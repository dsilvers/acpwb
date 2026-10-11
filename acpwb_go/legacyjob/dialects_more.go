package legacyjob

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// More walker dialects: statistics (R / S-PLUS, MATLAB / Octave, APL),
// Unix tools (awk, ksh, Tcl, bc, dc), web (ColdFusion, Classic ASP) and the
// research languages (Modula-2, Lisp, Smalltalk, Forth).

// closeLast appends text (closing brackets) to the last emitted code line.
func (e *dialectEmitter) closeLast(text string) {
	for i := len(e.out) - 1; i >= 0; i-- {
		if e.out[i].Code {
			e.out[i].Text += text
			return
		}
	}
	e.raw(text)
}

// ── R / S-PLUS ────────────────────────────────────────────────────────────

func rDialect(j *job) *dialect {
	x := j.x
	splus := j.id.LangKey == "splus"
	arrow := "<-"
	if splus && x.chance(0.4) {
		arrow = "_" // S's original assignment operator
	}
	inMain := true
	d := &dialect{ind: pick(x, []string{"  ", "    "}), cmt: pick(x, []string{"# ", "## "}),
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "%s %%/%% %s", mod: "%s %%%% %s", dble: "as.double(%s)", ladder: "trunc(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ",
		assign: "%s " + arrow + " %s", call: "%s(%s)", callNone: "%s()", stop: `stop("deck rejected")`,
		ifOpen: "if (%s) {", ifElse: "} else {", ifClose: "}",
		forUp: "for (%s in %s:%s) {", forDown: "for (%s in %s:%s) {", forClose: "}",
		concat: ", ", lit: dq, order: "procs-first",
		pstr: func(x string, w int) string { return fmt.Sprintf("formatC(%s, width = -%d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("formatC(%s, width = %d, format = \"d\")", x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("formatC(%s, width = %d, format = \"d\", flag = \"+\")", x, w)
		},
		print: func(p string) string { return "cat(" + p + ", \"\\n\", sep = \"\")" },
	}
	global := func(e *dialectEmitter, t string) bool {
		base := strings.SplitN(t, "[", 2)[0]
		return !inMain && e.globals[base]
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		ls := e.ex(l)
		if global(e, ls) {
			return ls + " <<- " + r, true
		}
		return ls + " " + arrow + " " + r, true
	}
	vec := x.chance(0.5)
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		tgt, outer, med, mult, ok := vecLadder(s)
		if !vec || !ok {
			return false
		}
		e.line(depth, fmt.Sprintf("%s[, %s] %s trunc(%s * %s)", tgt, e.ex(outer), map[bool]string{true: "<<-", false: arrow}[!inMain && e.globals[tgt]], e.ex(med), mult), tgt, mult)
		return true
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%s %s %s", e.constName(s.Name), arrow, s.Init[0])}
		}
		zero := map[string]string{"int": "0L", "dbl": "0", "str": "\"\""}[s.Type]
		if zero == "" {
			zero = "0L"
		}
		switch {
		case len(s.Dims) == 2:
			return []string{fmt.Sprintf("%s %s matrix(%s, nrow = %d, ncol = %d)", s.Name, arrow, zero, s.Dims[0], s.Dims[1])}
		case len(s.Dims) == 1 && len(s.Init) > 0:
			return []string{fmt.Sprintf("%s %s c(%s)", s.Name, arrow, initList(s, dq))}
		case len(s.Dims) == 1:
			return []string{fmt.Sprintf("%s %s %s(%d)", s.Name, arrow, map[string]string{"int": "integer", "dbl": "numeric", "str": "character"}[s.Type], s.Dims[0])}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("%s %s %s", s.Name, arrow, dq(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("%s %s %s", s.Name, arrow, s.Init[0])}
		}
		return []string{fmt.Sprintf("%s %s %s", s.Name, arrow, zero)}
	}
	d.header = func(e *dialectEmitter) {
		if !splus {
			e.raw("#!/usr/bin/env Rscript")
		}
		e.headerComments()
		e.line(0, "options(warn = 1, stringsAsFactors = FALSE)")
		e.globalsDecl("global")
		e.line(0, "deck "+arrow+" readLines(\"deck.dat\")")
		e.line(0, "cardno "+arrow+" 0L")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		inMain = false
		if isMain {
			e.raw("main " + arrow + " function() {")
			e.locals(p, 1)
			return
		}
		e.raw(fmt.Sprintf("%s %s function(%s) {", p.Name, arrow, strings.Join(p.Params, ", ")))
		e.locals(p, 1)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "invisible(NULL)")
		}
		e.raw("}")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw("main()")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "cardno <<- cardno + 1L\ncard "+arrow+" deck[cardno]")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		op := arrow
		if global(e, t) {
			op = "<<-"
		}
		if num {
			return fmt.Sprintf("%s %s as.integer(substr(card, %d, %d))", t, op, start+1, start+width)
		}
		return fmt.Sprintf("%s %s substr(card, %d, %d)", t, op, start+1, start+width)
	}
	return d
}

// ── MATLAB / GNU Octave ───────────────────────────────────────────────────

func matlabDialect(j *job) *dialect {
	x := j.x
	octave := j.id.LangKey == "octave"
	withEnd := j.in.Year >= 2005 || octave
	and, or := " && ", " || "
	if j.in.Year < 2002 && !octave {
		and, or = " & ", " | "
	}
	cmt := "% "
	if octave && x.chance(0.5) {
		cmt = "## "
	}
	d := &dialect{ind: pick(x, []string{"  ", "    "}), cmt: cmt,
		idx1: "%s(%s)", idx2: "%s(%s, %s)", intDiv: "fix(%s / %s)", mod: "mod(%s, %s)", dble: "double(%s)", ladder: "fix(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "==", "ne": "~=", "gt": ">", "ge": ">="}, and: and, or: or,
		assign: "%s = %s;", call: "%s(%s);", callNone: "%s;", stop: "error('" + strings.ToUpper(j.id.Program) + ":abend', 'deck rejected');",
		ifOpen: "if %s", ifElse: "else", ifClose: "end",
		forUp: "for %s = %s:%s", forDown: "for %s = %s:-1:%s", forClose: "end",
		concat: ", ", lit: sq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("sprintf('%%-%ds', %s)", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("sprintf('%%%dd', %s)", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("sprintf('%%+%dd', %s)", w, x) },
		print: func(p string) string { return "fprintf('%s\\n', [" + p + "]);" },
	}
	if octave {
		d.order = "procs-first"
	}
	vec := x.chance(0.5)
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		tgt, outer, med, mult, ok := vecLadder(s)
		if !vec || !ok {
			return false
		}
		e.line(depth, fmt.Sprintf("%s(:, %s) = fix(%s * %s(:));", tgt, e.ex(outer), e.ex(med), mult), tgt, mult)
		return true
	}
	// character data lives in cell arrays
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		if x.K == "idx" && e.symType(x.Name) == "str" {
			return fmt.Sprintf("%s{%s}", e.name(x.Name), e.ex(x.Args[0])), true
		}
		return "", false
	}
	globals := func(e *dialectEmitter) string {
		names := []string{"fid", "card"}
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				names = append(names, e.constName(s.Name))
			} else {
				names = append(names, s.Name)
			}
		}
		return "global " + strings.Join(names, " ")
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%s = %s;", e.constName(s.Name), s.Init[0])}
		}
		switch {
		case s.Type == "str" && len(s.Dims) > 0:
			return []string{fmt.Sprintf("%s = cell(%d, 1);", s.Name, s.Dims[0])}
		case len(s.Dims) == 2:
			return []string{fmt.Sprintf("%s = zeros(%d, %d);", s.Name, s.Dims[0], s.Dims[1])}
		case len(s.Dims) == 1 && len(s.Init) > 0:
			return []string{fmt.Sprintf("%s = [%s];", s.Name, strings.Join(s.Init, " "))}
		case len(s.Dims) == 1:
			return []string{fmt.Sprintf("%s = zeros(%d, 1);", s.Name, s.Dims[0])}
		case s.Type == "str" && len(s.Init) > 0:
			return []string{fmt.Sprintf("%s = %s;", s.Name, sq(s.Init[0]))}
		case s.Type == "str":
			return []string{fmt.Sprintf("%s = '';", s.Name)}
		}
		return []string{fmt.Sprintf("%s = 0;", s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		if octave {
			e.raw("1;  " + cmt + "script file, not a function file")
			e.headerComments()
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("function " + strings.ToLower(e.j.id.Program) + "()")
			if !octave {
				e.headerComments()
			}
			e.line(1, globals(e))
			e.globalsDeclAt(1)
			e.locals(p, 1)
			e.line(1, "fid = fopen('deck.dat', 'r');")
			e.line(1, "if fid < 0, error('cannot open deck.dat'); end")
			return
		}
		e.raw(fmt.Sprintf("function %s(%s)", p.Name, strings.Join(p.Params, ", ")))
		e.line(1, globals(e))
		e.locals(p, 1)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "fclose(fid);")
		}
		if withEnd {
			e.raw("end")
		}
	}
	d.footer = func(e *dialectEmitter) {
		if octave {
			e.blank()
			e.raw(strings.ToLower(e.j.id.Program) + "();")
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "card = fgetl(fid);") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = str2double(card(%d:%d));", t, start+1, start+width)
		}
		return fmt.Sprintf("%s = card(%d:%d);", t, start+1, start+width)
	}
	return d
}

// globalsDeclAt declares globals at an indent depth (MATLAB keeps them
// inside the main function).
func (e *dialectEmitter) globalsDeclAt(depth int) {
	for _, s := range e.j.p.Globals {
		for _, t := range e.d.decl(e, s, "global") {
			e.defLine(depth, t, s.Name)
		}
	}
}

// ── APL ───────────────────────────────────────────────────────────────────

func aplDialect(j *job) *dialect {
	start := 0
	sym := map[string]string{"+": "+", "-": "-", "*": "×", "/": "÷"}
	d := &dialect{ind: " ", cmt: "⍝ ", upper: true, byGlobal: true, order: "main-first",
		rel: map[string]string{"lt": "<", "le": "≤", "eq": "=", "ne": "≠", "gt": ">", "ge": "≥"}, and: "∧", or: "∨",
		assign: "%s←%s", callNone: "%s", call: "%s", stop: "→0",
		concat: ",", lit: func(s string) string { return sq(cardText(s)) },
		pstr:  func(x string, w int) string { return fmt.Sprintf("(%d↑%s)", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("(%d 0⍕%s)", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("(¯%d↑'+',⍕%s)", w, x) },
		print: func(p string) string { return "⎕←" + p },
	}
	num := func(v int) string {
		if v < 0 {
			return "¯" + strconv.Itoa(-v)
		}
		return strconv.Itoa(v)
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "int":
			return num(x.Val), true
		case "dlit":
			return strings.TrimPrefix(x.S, "0"), true
		case "idx":
			var ix []string
			for _, a := range x.Args {
				ix = append(ix, e.ex(a))
			}
			if e.symType(x.Name) == "str" {
				return x.Name + "[" + ix[0] + ";]", true
			}
			return x.Name + "[" + strings.Join(ix, ";") + "]", true
		case "bin":
			l, r := e.ex(x.Args[0]), e.ex(x.Args[1])
			if x.Op == "/" && !isDouble(x.Args[0]) && !isDouble(x.Args[1]) {
				return "(⌊" + l + "÷" + r + ")", true
			}
			return "(" + l + sym[x.Op] + r + ")", true
		case "call":
			if x.Name == "mod" {
				return "(" + e.ex(x.Args[1]) + "|" + e.ex(x.Args[0]) + ")", true
			}
			if x.Name == "dble" {
				return e.ex(x.Args[0]), true
			}
		case "ladder":
			return "(⌊" + e.ex(x.Args[0]) + "×" + e.ex(x.Args[1]) + ")", true
		}
		return "", false
	}
	d.condFn = func(e *dialectEmitter, c *Cond) (string, bool) {
		if c.Op == "and" || c.Op == "or" {
			return "(" + e.cond(c.A) + ")" + map[string]string{"and": "∧", "or": "∨"}[c.Op] + "(" + e.cond(c.B) + ")", true
		}
		return e.ex(c.L) + d.rel[c.Op] + e.ex(c.R), true
	}
	label := func(e *dialectEmitter) string { return fmt.Sprintf("L%d", e.newLabel()) }
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		if tgt, outer, med, mult, ok := vecLadder(s); ok {
			// the one place the FORTRAN shop let APL be APL
			e.line(0, fmt.Sprintf("%s[;%s]←⌊%s×%s", tgt, e.ex(outer), e.ex(med), mult), tgt, mult)
			return true
		}
		switch s.K {
		case "if":
			els, end := label(e), label(e)
			e.line(0, "→("+"~"+e.cond(s.Cond)+")/"+els, condRefs(s.Cond)...)
			e.body(s.Then, 0)
			if s.Else != nil {
				e.line(0, "→"+end)
			}
			e.line(0, els+":⍝")
			if s.Else != nil {
				e.body(s.Else, 0)
				e.line(0, end+":⍝")
			}
			return true
		case "do":
			top, done := label(e), label(e)
			rel, step := ">", "+"
			if s.Step < 0 {
				rel, step = "<", "-"
			}
			e.line(0, s.Var+"←"+e.ex(s.From), s.Var)
			e.line(0, top+":→("+s.Var+rel+e.ex(s.To)+")/"+done, s.Var)
			e.body(s.Body, 0)
			e.line(0, s.Var+"←"+s.Var+step+"1")
			e.line(0, "→"+top)
			e.line(0, done+":⍝")
			return true
		}
		return false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.raw(")CLEAR")
		e.raw(fmt.Sprintf(")WSID %s", e.j.id.Program))
		e.raw(")COPY 1 UTILITY FMT")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		locals := []string{}
		for _, l := range p.Locals {
			locals = append(locals, l.Name)
		}
		head := "∇ " + p.Name
		if isMain {
			head += ";⎕IO"
		}
		if len(locals) > 0 {
			head += ";" + strings.Join(locals, ";")
		}
		e.raw(head)
		start = len(e.out)
		if isMain {
			e.headerComments()
			e.line(0, "⎕IO←1")
			for _, s := range e.j.p.Globals {
				switch {
				case s.Role == "const":
					e.line(0, s.Name+"←"+s.Init[0], s.Name)
				case len(s.Init) > 0 && len(s.Dims) > 0:
					e.line(0, s.Name+"←"+strings.ReplaceAll(strings.Join(s.Init, " "), "0.", "."), s.Name)
				case len(s.Init) > 0:
					e.line(0, s.Name+"←"+sq(s.Init[0]), s.Name)
				case s.Type == "str" && len(s.Dims) > 0:
					e.line(0, fmt.Sprintf("%s←%d %d⍴' '", s.Name, s.Dims[0], maxi(s.Len, 1)), s.Name)
				case s.Type == "str":
					e.line(0, fmt.Sprintf("%s←%d⍴' '", s.Name, maxi(s.Len, 1)), s.Name)
				case len(s.Dims) > 0:
					var ds []string
					for _, dd := range s.Dims {
						ds = append(ds, strconv.Itoa(dd))
					}
					e.line(0, s.Name+"←"+strings.Join(ds, " ")+"⍴0", s.Name)
				default:
					e.line(0, s.Name+"←0", s.Name)
				}
			}
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		n := 0
		for i := start; i < len(e.out); i++ {
			n++
			t := strings.TrimLeft(e.out[i].Text, " ")
			if t == "" {
				t = "⍝"
			}
			e.out[i].Text = fmt.Sprintf("[%d] %s", n, t)
		}
		e.raw("∇")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(0, "CARD←80↑⍞") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s←⍎CARD[%d+⍳%d]", t, start, width)
		}
		return fmt.Sprintf("%s←CARD[%d+⍳%d]", t, start, width)
	}
	return d
}

// ── awk ───────────────────────────────────────────────────────────────────

func awkDialect(j *job) *dialect {
	x := j.x
	gawk := j.id.Language == "gawk"
	d := &dialect{ind: pick(x, []string{"    ", "  ", "\t"}), cmt: "# ",
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "int(%s / %s)", mod: "%s %% %s", dble: "%s", ladder: "int(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ",
		assign: "%s = %s", call: "%s(%s)", callNone: "%s()", stop: "exit 16",
		ifOpen: "if (%s) {", ifElse: "} else {", ifClose: "}",
		forUp: "for (%[1]s = %[2]s; %[1]s <= %[3]s; %[1]s++) {", forDown: "for (%[1]s = %[2]s; %[1]s >= %[3]s; %[1]s--) {", forClose: "}",
		concat: " ", lit: dq, order: pick(x, []string{"procs-first", "main-first"}),
		pstr:  func(x string, w int) string { return fmt.Sprintf("sprintf(\"%%-%ds\", %s)", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("sprintf(\"%%%dd\", %s)", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("sprintf(\"%%+%dd\", %s)", w, x) },
		print: func(p string) string { return "print " + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		if gawk {
			e.raw("#!/usr/bin/gawk -f")
		} else {
			e.raw("#!/usr/bin/nawk -f")
		}
		e.headerComments()
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("BEGIN {")
			e.line(1, `DECK = (ARGV[1] != "") ? ARGV[1] : "deck.dat"`)
			e.line(1, `ARGV[1] = ""`)
			constInits(e, 1)
			scalarInits(e, 1)
			elementInits(e, 1)
			return
		}
		// awk has no local variables: extra parameters after a gap are the idiom.
		var locals []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				locals = append(locals, l.Name)
			}
		}
		head := "function " + p.Name + "(" + strings.Join(p.Params, ", ")
		if len(locals) > 0 {
			if len(p.Params) > 0 {
				head += ","
			}
			head += "    " + strings.Join(locals, ", ")
		}
		e.raw(head + ") {")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "close(DECK)")
			e.line(1, "exit 0")
		}
		e.raw("}")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, `if ((getline card < DECK) <= 0) { print "deck short" > "/dev/stderr"; exit 2 }`)
	}
	style := x.intn(3)
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		switch {
		case num && style == 1:
			return fmt.Sprintf("%s = int(substr(card, %d, %d))", t, start+1, width)
		case num && style == 2:
			return fmt.Sprintf("%s = substr(card, %d, %d) * 1", t, start+1, width)
		case num:
			return fmt.Sprintf("%s = substr(card, %d, %d) + 0", t, start+1, width)
		case style == 2:
			return fmt.Sprintf("%s = sprintf(\"%%-%d.%ds\", substr(card, %d))", t, width, width, start+1)
		}
		return fmt.Sprintf("%s = substr(card, %d, %d)", t, start+1, width)
	}
	return d
}

// ── Korn shell ────────────────────────────────────────────────────────────

func kshDialect(j *job) *dialect {
	x := j.x
	ksh93 := j.in.Year >= 1995
	d := &dialect{ind: pick(x, []string{"    ", "  ", "\t"}), cmt: "# ",
		idx1: "%s[%s]", idx2: "%s[(%s - 1) * {D1} + %s]", intDiv: "%s / %s", mod: "%s %% %s", dble: "%s", ladder: "int(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ",
		assign: "(( %s = %s ))", stop: "exit 16",
		ifOpen: "if (( %s )); then", ifElse: "else", ifClose: "fi",
		forUp: "for (( %[1]s = %[2]s; %[1]s <= %[3]s; %[1]s++ )); do", forDown: "for (( %[1]s = %[2]s; %[1]s >= %[3]s; %[1]s-- )); do", forClose: "done",
		concat: "", order: "procs-first",
		lit: func(s string) string {
			return strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "$", "\\$", "`", "\\`").Replace(asciiText(s))
		},
		pstr:  func(x string, w int) string { return fmt.Sprintf("$(printf '%%-%ds' \"%s\")", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("$(printf '%%%dd' $(( %s )))", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("$(printf '%%+%dd' $(( %s )))", w, x) },
		print: func(p string) string { return "print -r -- \"" + p + "\"" },
	}
	if !ksh93 {
		d.whileFor, d.whileHd, d.whileEnd = true, "while (( %s )); do", "done"
		d.ladder = "(%s * %s) / 100"
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "dlit":
			if !ksh93 {
				return hundredths(x.S), true
			}
		case "var", "idx":
			if e.symType(x.Name) == "str" {
				if x.K == "var" {
					return "${" + x.Name + "}", true
				}
				return "${" + x.Name + "[" + e.ex(x.Args[0]) + "]}", true
			}
		}
		return "", false
	}
	unbrace := func(t string) string {
		if strings.HasPrefix(t, "${") {
			return t[2 : len(t)-1]
		}
		return t
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		if e.symType(l.Name) == "str" {
			return unbrace(e.ex(l)) + "=\"" + r + "\"", true
		}
		return "", false
	}
	d.callFn = func(e *dialectEmitter, name string, args []string) string {
		var as []string
		for _, a := range args {
			if strings.HasPrefix(a, "${") {
				as = append(as, "\""+a+"\"")
			} else {
				as = append(as, "$(( "+a+" ))")
			}
		}
		return strings.TrimSpace(name + " " + strings.Join(as, " "))
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		kw := "typeset -i"
		if s.Type == "str" {
			kw = "typeset"
		} else if s.Type == "dbl" {
			kw = "typeset -F"
			if !ksh93 {
				kw = "typeset -i"
			}
		}
		if s.Role == "const" {
			return []string{fmt.Sprintf("typeset -ri %s=%s", e.constName(s.Name), s.Init[0])}
		}
		if len(s.Dims) > 0 {
			if len(s.Init) > 0 {
				vals := s.Init
				if s.Type == "dbl" && !ksh93 {
					vals = nil
					for _, v := range s.Init {
						vals = append(vals, hundredths(v))
					}
				}
				if ksh93 {
					return []string{fmt.Sprintf("%s -a %s=(0 %s)", kw, s.Name, strings.Join(vals, " "))}
				}
				return []string{fmt.Sprintf("set -A %s 0 %s", s.Name, strings.Join(vals, " "))}
			}
			return []string{fmt.Sprintf("%s %s", kw, s.Name)}
		}
		if len(s.Init) > 0 {
			return []string{fmt.Sprintf("%s %s=%q", kw, s.Name, s.Init[0])}
		}
		if s.Type == "str" {
			return []string{fmt.Sprintf("%s %s=\"\"", kw, s.Name)}
		}
		return []string{fmt.Sprintf("%s %s=0", kw, s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		e.raw("#!/bin/ksh")
		e.headerComments()
		e.line(0, "set -o nounset")
		e.globalsDecl("global")
		e.defLine(0, "typeset card=\"\"", "card")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.blank()
			e.line(0, `exec 3< "${DECK:-deck.dat}"`)
			for _, l := range p.Locals {
				if l.Role != "param" {
					e.defLine(0, d.decl(e, l, "local")[0], l.Name)
				}
			}
			return
		}
		e.raw("function " + p.Name + " {")
		for i, q := range p.Params {
			e.defLine(1, fmt.Sprintf("typeset -i %s=$%d", q, i+1), q)
		}
		e.locals(p, 1)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "exec 3<&-")
			e.line(0, "exit 0")
			return
		}
		e.raw("}")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, `read -r -u3 card || { print -u2 "deck short"; exit 2; }`)
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		t = unbrace(t)
		sub := fmt.Sprintf("${card:%d:%d}", start, width)
		if !ksh93 {
			sub = fmt.Sprintf("$(print -r -- \"$card\" | cut -c%d-%d)", start+1, start+width)
		}
		if num {
			return fmt.Sprintf("(( %s = %s ))", t, sub)
		}
		return fmt.Sprintf("%s=\"%s\"", t, sub)
	}
	return d
}

// ── Tcl ───────────────────────────────────────────────────────────────────

var tclSimple = regexp.MustCompile(`^(\$[A-Za-z_][\w]*(\([^()\[\]]*\))?|-?[0-9.]+)$`)

func tclVal(s string) string {
	if tclSimple.MatchString(s) {
		return s
	}
	return "[expr {" + s + "}]"
}

func tclDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"    ", "  "}), cmt: "# ",
		intDiv: "%s / %s", mod: "%s %% %s", dble: "double(%s)", ladder: "int(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ",
		ifOpen: "if {%s} {", ifElse: "} else {", ifClose: "}", stop: "exit 16",
		concat: "", order: "procs-first",
		lit: func(s string) string {
			return strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "$", "\\$", "[", "\\[", "]", "\\]").Replace(asciiText(s))
		},
		pstr:  func(x string, w int) string { return fmt.Sprintf("[format %%-%ds %s]", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("[format %%%dd %s]", w, tclVal(x)) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("[format %%+%dd %s]", w, tclVal(x)) },
		print: func(p string) string { return "puts \"" + p + "\"" },
	}
	lval := func(e *dialectEmitter, l *Expr) string {
		if l.K == "idx" {
			var ix []string
			for _, a := range l.Args {
				ix = append(ix, tclVal(e.ex(a)))
			}
			return l.Name + "(" + strings.Join(ix, ",") + ")"
		}
		return l.Name
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "var":
			return "$" + e.name(x.Name), true
		case "idx":
			return "$" + lval(e, x), true
		}
		return "", false
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		return "set " + lval(e, l) + " " + tclVal(r), true
	}
	d.callFn = func(e *dialectEmitter, name string, args []string) string {
		var as []string
		for _, a := range args {
			as = append(as, tclVal(a))
		}
		return strings.TrimSpace(name + " " + strings.Join(as, " "))
	}
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		if s.K != "do" {
			return false
		}
		v := s.Var
		if s.Step < 0 {
			e.line(depth, fmt.Sprintf("for {set %s %s} {$%s >= %s} {incr %s -1} {", v, tclVal(e.ex(s.From)), v, e.ex(s.To), v), v)
		} else {
			e.line(depth, fmt.Sprintf("for {set %s %s} {$%s <= %s} {incr %s} {", v, tclVal(e.ex(s.From)), v, e.ex(s.To), v), v)
		}
		e.body(s.Body, depth+1)
		e.line(depth, "}")
		return true
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		switch {
		case s.Role == "const":
			return []string{fmt.Sprintf("set %s %s", e.constName(s.Name), s.Init[0])}
		case len(s.Dims) > 0 && len(s.Init) > 0:
			var kv []string
			for i, v := range s.Init {
				kv = append(kv, strconv.Itoa(i+1), v)
			}
			return []string{fmt.Sprintf("array set %s {%s}", s.Name, strings.Join(kv, " "))}
		case len(s.Dims) > 0:
			return []string{fmt.Sprintf("array set %s {}", s.Name)}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("set %s %s", s.Name, dq(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("set %s %s", s.Name, s.Init[0])}
		case s.Type == "str":
			return []string{fmt.Sprintf("set %s {}", s.Name)}
		}
		return []string{fmt.Sprintf("set %s 0", s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		e.raw("#!/bin/sh")
		e.raw("# the next line restarts using tclsh \\")
		e.raw("exec tclsh \"$0\" \"$@\"")
		e.headerComments()
		e.globalsDecl("global")
		e.line(0, "set card {}")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.blank()
			e.line(0, "set deck [open deck.dat r]")
			return
		}
		e.raw(fmt.Sprintf("proc %s {%s} {", p.Name, strings.Join(p.Params, " ")))
		gl := append([]string{"deck", "card"}, e.usedGlobals(p)...)
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				gl = append(gl, e.constName(s.Name))
			}
		}
		e.line(1, "global "+strings.Join(gl, " "))
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "close $deck")
			e.line(0, "exit 0")
			return
		}
		e.raw("}")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "if {[gets $deck card] < 0} { error \"deck short\" }")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		t = strings.TrimPrefix(t, "$")
		if num {
			return fmt.Sprintf("scan [string range $card %d %d] %%d %s", start, start+width-1, t)
		}
		return fmt.Sprintf("set %s [string range $card %d %d]", t, start, start+width-1)
	}
	return d
}

// ── bc ────────────────────────────────────────────────────────────────────

func bcDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"    ", "  "}), cmt: "/* ", cmtEnd: " */",
		idx1: "%s[%s]", idx2: "%s[(%s - 1) * {D1} + %s]", intDiv: "%s / %s", mod: "%s %% %s", dble: "%s", ladder: "t(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ",
		assign: "%s = %s", call: "z = %s(%s)", callNone: "z = %s()", stop: "halt",
		ifOpen: "if (%s) {", ifElse: "} else {", ifClose: "}",
		forUp: "for (%[1]s = %[2]s; %[1]s <= %[3]s; %[1]s++) {", forDown: "for (%[1]s = %[2]s; %[1]s >= %[3]s; %[1]s--) {", forClose: "}",
		concat: ", ", lit: dq, order: "procs-first",
		pstr:  func(x string, w int) string { return "" },
		pnum:  func(x string, w int) string { return x },
		psgn:  func(x string, w int) string { return x },
		print: func(p string) string { return "print " + p + ", \"\\n\"" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.commentLines(0, []string{"bc has no strings: the awk prestage strips names off the deck and paste(1)", "puts them back next to the numbers this program prints."})
		e.line(0, "scale = 0")
		e.blank()
		e.raw("/* truncate toward zero, whatever scale the caller left behind */")
		e.raw("define t(x) {")
		e.line(1, "auto s")
		e.line(1, "s = scale; scale = 0; x = x / 1; scale = s")
		e.line(1, "return (x)")
		e.raw("}")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.blank()
			constInits(e, 0)
			scalarInits(e, 0)
			elementInits(e, 0)
			return
		}
		var ps []string
		for _, q := range p.Params {
			if len(e.paramSym(p, q).Dims) > 0 {
				q += "[]"
			}
			ps = append(ps, q)
		}
		e.raw(fmt.Sprintf("define %s(%s) {", p.Name, strings.Join(ps, ", ")))
		var auto []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				n := l.Name
				if len(l.Dims) > 0 {
					n += "[]"
				}
				auto = append(auto, n)
			}
		}
		if len(auto) > 0 {
			e.line(1, "auto "+strings.Join(auto, ", "))
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
			e.line(0, "quit")
			return
		}
		e.line(1, "return (0)")
		e.raw("}")
	}
	d.readCard = func(e *dialectEmitter, depth int) {}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return t + " = read()"
		}
		return fmt.Sprintf("/* %s: columns %d-%d, carried by the awk prestage */", t, start+1, start+width)
	}
	return d
}

// ── dc ────────────────────────────────────────────────────────────────────

func dcDialect(j *job) *dialect {
	regs := map[string]string{}
	pool := []rune("abcdefghijklmnopqrstuvwxyz0123456789@#$&_")
	procPool := []rune("ZYXWVUTSRQPO")
	next, nextP := 0, 0
	reg := func(name string) string {
		if r, ok := regs[name]; ok {
			return r
		}
		var r string
		if next < len(pool) {
			r = string(pool[next])
		} else {
			r = string(rune(0xC0 + next))
		}
		next++
		regs[name] = r
		return r
	}
	preg := func(name string) string {
		if r, ok := regs["proc:"+name]; ok {
			return r
		}
		r := string(procPool[nextP%len(procPool)])
		nextP++
		regs["proc:"+name] = r
		return r
	}
	d := &dialect{ind: "  ", cmt: "# ", byGlobal: true, order: "topo",
		rel: map[string]string{"lt": "<", "le": "!>", "eq": "=", "ne": "!=", "gt": ">", "ge": "!<"}, stop: "q",
		concat: " ", lit: func(s string) string {
			return "[" + strings.NewReplacer("[", "(", "]", ")").Replace(asciiText(s)) + "]n"
		},
		pstr:  func(x string, w int) string { return "" },
		pnum:  func(x string, w int) string { return x + " n [ ]n" },
		psgn:  func(x string, w int) string { return x + " n [ ]n" },
		print: func(p string) string { return p + " 10P" },
	}
	num := func(v int) string {
		if v < 0 {
			return "_" + strconv.Itoa(-v)
		}
		return strconv.Itoa(v)
	}
	var addr func(e *dialectEmitter, x *Expr) string
	addr = func(e *dialectEmitter, x *Expr) string {
		if len(x.Args) == 2 {
			return fmt.Sprintf("%s 1- %s * %s +", e.ex(x.Args[0]), e.dim1(x.Name), e.ex(x.Args[1]))
		}
		return e.ex(x.Args[0])
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "var":
			if v, ok := e.consts[x.Name]; ok {
				return v, true
			}
			return "l" + reg(x.Name), true
		case "int":
			return num(x.Val), true
		case "dlit":
			return x.S, true
		case "idx":
			return addr(e, x) + " ;" + reg(x.Name), true
		case "bin":
			return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " " + x.Op, true
		case "call":
			if x.Name == "mod" {
				return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " %", true
			}
			return e.ex(x.Args[0]), true
		case "ladder":
			return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " * 1 /", true
		case "str":
			return "[" + asciiText(x.S) + "]", true
		}
		return "0", true
	}
	store := func(e *dialectEmitter, l *Expr, val string) string {
		if l.K == "idx" {
			return val + " " + addr(e, l) + " :" + reg(l.Name)
		}
		return val + " s" + reg(l.Name)
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) { return store(e, l, r), true }
	macro := func(depth int, base rune) string { return string(base + rune(depth%12)) }
	var test func(e *dialectEmitter, c *Cond, m string, invert bool) string
	test = func(e *dialectEmitter, c *Cond, m string, invert bool) string {
		if c.Op == "and" || c.Op == "or" {
			// dc has no boolean operators; test the first clause only and let
			// the macro re-check the rest (it never does).
			return test(e, c.A, m, invert)
		}
		op := d.rel[c.Op]
		if invert {
			op = map[string]string{"<": "!<", "!>": ">", "=": "!=", "!=": "=", ">": "!>", "!<": "<"}[op]
		}
		return e.ex(c.R) + " " + e.ex(c.L) + " " + op + m
	}
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		switch s.K {
		case "if":
			m := macro(depth, 'A')
			e.line(depth, "[")
			e.body(s.Then, depth+1)
			e.line(depth, "]s"+m)
			if s.Else != nil {
				me := macro(depth, 'a'+20)
				e.line(depth, "[")
				e.body(s.Else, depth+1)
				e.line(depth, "]s"+me)
				e.line(depth, test(e, s.Cond, me, true), condRefs(s.Cond)...)
			}
			e.line(depth, test(e, s.Cond, m, false), condRefs(s.Cond)...)
			return true
		case "do":
			m := macro(depth, 'M')
			v := reg(s.Var)
			op, step := "!>", "1+"
			if s.Step < 0 {
				op, step = "!<", "1-"
			}
			to := e.ex(s.To)
			e.line(depth, e.ex(s.From)+" s"+v+"   # "+s.Var, s.Var)
			e.line(depth, "[")
			e.body(s.Body, depth+1)
			e.line(depth+1, "l"+v+" "+step+" s"+v)
			e.line(depth+1, to+" l"+v+" "+op+m)
			e.line(depth, "]s"+m)
			e.line(depth, to+" l"+v+" "+op+m, s.Var)
			return true
		case "call":
			p := e.procs[s.Name]
			if p != nil {
				for i, a := range s.Args {
					if i < len(p.Params) {
						e.line(depth, store(e, V(p.Params[i]), e.ex(a)))
					}
				}
			}
			e.line(depth, "l"+preg(s.Name)+"x   # "+s.Name, s.Name)
			return true
		}
		return false
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.commentLines(0, []string{"dc keeps no strings: names are stripped from the deck by the awk prestage", "and pasted back beside these numbers.", "", "register map:"})
		for _, s := range e.j.p.Globals {
			if s.Role != "const" {
				e.commentLines(0, []string{fmt.Sprintf("  %s  %s", reg(s.Name), s.Name)})
			}
		}
		e.line(0, "0k")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			for _, s := range e.j.p.Globals {
				if s.Role == "const" {
					continue
				}
				if len(s.Init) > 0 && len(s.Dims) > 0 {
					for i, v := range s.Init {
						if s.Type == "str" {
							continue
						}
						e.line(0, fmt.Sprintf("%s %d :%s", v, i+1, reg(s.Name)), s.Name)
					}
				} else if len(s.Init) > 0 && s.Type != "str" {
					e.line(0, s.Init[0]+" s"+reg(s.Name), s.Name)
				} else if len(s.Init) > 0 {
					e.line(0, "['"+s.Init[0]+"']s"+reg(s.Name)+"   # "+s.Name+", never printed", s.Name)
				}
			}
			return
		}
		e.raw("[   # " + p.Name)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "q")
			return
		}
		e.raw("]s" + preg(p.Name))
	}
	d.readCard = func(e *dialectEmitter, depth int) {}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			// t is a load ("lx" or "... ;x"): turn it into the matching store
			if strings.HasPrefix(t, "l") && len(t) <= 3 {
				return "? s" + t[1:]
			}
			if i := strings.LastIndex(t, " ;"); i >= 0 {
				return "? " + t[:i] + " :" + t[i+2:]
			}
			return "? sz"
		}
		return fmt.Sprintf("# columns %d-%d: name, carried by the awk prestage", start+1, start+width)
	}
	return d
}

// ── ColdFusion ────────────────────────────────────────────────────────────

func cfmlDialect(j *job) *dialect {
	x := j.x
	mx := j.in.Year >= 2002 // CFFUNCTION arrived with ColdFusion MX
	q := func(s string) string {
		return "\"" + strings.NewReplacer("\"", "\"\"", "#", "##").Replace(asciiText(s)) + "\""
	}
	d := &dialect{ind: pick(x, []string{"  ", "    ", "\t"}), cmt: "<!--- ", cmtEnd: " --->",
		idx1: "%s[%s]", idx2: "%s[%s][%s]", intDiv: "%s \\ %s", mod: "%s MOD %s", dble: "%s", ladder: "Fix(%s * %s)",
		rel: map[string]string{"lt": "LT", "le": "LTE", "eq": "EQ", "ne": "NEQ", "gt": "GT", "ge": "GTE"}, and: " AND ", or: " OR ",
		assign: "<cfset %s = %s>", call: "<cfset %s(%s)>", callNone: "<cfset %s()>", stop: "<cfabort>",
		ifOpen: "<cfif %s>", ifElse: "<cfelse>", ifClose: "</cfif>",
		forUp: `<cfloop index="%s" from="#%s#" to="#%s#">`, forDown: `<cfloop index="%s" from="#%s#" to="#%s#" step="-1">`, forClose: "</cfloop>",
		concat: " & ", lit: q, order: "main-first",
		pstr: func(x string, w int) string { return fmt.Sprintf("LJustify(%s, %d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("RJustify(%s, %d)", x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("RJustify(NumberFormat(%s, \"+%s\"), %d)", x, strings.Repeat("_", w-1), w)
		},
		print: func(p string) string { return "<cfset WriteOutput(" + p + " & Chr(10))>" },
	}
	if !mx {
		d.byGlobal = true
		d.callNone = `<cfinclude template="%s.cfm">`
		d.call = d.callNone
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		switch {
		case s.Role == "const":
			return []string{fmt.Sprintf("<cfset %s = %s>", e.constName(s.Name), s.Init[0])}
		case len(s.Dims) > 0 && len(s.Init) > 0:
			return []string{fmt.Sprintf("<cfset %s = ListToArray(\"%s\")>", s.Name, strings.Join(s.Init, ","))}
		case len(s.Dims) > 0:
			return []string{fmt.Sprintf("<cfset %s = ArrayNew(%d)>", s.Name, len(s.Dims))}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("<cfset %s = %s>", s.Name, q(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("<cfset %s = %s>", s.Name, s.Init[0])}
		case s.Type == "str":
			return []string{fmt.Sprintf("<cfset %s = \"\">", s.Name)}
		}
		return []string{fmt.Sprintf("<cfset %s = 0>", s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		e.raw(`<cfsetting enablecfoutputonly="yes" requesttimeout="900" showdebugoutput="no">`)
		e.headerComments()
		e.raw(`<cfcontent type="text/plain">`)
		e.globalsDecl("global")
		if !mx {
			for _, s := range e.byGlobalSyms() {
				e.defLine(0, d.decl(e, s, "global")[0], s.Name)
			}
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, `<cffile action="read" file="#ExpandPath('deck.dat')#" variable="deck">`)
			e.line(0, "<cfset cardno = 0>")
			return
		}
		if !mx {
			e.raw("<!--- ===== " + p.Name + ".cfm (cfinclude'd; shares the caller's variables) ===== --->")
			return
		}
		e.raw(fmt.Sprintf(`<cffunction name="%s" access="private" returntype="void" output="true">`, p.Name))
		for _, q := range p.Params {
			typ := "numeric"
			if len(e.paramSym(p, q).Dims) > 0 {
				typ = "array"
			}
			e.defLine(1, fmt.Sprintf(`<cfargument name="%s" type="%s" required="yes">`, q, typ), q)
		}
		for _, l := range p.Locals {
			if l.Role != "param" {
				init := "0"
				if len(l.Dims) > 0 {
					init = fmt.Sprintf("ArrayNew(%d)", len(l.Dims))
				} else if l.Type == "str" {
					init = `""`
				}
				e.defLine(1, fmt.Sprintf("<cfset var %s = %s>", l.Name, init), l.Name)
			}
		}
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain || !mx {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "<cfsetting enablecfoutputonly=\"no\">")
			return
		}
		if mx {
			e.raw("</cffunction>")
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "<cfset cardno = cardno + 1>")
		e.line(depth, "<cfset card = ListGetAt(deck, cardno, Chr(10))>")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("<cfset %s = Val(Mid(card, %d, %d))>", t, start+1, width)
		}
		return fmt.Sprintf("<cfset %s = Mid(card, %d, %d)>", t, start+1, width)
	}
	return d
}

// ── Classic ASP (VBScript) ────────────────────────────────────────────────

func aspDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"    ", "  ", "\t"}), cmt: "' ",
		idx1: "%s(%s)", idx2: "%s(%s, %s)", intDiv: "%s \\ %s", mod: "%s Mod %s", dble: "CDbl(%s)", ladder: "Fix(CDbl(%s) * %s)",
		rel: stdRel, and: " And ", or: " Or ", assign: "%s = %s", call: "Call %s(%s)", callNone: "Call %s()", stop: "Response.End",
		ifOpen: "If %s Then", ifElse: "Else", ifClose: "End If",
		forUp: "For %s = %s To %s", forDown: "For %s = %s To %s Step -1", forClose: "Next",
		concat: " & ", lit: dq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("Left(%s & Space(%d), %d)", x, w, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("Right(Space(%d) & CStr(%s), %d)", w, x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("Right(Space(%d) & SignedStr(%s), %d)", w, x, w) },
		print: func(p string) string { return "Response.Write " + p + " & vbCrLf" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("Const %s = %s", e.name(s.Name), s.Init[0])}
		}
		if len(s.Dims) > 0 {
			return []string{fmt.Sprintf("Dim %s(%s)", s.Name, e.dimList(s, func(dn string) string { return dn }, ", "))}
		}
		return []string{"Dim " + s.Name}
	}
	d.header = func(e *dialectEmitter) {
		e.raw(`<%@ Language="VBScript" CodePage=1252 %>`)
		e.raw("<%")
		e.headerComments()
		e.raw("Option Explicit")
		e.raw("Response.ContentType = \"text/plain\"")
		e.raw("Server.ScriptTimeout = 900")
		e.globalsDecl("global")
		e.line(0, "Dim fso, ts, card")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.blank()
			scalarInits(e, 0)
			elementInits(e, 0)
			e.line(0, `Set fso = Server.CreateObject("Scripting.FileSystemObject")`)
			e.line(0, `Set ts = fso.OpenTextFile(Server.MapPath("deck.dat"), 1)`)
			return
		}
		e.raw(fmt.Sprintf("Sub %s(%s)", p.Name, strings.Join(p.Params, ", ")))
		e.locals(p, 1)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 0
		}
		return 1
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "ts.Close")
			e.line(0, "Set ts = Nothing: Set fso = Nothing")
			return
		}
		e.raw("End Sub")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw("Function SignedStr(n)")
		e.line(1, "If n > 0 Then SignedStr = \"+\" & CStr(n) Else SignedStr = CStr(n)")
		e.raw("End Function")
		e.raw("%>")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "card = ts.ReadLine") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = CLng(\"0\" & Trim(Mid(card, %d, %d)))", t, start+1, width)
		}
		return fmt.Sprintf("%s = Mid(card, %d, %d)", t, start+1, width)
	}
	return d
}

// ── Modula-2 ──────────────────────────────────────────────────────────────

func modula2Dialect(j *job) *dialect {
	x := j.x
	mod := strings.ToUpper(j.id.Program[:1]) + strings.ToLower(j.id.Program[1:])
	// each vendor shipped its own library modules
	lib := pick(x, []string{"logitech", "topspeed", "stonybrook"})
	if j.in.Year < 1991 {
		lib = "logitech"
	}
	numFn := pick(x, []string{"Num", "CardNum", "FieldInt", "ColVal"})
	sgnFn := pick(x, []string{"WriteSigned", "PutSigned", "WrSgn"})
	digitLoop := x.chance(0.5)
	typeName := func(s Sym) string {
		base := map[string]string{"int": "INTEGER", "dbl": "LONGREAL", "str": fmt.Sprintf("ARRAY [0..%d] OF CHAR", maxi(s.Len, 1)-1)}[s.Type]
		if base == "" {
			base = "INTEGER"
		}
		if len(s.Dims) == 0 {
			return base
		}
		var rs []string
		for _, dd := range s.Dims {
			rs = append(rs, fmt.Sprintf("[1..%d]", dd))
		}
		return "ARRAY " + strings.Join(rs, ", ") + " OF " + base
	}
	d := &dialect{ind: pick(x, []string{"  ", "   ", "    "}), cmt: "(* ", cmtEnd: " *)",
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "%s DIV %s", mod: "%s MOD %s", dble: "LFLOAT(%s)", ladder: "TRUNC(LFLOAT(%s) * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "#", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: "%s := %s;", call: "%s(%s);", callNone: "%s;", stop: "HALT;",
		ifOpen: "IF %s THEN", ifElse: "ELSE", ifClose: "END;",
		forUp: "FOR %s := %s TO %s DO", forDown: "FOR %s := %s TO %s BY -1 DO", forClose: "END;",
		concat: "; ", lit: func(s string) string { return "WriteString(" + dq(asciiText(s)) + ")" }, order: "procs-first",
		pstr:  func(x string, w int) string { return "WriteString(" + x + ")" },
		pnum:  func(x string, w int) string { return fmt.Sprintf("WriteInt(%s, %d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s(%s, %d)", sgnFn, x, w) },
		print: func(p string) string { return p + "; WriteLn;" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%s = %s;", e.constName(s.Name), s.Init[0])}
		}
		return []string{fmt.Sprintf("%s: %s;", s.Name, typeName(s))}
	}
	d.header = func(e *dialectEmitter) {
		e.raw("MODULE " + mod + ";")
		e.headerComments()
		switch lib {
		case "topspeed":
			e.raw("FROM IO IMPORT WrStr AS WriteString, WrInt, WrLn AS WriteLn;")
			e.raw("FROM FIO IMPORT File, Open, Close, RdStr, EOF;")
			e.raw("FROM Str IMPORT Copy, StrToInt;")
		case "stonybrook":
			e.raw("FROM STextIO IMPORT WriteString, WriteLn;")
			e.raw("FROM SWholeIO IMPORT WriteInt;")
			e.raw("FROM SeqFile IMPORT ChanId, OpenRead, Close, read;")
			e.raw("FROM TextIO IMPORT ReadString;")
			e.raw("FROM Strings IMPORT Extract AS Copy;")
			e.raw("FROM WholeStr IMPORT StrToInt;")
		default:
			e.raw("FROM InOut IMPORT WriteString, WriteInt, WriteLn, OpenInput, CloseInput, ReadString, Done;")
			e.raw("FROM Strings IMPORT Copy;")
			e.raw("FROM Conversions IMPORT StrToInt;")
		}
		if x.chance(0.5) {
			e.raw("FROM MathLib0 IMPORT entier;")
		}
		if lib == "topspeed" {
			e.raw("VAR deck: File;")
		} else if lib == "stonybrook" {
			e.raw("VAR deck: ChanId; res: CARDINAL;")
		}
		e.blank()
		e.raw("CONST")
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.raw("VAR")
		for _, s := range e.j.p.Globals {
			if s.Role != "const" {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.defLine(1, "card: ARRAY [0..79] OF CHAR;", "card")
		e.blank()
		e.raw("PROCEDURE " + numFn + "(VAR s: ARRAY OF CHAR; start, width: CARDINAL): INTEGER;")
		if digitLoop {
			e.line(1, "VAR i: CARDINAL; n: INTEGER;")
			e.raw("BEGIN")
			e.line(1, "n := 0;")
			e.line(1, "FOR i := start TO start + width - 1 DO")
			e.line(2, "IF (s[i] >= '0') AND (s[i] <= '9') THEN n := n * 10 + INTEGER(ORD(s[i]) - ORD('0')) END")
			e.line(1, "END;")
			e.line(1, "RETURN n")
		} else {
			e.line(1, "VAR buf: ARRAY [0..15] OF CHAR; n: INTEGER; ok: BOOLEAN;")
			e.raw("BEGIN")
			e.line(1, "Copy(s, start, width, buf); StrToInt(buf, n, ok);")
			e.line(1, "IF NOT ok THEN n := 0 END;")
			e.line(1, "RETURN n")
		}
		e.raw("END " + numFn + ";")
		e.blank()
		e.raw("PROCEDURE " + sgnFn + "(n: INTEGER; w: CARDINAL);")
		e.raw("BEGIN")
		if x.chance(0.5) {
			e.line(1, "IF n > 0 THEN WriteString(\"+\"); WriteInt(n, w - 1) ELSE WriteInt(n, w) END")
		} else {
			e.line(1, "IF n > 0 THEN")
			e.line(2, "WriteString(\"+\");")
			e.line(2, "WriteInt(n, w - 1)")
			e.line(1, "ELSE")
			e.line(2, "WriteInt(n, w)")
			e.line(1, "END")
		}
		e.raw("END " + sgnFn + ";")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("BEGIN (* " + mod + " *)")
			scalarInits(e, 1)
			elementInits(e, 1)
			switch lib {
			case "topspeed":
				e.line(1, "deck := Open(\"DECK.DAT\");")
			case "stonybrook":
				e.line(1, "OpenRead(deck, \"DECK.DAT\", read, res);")
			default:
				e.line(1, "OpenInput(\"DAT\");")
				e.line(1, "IF NOT Done THEN WriteString(\"NO DECK\"); WriteLn; HALT END;")
			}
			return
		}
		var ps []string
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			if len(s.Dims) > 0 {
				ps = append(ps, "VAR "+q+": ARRAY OF INTEGER")
			} else {
				ps = append(ps, q+": "+typeName(s))
			}
		}
		e.raw(fmt.Sprintf("PROCEDURE %s(%s);", p.Name, strings.Join(ps, "; ")))
		var vars []Sym
		for _, l := range p.Locals {
			if l.Role != "param" {
				vars = append(vars, l)
			}
		}
		if len(vars) > 0 {
			e.raw("VAR")
			for _, l := range vars {
				e.defLine(1, d.decl(e, l, "local")[0], l.Name)
			}
		}
		e.raw("BEGIN")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, map[string]string{"topspeed": "Close(deck)", "stonybrook": "Close(deck)", "logitech": "CloseInput"}[lib])
			e.raw("END " + mod + ".")
			return
		}
		e.raw("END " + p.Name + ";")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, map[string]string{"topspeed": "RdStr(deck, card);", "stonybrook": "ReadString(deck, card);", "logitech": "ReadString(card);"}[lib])
	}
	if x.chance(0.3) {
		d.whileFor, d.whileHd, d.whileEnd = true, "WHILE %s DO", "END;"
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s := %s(card, %d, %d);", t, numFn, start, width)
		}
		return fmt.Sprintf("Copy(card, %d, %d, %s);", start, width, t)
	}
	return d
}

// ── Lisp (Franz Lisp, Common Lisp) ────────────────────────────────────────

func lispDialect(j *job) *dialect {
	franz := j.id.Language == "Franz Lisp"
	ops := map[string]string{"+": "+", "-": "-", "*": "*", "/": "/"}
	d := &dialect{ind: "  ", cmt: ";; ", order: "topo",
		rel:    map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "/=", "gt": ">", "ge": ">="},
		stop:   "(throw 'abend 16)",
		concat: " ", lit: dq,
		pstr:  func(x string, w int) string { return fmt.Sprintf("(pad-right %s %d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("(format nil \"~%dD\" %s)", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("(format nil \"~%d@D\" %s)", w, x) },
		print: func(p string) string { return "(format t \"~{~A~}~%\" (list " + p + "))" },
	}
	if franz {
		d.rel["ne"] = "/="
		d.print = func(p string) string { return "(msg " + p + " N)" }
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "var":
			return e.name(x.Name), true
		case "dlit":
			return x.S + "d0", true
		case "idx":
			var ix []string
			for _, a := range x.Args {
				ix = append(ix, e.ex(a))
			}
			return "(aref " + x.Name + " " + strings.Join(ix, " ") + ")", true
		case "bin":
			l, r := e.ex(x.Args[0]), e.ex(x.Args[1])
			if x.Op == "/" && !isDouble(x.Args[0]) && !isDouble(x.Args[1]) {
				return "(truncate " + l + " " + r + ")", true
			}
			return "(" + ops[x.Op] + " " + l + " " + r + ")", true
		case "call":
			if x.Name == "mod" {
				return "(mod " + e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + ")", true
			}
			return "(float " + e.ex(x.Args[0]) + " 1d0)", true
		case "ladder":
			return "(values (truncate (* " + e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + ")))", true
		}
		return "", false
	}
	d.condFn = func(e *dialectEmitter, c *Cond) (string, bool) {
		if c.Op == "and" || c.Op == "or" {
			return "(" + c.Op + " " + e.cond(c.A) + " " + e.cond(c.B) + ")", true
		}
		return "(" + d.rel[c.Op] + " " + e.ex(c.L) + " " + e.ex(c.R) + ")", true
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		return "(setf " + e.ex(l) + " " + r + ")", true
	}
	d.callFn = func(e *dialectEmitter, name string, args []string) string {
		if len(args) == 0 {
			return "(" + name + ")"
		}
		return "(" + name + " " + strings.Join(args, " ") + ")"
	}
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		switch s.K {
		case "if":
			if s.Else == nil {
				e.line(depth, "(when "+e.cond(s.Cond), condRefs(s.Cond)...)
				e.body(s.Then, depth+1)
				e.closeLast(")")
				return true
			}
			e.line(depth, "(if "+e.cond(s.Cond), condRefs(s.Cond)...)
			e.line(depth+2, "(progn")
			e.body(s.Then, depth+3)
			e.closeLast(")")
			e.line(depth+2, "(progn")
			e.body(s.Else, depth+3)
			e.closeLast("))")
			return true
		case "do":
			v, from, to := s.Var, e.ex(s.From), e.ex(s.To)
			if franz {
				step, test := "1+", ">"
				if s.Step < 0 {
					step, test = "1-", "<"
				}
				e.line(depth, fmt.Sprintf("(do ((%s %s (%s %s))) ((%s %s %s))", v, from, step, v, test, v, to), v)
			} else {
				dir := "to"
				if s.Step < 0 {
					dir = "downto"
				}
				e.line(depth, fmt.Sprintf("(loop for %s from %s %s %s do", v, from, dir, to), v)
			}
			e.body(s.Body, depth+1)
			e.closeLast(")")
			return true
		}
		return false
	}
	init := func(s Sym) string {
		zero := map[string]string{"int": "0", "dbl": "0d0", "str": "\"\""}[s.Type]
		if zero == "" {
			zero = "0"
		}
		switch {
		case len(s.Dims) == 2:
			return fmt.Sprintf("(make-array '(%d %d) :initial-element %s)", s.Dims[0]+1, s.Dims[1]+1, zero)
		case len(s.Dims) == 1 && len(s.Init) > 0:
			var vs []string
			for _, v := range s.Init {
				if s.Type == "str" {
					vs = append(vs, dq(v))
				} else if s.Type == "dbl" {
					vs = append(vs, v+"d0")
				} else {
					vs = append(vs, v)
				}
			}
			return "(vector " + zero + " " + strings.Join(vs, " ") + ")"
		case len(s.Dims) == 1:
			return fmt.Sprintf("(make-array %d :initial-element %s)", s.Dims[0]+1, zero)
		case len(s.Init) > 0 && s.Type == "str":
			return dq(s.Init[0])
		case len(s.Init) > 0:
			return s.Init[0]
		}
		return zero
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("(defconstant %s %s)", e.constName(s.Name), s.Init[0])}
		}
		return []string{fmt.Sprintf("(defvar %s %s)", s.Name, init(s))}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		if !franz {
			pkg := strings.ToLower(e.j.id.Program)
			e.raw("(defpackage :" + pkg + " (:use :common-lisp))")
			e.raw("(in-package :" + pkg + ")")
		}
		e.globalsDecl("global")
		e.line(0, "(defvar *deck* nil)")
		e.line(0, "(defvar *card* \"\")")
		e.blank()
		e.raw("(defun pad-right (s w)")
		e.line(1, "(format nil \"~vA\" w (subseq s 0 (min w (length s)))))")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		name := p.Name
		if isMain {
			name = "main"
		}
		head := "(defun " + name + " (" + strings.Join(p.Params, " ")
		var aux []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				aux = append(aux, "("+l.Name+" "+init(l)+")")
			}
		}
		if len(aux) > 0 {
			if len(p.Params) > 0 {
				head += " "
			}
			head += "&aux " + strings.Join(aux, " ")
		}
		e.raw(head + ")")
		if isMain {
			e.line(1, "(setf *deck* (open \"deck.dat\" :direction :input))")
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "(close *deck*)")
		}
		e.closeLast(")")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw("(catch 'abend (main))")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "(setf *card* (read-line *deck* nil \"\"))")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("(setf %s (or (parse-integer *card* :start %d :end %d :junk-allowed t) 0))", t, start, start+width)
		}
		return fmt.Sprintf("(setf %s (subseq *card* %d %d))", t, start, start+width)
	}
	return d
}

// ── Smalltalk ─────────────────────────────────────────────────────────────

var stAt = regexp.MustCompile(`^\((.*) at: ([^()]*)\)$`)

func smalltalkDialect(j *job) *dialect {
	x := j.x
	class := strings.ToUpper(j.id.Program[:1]) + strings.ToLower(j.id.Program[1:])
	vw := j.id.Language != "Smalltalk-80"
	selector := func(name string, args []string) string {
		n := strings.ToLower(name[:1]) + name[1:]
		if len(args) == 0 {
			return n
		}
		parts := []string{n + ": " + args[0]}
		for _, a := range args[1:] {
			parts = append(parts, "with: "+a)
		}
		return strings.Join(parts, " ")
	}
	d := &dialect{ind: pick(x, []string{"    ", "\t"}), cmt: "\"", cmtEnd: "\"", parenAll: true,
		idx1: "(%s at: %s)", idx2: "((%s at: %s) at: %s)", intDiv: "%s // %s", mod: "%s \\\\ %s", dble: "%s asFloat", ladder: "(%s * %s) truncated",
		rel: map[string]string{"lt": " < ", "le": " <= ", "eq": " = ", "ne": " ~= ", "gt": " > ", "ge": " >= "}, and: " & ", or: " | ",
		assign: "%s := %s.", stop: "^self error: 'deck rejected'.",
		ifOpen: "(%s) ifTrue: [", ifElse: "] ifFalse: [", ifClose: "].",
		whileFor: true, whileHd: "[%s] whileTrue: [", whileEnd: "].",
		concat: ", ", lit: sq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("(self pad: %s to: %d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("(self num: %s width: %d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("(self signed: %s width: %d)", x, w) },
		print: func(p string) string { return "Transcript show: " + p + "; cr." },
	}
	put := func(t, v string) string {
		if m := stAt.FindStringSubmatch(t); m != nil {
			return fmt.Sprintf("%s at: %s put: %s.", strings.TrimSuffix(strings.TrimPrefix(m[1], "("), ")"), m[2], v)
		}
		return t + " := " + v + "."
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) { return put(e.ex(l), r), true }
	d.callFn = func(e *dialectEmitter, name string, args []string) string {
		return "self " + selector(name, args) + "."
	}
	method := func(e *dialectEmitter, cat string, head string) {
		e.blank()
		e.raw("!" + class + " methodsFor: '" + cat + "'!")
		e.raw(head)
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		var ivars []string
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				ivars = append(ivars, e.constName(s.Name))
			} else {
				ivars = append(ivars, s.Name)
			}
		}
		ivars = append(ivars, "deck", "card")
		e.raw("Object subclass: #" + class)
		e.raw("    instanceVariableNames: '" + strings.Join(ivars, " ") + "'")
		e.raw("    classVariableNames: ''")
		e.raw("    poolDictionaries: ''")
		if vw {
			e.raw("    category: 'ACPWB-Research'!")
		} else {
			e.raw("    category: 'Research-Tabulation'!")
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			method(e, "running", "run")
		} else {
			method(e, "tabulation", selector(p.Name, p.Params))
		}
		var temps []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				temps = append(temps, l.Name)
			}
		}
		if len(temps) > 0 {
			e.line(1, "| "+strings.Join(temps, " ")+" |")
		}
		if isMain {
			for _, s := range e.j.p.Globals {
				switch {
				case s.Role == "const":
					e.line(1, e.constName(s.Name)+" := "+s.Init[0]+".", s.Name)
				case len(s.Dims) == 2:
					e.line(1, fmt.Sprintf("%s := (1 to: %d) collect: [:each | Array new: %d withAll: 0].", s.Name, s.Dims[0], s.Dims[1]), s.Name)
				case len(s.Dims) == 1 && len(s.Init) > 0:
					var vs []string
					for _, v := range s.Init {
						if s.Type == "str" {
							v = sq(v)
						}
						vs = append(vs, v)
					}
					e.line(1, fmt.Sprintf("%s := #(%s) copy.", s.Name, strings.Join(vs, " ")), s.Name)
				case len(s.Dims) == 1:
					zero := "0"
					if s.Type == "str" {
						zero = "''"
					}
					e.line(1, fmt.Sprintf("%s := Array new: %d withAll: %s.", s.Name, s.Dims[0], zero), s.Name)
				case len(s.Init) > 0 && s.Type == "str":
					e.line(1, s.Name+" := "+sq(s.Init[0])+".", s.Name)
				case len(s.Init) > 0:
					e.line(1, s.Name+" := "+s.Init[0]+".", s.Name)
				default:
					zero := "0"
					if s.Type == "str" {
						zero = "''"
					}
					e.line(1, s.Name+" := "+zero+".", s.Name)
				}
			}
			if vw {
				e.line(1, "deck := 'deck.dat' asFilename readStream.")
			} else {
				e.line(1, "deck := (FileStream oldFileNamed: 'deck.dat').")
			}
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "deck close.")
		}
		e.raw("! !")
	}
	d.footer = func(e *dialectEmitter) {
		method(e, "printing", "pad: aString to: width")
		e.line(1, "^(aString , (String new: width withAll: $ )) copyFrom: 1 to: width")
		e.raw("! !")
		method(e, "printing", "num: anInteger width: width")
		e.line(1, "| s |")
		e.line(1, "s := anInteger printString.")
		e.line(1, "^(String new: (width - s size max: 0) withAll: $ ) , s")
		e.raw("! !")
		method(e, "printing", "signed: anInteger width: width")
		e.line(1, "^self num: (anInteger > 0 ifTrue: ['+' , anInteger printString] ifFalse: [anInteger printString]) width: width")
		e.raw("! !")
		e.blank()
		e.raw(class + " new run!")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "card := deck nextLine.") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		v := fmt.Sprintf("(card copyFrom: %d to: %d)", start+1, start+width)
		if num {
			v = "(" + v + " trimSeparators asNumber)"
		}
		return put(t, v)
	}
	return d
}

// ── Forth ─────────────────────────────────────────────────────────────────

func forthDialect(j *job) *dialect {
	ans := j.in.Year >= 1996
	d := &dialect{ind: "  ", cmt: "\\ ", upper: true, byGlobal: true, order: "topo",
		rel: map[string]string{"lt": "<", "le": "> 0=", "eq": "=", "ne": "= 0=", "gt": ">", "ge": "< 0="}, and: " AND", or: " OR",
		ifOpen: "%s IF", ifElse: "ELSE", ifClose: "THEN", stop: `ABORT" DECK REJECTED"`,
		concat: " ", lit: func(s string) string { return `." ` + strings.ReplaceAll(cardText(s), `"`, "'") + `"` },
		pstr:  func(x string, w int) string { return fmt.Sprintf("%s %d TYPE", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s %d .R", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s %d .R+", x, w) },
		print: func(p string) string { return "CR " + p },
	}
	if ans {
		d.rel["ne"] = "<>"
	}
	slot := func(e *dialectEmitter, x *Expr) string {
		var ix string
		if len(x.Args) == 2 {
			ix = fmt.Sprintf("%s 1- %s * %s +", e.ex(x.Args[0]), e.dim1(x.Name), e.ex(x.Args[1]))
		} else {
			ix = e.ex(x.Args[0])
		}
		if e.symType(x.Name) == "str" {
			return ix + " 40 * " + x.Name + " +"
		}
		return ix + " CELLS " + x.Name + " +"
	}
	d.exFn = func(e *dialectEmitter, x *Expr) (string, bool) {
		switch x.K {
		case "var":
			if _, ok := e.consts[x.Name]; ok {
				return x.Name, true
			}
			if e.symType(x.Name) == "str" {
				return x.Name, true
			}
			return x.Name + " @", true
		case "dlit":
			return hundredths(x.S), true
		case "idx":
			if e.symType(x.Name) == "str" {
				return slot(e, x), true
			}
			return slot(e, x) + " @", true
		case "bin":
			op := x.Op
			return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " " + op, true
		case "call":
			if x.Name == "mod" {
				return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " MOD", true
			}
			return e.ex(x.Args[0]), true
		case "ladder":
			return e.ex(x.Args[0]) + " " + e.ex(x.Args[1]) + " 100 */", true
		}
		return "", false
	}
	d.condFn = func(e *dialectEmitter, c *Cond) (string, bool) {
		if c.Op == "and" || c.Op == "or" {
			return e.cond(c.A) + " " + e.cond(c.B) + map[string]string{"and": " AND", "or": " OR"}[c.Op], true
		}
		return e.ex(c.L) + " " + e.ex(c.R) + " " + d.rel[c.Op], true
	}
	d.assignFn = func(e *dialectEmitter, l *Expr, r string) (string, bool) {
		if l.K == "idx" {
			return r + " " + slot(e, l) + " !", true
		}
		return r + " " + l.Name + " !", true
	}
	d.callFn = func(e *dialectEmitter, name string, args []string) string { return name }
	d.stmtFn = func(e *dialectEmitter, s *Stmt, depth int) bool {
		if s.K != "do" {
			return false
		}
		if s.Step < 0 {
			e.line(depth, fmt.Sprintf("%s 1- %s DO  I %s !", e.ex(s.To), e.ex(s.From), s.Var), s.Var)
			e.body(s.Body, depth+1)
			e.line(depth, "-1 +LOOP")
			return true
		}
		e.line(depth, fmt.Sprintf("%s 1+ %s DO  I %s !", e.ex(s.To), e.ex(s.From), s.Var), s.Var)
		e.body(s.Body, depth+1)
		e.line(depth, "LOOP")
		return true
	}
	decl := func(e *dialectEmitter, s Sym) string {
		switch {
		case s.Role == "const":
			return s.Init[0] + " CONSTANT " + s.Name
		case len(s.Dims) > 0 && len(s.Init) > 0 && s.Type != "str":
			var vs []string
			for _, v := range s.Init {
				if s.Type == "dbl" {
					v = hundredths(v)
				}
				vs = append(vs, v+" ,")
			}
			return "CREATE " + s.Name + " 0 , " + strings.Join(vs, " ")
		case len(s.Dims) > 0:
			n := 1
			for _, dd := range s.Dims {
				n *= dd
			}
			if s.Type == "str" {
				return fmt.Sprintf("CREATE %s %d 40 * ALLOT", s.Name, n+1)
			}
			return fmt.Sprintf("CREATE %s %d CELLS ALLOT", s.Name, n+1)
		case s.Type == "str" && len(s.Init) > 0:
			return fmt.Sprintf("CREATE %s %d ALLOT\n: %s$ ( -- addr len )  S\" '%s'\" ;", s.Name, maxi(s.Len, 1)+2, s.Name, s.Init[0])
		case s.Type == "str":
			return fmt.Sprintf("CREATE %s %d ALLOT", s.Name, maxi(s.Len, 1))
		}
		return "VARIABLE " + s.Name
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		if ans {
			e.raw("MARKER -" + e.j.id.Program)
		} else {
			e.raw("FORGET TASK  : TASK ;")
		}
		e.raw("DECIMAL")
		for _, s := range e.j.p.Globals {
			e.defLine(0, decl(e, s), s.Name)
		}
		for _, s := range e.byGlobalSyms() {
			e.defLine(0, decl(e, s), s.Name)
		}
		e.defLine(0, "CREATE CARD 80 ALLOT", "CARD")
		e.blank()
		e.raw(": .R+ ( n w -- )  OVER 0> IF  1- .\" +\"  THEN .R ;")
		e.raw(": >NUM ( addr len -- n )  0 0 2SWAP >NUMBER 2DROP DROP ;")
		if ans {
			e.raw("0 VALUE DECK-FILE")
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw(": " + e.j.id.Program + " ( -- )")
			if ans {
				e.line(1, `S" deck.dat" R/O OPEN-FILE THROW TO DECK-FILE`)
			}
			return
		}
		e.raw(": " + p.Name + " ( -- )")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain && ans {
			e.line(1, "DECK-FILE CLOSE-FILE THROW")
		}
		e.raw(";")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw(e.j.id.Program + " BYE")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		if ans {
			e.line(depth, "CARD 80 DECK-FILE READ-LINE THROW 2DROP")
			return
		}
		e.line(depth, "CARD 80 EXPECT")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		dest := strings.TrimSuffix(t, " @")
		if num {
			return fmt.Sprintf("CARD %d + %d >NUM %s !", start, width, dest)
		}
		return fmt.Sprintf("CARD %d + %s %d CMOVE", start, dest, width)
	}
	return d
}
