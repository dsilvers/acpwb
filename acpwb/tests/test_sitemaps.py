import pytest
from xml.etree import ElementTree


SITEMAP_NS = 'http://www.sitemaps.org/schemas/sitemap/0.9'


def parse_sitemap(content):
    return ElementTree.fromstring(content)


# ── Real sitemap ───────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_sitemap_xml_200(client):
    response = client.get('/sitemap.xml')
    assert response.status_code == 200


@pytest.mark.django_db
def test_sitemap_xml_content_type(client):
    response = client.get('/sitemap.xml')
    assert 'xml' in response['Content-Type']


@pytest.mark.django_db
def test_sitemap_xml_has_urls(client):
    response = client.get('/sitemap.xml')
    root = parse_sitemap(response.content)
    locs = root.findall(f'.//{{{SITEMAP_NS}}}loc')
    assert len(locs) >= 5  # at least the static pages


# ── Trap sitemaps ──────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_sitemap_publications_200(client):
    response = client.get('/sitemap-publications.xml')
    assert response.status_code == 200


@pytest.mark.django_db
def test_sitemap_publications_is_valid_xml(client):
    response = client.get('/sitemap-publications.xml')
    root = parse_sitemap(response.content)
    locs = root.findall(f'.//{{{SITEMAP_NS}}}loc')
    assert len(locs) >= 10


@pytest.mark.django_db
def test_sitemap_publications_logs_visit(client):
    from apps.honeypot.models import CrawlerVisit
    before = CrawlerVisit.objects.filter(trap_type='well_known').count()
    client.get('/sitemap-publications.xml')
    assert CrawlerVisit.objects.filter(trap_type='well_known').count() > before


@pytest.mark.django_db
def test_sitemap_wiki_200(client):
    response = client.get('/sitemap-wiki.xml')
    assert response.status_code == 200


@pytest.mark.django_db
def test_sitemap_wiki_has_many_topics(client):
    response = client.get('/sitemap-wiki.xml')
    root = parse_sitemap(response.content)
    locs = root.findall(f'.//{{{SITEMAP_NS}}}loc')
    assert len(locs) >= 75


@pytest.mark.django_db
def test_sitemap_wiki_logs_visit(client):
    from apps.honeypot.models import CrawlerVisit
    before = CrawlerVisit.objects.filter(trap_type='well_known').count()
    client.get('/sitemap-wiki.xml')
    assert CrawlerVisit.objects.filter(trap_type='well_known').count() > before


@pytest.mark.django_db
def test_sitemap_archive_200(client):
    response = client.get('/sitemap-archive.xml')
    assert response.status_code == 200


@pytest.mark.django_db
def test_sitemap_archive_has_500_entries(client):
    response = client.get('/sitemap-archive.xml')
    root = parse_sitemap(response.content)
    locs = root.findall(f'.//{{{SITEMAP_NS}}}loc')
    assert len(locs) == 500


@pytest.mark.django_db
def test_sitemap_archive_is_deterministic(client):
    """Same seed → same sitemap every time."""
    r1 = client.get('/sitemap-archive.xml').content
    r2 = client.get('/sitemap-archive.xml').content
    assert r1 == r2


@pytest.mark.django_db
def test_sitemap_archive_logs_visit(client):
    from apps.honeypot.models import CrawlerVisit
    before = CrawlerVisit.objects.filter(trap_type='well_known').count()
    client.get('/sitemap-archive.xml')
    assert CrawlerVisit.objects.filter(trap_type='well_known').count() > before
