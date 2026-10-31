from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('honeypot', '0032_compliance_completion_classification'),
    ]

    operations = [
        migrations.CreateModel(
            name='CrawlerMinute',
            fields=[
                ('id', models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name='ID')),
                ('minute', models.DateTimeField()),
                ('bot_type', models.CharField(blank=True, max_length=64)),
                ('bot_group', models.CharField(blank=True, max_length=64)),
                ('trap_type', models.CharField(max_length=32)),
                ('count', models.IntegerField(default=0)),
            ],
            options={
                'verbose_name': 'Crawler Minute',
                'ordering': ['-minute'],
            },
        ),
        migrations.AddIndex(
            model_name='crawlerminute',
            index=models.Index(fields=['bot_type', 'minute'], name='honeypot_cr_bot_typ_b92e2f_idx'),
        ),
        migrations.AddIndex(
            model_name='crawlerminute',
            index=models.Index(fields=['trap_type', 'minute'], name='honeypot_cr_trap_ty_2fb27d_idx'),
        ),
        migrations.AddIndex(
            model_name='crawlerminute',
            index=models.Index(fields=['bot_group', 'minute'], name='honeypot_cr_bot_gro_de9c2d_idx'),
        ),
        migrations.AddConstraint(
            model_name='crawlerminute',
            constraint=models.UniqueConstraint(fields=('minute', 'bot_type', 'bot_group', 'trap_type'),
                                               name='honeypot_crawlerminute_uniq'),
        ),
        # Hypertable, as 0026 does for the edge rollups. About 50k rows a
        # day, so monthly chunks. Never compressed: re-runs replace rows.
        migrations.RunSQL(
            sql="""
                CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

                ALTER TABLE honeypot_crawlerminute DROP CONSTRAINT honeypot_crawlerminute_pkey;
                SELECT create_hypertable('honeypot_crawlerminute', 'minute',
                    migrate_data => true, if_not_exists => true,
                    chunk_time_interval => INTERVAL '30 days');
                CREATE INDEX IF NOT EXISTS honeypot_crawlerminute_id_idx ON honeypot_crawlerminute (id);
            """,
            reverse_sql=migrations.RunSQL.noop,
        ),
    ]
