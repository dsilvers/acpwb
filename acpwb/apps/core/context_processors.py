import hashlib
import time


def honeypot_context(request):
    """Inject a per-request watermark token used in JSON-LD and prompt injection."""
    token = hashlib.md5(
        f"{request.path}{time.time()}{request.META.get('REMOTE_ADDR', '')}".encode()
    ).hexdigest()[:8]
    return {
        'honeypot_token': token,
    }
