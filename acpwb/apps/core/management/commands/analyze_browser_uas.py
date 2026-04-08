"""
Breakdown of CrawlerVisit rows classified as 'Other / Browser' in the last 24 hours.

Shows top user agent strings, paths, IPs, hosts, and trap types to identify
real browsers, misclassified bots, or new bot patterns worth adding to BOT_PATTERNS.

Usage:
    python manage.py analyze_browser_uas
    python manage.py analyze_browser_uas --limit 30
    python manage.py analyze_browser_uas --hours 48
    python manage.py analyze_browser_uas --sample       # print 20 recent raw rows
    python manage.py analyze_browser_uas --ua-sample    # group by UA prefix, show sample rows per group
"""
from collections import Counter, defaultdict
from datetime import timedelta

from django.core.management.base import BaseCommand
from django.db.models import Count
from django.utils import timezone

from apps.core.bot_classify import classify_ua_or_ip
from apps.honeypot.models import CrawlerVisit


class Command(BaseCommand):
    help = "Breakdown of 'Other / Browser' CrawlerVisit rows from the last N hours."

    def add_arguments(self, parser):
        parser.add_argument('--hours',     type=int, default=24,  help='Lookback window in hours (default 24)')
        parser.add_argument('--limit',     type=int, default=25,  help='Rows per table (default 25)')
        parser.add_argument('--sample',    action='store_true',   help='Print 20 recent raw rows instead of aggregates')
        parser.add_argument('--ua-sample', action='store_true',   help='Group by UA token, show one example row each')

    def handle(self, *args, **options):
        hours  = options['hours']
        limit  = options['limit']
        since  = timezone.now() - timedelta(hours=hours)

        qs = CrawlerVisit.objects.filter(
            bot_type='Other / Browser',
            timestamp__gte=since,
        )
        total = qs.count()

        self.stdout.write(
            self.style.SUCCESS(f"\n'Other / Browser' visits in the last {hours}h: {total:,}\n")
        )

        if not total:
            self.stdout.write("Nothing to report.")
            return

        if options['sample']:
            self.stdout.write("=== 20 Most Recent Raw Rows ===")
            for r in qs.order_by('-id')[:20]:
                self.stdout.write(
                    f"  [{r.id}] {r.timestamp:%Y-%m-%d %H:%M}  "
                    f"trap={r.trap_type:<15}  ip={r.ip_address:<16}  "
                    f"ua={r.user_agent[:100]!r}"
                )
            return

        if options['ua_sample']:
            self._ua_sample(qs, limit)
            return

        # ── Standard aggregate tables ─────────────────────────────────────────

        # Top full UAs
        self._table(
            "Top User Agents (full string)",
            qs.values('user_agent').annotate(c=Count('id')).order_by('-c')[:limit],
            'user_agent', total, truncate=120,
        )

        # Top UA "tokens" — first meaningful word before the first space or /
        self._ua_token_table(qs, limit, total)

        # Top paths
        self._table(
            "Top Paths",
            qs.values('path').annotate(c=Count('id')).order_by('-c')[:limit],
            'path', total, truncate=100,
        )

        # Top path prefixes (first 2 segments)
        self._table(
            "Top Path Prefixes (first 2 segments)",
            None, None, total,
            custom_rows=self._prefix_counts(qs, limit), label='prefix',
        )

        # Top IPs
        self._table(
            "Top IPs",
            qs.values('ip_address').annotate(c=Count('id')).order_by('-c')[:limit],
            'ip_address', total,
        )

        # Subnet breakdowns
        self._subnet_table(qs, limit, total, prefix_len=24, title="Top /24 Subnets")
        self._subnet_table(qs, limit, total, prefix_len=16, title="Top /16 Subnets")

        # Trap type breakdown
        self._table(
            "By Trap Type",
            qs.values('trap_type').annotate(c=Count('id')).order_by('-c'),
            'trap_type', total,
        )

        # By host
        self._table(
            "By Host",
            qs.exclude(host='').values('host').annotate(c=Count('id')).order_by('-c')[:limit],
            'host', total,
        )

        # Reclassification preview — run current classify_ua against DB rows
        self._reclassify_preview(qs, limit, total)

    # ── Helpers ───────────────────────────────────────────────────────────────

    def _table(self, title, rows, field, total, truncate=None, custom_rows=None, label=None):
        self.stdout.write(f"\n=== {title} ===")
        source = custom_rows if custom_rows is not None else rows
        label  = label or field
        for row in source:
            val   = row[label] if custom_rows is not None else row[field]
            count = row['c']
            pct   = count * 100 / total if total else 0
            if truncate and val and len(str(val)) > truncate:
                val = str(val)[:truncate] + '…'
            if not val:
                val = '(empty)'
            bar = '█' * min(int(pct / 2), 40)
            self.stdout.write(f"  {count:>8,}  {pct:5.1f}%  {bar:<40}  {val}")

    def _ua_token_table(self, qs, limit, total):
        """Group by the first slash-delimited token of the UA string."""
        counts = Counter()
        for ua in qs.values_list('user_agent', flat=True).iterator(chunk_size=5000):
            token = (ua or '').split('/')[0].split(' ')[0].strip() or '(empty)'
            counts[token] += 1
        self.stdout.write("\n=== Top UA Tokens (first word/product before '/' or space) ===")
        for token, count in counts.most_common(limit):
            pct = count * 100 / total if total else 0
            bar = '█' * min(int(pct / 2), 40)
            self.stdout.write(f"  {count:>8,}  {pct:5.1f}%  {bar:<40}  {token}")

    def _ua_sample(self, qs, limit):
        """Group by UA token; for each group show count + one example row."""
        groups = defaultdict(list)
        for r in qs.order_by('-id').values('id', 'user_agent', 'ip_address', 'path', 'trap_type', 'timestamp')[:50000]:
            token = (r['user_agent'] or '').split('/')[0].split(' ')[0].strip() or '(empty)'
            groups[token].append(r)

        ranked = sorted(groups.items(), key=lambda kv: len(kv[1]), reverse=True)
        self.stdout.write(f"\n=== UA Token Groups ({len(ranked)} distinct tokens) ===")
        for token, rows in ranked[:limit]:
            example = rows[0]
            self.stdout.write(
                f"\n  {len(rows):>6,}x  {token}\n"
                f"          example: [{example['id']}] {example['timestamp']:%Y-%m-%d %H:%M}  "
                f"ip={example['ip_address']}  trap={example['trap_type']}\n"
                f"          path={example['path']!r}\n"
                f"          ua={example['user_agent'][:120]!r}"
            )

    def _subnet_table(self, qs, limit, total, prefix_len, title):
        """Group IPs by their first N bits (i.e. /prefix_len subnet) and count."""
        import ipaddress
        counts = Counter()
        for ip in qs.values_list('ip_address', flat=True).iterator(chunk_size=10000):
            try:
                net = ipaddress.ip_interface(f"{ip}/{prefix_len}").network
                counts[str(net)] += 1
            except ValueError:
                counts['(invalid)'] += 1
        self.stdout.write(f"\n=== {title} ===")
        for subnet, count in counts.most_common(limit):
            pct = count * 100 / total if total else 0
            bar = '█' * min(int(pct / 2), 40)
            self.stdout.write(f"  {count:>8,}  {pct:5.1f}%  {bar:<40}  {subnet}")

    def _prefix_counts(self, qs, limit):
        counts = Counter()
        for path in qs.values_list('path', flat=True).iterator(chunk_size=10000):
            parts  = path.strip('/').split('/')
            prefix = '/' + '/'.join(parts[:2]) if len(parts) >= 2 else '/' + parts[0] if parts else '/'
            counts[prefix] += 1
        return [{'prefix': k, 'c': v} for k, v in counts.most_common(limit)]

    def _reclassify_preview(self, qs, limit, total):
        """
        Run classify_ua_or_ip() against stored UA + IP pairs.
        Shows what would change if BOT_PATTERNS / IP_BOT_RANGES were applied now —
        useful after adding new patterns before running backfill_bot_types.
        """
        new_labels = Counter()
        still_other = 0
        for ua, ip in qs.values_list('user_agent', 'ip_address').iterator(chunk_size=5000):
            result = classify_ua_or_ip(ua, ip or '')
            if result == 'Other / Browser':
                still_other += 1
            else:
                new_labels[result] += 1

        newly_matched = sum(new_labels.values())
        if not newly_matched:
            self.stdout.write(
                "\n=== Reclassification Preview ===\n"
                "  No rows would be reclassified by current BOT_PATTERNS or IP_BOT_RANGES.\n"
                "  All are genuine 'Other / Browser'.\n"
                "  Run --ua-sample or inspect top UAs above to find new patterns to add."
            )
            return

        self.stdout.write(
            f"\n=== Reclassification Preview (UA + IP ranges) ===\n"
            f"  {newly_matched:,} rows ({newly_matched*100/total:.1f}%) would be reclassified.\n"
            f"  Run: python manage.py backfill_bot_types --reclassify\n"
        )
        for label, count in new_labels.most_common(limit):
            pct = count * 100 / total
            self.stdout.write(f"  {count:>8,}  {pct:5.1f}%  → {label}")
        self.stdout.write(f"  {still_other:>8,}  {still_other*100/total:5.1f}%  → (still Other / Browser)")
