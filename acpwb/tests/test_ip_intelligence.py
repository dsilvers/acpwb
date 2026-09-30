import io
from datetime import timedelta
from unittest.mock import patch

import geoip2.errors
import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.ip_intel_classify import classify_hosting
from apps.core import tor_exit_list
from apps.honeypot.models import CrawlerVisit, IPIntelligence


class _FakeGeoReader:
    """Stands in for geoip2.database.Reader — every lookup is a miss, which
    is all the enrich_ip_intelligence pagination test needs to exercise."""
    def city(self, ip):
        raise geoip2.errors.AddressNotFoundError('not found')

    def asn(self, ip):
        raise geoip2.errors.AddressNotFoundError('not found')

    def metadata(self):
        class _Meta:
            build_epoch = 1735689600  # 2025-01-01, arbitrary fixed date
        return _Meta()

    def close(self):
        pass


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


@pytest.mark.django_db
def test_enrich_ip_intelligence_pages_through_all_rows_without_iterator():
    """Regression test: enrich_ip_intelligence used to stream via
    qs.iterator(), which breaks under PgBouncer transaction-pooling mode
    ("cursor does not exist" — a different backend can serve a later FETCH).
    It now pages via keyset pagination on IPIntelligence's real PK instead.
    """
    for i in range(25):
        IPIntelligence.objects.create(ip_address=f'10.0.{i}.1')

    with patch('geoip2.database.Reader', return_value=_FakeGeoReader()):
        call_command('enrich_ip_intelligence', batch_size=7, stdout=io.StringIO())

    assert IPIntelligence.objects.filter(enriched_at__isnull=True).count() == 0
    assert IPIntelligence.objects.filter(lookup_ok=False).count() == 25


class _HitGeoReader(_FakeGeoReader):
    """Every lookup succeeds — exercises the write path with real values,
    including a NULL latitude/longitude mixed in with non-NULL ones."""
    def city(self, ip):
        from types import SimpleNamespace as NS
        located = not ip.endswith('.9')
        return NS(
            country=NS(iso_code='BR', name='Brazil'),
            subdivisions=NS(most_specific=NS(name='São Paulo')),
            city=NS(name='Campinas'),
            location=NS(latitude=-22.9 if located else None,
                        longitude=-47.06 if located else None,
                        accuracy_radius=20 if located else None),
        )

    def asn(self, ip):
        from types import SimpleNamespace as NS
        return NS(autonomous_system_number=16509, autonomous_system_organization='AMAZON-02')


@pytest.mark.django_db
def test_enrich_ip_intelligence_writes_lookup_results():
    # 12 rows with batch_size=5 → several flushes; row .9 has NULL coordinates.
    for i in range(12):
        IPIntelligence.objects.create(ip_address=f'10.2.0.{i}')

    with patch('geoip2.database.Reader', return_value=_HitGeoReader()):
        call_command('enrich_ip_intelligence', batch_size=5, stdout=io.StringIO())

    assert IPIntelligence.objects.filter(enriched_at__isnull=True).count() == 0
    row = IPIntelligence.objects.get(ip_address='10.2.0.1')
    assert (row.country_code, row.country_name, row.region_name, row.city_name) == ('BR', 'Brazil', 'São Paulo', 'Campinas')
    assert row.latitude == pytest.approx(-22.9) and row.accuracy_radius_km == 20
    assert (row.asn, row.asn_org, row.is_hosting, row.lookup_ok) == (16509, 'AMAZON-02', True, True)
    assert row.geoip_db_date is not None and row.enrichment_note == ''
    unlocated = IPIntelligence.objects.get(ip_address='10.2.0.9')
    assert unlocated.latitude is None and unlocated.country_code == 'BR'


@pytest.mark.django_db
def test_enrich_ip_intelligence_respects_limit():
    for i in range(10):
        IPIntelligence.objects.create(ip_address=f'10.1.{i}.1')

    with patch('geoip2.database.Reader', return_value=_FakeGeoReader()):
        call_command('enrich_ip_intelligence', batch_size=3, limit=5, stdout=io.StringIO())

    assert IPIntelligence.objects.filter(enriched_at__isnull=False).count() == 5
