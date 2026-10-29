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
def test_archive_index_200(client):
    response = client.get('/archive/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_archive_year_serves_content(client):
    # /archive/<year>/ serves content directly on main domain (no redirect)
    response = client.get('/archive/2024/')
    assert response.status_code == 200
    assert b'2024' in response.content


@pytest.mark.django_db
def test_archive_month_serves_content(client):
    response = client.get('/archive/2020/6/')
    assert response.status_code == 200
    assert b'2020' in response.content


@pytest.mark.django_db
def test_archive_trap_serves_content(client):
    response = client.get('/archive/2024/3/15/some-slug/')
    assert response.status_code == 200
    assert b'2024' in response.content


@pytest.mark.django_db
def test_archive_deep_path_serves_content(client):
    response = client.get('/archive/2023/1/1/a/b/c/d/e/f/')
    assert response.status_code == 200
    assert b'2023' in response.content


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


# New public pages
@pytest.mark.django_db
def test_awards_200(client):
    response = client.get('/awards/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_patents_200(client):
    response = client.get('/patents/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_faq_200(client):
    response = client.get('/faq/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_do_not_sell_200(client):
    response = client.get('/privacy/do-not-sell/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_accessibility_200(client):
    response = client.get('/accessibility/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_certification_200(client):
    response = client.get('/certification/')
    assert response.status_code == 200
    content = response.content.decode()
    assert 'ACPWB 9001½' in content
    assert 'acpwb-9001-half.svg' in content
    # The legal notice must stay on the page.
    assert 'not affiliated with' in content


@pytest.mark.django_db
def test_certification_names_seeded_by_org(client):
    from apps.public.views import _certification_people

    first = client.get('/certification/?org=Pinnacle Group').content.decode()
    again = client.get('/certification/?org=Pinnacle Group').content.decode()
    other = client.get('/certification/?org=Apex Consulting').content.decode()
    pinnacle = _certification_people('Pinnacle Group')
    apex = _certification_people('Apex Consulting')

    assert pinnacle == _certification_people('Pinnacle Group')
    assert pinnacle['assessor']['name'] in first
    assert pinnacle['assessor']['name'] in again
    assert apex['assessor']['name'] in other
    assert pinnacle != apex
    # Annex Z sits after the certificate.
    assert first.index('id="certificate-print"') < first.index('id="annex-z"')


@pytest.mark.django_db
def test_awards_links_certification(client):
    response = client.get('/awards/')
    assert '/certification/' in response.content.decode()


@pytest.mark.django_db
def test_trademarks_200(client):
    response = client.get('/trademarks/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_site_map_200(client):
    response = client.get('/site-map/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_sitemap_pages_xml_200(client):
    response = client.get('/sitemap-pages.xml')
    assert response.status_code == 200


@pytest.mark.django_db
def test_internal_login_get_200(client):
    response = client.get('/internal/login/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_internal_login_post_logs_credentials(client):
    response = client.post('/internal/login/', {'username': 'admin', 'password': 'secret123'})
    assert response.status_code in (200, 302)
    from apps.honeypot.models import InternalLoginAttempt
    assert InternalLoginAttempt.objects.filter(username='admin').exists()


@pytest.mark.django_db
def test_press_releases_feed(client):
    from xml.etree import ElementTree

    from apps.public.press_data import _PRESS_RELEASES

    response = client.get('/press-releases/feed.xml')
    assert response.status_code == 200
    assert response['Content-Type'].startswith('application/rss+xml')
    root = ElementTree.fromstring(response.content)
    items = root.findall('./channel/item')
    assert len(items) == len(_PRESS_RELEASES)
    first = _PRESS_RELEASES[0]
    assert items[0].findtext('title') == first['headline']
    assert items[0].findtext('link').endswith(f"/{first['slug']}/")
    content = items[0].findtext('{http://purl.org/rss/1.0/modules/content/}encoded')
    assert 'href="/' not in content and 'src="/' not in content


@pytest.mark.django_db
def test_careers_has_no_openings(client):
    response = client.get('/careers/')
    assert b'We currently do not have any job openings. Please check back later.' in response.content
    assert b'/careers/jobs/' not in response.content


@pytest.mark.parametrize('path', [
    '/careers/jobs/84/full-stack-developer-client-portals/',
    '/careers/jobs/84/full-stack-developer-client-portals/apply/',
    '/careers/jobs/84/full-stack-developer-client-portals/applied/',
])
@pytest.mark.django_db
def test_job_links_404(client, path):
    response = client.get(path)
    assert response.status_code == 404


@pytest.mark.django_db
def test_press_release_answer_tree_block(client):
    """The `tree` block renders its interactive container with an image fallback, and the
    figure's data covers all six attempts."""
    import json
    from pathlib import Path

    slug = 'compliance-center-first-results'
    response = client.get(f'/press-releases/2026/10/26/{slug}/')
    assert response.status_code == 200
    html = response.content.decode()
    assert 'class="press-tree"' in html
    assert 'answer-tree.json' in html and 'answer-tree-attempt-3.svg' in html
    assert 'js/press-tree.js' in html
    assert 'data-attempt="3"' in html and 'data-winner="dbaaaa"' in html

    folder = Path(__file__).resolve().parents[1] / 'static' / 'img' / 'press' / slug
    data = json.loads((folder / 'answer-tree.json').read_text())
    assert sorted(data['att']) == ['1', '2', '3', '4', '5', '6']
    assert 'dbaaaa' in data['att']['3']['pass']
    assert (folder / 'answer-tree-attempt-3.svg').stat().st_size > 0

    feed = client.get('/press-releases/feed.xml').content.decode()
    assert 'answer-tree-attempt-3.svg' in feed
