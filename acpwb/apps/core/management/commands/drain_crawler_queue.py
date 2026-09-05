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
from datetime import datetime, timezone
from django.utils.dateparse import parse_datetime

from django.core.management.base import BaseCommand

from apps.core.crawler_queue import pop_crawler_visits, queue_length
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

    def _drain(self, options):
        batch_size = options['batch']
        max_batches = options['max_batches']
        max_seconds = options['max_seconds']
        total_inserted = 0
        batches_run = 0
        started = time.monotonic()

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                break
            items = pop_crawler_visits(batch_size)
            if not items:
                break

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
                    total_inserted += len(objs)
                except Exception as exc:
                    # A single bad row (e.g. an invalid ip_address format)
                    # fails the whole batch INSERT — retry one row at a time
                    # so the rest of an otherwise-valid batch isn't lost.
                    self._log(f'Batch insert failed ({exc}); retrying rows individually.')
                    for obj in objs:
                        try:
                            obj.save()
                            total_inserted += 1
                        except Exception:
                            pass

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = queue_length()
        depth_display = 'unknown (Redis unavailable)' if remaining < 0 else remaining
        self._log(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {depth_display}.'
        )
