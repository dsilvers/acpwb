from django.urls import path
from . import dashboard_views as views

urlpatterns = [
    path('',         views.overview, name='dashboard-overview'),
    path('crawlers/', views.crawlers, name='dashboard-crawlers'),
    path('archive/',  views.archive,  name='dashboard-archive'),
    path('emails/',   views.emails,   name='dashboard-emails'),
    path('people/',   views.people,   name='dashboard-people'),
]
