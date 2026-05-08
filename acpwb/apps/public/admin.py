from django.contrib import admin
from .models import ConferenceRegistration, DataOptOutRequest, Fortune500Company


@admin.register(DataOptOutRequest)
class DataOptOutRequestAdmin(admin.ModelAdmin):
    list_display = ('created_at', 'name', 'email', 'request_type', 'state', 'processed', 'ip_address')
    list_filter = ('request_type', 'processed', 'state')
    search_fields = ('name', 'email', 'ip_address')
    list_editable = ('processed',)
    readonly_fields = ('created_at', 'ip_address')
    ordering = ('-created_at',)


@admin.register(ConferenceRegistration)
class ConferenceRegistrationAdmin(admin.ModelAdmin):
    list_display = ('created_at', 'full_name', 'email', 'organization', 'registration_type', 'year', 'ip_address')
    list_filter = ('year', 'registration_type')
    search_fields = ('first_name', 'last_name', 'email', 'organization')
    readonly_fields = ('created_at', 'ip_address', 'user_agent')
    ordering = ('-created_at',)


@admin.register(Fortune500Company)
class Fortune500CompanyAdmin(admin.ModelAdmin):
    list_display = ('rank', 'name', 'industry', 'website')
    list_filter = ('industry',)
    search_fields = ('name', 'industry')
    ordering = ('rank',)
