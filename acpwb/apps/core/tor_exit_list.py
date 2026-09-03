"""
In-process cache over the flat-file Tor exit-node list (see
management/commands/refresh_tor_exit_list.py for how the file gets there).

Cheap to call repeatedly during a long enrich_ip_intelligence run: the file
is only re-read when its mtime changes, not on every is_tor_exit() call.
"""
import os
from django.conf import settings

_cache = {'mtime': None, 'ips': frozenset()}


def is_tor_exit(ip_str: str) -> bool:
    _reload_if_stale()
    return ip_str in _cache['ips']


def _reload_if_stale():
    path = settings.TOR_EXIT_LIST_PATH
    try:
        mtime = os.path.getmtime(path)
    except OSError:
        return
    if mtime != _cache['mtime']:
        with open(path) as f:
            _cache['ips'] = frozenset(
                line.strip() for line in f if line.strip() and not line.startswith('#')
            )
        _cache['mtime'] = mtime
