import json
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

        try:
            self._publish(request, response, elapsed_ms)
        except Exception:
            pass  # never let streaming break the response

        return response

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

        payload = json.dumps({
            'ip': ip_censored,
            'host': request.get_host(),
            'path': request.path,
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'response_ms': elapsed_ms,
            'response_bytes': response_bytes,
            'method': request.method,
            'status': response.status_code,
            'user_agent': request.META.get('HTTP_USER_AGENT', ''),
        })

        try:
            r.publish('request_stream', payload)
        except Exception:
            # Redis went down — reset so next request tries to reconnect after cooldown
            _redis_client = None
            _last_failure = time.monotonic()

    @staticmethod
    def _get_ip(request):
        x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
        if x_forwarded:
            return x_forwarded.split(',')[0].strip()
        return request.META.get('REMOTE_ADDR', '0.0.0.0')
