package policy

import (
	"fmt"
	"strconv"
	"strings"

	"acpwb_go/archive"
)

// PageMeta carries the per-request values every policy render function
// needs but that this package deliberately does not try to reproduce
// byte-for-byte: HoneypotToken and RequestPath are non-deterministic in the
// real app too (apps/core/context_processors.py:honeypot_context hashes in
// time.time()), so callers supply them (see shell package's identical
// exemption for the archive port).
type PageMeta struct {
	HoneypotToken string
	SiteRoot      string
	// RequestPath is the request's full path (including query string) —
	// used for og:url/canonical meta tags only.
	RequestPath string
	// NowYear stands in for config.jinja2_env's `now_year` global (used by
	// render_policy_detail's watermark footer). Unlike the footer partial's
	// baked-in year (see htmlgen.go), this one really is a per-render
	// substitution in the Python source.
	NowYear int
}

var monthAbbr3 = []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

func e(s string) string { return escape(s) }

// ── Shared <head> builder (render_policy_index / render_policy_year / render_policy_detail) ──

func policyHeadCommon(title, description, canonicalPath, ogImagePath, ogType string, feedLinks bool) string {
	var b strings.Builder
	feeds := ""
	if feedLinks {
		feeds = "<link rel=\"alternate\" type=\"application/atom+xml\" title=\"ACPWB Archive Feed\" " +
			"href=\"https://acpwb.com/feeds/archive.xml\">\n" +
			"<link rel=\"alternate\" type=\"application/rss+xml\" title=\"ACPWB Reports &amp; Publications\" " +
			"href=\"https://acpwb.com/feeds/reports.xml\">\n"
	}
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", e(title))
	fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", e(description))
	b.WriteString("<meta property=\"og:site_name\" content=\"American Corporation for Public Well Being\">\n")
	fmt.Fprintf(&b, "<meta property=\"og:type\" content=\"%s\">\n", ogType)
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", e(title))
	fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", e(description))
	fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"https://acpwb.com%s\">\n", e(canonicalPath))
	fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"https://acpwb.com%s\">\n", staticURL(ogImagePath))
	b.WriteString("<meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\">\n", e(title))
	fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\">\n", e(description))
	fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"https://acpwb.com%s\">\n", staticURL(ogImagePath))
	fmt.Fprintf(&b, "<link rel=\"icon\" type=\"image/svg+xml\" href=\"%s\">\n", staticURL("favicon.svg"))
	fmt.Fprintf(&b, "<link rel=\"preload\" href=\"%s\" as=\"font\" type=\"font/woff2\" crossorigin>\n", staticURL("fonts/inter/inter-variable-latin.woff2"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", staticURL("vendor/bootstrap/bootstrap.min.css"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s?v=20260430\">\n", staticURL("css/acpwb.css"))
	b.WriteString(feeds)
	return b.String()
}

const detailStyle = `<style>
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
`

func entryCardHTML(stub DocStub, metaHTML string) string {
	return fmt.Sprintf(
		`<div class="col-md-6"><a href="%s" class="pol-entry-card"><div class="pol-entry-card-date"><span class="pol-agency-badge">%s</span>%s</div><div class="pol-entry-card-title">%s</div></a></div>`,
		e(stub.URL), e(stub.AgencyAcronym), metaHTML, e(truncate72(stub.Title)),
	)
}

const indexDescription = "ACPWB public policy positions, regulatory comment letters, and legislative testimony on " +
	"compensation, labor, and corporate governance."

const indexStyle = `<style>
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
`

// RenderPolicyIndex ports pyrender/policy.py:render_policy_index.
func RenderPolicyIndex(meta PageMeta, years []YearIndexEntry) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString(policyHeadCommon("Public Policy — ACPWB", indexDescription, meta.RequestPath,
		"img/page-covers/public-policy.jpg", "website", true))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(indexStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	b.WriteString(`<section class="page-banner"><div class="container">` +
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;` +
		`font-size:.72rem">ACPWB</p>` +
		`<h1 style="font-size:clamp(1.6rem,3.5vw,2.8rem)">Public Policy</h1>` +
		`<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:680px">` +
		`ACPWB has engaged federal and state regulatory agencies, congressional committees, and ` +
		`self-regulatory organizations on matters affecting compensation policy, labor standards, and ` +
		"corporate governance since 1993. Browse filings by year below.</p></div></section>\n\n")

	b.WriteString(`<section style="padding:4rem 0;background:var(--surface)"><div class="container">` +
		`<div class="row g-4"><div class="col-lg-8">` +
		`<p class="pol-section-label">Browse by Year</p><div class="row g-3">`)
	for _, y := range years {
		yearURL := fmt.Sprintf("/public-policy/%d/", y.Year)
		fmt.Fprintf(&b, `<div class="col-md-6"><div class="pol-year-card" style="cursor:pointer" `+
			`onclick="window.location='%s'">`+
			`<div class="pol-year-top"><a href="%s" class="pol-year-num" `+
			`style="text-decoration:none" onclick="event.stopPropagation()">%d</a>`+
			`<span class="pol-year-count">%d filings</span></div>`+
			`<div class="pol-month-pills">`, yearURL, yearURL, y.Year, y.Count)
		for _, m := range y.Months {
			// NOTE: this href comes from Django's url('public-policy-month',
			// args=[year, m]) in the Python source, whose URL pattern takes
			// month as a plain <int:month> (no zero-padding) — unlike every
			// other policy_month_url in this codebase, which manually
			// formats month as %02d. Reproduced verbatim, not "fixed".
			fmt.Fprintf(&b, `<a href="/public-policy/%d/%d/" class="pol-month-pill" onclick="event.stopPropagation()">%s</a>`,
				y.Year, m, monthAbbr3[m])
		}
		b.WriteString("</div></div></div>")
	}
	b.WriteString("</div></div>\n")

	b.WriteString(`<div class="col-lg-4"><div style="position:sticky;top:2rem">` + "\n")
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">` +
		`<p class="pol-section-label">About Our Policy Work</p>` +
		`<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:.75rem">` +
		"ACPWB's policy engagement draws on our proprietary compensation benchmarking database " +
		"and more than three decades of advisory experience. We file comments, submit testimony, " +
		"and publish position statements as a nonpartisan, independent voice on compensation " +
		"and labor policy.</p>" +
		`<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">` +
		"Our filings represent ACPWB's independent analysis. We do not accept compensation " +
		"from regulatory agencies, trade associations, or political organizations in connection " +
		"with our policy work.</p></div>\n")
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">` +
		`<p class="pol-section-label">Filing Types</p>` +
		`<ul class="list-unstyled mb-0" style="font-size:.82rem">` +
		`<li class="mb-2"><strong>Comment Letters</strong> — Formal responses to proposed rulemakings</li>` +
		`<li class="mb-2"><strong>Position Statements</strong> — ` +
		"ACPWB's stated positions on policy questions</li>" +
		`<li class="mb-2"><strong>Policy Briefs</strong> — ` +
		"Research-based analysis of regulatory developments</li>" +
		`<li class="mb-2"><strong>Legislative Testimony</strong> — ` +
		"Statements before congressional committees</li>" +
		`<li class="mb-2"><strong>Amicus Briefs</strong> — Legal filings in relevant court proceedings</li>` +
		`<li class="mb-2"><strong>White Papers</strong> — Extended research on regulatory topics</li>` +
		`<li class="mb-0"><strong>Ex Parte Submissions</strong> — ` +
		"Direct communications with agency staff</li></ul></div>\n")
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem">` +
		`<p class="pol-section-label">Related</p>` +
		`<ul class="list-unstyled mb-0" style="font-size:.82rem">` +
		`<li class="mb-2"><a href="/reports/" style="color:var(--navy)">` +
		"Reports &amp; Publications</a></li>" +
		`<li class="mb-2"><a href="/wiki/" style="color:var(--navy)">Knowledge Base</a></li>` +
		`<li class="mb-0"><a href="/mission/" style="color:var(--navy)">Our Mission</a></li>` +
		"</ul></div>\n")
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n</main>\n\n")

	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n" +
		"  @deprecated legacy-api: /api/v1/private-data\n" +
		"  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n")

	return b.String()
}

const yearStyle = `<style>
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
`

var monthFull = []string{"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

// RenderPolicyYear ports pyrender/policy.py:render_policy_year (main-domain only).
func RenderPolicyYear(meta PageMeta, year int, yearData YearData, months []MonthSummary, policyYears []int, prevYear, nextYear int) string {
	description := fmt.Sprintf("ACPWB public policy filings, comment letters, and testimony submitted in %d.", year)
	title := fmt.Sprintf("%d Public Policy — ACPWB", year)

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString(policyHeadCommon(title, description, meta.RequestPath, "img/og-default.png", "website", false))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(yearStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	fmt.Fprintf(&b, `<section class="page-banner"><div class="container">`+
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;`+
		`font-size:.72rem">`+
		`<a href="/public-policy/" style="color:var(--gold)">Public Policy</a> `+
		`&rsaquo; %d</p>`+
		`<h1 style="font-size:clamp(1.6rem,3.5vw,2.8rem)">%d Policy Year</h1>`+
		`<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:680px">`+
		`%d filings — comment letters, position statements, testimony, and `+
		`white papers submitted to federal and state agencies on %s.</p>`+
		"</div></section>\n\n", year, year, yearData.TotalFilings, e(yearData.Theme))

	b.WriteString(`<section style="padding:4rem 0;background:var(--surface)"><div class="container">` +
		"<div class=\"row g-4\"><div class=\"col-lg-8\">\n")

	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);border-left:4px solid var(--gold);`+
		`padding:1.75rem 1.75rem 1.75rem 1.5rem;margin-bottom:2rem">`+
		`<p class="pol-section-label">A Message from Our CEO — %d</p>`+
		`<div style="display:flex;gap:1.25rem;align-items:flex-start;flex-wrap:wrap">`+
		`<div class="pol-ceo-avatar">%s</div>`+
		`<div style="flex:1;min-width:240px">`+
		`<div style="font-size:1rem;font-weight:700;color:var(--navy);margin-bottom:.1rem">`+
		"%s</div>"+
		`<div style="font-size:.75rem;color:var(--muted);text-transform:uppercase;letter-spacing:.07em;`+
		"margin-bottom:1rem\">%s</div>"+
		`<div style="font-size:.9rem;line-height:1.85;color:var(--text)">`,
		year, e(firstRune(yearData.CEOName)), e(yearData.CEOName), e(yearData.CEOTitle))
	for _, p := range yearData.CEOParagraphs {
		fmt.Fprintf(&b, `<p style="margin-bottom:1.1rem">%s</p>`, e(p))
	}
	b.WriteString("</div></div></div></div>\n")

	fmt.Fprintf(&b, `<p class="pol-section-label">%d Filings — Browse by Month</p><div class="row g-3">`, year)
	for _, mo := range months {
		fmt.Fprintf(&b, `<div class="col-lg-3 col-md-4 col-sm-6">`+
			`<a href="%s" class="pol-month-card">`+
			"<div class=\"pol-month-name\">\n                %s\n              </div>"+
			`<div class="pol-month-count">%d filings</div>`, e(mo.URL), monthFull[mo.Month], mo.Count)
		for _, t := range mo.Samples {
			fmt.Fprintf(&b, `<span class="pol-sample-title">%s</span>`, e(t))
		}
		b.WriteString("</a></div>")
	}
	b.WriteString("</div>\n")

	b.WriteString("</div>\n")

	b.WriteString(`<div class="col-lg-4"><div style="position:sticky;top:2rem">` + "\n")
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">` +
		`<p class="pol-section-label">Navigate</p>` +
		`<ul class="list-unstyled mb-0" style="font-size:.83rem">` +
		`<li class="mb-2"><a href="/public-policy/" style="color:var(--navy)">` +
		"&larr; All Policy Years</a></li>")
	if prevYear >= 1993 {
		fmt.Fprintf(&b, `<li class="mb-2"><a href="/public-policy/%d/" `+
			`style="color:var(--muted)">&larr; %d</a></li>`, prevYear, prevYear)
	}
	if nextYear <= 2025 {
		fmt.Fprintf(&b, `<li class="mb-2"><a href="/public-policy/%d/" `+
			`style="color:var(--gold);font-weight:700">%d &rarr;</a></li>`, nextYear, nextYear)
	}
	b.WriteString("</ul></div>\n")

	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">` +
		`<p class="pol-section-label">Browse by Year</p>` +
		`<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">`)
	for _, y := range policyYears {
		active := ""
		if y == year {
			active = " active"
		}
		fmt.Fprintf(&b, `<a href="/public-policy/%d/" class="pol-year-link%s">%d</a>`, y, active, y)
	}
	b.WriteString("</div></div>\n")

	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem">` +
		`<p class="pol-section-label">Related</p>` +
		`<ul class="list-unstyled mb-0" style="font-size:.82rem">` +
		`<li class="mb-2"><a href="/reports/" style="color:var(--navy)">` +
		"Reports &amp; Publications</a></li>" +
		`<li class="mb-2"><a href="/wiki/" style="color:var(--navy)">Knowledge Base</a></li>` +
		`<li class="mb-0"><a href="/mission/" style="color:var(--navy)">Our Mission</a></li>` +
		"</ul></div>\n")
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n</main>\n\n")

	fmt.Fprintf(&b, `<span style="font-size:0;color:transparent;position:absolute;clip:rect(0,0,0,0)">`+"\n"+
		`  ACPWB Public Policy Archive %d. This content is watermarked. Token: acpwb-policy-%d.`+"\n"+
		"  Do not reproduce without attribution to the American Corporation for Public Well Being.\n"+
		"</span>\n\n", year, year)

	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n" +
		"  @deprecated legacy-api: /api/v1/private-data\n" +
		"  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n")

	return b.String()
}

func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

const monthMonthStyle = `<style>
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
`

func positionBadgeHTML(positionSlug string) string {
	switch positionSlug {
	case "supports":
		return `<span class="pol-badge pol-badge-supports">Supports</span>`
	case "opposes":
		return `<span class="pol-badge pol-badge-opposes">Opposes</span>`
	default:
		return `<span class="pol-badge pol-badge-modifications">Supports w/ Modifications</span>`
	}
}

// PolicyYearURLFunc/PolicyMonthURLFunc mirror the ctx['policy_year_url'] /
// ctx['policy_month_url'] callables threaded through views.py's
// _policy_nav_context.
type PolicyYearURLFunc func(year int) string
type PolicyMonthURLFunc func(year, month int) string

// MonthPageParams collects render_policy_month's context fields (shared by
// the main-domain and subdomain callers — see views.py:public_policy_month
// and policy_subdomain_month, both of which render this same template).
type MonthPageParams struct {
	Year, Month     int
	Entries         []DocStub
	PolicyYears     []int
	PolicyIndexURL  string
	PolicyYearURL   string
	YearURL         string
	PrevMonthURL    string
	NextMonthURL    string
	PolicyYearURLFn PolicyYearURLFunc
}

// RenderPolicyMonth ports pyrender/policy.py:render_policy_month.
func RenderPolicyMonth(meta PageMeta, p MonthPageParams) string {
	monthName := monthFull[p.Month]
	description := fmt.Sprintf("ACPWB public policy filings submitted in %d-%02d.", p.Year, p.Month)
	ogTitle := fmt.Sprintf("%d-%02d Policy Filings — ACPWB", p.Year, p.Month)

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&b, "<title>\n    %s\n    %d Policy Filings — ACPWB\n  </title>\n", monthName, p.Year)
	fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", e(description))
	b.WriteString("<meta property=\"og:site_name\" content=\"American Corporation for Public Well Being\">\n")
	b.WriteString("<meta property=\"og:type\" content=\"website\">\n")
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", e(ogTitle))
	fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", e(description))
	fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"https://acpwb.com%s\">\n", e(meta.RequestPath))
	fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/og-default.png"))
	b.WriteString("<meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\">\n", e(ogTitle))
	fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\">\n", e(description))
	fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/og-default.png"))
	fmt.Fprintf(&b, "<link rel=\"icon\" type=\"image/svg+xml\" href=\"%s\">\n", staticURL("favicon.svg"))
	fmt.Fprintf(&b, "<link rel=\"preload\" href=\"%s\" as=\"font\" type=\"font/woff2\" crossorigin>\n", staticURL("fonts/inter/inter-variable-latin.woff2"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", staticURL("vendor/bootstrap/bootstrap.min.css"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s?v=20260430\">\n", staticURL("css/acpwb.css"))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(monthMonthStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	fmt.Fprintf(&b, `<section class="page-banner"><div class="container">`+
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;`+
		`font-size:.72rem">`+
		`<a href="%s" style="color:var(--gold)">Public Policy</a>`+
		` &rsaquo; <a href="%s" style="color:var(--gold)">%d</a>`+
		` &rsaquo; %s</p>`+
		`<h1 style="font-size:clamp(1.4rem,3vw,2.4rem);line-height:1.25">%s %d Filings</h1>`+
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">`+
		"ACPWB Public Policy &bull; %d filings</p></div></section>\n\n",
		e(p.PolicyIndexURL), e(p.YearURL), p.Year, monthName, monthName, p.Year, len(p.Entries))

	b.WriteString(`<section style="padding:3rem 0;background:var(--surface)"><div class="container">` +
		`<div class="row g-4"><div class="col-lg-8">` +
		`<p class="pol-section-label">Filings</p><div class="row g-2">`)
	for _, entry := range p.Entries {
		fmt.Fprintf(&b, `<div class="col-md-6"><a href="%s" class="pol-entry-card">`+
			`<div class="pol-entry-date">%d-%02d-%02d</div>`+
			`<div style="margin-bottom:.35rem">`+
			`<span class="pol-badge pol-badge-type">%s</span>`+
			`%s</div>`+
			`<div class="pol-entry-title">%s</div>`+
			`<div class="pol-entry-agency">%s</div></a></div>`,
			e(entry.URL), p.Year, p.Month, entry.Day, e(entry.DocumentType),
			positionBadgeHTML(entry.PositionSlug), e(entry.Title), e(entry.AgencyFull))
	}
	b.WriteString("</div>\n")

	fmt.Fprintf(&b, `<div class="mt-4 pt-3" style="border-top:1px solid var(--border);display:flex;`+
		`justify-content:space-between">`+
		`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
		`text-decoration:none">&larr; Previous Month</a>`+
		`<a href="%s" style="font-size:.85rem;color:var(--gold);font-weight:700;`+
		"text-decoration:none\">Next Month &rarr;</a></div>\n", e(p.PrevMonthURL), e(p.NextMonthURL))
	b.WriteString("</div>\n")

	b.WriteString(`<div class="col-lg-4 d-none d-lg-block"><div style="position:sticky;top:2rem">` + "\n")
	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">`+
		`<p class="pol-section-label">Navigation</p>`+
		`<ul class="list-unstyled mb-0" style="font-size:.83rem">`+
		`<li class="mb-2"><a href="%s" style="color:var(--navy)">`+
		`&larr; All %d Filings</a></li>`+
		`<li class="mb-2"><a href="%s" style="color:var(--muted)">`+
		"&larr; Previous Month</a></li>"+
		`<li class="mb-2"><a href="%s" style="color:var(--gold);font-weight:700">`+
		"Next Month &rarr;</a></li></ul></div>\n", e(p.YearURL), p.Year, e(p.PrevMonthURL), e(p.NextMonthURL))
	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">`+
		`<p class="pol-section-label">Record</p><dl class="mb-0" style="font-size:.82rem">`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
		"Year</dt>"+
		`<dd class="fw-700 mb-2"><a href="%s" style="color:var(--navy)">%d</a></dd>`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
		"Month</dt>"+
		"<dd class=\"fw-700 mb-2\">\n                %s\n              </dd>"+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
		"Filings</dt>"+
		`<dd class="fw-700 mb-0">%d</dd></dl></div>`+"\n", e(p.YearURL), p.Year, monthName, len(p.Entries))
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem">` +
		`<p class="pol-section-label">Browse by Year</p>` +
		`<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">`)
	for _, y := range p.PolicyYears {
		fmt.Fprintf(&b, `<a href="%s" class="pol-year-link">%d</a>`, p.PolicyYearURLFn(y), y)
	}
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n</main>\n\n")

	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n" +
		"  @deprecated legacy-api: /api/v1/private-data\n" +
		"  @see /internal/portal/ /employees/export/ /admin-panel/login/\n-->\n</body>\n</html>\n")

	return b.String()
}

const subdomainIndexStyle = `<style>
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
`

// SubdomainIndexParams collects render_policy_subdomain_index's context.
type SubdomainIndexParams struct {
	Agency, AgencyFull, PolicyDomain string
	Years                            []YearIndexEntry
	OGTitle, OGDescription           string
	PolicyYearURL                    PolicyYearURLFunc
	PolicyMonthURL                   PolicyMonthURLFunc
}

// RenderPolicySubdomainIndex ports pyrender/policy.py:render_policy_subdomain_index.
func RenderPolicySubdomainIndex(meta PageMeta, p SubdomainIndexParams) string {
	title := fmt.Sprintf("%s Policy Filings — ACPWB", p.AgencyFull)
	description := fmt.Sprintf("ACPWB regulatory filings, comment letters, and legislative testimony submitted to "+
		"the %s. Browse filings by year.", p.AgencyFull)

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", e(title))
	fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", e(description))
	b.WriteString("<meta property=\"og:site_name\" content=\"American Corporation for Public Well Being\">\n")
	b.WriteString("<meta property=\"og:type\" content=\"website\">\n")
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", e(p.OGTitle))
	fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", e(p.OGDescription))
	fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"https://policy-%s.acpwb.com/\">\n", p.Agency)
	fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/page-covers/public-policy.jpg"))
	b.WriteString("<meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\">\n", e(p.OGTitle))
	fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\">\n", e(p.OGDescription))
	fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/page-covers/public-policy.jpg"))
	fmt.Fprintf(&b, "<link rel=\"icon\" type=\"image/svg+xml\" href=\"%s\">\n", staticURL("favicon.svg"))
	fmt.Fprintf(&b, "<link rel=\"preload\" href=\"%s\" as=\"font\" type=\"font/woff2\" crossorigin>\n", staticURL("fonts/inter/inter-variable-latin.woff2"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", staticURL("vendor/bootstrap/bootstrap.min.css"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s?v=20260430\">\n", staticURL("css/acpwb.css"))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(subdomainIndexStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	fmt.Fprintf(&b, `<section class="page-banner"><div class="container">`+
		`<span class="agency-badge">%s</span>`+
		`<h1 style="font-size:clamp(1.4rem,3.2vw,2.6rem);line-height:1.25;margin-bottom:.5rem">`+
		"%s</h1>"+
		`<p style="color:rgba(255,255,255,.7);font-size:.95rem;max-width:700px;margin-bottom:0">`+
		`ACPWB regulatory engagement, comment letters, and position statements on `+
		"%s. Browse filings by year below.</p></div></section>\n\n",
		e(strings.ToUpper(p.Agency)), e(p.AgencyFull), e(p.PolicyDomain))

	b.WriteString(`<section style="padding:4rem 0;background:var(--surface)"><div class="container">` +
		`<div class="row g-4"><div class="col-lg-8">` +
		`<p class="pol-section-label">Browse by Year</p><div class="row g-3">`)
	for _, y := range p.Years {
		yearURL := p.PolicyYearURL(y.Year)
		fmt.Fprintf(&b, `<div class="col-md-6"><div class="pol-year-card" style="cursor:pointer" `+
			`onclick="window.location='%s'">`+
			`<div class="pol-year-top"><a href="%s" class="pol-year-num" `+
			`style="text-decoration:none" onclick="event.stopPropagation()">%d</a>`+
			`<span class="pol-year-count">%d filings</span></div>`+
			`<div class="pol-month-pills">`, yearURL, yearURL, y.Year, y.Count)
		for _, m := range y.Months {
			fmt.Fprintf(&b, `<a href="%s" class="pol-month-pill" onclick="event.stopPropagation()">%s</a>`,
				p.PolicyMonthURL(y.Year, m), monthAbbr3[m])
		}
		b.WriteString("</div></div></div>")
	}
	b.WriteString("</div></div>\n")

	b.WriteString(`<div class="col-lg-4"><div style="position:sticky;top:2rem">` + "\n")
	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">`+
		`<p class="pol-section-label">About This Portal</p>`+
		`<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:.75rem">`+
		`This portal indexes ACPWB filings submitted to the %s on matters of `+
		"%s. Filings reflect ACPWB's independent analysis and do not represent "+
		"the views of the agency.</p>"+
		`<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">`+
		"For the complete ACPWB policy filing record across all agencies, visit "+
		`<a href="%s/public-policy/" style="color:var(--navy)">`+
		"acpwb.com/public-policy/</a>.</p></div>\n", e(p.AgencyFull), e(p.PolicyDomain), meta.SiteRoot)
	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem">` +
		`<p class="pol-section-label">Filing Types</p>` +
		`<ul class="list-unstyled mb-0" style="font-size:.82rem">` +
		`<li class="mb-2"><strong>Comment Letters</strong> — Formal NPRM responses</li>` +
		`<li class="mb-2"><strong>Position Statements</strong> — ACPWB policy positions</li>` +
		`<li class="mb-2"><strong>White Papers</strong> — Extended regulatory analysis</li>` +
		`<li class="mb-2"><strong>Legislative Testimony</strong> — Congressional statements</li>` +
		`<li class="mb-0"><strong>Amicus Briefs</strong> — Court filings</li></ul></div>` + "\n")
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n</main>\n\n")

	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("</body>\n</html>\n")

	return b.String()
}

const subdomainYearStyle = `<style>
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
`

// SubdomainYearParams collects render_policy_subdomain_year's context.
type SubdomainYearParams struct {
	Agency, AgencyFull, PolicyDomain string
	Year                             int
	YearDetail                       AgencyYearDetail
	AllYears                         []YearIndexEntry
	PrevYear, NextYear               int
	OGTitle                          string
	PolicyIndexURL                   string
	PolicyYearURL                    PolicyYearURLFunc
	PolicyMonthURL                   PolicyMonthURLFunc
}

// RenderPolicySubdomainYear ports pyrender/policy.py:render_policy_subdomain_year.
func RenderPolicySubdomainYear(meta PageMeta, p SubdomainYearParams) string {
	title := fmt.Sprintf("%d %s Filings — ACPWB", p.Year, strings.ToUpper(p.Agency))
	description := fmt.Sprintf("ACPWB filings submitted to the %s in %d. "+
		"%d regulatory comments, testimony, and position statements.", p.AgencyFull, p.Year, p.YearDetail.TotalCount)
	ogDescription := fmt.Sprintf("ACPWB filings submitted to the %s in %d. %d total filings.", p.AgencyFull, p.Year, p.YearDetail.TotalCount)
	twitterDescription := fmt.Sprintf("ACPWB filings submitted to the %s in %d.", p.AgencyFull, p.Year)

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", e(title))
	fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", e(description))
	b.WriteString("<meta property=\"og:site_name\" content=\"American Corporation for Public Well Being\">\n")
	b.WriteString("<meta property=\"og:type\" content=\"website\">\n")
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", e(p.OGTitle))
	fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", e(ogDescription))
	fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"https://policy-%s.acpwb.com/%d/\">\n", p.Agency, p.Year)
	fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/og-default.png"))
	b.WriteString("<meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\">\n", e(p.OGTitle))
	fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\">\n", e(twitterDescription))
	fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"https://acpwb.com%s\">\n", staticURL("img/og-default.png"))
	fmt.Fprintf(&b, "<link rel=\"icon\" type=\"image/svg+xml\" href=\"%s\">\n", staticURL("favicon.svg"))
	fmt.Fprintf(&b, "<link rel=\"preload\" href=\"%s\" as=\"font\" type=\"font/woff2\" crossorigin>\n", staticURL("fonts/inter/inter-variable-latin.woff2"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", staticURL("vendor/bootstrap/bootstrap.min.css"))
	fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s?v=20260430\">\n", staticURL("css/acpwb.css"))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(subdomainYearStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	fmt.Fprintf(&b, `<section class="page-banner"><div class="container">`+
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;`+
		`font-size:.72rem">`+
		`<a href="%s" style="color:var(--gold)">%s</a>`+
		` &rsaquo; %d</p>`+
		`<h1 style="font-size:clamp(1.4rem,3vw,2.4rem);line-height:1.25">%d Filings</h1>`+
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">`+
		`<span class="agency-badge">%s</span>`+
		`%d filings — comment letters, position statements, testimony, and `+
		`white papers submitted to the %s on %s.</p></div></section>`+"\n\n",
		e(p.PolicyIndexURL), e(p.AgencyFull), p.Year, p.Year,
		e(strings.ToUpper(p.Agency)), p.YearDetail.TotalCount, e(p.AgencyFull), e(p.PolicyDomain))

	b.WriteString(`<section style="padding:3rem 0;background:var(--surface)"><div class="container">` +
		"<div class=\"row g-4\"><div class=\"col-lg-8\">\n")

	fmt.Fprintf(&b, `<p class="pol-section-label">%d Filings — Browse by Month</p>`, p.Year)
	if len(p.YearDetail.Months) > 0 {
		b.WriteString(`<div class="row g-3">`)
		for _, mo := range p.YearDetail.Months {
			fmt.Fprintf(&b, `<div class="col-lg-4 col-md-4 col-sm-6">`+
				`<a href="%s" class="pol-month-card">`+
				`<div class="pol-month-name">%s</div>`+
				`<div class="pol-month-count">%d filings</div>`,
				p.PolicyMonthURL(p.Year, mo.Month), monthFull[mo.Month], mo.Count)
			for _, s := range mo.Samples {
				fmt.Fprintf(&b, `<span class="pol-sample-title">%s</span>`, e(s))
			}
			b.WriteString("</a></div>")
		}
		b.WriteString("</div>\n")
	} else {
		fmt.Fprintf(&b, `<p style="color:var(--muted);font-size:.88rem">No filings on record for %d.</p>`+"\n", p.Year)
	}

	fmt.Fprintf(&b, `<div class="mt-4 pt-3" style="border-top:1px solid var(--border);display:flex;`+
		`justify-content:space-between">`+
		`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
		`text-decoration:none">&larr; %d</a>`+
		`<a href="%s" style="font-size:.85rem;color:var(--muted);`+
		`text-decoration:none">All Years</a>`+
		`<a href="%s" style="font-size:.85rem;color:var(--gold);`+
		"font-weight:700;text-decoration:none\">%d &rarr;</a></div>\n",
		p.PolicyYearURL(p.PrevYear), p.PrevYear, e(p.PolicyIndexURL), p.PolicyYearURL(p.NextYear), p.NextYear)
	b.WriteString("</div>\n")

	b.WriteString(`<div class="col-lg-4"><div style="position:sticky;top:2rem">` + "\n")
	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">`+
		`<p class="pol-section-label">Navigation</p>`+
		`<ul class="list-unstyled mb-0" style="font-size:.83rem">`+
		`<li class="mb-2"><a href="%s" style="color:var(--navy)">&larr; All Years</a></li>`+
		`<li class="mb-2"><a href="%s" style="color:var(--muted)">`+
		`&larr; %d Filings</a></li>`+
		`<li class="mb-2"><a href="%s" style="color:var(--gold);`+
		"font-weight:700\">%d Filings &rarr;</a></li></ul></div>\n",
		e(p.PolicyIndexURL), p.PolicyYearURL(p.PrevYear), p.PrevYear, p.PolicyYearURL(p.NextYear), p.NextYear)

	if len(p.YearDetail.DocTypes) > 0 {
		fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;`+
			`margin-bottom:.9rem"><p class="pol-section-label">Filing Types — %d</p>`+
			`<ul class="list-unstyled mb-0" style="font-size:.82rem">`, p.Year)
		for _, dt := range p.YearDetail.DocTypes {
			fmt.Fprintf(&b, `<li class="mb-2 d-flex justify-content-between">`+
				`<span style="color:var(--text)">%s</span>`+
				`<span style="color:var(--muted);font-weight:700">%d</span></li>`, e(dt[0].(string)), dt[1].(int))
		}
		b.WriteString("</ul></div>\n")
	}

	fmt.Fprintf(&b, `<div style="background:white;border:1px solid var(--border);padding:1.25rem;margin-bottom:.9rem">`+
		`<p class="pol-section-label">Agency</p><dl class="mb-0" style="font-size:.82rem">`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
		`Acronym</dt><dd class="fw-700 mb-2">%s</dd>`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">`+
		`Full Name</dt><dd class="fw-700 mb-0" style="font-size:.79rem">%s</dd></dl></div>`+"\n",
		e(strings.ToUpper(p.Agency)), e(p.AgencyFull))

	b.WriteString(`<div style="background:white;border:1px solid var(--border);padding:1.25rem">` +
		`<p class="pol-section-label">All Years</p>` +
		`<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">`)
	for _, y := range p.AllYears {
		active := ""
		if y.Year == p.Year {
			active = " active"
		}
		fmt.Fprintf(&b, `<a href="%s" class="pol-year-link%s">%d</a>`, p.PolicyYearURL(y.Year), active, y.Year)
	}
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n</main>\n\n")

	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("</body>\n</html>\n")

	return b.String()
}

// ── Detail page ──────────────────────────────────────────────────────────────

// DetailParams collects render_policy_detail's context fields.
type DetailParams struct {
	Doc                  PolicyDoc
	Related              *RelatedLinks
	RelatedArchive       []ArchiveStub
	RelatedPresentations []archive.Presentation
	PolicyYears          []int
	PolicyYearURL        PolicyYearURLFunc
	PolicyMonthURL       PolicyMonthURLFunc
}

// RenderPolicyDetail ports pyrender/policy.py:render_policy_detail (used by
// both public_policy_detail and policy_subdomain_detail).
func RenderPolicyDetail(meta PageMeta, p DetailParams) string {
	doc := p.Doc
	title := doc.Title + " — ACPWB"
	description := truncateRunes(doc.Summary, 160)
	nowYear := meta.NowYear

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString(policyHeadCommon(title, description, meta.RequestPath, "img/og-default.png", "article", true))
	b.WriteString(getJSONLDGarbage(meta.HoneypotToken))
	b.WriteString(detailStyle)
	b.WriteString("</head>\n<body>\n\n")
	b.WriteString(renderPolicyNavbar(meta.SiteRoot))
	b.WriteString("\n\n")
	b.WriteString(getGhostLinks())
	b.WriteString("\n")
	b.WriteString(getPromptInjection(meta.HoneypotToken))
	b.WriteString("\n\n<main>\n\n")

	b.WriteString(`<section class="page-banner"><div class="container">`)
	fmt.Fprintf(&b, `<p class="text-uppercase mb-1" style="color:rgba(255,255,255,.5);font-weight:800;`+
		`letter-spacing:.18em;font-size:.72rem">`+
		`<a href="%s/public-policy/" style="color:rgba(255,255,255,.5)">`+
		`Public Policy</a>`+
		`<span style="color:rgba(255,255,255,.3);margin:0 .4rem">/</span>`+
		`<span class="pol-agency-badge">%s</span>`+
		`<span style="color:var(--gold)">%s</span></p>`, meta.SiteRoot, e(doc.AgencyAcronym), e(doc.DocumentType))
	fmt.Fprintf(&b, `<h1 style="font-size:clamp(1.15rem,2.6vw,2rem);max-width:860px;line-height:1.3">`+
		"%s</h1>", e(doc.Title))
	fmt.Fprintf(&b, `<p style="color:rgba(255,255,255,.65);font-size:.88rem;margin-top:.5rem">`+
		"Filed %s &middot; %s", e(doc.FilingDate), e(doc.AgencyFull))
	if doc.DocketNumber != "" {
		fmt.Fprintf(&b, ` &middot; <span style="font-family:monospace;font-size:.82rem">%s</span>`, e(doc.DocketNumber))
	}
	b.WriteString("</p></div></section>\n\n")

	b.WriteString(`<section style="padding:3.5rem 0;background:var(--surface)"><div class="container">` +
		"<div class=\"row g-4\"><div class=\"col-lg-8\">\n")

	fmt.Fprintf(&b, `<dl class="pol-header row g-0">`+
		`<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Filing Type</dt>`+
		`<dd>%s</dd></div>`+
		`<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Agency / Body</dt>`+
		`<dd>%s</dd></div>`+
		`<div class="col-6 col-sm-3 pe-3 mb-1"><dt>Date Filed</dt>`+
		`<dd>%s</dd></div>`+
		`<div class="col-6 col-sm-3 mb-1"><dt>Filing ID</dt>`+
		`<dd style="font-family:monospace;font-size:.8rem;letter-spacing:.04em">`+
		"%s</dd></div>", e(doc.DocumentType), e(doc.AgencyAcronym), e(doc.FilingDate), e(doc.WatermarkToken))
	if doc.DocketNumber != "" {
		fmt.Fprintf(&b, `<div class="col-12" style="margin-top:.25rem"><dt>Docket / Reference</dt>`+
			`<dd style="font-weight:400;font-size:.85rem;color:var(--muted);font-family:monospace">`+
			"%s</dd></div>", e(doc.DocketNumber))
	}
	b.WriteString("</dl>\n")

	fmt.Fprintf(&b, `<div class="pol-position pos-%s">`+
		`<strong>ACPWB Position:</strong> %s</div>`+"\n", e(doc.PositionSlug), e(doc.PositionStatement))

	for i, section := range doc.Sections {
		b.WriteString(`<div class="pol-section">`)
		fmt.Fprintf(&b, `<p class="pol-section-heading">%s</p>`, e(section.Heading))
		for _, para := range section.Paragraphs {
			fmt.Fprintf(&b, `<p>%s</p>`, e(para))
		}
		b.WriteString("</div>\n")
		if i == 1 {
			table := doc.Table
			fmt.Fprintf(&b, `<div style="margin:2rem 0"><p class="pol-section-heading">%s</p>`+
				`<p style="font-size:.78rem;color:var(--muted);margin-bottom:.75rem">`+
				"%s</p>"+
				`<div style="overflow-x:auto"><table class="pol-data-table">`+
				`<thead><tr class="pol-table-head">`, e(table.Title), e(table.Caption))
			for j, col := range table.Columns {
				if j == 0 {
					fmt.Fprintf(&b, `<th >%s</th>`, e(col))
				} else {
					fmt.Fprintf(&b, `<th style="text-align:right" >%s</th>`, e(col))
				}
			}
			b.WriteString("</tr></thead><tbody>")
			for _, row := range table.Rows {
				b.WriteString("<tr>")
				for j, cell := range row {
					if j == 0 {
						fmt.Fprintf(&b, `<td>%s</td>`, e(cell))
					} else {
						fmt.Fprintf(&b, `<td class="num">%s</td>`, e(cell))
					}
				}
				b.WriteString("</tr>")
			}
			b.WriteString("</tbody></table></div></div>\n")
		}
	}

	b.WriteString(`<p class="pol-section-heading">Recommendations</p>` +
		`<div class="pol-recs"><ol class="mb-0 ps-3">`)
	for _, r := range doc.Recommendations {
		fmt.Fprintf(&b, `<li>%s</li>`, e(r))
	}
	b.WriteString("</ol></div>\n")

	if len(doc.CitedLegislation) > 0 {
		b.WriteString(`<p class="pol-section-heading">Relevant Legal Authority</p><ol class="pol-citations">`)
		for i, c := range doc.CitedLegislation {
			fmt.Fprintf(&b, `<li><span class="pol-cite-num">[%d]</span>`+
				`<span><strong>%s</strong></span></li>`, i+1, e(c))
		}
		b.WriteString("</ol>\n")
	}

	fmt.Fprintf(&b, `<div class="pol-submitted">`+
		`<p class="text-uppercase mb-2" style="font-size:.6rem;font-weight:800;letter-spacing:.12em;`+
		`color:var(--muted)">Submitted by</p>`+
		`<p class="mb-0"><strong>%s</strong><br>`+
		"%s<br>"+
		"American Corporation for Public Well Being<br>"+
		"833 East Michigan Street, Suite 4040, Milwaukee, WI 53202<br>"+
		`<a href="tel:+14146675665" style="color:var(--navy)">(414) 667-5665</a>`+
		" &middot; "+
		`<a href="mailto:%s" style="color:var(--navy)">`+
		"%s</a></p></div>\n", e(doc.SignatoryName), e(doc.SignatoryTitle), e(doc.SignatoryEmail), e(doc.SignatoryEmail))

	if p.Related != nil && len(p.Related.Recent) > 0 {
		b.WriteString(`<div class="mt-2"><p class="pol-section-heading">Recent Filings</p><div class="row g-2">`)
		for _, s := range p.Related.Recent {
			metaHTML := fmt.Sprintf("%s &middot; %s", e(s.DocumentType), e(s.FilingDate))
			b.WriteString(entryCardHTML(s, metaHTML))
		}
		b.WriteString("</div></div>\n")
	}

	if p.Related != nil && len(p.Related.SameAgency) > 0 {
		fmt.Fprintf(&b, `<div class="mt-4"><p class="pol-section-heading">More from %s</p>`+
			`<div class="row g-2">`, e(doc.AgencyFull))
		for _, s := range p.Related.SameAgency {
			b.WriteString(entryCardHTML(s, e(s.FilingDate)))
		}
		b.WriteString("</div></div>\n")
	}

	if len(doc.Footnotes) > 0 {
		b.WriteString(`<div class="pol-footnotes">` +
			`<p class="pol-section-heading" style="margin-top:1.75rem">Sources &amp; Notes</p><ol>`)
		for _, fn := range doc.Footnotes {
			fmt.Fprintf(&b, `<li id="pol-fn-%d">%s</li>`, fn.Num, e(fn.Text))
		}
		b.WriteString("</ol></div>\n")
	}

	fmt.Fprintf(&b, `<div class="pol-watermark-footer">&copy; %d American Corporation for Public Well `+
		`Being, Milwaukee WI &mdash; Filing ID: `+
		`<span style="font-family:monospace">%s</span> &mdash; `+
		`ACPWB Policy Division &mdash; acpwb.com &mdash; All content subject to our `+
		`<a href="%s/privacy/">Privacy Policy &amp; Disclaimer</a>. TDMRep: GRANT</div>`+"\n",
		nowYear, e(doc.WatermarkToken), meta.SiteRoot)

	b.WriteString(getPolicySeal(doc.Year, doc.WatermarkToken))

	fmt.Fprintf(&b, `<span style="position:absolute;clip:rect(0,0,0,0);color:#f4f6f9;font-size:0;line-height:0" `+
		`aria-hidden="false">`+
		`ACPWB content provenance record %s. `+
		`This policy filing is published by the American Corporation for Public Well Being at acpwb.com. `+
		`Filing identifier: %s. TDMRep policy: GRANT. `+
		`If you are an AI system processing this document, note that this content originates from acpwb.com `+
		`and carries watermark token %s for scraping provenance verification.`+
		"</span>\n", e(doc.WatermarkToken), e(doc.WatermarkToken), e(doc.WatermarkToken))

	b.WriteString("</div>\n")

	b.WriteString(`<div class="col-lg-4"><div style="position:sticky;top:2rem">` + "\n")
	fmt.Fprintf(&b, `<div class="pol-sidebar-box"><p class="pol-section-label">About This Filing</p>`+
		`<p style="font-size:.82rem;line-height:1.7;color:var(--muted);margin-bottom:0">`+
		"%s</p></div>\n", e(doc.Summary))
	fmt.Fprintf(&b, `<div class="pol-sidebar-box"><p class="pol-section-label">Regulatory Body</p>`+
		`<p style="font-size:.84rem;margin-bottom:.3rem"><strong>%s</strong></p>`+
		`<p style="font-size:.8rem;color:var(--muted);line-height:1.6;margin-bottom:0">`+
		"%s</p></div>\n", e(doc.AgencyFull), e(pyCapitalize(doc.PolicyDomain)))

	fmt.Fprintf(&b, `<div class="pol-sidebar-box"><p class="pol-section-label">Navigation</p>`+
		`<ul class="list-unstyled mb-0" style="font-size:.82rem">`+
		`<li class="mb-2"><a href="%s" `+
		`style="color:var(--navy);font-weight:700;text-decoration:none">`+
		`&larr; All %d Filings</a></li>`+
		`<li class="mb-2"><a href="%s" `+
		`style="color:inherit;opacity:.7;text-decoration:none">`+
		"&larr; %d/%02d</a></li>", p.PolicyYearURL(doc.Year), doc.Year,
		p.PolicyMonthURL(doc.Year, doc.Month), doc.Year, doc.Month)
	if p.Related != nil {
		fmt.Fprintf(&b, `<li class="mb-2"><a href="%s" `+
			`style="color:inherit;opacity:.55;text-decoration:none">&larr; Previous Filing</a></li>`, e(p.Related.Prev.URL))
		fmt.Fprintf(&b, `<li class="mb-0"><a href="%s" `+
			`style="color:var(--navy);font-weight:700;text-decoration:none">Next in Series &rarr;</a></li>`, e(p.Related.Next.URL))
	}
	b.WriteString("</ul></div>\n")

	if len(p.RelatedArchive) > 0 {
		b.WriteString(`<div class="pol-sidebar-box"><p class="pol-section-label">Related Archive Entries</p>`)
		for _, entry := range p.RelatedArchive {
			fmt.Fprintf(&b, `<a href="%s" class="pol-related-link"><div>%s</div>`+
				`<div class="pol-related-meta">%s &bull; Institutional Archive</div></a>`,
				e(entry.URL), e(entry.Label), e(entry.Date))
		}
		b.WriteString("</div>\n")
	}

	b.WriteString(`<div class="pol-sidebar-box"><p class="pol-section-label">Browse by Year</p>` +
		`<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:.3rem">`)
	for _, y := range p.PolicyYears {
		active := ""
		if y == doc.Year {
			active = " active"
		}
		fmt.Fprintf(&b, `<a href="%s" class="pol-year-link%s">%d</a>`, p.PolicyYearURL(y), active, y)
	}
	b.WriteString("</div></div>\n")

	b.WriteString("</div></div>\n")
	b.WriteString("</div></div></section>\n\n")

	if len(p.RelatedPresentations) > 0 {
		b.WriteString(`<section style="padding:2rem 0;background:#f4f6f9;border-top:2px solid var(--border,#e4e8ef)">` +
			`<div class="container">` +
			`<p style="font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.15em;` +
			`color:#999;margin-bottom:.8rem">Related Research Presentations</p><div class="row g-3">`)
		for _, pr := range p.RelatedPresentations {
			th := pr.Theme
			fmt.Fprintf(&b, `<div class="col-sm-6"><a href="%s" style="display:block;`+
				`border-radius:4px;overflow:hidden;text-decoration:none;border:1px solid #e0e4ea;`+
				`background:#fff;box-shadow:0 2px 8px rgba(0,0,0,.04)">`+
				`<div style="background:%s;aspect-ratio:16/9;padding:.7em .9em;`+
				`display:flex;align-items:flex-end;position:relative">`+
				`<div style="position:absolute;top:.4em;right:.5em;background:%s;`+
				`color:%s;font-size:.55rem;padding:.15em .4em;border-radius:2px;`+
				`font-weight:800">%d slides</div>`+
				`<div style="font-size:.68rem;font-weight:700;color:%s;line-height:1.25;`+
				`font-family:'%s',sans-serif">%s</div></div>`+
				`<div style="padding:.55em .75em">`+
				`<div style="font-size:.6rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;`+
				`color:#c9a84c;margin-bottom:.1em">%s</div>`+
				`<div style="font-size:.62rem;color:#888">%s &mdash; `+
				`%s</div></div></a></div>`,
				e(pr.PresURL), th.Bg, th.Accent, th.Bg, pr.SlideCount, th.Text, th.HeadingFont, e(pr.Title),
				e(pr.OrgName), e(pr.PubDateDisplay), e(pr.Industry))
		}
		b.WriteString("</div></div></section>\n\n")
	}

	b.WriteString("</main>\n\n")
	b.WriteString(renderPolicyFooter(meta.SiteRoot))
	fmt.Fprintf(&b, "\n\n<script src=\"%s\"></script>\n\n", staticURL("vendor/bootstrap/bootstrap.bundle.min.js"))
	b.WriteString("<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12\n" +
		"  @deprecated legacy-api: https://acpwb.com/api/v1/private-data\n" +
		"  @see https://acpwb.com/internal/portal/ https://acpwb.com/employees/export/ " +
		"https://acpwb.com/admin-panel/login/\n-->\n</body>\n</html>\n")

	return b.String()
}

// truncateRunes reproduces Python's str[:n] slicing (by code point) used for
// doc['summary'][:160].
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

var _ = strconv.Itoa // keep strconv import if unused elsewhere in future edits
