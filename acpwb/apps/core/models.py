from django.db import models


class DashboardStat(models.Model):
    key = models.CharField(max_length=128, unique=True, db_index=True)
    value = models.JSONField(default=dict)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        verbose_name = 'Dashboard Stat'

    def __str__(self):
        return self.key
