"""
Roll CrawlerVisit up into CrawlerMinute (one row per minute, bot and trap),
which per-bot and per-trap trend graphs read.

Each run recounts the last RECOUNT before the watermark as well as whatever
is new, so visits that reached the database late (the drains run every
minute) are picked up. A minute is RUN_DELAY old before it's first counted.
A range is replaced whole (delete + insert in one transaction), so re-running
one is safe; do that with --from/--to after a drain has been stalled for
longer than RECOUNT.

--backfill walks older history an hour at a time, up to the oldest minute
rolled up so far, resuming from a saved cursor. It takes a separate lock so
the cron run keeps going meanwhile; the two never touch the same minutes.

--reclassify rebuilds a range after bots were added to bot_classify, without
touching CrawlerVisit: raw chunks are compressed and segmented by bot_type, so
updating old rows would decompress nearly all of them. Stored labels are kept
except on unclassified rows, whose user_agent is looked up again. It resumes
from its own cursor and replaces --backfill, whose cursor it clears: anything
that later re-runs those hours from stored labels undoes it.

Run via cron every 5 minutes (see deploy/acpwb-crontab).
"""
import fcntl
import time
from collections import Counter
from datetime import datetime, timedelta, timezone as dt_timezone
from functools import lru_cache

from django.core.management.base import BaseCommand, CommandError
from django.db import connection, transaction
from django.db.models import Min
from django.utils import timezone

from apps.core.bot_classify import bot_type_to_group, classify_ua
from apps.core.management.commands.rollup_edge_hour import floor_hour
from apps.core.models import DashboardStat
from apps.honeypot.models import CrawlerMinute

_LOCK_FILE = '/tmp/acpwb-rollup-crawler-minute.lock'
_BACKFILL_LOCK_FILE = '/tmp/acpwb-rollup-crawler-minute-backfill.lock'
_WATERMARK_KEY = 'crawler_minute_watermark'  # {'minute': ISO of the first minute not yet counted}
_BACKFILL_KEY = 'crawler_minute_backfill'  # {'next': ISO of the next hour to backfill}
_RECLASSIFY_KEY = 'crawler_minute_reclassify'  # {'next': ISO of the next hour, 'end': ISO of the limit}
RUN_DELAY = timedelta(minutes=2)
RECOUNT = timedelta(minutes=30)
# Same reasoning as rollup_edge_hour: fail a window rather than hold locks
# that stall the compression policy; the next run retries it.
_LOCK_TIMEOUT = '30s'
_STATEMENT_TIMEOUT = '10min'
_WORK_MEM = '256MB'

_SQL = """
INSERT INTO honeypot_crawlerminute (minute, bot_type, bot_group, trap_type, count)
SELECT date_trunc('minute', timestamp), bot_type, bot_group, trap_type, count(*)
FROM honeypot_crawlervisit
WHERE timestamp >= %(start)s AND timestamp < %(end)s
GROUP BY 1, 2, 3, 4
"""

_UNCLASSIFIED = "('', 'Other / Browser')"
_SQL_CLASSIFIED = f"""
SELECT date_trunc('minute', timestamp), bot_type, bot_group, trap_type, count(*)
FROM honeypot_crawlervisit
WHERE timestamp >= %(start)s AND timestamp < %(end)s AND bot_type NOT IN {_UNCLASSIFIED}
GROUP BY 1, 2, 3, 4
"""
_SQL_UNCLASSIFIED = f"""
SELECT date_trunc('minute', timestamp), bot_type, bot_group, trap_type, user_agent, count(*)
FROM honeypot_crawlervisit
WHERE timestamp >= %(start)s AND timestamp < %(end)s AND bot_type IN {_UNCLASSIFIED}
GROUP BY 1, 2, 3, 4, 5
"""


def floor_minute(t):
    return t.replace(second=0, microsecond=0)


def _set_limits(cur):
    # SET LOCAL: reset at commit, so safe through pgbouncer.
    cur.execute(f"SET LOCAL lock_timeout = '{_LOCK_TIMEOUT}'")
    cur.execute(f"SET LOCAL statement_timeout = '{_STATEMENT_TIMEOUT}'")
    cur.execute(f"SET LOCAL work_mem = '{_WORK_MEM}'")


def rollup_range(start, end):
    """Replace the rollup for [start, end). Returns the rows written."""
    with transaction.atomic(), connection.cursor() as cur:
        _set_limits(cur)
        cur.execute('DELETE FROM honeypot_crawlerminute WHERE minute >= %s AND minute < %s', [start, end])
        cur.execute(_SQL, {'start': start, 'end': end})
        return cur.rowcount


@lru_cache(maxsize=200_000)
def _label_for(user_agent):
    """(bot_type, bot_group) if user_agent now names a bot, else None."""
    bot_type = classify_ua(user_agent)
    if bot_type in ('Other / Browser', '(empty user agent)'):
        return None
    return bot_type, bot_type_to_group(bot_type)


def reclassify_range(start, end):
    """Replace the rollup for [start, end) with today's labels. Returns the
    rows written. The slow read happens before the write transaction opens."""
    counts = Counter()
    params = {'start': start, 'end': end}
    with transaction.atomic(), connection.cursor() as cur:
        _set_limits(cur)
        cur.execute(_SQL_CLASSIFIED, params)
        for minute, bot_type, bot_group, trap_type, n in cur:
            counts[minute, bot_type, bot_group, trap_type] += n
        cur.execute(_SQL_UNCLASSIFIED, params)
        for minute, bot_type, bot_group, trap_type, user_agent, n in cur:
            bot_type, bot_group = _label_for(user_agent) or (bot_type, bot_group)
            counts[minute, bot_type, bot_group, trap_type] += n
    rows = [
        CrawlerMinute(minute=minute, bot_type=bot_type, bot_group=bot_group, trap_type=trap_type, count=n)
        for (minute, bot_type, bot_group, trap_type), n in counts.items()
    ]
    with transaction.atomic(), connection.cursor() as cur:
        _set_limits(cur)
        cur.execute('DELETE FROM honeypot_crawlerminute WHERE minute >= %s AND minute < %s', [start, end])
        CrawlerMinute.objects.bulk_create(rows, batch_size=5000)
    return len(rows)


def _windows(start, end):
    """[start, end) split at hour boundaries, so one statement never spans
    more than one raw chunk."""
    while start < end:
        stop = min(floor_hour(start) + timedelta(hours=1), end)
        yield start, stop
        start = stop


class Command(BaseCommand):
    help = 'Roll CrawlerVisit up into CrawlerMinute.'

    def add_arguments(self, parser):
        parser.add_argument('--from', dest='from_hour', help='Re-run from this hour (YYYY-MM-DDTHH, UTC); '
                                                             'with --backfill, where a new backfill starts')
        parser.add_argument('--to', dest='to_hour', help='With --from: stop before this hour (default: now)')
        parser.add_argument('--backfill', action='store_true', help='Fill in history older than the rollup')
        parser.add_argument('--reclassify', action='store_true',
                            help='Rebuild --from to --to with current bot labels; resumes from a saved cursor')
        parser.add_argument('--pause', type=float, default=0.5, help='Seconds to wait between hours')
        parser.add_argument('--max-seconds', type=int, default=240, help="Don't start another hour after this long; "
                                                                         '0 for no limit')

    def handle(self, *args, **options):
        batch = options['backfill'] or options['from_hour'] or options['reclassify']
        with open(_BACKFILL_LOCK_FILE if batch else _LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another rollup is already in progress — exiting.')
                return
            if options['reclassify']:
                self._reclassify(options)
            elif options['backfill']:
                self._backfill(options)
            elif options['from_hour']:
                self._rerun(options)
            else:
                self._run(options)

    def _parse_hour(self, s):
        try:
            return datetime.strptime(s, '%Y-%m-%dT%H').replace(tzinfo=dt_timezone.utc)
        except ValueError:
            raise CommandError(f'bad hour {s!r}: use YYYY-MM-DDTHH (UTC)')

    def _deadline(self, options):
        return time.monotonic() + options['max_seconds'] if options['max_seconds'] else float('inf')

    def _run(self, options):
        end = floor_minute(timezone.now() - RUN_DELAY)
        stat = DashboardStat.objects.filter(key=_WATERMARK_KEY).first()
        if stat and stat.value.get('minute'):
            mark = datetime.fromisoformat(stat.value['minute'])
            start = mark - RECOUNT
        else:
            mark = None
            start = end - RECOUNT
        deadline = self._deadline(options)
        rows = 0
        for a, b in _windows(start, end):
            if time.monotonic() >= deadline:
                break
            rows += rollup_range(a, b)
            if mark is None or b > mark:
                mark = b
                DashboardStat.objects.update_or_create(key=_WATERMARK_KEY, defaults={'value': {'minute': b.isoformat()}})
        self.stdout.write(f'Rolled up to {mark:%Y-%m-%dT%H:%M}: {rows} row(s).')

    def _rerun(self, options):
        start = self._parse_hour(options['from_hour'])
        end = self._parse_hour(options['to_hour']) if options['to_hour'] else floor_minute(timezone.now() - RUN_DELAY)
        rows = 0
        for a, b in _windows(start, end):
            rows += rollup_range(a, b)
            time.sleep(options['pause'])
        self.stdout.write(f'Re-ran from {start:%Y-%m-%dT%H}: {rows} row(s).')

    def _backfill(self, options):
        end = CrawlerMinute.objects.aggregate(m=Min('minute'))['m']
        stat = DashboardStat.objects.filter(key=_BACKFILL_KEY).first()
        if stat and stat.value.get('next'):
            nxt = datetime.fromisoformat(stat.value['next'])
            # The saved position is the backfill's upper limit once it has
            # started, because backfilled rows move the oldest minute back.
            end = datetime.fromisoformat(stat.value['end'])
        elif not options['from_hour']:
            raise CommandError('--backfill needs --from to start')
        elif end is None:
            raise CommandError('nothing rolled up yet: let the cron run once first')
        else:
            nxt = self._parse_hour(options['from_hour'])
        deadline = self._deadline(options)
        done = rows = 0
        for a, b in _windows(nxt, end):
            if time.monotonic() >= deadline:
                break
            t0 = time.monotonic()
            n = rollup_range(a, b)
            rows += n
            done += 1
            DashboardStat.objects.update_or_create(
                key=_BACKFILL_KEY, defaults={'value': {'next': b.isoformat(), 'end': end.isoformat()}})
            self.stdout.write(f'{a:%Y-%m-%dT%H:%M}: {n} rows ({time.monotonic() - t0:.1f}s)')
            time.sleep(options['pause'])
        self.stdout.write(f'Backfilled {done} hour(s), {rows} row(s).')

    def _reclassify(self, options):
        stat = DashboardStat.objects.filter(key=_RECLASSIFY_KEY).first()
        if stat and stat.value.get('next'):
            nxt = datetime.fromisoformat(stat.value['next'])
            end = datetime.fromisoformat(stat.value['end'])
        elif not (options['from_hour'] and options['to_hour']):
            raise CommandError('--reclassify needs --from and --to to start')
        else:
            nxt = self._parse_hour(options['from_hour'])
            end = self._parse_hour(options['to_hour'])
        DashboardStat.objects.filter(key=_BACKFILL_KEY).delete()
        deadline = self._deadline(options)
        done = rows = 0
        for a, b in _windows(nxt, end):
            if time.monotonic() >= deadline:
                break
            t0 = time.monotonic()
            n = reclassify_range(a, b)
            rows += n
            done += 1
            DashboardStat.objects.update_or_create(
                key=_RECLASSIFY_KEY, defaults={'value': {'next': b.isoformat(), 'end': end.isoformat()}})
            self.stdout.write(f'{a:%Y-%m-%dT%H:%M}: {n} rows ({time.monotonic() - t0:.1f}s)')
            self.stdout.flush()
            time.sleep(options['pause'])
        self.stdout.write(f'Reclassified {done} hour(s), {rows} row(s).')
