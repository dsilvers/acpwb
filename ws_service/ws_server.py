"""
Standalone asyncio WebSocket server for the dashboard's live request stream
(/acpwb-dashboard/live/).

acpwb_edge PUBLISHes one bundle a second on Redis channel "edge_live":
every request it saw that second (static files, redirects and failed
requests included, each with its connection's TLS details) and every
connection that closed — but only while WATCH_KEY exists, which this
service keeps alive while it has a viewer. Bundles are capped on the edge;
requests_total / conns_total stay exact, and "sampled" in a tick says the
cap was hit.

Each viewer sends a filter; for each bundle this service sends that viewer
one "tick": totals for all traffic, stats for the matching requests, rolling
WINDOW_SECONDS top-N breakdowns, connection outcomes, and up to max_rows
matching requests. Filtering happens here so a busy second costs the
browser one message and a bounded number of rows, not one message per
request.

Protocol (JSON text frames):
  server → {"type": "hello", "full_ips": bool, "window_s": 60, "max_rows_limit": 500, "fields": [...]}
  client → {"type": "filter", "filter": {...}, "max_rows": 100}
  server → {"type": "filter_ok"} | {"type": "filter_error", "field": "...", "message": "..."}
  server → {"type": "tick", ...} once a second ({"type": "tick", "stale": true} while no bundles arrive)

Environment variables:
  REDIS_URL         Redis URL (default: redis://redis:6379/0)
  WS_HOST           Bind host (default: 0.0.0.0)
  WS_PORT           Bind port (default: 8765)
  STREAM_WS_TOKEN   Required token in ?token= query param (optional; if unset,
                    open access — and IPv4 addresses are sent with the last
                    octet censored, as the old stream did)
"""

import asyncio
import collections
import fnmatch
import importlib.util
import ipaddress
import logging
import os
import re
import time
import urllib.parse
from functools import lru_cache
from pathlib import Path

import redis.asyncio as aioredis
from websockets.asyncio.server import serve

try:
    import orjson

    def dumps(v):
        return orjson.dumps(v).decode()

    loads = orjson.loads
except ImportError:  # pragma: no cover
    import json

    def dumps(v):
        return json.dumps(v, separators=(',', ':'))

    loads = json.loads

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

REDIS_URL = os.environ.get("REDIS_URL", "redis://redis:6379/0")
WS_HOST = os.environ.get("WS_HOST", "0.0.0.0")
WS_PORT = int(os.environ.get("WS_PORT", "8765"))
STREAM_WS_TOKEN = os.environ.get("STREAM_WS_TOKEN", "")

CHANNEL = "edge_live"                # acpwb_edge livefeed.Channel
WATCH_KEY = "acpwb:live:watching"    # acpwb_edge livefeed.WatchKey
WATCH_TTL = 5
WINDOW_SECONDS = 60
DEFAULT_ROWS = 100
MAX_ROWS = 500
SEND_TIMEOUT = 5.0

# ---------------------------------------------------------------------------
# Bot classification — the Django classifier, loaded from the repo (or a copy
# mounted next to this file in Docker) so there's one list of patterns.
# ---------------------------------------------------------------------------


def _load_bot_classify():
    here = Path(__file__).resolve().parent
    for path in (here / "bot_classify.py", here.parent / "acpwb" / "apps" / "core" / "bot_classify.py"):
        if path.exists():
            spec = importlib.util.spec_from_file_location("bot_classify", path)
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            return mod
    logger.warning("bot_classify.py not found — agents will all be 'Other / Browser'")
    return None


_bc = _load_bot_classify()
BROWSER = "Other / Browser"
EMPTY_UA = "(empty user agent)"


@lru_cache(maxsize=8192)
def _classify_ua(ua):
    if _bc is None:
        return EMPTY_UA if not ua.strip() else BROWSER
    return _bc.classify_ua(ua)


@lru_cache(maxsize=8192)
def _classify_ip(ip):
    return _bc.classify_ip(ip) if _bc is not None else None


@lru_cache(maxsize=256)
def _group(agent):
    return _bc.bot_type_to_group(agent) if _bc is not None else agent


_BROWSER_FAMILIES = (("edg/", "Edge"), ("opr/", "Opera"), ("firefox/", "Firefox"),
                     ("chrome/", "Chrome"), ("safari/", "Safari"))


@lru_cache(maxsize=8192)
def _browser_family(ua):
    low = ua.lower()
    for needle, name in _BROWSER_FAMILIES:
        if needle in low:
            return name
    return "Other"


def classify(ua, ip):
    """(agent, group): agent is the bot name, or "Browser · Chrome" etc. for
    anything not matched as a bot — the Django classifier lumps every
    browser UA (and every spoofed one) into 'Other / Browser'."""
    agent = _classify_ua(ua or "")
    if agent == BROWSER:
        agent = _classify_ip(ip) or BROWSER
    group = _group(agent)
    if agent == BROWSER:
        agent = "Browser · " + _browser_family(ua)
    return agent, group


def censor_ip(ip):
    parts = ip.split(".")
    if len(parts) == 4:
        parts[-1] = "xxx"
        return ".".join(parts)
    return ip


# ---------------------------------------------------------------------------
# Record preparation — once per bundle, shared by every viewer.
# ---------------------------------------------------------------------------


def prepare_request(r, full_ips=True):
    """Add the derived fields filters and stats use (all prefixed _)."""
    ip = r.get("ip_address") or ""
    r["_agent"], r["_group"] = classify(r.get("user_agent") or "", ip)
    if not full_ips:
        r["ip_address"] = censor_ip(ip)
    q = r.get("query") or ""
    r["_url"] = (r.get("path") or "") + ("?" + q if q else "")
    r["_ms"] = round((r.get("duration_us") or 0) / 1000, 1)
    r["_tls"] = (r.get("tls_version") or "TLS?") if r.get("tls") else "plain"
    return r


def prepare_conn(c, full_ips=True):
    if not full_ips:
        c["ip_address"] = censor_ip(c.get("ip_address") or "")
    c["_tls"] = c.get("tls_version") or ("plain" if c.get("outcome") == "plain" else "none")
    return c


# ---------------------------------------------------------------------------
# Filters
#
# Text fields: comma-separated terms. A term matches case-insensitively as a
# substring (IP, client and conn IDs: as a prefix; IP also as a CIDR), with
# * as a wildcard over the whole value; !term excludes. A value must match
# one of the positive terms (if any) and none of the excluded ones. A field
# starting with ~ is a single regular expression instead (~! negates it).
# ---------------------------------------------------------------------------

# filter key → (record key, mode)
TEXT_FIELDS = {
    "path": ("_url", "substr"),
    "host": ("host", "substr"),
    "ip": ("ip_address", "ip"),
    "ua": ("user_agent", "substr"),
    "agent": ("_agent", "substr"),
    "referer": ("referer", "substr"),
    "ja4": ("ja4", "substr"),
    "ja4h": ("ja4h", "substr"),
    "ja3": ("ja3", "substr"),
    "h2": ("h2_fingerprint", "substr"),
    "cipher": ("tls_cipher", "substr"),
    "lang": ("accept_language", "substr"),
    "ctype": ("content_type", "substr"),
    "client": ("client_id", "prefix"),
    "conn": ("conn_id", "prefix"),
}
# filter key → record key; the filter value is a list of allowed values
SET_FIELDS = {
    "method": "method",
    "proto": "protocol",
    "tls": "_tls",
    "route": "route",
    "group": "_group",
    "error": "error",
}
# Filters that also apply to connections, and the conn record key.
CONN_FIELDS = {"ip": "ip_address", "ja4": "ja4", "ja3": "ja3", "host": "sni", "cipher": "tls_cipher",
               "tls": "_tls", "conn": "conn_id"}


class FilterError(ValueError):
    def __init__(self, field, message):
        super().__init__(message)
        self.field = field


def _term_test(term, mode):
    low = term.lower()
    if "*" in term:
        rx = re.compile(fnmatch.translate(term), re.I | re.S)
        return lambda v: rx.match(v) is not None
    if mode == "ip":
        if "/" in term:
            try:
                net = ipaddress.ip_network(term, strict=False)
            except ValueError:
                pass
            else:
                def in_net(v):
                    try:
                        return ipaddress.ip_address(v) in net
                    except ValueError:
                        return False
                return in_net
        return lambda v: v.lower().startswith(low)
    if mode == "prefix":
        return lambda v: v.lower().startswith(low)
    return lambda v: low in v.lower()


def text_matcher(expr, mode="substr", field=""):
    """A predicate over strings for a text filter expression, or None if the
    expression is empty."""
    expr = (expr or "").strip()
    if not expr:
        return None
    if expr.startswith("~"):
        body, neg = expr[1:], False
        if body.startswith("!"):
            body, neg = body[1:], True
        try:
            rx = re.compile(body, re.I)
        except re.error as e:
            raise FilterError(field, f"bad regex: {e}") from None
        if neg:
            return lambda v: rx.search(v or "") is None
        return lambda v: rx.search(v or "") is not None
    pos, neg = [], []
    for term in expr.split(","):
        term = term.strip()
        if term.startswith("!"):
            term = term[1:].strip()
            if term:
                neg.append(_term_test(term, mode))
        elif term:
            pos.append(_term_test(term, mode))
    if not pos and not neg:
        return None

    def match(v):
        v = v or ""
        if pos and not any(t(v) for t in pos):
            return False
        return not any(t(v) for t in neg)
    return match


_STATUS_TERM = re.compile(r"^(\d)xx$|^(\d{1,3})$|^(\d{1,3})-(\d{1,3})$", re.I)


def status_matcher(expr):
    """"404", "4xx", "500-599", "!304", comma-separated; status 0 means no
    response was sent."""
    expr = (expr or "").strip()
    if not expr:
        return None
    pos, neg = [], []
    for term in expr.split(","):
        term = term.strip()
        if not term:
            continue
        target = pos
        if term.startswith("!"):
            term, target = term[1:].strip(), neg
        m = _STATUS_TERM.match(term)
        if not m:
            raise FilterError("status", f"bad status term {term!r} (use 404, 4xx, 500-599, !304)")
        if m.group(1):
            lo = int(m.group(1)) * 100
            target.append((lo, lo + 99))
        elif m.group(2):
            target.append((int(m.group(2)), int(m.group(2))))
        else:
            target.append((int(m.group(3)), int(m.group(4))))

    def match(s):
        s = s or 0
        if pos and not any(lo <= s <= hi for lo, hi in pos):
            return False
        return not any(lo <= s <= hi for lo, hi in neg)
    return match


class Filter:
    """A compiled viewer filter. Raises FilterError on bad input."""

    def __init__(self, spec=None):
        spec = spec or {}
        self.req_tests = []   # (record key, predicate)
        self.conn_tests = []
        self.request_only = False  # some filter can't apply to connections
        for key, (field, mode) in TEXT_FIELDS.items():
            m = text_matcher(spec.get(key), mode, key)
            if m:
                self.req_tests.append((field, m))
                if key in CONN_FIELDS:
                    self.conn_tests.append((CONN_FIELDS[key], m))
                else:
                    self.request_only = True
        for key, field in SET_FIELDS.items():
            vals = spec.get(key)
            if not vals:
                continue
            if isinstance(vals, str):
                vals = [vals]
            allowed = frozenset(str(v) for v in vals)
            if key == "error" and "*" in allowed:   # any error at all
                m = bool
            else:
                m = allowed.__contains__
            self.req_tests.append((field, m))
            if key in CONN_FIELDS:
                self.conn_tests.append((CONN_FIELDS[key], m))
            else:
                self.request_only = True
        st = status_matcher(spec.get("status"))
        if st:
            self.req_tests.append(("status", st))
            self.request_only = True
        try:
            min_ms = float(spec.get("min_ms") or 0)
        except (TypeError, ValueError):
            raise FilterError("min_ms", "not a number") from None
        if min_ms > 0:
            us = min_ms * 1000
            self.req_tests.append(("duration_us", lambda v: (v or 0) >= us))
            self.request_only = True
        if spec.get("new_client"):
            self.req_tests.append(("new_client", bool))
            self.request_only = True
        self.active = bool(self.req_tests)

    def match(self, r):
        for field, test in self.req_tests:
            if not test(r.get(field)):
                return False
        return True

    def match_conn(self, c):
        for field, test in self.conn_tests:
            if not test(c.get(field)):
                return False
        return True


# ---------------------------------------------------------------------------
# Stats
# ---------------------------------------------------------------------------

# Rolling breakdown dimensions: name → (key function, top N sent)
REQ_DIMS = {
    "path": (lambda r: r.get("path") or "", 12),
    "ip": (lambda r: r.get("ip_address") or "", 12),
    "agent": (lambda r: r["_agent"], 12),
    "group": (lambda r: r["_group"], 10),
    "host": (lambda r: r.get("host") or "", 10),
    "ja4": (lambda r: r.get("ja4") or "", 10),
    "ja4h": (lambda r: r.get("ja4h") or "", 10),
    "h2": (lambda r: r.get("h2_fingerprint") or "", 8),
    "status": (lambda r: str(r.get("status") or 0), 15),
    "tls": (lambda r: r["_tls"], 10),
    "proto": (lambda r: r.get("protocol") or "", 6),
    "route": (lambda r: r.get("route") or "", 15),
    "method": (lambda r: r.get("method") or "", 10),
    "error": (lambda r: r.get("error") or "", 15),
}
CONN_DIMS = {
    "outcome": (lambda c: c.get("outcome") or "", 10),
    "close": (lambda c: c.get("close_reason") or "", 12),
    "conn_tls": (lambda c: c["_tls"], 10),
    "fail_ja4": (lambda c: c.get("ja4") or "(none)" if c.get("outcome") != "ok" else None, 10),
    "fail_ip": (lambda c: c.get("ip_address") or "" if c.get("outcome") != "ok" else None, 10),
}


def percentile(sorted_vals, p):
    if not sorted_vals:
        return None
    i = min(len(sorted_vals) - 1, int(round(p / 100 * (len(sorted_vals) - 1))))
    return sorted_vals[i]


def second_counters(reqs, conns):
    out = {}
    for name, (key, _) in REQ_DIMS.items():
        out[name] = collections.Counter(key(r) for r in reqs)
    for name, (key, _) in CONN_DIMS.items():
        c = collections.Counter(key(x) for x in conns)
        c.pop(None, None)
        out[name] = c
    return out


class Window:
    """Rolling WINDOW_SECONDS of per-second counters, with running totals
    kept incrementally (add the new second, subtract the expired one)."""

    def __init__(self, seconds=WINDOW_SECONDS):
        self.seconds = seconds
        self.secs = collections.deque()
        self.totals = collections.defaultdict(collections.Counter)
        self.req = 0
        self.conns = 0

    def add(self, counters, nreq, nconn):
        self.secs.append((counters, nreq, nconn))
        self.req += nreq
        self.conns += nconn
        for name, c in counters.items():
            self.totals[name].update(c)
        while len(self.secs) > self.seconds:
            old, oreq, oconn = self.secs.popleft()
            self.req -= oreq
            self.conns -= oconn
            for name, c in old.items():
                tot = self.totals[name]
                for k, n in c.items():
                    left = tot[k] - n
                    if left > 0:
                        tot[k] = left
                    else:
                        del tot[k]

    def top(self):
        out = {}
        for name, (_, n) in {**REQ_DIMS, **CONN_DIMS}.items():
            out[name] = self.totals[name].most_common(n)
        out["unique_ips"] = len(self.totals["ip"])
        out["unique_paths"] = len(self.totals["path"])
        return out


def tick_for(viewer, bundle, reqs, conns, all_stats):
    """The tick message for one viewer from one prepared bundle."""
    f = viewer.filter
    matched = [r for r in reqs if f.match(r)] if f.active else reqs
    mconns = [c for c in conns if f.match_conn(c)] if f.conn_tests else conns
    if f.request_only:
        # A filter like status or path says nothing about connections that
        # carried no matching request; show only the ones that did.
        conn_ids = {r.get("conn_id") for r in matched}
        mconns = [c for c in mconns if c.get("conn_id") in conn_ids]

    viewer.window.add(second_counters(matched, mconns), len(matched), len(mconns))

    lat = sorted(r["_ms"] for r in matched)
    ttfb = sorted(r.get("ttfb_us", 0) / 1000 for r in matched if r.get("ttfb_us"))
    classes = collections.Counter()
    for r in matched:
        s = r.get("status") or 0
        classes[f"{s // 100}xx" if s else "none"] += 1

    rows = matched[-viewer.max_rows:] if viewer.max_rows else []
    return {
        "type": "tick",
        "t": bundle.get("t"),
        "sampled": all_stats["sampled"],
        "all": all_stats,
        "match": {
            "req": len(matched),
            "bytes": sum(r.get("response_bytes") or 0 for r in matched),
            "errors": sum(1 for r in matched if r.get("error")),
            "status": dict(classes),
            "new_clients": sum(1 for r in matched if r.get("new_client")),
            "ips": len({r.get("ip_address") for r in matched}),
            "p50": percentile(lat, 50), "p95": percentile(lat, 95), "p99": percentile(lat, 99),
            "max": lat[-1] if lat else None,
            "ttfb_p50": round(percentile(ttfb, 50), 1) if ttfb else None,
            "conns": len(mconns),
            "conn_fail": sum(1 for c in mconns if c.get("outcome") not in ("ok", "plain")),
        },
        "window": {"seconds": len(viewer.window.secs), "req": viewer.window.req,
                   "conns": viewer.window.conns, "top": viewer.window.top()},
        "rows": rows,
        "rows_omitted": len(matched) - len(rows),
        "conn_note": "request filters applied" if f.request_only else "",
    }


def all_stats_for(bundle, reqs, conns):
    total = bundle.get("requests_total", len(reqs))
    ctotal = bundle.get("conns_total", len(conns))
    return {
        "req": total,
        "req_sent": len(reqs),
        "conns": ctotal,
        "conns_sent": len(conns),
        "bytes": sum(r.get("response_bytes") or 0 for r in reqs),
        "conn_fail": sum(1 for c in conns if c.get("outcome") not in ("ok", "plain")),
        "sampled": total > len(reqs) or ctotal > len(conns),
    }


# ---------------------------------------------------------------------------
# Viewers
# ---------------------------------------------------------------------------


class Viewer:
    def __init__(self, ws):
        self.ws = ws
        self.filter = Filter()
        self.max_rows = DEFAULT_ROWS
        self.window = Window()
        self.sending = None  # the in-flight send task, if any

    def set_filter(self, spec, max_rows=None):
        self.filter = Filter(spec)  # raises FilterError before anything changes
        if max_rows is not None:
            self.max_rows = max(0, min(MAX_ROWS, int(max_rows)))
        self.window = Window()

    def send(self, text):
        # A viewer still receiving the previous tick (slow link, background
        # tab) skips this one rather than queueing up behind it.
        if self.sending is not None and not self.sending.done():
            return
        self.sending = asyncio.ensure_future(self._send(text))

    async def _send(self, text):
        try:
            await asyncio.wait_for(self.ws.send(text), SEND_TIMEOUT)
        except Exception:
            viewers.pop(self.ws, None)
            try:
                await self.ws.close()
            except Exception:
                pass


viewers: dict = {}


def _parse_token(path: str) -> str:
    """Extract ?token= from the WebSocket request path."""
    qs = urllib.parse.urlparse(path).query
    return urllib.parse.parse_qs(qs).get("token", [""])[0]


async def handle_client(websocket):
    """Accept a WebSocket connection, auth-check it, then read its filters."""
    if STREAM_WS_TOKEN:
        token = _parse_token(websocket.request.path)
        if token != STREAM_WS_TOKEN:
            await websocket.close(4401, "Unauthorized")
            return

    viewer = Viewer(websocket)
    viewers[websocket] = viewer
    logger.info("Client connected: %s (total: %d)", websocket.remote_address, len(viewers))
    try:
        await websocket.send(dumps({
            "type": "hello", "full_ips": bool(STREAM_WS_TOKEN), "window_s": WINDOW_SECONDS,
            "max_rows_limit": MAX_ROWS,
        }))
        async for raw in websocket:
            try:
                msg = loads(raw)
            except ValueError:
                continue
            if not isinstance(msg, dict) or msg.get("type") != "filter":
                continue
            try:
                viewer.set_filter(msg.get("filter") or {}, msg.get("max_rows"))
            except FilterError as e:
                await websocket.send(dumps({"type": "filter_error", "field": e.field, "message": str(e)}))
            except (TypeError, ValueError) as e:
                await websocket.send(dumps({"type": "filter_error", "field": "", "message": str(e)}))
            else:
                await websocket.send(dumps({"type": "filter_ok"}))
    except Exception:
        pass
    finally:
        viewers.pop(websocket, None)
        logger.info("Client disconnected (total: %d)", len(viewers))


def handle_bundle(bundle):
    if not viewers:
        return
    full_ips = bool(STREAM_WS_TOKEN)
    reqs = [prepare_request(r, full_ips) for r in bundle.get("requests") or ()]
    conns = [prepare_conn(c, full_ips) for c in bundle.get("conns") or ()]
    all_stats = all_stats_for(bundle, reqs, conns)
    for viewer in list(viewers.values()):
        try:
            viewer.send(dumps(tick_for(viewer, bundle, reqs, conns, all_stats)))
        except Exception:
            logger.exception("building tick")


async def redis_listener():
    """Keep WATCH_KEY alive while anyone is watching, and fan each edge
    bundle out to the viewers."""
    delay = 5
    while True:
        try:
            client = aioredis.from_url(REDIS_URL, decode_responses=False)
            pubsub = client.pubsub()
            await pubsub.subscribe(CHANNEL)
            logger.info("Connected to Redis, listening on %r", CHANNEL)
            delay = 5  # reset backoff on successful connect
            last_watch = last_bundle = last_stale = 0.0
            while True:
                now = time.monotonic()
                if viewers and now - last_watch >= 1:
                    await client.set(WATCH_KEY, b"1", ex=WATCH_TTL)
                    last_watch = now
                msg = await pubsub.get_message(ignore_subscribe_messages=True, timeout=1.0)
                if msg and msg["type"] == "message":
                    last_bundle = time.monotonic()
                    try:
                        bundle = loads(msg["data"])
                    except ValueError:
                        continue
                    handle_bundle(bundle)
                elif viewers and time.monotonic() - last_bundle > 3 and time.monotonic() - last_stale >= 1:
                    # The edge publishes within a second of seeing WATCH_KEY;
                    # tell viewers if it isn't (edge down, nginx rollback).
                    last_stale = time.monotonic()
                    stale = dumps({"type": "tick", "stale": True})
                    for viewer in list(viewers.values()):
                        viewer.send(stale)
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            logger.warning("Redis connection failed (%s), retrying in %ds", exc, delay)
            await asyncio.sleep(delay)
            delay = min(delay * 2, 60)


async def main():
    listener_task = asyncio.create_task(redis_listener())
    async with serve(handle_client, WS_HOST, WS_PORT, max_size=64 * 1024) as server:
        logger.info("WebSocket server listening on %s:%d", WS_HOST, WS_PORT)
        await server.serve_forever()
    listener_task.cancel()


if __name__ == "__main__":
    asyncio.run(main())
