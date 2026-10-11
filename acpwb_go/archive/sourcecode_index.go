package archive

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"acpwb_go/legacyjob"
)

// Source-code index (/source-code/, /source-code/YYYY/, /source-code/YYYY/MM/).
// The year/month combos and the per-month article list reuse the archive
// index generators (archidx_* / archmo_* seeds, same as Django's
// archive_index / archive_month views), so every listed entry is a real
// archive article and its /source-code/ page.

// SourceIndexURL, SourceYearURL and SourceMonthURL are main-domain paths.
const SourceIndexURL = "/source-code/"

func SourceYearURL(y int) string     { return fmt.Sprintf("/source-code/%d/", y) }
func SourceMonthURL(y, m int) string { return fmt.Sprintf("/source-code/%d/%02d/", y, m) }

// ArchiveVariant picks the article template for a path (mirrors
// apps/honeypot/views.py's variant hash). Only "default" articles carry the
// stats tables that a source-code page reproduces.
func ArchiveVariant(year, month, day int, slug string) string {
	h := md5Hex(fmt.Sprintf("variant_%d%d%d%s", year, month, day, slug))
	var mod int
	for _, c := range h {
		mod = (mod*16 + strings.IndexRune("0123456789abcdef", c)) % 20
	}
	switch {
	case mod < 3:
		return "compliance"
	case mod < 6:
		return "minutes"
	}
	return "default"
}

type SourceYear struct {
	Year   int
	Count  int
	Months []int
}

// SourceIndexYears lists every year with its real listing count and the
// months that have at least one listing (the archive index shows a random
// sample of months; the source-code index shows them all).
func SourceIndexYears() []SourceYear {
	var out []SourceYear
	for _, y := range archiveYearsDesc() {
		sy := SourceYear{Year: y}
		for m := 1; m <= 12; m++ {
			if n := len(SourceMonthEntries(y, m)); n > 0 {
				sy.Count += n
				sy.Months = append(sy.Months, m)
			}
		}
		out = append(out, sy)
	}
	return out
}

// SourceEntry is one archive article plus the program that tabulated it.
type SourceEntry struct {
	Year, Month, Day int
	Slug             string
	Title, Org       string
	Industry         string
	RecordID         string
	Metrics, Bench   int
	ID               legacyjob.ID
}

func (e SourceEntry) ArticleURL() string {
	return ArchiveURL(e.Year, intPtr(e.Month), intPtr(e.Day), e.Slug)
}

func (e SourceEntry) SourceURL() string { return SourceCodeURL(e.ArticleURL()) }

// EraURL is the article URL as seen from an archive subdomain of fromYear.
func (e SourceEntry) EraURL(fromYear int) string {
	return ArchiveURLEra(e.Year, intPtr(e.Month), intPtr(e.Day), e.Slug, fromYear)
}

func (e SourceEntry) Date() string { return fmt.Sprintf("%04d-%02d-%02d", e.Year, e.Month, e.Day) }

type monthKey struct{ y, m int }

// The month space is finite (41 years x 12), so entries are cached forever.
var sourceMonthCache sync.Map

// SourceMonthEntries mirrors archive_month's entry list (archmo_{y}_{m}),
// keeping only default-variant articles, sorted by day.
func SourceMonthEntries(year, month int) []SourceEntry {
	k := monthKey{year, month}
	if v, ok := sourceMonthCache.Load(k); ok {
		return v.([]SourceEntry)
	}
	r := rngA(fmt.Sprintf("archmo_%d_%d", year, month))
	count := int(r.RandInt(12, 42))
	type raw struct {
		day  int
		slug string
	}
	var raws []raw
	for i := 0; i < count; i++ {
		day := int(r.RandInt(1, 28))
		slug := fmt.Sprintf("%s-%d", choice(r, archiveSlugs), r.RandInt(1000, 9999))
		raws = append(raws, raw{day, slug})
	}
	sort.SliceStable(raws, func(i, j int) bool { return raws[i].day < raws[j].day })
	seen := map[string]bool{}
	out := []SourceEntry{}
	for _, w := range raws {
		key := fmt.Sprintf("%d/%s", w.day, w.slug)
		if seen[key] || ArchiveVariant(year, month, w.day, w.slug) != "default" {
			continue
		}
		seen[key] = true
		c := GenerateArchiveContent(year, month, w.day, w.slug)
		out = append(out, SourceEntry{
			Year: year, Month: month, Day: w.day, Slug: w.slug,
			Title: c.Title, Org: c.Org, Industry: c.Industry, RecordID: c.RecordID,
			Metrics: len(c.MetricRows), Bench: len(c.PercentileTable),
			ID: legacyjob.Identity(year, month, w.day, w.slug),
		})
	}
	sourceMonthCache.Store(k, out)
	return out
}

// MonthSummary is one month card on the year page.
type SourceMonthSummary struct {
	Month     int
	Count     int
	Languages []string
	Sample    []SourceEntry
}

func SourceYearMonths(year int) []SourceMonthSummary {
	var out []SourceMonthSummary
	for m := 1; m <= 12; m++ {
		es := SourceMonthEntries(year, m)
		langs := map[string]bool{}
		var ll []string
		for _, e := range es {
			if !langs[e.ID.Language] {
				langs[e.ID.Language] = true
				ll = append(ll, e.ID.Language)
			}
		}
		sort.Strings(ll)
		s := es
		if len(s) > 4 {
			s = s[:4]
		}
		out = append(out, SourceMonthSummary{Month: m, Count: len(es), Languages: ll, Sample: s})
	}
	return out
}

func prevMonth(y, m int) (int, int) {
	if m == 1 {
		return y - 1, 12
	}
	return y, m - 1
}

func nextMonth(y, m int) (int, int) {
	if m == 12 {
		return y + 1, 1
	}
	return y, m + 1
}

// sourceMore builds the "Related software projects" groups for the bottom of a
// source-code detail page: neighbours in the month listing, the rest of the
// month, the same language in other years, and the same calendar date in
// other years. url maps an entry to the URL scheme of the current page.
func sourceMore(year, month, day int, slug string, lang string, url func(SourceEntry) string) (groups []legacyjob.MoreGroup, prev, next *legacyjob.MoreLink) {
	link := func(e SourceEntry) legacyjob.MoreLink {
		return legacyjob.MoreLink{
			URL: SourceCodeURL(url(e)), Title: e.Title,
			Program: e.ID.Program + " rel. " + e.ID.Release, Language: e.ID.Language,
			Date: e.Date(), Org: e.Org, Job: e.ID.JobNumber,
		}
	}
	r := newSeededPicker(fmt.Sprintf("srcmore_%d_%d_%d_%s", year, month, day, slug))

	month0 := SourceMonthEntries(year, month)
	self := -1
	for i, e := range month0 {
		if e.Day == day && e.Slug == slug {
			self = i
		}
	}
	if self > 0 {
		l := link(month0[self-1])
		prev = &l
	}
	if self >= 0 && self < len(month0)-1 {
		l := link(month0[self+1])
		next = &l
	}

	var same []legacyjob.MoreLink
	for _, i := range r.perm(len(month0)) {
		if i != self && len(same) < 6 {
			same = append(same, link(month0[i]))
		}
	}
	if len(same) > 0 {
		groups = append(groups, legacyjob.MoreGroup{
			Title: fmt.Sprintf("Also run in %s %d", monthFullNames[month-1], year),
			URL:   SourceMonthURL(year, month), More: "All " + monthFullNames[month-1] + " jobs",
			Links: same,
		})
	}

	years := archiveYearsDesc()
	var langLinks []legacyjob.MoreLink
	for _, yi := range r.perm(len(years)) {
		if len(langLinks) >= 6 {
			break
		}
		y := years[yi]
		if y == year {
			continue
		}
		m := r.intn(12) + 1
		for _, e := range SourceMonthEntries(y, m) {
			if e.ID.Language == lang {
				langLinks = append(langLinks, link(e))
				break
			}
		}
	}
	if len(langLinks) > 0 {
		groups = append(groups, legacyjob.MoreGroup{Title: "Other " + lang + " programs", Links: langLinks})
	}

	var dated []legacyjob.MoreLink
	for _, dy := range []int{-10, -5, -1, 1, 5, 10} {
		y := year + dy
		if y < 1985 || y > 2025 {
			continue
		}
		es := SourceMonthEntries(y, month)
		if len(es) == 0 {
			continue
		}
		best := es[0]
		for _, e := range es {
			if abs(e.Day-day) < abs(best.Day-day) {
				best = e
			}
		}
		dated = append(dated, link(best))
	}
	if len(dated) > 0 {
		groups = append(groups, legacyjob.MoreGroup{Title: "Around this date in other years", Links: dated})
	}

	return groups, prev, next
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// seededPicker is a tiny deterministic helper for link selection (no Python
// parity needed — these pages have no Django counterpart).
type seededPicker struct {
	r interface{ RandInt(a, b int64) int64 }
}

func newSeededPicker(seed string) seededPicker { return seededPicker{rngA(seed)} }

func (p seededPicker) intn(n int) int { return int(p.r.RandInt(0, int64(n-1))) }

func (p seededPicker) perm(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := p.intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}
