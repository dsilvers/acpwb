"""
Analyze unique URLs served, using the pre-aggregated PathStat table.

Run `manage.py build_path_stats` first (and keep it on cron) to populate it.

Usage:
    manage.py analyze_paths
    manage.py analyze_paths --top 100
    manage.py analyze_paths --top 0              # show everything
    manage.py analyze_paths --min-hits 50
    manage.py analyze_paths --filter /reports/
    manage.py analyze_paths --host archives-2019.acpwb.com
    manage.py analyze_paths --include-archives
"""
from django.core.management.base import BaseCommand


class Command(BaseCommand):
    help = 'Show unique URLs by popularity from PathStat (fast — no table scan)'

    def add_arguments(self, parser):
        parser.add_argument(
            '--top', type=int, default=50,
            help='Number of top paths to show (default: 50, 0 = all)',
        )
        parser.add_argument(
            '--min-hits', type=int, default=1,
            help='Only show paths with at least this many hits (default: 1)',
        )
        parser.add_argument(
            '--filter', dest='filter_str', default='',
            help='Only show paths containing this substring',
        )
        parser.add_argument(
            '--host', dest='host_filter', default='',
            help='Filter by host (e.g. archives-2019.acpwb.com; blank = all hosts)',
        )
        parser.add_argument(
            '--include-archives', action='store_true',
            help='Alias for --host "" (archives are already included by default)',
        )

    def handle(self, *args, **options):
        from django.db.models import Sum

        from apps.honeypot.models import PathStat

        top = options['top']
        min_hits = options['min_hits']
        filter_str = options['filter_str']
        host_filter = options['host_filter']

        qs = PathStat.objects.all()
        if filter_str:
            qs = qs.filter(path__icontains=filter_str)
        if host_filter:
            qs = qs.filter(host=host_filter)
        if min_hits > 1:
            qs = qs.filter(count__gte=min_hits)

        totals = PathStat.objects.aggregate(total=Sum('count'))
        grand_total = totals['total'] or 0
        unique_paths = PathStat.objects.count()
        total_shown = qs.count()

        qs = qs.order_by('-count')
        if top:
            qs = qs[:top]

        self.stdout.write('')
        self.stdout.write(f'  Total requests  : {grand_total:,}')
        self.stdout.write(f'  Unique paths    : {unique_paths:,}')
        if min_hits > 1:
            self.stdout.write(f'  Paths >= {min_hits} hits : {total_shown:,}')
        if filter_str:
            self.stdout.write(f'  Filter          : "{filter_str}"')
        if host_filter:
            self.stdout.write(f'  Host            : {host_filter}')
        self.stdout.write('')

        rows = list(qs)
        if not rows:
            self.stdout.write('  No paths matched.')
            return

        header = f'  {"HITS":>10}  {"PCT":>5}  PATH'
        self.stdout.write(header)
        self.stdout.write('  ' + '-' * (len(header) - 2))

        for stat in rows:
            pct = stat.count / grand_total * 100 if grand_total else 0
            path = stat.path if len(stat.path) <= 100 else stat.path[:97] + '...'
            prefix = f'[{stat.host}]' if stat.host else ''
            display = f'{prefix}{path}'
            self.stdout.write(f'  {stat.count:10,}  {pct:4.1f}%  {display}')

        if top and total_shown > top:
            self.stdout.write(
                f'\n  ... {total_shown - top:,} more paths not shown '
                f'(use --top 0 to see all, or --min-hits to filter)'
            )
        self.stdout.write('')
