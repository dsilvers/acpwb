from django.urls import path
from . import views

urlpatterns = [
    path('', views.people_page, name='our-people'),
]
