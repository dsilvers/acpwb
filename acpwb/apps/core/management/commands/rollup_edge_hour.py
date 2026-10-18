"""
Roll each finished hour of edge traffic up into EdgeIPHour (one row per
client IP), EdgeFingerprintHour (one row per JA4 + HTTP/2 fingerprint) and
EdgeClientHour (one row per client cookie that came back).

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

# 'browser' rows from hosting networks predate the edge's headless class
# (Oct 14 2026): count them as headless, as the edge does now.
_HEADLESS = "(client_class = 'headless' OR (client_class = 'browser' AND network = 'hosting'))"
_BROWSER = "(client_class = 'browser' AND network <> 'hosting')"
_PAGE = "(route IN ('django', 'acpwb_go') AND method = 'GET')"
# An IP whose browser page loads came with none of their assets and no
# returning cookie: a fresh profile fetching documents with subresources
# blocked (a browser farm), not a person, who'd get the page's CSS and JS.
_BARE = "(r.browser_pages > 0 AND r.static = 0 AND r.cookie = 0)"
# A cookie seen from this many IPs in one hour is a pool's shared jar, not
# one person moving between Wi-Fi and mobile data.
POOL_IPS = 3
# Proof-of-automation trap hits, for joining to EdgeRequest by request_id.
# The app logs a hit at about the time the edge does; give it slack.
_TRAPS = """
    SELECT DISTINCT request_id FROM honeypot_crawlervisit
    WHERE timestamp >= %(start)s - interval '15 minutes' AND timestamp < %(spill)s
      AND trap_type = ANY(%(traps)s) AND request_id IS NOT NULL
"""

# Only clients that sent their cookie back at least once this hour: the
# rest are one-request clients that keep no cookies (most bots), and there
# are a million of those an hour.
_CLIENT_SQL = """
WITH t AS (""" + _TRAPS + """)
INSERT INTO honeypot_edgeclienthour (
    hour, client_id, first_seen, requests, page_reqs, static_reqs, trap_reqs,
    ips, asns, residential_ips, hosting_ips, ja4s, user_agents,
    browser_reqs, headless_reqs, impersonator_reqs, known_bot_reqs,
    top_ip, top_ja4, top_class, top_user_agent, signals
)
SELECT
    %(start)s, e.client_id, bool_or(e.new_client), count(*),
    count(*) FILTER (WHERE """ + _PAGE + """), count(*) FILTER (WHERE route = 'static'), count(t.request_id),
    count(DISTINCT ip_address), count(DISTINCT asn),
    count(DISTINCT ip_address) FILTER (WHERE network = 'residential'),
    count(DISTINCT ip_address) FILTER (WHERE network = 'hosting'),
    count(DISTINCT nullif(ja4, '')), count(DISTINCT user_agent),
    count(*) FILTER (WHERE """ + _BROWSER + """), count(*) FILTER (WHERE """ + _HEADLESS + """),
    count(*) FILTER (WHERE client_class = 'impersonator'), count(*) FILTER (WHERE client_class = 'known_bot'),
    mode() WITHIN GROUP (ORDER BY ip_address), coalesce(mode() WITHIN GROUP (ORDER BY ja4), ''),
    coalesce(mode() WITHIN GROUP (ORDER BY client_class), ''),
    left(coalesce(mode() WITHIN GROUP (ORDER BY user_agent), ''), 512),
    bit_or(signals)
FROM honeypot_edgerequest e
LEFT JOIN t USING (request_id)
WHERE e.timestamp >= %(start)s AND e.timestamp < %(end)s AND e.client_id <> ''
GROUP BY e.client_id
HAVING bool_or(NOT e.new_client)
"""

_IP_SQL = """
WITH r AS (
    SELECT
        ip_address, max(asn) AS asn, max(nullif(network, '')) AS network,
        count(*) AS requests, count(DISTINCT conn_id) AS conns,
        count(*) FILTER (WHERE """ + _BROWSER + """) AS browser,
        count(*) FILTER (WHERE """ + _HEADLESS + """) AS headless,
        count(*) FILTER (WHERE client_class = 'impersonator') AS impersonator,
        count(*) FILTER (WHERE client_class = 'known_bot') AS known_bot,
        count(*) FILTER (WHERE client_class = 'tool') AS tool,
        count(*) FILTER (WHERE """ + _PAGE + """) AS pages,
        count(*) FILTER (WHERE """ + _BROWSER + """ AND """ + _PAGE + """) AS browser_pages,
        count(p.pool_id) AS pool_cookie,
        count(*) FILTER (WHERE client_class IN ('unknown', '')) AS unknown,
        count(*) FILTER (WHERE route = 'static') AS static,
        count(*) FILTER (WHERE error <> '') AS errors,
        count(*) FILTER (WHERE new_client) AS new_client,
        count(*) FILTER (WHERE client_id <> '' AND NOT new_client) AS cookie,
        count(DISTINCT nullif(ja4, '')) AS ja4s, count(DISTINCT user_agent) AS uas,
        mode() WITHIN GROUP (ORDER BY ja4) AS top_ja4,
        mode() WITHIN GROUP (ORDER BY stack) AS top_stack,
        bit_or(signals) AS signals
    FROM honeypot_edgerequest e
    LEFT JOIN (
        SELECT client_id AS pool_id FROM honeypot_edgeclienthour WHERE hour = %(start)s AND ips >= %(pool_ips)s
    ) p ON p.pool_id = e.client_id
    WHERE timestamp >= %(start)s AND timestamp < %(end)s
    GROUP BY ip_address
),
c AS (
    SELECT
        ip_address, max(asn) AS asn, max(nullif(network, '')) AS network,
        sum(conn_count) FILTER (WHERE outcome = 'no_hello') AS no_hello,
        sum(conn_count) FILTER (WHERE outcome IN ('handshake_failed', 'timeout')) AS failed
    FROM honeypot_edgeconnection
    WHERE timestamp >= %(start)s AND timestamp < %(end)s AND outcome IN ('no_hello', 'handshake_failed', 'timeout')
    GROUP BY ip_address
),
t AS (
    SELECT ip_address, count(*) AS hits
    FROM honeypot_crawlervisit
    WHERE timestamp >= %(start)s AND timestamp < %(end)s AND trap_type = ANY(%(traps)s)
    GROUP BY ip_address
)
INSERT INTO honeypot_edgeiphour (
    hour, ip_address, asn, network, requests, conns,
    browser_reqs, headless_reqs, impersonator_reqs, known_bot_reqs, tool_reqs, unknown_reqs,
    page_reqs, static_reqs, error_reqs, new_client_reqs, cookie_reqs, pool_cookie_reqs, trap_hits,
    bare_page_reqs, behaviour_proven_reqs, no_hello_conns, failed_handshakes,
    ja4s, user_agents, top_ja4, top_stack, signals
)
SELECT
    %(start)s, coalesce(r.ip_address, c.ip_address), coalesce(r.asn, c.asn), coalesce(r.network, c.network, ''),
    coalesce(r.requests, 0), coalesce(r.conns, 0),
    coalesce(r.browser, 0), coalesce(r.headless, 0), coalesce(r.impersonator, 0), coalesce(r.known_bot, 0),
    coalesce(r.tool, 0), coalesce(r.unknown, 0),
    coalesce(r.pages, 0), coalesce(r.static, 0), coalesce(r.errors, 0), coalesce(r.new_client, 0),
    coalesce(r.cookie, 0), coalesce(r.pool_cookie, 0), coalesce(t.hits, 0),
    -- r is NULL for an IP with only failed / hello-less connections (and
    -- maybe a trap hit) this hour.
    CASE WHEN """ + _BARE + """ THEN r.browser_pages ELSE 0 END,
    CASE WHEN """ + _BARE + """ OR t.hits > 0 OR r.pool_cookie > 0
         THEN coalesce(r.browser + r.unknown, 0) ELSE 0 END,
    coalesce(c.no_hello, 0), coalesce(c.failed, 0),
    coalesce(r.ja4s, 0), coalesce(r.uas, 0), coalesce(r.top_ja4, ''), coalesce(r.top_stack, ''),
    coalesce(r.signals, 0)
FROM r
FULL JOIN c ON c.ip_address = r.ip_address
LEFT JOIN t ON t.ip_address = coalesce(r.ip_address, c.ip_address)
"""

_FP_SQL = """
WITH t AS (""" + _TRAPS + """),
c AS (
    SELECT conn_id, ja4, h2_fingerprint, ip_address, network, conn_count, request_count, first_user_agent
    FROM honeypot_edgeconnection
    WHERE timestamp >= %(start)s AND timestamp < %(end)s AND ja4 <> ''
),
r AS (
    SELECT conn_id,
        count(*) FILTER (WHERE """ + _BROWSER + """) AS browser,
        count(*) FILTER (WHERE """ + _HEADLESS + """) AS headless,
        count(*) FILTER (WHERE client_class = 'impersonator') AS impersonator,
        count(*) FILTER (WHERE client_class = 'known_bot') AS known_bot,
        count(t.request_id) AS traps,
        max(stack) AS stack
    FROM honeypot_edgerequest
    LEFT JOIN t USING (request_id)
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
    browser_reqs, headless_reqs, impersonator_reqs, known_bot_reqs, trap_reqs, trap_ips,
    user_agents, top_stack, top_user_agent
)
SELECT
    %(start)s, c.ja4, c.h2_fingerprint, sum(c.conn_count), sum(c.request_count),
    count(DISTINCT c.ip_address),
    count(DISTINCT c.ip_address) FILTER (WHERE c.network = 'residential'),
    count(DISTINCT c.ip_address) FILTER (WHERE c.network = 'hosting'),
    max(s.ips),
    coalesce(sum(r.browser), 0), coalesce(sum(r.headless), 0), coalesce(sum(r.impersonator), 0),
    coalesce(sum(r.known_bot), 0), coalesce(sum(r.traps), 0),
    count(DISTINCT c.ip_address) FILTER (WHERE r.traps > 0),
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
    (ip rows, fingerprint rows, client rows)."""
    params = {
        'start': start, 'end': start + timedelta(hours=1), 'spill': start + timedelta(hours=1) + _REQUEST_SPILL,
        'traps': list(ABSOLUTE_PROOF_TRAP_TYPES), 'pool_ips': POOL_IPS,
    }
    with transaction.atomic(), connection.cursor() as cur:
        for table in ('edgeiphour', 'edgefingerprinthour', 'edgeclienthour'):
            cur.execute(f'DELETE FROM honeypot_{table} WHERE hour = %s', [start])
        cur.execute(_CLIENT_SQL, params)  # first: the IP rollup reads its pool cookies
        clients = cur.rowcount
        cur.execute(_IP_SQL, params)
        ips = cur.rowcount
        cur.execute(_FP_SQL, params)
        return ips, cur.rowcount, clients


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
        ips, fps, clients = rollup_hour(start)
        self.stdout.write(f'{start:%Y-%m-%dT%H}: {ips} IPs, {fps} fingerprints, {clients} clients '
                          f'({time.monotonic() - t0:.1f}s)')

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
