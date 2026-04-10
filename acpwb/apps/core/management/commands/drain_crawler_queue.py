"""
Drain the Redis crawler visit queue into PostgreSQL.

Run via cron every minute:
    * * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml \
        exec -T web python manage.py drain_crawler_queue \
        >> /var/log/acpwb-crawler-drain.log 2>&1
"""
from django.core.management.base import BaseCommand

from apps.core.crawler_queue import pop_crawler_visits, queue_length
from apps.honeypot.models import CrawlerVisit


class Command(BaseCommand):
    help = 'Drain the Redis crawler visit queue into PostgreSQL via bulk_create'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch', type=int, default=500,
            help='Number of records to pop per batch (default: 500)',
        )
        parser.add_argument(
            '--max-batches', type=int, default=200,
            help='Maximum number of batches per run, 0 = unlimited (default: 200)',
        )

    def handle(self, *args, **options):
        batch_size = options['batch']
        max_batches = options['max_batches']
        total_inserted = 0
        batches_run = 0

        while True:
            items = pop_crawler_visits(batch_size)
            if not items:
                break

            objs = []
            for item in items:
                try:
                    objs.append(CrawlerVisit(**item))
                except Exception:
                    pass  # skip malformed entries

            if objs:
                CrawlerVisit.objects.bulk_create(objs, ignore_conflicts=True)
                total_inserted += len(objs)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        remaining = queue_length()
        self.stdout.write(
            f'Inserted {total_inserted} records in {batches_run} batch(es). '
            f'Queue depth: {remaining}.'
        )
