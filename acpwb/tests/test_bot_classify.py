import pytest
from apps.core.bot_classify import classify_ua, classify_ua_group


@pytest.mark.parametrize("ua,expected", [
    # AI crawlers
    ('GPTBot/1.0', 'OpenAI GPTBot'),
    ('OAI-SearchBot/1.0', 'OpenAI SearchBot'),
    ('ChatGPT-User/1.0', 'OpenAI ChatGPT'),
    ('ClaudeBot/0.5', 'Anthropic ClaudeBot'),
    ('Claude-Web/1.0', 'Anthropic Claude'),
    ('anthropic-ai/1.0', 'Anthropic'),
    ('PerplexityBot/1.0', 'Perplexity'),
    ('meta-externalagent/1.1 (+https://example.com/)', 'Meta ExternalAgent'),
    ('Amazonbot/0.1', 'Amazonbot'),
    ('Diffbot/3.0', 'Diffbot'),
    ('omgili/0.5 +http://omgili.com', 'Omgilibot'),
    ('CCBot/2.0', 'Common Crawl'),
    ('cohere-ai/1.0', 'Cohere'),
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ShapBot/0.1.0', 'Parallel ShapBot'),
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Reflectionbot/1.0; +https://reflection.ai/bot) Chrome/151.0.0.0 Safari/537.36', 'Reflection AI'),
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; YouBot/1.0; +https://docs.you.com/youbot; env:prod) Chrome/142.0.0.0 Safari/537.36', 'You.com YouBot'),
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; KeenableBot/1.0; +https://keenable.ai/bot)', 'Keenable'),
    ('LinkupBot/1.0 (LinkupBot for web indexing; https://linkup.so/bot; bot@linkup.so)', 'Linkup'),
    ('Mozilla/5.0 (compatible; Meta-WebIndexer/1.0; +https://developers.facebook.com/docs/sharing/webmasters/crawler)', 'Meta WebIndexer'),
    # Search engines
    ('Mozilla/5.0 (compatible; Bravebot/1.0; +https://brave.com/search/)', 'Bravebot'),
    ('Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36', 'Googlebot Mobile'),
    ('Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)', 'Googlebot'),
    ('GoogleOther/1.0', 'GoogleOther'),
    ('Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)', 'Bingbot'),
    ('Mozilla/5.0 (compatible; YandexBot/3.0)', 'YandexBot'),
    ('DuckDuckBot/1.0', 'DuckDuckBot'),
    # SEO/marketing
    ('SemrushBot/7~bl', 'SemrushBot'),
    ('AhrefsBot/7.0', 'AhrefsBot'),
    ('SERankingBacklinksBot/1.0', 'SERankingBacklinksBot'),
    ('ZoominfoBot/1.0', 'ZoominfoBot'),
    ('BitSightBot/1.0', 'BitSightBot'),
    ('zgrab/0.x', 'zgrab'),
    ('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36 (compatible; AionBot/1.0)', 'AionBot'),
    ('Hello from Palo Alto Networks, find out more about our scans in https://docs-cortex.paloaltonetworks.com/r/1/Cortex-Xpanse/Scanning-activity', 'Palo Alto Xpanse'),
    ('AdsBot-Google (+http://www.google.com/adsbot.html)', 'AdsBot-Google'),
    ('Mozilla/5.0 (Linux; Android 5.0) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; TikTokSpider; ttspider-feedback@tiktok.com)', 'TikTokSpider'),
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; HubSpot Crawler; +https://www.hubspot.com) Chrome/131.0.0.0 Safari/537.36', 'HubSpot Crawler'),
    ('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko; compatible; BuiltWith/1.4; rb.gy/xprgqj) Chrome/131.0.0.0 Safari/537.36', 'BuiltWith'),
    ('TechnicalSEO.com Spider Bot/2.0', 'TechnicalSEO'),
    ('pimeyes-downloader-api/0.1', 'PimEyes'),
    # Vendor names being spoofed on this site stay unclassified until real traffic shows up
    ('Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexity-user)', 'Other / Browser'),
    ('DuckAssistBot/1.1 (https://duckduckgo.com/duckassistbot)', 'Other / Browser'),
    # Generic scrapers
    ('python-requests/2.28.0', 'Python Requests'),
    ('curl/7.88.1', 'cURL'),
    ('Wget/1.21.1', 'Wget'),
    ('Scrapy/2.8.0 (+https://scrapy.org)', 'Scrapy'),
    ('Go-http-client/1.1', 'Go HTTP Client'),
    ('axios/1.3.4', 'axios'),
    ('node-fetch/2.6.9', 'node-fetch'),
    ('okhttp/4.10.0', 'OkHttp'),
    ('httpx/0.24.1', 'httpx'),
    ('aiohttp/3.8.4', 'aiohttp'),
    # Edge cases
    ('', '(empty user agent)'),
    ('   ', '(empty user agent)'),
    # Regular browser → no match
    ('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36', 'Other / Browser'),
])
def test_classify_ua(ua, expected):
    assert classify_ua(ua) == expected


def test_classify_ua_case_insensitive():
    """Pattern matching is case-insensitive."""
    assert classify_ua('GPTBOT/1.0') == 'OpenAI GPTBot'
    assert classify_ua('claudebot') == 'Anthropic ClaudeBot'


@pytest.mark.parametrize("ua,expected_group", [
    # AI crawlers
    ('GPTBot/1.0', 'AI Crawlers'),
    ('ClaudeBot/0.5', 'AI Crawlers'),
    ('anthropic-ai/1.0', 'AI Crawlers'),
    ('meta-externalagent/1.1', 'AI Crawlers'),
    ('Amazonbot/0.1', 'AI Crawlers'),
    ('Diffbot/3.0', 'AI Crawlers'),
    ('omgili/0.5', 'AI Crawlers'),
    ('CCBot/2.0', 'AI Crawlers'),
    ('cohere-ai/1.0', 'AI Crawlers'),
    ('ShapBot/0.1.0', 'AI Crawlers'),
    ('Reflectionbot/1.0', 'AI Crawlers'),
    ('YouBot/1.0', 'AI Crawlers'),
    ('KeenableBot/1.0', 'AI Crawlers'),
    ('LinkupBot/1.0', 'AI Crawlers'),
    ('meta-webindexer/1.1', 'AI Crawlers'),
    # Search engines
    ('Bravebot/1.0', 'Search Engines'),
    ('Mozilla/5.0 (compatible; Googlebot/2.1)', 'Search Engines'),
    ('Nexus 5X Build/MMB29P AppleWebKit/537.36', 'Search Engines'),
    ('GoogleOther/1.0', 'Search Engines'),
    ('bingbot/2.0', 'Search Engines'),
    ('YandexBot/3.0', 'Search Engines'),
    ('DuckDuckBot/1.0', 'Search Engines'),
    # Generic scrapers
    ('python-requests/2.28', 'Generic Scrapers'),
    ('curl/7.88.1', 'Generic Scrapers'),
    ('Wget/1.21', 'Generic Scrapers'),
    ('Go-http-client/1.1', 'Generic Scrapers'),
    ('axios/1.3.4', 'Generic Scrapers'),
    # SEO / Other
    ('SemrushBot/7', 'SEO / Other Bots'),
    ('AhrefsBot/7.0', 'SEO / Other Bots'),
    ('SERankingBacklinksBot/1.0', 'SEO / Other Bots'),
    ('zgrab/0.x', 'SEO / Other Bots'),
    ('AionBot/1.0', 'SEO / Other Bots'),
    # Pass-through groups
    ('', '(empty user agent)'),
    ('Mozilla/5.0 ... Chrome/120', 'Other / Browser'),
])
def test_classify_ua_group(ua, expected_group):
    assert classify_ua_group(ua) == expected_group
