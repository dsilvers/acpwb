package compliance

import (
	"fmt"
	"strconv"
	"strings"
)

// Generated question kinds shared by every course: glossary definitions,
// the course's own Key Requirements table, and facts shown on the course
// page. genQuestion draws from universalMix for part of every attempt, so
// even a course with a narrow topic mix rarely repeats itself.

type glossEntry struct{ Term, Def string }

type reqLine struct {
	Cat, Phrase, Unit string
	Min, Max          int
}

// courseReq is one row of a course's Key Requirements table.
type courseReq struct {
	Phrase, Unit string
	Value        int
	min, max     int
}

var (
	glossary     []glossEntry
	requirements []reqLine
)

var universalMix = []mixEntry{
	{"definition", 5}, {"requirement", 4}, {"agency", 2}, {"coursefact", 2},
	{"statute", 1}, {"wrongfact", 1}, {"report", 1}, {"chart", 2}, {"exhibit", 10},
}

func loadKinds() {
	for _, l := range lines(readData("data/glossary.txt")) {
		t, d, ok := strings.Cut(l, "|")
		if !ok {
			panic("compliance: bad glossary line: " + l)
		}
		glossary = append(glossary, glossEntry{strings.TrimSpace(t), strings.TrimSpace(d)})
	}
	for _, l := range lines(readData("data/requirements.txt")) {
		f := strings.Split(l, "|")
		if len(f) != 4 {
			panic("compliance: bad requirement line: " + l)
		}
		lo, hi, _ := strings.Cut(strings.TrimSpace(f[3]), "-")
		mn, _ := strconv.Atoi(lo)
		mx, _ := strconv.Atoi(hi)
		requirements = append(requirements, reqLine{strings.TrimSpace(f[0]), strings.TrimSpace(f[1]), strings.TrimSpace(f[2]), mn, mx})
	}
	generatedKinds["definition"] = true
	generatedKinds["requirement"] = true
	generatedKinds["coursefact"] = true
}

// courseRequirements picks the Key Requirements rows for a course: the
// "any" rows plus those matching its question categories, values fixed per
// course.
func courseRequirements(r *rnd, ct *courseType) []courseReq {
	cats := map[string]bool{"any": true}
	for _, m := range ct.Mix {
		cats[m.cat] = true
	}
	var pool []reqLine
	for _, q := range requirements {
		if cats[q.Cat] {
			pool = append(pool, q)
		}
	}
	var out []courseReq
	for _, q := range sample(r, pool, r.between(6, 9)) {
		out = append(out, courseReq{q.Phrase, q.Unit, r.between(q.Min, q.Max), q.Min, q.Max})
	}
	return out
}

// Standard renders a requirement's value: "Within 4 hours", "Above $50", ….
func (q courseReq) Standard() string { return reqText(q.Unit, q.Value) }

func reqText(unit string, v int) string {
	switch unit {
	case "dollars":
		return fmt.Sprintf("Expenses above $%d", v)
	case "number of bids":
		return fmt.Sprintf("At least %d bids", v)
	}
	if v == 1 {
		unit = strings.TrimSuffix(unit, "s")
	}
	return fmt.Sprintf("Within %d %s", v, unit)
}

func definitionQ(r *rnd) Question {
	e := pick(r, glossary)
	others := func() []glossEntry {
		var out []glossEntry
		for _, o := range sample(r, glossary, 8) {
			if o.Term != e.Term && len(out) < 3 {
				out = append(out, o)
			}
		}
		return out
	}()
	q := Question{key: "def/" + e.Term, Ref: "Ref: ACPWB Compliance Glossary"}
	var wrongs []string
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Which of the following best defines “%s”?",
			"In ACPWB policy, “%s” refers to:",
			"What is meant by “%s”?",
		}), e.Term)
		for _, o := range others {
			wrongs = append(wrongs, o.Def)
		}
		q.Choices, q.Correct = withChoices(r, e.Def, wrongs)
	case 1:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Which term describes the following: %s?",
			"“%s.” This describes which of the following?",
			"Which term is defined as: %s?",
		}), strings.TrimSuffix(lowerFirst(e.Def), "."))
		for _, o := range others {
			wrongs = append(wrongs, o.Term)
		}
		q.Choices, q.Correct = withChoices(r, e.Term, wrongs)
	default:
		def, truth := e.Def, true
		if r.chance(0.5) {
			def, truth = others[0].Def, false
		}
		q.Prompt = "True or False"
		q.Text = fmt.Sprintf("“%s” means %s.", e.Term, strings.TrimSuffix(lowerFirst(def), "."))
		q.Choices = []string{"True", "False"}
		if !truth {
			q.Correct = 1
		}
	}
	return q
}

func requirementQ(r *rnd, c *Course) Question {
	req := pick(r, c.Requirements)
	q := Question{key: "req/" + req.Phrase, Ref: fmt.Sprintf("Ref: %s, Key Requirements", c.Code)}
	q.Text = fmt.Sprintf(pick(r, []string{
		"According to the Key Requirements for %[1]s, what is the standard to %[2]s?",
		"Under %[1]s, associates must %[2]s:",
		"Per the %[3]s policy, what is the requirement to %[2]s?",
		"How quickly must associates %[2]s under %[1]s?",
		"The %[3]s Key Requirements table sets which standard to %[2]s?",
		"You need to %[2]s. According to %[1]s, the standard is:",
		"Which standard applies when you %[2]s, according to the %[3]s course page?",
		"Refer to the Key Requirements for %[1]s. What is the deadline to %[2]s?",
	}), c.Code, req.Phrase, c.Title)
	seen := map[int]bool{req.Value: true}
	var wrongs []string
	for tries := 0; len(wrongs) < 3 && tries < 50; tries++ {
		v := r.between(req.min, req.max)
		if tries > 20 {
			v = req.Value + r.between(1, 12)*(1+tries%2*-2)
		}
		if v >= 0 && !seen[v] {
			seen[v] = true
			wrongs = append(wrongs, reqText(req.Unit, v))
		}
	}
	q.Choices, q.Correct = withChoices(r, req.Standard(), wrongs)
	return q
}

func courseFactQ(r *rnd, c *Course) Question {
	q := Question{Ref: fmt.Sprintf("Ref: %s course overview", c.Code)}
	ask := func(key string, texts ...string) {
		q.key, q.Text = "fact/"+key, pick(r, texts)
	}
	title, code := c.Title, c.Code
	var right string
	var wrongs []string
	others := func(pool []string, n int) {
		for _, o := range sample(r, pool, len(pool)) {
			if o != right && !contains(wrongs, o) && len(wrongs) < n {
				wrongs = append(wrongs, o)
			}
		}
	}
	nums := func(v int, f func(int) string, deltas ...int) {
		right = f(v)
		for _, d := range sample(r, deltas, len(deltas)) {
			if w := f(v + d); v+d > 0 && w != right && !contains(wrongs, w) && len(wrongs) < 3 {
				wrongs = append(wrongs, w)
			}
		}
	}
	switch r.intn(11) {
	case 0:
		ask("agency", fmt.Sprintf("Which body mandates %s (%s)?", title, code),
			fmt.Sprintf("%s is required by which regulator?", title),
			fmt.Sprintf("According to the course overview, which agency mandates %s?", code))
		right = c.Agency.Name
		for len(wrongs) < 3 {
			if o := randomAgency(r); o.Name != right && !contains(wrongs, o.Name) {
				wrongs = append(wrongs, o.Name)
			}
		}
	case 1:
		ask("pass", "What is the passing score for this assessment?",
			fmt.Sprintf("What score must you achieve to pass %s?", code),
			fmt.Sprintf("To receive a certificate for %s, you must score at least:", title))
		right = fmt.Sprintf("%d%%", c.Pass)
		for _, p := range sample(r, []int{60, 65, 70, 75, 80, 85, 90, 95, 100}, 9) {
			if fmt.Sprintf("%d%%", p) != right && len(wrongs) < 3 {
				wrongs = append(wrongs, fmt.Sprintf("%d%%", p))
			}
		}
	case 2:
		ask("provider", fmt.Sprintf("Which firm delivers %s on behalf of ACPWB?", title),
			fmt.Sprintf("%s content is developed and delivered by:", code),
			fmt.Sprintf("Who is the outside provider for %s?", title))
		right = c.Provider
		others(orgs, 3)
	case 3:
		ask("dept", fmt.Sprintf("Which ACPWB department offers %s?", title),
			fmt.Sprintf("Questions about %s should be directed to which department?", code),
			fmt.Sprintf("%s is sponsored internally by:", title))
		right = c.Dept
		others(departments, 3)
	case 4:
		ask("questions", fmt.Sprintf("How many questions are on the %s assessment?", code),
			"How many questions does this assessment contain?",
			fmt.Sprintf("The %s assessment consists of:", title))
		nums(c.N, func(v int) string { return fmt.Sprintf("%d questions", v) }, -5, -3, -2, 2, 3, 5, 10)
	case 5:
		ask("minutes", fmt.Sprintf("What is the estimated duration of %s?", title),
			fmt.Sprintf("How long should you set aside to complete %s?", code))
		nums(c.Minutes, func(v int) string { return fmt.Sprintf("%d minutes", v) }, -20, -10, -5, 5, 10, 15, 30)
	case 6:
		ask("credits", fmt.Sprintf("How much continuing-education credit does %s carry?", title),
			fmt.Sprintf("Completing %s earns:", code))
		right = c.Credits
		for len(wrongs) < 3 {
			w := fmt.Sprintf("%.1f %s", float64(r.between(1, 6))*0.5, pick(r, creditKinds))
			if w != right && !contains(wrongs, w) {
				wrongs = append(wrongs, w)
			}
		}
	case 7:
		ask("version", fmt.Sprintf("Which version of %s is current?", title),
			"What is the current version of this course?")
		right = "Version " + c.Version
		for len(wrongs) < 3 {
			w := fmt.Sprintf("Version %d.%d.%d", r.between(1, 12), r.between(0, 9), r.between(0, 30))
			if w != right && !contains(wrongs, w) {
				wrongs = append(wrongs, w)
			}
		}
	case 8:
		ask("revised", fmt.Sprintf("In what year was %s last revised?", code),
			fmt.Sprintf("When was the %s course content last updated?", title))
		nums(c.Revised.Year(), func(v int) string { return fmt.Sprint(v) }, -3, -2, -1, 1, 2)
	case 9:
		ask("instructor", fmt.Sprintf("Who is the instructor of record for %s?", title),
			fmt.Sprintf("Which ACPWB associate is listed as the %s course contact?", code))
		right = c.Instructor.Name
		for len(wrongs) < 3 {
			if w := pick(r, firstNames) + " " + pick(r, lastNames); w != right && !contains(wrongs, w) {
				wrongs = append(wrongs, w)
			}
		}
	default:
		ask("authority", fmt.Sprintf("Which framework is cited as the authority for %s?", title),
			fmt.Sprintf("%s is aligned with which framework?", code))
		right = c.Framework
		others(frameworks, 3)
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	// Keep acronyms and proper names ("ACPWB", "PERCH") intact.
	if len(s) > 1 && s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z' {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
