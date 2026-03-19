from django.urls import path, re_path
from . import views

urlpatterns = [
    # Archive labyrinth — captures any depth via <path:slug>
    path('archive/<int:year>/<int:month>/<int:day>/', views.archive_trap, name='archive-trap-base', kwargs={'slug': ''}),
    path('archive/<int:year>/<int:month>/<int:day>/<path:slug>/', views.archive_trap, name='archive-trap'),

    # Wiki trap
    path('wiki/<slug:slug>/', views.wiki_page, name='wiki-page'),

    # Fake API (listed in HTML comment — not linked in nav)
    path('api/v1/private-data', views.fake_api, name='fake-api'),

    # PoW endpoints
    path('api/pow/challenge/', views.pow_challenge_view, name='pow-challenge'),
    path('api/pow/verify/', views.pow_verify_view, name='pow-verify'),

    # Well-known files (nginx must proxy these to Django)
    path('.well-known/ai-agent.json', views.ai_agent_file, name='ai-agent-file'),
    path('.well-known/robots.txt', views.fake_robots, name='fake-robots'),

    # Ghost link traps
    path('internal/portal/', views.ghost_trap, name='ghost-portal'),
    path('employees/export/', views.ghost_trap, name='ghost-export'),
    path('admin-panel/login/', views.ghost_trap, name='ghost-admin'),
]
