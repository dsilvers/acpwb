package legacyjob

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// unitInfo is what a compiler would know about one program unit, recovered
// from the emitted lines (the routine prologue's "Routine:" field marks each
// unit, in every language).
type unitInfo struct {
	name       string
	start, end int // line index range
	defs       map[string]int
	refs       map[string][]int
}

var routineRe = regexp.MustCompile(`(?i)routine:\s+([A-Za-z0-9_$-]+)`)

func analyzeUnits(src []Line) []*unitInfo {
	var units []*unitInfo
	cur := &unitInfo{name: "(prologue)", defs: map[string]int{}, refs: map[string][]int{}}
	for i, l := range src {
		if m := routineRe.FindStringSubmatch(l.Text); m != nil && !l.Code {
			cur.end = i
			units = append(units, cur)
			cur = &unitInfo{name: strings.ToUpper(m[1]), start: i, defs: map[string]int{}, refs: map[string][]int{}}
		}
		for _, d := range l.Defs {
			if _, ok := cur.defs[d]; !ok {
				cur.defs[d] = i + 1
			}
		}
		for _, r := range l.Refs {
			cur.refs[r] = append(cur.refs[r], i+1)
		}
	}
	cur.end = len(src)
	units = append(units, cur)
	return units
}

type symInfo struct {
	name, typ string
	size      int
	block     string
	defLine   int
	refs      []int
}

func symbolTable(j *job, units []*unitInfo) []*symInfo {
	m := map[string]*symInfo{}
	add := func(s Sym) {
		if _, ok := m[s.Name]; ok {
			return
		}
		size := 4
		switch s.Type {
		case "dbl":
			size = 8
		case "str":
			size = s.Len
		}
		for _, d := range s.Dims {
			size *= d
		}
		block := s.Common
		if s.Role == "const" {
			block = "(PARAMETER)"
		} else if block == "" {
			block = "(LOCAL)"
		}
		m[s.Name] = &symInfo{name: s.Name, typ: s.Type, size: size, block: block}
	}
	for _, s := range j.p.Globals {
		add(s)
	}
	for _, p := range j.p.Procs {
		for _, s := range p.Locals {
			add(s)
		}
	}
	for _, u := range units {
		for n, d := range u.defs {
			if si, ok := m[n]; ok && (si.defLine == 0 || d < si.defLine) {
				si.defLine = d
			}
		}
		for n, rs := range u.refs {
			if si, ok := m[n]; ok {
				si.refs = append(si.refs, rs...)
			}
		}
	}
	var out []*symInfo
	for _, s := range m {
		sort.Ints(s.refs)
		out = append(out, s)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].name < out[b].name })
	return out
}

// unusedPerUnit: symbols declared in a unit but never referenced there.
func unusedPerUnit(units []*unitInfo) [][2]string {
	var out [][2]string
	for _, u := range units {
		var names []string
		for d := range u.defs {
			if len(u.refs[d]) == 0 {
				names = append(names, d)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			out = append(out, [2]string{u.name, n})
		}
	}
	return out
}

func typeWord(t string, upper bool) string {
	w := map[string]string{"int": "INTEGER*4", "dbl": "REAL*8", "str": "CHARACTER"}[t]
	if w == "" {
		w = "INTEGER*4"
	}
	if !upper {
		return strings.ToLower(w)
	}
	return w
}

func refList(rs []int, max int) string {
	var parts []string
	for i, r := range rs {
		if i >= max {
			parts = append(parts, fmt.Sprintf("(+%d)", len(rs)-max))
			break
		}
		parts = append(parts, fmt.Sprint(r))
	}
	return strings.Join(parts, " ")
}

// compilerSections returns the compiler-output sections for the compute
// stage, and (for languages with a real compiler listing) a replacement for
// the plain source section.
func compilerSections(j *job, l *Listing, src []Line, srcName string) (listing *Section, extra []Section) {
	x := j.x
	units := analyzeUnits(src)
	syms := symbolTable(j, units)
	unused := unusedPerUnit(units)
	var dead []string
	for _, p := range j.p.Procs {
		if p.Dead {
			dead = append(dead, strings.ToUpper(p.Name))
		}
	}
	sort.Strings(dead)
	nCode := 0
	for _, s := range src {
		if s.Code {
			nCode++
		}
	}
	date := fmt.Sprintf("%s %02d, %04d", monthAbbr[j.in.Month-1], j.in.Day, j.in.Year)
	clock := fmt.Sprintf("%02d:%02d:%02d", x.between(0, 23), x.between(0, 59), x.between(0, 59))
	var diag, xref, smap []string
	lang := j.id.LangKey

	switch lang {
	case "f4", "f77":
		vax := strings.Contains(l.Compiler, "VAX") || strings.Contains(l.Compiler, "DEC")
		// numbered listing
		var lst []string
		page := 0
		hdr := func() {
			page++
			lst = append(lst, fmt.Sprintf("1%-24s %-28s DATE: %s  TIME: %s  PAGE: %4d", "LEVEL "+fmt.Sprintf("%d.%d.%d", x.between(1, 3), x.between(0, 9), x.between(0, 9)), strings.ToUpper(l.Compiler), date, clock, page))
			if page == 1 {
				lst = append(lst, " REQUESTED OPTIONS: "+pick(x, []string{"OPT(2),XREF,MAP,SOURCE", "NOOPT,XREF,MAP,LIST", "OPT(3),SOURCE,FLAG(I)"}),
					" OPTIONS IN EFFECT: NOLIST MAP XREF NOGOSTMT NODECK SOURCE TERM OBJECT FIXED NOTEST NOTRMFLG SRCFLG NOSYM NORENT SDUMP",
					"                    AUTODBL(NONE) NOSXM IL(DIM) NOVECTOR NOEC NOICA NODIRECTIVE NODC FLAG(I) LANGLVL(77) OPT(2)", "")
			}
			lst = append(lst, "     ISN *....*....1....*....2....*....3....*....4....*....5....*....6....*....7.*.......8", "")
		}
		hdr()
		isn := 0
		for i, s := range src {
			if i > 0 && i%58 == 0 {
				hdr()
			}
			if s.Code {
				isn++
				lst = append(lst, fmt.Sprintf("%8d  %s", isn, s.Text))
			} else {
				lst = append(lst, "          "+s.Text)
			}
		}
		listing = &Section{ID: "source", Title: "Compiler listing", Tab: fmt.Sprintf("%s — SOURCE LISTING, %d STATEMENTS", srcName, isn), Kind: "pre", Text: strings.Join(lst, "\n")}
		for _, u := range unused {
			if vax {
				diag = append(diag, fmt.Sprintf("%%FORT-I-VARUNUSED, Variable %s is declared but never used", u[1]), fmt.Sprintf("    in program unit %s", u[0]))
			} else {
				diag = append(diag, fmt.Sprintf(" ILF%04dI I  VARIABLE %s IS DECLARED BUT NOT REFERENCED IN %s.", x.between(2000, 2999), u[1], u[0]))
			}
		}
		for _, d := range dead {
			if vax {
				diag = append(diag, fmt.Sprintf("%%LINK-W-UDFSYM, %s is never called", d))
			} else {
				diag = append(diag, fmt.Sprintf(" ILF%04dI W  SUBPROGRAM %s IS NEVER REFERENCED.", x.between(3000, 3999), d))
			}
		}
		sort.Strings(diag)
		diag = append(diag, "", fmt.Sprintf(" *STATISTICS*  SOURCE STATEMENTS = %6d, PROGRAM SIZE = %6d BYTES, PROGRAM NAME = %s  PAGE: %d", isn, x.between(4000, 90000), j.id.Program, page),
			fmt.Sprintf(" *STATISTICS*  %d DIAGNOSTICS GENERATED, HIGHEST SEVERITY CODE IS %d", len(unused)+len(dead), pick(x, []int{0, 4})),
			" ****** END OF COMPILATION ******")
		xref = append(xref, " *****  CROSS REFERENCE DICTIONARY  *****", "", fmt.Sprintf(" %-12s %-12s %-10s %8s   %s", "NAME", "MODE", "BLOCK", "DEFINED", "REFERENCES"), "")
		for _, s := range syms {
			xref = append(xref, fmt.Sprintf(" %-12s %-12s %-10s %8d   %s", s.name, typeWord(s.typ, true), trunc(s.block, 10), s.defLine, refList(s.refs, 12)))
		}
		smap = storageMap(j, syms, true)
	case "gf":
		for _, u := range unused {
			ln := 0
			for _, uu := range units {
				if uu.name == u[0] {
					ln = uu.defs[u[1]]
				}
			}
			text := ""
			if ln > 0 && ln <= len(src) {
				text = strings.TrimRight(src[ln-1].Text, " ")
				if len(text) > 72 {
					text = text[:72]
				}
			}
			col := strings.Index(strings.ToUpper(text), strings.ToUpper(u[1])) + 1
			diag = append(diag, fmt.Sprintf("%s:%d:%d:", srcName, ln, col), "", fmt.Sprintf("%5d | %s", ln, text), fmt.Sprintf("      | %s1", strings.Repeat(" ", maxi(0, col-1))),
				fmt.Sprintf("Warning: Unused variable '%s' declared at (1) [-Wunused-variable]", strings.ToLower(u[1])))
		}
		for _, d := range dead {
			diag = append(diag, fmt.Sprintf("%s: Warning: Procedure '%s' defined but not used [-Wunused-procedure]", srcName, strings.ToLower(d)))
		}
		smap = nmTable(j, syms, units)
	case "cobol":
		var lst []string
		lst = append(lst, fmt.Sprintf("PP 5655-%s IBM %s   %s   Date %s  Time %s   Page %4d", pick(x, []string{"G53", "S71", "W32"}), l.Compiler, j.id.Program, date, clock, 1),
			"  LineID  PL SL  ----+-*A-1-B--+----2----+----3----+----4----+----5----+----6----+----7-|--+----8 Map and Cross Reference")
		for i, s := range src {
			lst = append(lst, fmt.Sprintf("  %06d         %s", i+1, s.Text))
		}
		listing = &Section{ID: "source", Title: "Compiler listing", Tab: fmt.Sprintf("%s — IGYCRCTL LISTING", srcName), Kind: "pre", Text: strings.Join(lst, "\n")}
		for _, u := range unused {
			diag = append(diag, fmt.Sprintf("  IGYPS%04d-I   \"%s\" was defined but was not referenced in paragraph %s.", x.between(2000, 2999), u[1], u[0]))
		}
		for _, d := range dead {
			diag = append(diag, fmt.Sprintf("  IGYPS%04d-W   Paragraph \"%s\" was never performed.", x.between(3000, 3999), d))
		}
		sort.Strings(diag)
		diag = append(diag, "", fmt.Sprintf("-Messages    Total    Informational    Warning    Error    Severe    Terminating"),
			fmt.Sprintf("0Printed:  %6d  %14d  %9d", len(unused)+len(dead), len(unused), len(dead)),
			fmt.Sprintf("-* Statistics for COBOL program %s:", j.id.Program),
			fmt.Sprintf(" *    Source records = %d", len(src)),
			fmt.Sprintf(" *    Data Division statements = %d", x.between(60, 400)),
			fmt.Sprintf(" *    Procedure Division statements = %d", nCode),
			fmt.Sprintf("0End of compilation 1,  program %s,  highest severity 4.", j.id.Program))
		xref = append(xref, " Data Division Map", "", " Data Definition Attribute codes (rightmost column) have the following meanings:",
			"     D = Object of OCCURS DEPENDING    G = GLOBAL     LSEQ= ORGANIZATION LINE SEQUENTIAL", "",
			fmt.Sprintf(" %-8s %-34s %-8s %-8s %s", "Source", "Hierarchy and Data Name", "Base", "Displ", "Data Type"))
		off := 0
		for _, s := range syms {
			xref = append(xref, fmt.Sprintf(" %8d %-34s BLW=0000 %08X %s", s.defLine, trunc(s.name, 34), off, map[string]string{"int": "Binary", "dbl": "Comp-2", "str": "Display"}[s.typ]))
			off += s.size
		}
		xref = append(xref, "", " Defined   Cross-reference of data names   References", "")
		for _, s := range syms {
			xref = append(xref, fmt.Sprintf(" %7d   %-32s %s", s.defLine, trunc(s.name, 32), refList(s.refs, 10)))
		}
	case "basic":
		for _, u := range unused {
			diag = append(diag, fmt.Sprintf("Warning: variable %s assigned but not used in %s", u[1], u[0]))
		}
		xref = append(xref, fmt.Sprintf("%s  Cross Reference  %s", strings.ToUpper(l.Compiler), date), "", fmt.Sprintf("%-12s %-8s %s", "Symbol", "Type", "Lines"))
		for _, s := range syms {
			var lines []string
			for _, r := range s.refs {
				if r-1 < len(src) {
					lines = append(lines, src[r-1].Label)
				}
			}
			xref = append(xref, fmt.Sprintf("%-12s %-8s %s", s.name, map[string]string{"int": "Double", "dbl": "Double", "str": "String"}[s.typ], strings.Join(firstN(lines, 14), " ")))
		}
		diag = append(diag, "", fmt.Sprintf("%d Bytes Available", x.between(200000, 600000)), fmt.Sprintf("%d Bytes Free", x.between(100000, 400000)), "", "0 Warning Error(s)", "0 Severe  Error(s)")
	case "c":
		for _, u := range unused {
			ln := 0
			for _, uu := range units {
				if uu.name == u[0] {
					ln = uu.defs[u[1]]
				}
			}
			diag = append(diag, fmt.Sprintf("\"%s\", line %d: warning: %s unused in function %s", srcName, ln, u[1], strings.ToLower(u[0])))
		}
		for _, d := range dead {
			diag = append(diag, fmt.Sprintf("%s defined( %s ), but never used", d, srcName))
		}
		sort.Strings(diag)
		diag = append([]string{"lint -hbx " + srcName, srcName + ":"}, diag...)
		smap = nmTable(j, syms, units)
	case "perl":
		diag = append(diag, "$ perl -wc "+srcName)
		for _, u := range unused {
			diag = append(diag, fmt.Sprintf("Name \"main::%s\" used only once: possible typo at %s line %d.", u[1], srcName, x.between(10, len(src))))
		}
		diag = append(diag, srcName+" syntax OK", "", "$ perlcritic --severity 3 "+srcName)
		for _, p := range j.p.Procs {
			diag = append(diag, fmt.Sprintf("Subroutine \"%s\" does not end with \"return\" at line %d, column 1.  See page 197 of PBP.  (Severity: 4)", p.Name, x.between(10, len(src))))
			if x.chance(0.4) {
				diag = append(diag, fmt.Sprintf("Mixed-case subroutine name at line %d, column 1.  See page 45 of PBP.  (Severity: 1)", x.between(10, len(src))))
			}
		}
		xref = append(xref, "$ perl -MO=Xref "+srcName, "File "+srcName)
		for _, u := range units {
			xref = append(xref, "  Subroutine "+u.name)
			var names []string
			for n := range u.refs {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				xref = append(xref, fmt.Sprintf("    Package main\n      $%-22s %s", n, refList(u.refs[n], 10)))
			}
		}
	case "py2":
		diag = append(diag, "************* Module "+strings.TrimSuffix(srcName, ".py"))
		for _, p := range j.p.Procs {
			if strings.ToUpper(p.Name) == p.Name {
				diag = append(diag, fmt.Sprintf("C:%3d, 0: Invalid function name \"%s\" (invalid-name)", x.between(10, len(src)), p.Name))
			}
			if len(p.Body) > 12 {
				diag = append(diag, fmt.Sprintf("R:%3d, 0: %s: Too many statements (%d/50) (too-many-statements)", x.between(10, len(src)), p.Name, len(p.Body)*x.between(2, 5)))
			}
		}
		for _, u := range unused {
			diag = append(diag, fmt.Sprintf("W:%3d, 4: Unused variable '%s' (unused-variable)", x.between(10, len(src)), u[1]))
		}
		for _, s := range syms {
			if s.block != "(LOCAL)" && s.block != "(PARAMETER)" && x.chance(0.3) {
				diag = append(diag, fmt.Sprintf("W:%3d, 4: Using the global statement (global-statement)", x.between(10, len(src))))
			}
		}
		diag = append(diag, "", "Report", "======", fmt.Sprintf("%d statements analysed.", nCode), "",
			fmt.Sprintf("Your code has been rated at %d.%02d/10 (previous run: %d.%02d/10, %+d.%02d)", x.between(-3, 6), x.between(0, 99), x.between(-5, 5), x.between(0, 99), x.between(-2, 2), x.between(0, 99)))
	}

	if m := metaOf(lang); len(diag) == 0 && m.diagTab != "" {
		sub := func(t string) string {
			return strings.NewReplacer("{SRC}", srcName, "{src}", srcName, "{PROG}", strings.ToUpper(j.id.Program),
				"{N}", fmt.Sprint(x.between(10, maxi(11, len(src))))).Replace(t)
		}
		for _, h := range m.diagHead {
			diag = append(diag, sub(h))
		}
		for _, u := range unused {
			if m.diagUnused != "" {
				diag = append(diag, strings.Split(sub(strings.NewReplacer("{VAR}", u[1], "{UNIT}", u[0]).Replace(m.diagUnused)), "\n")...)
			}
		}
		for _, d := range dead {
			if m.diagDead != "" {
				diag = append(diag, strings.Split(sub(strings.ReplaceAll(m.diagDead, "{ROUTINE}", d)), "\n")...)
			}
		}
		for _, t := range m.diagTail {
			diag = append(diag, sub(t))
		}
		if len(diag) == 0 {
			diag = append(diag, "0 warnings, 0 errors.")
		}
		xref = append(xref, fmt.Sprintf("%-24s %-10s %8s   %s", "SYMBOL", "TYPE", "DEFINED", "REFERENCES"))
		for _, s := range syms {
			xref = append(xref, fmt.Sprintf("%-24s %-10s %8d   %s", trunc(s.name, 24), map[string]string{"int": "integer", "dbl": "float", "str": "string"}[s.typ], s.defLine, refList(s.refs, 12)))
		}
	}

	if len(diag) > 0 {
		tab := map[string]string{"f4": "COMPILER DIAGNOSTICS", "f77": "COMPILER DIAGNOSTICS", "gf": "gfortran -Wall", "cobol": "IGYCRCTL MESSAGES", "basic": "BASCOM /X", "c": "lint", "perl": "perl -wc / perlcritic", "py2": "pylint", "perl4": "perl -wc", "cbtr": "lint"}[lang]
		if tab == "" {
			tab = metaOf(lang).diagTab
		}
		extra = append(extra, Section{ID: "diagnostics", Title: "Diagnostics", Tab: tab, Kind: "pre", Text: strings.Join(diag, "\n")})
	}
	if len(xref) > 0 {
		extra = append(extra, Section{ID: "xref", Title: "Cross-reference", Tab: "CROSS REFERENCE", Kind: "pre", Text: strings.Join(xref, "\n")})
	}
	if len(smap) > 0 {
		extra = append(extra, Section{ID: "storagemap", Title: "Storage map", Tab: "STORAGE MAP", Kind: "pre", Text: strings.Join(smap, "\n")})
	}
	return listing, extra
}

func firstN(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func storageMap(j *job, syms []*symInfo, upper bool) []string {
	x := j.x
	out := []string{" *****  STORAGE MAP  *****", "", fmt.Sprintf(" %-10s %-12s %-12s %-10s %-10s", "BLOCK", "NAME", "MODE", "OFFSET", "LENGTH"), ""}
	blocks := map[string][]*symInfo{}
	var order []string
	for _, s := range syms {
		if _, ok := blocks[s.block]; !ok {
			order = append(order, s.block)
		}
		blocks[s.block] = append(blocks[s.block], s)
	}
	sort.Strings(order)
	base := x.between(0x1000, 0x8000)
	for _, b := range order {
		off := 0
		for _, s := range blocks[b] {
			out = append(out, fmt.Sprintf(" %-10s %-12s %-12s %06X     %-10d", trunc(b, 10), s.name, typeWord(s.typ, upper), base+off, s.size))
			off += s.size
		}
		base += off + x.between(8, 256)
	}
	out = append(out, "", fmt.Sprintf(" TOTAL MEMORY REQUIREMENTS %06X BYTES", base))
	return out
}

func nmTable(j *job, syms []*symInfo, units []*unitInfo) []string {
	x := j.x
	prog := strings.ToLower(j.id.Program)
	out := []string{"$ nm -n " + prog}
	addr := 0x401000 + x.between(0, 0xfff)
	for _, u := range units {
		if u.name == "(prologue)" {
			continue
		}
		out = append(out, fmt.Sprintf("%016x T %s_", addr, strings.ToLower(u.name)))
		addr += x.between(0x40, 0x600)
	}
	addr = 0x601040 + x.between(0, 0xff)*8
	for _, s := range syms {
		if s.block == "(LOCAL)" || s.block == "(PARAMETER)" {
			continue
		}
		kind := "B"
		if strings.Contains(s.block, "W") {
			kind = "D"
		}
		out = append(out, fmt.Sprintf("%016x %s %s_", addr, kind, strings.ToLower(s.block)))
		addr += s.size
	}
	return out
}
