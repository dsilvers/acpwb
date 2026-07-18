"""
Analyze unique URLs served across the lifetime of the project.

Does a full GROUP BY on CrawlerVisit.path. With 800M+ rows and no index
on path this is a sequential scan — expect 30+ minutes on production.
Grab a coffee.

Usage:
    manage.py analyze_paths
    manage.py analyze_paths --top 100
    manage.py analyze_paths --top 0              # show everything
    manage.py analyze_paths --min-hits 50
    manage.py analyze_paths --filter /reports/
    manage.py analyze_paths --include-archives
    manage.py analyze_paths --with-query-strings
    manage.py analyze_paths --trap-type scanner_probe
"""
from django.core.management.base import BaseCommand
from django.db.models import Count


class Command(BaseCommand):
    help = 'Show unique URLs served, sorted by popularity (full table scan)'

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
            '--include-archives', action='store_true',
            help='Also include ArchiveVisit slugs in the results',
        )
        parser.add_argument(
            '--with-query-strings', action='store_true',
            help='Treat path+query_string as distinct URLs (default: path only)',
        )
        parser.add_argument(
            '--trap-type', default='',
            help='Filter by trap_type (e.g. scanner_probe, wp_probe)',
        )

    def handle(self, *args, **options):
        from apps.honeypot.models import ArchiveVisit, CrawlerVisit

        top = options['top']
        min_hits = options['min_hits']
        filter_str = options['filter_str']
        include_archives = options['include_archives']
        with_qs = options['with_query_strings']
        trap_type = options['trap_type']

        # ── CrawlerVisit ───────────────────────────────────────────────────────
        self.stdout.write('Querying CrawlerVisit (full table scan — may take a while) ...')
        cv_qs = CrawlerVisit.objects.all()
        if trap_type:
            cv_qs = cv_qs.filter(trap_type=trap_type)
        if filter_str:
            cv_qs = cv_qs.filter(path__icontains=filter_str)

        if with_qs:
            rows = (
                cv_qs
                .extra(select={'full_path': "CASE WHEN query_string = '' THEN path ELSE path || '?' || query_string END"})
                .values('full_path')
                .annotate(hits=Count('id'))
            )
            path_counts = {r['full_path']: r['hits'] for r in rows}
        else:
            path_counts = {r['path']: r['hits'] for r in cv_qs.values('path').annotate(hits=Count('id'))}

        total_cv_requests = sum(path_counts.values())
        unique_cv_paths = len(path_counts)
        self.stdout.write(f'  Done. {unique_cv_paths:,} unique paths, {total_cv_requests:,} total requests.')

        # ── ArchiveVisit (optional) ────────────────────────────────────────────
        total_av_requests = 0
        unique_av_paths = 0
        if include_archives:
            self.stdout.write('Querying ArchiveVisit ...')
            av_qs = ArchiveVisit.objects.all()
            if filter_str:
                av_qs = av_qs.filter(slug__icontains=filter_str)
            for r in av_qs.values('slug').annotate(hits=Count('id')):
                key = f'/archive/.../{r["slug"]}'
                path_counts[key] = path_counts.get(key, 0) + r['hits']
                total_av_requests += r['hits']
                unique_av_paths += 1
            self.stdout.write(f'  Done. {unique_av_paths:,} unique slugs, {total_av_requests:,} total requests.')

        # ── Filter, sort, truncate ─────────────────────────────────────────────
        if min_hits > 1:
            path_counts = {p: c for p, c in path_counts.items() if c >= min_hits}

        sorted_paths = sorted(path_counts.items(), key=lambda x: -x[1])
        total_shown = len(sorted_paths)
        if top:
            sorted_paths = sorted_paths[:top]

        grand_total = total_cv_requests + total_av_requests

        # ── Output ─────────────────────────────────────────────────────────────
        self.stdout.write('')
        self.stdout.write(f'  Total requests        : {grand_total:,}')
        self.stdout.write(f'  Unique paths (crawler): {unique_cv_paths:,}')
        if include_archives:
            self.stdout.write(f'  Unique slugs (archive): {unique_av_paths:,}')
        if min_hits > 1:
            self.stdout.write(f'  Paths with >= {min_hits} hits : {total_shown:,}')
        if filter_str:
            self.stdout.write(f'  Filter                : "{filter_str}"')
        self.stdout.write('')

        if not sorted_paths:
            self.stdout.write('  No paths matched.')
            return

        label = 'PATH + QUERY STRING' if with_qs else 'PATH'
        header = f'  {"HITS":>10}  {"PCT":>5}  {label}'
        self.stdout.write(header)
        self.stdout.write('  ' + '-' * (len(header) - 2))

        for path, hits in sorted_paths:
            pct = hits / grand_total * 100 if grand_total else 0
            display = path if len(path) <= 100 else path[:97] + '...'
            self.stdout.write(f'  {hits:10,}  {pct:4.1f}%  {display}')

        if top and total_shown > top:
            self.stdout.write(
                f'\n  ... {total_shown - top:,} more paths not shown '
                f'(use --top 0 to see all, or --min-hits to filter)'
            )
        self.stdout.write('')
