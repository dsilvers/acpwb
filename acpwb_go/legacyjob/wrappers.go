package legacyjob

import (
	"fmt"
	"strings"
)

type wrapperOut struct {
	name     string
	jobName  string
	pipeline []string
	sections []Section
}

// srcFileName is the conventional file / member name for the compute source.
func srcFileName(id ID, wrapper string) string {
	ext := metaOf(id.LangKey).ext
	name := id.Program
	switch wrapper {
	case "jcl":
		return "ACPWB.RSCH.SOURCE(" + name + ")"
	case "pick":
		return "BP " + name
	case "taskman":
		return "^" + name
	case "as400":
		return "RSCHLIB/QRPGSRC(" + name + ")"
	case "sapjob":
		return "Z" + name
	case "cfsched":
		return "/rsch/bench/" + strings.ToLower(name) + ext
	case "atari":
		return "D1:" + trunc(name, 8) + ext
	case "dcl", "bat", "wsched":
		return name + strings.ToUpper(ext)
	default:
		return strings.ToLower(name) + strings.ToLower(ext)
	}
}

func chooseWrapper(x *rnd, year int, lang string) string {
	wf := metaOf(lang).wrappers
	if wf == nil {
		wf = defaultWrappers
	}
	m := wf(year)
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sortStrings(names)
	ws := make([]int, len(names))
	for i, n := range names {
		ws[i] = m[n]
	}
	return names[x.weighted(ws)]
}

func buildWrapper(j *job, l *Listing, src []Line) wrapperOut {
	x := j.x
	w := chooseWrapper(x, j.in.Year, j.id.LangKey)
	out := wrapperOut{name: w}
	jobName := trunc("ACPWB"+j.id.Program[:3], 8)
	if x.chance(0.4) {
		jobName = trunc(j.id.Program+pick(x, []string{"RN", "Q", "X", "J1", "PR"}), 8)
	}
	out.jobName = jobName
	j.paths = buildPaths(j)
	srcName := srcFileName(j.id, w)
	if w == "jcl" {
		srcName = j.paths.dsn("SOURCE(" + j.id.Program + ")")
	}
	compute := j.id.Language + " " + j.id.Program
	out.pipeline = []string{compute}

	var ctl, log []string
	var ctlTab, logTab string
	switch w {
	case "jcl":
		ctl, log = jclWrapper(j, l, jobName)
		ctlTab, logTab = "JCL", "JES2 JOB LOG"
	case "dcl":
		ctl, log = dclWrapper(j, l, jobName, srcName)
		ctlTab, logTab = strings.ToUpper(jobName)+".COM", "BATCH LOG"
	case "bat":
		ctl, log = batWrapper(j, l, srcName)
		ctlTab, logTab = "RUN"+j.id.Program[:3]+".BAT", "CONSOLE"
	case "csh":
		ctl, log = cshWrapper(j, l, srcName)
		ctlTab, logTab = "Makefile + run_"+strings.ToLower(j.id.Program)+".csh", "make output"
	case "cron":
		ctl, log = cronWrapper(j, l, srcName)
		ctlTab, logTab = "crontab + run_"+strings.ToLower(j.id.Program)+".sh", "cron mail"
	case "atari":
		ctl, log = atariWrapper(j, l, srcName)
		ctlTab, logTab = "AUTORUN + DOS 2.5 MENU", "SCREEN"
	case "wsched":
		ctl, log = wschedWrapper(j, l, srcName)
		ctlTab, logTab = "Scheduled Task + RUN"+j.id.Program[:3]+".CMD", "task history"
	case "pick":
		ctl, log = pickWrapper(j, l, srcName)
		ctlTab, logTab = "PROC "+j.id.Program+".RUN", "TCL SESSION"
	case "taskman":
		ctl, log = taskmanWrapper(j, l, srcName)
		ctlTab, logTab = "TASKMAN OPTION + ^"+j.id.Program, "TASK LOG"
	case "as400":
		ctl, log = as400Wrapper(j, l, srcName)
		if j.id.Language == "RPG II" {
			ctlTab, logTab = "OCL PROCEDURE "+j.id.Program, "HISTORY FILE"
		} else {
			ctlTab, logTab = "SBMJOB + CL", "JOB LOG (DSPJOBLOG)"
		}
	case "sapjob":
		ctl, log = sapjobWrapper(j, l, srcName)
		ctlTab, logTab = "SM36 JOB DEFINITION", "SM37 JOB LOG"
	case "cfsched":
		ctl, log = cfschedWrapper(j, l, srcName)
		ctlTab, logTab = "CF ADMINISTRATOR — SCHEDULED TASK", "scheduler.log"
	default:
		ctl, log = jenkinsWrapper(j, l, srcName)
		ctlTab, logTab = "Jenkinsfile + Dockerfile", "console output"
	}

	source := Section{ID: "source", Title: "Source listing", Tab: fmt.Sprintf("%s — %d LINES", srcName, len(src)), Kind: "pre", Text: linesText(src)}
	listing, compOut := compilerSections(j, l, src, srcName)
	if listing != nil {
		source = *listing
	}
	out.sections = append(out.sections,
		Section{ID: "jobcontrol", Title: "Job control", Tab: ctlTab, Kind: "pre", Text: strings.Join(ctl, "\n")},
		source)
	out.sections = append(out.sections, compOut...)
	out.sections = append(out.sections,
		Section{ID: "sysin", Title: "Input deck", Tab: inputTab(w), Kind: "pre", Text: strings.Join(j.rep.Deck, "\n")},
		Section{ID: "sysprint", Title: "Output", Tab: outputTab(w), Kind: "pre", Text: strings.Join(j.rep.Printed, "\n")},
		Section{ID: "joblog", Title: "Job log", Tab: logTab, Kind: "pre", Text: strings.Join(log, "\n")},
	)
	return out
}

func inputTab(w string) string {
	switch w {
	case "jcl":
		return "GO.SYSIN DD *"
	case "dcl":
		return "FOR005 / SYSIN.DAT"
	case "bat", "wsched":
		return "DECK.DAT"
	case "atari":
		return "D1:DECK.DAT"
	case "pick":
		return "FILE DECK"
	case "taskman":
		return "DECK.DAT (HFS)"
	case "as400":
		return "RSCHLIB/DECK"
	case "sapjob":
		return "/usr/sap/RSCH/deck (AL11)"
	case "cfsched":
		return "deck.dat (cffile)"
	default:
		return "deck.dat (stdin)"
	}
}

func outputTab(w string) string {
	switch w {
	case "jcl":
		return "GO.SYSPRINT — SYSOUT=A"
	case "dcl":
		return "FOR006 / REPORT.LIS"
	case "bat", "wsched":
		return "REPORT.TXT"
	case "atari":
		return "P: (ATARI 1025)"
	case "pick":
		return "SPOOLER HOLD FILE"
	case "taskman":
		return "DEVICE: RSCH PRINTER"
	case "as400":
		return "QSYSPRT SPOOLED FILE"
	case "sapjob":
		return "SPOOL LIST (SP01)"
	case "cfsched":
		return "report.txt (published)"
	default:
		return "report.lis (stdout)"
	}
}

// siblingSteps returns cross-linked sibling programs (other archive pages'
// jobs) as (programName, markerText) pairs.
func siblingSteps(j *job, l *Listing, n int) []struct{ prog, marker string } {
	var out []struct{ prog, marker string }
	for _, cl := range sample(j.x, j.in.Links, n) {
		id := Identity(cl.Year, cl.Month, cl.Day, cl.Slug)
		idx := l.addLink(cl.URL)
		l.Related = append(l.Related, cl)
		out = append(out, struct{ prog, marker string }{id.Program, linkMarker(idx, id.Program)})
	}
	return out
}

func jclWrapper(j *job, l *Listing, jobName string) ([]string, []string) {
	x := j.x
	in := j.in
	rep := j.rep
	acct := fmt.Sprintf("(%04d,%s,%s)", x.between(1000, 9999), pick(x, []string{"RSCH", "COMP", "BMRK", "RD01", "CLNT"}), strings.ToUpper(in.RecordID))
	class := string(rune('A' + x.intn(6)))
	region := pick(x, []string{"2M", "4M", "512K", "1024K", "8M", "0M"})
	meta := metaOf(j.id.LangKey)
	proc := meta.jclProc
	if proc == "" {
		proc = "BPXBATCH"
	}
	if j.id.LangKey == "cobol" && in.Year < 1990 {
		proc = "COBUCLG"
	}
	sub := j.g.coder()
	var ctl []string
	ctl = append(ctl,
		fmt.Sprintf("//%-8s JOB %s,'%s RUN',CLASS=%s,MSGCLASS=%s,", jobName, acct, j.id.Program, class, pick(x, []string{"X", "A", "H", "T"})),
		fmt.Sprintf("//             NOTIFY=%s,REGION=%s,TIME=(%d,%02d)", pick(x, []string{"&SYSUID", sub.UserID}), region, x.between(0, 2), x.between(10, 59)),
	)
	if x.chance(0.6) {
		ctl = append(ctl, fmt.Sprintf("/*JOBPARM ROOM=%d,LINES=%d,FORMS=%s", x.between(100, 999), x.between(5, 99), pick(x, []string{"STD", "1PRT", "GRNB", "WIDE"})))
	}
	if x.chance(0.5) {
		ctl = append(ctl, fmt.Sprintf("/*ROUTE PRINT RMT%d", x.between(1, 40)))
	}
	box := func(lines ...string) {
		ctl = append(ctl, "//*"+strings.Repeat("*", 66))
		for _, ln := range lines {
			for _, w := range wrapWords(cardText(ln), 62) {
				ctl = append(ctl, "//* "+padRight(w, 64)+"*")
			}
		}
		ctl = append(ctl, "//*"+strings.Repeat("*", 66))
	}
	box(
		"ACPWB RESEARCH DIVISION -- "+j.id.Program+" BENCHMARK TABULATION",
		"ENGAGEMENT : "+in.Org,
		"ENG CODE   : "+in.EngCode,
		"RECORD     : "+strings.ToUpper(in.RecordID)+"   PROVENANCE: "+strings.ToUpper(in.Token),
		"SUBMITTED  : "+l.Submitted,
		j.g.comment(nil, "warnings", "superstition", "memos"),
	)
	cmpDD := meta.jclDD
	if cmpDD == "" {
		cmpDD = "GO"
	}
	switch j.id.LangKey {
	case "natural":
		ctl = append(ctl, "//STEP010  EXEC PGM=NATBAT"+pick(x, []string{"23", "31", "42", "82"})+",REGION=0M,",
			"//             PARM='IM=D,MADIO=0,MAXCL=0,INTENS=1'",
			"//STEPLIB  DD DSN=SAG.NAT"+pick(x, []string{"231", "319", "424", "826"})+".LOAD,DISP=SHR",
			"//DDCARD   DD *", "ADARUN DB="+fmt.Sprint(x.between(1, 250))+",MODE=MULTI,SVC=249", "/*",
			"//CMPRINT  DD SYSOUT=*", "//CMWKF01  DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//CMSYNIN  DD *", "LOGON RSCH", strings.ToUpper(j.id.Program), "FIN", "/*")
	case "rexx", "clist":
		lib, cmd := "SYSEXEC", "%"+j.id.Program
		if j.id.LangKey == "clist" {
			lib = "SYSPROC"
		}
		ctl = append(ctl, "//STEP010  EXEC PGM=IKJEFT01,DYNAMNBR=30,REGION=4M",
			"//"+padRight(lib, 8)+" DD DSN="+j.paths.dsn("EXEC")+",DISP=SHR",
			"//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//SYSTSPRT DD SYSOUT=*", "//SYSTSIN  DD *", cmd, "/*")
	case "ezt":
		ctl = append(ctl, "//STEP010  EXEC PGM=EZTPA00,REGION=2M",
			"//STEPLIB  DD DSN=SYS1.EZTP.LOADLIB,DISP=SHR",
			"//EZTVFM   DD UNIT=SYSDA,SPACE=(CYL,(5,5))",
			"//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//SYSPRINT DD SYSOUT=*", "//SYSOUT   DD SYSOUT=*")
	case "sas":
		ctl = append(ctl, "//STEP010  EXEC SAS,OPTIONS='NOCENTER LS=132'",
			"//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//WORK     DD UNIT=SYSDA,SPACE=(CYL,(20,10))")
	case "spss":
		ctl = append(ctl, "//STEP010  EXEC SPSSX,REGION=6M",
			"//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//SPSSWORK DD UNIT=SYSDA,SPACE=(CYL,(5,5))")
	case "focus", "nomad", "ramis":
		pgm := map[string]string{"focus": "FOCUS", "nomad": "NOMAD2", "ramis": "RAMIS"}[j.id.LangKey]
		ctl = append(ctl, "//STEP010  EXEC PGM="+pgm+",REGION=4M",
			"//STEPLIB  DD DSN=SYS2."+pgm+".LOAD,DISP=SHR",
			"//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR",
			"//MASTER   DD DSN="+j.paths.dsn("MASTER")+",DISP=SHR",
			"//FOCEXEC  DD DSN="+j.paths.dsn("FOCEXEC")+",DISP=SHR",
			"//OFFLINE  DD SYSOUT=*", "//SYSIN    DD *", "EX "+j.id.Program, "FIN", "/*")
	case "apl":
		ctl = append(ctl, "//STEP010  EXEC PGM=APL2,REGION=8M,PARM='WS(4M) QUIET'",
			"//STEPLIB  DD DSN=APL2.SAPLLOAD,DISP=SHR",
			"//SYSIN    DD *", ")LOAD 1 "+j.id.Program, j.id.Program, ")OFF HOLD", "//DECK     DD DSN="+j.paths.dsn("DECK")+",DISP=SHR", "//SYSPRINT DD SYSOUT=*")
	case "db2sql", "sqlds":
		ctl = append(ctl, "//STEP010  EXEC PGM=IKJEFT01,DYNAMNBR=20", "//SYSTSPRT DD SYSOUT=*", "//SYSTSIN  DD *",
			" DSN SYSTEM("+pick(x, []string{"DB2P", "DSNP", "DB2A"})+")",
			" RUN PROGRAM(DSNTEP2) PLAN(DSNTEP"+pick(x, []string{"22", "41", "71", "81"})+") -",
			"     LIB('DSN"+pick(x, []string{"410", "510", "710", "810"})+".RUNLIB.LOAD') PARMS('/ALIGN(MID)')", " END", "/*", "//SYSPRINT DD SYSOUT=*")
	default:
		ctl = append(ctl, fmt.Sprintf("//STEP010  EXEC %s,PARM.%s='%s'", proc, cmpDD,
			pick(x, []string{"OPT(2),XREF,MAP", "XREF,LIST,SOURCE", "NOOPT,MAP", "OPT(3),SOURCE,XREF", "LIST,MAP,OFFSET"})))
	}
	if langFamily(j.id.LangKey) == "perl" || j.id.LangKey == "py2" {
		interp := "/usr/lpp/perl/bin/perl"
		if j.id.LangKey == "py2" {
			interp = "/usr/lpp/IBM/cyp/v2r7/pyz/bin/python2"
		}
		ctl = append(ctl, fmt.Sprintf("//*  PARM='SH %s /u/%s/%s'", interp, strings.ToLower(j.paths.tso), strings.ToLower(j.id.Program)))
	}
	switch j.id.LangKey {
	case "natural", "rexx", "clist", "focus", "nomad", "ramis", "apl":
	case "ezt", "sas", "spss":
		ctl = append(ctl, fmt.Sprintf("//SYSIN    DD DSN=%s,DISP=SHR", j.paths.dsn("SOURCE("+j.id.Program+")")))
	case "db2sql", "sqlds":
		ctl = append(ctl, fmt.Sprintf("//SYSIN    DD DSN=%s,DISP=SHR", j.paths.dsn("SQL("+j.id.Program+")")))
	default:
		ctl = append(ctl, fmt.Sprintf("//%s.SYSIN DD DSN=%s,DISP=SHR", cmpDD, j.paths.dsn("SOURCE("+j.id.Program+")")))
	}
	for n := x.between(1, 4); n > 0; n-- {
		switch x.intn(5) {
		case 0:
			ctl = append(ctl, fmt.Sprintf("//GO.WORK%02d DD UNIT=SYSDA,SPACE=(CYL,(%d,%d)),DISP=(NEW,DELETE)", n, x.between(1, 20), x.between(1, 5)))
		case 1:
			ctl = append(ctl, fmt.Sprintf("//GO.TAPEIN DD DSN=%s.G%04dV00,DISP=OLD,", j.paths.dsn("PEERLIST"), x.between(1, 9999)),
				fmt.Sprintf("//             UNIT=TAPE,VOL=SER=%s%04d,LABEL=(1,SL)", pick(x, []string{"AC", "RD", "TP"}), x.between(1, 9999)))
		case 2:
			ctl = append(ctl, fmt.Sprintf("//GO.FT%02dF001 DD DUMMY", x.between(7, 20)))
		case 3:
			ctl = append(ctl, fmt.Sprintf("//GO.SYSUDUMP DD SYSOUT=%s", pick(x, []string{"A", "X", "*"})))
		default:
			ctl = append(ctl, fmt.Sprintf("//GO.HISTORY DD DSN=%s.Y%04d,DISP=MOD,UNIT=SYSDA", j.paths.dsn("HIST"), in.Year))
		}
	}
	ctl = append(ctl, "//GO.SYSIN DD *", fmt.Sprintf("//*  (%d CARD IMAGES -- SEE INPUT DECK)", len(rep.Deck)), "/*")
	ctl = append(ctl, fmt.Sprintf("//GO.SYSPRINT DD SYSOUT=%s,DCB=(RECFM=FBA,LRECL=133,BLKSIZE=%d)", pick(x, []string{"A", "*", "X"}), 133*x.between(10, 60)))
	sibs := siblingSteps(j, l, x.between(1, 3))
	for i, s := range sibs {
		ctl = append(ctl, fmt.Sprintf("//*  NEXT STEP: SEE %s FOR THE DOWNSTREAM TABULATION", strings.ToUpper(s.prog)))
		ctl = append(ctl, fmt.Sprintf("//STEP%03d  EXEC PGM=%s,COND=(4,LT)", (i+2)*10, s.marker))
		ctl = append(ctl, "//STEPLIB  DD DSN="+j.paths.dsn("LOADLIB")+",DISP=SHR")
	}
	ctl = append(ctl, "//")

	// JES2 log
	h, m, s := x.between(0, 23), x.between(0, 59), x.between(0, 59)
	t := func() string {
		s += x.between(0, 9)
		if s >= 60 {
			s -= 60
			m++
		}
		if m >= 60 {
			m -= 60
			h = (h + 1) % 24
		}
		return fmt.Sprintf("%02d.%02d.%02d", h, m, s)
	}
	jn := l.ID.JobNumber
	sys := pick(x, []string{"ACP1", "SYSA", "SYSB", "MKE1", "PRD1"})
	var log []string
	log = append(log,
		fmt.Sprintf("                    J E S 2  J O B  L O G  --  S Y S T E M  %s  --  N O D E  %s", spaced(sys), pick(x, []string{"MKEVM01", "ACPWBN1", "RSCHNODE", "N1"})),
		"",
		fmt.Sprintf("%s %s  $HASP373 %-8s STARTED - INIT %d - CLASS %s - SYS %s", t(), jn, jobName, x.between(1, 12), class, sys),
		fmt.Sprintf("%s %s  IEF403I %-8s - STARTED - TIME=%s", t(), jn, jobName, t()),
	)
	if x.chance(0.4) {
		log = append(log, fmt.Sprintf("%s %s  *IEC501A M %X,%s%04d,SL,6250 BPI,%s", t(), jn, x.between(0x180, 0x5FF), pick(x, []string{"AC", "RD"}), x.between(1, 9999), jobName))
		log = append(log, fmt.Sprintf("%s %s  IEC705I TAPE ON %X,%s%04d,SL,6250 BPI,%s", t(), jn, x.between(0x180, 0x5FF), pick(x, []string{"AC", "RD"}), x.between(1, 9999), jobName))
	}
	for _, step := range []string{"FORT", "LKED", "GO"} {
		log = append(log, fmt.Sprintf("%s %s  IEF142I %-8s STEP010 %s - STEP WAS EXECUTED - COND CODE %04d", t(), jn, jobName, step, pick(x, []int{0, 0, 0, 4})))
	}
	log = append(log, "------ JES2 JOB STATISTICS ------",
		fmt.Sprintf("  %02d %s %04d JOB EXECUTION DATE", in.Day, monthAbbr[in.Month-1], in.Year),
		fmt.Sprintf("  %10d CARDS READ", len(rep.Deck)+l.SrcLines),
		fmt.Sprintf("  %10d SYSOUT PRINT RECORDS", len(rep.Printed)+x.between(200, 2000)),
		fmt.Sprintf("  %10d SYSOUT PUNCH RECORDS", 0),
		fmt.Sprintf("  %10d SYSOUT SPOOL KBYTES", x.between(40, 900)),
		fmt.Sprintf("  %10.2f MINUTES EXECUTION TIME", float64(x.between(1, 300))/100),
		"",
	)
	for i, card := range ctl {
		log = append(log, fmt.Sprintf("%10d %s", i+1, stripMarkers(card)))
	}
	log = append(log, "",
		fmt.Sprintf("IEF236I ALLOC. FOR %s GO STEP010", jobName),
		fmt.Sprintf("IEF237I JES2 ALLOCATED TO SYSIN"),
		fmt.Sprintf("IEF237I JES2 ALLOCATED TO SYSPRINT"),
		fmt.Sprintf("IEF142I %s STEP010 GO - STEP WAS EXECUTED - COND CODE %s", jobName, l.CondCode),
		fmt.Sprintf("IEF285I   %s.%s.%s.JOB%s.D0000101.?         SYSOUT", sub.UserID, jobName, jn, jn[3:]),
		fmt.Sprintf("IEF373I STEP/GO      /START %04d%03d.%02d%02d", in.Year, doy(in.Year, in.Month, in.Day), h, m),
		fmt.Sprintf("IEF374I STEP/GO      /STOP  %04d%03d.%02d%02d CPU    0MIN %sSEC SRB    0MIN 00.%02dSEC VIRT %4dK SYS %4dK", in.Year, doy(in.Year, in.Month, in.Day), h, (m+1)%60, padLeft(l.CPU, 5), x.between(1, 40), x.between(64, 4096), x.between(200, 900)),
		fmt.Sprintf("IEF375I  JOB/%-8s/START %04d%03d.%02d%02d", jobName, in.Year, doy(in.Year, in.Month, in.Day), h, m),
		fmt.Sprintf("%s %s  $HASP395 %-8s ENDED", t(), jn, jobName),
	)
	if x.chance(0.3) {
		log = append(log, fmt.Sprintf("%s %s  +OPERATOR NOTE: %s", t(), jn, cardText(j.g.comment(nil, "superstition", "hardware", "misc"))))
	}
	return ctl, log
}

func doy(y, m, d int) int {
	cum := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	n := cum[m-1] + d
	if m > 2 && y%4 == 0 && (y%100 != 0 || y%400 == 0) {
		n++
	}
	return n
}

func dclWrapper(j *job, l *Listing, jobName, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	low := strings.ToUpper(j.id.Program)
	compile := map[string]string{"fortran": "$ FORTRAN/LIST/CROSS_REFERENCE/CHECK=BOUNDS " + srcName, "cobol": "$ COBOL/LIST/MAP " + srcName,
		"basic": "$ BASIC/LIST " + srcName, "c": "$ CC/LIST/DEBUG " + srcName}[langFamily(j.id.LangKey)]
	if m := metaOf(j.id.LangKey); len(m.dcl) > 0 {
		var cs []string
		for _, c := range m.dcl {
			cs = append(cs, expand(c, srcName, j.id))
		}
		compile = strings.Join(cs, "\n")
	}
	if compile == "" {
		compile = "$ FORTRAN/LIST " + srcName
	}
	c := j.g.coder()
	ctl := []string{
		"$ ! " + strings.Repeat("-", 66),
		"$ ! " + cardText(jobName+".COM -- "+j.id.Program+" BENCHMARK TABULATION FOR "+in.Org),
		"$ ! " + cardText("RECORD "+strings.ToUpper(in.RecordID)+"  SEAL "+strings.ToUpper(in.Token)+"  ENGAGEMENT "+in.EngCode),
	}
	for _, w := range wrapWords(cardText(j.g.comment(c, "warnings", "handoff", "superstition")), 64) {
		ctl = append(ctl, "$ ! "+w)
	}
	ctl = append(ctl, "$ ! "+strings.Repeat("-", 66),
		"$ SET NOON",
		"$ SET VERIFY",
		"$ ON ERROR THEN GOTO ABORT",
		"$ SET DEFAULT "+j.paths.vmsDir)
	ctl = append(ctl, strings.Split(compile, "\n")...)
	ctl = append(ctl,
		fmt.Sprintf("$ LINK/MAP/FULL %s", low),
		"$ ASSIGN SYSIN.DAT FOR005",
		"$ ASSIGN REPORT.LIS FOR006",
		fmt.Sprintf("$ RUN %s", low),
		"$ DEASSIGN FOR005",
		"$ DEASSIGN FOR006",
		fmt.Sprintf("$ PRINT/QUEUE=%s/FORM=%s/NOTIFY REPORT.LIS", pick(x, []string{"SYS$PRINT", "PRT3", "LPA0", "RSCH_LP"}), pick(x, []string{"GREENBAR", "WIDE", "DEFAULT"})),
	)
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "$ ! CHAIN TO "+strings.ToUpper(s.prog), "$ SUBMIT/NOPRINT/AFTER=\"+0:05\" "+s.marker+".COM")
	}
	ctl = append(ctl, "$ EXIT", "$ ABORT:", "$ MAIL/SUBJECT=\""+j.id.Program+" FAILED\" NL: "+c.UserID, "$ EXIT 44")

	var log []string
	log = append(log, fmt.Sprintf("$ Set NOON  ! Job %s (queue SYS$BATCH, entry %d) started on %s", jobName, x.between(100, 999), "SYS$BATCH"))
	for _, cl := range ctl {
		if strings.HasPrefix(cl, "$ !") {
			continue
		}
		log = append(log, stripMarkers(cl))
	}
	log = append(log,
		fmt.Sprintf("%%LINK-I-MAPFILE, map file %s.MAP created", low),
		fmt.Sprintf("Job %s queued to %s", jobName, "PRT3"),
		fmt.Sprintf("  %s      job terminated at %02d-%s-%04d %02d:%02d:%02d.%02d", c.UserID, in.Day, monthAbbr[in.Month-1], in.Year, x.between(0, 23), x.between(0, 59), x.between(0, 59), x.between(0, 99)),
		"",
		"  Accounting information:",
		fmt.Sprintf("  Buffered I/O count:          %6d         Peak working set size:  %6d", x.between(40, 900), x.between(400, 9000)),
		fmt.Sprintf("  Direct I/O count:            %6d         Peak page file size:    %6d", x.between(10, 400), x.between(1000, 20000)),
		fmt.Sprintf("  Page faults:                 %6d         Mounted volumes:        %6d", x.between(200, 5000), 0),
		fmt.Sprintf("  Charged CPU time:   0 00:00:%s          Elapsed time:  0 00:%02d:%02d.%02d", padLeft(l.CPU, 5), x.between(0, 9), x.between(0, 59), x.between(0, 99)),
	)
	return ctl, log
}

func batWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	exe := j.id.Program + ".EXE"
	compile := map[string][]string{
		"basic": {"BC /O /E /X " + srcName + ";", "LINK /EX " + j.id.Program + ".OBJ;"},
		"c":     {"CL /AL /Ox /W3 /c " + srcName, "LINK " + j.id.Program + ".OBJ,,NUL,LLIBCE.LIB;"},
		"cobol": {"COBOL " + srcName + " ANIM;", "LINK " + j.id.Program + ";"},
	}[langFamily(j.id.LangKey)]
	if m := metaOf(j.id.LangKey); len(m.bat) > 0 {
		compile = nil
		for _, c := range m.bat {
			compile = append(compile, expand(c, srcName, j.id))
		}
	}
	if compile == nil {
		compile = []string{"FL /FPi /Fs " + srcName}
	}
	c := j.g.coder()
	ctl := []string{"@ECHO OFF", "REM " + strings.Repeat("=", 66),
		"REM " + cardText("RUN"+j.id.Program[:3]+".BAT -- "+in.Org+" BENCHMARK TABULATION"),
		"REM " + cardText("RECORD "+strings.ToUpper(in.RecordID)+"  SEAL "+strings.ToUpper(in.Token))}
	for _, w := range wrapWords(cardText(j.g.comment(c)), 64) {
		ctl = append(ctl, "REM "+w)
	}
	ctl = append(ctl, "REM "+strings.Repeat("=", 66),
		fmt.Sprintf("SET LIB=C:\\%s\\LIB", pick(x, []string{"QB45", "MSC", "COBOL", "FORTRAN"})),
		"SET INCLUDE="+j.paths.dosDir[:2]+"\\INCLUDE",
		j.paths.dosDir[:2],
		"CD "+j.paths.dosDir[2:],
		"IF NOT EXIST DECK.DAT GOTO NODECK")
	ctl = append(ctl, compile...)
	ctl = append(ctl,
		"IF ERRORLEVEL 1 GOTO FAILED",
		exe+" < DECK.DAT > REPORT.TXT",
		"COPY REPORT.TXT LPT1:",
		fmt.Sprintf("COPY REPORT.TXT %s\\%s.TXT", j.paths.dosShare, strings.ToUpper(in.RecordID)))
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "REM NEXT: "+strings.ToUpper(s.prog), "CALL RUN"+s.marker+".BAT")
	}
	ctl = append(ctl, "GOTO END", ":NODECK", "ECHO DECK.DAT NOT FOUND. ASK "+cardText(c.Display(x))+".", "GOTO END", ":FAILED", "ECHO COMPILE FAILED. DO NOT PANIC.", ":END")

	log := []string{
		fmt.Sprintf("%s>RUN%s", j.paths.dosDir, j.id.Program[:3]),
		fmt.Sprintf("Microsoft (R) %s  Version %d.%02d", pick(x, []string{"QuickBASIC Compiler", "C Optimizing Compiler", "FORTRAN Optimizing Compiler", "Overlay Linker"}), x.between(3, 6), x.between(0, 99)),
		"Copyright (C) Microsoft Corp 1982-" + fmt.Sprint(mini(in.Year, 1999)) + ".  All rights reserved.",
		"",
		fmt.Sprintf("%d Bytes Available", x.between(200000, 600000)),
		fmt.Sprintf("%d Bytes Free", x.between(100000, 400000)),
		"",
		"0 Warning Error(s)",
		"0 Severe  Error(s)",
		"        1 file(s) copied",
		"        1 file(s) copied",
		"",
		fmt.Sprintf(" Volume in drive %s is %s", j.paths.dosDir[:1], j.paths.dosVolume),
		fmt.Sprintf(" REPORT   TXT     %5d %02d-%02d-%02d  %2d:%02dp", len(strings.Join(j.rep.Printed, "\r\n")), in.Month, in.Day, in.Year%100, x.between(1, 12), x.between(0, 59)),
	}
	return ctl, log
}

func cshWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToLower(j.id.Program)
	meta := metaOf(j.id.LangKey)
	cc := strings.Fields(expand(meta.unixBuild, srcName, j.id) + " true")[0]
	if meta.unixBuild == "" {
		cc = strings.Fields(expand(meta.unixRun, srcName, j.id) + " sh")[0]
	}
	c := j.g.coder()
	ctl := []string{
		"# Makefile for " + prog + " -- ACPWB Research Division",
		"# " + j.g.comment(c, "warnings", "todo", "workarounds"),
		"",
		"FC      = " + cc,
		"FFLAGS  = " + pick(x, []string{"-O", "-O2 -g", "-fast", "-xO3", "-g -C"}),
		"DECK    = deck.dat",
		"REPORT  = report.lis",
		"SEAL    = " + in.Token,
		"",
		"all: " + prog,
		"",
		prog + ": " + srcName,
		"\t$(FC) $(FFLAGS) -o " + prog + " " + srcName,
		"",
		"run: " + prog,
		"\t" + runLine(j, srcName) + " < $(DECK) > $(REPORT)",
		"\tlp -d " + pick(x, []string{"rsch_lp", "lw3", "prt3", "hp4si"}) + " -o landscape $(REPORT)",
		"",
		"clean:",
		"\trm -f " + prog + " *.o core $(REPORT)",
		"",
		"# ---- run_" + prog + ".csh ----",
		"#!/bin/csh -f",
		"# " + j.g.comment(nil, "superstition", "handoff"),
		"setenv ACPWB_RECORD " + strings.ToUpper(in.RecordID),
		"cd " + j.paths.unixWork,
		"make run >& make.log",
		"if ($status != 0) then",
		"    mail -s \"" + prog + " failed\" " + j.g.email(c) + " < make.log",
		"    exit 1",
		"endif",
	}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "# downstream: "+strings.ToLower(s.prog), "../"+s.marker+"/run.csh")
	}
	log := []string{
		fmt.Sprintf("%s %s %d %02d:%02d:%02d CST %d", pick(x, dayNames)[:3], strings.Title(strings.ToLower(monthAbbr[in.Month-1])), in.Day, x.between(0, 23), x.between(0, 59), x.between(0, 59), in.Year),
		cc + " -O -o " + prog + " " + srcName,
	}
	if langFamily(j.id.LangKey) == "fortran" {
		log = append(log, srcName+":", " MAIN "+strings.ToLower(j.id.Program)+":")
		for _, p := range j.p.Procs[1:] {
			log = append(log, "\t"+strings.ToLower(p.Name)+":")
		}
	}
	log = append(log, "./"+prog+" < deck.dat > report.lis",
		"lp -d prt3 -o landscape report.lis",
		fmt.Sprintf("request id is prt3-%d (1 file)", x.between(100, 9999)))
	return ctl, log
}

func cronWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToLower(j.id.Program)
	interp := runLine(j, srcName)
	if j.id.LangKey == "py2" && in.Year < 2010 {
		interp = "/usr/bin/python2.4 " + srcName
	}
	c := j.g.coder()
	ctl := []string{
		"# crontab -l (user " + j.paths.user + " on " + j.paths.host + ")",
		"MAILTO=" + j.g.email(c),
		fmt.Sprintf("%d %d * * %d  %s/bin/run_%s.sh >> %s/%s.log 2>&1", x.between(0, 59), x.between(1, 5), x.between(1, 5), j.paths.unixHome, prog, j.paths.unixLog, prog),
		"",
		"# ---- " + j.paths.unixHome + "/bin/run_" + prog + ".sh ----",
		"#!/bin/sh",
		"# " + j.g.comment(c, "warnings", "handoff", "todo"),
		"set -e",
		"RECORD=" + strings.ToUpper(in.RecordID),
		"SEAL=" + in.Token,
		"WORK=" + j.paths.unixWork,
		"cd $WORK",
		"echo \"$(date) starting " + prog + " for $RECORD\"",
		interp + " < deck.dat > report.lis",
		"lpr -P " + pick(x, []string{"rsch_hp", "lj4000", "floor3"}) + " report.lis",
		"cp report.lis " + j.paths.unixShare + "/$RECORD.lis",
	}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "# hand off to "+strings.ToLower(s.prog), j.paths.unixHome+"/bin/run_"+s.marker+".sh || true")
	}
	ctl = append(ctl, "echo \"$(date) done\"")
	log := []string{
		"From: Cron Daemon <root@" + j.paths.host + ">",
		"To: " + j.g.email(c),
		"Subject: Cron <" + j.paths.user + "@" + j.paths.host + "> " + j.paths.unixHome + "/bin/run_" + prog + ".sh",
		"",
		fmt.Sprintf("%s %s %2d %02d:%02d:%02d CST %d starting %s for %s", pick(x, dayNames)[:3], strings.Title(strings.ToLower(monthAbbr[in.Month-1])), in.Day, x.between(0, 23), x.between(0, 59), x.between(0, 59), in.Year, prog, strings.ToUpper(in.RecordID)),
	}
	if j.id.LangKey == "py2" && in.Year >= 2015 {
		log = append(log, "DEPRECATION: Python 2.7 will reach the end of its life on January 1st, 2020. Please upgrade your Python as Python 2.7 is no longer maintained.")
	}
	log = append(log, fmt.Sprintf("%s %s %2d %02d:%02d:%02d CST %d done", pick(x, dayNames)[:3], strings.Title(strings.ToLower(monthAbbr[in.Month-1])), in.Day, x.between(0, 23), x.between(0, 59), x.between(0, 59), in.Year))
	return ctl, log
}

func jenkinsWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToLower(j.id.Program)
	meta := metaOf(j.id.LangKey)
	base := map[string]string{"fortran": "ubuntu:14.04", "perl": "perl:5.20", "py2": "python:2.7-slim", "c": "gcc:4.9", "cobol": "centos:6"}[langFamily(j.id.LangKey)]
	if meta.dockerBase != "" {
		base = meta.dockerBase
	}
	if base == "" {
		base = "debian:stretch"
	}
	if in.Year >= 2018 && x.chance(0.5) {
		base = pick(x, []string{"ubuntu:18.04", "debian:stretch", "centos:7", "registry.acpwb.internal/legacy/base:" + fmt.Sprint(in.Year-1)})
	}
	pkgs := map[string]string{"fortran": "gfortran", "perl": "libtext-csv-perl", "py2": "python2.7", "c": "build-essential", "cobol": "open-cobol"}[langFamily(j.id.LangKey)]
	if meta.dockerPkgs != "" || meta.dockerBase != "" {
		pkgs = meta.dockerPkgs
	}
	if pkgs == "" {
		pkgs = "ca-certificates"
	}
	runCmd := runLine(j, srcName)
	c := j.g.coder()
	ctl := []string{
		"// Jenkinsfile -- " + prog + " (record " + strings.ToUpper(in.RecordID) + ")",
		"// " + j.g.comment(c, "todo", "warnings", "handoff"),
		"pipeline {",
		"    agent { label '" + pick(x, []string{"legacy-batch", "linux-docker", "rsch-node-02", "mainframe-bridge"}) + "' }",
		"    options { timeout(time: " + fmt.Sprint(x.between(5, 60)) + ", unit: 'MINUTES'); buildDiscarder(logRotator(numToKeepStr: '" + fmt.Sprint(x.between(5, 50)) + "')) }",
		"    environment {",
		"        RECORD = '" + strings.ToUpper(in.RecordID) + "'",
		"        SEAL   = '" + in.Token + "'",
		"        // TODO(" + strings.ToLower(c.Initials) + "): move to credentials store",
		"        PRINTER = '" + pick(x, []string{"rsch-hp-3", "floor3-mfp", "ricoh-c4504"}) + "'",
		"    }",
		"    stages {",
		"        stage('Checkout') { steps { checkout scm } }",
		"        stage('Build image') { steps { sh 'docker build -t acpwb/" + prog + ":${BUILD_NUMBER} .' } }",
		"        stage('Tabulate') {",
		"            steps {",
		"                sh 'docker run --rm -i acpwb/" + prog + ":${BUILD_NUMBER} < deck.dat > report.lis'",
		"            }",
		"        }",
		"        stage('Archive') { steps { archiveArtifacts artifacts: 'report.lis', fingerprint: true } }",
	}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "        stage('Downstream') { steps { build job: '"+s.marker+"', wait: false } }")
	}
	ctl = append(ctl, "    }", "    post { failure { mail to: '"+j.g.email(c)+"', subject: '"+prog+" failed', body: \"${env.BUILD_URL}\" } }", "}",
		"",
		"# ---- Dockerfile ----",
		"FROM "+base,
		"LABEL maintainer=\""+j.g.email(j.g.coder())+"\"",
		"# "+j.g.comment(nil, "vendors", "workarounds", "history"),
		"RUN apt-get update && apt-get install -y --no-install-recommends "+pkgs+" && rm -rf /var/lib/apt/lists/*",
		"WORKDIR "+j.paths.container,
		"COPY "+srcName+" .",
	)
	if langFamily(j.id.LangKey) == "fortran" {
		ctl = append(ctl, "RUN gfortran -std=legacy -O0 -o "+prog+" "+srcName+"   # -O0: see the comment about Thursdays")
	} else if j.id.LangKey == "c" {
		ctl = append(ctl, "RUN cc -O -o "+prog+" "+srcName)
	} else if meta.dockerBuild != "" {
		ctl = append(ctl, "RUN "+expand(meta.dockerBuild, srcName, j.id))
	}
	ctl = append(ctl, "ENTRYPOINT [\""+strings.Join(strings.Fields(runCmd), "\", \"")+"\"]")

	bn := x.between(100, 9000)
	log := []string{
		fmt.Sprintf("Started by timer"),
		fmt.Sprintf("Running in Durability level: MAX_SURVIVABILITY"),
		"[Pipeline] Start of Pipeline",
		"[Pipeline] node",
		fmt.Sprintf("Running on %s in %s", j.paths.host, j.paths.jenkinsWS),
		"[Pipeline] { (Checkout)",
		fmt.Sprintf(" > git rev-parse --resolve-git-dir %s/.git", j.paths.jenkinsWS),
		fmt.Sprintf("Checking out Revision %s (refs/remotes/origin/master)", strings.ToLower(j.g.hexToken())+in.Token),
		"[Pipeline] { (Build image)",
		fmt.Sprintf("+ docker build -t acpwb/%s:%d .", prog, bn),
		"Step 1/7 : FROM " + base,
		"Step 4/7 : RUN apt-get update && apt-get install -y --no-install-recommends " + pkgs,
		" ---> Using cache",
		fmt.Sprintf("Successfully tagged acpwb/%s:%d", prog, bn),
		"[Pipeline] { (Tabulate)",
		fmt.Sprintf("+ docker run --rm -i acpwb/%s:%d", prog, bn),
	}
	if j.id.LangKey == "py2" {
		log = append(log, "DEPRECATION: Python 2.7 reached the end of its life on January 1st, 2020.")
	}
	if langFamily(j.id.LangKey) == "fortran" && x.chance(0.6) {
		log = append(log, srcName+":"+fmt.Sprint(x.between(40, 900))+":72:", "Warning: Nonconforming tab character at (1) [-Wtabs]")
	}
	log = append(log, "[Pipeline] { (Archive)", "Archiving artifacts", "Recording fingerprints", "[Pipeline] End of Pipeline", "Finished: SUCCESS")
	return ctl, log
}

// runLine is how a Unix-style wrapper runs the compute program.
func runLine(j *job, srcName string) string {
	m := metaOf(j.id.LangKey)
	if m.unixRun != "" {
		return expand(m.unixRun, srcName, j.id)
	}
	return "./" + strings.ToLower(j.id.Program)
}

// atariWrapper: an Atari 800 with a 1050 drive, DOS 2.5 and a 1025 printer.
func atariWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	c := j.g.coder()
	ctl := []string{
		"REM ---- AUTORUN.SYS (built with the DOS 2.5 AUTORUN builder) ----",
		"REM " + cardText("BOOT DISK FOR "+j.id.Program+" -- "+in.Org),
		"REM " + cardText("RECORD "+strings.ToUpper(in.RecordID)+"  SEAL "+strings.ToUpper(in.Token)),
	}
	for _, w := range wrapWords(cardText(j.g.comment(c, "superstition", "hardware", "operators")), 36) {
		ctl = append(ctl, "REM "+w)
	}
	ctl = append(ctl, "", "RUN \""+srcName+"\"", "",
		"REM ---- OPERATOR CARD (TAPED TO THE 1050) ----",
		"1. TURN ON 1050 DRIVE, THEN 1025 PRINTER, THEN 800 WITH BASIC CARTRIDGE IN.",
		"2. INSERT DISK LABELED "+strings.ToUpper(j.id.Program)+" IN DRIVE 1.",
		"3. IF THE SCREEN SAYS ERROR- 138, THE PRINTER IS OFF. SEE STEP 1.",
		"4. DO NOT TYPE NEW. EVER. ("+cardText(c.Initials)+")")
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "5. THEN BOOT THE "+s.marker+" DISK.")
	}
	log := []string{"READY", "RUN \"" + srcName + "\"", "", fmt.Sprintf("%d FREE SECTORS", x.between(30, 600)), "READY", ""}
	if x.chance(0.4) {
		log = append(log, fmt.Sprintf("ERROR-  %d AT LINE %d", pick(x, []int{138, 139, 144, 170}), x.between(10, 900)), "READY", "RUN", "READY")
	}
	return ctl, log
}

// wschedWrapper: Windows scheduled task (AT on NT, Task Scheduler later)
// running a .CMD file.
func wschedWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	m := metaOf(j.id.LangKey)
	c := j.g.coder()
	cmd := "RUN" + j.id.Program[:3] + ".CMD"
	dir := `D:\RSCH\` + j.id.Program
	host := pick(x, []string{"RSCHAPP01", "BENCH-NT4", "MKE-BATCH2", "RSCHSQL01", strings.ToUpper(trunc(j.paths.user, 8)) + "-PC"})
	var ctl []string
	if in.Year < 2001 {
		ctl = []string{fmt.Sprintf(`C:\>AT \\%s %02d:%02d /every:M,T,W,Th,F "%s\%s"`, host, x.between(1, 5), x.between(0, 59), dir, cmd),
			"Added a new job with job ID = " + fmt.Sprint(x.between(1, 40))}
	} else {
		ctl = []string{fmt.Sprintf(`schtasks /create /s %s /tn "ACPWB %s" /tr "%s\%s" /sc weekly /d MON,TUE,WED,THU,FRI /st %02d:%02d:00 /ru RSCH\svc_batch`, host, j.id.Program, dir, cmd, x.between(1, 5), x.between(0, 59)),
			`SUCCESS: The scheduled task "ACPWB ` + j.id.Program + `" has successfully been created.`}
	}
	ctl = append(ctl, "", "REM ---- "+dir+`\`+cmd+" ----", "@ECHO OFF",
		"REM "+asciiText(j.g.comment(c, "warnings", "operators", "handoff")),
		"SET RECORD="+strings.ToUpper(in.RecordID),
		"SET SEAL="+in.Token,
		"CD /D "+dir)
	for _, w := range m.win {
		ctl = append(ctl, expand(w, srcName, j.id))
	}
	if j.id.LangKey == "vba" {
		ctl = append(ctl, "", "' ---- run_macro.vbs ----",
			`Set xl = CreateObject("Excel.Application")`,
			"xl.DisplayAlerts = False",
			"Set wb = xl.Workbooks.Open(WScript.Arguments(0))",
			`xl.Run "Run`+strings.Title(strings.ToLower(j.id.Program))+`"`,
			`wb.Worksheets("Report").SaveAs "`+dir+`\REPORT.TXT", -4158`,
			"wb.Close False",
			"xl.Quit")
	}
	ctl = append(ctl, "IF ERRORLEVEL 1 GOTO FAILED", `COPY REPORT.TXT \\RSCHFS01\REPORTS\%RECORD%.TXT`)
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, `CALL ..\`+s.marker+`\RUN.CMD`)
	}
	ctl = append(ctl, "GOTO END", ":FAILED", "NET SEND "+strings.ToUpper(trunc(alnum(c.Last), 8))+` "`+j.id.Program+` FAILED"`, ":END")
	log := []string{
		fmt.Sprintf(`"ACPWB %s" (%s) %d/%d/%d %d:%02d:%02d AM`, j.id.Program, cmd, in.Month, in.Day, in.Year, x.between(1, 5), x.between(0, 59), x.between(0, 59)),
		"    Started",
		fmt.Sprintf(`"ACPWB %s" (%s) Finished %d/%d/%d %d:%02d:%02d AM`, j.id.Program, cmd, in.Month, in.Day, in.Year, x.between(1, 5), x.between(0, 59), x.between(0, 59)),
		"    Result: The task completed with an exit code of (0).",
	}
	return ctl, log
}

// pickWrapper: a Pick PROC that compiles and runs the program, plus the TCL
// session it produced.
func pickWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToUpper(j.id.Program)
	ctl := []string{"PQ", "C " + cardText(prog+".RUN -- "+in.Org), "C " + cardText("RECORD "+strings.ToUpper(in.RecordID)+" SEAL "+strings.ToUpper(in.Token)),
		"HBASIC BP " + prog, "P", "HCATALOG BP " + prog, "P", "HSP-ASSIGN HS", "P", "H" + prog, "P", "HSP-ASSIGN", "P"}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "H"+s.marker+".RUN", "P")
	}
	ctl = append(ctl, "X"+cardText("DONE. SEE HOLD FILE "+fmt.Sprint(x.between(100, 999))+"."))
	log := []string{
		":LOGTO RSCH", ":" + prog + ".RUN",
		"Compiling: Source = 'BP/" + prog + "', Object = 'BP.O/" + prog + "'",
		strings.Repeat("*", x.between(10, 60)),
		"Compilation Complete.",
		"[241] '" + prog + "' cataloged.",
		fmt.Sprintf("Entry #%d", x.between(10, 999)),
		fmt.Sprintf("[%d] Hold file entry created.", x.between(100, 999)),
		":",
	}
	return ctl, log
}

// taskmanWrapper: a VA-style TaskMan option queueing the MUMPS routine.
func taskmanWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToUpper(j.id.Program)
	c := j.g.coder()
	ctl := []string{
		"OPTION NAME: RSCH " + prog + " NIGHTLY",
		"MENU TEXT: " + cardText(trunc(in.Title, 40)),
		"TYPE: run routine",
		"ROUTINE: EN^" + prog,
		"SCHEDULING RECOMMENDED: YES",
		fmt.Sprintf("QUEUED TO RUN AT WHAT TIME: T+1@%02d%02d", x.between(1, 5), x.between(0, 59)),
		"RESCHEDULING FREQUENCY: 1D",
		"DEVICE FOR QUEUED JOB OUTPUT: RSCH PRINTER;P-OTHER;132;66",
		"DESCRIPTION: " + cardText(j.g.comment(c, "operators", "warnings")),
		"",
		"RECORD: " + strings.ToUpper(in.RecordID) + "   SEAL: " + strings.ToUpper(in.Token),
	}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "FOLLOWED BY OPTION: RSCH "+s.marker+" NIGHTLY")
	}
	log := []string{
		fmt.Sprintf("Task #%d  EN^%s  -- RSCH %s NIGHTLY", x.between(100000, 999999), prog, prog),
		fmt.Sprintf("  Queued: %s %d, %04d@%02d:%02d   By: %s", strings.Title(strings.ToLower(monthAbbr[in.Month-1])), in.Day, in.Year, x.between(1, 5), x.between(0, 59), cardText(c.Last)),
		"  Device: RSCH PRINTER   Volume set: ROU   UCI: RSCH",
		"  Started running: " + fmt.Sprintf("%02d:%02d:%02d", x.between(1, 5), x.between(0, 59), x.between(0, 59)),
		"  Completed.  Elapsed time: " + fmt.Sprintf("%d sec", x.between(1, 90)),
	}
	return ctl, log
}
