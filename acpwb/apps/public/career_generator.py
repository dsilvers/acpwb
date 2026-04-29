import hashlib
import math
import random
import time
from datetime import date, timedelta

from django.utils.text import slugify

from .career_data import common
from .career_data import category_compensation_advisory as cat_comp
from .career_data import category_entry_level as cat_entry
from .career_data import category_coordinator as cat_coord
from .career_data import category_tech as cat_tech
from .career_data import category_office_support as cat_office

# (category_module, number_of_job_slots)
# IDs are assigned sequentially: comp=1-30, entry=31-50, coord=51-65, tech=66-90, office=91-100
# To add more jobs to a category: increment its slot count.
# Only append (grow last category or add new one) to avoid shifting existing job IDs.
CATEGORY_SLOTS = [
    (cat_comp,   30),
    (cat_entry,  20),
    (cat_coord,  15),
    (cat_tech,   25),
    (cat_office, 10),
]

_TOTAL_JOBS = sum(size for _, size in CATEGORY_SLOTS)


def _rng_from_seed(seed_str):
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


def _category_for_job(job_id: int):
    """Return the category module for a given job ID."""
    pos = 0
    for cat, size in CATEGORY_SLOTS:
        pos += size
        if job_id <= pos:
            return cat
    raise ValueError(f"job_id {job_id} out of range (max {_TOTAL_JOBS})")


def _build_base(cat, rng, title, department, mission_area):
    """Build the template format dict for a given category, consuming RNG as needed."""
    base = {
        'title': title,
        'department': department,
        'mission_area': mission_area,
    }
    # Compensation & advisory extras
    if hasattr(cat, 'CLIENT_TYPES'):
        base['client_type'] = rng.choice(cat.CLIENT_TYPES)
    if hasattr(cat, 'DELIVERABLE_TYPES'):
        base['deliverable'] = rng.choice(cat.DELIVERABLE_TYPES)
    if hasattr(cat, 'INDUSTRY_SECTORS'):
        base['industry'] = rng.choice(cat.INDUSTRY_SECTORS)
    if hasattr(cat, 'SURVEY_TOOLS'):
        base['tool'] = rng.choice(cat.SURVEY_TOOLS)
    # Technology extras
    if hasattr(cat, 'TECH_TOOLS'):
        base['tech_stack'] = rng.choice(cat.TECH_TOOLS)
    if hasattr(cat, 'SYSTEMS'):
        base['system'] = rng.choice(cat.SYSTEMS)
    if hasattr(cat, 'PLATFORMS'):
        base['platform'] = rng.choice(cat.PLATFORMS)
    # Entry-level extras
    if hasattr(cat, 'MENTOR_TITLES'):
        base['mentor_title'] = rng.choice(cat.MENTOR_TITLES)
    if hasattr(cat, 'PROGRAMS'):
        base['program'] = rng.choice(cat.PROGRAMS)
    # Coordinator extras
    if hasattr(cat, 'PROCESSES'):
        base['process'] = rng.choice(cat.PROCESSES)
    if hasattr(cat, 'TEAM_AREAS'):
        base['team_area'] = rng.choice(cat.TEAM_AREAS)
    return base


def generate_job(job_id: int) -> dict:
    """Generate a complete, deterministic job posting from an integer ID."""
    cat = _category_for_job(job_id)
    rng = _rng_from_seed(f"job_{job_id}")

    title = rng.choice(cat.JOB_TITLES)
    department = rng.choice(cat.DEPARTMENTS)
    location = rng.choice(common.LOCATIONS)
    job_type = rng.choice(common.JOB_TYPES)
    level, (sal_min_k, sal_max_k) = rng.choice(list(cat.LEVELS.items()))
    salary_min = (sal_min_k + rng.randint(-5, 5)) * 1000
    salary_max = (sal_max_k + rng.randint(-5, 5)) * 1000
    salary_min = max(salary_min, 44000)

    mission_area = rng.choice(cat.MISSION_AREAS)
    team_size = rng.randint(4, 22)
    years_exp = rng.randint(*cat.YEARS_EXP_RANGE)

    _base = _build_base(cat, rng, title, department, mission_area)

    # Posted date: Monday of current ISO week + small per-job offset (refreshes weekly)
    today = date.today()
    monday = today - timedelta(days=today.weekday())
    posting_offset = _rng_from_seed(
        f"job_posting_week_{today.isocalendar()[0]}_{today.isocalendar()[1]}_{job_id}"
    ).randint(0, 4)
    posted_date = monday + timedelta(days=posting_offset)

    deadline = posted_date + timedelta(days=rng.randint(18, 45))

    about_role_paras = [
        t.format(n=team_size, **_base)
        for t in rng.sample(cat.ABOUT_ROLE_TEMPLATES, rng.randint(2, 3))
    ]

    responsibilities = [
        t.format(n=team_size, **_base)
        for t in rng.sample(cat.RESPONSIBILITY_TEMPLATES, rng.randint(5, 7))
    ]

    requirements = [
        t.format(n=years_exp, **_base)
        for t in rng.sample(cat.REQUIREMENT_TEMPLATES, rng.randint(4, 6))
    ]

    preferred = rng.sample(cat.PREFERRED_TEMPLATES, rng.randint(3, 5))

    first_year_milestones = [
        t.format(**_base)
        for t in rng.sample(cat.FIRST_YEAR_MILESTONES, rng.randint(3, 5))
    ]

    reports_to = rng.choice(cat.REPORTS_TO_TEMPLATES).format(**_base)

    return {
        "id": job_id,
        "title": title,
        "slug": slugify(title),
        "department": department,
        "location": location,
        "job_type": job_type,
        "level": level,
        "salary_min": salary_min,
        "salary_max": salary_max,
        "salary_min_display": f"{salary_min:,}",
        "salary_max_display": f"{salary_max:,}",
        "posted_date": posted_date,
        "deadline": deadline,
        "about_role_paras": about_role_paras,
        "responsibilities": responsibilities,
        "requirements": requirements,
        "preferred": preferred,
        "why_acpwb": rng.choice(common.WHY_ACPWB_TEMPLATES),
        "interview_process": rng.choice(common.INTERVIEW_PROCESS_TEMPLATES),
        "team_size": team_size,
        "first_year_milestones": first_year_milestones,
        "reports_to": reports_to,
    }


def get_current_job_ids() -> list[int]:
    """Return active job IDs for the current hour: 2–3 from each category."""
    hour_bucket = math.floor(time.time() / 3600)
    result = []
    lo = 1
    for i, (cat, size) in enumerate(CATEGORY_SLOTS):
        hi = lo + size
        cat_rng = _rng_from_seed(f"careers_cat_{i}_{hour_bucket}")
        count = cat_rng.randint(2, 3)
        result.extend(cat_rng.sample(range(lo, hi), count))
        lo = hi
    _rng_from_seed(f"careers_order_{hour_bucket}").shuffle(result)
    return result


def get_current_jobs() -> list[dict]:
    """Return full job objects for the current hourly listing."""
    return [generate_job(jid) for jid in get_current_job_ids()]


def is_valid_job_id(job_id: int) -> bool:
    return 1 <= job_id <= _TOTAL_JOBS
