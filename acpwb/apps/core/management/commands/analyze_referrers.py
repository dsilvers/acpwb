"""
Show top referrers logged in CrawlerVisit.referrer.

Skips blank referrers. Single pass with WHERE referrer != '' to filter
early — on TimescaleDB this is applied per chunk so empty rows are
discarded before aggregation.

Usage:
    manage.py analyze_referrers
    manage.py analyze_referrers --top 100
    manage.py analyze_referrers --top 0        # all
    manage.py analyze_referrers --min-hits 10
    manage.py analyze_referrers --filter google
"""
import time

from django.core.management.base import BaseCommand
from django.db.models import Count


class Command(BaseCommand):
    help = 'Show top referrers from CrawlerVisit (full table scan)'

    def add_arguments(self, parser):
        parser.add_argument(
            '--top', type=int, default=50,
            help='Number of top referrers to show (default: 50, 0 = all)',
        )
        parser.add_argument(
            '--min-hits', type=int, default=1,
            help='Only show referrers with at least this many hits (default: 1)',
        )
        parser.add_argument(
            '--filter', dest='filter_str', default='',
            help='Only show referrers containing this substring',
        )

    def handle(self, *args, **options):
        from apps.honeypot.models import CrawlerVisit

        top = options['top']
        min_hits = options['min_hits']
        filter_str = options['filter_str']

        self.stdout.write('Scanning CrawlerVisit for referrers ...')
        t0 = time.monotonic()

        qs = CrawlerVisit.objects.exclude(referrer='')
        if filter_str:
            qs = qs.filter(referrer__icontains=filter_str)

        rows = qs.values('referrer').annotate(hits=Count('id'))
        referrer_counts = {r['referrer']: r['hits'] for r in rows}

        elapsed = time.monotonic() - t0
        self.stdout.write(
            f'  Done in {elapsed:.1f}s — '
            f'{len(referrer_counts):,} unique referrers, '
            f'{sum(referrer_counts.values()):,} total hits.'
        )

        if min_hits > 1:
            referrer_counts = {r: c for r, c in referrer_counts.items() if c >= min_hits}

        sorted_refs = sorted(referrer_counts.items(), key=lambda x: -x[1])
        total_shown = len(sorted_refs)
        grand_total = sum(referrer_counts.values())

        if top:
            sorted_refs = sorted_refs[:top]

        self.stdout.write('')
        self.stdout.write(f'  Unique referrers : {total_shown:,}')
        self.stdout.write(f'  Total hits       : {grand_total:,}')
        if filter_str:
            self.stdout.write(f'  Filter           : "{filter_str}"')
        self.stdout.write('')

        if not sorted_refs:
            self.stdout.write('  No referrers found.')
            return

        header = f'  {"HITS":>8}  {"PCT":>5}  REFERRER'
        self.stdout.write(header)
        self.stdout.write('  ' + '-' * (len(header) - 2))

        for referrer, hits in sorted_refs:
            pct = hits / grand_total * 100 if grand_total else 0
            display = referrer if len(referrer) <= 100 else referrer[:97] + '...'
            self.stdout.write(f'  {hits:8,}  {pct:4.1f}%  {display}')

        if top and total_shown > top:
            self.stdout.write(
                f'\n  ... {total_shown - top:,} more referrers not shown '
                f'(use --top 0 to see all, or --min-hits to filter)'
            )
        self.stdout.write('')
