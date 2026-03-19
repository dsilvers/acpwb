import pytest
from apps.core.middleware import BOT_UA_PATTERNS, VIEW_LOGGED_PATHS, BotTrackingMiddleware
from apps.honeypot.models import CrawlerVisit


# ── Pattern matching ───────────────────────────────────────────────────────────

@pytest.mark.parametrize('ua', [
    'Googlebot/2.1 (+http://www.google.com/bot.html)',
    'Mozilla/5.0 (compatible; bingbot/2.0)',
    'GPTBot/1.0 (+https://openai.com/gptbot)',
    'ClaudeBot/1.0',
    'python-requests/2.28.0',
    'scrapy/2.11.0',
    'curl/7.88.1',
    'wget/1.21',
    'Go-http-client/1.1',
    'Baiduspider/2.0',
    'Ahrefs/1.0',
])
def test_bot_ua_patterns_match(ua):
    assert BOT_UA_PATTERNS.search(ua) is not None, f"Should match bot UA: {ua}"


@pytest.mark.parametrize('ua', [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15',
    'Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0',
])
def test_bot_ua_patterns_no_match(ua):
    assert BOT_UA_PATTERNS.search(ua) is None, f"Should not match human UA: {ua}"


@pytest.mark.parametrize('path', [
    '/archive/2024/1/1/article/',
    '/wiki/governance/',
    '/internal/portal/',
    '/api/v1/private-data',
    '/.well-known/ai-agent.json',
    '/reports/some-report/',
    '/robots.txt',
])
def test_view_logged_paths_match(path):
    """These paths are handled by honeypot views — middleware should skip them."""
    assert VIEW_LOGGED_PATHS.match(path) is not None, f"Should match view-logged path: {path}"


@pytest.mark.parametrize('path', [
    '/',
    '/our-people/',
    '/careers/',
    '/mission/',
    '/projects/',
    '/partners/',
    '/privacy/',
    '/django-admin/',
])
def test_view_logged_paths_no_match(path):
    """Regular pages are not handled by honeypot views — middleware may log them."""
    assert VIEW_LOGGED_PATHS.match(path) is None, f"Should not match regular path: {path}"


# ── Middleware integration ─────────────────────────────────────────────────────

@pytest.mark.django_db
def test_bot_ua_logged_on_any_page(bot_client):
    count_before = CrawlerVisit.objects.count()
    bot_client.get('/')
    assert CrawlerVisit.objects.count() > count_before


@pytest.mark.django_db
def test_honeypot_path_logged_for_normal_ua(client):
    count_before = CrawlerVisit.objects.count()
    client.get('/api/v1/private-data')
    assert CrawlerVisit.objects.count() > count_before


@pytest.mark.django_db
def test_classify_path_archive():
    middleware = BotTrackingMiddleware(lambda r: None)
    assert middleware._classify_path('/archive/2024/1/1/x/') == 'archive'


@pytest.mark.django_db
def test_classify_path_wiki():
    middleware = BotTrackingMiddleware(lambda r: None)
    assert middleware._classify_path('/wiki/topic/') == 'wiki'


@pytest.mark.django_db
def test_classify_path_api():
    middleware = BotTrackingMiddleware(lambda r: None)
    assert middleware._classify_path('/api/v1/private-data') == 'api'


@pytest.mark.django_db
def test_classify_path_well_known():
    middleware = BotTrackingMiddleware(lambda r: None)
    assert middleware._classify_path('/.well-known/ai-agent.json') == 'well_known'


@pytest.mark.django_db
def test_classify_path_ghost_links():
    middleware = BotTrackingMiddleware(lambda r: None)
    assert middleware._classify_path('/internal/portal/') == 'ghost_link'
    assert middleware._classify_path('/employees/export/') == 'ghost_link'
    assert middleware._classify_path('/admin-panel/login/') == 'ghost_link'


@pytest.mark.django_db
def test_normal_request_not_logged(client):
    count_before = CrawlerVisit.objects.count()
    client.get('/careers/')
    # Normal UA + non-honeypot path should not be logged
    assert CrawlerVisit.objects.count() == count_before
