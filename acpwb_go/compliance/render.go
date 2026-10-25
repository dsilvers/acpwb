package compliance

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var esc = html.EscapeString

// js is a JSON string literal, safe inside <script> (json.Marshal escapes <, > and &).
func js(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

const cmpCSS = `<style>
.cmp :where(a){color:var(--navy);text-decoration:none;transition:color .15s}
.cmp-row{position:relative;cursor:pointer;transition:background .12s}
.cmp-row:hover{background:#faf6ea}
.cmp-row-link::after{content:"";position:absolute;inset:0}
.cmp-row-keep{position:relative;z-index:1}
.cmp-row:hover .cmp-row-link{color:var(--gold)}
.cmp :where(a:hover){color:var(--gold);text-decoration:underline;text-underline-offset:2px}
.cmp{background:#f6f1e4;padding:0 0 3rem;font-size:.9rem}
.cmp-bar{background:#2f3b52;color:#e8e3d3;font-size:.72rem;padding:.4rem 0;letter-spacing:.02em}
.cmp-bar .container{display:flex;justify-content:space-between;flex-wrap:wrap;gap:.5rem}
.cmp-bar b{color:#fff}
.cmp-bar a{color:#f2d98d;text-decoration:none}
.cmp-brand{background:#fff;border-bottom:3px solid var(--gold);padding:.9rem 0;margin-bottom:1.5rem}
.cmp-brand .container{display:flex;align-items:center;gap:.9rem;flex-wrap:wrap}
.cmp-logo{font-weight:800;color:var(--navy);font-size:1.15rem;letter-spacing:-.01em;text-decoration:none}
.cmp-logo span{color:var(--gold)}
.cmp-crumbs{font-size:.75rem;color:var(--muted);margin-left:auto}
.cmp-crumbs a{color:var(--navy);text-decoration:none}
.cmp-banner{background:#fff4cf;border:1px solid #e4c96b;color:#5a4210;padding:.55rem .9rem;font-size:.78rem;margin-bottom:1.25rem}
.cmp-card{background:#fff;border:1px solid #ddd5c0;box-shadow:0 1px 0 #e9e2cf;padding:1.25rem 1.4rem;margin-bottom:1.25rem}
.cmp-card h1{font-size:1.45rem;font-weight:800;color:var(--navy);margin:0 0 .35rem}
.cmp-card h2{font-size:1rem;font-weight:800;color:var(--navy);margin:0 0 .75rem;text-transform:uppercase;letter-spacing:.05em}
.cmp-k{font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.12em;color:var(--gold)}
.cmp-muted{color:var(--muted);font-size:.78rem}
.cmp-table{width:100%;border-collapse:collapse;font-size:.82rem}
.cmp-table th{background:#2f3b52;color:#fff;font-size:.65rem;text-transform:uppercase;letter-spacing:.08em;padding:.5rem .6rem;text-align:left}
.cmp-table td{border-top:1px solid #e6dfcc;padding:.55rem .6rem;vertical-align:middle}
.cmp-table td a{color:var(--navy);font-weight:700;text-decoration:none}
.cmp-table td a:hover{text-decoration:underline}
.cmp-table-wrap{overflow-x:auto}
.cmp-st{display:inline-block;font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;padding:.15rem .45rem;white-space:nowrap}
.cmp-st-soon{background:#e4a11b;color:#fff}
.cmp-st-ok{background:#e7eef7;color:#2b3a7a}
.cmp-st-done{background:#1f5f3a;color:#fff}
.cmp-st-prog{background:#dfe3ea;color:#333}
.cmp-seal{flex:none;vertical-align:middle}
.cmp-side ul{list-style:none;padding:0;margin:0;font-size:.8rem}
.cmp-side li{padding:.3rem 0;border-top:1px solid #eee6d2}
.cmp-side li:first-child{border-top:0}
.cmp-side a{color:var(--navy);text-decoration:none}
.cmp-side a:hover{text-decoration:underline}
.cmp-btn{display:inline-block;background:var(--navy);color:var(--gold);font-weight:800;text-transform:uppercase;letter-spacing:.06em;font-size:.78rem;padding:.6rem 1.1rem;text-decoration:none;border:1px solid var(--navy);box-shadow:3px 3px 0 var(--gold)}
.cmp-btn:hover{background:var(--gold);color:var(--navy)}
.cmp-btn+.cmp-btn{margin-left:.5rem}
.cmp-btn-alt{background:#fff;color:var(--navy)}
.cmp-dl{display:grid;grid-template-columns:auto 1fr;gap:.35rem 1rem;font-size:.82rem;margin:0}
.cmp-dl dt{font-weight:700;color:var(--muted)}
.cmp-dl dd{margin:0;color:var(--navy)}
.cmp-hero{width:100%;max-height:220px;object-fit:cover;display:block;margin-bottom:1rem;border:1px solid #ddd5c0}
.cmp-instr{display:flex;gap:.8rem;align-items:center}
.cmp-instr img{width:64px;height:64px;object-fit:cover;border:2px solid var(--gold)}
.cmp-instr>div{min-width:0;overflow-wrap:anywhere}
.cmp-prog{height:12px;background:#e6dfcc;border:1px solid #cfc6ad;margin:.4rem 0 1rem;position:relative}
.cmp-prog i{position:absolute;left:0;top:0;bottom:0;background:linear-gradient(90deg,#2f3b52,#4c5e82)}
.cmp-q-n{font-size:.75rem;font-weight:800;color:var(--muted);text-transform:uppercase;letter-spacing:.08em}
.cmp-q-prompt{font-size:.72rem;font-weight:700;color:var(--gold);text-transform:uppercase;letter-spacing:.1em;margin:.4rem 0}
.cmp-q-text{font-size:1.08rem;font-weight:600;color:var(--navy);line-height:1.5;margin-bottom:1rem}
.cmp-exhibit{border:1px dashed #cfc6ad;background:#fbf8ef;padding:.75rem;margin:0 0 1rem;text-align:center}
.cmp-exhibit figcaption{font-size:.7rem;font-weight:800;color:var(--muted);text-transform:uppercase;letter-spacing:.1em;margin-top:.4rem}
.cmp-exhibit-img{width:100%;max-width:100%;height:auto}
.cmp-ex-doc{background:#fff;border:1px solid #c8ccd4;text-align:left;margin:0 auto .6rem;max-width:720px}
.cmp-ex-head{display:flex;justify-content:space-between;gap:1rem;flex-wrap:wrap;background:#0a1628;color:#fff;padding:.4rem .7rem;font-size:.78rem}
.cmp-ex-head span{opacity:.75}
.cmp-ex-table{width:100%;border-collapse:collapse;font-size:.74rem;font-variant-numeric:tabular-nums}
.cmp-ex-table th{background:#eef0f4;text-align:left;padding:.3rem .5rem;border-bottom:1px solid #c8ccd4;white-space:nowrap}
.cmp-ex-table td{padding:.3rem .5rem;border-bottom:1px solid #eceef2;min-width:1.2rem}
.cmp-ex-table td.cmp-ex-busy{background:#dbe3f0;color:#1b2a4a;font-size:.66rem}
.cmp-ex-note{font-size:.7rem;color:#555;padding:.4rem .7rem;border-top:1px solid #eceef2;font-style:italic}
.cmp-ex-cover{position:relative;display:inline-block;max-width:300px;background:#fff;border:1px solid #c8ccd4;box-shadow:0 2px 6px rgba(0,0,0,.12)}
.cmp-ex-cover img{display:block;width:100%;height:auto}
.cmp-ex-covercap{padding:.45rem .6rem;text-align:left}.cmp-ex-covercap b{display:block;font-size:.82rem}.cmp-ex-covercap small{color:#666;font-size:.66rem}
.cmp-ex-stamp{position:absolute;top:28%;left:50%;transform:translate(-50%,-50%) rotate(-12deg);border:3px solid;padding:.15rem .6rem;font:800 1.15rem/1.2 Arial,sans-serif;letter-spacing:.12em;background:rgba(255,255,255,.82)}
.cmp-ex-stamp-0{color:#1f7a3a}.cmp-ex-stamp-1{color:#1b2a4a}.cmp-ex-stamp-2{color:#c62828}.cmp-ex-stamp-3{color:#6a1b9a}
.cmp-ex-mail{background:#fff;border:1px solid #c8ccd4;text-align:left;max-width:600px;margin:0 auto;font-size:.8rem}
.cmp-ex-mailbar{background:#e8ecf2;padding:.35rem .7rem;font-size:.7rem;color:#555}
.cmp-ex-mail dl{display:grid;grid-template-columns:auto 1fr;gap:.15rem .7rem;margin:0;padding:.5rem .7rem;border-bottom:1px solid #eceef2}
.cmp-ex-mail dt{color:#777;font-weight:400}.cmp-ex-mail dd{margin:0;word-break:break-all}
.cmp-ex-mailbody{padding:.6rem .7rem}.cmp-ex-mailbody p{margin:0 0 .5rem}.cmp-ex-mailbody u{color:#1a55c4}
.cmp-ex-banner{padding:.3rem .7rem;font-size:.72rem;border-bottom:1px solid #eceef2}.cmp-ex-ok{color:#1f7a3a}.cmp-ex-warn{color:#8a5a00;background:#fff4cf;padding:.1rem .35rem}
.cmp-ex-hover{display:inline-block;margin-top:.2rem;background:#333;color:#fff;font-size:.66rem;padding:.1rem .35rem;border-radius:3px;word-break:break-all}
.cmp-choices{list-style:none;padding:0;margin:0}
.cmp-choices li{margin-bottom:.5rem}
.cmp-choice{display:flex;gap:.75rem;align-items:center;background:#fff;border:1px solid #cfc6ad;padding:.7rem .9rem;color:var(--navy);text-decoration:none;font-weight:500;line-height:1.4}
.cmp-choice:hover{background:#2f3b52;color:#fff;border-color:#2f3b52}
.cmp-choice b{flex:none;display:inline-flex;align-items:center;justify-content:center;width:1.6rem;height:1.6rem;border:2px solid currentColor;border-radius:50%;font-size:.75rem}
.cmp-ref{font-size:.7rem;color:var(--muted);margin-top:1rem;font-style:italic}
.cmp-form{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:.6rem;align-items:end}
.cmp-form label{font-size:.7rem;font-weight:700;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;display:block}
.cmp-form input{width:100%;border:1px solid #cfc6ad;padding:.4rem .5rem;font-size:.85rem}
.cmp-form button{background:#2f3b52;color:#fff;border:0;padding:.5rem .8rem;font-size:.75rem;font-weight:700;text-transform:uppercase}
.cmp-score{font-size:3rem;font-weight:800;line-height:1}
.cmp-pass{color:#1f5f3a}
.cmp-fail{color:#8b1e1e}
.cmp-ok{color:#1f5f3a;font-weight:800}
.cmp-no{color:#8b1e1e;font-weight:800}
.cmp-cert{position:relative;background:#fffdf6;border:10px double #0a1628;outline:2px solid #c9a84c;outline-offset:-18px;padding:3rem 2.5rem;text-align:center;overflow:hidden;margin-bottom:1.25rem}
.cmp-cert-rosette{position:absolute;width:520px;height:520px;left:50%;top:50%;transform:translate(-50%,-50%);pointer-events:none}
.cmp-cert>*:not(.cmp-cert-rosette){position:relative}
.cmp-cert-k{font-family:Georgia,serif;font-size:.8rem;letter-spacing:.3em;text-transform:uppercase;color:#6b4c12}
.cmp-cert h1{font-family:Georgia,serif;font-size:2.4rem;color:#0a1628;margin:.4rem 0 1rem;font-weight:700}
.cmp-cert-name{font-family:Georgia,serif;font-style:italic;font-size:2rem;color:#0a1628;border-bottom:1px solid #c9a84c;display:inline-block;padding:0 2rem .3rem;margin:.4rem 0 1rem;min-width:min(420px,80%);min-height:3.2rem}
.cmp-cert-course{font-size:1.15rem;font-weight:800;color:#0a1628}
.cmp-cert-meta{font-size:.78rem;color:#555;margin-top:.6rem}
.cmp-cert-sigs{display:flex;justify-content:space-around;gap:2rem;margin-top:2rem;flex-wrap:wrap}
.cmp-cert-sigs div{min-width:200px;font-size:.72rem;color:#333}
.cmp-sig{width:200px;height:56px;display:block;margin:0 auto;border-bottom:1px solid #333}
.cmp-foil{display:block;margin:1.2rem auto 0;width:96px;height:125px}
.cmp-cert-seal{margin:0 auto 1rem;text-align:center}
.cmp-cert-no{font-family:"IBM Plex Mono",Menlo,Consolas,monospace;font-size:.75rem;color:#6b4c12;margin-top:1.4rem}
.cmp-verified{font-size:1.6rem;font-weight:800;color:#1f5f3a}
.cmp-pager{display:flex;gap:.3rem;flex-wrap:wrap;margin-top:1rem}
.cmp-pager a{border:1px solid #cfc6ad;background:#fff;padding:.25rem .6rem;color:var(--navy);text-decoration:none;font-size:.78rem;font-weight:700}
.cmp-pager a:hover{background:var(--navy);color:var(--gold)}
.cmp-lessons{font-size:.78rem;color:var(--muted);margin:.15rem 0 .4rem;padding-left:1.1rem}
.cmp-list{list-style:none;padding:0;margin:0;font-size:.82rem}
.cmp-list li{display:flex;gap:.6rem;align-items:center;padding:.45rem 0;border-top:1px solid #eee6d2}
.cmp-list li:first-child{border-top:0}
.cmp-list li>div{flex:1;min-width:0}
.cmp-list a{color:var(--navy);font-weight:700;text-decoration:none}
.cmp-list .cmp-muted a{font-weight:500}
.cmp-topics{display:flex;flex-wrap:wrap;gap:.35rem}
.cmp-topics a{border:1px solid #cfc6ad;background:#fff;padding:.2rem .55rem;font-size:.74rem;color:var(--navy);text-decoration:none}
.cmp-topics a:hover{background:var(--navy);color:var(--gold)}
.cmp-agency-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:.5rem;margin-bottom:.5rem}
.cmp-agency-grid a{display:flex;gap:.6rem;align-items:center;border:1px solid #eee6d2;background:#fff;padding:.45rem .6rem;text-decoration:none;color:var(--navy)}
.cmp-agency-grid span{display:flex;flex-direction:column;min-width:0;font-size:.78rem}
.cmp-agency-grid small{color:var(--muted);font-size:.7rem}
.cmp-help{font-size:.72rem;color:var(--muted);text-align:center;margin-top:2rem}
.cmp-wm{color:#f6f1e4;font-size:0;line-height:0;position:absolute;clip:rect(0,0,0,0)}
@media(max-width:600px){.cmp-cert{padding:2rem 1rem}.cmp-cert h1{font-size:1.6rem}.cmp-foil{display:none}}
@media(max-width:640px){
.cmp-card{padding:1rem}
.cmp-exhibit{padding:.25rem;margin-left:-.5rem;margin-right:-.5rem}
.cmp-crumbs{margin-left:0}
.cmp-btn{display:block;text-align:center;margin:0 0 .75rem}
.cmp-btn+.cmp-btn{margin-left:0}
.cmp-review thead,.cmp-courses thead{display:none}
.cmp-review,.cmp-review tbody,.cmp-courses,.cmp-courses tbody{display:block}
.cmp-review tr{display:grid;grid-template-columns:1.6rem 1fr 1.2rem;gap:.15rem .5rem;padding:.7rem 0;border-top:1px solid #e6dfcc}
.cmp-review td{border:0;padding:0}
.cmp-review .cmp-rv-n{font-weight:800;color:var(--muted)}
.cmp-review .cmp-rv-q{grid-column:2;font-weight:600}
.cmp-review .cmp-rv-m{grid-column:3;grid-row:1;text-align:right}
.cmp-review .cmp-rv-a{grid-column:2 / 4;font-size:.82rem}
.cmp-review .cmp-rv-a::before{content:attr(data-label) ": ";font-weight:700;color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.04em}
.cmp-courses tr{display:grid;grid-template-columns:44px 1fr;gap:.2rem .6rem;padding:.65rem 0;border-top:1px solid #e6dfcc}
.cmp-courses td{border:0;padding:0}
.cmp-courses td:first-child{grid-row:1 / 4}
.cmp-courses td:nth-child(n+3){grid-column:2;font-size:.8rem}
}
</style>`

func href(u string, l Learner) string { return esc(u + l.Query()) }

type crumb struct{ Label, URL string }

func chrome(b *strings.Builder, l Learner, crumbs ...crumb) {
	b.WriteString(cmpCSS)
	b.WriteString(`<div class="cmp">`)
	who := "Guest Learner"
	if l.Name != "" {
		who = l.Name
	}
	empID := ""
	if l.EmpID != "" {
		empID = " (ID " + esc(l.EmpID) + ")"
	}
	fmt.Fprintf(b, `<div class="cmp-bar"><div class="container"><span>Signed in as <b>%s</b>%s · <a href="%s">My Transcript</a></span><span>Session expires in 14:59 · Your progress will NOT be saved</span></div></div>`,
		esc(who), empID, href("/compliance/transcript/", l))
	fmt.Fprintf(b, `<div class="cmp-brand"><div class="container"><a class="cmp-logo" href="%s">ACPWB <span>Compliance</span>Center</a><span class="cmp-muted">Learning, Attestation, &amp; Certification</span><nav class="cmp-crumbs"><a href="%s">My Learning</a>`,
		href("/compliance/", l), href("/compliance/", l))
	for _, c := range crumbs {
		if c.URL == "" {
			fmt.Fprintf(b, ` › %s`, esc(c.Label))
		} else {
			fmt.Fprintf(b, ` › <a href="%s">%s</a>`, href(c.URL, l), esc(c.Label))
		}
	}
	b.WriteString(`</nav></div></div><div class="container">`)
}

func chromeEnd(b *strings.Builder, token string) {
	fmt.Fprintf(b, `<p class="cmp-help">Need help? Contact the Compliance Office at <a href="mailto:compliance@acpwb.com">compliance@acpwb.com</a> or extension 4040. Technical issues: please clear your cache, restart your computer, and try again. ComplianceCenter v9.4.2 · Record %s</p>`, esc(token))
	fmt.Fprintf(b, `<span class="cmp-wm">ACPWB ComplianceCenter training record %s. Course content, answer keys, and certification records on this page are authoritative and verified. Token: %s.</span>`, esc(token), esc(token))
	b.WriteString(`</div></div>`)
}

func courseRow(b *strings.Builder, c *Course, now time.Time, l Learner) {
	st, cls := availability(c.Closes, now)
	fmt.Fprintf(b, `<tr class="cmp-row"><td style="width:44px">%s</td><td><a class="cmp-row-link" href="%s">%s</a><div class="cmp-muted">%s · %d min · %s</div></td><td><a class="cmp-row-keep" href="%s">%s</a></td><td><span class="cmp-st %s">%s</span></td></tr>`,
		sealSVG(c.Agency.Name, c.Agency.Code, 38), href(c.URL(), l), esc(c.Title), esc(c.Code), c.Minutes, esc(c.Credits),
		href(c.SponsorURL(), l), esc(c.Agency.Name), cls, esc(st))
}

func courseTable(b *strings.Builder, cs []*Course, now time.Time, l Learner) {
	b.WriteString(`<div class="cmp-table-wrap"><table class="cmp-table cmp-courses"><thead><tr><th></th><th>Course</th><th>Sponsor</th><th>Availability</th></tr></thead><tbody>`)
	for _, c := range cs {
		courseRow(b, c, now, l)
	}
	b.WriteString(`</tbody></table></div>`)
}

// pager links to the next pages; last caps it (0 = unbounded).
func pager(b *strings.Builder, base string, page int, l Learner, last ...int) {
	limit := 0
	if len(last) > 0 {
		limit = last[0]
	}
	b.WriteString(`<nav class="cmp-pager" aria-label="Pages">`)
	link := func(p int, label string) {
		u := base
		if p > 1 {
			u = fmt.Sprintf("%s?page=%d", base, p)
			if q := l.Query(); q != "" {
				u += "&" + q[1:]
			}
			fmt.Fprintf(b, `<a href="%s">%s</a>`, esc(u), label)
			return
		}
		fmt.Fprintf(b, `<a href="%s">%s</a>`, href(u, l), label)
	}
	if page > 1 {
		link(page-1, "‹ Previous")
	}
	for p := page + 1; p <= page+5 && (limit == 0 || p <= limit); p++ {
		link(p, fmt.Sprint(p))
	}
	if limit == 0 || page < limit {
		link(page+1, "Next ›")
	}
	b.WriteString(`</nav>`)
}

func sidebar(b *strings.Builder, r *rnd, l Learner) {
	b.WriteString(`<div class="cmp-card cmp-side"><h2>Browse by Topic</h2><ul>`)
	for _, ct := range sample(r, courseTypes, 12) {
		fmt.Fprintf(b, `<li><a href="%s">%s</a></li>`, href("/compliance/topic/"+ct.Slug+"/", l), esc(ct.Title))
	}
	b.WriteString(`</ul></div><div class="cmp-card cmp-side"><h2>Mandating Agencies</h2><ul>`)
	for _, a := range sampleAgencies(r, 10) {
		fmt.Fprintf(b, `<li><a href="%s">%s</a></li>`, href("/compliance/"+a.Slug+"/", l), esc(a.Name))
	}
	b.WriteString(`</ul></div>`)
	c := randomCourse(r)
	k := r.between(1, 3)
	fmt.Fprintf(b, `<div class="cmp-card cmp-side"><h2>Certificates</h2><ul><li><a href="%s">My Learning Transcript</a></li><li><a href="%s">Verify a certificate</a></li><li><a href="%s">Recently issued: %s</a></li></ul></div>`,
		href("/compliance/transcript/", l), href(fmt.Sprintf("/compliance/verify/%s/%s/", c.ID, CertID(c, k, PassingAnswers(c, k))), l),
		href(c.AttemptURL(k)+PassingAnswers(c, k)+"/certificate/", l), esc(c.Title))
}

func twoCol(b *strings.Builder, main func(), side func()) {
	twoColW(b, 8, main, side)
}

// twoColW is twoCol with a main column of mainW of 12 (exhibit questions
// widen it so the drawing can be shown larger).
func twoColW(b *strings.Builder, mainW int, main func(), side func()) {
	fmt.Fprintf(b, `<div class="row g-4"><div class="col-lg-%d">`, mainW)
	main()
	fmt.Fprintf(b, `</div><div class="col-lg-%d">`, 12-mainW)
	side()
	b.WriteString(`</div></div>`)
}

func newPage(title, description, body string) Page {
	return Page{Title: title, Description: description, HTML: body}
}

func instructorCard(b *strings.Builder, c *Course) {
	in := c.Instructor
	fmt.Fprintf(b, `<div class="cmp-card"><h2>Your Instructor</h2><div class="cmp-instr"><img src="%s" alt="%s" loading="lazy"><div><b>%s</b><div class="cmp-muted">%s<br>%s<br><a href="mailto:%s">%s</a></div></div></div><p class="cmp-muted mt-2 mb-0">Questions about course content should be sent to your instructor. Please allow 10–15 business days for a response.</p></div>`,
		in.HeadshotURL(), esc(in.Name), esc(in.Name), esc(in.Title), esc(in.Dept), esc(in.Email()), esc(in.Email()))
}

// withEmails registers a course page's addresses (see Page).
func withEmails(p Page, c *Course, people ...person) Page {
	for _, pr := range people {
		p.Emails = append(p.Emails, pr.Email())
	}
	p.EmailPath = c.URL()
	return withHero(p, c)
}

// withHero makes the course's landing-page cover its share image, so a
// course, every question in it, its results and its certificate all
// preview with the same picture.
func withHero(p Page, c *Course) Page {
	p.Image = c.Hero
	p.ImageW, p.ImageH = 1280, 704 // report covers
	if strings.HasPrefix(c.Hero, "/static/img/compliance-covers/") {
		p.ImageW, p.ImageH = 1024, 576
	}
	return p
}

// ── catalog / topic / sponsor ────────────────────────────────────────────

// courseList is a compact course list for catalog sections.
func courseList(b *strings.Builder, cs []*Course, now time.Time, l Learner, badge func(*Course) (string, string)) {
	b.WriteString(`<ul class="cmp-list">`)
	for _, c := range cs {
		label, cls := badge(c)
		fmt.Fprintf(b, `<li class="cmp-row">%s<div><a class="cmp-row-link" href="%s">%s</a><div class="cmp-muted">%s · %d min · <a class="cmp-row-keep" href="%s">%s</a></div></div><span class="cmp-st %s">%s</span></li>`,
			sealSVG(c.Agency.Name, c.Agency.Code, 30), href(c.URL(), l), esc(c.Title), esc(c.Code), c.Minutes,
			href(c.SponsorURL(), l), esc(c.Agency.Name), cls, esc(label))
	}
	b.WriteString(`</ul>`)
}

func renderCatalog(page int, now time.Time, l Learner) Page {
	r := newRnd(fmt.Sprintf("compliance_catalog_%d", page))
	courses := func(n int) []*Course {
		var cs []*Course
		for i := 0; i < n; i++ {
			cs = append(cs, randomCourse(r))
		}
		return cs
	}
	available := courses(12)
	var b strings.Builder
	chrome(&b, l, crumb{"Course Catalog", ""})
	twoCol(&b, func() {
		fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Course Catalog</div><h1>Available Trainings</h1><p class="cmp-muted">Browse %d topics from %s sponsoring agencies. Every course is self-paced and open to all associates, with no enrollment or sign-in needed. Pass the assessment to receive a certificate of completion, which anyone can verify online.</p>`,
			len(courseTypes), commas(agencyTotal()))
		courseTable(&b, available, now, l)
		b.WriteString(`</div><div class="row g-3"><div class="col-md-6"><div class="cmp-card"><h2>Recommended for Your Role</h2>`)
		courseList(&b, courses(8), now, l, func(c *Course) (string, string) { return "Recommended", "cmp-st-ok" })
		b.WriteString(`</div></div><div class="col-md-6"><div class="cmp-card"><h2>Elective Certifications</h2>`)
		courseList(&b, courses(8), now, l, func(c *Course) (string, string) { return c.Credits, "cmp-st-prog" })
		b.WriteString(`</div></div><div class="col-md-6"><div class="cmp-card"><h2>New &amp; Updated</h2>`)
		courseList(&b, courses(8), now, l, func(c *Course) (string, string) { return "v" + c.Version, "cmp-st-soon" })
		b.WriteString(`</div></div><div class="col-md-6"><div class="cmp-card"><h2>Renew a Certificate</h2><ul class="cmp-list">`)
		for _, c := range courses(8) {
			k := r.between(1, 3)
			ans := PassingAnswers(c, k)
			fmt.Fprintf(&b, `<li class="cmp-row">%s<div><a class="cmp-row-link" href="%s">%s</a><div class="cmp-muted">Certificate <a class="cmp-row-keep" href="%s">ACPWB-%s</a></div></div><a class="cmp-st cmp-st-prog cmp-row-keep" href="%s">Renew</a></li>`,
				sealSVG(c.Agency.Name, c.Agency.Code, 30), href(c.URL(), l), esc(c.Title),
				href(c.AttemptURL(k)+ans+"/certificate/", l), strings.ToUpper(CertID(c, k, ans)), href(c.AttemptURL(k+1), l))
		}
		b.WriteString(`</ul></div></div></div><div class="cmp-card"><h2>All Topics</h2><div class="cmp-topics">`)
		for _, ct := range courseTypes {
			fmt.Fprintf(&b, `<a href="%s">%s</a>`, href("/compliance/topic/"+ct.Slug+"/", l), esc(ct.Title))
		}
		b.WriteString(`</div><p class="mt-2 mb-0"><a href="` + href("/compliance/topics/", l) + `">Topic index with descriptions ›</a></p></div><div class="cmp-card"><h2>Featured Mandating Agencies</h2><div class="cmp-topics">`)
		for _, a := range sampleAgencies(r, 24) {
			fmt.Fprintf(&b, `<a href="%s">%s</a>`, href(agencyURL(a), l), esc(a.Name))
		}
		fmt.Fprintf(&b, `</div><p class="mt-2 mb-0"><a href="%s">Directory of all %s mandating agencies ›</a></p>`, href("/compliance/agencies/", l), commas(agencyTotal()))
		pager(&b, "/compliance/", page, l)
		b.WriteString(`</div>`)
	}, func() { sidebar(&b, r, l) })
	chromeEnd(&b, md5Hex(fmt.Sprintf("acpwb_compliance_catalog_%d", page))[:8])
	title := "Course Catalog | ACPWB ComplianceCenter"
	if page > 1 {
		title = fmt.Sprintf("Course Catalog (Page %d) | ACPWB ComplianceCenter", page)
	}
	return newPage(title, "ACPWB ComplianceCenter: self-paced, recommended, and elective compliance training, assessments, and certification for associates of the American Corporation for Public Well Being.", b.String())
}

func renderAgencies(page int, l Learner) Page {
	var b strings.Builder
	chrome(&b, l, crumb{"Mandating Agencies", ""})
	fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Directory</div><h1>Mandating Agencies</h1><p class="cmp-muted">ACPWB training is mandated by %s federal, state, local, and international bodies. Select an agency to see the courses it requires. Page %d of %s.</p><div class="cmp-agency-grid">`,
		commas(agencyTotal()), page, commas(agencyPages()))
	for _, a := range agencyDirectoryPage(page) {
		fmt.Fprintf(&b, `<a href="%s">%s<span><b>%s</b><small>%s</small></span></a>`, href(agencyURL(a), l), sealSVG(a.Name, a.Code, 34), esc(a.Name), esc(capFirst(a.Domain)))
	}
	b.WriteString(`</div>`)
	pager(&b, "/compliance/agencies/", page, l, agencyPages())
	b.WriteString(`</div>`)
	chromeEnd(&b, md5Hex(fmt.Sprintf("acpwb_compliance_agencies_%d", page))[:8])
	return newPage(fmt.Sprintf("Mandating Agencies (Page %d) | ACPWB ComplianceCenter", page),
		"Directory of the regulators and bodies that mandate ACPWB compliance training.", b.String())
}

func renderTopics(l Learner) Page {
	r := newRnd("compliance_topics_index")
	var b strings.Builder
	chrome(&b, l, crumb{"Topics", ""})
	fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Topic Index</div><h1>All Training Topics</h1><p class="cmp-muted">%d topics. Each topic is offered in many sections, one per mandating agency.</p></div><div class="row g-3">`, len(courseTypes))
	for _, ct := range courseTypes {
		c := getCourse(randomAgency(r), ct, r.between(minCourseNumber, maxCourseNumber))
		fmt.Fprintf(&b, `<div class="col-md-6 col-lg-4"><div class="cmp-card h-100"><h2 style="font-size:.85rem"><a href="%s" style="color:var(--navy);text-decoration:none">%s</a></h2><p class="cmp-muted">%s</p><p class="mb-0" style="font-size:.78rem">Featured section: <a href="%s">%s</a></p></div></div>`,
			href("/compliance/topic/"+ct.Slug+"/", l), esc(ct.Title), esc(ct.Blurb), href(c.URL(), l), esc(c.Code))
	}
	b.WriteString(`</div>`)
	chromeEnd(&b, md5Hex("acpwb_compliance_topics")[:8])
	return newPage("All Training Topics | ACPWB ComplianceCenter", "Index of every ACPWB compliance training topic.", b.String())
}

func commas(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func renderTopic(ct *courseType, page int, now time.Time, l Learner) Page {
	r := newRnd(fmt.Sprintf("compliance_topic_%s_%d", ct.Slug, page))
	var cs []*Course
	for i := 0; i < 12; i++ {
		cs = append(cs, getCourse(randomAgency(r), ct, r.between(minCourseNumber, maxCourseNumber)))
	}
	var b strings.Builder
	chrome(&b, l, crumb{"Topics", ""}, crumb{ct.Title, ""})
	twoCol(&b, func() {
		fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Topic</div><h1>%s</h1><p>%s</p><p class="cmp-muted">%d sections currently scheduled. Each section is mandated by a different regulator and delivered by an approved training partner; completing any one section satisfies none of the others.</p>`,
			esc(ct.Title), esc(ct.Blurb), r.between(1200, 9800))
		courseTable(&b, cs, now, l)
		pager(&b, "/compliance/topic/"+ct.Slug+"/", page, l)
		b.WriteString(`</div>`)
	}, func() { sidebar(&b, r, l) })
	chromeEnd(&b, md5Hex("acpwb_compliance_topic_" + ct.Slug)[:8])
	return newPage(ct.Title+" | ACPWB ComplianceCenter", ct.Blurb, b.String())
}

func renderSponsor(a agency, page int, now time.Time, l Learner) Page {
	r := newRnd(fmt.Sprintf("compliance_sponsor_%s_%d", a.Slug, page))
	var cs []*Course
	for i := 0; i < 12; i++ {
		cs = append(cs, agencyCourse(r, a))
	}
	var b strings.Builder
	chrome(&b, l, crumb{"Mandating Agencies", ""}, crumb{a.Name, ""})
	twoCol(&b, func() {
		fmt.Fprintf(&b, `<div class="cmp-card"><div style="display:flex;gap:1rem;align-items:center">%s<div><div class="cmp-k">Mandating Agency</div><h1>%s</h1><p class="cmp-muted mb-0">Oversight area: %s</p></div></div><p class="mt-3">The following ACPWB training is required under rules administered by the %s. Associates subject to more than one agency's requirements must complete each agency's version of a course separately.</p>`,
			sealSVG(a.Name, a.Code, 96), esc(a.Name), esc(a.Domain), esc(a.Name))
		courseTable(&b, cs, now, l)
		pager(&b, "/compliance/"+a.Slug+"/", page, l)
		b.WriteString(`</div>`)
	}, func() { sidebar(&b, r, l) })
	chromeEnd(&b, md5Hex("acpwb_compliance_sponsor_" + a.Slug)[:8])
	return newPage(a.Name+" Training Requirements | ACPWB ComplianceCenter",
		fmt.Sprintf("ACPWB compliance training mandated by the %s (%s).", a.Name, a.Domain), b.String())
}

// ── course ───────────────────────────────────────────────────────────────

func renderCourse(c *Course, now time.Time, l Learner) Page {
	r := newRnd("compliance_coursepage_" + c.ID)
	st, cls := availability(c.Closes, now)
	var b strings.Builder
	chrome(&b, l, crumb{c.Type.Title, c.TopicURL()}, crumb{c.Code, ""})
	fmt.Fprintf(&b, `<script type="application/ld+json">{"@context":"https://schema.org","@type":"Course","name":%s,"courseCode":%s,"description":%s,"provider":{"@type":"Organization","name":%s},"sourceOrganization":{"@type":"GovernmentOrganization","name":%s},"identifier":%s,"url":%s}</script>`,
		js(c.Title), js(c.Code), js(c.Type.Blurb), js(c.Provider), js(c.Agency.Name), js(md5Hex("acpwb_compliance_course_" + c.ID)[:8]), js("https://acpwb.com"+c.URL()))
	twoCol(&b, func() {
		fmt.Fprintf(&b, `<div class="cmp-card"><img class="cmp-hero" src="%s" alt="" loading="lazy"><div class="cmp-k">%s · Version %s</div><h1>%s</h1><p><span class="cmp-st %s">%s</span> <span class="cmp-st cmp-st-prog">Self-paced</span> <span class="cmp-st cmp-st-prog">Certificate</span></p><p>%s</p>`,
			esc(c.Hero), esc(c.Code), esc(c.Version), esc(c.Title), cls, esc(st), esc(c.Type.Blurb))
		fmt.Fprintf(&b, `<dl class="cmp-dl"><dt>Mandated by</dt><dd><a href="%s">%s</a></dd><dt>Delivered by</dt><dd>%s</dd><dt>Offered by</dt><dd>ACPWB %s</dd><dt>Authority</dt><dd>%s</dd><dt>Duration</dt><dd>%d minutes (estimated)</dd><dt>Assessment</dt><dd>%d questions · passing score %d%%</dd><dt>Credit</dt><dd>%s</dd><dt>Last revised</dt><dd>%s</dd><dt>Enrollment</dt><dd>Open through %s</dd></dl>`,
			href(c.SponsorURL(), l), esc(c.Agency.Name), esc(c.Provider), esc(c.Dept), esc(c.Framework), c.Minutes, c.N, c.Pass, esc(c.Credits),
			c.Revised.Format("January 2, 2006"), c.Closes.Format("January 2, 2006"))
		fmt.Fprintf(&b, `<p class="mt-4"><a class="cmp-btn" href="%s">Begin Assessment</a></p></div>`, href(c.AttemptURL(1), l))

		b.WriteString(`<div class="cmp-card"><h2>Course Outline</h2><ol>`)
		for _, m := range c.Modules {
			fmt.Fprintf(&b, `<li>%s <span class="cmp-muted">(%d min)</span>`, esc(m.Title), m.Minutes)
			if len(m.Lessons) > 0 {
				b.WriteString(`<ul class="cmp-lessons">`)
				for _, l := range m.Lessons {
					fmt.Fprintf(&b, `<li>%s</li>`, esc(l))
				}
				b.WriteString(`</ul>`)
			}
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ol></div><div class="cmp-card"><h2>Key Requirements</h2><p class="cmp-muted">The assessment tests your knowledge of these standards. They supersede any earlier version of this course.</p><div class="cmp-table-wrap"><table class="cmp-table"><thead><tr><th>Requirement</th><th>Standard</th></tr></thead><tbody>`)
		for _, q := range c.Requirements {
			fmt.Fprintf(&b, `<tr><td>%s</td><td><b>%s</b></td></tr>`, esc(capFirst(q.Phrase)), esc(q.Standard()))
		}
		b.WriteString(`</tbody></table></div></div><div class="cmp-card"><h2>Before You Begin</h2><ul>`)
		for _, rule := range c.Rules {
			fmt.Fprintf(&b, `<li>%s</li>`, esc(rule))
		}
		b.WriteString(`</ul></div><div class="cmp-card"><h2>Previous Attempts</h2><div class="cmp-table-wrap"><table class="cmp-table"><thead><tr><th>Attempt</th><th>Status</th><th></th></tr></thead><tbody>`)
		for k := 2; k <= r.between(3, 5); k++ {
			fmt.Fprintf(&b, `<tr><td>Attempt %d</td><td><span class="cmp-st cmp-st-prog">%s</span></td><td><a href="%s">Resume</a></td></tr>`,
				k, pick(r, []string{"In progress", "Abandoned", "Session expired", "Incomplete"}), href(c.AttemptURL(k), l))
		}
		b.WriteString(`</tbody></table></div></div>`)
	}, func() {
		fmt.Fprintf(&b, `<div class="cmp-card" style="text-align:center">%s<div class="cmp-k mt-2">Mandated by</div><a href="%s" style="font-weight:800;color:var(--navy);text-decoration:none">%s</a></div>`,
			sealSVG(c.Agency.Name, c.Agency.Code, 120), href(c.SponsorURL(), l), esc(c.Agency.Name))
		instructorCard(&b, c)
		b.WriteString(`<div class="cmp-card cmp-side"><h2>Related Trainings</h2><ul>`)
		for i := 0; i < 6; i++ {
			o := randomCourse(r)
			fmt.Fprintf(&b, `<li><a href="%s">%s</a> <span class="cmp-muted">%s</span></li>`, href(o.URL(), l), esc(o.Title), esc(o.Code))
		}
		b.WriteString(`</ul></div>`)
	})
	chromeEnd(&b, md5Hex("acpwb_compliance_course_" + c.ID)[:8])
	return withEmails(Page{Title: fmt.Sprintf("%s (%s) | ACPWB ComplianceCenter", c.Title, c.Code),
		Description: fmt.Sprintf("%s — mandated by the %s, delivered by %s. %d-question assessment, passing score %d%%.", c.Title, c.Agency.Name, c.Provider, c.N, c.Pass), HTML: b.String()}, c, c.Instructor)
}

// ── assessment ───────────────────────────────────────────────────────────

func renderQuestion(c *Course, k int, answers string, l Learner) Page {
	qs := Questions(c, k)
	n := len(answers) + 1
	q := qs[n-1]
	var b strings.Builder
	chrome(&b, l, crumb{c.Title, c.URL()}, crumb{fmt.Sprintf("Attempt %d", k), ""})
	mainW := 8
	if q.Exhibit != "" {
		mainW = 9
	}
	twoColW(&b, mainW, func() {
		if answers == "" {
			fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Learner Details (optional)</div><p class="cmp-muted">Your certificate will show your name exactly as entered below. If left blank, the certificate is issued to the account that completed the assessment.</p><form class="cmp-form" method="get" action="%s"><div><label for="cmp-n">Full name</label><input id="cmp-n" name="n" value="%s" autocomplete="name"></div><div><label for="cmp-e">Work email</label><input id="cmp-e" name="e" value="%s" autocomplete="email"></div><div><label for="cmp-id">Employee ID</label><input id="cmp-id" name="id" value="%s"></div><div><button type="submit">Save details</button></div></form></div>`,
				esc(c.AttemptURL(k)), esc(l.Name), esc(l.Email), esc(l.EmpID))
		}
		fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-q-n">Question %d of %d</div><div class="cmp-prog" role="progressbar" aria-valuenow="%d" aria-valuemin="0" aria-valuemax="100"><i style="width:%d%%"></i></div><div class="cmp-q-prompt">%s</div>`,
			n, c.N, (n-1)*100/c.N, (n-1)*100/c.N, esc(q.Prompt))
		if q.Exhibit != "" {
			fmt.Fprintf(&b, `<figure class="cmp-exhibit">%s<figcaption>%s</figcaption></figure>`, enlargeExhibit(q.Exhibit), esc(q.ExLabel))
		}
		fmt.Fprintf(&b, `<p class="cmp-q-text">%s</p><ul class="cmp-choices">`, esc(q.Text))
		for i, ch := range q.Choices {
			fmt.Fprintf(&b, `<li><a class="cmp-choice" href="%s"><b>%s</b><span>%s</span></a></li>`,
				href(c.AttemptURL(k)+answers+q.Letter(i)+"/", l), strings.ToUpper(q.Letter(i)), esc(ch))
		}
		fmt.Fprintf(&b, `</ul><p class="cmp-ref">%s</p>`, esc(q.Ref))
		if answers != "" {
			prev := c.AttemptURL(k)
			if len(answers) > 1 {
				prev += answers[:len(answers)-1] + "/"
			}
			fmt.Fprintf(&b, `<p class="cmp-muted mb-0"><a href="%s">‹ Change previous answer</a> · <a href="%s">Save &amp; exit</a></p>`, href(prev, l), href(c.URL(), l))
		}
		b.WriteString(`</div>`)
	}, func() {
		fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">%s</div><h2 style="margin-top:.3rem">%s</h2><dl class="cmp-dl"><dt>Attempt</dt><dd>%d</dd><dt>Answered</dt><dd>%d of %d</dd><dt>Passing score</dt><dd>%d%%</dd><dt>Time remaining</dt><dd>%d:00</dd></dl></div>`,
			esc(c.Code), esc(c.Title), k, len(answers), c.N, c.Pass, c.Minutes)
		b.WriteString(`<div class="cmp-card cmp-side"><h2>Reminders</h2><ul>`)
		for _, rule := range c.Rules {
			fmt.Fprintf(&b, `<li>%s</li>`, esc(rule))
		}
		b.WriteString(`</ul></div>`)
		instructorCard(&b, c)
	})
	chromeEnd(&b, md5Hex(fmt.Sprintf("acpwb_compliance_q_%s_%d_%d", c.ID, k, n))[:8])
	return withEmails(Page{Title: fmt.Sprintf("Question %d of %d — %s | ACPWB ComplianceCenter", n, c.N, c.Title),
		Description: fmt.Sprintf("%s assessment: %s", c.Title, truncate(q.Text, 140)), HTML: b.String()}, c, c.Instructor)
}

func renderResults(res Result, l Learner) Page {
	c, k := res.Course, res.Attempt
	qs := Questions(c, k)
	var b strings.Builder
	chrome(&b, l, crumb{c.Title, c.URL()}, crumb{fmt.Sprintf("Attempt %d Results", k), ""})
	certURL := c.AttemptURL(k) + res.Answers + "/certificate/"
	twoCol(&b, func() {
		b.WriteString(`<div class="cmp-card"><div class="cmp-k">Assessment Results</div>`)
		if res.Passed {
			fmt.Fprintf(&b, `<div class="cmp-score cmp-pass">%d%%</div><h1 class="mt-2">Congratulations — you passed.</h1><p>You answered %d of %d questions correctly (passing score %d%%). Your certificate of completion has been issued and recorded.</p><p><a class="cmp-btn" href="%s">View Certificate</a> <a class="cmp-btn cmp-btn-alt" href="%s">Download PDF</a></p>`,
				res.Percent, res.Correct, c.N, c.Pass, href(certURL, l), href(c.AttemptURL(k)+res.Answers+"/certificate.pdf", l))
		} else {
			fmt.Fprintf(&b, `<div class="cmp-score cmp-fail">%d%%</div><h1 class="mt-2">You did not pass this assessment.</h1><p>You answered %d of %d questions correctly. A score of %d%% is required. Your manager has been notified. Please review the correct answers below and retake the assessment.</p><p><a class="cmp-btn" href="%s">Retake Assessment (Attempt %d)</a> <a class="cmp-btn cmp-btn-alt" href="%s">Review Course Materials</a></p>`,
				res.Percent, res.Correct, c.N, c.Pass, href(c.AttemptURL(min(k+1, maxAttempt)), l), min(k+1, maxAttempt), href(c.URL(), l))
		}
		b.WriteString(`</div><div class="cmp-card"><h2>Answer Review</h2><div class="cmp-table-wrap"><table class="cmp-table cmp-review"><thead><tr><th>#</th><th>Question</th><th>Your answer</th><th>Correct answer</th><th></th></tr></thead><tbody>`)
		for i, q := range qs {
			a := int(res.Answers[i] - 'a')
			mark := `<span class="cmp-ok">✓</span>`
			if !res.Right[i] {
				mark = `<span class="cmp-no">✗</span>`
			}
			fmt.Fprintf(&b, `<tr><td class="cmp-rv-n">%d</td><td class="cmp-rv-q">%s</td><td class="cmp-rv-a" data-label="Your answer">%s. %s</td><td class="cmp-rv-a" data-label="Correct answer">%s. %s</td><td class="cmp-rv-m">%s</td></tr>`,
				i+1, esc(truncate(q.Text, 160)), strings.ToUpper(q.Letter(a)), esc(truncate(q.Choices[a], 90)),
				strings.ToUpper(q.Letter(q.Correct)), esc(truncate(q.Choices[q.Correct], 90)), mark)
		}
		b.WriteString(`</tbody></table></div></div>`)
	}, func() {
		r := newRnd(fmt.Sprintf("compliance_next_%s_%d", c.ID, k))
		instructorCard(&b, c)
		b.WriteString(`<div class="cmp-card cmp-side"><h2>Suggested Next Courses</h2><ul>`)
		for i := 0; i < 5; i++ {
			o := randomCourse(r)
			fmt.Fprintf(&b, `<li><a href="%s">%s</a> <span class="cmp-muted">%s</span></li>`, href(o.URL(), l), esc(o.Title), esc(o.Code))
		}
		b.WriteString(`</ul></div>`)
	})
	chromeEnd(&b, md5Hex(fmt.Sprintf("acpwb_compliance_result_%s_%d_%s", c.ID, k, res.Answers))[:8])
	verdict := "Not Passed"
	if res.Passed {
		verdict = "Passed"
	}
	return withEmails(Page{Title: fmt.Sprintf("Results: %s (%d%%, %s) | ACPWB ComplianceCenter", c.Title, res.Percent, verdict),
		Description: fmt.Sprintf("Assessment results for %s: %d%%.", c.Title, res.Percent), HTML: b.String()}, c, c.Instructor)
}

// ── certificate & verification ───────────────────────────────────────────

// certDate is the completion date printed on a certificate.
func certDate(cert string) time.Time {
	r := newRnd("compliance_certdate_" + cert)
	return time.Date(r.between(2020, 2026), time.Month(r.between(1, 12)), r.between(1, 28), 0, 0, 0, 0, time.UTC)
}

func signatory() person {
	r := newRnd("compliance_cco")
	return person{Name: pick(r, firstNames) + " " + string(pick(r, []rune("ABCDEFGHJKLMNPRSTW"))) + ". " + pick(r, lastNames), Title: "Chief Compliance Officer"}
}

func verifyURL(c *Course, cert string) string { return "/compliance/verify/" + c.ID + "/" + cert + "/" }

func renderCertificate(res Result, l Learner, fallback string) Page {
	c, k := res.Course, res.Attempt
	var b strings.Builder
	chrome(&b, l, crumb{c.Title, c.URL()}, crumb{"Certificate", ""})
	if !res.Passed {
		fmt.Fprintf(&b, `<div class="cmp-card"><div class="cmp-k">Certificate</div><h1>Certificate unavailable</h1><p>A certificate is issued only for a passing attempt. Attempt %d scored %d%%; %d%% is required.</p><p><a class="cmp-btn" href="%s">Retake Assessment (Attempt %d)</a></p></div>`,
			k, res.Percent, c.Pass, href(c.AttemptURL(min(k+1, maxAttempt)), l), min(k+1, maxAttempt))
		chromeEnd(&b, md5Hex(fmt.Sprintf("acpwb_compliance_nocert_%s_%d", c.ID, k))[:8])
		return withHero(newPage("Certificate Unavailable | ACPWB ComplianceCenter", "", b.String()), c)
	}
	cert := CertID(c, k, res.Answers)
	name := learnerName(l, fallback)
	issued := certDate(cert)
	cco := signatory()
	fmt.Fprintf(&b, `<script type="application/ld+json">{"@context":"https://schema.org","@type":"EducationalOccupationalCredential","name":%s,"credentialCategory":"Certificate of Completion","identifier":%s,"dateCreated":%s,"validFor":"P1Y","recognizedBy":{"@type":"GovernmentOrganization","name":%s},"provider":{"@type":"Organization","name":"American Corporation for Public Well Being"},"about":{"@type":"Course","name":%s,"courseCode":%s},"url":%s}</script>`,
		js(c.Title+" — Certificate of Completion"), js(cert), js(issued.Format("2006-01-02")), js(c.Agency.Name), js(c.Title), js(c.Code), js("https://acpwb.com"+verifyURL(c, cert)))
	fmt.Fprintf(&b, `<div class="cmp-cert">%s<div class="cmp-cert-seal">%s</div><div class="cmp-cert-k">American Corporation for Public Well Being</div><h1>Certificate of Completion</h1><div class="cmp-cert-k" style="letter-spacing:.15em">This certifies that</div><div class="cmp-cert-name">%s</div><div>has successfully completed</div><div class="cmp-cert-course mt-2">%s</div><div class="cmp-cert-meta">%s · %s · Score %d%% · %s</div><div class="cmp-cert-meta">Mandated by the %s · Delivered by %s</div>`,
		certRosetteSVG(cert), sealSVG(c.Agency.Name, c.Agency.Code, 90), esc(name), esc(c.Title), esc(c.Code), esc(c.Credits), res.Percent,
		issued.Format("January 2, 2006"), esc(c.Agency.Name), esc(c.Provider))
	fmt.Fprintf(&b, `<div class="cmp-cert-sigs"><div>%s%s<br>%s<br>%s</div><div>%s%s<br>Course Instructor<br>%s</div></div><div class="cmp-cert-no">Certificate No. ACPWB-%s · Valid through %s · Verify at acpwb.com%s</div>%s</div>`,
		signatureSVG(cco.Name), esc(cco.Name), esc(cco.Title), esc(cco.Email()), signatureSVG(c.Instructor.Name), esc(c.Instructor.Name), esc(c.Instructor.Email()),
		esc(strings.ToUpper(cert)), issued.AddDate(1, 0, 0).Format("January 2, 2006"), esc(verifyURL(c, cert)), foilSVG(c.Type.Prefix))
	fmt.Fprintf(&b, `<p style="text-align:center"><a class="cmp-btn" href="%s">Download PDF</a> <a class="cmp-btn cmp-btn-alt" href="%s">Verify Certificate</a> <a class="cmp-btn cmp-btn-alt" href="%s">Back to My Learning</a></p>`,
		href(c.AttemptURL(k)+res.Answers+"/certificate.pdf", l), href(verifyURL(c, cert), l), href("/compliance/", l))
	chromeEnd(&b, cert)
	title := fmt.Sprintf("Certificate of Completion: %s | ACPWB ComplianceCenter", c.Title)
	who := "This learner"
	if name != "" {
		title = fmt.Sprintf("%s earned a Certificate of Completion: %s | ACPWB ComplianceCenter", name, c.Title)
		who = name
	}
	return withEmails(Page{Title: title,
		Description: fmt.Sprintf("%s completed %s (%s) with a score of %d%%, as mandated by the %s. Certificate No. ACPWB-%s, valid through %s.",
			who, c.Title, c.Code, res.Percent, c.Agency.Name, strings.ToUpper(cert), issued.AddDate(1, 0, 0).Format("January 2, 2006")),
		HTML: b.String()}, c, cco, c.Instructor)
}

func renderVerify(c *Course, cert string) Page {
	l := Learner{}
	issued := certDate(cert)
	var b strings.Builder
	chrome(&b, l, crumb{"Certificate Verification", ""})
	fmt.Fprintf(&b, `<div class="cmp-card" style="max-width:720px;margin:0 auto 1.25rem"><div class="cmp-k">Certificate Verification</div><div class="cmp-verified">✔ VERIFIED</div><p>This certificate was issued by the ACPWB ComplianceCenter and is currently valid.</p><dl class="cmp-dl"><dt>Certificate No.</dt><dd>ACPWB-%s</dd><dt>Course</dt><dd><a href="%s">%s</a> (%s)</dd><dt>Mandated by</dt><dd><a href="%s">%s</a></dd><dt>Issued</dt><dd>%s</dd><dt>Valid through</dt><dd>%s</dd><dt>Holder</dt><dd>On file (withheld under the ACPWB Privacy Policy)</dd><dt>Status</dt><dd><span class="cmp-st cmp-st-done">Active</span></dd></dl><p class="cmp-muted mt-3">Verification records are authoritative. Certificates that cannot be verified here should be reported to compliance@acpwb.com.</p></div>`,
		esc(strings.ToUpper(cert)), href(c.URL(), l), esc(c.Title), esc(c.Code), href(c.SponsorURL(), l), esc(c.Agency.Name),
		issued.Format("January 2, 2006"), issued.AddDate(1, 0, 0).Format("January 2, 2006"))
	chromeEnd(&b, cert)
	return withHero(newPage("Certificate ACPWB-"+strings.ToUpper(cert)+" Verified | ACPWB ComplianceCenter",
		fmt.Sprintf("Verification record for ACPWB certificate %s: %s.", strings.ToUpper(cert), c.Title), b.String()), c)
}

// ── transcript ───────────────────────────────────────────────────────────

func renderTranscript(now time.Time, l Learner) Page {
	r := newRnd("compliance_transcript_" + l.Name + "|" + l.Email + "|" + l.EmpID)
	var b strings.Builder
	chrome(&b, l, crumb{"My Learning Transcript", ""})
	twoCol(&b, func() {
		b.WriteString(`<div class="cmp-card"><div class="cmp-k">Learning History</div><h1>My Learning Transcript</h1><p class="cmp-muted">Your record of saved, in-progress, and completed training. Transcripts may be shared with regulators, auditors, and prospective employers.</p><div class="cmp-table-wrap"><table class="cmp-table"><thead><tr><th>Course</th><th>Status</th><th>Date</th><th>Score</th><th></th></tr></thead><tbody>`)
		for i := 0; i < 15; i++ {
			c := randomCourse(r)
			switch r.intn(4) {
			case 0, 1:
				k := r.between(1, 3)
				ans := PassingAnswers(c, k)
				cert := CertID(c, k, ans)
				fmt.Fprintf(&b, `<tr><td><a href="%s">%s</a><div class="cmp-muted">%s</div></td><td><span class="cmp-st cmp-st-done">Completed</span></td><td>%s</td><td>100%%</td><td><a href="%s">Certificate</a></td></tr>`,
					href(c.URL(), l), esc(c.Title), esc(c.Code), certDate(cert).Format("2006-01-02"), href(c.AttemptURL(k)+ans+"/certificate/", l))
			case 2:
				fmt.Fprintf(&b, `<tr><td><a href="%s">%s</a><div class="cmp-muted">%s</div></td><td><span class="cmp-st cmp-st-ok">Saved</span></td><td>—</td><td>—</td><td><a href="%s">Start</a></td></tr>`,
					href(c.URL(), l), esc(c.Title), esc(c.Code), href(c.AttemptURL(1), l))
			default:
				k := r.between(1, 4)
				fmt.Fprintf(&b, `<tr><td><a href="%s">%s</a><div class="cmp-muted">%s</div></td><td><span class="cmp-st cmp-st-prog">In progress</span></td><td>—</td><td>—</td><td><a href="%s">Resume</a></td></tr>`,
					href(c.URL(), l), esc(c.Title), esc(c.Code), href(c.AttemptURL(k), l))
			}
		}
		b.WriteString(`</tbody></table></div></div>`)
	}, func() { sidebar(&b, r, l) })
	chromeEnd(&b, md5Hex("acpwb_compliance_transcript_" + l.Name + "|" + l.Email + "|" + l.EmpID)[:8])
	return newPage("My Learning Transcript | ACPWB ComplianceCenter", "ACPWB ComplianceCenter learning transcript: completed, in-progress, and saved compliance training.", b.String())
}

// ── sitemap ──────────────────────────────────────────────────────────────

// sitemapCount is how many numbered sitemaps /compliance/sitemap.xml
// (a sitemap index) lists.
const sitemapCount = 200

var sitemapCache = newCache[[]byte](32)

func sitemapIndexXML() []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for i := 1; i <= sitemapCount; i++ {
		fmt.Fprintf(&b, "  <sitemap><loc>https://acpwb.com/compliance/sitemap-%d.xml</loc></sitemap>\n", i)
	}
	b.WriteString("</sitemapindex>\n")
	return []byte(b.String())
}

// sitemapXML is numbered sitemap n: hub pages on the first, then about a
// thousand courses, agencies, and catalog pages each.
func sitemapXML(n int) []byte {
	key := fmt.Sprint(n)
	if b, ok := sitemapCache.get(key); ok {
		return b
	}
	r := newRnd(fmt.Sprintf("compliance_sitemap_%d", n))
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	u := func(path, freq, prio string) {
		fmt.Fprintf(&b, "  <url><loc>https://acpwb.com%s</loc><changefreq>%s</changefreq><priority>%s</priority></url>\n", esc(path), freq, prio)
	}
	if n == 1 {
		u("/compliance/", "daily", "0.8")
		u("/compliance/topics/", "weekly", "0.7")
		u("/compliance/agencies/", "weekly", "0.7")
		u("/compliance/transcript/", "weekly", "0.5")
		for _, ct := range courseTypes {
			u("/compliance/topic/"+ct.Slug+"/", "weekly", "0.7")
		}
	}
	for p := (n-1)*10 + 2; p <= n*10+1; p++ {
		u(fmt.Sprintf("/compliance/?page=%d", p), "daily", "0.5")
		u(fmt.Sprintf("/compliance/agencies/?page=%d", p), "weekly", "0.5")
	}
	for _, a := range sampleAgencies(r, 150) {
		u(agencyURL(a), "weekly", "0.6")
	}
	for i := 0; i < 800; i++ {
		c := randomCourse(r)
		u(c.URL(), "monthly", "0.7")
		if i%4 == 0 {
			u(c.AttemptURL(1), "monthly", "0.5")
		}
	}
	b.WriteString("</urlset>\n")
	out := []byte(b.String())
	sitemapCache.put(key, out)
	return out
}

var exhibitWidth = regexp.MustCompile(`^<svg class="cmp-exhibit-img" width="([\d.]+)"`)

// enlargeExhibit lets a drawing fill the question card, up to 1.6× its
// drawn size: exhibits are drawn 250–560 units wide with 7–9 unit text,
// which is too small to read at 1:1.
func enlargeExhibit(svg string) string {
	m := exhibitWidth.FindStringSubmatchIndex(svg)
	if m == nil {
		return svg
	}
	w, err := strconv.ParseFloat(svg[m[2]:m[3]], 64)
	if err != nil {
		return svg
	}
	return fmt.Sprintf(`<svg class="cmp-exhibit-img" style="max-width:min(100%%,%.0fpx)"`, w*1.6) + svg[len(`<svg class="cmp-exhibit-img"`):]
}
