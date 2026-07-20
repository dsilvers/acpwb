"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Iterates TimescaleDB chunks by time range. Each chunk is queried with a
timestamp filter so TimescaleDB's chunk exclusion scopes the scan to one
physical table at a time. Rows are streamed via a server-side cursor
(Django iterator) and aggregated in Python — no SQL GROUP BY, no ORDER BY,
no cross-chunk sort.

Resumes by tracking the last fully-processed chunk range_end in
DashboardStat('path_stats_hwm').

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
import time

from django.core.management.base import BaseCommand

CURSOR_KEY = 'path_stats_hwm'
STREAM_CHUNK = 10000


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit (streamed, chunk-at-a-time)'

    def add_arguments(self, parser):
        parser.add_argument(
            '--reset', action='store_true',
            help='Truncate PathStat and reset cursor to reprocess all chunks',
        )

    def handle(self, *args, **options):
        from django.db import connection

        from apps.core.models import DashboardStat
        from apps.honeypot.models import CrawlerVisit, PathStat

        if options['reset']:
            self.stdout.write('Resetting PathStat ...')
            PathStat.objects.all().delete()
            DashboardStat.objects.filter(key=CURSOR_KEY).delete()
            self.stdout.write('  Done.')

        cursor_stat = DashboardStat.objects.filter(key=CURSOR_KEY).first()
        hwm = cursor_stat.value if cursor_stat else None

        with connection.cursor() as cur:
            cur.execute("""
                SELECT range_start, range_end
                FROM timescaledb_information.chunks
                WHERE hypertable_name = 'honeypot_crawlervisit'
                ORDER BY range_start
            """)
            chunks = cur.fetchall()

        if hwm:
            chunks = [(rs, re) for rs, re in chunks if str(re) > hwm]

        if not chunks:
            self.stdout.write('Nothing new to process.')
            return

        self.stdout.write(f'Processing {len(chunks)} chunks ...')
        t_start = time.monotonic()

        for range_start, range_end in chunks:
            t0 = time.monotonic()

            # Timestamp filter lets TimescaleDB exclude all other chunks.
            # iterator() opens a server-side cursor — rows stream in STREAM_CHUNK
            # batches without buffering the full result in Python or PostgreSQL.
            qs = (
                CrawlerVisit.objects
                .filter(timestamp__gte=range_start, timestamp__lt=range_end)
                .values('host', 'path')
            )

            counts = {}
            total_rows = 0
            for row in qs.iterator(chunk_size=STREAM_CHUNK):
                key = (row['host'] or '', row['path'])
                counts[key] = counts.get(key, 0) + 1
                total_rows += 1

            n_upserted = self._upsert(connection, counts)

            DashboardStat.objects.update_or_create(
                key=CURSOR_KEY, defaults={'value': str(range_end)},
            )

            self.stdout.write(
                f'  {range_start.date()}–{range_end.date()}'
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
