import hashlib
from pathlib import Path
from django import template
from django.templatetags.static import static
from django.utils.safestring import mark_safe

HEADSHOT_DIR = Path(__file__).resolve().parents[3] / "static" / "img" / "headshots"
HEADSHOT_COUNT = 400

register = template.Library()

# Palette of color pairs for CSS gradient avatars
AVATAR_PALETTES = [
    ('#0A1628', '#C9A84C'),
    ('#1a3a5c', '#4a9eda'),
    ('#2d5a27', '#7bc67e'),
    ('#5c1a1a', '#da4a4a'),
    ('#3d2b5c', '#9b6dd0'),
    ('#5c4a1a', '#d4a843'),
    ('#1a4a4a', '#43c5c5'),
    ('#4a2b1a', '#c57843'),
    ('#1a1a5c', '#4343da'),
    ('#4a1a3d', '#d043b5'),
]


@register.simple_tag
def avatar_card(seed, initials, size=80):
    """Render a CSS gradient avatar card with initials."""
    idx = int(hashlib.md5(str(seed).encode()).hexdigest(), 16) % len(AVATAR_PALETTES)
    color1, color2 = AVATAR_PALETTES[idx]
    style = (
        f'width:{size}px;height:{size}px;'
        f'background:linear-gradient(135deg,{color1},{color2});'
        f'border-radius:50%;display:flex;align-items:center;justify-content:center;'
        f'color:#fff;font-weight:700;font-size:{size // 3}px;'
        f'letter-spacing:0.05em;flex-shrink:0;'
    )
    return mark_safe(f'<div style="{style}">{initials}</div>')


@register.simple_tag
def headshot_or_avatar(seed, initials_text, size=80):
    """Use a generated headshot if available, otherwise fall back to CSS gradient avatar."""
    idx = int(hashlib.md5(str(seed).encode()).hexdigest(), 16) % HEADSHOT_COUNT
    img_path = HEADSHOT_DIR / f"{idx:03d}.webp"
    if img_path.exists():
        url = static(f"img/headshots/{idx:03d}.webp")
        style = (
            f'width:{size}px;height:{size}px;border-radius:50%;'
            f'object-fit:cover;object-position:center top;flex-shrink:0;'
        )
        return mark_safe(f'<img src="{url}" alt="{initials_text}" style="{style}">')
    return avatar_card(seed, initials_text, size)


PROJECT_COVER_COUNT = 80


@register.filter
def project_cover_idx(slug):
    """Map a project slug to a deterministic 3-digit cover image index (000–079)."""
    return str(int(hashlib.md5(str(slug).encode()).hexdigest(), 16) % PROJECT_COVER_COUNT).zfill(3)


@register.filter
def initials(name):
    """Return initials from a full name string."""
    parts = name.strip().split()
    if len(parts) >= 2:
        return f"{parts[0][0]}{parts[-1][0]}".upper()
    return name[:2].upper() if name else "??"
