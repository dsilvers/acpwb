from django.contrib import admin
from .models import DataOptOutRequest, Fortune500Company


@admin.register(DataOptOutRequest)
class DataOptOutRequestAdmin(admin.ModelAdmin):
    list_display = ('created_at', 'name', 'email', 'request_type', 'state', 'processed', 'ip_address')
    list_filter = ('request_type', 'processed', 'state')
    search_fields = ('name', 'email', 'ip_address')
    list_editable = ('processed',)
    readonly_fields = ('created_at', 'ip_address')
    ordering = ('-created_at',)


@admin.register(Fortune500Company)
class Fortune500CompanyAdmin(admin.ModelAdmin):
    list_display = ('rank', 'name', 'industry', 'website')
    list_filter = ('industry',)
    search_fields = ('name', 'industry')
    ordering = ('rank',)
