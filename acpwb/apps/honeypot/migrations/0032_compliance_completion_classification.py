from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('honeypot', '0031_compliance_completion'),
    ]

    operations = [
        migrations.AddField(
            model_name='compliancecompletion',
            name='bot_type',
            field=models.CharField(blank=True, db_index=True, max_length=64),
        ),
        migrations.AddField(
            model_name='compliancecompletion',
            name='client_class',
            field=models.CharField(blank=True, db_index=True, max_length=16),
        ),
    ]
