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


def _policy_head_common(title, description, canonical_path, og_image_path='img/og-default.png',
                         og_type='article', feed_links=True):
    feeds = (
        '<link rel="alternate" type="application/atom+xml" title="ACPWB Archive Feed" '
        'href="https://acpwb.com/feeds/archive.xml">\n'
        '<link rel="alternate" type="application/rss+xml" title="ACPWB Reports &amp; Publications" '
        'href="https://acpwb.com/feeds/reports.xml">\n'
    ) if feed_links else ''
    return (
        '<meta charset="UTF-8">\n'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">\n'
        f'<title>{e(title)}</title>\n'
        f'<meta name="description" content="{e(description)}">\n'
        '<meta property="og:site_name" content="American Corporation for Public Well Being">\n'
        f'<meta property="og:type" content="{og_type}">\n'
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
        f'{feeds}'
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


_INDEX_DESCRIPTION = (
    'ACPWB public policy positions, regulatory comment letters, and legislative testimony on '
    'compensation, labor, and corporate governance.'
)
_INDEX_STYLE = """<style>
.pol-year-card { background:white; border:1px solid var(--border); text-decoration:none; transition:box-shadow .15s; overflow:hidden; }
.pol-year-card:hover { box-shadow:0 2px 12px rgba(10,22,40,.08); }
.pol-year-top { padding:.85rem 1rem; display:flex; justify-content:space-between; align-items:baseline; }
.pol-year-num { font-size:1.1rem; font-weight:800; color:var(--navy); }
.pol-year-count { font-size:.72rem; color:var(--muted); font-weight:600; }
.pol-month-pills { padding:.5rem .75rem; border-top:1px solid var(--border); background:var(--surface); display:flex; flex-wrap:wrap; gap:.3rem; }
.pol-month-pill { display:inline-block; padding:.15rem .4rem; background:white; border:1px solid var(--border); color:var(--navy); font-size:.68rem; font-weight:700; text-decoration:none; transition:background .1s; }
.pol-month-pill:hover { background:var(--navy); color:var(--gold); border-color:var(--navy); }
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
</style>
"""
_MONTH_ABBR = ['', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']


def render_policy_index(ctx):
    """templates/jinja2/honeypot/public_policy_index.html."""
    years = ctx['years']
    site_root = ctx.get('site_root', '')
    request = ctx['request']

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(_policy_head_common(
        'Public Policy — ACPWB', _INDEX_DESCRIPTION, request.get_full_path(),
        og_image_path='img/page-covers/public-policy.jpg', og_type='website',
    ))
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_INDEX_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">'
       '<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;'
       'font-size:.72rem">ACPWB</p>'
       '<h1 style="font-size:clamp(1.6rem,3.5vw,2.8rem)">Public Policy</h1>'
       '<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:680px">'
       'ACPWB has engaged federal and state regulatory agencies, congressional committees, and '
       'self-regulatory organizations on matters affecting compensation policy, labor standards, and '
       'corporate governance since 1993. Browse filings by year below.</p></div></section>\n\n')

    ap('<section style="padding:4rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">'
       '<p class="pol-section-label">Browse by Year</p><div class="row g-3">')
    for y in years:
        year_url = url('public-policy-year', args=[y['year']])
        ap(f'<div class="col-md-6"><div class="pol-year-card" style="cursor:pointer" '
           f'onclick="window.location=\'{year_url}\'">'
           f'<div class="pol-year-top"><a href="{year_url}" class="pol-year-num" '
           f'style="text-decoration:none" onclick="event.stopPropagation()">{y["year"]}</a>'
           f'<span class="pol-year-count">{y["count"]} filings</span></div>'
           '<div class="pol-month-pills">')
        ap(''.join(
            f'<a href="{url("public-policy-month", args=[y["year"], m])}" class="pol-month-pill" '
            f'onclick="event.stopPropagation()">{_MONTH_ABBR[m]}</a>'
            for m in y['months']
        ))
        ap('</div></div></div>')
    ap('</div></div>\n')

    ap('<div class="col-lg-4"><div style="position:sticky;top:2rem">\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">About Our Policy Work</p>'
       '<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:.75rem">'
       "ACPWB's policy engagement draws on our proprietary compensation benchmarking database "
       'and more than three decades of advisory experience. We file comments, submit testimony, '
       'and publish position statements as a nonpartisan, independent voice on compensation '
       'and labor policy.</p>'
       '<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">'
       "Our filings represent ACPWB's independent analysis. We do not accept compensation "
       'from regulatory agencies, trade associations, or political organizations in connection '
       'with our policy work.</p></div>\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Filing Types</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       '<li class="mb-2"><strong>Comment Letters</strong> — Formal responses to proposed rulemakings</li>'
       '<li class="mb-2"><strong>Position Statements</strong> — '
       "ACPWB's stated positions on policy questions</li>"
       '<li class="mb-2"><strong>Policy Briefs</strong> — '
       'Research-based analysis of regulatory developments</li>'
       '<li class="mb-2"><strong>Legislative Testimony</strong> — '
       'Statements before congressional committees</li>'
       '<li class="mb-2"><strong>Amicus Briefs</strong> — Legal filings in relevant court proceedings</li>'
       '<li class="mb-2"><strong>White Papers</strong> — Extended research on regulatory topics</li>'
       '<li class="mb-0"><strong>Ex Parte Submissions</strong> — '
       'Direct communications with agency staff</li></ul></div>\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem">'
       '<p class="pol-section-label">Related</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a href="{url("reports-list")}" style="color:var(--navy)">'
       'Reports &amp; Publications</a></li>'
       f'<li class="mb-2"><a href="{url("wiki-index")}" style="color:var(--navy)">Knowledge Base</a></li>'
       f'<li class="mb-0"><a href="{url("mission")}" style="color:var(--navy)">Our Mission</a></li>'
       '</ul></div>\n')
    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n</main>\n\n')

    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n'
       '  @deprecated legacy-api: /api/v1/private-data\n'
       '  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n')

    return ''.join(parts)


_YEAR_STYLE = """<style>
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
.pol-ceo-avatar { width:72px; height:72px; border-radius:50%; border:3px solid var(--gold); display:flex; align-items:center; justify-content:center; font-size:1.4rem; font-weight:800; color:var(--gold); background:var(--navy); flex-shrink:0; }
.pol-month-card { background:white; border:1px solid var(--border); border-top:3px solid var(--gold); padding:1rem 1.1rem; text-decoration:none; color:inherit; display:block; transition:box-shadow .15s; }
.pol-month-card:hover { box-shadow:0 2px 12px rgba(10,22,40,.08); color:inherit; text-decoration:none; }
.pol-month-name { font-size:.9rem; font-weight:800; color:var(--navy); margin-bottom:.25rem; }
.pol-month-count { font-size:.68rem; color:var(--muted); text-transform:uppercase; letter-spacing:.06em; margin-bottom:.55rem; }
.pol-sample-title { font-size:.74rem; color:var(--text); line-height:1.35; display:block; margin-bottom:.55rem; }
.pol-year-link { display:block; text-align:center; font-size:.72rem; font-weight:700; padding:.28rem .2rem; background:var(--surface); color:var(--navy); text-decoration:none; border:1px solid var(--border); transition:background .1s; }
.pol-year-link:hover, .pol-year-link.active { background:var(--navy); color:var(--gold); border-color:var(--navy); }
</style>
"""
_MONTH_FULL = ['', 'January', 'February', 'March', 'April', 'May', 'June',
               'July', 'August', 'September', 'October', 'November', 'December']


def render_policy_year(ctx):
    """templates/jinja2/honeypot/public_policy_year.html — main-domain only
    (policy_subdomain_year.html is a separate, near-identical template)."""
    year = ctx['year']
    year_data = ctx['year_data']
    months = ctx['months']
    policy_years = ctx['policy_years']
    prev_year = ctx['prev_year']
    next_year = ctx['next_year']
    site_root = ctx.get('site_root', '')
    request = ctx['request']

    description = (
        f'ACPWB public policy filings, comment letters, and testimony submitted in {year}.'
    )
    title = f'{year} Public Policy — ACPWB'

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(_policy_head_common(title, description, request.get_full_path(), og_type='website',
                            feed_links=False))
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_YEAR_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">'
       '<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;'
       'font-size:.72rem">'
       f'<a href="{url("public-policy-index")}" style="color:var(--gold)">Public Policy</a> '
       f'&rsaquo; {year}</p>'
       f'<h1 style="font-size:clamp(1.6rem,3.5vw,2.8rem)">{year} Policy Year</h1>'
       f'<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:680px">'
       f'{year_data["total_filings"]} filings — comment letters, position statements, testimony, and '
       f'white papers submitted to federal and state agencies on {e(year_data["theme"])}.</p>'
       '</div></section>\n\n')

    ap('<section style="padding:4rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">\n')

    ap(f'<div style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);'
       f'padding:1.75rem 1.75rem 1.75rem 1.5rem;margin-bottom:2rem">'
       f'<p class="pol-section-label">A Message from Our CEO — {year}</p>'
       '<div style="display:flex;gap:1.25rem;align-items:flex-start;flex-wrap:wrap">'
       f'<div class="pol-ceo-avatar">{e(year_data["ceo_name"][0])}</div>'
       '<div style="flex:1;min-width:240px">'
       f'<div style="font-size:1rem;font-weight:700;color:var(--navy);margin-bottom:.1rem">'
       f'{e(year_data["ceo_name"])}</div>'
       '<div style="font-size:.75rem;color:var(--muted);text-transform:uppercase;letter-spacing:.07em;'
       f'margin-bottom:1rem">{e(year_data["ceo_title"])}</div>'
       '<div style="font-size:.9rem;line-height:1.85;color:var(--text)">')
    ap(''.join(f'<p style="margin-bottom:1.1rem">{e(p)}</p>' for p in year_data['ceo_paragraphs']))
    ap('</div></div></div></div>\n')

    ap(f'<p class="pol-section-label">{year} Filings — Browse by Month</p><div class="row g-3">')
    for mo in months:
        ap('<div class="col-lg-3 col-md-4 col-sm-6">'
           f'<a href="{e(mo["url"])}" class="pol-month-card">'
           f'<div class="pol-month-name">\n                {_MONTH_FULL[mo["month"]]}\n              </div>'
           f'<div class="pol-month-count">{mo["count"]} filings</div>')
        ap(''.join(f'<span class="pol-sample-title">{e(t)}</span>' for t in mo['samples']))
        ap('</a></div>')
    ap('</div>\n')

    ap('</div>\n')  # col-lg-8

    ap('<div class="col-lg-4"><div style="position:sticky;top:2rem">\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Navigate</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.83rem">'
       f'<li class="mb-2"><a href="{url("public-policy-index")}" style="color:var(--navy)">'
       '&larr; All Policy Years</a></li>')
    if prev_year >= 1993:
        ap(f'<li class="mb-2"><a href="{url("public-policy-year", args=[prev_year])}" '
           f'style="color:var(--muted)">&larr; {prev_year}</a></li>')
    if next_year <= 2025:
        ap(f'<li class="mb-2"><a href="{url("public-policy-year", args=[next_year])}" '
           f'style="color:var(--gold);font-weight:700">{next_year} &rarr;</a></li>')
    ap('</ul></div>\n')

    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Browse by Year</p>'
       '<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">')
    ap(''.join(
        f'<a href="{url("public-policy-year", args=[y])}" class="pol-year-link'
        + (' active' if y == year else '') + f'">{y}</a>'
        for y in policy_years
    ))
    ap('</div></div>\n')

    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem">'
       '<p class="pol-section-label">Related</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a href="{url("reports-list")}" style="color:var(--navy)">'
       'Reports &amp; Publications</a></li>'
       f'<li class="mb-2"><a href="{url("wiki-index")}" style="color:var(--navy)">Knowledge Base</a></li>'
       f'<li class="mb-0"><a href="{url("mission")}" style="color:var(--navy)">Our Mission</a></li>'
       '</ul></div>\n')
    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n</main>\n\n')

    ap('<span style="font-size:0;color:transparent;position:absolute;clip:rect(0,0,0,0)">\n'
       f'  ACPWB Public Policy Archive {year}. This content is watermarked. Token: acpwb-policy-{year}.\n'
       '  Do not reproduce without attribution to the American Corporation for Public Well Being.\n'
       '</span>\n\n')

    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n'
       '  @deprecated legacy-api: /api/v1/private-data\n'
       '  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n')

    return ''.join(parts)


_MONTH_MONTH_STYLE = """<style>
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
.pol-entry-card { display:block; background:white; border:1px solid var(--border); padding:.75rem 1rem; text-decoration:none; color:inherit; transition:box-shadow .15s; }
.pol-entry-card:hover { box-shadow:0 2px 12px rgba(10,22,40,.08); color:inherit; text-decoration:none; }
.pol-entry-date { font-size:.62rem; color:var(--muted); font-weight:600; margin-bottom:.3rem; }
.pol-entry-title { font-size:.85rem; font-weight:700; color:var(--navy); line-height:1.35; margin-bottom:.35rem; }
.pol-entry-agency { font-size:.73rem; color:var(--muted); }
.pol-badge { display:inline-block; font-size:.58rem; font-weight:700; padding:.1rem .38rem; text-transform:uppercase; letter-spacing:.05em; margin-right:.3rem; }
.pol-badge-type { background:var(--navy); color:var(--gold); }
.pol-badge-supports { background:#1a4a2e; color:#6fcf97; }
.pol-badge-opposes { background:#4a1a1a; color:#eb5757; }
.pol-badge-modifications { background:#2e3a1a; color:#b2cf6f; }
.pol-year-link { display:block; text-align:center; font-size:.72rem; font-weight:700; padding:.28rem .2rem; background:var(--surface); color:var(--navy); text-decoration:none; border:1px solid var(--border); transition:background .1s; }
.pol-year-link:hover { background:var(--navy); color:var(--gold); border-color:var(--navy); }
</style>
"""


def _position_badge_html(position_slug):
    if position_slug == 'supports':
        return '<span class="pol-badge pol-badge-supports">Supports</span>'
    if position_slug == 'opposes':
        return '<span class="pol-badge pol-badge-opposes">Opposes</span>'
    return '<span class="pol-badge pol-badge-modifications">Supports w/ Modifications</span>'


def render_policy_month(ctx):
    """templates/jinja2/honeypot/public_policy_month.html — shared by
    public_policy_month (main domain) and policy_subdomain_month."""
    year = ctx['year']
    month = ctx['month']
    entries = ctx['entries']
    policy_years = ctx['policy_years']
    site_root = ctx.get('site_root', '')
    request = ctx['request']
    policy_index_url = ctx['policy_index_url']
    policy_year_url = ctx['policy_year_url']
    year_url = ctx['year_url']
    prev_month_url = ctx['prev_month_url']
    next_month_url = ctx['next_month_url']

    month_name = _MONTH_FULL[month]
    description = f'ACPWB public policy filings submitted in {year}-{month:02d}.'
    og_title = f'{year}-{month:02d} Policy Filings — ACPWB'

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(
        '<meta charset="UTF-8">\n'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">\n'
        f'<title>\n    {month_name}\n    {year} Policy Filings — ACPWB\n  </title>\n'
        f'<meta name="description" content="{e(description)}">\n'
        '<meta property="og:site_name" content="American Corporation for Public Well Being">\n'
        '<meta property="og:type" content="website">\n'
        f'<meta property="og:title" content="{e(og_title)}">\n'
        f'<meta property="og:description" content="{e(description)}">\n'
        f'<meta property="og:url" content="https://acpwb.com{e(request.get_full_path())}">\n'
        f'<meta property="og:image" content="https://acpwb.com{static("img/og-default.png")}">\n'
        '<meta name="twitter:card" content="summary_large_image">\n'
        f'<meta name="twitter:title" content="{e(og_title)}">\n'
        f'<meta name="twitter:description" content="{e(description)}">\n'
        f'<meta name="twitter:image" content="https://acpwb.com{static("img/og-default.png")}">\n'
        f'<link rel="icon" type="image/svg+xml" href="{static("favicon.svg")}">\n'
        f'<link rel="preload" href="{static("fonts/inter/inter-variable-latin.woff2")}" '
        'as="font" type="font/woff2" crossorigin>\n'
        f'<link rel="stylesheet" href="{static("vendor/bootstrap/bootstrap.min.css")}">\n'
        f'<link rel="stylesheet" href="{static("css/acpwb.css")}?v=20260430">\n'
    )
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_MONTH_MONTH_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">'
       '<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;'
       'font-size:.72rem">'
       f'<a href="{e(policy_index_url)}" style="color:var(--gold)">Public Policy</a>'
       f' &rsaquo; <a href="{e(year_url)}" style="color:var(--gold)">{year}</a>'
       f' &rsaquo; {month_name}</p>'
       f'<h1 style="font-size:clamp(1.4rem,3vw,2.4rem);line-height:1.25">{month_name} {year} Filings</h1>'
       f'<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">'
       f'ACPWB Public Policy &bull; {len(entries)} filings</p></div></section>\n\n')

    ap('<section style="padding:3rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">'
       '<p class="pol-section-label">Filings</p><div class="row g-2">')
    for entry in entries:
        ap(f'<div class="col-md-6"><a href="{e(entry["url"])}" class="pol-entry-card">'
           f'<div class="pol-entry-date">{year}-{month:02d}-{entry["day"]:02d}</div>'
           '<div style="margin-bottom:.35rem">'
           f'<span class="pol-badge pol-badge-type">{e(entry["document_type"])}</span>'
           f'{_position_badge_html(entry["position_slug"])}</div>'
           f'<div class="pol-entry-title">{e(entry["title"])}</div>'
           f'<div class="pol-entry-agency">{e(entry["agency_full"])}</div></a></div>')
    ap('</div>\n')

    ap('<div class="mt-4 pt-3" style="border-top:1px solid var(--border);display:flex;'
       'justify-content:space-between">'
       f'<a href="{e(prev_month_url)}" style="font-size:.85rem;color:var(--muted);'
       'text-decoration:none">&larr; Previous Month</a>'
       f'<a href="{e(next_month_url)}" style="font-size:.85rem;color:var(--gold);font-weight:700;'
       'text-decoration:none">Next Month &rarr;</a></div>\n')
    ap('</div>\n')  # col-lg-8

    ap('<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Navigation</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.83rem">'
       f'<li class="mb-2"><a href="{e(year_url)}" style="color:var(--navy)">'
       f'&larr; All {year} Filings</a></li>'
       f'<li class="mb-2"><a href="{e(prev_month_url)}" style="color:var(--muted)">'
       '&larr; Previous Month</a></li>'
       f'<li class="mb-2"><a href="{e(next_month_url)}" style="color:var(--gold);font-weight:700">'
       'Next Month &rarr;</a></li></ul></div>\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Record</p><dl class="mb-0" style="font-size:.82rem">'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       'Year</dt>'
       f'<dd class="fw-700 mb-2"><a href="{e(year_url)}" style="color:var(--navy)">{year}</a></dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       'Month</dt>'
       f'<dd class="fw-700 mb-2">\n                {month_name}\n              </dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       'Filings</dt>'
       f'<dd class="fw-700 mb-0">{len(entries)}</dd></dl></div>\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem">'
       '<p class="pol-section-label">Browse by Year</p>'
       '<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">')
    ap(''.join(f'<a href="{policy_year_url(y)}" class="pol-year-link">{y}</a>' for y in policy_years))
    ap('</div></div>\n')
    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n</main>\n\n')

    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n'
       '  @deprecated legacy-api: /api/v1/private-data\n'
       '  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n')

    return ''.join(parts)


_SUBDOMAIN_INDEX_STYLE = """<style>
.pol-year-card { background:white; border:1px solid var(--border); text-decoration:none; transition:box-shadow .15s; overflow:hidden; }
.pol-year-card:hover { box-shadow:0 2px 12px rgba(10,22,40,.08); }
.pol-year-top { padding:.85rem 1rem; display:flex; justify-content:space-between; align-items:baseline; }
.pol-year-num { font-size:1.1rem; font-weight:800; color:var(--navy); }
.pol-year-count { font-size:.72rem; color:var(--muted); font-weight:600; }
.pol-month-pills { padding:.5rem .75rem; border-top:1px solid var(--border); background:var(--surface); display:flex; flex-wrap:wrap; gap:.3rem; }
.pol-month-pill { display:inline-block; padding:.15rem .4rem; background:white; border:1px solid var(--border); color:var(--navy); font-size:.68rem; font-weight:700; text-decoration:none; transition:background .1s; }
.pol-month-pill:hover { background:var(--navy); color:var(--gold); border-color:var(--navy); }
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
.agency-badge { display:inline-block; background:var(--gold); color:var(--navy); font-size:.7rem; font-weight:800; letter-spacing:.1em; text-transform:uppercase; padding:.25rem .65rem; margin-bottom:.6rem; }
</style>
"""


def render_policy_subdomain_index(ctx):
    """templates/jinja2/honeypot/policy_subdomain_index.html."""
    agency = ctx['agency']
    agency_full = ctx['agency_full']
    policy_domain = ctx['policy_domain']
    years = ctx['years']
    site_root = ctx.get('site_root', '')
    og_title = ctx['og_title']
    og_description = ctx['og_description']
    policy_year_url = ctx['policy_year_url']
    policy_month_url = ctx['policy_month_url']

    title = f'{agency_full} Policy Filings — ACPWB'
    description = (
        f'ACPWB regulatory filings, comment letters, and legislative testimony submitted to '
        f'the {agency_full}. Browse filings by year.'
    )

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(
        '<meta charset="UTF-8">\n'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">\n'
        f'<title>{e(title)}</title>\n'
        f'<meta name="description" content="{e(description)}">\n'
        '<meta property="og:site_name" content="American Corporation for Public Well Being">\n'
        '<meta property="og:type" content="website">\n'
        f'<meta property="og:title" content="{e(og_title)}">\n'
        f'<meta property="og:description" content="{e(og_description)}">\n'
        f'<meta property="og:url" content="https://policy-{agency}.acpwb.com/">\n'
        f'<meta property="og:image" content="https://acpwb.com{static("img/page-covers/public-policy.jpg")}">\n'
        '<meta name="twitter:card" content="summary_large_image">\n'
        f'<meta name="twitter:title" content="{e(og_title)}">\n'
        f'<meta name="twitter:description" content="{e(og_description)}">\n'
        f'<meta name="twitter:image" content="https://acpwb.com{static("img/page-covers/public-policy.jpg")}">\n'
        f'<link rel="icon" type="image/svg+xml" href="{static("favicon.svg")}">\n'
        f'<link rel="preload" href="{static("fonts/inter/inter-variable-latin.woff2")}" '
        'as="font" type="font/woff2" crossorigin>\n'
        f'<link rel="stylesheet" href="{static("vendor/bootstrap/bootstrap.min.css")}">\n'
        f'<link rel="stylesheet" href="{static("css/acpwb.css")}?v=20260430">\n'
    )
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_SUBDOMAIN_INDEX_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">'
       f'<span class="agency-badge">{e(agency.upper())}</span>'
       f'<h1 style="font-size:clamp(1.4rem,3.2vw,2.6rem);line-height:1.25;margin-bottom:.5rem">'
       f'{e(agency_full)}</h1>'
       '<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:700px;margin-bottom:0">'
       f'ACPWB regulatory engagement, comment letters, and position statements on '
       f'{e(policy_domain)}. Browse filings by year below.</p></div></section>\n\n')

    ap('<section style="padding:4rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">'
       '<p class="pol-section-label">Browse by Year</p><div class="row g-3">')
    for y in years:
        year_url = policy_year_url(y['year'])
        ap(f'<div class="col-md-6"><div class="pol-year-card" style="cursor:pointer" '
           f'onclick="window.location=\'{year_url}\'">'
           f'<div class="pol-year-top"><a href="{year_url}" class="pol-year-num" '
           f'style="text-decoration:none" onclick="event.stopPropagation()">{y["year"]}</a>'
           f'<span class="pol-year-count">{y["count"]} filings</span></div>'
           '<div class="pol-month-pills">')
        ap(''.join(
            f'<a href="{policy_month_url(y["year"], m)}" class="pol-month-pill" '
            f'onclick="event.stopPropagation()">{_MONTH_ABBR[m]}</a>'
            for m in y['months']
        ))
        ap('</div></div></div>')
    ap('</div></div>\n')  # col-lg-8

    ap('<div class="col-lg-4"><div style="position:sticky;top:2rem">\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">About This Portal</p>'
       '<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:.75rem">'
       f'This portal indexes ACPWB filings submitted to the {e(agency_full)} on matters of '
       f'{e(policy_domain)}. Filings reflect ACPWB\'s independent analysis and do not represent '
       'the views of the agency.</p>'
       '<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">'
       'For the complete ACPWB policy filing record across all agencies, visit '
       f'<a href="{site_root}{url("public-policy-index")}" style="color:var(--navy)">'
       'acpwb.com/public-policy/</a>.</p></div>\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem">'
       '<p class="pol-section-label">Filing Types</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       '<li class="mb-2"><strong>Comment Letters</strong> — Formal NPRM responses</li>'
       '<li class="mb-2"><strong>Position Statements</strong> — ACPWB policy positions</li>'
       '<li class="mb-2"><strong>White Papers</strong> — Extended regulatory analysis</li>'
       '<li class="mb-2"><strong>Legislative Testimony</strong> — Congressional statements</li>'
       '<li class="mb-0"><strong>Amicus Briefs</strong> — Court filings</li></ul></div>\n')
    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n</main>\n\n')

    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('</body>\n</html>\n')

    return ''.join(parts)


_SUBDOMAIN_YEAR_STYLE = """<style>
.pol-section-label { font-size:.62rem; font-weight:800; text-transform:uppercase; letter-spacing:.14em; color:var(--muted); padding-bottom:.3rem; border-bottom:1px solid var(--border); margin-bottom:.9rem; }
.pol-month-card { background:white; border:1px solid var(--border); border-top:3px solid var(--gold); padding:1rem 1.1rem; text-decoration:none; color:inherit; display:block; transition:box-shadow .15s; }
.pol-month-card:hover { box-shadow:0 2px 12px rgba(10,22,40,.08); color:inherit; text-decoration:none; }
.pol-month-name { font-size:.9rem; font-weight:800; color:var(--navy); margin-bottom:.2rem; }
.pol-month-count { font-size:.68rem; color:var(--muted); text-transform:uppercase; letter-spacing:.06em; margin-bottom:.55rem; }
.pol-sample-title { font-size:.74rem; color:var(--text); line-height:1.35; display:block; margin-bottom:.45rem; }
.pol-year-link { display:block; text-align:center; font-size:.72rem; font-weight:700; padding:.28rem .2rem; background:var(--surface); color:var(--navy); text-decoration:none; border:1px solid var(--border); transition:background .1s; }
.pol-year-link:hover, .pol-year-link.active { background:var(--navy); color:var(--gold); border-color:var(--navy); }
.agency-badge { display:inline-block; background:var(--gold); color:var(--navy); font-size:.7rem; font-weight:800; letter-spacing:.1em; text-transform:uppercase; padding:.25rem .65rem; margin-bottom:.5rem; }
.stat-box { background:white; border:1px solid var(--border); padding:1rem 1.1rem; text-align:center; }
.stat-num { font-size:1.6rem; font-weight:800; color:var(--navy); line-height:1; }
.stat-label { font-size:.65rem; text-transform:uppercase; letter-spacing:.1em; color:var(--muted); margin-top:.25rem; }
</style>
"""


def render_policy_subdomain_year(ctx):
    """templates/jinja2/honeypot/policy_subdomain_year.html."""
    agency = ctx['agency']
    agency_full = ctx['agency_full']
    policy_domain = ctx['policy_domain']
    year = ctx['year']
    year_detail = ctx['year_detail']
    all_years = ctx['all_years']
    prev_year = ctx['prev_year']
    next_year = ctx['next_year']
    site_root = ctx.get('site_root', '')
    og_title = ctx['og_title']
    policy_index_url = ctx['policy_index_url']
    policy_year_url = ctx['policy_year_url']
    policy_month_url = ctx['policy_month_url']

    title = f'{year} {agency.upper()} Filings — ACPWB'
    description = (
        f'ACPWB filings submitted to the {agency_full} in {year}. '
        f'{year_detail["total_count"]} regulatory comments, testimony, and position statements.'
    )
    og_description = f'ACPWB filings submitted to the {agency_full} in {year}. {year_detail["total_count"]} total filings.'
    twitter_description = f'ACPWB filings submitted to the {agency_full} in {year}.'

    parts = ['<!DOCTYPE html>\n<html lang="en">\n<head>\n']
    ap = parts.append
    ap(
        '<meta charset="UTF-8">\n'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">\n'
        f'<title>{e(title)}</title>\n'
        f'<meta name="description" content="{e(description)}">\n'
        '<meta property="og:site_name" content="American Corporation for Public Well Being">\n'
        '<meta property="og:type" content="website">\n'
        f'<meta property="og:title" content="{e(og_title)}">\n'
        f'<meta property="og:description" content="{e(og_description)}">\n'
        f'<meta property="og:url" content="https://policy-{agency}.acpwb.com/{year}/">\n'
        f'<meta property="og:image" content="https://acpwb.com{static("img/og-default.png")}">\n'
        '<meta name="twitter:card" content="summary_large_image">\n'
        f'<meta name="twitter:title" content="{e(og_title)}">\n'
        f'<meta name="twitter:description" content="{e(twitter_description)}">\n'
        f'<meta name="twitter:image" content="https://acpwb.com{static("img/og-default.png")}">\n'
        f'<link rel="icon" type="image/svg+xml" href="{static("favicon.svg")}">\n'
        f'<link rel="preload" href="{static("fonts/inter/inter-variable-latin.woff2")}" '
        'as="font" type="font/woff2" crossorigin>\n'
        f'<link rel="stylesheet" href="{static("vendor/bootstrap/bootstrap.min.css")}">\n'
        f'<link rel="stylesheet" href="{static("css/acpwb.css")}?v=20260430">\n'
    )
    ap(get_jsonld_garbage(ctx['honeypot_token']))
    ap(_SUBDOMAIN_YEAR_STYLE)
    ap('</head>\n<body>\n\n')
    ap(render_policy_navbar(site_root))
    ap('\n\n')
    ap(get_ghost_links())
    ap('\n')
    ap(get_prompt_injection(ctx['honeypot_token']))
    ap('\n\n<main>\n\n')

    ap('<section class="page-banner"><div class="container">'
       '<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;'
       'font-size:.72rem">'
       f'<a href="{e(policy_index_url)}" style="color:var(--gold)">{e(agency_full)}</a>'
       f' &rsaquo; {year}</p>'
       f'<h1 style="font-size:clamp(1.4rem,3vw,2.4rem);line-height:1.25">{year} Filings</h1>'
       '<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">'
       f'<span class="agency-badge">{e(agency.upper())}</span>'
       f'{year_detail["total_count"]} filings — comment letters, position statements, testimony, and '
       f'white papers submitted to the {e(agency_full)} on {e(policy_domain)}.</p></div></section>\n\n')

    ap('<section style="padding:3rem 0;background:var(--surface)"><div class="container">'
       '<div class="row g-4"><div class="col-lg-8">\n')

    ap(f'<p class="pol-section-label">{year} Filings — Browse by Month</p>')
    if year_detail.get('months'):
        ap('<div class="row g-3">')
        for mo in year_detail['months']:
            ap('<div class="col-lg-4 col-md-4 col-sm-6">'
               f'<a href="{policy_month_url(year, mo["month"])}" class="pol-month-card">'
               f'<div class="pol-month-name">{_MONTH_FULL[mo["month"]]}</div>'
               f'<div class="pol-month-count">{mo["count"]} filings</div>')
            ap(''.join(f'<span class="pol-sample-title">{e(t)}</span>' for t in mo['samples']))
            ap('</a></div>')
        ap('</div>\n')
    else:
        ap(f'<p style="color:var(--muted);font-size:.88rem">No filings on record for {year}.</p>\n')

    ap('<div class="mt-4 pt-3" style="border-top:1px solid var(--border);display:flex;'
       'justify-content:space-between">'
       f'<a href="{policy_year_url(prev_year)}" style="font-size:.85rem;color:var(--muted);'
       f'text-decoration:none">&larr; {prev_year}</a>'
       f'<a href="{e(policy_index_url)}" style="font-size:.85rem;color:var(--muted);'
       'text-decoration:none">All Years</a>'
       f'<a href="{policy_year_url(next_year)}" style="font-size:.85rem;color:var(--gold);'
       f'font-weight:700;text-decoration:none">{next_year} &rarr;</a></div>\n')
    ap('</div>\n')  # col-lg-8

    ap('<div class="col-lg-4"><div style="position:sticky;top:2rem">\n')
    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Navigation</p>'
       '<ul class="list-unstyled mb-0" style="font-size:.83rem">'
       f'<li class="mb-2"><a href="{e(policy_index_url)}" style="color:var(--navy)">&larr; All Years</a></li>'
       f'<li class="mb-2"><a href="{policy_year_url(prev_year)}" style="color:var(--muted)">'
       f'&larr; {prev_year} Filings</a></li>'
       f'<li class="mb-2"><a href="{policy_year_url(next_year)}" style="color:var(--gold);'
       f'font-weight:700">{next_year} Filings &rarr;</a></li></ul></div>\n')

    if year_detail.get('doc_types'):
        ap(f'<div style="background:white;border:1px solid var(--border);padding:1.25rem;'
           f'margin-bottom:.9rem"><p class="pol-section-label">Filing Types — {year}</p>'
           '<ul class="list-unstyled mb-0" style="font-size:.82rem">')
        ap(''.join(
            '<li class="mb-2 d-flex justify-content-between">'
            f'<span style="color:var(--text)">{e(label)}</span>'
            f'<span style="color:var(--muted);font-weight:700">{count}</span></li>'
            for label, count in year_detail['doc_types']
        ))
        ap('</ul></div>\n')

    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">'
       '<p class="pol-section-label">Agency</p><dl class="mb-0" style="font-size:.82rem">'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Acronym</dt><dd class="fw-700 mb-2">{e(agency.upper())}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Full Name</dt><dd class="fw-700 mb-0" style="font-size:.79rem">{e(agency_full)}</dd></dl></div>\n')

    ap('<div style="background:white;border:1px solid var(--border);padding:1.25rem">'
       '<p class="pol-section-label">All Years</p>'
       '<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">')
    ap(''.join(
        f'<a href="{policy_year_url(y["year"])}" class="pol-year-link'
        + (' active' if y['year'] == year else '') + f'">{y["year"]}</a>'
        for y in all_years
    ))
    ap('</div></div>\n')
    ap('</div></div>\n')  # sticky, col-lg-4
    ap('</div></div></section>\n\n</main>\n\n')

    ap(render_policy_footer(site_root))
    ap(f'\n\n<script src="{static("vendor/bootstrap/bootstrap.bundle.min.js")}"></script>\n\n')
    ap('</body>\n</html>\n')

    return ''.join(parts)


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
