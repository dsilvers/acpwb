"""
Profile archive/policy trap page rendering.

Hits real URLs through Django's test Client (full middleware stack, same as
production) rather than calling view functions directly, so the reported
numbers match how request time is actually measured elsewhere in this app
(e.g. RequestStreamMiddleware's own per-request timing).

Reports:
  - Wall-clock timing, cold (a unique URL per call, guaranteeing every
    lru_cache'd generator misses — representative of a bot crawling a URL
    for the first time) vs. warm (the same URL repeated — first call is a
    miss, the rest are cache hits, representative of a repeat crawl).
  - A cProfile breakdown AGGREGATED over many calls (not a single sample,
    which is noisy and — for archive specifically — confounded by variant
    selection depending on the slug hash, so a single cold call might land
    on any of 3 different templates/generators). Two views: cumulative time
    (includes callees) and self/"tottime" (excludes callees, so wrapper
    functions like render()/get_response() don't dominate the ranking) —
    both restricted to this app's own code (apps/...) so Django/Jinja2
    framework internals don't drown out what we can actually change.

Usage:
    python manage.py profile_traps --view archive --n 200
    python manage.py profile_traps --view policy --n 200
    python manage.py profile_traps --view both --n 200   (default)
"""
import cProfile
import io
import pstats
import statistics
import time

from django.core.management.base import BaseCommand
from django.test import Client

_APPS_PATTERN = r'\bapps/'


class Command(BaseCommand):
    help = 'Profile archive/policy trap page rendering (wall-clock + cProfile breakdown).'

    def add_arguments(self, parser):
        parser.add_argument('--view', choices=['archive', 'policy', 'both'], default='both')
        parser.add_argument('--n', type=int, default=200, help='Iterations per case (default: 200)')
        parser.add_argument(
            '--profile-n', type=int, default=50,
            help='Iterations to aggregate per cProfile pass (default: 50)',
        )

    def handle(self, *args, **options):
        n = options['n']
        pn = options['profile_n']
        view = options['view']
        client = Client(SERVER_NAME='acpwb.com')  # matches '.acpwb.com' in ALLOWED_HOSTS

        if view in ('archive', 'both'):
            self._profile_case(
                client, 'archive', n, pn,
                cold_url_fn=lambda i: f'/archive/2010/06/15/cold-profile-slug-{i}/',
                warm_url='/archive/2010/06/15/warm-profile-slug/',
            )
        if view in ('policy', 'both'):
            self._profile_case(
                client, 'policy', n, pn,
                cold_url_fn=lambda i: f'/public-policy/2020/05/10/sec/cold-profile-slug-{i}/',
                warm_url='/public-policy/2020/05/10/sec/warm-profile-slug/',
            )

    def _profile_case(self, client, label, n, pn, cold_url_fn, warm_url):
        self.stdout.write(self.style.MIGRATE_HEADING(f'\n=== {label} ==='))

        cold_urls = [cold_url_fn(i) for i in range(n)]
        cold_times = []
        for url in cold_urls:
            t0 = time.perf_counter()
            resp = client.get(url)
            cold_times.append((time.perf_counter() - t0) * 1000)
            if resp.status_code != 200:
                self.stderr.write(f'  WARNING: {url} returned {resp.status_code}')

        warm_times = []
        resp = None
        for _ in range(n):
            t0 = time.perf_counter()
            resp = client.get(warm_url)
            warm_times.append((time.perf_counter() - t0) * 1000)
        if resp is not None and resp.status_code != 200:
            self.stderr.write(f'  WARNING: {warm_url} returned {resp.status_code}')

        self._report('cold (unique URL each call)', cold_times)
        self._report('warm (same URL repeated, first call excluded)', warm_times[1:])

        # Fresh slugs, never touched by the wall-clock loop above, so this
        # cProfile pass is a genuine lru_cache-miss measurement.
        cprofile_cold_urls = [cold_url_fn(f'cprofile-{i}') for i in range(pn)]
        self.stdout.write(f'\n  --- cProfile: {pn} cold calls (aggregated) ---')
        self._cprofile_many(client, cprofile_cold_urls)
        self.stdout.write(f'\n  --- cProfile: {pn} warm calls (aggregated, cache hit) ---')
        self._cprofile_many(client, [warm_url] * pn)

    def _cprofile_many(self, client, urls):
        profiler = cProfile.Profile()
        profiler.enable()
        for url in urls:
            client.get(url)
        profiler.disable()

        self.stdout.write('    [top 12 by cumulative time, this app only]')
        self._print_stats(profiler, 'cumulative', 12, restrict=True)
        self.stdout.write('    [top 12 by self time (tottime), this app only]')
        self._print_stats(profiler, 'tottime', 12, restrict=True)
        self.stdout.write('    [top 25 by cumulative time, everything — for context]')
        self._print_stats(profiler, 'cumulative', 25, restrict=False)

    def _print_stats(self, profiler, sort_by, limit, restrict):
        stream = io.StringIO()
        stats = pstats.Stats(profiler, stream=stream).sort_stats(sort_by)
        if restrict:
            stats.print_stats(_APPS_PATTERN, limit)
        else:
            stats.print_stats(limit)
        # pstats prints a header (call count/time summary) every time; only
        # show the actual table rows after the first blank-line-preceded
        # "ncalls" header to avoid repeating the summary 3x per call.
        lines = stream.getvalue().splitlines()
        for line in lines:
            if line.strip() and not line.startswith(('Random listing order', 'Ordered by')):
                self.stdout.write(f'    {line}')

    def _report(self, label, times_ms):
        if not times_ms:
            self.stdout.write(f'  {label}: no samples')
            return
        self.stdout.write(
            f'  {label}: n={len(times_ms)} '
            f'median={statistics.median(times_ms):.2f}ms '
            f'mean={statistics.mean(times_ms):.2f}ms '
            f'min={min(times_ms):.2f}ms max={max(times_ms):.2f}ms'
        )
