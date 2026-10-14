import email as email_lib

from django.db import migrations, models


def backfill_message_id(apps, schema_editor):
    """Pipe (Cloudflare) emails never stored their Message-ID; read it from the
    raw message so redeliveries of already-stored emails are recognised."""
    InboundEmail = apps.get_model('webhooks', 'InboundEmail')
    for inbound in InboundEmail.objects.filter(message_id='').iterator(chunk_size=50):
        raw = (inbound.raw_payload or {}).get('raw')
        if not raw:
            continue
        message_id = (email_lib.message_from_string(raw).get('Message-ID') or '').strip()[:256]
        if message_id:
            InboundEmail.objects.filter(pk=inbound.pk).update(message_id=message_id)


class Migration(migrations.Migration):

    dependencies = [
        ('webhooks', '0004_honeypotmatch_source_code_email'),
    ]

    operations = [
        migrations.RenameField(
            model_name='inboundemail',
            old_name='mailgun_message_id',
            new_name='message_id',
        ),
        migrations.AlterField(
            model_name='inboundemail',
            name='message_id',
            field=models.CharField(blank=True, db_index=True, max_length=256),
        ),
        migrations.RunPython(backfill_message_id, migrations.RunPython.noop),
    ]
