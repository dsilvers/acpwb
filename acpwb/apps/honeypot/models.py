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
