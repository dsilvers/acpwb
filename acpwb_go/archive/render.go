package archive

import (
	"fmt"
	"strings"
)

// e is a short alias matching the Python source's `from apps.core.htmlgen import escape as e`.
func e(s string) string { return escape(s) }

func relatedReportsHTML(c *Context) string {
	if len(c.RelatedReports) == 0 {
		return ""
	}
	var items strings.Builder
	for _, r := range c.RelatedReports {
		fmt.Fprintf(&items,
			`<div style="background:white;border:1px solid var(--border);border-left:3px solid var(--gold);`+
				`padding:.75rem 1rem;margin-bottom:.75rem">`+
				`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;`+
				`color:var(--muted);margin-bottom:.2rem">`+
				`%s &bull; %s &bull; %s</div>`+
				`<a href="%s" style="font-size:.9rem;font-weight:700;`+
				`color:var(--navy);text-decoration:none">%s</a>`+
				`<p class="small text-muted mb-0" style="font-size:.78rem;margin-top:.25rem">`+
				`%s</p></div>`,
			e(r.Category), e(strings.ToUpper(r.FileType)), e(r.PubDateDisplay),
			e(r.DetailURL), e(r.Title),
			e(truncatechars(r.Summary, 160)),
		)
	}
	return `<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem">` +
		`Referenced Research &amp; Publications</h5>` + items.String() + `</div><hr class="gold-divider">`
}

func sidebarRelatedDocsHTML(c *Context, rid string) string {
	if len(c.RelatedDocs) == 0 {
		return ""
	}
	var items strings.Builder
	for _, d := range c.RelatedDocs {
		fmt.Fprintf(&items,
			`<a href="%s" style="display:block;background:#f4f6f9;border:1px solid var(--border);`+
				`border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">`+
				`<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">%s &bull; %s</div>`+
				`<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">%s</div></a>`,
			e(d.URL), e(d.Date), e(d.Phase), e(d.Label),
		)
	}
	return fmt.Sprintf(
		`<div id="acpwb-archive-%s-related-docs" class="acpwb-card mb-4">`+
			`<h6 class="card-title mb-3">Related Documents</h6>%s</div>`,
		rid, items.String(),
	)
}

func sidebarRelatedPolicyHTML(c *Context) string {
	if len(c.RelatedPolicy) == 0 {
		return ""
	}
	var items strings.Builder
	for _, s := range c.RelatedPolicy {
		fmt.Fprintf(&items,
			`<a href="%s" style="display:block;background:#f4f6f9;border:1px solid var(--border);`+
				`border-left:3px solid var(--gold);padding:.55rem .8rem;margin-bottom:.5rem;text-decoration:none">`+
				`<div style="font-size:.6rem;color:var(--muted);margin-bottom:.15rem">%s &bull; %s</div>`+
				`<div style="font-size:.75rem;font-weight:600;color:var(--navy);line-height:1.35">%s</div>`+
				`<div style="font-size:.65rem;color:var(--muted);margin-top:.15rem">%s</div></a>`,
			e(s.URL), e(s.FilingDate), e(s.DocumentType), e(s.Title), e(s.AgencyAcronym),
		)
	}
	return `<div class="acpwb-card mb-4"><h6 class="card-title mb-3">Related Public Policy</h6>` + items.String() + `</div>`
}

func sidebarRelatedPresentationsHTML(c *Context) string {
	if len(c.RelatedPresentations) == 0 {
		return ""
	}
	var items strings.Builder
	for _, p := range c.RelatedPresentations {
		items.WriteString(renderPresCard(p, "https://acpwb.com"))
	}
	return `<div class="acpwb-archive-pres-sidebar acpwb-card mb-4">` +
		`<h6 class="card-title mb-3">Related Presentations</h6>` + items.String() + `</div>`
}

func sidebarYearBrowserHTML(c *Context, rid string) string {
	var items strings.Builder
	for _, y := range c.ArchiveYears {
		fmt.Fprintf(&items,
			`<a id="acpwb-archive-%s-year-link-%d" href="https://archives-%d.acpwb.com/" `+
				`aria-label="Browse the %d archive" `+
				`style="font-size:.72rem;padding:.2rem .45rem;background:#f4f6f9;border:1px solid var(--border);`+
				`color:var(--navy);text-decoration:none">%d</a>`,
			rid, y, y, y, y,
		)
	}
	return fmt.Sprintf(
		`<div id="acpwb-archive-%s-year-browser" class="acpwb-card mb-4">`+
			`<h6 class="card-title mb-3">Browse by Year</h6>`+
			`<div style="display:flex;flex-wrap:wrap;gap:.35rem">%s</div></div>`,
		rid, items.String(),
	)
}

func bulkHexScriptHTML(c *Context, rid string) string {
	var cssVars strings.Builder
	for i, h := range c.BulkHexCSS {
		fmt.Fprintf(&cssVars, "  --acpwb-r%03d: %s;\n", i, h)
	}
	var jsRefs strings.Builder
	for i, h := range c.BulkHexJS {
		fmt.Fprintf(&jsRefs, `var _acpwbRef%d="%s";`+"\n", i+1, h)
	}
	var jsFuncs strings.Builder
	limit := len(c.BulkHexJS)
	if limit > 50 {
		limit = 50
	}
	for _, h := range c.BulkHexJS[:limit] {
		fmt.Fprintf(&jsFuncs, `function _acpwbArchiveRecordEntryMetadataLookup_%s(){return "%s";}`+"\n", h, rid)
	}
	return fmt.Sprintf(
		"<style>\n  :root {\n%s  }\n</style>\n<script>\n/* ACPWB archive index — %s */\n(function(){\n%s%s})();\n</script>\n",
		cssVars.String(), rid, jsRefs.String(), jsFuncs.String(),
	)
}

// RenderArchiveDefault ports
// pyrender/archive_main.py:render_archive_default (the main-domain branch of
// templates/honeypot/archive.html).
func RenderArchiveDefault(c *Context) string {
	rid := c.RecordID
	var b strings.Builder

	b.WriteString(`<section class="page-banner"><div class="container">`)
	fmt.Fprintf(&b,
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;`+
			`letter-spacing:.18em;font-size:.75rem">`+
			`<a href="/archive/" style="color:var(--gold)">Archive</a>`+
			` &rsaquo; <a href="%s" style="color:var(--gold)">%d</a>`+
			` &rsaquo; <a href="%s" style="color:var(--gold)">%02d</a>`+
			` &rsaquo; %02d</p>`,
		e(c.YearURL), c.Year, e(c.MonthURL), c.Month, c.Day,
	)
	fmt.Fprintf(&b, `<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">%s</h1>`, e(c.Title))
	fmt.Fprintf(&b,
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">`+
			`%s &bull; %s &bull; %s phase`+
			` &bull; Archived %d-%02d-%02d</p>`,
		e(c.Industry), e(c.Org), e(pyCapitalize(c.Phase)), c.Year, c.Month, c.Day,
	)
	b.WriteString(`</div></section>`)

	b.WriteString(`<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">`)
	fmt.Fprintf(&b, `<div id="acpwb-archive-%s-primary-content-col" class="col-lg-8">`, rid)

	execMB := "0"
	if len(c.ExecBullets) > 0 {
		execMB = ".75rem"
	}
	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-executive-summary-callout" `+
			`style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);`+
			`padding:1.25rem 1.5rem;margin-bottom:2rem">`+
			`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--gold);margin-bottom:.5rem">Executive Summary</div>`+
			`<p style="font-size:.88rem;color:var(--navy);margin-bottom:%s;font-weight:500;line-height:1.6">`+
			`This archive entry documents ACPWB's <strong>%s</strong> phase engagement with `+
			`<strong>%s</strong> in the <strong>%s</strong> sector.`+
			` The record ID <code style="font-size:.8rem;background:#f4f6f9;padding:.1rem .35rem">%s</code>`+
			` uniquely identifies this documentation set within ACPWB's institutional archive.</p>`,
		rid, execMB, e(c.Phase), e(c.Org), e(c.Industry), e(rid),
	)
	if len(c.ExecBullets) > 0 {
		fmt.Fprintf(&b, `<ul id="acpwb-archive-%s-exec-summary-bullets" style="padding-left:1.25rem;margin-bottom:0">`, rid)
		for _, bul := range c.ExecBullets {
			fmt.Fprintf(&b, `<li style="font-size:.83rem;color:#333;margin-bottom:.45rem;line-height:1.6">%s</li>`, e(bul))
		}
		b.WriteString(`</ul>`)
	}
	b.WriteString(`</div>`)

	if len(c.Findings) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-key-findings-block" class="mb-4">`+
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">`+
				`Key Findings</h5>`+
				`<ul id="acpwb-archive-%s-key-findings-list" style="padding-left:1.25rem;margin-bottom:0">`,
			rid, rid,
		)
		n := len(c.Findings)
		for i, f := range c.Findings {
			fmt.Fprintf(&b,
				`<li id="acpwb-archive-%s-finding-%d" data-ref="%s" `+
					`aria-label="Finding %d of %d: %s" `+
					`style="font-size:.88rem;color:#333;margin-bottom:.7rem;line-height:1.6">%s</li>`,
				rid, i+1, e(f.Ref), i+1, n, e(f.Text), e(f.Text),
			)
		}
		b.WriteString(`</ul></div>`)
	}

	if len(c.EngagementTeam) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-engagement-team" class="mb-4">`+
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">`+
				`Engagement Team</h5><table style="width:100%%;border-collapse:collapse;font-size:.83rem">`,
			rid,
		)
		for _, m := range c.EngagementTeam {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid var(--border)">`+
					`<td style="padding:.4rem .8rem .4rem 0;white-space:nowrap">`+
					`<div style="font-weight:700;color:var(--navy)">%s</div>`+
					`<div style="font-size:.72rem;color:var(--muted)">`+
					`<a href="mailto:%s" style="color:inherit">%s</a></div></td>`+
					`<td style="padding:.4rem 0;color:var(--muted)">%s</td></tr>`,
				e(m.Name), e(m.Email), e(m.Email), e(m.Title),
			)
		}
		b.WriteString(`</table></div>`)
	}

	fmt.Fprintf(&b, `<div id="acpwb-archive-%s-body-content" class="wiki-content mb-4">`, rid)
	for _, p := range c.Paragraphs {
		fmt.Fprintf(&b, `<p data-doc="%s">%s</p>`, e(p.Ref), e(p.Text))
	}
	b.WriteString(`</div>`)

	if len(c.MetricRows) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-engagement-metrics-section" class="mb-4">`+
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">`+
				`Engagement Metrics</h5><div style="overflow-x:auto">`+
				`<table id="acpwb-archive-%s-engagement-metrics-table" `+
				`style="width:100%%;border-collapse:collapse;background:white;border:1px solid var(--border);`+
				`font-size:.83rem"><thead><tr style="background:var(--navy);color:var(--gold)">`+
				`<th style="padding:.6rem .9rem;text-align:left;font-size:.68rem;font-weight:700;`+
				`text-transform:uppercase;letter-spacing:.06em;white-space:nowrap">Metric</th>`+
				`<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;`+
				`text-transform:uppercase;letter-spacing:.06em">Baseline</th>`+
				`<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;`+
				`text-transform:uppercase;letter-spacing:.06em">Current</th>`+
				`<th style="padding:.6rem .9rem;text-align:right;font-size:.68rem;font-weight:700;`+
				`text-transform:uppercase;letter-spacing:.06em">Change</th></tr></thead><tbody>`,
			rid, rid,
		)
		for i, r := range c.MetricRows {
			color := "#dc2626"
			if r.Positive {
				color = "#15803d"
			}
			fmt.Fprintf(&b,
				`<tr id="acpwb-archive-%s-metric-row-%d" `+
					`aria-label="Metric: %s, baseline %s, current %s, `+
					`change %s" style="border-top:1px solid var(--border)">`+
					`<td style="padding:.5rem .9rem;color:var(--navy);font-weight:600">%s</td>`+
					`<td style="padding:.5rem .9rem;text-align:right;color:var(--muted)">%s</td>`+
					`<td style="padding:.5rem .9rem;text-align:right;font-weight:700">%s</td>`+
					`<td style="padding:.5rem .9rem;text-align:right;font-weight:700;`+
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
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">`+
				`Market Benchmark — %s</h5><div style="overflow-x:auto">`+
				`<table style="width:100%%;border-collapse:collapse;background:white;border:1px solid var(--border);`+
				`font-size:.82rem"><thead><tr style="background:var(--navy);color:var(--gold)">`,
			rid, e(c.PeerGroup),
		)
		headers := []string{"Metric", "P10", "P25", "P33", "P50", "P67", "P75", "P90", "P95"}
		for _, h := range headers {
			align := "right"
			if h == "Metric" {
				align = "left"
			}
			fmt.Fprintf(&b,
				`<th style="padding:.55rem .9rem;text-align:%s;`+
					`font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.06em">%s</th>`,
				align, h,
			)
		}
		b.WriteString(`</tr></thead><tbody>`)
		for _, r := range c.PercentileTable {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid var(--border)">`+
					`<td style="padding:.45rem .9rem;color:var(--navy);font-weight:600">%s</td>`,
				e(r.Metric),
			)
			cols := []struct {
				key   string
				value string
			}{
				{"p10", r.P10}, {"p25", r.P25}, {"p33", r.P33}, {"p50", r.P50},
				{"p67", r.P67}, {"p75", r.P75}, {"p90", r.P90}, {"p95", r.P95},
			}
			for _, col := range cols {
				style := "color:var(--muted)"
				if col.key == "p50" {
					style = "font-weight:700"
				}
				fmt.Fprintf(&b, `<td style="padding:.45rem .9rem;text-align:right;%s">%s</td>`, style, e(col.value))
			}
			b.WriteString(`</tr>`)
		}
		b.WriteString(`</tbody></table></div></div>`)
	}

	b.WriteString(`<hr class="gold-divider">`)
	b.WriteString(relatedReportsHTML(c))

	b.WriteString(`<div style="background:var(--surface);border:1px solid var(--border);padding:.85rem 1rem;` +
		`margin-bottom:1.25rem;display:flex;align-items:center;justify-content:space-between;` +
		`flex-wrap:wrap;gap:.5rem"><div>` +
		`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;` +
		`color:var(--muted)">Raw Data Export</div>` +
		`<div style="font-size:.8rem;color:var(--navy);font-weight:600">` +
		`Download the underlying dataset for this archive entry</div></div>`)
	fmt.Fprintf(&b,
		`<a href="%s" style="font-size:.72rem;font-weight:700;padding:.3rem .85rem;`+
			`background:var(--navy);color:var(--gold);text-decoration:none;white-space:nowrap">`+
			`&#x2193; Download CSV</a></div>`,
		e(c.ExportCSVURL),
	)
	fmt.Fprintf(&b,
		`<a href="%s" style="display:block;background:var(--navy);color:white;`+
			`padding:1rem 1.5rem;text-decoration:none;font-weight:700;font-size:.9rem;margin-bottom:2rem">`+
			`Continue Reading: Next in Series &rarr;</a>`,
		e(c.NextEntryURL),
	)

	b.WriteString(`<div class="mt-2"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.85rem;margin-bottom:1rem">Related Archive Entries</h5>` +
		`<div class="row g-2">`)
	for _, ent := range c.RelatedPaths {
		fmt.Fprintf(&b,
			`<div class="col-md-6"><a href="%s" style="display:block;background:white;`+
				`border:1px solid var(--border);padding:.75rem 1rem;text-decoration:none">`+
				`<div style="font-size:.62rem;color:var(--muted);font-weight:600;margin-bottom:.2rem">`+
				`%s</div>`+
				`<div style="font-size:.8rem;color:var(--navy);font-weight:600;line-height:1.3">%s`+
				`</div></a></div>`,
			e(ent.URL), e(ent.Date), e(ent.Label),
		)
	}
	b.WriteString(`</div></div>`)

	if len(c.CrossYearReports) > 0 {
		b.WriteString(`<div class="mt-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
			`letter-spacing:.06em;font-size:.85rem;margin-bottom:1rem">` +
			`Related Archive Reports — Other Years</h5><div class="row g-2">`)
		for _, ent := range c.CrossYearReports {
			fmt.Fprintf(&b,
				`<div class="col-md-6"><a href="%s" style="display:block;background:white;`+
					`border:1px solid var(--border);border-left:3px solid var(--gold);padding:.75rem 1rem;`+
					`text-decoration:none">`+
					`<div style="font-size:.6rem;color:var(--gold);font-weight:800;text-transform:uppercase;`+
					`letter-spacing:.08em;margin-bottom:.2rem">%d Archive</div>`+
					`<div style="font-size:.8rem;color:var(--navy);font-weight:600;line-height:1.3">%s</div>`+
					`<div style="font-size:.62rem;color:var(--muted);margin-top:.2rem">%s</div></a></div>`,
				e(ent.URL), ent.Year, e(ent.Label), e(ent.Date),
			)
		}
		b.WriteString(`</div></div>`)
	}

	if len(c.Revisions) > 0 {
		fmt.Fprintf(&b,
			`<div id="acpwb-archive-%s-revision-history" class="mb-4 mt-4">`+
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)">`+
				`Revision History</h5><table style="width:100%%;border-collapse:collapse;font-size:.8rem">`+
				`<thead><tr style="color:var(--muted)">`+
				`<th style="padding:.3rem .6rem .3rem 0;font-size:.65rem;text-transform:uppercase;`+
				`letter-spacing:.06em;text-align:left;white-space:nowrap">Version</th>`+
				`<th style="padding:.3rem .6rem;font-size:.65rem;text-transform:uppercase;letter-spacing:.06em;`+
				`text-align:left;white-space:nowrap">Date</th>`+
				`<th style="padding:.3rem 0 .3rem .6rem;font-size:.65rem;text-transform:uppercase;`+
				`letter-spacing:.06em;text-align:left">Description</th></tr></thead><tbody>`,
			rid,
		)
		for _, r := range c.Revisions {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid var(--border)">`+
					`<td style="padding:.38rem .6rem .38rem 0;font-weight:700;color:var(--navy);white-space:nowrap;`+
					`font-family:monospace;font-size:.75rem">%s</td>`+
					`<td style="padding:.38rem .6rem;color:var(--muted);white-space:nowrap">%s</td>`+
					`<td style="padding:.38rem 0 .38rem .6rem;color:#444;line-height:1.45">%s `+
					`<span style="color:var(--muted)">— %s &lt;`+
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
				`style="border-top:1px solid var(--border);padding-top:1rem">`+
				`<h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;`+
				`font-size:.8rem;margin-bottom:.75rem">Sources &amp; Notes</h5>`+
				`<ol style="padding-left:1.25rem;margin-bottom:0">`,
			rid,
		)
		for _, fn := range c.Footnotes {
			fmt.Fprintf(&b,
				`<li id="acpwb-archive-%s-fn-%d" `+
					`style="font-size:.72rem;color:var(--muted);margin-bottom:.4rem;line-height:1.5">%s</li>`,
				rid, fn.Num, e(fn.Text),
			)
		}
		b.WriteString(`</ol></div>`)
	}

	fmt.Fprintf(&b,
		`<div class="mt-4 pt-3" style="border-top:1px solid var(--border)">`+
			`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
			`text-decoration:none">&larr; Previous in Series</a></div>`,
		e(c.PrevEntryURL),
	)

	b.WriteString(`<div class="mt-5 pt-3 text-center" style="border-top:1px solid var(--border);opacity:.75">`)
	b.WriteString(getArchiveSeal(c.Year, rid))
	b.WriteString(`</div>`)

	b.WriteString(`</div>`) // end col-lg-8

	fmt.Fprintf(&b, `<div id="acpwb-archive-%s-sidebar-col" class="col-lg-4 d-none d-lg-block">`+
		`<div style="position:sticky;top:2rem">`, rid)

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-sidebar-record-card" class="acpwb-card mb-4">`+
			`<h6 class="card-title mb-2">Archive Record</h6><dl class="mb-0" style="font-size:.82rem">`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>`+
			`<dd class="fw-700 mb-2">%d-%02d-%02d</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Phase</dt>`+
			`<dd class="fw-700 mb-2" style="text-transform:capitalize">%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Sector</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
			`Organization</dt><dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
			`Engagement Code</dt><dd class="mb-2" style="font-family:monospace;font-size:.75rem;`+
			`color:var(--muted)">%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Version</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;color:var(--muted)">`+
			`%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Pages</dt>`+
			`<dd class="mb-2" style="font-size:.8rem;color:var(--muted)">`+
			`%d (%d KB)</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
			`Distribution</dt><dd class="mb-2" style="font-size:.72rem;color:var(--muted);line-height:1.4">`+
			`%s</dd>`+
			`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
			`Record ID</dt><dd class="mb-0 text-muted" style="font-family:monospace;font-size:.75rem">`+
			`%s</dd></dl></div>`,
		rid, c.Year, c.Month, c.Day, e(c.Phase), e(c.Industry), e(c.Org), e(c.EngCode), e(c.DocVersion),
		c.PageCount, c.FileSizeKB, e(c.Distribution), e(rid),
	)

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-sidebar-nav-card" class="acpwb-card mb-4">`+
			`<h6 class="card-title mb-3">Archive Navigation</h6><ul class="list-unstyled mb-0" style="font-size:.82rem">`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-year" href="%s" `+
			`aria-label="Browse all %d archive records" style="color:var(--navy)">`+
			`&#8592; All %d Records</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-month" href="%s" `+
			`aria-label="Browse %d month %d archive records" style="color:var(--navy)">`+
			`&#8592; %d/%02d Records</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-prev" href="%s" `+
			`aria-label="Navigate to the previous entry in the archive series" style="color:var(--muted)">`+
			`&#8592; Previous Entry</a></li>`+
			`<li class="mb-2"><a id="acpwb-archive-%s-nav-next" href="%s" `+
			`aria-label="Navigate to the next entry in the %d archive series" `+
			`style="color:var(--gold);font-weight:700">Next in Series &rarr;</a></li></ul></div>`,
		rid,
		rid, e(c.YearURL), c.Year, c.Year,
		rid, e(c.MonthURL), c.Year, c.Month, c.Year, c.Month,
		rid, e(c.PrevEntryURL),
		rid, e(c.NextEntryURL), c.Year,
	)

	b.WriteString(sidebarRelatedDocsHTML(c, rid))
	b.WriteString(sidebarRelatedPolicyHTML(c))
	b.WriteString(sidebarRelatedPresentationsHTML(c))
	b.WriteString(sidebarYearBrowserHTML(c, rid))

	fmt.Fprintf(&b,
		`<div id="acpwb-archive-%s-research-division-card" class="acpwb-card">`+
			`<h6 class="card-title mb-3">Research Division</h6>`+
			`<p class="small text-muted mb-2" style="font-size:.8rem">`+
			`ACPWB's document archive spans our full operational history from 1985 to present.</p>`+
			`<a href="/reports/" style="font-size:.8rem;color:var(--gold);font-weight:700;`+
			`text-decoration:none">Browse Research Reports &rarr;</a></div>`,
		rid,
	)

	b.WriteString(`</div></div>`)           // end sticky, sidebar col
	b.WriteString(`</div></div></section>`) // end row, container, section

	b.WriteString(bulkHexScriptHTML(c, rid))
	return b.String()
}

// renderPresCard ports apps/core/htmlgen.py:render_pres_card.
func renderPresCard(pres Presentation, urlPrefix string) string {
	presURL := urlPrefix + pres.PresURL
	thumbStyle := fmt.Sprintf("background-color:%s;", pres.Theme.Bg)
	if pres.ThumbBg != "" {
		thumbStyle = fmt.Sprintf(
			"background-image:url('%s');background-size:cover;background-position:center;",
			staticURL(pres.ThumbBg),
		)
	}

	var authorsAvatars strings.Builder
	for _, a := range pres.Authors {
		authorsAvatars.WriteString(headshotOrAvatar(a.AvatarSeed, a.Initials, 24))
	}
	names := make([]string, len(pres.Authors))
	for i, a := range pres.Authors {
		names[i] = e(a.FullName)
	}
	authorsNames := strings.Join(names, ", ")

	return fmt.Sprintf(
		`<div class="pres-card">`+
			`<a href="%s" style="text-decoration:none">`+
			`<div class="pres-card-thumb" style="%s">`+
			`<div style="position:absolute;top:0.6em;right:0.7em;z-index:2;`+
			`background:%s;color:%s;font-size:.62rem;`+
			`padding:.2em .5em;border-radius:2px;font-weight:800;font-family:system-ui,sans-serif">`+
			`%d slides</div></div></a>`+
			`<div class="pres-card-body">`+
			`<div class="pres-card-org" style="display:flex;align-items:center;gap:.4em">`+
			`%s%s</div>`+
			`<a href="%s" class="pres-card-title" style="text-decoration:none">`+
			`%s</a>`+
			`<div class="pres-card-meta">%s &mdash; %s</div>`+
			`<div class="pres-card-authors">%s`+
			`<span style="font-size:.7rem;color:#555;margin-left:.3em">%s</span></div>`+
			`</div></div>`,
		e(presURL), thumbStyle,
		pres.Theme.Accent, pres.Theme.Bg,
		pres.SlideCount,
		orgLogo(pres.OrgSlug), e(pres.OrgName),
		e(presURL), e(truncatewords(pres.Title, 12)),
		e(pres.PubDateDisplay), e(pres.Industry),
		authorsAvatars.String(), authorsNames,
	)
}
