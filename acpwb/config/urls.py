from django.contrib import admin
from django.contrib.sitemaps.views import sitemap
from django.urls import path, include
from django.views.decorators.cache import cache_page
from django.views.generic import RedirectView

from apps.public.sitemaps import StaticPagesSitemap, ProjectStorySitemap
from apps.presentations.sitemaps import PresentationOrgSitemap, PresentationDetailSitemap

# Rebuilt from every ProjectStory row on each request (~800 KB, up to ~2s);
# crawlers fetch it hundreds of times an hour. Cached per worker for an
# hour, so new projects appear within the hour.
sitemap = cache_page(60 * 60)(sitemap)

_sitemaps = {
    'static': StaticPagesSitemap,
    'projects': ProjectStorySitemap,
}

urlpatterns = [
    path('django-admin/', admin.site.urls),
    path('sitemap.xml', sitemap, {'sitemaps': _sitemaps}, name='django-sitemap'),
    path('sitemap-pages.xml', sitemap, {'sitemaps': {'static': StaticPagesSitemap}}, name='django-sitemap-pages'),
    path('sitemap-presentations.xml', sitemap, {'sitemaps': {'orgs': PresentationOrgSitemap, 'detail': PresentationDetailSitemap}}, name='django-sitemap-presentations'),

    # Legacy PHP URL redirects (permanent 301)
    path('index.php', RedirectView.as_view(url='/', permanent=True)),
    # /compliance/ itself is served by acpwb_go (edge prefix route); only the
    # slashless form reaches Django.
    path('compliance', RedirectView.as_view(url='/compliance/', permanent=True)),
    path('profiles.php', RedirectView.as_view(url='/our-people/', permanent=True)),
    path('mission.php', RedirectView.as_view(url='/mission/', permanent=True)),
    path('disclaimer.php', RedirectView.as_view(url='/privacy/', permanent=True)),

    # Human-facing pages
    path('', include('apps.public.urls')),
    path('our-people/', include('apps.people.urls')),
    path('projects/', include('apps.projects.urls')),
    path('company-handbooks/', include('apps.company_handbooks.urls')),
    path('process-improvement/', include('apps.process_improvement.urls')),
    path('presentations/', include('apps.presentations.urls')),

    # Honeypot infrastructure
    path('', include('apps.honeypot.urls')),

    # Mailgun webhook
    path('webhooks/', include('apps.webhooks.urls')),

    # Staff dashboard
    path('acpwb-dashboard/', include('apps.core.dashboard_urls')),

    # Traffic-intelligence product API — see PublishedIPReputation
    path('reputation-api/', include('apps.core.reputation_api_urls')),
]

handler404 = 'apps.honeypot.views.scanner_probe_404'

# Customize admin site
admin.site.site_header = "ACPWB Administration"
admin.site.site_title = "ACPWB Admin"
admin.site.index_title = "American Corporation for Public Well Being"
