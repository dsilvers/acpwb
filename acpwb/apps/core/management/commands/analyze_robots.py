"""
Report which bot types checked robots.txt and which never did.

Uses a single table scan (WHERE path = '/robots.txt' GROUP BY bot_type)
for robots.txt counts, then reads total-per-bot from the pre-aggregated
DashboardStat to avoid a second full scan.

On TimescaleDB, the WHERE path filter is applied per chunk in parallel,
so most rows are skipped early — still a full scan but no worse than any
other single-pass query on the hypertable.

Usage:
    manage.py analyze_robots
    manage.py analyze_robots --min-hits 5    # ignore bots with < 5 total hits
"""
import time

from django.core.management.base import BaseCommand
from django.db.models import Count


class Command(BaseCommand):
    help = "Show which bot types checked robots.txt and which didn't"

    def add_arguments(self, parser):
        parser.add_argument(
            '--min-hits', type=int, default=0,
            help='Exclude bots with fewer than this many total hits (default: 0)',
        )

    def handle(self, *args, **options):
        from apps.core.models import DashboardStat
        from apps.honeypot.models import CrawlerVisit

        min_hits = options['min_hits']

        # ── Total hits per bot type from pre-aggregated stat (no table scan) ──
        try:
            totals = DashboardStat.objects.get(key='crawlers.by_bot_type').value
        except DashboardStat.DoesNotExist:
            totals = {}
            self.stdout.write(
                'Warning: crawlers.by_bot_type stat not found — '
                'total hit counts unavailable. Run precalc_dashboard first.'
            )

        if min_hits:
            totals = {k: v for k, v in totals.items() if v >= min_hits}

        # ── robots.txt hits per bot type (single table scan) ──────────────────
        self.stdout.write('Scanning CrawlerVisit for robots.txt hits ...')
        t0 = time.monotonic()

        robots_rows = (
            CrawlerVisit.objects
            .filter(path='/robots.txt')
            .values('bot_type')
            .annotate(hits=Count('id'))
        )
        robots_hits = {r['bot_type']: r['hits'] for r in robots_rows}

        elapsed = time.monotonic() - t0
        total_robots_requests = sum(robots_hits.values())
        self.stdout.write(
            f'  Done in {elapsed:.1f}s — '
            f'{total_robots_requests:,} robots.txt hits across '
            f'{len(robots_hits)} bot type(s).'
        )

        # ── Split into checked / ignored ───────────────────────────────────────
        all_bots = set(totals.keys()) | set(robots_hits.keys())
        if min_hits:
            all_bots = {b for b in all_bots if totals.get(b, 0) >= min_hits}

        checked = {b for b in all_bots if b in robots_hits}
        ignored = all_bots - checked

        grand_total = sum(totals.values()) if totals else 0

        # ── Section 1: bots that checked robots.txt ────────────────────────────
        self.stdout.write('')
        self.stdout.write(f'CHECKED ROBOTS.TXT ({len(checked)} bot type(s))')
        self.stdout.write('')

        if checked:
            col = max(len(b) for b in checked)
            col = min(max(col, 8), 40)
            header = f'  {"BOT TYPE":<{col}}  {"ROBOTS HITS":>12}  {"TOTAL HITS":>12}  {"% OF BOT":>9}  {"% OF ALL":>9}'
            self.stdout.write(header)
            self.stdout.write('  ' + '-' * (len(header) - 2))

            sorted_checked = sorted(checked, key=lambda b: robots_hits.get(b, 0), reverse=True)
            for bot in sorted_checked:
                rh = robots_hits.get(bot, 0)
                th = totals.get(bot, 0)
                pct_bot = rh / th * 100 if th else 0
                pct_all = rh / grand_total * 100 if grand_total else 0
                display = bot if bot else '(empty)'
                display = display[:col]
                self.stdout.write(
                    f'  {display:<{col}}  {rh:>12,}  {th:>12,}  {pct_bot:>8.1f}%  {pct_all:>8.1f}%'
                )
        else:
            self.stdout.write('  (none)')

        # ── Section 2: bots that never checked robots.txt ──────────────────────
        self.stdout.write('')
        self.stdout.write(f'IGNORED ROBOTS.TXT ({len(ignored)} bot type(s))')
        self.stdout.write('')

        if ignored:
            col2 = max(len(b) for b in ignored)
            col2 = min(max(col2, 8), 40)
            header2 = f'  {"BOT TYPE":<{col2}}  {"TOTAL HITS":>12}  {"% OF ALL":>9}'
            self.stdout.write(header2)
            self.stdout.write('  ' + '-' * (len(header2) - 2))

            sorted_ignored = sorted(ignored, key=lambda b: totals.get(b, 0), reverse=True)
            for bot in sorted_ignored:
                th = totals.get(bot, 0)
                pct_all = th / grand_total * 100 if grand_total else 0
                display = bot if bot else '(empty)'
                display = display[:col2]
                self.stdout.write(f'  {display:<{col2}}  {th:>12,}  {pct_all:>8.1f}%')
        else:
            self.stdout.write('  (none)')

        self.stdout.write('')
