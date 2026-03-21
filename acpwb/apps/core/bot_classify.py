"""
Bot user-agent classification utilities.
Shared between BotTrackingMiddleware and dashboard_views to avoid circular imports.
"""

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
    ('FacebookBot',         'Meta FacebookBot'),
    ('Applebot-Extended',   'Apple Applebot-Extended'),
    ('Bytespider',          'ByteDance Bytespider'),
    # Search engines
    ('Googlebot',           'Googlebot'),
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


def classify_ua_group(ua):
    """Coarser grouping for overview charts."""
    label = classify_ua(ua)
    if label in ('Other / Browser', '(empty user agent)'):
        return label
    ai_bots = {'OpenAI GPTBot', 'OpenAI SearchBot', 'OpenAI ChatGPT',
                'Anthropic ClaudeBot', 'Anthropic Claude', 'Anthropic',
                'Perplexity', 'Google-Extended (AI)', 'ByteDance Bytespider',
                'Meta FacebookBot', 'Apple Applebot-Extended'}
    if label in ai_bots:
        return 'AI Crawlers'
    search_bots = {'Googlebot', 'Bingbot', 'Bing Preview', 'MSN Bot',
                   'Baiduspider', 'YandexBot', 'Yahoo Slurp', 'DuckDuckBot',
                   'Applebot', 'Sogou', '360Spider', 'Seznam'}
    if label in search_bots:
        return 'Search Engines'
    scraper_bots = {'Python Requests', 'cURL', 'Wget', 'Scrapy', 'Go HTTP Client',
                    'Java HTTP Client', 'libwww-perl', 'axios', 'node-fetch',
                    'OkHttp', 'httpx', 'aiohttp', 'Faraday (Ruby)'}
    if label in scraper_bots:
        return 'Generic Scrapers'
    return 'SEO / Other Bots'
