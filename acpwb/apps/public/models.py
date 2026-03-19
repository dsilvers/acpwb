from django.db import models


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
