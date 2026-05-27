from django.urls import path
from . import views

urlpatterns = [
    path('', views.process_index, name='process-improvement-index'),
    path('<slug:instance>/', views.process_category, name='process-improvement-category'),
    path('<slug:instance>/<int:year>/', views.process_year, name='process-improvement-year'),
    path('<slug:instance>/<int:year>/page/<int:page>/', views.process_year_page, name='process-improvement-page'),
    path('<slug:instance>/<int:year>/<slug:initiative_slug>/', views.process_detail, name='process-improvement-detail'),
]
