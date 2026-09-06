// Package archive is a Go port of the main-domain ("default" variant, i.e.
// non-subdomain, non-compliance, non-minutes) archive detail page at
// /archive/<year>/<month>/<day>/<slug>/ from the Django project at
// /Users/dan/Projects/acpwb/acpwb.
//
// It reproduces, byte-for-byte:
//   - apps/honeypot/views.py: _generate_archive_content, _gen_nav_slugs,
//     _gen_related_path_data, _gen_cross_year_reports, _gen_related_docs_data,
//     _gen_presentations_count, _archive_url (main-domain branch only)
//   - apps/honeypot/report_generator.py: _enrich_report (+ REPORT_CATALOG)
//   - apps/honeypot/policy_generator.py: get_cross_policy_stubs (+ helpers)
//   - apps/presentations/generators.py: generate_presentations_for_context
//   - apps/presentations/image_selector.py: pick_background
//   - apps/presentations/logo_generator.py: generate_org_logo
//   - apps/core/templatetags/acpwb_tags.py: headshot_or_avatar, avatar_card
//   - apps/core/htmlgen.py: escape, truncatewords, truncatechars,
//     get_archive_seal, render_pres_card
//   - apps/honeypot/pyrender/archive_main.py: render_archive_default (+
//     private helpers)
//
// Deliberately NOT ported (out of scope — see the port's task notes):
// archives-YYYY.acpwb.com subdomain rendering (archive_era.py), the
// 'compliance' and 'minutes' archive variants, and the base.html page shell
// (nav/footer/ghost-links/jsonld-garbage/prompt-injection) that wraps
// render_archive_default's output in the real site — this package reproduces
// exactly the render_archive_default(ctx) content fragment, matching the
// verification approach used for its fixtures.
package archive

import (
	"acpwb_go/data"
)

// Archive content pools (apps/honeypot/archive_data.py via export_render_data.py)
var (
	archiveOrgs          = data.Strings("ARCHIVE_ORGS")
	archiveIndustries    = data.Strings("ARCHIVE_INDUSTRIES")
	archivePhases        = data.Strings("ARCHIVE_PHASES")
	archiveParaTemplates = data.Strings("ARCHIVE_PARA_TEMPLATES")
	archiveMetricNames   = data.Strings("ARCHIVE_METRIC_NAMES")
	archiveFindingTmpls  = data.Strings("ARCHIVE_FINDING_TEMPLATES")
	archiveMetricLabels  = data.Strings("ARCHIVE_METRIC_LABELS")
	archiveTitlePrefixes = data.Strings("ARCHIVE_TITLE_PREFIXES")
	archiveSlugs         = data.Strings("ARCHIVE_SLUGS")
	consultantTitles     = data.Strings("CONSULTANT_TITLES")
	execSummaryBullets   = data.Strings("EXEC_SUMMARY_BULLETS")
	archiveFootnoteTmpls = data.Strings("ARCHIVE_FOOTNOTE_TEMPLATES")
	revisionTypes        = data.TuplePairs("REVISION_TYPES")
	distributionClasses  = data.Strings("DISTRIBUTION_CLASSES")
	engagementCodes      = data.Strings("ENGAGEMENT_CODES")
	benchMetrics         = data.Strings("BENCH_METRICS")
	peerGroups           = data.Strings("PEER_GROUPS")
	archiveDocVersions   = data.Strings("ARCHIVE_DOC_VERSIONS")
	firstNames           = data.Strings("FIRST_NAMES")
	lastNames            = data.Strings("LAST_NAMES")
)

// Report pools (apps/honeypot/report_generator.py via export_render_data2.py)
type reportCatalogEntry struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Category string `json:"category"`
	FileType string `json:"file_type"`
}

var (
	reportCatalog        []reportCatalogEntry
	reportYearPool       []int
	reportSummaryTmpls   = data.Strings("REPORT_SUMMARY_TEMPLATES")
	reportFindingPhrases = data.Strings("REPORT_FINDING_PHRASES")
	reportAdjectives     = data.Strings("REPORT_ADJECTIVES")
	reportSubjects       = data.Strings("REPORT_SUBJECTS")
	reportSuffixes       = data.Strings("REPORT_SUFFIXES")
)

func init() {
	data.Unmarshal("REPORT_CATALOG", &reportCatalog)
	reportYearPool = data.Ints("REPORT_YEAR_POOL")
}

// Policy pools (apps/honeypot/policy_data.py via export_render_data.py)
var (
	agencies          = data.TupleMap("AGENCIES")
	documentTypes     = data.TuplePairs("DOCUMENT_TYPES")
	signatoryTitles   = data.Strings("SIGNATORY_TITLES")
	credentials       = data.Strings("CREDENTIALS")
	positions         = data.TuplePairs("POSITIONS")
	stubTitlePrefixes = data.StringSliceMap("STUB_TITLE_PREFIXES")
	policySlugs       = data.Strings("POLICY_SLUGS")
)

// Presentation pools (apps/presentations/data/* via export_render_data2.py)
type presTheme struct {
	Name        string `json:"name"`
	Bg          string `json:"bg"`
	Surface     string `json:"surface"`
	Accent      string `json:"accent"`
	Text        string `json:"text"`
	TextMuted   string `json:"text_muted"`
	HeadingFont string `json:"heading_font"`
	BodyFont    string `json:"body_font"`
}

var (
	presVerbs          = data.Strings("PRES_VERBS")
	presNouns          = data.Strings("PRES_NOUNS")
	presAdjectives     = data.Strings("PRES_ADJECTIVES")
	presTitleTemplates = data.Strings("PRES_TITLE_TEMPLATES")
	presDomains        = data.Strings("PRES_DOMAINS")
	presSubtitles      = data.Strings("PRES_SUBTITLES")
	presVenues         = data.TuplePairs("PRES_VENUES")
	presOrganizations  = data.Strings("PRES_ORGANIZATIONS")
	presOrgSlugMap     = data.StringMap("PRES_ORG_SLUG_MAP")
	presAcronyms       = data.StringMap("PRES_ACRONYMS")
	presTitleCaseLower map[string]bool
	presThemes         []presTheme
	peopleTitles       = data.Strings("PEOPLE_TITLES")
	peopleDepartments  = data.Strings("PEOPLE_DEPARTMENTS")
)

func init() {
	data.Unmarshal("PRES_THEMES", &presThemes)
	lower := data.Strings("PRES_TITLE_CASE_LOWER")
	presTitleCaseLower = make(map[string]bool, len(lower))
	for _, w := range lower {
		presTitleCaseLower[w] = true
	}
}

var archiveSealTemplate = data.Text("ARCHIVE_SEAL_TEMPLATE.html")
