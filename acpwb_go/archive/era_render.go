package archive

import (
	"fmt"
	"strings"
)

// RenderArchiveDefaultEra ports
// pyrender/archive_era.py:render_archive_default_era
// (templates/jinja2/honeypot/era/archive.html).
func RenderArchiveDefaultEra(c *EraContext) string {
	rid := c.RecordID
	yd := c.YearData
	var b strings.Builder

	fmt.Fprintf(&b,
		"<style>\n"+
			"  .era-archive-banner { background: %s; color: #fff; padding: 2rem 0 1.5rem; "+
			"font-family: %s, sans-serif; }\n"+
			"  .era-archive-content { padding: 3rem 0; background: %s; color: %s; "+
			"font-family: %s, sans-serif; }\n"+
			"  .era-callout { background: rgba(128,128,128,.08); border: 1px solid rgba(128,128,128,.2); "+
			"border-left: 4px solid %s; padding: 1.1rem 1.4rem; margin-bottom: 1.75rem; }\n"+
			"  .era-section-head { font-family: %s, sans-serif; font-size: .78rem; "+
			"font-weight: 800; text-transform: uppercase; letter-spacing: .1em; color: %s; "+
			"margin-bottom: .85rem; padding-bottom: .4rem; border-bottom: 2px solid %s; }\n"+
			"  .era-entry-card { background: rgba(128,128,128,.07); border: 1px solid rgba(128,128,128,.2); "+
			"border-left: 3px solid %s; padding: .7rem 1rem; text-decoration: none; "+
			"color: %s; display: block; }\n"+
			"  .era-entry-card:hover { border-left-color: %s; color: %s; }\n"+
			"  .era-nav-link { font-size: .85rem; color: %s; font-weight: 600; text-decoration: none; }\n"+
			"  .era-table-head { background: %s; }\n"+
			"  .era-cta { background: %s; color: #fff; }\n"+
			"  :root {\n%s  }\n"+
			"</style>\n",
		yd.Accent, yd.FontHead,
		yd.Bg, yd.TextColor, yd.FontBody,
		yd.Accent,
		yd.FontHead, yd.Accent, yd.Accent,
		yd.Accent, yd.TextColor,
		yd.Accent2, yd.TextColor,
		yd.Accent2,
		yd.Accent,
		yd.Accent,
		bulkHexCSSVars(c.BulkHexCSS),
	)

	b.WriteString(`<div class="era-archive-banner"><div class="container">`)
	fmt.Fprintf(&b,
		`<p class="text-uppercase mb-1" style="font-weight:800;letter-spacing:.18em;font-size:.75rem;opacity:.8">`+
			`<a href="/archive/" style="color:inherit">Archive</a>`+
			` &rsaquo; <a href="%s" style="color:inherit">%d</a>`+
			` &rsaquo; <a href="%s" style="color:inherit">%02d</a>`+
			` &rsaquo; %02d</p>`,
		e(c.YearURL), c.Year, e(c.MonthURL), c.Month, c.Day,
	)
	fmt.Fprintf(&b,
		`<h1 style="font-family:var(--era-font-head);font-size:clamp(1.1rem,3vw,1.9rem);line-height:1.25;`+
			`margin-bottom:.3rem">%s</h1>`,
		e(c.Title),
	)
	fmt.Fprintf(&b,
		`<p style="opacity:.75;font-size:.88rem;margin-bottom:0">`+
			`%s &bull; %s &bull; %s phase &bull; %d-%02d-%02d</p>`,
		e(c.Industry), e(c.Org), e(pyCapitalize(c.Phase)), c.Year, c.Month, c.Day,
	)
	b.WriteString(`</div></div>`)

	b.WriteString(`<div class="era-archive-content"><div class="container"><div class="row g-4"><div class="col-lg-8">`)

	execMb := "0"
	if len(c.ExecBullets) > 0 {
		execMb = ".75rem"
	}
	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-executive-summary-callout" class="era-callout">`+
			`<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--era-accent);margin-bottom:.45rem">Executive Summary</div>`+
			`<p style="font-size:.88rem;margin-bottom:%s;line-height:1.65">`+
			`This archive entry documents ACPWB's <strong>%s</strong> phase engagement with `+
			`<strong>%s</strong> in the <strong>%s</strong> sector.`+
			` Record ID <code style="font-size:.78rem;background:rgba(128,128,128,.15);padding:.1rem .3rem">`+
			`%s</code>.</p>`,
		rid, execMb, e(c.Phase), e(c.Org), e(c.Industry), e(rid),
	)
	if len(c.ExecBullets) > 0 {
		fmt.Fprintf(&b, `<ul id="acpwb-archive-%s-exec-summary-bullets" style="padding-left:1.25rem;margin-bottom:0">`, rid)
		for _, bu := range c.ExecBullets {
			fmt.Fprintf(&b, `<li style="font-size:.83rem;margin-bottom:.45rem;line-height:1.6">%s</li>`, e(bu))
		}
		b.WriteString(`</ul>`)
	}
	b.WriteString(`</div>`)

	if len(c.Findings) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-key-findings-block" class="mb-4">`+
				`<div class="era-section-head">Key Findings</div>`+
				`<ul id="acpwb-archive-%s-key-findings-list" style="padding-left:1.25rem;margin-bottom:0">`,
			rid, rid,
		)
		nFindings := len(c.Findings)
		for i, f := range c.Findings {
			fmt.Fprintf(&b,
				`<li id="acpwb-archive-%s-finding-%d" data-ref="%s" `+
					`aria-label="Finding %d of %d: %s" `+
					`style="font-size:.88rem;margin-bottom:.6rem;line-height:1.6">%s</li>`,
				rid, i+1, e(f.Ref), i+1, nFindings, e(f.Text), e(f.Text),
			)
		}
		b.WriteString(`</ul></div>`)
	}

	if len(c.EngagementTeam) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-engagement-team" class="mb-4">`+
				`<div class="era-section-head">Engagement Team</div>`+
				`<table style="width:100%%;border-collapse:collapse;font-size:.82rem">`,
			rid,
		)
		for _, m := range c.EngagementTeam {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid rgba(128,128,128,.15)">`+
					`<td style="padding:.35rem .7rem .35rem 0;white-space:nowrap">`+
					`<div style="font-weight:700">%s</div>`+
					`<div style="font-size:.72rem;opacity:.6"><a href="mailto:%s" `+
					`style="color:inherit">%s</a></div></td>`+
					`<td style="padding:.35rem 0;opacity:.75">%s</td></tr>`,
				e(m.Name), e(m.Email), e(m.Email), e(m.Title),
			)
		}
		b.WriteString(`</table></div>`)
	}

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-body-content" class="mb-4" style="font-family:var(--era-font-body);`+
			`line-height:1.85">`,
		rid,
	)
	for _, p := range c.Paragraphs {
		fmt.Fprintf(&b, `<p data-doc="%s">%s</p>`, e(p.Ref), e(p.Text))
	}
	b.WriteString(`</div>`)

	if len(c.MetricRows) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-engagement-metrics-section" class="mb-4">`+
				`<div class="era-section-head">Engagement Metrics</div><div style="overflow-x:auto">`+
				`<table id="acpwb-archive-%s-engagement-metrics-table" `+
				`style="width:100%%;border-collapse:collapse;font-size:.83rem;border:1px solid rgba(128,128,128,.25)">`+
				`<thead><tr class="era-table-head" style="color:#fff">`+
				`<th style="padding:.55rem .85rem;text-align:left;font-size:.66rem;text-transform:uppercase;`+
				`letter-spacing:.06em">Metric</th>`+
				`<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;`+
				`letter-spacing:.06em">Baseline</th>`+
				`<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;`+
				`letter-spacing:.06em">Current</th>`+
				`<th style="padding:.55rem .85rem;text-align:right;font-size:.66rem;text-transform:uppercase;`+
				`letter-spacing:.06em">Change</th></tr></thead><tbody>`,
			rid, rid,
		)
		for i, r := range c.MetricRows {
			color := "#dc2626"
			if r.Positive {
				color = "#16a34a"
			}
			fmt.Fprintf(&b,
				`<tr id="acpwb-archive-%s-metric-row-%d" `+
					`aria-label="Metric: %s, baseline %s, current %s, `+
					`change %s" style="border-top:1px solid rgba(128,128,128,.2)">`+
					`<td style="padding:.45rem .85rem;font-weight:600">%s</td>`+
					`<td style="padding:.45rem .85rem;text-align:right;opacity:.65">%s</td>`+
					`<td style="padding:.45rem .85rem;text-align:right;font-weight:700">%s</td>`+
					`<td style="padding:.45rem .85rem;text-align:right;font-weight:700;`+
					`color:%s">%s</td></tr>`,
				rid, i+1, e(r.Name), e(r.Baseline), e(r.Current), e(r.Delta),
				e(r.Name), e(r.Baseline), e(r.Current), color, e(r.Delta),
			)
		}
		b.WriteString(`</tbody></table></div></div>`)
	}

	if len(c.PercentileTable) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-benchmark-percentile-section" class="mb-4">`+
				`<div class="era-section-head">Market Benchmark — %s</div>`+
				`<div style="overflow-x:auto"><table style="width:100%%;border-collapse:collapse;font-size:.8rem;`+
				`border:1px solid rgba(128,128,128,.25)"><thead><tr class="era-table-head" style="color:#fff">`,
			rid, e(c.PeerGroup),
		)
		for _, h := range []string{"Metric", "P10", "P25", "P33", "P50", "P67", "P75", "P90", "P95"} {
			align := "right"
			if h == "Metric" {
				align = "left"
			}
			fmt.Fprintf(&b,
				`<th style="padding:.45rem .75rem;text-align:%s;`+
					`font-size:.63rem;text-transform:uppercase;letter-spacing:.06em">%s</th>`,
				align, h,
			)
		}
		b.WriteString(`</tr></thead><tbody>`)
		// Per-column style is NOT uniform in the real era template (unlike
		// the main-domain version) — p75 specifically uses opacity:.75,
		// distinct from the .6/.7 used elsewhere.
		for _, r := range c.PercentileTable {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid rgba(128,128,128,.2)">`+
					`<td style="padding:.38rem .75rem;font-weight:600">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.6">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.6">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.7">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;font-weight:700">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.7">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.75">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.6">%s</td>`+
					`<td style="padding:.38rem .75rem;text-align:right;opacity:.6">%s</td></tr>`,
				e(r.Metric), e(r.P10), e(r.P25), e(r.P33), e(r.P50), e(r.P67), e(r.P75), e(r.P90), e(r.P95),
			)
		}
		b.WriteString(`</tbody></table></div></div>`)
	}

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-csv-export-banner" style="background:rgba(128,128,128,.07);`+
			`border:1px solid rgba(128,128,128,.2);padding:.8rem 1rem;margin-bottom:1.25rem;display:flex;`+
			`align-items:center;justify-content:space-between;flex-wrap:wrap;gap:.5rem"><div>`+
			`<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;`+
			`opacity:.65">Raw Data Export</div>`+
			`<div style="font-size:.8rem;font-weight:600">Download the underlying dataset for this archive entry`+
			`</div></div>`+
			`<a id="acpwb-archive-%s-csv-download-link" href="%s" class="era-cta" `+
			`aria-label="Download CSV dataset for archive record %s" `+
			`style="font-size:.72rem;font-weight:700;padding:.3rem .85rem;text-decoration:none;white-space:nowrap">`+
			`&#x2193; Download CSV</a></div>`,
		rid, rid, e(c.ExportCSVURL), e(rid),
	)

	fmt.Fprintf(&b,
		`<a id="acpwb-archive-%s-next-entry-link" href="%s" class="era-cta" `+
			`aria-label="Continue to the next entry in the %d archive series" `+
			`style="display:block;padding:1rem 1.5rem;text-decoration:none;font-weight:700;font-size:.9rem;`+
			`margin-bottom:2rem;font-family:%s,sans-serif">`+
			`Continue Reading: Next in Series &rarr;</a>`,
		rid, e(c.NextEntryURL), c.Year, yd.FontHead,
	)

	if len(c.RelatedPaths) > 0 {
		b.WriteString(`<div class="mt-2"><div class="era-section-head">Related Archive Entries</div><div class="row g-2">`)
		for _, ent := range c.RelatedPaths {
			fmt.Fprintf(&b,
				`<div class="col-md-6"><a href="%s" class="era-entry-card">`+
					`<div style="font-size:.62rem;opacity:.6;font-weight:600;margin-bottom:.2rem">%s</div>`+
					`<div style="font-size:.8rem;font-weight:600;line-height:1.3">%s</div></a></div>`,
				e(ent.URL), e(ent.Date), e(ent.Label),
			)
		}
		b.WriteString(`</div></div>`)
	}

	if len(c.CrossYearReports) > 0 {
		b.WriteString(`<div class="mt-4"><div class="era-section-head">Related Archive Reports — Other Years</div>` +
			`<div class="row g-2">`)
		for _, ent := range c.CrossYearReports {
			fmt.Fprintf(&b,
				`<div class="col-md-6"><a href="%s" class="era-entry-card">`+
					`<div style="font-size:.6rem;opacity:.55;font-weight:700;text-transform:uppercase;`+
					`letter-spacing:.08em;margin-bottom:.2rem">%d Archive</div>`+
					`<div style="font-size:.8rem;font-weight:600;line-height:1.3">%s</div>`+
					`<div style="font-size:.62rem;opacity:.5;margin-top:.2rem">%s</div></a></div>`,
				e(ent.URL), ent.Year, e(ent.Label), e(ent.Date),
			)
		}
		b.WriteString(`</div></div>`)
	}

	if len(c.Revisions) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-revision-history" class="mb-4 mt-4">`+
				`<div class="era-section-head">Revision History</div>`+
				`<table style="width:100%%;border-collapse:collapse;font-size:.78rem"><thead>`+
				`<tr style="opacity:.55">`+
				`<th style="padding:.3rem .5rem .3rem 0;font-size:.62rem;text-transform:uppercase;`+
				`letter-spacing:.06em;text-align:left;white-space:nowrap">Version</th>`+
				`<th style="padding:.3rem .5rem;font-size:.62rem;text-transform:uppercase;letter-spacing:.06em;`+
				`text-align:left;white-space:nowrap">Date</th>`+
				`<th style="padding:.3rem 0 .3rem .5rem;font-size:.62rem;text-transform:uppercase;`+
				`letter-spacing:.06em;text-align:left">Description</th></tr></thead><tbody>`,
			rid,
		)
		for _, r := range c.Revisions {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid rgba(128,128,128,.15)">`+
					`<td style="padding:.35rem .5rem .35rem 0;font-weight:700;white-space:nowrap;`+
					`font-family:monospace;font-size:.72rem">%s</td>`+
					`<td style="padding:.35rem .5rem;opacity:.65;white-space:nowrap">%s</td>`+
					`<td style="padding:.35rem 0 .35rem .5rem;opacity:.8;line-height:1.45">%s `+
					`<span style="opacity:.55">— %s &lt;`+
					`<a href="mailto:%s" style="color:inherit">%s</a>&gt;`+
					`</span></td></tr>`,
				e(r.Version), e(r.Date), e(r.Description), e(r.Author), e(r.AuthorEmail), e(r.AuthorEmail),
			)
		}
		b.WriteString(`</tbody></table></div>`)
	}

	if len(c.Footnotes) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-footnotes" class="mb-4" `+
				`style="border-top:1px solid rgba(128,128,128,.15);padding-top:1rem">`+
				`<div class="era-section-head">Sources &amp; Notes</div>`+
				`<ol style="padding-left:1.25rem;margin-bottom:0">`,
			rid,
		)
		for _, fn := range c.Footnotes {
			fmt.Fprintf(&b,
				`<li id="acpwb-archive-%s-fn-%d" `+
					`style="font-size:.72rem;opacity:.7;margin-bottom:.4rem;line-height:1.5">%s</li>`,
				rid, fn.Num, e(fn.Text),
			)
		}
		b.WriteString(`</ol></div>`)
	}

	fmt.Fprintf(&b,
		`<div class="mt-4 pt-3" style="border-top:1px solid rgba(128,128,128,.2)">`+
			`<a href="%s" class="era-nav-link">&larr; Previous in Series</a></div>`,
		e(c.PrevEntryURL),
	)

	b.WriteString(`<div class="mt-5 pt-3 text-center" style="border-top:1px solid rgba(128,128,128,.12);opacity:.75">`)
	b.WriteString(getArchiveSeal(c.Year, rid))
	b.WriteString(`</div>`)

	b.WriteString(`</div>`) // end col-lg-8

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-sidebar-col" class="col-lg-4 d-none d-lg-block">`+
			`<div style="position:sticky;top:2rem">`,
		rid,
	)
	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-sidebar-record-card" style="background:rgba(128,128,128,.07);`+
			`border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">`+
			`<div class="era-section-head" style="margin-bottom:.6rem">Archive Record</div>`+
			`<dl class="mb-0" style="font-size:.82rem">`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>`+
			`<dd class="fw-700 mb-2">%d-%02d-%02d</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Phase</dt>`+
			`<dd class="fw-700 mb-2" style="text-transform:capitalize">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Sector</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Organization</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Engagement Code</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Version</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Pages</dt>`+
			`<dd class="mb-2" style="font-size:.8rem;opacity:.7">%d (%d KB)</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Distribution</dt>`+
			`<dd class="mb-2" style="font-size:.72rem;opacity:.65;line-height:1.4">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>`+
			`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd></dl></div>`,
		rid, c.Year, c.Month, c.Day, e(c.Phase), e(c.Industry), e(c.Org),
		e(c.EngCode), e(c.DocVersion), c.PageCount, c.FileSizeKB, e(c.Distribution), e(rid),
	)

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-sidebar-nav-card" style="background:rgba(128,128,128,.07);`+
			`border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">`+
			`<div class="era-section-head" style="margin-bottom:.6rem">Navigation</div>`+
			`<ul class="list-unstyled mb-0" style="font-size:.82rem">`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-year" href="%s" class="era-nav-link" `+
			`aria-label="Browse all %d archive records">&larr; All %d Records</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-month" href="%s" `+
			`aria-label="Browse %d month %d archive records" `+
			`style="color:inherit;opacity:.7;text-decoration:none">&larr; %d/%02d</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-prev" href="%s" `+
			`aria-label="Navigate to the previous entry in the archive series" `+
			`style="color:inherit;opacity:.55;text-decoration:none">&larr; Previous Entry</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-next" href="%s" class="era-nav-link" `+
			`aria-label="Navigate to the next entry in the %d archive series">Next in Series &rarr;</a></li>`+
			`</ul></div>`,
		rid,
		rid, e(c.YearURL), c.Year, c.Year,
		rid, e(c.MonthURL), c.Year, c.Month, c.Year, c.Month,
		rid, e(c.PrevEntryURL),
		rid, e(c.NextEntryURL), c.Year,
	)

	if len(c.RelatedDocs) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-related-docs" style="background:rgba(128,128,128,.07);`+
				`border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">`+
				`<div class="era-section-head" style="margin-bottom:.6rem">Related Documents</div>`,
			rid,
		)
		for _, d := range c.RelatedDocs {
			fmt.Fprintf(&b,
				`<a href="%s" class="era-entry-card" style="margin-bottom:.5rem">`+
					`<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">%s &bull; `+
					`%s</div>`+
					`<div style="font-size:.75rem;font-weight:600;line-height:1.35">%s</div></a>`,
				e(d.URL), e(d.Date), e(d.Phase), e(d.Label),
			)
		}
		b.WriteString(`</div>`)
	}

	if len(c.RelatedPolicy) > 0 {
		b.WriteString(`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;` +
			`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">` +
			`Related Public Policy</div>`)
		for _, s := range c.RelatedPolicy {
			fmt.Fprintf(&b,
				`<a href="https://acpwb.com%s" style="display:block;background:rgba(128,128,128,.05);`+
					`border:1px solid rgba(128,128,128,.2);border-left:3px solid var(--era-accent2);`+
					`padding:.55rem .75rem;margin-bottom:.5rem;text-decoration:none;color:inherit">`+
					`<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">%s &bull; `+
					`%s</div>`+
					`<div style="font-size:.74rem;font-weight:600;line-height:1.35">%s</div>`+
					`<div style="font-size:.65rem;opacity:.55;margin-top:.15rem">%s</div></a>`,
				e(s.URL), e(s.FilingDate), e(s.DocumentType), e(s.Title), e(s.AgencyAcronym),
			)
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(eraSidebarPresentationsHTML(c.RelatedPresentations, yd))

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-year-browser" style="background:rgba(128,128,128,.07);`+
			`border:1px solid rgba(128,128,128,.2);padding:1rem">`+
			`<div class="era-section-head" style="margin-bottom:.6rem">Browse by Year</div>`+
			`<div style="display:flex;flex-wrap:wrap;gap:.3rem">`,
		rid,
	)
	for _, y := range c.AllYears {
		extra := ""
		if y == c.Year {
			extra = ";font-weight:800;border-color:var(--era-accent);color:var(--era-accent)"
		}
		fmt.Fprintf(&b,
			`<a id="acpwb-archive-%s-year-link-%d" href="https://archives-%d.acpwb.com/" `+
				`aria-label="Browse the %d archive" `+
				`style="font-size:.68rem;padding:.2rem .4rem;border:1px solid rgba(128,128,128,.3);`+
				`text-decoration:none;color:inherit%s">%d</a>`,
			rid, y, y, y, extra, y,
		)
	}
	b.WriteString(`</div></div>`)

	b.WriteString(`</div></div>`)       // sticky, sidebar-col
	b.WriteString(`</div></div></div>`) // row, container, era-archive-content

	b.WriteString(eraBulkHexScriptHTML(c.BulkHexJS, rid))
	return b.String()
}
