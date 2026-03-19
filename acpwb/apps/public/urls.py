from django.urls import path
from . import views

urlpatterns = [
    path('', views.home, name='home'),
    path('careers/', views.careers, name='careers'),
    path('mission/', views.mission, name='mission'),
    path('partners/', views.partners, name='partners'),
    path('privacy/', views.privacy, name='privacy'),
]
