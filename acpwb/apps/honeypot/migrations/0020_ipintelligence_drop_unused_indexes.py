from django.db import migrations, models

# Production pg_stat_user_indexes (2026-09-30), honeypot_ipintelligence:
#   ip_address_key 290.8M scans, pkey 3.4M, enriched_at 28, (country_code,
#   is_hosting) 27 — everything below had 0–11, all from one ad-hoc report run.
# Each extra index was still written on every update. The first_seen/last_seen
# indexes were the expensive ones: discover_ip_intelligence rewrites those
# columns for every IP it sees, so no update could be HOT and each one wrote
# all 15 indexes (~1,600 rows/s; ~3 min per 1-hour discovery window).
#
# DROP INDEX CONCURRENTLY can't run in a transaction, hence atomic = False and
# one statement per index. IF EXISTS keeps it safe on databases where an index
# was never created. fillfactor leaves room on each page for HOT updates.
_DROPPED = [
    ('honeypot_ipintelligence_ip_version_4edfda90', 'btree (ip_version)'),
    ('honeypot_ipintelligence_country_code_a41397be', 'btree (country_code)'),
    ('honeypot_ipintelligence_country_code_a41397be_like', 'btree (country_code varchar_pattern_ops)'),
    ('honeypot_ipintelligence_asn_0bb0d898', 'btree (asn)'),
    ('honeypot_ipintelligence_asn_org_7c7fdb15', 'btree (asn_org)'),
    ('honeypot_ipintelligence_asn_org_7c7fdb15_like', 'btree (asn_org varchar_pattern_ops)'),
    ('honeypot_ipintelligence_is_hosting_b3b4dcc5', 'btree (is_hosting)'),
    ('honeypot_ipintelligence_is_tor_exit_f7f5c804', 'btree (is_tor_exit)'),
    ('honeypot_ipintelligence_first_seen_a1cd03e5', 'btree (first_seen)'),
    ('honeypot_ipintelligence_last_seen_94fd0e85', 'btree (last_seen)'),
    ('honeypot_ip_is_host_d82d2c_idx', 'btree (is_hosting, is_tor_exit)'),
]


class Migration(migrations.Migration):
    atomic = False

    dependencies = [
        ('honeypot', '0019_publishedipreputation_ipreputationscore_and_more'),
    ]

    operations = [
        migrations.SeparateDatabaseAndState(
            database_operations=[
                migrations.RunSQL(
                    sql=[f'DROP INDEX CONCURRENTLY IF EXISTS {name}' for name, _ in _DROPPED],
                    reverse_sql=[
                        f'CREATE INDEX CONCURRENTLY IF NOT EXISTS {name} ON honeypot_ipintelligence USING {definition}'
                        for name, definition in _DROPPED
                    ],
                ),
                migrations.RunSQL(
                    sql='ALTER TABLE honeypot_ipintelligence SET (fillfactor = 80)',
                    reverse_sql='ALTER TABLE honeypot_ipintelligence RESET (fillfactor)',
                ),
            ],
            state_operations=[
                migrations.RemoveIndex(model_name='ipintelligence', name='honeypot_ip_is_host_d82d2c_idx'),
                migrations.AlterField(model_name='ipintelligence', name='ip_version',
                                      field=models.PositiveSmallIntegerField(default=4)),
                migrations.AlterField(model_name='ipintelligence', name='country_code',
                                      field=models.CharField(blank=True, max_length=2)),
                migrations.AlterField(model_name='ipintelligence', name='asn',
                                      field=models.PositiveIntegerField(blank=True, null=True)),
                migrations.AlterField(model_name='ipintelligence', name='asn_org',
                                      field=models.CharField(blank=True, max_length=256)),
                migrations.AlterField(model_name='ipintelligence', name='is_hosting',
                                      field=models.BooleanField(default=False)),
                migrations.AlterField(model_name='ipintelligence', name='is_tor_exit',
                                      field=models.BooleanField(default=False)),
                migrations.AlterField(model_name='ipintelligence', name='first_seen',
                                      field=models.DateTimeField(blank=True, null=True)),
                migrations.AlterField(model_name='ipintelligence', name='last_seen',
                                      field=models.DateTimeField(blank=True, null=True)),
            ],
        ),
    ]
