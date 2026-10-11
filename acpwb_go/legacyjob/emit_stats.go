package legacyjob

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// statEmitter renders the job in the dataset-oriented languages: SAS, SPSS,
// Stata and the 4GL report writers (FOCUS, NOMAD, RAMIS). Like the SQL
// emitter, the deck is read whole, the change column and percentile ladder
// are computed per observation, and the report is a formatted listing. The
// dead routines become macros / programs / Dialogue Manager blocks, printed
// from the IR by a small statement printer so they vary page to page.
type statEmitter struct {
	j   *job
	k   string
	out []Line
	syn stmtSyntax
	lbl int
	lp  [3]string // label prefixes: else, end-if, repeat
}

// stmtSyntax is the little table the dead-code printer needs.
type stmtSyntax struct {
	cmt             string
	assign          func(l, r string) string
	ifOpen          func(c string) string
	ifElse, ifClose string
	doOpen          func(v, from, to string, step int) string
	doClose         string
	call            func(name string, args []string) string
	debug           func(n int) string
	v               func(name string) string
	idx             func(name string, ix []string) string
	intDiv, mod     string // fmt: a, b
	rel             map[string]string
	and, or         string
	upper           bool
}

func newStatEmitter(j *job) *statEmitter {
	e := &statEmitter{j: j, k: j.id.LangKey}
	e.lbl = j.x.between(0, 700)
	e.lp = [3]string{pick(j.x, []string{"ELS", "SKIP", "NOT", "ALT", "NO"}), pick(j.x, []string{"FI", "ENDIF", "JOIN", "CONT", "MERGE"}), pick(j.x, []string{"RPT", "LOOP", "AGAIN", "NXT", "CYC"})}
	e.syn = e.syntax()
	return e
}

func (e *statEmitter) code(t string) {
	for _, l := range strings.Split(t, "\n") {
		e.out = append(e.out, Line{Text: strings.TrimRight(l, " "), Code: true})
	}
}
func (e *statEmitter) codef(f string, a ...interface{}) { e.code(fmt.Sprintf(f, a...)) }
func (e *statEmitter) blank()                           { e.out = append(e.out, Line{Text: ""}) }

func (e *statEmitter) cmt(lines []string) {
	for _, l := range lines {
		switch e.k {
		case "sas":
			e.out = append(e.out, Line{Text: strings.TrimRight("* "+strings.ReplaceAll(l, ";", ",")+";", " ")})
		case "spss":
			e.out = append(e.out, Line{Text: strings.TrimRight("* "+strings.ReplaceAll(l, ".", ",")+".", " ")})
		default:
			e.out = append(e.out, Line{Text: strings.TrimRight(e.syn.cmt+l, " ")})
		}
	}
}

func (e *statEmitter) note(text string) {
	cs := e.j.cs
	cs.width = 68
	e.cmt(cs.para(text))
}

func (e *statEmitter) u(s string) string {
	if e.syn.upper {
		return strings.ToUpper(s)
	}
	return s
}

func (e *statEmitter) syntax() stmtSyntax {
	j := e.j
	year := j.in.Year
	switch e.k {
	case "sas":
		return stmtSyntax{cmt: "* ", upper: year < 1997,
			assign: func(l, r string) string { return l + " = " + r + ";" },
			ifOpen: func(c string) string { return "IF " + c + " THEN DO;" },
			ifElse: "END; ELSE DO;", ifClose: "END;",
			doOpen: func(v, f, t string, step int) string {
				if step < 0 {
					return fmt.Sprintf("DO %s = %s TO %s BY -1;", v, f, t)
				}
				return fmt.Sprintf("DO %s = %s TO %s;", v, f, t)
			},
			doClose: "END;",
			call:    func(n string, a []string) string { return "%" + n + "(" + strings.Join(a, ", ") + ");" },
			debug:   func(n int) string { return fmt.Sprintf("PUT ' DEBUG CHECKPOINT %04d';", n) },
			v:       func(n string) string { return n },
			idx:     func(n string, ix []string) string { return n + "[" + strings.Join(ix, ", ") + "]" },
			intDiv:  "INT(%s / %s)", mod: "MOD(%s, %s)", rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "^=", "gt": ">", "ge": ">="},
			and: " AND ", or: " OR "}
	case "spss":
		return stmtSyntax{cmt: "* ", upper: true,
			assign: func(l, r string) string { return "COMPUTE " + l + " = " + r + "." },
			ifOpen: func(c string) string { return "DO IF (" + c + ")." },
			ifElse: "ELSE.", ifClose: "END IF.",
			doOpen: func(v, f, t string, step int) string {
				if step < 0 {
					return fmt.Sprintf("LOOP %s = %s TO %s BY -1.", v, f, t)
				}
				return fmt.Sprintf("LOOP %s = %s TO %s.", v, f, t)
			},
			doClose: "END LOOP.",
			call:    func(n string, a []string) string { return "!" + n + " " + strings.Join(a, " ") + "." },
			debug:   func(n int) string { return fmt.Sprintf("PRINT / ' DEBUG CHECKPOINT %04d'.", n) },
			v:       func(n string) string { return "#" + n },
			idx:     func(n string, ix []string) string { return n + "(" + strings.Join(ix, ",") + ")" },
			intDiv:  "TRUNC(%s / %s)", mod: "MOD(%s, %s)", rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"},
			and: " AND ", or: " OR "}
	case "stata":
		return stmtSyntax{cmt: "* ",
			assign: func(l, r string) string { return "scalar " + l + " = " + r },
			ifOpen: func(c string) string { return "if " + c + " {" },
			ifElse: "}\nelse {", ifClose: "}",
			doOpen: func(v, f, t string, step int) string {
				if year < 2001 {
					return fmt.Sprintf("local %s = %s\nwhile `%s' <= %s {", v, f, v, t)
				}
				if step < 0 {
					return fmt.Sprintf("forvalues %s = %s(-1)%s {", v, f, t)
				}
				return fmt.Sprintf("forvalues %s = %s/%s {", v, f, t)
			},
			doClose: "}",
			call:    func(n string, a []string) string { return strings.TrimSpace(n + " " + strings.Join(a, " ")) },
			debug:   func(n int) string { return fmt.Sprintf("display \" DEBUG CHECKPOINT %04d\"", n) },
			v:       func(n string) string { return n },
			idx:     func(n string, ix []string) string { return n + "[" + strings.Join(ix, ", ") + "]" },
			intDiv:  "floor(%s / %s)", mod: "mod(%s, %s)", rel: cRelOps, and: " & ", or: " | "}
	default: // focus nomad ramis: Dialogue Manager
		return stmtSyntax{cmt: "-* ", upper: true,
			assign: func(l, r string) string { return "-SET " + l + " = " + r + ";" },
			ifOpen: func(c string) string { return "-IF NOT (" + c + ") GOTO {ELSE};" },
			ifElse: "-GOTO {END};\n-{ELSE}", ifClose: "-{END}",
			doOpen: func(v, f, t string, step int) string {
				if step < 0 {
					return fmt.Sprintf("-REPEAT {END} FOR %s FROM %s TO %s STEP -1", v, f, t)
				}
				return fmt.Sprintf("-REPEAT {END} FOR %s FROM %s TO %s", v, f, t)
			},
			doClose: "-{END}",
			call:    func(n string, a []string) string { return "-INCLUDE " + n },
			debug:   func(n int) string { return fmt.Sprintf("-TYPE DEBUG CHECKPOINT %04d", n) },
			v:       func(n string) string { return "&" + n },
			idx:     func(n string, ix []string) string { return "&" + n + "." + strings.Join(ix, ".") },
			intDiv:  "INT(%s / %s)", mod: "IMOD(%s, %s, 'I9')", rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"},
			and: " AND ", or: " OR "}
	}
}

// ── dead-code printer ──────────────────────────────────────────────────────

func (e *statEmitter) ex(x *Expr) string {
	s := e.syn
	switch x.K {
	case "var":
		return s.v(x.Name)
	case "int":
		return strconv.Itoa(x.Val)
	case "dlit":
		return x.S
	case "str":
		return "'" + x.S + "'"
	case "idx":
		var ix []string
		for _, a := range x.Args {
			ix = append(ix, e.ex(a))
		}
		return s.idx(x.Name, ix)
	case "bin":
		l, r := e.ex(x.Args[0]), e.ex(x.Args[1])
		if x.Op == "/" && !isDouble(x.Args[0]) && !isDouble(x.Args[1]) {
			return fmt.Sprintf(s.intDiv, l, r)
		}
		if x.Args[0].K == "bin" {
			l = "(" + l + ")"
		}
		if x.Args[1].K == "bin" {
			r = "(" + r + ")"
		}
		return l + " " + x.Op + " " + r
	case "call":
		if x.Name == "mod" {
			return fmt.Sprintf(s.mod, e.ex(x.Args[0]), e.ex(x.Args[1]))
		}
		return e.ex(x.Args[0])
	case "ladder":
		return "INT(" + e.ex(x.Args[0]) + " * " + e.ex(x.Args[1]) + ")"
	}
	return "0"
}

func (e *statEmitter) cond(c *Cond) string {
	if c.Op == "and" || c.Op == "or" {
		op := e.syn.and
		if c.Op == "or" {
			op = e.syn.or
		}
		return "(" + e.cond(c.A) + ")" + op + "(" + e.cond(c.B) + ")"
	}
	return e.ex(c.L) + " " + e.syn.rel[c.Op] + " " + e.ex(c.R)
}

func (e *statEmitter) emitLines(depth int, text string, lbl map[string]string) {
	for _, t := range strings.Split(text, "\n") {
		for k, v := range lbl {
			t = strings.ReplaceAll(t, "{"+k+"}", v)
		}
		ind := strings.Repeat("  ", depth)
		if strings.HasPrefix(t, "-") {
			ind = ""
		}
		e.code(ind + e.u(t))
	}
}

func (e *statEmitter) stmts(ss []*Stmt, depth int) {
	for _, s := range ss {
		switch s.K {
		case "comment":
			e.note(s.Text)
		case "blank":
			e.blank()
		case "assign":
			e.emitLines(depth, e.syn.assign(e.ex(s.L), e.ex(s.R)), nil)
		case "call":
			var args []string
			for _, a := range s.Args {
				args = append(args, e.ex(a))
			}
			e.emitLines(depth, e.syn.call(s.Name, args), nil)
		case "io":
			if s.Name == "debugmsg" && len(s.Args) > 0 {
				e.emitLines(depth, e.syn.debug(s.Args[0].Val), nil)
			}
		case "if":
			e.lbl++
			lbl := map[string]string{"ELSE": fmt.Sprintf("%s%03d", e.lp[0], e.lbl%1000), "END": fmt.Sprintf("%s%03d", e.lp[1], e.lbl%1000)}
			e.emitLines(depth, e.syn.ifOpen(e.cond(s.Cond)), lbl)
			e.stmts(s.Then, depth+1)
			if s.Else != nil {
				e.emitLines(depth, e.syn.ifElse, lbl)
				e.stmts(s.Else, depth+1)
			} else if strings.Contains(e.syn.ifElse, "{ELSE}") {
				e.emitLines(depth, "-{ELSE}", lbl)
			}
			e.emitLines(depth, e.syn.ifClose, lbl)
		case "do":
			e.lbl++
			lbl := map[string]string{"END": fmt.Sprintf("%s%03d", e.lp[2], e.lbl%1000)}
			e.emitLines(depth, e.syn.doOpen(e.syn.v(s.Var), e.ex(s.From), e.ex(s.To), s.Step), lbl)
			e.stmts(s.Body, depth+1)
			if e.k == "stata" && e.j.in.Year < 2001 {
				e.emitLines(depth+1, "local "+s.Var+" = `"+s.Var+"' + 1", nil)
			}
			e.emitLines(depth, e.syn.doClose, lbl)
		}
	}
}

// ── program ────────────────────────────────────────────────────────────────

func (e *statEmitter) emit() []Line {
	j := e.j
	e.cmt(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	e.cmt(j.historyLines())
	e.cmt(j.cs.box("Provenance", "Seal "+strings.ToUpper(j.in.Token)+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos")))
	e.cmt(j.prologue(j.p.Procs[0], nil, nil))
	switch e.k {
	case "sas":
		e.sas()
	case "spss":
		e.spss()
	case "stata":
		e.stata()
	case "nomad":
		e.nomad()
	default:
		e.focus()
	}
	if j.x.chance(0.4) {
		e.blank()
		e.note(j.g.comment(nil, "handoff", "todo", "misc"))
	}
	e.renameFields()
	return e.out
}

var statFieldRe = regexp.MustCompile(`(?i)\b(mname|bname|bsln|curr|delt|medn|nmet|nben|peer|recid)\b`)

// renameFields gives the dataset's variables the page's naming convention
// (the walker languages get theirs from the naming scheme directly).
func (e *statEmitter) renameFields() {
	nm := e.j.nm
	names := map[string]string{}
	for _, role := range []string{"mname", "bname", "bsln", "curr", "delt", "medn", "nmet", "nben", "peer", "recid"} {
		n := strings.Trim(nm.get(role), "#&$.")
		n = strings.Map(func(r rune) rune {
			if r == '-' || r == '.' {
				return '_'
			}
			return r
		}, n)
		if n == "" || strings.ContainsAny(n, " ") {
			n = role
		}
		names[role] = n
	}
	for i, l := range e.out {
		if !l.Code {
			continue
		}
		e.out[i].Text = statFieldRe.ReplaceAllStringFunc(l.Text, func(w string) string {
			n := names[strings.ToLower(w)]
			if w == strings.ToUpper(w) {
				return strings.ToUpper(n)
			}
			return strings.ToLower(n)
		})
	}
}

// dead emits every dead routine with the language's wrapper around it.
func (e *statEmitter) dead(open func(p *Proc), close func(p *Proc)) {
	for _, p := range e.j.p.Procs[1:] {
		if !p.Dead {
			continue
		}
		e.blank()
		e.cmt(e.j.prologue(p, nil, nil))
		open(p)
		e.stmts(p.Body, 1)
		close(p)
	}
}

func ladderList() string { return strings.Join(ladderLit[:], " ") }

// ── SAS ────────────────────────────────────────────────────────────────────

func (e *statEmitter) sas() {
	j := e.j
	x := j.x
	r := j.rep
	tok := strings.ToUpper(j.in.Token)
	prog := strings.ToLower(j.id.Program)
	mainframe := j.in.Year < 1995 || x.chance(0.3)
	e.blank()
	opts := sample(x, []string{"NODATE", "NONUMBER", "MPRINT", "SYMBOLGEN", "NOSOURCE2", "ERRORS=5", "OBS=MAX", "FMTERR", "MLOGIC"}, x.between(2, 5))
	e.code(e.u("OPTIONS NOCENTER LS=132 PS=66 " + strings.Join(opts, " ") + ";"))
	e.code(e.u("%LET seal = %STR('" + tok + "');"))
	if mainframe {
		e.code(e.u("FILENAME deck 'ACPWB.RSCH.DECK(" + j.id.Program + ")' DISP=SHR;"))
	} else {
		e.code("FILENAME deck '/rsch/deck/" + prog + ".dat';")
	}
	sym := "CALL SYMPUTX('nmet', nmet); CALL SYMPUTX('nben', nben);"
	if j.in.Year < 2002 {
		sym = "CALL SYMPUT('nmet', LEFT(PUT(nmet, 3.))); CALL SYMPUT('nben', LEFT(PUT(nben, 3.)));"
	}
	e.blank()
	e.note("Header cards: client, record, peer group, counts. " + j.g.comment(nil, "deck", "formats"))
	for _, l := range []string{
		"DATA work.hdr;",
		"  INFILE deck OBS=3 TRUNCOVER LRECL=80;",
		"  INPUT #1 @1 org $CHAR40. @41 recid $CHAR8.",
		"        #2 @1 peer $CHAR60.",
		"        #3 @1 nmet 3. @4 nben 3.;",
		"  " + sym,
		"RUN;"} {
		e.code(e.u(l))
	}
	cols := strings.ToLower(strings.Join(ladderCol[:], " "))
	readMet := pick(x, []string{
		"  INPUT @1 mname $CHAR40. @41 bsln 6. @47 curr 6.;",
		"  INPUT mname $ 1-40 bsln 41-46 curr 47-52;",
		"  INPUT (mname bsln curr) ($CHAR40. 6. 6.);",
	})
	readBen := pick(x, []string{
		"  INPUT @1 bname $CHAR40. @41 medn 7.;",
		"  INPUT bname $ 1-40 medn 41-47;",
		"  INPUT (bname medn) ($CHAR40. 7.);",
	})
	sqlDelta := j.in.Year >= 1990 && x.chance(0.35) // PROC SQL arrived with 6.06
	e.blank()
	e.note("Engagement metric cards. " + j.g.comment(nil, "data", "units", "warnings"))
	e.code(e.u("DATA work.met;"))
	e.code(e.u("  INFILE deck FIRSTOBS=4 OBS=%EVAL(3 + &nmet) TRUNCOVER LRECL=80;"))
	e.code(e.u(readMet))
	if !sqlDelta {
		e.code(e.u(pick(x, []string{"  delt = curr - bsln;", "  delt = SUM(curr, -bsln);", "  delt = curr + (-1 * bsln);"})))
	}
	e.code(e.u("RUN;"))
	if sqlDelta {
		e.code(e.u("PROC SQL NOPRINT;"))
		e.code(e.u("  CREATE TABLE work.met AS"))
		e.code(e.u("  SELECT mname, bsln, curr, curr - bsln AS delt"))
		e.code(e.u("    FROM work.met;"))
		e.code(e.u("QUIT;"))
	}
	e.blank()
	e.note("Benchmark medians and the ladder. Truncate, never round (audit 83-07). " + j.g.comment(nil, "magic", "lore"))
	switch x.intn(3) {
	case 0:
		for _, l := range []string{
			"DATA work.ben;",
			"  INFILE deck FIRSTOBS=%EVAL(4 + &nmet) TRUNCOVER LRECL=80;",
			readBen,
			"  ARRAY mult[8] _TEMPORARY_ (" + ladderList() + ");",
			"  ARRAY pct[8] " + cols + ";",
			"  DO k = 1 TO 8;",
			"    pct[k] = INT(medn * mult[k]);",
			"  END;",
			"  DROP k;",
			"RUN;"} {
			e.code(e.u(l))
		}
	case 1:
		e.code(e.u("DATA work.ben;"))
		e.code(e.u("  INFILE deck FIRSTOBS=%EVAL(4 + &nmet) TRUNCOVER LRECL=80;"))
		e.code(e.u(readBen))
		for k, c := range ladderCol {
			e.codef(e.u("  %s = INT(medn * %s);"), strings.ToLower(c), ladderLit[k])
		}
		e.code(e.u("RUN;"))
	default:
		e.note("Ladder as a dataset, joined back by PROC SQL and transposed. " + j.g.comment(nil, "conversions", "formats"))
		e.code(e.u("DATA work.ladder;"))
		e.code(e.u("  INPUT rung mult;"))
		e.code(e.u("  DATALINES;"))
		for k := range ladderLit {
			e.codef("%d %s", k+1, ladderLit[k])
		}
		e.code(";")
		e.code(e.u("RUN;"))
		e.code(e.u("DATA work.bin;"))
		e.code(e.u("  INFILE deck FIRSTOBS=%EVAL(4 + &nmet) TRUNCOVER LRECL=80;"))
		e.code(e.u(readBen))
		e.code(e.u("  seq = _N_;"))
		e.code(e.u("RUN;"))
		e.code(e.u("PROC SQL NOPRINT;"))
		e.code(e.u("  CREATE TABLE work.long AS"))
		e.code(e.u("  SELECT b.seq, b.bname, l.rung, INT(b.medn * l.mult) AS pct"))
		e.code(e.u("    FROM work.bin b, work.ladder l"))
		e.code(e.u("   ORDER BY b.seq, l.rung;"))
		e.code(e.u("QUIT;"))
		e.code(e.u("PROC TRANSPOSE DATA=work.long OUT=work.ben(DROP=_NAME_) PREFIX=r;"))
		e.code(e.u("  BY seq bname;"))
		e.code(e.u("  ID rung;"))
		e.code(e.u("  VAR pct;"))
		e.code(e.u("RUN;"))
		e.code(e.u("DATA work.ben;"))
		e.code(e.u("  SET work.ben;"))
		e.code(e.u("  RENAME " + func() string {
			var r []string
			for k, c := range ladderCol {
				r = append(r, fmt.Sprintf("r%d=%s", k+1, strings.ToLower(c)))
			}
			return strings.Join(r, " ")
		}() + ";"))
		e.code(e.u("RUN;"))
	}
	e.blank()
	e.note("Signed change column. " + j.g.comment(nil, "formats", "printing"))
	e.code(e.u("PROC FORMAT;"))
	e.code(e.u("  PICTURE sgn low-<0 = '00000009' (PREFIX='-')"))
	e.code(e.u("                0 = '9'"))
	e.code(e.u("          0<-high = '00000009' (PREFIX='+');"))
	e.code(e.u("RUN;"))
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.code(e.u("DATA _NULL_;"))
	e.code(e.u("  FILE PRINT NOTITLES;"))
	e.code(e.u("  SET work.hdr;"))
	e.codef("  PUT %s recid $CHAR8. %s;", q(r.HdrLeft), q(r.HdrMid+"    1"))
	e.codef("  PUT %s org $CHAR40.;", q(r.Title1))
	e.codef("  PUT %s;", q(r.ColHdr1))
	e.code(e.u("RUN;"))
	if j.in.Year >= 1992 && x.chance(0.35) {
		e.code(e.u("PROC REPORT DATA=work.met NOWINDOWS NOHEADER SPACING=0;"))
		e.code(e.u("  COLUMN mname bsln curr delt;"))
		e.code(e.u("  DEFINE mname / DISPLAY WIDTH=41 FORMAT=$CHAR41.;"))
		e.code(e.u("  DEFINE bsln  / DISPLAY WIDTH=9 FORMAT=9.;"))
		e.code(e.u("  DEFINE curr  / DISPLAY WIDTH=9 FORMAT=9.;"))
		e.code(e.u("  DEFINE delt  / DISPLAY WIDTH=9 FORMAT=sgn9.;"))
		e.code(e.u("RUN;"))
	} else {
		e.code(e.u("DATA _NULL_;"))
		e.code(e.u("  FILE PRINT NOTITLES MOD;"))
		e.code(e.u("  SET work.met;"))
		e.code(e.u(pick(x, []string{"  PUT @2 mname $CHAR40. bsln 9. curr 9. delt sgn9.;", "  PUT @2 mname $CHAR40. @42 bsln 9. @51 curr 9. @60 delt sgn9.;", "  PUT ' ' mname $CHAR40. (bsln curr) (9.) delt sgn9.;"})))
		e.code(e.u("RUN;"))
	}
	e.code(e.u("DATA _NULL_;"))
	e.code(e.u("  FILE PRINT NOTITLES MOD;"))
	e.code(e.u("  SET work.hdr;"))
	e.codef("  PUT %s peer $CHAR60.;", q(r.Title2))
	e.codef("  PUT %s;", q(r.ColHdr2))
	e.code(e.u("RUN;"))
	e.code(e.u("DATA _NULL_;"))
	e.code(e.u("  FILE PRINT NOTITLES MOD;"))
	e.code(e.u("  SET work.ben END=last;"))
	e.code(e.u("  PUT @2 bname $CHAR40. (" + strings.ToLower(strings.Join(ladderCol[:], " ")) + ") (9.);"))
	e.code(e.u("  IF last THEN PUT '0RECORDS PROCESSED' \"&nmet\" ' METRICS' \"&nben\" ' BENCHMARKS   SEAL ' &seal;"))
	e.code(e.u("RUN;"))
	e.dead(func(p *Proc) {
		e.code("%MACRO " + p.Name + "(" + strings.Join(p.Params, ", ") + ");")
		e.code(e.u("  DATA _NULL_;"))
		for _, q := range p.Params {
			e.code("    " + q + " = &" + q + ";")
		}
		for _, l := range p.Locals {
			if len(l.Dims) > 0 {
				e.codef("    ARRAY %s[%d] _TEMPORARY_;", l.Name, l.Dims[0])
			}
		}
	}, func(p *Proc) {
		e.code(e.u("  RUN;"))
		e.code("%MEND " + p.Name + ";")
	})
	e.blank()
	e.code(e.u(pick(x, []string{"PROC DATASETS LIBRARY=work NOLIST; DELETE hdr met ben; QUIT;", "PROC DATASETS LIB=work KILL NOLIST; QUIT;", "PROC DELETE DATA=work.hdr work.met work.ben; RUN;"})))
}

// ── SPSS ───────────────────────────────────────────────────────────────────

func (e *statEmitter) spss() {
	j := e.j
	x := j.x
	r := j.rep
	tok := strings.ToUpper(j.in.Token)
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	e.blank()
	e.code("SET WIDTH=132 LENGTH=66 PRINTBACK=NO HEADER=NO.")
	e.codef("FILE HANDLE DECK /NAME='ACPWB.RSCH.DECK(%s)'.", j.id.Program)
	e.code("DATA LIST FILE=DECK FIXED /CARD 1-80 (A).")
	e.code("COMPUTE SEQ = $CASENUM.")
	e.code("STRING SEAL (A8).")
	e.codef("COMPUTE SEAL = '%s'.", tok)
	e.blank()
	e.note("Counts from the third card, carried down every case with LEAVE. " + j.g.comment(nil, "deck", "formats"))
	e.code("DO IF (SEQ EQ 3).")
	e.code("+  COMPUTE NMET = NUMBER(SUBSTR(CARD,1,3),F3).")
	e.code("+  COMPUTE NBEN = NUMBER(SUBSTR(CARD,4,3),F3).")
	e.code("END IF.")
	e.code("LEAVE NMET NBEN.")
	e.blank()
	e.note("Metric and benchmark cards share the file; SEQ says which is which. " + j.g.comment(nil, "data", "units"))
	e.code("STRING MNAME BNAME (A40).")
	e.code("DO IF (SEQ GT 3 AND SEQ LE 3 + NMET).")
	e.code("+  COMPUTE MNAME = SUBSTR(CARD,1,40).")
	e.code("+  COMPUTE BSLN = NUMBER(SUBSTR(CARD,41,6),F6).")
	e.code("+  COMPUTE CURR = NUMBER(SUBSTR(CARD,47,6),F6).")
	e.code(pick(x, []string{"+  COMPUTE DELT = CURR - BSLN.", "+  COMPUTE DELT = SUM(CURR, -BSLN).", "+  COMPUTE DELT = CURR + (-1 * BSLN)."}))
	e.code("ELSE IF (SEQ GT 3 + NMET).")
	e.code("+  COMPUTE BNAME = SUBSTR(CARD,1,40).")
	e.code("+  COMPUTE MEDN = NUMBER(SUBSTR(CARD,41,7),F7).")
	e.code("END IF.")
	e.blank()
	e.note("The ladder. Ratified 1981-02-17. " + j.g.comment(nil, "magic", "lore", "memos"))
	switch x.intn(3) {
	case 0:
		e.code("VECTOR MULT(8) / PCT(8).")
		e.code("DO REPEAT M = MULT1 TO MULT8 / V = " + ladderList() + ".")
		e.code("+  COMPUTE M = V.")
		e.code("END REPEAT.")
		e.code("DO IF (SEQ GT 3 + NMET).")
		e.code("+  LOOP #K = 1 TO 8.")
		e.code("+    COMPUTE PCT(#K) = TRUNC(MEDN * MULT(#K)).")
		e.code("+  END LOOP.")
		e.code("END IF.")
	case 1:
		e.code("DO REPEAT P = PCT1 TO PCT8 / V = " + ladderList() + ".")
		e.code("+  IF (SEQ GT 3 + NMET) P = TRUNC(MEDN * V).")
		e.code("END REPEAT.")
	default:
		for k, v := range ladderLit {
			e.codef("IF (SEQ GT 3 + NMET) PCT%d = TRUNC(MEDN * %s).", k+1, v)
		}
	}
	e.code("FORMATS BSLN CURR DELT MEDN PCT1 TO PCT8 (F9.0).")
	e.code("EXECUTE.")
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.code("DO IF (SEQ EQ 1).")
	e.codef("+  PRINT / %s SUBSTR(CARD,41,8) (A8) / %s SUBSTR(CARD,1,40) (A40) / %s.", q(r.HdrLeft), q(r.Title1), q(r.ColHdr1))
	e.code("END IF.")
	if x.chance(0.4) {
		e.code("TEMPORARY.")
		e.code("SELECT IF (SEQ GT 3 AND SEQ LE 3 + NMET).")
		e.code(pick(x, []string{"LIST VARIABLES=MNAME BSLN CURR DELT /FORMAT=NUMBERED.", "LIST MNAME BSLN CURR DELT."}))
	} else {
		e.code("DO IF (SEQ GT 3 AND SEQ LE 3 + NMET).")
		e.code("+  PRINT / ' ' MNAME (A40) BSLN CURR DELT (3F9.0).")
		e.code("END IF.")
	}
	e.code("DO IF (SEQ EQ 3 + NMET + 1).")
	e.codef("+  PRINT / %s / %s.", q(r.Title2), q(r.ColHdr2))
	e.code("END IF.")
	e.code("DO IF (SEQ GT 3 + NMET).")
	e.code("+  PRINT / ' ' BNAME (A40) PCT1 TO PCT8 (8F9.0).")
	e.code("END IF.")
	e.code("EXECUTE.")
	e.dead(func(p *Proc) {
		var args []string
		for _, q := range p.Params {
			args = append(args, strings.ToUpper(q)+" = !TOKENS(1)")
		}
		e.code("DEFINE !" + p.Name + " (" + strings.Join(args, " / ") + ")")
	}, func(p *Proc) {
		e.code("!ENDDEFINE.")
	})
}

// ── Stata ──────────────────────────────────────────────────────────────────

func (e *statEmitter) stata() {
	j := e.j
	x := j.x
	r := j.rep
	tok := strings.ToUpper(j.in.Token)
	ver := 8
	switch y := j.in.Year; {
	case y < 1996:
		ver = 3
	case y < 2001:
		ver = 6
	case y < 2009:
		ver = 9
	case y < 2017:
		ver = 13
	default:
		ver = 15
	}
	q := func(s string) string { return "\"" + strings.ReplaceAll(s, "\"", "'") + "\"" }
	e.blank()
	e.codef("version %d", ver)
	e.code("clear")
	e.code("set more off")
	e.codef("local seal \"%s\"", tok)
	e.blank()
	e.note("Read the deck as card images. " + j.g.comment(nil, "deck", "formats"))
	e.code("infix str80 card 1-80 using \"deck.dat\", clear")
	e.code("generate seq = _n")
	e.code("local nmet = real(substr(card[3], 1, 3))")
	e.code("local nben = real(substr(card[3], 4, 3))")
	e.code("local first = 4 + `nmet'")
	e.blank()
	e.note("Metrics. " + j.g.comment(nil, "data", "units", "warnings"))
	e.code("generate str40 mname = substr(card, 1, 40) if seq > 3 & seq < `first'")
	e.code("generate bsln = real(substr(card, 41, 6)) if seq > 3 & seq < `first'")
	e.code("generate curr = real(substr(card, 47, 6)) if seq > 3 & seq < `first'")
	e.code(pick(x, []string{"generate delt = curr - bsln", "generate long delt = curr - bsln", "egen delt = rowtotal(curr), missing\nreplace delt = delt - bsln"}))
	e.code("generate str40 bname = substr(card, 1, 40) if seq >= `first'")
	e.code("generate medn = real(substr(card, 41, 7)) if seq >= `first'")
	e.blank()
	e.note("The ladder. Truncate, never round. " + j.g.comment(nil, "magic", "lore"))
	switch v := x.intn(3); {
	case v == 0 && j.in.Year >= 2001:
		e.codef("local rungs %s", ladderList())
		e.code("local k = 0")
		e.code("foreach m of local rungs {")
		e.code("    local ++k")
		e.code("    generate p`k' = floor(medn * `m')")
		e.code("}")
	case v == 1:
		for k, lit := range ladderLit {
			e.codef("generate p%d = floor(medn * %s)", k+1, lit)
		}
	default:
		for i, lit := range ladderLit {
			e.codef("local m%d = %s", i+1, lit)
		}
		if j.in.Year < 2001 {
			e.code("for num 1/8: generate pX = floor(medn * `mX')")
		} else {
			e.code("forvalues k = 1/8 {")
			e.code("    generate p`k' = floor(medn * `m`k'')")
			e.code("}")
		}
	}
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.codef("display %s substr(card[1], 41, 8) %s", q(r.HdrLeft), q(r.HdrMid+"    1"))
	e.codef("display %s substr(card[1], 1, 40)", q(r.Title1))
	e.codef("display %s", q(r.ColHdr1))
	loop := func(from, to string, body string) {
		if j.in.Year < 2001 {
			e.codef("local i = %s", from)
			e.codef("while `i' <= %s {", to)
			e.code("    " + body)
			e.code("    local i = `i' + 1")
			e.code("}")
			return
		}
		e.codef("forvalues i = %s/%s {", from, to)
		e.code("    " + body)
		e.code("}")
	}
	loop("4", "`=`first'-1'", "display \" \" %-40s mname[`i'] %9.0f bsln[`i'] %9.0f curr[`i'] %9s cond(delt[`i'] > 0, \"+\", \"\") + string(delt[`i'])")
	e.codef("display %s substr(card[2], 1, 60)", q(r.Title2))
	e.codef("display %s", q(r.ColHdr2))
	loop("`first'", "`=_N'", "display \" \" %-40s bname[`i'] %9.0f p1[`i'] p2[`i'] p3[`i'] p4[`i'] p5[`i'] p6[`i'] p7[`i'] p8[`i']")
	e.code("display \"0RECORDS PROCESSED\" %6.0f `nmet' \" METRICS\" %6.0f `nben' \" BENCHMARKS   SEAL `seal'\"")
	e.dead(func(p *Proc) {
		e.code("capture program drop " + p.Name)
		e.code("program define " + p.Name)
		if len(p.Params) > 0 {
			e.code("    args " + strings.Join(p.Params, " "))
		}
	}, func(p *Proc) {
		e.code("end")
	})
}

// ── FOCUS / RAMIS ──────────────────────────────────────────────────────────

func (e *statEmitter) focus() {
	j := e.j
	x := j.x
	r := j.rep
	verb := pick(x, []string{"PRINT", "PRINT", "LIST"})
	sel := func(c string) string { return "IF " + c }
	if j.in.Year >= 1988 && x.chance(0.6) {
		sel = func(c string) string { return "WHERE " + c }
	}
	compute := x.chance(0.4) // ladder as COMPUTEs inside the request
	tableOpts := sample(x, []string{"ON TABLE SET PAGE-NUM OFF", "ON TABLE SET LINES 66", "ON TABLE NOTOTAL", "ON TABLE SET STYLE *", "ON TABLE SET SPACES 1"}, x.between(1, 3))
	if j.in.Year < 2000 {
		tableOpts = sample(x, []string{"ON TABLE SET PAGE-NUM OFF", "ON TABLE SET LINES 66", "ON TABLE NOTOTAL"}, x.between(1, 2))
	}
	tok := strings.ToUpper(j.in.Token)
	ramis := e.k == "ramis"
	q := func(s string) string { return strings.ReplaceAll(s, "\"", "'") }
	e.blank()
	e.code("-SET &ECHO = 'OFF';")
	e.codef("-SET &SEAL = '%s';", tok)
	e.codef("-SET &RECID = '%s';", strings.ToUpper(j.in.RecordID))
	if ramis {
		e.code("FILEDEF DECK DISK ACPWB RSCHDECK A")
	} else {
		e.code("DYNAM ALLOC FILE DECK DA ACPWB.RSCH.DECK(" + j.id.Program + ") SHR REUSE")
	}
	e.code("-READ DECK &ORG.A40. &RID.A8.")
	e.code("-READ DECK &PEER.A60.")
	e.code("-READ DECK &NMET.I3. &NBEN.I3.")
	if x.chance(0.6) {
		lim := pick(x, []int{40, 50, 64, 99, 100})
		e.codef("-IF &NMET GT %d OR &NBEN GT %d THEN GOTO TOOMANY;", lim, lim)
	}
	e.blank()
	e.note("Master file description: one 80-column card per segment instance. " + j.g.comment(nil, "deck", "formats"))
	e.code("-* FILE=DECK, SUFFIX=FIX")
	e.code("-*   SEGNAME=CARDS, SEGTYPE=S0")
	e.code("-*     FIELD=CARD, ALIAS=C, FORMAT=A80, ACTUAL=A80, $")
	e.blank()
	e.note("Computed fields. The ladder truncates; the committee was clear. " + j.g.comment(nil, "magic", "lore", "memos"))
	kw := "DEFINE FILE DECK"
	if ramis {
		kw = "DEFINE FILE DECK ADD"
	}
	e.code(kw)
	e.code(" SEQ/I5      = LAST SEQ + 1;")
	e.code(" MNAME/A40   = SUBSTR(80, CARD, 1, 40, 40, 'A40');")
	e.code(" BSLN/I6     = EDIT(SUBSTR(80, CARD, 41, 46, 6, 'A6'));")
	e.code(" CURR/I6     = EDIT(SUBSTR(80, CARD, 47, 52, 6, 'A6'));")
	e.code(" DELT/I7S    = CURR - BSLN;")
	e.code(" MEDN/I7     = EDIT(SUBSTR(80, CARD, 41, 47, 7, 'A7'));")
	if !compute {
		for i, c := range ladderCol {
			e.codef(" %-11s = INT(MEDN * %s);", c+"/I9", ladderLit[i])
		}
	}
	e.code("END")
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.code("TABLE FILE DECK")
	e.code("HEADING")
	e.codef("\"%s<&RECID%s\"", q(r.HdrLeft), q(r.HdrMid+"    1"))
	e.codef("\"%s<&ORG\"", q(r.Title1))
	e.codef("\"%s\"", q(r.ColHdr1))
	e.code(verb + " MNAME BSLN CURR DELT")
	if x.chance(0.5) {
		e.code(sel("SEQ GT 3 AND SEQ LE &NMET + 3"))
	} else {
		e.code(sel("SEQ GT 3"))
		e.code(sel("SEQ LE &NMET + 3"))
	}
	for _, o := range tableOpts {
		e.code(o)
	}
	e.code("END")
	e.code("TABLE FILE DECK")
	e.code("HEADING")
	e.codef("\"%s<&PEER\"", q(r.Title2))
	e.codef("\"%s\"", q(r.ColHdr2))
	e.code(verb + " BNAME")
	if compute {
		for i, c := range ladderCol {
			e.codef("COMPUTE %s/I9 = INT(MEDN * %s);", c, ladderLit[i])
		}
	} else {
		e.code("      " + strings.Join(ladderCol[:], " "))
	}
	e.code(sel("SEQ GT &NMET + 3"))
	e.code("FOOTING")
	e.code("\"0RECORDS PROCESSED <&NMET METRICS <&NBEN BENCHMARKS   SEAL <&SEAL\"")
	e.code("END")
	e.code("-RUN")
	e.code("-GOTO DONE")
	e.dead(func(p *Proc) {
		e.code("-" + strings.ToUpper(p.Name))
	}, func(p *Proc) {
		e.code("-GOTO DONE")
	})
	if strings.Contains(strings.Join(func() []string {
		var t []string
		for _, l := range e.out {
			t = append(t, l.Text)
		}
		return t
	}(), "\n"), "GOTO TOOMANY") {
		e.code("-TOOMANY")
		e.code("-TYPE DECK HAS MORE CARDS THAN THE TABLES HOLD: &NMET / &NBEN")
		e.code("-EXIT")
	}
	e.code("-DONE")
	e.code("-EXIT")
}

// ── NOMAD ──────────────────────────────────────────────────────────────────

func (e *statEmitter) nomad() {
	j := e.j
	x := j.x
	r := j.rep
	verb := pick(x, []string{"LIST", "PRINT", "LIST"})
	trunc := pick(x, []string{"TRUNC", "INTEGER", "FLOOR"})
	tok := strings.ToUpper(j.in.Token)
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	e.blank()
	e.code("SCHEMA RSCHDECK;")
	e.code("  MASTER DECK INSERT=FIFO;")
	e.code("    ITEM SEQ   AS 99999;")
	e.code("    ITEM CARD  AS A80;")
	e.code("    DEFINE MNAME AS A40     EXPR=SUBSTR(CARD,1,40);")
	e.code("    DEFINE BSLN  AS 999999  EXPR=NUMBER(SUBSTR(CARD,41,6));")
	e.code("    DEFINE CURR  AS 999999  EXPR=NUMBER(SUBSTR(CARD,47,6));")
	e.code("    DEFINE DELT  AS S999999 EXPR=CURR - BSLN;")
	e.code("    DEFINE MEDN  AS 9999999 EXPR=NUMBER(SUBSTR(CARD,41,7));")
	for i, c := range ladderCol {
		e.codef("    DEFINE %-5s AS 999999999 EXPR="+trunc+"(MEDN * %s);", c, ladderLit[i])
	}
	e.code("END SCHEMA;")
	e.blank()
	e.codef("&SEAL = '%s';", tok)
	e.code("LOAD DECK FROM FILE DECKIN FORMAT CARD;")
	e.code("&NMET = NUMBER(SUBSTR(DECK(3).CARD,1,3));")
	e.code("&NBEN = NUMBER(SUBSTR(DECK(3).CARD,4,3));")
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.codef(verb+" HEADING %s || SUBSTR(DECK(1).CARD,41,8) / %s || SUBSTR(DECK(1).CARD,1,40) / %s", q(r.HdrLeft), q(r.Title1), q(r.ColHdr1))
	e.code("     MNAME BSLN CURR DELT WHERE SEQ > 3 AND SEQ <= &NMET + 3;")
	e.codef(verb+" HEADING %s || SUBSTR(DECK(2).CARD,1,60) / %s", q(r.Title2), q(r.ColHdr2))
	e.code("     BNAME=MNAME " + strings.Join(ladderCol[:], " ") + " WHERE SEQ > &NMET + 3")
	e.code("     FOOTING '0RECORDS PROCESSED ' || &NMET || ' METRICS ' || &NBEN || ' BENCHMARKS   SEAL ' || &SEAL;")
	e.dead(func(p *Proc) {
		e.code("PROCEDURE " + strings.ToUpper(p.Name) + ";")
	}, func(p *Proc) {
		e.code("END PROCEDURE;")
	})
}
