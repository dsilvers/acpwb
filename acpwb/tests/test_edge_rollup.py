import io
import uuid
from datetime import timedelta

import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.management.commands.rollup_edge_hour import floor_hour, rollup_hour
from apps.core.models import DashboardStat
from apps.honeypot.models import CrawlerVisit, EdgeConnection, EdgeFingerprintHour, EdgeIPHour, EdgeRequest

@pytest.fixture
def staff_client(client, db):
    from django.contrib.auth.models import User
    client.force_login(User.objects.create_user('staff', password='pass', is_staff=True))
    return client


FAKE_JA4 = 't13d311200_1d947a95fc68_1aa0e0f8a5b3'
CHROME_JA4 = 't13d1516h2_8daaf6152771_02713d6af862'
CHROME_H2 = '1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p'


def _hour(hours_ago=2):
    return floor_hour(timezone.now()) - timedelta(hours=hours_ago)


def _traffic(start):
    """A residential proxy pool (three IPs, one fake-Chrome connection
    each), and one real browser on a single connection with a static file
    and a returned cookie."""
    t = start + timedelta(minutes=5)
    conns, reqs = [], []
    for i in range(3):
        cid = uuid.uuid4()
        ip = f'203.0.113.{i + 1}'
        conns.append(EdgeConnection(timestamp=t, ip_address=ip, outcome='ok', ja4=FAKE_JA4, conn_id=cid,
                                    request_count=1, network='residential', asn=28573,
                                    first_user_agent='Mozilla/5.0 Chrome/131', idempotency_key=str(cid)))
        reqs.append(EdgeRequest(timestamp=t, request_id=uuid.uuid4(), conn_id=cid, ip_address=ip,
                                route='django', status=200, client_class='impersonator', network='residential',
                                asn=28573, ja4=FAKE_JA4, stack='t13d311200', signals=0b11, new_client=True,
                                user_agent='Mozilla/5.0 Chrome/131'))
    cid = uuid.uuid4()
    conns.append(EdgeConnection(timestamp=t, ip_address='198.51.100.9', outcome='ok', ja4=CHROME_JA4,
                                h2_fingerprint=CHROME_H2, conn_id=cid, request_count=2, network='residential',
                                first_user_agent='Mozilla/5.0 Chrome/131', idempotency_key=str(cid)))
    for route, new, client in (('django', True, 'c' * 22), ('static', False, 'c' * 22)):
        reqs.append(EdgeRequest(timestamp=t, request_id=uuid.uuid4(), conn_id=cid, ip_address='198.51.100.9',
                                route=route, status=200, client_class='browser', network='residential',
                                ja4=CHROME_JA4, stack='Chromium', new_client=new, client_id=client))
    # Outside the hour: ignored.
    reqs.append(EdgeRequest(timestamp=start + timedelta(hours=1, minutes=1), request_id=uuid.uuid4(),
                            ip_address='203.0.113.1', client_class='impersonator'))
    EdgeConnection.objects.bulk_create(conns)
    EdgeRequest.objects.bulk_create(reqs)
    CrawlerVisit.objects.create(ip_address='203.0.113.1', trap_type='ghost_link', path='/x', timestamp=t)


@pytest.mark.django_db
def test_rollup_hour_per_ip_and_fingerprint():
    start = _hour()
    _traffic(start)
    assert rollup_hour(start) == (4, 2)

    pool = EdgeIPHour.objects.get(hour=start, ip_address='203.0.113.1')
    assert (pool.requests, pool.conns, pool.impersonator_reqs, pool.browser_reqs) == (1, 1, 1, 0)
    assert (pool.network, pool.asn, pool.top_stack, pool.signals) == ('residential', 28573, 't13d311200', 0b11)
    assert (pool.new_client_reqs, pool.cookie_reqs, pool.trap_hits) == (1, 0, 1)

    browser = EdgeIPHour.objects.get(hour=start, ip_address='198.51.100.9')
    assert (browser.requests, browser.browser_reqs, browser.static_reqs, browser.cookie_reqs) == (2, 2, 1, 1)
    assert browser.trap_hits == 0

    fake = EdgeFingerprintHour.objects.get(hour=start, ja4=FAKE_JA4)
    assert (fake.conns, fake.ips, fake.residential_ips, fake.single_conn_ips) == (3, 3, 3, 3)
    assert (fake.impersonator_reqs, fake.browser_reqs, fake.top_stack) == (3, 0, 't13d311200')
    assert fake.top_user_agent == 'Mozilla/5.0 Chrome/131'
    real = EdgeFingerprintHour.objects.get(hour=start, ja4=CHROME_JA4)
    assert (real.h2_fingerprint, real.requests, real.browser_reqs, real.single_conn_ips) == (CHROME_H2, 2, 2, 1)

    # Re-running an hour replaces it.
    assert rollup_hour(start) == (4, 2)
    assert EdgeIPHour.objects.filter(hour=start).count() == 4


@pytest.mark.django_db
def test_rollup_command_follows_its_watermark():
    start = _hour(3)
    _traffic(start)
    call_command('rollup_edge_hour', since=start.strftime('%Y-%m-%dT%H'), stdout=io.StringIO())
    assert EdgeIPHour.objects.filter(hour=start).count() == 4
    mark = DashboardStat.objects.get(key='edge_rollup_watermark').value['hour']
    assert mark >= (start + timedelta(hours=1)).isoformat()

    # Caught up: the next run starts after the watermark and leaves the
    # rolled-up hour alone.
    EdgeIPHour.objects.filter(hour=start).delete()
    out = io.StringIO()
    call_command('rollup_edge_hour', stdout=out)
    assert not EdgeIPHour.objects.filter(hour=start).exists()

    call_command('rollup_edge_hour', hour=start.strftime('%Y-%m-%dT%H'), stdout=io.StringIO())
    assert EdgeIPHour.objects.filter(hour=start).count() == 4


@pytest.mark.django_db(databases=['default', 'direct'], transaction=True)
def test_precalc_edge_classes_and_dashboard(staff_client):
    start = _hour()
    _traffic(start)
    rollup_hour(start)
    call_command('precalc_dashboard', stdout=io.StringIO())

    s = DashboardStat.objects.get(key='edge.classes_24h').value
    res = next(r for r in s['by_network'] if r['network'] == 'residential')
    assert (res['requests'], res['impersonator'], res['browser'], res['ips'], res['fake_ips']) == (5, 3, 2, 4, 3)
    fp = s['top_fake_fps'][0]
    assert (fp['ja4'], fp['requests'], fp['ip_hours'], fp['single_conn'], fp['residential']) == (FAKE_JA4, 3, 3, 3, 3)
    assert s['top_proxy_asns'][0]['asn'] == 28573 and s['top_proxy_asns'][0]['ips'] == 3
    assert {i['ip_address']: i['traps'] for i in s['top_fake_ips']}['203.0.113.1'] == 1

    content = staff_client.get('/acpwb-dashboard/edge/').content.decode()
    assert "Who&#x27;s Really Visiting" in content or "Who's Really Visiting" in content
    assert FAKE_JA4 in content
