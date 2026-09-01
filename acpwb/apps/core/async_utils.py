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


def run_in_thread(fn, *args, **kwargs):
    """
    Run fn on gevent's native OS-thread pool and block the current greenlet
    until it completes, returning fn's result (exceptions from fn propagate
    normally). Use for CPU/C-library-bound work — e.g. weasyprint PDF
    rendering — that would otherwise run inline on the event loop and stall
    every other concurrent connection on this worker for its full duration,
    since cooperative scheduling only yields on I/O, not CPU-bound code.
    Falls back to running inline if gevent itself is unavailable.
    """
    try:
        import gevent
    except ImportError:
        return fn(*args, **kwargs)
    return gevent.get_hub().threadpool.spawn(fn, *args, **kwargs).get()
