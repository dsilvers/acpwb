"""
Rule-based residential-proxy/automation scoring — Phase 3 of the traffic-
intelligence work (see the plan this implements). Reads raw signal from
RequestFingerprint (Phase 2's per-request capture), cross-references
CrawlerVisit for absolute-proof trap hits and IPIntelligence for the
datacenter-vs-residential ASN split, and upserts IPReputationScore.

Windowed by RequestFingerprint's primary key (not time — this is a new,
much smaller table than CrawlerVisit, so simple ID-range batching is enough;
no TimescaleDB chunk-exclusion concerns here) and resumable via a watermark
in DashboardStat, same pattern as discover_ip_intelligence.

Deliberately does NOT touch PublishedIPReputation — see that model's
docstring for why the external-facing surface is synced separately
(publish_ip_reputation), on its own cadence, once there's confidence in the
scoring here.

Run via cron every few minutes once RequestFingerprint has real volume:
    python manage.py score_ip_reputation
"""
import fcntl
import time
from collections import defaultdict
from datetime import timedelta

from django.core.management.base import BaseCommand
from django.utils import timezone

from apps.core.models import DashboardStat
from apps.core.reputation_scoring import (
    ABSOLUTE_PROOF_TRAP_TYPES, compute_score, is_http1, is_old_tls_protocol, looks_like_modern_browser,
)
from apps.honeypot.models import CrawlerVisit, IPIntelligence, IPReputationScore, RequestFingerprint

_LOCK_FILE = '/tmp/acpwb-score-ip-reputation.lock'
_WATERMARK_KEY = 'ip_reputation_score_watermark'
_TAINT_LOOKBACK = timedelta(days=90)


class Command(BaseCommand):
    help = 'Score IPs for residential-proxy/automation likelihood from RequestFingerprint signal.'

    def add_arguments(self, parser):
        parser.add_argument('--batch', type=int, default=5000, help='RequestFingerprint rows per window (default: 5000)')
        parser.add_argument('--max-batches', type=int, default=0, help='0 = unlimited (default: 0)')
        parser.add_argument('--max-seconds', type=int, default=50)

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another scoring run is already in progress — exiting.')
                return
            self._run(options)

    def _get_watermark(self):
        stat, _ = DashboardStat.objects.get_or_create(key=_WATERMARK_KEY, defaults={'value': {'id': 0}})
        return stat.value.get('id', 0)

    def _set_watermark(self, last_id):
        DashboardStat.objects.update_or_create(key=_WATERMARK_KEY, defaults={'value': {'id': last_id}})

    def _aggregate_batch(self, rows):
        """Per-IP aggregation of raw per-request signal into the boolean
        flags compute_score() expects. "Any row in this batch shows the
        mismatch" rather than a majority/percentage threshold — simpler, and
        a v1 rule-based scorer erring toward sensitivity is an acceptable
        tradeoff (see the plan's phased rollout: this is validated against
        real traffic in the dashboard before anything downstream trusts it)."""
        by_ip = defaultdict(lambda: {
            'header_mismatch': False, 'protocol_mismatch': False,
            'tls_mismatch': False, 'new_ip_no_referrer': False,
        })
        for row in rows:
            flags = by_ip[row.ip_address]
            modern_browser = looks_like_modern_browser(row.user_agent)
            if modern_browser and not row.browser_headers_present:
                flags['header_mismatch'] = True
            if modern_browser and is_http1(row.client_protocol):
                flags['protocol_mismatch'] = True
            if modern_browser and is_old_tls_protocol(row.tls_protocol):
                flags['tls_mismatch'] = True
            if row.first_seen_ip and not row.referrer_present:
                flags['new_ip_no_referrer'] = True
        return by_ip

    def _tainted_ips(self, ip_addresses):
        cutoff = timezone.now() - _TAINT_LOOKBACK
        tainted = CrawlerVisit.objects.filter(
            ip_address__in=ip_addresses,
            trap_type__in=ABSOLUTE_PROOF_TRAP_TYPES,
            timestamp__gte=cutoff,
        ).values_list('ip_address', flat=True).distinct()
        return set(tainted)

    def _hosting_by_ip(self, ip_addresses):
        rows = IPIntelligence.objects.filter(ip_address__in=ip_addresses).values_list('ip_address', 'is_hosting', 'lookup_ok')
        return {ip: (is_hosting if lookup_ok else None) for ip, is_hosting, lookup_ok in rows}

    def _score_batch(self, rows):
        by_ip = self._aggregate_batch(rows)
        ip_addresses = list(by_ip.keys())
        if not ip_addresses:
            return 0

        tainted = self._tainted_ips(ip_addresses)
        hosting = self._hosting_by_ip(ip_addresses)
        now = timezone.now()
        scored = 0

        for ip, flags in by_ip.items():
            score, classification, evidence = compute_score(
                trap_tainted=ip in tainted,
                is_hosting=hosting.get(ip),
                **flags,
            )
            existing = IPReputationScore.objects.filter(ip_address=ip).first()
            first_flagged_at = existing.first_flagged_at if existing else None
            if first_flagged_at is None and classification != 'human':
                first_flagged_at = now

            IPReputationScore.objects.update_or_create(
                ip_address=ip,
                defaults={
                    'residential_proxy_score': score,
                    'classification': classification,
                    'evidence': evidence,
                    'first_flagged_at': first_flagged_at,
                    'last_scored_at': now,
                },
            )
            scored += 1
        return scored

    def _run(self, options):
        batch_size = options['batch']
        max_batches = options['max_batches']
        max_seconds = options['max_seconds']
        started = time.monotonic()

        watermark = self._get_watermark()
        total_scored = 0
        batches_run = 0

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                self.stdout.write(f'Time budget reached after {batches_run} batch(es).')
                break

            rows = list(
                RequestFingerprint.objects.filter(id__gt=watermark).order_by('id')[:batch_size]
            )
            if not rows:
                break

            total_scored += self._score_batch(rows)
            watermark = rows[-1].id
            self._set_watermark(watermark)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        self.stdout.write(self.style.SUCCESS(
            f'Scored {total_scored} IP(s) across {batches_run} batch(es); watermark now {watermark}.'
        ))
