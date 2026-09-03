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
def test_precalc_dashboard_crawlers_daily_preserves_old_days_outside_recompute_window():
    """Regression test: the daily chart used to do a full 60-day GROUP BY on
    every run, which forces TimescaleDB to decompress every compressed chunk
    (compression is segmented by bot_type/trap_type, not date) — at this
    project's volume that took 20+ minutes and pinned a PgBouncer connection.
    It now only recomputes a recent trailing window and merges the result
    into the stored dict, so a day outside that window must survive untouched
    even though no CrawlerVisit rows exist for it any more in a live query."""
    from apps.core.management.commands.precalc_dashboard import _DAILY_RECOMPUTE_WINDOW_DAYS

    stale_day = (timezone.now() - timedelta(days=_DAILY_RECOMPUTE_WINDOW_DAYS + 5)).date().isoformat()
    DashboardStat.objects.update_or_create(
        key='crawlers.daily', defaults={'value': {stale_day: 999}}
    )
    DashboardStat.objects.update_or_create(
        key='crawlers.daily_by_bot_type', defaults={'value': {stale_day: {'Googlebot': 999}}}
    )

    CrawlerVisit.objects.create(
        timestamp=timezone.now(), ip_address='1.2.3.4', path='/a', trap_type='archive', bot_type='Googlebot',
    )

    call_command('precalc_dashboard', stdout=io.StringIO())

    daily = DashboardStat.objects.get(key='crawlers.daily').value
    assert daily.get(stale_day) == 999
    daily_by_bot = DashboardStat.objects.get(key='crawlers.daily_by_bot_type').value
    assert daily_by_bot.get(stale_day) == {'Googlebot': 999}


@pytest.mark.django_db
def test_precalc_dashboard_archive_catches_up_across_multiple_runs():
    base = timezone.now() - timedelta(hours=5)
    ArchiveVisit.objects.bulk_create([
        ArchiveVisit(timestamp=base + timedelta(minutes=i * 5), ip_address=f'9.9.9.{i}', year=2024, month=1, day=1, slug=f's{i}', depth=i)
        for i in range(5)
    ])

    call_command('precalc_dashboard', stdout=io.StringIO())

    assert DashboardStat.objects.get(key='archive.total').value == 5
