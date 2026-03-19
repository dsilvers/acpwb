from django.urls import path
from . import views

urlpatterns = [
    path('mailgun/inbound/', views.mailgun_inbound, name='mailgun-inbound'),
]
