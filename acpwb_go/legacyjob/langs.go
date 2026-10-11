package legacyjob

import "strings"

// langMeta is everything outside the emitter that varies by compute
// language: naming family, file extension, toolchain names, how each job
// wrapper builds and runs it, and the flavour of its diagnostics.
//
// Command templates may use {src} (source file), {prog} (lower-case program
// name) and {PROG} (upper-case).
type langMeta struct {
	family string // naming family (see names.go)
	ext    string
	upper  bool // card-era language: comments folded to upper case

	// wrapper preference: weights by wrapper name for a given year
	wrappers func(year int) map[string]int

	jclProc, jclDD string // EXEC proc and compile DD name
	dcl            []string
	bat            []string
	unixBuild      string // "" = interpreted / no build step
	unixRun        string
	win            []string // Windows scheduled-task build+run lines
	dockerBase     string
	dockerPkgs     string
	dockerBuild    string

	diagTab    string
	diagUnused string // {VAR} {UNIT} {SRC} {N}
	diagDead   string // {ROUTINE} {SRC} {N}
	diagHead   []string
	diagTail   []string
}

func w(pairs ...interface{}) map[string]int {
	m := map[string]int{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i].(string)] = pairs[i+1].(int)
	}
	return m
}

func byEra(early, nineties, aughts, late map[string]int) func(int) map[string]int {
	return func(y int) map[string]int {
		switch {
		case y <= 1993:
			return early
		case y <= 2001:
			return nineties
		case y <= 2012:
			return aughts
		}
		return late
	}
}

var defaultWrappers = byEra(w("jcl", 60, "dcl", 30, "bat", 10), w("jcl", 45, "dcl", 25, "bat", 30),
	w("csh", 40, "cron", 40, "jcl", 20), w("jenkins", 50, "cron", 25, "jcl", 25))

var langMetas = map[string]*langMeta{
	"f4":  {family: "fortran", ext: ".FOR", upper: true, jclProc: "FORTGCLG", jclDD: "FORT", unixBuild: "f77 -O -o {prog} {src}", unixRun: "./{prog}"},
	"f77": {family: "fortran", ext: ".FOR", upper: true, jclProc: "FORTVCLG", jclDD: "FORT", unixBuild: "f77 -O -o {prog} {src}", unixRun: "./{prog}"},
	"gf":  {family: "fortran", ext: ".f", jclProc: "FORTVCLG", jclDD: "FORT", unixBuild: "gfortran -std=legacy -O0 -o {prog} {src}", unixRun: "./{prog}"},
	"cobol": {family: "cobol", ext: ".CBL", upper: true, jclProc: "IGYWCLG", jclDD: "COB", unixBuild: "cob -x -o {prog} {src}", unixRun: "./{prog}",
		wrappers: byEra(w("jcl", 85, "dcl", 15), w("jcl", 85, "dcl", 15), w("csh", 40, "cron", 40, "jcl", 20), w("jenkins", 50, "cron", 25, "jcl", 25))},
	"basic": {family: "basic", ext: ".BAS", upper: true, jclProc: "BASCLG", jclDD: "BAS",
		wrappers: byEra(w("dcl", 50, "bat", 50), w("dcl", 50, "bat", 50), w("bat", 100), w("bat", 100))},
	"atari": {family: "basic", ext: ".BAS", upper: true, wrappers: func(int) map[string]int { return w("atari", 100) },
		diagTab: "ATARI BASIC — LIST/SAVE", diagUnused: "ERROR-  {N}  (variable {VAR} never referenced after assignment)", diagDead: "LINE {N} UNREACHABLE ({ROUTINE})"},
	"c": {family: "c", ext: ".c", jclProc: "EDCCBG", jclDD: "C", unixBuild: "cc -O -o {prog} {src}", unixRun: "./{prog}",
		wrappers: byEra(w("dcl", 30, "bat", 30, "csh", 40), w("dcl", 30, "bat", 30, "csh", 40), w("csh", 40, "cron", 40, "jcl", 20), w("jenkins", 50, "cron", 25, "jcl", 25))},
	"cbtr": {family: "c", ext: ".C", bat: []string{"BTRIEVE /P:4096 /F:20 /M:64", "CL /AL /Ox /W3 {src} WBTRCALL.LIB"},
		win:      []string{"CL /O2 /W3 {src} WBTRV32.LIB", "{PROG}.EXE < DECK.DAT > REPORT.TXT"},
		wrappers: byEra(w("bat", 100), w("bat", 70, "wsched", 30), w("wsched", 100), w("wsched", 100))},
	"perl":  {family: "perl", ext: ".pl", jclProc: "BPXBATCH", jclDD: "GO", unixBuild: "perl -c {src}", unixRun: "/usr/bin/perl -w {src}"},
	"perl4": {family: "perl", ext: ".pl", unixBuild: "perl -c {src}", unixRun: "/usr/local/bin/perl {src}", wrappers: func(int) map[string]int { return w("csh", 60, "cron", 40) }},
	"py2":   {family: "py2", ext: ".py", jclProc: "BPXBATCH", jclDD: "GO", unixBuild: "python -m py_compile {src}", unixRun: "/usr/bin/python2.7 {src}"},
	"pick": {family: "pick", ext: "", upper: true, unixBuild: "uv \"BASIC BP {PROG}\"", unixRun: "uv \"RUN BP {PROG}\"",
		wrappers: byEra(w("pick", 100), w("pick", 70, "cron", 30), w("pick", 50, "cron", 50), w("cron", 100)),
		diagTab:  "BASIC BP " + "— COMPILER", diagUnused: "[B{N}] WARNING: Variable '{VAR}' assigned but never used.", diagDead: "[B{N}] WARNING: Label '{ROUTINE}' is never referenced.",
		diagHead: []string{"Compiling: Source = 'BP/{PROG}', Object = 'BP.O/{PROG}'"}, diagTail: []string{"", "Compilation Complete."}},
	"mumps": {family: "mumps", ext: ".m", upper: true, unixRun: "mumps -run ^{PROG}", dcl: []string{"$ DSM/UCI=RSCH/VOL=DSA", "D ^%RIMF", "D ^{PROG}"},
		wrappers: byEra(w("taskman", 60, "dcl", 40), w("taskman", 70, "cron", 30), w("taskman", 50, "cron", 50), w("cron", 60, "jenkins", 40)),
		diagTab:  "^%ZOSV / XINDEX", diagUnused: "{UNIT}+{N}  W - Variable {VAR} is SET but never referenced.", diagDead: "{ROUTINE}  W - Label is never referenced (DO or GOTO).",
		diagHead: []string{"V. A. CROSS REFERENCER (XINDEX)", "UCI: RSCH,DSA   Routines: ^{PROG}"}, diagTail: []string{"", "--- END ---"}},
	"ada": {family: "ada", ext: ".ada", unixBuild: "gnatmake -O2 {src}", unixRun: "./{prog}", dcl: []string{"$ ADA/LIST {src}", "$ ACS LINK {PROG}"},
		dockerBase: "debian:stable", dockerPkgs: "gnat", dockerBuild: "gnatmake -O2 {src}",
		wrappers: byEra(w("dcl", 60, "csh", 40), w("dcl", 50, "csh", 50), w("csh", 40, "cron", 40, "jenkins", 20), w("jenkins", 60, "cron", 40)),
		diagTab:  "ADA COMPILER", diagUnused: "{SRC}:{N}: warning: variable \"{VAR}\" is assigned but never read", diagDead: "{SRC}:{N}: warning: procedure \"{ROUTINE}\" is not referenced"},
	"pascal": {family: "pascal", ext: ".PAS", bat: []string{"TPC /B /$R+ {src}"}, dcl: []string{"$ PASCAL/LIST/CHECK=ALL {src}", "$ LINK {PROG}"},
		wrappers: byEra(w("bat", 70, "dcl", 30), w("bat", 80, "dcl", 20), w("bat", 100), w("bat", 100)),
		diagTab:  "TURBO PASCAL", diagUnused: "{SRC}({N}) Hint: Variable '{VAR}' is declared but never used in '{UNIT}'", diagDead: "{SRC}({N}) Hint: Procedure '{ROUTINE}' is not used",
		diagTail: []string{"", "{N} lines, 0.4 seconds, 18432 bytes code, 4096 bytes data."}},
	"pascalbtr": {family: "pascal", ext: ".PAS", bat: []string{"BTRIEVE /P:4096 /F:20", "TPC /B /$R+ {src}"},
		wrappers: byEra(w("bat", 100), w("bat", 100), w("bat", 100), w("bat", 100)),
		diagTab:  "TURBO PASCAL", diagUnused: "{SRC}({N}) Hint: Variable '{VAR}' is declared but never used in '{UNIT}'", diagDead: "{SRC}({N}) Hint: Procedure '{ROUTINE}' is not used"},
	"pli": {family: "pli", ext: ".PLI", upper: true, jclProc: "IBMZCBG", jclDD: "PLI", dcl: []string{"$ PLI/LIST {src}", "$ LINK {PROG}"},
		wrappers: byEra(w("jcl", 90, "dcl", 10), w("jcl", 90, "dcl", 10), w("jcl", 100), w("jcl", 80, "jenkins", 20)),
		diagTab:  "PL/I COMPILER MESSAGES", diagUnused: " IBM1213I I  {N}  The variable {VAR} is declared but not referenced.", diagDead: " IBM1215I W  {N}  The procedure {ROUTINE} is never invoked.",
		diagTail: []string{"", " File Reference Table", " Component    Return Code    Messages (Total/Suppressed)", " Compiler          4             {N}  /  0", " End of compilation of {PROG}"}},
	"pal": {family: "pal", ext: ".SC", bat: []string{"PARADOX -run {PROG}"}, win: []string{"PDOXWIN.EXE -q {PROG}.SDL"},
		wrappers: byEra(w("bat", 100), w("bat", 70, "wsched", 30), w("wsched", 100), w("wsched", 100)),
		diagTab:  "PAL DEBUGGER", diagUnused: "Script {PROG}, line {N}: variable {VAR} not used", diagDead: "Script {PROG}: procedure {ROUTINE} defined but never called"},
	"dbase": {family: "xbase", ext: ".PRG", upper: true, bat: []string{"DBASE {PROG}"},
		wrappers: func(int) map[string]int { return w("bat", 100) },
		diagTab:  "DBASE COMPILE (.DBO)", diagUnused: "Line {N}: Variable not found in procedure {UNIT}: {VAR}", diagDead: "Procedure {ROUTINE} not called"},
	"clipper": {family: "xbase", ext: ".PRG", upper: true, bat: []string{"CLIPPER {src} /M /N /W", "RTLINK FI {PROG} LIB CLIPPER,EXTEND"},
		wrappers: func(int) map[string]int { return w("bat", 100) },
		diagTab:  "CLIPPER /W", diagUnused: "{SRC}({N}) Warning C1003  Ambiguous variable reference '{VAR}'", diagDead: "{SRC}({N}) Warning C1007  Function '{ROUTINE}' is not called"},
	"foxpro": {family: "xbase", ext: ".PRG", bat: []string{"FOXPRO /T {PROG}"}, win: []string{"VFP6.EXE -T {PROG}.APP"},
		wrappers: byEra(w("bat", 100), w("bat", 50, "wsched", 50), w("wsched", 100), w("wsched", 100)),
		diagTab:  "FOXPRO COMPILE", diagUnused: "Program {PROG}, line {N}: Variable '{VAR}' is not found (warning)", diagDead: "Procedure '{ROUTINE}' is not referenced"},
	"natural": {family: "natural", ext: "", upper: true, jclProc: "NATBATCH", jclDD: "CMSYNIN",
		wrappers: byEra(w("jcl", 100), w("jcl", 100), w("jcl", 100), w("jcl", 70, "jenkins", 30)),
		diagTab:  "NATURAL CHECK / STOW", diagUnused: "NAT0{N} Field {VAR} defined but not referenced.", diagDead: "NAT0{N} Subroutine {ROUTINE} defined but not performed.",
		diagHead: []string{"CHECK {PROG}", "Library RSCH  Object {PROG}  Type PROGRAM"}, diagTail: []string{"", "STOW {PROG}", "Program {PROG} stowed in library RSCH."}},
	"informix": {family: "informix", ext: ".4gl", unixBuild: "fglpc {src}", unixRun: "fglgo {prog}",
		wrappers: func(int) map[string]int { return w("csh", 50, "cron", 50) },
		diagTab:  "fglpc", diagUnused: "| Warning: The variable '{VAR}' is defined but not used in {UNIT}.\n| See error number -{N}.", diagDead: "| Warning: The function '{ROUTINE}' is never called.\n| See error number -{N}."},
	"vb": {family: "vb", ext: ".bas", win: []string{"VB6.EXE /make {PROG}.VBP /out build.log", "{PROG}.EXE /deck=DECK.DAT /out=REPORT.TXT"},
		wrappers: func(int) map[string]int { return w("wsched", 100) },
		diagTab:  "VB6 /make build.log", diagUnused: "Module1.bas: variable {VAR} in {UNIT} is dimensioned but not used", diagDead: "Sub {ROUTINE} is not referenced in project"},
	"vba": {family: "vb", ext: ".bas", win: []string{"cscript //nologo run_macro.vbs \\\\RSCHFS01\\bench\\{PROG}.xls"},
		wrappers: func(int) map[string]int { return w("wsched", 100) },
		diagTab:  "Debug > Compile VBAProject", diagUnused: "Module1: variable '{VAR}' in {UNIT} is declared but never used (Rubberduck)", diagDead: "Module1: procedure '{ROUTINE}' is never called (Rubberduck)"},
	"vbnet": {family: "vbnet", ext: ".vb", win: []string{"vbc.exe /nologo /optimize+ /out:{PROG}.exe {src}", "{PROG}.exe < DECK.DAT > REPORT.TXT"},
		dockerBase: "mcr.microsoft.com/dotnet/sdk:6.0", dockerPkgs: "", dockerBuild: "dotnet build -c Release",
		wrappers: byEra(w("wsched", 100), w("wsched", 100), w("wsched", 70, "jenkins", 30), w("jenkins", 60, "wsched", 40)),
		diagTab:  "vbc", diagUnused: "{SRC}({N}) : warning BC42024: Unused local variable: '{VAR}'.", diagDead: "{SRC}({N}) : warning BC42305: Sub '{ROUTINE}' is never used."},
	"csharp": {family: "csharp", ext: ".cs", win: []string{"csc.exe /nologo /optimize+ /out:{PROG}.exe {src}", "{PROG}.exe < DECK.DAT > REPORT.TXT"},
		dockerBase: "mcr.microsoft.com/dotnet/sdk:6.0", dockerBuild: "dotnet build -c Release",
		wrappers: byEra(w("wsched", 100), w("wsched", 100), w("wsched", 70, "jenkins", 30), w("jenkins", 60, "wsched", 40)),
		diagTab:  "csc", diagUnused: "{SRC}({N},17): warning CS0168: The variable '{VAR}' is declared but never used", diagDead: "{SRC}({N},21): warning IDE0051: Private member '{ROUTINE}' is unused"},
	"php": {family: "php", ext: ".php", unixBuild: "php -l {src}", unixRun: "php -q {src}", dockerBase: "php:5.6-cli", dockerBuild: "php -l {src}",
		wrappers: byEra(w("cron", 100), w("cron", 100), w("cron", 70, "jenkins", 30), w("jenkins", 50, "cron", 50)),
		diagTab:  "php -l / phpcs", diagUnused: "{N} | WARNING | Variable ${VAR} is assigned but never used (in {UNIT})", diagDead: "{N} | WARNING | Function {ROUTINE}() is never called",
		diagHead: []string{"No syntax errors detected in {src}", "", "FILE: {src}", "----------------------------------------------------------------------"}},
	"plsql": {family: "sql", ext: ".sql", unixRun: "sqlplus -s rsch_batch/@RSCHPRD @{src}", dockerBase: "oraclelinux:7-slim", dockerPkgs: "oracle-instantclient-sqlplus",
		wrappers: byEra(w("csh", 50, "cron", 50), w("csh", 50, "cron", 50), w("cron", 70, "jenkins", 30), w("jenkins", 50, "cron", 50)),
		diagTab:  "SHOW ERRORS / PLW", diagUnused: "{N}/7  PLW-06009: procedure \"{UNIT}\" variable \"{VAR}\" is never used", diagDead: "{N}/1  PLW-06010: procedure \"{ROUTINE}\" is never called"},
	"sql": {family: "sql", ext: ".sql", unixRun: "isql -b rschdb < {src}", diagTab: "isql -b (warnings)",
		diagDead: "WARNING: view V_{ROUTINE} is not referenced by any query or report", diagTail: []string{"", "{N} statements executed."}, wrappers: byEra(w("cron", 60, "csh", 40), w("cron", 60, "csh", 40), w("cron", 70, "jenkins", 30), w("jenkins", 50, "cron", 50))},
	"sqlplus": {family: "sql", ext: ".sql", unixRun: "sqlplus -s rsch_batch/@RSCHPRD @{src}", diagTab: "SQL*Plus spool",
		diagDead: "SP2-0310: view V_{ROUTINE} created but never selected (audit note)", diagHead: []string{"SQL*Plus: Release 8.1.7.0.0 - Production", "Connected to: Oracle8i Enterprise Edition"}, diagTail: []string{"", "Disconnected from Oracle8i Enterprise Edition"}, wrappers: byEra(w("csh", 50, "cron", 50), w("csh", 50, "cron", 50), w("cron", 100), w("cron", 100))},
	"tsql": {family: "sql", ext: ".sql", unixRun: "isql -U rsch_batch -S RSCHSQL -i {src}", diagTab: "sqlcmd output",
		diagDead: "Msg 15335, Level 0, State 1: The object 'v_{ROUTINE}' is not referenced.", diagTail: []string{"(1 rows affected)", "Changed database context to 'Benchmarks'."}, win: []string{"sqlcmd -S RSCHSQL01 -d Benchmarks -E -b -i {src} -o REPORT.TXT"},
		wrappers: byEra(w("wsched", 100), w("wsched", 60, "cron", 40), w("wsched", 100), w("wsched", 60, "jenkins", 40))},
	"sqlpl": {family: "sql", ext: ".sql", unixRun: "db2 -td@ -vf {src}", diagTab: "db2 -td@ -vf",
		diagDead: "SQL0280W  View, trigger or materialized query table \"RSCH.V_{ROUTINE}\" has replaced an existing alias.  SQLSTATE=01595", diagTail: []string{"DB20000I  The SQL command completed successfully."}, wrappers: byEra(w("cron", 100), w("cron", 50, "jcl", 50), w("cron", 50, "jcl", 50), w("jenkins", 40, "cron", 30, "jcl", 30))},
	"db2sql": {family: "sql", ext: ".sql", upper: true, unixRun: "db2 -tvf {src}", diagTab: "DSNTEP2 OUTPUT",
		diagDead: "DSNT404I SQLCODE = 562, WARNING:  A GRANT OF A PRIVILEGE WAS IGNORED (V_{ROUTINE})", diagHead: []string{"PAGE    1", "***INPUT STATEMENT:"}, diagTail: []string{"", "SUCCESSFUL RETRIEVAL OF {N} ROW(S)"}, jclProc: "DSNTEP2", jclDD: "SYSIN", wrappers: byEra(w("jcl", 100), w("jcl", 100), w("jcl", 70, "cron", 30), w("jcl", 60, "jenkins", 40))},
	"sqlds": {family: "sql", ext: ".sql", upper: true, diagTab: "SQL/DS ISQL", diagDead: "ARI0505I SQL PROCESSING WAS SUCCESSFUL. VIEW V_{ROUTINE} NOT REFERENCED.", jclProc: "DSNTEP2", jclDD: "SYSIN", wrappers: func(int) map[string]int { return w("jcl", 100) }},
}

func metaOf(key string) *langMeta {
	if m, ok := langMetas[key]; ok {
		return m
	}
	return langMetas["f77"]
}

// expand fills {src} {prog} {PROG} in a command template.
func expand(t, src string, id ID) string {
	return strings.NewReplacer("{src}", src, "{prog}", strings.ToLower(id.Program), "{PROG}", strings.ToUpper(id.Program)).Replace(t)
}
