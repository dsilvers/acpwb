"""
Dump (input, context-lite, html) fixtures for all 6 public-policy page
template variants, for the Go port's regression tests (acpwb_go/policy).

Mirrors the pattern in dump_archive_default_fixtures.py: replicate the
context-assembly each views.py policy view function does (without touching
Redis/DB via _log_crawler), render with the real pyrender.policy functions,
and dump the final HTML (plus enough of the input to replay it in Go).

honeypot_token/site_root are the one known non-deterministic exemption (see
apps/core/context_processors.py:honeypot_context, which hashes in
time.time()) — fixtures use a FIXED honeypot_token so Go's replay can supply
the same fixed value and get a byte-identical comparison for everything else.

Usage (inside the web container):
    python manage.py dump_policy_fixtures
    docker cp <web-container>:/tmp/policy_fixtures/. ../acpwb_go/policy/testdata/
"""
import json
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot.policy_generator import (
    generate_policy_document, generate_related_links, get_cross_archive_stubs,
    get_policy_index_years, get_policy_year_data, get_policy_year_months,
    get_policy_month_entries, get_policy_agency_years, get_policy_agency_year_detail,
    get_policy_agency_month_entries,
)
from apps.honeypot.views import _policy_nav_context, _policy_url
from apps.honeypot.policy_data import AGENCIES
from apps.presentations.generators import generate_presentations_for_context
from apps.honeypot.pyrender import policy as pyrender_policy

FIXED_TOKEN = 'fixedtok01'


class _FakeRequest:
    def __init__(self, path, on_sub=False, agency_slug=None):
        self._path = path
        self.on_policy_subdomain = on_sub
        self.policy_agency_slug = agency_slug

    def get_full_path(self):
        return self._path


def _ctx_base(request):
    return {'honeypot_token': FIXED_TOKEN, 'site_root': 'https://acpwb.com' if request.on_policy_subdomain else '', 'request': request}


# ── policy_index ──────────────────────────────────────────────────────────────

def build_index_fixture():
    request = _FakeRequest('/public-policy/')
    ctx = {
        'years': get_policy_index_years(),
        'og_title': 'Public Policy — ACPWB',
        'og_description': 'ACPWB public policy positions, regulatory comment letters, and legislative testimony on compensation, labor, and corporate governance.',
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_index(ctx)
    return {'html': html}


# ── policy_year ───────────────────────────────────────────────────────────────

def build_year_fixture(year):
    request = _FakeRequest(f'/public-policy/{year}/')
    ctx = {
        'year': year,
        'year_data': get_policy_year_data(year),
        'months': get_policy_year_months(year),
        'policy_years': list(range(2025, 1992, -1)),
        'prev_year': year - 1,
        'next_year': year + 1,
        'og_title': f'{year} Public Policy — ACPWB',
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_year(ctx)
    return {'year': year, 'html': html}


# ── policy_month (main domain + subdomain) ───────────────────────────────────

def build_month_fixture(year, month, on_sub=False, agency=None):
    if on_sub:
        request = _FakeRequest(f'/{year}/{month:02d}/', on_sub=True, agency_slug=agency)
        entries = get_policy_agency_month_entries(agency, year, month, url_fn=lambda y, m, d, ag, sl: _policy_url(request, y, m, d, ag, sl))
        og_title = f'{agency.upper()} Policy {year}-{month:02d} — ACPWB'
    else:
        request = _FakeRequest(f'/public-policy/{year}/{month:02d}/')
        entries = get_policy_month_entries(year, month)
        og_title = f'Public Policy {year}-{month:02d} — ACPWB'

    prev_month = month - 1 if month > 1 else 12
    prev_year = year if month > 1 else year - 1
    next_month = month + 1 if month < 12 else 1
    next_year = year if month < 12 else year + 1
    nav = _policy_nav_context(request)
    ctx = {
        'year': year,
        'month': month,
        'entries': entries,
        'policy_years': list(range(2025, 1992, -1)),
        'prev_year': prev_year, 'prev_month': prev_month,
        'next_year': next_year, 'next_month': next_month,
        'year_url': nav['policy_year_url'](year),
        'prev_month_url': nav['policy_month_url'](prev_year, prev_month),
        'next_month_url': nav['policy_month_url'](next_year, next_month),
        'og_title': og_title,
        **nav,
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_month(ctx)
    return {'year': year, 'month': month, 'on_sub': on_sub, 'agency': agency, 'html': html}


# ── policy_detail (main domain + subdomain) ─────────────────────────────────

def build_detail_fixture(year, month, day, agency, slug, on_sub=False):
    doc = generate_policy_document(year, month, day, agency, slug)
    if on_sub:
        request = _FakeRequest(f'/{year}/{month:02d}/{day:02d}/{slug}/', on_sub=True, agency_slug=agency)
        url_fn = lambda y, m, d, ag, sl: _policy_url(request, y, m, d, ag, sl)
        doc = {**doc, 'url': url_fn(year, month, day, agency, slug)}
        related = generate_related_links(year, month, day, agency, slug, url_fn=url_fn)
        related_presentations = None
    else:
        request = _FakeRequest(f'/public-policy/{year}/{month:02d}/{day:02d}/{agency}/{slug}/')
        related = generate_related_links(year, month, day, agency, slug, url_fn=None)
        related_presentations = generate_presentations_for_context(
            f"policy_pres_{year}_{month}_{day}_{agency}_{slug[:32]}", count=4
        )
    related_archive = get_cross_archive_stubs(year, month, day, agency, slug)
    nav = _policy_nav_context(request)
    ctx = {
        'doc': doc,
        'related': related,
        'related_archive': related_archive,
        'related_presentations': related_presentations,
        'policy_years': list(range(2025, 1992, -1)),
        'og_title': f'{doc["title"]} — ACPWB',
        'og_description': doc['summary'][:160],
        **nav,
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_detail(ctx)
    return {'year': year, 'month': month, 'day': day, 'agency': agency, 'slug': slug, 'on_sub': on_sub, 'html': html}


# ── policy_subdomain_index ────────────────────────────────────────────────────

def build_subdomain_index_fixture(agency):
    from apps.honeypot.policy_data import AGENCIES
    request = _FakeRequest('/', on_sub=True, agency_slug=agency)
    agency_data = AGENCIES.get(agency, ('Unknown Agency', 'regulatory policy'))
    agency_full, policy_domain = agency_data
    nav = _policy_nav_context(request)
    ctx = {
        'agency': agency,
        'agency_full': agency_full,
        'policy_domain': policy_domain,
        'years': get_policy_agency_years(agency),
        'og_title': f'{agency.upper()} Policy Filings — ACPWB',
        'og_description': f'ACPWB regulatory filings, comment letters, and testimony submitted to the {agency_full}.',
        **nav,
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_subdomain_index(ctx)
    return {'agency': agency, 'html': html}


# ── policy_subdomain_year ─────────────────────────────────────────────────────

def build_subdomain_year_fixture(agency, year):
    from apps.honeypot.policy_data import AGENCIES
    request = _FakeRequest(f'/{year}/', on_sub=True, agency_slug=agency)
    agency_data = AGENCIES.get(agency, ('Unknown Agency', 'regulatory policy'))
    agency_full, policy_domain = agency_data
    nav = _policy_nav_context(request)
    all_years = get_policy_agency_years(agency)
    year_detail = get_policy_agency_year_detail(agency, year)
    ctx = {
        'agency': agency,
        'agency_full': agency_full,
        'policy_domain': policy_domain,
        'year': year,
        'year_detail': year_detail,
        'all_years': all_years,
        'prev_year': year - 1,
        'next_year': year + 1,
        'og_title': f'{year} {agency.upper()} Policy Filings — ACPWB',
        **nav,
        **_ctx_base(request),
    }
    html = pyrender_policy.render_policy_subdomain_year(ctx)
    return {'agency': agency, 'year': year, 'html': html}


AGENTS = ['sec', 'dol', 'ftc', 'eeoc', 'nlrb', 'irs', 'osha', 'epa', 'fcc', 'hud',
          'ca-dlse', 'ny-dol', 'tx-twc', 'uk-fca', 'jp-fsa', 'wi-dwd', 'finra', 'cfpb']
YEARS = [1993, 1998, 2002, 2006, 2011, 2015, 2019, 2022, 2024, 2025,
         2023, 2010, 2004, 1996, 2018, 2001, 2020, 2013]
SLUGS = [
    'executive-compensation-disclosure-requirements', 'ceo-pay-ratio-disclosure-rule',
    'overtime-threshold-adjustment', 'noncompete-agreement-enforcement-guidelines',
    'pay-equity-reporting-standards', 'collective-bargaining-unit-determination',
    'gig-worker-classification-standards', 'remote-work-tax-nexus-rules',
    'paid-family-leave-mandate', 'severance-pay-tax-treatment',
    'stock-option-accounting-transparency', 'proxy-statement-reform',
]


class Command(BaseCommand):
    help = 'Dump fixtures (input/html) for all 6 public-policy page template variants.'

    def add_arguments(self, parser):
        parser.add_argument('--out', default='/tmp/policy_fixtures')

    def handle(self, *args, **options):
        out_dir = Path(options['out'])
        out_dir.mkdir(parents=True, exist_ok=True)
        counts = {}

        def write(subdir, name, fixture):
            d = out_dir / subdir
            d.mkdir(parents=True, exist_ok=True)
            (d / f'{name}.json').write_text(json.dumps(fixture, ensure_ascii=False))
            counts[subdir] = counts.get(subdir, 0) + 1

        # policy_index: only one real input shape, but content depends on
        # the (non-parameterized) year list generator — dump a few times to
        # confirm stability, then once is enough since it's deterministic.
        write('index', 'index_1', build_index_fixture())

        for i, year in enumerate(sorted(set(YEARS + [1993, 2025, 2000, 2016] + list(range(1994, 2024, 3))))):
            write('year', f'year_{year}', build_year_fixture(year))

        month_pairs = []
        for i in range(30):
            y = 1993 + (i * 7) % 33
            m = (i * 5) % 12 + 1
            month_pairs.append((y, m))
        for y, m in sorted(set(month_pairs)):
            write('month_main', f'month_{y}_{m:02d}', build_month_fixture(y, m, on_sub=False))
        for i, (y, m) in enumerate(sorted(set(month_pairs))[:20]):
            ag = AGENTS[i % len(AGENTS)]
            write('month_sub', f'month_sub_{ag}_{y}_{m:02d}', build_month_fixture(y, m, on_sub=True, agency=ag))

        detail_inputs = []
        for i in range(40):
            y = YEARS[i % len(YEARS)]
            m = (i * 3) % 12 + 1
            d = (i * 11) % 28 + 1
            ag = AGENTS[i % len(AGENTS)]
            sl = SLUGS[i % len(SLUGS)]
            detail_inputs.append((y, m, d, ag, sl))
        # Featured-seed years explicitly, since those years often have
        # special-cased CEO-letter content elsewhere and are worth covering.
        featured_years = [2024, 2023, 2022, 2021, 2020]
        for i, y in enumerate(featured_years):
            detail_inputs.append((y, (i % 12) + 1, (i * 5) % 28 + 1, AGENTS[i % len(AGENTS)], SLUGS[i % len(SLUGS)]))
        seen = set()
        for y, m, d, ag, sl in detail_inputs:
            key = (y, m, d, ag, sl)
            if key in seen:
                continue
            seen.add(key)
            write('detail_main', f'detail_{y}_{m:02d}_{d:02d}_{ag}_{sl[:20]}', build_detail_fixture(y, m, d, ag, sl, on_sub=False))
        for i, (y, m, d, ag, sl) in enumerate(list(seen)[:20]):
            write('detail_sub', f'detail_sub_{ag}_{y}_{m:02d}_{d:02d}_{sl[:20]}', build_detail_fixture(y, m, d, ag, sl, on_sub=True))

        for ag in AGENTS:
            write('sub_index', f'subindex_{ag}', build_subdomain_index_fixture(ag))

        for i, ag in enumerate(AGENTS):
            y = YEARS[i % len(YEARS)]
            write('sub_year', f'subyear_{ag}_{y}', build_subdomain_year_fixture(ag, y))
            # a couple more years per agency to broaden coverage
            y2 = YEARS[(i + 5) % len(YEARS)]
            if y2 != y:
                write('sub_year', f'subyear_{ag}_{y2}', build_subdomain_year_fixture(ag, y2))

        for k, v in counts.items():
            self.stdout.write(f'  {k}: {v} fixtures')
        self.stdout.write(self.style.SUCCESS(f'Wrote fixtures to {out_dir}'))
