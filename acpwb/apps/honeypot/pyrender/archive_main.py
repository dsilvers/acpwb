"""
Hand-written Python builders for the main-domain (Django-backend) archive
templates: templates/honeypot/archive.html, archive_compliance.html,
archive_minutes.html.

Each of those template files also contains an `{% if on_archive_subdomain %}`
branch, but it's unreachable dead code in practice — archive_trap() picks a
completely different (Jinja2/era) template when on_sub is True, per
apps/honeypot/views.py. So only the `{% else %}` / main-domain branch is
reproduced here; the era/subdomain equivalents live in archive_era.py.

`ctx` is the same dict archive_trap() already builds for render(): the
content-variant dict (title/org/findings/etc.) merged with the shared
page-level fields (year/month/day/related_paths/next_entry_url/etc.).
"""
from apps.core.htmlgen import escape as e
from apps.core.htmlgen import get_archive_seal, render_pres_card, truncatechars


def _related_reports_html(c):
    if not c.get('related_reports'):
        return ''
    items = ''.join(
        f'<div style="background:white;border:1px solid var(--border);border-left:3px solid var(--gold);'
        f'padding:.75rem 1rem;margin-bottom:.75rem">'
        f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;'
        f'color:var(--muted);margin-bottom:.2rem">'
        f'{e(r.get("category", ""))} &bull; {e(str(r.get("file_type", "")).upper())} &bull; '
        f'{e(r.get("pub_date_display", ""))}</div>'
        f'<a href="{e(r.get("detail_url", "#"))}" style="font-size:.9rem;font-weight:700;'
        f'color:var(--navy);text-decoration:none">{e(r.get("title", ""))}</a>'
        f'<p class="small text-muted mb-0" style="font-size:.78rem;margin-top:.25rem">'
        f'{e(truncatechars(r.get("summary") or "", 160))}</p></div>'
        for r in c['related_reports']
    )
    return (
        '<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
        'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem">'
        'Referenced Research &amp; Publications</h5>' + items + '</div><hr class="gold-divider">'
    )


def _sidebar_related_docs_html(c, rid):
    if not c['related_docs']:
        return ''
    items = ''.join(
        f'<a href="{e(d["url"])}" style="display:block;background:#f4f6f9;border:1px solid var(--border);'
        f'border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">'
        f'<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">{e(d["date"])} &bull; '
        f'{e(d["phase"])}</div>'
        f'<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">{e(d["label"])}</div></a>'
        for d in c['related_docs']
    )
    return (
        f'<div id="acpwb-archive-{rid}-related-docs" class="acpwb-card mb-4">'
        f'<h6 class="card-title mb-3">Related Documents</h6>{items}</div>'
    )


def _sidebar_related_policy_html(c):
    if not c['related_policy']:
        return ''
    items = ''.join(
        f'<a href="{e(s["url"])}" style="display:block;background:#f4f6f9;border:1px solid var(--border);'
        f'border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">'
        f'<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">{e(s["filing_date"])} &bull; '
        f'{e(s["document_type"])}</div>'
        f'<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">{e(s["title"])}</div>'
        f'<div style="font-size:.65rem;color:var(--muted);margin-top:.15rem">{e(s["agency_acronym"])}</div></a>'
        for s in c['related_policy']
    )
    return f'<div class="acpwb-card mb-4"><h6 class="card-title mb-3">Related Public Policy</h6>{items}</div>'


def _sidebar_related_presentations_html(c):
    if not c['related_presentations']:
        return ''
    items = ''.join(render_pres_card(p, url_prefix='https://acpwb.com') for p in c['related_presentations'])
    return (
        '<div class="acpwb-archive-pres-sidebar acpwb-card mb-4">'
        f'<h6 class="card-title mb-3">Related Presentations</h6>{items}</div>'
    )


def _sidebar_year_browser_html(c, rid):
    items = ''.join(
        f'<a id="acpwb-archive-{rid}-year-link-{y}" href="https://archives-{y}.acpwb.com/" '
        f'aria-label="Browse the {y} archive" '
        f'style="font-size:.72rem;padding:.2rem .45rem;background:#f4f6f9;border:1px solid var(--border);'
        f'color:var(--navy);text-decoration:none">{y}</a>'
        for y in c['archive_years']
    )
    return (
        f'<div id="acpwb-archive-{rid}-year-browser" class="acpwb-card mb-4">'
        f'<h6 class="card-title mb-3">Browse by Year</h6>'
        f'<div style="display:flex;flex-wrap:wrap;gap:.35rem">{items}</div></div>'
    )


def _bulk_hex_script_html(c, rid):
    bulk_hex_js = c.get('bulk_hex_js', [])
    bulk_hex_css = c.get('bulk_hex_css', [])
    css_vars = ''.join(f'  --acpwb-r{i:03d}: {h};\n' for i, h in enumerate(bulk_hex_css))
    js_refs = ''.join(f'var _acpwbRef{i}="{h}";\n' for i, h in enumerate(bulk_hex_js, 1))
    js_funcs = ''.join(
        f'function _acpwbArchiveRecordEntryMetadataLookup_{h}(){{return "{rid}";}}\n'
        for h in bulk_hex_js[:50]
    )
    return (
        f'<style>\n  :root {{\n{css_vars}  }}\n</style>\n'
        f'<script>\n/* ACPWB archive index — {rid} */\n(function(){{\n{js_refs}{js_funcs}}})();\n</script>\n'
    )


def render_archive_default(ctx):
    """templates/honeypot/archive.html, main-domain branch."""
    c = ctx
    rid = c['record_id']
    parts = []
    ap = parts.append

    ap('<section class="page-banner"><div class="container">')
    ap(f'<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;'
       f'letter-spacing:.18em;font-size:.75rem">'
       f'<a href="/archive/" style="color:var(--gold)">Archive</a>'
       f' &rsaquo; <a href="{e(c["year_url"])}" style="color:var(--gold)">{c["year"]}</a>'
       f' &rsaquo; <a href="{e(c["month_url"])}" style="color:var(--gold)">{c["month"]:02d}</a>'
       f' &rsaquo; {c["day"]:02d}</p>')
    ap(f'<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">{e(c["title"])}</h1>')
    ap(f'<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">'
       f'{e(c["industry"])} &bull; {e(c["org"])} &bull; {e(c["phase"].capitalize())} phase'
       f' &bull; Archived {c["year"]}-{c["month"]:02d}-{c["day"]:02d}</p>')
    ap('</div></section>')

    ap('<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">')
    ap(f'<div id="acpwb-archive-{rid}-primary-content-col" class="col-lg-8">')

    exec_mb = '.75rem' if c.get('exec_bullets') else '0'
    ap(f'<div id="acpwb-archive-{rid}-executive-summary-callout" '
       f'style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);'
       f'padding:1.25rem 1.5rem;margin-bottom:2rem">'
       f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;'
       f'color:var(--gold);margin-bottom:.5rem">Executive Summary</div>'
       f'<p style="font-size:.88rem;color:var(--navy);margin-bottom:{exec_mb};font-weight:500;line-height:1.6">'
       f'This archive entry documents ACPWB\'s <strong>{e(c["phase"])}</strong> phase engagement with '
       f'<strong>{e(c["org"])}</strong> in the <strong>{e(c["industry"])}</strong> sector.'
       f' The record ID <code style="font-size:.8rem;background:#f4f6f9;padding:.1rem .35rem">{e(rid)}</code>'
       f' uniquely identifies this documentation set within ACPWB\'s institutional archive.</p>')
    if c.get('exec_bullets'):
        ap(f'<ul id="acpwb-archive-{rid}-exec-summary-bullets" style="padding-left:1.25rem;margin-bottom:0">')
        ap(''.join(
            f'<li style="font-size:.83rem;color:#333;margin-bottom:.45rem;line-height:1.6">{e(b)}</li>'
            for b in c['exec_bullets']
        ))
        ap('</ul>')
    ap('</div>')

    if c.get('findings'):
        ap(f'<div id="acpwb-archive-{rid}-key-findings-block" class="mb-4">'
           f'<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           f'font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">'
           f'Key Findings</h5>'
           f'<ul id="acpwb-archive-{rid}-key-findings-list" style="padding-left:1.25rem;margin-bottom:0">')
        n_findings = len(c['findings'])
        ap(''.join(
            f'<li id="acpwb-archive-{rid}-finding-{i}" data-ref="{e(f["ref"])}" '
            f'aria-label="Finding {i} of {n_findings}: {e(f["text"])}" '
            f'style="font-size:.88rem;color:#333;margin-bottom:.7rem;line-height:1.6">{e(f["text"])}</li>'
            for i, f in enumerate(c['findings'], 1)
        ))
        ap('</ul></div>')

    if c.get('engagement_team'):
        ap(f'<div id="acpwb-archive-{rid}-engagement-team" class="mb-4">'
           f'<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           f'font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">'
           f'Engagement Team</h5><table style="width:100%;border-collapse:collapse;font-size:.83rem">')
        ap(''.join(
            f'<tr style="border-top:1px solid var(--border)">'
            f'<td style="padding:.4rem .8rem .4rem 0;white-space:nowrap">'
            f'<div style="font-weight:700;color:var(--navy)">{e(m["name"])}</div>'
            f'<div style="font-size:.72rem;color:var(--muted)">'
            f'<a href="mailto:{e(m["email"])}" style="color:inherit">{e(m["email"])}</a></div></td>'
            f'<td style="padding:.4rem 0;color:var(--muted)">{e(m["title"])}</td></tr>'
            for m in c['engagement_team']
        ))
        ap('</table></div>')

    ap(f'<div id="acpwb-archive-{rid}-body-content" class="wiki-content mb-4">')
    ap(''.join(f'<p data-doc="{e(p["ref"])}">{e(p["text"])}</p>' for p in c.get('paragraphs', [])))
    ap('</div>')

    if c.get('metric_rows'):
        ap(f'<div id="acpwb-archive-{rid}-engagement-metrics-section" class="mb-4">'
           f'<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           f'font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">'
           f'Engagement Metrics</h5><div style="overflow-x:auto">'
           f'<table id="acpwb-archive-{rid}-engagement-metrics-table" '
           f'style="width:100%;border-collapse:collapse;background:white;border:1px solid var(--border);'
           f'font-size:.83rem"><thead><tr style="background:var(--navy);color:var(--gold)">'
           '<th style="padding:.6rem .9rem;text-align:left;font-size:.68rem;font-weight:700;'
           'text-transform:uppercase;letter-spacing:.06em;white-space:nowrap">Metric</th>'
           '<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;'
           'text-transform:uppercase;letter-spacing:.06em">Baseline</th>'
           '<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;'
           'text-transform:uppercase;letter-spacing:.06em">Current</th>'
           '<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;'
           'text-transform:uppercase;letter-spacing:.06em">Change</th></tr></thead><tbody>')
        ap(''.join(
            f'<tr id="acpwb-archive-{rid}-metric-row-{i}" '
            f'aria-label="Metric: {e(r["name"])}, baseline {e(r["baseline"])}, current {e(r["current"])}, '
            f'change {e(r["delta"])}" style="border-top:1px solid var(--border)">'
            f'<td style="padding:.5rem .9rem;color:var(--navy);font-weight:600">{e(r["name"])}</td>'
            f'<td style="padding:.5rem .9rem;text-align:right;color:var(--muted)">{e(r["baseline"])}</td>'
            f'<td style="padding:.5rem .9rem;text-align:right;font-weight:700">{e(r["current"])}</td>'
            f'<td style="padding:.5rem .9rem;text-align:right;font-weight:700;'
            f'color:{"#15803d" if r["positive"] else "#dc2626"}">{e(r["delta"])}</td></tr>'
            for i, r in enumerate(c['metric_rows'], 1)
        ))
        ap('</tbody></table></div></div>')

    if c.get('percentile_table'):
        ap(f'<div id="acpwb-archive-{rid}-benchmark-percentile-section" class="mb-4">'
           f'<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           f'font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">'
           f'Market Benchmark — {e(c.get("peer_group", ""))}</h5><div style="overflow-x:auto">'
           '<table style="width:100%;border-collapse:collapse;background:white;border:1px solid var(--border);'
           'font-size:.82rem"><thead><tr style="background:var(--navy);color:var(--gold)">'
           + ''.join(
               f'<th style="padding:.55rem .9rem;text-align:{"left" if h == "Metric" else "right"};'
               f'font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.06em">{h}</th>'
               for h in ['Metric', 'P10', 'P25', 'P33', 'P50', 'P67', 'P75', 'P90', 'P95']
           ) + '</tr></thead><tbody>')
        ap(''.join(
            f'<tr style="border-top:1px solid var(--border)">'
            f'<td style="padding:.45rem .9rem;color:var(--navy);font-weight:600">{e(r["metric"])}</td>'
            + ''.join(
                f'<td style="padding:.45rem .9rem;text-align:right;'
                f'{"font-weight:700" if k == "p50" else "color:var(--muted)"}">{e(r[k])}</td>'
                for k in ['p10', 'p25', 'p33', 'p50', 'p67', 'p75', 'p90', 'p95']
            ) + '</tr>'
            for r in c['percentile_table']
        ))
        ap('</tbody></table></div></div>')

    ap('<hr class="gold-divider">')
    ap(_related_reports_html(c))

    ap('<div style="background:var(--surface);border:1px solid var(--border);padding:.85rem 1rem;'
       'margin-bottom:1.25rem;display:flex;align-items:center;justify-content:space-between;'
       'flex-wrap:wrap;gap:.5rem"><div>'
       '<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;'
       'color:var(--muted)">Raw Data Export</div>'
       '<div style="font-size:.8rem;color:var(--navy);font-weight:600">'
       'Download the underlying dataset for this archive entry</div></div>'
       f'<a href="{e(c["export_csv_url"])}" style="font-size:.72rem;font-weight:700;padding:.3rem .85rem;'
       'background:var(--navy);color:var(--gold);text-decoration:none;white-space:nowrap">'
       '&#x2193; Download CSV</a></div>')
    ap(f'<a href="{e(c["next_entry_url"])}" style="display:block;background:var(--navy);color:white;'
       f'padding:1rem 1.5rem;text-decoration:none;font-weight:700;font-size:.9rem;margin-bottom:2rem">'
       f'Continue Reading: Next in Series &rarr;</a>')

    ap('<div class="mt-2"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.85rem;margin-bottom:1rem">Related Archive Entries</h5>'
       '<div class="row g-2">')
    ap(''.join(
        f'<div class="col-md-6"><a href="{e(ent["url"])}" style="display:block;background:white;'
        f'border:1px solid var(--border);padding:.75rem 1rem;text-decoration:none">'
        f'<div style="font-size:.62rem;color:var(--muted);font-weight:600;margin-bottom:.2rem">'
        f'{e(ent["date"])}</div>'
        f'<div style="font-size:.8rem;color:var(--navy);font-weight:600;line-height:1.3">{e(ent["label"])}'
        f'</div></a></div>'
        for ent in c['related_paths']
    ))
    ap('</div></div>')

    if c['cross_year_reports']:
        ap('<div class="mt-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
           'letter-spacing:.06em;font-size:.85rem;margin-bottom:1rem">'
           'Related Archive Reports — Other Years</h5><div class="row g-2">')
        ap(''.join(
            f'<div class="col-md-6"><a href="{e(ent["url"])}" style="display:block;background:white;'
            f'border:1px solid var(--border);border-left:3px solid var(--gold);padding:.75rem 1rem;'
            f'text-decoration:none">'
            f'<div style="font-size:.6rem;color:var(--gold);font-weight:800;text-transform:uppercase;'
            f'letter-spacing:.08em;margin-bottom:.2rem">{ent["year"]} Archive</div>'
            f'<div style="font-size:.8rem;color:var(--navy);font-weight:600;line-height:1.3">{e(ent["label"])}</div>'
            f'<div style="font-size:.62rem;color:var(--muted);margin-top:.2rem">{e(ent["date"])}</div></a></div>'
            for ent in c['cross_year_reports']
        ))
        ap('</div></div>')

    if c.get('revisions'):
        ap(f'<div id="acpwb-archive-{rid}-revision-history" class="mb-4 mt-4">'
           '<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           'font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">'
           'Revision History</h5><table style="width:100%;border-collapse:collapse;font-size:.8rem">'
           '<thead><tr style="color:var(--muted)">'
           '<th style="padding:.3rem .6rem .3rem 0;font-size:.65rem;text-transform:uppercase;'
           'letter-spacing:.06em;text-align:left;white-space:nowrap">Version</th>'
           '<th style="padding:.3rem .6rem;font-size:.65rem;text-transform:uppercase;letter-spacing:.06em;'
           'text-align:left;white-space:nowrap">Date</th>'
           '<th style="padding:.3rem 0 .3rem .6rem;font-size:.65rem;text-transform:uppercase;'
           'letter-spacing:.06em;text-align:left">Description</th></tr></thead><tbody>')
        ap(''.join(
            f'<tr style="border-top:1px solid var(--border)">'
            f'<td style="padding:.38rem .6rem .38rem 0;font-weight:700;color:var(--navy);white-space:nowrap;'
            f'font-family:monospace;font-size:.75rem">{e(r["version"])}</td>'
            f'<td style="padding:.38rem .6rem;color:var(--muted);white-space:nowrap">{e(r["date"])}</td>'
            f'<td style="padding:.38rem 0 .38rem .6rem;color:#444;line-height:1.45">{e(r["description"])} '
            f'<span style="color:var(--muted)">— {e(r["author"])} &lt;'
            f'<a href="mailto:{e(r["author_email"])}" style="color:inherit">{e(r["author_email"])}</a>&gt;'
            f'</span></td></tr>'
            for r in c['revisions']
        ))
        ap('</tbody></table></div>')

    if c.get('footnotes'):
        ap(f'<div id="acpwb-archive-{rid}-footnotes" class="mb-4" '
           'style="border-top:1px solid var(--border);padding-top:1rem">'
           '<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;'
           'font-size:.8rem;margin-bottom:.75rem">Sources &amp; Notes</h5>'
           '<ol style="padding-left:1.25rem;margin-bottom:0">')
        ap(''.join(
            f'<li id="acpwb-archive-{rid}-fn-{fn["num"]}" '
            f'style="font-size:.72rem;color:var(--muted);margin-bottom:.4rem;line-height:1.5">{e(fn["text"])}</li>'
            for fn in c['footnotes']
        ))
        ap('</ol></div>')

    ap(f'<div class="mt-4 pt-3" style="border-top:1px solid var(--border)">'
       f'<a href="{e(c["prev_entry_url"])}" style="font-size:.85rem;color:var(--muted);'
       f'text-decoration:none">&larr; Previous in Series</a></div>')

    ap('<div class="mt-5 pt-3 text-center" style="border-top:1px solid var(--border);opacity:.75">')
    ap(get_archive_seal(c['year'], rid))
    ap('</div>')

    ap('</div>')  # end col-lg-8

    ap(f'<div id="acpwb-archive-{rid}-sidebar-col" class="col-lg-4 d-none d-lg-block">'
       '<div style="position:sticky;top:2rem">')

    ap(f'<div id="acpwb-archive-{rid}-sidebar-record-card" class="acpwb-card mb-4">'
       '<h6 class="card-title mb-2">Archive Record</h6><dl class="mb-0" style="font-size:.82rem">'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>'
       f'<dd class="fw-700 mb-2">{c["year"]}-{c["month"]:02d}-{c["day"]:02d}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Phase</dt>'
       f'<dd class="fw-700 mb-2" style="text-transform:capitalize">{e(c["phase"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Sector</dt>'
       f'<dd class="fw-700 mb-2">{e(c["industry"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Organization</dt><dd class="fw-700 mb-2">{e(c["org"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Engagement Code</dt><dd class="mb-2" style="font-family:monospace;font-size:.75rem;'
       f'color:var(--muted)">{e(c["eng_code"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Version</dt>'
       f'<dd class="mb-2" style="font-family:monospace;font-size:.75rem;color:var(--muted)">'
       f'{e(c["doc_version"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Pages</dt>'
       f'<dd class="mb-2" style="font-size:.8rem;color:var(--muted)">'
       f'{c["page_count"]} ({c["file_size_kb"]} KB)</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Distribution</dt><dd class="mb-2" style="font-size:.72rem;color:var(--muted);line-height:1.4">'
       f'{e(c["distribution"])}</dd>'
       '<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">'
       f'Record ID</dt><dd class="mb-0 text-muted" style="font-family:monospace;font-size:.75rem">'
       f'{e(rid)}</dd></dl></div>')

    ap(f'<div id="acpwb-archive-{rid}-sidebar-nav-card" class="acpwb-card mb-4">'
       '<h6 class="card-title mb-3">Archive Navigation</h6><ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-year" href="{e(c["year_url"])}" '
       f'aria-label="Browse all {c["year"]} archive records" style="color:var(--navy)">'
       f'&#8592; All {c["year"]} Records</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-month" href="{e(c["month_url"])}" '
       f'aria-label="Browse {c["year"]} month {c["month"]} archive records" style="color:var(--navy)">'
       f'&#8592; {c["year"]}/{c["month"]:02d} Records</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-prev" href="{e(c["prev_entry_url"])}" '
       f'aria-label="Navigate to the previous entry in the archive series" style="color:var(--muted)">'
       f'&#8592; Previous Entry</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-next" href="{e(c["next_entry_url"])}" '
       f'aria-label="Navigate to the next entry in the {c["year"]} archive series" '
       f'style="color:var(--gold);font-weight:700">Next in Series &rarr;</a></li></ul></div>')

    ap(_sidebar_related_docs_html(c, rid))
    ap(_sidebar_related_policy_html(c))
    ap(_sidebar_related_presentations_html(c))
    ap(_sidebar_year_browser_html(c, rid))

    ap(f'<div id="acpwb-archive-{rid}-research-division-card" class="acpwb-card">'
       '<h6 class="card-title mb-3">Research Division</h6>'
       '<p class="small text-muted mb-2" style="font-size:.8rem">'
       'ACPWB\'s document archive spans our full operational history from 1985 to present.</p>'
       '<a href="/reports/" style="font-size:.8rem;color:var(--gold);font-weight:700;'
       'text-decoration:none">Browse Research Reports &rarr;</a></div>')

    ap('</div></div>')  # end sticky, sidebar col
    ap('</div></div></section>')  # end row, container, section

    ap(_bulk_hex_script_html(c, rid))
    return ''.join(parts)


def _risk_badge(risk, fallback_label='INFO'):
    """The real template uses a DIFFERENT fallback label in two places:
    "INFO" in the findings summary table, "INFORMATIONAL" in the detailed
    findings cards — same colors, different text."""
    colors = {
        'HIGH': ('#fee2e2', '#991b1b', '#fca5a5', 'HIGH'),
        'MEDIUM': ('#fef3c7', '#92400e', '#fcd34d', 'MEDIUM'),
        'LOW': ('#fef9c3', '#854d0e', '#fde047', 'LOW'),
    }
    bg, fg, border, label = colors.get(risk, ('#f1f5f9', '#475569', '#cbd5e1', fallback_label))
    return (
        f'<span style="font-size:.62rem;font-weight:700;padding:.15rem .4rem;background:{bg};'
        f'color:{fg};border:1px solid {border}">{label}</span>'
    )


def _year_browser_plain_html(c):
    """Compliance/minutes' Browse-by-Year sidebar: no id/aria-label per
    link (unlike the default variant's), current year bolded."""
    items = ''.join(
        f'<a href="https://archives-{y}.acpwb.com/" '
        f'style="font-size:.68rem;padding:.2rem .4rem;border:1px solid var(--border);'
        f'text-decoration:none;color:inherit'
        + (';font-weight:800;border-color:var(--gold);color:var(--navy)' if y == c['year'] else '')
        + f'">{y}</a>'
        for y in c['all_years']
    )
    return (
        '<div class="acpwb-card"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
        'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Browse by Year</div>'
        f'<div style="display:flex;flex-wrap:wrap;gap:.3rem">{items}</div></div>'
    )


def _status_badge(status):
    colors = {
        'OPEN': ('#fee2e2', '#991b1b', '#fca5a5', 'OPEN'),
        'IN PROGRESS': ('#fef3c7', '#92400e', '#fcd34d', 'IN PROGRESS'),
        'REMEDIATED': ('#dcfce7', '#15803d', '#86efac', 'REMEDIATED'),
        'DEFERRED': ('#f1f5f9', '#475569', '#cbd5e1', 'DEFERRED'),
    }
    bg, fg, border, label = colors.get(status, ('#eff6ff', '#1d4ed8', '#93c5fd', 'MONITORING'))
    return (
        f'<span style="font-size:.62rem;font-weight:700;padding:.15rem .4rem;background:{bg};'
        f'color:{fg};border:1px solid {border}">{label}</span>'
    )


def render_compliance_default(ctx):
    """templates/honeypot/archive_compliance.html, main-domain branch."""
    c = ctx
    rid = c['record_id']
    parts = []
    ap = parts.append

    ap('<section class="page-banner"><div class="container">')
    ap(f'<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;'
       f'letter-spacing:.18em;font-size:.75rem">'
       f'<a href="/archive/" style="color:var(--gold)">Archive</a>'
       f' &rsaquo; <a href="{e(c["year_url"])}" style="color:var(--gold)">{c["year"]}</a>'
       f' &rsaquo; <a href="{e(c["month_url"])}" style="color:var(--gold)">{c["month"]:02d}</a>'
       f' &rsaquo; {c["day"]:02d}</p>')
    ap(f'<p style="color:rgba(255,255,255,.6);font-size:.72rem;margin-bottom:.4rem;letter-spacing:.1em;'
       f'text-transform:uppercase">Compliance Review &bull; {e(c["audit_ref"])}</p>')
    ap(f'<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">{e(c["title"])}</h1>')
    ap(f'<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">'
       f'{e(c["industry"])} &bull; {e(c["org"])} &bull; Archived {e(c["date_str"])}</p>')
    ap('</div></section>')

    ap('<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">')
    ap('<div class="col-lg-8">')

    ap(f'<div id="acpwb-compliance-{rid}-header" style="background:white;border:1px solid var(--border);'
       f'border-left:4px solid var(--gold);padding:1.25rem 1.5rem;margin-bottom:2rem">'
       f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;'
       f'color:var(--gold);margin-bottom:.5rem">Document Information</div>'
       f'<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.83rem;'
       f'margin-bottom:.75rem">'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Client</span><br><strong>{e(c["org"])}</strong></div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Industry</span><br><strong>{e(c["industry"])}</strong></div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Audit Ref</span><br><code style="font-size:.75rem">{e(c["audit_ref"])}</code></div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Version</span><br><code style="font-size:.75rem">{e(c["doc_version"])}</code></div></div>'
       f'<div style="font-size:.8rem;color:#555;padding-top:.5rem;border-top:1px solid var(--border)">'
       f'Filed by <strong>{e(c["assessor"])}</strong>, {e(c["assessor_title"])} &mdash; '
       f'<a href="mailto:{e(c["assessor_email"])}" style="color:var(--navy)">{e(c["assessor_email"])}</a></div>')
    if c.get('frameworks_cited'):
        ap('<div style="margin-top:.6rem;display:flex;flex-wrap:wrap;gap:.3rem">')
        ap(''.join(
            f'<span style="font-size:.62rem;padding:.15rem .45rem;background:#f4f6f9;border:1px solid #dde1e8;'
            f'color:#444">{e(fw)}</span>'
            for fw in c['frameworks_cited']
        ))
        ap('</div>')
    ap('</div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       f'border-bottom:2px solid var(--gold)">1. Engagement Scope</h5>'
       f'<p style="font-size:.88rem;line-height:1.7;color:#333">{e(c["scope_para"])}</p></div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       f'border-bottom:2px solid var(--gold)">2. Methodology</h5>'
       f'<p style="font-size:.88rem;line-height:1.7;color:#333">{e(c["method_para"])}</p></div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       'border-bottom:2px solid var(--gold)">3. Findings Summary</h5>'
       '<div style="overflow-x:auto"><table class="dash-table w-100"><thead><tr>'
       '<th>Finding ID</th><th>Risk</th><th>Status</th><th>Owner</th><th>Due</th><th>Description</th>'
       '</tr></thead><tbody>')
    ap(''.join(
        f'<tr><td style="font-family:monospace;font-size:.72rem;white-space:nowrap">{e(f["id"])}</td>'
        f'<td style="white-space:nowrap">{_risk_badge(f["risk"])}</td>'
        f'<td style="white-space:nowrap">{_status_badge(f["status"])}</td>'
        f'<td style="font-size:.75rem">{e(f["owner"])}</td>'
        f'<td style="font-size:.72rem;white-space:nowrap">{e(f["due_date"])}</td>'
        f'<td style="font-size:.78rem;max-width:240px">{e(f["description"])}</td></tr>'
        for f in c['findings']
    ))
    ap('</tbody></table></div></div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       'border-bottom:2px solid var(--gold)">4. Detailed Findings</h5>')
    ap(''.join(
        f'<div style="background:white;border:1px solid var(--border);margin-bottom:1rem;padding:1rem 1.2rem">'
        f'<div style="display:flex;align-items:center;gap:.5rem;margin-bottom:.55rem">'
        f'<code style="font-size:.72rem;color:#666">{e(f["id"])}</code>'
        f'{_risk_badge(f["risk"], fallback_label="INFORMATIONAL")}'
        + (_status_badge('REMEDIATED') if f['status'] == 'REMEDIATED' else '')
        + '</div>'
        f'<div style="margin-bottom:.5rem">'
        f'<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;'
        f'color:#888;margin-bottom:.2rem">Observation</div>'
        f'<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:0">{e(f["description"])}</p></div>'
        f'<div style="margin-bottom:.5rem;padding-top:.5rem;border-top:1px solid var(--border)">'
        f'<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;'
        f'color:#888;margin-bottom:.2rem">Corrective Action Required</div>'
        f'<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:0">{e(f["corrective_action"])}</p></div>'
        f'<div style="padding-top:.5rem;border-top:1px solid var(--border)">'
        f'<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;'
        f'color:#888;margin-bottom:.2rem">Management Response</div>'
        f'<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:.4rem">{e(f["mgmt_response"])}</p>'
        f'<div style="font-size:.72rem;color:#666">Owner: <strong>{e(f["owner"])}</strong> &bull; '
        f'Due: <strong>{e(f["due_date"])}</strong></div></div></div>'
        for f in c['findings']
    ))
    ap('</div>')

    ap(f'<div style="margin-top:1.5rem;padding-top:1.25rem;border-top:1px solid var(--border);'
       f'display:flex;justify-content:space-between">'
       f'<a href="{e(c["prev_entry_url"])}" style="font-size:.85rem;color:var(--muted);'
       f'text-decoration:none">&larr; Previous</a>'
       f'<a href="{e(c["next_entry_url"])}" style="font-size:.85rem;color:var(--navy);font-weight:600;'
       f'text-decoration:none">Next &rarr;</a></div>')

    ap('</div>')  # end col-lg-8

    ap('<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">')
    ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
       'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Audit Record</div>'
       '<dl class="mb-0" style="font-size:.82rem">'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Audit Ref</dt>'
       f'<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(c["audit_ref"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>'
       f'<dd class="fw-700 mb-2">{e(c["date_str"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Organization</dt>'
       f'<dd class="fw-700 mb-2">{e(c["org"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Findings</dt>'
       f'<dd class="mb-2">{len(c["findings"])} total</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>'
       f'<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(rid)}</dd></dl></div>')
    ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
       'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Navigation</div>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a href="{e(c["year_url"])}" style="color:var(--navy);font-weight:600;'
       f'text-decoration:none">&larr; All {c["year"]} Records</a></li>'
       f'<li class="mb-2"><a href="{e(c["month_url"])}" style="color:inherit;opacity:.7;'
       f'text-decoration:none">&larr; {c["year"]}/{c["month"]:02d}</a></li>'
       f'<li class="mb-2"><a href="{e(c["prev_entry_url"])}" style="color:inherit;opacity:.55;'
       f'text-decoration:none">&larr; Previous Entry</a></li>'
       f'<li class="mb-2"><a href="{e(c["next_entry_url"])}" style="color:var(--navy);font-weight:600;'
       f'text-decoration:none">Next in Series &rarr;</a></li></ul></div>')
    if c['related_docs']:
        ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
           'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Related Documents</div>')
        ap(''.join(
            f'<a href="{e(d["url"])}" style="display:block;padding:.5rem .7rem;border:1px solid var(--border);'
            f'border-left:3px solid var(--gold);text-decoration:none;color:inherit;margin-bottom:.4rem">'
            f'<div style="font-size:.6rem;opacity:.5;margin-bottom:.1rem">{e(d["date"])}</div>'
            f'<div style="font-size:.75rem;font-weight:600;color:var(--navy)">{e(d["label"])}</div></a>'
            for d in c['related_docs']
        ))
        ap('</div>')
    ap(_year_browser_plain_html(c))
    ap('</div></div>')

    ap('</div></div></section>')
    # Note: unlike the default variant, the real archive_compliance.html
    # never renders a bulk_hex <style>/<script> block, even though
    # _generate_compliance_content still computes bulk_hex_js/css.
    return ''.join(parts)


def render_minutes_default(ctx):
    """templates/honeypot/archive_minutes.html, main-domain branch."""
    c = ctx
    rid = c['record_id']
    parts = []
    ap = parts.append

    ap('<section class="page-banner"><div class="container">')
    ap(f'<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;'
       f'letter-spacing:.18em;font-size:.75rem">'
       f'<a href="/archive/" style="color:var(--gold)">Archive</a>'
       f' &rsaquo; <a href="{e(c["year_url"])}" style="color:var(--gold)">{c["year"]}</a>'
       f' &rsaquo; <a href="{e(c["month_url"])}" style="color:var(--gold)">{c["month"]:02d}</a>'
       f' &rsaquo; {c["day"]:02d}</p>')
    ap(f'<p style="color:rgba(255,255,255,.6);font-size:.72rem;margin-bottom:.4rem;letter-spacing:.1em;'
       f'text-transform:uppercase">Meeting Minutes &bull; {e(c["meeting_ref"])}</p>')
    ap(f'<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">{e(c["title"])}</h1>')
    ap(f'<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">'
       f'{e(c["date_str"])} &bull; {e(c["location"])}</p>')
    ap('</div></section>')

    ap('<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">')
    ap('<div class="col-lg-8">')

    quorum_badge = (
        '<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#dcfce7;color:#15803d;'
        'border:1px solid #86efac">&#10003; QUORUM ESTABLISHED</span>'
        if c['quorum'] else
        '<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#fee2e2;color:#991b1b;'
        'border:1px solid #fca5a5">&#10007; QUORUM NOT MET</span>'
    )
    ap(f'<div style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);'
       f'padding:1.25rem 1.5rem;margin-bottom:2rem">'
       f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;'
       f'color:var(--gold);margin-bottom:.5rem">Meeting Information</div>'
       f'<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.83rem;'
       f'margin-bottom:.75rem">'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Committee</span><br><strong>{e(c["committee"])}</strong></div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Date</span><br><strong>{e(c["date_str"])}</strong></div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Location</span><br>{e(c["location"])}</div>'
       f'<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">'
       f'Meeting Ref</span><br><code style="font-size:.75rem">{e(c["meeting_ref"])}</code></div></div>'
       f'<div style="display:flex;align-items:center;gap:.5rem;padding-top:.5rem;border-top:1px solid var(--border)">'
       f'{quorum_badge}<span style="font-size:.72rem;color:#666">{c["num_present"]} of {c["total_seats"]} '
       f'members present</span></div></div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       'border-bottom:2px solid var(--gold)">Attendance</h5>'
       '<table class="dash-table w-100"><thead><tr><th>Name</th><th>Title</th><th>Role</th>'
       '<th style="text-align:center">Present</th></tr></thead><tbody>')
    ap(''.join(
        f'<tr{" style=\"opacity:.5\"" if not m["present"] else ""}>'
        f'<td style="font-weight:{"600" if m["present"] else "400"}">{e(m["name"])}</td>'
        f'<td style="font-size:.78rem">{e(m["title"])}</td>'
        f'<td style="font-size:.75rem;color:#666">{e(m["role"])}</td>'
        f'<td style="text-align:center">'
        + ('<span style="color:#15803d;font-weight:700">&#10003;</span>' if m['present']
           else '<span style="color:#ccc">&mdash;</span>')
        + '</td></tr>'
        for m in c['members']
    ))
    ap('</tbody></table></div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       'border-bottom:2px solid var(--gold)">Agenda</h5>')
    for item in c['items']:
        ap(f'<div style="margin-bottom:1.5rem;padding-bottom:1.25rem;border-bottom:1px solid var(--border)">'
           f'<div style="font-size:.68rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;'
           f'color:#999;margin-bottom:.2rem">Item {item["number"]}</div>'
           f'<h6 style="font-size:.92rem;font-weight:700;color:var(--navy);margin-bottom:.6rem">'
           f'{e(item["title"])}</h6>'
           f'<p style="font-size:.87rem;line-height:1.7;color:#333;margin-bottom:0">{e(item["discussion"])}</p>')
        if item.get('motion'):
            mo = item['motion']
            carried = (
                '<strong style="color:#15803d">CARRIED</strong>' if mo['carried']
                else '<strong style="color:#991b1b">FAILED</strong>'
            )
            ap(f'<div style="background:#f8f9fb;border:1px solid var(--border);border-left:3px solid var(--navy);'
               f'padding:.8rem 1rem;margin-top:.65rem">'
               f'<div style="font-size:.62rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;'
               f'color:#888;margin-bottom:.4rem">Motion</div>'
               f'<p style="margin-bottom:.4rem;font-size:.83rem;line-height:1.6;color:#333">'
               f'<strong>{e(mo["verb"])}:</strong> {e(mo["text"])}</p>'
               f'<div style="font-size:.8rem;color:#555;margin-bottom:.3rem">'
               f'<strong>Moved:</strong> {e(mo["moved_by"])} &nbsp;&bull;&nbsp; '
               f'<strong>Seconded:</strong> {e(mo["seconded_by"])}</div>'
               f'<div style="font-size:.8rem;color:#555">'
               f'<strong>Vote:</strong> {mo["yea"]} Yea &nbsp;/&nbsp; {mo["nay"]} Nay &nbsp;/&nbsp; '
               f'{mo["abstain"]} Abstain &nbsp;&mdash;&nbsp; {carried}</div></div>')
        ap('</div>')
    ap('</div>')

    ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
       'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;'
       'border-bottom:2px solid var(--gold)">Action Items</h5>'
       '<table class="dash-table w-100"><thead><tr><th style="width:2rem">#</th><th>Description</th>'
       '<th>Owner</th><th>Due</th></tr></thead><tbody>')
    ap(''.join(
        f'<tr><td style="color:#999;font-weight:600">{ai["number"]}</td>'
        f'<td style="font-size:.83rem">{e(ai["description"])}</td>'
        f'<td style="font-size:.78rem;font-weight:600;white-space:nowrap">{e(ai["owner"])}</td>'
        f'<td style="font-size:.72rem;color:#666;white-space:nowrap">{e(ai["due_date"])}</td></tr>'
        for ai in c['action_items']
    ))
    ap('</tbody></table></div>')

    ap(f'<div style="border:1px solid var(--border);padding:1rem 1.2rem;margin-bottom:1.5rem">'
       f'<p style="font-size:.85rem;color:#333;line-height:1.7;margin-bottom:.5rem">'
       f'There being no further business, a motion to adjourn was made and carried unanimously.'
       f' Meeting adjourned at {e(c["adjourn_time"])}.</p>'
       f'<p style="font-size:.83rem;margin-bottom:.75rem;color:#555"><strong>Next meeting:</strong> '
       f'{e(c["next_meeting"])}</p><div><p style="font-size:.78rem;color:#888;margin-bottom:.25rem">'
       f'Respectfully submitted,</p>'
       f'<p style="font-size:.85rem;font-weight:600;color:var(--navy);margin-bottom:.1rem">'
       f'{e(c["secretary"]["name"])}</p>'
       f'<p style="font-size:.78rem;color:#666;margin-bottom:.75rem">{e(c["secretary"]["title"])}, '
       f'{e(c["secretary"]["role"])}</p>'
       f'<div style="display:flex;gap:2rem">'
       f'<div><div style="width:160px;border-bottom:1px solid #aaa;margin-bottom:.2rem;height:1.5rem"></div>'
       f'<div style="font-size:.68rem;color:#888">Signature</div></div>'
       f'<div><div style="width:120px;border-bottom:1px solid #aaa;margin-bottom:.2rem;height:1.5rem"></div>'
       f'<div style="font-size:.68rem;color:#888">Approved: ___________</div></div></div></div></div>')

    ap(f'<div style="display:flex;justify-content:space-between;padding-top:1rem;border-top:1px solid var(--border)">'
       f'<a href="{e(c["prev_entry_url"])}" style="font-size:.85rem;color:var(--muted);'
       f'text-decoration:none">&larr; Previous</a>'
       f'<a href="{e(c["next_entry_url"])}" style="font-size:.85rem;color:var(--navy);font-weight:600;'
       f'text-decoration:none">Next &rarr;</a></div>')

    ap('</div>')  # end col-lg-8

    quorum_dl = (
        '<span style="color:#15803d;font-weight:700">Established</span>' if c['quorum']
        else '<span style="color:#991b1b;font-weight:700">Not Met</span>'
    )
    ap('<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">')
    ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
       'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Meeting Record</div>'
       '<dl class="mb-0" style="font-size:.82rem">'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Meeting Ref</dt>'
       f'<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(c["meeting_ref"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>'
       f'<dd class="fw-700 mb-2">{e(c["date_str"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Committee</dt>'
       f'<dd class="fw-700 mb-2">{e(c["committee"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Quorum</dt>'
       f'<dd class="mb-2">{quorum_dl}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Agenda Items</dt>'
       f'<dd class="mb-2">{len(c["items"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>'
       f'<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(rid)}</dd></dl></div>')
    ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
       'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Navigation</div>'
       '<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a href="{e(c["year_url"])}" style="color:var(--navy);font-weight:600;'
       f'text-decoration:none">&larr; All {c["year"]} Records</a></li>'
       f'<li class="mb-2"><a href="{e(c["month_url"])}" style="color:inherit;opacity:.7;'
       f'text-decoration:none">&larr; {c["year"]}/{c["month"]:02d}</a></li>'
       f'<li class="mb-2"><a href="{e(c["prev_entry_url"])}" style="color:inherit;opacity:.55;'
       f'text-decoration:none">&larr; Previous Entry</a></li>'
       f'<li class="mb-2"><a href="{e(c["next_entry_url"])}" style="color:var(--navy);font-weight:600;'
       f'text-decoration:none">Next in Series &rarr;</a></li></ul></div>')
    if c['related_docs']:
        ap('<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;'
           'letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Related Documents</div>')
        ap(''.join(
            f'<a href="{e(d["url"])}" style="display:block;padding:.5rem .7rem;border:1px solid var(--border);'
            f'border-left:3px solid var(--gold);text-decoration:none;color:inherit;margin-bottom:.4rem">'
            f'<div style="font-size:.6rem;opacity:.5;margin-bottom:.1rem">{e(d["date"])}</div>'
            f'<div style="font-size:.75rem;font-weight:600;color:var(--navy)">{e(d["label"])}</div></a>'
            for d in c['related_docs']
        ))
        ap('</div>')
    ap(_year_browser_plain_html(c))
    ap('</div></div>')

    ap('</div></div></section>')
    # Note: unlike the default variant, the real archive_minutes.html never
    # renders a bulk_hex <style>/<script> block, even though
    # _generate_minutes_content still computes bulk_hex_js/css.
    return ''.join(parts)
