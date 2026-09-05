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
    return re.sub(r'\s+', '', s)


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


_CASES = {
    'archive_default': _check_archive_default,
    'archive_compliance': _check_archive_compliance,
    'archive_minutes': _check_archive_minutes,
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
