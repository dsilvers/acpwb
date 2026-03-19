from django.contrib import admin
from .models import ProjectStory, ProjectPageVisit


@admin.register(ProjectStory)
class ProjectStoryAdmin(admin.ModelAdmin):
    list_display = ('title', 'industry_tag', 'impact_metric', 'page_number', 'generated_at')
    list_filter = ('industry_tag', 'page_number')
    search_fields = ('title', 'summary', 'slug')
    readonly_fields = ('generated_at', 'slug')
    ordering = ('page_number', 'slug')


@admin.register(ProjectPageVisit)
class ProjectPageVisitAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'page_number', 'user_agent_short')
    list_filter = ('timestamp', 'page_number')
    search_fields = ('ip_address',)

    @admin.display(description='UA')
    def user_agent_short(self, obj):
        return obj.user_agent[:60] if obj.user_agent else '—'
