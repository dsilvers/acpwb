"""
Deterministic image picker for pre-generated presentation image libraries.

Backgrounds: bg_{n:05d}.webp, pool of TOTAL_BACKGROUNDS slots (files generated separately).
Memes:       matrix_{n:05d}.webp, pool of TOTAL_MEMES slots.

Picks are computed purely from the seed — no directory scanning, no per-worker inconsistency.
Returns None for any slot whose file hasn't been generated yet; caller degrades gracefully.
"""

import hashlib
import random
from pathlib import Path

_STATIC_ROOT = Path(__file__).resolve().parents[2] / "static"
_BG_DIR      = _STATIC_ROOT / "img" / "presentations" / "backgrounds"
_MEME_DIR    = _STATIC_ROOT / "img" / "presentations" / "memes"

TOTAL_BACKGROUNDS = 5000
TOTAL_MEMES       = 5000


def _rng(seed_str: str) -> random.Random:
    seed_int = int(hashlib.md5(seed_str.encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)


def pick_background(pres_seed: str, slide_num: int) -> str | None:
    """Return a static-relative path for a background image, or None if not yet generated."""
    rng = _rng(f"{pres_seed}_bg_{slide_num}")
    n = rng.randint(0, TOTAL_BACKGROUNDS - 1)
    path = _BG_DIR / f"bg_{n:05d}.webp"
    if not path.exists():
        return None
    return f"img/presentations/backgrounds/bg_{n:05d}.webp"


def pick_meme(pres_seed: str, slide_num: int) -> str | None:
    """Return a static-relative path for a meme image, or None if not yet generated."""
    rng = _rng(f"{pres_seed}_meme_{slide_num}")
    n = rng.randint(0, TOTAL_MEMES - 1)
    path = _MEME_DIR / f"matrix_{n:05d}.webp"
    if not path.exists():
        return None
    return f"img/presentations/memes/matrix_{n:05d}.webp"
