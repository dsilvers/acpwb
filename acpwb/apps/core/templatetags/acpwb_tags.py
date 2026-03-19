import hashlib
from django import template

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
    return f'<div style="{style}">{initials}</div>'


@register.filter
def initials(name):
    """Return initials from a full name string."""
    parts = name.strip().split()
    if len(parts) >= 2:
        return f"{parts[0][0]}{parts[-1][0]}".upper()
    return name[:2].upper() if name else "??"
