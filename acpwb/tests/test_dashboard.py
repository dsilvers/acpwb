import pytest
from django.contrib.auth.models import User


@pytest.fixture
def staff_client(client, db):
    user = User.objects.create_user('staff', password='pass', is_staff=True)
    client.force_login(user)
    return client


# ── Auth gate ──────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_dashboard_redirects_anonymous(client):
    response = client.get('/acpwb-dashboard/')
    assert response.status_code == 302
    assert '/login/' in response['Location']


@pytest.mark.django_db
def test_dashboard_rejects_non_staff(client, db):
    user = User.objects.create_user('regular', password='pass', is_staff=False)
    client.force_login(user)
    response = client.get('/acpwb-dashboard/')
    assert response.status_code == 302


# ── Overview ───────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_dashboard_overview_200(staff_client):
    response = staff_client.get('/acpwb-dashboard/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_dashboard_overview_has_counts(staff_client):
    response = staff_client.get('/acpwb-dashboard/')
    assert b'Crawler Hits' in response.content
    assert b'Archive Visits' in response.content
    assert b'Inbound Emails' in response.content


# ── Sub-views ──────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_dashboard_crawlers_200(staff_client):
    response = staff_client.get('/acpwb-dashboard/crawlers/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_dashboard_archive_200(staff_client):
    response = staff_client.get('/acpwb-dashboard/archive/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_dashboard_emails_200(staff_client):
    response = staff_client.get('/acpwb-dashboard/emails/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_dashboard_people_200(staff_client):
    response = staff_client.get('/acpwb-dashboard/people/')
    assert response.status_code == 200


# ── Data populates after trap visits ──────────────────────────────────────────

@pytest.mark.django_db
def test_dashboard_shows_crawler_data(client, staff_client):
    client.get('/internal/portal/')
    response = staff_client.get('/acpwb-dashboard/')
    assert response.status_code == 200
    # recent_crawlers is a live query — shows trap type even before precalc runs
    assert b'ghost_link' in response.content


# ── TLS fingerprints (acpwb_edge) ──────────────────────────────────────────────

@pytest.mark.django_db
def test_dashboard_edge_200_empty(staff_client):
    response = staff_client.get('/acpwb-dashboard/edge/')
    assert response.status_code == 200
    assert b'TLS Fingerprints' in response.content


@pytest.mark.django_db
def test_dashboard_edge_shows_summary(staff_client):
    from apps.core.models import DashboardStat
    DashboardStat.objects.create(key='edge.summary_24h', value={
        'total': 10,
        'by_outcome': {'ok': 8, 'handshake_failed': 2},
        'by_alpn': {'h2': 5, 'http/1.1': 3, '(none)': 2},
        'by_tls_version': {'TLSv1.3': 8},
        'top_ja4': [{'ja4': 't13d1516h2_8daaf6152771_e5627efa2ab1', 'count': 5, 'ips': 2, 'uas': 4}],
        'top_failed_ja4': [],
        'top_h2_fingerprint': [{'h2_fingerprint': '1:65536|15663105|0|m,a,s,p', 'count': 5, 'ips': 2}],
    })
    content = staff_client.get('/acpwb-dashboard/edge/').content
    assert b't13d1516h2_8daaf6152771_e5627efa2ab1' in content
    assert b'20.0%' in content  # failed share
    assert b'50.0%' in content  # h2 share


@pytest.mark.django_db
def test_dashboard_edge_shows_process_counters(staff_client):
    from apps.core.models import DashboardStat
    DashboardStat.objects.create(key='edge.summary_24h', value={'total': 0})
    DashboardStat.objects.create(key='edge.process', value={
        'conns_open': 12, 'conns_total': 3456, 'records_pushed_conn': 3400, 'records_dropped_conn': 56,
        'records_pushed_request': 9000, 'records_dropped_request': 0,
        'records_aggregated': 0, 'rejected_requests': 0, 'parser_panics': 0, 'tap_panics': 0,
        'goroutines': 40, 'started_at': 1,
    })
    content = staff_client.get('/acpwb-dashboard/edge/').content.decode()
    assert 'Edge Process' in content
    assert '3456' in content
    assert 'color:#dc3545;font-weight:700">56<' in content  # dropped records flagged
    assert '9000' in content


@pytest.mark.django_db
def test_dashboard_edge_shows_requests(staff_client):
    from apps.core.models import DashboardStat
    from apps.honeypot.models import EdgeRequest
    DashboardStat.objects.create(key='edge.summary_24h', value={'total': 0})
    DashboardStat.objects.create(key='edge.requests_24h', value={
        'total': 200, 'clients': 7, 'new_clients': 3, 'ips': 9, 'failed': 20, 'bytes_out': 1,
        'by_route': [{'name': 'static', 'count': 150, 'avg_ms': 1.5, 'bytes': 2048}],
        'by_status': {'200': 180, '400': 20}, 'by_error': {'rejected': 20},
        'by_protocol': {'HTTP/2.0': 200},
    })
    DashboardStat.objects.create(key='edge.classes_24h', value={
        'cookie_pools': [{'client_id': 'AAAAAAAAAAAAAAAAAAAAAA', 'ip_hours': 4, 'max_ips': 4, 'hours': 1,
                          'requests': 30, 'client_class': 'browser', 'traps': 0}],
    })
    EdgeRequest.objects.create(request_id='0f8fad5b-d9cb-469f-a165-70867728950e', ip_address='192.0.2.5',
                               method='GET', host='acpwb.com', path='/static/x.css', status=200, route='static')
    content = staff_client.get('/acpwb-dashboard/edge/').content.decode()
    assert '10.0%' in content  # failed share
    assert 'AAAAAAAAAAAAAAAAAAAAAA' in content
    assert 'rejected' in content
    assert '/static/x.css' in content


# ── Privacy requests ───────────────────────────────────────────────────────────

@pytest.fixture
def privacy_requests_data(db):
    from apps.public.models import DataOptOutRequest
    a = DataOptOutRequest.objects.create(name='Ann', email='ann@example.com', request_type='do_not_sell', ip_address='1.2.3.4', message='stop selling')
    b = DataOptOutRequest.objects.create(name='Bob', email='bob@example.com', request_type='delete', ip_address='1.2.3.5', processed=True)
    return a, b


@pytest.mark.django_db
def test_privacy_requires_staff(client):
    assert client.get('/acpwb-dashboard/privacy/').status_code == 302


@pytest.mark.django_db
def test_privacy_lists_and_filters(staff_client, privacy_requests_data):
    r = staff_client.get('/acpwb-dashboard/privacy/')
    assert r.status_code == 200
    assert b'ann@example.com' in r.content and b'bob@example.com' in r.content
    r = staff_client.get('/acpwb-dashboard/privacy/?type=do_not_sell')
    assert b'ann@example.com' in r.content and b'bob@example.com' not in r.content
    r = staff_client.get('/acpwb-dashboard/privacy/?status=done')
    assert b'bob@example.com' in r.content and b'ann@example.com' not in r.content


@pytest.mark.django_db
def test_privacy_mark_processed(staff_client, privacy_requests_data):
    a, _ = privacy_requests_data
    r = staff_client.post('/acpwb-dashboard/privacy/', {'id': a.pk, 'processed': '1'})
    assert r.status_code == 302
    a.refresh_from_db()
    assert a.processed is True
