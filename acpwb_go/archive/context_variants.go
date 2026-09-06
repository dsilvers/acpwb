package archive

// sharedPageFields holds the page-level context fields archive_trap()
// (apps/honeypot/views.py:976) assembles identically regardless of which of
// the 3 variants (default/compliance/minutes) got picked — see BuildContext
// in context.go for the original (default-variant) port of this same
// assembly. It's factored out here so the compliance/minutes variants reuse
// it verbatim instead of re-deriving it.
type sharedPageFields struct {
	Year, Month, Day int
	Slug             string

	YearURL      string
	MonthURL     string
	PrevEntryURL string
	NextEntryURL string
	ExportCSVURL string

	RelatedPaths     []RelatedPathView
	CrossYearReports []CrossYearReport
	ArchiveYears     []int
	RelatedDocs      []RelatedDocView
	RelatedPolicy    []PolicyStub
	// Deliberately NOT included: RelatedPresentations. archive_trap() always
	// computes it, but render_compliance_default/render_minutes_default
	// never read related_presentations from the context (confirmed by
	// reading pyrender/archive_main.py in full) — including it here would be
	// dead weight for these two variants' render output.
}

// buildSharedPageFields ports the shared (variant-independent) portion of
// archive_trap()'s context assembly — identical to what BuildContext (in
// context.go) does for the default variant, minus the presentations call,
// which the compliance/minutes templates never reference.
func buildSharedPageFields(year, month, day int, slug string) sharedPageFields {
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

	return sharedPageFields{
		Year: year, Month: month, Day: day, Slug: slug,
		YearURL:      ArchiveURL(year, nil, nil, ""),
		MonthURL:     ArchiveURL(year, intPtr(month), nil, ""),
		PrevEntryURL: ArchiveURL(prevYear, intPtr(prevMonth), intPtr(prevDay), prevSlug),
		NextEntryURL: ArchiveURL(year, intPtr(month), intPtr(day), nextSlug),
		ExportCSVURL: ArchiveURL(year, intPtr(month), intPtr(day), slug) + "export.csv",

		RelatedPaths:     relatedPaths,
		CrossYearReports: crossYearReports,
		ArchiveYears:     archiveYearsDesc(),
		RelatedDocs:      relatedDocs,
		RelatedPolicy:    relatedPolicy,
	}
}

// ComplianceContext is the Go equivalent of the merged context dict
// archive_trap() builds when the 'compliance' variant is selected
// (_variant_int < 3).
type ComplianceContext struct {
	ComplianceContent
	sharedPageFields
	AllYears []int
}

// BuildComplianceContext ports archive_trap()'s context assembly for the
// 'compliance' variant branch (main-domain only).
func BuildComplianceContext(year, month, day int, slug string) ComplianceContext {
	content := GenerateComplianceContent(year, month, day, slug)
	shared := buildSharedPageFields(year, month, day, slug)
	return ComplianceContext{
		ComplianceContent: content,
		sharedPageFields:  shared,
		AllYears:          archiveYearsDesc(),
	}
}

// MinutesContext is the Go equivalent of the merged context dict
// archive_trap() builds when the 'minutes' variant is selected
// (3 <= _variant_int < 6).
type MinutesContext struct {
	MinutesContent
	sharedPageFields
	AllYears []int
}

// BuildMinutesContext ports archive_trap()'s context assembly for the
// 'minutes' variant branch (main-domain only).
func BuildMinutesContext(year, month, day int, slug string) MinutesContext {
	content := GenerateMinutesContent(year, month, day, slug)
	shared := buildSharedPageFields(year, month, day, slug)
	return MinutesContext{
		MinutesContent:   content,
		sharedPageFields: shared,
		AllYears:         archiveYearsDesc(),
	}
}
