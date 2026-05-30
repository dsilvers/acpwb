#!/usr/bin/env python3
"""
Standalone meme image generator for ACPWB presentation slides — no Django required.

Generates two meme formats using Pillow + system fonts:
  1. top_bottom — Impact text over dark gradient background
  2. matrix     — 2×2 BCG-style quadrant with axis labels

Text is drawn from the same pools used by slide generators so memes stay
content-relevant to the presentation subjects.

Install:
    pip install pillow

Usage:
    python meme_generator.py                    # generate all 80 memes
    python meme_generator.py --count 40         # generate 40 memes
    python meme_generator.py --format top_bottom  # only top/bottom format
    python meme_generator.py --format matrix      # only matrix format
    python meme_generator.py --force            # regenerate existing
    python meme_generator.py --dry-run          # preview text only
"""

import argparse
import hashlib
import random
import sys
from pathlib import Path
from textwrap import wrap

HERE = Path(__file__).resolve().parent
OUT_DIR = HERE / "acpwb" / "static" / "img" / "presentations" / "memes"

# ---------------------------------------------------------------------------
# Content pools (mirrors slide_templates.py vocabulary, no import needed)
# ---------------------------------------------------------------------------

_DOMAINS = [
    "Human Capital", "Change Management", "Digital Transformation", "Risk Management",
    "Operational Excellence", "Talent Acquisition", "Workforce Planning", "ESG Strategy",
    "Data Governance", "Compliance", "Innovation", "Customer Experience",
    "Leadership Development", "Supply Chain", "Strategic Planning",
]

_INDUSTRIES = [
    "Healthcare", "Financial Services", "Technology", "Manufacturing",
    "Energy", "Retail", "Government", "Education", "Defense", "Consulting",
]

_NOUNS = [
    "Roadmap", "Framework", "Initiative", "Strategy", "Pipeline",
    "Stakeholder", "Deliverable", "Synergy", "Bandwidth", "Alignment",
    "Ecosystem", "Playbook", "Leverage", "Traction", "Velocity",
]

_ADJECTIVES = [
    "Agile", "Scalable", "Robust", "Holistic", "Proactive",
    "Strategic", "Dynamic", "Transformational", "Integrated", "Data-Driven",
    "Cross-Functional", "Best-in-Class", "Innovative", "Streamlined",
]

_VERBS = [
    "Optimize", "Leverage", "Accelerate", "Transform", "Align",
    "Streamline", "Operationalize", "Prioritize", "Harmonize", "Democratize",
]

# (setup_text, punchline_text) pairs — {placeholders} filled by seeded RNG
_TOP_BOTTOM_TEMPLATES = [
    ("When the {noun} {roadmap} gets approved", "But there's no budget for implementation"),
    ("Finally aligns all {adj} stakeholders", "On the wrong {noun}"),
    ("Achieves {adj} {domain} transformation", "By renaming the existing {noun}"),
    ("Presents the Q4 {domain} strategy", "It's the Q2 strategy with new fonts"),
    ("Leadership asks for {adj} innovation", "Then rejects every proposed change"),
    ("The {domain} {noun} is on track", "The definition of 'on track' has changed"),
    ("Completes {adj} organizational redesign", "Everyone now reports to everyone else"),
    ("Builds {adj} cross-functional alignment", "Meeting scheduled: 47 attendees, 0 decisions"),
    ("Delivers {verb} {domain} framework", "Immediately launches Phase 2 framework"),
    ("{adj} {noun} deployed successfully", "Nobody told the {industry} team"),
    ("Data-driven decision making mandate", "The data says what leadership already decided"),
    ("Brings in external {industry} consultants", "They recommend what the team suggested in 2021"),
    ("Launches {adj} employee engagement survey", "Results filed under 'To Review Q3'"),
    ("Mandates {adj} {domain} training", "Completion metric: clicked through slides"),
    ("Strategic {noun} identified and prioritized", "Added to backlog, never to be seen again"),
    ("Achieves {adj} operational excellence", "By eliminating the people who knew how it worked"),
    ("The {adj} transformation is complete", "Phase 1 of 7 complete"),
    ("New {domain} {noun} announced", "Same {domain} {noun} as last year, different name"),
    ("{verb} the core {domain} capabilities", "Core capabilities remain unverified"),
    ("All {noun} milestones green on dashboard", "Dashboard updated, milestones not reviewed"),
    # --- batch 2 ---
    ("Schedules {adj} offsite to reset {domain} vision", "Same vision, different hotel"),
    ("Executive sponsors the {adj} {domain} initiative", "Then leaves the company in week two"),
    ("Announces {adj} culture transformation", "Hires consultant to define what culture means"),
    ("Requires {adj} weekly status reports", "Nobody reads the status reports"),
    ("Empowers {domain} team with {adj} autonomy", "Approvals now require four extra sign-offs"),
    ("{verb} the {adj} {noun} by end of quarter", "Quarter ends, noun undefined"),
    ("Posts {adj} {domain} thought leadership", "Written by intern, approved by committee"),
    ("Introduces {adj} {noun} governance model", "Nobody can explain the governance model"),
    ("Celebrates {adj} {domain} milestone", "Milestone was moved from last quarter"),
    ("We need more {adj} collaboration", "Adds another recurring Tuesday meeting"),
    ("Rolls out {adj} knowledge management system", "All knowledge stored in email threads"),
    ("Demands {adj} accountability across {domain}", "Accountability not defined or measured"),
    ("Launches pilot program to {verb} {noun}", "Pilot runs for three years, never scaled"),
    ("{adj} {domain} center of excellence established", "Three people, no budget, no mandate"),
    ("Creates {adj} task force to {verb} {domain}", "Task force creates a subcommittee"),
    ("Leadership commits to {adj} transparency", "Town hall questions screened in advance"),
    ("{verb} all {domain} processes end-to-end", "Discovered there are no documented processes"),
    ("Benchmarks against {adj} {industry} leaders", "Benchmark shows we are not the leader"),
    ("Achieves {adj} best-in-class {domain}", "As defined by our own internal metrics"),
    ("Announces zero-based {domain} budgeting", "Same budgets, new spreadsheet template"),
    # --- batch 3 ---
    ("Forms {adj} cross-functional tiger team", "Tiger team disbanded after first conflict"),
    ("Deploys {adj} {domain} chatbot solution", "Chatbot answers questions nobody asked"),
    ("{adj} {noun} workshop scheduled for Friday", "Workshop produces deck, no action items"),
    ("Rebrands {domain} team as a {adj} {noun} hub", "Same team, new lanyards"),
    ("Hires {adj} {domain} transformation lead", "Lead inherits broken {noun} from predecessor"),
    ("Sets {adj} OKRs for {domain} excellence", "OKRs not reviewed until Q4 retrospective"),
    ("Conducts {adj} root cause analysis", "Root cause: the previous root cause analysis"),
    ("Implements {adj} continuous improvement culture", "No improvements made this quarter"),
    ("Delivers {adj} executive briefing on {domain}", "Briefing requests a follow-up briefing"),
    ("Secures executive buy-in for {adj} {noun}", "Executive forgets buy-in by next sprint"),
    ("Runs {adj} design thinking sprint on {domain}", "Output: sticky notes and a parking lot"),
    ("Optimizes {adj} {domain} for scalability", "Solution cannot handle current load"),
    ("Issues {adj} {domain} white paper", "White paper is 80% definitions"),
    ("Requests {adj} {domain} business case", "Business case requires another business case"),
    ("Builds {adj} {noun} roadmap through 2027", "Roadmap revised in 2025"),
    ("Introduces {adj} {domain} operating model", "Nobody can draw the operating model"),
    ("{verb} {domain} outcomes for all stakeholders", "Stakeholder list not yet finalized"),
    ("Launches {adj} {industry} partnership program", "No partners have joined yet"),
    ("Identifies {noun} as a strategic priority", "Priority number seventeen of seventeen"),
    ("All {domain} teams aligned on {adj} {noun}", "Teams learn of alignment via press release"),
]

_MATRIX_TEMPLATES = [
    # (x_axis_label, y_axis_label, quadrant_labels: [bottom-left, bottom-right, top-left, top-right])
    (
        "Urgency",
        "Importance",
        ["Delegate", "Do First", "Drop", "Schedule"],
    ),
    (
        "Implementation Difficulty",
        "Strategic Impact",
        ["Avoid", "Quick Win", "Long-Term Bet", "Major Project"],
    ),
    (
        "Cost to Implement",
        "Expected ROI",
        ["Waste of Budget", "Low-Hanging Fruit", "Worth Exploring", "Strategic Priority"],
    ),
    (
        "{domain} Maturity",
        "Executive Buy-In",
        ["Orphaned Initiative", "Needs Enablement", "Grassroots Only", "Full Go"],
    ),
    (
        "Stakeholder Resistance",
        "Business Value",
        ["Kill It", "Force Through", "Evangelize", "Protect at All Costs"],
    ),
    (
        "Data Quality",
        "Decision Frequency",
        ["Ignore", "Automate Anyway", "Fix First", "Highest Priority"],
    ),
    (
        "Effort Required",
        "{adj} {domain} Value",
        ["Not Worth It", "Quick Win", "Strategic Bet", "Transform Now"],
    ),
    (
        "Change Resistance",
        "Productivity Impact",
        ["Abandon", "Train Harder", "Culture Work", "Drive It"],
    ),
    # --- batch 2 ---
    (
        "Process Complexity",
        "Frequency of Use",
        ["Document and Archive", "Simplify Now", "Automate Later", "Automate First"],
    ),
    (
        "Vendor Lock-In Risk",
        "Capability Gap Addressed",
        ["Hard Pass", "Proceed with Caution", "Negotiate Hard", "Sign Today"],
    ),
    (
        "Time to Value",
        "Confidence in Outcome",
        ["Avoid", "Experiment", "Plan Carefully", "Execute Now"],
    ),
    (
        "Regulatory Risk",
        "Revenue Impact",
        ["Deprioritize", "Monitor Closely", "Escalate", "Board-Level Priority"],
    ),
    (
        "Team Bandwidth",
        "{adj} Strategic Alignment",
        ["Backlog Forever", "Defer to Next Cycle", "Hire for It", "Start Monday"],
    ),
    (
        "Political Sensitivity",
        "Operational Necessity",
        ["Leave It Alone", "Handle Quietly", "Communicate Broadly", "Mandate It"],
    ),
    (
        "Customer Impact",
        "Implementation Speed",
        ["Low Priority", "Phase Two", "Fast Follow", "Drop Everything"],
    ),
    (
        "{domain} Debt",
        "Growth Dependency",
        ["Live With It", "Schedule Cleanup", "Address This Quarter", "Stop the Line"],
    ),
    (
        "Headcount Required",
        "Competitive Differentiation",
        ["Outsource", "Automate", "Upskill Internally", "Build a Team"],
    ),
    (
        "Execution Risk",
        "Market Timing",
        ["Not Now", "Pilot First", "Move Fast", "All In"],
    ),
    # --- batch 3 ---
    (
        "Stakeholder Enthusiasm",
        "Feasibility",
        ["Fantasy", "Needs a Champion", "Viable But Lonely", "Ship It"],
    ),
    (
        "Documentation Quality",
        "Team Dependence on Process",
        ["Nobody Cares", "Nice to Have", "Fix the Docs", "Critical Risk"],
    ),
    (
        "Budget Certainty",
        "Strategic Importance",
        ["Wish List", "Contingency Plan", "Find the Budget", "Non-Negotiable"],
    ),
    (
        "Reversibility",
        "Potential Upside",
        ["Just Don't", "Hedge First", "Limit Exposure", "Go Bold"],
    ),
    (
        "Cross-Team Dependency",
        "{adj} Business Impact",
        ["Solo Project", "Coordinate Lightly", "Needs a Sponsor", "Program Office"],
    ),
    (
        "Internal Capability",
        "Problem Urgency",
        ["Not Our Problem", "Buy a Solution", "Build It Slowly", "War Room"],
    ),
    (
        "Innovation Risk",
        "Customer Demand Signal",
        ["Interesting Idea", "Test with Segment", "Prototype Now", "Launch It"],
    ),
    (
        "Clarity of Success Metrics",
        "Leadership Visibility",
        ["Vague and Quiet", "Define First", "Manage Carefully", "Full Spotlight"],
    ),
    (
        "Technical Complexity",
        "Business Readiness",
        ["Revisit in Two Years", "Build Readiness First", "Phased Approach", "Ready to Launch"],
    ),
    (
        "Talent Availability",
        "{domain} ROI Confidence",
        ["Not Yet", "Hire First", "Retrain Internally", "Accelerate Now"],
    ),
]

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _rng(seed_str: str) -> random.Random:
    h = int(hashlib.md5(seed_str.encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(h)


def _fill(text: str, rng: random.Random) -> str:
    replacements = {
        "noun": rng.choice(_NOUNS),
        "adj": rng.choice(_ADJECTIVES),
        "domain": rng.choice(_DOMAINS),
        "industry": rng.choice(_INDUSTRIES),
        "verb": rng.choice(_VERBS),
        "roadmap": rng.choice(["roadmap", "framework", "strategy", "initiative"]),
    }
    try:
        return text.format(**replacements)
    except KeyError:
        return text


def _find_font(name: str, fallback: str = "DejaVuSans.ttf"):
    """Find a font file by name on common system paths."""
    search_dirs = [
        Path("/System/Library/Fonts/Supplemental"),
        Path("/System/Library/Fonts"),
        Path("/Library/Fonts"),
        Path.home() / "Library" / "Fonts",
        Path("/usr/share/fonts"),
        Path("/usr/local/share/fonts"),
    ]
    for d in search_dirs:
        for p in d.rglob("*.ttf"):
            if p.stem.lower().replace(" ", "") == name.lower().replace(" ", ""):
                return str(p)
    # DejaVu is bundled with Pillow on most systems
    import PIL
    pil_dir = Path(PIL.__file__).parent / "fonts"
    if (pil_dir / fallback).exists():
        return str(pil_dir / fallback)
    return None


# ---------------------------------------------------------------------------
# Background generators for top/bottom meme
# ---------------------------------------------------------------------------

# Corporate color palettes: (primary, secondary, accent) — all dark enough for white text
_BG_PALETTES = [
    ((10, 22, 55), (20, 45, 90), (180, 140, 40)),    # navy/gold
    ((18, 42, 28), (30, 65, 40), (80, 180, 100)),    # forest/green
    ((50, 18, 18), (85, 28, 28), (220, 80, 60)),     # deep red/crimson
    ((22, 18, 45), (38, 28, 75), (140, 100, 220)),   # deep purple
    ((15, 38, 48), (22, 60, 75), (40, 180, 190)),    # teal/slate
    ((42, 28, 10), (70, 48, 15), (210, 155, 50)),    # bronze/amber
    ((20, 20, 20), (40, 40, 45), (160, 160, 175)),   # charcoal/silver
    ((8, 28, 50), (14, 48, 80), (200, 220, 240)),    # steel blue
]


def _make_bg_diagonal_blocks(rng, width, height, Image, ImageDraw):
    """Two or three angled color bands — looks like a conference slide header."""
    pal = rng.choice(_BG_PALETTES)
    img = Image.new("RGB", (width, height), pal[0])
    draw = ImageDraw.Draw(img)

    # Main diagonal band (lighter secondary color)
    skew = rng.randint(width // 4, width // 2)
    band_w = rng.randint(width // 3, width * 2 // 3)
    poly = [
        (skew, 0), (skew + band_w, 0),
        (skew + band_w - height // 3, height), (skew - height // 3, height),
    ]
    draw.polygon(poly, fill=pal[1])

    # Thin accent stripe
    stripe_x = skew + band_w + rng.randint(-30, 30)
    stripe_w = rng.randint(8, 28)
    poly2 = [
        (stripe_x, 0), (stripe_x + stripe_w, 0),
        (stripe_x + stripe_w - height // 3, height), (stripe_x - height // 3, height),
    ]
    draw.polygon(poly2, fill=pal[2])

    # Optional second accent stripe
    if rng.random() > 0.4:
        stripe_x2 = skew - rng.randint(20, 60)
        stripe_w2 = rng.randint(4, 14)
        poly3 = [
            (stripe_x2, 0), (stripe_x2 + stripe_w2, 0),
            (stripe_x2 + stripe_w2 - height // 3, height), (stripe_x2 - height // 3, height),
        ]
        draw.polygon(poly3, fill=tuple(min(255, c + 40) for c in pal[2]))

    return img


def _make_bg_bokeh(rng, width, height, Image, ImageDraw):
    """Blurred circles of varying size/opacity — corporate photography bokeh look."""
    from PIL import ImageFilter
    pal = rng.choice(_BG_PALETTES)
    img = Image.new("RGB", (width, height), pal[0])
    draw = ImageDraw.Draw(img, "RGBA")

    # Scatter ~30 soft circles
    for _ in range(rng.randint(18, 35)):
        cx = rng.randint(-50, width + 50)
        cy = rng.randint(-50, height + 50)
        r = rng.randint(20, 130)
        alpha = rng.randint(25, 90)
        color_choice = rng.choice([pal[1], pal[2],
                                    tuple(min(255, c + 60) for c in pal[1]),
                                    (255, 255, 255)])
        draw.ellipse([cx - r, cy - r, cx + r, cy + r],
                     fill=color_choice + (alpha,))

    # Blur to get the bokeh softness
    img = img.filter(ImageFilter.GaussianBlur(radius=rng.randint(8, 18)))
    return img


def _make_bg_grid(rng, width, height, Image, ImageDraw):
    """Subtle geometric grid/dot pattern — data visualization aesthetic."""
    pal = rng.choice(_BG_PALETTES)

    # Gradient base
    img = Image.new("RGB", (width, height), pal[0])
    draw = ImageDraw.Draw(img)
    for y in range(height):
        t = y / height
        r = int(pal[0][0] + (pal[1][0] - pal[0][0]) * t * 0.6)
        g = int(pal[0][1] + (pal[1][1] - pal[0][1]) * t * 0.6)
        b = int(pal[0][2] + (pal[1][2] - pal[0][2]) * t * 0.6)
        draw.line([(0, y), (width, y)], fill=(r, g, b))

    style = rng.choice(["dots", "lines", "cross"])
    grid_size = rng.choice([28, 36, 44])
    dot_r = grid_size // 6

    overlay = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    o_draw = ImageDraw.Draw(overlay)
    accent_a = rng.randint(35, 65)
    accent = pal[2] + (accent_a,)

    if style == "dots":
        for gx in range(0, width + grid_size, grid_size):
            for gy in range(0, height + grid_size, grid_size):
                o_draw.ellipse([gx - dot_r, gy - dot_r, gx + dot_r, gy + dot_r], fill=accent)
    elif style == "lines":
        lw = max(1, grid_size // 10)
        for gx in range(0, width + grid_size, grid_size):
            o_draw.line([(gx, 0), (gx, height)], fill=accent, width=lw)
        for gy in range(0, height + grid_size, grid_size):
            o_draw.line([(0, gy), (width, gy)], fill=accent, width=lw)
    else:  # cross
        lw = max(1, grid_size // 12)
        cross_r = dot_r * 2
        for gx in range(0, width + grid_size, grid_size):
            for gy in range(0, height + grid_size, grid_size):
                o_draw.line([(gx - cross_r, gy), (gx + cross_r, gy)], fill=accent, width=lw)
                o_draw.line([(gx, gy - cross_r), (gx, gy + cross_r)], fill=accent, width=lw)

    img = Image.alpha_composite(img.convert("RGBA"), overlay).convert("RGB")
    return img


def _make_bg_split(rng, width, height, Image, ImageDraw):
    """Bold horizontal or vertical color split — high contrast."""
    pal = rng.choice(_BG_PALETTES)
    img = Image.new("RGB", (width, height), pal[0])
    draw = ImageDraw.Draw(img)

    if rng.random() > 0.5:
        # Horizontal split
        split_y = rng.randint(height // 3, height * 2 // 3)
        draw.rectangle([0, split_y, width, height], fill=pal[1])
        # Thin accent line at split
        line_h = rng.randint(3, 10)
        draw.rectangle([0, split_y - line_h // 2, width, split_y + line_h // 2], fill=pal[2])
    else:
        # Vertical split with slight angle
        split_x = rng.randint(width // 3, width * 2 // 3)
        angle_offset = rng.randint(-40, 40)
        poly = [(0, 0), (split_x, 0), (split_x + angle_offset, height), (0, height)]
        draw.polygon(poly, fill=pal[1])
        # Accent stripe at split
        line_w = rng.randint(4, 12)
        split_poly = [
            (split_x - line_w, 0), (split_x + line_w, 0),
            (split_x + angle_offset + line_w, height), (split_x + angle_offset - line_w, height),
        ]
        draw.polygon(split_poly, fill=pal[2])

    return img


_BG_STYLES = [_make_bg_diagonal_blocks, _make_bg_bokeh, _make_bg_grid, _make_bg_split]


def _generate_background(rng, width, height, Image, ImageDraw):
    """Pick a background style and generate it, then darken for text legibility."""
    style_fn = rng.choice(_BG_STYLES)
    img = style_fn(rng, width, height, Image, ImageDraw)

    # Dark overlay so white text always reads — preserves color but ensures contrast
    overlay = Image.new("RGBA", (width, height), (0, 0, 0, rng.randint(90, 140)))
    img = Image.alpha_composite(img.convert("RGBA"), overlay).convert("RGB")

    # Vignette
    vig = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    v_draw = ImageDraw.Draw(vig)
    steps = 60
    for i in range(steps):
        alpha = int((i / steps) ** 1.5 * 160)
        v_draw.rectangle([i, i, width - i, height - i], outline=(0, 0, 0, alpha))
    img = Image.alpha_composite(img.convert("RGBA"), vig).convert("RGB")

    return img


# ---------------------------------------------------------------------------
# Meme format: top / bottom Impact text
# ---------------------------------------------------------------------------

def _draw_text_outlined(draw, xy, text, font, fill, stroke_fill, stroke_width=3):
    x, y = xy
    for dx in range(-stroke_width, stroke_width + 1):
        for dy in range(-stroke_width, stroke_width + 1):
            if dx != 0 or dy != 0:
                draw.text((x + dx, y + dy), text, font=font, fill=stroke_fill)
    draw.text(xy, text, font=font, fill=fill)


def generate_top_bottom(seed: str, width: int = 800, height: int = 450) -> "Image.Image":
    from PIL import Image, ImageDraw, ImageFont

    rng = _rng(seed)
    tmpl_top, tmpl_bottom = rng.choice(_TOP_BOTTOM_TEMPLATES)
    top_text = _fill(tmpl_top, rng).upper()
    bottom_text = _fill(tmpl_bottom, rng).upper()

    img = _generate_background(rng, width, height, Image, ImageDraw)
    draw = ImageDraw.Draw(img)

    impact_path = _find_font("Impact")
    font_size = max(36, width // 14)
    if impact_path:
        try:
            font = ImageFont.truetype(impact_path, font_size)
            font_small = ImageFont.truetype(impact_path, max(28, font_size - 8))
        except Exception:
            font = font_small = ImageFont.load_default()
    else:
        font = font_small = ImageFont.load_default()

    margin = width // 12

    def _place_text(text, font, y_top, max_w):
        lines = wrap(text, width=max(8, max_w // (font_size // 2)))
        y = y_top
        for line in lines:
            bbox = font.getbbox(line)
            lw = bbox[2] - bbox[0]
            x = (width - lw) // 2
            _draw_text_outlined(draw, (x, y), line, font,
                                 fill=(255, 255, 255), stroke_fill=(0, 0, 0), stroke_width=4)
            y += bbox[3] - bbox[1] + 6
        return y

    _place_text(top_text, font, margin // 2, width - margin * 2)

    bottom_lines = wrap(bottom_text, width=max(8, (width - margin * 2) // (font_size // 2)))
    bottom_total = sum(font_small.getbbox(l)[3] + 6 for l in bottom_lines)
    _place_text(bottom_text, font_small, height - bottom_total - margin // 2, width - margin * 2)

    return img


# ---------------------------------------------------------------------------
# Meme format: 2×2 management consulting matrix
# ---------------------------------------------------------------------------

# Each palette: bg, grid_line, axis_text, low_end_text, and quadrant (fill, label_color)
# Quadrant order: BL (bad/low), BR (mixed), TL (mixed), TR (good/high)
_MATRIX_PALETTES = [
    # Traffic light — red/amber/amber/green
    {
        "bg": (248, 248, 250), "grid": (80, 80, 90), "axis": (40, 40, 50), "low": (160, 160, 170),
        "quads": [
            ((220, 60, 55), (255, 255, 255)),    # BL red
            ((230, 155, 40), (255, 255, 255)),   # BR amber
            ((210, 135, 30), (255, 255, 255)),   # TL amber-dark
            ((55, 160, 75), (255, 255, 255)),    # TR green
        ],
    },
    # Navy/gold corporate
    {
        "bg": (18, 28, 55), "grid": (80, 100, 150), "axis": (200, 210, 230), "low": (100, 120, 160),
        "quads": [
            ((28, 40, 75), (120, 140, 190)),     # BL dark navy/muted
            ((35, 55, 100), (160, 185, 230)),    # BR medium navy
            ((40, 65, 115), (180, 205, 245)),    # TL brighter navy
            ((180, 135, 30), (255, 245, 200)),   # TR gold
        ],
    },
    # Bold teal/coral
    {
        "bg": (245, 248, 250), "grid": (80, 100, 110), "axis": (30, 50, 60), "low": (140, 160, 168),
        "quads": [
            ((220, 100, 90), (255, 255, 255)),   # BL coral
            ((240, 175, 80), (60, 40, 10)),      # BR warm yellow (dark text)
            ((90, 175, 195), (255, 255, 255)),   # TL teal
            ((30, 130, 115), (255, 255, 255)),   # TR deep teal
        ],
    },
    # Slate/purple
    {
        "bg": (38, 32, 58), "grid": (110, 95, 150), "axis": (210, 200, 235), "low": (120, 110, 155),
        "quads": [
            ((55, 48, 82), (140, 130, 175)),     # BL darkest purple
            ((80, 60, 120), (185, 165, 225)),    # BR mid purple
            ((100, 75, 155), (215, 195, 255)),   # TL bright purple
            ((195, 145, 50), (255, 245, 200)),   # TR gold highlight
        ],
    },
    # Clean B&W with accent
    {
        "bg": (255, 255, 255), "grid": (60, 60, 60), "axis": (30, 30, 30), "low": (140, 140, 140),
        "quads": [
            ((240, 240, 240), (100, 100, 100)),  # BL light grey
            ((210, 228, 245), (30, 60, 100)),    # BR pale blue
            ((220, 238, 220), (30, 80, 40)),     # TL pale green
            ((20, 60, 120), (255, 255, 255)),    # TR bold navy
        ],
    },
    # Warm earth tones
    {
        "bg": (252, 248, 240), "grid": (140, 110, 70), "axis": (60, 40, 20), "low": (180, 155, 120),
        "quads": [
            ((235, 215, 185), (100, 70, 30)),    # BL sand
            ((210, 175, 110), (80, 50, 15)),     # BR tan
            ((175, 145, 85), (255, 248, 230)),   # TL bronze
            ((120, 80, 30), (255, 245, 210)),    # TR dark bronze
        ],
    },
    # High contrast dark
    {
        "bg": (15, 15, 20), "grid": (60, 65, 80), "axis": (200, 205, 220), "low": (80, 85, 100),
        "quads": [
            ((30, 30, 38), (90, 95, 115)),       # BL near-black
            ((40, 55, 75), (140, 170, 210)),     # BR dark blue
            ((35, 65, 50), (140, 210, 165)),     # TL dark green
            ((180, 140, 25), (255, 245, 180)),   # TR gold
        ],
    },
]


def _text_fits(font, text, max_w):
    """Check if text fits within max_w pixels."""
    bb = font.getbbox(text)
    return (bb[2] - bb[0]) <= max_w


def _draw_centered_text(draw, cx, cy, text, font, color, max_w, line_gap=5):
    """Draw multi-line centered text, wrapping to fit max_w. Returns total height drawn."""
    # Estimate chars per line from actual pixel width of 'M'
    m_w = font.getbbox("M")[2]
    chars_per_line = max(6, int(max_w / max(1, m_w)))
    lines = wrap(text, width=chars_per_line)
    if not lines:
        lines = [text[:chars_per_line]]

    line_heights = []
    for line in lines:
        bb = font.getbbox(line)
        line_heights.append(bb[3] - bb[1])

    total_h = sum(line_heights) + line_gap * (len(lines) - 1)
    y = cy - total_h // 2

    for line, lh in zip(lines, line_heights):
        bb = font.getbbox(line)
        lw = bb[2] - bb[0]
        draw.text((cx - lw // 2, y), line, font=font, fill=color)
        y += lh + line_gap

    return total_h


def generate_matrix(seed: str, width: int = 800, height: int = 450) -> "Image.Image":
    from PIL import Image, ImageDraw, ImageFont

    rng = _rng(seed)
    tmpl = rng.choice(_MATRIX_TEMPLATES)
    x_label = _fill(tmpl[0], rng)
    y_label = _fill(tmpl[1], rng)
    quadrant_labels = [_fill(q, rng) for q in tmpl[2]]  # BL, BR, TL, TR

    pal = rng.choice(_MATRIX_PALETTES)

    img = Image.new("RGB", (width, height), pal["bg"])
    draw = ImageDraw.Draw(img)

    # Layout constants
    pad_outer = 14        # outer margin
    axis_label_w = 20     # width reserved for rotated y-axis label text
    axis_label_h = 22     # height reserved for x-axis label text
    low_high_h = 16       # height for Low/High tick labels

    grid_x = pad_outer + axis_label_w + 8
    grid_y = pad_outer
    grid_w = width - grid_x - pad_outer
    grid_h = height - grid_y - axis_label_h - low_high_h - pad_outer
    mid_x = grid_x + grid_w // 2
    mid_y = grid_y + grid_h // 2

    # Quadrant fills (BL, BR, TL, TR)
    quads = [
        (grid_x, mid_y, mid_x, grid_y + grid_h),           # BL
        (mid_x, mid_y, grid_x + grid_w, grid_y + grid_h),  # BR
        (grid_x, grid_y, mid_x, mid_y),                    # TL
        (mid_x, grid_y, grid_x + grid_w, mid_y),            # TR
    ]
    for rect, (qfill, _qtxt) in zip(quads, pal["quads"]):
        draw.rectangle(rect, fill=qfill)

    # Grid border + dividers
    grid_lw = 2
    draw.rectangle([grid_x, grid_y, grid_x + grid_w, grid_y + grid_h],
                   outline=pal["grid"], width=grid_lw)
    draw.line([(mid_x, grid_y), (mid_x, grid_y + grid_h)], fill=pal["grid"], width=grid_lw)
    draw.line([(grid_x, mid_y), (grid_x + grid_w, mid_y)], fill=pal["grid"], width=grid_lw)

    # Arrow heads on axes
    arr = 7
    draw.polygon([(mid_x - arr, grid_y + arr + 2), (mid_x + arr, grid_y + arr + 2), (mid_x, grid_y - 2)],
                 fill=pal["grid"])
    draw.polygon([(grid_x + grid_w - 2, mid_y - arr), (grid_x + grid_w - 2, mid_y + arr),
                  (grid_x + grid_w + arr + 2, mid_y)], fill=pal["grid"])

    # Fonts — use bold variants for quad labels
    bold_path = (_find_font("ArialBold") or _find_font("Arial Bold") or
                 _find_font("Arial Narrow Bold") or _find_font("Arial"))
    reg_path = _find_font("Arial") or bold_path

    quad_font_size = max(16, width // 30)   # much larger than before
    axis_font_size = max(13, width // 55)
    low_high_size = max(11, width // 68)

    def _font(path, size):
        if path:
            try:
                return ImageFont.truetype(path, size)
            except Exception:
                pass
        return ImageFont.load_default()

    font_quad = _font(bold_path, quad_font_size)
    font_axis = _font(bold_path, axis_font_size)
    font_low_high = _font(reg_path, low_high_size)

    # Quadrant labels — centered in each quadrant
    quad_centers = [
        ((grid_x + mid_x) // 2,     (mid_y + grid_y + grid_h) // 2),   # BL
        ((mid_x + grid_x + grid_w) // 2, (mid_y + grid_y + grid_h) // 2),  # BR
        ((grid_x + mid_x) // 2,     (grid_y + mid_y) // 2),              # TL
        ((mid_x + grid_x + grid_w) // 2, (grid_y + mid_y) // 2),         # TR
    ]
    quad_w = grid_w // 2
    for qlabel, (cx, cy), (_qfill, qtxt_color) in zip(quadrant_labels, quad_centers, pal["quads"]):
        _draw_centered_text(draw, cx, cy, qlabel, font_quad, qtxt_color,
                            max_w=quad_w - 20, line_gap=6)

    # X-axis label (centered below grid)
    x_lbl_y = grid_y + grid_h + low_high_h + 2
    _draw_centered_text(draw, grid_x + grid_w // 2, x_lbl_y + axis_font_size // 2,
                        x_label, font_axis, pal["axis"], max_w=grid_w - 60)

    # Low / High for x-axis
    lh_y = grid_y + grid_h + 3
    draw.text((grid_x + 4, lh_y), "Low", font=font_low_high, fill=pal["low"])
    bb_hi = font_low_high.getbbox("High")
    draw.text((grid_x + grid_w - (bb_hi[2] - bb_hi[0]) - 4, lh_y),
              "High", font=font_low_high, fill=pal["low"])

    # Y-axis label (rotated, left side) — render then rotate
    y_label_short = y_label[:28]
    bb_yl = font_axis.getbbox(y_label_short)
    yl_w = bb_yl[2] - bb_yl[0]
    yl_h = bb_yl[3] - bb_yl[1]
    yl_img = Image.new("RGBA", (yl_w + 4, yl_h + 4), (0, 0, 0, 0))
    yl_draw = ImageDraw.Draw(yl_img)
    # Convert axis color to RGBA
    ac = pal["axis"] + (255,) if len(pal["axis"]) == 3 else pal["axis"]
    yl_draw.text((2, 2), y_label_short, font=font_axis, fill=ac)
    yl_img = yl_img.rotate(90, expand=True)
    yl_paste_x = pad_outer
    yl_paste_y = grid_y + (grid_h - yl_img.height) // 2
    img.paste(yl_img, (yl_paste_x, yl_paste_y), yl_img)

    # Low / High for y-axis (small, rotated)
    for val, rel_y in [("Low", grid_y + grid_h - low_high_size - 2),
                        ("High", grid_y + 2)]:
        bb_v = font_low_high.getbbox(val)
        v_w = bb_v[2] - bb_v[0]
        v_img = Image.new("RGBA", (v_w + 2, bb_v[3] - bb_v[1] + 2), (0, 0, 0, 0))
        v_draw = ImageDraw.Draw(v_img)
        lc = pal["low"] + (255,) if len(pal["low"]) == 3 else pal["low"]
        v_draw.text((1, 1), val, font=font_low_high, fill=lc)
        v_img = v_img.rotate(90, expand=True)
        img.paste(v_img, (pad_outer, rel_y), v_img)

    return img


# ---------------------------------------------------------------------------
# Batch generation
# ---------------------------------------------------------------------------

FORMATS = {
    "top_bottom": (generate_top_bottom, 40),
    "matrix": (generate_matrix, 40),
}


def main():
    parser = argparse.ArgumentParser(description="Generate meme images for presentations")
    parser.add_argument("--count", type=int, default=None,
                        help="Total memes per format (default: 40 each)")
    parser.add_argument("--format", choices=["top_bottom", "matrix", "all"], default="all")
    parser.add_argument("--force", action="store_true", help="Regenerate existing files")
    parser.add_argument("--dry-run", action="store_true", help="Preview text only, no images")
    args = parser.parse_args()

    if args.dry_run:
        print("--- Dry run: top/bottom memes ---")
        for i in range(5):
            seed = f"meme_top_bottom_{i}"
            rng = _rng(seed)
            t, b = rng.choice(_TOP_BOTTOM_TEMPLATES)
            print(f"  [{i}] TOP: {_fill(t, rng)}")
            rng = _rng(seed)
            rng.choice(_TOP_BOTTOM_TEMPLATES)
            print(f"       BOT: {_fill(b, rng)}")
        print("\n--- Dry run: matrix memes ---")
        for i in range(5):
            seed = f"meme_matrix_{i}"
            rng = _rng(seed)
            tmpl = rng.choice(_MATRIX_TEMPLATES)
            print(f"  [{i}] X={_fill(tmpl[0], rng)} Y={_fill(tmpl[1], rng)}")
        return

    try:
        from PIL import Image
    except ImportError:
        print("ERROR: Pillow not installed. Run: pip install pillow", file=sys.stderr)
        sys.exit(1)

    OUT_DIR.mkdir(parents=True, exist_ok=True)

    formats_to_run = list(FORMATS.items()) if args.format == "all" else [(args.format, FORMATS[args.format])]
    generated = 0
    skipped = 0

    for fmt_name, (generator_fn, default_count) in formats_to_run:
        count = args.count if args.count is not None else default_count
        print(f"\nGenerating {count} '{fmt_name}' memes…")
        for i in range(count):
            seed = f"meme_{fmt_name}_{i}"
            out_path = OUT_DIR / f"{fmt_name}_{i:05d}.webp"
            if out_path.exists() and not args.force:
                skipped += 1
                continue
            try:
                img = generator_fn(seed)
                img.save(str(out_path), "WEBP", quality=88)
                generated += 1
                if generated % 10 == 0:
                    print(f"  {generated} generated…")
            except Exception as e:
                print(f"  ERROR [{seed}]: {e}", file=sys.stderr)

    print(f"\nDone. Generated: {generated}, Skipped (existing): {skipped}")
    print(f"Output: {OUT_DIR}")


if __name__ == "__main__":
    main()
