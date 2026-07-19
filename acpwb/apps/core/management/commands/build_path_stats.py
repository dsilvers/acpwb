"""
Incrementally build the PathStat pre-aggregation table from CrawlerVisit rows.

Reads CrawlerVisit in batches ordered by id, aggregates (host, path) → count
in Python, then bulk-upserts into PathStat. Resumes from last processed id
stored in DashboardStat('path_stats_last_id').

Usage:
    manage.py build_path_stats
    manage.py build_path_stats --batch-size 10000
    manage.py build_path_stats --reset   # truncate PathStat and reprocess all
"""
from django.core.management.base import BaseCommand


CURSOR_KEY = 'path_stats_last_id'


class Command(BaseCommand):
    help = 'Incrementally populate PathStat from CrawlerVisit'

    def add_arguments(self, parser):
        parser.add_argument(
            '--batch-size', type=int, default=50000,
            help='CrawlerVisit rows to process per batch (default: 50000)',
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

        total_processed = 0
        total_upserted = 0
        max_id_seen = last_id

        qs = (
            CrawlerVisit.objects
            .filter(id__gt=last_id)
            .order_by('id')
            .values('id', 'host', 'path')
        )

        batch = {}
        batch_max_id = last_id

        for row in qs.iterator(chunk_size=batch_size):
            host = row['host'] or ''
            path = row['path']
            key = (host, path)
            batch[key] = batch.get(key, 0) + 1
            total_processed += 1
            if row['id'] > batch_max_id:
                batch_max_id = row['id']

            if total_processed % batch_size == 0:
                upserted = self._upsert_batch(connection, batch)
                total_upserted += upserted
                max_id_seen = batch_max_id
                self._save_cursor(DashboardStat, max_id_seen)
                self.stdout.write(
                    f'  Processed {total_processed:,} rows, '
                    f'up to id {max_id_seen:,}, '
                    f'{total_upserted:,} unique paths total'
                )
                batch = {}

        # Final partial batch
        if batch:
            upserted = self._upsert_batch(connection, batch)
            total_upserted += upserted
            max_id_seen = batch_max_id
            self._save_cursor(DashboardStat, max_id_seen)

        if total_processed == 0:
            self.stdout.write('  Nothing new to process.')
        else:
            self.stdout.write(
                f'Done. {total_processed:,} rows processed, '
                f'{total_upserted:,} unique paths, '
                f'cursor at id {max_id_seen:,}.'
            )

    def _upsert_batch(self, connection, batch):
        if not batch:
            return 0
        rows = [(host, path, count) for (host, path), count in batch.items()]
        placeholders = ','.join(['(%s,%s,%s)'] * len(rows))
        flat_params = [val for row in rows for val in row]
        with connection.cursor() as cur:
            cur.execute(
                f"""
                INSERT INTO honeypot_pathstat (host, path, count)
                VALUES {placeholders}
                ON CONFLICT (host, path)
                DO UPDATE SET count = honeypot_pathstat.count + EXCLUDED.count
                """,
                flat_params,
            )
        return len(rows)

    def _save_cursor(self, DashboardStat, last_id):
        DashboardStat.objects.update_or_create(
            key=CURSOR_KEY,
            defaults={'value': last_id},
        )
