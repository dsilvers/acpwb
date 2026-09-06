package archive

import (
	"fmt"
	"strings"
)

func minutesSidebarRelatedDocsHTML(c *MinutesContext) string {
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

// RenderMinutesDefault ports pyrender/archive_main.py:render_minutes_default
// (templates/honeypot/archive_minutes.html, main-domain branch).
func RenderMinutesDefault(c *MinutesContext) string {
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
			`text-transform:uppercase">Meeting Minutes &bull; %s</p>`,
		e(c.MeetingRef),
	)
	fmt.Fprintf(&b, `<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">%s</h1>`, e(c.Title))
	fmt.Fprintf(&b,
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">%s &bull; %s</p>`,
		e(c.DateStr), e(c.Location),
	)
	b.WriteString(`</div></section>`)

	b.WriteString(`<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4">`)
	b.WriteString(`<div class="col-lg-8">`)

	quorumBadge := `<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#fee2e2;color:#991b1b;` +
		`border:1px solid #fca5a5">&#10007; QUORUM NOT MET</span>`
	if c.Quorum {
		quorumBadge = `<span style="font-size:.65rem;font-weight:700;padding:.2rem .5rem;background:#dcfce7;color:#15803d;` +
			`border:1px solid #86efac">&#10003; QUORUM ESTABLISHED</span>`
	}
	fmt.Fprintf(&b,
		`<div style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);`+
			`padding:1.25rem 1.5rem;margin-bottom:2rem">`+
			`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;`+
			`color:var(--gold);margin-bottom:.5rem">Meeting Information</div>`+
			`<div style="display:grid;grid-template-columns:1fr 1fr;gap:.3rem .8rem;font-size:.83rem;`+
			`margin-bottom:.75rem">`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Committee</span><br><strong>%s</strong></div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Date</span><br><strong>%s</strong></div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Location</span><br>%s</div>`+
			`<div><span style="font-size:.68rem;opacity:.55;text-transform:uppercase;letter-spacing:.06em">`+
			`Meeting Ref</span><br><code style="font-size:.75rem">%s</code></div></div>`+
			`<div style="display:flex;align-items:center;gap:.5rem;padding-top:.5rem;border-top:1px solid var(--border)">`+
			`%s<span style="font-size:.72rem;color:#666">%d of %d `+
			`members present</span></div></div>`,
		e(c.Committee), e(c.DateStr), e(c.Location), e(c.MeetingRef), quorumBadge, c.NumPresent, c.TotalSeats,
	)

	b.WriteString(`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;` +
		`border-bottom:2px solid var(--gold)">Attendance</h5>` +
		`<table class="dash-table w-100"><thead><tr><th>Name</th><th>Title</th><th>Role</th>` +
		`<th style="text-align:center">Present</th></tr></thead><tbody>`)
	for _, m := range c.Members {
		rowStyle := ""
		if !m.Present {
			rowStyle = ` style="opacity:.5"`
		}
		fw := "400"
		if m.Present {
			fw = "600"
		}
		presentCell := `<span style="color:#ccc">&mdash;</span>`
		if m.Present {
			presentCell = `<span style="color:#15803d;font-weight:700">&#10003;</span>`
		}
		fmt.Fprintf(&b,
			`<tr%s><td style="font-weight:%s">%s</td>`+
				`<td style="font-size:.78rem">%s</td>`+
				`<td style="font-size:.75rem;color:#666">%s</td>`+
				`<td style="text-align:center">%s</td></tr>`,
			rowStyle, fw, e(m.Name), e(m.Title), e(m.Role), presentCell,
		)
	}
	b.WriteString(`</tbody></table></div>`)

	b.WriteString(`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;` +
		`border-bottom:2px solid var(--gold)">Agenda</h5>`)
	for _, item := range c.Items {
		fmt.Fprintf(&b,
			`<div style="margin-bottom:1.5rem;padding-bottom:1.25rem;border-bottom:1px solid var(--border)">`+
				`<div style="font-size:.68rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;`+
				`color:#999;margin-bottom:.2rem">Item %d</div>`+
				`<h6 style="font-size:.92rem;font-weight:700;color:var(--navy);margin-bottom:.6rem">%s</h6>`+
				`<p style="font-size:.87rem;line-height:1.7;color:#333;margin-bottom:0">%s</p>`,
			item.Number, e(item.Title), e(item.Discussion),
		)
		if item.Motion != nil {
			mo := item.Motion
			carried := `<strong style="color:#991b1b">FAILED</strong>`
			if mo.Carried {
				carried = `<strong style="color:#15803d">CARRIED</strong>`
			}
			fmt.Fprintf(&b,
				`<div style="background:#f8f9fb;border:1px solid var(--border);border-left:3px solid var(--navy);`+
					`padding:.8rem 1rem;margin-top:.65rem">`+
					`<div style="font-size:.62rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;`+
					`color:#888;margin-bottom:.4rem">Motion</div>`+
					`<p style="margin-bottom:.4rem;font-size:.83rem;line-height:1.6;color:#333">`+
					`<strong>%s:</strong> %s</p>`+
					`<div style="font-size:.8rem;color:#555;margin-bottom:.3rem">`+
					`<strong>Moved:</strong> %s &nbsp;&bull;&nbsp; `+
					`<strong>Seconded:</strong> %s</div>`+
					`<div style="font-size:.8rem;color:#555">`+
					`<strong>Vote:</strong> %d Yea &nbsp;/&nbsp; %d Nay &nbsp;/&nbsp; `+
					`%d Abstain &nbsp;&mdash;&nbsp; %s</div></div>`,
				e(mo.Verb), e(mo.Text), e(mo.MovedBy), e(mo.SecondedBy), mo.Yea, mo.Nay, mo.Abstain, carried,
			)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<div class="mb-4"><h5 style="font-weight:800;color:var(--navy);text-transform:uppercase;` +
		`letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;` +
		`border-bottom:2px solid var(--gold)">Action Items</h5>` +
		`<table class="dash-table w-100"><thead><tr><th style="width:2rem">#</th><th>Description</th>` +
		`<th>Owner</th><th>Due</th></tr></thead><tbody>`)
	for _, ai := range c.ActionItems {
		fmt.Fprintf(&b,
			`<tr><td style="color:#999;font-weight:600">%d</td>`+
				`<td style="font-size:.83rem">%s</td>`+
				`<td style="font-size:.78rem;font-weight:600;white-space:nowrap">%s</td>`+
				`<td style="font-size:.72rem;color:#666;white-space:nowrap">%s</td></tr>`,
			ai.Number, e(ai.Description), e(ai.Owner), e(ai.DueDate),
		)
	}
	b.WriteString(`</tbody></table></div>`)

	fmt.Fprintf(&b,
		`<div style="border:1px solid var(--border);padding:1rem 1.2rem;margin-bottom:1.5rem">`+
			`<p style="font-size:.85rem;color:#333;line-height:1.7;margin-bottom:.5rem">`+
			`There being no further business, a motion to adjourn was made and carried unanimously.`+
			` Meeting adjourned at %s.</p>`+
			`<p style="font-size:.83rem;margin-bottom:.75rem;color:#555"><strong>Next meeting:</strong> `+
			`%s</p><div><p style="font-size:.78rem;color:#888;margin-bottom:.25rem">`+
			`Respectfully submitted,</p>`+
			`<p style="font-size:.85rem;font-weight:600;color:var(--navy);margin-bottom:.1rem">%s</p>`+
			`<p style="font-size:.78rem;color:#666;margin-bottom:.75rem">%s, `+
			`%s</p>`+
			`<div style="display:flex;gap:2rem">`+
			`<div><div style="width:160px;border-bottom:1px solid #aaa;margin-bottom:.2rem;height:1.5rem"></div>`+
			`<div style="font-size:.68rem;color:#888">Signature</div></div>`+
			`<div><div style="width:120px;border-bottom:1px solid #aaa;margin-bottom:.2rem;height:1.5rem"></div>`+
			`<div style="font-size:.68rem;color:#888">Approved: ___________</div></div></div></div></div>`,
		e(c.AdjournTime), e(c.NextMeeting), e(c.Secretary.Name), e(c.Secretary.Title), e(c.Secretary.Role),
	)

	fmt.Fprintf(&b,
		`<div style="display:flex;justify-content:space-between;padding-top:1rem;border-top:1px solid var(--border)">`+
			`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
			`text-decoration:none">&larr; Previous</a>`+
			`<a href="%s" style="font-size:.85rem;color:var(--navy);font-weight:600;`+
			`text-decoration:none">Next &rarr;</a></div>`,
		e(c.PrevEntryURL), e(c.NextEntryURL),
	)

	b.WriteString(`</div>`) // end col-lg-8

	quorumDl := `<span style="color:#991b1b;font-weight:700">Not Met</span>`
	if c.Quorum {
		quorumDl = `<span style="color:#15803d;font-weight:700">Established</span>`
	}
	b.WriteString(`<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">`)
	fmt.Fprintf(&b,
		`<div class="acpwb-card mb-3"><div style="font-size:.65rem;font-weight:800;text-transform:uppercase;`+
			`letter-spacing:.1em;color:var(--gold);margin-bottom:.7rem">Meeting Record</div>`+
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
	b.WriteString(minutesSidebarRelatedDocsHTML(c))
	b.WriteString(yearBrowserPlainHTML(c.AllYears, c.Year))
	b.WriteString(`</div></div>`)

	b.WriteString(`</div></div></section>`)
	// Note: unlike the default variant, the real archive_minutes.html never
	// renders a bulk_hex <style>/<script> block, even though
	// _generate_minutes_content still computes bulk_hex_js/css.
	return b.String()
}
