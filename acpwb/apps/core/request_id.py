"""
acpwb_edge's per-request ID, carried from the X-Request-Id header it sets
(after stripping any client-sent copy) to the rows a request produces —
CrawlerVisit, ArchiveVisit, RequestFingerprint — so they join to
EdgeRequest.request_id, and through it to the connection and client.

RequestStreamMiddleware (outermost) binds the ID for the duration of the
request; queue_crawler_visit/queue_archive_visit stamp it onto their
payloads while still on the request's greenlet, before spawning the write.
"""
import contextvars
import uuid

_current = contextvars.ContextVar('acpwb_request_id', default=None)


def parse_request_id(value):
    """The header value as a canonical UUID string, or None if it isn't one
    (no edge in front, e.g. local runserver)."""
    if not value:
        return None
    try:
        return str(uuid.UUID(value))
    except (ValueError, TypeError, AttributeError):
        return None


def bind(request):
    """Bind request's edge request ID; returns a token for unbind()."""
    return _current.set(parse_request_id(request.META.get('HTTP_X_REQUEST_ID')))


def unbind(token):
    _current.reset(token)


def current_request_id():
    return _current.get()


def stamp_request_id(data: dict) -> dict:
    """data with request_id set from the current request, if it has one and
    data doesn't already."""
    rid = _current.get()
    if rid and not data.get('request_id'):
        data = {**data, 'request_id': rid}
    return data
