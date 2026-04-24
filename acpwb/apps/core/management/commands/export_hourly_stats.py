"""
Print hourly crawler request counts as CSV to stdout.

Usage:
    python manage.py export_hourly_stats
    python manage.py export_hourly_stats --bot-breakdown
"""
import csv
import sys

from django.core.management.base import BaseCommand
from django.db.models import Count
from django.db.models.functions import TruncHour

from apps.honeypot.models import CrawlerVisit


class Command(BaseCommand):
    help = 'Print hourly crawler request counts as CSV'

    def add_arguments(self, parser):
        parser.add_argument(
            '--bot-breakdown',
            action='store_true',
            help='Include a column per bot_type instead of a single total',
        )

    def handle(self, *args, **options):
        writer = csv.writer(sys.stdout)

        if options['bot_breakdown']:
            self._write_bot_breakdown(writer)
        else:
            self._write_totals(writer)

    def _write_totals(self, writer):
        writer.writerow(['hour', 'requests'])
        qs = (
            CrawlerVisit.objects
            .annotate(hour=TruncHour('timestamp'))
            .values('hour')
            .annotate(requests=Count('id'))
            .order_by('hour')
        )
        for row in qs:
            writer.writerow([row['hour'].strftime('%Y-%m-%d %H:00'), row['requests']])

    def _write_bot_breakdown(self, writer):
        qs = (
            CrawlerVisit.objects
            .annotate(hour=TruncHour('timestamp'))
            .values('hour', 'bot_type')
            .annotate(requests=Count('id'))
            .order_by('hour', 'bot_type')
        )

        # Collect all data first to determine bot columns
        from collections import defaultdict
        data = defaultdict(lambda: defaultdict(int))
        bot_types = set()
        hours = []

        for row in qs:
            h = row['hour'].strftime('%Y-%m-%d %H:00')
            b = row['bot_type'] or '(none)'
            if h not in data:
                hours.append(h)
            data[h][b] += row['requests']
            bot_types.add(b)

        bot_cols = sorted(bot_types)
        writer.writerow(['hour', 'total'] + bot_cols)
        for h in hours:
            total = sum(data[h].values())
            writer.writerow([h, total] + [data[h].get(b, 0) for b in bot_cols])
