from django.db import models


class PeoplePageVisit(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True, db_index=True)
    ip_address = models.GenericIPAddressField()
    user_agent = models.TextField(blank=True)
    referrer = models.TextField(blank=True)
    session_key = models.CharField(max_length=64, blank=True)

    class Meta:
        ordering = ['-timestamp']
        verbose_name = 'People Page Visit'

    def __str__(self):
        return f"{self.ip_address} @ {self.timestamp:%Y-%m-%d %H:%M}"


class GeneratedEmployee(models.Model):
    visit = models.ForeignKey(
        PeoplePageVisit,
        on_delete=models.CASCADE,
        related_name='employees',
    )
    first_name = models.CharField(max_length=64)
    last_name = models.CharField(max_length=64)
    email = models.EmailField(db_index=True)
    title = models.CharField(max_length=128)
    department = models.CharField(max_length=128)
    avatar_seed = models.CharField(max_length=32)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['last_name', 'first_name']
        verbose_name = 'Generated Employee'

    def __str__(self):
        return f"{self.first_name} {self.last_name} <{self.email}>"

    @property
    def full_name(self):
        return f"{self.first_name} {self.last_name}"

    @property
    def initials(self):
        return f"{self.first_name[0]}{self.last_name[0]}".upper()
