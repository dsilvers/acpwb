"""Tests for ws_server's filtering and aggregation (no Redis or sockets).

    pip install redis websockets orjson pytest && pytest ws_service/
"""
import pytest

import ws_server as ws


def req(**kw):
    r = {
        "timestamp": "2026-10-12T12:00:00.000000+00:00", "conn_id": "c1", "client_id": "abc",
        "ip_address": "203.0.113.7", "tls": True, "protocol": "HTTP/2.0", "host": "acpwb.com",
        "method": "GET", "path": "/wiki/topic/", "query": "", "user_agent": "Mozilla/5.0 Chrome/120",
        "status": 200, "route": "django", "error": "", "response_bytes": 1000,
        "duration_us": 20_000, "ttfb_us": 15_000, "tls_version": "TLSv1.3", "ja4": "t13d1516h2_aaa_bbb",
        "ja4h": "ge20nn", "h2_fingerprint": "1:65536|0|0|m,a,s,p", "new_client": False,
    }
    r.update(kw)
    return ws.prepare_request(r)


def conn(**kw):
    c = {"conn_id": "c1", "ip_address": "203.0.113.7", "outcome": "ok", "tls_version": "TLSv1.3",
         "ja4": "t13d1516h2_aaa_bbb", "sni": "acpwb.com", "close_reason": "client_closed"}
    c.update(kw)
    return ws.prepare_conn(c)


def matches(spec, r):
    return ws.Filter(spec).match(r)


def test_text_terms():
    r = req(path="/wiki/topic/", query="q=1")
    assert matches({"path": "wiki"}, r)
    assert matches({"path": "WIKI"}, r)
    assert matches({"path": "q=1"}, r)              # query is part of the URL
    assert not matches({"path": "!wiki"}, r)
    assert matches({"path": "/reports/, /wiki/"}, r)  # any positive term
    assert not matches({"path": "/wiki/, !topic"}, r)
    assert matches({"path": "/wiki/*/?q=1"}, r)      # glob over the whole value
    assert not matches({"path": "/wiki/*/"}, r)
    assert matches({"path": r"~^/wiki/\w+/"}, r)
    assert not matches({"path": r"~!^/wiki/"}, r)
    assert matches({"path": "  "}, r)


def test_bad_regex_names_the_field():
    with pytest.raises(ws.FilterError) as e:
        ws.Filter({"ua": "~(unclosed"})
    assert e.value.field == "ua"


def test_ip_prefix_and_cidr():
    r = req(ip_address="203.0.113.7")
    assert matches({"ip": "203.0.113."}, r)
    assert not matches({"ip": "03.0.113"}, r)       # prefix, not substring
    assert matches({"ip": "203.0.112.0/23"}, r)
    assert not matches({"ip": "10.0.0.0/8"}, r)
    assert not matches({"ip": "!203.0.113.0/24"}, r)
    assert matches({"ip": "2001:db8::/32"}, req(ip_address="2001:db8::1"))


def test_status():
    assert matches({"status": "2xx"}, req(status=204))
    assert matches({"status": "404, 5xx"}, req(status=503))
    assert not matches({"status": "404, 5xx"}, req(status=200))
    assert matches({"status": "400-499"}, req(status=451))
    assert not matches({"status": "!304"}, req(status=304))
    assert matches({"status": "!304"}, req(status=200))
    with pytest.raises(ws.FilterError):
        ws.Filter({"status": "abc"})


def test_sets_slow_and_flags():
    r = req(duration_us=1_500_000, error="", new_client=True)
    assert matches({"min_ms": "1000"}, r)
    assert not matches({"min_ms": 2000}, r)
    assert matches({"tls": ["TLSv1.3", "TLSv1.2"]}, r)
    assert not matches({"tls": ["TLSv1"]}, r)
    assert matches({"tls": ["plain"]}, req(tls=False, tls_version=""))
    assert matches({"proto": ["HTTP/2.0"]}, r)
    assert matches({"new_client": True}, r)
    assert not matches({"new_client": True}, req())
    assert not matches({"error": ["*"]}, r)
    assert matches({"error": ["*"]}, req(error="client_closed"))
    assert matches({"error": ["upstream_error"]}, req(error="upstream_error"))
    with pytest.raises(ws.FilterError):
        ws.Filter({"min_ms": "slow"})


def test_classification():
    assert req(user_agent="Mozilla/5.0 (compatible; GPTBot/1.0)")["_group"] == "AI Crawlers"
    b = req(user_agent="Mozilla/5.0 Firefox/130.0")
    assert b["_agent"] == "Browser · Firefox" and b["_group"] == "Other / Browser"
    assert req(user_agent="")["_group"] == "(empty user agent)"
    assert matches({"group": ["AI Crawlers"]}, req(user_agent="ClaudeBot"))
    assert matches({"agent": "firefox"}, b)


def test_edge_classification():
    # The edge's label wins over the UA-only guess.
    r = req(label="Residential · fake Chrome (Go TLS)", group="Residential bots", client_class="impersonator",
            network="residential", asn=28573, stack="Go TLS", signals=0b1000011)
    assert r["_agent"] == "Residential · fake Chrome (Go TLS)" and r["_group"] == "Residential bots"
    assert r["_signals"] == "tls_mismatch, no_h2, hosting"
    assert matches({"class": ["impersonator"], "network": ["residential"]}, r)
    assert matches({"group": ["Residential bots"], "asn": "285", "stack": "go tls"}, r)
    assert matches({"signal": "no_h2"}, r)
    assert not matches({"class": ["browser"]}, r)
    assert not matches({"asn": "1"}, r)  # a prefix of the ASN, not a substring
    # Records from an edge without classification fall back to the UA.
    assert req(user_agent="Mozilla/5.0 Firefox/130.0")["_signals"] == ""


def test_censoring():
    r = ws.prepare_request({"ip_address": "203.0.113.7", "user_agent": ""}, full_ips=False)
    assert r["ip_address"] == "203.0.113.xxx"
    r = ws.prepare_request({"ip_address": "2001:db8::1", "user_agent": ""}, full_ips=False)
    assert r["ip_address"] == "2001:db8::1"


def test_conn_filtering():
    f = ws.Filter({"ip": "203.0.113.", "ja4": "t13d"})
    assert f.match_conn(conn()) and not f.request_only
    assert not f.match_conn(conn(ip_address="198.51.100.1"))
    assert ws.Filter({"status": "404"}).request_only


class _WS:
    pass


def test_tick_and_window():
    v = ws.Viewer(_WS())
    v.set_filter({"path": "/wiki/"}, max_rows=2)
    reqs = [req(path=f"/wiki/{i}/", duration_us=i * 1000, status=200 if i % 2 else 404) for i in range(1, 6)]
    reqs.append(req(path="/static/x.css", conn_id="c2"))
    conns = [conn(), conn(conn_id="c2"), conn(conn_id="c3", outcome="handshake_failed", ja4="t12x")]
    bundle = {"t": 1, "requests_total": 10, "conns_total": 3}
    all_stats = ws.all_stats_for(bundle, reqs, conns)
    assert all_stats["sampled"] and all_stats["req"] == 10 and all_stats["conn_fail"] == 1

    t = ws.tick_for(v, bundle, reqs, conns, all_stats)
    m = t["match"]
    assert m["req"] == 5 and m["status"] == {"2xx": 3, "4xx": 2}
    assert m["p50"] == 3.0 and m["max"] == 5.0
    assert [r["path"] for r in t["rows"]] == ["/wiki/4/", "/wiki/5/"] and t["rows_omitted"] == 3
    # A path filter can't apply to connections, so only those that carried
    # a matching request are counted.
    assert m["conns"] == 1 and t["conn_note"]
    assert dict(t["window"]["top"]["status"]) == {"200": 3, "404": 2}

    for _ in range(ws.WINDOW_SECONDS + 5):
        t = ws.tick_for(v, bundle, reqs, conns, all_stats)
    w = t["window"]
    assert w["seconds"] == ws.WINDOW_SECONDS and w["req"] == 5 * ws.WINDOW_SECONDS
    assert dict(w["top"]["status"]) == {"200": 3 * ws.WINDOW_SECONDS, "404": 2 * ws.WINDOW_SECONDS}


def test_window_expires_keys():
    w = ws.Window(seconds=2)
    w.add({"ip": ws.collections.Counter({"a": 1})}, 1, 0)
    w.add({"ip": ws.collections.Counter({"b": 1})}, 1, 0)
    w.add({"ip": ws.collections.Counter({"b": 2})}, 2, 0)
    assert dict(w.totals["ip"]) == {"b": 3} and w.req == 3


def test_failed_handshakes_broken_out():
    v = ws.Viewer(_WS())
    conns = [conn(), conn(conn_id="x", outcome="handshake_failed", ja4="t12bad", ip_address="198.51.100.9")]
    t = ws.tick_for(v, {"t": 1}, [], conns, ws.all_stats_for({}, [], conns))
    top = t["window"]["top"]
    assert top["fail_ja4"] == [("t12bad", 1)] and top["fail_ip"] == [("198.51.100.9", 1)]
    assert t["match"]["conn_fail"] == 1
