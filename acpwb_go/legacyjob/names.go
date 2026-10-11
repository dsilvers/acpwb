package legacyjob

import (
	"fmt"
	"strings"
)

// roleWords: candidate stems for each logical variable role.
var roleWords = map[string][]string{
	"nmet":   {"NMET", "NMTR", "NUMMET", "METCNT", "NM", "MCOUNT", "NMETR"},
	"nben":   {"NBEN", "NBMK", "NUMBEN", "BENCNT", "NB", "BCOUNT", "NBNCH"},
	"bsln":   {"BSLN", "BASE", "BASLN", "BAS", "BL", "PRIOR", "OPEN"},
	"curr":   {"CURR", "CUR", "CRNT", "NOW", "CR", "LATEST", "CLOSE"},
	"delt":   {"DELT", "DLT", "CHG", "DIF", "DEL", "DELTA", "VAR"},
	"medn":   {"MEDN", "MED", "MDN", "MID", "P50", "CENTR", "MEDIAN"},
	"pctl":   {"PCTL", "PCT", "LAD", "PTL", "QTL", "RUNG", "TIER"},
	"mult":   {"PCTMUL", "FACTOR", "RUNG", "MULT", "FAC", "LADR", "COEF"},
	"shadow": {"SHADOW", "XCHK", "CHKPCT", "VERIFY", "DUPLAD", "MIRROR"},
	"idebug": {"IDEBUG", "DEBUG", "DBGFLG", "ITRACE", "IVERB", "DIAG"},
	"ipage":  {"IPAGE", "PAGNO", "NPAGE", "PAGE", "IPG"},
	"iline":  {"ILINE", "LINCNT", "NLINE", "LINE", "ILN"},
	"ierr":   {"IERR", "ERRCD", "IRC", "KODE", "ISTAT", "RETCD"},
	"chksum": {"ICHK", "CHKSUM", "CKSUM", "IHASH", "SEAL", "ISUM"},
	"mname":  {"MNAME", "METNAM", "CMNAME", "MLABEL", "MDESC"},
	"bname":  {"BNAME", "BENNAM", "CBNAME", "BLABEL", "BDESC"},
	"org":    {"CORGNM", "CLIENT", "ORGNAM", "CLNAME", "FIRM"},
	"recid":  {"CRECID", "RECID", "RECORD", "DOCID", "RID"},
	"peer":   {"CPEERG", "PEERGP", "PEER", "PGROUP", "COMPST"},
	"seal":   {"CSEAL", "DECKID", "PROVNC", "STAMP", "WMARK"},
	"maxmet": {"MAXMET", "MXMET", "LIMMET", "MMAX"},
	"maxben": {"MAXBEN", "MXBEN", "LIMBEN", "BMAX"},
	"nladr":  {"NLADR", "NRUNG", "NPCT", "LADSZ"},
	"icode":  {"ICODE", "KODE", "IERNO", "MSGNO", "NCODE"},
	"tmp":    {"ITMP", "KTMP", "IWORK", "ISAVE", "NTEMP", "JHOLD"},
	"i":      {"I", "II", "IM", "IX"},
	"j":      {"J", "JJ", "JB", "JX"},
	"k":      {"K", "KK", "KP", "KX"},
	"l":      {"L", "LL", "LC", "LX"},
}

var roleType = map[string]string{
	"mult": "dbl", "mname": "str", "bname": "str", "org": "str", "recid": "str", "peer": "str", "seal": "str",
}

type naming struct {
	x      *rnd
	family string // fortran cobol basic c perl
	maxLen int
	scheme int
	prefix string
	m      map[string]string
	used   map[string]bool
	n      int
}

func newNaming(x *rnd, langKey string, author *Coder) *naming {
	nm := &naming{x: x, family: langFamily(langKey), m: map[string]string{}, used: map[string]bool{}}
	switch langKey {
	case "f4", "f77":
		nm.maxLen = 6
	case "gf":
		nm.maxLen = 14
	case "basic":
		nm.maxLen = 6
	case "atari", "mumps":
		nm.maxLen = 8
	case "dbase", "clipper", "foxpro", "cl400":
		nm.maxLen = 10
	case "rpg":
		nm.maxLen = 6
	case "clist", "spss", "sas":
		nm.maxLen = 8
	case "focus", "nomad", "ramis":
		nm.maxLen = 12
	default:
		nm.maxLen = 24
	}
	nm.scheme = x.intn(6)
	if author != nil && len(author.Initials) >= 2 {
		nm.prefix = author.Initials[:2]
	} else {
		nm.prefix = "WK"
	}
	// reserve language keywords that a stem could collide with
	for _, kw := range []string{"DO", "IF", "GO", "TO", "END", "STOP", "DATA", "READ", "CALL", "FOR", "NEXT",
		"PRINT", "LET", "DIM", "REM", "RETURN", "INT", "MOD", "INPUT", "LINE", "PAGE", "FORMAT", "WRITE",
		"VALUE", "CLOSE", "OPEN", "DELETE", "VAR", "NOW",
		// keywords / builtins of the other dialects (Ada, Pascal, VB, C#, SQL...)
		"BASE", "CHECKED", "FIXED", "LOCK", "RANGE", "DELTA", "DIGITS", "RECORD", "TYPE", "BODY", "ENTRY",
		"ACCESS", "NEW", "ABS", "REM", "MOD", "DIV", "SET", "FILE", "LABEL", "BEGIN", "ELSE", "THEN", "CASE",
		"LOOP", "EXIT", "NAME", "DATE", "ERROR", "GET", "PUT", "SEEK", "CURRENT", "LEVEL", "COUNT", "SUM",
		"MIN", "MAX", "RANK", "ROW", "ROWS", "POSITION", "PERCENT", "TOP", "KEY", "USER", "SIZE", "MID",
		"LEN", "VAL", "STR", "OUT", "IN", "IS", "OR", "AND", "NOT", "MEDIAN", "LATEST", "STEP", "TIME",
		"PRIOR", "OF", "AT", "BY", "ON", "AS", "NULL", "DECIMAL", "EVENT", "PARAMS", "OBJECT", "STRING",
		"TABLE", "INDEX", "VIEW", "ORDER", "GROUP", "WHERE", "FROM", "SELECT", "UPDATE", "INSERT", "FIELD",
		"STAMP", "TIER", "LIST", "MAP", "DATA", "REF", "WORK", "SAVE", "STATUS", "RESULT", "RETURNS",
		"DEBUG", "ERR", "ITEM", "TEXT", "CELLS", "SHEET", "FORMAT", "TRACE", "VERIFY", "DIAG", "PCT", "SEQ"} {
		nm.used[kw] = true
	}
	return nm
}

func trunc(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// intLetter makes a FORTRAN-style name honour implicit typing: integers
// start with I-N, everything else with A-H / O-Z.
func intLetter(name, typ string) string {
	if name == "" {
		return name
	}
	c := name[0]
	isInt := c >= 'I' && c <= 'N'
	switch typ {
	case "int":
		if !isInt {
			return "K" + name
		}
	case "dbl":
		if isInt {
			return "D" + name
		}
	}
	return name
}

func (nm *naming) style(stem, typ, role string) string {
	switch nm.family {
	case "fortran", "basic":
		var s string
		switch nm.scheme {
		case 0:
			s = stem
		case 1: // Hungarian-ish
			s = map[string]string{"int": "IV", "dbl": "DV", "str": "CV"}[typ] + stem
		case 2: // vowel-dropped
			s = string(stem[0]) + strings.NewReplacer("A", "", "E", "", "I", "", "O", "", "U", "").Replace(stem[1:])
		case 3: // initials prefix
			s = nm.prefix + stem
		case 4: // numbered
			nm.n++
			s = fmt.Sprintf("%s%03d", map[string]string{"int": "N", "dbl": "X", "str": "C"}[typ], nm.n*7%1000)
		default:
			s = stem
		}
		if len(role) == 1 { // loop indices stay short
			s = stem
		}
		if nm.family == "basic" {
			if typ == "str" {
				s = trunc(s, 5) + "$"
				return s
			}
		}
		return trunc(intLetter(s, typ), nm.maxLen)
	case "cobol":
		pre := pick(nm.x, []string{"WS-", "W-", "WK-", "TB-", "CTL-", "RPT-"})
		if len(role) == 1 {
			pre = "WS-"
		}
		long := map[string]string{
			"BSLN": "BASELINE", "CURR": "CURRENT", "DELT": "CHANGE", "MEDN": "MEDIAN", "PCTL": "PERCENTILE",
			"NMET": "METRIC-COUNT", "NBEN": "BENCH-COUNT", "MNAME": "METRIC-NAME", "BNAME": "BENCH-NAME",
		}
		if l, ok := long[stem]; ok && nm.x.chance(0.6) {
			stem = l
		}
		if len(role) == 1 {
			stem = "SUB-" + stem
		}
		return pre + stem
	case "c":
		s := strings.ToLower(stem)
		switch nm.scheme % 4 {
		case 1:
			s = map[string]string{"int": "n", "dbl": "d", "str": "sz"}[typ] + strings.ToUpper(s[:1]) + s[1:]
		case 2:
			s = "g_" + s
		case 3:
			s = strings.ToLower(nm.prefix) + "_" + s
		}
		if len(role) == 1 {
			s = strings.ToLower(stem)
		}
		return s
	case "py2":
		s := strings.ToLower(stem)
		if nm.scheme%3 == 1 && len(role) > 1 {
			s = "_" + s
		} else if nm.scheme%3 == 2 && len(role) > 1 {
			s = s + "_" + pick(nm.x, []string{"list", "tbl", "map", "val", "data"})
		}
		return s
	case "perl":
		s := strings.ToLower(stem)
		if nm.scheme%3 == 1 && len(role) > 1 {
			s = s + "_" + pick(nm.x, []string{"ref", "tbl", "val", "list", "data"})
		}
		return s
	default:
		return nm.dialectStyle(stem, typ, role)
	}
}

// dialectStyle covers the naming conventions of the dialect-emitter
// languages (see emit_dialect.go).
func (nm *naming) dialectStyle(stem, typ, role string) string {
	short := len(role) == 1
	lower := strings.ToLower(stem)
	title := strings.ToUpper(lower[:1]) + lower[1:]
	sc := nm.scheme % 3
	switch nm.family {
	case "ada":
		if short {
			return title
		}
		return [3]string{title, "Ws_" + title, title + "_" + pick(nm.x, []string{"Val", "Tbl", "Cnt", "Rec"})}[sc]
	case "pascal":
		if short {
			return lower
		}
		return [3]string{lower, map[string]string{"int": "i", "dbl": "r", "str": "s"}[typ] + title, "g" + title}[sc]
	case "pli":
		if short {
			return stem
		}
		return [3]string{stem, "WS_" + stem, stem + "_" + pick(nm.x, []string{"TBL", "VAL", "CTR", "AREA"})}[sc]
	case "pal":
		if short {
			return lower
		}
		return [3]string{lower, lower + "Val", "the" + title}[sc]
	case "xbase":
		if short {
			return stem
		}
		return trunc([3]string{stem, "M_" + stem, map[string]string{"int": "N", "dbl": "N", "str": "C"}[typ] + stem}[sc], nm.maxLen)
	case "natural":
		if short {
			return "#" + stem
		}
		return [3]string{"#" + stem, "#WS-" + stem, "#" + stem + "-" + pick(nm.x, []string{"TAB", "VAL", "CNT", "FLD"})}[sc]
	case "informix":
		if short {
			return lower
		}
		return [3]string{lower, "l_" + lower, "g_" + lower}[sc]
	case "vb":
		if short {
			return lower
		}
		hung := map[string]string{"int": "lng", "dbl": "dbl", "str": "str"}[typ]
		return [3]string{hung + title, "m" + title, title}[sc]
	case "vbnet":
		if short {
			return lower
		}
		return [3]string{lower, "m_" + lower, "_" + lower}[sc]
	case "csharp":
		if short {
			return lower
		}
		return [3]string{lower, "_" + lower, "m_" + lower}[sc]
	case "php":
		if short {
			return lower
		}
		return [3]string{lower, lower + "_" + pick(nm.x, []string{"arr", "val", "tbl", "data"}), "g_" + lower}[sc]
	case "pick":
		if short {
			return stem
		}
		return [3]string{stem, stem + "." + pick(nm.x, []string{"TBL", "VAL", "REC", "ARR"}), "WS." + stem}[sc]
	case "mumps":
		s := strings.Map(func(r rune) rune {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, strings.ToUpper(stem))
		if !short && sc == 1 {
			s = "Z" + s
		}
		return trunc(s, 8)
	case "rexx":
		if short {
			return lower
		}
		return [3]string{lower, "w_" + lower, lower + "_" + pick(nm.x, []string{"cnt", "tab", "val"})}[sc]
	case "clist", "cl400", "rpg", "asm", "spss", "focus":
		u := strings.ToUpper(strings.Map(func(r rune) rune {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, stem))
		if !short && sc == 1 && nm.family != "asm" {
			u = "W" + u
		}
		if nm.family == "focus" && !short && sc == 2 {
			u = u + "_" + pick(nm.x, []string{"AMT", "CNT", "VAL"})
		}
		return trunc(u, nm.maxLen)
	case "ezt":
		if short {
			return "W-" + stem
		}
		return [3]string{"WS-" + stem, "W-" + stem, stem + "-" + pick(nm.x, []string{"CNT", "TBL", "AMT", "FLD"})}[sc]
	case "apl":
		if short {
			return stem
		}
		return [3]string{stem, "Δ" + stem, stem + "Δ"}[sc]
	case "abap":
		if short {
			return "lv_" + lower
		}
		return [3]string{"gv_" + lower, "g_" + lower, "w_" + lower}[sc]
	case "r":
		if short {
			return lower
		}
		return [3]string{lower, lower + "." + pick(nm.x, []string{"vec", "tab", "val", "n"}), "my" + title}[sc]
	case "matlab", "stata", "sas":
		if short {
			return lower
		}
		return trunc([3]string{lower, lower + "_" + pick(nm.x, []string{"vec", "tab", "val"}), "g" + title}[sc], maxi(nm.maxLen, 8))
	case "awk", "ksh", "tcl", "cfml", "prolog":
		if short {
			return lower
		}
		return [3]string{lower, lower + "_" + pick(nm.x, []string{"tab", "val", "cnt", "arr"}), "g_" + lower}[sc]
	case "bc", "dc":
		return lower
	case "asp":
		if short {
			return lower
		}
		return [3]string{map[string]string{"int": "int", "dbl": "dbl", "str": "str"}[typ] + title, "g" + title, title}[sc]
	case "modula2":
		if short {
			return lower
		}
		return [3]string{title, lower, "g" + title}[sc]
	case "lisp":
		if short || role == "x" {
			return strings.ReplaceAll(lower, "_", "-")
		}
		return "*" + [3]string{lower, lower + "-" + pick(nm.x, []string{"table", "vec", "count"}), "the-" + lower}[sc] + "*"
	case "smalltalk":
		if short {
			return lower
		}
		return [3]string{lower, lower + pick(nm.x, []string{"Table", "Count", "Value"}), "the" + title}[sc]
	case "forth":
		if short {
			return "V" + stem // I, J and K are the DO-loop index words
		}
		return [3]string{stem, "#" + stem, stem + "-" + pick(nm.x, []string{"TBL", "CNT", "VAR"})}[sc]
	default: // sql
		if short {
			return lower
		}
		return [3]string{lower, "v_" + lower, "l_" + lower}[sc]
	}
}

func (nm *naming) get(role string) string {
	if n, ok := nm.m[role]; ok {
		return n
	}
	typ := roleType[role]
	if typ == "" {
		typ = "int"
	}
	stems := roleWords[role]
	for _, stem := range sample(nm.x, stems, len(stems)) {
		n := nm.style(stem, typ, role)
		key := nameKey(n)
		if !nm.used[key] {
			nm.used[key] = true
			nm.m[role] = n
			return n
		}
	}
	n := nm.fresh(typ)
	nm.m[role] = n
	return n
}

var syllA = []string{"K", "M", "N", "J", "L", "I"}
var syllB = []string{"A", "O", "E", "U", "I"}
var syllC = []string{"T", "R", "S", "P", "X", "V", "B", "D", "G", "W"}
var syllD = []string{"CNT", "TOT", "WRK", "SAV", "TMP", "ACC", "IDX", "FLG", "BUF", "REG", "VAL", "OLD", "NEW", "PRV", "NXT"}

// fresh invents a plausible new identifier of the given type for dead code.
func (nm *naming) fresh(typ string) string {
	for tries := 0; tries < 50; tries++ {
		var s string
		switch nm.x.intn(3) {
		case 0:
			s = pick(nm.x, syllA) + pick(nm.x, syllB) + pick(nm.x, syllC) + pick(nm.x, syllD)
		case 1:
			s = pick(nm.x, syllD) + fmt.Sprint(nm.x.between(1, 99))
		default:
			s = pick(nm.x, syllA) + pick(nm.x, syllC) + pick(nm.x, syllB) + pick(nm.x, syllC)
		}
		n := nm.style(s, typ, "x")
		key := nameKey(n)
		if !nm.used[key] {
			nm.used[key] = true
			return n
		}
	}
	nm.n++
	return nm.style(fmt.Sprintf("Z%04d", nm.n), typ, "x")
}

// nameKey normalises a name for collision checks across sigils and case.
func nameKey(n string) string {
	return strings.ToUpper(strings.NewReplacer("$", "", "#", "", "_", "", ".", "", "-", "").Replace(n))
}

// routine invents a routine name (6 chars for old dialects).
func (nm *naming) routine(stem string) string {
	s := stem
	switch nm.family {
	case "fortran", "basic":
		s = trunc(strings.ToUpper(s), nm.maxLenRoutine())
	case "cobol":
		s = strings.ToUpper(s) + "-" + pick(nm.x, []string{"PARA", "RTN", "SECT", "PROC", "STEP"})
	case "c", "perl", "py2", "php", "informix", "sql", "rexx", "awk", "ksh", "tcl", "bc", "r", "matlab", "stata", "sas", "prolog", "abap", "dc":
		s = strings.ToLower(s)
	case "lisp":
		s = strings.ToLower(s)
	case "clist", "cl400", "rpg", "asm", "apl", "forth", "spss", "focus":
		s = trunc(strings.ToUpper(s), 6)
		if nm.family == "cl400" || nm.family == "focus" {
			s = trunc(strings.ToUpper(stem), 10)
		}
	case "ezt":
		s = strings.ToUpper(s) + "-" + pick(nm.x, []string{"RTN", "PROC", "PARA"})
	case "smalltalk":
		l := strings.ToLower(s)
		s = l + pick(nm.x, []string{"", "Step", "Routine"})
	case "pli", "pick", "xbase":
		s = strings.ToUpper(s)
		if nm.family == "xbase" {
			s = trunc(s, 10)
		}
	case "mumps":
		s = trunc(strings.ToUpper(s), 8)
	case "natural":
		s = strings.ToUpper(s) + "-" + pick(nm.x, []string{"RTN", "SUB", "PROC"})
	case "ada":
		l := strings.ToLower(s)
		s = strings.ToUpper(l[:1]) + l[1:] + pick(nm.x, []string{"", "", "_Proc", "_Rtn"})
	case "pal":
		l := strings.ToLower(s)
		s = l + pick(nm.x, []string{"", "Proc", "Rtn"})
	default: // pascal vb vbnet csharp
		l := strings.ToLower(s)
		s = strings.ToUpper(l[:1]) + l[1:]
	}
	key := strings.ToUpper(s)
	for nm.used[key] {
		nm.n++
		s = trunc(s, len(s)-1) + fmt.Sprint(nm.n%10)
		key = strings.ToUpper(s)
	}
	nm.used[key] = true
	return s
}

func (nm *naming) maxLenRoutine() int {
	if nm.maxLen < 6 {
		return 6
	}
	if nm.maxLen > 8 && nm.family == "fortran" {
		return 12
	}
	return 6
}
