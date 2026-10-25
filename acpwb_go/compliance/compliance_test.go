package compliance

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

func get(t *testing.T, rawURL string) Response {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return Serve(Request{Path: u.Path, Query: u.Query(), Now: testNow, Fallback: "OpenAI GPTBot"})
}

func mustOK(t *testing.T, rawURL string) Response {
	t.Helper()
	r := get(t, rawURL)
	if r.Status != http.StatusOK {
		t.Fatalf("%s: status %d (redirect %q)", rawURL, r.Status, r.Redirect)
	}
	return r
}

var hrefRe = regexp.MustCompile(`href="(/compliance/[^"]*)"`)

func links(r Response) []string {
	var out []string
	for _, m := range hrefRe.FindAllStringSubmatch(r.Page.HTML, -1) {
		out = append(out, strings.ReplaceAll(m[1], "&amp;", "&"))
	}
	return out
}

func firstCourse(t *testing.T) *Course {
	t.Helper()
	c, ok := ParseCourse("doc-pto", "access-badge-renewal-attestation-48213")
	if !ok {
		t.Fatal("course did not parse")
	}
	return c
}

func TestDataLoads(t *testing.T) {
	if len(courseTypes) < 30 || len(agencies) < 100 || len(wrongFacts) == 0 || len(perchFacts) == 0 || len(reportCovers) == 0 {
		t.Fatalf("pools: %d types, %d agencies, %d facts, %d perch, %d covers",
			len(courseTypes), len(agencies), len(wrongFacts), len(perchFacts), len(reportCovers))
	}
	for _, ct := range courseTypes {
		for _, m := range ct.Mix {
			if !generatedKinds[m.cat] && len(banks[m.cat]) == 0 {
				t.Errorf("%s: question category %q has no bank and no generator", ct.Slug, m.cat)
			}
		}
	}
	for cat, qs := range banks {
		for _, q := range qs {
			if len(q.Right) == 0 || len(q.Right)+len(q.Wrong) < 2 {
				t.Errorf("%s: %q needs a right answer and at least one other", cat, q.Text)
			}
		}
	}
}

// Every question of many courses: no unfilled slot, a valid key, 2–5 choices.
func TestQuestionsWellFormed(t *testing.T) {
	unfilled := regexp.MustCompile(`\{[A-Za-z_]+\}`)
	r := newRnd("test_questions")
	for i := 0; i < 400; i++ {
		c := randomCourse(r)
		for _, q := range Questions(c, r.between(1, 5)) {
			if len(q.Choices) < 2 || len(q.Choices) > len(letters) || q.Correct < 0 || q.Correct >= len(q.Choices) {
				t.Fatalf("%s q%d (%s): %d choices, key %d", c.ID, q.N, q.Kind, len(q.Choices), q.Correct)
			}
			all := q.Text + strings.Join(q.Choices, "|")
			if m := unfilled.FindString(all); m != "" {
				t.Fatalf("%s q%d (%s): unfilled slot %s in %q", c.ID, q.N, q.Kind, m, all)
			}
		}
	}
}

func TestNoRepeatsWithinAttempt(t *testing.T) {
	r := newRnd("test_repeats")
	repeats, total := 0, 0
	for i := 0; i < 300; i++ {
		c := randomCourse(r)
		seen := map[string]bool{}
		for _, q := range Questions(c, 1) {
			total++
			if seen[q.key] {
				repeats++
			}
			seen[q.key] = true
		}
	}
	// Small banks can still run dry on long courses; keep it rare.
	if repeats*50 > total {
		t.Fatalf("%d repeated items in %d questions", repeats, total)
	}
}

func TestDeterministic(t *testing.T) {
	a := mustOK(t, "/compliance/sec/insider-trading-awareness-48213/attempt-2/ab/")
	courseCache = newCache[*Course](8192)
	questionCache = newCache[[]Question](4096)
	b := mustOK(t, "/compliance/sec/insider-trading-awareness-48213/attempt-2/ab/")
	if a.Page.HTML != b.Page.HTML {
		t.Fatal("same URL rendered differently")
	}
	c := firstCourse(t)
	if CertID(c, 1, "abc") != CertID(c, 1, "abc") || CertID(c, 1, "abc") == CertID(c, 2, "abc") {
		t.Fatal("cert IDs not deterministic per attempt")
	}
}

func TestAnswerLinksAllResolve(t *testing.T) {
	c := firstCourse(t)
	r := mustOK(t, c.AttemptURL(1))
	q := Questions(c, 1)[0]
	var answers []string
	for _, l := range links(r) {
		if strings.HasPrefix(l, c.AttemptURL(1)) && l != c.AttemptURL(1) {
			answers = append(answers, l)
		}
	}
	if len(answers) != len(q.Choices) {
		t.Fatalf("got %d answer links, want %d", len(answers), len(q.Choices))
	}
	for _, l := range answers {
		mustOK(t, l)
	}
}

func TestPassAndFail(t *testing.T) {
	c := firstCourse(t)
	pass := PassingAnswers(c, 1)
	res := mustOK(t, c.AttemptURL(1)+pass+"/")
	if len(res.Events) != 1 || res.Events[0].Kind != "completed" || !res.Events[0].Passed || res.Events[0].Score != 100 {
		t.Fatalf("passing results events: %+v", res.Events)
	}
	if strings.Contains(res.Page.HTML, `class="cmp-choice"`) {
		t.Fatal("results page still offers answers")
	}
	cert := mustOK(t, c.AttemptURL(1)+pass+"/certificate/")
	if !strings.Contains(cert.Page.HTML, CertID(c, 1, pass)) || !strings.Contains(cert.Page.HTML, "OpenAI GPTBot (Unverified Learner)") {
		t.Fatal("certificate missing cert ID or bot learner name")
	}
	if cert.TrapType != "compliance_certificate" || cert.Events[0].Kind != "certificate" {
		t.Fatalf("certificate trap/events: %s %+v", cert.TrapType, cert.Events)
	}
	pdf := mustOK(t, c.AttemptURL(1)+pass+"/certificate.pdf")
	if !bytes.HasPrefix(pdf.Body, []byte("%PDF-")) || !bytes.HasSuffix(pdf.Body, []byte("%%EOF\n")) || pdf.ContentType != "application/pdf" {
		t.Fatal("bad certificate PDF")
	}

	// All wrong: fail, and offer the next attempt.
	qs := Questions(c, 1)
	wrong := make([]byte, c.N)
	for i, q := range qs {
		wrong[i] = letters[(q.Correct+1)%len(q.Choices)]
	}
	fail := mustOK(t, c.AttemptURL(1)+string(wrong)+"/")
	if fail.Events[0].Passed || !strings.Contains(fail.Page.HTML, c.AttemptURL(2)) {
		t.Fatal("failing attempt should link to attempt 2")
	}
	if get(t, c.AttemptURL(1)+string(wrong)+"/certificate.pdf").Status != http.StatusNotFound {
		t.Fatal("failed attempt served a PDF")
	}
}

func TestLearnerCarriedThrough(t *testing.T) {
	c := firstCourse(t)
	r := mustOK(t, c.AttemptURL(1)+"?n=Ada+Lovelace&e=ada%40example.com&id=E123")
	for _, l := range links(r) {
		if strings.HasPrefix(l, c.AttemptURL(1)) && !strings.Contains(l, "n=Ada+Lovelace") {
			t.Fatalf("learner dropped from %s", l)
		}
	}
	var ident *Event
	for i := range r.Events {
		if r.Events[i].Kind == "identified" {
			ident = &r.Events[i]
		}
	}
	if ident == nil || ident.Course != c.ID || ident.Attempt != 1 || ident.Learner.Email != "ada@example.com" {
		t.Fatalf("identified event: %+v", r.Events)
	}
	pass := PassingAnswers(c, 1)
	cert := mustOK(t, c.AttemptURL(1)+pass+"/certificate/?n=Ada+Lovelace")
	if !strings.Contains(cert.Page.HTML, "Ada Lovelace") || strings.Contains(cert.Page.HTML, "Unverified Learner") {
		t.Fatal("certificate should carry the typed name")
	}
}

func TestNotFound(t *testing.T) {
	c := firstCourse(t)
	long := strings.Repeat("a", c.N+1)
	for _, u := range []string{
		"/compliance/sec/no-such-course-12345/",
		"/compliance/sec/insider-trading-awareness-1234/",
		"/compliance/not-an-agency/insider-trading-awareness-48213/",
		"/compliance/insider-trading-awareness-48213/",
		"/compliance/sec/insider-trading-awareness-48213/attempt-0/",
		"/compliance/sec/insider-trading-awareness-48213/attempt-01/",
		"/compliance/sec/insider-trading-awareness-48213/attempt-1/abz/",
		c.AttemptURL(1) + long + "/",
		c.AttemptURL(1) + "a/certificate/",
		"/compliance/not-an-agency/",
		"/compliance/sponsor/sec/",
		"/compliance/verify/sec/insider-trading-awareness-48213/xyz/",
		"/compliance/verify/sec/insider-trading-awareness-48213/<script>/",
		"/compliance/verify/sec/insider-trading-awareness-48213/0123ABCD/",
		"/compliance/verify/insider-trading-awareness-48213/0123abcd/",
	} {
		if s := get(t, u).Status; s != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", u, s)
		}
	}
	// A letter past the last choice of a two-choice question is invalid.
	r := newRnd("test_tf")
	for i := 0; i < 2000; i++ {
		c := randomCourse(r)
		qs := Questions(c, 1)
		if len(qs[0].Choices) == 2 {
			if s := get(t, c.AttemptURL(1)+"e/").Status; s != http.StatusNotFound {
				t.Fatalf("%s: letter e on a 2-choice question: %d", c.ID, s)
			}
			return
		}
	}
}

func TestReservedNamesAreNotAgencies(t *testing.T) {
	for _, w := range []string{"topic", "topics", "agencies", "verify", "transcript", "sitemap.xml"} {
		if _, ok := agencyMap[w]; ok {
			t.Fatalf("agency slug %q collides with a /compliance/ route", w)
		}
	}
}

func TestRedirects(t *testing.T) {
	if r := get(t, "/compliance"); r.Status != http.StatusMovedPermanently || r.Redirect != "/compliance/" {
		t.Fatalf("/compliance: %+v", r)
	}
	if r := get(t, "/compliance/sec/insider-trading-awareness-48213?n=x"); r.Status != http.StatusMovedPermanently || r.Redirect != "/compliance/sec/insider-trading-awareness-48213/?n=x" {
		t.Fatalf("slashless course: %+v", r)
	}
}

func TestIndexPages(t *testing.T) {
	c := firstCourse(t)
	for _, u := range []string{
		"/compliance/", "/compliance/?page=7", "/compliance/transcript/", "/compliance/transcript/?n=Ada",
		c.SponsorURL(), c.SponsorURL() + "?page=3", c.TopicURL(), c.URL(),
		verifyURL(c, "0123abcd"),
	} {
		r := mustOK(t, u)
		if len(links(r)) < 5 {
			t.Errorf("%s: only %d internal links", u, len(links(r)))
		}
	}
	// Agency pages list only that agency's courses, and every course URL
	// carries its agency.
	for _, l := range links(mustOK(t, c.SponsorURL())) {
		seg := strings.Split(strings.Trim(l, "/"), "/")
		if len(seg) == 3 {
			if o, ok := ParseCourse(seg[1], seg[2]); ok && o.Agency != c.Agency && strings.Contains(l, "/compliance/"+c.Agency.Slug+"/") {
				t.Fatalf("agency %s lists %s", c.Agency.Slug, o.ID)
			}
		}
	}
	if c.URL() != "/compliance/doc-pto/access-badge-renewal-attestation-48213/" {
		t.Fatalf("course URL %s", c.URL())
	}
	// Same type and number under another agency is a different course.
	o, _ := ParseCourse("sec", c.Slug)
	if o.ID == c.ID || o.Agency == c.Agency {
		t.Fatal("agency is not part of the course identity")
	}
	// Transcript certificate links are real, passing certificates.
	for _, l := range links(mustOK(t, "/compliance/transcript/")) {
		if strings.HasSuffix(l, "/certificate/") {
			r := mustOK(t, l)
			if len(r.Events) == 0 || !r.Events[0].Passed {
				t.Fatalf("transcript certificate %s is not a pass", l)
			}
		}
	}
}

func TestSitemap(t *testing.T) {
	idx := mustOK(t, "/compliance/sitemap.xml")
	if !bytes.Contains(idx.Body, []byte("<sitemapindex")) || bytes.Count(idx.Body, []byte("<loc>")) != sitemapCount {
		t.Fatal("bad sitemap index")
	}
	for _, n := range []string{"1", "200"} {
		r := mustOK(t, "/compliance/sitemap-"+n+".xml")
		if !bytes.HasPrefix(r.Body, []byte(`<?xml`)) || bytes.Count(r.Body, []byte("<loc>")) < 900 {
			t.Fatalf("sitemap-%s too small", n)
		}
	}
	for _, u := range []string{"/compliance/sitemap-0.xml", "/compliance/sitemap-201.xml", "/compliance/sitemap-01.xml"} {
		if get(t, u).Status != http.StatusNotFound {
			t.Fatalf("%s should 404", u)
		}
	}
}

func TestGeneratedAgencies(t *testing.T) {
	if agencyTotal() < 40000 {
		t.Fatalf("only %d agencies", agencyTotal())
	}
	for _, i := range []int{0, len(agencies) - 1, len(agencies), len(agencies) + 12345, agencyTotal() - 1} {
		a := agencyAt(i)
		b, ok := lookupAgency(a.Slug)
		if !ok || b != a {
			t.Fatalf("agency %d (%s) does not round-trip", i, a.Slug)
		}
	}
	a, ok := lookupAgency("wi-board-workplace-ergonomics")
	if !ok || a.Name != "Wisconsin Board of Workplace Ergonomics" {
		t.Fatalf("got %+v", a)
	}
	for _, bad := range []string{"wi-board", "zz-board-workplace-ergonomics", "wi-board-nope"} {
		if _, ok := lookupAgency(bad); ok {
			t.Fatalf("%s should not resolve", bad)
		}
	}
	mustOK(t, "/compliance/wi-board-workplace-ergonomics/")
	c, _ := ParseCourse("wi-board-workplace-ergonomics", "clean-desk-policy-certification-12345")
	mustOK(t, c.AttemptURL(1))
}

func TestHubPages(t *testing.T) {
	cat := mustOK(t, "/compliance/")
	for _, want := range []string{"Available Trainings", "Recommended for Your Role", "Elective Certifications", "New &amp; Updated", "Renew a Certificate", "All Topics", "/compliance/agencies/"} {
		if !strings.Contains(cat.Page.HTML, want) {
			t.Errorf("catalog missing %q", want)
		}
	}
	if n := len(links(cat)); n < 150 {
		t.Errorf("catalog has only %d links", n)
	}
	mustOK(t, "/compliance/topics/")
	mustOK(t, "/compliance/agencies/")
	last := fmt.Sprintf("/compliance/agencies/?page=%d", agencyPages())
	if strings.Contains(mustOK(t, last).Page.HTML, fmt.Sprintf("page=%d", agencyPages()+1)) {
		t.Error("last directory page links past the end")
	}
	if get(t, fmt.Sprintf("/compliance/agencies/?page=%d", agencyPages()+1)).Status != http.StatusNotFound {
		t.Error("page past the directory end should 404")
	}
}

func TestInstructorEmailShownAndRegistered(t *testing.T) {
	c := firstCourse(t)
	email := c.Instructor.Email()
	if !regexp.MustCompile(`^[a-z0-9]+(\.[a-z0-9]+)*\.[a-z0-9]+@acpwb\.com$`).MatchString(email) {
		t.Fatalf("instructor email %q", email)
	}
	pass := PassingAnswers(c, 1)
	for _, u := range []string{c.URL(), c.AttemptURL(1), c.AttemptURL(1) + "a/", c.AttemptURL(1) + pass + "/", c.AttemptURL(1) + pass + "/certificate/"} {
		r := mustOK(t, u)
		if !strings.Contains(r.Page.HTML, "mailto:"+email) && !strings.Contains(r.Page.HTML, email) {
			t.Errorf("%s: instructor email not shown", u)
		}
		// Registered against the course, not the (endless) answer path.
		if r.Page.EmailPath != c.URL() || !contains(r.Page.Emails, email) {
			t.Errorf("%s: emails %v registered at %q", u, r.Page.Emails, r.Page.EmailPath)
		}
	}
	cert := mustOK(t, c.AttemptURL(1)+pass+"/certificate/")
	if !contains(cert.Page.Emails, signatory().Email()) || !strings.Contains(cert.Page.HTML, signatory().Email()) {
		t.Fatal("certificate should show and register the CCO's address")
	}
}

func TestAnonymousCertificateNameIsBlank(t *testing.T) {
	c := firstCourse(t)
	pass := PassingAnswers(c, 1)
	r := Serve(Request{Path: c.AttemptURL(1) + pass + "/certificate/", Query: url.Values{}, Now: testNow})
	if r.Status != http.StatusOK || !strings.Contains(r.Page.HTML, `class="cmp-cert-name"></div>`) {
		t.Fatal("anonymous certificate should have an empty name line")
	}
	if strings.Contains(r.Page.HTML, "Valued Learner") || strings.Contains(r.Page.HTML, "Unverified Learner") {
		t.Fatal("anonymous certificate should not carry a placeholder name")
	}
}

// Every page of a course shares the course's cover as its og:image, so a
// shared question, result or certificate previews like the course.
func TestCoursePagesShareTheHero(t *testing.T) {
	c := firstCourse(t)
	pass := PassingAnswers(c, 1)
	for _, p := range []string{c.URL(), c.AttemptURL(1), c.AttemptURL(1) + pass[:2] + "/",
		c.AttemptURL(1) + pass + "/", c.AttemptURL(1) + pass + "/certificate/"} {
		r := Serve(Request{Path: p, Query: url.Values{}, Now: testNow})
		if r.Status != http.StatusOK || r.Page.Image != c.Hero || r.Page.ImageW == 0 {
			t.Fatalf("%s: share image %q (%dx%d), want %q", p, r.Page.Image, r.Page.ImageW, r.Page.ImageH, c.Hero)
		}
	}
	r := Serve(Request{Path: c.AttemptURL(1) + pass + "/certificate/", Query: url.Values{"n": {"Pat Doe"}}, Now: testNow})
	if !strings.HasPrefix(r.Page.Title, "Pat Doe earned a Certificate of Completion") {
		t.Fatalf("certificate share title %q should name the learner", r.Page.Title)
	}
}

// Every grammar line (banks, fragments) expands cleanly for many seeds: no
// unknown slot or fragment, no stray bracket, and no "a"/"an" mismatch.
func TestGrammarExpandsCleanly(t *testing.T) {
	bad := regexp.MustCompile(`[\[\]{}|]|\b[Aa] [aeio][a-z]|\b[Aa]n [bcdfgjklmnpqrstvwxyz][a-z]|[a-z]\(|\b(?i:a|an|the|actual|simulated) (?i:a|an|the)\b`)
	c := firstCourse(t)
	var srcs []string
	for _, qs := range banks {
		for _, q := range qs {
			srcs = append(srcs, q.Text)
			srcs = append(srcs, q.Right...)
			srcs = append(srcs, q.Wrong...)
		}
	}
	for name := range fragments {
		srcs = append(srcs, "{@"+name+"}")
	}
	for i := 0; i < 60; i++ {
		r := newRnd(fmt.Sprintf("grammar_%d", i))
		g := gen{r, newSlots(r, c)}
		for _, s := range srcs {
			if out := g.expand(s); bad.MatchString(out) {
				t.Fatalf("%q expanded to %q", s, out)
			}
		}
	}
}

func TestOutlinesVary(t *testing.T) {
	r := newRnd("outline_variety")
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		c := randomCourse(r)
		if len(c.Modules) < 6 || len(c.Rules) < 3 || len(c.Requirements) < 6 {
			t.Fatalf("%s: %d modules, %d rules, %d requirements", c.ID, len(c.Modules), len(c.Rules), len(c.Requirements))
		}
		for _, m := range c.Modules {
			seen[m.Title] = true
		}
	}
	if len(seen) < 150 {
		t.Fatalf("only %d distinct module titles across 50 courses", len(seen))
	}
}

func TestLowerFirst(t *testing.T) {
	for in, want := range map[string]string{"A warning sign": "a warning sign", "ACPWB's channel": "ACPWB's channel", "The process": "the process"} {
		if got := lowerFirst(in); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGrammarSpacing(t *testing.T) {
	g := gen{newRnd("spacing"), slots{"X": "x"}}
	for i := 0; i < 50; i++ {
		out := g.expand("Global Considerations[? for Associates] and [?A|B|C] then [one|two[? more]] end")
		if strings.Contains(out, "Considerationsfor") || strings.Contains(out, "|") || strings.Contains(out, "  ") {
			t.Fatalf("bad spacing: %q", out)
		}
	}
}

func TestExhibitsWellFormed(t *testing.T) {
	c, _ := ParseCourse("sec", "access-badge-renewal-attestation-12345")
	if c == nil {
		t.Skip("sample course missing")
	}
	for _, kind := range exhibitNames {
		for i := 0; i < 300; i++ {
			r := newRnd(fmt.Sprintf("exhibit_test_%s_%d", kind, i))
			q := exhibitQ(r, c, newSlots(r, c), 3, kind)
			if len(q.Choices) < 2 || q.Correct < 0 || q.Correct >= len(q.Choices) {
				t.Fatalf("%s #%d: bad choices %v / %d", kind, i, q.Choices, q.Correct)
			}
			seen := map[string]bool{}
			for _, ch := range q.Choices {
				if seen[ch] || ch == "" {
					t.Fatalf("%s #%d: duplicate or empty choice %q in %v", kind, i, ch, q.Choices)
				}
				seen[ch] = true
			}
			for _, s := range []string{q.Text, q.Exhibit, strings.Join(q.Choices, "|")} {
				if strings.ContainsAny(s, "{}") || strings.Contains(s, "%!") || strings.Contains(s, "\x00") {
					t.Fatalf("%s #%d: unfilled or bad text: %.300s", kind, i, s)
				}
			}
			if !strings.Contains(q.Text, q.ExLabel) || q.Exhibit == "" {
				t.Fatalf("%s #%d: exhibit missing or not referenced: %s", kind, i, q.Text)
			}
		}
	}
}

func TestEvacAnswerIsNearestOpenExit(t *testing.T) {
	for i := 0; i < 500; i++ {
		r := newRnd(fmt.Sprintf("evac_%d", i))
		p := buildFloorPlan(r, slots{"FLOOR": "33"})
		best := -1
		for j, e := range p.exits {
			if j == p.blocked {
				continue
			}
			if best < 0 || math.Hypot(e.x-p.youX, e.y-p.youY) < math.Hypot(p.exits[best].x-p.youX, p.exits[best].y-p.youY) {
				best = j
			}
		}
		if best != p.answer || p.answer == p.blocked {
			t.Fatalf("plan %d: answer %d, nearest open %d, blocked %d", i, p.answer, best, p.blocked)
		}
	}
}

func TestExhibitsAppearInCourses(t *testing.T) {
	n := 0
	for i := 0; i < 40; i++ {
		c, _ := ParseCourse("sec", fmt.Sprintf("annual-security-awareness-%05d", 10000+i))
		for _, q := range Questions(c, 1) {
			if strings.HasPrefix(q.key, "ex/") {
				n++
			}
		}
	}
	if n < 20 {
		t.Fatalf("only %d exhibit questions across 40 courses", n)
	}
}

func TestChartMath(t *testing.T) {
	c, _ := ParseCourse("sec", "access-badge-renewal-attestation-12345")
	for i := 0; i < 300; i++ {
		r := newRnd(fmt.Sprintf("chartmath_%d", i))
		q := chartMathQ(r, c, 2, "lost badges", 2024, []string{"Q1", "Q2", "Q3", "Q4"}, []int{r.between(4, 60), r.between(4, 60), r.between(4, 60), r.between(4, 60)})
		seen := map[string]bool{}
		for _, ch := range q.Choices {
			if seen[ch] {
				t.Fatalf("#%d duplicate choice in %v", i, q.Choices)
			}
			seen[ch] = true
		}
	}
}

func TestNearestAED(t *testing.T) {
	for i := 0; i < 500; i++ {
		p := buildFloorPlan(newRnd(fmt.Sprintf("aed_%d", i)), slots{"FLOOR": "33rd"})
		if p.nearestAED < 0 {
			continue
		}
		for j, a := range p.aeds {
			if j != p.nearestAED && math.Hypot(a[0]-p.youX, a[1]-p.youY) <= math.Hypot(p.aeds[p.nearestAED][0]-p.youX, p.aeds[p.nearestAED][1]-p.youY) {
				t.Fatalf("plan %d: AED %d is not the nearest", i, p.nearestAED+1)
			}
		}
	}
}

func TestHybridGridSingleViolator(t *testing.T) {
	c, _ := ParseCourse("sec", "hybrid-work-agreement-attestation-12345")
	re := regexp.MustCompile(`at least (\d) office days per week\. (\w+) is the team anchor day`)
	days := map[string]int{"Monday": 0, "Tuesday": 1, "Wednesday": 2, "Thursday": 3, "Friday": 4}
	for i := 0; i < 400; i++ {
		r := newRnd(fmt.Sprintf("hg_%d", i))
		q := exhibitQ(r, c, newSlots(r, c), 1, "hybridgrid")
		m := re.FindStringSubmatch(q.Exhibit)
		if m == nil {
			t.Fatalf("#%d: no policy note", i)
		}
		min := int(m[1][0] - '0')
		anchor := days[m[2]]
		var under, missed []string
		for _, row := range regexp.MustCompile(`<tr><td>([^<]+)</td>(.*?)</tr>`).FindAllStringSubmatch(q.Exhibit, -1) {
			cells := regexp.MustCompile(`<td[^>]*>([^<]*)</td>`).FindAllStringSubmatch(row[2], -1)
			n, pto := 0, false
			for _, c := range cells {
				n += strings.Count(c[1], "Office")
				pto = pto || c[1] == "PTO"
			}
			if n < min && !pto {
				under = append(under, row[1])
			}
			if cells[anchor][1] == "Remote" {
				missed = append(missed, row[1])
			}
		}
		right := q.Choices[q.Correct]
		switch {
		case strings.Contains(q.Text, "minimum"):
			if len(under) != 1 || under[0] != right {
				t.Fatalf("#%d: under-minimum %v, keyed %q", i, under, right)
			}
		case strings.Contains(q.Text, "anchor"):
			if len(missed) != 1 || missed[0] != right {
				t.Fatalf("#%d: missed anchor %v, keyed %q", i, missed, right)
			}
		}
	}
}

func TestRulesAndModulesUseDistinctTemplates(t *testing.T) {
	norm := regexp.MustCompile(`\W+`)
	for i := 0; i < 300; i++ {
		c, _ := ParseCourse("sec", fmt.Sprintf("%s-%05d", courseTypes[i%len(courseTypes)].Slug, 10000+i))
		// The screenshot case: one rule a prefix of another ("…carefully." vs "…carefully and then read it again.").
		for a := range c.Rules {
			for b := range c.Rules {
				ra := norm.ReplaceAllString(strings.ToLower(c.Rules[a]), " ")
				rb := norm.ReplaceAllString(strings.ToLower(c.Rules[b]), " ")
				if a != b && len(ra) > 20 && strings.HasPrefix(rb, strings.TrimSpace(ra)) {
					t.Fatalf("%s: rules repeat each other:\n  %s\n  %s", c.ID, c.Rules[a], c.Rules[b])
				}
			}
		}
	}
}

func TestTopicCoversMatchCourseTypes(t *testing.T) {
	for slug := range topicCovers {
		if _, ok := courseTypeMap[slug]; !ok {
			t.Errorf("compliance_covers.txt names unknown course type %q (renamed or removed?)", slug)
		}
	}
}
