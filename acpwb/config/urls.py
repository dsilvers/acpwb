from django.contrib import admin
from django.contrib.sitemaps.views import sitemap
from django.urls import path, include
from django.views.generic import RedirectView

from apps.public.sitemaps import StaticPagesSitemap, ProjectStorySitemap

_sitemaps = {
    'static': StaticPagesSitemap,
    'projects': ProjectStorySitemap,
}

urlpatterns = [
    path('django-admin/', admin.site.urls),
    path('sitemap.xml', sitemap, {'sitemaps': _sitemaps}, name='django-sitemap'),

    # Legacy PHP URL redirects (permanent 301)
    path('index.php', RedirectView.as_view(url='/', permanent=True)),
    path('profiles.php', RedirectView.as_view(url='/our-people/', permanent=True)),
    path('mission.php', RedirectView.as_view(url='/mission/', permanent=True)),
    path('disclaimer.php', RedirectView.as_view(url='/privacy/', permanent=True)),

    # Human-facing pages
    path('', include('apps.public.urls')),
    path('our-people/', include('apps.people.urls')),
    path('contact/', RedirectView.as_view(url='/our-people/', permanent=False)),
    path('projects/', include('apps.projects.urls')),

    # Honeypot infrastructure
    path('', include('apps.honeypot.urls')),

    # Mailgun webhook
    path('webhooks/', include('apps.webhooks.urls')),
]

# Customize admin site
admin.site.site_header = "ACPWB Administration"
admin.site.site_title = "ACPWB Admin"
admin.site.index_title = "American Corporation for Public Well Being"
