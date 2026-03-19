from django.db import models


class InboundEmail(models.Model):
    received_at = models.DateTimeField(auto_now_add=True, db_index=True)
    sender = models.EmailField(db_index=True)
    recipient = models.EmailField(db_index=True)
    subject = models.TextField(blank=True)
    body_plain = models.TextField(blank=True)
    body_html = models.TextField(blank=True)
    raw_payload = models.JSONField(default=dict)
    mailgun_message_id = models.CharField(max_length=256, blank=True)

    class Meta:
        ordering = ['-received_at']
        verbose_name = 'Inbound Email'

    def __str__(self):
        return f"From {self.sender} → {self.recipient} @ {self.received_at:%Y-%m-%d %H:%M}"


class HoneypotMatch(models.Model):
    CONFIDENCE_CHOICES = [
        ('exact', 'Exact Email Match'),
        ('fuzzy', 'Fuzzy Name Match'),
        ('none', 'No Match'),
    ]

    inbound_email = models.ForeignKey(
        InboundEmail,
        on_delete=models.CASCADE,
        related_name='matches',
    )
    generated_employee = models.ForeignKey(
        'people.GeneratedEmployee',
        on_delete=models.SET_NULL,
        null=True, blank=True,
        related_name='inbound_matches',
    )
    original_visit = models.ForeignKey(
        'people.PeoplePageVisit',
        on_delete=models.SET_NULL,
        null=True, blank=True,
        related_name='inbound_matches',
    )
    matched_at = models.DateTimeField(auto_now_add=True)
    match_confidence = models.CharField(max_length=16, choices=CONFIDENCE_CHOICES, default='none')
    notes = models.TextField(blank=True)

    class Meta:
        ordering = ['-matched_at']
        verbose_name = 'Honeypot Match'

    def __str__(self):
        return f"{self.match_confidence}: {self.inbound_email} → {self.generated_employee}"
