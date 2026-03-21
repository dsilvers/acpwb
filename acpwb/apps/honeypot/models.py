from django.db import models


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
        ('other', 'Other'),
    ]

    timestamp = models.DateTimeField(auto_now_add=True, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
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
    timestamp = models.DateTimeField(auto_now_add=True, db_index=True)
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
