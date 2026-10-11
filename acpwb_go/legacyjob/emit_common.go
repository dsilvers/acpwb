package legacyjob

import (
	"fmt"
	"sort"
	"strings"
)

// Line is one emitted source line plus what the compiler-listing generator
// needs to know about it.
type Line struct {
	Text  string
	Code  bool     // a statement (vs comment / blank)
	Refs  []string // symbols referenced on this line
	Defs  []string // symbols declared on this line
	Label string   // statement label / paragraph / line number
}

// histEntry is one revision-history line in a program prologue.
type histEntry struct {
	Ver, Date string
	Who       *Coder
	Desc      string
}

// job bundles everything an emitter needs.
type job struct {
	g     *gen
	x     *rnd
	in    *Input
	id    ID
	b     *builder
	p     *Program
	nm    *naming
	rep   *report
	hist  []histEntry
	cs    cstyle
	paths *jobPaths
}

// ── comment styling ─────────────────────────────────────────────────────────

type cstyle struct {
	width  int  // usable text width
	upper  bool // fold to uppercase (card-era languages)
	banner int  // banner style 0..11
	fill   byte // box fill character
}

func newCStyle(x *rnd, width int, upper bool) cstyle {
	return cstyle{width: width, upper: upper, banner: x.intn(12), fill: pick(x, []byte{'*', '*', '-', '=', '+', '#'})}
}

func (cs cstyle) cased(s string) string {
	if cs.upper {
		return cardText(s)
	}
	return asciiText(s)
}

// para wraps a paragraph to the comment width.
func (cs cstyle) para(text string) []string {
	return wrapWords(cs.cased(text), cs.width)
}

// box draws a titled banner around optional body text, returning the inner
// content lines (the emitter adds its comment prefix).
func (cs cstyle) box(title, text string) []string {
	w := cs.width
	title = trunc(cs.cased(title), w-24)
	fill := string(cs.fill)
	var out []string
	body := cs.para(text)
	switch cs.banner {
	case 0, 1:
		rule := strings.Repeat(fill, w)
		out = append(out, rule)
		if title != "" {
			out = append(out, centre(title, w))
		}
		if len(body) > 0 {
			out = append(out, "")
			out = append(out, body...)
		}
		out = append(out, rule)
	case 2, 3:
		out = append(out, fill+fill+fill+" "+title+" "+strings.Repeat(fill, maxi(3, w-len(title)-5)))
		out = append(out, body...)
	case 4: // IBM-style framed
		rule := "+" + strings.Repeat("-", w-2) + "+"
		out = append(out, rule)
		out = append(out, "| "+padRight(title, w-4)+" |")
		for _, l := range body {
			out = append(out, "| "+padRight(l, w-4)+" |")
		}
		out = append(out, rule)
	case 5:
		out = append(out, strings.Repeat(fill+" ", w/2))
		out = append(out, "   "+spaced(title))
		out = append(out, body...)
		out = append(out, strings.Repeat(fill+" ", w/2))
	case 6:
		out = append(out, title)
		out = append(out, strings.Repeat("-", len(title)))
		out = append(out, body...)
	case 7:
		rule := strings.Repeat(fill, w)
		out = append(out, rule, rule)
		out = append(out, fill+fill+" "+padRight(title, w-6)+" "+fill+fill)
		for _, l := range body {
			out = append(out, fill+fill+" "+padRight(l, w-6)+" "+fill+fill)
		}
		out = append(out, rule, rule)
	case 8:
		out = append(out, ">>>>> "+title+" <<<<<")
		out = append(out, body...)
		out = append(out, "<<<<< END "+title+" >>>>>")
	case 9:
		out = append(out, "SECTION: "+title)
		out = append(out, body...)
	case 10:
		out = append(out, strings.Repeat("=", w))
		out = append(out, "== "+title)
		out = append(out, strings.Repeat("=", w))
		out = append(out, body...)
	default:
		out = append(out, "")
		out = append(out, "   *** "+title+" ***")
		out = append(out, body...)
		out = append(out, "")
	}
	return out
}

func centre(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return strings.Repeat(" ", (w-len(s))/2) + s
}

func spaced(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// prologue renders a routine header block (content lines only).
func (j *job) prologue(p *Proc, callers, callees []string) []string {
	cs := j.cs
	x := j.x
	var body []string
	field := func(k, v string) {
		lines := wrapWords(cs.cased(v), cs.width-14)
		for i, l := range lines {
			if i == 0 {
				body = append(body, padRight(cs.cased(k)+":", 14)+l)
			} else {
				body = append(body, strings.Repeat(" ", 14)+l)
			}
		}
	}
	field("Routine", p.Name)
	field("Purpose", p.Purpose)
	if p.Author != nil {
		field("Author", fmt.Sprintf("%s, %s", p.Author.Display(x), j.g.date(p.Author, p.Author.Start, p.Author.End)))
	}
	if len(p.Params) > 0 {
		field("Arguments", strings.Join(p.Params, ", "))
	}
	if len(callers) > 0 {
		field("Called by", strings.Join(callers, ", "))
	} else if p.Dead {
		field("Called by", pick(x, []string{"(none)", "nothing, since " + fmt.Sprint(x.between(1975, j.in.Year)), "unknown", "see " + pick(x, programNames)}))
	}
	if len(callees) > 0 {
		field("Calls", strings.Join(callees, ", "))
	}
	if x.chance(0.6) {
		field("Inputs", pick(x, []string{"common block data", "card images on unit 5", "arguments only", "none worth mentioning", "the work arrays"}))
	}
	if x.chance(0.6) {
		field("Outputs", pick(x, []string{"report lines on unit 6", "updated work arrays", "return code", "nothing", "the provenance seal"}))
	}
	for n := x.weighted([]int{6, 3, 1}); n > 0; n-- {
		field("Note", j.g.comment(p.Author))
	}
	if p.Dead && x.chance(0.7) {
		field("Status", pick(x, []string{"RETAINED FOR AUDIT", "DEAD - DO NOT CALL", "OBSOLETE", "DISABLED", "UNREFERENCED", "SEE NOTE"}))
	}
	return cs.boxLines(strings.ToUpper(p.Name), body)
}

// boxLines frames pre-formatted lines (no re-wrapping).
func (cs cstyle) boxLines(title string, body []string) []string {
	w := cs.width
	fill := string(cs.fill)
	switch cs.banner % 3 {
	case 0:
		out := []string{strings.Repeat(fill, w), centre(title, w), strings.Repeat(fill, w)}
		out = append(out, body...)
		return append(out, strings.Repeat(fill, w))
	case 1:
		out := []string{"+" + strings.Repeat("-", w-2) + "+"}
		for _, l := range body {
			out = append(out, "| "+padRight(l, w-4)+" |")
		}
		return append(out, "+"+strings.Repeat("-", w-2)+"+")
	default:
		out := []string{fill + fill + fill + " " + title + " " + strings.Repeat(fill, maxi(3, w-len(title)-5))}
		out = append(out, body...)
		return append(out, strings.Repeat(fill, w))
	}
}

// history renders the program revision log (content lines only).
func (j *job) historyLines() []string {
	cs := j.cs
	var out []string
	out = append(out, cs.cased("Revision history"), cs.cased("----------------"))
	for _, h := range j.hist {
		who := cs.cased(h.Who.Initials)
		ver := strings.Fields(cs.cased(h.Ver) + " ?")[0]
		head := fmt.Sprintf("%-5s %-11s %-6s ", trunc(ver, 5), h.Date, who)
		desc := wrapWords(cs.cased(h.Desc), maxi(20, cs.width-len(head)))
		for i, d := range desc {
			if i == 0 {
				out = append(out, head+d)
			} else {
				out = append(out, strings.Repeat(" ", len(head))+d)
			}
		}
	}
	return out
}

// callGraph returns callers/callees per procedure name.
func callGraph(p *Program) (callers, callees map[string][]string) {
	callers, callees = map[string][]string{}, map[string][]string{}
	var walk func(owner string, ss []*Stmt)
	walk = func(owner string, ss []*Stmt) {
		for _, s := range ss {
			if s.K == "call" {
				if !contains(callees[owner], s.Name) {
					callees[owner] = append(callees[owner], s.Name)
				}
				if !contains(callers[s.Name], owner) {
					callers[s.Name] = append(callers[s.Name], owner)
				}
			}
			walk(owner, s.Body)
			walk(owner, s.Then)
			walk(owner, s.Else)
		}
	}
	for _, pr := range p.Procs {
		walk(pr.Name, pr.Body)
	}
	for k := range callers {
		sort.Strings(callers[k])
	}
	return
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// refsOf collects variable names referenced by expressions.
func refsOf(es ...*Expr) []string {
	var out []string
	for _, e := range es {
		walkExpr(e, func(x *Expr) {
			if (x.K == "var" || x.K == "idx") && !contains(out, x.Name) {
				out = append(out, x.Name)
			}
		})
	}
	return out
}

func condRefs(c *Cond) []string {
	if c == nil {
		return nil
	}
	if c.Op == "and" || c.Op == "or" {
		return append(condRefs(c.A), condRefs(c.B)...)
	}
	return refsOf(c.L, c.R)
}

func symOf(p *Program, name string) *Sym {
	for i := range p.Globals {
		if p.Globals[i].Name == name {
			return &p.Globals[i]
		}
	}
	return nil
}

func roleSym(p *Program, role string) *Sym {
	for i := range p.Globals {
		if p.Globals[i].Role == role {
			return &p.Globals[i]
		}
	}
	return nil
}
