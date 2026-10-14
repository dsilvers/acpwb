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
from django.db.models import Avg, Count, Max, Q, Sum
from django.db.models.functions import TruncDate, TruncMinute
from django.utils import timezone

from apps.core.crawler_queue import oldest_pending_archive_visit, oldest_pending_crawler_visit
from apps.core.db_router import force_db
from apps.core.graph_gen import RECENT_BUCKET_MINUTES, RECENT_BUCKET_RETENTION_DAYS, _bucket_key
from apps.core.models import DashboardStat
from apps.honeypot.models import (
    ArchiveVisit, CanaryToken, CrawlerVisit, EdgeConnection, EdgeRequest, InternalLoginAttempt,
)
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
# Daily chart dicts (crawlers.daily, crawlers.daily_by_bot_type, archive.daily)
# are retained for this many days; older keys are pruned each run.
_DAILY_RETENTION_DAYS = {'crawlers': 60, 'archive': 30}
# CrawlerVisit/ArchiveVisit hold 90M+ rows/day. A daily chart used to be kept
# fresh by re-querying a date-range GROUP BY over the whole chart window on
# every run — first the full 60/30 days, then (after that turned out to force
# TimescaleDB to decompress every compressed chunk in range, since compression
# is segmented by bot_type/trap_type/depth, not by date) a narrowed trailing
# window. Both took 20+ minutes on this table and pinned a PgBouncer
# connection for the duration — and narrowing the window didn't actually fix
# it: even a 2-day, fully-uncompressed range took several minutes under cold
# disk cache, because the cost was the row volume itself, not compression.
# The actual fix is to never re-query historical data for this at all: bucket
# the same small `new_rows` window (already computed above, bounded to
# _MAX_HOURS_PER_RUN) by day and increment the stored dict in place, exactly
# like crawlers.by_trap_type/by_bot_type already do.
#
# The timestamp HWM assumes every row older than it is already in Postgres.
# That's false while the Redis drain is behind: on 2026-10-02 a traffic surge
# put the crawler queue ~2h behind, a handful of current-timestamp rows
# landed anyway, the HWM jumped past the whole backlog, and ~154M rows that
# drained afterwards were never counted. So the window also stops this far
# short of the oldest item still pending in Redis (queue head or an
# in-flight drain batch) — the margin covers items pushed slightly out of
# timestamp order by concurrent workers.
_PENDING_QUEUE_MARGIN = timedelta(minutes=5)


def _asn_orgs(ips):
    """ASN organization names for a few IPs, from the GeoLite2-ASN database
    the edge uses ({} if it isn't there)."""
    from django.conf import settings
    try:
        import geoip2.database
        import geoip2.errors
        reader = geoip2.database.Reader(settings.GEOIP2_ASN_DB_PATH)
    except Exception:
        return {}
    out = {}
    with reader:
        for ip in ips:
            try:
                out[ip] = reader.asn(ip).autonomous_system_organization or ''
            except (geoip2.errors.AddressNotFoundError, ValueError):
                pass
    return out


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

            # This command holds one connection through minutes of Python-side
            # aggregation between queries — route it straight to Postgres so
            # PgBouncer's client_idle_timeout (tuned low to protect the hot
            # request path) doesn't kill it mid-run. See apps.core.db_router.
            with force_db('direct'):
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
                self._update_edge()
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

    def _window_end(self, since, now, oldest_pending_fn, label):
        """Upper timestamp bound for this run's incremental window: at most
        _MAX_HOURS_PER_RUN past `since`, and never at or past rows still
        waiting in the Redis queue (see _PENDING_QUEUE_MARGIN). Returns
        (window_end, capped)."""
        window_end = min(now, since + timedelta(hours=_MAX_HOURS_PER_RUN))
        ok, oldest_pending = oldest_pending_fn()
        if not ok:
            self.stdout.write(f'  {label}: Redis unavailable — pending-queue bound not applied')
        elif oldest_pending is not None:
            window_end = min(window_end, oldest_pending - _PENDING_QUEUE_MARGIN)
        return window_end, window_end < now

    def _cap_new_max(self, hwm_value, actual_max):
        """Cap new_max to avoid processing too many rows in one run."""
        return min(actual_max, hwm_value + _MAX_ROWS_PER_RUN)

    def _set_stat(self, key, value):
        DashboardStat.objects.update_or_create(key=key, defaults={'value': value})

    def _agg_dict(self, qs, field):
        """GROUP BY field → {str(value): count} dict, ordered by count desc, pruned."""
        rows = qs.values(field).annotate(c=Count('id')).order_by('-c')[:_MAX_DICT_ENTRIES]
        return {str(r[field] or ''): r['c'] for r in rows}

    def _inc_daily(self, stat, new_rows, retention_days, now, bot_field=None):
        """Bucket new_rows (already bounded to _MAX_HOURS_PER_RUN) by day and
        add into stat.value in place, then prune anything past retention.
        If bot_field is given, builds {day: {bot_type: count}} instead of
        {day: count} — see the note above _MAX_HOURS_PER_RUN."""
        fields = ['d'] + ([bot_field] if bot_field else [])
        rows = new_rows.annotate(d=TruncDate('timestamp')).values(*fields).annotate(c=Count('id'))
        for row in rows:
            d = str(row['d'])
            if bot_field:
                k = row[bot_field] or '(empty user agent)'
                stat.value.setdefault(d, {})
                stat.value[d][k] = stat.value[d].get(k, 0) + row['c']
            else:
                stat.value[d] = stat.value.get(d, 0) + row['c']
        floor = (now - timedelta(days=retention_days)).date().isoformat()
        stat.value = {d: v for d, v in stat.value.items() if d >= floor}

    def _inc_recent_buckets(self, stat, new_rows, now):
        """Bucket new_rows by RECENT_BUCKET_MINUTES + bot_type into stat.value
        in place, then prune past RECENT_BUCKET_RETENTION_DAYS — feeds the 7d
        traffic graph (apps.core.graph_gen) without it ever re-scanning
        historical rows itself. See the note in graph_gen.py."""
        rows = (new_rows.annotate(bucket=TruncMinute('timestamp'))
                .values('bucket', 'bot_type').annotate(c=Count('id')))
        for row in rows:
            k = _bucket_key(row['bucket'], RECENT_BUCKET_MINUTES)
            bt = row['bot_type'] or '(empty user agent)'
            stat.value.setdefault(k, {})
            stat.value[k][bt] = stat.value[k].get(bt, 0) + row['c']
        floor_key = _bucket_key(now - timedelta(days=RECENT_BUCKET_RETENTION_DAYS), RECENT_BUCKET_MINUTES)
        stat.value = {k: v for k, v in stat.value.items() if k >= floor_key}

    # ── Full recompute ────────────────────────────────────────────────────────

    def _tune_session(self):
        from django.db import connections
        with connections['direct'].cursor() as c:
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
        window_end, capped = self._window_end(since, now, oldest_pending_crawler_visit, 'crawlers')
        if window_end <= since:
            self.stdout.write('  crawlers: waiting on Redis queue backlog')
            return
        new_rows = CrawlerVisit.objects.filter(timestamp__gt=since, timestamp__lte=window_end)
        actual_max_ts = new_rows.aggregate(m=Max('timestamp'))['m']
        if actual_max_ts is None:
            # No rows in this window, but time has moved on — advance the
            # HWM anyway so the next run doesn't re-scan the same empty
            # window (e.g. a quiet period with no traffic at all).
            hwm.value = window_end.isoformat()
            hwm.save()
            self.stdout.write('  crawlers: no new rows')
            return
        new_max_ts = actual_max_ts
        # Every query below must stop where the HWM will — rows landing in
        # (new_max_ts, window_end] mid-run would otherwise be counted here
        # by the later queries and again next run.
        new_rows = CrawlerVisit.objects.filter(timestamp__gt=since, timestamp__lte=new_max_ts)

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

        # Daily chart: incremented from the same bounded new_rows window used
        # above, not a separate date-range re-query — see the note above
        # _MAX_HOURS_PER_RUN.
        now = timezone.now()
        stat = self._upsert('crawlers.daily', {})
        self._inc_daily(stat, new_rows, retention_days=_DAILY_RETENTION_DAYS['crawlers'], now=now)
        stat.save()

        stat = self._upsert('crawlers.daily_by_bot_type', {})
        self._inc_daily(stat, new_rows, retention_days=_DAILY_RETENTION_DAYS['crawlers'], now=now, bot_field='bot_type')
        stat.save()

        stat = self._upsert('crawlers.recent_by_bucket', {})
        self._inc_recent_buckets(stat, new_rows, now)
        stat.save()

    # ── ArchiveVisit ──────────────────────────────────────────────────────────

    def _update_archive(self):
        hwm = self._upsert('hwm.archive_visit_ts', '')
        since = self._seed_since(ArchiveVisit, hwm)
        now = timezone.now()
        window_end, capped = self._window_end(since, now, oldest_pending_archive_visit, 'archive')
        if window_end <= since:
            self.stdout.write('  archive: waiting on Redis queue backlog')
            return
        new_rows = ArchiveVisit.objects.filter(timestamp__gt=since, timestamp__lte=window_end)
        actual_max_ts = new_rows.aggregate(m=Max('timestamp'))['m']
        if actual_max_ts is None:
            hwm.value = window_end.isoformat()
            hwm.save()
            self.stdout.write('  archive: no new rows')
            return
        new_max_ts = actual_max_ts
        # See _update_crawlers — keep every query aligned with the HWM.
        new_rows = ArchiveVisit.objects.filter(timestamp__gt=since, timestamp__lte=new_max_ts)

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

        # Daily chart: incremented from the same bounded new_rows window used
        # above, not a separate date-range re-query — see the note above
        # _MAX_HOURS_PER_RUN.
        now = timezone.now()
        stat = self._upsert('archive.daily', {})
        self._inc_daily(stat, new_rows, retention_days=_DAILY_RETENTION_DAYS['archive'], now=now)
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

    # ── EdgeConnection (rolling 24h window — always recomputed) ───────────────

    def _update_edge(self):
        """Connection-level TLS/HTTP fingerprint summary from acpwb_edge.
        A rolling window rather than HWM-incremental like the others: it
        only ever touches the last day's hypertable chunk, and "what does
        traffic look like right now" is the question it answers.

        Connections are Sum('conn_count'), never Count('id'): an IP past the
        edge's per-minute record budget is folded into aggregated rows."""
        since = timezone.now() - timedelta(hours=24)
        qs = EdgeConnection.objects.filter(timestamp__gte=since)
        conns = Sum('conn_count')

        def by(field, rows=qs):
            return {
                (r[field] or '(none)'): r['c']
                for r in rows.values(field).annotate(c=conns).order_by('-c')[:20]
            }

        def top_ja4(rows):
            return [
                {'ja4': r['ja4'], 'count': r['c'], 'ips': r['ips'], 'uas': r['uas']}
                for r in rows.exclude(ja4='').values('ja4').annotate(
                    c=conns,
                    ips=Count('ip_address', distinct=True),
                    uas=Count('first_user_agent', distinct=True),
                ).order_by('-c')[:20]
            ]

        summary = {
            'total': qs.aggregate(c=conns)['c'] or 0,
            'by_outcome': by('outcome'),
            'by_alpn': by('alpn'),
            'by_tls_version': by('tls_version'),
            'by_close_reason': by('close_reason'),
            'by_goaway': by('h2_goaway', qs.exclude(h2_goaway='')),
            'aggregated_conns': qs.filter(aggregated=True).aggregate(c=conns)['c'] or 0,
            'rejected_requests': qs.aggregate(c=Sum('rejected_requests'))['c'] or 0,
            'top_ja4': top_ja4(qs),
            'top_failed_ja4': top_ja4(qs.exclude(outcome__in=['ok', 'plain'])),
            'top_h2_fingerprint': [
                {'h2_fingerprint': r['h2_fingerprint'], 'count': r['c'], 'ips': r['ips']}
                for r in qs.exclude(h2_fingerprint='').values('h2_fingerprint').annotate(
                    c=conns, ips=Count('ip_address', distinct=True),
                ).order_by('-c')[:20]
            ],
            # HTTP/2 attack signatures, per IP: resets ≈ streams is rapid
            # reset; lots of continuations is a CONTINUATION flood.
            'top_h2_abuse': [
                {'ip_address': r['ip_address'], 'conns': r['c'], 'streams': r['streams'],
                 'resets': r['resets'], 'continuations': r['conts']}
                for r in qs.filter(alpn='h2').values('ip_address').annotate(
                    c=conns, streams=Sum('h2_streams'), resets=Sum('h2_resets'),
                    conts=Sum('h2_continuations'),
                ).filter(Q(resets__gt=0) | Q(conts__gt=0)).order_by('-resets', '-conts')[:20]
            ],
            # Connection floods: who opened the most connections.
            'top_conn_ips': [
                {'ip_address': r['ip_address'], 'conns': r['c'], 'requests': r['reqs'], 'ja4s': r['ja4s']}
                for r in qs.values('ip_address').annotate(
                    c=conns, reqs=Sum('request_count'), ja4s=Count('ja4', distinct=True, filter=~Q(ja4='')),
                ).order_by('-c')[:20]
            ],
        }
        self._set_stat('edge.summary_24h', summary)
        self._set_stat('edge.process', self._edge_process_stats())
        self.stdout.write(f'  edge: {summary["total"]} connections in last 24h')
        self._update_edge_requests(since)
        self._update_edge_classes()

    def _update_edge_requests(self, since):
        """Request-level summary from EdgeRequest (every request the edge
        saw, static files and failures included) over the same window."""
        qs = EdgeRequest.objects.filter(timestamp__gte=since)
        totals = qs.aggregate(
            total=Count('id'),
            clients=Count('client_id', distinct=True, filter=~Q(client_id='')),
            new_clients=Count('id', filter=Q(new_client=True)),
            ips=Count('ip_address', distinct=True),
            failed=Count('id', filter=~Q(error='')),
            bytes_out=Sum('response_bytes'),
        )
        summary = {k: v or 0 for k, v in totals.items()}
        summary['by_route'] = [
            {'name': r['route'] or '(none)', 'count': r['c'], 'avg_ms': round((r['avg_us'] or 0) / 1000, 1),
             'bytes': r['b'] or 0}
            for r in qs.values('route').annotate(
                c=Count('id'), avg_us=Avg('duration_us'), b=Sum('response_bytes'),
            ).order_by('-c')[:20]
        ]
        summary['by_status'] = {
            str(r['status']): r['c'] for r in qs.values('status').annotate(c=Count('id')).order_by('-c')[:20]
        }
        summary['by_error'] = {
            r['error']: r['c']
            for r in qs.exclude(error='').values('error').annotate(c=Count('id')).order_by('-c')
        }
        summary['by_protocol'] = {
            (r['protocol'] or '(none)'): r['c']
            for r in qs.values('protocol').annotate(c=Count('id')).order_by('-c')
        }
        self._set_stat('edge.requests_24h', summary)
        self.stdout.write(f'  edge: {summary["total"]} requests in last 24h')

    def _update_edge_classes(self):
        """Who's really visiting, from the hourly rollups (EdgeIPHour,
        EdgeFingerprintHour; rollup_edge_hour): requests by the edge's
        client class and network, the fingerprints faking browsers, and
        the residential networks those come from. The last 24 finished
        hours, plus 48 hours of per-hour totals."""
        from apps.honeypot.models import EdgeFingerprintHour, EdgeIPHour
        now = timezone.now()
        since = now.replace(minute=0, second=0, microsecond=0) - timedelta(hours=24)
        ips = EdgeIPHour.objects.filter(hour__gte=since)
        classes = ('browser', 'headless', 'impersonator', 'known_bot', 'tool', 'unknown')
        sums = {f'{c}_reqs': Sum(f'{c}_reqs') for c in classes}

        by_network = {}
        for r in ips.values('network').annotate(
            requests=Sum('requests'), ip_count=Count('ip_address', distinct=True),
            fake_ips=Count('ip_address', distinct=True, filter=Q(impersonator_reqs__gt=0)),
            no_hello_ips=Count('ip_address', distinct=True, filter=Q(no_hello_conns__gt=0)),
            no_hello=Sum('no_hello_conns'), failed=Sum('failed_handshakes'),
            **sums,
        ).order_by('-requests'):
            row = {
                'network': r['network'] or 'unknown', 'requests': r['requests'] or 0, 'ips': r['ip_count'],
                'fake_ips': r['fake_ips'], 'no_hello_ips': r['no_hello_ips'], 'no_hello': r['no_hello'] or 0,
                'failed_handshakes': r['failed'] or 0, **{c: r[f'{c}_reqs'] or 0 for c in classes},
            }
            # '' (rows from before the edge recorded networks) is unknown too.
            prev = by_network.get(row['network'])
            by_network[row['network']] = {k: v if k == 'network' else v + prev[k] for k, v in row.items()} if prev else row
        by_network = sorted(by_network.values(), key=lambda r: -r['requests'])

        hourly = [
            {'hour': r['hour'].isoformat(), 'requests': r['requests'] or 0,
             'residential_fake': r['residential_fake'] or 0, **{c: r[f'{c}_reqs'] or 0 for c in classes}}
            for r in EdgeIPHour.objects.filter(hour__gte=since - timedelta(hours=24)).values('hour').annotate(
                requests=Sum('requests'),
                residential_fake=Sum('impersonator_reqs', filter=Q(network='residential')),
                **sums,
            ).order_by('-hour')
        ]

        # Fingerprints faking browsers. ip_hours sums each hour's distinct
        # IPs: an IP active in three hours counts three times.
        top_fake_fps = [
            {'ja4': r['ja4'], 'h2_fingerprint': r['h2_fingerprint'], 'requests': r['imp'] or 0,
             'conns': r['conns'] or 0, 'ip_hours': r['ip_hours'] or 0, 'single_conn': r['single'] or 0,
             'residential': r['res'] or 0, 'hosting': r['host'] or 0,
             'stack': r['stack'], 'user_agent': r['ua']}
            for r in EdgeFingerprintHour.objects.filter(hour__gte=since, impersonator_reqs__gt=0)
            .values('ja4', 'h2_fingerprint').annotate(
                imp=Sum('impersonator_reqs'), conns=Sum('conns'), ip_hours=Sum('ips'),
                single=Sum('single_conn_ips'), res=Sum('residential_ips'), host=Sum('hosting_ips'),
                stack=Max('top_stack'), ua=Max('top_user_agent'),
            ).order_by('-ip_hours')[:20]
        ]

        # Residential networks sending fake browsers, or whose IPs open and
        # drop connections without a hello (proxy exits): candidate pools.
        asn_rows = list(
            ips.filter(Q(impersonator_reqs__gt=0) | Q(no_hello_conns__gt=0), network='residential')
            .values('asn').annotate(
                ip_count=Count('ip_address', distinct=True),
                fake_ips=Count('ip_address', distinct=True, filter=Q(impersonator_reqs__gt=0)),
                no_hello_ips=Count('ip_address', distinct=True, filter=Q(no_hello_conns__gt=0)),
                fake=Sum('impersonator_reqs'), requests=Sum('requests'), sample_ip=Max('ip_address'),
            ).order_by('-ip_count')[:20]
        )
        orgs = _asn_orgs([r['sample_ip'] for r in asn_rows])
        top_proxy_asns = [
            {'asn': r['asn'], 'org': orgs.get(r['sample_ip'], ''), 'ips': r['ip_count'],
             'fake_ips': r['fake_ips'], 'no_hello_ips': r['no_hello_ips'],
             'fake_requests': r['fake'] or 0, 'requests': r['requests'] or 0}
            for r in asn_rows
        ]

        top_fake_ips = [
            {'ip_address': r['ip_address'], 'network': r['network'] or 'unknown', 'asn': r['asn'],
             'fake_requests': r['fake'] or 0, 'requests': r['requests'] or 0, 'stack': r['stack'],
             'traps': r['traps'] or 0}
            for r in ips.filter(impersonator_reqs__gt=0).values('ip_address').annotate(
                fake=Sum('impersonator_reqs'), requests=Sum('requests'), network=Max('network'),
                asn=Max('asn'), stack=Max('top_stack'), traps=Sum('trap_hits'),
            ).order_by('-fake')[:20]
        ]

        self._set_stat('edge.classes_24h', {
            'by_network': by_network, 'hourly': hourly, 'top_fake_fps': top_fake_fps,
            'top_proxy_asns': top_proxy_asns, 'top_fake_ips': top_fake_ips,
            **self._edge_behaviour(since),
        })
        self.stdout.write(f'  edge: classes from {len(hourly)} rolled-up hour(s)')

    def _edge_behaviour(self, since):
        """What clients did, from EdgeClientHour (client cookies that came
        back) and EdgeFingerprintHour: automation that a single request
        can't show, however real its browser fingerprint.

        exposed: client-hours by the class of most of their requests, and
        how many of them behaved like no person does: the cookie on
        POOL_IPS+ IPs in one hour (a proxy pool's shared jar), a new client
        loading pages but never their assets (not rendering them), or a hit
        on an ABSOLUTE_PROOF_TRAP_TYPES trap (a ghost link no one can see).
        """
        from apps.core.management.commands.rollup_edge_hour import POOL_IPS
        from apps.honeypot.models import EdgeClientHour, EdgeFingerprintHour
        clients = EdgeClientHour.objects.filter(hour__gte=since)
        pooled = Q(ips__gte=POOL_IPS)
        no_render = Q(first_seen=True, page_reqs__gte=2, static_reqs=0)
        trapped = Q(trap_reqs__gt=0)
        exposed = [
            {'client_class': r['top_class'] or 'unclassified', 'client_hours': r['n'], 'pooled': r['pooled'],
             'no_render': r['no_render'], 'trapped': r['trapped'], 'any': r['any']}
            for r in clients.values('top_class').annotate(
                n=Count('client_id'), pooled=Count('client_id', filter=pooled),
                no_render=Count('client_id', filter=no_render), trapped=Count('client_id', filter=trapped),
                any=Count('client_id', filter=pooled | no_render | trapped),
            ).order_by('-n')
        ]
        multi = clients.filter(ips__gte=2).aggregate(
            multi=Count('client_id', distinct=True),
            pools=Count('client_id', distinct=True, filter=pooled),
            ten_plus=Count('client_id', distinct=True, filter=Q(ips__gte=10)),
            max_ips=Max('ips'),
        )
        multi['clients'] = clients.aggregate(n=Count('client_id', distinct=True))['n']
        # ip_hours sums each hour's distinct IPs, like the fingerprints'.
        cookie_pools = [
            {'client_id': r['client_id'], 'ip_hours': r['ip_hours'], 'max_ips': r['max_ips'], 'hours': r['hours'],
             'requests': r['requests'] or 0, 'residential': r['res'] or 0, 'hosting': r['host'] or 0,
             'traps': r['traps'] or 0, 'client_class': r['cls'], 'user_agent': r['ua'], 'ja4': r['ja4']}
            for r in clients.filter(pooled).values('client_id').annotate(
                ip_hours=Sum('ips'), max_ips=Max('ips'), hours=Count('hour'), requests=Sum('requests'),
                res=Sum('residential_ips'), host=Sum('hosting_ips'), traps=Sum('trap_reqs'),
                cls=Max('top_class'), ua=Max('top_user_agent'), ja4=Max('top_ja4'),
            ).order_by('-ip_hours')[:20]
        ]
        # Stacks whose requests walked into traps; browser_reqs > 0 means a
        # real browser fingerprint that a trap shows is driven by a bot.
        trap_fps = [
            {'ja4': r['ja4'], 'h2_fingerprint': r['h2_fingerprint'], 'trap_reqs': r['traps'] or 0,
             'trap_ips': r['trap_ips'] or 0, 'ip_hours': r['ip_hours'] or 0, 'requests': r['requests'] or 0,
             'browser': r['browser'] or 0, 'headless': r['headless'] or 0, 'impersonator': r['imp'] or 0,
             'stack': r['stack'], 'user_agent': r['ua']}
            for r in EdgeFingerprintHour.objects.filter(hour__gte=since, trap_reqs__gt=0)
            .values('ja4', 'h2_fingerprint').annotate(
                traps=Sum('trap_reqs'), trap_ips=Sum('trap_ips'), ip_hours=Sum('ips'), requests=Sum('requests'),
                browser=Sum('browser_reqs'), headless=Sum('headless_reqs'), imp=Sum('impersonator_reqs'),
                stack=Max('top_stack'), ua=Max('top_user_agent'),
            ).order_by('-trap_ips')[:20]
        ]
        return {
            'exposed': exposed, 'multi_ip': {k: v or 0 for k, v in multi.items()} | {'pool_ips': POOL_IPS},
            'cookie_pools': cookie_pools, 'trap_fps': trap_fps,
        }

    def _edge_process_stats(self):
        """acpwb_edge's own counters (acpwb:edge_stats, refreshed every 30s):
        what never becomes a row — connection and request records dropped
        because Redis was down or the queue full (records_dropped_conn /
        records_dropped_request), overload 503s, parser panics. Empty if
        the edge isn't running or Redis is unreachable."""
        from apps.core.crawler_queue import _get_consumer_client
        r = _get_consumer_client()
        if r is None:
            return {}
        try:
            raw = r.hgetall('acpwb:edge_stats')
        except Exception:
            return {}
        out = {}
        for k, v in raw.items():
            try:
                out[k] = int(v)
            except (TypeError, ValueError):
                pass
        return out
