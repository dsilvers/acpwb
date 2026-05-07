"""
Deterministic public policy document generator for ACPWB.
Seed: year + month + day + agency acronym + slug — same inputs always return same output.
"""
import datetime
import hashlib
import random

from apps.people.generators import FIRST_NAMES as _FIRST_NAMES, LAST_NAMES as _LAST_NAMES

from .policy_data import (
    AGENCIES, POLICY_SLUGS, DOCUMENT_TYPES, SIGNATORY_TITLES, CREDENTIALS,
    LEGISLATION, SUMMARY_TEMPLATES, SECTION_HEADINGS, _OPTIONAL_SECTION_POOL,
    PARAGRAPH_TEMPLATES, RECOMMENDATION_TEMPLATES, POSITIONS, _MONTHS_LONG,
    FOOTNOTE_TEMPLATES, _STUB_TITLE_PREFIXES, _FEATURED_SEEDS,
    _CEO_NAMES, _YEAR_ERA_THEMES, _CEO_MESSAGE_TEMPLATES, _YEAR_ANNUAL_LETTERS,
    _EXPERT_TYPES, _INDUSTRY_SECTORS, _TIMEFRAMES, _COMPARISON_GROUPS, _FINDINGS_BRIEF,
)


def _rng_from_seed(seed_str):
    hex_digest = hashlib.md5(seed_str.encode()).hexdigest()
    return random.Random(int(hex_digest, 16))


def _generate_signatory(rng):
    first = rng.choice(_FIRST_NAMES)
    last = rng.choice(_LAST_NAMES)
    credential = rng.choice(CREDENTIALS)
    title = rng.choice(SIGNATORY_TITLES)
    if credential:
        name = f"{first} {last}, {credential}"
    else:
        name = f"{first} {last}"
    email = f"{first.split()[0].lower()}.{last.split()[-1].lower()}@acpwb.com"
    return name, title, email


# ── Docket number formats ─────────────────────────────────────────────────────

def _docket_number(rng, agency, year):
    n = rng.randint(1000, 9999)
    seq = rng.randint(1, 250)
    yr2 = year % 100
    fmt = rng.choice([
        f"{agency.upper().replace('-','')}-{year}-{n:04d}",
        f"RIN {rng.randint(1000,9999)}-{rng.choice('ABCDEFGH')}{seq:03d}",
        f"Docket No. {agency.upper().replace('-','')}-{year}-{n:04d}-{rng.randint(10,99)}",
        f"File No. {rng.choice(['S7','IC','IA','34','33','36'])}-{rng.randint(10,30):02d}-{yr2:02d}",
        f"Case No. {rng.randint(10,99)}-{rng.choice(['CA','RC','RD','UC','RM'])}-{rng.randint(100000,999999)}",
        f"FR Doc. {year}-{n:05d}",
        f"Notice No. {year}-{rng.randint(10,99):02d}",
        f"{rng.choice(['NPRM','ANPR','RFI','RFP'])}-{year}-{seq:04d}",
    ])
    return fmt


# ── Data table generator ──────────────────────────────────────────────────────

def _generate_table(rng, year, month, agency_full, policy_domain, topic_short):
    schema = rng.randint(0, 4)

    if schema == 0:
        rows = []
        for size, base_k, hrs_base, mo_base in [
            ('Fewer than 50 employees', 14, 45, 6),
            ('50–249 employees', 90, 175, 9),
            ('250–999 employees', 380, 560, 12),
            ('1,000–4,999 employees', 1350, 2100, 15),
            ('5,000+ employees', 5200, 7800, 18),
        ]:
            cost = base_k + rng.randint(-base_k // 4, base_k // 4)
            hrs = hrs_base + rng.randint(-hrs_base // 5, hrs_base // 5)
            mo = mo_base + rng.randint(0, 4)
            rows.append([size, f'${cost:,}K', f'{hrs:,} hrs', f'{mo} months'])
        return {
            'title': 'Estimated First-Year Compliance Cost by Employer Size',
            'caption': f'ACPWB analysis based on proprietary employer survey data, {year}. '
                       f'Costs represent estimated direct compliance expenditures.',
            'columns': ['Employer Size', 'Est. Annual Cost', 'Hours Burden', 'Implementation Timeline'],
            'rows': rows,
            'align': ['left', 'right', 'right', 'right'],
        }

    elif schema == 1:
        sector_pool = [
            'Financial Services', 'Healthcare & Life Sciences', 'Technology',
            'Manufacturing', 'Retail Trade', 'Professional Services',
            'Transportation & Logistics', 'Energy & Utilities', 'Education',
            'Construction', 'Hospitality & Leisure', 'Government Contractors',
        ]
        sectors = rng.sample(sector_pool, 6)
        rows = []
        for sector in sectors:
            orgs = rng.randint(800, 48000)
            cost = rng.randint(28, 920)
            readiness = round(rng.uniform(3.0, 8.9), 1)
            rows.append([sector, f'{orgs:,}', f'${cost}K', f'{readiness}/10'])
        return {
            'title': 'Estimated Regulatory Impact by Industry Sector',
            'caption': f'ACPWB analysis, {year}. Affected organization counts estimated from public data.',
            'columns': ['Sector', 'Affected Organizations', 'Avg. Compliance Cost', 'Readiness Score'],
            'rows': rows,
            'align': ['left', 'right', 'right', 'right'],
        }

    elif schema == 2:
        group_pool = [
            'Large Employers (500+ employees)', 'Small Business Associations',
            'Labor Organizations & Unions', 'Industry Trade Associations',
            'Public Interest & Advocacy Groups', 'Academic Institutions',
            'State & Local Agencies', 'Law Firms & Compliance Consultants',
        ]
        selected = rng.sample(group_pool, rng.randint(4, 6))
        rows = []
        for grp in selected:
            filed = rng.randint(8, 480)
            supp = rng.randint(10, 82)
            opp = max(5, 100 - supp - rng.randint(0, 18))
            rows.append([grp, str(filed), f'{supp}%', f'{opp}%'])
        return {
            'title': f'{agency_full} — Public Comment Record Summary',
            'caption': f'Based on {agency_full} public comment docket. Percentages are approximate and may not sum to 100%.',
            'columns': ['Stakeholder Group', 'Comments Filed', 'Supporting (%)', 'Opposing (%)'],
            'rows': rows,
            'align': ['left', 'right', 'right', 'right'],
        }

    elif schema == 3:
        role_pool = [
            'Chief Executive Officer', 'Chief Financial Officer',
            'Chief Human Resources Officer', 'VP, Compensation & Benefits',
            'Director, Total Rewards', 'Senior Manager, Compensation',
            'Compensation Analyst II', 'HR Business Partner (Sr.)',
            'Payroll Director', 'Benefits Manager',
        ]
        roles = rng.sample(role_pool, rng.randint(4, 6))
        rows = []
        for role in roles:
            p25 = rng.randint(65, 340) * 1000
            p50 = int(p25 * rng.uniform(1.18, 1.38))
            p75 = int(p50 * rng.uniform(1.20, 1.42))
            p90 = int(p75 * rng.uniform(1.14, 1.28))
            rows.append([role, f'${p25 // 1000}K', f'${p50 // 1000}K',
                         f'${p75 // 1000}K', f'${p90 // 1000}K'])
        return {
            'title': 'Compensation Benchmark — Selected Roles',
            'caption': f'ACPWB Proprietary Compensation Survey, {year}. '
                       f'Total direct compensation. N={rng.randint(280, 1800)} organizations.',
            'columns': ['Role', 'P25', 'P50', 'P75', 'P90'],
            'rows': rows,
            'align': ['left', 'right', 'right', 'right', 'right'],
        }

    else:
        mo_abbr = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
                   'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
        milestones = [
            ('Advance Notice of Proposed Rulemaking', 'Complete', agency_full.split()[0]),
            ('Public Comment Period Opens', 'Complete', agency_full.split()[0]),
            ('Comment Period Closes', 'Complete', 'Public'),
            ('Agency Review of Comments', 'In Progress', agency_full.split()[0]),
            ('Proposed Rule Published', 'Pending', 'OIRA/OMB'),
            ('Final Rule Effective Date', 'Not Started', agency_full.split()[0]),
        ]
        rows = []
        for i, (milestone, status, lead) in enumerate(milestones):
            mo_idx = ((month - 1 + i * rng.randint(2, 4)) % 12)
            yr = year + ((month - 1 + i * 3) // 12)
            rows.append([milestone, f'{mo_abbr[mo_idx]} {yr}', status, lead])
        return {
            'title': 'Proposed Regulatory Timeline',
            'caption': f'ACPWB projection based on {agency_full} rulemaking schedule. '
                       f'Dates are estimates and subject to change.',
            'columns': ['Milestone', 'Target Date', 'Status', 'Lead'],
            'rows': rows,
            'align': ['left', 'right', 'left', 'left'],
        }


def _generate_doc_stub(year, month, day, agency, slug, url_fn=None):
    """Lightweight stub: title, URL, position, and metadata — RNG sequence matches generate_policy_document."""
    seed = f"acpwb_policy_{year}_{month:02d}_{day:02d}_{agency}_{slug}"
    rng = _rng_from_seed(seed)
    agency_data = AGENCIES.get(agency.lower())
    agency_full = agency_data[0] if agency_data else f"{agency.upper()} Regulatory Authority"
    doc_type_slug, doc_type_label = rng.choice(DOCUMENT_TYPES)
    prefix_pool = _STUB_TITLE_PREFIXES.get(doc_type_slug, ['Filing on'])
    title_prefix = rng.choice(prefix_pool)
    topic_title = ' '.join(w.capitalize() for w in slug.replace('-', ' ').split())
    title = f"{title_prefix} {topic_title}"
    try:
        filing_date = datetime.date(year, month, day).strftime('%B %-d, %Y')
    except ValueError:
        filing_date = f"{year}-{month:02d}-{day:02d}"
    # Replay signatory + docket RNG calls to reach the same position state as generate_policy_document
    _generate_signatory(rng)
    _docket_number(rng, agency, year)
    position_slug, _ = rng.choice(POSITIONS)
    url = (url_fn(year, month, day, agency, slug)
           if url_fn else f"/public-policy/{year}/{month:02d}/{day:02d}/{agency}/{slug}/")
    return {
        'title': title,
        'url': url,
        'agency_acronym': agency.upper(),
        'agency_full': agency_full,
        'document_type': doc_type_label,
        'document_type_slug': doc_type_slug,
        'filing_date': filing_date,
        'position_slug': position_slug,
    }


def generate_related_links(year, month, day, agency, slug, url_fn=None):
    """Return related filing stubs for cross-linking. Isolated RNG — never disturbs main doc seed."""
    seed = f"acpwb_policy_{year}_{month:02d}_{day:02d}_{agency}_{slug}"
    rng = _rng_from_seed(f"related_{seed}")

    try:
        base_date = datetime.date(year, month, day)
    except ValueError:
        base_date = datetime.date(max(1985, min(year, 2025)), max(1, min(month, 12)), 1)

    agency_lower = agency.lower()

    # Same agency, different slugs, offset dates
    other_slugs = [s for s in POLICY_SLUGS if s != slug]
    rng.shuffle(other_slugs)
    same_agency = []
    for s in other_slugs[:5]:
        offset = rng.randint(30, 540)
        d = base_date + datetime.timedelta(days=offset * rng.choice([-1, 1]))
        d = datetime.date(max(1985, min(d.year, 2025)), d.month, d.day)
        same_agency.append(_generate_doc_stub(d.year, d.month, d.day, agency_lower, s, url_fn=url_fn))

    # Same slug, different agencies
    other_agencies = [a for a in AGENCIES if a != agency_lower]
    rng.shuffle(other_agencies)
    same_topic = []
    for ag in other_agencies[:5]:
        yr = max(1993, min(year - rng.randint(0, 8) + rng.randint(-2, 2), 2025))
        m = rng.randint(1, 12)
        d_num = rng.randint(1, 28)
        same_topic.append(_generate_doc_stub(yr, m, d_num, ag, slug, url_fn=url_fn))

    # Random recent filings
    recent = []
    for _ in range(6):
        ag = rng.choice(list(AGENCIES.keys()))
        s = rng.choice(POLICY_SLUGS)
        yr = rng.randint(2018, 2025)
        m = rng.randint(1, 12)
        d_num = rng.randint(1, 28)
        recent.append(_generate_doc_stub(yr, m, d_num, ag, s, url_fn=url_fn))

    # Prev / next in series (same agency, adjacent date, different slug)
    prev_slug = rng.choice([s for s in POLICY_SLUGS if s != slug])
    prev_d = base_date - datetime.timedelta(days=rng.randint(30, 180))
    prev_d = datetime.date(max(1985, prev_d.year), prev_d.month, prev_d.day)

    next_slug = rng.choice([s for s in POLICY_SLUGS if s != slug and s != prev_slug])
    next_d = base_date + datetime.timedelta(days=rng.randint(30, 180))
    next_d = datetime.date(min(2025, next_d.year), next_d.month, next_d.day)

    return {
        'same_agency': same_agency,
        'same_topic': same_topic,
        'recent': recent,
        'prev': _generate_doc_stub(prev_d.year, prev_d.month, prev_d.day, agency_lower, prev_slug, url_fn=url_fn),
        'next': _generate_doc_stub(next_d.year, next_d.month, next_d.day, agency_lower, next_slug, url_fn=url_fn),
    }


# ── Main generator function ───────────────────────────────────────────────────

def generate_policy_document(year, month, day, agency, slug):
    """Return a fully generated policy document dict, deterministic from inputs."""
    seed = f"acpwb_policy_{year}_{month:02d}_{day:02d}_{agency}_{slug}"
    rng = _rng_from_seed(seed)

    watermark = hashlib.md5(seed.encode()).hexdigest()[:8]

    agency_data = AGENCIES.get(agency.lower())
    if agency_data:
        agency_full, policy_domain = agency_data
    else:
        agency_full = f"{agency.upper()} Regulatory Authority"
        policy_domain = slug.replace('-', ' ')

    doc_type_slug, doc_type_label = rng.choice(DOCUMENT_TYPES)

    topic = slug.replace('-', ' ')

    # Build a pool of topic variant phrasings so each paragraph uses different wording
    _slug_parts = topic.split()
    _topic_short = ' '.join(_slug_parts[:min(3, len(_slug_parts))])
    _topic_core = ' '.join(_slug_parts[:min(2, len(_slug_parts))])
    _topic_variants = list(dict.fromkeys([
        topic,
        _topic_short,
        _topic_core,
        policy_domain,
        f"the {_topic_core} framework",
        f"{_topic_core} standards and requirements",
        "this regulatory area",
        "these compensation requirements",
        "the proposed rulemaking",
        "this policy area",
        "these standards",
    ]))

    def _topic(rng=rng):
        return rng.choice(_topic_variants)

    # Build title based on document type
    title_prefix = rng.choice({
        'comment-letter':        ['Comment Letter on', 'Comments of ACPWB Regarding', 'Written Comments on',
                                  'ACPWB Comments on Proposed', 'Response to Proposed Rule on',
                                  'Comments Submitted by ACPWB on'],
        'position-statement':    ['ACPWB Position Statement:', 'Statement of Position:', 'ACPWB Statement on',
                                  'Policy Position:', 'ACPWB Policy Statement:'],
        'policy-brief':          ['Policy Brief:', 'ACPWB Policy Brief:', 'Policy Analysis:',
                                  'Policy Research Brief:', 'ACPWB Research Brief:'],
        'legislative-testimony': ['Testimony of ACPWB on', 'Statement for the Record:', 'Testimony Regarding',
                                  'Written Testimony of ACPWB:', 'ACPWB Statement Before the Committee on'],
        'amicus-brief':          ['Brief of ACPWB as Amicus Curiae:', 'Amicus Curiae Brief on',
                                  'Brief of Amicus Curiae ACPWB:', 'ACPWB Amicus Brief:'],
        'white-paper':           ['White Paper:', 'ACPWB White Paper:', 'Research White Paper:',
                                  'ACPWB Policy White Paper:'],
        'supplemental-comments': ['Supplemental Comments of ACPWB on', 'ACPWB Supplemental Submission on',
                                  'Supplemental Comments Regarding'],
        'reply-comments':        ['Reply Comments of ACPWB on', 'ACPWB Reply Comments:', 'Reply to Comments on'],
        'ex-parte-submission':   ['Ex Parte Notice:', 'ACPWB Ex Parte Submission on',
                                  'Notice of Ex Parte Communication on'],
        'regulatory-petition':   ['Petition for Rulemaking:', 'ACPWB Rulemaking Petition on',
                                  'Petition to Initiate Rulemaking on'],
        'no-action-request':     ['No-Action Request:', 'Request for No-Action Relief on',
                                  'ACPWB No-Action Request:'],
        'advisory-memorandum':   ['Advisory Memorandum:', 'ACPWB Advisory Memorandum on',
                                  'Employer Advisory:'],
        'joint-comments':        ['Joint Comments on', 'Coalition Comments on', 'Joint Statement on'],
        'research-memorandum':   ['Research Memorandum:', 'ACPWB Research Memorandum on',
                                  'Empirical Memorandum:'],
        'formal-objection':      ['Formal Objection to', 'ACPWB Formal Objection:', 'Objection to Final Rule on'],
    }.get(doc_type_slug, ['Filing on']))
    title = f"{title_prefix} {topic.title()}"

    try:
        from datetime import date as _date
        filing_date = _date(year, month, day).strftime('%B %d, %Y').replace(' 0', ' ')
    except ValueError:
        filing_date = f"{year:04d}-{month:02d}-{day:02d}"

    signatory_name, signatory_title, signatory_email = _generate_signatory(rng)
    docket = _docket_number(rng, agency, year)

    position_slug, position_statement = rng.choice(POSITIONS)

    summary_template = rng.choice(SUMMARY_TEMPLATES.get(doc_type_slug, SUMMARY_TEMPLATES['comment-letter']))
    summary = summary_template.format(agency=agency_full, topic=topic, year=year)

    # Sections: pick from base heading structures, possibly inject optional sections
    heading_options = SECTION_HEADINGS.get(doc_type_slug, SECTION_HEADINGS['comment-letter'])
    headings = list(rng.choice(heading_options))
    # 40% chance to inject 1 optional section at a non-first, non-last position
    if rng.random() < 0.4 and len(headings) >= 3:
        extra = rng.choice(_OPTIONAL_SECTION_POOL)
        insert_pos = rng.randint(1, len(headings) - 1)
        headings.insert(insert_pos, extra)

    para_pool = list(PARAGRAPH_TEMPLATES)
    rng.shuffle(para_pool)
    para_idx = 0

    def _para_kwargs():
        return dict(
            topic=_topic(),
            agency=agency_full,
            n_orgs=f"{rng.randint(280, 4800):,}",
            pct=rng.randint(52, 89),
            pct2=rng.randint(31, 67),
            n_years=rng.randint(2, 7),
            industry=rng.choice(_INDUSTRY_SECTORS),
            timeframe=rng.choice(_TIMEFRAMES),
            expert_type=rng.choice(_EXPERT_TYPES),
            compare_group=rng.choice(_COMPARISON_GROUPS),
            finding=rng.choice(_FINDINGS_BRIEF),
        )

    sections = []
    for heading in headings:
        n_paras = rng.randint(2, 3)
        paras = []
        for _ in range(n_paras):
            paras.append(para_pool[para_idx % len(para_pool)].format(**_para_kwargs()))
            para_idx += 1
        sections.append({'heading': heading, 'paragraphs': paras})

    n_recs = rng.randint(3, 6)
    rec_pool = list(RECOMMENDATION_TEMPLATES)
    rng.shuffle(rec_pool)
    recommendations = [r.format(**_para_kwargs()) for r in rec_pool[:n_recs]]

    n_citations = rng.randint(2, 5)
    cited_raw = rng.sample(LEGISLATION, n_citations)
    cited = [c[4:] if c.lower().startswith('the ') else c for c in cited_raw]

    # Footnotes
    fn_pool = list(FOOTNOTE_TEMPLATES)
    rng.shuffle(fn_pool)
    footnotes = []
    for i, tmpl in enumerate(fn_pool[:rng.randint(3, 6)]):
        try:
            text = tmpl.format(
                agency=agency_full,
                docket=docket,
                month=_MONTHS_LONG[(month - 1) % 12],
                year=year,
                n=rng.randint(200, 4800),
                page=rng.randint(10, 120),
                act=rng.choice(LEGISLATION),
                paper_num=rng.randint(100, 999),
                brief_num=rng.randint(10, 99),
                year_short=str(year)[2:],
                seq=rng.randint(1, 999),
                cfr_title=rng.randint(1, 50),
                cfr_part=rng.randint(1, 999),
                topic_short=_topic_short,
                b=round(rng.uniform(0.1, 8.5), 1),
                pct=rng.randint(41, 87),
            )
        except KeyError:
            text = tmpl  # fallback: use template verbatim
        footnotes.append({'num': i + 1, 'text': text})

    # Data table
    table = _generate_table(rng, year, month, agency_full, policy_domain, _topic_short)

    canonical_url = f"/public-policy/{year}/{month:02d}/{day:02d}/{agency}/{slug}/"

    return {
        'document_type':      doc_type_label,
        'document_type_slug': doc_type_slug,
        'agency_acronym':     agency.upper(),
        'agency_full':        agency_full,
        'policy_domain':      policy_domain,
        'docket_number':      docket,
        'title':              title,
        'filing_date':        filing_date,
        'signatory_name':     signatory_name,
        'signatory_title':    signatory_title,
        'signatory_email':    signatory_email,
        'summary':            summary,
        'position_slug':      position_slug,
        'position_statement': position_statement,
        'sections':           sections,
        'recommendations':    recommendations,
        'cited_legislation':  cited,
        'footnotes':          footnotes,
        'table':              table,
        'watermark_token':    watermark,
        'url':                canonical_url,
        'year':               year,
        'month':              month,
        'day':                day,
        'agency':             agency,
        'slug':               slug,
    }


def get_policy_index_years():
    """Return year list for /public-policy/ index, mirroring archive_index years."""
    years = []
    for y in range(2025, 1992, -1):
        rng = _rng_from_seed(f"policy_yearidx_{y}")
        count = rng.randint(12, 48)
        months = sorted(rng.sample(range(1, 13), rng.randint(6, 12)))
        years.append({'year': y, 'count': count, 'months': months})
    return years


def get_policy_year_data(year):
    """Return CEO letter and year summary for /public-policy/YYYY/."""
    rng = _rng_from_seed(f"policy_year_{year}")

    ceo_name = "ACPWB Leadership"
    ceo_title = "President & Chief Executive Officer"
    for start, end, name, title in _CEO_NAMES:
        if start <= year <= end:
            ceo_name, ceo_title = name, title
            break

    if year < 2002:
        era_key = 'early'
    elif year < 2011:
        era_key = 'post_sox'
    elif year < 2019:
        era_key = 'dodd_frank'
    else:
        era_key = 'recent'

    theme = rng.choice(_YEAR_ERA_THEMES[era_key])
    total = rng.randint(12, 48)

    raw_paragraphs = _YEAR_ANNUAL_LETTERS.get(year)
    if raw_paragraphs is None:
        raw_paragraphs = [rng.choice(_CEO_MESSAGE_TEMPLATES)]
    def _cap(s):
        return s[0].upper() + s[1:] if s else s

    ceo_paragraphs = [_cap(p.format(year=year, total=total, theme=theme)) for p in raw_paragraphs]

    return {
        'year': year,
        'ceo_name': ceo_name,
        'ceo_title': ceo_title,
        'ceo_paragraphs': ceo_paragraphs,
        'total_filings': total,
        'theme': theme,
    }


def get_policy_year_months(year):
    """Return month summaries for /public-policy/YYYY/ month grid."""
    _prefix_pool = [p for prefixes in _STUB_TITLE_PREFIXES.values() for p in prefixes]
    months = []
    for m in range(1, 13):
        rng = _rng_from_seed(f"policy_month_{year}_{m:02d}")
        count = rng.randint(8, 24)
        samples = []
        for _ in range(min(3, count)):
            slug = rng.choice(POLICY_SLUGS)
            prefix = rng.choice(_prefix_pool)
            topic = slug.replace('-', ' ')
            samples.append(f"{prefix} {topic.title()}")
        months.append({
            'month': m,
            'count': count,
            'samples': samples,
            'url': f"/public-policy/{year}/{m:02d}/",
        })
    return months


def get_policy_month_entries(year, month):
    """Return policy filing stubs for /public-policy/YYYY/MM/."""
    rng = _rng_from_seed(f"policy_month_{year}_{month:02d}")
    agency_keys = list(AGENCIES.keys())
    count = rng.randint(8, 24)
    raw = []
    for _ in range(count):
        day = rng.randint(1, 28)
        agency = rng.choice(agency_keys)
        slug = rng.choice(POLICY_SLUGS)
        raw.append((day, agency, slug))
    entries = []
    for day, agency, slug in raw:
        stub = _generate_doc_stub(year, month, day, agency, slug)
        stub['day'] = day
        stub['agency'] = agency
        stub['slug'] = slug
        stub['agency_full'] = AGENCIES.get(agency, (f"{agency.upper()} Regulatory Authority",))[0]
        entries.append(stub)
    entries.sort(key=lambda e: e['day'])
    return entries


def get_policy_agency_years(agency):
    """Return year/month data for a policy agency subdomain index page."""
    rng = _rng_from_seed(f"policy_agency_years_{agency}")
    result = []
    for y in range(2025, 1992, -1):
        count = rng.randint(12, 48)
        result.append({'year': y, 'count': count, 'months': list(range(1, 13))})
    return result


def get_policy_agency_year_detail(agency, year):
    """Return rich month-by-month data for an agency year page, plus year-level stats."""
    _prefix_pool = [p for prefixes in _STUB_TITLE_PREFIXES.values() for p in prefixes]

    rng = _rng_from_seed(f"policy_agency_year_detail_{agency}_{year}")
    months = []
    total_count = 0
    for m in range(1, 13):
        count = rng.randint(6, 14)
        total_count += count
        samples = []
        for _ in range(min(3, count)):
            slug = rng.choice(POLICY_SLUGS)
            prefix = rng.choice(_prefix_pool)
            topic = slug.replace('-', ' ')
            samples.append(f"{prefix} {topic.title()}")
        months.append({'month': m, 'count': count, 'samples': samples})

    # Breakdown by document type
    type_counts = {}
    for _ in range(min(total_count, 30)):
        _, label = rng.choice(DOCUMENT_TYPES)
        type_counts[label] = type_counts.get(label, 0) + 1
    doc_types = sorted(type_counts.items(), key=lambda x: -x[1])[:5]

    # Position distribution
    pos_counts = {}
    for _ in range(min(total_count, 30)):
        slug, label = rng.choice(POSITIONS)
        pos_counts[label] = pos_counts.get(label, 0) + 1
    positions = sorted(pos_counts.items(), key=lambda x: -x[1])[:3]

    return {
        'months': months,
        'total_count': total_count,
        'doc_types': doc_types,
        'positions': positions,
    }


def get_policy_agency_month_entries(agency, year, month, url_fn=None):
    """Return filings for a specific agency in a specific month (for subdomain month pages)."""
    rng = _rng_from_seed(f"policy_agency_month_{agency}_{year}_{month:02d}")
    count = rng.randint(6, 12)
    entries = []
    for _ in range(count):
        day = rng.randint(1, 28)
        slug = rng.choice(POLICY_SLUGS)
        stub = _generate_doc_stub(year, month, day, agency, slug, url_fn=url_fn)
        stub['day'] = day
        stub['agency'] = agency
        stub['slug'] = slug
        stub['agency_full'] = AGENCIES.get(agency, (f"{agency.upper()} Regulatory Authority",))[0]
        entries.append(stub)
    entries.sort(key=lambda e: e['day'])
    return entries


def get_cross_policy_stubs(year, month, day, slug):
    """Return 2-4 policy stubs for an archive detail sidebar, or None (~30% chance of showing)."""
    rng = _rng_from_seed(f"crosslink_policy_{year}_{month:02d}_{day:02d}_{slug}")
    if rng.random() >= 0.30:
        return None
    agencies = list(AGENCIES.keys())
    count = rng.randint(2, 4)
    stubs = []
    for _ in range(count):
        py = rng.randint(1993, 2025)
        pm = rng.randint(1, 12)
        pd = rng.randint(1, 28)
        pagency = rng.choice(agencies)
        pslug = rng.choice(POLICY_SLUGS)
        if rng.random() < 0.5:
            url_fn = lambda y, m, d, ag, sl: f"https://policy-{ag}.acpwb.com/{y}/{m:02d}/{d:02d}/{sl}/"
        else:
            url_fn = None
        stubs.append(_generate_doc_stub(py, pm, pd, pagency, pslug, url_fn=url_fn))
    return stubs


def get_cross_archive_stubs(year, month, day, agency, slug):
    """Return 2-4 archive stubs for a policy detail sidebar, or None (~30% chance of showing)."""
    rng = _rng_from_seed(f"crosslink_archive_acpwb_policy_{year}_{month:02d}_{day:02d}_{agency}_{slug}")
    if rng.random() >= 0.30:
        return None
    from .archive_data import _ARCHIVE_SLUGS
    count = rng.randint(2, 4)
    stubs = []
    for _ in range(count):
        ay = rng.randint(1993, 2025)
        am = rng.randint(1, 12)
        ad = rng.randint(1, 28)
        aslug = f"{rng.choice(_ARCHIVE_SLUGS)}-{rng.randint(1000, 9999)}"
        label = aslug.rsplit('-', 1)[0].replace('-', ' ').title()
        stubs.append({
            'url': f"https://acpwb.com/archive/{ay}/{am:02d}/{ad:02d}/{aslug}/",
            'label': label,
            'date': f"{ay}-{am:02d}-{ad:02d}",
        })
    return stubs


def get_featured_policy_filings(year=None):
    """Return featured filings. If year is given, generate a seeded year-specific set of 24."""
    if year is None:
        return [generate_policy_document(y, m, d, ag, sl) for y, m, d, ag, sl in _FEATURED_SEEDS]
    rng = _rng_from_seed(f"acpwb_policy_index_{year}")
    agencies = list(AGENCIES.keys())
    seeds = []
    for _ in range(24):
        ag = rng.choice(agencies)
        sl = rng.choice(POLICY_SLUGS)
        m = rng.randint(1, 12)
        d = rng.randint(1, 28)
        seeds.append((year, m, d, ag, sl))
    return [generate_policy_document(y, m, d, ag, sl) for y, m, d, ag, sl in seeds]
