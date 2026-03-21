from django.contrib import admin
from .models import CrawlerVisit, WikiPage, ArchiveVisit, PublicReport, InternalLoginAttempt


@admin.register(CrawlerVisit)
class CrawlerVisitAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'trap_type', 'path_short', 'user_agent_short')
    list_filter = ('trap_type', 'timestamp')
    search_fields = ('ip_address', 'path', 'user_agent')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)

    @admin.display(description='Path')
    def path_short(self, obj):
        return obj.path[:60]

    @admin.display(description='UA')
    def user_agent_short(self, obj):
        return obj.user_agent[:60] if obj.user_agent else '—'


@admin.register(WikiPage)
class WikiPageAdmin(admin.ModelAdmin):
    list_display = ('topic', 'title', 'watermark_token', 'generated_at')
    search_fields = ('topic', 'title', 'watermark_token')
    readonly_fields = ('generated_at', 'watermark_token')
    ordering = ('topic',)


@admin.register(ArchiveVisit)
class ArchiveVisitAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'year', 'month', 'day', 'depth', 'slug_short')
    list_filter = ('year', 'month')
    search_fields = ('ip_address', 'slug')
    readonly_fields = ('timestamp',)

    @admin.display(description='Slug')
    def slug_short(self, obj):
        return obj.slug[:50]


@admin.register(PublicReport)
class PublicReportAdmin(admin.ModelAdmin):
    list_display = ('title', 'category', 'file_type', 'pub_date', 'watermark_token', 'generated_at')
    list_filter = ('file_type', 'category')
    search_fields = ('title', 'slug', 'watermark_token')
    readonly_fields = ('generated_at', 'watermark_token', 'slug')
    ordering = ('-pub_date',)


@admin.register(InternalLoginAttempt)
class InternalLoginAttemptAdmin(admin.ModelAdmin):
    list_display = ('created_at', 'ip_address', 'username', 'password_short', 'next_url', 'user_agent_short')
    search_fields = ('ip_address', 'username', 'password', 'next_url')
    readonly_fields = ('created_at',)
    ordering = ('-created_at',)

    @admin.display(description='Password')
    def password_short(self, obj):
        return obj.password[:40] if obj.password else '—'

    @admin.display(description='UA')
    def user_agent_short(self, obj):
        return obj.user_agent[:60] if obj.user_agent else '—'
