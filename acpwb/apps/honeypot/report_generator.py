"""
Deterministic report generator for ACPWB Reports & Publications.
All output is seeded from the slug — same slug always returns same content.
"""
import csv
import hashlib
import io
import random
import re
from datetime import date

from django.utils.text import slugify


def _rng_from_seed(seed_str):
    """Deterministic random.Random from a string seed (same as projects/generators.py)."""
    hex_digest = hashlib.md5(seed_str.encode()).hexdigest()
    return random.Random(int(hex_digest, 16))


# ── Year distribution (weighted toward recent, goes back to 1993) ─────────────

YEAR_POOL = (
    list(range(1993, 2000)) * 1 +   # rare early archive
    list(range(2000, 2008)) * 2 +   # growing cadence
    list(range(2008, 2016)) * 3 +   # established org
    list(range(2016, 2023)) * 5 +   # prolific period
    list(range(2023, 2026)) * 8     # current era
)


# ── Canonical catalog ─────────────────────────────────────────────────────────

REPORT_CATALOG = [
    {'slug': 'salary-compensation-benchmarking-survey-2024',  'title': 'Salary & Compensation Benchmarking Survey 2024',  'category': 'Compensation', 'file_type': 'csv'},
    {'slug': 'annual-workforce-analytics-report-2024',        'title': 'Annual Workforce Analytics Report 2024',          'category': 'Workforce',    'file_type': 'pdf'},
    {'slug': 'executive-compensation-study-2024',             'title': 'Executive Compensation Study 2024',               'category': 'Compensation', 'file_type': 'csv'},
    {'slug': 'corporate-governance-best-practices-white-paper','title': 'Corporate Governance Best Practices White Paper', 'category': 'Governance',   'file_type': 'pdf'},
    {'slug': 'esg-disclosure-framework-2024',                 'title': 'ESG Disclosure Framework 2024',                   'category': 'ESG',          'file_type': 'pdf'},
    {'slug': 'q3-economic-indicators-2024',                   'title': 'Q3 Economic Indicators Report 2024',              'category': 'Economic',     'file_type': 'csv'},
    {'slug': 'benefits-cost-analysis-2024',                   'title': 'Benefits Cost Analysis 2024',                     'category': 'Benefits',     'file_type': 'csv'},
    {'slug': 'industry-wage-gap-analysis-2024',               'title': 'Industry Wage Gap Analysis 2024',                 'category': 'Compensation', 'file_type': 'pdf'},
    {'slug': 'fortune-500-ceo-pay-ratio-analysis',            'title': 'Fortune 500 CEO Pay Ratio Analysis',              'category': 'Compensation', 'file_type': 'csv'},
    {'slug': 'state-of-workplace-report-2024',                'title': 'State of the Workplace Report 2024',              'category': 'Workforce',    'file_type': 'pdf'},
    {'slug': 'regulatory-compliance-costs-2024',              'title': 'Regulatory Compliance Costs Report 2024',         'category': 'Governance',   'file_type': 'pdf'},
    {'slug': 'employee-satisfaction-survey-results-2024',     'title': 'Employee Satisfaction Survey Results 2024',       'category': 'Workforce',    'file_type': 'csv'},
    {'slug': 'regional-salary-differentials-2024',            'title': 'Regional Salary Differentials Study 2024',        'category': 'Compensation', 'file_type': 'csv'},
    {'slug': 'board-composition-study-2024',                  'title': 'Board Composition Study 2024',                   'category': 'Governance',   'file_type': 'csv'},
    {'slug': 'retirement-readiness-index-2024',               'title': 'Retirement Readiness Index 2024',                 'category': 'Benefits',     'file_type': 'csv'},
    {'slug': 'public-sector-compensation-guide',              'title': 'Public Sector Compensation Guide',                'category': 'Compensation', 'file_type': 'pdf'},
    {'slug': 'healthcare-benefits-landscape-2024',            'title': 'Healthcare Benefits Landscape 2024',              'category': 'Benefits',     'file_type': 'pdf'},
    {'slug': 'remote-work-impact-study-2024',                 'title': 'Remote Work Impact Study 2024',                   'category': 'Workforce',    'file_type': 'csv'},
    # Historical anchors
    {'slug': 'national-salary-survey-2001',                   'title': 'National Salary Survey 2001',                     'category': 'Compensation', 'file_type': 'pdf'},
    {'slug': 'corporate-pay-equity-study-1998',               'title': 'Corporate Pay Equity Study 1998',                 'category': 'Compensation', 'file_type': 'pdf'},
    {'slug': 'workforce-demographics-report-2006',            'title': 'Workforce Demographics Report 2006',              'category': 'Workforce',    'file_type': 'pdf'},
    {'slug': 'executive-benefits-survey-2003',                'title': 'Executive Benefits Survey 2003',                  'category': 'Benefits',     'file_type': 'csv'},
    {'slug': 'governance-disclosure-standards-2009',          'title': 'Governance Disclosure Standards 2009',            'category': 'Governance',   'file_type': 'pdf'},
    {'slug': 'compensation-benchmarking-methodology-1995',    'title': 'Compensation Benchmarking Methodology 1995',      'category': 'Compensation', 'file_type': 'pdf'},
    {'slug': 'annual-hr-metrics-report-2011',                 'title': 'Annual HR Metrics Report 2011',                   'category': 'Workforce',    'file_type': 'csv'},
    {'slug': 'job-grade-architecture-white-paper-2007',       'title': 'Job Grade Architecture White Paper 2007',         'category': 'Compensation', 'file_type': 'pdf'},
]

REPORT_CATEGORIES = [
    'Compensation', 'Workforce', 'Governance', 'ESG', 'Benefits',
    'Economic', 'Compliance', 'Diversity & Inclusion', 'Technology',
    'Leadership', 'Talent Acquisition', 'Retention',
]

REPORT_ADJECTIVES = [
    'Annual', 'Quarterly', 'Biennial', 'Longitudinal', 'Comprehensive',
    'National', 'Regional', 'Cross-Industry', 'Mid-Year', 'Benchmark',
    'Baseline', 'Trend', 'Comparative', 'Strategic', 'Preliminary',
    'Final', 'Supplemental', 'Revised', 'Expanded', 'Integrated',
]

REPORT_SUBJECTS = [
    'Compensation Benchmarking', 'Pay Equity', 'Workforce Planning',
    'Talent Retention', 'Benefits Utilization', 'Salary Transparency',
    'Board Effectiveness', 'Executive Pay', 'Compliance Risk',
    'Diversity Metrics', 'Turnover Analysis', 'Skills Gap Assessment',
    'Remote Work Productivity', 'Healthcare Cost Trends',
    'Retirement Plan Participation', 'Incentive Compensation',
    'Total Rewards Strategy', 'CEO Pay Ratio', 'Gender Pay Gap',
    'Geographic Pay Differentials', 'Wage Band Calibration',
    'Job Architecture Review', 'Market Pricing Methodology',
    'Employee Engagement Index', 'Hiring Cost Analysis',
    'Pay Transparency Compliance', 'Workforce Demographics',
]

REPORT_SUFFIXES = [
    'Report', 'Study', 'Survey', 'Analysis', 'Index', 'Guide',
    'Assessment', 'Framework', 'White Paper', 'Briefing',
]

SUMMARY_TEMPLATES = [
    "Drawing on proprietary survey data collected from {n} organizations across {states} states, this {adj} {subject} {suffix} documents {finding}. The analysis incorporates {years} years of longitudinal data and has been peer-reviewed by ACPWB's Research Advisory Panel.",
    "This {adj} {subject} {suffix} presents findings from ACPWB's national survey of {n} compensation professionals. Key findings include {finding}, with significant variation observed across industry sectors and geographic regions.",
    "Prepared by the ACPWB Research Division in collaboration with {n} contributing organizations, this {suffix} examines {finding} and provides actionable benchmarks for HR and finance professionals. Detailed methodology is available in Appendix B.",
    "An examination of {subject} trends across {n} participating organizations, this {suffix} draws on data collected between Q{q1} {year_from} and Q{q2} {year_to}. Findings are presented in aggregate to preserve respondent confidentiality.",
    "ACPWB's {adj} {suffix} synthesizes responses from {n} HR and finance leaders on the state of {subject} in American enterprise. The data reveals {finding}, underscoring the need for standardized benchmarking practices.",
    "Released by the ACPWB Research Division, this {adj} {suffix} on {subject} draws from a nationally representative sample of {n} employers. The report provides granular breakdowns by industry, geography, and company size.",
]

FINDING_PHRASES = [
    "median base salary increases of {p}% across all job families",
    "a {p}% year-over-year rise in total compensation costs",
    "widening pay disparity between executive and non-executive roles",
    "accelerating adoption of performance-based incentive structures",
    "significant geographic variation in compensation for comparable roles",
    "declining employee satisfaction with current benefits packages",
    "increased regulatory scrutiny of pay practices across sectors",
    "a {p}% correlation between salary transparency and employee retention",
    "median total rewards gaps of {p}% between demographic cohorts",
    "growing adoption of pay band disclosure among Fortune 500 firms",
    "a {p}-point decline in employee confidence in compensation fairness",
]

DOCUMENT_PARAGRAPH_TEMPLATES = [
    "Based on {n} survey responses collected during {quarter} {year}, ACPWB's analysis reveals a median of ${val:,} — representing a {dir} of {p}% compared to the prior-year period. These findings are consistent with broader macroeconomic trends documented by the Bureau of Labor Statistics and peer research organizations.",
    "The {industry} sector demonstrated notable outcomes in the measured period, with the {percentile}th percentile of respondents reporting total compensation of ${val:,}. Geographic dispersion remains pronounced, with respondents in high-cost-of-living markets earning {p}% more than peers in comparable roles in lower-cost regions.",
    "Regression analysis controlling for years of experience, company size, and geographic market indicates that role-level classification accounts for approximately {p}% of variance in total compensation. This finding is consistent across all industry segments examined in this study.",
    "ACPWB's proprietary index for this metric — calculated using a weighting methodology refined over {years} years of continuous research — stands at {idx:.1f} as of {quarter} {year}, representing a movement of {pts} basis points from the prior measurement period.",
    "Respondents citing {factor} as a primary driver of compensation decisions increased by {p}% year-over-year, reflecting a broader shift in how organizations approach total rewards strategy. ACPWB anticipates this trend will continue through the next measurement cycle.",
    "This finding aligns with reporting from the WorldatWork Total Rewards Survey and the Mercer Compensation Planning Survey, both of which identified {factor} as a key determinant of compensation outcomes in the {industry} sector for this performance year.",
    "Organizations in the top quartile of {factor} implementation reported {p}% lower voluntary turnover than their peers, and ${val:,} lower average cost-per-hire. These findings suggest a meaningful return on investment for employers who prioritize systematic compensation strategy.",
]

SECTION_HEADINGS = [
    'Methodology & Data Sources',
    'Industry Benchmarks',
    'Regional Analysis',
    'Year-Over-Year Trends',
    'Demographic Breakdown',
    'Peer Comparisons',
    'Compliance & Regulatory Context',
    'Outlook & Recommendations',
    'Total Rewards Landscape',
    'Pay Equity Findings',
    'Executive Compensation Highlights',
    'Benefits Utilization Patterns',
]

INDUSTRIES_LONG = [
    'Healthcare', 'Financial Services', 'Technology', 'Manufacturing',
    'Retail', 'Energy', 'Real Estate', 'Transportation', 'Education',
    'Government', 'Nonprofit', 'Pharmaceuticals', 'Defense',
    'Consumer Goods', 'Telecommunications', 'Media', 'Professional Services',
]

FACTORS = [
    'market-rate benchmarking',
    'internal pay equity adjustments',
    'skills-based pay structures',
    'geographic differential policies',
    'retention-driven merit increases',
    'pay transparency regulations',
    'executive compensation alignment',
    'total rewards optimization',
]

APPENDIX_ORGS = [
    'Meridian Group Holdings', 'Pinnacle Capital Advisors', 'Apex Workforce Solutions',
    'Summit Benefits Consulting', 'Vanguard HR Partners', 'Horizon Compensation Group',
    'Latitude Talent Analytics', 'Nexus Total Rewards', 'Catalyst People Solutions',
    'Benchmark Compensation Advisors', 'Keystone HR Consulting', 'Zenith Workforce Institute',
    'Atlas People Partners', 'Cornerstone Compensation Group', 'Forefront HR Advisory',
    'Paradigm Pay Consulting', 'Blueprint Talent Solutions', 'Vertex Workforce Analytics',
    'Apex Salary Intelligence', 'Summit Pay Research Group', 'Allied Workforce Partners',
    'National HR Benchmarking Consortium', 'Premier Compensation Research LLC',
]


# ── Watermark & dates ─────────────────────────────────────────────────────────

def _watermark_for(slug):
    return hashlib.md5(f"acpwb_report_{slug}".encode()).hexdigest()[:8]


def _pub_date_for(slug):
    rng = _rng_from_seed(f"pubdate_{slug}")
    m = re.search(r'(\d{4})', slug)
    if m:
        year = int(m.group(1))
        if 1990 <= year <= 2025:
            month = rng.randint(1, 12)
            if any(w in slug for w in ('annual', 'year', 'survey')):
                month = rng.choice([9, 10, 11, 12])
            return date(year, month, rng.randint(1, 28))
    year = rng.choice(YEAR_POOL)
    return date(year, rng.randint(1, 12), rng.randint(1, 28))


# ── Core enrichment ───────────────────────────────────────────────────────────

def _summary_for(slug, rng):
    template = rng.choice(SUMMARY_TEMPLATES)
    finding = rng.choice(FINDING_PHRASES).format(p=rng.randint(3, 22))
    return template.format(
        n=rng.randint(120, 2800),
        states=rng.randint(22, 50),
        adj=rng.choice(REPORT_ADJECTIVES).lower(),
        subject=rng.choice(REPORT_SUBJECTS).lower(),
        suffix=rng.choice(REPORT_SUFFIXES).lower(),
        finding=finding,
        years=rng.randint(3, 18),
        year=rng.randint(2010, 2024),
        q1=rng.randint(1, 2),
        q2=rng.randint(3, 4),
        year_from=rng.randint(2015, 2021),
        year_to=rng.randint(2022, 2024),
        p=rng.randint(3, 22),
    )


def _fake_size_for(file_type, rng):
    if file_type == 'csv':
        return f"{rng.randint(80, 820)} KB"
    return f"{rng.randint(1, 9)}.{rng.randint(1, 9)} MB"


def _enrich_report(entry):
    slug = entry['slug']
    rng = _rng_from_seed(f"enrich_{slug}")
    pub_date = _pub_date_for(slug)
    file_type = entry['file_type']
    return {
        'slug': slug,
        'title': entry['title'],
        'category': entry['category'],
        'file_type': file_type,
        'pub_date': pub_date.isoformat(),
        'pub_date_display': pub_date.strftime('%B %Y'),
        'pub_year': pub_date.year,
        'summary': _summary_for(slug, rng),
        'fake_size': _fake_size_for(file_type, rng),
        'watermark_token': _watermark_for(slug),
        'is_legacy': pub_date.year < 2005,
        'detail_url': f'/reports/{slug}/',
        'download_url': f'/reports/{slug}/download.csv' if file_type == 'csv' else f'/reports/{slug}/download.pdf',
        'row_count': rng.randint(300, 800) if file_type == 'csv' else None,
    }


# ── Synthetic generation (page 3+) ────────────────────────────────────────────

def _generate_synthetic(rng, seed):
    adj = rng.choice(REPORT_ADJECTIVES)
    subject = rng.choice(REPORT_SUBJECTS)
    suffix = rng.choice(REPORT_SUFFIXES)
    file_type = rng.choice(['csv', 'csv', 'pdf'])
    category = rng.choice(REPORT_CATEGORIES)
    title = f"{adj} {subject} {suffix}"
    year = rng.choice(YEAR_POOL)
    if year >= 2005:
        title += f" {year}"
    slug = slugify(title)[:96]
    # append seed fragment to guarantee uniqueness from catalog
    slug = f"{slug}-{hashlib.md5(seed.encode()).hexdigest()[:4]}"
    return {'slug': slug, 'title': title, 'category': category, 'file_type': file_type}


# ── Page generation ───────────────────────────────────────────────────────────

def generate_reports_for_page(page, count=12):
    catalog_start = (page - 1) * count
    catalog_end = catalog_start + count
    results = []

    for entry in REPORT_CATALOG[catalog_start:catalog_end]:
        results.append(_enrich_report(entry))

    needed = count - len(results)
    for i in range(needed):
        seed = f"synthetic_page{page}_item{i}"
        rng = _rng_from_seed(seed)
        entry = _generate_synthetic(rng, seed)
        results.append(_enrich_report(entry))

    results.sort(key=lambda r: r['pub_date'], reverse=True)
    return results


# ── Single-report lookup ───────────────────────────────────────────────────────

def get_or_generate_report_meta(slug):
    """Return enriched report dict for any slug — never 404."""
    from .models import PublicReport
    existing = PublicReport.objects.filter(slug=slug).first()
    if existing:
        pub_date = existing.pub_date
        file_type = existing.file_type
        return {
            'slug': existing.slug,
            'title': existing.title,
            'category': existing.category,
            'file_type': file_type,
            'pub_date': pub_date.isoformat(),
            'pub_date_display': pub_date.strftime('%B %Y'),
            'pub_year': pub_date.year,
            'summary': existing.summary,
            'fake_size': existing.summary[:4],  # unused on detail page
            'watermark_token': existing.watermark_token,
            'is_legacy': pub_date.year < 2005,
            'detail_url': f'/reports/{slug}/',
            'download_url': f'/reports/{slug}/download.csv' if file_type == 'csv' else f'/reports/{slug}/download.pdf',
        }

    for entry in REPORT_CATALOG:
        if entry['slug'] == slug:
            return _enrich_report(entry)

    rng = _rng_from_seed(f"arbitrary_{slug}")
    file_type = rng.choice(['csv', 'pdf'])
    category = rng.choice(REPORT_CATEGORIES)
    title = slug.replace('-', ' ').title()
    return _enrich_report({'slug': slug, 'title': title, 'category': category, 'file_type': file_type})


# ── CSV generation ────────────────────────────────────────────────────────────

_COMP_TITLES = [
    'Analyst', 'Senior Analyst', 'Manager', 'Senior Manager', 'Director',
    'Senior Director', 'VP', 'SVP', 'EVP', 'Principal',
    'Associate', 'Specialist', 'Consultant', 'Lead', 'Partner',
]
_DEPARTMENTS = [
    'Finance', 'Human Resources', 'Operations', 'Strategy', 'Technology',
    'Marketing', 'Legal', 'Compliance', 'Sales', 'Product',
    'Engineering', 'Research', 'Supply Chain', 'Risk Management',
]
_STATES = [
    'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA','HI','ID','IL','IN','IA',
    'KS','KY','LA','ME','MD','MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ',
    'NM','NY','NC','ND','OH','OK','OR','PA','RI','SC','SD','TN','TX','UT','VT',
    'VA','WA','WV','WI','WY',
]
_CITIES = [
    'New York','Los Angeles','Chicago','Houston','Phoenix','Philadelphia',
    'San Antonio','San Diego','Dallas','San Jose','Austin','Jacksonville',
    'Fort Worth','Columbus','Charlotte','San Francisco','Indianapolis',
    'Seattle','Denver','Nashville','Milwaukee','Boston','Atlanta',
    'Miami','Minneapolis','Portland','Las Vegas','Louisville','Memphis',
]
_GENDERS = ['M', 'F', 'F', 'M', 'M', 'F', 'N']
_EDUCATION = ["High School", "Bachelor's", "Bachelor's", "Master's", "Master's", 'MBA', 'PhD']
_COMPANY_SIZE = ['Small (1-200)', 'Mid-Market (201-1000)', 'Large (1001-5000)', 'Enterprise (5000+)']
_INDUSTRIES_SHORT = [
    'Healthcare', 'Financial Services', 'Technology', 'Manufacturing',
    'Retail', 'Energy', 'Professional Services', 'Government', 'Education',
]
_LEVEL_SALARY = {
    'L1': (42, 68),   'L2': (65, 98),   'L3': (90, 135),
    'L4': (125, 185), 'L5': (175, 260), 'L6': (245, 400),
    'L7': (380, 600), 'L8': (550, 950), 'L9': (900, 2200),
}
_LEVEL_BONUS = {
    'L1': (3, 8),   'L2': (5, 12),   'L3': (8, 18),
    'L4': (12, 25), 'L5': (18, 35),  'L6': (25, 50),
    'L7': (35, 75), 'L8': (50, 100), 'L9': (75, 150),
}
_CEO_NAMES = [
    ('James','Morrison'),('Patricia','Chen'),('Michael','Rodriguez'),
    ('Jennifer','Williams'),('Robert','Patel'),('Linda','Johnson'),
    ('William','Kim'),('Barbara','Thompson'),('David','Martinez'),
    ('Susan','Anderson'),('Richard','Taylor'),('Jessica','Brown'),
    ('Thomas','White'),('Sarah','Davis'),('Charles','Wilson'),
]
_TICKERS = [
    'MMM','ABT','ABBV','ACN','ADM','APA','AAPL','AMZN','AXP','T',
    'ADSK','AZO','BAC','BA','BMY','CAT','CVX','CSCO','C','KO',
    'DHR','DD','XOM','GE','GS','HD','HON','IBM','INTC','JNJ',
]
_REMOTE = ['Fully Remote','Hybrid 2d/wk','Hybrid 3d/wk','Fully On-site','Flexible']
_LEAVE_REASONS = [
    'Compensation','Career Growth','Work-Life Balance',
    'Management Quality','Company Culture','Benefits','Job Security',
]
_ROLE_LEVELS = ['Individual Contributor','Team Lead','Manager','Director','VP / Executive']
_REGIONS = ['Northeast','Southeast','Midwest','Southwest','West','Pacific Northwest']
_SIZES_SURVEY = ['1-50','51-200','201-500','501-2000','2001-10000','10000+']


def _get_schema(slug):
    s = slug.lower()
    if any(k in s for k in ('ceo', 'executive-comp', 'executive-ben', 'exec-pay')):
        return 'ceo'
    if any(k in s for k in ('benefit', 'healthcare', 'health-care', 'retirement', 'retire')):
        return 'benefits'
    if any(k in s for k in ('satisfaction', 'engagement', 'survey-result')):
        return 'survey'
    return 'comp'


def _comp_row(rng, i, year, watermark):
    level = rng.choice(list(_LEVEL_SALARY.keys()))
    lo, hi = _LEVEL_SALARY[level]
    base = rng.randint(lo * 1000, hi * 1000)
    bon_lo, bon_hi = _LEVEL_BONUS[level]
    bonus_target = rng.randint(bon_lo, bon_hi)
    bonus_actual = round(bonus_target * rng.uniform(0.6, 1.3), 1)
    total_cash = round(base * (1 + bonus_actual / 100))
    equity = rng.randint(0, base // 2) if level in ('L5', 'L6', 'L7', 'L8', 'L9') else 0
    benefits_val = rng.randint(8000, 22000)
    return [
        f"EMP-{10000 + i}", rng.choice(_COMP_TITLES), level,
        rng.choice(_DEPARTMENTS), rng.choice(_INDUSTRIES_SHORT),
        rng.choice(_STATES), rng.choice(_CITIES),
        base, bonus_target, round(bonus_actual, 1), total_cash,
        equity, benefits_val, total_cash + equity + benefits_val,
        rng.randint(1, 28), rng.choice(_GENDERS), rng.choice(_EDUCATION),
        rng.choice(_COMPANY_SIZE), year, watermark,
    ]


def _ceo_row(rng, i, year, watermark):
    fn, ln = rng.choice(_CEO_NAMES)
    base = rng.randint(800_000, 3_000_000)
    bonus = rng.randint(500_000, 5_000_000)
    equity = rng.randint(1_000_000, 30_000_000)
    total = base + bonus + equity
    median_w = rng.randint(45_000, 95_000)
    revenue = rng.choice(['<$1B', '$1B–$5B', '$5B–$20B', '$20B–$50B', '$50B+'])
    co_type = rng.choice(['Corp', 'Industries', 'Group', 'Holdings', 'Inc'])
    co_adj = rng.choice(['American', 'National', 'United', 'Premier', 'Allied'])
    return [
        f"CO-{1000 + i}", rng.choice(_TICKERS),
        f"{co_adj} {rng.choice(_INDUSTRIES_SHORT)} {co_type}",
        rng.choice(_INDUSTRIES_SHORT), revenue,
        fn, ln, base, bonus, equity, total,
        median_w, round(total / median_w, 1),
        rng.randint(1, 20), rng.randint(7, 15),
        year, watermark,
    ]


def _benefits_row(rng, i, year, watermark):
    medical = rng.randint(6_000, 18_000)
    dental = rng.randint(400, 1_400)
    vision = rng.randint(80, 350)
    total = medical + dental + vision + rng.randint(500, 3000)
    return [
        f"CO-{2000 + i}", rng.choice(_INDUSTRIES_SHORT),
        rng.choice(['1-50', '51-200', '201-1000', '1001-5000', '5000+']),
        rng.choice(_STATES),
        medical, dental, vision,
        round(rng.uniform(2.0, 6.0), 1),
        rng.randint(10, 25), rng.randint(4, 24),
        rng.choice(_REMOTE),
        rng.choice([0, 0, 150, 250, 500, 1000]),
        rng.choice([0, 0, 2000, 5000, 10000]),
        total, year, watermark,
    ]


def _survey_row(rng, i, year, watermark):
    return [
        f"R-{100000 + i}",
        rng.choice(_INDUSTRIES_SHORT),
        rng.choice(_ROLE_LEVELS),
        rng.choice(_REGIONS),
        rng.choice(_SIZES_SURVEY),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        rng.choice(['Yes', 'No', 'Unsure']),
        rng.choice(['Yes', 'No', 'Unsure']),
        rng.choice(_LEAVE_REASONS),
        f"Q{rng.randint(1, 4)}", year, watermark,
    ]


_SCHEMA_HEADERS = {
    'comp': ['employee_id', 'title', 'level', 'department', 'industry', 'state', 'city',
             'base_salary', 'bonus_target_pct', 'bonus_actual_pct', 'total_cash',
             'equity_value', 'benefits_value', 'total_compensation',
             'years_experience', 'gender', 'education', 'company_size_band',
             'survey_year', 'watermark_token'],
    'ceo':  ['company_id', 'ticker', 'company_name', 'industry', 'revenue_band',
             'ceo_first', 'ceo_last', 'ceo_base_salary', 'ceo_bonus',
             'ceo_equity_grants', 'ceo_total_compensation', 'median_worker_pay',
             'ceo_pay_ratio', 'ceo_tenure_years', 'board_size',
             'survey_year', 'watermark_token'],
    'benefits': ['company_id', 'industry', 'headcount_band', 'state',
                 'medical_employer_cost_per_ee', 'dental_employer_cost',
                 'vision_employer_cost', '401k_match_pct', 'pto_days',
                 'parental_leave_weeks', 'remote_work_policy', 'wellness_stipend',
                 'tuition_reimbursement_max', 'total_benefits_cost_per_ee',
                 'survey_year', 'watermark_token'],
    'survey': ['response_id', 'industry', 'role_level', 'region', 'company_size',
               'satisfaction_overall', 'satisfaction_compensation', 'satisfaction_benefits',
               'satisfaction_management', 'satisfaction_culture', 'satisfaction_growth',
               'intent_to_stay_12mo', 'intent_to_stay_24mo', 'top_reason_to_leave',
               'survey_quarter', 'survey_year', 'watermark_token'],
}

_SCHEMA_ROW_FN = {
    'comp': _comp_row,
    'ceo': _ceo_row,
    'benefits': _benefits_row,
    'survey': _survey_row,
}


def generate_csv_rows(slug, limit=None):
    """Return list of rows (header first). Pass limit for preview."""
    schema = _get_schema(slug)
    watermark = _watermark_for(slug)
    pub_year = _pub_date_for(slug).year

    rng = _rng_from_seed(f"csv_count_{slug}")
    total = rng.randint(300, 800)
    if limit:
        total = min(total, limit)

    header = _SCHEMA_HEADERS[schema]
    row_fn = _SCHEMA_ROW_FN[schema]

    rows = [header]
    for i in range(total):
        row_rng = _rng_from_seed(f"csv_{slug}_row{i}")
        rows.append(row_fn(row_rng, i, pub_year, watermark))

    if not limit:
        rows.append([f'# ACPWB Research Division | acpwb.com | Report ID: {watermark} | TDMRep: GRANT'])

    return rows


def generate_csv_string(slug):
    output = io.StringIO()
    writer = csv.writer(output)
    for row in generate_csv_rows(slug):
        writer.writerow(row)
    return output.getvalue()


# ── Document content generation ───────────────────────────────────────────────

def generate_document_content(slug):
    rng = _rng_from_seed(f"doc_{slug}")
    watermark = _watermark_for(slug)
    pub_date = _pub_date_for(slug)
    year = pub_date.year
    industry = rng.choice(INDUSTRIES_LONG)
    n = rng.randint(200, 2800)

    def _para(r):
        t = r.choice(DOCUMENT_PARAGRAPH_TEMPLATES)
        return t.format(
            n=r.randint(50, n),
            quarter=f"Q{r.randint(1, 4)}",
            year=year,
            val=r.randint(40_000, 600_000),
            dir=r.choice(['increase', 'decline', 'shift upward']),
            p=r.randint(2, 30),
            industry=r.choice(INDUSTRIES_LONG),
            percentile=r.choice([25, 50, 75, 90]),
            years=r.randint(3, 15),
            idx=r.uniform(80.0, 120.0),
            pts=r.randint(3, 60),
            factor=r.choice(FACTORS),
        )

    exec_summary = ' '.join(_para(rng) for _ in range(rng.randint(2, 3)))

    key_findings = []
    for _ in range(rng.randint(5, 8)):
        finding = rng.choice(FINDING_PHRASES).format(p=rng.randint(3, 24))
        key_findings.append(finding.capitalize())

    section_pool = rng.sample(SECTION_HEADINGS, rng.randint(3, 5))
    sections = []
    for heading in section_pool:
        paragraphs = [_para(rng) for _ in range(rng.randint(2, 3))]
        sections.append({'heading': heading, 'paragraphs': paragraphs})

    methodology = (
        f"Survey data for this report was collected between {pub_date.strftime('%B')} {year - 1} "
        f"and {pub_date.strftime('%B %Y')} via ACPWB's annual compensation and workforce survey instrument. "
        f"A total of {n:,} responses were received from HR and finance professionals across "
        f"{rng.randint(22, 50)} states. Respondents represent organizations ranging from fewer than 50 employees "
        f"to more than 10,000 employees. Data has been weighted to reflect the distribution of U.S. employers "
        f"by size and industry as reported by the Bureau of Labor Statistics Quarterly Census of Employment and Wages. "
        f"Outliers beyond the 99th percentile have been excluded from aggregate calculations."
    )

    return {
        'exec_summary': exec_summary,
        'key_findings': key_findings,
        'sections': sections,
        'methodology': methodology,
        'appendix_orgs': rng.sample(APPENDIX_ORGS, rng.randint(5, 10)),
        'watermark_token': watermark,
        'year': year,
        'n': n,
    }
