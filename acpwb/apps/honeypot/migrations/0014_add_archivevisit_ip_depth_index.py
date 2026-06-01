from django.db import migrations


class Migration(migrations.Migration):
    dependencies = [
        ('honeypot', '0013_add_presentation_trap_type'),
    ]
    operations = [
        migrations.RunSQL(
            sql="""
                CREATE INDEX IF NOT EXISTS honeypot_archivevisit_ip_depth_idx
                    ON honeypot_archivevisit (ip_address, depth);
            """,
            reverse_sql="DROP INDEX IF EXISTS honeypot_archivevisit_ip_depth_idx;",
        ),
    ]
