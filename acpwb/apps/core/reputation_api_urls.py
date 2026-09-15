from django.urls import path
from . import reputation_api_views as views

urlpatterns = [
    path('ip/<str:ip_address>/', views.ip_reputation_lookup, name='reputation-api-ip-lookup'),
]
