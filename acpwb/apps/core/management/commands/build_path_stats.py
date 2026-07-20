"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Streams CrawlerVisit in small id-ordered batches, aggregates (host, path) → count
in Python, then bulk-upserts into PathStat. Keeps a high-water mark in
DashboardStat('path_stats_hwm') so reruns only touch new rows.

The GROUP BY is intentionally done in Python (not SQL) so PostgreSQL never needs
to sort or hash-aggregate more than batch_size rows at once.

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --batch-size 10000
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
import time

from django.core.management.base import BaseCommand

CURSOR_KEY = 'path_stats_hwm'


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit (low-memory id-cursor)'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch-size', type=int, default=10000,
            help='CrawlerVisit rows to stream per batch (default: 10000)',
        )
        parser.add_argument(
            '--reset', action='store_true',
            help='Truncate PathStat and reset cursor to reprocess all rows',
        )

    def handle(self, *args, **options):
        from django.db import connection

        from apps.core.models import DashboardStat
        from apps.honeypot.models import CrawlerVisit, PathStat

        batch_size = options['batch_size']

        if options['reset']:
            self.stdout.write('Resetting PathStat ...')
            PathStat.objects.all().delete()
            DashboardStat.objects.filter(key=CURSOR_KEY).delete()
            self.stdout.write('  Done.')

        cursor_stat = DashboardStat.objects.filter(key=CURSOR_KEY).first()
        last_id = int(cursor_stat.value) if cursor_stat else 0

        self.stdout.write(f'Resuming from CrawlerVisit id > {last_id:,}')

        total_rows = 0
        total_unique = 0
        t_start = time.monotonic()

        qs = (
            CrawlerVisit.objects
            .filter(id__gt=last_id)
            .order_by('id')
            .values('id', 'host', 'path')
        )

        batch = {}
        batch_max_id = last_id

        for row in qs.iterator(chunk_size=batch_size):
            key = (row['host'] or '', row['path'])
            batch[key] = batch.get(key, 0) + 1
            total_rows += 1
            if row['id'] > batch_max_id:
                batch_max_id = row['id']

            if total_rows % batch_size == 0:
                n = self._upsert(connection, batch)
                total_unique += n
                self._save_hwm(DashboardStat, batch_max_id)
                elapsed = time.monotonic() - t_start
                self.stdout.write(
                    f'  {total_rows:>10,} rows  id={batch_max_id:,}  '
                    f'{total_unique:,} unique paths  {elapsed:.0f}s'
                )
                batch = {}

        if batch:
            n = self._upsert(connection, batch)
            total_unique += n
            self._save_hwm(DashboardStat, batch_max_id)

        if total_rows == 0:
            self.stdout.write('  Nothing new to process.')
        else:
            elapsed = time.monotonic() - t_start
            self.stdout.write(
                f'Done. {total_rows:,} rows → {total_unique:,} unique paths '
                f'in {elapsed:.0f}s (hwm id={batch_max_id:,})'
            )

    def _upsert(self, connection, batch):
        rows = list(batch.items())
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

    def _save_hwm(self, DashboardStat, last_id):
        DashboardStat.objects.update_or_create(
            key=CURSOR_KEY, defaults={'value': last_id},
        )
