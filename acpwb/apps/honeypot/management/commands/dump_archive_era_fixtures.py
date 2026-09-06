"""
Dump (input, html) fixtures for the archive SUBDOMAIN ("era") page variants —
templates/jinja2/honeypot/era/archive.html, archive_compliance.html,
archive_minutes.html, rendered by pyrender/archive_era.py — for the Go port's
regression tests (acpwb_go/archive).

Mirrors dump_archive_compliance_minutes_fixtures.py's approach, but simulates
on_archive_subdomain=True / request.archive_year=year the way
SubdomainMiddleware does for archives-YYYY.acpwb.com, so _archive_url()
produces subdomain-relative / cross-subdomain-absolute URLs instead of
/archive/<year>/... paths.

Usage (inside the web container):
    python manage.py dump_archive_era_fixtures
    docker cp <web-container>:/tmp/archive_era_fixtures/. ../acpwb_go/archive/testdata/
"""
import hashlib
import json
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot.views import (
    _generate_archive_content, _generate_compliance_content, _generate_minutes_content,
    _gen_nav_slugs, _gen_related_path_data,
    _gen_cross_year_reports, _gen_related_docs_data, _gen_presentations_count,
    _archive_url, _year_data,
)
from apps.honeypot.policy_generator import get_cross_policy_stubs
from apps.presentations.generators import generate_presentations_for_context
from apps.honeypot.pyrender.archive_era import (
    render_archive_default_era, render_compliance_default_era, render_minutes_default_era,
)


class _FakeSubRequest:
    """Simulates SubdomainMiddleware's attrs on archives-YYYY.acpwb.com."""
    on_archive_subdomain = True

    def __init__(self, year):
        self.archive_year = year


def build_context(year, month, day, slug, content, need_presentations):
    req = _FakeSubRequest(year)

    next_slug, prev_slug = _gen_nav_slugs(year, month, day, slug)
    prev_day = day - 1 if day > 1 else 28
    prev_month = month if day > 1 else (month - 1 if month > 1 else 12)
    prev_year = year if month > 1 or day > 1 else year - 1

    related_paths = [
        {
            'url': _archive_url(req, r['year'], r['month'], r['day'], r['full_slug']),
            'label': r['label'],
            'date': r['date'],
        }
        for r in _gen_related_path_data(year, month, day, slug)
    ]
    cross_year_reports = _gen_cross_year_reports(year, month, day, slug)
    related_docs = [
        {
            'label': d['label'],
            'url': _archive_url(req, year, month, d['day'], d['full_slug']),
            'date': d['date'],
            'phase': d['phase'],
        }
        for d in _gen_related_docs_data(year, month, day, slug)
    ]
    related_policy = get_cross_policy_stubs(year, month, day, slug)
    related_presentations = []
    if need_presentations:
        related_presentations = generate_presentations_for_context(
            f"archive_pres_{year}_{month}_{day}_{slug[:32]}",
            count=_gen_presentations_count(year, month, day, slug),
        )
    yd = _year_data(year)

    context = {
        'year': year, 'month': month, 'day': day, 'slug': slug,
        'depth': slug.count('/') + 1 if slug else 0,
        'next_slug': next_slug,
        'next_year': year, 'next_month': month, 'next_day': day,
        'prev_year': prev_year, 'prev_month': prev_month, 'prev_day': prev_day,
        'related_paths': related_paths,
        'cross_year_reports': cross_year_reports,
        'archive_years': list(range(2025, 1984, -1)),
        'on_archive_subdomain': True,
        'year_data': yd,
        'all_years': list(range(2025, 1984, -1)),
        'parent_template': 'honeypot/archive_subdomain_base.html',
        'year_url': _archive_url(req, year),
        'month_url': _archive_url(req, year, month),
        'prev_entry_url': _archive_url(req, prev_year, prev_month, prev_day, prev_slug),
        'next_entry_url': _archive_url(req, year, month, day, next_slug),
        'export_csv_url': _archive_url(req, year, month, day, slug) + 'export.csv',
        'related_docs': related_docs,
        'related_policy': related_policy,
        'related_presentations': related_presentations,
        'og_title': content.get('title', 'ACPWB Archive'),
        **content,
    }
    return context


def _variant_for(year, month, day, slug):
    return int(hashlib.md5(f"variant_{year}{month}{day}{slug}".encode()).hexdigest(), 16) % 20


_slugs = [
    '', 'q3-earnings-review-1234', 'annual-shareholder-letter-9981',
    'board-minutes-draft-0007', 'compliance-audit-findings-4432',
    'divisional-restructuring-plan-2210', 'internal/memo-5567',
    'legal/settlement-summary-8890', 'a/b/c-nested-path-3321',
    'workforce-reduction-notice-6654', 'benefits-overhaul-proposal-1122',
    'executive-succession-memo-7789', 'vendor-contract-renewal-3345',
    'data-breach-disclosure-9900', 'facilities-closure-notice-2244',
]

# Wide spread of (year, month, day, slug) combos, including year-boundary
# days (month=1/day=1) so prev_entry_url's cross-subdomain-year branch of
# _archive_url gets exercised. The command filters by variant int and only
# keeps combos landing in the target variant's bucket.
CANDIDATES = []
_years = [1985, 1990, 1998, 2003, 2008, 2012, 2016, 2019, 2021, 2024]
for i, year in enumerate(_years):
    month = (i % 12) + 1
    day = (i * 3 % 28) + 1
    slug = _slugs[i % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
for i in range(90):
    year = 1985 + (i * 7) % 40
    month = (i * 5 % 12) + 1
    day = (i * 11 % 28) + 1
    slug = _slugs[(i * 3) % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
for i in range(90):
    year = 1985 + (i * 13) % 41
    month = ((i * 7) % 12) + 1
    day = ((i * 17) % 28) + 1
    slug = _slugs[(i * 5 + 2) % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
for i in range(90):
    year = 1985 + (i * 17) % 41
    month = ((i * 3) % 12) + 1
    day = ((i * 23) % 28) + 1
    slug = _slugs[(i * 2 + 1) % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
# Year-boundary edge cases: month=1, day=1 => prev entry crosses into the
# PRIOR year's subdomain (cross-subdomain absolute URL for prev_entry_url).
CANDIDATES.append((2024, 1, 1, 'new-year-filing-0001'))
CANDIDATES.append((1986, 1, 1, ''))
CANDIDATES.append((2000, 1, 1, 'y2k-leap-year-review-1234'))
CANDIDATES.append((2024, 12, 28, 'year-end-close-out-9999'))
CANDIDATES.append((2023, 12, 28, 'a/b/c/d/deep-nested-9999'))
CANDIDATES.append((2000, 2, 29, 'leap-day-filing-1234'))
CANDIDATES.append((1985, 1, 1, ''))
CANDIDATES.append((2025, 6, 15, 'mid-decade-review-4567'))


class Command(BaseCommand):
    help = 'Dump fixtures (input/html) for the archive subdomain ("era") page variants.'

    def add_arguments(self, parser):
        parser.add_argument('--out', default='/tmp/archive_era_fixtures')
        parser.add_argument('--limit-each', type=int, default=35)

    def handle(self, *args, **options):
        out_dir = Path(options['out'])
        out_dir.mkdir(parents=True, exist_ok=True)
        limit_each = options['limit_each']

        counts = {'default': 0, 'compliance': 0, 'minutes': 0}
        seen = set()
        for year, month, day, slug in CANDIDATES:
            if all(counts[k] >= limit_each for k in counts):
                break
            key = (year, month, day, slug)
            if key in seen:
                continue
            seen.add(key)
            variant = _variant_for(year, month, day, slug)

            if variant < 3 and counts['compliance'] < limit_each:
                variant_name = 'compliance'
                content = _generate_compliance_content(year, month, day, slug)
                ctx = build_context(year, month, day, slug, content, need_presentations=False)
                html = render_compliance_default_era(ctx)
            elif 3 <= variant < 6 and counts['minutes'] < limit_each:
                variant_name = 'minutes'
                content = _generate_minutes_content(year, month, day, slug)
                ctx = build_context(year, month, day, slug, content, need_presentations=False)
                html = render_minutes_default_era(ctx)
            elif variant >= 6 and counts['default'] < limit_each:
                variant_name = 'default'
                content = _generate_archive_content(year, month, day, slug)
                ctx = build_context(year, month, day, slug, content, need_presentations=True)
                html = render_archive_default_era(ctx)
            else:
                self.stdout.write(f'  skip (variant={variant}, bucket full): {key}')
                continue

            fixture = {
                'variant': variant_name,
                'year': year, 'month': month, 'day': day, 'slug': slug,
                'variant_int': variant,
                'html': html,
            }
            name = f'era_{variant_name}_{year}_{month:02d}_{day:02d}_{slug.replace("/", "_") or "empty"}.json'
            (out_dir / name).write_text(json.dumps(fixture, ensure_ascii=False))
            self.stdout.write(f'  wrote {name} (variant={variant}, html_len={len(html)})')
            counts[variant_name] += 1

        self.stdout.write(self.style.SUCCESS(
            f'Wrote {counts["default"]} default + {counts["compliance"]} compliance + '
            f'{counts["minutes"]} minutes era fixtures to {out_dir}'
        ))
