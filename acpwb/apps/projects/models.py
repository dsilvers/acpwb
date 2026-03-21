from django.db import models


class ProjectStory(models.Model):
    slug = models.SlugField(max_length=128, unique=True, db_index=True)
    title = models.CharField(max_length=256)
    summary = models.TextField()
    body_paragraphs = models.JSONField(default=list)
    impact_metric = models.CharField(max_length=128)
    industry_tag = models.CharField(max_length=64)
    page_number = models.PositiveIntegerField(db_index=True, default=0)
    generated_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['page_number', 'slug']
        verbose_name = 'Project Story'
        verbose_name_plural = 'Project Stories'

    def __str__(self):
        return self.title


class ProjectPageVisit(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    page_number = models.PositiveIntegerField(default=1)
    pow_token = models.CharField(max_length=128, blank=True)

    class Meta:
        ordering = ['-timestamp']
        verbose_name = 'Project Page Visit'

    def __str__(self):
        return f"{self.ip_address} page {self.page_number} @ {self.timestamp:%Y-%m-%d %H:%M}"
