from django.contrib import admin
from django.urls import path, include
from django.views.generic import RedirectView

urlpatterns = [
    path('django-admin/', admin.site.urls),

    # Human-facing pages
    path('', include('apps.public.urls')),
    path('our-people/', include('apps.people.urls')),
    path('contact/', RedirectView.as_view(url='/our-people/', permanent=False)),
    path('projects/', include('apps.projects.urls')),

    # Honeypot infrastructure
    path('', include('apps.honeypot.urls')),

    # Mailgun webhook
    path('webhooks/', include('apps.webhooks.urls')),
]

# Customize admin site
admin.site.site_header = "ACPWB Administration"
admin.site.site_title = "ACPWB Admin"
admin.site.index_title = "American Corporation for Public Well Being"
