from django.db import migrations


class Migration(migrations.Migration):
    dependencies = [
        ('honeypot', '0009_timestamp_default'),
    ]
    operations = [
        migrations.RunSQL(
            sql="""
                -- Enable extension (required on existing production volume where
                -- the docker-entrypoint-initdb.d script was skipped)
                CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

                -- ── CrawlerVisit ──────────────────────────────────────────────
                -- TimescaleDB requires all unique constraints to include the
                -- partition column (timestamp). Drop the id PK first; the
                -- PostgreSQL sequence still guarantees unique values in practice,
                -- and no other models have FKs to this table.
                ALTER TABLE honeypot_crawlervisit
                    DROP CONSTRAINT honeypot_crawlervisit_pkey;

                -- Convert to hypertable (holds exclusive lock ~2-5 min on prod)
                SELECT create_hypertable(
                    'honeypot_crawlervisit', 'timestamp',
                    migrate_data => true,
                    if_not_exists => true,
                    chunk_time_interval => INTERVAL '7 days'
                );

                -- Restore fast id lookups (non-unique index; sequence is globally unique)
                CREATE INDEX IF NOT EXISTS honeypot_crawlervisit_id_idx
                    ON honeypot_crawlervisit (id);

                -- Compression: segment by low-cardinality columns for best ratio
                ALTER TABLE honeypot_crawlervisit SET (
                    timescaledb.compress,
                    timescaledb.compress_orderby = 'timestamp DESC',
                    timescaledb.compress_segmentby = 'bot_type, trap_type'
                );
                SELECT add_compression_policy(
                    'honeypot_crawlervisit', INTERVAL '7 days', if_not_exists => true
                );

                -- ── ArchiveVisit ──────────────────────────────────────────────
                ALTER TABLE honeypot_archivevisit
                    DROP CONSTRAINT honeypot_archivevisit_pkey;

                SELECT create_hypertable(
                    'honeypot_archivevisit', 'timestamp',
                    migrate_data => true,
                    if_not_exists => true,
                    chunk_time_interval => INTERVAL '7 days'
                );

                CREATE INDEX IF NOT EXISTS honeypot_archivevisit_id_idx
                    ON honeypot_archivevisit (id);

                ALTER TABLE honeypot_archivevisit SET (
                    timescaledb.compress,
                    timescaledb.compress_orderby = 'timestamp DESC',
                    timescaledb.compress_segmentby = 'depth'
                );
                SELECT add_compression_policy(
                    'honeypot_archivevisit', INTERVAL '7 days', if_not_exists => true
                );
            """,
            reverse_sql=migrations.RunSQL.noop,
        ),
    ]
