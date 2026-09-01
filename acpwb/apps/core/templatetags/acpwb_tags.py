import hashlib
import re
from pathlib import Path
from django import template
from django.templatetags.static import static
from django.utils.safestring import mark_safe

HEADSHOT_DIR = Path(__file__).resolve().parents[3] / "static" / "img" / "headshots"
HEADSHOT_COUNT = 400
SPEAKERS_DIR = Path(__file__).resolve().parents[3] / "static" / "img" / "speakers"

# These are build-time generated assets that don't change while the process
# is running, so list the directories once instead of stat'ing per template
# render (headshot_or_avatar/speaker_avatar are called many times per page).
_HEADSHOT_STEMS = frozenset(p.stem for p in HEADSHOT_DIR.glob('*.webp')) if HEADSHOT_DIR.is_dir() else frozenset()
_SPEAKER_STEMS = frozenset(p.stem for p in SPEAKERS_DIR.glob('*.webp')) if SPEAKERS_DIR.is_dir() else frozenset()

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
    stem = f"{idx:03d}"
    if stem in _HEADSHOT_STEMS:
        url = static(f"img/headshots/{stem}.webp")
        style = (
            f'width:{size}px;height:{size}px;border-radius:50%;'
            f'object-fit:cover;object-position:center top;flex-shrink:0;'
        )
        return mark_safe(f'<img src="{url}" alt="{initials_text}" style="{style}">')
    return avatar_card(seed, initials_text, size)


@register.filter
def schedule_speaker_name(value):
    """Extract the speaker name from a 'Name, Organization' schedule field."""
    return value.split(',')[0].strip() if value else ''


@register.simple_tag
def speaker_avatar(name, initials_text, size=80):
    """Show a generated speaker headshot if available, otherwise fall back to gradient avatar."""
    slug = re.sub(r'[^a-z0-9]+', '-', name.lower()).strip('-')
    if slug in _SPEAKER_STEMS:
        url = static(f"img/speakers/{slug}.webp")
        style = (
            f'width:{size}px;height:{size}px;border-radius:50%;'
            f'object-fit:cover;object-position:center top;flex-shrink:0;'
        )
        return mark_safe(f'<img src="{url}" alt="{initials_text}" style="{style}">')
    return avatar_card(name, initials_text, size)


PROJECT_COVER_COUNT = 80


@register.filter
def project_cover_idx(slug):
    """Map a project slug to a deterministic 3-digit cover image index (000–079)."""
    return str(int(hashlib.md5(str(slug).encode()).hexdigest(), 16) % PROJECT_COVER_COUNT).zfill(3)


@register.filter
def initials(name):
    """Return initials from a full name string."""
    parts = [w for w in name.strip().split() if w and w[0].isalpha()]
    if len(parts) >= 2:
        return f"{parts[0][0]}{parts[-1][0]}".upper()
    if parts:
        return parts[0][:2].upper()
    return "??"


@register.simple_tag
def org_logo(org_slug, size=40):
    """Render a deterministic inline SVG logo for a fake consulting org."""
    from apps.presentations.logo_generator import generate_org_logo
    return mark_safe(generate_org_logo(org_slug, int(size)))


@register.filter
def times_range(value):
    """Return range(1, value+1) for use in for loops: {% for n in count|times_range %}"""
    try:
        return range(1, int(value) + 1)
    except (TypeError, ValueError):
        return range(0)
