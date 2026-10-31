"""
Bot user-agent classification utilities.
Shared between BotTrackingMiddleware and dashboard_views to avoid circular imports.

Two classification paths:
  classify_ua(ua)            — match by user-agent string against BOT_PATTERNS
  classify_ip(ip)            — match by source IP against known bot CIDR ranges
  classify_ua_or_ip(ua, ip)  — UA first; falls back to IP when UA returns 'Other / Browser',
                               and relabels a vendor's name from outside its published ranges
"""
import bisect
import ipaddress
from pathlib import Path

# ---------------------------------------------------------------------------
# Known bot IP CIDR ranges
# Format: (cidr_string, display_name)
# Checked in order; first match wins.
# ---------------------------------------------------------------------------
_IP_BOT_RANGE_DEFS = [
    # Alibaba / Qwen AI crawler
    ('47.79.0.0/16',   'Alibaba Qwen'),  # observed traffic
    ('47.82.60.0/22',  'Alibaba Qwen'),  # 47.82.60–63
    ('8.219.0.0/16',   'Alibaba Qwen'),  # observed traffic
    ('43.156.0.0/16', 'Tencent'),
    ('43.172.0.0/16', 'Tencent'),
    ('43.173.0.0/16', 'Tencent'),
    # INTERNEXUS, LLC scraper pool (uniform Chrome/Mac UA, throttled per-IP)
    ('207.180.11.0/24', 'INTERNEXUS Scraper Pool'),
    ('216.75.132.0/24', 'INTERNEXUS Scraper Pool'),
    # IPXO-leased blocks, single HK customer (ORG-PC1271-RIPE), same fingerprint
    ('143.20.253.0/24', 'IPXO Scraper Pool (HK)'),
    ('143.14.6.0/24',   'IPXO Scraper Pool (HK)'),
    ('144.31.35.0/24',  'IPXO Scraper Pool (HK)'),
]

# Pre-parse networks at import time so per-request lookup is fast
IP_BOT_RANGES = [
    (ipaddress.ip_network(cidr, strict=False), name)
    for cidr, name in _IP_BOT_RANGE_DEFS
]


def classify_ip(ip_str):
    """
    Return a bot name if ip_str falls within a known bot IP range, else None.
    Returns None (not 'Other / Browser') so callers can distinguish "no match".
    """
    try:
        addr = ipaddress.ip_address(ip_str.strip())
    except ValueError:
        return None
    for network, name in IP_BOT_RANGES:
        if addr in network:
            return name
    return None


BOT_PATTERNS = [
    # AI crawlers — most interesting
    ('GPTBot',              'OpenAI GPTBot'),
    ('OAI-SearchBot',       'OpenAI SearchBot'),
    ('ChatGPT-User',        'OpenAI ChatGPT'),
    ('OAI-AdsBot',          'OpenAI AdsBot'),
    ('Claude-User',         'Anthropic Claude-User'),  # before ClaudeBot: UA ends +claudebot@anthropic.com
    ('Claude-SearchBot',    'Anthropic Claude-SearchBot'),  # same
    ('ClaudeBot',           'Anthropic ClaudeBot'),
    ('Claude-Web',          'Anthropic Claude'),
    ('anthropic-ai',        'Anthropic'),
    ('PerplexityBot',       'Perplexity'),
    ('Perplexity-User',     'Perplexity-User'),
    ('DuckAssistBot',       'DuckAssistBot'),
    ('Google-Extended',     'Google-Extended (AI)'),
    ('Google-Agent',        'Google-Agent'),
    ('meta-externalagent',  'Meta ExternalAgent'),
    ('meta-webindexer',     'Meta WebIndexer'),
    ('FacebookBot',         'Meta FacebookBot'),
    ('Applebot-Extended',   'Apple Applebot-Extended'),
    ('Bytespider',          'ByteDance Bytespider'),
    ('Amazonbot',           'Amazonbot'),
    ('Diffbot',             'Diffbot'),
    ('omgili',              'Omgilibot'),
    ('webzio-extended',     'Webzio'),
    ('CCBot',               'Common Crawl'),
    ('cohere-ai',           'Cohere'),
    ('Timpibot',            'Timpi'),
    ('ShapBot',             'Parallel ShapBot'),
    ('Reflectionbot',       'Reflection AI'),
    ('YouBot',              'You.com YouBot'),
    ('Keenable',            'Keenable'),  # KeenableBot, Keenable-SearchBot, Keenable-User
    ('LinkupBot',           'Linkup'),
    # Search engines
    ('Nexus 5X Build/MMB29P', 'Googlebot Mobile'),  # Googlebot mobile fingerprint (before Googlebot)
    ('Googlebot',           'Googlebot'),
    ('GoogleOther',         'GoogleOther'),
    ('bingbot',             'Bingbot'),
    ('BingPreview',         'Bing Preview'),
    ('msnbot',              'MSN Bot'),
    ('Baiduspider',         'Baiduspider'),
    ('YandexBot',           'YandexBot'),
    ('Slurp',               'Yahoo Slurp'),
    ('DuckDuckBot',         'DuckDuckBot'),
    ('Applebot',            'Applebot'),
    ('sogou',               'Sogou'),
    ('360Spider',           '360Spider'),
    ('SeznamBot',           'Seznam'),
    ('Bravebot',            'Bravebot'),
    # SEO/marketing crawlers
    ('SemrushBot',          'SemrushBot'),
    ('AhrefsBot',           'AhrefsBot'),
    ('MJ12bot',             'Majestic MJ12'),
    ('DotBot',              'DotBot'),
    ('DataForSeoBot',       'DataForSEO'),
    ('PetalBot',            'Huawei PetalBot'),
    ('PiplBot',             'Pipl'),
    ('SERankingBacklinksBot', 'SERankingBacklinksBot'),
    ('ZoominfoBot',         'ZoominfoBot'),
    ('AwarioBot',           'AwarioBot'),
    ('BitSightBot',         'BitSightBot'),
    ('zgrab',               'zgrab'),
    ('AionBot',             'AionBot'),  # operator unknown; Oracle Cloud, ignores robots.txt
    ('Cortex-Xpanse',       'Palo Alto Xpanse'),
    ('AdsBot-Google',       'AdsBot-Google'),
    ('TikTokSpider',        'TikTokSpider'),
    ('HubSpot Crawler',     'HubSpot Crawler'),
    ('BuiltWith',           'BuiltWith'),
    ('TechnicalSEO.com',    'TechnicalSEO'),
    ('PimEyes',             'PimEyes'),
    # Social
    ('Twitterbot',          'Twitterbot'),
    ('facebookexternalhit', 'Facebook Scraper'),
    ('LinkedInBot',         'LinkedIn'),
    # Archives
    ('ia_archiver',         'Internet Archive'),
    ('archive.org_bot',     'Internet Archive'),
    # Generic HTTP clients (likely scrapers/bots)
    ('python-requests',     'Python Requests'),
    ('curl/',               'cURL'),
    ('wget',                'Wget'),
    ('scrapy',              'Scrapy'),
    ('Go-http-client',      'Go HTTP Client'),
    ('Java/',               'Java HTTP Client'),
    ('libwww-perl',         'libwww-perl'),
    ('axios',               'axios'),
    ('node-fetch',          'node-fetch'),
    ('okhttp',              'OkHttp'),
    ('httpx',               'httpx'),
    ('aiohttp',             'aiohttp'),
    ('Faraday',             'Faraday (Ruby)'),
]


# Pre-lowered once at import time so per-request matching doesn't re-lower
# every pattern string on every call.
_BOT_PATTERNS_LOWER = [(pattern.lower(), name) for pattern, name in BOT_PATTERNS]


def classify_ua(ua):
    if not ua or not ua.strip():
        return '(empty user agent)'
    ua_lower = ua.lower()
    for pattern, name in _BOT_PATTERNS_LOWER:
        if pattern in ua_lower:
            return name
    return 'Other / Browser'


# ---------------------------------------------------------------------------
# Vendors that publish their crawlers' IPs. A request naming one of these bots
# from anywhere else is an impersonator: 'Spoofed <vendor>'.
# vendor_ranges.tsv is generated by deploy/refresh_vendor_ranges.py. Without it
# (ws_service mounts this file alone) nothing is reported as spoofed.
# ---------------------------------------------------------------------------
_VENDOR_BOTS = {
    'Anthropic': ('Anthropic ClaudeBot', 'Anthropic Claude', 'Anthropic', 'Anthropic Claude-User',
                  'Anthropic Claude-SearchBot'),
    'Bing': ('Bingbot',),
    'DuckDuckGo': ('DuckDuckBot', 'DuckAssistBot'),
    'Google': ('Googlebot', 'Googlebot Mobile', 'GoogleOther', 'Google-Extended (AI)', 'Google-Agent',
               'AdsBot-Google'),
    'OpenAI': ('OpenAI GPTBot', 'OpenAI SearchBot', 'OpenAI ChatGPT', 'OpenAI AdsBot'),
    'Perplexity': ('Perplexity', 'Perplexity-User'),
}
_BOT_VENDOR = {bot: vendor for vendor, bots in _VENDOR_BOTS.items() for bot in bots}

# Never spoofed: dev machines, tests and anything else that isn't a public source address.
_LOCAL_NETWORKS = [ipaddress.ip_network(cidr) for cidr in (
    '10.0.0.0/8', '172.16.0.0/12', '192.168.0.0/16', '127.0.0.0/8', '169.254.0.0/16',
    'fc00::/7', '::1/128', 'fe80::/10',
)]


def _load_vendor_ranges():
    """{vendor: ([(version, first address)], [last address])}, sorted for bisect."""
    path = Path(__file__).with_name('vendor_ranges.tsv')
    if not path.exists():
        return {}
    ranges = {}
    for line in path.read_text().splitlines():
        if line and not line.startswith('#'):
            vendor, cidr = line.split('\t')
            net = ipaddress.ip_network(cidr)
            starts, ends = ranges.setdefault(vendor, ([], []))
            starts.append((net.version, int(net.network_address)))
            ends.append(int(net.broadcast_address))
    return ranges


_VENDOR_RANGES = _load_vendor_ranges()


def spoofed_vendor(bot_type, ip_str):
    """'Spoofed <vendor>' if bot_type is a verifiable vendor's bot and ip_str isn't theirs, else None."""
    vendor = _BOT_VENDOR.get(bot_type)
    if vendor not in _VENDOR_RANGES:
        return None
    try:
        addr = ipaddress.ip_address(ip_str.strip())
    except ValueError:
        return None
    addr = getattr(addr, 'ipv4_mapped', None) or addr
    if any(addr in net for net in _LOCAL_NETWORKS if net.version == addr.version):
        return None
    starts, ends = _VENDOR_RANGES[vendor]
    i = bisect.bisect_right(starts, (addr.version, int(addr))) - 1
    if i >= 0 and starts[i][0] == addr.version and int(addr) <= ends[i]:
        return None
    return f'Spoofed {vendor}'


def classify_ua_or_ip(ua, ip):
    """
    Classify by UA first; if that returns 'Other / Browser', fall back to IP range lookup.
    A vendor bot's name from outside the vendor's published ranges becomes 'Spoofed <vendor>'.
    Use this in preference to classify_ua() wherever the source IP is available.
    """
    result = classify_ua(ua)
    if result == 'Other / Browser':
        ip_result = classify_ip(ip)
        if ip_result:
            return ip_result
    return spoofed_vendor(result, ip) or result


_AI_BOTS = {
    'OpenAI GPTBot', 'OpenAI SearchBot', 'OpenAI ChatGPT', 'OpenAI AdsBot',
    'Anthropic ClaudeBot', 'Anthropic Claude', 'Anthropic',
    'Anthropic Claude-User', 'Anthropic Claude-SearchBot',
    'Perplexity', 'Perplexity-User', 'DuckAssistBot', 'Google-Agent',
    'Google-Extended (AI)', 'ByteDance Bytespider',
    'Meta FacebookBot', 'Meta ExternalAgent', 'Apple Applebot-Extended',
    'Amazonbot', 'Diffbot', 'Omgilibot', 'Webzio',
    'Common Crawl', 'Cohere', 'Timpi',
    'Meta WebIndexer', 'Parallel ShapBot', 'Reflection AI', 'You.com YouBot',
    'Keenable', 'Linkup',
    'Alibaba Qwen',  # IP-range classified
    'Tencent', # Tencent
}

_SEARCH_BOTS = {
    'Googlebot', 'Googlebot Mobile', 'GoogleOther',
    'Bingbot', 'Bing Preview', 'MSN Bot',
    'Baiduspider', 'YandexBot', 'Yahoo Slurp', 'DuckDuckBot',
    'Applebot', 'Sogou', '360Spider', 'Seznam', 'Bravebot',
}

_SCRAPER_BOTS = {
    'Python Requests', 'cURL', 'Wget', 'Scrapy', 'Go HTTP Client',
    'Java HTTP Client', 'libwww-perl', 'axios', 'node-fetch',
    'OkHttp', 'httpx', 'aiohttp', 'Faraday (Ruby)',
}


def bot_type_to_group(bot_type):
    """Map a bot_type label (from classify_ua_or_ip) to a coarse group."""
    if bot_type in ('Other / Browser', '(empty user agent)'):
        return bot_type
    if bot_type.startswith('Spoofed '):
        return 'Spoofed Vendors'
    if bot_type in _AI_BOTS:
        return 'AI Crawlers'
    if bot_type in _SEARCH_BOTS:
        return 'Search Engines'
    if bot_type in _SCRAPER_BOTS:
        return 'Generic Scrapers'
    return 'SEO / Other Bots'


def classify_ua_group(ua):
    """Coarser grouping for overview charts — UA only. Use bot_type_to_group() when IP is available."""
    return bot_type_to_group(classify_ua(ua))
