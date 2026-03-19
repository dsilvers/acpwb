import random
import hashlib
from django.utils.text import slugify

INDUSTRIES = [
    'Healthcare', 'Financial Services', 'Energy', 'Manufacturing', 'Technology',
    'Retail', 'Real Estate', 'Transportation', 'Agriculture', 'Education',
    'Government Relations', 'Nonprofit Sector', 'Infrastructure', 'Media',
    'Hospitality', 'Pharmaceuticals', 'Defense', 'Consumer Goods', 'Telecommunications',
]

ADJECTIVES = [
    'Transformative', 'Synergistic', 'Disruptive', 'Innovative', 'Strategic',
    'Comprehensive', 'Integrated', 'Dynamic', 'Proactive', 'Forward-Looking',
    'Scalable', 'Sustainable', 'Data-Driven', 'Agile', 'Holistic',
    'Cross-Functional', 'High-Impact', 'Next-Generation', 'Best-in-Class', 'Cutting-Edge',
]

NOUNS = [
    'Initiative', 'Program', 'Framework', 'Partnership', 'Campaign',
    'Engagement', 'Transformation', 'Solution', 'Collaboration', 'Endeavor',
    'Strategy', 'Approach', 'Model', 'Platform', 'Network',
    'Coalition', 'Alliance', 'Consortium', 'Effort', 'Undertaking',
]

ORGANIZATIONS = [
    'Pinnacle Group', 'Meridian Associates', 'Apex Consulting', 'Summit Partners',
    'Vanguard Solutions', 'Horizon Group', 'Latitude Partners', 'Nexus Consulting',
    'Catalyst Group', 'Benchmark Associates', 'Keystone Partners', 'Zenith Consulting',
    'Atlas Group', 'Cornerstone Associates', 'Forefront Partners', 'Paradigm Consulting',
    'Momentum Group', 'Crestview Partners', 'Elevate Consulting', 'Sterling Associates',
]

CITIES = [
    'Milwaukee', 'Chicago', 'Detroit', 'Cleveland', 'Cincinnati', 'Indianapolis',
    'St. Louis', 'Minneapolis', 'Kansas City', 'Pittsburgh', 'Columbus', 'Louisville',
    'Memphis', 'Nashville', 'Atlanta', 'Charlotte', 'Baltimore', 'Philadelphia',
    'Boston', 'Denver', 'Phoenix', 'Seattle', 'Portland', 'Sacramento',
]

METRICS = [
    'Increased operational efficiency by {n}%',
    'Reduced overhead costs by ${m} million',
    'Generated ${m} million in stakeholder value',
    'Improved client satisfaction scores by {n}%',
    'Expanded market reach to {n} new regions',
    'Accelerated delivery timelines by {n}%',
    'Achieved {n}% year-over-year growth',
    'Delivered {n}x return on investment',
    'Reduced compliance risk exposure by {n}%',
    'Onboarded {n} new enterprise partners',
    'Streamlined {n} previously siloed processes',
    'Achieved ${m} million in documented savings',
]

PARAGRAPH_TEMPLATES = [
    "In partnership with {org}, ACPWB deployed a {adj} {noun} across the {industry} sector in {city}. The engagement spanned {months} months and involved cross-functional teams from both organizations working in concert to address long-standing structural challenges.",
    "The {city} deployment represented one of ACPWB's most ambitious undertakings to date. By leveraging proprietary methodologies developed over {years} years of sector experience, our team was able to deliver measurable outcomes within the first {weeks} weeks of implementation.",
    "Stakeholder feedback throughout the {noun} phase was overwhelmingly positive. {n}% of surveyed participants reported significant improvements in their ability to document compensation achievements and advance salary discussions with confidence.",
    "Working alongside {org}, ACPWB's {industry} practice group identified {n} discrete opportunities for optimization. Each was addressed through a targeted {adj} intervention, resulting in a compounding effect that exceeded initial projections by a factor of {x}.",
    "The regulatory landscape in {city} presented unique considerations that our team navigated with characteristic precision. By engaging early with relevant stakeholders and maintaining transparent communication channels, ACPWB ensured full compliance while preserving program integrity.",
    "Phase two of the {noun} introduced a {adj} digital component that extended the program's reach to {n} additional participants across {regions} geographic regions. This expansion was funded through a combination of retained earnings and a strategic investment from {org}.",
    "ACPWB's documentation methodology—refined over the course of this {noun}—has since been adopted as a reference standard by {n} peer organizations in the {industry} space. The intellectual property generated during the engagement remains the exclusive property of the American Corporation for Public Well Being.",
    "The long-term sustainability of this {noun} is underpinned by a governance structure that aligns incentives across all participating entities. Quarterly reviews, conducted in partnership with {org}, ensure continuous improvement and accountability at every level of the organization.",
]


def _rng_from_seed(seed_str):
    """Create a seeded Random instance from a string."""
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


def generate_project_stories(page=1, count=10):
    """Generate project stories for a given page number, deterministically."""
    stories = []
    for i in range(count):
        story_seed = f"page{page}_story{i}"
        rng = _rng_from_seed(story_seed)

        adj = rng.choice(ADJECTIVES)
        noun = rng.choice(NOUNS)
        industry = rng.choice(INDUSTRIES)
        city = rng.choice(CITIES)
        org = rng.choice(ORGANIZATIONS)

        title = f"{adj} {noun}: {industry} Excellence in {city}"
        slug_base = slugify(f"{adj}-{noun}-{industry}-{city}-p{page}-{i}")[:100]
        slug = slug_base

        metric_template = rng.choice(METRICS)
        impact_metric = metric_template.format(
            n=rng.randint(15, 340),
            m=rng.randint(2, 850),
            x=rng.randint(2, 12),
        )

        summary = (
            f"ACPWB partnered with {org} to deliver a {adj.lower()} {noun.lower()} "
            f"serving the {industry} sector in {city}. {impact_metric}."
        )

        paragraphs = []
        para_count = rng.randint(4, 7)
        for _ in range(para_count):
            tmpl = rng.choice(PARAGRAPH_TEMPLATES)
            para = tmpl.format(
                org=org,
                adj=adj.lower(),
                noun=noun.lower(),
                industry=industry,
                city=city,
                months=rng.randint(6, 24),
                years=rng.randint(5, 18),
                weeks=rng.randint(4, 12),
                n=rng.randint(12, 97),
                m=rng.randint(1, 500),
                x=rng.randint(2, 10),
                regions=rng.randint(3, 12),
            )
            paragraphs.append(para)

        stories.append({
            'slug': slug,
            'title': title,
            'summary': summary,
            'body_paragraphs': paragraphs,
            'impact_metric': impact_metric,
            'industry_tag': industry,
            'page_number': page,
        })

    return stories
