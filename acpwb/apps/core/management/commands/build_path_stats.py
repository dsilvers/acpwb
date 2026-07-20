"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Iterates TimescaleDB chunks one at a time. Each chunk is a small physical table
so PostgreSQL does a simple sequential scan with no cross-chunk sort. Rows are
streamed to Python in sub-batches for aggregation; the SQL GROUP BY is never
used. Resumes by tracking the last fully-processed chunk range_end in
DashboardStat('path_stats_hwm').

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
import time

from django.core.management.base import BaseCommand

CURSOR_KEY = 'path_stats_hwm'
FETCH_SIZE = 10000


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit (chunk-at-a-time, Python aggregation)'

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
                ORDER BY range_start
            """)
            chunks = cur.fetchall()

        if hwm:
            chunks = [(s, n, rs, re) for s, n, rs, re in chunks if str(re) > hwm]

        if not chunks:
            self.stdout.write('Nothing new to process.')
            return

        self.stdout.write(f'Processing {len(chunks)} chunks ...')
        t_start = time.monotonic()

        for chunk_schema, chunk_name, range_start, range_end in chunks:
            t0 = time.monotonic()
            full_name = f'"{chunk_schema}"."{chunk_name}"'
            counts = {}
            total_rows = 0

            with connection.cursor() as cur:
                cur.execute(f'SELECT COALESCE(host, \'\'), path FROM {full_name}')
                while True:
                    rows = cur.fetchmany(FETCH_SIZE)
                    if not rows:
                        break
                    for host, path in rows:
                        key = (host, path)
                        counts[key] = counts.get(key, 0) + 1
                    total_rows += len(rows)

            n_upserted = self._upsert(connection, counts)

            DashboardStat.objects.update_or_create(
                key=CURSOR_KEY, defaults={'value': str(range_end)},
            )

            self.stdout.write(
                f'  {chunk_name}  {range_start.date()}–{range_end.date()}'
                f'  {total_rows:,} rows  {n_upserted:,} unique paths'
                f'  ({time.monotonic()-t0:.1f}s)'
            )

        self.stdout.write(f'Done in {time.monotonic()-t_start:.0f}s total.')

    def _upsert(self, connection, counts):
        if not counts:
            return 0
        rows = list(counts.items())
        placeholders = ','.join(['(%s,%s,%s)'] * len(rows))
        params = [val for (host, path), count in rows for val in (host, path, count)]
        with connection.cursor() as cur:
            cur.execute(
                f"""
                INSERT INTO honeypot_pathstat (host, path, count)
                VALUES {placeholders}
                ON CONFLICT (host, path)
                DO UPDATE SET count = honeypot_pathstat.count + EXCLUDED.count
                """,
                params,
            )
        return len(rows)
