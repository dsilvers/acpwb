package legacyjob

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Maintenance layers: what decades of hands do to a source file. Applied to
// the finished listing of any language, never touching live statements:
//
//   - a commented-out "previous version" of a stretch of real code, with the
//     constants it used back then, kept "for audit" (sometimes as Y2K work)
//   - conditional compilation: C #ifdef blocks, COBOL debugging lines (D in
//     column 7), DEC FORTRAN D lines
//   - one routine visibly reformatted by a later maintainer

type cmtSyntax struct {
	pre, post string
	col       int // fixed-format: overwrite this column with pre[0] instead
}

func commentSyntax(key string) (cmtSyntax, bool) {
	switch key {
	case "f4", "f77", "gf":
		return cmtSyntax{pre: "C", col: 1}, true
	case "cobol":
		return cmtSyntax{pre: "*", col: 7}, true
	case "basic", "atari", "apl", "rpg", "py2":
		return cmtSyntax{}, false // line numbers / layout carry meaning
	}
	m := map[string]cmtSyntax{
		"c": {pre: "/* ", post: " */"}, "cbtr": {pre: "/* ", post: " */"}, "perl": {pre: "# "}, "perl4": {pre: "# "},
		"php": {pre: "// "}, "csharp": {pre: "// "}, "vbnet": {pre: "' "}, "vb": {pre: "' "}, "vba": {pre: "' "}, "asp": {pre: "' "},
		"ada": {pre: "-- "}, "pascal": {pre: "{ ", post: " }"}, "pascalbtr": {pre: "{ ", post: " }"}, "modula2": {pre: "(* ", post: " *)"},
		"pli": {pre: "/* ", post: " */"}, "pal": {pre: "; "}, "dbase": {pre: "* "}, "clipper": {pre: "// "}, "foxpro": {pre: "* "},
		"natural": {pre: "* "}, "informix": {pre: "# "}, "plsql": {pre: "-- "}, "sql": {pre: "-- "}, "db2sql": {pre: "-- "},
		"sqlds": {pre: "-- "}, "sqlplus": {pre: "-- "}, "tsql": {pre: "-- "}, "sqlpl": {pre: "-- "}, "pick": {pre: "* "},
		"mumps": {pre: " ;"}, "rexx": {pre: "/* ", post: " */"}, "clist": {pre: "/* ", post: " */"}, "cl400": {pre: "/* ", post: " */"},
		"ezt": {pre: "* "}, "abap": {pre: "\" "}, "r": {pre: "# "}, "splus": {pre: "# "}, "matlab": {pre: "% "}, "octave": {pre: "% "},
		"awk": {pre: "# "}, "ksh": {pre: "# "}, "tcl": {pre: "# "}, "bc": {pre: "/* ", post: " */"}, "dc": {pre: "# "},
		"cfml": {pre: "<!--- ", post: " --->"}, "lisp": {pre: ";; "}, "prolog": {pre: "% "}, "forth": {pre: "\\ "},
		"smalltalk": {pre: "\"", post: "\""}, "sas": {pre: "* ", post: ";"}, "spss": {pre: "* ", post: "."}, "stata": {pre: "* "},
		"focus": {pre: "-* "}, "nomad": {pre: "-* "}, "ramis": {pre: "-* "},
	}
	s, ok := m[key]
	return s, ok
}

// commentOut turns one source line into a comment of the language.
func (cs cmtSyntax) commentOut(t string) string {
	if cs.col > 0 {
		r := []rune(t)
		for len(r) < cs.col {
			r = append(r, ' ')
		}
		r[cs.col-1] = rune(cs.pre[0])
		return string(r)
	}
	if cs.post == "." { // SPSS: only a period at the end of a line closes the comment
		return strings.TrimRight(cs.pre+strings.TrimRight(t, ". "), " ") + "."
	}
	if end := strings.TrimSpace(cs.post); end != "" {
		t = strings.ReplaceAll(t, end, "'")
		if cs.pre == "\"" {
			t = strings.ReplaceAll(t, "\"", "'")
		}
	}
	return strings.TrimRight(cs.pre+t+cs.post, " ")
}

// cosmetic-only indentation: the routine can be re-indented safely
var reindentOK = map[string]bool{"c": true, "cbtr": true, "perl": true, "perl4": true, "php": true, "csharp": true, "awk": true,
	"tcl": true, "r": true, "splus": true, "matlab": true, "octave": true, "pascal": true, "pascalbtr": true, "ada": true,
	"modula2": true, "vb": true, "vba": true, "vbnet": true, "asp": true, "cfml": true, "plsql": true, "informix": true,
	"ksh": true, "lisp": true, "smalltalk": true, "rexx": true, "pli": true, "natural": true, "clipper": true, "foxpro": true, "dbase": true}

var numLit = regexp.MustCompile(`\b\d+(\.\d+)?\b`)

func maintain(j *job, src []Line) []Line {
	x := j.x.fork("maintenance")
	key := j.id.LangKey
	cs, ok := commentSyntax(key)
	if ok && x.chance(0.55) {
		src = oldVersionBlock(j, x, src, cs)
	}
	src = condCompile(j, x, src)
	if reindentOK[key] && x.chance(0.35) {
		src = reformatRoutine(j, x, src, cs)
	}
	return src
}

// routineStarts returns the line indexes where each routine's prologue
// begins (the "Routine:" field).
func routineStarts(src []Line) []int {
	var out []int
	for i, l := range src {
		if !l.Code && strings.Contains(strings.ToUpper(l.Text), "ROUTINE:") {
			out = append(out, i)
		}
	}
	return out
}

// oldVersionBlock copies a window of real statements, ages their constants
// and inserts them, commented out, ahead of the live code.
func oldVersionBlock(j *job, x *rnd, src []Line, cs cmtSyntax) []Line {
	var code []int
	for i, l := range src {
		if l.Code && len(strings.TrimSpace(l.Text)) > 3 {
			code = append(code, i)
		}
	}
	if len(code) < 30 {
		return src
	}
	n := x.between(4, 14)
	k := x.between(len(code)/5, len(code)-n-1)
	from, to := code[k], code[k+n]
	who := j.g.coder()
	yr := x.between(maxi(1975, j.in.Year-25), j.in.Year)
	title := pick(x, []string{
		"Previous version (rel %d.%d, %d), retained for audit.",
		"Superseded %d.%d logic from %d. Do not delete: cited in a client deliverable.",
		"Original %d.%d code, %d. Kept until the parallel run is signed off.",
		"Old code, rel %d.%d (%d). Commented out, not removed, per records retention.",
	})
	head := fmt.Sprintf(title, x.between(1, 3), x.between(0, 9), yr)
	if j.in.Year >= 1998 && x.chance(0.4) {
		head = fmt.Sprintf("Y2K remediation %s: two-digit year handling below replaced. Pre-remediation code kept for the certification binder.", pick(x, []string{"1998", "1999", "phase 2", "wave 3"}))
	}
	var block []Line
	note := j.cs
	note.width = 60
	for _, t := range note.para(head + " -- " + who.Initials) {
		block = append(block, Line{Text: cs.commentOut(t)})
	}
	for i := from; i < to; i++ {
		t := src[i].Text
		if src[i].Code && x.chance(0.35) {
			done := false
			t = numLit.ReplaceAllStringFunc(t, func(m string) string {
				if done {
					return m
				}
				done = true
				if strings.Contains(m, ".") {
					f, err := strconv.ParseFloat(m, 64)
					if err != nil {
						return m
					}
					return strconv.FormatFloat(f+float64(x.between(-3, 3))/100, 'f', len(m)-strings.Index(m, ".")-1, 64)
				}
				v, _ := strconv.Atoi(m)
				return strconv.Itoa(maxi(0, v+x.between(-2, 2)))
			})
		}
		block = append(block, Line{Text: cs.commentOut(t)})
	}
	out := append([]Line{}, src[:from]...)
	out = append(out, block...)
	return append(out, src[from:]...)
}

// condCompile adds the conditional-compilation habits of the C, COBOL and
// FORTRAN shops.
func condCompile(j *job, x *rnd, src []Line) []Line {
	switch j.id.LangKey {
	case "c", "cbtr":
		if !x.chance(0.5) {
			return src
		}
		sym := pick(x, []string{"OLD_LADDER", "DEBUG", "PRE_Y2K", "VAX", "TRACE_DECK", "MVS"})
		var out []Line
		wrapped := false
		for _, l := range src {
			if !wrapped && l.Code && strings.Contains(l.Text, "DEBUG CHECKPOINT") {
				ind := l.Text[:len(l.Text)-len(strings.TrimLeft(l.Text, " \t"))]
				out = append(out, Line{Text: "#ifdef " + sym}, l, Line{Text: "#else"},
					Line{Text: ind + "/* checkpoint suppressed in production (" + j.g.memo() + ") */"}, Line{Text: "#endif"})
				wrapped = true
				continue
			}
			out = append(out, l)
		}
		if !wrapped {
			out = append([]Line{{Text: "#if 0"}, {Text: "/* " + asciiText(j.g.comment(nil, "history", "warnings")) + " */"}, {Text: "#endif"}}, out...)
		}
		return out
	case "cobol":
		if !x.chance(0.5) {
			return src
		}
		for i, l := range src {
			if l.Code && strings.Contains(l.Text, "DEBUG CHECKPOINT") && len(l.Text) > 7 && x.chance(0.6) {
				r := []rune(l.Text)
				r[6] = 'D' // debugging line: compiled only WITH DEBUGGING MODE
				src[i].Text = string(r)
				src[i].Code = false
			}
		}
	case "f77":
		if j.in.Year > 1997 || !x.chance(0.4) {
			return src
		}
		for i, l := range src {
			if l.Code && strings.HasPrefix(l.Text, " ") && strings.Contains(strings.ToUpper(l.Text), "WRITE") && x.chance(0.25) {
				src[i].Text = "D" + l.Text[1:] // DEC extended source: compiled with /D_LINES
				src[i].Code = false
			}
		}
	}
	return src
}

// reformatRoutine: one routine re-indented by a later maintainer who
// preferred a different indent, with a note saying so.
func reformatRoutine(j *job, x *rnd, src []Line, cs cmtSyntax) []Line {
	starts := routineStarts(src)
	if len(starts) < 3 {
		return src
	}
	i := x.intn(len(starts) - 1)
	from, to := starts[i], starts[i+1]
	unit := pick(x, []string{"\t", "  ", "    ", "   ", "        "})
	who := j.g.coder()
	for k := from; k < to; k++ {
		t := src[k].Text
		if !src[k].Code {
			continue
		}
		lead := len(t) - len(strings.TrimLeft(t, " \t"))
		if lead == 0 {
			continue
		}
		width := strings.Count(t[:lead], "\t")*4 + strings.Count(t[:lead], " ")
		levels := maxi(1, (width+2)/4)
		src[k].Text = strings.Repeat(unit, levels) + strings.TrimLeft(t, " \t")
	}
	notes := []string{
		fmt.Sprintf("Reformatted by %s, %d, using the department's new layout standard. Logic unchanged.", who.Display(x), x.between(maxi(1980, j.in.Year-15), j.in.Year)),
		fmt.Sprintf("%s re-indented this routine with an editor macro in %d. It has looked like this ever since.", who.Display(x), x.between(maxi(1980, j.in.Year-15), j.in.Year)),
	}
	note := Line{Text: cs.commentOut(asciiText(pick(x, notes)))}
	out := append([]Line{}, src[:from]...)
	out = append(out, note)
	return append(out, src[from:]...)
}
