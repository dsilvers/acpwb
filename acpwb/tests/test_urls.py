import pytest


@pytest.mark.django_db
def test_home_200(client):
    response = client.get('/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_our_people_200(client):
    response = client.get('/our-people/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_careers_200(client):
    response = client.get('/careers/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_mission_200(client):
    response = client.get('/mission/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_partners_200(client):
    response = client.get('/partners/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_privacy_200(client):
    response = client.get('/privacy/')
    assert response.status_code == 200


# PHP legacy redirects
@pytest.mark.django_db
def test_php_index_redirect(client):
    response = client.get('/index.php')
    assert response.status_code == 301
    assert response['Location'] == '/'


@pytest.mark.django_db
def test_php_profiles_redirect(client):
    response = client.get('/profiles.php')
    assert response.status_code == 301
    assert response['Location'] == '/our-people/'


@pytest.mark.django_db
def test_php_mission_redirect(client):
    response = client.get('/mission.php')
    assert response.status_code == 301
    assert response['Location'] == '/mission/'


@pytest.mark.django_db
def test_php_disclaimer_redirect(client):
    response = client.get('/disclaimer.php')
    assert response.status_code == 301
    assert response['Location'] == '/privacy/'


# Honeypot endpoints
@pytest.mark.django_db
def test_ghost_trap_internal_portal(client):
    response = client.get('/internal/portal/')
    assert response.status_code == 403


@pytest.mark.django_db
def test_ghost_trap_employees_export(client):
    response = client.get('/employees/export/')
    assert response.status_code == 403


@pytest.mark.django_db
def test_ghost_trap_admin_panel(client):
    response = client.get('/admin-panel/login/')
    assert response.status_code == 403


@pytest.mark.django_db
def test_fake_api_200(client):
    response = client.get('/api/v1/private-data')
    assert response.status_code == 200


@pytest.mark.django_db
def test_archive_any_path(client):
    response = client.get('/archive/2024/3/15/some-slug/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_archive_deep_path(client):
    response = client.get('/archive/2023/1/1/a/b/c/d/e/f/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_wiki_page(client):
    response = client.get('/wiki/corporate-governance/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_well_known_ai_agent(client):
    response = client.get('/.well-known/ai-agent.json')
    assert response.status_code == 200


@pytest.mark.django_db
def test_well_known_robots(client):
    response = client.get('/.well-known/robots.txt')
    assert response.status_code == 200


@pytest.mark.django_db
def test_contact_redirects_to_our_people(client):
    response = client.get('/contact/')
    assert response.status_code in (301, 302)
    assert '/our-people/' in response['Location']
