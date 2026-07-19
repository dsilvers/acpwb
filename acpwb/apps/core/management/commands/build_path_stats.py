"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Processes one TimescaleDB chunk at a time, running a GROUP BY directly against
each chunk table to avoid cross-chunk scans. Resumes by tracking the last
fully-processed chunk range_end in DashboardStat('path_stats_last_id').

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
import time

from django.core.management.base import BaseCommand

CURSOR_KEY = 'path_stats_hwm'


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit (chunk-at-a-time)'

    def add_arguments(self, parser):
        parser.add_argument(
            '--reset', action='store_true',
            help='Truncate PathStat and reset cursor to reprocess all chunks',
        )

    def handle(self, *args, **options):
        from django.db import connection
        from apps.core.models import DashboardStat
        from apps.honeypot.models import PathStat

        if options['reset']:
            self.stdout.write('Resetting PathStat ...')
            PathStat.objects.all().delete()
            DashboardStat.objects.filter(key=CURSOR_KEY).delete()
            self.stdout.write('  Done.')

        cursor_stat = DashboardStat.objects.filter(key=CURSOR_KEY).first()
        hwm = cursor_stat.value if cursor_stat else None

        with connection.cursor() as cur:
            cur.execute("""
                SELECT chunk_schema, chunk_name, range_start, range_end
                FROM timescaledb_information.chunks
                WHERE hypertable_name = 'honeypot_crawlervisit'
                ORDER BY range_start DESC
            """)
            chunks = cur.fetchall()

        if hwm:
            chunks = [(s, n, rs, re) for s, n, rs, re in chunks if str(re) > hwm]

        if not chunks:
            self.stdout.write('Nothing new to process.')
            return

        self.stdout.write(f'Processing {len(chunks)} chunks ...')
        t_start = time.monotonic()
        total_paths = 0

        for chunk_schema, chunk_name, range_start, range_end in chunks:
            t0 = time.monotonic()
            full_name = f'"{chunk_schema}"."{chunk_name}"'

            with connection.cursor() as cur:
                cur.execute(f"""
                    INSERT INTO honeypot_pathstat (host, path, count)
                    SELECT COALESCE(host, ''), path, COUNT(*)
                    FROM {full_name}
                    GROUP BY COALESCE(host, ''), path
                    ON CONFLICT (host, path)
                    DO UPDATE SET count = honeypot_pathstat.count + EXCLUDED.count
                """)
                paths = cur.rowcount

            total_paths += paths
            DashboardStat.objects.update_or_create(
                key=CURSOR_KEY, defaults={'value': str(range_end)},
            )
            self.stdout.write(
                f'  {chunk_name}  {range_start.date()}–{range_end.date()}'
                f'  {paths:,} paths  ({time.monotonic()-t0:.1f}s)'
            )

        self.stdout.write(
            f'Done. {total_paths:,} path upserts across {len(chunks)} chunks '
            f'in {time.monotonic()-t_start:.1f}s'
        )
