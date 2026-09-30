"""
JPEG copies of the presentation image library for the PPTX/PDF exports.

The library is WebP, which neither format can embed natively, so embedding
one means a decode + re-encode (the dominant cost of the old WeasyPrint PDF
path). JPEG embeds untouched in both (DCTDecode in PDF, image/jpeg in PPTX),
and the pool is fixed (image_selector's TOTAL_BACKGROUNDS + TOTAL_MEMES), so
each image is converted at most once and cached on disk.
"""
import os
import tempfile
from functools import lru_cache
from io import BytesIO
from pathlib import Path

from .image_selector import _STATIC_ROOT

CACHE_DIR = Path(os.environ.get(
    'PRESENTATION_JPEG_CACHE_DIR',
    Path(tempfile.gettempdir()) / 'acpwb-presentation-jpeg',
))
JPEG_QUALITY = 72


def _convert(src, dest):
    from PIL import Image as PILImage
    with PILImage.open(src) as im:
        im = im.convert('RGB')
        buf = BytesIO()
        im.save(buf, 'JPEG', quality=JPEG_QUALITY)
    dest.parent.mkdir(parents=True, exist_ok=True)
    # Write-then-rename so a concurrent reader never sees a partial file.
    fd, tmp = tempfile.mkstemp(dir=dest.parent, suffix='.tmp')
    with os.fdopen(fd, 'wb') as f:
        f.write(buf.getvalue())
    os.replace(tmp, dest)
    return buf.getvalue()


def _jpeg_size(data):
    """(width, height) from a baseline/progressive JPEG's SOF marker."""
    i = 2
    n = len(data)
    while i < n:
        if data[i] != 0xFF:
            i += 1
            continue
        marker = data[i + 1]
        if marker in (0xC0, 0xC1, 0xC2):
            h = int.from_bytes(data[i + 5:i + 7], 'big')
            w = int.from_bytes(data[i + 7:i + 9], 'big')
            return w, h
        seg_len = int.from_bytes(data[i + 2:i + 4], 'big')
        i += 2 + seg_len
    raise ValueError('no SOF marker')


@lru_cache(maxsize=64)
def load_jpeg(static_path):
    """
    Return (jpeg_bytes, width, height) for a static-relative image path, or
    None if the source image doesn't exist.
    """
    src = _STATIC_ROOT / static_path
    dest = CACHE_DIR / Path(static_path).with_suffix('.jpg')
    try:
        data = dest.read_bytes()
    except FileNotFoundError:
        if not src.exists():
            return None
        data = _convert(src, dest)
    w, h = _jpeg_size(data)
    return data, w, h
