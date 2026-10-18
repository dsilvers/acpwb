import io
import uuid
from datetime import timedelta

import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.management.commands.rollup_edge_hour import floor_hour, rollup_hour
from apps.core.models import DashboardStat
from apps.honeypot.models import (
    CrawlerVisit, EdgeClientHour, EdgeConnection, EdgeFingerprintHour, EdgeIPHour, EdgeRequest,
)

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
                                method='GET', route=route, status=200, client_class='browser', network='residential',
                                ja4=CHROME_JA4, stack='Chromium', new_client=new, client_id=client))
    # A browser in a datacenter, classified before the edge had a headless
    # class: rolled up as headless.
    reqs.append(EdgeRequest(timestamp=t, request_id=uuid.uuid4(), ip_address='192.0.2.50', route='django',
                            status=200, client_class='browser', network='hosting', ja4=CHROME_JA4))
    # A proxy exit that only opened and dropped connections, and a failed
    # handshake from the pool.
    conns.append(EdgeConnection(timestamp=t, ip_address='203.0.113.99', outcome='no_hello', conn_count=2,
                                network='residential', asn=28573, idempotency_key=uuid.uuid4()))
    conns.append(EdgeConnection(timestamp=t, ip_address='203.0.113.2', outcome='handshake_failed', ja4=FAKE_JA4,
                                network='residential', asn=28573, idempotency_key=uuid.uuid4()))
    # Outside the hour: ignored.
    reqs.append(EdgeRequest(timestamp=start + timedelta(hours=1, minutes=1), request_id=uuid.uuid4(),
                            ip_address='203.0.113.1', client_class='impersonator'))
    EdgeConnection.objects.bulk_create(conns)
    EdgeRequest.objects.bulk_create(reqs)
    CrawlerVisit.objects.create(ip_address='203.0.113.1', trap_type='ghost_link', path='/x', timestamp=t,
                                request_id=reqs[0].request_id)
    # A trap hit from the proxy exit that has no requests this hour.
    CrawlerVisit.objects.create(ip_address='203.0.113.99', trap_type='env_probe', path='/.env', timestamp=t)


@pytest.mark.django_db
def test_rollup_hour_per_ip_and_fingerprint():
    start = _hour()
    _traffic(start)
    assert rollup_hour(start) == (6, 2, 1)

    pool = EdgeIPHour.objects.get(hour=start, ip_address='203.0.113.1')
    assert (pool.requests, pool.conns, pool.impersonator_reqs, pool.browser_reqs) == (1, 1, 1, 0)
    assert (pool.network, pool.asn, pool.top_stack, pool.signals) == ('residential', 28573, 't13d311200', 0b11)
    assert (pool.new_client_reqs, pool.cookie_reqs, pool.trap_hits) == (1, 0, 1)

    assert EdgeIPHour.objects.get(hour=start, ip_address='203.0.113.2').failed_handshakes == 1
    exit_only = EdgeIPHour.objects.get(hour=start, ip_address='203.0.113.99')
    assert (exit_only.requests, exit_only.no_hello_conns, exit_only.network, exit_only.asn) == (0, 2, 'residential', 28573)
    assert (exit_only.trap_hits, exit_only.behaviour_proven_reqs, exit_only.bare_page_reqs) == (1, 0, 0)
    dc = EdgeIPHour.objects.get(hour=start, ip_address='192.0.2.50')
    assert (dc.headless_reqs, dc.browser_reqs) == (1, 0)

    browser = EdgeIPHour.objects.get(hour=start, ip_address='198.51.100.9')
    assert (browser.requests, browser.browser_reqs, browser.static_reqs, browser.cookie_reqs) == (2, 2, 1, 1)
    assert (browser.page_reqs, browser.pool_cookie_reqs) == (1, 0)
    assert browser.trap_hits == 0

    fake = EdgeFingerprintHour.objects.get(hour=start, ja4=FAKE_JA4)
    # The failed handshake is one more connection from an IP already counted.
    assert (fake.conns, fake.ips, fake.residential_ips, fake.single_conn_ips) == (4, 3, 3, 2)
    assert (fake.impersonator_reqs, fake.browser_reqs, fake.top_stack) == (3, 0, 't13d311200')
    assert fake.top_user_agent == 'Mozilla/5.0 Chrome/131'
    assert (fake.trap_reqs, fake.trap_ips) == (1, 1)  # the trap hit, by its request_id
    real = EdgeFingerprintHour.objects.get(hour=start, ja4=CHROME_JA4)
    assert (real.h2_fingerprint, real.requests, real.browser_reqs, real.single_conn_ips) == (CHROME_H2, 2, 2, 1)

    # Only the browser sent its cookie back.
    client = EdgeClientHour.objects.get(hour=start)
    assert (client.client_id, client.first_seen, client.requests, client.page_reqs, client.static_reqs) == ('c' * 22, True, 2, 1, 1)
    assert (client.ips, client.residential_ips, client.browser_reqs, client.top_ip) == (1, 1, 2, '198.51.100.9')

    # Re-running an hour replaces it.
    assert rollup_hour(start) == (6, 2, 1)
    assert EdgeIPHour.objects.filter(hour=start).count() == 6


@pytest.mark.django_db
def test_rollup_clients_pools_and_traps():
    """One cookie jar shared by a proxy pool (three IPs), a scraper that
    loads pages and never their assets, and a real-Chrome client caught by
    a ghost link."""
    start = _hour()
    t = start + timedelta(minutes=10)

    def req(ip, client, new=False, route='acpwb_go', cls='browser', **kw):
        return EdgeRequest(timestamp=t, request_id=uuid.uuid4(), ip_address=ip, method='GET', route=route,
                           status=200, client_class=cls, network='residential', ja4=CHROME_JA4, stack='Chromium',
                           client_id=client, new_client=new, **kw)
    pool = 'p' * 22
    scraper = 's' * 22
    caught = 'g' * 22
    reqs = [req('203.0.113.10', pool, new=True), req('203.0.113.11', pool), req('203.0.113.12', pool),
            req('198.51.100.30', scraper, new=True), req('198.51.100.30', scraper), req('198.51.100.30', scraper),
            req('198.51.100.40', caught, new=True), req('198.51.100.40', caught, route='static'),
            req('198.51.100.40', caught, route='django', cls='headless'),
            # A one-request client: no row.
            req('198.51.100.50', 'n' * 22, new=True),
            # A browser farm: a fresh profile per IP fetching one page, no assets.
            req('198.51.100.60', 'f' * 22, new=True)]
    EdgeRequest.objects.bulk_create(reqs)
    CrawlerVisit.objects.create(ip_address='198.51.100.40', trap_type='ghost_link', path='/employees/export/',
                                timestamp=t + timedelta(seconds=1), request_id=reqs[8].request_id)
    rollup_hour(start)

    rows = {c.client_id: c for c in EdgeClientHour.objects.filter(hour=start)}
    assert set(rows) == {pool, scraper, caught}
    assert (rows[pool].ips, rows[pool].residential_ips, rows[pool].requests) == (3, 3, 3)
    assert (rows[scraper].first_seen, rows[scraper].page_reqs, rows[scraper].static_reqs) == (True, 3, 0)
    assert (rows[caught].trap_reqs, rows[caught].static_reqs, rows[caught].headless_reqs) == (1, 1, 1)
    assert rows[caught].top_class == 'browser'

    ips = {r.ip_address: r for r in EdgeIPHour.objects.filter(hour=start)}
    assert [ips[f'203.0.113.{i}'].pool_cookie_reqs for i in (10, 11, 12)] == [1, 1, 1]
    assert (ips['198.51.100.30'].pool_cookie_reqs, ips['198.51.100.30'].page_reqs) == (0, 3)

    # Convicted by behaviour: a bare page load, a pool cookie, a trap hit.
    # The scraper returns its cookie, so its pages aren't bare.
    bare = {ip: (r.bare_page_reqs, r.behaviour_proven_reqs) for ip, r in ips.items()}
    assert bare['198.51.100.60'] == (1, 1)
    assert bare['203.0.113.11'] == (0, 1)
    assert bare['198.51.100.40'] == (0, 2)  # its browser requests; the headless one was proven already
    assert bare['198.51.100.30'] == (0, 0)


@pytest.mark.django_db
def test_rollup_command_follows_its_watermark():
    start = _hour(3)
    _traffic(start)
    call_command('rollup_edge_hour', since=start.strftime('%Y-%m-%dT%H'), stdout=io.StringIO())
    assert EdgeIPHour.objects.filter(hour=start).count() == 6
    mark = DashboardStat.objects.get(key='edge_rollup_watermark').value['hour']
    assert mark >= (start + timedelta(hours=1)).isoformat()

    # Caught up: the next run starts after the watermark and leaves the
    # rolled-up hour alone.
    EdgeIPHour.objects.filter(hour=start).delete()
    out = io.StringIO()
    call_command('rollup_edge_hour', stdout=out)
    assert not EdgeIPHour.objects.filter(hour=start).exists()

    call_command('rollup_edge_hour', hour=start.strftime('%Y-%m-%dT%H'), stdout=io.StringIO())
    assert EdgeIPHour.objects.filter(hour=start).count() == 6


@pytest.mark.django_db(databases=['default', 'direct'], transaction=True)
def test_precalc_edge_classes_and_dashboard(staff_client):
    start = _hour()
    _traffic(start)
    rollup_hour(start)
    call_command('precalc_dashboard', stdout=io.StringIO())

    s = DashboardStat.objects.get(key='edge.classes_24h').value
    res = next(r for r in s['by_network'] if r['network'] == 'residential')
    assert (res['requests'], res['impersonator'], res['browser'], res['ips'], res['fake_ips']) == (5, 3, 2, 5, 3)
    assert (res['no_hello'], res['no_hello_ips'], res['failed_handshakes']) == (2, 1, 1)
    hosting = next(r for r in s['by_network'] if r['network'] == 'hosting')
    assert (hosting['headless'], hosting['browser']) == (1, 0)
    fp = s['top_fake_fps'][0]
    assert (fp['ja4'], fp['requests'], fp['ip_hours'], fp['single_conn'], fp['residential']) == (FAKE_JA4, 3, 3, 2, 3)
    asn = s['top_proxy_asns'][0]
    assert (asn['asn'], asn['ips'], asn['fake_ips'], asn['no_hello_ips']) == (28573, 4, 3, 1)
    assert {i['ip_address']: i['traps'] for i in s['top_fake_ips']}['203.0.113.1'] == 1

    # The real browser fetched assets and returned its cookie: unproven.
    assert s['headline']['3h'] | {'since': ''} == {'total': 6, 'by_class': 4, 'by_behaviour': 0, 'unproven': 2,
                                                   'bare_pages': 0, 'bare_ips': 0, 'since': ''}

    # The real browser's cookie came from one IP; nothing on 2+.
    assert (s['multi_ip']['clients'], s['multi_ip']['multi']) == (1, 0)
    assert s['exposed'] == [{'client_class': 'browser', 'client_hours': 1, 'pooled': 0, 'no_render': 0,
                             'trapped': 0, 'any': 0}]
    assert s['trap_fps'][0]['ja4'] == FAKE_JA4 and s['trap_fps'][0]['trap_ips'] == 1

    content = staff_client.get('/acpwb-dashboard/edge/').content.decode()
    assert 'Proven Automated' in content and 'Cookies on 2+ IPs' in content and 'Caught by Behaviour' in content
    assert "Who&#x27;s Really Visiting" in content or "Who's Really Visiting" in content
    assert FAKE_JA4 in content
