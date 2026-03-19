from django.urls import path
from . import views

urlpatterns = [
    # Archive labyrinth — both slash and non-slash variants to avoid 301 churn
    path('archive/<int:year>/<int:month>/<int:day>/', views.archive_trap, name='archive-trap-base', kwargs={'slug': ''}),
    path('archive/<int:year>/<int:month>/<int:day>/<path:slug>/', views.archive_trap, name='archive-trap'),
    path('archive/<int:year>/<int:month>/<int:day>', views.archive_trap, kwargs={'slug': ''}),
    path('archive/<int:year>/<int:month>/<int:day>/<path:slug>', views.archive_trap),

    # Wiki trap
    path('wiki/<slug:slug>/', views.wiki_page, name='wiki-page'),

    # Fake API (listed in HTML comment — not linked in nav)
    path('api/v1/private-data', views.fake_api, name='fake-api'),

    # PoW endpoints
    path('api/pow/challenge/', views.pow_challenge_view, name='pow-challenge'),
    path('api/pow/verify/', views.pow_verify_view, name='pow-verify'),

    # Well-known files (nginx must proxy these to Django)
    path('robots.txt', views.fake_robots, name='robots-txt'),
    path('.well-known/ai-agent.json', views.ai_agent_file, name='ai-agent-file'),
    path('.well-known/robots.txt', views.fake_robots, name='fake-robots'),

    # Reports & Publications
    path('reports/', views.reports_list, name='reports-list'),
    path('reports/page/<int:page>/', views.reports_page_api, name='reports-page-api'),
    path('reports/<slug:slug>/download.csv', views.report_download, name='report-download'),
    path('reports/<slug:slug>/download.pdf', views.report_download_pdf, name='report-download-pdf'),
    path('reports/<slug:slug>/', views.report_detail, name='report-detail'),

    # Trap sitemaps (referenced in robots.txt)
    path('sitemap-publications.xml', views.sitemap_publications, name='sitemap-publications'),
    path('sitemap-wiki.xml',     views.sitemap_wiki,     name='sitemap-wiki'),
    path('sitemap-archive.xml',  views.sitemap_archive,  name='sitemap-archive'),

    # Ghost link traps
    path('internal/portal/', views.ghost_trap, name='ghost-portal'),
    path('employees/export/', views.ghost_trap, name='ghost-export'),
    path('admin-panel/login/', views.ghost_trap, name='ghost-admin'),
]
