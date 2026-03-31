"""
Incrementally update dashboard stats in the database using PK high-water marks.

Each model's processed rows are tracked by a high-water mark stat row
(e.g. hwm.crawler_visit = 12345).  On each run only rows with id > hwm are
processed; cumulative dict/int stats are updated in place.

Daily charts are recomputed only when new rows exist (skipped on quiet ticks
to avoid unnecessary full-table GROUP BY scans).

Run on a schedule (every 30 minutes recommended).

Usage:
    python manage.py precalc_dashboard

Cron (host crontab):
    */30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web \\
        python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
"""
import fcntl
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
_MAX_ROWS_PER_RUN = 500_000


class Command(BaseCommand):
    help = 'Incrementally update dashboard stats in the database.'

    def handle(self, *args, **options):
        with open('/tmp/precalc_dashboard.lock', 'w') as lockfile:
            try:
                fcntl.flock(lockfile, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                self.stdout.write('precalc_dashboard already running — skipping.')
                return

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
            self._update_graphs()

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

    def _cap_new_max(self, hwm_value, actual_max):
        """Cap new_max to avoid processing too many rows in one run."""
        return min(actual_max, hwm_value + _MAX_ROWS_PER_RUN)

    # ── CrawlerVisit ──────────────────────────────────────────────────────────

    def _update_crawlers(self):
        hwm = self._upsert('hwm.crawler_visit', 0)
        actual_max = CrawlerVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
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

            cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
            self.stdout.write(f'  crawlers: updated to hwm={new_max}{cap_note}')

            # Daily chart recomputed only when there are new rows
            now = timezone.now()
            rows = (CrawlerVisit.objects
                    .filter(timestamp__gte=now - timedelta(days=60))
                    .annotate(d=TruncDate('timestamp'))
                    .values('d').annotate(c=Count('id')))
            stat = self._upsert('crawlers.daily', {})
            stat.value = {str(r['d']): r['c'] for r in rows}
            stat.save()

            # Per-bot-type daily breakdown for traffic graphs
            bot_rows = (CrawlerVisit.objects
                        .filter(timestamp__gte=now - timedelta(days=60))
                        .annotate(d=TruncDate('timestamp'))
                        .values('d', 'bot_type')
                        .annotate(c=Count('id')))
            daily_by_bot = {}
            for r in bot_rows:
                d = str(r['d'])
                bt = r['bot_type'] or '(empty user agent)'
                if d not in daily_by_bot:
                    daily_by_bot[d] = {}
                daily_by_bot[d][bt] = r['c']
            stat = self._upsert('crawlers.daily_by_bot_type', {})
            stat.value = daily_by_bot
            stat.save()
        else:
            self.stdout.write('  crawlers: no new rows')

    # ── ArchiveVisit ──────────────────────────────────────────────────────────

    def _update_archive(self):
        hwm = self._upsert('hwm.archive_visit', 0)
        actual_max = ArchiveVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
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

            # Bot type via UA classification — chunked to avoid large cursors
            from apps.core.bot_classify import classify_ua
            stat = self._upsert('archive.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True).iterator(chunk_size=5000):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            hwm.value = new_max
            hwm.save()

            cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
            self.stdout.write(f'  archive: updated to hwm={new_max}{cap_note}')

            # Daily chart recomputed only when there are new rows
            now = timezone.now()
            rows = (ArchiveVisit.objects
                    .filter(timestamp__gte=now - timedelta(days=30))
                    .annotate(d=TruncDate('timestamp'))
                    .values('d').annotate(c=Count('id')))
            stat = self._upsert('archive.daily', {})
            stat.value = {str(r['d']): r['c'] for r in rows}
            stat.save()
        else:
            self.stdout.write('  archive: no new rows')

    # ── InboundEmail ──────────────────────────────────────────────────────────

    def _update_emails(self):
        hwm = self._upsert('hwm.inbound_email', 0)
        actual_max = InboundEmail.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
            new_rows = InboundEmail.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('emails.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            stat = self._upsert('emails.by_domain', {})
            for sender in new_rows.values_list('sender', flat=True).iterator(chunk_size=5000):
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

            cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
            self.stdout.write(f'  emails: updated to hwm={new_max}{cap_note}')

            # Daily chart recomputed only when there are new rows
            now = timezone.now()
            rows = (InboundEmail.objects
                    .filter(received_at__gte=now - timedelta(days=30))
                    .annotate(d=TruncDate('received_at'))
                    .values('d').annotate(c=Count('id')))
            stat = self._upsert('emails.daily', {})
            stat.value = {str(r['d']): r['c'] for r in rows}
            stat.save()
        else:
            self.stdout.write('  emails: no new rows')

    # ── PeoplePageVisit ───────────────────────────────────────────────────────

    def _update_people(self):
        hwm = self._upsert('hwm.people_visit', 0)
        actual_max = PeoplePageVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
            new_rows = PeoplePageVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('people.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            from apps.core.bot_classify import classify_ua
            stat = self._upsert('people.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True).iterator(chunk_size=5000):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            stat = self._upsert('people.by_ip', {})
            self._inc_dict(stat, new_rows, 'ip_address')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()

            cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
            self.stdout.write(f'  people: updated to hwm={new_max}{cap_note}')

            # Daily chart recomputed only when there are new rows
            now = timezone.now()
            rows = (PeoplePageVisit.objects
                    .filter(timestamp__gte=now - timedelta(days=30))
                    .annotate(d=TruncDate('timestamp'))
                    .values('d').annotate(c=Count('id')))
            stat = self._upsert('people.daily', {})
            stat.value = {str(r['d']): r['c'] for r in rows}
            stat.save()
        else:
            self.stdout.write('  people: no new rows')

    # ── ProjectPageVisit ──────────────────────────────────────────────────────

    def _update_projects(self):
        hwm = self._upsert('hwm.project_visit', 0)
        actual_max = ProjectPageVisit.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
            new_rows = ProjectPageVisit.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('projects.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            from apps.core.bot_classify import classify_ua
            stat = self._upsert('projects.by_bot_type', {})
            for ua in new_rows.values_list('user_agent', flat=True).iterator(chunk_size=5000):
                k = classify_ua(ua or '')
                stat.value[k] = stat.value.get(k, 0) + 1
            stat.save()

            stat = self._upsert('projects.by_ip', {})
            self._inc_dict(stat, new_rows, 'ip_address')
            stat.value = self._prune(stat.value)
            stat.save()

            hwm.value = new_max
            hwm.save()

            cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
            self.stdout.write(f'  projects: updated to hwm={new_max}{cap_note}')

            # Daily chart recomputed only when there are new rows
            now = timezone.now()
            rows = (ProjectPageVisit.objects
                    .filter(timestamp__gte=now - timedelta(days=30))
                    .annotate(d=TruncDate('timestamp'))
                    .values('d').annotate(c=Count('id')))
            stat = self._upsert('projects.daily', {})
            stat.value = {str(r['d']): r['c'] for r in rows}
            stat.save()
        else:
            self.stdout.write('  projects: no new rows')

    # ── InternalLoginAttempt ──────────────────────────────────────────────────

    def _update_login_attempts(self):
        hwm = self._upsert('hwm.login_attempt', 0)
        actual_max = InternalLoginAttempt.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']
        if not actual_max:
            self.stdout.write('  logins: no new rows')
            return

        new_max = self._cap_new_max(hwm.value, actual_max)
        new_rows = InternalLoginAttempt.objects.filter(id__gt=hwm.value, id__lte=new_max)

        total_stat = self._upsert('login_attempts.total', 0)
        total_stat.value = total_stat.value + new_rows.count()
        total_stat.save()

        by_username = self._upsert('login_attempts.by_username', {})
        self._inc_dict(by_username, new_rows, 'username')
        by_username.value = self._prune(by_username.value)
        by_username.save()

        by_ip = self._upsert('login_attempts.by_ip', {})
        self._inc_dict(by_ip, new_rows, 'ip_address')
        by_ip.value = self._prune(by_ip.value)
        by_ip.save()

        # Group by source (the login page hit): derive from next_url
        by_source = self._upsert('login_attempts.by_source', {})
        for row in new_rows.values('next_url').annotate(c=Count('id')):
            nu = row['next_url'] or ''
            if 'wp-login' in nu or 'wp_login' in nu:
                src = 'wp-login'
            elif 'xmlrpc' in nu:
                src = 'xmlrpc'
            elif 'internal' in nu or nu == '':
                src = 'internal'
            else:
                src = nu[:60] or 'other'
            by_source.value[src] = by_source.value.get(src, 0) + row['c']
        by_source.save()

        # Daily chart — recomputed only when new rows exist
        now = timezone.now()
        daily_stat = self._upsert('login_attempts.daily', {})
        daily_rows = (
            InternalLoginAttempt.objects
            .filter(created_at__gte=now - timedelta(days=60))
            .annotate(d=TruncDate('created_at'))
            .values('d')
            .annotate(c=Count('id'))
        )
        daily_stat.value = {str(r['d']): r['c'] for r in daily_rows}
        daily_stat.save()

        hwm.value = new_max
        hwm.save()

        cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
        self.stdout.write(f'  logins: updated to hwm={new_max}{cap_note}')

    # ── DataOptOutRequest ─────────────────────────────────────────────────────

    def _update_opt_outs(self):
        hwm = self._upsert('hwm.opt_out', 0)
        actual_max = DataOptOutRequest.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']
        if not actual_max:
            self.stdout.write('  optouts: no new rows')
            return

        new_max = self._cap_new_max(hwm.value, actual_max)
        total_stat = self._upsert('optouts.total', 0)
        total_stat.value = total_stat.value + DataOptOutRequest.objects.filter(id__gt=hwm.value, id__lte=new_max).count()
        total_stat.save()

        hwm.value = new_max
        hwm.save()

        cap_note = f' (capped, actual_max={actual_max})' if new_max < actual_max else ''
        self.stdout.write(f'  optouts: updated to hwm={new_max}{cap_note}')

    # ── Traffic graphs ────────────────────────────────────────────────────────

    def _update_graphs(self):
        from pathlib import Path

        from django.conf import settings

        from apps.core.graph_gen import generate_traffic_graphs

        self.stdout.write('  graphs:')
        t0 = time.monotonic()
        try:
            graph_dir = Path(settings.STATIC_ROOT) / 'graphs'
            generate_traffic_graphs(graph_dir, stdout=self.stdout)
            elapsed = time.monotonic() - t0
            self.stdout.write(f'  graphs: done in {elapsed:.1f}s')
        except Exception as exc:
            self.stdout.write(f'  graphs: error — {exc}')

    # ── CanaryToken (always full recompute — cheap) ───────────────────────────

    def _update_canary(self):
        triggered = CanaryToken.objects.filter(triggered=True)
        count = triggered.count()

        stat = self._upsert('canary.triggered_count', 0)
        stat.value = count
        stat.save()

        self.stdout.write(f'  canary: {count} triggered')
