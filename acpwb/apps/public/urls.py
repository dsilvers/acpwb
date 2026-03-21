from django.urls import path
from . import views

urlpatterns = [
    path('', views.home, name='home'),
    path('careers/', views.careers, name='careers'),
    path('awards/', views.awards, name='awards'),
    path('patents/', views.patents, name='patents'),
    path('mission/', views.mission, name='mission'),
    path('partners/', views.partners, name='partners'),
    path('privacy/', views.privacy, name='privacy'),
    path('privacy/do-not-sell/', views.do_not_sell, name='do-not-sell'),
    path('accessibility/', views.accessibility, name='accessibility'),
    path('trademarks/', views.trademarks, name='trademarks'),
    path('site-map/', views.sitemap_page, name='site-map'),
]
