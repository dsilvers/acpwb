"""URL conf for archives-YYYY.acpwb.com subdomains.

Year is baked into the host — paths start at the subdomain root.
The main site URL conf is included at the end so that {% url 'home' %} and
other named URLs in base.html continue to resolve correctly on subdomain pages.
"""
from django.urls import include, path
from . import views

urlpatterns = [
    path('robots.txt', views.archive_subdomain_robots, name='archive-sub-robots'),
    path('', views.archive_subdomain_index, name='archive-sub-index'),
    path('<int:month>/', views.archive_month, name='archive-sub-month'),
    path('<int:month>', views.archive_month),
    path('<int:month>/<int:day>/', views.archive_trap, kwargs={'slug': ''}, name='archive-sub-trap-base'),
    path('<int:month>/<int:day>/<path:slug>/', views.archive_trap, name='archive-sub-trap'),
    path('<int:month>/<int:day>', views.archive_trap, kwargs={'slug': ''}),
    path('<int:month>/<int:day>/<path:slug>', views.archive_trap),
    path('<int:month>/<int:day>/<path:slug>/export.csv', views.archive_export_csv, name='archive-sub-export'),
    path('<int:month>/<int:day>/export.csv', views.archive_export_csv, kwargs={'slug': ''}, name='archive-sub-export-base'),
    # Catch-all: any non-archive path (e.g. /mission/, /reports/) → redirect to main domain.
    # Must come before include('config.urls') so requests are redirected, not served.
    # URL reversal ({% url 'home' %} etc.) still works because config.urls is still included
    # and reversal scans all patterns regardless of order.
    path('<path:rest>', views.archive_subdomain_non_archive_redirect),
    # Include main site URLs so named URLs (home, careers, etc.) resolve in templates
    path('', include('config.urls')),
]
