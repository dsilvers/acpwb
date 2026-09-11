from django.db import migrations


class Migration(migrations.Migration):
    dependencies = [
        ('honeypot', '0017_crawler_archive_idempotency_key'),
    ]
    operations = [
        migrations.RunSQL(
            sql="""
                -- The original 0010 migration paired chunk_time_interval and
                -- compress_after at 7 days each. A chunk isn't compression-eligible
                -- until it's fully closed *and* compress_after old, so that pairing
                -- let up to ~2 weeks of data sit fully uncompressed at once — fine
                -- at low traffic, but a 2026-09-05 traffic spike left ~450GB
                -- uncompressed across both hypertables and nearly filled the disk.
                --
                -- Only affects chunks created after this runs — existing chunks
                -- keep their original 7-day boundaries until they close naturally.
                SELECT set_chunk_time_interval('honeypot_crawlervisit', INTERVAL '1 day');
                SELECT set_chunk_time_interval('honeypot_archivevisit', INTERVAL '1 day');

                -- 2-day buffer (not 1) so the policy never tries to compress a
                -- chunk while drain_crawler_queue/drain_archive_queue (1-min cron,
                -- Redis-queued writes) are still landing stragglers for that day.
                SELECT remove_compression_policy('honeypot_crawlervisit', if_exists => true);
                SELECT add_compression_policy(
                    'honeypot_crawlervisit', INTERVAL '2 days', if_not_exists => true
                );

                SELECT remove_compression_policy('honeypot_archivevisit', if_exists => true);
                SELECT add_compression_policy(
                    'honeypot_archivevisit', INTERVAL '2 days', if_not_exists => true
                );

                -- With daily chunks, the default 12h job check cadence means a
                -- chunk that just closed could sit for up to half a day before the
                -- job even looks at it. 6h keeps that window tighter.
                SELECT alter_job(job_id, schedule_interval => INTERVAL '6 hours')
                FROM timescaledb_information.jobs
                WHERE hypertable_name = 'honeypot_crawlervisit' AND proc_name = 'policy_compression';

                SELECT alter_job(job_id, schedule_interval => INTERVAL '6 hours')
                FROM timescaledb_information.jobs
                WHERE hypertable_name = 'honeypot_archivevisit' AND proc_name = 'policy_compression';
            """,
            reverse_sql=migrations.RunSQL.noop,
        ),
    ]
