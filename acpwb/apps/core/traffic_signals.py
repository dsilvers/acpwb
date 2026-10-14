"""
Names for EdgeRequest.signals, the bitmask acpwb_edge's request classifier
sets (acpwb_edge/classify/classify.go, the Sig* constants — same order).
The decisive ones (DECISIVE) make a browser claim an impersonator; the rest
are evidence for reputation scoring.
"""

SIGNALS = [
    ('tls_mismatch', "TLS stack isn't the claimed browser's"),
    ('no_h2', "browser that didn't offer HTTP/2"),
    ('h2_mismatch', "HTTP/2 framing isn't the claimed browser's"),
    ('h2_stale', 'Chrome 106+ UA with older Chrome HTTP/2 settings'),
    ('no_accept_language', 'browser claim without Accept-Language'),
    ('cookie_dropped', 'client cookie not returned on the same connection'),
    ('hosting', 'hosting/cloud network'),
    ('old_tls', 'TLS 1.1 or older from a modern browser UA'),
    ('bot_residential', 'self-declared crawler from a residential network'),
    ('known_range', 'known bot IP range'),
    ('header_order', "headers in an order the claimed browser never sends"),
    ('exploit_probe', 'exploit header (x-middleware-subrequest: Next.js auth bypass)'),
    ('client_hints', "Chromium's sec-ch-ua hints missing or contradicting the UA"),
    ('hidden_link', 'page only linked invisibly (ghost links): no person asks for it'),
]

BIT = {name: 1 << i for i, (name, _) in enumerate(SIGNALS)}
DECISIVE = ('tls_mismatch', 'no_h2', 'h2_mismatch', 'header_order', 'client_hints')


def signal_names(mask):
    """The names of the signals set in mask, in bit order."""
    return [name for i, (name, _) in enumerate(SIGNALS) if mask & (1 << i)]
