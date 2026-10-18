import email as email_lib
from email.header import decode_header, make_header

from django.db import migrations, models


def _decode_part(part):
    data = part.get_payload(decode=True) or b''
    try:
        return data.decode(part.get_content_charset() or 'utf-8', errors='replace')
    except LookupError:
        return data.decode('utf-8', errors='replace')


def fix_stored_emails(apps, schema_editor):
    """Repair rows stored by the old pipe_inbound:
    - from_header filled from the raw From: header;
    - subjects left RFC 2047-encoded ('=?UTF-8?Q?…?=') → decoded text;
    - single-part text/html bodies saved in body_plain → body_html;
    - bodies decoded as UTF-8 regardless of charset → re-decoded with it."""
    InboundEmail = apps.get_model('webhooks', 'InboundEmail')
    for inbound in InboundEmail.objects.iterator(chunk_size=50):
        changes = {}
        if not inbound.from_header:
            source = ''
            if (inbound.raw_payload or {}).get('raw'):
                source = str(email_lib.message_from_string(inbound.raw_payload['raw']).get('From') or '')
            else:
                source = (inbound.raw_payload or {}).get('from', '')  # legacy Mailgun webhook
            if source:
                try:
                    changes['from_header'] = str(make_header(decode_header(source))).strip()[:512]
                except Exception:
                    changes['from_header'] = source.strip()[:512]
        if '=?' in inbound.subject:
            try:
                changes['subject'] = str(make_header(decode_header(inbound.subject)))[:512]
            except Exception:
                pass
        raw = (inbound.raw_payload or {}).get('raw')
        if raw:
            msg = email_lib.message_from_string(raw)
            plain = html = ''
            if msg.is_multipart():
                for part in msg.walk():
                    ct = part.get_content_type()
                    if ct == 'text/plain' and not plain:
                        plain = _decode_part(part)
                    elif ct == 'text/html' and not html:
                        html = _decode_part(part)
            elif msg.get_content_type() == 'text/html':
                html = _decode_part(msg)
            else:
                plain = _decode_part(msg)
            if plain != inbound.body_plain:
                changes['body_plain'] = plain
            if html != inbound.body_html:
                changes['body_html'] = html
        if changes:
            InboundEmail.objects.filter(pk=inbound.pk).update(**changes)


class Migration(migrations.Migration):

    dependencies = [
        ('webhooks', '0006_outboundemail'),
    ]

    operations = [
        migrations.AddField(
            model_name='inboundemail',
            name='from_header',
            field=models.CharField(blank=True, max_length=512),
        ),
        migrations.RunPython(fix_stored_emails, migrations.RunPython.noop),
    ]
