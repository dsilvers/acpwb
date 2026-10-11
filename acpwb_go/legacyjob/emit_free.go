package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

// freeEmitter renders the IR as C, Perl 5 or Python 2. The three share
// statement structure; they differ in block syntax, sigils, I/O idiom and
// comment style.
type freeEmitter struct {
	j          *job
	lang       string // c perl py2
	out        []Line
	ind        string // one indent unit
	kr         bool   // K&R C
	c99        bool   // // comments in C
	oldP       bool   // pre-5.6 Perl: use vars, &calls, no 'our'
	xr         bool   // python: xrange
	floorDiv   bool
	whileLoops float64
	unless     float64
	globals    map[string]bool
	consts     map[string]string
	arrays     map[string]bool
	cardStyle  int // card-field idiom: 0 helper/substr/slice, 1 and 2 alternatives
}

func newFreeEmitter(j *job, lang string) *freeEmitter {
	x := j.x
	e := &freeEmitter{j: j, lang: lang, globals: map[string]bool{}, consts: map[string]string{}, arrays: map[string]bool{}}
	e.ind = pick(x, []string{"    ", "  ", "\t", "    ", "   "})
	if lang == "py2" {
		e.ind = pick(x, []string{"    ", "    ", "  ", "\t"})
	}
	e.kr = lang == "c" && (j.in.Year < 1992 || x.chance(0.3))
	e.c99 = lang == "c" && j.in.Year >= 2003 && x.chance(0.5)
	e.oldP = lang == "perl" && (j.in.Year < 2003 || x.chance(0.25))
	e.xr = x.chance(0.6)
	e.floorDiv = x.chance(0.4)
	e.whileLoops = []float64{0, 0.15, 0.35}[x.intn(3)]
	e.cardStyle = x.intn(3)
	e.unless = []float64{0, 0.3}[x.intn(2)]
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
	return e
}

func (e *freeEmitter) emitLine(depth int, text string, refs []string) {
	e.out = append(e.out, Line{Text: strings.Repeat(e.ind, depth) + text, Code: true, Refs: refs})
}

func (e *freeEmitter) blank() { e.out = append(e.out, Line{Text: ""}) }

// ── comments ───────────────────────────────────────────────────────────────

func (e *freeEmitter) commentText(depth int, lines []string) {
	pad := strings.Repeat(e.ind, depth)
	switch e.lang {
	case "c":
		if e.c99 {
			for _, l := range lines {
				e.out = append(e.out, Line{Text: pad + "// " + l})
			}
			return
		}
		if len(lines) == 1 {
			e.out = append(e.out, Line{Text: pad + "/* " + lines[0] + " */"})
			return
		}
		e.out = append(e.out, Line{Text: pad + "/*"})
		for _, l := range lines {
			e.out = append(e.out, Line{Text: pad + " * " + l})
		}
		e.out = append(e.out, Line{Text: pad + " */"})
	default:
		for _, l := range lines {
			e.out = append(e.out, Line{Text: strings.TrimRight(pad+"# "+l, " ")})
		}
	}
}

func (e *freeEmitter) comment(depth int, text string) {
	cs := e.j.cs
	cs.width = maxi(40, 76-len(e.ind)*depth)
	e.commentText(depth, cs.para(text))
}

func (e *freeEmitter) box(depth int, lines []string) {
	if e.lang == "c" && !e.c99 {
		pad := strings.Repeat(e.ind, depth)
		e.out = append(e.out, Line{Text: pad + "/" + strings.Repeat("*", 74)})
		for _, l := range lines {
			e.out = append(e.out, Line{Text: pad + " * " + l})
		}
		e.out = append(e.out, Line{Text: pad + " " + strings.Repeat("*", 74) + "/"})
		return
	}
	e.commentText(depth, lines)
}

// ── names / expressions ────────────────────────────────────────────────────

func (e *freeEmitter) name(n string, arrayCtx bool) string {
	if v, ok := e.consts[n]; ok {
		switch e.lang {
		case "c":
			return strings.ToUpper(n)
		case "perl":
			_ = v
			return strings.ToUpper(n)
		default:
			return strings.ToUpper(n)
		}
	}
	if e.lang == "perl" {
		if arrayCtx {
			return "$" + n
		}
		if e.arrays[n] {
			return "@" + n
		}
		return "$" + n
	}
	return n
}

func (e *freeEmitter) ex(x *Expr) string {
	switch x.K {
	case "var":
		return e.name(x.Name, false)
	case "idx":
		s := e.name(x.Name, true)
		for _, a := range x.Args {
			s += "[" + e.ex(a) + "]"
		}
		if len(x.Args) == 2 {
			// IR order is (rung, column); store as [rung][column].
			s = e.name(x.Name, true) + "[" + e.ex(x.Args[0]) + "][" + e.ex(x.Args[1]) + "]"
		}
		return s
	case "int":
		return strconv.Itoa(x.Val)
	case "dlit":
		return x.S
	case "bin":
		l, r := x.Args[0], x.Args[1]
		ls, rs := e.ex(l), e.ex(r)
		if l.K == "bin" && prec(l.Op) < prec(x.Op) {
			ls = "(" + ls + ")"
		}
		if r.K == "bin" && (prec(r.Op) < prec(x.Op) || (prec(r.Op) == prec(x.Op) && (x.Op == "-" || x.Op == "/"))) {
			rs = "(" + rs + ")"
		}
		if x.Op == "/" && !isDouble(l) && !isDouble(r) {
			switch e.lang {
			case "perl":
				return "int(" + ls + " / " + rs + ")"
			case "py2":
				if e.floorDiv {
					return ls + " // " + rs
				}
			}
		}
		return ls + " " + x.Op + " " + rs
	case "call":
		a := e.ex(x.Args[0])
		switch x.Name {
		case "mod":
			return wrapIfBin(x.Args[0], a) + " % " + wrapIfBin(x.Args[1], e.ex(x.Args[1]))
		case "dble":
			switch e.lang {
			case "c":
				return "(double)" + wrapIfBin(x.Args[0], a)
			case "py2":
				return "float(" + a + ")"
			}
			return a
		}
		return x.Name + "(" + a + ")"
	case "ladder":
		m, f := e.ex(x.Args[0]), e.ex(x.Args[1])
		switch e.lang {
		case "c":
			return "(int)((double)" + m + " * " + wrapIfBin(x.Args[1], f) + ")"
		default:
			return "int(" + m + " * " + wrapIfBin(x.Args[1], f) + ")"
		}
	case "str":
		return "\"" + x.S + "\""
	}
	return "?"
}

func isDouble(x *Expr) bool { return x.K == "dlit" || (x.K == "call" && x.Name == "dble") }

func wrapIfBin(x *Expr, s string) string {
	if x.K == "bin" {
		return "(" + s + ")"
	}
	return s
}

var cRel = map[string]string{"lt": "<", "le": "<=", "eq": "==", "ne": "!=", "gt": ">", "ge": ">="}

func (e *freeEmitter) cond(c *Cond) string {
	and, or := " && ", " || "
	if e.lang == "py2" {
		and, or = " and ", " or "
	}
	switch c.Op {
	case "and":
		return e.cond(c.A) + and + e.cond(c.B)
	case "or":
		return e.cond(c.A) + or + e.cond(c.B)
	}
	return e.ex(c.L) + " " + cRel[c.Op] + " " + e.ex(c.R)
}

// ── statements ──────────────────────────────────────────────────────────────

func (e *freeEmitter) open(depth int, head string, refs []string) {
	if e.lang == "py2" {
		e.emitLine(depth, head+":", refs)
		return
	}
	if e.lang == "c" && e.j.x.chance(0.3) {
		e.emitLine(depth, head, refs)
		e.emitLine(depth, "{", nil)
		return
	}
	e.emitLine(depth, head+" {", refs)
}

func (e *freeEmitter) close(depth int) {
	if e.lang != "py2" {
		e.emitLine(depth, "}", nil)
	}
}

func (e *freeEmitter) body(ss []*Stmt, depth int) {
	n := 0
	for _, s := range ss {
		if s.K != "comment" && s.K != "blank" {
			n++
		}
		e.one(s, depth)
	}
	if n == 0 && e.lang == "py2" {
		e.emitLine(depth, "pass", nil)
	}
}

func (e *freeEmitter) semi() string {
	if e.lang == "py2" {
		return ""
	}
	return ";"
}

func (e *freeEmitter) one(s *Stmt, depth int) {
	x := e.j.x
	switch s.K {
	case "comment":
		if s.Banner {
			e.box(depth, e.j.cs.box(s.Title, s.Text))
		} else {
			e.comment(depth, s.Text)
		}
	case "blank":
		e.blank()
	case "assign":
		e.emitLine(depth, e.ex(s.L)+" = "+e.ex(s.R)+e.semi(), refsOf(s.L, s.R))
	case "call":
		var args []string
		for _, a := range s.Args {
			args = append(args, e.ex(a))
		}
		call := s.Name + "(" + strings.Join(args, ", ") + ")"
		if e.lang == "perl" && e.oldP {
			call = "&" + call
		}
		e.emitLine(depth, call+e.semi(), append([]string{s.Name}, refsOf(s.Args...)...))
	case "stop":
		switch e.lang {
		case "c":
			e.emitLine(depth, pick(x, []string{"exit(0);", "exit(EXIT_SUCCESS);"}), nil)
		case "perl":
			e.emitLine(depth, "exit 0;", nil)
		default:
			e.emitLine(depth, "sys.exit(0)", nil)
		}
	case "io":
		e.io(s, depth)
	case "do":
		e.do(s, depth)
	case "if":
		e.ifStmt(s, depth)
	}
}

func (e *freeEmitter) do(s *Stmt, depth int) {
	x := e.j.x
	v := e.name(s.Var, false)
	from, to := e.ex(s.From), e.ex(s.To)
	refs := append([]string{s.Var}, refsOf(s.From, s.To)...)
	cmp, inc := "<=", "++"
	if s.Step < 0 {
		cmp, inc = ">=", "--"
	}
	if x.float() < e.whileLoops && e.lang != "py2" {
		e.emitLine(depth, v+" = "+from+";", refs)
		e.open(depth, "while ("+v+" "+cmp+" "+to+")", refs)
		e.body(s.Body, depth+1)
		e.emitLine(depth+1, v+inc+";", nil)
		e.close(depth)
		return
	}
	switch e.lang {
	case "c":
		e.open(depth, fmt.Sprintf("for (%s = %s; %s %s %s; %s%s)", v, from, v, cmp, to, v, inc), refs)
	case "perl":
		if x.chance(0.5) {
			rng := "(" + from + " .. " + to + ")"
			if s.Step < 0 {
				rng = "(reverse " + to + " .. " + from + ")"
			}
			e.open(depth, "foreach "+v+" "+rng, refs)
		} else {
			e.open(depth, fmt.Sprintf("for (%s = %s; %s %s %s; %s%s)", v, from, v, cmp, to, v, inc), refs)
		}
	default:
		rf := "range"
		if e.xr {
			rf = "xrange"
		}
		if s.Step < 0 {
			e.open(depth, fmt.Sprintf("for %s in %s(%s, %s - 1, -1)", v, rf, from, to), refs)
		} else {
			e.open(depth, fmt.Sprintf("for %s in %s(%s, %s + 1)", v, rf, from, to), refs)
		}
	}
	e.body(s.Body, depth+1)
	e.close(depth)
}

func (e *freeEmitter) ifStmt(s *Stmt, depth int) {
	x := e.j.x
	refs := condRefs(s.Cond)
	if e.lang == "perl" && s.Else == nil && x.float() < e.unless {
		e.open(depth, "unless ("+e.cond(negate(s.Cond))+")", refs)
	} else if e.lang == "py2" {
		e.open(depth, "if "+e.cond(s.Cond), refs)
	} else {
		e.open(depth, "if ("+e.cond(s.Cond)+")", refs)
	}
	e.body(s.Then, depth+1)
	if s.Else != nil {
		switch e.lang {
		case "py2":
			e.emitLine(depth, "else:", nil)
		default:
			e.emitLine(depth, "} else {", nil)
		}
		e.body(s.Else, depth+1)
	}
	e.close(depth)
}

// ── I/O ────────────────────────────────────────────────────────────────────

func cq(s string) string { // C / Perl double-quoted literal
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "%", "%%", "$", "\\$", "@", "\\@").Replace(s)
}

func pyq(s string) string { return strings.NewReplacer("\\", "\\\\", "'", "\\'", "%", "%%").Replace(s) }

func (e *freeEmitter) printf(depth int, format string, args []string, refs []string) {
	switch e.lang {
	case "c":
		f := strings.ReplaceAll(cq(format), "%%", "%")
		if len(args) == 0 {
			e.emitLine(depth, "printf(\""+f+"\\n\");", refs)
			return
		}
		e.emitLine(depth, "printf(\""+f+"\\n\", "+strings.Join(args, ", ")+");", refs)
	case "perl":
		f := strings.ReplaceAll(cq(format), "%%", "%")
		if len(args) == 0 {
			e.emitLine(depth, "printf(\""+f+"\\n\");", refs)
			return
		}
		e.emitLine(depth, "printf(\""+f+"\\n\", "+strings.Join(args, ", ")+");", refs)
	default:
		f := strings.ReplaceAll(pyq(format), "%%", "%")
		if len(args) == 0 {
			e.emitLine(depth, "print '"+strings.ReplaceAll(f, "%", "%%")+"'", refs)
			return
		}
		if len(args) == 1 {
			e.emitLine(depth, "print '"+f+"' % "+args[0], refs)
			return
		}
		e.emitLine(depth, "print '"+f+"' % ("+strings.Join(args, ", ")+")", refs)
	}
}

// lit escapes literal text inside a printf format.
func flit(s string) string { return strings.ReplaceAll(s, "%", "%%") }

func (e *freeEmitter) readCard(depth int) string {
	switch e.lang {
	case "c":
		e.emitLine(depth, "if (fgets(card, sizeof card, stdin) == NULL) "+pick(e.j.x, []string{"eof_error();", "exit(8);", "goto bad_deck;"}), nil)
		return "card"
	case "perl":
		e.emitLine(depth, "$card = <STDIN>;", nil)
		e.emitLine(depth, "chomp($card);", nil)
		return "$card"
	default:
		e.emitLine(depth, "card = sys.stdin.readline().rstrip('\\n')", nil)
		return "card"
	}
}

func (e *freeEmitter) field(depth int, target, card string, start, width int, num bool, refs []string) {
	switch e.lang {
	case "c":
		switch {
		case e.cardStyle == 1 && num:
			e.emitLine(depth, fmt.Sprintf("{ char t[16]; memcpy(t, %s + %d, %d); t[%d] = '\\0'; %s = atoi(t); }", card, start, width, width, target), refs)
		case e.cardStyle == 2 && num:
			e.emitLine(depth, fmt.Sprintf("if (sscanf(%s + %d, \"%%%dd\", &%s) != 1) %s = 0;", card, start, width, target, target), refs)
		case e.cardStyle > 0:
			e.emitLine(depth, fmt.Sprintf("sprintf(%s, \"%%-%d.%ds\", %s + %d);", target, width, width, card, start), refs)
		case num:
			e.emitLine(depth, fmt.Sprintf("%s = fieldnum(%s, %d, %d);", target, card, start, width), refs)
		default:
			e.emitLine(depth, fmt.Sprintf("fieldstr(%s, %s, %d, %d);", target, card, start, width), refs)
		}
	case "perl":
		switch {
		case e.cardStyle == 1 && num:
			e.emitLine(depth, fmt.Sprintf("(%s) = unpack(\"x%d A%d\", %s); %s += 0;", target, start, width, card, target), refs)
		case e.cardStyle == 1:
			e.emitLine(depth, fmt.Sprintf("%s = sprintf(\"%%-%ds\", unpack(\"x%d a%d\", %s));", target, width, start, width, card), refs)
		case e.cardStyle == 2 && num:
			e.emitLine(depth, fmt.Sprintf("(%s) = %s =~ /^.{%d}\\s*(\\d*)/; %s ||= 0;", target, card, start, target), refs)
		case e.cardStyle == 2:
			e.emitLine(depth, fmt.Sprintf("(%s) = (%s . (' ' x %d)) =~ /^.{%d}(.{%d})/;", target, card, start+width, start, width), refs)
		case num:
			e.emitLine(depth, fmt.Sprintf("%s = substr(%s, %d, %d) + 0;", target, card, start, width), refs)
		default:
			e.emitLine(depth, fmt.Sprintf("%s = sprintf(\"%%-%ds\", substr(%s, %d, %d));", target, width, card, start, width), refs)
		}
	default:
		switch {
		case e.cardStyle == 1 && num:
			e.emitLine(depth, fmt.Sprintf("%s = int(%s[%d:%d].strip() or '0')", target, card, start, start+width), refs)
		case e.cardStyle == 2 && num:
			e.emitLine(depth, fmt.Sprintf("%s = int(filter(str.isdigit, %s[%d:%d]) or 0)", target, card, start, start+width), refs)
		case e.cardStyle == 2:
			e.emitLine(depth, fmt.Sprintf("%s = '%%-%d.%ds' %% %s[%d:]", target, width, width, card, start), refs)
		case num:
			e.emitLine(depth, fmt.Sprintf("%s = int(%s[%d:%d] or 0)", target, card, start, start+width), refs)
		default:
			e.emitLine(depth, fmt.Sprintf("%s = %s[%d:%d].ljust(%d)", target, card, start, start+width, width), refs)
		}
	}
}

func (e *freeEmitter) io(s *Stmt, depth int) {
	nm := e.j.nm
	r := e.j.rep
	var a *Expr
	if len(s.Args) > 0 {
		a = s.Args[0]
	}
	I := func(role string) string { return e.ex(Ix(nm.get(role), a)) }
	G := func(role string) string { return e.ex(V(nm.get(role))) }
	switch s.Name {
	case "readhdr":
		c := e.readCard(depth)
		e.field(depth, G("org"), c, 0, 40, false, []string{nm.get("org")})
		e.field(depth, G("recid"), c, 40, 8, false, []string{nm.get("recid")})
		c = e.readCard(depth)
		e.field(depth, G("peer"), c, 0, 60, false, []string{nm.get("peer")})
		c = e.readCard(depth)
		e.field(depth, G("nmet"), c, 0, 3, true, []string{nm.get("nmet")})
		e.field(depth, G("nben"), c, 3, 3, true, []string{nm.get("nben")})
	case "readmet":
		c := e.readCard(depth)
		e.field(depth, I("mname"), c, 0, 40, false, []string{nm.get("mname")})
		e.field(depth, I("bsln"), c, 40, 6, true, []string{nm.get("bsln")})
		e.field(depth, I("curr"), c, 46, 6, true, []string{nm.get("curr")})
	case "readben":
		c := e.readCard(depth)
		e.field(depth, I("bname"), c, 0, 40, false, []string{nm.get("bname")})
		e.field(depth, I("medn"), c, 40, 7, true, []string{nm.get("medn")})
	case "pagehdr":
		e.printf(depth, flit(r.HdrLeft)+"%-8s"+flit(r.HdrMid)+"%5d", []string{G("recid"), G("ipage")}, []string{nm.get("recid"), nm.get("ipage")})
	case "title1":
		e.printf(depth, flit(r.Title1)+"%s", []string{G("org")}, []string{nm.get("org")})
	case "colhdr1":
		e.printf(depth, flit(r.ColHdr1), nil, nil)
	case "rowmet":
		e.printf(depth, " %-40.40s%9d%9d%+9d", []string{I("mname"), I("bsln"), I("curr"), I("delt")}, []string{nm.get("mname"), nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	case "title2":
		e.printf(depth, flit(r.Title2)+"%s", []string{G("peer")}, []string{nm.get("peer")})
	case "colhdr2":
		e.printf(depth, flit(r.ColHdr2), nil, nil)
	case "rowben":
		var args []string
		for k := 1; k <= 8; k++ {
			args = append(args, e.ex(Ix(nm.get("pctl"), N(k), a)))
		}
		e.printf(depth, " %-40.40s"+strings.Repeat("%9d", 8), append([]string{I("bname")}, args...), []string{nm.get("bname"), nm.get("pctl")})
	case "trailer":
		e.printf(depth, "0RECORDS PROCESSED%6d METRICS%6d BENCHMARKS   SEAL %6d", []string{G("nmet"), G("nben"), G("chksum")}, []string{nm.get("nmet"), nm.get("nben"), nm.get("chksum")})
	case "errmsg":
		e.printf(depth, flit(fmt.Sprintf(" *** %s%03dE %s ***", e.j.id.Program[:3], a.Val*10+e.j.x.intn(10), errText[a.Val])), nil, nil)
	case "debugmsg":
		e.printf(depth, fmt.Sprintf(" DEBUG CHECKPOINT %04d", a.Val), nil, nil)
	case "debugmet":
		e.printf(depth, " DBG MET%4d%10d%10d%10d", []string{e.ex(a), I("bsln"), I("curr"), I("delt")}, []string{nm.get("bsln")})
	case "debugben":
		e.printf(depth, " DBG BEN%4d%10d", []string{e.ex(a), I("medn")}, []string{nm.get("medn")})
	}
}

// ── declarations & units ───────────────────────────────────────────────────

func (e *freeEmitter) cType(s Sym) string {
	switch s.Type {
	case "dbl":
		return "double"
	case "str":
		return "char"
	}
	return pick(e.j.x, []string{"int", "int", "long"})
}

func (e *freeEmitter) cDecl(s Sym, static bool) string {
	t := e.cType(s)
	n := s.Name
	if s.Type == "str" {
		if len(s.Dims) > 0 {
			n += fmt.Sprintf("[%s + 1][%d]", e.dimName(s.Dims[0]), s.Len+1)
		} else {
			n += fmt.Sprintf("[%d]", s.Len+1)
		}
	} else if len(s.Dims) == 2 {
		n += fmt.Sprintf("[%d][%s + 1]", s.Dims[0]+1, e.dimName(s.Dims[1]))
	} else if len(s.Dims) == 1 {
		n += fmt.Sprintf("[%s + 1]", e.dimName(s.Dims[0]))
	}
	pre := ""
	if static {
		pre = "static "
	}
	if len(s.Init) > 0 {
		vals := []string{"0"}
		if s.Type == "str" {
			return pre + "char " + s.Name + "[] = \"" + s.Init[0] + "\";"
		}
		vals = append(vals, s.Init...)
		if len(s.Dims) == 0 {
			vals = s.Init
			return pre + t + " " + n + " = " + vals[0] + ";"
		}
		return pre + t + " " + n + " = { " + strings.Join(vals, ", ") + " };"
	}
	return pre + t + " " + n + ";"
}

func (e *freeEmitter) dimName(d int) string {
	for _, g := range e.j.p.Globals {
		if g.Role == "const" && g.Init[0] == strconv.Itoa(d) {
			return strings.ToUpper(g.Name)
		}
	}
	return strconv.Itoa(d)
}

func (e *freeEmitter) initVal(s Sym) string {
	size := func(d int) string { return e.dimName(d) + " + 1" }
	switch e.lang {
	case "perl":
		if len(s.Init) > 0 {
			if s.Type == "str" {
				return "'" + s.Init[0] + "'"
			}
			if len(s.Dims) == 0 {
				return s.Init[0]
			}
			return "(undef, " + strings.Join(s.Init, ", ") + ")"
		}
		if len(s.Dims) > 0 {
			return "()"
		}
		if s.Type == "str" {
			return "''"
		}
		return "0"
	default:
		if len(s.Init) > 0 {
			if s.Type == "str" {
				return "'" + s.Init[0] + "'"
			}
			if len(s.Dims) == 0 {
				return s.Init[0]
			}
			return "[None, " + strings.Join(s.Init, ", ") + "]"
		}
		if len(s.Dims) == 2 {
			return fmt.Sprintf("[[0] * (%s) for _k in range(%d)]", size(s.Dims[1]), s.Dims[0]+1)
		}
		if len(s.Dims) == 1 {
			if s.Type == "str" {
				return fmt.Sprintf("[''] * (%s)", size(s.Dims[0]))
			}
			return fmt.Sprintf("[0] * (%s)", size(s.Dims[0]))
		}
		if s.Type == "str" {
			return "''"
		}
		return "0"
	}
}

func (e *freeEmitter) header() {
	j := e.j
	x := j.x
	var head []string
	head = append(head, j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy"))...)
	head = append(head, j.historyLines()...)
	head = append(head, j.cs.box("Provenance", "Seal "+j.in.Token+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos"))...)
	switch e.lang {
	case "perl":
		e.emitLine(0, pick(x, []string{"#!/usr/bin/perl -w", "#!/usr/local/bin/perl", "#!/usr/bin/env perl"}), nil)
		e.box(0, head)
		if !e.oldP {
			e.emitLine(0, "use strict;", nil)
			e.emitLine(0, "use warnings;", nil)
		} else {
			e.emitLine(0, "require 5.004;", nil)
		}
		e.emitLine(0, "use constant {", nil)
		for _, s := range j.p.Globals {
			if s.Role == "const" {
				e.emitLine(1, strings.ToUpper(s.Name)+" => "+s.Init[0]+",", nil)
			}
		}
		e.emitLine(0, "};", nil)
		var vars []string
		for _, s := range j.p.Globals {
			if s.Role != "const" {
				vars = append(vars, e.name(s.Name, false))
			}
		}
		vars = append(vars, "$card")
		if e.oldP {
			e.emitLine(0, "use vars qw("+strings.Join(vars, " ")+");", nil)
		} else {
			e.emitLine(0, "our ("+strings.Join(vars, ", ")+");", nil)
		}
		for _, s := range j.p.Globals {
			if s.Role != "const" && (len(s.Init) > 0) {
				e.emitLine(0, e.name(s.Name, false)+" = "+e.initVal(s)+";", []string{s.Name})
			}
		}
		e.blank()
		// POD
		e.out = append(e.out, Line{Text: "=head1 NAME"}, Line{Text: ""}, Line{Text: strings.ToLower(j.id.Program) + " - " + j.in.Title}, Line{Text: ""},
			Line{Text: "=head1 DESCRIPTION"}, Line{Text: ""})
		for _, p := range wrapWords(asciiText(j.g.comment(nil, "history", "data", "units")), 76) {
			e.out = append(e.out, Line{Text: p})
		}
		e.out = append(e.out, Line{Text: ""}, Line{Text: "=head1 AUTHOR"}, Line{Text: ""}, Line{Text: j.g.roster[0].Name() + " <" + j.g.email(j.g.roster[0]) + ">"}, Line{Text: ""}, Line{Text: "=cut"}, Line{Text: ""})
	case "py2":
		e.emitLine(0, pick(x, []string{"#!/usr/bin/env python", "#!/usr/bin/python", "#!/usr/bin/python2.7"}), nil)
		e.emitLine(0, "# -*- coding: "+pick(x, []string{"latin-1", "utf-8", "ascii"})+" -*-", nil)
		e.out = append(e.out, Line{Text: "\"\"\"" + strings.ToLower(j.id.Program) + " -- " + asciiText(j.in.Title)})
		e.out = append(e.out, Line{Text: ""})
		for _, p := range wrapWords(asciiText(j.g.comment(nil, "history", "data")), 76) {
			e.out = append(e.out, Line{Text: p})
		}
		e.out = append(e.out, Line{Text: "\"\"\""})
		e.box(0, head)
		if x.chance(0.5) {
			e.emitLine(0, "from __future__ import "+pick(x, []string{"with_statement", "print_function # TODO: actually use it", "absolute_import"}), nil)
		}
		e.emitLine(0, "import sys", nil)
		if x.chance(0.6) {
			e.emitLine(0, "import "+pick(x, []string{"string", "os, time", "getopt", "re, string"}), nil)
		}
		e.blank()
		for _, s := range j.p.Globals {
			if s.Role == "const" {
				e.emitLine(0, strings.ToUpper(s.Name)+" = "+s.Init[0], nil)
			}
		}
		for _, s := range j.p.Globals {
			if s.Role != "const" {
				e.emitLine(0, s.Name+" = "+e.initVal(s), []string{s.Name})
			}
		}
		e.emitLine(0, "card = ''", nil)
	default: // C
		e.box(0, head)
		e.emitLine(0, "#include <stdio.h>", nil)
		e.emitLine(0, "#include <stdlib.h>", nil)
		e.emitLine(0, "#include <string.h>", nil)
		if j.id.LangKey == "cbtr" {
			e.emitLine(0, "#include \"btrapi.h\"   /* Btrieve: BTRV(), B_OPEN, B_INSERT ... */", nil)
		}
		if x.chance(0.4) {
			e.emitLine(0, "#include <"+pick(x, []string{"ctype.h", "math.h", "time.h"})+">", nil)
		}
		e.blank()
		for _, s := range j.p.Globals {
			if s.Role == "const" {
				e.emitLine(0, fmt.Sprintf("#define %-10s %s", strings.ToUpper(s.Name), s.Init[0]), nil)
			}
		}
		e.blank()
		e.comment(0, "Arrays are one-based, as FORTRAN intended. Element zero is unused. "+j.g.comment(nil, "workarounds", "history"))
		for _, s := range j.p.Globals {
			if s.Role != "const" {
				e.emitLine(0, e.cDecl(s, true), []string{s.Name})
				e.out[len(e.out)-1].Defs = []string{s.Name}
			}
		}
		e.emitLine(0, "static char card[132];", nil)
		e.blank()
		// card helpers
		e.comment(0, "Card field helpers. Columns are zero-based here and one-based everywhere else. Sorry. -- "+j.g.coder().Initials)
		if e.kr {
			e.emitLine(0, "static int fieldnum(c, start, width)", nil)
			e.emitLine(0, "char *c; int start, width;", nil)
		} else {
			e.emitLine(0, "static int fieldnum(const char *c, int start, int width)", nil)
		}
		e.emitLine(0, "{", nil)
		e.emitLine(1, "char buf[16];", nil)
		e.emitLine(1, "strncpy(buf, c + start, width);", nil)
		e.emitLine(1, "buf[width] = '\\0';", nil)
		e.emitLine(1, "return atoi(buf);", nil)
		e.emitLine(0, "}", nil)
		e.blank()
		if e.kr {
			e.emitLine(0, "static void fieldstr(dst, c, start, width)", nil)
			e.emitLine(0, "char *dst; char *c; int start, width;", nil)
		} else {
			e.emitLine(0, "static void fieldstr(char *dst, const char *c, int start, int width)", nil)
		}
		e.emitLine(0, "{", nil)
		e.emitLine(1, "int n = (int)strlen(c) - start;", nil)
		e.emitLine(1, "memset(dst, ' ', width);", nil)
		e.emitLine(1, "if (n > 0) memcpy(dst, c + start, n < width ? n : width);", nil)
		e.emitLine(1, "dst[width] = '\\0';", nil)
		e.emitLine(0, "}", nil)
		e.blank()
		if !e.kr {
			for _, p := range j.p.Procs[1:] {
				e.emitLine(0, "static void "+p.Name+"("+e.cParams(p, true)+");", nil)
			}
			e.emitLine(0, "static void eof_error(void);", nil)
			e.blank()
		}
	}
}

func (e *freeEmitter) cParams(p *Proc, proto bool) string {
	if len(p.Params) == 0 {
		if proto && !e.kr {
			return "void"
		}
		return ""
	}
	var ps []string
	for _, q := range p.Params {
		var s Sym
		for _, l := range p.Locals {
			if l.Name == q {
				s = l
			}
		}
		t := "int"
		if s.Type == "dbl" {
			t = "double"
		}
		if len(s.Dims) > 0 {
			ps = append(ps, t+" *"+q)
		} else {
			ps = append(ps, t+" "+q)
		}
	}
	if e.kr && !proto {
		var names []string
		for _, q := range p.Params {
			names = append(names, q)
		}
		return strings.Join(names, ", ")
	}
	return strings.Join(ps, ", ")
}

// assignedGlobals: scalar globals assigned in a body (Python 'global').
func (e *freeEmitter) assignedGlobals(ss []*Stmt, acc map[string]bool) {
	for _, s := range ss {
		if s.K == "assign" && s.L.K == "var" && e.globals[s.L.Name] && !e.arrays[s.L.Name] {
			acc[s.L.Name] = true
		}
		if s.K == "do" && e.globals[s.Var] {
			acc[s.Var] = true
		}
		if s.K == "io" {
			for _, r := range []string{"org", "recid", "peer", "nmet", "nben"} {
				if s.Name == "readhdr" {
					acc[e.j.nm.get(r)] = true
				}
			}
		}
		e.assignedGlobals(s.Body, acc)
		e.assignedGlobals(s.Then, acc)
		e.assignedGlobals(s.Else, acc)
	}
}

func (e *freeEmitter) unit(p *Proc, callers, callees []string, isMain bool) {
	j := e.j
	e.blank()
	e.box(0, j.prologue(p, callers, callees))
	switch e.lang {
	case "c":
		if isMain {
			if e.kr {
				e.emitLine(0, "main(argc, argv)", nil)
				e.emitLine(0, "int argc; char **argv;", nil)
			} else {
				e.emitLine(0, "int main(int argc, char *argv[])", nil)
			}
		} else {
			e.emitLine(0, "static void "+p.Name+"("+e.cParams(p, false)+")", nil)
			if e.kr && len(p.Params) > 0 {
				e.emitLine(0, e.cParams(p, true)+";", nil)
			}
		}
		e.emitLine(0, "{", nil)
		for _, l := range p.Locals {
			if l.Role == "param" {
				continue
			}
			e.emitLine(1, e.cDecl(l, len(l.Init) > 0), []string{l.Name})
			e.out[len(e.out)-1].Defs = []string{l.Name}
		}
		e.body(p.Body, 1)
		if isMain && j.id.LangKey == "cbtr" {
			e.btrieve()
		}
		if isMain {
			e.emitLine(1, "return 0;", nil)
		}
		e.emitLine(0, "}", nil)
	case "perl":
		e.emitLine(0, "sub "+p.Name+" {", nil)
		if len(p.Params) > 0 {
			var ps []string
			for _, q := range p.Params {
				ps = append(ps, e.name(q, false))
			}
			e.emitLine(1, "my ("+strings.Join(ps, ", ")+") = @_;", p.Params)
		}
		for _, l := range p.Locals {
			if l.Role == "param" {
				continue
			}
			e.emitLine(1, "my "+e.name(l.Name, false)+" = "+e.initVal(l)+";", []string{l.Name})
			e.out[len(e.out)-1].Defs = []string{l.Name}
		}
		e.body(p.Body, 1)
		e.emitLine(0, "}", nil)
	default:
		e.emitLine(0, "def "+p.Name+"("+strings.Join(p.Params, ", ")+"):", nil)
		e.out = append(e.out, Line{Text: e.ind + "\"\"\"" + asciiText(p.Purpose) + "\"\"\""})
		acc := map[string]bool{}
		e.assignedGlobals(p.Body, acc)
		if len(acc) > 0 {
			var gs []string
			for k := range acc {
				gs = append(gs, k)
			}
			sortStrings(gs)
			e.emitLine(1, "global "+strings.Join(gs, ", "), nil)
		}
		for _, l := range p.Locals {
			if l.Role == "param" {
				continue
			}
			e.emitLine(1, l.Name+" = "+e.initVal(l), []string{l.Name})
			e.out[len(e.out)-1].Defs = []string{l.Name}
		}
		e.body(p.Body, 1)
	}
}

func (e *freeEmitter) emit() []Line {
	j := e.j
	callers, callees := callGraph(j.p)
	e.header()
	for i, p := range j.p.Procs {
		if i == 0 && e.lang != "c" {
			continue
		}
		e.unit(p, callers[p.Name], callees[p.Name], i == 0)
	}
	switch e.lang {
	case "c":
		e.blank()
		if e.kr {
			e.emitLine(0, "static void eof_error()", nil)
		} else {
			e.emitLine(0, "static void eof_error(void)", nil)
		}
		e.emitLine(0, "{", nil)
		e.emitLine(1, "fprintf(stderr, \"*** UNEXPECTED END OF DECK -- CHECK THE CARD COUNT ***\\n\");", nil)
		e.emitLine(1, "exit(12);", nil)
		e.emitLine(0, "}", nil)
	case "perl":
		e.unit(j.p.Procs[0], nil, callees[j.p.Procs[0].Name], true)
		e.blank()
		e.emitLine(0, j.p.Procs[0].Name+"();", nil)
		e.emitLine(0, "exit 0;", nil)
		e.blank()
		e.emitLine(0, "__END__", nil)
		for n := j.x.between(1, 4); n > 0; n-- {
			e.out = append(e.out, Line{Text: asciiText(j.g.comment(nil, "handoff", "misc", "history"))})
		}
	default:
		e.unit(j.p.Procs[0], nil, callees[j.p.Procs[0].Name], true)
		e.blank()
		e.blank()
		e.emitLine(0, "if __name__ == '__main__':", nil)
		e.emitLine(1, j.p.Procs[0].Name+"()", nil)
	}
	return e.out
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for k := i; k > 0 && s[k] < s[k-1]; k-- {
			s[k], s[k-1] = s[k-1], s[k]
		}
	}
}

// btrieve appends the run to a Btrieve history file (C + Btrieve jobs).
func (e *freeEmitter) btrieve() {
	j := e.j
	nm := j.nm
	e.comment(1, "Append this run to the Btrieve history file. "+j.g.comment(nil, "data", "vendors", "workarounds"))
	e.emitLine(1, "{", nil)
	e.emitLine(2, "static char posBlk[128];", nil)
	e.emitLine(2, "struct { char recid[8]; char metric[40]; long base, curr, chg; } hist;", nil)
	e.emitLine(2, "unsigned short len = sizeof hist;", nil)
	e.emitLine(2, "char key[64] = \"BNCHHIST.BTR\";", nil)
	e.emitLine(2, "int st, k;", nil)
	e.emitLine(2, "st = BTRV(B_OPEN, posBlk, &hist, &len, key, 0);", nil)
	e.emitLine(2, "if (st != 0) {", nil)
	e.emitLine(3, "fprintf(stderr, \"BTRIEVE STATUS %d ON OPEN -- IS BTRIEVE.EXE LOADED?\\n\", st);", nil)
	e.emitLine(2, "} else {", nil)
	e.emitLine(3, "for (k = 1; k <= "+e.name(nm.get("nmet"), false)+"; k++) {", []string{nm.get("nmet")})
	e.emitLine(4, "memcpy(hist.recid, "+e.name(nm.get("recid"), false)+", 8);", []string{nm.get("recid")})
	e.emitLine(4, "memcpy(hist.metric, "+e.name(nm.get("mname"), true)+"[k], 40);", []string{nm.get("mname")})
	e.emitLine(4, "hist.base = "+e.name(nm.get("bsln"), true)+"[k]; hist.curr = "+e.name(nm.get("curr"), true)+"[k]; hist.chg = "+e.name(nm.get("delt"), true)+"[k];", []string{nm.get("bsln"), nm.get("curr"), nm.get("delt")})
	e.emitLine(4, "st = BTRV(B_INSERT, posBlk, &hist, &len, key, 0);", nil)
	e.emitLine(4, "if (st == 5) st = 0;  /* duplicate key: loaded by an earlier run */", nil)
	e.emitLine(3, "}", nil)
	e.emitLine(3, "BTRV(B_CLOSE, posBlk, &hist, &len, key, 0);", nil)
	e.emitLine(2, "}", nil)
	e.emitLine(1, "}", nil)
}
