package archive

import (
	"fmt"
	"strings"
)

// RenderMinutesDefaultEra ports
// pyrender/archive_era.py:render_minutes_default_era
// (templates/jinja2/honeypot/era/archive_minutes.html).
func RenderMinutesDefaultEra(c *EraMinutesContext) string {
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
			"  .era-motion-block { background: rgba(128,128,128,.06); border: 1px solid rgba(128,128,128,.18); "+
			"border-left: 3px solid %s; padding: .8rem 1rem; margin-top: .65rem; font-size: .82rem; }\n"+
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
		yd.Accent2,
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
			`opacity:.7">Meeting Minutes &bull; %s</p>`,
		e(c.MeetingRef),
	)
	fmt.Fprintf(&b,
		`<h1 style="font-family:var(--era-font-head);font-size:clamp(1.1rem,3vw,1.9rem);line-height:1.25;`+
			`margin-bottom:.3rem">%s</h1>`,
		e(c.Title),
	)
	fmt.Fprintf(&b,
		`<p style="opacity:.75;font-size:.88rem;margin-bottom:0">`+
			`%s &bull; %s &bull; Called to order %s</p>`,
		e(c.DateStr), e(c.Location), e(c.CallToOrder),
	)
	b.WriteString(`</div></div>`)

	b.WriteString(`<div class="era-archive-content"><div class="container"><div class="row g-4"><div class="col-lg-8">`)

	var quorumBadge string
	if c.Quorum {
		quorumBadge = `<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#dcfce7;color:#15803d;` +
			`border:1px solid #86efac">&#10003; QUORUM ESTABLISHED</span>`
	} else {
		quorumBadge = `<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#fee2e2;color:#991b1b;` +
			`border:1px solid #fca5a5">&#10007; QUORUM NOT MET</span>`
	}
	fmt.Fprintf(&b,
		`<div id="acpwb-minutes-%s-header" class="era-callout">`+
			`<div style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--era-accent);margin-bottom:.6rem">Meeting Information</div>`+
			`<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.8rem;`+
			`margin-bottom:.65rem">`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Committee</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Date</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Location</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Meeting Ref</span><br><code style="font-size:.75rem">%s</code></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Called to Order</span><br><strong>%s</strong></div>`+
			`<div><span style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.06em">`+
			`Adjourned</span><br><strong>%s</strong></div></div>`+
			`<div style="display:flex;align-items:center;gap:.5rem">%s`+
			`<span style="font-size:.72rem;opacity:.65">%d of %d `+
			`members present</span></div></div>`,
		rid, e(c.Committee), e(c.DateStr), e(c.Location), e(c.MeetingRef),
		e(c.CallToOrder), e(c.AdjournTime), quorumBadge, c.NumPresent, c.TotalSeats,
	)

	fmt.Fprintf(&b,
		`<div id="acpwb-minutes-%s-attendance" class="mb-4"><div class="era-section-head">Attendance</div>`+
			`<div style="overflow-x:auto"><table style="width:100%%;border-collapse:collapse;font-size:.82rem">`+
			`<thead><tr class="era-table-head" style="color:#fff">`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Name</th>`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Title</th>`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Role</th>`+
			`<th style="padding:.4rem .7rem;text-align:center;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Present</th></tr></thead><tbody>`,
		rid,
	)
	for _, m := range c.Members {
		opacity := ""
		if !m.Present {
			opacity = ";opacity:.5"
		}
		weight := "400"
		if m.Present {
			weight = "600"
		}
		presentCell := `<span style="opacity:.4">&mdash;</span>`
		if m.Present {
			presentCell = `<span style="color:#15803d;font-weight:700">&#10003;</span>`
		}
		fmt.Fprintf(&b,
			`<tr style="border-top:1px solid rgba(128,128,128,.15)%s">`+
				`<td style="padding:.35rem .7rem;font-weight:%s">%s</td>`+
				`<td style="padding:.35rem .7rem;font-size:.78rem;opacity:.8">%s</td>`+
				`<td style="padding:.35rem .7rem;font-size:.75rem;opacity:.7">%s</td>`+
				`<td style="padding:.35rem .7rem;text-align:center">%s</td></tr>`,
			opacity, weight, e(m.Name), e(m.Title), e(m.Role), presentCell,
		)
	}
	b.WriteString(`</tbody></table></div>`)
	quorumNote := "Quorum not established; meeting proceeded in advisory capacity only."
	if c.Quorum {
		quorumNote = "Quorum established."
	}
	fmt.Fprintf(&b,
		`<p style="font-size:.72rem;opacity:.6;margin-top:.5rem;margin-bottom:0">`+
			`%d of %d members present. %s</p></div>`,
		c.NumPresent, c.TotalSeats, quorumNote,
	)

	fmt.Fprintf(&b, `<div id="acpwb-minutes-%s-agenda" class="mb-4"><div class="era-section-head">Agenda</div>`, rid)
	for _, item := range c.Items {
		fmt.Fprintf(&b,
			`<div id="acpwb-minutes-%s-item-%d" style="margin-bottom:1.5rem;`+
				`padding-bottom:1.25rem;border-bottom:1px solid rgba(128,128,128,.12)">`+
				`<div style="font-size:.72rem;font-weight:800;text-transform:uppercase;letter-spacing:.1em;`+
				`opacity:.5;margin-bottom:.2rem">Item %d</div>`+
				`<h6 style="font-size:.92rem;font-weight:700;margin-bottom:.6rem;line-height:1.3">`+
				`%s</h6>`+
				`<p style="font-size:.85rem;line-height:1.7;margin-bottom:0">%s</p>`,
			rid, item.Number, item.Number, e(item.Title), e(item.Discussion),
		)
		if item.Motion != nil {
			mo := item.Motion
			carried := `<span style="font-weight:800;color:#991b1b">FAILED</span>`
			if mo.Carried {
				carried = `<span style="font-weight:800;color:#15803d">CARRIED</span>`
			}
			fmt.Fprintf(&b,
				`<div class="era-motion-block">`+
					`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.1em;`+
					`opacity:.55;margin-bottom:.4rem">Motion</div>`+
					`<p style="margin-bottom:.5rem;font-size:.83rem;line-height:1.6">`+
					`<strong>%s:</strong> %s</p>`+
					`<div style="font-size:.8rem;margin-bottom:.35rem"><strong>Moved:</strong> %s `+
					`&nbsp;&bull;&nbsp; <strong>Seconded:</strong> %s</div>`+
					`<div style="font-size:.8rem"><strong>Vote:</strong> %d Yea &nbsp;/&nbsp; `+
					`%d Nay &nbsp;/&nbsp; %d Abstain &nbsp;&mdash;&nbsp; %s</div></div>`,
				e(mo.Verb), e(mo.Text), e(mo.MovedBy), e(mo.SecondedBy), mo.Yea, mo.Nay, mo.Abstain, carried,
			)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b,
		`<div id="acpwb-minutes-%s-action-items" class="mb-4"><div class="era-section-head">`+
			`Action Items</div><div style="overflow-x:auto">`+
			`<table style="width:100%%;border-collapse:collapse;font-size:.82rem"><thead>`+
			`<tr class="era-table-head" style="color:#fff">`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase;width:2.5rem">#</th>`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase">Description</th>`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase;white-space:nowrap">Owner</th>`+
			`<th style="padding:.4rem .7rem;text-align:left;font-size:.68rem;letter-spacing:.06em;`+
			`text-transform:uppercase;white-space:nowrap">Due</th></tr></thead><tbody>`,
		rid,
	)
	for _, ai := range c.ActionItems {
		fmt.Fprintf(&b,
			`<tr style="border-top:1px solid rgba(128,128,128,.15)">`+
				`<td style="padding:.35rem .7rem;opacity:.5;font-weight:600">%d</td>`+
				`<td style="padding:.35rem .7rem;font-size:.82rem;line-height:1.5">%s</td>`+
				`<td style="padding:.35rem .7rem;font-size:.78rem;white-space:nowrap;font-weight:600">`+
				`%s</td>`+
				`<td style="padding:.35rem .7rem;font-size:.72rem;white-space:nowrap;opacity:.7">`+
				`%s</td></tr>`,
			ai.Number, e(ai.Description), e(ai.Owner), e(ai.DueDate),
		)
	}
	b.WriteString(`</tbody></table></div></div>`)

	fmt.Fprintf(&b,
		`<div id="acpwb-minutes-%s-adjournment" class="mb-4">`+
			`<div class="era-section-head">Adjournment</div>`+
			`<p style="font-size:.85rem;line-height:1.7;margin-bottom:.5rem">`+
			`There being no further business, a motion to adjourn was made and carried unanimously.`+
			` The meeting was adjourned at %s.</p>`+
			`<p style="font-size:.82rem;margin-bottom:.5rem"><strong>Next meeting:</strong> `+
			`%s</p>`+
			`<div style="margin-top:1rem;padding-top:.85rem;border-top:1px solid rgba(128,128,128,.15)">`+
			`<p style="font-size:.8rem;margin-bottom:.35rem;opacity:.7">Respectfully submitted,</p>`+
			`<p style="font-size:.83rem;font-weight:600;margin-bottom:.1rem">%s</p>`+
			`<p style="font-size:.78rem;opacity:.65;margin-bottom:.6rem">%s, `+
			`%s</p>`+
			`<div style="display:flex;gap:2rem;flex-wrap:wrap"><div>`+
			`<div style="width:160px;border-bottom:1px solid rgba(128,128,128,.4);margin-bottom:.2rem;`+
			`height:1.5rem"></div><div style="font-size:.68rem;opacity:.55">Signature</div></div><div>`+
			`<div style="width:120px;border-bottom:1px solid rgba(128,128,128,.4);margin-bottom:.2rem;`+
			`height:1.5rem"></div>`+
			`<div style="font-size:.68rem;opacity:.55">Approved: ___________</div></div></div></div></div>`,
		rid, e(c.AdjournTime), e(c.NextMeeting), e(c.Secretary.Name), e(c.Secretary.Title), e(c.Secretary.Role),
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
		`<div id="acpwb-minutes-%s-sidebar" class="col-lg-4 d-none d-lg-block">`+
			`<div style="position:sticky;top:2rem">`,
		rid,
	)
	quorumDl := `<span style="color:#991b1b;font-weight:700">Not Met</span>`
	if c.Quorum {
		quorumDl = `<span style="color:#15803d;font-weight:700">Established</span>`
	}
	fmt.Fprintf(&b,
		`<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;`+
			`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.6rem">Meeting Record</div>`+
			`<dl class="mb-0" style="font-size:.82rem">`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Meeting Ref</dt>`+
			`<dd class="mb-2" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Date</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Committee</dt>`+
			`<dd class="fw-700 mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Quorum</dt>`+
			`<dd class="mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Agenda Items</dt>`+
			`<dd class="mb-2">%d</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Record ID</dt>`+
			`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd></dl></div>`,
		e(c.MeetingRef), e(c.DateStr), e(c.Committee), quorumDl, len(c.Items), e(rid),
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
