package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

// dialectEmitter renders the IR for the block-structured languages that
// share one statement walker: Ada, Pascal, PL/I, Paradox PAL, xBase
// (dBASE / Clipper / FoxPro), Natural, Informix 4GL, Visual Basic / VBA /
// VB.NET, C#, PHP, Pick BASIC, MUMPS and PL/SQL. Each language is a
// dialect: a table of syntax templates plus a few hooks for the parts that
// do not fit a template (program skeleton, declarations, card reading).
type dialectEmitter struct {
	j       *job
	d       *dialect
	out     []Line
	globals map[string]bool
	consts  map[string]string
	arrays  map[string]bool
	params  map[string]bool // dialects that pass arguments through globals
	procs   map[string]*Proc
	cardVar string
	pending []Line // hoisted statements flushed before the next line (dialect.hoist)
	tmpN    map[string]int
	tmpMax  map[string]int
	label   int
}

type dialect struct {
	ind      string
	cmt      string // line comment prefix
	cmtEnd   string // closing text for bracketed comments
	sigil    string // $ (PHP) — prefixed to every variable
	idx1     string // fmt: name, i
	idx2     string // fmt: name, i, j
	intDiv   string // fmt: a, b
	mod      string // fmt: a, b
	dble     string // fmt: a
	ladder   string // fmt: median, multiplier
	rel      map[string]string
	and, or  string
	parenAll bool // MUMPS: strict left-to-right, parenthesise everything
	assign   string
	call     string // fmt: name, args
	callNone string // fmt: name
	stop     string
	ifOpen   string // fmt: cond
	ifElse   string
	ifClose  string
	forUp    string // fmt: v, from, to
	forDown  string // fmt: v, from(high), to(low)
	forClose string // fmt: v
	whileFor bool   // no FOR statement (dBASE, T-SQL style): v = from; WHILE ... ; v = v + 1
	whileHd  string // fmt: cond
	whileEnd string
	dots     bool // MUMPS argumentless-DO dot blocks
	byGlobal bool // arguments assigned to globals before the call (GOSUB / PERFORM)
	upper    bool // keywords and identifiers in upper case

	concat string
	lit    func(s string) string
	pstr   func(x string, w int) string
	pnum   func(x string, w int) string
	psgn   func(x string, w int) string
	print  func(parts string) string

	// hooks
	header    func(e *dialectEmitter)
	decl      func(e *dialectEmitter, s Sym, scope string) []string
	procOpen  func(e *dialectEmitter, p *Proc, isMain bool)
	procClose func(e *dialectEmitter, p *Proc, isMain bool)
	readCard  func(e *dialectEmitter, depth int)
	field     func(e *dialectEmitter, target string, start, width int, num bool) string
	footer    func(e *dialectEmitter)
	order     string // "main-first" (default), "procs-first", "nested"
	bodyDepth func(isMain bool) int
	post      func(e *dialectEmitter) // final pass over the emitted lines

	// optional overrides for languages whose shape the templates above
	// cannot express (prefix / postfix / message-send syntax, statement
	// lowering). Each returns ok=false to fall back to the default.
	argSep   string
	exFn     func(e *dialectEmitter, x *Expr) (string, bool)
	condFn   func(e *dialectEmitter, c *Cond) (string, bool)
	assignFn func(e *dialectEmitter, l *Expr, r string) (string, bool)
	callFn   func(e *dialectEmitter, name string, args []string) string
	stmtFn   func(e *dialectEmitter, s *Stmt, depth int) bool
	// hoist: array element reads are moved into temporaries by a statement
	// emitted just before the line that uses them (languages without array
	// expressions, e.g. classic ABAP internal tables). Format: temp, array, index.
	hoist    string
	hoistTmp func(e *dialectEmitter, typ string, n int) string
	// mainEnds: procClose of the main routine terminates the program, so
	// main's trailing STOP is dropped (STOP is then only the error exit).
	mainEnds bool
}

func newDialectEmitter(j *job, d *dialect) *dialectEmitter {
	e := &dialectEmitter{j: j, d: d, globals: map[string]bool{}, consts: map[string]string{}, arrays: map[string]bool{},
		params: map[string]bool{}, procs: map[string]*Proc{}, cardVar: "card", tmpN: map[string]int{}, tmpMax: map[string]int{}}
	for _, s := range j.p.Globals {
		if s.Role == "const" {
			e.consts[s.Name] = s.Init[0]
		} else {
			e.globals[s.Name] = true
		}
		if len(s.Dims) > 0 {
			e.arrays[s.Name] = true
		}
	}
	for _, p := range j.p.Procs {
		e.procs[p.Name] = p
		for _, l := range p.Locals {
			if len(l.Dims) > 0 {
				e.arrays[l.Name] = true
			}
		}
	}
	return e
}

// ── output ─────────────────────────────────────────────────────────────────

func (e *dialectEmitter) indent(depth int) string {
	if e.d.dots {
		return " " + strings.Repeat(". ", maxi(0, depth-1))
	}
	return strings.Repeat(e.d.ind, depth)
}

func (e *dialectEmitter) line(depth int, text string, refs ...string) {
	for _, p := range e.pending {
		p.Text = strings.TrimRight(e.indent(depth)+p.Text, " ")
		e.out = append(e.out, p)
	}
	e.pending = e.pending[:0]
	for k := range e.tmpN {
		e.tmpN[k] = 0
	}
	if text == "" {
		return
	}
	for _, t := range strings.Split(text, "\n") {
		e.out = append(e.out, Line{Text: strings.TrimRight(e.indent(depth)+t, " "), Code: true, Refs: refs})
	}
}

// newLabel returns a fresh statement label number (GOTO-structured dialects).
func (e *dialectEmitter) newLabel() int {
	e.label++
	return e.label
}

// symType reports the IR type of a named symbol ("int" when unknown).
func (e *dialectEmitter) symType(name string) string {
	for _, g := range e.j.p.Globals {
		if g.Name == name {
			return g.Type
		}
	}
	for _, p := range e.j.p.Procs {
		for _, l := range p.Locals {
			if l.Name == name {
				return l.Type
			}
		}
	}
	return "int"
}

func (e *dialectEmitter) exprType(x *Expr) string {
	switch x.K {
	case "var", "idx":
		return e.symType(x.Name)
	case "str":
		return "str"
	case "dlit":
		return "dbl"
	case "bin":
		if isDouble(x.Args[0]) || isDouble(x.Args[1]) {
			return "dbl"
		}
	}
	return "int"
}

// hoisted moves an array element read into a temporary.
func (e *dialectEmitter) hoisted(x *Expr, rendered string) string {
	typ := e.symType(x.Name)
	e.tmpN[typ]++
	if e.tmpN[typ] > e.tmpMax[typ] {
		e.tmpMax[typ] = e.tmpN[typ]
	}
	t := e.d.hoistTmp(e, typ, e.tmpN[typ])
	var idx []string
	for _, a := range x.Args {
		idx = append(idx, e.ex(a))
	}
	ix := strings.Join(idx, e.argSep())
	if len(x.Args) == 2 {
		ix = fmt.Sprintf("( %s - 1 ) * %s + %s", idx[0], e.dim1(x.Name), idx[1])
	}
	e.pending = append(e.pending, Line{Text: fmt.Sprintf(e.d.hoist, t, e.name(x.Name), ix), Code: true, Refs: []string{x.Name}})
	return t
}

func (e *dialectEmitter) raw(text string) { e.out = append(e.out, Line{Text: text, Code: true}) }

func (e *dialectEmitter) defLine(depth int, text string, defs ...string) {
	e.out = append(e.out, Line{Text: e.indent(depth) + text, Code: true, Defs: defs})
}

func (e *dialectEmitter) blank() {
	if e.d.dots {
		e.out = append(e.out, Line{Text: " ;"})
		return
	}
	e.out = append(e.out, Line{Text: ""})
}

func (e *dialectEmitter) commentLines(depth int, lines []string) {
	pad := e.indent(depth)
	for _, l := range lines {
		if end := strings.TrimSpace(e.d.cmtEnd); end != "" {
			l = strings.ReplaceAll(l, end, "'")
		}
		t := pad + e.d.cmt + l
		if e.d.cmtEnd != "" {
			t = pad + e.d.cmt + padRight(l, e.j.cs.width) + e.d.cmtEnd
		}
		e.out = append(e.out, Line{Text: strings.TrimRight(t, " ")})
	}
}

func (e *dialectEmitter) comment(depth int, text string) {
	cs := e.j.cs
	cs.width = maxi(40, 72-len(e.indent(depth)))
	e.commentLines(depth, cs.para(text))
}

// ── expressions ────────────────────────────────────────────────────────────

func (e *dialectEmitter) name(n string) string {
	if _, ok := e.consts[n]; ok {
		return e.constName(n)
	}
	return e.d.sigil + n
}

func (e *dialectEmitter) constName(n string) string {
	switch e.j.nm.family {
	case "natural", "mumps", "pick", "xbase", "pli", "bc", "dc", "forth", "apl", "clist", "cl400", "ezt", "rpg", "abap", "awk":
		return n
	case "lisp":
		return "+" + strings.Trim(n, "*+") + "+"
	case "smalltalk":
		return strings.ToLower(n[:1]) + n[1:]
	}
	return strings.ToUpper(n)
}

func (e *dialectEmitter) ex(x *Expr) string {
	d := e.d
	if d.exFn != nil {
		if s, ok := d.exFn(e, x); ok {
			return s
		}
	}
	if d.hoist != "" && x.K == "idx" {
		return e.hoisted(x, "")
	}
	switch x.K {
	case "var":
		return e.name(x.Name)
	case "idx":
		if len(x.Args) == 2 {
			f := strings.ReplaceAll(d.idx2, "{D1}", e.dim1(x.Name))
			return fmt.Sprintf(f, e.name(x.Name), e.ex(x.Args[0]), e.ex(x.Args[1]))
		}
		return fmt.Sprintf(d.idx1, e.name(x.Name), e.ex(x.Args[0]))
	case "int":
		return strconv.Itoa(x.Val)
	case "dlit":
		if e.j.nm.family == "mumps" || e.j.nm.family == "pick" {
			return strings.TrimPrefix(x.S, "0")
		}
		return x.S
	case "str":
		return d.lit(x.S)
	case "bin":
		l, r := x.Args[0], x.Args[1]
		ls, rs := e.ex(l), e.ex(r)
		if x.Op == "/" && !isDouble(l) && !isDouble(r) {
			return fmt.Sprintf(d.intDiv, e.wrap(l, ls), e.wrap(r, rs))
		}
		if d.parenAll {
			return e.wrap(l, ls) + x.Op + e.wrap(r, rs)
		}
		if l.K == "bin" && prec(l.Op) < prec(x.Op) {
			ls = "(" + ls + ")"
		}
		if r.K == "bin" && (prec(r.Op) < prec(x.Op) || (prec(r.Op) == prec(x.Op) && (x.Op == "-" || x.Op == "/"))) {
			rs = "(" + rs + ")"
		}
		return ls + " " + x.Op + " " + rs
	case "call":
		switch x.Name {
		case "mod":
			return fmt.Sprintf(d.mod, e.wrap(x.Args[0], e.ex(x.Args[0])), e.wrap(x.Args[1], e.ex(x.Args[1])))
		case "dble":
			return fmt.Sprintf(d.dble, e.wrap(x.Args[0], e.ex(x.Args[0])))
		}
		var args []string
		for _, a := range x.Args {
			args = append(args, e.ex(a))
		}
		return x.Name + "(" + strings.Join(args, e.argSep()) + ")"
	case "ladder":
		return fmt.Sprintf(d.ladder, e.wrap(x.Args[0], e.ex(x.Args[0])), e.wrap(x.Args[1], e.ex(x.Args[1])))
	}
	return "0"
}

func (e *dialectEmitter) argSep() string {
	if e.d.argSep != "" {
		return e.d.argSep
	}
	if e.d.dots {
		return ","
	}
	return ", "
}

// dim1 is the second-dimension bound of a 2-D array (for dialects that
// flatten matrices into vectors).
func (e *dialectEmitter) dim1(name string) string {
	for _, g := range e.j.p.Globals {
		if g.Name == name && len(g.Dims) == 2 {
			return e.dimName(g.Dims[1])
		}
	}
	for _, p := range e.j.p.Procs {
		for _, l := range p.Locals {
			if l.Name == name && len(l.Dims) == 2 {
				return e.dimName(l.Dims[1])
			}
		}
	}
	return "40"
}

// assigns reports whether a routine body assigns to (or loops over) a name.
func assigns(ss []*Stmt, name string) bool {
	for _, s := range ss {
		if s.K == "assign" && s.L != nil && s.L.Name == name {
			return true
		}
		if s.K == "do" && s.Var == name {
			return true
		}
		if assigns(s.Body, name) || assigns(s.Then, name) || assigns(s.Else, name) {
			return true
		}
	}
	return false
}

func (e *dialectEmitter) wrap(x *Expr, s string) string {
	if x.K == "bin" {
		return "(" + s + ")"
	}
	return s
}

func (e *dialectEmitter) cond(c *Cond) string {
	if e.d.condFn != nil {
		if s, ok := e.d.condFn(e, c); ok {
			return s
		}
	}
	switch c.Op {
	case "and", "or":
		op := e.d.and
		if c.Op == "or" {
			op = e.d.or
		}
		a, b := e.cond(c.A), e.cond(c.B)
		if e.d.parenAll || e.j.nm.family == "pascal" {
			return "(" + a + ")" + op + "(" + b + ")"
		}
		return a + op + b
	}
	l, r := e.ex(c.L), e.ex(c.R)
	if e.d.parenAll {
		return l + e.d.rel[c.Op] + r
	}
	return l + " " + e.d.rel[c.Op] + " " + r
}

// ── statements ─────────────────────────────────────────────────────────────

func (e *dialectEmitter) body(ss []*Stmt, depth int) {
	n := 0
	for _, s := range ss {
		if s.K != "comment" && s.K != "blank" {
			n++
		}
		e.one(s, depth)
	}
	if n == 0 {
		switch e.j.nm.family {
		case "ada", "sql":
			e.line(depth, "null;")
		case "pick", "natural":
			e.line(depth, pick(e.j.x, []string{"NULL", "IGNORE"}))
		case "mumps":
			e.line(depth, "Q")
		}
	}
}

func (e *dialectEmitter) one(s *Stmt, depth int) {
	d := e.d
	if d.stmtFn != nil && d.stmtFn(e, s, depth) {
		return
	}
	switch s.K {
	case "comment":
		if s.Banner {
			e.commentLines(depth, e.j.cs.box(s.Title, s.Text))
		} else {
			e.comment(depth, s.Text)
		}
	case "blank":
		e.blank()
	case "assign":
		e.line(depth, e.assignText(s.L, s.R), refsOf(s.L, s.R)...)
	case "call":
		p := e.procs[s.Name]
		var args []string
		for _, a := range s.Args {
			args = append(args, e.ex(a))
		}
		refs := append([]string{s.Name}, refsOf(s.Args...)...)
		if d.byGlobal && p != nil {
			for i, a := range s.Args {
				if i < len(p.Params) {
					e.line(depth, e.assignText(V(p.Params[i]), a), refsOf(a)...)
				}
			}
			if d.callFn != nil {
				e.line(depth, d.callFn(e, s.Name, nil), s.Name)
				return
			}
			e.line(depth, fmt.Sprintf(d.callNone, s.Name), s.Name)
			return
		}
		if d.callFn != nil {
			e.line(depth, d.callFn(e, s.Name, args), refs...)
			return
		}
		if len(args) == 0 {
			e.line(depth, fmt.Sprintf(d.callNone, s.Name), refs...)
		} else {
			e.line(depth, fmt.Sprintf(d.call, s.Name, strings.Join(args, e.argSep())), refs...)
		}
	case "stop":
		e.line(depth, d.stop)
	case "io":
		e.io(s, depth)
	case "do":
		e.do(s, depth)
	case "if":
		refs := condRefs(s.Cond)
		e.line(depth, fmt.Sprintf(d.ifOpen, e.cond(s.Cond)), refs...)
		e.body(s.Then, depth+1)
		if s.Else != nil {
			e.line(depth, d.ifElse)
			e.body(s.Else, depth+1)
		}
		e.line(depth, d.ifClose)
	}
}

// assignText renders l := r, through the dialect's assignFn when it has one.
func (e *dialectEmitter) assignText(l, r *Expr) string {
	rs := e.ex(r)
	if e.d.assignFn != nil {
		if t, ok := e.d.assignFn(e, l, rs); ok {
			return t
		}
	}
	if e.d.hoist != "" && l.K == "idx" {
		// element store: render the target without hoisting it
		h := e.d.hoist
		e.d.hoist = ""
		ls := e.ex(l)
		e.d.hoist = h
		return fmt.Sprintf(e.d.assign, ls, rs)
	}
	return fmt.Sprintf(e.d.assign, e.ex(l), rs)
}

func (e *dialectEmitter) do(s *Stmt, depth int) {
	d := e.d
	v := e.name(s.Var)
	from, to := e.ex(s.From), e.ex(s.To)
	refs := append([]string{s.Var}, refsOf(s.From, s.To)...)
	if d.whileFor {
		rel, step := "le", "+"
		if s.Step < 0 {
			rel, step = "ge", "-"
		}
		e.line(depth, e.assignText(V(s.Var), s.From), refs...)
		e.line(depth, fmt.Sprintf(d.whileHd, e.cond(C(rel, V(s.Var), s.To))), refs...)
		e.body(s.Body, depth+1)
		e.line(depth+1, e.assignText(V(s.Var), Bin(step, V(s.Var), N(1))))
		e.line(depth, d.whileEnd)
		return
	}
	if s.Step < 0 {
		e.line(depth, fmt.Sprintf(d.forDown, v, from, to), refs...)
	} else {
		e.line(depth, fmt.Sprintf(d.forUp, v, from, to), refs...)
	}
	e.body(s.Body, depth+1)
	if strings.Contains(d.forClose, "%") {
		e.line(depth, fmt.Sprintf(d.forClose, v))
	} else if d.forClose != "" {
		e.line(depth, d.forClose)
	}
}

// ── I/O ────────────────────────────────────────────────────────────────────

type pitem struct {
	lit  string
	x    string
	w    int
	kind byte // s n + (signed)
}

func (e *dialectEmitter) printItems(depth int, items []pitem, refs []string) {
	d := e.d
	var parts []string
	for _, it := range items {
		switch {
		case it.x == "":
			if it.lit != "" {
				parts = append(parts, d.lit(it.lit))
			}
		case it.kind == 's':
			parts = append(parts, d.pstr(it.x, it.w))
		case it.kind == '+':
			parts = append(parts, d.psgn(it.x, it.w))
		default:
			parts = append(parts, d.pnum(it.x, it.w))
		}
	}
	kept := parts[:0]
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	parts = kept
	if len(parts) == 0 {
		parts = []string{d.lit(" ")}
	}
	e.line(depth, d.print(strings.Join(parts, d.concat)), refs...)
}

func (e *dialectEmitter) io(s *Stmt, depth int) {
	nm := e.j.nm
	r := e.j.rep
	var a *Expr
	if len(s.Args) > 0 {
		a = s.Args[0]
	}
	I := func(role string) string { return e.ex(Ix(nm.get(role), a)) }
	G := func(role string) string { return e.name(nm.get(role)) }
	fld := func(target string, start, width int, num bool, role string) {
		e.line(depth, e.d.field(e, target, start, width, num), nm.get(role))
	}
	switch s.Name {
	case "readhdr":
		e.d.readCard(e, depth)
		fld(G("org"), 0, 40, false, "org")
		fld(G("recid"), 40, 8, false, "recid")
		e.d.readCard(e, depth)
		fld(G("peer"), 0, 60, false, "peer")
		e.d.readCard(e, depth)
		fld(G("nmet"), 0, 3, true, "nmet")
		fld(G("nben"), 3, 3, true, "nben")
	case "readmet":
		e.d.readCard(e, depth)
		fld(I("mname"), 0, 40, false, "mname")
		fld(I("bsln"), 40, 6, true, "bsln")
		fld(I("curr"), 46, 6, true, "curr")
	case "readben":
		e.d.readCard(e, depth)
		fld(I("bname"), 0, 40, false, "bname")
		fld(I("medn"), 40, 7, true, "medn")
	case "pagehdr":
		e.printItems(depth, []pitem{{lit: r.HdrLeft}, {x: G("recid"), w: 8, kind: 's'}, {lit: r.HdrMid}, {x: G("ipage"), w: 5, kind: 'n'}}, []string{nm.get("recid"), nm.get("ipage")})
	case "title1":
		e.printItems(depth, []pitem{{lit: r.Title1}, {x: G("org"), w: 40, kind: 's'}}, []string{nm.get("org")})
	case "colhdr1":
		e.printItems(depth, []pitem{{lit: r.ColHdr1}}, nil)
	case "rowmet":
		e.printItems(depth, []pitem{{lit: " "}, {x: I("mname"), w: 40, kind: 's'}, {x: I("bsln"), w: 9, kind: 'n'}, {x: I("curr"), w: 9, kind: 'n'}, {x: I("delt"), w: 9, kind: '+'}},
			[]string{nm.get("mname"), nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	case "title2":
		e.printItems(depth, []pitem{{lit: r.Title2}, {x: G("peer"), w: 60, kind: 's'}}, []string{nm.get("peer")})
	case "colhdr2":
		e.printItems(depth, []pitem{{lit: r.ColHdr2}}, nil)
	case "rowben":
		items := []pitem{{lit: " "}, {x: I("bname"), w: 40, kind: 's'}}
		for k := 1; k <= 8; k++ {
			items = append(items, pitem{x: e.ex(Ix(nm.get("pctl"), N(k), a)), w: 9, kind: 'n'})
		}
		e.printItems(depth, items, []string{nm.get("bname"), nm.get("pctl")})
	case "trailer":
		e.printItems(depth, []pitem{{lit: "0RECORDS PROCESSED"}, {x: G("nmet"), w: 6, kind: 'n'}, {lit: " METRICS"}, {x: G("nben"), w: 6, kind: 'n'}, {lit: " BENCHMARKS   SEAL "}, {x: G("chksum"), w: 6, kind: 'n'}},
			[]string{nm.get("nmet"), nm.get("nben"), nm.get("chksum")})
	case "errmsg":
		e.printItems(depth, []pitem{{lit: fmt.Sprintf(" *** %s%03dE %s ***", e.j.id.Program[:3], a.Val*10+e.j.x.intn(10), errText[a.Val])}}, nil)
	case "debugmsg":
		e.printItems(depth, []pitem{{lit: fmt.Sprintf(" DEBUG CHECKPOINT %04d", a.Val)}}, nil)
	case "debugmet":
		e.printItems(depth, []pitem{{lit: " DBG MET"}, {x: e.ex(a), w: 4, kind: 'n'}, {x: I("bsln"), w: 10, kind: 'n'}, {x: I("curr"), w: 10, kind: 'n'}, {x: I("delt"), w: 10, kind: 'n'}}, []string{nm.get("bsln")})
	case "debugben":
		e.printItems(depth, []pitem{{lit: " DBG BEN"}, {x: e.ex(a), w: 4, kind: 'n'}, {x: I("medn"), w: 10, kind: 'n'}}, []string{nm.get("medn")})
	}
}

// ── program structure ──────────────────────────────────────────────────────

func (e *dialectEmitter) dimName(d int) string {
	for _, g := range e.j.p.Globals {
		if g.Role == "const" && g.Init[0] == strconv.Itoa(d) {
			return e.name(g.Name)
		}
	}
	return strconv.Itoa(d)
}

// headerComments is the standard program banner, revision log and seal box.
func (e *dialectEmitter) headerComments() {
	j := e.j
	e.commentLines(0, j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	e.commentLines(0, j.historyLines())
	e.commentLines(0, j.cs.box("Provenance", "Seal "+strings.ToUpper(j.in.Token)+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos")))
}

func (e *dialectEmitter) globalsDecl(scope string) {
	for _, s := range e.j.p.Globals {
		for _, t := range e.d.decl(e, s, scope) {
			e.defLine(0, t, s.Name)
		}
	}
}

// byGlobalSyms: dialects without real parameters declare every routine's
// parameters and locals as program data.
func (e *dialectEmitter) byGlobalSyms() []Sym {
	var out []Sym
	seen := map[string]bool{}
	for _, p := range e.j.p.Procs {
		for _, l := range p.Locals {
			if !seen[l.Name] && !e.globals[l.Name] {
				seen[l.Name] = true
				out = append(out, l)
			}
		}
	}
	return out
}

func (e *dialectEmitter) locals(p *Proc, depth int) {
	if e.d.byGlobal {
		return
	}
	for _, l := range p.Locals {
		if l.Role == "param" {
			continue
		}
		for _, t := range e.d.decl(e, l, "local") {
			e.defLine(depth, t, l.Name)
		}
	}
}

func (e *dialectEmitter) unit(p *Proc, isMain bool, callers, callees []string) {
	e.blank()
	e.commentLines(0, e.j.prologue(p, callers, callees))
	e.d.procOpen(e, p, isMain)
	depth := 1
	if e.d.bodyDepth != nil {
		depth = e.d.bodyDepth(isMain)
	}
	body := p.Body
	if isMain && e.d.mainEnds && len(body) > 0 && body[len(body)-1].K == "stop" {
		body = body[:len(body)-1]
	}
	e.body(body, depth)
	e.d.procClose(e, p, isMain)
}

func (e *dialectEmitter) emit() []Line {
	j := e.j
	callers, callees := callGraph(j.p)
	e.d.header(e)
	main := j.p.Procs[0]
	rest := j.p.Procs[1:]
	if e.d.order == "topo" {
		rest = topoProcs(rest, callees)
	}
	switch e.d.order {
	case "procs-first", "nested", "topo":
		for _, p := range rest {
			e.unit(p, false, callers[p.Name], callees[p.Name])
		}
		e.unit(main, true, nil, callees[main.Name])
	default:
		e.unit(main, true, nil, callees[main.Name])
		for _, p := range rest {
			e.unit(p, false, callers[p.Name], callees[p.Name])
		}
	}
	if e.d.footer != nil {
		e.d.footer(e)
	}
	if e.d.post != nil {
		e.d.post(e)
	}
	return e.out
}

// vecLadder recognises the inner ladder loop
//
//	DO k = 1, NLADR: PCTL(k, j) = INT(MEDN(j) * MULT(k))
//
// so array languages can write it as one whole-vector statement. It returns
// the target array, the outer subscript, the median expression and the
// multiplier array.
func vecLadder(s *Stmt) (tgt string, outer, med *Expr, mult string, ok bool) {
	if s.K != "do" || s.Step != 1 {
		return
	}
	var body []*Stmt
	for _, b := range s.Body {
		if b.K != "comment" && b.K != "blank" {
			body = append(body, b)
		}
	}
	if len(body) != 1 || body[0].K != "assign" {
		return
	}
	a := body[0]
	if a.R.K != "ladder" || a.L.K != "idx" || len(a.L.Args) != 2 || a.L.Args[0].K != "var" || a.L.Args[0].Name != s.Var {
		return
	}
	m := a.R.Args[1]
	if m.K != "idx" || len(m.Args) != 1 || m.Args[0].K != "var" || m.Args[0].Name != s.Var {
		return
	}
	for _, r := range refsOf(a.R.Args[0], a.L.Args[1]) {
		if r == s.Var {
			return
		}
	}
	return a.L.Name, a.L.Args[1], a.R.Args[0], m.Name, true
}

// topoProcs orders routines so every callee precedes its callers
// (languages that compile definitions in one pass: Forth, dc, Lisp).
func topoProcs(ps []*Proc, callees map[string][]string) []*Proc {
	byName := map[string]*Proc{}
	for _, p := range ps {
		byName[p.Name] = p
	}
	done := map[string]bool{}
	var out []*Proc
	var visit func(p *Proc)
	visit = func(p *Proc) {
		if done[p.Name] {
			return
		}
		done[p.Name] = true
		for _, c := range callees[p.Name] {
			if q, ok := byName[c]; ok {
				visit(q)
			}
		}
		out = append(out, p)
	}
	for _, p := range ps {
		visit(p)
	}
	return out
}

// usedGlobals lists the scalar/array globals a procedure body touches (PHP
// needs a `global` statement for them).
func (e *dialectEmitter) usedGlobals(p *Proc) []string {
	seen := map[string]bool{}
	var out []string
	var walkS func(ss []*Stmt)
	add := func(es ...*Expr) {
		for _, n := range refsOf(es...) {
			if e.globals[n] && !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	walkS = func(ss []*Stmt) {
		for _, s := range ss {
			add(s.L, s.R, s.From, s.To)
			add(s.Args...)
			if s.Cond != nil {
				for _, n := range condRefs(s.Cond) {
					if e.globals[n] && !seen[n] {
						seen[n] = true
						out = append(out, n)
					}
				}
			}
			if s.K == "do" && e.globals[s.Var] && !seen[s.Var] {
				seen[s.Var] = true
				out = append(out, s.Var)
			}
			if s.K == "io" {
				for _, role := range []string{"org", "recid", "peer", "nmet", "nben", "ipage", "chksum", "mname", "bname", "bsln", "curr", "delt", "medn", "pctl"} {
					n := e.j.nm.get(role)
					if e.globals[n] && !seen[n] {
						seen[n] = true
						out = append(out, n)
					}
				}
			}
			walkS(s.Body)
			walkS(s.Then)
			walkS(s.Else)
		}
	}
	walkS(p.Body)
	sortStrings(out)
	return out
}

func (e *dialectEmitter) paramSym(p *Proc, q string) Sym {
	for _, l := range p.Locals {
		if l.Name == q {
			return l
		}
	}
	return Sym{Name: q, Type: "int"}
}

func dims(s Sym) int { return len(s.Dims) }

// quoting helpers
func dq(s string) string { return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\"" }
func sq(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func bsq(s string) string {
	return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(s) + "'"
}

// recaseKeywords rewrites the given keywords in code lines (outside string
// literals and comments) with f; for case-insensitive languages whose shops
// each had their own house style.
func recaseKeywords(e *dialectEmitter, kws []string, f func(string) string, quote byte, cmtOpen string) {
	set := map[string]bool{}
	for _, k := range kws {
		set[strings.ToLower(k)] = true
	}
	for i, l := range e.out {
		if !l.Code {
			continue
		}
		var b strings.Builder
		t := l.Text
		inQ := false
		for k := 0; k < len(t); {
			c := t[k]
			if !inQ && cmtOpen != "" && strings.HasPrefix(t[k:], cmtOpen) {
				b.WriteString(t[k:])
				break
			}
			if c == quote {
				inQ = !inQ
				b.WriteByte(c)
				k++
				continue
			}
			if !inQ && (c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_') {
				m := k
				for m < len(t) && (t[m] >= 'A' && t[m] <= 'Z' || t[m] >= 'a' && t[m] <= 'z' || t[m] >= '0' && t[m] <= '9' || t[m] == '_') {
					m++
				}
				w := t[k:m]
				if set[strings.ToLower(w)] {
					w = f(strings.ToLower(w))
				}
				b.WriteString(w)
				k = m
				continue
			}
			b.WriteByte(c)
			k++
		}
		e.out[i].Text = b.String()
	}
}
