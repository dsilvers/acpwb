package compliance

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Question is one generated assessment item. The answer key (Correct) is
// derived from the same seed as the text, so a score is a pure function of
// (course, attempt, answer letters).
type Question struct {
	N       int
	Kind    string // bank category or generated kind
	Prompt  string // "Select the BEST answer.", "True or False", …
	Text    string
	Choices []string
	Correct int
	Exhibit string // inline SVG, or ""
	ExLabel string // "Exhibit 4-B"
	Ref     string // policy/framework reference line

	key string // identifies the underlying item, to avoid repeats in an attempt
}

const letters = "abcde"

func (q Question) Letter(i int) string { return letters[i : i+1] }

// questionCache holds the question sets of recent attempts (a crawler walks
// every answer path of one attempt, re-deriving the same N questions).
var questionCache = newCache[[]Question](4096)

// Questions returns all N questions of an attempt.
func Questions(c *Course, attempt int) []Question {
	key := fmt.Sprintf("%s/%d", c.ID, attempt)
	if qs, ok := questionCache.get(key); ok {
		return qs
	}
	qs := make([]Question, c.N)
	used := map[string]bool{}
	for n := 1; n <= c.N; n++ {
		// Redraw (up to 12 times) rather than repeat an item within one attempt.
		for try := 0; ; try++ {
			qs[n-1] = genQuestion(c, attempt, n, try)
			if !used[qs[n-1].key] || try == 12 {
				break
			}
		}
		used[qs[n-1].key] = true
	}
	questionCache.put(key, qs)
	return qs
}

func genQuestion(c *Course, attempt, n, try int) Question {
	seed := fmt.Sprintf("compliance_q_%s_%d_%d", c.ID, attempt, n)
	if try > 0 {
		seed += fmt.Sprintf("_%d", try)
	}
	r := newRnd(seed)
	mix := c.Type.Mix
	// Life-safety courses stay on topic: only a few questions come from
	// the shared pool.
	universal := 0.4
	if len(mix) > 0 && lifeSafety[mix[0].cat] {
		universal = 0.1
	}
	if r.chance(universal) {
		mix = universalMix
	}
	weights := make([]int, len(mix))
	for i, m := range mix {
		weights[i] = m.w
	}
	cat := mix[weighted(r, weights)].cat
	sl := newSlots(r, c)
	var q Question
	switch {
	case cat == "badge" && r.chance(0.4):
		q = badgeExhibitQ(r, c, sl, n)
	case cat == "exhibit":
		q = exhibitQ(r, c, sl, n, anyExhibit(r))
	case strings.HasPrefix(cat, "ex-") && exhibitGens[cat[3:]] != nil:
		q = exhibitQ(r, c, sl, n, cat[3:])
	case len(categoryExhibits[cat]) > 0 && r.chance(exhibitChance):
		q = exhibitQ(r, c, sl, n, pick(r, categoryExhibits[cat]))
	case cat == "agency":
		q = agencyQ(r)
	case cat == "statute":
		q = statuteQ(r, c)
	case cat == "wrongfact":
		q = wrongFactQ(r)
	case cat == "scenario":
		q = scenarioQ(r, c, sl)
	case cat == "perch":
		q = perchQ(r)
	case cat == "report":
		q = reportQ(r, sl)
	case cat == "chart":
		q = chartQ(r, c, n)
	case cat == "definition":
		q = definitionQ(r)
	case cat == "requirement" && len(c.Requirements) > 0:
		q = requirementQ(r, c)
	case cat == "coursefact" || cat == "requirement":
		q = courseFactQ(r, c)
	default:
		q = bankQuestion(r, cat, sl)
	}
	q.N, q.Kind = n, cat
	if q.Prompt == "" {
		q.Prompt = pick(r, []string{"Select the BEST answer.", "Select ONE answer.", "Choose the most appropriate response.", "Select the correct answer."})
	}
	if q.Ref == "" {
		q.Ref = fmt.Sprintf("Ref: %s, Module %d; %s", c.Code, r.between(1, len(c.Modules)), c.Framework)
	}
	return q
}

// lifeSafety banks: a course whose mix leads with one of these draws few
// shared-pool questions. Their "+" answers are real guidance (see CLAUDE.md).
var lifeSafety = map[string]bool{
	"cpr": true, "firstaid": true, "threat": true, "bombthreat": true, "weather": true,
	"violence": true, "bloodborne": true, "hazcom": true, "elevator": true, "safety": true,
	"fire": true, "heat": true, "ladder": true, "lockout": true, "mentalhealth": true,
	"winter": true, "electrical": true, "spaces": true, "ppe": true, "medical": true,
	"earthquake": true, "foodsafety": true, "mail": true, "evacassist": true, "warehouse": true,
	"ramp": true, "cabin": true, "dgr": true,
	"fallprotection": true, "forklift": true, "hotwork": true, "chemical": true, "lab": true,
	"machineguarding": true, "lifting": true, "hearing": true, "slips": true, "respirator": true,
	"lithium": true, "nearmiss": true, "fatigue": true,
}

func weighted(r *rnd, weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return 0
	}
	x := r.intn(total)
	for i, w := range weights {
		if x < w {
			return i
		}
		x -= w
	}
	return len(weights) - 1
}

// withChoices shuffles right + wrongs into a question, recording where the
// right answer landed.
func withChoices(r *rnd, right string, wrongs []string) ([]string, int) {
	all := append([]string{right}, wrongs...)
	idx := make([]int, len(all))
	for i := range idx {
		idx[i] = i
	}
	shuffle(r, idx)
	out := make([]string, len(all))
	correct := 0
	for to, from := range idx {
		out[to] = all[from]
		if from == 0 {
			correct = to
		}
	}
	return out, correct
}

func bankQuestion(r *rnd, cat string, sl slots) Question {
	g := gen{r, sl}
	i := r.intn(len(banks[cat]))
	bq := banks[cat][i]
	text := g.expand(bq.Text)
	if r.chance(0.45) && len(fragments["preamble"]) > 0 {
		text = g.expand("{@preamble}") + " " + text
	}
	q := Question{Text: text, key: fmt.Sprintf("%s/%d", cat, i)}
	if len(bq.Right) >= 2 && r.chance(0.6) {
		// "All of the above": every listed option is right.
		for _, s := range sample(r, bq.Right, 3) {
			q.Choices = append(q.Choices, g.expand(s))
		}
		q.Choices = append(q.Choices, pick(r, []string{"All of the above", "All of the above.", "All of these"}))
		q.Correct = len(q.Choices) - 1
		q.Prompt = "Select the BEST answer."
		return q
	}
	var wrongs []string
	for _, s := range sample(r, bq.Wrong, 3) {
		wrongs = append(wrongs, g.expand(s))
	}
	q.Choices, q.Correct = withChoices(r, g.expand(pick(r, bq.Right)), wrongs)
	if len(q.Choices) < 5 && r.chance(0.15) {
		q.Choices = append(q.Choices, "None of the above")
	}
	return q
}

func agencyQ(r *rnd) Question {
	a := randomAgency(r)
	var others []agency
	seen := map[string]bool{a.Name: true, a.Domain: true}
	for len(others) < 3 {
		o := randomAgency(r)
		if !seen[o.Name] && !seen[o.Domain] {
			seen[o.Name], seen[o.Domain] = true, true
			others = append(others, o)
		}
	}
	q := Question{key: "agency/" + a.Slug, Ref: "Ref: ACPWB Regulatory Map, " + a.Code}
	var wrongs []string
	if r.chance(0.5) {
		q.Text = fmt.Sprintf(pick(r, []string{
			"Which agency is primarily responsible for %s?",
			"Questions about %s should be directed to which regulator?",
			"Under the ACPWB Regulatory Map, oversight of %s rests with:",
			"An associate has a concern about %s. Which body has jurisdiction?",
		}), a.Domain)
		for _, o := range others {
			wrongs = append(wrongs, o.Name)
		}
		q.Choices, q.Correct = withChoices(r, a.Name, wrongs)
		return q
	}
	q.Text = fmt.Sprintf(pick(r, []string{
		"The %s primarily oversees which of the following?",
		"Which area falls within the jurisdiction of the %s?",
		"ACPWB associates contact the %s about:",
	}), a.Name)
	for _, o := range others {
		wrongs = append(wrongs, capFirst(o.Domain))
	}
	q.Choices, q.Correct = withChoices(r, capFirst(a.Domain), wrongs)
	return q
}

var fakeActWords = []string{"Workforce", "Remuneration", "Fair Wage", "Disclosure", "Equity", "Accountability",
	"Transparency", "Employee Dignity", "Stakeholder", "Compensation Integrity", "Benefits Modernization"}
var fakeActKinds = []string{"Act", "Reform Act", "Protection Act", "Improvement Act", "Clarification Act"}

func fakeStatute(r *rnd) string {
	return fmt.Sprintf("the %s %s of %d", pick(r, fakeActWords), pick(r, fakeActKinds), r.between(1971, 2024))
}

func statuteQ(r *rnd, c *Course) Question {
	if r.chance(0.5) {
		// A framework citation and what it governs, e.g. "FLSA §207" → "overtime provisions".
		fw := pick(r, frameworks)
		code, desc, ok := strings.Cut(fw, " (")
		if ok {
			desc = strings.TrimSuffix(desc, ")")
			var wrongs []string
			seen := map[string]bool{code: true}
			for len(wrongs) < 3 {
				o, _, _ := strings.Cut(pick(r, frameworks), " (")
				if !seen[o] {
					seen[o] = true
					wrongs = append(wrongs, o)
				}
			}
			q := Question{Text: fmt.Sprintf("Which provision governs %s?", desc), key: "fw/" + code}
			q.Choices, q.Correct = withChoices(r, code, wrongs)
			return q
		}
	}
	law := pick(r, legislation)
	q := Question{Text: fmt.Sprintf("Which of the following is cited in the %s policy as a governing statute?", c.Type.Title), key: "law/" + law}
	var wrongs []string
	for len(wrongs) < 3 {
		wrongs = append(wrongs, fakeStatute(r))
	}
	q.Choices, q.Correct = withChoices(r, capFirst(law), capAll(wrongs))
	return q
}

func wrongFactQ(r *rnd) Question {
	fact := pick(r, wrongFacts)
	q := Question{
		key:     "fact/" + fact,
		Prompt:  "True or False",
		Text:    fact,
		Choices: []string{"True", "False"},
		Ref:     "Ref: ACPWB Corporate Governance Wiki",
	}
	if r.chance(0.3) {
		q.Choices, q.Correct = []string{"False", "True"}, 1
	}
	return q
}

func scenarioQ(r *rnd, c *Course, sl slots) Question {
	finding := sl.fill(pick(r, findings))
	right := firstSentence(sl.fill(pick(r, correctives)))
	var wrongs []string
	for _, s := range sample(r, correctives, 2) {
		if w := firstSentence(sl.fill(s)); w != right {
			wrongs = append(wrongs, w)
		}
	}
	wrongs = append(wrongs, pick(r, []string{
		"Wait for the next scheduled audit cycle before taking any action.",
		"Document the finding in a personal notebook and revisit it next fiscal year.",
		"Notify the finding's subject informally so they can correct it quietly.",
		"Close the finding, since it was identified internally.",
	}))
	q := Question{
		key:    "scenario/" + finding,
		Prompt: "Scenario",
		Text: fmt.Sprintf("During a %s review, %s (%s, %s) identifies the following finding: “%s.” Under %s, what is the REQUIRED first step?",
			c.Provider, sl["NAME"], sl["TITLE"], sl["DEPT"], strings.TrimSuffix(finding, "."), c.Framework),
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

func perchQ(r *rnd) Question {
	var held []perchFact
	for _, f := range perchFacts {
		if !f.Cancelled && f.Venue != "" {
			held = append(held, f)
		}
	}
	f := pick(r, held)
	others := func(get func(perchFact) string) []string {
		seen := map[string]bool{get(f): true}
		var out []string
		for _, o := range sample(r, held, len(held)) {
			if v := get(o); !seen[v] && len(out) < 3 {
				seen[v] = true
				out = append(out, v)
			}
		}
		return out
	}
	q := Question{key: fmt.Sprintf("perch/%d", f.Year)}
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Where was PERCH %d held?", f.Year)
		q.Choices, q.Correct = withChoices(r, f.Venue, others(func(p perchFact) string { return p.Venue }))
	case 1:
		q.Text = fmt.Sprintf("What was the theme of PERCH %d?", f.Year)
		q.Choices, q.Correct = withChoices(r, "“"+f.Theme+"”", quoteAll(others(func(p perchFact) string { return p.Theme })))
	default:
		q.Text = fmt.Sprintf("On what dates was PERCH %d held?", f.Year)
		q.Choices, q.Correct = withChoices(r, f.Dates, others(func(p perchFact) string { return p.Dates }))
	}
	q.Ref = fmt.Sprintf("Ref: PERCH %d Program, /perch-conference/%d/", f.Year, f.Year)
	return q
}

func reportQ(r *rnd, sl slots) Question {
	rep := pick(r, reports)
	var wrongs []string
	seen := map[string]bool{rep.Title: true}
	for _, o := range sample(r, reports, len(reports)) {
		if o.Category != rep.Category && !seen[o.Title] && len(wrongs) < 3 {
			seen[o.Title] = true
			wrongs = append(wrongs, o.Title)
		}
	}
	cat := strings.ToLower(rep.Category)
	q := Question{key: "report/" + rep.Slug}
	if r.chance(0.3) {
		// Reverse: which series does a report belong to?
		var cats []string
		for _, o := range reports {
			if o.Category != rep.Category && !contains(cats, o.Category) {
				cats = append(cats, o.Category)
			}
		}
		if len(cats) >= 3 {
			q.Text = fmt.Sprintf(pick(r, []string{
				"The ACPWB report \"%s\" belongs to which publication series?",
				"Under which category is \"%s\" filed in the ACPWB Reports & Publications archive?",
				"\"%s\" is primarily a source of guidance on:",
			}), rep.Title)
			q.Choices, q.Correct = withChoices(r, rep.Category, sample(r, cats, 3))
			q.Ref = "Ref: ACPWB Reports & Publications, /reports/" + rep.Slug + "/"
			return q
		}
	}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Which ACPWB publication should you consult for current %s guidance?",
		"{NAME} in {DEPT} needs the latest %s data for a briefing. Which ACPWB report should {FIRST} cite?",
		"Which of the following is an ACPWB %s publication?",
		"A regulator asks which ACPWB report documents your %s practices. You should point them to:",
		"Which report in the ACPWB archive covers %s?",
		"Your manager asks for the ACPWB reference on %s. Which publication do you send?",
	}), cat)
	q.Text = sl.fill(q.Text)
	q.Choices, q.Correct = withChoices(r, rep.Title, wrongs)
	q.Ref = "Ref: ACPWB Reports & Publications, /reports/" + rep.Slug + "/"
	return q
}

var chartSubjects = []string{
	"tailgating incidents", "lost badges", "reported phishing emails", "unlocked workstations",
	"clean desk violations", "late expense reports", "gift register entries", "slip-and-fall near misses",
	"Integrity Line calls", "privacy incidents", "unescorted visitors", "propped doors",
}

func chartQ(r *rnd, c *Course, n int) Question {
	subj := pick(r, chartSubjects)
	year := r.between(2015, 2026)
	labels := []string{"Q1", "Q2", "Q3", "Q4"}
	vals := make([]int, 4)
	for i := range vals {
		vals[i] = r.between(4, 60)
	}
	if r.chance(0.45) {
		return chartMathQ(r, c, n, subj, year, labels, vals)
	}
	// Make the extreme unique so the key is unambiguous.
	most := r.chance(0.6)
	target := r.intn(4)
	vals[target] = 75
	if !most {
		vals[target] = 2
	}
	ex := exhibitLabel(r, n)
	q := Question{
		key:     "chart/" + subj,
		ExLabel: ex,
		Exhibit: barChartSVG(fmt.Sprintf("Reported %s by quarter, %d", subj, year), labels, vals),
	}
	if most {
		q.Text = fmt.Sprintf("Refer to %s. In which quarter were the MOST %s reported?", ex, subj)
	} else {
		q.Text = fmt.Sprintf("Refer to %s. In which quarter were the FEWEST %s reported?", ex, subj)
	}
	q.Choices = []string{"Q1", "Q2", "Q3", "Q4"}
	q.Correct = target
	q.Ref = fmt.Sprintf("Source: %s Incident Dashboard, FY%d", c.Dept, year)
	return q
}

// chartMathQ asks for arithmetic on a labelled chart: a total, a change
// between quarters, or an average.
func chartMathQ(r *rnd, c *Course, n int, subj string, year int, labels []string, vals []int) Question {
	ex := exhibitLabel(r, n)
	q := Question{key: "chartmath/" + subj, ExLabel: ex, Prompt: "Use the exhibit to answer.",
		Exhibit: barChartSVG(fmt.Sprintf("Reported %s by quarter, %d", subj, year), labels, vals, true),
		Ref:     fmt.Sprintf("Source: %s Incident Dashboard, FY%d", c.Dept, year)}
	switch r.intn(3) {
	case 0:
		total := 0
		for _, v := range vals {
			total += v
		}
		q.Text = fmt.Sprintf("Refer to %s. How many %s were reported in %d in total?", ex, subj, year)
		q.Choices, q.Correct = numChoices(r, total, []int{-10, 10, -vals[0], vals[3], -1, 1}, itoa)
	case 1:
		a, b := 0, 1+r.intn(3)
		if vals[a] == vals[b] {
			vals[b]++
			q.Exhibit = barChartSVG(fmt.Sprintf("Reported %s by quarter, %d", subj, year), labels, vals, true)
		}
		d := vals[b] - vals[a]
		word := "increase"
		if d < 0 {
			word, d = "decrease", -d
		}
		q.Text = fmt.Sprintf("Refer to %s. By how much did reported %s %s from %s to %s?", ex, subj, word, labels[a], labels[b])
		q.Choices, q.Correct = numChoices(r, d, []int{-5, 5, -1, 1, vals[a] - d, 10}, itoa)
	default:
		for (vals[0]+vals[1]+vals[2]+vals[3])%4 != 0 {
			vals[3]++
		}
		q.Exhibit = barChartSVG(fmt.Sprintf("Reported %s by quarter, %d", subj, year), labels, vals, true)
		avg := (vals[0] + vals[1] + vals[2] + vals[3]) / 4
		q.Text = fmt.Sprintf("Refer to %s. What was the average number of %s reported per quarter in %d?", ex, subj, year)
		q.Choices, q.Correct = numChoices(r, avg, []int{-3, 3, -1, 1, 5, -5}, itoa)
	}
	return q
}

func badgeExhibitQ(r *rnd, c *Course, sl slots, n int) Question {
	b := badgeData{
		Person:  person{Name: sl["NAME"], Title: sl["TITLE"], Dept: sl["DEPT"], Headshot: r.intn(400)},
		ID:      fmt.Sprintf("%06d", r.between(100000, 999999)),
		Floors:  sample(r, []string{"L", "2", "12", "18", "27", "33", "39", "40", "P1", "P2"}, r.between(2, 4)),
		Stripe:  pick(r, []string{"#0a1628", "#c9a84c", "#8b1e1e", "#1f5f3a", "#3b3f8f"}),
		Kind:    pick(r, []string{"ASSOCIATE", "CONTRACTOR", "VISITOR — ESCORT REQUIRED", "INTERN"}),
		Expires: time.Date(r.between(2019, 2027), time.Month(r.between(1, 12)), 28, 0, 0, 0, 0, time.UTC),
	}
	ex := exhibitLabel(r, n)
	q := Question{ExLabel: ex, Exhibit: badgeSVG(b), key: "badgeex/" + b.ID}
	if r.chance(0.5) {
		on := b.Expires.AddDate(0, r.between(-14, 14), 0)
		if on.Equal(b.Expires) {
			on = on.AddDate(0, 1, 0)
		}
		q.Text = fmt.Sprintf("Refer to %s. %s presents this badge at the %s on %s. Should entry be permitted?",
			ex, b.Person.First(), sl["DOOR"], on.Format("January 2, 2006"))
		valid := on.Before(b.Expires)
		right := "No. The badge has expired; direct the holder to the Security Office"
		if valid {
			right = "Yes. The badge is current"
			if strings.HasPrefix(b.Kind, "VISITOR") {
				right = "Yes, but only with an ACPWB escort present"
			}
		}
		wrongs := []string{"Yes, if the holder is known to you", "Only if the badge photo is in color", "No. Badges are never valid at this entrance"}
		if !valid {
			wrongs[0] = "Yes. The badge is current"
		}
		q.Choices, q.Correct = withChoices(r, right, wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Which of the following floors may this badge access?", ex)
	right := "Floor " + pick(r, b.Floors)
	var wrongs []string
	for _, f := range []string{"L", "2", "12", "18", "27", "33", "39", "40", "P1", "P2"} {
		if !contains(b.Floors, f) && len(wrongs) < 3 {
			wrongs = append(wrongs, "Floor "+f)
		}
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

func exhibitLabel(r *rnd, n int) string {
	return fmt.Sprintf("Exhibit %d-%c", n, 'A'+rune(r.intn(6)))
}

// ── slots ────────────────────────────────────────────────────────────────

type slots map[string]string

func newSlots(r *rnd, c *Course) slots {
	first, last := pick(r, firstNames), pick(r, lastNames)
	first2, last2 := pick(r, firstNames), pick(r, lastNames)
	s := slots{
		"NAME":      first + " " + last,
		"FIRST":     first,
		"NAME2":     first2 + " " + last2,
		"FIRST2":    first2,
		"DEPT2":     pick(r, departments),
		"TITLE2":    pick(r, titles),
		"DEPT":      pick(r, departments),
		"TITLE":     pick(r, titles),
		"N":         fmt.Sprint(r.between(2, 12)),
		"DAYS":      fmt.Sprint(pick(r, []int{3, 5, 7, 10, 14, 30, 45, 60, 90})),
		"AMOUNT":    fmt.Sprintf("$%d", pick(r, []int{25, 50, 75, 100, 150, 250, 500})),
		"FORM":      fmt.Sprintf("%s-%d%c", pick(r, []string{"SEC", "FAC", "HR", "CMP", "GC", "IT"}), r.between(10, 99), 'A'+rune(r.intn(6))),
		"POLICY":    fmt.Sprintf("%d.%d.%d", r.between(1, 12), r.between(1, 20), r.between(1, 9)),
		"FLOOR":     ordinal(r.between(12, 40)), // "33rd"
		"AGENCY":    randomAgency(r).Name,
		"ORG":       pick(r, orgs),
		"COMMITTEE": pick(r, committees),
		"COURSE":    c.Title,
		"TOPIC":     c.Type.Title,
		"YEAR":      fmt.Sprint(r.between(2014, 2026)),
		// lowercase slots used by the shared archive compliance pools
		"n":           fmt.Sprint(r.between(3, 140)),
		"pct":         fmt.Sprint(r.between(3, 22)),
		"regions":     fmt.Sprint(r.between(2, 9)),
		"year":        fmt.Sprint(r.between(2008, 2025)),
		"frameworks":  c.Framework,
		"doc_version": c.Version,
		"org":         c.Provider,
	}
	for _, k := range varNames { // sorted: map order would make slots non-deterministic
		s[k] = pick(r, vars[k])
	}
	return s
}

var slotRe = regexp.MustCompile(`\{([A-Za-z_]+)\}`)

func (s slots) fill(t string) string {
	return slotRe.ReplaceAllStringFunc(t, func(m string) string {
		if v, ok := s[m[1:len(m)-1]]; ok {
			return v
		}
		return m
	})
}

// ── small helpers ────────────────────────────────────────────────────────

func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}

func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func capAll(s []string) []string {
	out := make([]string, len(s))
	for i, v := range s {
		out[i] = capFirst(v)
	}
	return out
}

func quoteAll(s []string) []string {
	out := make([]string, len(s))
	for i, v := range s {
		out[i] = "“" + v + "”"
	}
	return out
}

// ordinal is "1st", "22nd", "13th"…
func ordinal(n int) string {
	suf := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suf = "st"
		case 2:
			suf = "nd"
		case 3:
			suf = "rd"
		}
	}
	return fmt.Sprint(n) + suf
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
