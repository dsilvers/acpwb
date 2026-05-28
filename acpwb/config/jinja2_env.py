from datetime import datetime

from django.contrib.staticfiles.storage import staticfiles_storage
from django.urls import reverse
from jinja2 import Environment


def environment(**options):
    env = Environment(**options)
    env.globals.update({
        'static': staticfiles_storage.url,
        'url': reverse,
        'now_year': datetime.now().year,
        'site_root': '',
    })
    env.filters['capfirst'] = lambda s: (s[:1].upper() + s[1:]) if s else s
    env.filters['truncatechars'] = lambda s, n: (s[:n - 1] + '\u2026') if s and len(s) > n else (s or '')
    return env
