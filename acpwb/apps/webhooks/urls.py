from django.urls import path
from . import views

urlpatterns = [
    path('mailgun/inbound/', views.mailgun_inbound, name='mailgun-inbound'),
    path('pipe/inbound/', views.pipe_inbound, name='pipe-inbound'),
    path('canary-trigger/', views.canary_trigger_webhook, name='canary-trigger'),
]
