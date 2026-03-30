"""
Incrementally update dashboard stats in the database using PK high-water marks.

Each model's processed rows are tracked by a high-water mark stat row
(e.g. hwm.crawler_visit = 12345).  On each run only rows with id > hwm are
processed; cumulative dict/int stats are updated in place.

Daily charts are always fully recomputed (last 60/30 days) since they are
windowed time queries that cannot be made truly incremental.

Run on a schedule (every 30 minutes recommended).

Usage:
    python manage.py precalc_dashboard

Cron (host crontab):
    */30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web \\
        python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
"""
import time
from datetime import timedelta

from django.core.management.base import BaseCommand
from django.db.models import Count, Max
from django.db.models.functions import TruncDate
from django.utils import timezone

from apps.core.models import DashboardStat
from apps.honeypot.models import ArchiveVisit, CanaryToken, CrawlerVisit, InternalLoginAttempt
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.public.models import DataOptOutRequest
from apps.webhooks.models import InboundEmail

_MAX_DICT_ENTRIES = 200


class Command(BaseCommand):
    help = 'Incrementally update dashboard stats in the database.'

    def handle(self, *args, **options):
        run_start = time.monotonic()
        now_str = timezone.now().strftime('%Y-%m-%d %H:%M:%S %Z')
        self.stdout.write(f'precalc_dashboard starting at {now_str}')

        self._update_crawlers()
        self._update_archive()
        self._update_emails()
        self._update_people()
        self._update_projects()
        self._update_login_attempts()
        self._update_opt_outs()
        self._update_canary()

        elapsed = time.monotonic() - run_start
        self.stdout.write(f'Done in {elapsed:.1f}s')

    # ── Helpers ───────────────────────────────────────────────────────────────

    def _upsert(self, key, default_value):
        stat, _ = DashboardStat.objects.get_or_create(
            key=key, defaults={'value': default_value}
        )
        return stat

    def _prune(self, d):
        if len(d) > _MAX_DICT_ENTRIES:
            return dict(sorted(d.items(), key=lambda x: x[1], reverse=True)[:_MAX_DICT_ENTRIES])
        return d

    def _inc_dict(self, stat, rows_qs, field):
        """Add counts from rows_qs.values(field) into stat.value dict."""
        for row in rows_qs.values(field).annotate(c=Count('id')):
            k = str(row[field] or '')
            stat.value[k] = stat.value.get(k, 0) + row['c']

    # ── CrawlerVisit ──────────────────────────────────────────────────────────

    def _update_crawlers(self):
        hwm = self._upsert('hwm.crawler_visit', 0)
        new_max = CrawlerVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if new_max:
            new_rows = CrawlerVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('crawlers.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            for key, field in [
                ('crawlers.by_trap_type', 'trap_type'),
                ('crawlers.by_bot_type',  'bot_type'),
                ('crawlers.by_bot_group', 'bot_group'),
            ]:
                stat = self._upsert(key, {})
                self._inc_dict(stat, new_rows, field)
                stat.save()

            for key, field in [
                ('crawlers.by_ip',   'ip_address'),
                ('crawlers.by_path', 'path'),
            ]:
                stat = self._upsert(key, {})
                self._inc_dict(stat, new_rows, field)
                stat.value = self._prune(stat.value)
                stat.save()

            stat = self._upsert('crawlers.by_host', {})
            self._inc_dict(stat, new_rows.exclude(host=''), 'host')
            stat.value = self._prune(stat.value)
            stat.save()

            probe_types = ['env_probe', 'wp_probe', 'webshell_probe', 'scanner_probe']
            stat = self._upsert('crawlers.probe_by_path', {})
            self._inc_dict(stat, new_rows.filter(trap_type__in=probe_types), 'path')
            stat.value = self._prune(stat.value)
            stat.save()

            stat = self._upsert('crawlers.webshell_cmds', {})
            self._inc_dict(stat, new_rows.filter(trap_type='webshell_probe').exclude(query_string=''), 'query_string')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()
            self.stdout.write(f'  crawlers: updated to hwm={new_max}')
        else:
            self.stdout.write('  crawlers: no new rows')

        # Daily chart always recomputed
        now = timezone.now()
        rows = (CrawlerVisit.objects
                .filter(timestamp__gte=now - timedelta(days=60))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('crawlers.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

    # ── ArchiveVisit ──────────────────────────────────────────────────────────

    def _update_archive(self):
        hwm = self._upsert('hwm.archive_visit', 0)
        new_max = ArchiveVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if new_max:
            new_rows = ArchiveVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('archive.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            stat = self._upsert('archive.by_depth', {})
            self._inc_dict(stat, new_rows, 'depth')
            stat.save()

            stat = self._upsert('archive.by_ip', {})
            self._inc_dict(stat, new_rows, 'ip_address')
            stat.value = self._prune(stat.value)
            stat.save()

            stat = self._upsert('archive.by_slug', {})
            self._inc_dict(stat, new_rows, 'slug')
            stat.value = self._prune(stat.value)
            stat.save()

            # Bot type via UA classification
            from apps.core.bot_classify import classify_ua
            stat = self._upsert('archive.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            hwm.value = new_max
            hwm.save()
            self.stdout.write(f'  archive: updated to hwm={new_max}')
        else:
            self.stdout.write('  archive: no new rows')

        now = timezone.now()
        rows = (ArchiveVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('archive.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

    # ── InboundEmail ──────────────────────────────────────────────────────────

    def _update_emails(self):
        hwm = self._upsert('hwm.inbound_email', 0)
        new_max = InboundEmail.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if new_max:
            new_rows = InboundEmail.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('emails.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            stat = self._upsert('emails.by_domain', {})
            for sender in new_rows.values_list('sender', flat=True):
                domain = sender.split('@')[-1].lower() if '@' in sender else sender
                stat.value[domain] = stat.value.get(domain, 0) + 1
            stat.value = self._prune(stat.value)
            stat.save()

            stat = self._upsert('emails.by_recipient', {})
            self._inc_dict(stat, new_rows, 'recipient')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()
            self.stdout.write(f'  emails: updated to hwm={new_max}')
        else:
            self.stdout.write('  emails: no new rows')

        now = timezone.now()
        rows = (InboundEmail.objects
                .filter(received_at__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('received_at'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('emails.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

    # ── PeoplePageVisit ───────────────────────────────────────────────────────

    def _update_people(self):
        hwm = self._upsert('hwm.people_visit', 0)
        new_max = PeoplePageVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if new_max:
            new_rows = PeoplePageVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('people.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            from apps.core.bot_classify import classify_ua
            stat = self._upsert('people.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            stat = self._upsert('people.by_ip', {})
            self._inc_dict(stat, new_rows, 'ip_address')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()
            self.stdout.write(f'  people: updated to hwm={new_max}')
        else:
            self.stdout.write('  people: no new rows')

        now = timezone.now()
        rows = (PeoplePageVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('people.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

    # ── ProjectPageVisit ──────────────────────────────────────────────────────

    def _update_projects(self):
        hwm = self._upsert('hwm.project_visit', 0)
        new_max = ProjectPageVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if new_max:
            new_rows = ProjectPageVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('projects.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            from apps.core.bot_classify import classify_ua
            stat = self._upsert('projects.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            stat = self._upsert('projects.by_ip', {})
            self._inc_dict(stat, new_rows, 'ip_address')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()
            self.stdout.write(f'  projects: updated to hwm={new_max}')
        else:
            self.stdout.write('  projects: no new rows')

        now = timezone.now()
        rows = (ProjectPageVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('projects.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

    # ── InternalLoginAttempt ──────────────────────────────────────────────────

    def _update_login_attempts(self):
        hwm = self._upsert('hwm.login_attempt', 0)
        new_max = InternalLoginAttempt.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']
        if not new_max:
            self.stdout.write('  logins: no new rows')
            return

        total_stat = self._upsert('login_attempts.total', 0)
        total_stat.value = total_stat.value + InternalLoginAttempt.objects.filter(id__gt=hwm.value, id__lte=new_max).count()
        total_stat.save()

        hwm.value = new_max
        hwm.save()
        self.stdout.write(f'  logins: updated to hwm={new_max}')

    # ── DataOptOutRequest ─────────────────────────────────────────────────────

    def _update_opt_outs(self):
        hwm = self._upsert('hwm.opt_out', 0)
        new_max = DataOptOutRequest.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']
        if not new_max:
            self.stdout.write('  optouts: no new rows')
            return

        total_stat = self._upsert('optouts.total', 0)
        total_stat.value = total_stat.value + DataOptOutRequest.objects.filter(id__gt=hwm.value, id__lte=new_max).count()
        total_stat.save()

        hwm.value = new_max
        hwm.save()
        self.stdout.write(f'  optouts: updated to hwm={new_max}')

    # ── CanaryToken (always full recompute — cheap) ───────────────────────────

    def _update_canary(self):
        triggered = CanaryToken.objects.filter(triggered=True)
        count = triggered.count()

        stat = self._upsert('canary.triggered_count', 0)
        stat.value = count
        stat.save()

        self.stdout.write(f'  canary: {count} triggered')
