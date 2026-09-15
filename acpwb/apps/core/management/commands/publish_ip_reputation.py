"""
Sync IPReputationScore -> PublishedIPReputation — the last, deliberately
separate step of the traffic-intelligence pipeline (Phase 5 of the plan
this implements).

Kept as its own command, on its own cadence, rather than folded into
score_ip_reputation: PublishedIPReputation is the external-facing contract
(what a future blacklist/API customer reads), and the plan is explicit that
committing to that contract's cadence should stay decoupled from how often
the internal scoring itself iterates.

Watermarked by IPReputationScore.id, same resumable pattern as
score_ip_reputation/discover_ip_intelligence.

Run via cron once there's confidence in score_ip_reputation's output:
    python manage.py publish_ip_reputation
"""
import fcntl
import time

from django.core.management.base import BaseCommand

from apps.core.models import DashboardStat
from apps.honeypot.models import IPReputationScore, PublishedIPReputation

_LOCK_FILE = '/tmp/acpwb-publish-ip-reputation.lock'
_WATERMARK_KEY = 'ip_reputation_publish_watermark'


class Command(BaseCommand):
    help = 'Sync IPReputationScore into the external-facing PublishedIPReputation surface.'

    def add_arguments(self, parser):
        parser.add_argument('--batch', type=int, default=5000)
        parser.add_argument('--max-batches', type=int, default=0)
        parser.add_argument('--max-seconds', type=int, default=50)

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another publish run is already in progress — exiting.')
                return
            self._run(options)

    def _get_watermark(self):
        stat, _ = DashboardStat.objects.get_or_create(key=_WATERMARK_KEY, defaults={'value': {'id': 0}})
        return stat.value.get('id', 0)

    def _set_watermark(self, last_id):
        DashboardStat.objects.update_or_create(key=_WATERMARK_KEY, defaults={'value': {'id': last_id}})

    def _run(self, options):
        batch_size = options['batch']
        max_batches = options['max_batches']
        max_seconds = options['max_seconds']
        started = time.monotonic()

        watermark = self._get_watermark()
        total_published = 0
        batches_run = 0

        while True:
            if max_seconds and time.monotonic() - started >= max_seconds:
                self.stdout.write(f'Time budget reached after {batches_run} batch(es).')
                break

            rows = list(
                IPReputationScore.objects.filter(id__gt=watermark).order_by('id')[:batch_size]
            )
            if not rows:
                break

            for row in rows:
                PublishedIPReputation.objects.update_or_create(
                    ip_address=row.ip_address,
                    defaults={
                        'classification': row.classification,
                        'confidence': row.residential_proxy_score,
                    },
                )
            total_published += len(rows)
            watermark = rows[-1].id
            self._set_watermark(watermark)

            batches_run += 1
            if max_batches and batches_run >= max_batches:
                break

        self.stdout.write(self.style.SUCCESS(
            f'Published {total_published} IP(s) across {batches_run} batch(es); watermark now {watermark}.'
        ))
