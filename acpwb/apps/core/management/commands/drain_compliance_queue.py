"""
Drain the Redis compliance queue (/compliance/ milestones pushed by
acpwb_go: finished attempts, certificates served, learner details typed in)
into honeypot.ComplianceCompletion.

Low volume — acpwb_go queues each (ip, course, attempt, event) at most once
per 30 days. Per-question answers are not queued: the CrawlerVisit path holds
them (decode with `acpwb_go compliance-score`).

Run via cron every minute, same as drain_code_email_queue.
"""
import fcntl
import time
import uuid
from datetime import datetime, timezone as dt_timezone

from django.core.management.base import BaseCommand
from django.utils import timezone
from django.utils.dateparse import parse_datetime

from apps.core.crawler_queue import (
    compliance_queue_length, finalize_batch, pop_compliance_completions, recover_compliance_completions,
)
from apps.honeypot.models import ComplianceCompletion

_LOCK_FILE = '/tmp/acpwb-compliance-drain.lock'
_EVENTS = {k for k, _ in ComplianceCompletion.EVENT_CHOICES}


def _int(v, lo=0, hi=32767):
    try:
        return max(lo, min(hi, int(v)))
    except (TypeError, ValueError):
        return lo


def build_completion(item):
    """One queued payload → an unsaved ComplianceCompletion, or None if it's
    malformed."""
    event = item.get('event')
    ip = item.get('ip_address')
    if event not in _EVENTS or not ip:
        return None
    request_id = item.get('request_id')
    try:
        request_id = uuid.UUID(request_id) if request_id else None
    except (TypeError, ValueError):
        request_id = None
    return ComplianceCompletion(
        timestamp=parse_datetime(item.get('timestamp') or '') or timezone.now(),
        event=event,
        ip_address=ip,
        user_agent=item.get('user_agent') or '',
        host=(item.get('host') or '')[:253],
        request_id=request_id,
        course=(item.get('course') or '')[:80],
        attempt=_int(item.get('attempt'), hi=2**31 - 1),
        answers=(item.get('answers') or '')[:32],
        score=_int(item.get('score'), hi=100),
        passed=bool(item.get('passed')),
        cert_id=(item.get('cert_id') or '')[:8],
        learner_name=(item.get('learner_name') or '')[:120],
        learner_email=(item.get('learner_email') or '')[:254],
        learner_employee_id=(item.get('learner_employee_id') or '')[:64],
        bot_type=(item.get('bot_type') or '')[:64],
        client_class=(item.get('client_class') or '')[:16],
    )


class Command(BaseCommand):
    help = 'Drain the Redis compliance queue into honeypot.ComplianceCompletion'

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
        rows = []
        for item in items:
            try:
                row = build_completion(item)
            except Exception:
                row = None  # skip malformed entries
            if row is not None:
                rows.append(row)
        ComplianceCompletion.objects.bulk_create(rows)
        finalize_batch(batch_key)
        return len(rows)

    def _drain(self, options):
        started = time.monotonic()
        created = 0
        for items, batch_key in recover_compliance_completions():
            created += self._process(items, batch_key)
        while not options['max_seconds'] or time.monotonic() - started < options['max_seconds']:
            items, batch_key = pop_compliance_completions(options['batch'])
            if not batch_key:
                break
            created += self._process(items, batch_key)
        remaining = compliance_queue_length()
        self._log(f'Inserted {created} compliance milestone(s). Queue depth: '
                  f'{"unknown (Redis unavailable)" if remaining < 0 else remaining}.')
