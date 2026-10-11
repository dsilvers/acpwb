package legacyjob

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type fortranEmitter struct {
	j       *job
	dialect string // f4 f77 gf
	out     []Line

	labelNext, labelStep int
	labelChaos           bool
	fmtNext              int
	cont                 string
	cchar                string
	indentW              int
	labelRight           bool
	fmtInline            bool
	gotoLoops            float64
	arithIf              float64
	modernDo             float64
	lower                bool
	declPerLine          int

	procFmts  []Line
	fmtLabels map[string]int
	seq       bool
	seqPrefix string
	seqStep   int
	dropped   bool
	strBlocks map[string]string
}

func newFortranEmitter(j *job, dialect string) *fortranEmitter {
	x := j.x
	f := &fortranEmitter{j: j, dialect: dialect}
	f.labelStep = pick(x, []int{5, 10, 10, 10, 100, 1})
	f.labelNext = f.labelStep
	if f.labelStep == 1 {
		f.labelNext = x.between(1, 20)
		f.labelChaos = true
		f.labelStep = 7
	}
	f.fmtNext = pick(x, []int{9000, 900, 800, 9100, 500})
	f.cont = pick(x, []string{"1", "&", "$", "*", "+", "X", "C"})
	if f.cont == "C" {
		f.cont = "2"
	}
	f.cchar = pick(x, []string{"C", "C", "C", "*", "c"})
	f.indentW = pick(x, []int{0, 1, 2, 3, 3, 4})
	f.labelRight = x.chance(0.7)
	f.fmtInline = x.chance(0.5)
	f.gotoLoops = []float64{0, 0.1, 0.3}[x.intn(3)]
	f.arithIf = []float64{0, 0.2, 0.5}[x.intn(3)]
	f.declPerLine = x.between(1, 5)
	switch dialect {
	case "f4":
		f.gotoLoops += 0.1
		f.arithIf += 0.2
	case "gf":
		f.modernDo = []float64{0.3, 0.7, 1}[x.intn(3)]
		f.lower = x.chance(0.4)
		if f.lower {
			f.cchar = pick(x, []string{"c", "!", "C"})
		}
	}
	f.seq = x.chance(0.85)
	f.seqPrefix = j.id.Program[:3]
	f.seqStep = pick(x, []int{10, 10, 100, 5, 20})
	f.dropped = x.chance(0.1)
	f.strBlocks = map[string]string{}
	return f
}

func (f *fortranEmitter) kw(s string) string {
	if f.lower {
		return strings.ToLower(s)
	}
	return s
}

func (f *fortranEmitter) nextLabel() int {
	l := f.labelNext
	if f.labelChaos {
		f.labelNext += f.j.x.between(1, 40)
	} else {
		f.labelNext += f.labelStep
	}
	if f.labelNext > 8999 {
		f.labelNext = 10
	}
	return l
}

func (f *fortranEmitter) labelField(l int) string {
	if l == 0 {
		return "     "
	}
	s := strconv.Itoa(l)
	if f.labelRight {
		return padLeft(s, 5)
	}
	return padRight(s, 5)
}

// stmt emits a fixed-form statement, splitting it across continuation lines
// at column 72.
func (f *fortranEmitter) stmt(label int, depth int, text string, refs []string) {
	f.stmtInto(&f.out, label, depth, text, refs)
}

func (f *fortranEmitter) stmtInto(dst *[]Line, label int, depth int, text string, refs []string) {
	ind := strings.Repeat(" ", f.indentW*depth)
	body := ind + text
	first := true
	for len(body) > 0 {
		chunk := body
		if len(chunk) > 66 {
			cut := 66
			for c := 65; c > 40; c-- {
				if body[c] == ',' || body[c] == ' ' || body[c] == '+' || body[c] == '*' {
					cut = c + 1
					break
				}
			}
			chunk = body[:cut]
		}
		body = body[len(chunk):]
		lf := "     "
		cc := " "
		lbl := ""
		if first {
			if label != 0 {
				lbl = strconv.Itoa(label)
			}
			lf = f.labelField(label)
		} else {
			cc = f.cont
			if f.cont == "1" {
				cc = strconv.Itoa(minInt(9, len(*dst)%9+1))
			}
			if len(body) > 0 || !first {
				chunk = ind + "  " + strings.TrimLeft(chunk, " ")
			}
		}
		*dst = append(*dst, Line{Text: lf + cc + chunk, Code: true, Refs: refs, Label: lbl})
		first = false
		refs = nil
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (f *fortranEmitter) comment(text string) {
	for _, l := range f.j.cs.para(text) {
		f.out = append(f.out, Line{Text: f.cchar + "     " + l})
	}
}

func (f *fortranEmitter) commentLines(lines []string) {
	for _, l := range lines {
		f.out = append(f.out, Line{Text: f.cchar + " " + l})
	}
}

// ── expressions ─────────────────────────────────────────────────────────────

func prec(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 3
}

func (f *fortranEmitter) isConst(name string) (string, bool) {
	s := symOf(f.j.p, name)
	if s != nil && s.Role == "const" {
		return s.Init[0], true
	}
	return "", false
}

func (f *fortranEmitter) ex(e *Expr) string {
	switch e.K {
	case "var":
		if v, ok := f.isConst(e.Name); ok && f.dialect == "f4" {
			return v
		}
		return f.kw(e.Name)
	case "idx":
		var parts []string
		for _, a := range e.Args {
			parts = append(parts, f.ex(a))
		}
		return f.kw(e.Name) + "(" + strings.Join(parts, ",") + ")"
	case "int":
		if e.Val < 0 {
			return "(" + strconv.Itoa(e.Val) + ")"
		}
		return strconv.Itoa(e.Val)
	case "dlit":
		s := e.S
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s + f.kw("D0")
	case "bin":
		l, r := e.Args[0], e.Args[1]
		ls, rs := f.ex(l), f.ex(r)
		if l.K == "bin" && prec(l.Op) < prec(e.Op) {
			ls = "(" + ls + ")"
		}
		if r.K == "bin" && (prec(r.Op) < prec(e.Op) || (prec(r.Op) == prec(e.Op) && (e.Op == "-" || e.Op == "/"))) {
			rs = "(" + rs + ")"
		}
		return ls + e.Op + rs
	case "call":
		var parts []string
		for _, a := range e.Args {
			parts = append(parts, f.ex(a))
		}
		return f.kw(strings.ToUpper(e.Name)) + "(" + strings.Join(parts, ",") + ")"
	case "ladder":
		fn := "INT"
		if f.dialect == "f4" {
			fn = "IDINT"
		}
		return f.kw(fn) + "(" + f.kw("DBLE") + "(" + f.ex(e.Args[0]) + ")*" + f.ex(e.Args[1]) + ")"
	case "str":
		return "'" + e.S + "'"
	}
	return "?"
}

var relOps = map[string]string{"lt": ".LT.", "le": ".LE.", "eq": ".EQ.", "ne": ".NE.", "gt": ".GT.", "ge": ".GE."}
var negRel = map[string]string{"lt": "ge", "le": "gt", "eq": "ne", "ne": "eq", "gt": "le", "ge": "lt"}

func (f *fortranEmitter) cond(c *Cond) string {
	switch c.Op {
	case "and":
		return f.cond(c.A) + f.kw(".AND.") + f.cond(c.B)
	case "or":
		return f.cond(c.A) + f.kw(".OR.") + f.cond(c.B)
	}
	return f.ex(c.L) + " " + f.kw(relOps[c.Op]) + " " + f.ex(c.R)
}

func negate(c *Cond) *Cond {
	switch c.Op {
	case "and":
		return Or(negate(c.A), negate(c.B))
	case "or":
		return And(negate(c.A), negate(c.B))
	}
	return C(negRel[c.Op], c.L, c.R)
}

// ── statements ──────────────────────────────────────────────────────────────

func (f *fortranEmitter) body(ss []*Stmt, depth int) {
	for _, s := range ss {
		f.one(s, depth)
	}
}

func (f *fortranEmitter) one(s *Stmt, depth int) {
	x := f.j.x
	switch s.K {
	case "comment":
		if s.Banner {
			f.commentLines(f.j.cs.box(s.Title, s.Text))
		} else {
			f.comment(s.Text)
		}
	case "blank":
		f.out = append(f.out, Line{Text: f.cchar})
	case "assign":
		f.stmt(0, depth, f.ex(s.L)+" = "+f.ex(s.R), refsOf(s.L, s.R))
	case "call":
		var args []string
		for _, a := range s.Args {
			args = append(args, f.ex(a))
		}
		t := f.kw("CALL ") + f.kw(s.Name)
		if len(args) > 0 {
			t += "(" + strings.Join(args, ", ") + ")"
		}
		f.stmt(0, depth, t, append([]string{s.Name}, refsOf(s.Args...)...))
	case "stop":
		f.stmt(0, depth, f.kw(pick(x, []string{"STOP", "STOP", "STOP 0", "STOP 16", "STOP 4"})), nil)
	case "return":
		f.stmt(0, depth, f.kw("RETURN"), nil)
	case "io":
		f.io(s, depth)
	case "do":
		f.do(s, depth)
	case "if":
		f.ifStmt(s, depth)
	}
}

func (f *fortranEmitter) do(s *Stmt, depth int) {
	x := f.j.x
	v := f.kw(s.Var)
	from, to := f.ex(s.From), f.ex(s.To)
	refs := append([]string{s.Var}, refsOf(s.From, s.To)...)
	useGoto := x.float() < f.gotoLoops || (f.dialect == "f4" && s.Step < 0)
	if useGoto {
		top, done := f.nextLabel(), f.nextLabel()
		cmp := ".GT."
		inc := v + " + 1"
		if s.Step < 0 {
			cmp = ".LT."
			inc = v + " - 1"
		}
		f.stmt(0, depth, v+" = "+from, refs)
		f.stmt(top, depth, f.kw("IF")+" ("+v+" "+f.kw(cmp)+" "+to+") "+f.kw("GO TO ")+strconv.Itoa(done), refs)
		f.body(s.Body, depth+1)
		f.stmt(0, depth+1, v+" = "+inc, []string{s.Var})
		f.stmt(0, depth+1, f.kw("GO TO ")+strconv.Itoa(top), nil)
		f.stmt(done, depth, f.kw("CONTINUE"), nil)
		return
	}
	step := ""
	if s.Step < 0 {
		step = ", -1"
	}
	if f.dialect == "gf" && x.float() < f.modernDo {
		f.stmt(0, depth, f.kw("DO ")+v+" = "+from+", "+to+step, refs)
		f.body(s.Body, depth+1)
		f.stmt(0, depth, f.kw("END DO"), nil)
		return
	}
	lbl := f.nextLabel()
	f.stmt(0, depth, f.kw("DO ")+strconv.Itoa(lbl)+" "+v+" = "+from+", "+to+step, refs)
	f.body(s.Body, depth+1)
	f.stmt(lbl, depth, f.kw("CONTINUE"), nil)
}

func simpleStmt(ss []*Stmt) *Stmt {
	var only *Stmt
	for _, s := range ss {
		if s.K == "comment" || s.K == "blank" {
			continue
		}
		if only != nil {
			return nil
		}
		only = s
	}
	if only != nil && (only.K == "assign" || only.K == "call" || only.K == "stop") {
		return only
	}
	return nil
}

func (f *fortranEmitter) inline(s *Stmt) string {
	var tmp []Line
	save := f.out
	f.out = nil
	f.one(s, 0)
	tmp = f.out
	f.out = save
	if len(tmp) != 1 {
		return ""
	}
	return strings.TrimSpace(tmp[0].Text[6:])
}

func (f *fortranEmitter) ifStmt(s *Stmt, depth int) {
	x := f.j.x
	refs := condRefs(s.Cond)
	single := simpleStmt(s.Then)
	// arithmetic IF for simple integer comparisons
	if s.Else == nil && s.Cond.Op != "and" && s.Cond.Op != "or" && x.float() < f.arithIf && f.dialect != "gf" {
		then, skip := f.nextLabel(), f.nextLabel()
		lt, eq, gt := skip, skip, skip
		switch s.Cond.Op {
		case "lt":
			lt = then
		case "le":
			lt, eq = then, then
		case "eq":
			eq = then
		case "ne":
			lt, gt = then, then
		case "gt":
			gt = then
		case "ge":
			eq, gt = then, then
		}
		f.stmt(0, depth, fmt.Sprintf("%s (%s-(%s)) %d,%d,%d", f.kw("IF"), f.ex(s.Cond.L), f.ex(s.Cond.R), lt, eq, gt), refs)
		f.stmt(then, depth, f.kw("CONTINUE"), nil)
		f.body(s.Then, depth+1)
		f.stmt(skip, depth, f.kw("CONTINUE"), nil)
		return
	}
	if s.Else == nil && single != nil && (f.dialect == "f4" || x.chance(0.5)) {
		if in := f.inline(single); in != "" {
			f.stmt(0, depth, f.kw("IF")+" ("+f.cond(s.Cond)+") "+in, append(refs, single.Name))
			return
		}
	}
	if f.dialect == "f4" {
		skip := f.nextLabel()
		f.stmt(0, depth, f.kw("IF")+" ("+f.cond(negate(s.Cond))+") "+f.kw("GO TO ")+strconv.Itoa(skip), refs)
		f.body(s.Then, depth+1)
		if s.Else != nil {
			end := f.nextLabel()
			f.stmt(0, depth+1, f.kw("GO TO ")+strconv.Itoa(end), nil)
			f.stmt(skip, depth, f.kw("CONTINUE"), nil)
			f.body(s.Else, depth+1)
			f.stmt(end, depth, f.kw("CONTINUE"), nil)
			return
		}
		f.stmt(skip, depth, f.kw("CONTINUE"), nil)
		return
	}
	f.stmt(0, depth, f.kw("IF")+" ("+f.cond(s.Cond)+") "+f.kw("THEN"), refs)
	f.body(s.Then, depth+1)
	if s.Else != nil {
		f.stmt(0, depth, f.kw("ELSE"), nil)
		f.body(s.Else, depth+1)
	}
	f.stmt(0, depth, f.kw(pick(x, []string{"END IF", "ENDIF"})), nil)
}

// ── I/O ─────────────────────────────────────────────────────────────────────

// lit renders a FORMAT literal: Hollerith for FORTRAN IV, quoted otherwise,
// with runs of blanks compressed to nX edit descriptors.
func (f *fortranEmitter) lit(s string) string {
	var parts []string
	for len(s) > 0 {
		i := strings.Index(s, "    ")
		if i < 0 {
			parts = append(parts, f.quote(s))
			break
		}
		if i > 0 {
			parts = append(parts, f.quote(s[:i]))
		}
		j := i
		for j < len(s) && s[j] == ' ' {
			j++
		}
		parts = append(parts, strconv.Itoa(j-i)+"X")
		s = s[j:]
	}
	return strings.Join(parts, ",")
}

func (f *fortranEmitter) quote(s string) string {
	if f.dialect == "f4" {
		return strconv.Itoa(len(s)) + "H" + s
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (f *fortranEmitter) words(n int) int { return (n + 3) / 4 }

// strRef renders a string variable reference for READ/WRITE lists: plain
// for CHARACTER dialects, implied-DO over the A4 words for FORTRAN IV.
func (f *fortranEmitter) strRef(role string, idx *Expr) (string, string) {
	s := roleSym(f.j.p, role)
	name := f.kw(s.Name)
	if f.dialect != "f4" {
		if idx != nil {
			return name + "(" + f.ex(idx) + ")", "A" + strconv.Itoa(s.Len)
		}
		return name, "A" + strconv.Itoa(s.Len)
	}
	l := f.kw(f.j.nm.get("l"))
	w := f.words(s.Len)
	if idx != nil {
		return fmt.Sprintf("(%s(%s,%s),%s=1,%d)", name, l, f.ex(idx), l, w), fmt.Sprintf("%dA4", w)
	}
	return fmt.Sprintf("(%s(%s),%s=1,%d)", name, l, l, w), fmt.Sprintf("%dA4", w)
}

func (f *fortranEmitter) fmtLabel(spec string) int {
	if l, ok := f.fmtLabels[spec]; ok {
		return l
	}
	l := f.fmtNext
	f.fmtNext += pick(f.j.x, []int{1, 1, 5, 10})
	f.fmtLabels[spec] = l
	line := f.kw("FORMAT") + " (" + spec + ")"
	if f.fmtInline {
		f.stmt(l, 0, line, nil)
	} else {
		f.stmtInto(&f.procFmts, l, 0, line, nil)
	}
	return l
}

func (f *fortranEmitter) rw(depth int, unit int, list []string, spec string, refs []string) {
	verb := "WRITE"
	if unit == 5 {
		verb = "READ"
	}
	l := f.fmtLabel(spec)
	t := fmt.Sprintf("%s (%d,%d)", f.kw(verb), unit, l)
	if len(list) > 0 {
		t += " " + strings.Join(list, ", ")
	}
	f.stmt(0, depth, t, refs)
}

var errText = map[int]string{
	1: "VALUE OUT OF RANGE ON METRIC CARD",
	2: "BENCHMARK MEDIAN NOT POSITIVE",
	3: "CARD COUNT EXCEEDS TABLE SIZE",
	4: "PERCENTILE LADDER CROSS-CHECK FAILED",
}

func (f *fortranEmitter) io(s *Stmt, depth int) {
	j := f.j
	nm := j.nm
	r := j.rep
	K := func(role string) string { return f.kw(nm.get(role)) }
	var a *Expr
	if len(s.Args) > 0 {
		a = s.Args[0]
	}
	switch s.Name {
	case "readhdr":
		o, of := f.strRef("org", nil)
		ri, rf := f.strRef("recid", nil)
		f.rw(depth, 5, []string{o, ri}, of+","+rf, []string{nm.get("org"), nm.get("recid")})
		p, pf := f.strRef("peer", nil)
		f.rw(depth, 5, []string{p}, pf, []string{nm.get("peer")})
		f.rw(depth, 5, []string{K("nmet"), K("nben")}, "2I3", []string{nm.get("nmet"), nm.get("nben")})
	case "readmet":
		mn, mf := f.strRef("mname", a)
		f.rw(depth, 5, []string{mn, f.ex(Ix(nm.get("bsln"), a)), f.ex(Ix(nm.get("curr"), a))}, mf+",2I6",
			[]string{nm.get("mname"), nm.get("bsln"), nm.get("curr")})
	case "readben":
		bn, bf := f.strRef("bname", a)
		f.rw(depth, 5, []string{bn, f.ex(Ix(nm.get("medn"), a))}, bf+",I7", []string{nm.get("bname"), nm.get("medn")})
	case "pagehdr":
		ri, rf := f.strRef("recid", nil)
		f.rw(depth, 6, []string{ri, K("ipage")}, f.lit(r.HdrLeft)+","+rf+","+f.lit(r.HdrMid)+",I5", []string{nm.get("recid"), nm.get("ipage")})
	case "title1":
		o, of := f.strRef("org", nil)
		f.rw(depth, 6, []string{o}, f.lit(r.Title1)+","+of, []string{nm.get("org")})
	case "colhdr1":
		f.rw(depth, 6, nil, f.lit(r.ColHdr1), nil)
	case "rowmet":
		mn, mf := f.strRef("mname", a)
		sign := "SP,"
		if f.dialect == "f4" {
			sign = ""
		}
		f.rw(depth, 6, []string{mn, f.ex(Ix(nm.get("bsln"), a)), f.ex(Ix(nm.get("curr"), a)), f.ex(Ix(nm.get("delt"), a))},
			"1X,"+mf+",2I9,"+sign+"I9", []string{nm.get("mname"), nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	case "title2":
		p, pf := f.strRef("peer", nil)
		f.rw(depth, 6, []string{p}, f.lit(r.Title2)+","+pf, []string{nm.get("peer")})
	case "colhdr2":
		f.rw(depth, 6, nil, f.lit(r.ColHdr2), nil)
	case "rowben":
		bn, bf := f.strRef("bname", a)
		k := K("k")
		f.rw(depth, 6, []string{bn, fmt.Sprintf("(%s(%s,%s),%s=1,%s)", K("pctl"), k, f.ex(a), k, f.ex(V(nm.get("nladr"))))},
			"1X,"+bf+",8I9", []string{nm.get("bname"), nm.get("pctl"), nm.get("k")})
	case "trailer":
		f.rw(depth, 6, []string{K("nmet"), K("nben"), K("chksum")},
			f.lit("0RECORDS PROCESSED")+",I6,"+f.lit(" METRICS")+",I6,"+f.lit(" BENCHMARKS   SEAL ")+",I6",
			[]string{nm.get("nmet"), nm.get("nben"), nm.get("chksum")})
	case "errmsg":
		code := a.Val
		f.rw(depth, 6, nil, f.lit(fmt.Sprintf(" *** %s%03dE %s ***", j.id.Program[:3], code*10+j.x.intn(10), errText[code])), nil)
	case "debugmsg":
		f.rw(depth, 6, nil, f.lit(fmt.Sprintf(" DEBUG CHECKPOINT %04d", a.Val)), nil)
	case "debugmet":
		f.rw(depth, 6, []string{f.ex(a), f.ex(Ix(nm.get("bsln"), a)), f.ex(Ix(nm.get("curr"), a)), f.ex(Ix(nm.get("delt"), a))},
			f.lit(" DBG MET")+",I4,3I10", []string{nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	case "debugben":
		k := K("k")
		f.rw(depth, 6, []string{f.ex(a), f.ex(Ix(nm.get("medn"), a)), fmt.Sprintf("(%s(%s,%s),%s=1,%s)", K("pctl"), k, f.ex(a), k, f.ex(V(nm.get("nladr"))))},
			f.lit(" DBG BEN")+",I4,I10,8I9", []string{nm.get("medn"), nm.get("pctl")})
	}
}

// ── declarations ────────────────────────────────────────────────────────────

func (f *fortranEmitter) typeName(s Sym) string {
	switch s.Type {
	case "dbl":
		return "DOUBLE PRECISION"
	case "str":
		if f.dialect == "f4" {
			return "INTEGER"
		}
		return "CHARACTER*" + strconv.Itoa(s.Len)
	}
	return "INTEGER"
}

func (f *fortranEmitter) dimText(s Sym) string {
	var dims []string
	if s.Type == "str" && f.dialect == "f4" {
		dims = append(dims, strconv.Itoa(f.words(s.Len)))
	}
	for _, d := range s.Dims {
		ds := strconv.Itoa(d)
		if f.dialect != "f4" {
			for _, g := range f.j.p.Globals {
				if g.Role == "const" && g.Init[0] == ds && g.Name != f.j.nm.get("nladr") {
					ds = f.kw(g.Name)
					break
				}
			}
			if d == 8 {
				ds = f.kw(f.j.nm.get("nladr"))
			}
		}
		dims = append(dims, ds)
	}
	if len(dims) == 0 {
		return f.kw(s.Name)
	}
	return f.kw(s.Name) + "(" + strings.Join(dims, ",") + ")"
}

func (f *fortranEmitter) declare(syms []Sym, depth int) {
	groups := map[string][]Sym{}
	var order []string
	for _, s := range syms {
		if s.Role == "const" {
			continue
		}
		t := f.typeName(s)
		if _, ok := groups[t]; !ok {
			order = append(order, t)
		}
		groups[t] = append(groups[t], s)
	}
	for _, t := range order {
		ss := groups[t]
		for i := 0; i < len(ss); i += f.declPerLine {
			end := minInt(len(ss), i+f.declPerLine)
			var parts []string
			var defs []string
			for _, s := range ss[i:end] {
				parts = append(parts, f.dimText(s))
				defs = append(defs, s.Name)
			}
			f.stmt(0, depth, f.kw(t)+" "+strings.Join(parts, ", "), nil)
			f.out[len(f.out)-1].Defs = defs
		}
	}
}

func (f *fortranEmitter) blockOf(s Sym) string {
	b := s.Common
	if b == "" {
		b = "LADTAB"
	}
	if s.Type == "str" && f.dialect != "f4" {
		if cb, ok := f.strBlocks[b]; ok {
			return cb
		}
		cb := trunc("C"+b, 6)
		f.strBlocks[b] = cb
		return cb
	}
	return b
}

func (f *fortranEmitter) commons(depth int) {
	blocks := map[string][]string{}
	var order []string
	for _, s := range f.j.p.Globals {
		if s.Role == "const" {
			continue
		}
		b := f.blockOf(s)
		if _, ok := blocks[b]; !ok {
			order = append(order, b)
		}
		blocks[b] = append(blocks[b], f.kw(s.Name))
	}
	for _, b := range order {
		f.stmt(0, depth, f.kw("COMMON")+" /"+f.kw(b)+"/ "+strings.Join(blocks[b], ", "), blocks[b])
	}
}

func (f *fortranEmitter) params(depth int) {
	if f.dialect == "f4" {
		return
	}
	var names, parts []string
	for _, s := range f.j.p.Globals {
		if s.Role == "const" {
			names = append(names, f.kw(s.Name))
			parts = append(parts, f.kw(s.Name)+" = "+s.Init[0])
		}
	}
	f.stmt(0, depth, f.kw("INTEGER")+" "+strings.Join(names, ", "), nil)
	f.out[len(f.out)-1].Defs = names
	f.stmt(0, depth, f.kw("PARAMETER")+" ("+strings.Join(parts, ", ")+")", nil)
}

func (f *fortranEmitter) dataStmt(s Sym, depth int) {
	if len(s.Init) == 0 {
		return
	}
	var vals []string
	for _, v := range s.Init {
		switch s.Type {
		case "dbl":
			vals = append(vals, f.ex(DL(v)))
		case "str":
			if f.dialect == "f4" {
				for w := 0; w < len(v); w += 4 {
					vals = append(vals, "4H"+padRight(v[w:minInt(len(v), w+4)], 4))
				}
			} else {
				vals = append(vals, "'"+v+"'")
			}
		default:
			vals = append(vals, v)
		}
	}
	f.stmt(0, depth, f.kw("DATA")+" "+f.kw(s.Name)+" /"+strings.Join(vals, ", ")+"/", []string{s.Name})
}

// ── program units ───────────────────────────────────────────────────────────

func (f *fortranEmitter) unit(p *Proc, callers, callees []string, isMain bool) {
	j := f.j
	f.fmtLabels = map[string]int{}
	f.procFmts = nil
	f.out = append(f.out, Line{Text: f.cchar})
	f.commentLines(j.prologue(p, callers, callees))
	if isMain && f.dialect == "f4" {
		f.comment("Main program. FORTRAN IV has no PROGRAM statement; this card marks the spot.")
	}
	switch {
	case isMain && f.dialect != "f4":
		f.stmt(0, 0, f.kw("PROGRAM ")+f.kw(p.Name), nil)
	case !isMain:
		t := f.kw("SUBROUTINE ") + f.kw(p.Name)
		if len(p.Params) > 0 {
			var ps []string
			for _, q := range p.Params {
				ps = append(ps, f.kw(q))
			}
			t += "(" + strings.Join(ps, ", ") + ")"
		}
		f.stmt(0, 0, t, nil)
	}
	if f.dialect == "gf" || (f.dialect == "f77" && j.x.chance(0.4)) {
		f.stmt(0, 0, f.kw("IMPLICIT NONE"), nil)
	}
	usesGlobals := p.Core || j.x.chance(0.4)
	if usesGlobals {
		f.params(0)
		f.declare(j.p.Globals, 0)
		f.commons(0)
	}
	f.declare(p.Locals, 0)
	for _, s := range p.Locals {
		f.dataStmt(s, 0)
	}
	if !isMain && usesGlobals && j.x.chance(0.1) {
		f.comment(j.g.comment(p.Author, "workarounds", "warnings"))
	}
	f.body(p.Body, 1)
	if !isMain {
		f.stmt(0, 1, f.kw("RETURN"), nil)
	}
	if len(f.procFmts) > 0 {
		if j.x.chance(0.6) {
			f.comment(pick(j.x, []string{"Format statements.", "FORMATS", "Output formats. Widths are load-bearing.", "Formats - see memo " + j.g.memo() + "."}))
		}
		f.out = append(f.out, f.procFmts...)
	}
	f.stmt(0, 0, f.kw("END"), nil)
}

func (f *fortranEmitter) blockData() {
	j := f.j
	var syms []Sym
	for _, s := range j.p.Globals {
		if s.Role != "const" && len(s.Init) > 0 {
			syms = append(syms, s)
		}
	}
	if len(syms) == 0 {
		return
	}
	f.out = append(f.out, Line{Text: f.cchar})
	f.commentLines(j.cs.box("BLOCK DATA", "Initial values for COMMON. The percentile ladder and the provenance seal live here. "+j.g.comment(nil, "warnings", "magic")))
	f.stmt(0, 0, f.kw("BLOCK DATA ")+f.kw(trunc("BD"+j.id.Program, 6)), nil)
	if f.dialect == "gf" {
		f.stmt(0, 0, f.kw("IMPLICIT NONE"), nil)
	}
	f.params(0)
	f.declare(j.p.Globals, 0)
	f.commons(0)
	for _, s := range syms {
		f.dataStmt(s, 0)
	}
	f.stmt(0, 0, f.kw("END"), nil)
}

func (f *fortranEmitter) emit() []Line {
	j := f.j
	callers, callees := callGraph(j.p)
	// program banner + revision history
	f.commentLines(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	f.commentLines(j.historyLines())
	f.commentLines(j.cs.box("Watermark", "Provenance seal "+strings.ToUpper(j.in.Token)+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos")))
	for n := j.x.between(0, 2); n > 0; n-- {
		f.comment(j.g.comment(nil))
		f.out = append(f.out, Line{Text: f.cchar})
	}
	bdFirst := j.x.chance(0.3)
	if bdFirst {
		f.blockData()
	}
	for i, p := range j.p.Procs {
		f.unit(p, callers[p.Name], callees[p.Name], i == 0)
	}
	if !bdFirst {
		f.blockData()
	}
	f.finish()
	return f.out
}

// finish pads to column 72 and adds deck sequence numbers in 73-80
// (optionally with a hand-resequenced, out-of-order region).
func (f *fortranEmitter) finish() {
	if !f.seq {
		return
	}
	n := len(f.out)
	nums := make([]int, n)
	for i := range nums {
		nums[i] = (i + 1) * f.seqStep
	}
	if f.dropped && n > 60 {
		start := f.j.x.between(20, n-30)
		size := f.j.x.between(6, 14)
		region := nums[start : start+size]
		shuffle(f.j.x, region)
		c := f.j.g.coder()
		note := Line{Text: f.cchar + "     " + cardText(fmt.Sprintf("RESEQUENCED BY HAND AFTER THE DECK WAS DROPPED, %s --%s", f.j.g.date(c, c.Start, c.End), c.Initials))}
		f.out = append(f.out[:start], append([]Line{note}, f.out[start:]...)...)
		nums = append(nums[:start], append([]int{nums[start]}, nums[start:]...)...)
	}
	digits := 8 - len(f.seqPrefix)
	if len(f.out)*f.seqStep >= pow10(digits) {
		step := pow10(digits) / (len(f.out) + 1)
		for i := range nums {
			nums[i] = nums[i] / f.seqStep * maxi(1, step/10*10)
		}
	}
	for i := range f.out {
		t := f.out[i].Text
		if visibleLen(t) > 72 {
			t = t[:72]
		}
		f.out[i].Text = t + strings.Repeat(" ", 72-visibleLen(t)) + f.seqPrefix + fmt.Sprintf("%0*d", digits, nums[i]%pow10(digits))
	}
}

func pow10(n int) int {
	p := 1
	for ; n > 0; n-- {
		p *= 10
	}
	return p
}

// symbols returns the declared symbols in a stable order (for the XREF).
func sortedKeys(m map[string][]int) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
