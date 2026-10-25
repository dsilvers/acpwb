package compliance

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	sitedata "acpwb_go/data"
)

// Compliance-only content lives in data/ (course types, question banks,
// slot vocabularies) as plain text so it can be extended without Go
// changes; everything shared with the rest of the site comes from the
// embedded Python data pools in acpwb_go/data.
//
//go:embed data
var dataFS embed.FS

type mixEntry struct {
	cat string
	w   int
}

type courseType struct {
	Slug, Title, Prefix string
	MinQ, MaxQ, Pass    int
	Mix                 []mixEntry
	Blurb               string
}

type bankQ struct {
	Text         string
	Right, Wrong []string
}

type perchFact struct {
	Year      int    `json:"year"`
	Theme     string `json:"theme"`
	Dates     string `json:"dates"`
	Venue     string `json:"venue"`
	Attendees int    `json:"attendees"`
	Cancelled bool   `json:"cancelled"`
	Virtual   bool   `json:"virtual"`
}

type reportEntry struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// generatedKinds are question categories built in code (question.go) rather
// than read from a bank file.
var generatedKinds = map[string]bool{
	"agency": true, "statute": true, "wrongfact": true, "scenario": true,
	"perch": true, "report": true, "chart": true,
}

var (
	courseTypes   []*courseType
	courseTypeMap = map[string]*courseType{}
	banks         = map[string][]bankQ{}
	vars          = map[string][]string{}
	varNames      []string
	reportCovers  []string
	topicCovers   = map[string][]string{} // course-type slug → generated cover names

	agencies    []agency
	agencyMap   = map[string]int{}
	orgs        []string
	legislation []string
	frameworks  []string
	findings    []string
	correctives []string
	firstNames  []string
	lastNames   []string
	departments []string
	titles      []string
	wrongFacts  []string
	committees  []string
	perchFacts  []perchFact
	reports     []reportEntry
)

func init() {
	loadCourseTypes()
	loadBanks()
	loadVars()
	loadBodies()
	loadFragments()
	loadKinds()
	reportCovers = lines(readData("data/report_covers.txt"))
	// Topic covers from tools/generate_compliance_covers.py ("<slug>-<n>").
	for _, name := range lines(readData("data/compliance_covers.txt")) {
		if i := strings.LastIndex(name, "-"); i > 0 {
			topicCovers[name[:i]] = append(topicCovers[name[:i]], name)
		}
	}

	var raw map[string]struct {
		T []string `json:"__tuple__"`
	}
	sitedata.Unmarshal("AGENCIES", &raw)
	for _, k := range sitedata.Strings("AGENCY_KEYS_ORDER") {
		a, ok := raw[k]
		if !ok || len(a.T) < 2 {
			continue
		}
		agencyMap[k] = len(agencies)
		agencies = append(agencies, agency{Slug: k, Name: a.T[0], Domain: a.T[1], Code: strings.ToUpper(k)})
	}
	orgs = dedupe(sitedata.Strings("ARCHIVE_ORGS"))
	legislation = sitedata.Strings("LEGISLATION")
	frameworks = sitedata.Strings("COMPLIANCE_FRAMEWORKS")
	findings = sitedata.Strings("COMPLIANCE_FINDING_TYPES")
	correctives = sitedata.Strings("CORRECTIVE_ACTION_TEMPLATES")
	firstNames = sitedata.Strings("FIRST_NAMES")
	lastNames = sitedata.Strings("LAST_NAMES")
	departments = sitedata.Strings("PEOPLE_DEPARTMENTS")
	titles = sitedata.Strings("PEOPLE_TITLES")
	wrongFacts = sitedata.Strings("WRONG_FACTS")
	committees = sitedata.Strings("COMMITTEE_NAMES")
	sitedata.Unmarshal("PERCH_FACTS", &perchFacts)
	sitedata.Unmarshal("REPORT_CATALOG", &reports)
}

func readData(name string) string {
	b, err := dataFS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("compliance: missing %s: %v", name, err))
	}
	return string(b)
}

// lines returns the non-blank, non-comment lines of s, trimmed.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

func dedupe(s []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func loadCourseTypes() {
	for _, l := range lines(readData("data/course_types.txt")) {
		f := strings.Split(l, "|")
		if len(f) != 7 {
			panic("compliance: bad course type line: " + l)
		}
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		ct := &courseType{Slug: f[0], Title: f[1], Prefix: f[2], Blurb: f[6]}
		lo, hi, _ := strings.Cut(f[3], "-")
		ct.MinQ, _ = strconv.Atoi(lo)
		ct.MaxQ, _ = strconv.Atoi(hi)
		ct.Pass, _ = strconv.Atoi(f[4])
		for _, m := range strings.Split(f[5], ",") {
			cat, w, _ := strings.Cut(strings.TrimSpace(m), ":")
			n, _ := strconv.Atoi(w)
			ct.Mix = append(ct.Mix, mixEntry{cat, n})
		}
		if ct.MinQ < 1 || ct.MaxQ < ct.MinQ || ct.MaxQ > maxQuestions || ct.Pass < 1 || ct.Pass > 100 {
			panic("compliance: bad course type numbers: " + l)
		}
		courseTypes = append(courseTypes, ct)
		courseTypeMap[ct.Slug] = ct
	}
}

func loadBanks() {
	entries, _ := dataFS.ReadDir("data/questions")
	for _, e := range entries {
		cat := strings.TrimSuffix(e.Name(), ".txt")
		var qs []bankQ
		var cur *bankQ
		for _, l := range strings.Split(readData(path.Join("data/questions", e.Name())), "\n") {
			l = strings.TrimSpace(l)
			switch {
			case strings.HasPrefix(l, "Q:"):
				qs = append(qs, bankQ{Text: strings.TrimSpace(l[2:])})
				cur = &qs[len(qs)-1]
			case cur != nil && strings.HasPrefix(l, "+"):
				cur.Right = append(cur.Right, strings.TrimSpace(l[1:]))
			case cur != nil && strings.HasPrefix(l, "-"):
				cur.Wrong = append(cur.Wrong, strings.TrimSpace(l[1:]))
			}
		}
		banks[cat] = qs
	}
}

func loadVars() {
	entries, _ := dataFS.ReadDir("data/vars")
	for _, e := range entries {
		vars[strings.TrimSuffix(e.Name(), ".txt")] = lines(readData(path.Join("data/vars", e.Name())))
	}
	for k := range vars {
		varNames = append(varNames, k)
	}
	sort.Strings(varNames)
}

// bankNames lists the question-bank categories, sorted (tests, topic list).
func bankNames() []string {
	var out []string
	for k := range banks {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
