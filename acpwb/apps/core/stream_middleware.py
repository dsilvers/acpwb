import orjson
import time
import logging
from datetime import datetime, timezone

logger = logging.getLogger(__name__)

_redis_client = None
_last_failure = 0.0
_CIRCUIT_BREAKER_COOLDOWN = 30.0  # seconds to wait after a Redis failure before retrying


def _get_redis():
    """Lazily create a synchronous Redis client. Returns None if unavailable."""
    global _redis_client, _last_failure

    if _redis_client is not None:
        return _redis_client

    # Circuit breaker: don't hammer a downed Redis
    if time.monotonic() - _last_failure < _CIRCUIT_BREAKER_COOLDOWN:
        return None

    try:
        import redis as redis_lib
        from django.conf import settings
        url = getattr(settings, 'REDIS_URL', 'redis://redis:6379/0')
        _redis_client = redis_lib.from_url(
            url,
            socket_connect_timeout=1,
            socket_timeout=0.1,
            decode_responses=True,
            max_connections=100,
        )
        return _redis_client
    except Exception:
        _last_failure = time.monotonic()
        return None


class RequestStreamMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        t0 = time.monotonic()
        response = self.get_response(request)
        elapsed_ms = round((time.monotonic() - t0) * 1000)

        # Fire-and-forget: nothing in the response path depends on this, so
        # publish it on a background greenlet instead of blocking the
        # response on a Redis round-trip.
        from apps.core.async_utils import spawn
        spawn(self._safe_publish, request, response, elapsed_ms)

        return response

    def _safe_publish(self, request, response, elapsed_ms):
        try:
            self._publish(request, response, elapsed_ms)
        except Exception:
            pass  # never let streaming break the response

    def _publish(self, request, response, elapsed_ms):
        global _redis_client, _last_failure

        r = _get_redis()
        if r is None:
            return

        ip = self._get_ip(request)
        parts = ip.split('.')
        if len(parts) == 4:
            parts[-1] = 'xxx'
            ip_censored = '.'.join(parts)
        else:
            ip_censored = ip  # IPv6 — pass through

        response_bytes = 0
        if hasattr(response, 'content'):
            try:
                response_bytes = len(response.content)
            except Exception:
                pass

        ua = request.META.get('HTTP_USER_AGENT', '')
        cached = getattr(request, '_bot_classification', None)
        if cached is not None:
            bot_type, bot_group = cached
        else:
            # BotTrackingMiddleware normally computes and caches this per
            # request; fall back to computing it here if it didn't run.
            try:
                from apps.core.bot_classify import bot_type_to_group, classify_ua_or_ip
                bot_type = classify_ua_or_ip(ua, ip)
                bot_group = bot_type_to_group(bot_type)
            except Exception:
                bot_type = ''
                bot_group = ''

        payload = orjson.dumps({
            'ip': ip_censored,
            'host': request.get_host(),
            'path': request.path,
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'response_ms': elapsed_ms,
            'response_bytes': response_bytes,
            'method': request.method,
            'status': response.status_code,
            'user_agent': ua,
            'bot_type': bot_type,
            'bot_group': bot_group,
        })

        try:
            r.publish('request_stream', payload)
        except Exception:
            # Redis went down — reset so next request tries to reconnect after cooldown
            _redis_client = None
            _last_failure = time.monotonic()

        self._queue_fingerprint(request, ip, ua)

    @staticmethod
    def _get_ip(request):
        from apps.core.ip_utils import get_client_ip
        return get_client_ip(request)

    @staticmethod
    def _queue_fingerprint(request, ip, ua):
        """Queue a RequestFingerprint row for every request, not just
        UA-matched bot traffic — residential-proxy traffic is exactly the
        traffic that looks like a normal browser, so it has to be captured
        here (this middleware runs unconditionally) rather than gated behind
        BotTrackingMiddleware's BOT_UA_PATTERNS check.

        Gated by FINGERPRINT_CAPTURE_ENABLED (default off) and
        FINGERPRINT_SAMPLE_RATE, checked before anything touches Redis —
        including the per-IP first-seen key — because at full production
        volume this is one queue entry per request."""
        try:
            from django.conf import settings
            from apps.core.signal_capture import fingerprint_sampled
            if not settings.FINGERPRINT_CAPTURE_ENABLED:
                return
            if not fingerprint_sampled(ip, settings.FINGERPRINT_SAMPLE_RATE):
                return

            from django.utils import timezone
            from apps.core.crawler_queue import queue_fingerprint_signal, check_first_seen_ip
            from apps.core.signal_capture import BROWSER_SIGNAL_HEADERS, browser_headers_present

            present_headers = (
                name for name in BROWSER_SIGNAL_HEADERS
                if request.META.get(f'HTTP_{name.upper().replace("-", "_")}')
            )
            data = {
                'timestamp': timezone.now().isoformat(),
                'ip_address': ip,
                'host': request.get_host()[:253],
                'user_agent': ua[:512] if ua else '',
                'referrer_present': bool(request.META.get('HTTP_REFERER')),
                'first_seen_ip': check_first_seen_ip(ip),
                'client_protocol': request.META.get('HTTP_X_CLIENT_PROTOCOL', '')[:16],
                'tls_protocol': request.META.get('HTTP_X_TLS_PROTOCOL', '')[:16],
                'tls_cipher': request.META.get('HTTP_X_TLS_CIPHER', '')[:64],
                'browser_headers_present': browser_headers_present(present_headers),
            }
            queue_fingerprint_signal(data)
        except Exception:
            pass  # never let signal capture break the response
