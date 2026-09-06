package archive

import (
	"fmt"
	"regexp"
	"strings"
)

type Paragraph struct {
	Text string
	Ref  string
}

type Finding struct {
	Text string
	Ref  string
}

type MetricRow struct {
	Name     string
	Baseline string
	Current  string
	Delta    string
	Positive bool
}

type EngagementTeamMember struct {
	Name  string
	Title string
	Email string
}

type PercentileRow struct {
	Metric                                 string
	P10, P25, P33, P50, P67, P75, P90, P95 string
}

type Footnote struct {
	Num  int
	Text string
}

type Revision struct {
	Version     string
	Date        string
	Description string
	Author      string
	AuthorEmail string
}

// ArchiveContent is the Go equivalent of _generate_archive_content()'s
// return dict (apps/honeypot/views.py:98).
type ArchiveContent struct {
	Title          string
	BaseTitle      string
	Org            string
	Industry       string
	Phase          string
	Paragraphs     []Paragraph
	Findings       []Finding
	MetricRows     []MetricRow
	RelatedReports []RelatedReport
	RecordID       string
	BulkHexJS      []string
	BulkHexCSS     []string

	EngCode      string
	DocVersion   string
	Distribution string
	PageCount    int
	FileSizeKB   int

	EngagementTeam  []EngagementTeamMember
	ExecBullets     []string
	PeerGroup       string
	PercentileTable []PercentileRow
	Footnotes       []Footnote
	Revisions       []Revision
}

var percentileLabels = []string{
	"10th", "15th", "20th", "25th", "30th", "33rd", "35th",
	"40th", "45th", "50th", "55th", "60th", "65th", "67th",
	"70th", "75th", "80th", "85th", "90th", "95th", "99th",
}

var trailingNumericID = regexp.MustCompile(`-\d{3,}$`)

// GenerateArchiveContent ports apps/honeypot/views.py:_generate_archive_content
// (the 'default' variant's content generator — the @functools.lru_cache
// decorator there is a pure memoization no-op for a byte-identical port and
// isn't reproduced).
func GenerateArchiveContent(year, month, day int, slug string) ArchiveContent {
	rng := rngA(fmt.Sprintf("content_%d%d%d%s", year, month, day, slug))
	org := choice(rng, archiveOrgs)
	industry := choice(rng, archiveIndustries)
	phase := choice(rng, archivePhases)
	dateStr := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	endYear := min(year+int(rng.RandInt(1, 3)), 2024)
	metric := choice(rng, archiveMetricLabels)

	base := func() map[string]string {
		return map[string]string{
			"org": org, "industry": industry, "phase": phase,
			"date": dateStr, "year": itoaCache(year), "endyear": itoaCache(endYear),
			"metric": metric,
		}
	}

	// Paragraphs
	nPara := int(rng.RandInt(5, 7))
	paraTmpls := sample(rng, archiveParaTemplates, nPara)
	var paragraphs []string
	for _, tmpl := range paraTmpls {
		kw := base()
		kw["n"] = itoaCache(int(rng.RandInt(12, 280)))
		kw["regions"] = itoaCache(int(rng.RandInt(3, 47)))
		kw["pct"] = itoaCache(int(rng.RandInt(3, 18)))
		kw["percentile"] = choice(rng, percentileLabels)
		paragraphs = append(paragraphs, pyFormat(tmpl, kw))
	}

	// Findings
	nFind := int(rng.RandInt(3, 5))
	findTmpls := sample(rng, archiveFindingTmpls, nFind)
	var findings []string
	for _, tmpl := range findTmpls {
		kw := base()
		kw["n"] = itoaCache(int(rng.RandInt(12, 280)))
		kw["regions"] = itoaCache(int(rng.RandInt(3, 47)))
		kw["pct"] = itoaCache(int(rng.RandInt(3, 18)))
		kw["percentile"] = choice(rng, percentileLabels)
		findings = append(findings, pyFormat(tmpl, kw))
	}

	// Metrics table
	nMetrics := int(rng.RandInt(6, 8))
	metricNames := sample(rng, archiveMetricNames, nMetrics)
	var metricRows []MetricRow
	for _, name := range metricNames {
		baseline := int(rng.RandInt(20, 980))
		delta := int(rng.RandInt(-18, 42))
		current := baseline + delta
		if current < 0 {
			current = 0
		}
		metricRows = append(metricRows, MetricRow{
			Name: name, Baseline: commaInt(baseline), Current: commaInt(current),
			Delta: signedInt(delta), Positive: delta >= 0,
		})
	}

	// Related reports
	k := 3
	if len(reportCatalog) < k {
		k = len(reportCatalog)
	}
	sampledReports := sample(rng, reportCatalog, k)
	var relatedReports []RelatedReport
	for _, e := range sampledReports {
		relatedReports = append(relatedReports, EnrichReport(e))
	}

	// Title
	tail := slug
	if slug != "" {
		parts := strings.Split(slug, "/")
		tail = parts[len(parts)-1]
	} else {
		tail = fmt.Sprintf("%d-%02d-%02d-archive", year, month, day)
	}
	cleanTail := trailingNumericID.ReplaceAllString(tail, "")
	baseTitle := pyTitle(strings.ReplaceAll(cleanTail, "-", " "))
	prefix := choice(rng, archiveTitlePrefixes)
	title := prefix + " " + baseTitle

	recordID := md5Hex(fmt.Sprintf("archive_%d_%d_%d_%s", year, month, day, slug))[:8]

	// Bulk hex
	bulkHex := make([]string, 350)
	for i := range bulkHex {
		bulkHex[i] = fmt.Sprintf("%016x", rng.GetRandBits64(64))
	}
	bulkHexJS := bulkHex[:200]
	bulkHexCSS := bulkHex[200:350]

	var findingsRich []Finding
	for j, f := range findings {
		findingsRich = append(findingsRich, Finding{Text: f, Ref: bulkHex[50+j]})
	}
	var paragraphsRich []Paragraph
	for j, p := range paragraphs {
		paragraphsRich = append(paragraphsRich, Paragraph{Text: p, Ref: bulkHex[60+j]})
	}

	// Structured metadata
	engCode := fmt.Sprintf("ENG-%d-%s-%d", year, choice(rng, engagementCodes), rng.RandInt(10000, 99999))
	docVersion := choice(rng, archiveDocVersions)
	distribution := choice(rng, distributionClasses)
	pageCount := int(rng.RandInt(28, 214))
	fileSizeKB := pageCount * int(rng.RandInt(38, 92))

	// Engagement team
	teamSize := int(rng.RandInt(4, 6))
	var engagementTeam []EngagementTeamMember
	for i := 0; i < teamSize; i++ {
		fname := choice(rng, firstNames)
		lname := choice(rng, lastNames)
		titleT := choice(rng, consultantTitles)
		email := strings.ToLower(fname) + "." + strings.ToLower(lname) + "@acpwb.com"
		engagementTeam = append(engagementTeam, EngagementTeamMember{
			Name: fname + " " + lname, Title: titleT, Email: email,
		})
	}

	// Exec bullets
	nExec := int(rng.RandInt(4, 6))
	execTmpls := sample(rng, execSummaryBullets, nExec)
	var execBullets []string
	for _, tmpl := range execTmpls {
		n := int(rng.RandInt(12, 280))
		pct := int(rng.RandInt(3, 18))
		total := n + int(rng.RandInt(5, 30))
		perc := choice(rng, percentileLabels)
		// NOTE: intentionally does NOT include "phase" — matches
		// _generate_archive_content's real exec_bullets .format() call,
		// which omits phase= even though some EXEC_SUMMARY_BULLETS
		// templates reference {phase}; those genuinely fall back to raw
		// unformatted text in the real app (see pyFormat's doc comment).
		kw := map[string]string{
			"org": org, "industry": industry, "metric": metric,
			"year": itoaCache(year), "endyear": itoaCache(endYear),
			"date":       dateStr,
			"n":          itoaCache(n),
			"regions":    itoaCache(int(rng.RandInt(3, 47))),
			"pct":        itoaCache(pct),
			"total":      itoaCache(total),
			"percentile": perc,
		}
		execBullets = append(execBullets, pyFormat(tmpl, kw))
	}

	// Benchmark percentile table
	nBench := int(rng.RandInt(4, 6))
	benchNames := sample(rng, benchMetrics, nBench)
	peerGroupTmpl := choice(rng, peerGroups)
	peerGroup := pyFormat(peerGroupTmpl, map[string]string{
		"industry": industry,
		"regions":  itoaCache(int(rng.RandInt(3, 47))),
		"n":        itoaCache(int(rng.RandInt(12, 280))),
	})
	var percentileTable []PercentileRow
	for _, bm := range benchNames {
		b := int(rng.RandInt(45000, 320000))
		percentileTable = append(percentileTable, PercentileRow{
			Metric: bm,
			P10:    dollarComma(int(float64(b) * 0.58)),
			P25:    dollarComma(int(float64(b) * 0.78)),
			P33:    dollarComma(int(float64(b) * 0.88)),
			P50:    dollarComma(b),
			P67:    dollarComma(int(float64(b) * 1.14)),
			P75:    dollarComma(int(float64(b) * 1.28)),
			P90:    dollarComma(int(float64(b) * 1.62)),
			P95:    dollarComma(int(float64(b) * 1.84)),
		})
	}

	// Footnotes
	nFoot := int(rng.RandInt(4, 7))
	footTmpls := sample(rng, archiveFootnoteTmpls, nFoot)
	var footnotes []Footnote
	for i, tmpl := range footTmpls {
		q := int(rng.RandInt(1, 4))
		kw := map[string]string{
			"org": org, "industry": industry, "year": itoaCache(year),
			"endyear": itoaCache(endYear), "date": dateStr, "q": itoaCache(q),
			"n":       itoaCache(int(rng.RandInt(12, 280))),
			"regions": itoaCache(int(rng.RandInt(3, 47))),
		}
		footnotes = append(footnotes, Footnote{Num: i + 1, Text: pyFormat(tmpl, kw)})
	}

	// Revisions
	numRevisions := int(rng.RandInt(3, 5))
	maxStart := len(revisionTypes) - numRevisions
	if maxStart < 0 {
		maxStart = 0
	}
	start := int(rng.RandInt(0, int64(maxStart)))
	revSample := revisionTypes[start : start+numRevisions]
	var revisions []Revision
	for i, pair := range revSample {
		verLabel, rdesc := pair[0], pair[1]
		rMonth := month - (numRevisions - 1 - i)
		if rMonth < 1 {
			rMonth = 1
		}
		if rMonth > 12 {
			rMonth = 12
		}
		rDay := int(rng.RandInt(1, 28))
		rDate := fmt.Sprintf("%d-%02d-%02d", year, rMonth, rDay)
		fname := choice(rng, firstNames)
		lname := choice(rng, lastNames)
		author := fname + " " + lname
		authorEmail := strings.ToLower(fname) + "." + strings.ToLower(lname) + "@acpwb.com"
		q := int(rng.RandInt(1, 4))
		kw := map[string]string{
			"org": org, "date": rDate, "q": itoaCache(q),
			"year": itoaCache(year), "endyear": itoaCache(endYear),
			"pct": itoaCache(int(rng.RandInt(3, 18))),
			"n":   itoaCache(int(rng.RandInt(12, 280))),
		}
		desc := pyFormat(rdesc, kw)
		revisions = append(revisions, Revision{
			Version: verLabel, Date: rDate, Description: desc,
			Author: author, AuthorEmail: authorEmail,
		})
	}
	// stable sort by date, matching Python's list.sort(key=...) (Timsort is stable)
	stableSortRevisionsByDate(revisions)

	return ArchiveContent{
		Title: title, BaseTitle: baseTitle, Org: org, Industry: industry, Phase: phase,
		Paragraphs: paragraphsRich, Findings: findingsRich, MetricRows: metricRows,
		RelatedReports: relatedReports, RecordID: recordID,
		BulkHexJS: bulkHexJS, BulkHexCSS: bulkHexCSS,
		EngCode: engCode, DocVersion: docVersion, Distribution: distribution,
		PageCount: pageCount, FileSizeKB: fileSizeKB,
		EngagementTeam: engagementTeam, ExecBullets: execBullets,
		PeerGroup: peerGroup, PercentileTable: percentileTable,
		Footnotes: footnotes, Revisions: revisions,
	}
}

func stableSortRevisionsByDate(rs []Revision) {
	// insertion sort: stable, fine for the tiny (3-5 element) slices involved.
	for i := 1; i < len(rs); i++ {
		j := i
		for j > 0 && rs[j-1].Date > rs[j].Date {
			rs[j-1], rs[j] = rs[j], rs[j-1]
			j--
		}
	}
}

// itoaCache avoids importing strconv everywhere in this file's call sites.
func itoaCache(n int) string {
	return fmt.Sprintf("%d", n)
}

// ── Nav / related-content helpers (views.py) ────────────────────────────────

// GenNavSlugs ports views.py:_gen_nav_slugs.
func GenNavSlugs(year, month, day int, slug string) (nextSlug, prevSlug string) {
	rng := rngA(fmt.Sprintf("navslugs_%d%d%d%s", year, month, day, slug))
	if slug != "" {
		nextSlug = fmt.Sprintf("%s/%s-%d", slug, choice(rng, archiveSlugs), rng.RandInt(1000, 9999))
	} else {
		nextSlug = fmt.Sprintf("%s-%d", choice(rng, archiveSlugs), rng.RandInt(1000, 9999))
	}
	prevSlug = fmt.Sprintf("%s-%d", choice(rng, archiveSlugs), rng.RandInt(1000, 9999))
	return
}

type RelatedPathRaw struct {
	Year, Month, Day int
	FullSlug         string
	Label            string
	Date             string
}

// GenRelatedPathData ports views.py:_gen_related_path_data.
func GenRelatedPathData(year, month, day int, slug string) []RelatedPathRaw {
	rng := rngA(fmt.Sprintf("relpaths_%d%d%d%s", year, month, day, slug))
	out := make([]RelatedPathRaw, 0, 10)
	for i := 0; i < 10; i++ {
		rYear := int(rng.RandInt(1985, 2025))
		rMonth := int(rng.RandInt(1, 12))
		rDay := int(rng.RandInt(1, 28))
		rSlug := choice(rng, archiveSlugs)
		rID := int(rng.RandInt(1000, 9999))
		out = append(out, RelatedPathRaw{
			Year: rYear, Month: rMonth, Day: rDay,
			FullSlug: fmt.Sprintf("%s-%d", rSlug, rID),
			Label:    pyTitle(strings.ReplaceAll(rSlug, "-", " ")),
			Date:     fmt.Sprintf("%d-%02d-%02d", rYear, rMonth, rDay),
		})
	}
	return out
}

type CrossYearReport struct {
	URL   string
	Label string
	Date  string
	Year  int
}

func crossYearArchiveURL(year, month, day int, slug string) string {
	return fmt.Sprintf("https://archives-%d.acpwb.com/%02d/%02d/%s/", year, month, day, slug)
}

// GenCrossYearReports ports views.py:_gen_cross_year_reports.
func GenCrossYearReports(year, month, day int, slug string) []CrossYearReport {
	rng := rngA(fmt.Sprintf("crossyear_%d%d%d%s", year, month, day, slug))
	n := int(rng.RandInt(1, 5))
	out := make([]CrossYearReport, 0, n)
	for i := 0; i < n; i++ {
		cyYear := int(rng.RandInt(1985, 2025))
		for cyYear == year {
			cyYear = int(rng.RandInt(1985, 2025))
		}
		cyMonth := int(rng.RandInt(1, 12))
		cyDay := int(rng.RandInt(1, 28))
		cySlug := choice(rng, archiveSlugs)
		cyID := int(rng.RandInt(1000, 9999))
		out = append(out, CrossYearReport{
			URL:   crossYearArchiveURL(cyYear, cyMonth, cyDay, fmt.Sprintf("%s-%d", cySlug, cyID)),
			Label: pyTitle(strings.ReplaceAll(cySlug, "-", " ")),
			Date:  fmt.Sprintf("%d-%02d-%02d", cyYear, cyMonth, cyDay),
			Year:  cyYear,
		})
	}
	return out
}

type RelatedDocRaw struct {
	Label    string
	Day      int
	FullSlug string
	Date     string
	Phase    string
}

// GenRelatedDocsData ports views.py:_gen_related_docs_data.
func GenRelatedDocsData(year, month, day int, slug string) []RelatedDocRaw {
	rng := rngA(fmt.Sprintf("reldocs_%d%d%d%s", year, month, day, slug))
	n := int(rng.RandInt(2, 4))
	out := make([]RelatedDocRaw, 0, n)
	for i := 0; i < n; i++ {
		sibSlug := fmt.Sprintf("%s-%d", choice(rng, archiveSlugs), rng.RandInt(1000, 9999))
		sibDay := int(rng.RandInt(1, 28))
		sibPrefix := choice(rng, archiveTitlePrefixes)
		parts := strings.Split(sibSlug, "-")
		sibBase := pyTitle(strings.ReplaceAll(strings.Join(parts[:len(parts)-1], "-"), "-", " "))
		out = append(out, RelatedDocRaw{
			Label:    sibPrefix + " " + sibBase,
			Day:      sibDay,
			FullSlug: sibSlug,
			Date:     fmt.Sprintf("%d-%02d-%02d", year, month, sibDay),
			Phase:    choice(rng, archivePhases),
		})
	}
	return out
}

// GenPresentationsCount ports views.py:_gen_presentations_count.
func GenPresentationsCount(year, month, day int, slug string) int {
	rng := rngA(fmt.Sprintf("prescount_%d%d%d%s", year, month, day, slug))
	return int(choice(rng, []int64{2, 3, 4}))
}

// ArchiveURL ports the main-domain-only branch of views.py:_archive_url
// (on_archive_subdomain is always false in scope for this port).
func ArchiveURL(year int, month, day *int, slug string) string {
	if month == nil {
		return fmt.Sprintf("/archive/%d/", year)
	}
	if day == nil {
		return fmt.Sprintf("/archive/%d/%02d/", year, *month)
	}
	if slug != "" {
		return fmt.Sprintf("/archive/%d/%02d/%02d/%s/", year, *month, *day, slug)
	}
	return fmt.Sprintf("/archive/%d/%02d/%02d/", year, *month, *day)
}

func intPtr(i int) *int { return &i }
