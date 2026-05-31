"""
Redis-backed queues for deferred DB writes.

CrawlerVisit:
  Request path: push_crawler_visit() → RPUSH to acpwb:crawler_queue
  Consumer:     pop_crawler_visits() → LPOP key count → bulk_create in DB

ArchiveVisit:
  Request path: push_archive_visit() → RPUSH to acpwb:archive_queue
  Consumer:     pop_archive_visits() → LPOP key count → bulk_create in DB

Falls back gracefully if Redis is unavailable (caller handles the fallback).
"""
import json
import time

_QUEUE_KEY = 'acpwb:crawler_queue'
_ARCHIVE_QUEUE_KEY = 'acpwb:archive_queue'
_CIRCUIT_BREAKER_COOLDOWN = 30.0

_redis_client = None
_last_failure = 0.0


def _get_client():
    global _redis_client, _last_failure

    if _redis_client is not None:
        return _redis_client

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


def _mark_failure():
    global _redis_client, _last_failure
    _redis_client = None
    _last_failure = time.monotonic()


def push_crawler_visit(data: dict) -> bool:
    """
    Serialize `data` and RPUSH it onto the crawler queue.

    Returns True on success, False if Redis is unavailable (caller should
    fall back to a direct DB write).
    """
    r = _get_client()
    if r is None:
        return False
    try:
        pipe = r.pipeline(transaction=False)
        pipe.rpush(_QUEUE_KEY, json.dumps(data))
        # pipe.ltrim(_QUEUE_KEY, -_MAX_QUEUE, -1)
        pipe.execute()
        return True
    except Exception:
        _mark_failure()
        return False


def pop_crawler_visits(count: int = 500) -> list:
    """
    Pop up to `count` items from the left of the queue (FIFO).
    Returns a list of dicts; stops early if the queue is exhausted.
    """
    r = _get_client()
    if r is None:
        return []
    try:
        results = r.lpop(_QUEUE_KEY, count)
        if not results:
            return []
        return [json.loads(raw) for raw in results]
    except Exception:
        _mark_failure()
        return []


def queue_length() -> int:
    """Return the current queue depth, or -1 if Redis is unavailable."""
    r = _get_client()
    if r is None:
        return -1
    try:
        return r.llen(_QUEUE_KEY)
    except Exception:
        _mark_failure()
        return -1


def push_archive_visit(data: dict) -> bool:
    """
    Serialize `data` and RPUSH it onto the archive visit queue.

    Returns True on success, False if Redis is unavailable (caller should
    fall back to a direct DB write).
    """
    r = _get_client()
    if r is None:
        return False
    try:
        r.rpush(_ARCHIVE_QUEUE_KEY, json.dumps(data))
        return True
    except Exception:
        _mark_failure()
        return False


def pop_archive_visits(count: int = 500) -> list:
    """
    Pop up to `count` items from the left of the archive queue (FIFO).
    Returns a list of dicts; stops early if the queue is exhausted.
    """
    r = _get_client()
    if r is None:
        return []
    try:
        results = r.lpop(_ARCHIVE_QUEUE_KEY, count)
        if not results:
            return []
        return [json.loads(raw) for raw in results]
    except Exception:
        _mark_failure()
        return []


def archive_queue_length() -> int:
    """Return the archive queue depth, or -1 if Redis is unavailable."""
    r = _get_client()
    if r is None:
        return -1
    try:
        return r.llen(_ARCHIVE_QUEUE_KEY)
    except Exception:
        _mark_failure()
        return -1
