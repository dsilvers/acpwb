package archive

import (
	"fmt"
	"strings"
)

// RenderComplianceDefaultEra ports
// pyrender/archive_era.py:render_compliance_default_era
// (templates/jinja2/honeypot/era/archive_compliance.html).
func RenderComplianceDefaultEra(c *EraComplianceContext) string {
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
		`<p class="mb-1" style="font-size:.68rem;font-weight:700;letter-spacing:.14em;text-transform:uppercase;`+
			`opacity:.7">Compliance Review &bull; Audit Ref: %s</p>`,
		e(c.AuditRef),
	)
	fmt.Fprintf(&b,
		`<h1 style="font-family:var(--era-font-head);font-size:clamp(1.1rem,3vw,1.9rem);line-height:1.25;`+
			`margin-bottom:.3rem">%s</h1>`,
		e(c.Title),
	)
	fmt.Fprintf(&b,
		`<p style="opacity:.75;font-size:.88rem;margin-bottom:0">`+
			`%s &bull; %s &bull; %s</p>`,
		e(c.Industry), e(c.Org), e(c.DateStr),
	)
	b.WriteString(`</div></div>`)

	b.WriteString(`<div class="era-archive-content"><div class="container"><div class="row g-4"><div class="col-lg-8">`)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-header" class="era-callout" style="margin-bottom:1.5rem">`+
			`<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--era-accent);margin-bottom:.6rem">Document Information</div>`+
			`<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.8rem">`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Client</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Industry</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Audit Ref</span><br><code style="font-size:.75rem">%s</code></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Date</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Version</span><br><code style="font-size:.75rem">%s</code></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Record ID</span><br><code style="font-size:.75rem;opacity:.7">%s</code></div></div>`+
			`<div style="margin-top:.75rem;padding-top:.6rem;border-top:1px solid rgba(128,128,128,.18);`+
			`font-size:.78rem"><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;`+
			`letter-spacing:.06em">Filed by</span>&nbsp; <strong>%s</strong>, `+
			`%s &mdash; <a href="mailto:%s" `+
			`style="color:inherit;opacity:.7">%s</a></div>`,
		rid, e(c.Org), e(c.Industry), e(c.AuditRef), e(c.DateStr), e(c.DocVersion), e(rid),
		e(c.Assessor), e(c.AssessorTitle), e(c.AssessorEmail), e(c.AssessorEmail),
	)
	if len(c.FrameworksCited) > 0 {
		b.WriteString(`<div style="margin-top:.6rem;display:flex;flex-wrap:wrap;gap:.3rem">`)
		for _, fw := range c.FrameworksCited {
			fmt.Fprintf(&b,
				`<span style="font-size:.62rem;padding:.15rem .45rem;background:rgba(128,128,128,.12);`+
					`border:1px solid rgba(128,128,128,.25)">%s</span>`,
				e(fw),
			)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-scope" class="mb-4"><div class="era-section-head">`+
			`1. Engagement Scope</div>`+
			`<p style="font-size:.88rem;line-height:1.7;margin-bottom:0">%s</p></div>`,
		rid, e(c.ScopePara),
	)
	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-methodology" class="mb-4"><div class="era-section-head">`+
			`2. Methodology</div>`+
			`<p style="font-size:.88rem;line-height:1.7;margin-bottom:0">%s</p></div>`,
		rid, e(c.MethodPara),
	)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-findings-summary" class="mb-4">`+
			`<div class="era-section-head">3. Findings Summary</div>`+
			`<div style="overflow-x:auto"><table style="width:100%%;border-collapse:collapse;font-size:.8rem">`+
			`<thead><tr class="era-table-head" style="color:#fff">`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase;white-space:nowrap">Finding ID</th>`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Risk</th>`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Status</th>`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Owner</th>`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Due</th>`+
			`<th style="padding:.45rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Description</th></tr></thead><tbody>`,
		rid,
	)
	for _, f := range c.Findings {
		fmt.Fprintf(&b,
			`<tr style="border-top:1px solid rgba(128,128,128,.15)">`+
				`<td style="padding:.4rem .7rem;font-family:monospace;font-size:.72rem;white-space:nowrap">`+
				`%s</td>`+
				`<td style="padding:.4rem .7rem;white-space:nowrap">%s</td>`+
				`<td style="padding:.4rem .7rem;white-space:nowrap">%s</td>`+
				`<td style="padding:.4rem .7rem;font-size:.75rem;opacity:.8">%s</td>`+
				`<td style="padding:.4rem .7rem;font-size:.72rem;white-space:nowrap;opacity:.7">%s</td>`+
				`<td style="padding:.4rem .7rem;font-size:.78rem;max-width:280px">%s</td></tr>`,
			e(f.ID), riskBadge(f.Risk, "INFO"), statusBadge(f.Status), e(f.Owner), e(f.DueDate), e(f.Description),
		)
	}
	b.WriteString(`</tbody></table></div></div>`)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-findings-detail" class="mb-4">`+
			`<div class="era-section-head">4. Detailed Findings</div>`,
		rid,
	)
	for _, f := range c.Findings {
		fmt.Fprintf(&b,
			`<div id="acpwb-compliance-%s-finding-%s" style="background:rgba(128,128,128,.06);`+
				`border:1px solid rgba(128,128,128,.18);margin-bottom:1rem;padding:1rem 1.1rem">`+
				`<div style="display:flex;align-items:center;gap:.5rem;margin-bottom:.6rem">`+
				`<code style="font-size:.72rem;opacity:.75">%s</code>`+
				`%s`+
				`%s</div>`+
				`<div style="margin-bottom:.55rem">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`opacity:.5;margin-bottom:.2rem">Observation</div>`+
				`<p style="font-size:.83rem;line-height:1.65;margin-bottom:0">%s</p></div>`+
				`<div style="margin-bottom:.55rem;padding-top:.5rem;border-top:1px solid rgba(128,128,128,.12)">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`opacity:.5;margin-bottom:.2rem">Corrective Action Required</div>`+
				`<p style="font-size:.83rem;line-height:1.65;margin-bottom:0">%s</p></div>`+
				`<div style="padding-top:.5rem;border-top:1px solid rgba(128,128,128,.12)">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`opacity:.5;margin-bottom:.2rem">Management Response</div>`+
				`<p style="font-size:.83rem;line-height:1.65;margin-bottom:.45rem">%s</p>`+
				`<div style="font-size:.72rem;opacity:.65">Owner: <strong>%s</strong> &bull; `+
				`Due: <strong>%s</strong></div></div></div>`,
			rid, f.ID, e(f.ID), riskBadge(f.Risk, "INFORMATIONAL"), statusBadge(f.Status),
			e(f.Description), e(f.CorrectiveAction), e(f.MgmtResponse), e(f.Owner), e(f.DueDate),
		)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-certification" class="mb-4">`+
			`<div class="era-section-head">5. Distribution &amp; Certification</div>`,
		rid,
	)
	if len(c.DistList) > 0 {
		b.WriteString(`<table style="width:100%;border-collapse:collapse;font-size:.8rem;margin-bottom:1.25rem">` +
			`<thead><tr class="era-table-head" style="color:#fff">` +
			`<th style="padding:.35rem .7rem;text-align:left;font-size:.68rem;text-transform:uppercase;` +
			`letter-spacing:.06em">Recipient</th>` +
			`<th style="padding:.35rem .7rem;text-align:left;font-size:.68rem;text-transform:uppercase;` +
			`letter-spacing:.06em">Title</th>` +
			`<th style="padding:.35rem .7rem;text-align:left;font-size:.68rem;text-transform:uppercase;` +
			`letter-spacing:.06em">Email</th></tr></thead><tbody>`)
		for _, d := range c.DistList {
			fmt.Fprintf(&b,
				`<tr style="border-top:1px solid rgba(128,128,128,.15)">`+
					`<td style="padding:.35rem .7rem;font-weight:600">%s</td>`+
					`<td style="padding:.35rem .7rem;opacity:.8">%s</td>`+
					`<td style="padding:.35rem .7rem;font-family:monospace;font-size:.72rem;opacity:.7">`+
					`<a href="mailto:%s" style="color:inherit">%s</a></td></tr>`,
				e(d.Name), e(d.Title), e(d.Email), e(d.Email),
			)
		}
		b.WriteString(`</tbody></table>`)
	}
	fmt.Fprintf(&b,
		`<div style="border:1px solid rgba(128,128,128,.2);padding:1rem 1.1rem">`+
			`<p style="font-size:.8rem;line-height:1.6;margin-bottom:.75rem">`+
			`The undersigned attests that the information contained in this report is accurate and complete`+
			` to the best of their knowledge and belief, and that this review was conducted in accordance`+
			` with ACPWB professional standards.</p>`+
			`<div style="display:flex;gap:2rem;flex-wrap:wrap"><div>`+
			`<div style="width:160px;border-bottom:1px solid rgba(128,128,128,.4);margin-bottom:.25rem;`+
			`height:1.5rem"></div>`+
			`<div style="font-size:.72rem;opacity:.65">%s<br>%s<br>`+
			`%s</div></div><div>`+
			`<div style="width:120px;border-bottom:1px solid rgba(128,128,128,.4);margin-bottom:.25rem;`+
			`height:1.5rem"></div>`+
			`<div style="font-size:.72rem;opacity:.65">Date: %s</div></div></div></div></div>`,
		e(c.Assessor), e(c.AssessorTitle), e(c.AssessorEmail), e(c.DateStr),
	)

	fmt.Fprintf(&b,
		`<div style="display:flex;justify-content:space-between;align-items:center;padding-top:1.25rem;`+
			`border-top:1px solid rgba(128,128,128,.15);font-size:.82rem">`+
			`<a href="%s" class="era-nav-link">&larr; Previous</a>`+
			`<a href="%s" class="era-nav-link">Next &rarr;</a></div>`,
		e(c.PrevEntryURL), e(c.NextEntryURL),
	)

	if len(c.CrossYearReports) > 0 {
		b.WriteString(`<div class="mt-4"><div class="era-section-head">Related Records — Other Years</div>` +
			`<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:.5rem">`)
		for _, ent := range c.CrossYearReports {
			fmt.Fprintf(&b,
				`<a href="%s" class="era-entry-card">`+
					`<div style="font-size:.6rem;opacity:.5;margin-bottom:.1rem">%d &bull; `+
					`%s</div>`+
					`<div style="font-size:.73rem;font-weight:600;line-height:1.3">%s</div></a>`,
				e(ent.URL), ent.Year, e(ent.Date), e(ent.Label),
			)
		}
		b.WriteString(`</div></div>`)
	}

	b.WriteString(`<div class="mt-5 pt-3 text-center" style="border-top:1px solid rgba(128,128,128,.12);opacity:.75">`)
	b.WriteString(getArchiveSeal(c.Year, rid))
	b.WriteString(`</div>`)

	b.WriteString(`</div>`) // end col-lg-8

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-sidebar" class="col-lg-4 d-none d-lg-block">`+
			`<div style="position:sticky;top:2rem">`,
		rid,
	)
	fmt.Fprintf(&b,
		`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;`+
			`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">Audit Record</div>`+
			`<dl class="mb-0" style="font-size:.82rem">`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Audit Ref</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Client</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Industry</dt>`+
			`<dd class="mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Findings</dt>`+
			`<dd class="mb-2">%d total</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>`+
			`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd></dl></div>`,
		e(c.AuditRef), e(c.DateStr), e(c.Org), e(c.Industry), len(c.Findings), e(rid),
	)
	fmt.Fprintf(&b,
		`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;`+
			`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">Navigation</div>`+
			`<ul class="list-unstyled mb-0" style="font-size:.82rem">`+
			`<li class="mb-2"><a href="%s" class="era-nav-link">&larr; All %d Records</a></li>`+
			`<li class="mb-2"><a href="%s" style="color:inherit;opacity:.7;`+
			`text-decoration:none">&larr; %d/%02d</a></li>`+
			`<li class="mb-2"><a href="%s" style="color:inherit;opacity:.55;`+
			`text-decoration:none">&larr; Previous Entry</a></li>`+
			`<li class="mb-2"><a href="%s" class="era-nav-link">`+
			`Next in Series &rarr;</a></li></ul></div>`,
		e(c.YearURL), c.Year, e(c.MonthURL), c.Year, c.Month, e(c.PrevEntryURL), e(c.NextEntryURL),
	)
	if len(c.RelatedDocs) > 0 {
		b.WriteString(`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;` +
			`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">` +
			`Related Documents</div>`)
		for _, d := range c.RelatedDocs {
			fmt.Fprintf(&b,
				`<a href="%s" class="era-entry-card" style="margin-bottom:.5rem">`+
					`<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">%s</div>`+
					`<div style="font-size:.75rem;font-weight:600;line-height:1.35">%s</div></a>`,
				e(d.URL), e(d.Date), e(d.Label),
			)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem">` +
		`<div class="era-section-head" style="margin-bottom:.6rem">Browse by Year</div>` +
		`<div style="display:flex;flex-wrap:wrap;gap:.3rem">`)
	for _, y := range c.AllYears {
		extra := ""
		if y == c.Year {
			extra = ";font-weight:800;border-color:var(--era-accent);color:var(--era-accent)"
		}
		fmt.Fprintf(&b,
			`<a href="https://archives-%d.acpwb.com/" `+
				`style="font-size:.68rem;padding:.2rem .4rem;border:1px solid rgba(128,128,128,.3);`+
				`text-decoration:none;color:inherit%s">%d</a>`,
			y, extra, y,
		)
	}
	b.WriteString(`</div></div>`)
	b.WriteString(`</div></div>`)       // sticky, sidebar col
	b.WriteString(`</div></div></div>`) // row, container, era-archive-content
	return b.String()
}
