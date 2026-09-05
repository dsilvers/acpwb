"""
Hand-written Python builders for the era/subdomain (Jinja2-backend) archive
templates: templates/jinja2/honeypot/era/archive.html,
archive_compliance.html, archive_minutes.html.

`ctx` is the same merged dict archive_trap() builds for render() — same
content-variant dict merged with the shared page-level fields.
"""
import jinja2
from jinja2.filters import do_truncate

from apps.core.htmlgen import escape as e
from apps.core.htmlgen import get_archive_seal

# do_truncate needs a real Environment to read its default truncation
# policies (leeway, etc.) — a fresh default-config Environment matches
# Django's Jinja2 backend defaults for this filter.
_JINJA_ENV = jinja2.Environment()


def _truncate(s, length):
    return do_truncate(_JINJA_ENV, s, length)


def _bulk_hex_css_vars(bulk_hex_css):
    return ''.join(f'--acpwb-r{i:03d}: {h};\n' for i, h in enumerate(bulk_hex_css))


def _bulk_hex_script(bulk_hex_js, rid):
    refs = ''.join(f'var _acpwbRef{i}="{h}";\n' for i, h in enumerate(bulk_hex_js, 1))
    funcs = ''.join(
        f'function _acpwbArchiveRecordEntryMetadataLookup_{h}(){{return "{rid}";}}\n'
        for h in bulk_hex_js[:50]
    )
    return f'<script>\n/* ACPWB archive index — {rid} */\n(function(){{\n{refs}{funcs}}})();\n</script>\n'


def _era_sidebar_presentations_html(c, yd):
    if not c['related_presentations']:
        return ''
    from django.templatetags.static import static
    items = []
    for pres in c['related_presentations']:
        if pres.get('thumb_bg'):
            bg_style = f"background-image:url('{static(pres['thumb_bg'])}');background-size:cover;background-position:center;"
        else:
            bg_style = f"background:{pres['theme']['bg']};"
        items.append(
            f'<a href="https://acpwb.com{e(pres["pres_url"])}" style="display:block;margin-bottom:.75rem;'
            f'text-decoration:none;color:inherit;border:1px solid rgba(128,128,128,.2);overflow:hidden">'
            f'<div style="{bg_style}aspect-ratio:16/9;position:relative;display:flex;align-items:flex-end;'
            f'padding:.5em .6em">'
            f'<div style="position:absolute;top:.4em;right:.5em;background:{pres["theme"]["accent"]};'
            f'color:{pres["theme"]["bg"]};font-size:.55rem;padding:.15em .4em;font-weight:800">'
            f'{pres["slide_count"]} slides</div>'
            f'<div style="font-size:.62rem;font-weight:700;color:{pres["theme"]["text"]};line-height:1.2;'
            f'text-shadow:0 1px 3px rgba(0,0,0,.7)">{e(_truncate(pres["title"], 55))}</div></div>'
            f'<div style="padding:.45em .65em;background:rgba(128,128,128,.06)">'
            f'<div style="font-size:.6rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;'
            f'color:{yd["accent"]};margin-bottom:.1em">{e(pres["org_name"])}</div>'
            f'<div style="font-size:.62rem;opacity:.65">{e(pres["pub_date_display"])} &mdash; '
            f'{e(pres["industry"])}</div></div></a>'
        )
    return (
        '<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;'
        'margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.8rem">Related Presentations</div>'
        + ''.join(items) + '</div>'
    )


def render_archive_default_era(ctx):
    """templates/jinja2/honeypot/era/archive.html."""
    c = ctx
    rid = c['record_id']
    yd = c['year_data']
    parts = []
    ap = parts.append

    ap(
        '<style>\n'
        f'  .era-archive-banner {{ background: {yd["accent"]}; color: #fff; padding: 2rem 0 1.5rem; '
        f'font-family: {yd["font_head"]}, sans-serif; }}\n'
        f'  .era-archive-content {{ padding: 3rem 0; background: {yd["bg"]}; color: {yd["text_color"]}; '
        f'font-family: {yd["font_body"]}, sans-serif; }}\n'
        f'  .era-callout {{ background: rgba(128,128,128,.08); border: 1px solid rgba(128,128,128,.2); '
        f'border-left: 4px solid {yd["accent"]}; padding: 1.1rem 1.4rem; margin-bottom: 1.75rem; }}\n'
        f'  .era-section-head {{ font-family: {yd["font_head"]}, sans-serif; font-size: .78rem; '
        f'font-weight: 800; text-transform: uppercase; letter-spacing: .1em; color: {yd["accent"]}; '
        f'margin-bottom: .85rem; padding-bottom: .4rem; border-bottom: 2px solid {yd["accent"]}; }}\n'
        f'  .era-entry-card {{ background: rgba(128,128,128,.07); border: 1px solid rgba(128,128,128,.2); '
        f'border-left: 3px solid {yd["accent"]}; padding: .7rem 1rem; text-decoration: none; '
        f'color: {yd["text_color"]}; display: block; }}\n'
        f'  .era-entry-card:hover {{ border-left-color: {yd["accent2"]}; color: {yd["text_color"]}; }}\n'
        f'  .era-nav-link {{ font-size: .85rem; color: {yd["accent2"]}; font-weight: 600; text-decoration: none; }}\n'
        f'  .era-table-head {{ background: {yd["accent"]}; }}\n'
        f'  .era-cta {{ background: {yd["accent"]}; color: #fff; }}\n'
        f'  :root {{\n{_bulk_hex_css_vars(c.get("bulk_hex_css", []))}  }}\n'
        '</style>\n'
    )

    ap('<div class="era-archive-banner"><div class="container">')
    ap(f'<p class="text-uppercase mb-1" style="font-weight:800;letter-spacing:.18em;font-size:.75rem;opacity:.8">'
       f'<a href="/archive/" style="color:inherit">Archive</a>'
       f' &rsaquo; <a href="{e(c["year_url"])}" style="color:inherit">{c["year"]}</a>'
       f' &rsaquo; <a href="{e(c["month_url"])}" style="color:inherit">{c["month"]:02d}</a>'
       f' &rsaquo; {c["day"]:02d}</p>')
    ap(f'<h1 style="font-family:var(--era-font-head);font-size:clamp(1.1rem,3vw,1.9rem);line-height:1.25;'
       f'margin-bottom:.3rem">{e(c["title"])}</h1>')
    ap(f'<p style="opacity:.75;font-size:.88rem;margin-bottom:0">'
       f'{e(c["industry"])} &bull; {e(c["org"])} &bull; {e(c["phase"].capitalize())} phase &bull; '
       f'{c["year"]}-{c["month"]:02d}-{c["day"]:02d}</p>')
    ap('</div></div>')

    ap('<div class="era-archive-content"><div class="container"><div class="row g-4"><div class="col-lg-8">')

    exec_mb = '.75rem' if c.get('exec_bullets') else '0'
    ap(f'<div id="acpwb-archive-{rid}-executive-summary-callout" class="era-callout">'
       f'<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;'
       f'color:var(--era-accent);margin-bottom:.45rem">Executive Summary</div>'
       f'<p style="font-size:.88rem;margin-bottom:{exec_mb};line-height:1.65">'
       f'This archive entry documents ACPWB\'s <strong>{e(c["phase"])}</strong> phase engagement with '
       f'<strong>{e(c["org"])}</strong> in the <strong>{e(c["industry"])}</strong> sector.'
       f' Record ID <code style="font-size:.78rem;background:rgba(128,128,128,.15);padding:.1rem .3rem">'
       f'{e(rid)}</code>.</p>')
    if c.get('exec_bullets'):
        ap(f'<ul id="acpwb-archive-{rid}-exec-summary-bullets" style="padding-left:1.25rem;margin-bottom:0">')
        ap(''.join(
            f'<li style="font-size:.83rem;margin-bottom:.45rem;line-height:1.6">{e(b)}</li>'
            for b in c['exec_bullets']
        ))
        ap('</ul>')
    ap('</div>')

    if c.get('findings'):
        ap(f'<div id="acpwb-archive-{rid}-key-findings-block" class="mb-4">'
           f'<div class="era-section-head">Key Findings</div>'
           f'<ul id="acpwb-archive-{rid}-key-findings-list" style="padding-left:1.25rem;margin-bottom:0">')
        n_findings = len(c['findings'])
        ap(''.join(
            f'<li id="acpwb-archive-{rid}-finding-{i}" data-ref="{e(f["ref"])}" '
            f'aria-label="Finding {i} of {n_findings}: {e(f["text"])}" '
            f'style="font-size:.88rem;margin-bottom:.6rem;line-height:1.6">{e(f["text"])}</li>'
            for i, f in enumerate(c['findings'], 1)
        ))
        ap('</ul></div>')

    if c.get('engagement_team'):
        ap(f'<div id="acpwb-archive-{rid}-engagement-team" class="mb-4">'
           f'<div class="era-section-head">Engagement Team</div>'
           f'<table style="width:100%;border-collapse:collapse;font-size:.82rem">')
        ap(''.join(
            f'<tr style="border-top:1px solid rgba(128,128,128,.15)">'
            f'<td style="padding:.35rem .7rem .35rem 0;white-space:nowrap">'
            f'<div style="font-weight:700">{e(m["name"])}</div>'
            f'<div style="font-size:.72rem;opacity:.6"><a href="mailto:{e(m["email"])}" '
            f'style="color:inherit">{e(m["email"])}</a></div></td>'
            f'<td style="padding:.35rem 0;opacity:.75">{e(m["title"])}</td></tr>'
            for m in c['engagement_team']
        ))
        ap('</table></div>')

    ap(f'<div id="acpwb-archive-{rid}-body-content" class="mb-4" style="font-family:var(--era-font-body);'
       f'line-height:1.85">')
    ap(''.join(f'<p data-doc="{e(p["ref"])}">{e(p["text"])}</p>' for p in c.get('paragraphs', [])))
    ap('</div>')

    if c.get('metric_rows'):
        ap(f'<div id="acpwb-archive-{rid}-engagement-metrics-section" class="mb-4">'
           f'<div class="era-section-head">Engagement Metrics</div><div style="overflow-x:auto">'
           f'<table id="acpwb-archive-{rid}-engagement-metrics-table" '
           f'style="width:100%;border-collapse:collapse;font-size:.83rem;border:1px solid rgba(128,128,128,.25)">'
           f'<thead><tr class="era-table-head" style="color:#fff">'
           '<th style="padding:.55rem .85rem;text-align:left;font-size:.66rem;text-transform:uppercase;'
           'letter-spacing:.06em">Metric</th>'
           '<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;'
           'letter-spacing:.06em">Baseline</th>'
           '<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;'
           'letter-spacing:.06em">Current</th>'
           '<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;'
           'letter-spacing:.06em">Change</th></tr></thead><tbody>')
        ap(''.join(
            f'<tr id="acpwb-archive-{rid}-metric-row-{i}" '
            f'aria-label="Metric: {e(r["name"])}, baseline {e(r["baseline"])}, current {e(r["current"])}, '
            f'change {e(r["delta"])}" style="border-top:1px solid rgba(128,128,128,.2)">'
            f'<td style="padding:.45rem .85rem;font-weight:600">{e(r["name"])}</td>'
            f'<td style="padding:.45rem .85rem;text-align:right;opacity:.65">{e(r["baseline"])}</td>'
            f'<td style="padding:.45rem .85rem;text-align:right;font-weight:700">{e(r["current"])}</td>'
            f'<td style="padding:.45rem .85rem;text-align:right;font-weight:700;'
            f'color:{"#16a34a" if r["positive"] else "#dc2626"}">{e(r["delta"])}</td></tr>'
            for i, r in enumerate(c['metric_rows'], 1)
        ))
        ap('</tbody></table></div></div>')

    if c.get('percentile_table'):
        ap(f'<div id="acpwb-archive-{rid}-benchmark-percentile-section" class="mb-4">'
           f'<div class="era-section-head">Market Benchmark — {e(c.get("peer_group", ""))}</div>'
           f'<div style="overflow-x:auto"><table style="width:100%;border-collapse:collapse;font-size:.8rem;'
           f'border:1px solid rgba(128,128,128,.25)"><thead><tr class="era-table-head" style="color:#fff">'
           + ''.join(
               f'<th style="padding:.45rem .75rem;text-align:{"left" if h == "Metric" else "right"};'
               f'font-size:.63rem;text-transform:uppercase;letter-spacing:.06em">{h}</th>'
               for h in ['Metric', 'P10', 'P25', 'P33', 'P50', 'P67', 'P75', 'P90', 'P95']
           ) + '</tr></thead><tbody>')
        # Per-column style is NOT uniform in the real era template (unlike
        # the main-domain version) — p75 specifically uses opacity:.75,
        # distinct from the .6/.7 used elsewhere.
        _pctile_col_style = {
            'p10': 'opacity:.6', 'p25': 'opacity:.6', 'p33': 'opacity:.7',
            'p50': 'font-weight:700', 'p67': 'opacity:.7', 'p75': 'opacity:.75',
            'p90': 'opacity:.6', 'p95': 'opacity:.6',
        }
        ap(''.join(
            f'<tr style="border-top:1px solid rgba(128,128,128,.2)">'
            f'<td style="padding:.38rem .75rem;font-weight:600">{e(r["metric"])}</td>'
            + ''.join(
                f'<td style="padding:.38rem .75rem;text-align:right;{_pctile_col_style[k]}">{e(r[k])}</td>'
                for k in ['p10', 'p25', 'p33', 'p50', 'p67', 'p75', 'p90', 'p95']
            ) + '</tr>'
            for r in c['percentile_table']
        ))
        ap('</tbody></table></div></div>')

    ap(f'<div id="acpwb-archive-{rid}-csv-export-banner" style="background:rgba(128,128,128,.07);'
       f'border:1px solid rgba(128,128,128,.2);padding:.8rem 1rem;margin-bottom:1.25rem;display:flex;'
       f'align-items:center;justify-content:space-between;flex-wrap:wrap;gap:.5rem"><div>'
       f'<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;'
       f'opacity:.65">Raw Data Export</div>'
       f'<div style="font-size:.8rem;font-weight:600">Download the underlying dataset for this archive entry'
       f'</div></div>'
       f'<a id="acpwb-archive-{rid}-csv-download-link" href="{e(c["export_csv_url"])}" class="era-cta" '
       f'aria-label="Download CSV dataset for archive record {e(rid)}" '
       f'style="font-size:.72rem;font-weight:700;padding:.3rem .85rem;text-decoration:none;white-space:nowrap">'
       f'&#x2193; Download CSV</a></div>')

    ap(f'<a id="acpwb-archive-{rid}-next-entry-link" href="{e(c["next_entry_url"])}" class="era-cta" '
       f'aria-label="Continue to the next entry in the {c["year"]} archive series" '
       f'style="display:block;padding:1rem 1.5rem;text-decoration:none;font-weight:700;font-size:.9rem;'
       f'margin-bottom:2rem;font-family:{yd["font_head"]},sans-serif">'
       f'Continue Reading: Next in Series &rarr;</a>')

    if c['related_paths']:
        ap('<div class="mt-2"><div class="era-section-head">Related Archive Entries</div><div class="row g-2">')
        ap(''.join(
            f'<div class="col-md-6"><a href="{e(ent["url"])}" class="era-entry-card">'
            f'<div style="font-size:.62rem;opacity:.6;font-weight:600;margin-bottom:.2rem">{e(ent["date"])}</div>'
            f'<div style="font-size:.8rem;font-weight:600;line-height:1.3">{e(ent["label"])}</div></a></div>'
            for ent in c['related_paths']
        ))
        ap('</div></div>')

    if c['cross_year_reports']:
        ap('<div class="mt-4"><div class="era-section-head">Related Archive Reports — Other Years</div>'
           '<div class="row g-2">')
        ap(''.join(
            f'<div class="col-md-6"><a href="{e(ent["url"])}" class="era-entry-card">'
            f'<div style="font-size:.6rem;opacity:.55;font-weight:700;text-transform:uppercase;'
            f'letter-spacing:.08em;margin-bottom:.2rem">{ent["year"]} Archive</div>'
            f'<div style="font-size:.8rem;font-weight:600;line-height:1.3">{e(ent["label"])}</div>'
            f'<div style="font-size:.62rem;opacity:.5;margin-top:.2rem">{e(ent["date"])}</div></a></div>'
            for ent in c['cross_year_reports']
        ))
        ap('</div></div>')

    if c.get('revisions'):
        ap(f'<div id="acpwb-archive-{rid}-revision-history" class="mb-4 mt-4">'
           f'<div class="era-section-head">Revision History</div>'
           f'<table style="width:100%;border-collapse:collapse;font-size:.78rem"><thead>'
           '<tr style="opacity:.55">'
           '<th style="padding:.3rem .5rem .3rem 0;font-size:.62rem;text-transform:uppercase;'
           'letter-spacing:.06em;text-align:left;white-space:nowrap">Version</th>'
           '<th style="padding:.3rem .5rem;font-size:.62rem;text-transform:uppercase;letter-spacing:.06em;'
           'text-align:left;white-space:nowrap">Date</th>'
           '<th style="padding:.3rem 0 .3rem .5rem;font-size:.62rem;text-transform:uppercase;'
           'letter-spacing:.06em;text-align:left">Description</th></tr></thead><tbody>')
        ap(''.join(
            f'<tr style="border-top:1px solid rgba(128,128,128,.15)">'
            f'<td style="padding:.35rem .5rem .35rem 0;font-weight:700;white-space:nowrap;'
            f'font-family:monospace;font-size:.72rem">{e(r["version"])}</td>'
            f'<td style="padding:.35rem .5rem;opacity:.65;white-space:nowrap">{e(r["date"])}</td>'
            f'<td style="padding:.35rem 0 .35rem .5rem;opacity:.8;line-height:1.45">{e(r["description"])} '
            f'<span style="opacity:.55">— {e(r["author"])} &lt;'
            f'<a href="mailto:{e(r["author_email"])}" style="color:inherit">{e(r["author_email"])}</a>&gt;'
            f'</span></td></tr>'
            for r in c['revisions']
        ))
        ap('</tbody></table></div>')

    if c.get('footnotes'):
        ap(f'<div id="acpwb-archive-{rid}-footnotes" class="mb-4" '
           f'style="border-top:1px solid rgba(128,128,128,.15);padding-top:1rem">'
           f'<div class="era-section-head">Sources &amp; Notes</div>'
           f'<ol style="padding-left:1.25rem;margin-bottom:0">')
        ap(''.join(
            f'<li id="acpwb-archive-{rid}-fn-{fn["num"]}" '
            f'style="font-size:.72rem;opacity:.7;margin-bottom:.4rem;line-height:1.5">{e(fn["text"])}</li>'
            for fn in c['footnotes']
        ))
        ap('</ol></div>')

    ap(f'<div class="mt-4 pt-3" style="border-top:1px solid rgba(128,128,128,.2)">'
       f'<a href="{e(c["prev_entry_url"])}" class="era-nav-link">&larr; Previous in Series</a></div>')

    ap('<div class="mt-5 pt-3 text-center" style="border-top:1px solid rgba(128,128,128,.12);opacity:.75">')
    ap(get_archive_seal(c['year'], rid))
    ap('</div>')

    ap('</div>')  # end col-lg-8

    ap(f'<div id="acpwb-archive-{rid}-sidebar-col" class="col-lg-4 d-none d-lg-block">'
       f'<div style="position:sticky;top:2rem">')
    ap(f'<div id="acpwb-archive-{rid}-sidebar-record-card" style="background:rgba(128,128,128,.07);'
       f'border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">'
       f'<div class="era-section-head" style="margin-bottom:.6rem">Archive Record</div>'
       f'<dl class="mb-0" style="font-size:.82rem">'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>'
       f'<dd class="fw-700 mb-2">{c["year"]}-{c["month"]:02d}-{c["day"]:02d}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Phase</dt>'
       f'<dd class="fw-700 mb-2" style="text-transform:capitalize">{e(c["phase"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Sector</dt>'
       f'<dd class="fw-700 mb-2">{e(c["industry"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Organization</dt>'
       f'<dd class="fw-700 mb-2">{e(c["org"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Engagement Code</dt>'
       f'<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(c["eng_code"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Version</dt>'
       f'<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(c["doc_version"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Pages</dt>'
       f'<dd class="mb-2" style="font-size:.8rem;opacity:.7">{c["page_count"]} ({c["file_size_kb"]} KB)</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Distribution</dt>'
       f'<dd class="mb-2" style="font-size:.72rem;opacity:.65;line-height:1.4">{e(c["distribution"])}</dd>'
       '<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>'
       f'<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(rid)}</dd></dl></div>')

    ap(f'<div id="acpwb-archive-{rid}-sidebar-nav-card" style="background:rgba(128,128,128,.07);'
       f'border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">'
       f'<div class="era-section-head" style="margin-bottom:.6rem">Navigation</div>'
       f'<ul class="list-unstyled mb-0" style="font-size:.82rem">'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-year" href="{e(c["year_url"])}" class="era-nav-link" '
       f'aria-label="Browse all {c["year"]} archive records">&larr; All {c["year"]} Records</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-month" href="{e(c["month_url"])}" '
       f'aria-label="Browse {c["year"]} month {c["month"]} archive records" '
       f'style="color:inherit;opacity:.7;text-decoration:none">&larr; {c["year"]}/{c["month"]:02d}</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-prev" href="{e(c["prev_entry_url"])}" '
       f'aria-label="Navigate to the previous entry in the archive series" '
       f'style="color:inherit;opacity:.55;text-decoration:none">&larr; Previous Entry</a></li>'
       f'<li class="mb-2"><a id="acpwb-archive-{rid}-nav-next" href="{e(c["next_entry_url"])}" class="era-nav-link" '
       f'aria-label="Navigate to the next entry in the {c["year"]} archive series">Next in Series &rarr;</a></li>'
       f'</ul></div>')

    if c['related_docs']:
        ap(f'<div id="acpwb-archive-{rid}-related-docs" style="background:rgba(128,128,128,.07);'
           f'border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">'
           f'<div class="era-section-head" style="margin-bottom:.6rem">Related Documents</div>')
        ap(''.join(
            f'<a href="{e(d["url"])}" class="era-entry-card" style="margin-bottom:.5rem">'
            f'<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">{e(d["date"])} &bull; '
            f'{e(d["phase"])}</div>'
            f'<div style="font-size:.75rem;font-weight:600;line-height:1.35">{e(d["label"])}</div></a>'
            for d in c['related_docs']
        ))
        ap('</div>')

    if c['related_policy']:
        ap('<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;'
           'margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">'
           'Related Public Policy</div>')
        ap(''.join(
            f'<a href="https://acpwb.com{e(s["url"])}" style="display:block;background:rgba(128,128,128,.05);'
            f'border:1px solid rgba(128,128,128,.2);border-left:3px solid var(--era-accent2);'
            f'padding:.55rem .75rem;margin-bottom:.5rem;text-decoration:none;color:inherit">'
            f'<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">{e(s["filing_date"])} &bull; '
            f'{e(s["document_type"])}</div>'
            f'<div style="font-size:.74rem;font-weight:600;line-height:1.35">{e(s["title"])}</div>'
            f'<div style="font-size:.65rem;opacity:.55;margin-top:.15rem">{e(s["agency_acronym"])}</div></a>'
            for s in c['related_policy']
        ))
        ap('</div>')

    ap(_era_sidebar_presentations_html(c, yd))

    ap(f'<div id="acpwb-archive-{rid}-year-browser" style="background:rgba(128,128,128,.07);'
       f'border:1px solid rgba(128,128,128,.2);padding:1rem">'
       f'<div class="era-section-head" style="margin-bottom:.6rem">Browse by Year</div>'
       f'<div style="display:flex;flex-wrap:wrap;gap:.3rem">')
    ap(''.join(
        f'<a id="acpwb-archive-{rid}-year-link-{y}" href="https://archives-{y}.acpwb.com/" '
        f'aria-label="Browse the {y} archive" '
        f'style="font-size:.68rem;padding:.2rem .4rem;border:1px solid rgba(128,128,128,.3);'
        f'text-decoration:none;color:inherit'
        + (';font-weight:800;border-color:var(--era-accent);color:var(--era-accent)' if y == c['year'] else '')
        + f'">{y}</a>'
        for y in c['all_years']
    ))
    ap('</div></div>')

    ap('</div></div>')  # sticky, sidebar-col
    ap('</div></div></div>')  # row, container, era-archive-content

    ap(_bulk_hex_script(c.get('bulk_hex_js', []), rid))
    return ''.join(parts)
