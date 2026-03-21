from django.db import models


class DataOptOutRequest(models.Model):
    REQUEST_CHOICES = [
        ('do_not_sell', 'Do Not Sell My Personal Information'),
        ('delete',      'Delete My Personal Information'),
        ('access',      'Access My Personal Information'),
    ]

    name         = models.CharField(max_length=256)
    email        = models.EmailField(db_index=True)
    request_type = models.CharField(max_length=32, choices=REQUEST_CHOICES, default='do_not_sell')
    state        = models.CharField(max_length=64, blank=True)
    message      = models.TextField(blank=True)
    ip_address   = models.GenericIPAddressField()
    processed    = models.BooleanField(default=False)
    created_at   = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Data Opt-Out Request'
        verbose_name_plural = 'Data Opt-Out Requests'

    def __str__(self):
        return f"{self.name} <{self.email}> [{self.get_request_type_display()}] @ {self.created_at:%Y-%m-%d}"


class Fortune500Company(models.Model):
    rank = models.PositiveIntegerField(unique=True)
    name = models.CharField(max_length=256)
    website = models.URLField()
    industry = models.CharField(max_length=128, blank=True)

    class Meta:
        ordering = ['rank']
        verbose_name = 'Fortune 500 Company'
        verbose_name_plural = 'Fortune 500 Companies'

    def __str__(self):
        return f"#{self.rank} {self.name}"

    @property
    def initials(self):
        words = [w for w in self.name.split() if w not in ('&', 'and', 'the', 'of', 'for')]
        if len(words) >= 2:
            return f"{words[0][0]}{words[1][0]}".upper()
        return self.name[:2].upper()
