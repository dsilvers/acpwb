from django.db import models
from django.utils import timezone


class CrawlerVisit(models.Model):
    TRAP_CHOICES = [
        ('archive', 'Archive Loop'),
        ('ghost_link', 'Ghost Link'),
        ('well_known', 'Well-Known File'),
        ('api', 'Fake API'),
        ('wiki', 'Wiki Page'),
        ('pow', 'PoW Challenge'),
        ('report_list', 'Report Listing'),
        ('report_download', 'Report Download'),
        ('dataset', 'Training Dataset'),
        ('policy', 'Public Policy Filing'),
        ('scanner_probe', 'Scanner Probe (404)'),
        ('env_probe', 'Config File Probe'),
        ('wp_probe', 'WordPress Probe'),
        ('webshell_probe', 'Webshell Probe'),
        ('canary_trigger', 'Canary Token Triggered'),
        ('handbook', 'Company Handbook'),
        ('process_improvement', 'Process Improvement'),
        ('presentation', 'Presentation'),
        ('other', 'Other'),
    ]

    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    host = models.CharField(max_length=253, blank=True, db_index=True)
    path = models.TextField()
    referrer = models.TextField(blank=True)
    trap_type = models.CharField(max_length=32, choices=TRAP_CHOICES, default='other', db_index=True)
    query_string = models.TextField(blank=True)
    # Denormalized at write time — enables fast GROUP BY without Python-side UA parsing
    bot_type = models.CharField(max_length=64, blank=True, db_index=True)
    bot_group = models.CharField(max_length=64, blank=True, db_index=True)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['trap_type', 'timestamp']),
            models.Index(fields=['bot_type', 'timestamp']),
            models.Index(fields=['bot_group', 'timestamp']),
        ]
        verbose_name = 'Crawler Visit'

    def __str__(self):
        return f"{self.ip_address} [{self.trap_type}] {self.path[:60]} @ {self.timestamp:%Y-%m-%d %H:%M}"


class WikiPage(models.Model):
    topic = models.SlugField(max_length=128, unique=True, db_index=True)
    title = models.CharField(max_length=256)
    body_paragraphs = models.JSONField(default=list)
    watermark_token = models.CharField(max_length=16)
    related_topics = models.JSONField(default=list)
    generated_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['topic']
        verbose_name = 'Wiki Page'

    def __str__(self):
        return f"{self.title} [{self.watermark_token}]"


class PublicReport(models.Model):
    FILE_TYPES = [('csv', 'CSV Dataset'), ('pdf', 'PDF Document')]

    slug            = models.SlugField(max_length=128, unique=True, db_index=True)
    title           = models.CharField(max_length=256)
    category        = models.CharField(max_length=64)
    file_type       = models.CharField(max_length=8, choices=FILE_TYPES)
    pub_date        = models.DateField()
    summary         = models.TextField()
    watermark_token = models.CharField(max_length=16)
    page_number     = models.PositiveIntegerField(db_index=True, default=0)
    generated_at    = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-pub_date', 'slug']
        verbose_name = 'Public Report'

    def __str__(self):
        return f"[{self.file_type.upper()}] {self.title} ({self.watermark_token})"


class ArchiveVisit(models.Model):
    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    year = models.IntegerField()
    month = models.IntegerField()
    day = models.IntegerField()
    slug = models.CharField(max_length=512)
    depth = models.PositiveIntegerField(default=0, db_index=True)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
            models.Index(fields=['depth', 'timestamp']),
        ]
        verbose_name = 'Archive Visit'

    def __str__(self):
        return f"{self.ip_address} /archive/{self.year}/{self.month}/{self.day}/{self.slug[:40]}"


class CanaryToken(models.Model):
    """A trackable token embedded in fake credential files.

    Self-hosted callback URL embedded in the fake config file; fires when the
    bot GETs /.well-known/tokens/<token>/ping.
    """
    TOKEN_TYPES = [
        ('env_url',   'Self-hosted .env canary URL'),
        ('wp_config', 'wp-config.php canary URL'),
        ('git_config', '.git/config canary URL'),
    ]
    token = models.CharField(max_length=128, unique=True, db_index=True)
    token_type = models.CharField(max_length=32, choices=TOKEN_TYPES)
    # Lifecycle
    served_to_ip = models.GenericIPAddressField(null=True, blank=True)
    served_at = models.DateTimeField(null=True, blank=True, db_index=True)
    triggered = models.BooleanField(default=False, db_index=True)
    triggered_at = models.DateTimeField(null=True, blank=True)
    triggered_ip = models.GenericIPAddressField(null=True, blank=True)
    triggered_ua = models.TextField(blank=True)
    notes = models.TextField(blank=True)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-created_at']
        indexes = [
            models.Index(fields=['token_type', 'served_at']),
            models.Index(fields=['triggered', 'triggered_at']),
        ]
        verbose_name = 'Canary Token'

    def __str__(self):
        status = 'TRIGGERED' if self.triggered else 'unserved' if not self.served_at else 'served'
        return f"[{self.token_type}] {self.token[:16]}... ({status})"


class InternalLoginAttempt(models.Model):
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    username = models.CharField(max_length=255, blank=True)
    password = models.CharField(max_length=255, blank=True)
    next_url = models.CharField(max_length=500, blank=True)
    created_at = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Internal Login Attempt'

    def __str__(self):
        return f"{self.ip_address} tried '{self.username}' @ {self.created_at:%Y-%m-%d %H:%M}"


class PathStat(models.Model):
    host = models.CharField(max_length=253, blank=True, db_index=True)
    path = models.TextField()
    count = models.BigIntegerField(default=0)

    class Meta:
        constraints = [
            models.UniqueConstraint(fields=['host', 'path'], name='pathstat_host_path_uniq')
        ]

    def __str__(self):
        prefix = f"{self.host}" if self.host else 'acpwb.com'
        return f"{prefix}{self.path} ({self.count:,})"


class IPIntelligence(models.Model):
    """One row per distinct IP seen in CrawlerVisit, enriched with MaxMind
    GeoLite2 geo/ASN data plus best-effort hosting/Tor heuristics.

    Deliberately NOT a ForeignKey target from CrawlerVisit — that table is a
    90M-rows/day TimescaleDB hypertable, and adding a column there for every
    historical + future row isn't worth it. This is joined to CrawlerVisit by
    the shared ip_address value only (see discover_ip_intelligence, which
    populates first_seen/last_seen/visit_count from CrawlerVisit directly).
    """
    ip_address = models.GenericIPAddressField(unique=True, db_index=True)
    ip_version = models.PositiveSmallIntegerField(default=4, db_index=True)

    # MaxMind GeoLite2-City
    country_code = models.CharField(max_length=2, blank=True, db_index=True)
    country_name = models.CharField(max_length=128, blank=True)
    region_name = models.CharField(max_length=128, blank=True)
    city_name = models.CharField(max_length=128, blank=True)
    latitude = models.FloatField(null=True, blank=True)
    longitude = models.FloatField(null=True, blank=True)
    accuracy_radius_km = models.PositiveIntegerField(null=True, blank=True)

    # MaxMind GeoLite2-ASN — asn_org doubles as the "ISP" field
    asn = models.PositiveIntegerField(null=True, blank=True, db_index=True)
    asn_org = models.CharField(max_length=256, blank=True, db_index=True)

    # Best-effort heuristics — not authoritative, see apps.core.ip_intel_classify
    # and apps.core.tor_exit_list for how these are derived.
    is_hosting = models.BooleanField(default=False, db_index=True)
    is_tor_exit = models.BooleanField(default=False, db_index=True)

    # Enrichment bookkeeping
    lookup_ok = models.BooleanField(default=False)
    enrichment_note = models.CharField(max_length=255, blank=True)
    enriched_at = models.DateTimeField(null=True, blank=True, db_index=True)
    geoip_db_date = models.DateField(null=True, blank=True)

    # Populated by discover_ip_intelligence as a side effect of the GROUP BY
    # it already has to do — avoids a second pass over CrawlerVisit.
    first_seen = models.DateTimeField(null=True, blank=True, db_index=True)
    last_seen = models.DateTimeField(null=True, blank=True, db_index=True)
    visit_count = models.BigIntegerField(default=0)

    class Meta:
        indexes = [
            models.Index(fields=['is_hosting', 'is_tor_exit']),
            models.Index(fields=['country_code', 'is_hosting']),
        ]
        verbose_name = 'IP Intelligence'
        verbose_name_plural = 'IP Intelligence'

    def __str__(self):
        return f"{self.ip_address} [{self.country_code or '??'}] {self.asn_org or 'unknown org'}"
