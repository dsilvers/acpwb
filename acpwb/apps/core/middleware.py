import re

# Known bot/crawler user-agent patterns
BOT_UA_PATTERNS = re.compile(
    r'(bot|crawler|spider|scraper|crawl|fetch|wget|curl|python-requests|'
    r'scrapy|httpclient|java/|go-http|ruby|perl|libwww|mechanize|'
    r'gptbot|chatgpt|claude|anthropic|openai|bingbot|googlebot|'
    r'yandex|baidu|duckduck|semrush|ahrefs|moz\.com|dataprovider|'
    r'zgrab|masscan|nmap|nikto|nuclei)',
    re.IGNORECASE
)

# Honeypot paths that should always be logged regardless of UA
HONEYPOT_PATHS = re.compile(
    r'^/(archive|wiki|internal|employees/export|admin-panel|api/v1|\.well-known)/'
)


class BotTrackingMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        user_agent = request.META.get('HTTP_USER_AGENT', '')
        path = request.path

        # Log bot UA hits on any page, or any visit to honeypot paths
        if BOT_UA_PATTERNS.search(user_agent) or HONEYPOT_PATHS.match(path):
            self._log_bot_visit(request, user_agent, path)

        response = self.get_response(request)
        return response

    def _log_bot_visit(self, request, user_agent, path):
        # Deferred import to avoid circular issues at middleware load time
        try:
            from apps.honeypot.models import CrawlerVisit
            ip = self._get_ip(request)
            trap_type = self._classify_path(path)
            CrawlerVisit.objects.create(
                ip_address=ip,
                user_agent=user_agent[:512],
                path=path[:512],
                referrer=request.META.get('HTTP_REFERER', '')[:256],
                trap_type=trap_type,
                query_string=request.META.get('QUERY_STRING', '')[:256],
            )
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
