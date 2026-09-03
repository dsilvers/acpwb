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
from django.db import transaction
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
# CrawlerVisit/ArchiveVisit only track a timestamp HWM (not id — see the
# hypertable migration), so their per-run cap can't be a cheap `hwm + N`
# arithmetic bound like the id-keyed models below. It used to be computed by
# peeking at the timestamp of the (hwm + 500_000)th row via
# qs[500_000:500_001] — an OFFSET 500000 query that Postgres must scan/sort
# (and, on these compressed TimescaleDB hypertables, decompress) in full just
# to reach that one row. As traffic grew, that peek alone started taking
# 25+ minutes and pinning several PgBouncer connections — capping by a fixed
# wall-clock window instead needs no such peek: the window bound is pure
# arithmetic, and the resulting bounded-timestamp queries are ordinary
# indexed range scans.
_MAX_HOURS_PER_RUN = 2


class Command(BaseCommand):
    help = 'Incrementally update dashboard stats in the database.'

    def add_arguments(self, parser):
        parser.add_argument(
            '--reset-bot-types', action='store_true',
            help='Full recompute of bot_type/bot_group from all rows (use after backfill_bot_types)',
        )
        parser.add_argument(
            '--full-recompute', action='store_true',
            help='Recompute all stats from scratch via aggregate queries, then advance HWMs. '
                 'Skips Python-classified bot breakdowns (archive/people/projects.by_bot_type).',
        )

    def handle(self, *args, **options):
        self.reset_bot_types = options['reset_bot_types']
        with open('/tmp/precalc_dashboard.lock', 'w') as lockfile:
            try:
                fcntl.flock(lockfile, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                self.stdout.write('precalc_dashboard already running — skipping.')
                return

            run_start = time.monotonic()
            now_str = timezone.now().strftime('%Y-%m-%d %H:%M:%S %Z')
            self.stdout.write(f'precalc_dashboard starting at {now_str}')

            if options['full_recompute']:
                self.stdout.write('Mode: full recompute')
                self._tune_session()
                self._full_recompute_crawlers()
                self._full_recompute_archive()
                self._full_recompute_emails()
                self._full_recompute_people()
                self._full_recompute_projects()
                self._full_recompute_login_attempts()
                self._full_recompute_opt_outs()
            else:
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

    def _seed_since(self, model, hwm):
        """Starting point for a timestamp-HWM model's incremental scan.

        On a fresh (never-run) watermark, seed from the model's actual
        earliest row instead of datetime.min — otherwise the time-windowed
        cap below would need to advance _MAX_HOURS_PER_RUN at a time
        starting from year 1, taking an enormous number of cron ticks to
        reach real data. Finding the earliest row is cheap (hits the
        leading edge of the timestamp index/oldest chunk), unlike the old
        approach this file used to bound each run's row count with.
        """
        from datetime import datetime, timedelta as td, timezone as dt_tz
        if hwm.value:
            return datetime.fromisoformat(hwm.value)
        earliest = model.objects.order_by('timestamp').values_list('timestamp', flat=True).first()
        if earliest is None:
            return datetime.min.replace(tzinfo=dt_tz.utc)
        return earliest - td(microseconds=1)

    def _cap_new_max(self, hwm_value, actual_max):
        """Cap new_max to avoid processing too many rows in one run."""
        return min(actual_max, hwm_value + _MAX_ROWS_PER_RUN)

    def _set_stat(self, key, value):
        DashboardStat.objects.update_or_create(key=key, defaults={'value': value})

    def _agg_dict(self, qs, field):
        """GROUP BY field → {str(value): count} dict, ordered by count desc, pruned."""
        rows = qs.values(field).annotate(c=Count('id')).order_by('-c')[:_MAX_DICT_ENTRIES]
        return {str(r[field] or ''): r['c'] for r in rows}

    # ── Full recompute ────────────────────────────────────────────────────────

    def _tune_session(self):
        from django.db import connection
        with connection.cursor() as c:
            c.execute("SET work_mem = '4GB'")
            c.execute("SET max_parallel_workers_per_gather = 8")
            c.execute("SET enable_partitionwise_aggregate = on")
        self.stdout.write('  session tuned (work_mem=4GB, parallel_workers=8, partitionwise_aggregate=on)')

    def _step(self, msg):
        self.stdout.write(f'    {msg}', ending='')
        self.stdout.flush()

    def _done(self, msg=''):
        self.stdout.write(f' done{(" — " + msg) if msg else ""}')

    def _full_recompute_crawlers(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  crawlers:')

        self._step('total...')
        total = CrawlerVisit.objects.count()
        self._set_stat('crawlers.total', total)
        self._done(f'{total:,}')

        for label, key, qs, field in [
            ('by_trap_type', 'crawlers.by_trap_type', CrawlerVisit.objects,                          'trap_type'),
            ('by_bot_type',  'crawlers.by_bot_type',  CrawlerVisit.objects,                          'bot_type'),
            ('by_bot_group', 'crawlers.by_bot_group', CrawlerVisit.objects,                          'bot_group'),
            ('by_ip',        'crawlers.by_ip',         CrawlerVisit.objects,                         'ip_address'),
            ('by_path',      'crawlers.by_path',       CrawlerVisit.objects,                         'path'),
            ('by_host',      'crawlers.by_host',       CrawlerVisit.objects.exclude(host=''),        'host'),
        ]:
            self._step(f'{label}...')
            self._set_stat(key, self._agg_dict(qs, field))
            self._done()

        probe_types = ['env_probe', 'wp_probe', 'webshell_probe', 'scanner_probe']
        self._step('probe_by_path...')
        self._set_stat('crawlers.probe_by_path',
            self._agg_dict(CrawlerVisit.objects.filter(trap_type__in=probe_types), 'path'))
        self._done()

        self._step('webshell_cmds...')
        self._set_stat('crawlers.webshell_cmds',
            self._agg_dict(CrawlerVisit.objects.filter(trap_type='webshell_probe').exclude(query_string=''), 'query_string'))
        self._done()

        self._step('daily (60d)...')
        rows = (CrawlerVisit.objects
                .filter(timestamp__gte=now - timedelta(days=60))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        self._set_stat('crawlers.daily', {str(r['d']): r['c'] for r in rows})
        self._done()

        self._step('daily_by_bot_type (60d)...')
        bot_rows = (CrawlerVisit.objects
                    .filter(timestamp__gte=now - timedelta(days=60))
                    .annotate(d=TruncDate('timestamp'))
                    .values('d', 'bot_type').annotate(c=Count('id')))
        daily_by_bot = {}
        for r in bot_rows:
            d = str(r['d'])
            bt = r['bot_type'] or '(empty user agent)'
            daily_by_bot.setdefault(d, {})[bt] = r['c']
        self._set_stat('crawlers.daily_by_bot_type', daily_by_bot)
        self._done()

        self._step('advancing HWM...')
        max_ts = CrawlerVisit.objects.aggregate(m=Max('timestamp'))['m']
        self._set_stat('hwm.crawler_visit_ts', max_ts.isoformat() if max_ts else '')
        self._done()

    def _full_recompute_archive(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  archive: (by_bot_type skipped — Python-classified)')

        self._step('total...')
        total = ArchiveVisit.objects.count()
        self._set_stat('archive.total', total)
        self._done(f'{total:,}')

        computed = {}
        for label, key, field in [
            ('by_depth', 'archive.by_depth', 'depth'),
            ('by_ip',    'archive.by_ip',    'ip_address'),
            ('by_slug',  'archive.by_slug',  'slug'),
        ]:
            self._step(f'{label}...')
            d = self._agg_dict(ArchiveVisit.objects, field)
            self._set_stat(key, d)
            computed[label] = d
            self._done()

        self._step('max_depth_by_ip...')
        # Scoped to by_ip's own key set (rather than pruned by depth value)
        # so this stays consistent with what the dashboard actually looks up.
        ip_keys = list(computed['by_ip'].keys())
        rows = (ArchiveVisit.objects.filter(ip_address__in=ip_keys)
                .values('ip_address').annotate(m=Max('depth')))
        self._set_stat('archive.max_depth_by_ip', {str(r['ip_address'] or ''): r['m'] for r in rows})
        self._done()

        self._step('daily (30d)...')
        rows = (ArchiveVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        self._set_stat('archive.daily', {str(r['d']): r['c'] for r in rows})
        self._done()

        self._step('advancing HWM...')
        max_ts = ArchiveVisit.objects.aggregate(m=Max('timestamp'))['m']
        self._set_stat('hwm.archive_visit_ts', max_ts.isoformat() if max_ts else '')
        self._done()

    def _full_recompute_emails(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  emails:')

        self._step('total...')
        total = InboundEmail.objects.count()
        self._set_stat('emails.total', total)
        self._done(f'{total:,}')

        self._step('by_recipient...')
        self._set_stat('emails.by_recipient', self._agg_dict(InboundEmail.objects, 'recipient'))
        self._done()

        self._step('by_domain...')
        by_domain = {}
        # PgBouncer transaction-pooling mode can hand this iterator()'s
        # server-side cursor to a different backend between FETCHes unless
        # the whole streamed loop stays in one transaction.
        with transaction.atomic():
            for sender in InboundEmail.objects.values_list('sender', flat=True).iterator(chunk_size=5000):
                domain = sender.split('@')[-1].lower() if '@' in sender else sender
                by_domain[domain] = by_domain.get(domain, 0) + 1
        self._set_stat('emails.by_domain', self._prune(by_domain))
        self._done()

        self._step('daily (30d)...')
        rows = (InboundEmail.objects
                .filter(received_at__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('received_at'))
                .values('d').annotate(c=Count('id')))
        self._set_stat('emails.daily', {str(r['d']): r['c'] for r in rows})
        self._done()

        self._step('advancing HWM...')
        max_id = InboundEmail.objects.aggregate(m=Max('id'))['m'] or 0
        self._set_stat('hwm.inbound_email', max_id)
        self._done()

    def _full_recompute_people(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  people: (by_bot_type skipped — Python-classified)')

        self._step('total...')
        total = PeoplePageVisit.objects.count()
        self._set_stat('people.total', total)
        self._done(f'{total:,}')

        self._step('by_ip...')
        self._set_stat('people.by_ip', self._agg_dict(PeoplePageVisit.objects, 'ip_address'))
        self._done()

        self._step('daily (30d)...')
        rows = (PeoplePageVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        self._set_stat('people.daily', {str(r['d']): r['c'] for r in rows})
        self._done()

        self._step('advancing HWM...')
        max_id = PeoplePageVisit.objects.aggregate(m=Max('id'))['m'] or 0
        self._set_stat('hwm.people_visit', max_id)
        self._done()

    def _full_recompute_projects(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  projects: (by_bot_type skipped — Python-classified)')

        self._step('total...')
        total = ProjectPageVisit.objects.count()
        self._set_stat('projects.total', total)
        self._done(f'{total:,}')

        self._step('by_ip...')
        self._set_stat('projects.by_ip', self._agg_dict(ProjectPageVisit.objects, 'ip_address'))
        self._done()

        self._step('daily (30d)...')
        rows = (ProjectPageVisit.objects
                .filter(timestamp__gte=now - timedelta(days=30))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        self._set_stat('projects.daily', {str(r['d']): r['c'] for r in rows})
        self._done()

        self._step('advancing HWM...')
        max_id = ProjectPageVisit.objects.aggregate(m=Max('id'))['m'] or 0
        self._set_stat('hwm.project_visit', max_id)
        self._done()

    def _full_recompute_login_attempts(self):
        from datetime import timedelta
        now = timezone.now()
        self.stdout.write('  logins:')

        self._step('total...')
        total = InternalLoginAttempt.objects.count()
        self._set_stat('login_attempts.total', total)
        self._done(f'{total:,}')

        self._step('by_username...')
        self._set_stat('login_attempts.by_username', self._agg_dict(InternalLoginAttempt.objects, 'username'))
        self._done()

        self._step('by_ip...')
        self._set_stat('login_attempts.by_ip', self._agg_dict(InternalLoginAttempt.objects, 'ip_address'))
        self._done()

        self._step('by_source...')
        by_source = {}
        for row in InternalLoginAttempt.objects.values('next_url').annotate(c=Count('id')):
            nu = row['next_url'] or ''
            if 'wp-login' in nu or 'wp_login' in nu:
                src = 'wp-login'
            elif 'xmlrpc' in nu:
                src = 'xmlrpc'
            elif 'internal' in nu or nu == '':
                src = 'internal'
            else:
                src = nu[:60] or 'other'
            by_source[src] = by_source.get(src, 0) + row['c']
        self._set_stat('login_attempts.by_source', by_source)
        self._done()

        self._step('daily (60d)...')
        daily_rows = (InternalLoginAttempt.objects
                      .filter(created_at__gte=now - timedelta(days=60))
                      .annotate(d=TruncDate('created_at'))
                      .values('d').annotate(c=Count('id')))
        self._set_stat('login_attempts.daily', {str(r['d']): r['c'] for r in daily_rows})
        self._done()

        self._step('advancing HWM...')
        max_id = InternalLoginAttempt.objects.aggregate(m=Max('id'))['m'] or 0
        self._set_stat('hwm.login_attempt', max_id)
        self._done()

    def _full_recompute_opt_outs(self):
        from apps.public.models import DataOptOutRequest
        self.stdout.write('  optouts:')
        self._step('total...')
        total = DataOptOutRequest.objects.count()
        self._set_stat('optouts.total', total)
        max_id = DataOptOutRequest.objects.aggregate(m=Max('id'))['m'] or 0
        self._set_stat('hwm.opt_out', max_id)
        self._done(f'{total:,}')

    # ── CrawlerVisit ──────────────────────────────────────────────────────────

    def _update_crawlers(self):
        hwm = self._upsert('hwm.crawler_visit_ts', '')
        since = self._seed_since(CrawlerVisit, hwm)
        now = timezone.now()
        window_end = min(now, since + timedelta(hours=_MAX_HOURS_PER_RUN))
        capped = window_end < now
        new_rows = CrawlerVisit.objects.filter(timestamp__gt=since, timestamp__lte=window_end)
        actual_max_ts = new_rows.aggregate(m=Max('timestamp'))['m']
        if actual_max_ts is None:
            if window_end > since:
                # No rows in this window, but time has moved on — advance the
                # HWM anyway so the next run doesn't re-scan the same empty
                # window (e.g. a quiet period with no traffic at all).
                hwm.value = window_end.isoformat()
                hwm.save()
            self.stdout.write('  crawlers: no new rows')
            return
        new_max_ts = actual_max_ts

        total_stat = self._upsert('crawlers.total', 0)
        total_stat.value = total_stat.value + new_rows.count()
        total_stat.save()

        stat = self._upsert('crawlers.by_trap_type', {})
        self._inc_dict(stat, new_rows, 'trap_type')
        stat.save()

        for key, field in [
            ('crawlers.by_bot_type',  'bot_type'),
            ('crawlers.by_bot_group', 'bot_group'),
        ]:
            stat = self._upsert(key, {})
            if self.reset_bot_types:
                stat.value = {
                    str(r[field] or ''): r['c']
                    for r in CrawlerVisit.objects.values(field).annotate(c=Count('id'))
                }
            else:
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

        hwm.value = new_max_ts.isoformat()
        hwm.save()

        cap_note = ' (window bounded — backlog may remain)' if capped else ''
        self.stdout.write(f'  crawlers: updated to hwm_ts={new_max_ts}{cap_note}')

        # Daily chart: full 60-day recompute — always authoritative, never overwrites
        # a correct stored value with a partial/stale query result.
        now = timezone.now()
        rows = (CrawlerVisit.objects
                .filter(timestamp__gte=now - timedelta(days=60))
                .annotate(d=TruncDate('timestamp'))
                .values('d').annotate(c=Count('id')))
        stat = self._upsert('crawlers.daily', {})
        stat.value = {str(r['d']): r['c'] for r in rows}
        stat.save()

        # Per-bot-type daily breakdown for all-time traffic graph
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

    # ── ArchiveVisit ──────────────────────────────────────────────────────────

    def _update_archive(self):
        hwm = self._upsert('hwm.archive_visit_ts', '')
        since = self._seed_since(ArchiveVisit, hwm)
        now = timezone.now()
        window_end = min(now, since + timedelta(hours=_MAX_HOURS_PER_RUN))
        capped = window_end < now
        new_rows = ArchiveVisit.objects.filter(timestamp__gt=since, timestamp__lte=window_end)
        actual_max_ts = new_rows.aggregate(m=Max('timestamp'))['m']
        if actual_max_ts is None:
            if window_end > since:
                hwm.value = window_end.isoformat()
                hwm.save()
            self.stdout.write('  archive: no new rows')
            return
        new_max_ts = actual_max_ts

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
        by_ip_keys = set(stat.value.keys())

        stat = self._upsert('archive.by_slug', {})
        self._inc_dict(stat, new_rows, 'slug')
        stat.value = self._prune(stat.value)
        stat.save()

        # Scoped to by_ip's current key set (rather than pruned by depth
        # value) so this stays consistent with what the dashboard looks up.
        stat = self._upsert('archive.max_depth_by_ip', {})
        for row in new_rows.values('ip_address').annotate(m=Max('depth')):
            ip = str(row['ip_address'] or '')
            if ip in by_ip_keys:
                stat.value[ip] = max(stat.value.get(ip, 0), row['m'])
        stat.value = {ip: d for ip, d in stat.value.items() if ip in by_ip_keys}
        stat.save()

        # Bot type via UA + IP classification — chunked to avoid large cursors.
        # PgBouncer transaction-pooling mode can hand this iterator()'s
        # server-side cursor to a different backend between FETCHes unless
        # the whole streamed loop stays in one transaction.
        from apps.core.bot_classify import classify_ua_or_ip
        stat = self._upsert('archive.by_bot_type', {})
        with transaction.atomic():
            for ua, ip in new_rows.values_list('user_agent', 'ip_address').iterator(chunk_size=5000):
                k = classify_ua_or_ip(ua or '', ip or '')
                stat.value[k] = stat.value.get(k, 0) + 1
        stat.save()

        hwm.value = new_max_ts.isoformat()
        hwm.save()

        cap_note = ' (window bounded — backlog may remain)' if capped else ''
        self.stdout.write(f'  archive: updated to hwm_ts={new_max_ts}{cap_note}')

        # Daily chart recomputed only when there are new rows
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
        actual_max = InboundEmail.objects.filter(id__gt=hwm.value).aggregate(m=Max('id'))['m']

        if actual_max:
            new_max = self._cap_new_max(hwm.value, actual_max)
            new_rows = InboundEmail.objects.filter(id__gt=hwm.value, id__lte=new_max)
            total_stat = self._upsert('emails.total', 0)
            total_stat.value = total_stat.value + new_rows.count()
            total_stat.save()

            stat = self._upsert('emails.by_domain', {})
            with transaction.atomic():  # PgBouncer transaction pooling + server-side cursor
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

            from apps.core.bot_classify import classify_ua_or_ip
            stat = self._upsert('people.by_bot_type', {})
            with transaction.atomic():  # PgBouncer transaction pooling + server-side cursor
                for ua, ip in new_rows.values_list('user_agent', 'ip_address').iterator(chunk_size=5000):
                    k = classify_ua_or_ip(ua or '', ip or '')
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

            from apps.core.bot_classify import classify_ua_or_ip
            stat = self._upsert('projects.by_bot_type', {})
            with transaction.atomic():  # PgBouncer transaction pooling + server-side cursor
                for ua, ip in new_rows.values_list('user_agent', 'ip_address').iterator(chunk_size=5000):
                    k = classify_ua_or_ip(ua or '', ip or '')
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
