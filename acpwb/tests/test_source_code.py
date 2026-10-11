"""Legacy source-code pages: Django-side pieces (identity mirror, article
link markup, code-email registry + inbound matching, trap type). The page
itself is rendered by acpwb_go; see acpwb_go/legacyjob for its tests."""
import pytest

from apps.honeypot.legacy_identity import legacy_identity, source_code_href
from apps.honeypot.models import CrawlerVisit, SourceCodeEmail
from apps.webhooks.models import HoneypotMatch, InboundEmail

SLUG = 'human-factors-and-usability-engineering-report-9293'


def test_identity_matches_go():
    # Same values acpwb_go renders for this article (see archive fixtures).
    ident = legacy_identity(2022, 4, 6, SLUG)
    assert ident['program'] == 'EQUTAB'
    assert ident['job_number'] == 'JOB88624'
    assert ident == legacy_identity(2022, 4, 6, SLUG)


@pytest.mark.parametrize('year', [1985, 1987, 1993, 1997, 2001, 2008, 2016, 2020, 2025])
def test_identity_era_pools(year):
    from apps.honeypot.legacy_identity import _LANGUAGE_TABLE
    live = {k for k, _label, a, b, _w in _LANGUAGE_TABLE if a <= year <= b}
    for i in range(200):
        assert legacy_identity(year, 1 + i % 12, 1 + i % 28, f's-{i}')['lang_key'] in live


def test_source_code_href():
    assert source_code_href('/archive/2022/04/06/x/export.csv') == '/archive/2022/04/06/x/source-code/'
    assert source_code_href('/04/06/x/export.csv') == '/04/06/x/source-code/'


def test_article_render_links_to_source_code():
    from apps.honeypot.management.commands.dump_archive_default_fixtures import build_context
    from apps.honeypot.pyrender.archive_main import render_archive_default
    html = render_archive_default(build_context(2022, 4, 6, SLUG))
    assert html.count('/archive/2022/04/06/' + SLUG + '/source-code/') == 4
    assert 'Computation Source' in html


def test_trap_type_choice():
    assert 'source_code' in dict(CrawlerVisit.TRAP_CHOICES)


@pytest.mark.django_db
def test_drain_upsert_creates_then_updates():
    from apps.core.management.commands.drain_code_email_queue import upsert_code_email
    item = {'address': 'Xiao.Whitehead@acpwb.com', 'host': 'archives-2022.acpwb.com',
            'path': '/04/06/' + SLUG + '/source-code/', 'ip_address': '10.1.2.3',
            'timestamp': '2026-10-10T05:00:00+00:00'}
    assert upsert_code_email(item) is True
    assert upsert_code_email({**item, 'timestamp': '2026-10-11T05:00:00+00:00'}) is False
    row = SourceCodeEmail.objects.get()
    assert row.address == 'xiao.whitehead@acpwb.com'
    assert row.times_served == 2
    assert row.first_ip == '10.1.2.3'
    assert row.last_seen > row.first_seen
    assert upsert_code_email({'address': '', 'path': '/x/'}) is False


@pytest.mark.django_db
def test_inbound_email_matches_source_code_page():
    from apps.webhooks.views import _match_honeypot
    page = SourceCodeEmail.objects.create(address='gail.hargrove@acpwb.com', host='acpwb.com',
                                          path='/archive/2022/04/06/' + SLUG + '/source-code/', first_ip='10.9.9.9')
    inbound = InboundEmail.objects.create(sender='spam@example.com', recipient='Gail.Hargrove@acpwb.com')
    _match_honeypot(inbound)
    m = HoneypotMatch.objects.get(inbound_email=inbound)
    assert m.match_confidence == 'exact'
    assert m.source_code_email == page
    assert page.path in m.notes


@pytest.mark.django_db
def test_inbound_email_no_match():
    from apps.webhooks.views import _match_honeypot
    inbound = InboundEmail.objects.create(sender='spam@example.com', recipient='nobody.here@acpwb.com')
    _match_honeypot(inbound)
    assert HoneypotMatch.objects.get(inbound_email=inbound).match_confidence == 'none'


@pytest.mark.django_db
def test_archive_subdomain_robots_allows_source_code(client):
    with _no_queue():
        body = client.get('/robots.txt?__year=1993').content.decode()
    assert 'Allow: /source-code/' in body
    assert 'Allow: /*/source-code/' in body
    assert 'Sitemap: https://archives-1993.acpwb.com/sitemap-source-code.xml' in body


@pytest.mark.django_db
def test_archive_subdomain_source_code_sitemap(client):
    with _no_queue():
        body = client.get('/sitemap-source-code.xml?__year=1993').content.decode()
    assert '<loc>https://archives-1993.acpwb.com/source-code/</loc>' in body
    for m in range(1, 13):
        assert f'<loc>https://archives-1993.acpwb.com/source-code/{m:02d}/</loc>' in body
    assert body.count('/source-code/</loc>') > 24  # record listings, subdomain-relative
    assert 'acpwb.com/archive/' not in body


@pytest.mark.django_db
def test_main_source_code_sitemap_lists_every_month(client):
    with _no_queue():
        body = client.get('/sitemap-source-code.xml').content.decode()
    for m in range(1, 13):
        assert f'/source-code/1993/{m:02d}/</loc>' in body


def _no_queue():
    from unittest.mock import patch
    return patch('apps.core.crawler_queue.push_crawler_visit', return_value=True)
