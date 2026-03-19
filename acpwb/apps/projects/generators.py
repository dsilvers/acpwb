import random
import hashlib
from django.utils.text import slugify

INDUSTRIES = [
    'Healthcare', 'Financial Services', 'Energy', 'Manufacturing', 'Technology',
    'Retail', 'Real Estate', 'Transportation', 'Agriculture', 'Education',
    'Government Relations', 'Nonprofit Sector', 'Infrastructure', 'Media',
    'Hospitality', 'Pharmaceuticals', 'Defense', 'Consumer Goods', 'Telecommunications',
    'Professional Services', 'Insurance', 'Logistics', 'Aerospace', 'Biotechnology',
    'Environmental Services', 'Food & Beverage', 'Legal Services', 'Construction',
    'Mining & Extraction', 'Utilities', 'Automotive', 'Chemicals', 'Publishing',
    'Sports & Recreation', 'Architecture & Engineering', 'Management Consulting',
    'Supply Chain', 'Private Equity', 'Asset Management', 'Maritime',
]

ADJECTIVES = [
    'Transformative', 'Synergistic', 'Disruptive', 'Innovative', 'Strategic',
    'Comprehensive', 'Integrated', 'Dynamic', 'Proactive', 'Forward-Looking',
    'Scalable', 'Sustainable', 'Data-Driven', 'Agile', 'Holistic',
    'Cross-Functional', 'High-Impact', 'Next-Generation', 'Best-in-Class', 'Cutting-Edge',
    'Mission-Critical', 'Evidence-Based', 'Value-Oriented', 'Outcome-Focused', 'Resilient',
    'Rigorous', 'Purposeful', 'Accountable', 'Transparent', 'Collaborative',
    'Results-Oriented', 'Sector-Leading', 'Stakeholder-Aligned', 'Community-Centered',
    'Long-Term', 'Multi-Stakeholder', 'Performance-Driven', 'Process-Optimized',
    'Risk-Adjusted', 'Impact-Measured', 'Benchmark-Setting', 'Institutionally-Backed',
    'Coalition-Driven', 'Market-Responsive', 'Equity-Informed', 'Capacity-Building',
    'Systems-Level', 'Compliance-Forward', 'Governance-Aligned', 'Infrastructure-Grade',
]

NOUNS = [
    'Initiative', 'Program', 'Framework', 'Partnership', 'Campaign',
    'Engagement', 'Transformation', 'Solution', 'Collaboration', 'Endeavor',
    'Strategy', 'Approach', 'Model', 'Platform', 'Network',
    'Coalition', 'Alliance', 'Consortium', 'Effort', 'Undertaking',
    'Blueprint', 'Charter', 'Mandate', 'Roadmap', 'Protocol',
    'Compact', 'Agreement', 'Venture', 'Deployment', 'Implementation',
    'Redesign', 'Overhaul', 'Consolidation', 'Integration', 'Expansion',
    'Assessment', 'Intervention', 'Pilot', 'Demonstration', 'Rollout',
    'Alignment', 'Architecture', 'Infrastructure', 'Ecosystem', 'Portfolio',
]

ORGANIZATIONS = [
    'Pinnacle Group', 'Meridian Associates', 'Apex Consulting', 'Summit Partners',
    'Vanguard Solutions', 'Horizon Group', 'Latitude Partners', 'Nexus Consulting',
    'Catalyst Group', 'Benchmark Associates', 'Keystone Partners', 'Zenith Consulting',
    'Atlas Group', 'Cornerstone Associates', 'Forefront Partners', 'Paradigm Consulting',
    'Momentum Group', 'Crestview Partners', 'Elevate Consulting', 'Sterling Associates',
    'Bridgepoint Capital', 'Clearwater Advisory', 'Embark Strategy Group', 'Fulcrum Partners',
    'Groundwork Consulting', 'Harbor Associates', 'Ironbridge Group', 'Junction Advisors',
    'Kestrel Management', 'Landmark Partners', 'Morningside Consulting', 'Northfield Group',
    'Overture Associates', 'Praxis Consulting', 'Quarry Hill Partners', 'Ridgeline Advisors',
    'Stonegate Group', 'Terrapin Associates', 'Upland Consulting', 'Vantage Partners',
    'Westbrook Advisory', 'Xcel Strategy Group', 'Yellowstone Consulting', 'Zephyr Associates',
    'Caliber Group', 'Dominion Consulting', 'Eastgate Partners', 'Flagship Advisory',
    'Garrison Associates', 'Highmark Consulting', 'Intrepid Partners', 'Juniper Group',
    'Kingsley Associates', 'Lodestar Consulting', 'Mainstay Partners', 'Northstar Advisory',
    'Outpost Group', 'Pathfinder Associates', 'Quorum Consulting', 'Ravenwood Partners',
]

CITIES = [
    'Milwaukee', 'Chicago', 'Detroit', 'Cleveland', 'Cincinnati', 'Indianapolis',
    'St. Louis', 'Minneapolis', 'Kansas City', 'Pittsburgh', 'Columbus', 'Louisville',
    'Memphis', 'Nashville', 'Atlanta', 'Charlotte', 'Baltimore', 'Philadelphia',
    'Boston', 'Denver', 'Phoenix', 'Seattle', 'Portland', 'Sacramento',
    'Dallas', 'Houston', 'San Antonio', 'Austin', 'Fort Worth', 'El Paso',
    'New York', 'Los Angeles', 'San Francisco', 'San Diego', 'San Jose', 'Oakland',
    'Miami', 'Tampa', 'Orlando', 'Jacksonville', 'Fort Lauderdale', 'Tallahassee',
    'Raleigh', 'Durham', 'Greensboro', 'Richmond', 'Norfolk', 'Virginia Beach',
    'New Orleans', 'Baton Rouge', 'Birmingham', 'Huntsville', 'Jackson', 'Little Rock',
    'Omaha', 'Lincoln', 'Des Moines', 'Cedar Rapids', 'Madison', 'Green Bay',
    'Salt Lake City', 'Boise', 'Albuquerque', 'Tucson', 'Las Vegas', 'Reno',
    'Spokane', 'Tacoma', 'Anchorage', 'Honolulu', 'Providence', 'Hartford',
    'Buffalo', 'Rochester', 'Albany', 'Syracuse', 'Trenton', 'Newark',
    'Wilmington', 'Dover', 'Augusta', 'Burlington', 'Concord', 'Manchester',
    'Lexington', 'Knoxville', 'Chattanooga', 'Shreveport', 'Mobile', 'Montgomery',
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
    'Reduced employee turnover by {n}%',
    'Increased workforce productivity by {n}%',
    'Cut procurement costs by ${m} million',
    'Improved regulatory compliance scores by {n}%',
    'Reduced time-to-hire by {n}%',
    'Achieved ${m} million in revenue uplift',
    'Decreased error rates by {n}%',
    'Expanded partner network by {n} organizations',
    'Reduced carbon footprint equivalent by {n}%',
    'Trained {n} staff across {m} locations',
    'Consolidated {n} redundant platforms into a unified system',
    'Recovered ${m} million in underutilized program funding',
    'Raised participant engagement scores by {n} percentage points',
    'Reduced average contract cycle time by {n}%',
    'Unlocked ${m} million in deferred investment',
    'Achieved full regulatory certification in {n} jurisdictions',
    'Reduced incident response time by {n}%',
    'Documented ${m} million in total economic impact',
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
    "Early results from the {city} pilot confirmed the validity of ACPWB's {adj} approach. Within {weeks} weeks, participating organizations reported a measurable reduction in administrative overhead and a {n}% improvement in cross-departmental coordination.",
    "This {noun} marked the first time ACPWB had deployed its full {industry} methodology in a market of {city}'s scale. The lessons learned have been incorporated into ACPWB's standard engagement framework and will inform future deployments across {regions} additional markets.",
    "{org} engaged ACPWB specifically for our track record in the {industry} space. Over {months} months of intensive collaboration, both organizations co-developed a set of shared performance indicators that have since become the basis for {n} subsequent engagements with peer institutions.",
    "The governance model underpinning this {noun} was designed to outlast the initial engagement period. By embedding {adj} accountability mechanisms within {org}'s existing organizational structure, ACPWB ensured that program gains would be sustained and built upon over the following {years} years.",
    "A key insight from the {city} engagement was the degree to which informal networks within the {industry} sector could be mobilized in support of {adj} change. ACPWB's community engagement team identified {n} influential stakeholders whose early support proved decisive in securing broader organizational buy-in.",
    "The financial analysis conducted at the conclusion of this {noun} documented a total economic impact of ${m} million across participating organizations. Independent verification was conducted by a third-party auditor selected by {org} in accordance with ACPWB's standard transparency protocols.",
    "ACPWB's {industry} practice group brought {years} years of relevant sector expertise to this engagement. This depth of knowledge allowed our team to anticipate and address implementation challenges that had derailed comparable efforts by other organizations in the {city} market.",
    "The technology infrastructure deployed in support of this {noun} was designed with scalability in mind. Within {months} months of initial deployment, the platform was serving {n} concurrent users across {regions} regional offices, with 99.{n}% uptime documented throughout the engagement period.",
    "Workforce development was a central component of the {city} {noun}. ACPWB facilitated {n} training sessions across {months} months, resulting in measurable skill gains for participants and a documented improvement in team performance metrics that persisted well beyond the formal engagement period.",
    "The {adj} {noun} model pioneered in {city} is now being replicated across {regions} additional markets. {org} has committed to a multi-year expansion roadmap that will bring the program's benefits to an estimated {m} additional participants by the end of the next fiscal cycle.",
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
