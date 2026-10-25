// Package compliance serves /compliance/, the ACPWB ComplianceCenter: an
// endless fake corporate LMS whose courses, questions, answer keys and
// certificates are all generated from the URL.
//
// The test state lives in the path:
// /compliance/<agency>/<course>/attempt-<k>/<answers>/
// where <answers> holds one letter per question answered so far. Every
// choice is a plain link that appends a letter, so a crawler "takes" the
// test by following links and, with enough branches, earns a watermarked
// certificate. Nothing here has a Python counterpart.
package compliance

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Learner is the optional identity a test-taker typed above question 1. It
// rides along on every answer link as ?n=&e=&id=.
type Learner struct{ Name, Email, EmpID string }

func learnerFrom(q url.Values) Learner {
	clean := func(s string) string {
		s = strings.TrimSpace(s)
		if !utf8.ValidString(s) {
			s = strings.ToValidUTF8(s, "")
		}
		if r := []rune(s); len(r) > 80 {
			s = string(r[:80])
		}
		return s
	}
	return Learner{clean(q.Get("n")), clean(q.Get("e")), clean(q.Get("id"))}
}

func (l Learner) Empty() bool { return l.Name == "" && l.Email == "" && l.EmpID == "" }

// Query is the learner's query-string suffix ("" or "?n=…").
func (l Learner) Query() string {
	if l.Empty() {
		return ""
	}
	v := url.Values{}
	if l.Name != "" {
		v.Set("n", l.Name)
	}
	if l.Email != "" {
		v.Set("e", l.Email)
	}
	if l.EmpID != "" {
		v.Set("id", l.EmpID)
	}
	return "?" + v.Encode()
}

// Request is what Serve needs from an HTTP request.
type Request struct {
	Path  string
	Query url.Values
	Now   time.Time
	// Fallback names an anonymous learner on a certificate, e.g. the
	// classified crawler ("OpenAI GPTBot").
	Fallback string
}

// Event is a milestone worth recording (see visitqueue.PushComplianceCompletion).
type Event struct {
	Kind    string // completed / certificate / certificate_pdf / identified
	Course  string
	Attempt int
	Answers string
	Score   int
	Passed  bool
	CertID  string
	Learner Learner
}

// Page is the HTML body (inside the site shell) plus SEO fields, and the
// @acpwb.com addresses it shows. Those are registered for spam matching
// against EmailPath, the course URL rather than the page's own path: every
// answer path is a new page, and one row per (address, course) is enough to
// trace a harvested address back.
type Page struct {
	Title, Description, HTML string
	// Image is the share (og:image) picture, site-relative, with its pixel
	// size; empty keeps the section default.
	Image          string
	ImageW, ImageH int
	Emails         []string
	EmailPath      string
}

type Response struct {
	Status      int
	Redirect    string
	Page        *Page  // HTML page in the main-site shell
	ContentType string // non-HTML bodies (PDF, XML)
	Body        []byte
	Filename    string // Content-Disposition for downloads
	TrapType    string // compliance / compliance_certificate
	Events      []Event
	Stat        Stat // what kind of page this was, for the dashboard counters
}

// Stat describes a served page for the per-day dashboard counters
// (visitqueue.PushComplianceStats): page kind, topic and agency, how deep
// into an attempt it is, and which exhibit (if any) the question shows.
type Stat struct {
	Kind    string // catalog, agencies, topics, topic, sponsor, course, question, results, certificate, certificate_pdf, verify, transcript, sitemap
	Topic   string // course-type slug
	Agency  string // agency slug
	Course  string // course ID
	Attempt int
	Answers string // answer letters so far (depth = len)
	Exhibit string // exhibit kind on a question page ("evac", "chart", …)
}

func notFound() Response { return Response{Status: http.StatusNotFound} }

// Serve renders any /compliance/ URL.
func Serve(req Request) Response {
	rest := strings.TrimPrefix(req.Path, "/compliance")
	if rest == req.Path {
		return notFound()
	}
	if rest == "" {
		return Response{Status: http.StatusMovedPermanently, Redirect: "/compliance/"}
	}
	if req.Now.IsZero() {
		req.Now = time.Now()
	}
	learner := learnerFrom(req.Query)
	seg := strings.Split(strings.Trim(rest, "/"), "/")
	resp := route(req, learner, seg, strings.HasSuffix(rest, "/"))
	if resp.Status == http.StatusOK {
		resp.Stat = statFor(seg)
	}
	if resp.Status == http.StatusOK && !learner.Empty() {
		ev := Event{Kind: "identified", Learner: learner}
		for _, e := range resp.Events {
			ev.Course, ev.Attempt = e.Course, e.Attempt
		}
		if ev.Course == "" {
			ev.Course, ev.Attempt = courseAttemptOf(req.Path)
		}
		resp.Events = append(resp.Events, ev)
	}
	return resp
}

func courseAttemptOf(path string) (string, int) {
	seg := strings.Split(strings.Trim(strings.TrimPrefix(path, "/compliance"), "/"), "/")
	if len(seg) < 2 {
		return "", 0
	}
	c, ok := ParseCourse(seg[0], seg[1])
	if !ok {
		return "", 0
	}
	if len(seg) > 2 {
		if k, ok := parseAttempt(seg[2]); ok {
			return c.ID, k
		}
	}
	return c.ID, 0
}

func parseAttempt(s string) (int, bool) {
	n, ok := strings.CutPrefix(s, "attempt-")
	if !ok || n == "" || n[0] == '0' || len(n) > 4 {
		return 0, false
	}
	k, err := strconv.Atoi(n)
	return k, err == nil && k >= 1 && k <= maxAttempt
}

func validAnswers(s string) bool {
	if s == "" || len(s) > maxQuestions {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !strings.ContainsRune(letters, rune(s[i])) {
			return false
		}
	}
	return true
}

func pageParam(q url.Values) int {
	p, err := strconv.Atoi(q.Get("page"))
	if err != nil || p < 1 {
		return 1
	}
	if p > 1_000_000 {
		return 1_000_000
	}
	return p
}

func ok(p Page) Response { return Response{Status: http.StatusOK, Page: &p, TrapType: "compliance"} }

func route(req Request, l Learner, seg []string, slash bool) Response {
	// Directory-style pages need their trailing slash, as on the rest of
	// the site.
	dir := func(r Response) Response {
		if !slash && r.Status == http.StatusOK && r.Page != nil {
			loc := req.Path + "/"
			if raw := req.Query.Encode(); raw != "" {
				loc += "?" + raw
			}
			return Response{Status: http.StatusMovedPermanently, Redirect: loc}
		}
		return r
	}
	switch {
	case len(seg) == 1 && seg[0] == "":
		return ok(renderCatalog(pageParam(req.Query), req.Now, l))
	case len(seg) == 1 && seg[0] == "sitemap.xml":
		return Response{Status: http.StatusOK, ContentType: "application/xml; charset=utf-8", Body: sitemapIndexXML(), TrapType: "compliance"}
	case len(seg) == 1 && strings.HasPrefix(seg[0], "sitemap-") && strings.HasSuffix(seg[0], ".xml"):
		n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(seg[0], "sitemap-"), ".xml"))
		if err != nil || n < 1 || n > sitemapCount || seg[0] != fmt.Sprintf("sitemap-%d.xml", n) {
			return notFound()
		}
		return Response{Status: http.StatusOK, ContentType: "application/xml; charset=utf-8", Body: sitemapXML(n), TrapType: "compliance"}
	case len(seg) == 1 && seg[0] == "agencies":
		p := pageParam(req.Query)
		if p > agencyPages() {
			return notFound()
		}
		return dir(ok(renderAgencies(p, l)))
	case len(seg) == 1 && seg[0] == "topics":
		return dir(ok(renderTopics(l)))
	case len(seg) == 1 && seg[0] == "transcript":
		return dir(ok(renderTranscript(req.Now, l)))
	case len(seg) == 2 && seg[0] == "topic":
		ct, found := courseTypeMap[seg[1]]
		if !found {
			return notFound()
		}
		return dir(ok(renderTopic(ct, pageParam(req.Query), req.Now, l)))
	case len(seg) == 4 && seg[0] == "verify":
		c, found := ParseCourse(seg[1], seg[2])
		if !found || !isHex8(seg[3]) {
			return notFound()
		}
		return dir(ok(renderVerify(c, seg[3])))
	}

	// Everything else is under an agency: /compliance/<agency>/ lists its
	// courses, /compliance/<agency>/<course>/… is a course.
	a, found := lookupAgency(seg[0])
	if !found {
		return notFound()
	}
	if len(seg) == 1 {
		return dir(ok(renderSponsor(a, pageParam(req.Query), req.Now, l)))
	}
	c, found := ParseCourse(seg[0], seg[1])
	if !found {
		return notFound()
	}
	seg = seg[1:]
	if len(seg) == 1 {
		return dir(ok(renderCourse(c, req.Now, l)))
	}
	k, found := parseAttempt(seg[1])
	if !found {
		return notFound()
	}
	if len(seg) == 2 {
		return dir(ok(renderQuestion(c, k, "", l)))
	}
	answers := seg[2]
	if !validAnswers(answers) {
		return notFound()
	}
	res, valid := Score(c, k, answers)
	if !valid {
		return notFound()
	}
	ev := Event{Course: c.ID, Attempt: k, Answers: answers, Score: res.Percent, Passed: res.Passed, Learner: l}
	if res.Passed {
		ev.CertID = CertID(c, k, answers)
	}
	switch {
	case len(seg) == 3 && !res.Complete:
		return dir(ok(renderQuestion(c, k, answers, l)))
	case len(seg) == 3:
		ev.Kind = "completed"
		r := dir(ok(renderResults(res, l)))
		if r.Status == http.StatusOK {
			r.Events = []Event{ev}
		}
		return r
	case len(seg) == 4 && seg[3] == "certificate" && res.Complete:
		r := dir(ok(renderCertificate(res, l, req.Fallback)))
		r.TrapType = "compliance_certificate"
		if r.Status == http.StatusOK && res.Passed {
			ev.Kind = "certificate"
			r.Events = []Event{ev}
		}
		return r
	case len(seg) == 4 && seg[3] == "certificate.pdf" && res.Passed:
		ev.Kind = "certificate_pdf"
		return Response{
			Status: http.StatusOK, ContentType: "application/pdf",
			Body:     certificatePDF(res, learnerName(l, req.Fallback)),
			Filename: "ACPWB-Certificate-" + ev.CertID + ".pdf",
			TrapType: "compliance_certificate", Events: []Event{ev},
		}
	}
	return notFound()
}

func isHex8(s string) bool {
	if len(s) != 8 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !strings.ContainsRune("0123456789abcdef", rune(s[i])) {
			return false
		}
	}
	return true
}

// learnerName is the name printed on a certificate.
func learnerName(l Learner, fallback string) string {
	if l.Name != "" {
		return l.Name
	}
	if fallback != "" {
		return fallback + " (Unverified Learner)"
	}
	return "" // blank name line, as on an unsigned paper certificate
}

// statFor classifies a served path (already known to be valid).
func statFor(seg []string) Stat {
	switch {
	case len(seg) == 1 && seg[0] == "":
		return Stat{Kind: "catalog"}
	case len(seg) == 1 && strings.HasPrefix(seg[0], "sitemap"):
		return Stat{Kind: "sitemap"}
	case len(seg) == 1 && (seg[0] == "agencies" || seg[0] == "topics" || seg[0] == "transcript"):
		return Stat{Kind: seg[0]}
	case seg[0] == "topic":
		return Stat{Kind: "topic", Topic: seg[1]}
	case seg[0] == "verify":
		st := Stat{Kind: "verify", Agency: seg[1]}
		if c, ok := ParseCourse(seg[1], seg[2]); ok {
			st.Topic, st.Course = c.Type.Slug, c.ID
		}
		return st
	case len(seg) == 1:
		return Stat{Kind: "sponsor", Agency: seg[0]}
	}
	c, ok := ParseCourse(seg[0], seg[1])
	if !ok {
		return Stat{}
	}
	st := Stat{Kind: "course", Topic: c.Type.Slug, Agency: seg[0], Course: c.ID}
	if len(seg) < 3 {
		return st
	}
	st.Attempt, _ = parseAttempt(seg[2])
	if len(seg) > 3 {
		st.Answers = seg[3]
	}
	switch {
	case len(seg) > 4 && seg[4] == "certificate.pdf":
		st.Kind = "certificate_pdf"
	case len(seg) > 4:
		st.Kind = "certificate"
	case len(st.Answers) >= c.N:
		st.Kind = "results"
	default:
		st.Kind = "question"
		qs := Questions(c, st.Attempt)
		if i := len(st.Answers); i < len(qs) {
			if k := qs[i].key; strings.HasPrefix(k, "ex/") {
				st.Exhibit = strings.SplitN(k[3:], "/", 2)[0]
			} else if strings.HasPrefix(k, "chart") || strings.HasPrefix(k, "badgeex") {
				st.Exhibit = strings.SplitN(k, "/", 2)[0]
			}
		}
	}
	return st
}
