"""
Pre-compute all dashboard stats and warm the Redis cache.

Daily charts are preset-independent (always show last 30/60 days regardless of
the date-range picker), so they are computed ONCE here and passed into each
_compute_* call rather than being recomputed 6× per view.

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
    _daily_chart,
)
from apps.honeypot.models import ArchiveVisit, CrawlerVisit
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.webhooks.models import InboundEmail


class Command(BaseCommand):
    help = 'Pre-compute dashboard stats for all preset date ranges and write to Redis cache.'

    def handle(self, *args, **options):
        run_start = time.monotonic()
        now_str = timezone.now().strftime('%Y-%m-%d %H:%M:%S %Z')
        self.stdout.write(f'precalc_dashboard starting at {now_str}')

        # ── Step 1: Compute preset-independent daily charts ONCE ──────────────
        # These scan the full table and are identical for every preset, so we
        # compute them once here and pass them into each _compute_* call.
        self.stdout.write('\nComputing daily charts (preset-independent)...')
        charts = {}
        chart_specs = [
            ('overview',  lambda: CrawlerVisit.objects.all(),  30, 'timestamp'),
            ('crawlers',  lambda: CrawlerVisit.objects.all(),  60, 'timestamp'),
            ('archive',   lambda: ArchiveVisit.objects.all(),  30, 'timestamp'),
            ('emails',    lambda: InboundEmail.objects.all(),  30, 'received_at'),
            ('people',    lambda: PeoplePageVisit.objects.all(), 30, 'timestamp'),
            ('projects',  lambda: ProjectPageVisit.objects.all(), 30, 'timestamp'),
        ]
        for name, qs_fn, days, field in chart_specs:
            t0 = time.monotonic()
            try:
                charts[name] = _daily_chart(qs_fn(), days=days, field=field)
                # Also warm the per-chart Redis keys used by on-demand view misses
                redis_key = f'dashboard:daily:{name}:{days}'
                cache.set(redis_key, charts[name], _DASH_CACHE_TTL)
                self.stdout.write(f'  chart:{name}:{days}d  ({time.monotonic() - t0:.2f}s)')
            except Exception as exc:
                self.stderr.write(f'  ERR chart:{name}  ({time.monotonic() - t0:.2f}s): {exc}')
                charts[name] = None

        # ── Step 2: Per-preset aggregations ───────────────────────────────────
        # Each _compute_* call now only runs the preset-filtered aggregations
        # (bot breakdown, top IPs, trap counts, etc.) — no full-table scans.
        VIEWS = [
            ('overview', _compute_overview, {'daily':        charts.get('overview')}),
            ('crawlers', _compute_crawlers, {'daily':        charts.get('crawlers')}),
            ('archive',  _compute_archive,  {'daily':        charts.get('archive')}),
            ('emails',   _compute_emails,   {'daily':        charts.get('emails')}),
            ('people',   _compute_people,   {'people_daily': charts.get('people'),
                                             'project_daily': charts.get('projects')}),
        ]

        total_keys = 0
        errors = 0

        for preset in PRECALC_PRESETS:
            dr = _build_date_range(preset)
            self.stdout.write(f'\n  [{preset}]')

            for view_name, compute_fn, kwargs in VIEWS:
                key = f'dashboard:{view_name}:{preset}'
                t0 = time.monotonic()
                try:
                    ctx = compute_fn(dr, **kwargs)
                    cache.set(key, ctx, _DASH_CACHE_TTL)
                    self.stdout.write(f'    OK  {key}  ({time.monotonic() - t0:.2f}s)')
                    total_keys += 1
                except Exception as exc:
                    self.stderr.write(f'    ERR {key}  ({time.monotonic() - t0:.2f}s): {exc}')
                    errors += 1

        total_elapsed = time.monotonic() - run_start
        self.stdout.write(
            f'\nDone: {total_keys} keys written, {errors} errors, '
            f'{total_elapsed:.1f}s total'
        )
