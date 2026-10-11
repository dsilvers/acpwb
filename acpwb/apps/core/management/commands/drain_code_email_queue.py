"""
Drain the Redis code-email queue (programmer addresses shown on legacy
source-code pages, pushed by acpwb_go) into honeypot.SourceCodeEmail.

Each item is one (address, page) pair; repeats update last_seen and
times_served instead of inserting duplicates. Low volume — acpwb_go dedupes
each pair for 30 days before re-queuing it.

Run via cron every minute, same as drain_fingerprint_queue.
"""
import fcntl
import time
from datetime import datetime, timezone as dt_timezone

from django.core.management.base import BaseCommand
from django.db.models import F
from django.utils import timezone
from django.utils.dateparse import parse_datetime

from apps.core.crawler_queue import (
    code_email_queue_length, finalize_batch, pop_code_emails, recover_code_emails,
)
from apps.honeypot.models import SourceCodeEmail

_LOCK_FILE = '/tmp/acpwb-code-email-drain.lock'


def upsert_code_email(item):
    """Insert or refresh one (address, path) registry row. Returns True if a
    new row was created."""
    address = (item.get('address') or '').strip().lower()
    path = (item.get('path') or '').strip()[:512]
    if not address or not path:
        return False
    ts = parse_datetime(item.get('timestamp') or '') or timezone.now()
    ip = item.get('ip_address') or None
    obj, created = SourceCodeEmail.objects.get_or_create(
        address=address, path=path,
        defaults={'host': (item.get('host') or '')[:253], 'first_seen': ts, 'last_seen': ts, 'first_ip': ip},
    )
    if not created:
        SourceCodeEmail.objects.filter(pk=obj.pk).update(last_seen=ts, times_served=F('times_served') + 1)
    return created


class Command(BaseCommand):
    help = 'Drain the Redis code-email queue into honeypot.SourceCodeEmail'

    def add_arguments(self, parser):
        parser.add_argument('--batch', type=int, default=500)
        parser.add_argument('--max-seconds', type=int, default=55)

    def _log(self, msg):
        self.stdout.write(f'[{datetime.now(dt_timezone.utc).isoformat()}] {msg}')

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self._log('Another drain is already running — exiting.')
                return
            self._drain(options)

    def _process(self, items, batch_key):
        created = 0
        for item in items:
            try:
                created += int(upsert_code_email(item))
            except Exception:
                pass  # skip malformed entries
        finalize_batch(batch_key)
        return created

    def _drain(self, options):
        started = time.monotonic()
        created = 0
        for items, batch_key in recover_code_emails():
            created += self._process(items, batch_key)
        while not options['max_seconds'] or time.monotonic() - started < options['max_seconds']:
            items, batch_key = pop_code_emails(options['batch'])
            if not batch_key:
                break
            created += self._process(items, batch_key)
        remaining = code_email_queue_length()
        self._log(f'Created {created} new address/page pair(s). Queue depth: '
                  f'{"unknown (Redis unavailable)" if remaining < 0 else remaining}.')
