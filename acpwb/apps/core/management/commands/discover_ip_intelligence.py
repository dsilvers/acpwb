"""
Populate IPIntelligence with every distinct IP seen in CrawlerVisit, without
ever running a full-table DISTINCT (that took ~5 minutes on the 90M+
rows/day honeypot_crawlervisit hypertable — see the hourly_traffic_report
postmortem this command replaces for IP discovery purposes).

Serves both the one-time historical backfill and ongoing incremental
discovery — the only difference is how large the gap between the stored
watermark and now happens to be, which shrinks to a normal small tail after
the first few runs.

Mechanics: step through time in fixed windows from a high-watermark (stored
in DashboardStat) up to now() minus a safety margin (to avoid racing rows
still being drained from Redis by drain_crawler_queue). Each window does
discovery + aggregation + upsert in a single raw SQL statement — the
GROUP BY only ever touches that window's rows, so cost stays bounded
regardless of total table size.

New IPIntelligence rows land with enriched_at = NULL; enrich_ip_intelligence
picks those up separately (this command never touches MaxMind).

Usage:
    python manage.py discover_ip_intelligence                  # incremental
    python manage.py discover_ip_intelligence --step-hours 24  # deep catch-up
    python manage.py discover_ip_intelligence --dry-run
    python manage.py discover_ip_intelligence --since 2024-01-01
    python manage.py discover_ip_intelligence --full-history    # scan back to the first-ever row

On a table this size (500M+ rows and growing, 7-day TimescaleDB chunks
compressed after 7 days), walking the full history by default would mean
decompressing every old chunk back to day one just to seed the watermark —
exactly the query shape that caused a production Postgres connection
exhaustion incident (see deploy/README.md, "Incident: Postgres connection
exhaustion"). So the FIRST run (no stored watermark yet) defaults to
`now - --max-lookback-days` (14 days) instead of the true earliest
CrawlerVisit row. Pass --full-history to opt into the true earliest row, or
--since to pick an explicit start date. Once a watermark is stored, every
later run resumes from it regardless of these flags.

Run every 5 minutes via cron for ongoing discovery once the initial backfill
catches up. On production, run small windows (--step-hours 1, the default)
and watch the printed per-window timing before trusting it to cron — do NOT
default to large windows or --full-history there without confirming timing
first.
"""
import fcntl
import time
from datetime import timedelta

from django.core.management.base import BaseCommand
from django.db import connection
from django.utils import timezone
from django.utils.dateparse import parse_date

from apps.core.models import DashboardStat
from apps.honeypot.models import CrawlerVisit

_LOCK_FILE = '/tmp/acpwb-ip-intel-discover.lock'
_WATERMARK_KEY = 'ip_intel_discover_watermark'

_UPSERT_SQL = """
    INSERT INTO honeypot_ipintelligence (
        ip_address, ip_version, first_seen, last_seen, visit_count,
        country_code, country_name, region_name, city_name, asn_org,
        enrichment_note, is_hosting, is_tor_exit, lookup_ok
    )
    SELECT
        ip_address,
        CASE WHEN family(ip_address) = 6 THEN 6 ELSE 4 END,
        MIN(timestamp), MAX(timestamp), COUNT(*),
        '', '', '', '', '',
        '', false, false, false
    FROM honeypot_crawlervisit
    WHERE timestamp >= %s AND timestamp < %s
    GROUP BY ip_address
    ON CONFLICT (ip_address) DO UPDATE SET
        first_seen  = LEAST(honeypot_ipintelligence.first_seen, EXCLUDED.first_seen),
        last_seen   = GREATEST(honeypot_ipintelligence.last_seen, EXCLUDED.last_seen),
        visit_count = honeypot_ipintelligence.visit_count + EXCLUDED.visit_count
"""

_COUNT_SQL = """
    SELECT COUNT(DISTINCT ip_address) FROM honeypot_crawlervisit
    WHERE timestamp >= %s AND timestamp < %s
"""


class Command(BaseCommand):
    help = 'Discover distinct IPs from CrawlerVisit into IPIntelligence, windowed and resumable.'

    def add_arguments(self, parser):
        parser.add_argument('--step-hours', type=int, default=1, help='Window size in hours (default: 1)')
        parser.add_argument('--max-seconds', type=int, default=50, help='Stop starting new windows after this many seconds (default: 50, 0=unlimited)')
        parser.add_argument('--safety-margin-seconds', type=int, default=120, help='Do not process the last N seconds, in case rows are still being drained (default: 120)')
        parser.add_argument('--since', type=str, default=None, metavar='YYYY-MM-DD', help='Reset the watermark to this date instead of using the stored one')
        parser.add_argument('--dry-run', action='store_true', help='Print the window list and per-window distinct-IP counts without writing')
        parser.add_argument('--max-lookback-days', type=int, default=14, help='First-run only: how far back to seed the watermark instead of the true earliest row (default: 14)')
        parser.add_argument('--full-history', action='store_true', help='First-run only: seed the watermark from the true earliest CrawlerVisit row instead of --max-lookback-days')

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another discovery run is already in progress — exiting.')
                return
            self._run(options)

    def _run(self, options):
        step = timedelta(hours=options['step_hours'])
        max_seconds = options['max_seconds']
        safety_margin = timedelta(seconds=options['safety_margin_seconds'])
        upper_bound = timezone.now() - safety_margin

        watermark = self._get_watermark(options['since'], options['max_lookback_days'], options['full_history'])
        if watermark is None:
            self.stdout.write('No CrawlerVisit rows exist — nothing to discover.')
            return

        if watermark >= upper_bound:
            self.stdout.write(f'Already caught up (watermark {watermark.isoformat()}).')
            return

        self._tune_session()

        started = time.monotonic()
        cursor_ts = watermark
        windows_run = 0

        while cursor_ts < upper_bound:
            if max_seconds and time.monotonic() - started >= max_seconds:
                self.stdout.write(f'Time budget reached after {windows_run} window(s); resuming next run from {cursor_ts.isoformat()}.')
                break

            window_end = min(cursor_ts + step, upper_bound)

            if options['dry_run']:
                with connection.cursor() as cur:
                    cur.execute(_COUNT_SQL, [cursor_ts, window_end])
                    count = cur.fetchone()[0]
                self.stdout.write(f'  {cursor_ts.isoformat()} -> {window_end.isoformat()}  — {count:,} distinct IPs')
            else:
                t0 = time.monotonic()
                with connection.cursor() as cur:
                    cur.execute(_UPSERT_SQL, [cursor_ts, window_end])
                elapsed = time.monotonic() - t0
                self._set_watermark(window_end)
                self.stdout.write(f'  {cursor_ts.isoformat()} -> {window_end.isoformat()}  — done ({elapsed:.1f}s)')

            cursor_ts = window_end
            windows_run += 1

        if not options['dry_run']:
            self.stdout.write(self.style.SUCCESS(f'\nProcessed {windows_run} window(s); watermark now {cursor_ts.isoformat()}.'))

    def _tune_session(self):
        with connection.cursor() as cur:
            cur.execute("SET work_mem = '128MB'")
            cur.execute("SET synchronous_commit = off")

    def _get_watermark(self, since_str, max_lookback_days, full_history):
        if since_str:
            since_date = parse_date(since_str)
            if since_date is None:
                raise ValueError(f'Invalid --since date: {since_str}')
            return timezone.make_aware(timezone.datetime.combine(since_date, timezone.datetime.min.time()))

        stat, _ = DashboardStat.objects.get_or_create(key=_WATERMARK_KEY, defaults={'value': {}})
        ts_str = stat.value.get('ts')
        if ts_str:
            return timezone.datetime.fromisoformat(ts_str)

        # No stored watermark yet — this is the first-ever run. Don't default
        # to the true earliest row: on a 500M+ row hypertable with 7-day
        # compressed chunks, that walks every chunk back to day one (the
        # query shape behind the connection-exhaustion incident in
        # deploy/README.md). Cap to a recent lookback unless explicitly
        # opted out via --full-history.
        earliest = CrawlerVisit.objects.order_by('timestamp').values_list('timestamp', flat=True).first()
        if earliest is None:
            return None
        if full_history:
            return earliest
        capped = timezone.now() - timedelta(days=max_lookback_days)
        return max(earliest, capped)

    def _set_watermark(self, ts):
        DashboardStat.objects.update_or_create(
            key=_WATERMARK_KEY, defaults={'value': {'ts': ts.isoformat()}}
        )
