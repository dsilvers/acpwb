import hashlib
import random


def _rng_from_seed(seed_str):
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


_COLORS = [
    # (bg, text)
    ("#1a2e4a", "#c8a84b"),  # navy / gold
    ("#2c5f2e", "#f0e6c8"),  # forest / cream
    ("#6b2737", "#c8a84b"),  # burgundy / gold
    ("#3d5a6e", "#e8c87a"),  # slate / amber
    ("#1a5c5a", "#c8a84b"),  # teal / gold
    ("#4a2060", "#c8d4e8"),  # purple / lavender
    ("#2d2d2d", "#e8a040"),  # charcoal / amber
    ("#1c3a5e", "#e0c882"),  # cobalt / sand
    ("#7c3238", "#e8d4b0"),  # crimson / bone
    ("#1e4d3a", "#d4b896"),  # hunter / tan
    ("#3a2a1a", "#c8b480"),  # espresso / wheat
    ("#263850", "#e0d0a8"),  # midnight / parchment
    ("#4a1c2a", "#d4c4a0"),  # plum / buff
    ("#1a3a2e", "#d0c890"),  # emerald / straw
    ("#5a3020", "#e8d0a8"),  # brown / cream
    ("#1e2a4a", "#90b8d8"),  # oxford / powder
    ("#2a4a3a", "#c8d890"),  # sage-dark / lime
    ("#3a3a1a", "#d8d090"),  # olive-dark / khaki
    ("#4a2a4a", "#d0c0e0"),  # eggplant / lilac
    ("#1a3a4a", "#88d0d0"),  # prussian / aqua
    ("#3a1a1a", "#e0b8a0"),  # mahogany / peach
    ("#1e3a1e", "#a0d890"),  # dark-green / mint
    ("#2a2a4a", "#a8b8e8"),  # indigo / periwinkle
    ("#4a3a1a", "#e8d0a0"),  # bronze-dark / buff
]

_SHAPES = ["circle", "roundrect", "hexagon", "diamond", "shield", "pill", "octagon"]

_MARKS = [
    # (id, path_d) — all normalized to 0 0 20 20 viewBox
    ("arrow_up", "M10 3 L17 13 H13 V18 H7 V13 H3 Z"),
    ("bar_chart", "M2 18 V10 H6 V18 Z M8 18 V6 H12 V18 Z M14 18 V13 H18 V18 Z"),
    ("network", "M10 3 A2 2 0 1 1 10 7 A2 2 0 1 1 10 3 Z M3 14 A2 2 0 1 1 3 18 A2 2 0 1 1 3 14 Z M17 14 A2 2 0 1 1 17 18 A2 2 0 1 1 17 14 Z M10 7 L3 14 M10 7 L17 14"),
    ("shield_check", "M10 2 L18 6 V12 C18 16 14 19 10 20 C6 19 2 16 2 12 V6 Z M6 11 L9 14 L14 9"),
    ("diamond_split", "M10 2 L18 10 L10 18 L2 10 Z M10 2 L10 18 M2 10 L18 10"),
    ("compass", "M10 2 A8 8 0 1 1 10 18 A8 8 0 1 1 10 2 M10 2 V5 M10 15 V18 M2 10 H5 M15 10 H18 M10 10 L14 6"),
    ("layers", "M10 2 L18 6 L10 10 L2 6 Z M2 10 L10 14 L18 10 M2 14 L10 18 L18 14"),
    ("target", "M10 2 A8 8 0 1 1 10 18 A8 8 0 1 1 10 2 M10 5 A5 5 0 1 1 10 15 A5 5 0 1 1 10 5 M10 8 A2 2 0 1 1 10 12 A2 2 0 1 1 10 8"),
    ("lightning", "M12 2 L6 11 H10 L8 18 L16 9 H12 Z"),
    ("grid", "M2 2 H8 V8 H2 Z M12 2 H18 V8 H12 Z M2 12 H8 V18 H2 Z M12 12 H18 V18 H12 Z"),
    ("crown", "M2 16 L2 10 L6 14 L10 4 L14 14 L18 10 L18 16 Z"),
    ("star", "M10 2 L12 8 H18 L13 12 L15 18 L10 14 L5 18 L7 12 L2 8 H8 Z"),
]


def _get_initials(org_name):
    words = [w for w in org_name.replace('-', ' ').split() if w]
    stop = {"the", "of", "and", "for", "a", "an", "&"}
    significant = [w for w in words if w.lower() not in stop]
    if not significant:
        significant = words
    if len(significant) == 1:
        return significant[0][:2].upper()
    return (significant[0][0] + significant[1][0]).upper()


def _shape_clip_path(shape, size, corner_r, uid):
    cx, cy, r = size / 2, size / 2, size / 2
    if shape == "circle":
        return f'<clipPath id="c{uid}"><circle cx="{cx}" cy="{cy}" r="{r}"/></clipPath>', f'<circle cx="{cx}" cy="{cy}" r="{r}"/>'
    if shape == "roundrect":
        return (f'<clipPath id="c{uid}"><rect width="{size}" height="{size}" rx="{corner_r}"/></clipPath>',
                f'<rect width="{size}" height="{size}" rx="{corner_r}"/>')
    if shape == "pill":
        pr = size // 3
        return (f'<clipPath id="c{uid}"><rect width="{size}" height="{size}" rx="{pr}"/></clipPath>',
                f'<rect width="{size}" height="{size}" rx="{pr}"/>')
    if shape == "diamond":
        pts = f"{cx},{2} {size-2},{cy} {cx},{size-2} {2},{cy}"
        return (f'<clipPath id="c{uid}"><polygon points="{pts}"/></clipPath>',
                f'<polygon points="{pts}"/>')
    if shape == "hexagon":
        h = size * 0.866
        yo = (size - h) / 2
        pts = (f"{cx},{yo+1} {size-3},{yo + h*0.25} {size-3},{yo + h*0.75} "
               f"{cx},{yo+h-1} {3},{yo + h*0.75} {3},{yo + h*0.25}")
        return (f'<clipPath id="c{uid}"><polygon points="{pts}"/></clipPath>',
                f'<polygon points="{pts}"/>')
    if shape == "shield":
        pts = f"{cx},{2} {size-3},{size*0.35} {size-3},{size*0.65} {cx},{size-2} {3},{size*0.65} {3},{size*0.35}"
        return (f'<clipPath id="c{uid}"><polygon points="{pts}"/></clipPath>',
                f'<polygon points="{pts}"/>')
    if shape == "octagon":
        o = size * 0.29
        pts = (f"{o},{2} {size-o},{2} {size-2},{o} {size-2},{size-o} "
               f"{size-o},{size-2} {o},{size-2} {2},{size-o} {2},{o}")
        return (f'<clipPath id="c{uid}"><polygon points="{pts}"/></clipPath>',
                f'<polygon points="{pts}"/>')
    # fallback: circle
    return f'<clipPath id="c{uid}"><circle cx="{cx}" cy="{cy}" r="{r}"/></clipPath>', f'<circle cx="{cx}" cy="{cy}" r="{r}"/>'


def generate_org_logo(org_slug, size=40):
    rng = _rng_from_seed(f"logo_{org_slug}")
    uid = org_slug.replace("-", "")[:8]

    bg, fg = rng.choice(_COLORS)
    shape = rng.choice(_SHAPES)
    corner_r = rng.randint(4, size // 4) if shape in ("roundrect",) else 0

    # icon_style: 0 = initials only, 1 = mark + initials, 2 = mark only
    icon_style = rng.randint(0, 2)
    mark_id, mark_path = rng.choice(_MARKS)

    # org name from slug
    org_name = org_slug.replace("-", " ").title()
    initials = _get_initials(org_name)

    clip_def, bg_shape = _shape_clip_path(shape, size, corner_r, uid)

    # Scale font and mark to size
    font_size = round(size * 0.36) if icon_style == 1 else round(size * 0.46)
    text_y = size * 0.68 if icon_style == 1 else size * 0.63

    # Mark scaled to fit upper ~40% of shape
    mark_scale = size * 0.022  # 20-unit path → size*0.44 wide
    mark_tx = size * 0.07
    mark_ty = size * 0.06 if icon_style == 1 else size * 0.28

    inner = ""
    if icon_style in (1, 2):
        inner += (f'<g transform="translate({mark_tx},{mark_ty}) scale({mark_scale})" '
                  f'fill="none" stroke="{fg}" stroke-width="{round(1.4/mark_scale, 2)}" '
                  f'stroke-linecap="round" stroke-linejoin="round">'
                  f'<path d="{mark_path}"/></g>')
    if icon_style in (0, 1):
        inner += (f'<text x="{size/2}" y="{text_y}" text-anchor="middle" '
                  f'font-family="Arial,Helvetica,sans-serif" font-weight="700" '
                  f'font-size="{font_size}" fill="{fg}" letter-spacing="1">{initials}</text>')

    svg = (
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{size}" height="{size}" '
        f'viewBox="0 0 {size} {size}" role="img" aria-label="{org_name} logo">'
        f'<defs>{clip_def}</defs>'
        f'<g clip-path="url(#c{uid})">'
        f'<rect width="{size}" height="{size}" fill="{bg}"/>'
        f'{inner}'
        f'</g>'
        f'</svg>'
    )
    return svg
