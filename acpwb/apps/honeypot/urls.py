from django.urls import path
from . import views

urlpatterns = [
    # Archive labyrinth — both slash and non-slash variants to avoid 301 churn
    path('archive/', views.archive_index, name='archive-index'),
    path('archive/<int:year>/', views.archive_year, name='archive-year'),
    path('archive/<int:year>', views.archive_year),
    path('archive/<int:year>/<int:month>/', views.archive_month, name='archive-month'),
    path('archive/<int:year>/<int:month>', views.archive_month),
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

    # Internal portal (accidentally exposed intranet)
    path('internal/', views.internal_portal, name='internal-portal'),
    path('internal/login/', views.internal_login, name='internal-login'),
    path('internal/employee-records/', views.internal_employee_records, name='internal-emp-records'),
    path('internal/employee-records/export.csv', views.internal_employee_records_csv, name='internal-emp-csv'),
    path('internal/salary-database/', views.internal_salary_database, name='internal-salary-db'),
    path('internal/salary-database/export.csv', views.internal_salary_database_csv, name='internal-salary-csv'),
    path('internal/acquisition-targets/', views.internal_acquisition_targets, name='internal-acq'),
    path('internal/acquisition-targets/export.csv', views.internal_acquisition_targets_csv, name='internal-acq-csv'),
    path('internal/litigation-hold/', views.internal_litigation_hold, name='internal-lit-hold'),

    # Archive CSV export
    path('archive/<int:year>/<int:month>/<int:day>/<path:slug>/export.csv', views.archive_export_csv, name='archive-export-csv'),
    path('archive/<int:year>/<int:month>/<int:day>/export.csv', views.archive_export_csv, kwargs={'slug': ''}, name='archive-export-csv-base'),

    # RSS / Atom feeds
    path('feeds/', views.feeds_index, name='feeds-index'),
    path('feeds/archive.xml', views.feed_archive, name='feed-archive'),
    path('feeds/reports.xml', views.feed_reports, name='feed-reports'),

    # API v1
    path('api/v1/', views.api_v1_index, name='api-v1-index'),
    path('api/v1/openapi.json', views.openapi_spec, name='openapi-spec'),

    # Training datasets
    path('datasets/', views.datasets_index, name='datasets-index'),
    path('datasets/<slug:slug>/', views.dataset_detail, name='dataset-detail'),
    path('datasets/<slug:slug>/data.jsonl', views.dataset_download, name='dataset-download'),
]
