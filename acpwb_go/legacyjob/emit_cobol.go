package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

type cobolEmitter struct {
	j        *job
	out      []Line
	cob74    bool
	verbose  float64
	seq      bool
	seqStep  int
	paraNum  int
	paraStep int
	deferred [][]Line // out-of-line paragraphs (COBOL-74 loops / IF bodies)
	cur      *[]Line
	params   map[string]string // proc param -> WS field
	procPara map[string]string // proc name -> paragraph name
	nlit     int
	evaluate bool // COBOL-85 EVALUATE for runs of equality tests
}

func newCobolEmitter(j *job) *cobolEmitter {
	x := j.x
	c := &cobolEmitter{j: j, params: map[string]string{}, procPara: map[string]string{}}
	c.cob74 = j.in.Year < 1988 || x.chance(0.25)
	c.verbose = []float64{0.2, 0.5, 0.9}[x.intn(3)]
	c.seq = x.chance(0.8)
	c.seqStep = pick(x, []int{10, 100, 10})
	c.paraStep = pick(x, []int{100, 1000, 50})
	c.paraNum = c.paraStep
	c.evaluate = x.chance(0.6)
	return c
}

func (c *cobolEmitter) dst() *[]Line {
	if c.cur != nil {
		return c.cur
	}
	return &c.out
}

// areaA / areaB lines; long statements wrap onto further Area B lines.
func (c *cobolEmitter) a(text string) {
	*c.dst() = append(*c.dst(), Line{Text: "       " + text, Code: true})
}

func (c *cobolEmitter) b(depth int, text string, refs []string) {
	ind := "           " + strings.Repeat("   ", depth)
	for _, l := range wrapWords(text, 72-len(ind)) {
		*c.dst() = append(*c.dst(), Line{Text: ind + l, Code: true, Refs: refs})
		refs = nil
		ind = "           " + strings.Repeat("   ", depth+1)
	}
}

func (c *cobolEmitter) comment(text string) {
	for _, l := range c.j.cs.para(text) {
		*c.dst() = append(*c.dst(), Line{Text: "      *" + l})
	}
}

func (c *cobolEmitter) commentLines(ls []string) {
	for _, l := range ls {
		*c.dst() = append(*c.dst(), Line{Text: "      *" + l})
	}
}

func (c *cobolEmitter) nextPara(stem string) string {
	n := fmt.Sprintf("%04d-%s", c.paraNum, stem)
	c.paraNum += c.paraStep
	return n
}

// ── expressions ─────────────────────────────────────────────────────────────

func (c *cobolEmitter) ex(e *Expr) string {
	switch e.K {
	case "var":
		if p, ok := c.params[e.Name]; ok {
			return p
		}
		return e.Name
	case "idx":
		var parts []string
		for _, a := range e.Args {
			parts = append(parts, c.ex(a))
		}
		return e.Name + " (" + strings.Join(parts, ", ") + ")"
	case "int":
		return strconv.Itoa(e.Val)
	case "dlit":
		return e.S
	case "bin":
		l, r := e.Args[0], e.Args[1]
		ls, rs := c.ex(l), c.ex(r)
		if l.K == "bin" && prec(l.Op) < prec(e.Op) {
			ls = "(" + ls + ")"
		}
		if r.K == "bin" && (prec(r.Op) <= prec(e.Op)) {
			rs = "(" + rs + ")"
		}
		return ls + " " + e.Op + " " + rs
	case "call":
		switch e.Name {
		case "mod":
			return "FUNCTION MOD (" + c.ex(e.Args[0]) + ", " + c.ex(e.Args[1]) + ")"
		case "dble":
			return c.ex(e.Args[0])
		}
	case "ladder":
		return c.ex(e.Args[0]) + " * " + c.ex(e.Args[1])
	}
	return "ZERO"
}

var cobRelShort = map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "NOT =", "gt": ">", "ge": ">="}
var cobRelLong = map[string]string{"lt": "IS LESS THAN", "le": "IS NOT GREATER THAN", "eq": "IS EQUAL TO", "ne": "IS NOT EQUAL TO", "gt": "IS GREATER THAN", "ge": "IS NOT LESS THAN"}

func (c *cobolEmitter) cond(cd *Cond) string {
	switch cd.Op {
	case "and":
		return c.cond(cd.A) + " AND " + c.cond(cd.B)
	case "or":
		return c.cond(cd.A) + " OR " + c.cond(cd.B)
	}
	rel := cobRelShort[cd.Op]
	if c.j.x.float() < c.verbose || c.cob74 {
		rel = cobRelLong[cd.Op]
	}
	return c.ex(cd.L) + " " + rel + " " + c.ex(cd.R)
}

// ── statements ──────────────────────────────────────────────────────────────

func (c *cobolEmitter) body(ss []*Stmt, depth int) {
	for i := 0; i < len(ss); i++ {
		if n := c.evaluateRun(ss[i:], depth); n > 0 {
			i += n - 1
			continue
		}
		c.one(ss[i], depth)
	}
}

// evaluateRun renders a run of "IF v = constant" tests on one variable (no
// ELSE) as a single EVALUATE (COBOL-85). It returns how many statements it
// consumed, 0 if the run is too short or the page prefers IFs.
func (c *cobolEmitter) evaluateRun(ss []*Stmt, depth int) int {
	if c.cob74 || !c.evaluate {
		return 0
	}
	isCase := func(s *Stmt) bool {
		return s.K == "if" && s.Else == nil && s.Cond.Op == "eq" && s.Cond.L.K == "var" && s.Cond.R.K == "int"
	}
	if len(ss) < 3 || !isCase(ss[0]) {
		return 0
	}
	v := ss[0].Cond.L.Name
	n := 0
	seen := map[int]bool{}
	for _, s := range ss {
		if !isCase(s) || s.Cond.L.Name != v || seen[s.Cond.R.Val] {
			break
		}
		seen[s.Cond.R.Val] = true
		n++
	}
	if n < 3 {
		return 0
	}
	c.b(depth, "EVALUATE "+c.ex(V(v)), []string{v})
	for _, s := range ss[:n] {
		c.b(depth+1, "WHEN "+c.ex(s.Cond.R), nil)
		c.body(s.Then, depth+2)
		if len(s.Then) == 0 {
			c.b(depth+2, "CONTINUE", nil)
		}
	}
	c.b(depth+1, "WHEN OTHER", nil)
	c.b(depth+2, "CONTINUE", nil)
	c.b(depth, "END-EVALUATE", nil)
	return n
}

func (c *cobolEmitter) period() string {
	if c.cob74 {
		return "."
	}
	return ""
}

// outOfLine moves a statement list into its own paragraph (COBOL-74 style)
// and returns the paragraph name.
func (c *cobolEmitter) outOfLine(stem string, ss []*Stmt) string {
	name := c.nextPara(stem)
	var lines []Line
	save := c.cur
	c.cur = &lines
	c.a(name + ".")
	c.body(ss, 0)
	if len(ss) == 0 {
		c.b(0, "CONTINUE.", nil)
	}
	c.cur = save
	c.deferred = append(c.deferred, lines)
	return name
}

func (c *cobolEmitter) one(s *Stmt, depth int) {
	x := c.j.x
	switch s.K {
	case "comment":
		if s.Banner {
			c.commentLines(c.j.cs.box(s.Title, s.Text))
		} else {
			c.comment(s.Text)
		}
	case "blank":
		*c.dst() = append(*c.dst(), Line{Text: "      *"})
	case "assign":
		refs := refsOf(s.L, s.R)
		if s.R.K == "var" || s.R.K == "int" || s.R.K == "idx" {
			v := c.ex(s.R)
			if s.R.K == "int" && s.R.Val == 0 && x.chance(0.6) {
				v = pick(x, []string{"ZERO", "ZEROS", "ZEROES"})
			}
			c.b(depth, "MOVE "+v+" TO "+c.ex(s.L)+c.period(), refs)
		} else if s.R.K == "bin" && s.R.Op == "+" && s.R.Args[1].K == "int" && s.R.Args[1].Val == 1 && s.R.Args[0].K == "var" && s.R.Args[0].Name == s.L.Name {
			c.b(depth, "ADD 1 TO "+c.ex(s.L)+c.period(), refs)
		} else {
			c.b(depth, "COMPUTE "+c.ex(s.L)+" = "+c.ex(s.R)+c.period(), refs)
		}
	case "call":
		para := c.procPara[s.Name]
		if para == "" {
			para = s.Name
		}
		if p := c.findProc(s.Name); p != nil {
			for i, a := range s.Args {
				if i < len(p.Params) {
					c.b(depth, "MOVE "+c.ex(a)+" TO "+c.paramField(p, p.Params[i])+c.period(), refsOf(a))
				}
			}
		}
		c.b(depth, "PERFORM "+para+c.period(), []string{s.Name})
	case "stop":
		c.b(depth, pick(x, []string{"STOP RUN.", "STOP RUN.", "GOBACK."}), nil)
	case "io":
		c.io(s, depth)
	case "do":
		from, to := c.ex(s.From), c.ex(s.To)
		v := c.ex(V(s.Var))
		by, until := "BY 1", "UNTIL "+v+" > "+to
		if s.Step < 0 {
			by, until = "BY -1", "UNTIL "+v+" < "+to
		}
		refs := append([]string{s.Var}, refsOf(s.From, s.To)...)
		if c.cob74 {
			para := c.outOfLine("LOOP", s.Body)
			c.b(depth, fmt.Sprintf("PERFORM %s VARYING %s FROM %s %s %s.", para, v, from, by, until), refs)
			return
		}
		c.b(depth, fmt.Sprintf("PERFORM VARYING %s FROM %s %s %s", v, from, by, until), refs)
		c.body(s.Body, depth+1)
		c.b(depth, "END-PERFORM", nil)
	case "if":
		refs := condRefs(s.Cond)
		if c.cob74 {
			then := c.outOfLine("THEN", s.Then)
			if s.Else != nil {
				els := c.outOfLine("ELSE", s.Else)
				c.b(depth, "IF "+c.cond(s.Cond)+" PERFORM "+then+" ELSE PERFORM "+els+".", refs)
			} else {
				c.b(depth, "IF "+c.cond(s.Cond)+" PERFORM "+then+".", refs)
			}
			return
		}
		c.b(depth, "IF "+c.cond(s.Cond), refs)
		if x.chance(0.3) {
			c.b(depth+1, "THEN", nil)
		}
		c.body(s.Then, depth+1)
		if s.Else != nil {
			c.b(depth, "ELSE", nil)
			c.body(s.Else, depth+1)
		}
		c.b(depth, "END-IF", nil)
	}
}

func (c *cobolEmitter) findProc(name string) *Proc {
	for _, p := range c.j.p.Procs {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (c *cobolEmitter) paramField(p *Proc, param string) string {
	key := p.Name + "|" + param
	if f, ok := c.params[key]; ok {
		return f
	}
	f := "LK-" + strings.TrimPrefix(strings.TrimPrefix(p.Name, "WS-"), "W-")
	f = trunc(strings.Split(f, "-PARA")[0], 18) + "-" + strconv.Itoa(len(c.params)+1)
	c.params[key] = f
	return f
}

// ── I/O ─────────────────────────────────────────────────────────────────────

func (c *cobolEmitter) io(s *Stmt, depth int) {
	j := c.j
	nm := j.nm
	var a *Expr
	if len(s.Args) > 0 {
		a = s.Args[0]
	}
	I := func(role string) string { return c.ex(Ix(nm.get(role), a)) }
	read := func() {
		c.b(depth, "READ CARD-FILE INTO WS-CARD", nil)
		c.b(depth+1, "AT END MOVE 'Y' TO WS-EOF-SW", nil)
		if !c.cob74 {
			c.b(depth, "END-READ", nil)
		}
	}
	adv := func(rec string, n string) {
		c.b(depth, "WRITE PRINT-REC FROM "+rec, nil)
		c.b(depth+1, "AFTER ADVANCING "+n+c.period(), nil)
	}
	switch s.Name {
	case "readhdr":
		read()
		c.b(depth, "MOVE CRD-ORG TO "+nm.get("org")+c.period(), []string{nm.get("org")})
		c.b(depth, "MOVE CRD-RECID TO "+nm.get("recid")+c.period(), []string{nm.get("recid")})
		read()
		c.b(depth, "MOVE WS-CARD TO "+nm.get("peer")+c.period(), []string{nm.get("peer")})
		read()
		c.b(depth, "MOVE CRD-NMET TO "+nm.get("nmet")+c.period(), []string{nm.get("nmet")})
		c.b(depth, "MOVE CRD-NBEN TO "+nm.get("nben")+c.period(), []string{nm.get("nben")})
	case "readmet":
		read()
		c.b(depth, "MOVE CRD-MNAME TO "+I("mname")+c.period(), []string{nm.get("mname")})
		c.b(depth, "MOVE CRD-BASE TO "+I("bsln")+c.period(), []string{nm.get("bsln")})
		c.b(depth, "MOVE CRD-CURR TO "+I("curr")+c.period(), []string{nm.get("curr")})
	case "readben":
		read()
		c.b(depth, "MOVE CRD-BNAME TO "+I("bname")+c.period(), []string{nm.get("bname")})
		c.b(depth, "MOVE CRD-MED TO "+I("medn")+c.period(), []string{nm.get("medn")})
	case "pagehdr":
		c.b(depth, "MOVE "+nm.get("recid")+" TO PH-RECID"+c.period(), nil)
		c.b(depth, "MOVE "+nm.get("ipage")+" TO PH-PAGE"+c.period(), nil)
		adv("PRT-PAGE-HDR", "PAGE")
	case "title1":
		c.b(depth, "MOVE "+nm.get("org")+" TO T1-ORG"+c.period(), nil)
		adv("PRT-TITLE-1", "2 LINES")
	case "colhdr1":
		adv("PRT-COLHDR-1", "1 LINE")
	case "rowmet":
		c.b(depth, "MOVE "+I("mname")+" TO D1-NAME"+c.period(), nil)
		c.b(depth, "MOVE "+I("bsln")+" TO D1-BASE"+c.period(), nil)
		c.b(depth, "MOVE "+I("curr")+" TO D1-CURR"+c.period(), nil)
		c.b(depth, "MOVE "+I("delt")+" TO D1-CHG"+c.period(), nil)
		adv("PRT-DETAIL-1", "1 LINE")
	case "title2":
		c.b(depth, "MOVE "+nm.get("peer")+" TO T2-PEER"+c.period(), nil)
		adv("PRT-TITLE-2", "2 LINES")
	case "colhdr2":
		adv("PRT-COLHDR-2", "1 LINE")
	case "rowben":
		c.b(depth, "MOVE "+I("bname")+" TO D2-NAME"+c.period(), nil)
		for k := 1; k <= 8; k++ {
			c.b(depth, fmt.Sprintf("MOVE %s TO D2-P (%d)%s", c.ex(Ix(nm.get("pctl"), N(k), a)), k, c.period()), nil)
		}
		adv("PRT-DETAIL-2", "1 LINE")
	case "trailer":
		c.b(depth, "MOVE "+nm.get("nmet")+" TO TR-NMET"+c.period(), nil)
		c.b(depth, "MOVE "+nm.get("nben")+" TO TR-NBEN"+c.period(), nil)
		c.b(depth, "MOVE "+nm.get("chksum")+" TO TR-SEAL"+c.period(), nil)
		adv("PRT-TRAILER", "2 LINES")
	case "errmsg":
		c.b(depth, fmt.Sprintf("DISPLAY '*** %s%03dE %s ***'%s", j.id.Program[:3], a.Val*10+j.x.intn(10), errText[a.Val], c.period()), nil)
	case "debugmsg":
		c.b(depth, fmt.Sprintf("DISPLAY 'DEBUG CHECKPOINT %04d'%s", a.Val, c.period()), nil)
	case "debugmet":
		c.b(depth, "DISPLAY 'DBG MET ' "+c.ex(a)+" ' ' "+I("bsln")+" ' ' "+I("curr")+c.period(), nil)
	case "debugben":
		c.b(depth, "DISPLAY 'DBG BEN ' "+c.ex(a)+" ' ' "+I("medn")+c.period(), nil)
	}
}

// ── data division ───────────────────────────────────────────────────────────

func (c *cobolEmitter) pic(s Sym) string {
	switch s.Type {
	case "dbl":
		if c.cob74 {
			return "PIC 9V99"
		}
		return pick(c.j.x, []string{"COMP-2", "PIC 9V99 COMP-3", "PIC 9V99"})
	case "str":
		return fmt.Sprintf("PIC X(%d)", s.Len)
	}
	return pick(c.j.x, []string{"PIC S9(9) COMP", "PIC S9(9) COMP-3", "PIC S9(8) COMP", "PIC 9(9)"})
}

func (c *cobolEmitter) field(level int, name, pic string, extra string) {
	t := fmt.Sprintf("%02d  %-28s %s%s.", level, name, pic, extra)
	if level == 1 || level == 77 {
		c.a(t)
		return
	}
	c.b(strings.Count(fmt.Sprint(level), "")/3, t, nil)
}

// fillers splits a literal into FILLER VALUE items of at most 30 chars.
func (c *cobolEmitter) fillers(level int, s string) {
	for len(s) > 0 {
		n := minInt(len(s), 30)
		part := s[:n]
		s = s[n:]
		c.b(1, fmt.Sprintf("%02d  FILLER %s VALUE '%s'.", level, fmt.Sprintf("PIC X(%d)", len(part)), strings.ReplaceAll(part, "'", "''")), nil)
	}
}

func (c *cobolEmitter) dataDivision() {
	j := c.j
	r := j.rep
	c.a("DATA DIVISION.")
	c.a("FILE SECTION.")
	c.a("FD  CARD-FILE")
	c.b(0, "LABEL RECORDS ARE OMITTED", nil)
	c.b(0, "RECORD CONTAINS 80 CHARACTERS.", nil)
	c.a("01  CARD-REC                     PIC X(80).")
	c.a("FD  PRINT-FILE")
	c.b(0, "LABEL RECORDS ARE OMITTED", nil)
	c.b(0, "RECORD CONTAINS 133 CHARACTERS.", nil)
	c.a("01  PRINT-REC                    PIC X(133).")
	c.a("WORKING-STORAGE SECTION.")
	c.comment(j.g.comment(nil, "magic", "workarounds", "warnings"))
	c.a("77  WS-EOF-SW                    PIC X VALUE 'N'.")
	for _, s := range j.p.Globals {
		if s.Role == "const" {
			c.a(fmt.Sprintf("77  %-28s PIC 9(4) COMP VALUE %s.", s.Name, s.Init[0]))
		}
	}
	for _, s := range j.p.Globals {
		if s.Role == "const" || len(s.Dims) > 0 {
			continue
		}
		val := ""
		if len(s.Init) > 0 {
			if s.Type == "str" {
				val = " VALUE '" + s.Init[0] + "'"
			} else {
				val = " VALUE " + s.Init[0]
			}
		} else if s.Type != "str" {
			val = " VALUE ZERO"
		}
		c.a(fmt.Sprintf("77  %-28s %s%s.", s.Name, c.pic(s), val))
	}
	c.a("01  WS-TABLES.")
	for _, s := range j.p.Globals {
		if s.Role == "const" || len(s.Dims) == 0 {
			continue
		}
		if len(s.Init) > 0 {
			continue
		}
		if len(s.Dims) == 2 {
			c.b(0, fmt.Sprintf("05  %s-ROW OCCURS %d TIMES.", s.Name, s.Dims[0]), nil)
			c.b(1, fmt.Sprintf("10  %s OCCURS %d TIMES %s.", s.Name, s.Dims[1], c.pic(s)), nil)
			continue
		}
		c.b(0, fmt.Sprintf("05  %s OCCURS %d TIMES %s.", s.Name, s.Dims[0], c.pic(s)), nil)
	}
	for _, s := range j.p.Globals {
		if s.Role == "const" || len(s.Dims) == 0 || len(s.Init) == 0 {
			continue
		}
		vname := s.Name + "-VALUES"
		c.a("01  " + vname + ".")
		for _, v := range s.Init {
			if s.Type == "dbl" {
				c.b(0, fmt.Sprintf("05  FILLER PIC 9V99 VALUE %s.", v), nil)
			} else {
				c.b(0, fmt.Sprintf("05  FILLER PIC 9(4) VALUE %s.", v), nil)
			}
		}
		c.a(fmt.Sprintf("01  FILLER REDEFINES %s.", vname))
		pic := "PIC 9V99"
		if s.Type != "dbl" {
			pic = "PIC 9(4)"
		}
		c.b(0, fmt.Sprintf("05  %s OCCURS %d TIMES %s.", s.Name, len(s.Init), pic), nil)
	}
	// dead-routine locals and parameter fields
	for _, p := range j.p.Procs {
		if len(p.Locals) == 0 {
			continue
		}
		c.comment("Work fields for " + p.Name + ". " + pick(j.x, []string{"", "Retained.", "Do not reuse.", "Shared with nothing."}))
		for _, l := range p.Locals {
			name := l.Name
			if l.Role == "param" {
				name = c.paramField(p, l.Name)
				c.params[l.Name] = name
			}
			if len(l.Dims) > 0 {
				c.a(fmt.Sprintf("01  %s-TBL.", name))
				c.b(0, fmt.Sprintf("05  %s OCCURS %d TIMES %s.", name, l.Dims[0], c.pic(l)), nil)
				continue
			}
			c.a(fmt.Sprintf("77  %-28s %s.", name, c.pic(l)))
		}
	}
	// card layouts
	c.a("01  WS-CARD                      PIC X(80).")
	c.a("01  CRD-HDR REDEFINES WS-CARD.")
	c.b(0, "05  CRD-ORG                  PIC X(40).", nil)
	c.b(0, "05  CRD-RECID                PIC X(8).", nil)
	c.b(0, "05  FILLER                   PIC X(32).", nil)
	c.a("01  CRD-CTL REDEFINES WS-CARD.")
	c.b(0, "05  CRD-NMET                 PIC 9(3).", nil)
	c.b(0, "05  CRD-NBEN                 PIC 9(3).", nil)
	c.b(0, "05  FILLER                   PIC X(74).", nil)
	c.a("01  CRD-MET REDEFINES WS-CARD.")
	c.b(0, "05  CRD-MNAME                PIC X(40).", nil)
	c.b(0, "05  CRD-BASE                 PIC 9(6).", nil)
	c.b(0, "05  CRD-CURR                 PIC 9(6).", nil)
	c.b(0, "05  FILLER                   PIC X(28).", nil)
	c.a("01  CRD-BEN REDEFINES WS-CARD.")
	c.b(0, "05  CRD-BNAME                PIC X(40).", nil)
	c.b(0, "05  CRD-MED                  PIC 9(7).", nil)
	c.b(0, "05  FILLER                   PIC X(33).", nil)
	// print lines
	c.comment("Print line layouts. " + j.g.comment(nil, "units", "memos"))
	c.a("01  PRT-PAGE-HDR.")
	c.fillers(5, r.HdrLeft[1:])
	c.b(1, "05  PH-RECID                 PIC X(8).", nil)
	c.fillers(5, r.HdrMid)
	c.b(1, "05  PH-PAGE                  PIC ZZZZ9.", nil)
	c.a("01  PRT-TITLE-1.")
	c.fillers(5, r.Title1[1:])
	c.b(1, "05  T1-ORG                   PIC X(40).", nil)
	c.a("01  PRT-COLHDR-1.")
	c.fillers(5, r.ColHdr1[1:])
	c.a("01  PRT-DETAIL-1.")
	c.b(1, "05  D1-NAME                  PIC X(40).", nil)
	c.b(1, "05  D1-BASE                  PIC Z(8)9.", nil)
	c.b(1, "05  D1-CURR                  PIC Z(8)9.", nil)
	c.b(1, "05  D1-CHG                   PIC +(8)9.", nil)
	c.a("01  PRT-TITLE-2.")
	c.fillers(5, r.Title2[1:])
	c.b(1, "05  T2-PEER                  PIC X(60).", nil)
	c.a("01  PRT-COLHDR-2.")
	c.fillers(5, r.ColHdr2[1:])
	c.a("01  PRT-DETAIL-2.")
	c.b(1, "05  D2-NAME                  PIC X(40).", nil)
	c.b(1, "05  D2-P OCCURS 8 TIMES      PIC Z(8)9.", nil)
	c.a("01  PRT-TRAILER.")
	c.fillers(5, "RECORDS PROCESSED")
	c.b(1, "05  TR-NMET                  PIC ZZZZZ9.", nil)
	c.fillers(5, " METRICS")
	c.b(1, "05  TR-NBEN                  PIC ZZZZZ9.", nil)
	c.fillers(5, " BENCHMARKS   SEAL ")
	c.b(1, "05  TR-SEAL                  PIC ZZZZZ9.", nil)
}

func (c *cobolEmitter) emit() []Line {
	j := c.j
	x := j.x
	callers, callees := callGraph(j.p)
	main := j.p.Procs[0]
	for i, p := range j.p.Procs {
		if i == 0 {
			c.procPara[p.Name] = "0000-MAIN"
			continue
		}
		c.procPara[p.Name] = c.nextPara(strings.Split(strings.TrimSuffix(p.Name, "-"), "-")[0])
	}
	c.commentLines(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	c.commentLines(j.historyLines())
	c.a("IDENTIFICATION DIVISION.")
	c.a("PROGRAM-ID. " + j.id.Program + ".")
	a0 := j.g.roster[0]
	c.a("AUTHOR. " + cardText(a0.Name()) + ".")
	c.a("INSTALLATION. ACPWB RESEARCH DIVISION, MILWAUKEE.")
	c.a("DATE-WRITTEN. " + cardText(j.g.date(a0, a0.Start, a0.Start)) + ".")
	c.a("DATE-COMPILED.")
	c.a("SECURITY. " + cardText(trunc(j.in.Distribution, 50)) + ".")
	if c.cob74 {
		c.a("REMARKS.")
		for _, l := range wrapWords(cardText(j.g.comment(nil, "data", "units", "memos")), 56) {
			c.b(0, l, nil)
		}
	}
	c.a("ENVIRONMENT DIVISION.")
	c.a("CONFIGURATION SECTION.")
	mach := pick(x, []string{"IBM-370", "IBM-4381", "IBM-3090", "IBM-ES9000", "VAX-8600", "IBM-390"})
	c.a("SOURCE-COMPUTER. " + mach + ".")
	c.a("OBJECT-COMPUTER. " + mach + ".")
	if x.chance(0.4) {
		c.a("SPECIAL-NAMES. C01 IS TOP-OF-PAGE.")
	}
	c.a("INPUT-OUTPUT SECTION.")
	c.a("FILE-CONTROL.")
	c.b(0, "SELECT CARD-FILE ASSIGN TO "+pick(x, []string{"UT-S-SYSIN", "SYSIN", "CARDIN"})+".", nil)
	c.b(0, "SELECT PRINT-FILE ASSIGN TO "+pick(x, []string{"UT-S-SYSPRINT", "SYSPRINT", "PRTOUT"})+".", nil)
	c.dataDivision()
	c.a("PROCEDURE DIVISION.")
	c.commentLines(j.prologue(main, nil, callees[main.Name]))
	c.a("0000-MAIN.")
	c.b(0, "OPEN INPUT CARD-FILE OUTPUT PRINT-FILE"+c.period(), nil)
	for _, s := range main.Body {
		if s.K == "stop" {
			c.b(0, "CLOSE CARD-FILE PRINT-FILE"+c.period(), nil)
		}
		c.one(s, 0)
	}
	for _, p := range j.p.Procs[1:] {
		c.commentLines(j.prologue(p, callers[p.Name], callees[p.Name]))
		c.a(c.procPara[p.Name] + ".")
		c.body(p.Body, 0)
		c.b(0, "EXIT.", nil)
	}
	for _, d := range c.deferred {
		c.out = append(c.out, d...)
	}
	c.finish()
	return c.out
}

func (c *cobolEmitter) finish() {
	id := padRight(c.j.id.Program, 8)
	for i := range c.out {
		t := c.out[i].Text
		if len(t) > 72 {
			t = t[:72]
		}
		seq := "      "
		if c.seq {
			seq = fmt.Sprintf("%06d", ((i+1)*c.seqStep)%1000000)
		}
		if len(t) >= 6 {
			t = seq + t[6:]
		}
		c.out[i].Text = padRight(t, 72) + id
	}
}
