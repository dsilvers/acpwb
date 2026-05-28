from django.urls import path
from . import views

urlpatterns = [
    path('', views.handbook_index, name='handbook-index'),
    path('<slug:instance>/', views.handbook_agency, name='handbook-agency'),
    path('<slug:instance>/<int:year>/', views.handbook_year, name='handbook-year'),
    path('<slug:instance>/<int:year>/rev/<int:revision>/', views.handbook_revision, name='handbook-revision'),
    path('<slug:instance>/<int:year>/rev/<int:revision>/<slug:group_slug>/', views.handbook_group, name='handbook-group'),
    path('<slug:instance>/<int:year>/rev/<int:revision>/<slug:group_slug>/page/<int:page>/', views.handbook_group, name='handbook-group-page'),
]
