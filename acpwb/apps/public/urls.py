from django.urls import path
from . import views

urlpatterns = [
    path('', views.home, name='home'),
    path('careers/', views.careers, name='careers'),
    path('careers/jobs/<int:job_id>/<str:job_slug>/', views.job_detail, name='job-detail'),
    path('careers/jobs/<int:job_id>/<str:job_slug>/apply/', views.job_apply, name='job-apply'),
    path('careers/jobs/<int:job_id>/<str:job_slug>/applied/', views.job_applied, name='job-applied'),
    path('awards/', views.awards, name='awards'),
    path('patents/', views.patents, name='patents'),
    path('mission/', views.mission, name='mission'),
    path('partners/', views.partners, name='partners'),
    path('privacy/', views.privacy, name='privacy'),
    path('privacy/do-not-sell/', views.do_not_sell, name='do-not-sell'),
    path('accessibility/', views.accessibility, name='accessibility'),
    path('trademarks/', views.trademarks, name='trademarks'),
    path('site-map/', views.sitemap_page, name='site-map'),
    path('faq/', views.faq, name='faq'),
    path('contact/', views.contact, name='contact'),
    path('press-releases/', views.press_releases, name='press_releases'),
    path('press-releases/<int:year>/<int:month>/<int:day>/<slug:slug>/', views.press_release_detail, name='press_release_detail'),
]
