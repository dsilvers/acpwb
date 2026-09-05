import threading
from contextlib import contextmanager

_local = threading.local()


class DirectDBRouter:
    """Sends ORM traffic to whatever alias `force_db` last set for this
    thread/greenlet, falling back to Django's normal alias selection
    otherwise. See DATABASES['direct'] in settings for why this exists."""

    def db_for_read(self, model, **hints):
        return getattr(_local, 'alias', None)

    def db_for_write(self, model, **hints):
        return getattr(_local, 'alias', None)


@contextmanager
def force_db(alias):
    previous = getattr(_local, 'alias', None)
    _local.alias = alias
    try:
        yield
    finally:
        _local.alias = previous
