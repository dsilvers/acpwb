"""
Drain the Redis archive visit queue into PostgreSQL.

Only one instance runs at a time — an exclusive flock on /tmp/acpwb-archive-drain.lock
prevents overlap. The OS releases the lock automatically if the process dies,
so a crashed run never blocks the next cron tick.

Run via cron every minute:
    * * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml \
        exec -T web python manage.py drain_archive_queue \
        >> /var/log/acpwb-archive-drain.log 2>&1
"""
import fcntl
import time
from django.utils.dateparse import parse_datetime

from django.core.management.base import BaseCommand

from apps.core.crawler_queue import pop_archive_visits, archive_queue_length
from apps.honeypot.models import ArchiveVisit

_LOCK_FILE = '/tmp/acpwb-archive-drain.lock'


class Command(BaseCommand):
    help = 'Drain the Redis archive visit queue into PostgreSQL via bulk_create'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch', type=int, default=500,
            help='Number of records to pop per batch (default: 500)',
        )
        parser.add_argument(
            '--max-batches', type=int, default=200,
            help='Maximum number of batches per run, 0 = unlimited (default: 200)',
        )
        parser.add_argument(
            '--max-seconds', type=int, default=50,
            help='Stop fetching new batches after this many seconds (default: 50)',
        )

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another drain is already running — exiting.')
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
            items = pop_archive_visits(batch_size)
            if not items:
                break

            objs = []
            for item in items:
                try:
                    if 'timestamp' in item:
                        ts = parse_datetime(item['timestamp'])
                        item['timestamp'] = ts if ts else item.pop('timestamp')
                    objs.append(ArchiveVisit(**item))
                except Exception:
                    pass  # skip malformed entries

            if objs:
                ArchiveVisit.objects.bulk_create(objs, ignore_conflicts=True)
                total_inserted += len(objs)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = archive_queue_length()
        self.stdout.write(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {remaining}.'
        )
