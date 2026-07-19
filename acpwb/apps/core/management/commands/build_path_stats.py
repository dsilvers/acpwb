"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Aggregates (host, path) → count in PostgreSQL in id-range batches, then
upserts into PathStat. Resumes from last processed id stored in
DashboardStat('path_stats_last_id').

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --batch-size 50000
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
import time

from django.core.management.base import BaseCommand


CURSOR_KEY = 'path_stats_last_id'
DEFAULT_BATCH = 50_000


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch-size', type=int, default=DEFAULT_BATCH,
            help=f'CrawlerVisit rows per batch (default: {DEFAULT_BATCH:,})',
        )
        parser.add_argument(
            '--reset', action='store_true',
            help='Truncate PathStat and reset cursor to reprocess all rows',
        )

    def handle(self, *args, **options):
        from django.db import connection
        from apps.core.models import DashboardStat
        from apps.honeypot.models import PathStat

        batch_size = options['batch_size']

        if options['reset']:
            self.stdout.write('Resetting PathStat ...')
            PathStat.objects.all().delete()
            DashboardStat.objects.filter(key=CURSOR_KEY).delete()
            self.stdout.write('  Done.')

        cursor_stat = DashboardStat.objects.filter(key=CURSOR_KEY).first()
        last_id = int(cursor_stat.value) if cursor_stat else 0

        with connection.cursor() as cur:
            cur.execute('SELECT MAX(id) FROM honeypot_crawlervisit WHERE id > %s', [last_id])
            max_id = cur.fetchone()[0]

        if max_id is None:
            self.stdout.write('Nothing new to process.')
            return

        total = max_id - last_id
        self.stdout.write(f'Resuming from id > {last_id:,} — processing up to {max_id:,} ({total:,} rows)')

        t_start = time.monotonic()
        total_paths = 0
        batch_start = last_id

        while batch_start < max_id:
            batch_end = min(batch_start + batch_size, max_id)
            t0 = time.monotonic()

            with connection.cursor() as cur:
                cur.execute("""
                    INSERT INTO honeypot_pathstat (host, path, count)
                    SELECT COALESCE(host, ''), path, COUNT(*)
                    FROM honeypot_crawlervisit
                    WHERE id > %s AND id <= %s
                    GROUP BY COALESCE(host, ''), path
                    ON CONFLICT (host, path)
                    DO UPDATE SET count = honeypot_pathstat.count + EXCLUDED.count
                """, [batch_start, batch_end])
                paths = cur.rowcount

            total_paths += paths
            DashboardStat.objects.update_or_create(
                key=CURSOR_KEY, defaults={'value': batch_end},
            )

            pct = (batch_end - last_id) / total * 100
            self.stdout.write(
                f'  id {batch_start:,}–{batch_end:,}  {paths:,} paths  '
                f'{pct:.0f}%  ({time.monotonic()-t0:.1f}s)'
            )
            batch_start = batch_end

        self.stdout.write(
            f'Done. {total_paths:,} path upserts in {time.monotonic()-t_start:.1f}s'
        )
