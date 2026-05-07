import re

from django.conf import settings
from django.http import HttpResponseRedirect

# Matches archives-YYYY.acpwb.com (production) and archives-YYYY.acpwb.example (local dev)
ARCHIVE_SUBDOMAIN_RE = re.compile(
    r'^archives-(\d{4})\.acpwb\.(com|example)(?::\d+)?$', re.IGNORECASE
)
ARCHIVE_VALID_YEARS = range(1985, 2026)

# Matches policy-<agency>.acpwb.com (production) and policy-<agency>.acpwb.example (local dev)
POLICY_SUBDOMAIN_RE = re.compile(
    r'^policy-([a-z0-9][a-z0-9\-]*)\.acpwb\.(com|example)(?::\d+)?$', re.IGNORECASE
)

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
        # DEBUG shortcut: ?__year=YYYY bypasses DNS for archive subdomains
        if settings.DEBUG:
            year_param = request.GET.get('__year')
            if year_param and year_param.isdigit():
                year = int(year_param)
                if year in ARCHIVE_VALID_YEARS:
                    request.archive_year = year
                    request.on_archive_subdomain = True
                    request.urlconf = 'apps.honeypot.archive_subdomain_urls'
                    return self.get_response(request)

            # DEBUG shortcut: ?__agency=<slug> bypasses DNS for policy subdomains
            agency_param = request.GET.get('__agency', '').lower().strip()
            if agency_param:
                from apps.honeypot.policy_data import AGENCIES
                if agency_param in AGENCIES:
                    request.policy_agency_slug = agency_param
                    request.on_policy_subdomain = True
                    request.urlconf = 'apps.honeypot.policy_subdomain_urls'
                    return self.get_response(request)

        # Strip port so localhost:8001 matches the same as localhost
        raw_host = request.get_host().lower()
        host = raw_host.rsplit(':', 1)[0] if ':' in raw_host else raw_host

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

        # Policy subdomain: policy-<agency>.acpwb.com or policy-<agency>.acpwb.example
        pm = POLICY_SUBDOMAIN_RE.match(host)
        if pm:
            from apps.honeypot.policy_data import AGENCIES
            agency_slug = pm.group(1).lower()
            if agency_slug in AGENCIES:
                request.policy_agency_slug = agency_slug
                request.on_policy_subdomain = True
                request.urlconf = 'apps.honeypot.policy_subdomain_urls'
                return self.get_response(request)

        # Unrecognized subdomain → redirect to appropriate main domain
        tld = 'example' if host.endswith('.acpwb.example') else 'com'
        main_domain = f'acpwb.{tld}'
        return HttpResponseRedirect(f'https://{main_domain}{request.get_full_path()}')
