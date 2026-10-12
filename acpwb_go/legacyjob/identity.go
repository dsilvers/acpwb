package legacyjob

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
)

// ID is the small, article-visible identity of a page's batch job: what the
// caption under each table, the sidebar card and the references entry show.
//
// It is the ONLY part of this package mirrored in Python
// (acpwb/apps/honeypot/legacy_identity.py) — the article HTML is
// byte-compared against Django's pyrender output, so any change to the
// tables below must be made in both places and the archive fixtures
// re-dumped.
type ID struct {
	Program   string // BNCHMK
	Release   string // 2.4A
	LangKey   string // f77
	Language  string // FORTRAN 77
	JobNumber string // JOB04471
}

type langOpt struct {
	key, label string
	from, to   int
	weight     int
}

// languageTable: every compute language with the archive years it was in
// service and its weight among the languages live that year. A key may
// appear on several rows (different label / weight per period). Add a
// language here AND in the Python mirror together, then re-dump fixtures.
var languageTable = []langOpt{
	{"f4", "FORTRAN IV", 1985, 1993, 12},
	{"f77", "FORTRAN 77", 1985, 2001, 18},
	{"f77", "FORTRAN 77", 2002, 2012, 10},
	{"gf", "Fortran (gfortran)", 2013, 2025, 14},
	{"cobol", "COBOL", 1985, 2001, 20},
	{"cobol", "COBOL", 2002, 2025, 6},
	{"basic", "BASIC", 1985, 1993, 10},
	{"atari", "Atari BASIC", 1985, 1991, 5},
	{"pick", "Pick BASIC", 1985, 2008, 6},
	{"mumps", "MUMPS", 1985, 2025, 5},
	{"ada", "Ada 83", 1985, 1995, 6},
	{"ada", "Ada 95", 1996, 2012, 4},
	{"ada", "Ada 2012", 2013, 2025, 2},
	{"pascal", "Turbo Pascal", 1985, 1998, 7},
	{"pascalbtr", "Turbo Pascal + Btrieve", 1986, 2001, 4},
	{"pli", "PL/I", 1985, 2005, 8},
	{"pli", "Enterprise PL/I", 2006, 2025, 2},
	{"pal", "Paradox PAL", 1987, 1999, 4},
	{"dbase", "dBASE IV", 1989, 1994, 5},
	{"clipper", "Clipper", 1988, 1999, 5},
	{"foxpro", "FoxPro", 1991, 2007, 4},
	{"natural", "Natural (Adabas)", 1985, 2025, 5},
	{"informix", "Informix 4GL", 1987, 2008, 5},
	{"c", "C", 1985, 2001, 16},
	{"c", "C", 2002, 2025, 8},
	{"cbtr", "C + Btrieve", 1988, 2004, 4},
	{"vb", "Visual Basic 3.0", 1993, 1997, 5},
	{"vb", "Visual Basic 6.0", 1998, 2012, 6},
	{"vba", "Excel VBA", 1994, 2025, 7},
	{"vbnet", "VB.NET", 2002, 2025, 6},
	{"csharp", "C#", 2002, 2025, 7},
	{"perl4", "Perl 4", 1991, 1996, 5},
	{"perl", "Perl 5", 1995, 2025, 14},
	{"php", "PHP 4", 2000, 2004, 6},
	{"php", "PHP 5", 2005, 2015, 8},
	{"php", "PHP 7", 2016, 2025, 6},
	{"py2", "Python 2", 2001, 2012, 10},
	{"py2", "Python 2.7", 2013, 2025, 10},
	{"sqlds", "SQL/DS", 1985, 1995, 3},
	{"db2sql", "DB2 SQL", 1985, 2025, 6},
	{"sqlplus", "Oracle SQL*Plus", 1988, 2012, 5},
	{"plsql", "PL/SQL", 1992, 2025, 6},
	{"tsql", "Transact-SQL", 1993, 2025, 6},
	{"sqlpl", "DB2 SQL PL", 1999, 2025, 4},
	{"sql", "SQL", 1990, 2025, 5},
	{"rpg", "RPG II", 1985, 1990, 4},
	{"rpg", "RPG/400", 1988, 1996, 5},
	{"rpg", "ILE RPG IV", 1995, 2012, 4},
	{"rpg", "RPG IV (free-form)", 2013, 2025, 3},
	{"rexx", "REXX", 1985, 2025, 5},
	{"clist", "TSO CLIST", 1985, 2001, 3},
	{"cl400", "AS/400 CL", 1988, 2007, 3},
	{"cl400", "IBM i CL", 2008, 2025, 2},
	{"ezt", "Easytrieve Plus", 1985, 2025, 4},
	{"abap", "ABAP/4", 1993, 2012, 4},
	{"abap", "ABAP", 2013, 2025, 4},
	{"apl", "VS APL", 1985, 1988, 2},
	{"apl", "APL2", 1989, 2004, 2},
	{"apl", "Dyalog APL", 2005, 2025, 2},
	{"focus", "FOCUS", 1985, 2010, 3},
	{"nomad", "NOMAD", 1985, 1998, 2},
	{"ramis", "RAMIS II", 1985, 1994, 2},
	{"sas", "SAS", 1985, 2025, 6},
	{"spss", "SPSS-X", 1985, 1993, 3},
	{"spss", "SPSS", 1994, 2025, 3},
	{"splus", "S-PLUS", 1988, 2007, 2},
	{"r", "R", 2000, 2025, 4},
	{"matlab", "MATLAB", 1988, 2025, 3},
	{"octave", "GNU Octave", 1997, 2025, 2},
	{"stata", "Stata", 1986, 2025, 3},
	{"awk", "nawk", 1988, 1995, 4},
	{"awk", "gawk", 1996, 2025, 4},
	{"ksh", "Korn shell", 1988, 2025, 4},
	{"tcl", "Tcl", 1991, 2025, 3},
	{"dc", "dc", 1985, 2025, 1},
	{"bc", "bc", 1985, 2025, 2},
	{"cfml", "ColdFusion", 1996, 2025, 3},
	{"asp", "Classic ASP", 1997, 2025, 3},
	{"modula2", "Modula-2", 1985, 2000, 2},
	{"lisp", "Franz Lisp", 1985, 1989, 2},
	{"lisp", "Common Lisp", 1990, 2025, 2},
	{"prolog", "Turbo Prolog", 1986, 1991, 2},
	{"prolog", "Quintus Prolog", 1992, 2000, 1},
	{"prolog", "SWI-Prolog", 2001, 2025, 1},
	{"forth", "Forth-83", 1985, 1995, 2},
	{"forth", "ANS Forth", 1996, 2008, 1},
	{"smalltalk", "Smalltalk-80", 1985, 1995, 1},
	{"smalltalk", "VisualWorks Smalltalk", 1996, 2012, 2},
}

// eraLanguages returns the weighted pool for an article year, in table order.
func eraLanguages(year int) []langOpt {
	var out []langOpt
	lo, hi := languageTable[0].from, languageTable[0].to
	for _, l := range languageTable {
		if year >= l.from && year <= l.to {
			out = append(out, l)
		}
		lo, hi = min(lo, l.from), max(hi, l.to)
	}
	// Years outside every language's service span (e.g. a 1984 archive page)
	// borrow the nearest covered year rather than leaving Identity nothing to
	// pick from. Mirrored in apps/honeypot/legacy_identity.py:_era_languages.
	if len(out) == 0 && (year < lo || year > hi) {
		return eraLanguages(max(lo, min(year, hi)))
	}
	return out
}

// LanguageKeys lists every distinct compute-language key (for tests).
func LanguageKeys() []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range languageTable {
		if !seen[l.key] {
			seen[l.key] = true
			out = append(out, l.key)
		}
	}
	return out
}

var programNames = []string{
	"BNCHMK", "PCTLAD", "ENGTAB", "STATPK", "CMPBNK", "RWDTAB", "MKTPOS", "PAYLAD",
	"SURVTB", "BMKRPT", "QTLCLC", "ENGMET", "DLTCLC", "TABGEN", "PCTRPT", "WGSTAT",
	"COMPRA", "BENCHX", "KPITAB", "MEDLAD", "PERCNT", "SALSRV", "TRWTAB", "EQUTAB",
	"MRKBNK", "LADDER", "RPTPCT", "ENGDLT", "CMPTAB", "STPCTL", "BNKSUM", "QRTILE",
	"PAYSTA", "WAGRPT", "MKTLAD", "SVYTAB", "DELTAS", "PCTGEN", "BMKTAB", "RSCHTB",
}

// Identity is pure md5 arithmetic (no RNG state) so it is trivial to mirror
// in Python exactly.
func Identity(year, month, day int, slug string) ID {
	sum := md5.Sum([]byte(fmt.Sprintf("legacyid_%d_%d_%d_%s", year, month, day, slug)))
	h := hex.EncodeToString(sum[:])
	n := func(i int) uint64 {
		v, _ := strconv.ParseUint(h[i*8:(i+1)*8], 16, 64)
		return v
	}
	n0, n1, n2, n3 := n(0), n(1), n(2), n(3)

	langs := eraLanguages(year)
	total := 0
	for _, l := range langs {
		total += l.weight
	}
	p := int(n0 % uint64(total))
	lang := langs[len(langs)-1]
	for _, l := range langs {
		if p < l.weight {
			lang = l
			break
		}
		p -= l.weight
	}

	release := fmt.Sprintf("%d.%d", 1+n2%4, (n2/4)%10)
	if (n2/40)%3 == 0 {
		release += string(rune('A' + (n2/120)%3))
	}

	return ID{
		Program:   programNames[n1%uint64(len(programNames))],
		Release:   release,
		LangKey:   lang.key,
		Language:  lang.label,
		JobNumber: fmt.Sprintf("JOB%05d", 10000+n3%90000),
	}
}

// LanguageLive reports whether language key k is in service in year.
func LanguageLive(k string, year int) bool {
	for _, l := range eraLanguages(year) {
		if l.key == k {
			return true
		}
	}
	return false
}

// LangSpan is one language label and its service years, for index pages.
type LangSpan struct {
	Key, Label, Category string
	From, To, Weight     int
}

var langCategory = map[string]string{
	"f4": "Scientific", "f77": "Scientific", "gf": "Scientific", "ada": "Scientific", "pascal": "Scientific", "pli": "Mainframe & business",
	"cobol": "Mainframe & business", "natural": "Mainframe & business", "mumps": "Mainframe & business", "pick": "Mainframe & business",
	"basic": "Desktop & PC", "atari": "Desktop & PC", "pascalbtr": "Desktop & PC", "cbtr": "Desktop & PC", "pal": "Desktop & PC",
	"dbase": "Desktop & PC", "clipper": "Desktop & PC", "foxpro": "Desktop & PC", "vb": "Desktop & PC", "vba": "Desktop & PC",
	"informix": "Database & 4GL", "sql": "Database & 4GL", "db2sql": "Database & 4GL", "sqlds": "Database & 4GL", "sqlplus": "Database & 4GL",
	"plsql": "Database & 4GL", "tsql": "Database & 4GL", "sqlpl": "Database & 4GL",
	"c": "Systems & scripting", "perl4": "Systems & scripting", "perl": "Systems & scripting", "py2": "Systems & scripting",
	"php": "Systems & scripting", "vbnet": "Systems & scripting", "csharp": "Systems & scripting",
	"rpg": "Mainframe & business", "rexx": "Mainframe & business", "clist": "Mainframe & business",
	"cl400": "Mainframe & business", "ezt": "Mainframe & business", "abap": "Mainframe & business",
	"focus": "Database & 4GL", "nomad": "Database & 4GL", "ramis": "Database & 4GL",
	"sas": "Statistics & analytics", "spss": "Statistics & analytics", "splus": "Statistics & analytics", "r": "Statistics & analytics",
	"matlab": "Statistics & analytics", "octave": "Statistics & analytics", "stata": "Statistics & analytics", "apl": "Statistics & analytics",
	"awk": "Systems & scripting", "ksh": "Systems & scripting", "tcl": "Systems & scripting", "dc": "Systems & scripting", "bc": "Systems & scripting",
	"cfml": "Systems & scripting", "asp": "Systems & scripting", "modula2": "Scientific",
	"lisp": "AI & research", "prolog": "AI & research", "forth": "AI & research", "smalltalk": "AI & research",
}

// LanguageCategories is the display order of LanguageSpans categories.
var LanguageCategories = []string{"Scientific", "Statistics & analytics", "Mainframe & business", "Database & 4GL", "Desktop & PC", "Systems & scripting", "AI & research"}

// LanguageSpans merges table rows that share a label (e.g. FORTRAN 77's two
// periods) into one span each.
func LanguageSpans() []LangSpan {
	var out []LangSpan
	idx := map[string]int{}
	for _, l := range languageTable {
		if i, ok := idx[l.label]; ok {
			if l.to > out[i].To {
				out[i].To = l.to
			}
			if l.weight > out[i].Weight {
				out[i].Weight = l.weight
			}
			continue
		}
		idx[l.label] = len(out)
		out = append(out, LangSpan{Key: l.key, Label: l.label, Category: langCategory[l.key], From: l.from, To: l.to, Weight: l.weight})
	}
	return out
}
