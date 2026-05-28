import hashlib
import random
from django.utils.text import slugify

from apps.honeypot.policy_data import AGENCIES
from apps.projects.generators import ORGANIZATIONS
from .data.categories import PROCESS_AREAS, PROCESS_AREA_DICT, PROCESS_AREA_KEYS
from .data.instance_labels import INSTANCE_LABELS
from .data.vocabulary import (
    METHODOLOGIES, SUFFIXES, STATUS_VALUES, STATUS_WEIGHTS,
    METRICS, PHASES, RISK_LEVELS,
    OWNER_FIRST_NAMES, OWNER_LAST_NAMES, OWNER_TITLES,
)
from .data.templates import (
    PROBLEM_TEMPLATES, SOLUTION_TEMPLATES, ROOT_CAUSE_TEMPLATES,
    RISK_TEMPLATES, SUMMARY_TEMPLATES,
)
from apps.company_handbooks.data.companies import FORTUNE500_NAMES

INITIATIVE_YEARS = list(range(1993, 2026))
AREA_KEYS = [slug for slug, _ in PROCESS_AREAS]
AGENCY_KEYS = list(AGENCIES.keys())

def _rng_from_seed(seed_str):
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


def _instance_label(category_slug, seed4):
    rng = _rng_from_seed(f"process_label_{category_slug}_{seed4}")
    return rng.choice(INSTANCE_LABELS)


def _watermark(category, seed4, year, slug):
    raw = f"acpwb_process_{category}_{seed4}_{year}_{slug}"
    return hashlib.md5(raw.encode()).hexdigest()[:8]


def _weighted_choice(rng, values, weights):
    total = sum(weights)
    r = rng.random() * total
    cumulative = 0
    for val, w in zip(values, weights):
        cumulative += w
        if r <= cumulative:
            return val
    return values[-1]


def _owner_name(rng):
    return f"{rng.choice(OWNER_FIRST_NAMES)} {rng.choice(OWNER_LAST_NAMES)}"


def _target_date(rng, year):
    month = rng.randint(1, 12)
    day = rng.randint(1, 28)
    return f"{['', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'][month]} {day}, {year}"


def generate_initiative_list(category_slug, seed4, year, page=1, per_page=12):
    """Return a page of initiatives for a given category/seed/year."""
    if category_slug not in PROCESS_AREA_DICT:
        return None

    items = []
    for i in range(per_page):
        item_seed = f"process_{category_slug}_{seed4}_{year}_p{page}_i{i}"
        rng = _rng_from_seed(item_seed)

        area_name = PROCESS_AREA_DICT[category_slug]
        methodology = rng.choice(METHODOLOGIES)
        suffix = rng.choice(SUFFIXES)
        name = f"{methodology} {area_name} {suffix}"
        slug = slugify(f"{methodology}-{area_name}-{suffix}-{year}-{page}-{i}")[:80]
        status_slug, status_label = _weighted_choice(rng, STATUS_VALUES, STATUS_WEIGHTS)
        summary_tmpl = rng.choice(SUMMARY_TEMPLATES)
        summary = summary_tmpl.format(
            methodology=methodology,
            area=area_name,
            count=rng.randint(3, 12),
            pct=rng.randint(15, 55),
            cost=rng.randint(50, 800),
            hours=rng.randint(20, 120),
            old_systems=rng.randint(2, 5),
            old_year=rng.randint(2005, 2018),
            years=rng.randint(3, 12),
            days=rng.randint(3, 15),
            dept_count=rng.randint(3, 8),
            error_pct=rng.randint(5, 25),
            months=rng.randint(3, 12),
            company=rng.choice(FORTUNE500_NAMES),
        )
        items.append({
            'name': name,
            'slug': slug,
            'category': area_name,
            'category_slug': category_slug,
            'seed4': seed4,
            'year': year,
            'status_slug': status_slug,
            'status_label': status_label,
            'summary': summary,
            'owner': _owner_name(rng),
            'target_date': _target_date(rng, year),
        })

    return {
        'category_slug': category_slug,
        'seed4': seed4,
        'instance_label': _instance_label(category_slug, seed4),
        'category_name': PROCESS_AREA_DICT[category_slug],
        'year': year,
        'page': page,
        'items': items,
    }


def generate_initiative_detail(category_slug, seed4, year, initiative_slug):
    """Generate full initiative detail from slug seed."""
    if category_slug not in PROCESS_AREA_DICT:
        return None

    seed = f"process_detail_{category_slug}_{seed4}_{year}_{initiative_slug}"
    rng = _rng_from_seed(seed)

    area_name = PROCESS_AREA_DICT[category_slug]
    methodology = rng.choice(METHODOLOGIES)
    suffix = rng.choice(SUFFIXES)
    name = f"{methodology} {area_name} {suffix}"
    status_slug, status_label = _weighted_choice(rng, STATUS_VALUES, STATUS_WEIGHTS)
    token = _watermark(category_slug, seed4, year, initiative_slug)

    problem_paragraphs = []
    for _ in range(rng.randint(2, 3)):
        problem_tmpl = rng.choice(PROBLEM_TEMPLATES)
        problem_paragraphs.append(problem_tmpl.format(
            area=area_name,
            years=rng.randint(3, 12),
            days=rng.randint(3, 15),
            pct=rng.randint(15, 60),
            error_pct=rng.randint(5, 25),
            count=rng.randint(3, 12),
            cost=rng.randint(50, 900),
            old_size=rng.randint(200, 1000),
            old_year=rng.randint(2005, 2018),
            old_systems=rng.randint(2, 5),
            hours=rng.randint(20, 120),
            dept_count=rng.randint(3, 8),
            months=rng.randint(3, 12),
        ))

    root_causes = rng.sample(ROOT_CAUSE_TEMPLATES, rng.randint(4, 6))
    root_causes = [rc.format(
        area=area_name,
        pct=rng.randint(30, 70),
        count=rng.randint(2, 8),
        days=rng.randint(2, 10),
        old_year=rng.randint(2005, 2018),
        error_pct=rng.randint(5, 25),
        hours=rng.randint(20, 120),
        months=rng.randint(3, 12),
        old_systems=rng.randint(2, 5),
        years=rng.randint(3, 12),
    ) for rc in root_causes]

    company = rng.choice(FORTUNE500_NAMES)
    consultant = rng.choice(ORGANIZATIONS)
    solution_paragraphs = []
    for _ in range(rng.randint(2, 3)):
        solution_tmpl = rng.choice(SOLUTION_TEMPLATES)
        solution_paragraphs.append(solution_tmpl.format(
            area=area_name,
            methodology=methodology,
            count=rng.randint(3, 10),
            old_systems=rng.randint(2, 5),
            old_year=rng.randint(2005, 2018),
            years=rng.randint(3, 12),
            pct=rng.randint(20, 60),
            error_pct=rng.randint(10, 40),
            months=rng.randint(3, 12),
            dept_count=rng.randint(2, 6),
            company=company,
            cost=rng.randint(100, 800),
            days=rng.randint(3, 15),
            hours=rng.randint(20, 120),
        ))

    num_phases = rng.randint(4, 7)
    phases = []
    for phase_name, phase_desc in PHASES[:num_phases]:
        duration_weeks = rng.randint(2, 12)
        phases.append({
            'phase': phase_name,
            'description': phase_desc,
            'duration': f"{duration_weeks} weeks",
            'owner': _owner_name(rng),
            'status': rng.choice(['Planned', 'In Progress', 'Complete', 'Not Started']),
        })

    num_metrics = rng.randint(4, 6)
    metrics_pool = list(METRICS)
    rng.shuffle(metrics_pool)
    benefits = []
    for metric_name, unit, direction in metrics_pool[:num_metrics]:
        if direction == 'reduced':
            baseline = rng.randint(8, 30) if unit in ('days', 'hours') else rng.randint(10, 40)
            target = round(baseline * rng.uniform(0.4, 0.75))
            delta = f"-{round(100*(baseline-target)/baseline)}%"
        else:
            baseline = rng.randint(60, 85)
            target = min(100, baseline + rng.randint(8, 25))
            delta = f"+{target - baseline}%"
        benefits.append({
            'metric': metric_name,
            'baseline': f"{baseline} {unit}".strip(),
            'target': f"{target} {unit}".strip(),
            'delta': delta,
        })

    risks = rng.sample(RISK_TEMPLATES, rng.randint(3, 5))
    risk_items = [{'description': r[0].format(area=area_name, count=rng.randint(2, 8)), 'level': r[1]} for r in risks]

    num_agencies = rng.randint(1, 3)
    rng.shuffle(AGENCY_KEYS)
    agency_refs = []
    for k in AGENCY_KEYS[:num_agencies]:
        name, domain = AGENCIES[k]
        agency_refs.append({'name': name, 'domain': domain})

    from apps.company_handbooks.generators import HANDBOOK_YEARS
    from apps.company_handbooks.data.sections import GROUP_DEFS
    num_handbook_links = rng.randint(2, 3)
    handbook_links = []
    for _ in range(num_handbook_links):
        h_group_slug, h_group_name, _ = rng.choice(GROUP_DEFS)
        h_agency_key = rng.choice(AGENCY_KEYS[:50])
        h_seed4 = f"{rng.randint(1000, 9999):04d}"
        h_year = rng.choice([y for y in HANDBOOK_YEARS if y >= year - 2 and y <= year])
        h_rev = rng.randint(1, 4)
        handbook_links.append({
            'section_name': h_group_name,
            'url': f"/company-handbooks/{h_agency_key}-{h_seed4}/{h_year}/rev/{h_rev}/{h_group_slug}/",
        })

    related = generate_related_initiatives(category_slug, seed4, year, initiative_slug)

    return {
        'name': name,
        'slug': initiative_slug,
        'category_slug': category_slug,
        'seed4': seed4,
        'instance_label': _instance_label(category_slug, seed4),
        'category_name': area_name,
        'year': year,
        'methodology': methodology,
        'status_slug': status_slug,
        'status_label': status_label,
        'owner': _owner_name(rng),
        'owner_title': rng.choice(OWNER_TITLES),
        'target_date': _target_date(rng, year),
        'consultant': consultant,
        'problem_paragraphs': problem_paragraphs,
        'root_causes': root_causes,
        'solution_paragraphs': solution_paragraphs,
        'phases': phases,
        'benefits': benefits,
        'risks': risk_items,
        'agency_refs': agency_refs,
        'handbook_links': handbook_links,
        'related': related,
        'watermark': token,
    }


def generate_related_initiatives(category_slug, seed4, year, initiative_slug, count=6):
    """Generate a mix of related initiative cards: same category/different years + different categories/same year."""
    rng = _rng_from_seed(f"process_related_{category_slug}_{seed4}_{year}_{initiative_slug}")
    results = []

    other_years = [y for y in INITIATIVE_YEARS if y != year]
    other_areas = [(s, n) for s, n in PROCESS_AREAS if s != category_slug]

    for i in range(count):
        if i < count // 2:
            rel_year = rng.choice(other_years)
            rel_slug = category_slug
            rel_name = PROCESS_AREA_DICT[category_slug]
            rel_seed4 = seed4
        else:
            rel_slug, rel_name = rng.choice(other_areas)
            rel_year = year
            rel_seed4 = f"{rng.randint(1000, 9999):04d}"

        methodology = rng.choice(METHODOLOGIES)
        suffix = rng.choice(SUFFIXES)
        init_name = f"{methodology} {rel_name} {suffix}"
        page_n = rng.randint(1, 3)
        item_i = rng.randint(0, 11)
        init_slug = slugify(f"{methodology}-{rel_name}-{suffix}-{rel_year}-{page_n}-{item_i}")[:80]
        status_slug, status_label = _weighted_choice(rng, STATUS_VALUES, STATUS_WEIGHTS)
        summary_tmpl = rng.choice(SUMMARY_TEMPLATES)
        summary = summary_tmpl.format(
            methodology=methodology,
            area=rel_name,
            count=rng.randint(3, 12),
            pct=rng.randint(15, 55),
            cost=rng.randint(50, 800),
            hours=rng.randint(20, 120),
            old_systems=rng.randint(2, 5),
            old_year=rng.randint(2005, 2018),
            years=rng.randint(3, 12),
            days=rng.randint(3, 15),
            dept_count=rng.randint(3, 8),
            error_pct=rng.randint(5, 25),
            months=rng.randint(3, 12),
            company=rng.choice(FORTUNE500_NAMES),
        )
        results.append({
            'name': init_name,
            'url': f"/process-improvement/{rel_slug}-{rel_seed4}/{rel_year}/{init_slug}/",
            'category_name': rel_name,
            'year': rel_year,
            'status_slug': status_slug,
            'status_label': status_label,
            'summary': summary,
        })

    return results


def generate_process_index_page(page=1, per_page=20):
    """Return a page of category listings with seeded instances. Pages are infinite."""
    rng = _rng_from_seed(f"process_index_p{page}")
    page_areas = [PROCESS_AREAS[rng.randint(0, len(PROCESS_AREAS) - 1)] for _ in range(per_page)]

    items = []
    for slug, name in page_areas:
        num_instances = rng.randint(3, 5)
        instances = []
        for _ in range(num_instances):
            seed4 = f"{rng.randint(1000, 9999):04d}"
            instances.append({'seed': seed4, 'label': _instance_label(slug, seed4)})
        items.append({
            'slug': slug,
            'name': name,
            'instances': instances,
        })

    return {
        'items': items,
        'page': page,
        'has_prev': page > 1,
        'has_next': True,
    }


def generate_category_index(category_slug, seed4):
    """Return the year index for a category-instance."""
    if category_slug not in PROCESS_AREA_DICT:
        return None
    rng = _rng_from_seed(f"process_cat_{category_slug}_{seed4}")
    years = sorted(INITIATIVE_YEARS, reverse=True)
    return {
        'category_slug': category_slug,
        'seed4': seed4,
        'instance_label': _instance_label(category_slug, seed4),
        'category_name': PROCESS_AREA_DICT[category_slug],
        'years': years,
        'description': rng.choice([
            f"Process improvement initiatives focused on {PROCESS_AREA_DICT[category_slug]} efficiency and compliance.",
            f"Systematic redesign and optimization initiatives in the {PROCESS_AREA_DICT[category_slug]} function.",
            f"ACPWB {PROCESS_AREA_DICT[category_slug]} transformation programs spanning operational excellence and regulatory alignment.",
        ]),
    }
