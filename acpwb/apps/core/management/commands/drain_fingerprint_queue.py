"""
Drain the Redis fingerprint-signal queue into PostgreSQL.

Simpler than drain_crawler_queue: RequestFingerprint is a brand-new table
(not an existing hypertable), so idempotency_key has a real unique DB
constraint and bulk_create(ignore_conflicts=True) handles recovery-replay
dedup on its own — no need for CrawlerVisit's app-level "check before
re-insert" workaround.

Run via cron every minute, same as drain_crawler_queue/drain_archive_queue:
    * * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml \
        exec -T web python manage.py drain_fingerprint_queue \
        >> /var/log/acpwb-fingerprint-drain.log 2>&1
"""
import fcntl
import time
from datetime import datetime, timezone

from django.core.management.base import BaseCommand
from django.utils.dateparse import parse_datetime

from apps.core.crawler_queue import (
    finalize_batch, fingerprint_queue_length, pop_fingerprint_signals, recover_fingerprint_signals,
)
from apps.honeypot.models import RequestFingerprint

class Command(BaseCommand):
    help = 'Drain the Redis fingerprint-signal queue into PostgreSQL via bulk_create'

    # drain_edge_conn_queue subclasses this command and swaps these out.
    model = RequestFingerprint
    lock_file = '/tmp/acpwb-fingerprint-drain.lock'
    pop = staticmethod(pop_fingerprint_signals)
    recover = staticmethod(recover_fingerprint_signals)
    queue_length = staticmethod(fingerprint_queue_length)

    def add_arguments(self, parser):
        parser.add_argument('--batch', type=int, default=500)
        parser.add_argument('--max-batches', type=int, default=0)
        parser.add_argument('--max-seconds', type=int, default=55)

    def _log(self, msg):
        ts = datetime.now(timezone.utc).isoformat()
        self.stdout.write(f'[{ts}] {msg}')

    def handle(self, *args, **options):
        with open(self.lock_file, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self._log('Another drain is already running — exiting.')
                return
            self._drain(options)

    def _process_batch(self, items, batch_key):
        # Keys the model doesn't know are dropped rather than failing the
        # row: a producer (acpwb_go, acpwb_edge) that ships a new field
        # before this side's migration would otherwise lose every record.
        fields = {f.attname for f in self.model._meta.concrete_fields}
        objs = []
        for item in items:
            try:
                item = {k: v for k, v in item.items() if k in fields}
                if 'timestamp' in item:
                    ts = parse_datetime(item['timestamp'])
                    item['timestamp'] = ts if ts else item.pop('timestamp')
                if not item.get('ip_address'):
                    item['ip_address'] = '0.0.0.0'
                objs.append(self.model(**item))
            except Exception:
                pass  # skip malformed entries

        inserted = 0
        if objs:
            try:
                self.model.objects.bulk_create(objs, ignore_conflicts=True)
                inserted = len(objs)
            except Exception as exc:
                self._log(f'Batch insert failed ({exc}); retrying rows individually.')
                for obj in objs:
                    try:
                        obj.save()
                        inserted += 1
                    except Exception:
                        pass

        finalize_batch(batch_key)
        return inserted

    def _drain(self, options):
        batch_size = options['batch']
        max_batches = options['max_batches']
        max_seconds = options['max_seconds']
        total_inserted = 0
        batches_run = 0
        started = time.monotonic()

        recovered = 0
        for items, batch_key in self.recover():
            recovered += self._process_batch(items, batch_key)
        if recovered:
            self._log(f'Recovered {recovered} record(s) from a prior interrupted run.')

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                break
            items, batch_key = self.pop(batch_size)
            if not batch_key:
                break

            total_inserted += self._process_batch(items, batch_key)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = self.queue_length()
        depth_display = 'unknown (Redis unavailable)' if remaining < 0 else remaining
        self._log(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {depth_display}.'
        )
