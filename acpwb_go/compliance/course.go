package compliance

import (
	"acpwb_go/legacyjob"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// A course lives at /compliance/<agency slug>/<course-type slug>-<5 digits>/,
// e.g. /compliance/sec/insider-trading-awareness-48213/. The mandating agency
// is part of the course's identity: the same type and number under another
// agency is a different course. Course.ID ("sec/insider-trading-awareness-48213")
// seeds everything else.
const (
	minCourseNumber = 10000
	maxCourseNumber = 99999
	maxQuestions    = 25
	maxAttempt      = 9999
)

type person struct {
	Name, Title, Dept string
	Headshot          int // static/img/headshots/NNN.webp
}

// module is one course-outline entry with its lessons.
type module struct {
	Title   string
	Minutes int
	Lessons []string
}

func (p person) First() string { f, _, _ := strings.Cut(p.Name, " "); return f }

// Email is the person's address in the site-wide first.last@acpwb.com format.
func (p person) Email() string {
	f := strings.Fields(p.Name)
	if len(f) == 0 {
		return ""
	}
	return legacyjob.CodeEmail(f[0], f[len(f)-1])
}

func (p person) HeadshotURL() string {
	return fmt.Sprintf("/static/img/headshots/%03d.webp", p.Headshot)
}

type Course struct {
	ID         string // "<agency>/<slug>", the seed and record key
	Slug       string // "<course-type slug>-<number>"
	Type       *courseType
	Title      string // this course's own title: the type title plus audience/level/edition variants
	Number     int
	Code       string
	Agency     agency // the regulator whose rule "mandates" it
	Provider   string // the consultancy that "delivers" it
	Dept       string // the ACPWB department that "assigns" it
	Framework  string
	N          int // questions per attempt
	Pass       int // pass mark, percent
	Minutes    int
	Credits    string
	Version    string
	Revised    time.Time
	Closes     time.Time // enrollment window end
	Modules    []module
	Instructor person
	Hero       string
	Rules      []string
	// Requirements is the Key Requirements table on the course page,
	// which requirement questions ask about.
	Requirements []courseReq
}

func (c *Course) URL() string { return "/compliance/" + c.ID + "/" }

func (c *Course) AttemptURL(k int) string { return fmt.Sprintf("%sattempt-%d/", c.URL(), k) }

func (c *Course) SponsorURL() string { return agencyURL(c.Agency) }

func (c *Course) TopicURL() string { return "/compliance/topic/" + c.Type.Slug + "/" }

func courseSlug(ct *courseType, number int) string { return fmt.Sprintf("%s-%d", ct.Slug, number) }

// ParseCourse resolves an agency slug plus course slug, or reports false for
// anything that isn't a known agency, a known course type, and a 5-digit
// number.
func ParseCourse(agencySlug, slug string) (*Course, bool) {
	a, ok := lookupAgency(agencySlug)
	if !ok {
		return nil, false
	}
	i := strings.LastIndexByte(slug, '-')
	if i < 0 || len(slug)-i-1 != 5 {
		return nil, false
	}
	ct, ok := courseTypeMap[slug[:i]]
	if !ok {
		return nil, false
	}
	n, err := strconv.Atoi(slug[i+1:])
	if err != nil || n < minCourseNumber || n > maxCourseNumber {
		return nil, false
	}
	return getCourse(a, ct, n), true
}

// ParseCourseID resolves a Course.ID ("<agency>/<slug>").
func ParseCourseID(id string) (*Course, bool) {
	ag, slug, ok := strings.Cut(id, "/")
	if !ok {
		return nil, false
	}
	return ParseCourse(ag, slug)
}

// Courses are re-derived on every answer page, so recent ones are cached.
var courseCache = newCache[*Course](8192)

func getCourse(a agency, ct *courseType, number int) *Course {
	id := a.Slug + "/" + courseSlug(ct, number)
	if c, ok := courseCache.get(id); ok {
		return c
	}
	c := buildCourse(a, ct, number)
	courseCache.put(id, c)
	return c
}

var creditKinds = []string{"CPE", "CLE", "SHRM-PDC", "HRCI", "CCEP", "CEU"}

func buildCourse(a agency, ct *courseType, number int) *Course {
	slug := courseSlug(ct, number)
	id := a.Slug + "/" + slug
	r := newRnd("compliance_course_" + id)
	c := &Course{ID: id, Slug: slug, Type: ct, Number: number, Agency: a}
	c.Code = fmt.Sprintf("%s-%04d-R%d", ct.Prefix, number%10000, r.between(1, 9))
	c.Provider = pick(r, orgs)
	c.Dept = pick(r, departments)
	c.Framework = pick(r, frameworks)
	c.N = r.between(ct.MinQ, ct.MaxQ)
	c.Pass = ct.Pass
	if r.chance(0.08) {
		c.Pass = 100
	}
	c.Minutes = r.between(3, 24) * 5
	c.Credits = fmt.Sprintf("%.1f %s", float64(r.between(1, 6))*0.5, pick(r, creditKinds))
	c.Version = fmt.Sprintf("%d.%d.%d", r.between(1, 12), r.between(0, 9), r.between(0, 30))
	c.Revised = time.Date(r.between(2019, 2026), time.Month(r.between(1, 12)), r.between(1, 28), 0, 0, 0, 0, time.UTC)
	c.Closes = time.Date(r.between(2025, 2027), time.Month(r.between(1, 12)), r.between(1, 28), 0, 0, 0, 0, time.UTC)
	c.Title = courseTitle(newRnd("compliance_title_"+id), ct, a, c.Revised.Year())
	c.Instructor = person{
		Name:     pick(r, firstNames) + " " + pick(r, lastNames),
		Title:    pick(r, titles),
		Dept:     pick(r, departments),
		Headshot: r.intn(400),
	}
	c.Hero = "/static/img/report-covers/" + pick(r, reportCovers) + ".webp"
	if tc := topicCovers[ct.Slug]; len(tc) > 0 && r.chance(0.85) {
		c.Hero = "/static/img/compliance-covers/" + pick(r, tc) + ".webp"
	}
	buildOutline(c)
	c.Requirements = courseRequirements(newRnd("compliance_reqs_"+id), ct)
	return c
}

// buildOutline writes the course outline and "Before You Begin" rules
// from the module/lesson/rule grammars, with the course's own slots.
func buildOutline(c *Course) {
	r := newRnd("compliance_outline_" + c.ID)
	// Fresh slots per line, so modules and lessons don't all name the same
	// person or document; AGENCY is always the course's own.
	fresh := func() gen {
		sl := newSlots(r, c)
		sl["AGENCY"] = c.Agency.Name
		return gen{r, sl}
	}
	used := map[string]bool{}
	distinct := func(src string) string {
		for try := 0; ; try++ {
			if s := fresh().expand(src); !used[s] || try == 10 {
				used[s] = true
				return s
			}
		}
	}
	// fragLine expands a fragment template not already used in this course,
	// so two outline entries or rules never come from the same line (their
	// expansions differ only in the optional bits, which reads as a repeat).
	usedLine := map[string]bool{}
	fragLine := func(name string) string {
		lines := fragments[name]
		for try := 0; try < 20; try++ {
			i := r.intn(len(lines))
			if k := fmt.Sprintf("%s/%d", name, i); !usedLine[k] || try == 19 {
				usedLine[k] = true
				return distinct(lines[i])
			}
		}
		return distinct("{@" + name + "}")
	}
	add := func(title string) {
		m := module{Title: title, Minutes: r.between(3, 18)}
		for i := r.between(0, 3); i > 0; i-- {
			m.Lessons = append(m.Lessons, fragLine("lesson"))
		}
		c.Modules = append(c.Modules, m)
	}
	add(distinct("Introduction[? & Purpose| to {TOPIC}| from the {DEPT} Team]"))
	for i := r.between(4, 9); i > 0; i-- {
		add(fragLine("module"))
	}
	add(distinct("[Knowledge Check|Final Assessment|Assessment & Attestation][? ({N} Questions)]"))
	for i := r.between(3, 6); i > 0; i-- {
		c.Rules = append(c.Rules, fragLine("rule"))
	}
}

// randomCourse picks a course from r: any agency, type, and number.
func randomCourse(r *rnd) *Course {
	return agencyCourse(r, randomAgency(r))
}

// agencyCourse picks one of an agency's courses.
func agencyCourse(r *rnd, a agency) *Course {
	return getCourse(a, pick(r, courseTypes), r.between(minCourseNumber, maxCourseNumber))
}

// availability is the catalog badge for a course's enrollment window,
// relative to now. Deliberately calm: no overdue or deadline pressure, so the
// catalog reads as an open course listing rather than a phishing-style
// "action required" notice. A window that has passed simply reopens.
func availability(closes, now time.Time) (label, class string) {
	days := int(closes.Sub(now).Hours() / 24)
	switch {
	case days < 0:
		return "Open enrollment", "cmp-st-ok"
	case days <= 14:
		return "New session open", "cmp-st-prog"
	}
	return "Available through " + closes.Format("Jan 2, 2006"), "cmp-st-ok"
}

// Title variants, so the same course type reads differently under each
// agency: "Active Shooter Response Training for Front-Desk Staff",
// "Fundamentals of Records Retention & Disposal — Level 2 (Practitioner)",
// "Clean Desk Policy Certification (2024 Edition)". Seeded by course ID, on
// its own RNG so adding variants never shifts any other course field.
var (
	titlePrefixes  = []string{"Introduction to", "Essentials of", "Fundamentals of", "Foundations of", "Applied", "Advanced", "Practical", "Core", "Understanding", "Principles of", "A Field Guide to", "Getting Started with"}
	titleAudiences = []string{"for People Managers", "for Contractors & Temporary Staff", "for New Associates", "for Remote & Hybrid Associates",
		"for Executives & Officers", "for Front-Desk Staff", "for Field & Facilities Staff", "for Interns & Co-ops", "for Board Members",
		"for Finance Teams", "for Senior Leaders", "for Part-Time Associates", "for Floor Wardens", "for Shift Supervisors",
		"for Customer-Facing Teams", "for Project Leads", "for Executive Assistants", "for Returning Associates", "for Multi-Site Teams",
		"for Consultants & Advisors", "for Data Stewards", "for Event Staff & Volunteers"}
	titleLevels = []string{"Level 1", "Level 2 (Practitioner)", "Level 3 (Expert)", "Advanced Module", "Refresher", "Recertification",
		"Scenario Workshop", "Self-Study Edition", "Micro-Module", "Knowledge Check", "Instructor-Led Replay", "Case Study Series",
		"Quick Reference Edition", "Deep Dive", "Lunch & Learn", "Tabletop Exercise", "Bootcamp", "Practicum", "Supplemental Module",
		"Leadership Edition", "Compliance Clinic", "Policy Walkthrough"}
	titleEditions = []string{"%d Edition", "%d Update", "Revised %d", "%d Cycle", "FY%d"}
)

func courseTitle(r *rnd, ct *courseType, a agency, year int) string {
	t := ct.Title
	lower := strings.ToLower(t)
	hasFor := strings.Contains(lower, " for ")
	hasSep := strings.ContainsAny(t, ":—")
	paren := strings.HasSuffix(t, ")")
	if r.chance(0.2) {
		return t
	}
	if r.chance(0.3) {
		p := pick(r, titlePrefixes)
		if !strings.Contains(lower, strings.ToLower(strings.Fields(p)[0])) {
			t = p + " " + t
		}
	}
	added := false
	if !hasFor && r.chance(0.45) {
		t += " " + pick(r, titleAudiences)
		added = true
	}
	switch {
	case !hasSep && (!added || r.chance(0.35)) && r.chance(0.6):
		lvl := pick(r, titleLevels)
		if r.chance(0.12) {
			m := r.between(2, 6)
			lvl = fmt.Sprintf("Part %d of %d", r.between(1, m), m)
		}
		if r.chance(0.15) && a.Code != "" {
			lvl = a.Code + " Requirements"
		}
		if !strings.Contains(lower, strings.ToLower(lvl)) {
			t += " — " + lvl
		}
	case !paren && r.chance(0.5):
		t += " (" + fmt.Sprintf(pick(r, titleEditions), year) + ")"
	}
	return t
}
