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


class VoicemailRecording(models.Model):
    TRANSCRIPTION_STATUS_CHOICES = [
        ('pending',   'Pending'),
        ('completed', 'Completed'),
        ('failed',    'Failed'),
    ]

    call_sid             = models.CharField(max_length=64, unique=True, db_index=True)
    recording_sid        = models.CharField(max_length=64, unique=True, db_index=True)
    recording_url        = models.URLField(max_length=512)
    recording_duration   = models.IntegerField(default=0, help_text="Duration in seconds")
    caller_number        = models.CharField(max_length=32, blank=True)
    transcription_status = models.CharField(
        max_length=16,
        choices=TRANSCRIPTION_STATUS_CHOICES,
        default='pending',
        db_index=True,
    )
    transcription_text   = models.TextField(blank=True, null=True)
    raw_payload          = models.JSONField(default=dict)
    received_at          = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-received_at']
        verbose_name = 'Voicemail Recording'
        verbose_name_plural = 'Voicemail Recordings'

    def __str__(self):
        return f"Voicemail from {self.caller_number} @ {self.received_at:%Y-%m-%d %H:%M} ({self.recording_sid})"

    @property
    def duration_display(self):
        m, s = divmod(self.recording_duration, 60)
        return f"{m}:{s:02d}"


class CallLog(models.Model):
    CALL_STATUS_CHOICES = [
        ('completed',  'Completed'),
        ('busy',       'Busy'),
        ('no-answer',  'No Answer'),
        ('failed',     'Failed'),
        ('canceled',   'Canceled'),
    ]

    call_sid      = models.CharField(max_length=64, unique=True, db_index=True)
    caller_number = models.CharField(max_length=32, blank=True)
    call_status   = models.CharField(max_length=16, choices=CALL_STATUS_CHOICES, db_index=True)
    call_duration = models.IntegerField(default=0, help_text="Duration in seconds (0 for non-completed calls)")
    raw_payload   = models.JSONField(default=dict)
    received_at   = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-received_at']
        verbose_name = 'Call Log'
        verbose_name_plural = 'Call Logs'

    def __str__(self):
        return f"Call from {self.caller_number} [{self.call_status}] @ {self.received_at:%Y-%m-%d %H:%M}"

    @property
    def duration_display(self):
        if not self.call_duration:
            return '—'
        m, s = divmod(self.call_duration, 60)
        return f"{m}:{s:02d}"

    @property
    def left_voicemail(self):
        return VoicemailRecording.objects.filter(call_sid=self.call_sid).exists()
