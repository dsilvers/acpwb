"""Pure rule-based scoring for residential-proxy/automation detection —
kept side-effect-free (no DB/Redis) so apps.core.management.commands.
score_ip_reputation and its tests can exercise the actual scoring logic
directly, without needing RequestFingerprint/CrawlerVisit/IPIntelligence
rows in a real database.

Score/classification meaning is versioned via IPReputationScore.score_version
— revise the weights/thresholds here freely, but bump that constant when a
change would make an old row's score mean something different than it did
when it was written.
"""

SCORE_VERSION = 'v1'

# CrawlerVisit.trap_type values no real human browsing could ever trigger —
# hitting any of these proves automation regardless of what the UA/headers on
# that IP's *other* requests look like (see score_ip_reputation's taint
# lookup). Deliberately excludes trap types like 'archive'/'wiki' that get
# real (if rare) human traffic via search-engine indexing.
ABSOLUTE_PROOF_TRAP_TYPES = ('ghost_link', 'canary_trigger', 'env_probe', 'wp_probe', 'webshell_probe')

# UA substrings a scripted client generally doesn't bother replicating, and
# that a modern desktop/mobile browser reliably does — used only to decide
# whether "this UA claims a browser" is worth checking header/protocol/TLS
# consistency against, not as a bot classifier on its own (apps.core.
# bot_classify already owns that).
_MODERN_BROWSER_MARKERS = ('chrome/', 'firefox/', 'edg/', 'crios/', 'fxios/')
_NON_BROWSER_MARKERS = (
    'bot', 'crawler', 'spider', 'curl/', 'wget/', 'python-requests',
    'scrapy', 'go-http-client', 'okhttp', 'aiohttp', 'libwww',
)

# TLS versions no current browser release negotiates by default — nginx
# forwards $ssl_protocol as-is (see nginx/acpwb.com's X-TLS-Protocol header).
_OLD_TLS_PROTOCOLS = {'SSLv2', 'SSLv3', 'TLSv1', 'TLSv1.1'}

# Rule weights. Deliberately simple/additive for a v1 — see module docstring
# re: bumping SCORE_VERSION before changing these.
_WEIGHTS = {
    'trap_tainted': 40,
    'header_mismatch': 25,
    'protocol_mismatch': 20,
    'tls_mismatch': 15,
    'new_ip_no_referrer': 10,
}
_HUMAN_THRESHOLD = 20


def looks_like_modern_browser(user_agent):
    """True if `user_agent` claims to be a modern desktop/mobile browser —
    the population these mismatch checks apply to. Not a bot classifier
    (apps.core.bot_classify already owns that); a scripted client that
    doesn't bother claiming a browser UA at all isn't what this function, or
    residential-proxy detection generally, is trying to catch — it's UA
    spoofing specifically that this signal is built to see through."""
    if not user_agent:
        return False
    ua = user_agent.lower()
    if any(marker in ua for marker in _NON_BROWSER_MARKERS):
        return False
    if any(marker in ua for marker in _MODERN_BROWSER_MARKERS):
        return True
    return 'safari/' in ua and 'version/' in ua


def compute_score(*, trap_tainted, header_mismatch, protocol_mismatch, tls_mismatch,
                   new_ip_no_referrer, is_hosting):
    """Combine boolean signals into (score, classification, evidence).

    is_hosting: True/False from IPIntelligence, or None if no IPIntelligence
    row exists yet for this IP (discover_ip_intelligence hasn't caught up).
    A None is treated the same as "score is high but we can't tell datacenter
    from residential yet" -> 'unknown', rather than guessing.
    """
    evidence = []
    score = 0
    for name, fired in (
        ('trap_tainted', trap_tainted),
        ('header_mismatch', header_mismatch),
        ('protocol_mismatch', protocol_mismatch),
        ('tls_mismatch', tls_mismatch),
        ('new_ip_no_referrer', new_ip_no_referrer),
    ):
        if fired:
            score += _WEIGHTS[name]
            evidence.append(name)
    score = min(score, 100)

    if score < _HUMAN_THRESHOLD:
        classification = 'human'
    elif is_hosting is True:
        classification = 'datacenter_bot'
    elif is_hosting is False:
        classification = 'residential_proxy_bot'
    else:
        classification = 'unknown'

    return score, classification, evidence


def is_old_tls_protocol(tls_protocol):
    return bool(tls_protocol) and tls_protocol in _OLD_TLS_PROTOCOLS


def is_http1(client_protocol):
    """True for both nginx's exact "HTTP/1.1" and HAProxy's coarser
    "HTTP/1" (only the major version is available from HAProxy's
    fc_http_major fetch — see haproxy/haproxy.cfg) — this signal needs to
    mean the same thing regardless of which layer captured it."""
    return bool(client_protocol) and client_protocol.startswith('HTTP/1')
