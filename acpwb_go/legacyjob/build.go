package legacyjob

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Anchor is a jump target inside a pre section (a file, routine, step...).
type Anchor struct {
	Line  int // line index within Section.Text
	ID    string
	Label string
}

// Section is one block of the source-code page.
type Section struct {
	Anchors []Anchor
	ID      string // anchor suffix
	Title   string // sidebar contents label
	Tab     string // small header tab on the printout block
	Kind    string // pre | html
	Text    string // pre: listing text (may contain link markers)
	HTML    string // html: pre-rendered markup
}

// Listing is the whole generated job for one page.
type Listing struct {
	sourceText string   // plain source (the "source" section may be a compiler listing)
	codeLines  []string // statement lines only (variety measurement)
	ID         ID
	In         *Input
	Host       string
	OS         string
	Compiler   string
	Wrapper    string
	Pipeline   []string
	Submitted  string
	Approved   string
	CPU        string
	CondCode   string
	Sections   []Section
	Roster     []*Coder
	Emails     []string // honeypot addresses shown on the page
	Links      []string // link-marker targets (index = marker number)
	Related    []CrossLink
	SrcLines   int
	License    string
}

func (l *Listing) addLink(url string) int {
	l.Links = append(l.Links, url)
	return len(l.Links) - 1
}

// ── bounded cache ───────────────────────────────────────────────────────────

const cacheMax = 256

var (
	cacheMu  sync.Mutex
	cacheLRU = list.New()
	cacheMap = map[string]*list.Element{}
)

type cacheEntry struct {
	key string
	l   *Listing
}

// Build generates (or returns the cached) listing for an input. Recursive
// crawlers re-hit the same URLs, so a small LRU saves regeneration.
func Build(in *Input) *Listing {
	key := fmt.Sprintf("%d/%d/%d/%s|%s", in.Year, in.Month, in.Day, in.Slug, in.ArticleURL)
	cacheMu.Lock()
	if el, ok := cacheMap[key]; ok {
		cacheLRU.MoveToFront(el)
		cacheMu.Unlock()
		return el.Value.(*cacheEntry).l
	}
	cacheMu.Unlock()

	l := build(in)

	cacheMu.Lock()
	defer cacheMu.Unlock()
	cacheMap[key] = cacheLRU.PushFront(&cacheEntry{key: key, l: l})
	for cacheLRU.Len() > cacheMax {
		old := cacheLRU.Back()
		cacheLRU.Remove(old)
		delete(cacheMap, old.Value.(*cacheEntry).key)
	}
	return l
}

func langUpper(langKey string) bool { return metaOf(langKey).upper }

// debrace removes template braces from article-derived text (some article
// fields deliberately contain unfilled "{n}" placeholders).
func debrace(s string) string { return strings.NewReplacer("{", "", "}", "").Replace(s) }

func build(orig *Input) *Listing {
	cp := *orig
	in := &cp
	in.Title, in.Org, in.Industry, in.Phase = debrace(in.Title), debrace(in.Org), debrace(in.Industry), debrace(in.Phase)
	in.Distribution, in.PeerGroup, in.DocVersion = debrace(in.Distribution), debrace(in.PeerGroup), debrace(in.DocVersion)
	in.Sentences = append([]string(nil), in.Sentences...)
	for i := range in.Sentences {
		in.Sentences[i] = debrace(in.Sentences[i])
	}
	in.Revisions = append([]Revision(nil), in.Revisions...)
	for i := range in.Revisions {
		in.Revisions[i].Description = debrace(in.Revisions[i].Description)
	}
	id := Identity(in.Year, in.Month, in.Day, in.Slug)
	x := newRnd(fmt.Sprintf("legacyjob_%d_%d_%d_%s", in.Year, in.Month, in.Day, in.Slug))

	g := &gen{x: x.fork("comments"), in: in, id: id, emails: map[string]bool{}, usedTpl: map[string]int{}}
	g.roster = buildRoster(x.fork("roster"), in)
	g.langName = id.Language
	g.lang2 = pick(x, []string{"FORTRAN IV", "COBOL", "RPG II", "PL/I", "assembler", "BASIC", "APL", "Pascal", "MUMPS", "Natural", "dBASE", "Clipper", "Ada", "SQL", "Pick BASIC"})
	g.dist = in.Distribution
	vp := pick(x, g.roster)
	for _, c := range g.roster {
		if strings.Contains(strings.ToLower(c.Role), "vice president") || strings.Contains(strings.ToLower(c.Role), "manager") {
			vp = c
		}
	}
	g.vp = vp.First + " " + vp.Last
	if vp.First == "" {
		g.vp = "the Vice President, Research Division"
	}

	l := &Listing{ID: id, In: in, Roster: g.roster}
	g.holder = pick(x, []string{"American Corporation for Public Well Being", "American Corporation for Public Well Being", "ACPWB Research Division",
		"American Corporation for Public Well Being, Milwaukee, Wisconsin"})
	g.yearFrom = g.roster[0].Start

	nm := newNaming(x.fork("names"), id.LangKey, g.roster[0])
	b := newBuilder(g, x.fork("program"), nm)
	prog := b.build()
	rep := buildReport(in, id, b)

	j := &job{g: g, x: x.fork("emit"), in: in, id: id, b: b, p: prog, nm: nm, rep: rep}
	j.hist = buildHistory(j)
	j.cs = newCStyle(j.x, 66, langUpper(id.LangKey) || (id.LangKey == "gf" && j.x.chance(0.5)))

	src := maintain(j, emitSource(j))
	l.SrcLines = len(src)
	l.sourceText = linesText(src)
	for _, ln := range src {
		if ln.Code {
			l.codeLines = append(l.codeLines, ln.Text)
		}
	}

	fam := langFamily(id.LangKey)
	cl := compilers[id.LangKey]
	if cl == nil {
		cl = compilers[fam]
	}
	l.Compiler = eraPick(x, cl, in.Year)
	l.OS = eraPick(x, operatingSystems, in.Year)
	l.Host = eraPick(x, hardware, in.Year)
	l.CPU = fmt.Sprintf("%d.%02d", x.between(0, 2), x.between(1, 99))
	l.CondCode = "0000"
	team := g.teamMember()
	l.Submitted = team.Name + ", " + team.Title
	appr := g.coder()
	l.Approved = appr.Display(x) + ", " + appr.Role

	wr := buildWrapper(j, l, src)
	l.Wrapper = wr.name
	l.Pipeline = wr.pipeline
	jobName := wr.jobName
	userID := g.roster[len(g.roster)-1].UserID
	// Line-printer separator pages only belong to mainframe / VAX jobs, and
	// punched cards only to the card era.
	printerEra := wr.name == "jcl" || wr.name == "dcl"
	cardEra := printerEra && (in.Year <= 1995 || id.LangKey == "f4")
	if printerEra {
		l.Sections = append(l.Sections, Section{ID: "banner", Title: "Separator page", Tab: "PRINT SEPARATOR — START", Kind: "pre", Text: bannerPage(j, l, jobName, userID, false)})
	}
	if cardEra {
		l.Sections = append(l.Sections, Section{ID: "cards", Title: "Punch cards", Tab: "CARD IMAGES — IBM 029", Kind: "html", HTML: cardsHTML(j, src, rep.Deck)})
	}
	formTab := "FORM RD-114 — RUN AUTHORIZATION"
	if in.Year >= 2005 {
		formTab = "FORM RD-114 — RUN AUTHORIZATION (SCANNED)"
	}
	l.Sections = append(l.Sections, Section{ID: "rd114", Title: "Form RD-114", Tab: formTab, Kind: "html", HTML: rd114HTML(j, l)})
	l.Sections = append(l.Sections, wr.sections...)
	if printerEra {
		l.Sections = append(l.Sections, Section{ID: "trailer", Title: "Separator page (end)", Tab: "PRINT SEPARATOR — END", Kind: "pre", Text: bannerPage(j, l, jobName, userID, true)})
	}

	lname, ltext := buildLicense(j)
	l.License = lname
	l.Sections = append(l.Sections, Section{ID: "license", Title: "License", Tab: "LICENSE — " + strings.ToUpper(lname), Kind: "pre", Text: ltext})

	annotate(l, rep)
	for _, c := range g.roster {
		g.email(c)
	}
	for addr := range g.emails {
		l.Emails = append(l.Emails, addr)
	}
	sort.Strings(l.Emails)
	return l
}

// buildHistory produces a program revision log consistent with the roster
// and never dated after the article.
func buildHistory(j *job) []histEntry {
	x := j.x
	n := x.between(4, 14)
	ros := j.g.roster
	first := ros[0].Start
	if first > j.in.Year {
		first = j.in.Year
	}
	years := make([]int, n)
	for i := range years {
		years[i] = x.between(first, j.in.Year)
	}
	sort.Ints(years)
	years[0] = first
	var out []histEntry
	major, minor := 1, 0
	for i, y := range years {
		c := ros[0]
		for _, cand := range sample(x, ros, len(ros)) {
			if cand.Start <= y && cand.End >= y {
				c = cand
				break
			}
		}
		desc := "Original program."
		if i > 0 {
			for tries := 0; tries < 8; tries++ {
				if t, ok := eraText(pick(x, getLibrary().comments["revision"]), j.in.Year); ok {
					desc = j.g.fill(t, c)
					break
				}
			}
		}
		ver := fmt.Sprintf("%d.%d", major, minor)
		if x.chance(0.2) {
			ver += string(rune('A' + x.intn(3)))
		}
		out = append(out, histEntry{Ver: ver, Date: j.g.dateIn(c, y), Who: c, Desc: desc})
		if x.chance(0.3) {
			major++
			minor = 0
		} else {
			minor += x.between(1, 3)
		}
	}
	article := fmt.Sprintf("%04d-%02d-%02d", j.in.Year, j.in.Month, j.in.Day)
	for _, r := range j.in.Revisions {
		if r.Date <= article && x.chance(0.35) {
			c := pick(x, ros)
			out = append(out, histEntry{Ver: strings.Fields(r.Version + " ?")[0], Date: r.Date, Who: c, Desc: r.Description})
		}
	}
	last := ros[len(ros)-1]
	out = append(out, histEntry{Ver: j.id.Release, Date: fmt.Sprintf("%04d-%02d-%02d", j.in.Year, j.in.Month, j.in.Day), Who: last,
		Desc: pick(x, []string{"Recompiled. No source changes.", "Automated rebuild for record " + strings.ToUpper(j.in.RecordID) + ".", "Production run for " + j.in.Org + "."})})
	return out
}

// emitSource dispatches to the compute language's emitter.
func emitSource(j *job) []Line {
	switch j.id.LangKey {
	case "f4", "f77", "gf":
		return newFortranEmitter(j, j.id.LangKey).emit()
	default:
		if em := otherEmitter(j); em != nil {
			return em
		}
		return newFortranEmitter(j, "f77").emit()
	}
}

func otherEmitter(j *job) []Line {
	switch j.id.LangKey {
	case "c", "perl", "py2":
		return newFreeEmitter(j, j.id.LangKey).emit()
	case "cbtr":
		return newFreeEmitter(j, "c").emit()
	case "perl4":
		e := newFreeEmitter(j, "perl")
		e.oldP = true
		return e.emit()
	case "cobol":
		return newCobolEmitter(j).emit()
	case "basic", "atari":
		return newBasicEmitter(j).emit()
	case "sql", "db2sql", "sqlds", "sqlplus", "tsql", "sqlpl":
		return newSQLEmitter(j).emit()
	case "sas", "spss", "stata", "focus", "nomad", "ramis":
		return newStatEmitter(j).emit()
	case "prolog":
		return newPrologEmitter(j).emit()
	}
	if d := dialectFor(j); d != nil {
		return newDialectEmitter(j, d).emit()
	}
	return nil
}

func linesText(ls []Line) string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

// BuildFresh generates a listing bypassing the cache (tests, determinism checks).
func BuildFresh(in *Input) *Listing { return build(in) }
