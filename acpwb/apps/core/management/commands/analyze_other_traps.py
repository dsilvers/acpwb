"""
Breakdown of CrawlerVisit rows with trap_type='other'.

Shows top paths, user agents, and IPs to identify patterns that should
either get a dedicated trap type or be added to VIEW_LOGGED_PATHS.

Usage:
    python manage.py analyze_other_traps
    python manage.py analyze_other_traps --limit 30
    python manage.py analyze_other_traps --sample   # show 20 recent raw rows
"""
from django.core.management.base import BaseCommand
from django.db import transaction
from django.db.models import Count

from apps.honeypot.models import CrawlerVisit


class Command(BaseCommand):
    help = "Breakdown of CrawlerVisit rows with trap_type='other'."

    def add_arguments(self, parser):
        parser.add_argument('--limit', type=int, default=20, help='Rows per table (default 20)')
        parser.add_argument('--sample', action='store_true', help='Print 20 recent raw rows instead of aggregates')

    def handle(self, *args, **options):
        limit = options['limit']
        qs = CrawlerVisit.objects.filter(trap_type='other')
        total = qs.count()

        self.stdout.write(f"\nTotal 'other' trap rows: {total:,}\n")

        if options['sample']:
            self.stdout.write("=== 20 Most Recent Raw Rows ===")
            for r in qs.order_by('-id')[:20]:
                self.stdout.write(
                    f"  [{r.id}] {r.timestamp:%Y-%m-%d %H:%M}  "
                    f"ip={r.ip_address}  path={r.path!r}  "
                    f"ua={r.user_agent[:80]!r}"
                )
            return

        if not total:
            self.stdout.write("No 'other' rows found.")
            return

        self._table(
            "Top Paths",
            qs.values('path').annotate(c=Count('id')).order_by('-c')[:limit],
            'path', total,
        )

        self._table(
            "Top Path Prefixes (first two segments)",
            None, None, total,
            custom_rows=self._prefix_counts(qs, limit),
            label='prefix',
        )

        self._table(
            "Top User Agents",
            qs.values('user_agent').annotate(c=Count('id')).order_by('-c')[:limit],
            'user_agent', total,
            truncate=80,
        )

        self._table(
            "Top IPs",
            qs.values('ip_address').annotate(c=Count('id')).order_by('-c')[:limit],
            'ip_address', total,
        )

        self._table(
            "Top Bot Types",
            qs.values('bot_type').annotate(c=Count('id')).order_by('-c')[:limit],
            'bot_type', total,
        )

        self._table(
            "Top Hosts",
            qs.exclude(host='').values('host').annotate(c=Count('id')).order_by('-c')[:limit],
            'host', total,
        )

    def _table(self, title, rows, field, total, truncate=None, custom_rows=None, label=None):
        self.stdout.write(f"\n=== {title} ===")
        source = custom_rows if custom_rows is not None else rows
        label = label or field
        for row in source:
            val = row[label] if custom_rows is not None else row[field]
            count = row['c']
            pct = count * 100 / total if total else 0
            if truncate and val and len(val) > truncate:
                val = val[:truncate] + '…'
            self.stdout.write(f"  {count:>8,}  {pct:5.1f}%  {val}")

    def _prefix_counts(self, qs, limit):
        """Group paths by their first two URL segments."""
        from collections import Counter
        counts = Counter()
        with transaction.atomic():  # PgBouncer transaction pooling + server-side cursor
            for path in qs.values_list('path', flat=True).iterator(chunk_size=10000):
                parts = path.strip('/').split('/')
                prefix = '/' + '/'.join(parts[:2]) if len(parts) >= 2 else '/' + parts[0] if parts else '/'
                counts[prefix] += 1
        return [{'prefix': k, 'c': v} for k, v in counts.most_common(limit)]
