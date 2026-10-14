"""
Roll each finished hour of edge traffic up into EdgeIPHour (one row per
client IP) and EdgeFingerprintHour (one row per JA4 + HTTP/2 fingerprint).

These are what the statistics and IP reputation scoring read: the raw
EdgeRequest/EdgeConnection tables run to tens of millions of rows a day,
the rollups to a couple of million. Like the raw tables they're kept
forever.

An hour is rolled up once it is ROLLUP_DELAY past its end (drains and
connections closing need the slack), resuming from a watermark in
DashboardStat. Each hour is replaced whole (delete + insert in one
transaction), so re-running an hour (--hour) is safe.

Run via cron every 15 minutes (see deploy/acpwb-crontab).
"""
import fcntl
import time
from datetime import datetime, timedelta, timezone as dt_timezone

from django.core.management.base import BaseCommand, CommandError
from django.db import connection, transaction
from django.utils import timezone

from apps.core.models import DashboardStat
from apps.core.reputation_scoring import ABSOLUTE_PROOF_TRAP_TYPES

_LOCK_FILE = '/tmp/acpwb-rollup-edge-hour.lock'
_WATERMARK_KEY = 'edge_rollup_watermark'  # {'hour': ISO of the last hour rolled up}
ROLLUP_DELAY = timedelta(minutes=15)
# Requests of a connection accepted near the end of an hour land in the
# next one; join them for this long past the hour.
_REQUEST_SPILL = timedelta(minutes=15)

_IP_SQL = """
INSERT INTO honeypot_edgeiphour (
    hour, ip_address, asn, network, requests, conns,
    browser_reqs, impersonator_reqs, known_bot_reqs, tool_reqs, unknown_reqs,
    static_reqs, error_reqs, new_client_reqs, cookie_reqs, trap_hits,
    ja4s, user_agents, top_ja4, top_stack, signals
)
SELECT
    %(start)s, r.ip_address, max(r.asn), coalesce(max(nullif(r.network, '')), ''),
    count(*), count(DISTINCT r.conn_id),
    count(*) FILTER (WHERE r.client_class = 'browser'),
    count(*) FILTER (WHERE r.client_class = 'impersonator'),
    count(*) FILTER (WHERE r.client_class = 'known_bot'),
    count(*) FILTER (WHERE r.client_class = 'tool'),
    count(*) FILTER (WHERE r.client_class IN ('unknown', '')),
    count(*) FILTER (WHERE r.route = 'static'),
    count(*) FILTER (WHERE r.error <> ''),
    count(*) FILTER (WHERE r.new_client),
    count(*) FILTER (WHERE r.client_id <> '' AND NOT r.new_client),
    coalesce(max(t.hits), 0),
    count(DISTINCT nullif(r.ja4, '')), count(DISTINCT r.user_agent),
    coalesce(mode() WITHIN GROUP (ORDER BY r.ja4), ''),
    coalesce(mode() WITHIN GROUP (ORDER BY r.stack), ''),
    coalesce(bit_or(r.signals), 0)
FROM honeypot_edgerequest r
LEFT JOIN (
    SELECT ip_address, count(*) AS hits
    FROM honeypot_crawlervisit
    WHERE timestamp >= %(start)s AND timestamp < %(end)s AND trap_type = ANY(%(traps)s)
    GROUP BY ip_address
) t ON t.ip_address = r.ip_address
WHERE r.timestamp >= %(start)s AND r.timestamp < %(end)s
GROUP BY r.ip_address
"""

_FP_SQL = """
WITH c AS (
    SELECT conn_id, ja4, h2_fingerprint, ip_address, network, conn_count, request_count, first_user_agent
    FROM honeypot_edgeconnection
    WHERE timestamp >= %(start)s AND timestamp < %(end)s AND ja4 <> ''
),
r AS (
    SELECT conn_id,
        count(*) FILTER (WHERE client_class = 'browser') AS browser,
        count(*) FILTER (WHERE client_class = 'impersonator') AS impersonator,
        count(*) FILTER (WHERE client_class = 'known_bot') AS known_bot,
        max(stack) AS stack
    FROM honeypot_edgerequest
    WHERE timestamp >= %(start)s AND timestamp < %(spill)s AND conn_id IS NOT NULL
    GROUP BY conn_id
),
per_ip AS (
    SELECT ja4, h2_fingerprint, ip_address, sum(conn_count) AS n
    FROM c GROUP BY ja4, h2_fingerprint, ip_address
),
single AS (
    SELECT ja4, h2_fingerprint, count(*) FILTER (WHERE n = 1) AS ips
    FROM per_ip GROUP BY ja4, h2_fingerprint
)
INSERT INTO honeypot_edgefingerprinthour (
    hour, ja4, h2_fingerprint, conns, requests, ips, residential_ips, hosting_ips, single_conn_ips,
    browser_reqs, impersonator_reqs, known_bot_reqs, user_agents, top_stack, top_user_agent
)
SELECT
    %(start)s, c.ja4, c.h2_fingerprint, sum(c.conn_count), sum(c.request_count),
    count(DISTINCT c.ip_address),
    count(DISTINCT c.ip_address) FILTER (WHERE c.network = 'residential'),
    count(DISTINCT c.ip_address) FILTER (WHERE c.network = 'hosting'),
    max(s.ips),
    coalesce(sum(r.browser), 0), coalesce(sum(r.impersonator), 0), coalesce(sum(r.known_bot), 0),
    count(DISTINCT nullif(c.first_user_agent, '')),
    coalesce(mode() WITHIN GROUP (ORDER BY r.stack), ''),
    left(coalesce(mode() WITHIN GROUP (ORDER BY nullif(c.first_user_agent, '')), ''), 512)
FROM c
LEFT JOIN r USING (conn_id)
JOIN single s ON s.ja4 = c.ja4 AND s.h2_fingerprint = c.h2_fingerprint
GROUP BY c.ja4, c.h2_fingerprint
"""


def floor_hour(t):
    return t.replace(minute=0, second=0, microsecond=0)


def rollup_hour(start):
    """Replace the rollups for the hour starting at start. Returns
    (ip rows, fingerprint rows)."""
    params = {
        'start': start, 'end': start + timedelta(hours=1), 'spill': start + timedelta(hours=1) + _REQUEST_SPILL,
        'traps': list(ABSOLUTE_PROOF_TRAP_TYPES),
    }
    with transaction.atomic(), connection.cursor() as cur:
        cur.execute('DELETE FROM honeypot_edgeiphour WHERE hour = %s', [start])
        cur.execute('DELETE FROM honeypot_edgefingerprinthour WHERE hour = %s', [start])
        cur.execute(_IP_SQL, params)
        ips = cur.rowcount
        cur.execute(_FP_SQL, params)
        return ips, cur.rowcount


class Command(BaseCommand):
    help = 'Roll finished hours of EdgeRequest/EdgeConnection up into EdgeIPHour and EdgeFingerprintHour.'

    def add_arguments(self, parser):
        parser.add_argument('--hour', help='Roll up just this hour (YYYY-MM-DDTHH, UTC); the watermark is left alone')
        parser.add_argument('--since', help='First run only: start from this hour (YYYY-MM-DDTHH, UTC) '
                                            'instead of the last finished one')
        parser.add_argument('--max-seconds', type=int, default=600, help="Don't start another hour after this long")

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another rollup is already in progress — exiting.')
                return
            if options['hour']:
                start = self._parse_hour(options['hour'])
                self._one(start)
                return
            self._run(options)

    def _parse_hour(self, s):
        try:
            return datetime.strptime(s, '%Y-%m-%dT%H').replace(tzinfo=dt_timezone.utc)
        except ValueError:
            raise CommandError(f'bad hour {s!r}: use YYYY-MM-DDTHH (UTC)')

    def _one(self, start):
        t0 = time.monotonic()
        ips, fps = rollup_hour(start)
        self.stdout.write(f'{start:%Y-%m-%dT%H}: {ips} IPs, {fps} fingerprints ({time.monotonic() - t0:.1f}s)')

    def _run(self, options):
        last_finished = floor_hour(timezone.now() - ROLLUP_DELAY) - timedelta(hours=1)
        stat = DashboardStat.objects.filter(key=_WATERMARK_KEY).first()
        if stat and stat.value.get('hour'):
            nxt = datetime.fromisoformat(stat.value['hour']) + timedelta(hours=1)
        elif options['since']:
            nxt = self._parse_hour(options['since'])
        else:
            nxt = last_finished
        deadline = time.monotonic() + options['max_seconds']
        done = 0
        while nxt <= last_finished and time.monotonic() < deadline:
            self._one(nxt)
            DashboardStat.objects.update_or_create(key=_WATERMARK_KEY, defaults={'value': {'hour': nxt.isoformat()}})
            nxt += timedelta(hours=1)
            done += 1
        behind = int((last_finished - nxt) / timedelta(hours=1)) + 1 if nxt <= last_finished else 0
        self.stdout.write(f'Rolled up {done} hour(s); {behind} still to do.')
