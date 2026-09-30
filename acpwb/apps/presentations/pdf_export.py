"""
Serialize export_layout slides straight to PDF.

Replaces the WeasyPrint print-template render (~3s CPU per deck: CSS cascade,
font subsetting, WebP -> PDF image re-encoding). Here text uses the standard-14
Helvetica fonts (nothing to embed or subset — wrapping uses their published
metrics), images are JPEGs passed through as-is (DCTDecode), and the only real
work is string formatting.
"""
import zlib
from datetime import datetime, timezone

from .data.helvetica_widths import BOLD as _BOLD_WIDTHS, REGULAR as _REGULAR_WIDTHS
from .export_images import load_jpeg
from .export_layout import (
    EMU_PER_PT, H, INSET_X, INSET_Y, W, Image, Rect, Text, layout_slides, provenance_text,
)

PAGE_W = W / EMU_PER_PT
PAGE_H = H / EMU_PER_PT
LINE_HEIGHT = 1.2       # x font size, matches PowerPoint single spacing
ASCENT = 0.905          # Helvetica ascender, x font size

_FONTS = {
    (False, False): ('F1', 'Helvetica'),
    (True, False): ('F2', 'Helvetica-Bold'),
    (False, True): ('F3', 'Helvetica-Oblique'),
    (True, True): ('F4', 'Helvetica-BoldOblique'),
}


def _pt(emu):
    return emu / EMU_PER_PT


def _rgb(hex_color):
    return ' '.join(f'{int(hex_color[i:i + 2], 16) / 255:.3f}' for i in (0, 2, 4))


def _encode(text):
    return text.encode('cp1252', 'replace')


def _pdf_str(raw):
    return b'(' + raw.replace(b'\\', b'\\\\').replace(b'(', b'\\(').replace(b')', b'\\)') + b')'


def _width(raw, size, bold):
    widths = _BOLD_WIDTHS if bold else _REGULAR_WIDTHS
    return sum(widths[b] for b in raw) * size / 1000


def _wrap(raw, size, bold, max_w):
    words = raw.split(b' ')
    lines, cur = [], b''
    for word in words:
        trial = cur + b' ' + word if cur else word
        if cur and _width(trial, size, bold) > max_w:
            lines.append(cur)
            cur = word
        else:
            cur = trial
    lines.append(cur)
    return lines


def _text_ops(t):
    inner_w = _pt(t.w - 2 * INSET_X)
    x0 = _pt(t.x + INSET_X)
    blocks = []
    total_h = 0.0
    for i, p in enumerate(t.paras):
        raw = _encode(p.text)
        lines = _wrap(raw, p.size, p.bold, inner_w) if t.wrap else [raw]
        before = p.space_before if i else 0
        blocks.append((p, lines, before))
        total_h += before + len(lines) * p.size * LINE_HEIGHT + p.space_after

    top = _pt(t.y + INSET_Y)
    if t.anchor == 'ctr':
        top += max(0.0, (_pt(t.h - 2 * INSET_Y) - total_h) / 2)

    out = [b'BT']
    y = top
    for p, lines, before in blocks:
        y += before
        font = _FONTS[(p.bold, p.italic)][0]
        out.append(f'/{font} {p.size:g} Tf {_rgb(p.color)} rg'.encode())
        for line in lines:
            baseline = y + p.size * (LINE_HEIGHT - 1) / 2 + p.size * ASCENT
            if p.align == 'l':
                lx = x0
            else:
                lw = _width(line, p.size, p.bold)
                lx = x0 + (inner_w - lw) / (2 if p.align == 'ctr' else 1)
            out.append(f'1 0 0 1 {lx:.2f} {PAGE_H - baseline:.2f} Tm '.encode()
                       + _pdf_str(line) + b' Tj')
            y += p.size * LINE_HEIGHT
        y += p.space_after
    out.append(b'ET')
    return b'\n'.join(out)


def _rect_ops(r, alphas):
    x, y, w, h = _pt(r.x), PAGE_H - _pt(r.y + r.h), _pt(r.w), _pt(r.h)
    ops = f'{_rgb(r.fill)} rg {x:.2f} {y:.2f} {w:.2f} {h:.2f} re '
    if r.line:
        ops += f'{_rgb(r.line)} RG 1 w B'
    else:
        ops += 'f'
    if r.alpha < 1.0:
        gs = f'GS{int(r.alpha * 100)}'
        alphas[gs] = r.alpha
        ops = f'q /{gs} gs {ops} Q'
    return ops.encode()


def _image_ops(img, name, iw, ih):
    x, y, w, h = _pt(img.x), PAGE_H - _pt(img.y + img.h), _pt(img.w), _pt(img.h)
    scale = (max if img.fit == 'cover' else min)(w / iw, h / ih)
    dw, dh = iw * scale, ih * scale
    dx, dy = x + (w - dw) / 2, y + (h - dh) / 2
    return (f'q {x:.2f} {y:.2f} {w:.2f} {h:.2f} re W n '
            f'{dw:.2f} 0 0 {dh:.2f} {dx:.2f} {dy:.2f} cm /{name} Do Q').encode()


def _info_str(text):
    """PDF text string; UTF-16BE so non-Latin-1 titles survive."""
    return b'<FEFF' + text.encode('utf-16-be').hex().upper().encode() + b'>'


def generate_pdf_bytes(pres_meta, slides):
    laid_out = layout_slides(pres_meta, slides)
    images = {}         # static path -> (name, jpeg bytes, w, h)
    alphas = {}
    hidden = (b'q BT 3 Tr /F1 1 Tf 1 0 0 1 4 4 Tm '
              + _pdf_str(_encode(provenance_text(pres_meta))) + b' Tj ET Q')

    pages = []
    for slide in laid_out:
        ops = [f'{_rgb(slide.bg)} rg 0 0 {PAGE_W:.2f} {PAGE_H:.2f} re f'.encode()]
        for shape in slide.shapes:
            if isinstance(shape, Rect):
                ops.append(_rect_ops(shape, alphas))
            elif isinstance(shape, Text):
                ops.append(_text_ops(shape))
            elif isinstance(shape, Image):
                if shape.path not in images:
                    jpeg = load_jpeg(shape.path)
                    if jpeg is None:
                        continue
                    images[shape.path] = (f'Im{len(images) + 1}',) + jpeg
                name, _, iw, ih = images[shape.path]
                ops.append(_image_ops(shape, name, iw, ih))
        ops.append(hidden)
        pages.append(zlib.compress(b'\n'.join(ops), 1))

    # Object layout: 1 catalog, 2 pages, 3 resources, 4-7 fonts, 8 info,
    # then images, then (page, content) pairs.
    objs = {}
    for (bold, italic), (fname, base) in _FONTS.items():
        objs[4 + int(fname[1]) - 1] = (f'<< /Type /Font /Subtype /Type1 /BaseFont /{base} '
                                       f'/Encoding /WinAnsiEncoding >>').encode()

    token = pres_meta.get('watermark_token', '')
    authors = pres_meta.get('authors') or []
    created = datetime(pres_meta['year'], pres_meta['month'], min(pres_meta['day'], 28),
                       tzinfo=timezone.utc).strftime("D:%Y%m%d090000Z")
    objs[8] = (b'<< /Title ' + _info_str(pres_meta.get('title', ''))
               + b' /Author ' + _info_str(authors[0].get('full_name', '') if authors
                                          else pres_meta.get('org_name', ''))
               + b' /Subject ' + _info_str(pres_meta.get('subtitle', ''))
               + b' /Keywords ' + _info_str(f"{pres_meta.get('org_name', '')}; {token}")
               + f' /CreationDate ({created}) /ModDate ({created}) >>'.encode())

    num = 9
    xobjects = []
    for name, data, iw, ih in images.values():
        objs[num] = (f'<< /Type /XObject /Subtype /Image /Width {iw} /Height {ih} '
                     f'/ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode '
                     f'/Length {len(data)} >>\nstream\n').encode() + data + b'\nendstream'
        xobjects.append(f'/{name} {num} 0 R')
        num += 1

    kids = []
    for content in pages:
        objs[num] = (f'<< /Type /Page /Parent 2 0 R /MediaBox [0 0 {PAGE_W:.2f} {PAGE_H:.2f}] '
                     f'/Resources 3 0 R /Contents {num + 1} 0 R >>').encode()
        objs[num + 1] = (f'<< /Length {len(content)} /Filter /FlateDecode >>\nstream\n'.encode()
                         + content + b'\nendstream')
        kids.append(f'{num} 0 R')
        num += 2

    objs[1] = b'<< /Type /Catalog /Pages 2 0 R >>'
    objs[2] = f'<< /Type /Pages /Kids [{" ".join(kids)}] /Count {len(kids)} >>'.encode()
    gs = ''.join(f'/{k} << /Type /ExtGState /ca {v:.2f} /CA {v:.2f} >>' for k, v in alphas.items())
    objs[3] = (f'<< /Font << /F1 4 0 R /F2 5 0 R /F3 6 0 R /F4 7 0 R >> '
               f'/ExtGState << {gs} >> /XObject << {" ".join(xobjects)} >> >>').encode()

    out = [b'%PDF-1.4\n%\xe2\xe3\xcf\xd3\n']
    pos = len(out[0])
    offsets = []
    for n in range(1, num):
        chunk = f'{n} 0 obj\n'.encode() + objs[n] + b'\nendobj\n'
        offsets.append(pos)
        out.append(chunk)
        pos += len(chunk)
    xref = [f'xref\n0 {num}\n0000000000 65535 f \n'.encode()]
    xref += [f'{o:010d} 00000 n \n'.encode() for o in offsets]
    out += xref
    out.append(f'trailer\n<< /Size {num} /Root 1 0 R /Info 8 0 R >>\nstartxref\n{pos}\n%%EOF\n'
               .encode())
    return b''.join(out)
