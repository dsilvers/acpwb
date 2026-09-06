package archive

import "fmt"

// ArchiveURLEra ports the on_archive_subdomain=True branches of
// apps/honeypot/views.py:_archive_url (lines ~679-697): links to the SAME
// year as the current page are subdomain-relative; links to a DIFFERENT year
// are absolute URLs to that year's own archives-YYYY.acpwb.com subdomain.
// currentYear is the era page's own year (request.archive_year, always equal
// to the year param archive_trap() was called with on a subdomain request).
func ArchiveURLEra(targetYear int, month, day *int, slug string, currentYear int) string {
	if targetYear == currentYear {
		if month == nil {
			return "/"
		}
		if day == nil {
			return fmt.Sprintf("/%02d/", *month)
		}
		if slug != "" {
			return fmt.Sprintf("/%02d/%02d/%s/", *month, *day, slug)
		}
		return fmt.Sprintf("/%02d/%02d/", *month, *day)
	}
	base := fmt.Sprintf("https://archives-%d.acpwb.com", targetYear)
	if month == nil {
		return base + "/"
	}
	if day == nil {
		return fmt.Sprintf("%s/%02d/", base, *month)
	}
	if slug != "" {
		return fmt.Sprintf("%s/%02d/%02d/%s/", base, *month, *day, slug)
	}
	return fmt.Sprintf("%s/%02d/%02d/", base, *month, *day)
}

// sharedEraFields holds the page-level context fields archive_trap()
// assembles identically across all 3 variants when on_archive_subdomain is
// True, mirroring sharedPageFields (context_variants.go) for the main-domain
// port but with era-aware (subdomain-relative / cross-subdomain-absolute)
// URLs built via ArchiveURLEra instead of ArchiveURL.
type sharedEraFields struct {
	Year, Month, Day int
	Slug             string

	YearURL      string
	MonthURL     string
	PrevEntryURL string
	NextEntryURL string
	ExportCSVURL string

	RelatedPaths     []RelatedPathView
	CrossYearReports []CrossYearReport
	AllYears         []int
	RelatedDocs      []RelatedDocView
	RelatedPolicy    []PolicyStub
	YearData         YearData
}

// buildSharedEraFields ports the shared (variant-independent) portion of
// archive_trap()'s context assembly for the on_archive_subdomain=True branch.
func buildSharedEraFields(year, month, day int, slug string) sharedEraFields {
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
			URL:   ArchiveURLEra(r.Year, intPtr(r.Month), intPtr(r.Day), r.FullSlug, year),
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
			URL:   ArchiveURLEra(year, intPtr(month), intPtr(d.Day), d.FullSlug, year),
			Date:  d.Date,
			Phase: d.Phase,
		})
	}

	relatedPolicy := GetCrossPolicyStubs(year, month, day, slug)

	return sharedEraFields{
		Year: year, Month: month, Day: day, Slug: slug,
		YearURL:      ArchiveURLEra(year, nil, nil, "", year),
		MonthURL:     ArchiveURLEra(year, intPtr(month), nil, "", year),
		PrevEntryURL: ArchiveURLEra(prevYear, intPtr(prevMonth), intPtr(prevDay), prevSlug, year),
		NextEntryURL: ArchiveURLEra(year, intPtr(month), intPtr(day), nextSlug, year),
		ExportCSVURL: ArchiveURLEra(year, intPtr(month), intPtr(day), slug, year) + "export.csv",

		RelatedPaths:     relatedPaths,
		CrossYearReports: crossYearReports,
		AllYears:         archiveYearsDesc(),
		RelatedDocs:      relatedDocs,
		RelatedPolicy:    relatedPolicy,
		YearData:         YearDataFor(year),
	}
}

// EraContext is the Go equivalent of the merged context dict archive_trap()
// builds for the archives-YYYY.acpwb.com 'default' era variant.
type EraContext struct {
	ArchiveContent
	sharedEraFields
	RelatedPresentations []Presentation
}

// BuildEraContext ports archive_trap()'s context assembly for the
// on_archive_subdomain=True, 'default'-variant branch.
func BuildEraContext(year, month, day int, slug string) EraContext {
	content := GenerateArchiveContent(year, month, day, slug)
	shared := buildSharedEraFields(year, month, day, slug)

	presCount := GenPresentationsCount(year, month, day, slug)
	truncSlug := slug
	if len(truncSlug) > 32 {
		truncSlug = truncSlug[:32]
	}
	contextSeed := fmt.Sprintf("archive_pres_%d_%d_%d_%s", year, month, day, truncSlug)
	relatedPresentations := GeneratePresentationsForContext(contextSeed, presCount)

	return EraContext{
		ArchiveContent:       content,
		sharedEraFields:      shared,
		RelatedPresentations: relatedPresentations,
	}
}

// EraComplianceContext is the Go equivalent of the merged context dict
// archive_trap() builds for the archives-YYYY.acpwb.com 'compliance' era
// variant.
type EraComplianceContext struct {
	ComplianceContent
	sharedEraFields
}

// BuildEraComplianceContext ports archive_trap()'s context assembly for the
// on_archive_subdomain=True, 'compliance'-variant branch.
func BuildEraComplianceContext(year, month, day int, slug string) EraComplianceContext {
	content := GenerateComplianceContent(year, month, day, slug)
	shared := buildSharedEraFields(year, month, day, slug)
	return EraComplianceContext{ComplianceContent: content, sharedEraFields: shared}
}

// EraMinutesContext is the Go equivalent of the merged context dict
// archive_trap() builds for the archives-YYYY.acpwb.com 'minutes' era
// variant.
type EraMinutesContext struct {
	MinutesContent
	sharedEraFields
}

// BuildEraMinutesContext ports archive_trap()'s context assembly for the
// on_archive_subdomain=True, 'minutes'-variant branch.
func BuildEraMinutesContext(year, month, day int, slug string) EraMinutesContext {
	content := GenerateMinutesContent(year, month, day, slug)
	shared := buildSharedEraFields(year, month, day, slug)
	return EraMinutesContext{MinutesContent: content, sharedEraFields: shared}
}
