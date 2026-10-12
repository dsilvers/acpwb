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
