import pytest
from unittest.mock import patch
from apps.honeypot.models import CrawlerVisit

# A known-valid agency/slug pair (apps/honeypot/policy_data.py) used across
# these tests so the same deterministic content is exercised repeatedly.
AGENCY = 'sec'
SLUG = 'executive-compensation-disclosure-requirements'
YEAR, MONTH, DAY = 2020, 5, 10


# ── Main-domain policy pages ────────────────────────────────────────────────

@pytest.mark.django_db
def test_policy_index_serves_content(client):
    response = client.get('/public-policy/')
    assert response.status_code == 200
    assert b'Public Policy' in response.content


@pytest.mark.django_db
def test_policy_year_serves_content(client):
    response = client.get(f'/public-policy/{YEAR}/')
    assert response.status_code == 200
    assert str(YEAR).encode() in response.content


@pytest.mark.django_db
def test_policy_month_serves_content(client):
    response = client.get(f'/public-policy/{YEAR}/{MONTH}/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_policy_detail_serves_content(client):
    response = client.get(f'/public-policy/{YEAR}/{MONTH}/{DAY}/{AGENCY}/{SLUG}/')
    assert response.status_code == 200
    content = response.content.decode()
    assert 'Compensation'.lower() in content.lower() or 'compensation' in content.lower()


@pytest.mark.django_db
def test_policy_detail_never_404_on_unknown_slug(client):
    # Content is deterministically generated from the URL, not looked up —
    # an unrecognized slug should still render, not 404.
    response = client.get(f'/public-policy/{YEAR}/{MONTH}/{DAY}/{AGENCY}/totally-made-up-slug/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_policy_detail_has_related_sections(client):
    response = client.get(f'/public-policy/{YEAR}/{MONTH}/{DAY}/{AGENCY}/{SLUG}/')
    content = response.content.decode()
    # Sidebar sections that should be present given fixed, deterministic seeds.
    assert 'Related' in content


@pytest.mark.django_db
def test_policy_logs_visit(client):
    with patch('apps.core.crawler_queue.push_crawler_visit', return_value=False):
        assert CrawlerVisit.objects.filter(trap_type='policy').count() == 0
        client.get(f'/public-policy/{YEAR}/{MONTH}/{DAY}/{AGENCY}/{SLUG}/')
        assert CrawlerVisit.objects.filter(trap_type='policy').count() == 1


# ── Policy subdomain (policy-<agency>.acpwb.com) via the DEBUG __agency shortcut ──
# See apps/core/subdomain_middleware.py: ?__agency=<slug> bypasses DNS in DEBUG,
# same mechanism the archive tests use with ?__year=.

@pytest.mark.django_db
def test_policy_subdomain_index_serves_content(client):
    response = client.get(f'/?__agency={AGENCY}')
    assert response.status_code == 200


@pytest.mark.django_db
def test_policy_subdomain_year_serves_content(client):
    response = client.get(f'/{YEAR}/?__agency={AGENCY}')
    assert response.status_code == 200


@pytest.mark.django_db
def test_policy_subdomain_month_serves_content(client):
    response = client.get(f'/{YEAR}/{MONTH}/?__agency={AGENCY}')
    assert response.status_code == 200


@pytest.mark.django_db
def test_policy_subdomain_detail_serves_content(client):
    response = client.get(f'/{YEAR}/{MONTH}/{DAY}/{SLUG}/?__agency={AGENCY}')
    assert response.status_code == 200
    assert response.content  # non-empty body
