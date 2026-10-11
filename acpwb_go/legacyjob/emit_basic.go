package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

// basicEmitter renders GW-BASIC / VAX BASIC style line-numbered source:
// procedures become GOSUB blocks, arguments are passed through globals,
// tables are loaded with READ/DATA.
type basicEmitter struct {
	j        *job
	out      []Line
	lineNo   int
	step     int
	apos     bool // ' comments instead of REM
	multi    bool // pack short statements with ':'
	sub      map[string]int
	pending  []string // deferred DATA lines
	paramMap map[string]string
	atari    bool // Atari 8-bit BASIC: REM only, A$(i,j) substrings, no PRINT USING
	strArr   map[string]int
}

func newBasicEmitter(j *job) *basicEmitter {
	x := j.x
	b := &basicEmitter{j: j, step: pick(x, []int{10, 10, 5, 20}), apos: x.chance(0.3), multi: x.chance(0.5),
		sub: map[string]int{}, paramMap: map[string]string{}, strArr: map[string]int{}}
	if j.id.LangKey == "atari" {
		b.atari, b.apos = true, false
		for _, s := range j.p.Globals {
			if s.Type == "str" && len(s.Dims) > 0 {
				b.strArr[s.Name] = s.Len
			}
		}
	}
	return b
}

func (b *basicEmitter) emitLine(text string, code bool, refs []string) {
	b.lineNo += b.step
	b.out = append(b.out, Line{Text: fmt.Sprintf("%d %s", b.lineNo, text), Code: code, Refs: refs, Label: strconv.Itoa(b.lineNo)})
}

func (b *basicEmitter) rem(text string) {
	kw := "REM "
	if b.apos {
		kw = "' "
	}
	for _, l := range b.j.cs.para(text) {
		b.emitLine(kw+l, false, nil)
	}
}

func (b *basicEmitter) remLines(ls []string) {
	kw := "REM "
	if b.apos {
		kw = "' "
	}
	for _, l := range ls {
		b.emitLine(strings.TrimRight(kw+l, " "), false, nil)
	}
}

func (b *basicEmitter) name(n string) string {
	if p, ok := b.paramMap[n]; ok {
		return p
	}
	return n
}

func (b *basicEmitter) ex(e *Expr) string {
	switch e.K {
	case "var":
		return b.name(e.Name)
	case "idx":
		var parts []string
		for _, a := range e.Args {
			parts = append(parts, b.ex(a))
		}
		if w, ok := b.strArr[e.Name]; ok && len(parts) == 1 {
			// Atari BASIC has no string arrays: one long string, sliced.
			i := wrapIfBin(e.Args[0], parts[0])
			return fmt.Sprintf("%s(%s*%d-%d,%s*%d)", b.name(e.Name), i, w, w-1, i, w)
		}
		return b.name(e.Name) + "(" + strings.Join(parts, ",") + ")"
	case "int":
		return strconv.Itoa(e.Val)
	case "dlit":
		return strings.TrimPrefix(e.S, "0")
	case "bin":
		l, r := e.Args[0], e.Args[1]
		ls, rs := b.ex(l), b.ex(r)
		if l.K == "bin" && prec(l.Op) < prec(e.Op) {
			ls = "(" + ls + ")"
		}
		if r.K == "bin" && prec(r.Op) <= prec(e.Op) {
			rs = "(" + rs + ")"
		}
		if e.Op == "/" && !isDouble(l) && !isDouble(r) {
			if b.j.x.chance(0.5) && !b.atari {
				return ls + " \\ " + rs
			}
			return "INT(" + ls + " / " + rs + ")"
		}
		return ls + " " + e.Op + " " + rs
	case "call":
		switch e.Name {
		case "mod":
			a, c := wrapIfBin(e.Args[0], b.ex(e.Args[0])), wrapIfBin(e.Args[1], b.ex(e.Args[1]))
			if b.atari {
				return "(" + a + "-INT(" + a + "/" + c + ")*" + c + ")"
			}
			return a + " MOD " + c
		case "dble":
			if b.atari {
				return b.ex(e.Args[0])
			}
			return "CDBL(" + b.ex(e.Args[0]) + ")"
		}
	case "ladder":
		return "INT(" + b.ex(e.Args[0]) + " * " + wrapIfBin(e.Args[1], b.ex(e.Args[1])) + ")"
	}
	return "0"
}

var basRel = map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "<>", "gt": ">", "ge": ">="}

func (b *basicEmitter) cond(c *Cond) string {
	switch c.Op {
	case "and":
		return b.cond(c.A) + " AND " + b.cond(c.B)
	case "or":
		return b.cond(c.A) + " OR " + b.cond(c.B)
	}
	return b.ex(c.L) + " " + basRel[c.Op] + " " + b.ex(c.R)
}

func (b *basicEmitter) body(ss []*Stmt) {
	for _, s := range ss {
		b.one(s)
	}
}

func (b *basicEmitter) inline(s *Stmt) string {
	switch s.K {
	case "assign":
		return b.assign(s)
	case "call":
		return b.callText(s)
	case "stop":
		return "END"
	}
	return ""
}

func (b *basicEmitter) assign(s *Stmt) string {
	t := b.ex(s.L) + " = " + b.ex(s.R)
	if b.j.x.chance(0.15) {
		t = "LET " + t
	}
	return t
}

func (b *basicEmitter) callText(s *Stmt) string {
	var parts []string
	if p := b.findProc(s.Name); p != nil {
		for i, a := range s.Args {
			if i < len(p.Params) {
				parts = append(parts, b.name(p.Params[i])+" = "+b.ex(a))
			}
		}
	}
	parts = append(parts, "GOSUB @SUB:"+s.Name+"@")
	return strings.Join(parts, " : ")
}

func (b *basicEmitter) findProc(name string) *Proc {
	for _, p := range b.j.p.Procs {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (b *basicEmitter) one(s *Stmt) {
	x := b.j.x
	switch s.K {
	case "comment":
		if s.Banner {
			b.remLines(b.j.cs.box(s.Title, s.Text))
		} else {
			b.rem(s.Text)
		}
	case "blank":
		b.emitLine("REM", false, nil)
	case "assign":
		b.emitLine(b.assign(s), true, refsOf(s.L, s.R))
	case "call":
		b.emitLine(b.callText(s), true, append([]string{s.Name}, refsOf(s.Args...)...))
	case "stop":
		if b.atari {
			b.emitLine(pick(x, []string{"END", "STOP"}), true, nil)
		} else {
			b.emitLine(pick(x, []string{"END", "STOP", "SYSTEM"}), true, nil)
		}
	case "io":
		b.io(s)
	case "do":
		v := b.name(s.Var)
		step := ""
		if s.Step < 0 {
			step = " STEP -1"
		}
		refs := append([]string{s.Var}, refsOf(s.From, s.To)...)
		b.emitLine(fmt.Sprintf("FOR %s = %s TO %s%s", v, b.ex(s.From), b.ex(s.To), step), true, refs)
		b.body(s.Body)
		if x.chance(0.6) || b.atari {
			b.emitLine("NEXT "+v, true, nil)
		} else {
			b.emitLine("NEXT", true, nil)
		}
	case "if":
		refs := condRefs(s.Cond)
		if s.Else == nil {
			if one := simpleStmt(s.Then); one != nil {
				if t := b.inline(one); t != "" {
					b.emitLine("IF "+b.cond(s.Cond)+" THEN "+t, true, refs)
					return
				}
			}
		}
		// IF NOT ... THEN GOTO skip
		b.emitLine("IF "+b.cond(negate(s.Cond))+" THEN GOTO @SKIP@", true, refs)
		gotoIdx := len(b.out) - 1
		b.body(s.Then)
		if s.Else != nil {
			b.emitLine("GOTO @END@", true, nil)
			endIdx := len(b.out) - 1
			b.emitLine("REM ELSE", false, nil)
			b.patch(gotoIdx, "@SKIP@", b.lineNo)
			b.body(s.Else)
			b.emitLine("REM END IF", false, nil)
			b.patch(endIdx, "@END@", b.lineNo)
			return
		}
		b.emitLine("REM END IF", false, nil)
		b.patch(gotoIdx, "@SKIP@", b.lineNo)
	}
}

func (b *basicEmitter) patch(idx int, marker string, line int) {
	b.out[idx].Text = strings.Replace(b.out[idx].Text, marker, strconv.Itoa(line), 1)
}

// using builds a PRINT USING image: "\  \" string fields and # numerics.
func strField(n int) string { return "\\" + strings.Repeat(" ", n-2) + "\\" }

func numField(n int, signed bool) string {
	if signed {
		return "+" + strings.Repeat("#", n-1)
	}
	return strings.Repeat("#", n)
}

func basq(s string) string { return strings.ReplaceAll(s, "\"", "'") }

func (b *basicEmitter) io(s *Stmt) {
	j := b.j
	nm := j.nm
	r := j.rep
	var a *Expr
	if len(s.Args) > 0 {
		a = s.Args[0]
	}
	I := func(role string) string { return b.ex(Ix(nm.get(role), a)) }
	G := func(role string) string { return b.name(nm.get(role)) }
	read := func() { b.emitLine("LINE INPUT #1, CARD$", true, nil) }
	if b.atari {
		b.atariIO(s, a, I, G)
		return
	}
	switch s.Name {
	case "readhdr":
		read()
		b.emitLine(fmt.Sprintf("%s = MID$(CARD$, 1, 40) : %s = MID$(CARD$, 41, 8)", G("org"), G("recid")), true, []string{nm.get("org"), nm.get("recid")})
		read()
		b.emitLine(fmt.Sprintf("%s = CARD$", G("peer")), true, []string{nm.get("peer")})
		read()
		b.emitLine(fmt.Sprintf("%s = VAL(MID$(CARD$, 1, 3)) : %s = VAL(MID$(CARD$, 4, 3))", G("nmet"), G("nben")), true, []string{nm.get("nmet"), nm.get("nben")})
	case "readmet":
		read()
		b.emitLine(fmt.Sprintf("%s = MID$(CARD$, 1, 40)", I("mname")), true, []string{nm.get("mname")})
		b.emitLine(fmt.Sprintf("%s = VAL(MID$(CARD$, 41, 6)) : %s = VAL(MID$(CARD$, 47, 6))", I("bsln"), I("curr")), true, []string{nm.get("bsln"), nm.get("curr")})
	case "readben":
		read()
		b.emitLine(fmt.Sprintf("%s = MID$(CARD$, 1, 40) : %s = VAL(MID$(CARD$, 41, 7))", I("bname"), I("medn")), true, []string{nm.get("bname"), nm.get("medn")})
	case "pagehdr":
		b.emitLine(fmt.Sprintf("PRINT \"%s\"; %s; \"%s\"; USING \"#####\"; %s", basq(r.HdrLeft), G("recid"), basq(r.HdrMid), G("ipage")), true, []string{nm.get("recid"), nm.get("ipage")})
	case "title1":
		b.emitLine(fmt.Sprintf("PRINT \"%s\"; %s", basq(r.Title1), G("org")), true, []string{nm.get("org")})
	case "colhdr1":
		b.emitLine("PRINT \""+basq(r.ColHdr1)+"\"", true, nil)
	case "rowmet":
		img := " " + strField(40) + numField(9, false) + numField(9, false) + numField(9, true)
		b.emitLine(fmt.Sprintf("PRINT USING \"%s\"; %s; %s; %s; %s", img, I("mname"), I("bsln"), I("curr"), I("delt")), true, []string{nm.get("mname"), nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	case "title2":
		b.emitLine(fmt.Sprintf("PRINT \"%s\"; %s", basq(r.Title2), G("peer")), true, []string{nm.get("peer")})
	case "colhdr2":
		b.emitLine("PRINT \""+basq(r.ColHdr2)+"\"", true, nil)
	case "rowben":
		img := " " + strField(40) + strings.Repeat(numField(9, false), 8)
		var vals []string
		for k := 1; k <= 8; k++ {
			vals = append(vals, b.ex(Ix(nm.get("pctl"), N(k), a)))
		}
		b.emitLine(fmt.Sprintf("PRINT USING \"%s\"; %s; %s", img, I("bname"), strings.Join(vals, "; ")), true, []string{nm.get("bname"), nm.get("pctl")})
	case "trailer":
		b.emitLine(fmt.Sprintf("PRINT USING \"0RECORDS PROCESSED###### METRICS###### BENCHMARKS   SEAL ######\"; %s; %s; %s", G("nmet"), G("nben"), G("chksum")), true, []string{nm.get("nmet"), nm.get("nben"), nm.get("chksum")})
	case "errmsg":
		b.emitLine(fmt.Sprintf("PRINT \" *** %s%03dE %s ***\"", j.id.Program[:3], a.Val*10+j.x.intn(10), errText[a.Val]), true, nil)
	case "debugmsg":
		b.emitLine(fmt.Sprintf("PRINT \" DEBUG CHECKPOINT %04d\"", a.Val), true, nil)
	case "debugmet":
		b.emitLine(fmt.Sprintf("PRINT \" DBG MET\"; %s; %s; %s; %s", b.ex(a), I("bsln"), I("curr"), I("delt")), true, []string{nm.get("bsln")})
	case "debugben":
		b.emitLine(fmt.Sprintf("PRINT \" DBG BEN\"; %s; %s", b.ex(a), I("medn")), true, []string{nm.get("medn")})
	}
}

func (b *basicEmitter) emit() []Line {
	j := b.j
	x := j.x
	callers, callees := callGraph(j.p)
	// assign subroutine line numbers up front (1000, 2000, ...)
	base := pick(x, []int{1000, 500, 2000})
	for _, p := range j.p.Procs[1:] {
		for _, l := range p.Locals {
			if l.Role == "param" {
				b.paramMap[l.Name] = l.Name
			}
		}
	}

	b.remLines(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	b.remLines(j.historyLines())
	b.remLines(j.cs.box("Provenance", "Seal "+strings.ToUpper(j.in.Token)+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "memos")))
	if b.atari {
		b.emitLine("DIM CARD$(80),LN$(140)", true, nil)
		for _, s := range j.p.Globals {
			if s.Type == "str" {
				n := s.Len
				if len(s.Dims) > 0 {
					n = s.Len * s.Dims[0]
				}
				b.emitLine(fmt.Sprintf("DIM %s(%d)", s.Name, n), true, []string{s.Name})
			}
		}
		b.emitLine("POKE 82,0:REM LEFT MARGIN 0 SO THE 40-COLUMN SCREEN SHOWS COLUMN 1", true, nil)
	} else {
		b.emitLine("DEFDBL A-Z", true, nil)
		if x.chance(0.5) {
			b.emitLine("DEFINT I-N : REM COUNTERS ONLY. DO NOT PUT MONEY IN AN I-N VARIABLE.", true, nil)
		}
	}
	for _, s := range j.p.Globals {
		if s.Role == "const" {
			b.emitLine(fmt.Sprintf("%s = %s", s.Name, s.Init[0]), true, []string{s.Name})
		}
	}
	var dims []string
	for _, s := range j.p.Globals {
		if len(s.Dims) > 0 && !(b.atari && s.Type == "str") {
			var ds []string
			for _, d := range s.Dims {
				ds = append(ds, strconv.Itoa(d))
			}
			dims = append(dims, s.Name+"("+strings.Join(ds, ",")+")")
		}
	}
	for i := 0; i < len(dims); i += 4 {
		b.emitLine("DIM "+strings.Join(dims[i:minInt(len(dims), i+4)], ", "), true, nil)
	}
	for _, s := range j.p.Globals {
		if len(s.Init) == 0 {
			continue
		}
		if len(s.Dims) == 0 {
			v := s.Init[0]
			if s.Type == "str" {
				v = "\"" + v + "\""
			}
			b.emitLine(s.Name+" = "+v, true, []string{s.Name})
			continue
		}
		k := j.nm.get("k")
		b.emitLine(fmt.Sprintf("RESTORE @DATA:%d@", len(b.pending)), true, nil)
		b.emitLine(fmt.Sprintf("FOR %s = 1 TO %d : READ %s(%s) : NEXT %s", k, len(s.Init), s.Name, k, k), true, []string{s.Name})
		var vals []string
		for _, v := range s.Init {
			vals = append(vals, strings.TrimPrefix(v, "0"))
		}
		b.pending = append(b.pending, "DATA "+strings.Join(vals, ", "))
	}
	if b.atari {
		b.emitLine("OPEN #1,4,0,\"D1:DECK.DAT\"", true, nil)
		b.emitLine("OPEN #2,8,0,\"P:\"", true, nil)
	} else {
		b.emitLine(fmt.Sprintf("OPEN \"%s\" FOR INPUT AS #1", pick(x, []string{"DECK.DAT", "A:DECK.DAT", "SYS$INPUT:"})), true, nil)
	}
	main := j.p.Procs[0]
	b.remLines(j.prologue(main, nil, callees[main.Name]))
	for _, s := range main.Body {
		if s.K == "stop" {
			if b.atari {
				b.emitLine("CLOSE #1:CLOSE #2", true, nil)
			} else {
				b.emitLine("CLOSE #1", true, nil)
			}
		}
		b.one(s)
	}
	for _, p := range j.p.Procs[1:] {
		if x.chance(0.5) {
			b.lineNo = (b.lineNo/base + 1) * base
			b.lineNo -= b.step
		}
		b.sub[p.Name] = b.lineNo + b.step
		b.remLines(j.prologue(p, callers[p.Name], callees[p.Name]))
		b.body(p.Body)
		b.emitLine("RETURN", true, nil)
	}
	if b.lineNo < 9000-b.step {
		b.lineNo = 9000 - b.step
	}
	b.rem("Ladder and seal tables. " + j.g.comment(nil, "magic", "warnings"))
	dataLine := map[string]int{}
	for i, d := range b.pending {
		b.emitLine(d, true, nil)
		dataLine[fmt.Sprintf("@DATA:%d@", i)] = b.lineNo
	}
	for i := range b.out {
		t := b.out[i].Text
		for strings.Contains(t, "@SUB:") {
			a := strings.Index(t, "@SUB:")
			e := strings.Index(t[a+5:], "@") + a + 5
			t = t[:a] + strconv.Itoa(b.sub[t[a+5:e]]) + t[e+1:]
		}
		for k, v := range dataLine {
			t = strings.ReplaceAll(t, k, strconv.Itoa(v))
		}
		b.out[i].Text = t
	}
	return b.out
}

// atariIO: Atari BASIC has INPUT # (no LINE INPUT), A$(i,j) substrings
// instead of MID$, and no PRINT USING — columns come from the printer's
// tab stops, which is why the report only lines up on the 1025.
func (b *basicEmitter) atariIO(s *Stmt, a *Expr, I, G func(string) string) {
	j := b.j
	nm := j.nm
	r := j.rep
	sub := func(start, width int) string { return fmt.Sprintf("CARD$(%d,%d)", start+1, start+width) }
	read := func() { b.emitLine("INPUT #1,CARD$", true, nil) }
	pr := func(parts ...string) { b.emitLine("PRINT #2;"+strings.Join(parts, ";"), true, nil) }
	q := func(t string) string { return "\"" + basq(cardText(t)) + "\"" }
	switch s.Name {
	case "readhdr":
		read()
		b.emitLine(fmt.Sprintf("%s=%s:%s=%s", G("org"), sub(0, 40), G("recid"), sub(40, 8)), true, []string{nm.get("org"), nm.get("recid")})
		read()
		b.emitLine(fmt.Sprintf("%s=CARD$", G("peer")), true, []string{nm.get("peer")})
		read()
		b.emitLine(fmt.Sprintf("%s=VAL(%s):%s=VAL(%s)", G("nmet"), sub(0, 3), G("nben"), sub(3, 3)), true, []string{nm.get("nmet"), nm.get("nben")})
	case "readmet":
		read()
		b.emitLine(fmt.Sprintf("%s=%s", I("mname"), sub(0, 40)), true, []string{nm.get("mname")})
		b.emitLine(fmt.Sprintf("%s=VAL(%s):%s=VAL(%s)", I("bsln"), sub(40, 6), I("curr"), sub(46, 6)), true, []string{nm.get("bsln"), nm.get("curr")})
	case "readben":
		read()
		b.emitLine(fmt.Sprintf("%s=%s:%s=VAL(%s)", I("bname"), sub(0, 40), I("medn"), sub(40, 7)), true, []string{nm.get("bname"), nm.get("medn")})
	case "pagehdr":
		pr(q(r.HdrLeft), G("recid"), q(r.HdrMid), G("ipage"))
	case "title1":
		pr(q(r.Title1), G("org"))
	case "colhdr1":
		pr(q(r.ColHdr1))
	case "rowmet":
		pr("\" \"", I("mname"), "\" \"", I("bsln"), "\" \"", I("curr"), "\" \"", I("delt"))
	case "title2":
		pr(q(r.Title2), G("peer"))
	case "colhdr2":
		pr(q(r.ColHdr2))
	case "rowben":
		parts := []string{"\" \"", I("bname")}
		for k := 1; k <= 8; k++ {
			parts = append(parts, "\" \"", b.ex(Ix(nm.get("pctl"), N(k), a)))
		}
		pr(parts...)
	case "trailer":
		pr("\"0RECORDS PROCESSED \"", G("nmet"), "\" METRICS \"", G("nben"), "\" BENCHMARKS   SEAL \"", G("chksum"))
	case "errmsg":
		pr(q(fmt.Sprintf(" *** %s%03dE %s ***", j.id.Program[:3], a.Val*10+j.x.intn(10), errText[a.Val])))
	case "debugmsg":
		pr(q(fmt.Sprintf(" DEBUG CHECKPOINT %04d", a.Val)))
	case "debugmet":
		pr("\" DBG MET \"", b.ex(a), "\" \"", I("bsln"), "\" \"", I("curr"), "\" \"", I("delt"))
	case "debugben":
		pr("\" DBG BEN \"", b.ex(a), "\" \"", I("medn"))
	}
}
