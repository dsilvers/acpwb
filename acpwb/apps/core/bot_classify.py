"""
Bot user-agent classification utilities.
Shared between BotTrackingMiddleware and dashboard_views to avoid circular imports.

Two classification paths:
  classify_ua(ua)            — match by user-agent string against BOT_PATTERNS
  classify_ip(ip)            — match by source IP against known bot CIDR ranges
  classify_ua_or_ip(ua, ip)  — UA first; falls back to IP when UA returns 'Other / Browser'
"""
import ipaddress

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
    ('ClaudeBot',           'Anthropic ClaudeBot'),
    ('Claude-Web',          'Anthropic Claude'),
    ('anthropic-ai',        'Anthropic'),
    ('PerplexityBot',       'Perplexity'),
    ('Google-Extended',     'Google-Extended (AI)'),
    ('meta-externalagent',  'Meta ExternalAgent'),
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


def classify_ua(ua):
    if not ua or not ua.strip():
        return '(empty user agent)'
    for pattern, name in BOT_PATTERNS:
        if pattern.lower() in ua.lower():
            return name
    return 'Other / Browser'


def classify_ua_or_ip(ua, ip):
    """
    Classify by UA first; if that returns 'Other / Browser', fall back to IP range lookup.
    Use this in preference to classify_ua() wherever the source IP is available.
    """
    result = classify_ua(ua)
    if result == 'Other / Browser':
        ip_result = classify_ip(ip)
        if ip_result:
            return ip_result
    return result


_AI_BOTS = {
    'OpenAI GPTBot', 'OpenAI SearchBot', 'OpenAI ChatGPT',
    'Anthropic ClaudeBot', 'Anthropic Claude', 'Anthropic',
    'Perplexity', 'Google-Extended (AI)', 'ByteDance Bytespider',
    'Meta FacebookBot', 'Meta ExternalAgent', 'Apple Applebot-Extended',
    'Amazonbot', 'Diffbot', 'Omgilibot', 'Webzio',
    'Common Crawl', 'Cohere', 'Timpi',
    'Alibaba Qwen',  # IP-range classified
    'Tencent', # Tencent
}

_SEARCH_BOTS = {
    'Googlebot', 'Googlebot Mobile', 'GoogleOther',
    'Bingbot', 'Bing Preview', 'MSN Bot',
    'Baiduspider', 'YandexBot', 'Yahoo Slurp', 'DuckDuckBot',
    'Applebot', 'Sogou', '360Spider', 'Seznam',
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
