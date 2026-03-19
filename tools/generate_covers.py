"""
Generate corporate stock images for reports and projects using SDXL on Apple Silicon (MPS).

Usage:
    pip install torch diffusers transformers accelerate pillow
    python tools/generate_covers.py
    python tools/generate_covers.py --type reports --count 60
    python tools/generate_covers.py --type projects --count 60
    python tools/generate_covers.py --type both --count 100

Outputs WebP images to:
    acpwb/static/img/report-covers/   (named by report slug)
    acpwb/static/img/project-covers/  (named 000.webp, 001.webp, ...)
"""

import argparse
import json
import random
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent
ACPWB_ROOT = REPO_ROOT / "acpwb"
REPORT_COVERS_DIR = ACPWB_ROOT / "static" / "img" / "report-covers"
PROJECT_COVERS_DIR = ACPWB_ROOT / "static" / "img" / "project-covers"

MODEL_ID = "stabilityai/stable-diffusion-xl-base-1.0"

# ── Report prompts by category ────────────────────────────────────────────────

REPORT_CATEGORY_PROMPTS = {
    "Compensation": [
        "professional business meeting in sleek conference room, people reviewing salary charts on laptops, overhead view of desk with financial documents and graphs, corporate office, warm natural light",
        "close-up of financial data charts and bar graphs on a glass desk, office environment with city view, depth of field blur, corporate documentary photography",
        "executive boardroom discussion, three professionals in business attire examining printed reports, navy and gold color scheme, editorial photography",
        "hands exchanging professional documents across polished conference table, financial charts visible, soft office lighting, corporate stock photography",
        "modern open-plan office with professionals at standing desks, large monitors showing compensation dashboards, clean corporate aesthetic",
    ],
    "Workforce": [
        "diverse team of professionals in bright modern office, standing around a glass whiteboard with diagrams, corporate documentary style",
        "aerial view of busy corporate open office floor, rows of workstations, natural light through floor-to-ceiling windows, editorial photography",
        "small team meeting around a circular table in a glass-walled conference room, laptops and notebooks, professional corporate atmosphere",
        "two professionals shaking hands in a lobby with marble floors, business attire, natural light, corporate editorial",
        "employees collaborating around a large monitor showing workforce analytics data, modern office, shallow depth of field",
    ],
    "Governance": [
        "empty boardroom with long polished mahogany table and leather chairs, city skyline visible through windows, golden hour light",
        "close-up of a gavel on a polished wood desk with blurred legal documents in background, corporate governance concept",
        "row of professional executives seated at a panel table, formal setting, directional studio lighting, editorial photograph",
        "stacks of organized corporate documents with a pen and reading glasses on a mahogany desk, warm office lighting",
        "glass and steel corporate headquarters exterior shot at dusk, reflections on windows, architectural photography",
    ],
    "ESG": [
        "modern glass office building with green living wall, solar panels on roof visible, blue sky, architectural photography",
        "professional in business attire reviewing sustainability metrics on tablet with urban greenery in background, natural light",
        "aerial view of corporate campus with manicured grounds and trees, clean geometric architecture, drone photography",
        "wind turbines visible through the floor-to-ceiling window of a modern executive office, symbolic corporate environmental photo",
        "team of diverse professionals planting a tree near their corporate campus, branded corporate social responsibility photography",
    ],
    "Benefits": [
        "bright modern employee wellness center inside a corporate campus, yoga mats and natural light, lifestyle photography",
        "HR professional presenting employee benefits on a large screen in a well-lit conference room, attentive audience",
        "close-up of hands filling out benefit enrollment forms on a clean white desk, soft office lighting, documentary",
        "modern corporate cafeteria with healthy food options and employees dining, natural light, lifestyle editorial",
        "smiling employee at ergonomic standing desk in open-plan office, corporate wellbeing concept, shallow depth of field",
    ],
    "Economic": [
        "stock market data on multiple computer monitors in a financial trading environment, dramatic blue lighting",
        "economist presenting charts and graphs to a corporate audience in an auditorium, editorial photography",
        "abstract visualization of economic data — rising bar charts overlaid on an aerial city photograph, double exposure",
        "professional reviewing The Wall Street Journal and financial printouts at a mahogany desk, classic editorial",
        "modern financial district skyline at blue hour, glass towers reflected in a still water feature, architectural",
    ],
    "Compliance": [
        "attorney and corporate executive reviewing legal documents at a conference table, formal setting, directional light",
        "close-up of a professional signing official corporate documents with a fountain pen, notary seal visible, dark background",
        "organized law books and corporate compliance binders on dark wood shelving, shallow depth of field, editorial",
        "two compliance officers in business attire reviewing a checklist on a tablet in a corporate corridor, documentary",
        "secure corporate data center with blurred server racks and a professional in background, blue technical lighting",
    ],
    "Diversity & Inclusion": [
        "diverse team of six professionals in a bright meeting room, various ages and backgrounds, genuine collaboration, natural light",
        "corporate mentoring session, senior executive and junior employee at a shared desk with laptops, warm editorial",
        "panel of four diverse speakers at a corporate conference, branded stage backdrop, event photography",
        "inclusive corporate workshop in a modern space, participants at round tables, colorful sticky notes on walls",
        "employees from different backgrounds celebrating a team success in an open office, candid corporate photography",
    ],
    "Technology": [
        "data center corridor with glowing server racks receding to infinity, blue and white light, wide angle, dramatic",
        "software developer at dual-monitor setup with code displayed, dark themed IDE, soft desk lamp light, editorial",
        "team of engineers collaborating around a large digital display showing architecture diagrams, corporate tech photo",
        "modern corporate IT operations center with multiple screens showing dashboards, technicians working, blue ambient light",
        "close-up of hands on a keyboard with code reflected in glasses, depth of field, technology editorial photography",
    ],
    "Leadership": [
        "executive giving keynote on a stage in a large corporate conference hall, dramatic spotlight, wide shot",
        "senior leader one-on-one coaching session, two professionals at a round table with coffee, warm window light",
        "confident executive walking through a modern office lobby, employees in background, corporate portrait photography",
        "CEO presenting annual results on a large screen to a full auditorium, corporate event photography",
        "leadership team portrait in executive conference room, five professionals, formal lighting, editorial photography",
    ],
    "Talent Acquisition": [
        "professional recruiter interviewing a candidate in a bright modern meeting room, across a glass table, editorial",
        "job fair at a corporate campus, booths and professionals networking, branded environment, event photography",
        "HR team reviewing resumes on laptops in a collaborative office space, notebooks and coffee, natural light",
        "two professionals shaking hands in a lobby after an interview, business formal, glass architecture background",
        "onboarding session for new employees in a training room, presenter at whiteboard, engaged audience, corporate",
    ],
    "Retention": [
        "long-tenure employee receiving recognition award in corporate ceremony, applause, event photography",
        "manager and employee in a private glass-walled office having a relaxed one-on-one, plants and natural light",
        "team of professionals celebrating a project milestone in a modern breakout space, candid corporate photography",
        "employee satisfaction survey on a tablet, held in open office environment, lifestyle corporate photography",
        "multi-generational team working collaboratively around a table, diverse ages, corporate documentary style",
    ],
}

REPORT_GENERIC_PROMPTS = [
    "corporate office interior with large windows, professionals working at desks, clean minimal aesthetic, natural light",
    "overhead view of polished conference table with documents, pens, glasses of water, professional setting",
    "business professional reviewing printed reports, pen in hand, warm desk lamp, blurred office background",
]

# ── Project prompts by industry ───────────────────────────────────────────────

PROJECT_INDUSTRY_PROMPTS = {
    "Healthcare": [
        "modern hospital exterior with glass facade and healing garden, blue sky, architectural photography",
        "medical professionals reviewing data on screens in a clean hospital operations center, blue ambient light",
        "state-of-the-art medical equipment in a bright clinical setting, no people, product photography aesthetic",
        "healthcare executives in meeting with doctors in hospital boardroom, corporate documentary",
        "expansive hospital atrium with natural light, wayfinding signage, patients and staff in background",
    ],
    "Finance": [
        "exterior of a glass financial district skyscraper at dusk, logo signage illuminated, architectural photography",
        "trading floor with multiple monitors showing financial data, professionals at workstations, dramatic blue lighting",
        "bank lobby with marble floors and high ceilings, teller counters in background, wide angle architecture",
        "financial advisor and client reviewing portfolio at a glass desk, warm window light, corporate portrait",
        "stacks of organized financial reports on mahogany desk, pen and calculator, editorial still life",
    ],
    "Energy": [
        "industrial energy facility at sunset, dramatic sky, editorial documentary photography",
        "wind farm on rolling hills, golden hour light, aerial perspective, environmental corporate",
        "engineer in hard hat and safety vest reviewing plans on a tablet at an industrial site",
        "interior of a modern energy control room, multiple screens with grid data, blue ambient light",
        "solar panel array on flat commercial rooftop, city skyline in background, corporate environmental",
    ],
    "Technology": [
        "modern tech campus building with glass and steel architecture, manicured grounds, architectural photography",
        "software engineering team in open-plan office with large monitors showing dashboards, natural light",
        "clean white server room with organized cable management, cold blue LED lighting, technical photography",
        "product team doing a sprint review with sticky notes on glass whiteboard, collaborative, editorial",
        "rooftop satellite dish and antenna array against a blue sky, corporate infrastructure photography",
    ],
    "Manufacturing": [
        "modern clean manufacturing floor with robotic assembly arms and industrial lighting, wide angle",
        "quality control inspector examining a product on a conveyor belt, hard hat and lab coat, documentary",
        "factory exterior shot at dusk, smokestacks with clean emissions, industrial architecture",
        "engineer reviewing blueprints on a drafting table in a manufacturing facility, overhead light",
        "aerial view of logistics distribution center surrounded by trucks, bird's eye view photography",
    ],
    "Retail": [
        "flagship retail store interior with clean shelving, warm lighting, editorial architectural photography",
        "corporate retail headquarters lobby with brand signage and modern furniture",
        "supply chain professionals in a distribution warehouse reviewing inventory on tablets",
        "retail executive meeting in a corporate setting with mood boards and product samples on table",
        "busy flagship store exterior on a city street, shoppers visible, brand signage, street photography",
    ],
    "Government": [
        "neoclassical government building exterior, American flag, blue sky, wide establishing shot",
        "government officials in formal meeting room with flags and wood paneling, documentary photography",
        "public service professionals at modern workstations reviewing data, open government office",
        "city hall exterior at dawn, columns and stone steps, architectural photography",
        "town hall meeting in a civic auditorium, presenter at podium, engaged audience, editorial",
    ],
    "Education": [
        "modern university research building with glass facade, students on plaza, architectural photography",
        "corporate training session in a bright classroom, presenter at front, engaged adult learners",
        "empty executive education classroom with stadium seating, projector screen, natural light",
        "academic library interior with high shelves and reading tables, warm light, editorial photography",
        "graduation ceremony on university campus quad, corporate sponsorship banners, event photography",
    ],
}

NEGATIVE_PROMPT = (
    "cartoon, illustration, anime, painting, text, watermark, logo overlay, "
    "low quality, blurry, out of focus, distorted, ugly, deformed, "
    "oversaturated, HDR, amateur photography, stock photo clichés, "
    "excessive lens flare, fake bokeh, composite obvious seams"
)


def build_report_prompt(category, slug, rng):
    prompts = REPORT_CATEGORY_PROMPTS.get(category, REPORT_GENERIC_PROMPTS)
    base = rng.choice(prompts)
    suffix = "professional corporate photography, 16:9 aspect ratio, muted navy and gold color palette"
    return f"{base}, {suffix}"


def build_project_prompt(industry, rng):
    prompts = PROJECT_INDUSTRY_PROMPTS.get(industry, PROJECT_INDUSTRY_PROMPTS["Technology"])
    base = rng.choice(prompts)
    suffix = "professional corporate photography, 16:9 aspect ratio, muted navy and gold color palette"
    return f"{base}, {suffix}"


def load_report_entries():
    """Import report catalog + categories without needing Django setup."""
    sys.path.insert(0, str(ACPWB_ROOT))
    # Minimal Django setup just to import the generators
    import django
    import os
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "config.settings.local")
    try:
        django.setup()
    except Exception:
        pass  # fine if it partially fails — we just need the data
    from apps.honeypot.report_generator import REPORT_CATALOG, REPORT_CATEGORIES, _rng_from_seed, _generate_synthetic
    import hashlib
    entries = list(REPORT_CATALOG)
    # Generate a representative set of synthetic reports
    for i in range(74):  # enough to cover all categories
        seed = f"synthetic_cover_item{i}"
        rng = _rng_from_seed(seed)
        entry = _generate_synthetic(rng, seed)
        entries.append(entry)
    return entries


def load_project_industries():
    from apps.projects.generators import INDUSTRIES
    return INDUSTRIES


def generate_report_covers(pipe, count, rng_seed):
    REPORT_COVERS_DIR.mkdir(parents=True, exist_ok=True)
    manifest_path = REPORT_COVERS_DIR / "manifest.json"
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}

    entries = load_report_entries()
    if count:
        entries = entries[:count]

    import torch
    generated = 0
    for entry in entries:
        slug = entry["slug"]
        out_path = REPORT_COVERS_DIR / f"{slug}.webp"
        if out_path.exists():
            print(f"  skip {slug} (exists)")
            continue

        rng = random.Random(f"report_{slug}_{rng_seed}")
        prompt = build_report_prompt(entry["category"], slug, rng)
        print(f"  [{generated + 1}] {slug[:55]}")
        print(f"       {prompt[:90]}...")

        seed_int = int.from_bytes(slug.encode()[:8].ljust(8, b"\x00"), "little") & 0x7FFFFFFF
        result = pipe(
            prompt=prompt,
            negative_prompt=NEGATIVE_PROMPT,
            num_inference_steps=30,
            guidance_scale=7.5,
            height=704,
            width=1280,
            generator=torch.Generator("mps").manual_seed(seed_int),
        )
        img = result.images[0]
        img.save(out_path, "WEBP", quality=85, method=6)
        manifest[slug] = {"prompt": prompt, "category": entry["category"], "path": f"img/report-covers/{slug}.webp"}
        manifest_path.write_text(json.dumps(manifest, indent=2))
        generated += 1

    print(f"\n  Report covers done: {generated} generated, {len(entries) - generated} skipped.")


def generate_project_covers(pipe, count, rng_seed):
    PROJECT_COVERS_DIR.mkdir(parents=True, exist_ok=True)
    manifest_path = PROJECT_COVERS_DIR / "manifest.json"
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}

    industries = load_project_industries()
    # Generate count images spread evenly across industries
    total = count or 80
    entries = []
    for i in range(total):
        industry = industries[i % len(industries)]
        entries.append((i, industry))

    import torch
    generated = 0
    for idx, industry in entries:
        out_path = PROJECT_COVERS_DIR / f"{idx:03d}.webp"
        if out_path.exists():
            print(f"  skip {idx:03d} ({industry}) (exists)")
            continue

        rng = random.Random(f"project_{idx}_{rng_seed}")
        prompt = build_project_prompt(industry, rng)
        print(f"  [{idx + 1}/{total}] {industry}: {prompt[:80]}...")

        result = pipe(
            prompt=prompt,
            negative_prompt=NEGATIVE_PROMPT,
            num_inference_steps=30,
            guidance_scale=7.5,
            height=704,
            width=1280,
            generator=torch.Generator("mps").manual_seed(idx),
        )
        img = result.images[0]
        img.save(out_path, "WEBP", quality=85, method=6)
        manifest[str(idx)] = {"prompt": prompt, "industry": industry, "path": f"img/project-covers/{idx:03d}.webp"}
        manifest_path.write_text(json.dumps(manifest, indent=2))
        generated += 1

    print(f"\n  Project covers done: {generated} generated, {total - generated} skipped.")


def main():
    parser = argparse.ArgumentParser(description="Generate corporate stock images for ACPWB reports and projects.")
    parser.add_argument("--type", choices=["reports", "projects", "both"], default="both",
                        help="Which images to generate (default: both)")
    parser.add_argument("--count", type=int, default=None,
                        help="Max images per type (default: all reports + 80 projects)")
    parser.add_argument("--seed", type=int, default=42,
                        help="Random seed for prompt variation (default: 42)")
    args = parser.parse_args()

    try:
        import torch
        from diffusers import StableDiffusionXLPipeline
    except ImportError:
        print("Missing dependencies. Run:")
        print("  pip install torch diffusers transformers accelerate pillow")
        sys.exit(1)

    if not torch.backends.mps.is_available():
        print("MPS not available — are you on Apple Silicon?")
        sys.exit(1)

    print(f"Loading {MODEL_ID} (~7GB download on first run) ...")
    pipe = StableDiffusionXLPipeline.from_pretrained(
        MODEL_ID,
        torch_dtype=torch.float32,
        use_safetensors=True,
    )
    pipe = pipe.to("mps")
    pipe.enable_attention_slicing()

    if args.type in ("reports", "both"):
        print("\nGenerating report covers ...")
        generate_report_covers(pipe, args.count, args.seed)

    if args.type in ("projects", "both"):
        print("\nGenerating project covers ...")
        generate_project_covers(pipe, args.count, args.seed)

    print("\nAll done.")


if __name__ == "__main__":
    main()
