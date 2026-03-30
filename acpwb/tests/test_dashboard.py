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
