import re

from django.conf import settings
from django.http import HttpResponseRedirect

# Matches archives-YYYY.acpwb.com (production) and archives-YYYY.acpwb.example (local dev)
ARCHIVE_SUBDOMAIN_RE = re.compile(
    r'^archives-(\d{4})\.acpwb\.(com|example)(?::\d+)?$', re.IGNORECASE
)
ARCHIVE_VALID_YEARS = range(1985, 2026)

# Hosts treated as the main domain (no subdomain routing)
_MAIN_HOSTS = frozenset([
    'acpwb.com', 'www.acpwb.com',
    'acpwb.example', 'www.acpwb.example',
    'localhost', '127.0.0.1', 'testserver',
])


class SubdomainMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        # DEBUG shortcut: ?__year=YYYY bypasses DNS — works with localhost or any host
        if settings.DEBUG:
            year_param = request.GET.get('__year')
            if year_param and year_param.isdigit():
                year = int(year_param)
                if year in ARCHIVE_VALID_YEARS:
                    request.archive_year = year
                    request.on_archive_subdomain = True
                    request.urlconf = 'apps.honeypot.archive_subdomain_urls'
                    return self.get_response(request)

        host = request.get_host().lower()

        # Main domain and dev hosts — normal routing
        if (host in _MAIN_HOSTS
                or host.startswith('127.')
                or host.startswith('192.168.')):
            return self.get_response(request)

        # Archive subdomain: archives-YYYY.acpwb.com or archives-YYYY.acpwb.example
        m = ARCHIVE_SUBDOMAIN_RE.match(host)
        if m:
            year = int(m.group(1))
            if year in ARCHIVE_VALID_YEARS:
                request.archive_year = year
                request.on_archive_subdomain = True
                request.urlconf = 'apps.honeypot.archive_subdomain_urls'
                return self.get_response(request)

        # Unrecognized subdomain → redirect to appropriate main domain
        tld = m.group(2) if m else ('example' if host.endswith('.example') else 'com')
        main_domain = f'acpwb.{tld}'
        return HttpResponseRedirect(f'https://{main_domain}{request.get_full_path()}')
