import io
from datetime import timedelta
from unittest.mock import patch

import pytest
from django.core.management import CommandError, call_command
from django.utils import timezone

from apps.core.management.commands.rollup_crawler_minute import (
    RECOUNT, RUN_DELAY, _BACKFILL_KEY, _RECLASSIFY_KEY, _WATERMARK_KEY, _windows, floor_minute, reclassify_range,
    rollup_range,
)
from apps.core.management.commands.rollup_edge_hour import floor_hour
from apps.core.models import DashboardStat
from apps.honeypot.models import CrawlerMinute, CrawlerVisit


AION = 'Mozilla/5.0 (X11; Linux x86_64) Chrome/149.0.0.0 Safari/537.36 (compatible; AionBot/1.0)'
CHROME = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'


def _visit(t, bot='GPTBot', group='AI Crawlers', trap='archive', ua=''):
    return CrawlerVisit(timestamp=t, ip_address='203.0.113.1', path='/', trap_type=trap, bot_type=bot, bot_group=group,
                        user_agent=ua)


def _unclassified(t, ua, bot='Other / Browser', trap='archive'):
    return _visit(t, bot=bot, group=bot, trap=trap, ua=ua)


def _counts():
    return {(r.minute, r.bot_type, r.trap_type): r.count for r in CrawlerMinute.objects.all()}


def _run(**kw):
    out = io.StringIO()
    call_command('rollup_crawler_minute', stdout=out, pause=0, **kw)
    return out.getvalue()


@pytest.mark.django_db
class TestRollupRange:
    def test_counts_per_minute_bot_and_trap(self):
        t = floor_minute(timezone.now()) - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([
            _visit(t), _visit(t + timedelta(seconds=59)), _visit(t, trap='policy'),
            _visit(t, bot='Other / Browser', group='Other / Browser'), _visit(t + timedelta(minutes=1)),
        ])
        assert rollup_range(t, t + timedelta(minutes=5)) == 4
        assert _counts() == {
            (t, 'GPTBot', 'archive'): 2, (t, 'GPTBot', 'policy'): 1, (t, 'Other / Browser', 'archive'): 1,
            (t + timedelta(minutes=1), 'GPTBot', 'archive'): 1,
        }

    def test_blank_bot_is_kept(self):
        t = floor_minute(timezone.now()) - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([_visit(t, bot='', group='')])
        rollup_range(t, t + timedelta(minutes=1))
        assert _counts() == {(t, '', 'archive'): 1}

    def test_rerun_replaces_rather_than_adds(self):
        t = floor_minute(timezone.now()) - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([_visit(t)])
        rollup_range(t, t + timedelta(minutes=1))
        CrawlerVisit.objects.bulk_create([_visit(t)])
        rollup_range(t, t + timedelta(minutes=1))
        assert _counts() == {(t, 'GPTBot', 'archive'): 2}

    def test_end_is_exclusive(self):
        t = floor_minute(timezone.now()) - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([_visit(t), _visit(t + timedelta(minutes=1))])
        rollup_range(t, t + timedelta(minutes=1))
        assert list(_counts()) == [(t, 'GPTBot', 'archive')]


class TestWindows:
    def test_split_at_hours(self):
        t = floor_hour(timezone.now()) - timedelta(hours=3)
        a, b = t + timedelta(minutes=50), t + timedelta(hours=2, minutes=10)
        assert list(_windows(a, b)) == [
            (a, t + timedelta(hours=1)), (t + timedelta(hours=1), t + timedelta(hours=2)), (t + timedelta(hours=2), b),
        ]


@pytest.mark.django_db
class TestCron:
    def test_first_run_starts_at_recount_and_sets_watermark(self):
        end = floor_minute(timezone.now() - RUN_DELAY)
        CrawlerVisit.objects.bulk_create([
            _visit(end - timedelta(minutes=1)), _visit(end - RECOUNT - timedelta(minutes=5)),
        ])
        _run()
        assert list(_counts()) == [(end - timedelta(minutes=1), 'GPTBot', 'archive')]
        mark = DashboardStat.objects.get(key=_WATERMARK_KEY).value['minute']
        assert mark >= end.isoformat()

    def test_late_visit_is_picked_up(self):
        end = floor_minute(timezone.now() - RUN_DELAY)
        t = end - timedelta(minutes=5)
        CrawlerVisit.objects.bulk_create([_visit(t)])
        _run()
        CrawlerVisit.objects.bulk_create([_visit(t)])  # the drain caught up
        _run()
        assert _counts() == {(t, 'GPTBot', 'archive'): 2}

    def test_catches_up_after_an_outage(self):
        end = floor_minute(timezone.now() - RUN_DELAY)
        t = end - timedelta(hours=5)
        DashboardStat.objects.create(key=_WATERMARK_KEY, value={'minute': (t - timedelta(minutes=1)).isoformat()})
        CrawlerVisit.objects.bulk_create([_visit(t)])
        _run()
        assert _counts() == {(t, 'GPTBot', 'archive'): 1}

    def test_deadline_keeps_the_watermark_where_it_stopped(self):
        end = floor_minute(timezone.now() - RUN_DELAY)
        before = end - timedelta(hours=5)
        DashboardStat.objects.create(key=_WATERMARK_KEY, value={'minute': before.isoformat()})
        with patch('apps.core.management.commands.rollup_crawler_minute.time') as fake:
            fake.monotonic.side_effect = [0, 0, 1000]
            _run(max_seconds=10)
        mark = DashboardStat.objects.get(key=_WATERMARK_KEY).value['minute']
        assert before.isoformat() <= mark < end.isoformat()


@pytest.mark.django_db
class TestBackfill:
    def _rolled(self):
        t = floor_hour(timezone.now()) - timedelta(hours=1)
        CrawlerMinute.objects.create(minute=t + timedelta(minutes=30), bot_type='GPTBot', trap_type='archive', count=1)
        return t

    def test_fills_up_to_oldest_rolled_minute(self):
        t = self._rolled()
        old = t - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([_visit(old), _visit(t + timedelta(minutes=10))])
        _run(backfill=True, from_hour=f'{old:%Y-%m-%dT%H}', max_seconds=0)
        assert _counts()[(old, 'GPTBot', 'archive')] == 1
        assert _counts()[(t + timedelta(minutes=10), 'GPTBot', 'archive')] == 1
        assert _counts()[(t + timedelta(minutes=30), 'GPTBot', 'archive')] == 1  # left alone

    def test_resumes_from_cursor_with_original_limit(self):
        t = self._rolled()
        old = t - timedelta(hours=2)
        CrawlerVisit.objects.bulk_create([_visit(old), _visit(old + timedelta(hours=1))])
        end = (t + timedelta(minutes=30)).isoformat()
        DashboardStat.objects.create(key=_BACKFILL_KEY, value={'next': (old + timedelta(hours=1)).isoformat(), 'end': end})
        _run(backfill=True)
        assert (old, 'GPTBot', 'archive') not in _counts()
        assert (old + timedelta(hours=1), 'GPTBot', 'archive') in _counts()

    def test_second_run_after_finishing_does_nothing(self):
        t = self._rolled()
        _run(backfill=True, from_hour=f'{t:%Y-%m-%dT%H}', max_seconds=0)
        assert 'Backfilled 0 hour' in _run(backfill=True, max_seconds=0)

    def test_needs_a_start(self):
        with pytest.raises(CommandError):
            _run(backfill=True)

    def test_needs_the_cron_to_have_run(self):
        with pytest.raises(CommandError):
            _run(backfill=True, from_hour='2026-07-30T00')


@pytest.mark.django_db
def test_rerun_range_picks_up_reclassification():
    t = floor_hour(timezone.now()) - timedelta(hours=3)
    visit = _visit(t, bot='Other / Browser', group='Other / Browser')
    CrawlerVisit.objects.bulk_create([visit])
    rollup_range(t, t + timedelta(hours=1))
    CrawlerVisit.objects.filter(timestamp=t).update(bot_type='AionBot', bot_group='SEO / Other Bots')
    _run(from_hour=f'{t:%Y-%m-%dT%H}', to_hour=f'{t + timedelta(hours=1):%Y-%m-%dT%H}')
    assert _counts() == {(t, 'AionBot', 'archive'): 1}


@pytest.mark.django_db
class TestReclassify:
    def _hour(self):
        return floor_hour(timezone.now()) - timedelta(hours=3)

    def test_relabels_rows_that_now_match_and_merges_counts(self):
        t = self._hour()
        CrawlerVisit.objects.bulk_create([_unclassified(t, AION), _unclassified(t, AION), _unclassified(t, CHROME)])
        assert reclassify_range(t, t + timedelta(hours=1)) == 2
        assert _counts() == {(t, 'AionBot', 'archive'): 2, (t, 'Other / Browser', 'archive'): 1}
        assert CrawlerMinute.objects.get(bot_type='AionBot').bot_group == 'SEO / Other Bots'

    def test_merges_into_a_label_already_stored_that_minute(self):
        t = self._hour()
        CrawlerVisit.objects.bulk_create([
            _visit(t, bot='AionBot', group='SEO / Other Bots', ua=AION), _unclassified(t, AION),
        ])
        reclassify_range(t, t + timedelta(hours=1))
        assert _counts() == {(t, 'AionBot', 'archive'): 2}

    def test_stored_labels_are_kept(self):
        # An IP-range label has a browser UA, which must not undo it.
        t = self._hour()
        CrawlerVisit.objects.bulk_create([_visit(t, bot='Tencent', group='AI Crawlers', ua=CHROME)])
        reclassify_range(t, t + timedelta(hours=1))
        assert _counts() == {(t, 'Tencent', 'archive'): 1}

    def test_blank_bot_is_kept_unless_the_ua_matches(self):
        t = self._hour()
        CrawlerVisit.objects.bulk_create([_unclassified(t, CHROME, bot=''), _unclassified(t, AION, bot='')])
        reclassify_range(t, t + timedelta(hours=1))
        assert _counts() == {(t, '', 'archive'): 1, (t, 'AionBot', 'archive'): 1}

    def test_does_not_touch_raw_rows(self):
        t = self._hour()
        CrawlerVisit.objects.bulk_create([_unclassified(t, AION)])
        reclassify_range(t, t + timedelta(hours=1))
        assert CrawlerVisit.objects.get().bot_type == 'Other / Browser'

    def test_replaces_the_range_and_leaves_the_rest(self):
        t = self._hour()
        CrawlerVisit.objects.bulk_create([_unclassified(t, AION), _unclassified(t + timedelta(hours=1), AION)])
        rollup_range(t, t + timedelta(hours=2))
        reclassify_range(t, t + timedelta(hours=1))
        assert _counts() == {
            (t, 'AionBot', 'archive'): 1, (t + timedelta(hours=1), 'Other / Browser', 'archive'): 1,
        }

    def test_command_runs_saves_cursor_and_clears_backfill(self):
        t = self._hour()
        end = t + timedelta(hours=2)
        DashboardStat.objects.create(key=_BACKFILL_KEY, value={'next': t.isoformat(), 'end': t.isoformat()})
        CrawlerVisit.objects.bulk_create([_unclassified(t, AION), _unclassified(end, AION)])
        out = _run(reclassify=True, from_hour=f'{t:%Y-%m-%dT%H}', to_hour=f'{end:%Y-%m-%dT%H}', max_seconds=0)
        assert 'Reclassified 2 hour' in out
        assert _counts() == {(t, 'AionBot', 'archive'): 1}
        assert DashboardStat.objects.get(key=_RECLASSIFY_KEY).value['next'] == end.isoformat()
        assert not DashboardStat.objects.filter(key=_BACKFILL_KEY).exists()

    def test_resumes_from_cursor_over_from(self):
        t = self._hour()
        end = t + timedelta(hours=2)
        DashboardStat.objects.create(key=_RECLASSIFY_KEY, value={'next': (t + timedelta(hours=1)).isoformat(), 'end': end.isoformat()})
        CrawlerVisit.objects.bulk_create([_unclassified(t, AION), _unclassified(t + timedelta(hours=1), AION)])
        _run(reclassify=True, from_hour=f'{t:%Y-%m-%dT%H}', to_hour=f'{end:%Y-%m-%dT%H}', max_seconds=0)
        assert _counts() == {(t + timedelta(hours=1), 'AionBot', 'archive'): 1}

    def test_finished_run_does_nothing(self):
        t = self._hour()
        DashboardStat.objects.create(key=_RECLASSIFY_KEY, value={'next': t.isoformat(), 'end': t.isoformat()})
        assert 'Reclassified 0 hour' in _run(reclassify=True, max_seconds=0)

    def test_needs_a_range_to_start(self):
        with pytest.raises(CommandError):
            _run(reclassify=True, from_hour='2026-07-30T00')


def test_bad_hour_is_rejected():
    with pytest.raises(CommandError):
        _run(from_hour='yesterday')
