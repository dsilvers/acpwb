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
        ('other', 'Other'),
    ]

    timestamp = models.DateTimeField(auto_now_add=True, db_index=True)
    ip_address = models.GenericIPAddressField()
    user_agent = models.TextField(blank=True)
    path = models.TextField()
    referrer = models.TextField(blank=True)
    trap_type = models.CharField(max_length=32, choices=TRAP_CHOICES, default='other')
    query_string = models.TextField(blank=True)

    class Meta:
        ordering = ['-timestamp']
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
    timestamp = models.DateTimeField(auto_now_add=True)
    ip_address = models.GenericIPAddressField()
    user_agent = models.TextField(blank=True)
    year = models.IntegerField()
    month = models.IntegerField()
    day = models.IntegerField()
    slug = models.CharField(max_length=512)
    depth = models.PositiveIntegerField(default=0)

    class Meta:
        ordering = ['-timestamp']
        verbose_name = 'Archive Visit'

    def __str__(self):
        return f"{self.ip_address} /archive/{self.year}/{self.month}/{self.day}/{self.slug[:40]}"
