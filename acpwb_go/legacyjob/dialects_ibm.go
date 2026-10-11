package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

// Dialects for the IBM midrange / mainframe family that the shared walker
// can drive: REXX, TSO CLIST, AS/400 CL, Easytrieve Plus, ABAP and RPG.

var moreDialects = map[string]func(j *job) *dialect{
	"rexx": rexxDialect, "clist": clistDialect, "cl400": cl400Dialect, "ezt": eztDialect, "abap": abapDialect, "rpg": rpgDialect,
	"r": rDialect, "splus": rDialect, "matlab": matlabDialect, "octave": matlabDialect, "apl": aplDialect,
	"awk": awkDialect, "ksh": kshDialect, "tcl": tclDialect, "bc": bcDialect, "dc": dcDialect,
	"cfml": cfmlDialect, "asp": aspDialect, "modula2": modula2Dialect, "lisp": lispDialect, "smalltalk": smalltalkDialect, "forth": forthDialect,
}

// hundredths renders a decimal literal for integer-only languages: ladder
// rungs (< 10) become hundredths, anything larger keeps its integer part.
func hundredths(lit string) string {
	f, err := strconv.ParseFloat(strings.TrimSuffix(strings.ToUpper(lit), "D0"), 64)
	if err != nil {
		return lit
	}
	if f < 10 {
		return strconv.Itoa(int(f*100 + 0.5))
	}
	return strconv.Itoa(int(f))
}

// simpleIdx reports whether every subscript is a plain variable or literal.
func simpleIdx(x *Expr) bool {
	for _, a := range x.Args {
		if a.K != "var" && a.K != "int" {
			return false
		}
	}
	return true
}

// contLines joins words into lines no wider than w, ending every line but
// the last with cont.
func contLines(prefix string, words []string, w int, cont string) []string {
	var out []string
	cur := prefix
	for _, wd := range words {
		if len(cur)+len(wd)+1 > w && cur != prefix {
			out = append(out, cur+cont)
			cur = strings.Repeat(" ", len(prefix))
		}
		if cur != prefix && !strings.HasSuffix(cur, " ") {
			cur += " "
		}
		cur += wd
	}
	return append(out, cur)
}

// ── REXX ──────────────────────────────────────────────────────────────────

func rexxDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"  ", "   ", "    "}), cmt: "/* ", cmtEnd: " */",
		idx1: "%s.%s", idx2: "%s.%s.%s", intDiv: "%s %% %s", mod: "%s // %s", dble: "%s", ladder: "trunc(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "\\=", "gt": ">", "ge": ">="}, and: " & ", or: " | ",
		assign: "%s = %s", call: "call %s %s", callNone: "call %s", stop: "exit 16",
		ifOpen: "if %s then do", ifElse: "end\nelse do", ifClose: "end",
		forUp: "do %s = %s to %s", forDown: "do %s = %s to %s by -1", forClose: "end %s",
		concat: " || ", lit: sq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("left(%s, %d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("right(%s, %d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("right(sgn(%s), %d)", x, w) },
		print: func(p string) string { return "say " + p },
	}
	stemRef := func(e *dialectEmitter, x *Expr) string {
		var parts []string
		for _, a := range x.Args {
			parts = append(parts, "("+e.ex(a)+")")
		}
		return "'" + x.Name + ".'" + strings.Join(parts, "'.'")
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		if x.K == "idx" && !simpleIdx(x) {
			return "value(" + stemRef(e, x) + ")", true
		}
		return "", false
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		if l.K == "idx" && !simpleIdx(l) {
			return "call value " + stemRef(e, l) + ", " + r, true
		}
		return "", false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.raw("/* REXX " + e.j.id.Program + " -- " + asciiText(trunc(e.j.in.Title, 50)) + " */")
		e.headerComments()
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "trace "+pick(e.j.x, []string{"o", "n", "off", "normal"}))
			e.line(0, "numeric digits 15")
			e.line(0, "address tso")
			e.line(0, `"ALLOC F(DECK) DA('ACPWB.RSCH.DECK(`+e.j.id.Program+`)') SHR REUSE"`)
			e.line(0, `"EXECIO * DISKR DECK (STEM deck. FINIS"`)
			e.line(0, "if rc <> 0 then exit rc")
			e.line(0, "ncard = 0")
			constInits(e, 0)
			scalarInits(e, 0)
			elementInits(e, 0)
			return
		}
		expose := []string{"deck.", "ncard", "card"}
		for _, g := range e.usedGlobals(p) {
			if e.arrays[g] {
				g += "."
			}
			expose = append(expose, g)
		}
		e.raw(p.Name + ": procedure expose " + strings.Join(expose, " "))
		if len(p.Params) > 0 {
			e.line(1, "parse arg "+strings.Join(p.Params, ", "))
		}
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, `"FREE F(DECK)"`)
			e.line(0, "exit 0")
			return
		}
		e.line(1, "return")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw("sgn: procedure")
		e.line(1, "parse arg n")
		e.line(1, "if n > 0 then return '+'n")
		e.line(1, "return n")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "ncard = ncard + 1\ncard = deck.ncard") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = strip(substr(card, %d, %d)) + 0", t, start+1, width)
		}
		return fmt.Sprintf("%s = substr(card, %d, %d)", t, start+1, width)
	}
	return d
}

// ── TSO CLIST ─────────────────────────────────────────────────────────────

func clistDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"  ", "   "}), cmt: "/* ", cmtEnd: " */", sigil: "&", upper: true, byGlobal: true,
		idx1: "&%s%s", idx2: "&%s%sX%s", intDiv: "%s / %s", mod: "%s // %s", dble: "%s", ladder: "(%s * %s) / 100",
		rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"}, and: " AND ", or: " OR ",
		assign: "SET %s = %s", call: "SYSCALL %s", callNone: "SYSCALL %s", stop: "EXIT CODE(16)",
		ifOpen: "IF %s THEN DO", ifElse: "END\nELSE DO", ifClose: "END",
		forUp: "DO %s = %s TO %s", forDown: "DO %s = %s TO %s BY -1", forClose: "END",
		concat: "", lit: func(s string) string { return strings.ReplaceAll(cardText(s), "&", "&&") }, order: "main-first",
		pstr: func(x string, w int) string {
			return fmt.Sprintf("&SUBSTR(1:%d,&STR(%s%s))", w, x, strings.Repeat(" ", w))
		},
		pnum: func(x string, w int) string {
			return fmt.Sprintf("&SUBSTR(1:%d,&STR(%s)%s)", w, x, strings.Repeat(" ", w))
		},
		psgn: func(x string, w int) string {
			return fmt.Sprintf("&SUBSTR(1:%d,&STR(%s)%s)", w, x, strings.Repeat(" ", w))
		},
		print: func(p string) string { return "WRITE " + p },
	}
	// CLIST has no arrays: element I of BSLN is the symbolic variable
	// BSLN<value of I>, read back with double ampersands.
	sub := func(e *dialectEmitter, x *Expr) string {
		var b strings.Builder
		for i, a := range x.Args {
			if i > 0 {
				b.WriteString("X")
			}
			switch a.K {
			case "int":
				b.WriteString(strconv.Itoa(a.Val))
			case "var":
				b.WriteString("&" + a.Name)
				if i < len(x.Args)-1 {
					b.WriteString(".")
				}
			default:
				e.tmpN["ix"]++
				t := fmt.Sprintf("IX%d", e.tmpN["ix"])
				e.pending = append(e.pending, Line{Text: fmt.Sprintf("SET &%s = %s", t, e.ex(a)), Code: true})
				b.WriteString("&" + t)
				if i < len(x.Args)-1 {
					b.WriteString(".")
				}
			}
		}
		return b.String()
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "dlit":
			return hundredths(x.S), true
		case "idx":
			return "&&" + x.Name + sub(e, x), true
		}
		return "", false
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		if l.K == "idx" {
			return "SET &" + l.Name + sub(e, l) + " = " + r, true
		}
		return "", false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("PROC 0 DEBUG")
		e.line(0, "CONTROL NOMSG NOLIST NOCONLIST NOSYMLIST NOFLUSH")
		var names []string
		for _, s := range e.j.p.Globals {
			if len(s.Dims) == 0 {
				names = append(names, "&"+s.Name)
			}
		}
		for _, s := range e.byGlobalSyms() {
			if len(s.Dims) == 0 {
				names = append(names, "&"+s.Name)
			}
		}
		names = append(names, "&DECK", "&IX1", "&IX2", "&IX3")
		for _, l := range contLines("NGLOBAL ", names, 70, " -") {
			e.line(0, l)
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "ALLOCATE FILE(DECK) DATASET('ACPWB.RSCH.DECK("+e.j.id.Program+")') SHR REUSE")
			e.line(0, "OPENFILE DECK")
			constInits(e, 0)
			for _, s := range e.j.p.Globals {
				if s.Role != "const" && len(s.Dims) == 0 && len(s.Init) > 0 {
					v := s.Init[0]
					if s.Type == "str" {
						v = "&STR('" + v + "')" // kept quoted, as the audit log expects
					}
					e.line(0, fmt.Sprintf("SET &%s = %s", s.Name, v), s.Name)
				}
			}
			for _, s := range e.j.p.Globals {
				if len(s.Dims) == 0 || len(s.Init) == 0 {
					continue
				}
				for i, v := range s.Init {
					if s.Type == "dbl" {
						v = hundredths(v)
					}
					e.line(0, fmt.Sprintf("SET &%s%d = %s", s.Name, i+1, v), s.Name)
				}
			}
			return
		}
		e.raw(p.Name + ": PROC 0")
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "CLOSFILE DECK")
			e.line(0, "FREE FILE(DECK)")
			e.line(0, "EXIT CODE(0)")
			return
		}
		e.raw("END")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "GETFILE DECK") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		return fmt.Sprintf("SET %s = &SUBSTR(%d:%d,&STR(&DECK))", t, start+1, start+width)
	}
	return d
}

// ── AS/400 CL ─────────────────────────────────────────────────────────────

func cl400Dialect(j *job) *dialect {
	modern := j.in.Year >= 2006 // SUBR / CALLSUBR / DOFOR
	cmd := func(c, rest string) string { return fmt.Sprintf("%-10s %s", c, rest) }
	d := &dialect{ind: "  ", cmt: "/* ", cmtEnd: " */", sigil: "&", upper: true, byGlobal: true,
		idx1: "%s(%s)", idx2: "%s(%s %s)", intDiv: "(%s / %s)", mod: "(%[1]s - ((%[1]s / %[2]s) * %[2]s))", dble: "%s", ladder: "(%s * %s)",
		rel: map[string]string{"lt": "*LT", "le": "*LE", "eq": "*EQ", "ne": "*NE", "gt": "*GT", "ge": "*GE"}, and: " *AND ", or: " *OR ",
		assign: cmd("CHGVAR", "VAR(%s) VALUE(%s)"), call: cmd("CALLSUBR", "SUBR(%s)"), callNone: cmd("CALLSUBR", "SUBR(%s)"),
		stop:   cmd("RETURN", ""),
		ifOpen: cmd("IF", "COND(%s) THEN(DO)"), ifElse: cmd("ENDDO", "") + "\n" + cmd("ELSE", "CMD(DO)"), ifClose: cmd("ENDDO", ""),
		forUp: cmd("DOFOR", "VAR(%s) FROM(%s) TO(%s)"), forDown: cmd("DOFOR", "VAR(%s) FROM(%s) TO(%s) BY(-1)"), forClose: cmd("ENDDO", ""),
		concat: " *CAT ", lit: func(s string) string { return sq(cardText(s)) }, order: "main-first",
		pstr: func(x string, w int) string { return x },
		pnum: func(x string, w int) string { return fmt.Sprintf("%%SST(%%CHAR(%s) 1 %d)", x, w) },
		psgn: func(x string, w int) string { return fmt.Sprintf("%%SST(%%CHAR(%s) 1 %d)", x, w) },
		print: func(p string) string {
			return cmd("CHGVAR", "VAR(&LINE) VALUE("+p+")") + "\n" + cmd("CALL", "PGM(RSCHLIB/PRTLIN) PARM(&LINE)")
		},
	}
	if !modern {
		d.call, d.callNone = cmd("CALL", "PGM(RSCHLIB/%s)"), cmd("CALL", "PGM(RSCHLIB/%s)")
	}
	width := func(s Sym) int {
		switch s.Type {
		case "str":
			return maxi(s.Len, 1)
		case "dbl":
			return 15
		}
		return 9
	}
	// Arrays are character variables carved up with %SST.
	elem := func(e *dialectEmitter, x *Expr) string {
		w := width(Sym{Type: e.symType(x.Name), Len: 40})
		var ix string
		if len(x.Args) == 2 {
			ix = fmt.Sprintf("((%s - 1) * %s + %s)", e.ex(x.Args[0]), e.dim1(x.Name), e.ex(x.Args[1]))
		} else {
			ix = e.ex(x.Args[0])
		}
		return fmt.Sprintf("%%SST(&%s ((%s - 1) * %d + 1) %d)", x.Name, ix, w, w)
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		if x.K == "idx" {
			return elem(e, x), true
		}
		return "", false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		typ := "TYPE(*DEC) LEN(9 0)"
		switch s.Type {
		case "str":
			typ = fmt.Sprintf("TYPE(*CHAR) LEN(%d)", maxi(s.Len, 1))
		case "dbl":
			typ = "TYPE(*DEC) LEN(15 5)"
		}
		if len(s.Dims) > 0 {
			n := width(s)
			for _, dd := range s.Dims {
				n *= dd
			}
			typ = fmt.Sprintf("TYPE(*CHAR) LEN(%d)", n)
		}
		val := ""
		if len(s.Init) > 0 && len(s.Dims) == 0 {
			if s.Type == "str" {
				val = " VALUE(" + sq(s.Init[0]) + ")"
			} else {
				val = " VALUE(" + s.Init[0] + ")"
			}
		}
		return []string{cmd("DCL", fmt.Sprintf("VAR(&%s) %s%s", s.Name, typ, val))}
	}
	decls := func(e *dialectEmitter) {
		for _, s := range e.j.p.Globals {
			e.defLine(0, d.decl(e, s, "global")[0], s.Name)
		}
		for _, s := range e.byGlobalSyms() {
			e.defLine(0, d.decl(e, s, "global")[0], s.Name)
		}
		e.defLine(0, cmd("DCL", "VAR(&LINE) TYPE(*CHAR) LEN(132)"), "LINE")
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw(cmd("PGM", "PARM(&DEBUG)"))
		e.line(0, cmd("DCL", "VAR(&DEBUG) TYPE(*CHAR) LEN(1)"))
		e.line(0, cmd("DCLF", "FILE(RSCHLIB/DECK)"))
		decls(e)
		e.line(0, cmd("MONMSG", "MSGID(CPF0000) EXEC(GOTO CMDLBL(ABEND))"))
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			for _, s := range e.j.p.Globals {
				if len(s.Dims) == 0 || len(s.Init) == 0 {
					continue
				}
				for i, v := range s.Init {
					if s.Type == "str" {
						v = sq(v)
					}
					e.line(0, fmt.Sprintf(d.assign, elem(e, Ix(s.Name, N(i+1))), v), s.Name)
				}
			}
			return
		}
		if modern {
			e.raw(cmd("SUBR", "SUBR("+p.Name+")"))
			return
		}
		e.raw("/* ===== SOURCE MEMBER " + p.Name + " (TYPE CLP) ===== */")
		e.raw(cmd("PGM", ""))
		decls(e)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			if modern {
				e.line(0, cmd("RETURN", ""))
				return
			}
			e.line(0, cmd("RETURN", ""))
			e.raw("ABEND:     " + cmd("SNDPGMMSG", "MSGID(CPF9898) MSGF(QCPFMSG) MSGDTA('"+e.j.id.Program+" FAILED') MSGTYPE(*ESCAPE)"))
			e.raw(cmd("ENDPGM", ""))
			return
		}
		if modern {
			e.raw(cmd("ENDSUBR", ""))
			return
		}
		e.raw(cmd("ENDPGM", ""))
	}
	d.footer = func(e *dialectEmitter) {
		if modern {
			e.raw("ABEND:     " + cmd("SNDPGMMSG", "MSGID(CPF9898) MSGF(QCPFMSG) MSGDTA('"+e.j.id.Program+" FAILED') MSGTYPE(*ESCAPE)"))
			e.raw(cmd("ENDPGM", ""))
		}
	}
	if !modern {
		// DOFOR arrived in V5R3: older programs loop with labels and GOTO.
		d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
			if s.K != "do" {
				return false
			}
			n := e.newLabel()
			rel, step := "*GT", "+"
			if s.Step < 0 {
				rel, step = "*LT", "-"
			}
			v := "&" + s.Var
			e.line(depth, fmt.Sprintf(d.assign, v, e.ex(s.From)), s.Var)
			e.line(depth, fmt.Sprintf("LOOP%02d:    ", n)+cmd("IF", fmt.Sprintf("COND(%s %s %s) THEN(GOTO CMDLBL(DONE%02d))", v, rel, e.ex(s.To), n)), s.Var)
			e.body(s.Body, depth+1)
			e.line(depth+1, fmt.Sprintf(d.assign, v, v+" "+step+" 1"))
			e.line(depth+1, cmd("GOTO", fmt.Sprintf("CMDLBL(LOOP%02d)", n)))
			e.line(depth, fmt.Sprintf("DONE%02d:    ", n)+cmd("CHGVAR", "VAR(&LINE) VALUE(&LINE)"))
			return true
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, cmd("RCVF", "RCDFMT(DECKR)"))
		e.line(depth, cmd("MONMSG", "MSGID(CPF0864) EXEC(GOTO CMDLBL(ABEND))"))
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		return fmt.Sprintf(d.assign, t, fmt.Sprintf("%%SST(&CARD %d %d)", start+1, width))
	}
	return d
}

// ── Easytrieve Plus ───────────────────────────────────────────────────────

var deckFields = [][2]int{{0, 40}, {40, 8}, {0, 60}, {0, 3}, {3, 3}, {40, 6}, {46, 6}, {40, 7}}

func eztDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"  ", "   "}), cmt: "* ", upper: true, byGlobal: true,
		idx1: "%s (%s)", idx2: "%s ((%s - 1) * {D1} + %s)", intDiv: "%s / %s", mod: "%[1]s - (%[1]s / %[2]s) * %[2]s", dble: "%s", ladder: "%s * %s",
		rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"}, and: " AND ", or: " OR ",
		assign: "%s = %s", call: "PERFORM %s", callNone: "PERFORM %s", stop: "STOP EXECUTE",
		ifOpen: "IF %s", ifElse: "ELSE", ifClose: "END-IF",
		whileFor: true, whileHd: "DO WHILE %s", whileEnd: "END-DO",
		concat: " ", lit: func(s string) string { return sq(cardText(s)) }, order: "main-first",
		pstr:  func(x string, w int) string { return x },
		pnum:  func(x string, w int) string { return x },
		psgn:  func(x string, w int) string { return x },
		print: func(p string) string { return "DISPLAY " + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		t := "4 B 0"
		switch s.Type {
		case "dbl":
			t = "8 P 4"
		case "str":
			t = fmt.Sprintf("%d A", maxi(s.Len, 1))
		}
		occ := ""
		if len(s.Dims) > 0 {
			n := 1
			for _, dd := range s.Dims {
				n *= dd
			}
			occ = fmt.Sprintf(" OCCURS %d", n)
		}
		val := ""
		if len(s.Init) > 0 && len(s.Dims) == 0 {
			if s.Type == "str" {
				val = " VALUE " + sq(s.Init[0])
			} else {
				val = " VALUE " + s.Init[0]
			}
		}
		return []string{fmt.Sprintf("%-12s W %s%s%s", s.Name, t, occ, val)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("PARM DEBUG (STATE FLOW) LINK(" + e.j.id.Program + ")")
		e.raw("*")
		e.raw("FILE DECK")
		e.raw(fmt.Sprintf("%-12s  1  80 A", "CARD"))
		for _, f := range deckFields {
			e.raw(fmt.Sprintf("C%02d%02d         %3d %3d A", f[0]+1, f[1], f[0]+1, f[1]))
			if f[1] < 10 {
				e.raw(fmt.Sprintf("N%02d%02d         %3d %3d N 0", f[0]+1, f[1], f[0]+1, f[1]))
			}
		}
		e.raw("*")
		e.globalsDecl("global")
		for _, s := range e.byGlobalSyms() {
			e.defLine(0, d.decl(e, s, "global")[0], s.Name)
		}
		e.raw("*")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("JOB INPUT NULL NAME " + e.j.id.Program)
			elementInits(e, 1)
			return
		}
		e.raw(p.Name + ". PROC")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "STOP")
			return
		}
		e.raw("END-PROC")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "GET DECK")
		e.line(depth, "IF EOF DECK\n"+d.ind+"PERFORM DECK-SHORT\nEND-IF")
	}
	d.footer = func(e *dialectEmitter) {
		e.raw("DECK-SHORT. PROC")
		e.line(1, "DISPLAY '*** DECK ENDED EARLY ***'")
		e.line(1, "STOP EXECUTE")
		e.raw("END-PROC")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = N%02d%02d", t, start+1, width)
		}
		return fmt.Sprintf("%s = C%02d%02d", t, start+1, width)
	}
	return d
}

// ── ABAP ──────────────────────────────────────────────────────────────────

func abapDialect(j *job) *dialect {
	modern := j.in.Year >= 2013 // table expressions, string templates
	typ := func(s Sym) string {
		switch s.Type {
		case "dbl":
			return "TYPE f"
		case "str":
			if j.in.Year >= 2002 {
				return fmt.Sprintf("TYPE c LENGTH %d", maxi(s.Len, 1))
			}
			return fmt.Sprintf("\x00(%d) TYPE c", maxi(s.Len, 1))
		}
		return "TYPE i"
	}
	d := &dialect{ind: "  ", cmt: "\" ",
		idx1: "%s[ %s ]", idx2: "%s[ ( %s - 1 ) * {D1} + %s ]", intDiv: "%s DIV %s", mod: "%s MOD %s", dble: "%s", ladder: "trunc( %s * %s )",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "<>", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: "%s = %s.", call: "PERFORM %s USING %s.", callNone: "PERFORM %s.", stop: "LEAVE PROGRAM.",
		ifOpen: "IF %s.", ifElse: "ELSE.", ifClose: "ENDIF.",
		whileFor: true, whileHd: "WHILE %s.", whileEnd: "ENDWHILE.", argSep: " ",
		concat: ", ", lit: func(s string) string { return sq(asciiText(s)) }, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("%s(%d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s(%d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s(%d)", x, w) },
		print: func(p string) string { return "WRITE: / " + p + "." },
	}
	if modern {
		d.dble = "CONV f( %s )"
		tmpl := func(s string) string {
			return strings.NewReplacer("|", "\\|", "{", "\\{", "}", "\\}").Replace(asciiText(s))
		}
		d.concat, d.lit = "", tmpl
		d.pstr = func(x string, w int) string { return fmt.Sprintf("{ %s WIDTH = %d ALIGN = LEFT }", x, w) }
		d.pnum = func(x string, w int) string { return fmt.Sprintf("{ %s WIDTH = %d ALIGN = RIGHT }", x, w) }
		d.psgn = func(x string, w int) string {
			return fmt.Sprintf("{ %s SIGN = LEFTPLUS WIDTH = %d ALIGN = RIGHT }", x, w)
		}
		d.print = func(p string) string { return "WRITE: / |" + p + "|." }
	} else {
		// Internal tables have no element expressions before 7.40: every read
		// is a READ TABLE into a work field, every store a MODIFY.
		d.hoist = "READ TABLE %[2]s INDEX %[3]s INTO %[1]s."
		d.hoistTmp = func(e *dialectEmitter, typ string, n int) string {
			return map[string]string{"int": "gv_t", "dbl": "gv_f", "str": "gv_c"}[typ] + strconv.Itoa(n)
		}
		d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
			if l.K != "idx" {
				return "", false
			}
			h := d.hoist
			d.hoist = ""
			var idx []string
			for _, a := range l.Args {
				idx = append(idx, e.ex(a))
			}
			d.hoist = h
			ix := idx[0]
			if len(idx) == 2 {
				ix = fmt.Sprintf("( %s - 1 ) * %s + %s", idx[0], e.dim1(l.Name), idx[1])
			}
			wa := map[string]string{"int": "gv_wa_i", "dbl": "gv_wa_f", "str": "gv_wa_c"}[e.symType(l.Name)]
			return fmt.Sprintf("%s = %s.\nMODIFY %s FROM %s INDEX %s.", wa, r, l.Name, wa, ix), true
		}
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		if x.K == "dlit" {
			return "'" + x.S + "'", true
		}
		return "", false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("CONSTANTS: %s TYPE i VALUE %s.", s.Name, s.Init[0])}
		}
		if len(s.Dims) > 0 {
			if j.in.Year < 1998 {
				return []string{strings.Replace(fmt.Sprintf("DATA: %s %s OCCURS %d.", s.Name, typ(s), s.Dims[0]), " \x00", "", 1)}
			}
			elem := strings.TrimPrefix(typ(s), "TYPE ")
			if s.Type == "str" {
				elem = fmt.Sprintf("char%d", maxi(s.Len, 1)) // DDIC data element
			}
			return []string{fmt.Sprintf("DATA: %s TYPE STANDARD TABLE OF %s.", s.Name, elem)}
		}
		val := ""
		if len(s.Init) > 0 {
			val = " VALUE " + sq(s.Init[0])
			if s.Type != "str" {
				val = " VALUE " + s.Init[0]
			}
		}
		return []string{strings.Replace(fmt.Sprintf("DATA: %s %s%s.", s.Name, typ(s), val), " \x00", "", 1)}
	}
	d.header = func(e *dialectEmitter) {
		e.raw(fmt.Sprintf("REPORT z%s LINE-SIZE 132 NO STANDARD PAGE HEADING.", strings.ToLower(e.j.id.Program)))
		e.headerComments()
		e.globalsDecl("global")
		e.defLine(0, "DATA: gv_card(80) TYPE c.", "gv_card")
		e.defLine(0, "DATA: gv_file TYPE string VALUE '/usr/sap/RSCH/deck/"+strings.ToLower(e.j.id.Program)+".dat'.", "gv_file")
		if !modern {
			var ts []string
			for i := 1; i <= 12; i++ {
				ts = append(ts, fmt.Sprintf("gv_t%d", i))
			}
			e.line(0, "DATA: "+strings.Join(ts, ", ")+" TYPE i.")
			e.line(0, "DATA: gv_f1, gv_f2, gv_f3, gv_f4 TYPE f.")
			e.line(0, "DATA: gv_c1(60) TYPE c, gv_c2(60) TYPE c.")
			e.line(0, "DATA: gv_wa_i TYPE i, gv_wa_f TYPE f, gv_wa_c(60) TYPE c.")
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("START-OF-SELECTION.")
			for _, l := range p.Locals {
				if l.Role != "param" {
					e.defLine(1, d.decl(e, l, "local")[0], l.Name)
				}
			}
			for _, s := range e.j.p.Globals {
				if len(s.Dims) == 0 {
					continue
				}
				if len(s.Init) > 0 {
					for _, v := range s.Init {
						if s.Type == "str" {
							v = sq(v)
						} else if s.Type == "dbl" {
							v = "'" + v + "'"
						}
						e.line(1, fmt.Sprintf("APPEND %s TO %s.", v, s.Name), s.Name)
					}
					continue
				}
				n := 1
				for _, dd := range s.Dims {
					n *= dd
				}
				e.line(1, fmt.Sprintf("DO %d TIMES.\n%sAPPEND INITIAL LINE TO %s.\nENDDO.", n, d.ind, s.Name), s.Name)
			}
			enc := ""
			if j.in.Year >= 2001 {
				enc = " ENCODING DEFAULT"
			}
			e.line(1, "OPEN DATASET gv_file FOR INPUT IN TEXT MODE"+enc+".")
			e.line(1, "IF sy-subrc <> 0.\n"+d.ind+"MESSAGE e001(zrsch) WITH gv_file.\nENDIF.")
			return
		}
		var ps []string
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			if len(s.Dims) > 0 {
				ps = append(ps, q+" TYPE STANDARD TABLE")
			} else {
				ps = append(ps, q+" "+typ(s))
			}
		}
		if len(ps) > 0 {
			e.raw("FORM " + p.Name + " USING " + strings.Join(ps, " ") + ".")
		} else {
			e.raw("FORM " + p.Name + ".")
		}
		e.locals(p, 1)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "CLOSE DATASET gv_file.")
			return
		}
		e.raw("ENDFORM.")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "READ DATASET gv_file INTO gv_card.") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		return fmt.Sprintf("%s = gv_card+%d(%d).", t, start, width)
	}
	return d
}

// ── RPG ───────────────────────────────────────────────────────────────────
//
// RPG II / RPG/400 are fixed-column calculation specs with no expressions:
// the walker's statements are lowered to one opcode per line (ADD, SUB,
// MULT, DIV/MVR, Z-ADD, MOVEL, IFxx/COMP...), with array subscripts and
// intermediate results in calculation-defined work fields, and printing via
// EXCPT output specs collected at the end of the source. ILE RPG IV uses
// EVAL with free expressions, and the free-form years drop the columns.

type rpgState struct {
	tmp     int
	defined map[string]bool
	ospecs  []string
	excpt   int
	style   string // ii iii iv free
}

func rpgDialect(j *job) *dialect {
	st := &rpgState{defined: map[string]bool{}}
	switch j.id.Language {
	case "RPG II":
		st.style = "ii"
	case "RPG/400":
		st.style = "iii"
	case "RPG IV (free-form)":
		st.style = "free"
	default:
		st.style = "iv"
	}
	if st.style == "ii" || st.style == "iii" {
		return rpgFixedDialect(j, st)
	}
	return rpgIVDialect(j, st)
}

// cspec formats an RPG II/III calculation line.
func cspec(ind, f1, op, f2, res, ln, dec, resInd string) string {
	return strings.TrimRight(fmt.Sprintf("     C%-11s%-10s%-5s%-10s%-6s%3s%1s %-6s", "  "+ind, f1, op, f2, res, ln, dec, resInd), " ")
}

func rpgFixedDialect(j *job, st *rpgState) *dialect {
	d := &dialect{ind: "", cmt: "     C* ", upper: true, byGlobal: true, order: "main-first",
		idx1: "%s,%s", idx2: "%s,%s", intDiv: "", mod: "", dble: "%s", ladder: "",
		rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"}, and: "", or: "",
		assign: "", call: "", callNone: "", stop: cspec("", "", "SETON", "", "", "", "", "LR") + "\n" + cspec("", "", "RETRN", "", "", "", "", ""),
		concat: "\x1e", lit: func(s string) string { return "L\x1f" + strings.ReplaceAll(cardText(s), "'", "''") },
		pstr: func(x string, w int) string { return fmt.Sprintf("S\x1f%s\x1f%d", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("N\x1f%s\x1f%d", x, w) },
		psgn: func(x string, w int) string { return fmt.Sprintf("G\x1f%s\x1f%d", x, w) },
	}
	temp := func(e *dialectEmitter, typ string) (string, string, string) {
		st.tmp++
		n := fmt.Sprintf("WK%03d", st.tmp)
		if typ == "dbl" {
			return n, "15", "5"
		}
		return n, "9", "0"
	}
	var lower func(e *dialectEmitter, x *Expr) string
	emit := func(e *dialectEmitter, f1, op, f2, res, ln, dec string) {
		if res != "" && st.defined[res] {
			ln, dec = "", ""
		} else if res != "" && ln != "" {
			st.defined[res] = true
		}
		var refs []string
		if res != "" {
			refs = []string{res}
		}
		e.pending = append(e.pending, Line{Text: cspec("", f1, op, f2, res, ln, dec, ""), Code: true, Refs: refs})
	}
	opOf := map[string]string{"+": "ADD", "-": "SUB", "*": "MULT", "/": "DIV"}
	lowerInto := func(e *dialectEmitter, x *Expr, dst, ln, dec string) bool {
		switch x.K {
		case "bin":
			a, b := lower(e, x.Args[0]), lower(e, x.Args[1])
			emit(e, a, opOf[x.Op], b, dst, ln, dec)
			return true
		case "ladder":
			a, b := lower(e, x.Args[0]), lower(e, x.Args[1])
			emit(e, a, "MULT", b, dst, ln, dec) // no half-adjust: truncates
			return true
		case "call":
			if x.Name == "mod" {
				a, b := lower(e, x.Args[0]), lower(e, x.Args[1])
				q, ql, qd := temp(e, "int")
				emit(e, a, "DIV", b, q, ql, qd)
				emit(e, "", "MVR", "", dst, ln, dec)
				return true
			}
		}
		return false
	}
	lower = func(e *dialectEmitter, x *Expr) string {
		switch x.K {
		case "var":
			return x.Name
		case "int":
			return strconv.Itoa(x.Val)
		case "dlit":
			return strings.TrimPrefix(x.S, "0")
		case "str":
			return "'" + cardText(x.S) + "'"
		case "idx":
			ix := lower(e, x.Args[0])
			if len(x.Args) == 2 {
				t, l, dd := temp(e, "int")
				k := lower(e, x.Args[0])
				emit(e, k, "SUB", "1", t, l, dd)
				emit(e, t, "MULT", e.dim1(x.Name), t, "", "")
				emit(e, t, "ADD", lower(e, x.Args[1]), t, "", "")
				ix = t
			} else if x.Args[0].K != "var" && x.Args[0].K != "int" {
				t, l, dd := temp(e, "int")
				emit(e, "", "Z-ADD", ix, t, l, dd)
				ix = t
			}
			return trunc(x.Name, 5) + "," + ix
		case "call":
			if x.Name == "dble" {
				return lower(e, x.Args[0])
			}
		}
		t, l, dd := temp(e, e.exprType(x))
		if !lowerInto(e, x, t, l, dd) {
			emit(e, "", "Z-ADD", "0", t, l, dd)
		}
		return t
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) { return lower(e, x), true }
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) { return "", false }
	assign := func(e *dialectEmitter, l, r *Expr) {
		dst := lower(e, l)
		ln, dec := "9", "0"
		if e.exprType(l) == "dbl" {
			ln, dec = "15", "5"
		}
		if strings.Contains(dst, ",") {
			ln, dec = "", ""
		}
		if lowerInto(e, r, dst, ln, dec) {
			e.line(0, "")
			return
		}
		op := "Z-ADD"
		if e.exprType(l) == "str" {
			op = "MOVEL"
		}
		emit(e, "", op, lower(e, r), dst, ln, dec)
		e.line(0, "")
	}
	compare := func(e *dialectEmitter, c *Cond) []string {
		var out []string
		var walk func(c *Cond, join string)
		walk = func(c *Cond, join string) {
			if c.Op == "and" || c.Op == "or" {
				walk(c.A, join)
				walk(c.B, map[string]string{"and": "AND", "or": "OR"}[c.Op])
				return
			}
			a, b := lower(e, c.L), lower(e, c.R)
			out = append(out, a+"\x1f"+join+d.rel[c.Op]+"\x1f"+b)
		}
		walk(c, "IF")
		return out
	}
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		switch s.K {
		case "assign":
			assign(e, s.L, s.R)
			return true
		case "call":
			p := e.procs[s.Name]
			if p != nil {
				for i, a := range s.Args {
					if i < len(p.Params) {
						assign(e, V(p.Params[i]), a)
					}
				}
			}
			e.line(0, cspec("", "", "EXSR", s.Name, "", "", "", ""), s.Name)
			return true
		case "if":
			cs := compare(e, s.Cond)
			if st.style == "ii" {
				// RPG II: COMP sets indicators, GOTO/TAG do the rest.
				n := e.newLabel()
				ind := map[string]string{"LT": "  50  ", "LE": "  5050", "EQ": "    50", "NE": "5050  ", "GT": "50    ", "GE": "50  50"}
				for _, c := range cs {
					f := strings.Split(c, "\x1f")
					rel := f[1][len(f[1])-2:]
					e.pending = append(e.pending, Line{Text: cspec("", f[0], "COMP", f[2], "", "", "", ind[rel]), Code: true})
				}
				e.line(0, cspec("N50", "", "GOTO", fmt.Sprintf("ELS%03d", n), "", "", "", ""))
				e.body(s.Then, 0)
				e.line(0, cspec("", "", "GOTO", fmt.Sprintf("END%03d", n), "", "", "", ""))
				e.line(0, cspec("", fmt.Sprintf("ELS%03d", n), "TAG", "", "", "", "", ""))
				e.body(s.Else, 0)
				e.line(0, cspec("", fmt.Sprintf("END%03d", n), "TAG", "", "", "", "", ""))
				return true
			}
			for _, c := range cs {
				f := strings.Split(c, "\x1f")
				e.pending = append(e.pending, Line{Text: cspec("", f[0], f[1], f[2], "", "", "", ""), Code: true})
			}
			e.line(0, "", condRefs(s.Cond)...)
			e.body(s.Then, 0)
			if s.Else != nil {
				e.line(0, cspec("", "", "ELSE", "", "", "", "", ""))
				e.body(s.Else, 0)
			}
			e.line(0, cspec("", "", "END", "", "", "", "", ""))
			return true
		case "do":
			from, to := lower(e, s.From), lower(e, s.To)
			v := s.Var
			if s.Step > 0 && st.style == "iii" {
				ln, dec := "5", "0"
				if st.defined[v] {
					ln, dec = "", ""
				}
				st.defined[v] = true
				e.line(0, cspec("", from, "DO", to, v, ln, dec, ""), v)
				e.body(s.Body, 0)
				e.line(0, cspec("", "", "END", "", "", "", "", ""))
				return true
			}
			n := e.newLabel()
			emit(e, "", "Z-ADD", from, v, "5", "0")
			e.line(0, cspec("", fmt.Sprintf("TOP%03d", n), "TAG", "", "", "", "", ""), v)
			hi := "51    "
			if s.Step < 0 {
				hi = "  51  "
			}
			e.line(0, cspec("", v, "COMP", to, "", "", "", hi), v)
			e.line(0, cspec(" 51", "", "GOTO", fmt.Sprintf("BOT%03d", n), "", "", "", ""))
			e.body(s.Body, 0)
			op := "ADD"
			if s.Step < 0 {
				op = "SUB"
			}
			e.line(0, cspec("", v, op, "1", v, "", "", ""))
			e.line(0, cspec("", "", "GOTO", fmt.Sprintf("TOP%03d", n), "", "", "", ""))
			e.line(0, cspec("", fmt.Sprintf("BOT%03d", n), "TAG", "", "", "", "", ""))
			return true
		}
		return false
	}
	d.print = func(p string) string {
		st.excpt++
		name := fmt.Sprintf("DTL%03d", st.excpt)
		st.ospecs = append(st.ospecs, fmt.Sprintf("     OQSYSPRT E  1           %s", name))
		end := 0
		for _, part := range strings.Split(p, "\x1e") {
			f := strings.Split(part, "\x1f")
			switch f[0] {
			case "L":
				end += len(f[1])
				st.ospecs = append(st.ospecs, fmt.Sprintf("     O%32d '%s'", end, f[1]))
			default:
				w, _ := strconv.Atoi(f[2])
				end += w
				code := " "
				if f[0] == "N" {
					code = "Z"
				} else if f[0] == "G" {
					code = "L"
				}
				st.ospecs = append(st.ospecs, fmt.Sprintf("     O%23s%-6s%s%3d", "", f[1], code, end))
			}
		}
		return cspec("", "", "EXCPT", name, "", "", "", "")
	}
	arrName := func(n string) string { return trunc(n, 5) }
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		j := e.j
		e.raw(fmt.Sprintf("     H* %s -- %s", j.id.Program, cardText(trunc(j.in.Title, 50))))
		e.headerComments()
		e.raw("     H                                    1")
		e.raw("     FDECK    IP  F  80  80            DISK")
		e.raw("     FQSYSPRT O   F 132 132     OF     PRINTER")
		var ctdata []Sym
		for _, s := range append(append([]Sym{}, j.p.Globals...), e.byGlobalSyms()...) {
			if len(s.Dims) == 0 {
				continue
			}
			n := 1
			for _, dd := range s.Dims {
				n *= dd
			}
			ln, dec := 9, "0"
			switch s.Type {
			case "str":
				ln, dec = maxi(s.Len, 1), " "
			case "dbl":
				ln, dec = 3, "2"
			}
			per := ""
			if len(s.Init) > 0 {
				per = fmt.Sprintf("%3d", len(s.Init))
				ctdata = append(ctdata, s)
			}
			e.defLine(0, strings.TrimRight(fmt.Sprintf("     E%17s%-6s%3s%4d%3d %s", "", arrName(s.Name), per, n, ln, dec), " "), s.Name)
			st.defined[arrName(s.Name)] = true
		}
		e.raw("     IDECK    NS  01")
		e.raw("     I                                        1  80 CARD")
		e.raw("     ICARDDS      DS")
		for _, f := range deckFields {
			e.raw(fmt.Sprintf("     I%38d%4d C%02d%02d", f[0]+1, f[0]+f[1], f[0]+1, f[1]))
			if f[1] < 10 {
				e.raw(fmt.Sprintf("     I%38d%4d0N%02d%02d", f[0]+1, f[0]+f[1], f[0]+1, f[1]))
			}
		}
		st.ospecs = append(st.ospecs, "     OQSYSPRT H  201   1P", "     O                                   9 'RECORD '")
		e.pending = nil
		// compile-time array data goes after the output specs
		d.footer = func(e *dialectEmitter) {
			e.commentLines(0, []string{"OUTPUT SPECIFICATIONS -- ONE EXCPT RECORD PER REPORT LINE"})
			for _, o := range st.ospecs {
				e.raw(o)
			}
			for _, s := range ctdata {
				e.raw("**")
				var b strings.Builder
				for _, v := range s.Init {
					if s.Type == "dbl" {
						b.WriteString(fmt.Sprintf("%03s", hundredths(v)))
					} else if s.Type == "str" {
						b.WriteString(padRight(cardText(v), maxi(s.Len, 1)))
					} else {
						b.WriteString(fmt.Sprintf("%09s", v))
					}
				}
				e.raw(b.String())
			}
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			for _, s := range e.j.p.Globals {
				if s.Role == "const" {
					emit(e, "", "Z-ADD", s.Init[0], s.Name, "5", "0")
					e.line(0, "")
				} else if s.Type == "str" && len(s.Init) > 0 && len(s.Dims) == 0 {
					emit(e, "", "MOVEL", "'"+cardText(s.Init[0])+"'", s.Name, strconv.Itoa(maxi(s.Len, 1)), "")
					e.line(0, "")
				}
			}
			return
		}
		e.raw(strings.TrimRight(fmt.Sprintf("     CSR%9s%-10s%-5s", "", p.Name, "BEGSR"), " "))
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, cspec("", "", "SETON", "", "", "", "", "LR"))
			return
		}
		e.raw("     CSR                   ENDSR")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(0, cspec("", "", "READ", "DECK", "", "", "", "    LR"))
		e.line(0, cspec("", "", "MOVEL", "CARD", "CARDDS", "", "", ""))
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return cspec("", "", "Z-ADD", fmt.Sprintf("N%02d%02d", start+1, width), t, "", "", "")
		}
		return cspec("", "", "MOVEL", fmt.Sprintf("C%02d%02d", start+1, width), t, "", "", "")
	}
	return d
}

func rpgIVDialect(j *job, st *rpgState) *dialect {
	free := st.style == "free"
	forOK := j.in.Year >= 1999
	c := func(op, ext string) string {
		if free {
			if ext == "" {
				return strings.ToLower(op) + ";"
			}
			return strings.ToLower(op) + " " + ext + ";"
		}
		return strings.TrimRight(fmt.Sprintf("     C%19s%-10s%s", "", op, ext), " ")
	}
	d := &dialect{ind: "", cmt: "     C* ", byGlobal: true, order: "main-first",
		idx1: "%s(%s)", idx2: "%s((%s - 1) * {D1} + %s)", intDiv: "%%DIV(%s:%s)", mod: "%%REM(%s:%s)", dble: "%s", ladder: "%%INT(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "<>", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: c("EVAL", "%s = %s"), call: c("EXSR", "%s"), callNone: c("EXSR", "%s"), stop: c("RETURN", ""),
		ifOpen: c("IF", "%s"), ifElse: c("ELSE", ""), ifClose: c("ENDIF", ""),
		forUp: c("FOR", "%s = %s TO %s"), forDown: c("FOR", "%s = %s DOWNTO %s"), forClose: c("ENDFOR", ""),
		concat: " + ", lit: func(s string) string { return sq(asciiText(s)) },
		pstr: func(x string, w int) string { return fmt.Sprintf("%%SUBST(%s:1:%d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("%%EDITC(%s:'3')", x) },
		psgn: func(x string, w int) string { return fmt.Sprintf("%%EDITC(%s:'L')", x) },
		print: func(p string) string {
			return c("EVAL", "PRTDS = "+p) + "\n" + c("WRITE", "QSYSPRT PRTDS")
		},
	}
	if j.in.Year < 1999 {
		d.intDiv, d.mod, d.ladder = "%%INT(%s / %s)", "(%[1]s - %%INT(%[1]s / %[2]s) * %[2]s)", "%%INT(%s * %s)"
	}
	if free {
		d.ind, d.cmt = "  ", "// "
		d.assign = "%s = %s;"
		d.lit = func(s string) string { return sq(asciiText(s)) }
	}
	if !forOK {
		// FOR arrived with V4R4; before that, DO with an index field.
		d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
			if s.K != "do" {
				return false
			}
			from, to := e.ex(s.From), e.ex(s.To)
			if s.Step > 0 {
				e.line(depth, strings.TrimRight(fmt.Sprintf("     C     %-14s%-10s%-14s%s", from, "DO", to, s.Var), " "), s.Var)
				e.body(s.Body, depth)
				e.line(depth, c("ENDDO", ""))
				return true
			}
			e.line(depth, fmt.Sprintf(d.assign, s.Var, from), s.Var)
			e.line(depth, c("DOW", s.Var+" >= "+to), s.Var)
			e.body(s.Body, depth)
			e.line(depth, fmt.Sprintf(d.assign, s.Var, s.Var+" - 1"))
			e.line(depth, c("ENDDO", ""))
			return true
		}
	}
	dtype := func(s Sym) string {
		switch s.Type {
		case "dbl":
			if free {
				return "packed(15:5)"
			}
			return "15P 5"
		case "str":
			if free {
				return fmt.Sprintf("char(%d)", maxi(s.Len, 1))
			}
			return fmt.Sprintf("%dA", maxi(s.Len, 1))
		}
		if free {
			return "packed(9:0)"
		}
		return "9P 0"
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		kw := ""
		if len(s.Dims) > 0 {
			n := 1
			for _, dd := range s.Dims {
				n *= dd
			}
			kw = fmt.Sprintf(" DIM(%d)", n)
			if len(s.Init) > 0 {
				kw += fmt.Sprintf(" CTDATA PERRCD(%d)", len(s.Init))
			}
		}
		if s.Role == "const" {
			if free {
				return []string{fmt.Sprintf("dcl-c %s %s;", s.Name, s.Init[0])}
			}
			return []string{fmt.Sprintf("     D %-15s C                   CONST(%s)", s.Name, s.Init[0])}
		}
		if len(s.Init) > 0 && len(s.Dims) == 0 {
			v := s.Init[0]
			if s.Type == "str" {
				v = sq(v)
			}
			kw += " INZ(" + v + ")"
		}
		if free {
			return []string{fmt.Sprintf("dcl-s %s %s%s;", s.Name, dtype(s), strings.ToLower(kw))}
		}
		return []string{strings.TrimRight(fmt.Sprintf("     D %-15s S%17s%s", s.Name, dtype(s), kw), " ")}
	}
	var ctdata []Sym
	d.header = func(e *dialectEmitter) {
		j := e.j
		if free {
			e.raw("**FREE")
			e.headerComments()
			e.raw("ctl-opt dftactgrp(*no) actgrp(*new) option(*srcstmt);")
			e.raw("dcl-f deck disk(80) usage(*input);")
			e.raw("dcl-f qsysprt printer(132);")
			e.raw("dcl-ds prtds len(132) end-ds;")
			e.raw("dcl-ds cardds len(80);")
			for _, f := range deckFields {
				e.raw(fmt.Sprintf("  c%02d%02d char(%d) pos(%d);", f[0]+1, f[1], f[1], f[0]+1))
				if f[1] < 10 {
					e.raw(fmt.Sprintf("  n%02d%02d zoned(%d:0) pos(%d);", f[0]+1, f[1], f[1], f[0]+1))
				}
			}
			e.raw("end-ds;")
		} else {
			e.raw("     H DFTACTGRP(*NO) ACTGRP(*NEW) OPTION(*SRCSTMT)")
			e.headerComments()
			e.raw("     FDECK      IF   F   80        DISK")
			e.raw("     FQSYSPRT   O    F  132        PRINTER")
			e.raw("     D PRTDS           DS           132")
			e.raw("     D CARDDS          DS            80")
			for _, f := range deckFields {
				e.raw(fmt.Sprintf("     D  C%02d%02d%18d%7d", f[0]+1, f[1], f[0]+1, f[0]+f[1]))
				if f[1] < 10 {
					e.raw(fmt.Sprintf("     D  N%02d%02d%18d%7dS 0", f[0]+1, f[1], f[0]+1, f[0]+f[1]))
				}
			}
		}
		for _, s := range append(append([]Sym{}, j.p.Globals...), e.byGlobalSyms()...) {
			for _, t := range d.decl(e, s, "global") {
				e.defLine(0, t, s.Name)
			}
			if len(s.Init) > 0 && len(s.Dims) > 0 {
				ctdata = append(ctdata, s)
			}
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			return
		}
		if free {
			e.raw("begsr " + p.Name + ";")
			return
		}
		e.raw(strings.TrimRight(fmt.Sprintf("     CSR   %-14s%s", p.Name, "BEGSR"), " "))
	}
	d.bodyDepth = func(isMain bool) int {
		if free && !isMain {
			return 1
		}
		return 0
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, c("EVAL", "*INLR = *ON"))
			if free {
				e.out[len(e.out)-1].Text = "*inlr = *on;"
			}
			e.line(0, c("RETURN", ""))
			return
		}
		if free {
			e.raw("endsr;")
			return
		}
		e.raw("     CSR                 ENDSR")
	}
	d.footer = func(e *dialectEmitter) {
		for _, s := range ctdata {
			if free {
				e.raw("// compile-time data for " + s.Name + " is loaded by *INZSR in the free-form years")
				continue
			}
			e.raw("**CTDATA " + s.Name)
			var b strings.Builder
			for _, v := range s.Init {
				if s.Type == "str" {
					b.WriteString(padRight(cardText(v), maxi(s.Len, 1)))
				} else if s.Type == "dbl" {
					b.WriteString(fmt.Sprintf("%015s", strings.ReplaceAll(v, ".", "")+"000"))
				} else {
					b.WriteString(fmt.Sprintf("%09s", v))
				}
			}
			e.raw(b.String())
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		if free {
			e.line(depth, "read deck cardds;")
			return
		}
		e.line(depth, strings.TrimRight(fmt.Sprintf("     C%19s%-10s%-14s%-14s%5s", "", "READ", "DECK", "CARDDS", "LR"), " "))
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		n := fmt.Sprintf("C%02d%02d", start+1, width)
		if num {
			n = fmt.Sprintf("N%02d%02d", start+1, width)
		}
		if free {
			n = strings.ToLower(n)
		}
		return fmt.Sprintf(d.assign, t, n)
	}
	return d
}
