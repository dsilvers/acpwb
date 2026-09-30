"""
Serialize export_layout slides to a .pptx as raw PresentationML.

python-pptx is only used once per process, to produce the static package
parts (slide master, layouts, themes, notes master). Per request, slides are
emitted as XML strings and zipped around those cached parts — building them
through python-pptx's object model cost ~0.4s per deck, which a crawler
walking the (unbounded) presentation URL space turned into a DoS.
"""
import re
import zipfile
from datetime import datetime, timezone
from functools import lru_cache
from io import BytesIO
from xml.sax.saxutils import escape

from .export_images import load_jpeg
from .export_layout import H, INSET_X, INSET_Y, W, Image, Rect, Text, layout_slides

_NS = ('xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" '
       'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" '
       'xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"')
_XML_DECL = "<?xml version='1.0' encoding='UTF-8' standalone='yes'?>\n"
_REL_NS = 'http://schemas.openxmlformats.org/package/2006/relationships'
_RT = 'http://schemas.openxmlformats.org/officeDocument/2006/relationships/'
_CT_SLIDE = 'application/vnd.openxmlformats-officedocument.presentationml.slide+xml'
_CT_NOTES = 'application/vnd.openxmlformats-officedocument.presentationml.notesSlide+xml'

_SLIDE_LAYOUT = '../slideLayouts/slideLayout7.xml'   # "Blank" in the default template
_NOTES_MASTER = '../notesMasters/notesMaster1.xml'

# Arial is metric-compatible with the Helvetica the PDF export uses, so text
# wraps the same way in both formats.
_FONT = 'Arial'


@lru_cache(maxsize=1)
def _base_parts():
    """Static package parts from python-pptx's default template (built once)."""
    from pptx import Presentation

    prs = Presentation()
    prs.slide_width = W
    prs.slide_height = H
    layout = prs.slide_layouts[6]
    assert layout.part.partname == '/ppt/slideLayouts/slideLayout7.xml', layout.part.partname
    # A throwaway slide with notes forces python-pptx to add the notes master
    # to the package; the slide itself is stripped out below.
    prs.slides.add_slide(layout).notes_slide
    buf = BytesIO()
    prs.save(buf)

    parts = {}
    with zipfile.ZipFile(buf) as z:
        for name in z.namelist():
            if name.startswith(('ppt/slides/', 'ppt/notesSlides/')) or name == 'docProps/core.xml':
                continue
            parts[name] = z.read(name)

    pres = parts['ppt/presentation.xml'].decode()
    pres = re.sub(r'<p:sldIdLst>.*?</p:sldIdLst>', '{SLD_ID_LST}', pres)
    pres = pres.replace(' type="screen4x3"', '')
    assert '{SLD_ID_LST}' in pres

    rels = parts['ppt/_rels/presentation.xml.rels'].decode()
    rels = re.sub(r'<Relationship [^>]*relationships/slide"[^>]*/>', '', rels)
    rels = rels.replace('</Relationships>', '{SLIDE_RELS}</Relationships>')

    # python-pptx relates the notes master but never lists it in
    # presentation.xml, which the spec requires. PowerPoint and Google Slides
    # tolerate that; Keynote refuses to open any deck with speaker notes.
    nm_rid = re.search(r'Id="(rId\d+)" Type="[^"]*/notesMaster"', rels).group(1)
    pres = pres.replace('</p:sldMasterIdLst>',
                        f'</p:sldMasterIdLst><p:notesMasterIdLst>'
                        f'<p:notesMasterId r:id="{nm_rid}"/></p:notesMasterIdLst>', 1)

    ct = parts['[Content_Types].xml'].decode()
    ct = re.sub(r'<Override PartName="/ppt/(slides|notesSlides)/[^"]+"[^>]*/>', '', ct)
    ct = ct.replace('</Types>', '{OVERRIDES}</Types>')

    templates = {
        'ppt/presentation.xml': pres,
        'ppt/_rels/presentation.xml.rels': rels,
        '[Content_Types].xml': ct,
    }
    static = {k: v for k, v in parts.items() if k not in templates}
    return templates, static


# ── shape XML ───────────────────────────────────────────────────────────────

def _xfrm(x, y, w, h):
    return f'<a:xfrm><a:off x="{x}" y="{y}"/><a:ext cx="{max(w, 0)}" cy="{max(h, 0)}"/></a:xfrm>'


def _fill(color, alpha=1.0):
    if alpha < 1.0:
        return (f'<a:solidFill><a:srgbClr val="{color}"><a:alpha val="{int(alpha * 100000)}"/>'
                f'</a:srgbClr></a:solidFill>')
    return f'<a:solidFill><a:srgbClr val="{color}"/></a:solidFill>'


def _rect_xml(sid, r):
    line = (f'<a:ln w="12700">{_fill(r.line)}</a:ln>' if r.line else '<a:ln><a:noFill/></a:ln>')
    return (f'<p:sp><p:nvSpPr><p:cNvPr id="{sid}" name="Rectangle {sid}"/><p:cNvSpPr/><p:nvPr/>'
            f'</p:nvSpPr><p:spPr>{_xfrm(r.x, r.y, r.w, r.h)}<a:prstGeom prst="rect"><a:avLst/>'
            f'</a:prstGeom>{_fill(r.fill, r.alpha)}{line}</p:spPr></p:sp>')


def _para_xml(p):
    ppr = f'<a:pPr algn="{p.align}">'
    if p.space_before:
        ppr += f'<a:spcBef><a:spcPts val="{int(p.space_before * 100)}"/></a:spcBef>'
    if p.space_after:
        ppr += f'<a:spcAft><a:spcPts val="{int(p.space_after * 100)}"/></a:spcAft>'
    ppr += '</a:pPr>'
    attrs = f'lang="en-US" sz="{int(p.size * 100)}"'
    if p.bold:
        attrs += ' b="1"'
    if p.italic:
        attrs += ' i="1"'
    return (f'<a:p>{ppr}<a:r><a:rPr {attrs} dirty="0">{_fill(p.color)}'
            f'<a:latin typeface="{_FONT}"/><a:cs typeface="{_FONT}"/></a:rPr>'
            f'<a:t>{escape(p.text)}</a:t></a:r></a:p>')


def _text_xml(sid, t):
    wrap = 'square' if t.wrap else 'none'
    body = (f'<a:bodyPr wrap="{wrap}" lIns="{INSET_X}" tIns="{INSET_Y}" rIns="{INSET_X}" '
            f'bIns="{INSET_Y}" anchor="{t.anchor}" rtlCol="0"><a:noAutofit/></a:bodyPr>')
    return (f'<p:sp><p:nvSpPr><p:cNvPr id="{sid}" name="TextBox {sid}"/><p:cNvSpPr txBox="1"/>'
            f'<p:nvPr/></p:nvSpPr><p:spPr>{_xfrm(t.x, t.y, t.w, t.h)}<a:prstGeom prst="rect">'
            f'<a:avLst/></a:prstGeom><a:noFill/></p:spPr><p:txBody>{body}<a:lstStyle/>'
            f'{"".join(_para_xml(p) for p in t.paras)}</p:txBody></p:sp>')


def _pic_xml(sid, img, rid, iw, ih):
    x, y, w, h = img.x, img.y, img.w, img.h
    src_rect = ''
    box_aspect, img_aspect = w / h, iw / ih
    if img.fit == 'contain':
        if img_aspect > box_aspect:
            nh = int(w / img_aspect)
            y, h = y + (h - nh) // 2, nh
        else:
            nw = int(h * img_aspect)
            x, w = x + (w - nw) // 2, nw
    elif img_aspect > box_aspect:
        crop = int((1 - box_aspect / img_aspect) / 2 * 100000)
        src_rect = f'<a:srcRect l="{crop}" r="{crop}"/>'
    elif img_aspect < box_aspect:
        crop = int((1 - img_aspect / box_aspect) / 2 * 100000)
        src_rect = f'<a:srcRect t="{crop}" b="{crop}"/>'
    return (f'<p:pic><p:nvPicPr><p:cNvPr id="{sid}" name="Picture {sid}"/><p:cNvPicPr>'
            f'<a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr><p:blipFill>'
            f'<a:blip r:embed="{rid}"/>{src_rect}<a:stretch><a:fillRect/></a:stretch></p:blipFill>'
            f'<p:spPr>{_xfrm(x, y, w, h)}<a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr>'
            f'</p:pic>')


def _rels_xml(rels):
    body = ''.join(f'<Relationship Id="{rid}" Type="{_RT}{typ}" Target="{target}"/>'
                   for rid, typ, target in rels)
    return f'{_XML_DECL}<Relationships xmlns="{_REL_NS}">{body}</Relationships>'


def _notes_xml(text):
    return (
        f'{_XML_DECL}<p:notes {_NS}><p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/>'
        '<p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/>'
        '<p:sp><p:nvSpPr><p:cNvPr id="2" name="Slide Image Placeholder 1"/><p:cNvSpPr>'
        '<a:spLocks noGrp="1" noRot="1" noChangeAspect="1"/></p:cNvSpPr><p:nvPr>'
        '<p:ph type="sldImg" idx="2"/></p:nvPr></p:nvSpPr><p:spPr/></p:sp>'
        '<p:sp><p:nvSpPr><p:cNvPr id="3" name="Notes Placeholder 2"/><p:cNvSpPr>'
        '<a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="body" idx="3" sz="quarter"/></p:nvPr>'
        '</p:nvSpPr><p:spPr/><p:txBody><a:bodyPr/><a:lstStyle/>'
        f'<a:p><a:r><a:rPr lang="en-US" dirty="0"/><a:t>{escape(text)}</a:t></a:r></a:p>'
        '</p:txBody></p:sp></p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>'
        '</p:notes>'
    )


def _core_xml(meta):
    now = datetime(meta['year'], meta['month'], min(meta['day'], 28),
                   tzinfo=timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    authors = meta.get('authors') or []
    creator = authors[0].get('full_name', '') if authors else meta.get('org_name', '')
    token = meta.get('watermark_token', '')
    return (
        f'{_XML_DECL}<cp:coreProperties '
        'xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" '
        'xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" '
        'xmlns:dcmitype="http://purl.org/dc/dcmitype/" '
        'xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">'
        f'<dc:title>{escape(meta.get("title", ""))}</dc:title>'
        f'<dc:subject>{escape(meta.get("subtitle", ""))}</dc:subject>'
        f'<dc:creator>{escape(creator)}</dc:creator>'
        f'<cp:keywords>{escape(meta.get("industry", ""))}; {escape(meta.get("domain", ""))}; {token}</cp:keywords>'
        f'<dc:description>{escape(meta.get("org_name", ""))} — Presentation ID: {token}</dc:description>'
        f'<cp:lastModifiedBy>{escape(creator)}</cp:lastModifiedBy><cp:revision>3</cp:revision>'
        f'<dcterms:created xsi:type="dcterms:W3CDTF">{now}</dcterms:created>'
        f'<dcterms:modified xsi:type="dcterms:W3CDTF">{now}</dcterms:modified>'
        f'<cp:category>{escape(meta.get("org_name", ""))}</cp:category></cp:coreProperties>'
    )


# ── package ─────────────────────────────────────────────────────────────────

def generate_pptx_bytes(pres_meta, slides):
    templates, static = _base_parts()
    laid_out = layout_slides(pres_meta, slides)

    files = {}
    media = {}          # static path -> (part name, width, height)
    sld_ids, pres_rels, overrides = [], [], []

    for n, slide in enumerate(laid_out, 1):
        rels = [('rId1', 'slideLayout', _SLIDE_LAYOUT)]
        shapes = []
        for sid, shape in enumerate(slide.shapes, 2):
            if isinstance(shape, Rect):
                shapes.append(_rect_xml(sid, shape))
            elif isinstance(shape, Text):
                shapes.append(_text_xml(sid, shape))
            elif isinstance(shape, Image):
                if shape.path not in media:
                    jpeg = load_jpeg(shape.path)
                    if jpeg is None:
                        continue
                    name = f'ppt/media/image{len(media) + 1}.jpeg'
                    files[name] = jpeg[0]
                    media[shape.path] = (name, jpeg[1], jpeg[2])
                name, iw, ih = media[shape.path]
                rid = f'rId{len(rels) + 1}'
                rels.append((rid, 'image', '../media/' + name.rsplit('/', 1)[1]))
                shapes.append(_pic_xml(sid, shape, rid, iw, ih))

        if slide.notes:
            rels.append((f'rId{len(rels) + 1}', 'notesSlide', f'../notesSlides/notesSlide{n}.xml'))
            files[f'ppt/notesSlides/notesSlide{n}.xml'] = _notes_xml(slide.notes)
            files[f'ppt/notesSlides/_rels/notesSlide{n}.xml.rels'] = _rels_xml([
                ('rId1', 'notesMaster', _NOTES_MASTER),
                ('rId2', 'slide', f'../slides/slide{n}.xml'),
            ])
            overrides.append(f'<Override PartName="/ppt/notesSlides/notesSlide{n}.xml" '
                             f'ContentType="{_CT_NOTES}"/>')

        files[f'ppt/slides/slide{n}.xml'] = (
            f'{_XML_DECL}<p:sld {_NS}><p:cSld><p:bg><p:bgPr>{_fill(slide.bg)}<a:effectLst/>'
            '</p:bgPr></p:bg><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/>'
            '<p:nvPr/></p:nvGrpSpPr><p:grpSpPr/>'
            f'{"".join(shapes)}</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/>'
            '</p:clrMapOvr></p:sld>'
        )
        files[f'ppt/slides/_rels/slide{n}.xml.rels'] = _rels_xml(rels)
        overrides.append(f'<Override PartName="/ppt/slides/slide{n}.xml" ContentType="{_CT_SLIDE}"/>')
        sld_ids.append(f'<p:sldId id="{255 + n}" r:id="rId{1000 + n}"/>')
        pres_rels.append(f'<Relationship Id="rId{1000 + n}" Type="{_RT}slide" '
                         f'Target="slides/slide{n}.xml"/>')

    files['docProps/core.xml'] = _core_xml(pres_meta)

    buf = BytesIO()
    with zipfile.ZipFile(buf, 'w', zipfile.ZIP_DEFLATED, compresslevel=1) as z:
        z.writestr('[Content_Types].xml',
                   templates['[Content_Types].xml'].replace('{OVERRIDES}', ''.join(overrides)))
        z.writestr('ppt/presentation.xml', templates['ppt/presentation.xml'].replace(
            '{SLD_ID_LST}', f'<p:sldIdLst>{"".join(sld_ids)}</p:sldIdLst>'))
        z.writestr('ppt/_rels/presentation.xml.rels',
                   templates['ppt/_rels/presentation.xml.rels'].replace(
                       '{SLIDE_RELS}', ''.join(pres_rels)))
        for name, data in static.items():
            z.writestr(name, data)
        for name, data in files.items():
            if name.endswith('.jpeg'):
                z.writestr(name, data, compress_type=zipfile.ZIP_STORED)
            else:
                z.writestr(name, data)
    return buf.getvalue()
