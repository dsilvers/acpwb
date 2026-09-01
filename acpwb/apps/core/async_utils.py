"""
Shared helper for deferring non-critical, best-effort work (honeypot
logging, live-stream publishing) off the request's response path.
"""


def spawn(fn, *args, **kwargs):
    """
    Run fn on a background greenlet so the caller never blocks on Redis or
    Postgres I/O. Falls back to running inline under pytest (see
    settings.TESTING) — gunicorn's gevent worker monkey-patches the process
    so a spawned greenlet gets a chance to run, but nothing does that under
    the test runner, so tests need these writes to stay synchronous — or if
    gevent isn't the active worker model (e.g. local `runserver`).
    """
    from django.conf import settings
    if getattr(settings, 'TESTING', False):
        fn(*args, **kwargs)
        return
    try:
        import gevent
        gevent.spawn(fn, *args, **kwargs)
    except Exception:
        fn(*args, **kwargs)
