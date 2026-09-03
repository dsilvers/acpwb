import io
from datetime import timedelta

import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.ip_intel_classify import classify_hosting
from apps.core import tor_exit_list
from apps.honeypot.models import CrawlerVisit, IPIntelligence


@pytest.mark.parametrize("asn_org,expected", [
    ('AMAZON-02', True),
    ('Google LLC', True),
    ('Hetzner Online GmbH', True),
    ('OVH SAS', True),
    ('DIGITALOCEAN-ASN', True),
    ('Telefonica Germany', False),
    ('Comcast Cable Communications, LLC', False),
    ('', False),
    (None, False),
])
def test_classify_hosting(asn_org, expected):
    assert classify_hosting(asn_org) is expected


def test_is_tor_exit(tmp_path, settings):
    list_path = tmp_path / 'tor_exit_nodes.txt'
    list_path.write_text('1.2.3.4\n5.6.7.8\n')
    settings.TOR_EXIT_LIST_PATH = str(list_path)
    tor_exit_list._cache['mtime'] = None  # force reload for this test's file

    assert tor_exit_list.is_tor_exit('1.2.3.4') is True
    assert tor_exit_list.is_tor_exit('9.9.9.9') is False


@pytest.mark.django_db
def test_discover_ip_intelligence_aggregates_and_is_resumable():
    base = timezone.now() - timedelta(hours=5)
    CrawlerVisit.objects.bulk_create([
        CrawlerVisit(timestamp=base, ip_address='1.2.3.4', path='/a', trap_type='archive'),
        CrawlerVisit(timestamp=base + timedelta(minutes=10), ip_address='1.2.3.4', path='/b', trap_type='archive'),
        CrawlerVisit(timestamp=base + timedelta(minutes=20), ip_address='5.6.7.8', path='/c', trap_type='policy'),
        CrawlerVisit(timestamp=base + timedelta(hours=2), ip_address='1.2.3.4', path='/d', trap_type='archive'),
    ])

    call_command('discover_ip_intelligence', step_hours=1, max_seconds=0, stdout=io.StringIO())

    by_ip = {r.ip_address: r for r in IPIntelligence.objects.all()}
    assert by_ip['1.2.3.4'].visit_count == 3
    assert by_ip['5.6.7.8'].visit_count == 1
    assert by_ip['1.2.3.4'].first_seen == base
    assert by_ip['1.2.3.4'].last_seen == base + timedelta(hours=2)
    assert by_ip['1.2.3.4'].enriched_at is None

    # A second run with nothing new to discover should be a no-op (watermark
    # already caught up) rather than double-counting.
    call_command('discover_ip_intelligence', step_hours=1, max_seconds=0, stdout=io.StringIO())
    by_ip = {r.ip_address: r for r in IPIntelligence.objects.all()}
    assert by_ip['1.2.3.4'].visit_count == 3


@pytest.mark.django_db
def test_discover_ip_intelligence_first_run_caps_lookback_by_default():
    """On a huge table, the first-ever run must not walk back to the
    earliest row (that's the query shape behind the connection-exhaustion
    incident in deploy/README.md) — it should cap to --max-lookback-days."""
    old = timezone.now() - timedelta(days=90)
    recent = timezone.now() - timedelta(hours=1)
    CrawlerVisit.objects.bulk_create([
        CrawlerVisit(timestamp=old, ip_address='10.0.0.1', path='/old', trap_type='archive'),
        CrawlerVisit(timestamp=recent, ip_address='10.0.0.2', path='/new', trap_type='archive'),
    ])

    call_command('discover_ip_intelligence', step_hours=24, max_seconds=0, max_lookback_days=14, stdout=io.StringIO())

    ips = set(IPIntelligence.objects.values_list('ip_address', flat=True))
    assert '10.0.0.2' in ips
    assert '10.0.0.1' not in ips  # older than the 14-day cap, correctly skipped

    call_command('discover_ip_intelligence', step_hours=24, max_seconds=0, full_history=True, stdout=io.StringIO())
    # full_history only affects seeding the very first watermark, which is
    # already stored now — it stays skipped without a --since reset.
    ips = set(IPIntelligence.objects.values_list('ip_address', flat=True))
    assert '10.0.0.1' not in ips


@pytest.mark.django_db
def test_discover_ip_intelligence_since_does_not_roll_back_existing_watermark():
    """Regression test: repeating --since on every invocation (a real
    footgun hit in production) must NOT keep resetting progress back to
    that date once a watermark is already stored — only --force may do that."""
    base = timezone.now() - timedelta(hours=10)
    CrawlerVisit.objects.bulk_create([
        CrawlerVisit(timestamp=base, ip_address='1.1.1.1', path='/a', trap_type='archive'),
        CrawlerVisit(timestamp=base + timedelta(hours=5), ip_address='2.2.2.2', path='/b', trap_type='archive'),
    ])

    out = io.StringIO()
    call_command('discover_ip_intelligence', step_hours=1, max_seconds=0, since=base.date().isoformat(), stdout=out)
    from apps.core.models import DashboardStat
    watermark_after_first = DashboardStat.objects.get(key='ip_intel_discover_watermark').value['ts']

    # Re-running with the SAME --since must resume, not reset, now that a
    # watermark is stored (it may still creep forward slightly on its own —
    # real time passes between calls and the watermark tracks "now" once
    # caught up — the point is it must never go BACKWARD to --since's date).
    out2 = io.StringIO()
    call_command('discover_ip_intelligence', step_hours=1, max_seconds=0, since=base.date().isoformat(), stdout=out2)
    assert 'ignored' in out2.getvalue()
    watermark_after_second = DashboardStat.objects.get(key='ip_intel_discover_watermark').value['ts']
    assert watermark_after_second >= watermark_after_first

    # --force must still allow a deliberate rollback.
    out3 = io.StringIO()
    call_command('discover_ip_intelligence', step_hours=1, max_seconds=0, since=base.date().isoformat(), force=True, stdout=out3)
    assert 'ignored' not in out3.getvalue()
