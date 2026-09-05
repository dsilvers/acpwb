"""
Redis-backed queues for deferred DB writes.

CrawlerVisit:
  Request path: push_crawler_visit() → RPUSH to acpwb:crawler_queue
  Consumer:     pop_crawler_visits() → LMOVE into a per-batch processing key
                → bulk_create in DB → finalize_batch() deletes that key

ArchiveVisit:
  Request path: push_archive_visit() → RPUSH to acpwb:archive_queue
  Consumer:     pop_archive_visits() → same LMOVE/finalize_batch pattern

Reliable-queue design: pop_*_visits() never destructively removes an item
from the main queue — it LMOVEs the batch into a freshly-named
"<queue>:processing:<uuid>" list, which is only deleted (via finalize_batch)
after the batch is confirmed durably written to Postgres. If the consumer
process dies anywhere in between (crash, OOM-kill, connection loss), that
named key survives with exactly that batch's items and nothing else.
recover_*_visits() finds any such leftover keys from a prior run and returns
them the same way, so the caller can retry the DB insert. A retried insert
is safe to repeat because every queued item carries an idempotency_key
(minted once, at push time) backed by a unique DB constraint — re-inserting
an already-committed batch is a harmless no-op, not a duplicate.

Falls back gracefully if Redis is unavailable (caller handles the fallback).
"""
import json
import time
import uuid

_QUEUE_KEY = 'acpwb:crawler_queue'
_ARCHIVE_QUEUE_KEY = 'acpwb:archive_queue'
_CIRCUIT_BREAKER_COOLDOWN = 30.0

_redis_client = None
_last_failure = 0.0

# Consumer (drain command) side gets its own client, separate from the
# request-path one above. The request path is fire-and-forget off a spawned
# greenlet — a tight socket_timeout and long cooldown are right there, since
# it should fail fast and stay quiet rather than pile up retries under load.
# The drain commands are long-lived (up to ~55s), single-threaded bulk
# consumers that can afford to wait longer for a reply, and a single slow
# call shouldn't sideline the rest of that run for 30s.
_CONSUMER_CIRCUIT_BREAKER_COOLDOWN = 3.0

_consumer_redis_client = None
_consumer_last_failure = 0.0


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
            max_connections=100,
        )
        return _redis_client
    except Exception:
        _last_failure = time.monotonic()
        return None


def _mark_failure():
    global _redis_client, _last_failure
    _redis_client = None
    _last_failure = time.monotonic()


def _get_consumer_client():
    global _consumer_redis_client, _consumer_last_failure

    if _consumer_redis_client is not None:
        return _consumer_redis_client

    if time.monotonic() - _consumer_last_failure < _CONSUMER_CIRCUIT_BREAKER_COOLDOWN:
        return None

    try:
        import redis as redis_lib
        from django.conf import settings
        url = getattr(settings, 'REDIS_URL', 'redis://redis:6379/0')
        _consumer_redis_client = redis_lib.from_url(
            url,
            socket_connect_timeout=1,
            socket_timeout=2.0,
            decode_responses=True,
            max_connections=10,
        )
        return _consumer_redis_client
    except Exception:
        _consumer_last_failure = time.monotonic()
        return None


def _mark_consumer_failure():
    global _consumer_redis_client, _consumer_last_failure
    _consumer_redis_client = None
    _consumer_last_failure = time.monotonic()


def push_crawler_visit(data: dict) -> bool:
    """
    Serialize `data` and RPUSH it onto the crawler queue.

    Mints a fresh idempotency_key for this item (kept only in the local
    payload copy, not written back into `data`) so the consumer side can
    safely retry an insert after a crash without risking a duplicate.

    Returns True on success, False if Redis is unavailable (caller should
    fall back to a direct DB write).
    """
    r = _get_client()
    if r is None:
        return False
    payload = dict(data)
    payload.setdefault('idempotency_key', str(uuid.uuid4()))
    try:
        pipe = r.pipeline(transaction=False)
        pipe.rpush(_QUEUE_KEY, json.dumps(payload))
        pipe.execute()
        return True
    except Exception:
        _mark_failure()
        return False


def push_archive_visit(data: dict) -> bool:
    """
    Serialize `data` and RPUSH it onto the archive visit queue.

    See push_crawler_visit() re: idempotency_key.

    Returns True on success, False if Redis is unavailable (caller should
    fall back to a direct DB write).
    """
    r = _get_client()
    if r is None:
        return False
    payload = dict(data)
    payload.setdefault('idempotency_key', str(uuid.uuid4()))
    try:
        r.rpush(_ARCHIVE_QUEUE_KEY, json.dumps(payload))
        return True
    except Exception:
        _mark_failure()
        return False


def _pop_visits(queue_key: str, count: int):
    """
    Shared implementation behind pop_crawler_visits()/pop_archive_visits().

    LMOVEs up to `count` items from `queue_key` into a freshly-named
    per-batch processing key, one LMOVE per item pipelined into a single
    round trip. Returns (items, batch_key); batch_key is None if there was
    nothing to move or Redis was unavailable, in which case there is no
    batch to finalize.
    """
    r = _get_consumer_client()
    if r is None:
        return [], None

    batch_key = f'{queue_key}:processing:{uuid.uuid4().hex}'
    try:
        pipe = r.pipeline(transaction=False)
        for _ in range(count):
            pipe.lmove(queue_key, batch_key, 'LEFT', 'RIGHT')
        results = pipe.execute()
    except Exception:
        _mark_consumer_failure()
        return [], None

    raw_items = [item for item in results if item is not None]
    if not raw_items:
        return [], None

    items = []
    for raw in raw_items:
        try:
            items.append(json.loads(raw))
        except Exception:
            pass  # malformed entry — still physically in batch_key, cleaned
            # up by finalize_batch() same as a successfully-parsed one.
    return items, batch_key


def pop_crawler_visits(count: int = 500):
    """
    Move up to `count` items from the crawler queue into a fresh per-batch
    processing key (never a destructive pop). Returns (items, batch_key).

    Caller must call finalize_batch(batch_key) once the batch is durably
    written to the DB — until then, the items remain recoverable via
    recover_crawler_visits().
    """
    return _pop_visits(_QUEUE_KEY, count)


def pop_archive_visits(count: int = 500):
    """Same contract as pop_crawler_visits(), for the archive queue."""
    return _pop_visits(_ARCHIVE_QUEUE_KEY, count)


def finalize_batch(batch_key) -> None:
    """
    Mark a per-batch processing key fully and durably processed by deleting
    it. Call this ONLY after the batch's records are confirmed committed to
    the DB — shared by both queues, since batch_key already encodes which
    queue it came from.

    A no-op if batch_key is falsy (e.g. pop_*_visits() found nothing).
    Failing to delete isn't fatal: recover_*_visits() will pick the key back
    up and safely re-attempt the now-idempotent insert on the next call.
    """
    if not batch_key:
        return
    r = _get_consumer_client()
    if r is None:
        return
    try:
        r.delete(batch_key)
    except Exception:
        _mark_consumer_failure()


def _recover_visits(queue_key: str):
    """
    Find any "<queue_key>:processing:*" keys left behind by a drain run that
    died before calling finalize_batch() (crash, OOM-kill, connection loss),
    and return them as (items, batch_key) pairs — same shape as
    _pop_visits() — so the caller can retry the insert and finalize each one.

    Safe to call whether or not there's anything to recover.
    """
    r = _get_consumer_client()
    if r is None:
        return []

    batches = []
    try:
        for key in r.scan_iter(match=f'{queue_key}:processing:*', count=100):
            raw_items = r.lrange(key, 0, -1)
            items = []
            for raw in raw_items:
                try:
                    items.append(json.loads(raw))
                except Exception:
                    pass
            batches.append((items, key))
    except Exception:
        _mark_consumer_failure()
    return batches


def recover_crawler_visits():
    """Recover any orphaned crawler-queue batches from a prior crashed run."""
    return _recover_visits(_QUEUE_KEY)


def recover_archive_visits():
    """Recover any orphaned archive-queue batches from a prior crashed run."""
    return _recover_visits(_ARCHIVE_QUEUE_KEY)


def _write_crawler_visit(data: dict):
    if not push_crawler_visit(data):
        try:
            from apps.honeypot.models import CrawlerVisit
            CrawlerVisit.objects.create(**data)
        except Exception:
            pass


def queue_crawler_visit(data: dict) -> None:
    """
    Fire-and-forget: RPUSH `data` onto the crawler queue, falling back to a
    direct DB write if Redis is unavailable — same reliability as calling
    push_crawler_visit() directly, but off the request's critical path.
    """
    from apps.core.async_utils import spawn
    spawn(_write_crawler_visit, data)


def _write_archive_visit(data: dict):
    if not push_archive_visit(data):
        try:
            from apps.honeypot.models import ArchiveVisit
            ArchiveVisit.objects.create(**data)
        except Exception:
            pass


def queue_archive_visit(data: dict) -> None:
    """
    Fire-and-forget: RPUSH `data` onto the archive queue, falling back to a
    direct DB write if Redis is unavailable — same reliability as calling
    push_archive_visit() directly, but off the request's critical path.
    """
    from apps.core.async_utils import spawn
    spawn(_write_archive_visit, data)


def queue_length() -> int:
    """Return the current queue depth, or -1 if Redis is unavailable."""
    r = _get_consumer_client()
    if r is None:
        return -1
    try:
        return r.llen(_QUEUE_KEY)
    except Exception:
        _mark_consumer_failure()
        return -1


def archive_queue_length() -> int:
    """Return the archive queue depth, or -1 if Redis is unavailable."""
    r = _get_consumer_client()
    if r is None:
        return -1
    try:
        return r.llen(_ARCHIVE_QUEUE_KEY)
    except Exception:
        _mark_consumer_failure()
        return -1
