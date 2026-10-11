package archive

import (
	"fmt"
	"strings"

	"acpwb_go/legacyjob"
)

// SourceCodeURL is the /source-code/ page sitting next to an article URL.
func SourceCodeURL(articleURL string) string { return articleURL + "source-code/" }

// SourceCodeInput adapts a default-variant archive page into the generator's
// surface-agnostic input. era selects archive-subdomain (relative /
// cross-subdomain) URLs versus main-domain /archive/YYYY/... URLs.
func SourceCodeInput(year, month, day int, slug string, era bool) legacyjob.Input {
	c := GenerateArchiveContent(year, month, day, slug)
	url := func(y, m, d int, s string) string {
		if era {
			return ArchiveURLEra(y, intPtr(m), intPtr(d), s, year)
		}
		return ArchiveURL(y, intPtr(m), intPtr(d), s)
	}
	article := url(year, month, day, slug)

	in := legacyjob.Input{
		Year: year, Month: month, Day: day, Slug: slug,
		Path:     fmt.Sprintf("/archive/%d/%02d/%02d/%s", year, month, day, slug),
		Token:    legacyjob.WatermarkToken(slug),
		RecordID: c.RecordID,
		Title:    c.Title, BaseTitle: c.BaseTitle, Org: c.Org, Industry: c.Industry, Phase: c.Phase,
		EngCode: c.EngCode, DocVersion: c.DocVersion, Distribution: c.Distribution,
		PageCount: c.PageCount, FileSizeKB: c.FileSizeKB, PeerGroup: c.PeerGroup,
		FirstNames:    firstNames,
		LastNames:     lastNames,
		ArticleURL:    article,
		ExportCSVURL:  article + "export.csv",
		MetricsAnchor: fmt.Sprintf("acpwb-archive-%s-engagement-metrics-section", c.RecordID),
		BenchAnchor:   fmt.Sprintf("acpwb-archive-%s-benchmark-percentile-section", c.RecordID),
	}
	for _, r := range c.MetricRows {
		in.Metrics = append(in.Metrics, legacyjob.MetricIn{Name: r.Name, Baseline: legacyjob.ParseInt(r.Baseline), Current: legacyjob.ParseInt(r.Current)})
	}
	for _, r := range c.PercentileTable {
		bi := legacyjob.BenchIn{Name: r.Metric}
		for k, v := range []string{r.P10, r.P25, r.P33, r.P50, r.P67, r.P75, r.P90, r.P95} {
			bi.P[k] = legacyjob.ParseInt(v)
		}
		bi.Median = bi.P[3]
		in.Bench = append(in.Bench, bi)
	}
	for _, m := range c.EngagementTeam {
		in.Team = append(in.Team, legacyjob.Person{Name: m.Name, Title: m.Title, Email: m.Email})
	}
	for _, r := range c.Revisions {
		in.Revisions = append(in.Revisions, legacyjob.Revision{Version: r.Version, Date: r.Date, Description: r.Description, Author: r.Author})
	}
	for _, f := range c.Findings {
		in.Sentences = append(in.Sentences, f.Text)
	}
	in.Sentences = append(in.Sentences, c.ExecBullets...)
	for _, p := range c.Paragraphs {
		in.Sentences = append(in.Sentences, p.Text)
	}
	for _, f := range c.Footnotes {
		in.Sentences = append(in.Sentences, f.Text)
	}
	for _, r := range c.RelatedReports {
		in.RelatedTitles = append(in.RelatedTitles, r.Title)
	}
	if len(c.BulkHexJS) > 40 {
		in.Hex = append(in.Hex, c.BulkHexJS[:40]...)
	}
	for _, r := range GenRelatedPathData(year, month, day, slug) {
		in.Links = append(in.Links, legacyjob.CrossLink{
			URL:   SourceCodeURL(url(r.Year, r.Month, r.Day, r.FullSlug)),
			Label: r.Label, Date: r.Date,
			Year: r.Year, Month: r.Month, Day: r.Day, Slug: r.FullSlug,
		})
	}
	root := ""
	if era {
		// Archive subdomains link to main-domain pages absolutely, the same
		// way Django's site_root context var does.
		root = "https://acpwb.com"
	}
	// Archive subdomains carry their own year/month source-code index
	// (archives-YYYY.acpwb.com/source-code/[MM/]); "all years" is main-domain.
	lk := srcLinks{}
	if era {
		lk.sub = year
	}
	in.IndexURL = lk.index()
	in.YearURL = lk.year(year)
	in.MonthURL = lk.month(year, month)
	id := legacyjob.Identity(year, month, day, slug)
	in.More, in.Prev, in.Next = sourceMore(year, month, day, slug, id.Language, func(e SourceEntry) string {
		if era {
			return e.EraURL(year)
		}
		return e.ArticleURL()
	})
	for i := range in.More {
		if in.More[i].URL == SourceMonthURL(year, month) {
			in.More[i].URL = lk.month(year, month)
		} else if in.More[i].URL != "" {
			in.More[i].URL = root + in.More[i].URL
		}
	}
	return in
}

// AllArchiveYears is the year list the era shell's year browser shows.
func AllArchiveYears() []int { return archiveYearsDesc() }

// sourceCodeHref derives the /source-code/ URL from the page's export.csv URL
// (both sit next to the article URL, in either URL scheme).
func sourceCodeHref(exportCSVURL string) string {
	return strings.TrimSuffix(exportCSVURL, "export.csv") + "source-code/"
}

func sourceCaptionHTML(id legacyjob.ID, href string, era bool) string {
	if era {
		return fmt.Sprintf(`<div class="acpwb-source-caption" style="font-size:.72rem;opacity:.7;margin-top:.4rem">`+
			`Tabulated by <span style="font-family:monospace">%s rel. %s</span> (%s) &middot; `+
			`<a href="%s" class="era-nav-link">view source code &rarr;</a></div>`,
			e(id.Program), e(id.Release), e(id.Language), e(href))
	}
	return fmt.Sprintf(`<div class="acpwb-source-caption" style="font-size:.72rem;color:var(--muted);margin-top:.4rem">`+
		`Tabulated by <span style="font-family:monospace">%s rel. %s</span> (%s) &middot; `+
		`<a href="%s" style="color:var(--navy);font-weight:700">view source code &rarr;</a></div>`,
		e(id.Program), e(id.Release), e(id.Language), e(href))
}

func sourceSidebarCardHTML(id legacyjob.ID, href, rid string, era bool) string {
	if era {
		return fmt.Sprintf(`<div id="acpwb-archive-%s-sidebar-source-card" style="background:rgba(128,128,128,.07);`+
			`border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">`+
			`<div class="era-section-head" style="margin-bottom:.6rem">Computation Source</div>`+
			`<dl class="mb-2" style="font-size:.82rem">`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Program</dt>`+
			`<dd class="fw-700 mb-2">%s rel. %s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Language</dt>`+
			`<dd class="mb-2">%s</dd>`+
			`<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Job</dt>`+
			`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">%s</dd></dl>`+
			`<a href="%s" class="era-nav-link" style="font-size:.8rem">View source code &rarr;</a></div>`,
			rid, e(id.Program), e(id.Release), e(id.Language), e(id.JobNumber), e(href))
	}
	return fmt.Sprintf(`<div id="acpwb-archive-%s-sidebar-source-card" class="acpwb-card mb-4">`+
		`<h6 class="card-title mb-2">Computation Source</h6><dl class="mb-2" style="font-size:.82rem">`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Program</dt>`+
		`<dd class="fw-700 mb-2">%s rel. %s</dd>`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Language</dt>`+
		`<dd class="mb-2">%s</dd>`+
		`<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Job</dt>`+
		`<dd class="mb-0" style="font-family:monospace;font-size:.75rem;color:var(--muted)">%s</dd></dl>`+
		`<a href="%s" style="font-size:.8rem;color:var(--gold);font-weight:700;text-decoration:none">View source code &rarr;</a></div>`,
		rid, e(id.Program), e(id.Release), e(id.Language), e(id.JobNumber), e(href))
}

func sourceRefEntryHTML(id legacyjob.ID, href, date string) string {
	return fmt.Sprintf(`<div style="background:white;border:1px solid var(--border);border-left:3px solid var(--gold);`+
		`padding:.75rem 1rem;margin-bottom:.75rem">`+
		`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;`+
		`color:var(--muted);margin-bottom:.2rem">Source &bull; JOB LISTING &bull; %s</div>`+
		`<a href="%s" style="font-size:.9rem;font-weight:700;color:var(--navy);text-decoration:none">%s Job Listing %s</a>`+
		`<p class="small text-muted mb-0" style="font-size:.78rem;margin-top:.25rem">`+
		`%s batch job that tabulated the engagement metrics and market benchmark tables in this record.</p></div>`,
		e(date), e(href), e(id.Program), e(id.JobNumber), e(id.Language))
}

func sourceRefEntryEraHTML(id legacyjob.ID, href, date string) string {
	return fmt.Sprintf(`<a href="%s" class="era-entry-card" style="margin-bottom:.5rem">`+
		`<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">%s &bull; source</div>`+
		`<div style="font-size:.75rem;font-weight:600;line-height:1.35">%s Job Listing %s</div></a>`,
		e(href), e(date), e(id.Program), e(id.JobNumber))
}
