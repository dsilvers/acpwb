import email as email_lib
import hashlib
import hmac
import json
import logging

from django.conf import settings
from django.http import HttpResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_POST

from .models import InboundEmail, HoneypotMatch

logger = logging.getLogger(__name__)


def _verify_mailgun_signature(signing_key, token, timestamp, signature):
    """Verify Mailgun webhook HMAC-SHA256 signature."""
    if not signing_key:
        logger.warning("MAILGUN_WEBHOOK_SIGNING_KEY not configured — skipping verification")
        return True  # Allow in dev when key not set

    value = f"{timestamp}{token}".encode('utf-8')
    expected = hmac.new(
        signing_key.encode('utf-8'),
        value,
        hashlib.sha256,
    ).hexdigest()
    return hmac.compare_digest(expected, signature)


@csrf_exempt
@require_POST
def mailgun_inbound(request):
    # Parse multipart or JSON body
    post = request.POST

    timestamp = post.get('timestamp', '')
    token = post.get('token', '')
    signature = post.get('signature', '')

    if not _verify_mailgun_signature(
        settings.MAILGUN_WEBHOOK_SIGNING_KEY,
        token, timestamp, signature,
    ):
        logger.warning(
            "Mailgun webhook signature verification failed from %s",
            request.META.get('REMOTE_ADDR'),
        )
        return HttpResponse(status=406)

    sender = post.get('sender', '') or post.get('from', '')
    recipient = post.get('recipient', '') or post.get('To', '')
    subject = post.get('subject', '')
    body_plain = post.get('body-plain', '')
    body_html = post.get('body-html', '')
    message_id = post.get('Message-Id', '')

    # Capture full payload for audit
    raw_payload = {k: v for k, v in post.items()}

    inbound = InboundEmail.objects.create(
        sender=sender[:254],
        recipient=recipient[:254],
        subject=subject[:512],
        body_plain=body_plain,
        body_html=body_html,
        raw_payload=raw_payload,
        mailgun_message_id=message_id[:256],
    )

    # Try to match recipient against generated employees
    _match_honeypot(inbound)

    # Must return 200 or Mailgun will retry
    return HttpResponse(status=200)


@csrf_exempt
@require_POST
def pipe_inbound(request):
    secret = getattr(settings, 'PIPE_WEBHOOK_SECRET', '')
    if secret and request.headers.get('X-Webhook-Secret') != secret:
        return HttpResponse(status=403)

    try:
        data = json.loads(request.body)
    except json.JSONDecodeError:
        return HttpResponse(status=400)

    # Parse the raw RFC 2822 email for body text
    raw = data.get('raw', '')
    body_plain = ''
    body_html = ''
    if raw:
        msg = email_lib.message_from_string(raw)
        if msg.is_multipart():
            for part in msg.walk():
                ct = part.get_content_type()
                if ct == 'text/plain' and not body_plain:
                    body_plain = part.get_payload(decode=True).decode(errors='replace')
                elif ct == 'text/html' and not body_html:
                    body_html = part.get_payload(decode=True).decode(errors='replace')
        else:
            body_plain = msg.get_payload(decode=True).decode(errors='replace')

    inbound = InboundEmail.objects.create(
        sender=data.get('sender', '')[:254],
        recipient=data.get('recipient', '')[:254],
        subject=data.get('subject', '')[:512],
        body_plain=body_plain,
        body_html=body_html,
        raw_payload=data,
    )

    _match_honeypot(inbound)
    return HttpResponse(status=200)


def _match_honeypot(inbound_email):
    """Find GeneratedEmployee records matching the recipient address."""
    from apps.people.models import GeneratedEmployee

    recipient = inbound_email.recipient.lower().strip()

    matches = GeneratedEmployee.objects.filter(
        email__iexact=recipient,
    ).select_related('visit')

    if matches.exists():
        for employee in matches:
            HoneypotMatch.objects.create(
                inbound_email=inbound_email,
                generated_employee=employee,
                original_visit=employee.visit,
                match_confidence='exact',
                notes=(
                    f"Recipient {recipient} matched employee {employee.full_name} "
                    f"generated during visit {employee.visit.id} "
                    f"from IP {employee.visit.ip_address} "
                    f"at {employee.visit.timestamp:%Y-%m-%d %H:%M:%S}"
                ),
            )
        logger.info(
            "Honeypot match: %s → %d employee record(s)",
            recipient, matches.count(),
        )
    else:
        # No exact match — log it anyway with no match
        HoneypotMatch.objects.create(
            inbound_email=inbound_email,
            match_confidence='none',
            notes=f"No GeneratedEmployee found for recipient: {recipient}",
        )
        logger.info("No honeypot match found for recipient: %s", recipient)
