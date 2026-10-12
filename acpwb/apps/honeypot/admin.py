from django.contrib import admin
from .models import (
    SourceCodeEmail,
    CrawlerVisit, WikiPage, ArchiveVisit, PublicReport, InternalLoginAttempt, IPIntelligence,
    RequestFingerprint, IPReputationScore, PublishedIPReputation, EdgeConnection,
)


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


@admin.register(IPIntelligence)
class IPIntelligenceAdmin(admin.ModelAdmin):
    list_display = ('ip_address', 'country_code', 'asn_org', 'is_hosting', 'is_tor_exit', 'visit_count', 'enriched_at')
    list_filter = ('is_hosting', 'is_tor_exit', 'lookup_ok', 'country_code')
    search_fields = ('ip_address', 'asn_org', 'country_name', 'city_name')
    readonly_fields = ('first_seen', 'last_seen', 'enriched_at')
    ordering = ('-visit_count',)

    @admin.display(description='UA')
    def user_agent_short(self, obj):
        return obj.user_agent[:60] if obj.user_agent else '—'


@admin.register(RequestFingerprint)
class RequestFingerprintAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'client_protocol', 'tls_protocol', 'ja4', 'first_seen_ip', 'referrer_present')
    list_filter = ('client_protocol', 'tls_protocol', 'first_seen_ip', 'referrer_present')
    search_fields = ('ip_address', 'user_agent', 'browser_headers_present', 'ja4', 'ja3', 'ja4h')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)


@admin.register(EdgeConnection)
class EdgeConnectionAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'outcome', 'alpn', 'tls_version', 'ja4', 'request_count', 'first_path')
    list_filter = ('outcome', 'alpn', 'tls_version')
    search_fields = ('ip_address', 'ja4', 'ja3', 'sni', 'first_user_agent')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)
    show_full_result_count = False


@admin.register(IPReputationScore)
class IPReputationScoreAdmin(admin.ModelAdmin):
    list_display = ('ip_address', 'classification', 'residential_proxy_score', 'score_version', 'last_scored_at')
    list_filter = ('classification', 'score_version')
    search_fields = ('ip_address',)
    readonly_fields = ('first_flagged_at', 'last_scored_at')
    ordering = ('-residential_proxy_score',)


@admin.register(PublishedIPReputation)
class PublishedIPReputationAdmin(admin.ModelAdmin):
    list_display = ('ip_address', 'classification', 'confidence', 'published_at')
    list_filter = ('classification',)
    search_fields = ('ip_address',)
    readonly_fields = ('published_at',)
    ordering = ('-published_at',)


@admin.register(SourceCodeEmail)
class SourceCodeEmailAdmin(admin.ModelAdmin):
    list_display = ('address', 'host', 'path', 'first_seen', 'last_seen', 'times_served', 'first_ip')
    search_fields = ('address', 'path', 'first_ip')
    list_filter = ('first_seen',)
    readonly_fields = ('first_seen', 'last_seen', 'times_served')
