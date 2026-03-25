"""
Pre-compute all dashboard stats and warm the Redis cache.

Run this on a schedule (every 30 minutes recommended) so the dashboard always
serves instantly from cache rather than running expensive aggregations live.

Usage:
    python manage.py precalc_dashboard

Cron (host crontab):
    */30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web \\
        python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
"""
import time

from django.core.cache import cache
from django.core.management.base import BaseCommand
from django.utils import timezone

from apps.core.dashboard_views import (
    PRECALC_PRESETS,
    _DASH_CACHE_TTL,
    _build_date_range,
    _compute_archive,
    _compute_crawlers,
    _compute_emails,
    _compute_overview,
    _compute_people,
)

VIEWS = [
    ('overview', _compute_overview),
    ('crawlers', _compute_crawlers),
    ('archive',  _compute_archive),
    ('emails',   _compute_emails),
    ('people',   _compute_people),
]


class Command(BaseCommand):
    help = 'Pre-compute dashboard stats for all preset date ranges and write to Redis cache.'

    def handle(self, *args, **options):
        run_start = time.monotonic()
        now_str = timezone.now().strftime('%Y-%m-%d %H:%M:%S %Z')
        self.stdout.write(f'precalc_dashboard starting at {now_str}')

        total_keys = 0
        errors = 0

        for preset in PRECALC_PRESETS:
            dr = _build_date_range(preset)
            self.stdout.write(f'\n  [{preset}]')

            for view_name, compute_fn in VIEWS:
                key = f'dashboard:{view_name}:{preset}'
                t0 = time.monotonic()
                try:
                    ctx = compute_fn(dr)
                    cache.set(key, ctx, _DASH_CACHE_TTL)
                    elapsed = time.monotonic() - t0
                    self.stdout.write(f'    OK  {key}  ({elapsed:.2f}s)')
                    total_keys += 1
                except Exception as exc:
                    elapsed = time.monotonic() - t0
                    self.stderr.write(f'    ERR {key}  ({elapsed:.2f}s): {exc}')
                    errors += 1

        total_elapsed = time.monotonic() - run_start
        self.stdout.write(
            f'\nDone: {total_keys} keys written, {errors} errors, '
            f'{total_elapsed:.1f}s total'
        )
