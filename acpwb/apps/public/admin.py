from django.contrib import admin
from .models import Fortune500Company


@admin.register(Fortune500Company)
class Fortune500CompanyAdmin(admin.ModelAdmin):
    list_display = ('rank', 'name', 'industry', 'website')
    list_filter = ('industry',)
    search_fields = ('name', 'industry')
    ordering = ('rank',)
