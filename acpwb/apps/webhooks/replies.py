"""Reading attachments out of stored inbound mail, and sending dashboard replies
through Mailgun (Django's SMTP backend, configured in settings)."""
import email as email_lib
import re
from email import policy
from email.utils import formataddr, getaddresses, make_msgid

from django.conf import settings
from django.core.exceptions import ValidationError
from django.core.mail import EmailMessage
from django.core.validators import validate_email

from .models import OutboundEmail

LOCAL_PART_RE = re.compile(r"^[A-Za-z0-9!#$%&'*+/=?^_`{|}~-]+(\.[A-Za-z0-9!#$%&'*+/=?^_`{|}~-]+)*$")


def parse_raw(inbound):
    """The original RFC 2822 message, or None. Only Cloudflare (pipe) deliveries
    store it; the legacy Mailgun webhook kept the parsed fields only."""
    raw = (inbound.raw_payload or {}).get('raw')
    if not raw:
        return None
    try:
        return email_lib.message_from_string(raw, policy=policy.default)
    except Exception:
        return email_lib.message_from_string(raw)


def _attachment_parts(msg):
    if msg is None:
        return []
    return [
        part for part in msg.walk()
        if not part.is_multipart()
        and (part.get_filename() or part.get_content_disposition() == 'attachment')
    ]


def attachments(inbound):
    """[{index, filename, content_type, size}] in MIME order (inline images included)."""
    out = []
    for i, part in enumerate(_attachment_parts(parse_raw(inbound))):
        data = part.get_payload(decode=True) or b''
        out.append({
            'index': i,
            'filename': part.get_filename() or f'attachment-{i + 1}',
            'content_type': part.get_content_type(),
            'size': len(data),
        })
    return out


def attachment(inbound, index):
    """(filename, content_type, bytes) for one attachment, or None."""
    parts = _attachment_parts(parse_raw(inbound))
    if not 0 <= index < len(parts):
        return None
    part = parts[index]
    return (
        part.get_filename() or f'attachment-{index + 1}',
        part.get_content_type(),
        part.get_payload(decode=True) or b'',
    )


def html_to_text(html):
    """Readable text from an HTML body: drops head/style/script, turns block
    ends into line breaks, decodes entities, collapses blank runs."""
    from html import unescape
    text = re.sub(r'(?is)<(head|style|script|title)\b.*?</\1\s*>', '', html or '')
    text = re.sub(r'(?is)<!--.*?-->', '', text)
    text = re.sub(r'(?i)<br\s*/?>', '\n', text)
    text = re.sub(r'(?i)</(p|div|tr|li|h[1-6]|table|blockquote)\s*>', '\n', text)
    text = re.sub(r'(?s)<[^>]+>', '', text)
    lines = [re.sub(r'[ \t\xa0]+', ' ', line).strip() for line in unescape(text).splitlines()]
    return re.sub(r'\n{3,}', '\n\n', '\n'.join(lines)).strip()


def body_text(inbound):
    return inbound.body_plain or html_to_text(inbound.body_html)


def _header(msg, name):
    return str(msg.get(name, '')) if msg is not None else ''


def header(inbound, name):
    """One decoded header of the original message ('' when absent or no raw)."""
    return _header(parse_raw(inbound), name)


def reply_defaults(inbound):
    """Initial values for the reply form."""
    msg = parse_raw(inbound)
    domain = settings.REPLY_FROM_DOMAIN

    # From: the address the mail was delivered to, with the display name the
    # sender used for it in To/Cc. Names that just repeat the address (or the
    # part before the @) are dropped.
    recipient = inbound.recipient.strip().lower()
    local, _, rdomain = recipient.partition('@')
    from_name = ''
    for name, addr in getaddresses([_header(msg, 'To'), _header(msg, 'Cc')]):
        if addr.lower() == recipient and name.lower() not in (recipient, local):
            from_name = name
            break
    if rdomain != domain or not LOCAL_PART_RE.match(local):
        local = 'general-inquiry'

    # To: Reply-To, then From; the envelope sender (often an SRS/BATV address
    # like prvs=…=user@host) only as a last resort.
    to = _header(msg, 'Reply-To') or _header(msg, 'From') or inbound.sender

    # The parsed header, decoded; Cloudflare's subject field can still be RFC 2047-encoded.
    subject = _header(msg, 'Subject') or inbound.subject
    if not re.match(r'^\s*re\s*:', subject, re.I):
        subject = f'Re: {subject}'.strip()

    quoted_from = _header(msg, 'From') or inbound.sender
    date = _header(msg, 'Date') or f'{inbound.received_at:%a, %d %b %Y %H:%M:%S} +0000'
    original = body_text(inbound).strip('\n')
    quote = '\n'.join('>' + (' ' + line if line else '') for line in original.splitlines())

    return {
        'from_name': from_name,
        'from_local': local,
        'to': to,
        'cc': '',
        'subject': subject,
        'body': f'\n\n\nOn {date}, {quoted_from} wrote:\n{quote}\n',
    }


def parse_recipients(value, field, required=False):
    """Comma-separated addresses → ['Name <addr>', …]; raises ValidationError."""
    out = []
    for name, addr in getaddresses([value or '']):
        if not addr:
            continue
        try:
            validate_email(addr)
        except ValidationError:
            raise ValidationError(f'{field}: "{addr}" is not a valid email address.')
        out.append(formataddr((name, addr)))
    if required and not out:
        raise ValidationError(f'{field}: at least one address is required.')
    return out


def from_address(from_name, from_local):
    local = (from_local or '').strip()
    if '@' in local:
        raise ValidationError(f'From: enter only the part before @{settings.REPLY_FROM_DOMAIN}.')
    if not LOCAL_PART_RE.match(local) or len(local) > 64:
        raise ValidationError(f'From: "{local}" is not a valid address before @{settings.REPLY_FROM_DOMAIN}.')
    name = re.sub(r'[\r\n]+', ' ', from_name or '').strip()
    return formataddr((name, f'{local}@{settings.REPLY_FROM_DOMAIN}'))


def send_reply(inbound, *, from_name, from_local, to, cc, subject, body, sent_by=''):
    """Validate, send through Mailgun, and log an OutboundEmail. Raises
    ValidationError for bad input; SMTP errors propagate."""
    sender = from_address(from_name, from_local)
    to_list = parse_recipients(to, 'To', required=True)
    cc_list = parse_recipients(cc, 'Cc')
    subject = re.sub(r'[\r\n]+', ' ', subject or '').strip()

    message_id = make_msgid(domain=settings.REPLY_FROM_DOMAIN)
    headers = {'Message-ID': message_id}
    if inbound.message_id:
        refs = _header(parse_raw(inbound), 'References').split()
        headers['In-Reply-To'] = inbound.message_id
        headers['References'] = ' '.join(refs + [inbound.message_id])

    EmailMessage(
        subject=subject,
        body=body,
        from_email=sender,
        to=to_list,
        cc=cc_list,
        headers=headers,
    ).send()

    return OutboundEmail.objects.create(
        in_reply_to=inbound,
        sent_by=sent_by,
        from_email=sender,
        to=', '.join(to_list),
        cc=', '.join(cc_list),
        subject=subject,
        body=body,
        message_id=message_id,
    )

