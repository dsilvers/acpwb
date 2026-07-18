from django.db import models


class TrafficMinuteStat(models.Model):
    """Pre-aggregated per-minute request counts from CrawlerVisit, by bot_type."""
    minute = models.DateTimeField(db_index=True)
    bot_type = models.CharField(max_length=64, blank=True, db_index=True)
    count = models.PositiveIntegerField(default=0)

    class Meta:
        unique_together = [('minute', 'bot_type')]
        indexes = [
            models.Index(fields=['minute', 'bot_type']),
        ]

    def __str__(self):
        return f'{self.minute.isoformat()} | {self.bot_type} | {self.count}'


class DashboardStat(models.Model):
    key = models.CharField(max_length=128, unique=True, db_index=True)
    value = models.JSONField(default=dict)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        verbose_name = 'Dashboard Stat'

    def __str__(self):
        return self.key
