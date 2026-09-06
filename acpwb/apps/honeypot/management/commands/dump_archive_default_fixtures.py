"""
Dump (input, context, html) fixtures for the main-domain 'default' archive
page variant, for the Go port's regression tests (acpwb_go/archive).

For each (year, month, day, slug) input that lands in the 'default' variant
(~70% of cases — see _variant_int in archive_trap()), this replicates the
exact context-assembly archive_trap() does (main-domain branch only, i.e.
on_archive_subdomain=False) and renders it with the real
pyrender.archive_main.render_archive_default(), then dumps:
  - the input params
  - the assembled context (JSON, for localizing any Go-port mismatches)
  - the rendered HTML

Usage (inside the web container):
    python manage.py dump_archive_default_fixtures
    docker cp <web-container>:/tmp/archive_fixtures/. ../acpwb_go/archive/testdata/
"""
import hashlib
import json
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot.views import (
    _generate_archive_content, _gen_nav_slugs, _gen_related_path_data,
    _gen_cross_year_reports, _gen_related_docs_data, _gen_presentations_count,
    _archive_url, _year_data,
)
from apps.honeypot.policy_generator import get_cross_policy_stubs
from apps.presentations.generators import generate_presentations_for_context
from apps.honeypot.pyrender.archive_main import render_archive_default


class _FakeRequest:
    """Stand-in for the Django request — _archive_url only reads these two
    attrs (via getattr with defaults), and archive_trap's main-domain branch
    corresponds to both being absent/False."""
    on_archive_subdomain = False
    archive_year = None


def build_context(year, month, day, slug):
    req = _FakeRequest()
    content = _generate_archive_content(year, month, day, slug)

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
        'on_archive_subdomain': False,
        'year_data': yd,
        'all_years': list(range(2025, 1984, -1)),
        'parent_template': 'base.html',
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


# (year, month, day, slug) — hand-picked to vary widely: early/mid/late years,
# every month, slug depths (0/1/2), numeric-suffixed slugs, empty slug.
CANDIDATES = []
_slugs = [
    '', 'q3-earnings-review-1234', 'annual-shareholder-letter-9981',
    'board-minutes-draft-0007', 'compliance-audit-findings-4432',
    'divisional-restructuring-plan-2210', 'internal/memo-5567',
    'legal/settlement-summary-8890', 'a/b/c-nested-path-3321',
    'workforce-reduction-notice-6654', 'benefits-overhaul-proposal-1122',
    'executive-succession-memo-7789', 'vendor-contract-renewal-3345',
    'data-breach-disclosure-9900', 'facilities-closure-notice-2244',
]
_years = [1985, 1990, 1998, 2003, 2008, 2012, 2016, 2019, 2021, 2024]
for i, year in enumerate(_years):
    month = (i % 12) + 1
    day = (i * 3 % 28) + 1
    slug = _slugs[i % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
# A few more explicit combos to broaden coverage past 20 passing cases.
for i in range(20):
    year = 1985 + (i * 7) % 40
    month = (i * 5 % 12) + 1
    day = (i * 11 % 28) + 1
    slug = _slugs[(i * 3) % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
# Third batch: push past 20 passing default-variant cases with more spread
# (leap-year days, year boundaries near the endyear=min(year+.., 2024) clamp,
# deeply nested slugs, month=12/day=28 edges).
for i in range(30):
    year = 1985 + (i * 13) % 41
    month = ((i * 7) % 12) + 1
    day = ((i * 17) % 28) + 1
    slug = _slugs[(i * 5 + 2) % len(_slugs)]
    CANDIDATES.append((year, month, day, slug))
CANDIDATES.append((2024, 12, 28, 'year-end-close-out-9999'))
CANDIDATES.append((2023, 12, 31 - 3, 'a/b/c/d/deep-nested-9999'))
CANDIDATES.append((2000, 2, 29, 'leap-day-filing-1234'))


class Command(BaseCommand):
    help = 'Dump fixtures (input/context/html) for the default archive page variant.'

    def add_arguments(self, parser):
        parser.add_argument('--out', default='/tmp/archive_fixtures')
        parser.add_argument('--limit', type=int, default=60)

    def handle(self, *args, **options):
        out_dir = Path(options['out'])
        out_dir.mkdir(parents=True, exist_ok=True)
        limit = options['limit']

        count = 0
        seen = set()
        for year, month, day, slug in CANDIDATES:
            if count >= limit:
                break
            key = (year, month, day, slug)
            if key in seen:
                continue
            seen.add(key)
            variant = _variant_for(year, month, day, slug)
            if variant < 6:
                self.stdout.write(f'  skip (variant={variant}, not default): {key}')
                continue
            ctx = build_context(year, month, day, slug)
            html = render_archive_default(ctx)

            fixture = {
                'year': year, 'month': month, 'day': day, 'slug': slug,
                'variant_int': variant,
                'html': html,
            }
            name = f'default_{year}_{month:02d}_{day:02d}_{slug.replace("/", "_") or "empty"}.json'
            (out_dir / name).write_text(json.dumps(fixture, ensure_ascii=False))
            self.stdout.write(f'  wrote {name} (variant={variant}, html_len={len(html)})')
            count += 1

        self.stdout.write(self.style.SUCCESS(f'Wrote {count} default-variant fixtures to {out_dir}'))
