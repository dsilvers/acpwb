import hashlib
import random

from apps.honeypot.policy_data import AGENCIES
from .data.sections import SECTIONS, SECTION_DICT, GROUP_DEFS, GROUP_SLUG_LIST, GROUP_NAMES, GROUP_SECTIONS
from .data.templates import (
    OPENING_TEMPLATES, BODY_TEMPLATES, CONTEXT_FACTS,
    AMENDMENT_NOTES_TEMPLATES, APPROVER_NAMES, MONTH_DAYS,
    CLOSING_TEMPLATES, SECTION_SPOTLIGHT_FACTS, REGULATORY_CITATIONS,
    DEFINITIONS, POLICY_NOTES, SUBSECTION_HEADER_SETS, _GROUP_CODES,
)
from .data.thresholds import (
    PTO_DAYS_BY_YEAR, SICK_DAYS_BY_YEAR, EXPENSE_APPROVAL_THRESHOLD_BY_YEAR,
    MEAL_PER_DIEM_BY_YEAR, HOTEL_PER_DIEM_BY_YEAR, PARENTAL_LEAVE_WEEKS_BY_YEAR,
    NOTICE_PERIODS_BY_YEAR, REMOTE_WORK_POLICY_BY_YEAR, TUITION_REIMBURSEMENT_BY_YEAR,
    BEREAVEMENT_DAYS, get_threshold,
)

HANDBOOK_YEARS = list(range(1993, 2026))
AGENCY_KEYS = list(AGENCIES.keys())

_SECTION_AGENCY_KEYWORDS = {
    'financial': ['sec', 'cftc', 'finra', 'fdic', 'occ', 'frb', 'cfpb', 'pcaob'],
    'labor':     ['dol', 'nlrb', 'whd', 'ofccp', 'fmcs', 'flra', 'bls', 'dol-eta'],
    'safety':    ['osha', 'oshrc', 'niosh', 'epa', 'dot-fmcsa'],
    'health':    ['hhs', 'cms', 'fda', 'ebsa', 'eeoc'],
    'equity':    ['eeoc', 'ofccp', 'doj-crt', 'hud-fheo', 'ed-ocr'],
    'data':      ['ftc', 'sec', 'nist', 'cisa', 'dhs'],
    'legal':     ['doj-civil', 'doj-crt', 'sec', 'ftc'],
    'benefits':  ['ebsa', 'dol', 'irs', 'hhs', 'pbgc'],
    'export':    ['bis', 'treasury-ofac', 'state-pm', 'ddtc'],
    'tax':       ['irs', 'treasury-do', 'treasury-ofac'],
}

_SECTION_TO_AGENCY_CATEGORY = {
    'payroll-procedures':    'labor',
    'overtime-policy':       'labor',
    'pto-leave':             'labor',
    'sick-leave':            'labor',
    'parental-leave':        'labor',
    'bereavement-leave':     'labor',
    'jury-duty':             'labor',
    'military-leave':        'labor',
    'fmla':                  'labor',
    'anti-harassment':       'equity',
    'equal-opportunity':     'equity',
    'diversity-inclusion':   'equity',
    'accommodation-disability': 'equity',
    'benefits-overview':     'benefits',
    'health-insurance':      'health',
    'dental-vision':         'health',
    'retirement-401k':       'benefits',
    'equity-compensation':   'financial',
    'insider-trading':       'financial',
    'safety-security':       'safety',
    'emergency-procedures':  'safety',
    'workplace-violence':    'safety',
    'ergonomics':            'safety',
    'first-aid':             'safety',
    'data-privacy':          'data',
    'cybersecurity':         'data',
    'technology-use':        'data',
    'export-controls':       'export',
    'anti-corruption':       'legal',
    'whistleblower':         'legal',
    'conflict-of-interest':  'legal',
    'drug-alcohol':          'safety',
}

# Build a lookup from section_slug -> (group_slug, position_in_group)
_SECTION_GROUP_MAP = {}
for _g_slug, _g_name, _g_sections in GROUP_DEFS:
    for _pos, (_s_slug, _s_name) in enumerate(_g_sections, start=1):
        _SECTION_GROUP_MAP[_s_slug] = (_g_slug, _pos)


class _SafeFormatMap(dict):
    def __missing__(self, key):
        return f'{{{key}}}'


def _rng_from_seed(seed_str):
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


def _watermark(agency_slug, seed4, year, revision, section):
    raw = f"acpwb_handbook_{agency_slug}_{seed4}_{year}_rev{revision}_{section}"
    return hashlib.md5(raw.encode()).hexdigest()[:8]


def _effective_date(year, revision):
    idx = (revision - 1) % len(MONTH_DAYS)
    month, day = MONTH_DAYS[idx]
    months = ['', 'January', 'February', 'March', 'April', 'May', 'June',
              'July', 'August', 'September', 'October', 'November', 'December']
    return f"{months[month]} {day}, {year}"


def _agency_display(agency_slug):
    if agency_slug in AGENCIES:
        name, domain = AGENCIES[agency_slug]
        return {'slug': agency_slug, 'name': name, 'domain': domain}
    return None


def _pick_agencies_for_section(section_slug, rng, count=3):
    category = _SECTION_TO_AGENCY_CATEGORY.get(section_slug, 'labor')
    candidates = _SECTION_AGENCY_KEYWORDS.get(category, _SECTION_AGENCY_KEYWORDS['labor'])
    valid = [k for k in candidates if k in AGENCIES]
    if not valid:
        valid = AGENCY_KEYS[:20]
    rng.shuffle(valid)
    return [_agency_display(k) for k in valid[:count]]


def _thresholds_for_section(_, year):
    return {
        'pto_days':          get_threshold(PTO_DAYS_BY_YEAR, year),
        'sick_days':         get_threshold(SICK_DAYS_BY_YEAR, year),
        'expense_threshold': get_threshold(EXPENSE_APPROVAL_THRESHOLD_BY_YEAR, year),
        'meal_per_diem':     get_threshold(MEAL_PER_DIEM_BY_YEAR, year),
        'hotel_per_diem':    get_threshold(HOTEL_PER_DIEM_BY_YEAR, year),
        'parental_weeks':    get_threshold(PARENTAL_LEAVE_WEEKS_BY_YEAR, year),
        'notice_period':     get_threshold(NOTICE_PERIODS_BY_YEAR, year),
        'remote_policy':     get_threshold(REMOTE_WORK_POLICY_BY_YEAR, year),
        'tuition_max':       get_threshold(TUITION_REIMBURSEMENT_BY_YEAR, year),
        'bereavement_immediate': BEREAVEMENT_DAYS['immediate'],
    }


def _agency_name(agency_slug):
    return AGENCIES[agency_slug][0] if agency_slug in AGENCIES else agency_slug


def _policy_number(section_slug, year, revision):
    group_slug, pos = _SECTION_GROUP_MAP.get(section_slug, ('hr', 0))
    code = _GROUP_CODES.get(group_slug, 'HR')
    # Derive a stable 4-digit sequence from slug hash so numbers look real
    seq_hash = int(hashlib.md5(section_slug.encode()).hexdigest(), 16) % 900 + 100
    return f"HR.{code}.{seq_hash:04d}-{year}-R{revision}"


def _build_subsections(body_pool, header_set, fmt_vars, rng):
    """Distribute body paragraphs across subsection headers."""
    pool = list(body_pool)
    rng.shuffle(pool)
    subsections = []
    used = 0
    for i, header in enumerate(header_set):
        remaining_headers = len(header_set) - i
        remaining_paras = len(pool) - used
        # Give each subsection at least 1 paragraph; last one gets remainder
        if remaining_headers == 1:
            n = max(1, remaining_paras)
        else:
            n = rng.randint(1, max(1, remaining_paras - remaining_headers + 1))
        paras = [p.format_map(_SafeFormatMap(fmt_vars)) for p in pool[used:used + n]]
        used += n
        subsections.append({'header': header, 'paragraphs': paras})
        if used >= len(pool):
            break
    return subsections


def _build_threshold_table(thresholds):
    rows = [
        ("PTO (Annual)", f"{thresholds['pto_days']} days"),
        ("Sick Leave (Annual)", f"{thresholds['sick_days']} days"),
        ("Expense Pre-Approval", f"${thresholds['expense_threshold']:,}+"),
        ("Meal Per Diem", f"${thresholds['meal_per_diem']}/day"),
        ("Hotel Per Diem", f"${thresholds['hotel_per_diem']}/night"),
        ("Parental Leave", f"{thresholds['parental_weeks']} weeks paid"),
        ("Tuition Reimbursement", f"${thresholds['tuition_max']:,}/year"),
        ("Bereavement (Immediate)", f"{thresholds['bereavement_immediate']} days"),
        ("Resignation Notice", thresholds['notice_period']),
    ]
    return rows


def generate_handbook_index(agency_slug, seed4):
    """Return metadata for an agency-instance handbook index page."""
    if agency_slug not in AGENCIES:
        return None
    agency_name, agency_domain = AGENCIES[agency_slug]
    rng = _rng_from_seed(f"handbook_index_{agency_slug}_{seed4}")
    years = sorted(HANDBOOK_YEARS, reverse=True)
    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'agency_name': agency_name,
        'agency_domain': agency_domain,
        'years': years,
        'revision_count': rng.randint(3, 8),
    }


def generate_year_index(agency_slug, seed4, year):
    """Return revision list for a given agency-instance + year."""
    rng = _rng_from_seed(f"handbook_year_{agency_slug}_{seed4}_{year}")
    num_revisions = rng.randint(2, 6)
    revisions = []
    for rev in range(1, num_revisions + 1):
        revisions.append({
            'number': rev,
            'effective_date': _effective_date(year, rev),
        })
    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'agency_name': _agency_name(agency_slug),
        'year': year,
        'revisions': revisions,
    }


def generate_revision_toc(agency_slug, seed4, year, revision):
    """Return a revision's table of contents with amendment notes."""
    seed = f"handbook_rev_{agency_slug}_{seed4}_{year}_{revision}"
    rng = _rng_from_seed(seed)

    effective_date = _effective_date(year, revision)
    approver_name, approver_title = rng.choice(APPROVER_NAMES)

    num_amendments = rng.randint(2, 5)
    amendments = []
    for _ in range(num_amendments):
        _, section_name = rng.choice(SECTIONS)
        note_tmpl = rng.choice(AMENDMENT_NOTES_TEMPLATES)
        note = note_tmpl.format(
            section_name=section_name,
            effective_date=effective_date,
            year=year,
        )
        amendments.append({'section': section_name, 'note': note})

    all_groups = [(g[0], g[1]) for g in GROUP_DEFS]
    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'agency_name': _agency_name(agency_slug),
        'year': year,
        'revision': revision,
        'effective_date': effective_date,
        'approver_name': approver_name,
        'approver_title': approver_title,
        'amendments': amendments,
        'all_groups': all_groups,
    }


def generate_section(agency_slug, seed4, year, revision, section_slug):
    """Generate full content for a single handbook section."""
    if section_slug not in SECTION_DICT:
        return None

    seed = f"acpwb_handbook_{agency_slug}_{seed4}_{year}_rev{revision}_{section_slug}"
    rng = _rng_from_seed(seed)

    section_name = SECTION_DICT[section_slug]
    effective_date = _effective_date(year, revision)
    token = _watermark(agency_slug, seed4, year, revision, section_slug)
    thresholds = _thresholds_for_section(section_slug, year)

    agency_name = _agency_name(agency_slug)
    fmt_vars = dict(section_name=section_name, effective_date=effective_date, org_name=agency_name, **thresholds)

    opening = rng.choice(OPENING_TEMPLATES).format_map(_SafeFormatMap(fmt_vars))

    # Subsections: pick a header set, shuffle a larger body pool, distribute
    header_set = rng.choice(SUBSECTION_HEADER_SETS)
    body_pool = list(BODY_TEMPLATES)
    n_body = rng.randint(5, 8)
    rng.shuffle(body_pool)
    body_pool = body_pool[:n_body]
    subsections = _build_subsections(body_pool, header_set, fmt_vars, rng)

    # Policy notes: 1-2 callout boxes inserted at random subsection boundaries
    note_pool = list(POLICY_NOTES)
    rng.shuffle(note_pool)
    n_notes = rng.randint(1, 2)
    policy_notes = [{'type': t, 'text': txt} for t, txt in note_pool[:n_notes]]

    # Insert note index: after which subsection to show the first note
    note_after_idx = rng.randint(0, max(0, len(subsections) - 2))

    context_fact = rng.choice(CONTEXT_FACTS)
    spotlight_fact = rng.choice(SECTION_SPOTLIGHT_FACTS)
    regulatory_citation = rng.choice(REGULATORY_CITATIONS)

    agencies = _pick_agencies_for_section(section_slug, rng, count=rng.randint(2, 4))

    approver_name, approver_title = rng.choice(APPROVER_NAMES)

    # Definitions: pick 4-5 from the pool
    def_pool = list(DEFINITIONS)
    rng.shuffle(def_pool)
    n_defs = rng.randint(4, 5)
    definitions = [{'term': t, 'definition': d} for t, d in def_pool[:n_defs]]

    # Amendment history: show up to 3 prior revisions + current
    history = []
    num_prior = min(revision - 1, rng.randint(2, 3))
    for j in range(num_prior, 0, -1):
        prior_rev = revision - j
        prior_date = _effective_date(year, prior_rev)
        note_tmpl = rng.choice(AMENDMENT_NOTES_TEMPLATES)
        note = note_tmpl.format(section_name=section_name, effective_date=prior_date, year=year)
        history.append({'version': f'Rev. {prior_rev}', 'date': prior_date, 'summary': note})
    history.append({
        'version': f'Rev. {revision} (Current)',
        'date': effective_date,
        'summary': 'Current revision. See the Revision Table of Contents for full amendment notes.',
    })

    policy_number = _policy_number(section_slug, year, revision)
    threshold_table = _build_threshold_table(thresholds)

    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'year': year,
        'revision': revision,
        'section_slug': section_slug,
        'section_name': section_name,
        'effective_date': effective_date,
        'opening': opening,
        'body_paragraphs': [p for sub in subsections for p in sub['paragraphs']],  # kept for compat
        'subsections': subsections,
        'policy_notes': policy_notes,
        'note_after_idx': note_after_idx,
        'context_fact': context_fact,
        'spotlight_fact': spotlight_fact,
        'regulatory_citation': regulatory_citation,
        'agencies': agencies,
        'approver_name': approver_name,
        'approver_title': approver_title,
        'definitions': definitions,
        'amendment_history': history,
        'policy_number': policy_number,
        'threshold_table': threshold_table,
        'thresholds': thresholds,
        'watermark': token,
        'sections': SECTIONS,
    }


HANDBOOK_YEARS = list(range(2025, 1992, -1))


def generate_group_page(agency_slug, seed4, year, revision, group_slug):
    """Generate all section content for one handbook chapter (group)."""
    if group_slug not in GROUP_SECTIONS:
        return None
    name = _agency_name(agency_slug)
    group_section_list = GROUP_SECTIONS[group_slug]
    sections_content = []
    for s_slug, _ in group_section_list:
        data = generate_section(agency_slug, seed4, year, revision, s_slug)
        if data:
            sections_content.append(data)
    idx = GROUP_SLUG_LIST.index(group_slug)
    all_groups = [(g[0], g[1]) for g in GROUP_DEFS]

    rel_rng = _rng_from_seed(f"related_{agency_slug}_{seed4}_{year}_{revision}_{group_slug}")
    num_related = rel_rng.randint(4, 7)
    related = []
    seen = set()
    for _ in range(num_related * 3):
        if len(related) >= num_related:
            break
        r_slug = rel_rng.choice(AGENCY_KEYS)
        r_seed4 = f"{rel_rng.randint(1000, 9999):04d}"
        r_year = rel_rng.choice(HANDBOOK_YEARS)
        r_rev = rel_rng.randint(1, 4)
        key = (r_slug, r_seed4)
        if key not in seen and r_slug != agency_slug:
            seen.add(key)
            related.append({
                'agency_slug': r_slug,
                'seed4': r_seed4,
                'year': r_year,
                'revision': r_rev,
                'agency_name': _agency_name(r_slug),
            })

    # Group-level sidebar data
    thresholds = _thresholds_for_section(None, year)
    threshold_table = _build_threshold_table(thresholds)

    grp_rng = _rng_from_seed(f"group_sidebar_{agency_slug}_{seed4}_{year}_{revision}_{group_slug}")
    group_spotlight = grp_rng.choice(SECTION_SPOTLIGHT_FACTS)

    # Pick agencies representative of this group's first section
    first_slug = group_section_list[0][0] if group_section_list else None
    group_agencies = _pick_agencies_for_section(first_slug or '', grp_rng, count=3) if first_slug else []

    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'agency_name': name,
        'year': year,
        'revision': revision,
        'group_slug': group_slug,
        'group_name': GROUP_NAMES[group_slug],
        'sections': sections_content,
        'all_groups': all_groups,
        'prev_group': GROUP_SLUG_LIST[idx - 1] if idx > 0 else None,
        'prev_group_name': GROUP_NAMES[GROUP_SLUG_LIST[idx - 1]] if idx > 0 else None,
        'next_group': GROUP_SLUG_LIST[idx + 1] if idx < len(GROUP_SLUG_LIST) - 1 else None,
        'next_group_name': GROUP_NAMES[GROUP_SLUG_LIST[idx + 1]] if idx < len(GROUP_SLUG_LIST) - 1 else None,
        'watermark': _watermark(agency_slug, seed4, year, revision, group_slug),
        'effective_date': _effective_date(year, revision),
        'related_handbooks': related,
        'threshold_table': threshold_table,
        'group_spotlight': group_spotlight,
        'group_agencies': group_agencies,
    }


def generate_agency_index_page(page=1, per_page=20):
    """Return a page of agency listings with year links using a seeded instance per agency."""
    agency_list = AGENCY_KEYS
    start = (page - 1) * per_page
    end = start + per_page
    page_agencies = agency_list[start:end]
    total_pages = (len(agency_list) + per_page - 1) // per_page

    items = []
    for slug in page_agencies:
        name, domain = AGENCIES[slug]
        rng = _rng_from_seed(f"handbook_instances_{slug}")
        num_instances = rng.randint(3, 5)
        instance_seeds = [f"{rng.randint(1000, 9999):04d}" for _ in range(num_instances)]
        items.append({
            'slug': slug,
            'name': name,
            'domain': domain,
            'primary_seed': instance_seeds[0],
            'instances': instance_seeds,
            'years': HANDBOOK_YEARS,
        })

    return {
        'items': items,
        'page': page,
        'total_pages': total_pages,
        'has_prev': page > 1,
        'has_next': page < total_pages,
    }
