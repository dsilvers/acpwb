#!/usr/bin/env python3
"""
Standalone background image generator for ACPWB presentation slides — no Django required.

Generates corporate stock-photography-style backgrounds using SDXL-Turbo on Apple MPS.
No people in any image. Images are saved as WebP to static/img/presentations/backgrounds/.

Install:
    pip install diffusers transformers accelerate pillow torch

Usage:
    python generate_presentation_images.py                    # 60 backgrounds
    python generate_presentation_images.py --count 60
    python generate_presentation_images.py --force            # regenerate existing
    python generate_presentation_images.py --dry-run          # preview prompts only
    python generate_presentation_images.py --steps 2          # inference steps (default 2)
    python generate_presentation_images.py --model stabilityai/sdxl-turbo

Notes:
    - Requires Apple Silicon Mac for MPS acceleration; falls back to CPU
    - Images are 800×450 px, saved as WebP quality 85
    - After generation run: docker compose exec web python manage.py collectstatic --noinput
"""

import argparse
import hashlib
import random
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT_DIR = HERE / "acpwb" / "static" / "img" / "presentations" / "backgrounds"

# ---------------------------------------------------------------------------
# Prompt library
# ---------------------------------------------------------------------------

_CATEGORIES = {
    "boardroom": [
        "empty modern boardroom, floor-to-ceiling windows, city skyline view, no people, "
        "professional architectural photography, clean minimal",
        "empty executive conference room, long dark wood table, leather chairs, panoramic windows, "
        "no people, corporate interior photography",
        "minimalist boardroom, frosted glass walls, marble floor, no people, "
        "clean lines, architectural photography",
        "circular boardroom with wraparound windows, high-back chairs around oval table, "
        "no people, architectural photography",
        "executive boardroom with dark paneled walls, oil paintings, brass fixtures, "
        "no people, traditional corporate interior",
        "glass-walled boardroom overlooking open office floor below, no people, "
        "modern corporate architecture",
        "minimalist white boardroom, single long table, tulip chairs, "
        "no people, Scandinavian corporate design",
        "boardroom with green marble table, gold trim, no people, "
        "luxury corporate interior photography",
        "boardroom with low credenza, abstract sculpture, indirect cove lighting, "
        "no people, contemporary corporate interior",
        "narrow boardroom with floor-to-ceiling shelving one wall, glass opposite wall, "
        "no people, modern executive interior",
        "octagonal boardroom, coffered ceiling, herringbone wood floor, "
        "no people, traditional architecture photography",
        "boardroom with acoustic ceiling panels, integrated AV screen, matte surfaces, "
        "no people, modern office interior",
        "corner boardroom, two walls of glass meeting at angle, city view both directions, "
        "no people, high-rise interior photography",
        "boardroom with deep teal walls, brass pendant lights, velvet chairs, "
        "no people, luxury interior photography",
        "biophilic boardroom, living moss wall, wood table, floor plants, skylights, "
        "no people, contemporary sustainable design",
        "boardroom with frosted glass privacy panels, integrated power columns, "
        "no people, corporate interior minimalism",
        "historic boardroom, wainscoting, portrait paintings, brass chandelier, "
        "no people, heritage corporate interior",
        "loft-style boardroom, exposed concrete ceiling, steel windows, industrial aesthetic, "
        "no people, editorial interior photography",
        "boardroom with retractable glass partition, dual-purpose event space, "
        "no people, flexible modern workplace",
        "boardroom with wraparound whiteboard walls, rolling stools, no people, "
        "creative corporate interior",
    ],
    "office": [
        "open plan office with empty desks, modern ergonomic furniture, plants, "
        "natural light, no people, professional interior photography",
        "corporate lobby interior, polished floors, reception area, no people, "
        "architectural photography",
        "empty co-working space, exposed brick, pendant lights, standing desks, "
        "plants, no people, editorial photography",
        "modern office kitchen and lounge area, no people, clean and contemporary, "
        "natural light",
        "empty corner office, floor-to-ceiling windows, single desk, city view, "
        "no people, executive interior photography",
        "open trading floor with rows of monitors and empty chairs, no people, "
        "corporate finance interior",
        "creative agency office, bright colors, whiteboards covered in diagrams, "
        "no people, startup interior photography",
        "law firm office, rows of bookshelves, dark wood furniture, leather chairs, "
        "no people, professional interior",
        "corporate wellness room, yoga mats rolled out, plants, diffused light, "
        "no people, modern office amenity",
        "empty office hallway, glass-walled meeting rooms on both sides, "
        "no people, architectural photography, perspective shot",
        "rooftop terrace of office building, outdoor seating, city view, "
        "no people, corporate amenity photography",
        "phone booth pods in open office, acoustic panels, soft lighting, "
        "no people, modern workplace interior",
        "empty open-plan floor at dusk, city lights visible through windows, "
        "ambient desk lamps on, no people, moody office photography",
        "minimalist private office, single floating desk, one framed print, "
        "no people, executive interior, zen aesthetic",
        "office atrium with spiral staircase, glass balustrades, skylight above, "
        "no people, architectural interior photography",
        "newsroom-style office, long rows of monitors, track lighting, "
        "no people, corporate media interior",
        "office library nook, built-in shelving, club chairs, reading lamps, "
        "no people, boutique corporate interior",
        "open office with exposed ductwork, polished concrete, pendant Edison bulbs, "
        "no people, industrial office photography",
        "executive suite anteroom, assistant's desk, fresh flowers, marble floors, "
        "no people, luxury corporate photography",
        "hot-desking floor, clean unassigned workstations, lockers along wall, "
        "no people, modern agile workplace",
        "office print and mail room, rows of equipment, organized shelving, "
        "no people, corporate interior detail",
        "executive boardroom anteroom, leather bench seating, artwork, "
        "no people, upscale corporate reception",
        "corner office with vintage credenza, globe, leather blotter, "
        "no people, classic executive interior",
        "open plan office with frosted glass pods, acoustic carpeting, "
        "no people, contemporary workplace design",
    ],
    "exterior": [
        "glass and steel corporate headquarters exterior, blue sky, "
        "architectural photography, no people, sharp detail",
        "modern office park, manicured grounds, reflecting pool, no people, "
        "architectural photography",
        "corporate campus exterior, multiple interconnected glass buildings, "
        "no people, professional photography, wide angle",
        "downtown office tower from street level, glass facade, no people, "
        "architectural photography",
        "brutalist concrete office building exterior, overcast sky, "
        "no people, architectural photography, dramatic shadows",
        "mid-century modern corporate campus, low-rise glass buildings, manicured lawn, "
        "no people, architectural photography",
        "contemporary headquarters with green roof garden, solar panels, "
        "no people, sustainable architecture photography",
        "financial district viewed from across a river, skyline reflection in water, "
        "no people, wide angle photography",
        "glass tower entrance canopy, stainless steel lettering above doors, "
        "no people, architectural detail photography",
        "curved glass skyscraper exterior detail, reflections of clouds, "
        "no people, architectural photography",
        "corporate campus courtyard between buildings, sculpture, paving stones, "
        "no people, architectural photography",
        "parking structure with dramatic geometric shadows, "
        "no people, architectural detail photography",
        "postmodern office tower, decorative crown, setbacks, terracotta cladding, "
        "no people, architectural photography",
        "glass office tower base, public plaza, granite paving, mature trees, "
        "no people, urban architectural photography",
        "corporate headquarters at night, illuminated facade, empty forecourt, "
        "no people, architectural night photography",
        "regional office park, low buildings, American flag, manicured hedge, "
        "no people, corporate real estate photography",
        "neoclassical corporate headquarters, columns, stone steps, brass doors, "
        "no people, traditional architectural photography",
        "glass curtain wall detail, structural fins, abstract geometry, "
        "no people, architectural abstraction",
        "corporate headquarters seen through ornamental gate, long driveway, "
        "no people, estate architecture photography",
        "campus connector bridge between two towers, glass enclosed, "
        "no people, architectural photography",
        "headquarters courtyard water feature, corporate sculpture, "
        "no people, landscape architecture photography",
        "office tower base retail podium, empty plaza, granite benches, "
        "no people, urban architecture photography",
    ],
    "abstract": [
        "blurred bokeh office background, professional corporate, "
        "warm neutral tones, depth of field",
        "abstract close-up of glass building facade, geometric reflections, "
        "blue and grey tones",
        "shallow depth of field view through office window, city buildings blurred, "
        "warm tones",
        "abstract dark blue and gold gradient background, subtle texture, "
        "corporate presentation style",
        "macro close-up of brushed stainless steel surface, soft reflections, "
        "abstract corporate texture",
        "abstract overhead view of empty conference table, "
        "geometric composition, cool tones, minimal",
        "blurred background of office plants and natural light, "
        "soft green and cream tones",
        "close-up of woven fabric office chair, geometric pattern, "
        "neutral tones, abstract texture",
        "abstract view through rain-spotted office window, city lights blurred, "
        "moody, desaturated tones",
        "overhead drone shot of corporate parking lot, geometric lines, "
        "abstract composition, muted tones, no people",
        "close-up of modern building exterior louvres, repeating pattern, "
        "abstract architectural photography",
        "marble reception desk surface close-up, white and grey veining, "
        "abstract luxury corporate texture",
        "close-up of perforated aluminum ceiling tile, repeating dot grid, "
        "abstract corporate texture",
        "long exposure of empty office corridor, light trails from cleaning cart, "
        "abstract motion blur, no people",
        "frosted glass partition close-up, blurred shapes behind, "
        "abstract corporate privacy glass",
        "glass elevator shaft interior, abstract geometry, cable and counterweight, "
        "no people, architectural abstraction",
        "corporate carpet tile close-up, geometric pattern, muted blue and grey, "
        "abstract texture photography",
        "tinted window reflection of building opposite, abstract distortion, "
        "architectural photography",
        "escalator handrail close-up, moving blur, abstract corporate motion, "
        "long exposure photography",
        "wood veneer wall panel close-up, straight grain, warm tones, "
        "abstract corporate material texture",
    ],
    "data_tech": [
        "server room corridor, dramatic lighting, blue tones, no people, "
        "technology photography, sharp detail",
        "data center rows of servers, ambient blue LED lighting, no people, "
        "professional technology photography",
        "close-up of network cable panel, clean organized cables, "
        "no people, technology detail photography",
        "empty control room with multiple monitors displaying data, "
        "no people, dramatic lighting, technology photography",
        "fiber optic cables glowing in darkness, abstract technology, "
        "no people, macro photography, vivid colors",
        "clean room interior, white walls, filtered lighting, "
        "no people, technology facility photography",
        "UPS battery room, rows of black server racks, status LEDs, "
        "no people, technology photography",
        "network operations center, rows of empty workstations facing screen wall, "
        "no people, technology interior",
        "cooling infrastructure beneath raised data center floor, "
        "no people, industrial technology photography",
        "close-up of circuit board, green and gold, macro photography, "
        "no people, technology abstract",
        "tape library robot arm extended in darkened vault, no people, "
        "technology photography, dramatic light",
        "server room hot aisle containment, orange overhead panels, "
        "no people, modern data center photography",
        "satellite dish array on rooftop, no people, "
        "technology infrastructure photography",
        "power distribution unit close-up, LED indicators, cable management, "
        "no people, technology detail photography",
        "empty trading floor at night, screens glowing with market data, "
        "no people, finance technology photography",
        "structured cabling room, patch panels, colored cables, "
        "no people, infrastructure photography",
        "raised floor data center aisle, perforated tiles, cool blue light, "
        "no people, technology photography",
        "biometric security door reader, keypad, stainless frame, "
        "no people, corporate security technology",
        "solar panel array on corporate roof, geometric grid, blue sky, "
        "no people, sustainable technology photography",
        "outdoor communications tower, blinking red lights, dusk sky, "
        "no people, infrastructure photography",
    ],
    "city": [
        "downtown financial district skyline at dusk, city lights beginning, "
        "no people, professional photography",
        "aerial view of city grid, buildings and streets, golden hour light, "
        "no people, drone photography",
        "city street lined with corporate buildings, late afternoon, "
        "no people visible, professional photography",
        "foggy city skyline at dawn, glass towers emerging from mist, "
        "no people, moody atmospheric photography",
        "aerial view of downtown at night, office windows lit, "
        "no people, long exposure, city lights",
        "looking up between two glass skyscrapers, narrow strip of sky, "
        "no people, dramatic urban photography",
        "elevated train platform with city skyline behind, no people, "
        "urban photography",
        "quiet downtown plaza, fountain, office towers surrounding, "
        "no people, architectural photography",
        "highway overpass with financial district skyline in background, "
        "no people, urban photography",
        "rooftop view of city from office building, no people, "
        "drone perspective, wide angle",
        "bridge over river with financial district skyline beyond, no people, "
        "urban landscape photography",
        "city canyon shot looking down empty street, tall buildings both sides, "
        "no people, urban photography",
        "downtown at blue hour, wet reflective pavement, office lights, "
        "no people, atmospheric city photography",
        "helicopter pad on skyscraper roof, city sprawl beyond, "
        "no people, aerial photography",
        "waterfront promenade, office towers reflected in bay, "
        "no people, urban landscape photography",
        "freeway interchange, elevated roads looping, city behind, "
        "no people, infrastructure photography",
        "city park with office towers visible above tree canopy, "
        "no people, urban green space photography",
        "downtown pedestrian bridge, cable-stayed, empty, towers behind, "
        "no people, architectural photography",
        "city marina, sailboats, financial district skyline, "
        "no people, urban waterfront photography",
        "rooftop mechanical penthouse, HVAC units, antenna, city panorama, "
        "no people, urban rooftop photography",
    ],
    "nature_corporate": [
        "manicured corporate campus lawn, ornamental trees in autumn color, "
        "no people, landscape photography",
        "minimalist Japanese garden adjacent to glass office building, "
        "no people, landscape photography",
        "formal hedge maze pattern on corporate grounds, aerial view, "
        "no people, landscape photography",
        "wide prairie at sunrise behind silhouette of low corporate campus, "
        "no people, landscape photography",
        "mountain range visible behind suburban office park, clear sky, "
        "no people, landscape photography",
        "tree-lined corporate driveway, autumn foliage, "
        "no people, landscape photography",
        "corporate campus wildflower meadow with building in background, "
        "no people, editorial photography",
        "industrial waterfront with glass office towers reflected in still harbor, "
        "no people, landscape photography",
        "winter snowfall on corporate campus, bare trees, lit windows, "
        "no people, atmospheric photography",
        "spring cherry blossoms framing corporate headquarters entry, "
        "no people, seasonal landscape photography",
        "corporate campus pond, weeping willows, glass building reflection, "
        "no people, landscape photography",
        "foggy morning on corporate grounds, dew on grass, mist in trees, "
        "no people, atmospheric landscape",
        "corporate campus in autumn, fallen leaves on plaza, brick path, "
        "no people, seasonal photography",
        "desert landscape with modern low-rise corporate campus, cacti, clear sky, "
        "no people, arid landscape photography",
        "coastal corporate campus, dramatic cliffs, glass building edge, "
        "no people, landscape architecture photography",
        "corporate campus in heavy rain, puddles reflecting building, umbrellas absent, "
        "no people, moody weather photography",
        "formal allée of plane trees leading to corporate building, "
        "no people, landscape photography",
        "corporate campus lake at sunset, geese absent, glassy reflection, "
        "no people, golden hour landscape",
    ],
    "hospitality_venue": [
        "empty hotel ballroom, chandeliers, round tables set for event, "
        "no people, event venue photography",
        "corporate retreat lodge, stone fireplace, leather sofas, mountain view, "
        "no people, hospitality photography",
        "convention center atrium, soaring glass ceiling, escalators, "
        "no people, architectural photography",
        "empty auditorium with theater-style seating facing stage with podium, "
        "no people, event photography",
        "hotel rooftop bar, city skyline view, empty bar stools, "
        "no people, hospitality photography",
        "executive airport lounge, leather chairs, city view, "
        "no people, hospitality interior photography",
        "empty museum atrium with marble floors and skylight, "
        "no people, architectural photography",
        "upscale hotel corridor, repeating doors, artwork, warm sconces, "
        "no people, hospitality photography",
        "private dining room, round table for ten, crystal glasses set, chandelier, "
        "no people, fine dining photography",
        "conference center pre-function space, cocktail tables, empty, "
        "no people, event venue photography",
        "corporate retreat spa, stone basin, candles, bamboo, "
        "no people, wellness venue photography",
        "yacht club dining room, nautical decor, harbor view, "
        "no people, hospitality interior",
        "historic hotel library bar, dark wood, brass, leather club chairs, "
        "no people, hospitality interior photography",
        "rooftop event terrace, pergola, city view, string lights unlit, "
        "no people, event venue photography",
        "breakout session room, rounds of four, flip chart stands, "
        "no people, meeting venue photography",
        "amphitheater outdoor venue, curved seating, stage, "
        "no people, event space photography",
        "high-speed rail first class lounge, marble, champagne bar empty, "
        "no people, luxury travel photography",
        "corporate retreat activity room, billiard table, chess boards, "
        "no people, upscale amenity photography",
    ],
    "manufacturing_industrial": [
        "clean modern factory floor, automated assembly line, no people, "
        "industrial photography, dramatic overhead lighting",
        "empty logistics warehouse, rows of shelving, high ceiling, "
        "no people, industrial interior photography",
        "pharmaceutical manufacturing clean room, stainless steel equipment, "
        "no people, industrial photography",
        "empty shipping dock, loading bays, industrial doors open, "
        "no people, industrial photography",
        "large industrial printing press room, enormous machinery, "
        "no people, industrial photography",
        "chemical plant exterior at dusk, pipes and towers, safety lighting, "
        "no people, industrial photography",
        "automated robotic welding station, sparks implied, no people, "
        "industrial technology photography",
        "food production facility, stainless conveyors, hygienic white walls, "
        "no people, industrial interior photography",
        "wind turbine blade factory, giant composite forms on fixtures, "
        "no people, industrial manufacturing photography",
        "aerospace assembly hangar, large aircraft components on jigs, "
        "no people, industrial interior photography",
        "cold storage distribution center, frost-covered racking, forklifts absent, "
        "no people, industrial photography",
        "automated high-bay warehouse, conveyor belts, sorting systems, "
        "no people, logistics photography",
        "steel mill furnace room, glowing metal, heat haze, no people, "
        "industrial photography, dramatic orange light",
        "microchip fab cleanroom, yellow filtered lights, equipment, "
        "no people, semiconductor manufacturing photography",
        "engine test cell, turbine on stand, instrumentation panels, "
        "no people, industrial testing photography",
        "cement plant interior, kilns, dust-covered machinery, "
        "no people, heavy industrial photography",
        "bottling plant, empty conveyor lines, glass bottles in rows, "
        "no people, food beverage manufacturing",
        "shipyard dry dock, vessel in cradle, scaffolding, "
        "no people, industrial photography",
    ],
    "research_education": [
        "empty university lecture hall, tiered seating, projection screen, "
        "no people, architectural photography",
        "corporate R&D lab, white benches, equipment, clean, "
        "no people, science photography",
        "empty library reading room, long tables, lamps, bookshelves, "
        "no people, interior photography",
        "innovation center open studio, whiteboards, maker equipment, "
        "no people, interior photography",
        "telescope dome exterior at twilight, star field beginning, "
        "no people, science photography",
        "empty operating theater from observation gallery, surgical lights, "
        "no people, medical photography",
        "law library, floor-to-ceiling bound volumes, brass rail ladder, "
        "no people, interior photography",
        "pharmaceutical research lab, fume hoods, sample racks, no people, "
        "science interior photography",
        "wind tunnel test section, scale model mounted, empty, "
        "no people, aerospace research photography",
        "empty MRI suite, white scanner, clinical environment, "
        "no people, medical facility photography",
        "university atrium, exposed concrete, student commons empty, "
        "no people, academic architecture photography",
        "corporate training center, tiered classroom, podium, screens, "
        "no people, modern education photography",
        "think tank seminar room, circular table, whiteboards, "
        "no people, research interior photography",
        "patent office reading room, long tables, reference volumes, "
        "no people, institutional interior photography",
        "corporate archive vault, shelved binders, file cabinets, dim light, "
        "no people, institutional photography",
        "simulation lab, wall-mounted screens, control station, "
        "no people, technology research photography",
        "natural history museum storage, specimen drawers, cabinets, "
        "no people, institutional photography",
        "corporate test kitchen, stainless counters, commercial range, "
        "no people, culinary research photography",
    ],
    "finance_trading": [
        "empty stock exchange trading floor, horseshoe booths, screens dark, "
        "no people, financial photography",
        "investment bank dealing room, rows of multi-screen workstations empty, "
        "no people, corporate finance interior",
        "hedge fund portfolio management floor, panoramic city view, empty chairs, "
        "no people, financial interior photography",
        "private equity boardroom, dark walnut, leather, framed deal tombstones, "
        "no people, corporate interior",
        "bank vault interior, circular door open, safe deposit boxes, "
        "no people, financial institution photography",
        "central bank gold reserve room, stacked gold bars, security lighting, "
        "no people, institutional photography",
        "currency trading pit, octagonal booth, screens, empty, "
        "no people, financial exchange photography",
        "wealth management client suite, contemporary art, harbor view, "
        "no people, luxury financial interior",
        "compliance monitoring room, multiple screens, alert dashboards, "
        "no people, financial technology photography",
        "bank branch lobby, marble counters, teller windows, "
        "no people, retail banking photography",
        "actuarial office, rows of cubicles, dual monitors, empty, "
        "no people, insurance corporate interior",
        "prime brokerage floor, risk screens, order terminals, "
        "no people, financial services photography",
        "economic research library, journals, data terminals, "
        "no people, financial research interior",
        "commodities exchange floor, open outcry pit, hand signal boards, "
        "no people, exchange photography",
        "private banking consultation room, low lighting, bespoke furniture, "
        "no people, luxury banking interior",
    ],
    "transport_infrastructure": [
        "empty international airport terminal, gates, moving walkway, "
        "no people, architectural photography",
        "high-speed train on platform, sleek exterior, empty platform, "
        "no people, transportation photography",
        "container port aerial view, stacked containers, cranes, "
        "no people, logistics photography",
        "highway interchange aerial, geometric road curves, "
        "no people, infrastructure photography",
        "airport control tower and runway at dusk, aircraft absent, "
        "no people, aviation photography",
        "suspension bridge pylon close-up, cables radiating, "
        "no people, infrastructure photography",
        "corporate jet on ramp, executive terminal, empty, "
        "no people, business aviation photography",
        "metro station platform, arched tunnel, empty, "
        "no people, transit architecture photography",
        "cargo aircraft interior, empty freight hold, pallets, "
        "no people, aviation logistics photography",
        "deep-water port quay, mooring bollards, empty berth, "
        "no people, maritime infrastructure photography",
        "logistics hub at night, trucks absent, illuminated loading apron, "
        "no people, industrial infrastructure photography",
        "monorail station platform, futuristic canopy, city below, "
        "no people, urban transit photography",
        "pipeline pumping station exterior, valves, gauges, no people, "
        "energy infrastructure photography",
        "electric vehicle charging plaza, canopied stalls empty, "
        "no people, sustainable transport photography",
        "cable car aerial tramway, cabin descending, mountain corporate campus, "
        "no people, mountain transport photography",
    ],
}

# Lighting modifiers
_LIGHTING_MODS = [
    "golden hour lighting",
    "overcast diffused light",
    "dramatic side lighting",
    "soft morning light",
    "blue hour ambient light",
    "bright midday sun",
    "warm incandescent ambient",
    "cool fluorescent overhead",
    "dusk twilight lighting",
    "crisp winter daylight",
    "pre-dawn blue light",
    "harsh noon shadows",
    "warm sunset backlight",
    "diffused cloudy sky",
    "dramatic storm light",
    "neon-lit night scene",
    "candlelight warmth",
    "soft studio strobe",
    "raking late afternoon sun",
    "fog-filtered pale light",
]

# Photographic style modifiers
_STYLE_MODS = [
    "photorealistic, 8k, sharp focus",
    "professional photography, high resolution",
    "editorial photography, crisp detail",
    "architectural photography, award winning",
    "commercial photography, magazine quality",
    "fine art photography, exquisite detail",
    "documentary photography, natural tones",
    "luxury real estate photography",
    "corporate annual report photography",
    "Architectural Digest style photography",
    "minimalist photography, clean composition",
    "dramatic editorial photography",
    "travel photography, vivid tones",
    "environmental photography, rich detail",
]

# Atmospheric / weather modifiers
_ATMOSPHERE_MODS = [
    "clear sky",
    "light morning haze",
    "dramatic storm clouds",
    "light rain, wet surfaces",
    "heavy fog",
    "fresh snow on ground",
    "autumn leaves on ground",
    "heat shimmer",
    "thin cirrus clouds",
    "post-rain freshness, puddles",
    "wildfire smoke tint, amber cast",
    "crisp clear winter air",
]

# Lens / composition modifiers
_LENS_MODS = [
    "wide angle lens",
    "telephoto compression",
    "tilt-shift miniature effect",
    "fisheye perspective",
    "macro close-up detail",
    "long exposure blur",
    "drone bird's eye view",
    "worm's eye upward angle",
    "symmetrical composition",
    "rule of thirds framing",
    "leading lines composition",
    "frame within frame",
]

_NEGATIVE_PROMPT = (
    "people, person, human, face, body, hands, crowd, portrait, selfie, "
    "text, watermark, logo, cartoon, illustration, painting, drawing, "
    "low quality, blurry, overexposed, underexposed"
)


def _build_prompts(count: int, seed_prefix: str = "bg") -> list[tuple[str, str, str]]:
    """Return list of (seed, category, prompt) tuples."""
    all_prompts = []
    for cat, prompts in _CATEGORIES.items():
        for prompt in prompts:
            all_prompts.append((cat, prompt))

    results = []
    rng = random.Random(int(hashlib.md5(seed_prefix.encode()).hexdigest(), 16) % (2 ** 32))

    for i in range(count):
        # Random selection (seeded) instead of cycling — avoids identical prompt repeats
        cat, base_prompt = rng.choice(all_prompts)
        lighting = rng.choice(_LIGHTING_MODS)
        style = rng.choice(_STYLE_MODS)
        atmosphere = rng.choice(_ATMOSPHERE_MODS)
        lens = rng.choice(_LENS_MODS)
        prompt = f"{base_prompt}, {lighting}, {atmosphere}, {lens}, {style}"
        results.append((f"{seed_prefix}_{i:05d}", cat, prompt))
    return results


def _pick_device(torch):
    if torch.backends.mps.is_available():
        print("Using Apple MPS (Metal Performance Shaders)")
        return "mps"
    if torch.cuda.is_available():
        print(f"Using CUDA: {torch.cuda.get_device_name(0)}")
        return "cuda"
    print("Using CPU (slow)")
    return "cpu"


def _seed_for(seed_str: str) -> int:
    return int(hashlib.md5(seed_str.encode()).hexdigest(), 16) % (2 ** 32)


def main():
    parser = argparse.ArgumentParser(description="Generate background images for presentations")
    parser.add_argument("--count", type=int, default=60, help="Number of images to generate")
    parser.add_argument("--force", action="store_true", help="Regenerate existing files")
    parser.add_argument("--dry-run", action="store_true", help="Preview prompts only")
    parser.add_argument("--steps", type=int, default=2, help="Diffusion steps (default 2)")
    parser.add_argument("--model", default="stabilityai/sdxl-turbo",
                        help="Model ID (default: stabilityai/sdxl-turbo)")
    parser.add_argument("--width", type=int, default=800)
    parser.add_argument("--height", type=int, default=448)  # must be divisible by 8 for SDXL
    args = parser.parse_args()

    prompts = _build_prompts(args.count)

    if args.dry_run:
        print(f"Would generate {len(prompts)} images to {OUT_DIR}")
        for seed, cat, prompt in prompts[:10]:
            print(f"\n  [{cat}] {seed}")
            print(f"  {prompt[:100]}…")
        if len(prompts) > 10:
            print(f"\n  … and {len(prompts) - 10} more")
        return

    try:
        import torch
        from diffusers import AutoPipelineForText2Image
        from PIL import Image
    except ImportError as e:
        print(f"ERROR: Missing dependency: {e}", file=sys.stderr)
        print("Run: pip install diffusers transformers accelerate pillow torch", file=sys.stderr)
        sys.exit(1)

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    device = _pick_device(torch)

    print(f"Loading model {args.model}…")
    dtype = torch.float16 if device != "cpu" else torch.float32
    pipe = AutoPipelineForText2Image.from_pretrained(
        args.model,
        torch_dtype=dtype,
        variant="fp16" if device != "cpu" else None,
    ).to(device)
    pipe.set_progress_bar_config(disable=True)

    generated = skipped = errors = 0
    for seed_str, cat, prompt in prompts:
        out_path = OUT_DIR / f"{seed_str}.webp"
        if out_path.exists() and not args.force:
            skipped += 1
            continue
        try:
            generator = torch.Generator(device=device).manual_seed(_seed_for(seed_str))
            result = pipe(
                prompt=prompt,
                negative_prompt=_NEGATIVE_PROMPT,
                num_inference_steps=args.steps,
                guidance_scale=0.0,  # SDXL-Turbo uses guidance_scale=0
                width=args.width,
                height=args.height,
                generator=generator,
            )
            img = result.images[0]
            img.save(str(out_path), "WEBP", quality=85)
            generated += 1
            if generated % 5 == 0:
                print(f"  {generated}/{len(prompts)} generated…")
        except Exception as e:
            errors += 1
            print(f"  ERROR [{seed_str}]: {e}", file=sys.stderr)

    print(f"\nDone. Generated: {generated}, Skipped: {skipped}, Errors: {errors}")
    print(f"Output: {OUT_DIR}")
    if generated:
        print("\nNext step: docker compose exec web python manage.py collectstatic --noinput")


if __name__ == "__main__":
    main()
