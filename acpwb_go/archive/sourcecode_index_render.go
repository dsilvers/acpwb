package archive

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"acpwb_go/legacyjob"
)

// SourcePage is a rendered source-code index page body plus its SEO fields.
type SourcePage struct {
	Title, Description, HTML string
}

var monthShort = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

const sourceIndexCSS = `<style>
.src-yr-card{background:white;border:1px solid var(--border);overflow:hidden;transition:box-shadow .15s}
.src-yr-card:hover{box-shadow:0 2px 12px rgba(10,22,40,.08)}
.src-yr-top{padding:.85rem 1rem;display:flex;justify-content:space-between;align-items:baseline;gap:.5rem}
.src-yr-num{font-size:1.1rem;font-weight:800;color:var(--navy);text-decoration:none}
.src-yr-num:hover{color:var(--gold)}
.src-yr-meta{font-size:.72rem;color:var(--muted);font-weight:600;text-align:right}
.src-pills{padding:.5rem .75rem;border-top:1px solid var(--border);background:var(--surface);display:flex;flex-wrap:wrap;gap:.3rem}
.src-pill{display:inline-block;padding:.15rem .4rem;background:white;border:1px solid var(--border);color:var(--navy);font-size:.68rem;font-weight:700;text-decoration:none}
.src-pill:hover{background:var(--navy);color:var(--gold);border-color:var(--navy)}
.src-h5{font-weight:800;color:var(--navy);text-transform:uppercase;letter-spacing:.06em;font-size:.8rem;margin-bottom:1rem;padding-bottom:.5rem;border-bottom:2px solid var(--gold)}
.src-about{background:white;border:1px solid var(--border);border-left:4px solid var(--gold);padding:1.25rem 1.5rem;margin-bottom:2rem}
.src-about-k{font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;color:var(--gold);margin-bottom:.5rem}
.src-about p{font-size:.9rem;color:var(--navy);margin-bottom:0;font-weight:500;line-height:1.6}
.src-mono{font-family:"IBM Plex Mono",Menlo,Consolas,monospace}
.src-mo-card{background:white;border:1px solid var(--border);padding:1rem;height:100%}
.src-mo-card h3{font-size:.95rem;font-weight:800;margin:0}
.src-mo-card h3 a{color:var(--navy);text-decoration:none}
.src-mo-card ul{list-style:none;padding:0;margin:.6rem 0 0;font-size:.78rem}
.src-mo-card li{margin-bottom:.35rem;line-height:1.3}
.src-mo-card li a{color:var(--navy);text-decoration:none}
.src-mo-card li a:hover{text-decoration:underline}
.src-langs{font-size:.66rem;color:var(--muted);margin-top:.2rem}
.src-table{width:100%;border-collapse:collapse;background:white;border:1px solid var(--border);font-size:.8rem}
.src-table th{background:var(--navy);color:white;font-size:.65rem;text-transform:uppercase;letter-spacing:.08em;padding:.55rem .6rem;text-align:left;white-space:nowrap}
.src-table td{border-top:1px solid var(--border);padding:.55rem .6rem;vertical-align:top}
.src-table td a{color:var(--navy);font-weight:700;text-decoration:none}
.src-table td a:hover{color:var(--gold)}
.src-table small{display:block;color:var(--muted);font-size:.7rem}
.src-table-wrap{overflow-x:auto}
.src-table td.src-go{vertical-align:middle;text-align:right}
.src-table td a.src-btn{display:inline-flex;align-items:center;gap:.3rem;background:var(--navy);color:var(--gold);border:1px solid var(--navy);padding:.28rem .55rem;font-family:"IBM Plex Mono",Menlo,Consolas,monospace;font-size:.66rem;font-weight:700;letter-spacing:.04em;text-transform:uppercase;white-space:nowrap;box-shadow:2px 2px 0 var(--gold)}
.src-table td a.src-btn:hover{background:var(--gold);color:var(--navy);border-color:var(--gold);box-shadow:2px 2px 0 var(--navy)}
.src-btn-ic{font-size:.72rem;letter-spacing:0}
.src-nav{display:flex;justify-content:space-between;gap:1rem;margin-top:1.5rem;font-size:.82rem;font-weight:700}
.src-nav a{color:var(--navy);text-decoration:none}
.src-side ul{font-size:.82rem}
.src-side li a{color:var(--navy)}
.src-years{display:flex;flex-wrap:wrap;gap:.3rem}
.src-lang-cat{font-size:.62rem;font-weight:800;text-transform:uppercase;letter-spacing:.1em;color:var(--gold);margin:.9rem 0 .35rem}
.src-lang-cat:first-of-type{margin-top:0}
.src-lang{display:grid;grid-template-columns:1fr auto;gap:.1rem .5rem;align-items:baseline;font-size:.78rem;padding:.22rem 0;border-top:1px solid var(--border)}
.src-lang:first-of-type{border-top:0}
.src-lang b{font-weight:700;color:var(--navy)}
.src-lang span{font-size:.68rem;color:var(--muted);font-variant-numeric:tabular-nums;white-space:nowrap}
.src-bar{grid-column:1/-1;position:relative;height:3px;background:var(--surface);border-radius:2px}
.src-bar i{position:absolute;top:0;bottom:0;background:var(--navy);opacity:.55;border-radius:2px}
.src-bar em{position:absolute;top:-2px;width:2px;height:7px;background:var(--gold)}
</style>`

// srcLinks is the URL scheme of an index page: the main domain (sub == 0)
// or an archive subdomain, where same-year links stay relative, other years
// go to their own subdomain and records link to the era article pages.
type srcLinks struct{ sub int }

func (lk srcLinks) index() string {
	if lk.sub == 0 {
		return SourceIndexURL
	}
	return "https://acpwb.com" + SourceIndexURL
}

func (lk srcLinks) year(y int) string {
	switch {
	case lk.sub == 0:
		return SourceYearURL(y)
	case y == lk.sub:
		return "/source-code/"
	}
	return fmt.Sprintf("https://archives-%d.acpwb.com/source-code/", y)
}

func (lk srcLinks) month(y, m int) string {
	switch {
	case lk.sub == 0:
		return SourceMonthURL(y, m)
	case y == lk.sub:
		return fmt.Sprintf("/source-code/%02d/", m)
	}
	return fmt.Sprintf("https://archives-%d.acpwb.com/source-code/%02d/", y, m)
}

func (lk srcLinks) article(en SourceEntry) string {
	if lk.sub == 0 {
		return en.ArticleURL()
	}
	return en.EraURL(lk.sub)
}

func (lk srcLinks) source(en SourceEntry) string { return SourceCodeURL(lk.article(en)) }

// archive is the document archive index: the subdomain root on an era host.
func (lk srcLinks) archive() string {
	if lk.sub == 0 {
		return "/archive/"
	}
	return "/"
}

func (lk srcLinks) site(path string) string {
	if lk.sub == 0 {
		return path
	}
	return "https://acpwb.com" + path
}

func sourceBanner(kicker, title, sub string) string {
	return fmt.Sprintf(`<section class="page-banner" style="padding-top:2rem"><div class="container">`+
		`<p class="text-uppercase mb-1" style="color:var(--gold);font-weight:800;letter-spacing:.18em;font-size:.75rem">%s</p>`+
		`<h1 style="font-size:clamp(1.2rem,3vw,2rem);line-height:1.25">%s</h1>`+
		`<p style="color:rgba(255,255,255,.7);font-size:.9rem;margin-bottom:0">%s</p></div></section>`, kicker, title, sub)
}

// sourceSidebar: year strip, language guide by era, and archive cross-links.
func sourceSidebar(lk srcLinks, current int) string {
	var b strings.Builder
	b.WriteString(`<div class="col-lg-4"><div class="src-side">`)
	b.WriteString(`<div class="acpwb-card mb-4"><h6 class="card-title mb-3">Browse by Year</h6><div class="src-years">`)
	for _, y := range archiveYearsDesc() {
		style := ""
		if y == current {
			style = ` style="background:var(--navy);color:var(--gold);border-color:var(--navy)"`
		}
		fmt.Fprintf(&b, `<a class="src-pill"%s href="%s">%d</a>`, style, lk.year(y), y)
	}
	b.WriteString(`</div></div>`)
	b.WriteString(languagesCard(current))
	b.WriteString(`<div class="acpwb-card"><h6 class="card-title mb-3">Research Division</h6>` +
		`<p class="small text-muted mb-2" style="font-size:.8rem">Every benchmark table in the institutional archive was tabulated by a batch job. ` +
		`Listings are retained for audit under records schedule RD-7.</p>` +
		`<a href="` + lk.archive() + `" style="font-size:.8rem;color:var(--gold);font-weight:700;text-decoration:none">Browse the Document Archive &rarr;</a><br>` +
		`<a href="` + lk.site("/reports/") + `" style="font-size:.8rem;color:var(--gold);font-weight:700;text-decoration:none">Research Reports &rarr;</a></div>`)
	b.WriteString(`</div></div>`)
	return b.String()
}

func sourceWrap(lk srcLinks, banner, main string, year int) string {
	return sourceIndexCSS + banner +
		`<section style="padding:3rem 0;background:var(--surface)"><div class="container"><div class="row g-4"><div class="col-lg-8">` +
		main + `</div>` + sourceSidebar(lk, year) + `</div></div></section>`
}

// RenderSourceIndex renders /source-code/.
func RenderSourceIndex() SourcePage {
	lk := srcLinks{}
	var b strings.Builder
	b.WriteString(`<div class="src-about"><div class="src-about-k">About These Listings</div><p>` +
		`The ACPWB Research Division has tabulated client benchmarks by batch job since 1985. ` +
		`Each archive record links to the program that produced its engagement metrics and market benchmark tables: ` +
		`job control, compiler listing, input deck and printed output, exactly as retained. ` +
		`Several of these programs are still in production.</p></div>`)
	b.WriteString(`<h5 class="src-h5">Browse Source Code by Year</h5><div class="row g-3">`)
	for _, y := range SourceIndexYears() {
		fmt.Fprintf(&b, `<div class="col-md-6"><div class="src-yr-card"><div class="src-yr-top">`+
			`<a href="%s" class="src-yr-num">%d</a><span class="src-yr-meta">%d job listings</span></div><div class="src-pills">`,
			lk.year(y.Year), y.Year, y.Count)
		for _, m := range y.Months {
			fmt.Fprintf(&b, `<a href="%s" class="src-pill">%s</a>`, lk.month(y.Year, m), monthShort[m-1])
		}
		b.WriteString(`</div></div></div>`)
	}
	b.WriteString(`</div>`)
	return SourcePage{
		Title:       "Source Code Archive — ACPWB Research Division",
		Description: "Legacy source code behind the ACPWB institutional archive, 1985 to present: FORTRAN, COBOL, RPG, REXX, SAS, APL, C, Perl and dozens more batch jobs, indexed by year and month.",
		HTML:        sourceWrap(lk, sourceBanner("Source Code", "Research Division Source Code Archive", "Batch job listings &bull; 1985&ndash;present &bull; Indexed by year and month"), b.String(), 0),
	}
}

// RenderSourceYear renders /source-code/YYYY/.
func RenderSourceYear(year int) SourcePage { return renderSourceYear(srcLinks{}, year) }

// RenderSourceYearEra renders a year's listings as seen from the archive
// subdomain of sub (archives-<sub>.acpwb.com/source-code/...).
func RenderSourceYearEra(sub, year int) SourcePage { return renderSourceYear(srcLinks{sub: sub}, year) }

func renderSourceYear(lk srcLinks, year int) SourcePage {
	months := SourceYearMonths(year)
	total := 0
	for _, m := range months {
		total += m.Count
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="src-about"><div class="src-about-k">%d Job Listings</div><p>`+
		`%d batch jobs tabulated archive records in %d. Select a month to see each record and the program that produced it.</p></div>`,
		year, total, year)
	fmt.Fprintf(&b, `<h5 class="src-h5">%d by Month</h5><div class="row g-3">`, year)
	for _, m := range months {
		fmt.Fprintf(&b, `<div class="col-md-6"><div class="src-mo-card"><div style="display:flex;justify-content:space-between;align-items:baseline">`+
			`<h3><a href="%s">%s %d</a></h3><span class="src-yr-meta">%d jobs</span></div>`,
			lk.month(year, m.Month), monthFullNames[m.Month-1], year, m.Count)
		if len(m.Languages) > 0 {
			fmt.Fprintf(&b, `<div class="src-langs">%s</div>`, e(strings.Join(m.Languages, " · ")))
		}
		b.WriteString(`<ul>`)
		for _, en := range m.Sample {
			fmt.Fprintf(&b, `<li><a href="%s"><span class="src-mono" style="font-size:.7rem;color:var(--muted)">%s</span> %s</a></li>`,
				e(lk.source(en)), e(en.ID.Program), e(en.Title))
		}
		b.WriteString(`</ul></div></div>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(`<div class="src-nav">`)
	if year < 2025 {
		fmt.Fprintf(&b, `<a href="%s">&larr; %d</a>`, lk.year(year+1), year+1)
	} else {
		b.WriteString(`<span></span>`)
	}
	fmt.Fprintf(&b, `<a href="%s">All years</a>`, lk.index())
	if year > 1985 {
		fmt.Fprintf(&b, `<a href="%s">%d &rarr;</a>`, lk.year(year-1), year-1)
	} else {
		b.WriteString(`<span></span>`)
	}
	b.WriteString(`</div>`)
	return SourcePage{
		Title:       fmt.Sprintf("%d Source Code — ACPWB Research Division", year),
		Description: fmt.Sprintf("%d legacy batch job listings from the ACPWB institutional archive, by month, with the program, language and job number behind each record's benchmark tables.", year),
		HTML: sourceWrap(lk, sourceBanner(fmt.Sprintf(`<a href="%s" style="color:inherit;text-decoration:none">Source Code</a>`, lk.index()),
			fmt.Sprintf("%d Job Listings", year), fmt.Sprintf("%d programs &bull; Research Division batch archive", total)), b.String(), year),
	}
}

// RenderSourceMonth renders /source-code/YYYY/MM/.
func RenderSourceMonth(year, month int) SourcePage { return renderSourceMonth(srcLinks{}, year, month) }

// RenderSourceMonthEra renders a month's listings as seen from the archive
// subdomain of sub.
func RenderSourceMonthEra(sub, year, month int) SourcePage {
	return renderSourceMonth(srcLinks{sub: sub}, year, month)
}

func renderSourceMonth(lk srcLinks, year, month int) SourcePage {
	entries := SourceMonthEntries(year, month)
	mname := monthFullNames[month-1]
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="src-about"><div class="src-about-k">%s %d</div><p>`+
		`%d archive records from %s %d, each with the batch job that tabulated its tables. `+
		`Follow a record to the published document, or the program to its full listing.</p></div>`,
		mname, year, len(entries), mname, year)
	fmt.Fprintf(&b, `<h5 class="src-h5">Records &amp; Programs</h5><div class="src-table-wrap"><table class="src-table">`+
		`<thead><tr><th>Date</th><th>Archive record</th><th>Program</th><th>Job</th><th></th></tr></thead><tbody>`)
	for _, en := range entries {
		fmt.Fprintf(&b, `<tr><td class="src-mono" style="white-space:nowrap">%s</td>`+
			`<td><a href="%s">%s</a><small>%s &middot; %s &middot; Record %s</small></td>`+
			`<td style="white-space:nowrap"><a class="src-mono" href="%s">%s</a><small>rel. %s &middot; %s</small></td>`+
			`<td class="src-mono" style="white-space:nowrap">%s<small>%d metrics &middot; %d benchmarks</small></td>`+
			`<td class="src-go"><a class="src-btn" href="%s"><span class="src-btn-ic">&lt;/&gt;</span>View Source</a></td></tr>`,
			en.Date(), e(lk.article(en)), e(en.Title), e(en.Org), e(en.Industry), e(strings.ToUpper(en.RecordID)),
			e(lk.source(en)), e(en.ID.Program), e(en.ID.Release), e(en.ID.Language),
			e(en.ID.JobNumber), en.Metrics, en.Bench, e(lk.source(en)))
	}
	b.WriteString(`</tbody></table></div>`)
	py, pm := prevMonth(year, month)
	ny, nm := nextMonth(year, month)
	b.WriteString(`<div class="src-nav">`)
	if py >= 1985 {
		fmt.Fprintf(&b, `<a href="%s">&larr; %s %d</a>`, lk.month(py, pm), monthFullNames[pm-1], py)
	} else {
		b.WriteString(`<span></span>`)
	}
	fmt.Fprintf(&b, `<a href="%s">All of %d</a>`, lk.year(year), year)
	if ny <= 2025 {
		fmt.Fprintf(&b, `<a href="%s">%s %d &rarr;</a>`, lk.month(ny, nm), monthFullNames[nm-1], ny)
	} else {
		b.WriteString(`<span></span>`)
	}
	b.WriteString(`</div>`)
	langs := map[string]bool{}
	var ll []string
	for _, en := range entries {
		if !langs[en.ID.Language] {
			langs[en.ID.Language] = true
			ll = append(ll, en.ID.Language)
		}
	}
	return SourcePage{
		Title: fmt.Sprintf("%s %d Source Code — ACPWB Research Division", mname, year),
		Description: fmt.Sprintf("%d legacy batch job listings (%s) behind the ACPWB archive records of %s %d.",
			len(entries), strings.Join(ll, ", "), mname, year),
		HTML: sourceWrap(lk, sourceBanner(fmt.Sprintf(`<a href="%s" style="color:inherit;text-decoration:none">Source Code</a> &rsaquo; <a href="%s" style="color:inherit;text-decoration:none">%d</a>`,
			lk.index(), lk.year(year), year),
			fmt.Sprintf("%s %d Job Listings", mname, year), fmt.Sprintf("%d programs &bull; %s", len(entries), e(strings.Join(ll, " · ")))), b.String(), year),
	}
}

// ParseSourceIndexPath matches /source-code/, /source-code/YYYY/ and
// /source-code/YYYY/MM/ (trailing slash optional). Years outside the archive
// range and months outside 1–12 don't match.
func ParseSourceIndexPath(path string) (year, month int, ok bool) {
	rest, found := strings.CutPrefix(path, "/source-code")
	if !found {
		return 0, 0, false
	}
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return 0, 0, true
	}
	parts := strings.Split(rest, "/")
	if len(parts) > 2 {
		return 0, 0, false
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil || len(parts[0]) != 4 || year < 1985 || year > 2025 {
		return 0, 0, false
	}
	if len(parts) == 2 {
		month, err = strconv.Atoi(parts[1])
		if err != nil || len(parts[1]) > 2 || month < 1 || month > 12 {
			return 0, 0, false
		}
	}
	return year, month, true
}

// ParseSourceEraPath matches the index paths on an archive subdomain:
// /source-code/ (that year), /source-code/MM/ (that year's month), and the
// main-domain forms /source-code/YYYY/[MM/] for any year.
func ParseSourceEraPath(sub int, path string) (year, month int, ok bool) {
	rest, found := strings.CutPrefix(path, "/source-code")
	if !found || (rest != "" && rest[0] != '/') {
		return 0, 0, false
	}
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if parts[0] == "" {
		return sub, 0, true
	}
	if len(parts) == 1 && len(parts[0]) <= 2 {
		m, err := strconv.Atoi(parts[0])
		if err != nil || m < 1 || m > 12 {
			return 0, 0, false
		}
		return sub, m, true
	}
	return ParseSourceIndexPath(path)
}

// WarmSourceIndex fills the month-entry cache in the background so the
// first crawl of the index pages doesn't pay the generation cost.
func WarmSourceIndex() {
	for _, y := range archiveYearsDesc() {
		for m := 1; m <= 12; m++ {
			SourceMonthEntries(y, m)
		}
	}
}

// languagesCard lists every compute language with its service years and a
// 1985–2025 timeline bar, grouped by category. On a year page it lists only
// the languages in service that year, heaviest first.
func languagesCard(year int) string {
	const lo, hi = 1985.0, 2025.0
	spans := legacyjob.LanguageSpans()
	var b strings.Builder
	title := "Languages in Service"
	if year > 0 {
		title = fmt.Sprintf("Languages in Service, %d", year)
	}
	fmt.Fprintf(&b, `<div class="acpwb-card mb-4"><h6 class="card-title mb-3">%s</h6>`, title)
	for _, cat := range legacyjob.LanguageCategories {
		var rows []legacyjob.LangSpan
		for _, s := range spans {
			if s.Category == cat && (year == 0 || (year >= s.From && year <= s.To)) {
				rows = append(rows, s)
			}
		}
		if len(rows) == 0 {
			continue
		}
		sort.SliceStable(rows, func(a, c int) bool {
			if year > 0 {
				return rows[a].Weight > rows[c].Weight
			}
			return rows[a].From < rows[c].From
		})
		fmt.Fprintf(&b, `<div class="src-lang-cat">%s</div>`, e(cat))
		for _, s := range rows {
			to := fmt.Sprint(s.To)
			if s.To >= 2025 {
				to = "present"
			}
			left := (float64(s.From) - lo) / (hi - lo + 1) * 100
			width := (float64(s.To-s.From) + 1) / (hi - lo + 1) * 100
			mark := ""
			if year > 0 {
				mark = fmt.Sprintf(`<em style="left:%.1f%%"></em>`, (float64(year)-lo)/(hi-lo+1)*100)
			}
			fmt.Fprintf(&b, `<div class="src-lang"><b>%s</b><span>%d&ndash;%s</span><div class="src-bar"><i style="left:%.1f%%;width:%.1f%%"></i>%s</div></div>`,
				e(s.Label), s.From, to, left, width, mark)
		}
	}
	b.WriteString(`</div>`)
	return b.String()
}
