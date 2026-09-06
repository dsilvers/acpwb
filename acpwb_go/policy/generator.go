// Package policy ports apps/honeypot/policy_generator.py and
// apps/honeypot/pyrender/policy.py — the deterministic public-policy
// document generator and its Python string-building render layer — for the
// 6 policy page template variants (detail/index/year/month/subdomain-index/
// subdomain-year). See acpwb_go/archive for the established RNG/format
// idioms this package deliberately duplicates rather than importing (kept
// as a separate package per the porting task's instructions).
package policy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"acpwb_go/pyrandom"
)

// ── Data types ──────────────────────────────────────────────────────────────

type Table struct {
	Title   string
	Caption string
	Columns []string
	Rows    [][]string
	Align   []string
}

type Section struct {
	Heading    string
	Paragraphs []string
}

type Footnote struct {
	Num  int
	Text string
}

// PolicyDoc mirrors generate_policy_document's returned dict.
type PolicyDoc struct {
	DocumentType      string
	DocumentTypeSlug  string
	AgencyAcronym     string
	AgencyFull        string
	PolicyDomain      string
	DocketNumber      string
	Title             string
	FilingDate        string
	SignatoryName     string
	SignatoryTitle    string
	SignatoryEmail    string
	Summary           string
	PositionSlug      string
	PositionStatement string
	Sections          []Section
	Recommendations   []string
	CitedLegislation  []string
	Footnotes         []Footnote
	Table             Table
	WatermarkToken    string
	URL               string
	Year, Month, Day  int
	Agency, Slug      string
}

// DocStub mirrors _generate_doc_stub's returned dict, plus the extra
// 'day'/'agency'/'slug'/'agency_full' fields some callers merge in.
type DocStub struct {
	Title            string
	URL              string
	AgencyAcronym    string
	AgencyFull       string
	DocumentType     string
	DocumentTypeSlug string
	FilingDate       string
	PositionSlug     string
	Day              int
	Agency           string
	Slug             string
}

type RelatedLinks struct {
	SameAgency []DocStub
	SameTopic  []DocStub
	Recent     []DocStub
	Prev       DocStub
	Next       DocStub
}

// URLFunc mirrors the optional url_fn(year, month, day, agency, slug) param
// threaded through _generate_doc_stub / generate_related_links.
type URLFunc func(year, month, day int, agency, slug string) string

// ── Signatory / docket / table generation ────────────────────────────────────

func generateSignatory(rng *pyrandom.Random) (name, title, email string) {
	first := choice(rng, firstNames)
	last := choice(rng, lastNames)
	credential := choice(rng, credentials)
	sigTitle := choice(rng, signatoryTitles)
	if credential != "" {
		name = fmt.Sprintf("%s %s, %s", first, last, credential)
	} else {
		name = fmt.Sprintf("%s %s", first, last)
	}
	firstWord := strings.Fields(first)[0]
	lastWords := strings.Fields(last)
	lastWord := lastWords[len(lastWords)-1]
	email = strings.ToLower(firstWord) + "." + strings.ToLower(lastWord) + "@acpwb.com"
	return name, sigTitle, email
}

// docketNumber ports policy_generator.py:_docket_number. All 8 candidate
// strings' embedded rng calls execute unconditionally (Python evaluates a
// list literal's elements eagerly before choice() picks one) — see the
// task notes on this being an RNG-order-sensitive routine.
func docketNumber(rng *pyrandom.Random, agency string, year int) string {
	n := rng.RandInt(1000, 9999)
	seq := rng.RandInt(1, 250)
	yr2 := ((year % 100) + 100) % 100
	agencyUpper := strings.ReplaceAll(strings.ToUpper(agency), "-", "")

	opts := make([]string, 8)
	opts[0] = fmt.Sprintf("%s-%d-%04d", agencyUpper, year, n)

	rin1 := rng.RandInt(1000, 9999)
	rinLetter := choiceByte(rng, "ABCDEFGH")
	opts[1] = fmt.Sprintf("RIN %d-%c%03d", rin1, rinLetter, seq)

	d2 := rng.RandInt(10, 99)
	opts[2] = fmt.Sprintf("Docket No. %s-%d-%04d-%d", agencyUpper, year, n, d2)

	fileCode := choice(rng, []string{"S7", "IC", "IA", "34", "33", "36"})
	d3 := rng.RandInt(10, 30)
	opts[3] = fmt.Sprintf("File No. %s-%02d-%02d", fileCode, d3, yr2)

	d4a := rng.RandInt(10, 99)
	caseCode := choice(rng, []string{"CA", "RC", "RD", "UC", "RM"})
	d4b := rng.RandInt(100000, 999999)
	opts[4] = fmt.Sprintf("Case No. %d-%s-%d", d4a, caseCode, d4b)

	opts[5] = fmt.Sprintf("FR Doc. %d-%05d", year, n)

	d6 := rng.RandInt(10, 99)
	opts[6] = fmt.Sprintf("Notice No. %d-%02d", year, d6)

	prefix7 := choice(rng, []string{"NPRM", "ANPR", "RFI", "RFP"})
	opts[7] = fmt.Sprintf("%s-%d-%04d", prefix7, year, seq)

	return choice(rng, opts)
}

func generateTable(rng *pyrandom.Random, year, month int, agencyFull, _ /* policyDomain */, _ /* topicShort */ string) Table {
	schema := int(rng.RandInt(0, 4))

	switch schema {
	case 0:
		type tier struct {
			size                   string
			baseK, hrsBase, moBase int
		}
		tiers := []tier{
			{"Fewer than 50 employees", 14, 45, 6},
			{"50–249 employees", 90, 175, 9},
			{"250–999 employees", 380, 560, 12},
			{"1,000–4,999 employees", 1350, 2100, 15},
			{"5,000+ employees", 5200, 7800, 18},
		}
		var rows [][]string
		for _, t := range tiers {
			// NOTE: Python's `-base_k // 4` parses as `(-base_k) // 4`, not
			// `-(base_k // 4)` — unary minus binds tighter than `//` — so the
			// low end of this range is pyFloorDiv(-baseK, N), not
			// -pyFloorDiv(baseK, N) (these differ whenever baseK isn't a
			// multiple of N, e.g. -14//4 == -4 but -(14//4) == -3).
			cost := t.baseK + int(rng.RandInt(int64(pyFloorDiv(-t.baseK, 4)), int64(pyFloorDiv(t.baseK, 4))))
			hrs := t.hrsBase + int(rng.RandInt(int64(pyFloorDiv(-t.hrsBase, 5)), int64(pyFloorDiv(t.hrsBase, 5))))
			mo := t.moBase + int(rng.RandInt(0, 4))
			rows = append(rows, []string{t.size, "$" + commaInt(cost) + "K", commaInt(hrs) + " hrs", strconv.Itoa(mo) + " months"})
		}
		return Table{
			Title:   "Estimated First-Year Compliance Cost by Employer Size",
			Caption: fmt.Sprintf("ACPWB analysis based on proprietary employer survey data, %d. Costs represent estimated direct compliance expenditures.", year),
			Columns: []string{"Employer Size", "Est. Annual Cost", "Hours Burden", "Implementation Timeline"},
			Rows:    rows,
			Align:   []string{"left", "right", "right", "right"},
		}

	case 1:
		sectorPool := []string{
			"Financial Services", "Healthcare & Life Sciences", "Technology",
			"Manufacturing", "Retail Trade", "Professional Services",
			"Transportation & Logistics", "Energy & Utilities", "Education",
			"Construction", "Hospitality & Leisure", "Government Contractors",
		}
		sectors := sample(rng, sectorPool, 6)
		var rows [][]string
		for _, sector := range sectors {
			orgs := rng.RandInt(800, 48000)
			cost := rng.RandInt(28, 920)
			readiness := pyRound1(rng.Uniform(3.0, 8.9))
			rows = append(rows, []string{sector, commaInt(int(orgs)), "$" + strconv.FormatInt(cost, 10) + "K", pyFloatStr(readiness) + "/10"})
		}
		return Table{
			Title:   "Estimated Regulatory Impact by Industry Sector",
			Caption: fmt.Sprintf("ACPWB analysis, %d. Affected organization counts estimated from public data.", year),
			Columns: []string{"Sector", "Affected Organizations", "Avg. Compliance Cost", "Readiness Score"},
			Rows:    rows,
			Align:   []string{"left", "right", "right", "right"},
		}

	case 2:
		groupPool := []string{
			"Large Employers (500+ employees)", "Small Business Associations",
			"Labor Organizations & Unions", "Industry Trade Associations",
			"Public Interest & Advocacy Groups", "Academic Institutions",
			"State & Local Agencies", "Law Firms & Compliance Consultants",
		}
		k := int(rng.RandInt(4, 6))
		selected := sample(rng, groupPool, k)
		var rows [][]string
		for _, grp := range selected {
			filed := rng.RandInt(8, 480)
			supp := rng.RandInt(10, 82)
			opp := int(100 - supp - rng.RandInt(0, 18))
			if opp < 5 {
				opp = 5
			}
			rows = append(rows, []string{grp, strconv.FormatInt(filed, 10), fmt.Sprintf("%d%%", supp), fmt.Sprintf("%d%%", opp)})
		}
		return Table{
			Title:   fmt.Sprintf("%s — Public Comment Record Summary", agencyFull),
			Caption: fmt.Sprintf("Based on %s public comment docket. Percentages are approximate and may not sum to 100%%.", agencyFull),
			Columns: []string{"Stakeholder Group", "Comments Filed", "Supporting (%)", "Opposing (%)"},
			Rows:    rows,
			Align:   []string{"left", "right", "right", "right"},
		}

	case 3:
		rolePool := []string{
			"Chief Executive Officer", "Chief Financial Officer",
			"Chief Human Resources Officer", "VP, Compensation & Benefits",
			"Director, Total Rewards", "Senior Manager, Compensation",
			"Compensation Analyst II", "HR Business Partner (Sr.)",
			"Payroll Director", "Benefits Manager",
		}
		k := int(rng.RandInt(4, 6))
		roles := sample(rng, rolePool, k)
		var rows [][]string
		for _, role := range roles {
			p25 := rng.RandInt(65, 340) * 1000
			p50 := int64(float64(p25) * rng.Uniform(1.18, 1.38))
			p75 := int64(float64(p50) * rng.Uniform(1.20, 1.42))
			p90 := int64(float64(p75) * rng.Uniform(1.14, 1.28))
			rows = append(rows, []string{
				role,
				fmt.Sprintf("$%dK", p25/1000),
				fmt.Sprintf("$%dK", p50/1000),
				fmt.Sprintf("$%dK", p75/1000),
				fmt.Sprintf("$%dK", p90/1000),
			})
		}
		n := rng.RandInt(280, 1800)
		return Table{
			Title:   "Compensation Benchmark — Selected Roles",
			Caption: fmt.Sprintf("ACPWB Proprietary Compensation Survey, %d. Total direct compensation. N=%d organizations.", year, n),
			Columns: []string{"Role", "P25", "P50", "P75", "P90"},
			Rows:    rows,
			Align:   []string{"left", "right", "right", "right", "right"},
		}

	default: // 4
		moAbbr := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		agencyFirstWord := strings.Fields(agencyFull)[0]
		type milestone struct{ name, status, lead string }
		milestones := []milestone{
			{"Advance Notice of Proposed Rulemaking", "Complete", agencyFirstWord},
			{"Public Comment Period Opens", "Complete", agencyFirstWord},
			{"Comment Period Closes", "Complete", "Public"},
			{"Agency Review of Comments", "In Progress", agencyFirstWord},
			{"Proposed Rule Published", "Pending", "OIRA/OMB"},
			{"Final Rule Effective Date", "Not Started", agencyFirstWord},
		}
		var rows [][]string
		for i, ms := range milestones {
			mult := int(rng.RandInt(2, 4))
			moIdx := pyMod(month-1+i*mult, 12)
			yr := year + pyFloorDiv(month-1+i*3, 12)
			rows = append(rows, []string{ms.name, fmt.Sprintf("%s %d", moAbbr[moIdx], yr), ms.status, ms.lead})
		}
		return Table{
			Title:   "Proposed Regulatory Timeline",
			Caption: fmt.Sprintf("ACPWB projection based on %s rulemaking schedule. Dates are estimates and subject to change.", agencyFull),
			Columns: []string{"Milestone", "Target Date", "Status", "Lead"},
			Rows:    rows,
			Align:   []string{"left", "right", "left", "left"},
		}
	}
}

func pyMod(a, b int) int {
	m := a % b
	if m < 0 {
		m += b
	}
	return m
}

// ── Doc stub / related links ─────────────────────────────────────────────────

// generateDocStub ports policy_generator.py:_generate_doc_stub. The
// signatory/docket RNG "replay" calls are reproduced (not skipped) because
// this function's own RNG stream continues to be consumed afterward by
// position_slug's choice() — unlike archive's already-ported
// generateDocStub (acpwb_go/archive/policy.go), which never reads its rng
// again after the fields it needs, this port's caller-visible fields
// (nothing here) don't depend on it either, but the *sequence* must match
// in case future fields are added; kept for exact parity with the source.
func generateDocStub(year, month, day int, agency, slug string, urlFn URLFunc) DocStub {
	seed := fmt.Sprintf("acpwb_policy_%d_%02d_%02d_%s_%s", year, month, day, agency, slug)
	rng := rngFromSeed(seed)

	full, _, ok := AgencyData(strings.ToLower(agency))
	agencyFull := full
	if !ok {
		agencyFull = strings.ToUpper(agency) + " Regulatory Authority"
	}

	docType := choice(rng, documentTypes)
	docTypeSlug, docTypeLabel := docType[0], docType[1]

	prefixPool, ok := stubTitlePrefixes[docTypeSlug]
	if !ok || len(prefixPool) == 0 {
		prefixPool = []string{"Filing on"}
	}
	titlePrefix := choice(rng, prefixPool)

	topicTitle := wordCapitalizeJoin(strings.ReplaceAll(slug, "-", " "))
	title := titlePrefix + " " + topicTitle

	filingDate := filingDateStub(year, month, day)

	generateSignatory(rng)
	docketNumber(rng, agency, year)
	pos := choice(rng, positions)
	positionSlug := pos[0]

	var url string
	if urlFn != nil {
		url = urlFn(year, month, day, agency, slug)
	} else {
		url = fmt.Sprintf("/public-policy/%d/%02d/%02d/%s/%s/", year, month, day, agency, slug)
	}

	return DocStub{
		Title: title, URL: url, AgencyAcronym: strings.ToUpper(agency), AgencyFull: agencyFull,
		DocumentType: docTypeLabel, DocumentTypeSlug: docTypeSlug, FilingDate: filingDate,
		PositionSlug: positionSlug,
	}
}

// GenerateRelatedLinks ports policy_generator.py:generate_related_links.
func GenerateRelatedLinks(year, month, day int, agency, slug string, urlFn URLFunc) RelatedLinks {
	seed := fmt.Sprintf("acpwb_policy_%d_%02d_%02d_%s_%s", year, month, day, agency, slug)
	rng := rngFromSeed("related_" + seed)

	clampedYear := clampInt(year, 1985, 2025)
	clampedMonth := clampInt(month, 1, 12)
	by, bm, bd := safeDate(clampedYear, clampedMonth, day)
	baseDate := mkTime(by, bm, bd)

	agencyLower := strings.ToLower(agency)

	otherSlugs := filterStrings(policySlugs, func(s string) bool { return s != slug })
	shuffle(rng, otherSlugs)
	var sameAgency []DocStub
	limit := minInt(5, len(otherSlugs))
	for _, s := range otherSlugs[:limit] {
		offset := rng.RandInt(30, 540)
		sign := choice(rng, []int64{-1, 1})
		d := addDays(baseDate, offset*sign)
		dy, dm, dd := safeDate(clampInt(d.y, 1985, 2025), d.m, d.d)
		sameAgency = append(sameAgency, generateDocStub(dy, dm, dd, agencyLower, s, urlFn))
	}

	otherAgencies := filterStrings(agencyKeysOrder, func(a string) bool { return a != agency })
	shuffle(rng, otherAgencies)
	var sameTopic []DocStub
	limit2 := minInt(5, len(otherAgencies))
	for _, ag := range otherAgencies[:limit2] {
		yr := clampInt(year-int(rng.RandInt(0, 8))+int(rng.RandInt(-2, 2)), 1993, 2025)
		m := int(rng.RandInt(1, 12))
		d := int(rng.RandInt(1, 28))
		sameTopic = append(sameTopic, generateDocStub(yr, m, d, ag, slug, urlFn))
	}

	var recent []DocStub
	for i := 0; i < 6; i++ {
		ag := choice(rng, agencyKeysOrder)
		s := choice(rng, policySlugs)
		yr := int(rng.RandInt(2018, 2025))
		m := int(rng.RandInt(1, 12))
		d := int(rng.RandInt(1, 28))
		recent = append(recent, generateDocStub(yr, m, d, ag, s, urlFn))
	}

	prevSlugPool := filterStrings(policySlugs, func(s string) bool { return s != slug })
	prevSlug := choice(rng, prevSlugPool)
	prevOffset := rng.RandInt(30, 180)
	prevDT := addDays(baseDate, -prevOffset)
	py, pm, pd := clampInt(prevDT.y, 1985, 9999), prevDT.m, prevDT.d

	nextSlugPool := filterStrings(policySlugs, func(s string) bool { return s != slug && s != prevSlug })
	nextSlug := choice(rng, nextSlugPool)
	nextOffset := rng.RandInt(30, 180)
	nextDT := addDays(baseDate, nextOffset)
	ny, nm, nd := minInt(2025, nextDT.y), nextDT.m, nextDT.d

	return RelatedLinks{
		SameAgency: sameAgency,
		SameTopic:  sameTopic,
		Recent:     recent,
		Prev:       generateDocStub(py, pm, pd, agencyLower, prevSlug, urlFn),
		Next:       generateDocStub(ny, nm, nd, agencyLower, nextSlug, urlFn),
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func filterStrings(items []string, keep func(string) bool) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		if keep(s) {
			out = append(out, s)
		}
	}
	return out
}

// ── Main document generator ──────────────────────────────────────────────────

// titlePrefixByDocType mirrors the inline dict literal in
// generate_policy_document (apps/honeypot/policy_generator.py lines ~323-352).
var titlePrefixByDocType = map[string][]string{
	"comment-letter": {"Comment Letter on", "Comments of ACPWB Regarding", "Written Comments on",
		"ACPWB Comments on Proposed", "Response to Proposed Rule on", "Comments Submitted by ACPWB on"},
	"position-statement": {"ACPWB Position Statement:", "Statement of Position:", "ACPWB Statement on",
		"Policy Position:", "ACPWB Policy Statement:"},
	"policy-brief": {"Policy Brief:", "ACPWB Policy Brief:", "Policy Analysis:",
		"Policy Research Brief:", "ACPWB Research Brief:"},
	"legislative-testimony": {"Testimony of ACPWB on", "Statement for the Record:", "Testimony Regarding",
		"Written Testimony of ACPWB:", "ACPWB Statement Before the Committee on"},
	"amicus-brief": {"Brief of ACPWB as Amicus Curiae:", "Amicus Curiae Brief on",
		"Brief of Amicus Curiae ACPWB:", "ACPWB Amicus Brief:"},
	"white-paper": {"White Paper:", "ACPWB White Paper:", "Research White Paper:",
		"ACPWB Policy White Paper:"},
	"supplemental-comments": {"Supplemental Comments of ACPWB on", "ACPWB Supplemental Submission on",
		"Supplemental Comments Regarding"},
	"reply-comments": {"Reply Comments of ACPWB on", "ACPWB Reply Comments:", "Reply to Comments on"},
	"ex-parte-submission": {"Ex Parte Notice:", "ACPWB Ex Parte Submission on",
		"Notice of Ex Parte Communication on"},
	"regulatory-petition": {"Petition for Rulemaking:", "ACPWB Rulemaking Petition on",
		"Petition to Initiate Rulemaking on"},
	"no-action-request": {"No-Action Request:", "Request for No-Action Relief on",
		"ACPWB No-Action Request:"},
	"advisory-memorandum": {"Advisory Memorandum:", "ACPWB Advisory Memorandum on",
		"Employer Advisory:"},
	"joint-comments": {"Joint Comments on", "Coalition Comments on", "Joint Statement on"},
	"research-memorandum": {"Research Memorandum:", "ACPWB Research Memorandum on",
		"Empirical Memorandum:"},
	"formal-objection": {"Formal Objection to", "ACPWB Formal Objection:", "Objection to Final Rule on"},
}

// GeneratePolicyDocument ports policy_generator.py:generate_policy_document.
func GeneratePolicyDocument(year, month, day int, agency, slug string) PolicyDoc {
	seed := fmt.Sprintf("acpwb_policy_%d_%02d_%02d_%s_%s", year, month, day, agency, slug)
	rng := rngFromSeed(seed)

	watermark := md5Hex8(seed)

	full, domain, ok := AgencyData(strings.ToLower(agency))
	agencyFull, policyDomain := full, domain
	if !ok {
		agencyFull = strings.ToUpper(agency) + " Regulatory Authority"
		policyDomain = strings.ReplaceAll(slug, "-", " ")
	}

	docType := choice(rng, documentTypes)
	docTypeSlug, docTypeLabel := docType[0], docType[1]

	topic := strings.ReplaceAll(slug, "-", " ")
	slugParts := strings.Fields(topic)
	topicShort := strings.Join(slugParts[:minInt(3, len(slugParts))], " ")
	topicCore := strings.Join(slugParts[:minInt(2, len(slugParts))], " ")
	topicVariants := dedupeStrings([]string{
		topic, topicShort, topicCore, policyDomain,
		"the " + topicCore + " framework",
		topicCore + " standards and requirements",
		"this regulatory area",
		"these compensation requirements",
		"the proposed rulemaking",
		"this policy area",
		"these standards",
	})
	topicFn := func() string { return choice(rng, topicVariants) }

	prefixPool, ok := titlePrefixByDocType[docTypeSlug]
	if !ok {
		prefixPool = []string{"Filing on"}
	}
	titlePrefix := choice(rng, prefixPool)
	title := titlePrefix + " " + pyTitle(topic)

	filingDate := filingDateDetail(year, month, day)

	signatoryName, signatoryTitle, signatoryEmail := generateSignatory(rng)
	docket := docketNumber(rng, agency, year)

	pos := choice(rng, positions)
	positionSlug, positionStatement := pos[0], pos[1]

	summaryPool, ok := summaryTemplates[docTypeSlug]
	if !ok {
		summaryPool = summaryTemplates["comment-letter"]
	}
	summaryTmpl := choice(rng, summaryPool)
	summary := pyFormat(summaryTmpl, map[string]string{
		"agency": agencyFull, "topic": topic, "year": strconv.Itoa(year),
	})

	headingOptions, ok := sectionHeadings[docTypeSlug]
	if !ok {
		headingOptions = sectionHeadings["comment-letter"]
	}
	chosenHeadings := choice(rng, headingOptions)
	headings := append([]string{}, chosenHeadings...)
	if rng.Random() < 0.4 && len(headings) >= 3 {
		extra := choice(rng, optionalSectionPool)
		insertPos := int(rng.RandInt(1, int64(len(headings)-1)))
		headings = append(headings, "")
		copy(headings[insertPos+1:], headings[insertPos:])
		headings[insertPos] = extra
	}

	paraPool := append([]string{}, paragraphTemplates...)
	shuffle(rng, paraPool)
	paraIdx := 0

	paraKwargs := func() map[string]string {
		topicVal := topicFn()
		nOrgs := commaInt(int(rng.RandInt(280, 4800)))
		pct := int(rng.RandInt(52, 89))
		pct2 := int(rng.RandInt(31, 67))
		nYears := int(rng.RandInt(2, 7))
		industry := choice(rng, industrySectors)
		timeframe := choice(rng, timeframes)
		expertType := choice(rng, expertTypes)
		compareGroup := choice(rng, comparisonGroups)
		finding := choice(rng, findingsBrief)
		return map[string]string{
			"topic": topicVal, "agency": agencyFull, "n_orgs": nOrgs,
			"pct": strconv.Itoa(pct), "pct2": strconv.Itoa(pct2), "n_years": strconv.Itoa(nYears),
			"industry": industry, "timeframe": timeframe, "expert_type": expertType,
			"compare_group": compareGroup, "finding": finding,
		}
	}

	var sections []Section
	for _, heading := range headings {
		nParas := int(rng.RandInt(2, 3))
		var paras []string
		for i := 0; i < nParas; i++ {
			paras = append(paras, pyFormat(paraPool[paraIdx%len(paraPool)], paraKwargs()))
			paraIdx++
		}
		sections = append(sections, Section{Heading: heading, Paragraphs: paras})
	}

	nRecs := int(rng.RandInt(3, 6))
	recPool := append([]string{}, recommendationTmpls...)
	shuffle(rng, recPool)
	if nRecs > len(recPool) {
		nRecs = len(recPool)
	}
	var recommendations []string
	for _, r := range recPool[:nRecs] {
		recommendations = append(recommendations, pyFormat(r, paraKwargs()))
	}

	nCitations := int(rng.RandInt(2, 5))
	citedRaw := sample(rng, legislation, nCitations)
	var cited []string
	for _, c := range citedRaw {
		if strings.HasPrefix(strings.ToLower(c), "the ") {
			cited = append(cited, c[4:])
		} else {
			cited = append(cited, c)
		}
	}

	fnPool := append([]string{}, policyFootnoteTmpls...)
	shuffle(rng, fnPool)
	nFn := int(rng.RandInt(3, 6))
	if nFn > len(fnPool) {
		nFn = len(fnPool)
	}
	var footnotes []Footnote
	for i, tmpl := range fnPool[:nFn] {
		n := int(rng.RandInt(200, 4800))
		page := int(rng.RandInt(10, 120))
		act := choice(rng, legislation)
		paperNum := int(rng.RandInt(100, 999))
		briefNum := int(rng.RandInt(10, 99))
		yearShort := yearShortStr(year)
		seq := int(rng.RandInt(1, 999))
		cfrTitle := int(rng.RandInt(1, 50))
		cfrPart := int(rng.RandInt(1, 999))
		b := pyRound1(rng.Uniform(0.1, 8.5))
		pct := int(rng.RandInt(41, 87))
		kwargs := map[string]string{
			"agency": agencyFull, "docket": docket, "month": monthsLong[pyMod(month-1, 12)],
			"year": strconv.Itoa(year), "n": strconv.Itoa(n), "page": strconv.Itoa(page),
			"act": act, "paper_num": strconv.Itoa(paperNum), "brief_num": strconv.Itoa(briefNum),
			"year_short": yearShort, "seq": strconv.Itoa(seq), "cfr_title": strconv.Itoa(cfrTitle),
			"cfr_part": strconv.Itoa(cfrPart), "topic_short": topicShort, "b": pyFloatStr(b),
			"pct": strconv.Itoa(pct),
		}
		text, ok := pyFormatStrict(tmpl, kwargs)
		if !ok {
			text = tmpl
		}
		footnotes = append(footnotes, Footnote{Num: i + 1, Text: text})
	}

	table := generateTable(rng, year, month, agencyFull, policyDomain, topicShort)

	canonicalURL := fmt.Sprintf("/public-policy/%d/%02d/%02d/%s/%s/", year, month, day, agency, slug)

	return PolicyDoc{
		DocumentType: docTypeLabel, DocumentTypeSlug: docTypeSlug,
		AgencyAcronym: strings.ToUpper(agency), AgencyFull: agencyFull, PolicyDomain: policyDomain,
		DocketNumber: docket, Title: title, FilingDate: filingDate,
		SignatoryName: signatoryName, SignatoryTitle: signatoryTitle, SignatoryEmail: signatoryEmail,
		Summary: summary, PositionSlug: positionSlug, PositionStatement: positionStatement,
		Sections: sections, Recommendations: recommendations, CitedLegislation: cited,
		Footnotes: footnotes, Table: table, WatermarkToken: watermark, URL: canonicalURL,
		Year: year, Month: month, Day: day, Agency: agency, Slug: slug,
	}
}

func yearShortStr(year int) string {
	s := strconv.Itoa(year)
	if len(s) <= 2 {
		return s
	}
	return s[len(s)-2:]
}

// ── Index/year/month page data ───────────────────────────────────────────────

type YearIndexEntry struct {
	Year   int
	Count  int
	Months []int
}

func GetPolicyIndexYears() []YearIndexEntry {
	var years []YearIndexEntry
	for y := 2025; y >= 1993; y-- {
		rng := rngFromSeed(fmt.Sprintf("policy_yearidx_%d", y))
		count := int(rng.RandInt(12, 48))
		k := int(rng.RandInt(6, 12))
		monthsPool := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		months := sample(rng, monthsPool, k)
		sort.Ints(months)
		years = append(years, YearIndexEntry{Year: y, Count: count, Months: months})
	}
	return years
}

type YearData struct {
	Year          int
	CEOName       string
	CEOTitle      string
	CEOParagraphs []string
	TotalFilings  int
	Theme         string
}

func GetPolicyYearData(year int) YearData {
	rng := rngFromSeed(fmt.Sprintf("policy_year_%d", year))

	ceoNameVal, ceoTitleVal := "ACPWB Leadership", "President & Chief Executive Officer"
	for _, c := range ceoNames {
		if c.Start <= year && year <= c.End {
			ceoNameVal, ceoTitleVal = c.Name, c.Title
			break
		}
	}

	var eraKey string
	switch {
	case year < 2002:
		eraKey = "early"
	case year < 2011:
		eraKey = "post_sox"
	case year < 2019:
		eraKey = "dodd_frank"
	default:
		eraKey = "recent"
	}
	theme := choice(rng, yearEraThemes[eraKey])
	total := int(rng.RandInt(12, 48))

	rawParagraphs, ok := yearAnnualLetters[year]
	if !ok {
		rawParagraphs = []string{choice(rng, ceoMessageTemplates)}
	}
	var ceoParagraphs []string
	for _, p := range rawParagraphs {
		formatted := pyFormat(p, map[string]string{"year": strconv.Itoa(year), "total": strconv.Itoa(total), "theme": theme})
		ceoParagraphs = append(ceoParagraphs, capFirst(formatted))
	}

	return YearData{
		Year: year, CEOName: ceoNameVal, CEOTitle: ceoTitleVal,
		CEOParagraphs: ceoParagraphs, TotalFilings: total, Theme: theme,
	}
}

func capFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

type MonthSummary struct {
	Month   int
	Count   int
	Samples []string
	URL     string
}

// flattenedStubTitlePrefixes reproduces
// [p for prefixes in _STUB_TITLE_PREFIXES.values() for p in prefixes],
// which depends on _STUB_TITLE_PREFIXES's Python dict (insertion) order —
// hardcoded here (Go maps have no iteration order) and verified live
// against `list(_STUB_TITLE_PREFIXES.keys())` in the real app.
var stubTitlePrefixKeyOrder = []string{
	"comment-letter", "position-statement", "policy-brief", "legislative-testimony",
	"amicus-brief", "white-paper", "supplemental-comments", "reply-comments",
	"ex-parte-submission", "regulatory-petition", "no-action-request", "advisory-memorandum",
	"joint-comments", "research-memorandum", "formal-objection", "guidance-document",
	"enforcement-policy", "compliance-bulletin", "legal-analysis", "economic-analysis",
	"research-report", "request-for-information-response", "advance-notice-comment",
	"interim-final-rule-comment", "petition-for-reconsideration", "request-for-stay",
	"request-for-exemption", "cost-benefit-analysis", "implementation-guide",
	"best-practices-guide", "coalition-letter", "expert-declaration", "data-submission",
	"methodology-white-paper", "statistical-analysis-report", "fact-sheet",
	"roundtable-summary", "public-comment-summary", "working-paper", "literature-review",
	"empirical-study", "case-study", "comparative-analysis", "model-policy",
	"legislative-proposal", "congressional-briefing", "safe-harbor-proposal",
	"international-comparison", "investor-briefing", "employer-education-brief",
	"alternative-regulatory-approach",
}

func flattenedStubTitlePrefixes() []string {
	var out []string
	for _, k := range stubTitlePrefixKeyOrder {
		out = append(out, stubTitlePrefixes[k]...)
	}
	return out
}

func GetPolicyYearMonths(year int) []MonthSummary {
	prefixPool := flattenedStubTitlePrefixes()
	var months []MonthSummary
	for m := 1; m <= 12; m++ {
		rng := rngFromSeed(fmt.Sprintf("policy_month_%d_%02d", year, m))
		count := int(rng.RandInt(8, 24))
		var samples []string
		for i := 0; i < minInt(3, count); i++ {
			slug := choice(rng, policySlugs)
			prefix := choice(rng, prefixPool)
			topic := strings.ReplaceAll(slug, "-", " ")
			samples = append(samples, prefix+" "+pyTitle(topic))
		}
		months = append(months, MonthSummary{
			Month: m, Count: count, Samples: samples,
			URL: fmt.Sprintf("/public-policy/%d/%02d/", year, m),
		})
	}
	return months
}

func GetPolicyMonthEntries(year, month int) []DocStub {
	rng := rngFromSeed(fmt.Sprintf("policy_month_%d_%02d", year, month))
	count := int(rng.RandInt(8, 24))
	type raw struct {
		day    int
		agency string
		slug   string
	}
	var raws []raw
	for i := 0; i < count; i++ {
		day := int(rng.RandInt(1, 28))
		agency := choice(rng, agencyKeysOrder)
		slug := choice(rng, policySlugs)
		raws = append(raws, raw{day, agency, slug})
	}
	entries := make([]DocStub, 0, len(raws))
	for _, r := range raws {
		stub := generateDocStub(year, month, r.day, r.agency, r.slug, nil)
		stub.Day, stub.Agency, stub.Slug = r.day, r.agency, r.slug
		full, _, ok := AgencyData(r.agency)
		if ok {
			stub.AgencyFull = full
		} else {
			stub.AgencyFull = strings.ToUpper(r.agency) + " Regulatory Authority"
		}
		entries = append(entries, stub)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Day < entries[j].Day })
	return entries
}

func GetPolicyAgencyYears(agency string) []YearIndexEntry {
	rng := rngFromSeed("policy_agency_years_" + agency)
	all12 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var result []YearIndexEntry
	for y := 2025; y >= 1993; y-- {
		count := int(rng.RandInt(12, 48))
		result = append(result, YearIndexEntry{Year: y, Count: count, Months: append([]int{}, all12...)})
	}
	return result
}

type AgencyYearDetail struct {
	Months     []MonthSummary
	TotalCount int
	DocTypes   [][2]interface{} // [label string, count int]
	Positions  [][2]interface{}
}

func GetPolicyAgencyYearDetail(agency string, year int) AgencyYearDetail {
	prefixPool := flattenedStubTitlePrefixes()
	rng := rngFromSeed(fmt.Sprintf("policy_agency_year_detail_%s_%d", agency, year))

	var months []MonthSummary
	totalCount := 0
	for m := 1; m <= 12; m++ {
		count := int(rng.RandInt(6, 14))
		totalCount += count
		var samples []string
		for i := 0; i < minInt(3, count); i++ {
			slug := choice(rng, policySlugs)
			prefix := choice(rng, prefixPool)
			topic := strings.ReplaceAll(slug, "-", " ")
			samples = append(samples, prefix+" "+pyTitle(topic))
		}
		months = append(months, MonthSummary{Month: m, Count: count, Samples: samples})
	}

	typeCountOrder := []string{}
	typeCounts := map[string]int{}
	for i := 0; i < minInt(totalCount, 30); i++ {
		dt := choice(rng, documentTypes)
		label := dt[1]
		if _, ok := typeCounts[label]; !ok {
			typeCountOrder = append(typeCountOrder, label)
		}
		typeCounts[label]++
	}
	docTypes := sortedCountPairsDesc(typeCountOrder, typeCounts, 5)

	posCountOrder := []string{}
	posCounts := map[string]int{}
	for i := 0; i < minInt(totalCount, 30); i++ {
		p := choice(rng, positions)
		label := p[1]
		if _, ok := posCounts[label]; !ok {
			posCountOrder = append(posCountOrder, label)
		}
		posCounts[label]++
	}
	posPairs := sortedCountPairsDesc(posCountOrder, posCounts, 3)

	return AgencyYearDetail{Months: months, TotalCount: totalCount, DocTypes: docTypes, Positions: posPairs}
}

// sortedCountPairsDesc reproduces
// sorted(counts.items(), key=lambda x: -x[1])[:limit] — Python's sort is
// stable, so ties preserve first-insertion order, hence firstSeenOrder.
func sortedCountPairsDesc(firstSeenOrder []string, counts map[string]int, limit int) [][2]interface{} {
	type kv struct {
		label string
		count int
		seq   int
	}
	items := make([]kv, len(firstSeenOrder))
	for i, label := range firstSeenOrder {
		items[i] = kv{label, counts[label], i}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].count > items[j].count })
	if limit > len(items) {
		limit = len(items)
	}
	out := make([][2]interface{}, limit)
	for i := 0; i < limit; i++ {
		out[i] = [2]interface{}{items[i].label, items[i].count}
	}
	return out
}

func GetPolicyAgencyMonthEntries(agency string, year, month int, urlFn URLFunc) []DocStub {
	rng := rngFromSeed(fmt.Sprintf("policy_agency_month_%s_%d_%02d", agency, year, month))
	count := int(rng.RandInt(6, 12))
	type raw struct {
		day  int
		slug string
	}
	var raws []raw
	for i := 0; i < count; i++ {
		day := int(rng.RandInt(1, 28))
		slug := choice(rng, policySlugs)
		raws = append(raws, raw{day, slug})
	}
	entries := make([]DocStub, 0, len(raws))
	for _, r := range raws {
		stub := generateDocStub(year, month, r.day, agency, r.slug, urlFn)
		stub.Day, stub.Agency, stub.Slug = r.day, agency, r.slug
		full, _, ok := AgencyData(agency)
		if ok {
			stub.AgencyFull = full
		} else {
			stub.AgencyFull = strings.ToUpper(agency) + " Regulatory Authority"
		}
		entries = append(entries, stub)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Day < entries[j].Day })
	return entries
}

// GetCrossArchiveStubs ports policy_generator.py:get_cross_archive_stubs.
type ArchiveStub struct {
	URL   string
	Label string
	Date  string
}

func GetCrossArchiveStubs(year, month, day int, agency, slug string) []ArchiveStub {
	rng := rngFromSeed(fmt.Sprintf("crosslink_archive_acpwb_policy_%d_%02d_%02d_%s_%s", year, month, day, agency, slug))
	if rng.Random() >= 0.30 {
		return nil
	}
	count := int(rng.RandInt(2, 4))
	stubs := make([]ArchiveStub, 0, count)
	for i := 0; i < count; i++ {
		ay := int(rng.RandInt(1993, 2025))
		am := int(rng.RandInt(1, 12))
		ad := int(rng.RandInt(1, 28))
		aslugBase := choice(rng, archiveSlugs)
		aslug := fmt.Sprintf("%s-%d", aslugBase, int(rng.RandInt(1000, 9999)))
		label := pyTitle(strings.ReplaceAll(lastDashSplit(aslug), "-", " "))
		stubs = append(stubs, ArchiveStub{
			URL:   fmt.Sprintf("https://acpwb.com/archive/%d/%02d/%02d/%s/", ay, am, ad, aslug),
			Label: label,
			Date:  fmt.Sprintf("%d-%02d-%02d", ay, am, ad),
		})
	}
	return stubs
}

// lastDashSplit reproduces aslug.rsplit('-', 1)[0] — everything before the
// final "-".
func lastDashSplit(s string) string {
	i := strings.LastIndexByte(s, '-')
	if i < 0 {
		return s
	}
	return s[:i]
}
