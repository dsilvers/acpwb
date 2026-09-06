package archive

import "fmt"

// Context is the Go equivalent of the merged context dict archive_trap()
// builds for the main-domain 'default' variant (apps/honeypot/views.py:976),
// restricted to the fields render_archive_default actually reads.
type Context struct {
	ArchiveContent

	Year, Month, Day int
	Slug             string

	YearURL      string
	MonthURL     string
	PrevEntryURL string
	NextEntryURL string
	ExportCSVURL string

	RelatedPaths         []RelatedPathView
	CrossYearReports     []CrossYearReport
	ArchiveYears         []int
	RelatedDocs          []RelatedDocView
	RelatedPolicy        []PolicyStub
	RelatedPresentations []Presentation
}

type RelatedPathView struct {
	URL   string
	Label string
	Date  string
}

type RelatedDocView struct {
	Label string
	URL   string
	Date  string
	Phase string
}

// BuildContext ports the context-assembly done inline in
// apps/honeypot/views.py:archive_trap() for the main-domain, 'default'
// variant branch only (on_archive_subdomain is always false in this port's
// scope).
func BuildContext(year, month, day int, slug string) Context {
	content := GenerateArchiveContent(year, month, day, slug)

	nextSlug, prevSlug := GenNavSlugs(year, month, day, slug)

	var prevDay, prevMonth, prevYear int
	if day > 1 {
		prevDay = day - 1
		prevMonth = month
		prevYear = year
	} else {
		prevDay = 28
		if month > 1 {
			prevMonth = month - 1
			prevYear = year
		} else {
			prevMonth = 12
			prevYear = year - 1
		}
	}

	relatedPathsRaw := GenRelatedPathData(year, month, day, slug)
	relatedPaths := make([]RelatedPathView, 0, len(relatedPathsRaw))
	for _, r := range relatedPathsRaw {
		relatedPaths = append(relatedPaths, RelatedPathView{
			URL:   ArchiveURL(r.Year, intPtr(r.Month), intPtr(r.Day), r.FullSlug),
			Label: r.Label,
			Date:  r.Date,
		})
	}

	crossYearReports := GenCrossYearReports(year, month, day, slug)

	relatedDocsRaw := GenRelatedDocsData(year, month, day, slug)
	relatedDocs := make([]RelatedDocView, 0, len(relatedDocsRaw))
	for _, d := range relatedDocsRaw {
		relatedDocs = append(relatedDocs, RelatedDocView{
			Label: d.Label,
			URL:   ArchiveURL(year, intPtr(month), intPtr(d.Day), d.FullSlug),
			Date:  d.Date,
			Phase: d.Phase,
		})
	}

	relatedPolicy := GetCrossPolicyStubs(year, month, day, slug)

	presCount := GenPresentationsCount(year, month, day, slug)
	truncSlug := slug
	if len(truncSlug) > 32 {
		truncSlug = truncSlug[:32]
	}
	contextSeed := fmt.Sprintf("archive_pres_%d_%d_%d_%s", year, month, day, truncSlug)
	relatedPresentations := GeneratePresentationsForContext(contextSeed, presCount)

	return Context{
		ArchiveContent: content,
		Year:           year, Month: month, Day: day, Slug: slug,
		YearURL:      ArchiveURL(year, nil, nil, ""),
		MonthURL:     ArchiveURL(year, intPtr(month), nil, ""),
		PrevEntryURL: ArchiveURL(prevYear, intPtr(prevMonth), intPtr(prevDay), prevSlug),
		NextEntryURL: ArchiveURL(year, intPtr(month), intPtr(day), nextSlug),
		ExportCSVURL: ArchiveURL(year, intPtr(month), intPtr(day), slug) + "export.csv",

		RelatedPaths:         relatedPaths,
		CrossYearReports:     crossYearReports,
		ArchiveYears:         archiveYearsDesc(),
		RelatedDocs:          relatedDocs,
		RelatedPolicy:        relatedPolicy,
		RelatedPresentations: relatedPresentations,
	}
}

func archiveYearsDesc() []int {
	out := make([]int, 0, 2025-1985+1)
	for y := 2025; y >= 1985; y-- {
		out = append(out, y)
	}
	return out
}
