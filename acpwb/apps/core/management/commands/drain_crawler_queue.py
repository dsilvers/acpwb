"""
Drain the Redis crawler visit queue into PostgreSQL.

Only one instance runs at a time — an exclusive flock on /tmp/acpwb-drain.lock
prevents overlap. The OS releases the lock automatically if the process dies,
so a crashed run never blocks the next cron tick.

Run via cron every minute:
    * * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml \
        exec -T web python manage.py drain_crawler_queue \
        >> /var/log/acpwb-crawler-drain.log 2>&1
"""
import fcntl
import time
from datetime import datetime, timedelta, timezone
from django.utils.dateparse import parse_datetime

from django.core.management.base import BaseCommand

from apps.core.crawler_queue import (
    finalize_batch, pop_crawler_visits, queue_length, recover_crawler_visits,
)
from apps.honeypot.models import CrawlerVisit

_LOCK_FILE = '/tmp/acpwb-drain.lock'


class Command(BaseCommand):
    help = 'Drain the Redis crawler visit queue into PostgreSQL via bulk_create'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch', type=int, default=500,
            help='Number of records to pop per batch (default: 500)',
        )
        parser.add_argument(
            '--max-batches', type=int, default=0,
            help='Maximum number of batches per run, 0 = unlimited (default: 0)',
        )
        parser.add_argument(
            '--max-seconds', type=int, default=55,
            help='Stop fetching new batches after this many seconds (default: 55)',
        )

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
            # lock released automatically when with-block exits

    def _drop_already_inserted(self, items):
        """
        Recovery-only: a crash could have landed after the DB commit but
        before finalize_batch() ran, leaving an already-inserted batch
        looking unprocessed. There's no DB-level unique constraint to lean
        on here (see CrawlerVisit.idempotency_key — TimescaleDB hypertables
        don't support the concurrent index build needed to add one without
        blocking live writes), so check explicitly before re-inserting.

        Bounded to the last hour so TimescaleDB can chunk-exclude the rest
        of this (large, historical) hypertable instead of scanning it all —
        recovered batches are always from a run that started at most a
        couple of cron ticks ago.
        """
        keys = [item['idempotency_key'] for item in items if item.get('idempotency_key')]
        if not keys:
            return items
        existing = {
            str(k) for k in CrawlerVisit.objects.filter(
                idempotency_key__in=keys,
                timestamp__gte=datetime.now(timezone.utc) - timedelta(hours=1),
            ).values_list('idempotency_key', flat=True)
        }
        if not existing:
            return items
        return [item for item in items if item.get('idempotency_key') not in existing]

    def _process_batch(self, items, batch_key, is_recovery=False):
        """
        Insert `items` into the DB (bulk, falling back to per-row retry on a
        batch-level failure), then finalize_batch() the corresponding Redis
        processing key. Returns the number of records inserted.

        finalize_batch() only runs once this has done everything it's going
        to for this batch — if the process dies before returning, the batch
        key survives for recover_crawler_visits() to retry.
        """
        if is_recovery and items:
            items = self._drop_already_inserted(items)

        inserted = 0
        objs = []
        for item in items:
            try:
                if 'timestamp' in item:
                    ts = parse_datetime(item['timestamp'])
                    item['timestamp'] = ts if ts else item.pop('timestamp')
                if not item.get('ip_address'):
                    item['ip_address'] = '0.0.0.0'
                objs.append(CrawlerVisit(**item))
            except Exception:
                pass  # skip malformed entries

        if objs:
            try:
                CrawlerVisit.objects.bulk_create(objs, ignore_conflicts=True)
                inserted = len(objs)
            except Exception as exc:
                # A single bad row (e.g. an invalid ip_address format)
                # fails the whole batch INSERT — retry one row at a time so
                # the rest of an otherwise-valid batch isn't lost.
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
        for items, batch_key in recover_crawler_visits():
            recovered += self._process_batch(items, batch_key, is_recovery=True)
        if recovered:
            self._log(f'Recovered {recovered} record(s) from a prior interrupted run.')

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                break
            items, batch_key = pop_crawler_visits(batch_size)
            # Check batch_key, not items: a batch can be entirely malformed
            # JSON (items == []) while still having moved real entries that
            # need finalizing — only a falsy batch_key means the main queue
            # was actually empty (or Redis was unavailable).
            if not batch_key:
                break

            total_inserted += self._process_batch(items, batch_key)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = queue_length()
        depth_display = 'unknown (Redis unavailable)' if remaining < 0 else remaining
        self._log(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {depth_display}.'
        )
