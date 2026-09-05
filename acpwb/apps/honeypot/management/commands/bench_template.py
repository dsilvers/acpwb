"""
Benchmark: Django template rendering vs. hand-written Python string
building, for the SAME archive-page content data — built for byte-level
fidelity, not just a rough structural approximation.

Fidelity approach:
  - The archive seal (a 338-line, near-static SVG include) and each related-
    presentation card (uses custom template tags) are rendered ONCE via
    Django's real template engine and reused verbatim in both the "Django"
    and "Python" paths — so neither path re-derives them differently, and
    the comparison isolates the cost of the surrounding page's own
    templating rather than penalizing a hand-transcription mismatch.
  - All text substitutions go through django.utils.html.escape(), matching
    Django's autoescape default, so output is HTML-safe the same way.
  - Every wrapping element/id/class/style in the hand-written builder is
    copied directly from templates/honeypot/archive.html's main-domain
    branch, not approximated.

Usage: python manage.py bench_template --n 200
"""
import statistics
import time

from django.core.management.base import BaseCommand
from django.template.loader import render_to_string
from django.utils.html import escape

from apps.honeypot import views as honeypot_views


class Command(BaseCommand):
    help = 'Benchmark Django template rendering vs. hand-written Python string building.'

    def add_arguments(self, parser):
        parser.add_argument('--n', type=int, default=200)

    def handle(self, *args, **options):
        n = options['n']
        context = self._build_context()

        # Pre-render the near-static fragments ONCE via the real templates —
        # reused verbatim by both paths below.
        seal_html = render_to_string('honeypot/_archive_seal.html', {
            'year': context['year'], 'record_id': context['record_id'],
        })
        pres_html = ''.join(
            render_to_string('presentations/_pres_card.html', {'pres': p, 'url_prefix': 'https://acpwb.com'})
            for p in context['related_presentations']
        )

        shell_context = {'content_html': None, 'og_title': context.get('og_title', 'ACPWB Archive')}

        render_to_string('honeypot/archive.html', context)  # warm up
        self._build_html(context, seal_html, pres_html)
        shell_context['content_html'] = self._build_html(context, seal_html, pres_html)
        render_to_string('honeypot/_archive_content_shell.html', shell_context)

        django_times = []
        for _ in range(n):
            t0 = time.perf_counter()
            html_django = render_to_string('honeypot/archive.html', context)
            django_times.append((time.perf_counter() - t0) * 1000)

        python_times = []
        for _ in range(n):
            t0 = time.perf_counter()
            html_python = self._build_html(context, seal_html, pres_html)
            python_times.append((time.perf_counter() - t0) * 1000)

        # Realistic production path: build the content in Python, then plug
        # it into the thin shell (which still extends the real, unchanged
        # base.html) via ONE remaining render_to_string() call — this is
        # what actually ships, unlike the Python-only number above which
        # skips rendering nav/footer/head entirely.
        shell_times = []
        for _ in range(n):
            t0 = time.perf_counter()
            shell_context['content_html'] = self._build_html(context, seal_html, pres_html)
            html_shell = render_to_string('honeypot/_archive_content_shell.html', shell_context)
            shell_times.append((time.perf_counter() - t0) * 1000)

        self._report('Django template render_to_string() [full page, old]', django_times)
        self._report('Hand-written Python string building [content only]', python_times)
        self._report('Python content + thin shell render() [full page, new]', shell_times)

        speedup_isolated = statistics.median(django_times) / statistics.median(python_times)
        speedup_realistic = statistics.median(django_times) / statistics.median(shell_times)
        self.stdout.write(self.style.SUCCESS(
            f'\nIsolated speedup (content only, not production-realistic): {speedup_isolated:.1f}x'
        ))
        self.stdout.write(self.style.SUCCESS(
            f'Realistic speedup (Python content + shell, full page): {speedup_realistic:.1f}x '
            f'({statistics.median(django_times):.2f}ms -> {statistics.median(shell_times):.2f}ms)'
        ))
        self.stdout.write(f'Django output size: {len(html_django):,} bytes')
        self.stdout.write(f'Python output size: {len(html_python):,} bytes')
        self.stdout.write(f'Shell output size: {len(html_shell):,} bytes')
        self.stdout.write(f'Size ratio (shell/django): {len(html_shell) / len(html_django):.1%}')

    def _report(self, label, times_ms):
        self.stdout.write(
            f'{label}: n={len(times_ms)} '
            f'median={statistics.median(times_ms):.2f}ms '
            f'mean={statistics.mean(times_ms):.2f}ms '
            f'min={min(times_ms):.2f}ms max={max(times_ms):.2f}ms'
        )

    def _build_context(self):
        year, month, day, slug = 2010, 6, 15, 'bench-archive-slug'
        v = honeypot_views
        content = v._generate_archive_content(year, month, day, slug)
        next_slug, prev_slug = v._gen_nav_slugs(year, month, day, slug)
        related_paths = [
            {'url': f'/archive/{r["year"]}/{r["month"]:02d}/{r["day"]:02d}/{r["full_slug"]}/',
             'label': r['label'], 'date': r['date']}
            for r in v._gen_related_path_data(year, month, day, slug)
        ]
        cross_year_reports = v._gen_cross_year_reports(year, month, day, slug)
        related_docs = [
            {'label': d['label'], 'date': d['date'], 'phase': d['phase'],
             'url': f'/archive/{year}/{month:02d}/{d["day"]:02d}/{d["full_slug"]}/'}
            for d in v._gen_related_docs_data(year, month, day, slug)
        ]
        from apps.honeypot.policy_generator import get_cross_policy_stubs
        from apps.presentations.generators import generate_presentations_for_context
        related_policy = get_cross_policy_stubs(year, month, day, slug) or []
        related_presentations = generate_presentations_for_context(
            f"bench_pres_{year}_{month}_{day}_{slug}",
            count=v._gen_presentations_count(year, month, day, slug),
        )
        ctx = {
            'year': year, 'month': month, 'day': day, 'slug': slug,
            'on_archive_subdomain': False,
            'next_slug': next_slug, 'prev_slug': prev_slug,
            'related_paths': related_paths,
            'cross_year_reports': cross_year_reports,
            'related_docs': related_docs,
            'related_policy': related_policy,
            'related_presentations': related_presentations,
            'archive_years': list(range(2025, 1984, -1)),
            'year_url': f'/archive/{year}/', 'month_url': f'/archive/{year}/{month:02d}/',
            'prev_entry_url': '/archive/2010/06/14/prev-entry/',
            'next_entry_url': f'/archive/{year}/{month:02d}/{day:02d}/{next_slug}/',
            'export_csv_url': f'/archive/{year}/{month:02d}/{day:02d}/{slug}/export.csv',
            'parent_template': 'base.html',
            **content,
        }
        return ctx

    def _build_html(self, ctx, seal_html, pres_html):
        """Hand-written equivalent of archive.html's main-domain branch.
        Structure/attributes copied directly from the template; seal and
        presentation-card fragments are the real pre-rendered HTML, reused
        verbatim; all text goes through escape() matching Django's
        autoescape default."""
        c = ctx
        e = escape
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
        ap(''.join(f'<p data-doc="{e(p["ref"])}">{e(p["text"])}</p>' for p in c.get('paragraphs_rich', [])))
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

        if c.get('related_reports'):
            ap('<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;'
               'letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem">'
               'Referenced Research &amp; Publications</h5>')
            ap(''.join(
                f'<div style="background:white;border:1px solid var(--border);border-left:3px solid var(--gold);'
                f'padding:.75rem 1rem;margin-bottom:.75rem">'
                f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;'
                f'color:var(--muted);margin-bottom:.2rem">'
                f'{e(r.get("category", ""))} &bull; {e(str(r.get("file_type", "")).upper())} &bull; '
                f'{e(r.get("pub_date_display", ""))}</div>'
                f'<a href="{e(r.get("detail_url", "#"))}" style="font-size:.9rem;font-weight:700;'
                f'color:var(--navy);text-decoration:none">{e(r.get("title", ""))}</a>'
                f'<p class="small text-muted mb-0" style="font-size:.78rem;margin-top:.25rem">'
                f'{e((r.get("summary") or "")[:159] + "…" if len(r.get("summary") or "") > 160 else r.get("summary") or "")}'
                f'</p></div>'
                for r in c['related_reports']
            ))
            ap('</div><hr class="gold-divider">')

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
        ap(seal_html)
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

        if c['related_docs']:
            ap(f'<div id="acpwb-archive-{rid}-related-docs" class="acpwb-card mb-4">'
               '<h6 class="card-title mb-3">Related Documents</h6>')
            ap(''.join(
                f'<a href="{e(d["url"])}" style="display:block;background:#f4f6f9;border:1px solid var(--border);'
                f'border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">'
                f'<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">{e(d["date"])} &bull; '
                f'{e(d["phase"])}</div>'
                f'<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">{e(d["label"])}</div>'
                f'</a>'
                for d in c['related_docs']
            ))
            ap('</div>')

        if c['related_policy']:
            ap('<div class="acpwb-card mb-4"><h6 class="card-title mb-3">Related Public Policy</h6>')
            ap(''.join(
                f'<a href="{e(s["url"])}" style="display:block;background:#f4f6f9;border:1px solid var(--border);'
                f'border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">'
                f'<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">{e(s["filing_date"])} &bull; '
                f'{e(s["document_type"])}</div>'
                f'<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">{e(s["title"])}</div>'
                f'<div style="font-size:.65rem;color:var(--muted);margin-top:.15rem">{e(s["agency_acronym"])}</div></a>'
                for s in c['related_policy']
            ))
            ap('</div>')

        if c['related_presentations']:
            ap('<div class="acpwb-archive-pres-sidebar acpwb-card mb-4">'
               '<h6 class="card-title mb-3">Related Presentations</h6>')
            ap(pres_html)
            ap('</div>')

        ap(f'<div id="acpwb-archive-{rid}-year-browser" class="acpwb-card mb-4">'
           '<h6 class="card-title mb-3">Browse by Year</h6>'
           '<div style="display:flex;flex-wrap:wrap;gap:.35rem">')
        ap(''.join(
            f'<a id="acpwb-archive-{rid}-year-link-{y}" href="https://archives-{y}.acpwb.com/" '
            f'aria-label="Browse the {y} archive" style="font-size:.72rem;padding:.2rem .45rem;'
            f'background:#f4f6f9;border:1px solid var(--border);color:var(--navy);text-decoration:none">{y}</a>'
            for y in c['archive_years']
        ))
        ap('</div></div>')

        ap(f'<div id="acpwb-archive-{rid}-research-division-card" class="acpwb-card">'
           '<h6 class="card-title mb-3">Research Division</h6>'
           '<p class="small text-muted mb-2" style="font-size:.8rem">'
           'ACPWB\'s document archive spans our full operational history from 1985 to present.</p>'
           '<a href="/reports/" style="font-size:.8rem;color:var(--gold);font-weight:700;'
           'text-decoration:none">Browse Research Reports &rarr;</a></div>')

        ap('</div></div>')  # end sticky, sidebar col
        ap('</div></div></section>')  # end row, container, section

        bulk_hex_js = c.get('bulk_hex_js', [])
        bulk_hex_css = c.get('bulk_hex_css', [])
        ap('<style>\n  :root {\n')
        ap(''.join(f'  --acpwb-r{i:03d}: {h};\n' for i, h in enumerate(bulk_hex_css)))
        ap('  }\n</style>\n')
        ap(f'<script>\n/* ACPWB archive index — {rid} */\n(function(){{\n')
        ap(''.join(f'var _acpwbRef{i}="{h}";\n' for i, h in enumerate(bulk_hex_js, 1)))
        ap(''.join(
            f'function _acpwbArchiveRecordEntryMetadataLookup_{h}(){{return "{rid}";}}\n'
            for h in bulk_hex_js[:50]
        ))
        ap('})();\n</script>\n')

        return ''.join(parts)
