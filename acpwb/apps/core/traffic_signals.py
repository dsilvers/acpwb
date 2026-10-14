"""
Names for EdgeRequest.signals, the bitmask acpwb_edge's request classifier
sets (acpwb_edge/classify/classify.go, the Sig* constants — same order).
The first three are decisive: any one of them on a browser claim makes the
request an impersonator. The rest are evidence for reputation scoring.
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
]

BIT = {name: 1 << i for i, (name, _) in enumerate(SIGNALS)}


def signal_names(mask):
    """The names of the signals set in mask, in bit order."""
    return [name for i, (name, _) in enumerate(SIGNALS) if mask & (1 << i)]
