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
        "printed salary survey binders stacked neatly on a credenza in an executive office, morning light through venetian blinds, editorial still life",
        "wide angle of an empty glass-walled boardroom with a projection screen showing a pay-equity bar chart, city skyline backdrop",
        "close-up of a mechanical pencil resting on a printed compensation matrix spreadsheet, warm desk lamp, shallow depth of field",
        "dual monitors on a modern desk showing side-by-side salary benchmark dashboards, tidy home office, natural window light",
        "aerial view of an organized desk: printed reports, coffee cup, reading glasses, ruler, all perfectly arranged, overhead editorial",
        "a sleek tablet propped against a stack of binders in a glass-walled office, compensation data visible on screen, corporate lifestyle",
        "financial district at sunrise through floor-to-ceiling office windows, reflecting the golden light across a polished conference table with documents",
    ],
    "Workforce": [
        "bright modern office with a glass whiteboard covered in diagrams and sticky notes, representing a collaborative environment, corporate documentary style",
        "aerial view of busy corporate open office floor, rows of workstations, natural light through floor-to-ceiling windows, editorial photography",
        "glass-walled conference room with laptops and notebooks open on a circular table, suggesting a recent meeting, professional corporate atmosphere",
        "modern corporate lobby with marble floors and glass architecture, suggesting a business agreement, natural light, corporate editorial",
        "large monitor in a modern office displaying workforce analytics data, with empty chairs around it, shallow depth of field",
        "long rows of ergonomic workstations in an open-plan corporate floor, overhead industrial lighting, symmetry, no people",
        "a corporate employee bulletin board covered in charts, announcements, and analytics printouts, warm hallway light, editorial documentary",
        "a large digital display on an office wall showing an org chart and headcount metrics, empty conference chairs in foreground",
        "close-up of a printed workforce planning document with highlighted sections and margin notes, dark mahogany desk, warm lamp",
        "modern HR operations office: standing desks, multiple monitors, a half-filled coffee mug, late afternoon golden light",
        "corporate campus outdoor seating area with empty benches and architectural landscaping, suggesting a workforce at work inside, atmospheric",
        "elevator bank of a glass corporate tower at midday, polished steel doors, marble floor, implying scale and workforce size",
    ],
    "Governance": [
        "empty boardroom with long polished mahogany table and leather chairs, city skyline visible through windows, golden hour light",
        "close-up of a gavel on a polished wood desk with blurred legal documents in background, corporate governance concept",
        "a panel table in a formal setting with microphones and water glasses, prepared for an executive panel, directional studio lighting, editorial photograph",
        "stacks of organized corporate documents with a pen and reading glasses on a mahogany desk, warm office lighting",
        "glass and steel corporate headquarters exterior shot at dusk, reflections on windows, architectural photography",
        "a nameplate and water glass at a formal board meeting table, leather portfolio open in front, blurred room in background, editorial",
        "circular corporate governance binders arranged by year on dark wood shelving, a single shaft of light from a narrow window",
        "an executive suite hallway with framed annual report covers lining both walls, receding perspective, corporate heritage photography",
        "wide shot of an empty shareholders meeting hall, rows of seats, a dais with a podium and microphone, dramatic overhead lighting",
        "a brass plaque mounted on a marble wall reading 'Board of Directors', institutional photography, shallow depth of field",
        "proxy statement and corporate charter documents fanned out on a mahogany conference table, gold pen beside them, formal editorial",
        "aerial view of a symmetrical corporate campus from directly above, showing the scale and precision of institutional architecture",
    ],
    "ESG": [
        "modern glass office building with green living wall, solar panels on roof visible, blue sky, architectural photography",
        "a tablet displaying sustainability metrics on a desk, with a background of urban greenery, natural light",
        "aerial view of corporate campus with manicured grounds and trees, clean geometric architecture, drone photography",
        "wind turbines visible through the floor-to-ceiling window of a modern executive office, symbolic corporate environmental photo",
        "a newly planted sapling with gardening tools beside it on the grounds of a corporate campus, branded corporate social responsibility photography",
        "rooftop garden of a corporate headquarters, raised plant beds with city skyline in background, environmental CSR editorial",
        "solar panels covering the roof of a distribution center, aerial drone shot, clean geometric lines, blue sky",
        "a bicycle-share station in front of a glass corporate headquarters, sustainability initiative, architectural lifestyle",
        "rain water collection system and native plant landscaping at a corporate campus entrance, environmental branding photography",
        "a printed ESG annual report open to a full-bleed infographic page on a clean white desk, natural light, editorial still life",
        "corporate sustainability dashboard on a large wall-mounted display in an empty lobby, showing carbon metrics and green targets",
        "interior of a LEED-certified office building: exposed timber, natural ventilation grilles, plants, daylight harvesting design",
    ],
    "Benefits": [
        "bright modern employee wellness center inside a corporate campus, yoga mats and natural light, lifestyle photography",
        "a large screen in a well-lit conference room displaying an employee benefits presentation, empty chairs facing the screen",
        "employee benefit enrollment forms on a clean white desk with a pen, soft office lighting, documentary",
        "modern corporate cafeteria with healthy food options displayed, clean and bright, natural light, lifestyle editorial, no people",
        "ergonomic standing desk in a modern open-plan office, with a laptop and a plant, corporate wellbeing concept, shallow depth of field, no people",
        "a corporate fitness center with treadmills and free weights visible through a glass wall, no people, lifestyle corporate photography",
        "a tray of fresh fruit on a communal kitchen counter in a modern office, morning light, healthy lifestyle editorial",
        "wide shot of a bright corporate meditation or quiet room with soft cushions and diffused lighting, no people",
        "a benefits booklet and summary plan description fanned out on a clean white surface, clinical editorial still life",
        "employee lounge with comfortable seating, a foosball table, and natural light, suggesting modern corporate culture, empty",
        "a dental and health insurance card resting on a marble surface, macro still life, shallow depth of field",
        "corporate on-site childcare center exterior with colorful mural, playground equipment, blue sky, lifestyle branding photography",
    ],
    "Economic": [
        "stock market data on multiple computer monitors in a financial trading environment, dramatic blue lighting",
        "an auditorium with a large screen displaying economic charts and graphs, empty seats, editorial photography",
        "abstract visualization of economic data — rising bar charts overlaid on an aerial city photograph, double exposure",
        "The Wall Street Journal and financial printouts spread on a mahogany desk, classic editorial",
        "modern financial district skyline at blue hour, glass towers reflected in a still water feature, architectural",
        "a row of Bloomberg terminals on a trading desk, screens glowing with data, blue and green ambient light, no people",
        "aerial view of a container port at dusk, cranes and stacked shipping containers, symbolizing global economic scale",
        "a printed quarterly earnings report open on a conference table, reading glasses resting on the page, editorial still life",
        "abstract long exposure of a busy financial district intersection at night, light trails from traffic, economic energy",
        "a glass building lobby with an LED ticker showing market data scrolling across the screen, architectural corporate",
        "GDP growth chart printed large on the wall of a think tank conference room, empty Eames chairs, editorial",
        "overhead aerial of a dense urban commercial district, implying the scale and interconnection of local economies",
    ],
    "Compliance": [
        "legal documents and binders spread across a conference table in a formal setting, directional light",
        "an official corporate document with a fountain pen resting on the signature line, notary seal visible, dark background",
        "organized law books and corporate compliance binders on dark wood shelving, shallow depth of field, editorial",
        "a tablet displaying a compliance checklist on a table in a corporate corridor, documentary style",
        "secure corporate data center with blurred server racks, blue technical lighting, emphasizing security and technology, no people",
        "a red 'Confidential' stamp and ink pad on a desk beside a sealed manila envelope, editorial still life, dark background",
        "a formal legal library inside a corporate headquarters, rows of bound regulatory codes, warm leather-and-wood aesthetic",
        "close-up of a fingerprint scanner and keycard access panel on a secure corporate door, technology security concept",
        "a signed compliance acknowledgment form under a desk lamp, pen still resting on it, institutional editorial photography",
        "a printed GDPR or regulatory compliance framework diagram spread across a conference table, corporate policy meeting atmosphere",
        "server room with locked cages and access logs posted on the door, security audit concept, fluorescent lighting",
        "a corporate whistleblower hotline poster mounted in a clean hallway, institutional branding, editorial documentary",
    ],
    "Diversity & Inclusion": [
        "abstract representation of diversity and inclusion in a corporate setting, interconnected nodes of different colors on a digital display, bright meeting room background",
        "two coffee cups and open notebooks on a shared desk in a warm, modern office, suggesting a mentoring session, warm editorial",
        "an empty stage set for a corporate panel discussion, four chairs and microphones, branded stage backdrop, event photography",
        "a modern workshop space with round tables, colorful sticky notes with ideas on a large wall, representing collaboration",
        "abstract visualization of teamwork and success, colorful intersecting light trails in a modern office background, candid corporate photography",
        "a company pride flag and mission statement banner hanging side by side in a corporate lobby, institutional photography",
        "an empty inclusive restroom sign beside a well-lit corporate hallway, modern DE&I facility design",
        "a multilingual welcome sign in a glass corporate lobby, inclusive branding, warm architectural interior light",
        "a breakout room whiteboard covered in equity framework diagrams from a recent DEI workshop, empty chairs, editorial",
        "abstract mosaic wall art inside a corporate HQ lobby, many small tiles forming a unified design, symbolizing inclusion",
        "a corporate ERG (employee resource group) bulletin board with colorful flyers and events calendar, hallway editorial",
        "close-up of a printed pay equity audit report on a glass conference table, gold pen, soft office light",
    ],
    "Technology": [
        "data center corridor with glowing server racks receding to infinity, blue and white light, wide angle, dramatic",
        "dual-monitor setup with code displayed on screen, dark themed IDE, soft desk lamp light, editorial, no person",
        "a large digital display showing complex architecture diagrams in a modern tech office, empty chairs around, corporate tech photo",
        "modern corporate IT operations center with multiple screens showing dashboards, empty workstations, blue ambient light",
        "close-up of a modern ergonomic keyboard and mouse on a desk, with code reflected in a nearby screen, depth of field, technology editorial photography",
        "ethernet patch panel in a corporate network closet, colorful cables organized with velcro ties, technical documentary",
        "a MacBook Pro and external monitor showing a cloud infrastructure diagram on a minimal white desk, clean tech editorial",
        "server cooling units and power distribution rows in a modern hyperscale data center, industrial scale, no people",
        "a Kanban board on a glass wall covered in user story cards from a recent product sprint, tech office editorial",
        "a smartboard in a tech company war room covered in system architecture drawings, dramatic overhead lighting",
        "a corporate satellite uplink dish farm on a rooftop at dawn, clean geometric shapes against a gradient sky",
        "a printed IT roadmap document on a standing desk beside an ergonomic chair, modern tech workspace editorial",
    ],
    "Leadership": [
        "an empty stage in a large corporate conference hall, prepared for a keynote speech, dramatic spotlight on a lectern, wide shot",
        "a round table in a quiet office corner with two coffee cups and notebooks, suggesting a one-on-one coaching session, warm window light",
        "a modern, expansive office lobby with striking architecture, suggesting corporate power and confidence, corporate portrait photography style but with no people",
        "a large screen in an empty auditorium displaying annual results, corporate event photography",
        "an executive conference room with a long table and empty chairs, prepared for a board meeting, formal lighting, editorial photography",
        "a corner executive office with a view of the city, a clear desk, and a single leather chair, power and solitude, architectural editorial",
        "an empty CEO suite antechamber with a receptionist desk, live orchid, and framed accolades on the wall, corporate prestige",
        "a leadership development retreat venue: a mountain lodge conference room with floor-to-ceiling windows and empty chairs, aspirational",
        "a printed succession planning document and organizational chart on a mahogany desk, gold letter opener, formal editorial",
        "a trophy case in a corporate hallway filled with awards, plaques, and recognitions, institutional pride photography",
        "a mentor's notebook open to a page of hand-written leadership principles, pen resting on it, warm personal editorial",
        "wide shot of a dimly lit auditorium before an executive all-hands event, stage lit with a single spotlight, anticipation",
    ],
    "Talent Acquisition": [
        "a bright modern meeting room with two empty chairs facing each other across a glass table, suggesting an interview, editorial",
        "an empty corporate job fair booth with branding and informational pamphlets, event photography",
        "laptops open to resume templates in a collaborative office space, notebooks and coffee on the table, natural light",
        "the entrance to a modern glass office building, suggesting a successful meeting or new beginning, business formal architecture",
        "a corporate training room with a whiteboard covered in notes, empty chairs facing it, suggesting a recent onboarding session",
        "a printed offer letter on a clean white desk with a pen beside it, an implied moment of decision, editorial still life",
        "a glass bowl of business cards on a recruiter's desk at a job fair, out-of-focus event hall in background",
        "a corporate career portal on a large office monitor, clean UI showing open positions, modern HR tech visual",
        "a freshly set up new-hire welcome kit on a desk: branded notebook, lanyard, access badge, morning light",
        "a university campus career fair setup: empty booths, a wide-angle shot before the event begins, anticipation",
        "a whiteboard in a recruiting office with candidate pipeline stages written in marker, editorial documentary",
        "a printed stack of resumes with a red pen resting on top, warm desk lamp, shallow depth of field editorial",
    ],
    "Retention": [
        "a corporate recognition award trophy on a stage podium, empty ceremony hall in background, event photography",
        "a private glass-walled office with two empty chairs and a plant, suggesting a relaxed one-on-one meeting, natural light",
        "a modern corporate breakout space with remnants of a celebration, like confetti and empty glasses, candid corporate photography style",
        "a tablet displaying an employee satisfaction survey on a desk in an open office environment, lifestyle corporate photography",
        "abstract visualization of collaboration between different generations, overlapping transparent circles of different sizes and colors, corporate documentary style",
        "a printed years-of-service award certificate in a frame on a credenza, behind it a blurred hallway, corporate milestone",
        "a 'Years of Service' wall in a corporate headquarters hallway, brass plaques with names and years, institutional editorial",
        "a manager's desk with a handwritten thank-you note and a small gift, suggesting employee recognition, warm editorial still life",
        "a company all-hands celebration event space, empty before the party: tables set, banners hung, balloons not yet released",
        "a modern employee lounge with a 'Employee of the Month' display board, spotlit, rest of room in shadow, corporate culture",
        "a printed engagement survey results report on a conference table, sticky note flags marking key sections, editorial documentary",
        "an exit interview room: two chairs, a table, a box of tissues, and a window, sparse and quietly corporate, editorial",
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
        "a modern medical research lab with stainless steel benches and equipment, clean white lighting, no people, editorial",
        "hospital pharmacy interior with organized medication dispensaries and clean clinical shelving, institutional photography",
        "a telehealth command center with rows of monitors, headsets at each workstation, empty, blue ambient light",
        "an outdoor healing garden on a hospital campus, stone path, native plants, architectural landscape photography",
        "a printed patient outcome metrics report on a hospital administrator's desk, stethoscope resting beside it, editorial still life",
        "a clean sterile operating suite with equipment ready, overhead surgical lights on, no people, technical documentary",
        "a large hospital campus aerial view at dusk, parking structure and main entrance lit, institutional architectural photography",
    ],
    "Finance": [
        "exterior of a glass financial district skyscraper at dusk, logo signage illuminated, architectural photography",
        "trading floor with multiple monitors showing financial data, empty workstations, dramatic blue lighting",
        "bank lobby with marble floors and high ceilings, teller counters in background, wide angle architecture",
        "a glass desk with a tablet showing a financial portfolio, two empty chairs, warm window light, corporate portrait style but no people",
        "stacks of organized financial reports on mahogany desk, pen and calculator, editorial still life",
        "a polished brass revolving door at a bank entrance, institutional architecture, symmetrical, golden hour",
        "a modern private wealth office with leather seating and abstract art, suggesting high-net-worth clientele, editorial interior",
        "a printed annual report open to a financial highlights page on a glass conference table, corporate editorial photography",
        "a bank vault door slightly ajar, heavy steel, dramatic lighting from within, symbolic financial security",
        "an empty stock exchange trading floor at closing bell, screens still glowing, wide establishing shot",
        "aerial view of a financial district at night, lights forming a grid of commerce, drone photography",
        "a financial analyst's multiple-monitor workstation with risk dashboards, empty ergonomic chair, late evening office light",
    ],
    "Energy": [
        "industrial energy facility at sunset, dramatic sky, editorial documentary photography",
        "wind farm on rolling hills, golden hour light, aerial perspective, environmental corporate",
        "a tablet displaying industrial plans on a table at an industrial site, hard hat resting next to it",
        "interior of a modern energy control room, multiple screens with grid data, blue ambient light",
        "solar panel array on flat commercial rooftop, city skyline in background, corporate environmental",
        "an offshore wind platform from a distance, North Sea atmosphere, industrial scale, dramatic clouds",
        "a natural gas processing facility at dusk, flare stacks and industrial piping, editorial documentary",
        "a utility-scale battery storage installation, rows of white containers in a desert landscape, aerial view",
        "a smart grid operations center, wall of screens showing regional energy distribution maps, no people, blue light",
        "a drilling platform supply vessel at a dock, industrial maritime, golden morning light",
        "close-up of industrial-grade electrical switchgear and transformers at a substation, technical documentary",
        "a printed energy sector project feasibility report on a field desk beside a hard hat, industrial editorial still life",
    ],
    "Technology": [
        "modern tech campus building with glass and steel architecture, manicured grounds, architectural photography",
        "an open-plan office with large monitors showing dashboards and code, empty desks, natural light",
        "clean white server room with organized cable management, cold blue LED lighting, technical photography",
        "a glass whiteboard covered in colorful sticky notes from a sprint review, collaborative but empty room, editorial",
        "rooftop satellite dish and antenna array against a blue sky, corporate infrastructure photography",
        "a hyperscale data center exterior at night, backlit cooling units and industrial scale, atmospheric editorial",
        "a fiber optic cable cross-section under macro lens, light traveling through glass strands, abstract technology",
        "a modern product design studio with hardware prototypes on white tables, clean overhead lighting, no people",
        "a corporate security operations center with a curved wall of screens showing global threat maps, no people, red and blue light",
        "a 3D printer completing a plastic prototype part in a modern innovation lab, close-up, technical editorial",
        "a clean anechoic testing chamber inside a tech facility, foam-lined walls, a single piece of test equipment, technical documentary",
        "an empty chip fabrication clean room corridor, white suits on hooks by the door, implying precision and scale",
    ],
    "Manufacturing": [
        "modern clean manufacturing floor with robotic assembly arms and industrial lighting, wide angle",
        "a product on a conveyor belt under a bright inspection light, robotic arm nearby, documentary style, no people",
        "factory exterior shot at dusk, smokestacks with clean emissions, industrial architecture",
        "blueprints spread out on a drafting table in a manufacturing facility, overhead light, no person",
        "aerial view of logistics distribution center surrounded by trucks, bird's eye view photography",
        "a precision CNC machining cell with metal shavings and cutting tools, clean industrial editorial photography",
        "a quality control inspection table in a factory, printed measurement reports and calipers, no people",
        "an automated guided vehicle traveling through a wide factory aisle, no people, overhead industrial lighting",
        "a printed production floor layout and capacity plan pinned to a corkboard, manufacturing planning editorial",
        "a welding jig in a fabrication shop with recently completed parts cooling, industrial still life, no people",
        "a supply chain visualization projected on a wall screen in a manufacturing war room, empty chairs, editorial",
        "aerial drone view of a manufacturing campus with multiple buildings, loading docks, and a rail spur, dawn light",
    ],
    "Retail": [
        "flagship retail store interior with clean shelving, warm lighting, editorial architectural photography",
        "corporate retail headquarters lobby with brand signage and modern furniture",
        "a distribution warehouse with shelves of inventory, a tablet on a crate showing stock levels",
        "a corporate meeting room with mood boards and product samples on the table, prepared for a retail executive meeting",
        "flagship store exterior on a city street, architectural focus, long exposure to blur any people into streaks of motion",
        "a retail analytics dashboard on a large screen in a quiet back office, empty chairs, corporate tech editorial",
        "a merchandising planogram printed and taped to a retail stockroom wall, organizational documentary",
        "a pristine retail store opening setup: fixtures installed, lighting on, mannequins dressed, no people yet",
        "an e-commerce fulfillment center with rows of shelving and conveyor belts, overhead view, industrial lifestyle",
        "a printed quarterly sales performance report with a red pen and sticky notes on a retail manager's desk",
        "a store window display in its design phase, garment stands and props half-assembled, editorial behind the scenes",
        "aerial view of a major shopping destination at peak season, traffic and parking visible, commercial real estate editorial",
    ],
    "Government": [
        "neoclassical government building exterior, American flag, blue sky, wide establishing shot",
        "a formal government meeting room with flags and wood paneling, empty chairs around a large table, documentary photography",
        "a modern open government office with empty workstations, computers displaying public data, natural light",
        "city hall exterior at dawn, columns and stone steps, architectural photography",
        "an empty civic auditorium, stage with a podium, prepared for a town hall meeting, editorial",
        "a municipal emergency operations center with large screens and workstations, no people, official documentary",
        "a printed government RFP document on a clean desk with a seal stamp and official envelope, institutional editorial",
        "a transportation authority control room, screens showing transit maps and schedules, blue ambient light, no people",
        "an empty public hearing room with a dais, name placards, and microphones, government editorial photography",
        "a federal courthouse entrance with columned portico and mounted eagle seal, symmetrical, morning light",
        "a government printing office with organized document trays and an institutional printing press, documentary editorial",
        "wide angle of a state capitol building interior rotunda, looking up at the dome, architectural photography",
    ],
    "Education": [
        "modern university research building with glass facade, architectural focus, empty plaza, architectural photography",
        "a bright corporate classroom with a presenter's screen and empty chairs for adult learners",
        "empty executive education classroom with stadium seating, projector screen, natural light",
        "academic library interior with high shelves and reading tables, warm light, editorial photography",
        "graduation ceremony on university campus quad, corporate sponsorship banners, event photography",
        "an empty lecture hall at a business school, tiered seating, a case-method discussion seating arrangement, editorial",
        "a university research lab with organized equipment and printed lab manuals on clean benches, academic editorial",
        "a learning management system dashboard displayed on a large screen in an empty training room, corporate EdTech concept",
        "a corporate training center hallway with numbered classroom doors and motivational signage, institutional photography",
        "a printed curriculum design document and course syllabus on a faculty desk, warm lamp, academic editorial still life",
        "aerial view of a university campus in autumn, colorful foliage around brick buildings, editorial overhead photography",
        "a corporate tuition reimbursement program poster on an office bulletin board, lifestyle institutional photography",
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
