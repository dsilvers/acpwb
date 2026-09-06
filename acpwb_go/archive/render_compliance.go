package archive

import (
	"fmt"
	"strings"
)

// riskBadge ports pyrender/archive_main.py:_risk_badge. The real template
// uses a DIFFERENT fallback label in two places: "INFO" in the findings
// summary table, "INFORMATIONAL" in the detailed findings cards — same
// colors, different text.
func riskBadge(risk, fallbackLabel string) string {
	var bg, fg, border, label string
	switch risk {
	case "HIGH":
		bg, fg, border, label = "#fee2e2", "#991b1b", "#fca5a5", "HIGH"
	case "MEDIUM":
		bg, fg, border, label = "#fef3c7", "#92400e", "#fcd34d", "MEDIUM"
	case "LOW":
		bg, fg, border, label = "#fef9c3", "#854d0e", "#fde047", "LOW"
	default:
		bg, fg, border, label = "#f1f5f9", "#475569", "#cbd5e1", fallbackLabel
	}
	return fmt.Sprintf(
		`<span style="font-size:.62rem;font-weight:700;padding:.15rem .4rem;background:%s;`+
			`color:%s;border:1px solid %s">%s</span>`,
		bg, fg, border, label,
	)
}

// statusBadge ports pyrender/archive_main.py:_status_badge.
func statusBadge(status string) string {
	var bg, fg, border, label string
	switch status {
	case "OPEN":
		bg, fg, border, label = "#fee2e2", "#991b1b", "#fca5a5", "OPEN"
	case "IN PROGRESS":
		bg, fg, border, label = "#fef3c7", "#92400e", "#fcd34d", "IN PROGRESS"
	case "REMEDIATED":
		bg, fg, border, label = "#dcfce7", "#15803d", "#86efac", "REMEDIATED"
	case "DEFERRED":
		bg, fg, border, label = "#f1f5f9", "#475569", "#cbd5e1", "DEFERRED"
	default:
		bg, fg, border, label = "#eff6ff", "#1d4ed8", "#93c5fd", "MONITORING"
	}
	return fmt.Sprintf(
		`<span style="font-size:.62rem;font-weight:700;padding:.15rem .4rem;background:%s;`+
			`color:%s;border:1px solid %s">%s</span>`,
		bg, fg, border, label,
	)
}

// yearBrowserPlainHTML ports pyrender/archive_main.py:_year_browser_plain_html
// — the compliance/minutes "Browse by Year" sidebar: no id/aria-label per
// link (unlike the default variant's), current year bolded.
func yearBrowserPlainHTML(allYears []int, currentYear int) string {
	var items strings.Builder
	for _, y := range allYears {
		extra := ""
		if y == currentYear {
			extra = ";font-weight:800;border-color:var(--gold);color:var(--navy)"
		}
		fmt.Fprintf(&items,
			`<a href="https://archives-%d.acpwb.com/" style="font-size:.68rem;padding:.2rem .4rem;`+
				`border:1px solid var(--border);text-decoration:none;color:inherit%s">%d</a>`,
			y, extra, y,
		)
	}
	return `<div class="acpwb-card"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;` +
		`letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Browse by Year</div>` +
		`<div style="display:flex;flex-wrap:wrap;gap:.3rem">` + items.String() + `</div></div>`
}

func complianceSidebarRelatedDocsHTML(c *ComplianceContext) string {
	if len(c.RelatedDocs) == 0 {
		return ""
	}
	var items strings.Builder
	for _, d := range c.RelatedDocs {
		fmt.Fprintf(&items,
			`<a href="%s" style="display:block;padding:.5rem .7rem;border:1px solid var(--border);`+
				`border-left:3px solid var(--gold);text-decoration:none;color:inherit;margin-bottom:.4rem">`+
				`<div style="font-size:.6rem;opacity:.5;margin-bottom:.1rem">%s</div>`+
				`<div style="font-size:.75rem;font-weight:600;color:var(--navy)">%s</div></a>`,
			e(d.URL), e(d.Date), e(d.Label),
		)
	}
	return `<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;` +
		`letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Related Documents</div>` + items.String() + `</div>`
}

// RenderComplianceDefault ports pyrender/archive_main.py:render_compliance_default
// (templates/honeypot/archive_compliance.html, main-domain branch).
func RenderComplianceDefault(c *ComplianceContext) string {
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
	fmt.Fprintf(&b,
		`<p style="color:rgba(255,255,255,.6);font-size:.72rem;margin-bottom:.4rem;letter-spacing:.1em;`+
			`text-transform:uppercase">Compliance Review &bull; %s</p>`,
		e(c.AuditRef),
	)
	fmt.Fprintf(&b, `<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">%s</h1>`, e(c.Title))
	fmt.Fprintf(&b,
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">`+
			`%s &bull; %s &bull; Archived %s</p>`,
		e(c.Industry), e(c.Org), e(c.DateStr),
	)
	b.WriteString(`</div></section>`)

	b.WriteString(`<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">`)
	b.WriteString(`<div class="col-lg-8">`)

	fmt.Fprintf(&b,
		`<div id="acpwb-compliance-%s-header" style="background:white;border:1px solid var(--border);`+
			`border-left:4px solid var(--gold);padding:1.25rem 1.5rem;margin-bottom:2rem">`+
			`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--gold);margin-bottom:.5rem">Document Information</div>`+
			`<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.83rem;`+
			`margin-bottom:.75rem">`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Client</span><br><strong>%s</strong></div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Industry</span><br><strong>%s</strong></div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Audit Ref</span><br><code style="font-size:.75rem">%s</code></div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Version</span><br><code style="font-size:.75rem">%s</code></div></div>`+
			`<div style="font-size:.8rem;color:#555;padding-top:.5rem;border-top:1px solid var(--border)">`+
			`Filed by <strong>%s</strong>, %s &mdash; `+
			`<a href="mailto:%s" style="color:var(--navy)">%s</a></div>`,
		rid, e(c.Org), e(c.Industry), e(c.AuditRef), e(c.DocVersion),
		e(c.Assessor), e(c.AssessorTitle), e(c.AssessorEmail), e(c.AssessorEmail),
	)
	if len(c.FrameworksCited) > 0 {
		b.WriteString(`<div style="margin-top:.6rem;display:flex;flex-wrap:wrap;gap:.3rem">`)
		for _, fw := range c.FrameworksCited {
			fmt.Fprintf(&b,
				`<span style="font-size:.62rem;padding:.15rem .45rem;background:#f4f6f9;border:1px solid #dde1e8;`+
					`color:#444">%s</span>`,
				e(fw),
			)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b,
		`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;`+
			`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;`+
			`border-bottom:2px solid var(--gold)">1. Engagement Scope</h5>`+
			`<p style="font-size:.88rem;line-height:1.7;color:#333">%s</p></div>`,
		e(c.ScopePara),
	)

	fmt.Fprintf(&b,
		`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;`+
			`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;`+
			`border-bottom:2px solid var(--gold)">2. Methodology</h5>`+
			`<p style="font-size:.88rem;line-height:1.7;color:#333">%s</p></div>`,
		e(c.MethodPara),
	)

	b.WriteString(`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;` +
		`border-bottom:2px solid var(--gold)">3. Findings Summary</h5>` +
		`<div style="overflow-x:auto"><table class="dash-table w-100"><thead><tr>` +
		`<th>Finding ID</th><th>Risk</th><th>Status</th><th>Owner</th><th>Due</th><th>Description</th>` +
		`</tr></thead><tbody>`)
	for _, f := range c.Findings {
		fmt.Fprintf(&b,
			`<tr><td style="font-family:monospace;font-size:.72rem;white-space:nowrap">%s</td>`+
				`<td style="white-space:nowrap">%s</td>`+
				`<td style="white-space:nowrap">%s</td>`+
				`<td style="font-size:.75rem">%s</td>`+
				`<td style="font-size:.72rem;white-space:nowrap">%s</td>`+
				`<td style="font-size:.78rem;max-width:240px">%s</td></tr>`,
			e(f.ID), riskBadge(f.Risk, "INFO"), statusBadge(f.Status), e(f.Owner), e(f.DueDate), e(f.Description),
		)
	}
	b.WriteString(`</tbody></table></div></div>`)

	b.WriteString(`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;` +
		`border-bottom:2px solid var(--gold)">4. Detailed Findings</h5>`)
	for _, f := range c.Findings {
		statusExtra := ""
		if f.Status == "REMEDIATED" {
			statusExtra = statusBadge("REMEDIATED")
		}
		fmt.Fprintf(&b,
			`<div style="background:white;border:1px solid var(--border);margin-bottom:1rem;padding:1rem 1.2rem">`+
				`<div style="display:flex;align-items:center;gap:.5rem;margin-bottom:.55rem">`+
				`<code style="font-size:.72rem;color:#666">%s</code>%s%s</div>`+
				`<div style="margin-bottom:.5rem">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`color:#888;margin-bottom:.2rem">Observation</div>`+
				`<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:0">%s</p></div>`+
				`<div style="margin-bottom:.5rem;padding-top:.5rem;border-top:1px solid var(--border)">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`color:#888;margin-bottom:.2rem">Corrective Action Required</div>`+
				`<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:0">%s</p></div>`+
				`<div style="padding-top:.5rem;border-top:1px solid var(--border)">`+
				`<div style="font-size:.65rem;font-weight:700;text-transform:uppercase;letter-spacing:.07em;`+
				`color:#888;margin-bottom:.2rem">Management Response</div>`+
				`<p style="font-size:.85rem;line-height:1.65;color:#333;margin-bottom:.4rem">%s</p>`+
				`<div style="font-size:.72rem;color:#666">Owner: <strong>%s</strong> &bull; `+
				`Due: <strong>%s</strong></div></div></div>`,
			e(f.ID), riskBadge(f.Risk, "INFORMATIONAL"), statusExtra,
			e(f.Description), e(f.CorrectiveAction), e(f.MgmtResponse), e(f.Owner), e(f.DueDate),
		)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b,
		`<div style="margin-top:1.5rem;padding-top:1.25rem;border-top:1px solid var(--border);`+
			`display:flex;justify-content:space-between">`+
			`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
			`text-decoration:none">&larr; Previous</a>`+
			`<a href="%s" style="font-size:.85rem;color:var(--navy);font-weight:600;`+
			`text-decoration:none">Next &rarr;</a></div>`,
		e(c.PrevEntryURL), e(c.NextEntryURL),
	)

	b.WriteString(`</div>`) // end col-lg-8

	b.WriteString(`<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">`)
	fmt.Fprintf(&b,
		`<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;`+
			`letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Audit Record</div>`+
			`<dl class="mb-0" style="font-size:.82rem">`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Audit Ref</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Organization</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Findings</dt>`+
			`<dd class="mb-2">%d total</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>`+
			`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd></dl></div>`,
		e(c.AuditRef), e(c.DateStr), e(c.Org), len(c.Findings), e(rid),
	)
	fmt.Fprintf(&b,
		`<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;`+
			`letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Navigation</div>`+
			`<ul class="list-unstyled mb-0" style="font-size:.82rem">`+
			`<li class="mb-2"><a href="%s" style="color:var(--navy);font-weight:600;`+
			`text-decoration:none">&larr; All %d Records</a></li>`+
			`<li class="mb-2"><a href="%s" style="color:inherit;opacity:.7;`+
			`text-decoration:none">&larr; %d/%02d</a></li>`+
			`<li class="mb-2"><a href="%s" style="color:inherit;opacity:.55;`+
			`text-decoration:none">&larr; Previous Entry</a></li>`+
			`<li class="mb-2"><a href="%s" style="color:var(--navy);font-weight:600;`+
			`text-decoration:none">Next in Series &rarr;</a></li></ul></div>`,
		e(c.YearURL), c.Year, e(c.MonthURL), c.Year, c.Month, e(c.PrevEntryURL), e(c.NextEntryURL),
	)
	b.WriteString(complianceSidebarRelatedDocsHTML(c))
	b.WriteString(yearBrowserPlainHTML(c.AllYears, c.Year))
	b.WriteString(`</div></div>`)

	b.WriteString(`</div></div></section>`)
	// Note: unlike the default variant, the real archive_compliance.html
	// never renders a bulk_hex <style>/<script> block, even though
	// _generate_compliance_content still computes bulk_hex_js/css.
	return b.String()
}
