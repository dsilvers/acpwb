"""
Report on IPIntelligence — country/ASN/hosting/Tor breakdowns for the IPs
CrawlerVisit has seen. Only ever queries IPIntelligence (small, well-indexed)
except for a handful of bounded point-lookups against CrawlerVisit for the
top outliers section — never a bulk scan of CrawlerVisit itself.

Usage:
    python manage.py ip_intelligence_report
    python manage.py ip_intelligence_report --top 10
    python manage.py ip_intelligence_report --since 2026-09-01
    python manage.py ip_intelligence_report --export-csv /tmp/ip-report
"""
import csv
import os

from django.core.management.base import BaseCommand
from django.db.models import Count, Sum

from apps.honeypot.models import CrawlerVisit, IPIntelligence


class Command(BaseCommand):
    help = 'Report country/ASN/hosting/Tor breakdowns from IPIntelligence.'

    def add_arguments(self, parser):
        parser.add_argument('--top', type=int, default=25, help='Rows per top-N table (default: 25)')
        parser.add_argument('--since', type=str, default=None, metavar='YYYY-MM-DD', help='Only IPs with last_seen on/after this date')
        parser.add_argument('--export-csv', type=str, default=None, metavar='DIR', help='Also write country/ASN tables as CSV into this directory')

    def handle(self, *args, **options):
        top = options['top']
        qs = IPIntelligence.objects.all()
        if options['since']:
            qs = qs.filter(last_seen__date__gte=options['since'])

        self._coverage(qs)
        self._top_countries(qs, top, options['export_csv'])
        self._top_asns(qs, top, options['export_csv'])
        self._hosting_split(qs)
        self._tor_crosstab(qs)
        self._country_hosting_ratio(qs, top)
        self._top_outliers(qs, top)

    def _coverage(self, qs):
        total = qs.count()
        if total == 0:
            self.stdout.write(self.style.WARNING('No IPIntelligence rows match this filter.'))
            return
        pending = qs.filter(enriched_at__isnull=True).count()
        unmapped = qs.filter(enriched_at__isnull=False, lookup_ok=False).count()
        ipv4 = qs.filter(ip_version=4).count()
        ipv6 = qs.filter(ip_version=6).count()
        visit_sum = qs.aggregate(s=Sum('visit_count'))['s'] or 0

        self.stdout.write(self.style.MIGRATE_HEADING('=== Coverage ==='))
        self.stdout.write(f'  Total distinct IPs tracked : {total:,}')
        self.stdout.write(f'  Enriched                   : {total - pending:,} ({100 * (total - pending) / total:.1f}%)')
        self.stdout.write(f'  Pending enrichment         : {pending:,} ({100 * pending / total:.1f}%)')
        self.stdout.write(f'  Enriched but unmapped      : {unmapped:,} (private/reserved/not found)')
        self.stdout.write(f'  IPv4 / IPv6                : {ipv4:,} / {ipv6:,}')
        self.stdout.write(f'  Sum of visit_count         : {visit_sum:,} (sanity check vs. CrawlerVisit row count)')
        self.stdout.write('')

    def _top_countries(self, qs, top, export_dir):
        self.stdout.write(self.style.MIGRATE_HEADING('=== Top countries (by distinct IP / by visit_count) ==='))
        rows = list(
            qs.exclude(country_code='')
            .values('country_code', 'country_name')
            .annotate(ip_count=Count('id'), visits=Sum('visit_count'))
            .order_by('-ip_count')[:top]
        )
        self._print_table(rows, [
            ('country_code', 'Country'), ('country_name', 'Name'),
            ('ip_count', 'IPs'), ('visits', 'Visits'),
        ])
        if export_dir:
            self._write_csv(export_dir, 'top_countries.csv', rows)
        self.stdout.write('')

    def _top_asns(self, qs, top, export_dir):
        self.stdout.write(self.style.MIGRATE_HEADING('=== Top ASN orgs / ISPs (by distinct IP / by visit_count) ==='))
        rows = list(
            qs.exclude(asn_org='')
            .values('asn', 'asn_org')
            .annotate(ip_count=Count('id'), visits=Sum('visit_count'))
            .order_by('-ip_count')[:top]
        )
        self._print_table(rows, [
            ('asn', 'ASN'), ('asn_org', 'Organization'),
            ('ip_count', 'IPs'), ('visits', 'Visits'),
        ])
        if export_dir:
            self._write_csv(export_dir, 'top_asns.csv', rows)
        self.stdout.write('')

    def _hosting_split(self, qs):
        self.stdout.write(self.style.MIGRATE_HEADING('=== Hosting vs. residential (heuristic, ASN-org-keyword-based — not authoritative) ==='))
        total = qs.filter(enriched_at__isnull=False).count()
        if total == 0:
            self.stdout.write('  (nothing enriched yet)\n')
            return
        hosting = qs.filter(is_hosting=True).count()
        unmapped = qs.filter(lookup_ok=False).count()
        residential = total - hosting - unmapped
        self.stdout.write(f'  Hosting / datacenter : {hosting:,} ({100 * hosting / total:.1f}%)')
        self.stdout.write(f'  Likely residential   : {residential:,} ({100 * residential / total:.1f}%)')
        self.stdout.write(f'  Unmapped/unknown     : {unmapped:,} ({100 * unmapped / total:.1f}%)')
        self.stdout.write('')

    def _tor_crosstab(self, qs):
        self.stdout.write(self.style.MIGRATE_HEADING('=== Tor exit-node flag (best-effort, current exit list) ==='))
        total = qs.filter(enriched_at__isnull=False).count()
        if total == 0:
            self.stdout.write('  (nothing enriched yet)\n')
            return
        tor = qs.filter(is_tor_exit=True).count()
        self.stdout.write(f'  Tor exit IPs seen: {tor:,} ({100 * tor / total:.1f}% of enriched IPs)')
        self.stdout.write('')
        self.stdout.write('  Cross-tab:')
        for hosting_flag, label in [(True, 'Hosting'), (False, 'Non-hosting')]:
            for tor_flag, tor_label in [(True, 'Tor'), (False, 'Non-Tor')]:
                n = qs.filter(is_hosting=hosting_flag, is_tor_exit=tor_flag).count()
                self.stdout.write(f'    {label:<12} x {tor_label:<8}: {n:,}')
        self.stdout.write('')

    def _country_hosting_ratio(self, qs, top):
        self.stdout.write(self.style.MIGRATE_HEADING('=== Per-country hosting ratio (top countries by IP count) ==='))
        countries = list(
            qs.exclude(country_code='')
            .values('country_code')
            .annotate(ip_count=Count('id'))
            .order_by('-ip_count')[:top]
        )
        self.stdout.write(f"  {'Country':<10}{'IPs':>10}{'Hosting %':>12}")
        for c in countries:
            code = c['country_code']
            n = c['ip_count']
            hosting_n = qs.filter(country_code=code, is_hosting=True).count()
            pct = 100 * hosting_n / n if n else 0
            self.stdout.write(f"  {code:<10}{n:>10,}{pct:>11.1f}%")
        self.stdout.write('')

    def _top_outliers(self, qs, top):
        self.stdout.write(self.style.MIGRATE_HEADING(f'=== Top {top} IPs by visit_count ==='))
        rows = list(qs.order_by('-visit_count')[:top])
        for r in rows:
            self.stdout.write(
                f"  {r.ip_address:<40} visits={r.visit_count:>10,}  "
                f"{r.country_code or '??':<3} {r.asn_org or 'unknown org':<40} "
                f"hosting={r.is_hosting} tor={r.is_tor_exit}"
            )
            trap_breakdown = (
                CrawlerVisit.objects.filter(ip_address=r.ip_address)
                .values('trap_type').annotate(n=Count('id')).order_by('-n')[:5]
            )
            trap_str = ', '.join(f"{t['trap_type']}={t['n']}" for t in trap_breakdown)
            if trap_str:
                self.stdout.write(f"      traps: {trap_str}")
        self.stdout.write('')

    def _print_table(self, rows, columns):
        if not rows:
            self.stdout.write('  (none)')
            return
        header = ''.join(f'{label:<20}' for _, label in columns)
        self.stdout.write(f'  {header}')
        for row in rows:
            line = ''.join(f'{str(row.get(key, "") or ""):<20}' for key, _ in columns)
            self.stdout.write(f'  {line}')

    def _write_csv(self, export_dir, filename, rows):
        os.makedirs(export_dir, exist_ok=True)
        path = os.path.join(export_dir, filename)
        if not rows:
            return
        with open(path, 'w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=list(rows[0].keys()))
            writer.writeheader()
            writer.writerows(rows)
        self.stdout.write(f'  Wrote {path}')
