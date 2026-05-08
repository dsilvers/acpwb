import uuid

from django.db import models
from django.utils import timezone


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


class JobApplication(models.Model):
    job_id    = models.PositiveIntegerField(db_index=True)
    job_title = models.CharField(max_length=256)
    name      = models.CharField(max_length=256)
    email     = models.EmailField(db_index=True)
    phone     = models.CharField(max_length=32, blank=True)
    cover_letter         = models.TextField(blank=True)
    resume_filename      = models.CharField(max_length=256, blank=True)
    resume_data          = models.BinaryField(blank=True)
    resume_content_type  = models.CharField(max_length=128, blank=True)
    ip_address  = models.GenericIPAddressField()
    user_agent  = models.TextField(blank=True)
    created_at  = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Job Application'
        verbose_name_plural = 'Job Applications'

    def __str__(self):
        return f"{self.name} → {self.job_title} @ {self.created_at:%Y-%m-%d}"


class JobApplicationDocument(models.Model):
    application  = models.ForeignKey(JobApplication, on_delete=models.CASCADE, related_name='documents')
    filename     = models.CharField(max_length=256)
    data         = models.BinaryField()
    content_type = models.CharField(max_length=128, blank=True)
    created_at   = models.DateTimeField(auto_now_add=True)

    class Meta:
        verbose_name = 'Job Application Document'
        verbose_name_plural = 'Job Application Documents'

    def __str__(self):
        return f"{self.filename} ({self.application})"


class ConferenceRegistration(models.Model):
    REGISTRATION_TYPES = [
        ('full', 'Full Conference'),
        ('day1', 'Day 1 Only'),
        ('day2', 'Day 2 Only'),
    ]

    year              = models.IntegerField(default=2026, db_index=True)
    # Contact
    first_name        = models.CharField(max_length=100)
    last_name         = models.CharField(max_length=100)
    email             = models.EmailField(db_index=True)
    phone             = models.CharField(max_length=30, blank=True)
    # Address
    address_line1     = models.CharField(max_length=200)
    address_line2     = models.CharField(max_length=200, blank=True)
    city              = models.CharField(max_length=100)
    state             = models.CharField(max_length=100)
    postal_code       = models.CharField(max_length=20)
    country           = models.CharField(max_length=100, default='United States')
    # Badge
    badge_name        = models.CharField(max_length=100)
    badge_title       = models.CharField(max_length=150, blank=True)
    badge_company     = models.CharField(max_length=150, blank=True)
    # Professional details
    job_title         = models.CharField(max_length=150, blank=True)
    organization      = models.CharField(max_length=150, blank=True)
    years_in_field    = models.CharField(max_length=20, blank=True)
    professional_membership = models.CharField(max_length=200, blank=True)
    registration_type = models.CharField(max_length=10, choices=REGISTRATION_TYPES, default='full')
    # Metadata — CC data is NEVER stored
    token             = models.UUIDField(default=uuid.uuid4, editable=False, unique=True, db_index=True)
    ip_address        = models.GenericIPAddressField(null=True, blank=True)
    user_agent        = models.CharField(max_length=512, blank=True)
    created_at        = models.DateTimeField(default=timezone.now, editable=False)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Conference Registration'
        verbose_name_plural = 'Conference Registrations'

    def __str__(self):
        return f"{self.first_name} {self.last_name} — PERCH {self.year} ({self.get_registration_type_display()})"

    @property
    def full_name(self):
        return f"{self.first_name} {self.last_name}"
