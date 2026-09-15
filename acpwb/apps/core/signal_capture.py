"""Pure signal-extraction helpers shared by BotTrackingMiddleware/
RequestStreamMiddleware (Django) and mirrored in acpwb_go's fingerprint
capture (Go) — kept side-effect-free so both callers and tests can use them
without touching Redis/DB.
"""

# Headers a real browser sends that a scripted HTTP client generally doesn't
# bother replicating, even when it fakes a browser User-Agent — residential
# proxies only touch the IP layer, so this leaks through regardless of proxy
# use. Fixed, small candidate list (not "every header seen") so the derived
# signal has low, meaningful cardinality instead of encoding arbitrary junk
# headers a bot happens to send.
BROWSER_SIGNAL_HEADERS = [
    'accept-language',
    'sec-fetch-site',
    'sec-fetch-mode',
    'sec-fetch-dest',
    'sec-fetch-user',
    'sec-ch-ua',
    'sec-ch-ua-mobile',
    'sec-ch-ua-platform',
    'upgrade-insecure-requests',
]


def browser_headers_present(header_names):
    """header_names: iterable of lowercase header names present on the
    request. Returns a sorted, comma-joined subset of BROWSER_SIGNAL_HEADERS
    that were actually present — deterministic regardless of wire order, so
    it means the same thing whether captured by Django or acpwb_go."""
    present = set(header_names) & set(BROWSER_SIGNAL_HEADERS)
    return ','.join(sorted(present))
