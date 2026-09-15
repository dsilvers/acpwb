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

_LOCK_FILE = '/tmp/acpwb-fingerprint-drain.lock'


class Command(BaseCommand):
    help = 'Drain the Redis fingerprint-signal queue into PostgreSQL via bulk_create'

    def add_arguments(self, parser):
        parser.add_argument('--batch', type=int, default=500)
        parser.add_argument('--max-batches', type=int, default=0)
        parser.add_argument('--max-seconds', type=int, default=55)

    def _log(self, msg):
        ts = datetime.now(timezone.utc).isoformat()
        self.stdout.write(f'[{ts}] {msg}')

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self._log('Another drain is already running — exiting.')
                return
            self._drain(options)

    def _process_batch(self, items, batch_key):
        objs = []
        for item in items:
            try:
                if 'timestamp' in item:
                    ts = parse_datetime(item['timestamp'])
                    item['timestamp'] = ts if ts else item.pop('timestamp')
                if not item.get('ip_address'):
                    item['ip_address'] = '0.0.0.0'
                objs.append(RequestFingerprint(**item))
            except Exception:
                pass  # skip malformed entries

        inserted = 0
        if objs:
            try:
                RequestFingerprint.objects.bulk_create(objs, ignore_conflicts=True)
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
        for items, batch_key in recover_fingerprint_signals():
            recovered += self._process_batch(items, batch_key)
        if recovered:
            self._log(f'Recovered {recovered} record(s) from a prior interrupted run.')

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                break
            items, batch_key = pop_fingerprint_signals(batch_size)
            if not batch_key:
                break

            total_inserted += self._process_batch(items, batch_key)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = fingerprint_queue_length()
        depth_display = 'unknown (Redis unavailable)' if remaining < 0 else remaining
        self._log(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {depth_display}.'
        )
