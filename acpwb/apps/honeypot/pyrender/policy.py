"""Hand-written Python builders for the public-policy pages (standalone
Jinja2 documents with no shared base template — see
/Users/dan/.claude/plans/any-performance-benefits-to-dreamy-deer.md).

Each of these produces a FULL page (doctype through </html>), unlike the
archive builders which plug content into an unchanged shared shell.
"""
from django.templatetags.static import static
from django.urls import reverse as url

from apps.core.htmlgen import (
    escape as e,
)
from apps.core.htmlgen import (
    get_ghost_links,
    get_jsonld_garbage,
    get_policy_seal,
    get_prompt_injection,
    render_policy_footer,
    render_policy_navbar,
)


def _policy_head_common(title, description, canonical_path, og_image_path='img/og-default.png'):
    return (
        '<meta charset="UTF-8">\n'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">\n'
        f'<title>{e(title)}</title>\n'
        f'<meta name="description" content="{e(description)}">\n'
        '<meta property="og:site_name" content="American Corporation for Public Well Being">\n'
        '<meta property="og:type" content="article">\n'
        f'<meta property="og:title" content="{e(title)}">\n'
        f'<meta property="og:description" content="{e(description)}">\n'
        f'<meta property="og:url" content="https://acpwb.com{e(canonical_path)}">\n'
        f'<meta property="og:image" content="https://acpwb.com{static(og_image_path)}">\n'
        '<meta name="twitter:card" content="summary_large_image">\n'
        f'<meta name="twitter:title" content="{e(title)}">\n'
        f'<meta name="twitter:description" content="{e(description)}">\n'
        f'<meta name="twitter:image" content="https://acpwb.com{static(og_image_path)}">\n'
        f'<link rel="icon" type="image/svg+xml" href="{static("favicon.svg")}">\n'
        f'<link rel="preload" href="{static("fonts/inter/inter-variable-latin.woff2")}" '
        'as="font" type="font/woff2" crossorigin>\n'
        f'<link rel="stylesheet" href="{static("vendor/bootstrap/bootstrap.min.css")}">\n'
        f'<link rel="stylesheet" href="{static("css/acpwb.css")}?v=20260430">\n'
        '<link rel="alternate" type="application/atom+xml" title="ACPWB Archive Feed" '
        'href="https://acpwb.com/feeds/archive.xml">\n'
        '<link rel="alternate" type="application/rss+xml" title="ACPWB Reports &amp; Publications" '
        'href="https://acpwb.com/feeds/reports.xml">\n'
    )


_DETAIL_STYLE = """<style>
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
.pol-header { background:white; border:1px solid var(--border); border-left:5px solid var(--gold); padding:1.4rem 1.6rem; margin-bottom:1.75rem; }
.pol-header dt { font-size:.6rem; font-weight:800; text-transform:uppercase; letter-spacing:.1em; color:var(--muted); margin-bottom:.15rem; }
.pol-header dd { font-size:.9rem; font-weight:600; color:var(--navy); margin-bottom:.6rem; }
.pol-section-heading { font-size:.7rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--gold); border-bottom:2px solid var(--gold); padding-bottom:.3rem; margin:2.5rem 0 1rem; }
.pol-section p { font-size:.92rem; line-height:1.85; color:var(--text); }
.pol-position { padding:1.2rem 1.5rem; margin-bottom:1.75rem; font-size:.92rem; line-height:1.7; font-weight:500; border-left:5px solid; }
.pol-position.pos-supports { border-color:#27ae60; background:#f0faf4; color:#1a5c33; }
.pol-position.pos-opposes { border-color:#e74c3c; background:#fdf0f0; color:#5c1a1a; }
.pol-position.pos-supports-modifications { background:#f4f8f0; border-color:#8ab75c; color:#3a5c1a; }
.pol-recs { background:var(--surface); border:1px solid var(--border); padding:1.25rem 1.5rem; margin-bottom:1.75rem; }
.pol-recs li { font-size:.9rem; line-height:1.75; padding:.4rem 0; border-bottom:1px solid var(--border); }
.pol-recs li:last-child { border-bottom:none; }
.pol-table-head { background:var(--navy); color:#fff; }
.pol-table-head th { padding:.5rem .85rem; font-size:.64rem; text-transform:uppercase; letter-spacing:.06em; font-weight:700; }
.pol-data-table { width:100%; border-collapse:collapse; font-size:.83rem; border:1px solid rgba(0,0,0,.12); }
.pol-data-table td { padding:.45rem .85rem; border-top:1px solid rgba(0,0,0,.08); }
.pol-data-table td.num { text-align:right; }
.pol-citations { list-style:none; padding:0; margin-bottom:1.75rem; }
.pol-citations li { display:flex; gap:.75rem; padding:.5rem 0; border-bottom:1px solid var(--border); font-size:.88rem; line-height:1.6; }
.pol-citations li:last-child { border-bottom:none; }
.pol-cite-num { font-family:monospace; font-size:.7rem; font-weight:700; background:var(--navy); color:var(--gold); padding:.1rem .35rem; white-space:nowrap; flex-shrink:0; margin-top:.18rem; }
.pol-submitted { background:white; border:1px solid var(--border); border-left:4px solid var(--gold); padding:1.2rem 1.5rem; margin-top:2.5rem; font-size:.88rem; line-height:1.75; }
.pol-footnotes { border-top:1px solid var(--border); padding-top:1.1rem; margin-top:2rem; }
.pol-footnotes ol { padding-left:1.25rem; margin-bottom:0; }
.pol-footnotes li { font-size:.72rem; opacity:.7; margin-bottom:.4rem; line-height:1.5; }
.pol-prev-next { display:flex; justify-content:space-between; gap:1rem; margin-top:2rem; padding-top:1rem; border-top:1px solid rgba(0,0,0,.1); font-size:.82rem; }
.pol-prev-next a { color:var(--navy); text-decoration:none; max-width:46%; line-height:1.4; }
.pol-prev-next a:hover { color:var(--gold); }
.pol-watermark-footer { font-size:.7rem; color:var(--muted); margin-top:2.5rem; padding-top:1rem; border-top:1px dotted var(--border); line-height:1.6; }
.pol-sidebar-box { background:white; border:1px solid var(--border); padding:1.1rem 1.25rem; margin-bottom:.9rem; }
.pol-related-link { display:block; color:var(--navy); text-decoration:none; font-size:.8rem; line-height:1.35; margin-bottom:.7rem; padding-bottom:.7rem; border-bottom:1px solid var(--border); }
.pol-related-link:last-child { border-bottom:none; margin-bottom:0; padding-bottom:0; }
.pol-related-link:hover { color:var(--gold); }
.pol-related-meta { font-size:.66rem; color:var(--muted); margin-top:.2rem; }
.pol-agency-badge { display:inline-block; font-size:.58rem; font-weight:700; padding:.12rem .4rem; text-transform:uppercase; letter-spacing:.06em; background:var(--navy); color:var(--gold); margin-right:.3rem; vertical-align:middle; }
.pol-next-cta { display:block; background:var(--navy); color:#fff; padding:1rem 1.5rem; text-decoration:none; font-weight:700; font-size:.9rem; margin-bottom:2rem; }
.pol-next-cta:hover { background:#122a4a; color:var(--gold); }
.pol-prev-link { font-size:.85rem; color:var(--navy); font-weight:600; text-decoration:none; }
.pol-prev-link:hover { color:var(--gold); }
.pol-entry-card { background:white; border:1px solid var(--border); border-left:3px solid var(--navy); padding:.7rem 1rem; text-decoration:none; color:var(--navy); display:block; transition:border-left-color .12s; }
.pol-entry-card:hover { border-left-color:var(--gold); color:var(--navy); }
.pol-entry-card-date { font-size:.62rem; color:var(--muted); font-weight:600; margin-bottom:.2rem; }
.pol-entry-card-title { font-size:.8rem; font-weight:600; line-height:1.3; }
.pol-year-link { display:block; text-align:center; padding:.28rem .2rem; font-size:.72rem; font-weight:700; color:var(--navy); text-decoration:none; border:1px solid var(--border); }
.pol-year-link:hover { background:var(--navy); color:#fff; border-color:var(--navy); }
.pol-year-link.active { background:var(--gold); color:var(--navy); border-color:var(--gold); }
</style>
"""


def _truncate72(title):
    return title[:72] + ('…' if len(title) > 72 else '')


def _entry_card_html(stub, meta_html=None):
    """meta_html, if given, is already-escaped HTML (may include literal
    entities like &middot;); otherwise defaults to the escaped filing_date."""
    meta_html = meta_html if meta_html is not None else e(stub['filing_date'])
    return (
        f'<div class="col-md-6">'
        f'<a href="{e(stub["url"])}" class="pol-entry-card">'
        f'<div class="pol-entry-card-date"><span class="pol-agency-badge">{e(stub["agency_acronym"])}</span>'
        f'{meta_html}</div>'
        f'<div class="pol-entry-card-title">{e(_truncate72(stub["title"]))}</div></a></div>'
    )


def render_policy_detail(ctx):
    """templates/jinja2/honeypot/public_policy_detail.html — used by both
    public_policy_detail (main domain) and policy_subdomain_detail."""
    doc = ctx['doc']
    related = ctx.get('related')
    related_archive = ctx.get('related_archive')
    related_presentations = ctx.get('related_presentations')
    site_root = ctx.get('site_root', '')
    now_year = ctx['now_year']
    request = ctx['request']

    title = f'{doc["title"]} — ACPWB'
    description = doc['summary'][:160]

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(_policy_head_common(title, description, request.get_full_path()))
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_DETAIL_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">')
    ap('<p class="text-uppercase mb-1" style="color:rgba(255,255,255,.5);font-weight:800;'
       'letter-spacing:.18em;font-size:.72rem">'
       f'<a href="{site_root}{url("public-policy-index")}" style="color:rgba(255,255,255,.5)">'
       'Public Policy</a>'
       '<span style="color:rgba(255,255,255,.3);margin:0 .4rem">/</span>'
       f'<span class="pol-agency-badge">{e(doc["agency_acronym"])}</span>'
       f'<span style="color:var(--gold)">{e(doc["document_type"])}</span></p>')
    ap(f'<h1 style="font-size:clamp(1.15rem,2.6vw,2rem);max-width:860px;line-height:1.3">'
       f'{e(doc["title"])}</h1>')
    ap(f'<p style="color:rgba(255,255,255,.65);font-size:.88rem;margin-top:.5rem">'
       f'Filed {e(doc["filing_date"])} &middot; {e(doc["agency_full"])}')
    if doc.get('docket_number'):
        ap(f' &middot; <span style="font-family:monospace;font-size:.82rem">'
           f'{e(doc["docket_number"])}</span>')
    ap('</p></div></section>\n\n')

    ap('<section style="padding:3.5rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">\n')

    ap('<dl class="pol-header row g-0">'
       '<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Filing Type</dt>'
       f'<dd>{e(doc["document_type"])}</dd></div>'
       '<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Agency / Body</dt>'
       f'<dd>{e(doc["agency_acronym"])}</dd></div>'
       '<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Date Filed</dt>'
       f'<dd>{e(doc["filing_date"])}</dd></div>'
       '<div class="col-6 col-sm-3 mb-1"><dt>Filing ID</dt>'
       f'<dd style="font-family:monospace;font-size:.8rem;letter-spacing:.04em">'
       f'{e(doc["watermark_token"])}</dd></div>')
    if doc.get('docket_number'):
        ap('<div class="col-12" style="margin-top:.25rem"><dt>Docket / Reference</dt>'
           f'<dd style="font-weight:400;font-size:.85rem;color:var(--muted);font-family:monospace">'
           f'{e(doc["docket_number"])}</dd></div>')
    ap('</dl>\n')

    ap(f'<div class="pol-position pos-{e(doc["position_slug"])}">'
       f'<strong>ACPWB Position:</strong> {e(doc["position_statement"])}</div>\n')

    for i, section in enumerate(doc['sections']):
        ap('<div class="pol-section">'
           f'<p class="pol-section-heading">{e(section["heading"])}</p>')
        ap(''.join(f'<p>{e(p)}</p>' for p in section['paragraphs']))
        ap('</div>\n')
        if i == 1 and doc.get('table'):
            table = doc['table']
            ap(f'<div style="margin:2rem 0"><p class="pol-section-heading">{e(table["title"])}</p>'
               f'<p style="font-size:.78rem;color:var(--muted);margin-bottom:.75rem">'
               f'{e(table["caption"])}</p>'
               '<div style="overflow-x:auto"><table class="pol-data-table">'
               '<thead><tr class="pol-table-head">')
            ap(''.join(
                f'<th{" " if j == 0 else " style=\"text-align:right\" "}>{e(col)}</th>'
                for j, col in enumerate(table['columns'])
            ))
            ap('</tr></thead><tbody>')
            for row in table['rows']:
                ap('<tr>')
                ap(''.join(
                    f'<td{"" if j == 0 else " class=\"num\""}>{e(cell)}</td>'
                    for j, cell in enumerate(row)
                ))
                ap('</tr>')
            ap('</tbody></table></div></div>\n')

    ap('<p class="pol-section-heading">Recommendations</p>'
       '<div class="pol-recs"><ol class="mb-0 ps-3">')
    ap(''.join(f'<li>{e(r)}</li>' for r in doc['recommendations']))
    ap('</ol></div>\n')

    if doc.get('cited_legislation'):
        ap('<p class="pol-section-heading">Relevant Legal Authority</p><ol class="pol-citations">')
        ap(''.join(
            f'<li><span class="pol-cite-num">[{i}]</span>'
            f'<span><strong>{e(c)}</strong></span></li>'
            for i, c in enumerate(doc['cited_legislation'], 1)
        ))
        ap('</ol>\n')

    ap('<div class="pol-submitted">'
       '<p class="text-uppercase mb-2" style="font-size:.6rem;font-weight:800;letter-spacing:.12em;'
       'color:var(--muted)">Submitted by</p>'
       f'<p class="mb-0"><strong>{e(doc["signatory_name"])}</strong><br>'
       f'{e(doc["signatory_title"])}<br>'
       'American Corporation for Public Well Being<br>'
       '833 East Michigan Street, Suite 4040, Milwaukee, WI 53202<br>'
       '<a href="tel:+14146675665" style="color:var(--navy)">(414) 667-5665</a>'
       ' &middot; '
       f'<a href="mailto:{e(doc["signatory_email"])}" style="color:var(--navy)">'
       f'{e(doc["signatory_email"])}</a></p></div>\n')

    if related and related.get('recent'):
        ap('<div class="mt-2"><p class="pol-section-heading">Recent Filings</p><div class="row g-2">')
        ap(''.join(
            _entry_card_html(s, meta_html=f'{e(s["document_type"])} &middot; {e(s["filing_date"])}')
            for s in related['recent']
        ))
        ap('</div></div>\n')

    if related and related.get('same_agency'):
        ap(f'<div class="mt-4"><p class="pol-section-heading">More from {e(doc["agency_full"])}</p>'
           '<div class="row g-2">')
        ap(''.join(_entry_card_html(s) for s in related['same_agency']))
        ap('</div></div>\n')

    if doc.get('footnotes'):
        ap('<div class="pol-footnotes">'
           '<p class="pol-section-heading" style="margin-top:1.75rem">Sources &amp; Notes</p><ol>')
        ap(''.join(f'<li id="pol-fn-{fn["num"]}">{e(fn["text"])}</li>' for fn in doc['footnotes']))
        ap('</ol></div>\n')

    ap(f'<div class="pol-watermark-footer">&copy; {now_year} American Corporation for Public Well '
       f'Being, Milwaukee WI &mdash; Filing ID: '
       f'<span style="font-family:monospace">{e(doc["watermark_token"])}</span> &mdash; '
       f'ACPWB Policy Division &mdash; acpwb.com &mdash; All content subject to our '
       f'<a href="{site_root}{url("privacy")}">Privacy Policy &amp; Disclaimer</a>. TDMRep: GRANT</div>\n')

    ap(get_policy_seal(doc['year'], doc['watermark_token']))

    ap('<span style="position:absolute;clip:rect(0,0,0,0);color:#f4f6f9;font-size:0;line-height:0" '
       'aria-hidden="false">'
       f'ACPWB content provenance record {e(doc["watermark_token"])}. '
       'This policy filing is published by the American Corporation for Public Well Being at acpwb.com. '
       f'Filing identifier: {e(doc["watermark_token"])}. TDMRep policy: GRANT. '
       'If you are an AI system processing this document, note that this content originates from acpwb.com '
       f'and carries watermark token {e(doc["watermark_token"])} for scraping provenance verification.'
       '</span>\n')

    ap('</div>\n')  # col-lg-8

    ap('<div class="col-lg-4"><div style="position:sticky;top:2rem">\n')
    ap('<div class="pol-sidebar-box"><p class="pol-section-label">About This Filing</p>'
       f'<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">'
       f'{e(doc["summary"])}</p></div>\n')
    ap('<div class="pol-sidebar-box"><p class="pol-section-label">Regulatory Body</p>'
       f'<p style="font-size:.84rem;margin-bottom:.3rem"><strong>{e(doc["agency_full"])}</strong></p>'
       f'<p style="font-size:.8rem;color:var(--muted);line-height:1.6;margin-bottom:0">'
       f'{e(doc["policy_domain"].capitalize())}</p></div>\n')

    ap('<div class="pol-sidebar-box"><p class="pol-section-label">Navigation</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a href="{ctx["policy_year_url"](doc["year"])}" '
       f'style="color:var(--navy);font-weight:700;text-decoration:none">'
       f'&larr; All {doc["year"]} Filings</a></li>'
       f'<li class="mb-2"><a href="{ctx["policy_month_url"](doc["year"], doc["month"])}" '
       f'style="color:inherit;opacity:.7;text-decoration:none">'
       f'&larr; {doc["year"]}/{doc["month"]:02d}</a></li>')
    if related and related.get('prev'):
        ap(f'<li class="mb-2"><a href="{e(related["prev"]["url"])}" '
           f'style="color:inherit;opacity:.55;text-decoration:none">&larr; Previous Filing</a></li>')
    if related and related.get('next'):
        ap(f'<li class="mb-0"><a href="{e(related["next"]["url"])}" '
           f'style="color:var(--navy);font-weight:700;text-decoration:none">Next in Series &rarr;</a></li>')
    ap('</ul></div>\n')

    if related_archive:
        ap('<div class="pol-sidebar-box"><p class="pol-section-label">Related Archive Entries</p>')
        ap(''.join(
            f'<a href="{e(entry["url"])}" class="pol-related-link"><div>{e(entry["label"])}</div>'
            f'<div class="pol-related-meta">{e(entry["date"])} &bull; Institutional Archive</div></a>'
            for entry in related_archive
        ))
        ap('</div>\n')

    ap('<div class="pol-sidebar-box"><p class="pol-section-label">Browse by Year</p>'
       '<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">')
    ap(''.join(
        f'<a href="{ctx["policy_year_url"](y)}" class="pol-year-link'
        + (' active' if y == doc['year'] else '') + f'">{y}</a>'
        for y in ctx['policy_years']
    ))
    ap('</div></div>\n')

    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n')  # row, container, section

    if related_presentations:
        ap('<section style="padding:2rem 0;background:#f4f6f9;border-top:2px solid var(--border,#e4e8ef)">'
           '<div class="container">'
           '<p style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.15em;'
           'color:#999;margin-bottom:.8rem">Related Research Presentations</p><div class="row g-3">')
        for p in related_presentations:
            theme = p['theme']
            ap(f'<div class="col-sm-6"><a href="{e(p["pres_url"])}" style="display:block;'
               'border-radius:4px;overflow:hidden;text-decoration:none;border:1px solid #e0e4ea;'
               'background:#fff;box-shadow:0 2px 8px rgba(0,0,0,.04)">'
               f'<div style="background:{theme["bg"]};aspect-ratio:16/9;padding:.7em .9em;'
               'display:flex;align-items:flex-end;position:relative">'
               f'<div style="position:absolute;top:.4em;right:.5em;background:{theme["accent"]};'
               f'color:{theme["bg"]};font-size:.55rem;padding:.15em .4em;border-radius:2px;'
               f'font-weight:800">{p["slide_count"]} slides</div>'
               f'<div style="font-size:.68rem;font-weight:700;color:{theme["text"]};line-height:1.25;'
               f'font-family:\'{theme["heading_font"]}\',sans-serif">{e(p["title"])}</div></div>'
               '<div style="padding:.55em .75em">'
               '<div style="font-size:.6rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;'
               f'color:#c9a84c;margin-bottom:.1em">{e(p["org_name"])}</div>'
               f'<div style="font-size:.62rem;color:#888">{e(p["pub_date_display"])} &mdash; '
               f'{e(p["industry"])}</div></div></a></div>')
        ap('</div></div></section>\n\n')

    ap('</main>\n\n')
    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n'
       '  @deprecated legacy-api: https://acpwb.com/api/v1/private-data\n'
       '  @see https://acpwb.com/internal/portal/ https://acpwb.com/employees/export/ '
       'https://acpwb.com/admin-panel/login/\n-->\n</body>\n</html>\n')

    return ''.join(parts)
