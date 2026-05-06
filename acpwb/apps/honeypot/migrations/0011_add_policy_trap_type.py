from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('honeypot', '0010_timescaledb_hypertables'),
    ]

    operations = [
        migrations.AlterField(
            model_name='crawlervisit',
            name='trap_type',
            field=models.CharField(
                choices=[
                    ('archive', 'Archive Loop'),
                    ('ghost_link', 'Ghost Link'),
                    ('well_known', 'Well-Known File'),
                    ('api', 'Fake API'),
                    ('wiki', 'Wiki Page'),
                    ('pow', 'PoW Challenge'),
                    ('report_list', 'Report Listing'),
                    ('report_download', 'Report Download'),
                    ('dataset', 'Training Dataset'),
                    ('policy', 'Public Policy Filing'),
                    ('scanner_probe', 'Scanner Probe (404)'),
                    ('env_probe', 'Config File Probe'),
                    ('wp_probe', 'WordPress Probe'),
                    ('webshell_probe', 'Webshell Probe'),
                    ('canary_trigger', 'Canary Token Triggered'),
                    ('other', 'Other'),
                ],
                db_index=True,
                default='other',
                max_length=32,
            ),
        ),
    ]
