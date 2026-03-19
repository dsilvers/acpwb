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


GIT_REPO_ROOT = Path(__file__).parent.parent
DJANGO_PROJECT_ROOT = GIT_REPO_ROOT / "acpwb"
REPORT_COVERS_DIR = DJANGO_PROJECT_ROOT / "static" / "img" / "report-covers"
PROJECT_COVERS_DIR = DJANGO_PROJECT_ROOT / "static" / "img" / "project-covers"

MODEL_ID = "stabilityai/stable-diffusion-3-medium-diffusers"

# ── Report prompts by category ────────────────────────────────────────────────

REPORT_CATEGORY_PROMPTS = {
    "Compensation": [
        "sleek conference room with laptops open to salary charts on the table, overhead view of desk with financial documents and graphs, corporate office, warm natural light",
        "close-up of financial data charts and bar graphs on a glass desk, office environment with city view, depth of field blur, corporate documentary photography",
        "executive boardroom with printed reports and charts laid out on a polished table, navy and gold color scheme, editorial photography",
        "professional documents and financial charts on a polished conference table, soft office lighting, corporate stock photography",
        "modern open-plan office with standing desks, large monitors showing compensation dashboards, clean corporate aesthetic, no people",
    ],
    "Workforce": [
        "bright modern office with a glass whiteboard covered in diagrams and sticky notes, representing a collaborative environment, corporate documentary style",
        "aerial view of busy corporate open office floor, rows of workstations, natural light through floor-to-ceiling windows, editorial photography",
        "glass-walled conference room with laptops and notebooks open on a circular table, suggesting a recent meeting, professional corporate atmosphere",
        "modern corporate lobby with marble floors and glass architecture, suggesting a business agreement, natural light, corporate editorial",
        "large monitor in a modern office displaying workforce analytics data, with empty chairs around it, shallow depth of field",
    ],
    "Governance": [
        "empty boardroom with long polished mahogany table and leather chairs, city skyline visible through windows, golden hour light",
        "close-up of a gavel on a polished wood desk with blurred legal documents in background, corporate governance concept",
        "a panel table in a formal setting with microphones and water glasses, prepared for an executive panel, directional studio lighting, editorial photograph",
        "stacks of organized corporate documents with a pen and reading glasses on a mahogany desk, warm office lighting",
        "glass and steel corporate headquarters exterior shot at dusk, reflections on windows, architectural photography",
    ],
    "ESG": [
        "modern glass office building with green living wall, solar panels on roof visible, blue sky, architectural photography",
        "a tablet displaying sustainability metrics on a desk, with a background of urban greenery, natural light",
        "aerial view of corporate campus with manicured grounds and trees, clean geometric architecture, drone photography",
        "wind turbines visible through the floor-to-ceiling window of a modern executive office, symbolic corporate environmental photo",
        "a newly planted sapling with gardening tools beside it on the grounds of a corporate campus, branded corporate social responsibility photography",
    ],
    "Benefits": [
        "bright modern employee wellness center inside a corporate campus, yoga mats and natural light, lifestyle photography",
        "a large screen in a well-lit conference room displaying an employee benefits presentation, empty chairs facing the screen",
        "employee benefit enrollment forms on a clean white desk with a pen, soft office lighting, documentary",
        "modern corporate cafeteria with healthy food options displayed, clean and bright, natural light, lifestyle editorial, no people",
        "ergonomic standing desk in a modern open-plan office, with a laptop and a plant, corporate wellbeing concept, shallow depth of field, no people",
    ],
    "Economic": [
        "stock market data on multiple computer monitors in a financial trading environment, dramatic blue lighting",
        "an auditorium with a large screen displaying economic charts and graphs, empty seats, editorial photography",
        "abstract visualization of economic data — rising bar charts overlaid on an aerial city photograph, double exposure",
        "The Wall Street Journal and financial printouts spread on a mahogany desk, classic editorial",
        "modern financial district skyline at blue hour, glass towers reflected in a still water feature, architectural",
    ],
    "Compliance": [
        "legal documents and binders spread across a conference table in a formal setting, directional light",
        "an official corporate document with a fountain pen resting on the signature line, notary seal visible, dark background",
        "organized law books and corporate compliance binders on dark wood shelving, shallow depth of field, editorial",
        "a tablet displaying a compliance checklist on a table in a corporate corridor, documentary style",
        "secure corporate data center with blurred server racks, blue technical lighting, emphasizing security and technology, no people",
    ],
    "Diversity & Inclusion": [
        "abstract representation of diversity and inclusion in a corporate setting, interconnected nodes of different colors on a digital display, bright meeting room background",
        "two coffee cups and open notebooks on a shared desk in a warm, modern office, suggesting a mentoring session, warm editorial",
        "an empty stage set for a corporate panel discussion, four chairs and microphones, branded stage backdrop, event photography",
        "a modern workshop space with round tables, colorful sticky notes with ideas on a large wall, representing collaboration",
        "abstract visualization of teamwork and success, colorful intersecting light trails in a modern office background, candid corporate photography",
    ],
    "Technology": [
        "data center corridor with glowing server racks receding to infinity, blue and white light, wide angle, dramatic",
        "dual-monitor setup with code displayed on screen, dark themed IDE, soft desk lamp light, editorial, no person",
        "a large digital display showing complex architecture diagrams in a modern tech office, empty chairs around, corporate tech photo",
        "modern corporate IT operations center with multiple screens showing dashboards, empty workstations, blue ambient light",
        "close-up of a modern ergonomic keyboard and mouse on a desk, with code reflected in a nearby screen, depth of field, technology editorial photography",
    ],
    "Leadership": [
        "an empty stage in a large corporate conference hall, prepared for a keynote speech, dramatic spotlight on a lectern, wide shot",
        "a round table in a quiet office corner with two coffee cups and notebooks, suggesting a one-on-one coaching session, warm window light",
        "a modern, expansive office lobby with striking architecture, suggesting corporate power and confidence, corporate portrait photography style but with no people",
        "a large screen in an empty auditorium displaying annual results, corporate event photography",
        "an executive conference room with a long table and empty chairs, prepared for a board meeting, formal lighting, editorial photography",
    ],
    "Talent Acquisition": [
        "a bright modern meeting room with two empty chairs facing each other across a glass table, suggesting an interview, editorial",
        "an empty corporate job fair booth with branding and informational pamphlets, event photography",
        "laptops open to resume templates in a collaborative office space, notebooks and coffee on the table, natural light",
        "the entrance to a modern glass office building, suggesting a successful meeting or new beginning, business formal architecture",
        "a corporate training room with a whiteboard covered in notes, empty chairs facing it, suggesting a recent onboarding session",
    ],
    "Retention": [
        "a corporate recognition award trophy on a stage podium, empty ceremony hall in background, event photography",
        "a private glass-walled office with two empty chairs and a plant, suggesting a relaxed one-on-one meeting, natural light",
        "a modern corporate breakout space with remnants of a celebration, like confetti and empty glasses, candid corporate photography style",
        "a tablet displaying an employee satisfaction survey on a desk in an open office environment, lifestyle corporate photography",
        "abstract visualization of collaboration between different generations, overlapping transparent circles of different sizes and colors, corporate documentary style",
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
        "a clean hospital operations center with multiple screens displaying medical data, blue ambient light, no people",
        "state-of-the-art medical equipment in a bright clinical setting, no people, product photography aesthetic",
        "a hospital boardroom with a large table, prepared for a meeting between executives and doctors, corporate documentary",
        "expansive hospital atrium with natural light, wayfinding signage, architectural focus, no people or blurred people in distant background",
    ],
    "Finance": [
        "exterior of a glass financial district skyscraper at dusk, logo signage illuminated, architectural photography",
        "trading floor with multiple monitors showing financial data, empty workstations, dramatic blue lighting",
        "bank lobby with marble floors and high ceilings, teller counters in background, wide angle architecture",
        "a glass desk with a tablet showing a financial portfolio, two empty chairs, warm window light, corporate portrait style but no people",
        "stacks of organized financial reports on mahogany desk, pen and calculator, editorial still life",
    ],
    "Energy": [
        "industrial energy facility at sunset, dramatic sky, editorial documentary photography",
        "wind farm on rolling hills, golden hour light, aerial perspective, environmental corporate",
        "a tablet displaying industrial plans on a table at an industrial site, hard hat resting next to it",
        "interior of a modern energy control room, multiple screens with grid data, blue ambient light",
        "solar panel array on flat commercial rooftop, city skyline in background, corporate environmental",
    ],
    "Technology": [
        "modern tech campus building with glass and steel architecture, manicured grounds, architectural photography",
        "an open-plan office with large monitors showing dashboards and code, empty desks, natural light",
        "clean white server room with organized cable management, cold blue LED lighting, technical photography",
        "a glass whiteboard covered in colorful sticky notes from a sprint review, collaborative but empty room, editorial",
        "rooftop satellite dish and antenna array against a blue sky, corporate infrastructure photography",
    ],
    "Manufacturing": [
        "modern clean manufacturing floor with robotic assembly arms and industrial lighting, wide angle",
        "a product on a conveyor belt under a bright inspection light, robotic arm nearby, documentary style, no people",
        "factory exterior shot at dusk, smokestacks with clean emissions, industrial architecture",
        "blueprints spread out on a drafting table in a manufacturing facility, overhead light, no person",
        "aerial view of logistics distribution center surrounded by trucks, bird's eye view photography",
    ],
    "Retail": [
        "flagship retail store interior with clean shelving, warm lighting, editorial architectural photography",
        "corporate retail headquarters lobby with brand signage and modern furniture",
        "a distribution warehouse with shelves of inventory, a tablet on a crate showing stock levels",
        "a corporate meeting room with mood boards and product samples on the table, prepared for a retail executive meeting",
        "flagship store exterior on a city street, architectural focus, long exposure to blur any people into streaks of motion",
    ],
    "Government": [
        "neoclassical government building exterior, American flag, blue sky, wide establishing shot",
        "a formal government meeting room with flags and wood paneling, empty chairs around a large table, documentary photography",
        "a modern open government office with empty workstations, computers displaying public data, natural light",
        "city hall exterior at dawn, columns and stone steps, architectural photography",
        "an empty civic auditorium, stage with a podium, prepared for a town hall meeting, editorial",
    ],
    "Education": [
        "modern university research building with glass facade, architectural focus, empty plaza, architectural photography",
        "a bright corporate classroom with a presenter's screen and empty chairs for adult learners",
        "empty executive education classroom with stadium seating, projector screen, natural light",
        "academic library interior with high shelves and reading tables, warm light, editorial photography",
        "graduation ceremony on university campus quad, corporate sponsorship banners, event photography",
    ],
}

NEGATIVE_PROMPT = (
    "people, person, man, woman, hands, fingers, malformed hands, "
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


def load_data_from_json():
    """Loads data from the JSON file exported by the management command."""
    json_path = GIT_REPO_ROOT / "acpwb/gen_data.json"
    print(json_path)
    if not json_path.exists():
        print(f"Error: Data file not found at {json_path}")
        print("\nPlease run the management command inside the container first:")
        print("  docker compose exec web python manage.py export_gen_data\n")
        sys.exit(1)

    print(f"Loading data from {json_path}...")
    data = json.loads(json_path.read_text())
    # The management command generates the full list of reports and project industries.
    return data["reports"], data["projects"]["industries"]


def generate_report_covers(pipe, count, rng_seed, entries):
    REPORT_COVERS_DIR.mkdir(parents=True, exist_ok=True)
    manifest_path = REPORT_COVERS_DIR / "manifest.json"
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}
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


def generate_project_covers(pipe, count, rng_seed, industries):
    PROJECT_COVERS_DIR.mkdir(parents=True, exist_ok=True)
    manifest_path = PROJECT_COVERS_DIR / "manifest.json"
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}
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
        from diffusers import StableDiffusion3Pipeline
    except ImportError:
        print("Missing dependencies. Run:")
        print("  pip install torch diffusers transformers accelerate pillow")
        sys.exit(1)

    if not torch.backends.mps.is_available():
        print("MPS not available — are you on Apple Silicon?")
        sys.exit(1)

    print(f"Loading {MODEL_ID} (~7GB download on first run) ...")
    pipe = StableDiffusion3Pipeline.from_pretrained(
        MODEL_ID,
        torch_dtype=torch.float16,
        use_safetensors=True,
    )
    pipe = pipe.to("mps")
    pipe.enable_attention_slicing()

    report_entries, project_industries = load_data_from_json()

    if args.type in ("reports", "both"):
        print("\nGenerating report covers ...")
        generate_report_covers(pipe, args.count, args.seed, report_entries)

    if args.type in ("projects", "both"):
        print("\nGenerating project covers ...")
        generate_project_covers(pipe, args.count, args.seed, project_industries)

    print("\nAll done.")


if __name__ == "__main__":
    main()
