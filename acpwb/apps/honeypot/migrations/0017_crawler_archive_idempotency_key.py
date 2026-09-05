from django.db import migrations, models


class Migration(migrations.Migration):
    dependencies = [
        ('honeypot', '0016_ipintelligence'),
    ]

    operations = [
        # Nullable column, no server-side default — ADD COLUMN is a fast,
        # metadata-only operation in Postgres regardless of table size, and
        # existing rows simply stay NULL forever (no backfill).
        #
        # Deliberately no unique constraint/index here: both tables are
        # TimescaleDB hypertables, which reject CREATE INDEX CONCURRENTLY
        # outright, and a non-concurrent build would block live writes on
        # these tables for the build's duration. The drain commands check
        # for already-inserted idempotency_keys with a plain (unindexed,
        # timestamp-bounded) query instead — see drain_crawler_queue.py /
        # drain_archive_queue.py.
        migrations.AddField(
            model_name='crawlervisit',
            name='idempotency_key',
            field=models.UUIDField(blank=True, default=None, null=True),
        ),
        migrations.AddField(
            model_name='archivevisit',
            name='idempotency_key',
            field=models.UUIDField(blank=True, default=None, null=True),
        ),
    ]
