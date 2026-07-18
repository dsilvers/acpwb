"""
Build/update the TrafficMinuteStat per-minute aggregation table.

Queries CrawlerVisit grouped by (TruncMinute(timestamp), bot_type) and
upserts the counts into TrafficMinuteStat.  Processes one day at a time to
keep memory use flat and print live progress.

Usage:
    manage.py build_minute_stats           # incremental from last high-water mark
    manage.py build_minute_stats --full    # full rebuild (clears table first)
"""
import time
from datetime import datetime, timedelta, timezone as dt_tz

from django.core.management.base import BaseCommand
from django.db.models import Count
from django.db.models.functions import TruncMinute

from apps.core.models import DashboardStat, TrafficMinuteStat
from apps.honeypot.models import CrawlerVisit

_HWM_KEY = 'stats.minute_hwm'


class Command(BaseCommand):
    help = 'Build/update TrafficMinuteStat per-minute aggregation table from CrawlerVisit'

    def add_arguments(self, parser):
        parser.add_argument(
            '--full', action='store_true',
            help='Clear the table and rebuild from the earliest CrawlerVisit',
        )

    def handle(self, *args, **options):
        t_start = time.monotonic()
        full = options['full']

        earliest_qs = CrawlerVisit.objects.order_by('timestamp').values_list('timestamp', flat=True)
        earliest = earliest_qs.first()
        if earliest is None:
            self.stdout.write('No CrawlerVisit records found — nothing to do.')
            return

        if not earliest.tzinfo:
            earliest = earliest.replace(tzinfo=dt_tz.utc)

        if full:
            self.stdout.write('Full rebuild: clearing TrafficMinuteStat ...')
            deleted, _ = TrafficMinuteStat.objects.all().delete()
            self.stdout.write(f'  Deleted {deleted:,} existing rows.')
            DashboardStat.objects.filter(key=_HWM_KEY).delete()
            start_day = earliest.replace(hour=0, minute=0, second=0, microsecond=0)
        else:
            hwm_stat = DashboardStat.objects.filter(key=_HWM_KEY).first()
            if hwm_stat:
                hwm_dt = datetime.fromisoformat(hwm_stat.value)
                if not hwm_dt.tzinfo:
                    hwm_dt = hwm_dt.replace(tzinfo=dt_tz.utc)
                # Reprocess from 1 day before hwm to catch any late-arriving records
                start_day = (hwm_dt - timedelta(days=1)).replace(hour=0, minute=0, second=0, microsecond=0)
            else:
                start_day = earliest.replace(hour=0, minute=0, second=0, microsecond=0)

        now = datetime.now(dt_tz.utc)
        end_day = now.replace(hour=0, minute=0, second=0, microsecond=0) + timedelta(days=1)

        total_buckets = 0
        total_requests = 0
        current = start_day

        self.stdout.write(
            f'Processing {start_day.date()} → {(end_day - timedelta(days=1)).date()} inclusive ...\n'
        )

        while current < end_day:
            day_end = current + timedelta(days=1)
            day_t0 = time.monotonic()

            rows = list(
                CrawlerVisit.objects
                .filter(timestamp__gte=current, timestamp__lt=day_end)
                .annotate(minute=TruncMinute('timestamp'))
                .values('minute', 'bot_type')
                .annotate(c=Count('id'))
            )

            day_requests = sum(r['c'] for r in rows)

            if rows:
                objs = [
                    TrafficMinuteStat(
                        minute=r['minute'],
                        bot_type=r['bot_type'] or '',
                        count=r['c'],
                    )
                    for r in rows
                ]
                TrafficMinuteStat.objects.bulk_create(
                    objs,
                    update_conflicts=True,
                    unique_fields=['minute', 'bot_type'],
                    update_fields=['count'],
                )

            elapsed = time.monotonic() - day_t0
            total_buckets += len(rows)
            total_requests += day_requests
            self.stdout.write(
                f'  {current.date()}  {len(rows):5d} minute-buckets  '
                f'{day_requests:9,} requests  ({elapsed:.1f}s)'
            )

            current = day_end

        DashboardStat.objects.update_or_create(
            key=_HWM_KEY,
            defaults={'value': now.isoformat()},
        )

        elapsed_total = time.monotonic() - t_start
        self.stdout.write(
            f'\nDone. {total_buckets:,} minute-buckets, '
            f'{total_requests:,} total requests in {elapsed_total:.1f}s'
        )
