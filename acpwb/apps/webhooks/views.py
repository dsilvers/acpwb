import base64
import email as email_lib
import hashlib
import hmac
import json
import logging

from django.conf import settings
from django.http import HttpResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_POST

from .models import CallLog, InboundEmail, HoneypotMatch, VoicemailRecording

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


# ── Twilio ─────────────────────────────────────────────────────────────────────

def _verify_twilio_signature(auth_token, request):
    """
    Verify X-Twilio-Signature using HMAC-SHA1.
    Twilio signs: url + sorted(key+value for each POST param concatenated).
    Returns True if valid, or True with a warning if auth_token not configured (dev bypass).
    """
    if not auth_token:
        logger.warning("TWILIO_AUTH_TOKEN not configured — skipping Twilio signature verification")
        return True

    signature = request.headers.get('X-Twilio-Signature', '')
    url = request.build_absolute_uri()
    s = url + ''.join(k + request.POST[k] for k in sorted(request.POST))
    expected = base64.b64encode(
        hmac.new(auth_token.encode('utf-8'), s.encode('utf-8'), hashlib.sha1).digest()
    ).decode('utf-8')
    return hmac.compare_digest(expected, signature)


@csrf_exempt
@require_POST
def twilio_recording(request):
    """Receives recording-complete callback from Twilio Studio."""
    if not _verify_twilio_signature(settings.TWILIO_AUTH_TOKEN, request):
        logger.warning(
            "Twilio recording webhook signature verification failed from %s",
            request.META.get('REMOTE_ADDR'),
        )
        return HttpResponse(status=403)

    post = request.POST
    call_sid      = post.get('CallSid', '')
    recording_sid = post.get('RecordingSid', '')
    recording_url = post.get('RecordingUrl', '')
    duration_str  = post.get('RecordingDuration', '0')
    caller_number = post.get('From', '')

    if not call_sid or not recording_sid:
        logger.warning("Twilio recording webhook missing required fields")
        return HttpResponse(status=400)

    VoicemailRecording.objects.update_or_create(
        call_sid=call_sid,
        defaults={
            'recording_sid':      recording_sid,
            'recording_url':      recording_url,
            'recording_duration': int(duration_str) if duration_str.isdigit() else 0,
            'caller_number':      caller_number,
            'raw_payload':        {k: v for k, v in post.items()},
        },
    )

    return HttpResponse(status=204)


@csrf_exempt
@require_POST
def twilio_transcription(request):
    """Receives transcription-complete callback from Twilio."""
    if not _verify_twilio_signature(settings.TWILIO_AUTH_TOKEN, request):
        logger.warning(
            "Twilio transcription webhook signature verification failed from %s",
            request.META.get('REMOTE_ADDR'),
        )
        return HttpResponse(status=403)

    post = request.POST
    recording_sid        = post.get('RecordingSid', '')
    transcription_text   = post.get('TranscriptionText', '')
    transcription_status = post.get('TranscriptionStatus', 'failed')

    if not recording_sid:
        return HttpResponse(status=400)

    if transcription_status not in ('completed', 'failed'):
        transcription_status = 'failed'

    updated = VoicemailRecording.objects.filter(recording_sid=recording_sid).update(
        transcription_status=transcription_status,
        transcription_text=transcription_text or None,
    )

    if not updated:
        logger.warning(
            "Twilio transcription arrived for unknown RecordingSid: %s", recording_sid
        )

    return HttpResponse(status=204)


@csrf_exempt
@require_POST
def twilio_call_status(request):
    """Receives call status callback from Twilio (fired on every terminal call state)."""
    if not _verify_twilio_signature(settings.TWILIO_AUTH_TOKEN, request):
        logger.warning(
            "Twilio call-status webhook signature verification failed from %s",
            request.META.get('REMOTE_ADDR'),
        )
        return HttpResponse(status=403)

    post = request.POST
    call_sid      = post.get('CallSid', '')
    call_status   = post.get('CallStatus', '')
    caller_number = post.get('From', '')
    duration_str  = post.get('CallDuration', '0')

    if not call_sid or not call_status:
        return HttpResponse(status=400)

    # Only persist terminal states — Twilio fires intermediate events too (ringing, in-progress)
    terminal = {'completed', 'busy', 'no-answer', 'failed', 'canceled'}
    if call_status not in terminal:
        return HttpResponse(status=204)

    CallLog.objects.update_or_create(
        call_sid=call_sid,
        defaults={
            'caller_number': caller_number,
            'call_status':   call_status,
            'call_duration': int(duration_str) if duration_str.isdigit() else 0,
            'raw_payload':   {k: v for k, v in post.items()},
        },
    )

    return HttpResponse(status=204)
