"""
Compare the old (Django/Jinja2 template) and new (hand-written Python
builder) render paths for archive/policy pages, for the SAME generated
content, and report whether they match once insignificant whitespace is
stripped (the accepted fidelity bar for this work — see
/Users/dan/.claude/plans/any-performance-benefits-to-dreamy-deer.md).

This is meant to be run by a developer during/after converting a template,
not as part of the automated test suite. It builds real context via the
actual generator functions (same deterministic seeds each run) and prints
the first several real differences if any are found, after normalizing
away pure-whitespace-between-tags noise.

Usage:
    python manage.py diff_render --case archive_default
    python manage.py diff_render --case all
"""
import difflib
import re

from django.core.management.base import BaseCommand, CommandError
from django.template.loader import render_to_string


def _strip_ws(s):
    # Strip HTML source comments (<!-- ... -->) before whitespace collapsing —
    # they're developer annotations in the old templates, never rendered
    # visible content, and the Python builders correctly don't reproduce them.
    # (The trailing "v2.4.1 | build..." comment is deliberately visible-in-source
    # honeypot content and IS reproduced by the builders, so this only removes
    # comments that don't survive into both sides' output anyway once matched.)
    s = re.sub(r'<!--(?!\[if).*?-->', '', s, flags=re.DOTALL)
    s = re.sub(r'\s+', '', s)
    # Jinja2 HTML-entity-escapes apostrophes even inside <style> blocks
    # (e.g. font-family: &#39;Courier New&#39; instead of 'Courier New') —
    # a pre-existing bug in the era templates, since <style> is a raw-text
    # element and browsers never decode entities there. The Python builders
    # deliberately emit the correct, unescaped form instead of reproducing
    # it, so normalize both representations as equivalent for comparison.
    s = s.replace('&#39;', "'").replace('&#x27;', "'")
    # Django's escape() emits &quot; for "; Jinja2/markupsafe emits &#34; —
    # both correct HTML, same normalization rationale as apostrophes above.
    s = s.replace('&#34;', '&quot;')
    return s


def _report_diff(stdout, real, mine, label):
    real_n, mine_n = _strip_ws(real), _strip_ws(mine)
    if real_n == mine_n:
        stdout.write(f'  [{label}] MATCH (content-equivalent; real={len(real):,}B mine={len(mine):,}B raw)')
        return True
    stdout.write(f'  [{label}] MISMATCH real={len(real_n):,}B mine={len(mine_n):,}B (post-whitespace-strip)')
    sm = difflib.SequenceMatcher(None, real_n, mine_n)
    shown = 0
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag != 'equal' and (i2 - i1 > 8 or j2 - j1 > 8):
            stdout.write(f'    {tag} real[{i1}:{i2}]: {real_n[max(0, i1-30):i2+30]!r}')
            stdout.write(f'    {tag} mine[{j1}:{j2}]: {mine_n[max(0, j1-30):j2+30]!r}')
            shown += 1
            if shown >= 6:
                stdout.write('    ...(more differences omitted)')
                break
    return False


def _archive_ctx(year, month, day, slug):
    from apps.honeypot import views as v
    content = v._generate_archive_content(year, month, day, slug)
    return _archive_ctx_common(v, year, month, day, slug, content)


def _archive_ctx_era(year, month, day, slug, generator_name):
    """Same shared fields, but on_archive_subdomain=True + year_data, and
    era-appropriate (absolute, cross-subdomain-safe) URLs for related_paths/
    related_docs, matching what archive_trap() actually builds when
    on_sub=True."""
    from apps.honeypot import views as v
    content = getattr(v, generator_name)(year, month, day, slug)
    ctx = _archive_ctx_common(v, year, month, day, slug, content)
    ctx['on_archive_subdomain'] = True
    ctx['year_data'] = v._year_data(year)
    # On the subdomain, same-year related paths are relative; the shared
    # ctx builder above always builds main-domain-style absolute paths, so
    # rebuild them the way _archive_url() would for on_sub=True, same year.
    ctx['related_paths'] = [
        {'url': f'/{r["month"]:02d}/{r["day"]:02d}/{r["full_slug"]}/' if r['year'] == year
         else f'https://archives-{r["year"]}.acpwb.com/{r["month"]:02d}/{r["day"]:02d}/{r["full_slug"]}/',
         'label': r['label'], 'date': r['date']}
        for r in v._gen_related_path_data(year, month, day, slug)
    ]
    ctx['related_docs'] = [
        {'label': d['label'], 'date': d['date'], 'phase': d['phase'],
         'url': f'/{month:02d}/{d["day"]:02d}/{d["full_slug"]}/'}
        for d in v._gen_related_docs_data(year, month, day, slug)
    ]
    ctx['year_url'] = '/'
    ctx['month_url'] = f'/{month:02d}/'
    ctx['prev_entry_url'] = f'/{month:02d}/{max(day - 1, 1)}/prev/'
    ctx['next_entry_url'] = f'/{month:02d}/{day:02d}/{ctx["next_slug"]}/'
    ctx['export_csv_url'] = f'/{month:02d}/{day:02d}/{slug}/export.csv'
    # archive_era_base.html reads request.get_full_path() directly (Jinja2
    # doesn't get Django's RequestContext auto-injection the way the
    # Django backend does), so a real request object is needed here.
    from django.test import RequestFactory
    ctx['request'] = RequestFactory().get(f'/{month:02d}/{day:02d}/{slug}/')
    return ctx


def _compliance_ctx(year, month, day, slug):
    from apps.honeypot import views as v
    content = v._generate_compliance_content(year, month, day, slug)
    return _archive_ctx_common(v, year, month, day, slug, content)


def _minutes_ctx(year, month, day, slug):
    from apps.honeypot import views as v
    content = v._generate_minutes_content(year, month, day, slug)
    return _archive_ctx_common(v, year, month, day, slug, content)


def _archive_ctx_common(v, year, month, day, slug, content):
    from apps.honeypot.policy_generator import get_cross_policy_stubs
    from apps.presentations.generators import generate_presentations_for_context

    next_slug, prev_slug = v._gen_nav_slugs(year, month, day, slug)
    related_paths = [
        {'url': f'/archive/{r["year"]}/{r["month"]:02d}/{r["day"]:02d}/{r["full_slug"]}/',
         'label': r['label'], 'date': r['date']}
        for r in v._gen_related_path_data(year, month, day, slug)
    ]
    cross_year_reports = v._gen_cross_year_reports(year, month, day, slug)
    related_docs = [
        {'label': d['label'], 'date': d['date'], 'phase': d['phase'],
         'url': f'/archive/{year}/{month:02d}/{d["day"]:02d}/{d["full_slug"]}/'}
        for d in v._gen_related_docs_data(year, month, day, slug)
    ]
    related_policy = get_cross_policy_stubs(year, month, day, slug) or []
    related_presentations = generate_presentations_for_context(
        f'diffrender_pres_{year}_{month}_{day}_{slug}',
        count=v._gen_presentations_count(year, month, day, slug),
    )
    return {
        'year': year, 'month': month, 'day': day, 'slug': slug,
        'on_archive_subdomain': False,
        'next_slug': next_slug, 'prev_slug': prev_slug,
        'related_paths': related_paths, 'cross_year_reports': cross_year_reports,
        'related_docs': related_docs, 'related_policy': related_policy,
        'related_presentations': related_presentations,
        'archive_years': list(range(2025, 1984, -1)), 'all_years': list(range(2025, 1984, -1)),
        'year_url': f'/archive/{year}/', 'month_url': f'/archive/{year}/{month:02d}/',
        'prev_entry_url': f'/archive/{year}/{month:02d}/{max(day - 1, 1)}/prev/',
        'next_entry_url': f'/archive/{year}/{month:02d}/{day:02d}/{next_slug}/',
        'export_csv_url': f'/archive/{year}/{month:02d}/{day:02d}/{slug}/export.csv',
        'parent_template': 'base.html',
        'og_title': content.get('title', 'ACPWB Archive'),
        **content,
    }


def _check_archive_default(stdout):
    from apps.honeypot.pyrender import archive_main
    ctx = _archive_ctx(2024, 3, 15, 'diff-render-default-slug')
    real = render_to_string('honeypot/archive.html', ctx)
    mine_content = archive_main.render_archive_default(ctx)
    mine = render_to_string('honeypot/_archive_content_shell.html', {
        'content_html': mine_content, 'og_title': ctx['og_title'], 'needs_presentations_css': True,
    })
    return _report_diff(stdout, real, mine, 'archive_default')


def _check_archive_default_era(stdout):
    from apps.honeypot.pyrender import archive_era
    ctx = _archive_ctx_era(2024, 3, 15, 'diff-render-era-default-slug', '_generate_archive_content')
    real = render_to_string('honeypot/era/archive.html', ctx)
    mine_era_content = archive_era.render_archive_default_era(ctx)
    og_description = (
        f'{ctx["industry"]} sector engagement documentation archived '
        f'{ctx["year"]}-{ctx["month"]:02d}-{ctx["day"]:02d}. '
        f'{ctx["phase"].capitalize()} phase record. ACPWB Research Division.'
    )
    mine = render_to_string('honeypot/_archive_era_content_shell.html', {
        'title': ctx['title'], 'og_description': og_description, 'era_content_html': mine_era_content,
        'request': ctx['request'], 'year_data': ctx['year_data'],
        'year': ctx['year'], 'all_years': ctx['all_years'],
    })
    return _report_diff(stdout, real, mine, 'archive_default_era')


def _check_archive_compliance(stdout):
    from apps.honeypot.pyrender import archive_main
    # Slug chosen (see the session that built this) to land in the
    # compliance bucket of archive_trap's variant hash.
    ctx = _compliance_ctx(2024, 3, 15, 'article-19')
    real = render_to_string('honeypot/archive_compliance.html', ctx)
    mine_content = archive_main.render_compliance_default(ctx)
    mine = render_to_string('honeypot/_archive_content_shell.html', {
        'content_html': mine_content, 'og_title': ctx['og_title'],
    })
    return _report_diff(stdout, real, mine, 'archive_compliance')


def _check_archive_minutes(stdout):
    from apps.honeypot.pyrender import archive_main
    ctx = _minutes_ctx(2024, 3, 15, 'article-1')
    real = render_to_string('honeypot/archive_minutes.html', ctx)
    mine_content = archive_main.render_minutes_default(ctx)
    mine = render_to_string('honeypot/_archive_content_shell.html', {
        'content_html': mine_content, 'og_title': ctx['og_title'],
    })
    return _report_diff(stdout, real, mine, 'archive_minutes')


def _check_archive_compliance_era(stdout):
    from apps.honeypot.pyrender import archive_era
    ctx = _archive_ctx_era(2024, 3, 15, 'article-19', '_generate_compliance_content')
    real = render_to_string('honeypot/era/archive_compliance.html', ctx)
    mine_era_content = archive_era.render_compliance_default_era(ctx)
    og_description = (
        f'{ctx["industry"]} sector compliance review archived '
        f'{ctx["year"]}-{ctx["month"]:02d}-{ctx["day"]:02d}. Audit ref {ctx["audit_ref"]}. '
        f'ACPWB Regulatory Practice.'
    )
    mine = render_to_string('honeypot/_archive_era_content_shell.html', {
        'title': ctx['title'], 'title_suffix': 'ACPWB Compliance Archive',
        'og_description': og_description, 'era_content_html': mine_era_content,
        'request': ctx['request'], 'year_data': ctx['year_data'],
        'year': ctx['year'], 'all_years': ctx['all_years'],
    })
    return _report_diff(stdout, real, mine, 'archive_compliance_era')


def _check_archive_minutes_era(stdout):
    from apps.honeypot.pyrender import archive_era
    ctx = _archive_ctx_era(2024, 3, 15, 'article-1', '_generate_minutes_content')
    real = render_to_string('honeypot/era/archive_minutes.html', ctx)
    mine_era_content = archive_era.render_minutes_default_era(ctx)
    og_description = (
        f'{ctx["committee"]} meeting minutes archived '
        f'{ctx["year"]}-{ctx["month"]:02d}-{ctx["day"]:02d}. Meeting ref {ctx["meeting_ref"]}. '
        f'ACPWB Institutional Records.'
    )
    mine = render_to_string('honeypot/_archive_era_content_shell.html', {
        'title': ctx['title'], 'og_description': og_description, 'era_content_html': mine_era_content,
        'request': ctx['request'], 'year_data': ctx['year_data'],
        'year': ctx['year'], 'all_years': ctx['all_years'],
    })
    return _report_diff(stdout, real, mine, 'archive_minutes_era')


def _policy_detail_ctx(year, month, day, agency, slug, on_sub=False):
    from django.test import RequestFactory

    from apps.core.context_processors import honeypot_context
    from apps.honeypot import views as v
    from apps.honeypot.policy_generator import (
        generate_policy_document,
        generate_related_links,
        get_cross_archive_stubs,
    )
    from apps.presentations.generators import generate_presentations_for_context

    doc = generate_policy_document(year, month, day, agency, slug)
    related = generate_related_links(year, month, day, agency, slug)
    related_archive = get_cross_archive_stubs(year, month, day, agency, slug)
    related_presentations = generate_presentations_for_context(
        f"policy_pres_{year}_{month}_{day}_{agency}_{slug[:32]}", count=4
    )
    path = f'/public-policy/{year}/{month:02d}/{day:02d}/{agency}/{slug}/'
    request = RequestFactory().get(path)
    request.on_policy_subdomain = on_sub
    nav = v._policy_nav_context(request)
    return {
        'doc': doc,
        'related': related,
        'related_archive': related_archive,
        'related_presentations': related_presentations,
        'policy_years': list(range(2025, 1992, -1)),
        'og_title': f'{doc["title"]} — ACPWB',
        'og_description': doc['summary'][:160],
        'request': request,
        'now_year': __import__('datetime').datetime.now().year,
        **nav,
        **honeypot_context(request),
    }


def _check_policy_detail(stdout):
    from apps.honeypot.pyrender import policy
    ctx = _policy_detail_ctx(2024, 3, 15, 'sec', 'diff-render-policy-detail-slug')
    real = render_to_string('honeypot/public_policy_detail.html', ctx)
    mine = policy.render_policy_detail(ctx)
    return _report_diff(stdout, real, mine, 'policy_detail')


_CASES = {
    'archive_default': _check_archive_default,
    'archive_default_era': _check_archive_default_era,
    'archive_compliance': _check_archive_compliance,
    'archive_minutes': _check_archive_minutes,
    'archive_compliance_era': _check_archive_compliance_era,
    'archive_minutes_era': _check_archive_minutes_era,
    'policy_detail': _check_policy_detail,
}


class Command(BaseCommand):
    help = 'Compare old-template vs new-Python-builder render output for archive/policy pages.'

    def add_arguments(self, parser):
        parser.add_argument('--case', default='all', choices=list(_CASES) + ['all'])

    def handle(self, *args, **options):
        case = options['case']
        cases = list(_CASES) if case == 'all' else [case]
        all_ok = True
        for name in cases:
            ok = _CASES[name](self.stdout)
            all_ok = all_ok and ok
        if not all_ok:
            raise CommandError('One or more cases mismatched — see output above.')
        self.stdout.write(self.style.SUCCESS('\nAll checked cases match.'))
