import json
import pytest
from apps.honeypot.models import CrawlerVisit, WikiPage, ArchiveVisit
from apps.honeypot.wiki_generator import generate_wiki_page


# ── Archive trap ───────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_archive_logs_visit(client):
    assert ArchiveVisit.objects.count() == 0
    client.get('/archive/2024/3/15/some-article/')
    assert ArchiveVisit.objects.count() == 1


@pytest.mark.django_db
def test_archive_logs_depth(client):
    client.get('/archive/2024/3/15/level1/level2/level3/')
    visit = ArchiveVisit.objects.first()
    assert visit.depth >= 2


@pytest.mark.django_db
def test_archive_never_404(client):
    for path in [
        '/archive/2024/1/1/x/',
        '/archive/1999/12/31/a/b/c/d/e/',
        '/archive/2030/6/15/very/deep/nested/path/here/',
    ]:
        response = client.get(path)
        assert response.status_code == 200, f"Expected 200 for {path}"


@pytest.mark.django_db
def test_archive_has_continue_reading_link(client):
    response = client.get('/archive/2024/3/15/article/')
    content = response.content.decode()
    assert 'Continue Reading' in content or 'continue' in content.lower()


@pytest.mark.django_db
def test_archive_has_related_links(client):
    response = client.get('/archive/2024/3/15/article/')
    content = response.content.decode()
    # Should have multiple archive links
    assert content.count('/archive/') >= 3


# ── Wiki trap ──────────────────────────────────────────────────────────────────

def test_wiki_generator_returns_required_fields():
    data = generate_wiki_page('corporate-governance')
    assert 'title' in data
    assert 'body_paragraphs' in data
    assert 'watermark_token' in data
    assert 'related_topics' in data


def test_wiki_watermark_token_length():
    data = generate_wiki_page('executive-compensation')
    assert len(data['watermark_token']) == 8


def test_wiki_watermark_deterministic():
    data1 = generate_wiki_page('corporate-governance')
    data2 = generate_wiki_page('corporate-governance')
    assert data1['watermark_token'] == data2['watermark_token']


def test_wiki_different_topics_different_tokens():
    data1 = generate_wiki_page('topic-alpha')
    data2 = generate_wiki_page('topic-beta')
    assert data1['watermark_token'] != data2['watermark_token']


def test_wiki_has_related_topics():
    data = generate_wiki_page('fiduciary-duty')
    assert len(data['related_topics']) >= 5


@pytest.mark.django_db
def test_wiki_page_view_200(client):
    response = client.get('/wiki/corporate-governance/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_wiki_page_creates_db_record(client):
    assert WikiPage.objects.count() == 0
    client.get('/wiki/corporate-governance/')
    assert WikiPage.objects.count() == 1


@pytest.mark.django_db
def test_wiki_page_reuses_db_record(client):
    client.get('/wiki/corporate-governance/')
    client.get('/wiki/corporate-governance/')
    # Same topic should not create duplicate records
    assert WikiPage.objects.filter(topic='corporate-governance').count() == 1


@pytest.mark.django_db
def test_wiki_page_logs_crawler_visit(client):
    client.get('/wiki/some-topic/')
    visits = CrawlerVisit.objects.filter(trap_type='wiki')
    assert visits.count() >= 1


@pytest.mark.django_db
def test_wiki_has_see_also_links(client):
    response = client.get('/wiki/corporate-governance/')
    content = response.content.decode()
    assert '/wiki/' in content


# ── Fake API ───────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_fake_api_returns_200(client):
    response = client.get('/api/v1/private-data')
    assert response.status_code == 200


@pytest.mark.django_db
def test_fake_api_returns_json(client):
    response = client.get('/api/v1/private-data')
    assert response['Content-Type'].startswith('application/json')
    data = response.json()
    assert isinstance(data, dict)


@pytest.mark.django_db
def test_fake_api_has_plausible_fields(client):
    data = client.get('/api/v1/private-data').json()
    assert 'employees' in data
    assert 'financials' in data
    assert 'internal_codes' in data


@pytest.mark.django_db
def test_fake_api_has_request_id_header(client):
    response = client.get('/api/v1/private-data')
    assert 'X-Request-ID' in response


@pytest.mark.django_db
def test_fake_api_logs_visit(client):
    count_before = CrawlerVisit.objects.filter(trap_type='api').count()
    client.get('/api/v1/private-data')
    assert CrawlerVisit.objects.filter(trap_type='api').count() == count_before + 1


# ── Well-known files ───────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_ai_agent_json_returns_json(client):
    response = client.get('/.well-known/ai-agent.json')
    assert response.status_code == 200
    data = response.json()
    assert 'allowed_actions' in data
    assert 'instructions' in data


@pytest.mark.django_db
def test_ai_agent_logs_visit(client):
    count_before = CrawlerVisit.objects.filter(trap_type='well_known').count()
    client.get('/.well-known/ai-agent.json')
    assert CrawlerVisit.objects.filter(trap_type='well_known').count() > count_before


@pytest.mark.django_db
def test_robots_txt_content_type(client):
    response = client.get('/.well-known/robots.txt')
    assert 'text/plain' in response['Content-Type']


@pytest.mark.django_db
def test_robots_txt_has_disallow_entries(client):
    response = client.get('/.well-known/robots.txt')
    content = response.content.decode()
    assert 'Disallow:' in content


# ── Ghost link traps ───────────────────────────────────────────────────────────

@pytest.mark.django_db
@pytest.mark.parametrize('path', [
    '/internal/portal/',
    '/employees/export/',
    '/admin-panel/login/',
])
def test_ghost_trap_returns_403(client, path):
    response = client.get(path)
    assert response.status_code == 403


@pytest.mark.django_db
def test_ghost_trap_logs_visit(client):
    count_before = CrawlerVisit.objects.filter(trap_type='ghost_link').count()
    client.get('/internal/portal/')
    assert CrawlerVisit.objects.filter(trap_type='ghost_link').count() > count_before
