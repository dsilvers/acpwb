import io
from datetime import timedelta

import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.models import DashboardStat
from apps.honeypot.models import ArchiveVisit, CrawlerVisit


@pytest.mark.django_db
def test_precalc_dashboard_crawlers_catches_up_across_multiple_runs():
    """Regression test for the OFFSET(500_000) -> time-window fix.

    Seeds two clusters of CrawlerVisit rows more than _MAX_HOURS_PER_RUN
    apart, so a single run can't process everything — it must take several
    runs to fully catch up, without ever getting stuck re-scanning an empty
    window forever.
    """
    base = timezone.now() - timedelta(hours=5)
    early = [
        CrawlerVisit(timestamp=base + timedelta(minutes=i * 10), ip_address=f'1.2.3.{i}', path=f'/a{i}', trap_type='archive')
        for i in range(5)
    ]
    late = [
        CrawlerVisit(timestamp=base + timedelta(hours=3, minutes=i * 10), ip_address=f'5.6.7.{i}', path=f'/b{i}', trap_type='policy')
        for i in range(5)
    ]
    CrawlerVisit.objects.bulk_create(early + late)

    for _ in range(4):
        call_command('precalc_dashboard', stdout=io.StringIO())

    assert DashboardStat.objects.get(key='crawlers.total').value == 10
    hwm = DashboardStat.objects.get(key='hwm.crawler_visit_ts').value
    assert hwm  # advanced all the way to "now" once caught up

    # Further runs must not re-count already-processed rows.
    call_command('precalc_dashboard', stdout=io.StringIO())
    assert DashboardStat.objects.get(key='crawlers.total').value == 10


@pytest.mark.django_db
def test_precalc_dashboard_crawlers_daily_increments_without_rescanning_history():
    """Regression test: the daily chart used to be kept fresh by re-querying
    a date-range GROUP BY over the whole chart window (first 60 days, then a
    narrowed "recent" window) on every run — both approaches forced scanning/
    decompressing huge amounts of this table's 90M+ rows/day and took 20+
    minutes, pinning a PgBouncer connection. It's now incremented from the
    same small new_rows window already used for by_trap_type/by_bot_type, so:
    a day within retention that today's run doesn't touch must survive
    unchanged (no re-query of history), new rows must increment rather than
    overwrite same-day counts, and a day past retention must get pruned."""
    from apps.core.management.commands.precalc_dashboard import _DAILY_RETENTION_DAYS

    untouched_day = (timezone.now() - timedelta(days=3)).date().isoformat()
    today = timezone.now().date().isoformat()
    stale_day = (timezone.now() - timedelta(days=_DAILY_RETENTION_DAYS['crawlers'] + 5)).date().isoformat()
    DashboardStat.objects.update_or_create(
        key='crawlers.daily', defaults={'value': {untouched_day: 5, today: 2, stale_day: 999}}
    )
    DashboardStat.objects.update_or_create(
        key='crawlers.daily_by_bot_type',
        defaults={'value': {untouched_day: {'Googlebot': 5}, today: {'Googlebot': 2}, stale_day: {'Googlebot': 999}}},
    )

    CrawlerVisit.objects.create(
        timestamp=timezone.now(), ip_address='1.2.3.4', path='/a', trap_type='archive', bot_type='Googlebot',
    )

    call_command('precalc_dashboard', stdout=io.StringIO())

    daily = DashboardStat.objects.get(key='crawlers.daily').value
    assert daily.get(untouched_day) == 5  # not re-queried, survives untouched
    assert daily.get(today) == 3  # incremented, not overwritten
    assert stale_day not in daily  # past retention, pruned

    daily_by_bot = DashboardStat.objects.get(key='crawlers.daily_by_bot_type').value
    assert daily_by_bot.get(untouched_day) == {'Googlebot': 5}
    assert daily_by_bot.get(today) == {'Googlebot': 3}
    assert stale_day not in daily_by_bot


@pytest.mark.django_db
def test_precalc_dashboard_archive_catches_up_across_multiple_runs():
    base = timezone.now() - timedelta(hours=5)
    ArchiveVisit.objects.bulk_create([
        ArchiveVisit(timestamp=base + timedelta(minutes=i * 5), ip_address=f'9.9.9.{i}', year=2024, month=1, day=1, slug=f's{i}', depth=i)
        for i in range(5)
    ])

    call_command('precalc_dashboard', stdout=io.StringIO())

    assert DashboardStat.objects.get(key='archive.total').value == 5
