"""
Format-neutral slide layout for the PPTX and PDF downloads.

Each slide type is laid out once, in EMU on a 13.333in x 7.5in page, as a flat
list of primitives (Rect / Text / Image). pptx_export and pdf_export each
serialize those primitives directly — no python-pptx object model and no HTML
layout engine per request, since these endpoints are crawled at volume over an
unbounded URL space (every presentation URL is unique, so caching can't help;
only cheap generation does).
"""
from dataclasses import dataclass, field

EMU_PER_INCH = 914400
EMU_PER_PT = 12700


def inches(v):
    return int(v * EMU_PER_INCH)


W = inches(13.333)
H = inches(7.5)

# Default text-box insets (match PowerPoint's own defaults so both renderers
# wrap text at the same width).
INSET_X = 91440
INSET_Y = 45720

# Brand colors
NAVY = '0A1628'
NAVY_MID = '122040'
NAVY_LIGHT = '1E3560'
GOLD = 'C9A84C'
GOLD_LIGHT = 'E0C06E'
WHITE = 'FFFFFF'
LIGHT_GRAY = 'F4F6F9'
MID_GRAY = 'E4E8EF'
DARK_TEXT = '222233'
MID_TEXT = '555566'


@dataclass
class Para:
    text: str
    size: float
    bold: bool = False
    italic: bool = False
    color: str = WHITE
    align: str = 'l'          # 'l' | 'ctr' | 'r'
    space_before: float = 0   # pt
    space_after: float = 0    # pt


@dataclass
class Rect:
    x: int
    y: int
    w: int
    h: int
    fill: str
    line: str | None = None
    alpha: float = 1.0


@dataclass
class Text:
    x: int
    y: int
    w: int
    h: int
    paras: list
    wrap: bool = True
    anchor: str = 't'         # 't' | 'ctr'


@dataclass
class Image:
    x: int
    y: int
    w: int
    h: int
    path: str                 # static-relative, e.g. img/presentations/...
    fit: str = 'cover'        # 'cover' | 'contain'


@dataclass
class Slide:
    bg: str
    shapes: list = field(default_factory=list)
    notes: str = ''

    def rect(self, x, y, w, h, fill, line=None, alpha=1.0):
        self.shapes.append(Rect(int(x), int(y), int(w), int(h), fill, line, alpha))

    def text(self, x, y, w, h, *paras, wrap=True, anchor='t'):
        paras = [p for p in paras if p.text]
        if paras:
            self.shapes.append(Text(int(x), int(y), int(w), int(h), paras, wrap, anchor))

    def image(self, x, y, w, h, path, fit='cover'):
        if path:
            self.shapes.append(Image(int(x), int(y), int(w), int(h), path, fit))


# ── shared chrome ──────────────────────────────────────────────────────────

def _gold_bar(s, top, width=W, height=inches(0.045), left=0):
    s.rect(left, top, width, height, GOLD)


def _header_band(s, title, height_frac=0.22):
    """Dark navy header band with gold accent + white title text."""
    band_h = int(H * height_frac)
    s.rect(0, 0, W, band_h, NAVY)
    _gold_bar(s, top=band_h - inches(0.045))
    margin = inches(0.55)
    s.text(margin, inches(0.12), W - margin * 2, band_h - inches(0.2),
           Para(title, 22, bold=True), anchor='ctr')
    return band_h


def _footer(s, meta, slide):
    footer_h = inches(0.32)
    top = H - footer_h
    s.rect(0, top, W, footer_h, NAVY_MID)
    s.text(inches(0.35), top + inches(0.02), W - inches(0.7), footer_h,
           Para(meta.get('org_name', 'ACPWB'), 8, bold=True, color=GOLD), wrap=False)
    s.text(W - inches(1.5), top + inches(0.02), inches(1.2), footer_h,
           Para(f"Slide {slide['num']} / {slide['total']}", 8, color=MID_GRAY, align='r'),
           wrap=False)


def _footnote(s, slide):
    fn = slide.get('footnote')
    if fn:
        s.text(inches(0.55), H - inches(0.62), W - inches(1.1), inches(0.3),
               Para(fn, 8, italic=True, color=MID_TEXT))


def _bullets(items, size, bold=False, space_before=4, space_after=0):
    return [Para(f"•  {b}", size, bold=bold, color=DARK_TEXT,
                 space_before=space_before, space_after=space_after) for b in items]


def _photo_backdrop(s, path):
    """Full-bleed photo under a dark scrim, for the title / Q&A slides."""
    if path:
        s.image(0, 0, W, H, path)
        s.rect(0, 0, W, H, '000000', alpha=0.72)


# ── slide types ────────────────────────────────────────────────────────────

def _title(meta, slide):
    s = Slide(NAVY)
    _photo_backdrop(s, slide.get('bg_image'))
    _gold_bar(s, top=inches(1.5), width=inches(0.9), left=inches(0.65), height=inches(0.055))
    s.text(inches(0.65), inches(1.65), inches(10.5), inches(2.4),
           Para(slide.get('heading', meta.get('title', '')), 34, bold=True))
    s.text(inches(0.65), inches(4.1), inches(10.5), inches(1.2),
           Para(slide.get('subheading', meta.get('subtitle', '')), 16, color=GOLD_LIGHT))
    authors = slide.get('authors') or []
    if authors:
        a = authors[0]
        s.text(inches(0.65), inches(5.35), inches(9.5), inches(1.1),
               Para(a.get('full_name', ''), 13, bold=True),
               Para(a.get('title', ''), 11, color=GOLD_LIGHT))
    s.text(inches(0.65), inches(6.6), inches(7.5), inches(0.6),
           Para(meta.get('org_name', '').upper(), 11, bold=True, color=GOLD))
    date_str = f"{meta.get('pub_date_display', meta.get('year', ''))} · {meta.get('industry', '')}"
    s.text(W - inches(4.5), inches(6.6), inches(4.2), inches(0.6),
           Para(date_str, 10, color=MID_GRAY, align='r'))
    return s


def _agenda(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', 'Agenda'))
    items = slide.get('items', [])
    s.text(inches(0.75), band_h + inches(0.35), W - inches(1.5), H - band_h - inches(0.7),
           *[Para(f"{i}.  {item}", 14, color=DARK_TEXT, space_before=6)
             for i, item in enumerate(items, 1)])
    _footer(s, meta, slide)
    return s


def _content(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    s.text(inches(0.75), band_h + inches(0.35), W - inches(1.5), H - band_h - inches(0.75),
           *_bullets(slide.get('bullets', []), 13, space_after=2))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _stat(meta, slide):
    s = Slide(NAVY)
    _gold_bar(s, top=0, height=inches(0.05))
    s.text(inches(0.65), inches(0.25), W - inches(1.3), inches(0.7),
           Para(slide.get('heading', ''), 16, bold=True))
    stats = slide.get('stats', [])
    col_w = (W - inches(1.0)) // max(len(stats), 1)
    for i, stat in enumerate(stats):
        left = inches(0.5) + i * col_w
        s.rect(left + inches(0.1), inches(1.2), col_w - inches(0.2), inches(4.8), NAVY_MID)
        s.text(left + inches(0.2), inches(1.9), col_w - inches(0.4), inches(1.6),
               Para(stat.get('value', ''), 36, bold=True, color=GOLD, align='ctr'))
        s.text(left + inches(0.2), inches(3.8), col_w - inches(0.4), inches(1.8),
               Para(stat.get('label', ''), 11, color=LIGHT_GRAY, align='ctr'))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _quote(meta, slide):
    s = Slide(NAVY_LIGHT)
    _gold_bar(s, top=0, height=inches(0.06))
    s.text(inches(1.2), inches(1.2), W - inches(2.4), inches(4.2),
           Para(f"“{slide.get('quote', '')}”", 20, italic=True, align='ctr'), anchor='ctr')
    s.text(inches(1.2), inches(5.5), W - inches(2.4), inches(0.8),
           Para(f"— {slide.get('attribution', '')}", 13, color=GOLD, align='ctr'))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _chart(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    chart_type = slide.get('chart_type', '')
    top = band_h + inches(0.45)
    avail_h = H - top - inches(1.1)

    if chart_type in ('bar_h', 'bar_v'):
        bars = slide.get('chart_bars', [])
        row_h = min(inches(0.5), avail_h // max(len(bars), 1))
        bar_left = inches(4.6)
        bar_max_w = W - bar_left - inches(1.6)
        for i, bar in enumerate(bars):
            y = top + i * row_h
            s.text(inches(0.75), y, inches(3.7), row_h,
                   Para(bar.get('label', ''), 10, color=DARK_TEXT, align='r'), anchor='ctr')
            bw = max(inches(0.05), bar_max_w * bar.get('pct', 0) // 100)
            s.rect(bar_left, y + row_h // 6, bw, row_h * 2 // 3,
                   (bar.get('color') or '#' + GOLD).lstrip('#').upper())
            s.text(bar_left + bw + inches(0.08), y, inches(1.2), row_h,
                   Para(str(bar.get('value', '')), 10, bold=True, color=DARK_TEXT),
                   wrap=False, anchor='ctr')
    elif chart_type == 'line':
        pts = slide.get('chart_line_pts', [])
        n = max(len(pts), 1)
        col_w = (W - inches(1.5)) // n
        base_y = top + avail_h - inches(0.4)
        plot_h = avail_h - inches(0.8)
        for i, pt in enumerate(pts):
            x = inches(0.75) + i * col_w
            v = pt.get('value', 0) or 0
            bh = max(inches(0.04), plot_h * v // 100)
            s.rect(x + col_w // 4, base_y - bh, col_w // 2, bh, NAVY_LIGHT)
            s.text(x, base_y - bh - inches(0.35), col_w, inches(0.3),
                   Para(str(v), 9, bold=True, color=DARK_TEXT, align='ctr'))
            s.text(x, base_y + inches(0.05), col_w, inches(0.35),
                   Para(pt.get('label', ''), 8, color=MID_TEXT, align='ctr'))
    elif chart_type == 'donut':
        arcs = slide.get('chart_arcs', [])
        row_h = min(inches(0.5), avail_h // max(len(arcs), 1))
        for i, arc in enumerate(arcs):
            y = top + i * row_h
            s.rect(inches(0.75), y + row_h // 4, row_h // 2, row_h // 2,
                   (arc.get('color') or '#' + GOLD).lstrip('#').upper())
            s.text(inches(0.75) + row_h, y, inches(6.5), row_h,
                   Para(f"{arc.get('label', '')}  —  {arc.get('pct', '')}%", 11, color=DARK_TEXT),
                   anchor='ctr')

    src = slide.get('chart_source', '')
    if src:
        s.text(inches(0.75), H - inches(0.95), W - inches(1.5), inches(0.3),
               Para(f"Source: {src}", 8, italic=True, color=MID_TEXT))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _image(meta, slide):
    s = Slide(NAVY)
    s.image(0, 0, W, H, slide.get('image_path'))
    caption = slide.get('caption', '')
    if caption:
        s.rect(0, H - inches(1.5), W, inches(1.18), '000000', alpha=0.6)
        s.text(inches(0.65), H - inches(1.5), W - inches(1.3), inches(1.18),
               Para(caption, 14, align='ctr'), anchor='ctr')
    _footer(s, meta, slide)
    return s


def _meme(meta, slide):
    s = Slide(NAVY_MID)
    path = slide.get('meme_path')
    if path:
        s.image(inches(0.4), inches(0.3), W - inches(0.8), H - inches(0.9), path, fit='contain')
    _footer(s, meta, slide)
    return s


def _summary(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', 'Key Takeaways'))
    s.text(inches(0.75), band_h + inches(0.4), W - inches(1.5), H - band_h - inches(0.8),
           *_bullets(slide.get('bullets', []), 13, bold=True, space_before=5))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _two_column(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    mid_x = W // 2
    content_top = band_h + inches(0.2)
    content_h = H - band_h - inches(0.5)
    col_w = mid_x - inches(0.75)
    s.rect(mid_x - inches(0.01), content_top, inches(0.02), content_h, MID_GRAY)
    for left_x, label_key, items_key in [
        (inches(0.55), 'left_label', 'left_items'),
        (mid_x + inches(0.2), 'right_label', 'right_items'),
    ]:
        s.text(left_x, content_top + inches(0.15), col_w, inches(0.45),
               Para(slide.get(label_key, ''), 12, bold=True, color=GOLD))
        s.text(left_x, content_top + inches(0.7), col_w, content_h - inches(0.8),
               *_bullets(slide.get(items_key, []), 11, space_before=3))
    _footnote(s, slide)
    _footer(s, meta, slide)
    return s


def _timeline(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    milestones = slide.get('milestones', [])
    n = max(len(milestones), 1)
    avail_w = W - inches(1.0)
    col_w = avail_w // n
    box_top = band_h + inches(0.45)
    box_h = H - band_h - inches(0.9)
    line_y = box_top + inches(0.45)
    s.rect(inches(0.5) + col_w // 2, line_y, avail_w - col_w, inches(0.04), GOLD)
    for i, ms in enumerate(milestones):
        left = inches(0.5) + i * col_w
        s.rect(left + col_w // 2 - inches(0.12), line_y - inches(0.1), inches(0.24), inches(0.24),
               NAVY, line=GOLD)
        s.text(left + inches(0.1), box_top - inches(0.1), col_w - inches(0.2), inches(0.38),
               Para(ms.get('label', ''), 9, bold=True, color=NAVY, align='ctr'))
        s.text(left + inches(0.1), line_y + inches(0.3), col_w - inches(0.2), box_h - inches(0.8),
               Para(ms.get('desc', ''), 9, color=MID_TEXT, align='ctr'))
    _footer(s, meta, slide)
    return s


def _section_divider(meta, slide):
    s = Slide(NAVY)
    s.text(inches(0.65), inches(2.55), inches(3), inches(0.5),
           Para(f"Section {slide.get('section_num', '')}", 11, bold=True, color=GOLD))
    _gold_bar(s, top=inches(3.1), width=inches(1.2), left=inches(0.65), height=inches(0.065))
    s.text(inches(0.65), inches(3.25), W - inches(1.3), inches(2.5),
           Para(slide.get('heading', ''), 32, bold=True))
    _footer(s, meta, slide)
    return s


def _process(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    steps = slide.get('steps', [])
    step_w = (W - inches(1.0)) // max(len(steps), 1)
    box_top = band_h + inches(0.4)
    box_h = H - band_h - inches(0.85)
    for i, step in enumerate(steps):
        left = inches(0.5) + i * step_w
        s.rect(left + inches(0.08), box_top, step_w - inches(0.16), box_h,
               NAVY if i % 2 == 0 else NAVY_MID)
        s.text(left + inches(0.15), box_top + inches(0.12), step_w - inches(0.3), inches(0.5),
               Para(str(i + 1), 22, bold=True, color=GOLD, align='ctr'))
        s.text(left + inches(0.1), box_top + inches(0.7), step_w - inches(0.2), inches(0.55),
               Para(step.get('name', ''), 10, bold=True, align='ctr'))
        s.text(left + inches(0.1), box_top + inches(1.3), step_w - inches(0.2), box_h - inches(1.4),
               Para(step.get('desc', ''), 9, color=LIGHT_GRAY, align='ctr'))
    _footer(s, meta, slide)
    return s


def _case_study(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''))
    s.text(inches(0.65), band_h + inches(0.15), W - inches(1.3), inches(0.35),
           Para(slide.get('org_type', ''), 10, italic=True, color=MID_TEXT))
    section_w = (W - inches(1.0)) // 3
    for i, (label, content, bg) in enumerate([
        ('Challenge', slide.get('challenge', ''), NAVY),
        ('Approach', slide.get('approach', ''), NAVY_MID),
        ('Result', slide.get('result', ''), NAVY_LIGHT),
    ]):
        left = inches(0.5) + i * section_w
        s.rect(left + inches(0.05), band_h + inches(0.65), section_w - inches(0.1),
               H - band_h - inches(1.1), bg)
        s.text(left + inches(0.15), band_h + inches(0.75), section_w - inches(0.3), inches(0.4),
               Para(label.upper(), 9, bold=True, color=GOLD))
        s.text(left + inches(0.15), band_h + inches(1.2), section_w - inches(0.3),
               H - band_h - inches(1.7),
               Para(content, 10, color=LIGHT_GRAY))
    _footer(s, meta, slide)
    return s


def _callout(meta, slide):
    s = Slide(NAVY)
    _gold_bar(s, top=0, height=inches(0.07))
    s.text(inches(0.8), inches(1.2), W - inches(1.6), inches(2.8),
           Para(slide.get('stat', ''), 72, bold=True, color=GOLD, align='ctr'), anchor='ctr')
    s.text(inches(1.5), inches(4.2), W - inches(3.0), inches(2.2),
           Para(slide.get('description', ''), 17, align='ctr'))
    _footer(s, meta, slide)
    return s


def _appendix(meta, slide):
    s = Slide(LIGHT_GRAY)
    band_h = _header_band(s, slide.get('heading', ''), height_frac=0.18)
    s.text(inches(0.65), band_h + inches(0.3), W - inches(1.3), H - band_h - inches(0.75),
           Para(slide.get('content', ''), 11, color=DARK_TEXT))
    _footer(s, meta, slide)
    return s


def _qanda(meta, slide):
    s = Slide(NAVY)
    _photo_backdrop(s, slide.get('bg_image'))
    _gold_bar(s, top=0, height=inches(0.07))
    s.text(inches(1.0), inches(1.6), W - inches(2.0), inches(1.4),
           Para(slide.get('heading', 'Questions & Discussion'), 40, bold=True, align='ctr'),
           anchor='ctr')
    paras = [Para(slide.get('org_name', meta.get('org_name', '')), 14, bold=True,
                  color=GOLD, align='ctr', space_after=6)]
    for a in meta.get('authors') or []:
        paras.append(Para(f"{a.get('full_name', '')} — {a.get('title', '')}", 12,
                          color=LIGHT_GRAY, align='ctr', space_before=4))
        paras.append(Para(a.get('email', ''), 11, color=GOLD_LIGHT, align='ctr'))
    if not meta.get('authors') and slide.get('contact_email'):
        paras.append(Para(slide['contact_email'], 12, color=LIGHT_GRAY, align='ctr'))
    s.text(inches(1.5), inches(3.3), W - inches(3.0), inches(2.6), *paras)
    s.text(inches(1.5), H - inches(0.85), W - inches(3.0), inches(0.35),
           Para(f"Presentation ID: {meta.get('watermark_token', '')}", 8,
                color=MID_GRAY, align='ctr'))
    _footer(s, meta, slide)
    return s


_BUILDERS = {
    'title': _title,
    'agenda': _agenda,
    'content': _content,
    'stat': _stat,
    'quote': _quote,
    'chart': _chart,
    'image': _image,
    'meme': _meme,
    'summary': _summary,
    'two_column': _two_column,
    'timeline': _timeline,
    'section_divider': _section_divider,
    'process': _process,
    'case_study': _case_study,
    'callout': _callout,
    'appendix': _appendix,
    'qanda': _qanda,
}


def layout_slides(pres_meta, slides):
    out = []
    for slide in slides:
        s = _BUILDERS.get(slide.get('type'), _content)(pres_meta, slide)
        s.notes = slide.get('speaker_note', '') or ''
        out.append(s)
    return out


def provenance_text(pres_meta):
    token = pres_meta.get('watermark_token', '')
    return (f"ACPWB Presentation {token} — American Corporation for Public Well Being. "
            f"Content provenance token: {token}.")
