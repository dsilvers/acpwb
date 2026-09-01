import functools
import hashlib
import math
import random
from django.utils.text import slugify

from apps.people.generators import FIRST_NAMES, LAST_NAMES, TITLES, DEPARTMENTS

from .data.verbs import VERBS
from .data.adjectives import ADJECTIVES
from .data.nouns import NOUNS
from .data.domains import DOMAINS, INDUSTRIES
from .data.organizations import ORGANIZATIONS, ORG_SLUG_MAP
from .data.themes import THEMES
from .data.slogans import ORG_SLOGAN_TEMPLATES
from .data.charts import CHART_COLORS
from .data.footnotes import FOOTNOTE_TEMPLATES
from .data.text import TITLE_CASE_LOWER, ACRONYMS
from .data.slide_templates import (
    TITLE_TEMPLATES, SUBTITLES, AGENDA_SECTIONS, CONTENT_TEMPLATES, QUOTE_TEMPLATES,
    STAT_TEMPLATES, TAKEAWAY_TEMPLATES, CHART_TEMPLATES, CHART_LABELS_POOL, VENUES,
    IMAGE_CAPTIONS,
)
from .data.section_titles import SECTION_TITLES
from .data.callout import CALLOUT_TEMPLATES
from .data.two_column import TWO_COLUMN_TEMPLATES
from .data.timeline import TIMELINE_TEMPLATES
from .data.process import PROCESS_TEMPLATES
from .data.case_studies import CASE_STUDY_TEMPLATES
from .data.speaker_notes import SPEAKER_NOTES_TEMPLATES
from .data.appendix import APPENDIX_TEMPLATES
from .image_selector import pick_background, pick_meme


def _rng_from_seed(seed_str):
    seed_int = int(hashlib.md5(str(seed_str).encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed_int)



def _donut_arcs(items):
    """Compute SVG arc path strings for a donut chart (viewBox 0 0 200 200)."""
    cx, cy, r_out, r_in = 95, 95, 82, 52
    arcs = []
    angle = -90.0
    for i, item in enumerate(items):
        if item['pct'] <= 0:
            continue
        sweep = item['pct'] / 100 * 360
        end_angle = angle + sweep
        a1, a2 = math.radians(angle), math.radians(end_angle)
        x1 = round(cx + r_out * math.cos(a1), 1)
        y1 = round(cy + r_out * math.sin(a1), 1)
        x2 = round(cx + r_out * math.cos(a2), 1)
        y2 = round(cy + r_out * math.sin(a2), 1)
        ix1 = round(cx + r_in * math.cos(a2), 1)
        iy1 = round(cy + r_in * math.sin(a2), 1)
        ix2 = round(cx + r_in * math.cos(a1), 1)
        iy2 = round(cy + r_in * math.sin(a1), 1)
        large = 1 if sweep > 180 else 0
        path = (f"M{x1},{y1} A{r_out},{r_out} 0 {large},1 {x2},{y2} "
                f"L{ix1},{iy1} A{r_in},{r_in} 0 {large},0 {ix2},{iy2} Z")
        arcs.append({
            'path': path,
            'color': CHART_COLORS[i % len(CHART_COLORS)],
            'label': item['label'],
            'pct': item['pct'],
        })
        angle = end_angle
    return arcs


def _line_points(items):
    """Compute SVG x/y coordinates for a line chart (viewBox 0 0 380 140)."""
    pl, pr, pt, pb = 8, 8, 10, 22
    w, h = 380 - pl - pr, 140 - pt - pb
    vals = [item['value'] for item in items]
    lo, hi = min(vals), max(vals)
    span = max(hi - lo, 1)
    n = len(items)
    points = []
    for i, item in enumerate(items):
        x = round(pl + (i / (n - 1)) * w, 1) if n > 1 else pl + w / 2
        y = round(pt + (1 - (item['value'] - lo) / span) * h, 1)
        points.append({'x': x, 'y': y, 'label': item['label'], 'value': item['value']})
    path_d = 'M ' + ' L '.join(f"{p['x']},{p['y']}" for p in points)
    return points, path_d


def _render_chart_svg(slide):
    """Return an SVG string for the chart slide. All rendering logic lives here."""
    chart_type = slide.get('chart_type', 'bar_h')

    if chart_type == 'bar_h':
        bars = slide['chart_bars']
        row_h = 20
        pad_l, pad_r, pad_t = 96, 46, 4
        chart_w = 380 - pad_l - pad_r
        height = pad_t + len(bars) * row_h + (len(bars) - 1) * 5 + 4
        parts = [f'<svg viewBox="0 0 380 {height}" preserveAspectRatio="xMidYMid meet" style="width:100%;overflow:visible">']
        for i, bar in enumerate(bars):
            y = pad_t + i * (row_h + 5)
            bar_w = round(bar['pct'] / 100 * chart_w)
            lbl = bar['label'][:16]
            parts += [
                f'<text x="{pad_l - 6}" y="{y + row_h * 0.72:.0f}" text-anchor="end" '
                f'font-size="9" fill="currentColor" opacity=".65">{lbl}</text>',
                f'<rect x="{pad_l}" y="{y}" width="{bar_w}" height="{row_h}" '
                f'fill="{bar["color"]}" rx="2"/>',
                f'<text x="{pad_l + bar_w + 4}" y="{y + row_h * 0.72:.0f}" '
                f'font-size="9" font-weight="700" fill="{bar["color"]}">{bar["value"]}%</text>',
            ]
        parts.append('</svg>')
        return '\n'.join(parts)

    elif chart_type == 'bar_v':
        bars = slide['chart_bars']
        n = len(bars)
        pad_l, pad_r, pad_t, pad_b = 6, 6, 8, 30
        chart_h = 120
        bar_w = max(14, min(36, (380 - pad_l - pad_r - (n - 1) * 4) // n))
        total_w = n * bar_w + (n - 1) * 4
        start_x = (380 - total_w) // 2
        height = pad_t + chart_h + pad_b
        parts = [f'<svg viewBox="0 0 380 {height}" preserveAspectRatio="xMidYMid meet" style="width:100%">']
        for i, bar in enumerate(bars):
            x = start_x + i * (bar_w + 4)
            bh = round(bar['pct'] / 100 * chart_h)
            by = pad_t + chart_h - bh
            cx = x + bar_w // 2
            lbl = bar['label'][:10]
            parts += [
                f'<rect x="{x}" y="{by}" width="{bar_w}" height="{bh}" fill="{bar["color"]}" rx="2"/>',
                f'<text x="{cx}" y="{by - 3}" text-anchor="middle" '
                f'font-size="8" font-weight="700" fill="{bar["color"]}">{bar["value"]}%</text>',
                f'<text x="{cx}" y="{pad_t + chart_h + 14}" text-anchor="middle" '
                f'font-size="8" fill="currentColor" opacity=".6">{lbl}</text>',
            ]
        parts.append('</svg>')
        return '\n'.join(parts)

    elif chart_type == 'line':
        pts, path_d = slide['chart_line_pts'], slide['chart_line_path']
        if not pts:
            return ''
        uid = slide.get('uid', 'lc')
        area_close = f" L{pts[-1]['x']},118 L{pts[0]['x']},118 Z"
        dots = ''.join(
            f'<circle cx="{p["x"]}" cy="{p["y"]}" r="3.5" fill="var(--slide-accent,#c9a84c)"/>'
            for p in pts
        )
        n_pts = len(pts)
        def _label_anchor(idx):
            if idx == 0:
                return 'start'
            if idx == n_pts - 1:
                return 'end'
            return 'middle'
        labels = ''.join(
            f'<text x="{p["x"]}" y="136" text-anchor="{_label_anchor(i)}" font-size="8" '
            f'fill="currentColor" opacity=".55">{p["label"][:14]}</text>'
            for i, p in enumerate(pts)
        )
        pts_str = ' '.join(f'{p["x"]},{p["y"]}' for p in pts)
        return (
            f'<svg viewBox="0 0 380 140" preserveAspectRatio="xMidYMid meet" style="width:100%">'
            f'<defs><linearGradient id="lg{uid}" x1="0" y1="0" x2="0" y2="1">'
            f'<stop offset="0%" stop-color="var(--slide-accent,#c9a84c)" stop-opacity="0.3"/>'
            f'<stop offset="100%" stop-color="var(--slide-accent,#c9a84c)" stop-opacity="0.02"/>'
            f'</linearGradient></defs>'
            f'<path d="{path_d}{area_close}" fill="url(#lg{uid})"/>'
            f'<polyline points="{pts_str}" fill="none" stroke="var(--slide-accent,#c9a84c)" '
            f'stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round"/>'
            f'{dots}{labels}</svg>'
        )

    elif chart_type == 'donut':
        arcs = slide['chart_arcs']
        arc_paths = ''.join(f'<path d="{a["path"]}" fill="{a["color"]}"/>' for a in arcs)
        legend_items = ''.join(
            f'<div class="slide-donut-legend-item">'
            f'<span class="slide-donut-swatch" style="background:{a["color"]}"></span>'
            f'<span class="slide-donut-label">{a["label"]}</span>'
            f'<span class="slide-donut-pct">{a["pct"]}%</span>'
            f'</div>'
            for a in arcs
        )
        donut_svg = (
            f'<svg viewBox="0 0 200 200" preserveAspectRatio="xMidYMid meet" class="slide-donut-svg">'
            f'{arc_paths}</svg>'
        )
        return f'<div class="slide-donut-wrap">{donut_svg}<div class="slide-donut-legend">{legend_items}</div></div>'

    return ''


def generate_org_slogan(org_slug):
    rng = _rng_from_seed(f"orgslogan_{org_slug}")
    industry = rng.choice(INDUSTRIES)
    domain = rng.choice(DOMAINS)
    tmpl = rng.choice(ORG_SLOGAN_TEMPLATES)
    return tmpl.format(domain=domain, industry=industry)


def _watermark(seed):
    return hashlib.md5(f"acpwb_pres_{seed}".encode()).hexdigest()[:8]


def _build_pres_seed(org_slug, year, month, day, slug):
    return f"pres_{org_slug}_{year}_{month:02d}_{day:02d}_{slug}"



def _smart_title(s):
    """Title-case a string, keeping common prepositions/articles lowercase and restoring acronyms."""
    words = s.split()
    out = []
    for i, w in enumerate(words):
        lower = w.lower()
        if lower in ACRONYMS:
            out.append(ACRONYMS[lower])
        elif i == 0 or lower not in TITLE_CASE_LOWER:
            out.append(w.capitalize())
        else:
            out.append(w)
    return ' '.join(out)


def _title_from_slug(slug):
    """Recover display title from a slug of the form <words>-<NNNN>."""
    parts = slug.rsplit('-', 1)
    body = parts[0] if (len(parts) == 2 and parts[1].isdigit()) else slug
    return _smart_title(body.replace('-', ' '))


def _slug_from_title(title, num):
    """Build a URL slug from a title + 4-digit number."""
    return f"{slugify(title)}-{num}"


def _generate_authors(pres_seed, count):
    rng = _rng_from_seed(f"authors_{pres_seed}")
    authors = []
    used_emails = set()
    for i in range(count):
        first = rng.choice(FIRST_NAMES)
        last = rng.choice(LAST_NAMES)
        email_base = f"{first.split()[0].lower()}.{last.lower()}@acpwb.com"
        email = email_base
        n = 2
        while email in used_emails:
            email = f"{first.split()[0].lower()}.{last.lower()}{n}@acpwb.com"
            n += 1
        used_emails.add(email)
        avatar_seed = hashlib.md5(f"{first}{last}{i}".encode()).hexdigest()[:16]
        parts = (first[0] + last[0]).upper()
        authors.append({
            'first_name': first,
            'last_name': last,
            'full_name': f"{first} {last}",
            'title': rng.choice(TITLES),
            'department': rng.choice(DEPARTMENTS),
            'email': email,
            'avatar_seed': avatar_seed,
            'initials': parts,
        })
    return authors


def _fill(template, rng, **extra):
    subs = {
        'n': rng.randint(15, 87),
        'm': rng.randint(2, 480),
        'x': rng.randint(2, 10),
        'pct': rng.randint(12, 94),
        'months': rng.randint(3, 18),
        'years': rng.randint(3, 20),
        'regions': rng.randint(4, 12),
        'year': rng.randint(2015, 2025),
    }
    subs.update(extra)
    try:
        return template.format(**subs)
    except KeyError:
        return template


def generate_presentation_meta(org_slug, year, month, day, slug):
    pres_seed = _build_pres_seed(org_slug, year, month, day, slug)
    rng = _rng_from_seed(pres_seed)

    org_name = ORG_SLUG_MAP.get(org_slug, org_slug.replace('-', ' ').title())
    industry = rng.choice(INDUSTRIES)
    domain = rng.choice(DOMAINS)
    verb = rng.choice(VERBS)
    noun = rng.choice(NOUNS)

    # Title is always decoded from the slug so the URL always matches what's shown
    title = _title_from_slug(slug)

    adj = rng.choice(ADJECTIVES)
    subtitle_tmpl = rng.choice(SUBTITLES)
    subtitle = _fill(subtitle_tmpl, rng, industry=industry, domain=domain, org=org_name,
                     verb=verb, noun=noun, adj=adj)

    venue_city, venue_name_tmpl = rng.choice(VENUES)
    venue_name = _fill(venue_name_tmpl, rng, industry=industry, domain=domain)

    author_count = rng.choices([1, 2, 3], weights=[30, 45, 25])[0]
    authors = _generate_authors(pres_seed, author_count)

    theme = rng.choice(THEMES)
    slide_count = rng.randint(10, 20)

    thumb_bg = pick_background(pres_seed, 0)

    pres_url = f"/presentations/{org_slug}/{year}/{month:02d}/{day:02d}/{slug}/"
    return {
        'org_slug': org_slug,
        'org_name': org_name,
        'year': year,
        'month': month,
        'day': day,
        'slug': slug,
        'title': title,
        'subtitle': subtitle,
        'authors': authors,
        'theme': theme,
        'slide_count': slide_count,
        'watermark_token': _watermark(pres_seed),
        'industry': industry,
        'domain': domain,
        'pres_seed': pres_seed,
        'pub_date_display': _pub_date_display(year, month, day),
        'venue_city': venue_city,
        'venue_name': venue_name,
        'thumb_bg': thumb_bg,
        'pres_url': pres_url,
    }


def _pub_date_display(year, month, day):
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
              'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    return f"{months[month - 1]} {day}, {year}"


def generate_slide(pres_meta, slide_num):
    pres_seed = pres_meta['pres_seed']
    slide_count = pres_meta['slide_count']
    industry = pres_meta['industry']
    domain = pres_meta['domain']
    org_name = pres_meta['org_name']
    authors = pres_meta['authors']

    rng = _rng_from_seed(f"{pres_seed}_slide{slide_num}")

    fill_kwargs = dict(
        industry=industry, domain=domain, org=org_name,
        verb=rng.choice(VERBS), adj=rng.choice(ADJECTIVES),
        noun=rng.choice(NOUNS),
    )
    if authors:
        fill_kwargs.update({
            'first_name': authors[0]['first_name'],
            'last_name': authors[0]['last_name'],
            'title': authors[0]['title'],
        })

    # Determine slide type
    if slide_num == 1:
        slide_type = 'title'
    elif slide_num == 2:
        slide_type = 'agenda'
    elif slide_num == slide_count:
        slide_type = 'qanda'
    elif slide_num == slide_count - 1:
        slide_type = 'summary'
    else:
        # Interior slides: weighted random from content types
        slide_type = rng.choices(
            ['content', 'stat', 'quote', 'chart', 'image', 'meme',
             'two_column', 'timeline', 'section_divider', 'process', 'case_study', 'callout', 'appendix'],
            weights=[30, 12, 10, 12, 15, 10, 14, 8, 6, 8, 10, 10, 5],
        )[0]

    slide = {
        'num': slide_num,
        'total': slide_count,
        'type': slide_type,
        'theme': pres_meta['theme'],
        'bg_image': pick_background(pres_seed, 0 if slide_type == 'title' else slide_num),
    }

    if slide_type == 'title':
        slide['heading'] = pres_meta['title']
        slide['subheading'] = pres_meta['subtitle']
        slide['authors'] = authors

    elif slide_type == 'agenda':
        agenda = rng.choice(AGENDA_SECTIONS)
        slide['heading'] = rng.choice([
            "Today's Agenda",
            "Session Overview",
            "What We'll Cover",
            "Agenda",
            "Our Roadmap for Today",
            "Discussion Framework",
            "Program Outline",
            "What to Expect",
        ])
        slide['items'] = agenda

    elif slide_type == 'content':
        tmpl_heading, tmpl_bullets = rng.choice(CONTENT_TEMPLATES)
        slide['heading'] = _fill(tmpl_heading, rng, **fill_kwargs)
        if all(b.startswith("Stage ") for b in tmpl_bullets):
            bullets = tmpl_bullets
        else:
            bullets = rng.sample(tmpl_bullets, rng.randint(3, min(6, len(tmpl_bullets))))
        slide['bullets'] = [_fill(b, rng, **fill_kwargs) for b in bullets]

    elif slide_type == 'stat':
        stat_group = rng.choice(STAT_TEMPLATES)
        stats = []
        for val_tmpl, label_tmpl in stat_group:
            stats.append({
                'value': _fill(val_tmpl, rng, **fill_kwargs),
                'label': _fill(label_tmpl, rng, **fill_kwargs),
            })
        slide['stats'] = stats
        slide['heading'] = f"{industry} {domain}: By the Numbers"

    elif slide_type == 'quote':
        quote_tmpl, attr_tmpl = rng.choice(QUOTE_TEMPLATES)
        slide['quote'] = _fill(quote_tmpl, rng, **fill_kwargs)
        slide['attribution'] = _fill(attr_tmpl, rng, **fill_kwargs)

    elif slide_type == 'chart':
        chart_tmpl = rng.choice(CHART_TEMPLATES)
        slide['heading'] = _fill(chart_tmpl['title'], rng, **fill_kwargs)
        slide['chart_source'] = _fill(chart_tmpl['source'], rng, **fill_kwargs)
        chart_type = rng.choice(['bar_h', 'bar_h', 'bar_v', 'line', 'donut'])
        slide['chart_type'] = chart_type
        labels = rng.choice(CHART_LABELS_POOL)

        if chart_type in ('bar_h', 'bar_v'):
            values = [rng.randint(15, 95) for _ in labels]
            max_val = max(values)
            slide['chart_bars'] = [
                {'label': lbl, 'value': val, 'pct': round(val / max_val * 100),
                 'color': CHART_COLORS[i % len(CHART_COLORS)]}
                for i, (lbl, val) in enumerate(zip(labels, values))
            ]
        elif chart_type == 'line':
            base = rng.randint(20, 55)
            v = base
            items = []
            for lbl in labels:
                v = max(5, min(97, v + rng.randint(-10, 18)))
                items.append({'label': lbl, 'value': v})
            slide['chart_line_pts'], slide['chart_line_path'] = _line_points(items)
        elif chart_type == 'donut':
            raw = [rng.randint(5, 40) for _ in labels]
            total = sum(raw)
            pcts = [round(r / total * 100) for r in raw]
            pcts[0] += 100 - sum(pcts)  # fix rounding
            items = [{'label': lbl, 'pct': pct} for lbl, pct in zip(labels, pcts)]
            slide['chart_arcs'] = _donut_arcs(items)
        slide['uid'] = slide.get('uid', f"{pres_seed}_{slide_num}")
        slide['chart_svg'] = _render_chart_svg(slide)

    elif slide_type == 'image':
        img_path = pick_background(pres_seed, slide_num)
        if img_path:
            slide['image_path'] = img_path
            slide['caption'] = _fill(rng.choice(IMAGE_CAPTIONS), rng,
                                     venue_city=pres_meta.get('venue_city', ''),
                                     org_name=pres_meta['org_name'],
                                     **fill_kwargs)
        else:
            # No background library yet — degrade to content slide
            slide['type'] = 'content'
            tmpl_heading, tmpl_bullets = rng.choice(CONTENT_TEMPLATES)
            slide['heading'] = _fill(tmpl_heading, rng, **fill_kwargs)
            bullets = rng.sample(tmpl_bullets, rng.randint(3, min(5, len(tmpl_bullets))))
            slide['bullets'] = [_fill(b, rng, **fill_kwargs) for b in bullets]

    elif slide_type == 'meme':
        meme_path = pick_meme(pres_seed, slide_num)
        if meme_path:
            slide['meme_path'] = meme_path
        else:
            # No meme library yet — degrade to quote slide
            slide['type'] = 'quote'
            quote_tmpl, attr_tmpl = rng.choice(QUOTE_TEMPLATES)
            slide['quote'] = _fill(quote_tmpl, rng, **fill_kwargs)
            slide['attribution'] = _fill(attr_tmpl, rng, **fill_kwargs)

    elif slide_type == 'summary':
        takeaways_group = rng.choice(TAKEAWAY_TEMPLATES)
        slide['heading'] = "Key Takeaways"
        slide['bullets'] = [_fill(t, rng, **fill_kwargs) for t in takeaways_group]

    elif slide_type == 'two_column':
        tmpl = rng.choice(TWO_COLUMN_TEMPLATES)
        slide['heading'] = _fill(tmpl['heading'], rng, **fill_kwargs)
        slide['left_label'] = _fill(tmpl['left_label'], rng, **fill_kwargs)
        slide['right_label'] = _fill(tmpl['right_label'], rng, **fill_kwargs)
        slide['left_items'] = [_fill(item, rng, **fill_kwargs) for item in tmpl['left_items']]
        slide['right_items'] = [_fill(item, rng, **fill_kwargs) for item in tmpl['right_items']]

    elif slide_type == 'timeline':
        tmpl = rng.choice(TIMELINE_TEMPLATES)
        slide['heading'] = _fill(tmpl['heading'], rng, **fill_kwargs)
        slide['milestones'] = [
            {'label': _fill(label, rng, **fill_kwargs), 'desc': _fill(desc, rng, **fill_kwargs)}
            for label, desc in tmpl['milestones']
        ]

    elif slide_type == 'section_divider':
        slide['heading'] = rng.choice(SECTION_TITLES)
        slide['section_num'] = slide_num

    elif slide_type == 'process':
        tmpl = rng.choice(PROCESS_TEMPLATES)
        slide['heading'] = _fill(tmpl['heading'], rng, **fill_kwargs)
        slide['steps'] = [
            {'name': _fill(name, rng, **fill_kwargs), 'desc': _fill(desc, rng, **fill_kwargs)}
            for name, desc in tmpl['steps']
        ]

    elif slide_type == 'case_study':
        tmpl = rng.choice(CASE_STUDY_TEMPLATES)
        slide['heading'] = _fill(tmpl['heading'], rng, **fill_kwargs)
        slide['org_type'] = _fill(tmpl['org_type'], rng, **fill_kwargs)
        slide['challenge'] = _fill(tmpl['challenge'], rng, **fill_kwargs)
        slide['approach'] = _fill(tmpl['approach'], rng, **fill_kwargs)
        slide['result'] = _fill(tmpl['result'], rng, **fill_kwargs)

    elif slide_type == 'callout':
        stat_tmpl, desc_tmpl = rng.choice(CALLOUT_TEMPLATES)
        slide['stat'] = _fill(stat_tmpl, rng, **fill_kwargs)
        slide['description'] = _fill(desc_tmpl, rng, **fill_kwargs)

    elif slide_type == 'appendix':
        tmpl = rng.choice(APPENDIX_TEMPLATES)
        slide['heading'] = _fill(tmpl['heading'], rng, **fill_kwargs)
        slide['content'] = _fill(tmpl['content'], rng, **fill_kwargs)

    elif slide_type == 'qanda':
        slide['heading'] = "Questions & Discussion"
        slide['contact_email'] = authors[0]['email'] if authors else 'info@acpwb.com'
        slide['org_name'] = org_name

    # Footnote: ~40% chance on interior content-heavy slides
    if slide_type in ('content', 'stat', 'quote', 'summary', 'two_column', 'chart') and rng.random() < 0.40:
        slide['footnote'] = _fill(rng.choice(FOOTNOTE_TEMPLATES), rng, **fill_kwargs)

    # Speaker note: ~30% chance on interior slides
    if slide_type not in ('title', 'agenda', 'qanda', 'section_divider', 'appendix') and rng.random() < 0.30:
        slide['speaker_note'] = _fill(rng.choice(SPEAKER_NOTES_TEMPLATES), rng, **fill_kwargs)

    return slide


def _generate_title(rng, industry, domain, verb, noun):
    """Generate a readable presentation title from TITLE_TEMPLATES."""
    tmpl = rng.choice(TITLE_TEMPLATES)
    adj_clean = rng.choice(ADJECTIVES).replace('-', ' ')
    return _fill(tmpl, rng, industry=industry, domain=domain,
                 verb=verb, adj=adj_clean, noun=noun)


@functools.lru_cache(maxsize=2048)
def generate_presentations_for_context(context_seed, count=12):
    """Generate a list of presentation metadata dicts for list pages."""
    results = []
    used_slugs = set()

    for i in range(count):
        item_rng = _rng_from_seed(f"{context_seed}_item{i}")
        org_name = item_rng.choice(ORGANIZATIONS)
        org_slug = slugify(org_name)
        year = item_rng.randint(2008, 2025)
        month = item_rng.randint(1, 12)
        day = item_rng.randint(1, 28)

        # Build title first, then derive slug from it
        industry = item_rng.choice(INDUSTRIES)
        domain = item_rng.choice(DOMAINS)
        verb = item_rng.choice(VERBS)
        noun = item_rng.choice(NOUNS)
        title = _generate_title(item_rng, industry, domain, verb, noun)
        num = item_rng.randint(1000, 9999)
        slug = _slug_from_title(title, num)
        while slug in used_slugs:
            num = item_rng.randint(1000, 9999)
            slug = _slug_from_title(title, num)
        used_slugs.add(slug)

        meta = generate_presentation_meta(org_slug, year, month, day, slug)
        results.append(meta)

    return results


def generate_landing_orgs(page=1):
    """Generate the featured organizations and their presentations for the landing page."""
    doubled = ORGANIZATIONS * 2  # allows wrap-around for any page
    start = ((page - 1) * 8) % len(ORGANIZATIONS)
    featured_org_names = doubled[start:start + 8]
    sections = []
    for org_name in featured_org_names:
        org_slug = slugify(org_name)
        context_seed = f"presorg_{org_slug}_p1"
        presentations = generate_presentations_for_context(context_seed, count=3)
        sections.append({
            'org_name': org_name,
            'org_slug': org_slug,
            'org_slogan': generate_org_slogan(org_slug),
            'presentations': presentations,
        })
    return sections


def generate_related(pres_seed, org_slug, pres_meta):
    """Return same_org and related_topic presentation lists for the sidebar."""
    same_org = generate_presentations_for_context(
        f"presrelated_org_{org_slug}_{pres_seed}", count=4
    )
    # Filter out the current presentation
    same_org = [p for p in same_org if p['slug'] != pres_meta['slug']][:4]

    domain = pres_meta['domain']
    related = generate_presentations_for_context(
        f"presrelated_domain_{domain}_{pres_seed}", count=4
    )
    related = [p for p in related if p['slug'] != pres_meta['slug']][:4]

    return same_org, related
