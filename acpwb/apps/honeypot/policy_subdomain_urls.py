"""URL conf for policy-<agency>.acpwb.com subdomains.

Agency slug is baked into the host — paths are YYYY/MM/DD/<slug>/ from the subdomain root.
The main site URL conf is included at the end so {% url 'home' %} and other named URLs
in base.html continue to resolve correctly on subdomain pages.
"""
from django.urls import include, path
from . import views

urlpatterns = [
    path('robots.txt', views.policy_subdomain_robots, name='policy-sub-robots'),
    path('sitemap.xml', views.policy_subdomain_sitemap, name='policy-sub-sitemap'),
    path('', views.policy_subdomain_index, name='policy-sub-index'),
    path('<int:year>/', views.policy_subdomain_year, name='policy-sub-year'),
    path('<int:year>/<int:month>/', views.policy_subdomain_month, name='policy-sub-month'),
    path('<int:year>/<int:month>/<int:day>/<slug:slug>/', views.policy_subdomain_detail, name='policy-sub-detail'),
    # Catch-all: any non-policy path → redirect to main domain.
    # Must come before include('config.urls') so requests redirect instead of being served.
    path('<path:rest>', views.policy_subdomain_redirect),
    # Include main site URLs so named URLs (home, careers, etc.) resolve in templates.
    path('', include('config.urls')),
]
