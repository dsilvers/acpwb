package legacyjob

import (
	"fmt"
	"strings"
)

// as400Wrapper: System/36 OCL for RPG II, otherwise SBMJOB of a CL driver
// with the job log it left behind.
func as400Wrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToUpper(j.id.Program)
	c := j.g.coder()
	user := strings.ToUpper(trunc(alnum(c.Last), 8))
	if j.id.Language == "RPG II" {
		ctl := []string{
			"// * " + cardText(prog+" -- "+in.Org),
			"// * RECORD " + strings.ToUpper(in.RecordID) + "  SEAL " + strings.ToUpper(in.Token),
			"// RPGC " + prog + ",RSCHLIB",
			"// LOAD " + prog + ",RSCHLIB",
			"// FILE NAME-DECK,LABEL-" + prog + "DK,DISP-SHR",
			"// PRINTER NAME-QSYSPRT,LINES-66,CPI-10",
			"// RUN",
		}
		for _, s := range siblingSteps(j, l, x.between(1, 2)) {
			ctl = append(ctl, "// INCLUDE "+s.marker+",RSCHLIB")
		}
		log := []string{
			fmt.Sprintf("%02d/%02d/%02d %02d.%02d.%02d  RPGC %s  COMPILE COMPLETE  0 TERMINAL ERRORS", in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59), prog),
			fmt.Sprintf("%02d/%02d/%02d %02d.%02d.%02d  LOAD %s  JOB %s  USER %s", in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59), prog, l.ID.JobNumber, user),
			"SYS-1230 END OF JOB",
		}
		return ctl, log
	}
	jobq := pick(x, []string{"QBATCH", "QPGMR", "RSCHBATCH", "QNIGHT"})
	ctl := []string{
		fmt.Sprintf("SBMJOB CMD(CALL PGM(RSCHLIB/%sCL) PARM('N')) JOB(%s) JOBQ(%s) USER(%s) +", prog[:minInt(len(prog), 6)], prog, jobq, user),
		fmt.Sprintf("       JOBD(RSCHLIB/BATCH) SCDDATE(*CURRENT) SCDTIME('%02d%02d00') LOG(4 00 *SECLVL)", x.between(1, 5), x.between(0, 59)),
		"",
		"/* ===== RSCHLIB/QCLSRC(" + prog[:minInt(len(prog), 6)] + "CL) ===== */",
		"             PGM        PARM(&DEBUG)",
		"             DCL        VAR(&DEBUG) TYPE(*CHAR) LEN(1)",
		"/* " + padRight(cardText("RECORD "+strings.ToUpper(in.RecordID)+"  SEAL "+strings.ToUpper(in.Token)), 60) + " */",
		"             OVRDBF     FILE(DECK) TOFILE(RSCHLIB/DECK) MBR(" + prog + ")",
		"             OVRPRTF    FILE(QSYSPRT) PAGESIZE(66 132) OUTQ(RSCHPRT) HOLD(*YES)",
	}
	if j.id.LangKey == "rpg" {
		cmd := "CRTRPGPGM"
		if j.id.Language != "RPG/400" {
			cmd = "CRTBNDRPG"
		}
		ctl = append(ctl, fmt.Sprintf("             %-10s PGM(RSCHLIB/%s) SRCFILE(RSCHLIB/QRPGSRC)", cmd, prog))
	} else {
		ctl = append(ctl, fmt.Sprintf("             CRTCLPGM   PGM(RSCHLIB/%s) SRCFILE(RSCHLIB/QCLSRC)", prog))
	}
	ctl = append(ctl, fmt.Sprintf("             CALL       PGM(RSCHLIB/%s) PARM(&DEBUG)", prog),
		"             MONMSG     MSGID(CPF0000) EXEC(SNDMSG MSG('"+prog+" FAILED') TOUSR("+user+"))")
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "             CALL       PGM(RSCHLIB/"+s.marker+")")
	}
	ctl = append(ctl, "             DLTOVR     FILE(*ALL)", "             ENDPGM")
	jn := fmt.Sprintf("%06d/%s/%s", x.between(100000, 999999), user, prog)
	log := []string{
		"                                        Display Job Log",
		fmt.Sprintf(" Job . . : %-10s  User . . : %-10s  Number . . . : %s", prog, user, strings.SplitN(jn, "/", 2)[0]),
		"",
		fmt.Sprintf(" CPI1125  Information   Job %s submitted.", jn),
		fmt.Sprintf(" CPF1124  Information   Job %s started on %02d/%02d/%02d at %02d:%02d:%02d in subsystem QBATCH.", jn, in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59)),
		" CPC2205  Completion    Override for file DECK in RSCHLIB.",
		fmt.Sprintf(" CPF1164  Completion    Job %s ended on %02d/%02d/%02d; %d seconds used; end code 0 .", jn, in.Month, in.Day, in.Year%100, x.between(1, 40)),
	}
	return ctl, log
}

// sapjobWrapper: an SM36 background job definition and the SM37 log.
func sapjobWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := "Z" + strings.ToUpper(j.id.Program)
	c := j.g.coder()
	user := strings.ToUpper(trunc(alnum(c.Last), 12))
	jobName := "ZRSCH_" + strings.ToUpper(j.id.Program) + "_NIGHTLY"
	count := fmt.Sprintf("%02d%02d%02d%02d", x.between(0, 23), x.between(0, 59), x.between(0, 59), x.between(1, 99))
	ctl := []string{
		"Job name ........... " + jobName,
		"Job class .......... " + pick(x, []string{"A", "B", "C"}),
		"Target server ...... " + pick(x, []string{"rschapp01_RSP_00", "sapbtc02_RSP_01", "mkesap01_PRD_00"}),
		"",
		"Step 1: ABAP program",
		"  Program ......... " + prog,
		"  Variant ......... " + pick(x, []string{"NIGHTLY", "DEFAULT", "RD114"}),
		"  User ............ " + user,
		"  Language ........ EN",
		"  Print parameters: LOCL / X_65_132, spool retention 8 days",
		"",
		"Start condition: date/time " + fmt.Sprintf("%02d.%02d.%04d %02d:%02d:00", in.Day, in.Month, in.Year, x.between(1, 5), x.between(0, 59)) + ", periodic: daily",
		"",
		"Record " + strings.ToUpper(in.RecordID) + "  Seal " + strings.ToUpper(in.Token),
		"Note: " + asciiText(j.g.comment(c, "operators", "warnings")),
	}
	for i, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, fmt.Sprintf("Step %d: ABAP program Z%s", i+2, s.marker))
	}
	t := func() string {
		return fmt.Sprintf("%02d.%02d.%04d %02d:%02d:%02d", in.Day, in.Month, in.Year, x.between(1, 5), x.between(0, 59), x.between(0, 59))
	}
	log := []string{
		"Job log overview for job: " + jobName + " / " + count,
		"",
		t() + " Job started                                              00 516 S",
		t() + " Step 001 started (program " + prog + ", variant NIGHTLY, user ID " + user + ")   00 550 S",
		t() + " Spool request (number " + fmt.Sprint(x.between(10000, 99999)) + ") created without immediate output   SPOOL 371 S",
		t() + " Job finished                                             00 517 S",
	}
	return ctl, log
}

// cfschedWrapper: a ColdFusion Administrator scheduled task requesting the
// template, publishing the output to a file.
func cfschedWrapper(j *job, l *Listing, srcName string) ([]string, []string) {
	x := j.x
	in := j.in
	prog := strings.ToLower(j.id.Program)
	host := pick(x, []string{"rschweb01", "intranet", "bench.acpwb.local", "mkecf02"})
	ctl := []string{
		"Task Name ............ RSCH " + strings.ToUpper(j.id.Program) + " nightly",
		"Duration ............. Start Date " + fmt.Sprintf("%d/%d/%d", in.Month, in.Day, in.Year),
		"Frequency ............ Daily every 1 day at " + fmt.Sprintf("%d:%02d AM", x.between(1, 5), x.between(0, 59)),
		"URL .................. http://" + host + "/bench/" + prog + ".cfm?record=" + strings.ToUpper(in.RecordID),
		"User Name ............ svc_cfsched",
		"Timeout (seconds) .... 900",
		"Publish .............. [x] Save output to a file",
		"File ................. D:\\RSCH\\REPORTS\\" + strings.ToUpper(in.RecordID) + ".TXT",
		"Resolve URL .......... [ ]",
		"",
		"<!--- seal " + strings.ToUpper(in.Token) + " --->",
	}
	for _, s := range siblingSteps(j, l, x.between(1, 2)) {
		ctl = append(ctl, "Chained task ......... RSCH "+s.marker+" nightly")
	}
	log := []string{
		fmt.Sprintf("\"Information\",\"scheduler-1\",\"%02d/%02d/%02d\",\"%02d:%02d:%02d\",,\"Task RSCH %s nightly triggered.\"", in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59), strings.ToUpper(j.id.Program)),
		fmt.Sprintf("\"Information\",\"scheduler-1\",\"%02d/%02d/%02d\",\"%02d:%02d:%02d\",,\"[%s.cfm] Executing at %s\"", in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59), prog, host),
		fmt.Sprintf("\"Information\",\"scheduler-1\",\"%02d/%02d/%02d\",\"%02d:%02d:%02d\",,\"Task RSCH %s nightly completed; output published.\"", in.Month, in.Day, in.Year%100, x.between(1, 5), x.between(0, 59), x.between(0, 59), strings.ToUpper(j.id.Program)),
	}
	return ctl, log
}
