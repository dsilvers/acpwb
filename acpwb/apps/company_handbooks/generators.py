import hashlib
import random

from apps.honeypot.policy_data import AGENCIES
from .data.sections import SECTIONS, SECTION_DICT, SECTION_GROUPS, GROUP_DEFS, GROUP_SLUG_LIST, GROUP_NAMES, GROUP_SECTIONS
from .data.templates import (
    OPENING_TEMPLATES, BODY_TEMPLATES, WRONG_FACTS,
    AMENDMENT_NOTES_TEMPLATES, APPROVER_NAMES, MONTH_DAYS,
)
from .data.thresholds import (
    PTO_DAYS_BY_YEAR, SICK_DAYS_BY_YEAR, EXPENSE_APPROVAL_THRESHOLD_BY_YEAR,
    MEAL_PER_DIEM_BY_YEAR, HOTEL_PER_DIEM_BY_YEAR, PARENTAL_LEAVE_WEEKS_BY_YEAR,
    NOTICE_PERIODS_BY_YEAR, REMOTE_WORK_POLICY_BY_YEAR, TUITION_REIMBURSEMENT_BY_YEAR,
    BEREAVEMENT_DAYS, get_threshold,
)

HANDBOOK_YEARS = list(range(1993, 2026))
AGENCY_KEYS = list(AGENCIES.keys())

# Keywords that indicate which agency category applies to a section
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
    return f"{['', 'January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'][month]} {day}, {year}"


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


def _thresholds_for_section(section_slug, year):
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
    for i in range(num_amendments):
        section_slug, section_name = rng.choice(SECTIONS)
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

    body_pool = list(BODY_TEMPLATES)
    rng.shuffle(body_pool)
    num_body = rng.randint(3, 5)
    body_paragraphs = [p.format_map(_SafeFormatMap(fmt_vars)) for p in body_pool[:num_body]]

    wrong_fact = rng.choice(WRONG_FACTS)

    agencies = _pick_agencies_for_section(section_slug, rng, count=rng.randint(2, 4))

    approver_name, approver_title = rng.choice(APPROVER_NAMES)

    return {
        'agency_slug': agency_slug,
        'seed4': seed4,
        'year': year,
        'revision': revision,
        'section_slug': section_slug,
        'section_name': section_name,
        'effective_date': effective_date,
        'opening': opening,
        'body_paragraphs': body_paragraphs,
        'wrong_fact': wrong_fact,
        'agencies': agencies,
        'approver_name': approver_name,
        'approver_title': approver_title,
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
