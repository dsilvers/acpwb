import re

from django.contrib.auth.models import AnonymousUser

# Known bot/crawler user-agent patterns
BOT_UA_PATTERNS = re.compile(
    r'(bot|crawler|spider|scraper|crawl|fetch|wget|curl|python-requests|'
    r'scrapy|httpclient|java/|go-http|ruby|perl|libwww|mechanize|'
    r'gptbot|chatgpt|claude|anthropic|openai|bingbot|googlebot|'
    r'yandex|baidu|duckduck|semrush|ahrefs|moz\.com|dataprovider|'
    r'zgrab|masscan|nmap|nikto|nuclei)',
    re.IGNORECASE
)

# Paths where the view already calls _log_crawler — middleware must NOT also log
# these or every request creates two CrawlerVisit records.
VIEW_LOGGED_PATHS = re.compile(
    r'^/(archive|wiki|internal|employees/export|admin-panel|api/v1|'
    r'reports|sitemap|robots\.txt|\.well-known|datasets|feeds|process-improvement|presentations|'
    r'public-policy|company-handbooks)(/|$)'
)

# Subdomain views log their own CrawlerVisits — skip middleware logging
# for any request that SubdomainMiddleware has routed to a subdomain urlconf.
ARCHIVE_SUBDOMAIN_URLCONF = 'apps.honeypot.archive_subdomain_urls'
POLICY_SUBDOMAIN_URLCONF = 'apps.honeypot.policy_subdomain_urls'


_DB_REQUIRED = re.compile(r'^/(django-admin|acpwb-dashboard)(/|$)')


class _NoOpSession:
    modified = False
    accessed = False
    session_key = None

    def get(self, key, default=None): return default
    def __contains__(self, key): return False
    def __getitem__(self, key): raise KeyError(key)
    def __setitem__(self, key, value): pass
    def __delitem__(self, key): pass
    def flush(self): pass
    def cycle_key(self): pass
    def save(self, *args, **kwargs): pass
    def keys(self): return []


class ConditionalAuthMiddleware:
    """Runs Session+Auth+Message middleware only for paths that need DB access."""

    def __init__(self, get_response):
        from django.contrib.sessions.middleware import SessionMiddleware
        from django.contrib.auth.middleware import AuthenticationMiddleware
        from django.contrib.messages.middleware import MessageMiddleware
        self._full_chain = SessionMiddleware(AuthenticationMiddleware(MessageMiddleware(get_response)))
        self.get_response = get_response

    def __call__(self, request):
        if _DB_REQUIRED.match(request.path_info):
            return self._full_chain(request)
        request.session = _NoOpSession()
        request.user = AnonymousUser()
        return self.get_response(request)


class BotTrackingMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        user_agent = request.META.get('HTTP_USER_AGENT', '')
        path = request.path

        # Classify once per request and cache it on the request object so
        # RequestStreamMiddleware (which needs bot_type/bot_group for every
        # request anyway) doesn't redo the same UA/IP matching.
        try:
            from apps.core.bot_classify import bot_type_to_group, classify_ua_or_ip
            bot_type = classify_ua_or_ip(user_agent, self._get_ip(request))
            bot_group = bot_type_to_group(bot_type)
        except Exception:
            bot_type, bot_group = '', ''
        request._bot_classification = (bot_type, bot_group)

        # Log bot UA hits only on paths NOT already logged by a honeypot view.
        # Honeypot views call _log_crawler themselves with more specific trap types.
        # Also skip archive subdomain requests — those views log their own CrawlerVisits.
        if (BOT_UA_PATTERNS.search(user_agent)
                and not VIEW_LOGGED_PATHS.match(path)
                and getattr(request, 'urlconf', None) not in (ARCHIVE_SUBDOMAIN_URLCONF, POLICY_SUBDOMAIN_URLCONF)):
            self._log_bot_visit(request, user_agent, path, bot_type, bot_group)

        response = self.get_response(request)
        return response

    def _log_bot_visit(self, request, user_agent, path, bot_type, bot_group):
        # Deferred imports to avoid circular issues at middleware load time
        try:
            from django.utils import timezone
            from apps.core.crawler_queue import queue_crawler_visit
            ip = self._get_ip(request)
            trap_type = self._classify_path(path)
            ua = user_agent or ''
            data = {
                'timestamp': timezone.now().isoformat(),
                'ip_address': ip,
                'user_agent': ua[:512],
                'path': path[:512],
                'referrer': request.META.get('HTTP_REFERER', '')[:256],
                'trap_type': trap_type,
                'query_string': request.META.get('QUERY_STRING', '')[:256],
                'bot_type': bot_type,
                'bot_group': bot_group,
            }
            queue_crawler_visit(data)
        except Exception:
            pass  # Never let honeypot logging break the response

    def _get_ip(self, request):
        x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
        if x_forwarded:
            return x_forwarded.split(',')[0].strip()
        return request.META.get('REMOTE_ADDR', '0.0.0.0')

    def _classify_path(self, path):
        if path.startswith('/archive/'):
            return 'archive'
        if path.startswith('/wiki/'):
            return 'wiki'
        if path.startswith('/api/v1/'):
            return 'api'
        if path.startswith('/.well-known/'):
            return 'well_known'
        if path.startswith('/internal/') or path.startswith('/employees/') or path.startswith('/admin-panel/'):
            return 'ghost_link'
        return 'other'
