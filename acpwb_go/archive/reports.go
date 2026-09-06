package archive

import (
	"fmt"
	"regexp"
	"strings"

	"acpwb_go/pyrandom"
)

// RelatedReport carries exactly the fields
// pyrender/archive_main.py:_related_reports_html reads from
// _enrich_report()'s return dict (apps/honeypot/report_generator.py).
type RelatedReport struct {
	Category       string
	FileType       string
	PubDateDisplay string
	DetailURL      string
	Title          string
	Summary        string
}

var reportYearRe = regexp.MustCompile(`(\d{4})`)

var monthAbbrev = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// pubDateFor ports report_generator.py:_pub_date_for. Returns (year, month, day).
func pubDateFor(slug string) (int, int, int) {
	rng := rngB("pubdate_" + slug)
	if m := reportYearRe.FindString(slug); m != "" {
		var year int
		fmt.Sscanf(m, "%d", &year)
		if year >= 1990 && year <= 2025 {
			month := int(rng.RandInt(1, 12))
			if strings.Contains(slug, "annual") || strings.Contains(slug, "year") || strings.Contains(slug, "survey") {
				month = int(choice(rng, []int64{9, 10, 11, 12}))
			}
			day := int(rng.RandInt(1, 28))
			return year, month, day
		}
	}
	year := int(choice(rng, toInt64Slice(reportYearPool)))
	month := int(rng.RandInt(1, 12))
	day := int(rng.RandInt(1, 28))
	return year, month, day
}

func toInt64Slice(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}

// summaryFor ports report_generator.py:_summary_for, consuming rng in the
// exact same order as the Python (template choice, finding phrase +
// its {p} fill, then the summary template's own kwargs in the order
// they're written in the .format(...) call).
func summaryFor(rng *pyrandom.Random, slug string) string {
	template := choice(rng, reportSummaryTmpls)
	findingTmpl := choice(rng, reportFindingPhrases)
	findingP := int(rng.RandInt(3, 22))
	finding := pyFormat(findingTmpl, map[string]string{"p": itoaCache(findingP)})

	n := int(rng.RandInt(120, 2800))
	states := int(rng.RandInt(22, 50))
	adj := strings.ToLower(choice(rng, reportAdjectives))
	subject := strings.ToLower(choice(rng, reportSubjects))
	suffix := strings.ToLower(choice(rng, reportSuffixesPool()))
	years := int(rng.RandInt(3, 18))
	year := int(rng.RandInt(2010, 2024))
	q1 := int(rng.RandInt(1, 2))
	q2 := int(rng.RandInt(3, 4))
	yearFrom := int(rng.RandInt(2015, 2021))
	yearTo := int(rng.RandInt(2022, 2024))
	p := int(rng.RandInt(3, 22))

	kw := map[string]string{
		"n": itoaCache(n), "states": itoaCache(states), "adj": adj,
		"subject": subject, "suffix": suffix, "finding": finding,
		"years": itoaCache(years), "year": itoaCache(year),
		"q1": itoaCache(q1), "q2": itoaCache(q2),
		"year_from": itoaCache(yearFrom), "year_to": itoaCache(yearTo),
		"p": itoaCache(p),
	}
	return pyFormat(template, kw)
}

func reportSuffixesPool() []string {
	return reportSuffixes
}

// EnrichReport ports report_generator.py:_enrich_report for the fields the
// default archive variant's _related_reports_html actually renders
// (category, file_type, pub_date_display, detail_url, title, summary).
// fake_size/row_count are intentionally not computed: they're drawn from the
// same rng AFTER summary in the Python source, so skipping them doesn't
// change summary's value and they're never rendered by this page.
func EnrichReport(e reportCatalogEntry) RelatedReport {
	rng := rngB("enrich_" + e.Slug)
	year, month, _ := pubDateFor(e.Slug)
	summary := summaryFor(rng, e.Slug)
	return RelatedReport{
		Category:       e.Category,
		FileType:       e.FileType,
		PubDateDisplay: fmt.Sprintf("%s %d", monthAbbrev[month-1], year),
		DetailURL:      "/reports/" + e.Slug + "/",
		Title:          e.Title,
		Summary:        truncatechars(summary, 160),
	}
}
