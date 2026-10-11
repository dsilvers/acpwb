package legacyjob

import (
	"fmt"
	"regexp"
	"strings"
)

// sqlEmitter renders the batch job as a set-based SQL script: the deck is
// loaded into a staging table, parsed with SUBSTR, the change column and the
// percentile ladder are computed with UPDATE / INSERT ... SELECT, and the
// report is a formatted SELECT (or a cursor loop, for the procedural
// flavours). Flavours: ANSI SQL, DB2 (DSNTEP2), SQL/DS, Oracle SQL*Plus,
// Transact-SQL and DB2 SQL PL.
type sqlEmitter struct {
	j      *job
	out    []Line
	k      string
	upper  bool
	cat    string // concatenation operator
	term   string // statement terminator
	tables map[string]string
	opt    sqlOpts
	padAlt bool
}

// sqlOpts are the per-page structural choices; every combination produces
// the same rows.
type sqlOpts struct {
	join      string // cross inner comma
	ladder    string // table union rung correlated
	change    string // update merge correlated
	pivot     string // case pivot scalar selfjoin
	parse     string // subquery join
	cte       bool
	rownum    bool
	whileLoop bool // T-SQL report by WHILE over seq instead of cursors
}

func newSQLEmitter(j *job) *sqlEmitter {
	k := j.id.LangKey
	e := &sqlEmitter{j: j, k: k, upper: k == "db2sql" || k == "sqlds", cat: " || ", term: ";"}
	if k == "tsql" {
		e.cat = " + "
	}
	if k == "sqlpl" {
		e.term = "@"
	}
	x := j.x
	pre := pick(x, []string{"bench_", "rsch_", "bm_", "t_", "wk_"})
	e.opt = e.chooseOpts()
	e.padAlt = x.chance(0.5)
	e.tables = map[string]string{
		"deck": pre + pick(x, []string{"deck_cards", "sysin", "card_image", "input_deck"}),
		"met":  pre + pick(x, []string{"metric", "engagement_metric", "met_work", "metrics"}),
		"med":  pre + pick(x, []string{"median", "benchmark_median", "bench_work", "medians"}),
		"lad":  pre + pick(x, []string{"pct_ladder", "ladder", "rung", "pctmul"}),
		"pct":  pre + pick(x, []string{"pctl", "percentile", "ladder_out", "pct_result"}),
		"ctl":  pre + pick(x, []string{"run_control", "control", "ctl", "run_hdr"}),
	}
	return e
}

// at reports whether the page's database/era supports a feature.
func (e *sqlEmitter) at(feature string) bool {
	y := e.j.in.Year
	since := map[string]map[string]int{
		"cte":    {"sql": 2003, "tsql": 2005, "sqlplus": 2002, "db2sql": 2004, "sqlpl": 2004},
		"window": {"sql": 2009, "tsql": 2005, "sqlplus": 1999, "db2sql": 2004, "sqlpl": 2002},
		"merge":  {"tsql": 2008, "sqlplus": 2002, "db2sql": 2004, "sqlpl": 2004},
		"pivot":  {"tsql": 2005, "sqlplus": 2007},
		"ansi":   {"sql": 1990, "tsql": 1996, "sqlplus": 2001, "db2sql": 1993, "sqlpl": 1999},
	}[feature]
	from, ok := since[e.k]
	return ok && y >= from
}

func (e *sqlEmitter) chooseOpts() sqlOpts {
	x := e.j.x
	o := sqlOpts{join: "comma", ladder: "table", change: "update", pivot: "case", parse: "subquery"}
	if e.at("ansi") {
		o.join = pick(x, []string{"cross", "inner", "comma", "cross"})
	}
	o.ladder = pick(x, []string{"table", "table", "union", "rung", "correlated"})
	o.change = pick(x, []string{"update", "update", "correlated"})
	if e.at("merge") && x.chance(0.4) {
		o.change = "merge"
	}
	pivots := []string{"case", "case", "scalar", "selfjoin"}
	if e.at("pivot") {
		pivots = append(pivots, "pivot", "pivot")
	}
	o.pivot = pick(x, pivots)
	o.parse = pick(x, []string{"subquery", "join"})
	o.cte = e.at("cte") && x.chance(0.5)
	o.rownum = e.at("window") && x.chance(0.4)
	o.whileLoop = x.chance(0.4)
	return o
}

// dual is the one-row table a SELECT of constants needs on this database.
func (e *sqlEmitter) dual() string {
	switch e.k {
	case "sqlplus":
		return " FROM DUAL"
	case "db2sql", "sqlds", "sqlpl":
		return e.c(" FROM SYSIBM.SYSDUMMY1")
	}
	return ""
}

func (e *sqlEmitter) c(s string) string {
	if e.upper {
		return strings.NewReplacer("%S", "%s", "%D", "%d").Replace(strings.ToUpper(s))
	}
	return s
}

func (e *sqlEmitter) t(name string) string { return e.c(e.tables[name]) }

func (e *sqlEmitter) code(text string) {
	e.out = append(e.out, Line{Text: text, Code: true})
}

func (e *sqlEmitter) codef(f string, a ...interface{}) { e.code(fmt.Sprintf(f, a...)) }

func (e *sqlEmitter) blank() { e.out = append(e.out, Line{Text: ""}) }

func (e *sqlEmitter) cmt(lines []string) {
	for _, l := range lines {
		e.out = append(e.out, Line{Text: strings.TrimRight("-- "+l, " ")})
	}
}

func (e *sqlEmitter) note(text string) {
	cs := e.j.cs
	cs.width = 72
	e.cmt(cs.para(text))
}

func (e *sqlEmitter) trunc(x string) string {
	switch e.k {
	case "sqlplus":
		return "TRUNC(" + x + ")"
	case "tsql":
		return "CAST(ROUND(" + x + ", 0, 1) AS INT)"
	case "db2sql", "sqlds", "sqlpl":
		return e.c("INTEGER(" + x + ")")
	}
	return "CAST(FLOOR(" + x + ") AS INTEGER)"
}

func (e *sqlEmitter) toInt(x string) string {
	switch e.k {
	case "sqlplus":
		return "TO_NUMBER(" + x + ")"
	case "db2sql", "sqlds", "sqlpl":
		return e.c("INTEGER(" + x + ")")
	}
	return "CAST(" + x + " AS INTEGER)"
}

func (e *sqlEmitter) substr(col string, start, width int) string {
	if e.k == "tsql" {
		return fmt.Sprintf("SUBSTRING(%s, %d, %d)", col, start+1, width)
	}
	return e.c(fmt.Sprintf("SUBSTR(%s, %d, %d)", col, start+1, width))
}

func (e *sqlEmitter) padR(x string, w int) string {
	switch e.k {
	case "sqlplus":
		return fmt.Sprintf("RPAD(%s, %d)", x, w)
	case "tsql":
		return fmt.Sprintf("LEFT(%s + SPACE(%d), %d)", x, w, w)
	case "db2sql", "sqlds", "sqlpl":
		return e.c(fmt.Sprintf("CHAR(%s, %d)", x, w))
	}
	return fmt.Sprintf("CAST(%s AS CHAR(%d))", x, w)
}

func (e *sqlEmitter) padN(x string, w int) string {
	if e.padAlt {
		switch e.k {
		case "sqlplus":
			return fmt.Sprintf("TO_CHAR(%s, '%s')", x, strings.Repeat("9", w-1))
		case "tsql":
			return fmt.Sprintf("RIGHT(SPACE(%d) + CAST(%s AS VARCHAR(12)), %d)", w, x, w)
		case "db2sql", "sqlds", "sqlpl":
			return e.c(fmt.Sprintf("RIGHT('%s' || RTRIM(CHAR(%s)), %d)", strings.Repeat(" ", w), x, w))
		default:
			return fmt.Sprintf("RIGHT('%s' || CAST(%s AS VARCHAR(12)), %d)", strings.Repeat(" ", w), x, w)
		}
	}
	switch e.k {
	case "sqlplus":
		return fmt.Sprintf("LPAD(TO_CHAR(%s), %d)", x, w)
	case "tsql":
		return fmt.Sprintf("STR(%s, %d)", x, w)
	case "db2sql", "sqlds", "sqlpl":
		return e.c(fmt.Sprintf("RIGHT(REPEAT(' ', %d) || RTRIM(CHAR(%s)), %d)", w, x, w))
	}
	return fmt.Sprintf("LPAD(CAST(%s AS VARCHAR(12)), %d)", x, w)
}

func (e *sqlEmitter) padS(x string, w int) string {
	switch e.k {
	case "sqlplus":
		return fmt.Sprintf("LPAD(TO_CHAR(%s, 'S99999990'), %d)", x, w)
	case "tsql":
		return fmt.Sprintf("RIGHT(SPACE(%d) + CASE WHEN %s > 0 THEN '+' ELSE '' END + CAST(%s AS VARCHAR(12)), %d)", w, x, x, w)
	}
	inner := fmt.Sprintf("CASE WHEN %s > 0 THEN '+' ELSE '' END || RTRIM(CHAR(%s))", x, x)
	if e.k == "sql" {
		inner = fmt.Sprintf("CASE WHEN %s > 0 THEN '+' ELSE '' END || CAST(%s AS VARCHAR(12))", x, x)
		return fmt.Sprintf("LPAD(%s, %d)", inner, w)
	}
	return e.c(fmt.Sprintf("RIGHT(REPEAT(' ', %d) || %s, %d)", w, inner, w))
}

// rename swaps fixed identifiers in the lines emitted since start for a
// per-page choice from each candidate list (first entry = the default).
func (e *sqlEmitter) rename(start int, cands map[string][]string) {
	keys := make([]string, 0, len(cands))
	for k := range cands {
		keys = append(keys, k)
	}
	sortStrings(keys)
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k, pick(e.j.x, cands[k]))
	}
	r := strings.NewReplacer(pairs...)
	for i := start; i < len(e.out); i++ {
		e.out[i].Text = r.Replace(e.out[i].Text)
	}
}

func (e *sqlEmitter) intType() string {
	if e.k == "sqlplus" {
		return "NUMBER(9)"
	}
	return e.c("INTEGER")
}

func (e *sqlEmitter) emit() []Line {
	j := e.j
	x := j.x
	callers, _ := callGraph(j.p)
	_ = callers
	e.cmt(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	e.cmt(j.historyLines())
	e.cmt(j.cs.box("Provenance", "Seal "+strings.ToUpper(j.in.Token)+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos")))
	e.session()
	main := j.p.Procs[0]
	e.cmt(j.prologue(main, nil, nil))
	e.ddl()
	e.load()
	e.parse()
	e.compute()
	e.report()
	for _, p := range j.p.Procs[1:] {
		if p.Dead {
			e.deadObject(p)
		}
	}
	e.cleanup()
	if x.chance(0.4) {
		e.blank()
		e.note(j.g.comment(nil, "handoff", "todo", "misc"))
	}
	e.renameColumns()
	return e.out
}

var sqlColumns = []struct {
	name  string
	cands []string
}{
	{"seq", []string{"seq", "seq_no", "card_no", "line_no"}},
	{"card", []string{"card", "card_image", "card_text", "rec_data"}},
	{"metric_name", []string{"metric_name", "metric", "met_desc", "met_label"}},
	{"baseline", []string{"baseline", "base_val", "prior_val", "opening"}},
	{"current_val", []string{"current_val", "curr_val", "latest_val", "closing"}},
	{"change_val", []string{"change_val", "delta", "chg", "diff_val"}},
	{"bench_name", []string{"bench_name", "benchmark", "bm_desc", "bench_label"}},
	{"median_val", []string{"median_val", "med_val", "p50_val", "mid_val"}},
	{"rung", []string{"rung", "tier", "step_no", "lvl"}},
	{"mult", []string{"mult", "factor", "coeff", "multiplier"}},
	{"pct_val", []string{"pct_val", "pctl", "value_out", "rung_val"}},
	{"record_id", []string{"record_id", "rec_id", "doc_id"}},
	{"client", []string{"client", "client_name", "org_name"}},
	{"peer_group", []string{"peer_group", "peer_grp", "comp_set"}},
	{"n_met", []string{"n_met", "met_count", "nmet"}},
	{"n_ben", []string{"n_ben", "ben_count", "nben"}},
	// table aliases
	{"m", []string{"m", "a", "mm", "b1"}},
	{"l", []string{"l", "lad", "x", "r1"}},
	{"p", []string{"p", "pc", "q", "z"}},
	{"c", []string{"c", "k", "ctl", "h"}},
	{"m2", []string{"m2", "a2", "chk"}},
	{"l2", []string{"l2", "x2", "lk"}},
}

// renameColumns gives the work tables' columns and aliases per-page names,
// leaving string literals (report headings, card images) untouched.
func (e *sqlEmitter) renameColumns() {
	x := e.j.x
	m := map[string]string{}
	var alts []string
	for _, c := range sqlColumns {
		n := pick(x, c.cands)
		m[c.name] = n
		alts = append(alts, regexp.QuoteMeta(c.name))
	}
	re := regexp.MustCompile(`(?i)\b(` + strings.Join(alts, "|") + `)\b`)
	for i, l := range e.out {
		if !l.Code {
			continue
		}
		parts := strings.Split(l.Text, "'")
		for k := 0; k < len(parts); k += 2 { // even parts are outside quotes
			parts[k] = re.ReplaceAllStringFunc(parts[k], func(w string) string {
				n, ok := m[strings.ToLower(w)]
				if !ok {
					return w
				}
				if w == strings.ToUpper(w) && w != strings.ToLower(w) {
					return strings.ToUpper(n)
				}
				return n
			})
		}
		e.out[i].Text = strings.Join(parts, "'")
	}
}

func (e *sqlEmitter) session() {
	j := e.j
	x := j.x
	e.blank()
	switch e.k {
	case "sqlplus":
		e.code("SET PAGESIZE 66 LINESIZE 132 FEEDBACK OFF VERIFY OFF HEADING OFF TRIMSPOOL ON")
		e.code("WHENEVER SQLERROR EXIT SQL.SQLCODE ROLLBACK")
		e.codef("SPOOL %s.lis", strings.ToLower(j.id.Program))
	case "tsql":
		for _, o := range sample(x, []string{"SET NOCOUNT ON", "SET ANSI_NULLS ON", "SET QUOTED_IDENTIFIER OFF", "SET XACT_ABORT ON", "SET ANSI_WARNINGS OFF", "SET ARITHABORT ON", "SET CONCAT_NULL_YIELDS_NULL OFF"}, x.between(2, 4)) {
			e.code(o)
		}
		e.codef("USE %s", pick(x, []string{"Benchmarks", "RSCH", "Research", "bench_prod"}))
		e.code("GO")
	case "db2sql", "sqlds":
		e.codef("SET CURRENT SQLID = '%s';", strings.ToUpper(trunc(alnum(j.g.roster[0].Last), 7)+"1"))
	case "sqlpl":
		e.code("--#SET TERMINATOR @")
		e.codef("CONNECT TO %s@", pick(x, []string{"RSCHPRD", "BENCHDB", "DB2PRD1"}))
		e.codef("SET CURRENT SCHEMA = 'RSCH'@")
	default:
		e.note("Portable SQL. Runs on whatever the department is using this year. " + j.g.comment(nil, "vendors", "conversions"))
	}
}

func (e *sqlEmitter) ddl() {
	j := e.j
	it := e.intType()
	ch := func(n int) string { return e.c(fmt.Sprintf("CHAR(%d)", n)) }
	dec := e.c("DECIMAL(5,2)")
	if e.k == "sqlplus" {
		dec = "NUMBER(5,2)"
	}
	tmp := ""
	if e.k == "tsql" {
		tmp = "#"
	}
	for k, v := range e.tables {
		e.tables[k] = tmp + strings.TrimPrefix(v, "#")
	}
	e.blank()
	e.note("Work tables. " + j.g.comment(nil, "data", "warnings", "deck"))
	e.codef(e.c("CREATE TABLE %s (seq %s NOT NULL, card %s)")+e.term, e.t("deck"), it, ch(80))
	e.codef(e.c("CREATE TABLE %s (seq %s NOT NULL, metric_name %s, baseline %s, current_val %s, change_val %s)")+e.term, e.t("met"), it, ch(40), it, it, it)
	e.codef(e.c("CREATE TABLE %s (seq %s NOT NULL, bench_name %s, median_val %s)")+e.term, e.t("med"), it, ch(40), it)
	e.codef(e.c("CREATE TABLE %s (rung %s NOT NULL, label %s, mult %s)")+e.term, e.t("lad"), it, ch(3), dec)
	e.codef(e.c("CREATE TABLE %s (seq %s NOT NULL, rung %s NOT NULL, pct_val %s)")+e.term, e.t("pct"), it, it, it)
	e.codef(e.c("CREATE TABLE %s (record_id %s, client %s, peer_group %s, seal %s, n_met %s, n_ben %s)")+e.term, e.t("ctl"), ch(8), ch(40), ch(60), ch(8), it, it)
	e.blank()
	e.note("The percentile ladder. Ratified 1981-02-17. " + j.g.comment(nil, "magic", "lore", "memos"))
	for k := 0; k < 8; k++ {
		e.codef(e.c("INSERT INTO %s VALUES (%d, '%s', %s)")+e.term, e.t("lad"), k+1, ladderCol[k], ladderLit[k])
	}
	e.codef(e.c("INSERT INTO %s (seal) VALUES ('%s')")+e.term, e.t("ctl"), strings.ToUpper(j.in.Token))
}

func (e *sqlEmitter) load() {
	j := e.j
	e.blank()
	e.note("Card images from the input deck, one row per card. Generated by " + pick(j.x, []string{"DSNTIAUL", "SQL*Loader", "bcp", "dbload", "a Perl one-liner", "hand"}) + ". " + j.g.comment(nil, "deck", "data", "operators"))
	q := func(card string) string { return "'" + strings.ReplaceAll(card, "'", "''") + "'" }
	cols := ""
	if j.x.chance(0.5) {
		cols = e.c(" (seq, card)")
	}
	multi := map[string]int{"sql": 2006, "tsql": 2008, "db2sql": 2004, "sqlpl": 2004}[e.k]
	style := j.x.intn(4)
	switch {
	case style == 1 && multi > 0 && j.in.Year >= multi:
		// one multi-row VALUES statement
		e.codef(e.c("INSERT INTO %s%s VALUES"), e.t("deck"), cols)
		for i, card := range j.rep.Deck {
			sep := ","
			if i == len(j.rep.Deck)-1 {
				sep = e.term
			}
			e.codef("  (%d, %s)%s", i+1, q(card), sep)
		}
	case style == 1 && e.k == "sqlplus" && j.in.Year >= 2002:
		e.code("INSERT ALL")
		for i, card := range j.rep.Deck {
			e.codef("  INTO %s%s VALUES (%d, %s)", e.t("deck"), cols, i+1, q(card))
		}
		e.code("SELECT * FROM DUAL;")
	case style == 2:
		// INSERT ... SELECT ... UNION ALL
		e.codef(e.c("INSERT INTO %s%s"), e.t("deck"), cols)
		for i, card := range j.rep.Deck {
			kw := e.c("UNION ALL SELECT")
			if i == 0 {
				kw = e.c("          SELECT")
			}
			end := ""
			if i == len(j.rep.Deck)-1 {
				end = e.term
			}
			e.codef("%s %d, %s%s%s", kw, i+1, q(card), e.dual(), end)
		}
	default:
		f := pick(j.x, []string{"INSERT INTO %s%s VALUES (%d, %s)", "INSERT INTO %s%s VALUES ( %d, %s )", "INSERT INTO %s%s\n    VALUES (%d, %s)", "INSERT INTO %s%s VALUES (%3d, %s)"})
		for i, card := range j.rep.Deck {
			for _, l := range strings.Split(fmt.Sprintf(strings.ReplaceAll(e.c(f), "%3D", "%3d"), e.t("deck"), cols, i+1, q(card))+e.term, "\n") {
				e.code(l)
			}
		}
	}
}

func (e *sqlEmitter) parse() {
	j := e.j
	d, ctl := e.t("deck"), e.t("ctl")
	nMet := "(" + e.c("SELECT "+e.toInt(e.substr("card", 0, 3))+" FROM "+d+" WHERE seq = 3") + ")"
	e.blank()
	e.note("Header cards: client, peer group, counts. " + j.g.comment(nil, "data", "formats"))
	e.codef(e.c("UPDATE %s SET record_id = (SELECT %s FROM %s WHERE seq = 1),")+"", ctl, e.substr("card", 40, 8), d)
	e.codef(e.c("    client = (SELECT %s FROM %s WHERE seq = 1),"), e.substr("card", 0, 40), d)
	e.codef(e.c("    peer_group = (SELECT %s FROM %s WHERE seq = 2),"), e.substr("card", 0, 60), d)
	e.codef(e.c("    n_met = %s,"), nMet)
	e.codef(e.c("    n_ben = (SELECT %s FROM %s WHERE seq = 3)")+e.term, e.toInt(e.substr("card", 3, 3)), d)
	from, nm := d, nMet
	if e.opt.parse == "join" {
		nm = e.c("c.n_met")
		switch {
		case e.opt.join == "comma" || !e.at("ansi"):
			from = d + e.c(" d, ") + ctl + e.c(" c")
		default:
			from = d + e.c(" d INNER JOIN ") + ctl + e.c(" c ON c.n_met IS NOT NULL")
		}
	}
	e.blank()
	e.note("Engagement metric cards. " + j.g.comment(nil, "data", "deck", "units"))
	e.codef(e.c("INSERT INTO %s (seq, metric_name, baseline, current_val)"), e.t("met"))
	e.codef(e.c("SELECT seq - 3, %s, %s, %s"), e.substr("card", 0, 40), e.toInt(e.substr("card", 40, 6)), e.toInt(e.substr("card", 46, 6)))
	e.codef(e.c("  FROM %s"), from)
	e.codef(e.c(" WHERE seq BETWEEN 4 AND 3 + %s")+e.term, nm)
	e.blank()
	e.note("Benchmark median cards. " + j.g.comment(nil, "data", "magic"))
	e.codef(e.c("INSERT INTO %s (seq, bench_name, median_val)"), e.t("med"))
	e.codef(e.c("SELECT seq - 3 - %s, %s, %s"), nm, e.substr("card", 0, 40), e.toInt(e.substr("card", 40, 7)))
	e.codef(e.c("  FROM %s"), from)
	e.codef(e.c(" WHERE seq > 3 + %s")+e.term, nm)
}

func (e *sqlEmitter) compute() {
	j := e.j
	met, med, lad, pct := e.t("met"), e.t("med"), e.t("lad"), e.t("pct")
	e.blank()
	e.note("Period-over-period change. " + j.g.comment(nil, "warnings", "units", "revision"))
	switch e.opt.change {
	case "merge":
		e.codef(e.c("MERGE INTO %s t"), met)
		e.codef(e.c("USING (SELECT seq, current_val - baseline AS chg FROM %s) s"), met)
		e.code(e.c("   ON (t.seq = s.seq)"))
		e.code(e.c(" WHEN MATCHED THEN UPDATE SET t.change_val = s.chg") + e.term)
	case "correlated":
		e.codef(e.c("UPDATE %s SET change_val ="), met)
		e.codef(e.c("       (SELECT m2.current_val - m2.baseline FROM %s m2 WHERE m2.seq = %s.seq)")+e.term, met, met)
	default:
		e.codef(e.c("UPDATE %s SET change_val = current_val - baseline")+e.term, met)
	}
	e.blank()
	e.note("Apply the ladder. Truncate, never round (audit 83-07). " + j.g.comment(nil, "magic", "lore", "philosophy"))
	if e.opt.ladder == "rung" {
		e.note("One statement per rung, so the auditors can tick them off. " + j.g.comment(nil, "testing", "regulatory"))
		for k := 0; k < 8; k++ {
			e.codef(e.c("INSERT INTO %s (seq, rung, pct_val) SELECT m.seq, %d, %s FROM %s m")+e.term, pct, k+1, e.trunc("m.median_val * "+ladderLit[k]), med)
		}
		return
	}
	src := lad + e.c(" l")
	switch e.opt.ladder {
	case "union":
		var parts []string
		for k := 0; k < 8; k++ {
			parts = append(parts, e.c(fmt.Sprintf("SELECT %d AS rung, %s AS mult", k+1, ladderLit[k]))+e.dual())
		}
		src = "(" + strings.Join(parts, e.c(" UNION ALL ")) + e.c(") l")
	case "correlated":
		src = lad + e.c(" l")
	}
	if e.opt.cte && e.opt.ladder != "union" {
		if e.k == "tsql" {
			e.codef("WITH rungs (rung, mult) AS (SELECT rung, mult FROM %s)", lad)
			e.codef("INSERT INTO %s (seq, rung, pct_val)", pct)
		} else {
			e.codef(e.c("INSERT INTO %s (seq, rung, pct_val)"), pct)
			e.codef(e.c("WITH rungs (rung, mult) AS (SELECT rung, mult FROM %s)"), lad)
		}
		src = e.c("rungs l")
	} else {
		e.codef(e.c("INSERT INTO %s (seq, rung, pct_val)"), pct)
	}
	val := e.trunc("m.median_val * l.mult")
	if e.opt.ladder == "correlated" {
		val = e.trunc(e.c("m.median_val * (SELECT l2.mult FROM ") + lad + e.c(" l2 WHERE l2.rung = l.rung)"))
	}
	e.codef(e.c("SELECT m.seq, l.rung, %s"), val)
	switch e.opt.join {
	case "inner":
		e.codef(e.c("  FROM %s m INNER JOIN %s ON l.rung BETWEEN 1 AND 8")+e.term, med, src)
	case "cross":
		e.codef(e.c("  FROM %s m CROSS JOIN %s")+e.term, med, src)
	default:
		e.codef(e.c("  FROM %s m, %s")+e.term, med, src)
	}
}

func (e *sqlEmitter) report() {
	j := e.j
	r := j.rep
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "formats", "clients"))
	if e.k == "tsql" {
		e.reportCursor()
		return
	}
	if e.k == "sqlplus" {
		e.codef("TTITLE LEFT %s SKIP 1", q(strings.TrimSpace(r.HdrLeft)))
		e.code("BTITLE OFF")
	}
	e.codef(e.c("SELECT %s%s%s FROM %s")+e.term, q(r.Title1), e.cat, e.c("client"), e.t("ctl"))
	e.codef(e.c("SELECT %s FROM %s")+e.term, q(r.ColHdr1), e.t("ctl"))
	e.codef(e.c("SELECT ' '%s%s%s%s%s%s%s%s"), e.cat, e.padR("metric_name", 40), e.cat, e.padN("baseline", 9), e.cat, e.padN("current_val", 9), e.cat, e.padS("change_val", 9))
	e.codef(e.c("  FROM %s ORDER BY seq")+e.term, e.t("met"))
	e.codef(e.c("SELECT %s%s%s FROM %s")+e.term, q(r.Title2), e.cat, e.c("peer_group"), e.t("ctl"))
	e.codef(e.c("SELECT %s FROM %s")+e.term, q(r.ColHdr2), e.t("ctl"))
	e.benchReport()
	if e.k == "sqlpl" {
		e.sqlplProc()
	}
	e.codef(e.c("SELECT '0RECORDS PROCESSED'%s%s%s' METRICS'%s%s%s' BENCHMARKS   SEAL '%sseal FROM %s")+e.term,
		e.cat, e.padN("n_met", 6), e.cat, e.cat, e.padN("n_ben", 6), e.cat, e.cat, e.t("ctl"))
}

// benchReport: the ladder pivoted back to one row per benchmark, written
// one of several equivalent ways.
func (e *sqlEmitter) benchReport() {
	med, pct := e.t("med"), e.t("pct")
	order := e.c(" ORDER BY m.seq")
	if e.opt.rownum {
		order = e.c(" ORDER BY ROW_NUMBER() OVER (ORDER BY m.seq)")
	}
	switch e.opt.pivot {
	case "pivot":
		var in, cols []string
		for k := 1; k <= 8; k++ {
			in = append(in, fmt.Sprintf("[%d]", k))
			cols = append(cols, e.padN(fmt.Sprintf("pv.[%d]", k), 9))
		}
		if e.k == "sqlplus" {
			in, cols = nil, nil
			for k := 1; k <= 8; k++ {
				in = append(in, fmt.Sprintf("%d AS r%d", k, k))
				cols = append(cols, e.padN(fmt.Sprintf("pv.r%d", k), 9))
			}
		}
		e.codef("SELECT ' '%s%s", e.cat, e.padR("m.bench_name", 40))
		for _, c := range cols {
			e.codef("       %s%s", e.cat, c)
		}
		e.codef("  FROM %s m", med)
		e.codef("  JOIN (SELECT seq, rung, pct_val FROM %s) src", pct)
		e.codef("       PIVOT (MAX(pct_val) FOR rung IN (%s)) pv ON pv.seq = m.seq", strings.Join(in, ", "))
		e.code(" ORDER BY m.seq" + e.term)
		return
	case "scalar":
		e.codef(e.c("SELECT ' '%s%s"), e.cat, e.padR("m.bench_name", 40))
		for k := 1; k <= 8; k++ {
			e.codef("       %s%s", e.cat, e.c(e.padN(fmt.Sprintf("(SELECT p.pct_val FROM %s p WHERE p.seq = m.seq AND p.rung = %d)", pct, k), 9)))
		}
		e.codef(e.c("  FROM %s m")+"%s"+e.term, med, order)
		return
	case "selfjoin":
		e.codef(e.c("SELECT ' '%s%s"), e.cat, e.padR("m.bench_name", 40))
		for k := 1; k <= 8; k++ {
			e.codef("       %s%s", e.cat, e.c(e.padN(fmt.Sprintf("p%d.pct_val", k), 9)))
		}
		if e.at("ansi") && e.opt.join != "comma" {
			e.codef(e.c("  FROM %s m"), med)
			for k := 1; k <= 8; k++ {
				e.codef(e.c("  JOIN %s p%d ON p%d.seq = m.seq AND p%d.rung = %d"), pct, k, k, k, k)
			}
			e.code(order + e.term)
			return
		}
		var from, where []string
		for k := 1; k <= 8; k++ {
			from = append(from, fmt.Sprintf("%s p%d", pct, k))
			outer := ""
			if e.k == "sqlplus" {
				outer = "(+)" // Oracle outer join, the way the shop always wrote it
			}
			where = append(where, fmt.Sprintf("p%d.seq%s = m.seq AND p%d.rung%s = %d", k, outer, k, outer, k))
		}
		e.codef(e.c("  FROM %s m, %s"), med, e.c(strings.Join(from, ", ")))
		for i, w := range where {
			kw := "   AND "
			if i == 0 {
				kw = " WHERE "
			}
			e.code(e.c(kw + w))
		}
		e.code(order + e.term)
		return
	}
	var cols []string
	for k := 1; k <= 8; k++ {
		cols = append(cols, e.padN(fmt.Sprintf("MAX(CASE WHEN p.rung = %d THEN p.pct_val END)", k), 9))
	}
	e.codef(e.c("SELECT ' '%s%s"), e.cat, e.padR("MAX(m.bench_name)", 40))
	for _, c := range cols {
		e.codef("       %s%s", e.cat, e.c(c))
	}
	switch {
	case e.k == "sqlplus" && !e.at("ansi"):
		e.codef("  FROM %s m, %s p WHERE p.seq(+) = m.seq", med, pct)
	case e.opt.join == "comma" || !e.at("ansi"):
		e.codef(e.c("  FROM %s m, %s p WHERE p.seq = m.seq"), med, pct)
	case e.opt.join == "inner":
		e.codef(e.c("  FROM %s m INNER JOIN %s p ON p.seq = m.seq"), med, pct)
	default:
		e.codef(e.c("  FROM %s m LEFT OUTER JOIN %s p ON p.seq = m.seq"), med, pct)
	}
	e.code(e.c(" GROUP BY m.seq ORDER BY m.seq") + e.term)
}

// reportCursor: Transact-SQL prints row by row through cursors.
func (e *sqlEmitter) reportCursor() {
	j := e.j
	r := j.rep
	start := len(e.out)
	defer e.rename(start, map[string][]string{
		"@client": {"@client", "@clnt", "@org", "@firm"}, "@peer": {"@peer", "@peergrp", "@grp"},
		"@name": {"@name", "@nm", "@label", "@desc"}, "@base": {"@base", "@bsln", "@prior"},
		"@curr": {"@curr", "@cur", "@now"}, "@chg": {"@chg", "@delta", "@diff"}, "@seq": {"@seq", "@row", "@i"},
		"@line": {"@line", "@out", "@buf", "@prt"}, "c_met": {"c_met", "curMetrics", "crs_met", "mcur"},
		"c_ben": {"c_ben", "curBench", "crs_ben", "bcur"}, "@n ": {"@n ", "@last ", "@maxseq "},
	})
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	e.code("DECLARE @client CHAR(40), @peer CHAR(60), @name CHAR(40), @base INT, @curr INT, @chg INT, @seq INT, @line VARCHAR(140)")
	e.codef("SELECT @client = client, @peer = peer_group FROM %s", e.t("ctl"))
	e.codef("PRINT %s + @client", q(r.Title1))
	e.codef("PRINT %s", q(r.ColHdr1))
	e.codef("DECLARE c_met CURSOR LOCAL FAST_FORWARD FOR SELECT metric_name, baseline, current_val, change_val FROM %s ORDER BY seq", e.t("met"))
	e.code("OPEN c_met")
	e.code("FETCH NEXT FROM c_met INTO @name, @base, @curr, @chg")
	e.code("WHILE @@FETCH_STATUS = 0")
	e.code("BEGIN")
	e.codef("    SET @line = ' ' + %s + %s + %s + %s", e.padR("@name", 40), e.padN("@base", 9), e.padN("@curr", 9), e.padS("@chg", 9))
	e.code("    PRINT @line")
	e.code("    FETCH NEXT FROM c_met INTO @name, @base, @curr, @chg")
	e.code("END")
	e.code("CLOSE c_met")
	e.code("DEALLOCATE c_met")
	e.codef("PRINT %s + @peer", q(r.Title2))
	e.codef("PRINT %s", q(r.ColHdr2))
	if e.opt.whileLoop {
		e.code("DECLARE @n INT")
		e.codef("SELECT @seq = MIN(seq), @n = MAX(seq) FROM %s", e.t("med"))
		e.code("WHILE @seq <= @n")
		e.code("BEGIN")
		e.codef("    SELECT @name = bench_name FROM %s WHERE seq = @seq", e.t("med"))
		e.codef("    SET @line = ' ' + %s", e.padR("@name", 40))
		e.codef("    SELECT @line = @line + STR(pct_val, 9) FROM %s WHERE seq = @seq ORDER BY rung", e.t("pct"))
		e.code("    PRINT @line")
		e.code("    SET @seq = @seq + 1")
		e.code("END")
		e.codef("SELECT @line = '0RECORDS PROCESSED' + STR(n_met, 6) + ' METRICS' + STR(n_ben, 6) + ' BENCHMARKS   SEAL ' + seal FROM %s", e.t("ctl"))
		e.code("PRINT @line")
		e.code("GO")
		return
	}
	e.codef("DECLARE c_ben CURSOR LOCAL FAST_FORWARD FOR SELECT seq, bench_name FROM %s ORDER BY seq", e.t("med"))
	e.code("OPEN c_ben")
	e.code("FETCH NEXT FROM c_ben INTO @seq, @name")
	e.code("WHILE @@FETCH_STATUS = 0")
	e.code("BEGIN")
	e.codef("    SET @line = ' ' + %s", e.padR("@name", 40))
	e.codef("    SELECT @line = @line + STR(pct_val, 9) FROM %s WHERE seq = @seq ORDER BY rung", e.t("pct"))
	e.code("    PRINT @line")
	e.code("    FETCH NEXT FROM c_ben INTO @seq, @name")
	e.code("END")
	e.code("CLOSE c_ben")
	e.code("DEALLOCATE c_ben")
	e.codef("SELECT @line = '0RECORDS PROCESSED' + STR(n_met, 6) + ' METRICS' + STR(n_ben, 6) + ' BENCHMARKS   SEAL ' + seal FROM %s", e.t("ctl"))
	e.code("PRINT @line")
	e.code("GO")
}

// sqlplProc: DB2 SQL PL keeps a stored procedure that re-derives the change
// column row by row "for the auditors".
func (e *sqlEmitter) sqlplProc() {
	j := e.j
	start := len(e.out)
	defer e.rename(start, map[string][]string{
		"p_mismatches": {"p_mismatches", "o_bad", "p_diffs", "out_errs"}, "v_chg": {"v_chg", "v_delta", "l_chg", "w_diff"},
		" r.": {" r.", " rec.", " m."}, " r AS ": {" r AS ", " rec AS ", " m AS "}, "c1 CURSOR": {"c1 CURSOR", "cmet CURSOR", "c_chk CURSOR"},
	})
	name := "P_" + j.id.Program + "_XCHK"
	e.blank()
	e.cmt(j.prologue(&Proc{Name: name, Purpose: "Recomputes the change column row by row and compares it with the set-based result. " + j.g.comment(nil, "testing", "regulatory"), Author: j.g.coder()}, nil, nil))
	e.codef("CREATE OR REPLACE PROCEDURE %s (OUT p_mismatches INTEGER)", name)
	e.code("LANGUAGE SQL")
	e.code("BEGIN")
	e.code("  DECLARE v_chg INTEGER DEFAULT 0;")
	e.code("  SET p_mismatches = 0;")
	e.codef("  FOR r AS c1 CURSOR FOR SELECT seq, baseline, current_val, change_val FROM %s DO", e.t("met"))
	e.code("    SET v_chg = r.current_val - r.baseline;")
	e.code("    IF v_chg <> r.change_val THEN")
	e.code("      SET p_mismatches = p_mismatches + 1;")
	e.code("    END IF;")
	e.code("  END FOR;")
	e.code("END@")
	e.codef("CALL %s(?)@", name)
}

// deadObject: one never-used view (or, for Transact-SQL, procedure — views
// cannot see #temp tables) per dead routine.
func (e *sqlEmitter) deadObject(p *Proc) {
	j := e.j
	x := j.x
	e.blank()
	e.cmt(j.prologue(p, nil, nil))
	if x.chance(0.45) {
		e.deadDDL(p)
		return
	}
	name := e.c(strings.ToLower(p.Name))
	mod := func(a string, b int) string {
		if e.k == "tsql" {
			return fmt.Sprintf("%s %% %d", a, b)
		}
		return e.c(fmt.Sprintf("MOD(%s, %d)", a, b))
	}
	var body []string
	switch x.intn(9) {
	case 0:
		body = []string{fmt.Sprintf("SELECT m.seq, m.median_val, %s AS p99", e.trunc("m.median_val * 2.33")), "  FROM " + e.t("med") + " m"}
	case 1:
		body = []string{fmt.Sprintf("SELECT seq, CASE WHEN %s BETWEEN 11 AND 13 THEN 'TH' WHEN %s = 1 THEN 'ST' WHEN %s = 2 THEN 'ND' WHEN %s = 3 THEN 'RD' ELSE 'TH' END AS suffix",
			mod("seq", 100), mod("seq", 10), mod("seq", 10), mod("seq", 10)), "  FROM " + e.t("lad")}
	case 2:
		body = []string{fmt.Sprintf("SELECT seq, baseline, current_val, (current_val - baseline) * %d AS shadow_chg", x.between(1, 3)),
			fmt.Sprintf("  FROM %s WHERE baseline > %d", e.t("met"), x.between(0, 900))}
	case 3:
		body = []string{"SELECT COUNT(*) AS n_cards, SUM(seq) AS seq_sum, MAX(seq) AS last_seq", "  FROM " + e.t("deck")}
	case 4:
		y := x.between(1900, 2099)
		body = []string{fmt.Sprintf("SELECT %d AS yr, CASE WHEN %s = 0 AND (%s <> 0 OR %s = 0) THEN 1 ELSE 0 END AS is_leap", y, mod(fmt.Sprint(y), 4), mod(fmt.Sprint(y), 100), mod(fmt.Sprint(y), 400)),
			"  FROM " + e.t("ctl")}
	case 5:
		body = []string{fmt.Sprintf("SELECT seq, %s AS randu", mod(fmt.Sprintf("seq * 65539 + %d", x.between(1, 99999)), 2147483647)), "  FROM " + e.t("deck"),
			fmt.Sprintf(" WHERE seq <= %d", x.between(3, 40))}
	case 6:
		body = []string{fmt.Sprintf("SELECT m.seq, ROUND(m.median_val * l.mult, %d) AS rounded_val", x.between(-3, 0)),
			"  FROM " + e.t("med") + " m JOIN " + e.t("lad") + " l ON l.rung = " + fmt.Sprint(x.between(1, 8))}
	case 7:
		body = []string{fmt.Sprintf("SELECT %s AS checksum", mod("SUM(baseline * 17 + current_val)", 65521)), "  FROM " + e.t("met")}
	default:
		body = []string{"SELECT c.record_id, c.peer_group, COUNT(*) AS n_rows",
			"  FROM " + e.t("ctl") + " c, " + e.t("met") + " m", " GROUP BY c.record_id, c.peer_group"}
	}
	for i := range body {
		body[i] = e.c(body[i])
	}
	if e.k == "tsql" {
		e.codef("CREATE PROCEDURE dbo.p_%s AS", name)
		e.code("BEGIN")
		for _, l := range body {
			e.code("    " + l)
		}
		e.code("END")
		e.code("GO")
	} else {
		e.codef(e.c("CREATE VIEW v_%s AS"), name)
		for i, l := range body {
			if i == len(body)-1 {
				l += e.term
			}
			e.code(l)
		}
	}
	if x.chance(0.4) {
		e.note(j.g.comment(p.Author, "warnings", "history", "lore"))
	}
}

// deadDDL: an index, grant, comment, sequence or trigger that nothing uses.
func (e *sqlEmitter) deadDDL(p *Proc) {
	x := e.j.x
	name := e.c(strings.ToLower(p.Name))
	tbl := e.t(pick(x, []string{"met", "med", "pct", "deck"}))
	col := e.c(pick(x, []string{"seq", "seq", "rung"}))
	if strings.HasSuffix(tbl, e.t("met")) || strings.HasSuffix(tbl, e.t("med")) || strings.HasSuffix(tbl, e.t("deck")) {
		col = e.c("seq")
	}
	user := e.c(pick(x, []string{"rsch_read", "audit_ro", "bench_rpt", "public"}))
	switch x.intn(5) {
	case 0:
		e.codef(e.c("CREATE INDEX ix_%s ON %s (%s)")+e.term, name, tbl, col)
	case 1:
		e.codef(e.c("GRANT SELECT ON %s TO %s")+e.term, tbl, user)
	case 2:
		if e.k == "tsql" {
			e.codef("EXEC sp_addextendedproperty 'MS_Description', '%s', 'SCHEMA', 'dbo', 'TABLE', '%s'", strings.ReplaceAll(asciiText(trunc(p.Purpose, 60)), "'", "''"), tbl)
			return
		}
		e.codef(e.c("COMMENT ON TABLE %s IS '%s'")+e.term, tbl, strings.ReplaceAll(asciiText(trunc(p.Purpose, 60)), "'", "''"))
	case 3:
		switch e.k {
		case "tsql":
			e.codef("CREATE TABLE seq_%s (id INT IDENTITY(1,1) NOT NULL, stamp DATETIME DEFAULT GETDATE())", name)
		case "sqlds":
			e.codef(e.c("CREATE TABLE seq_%s (id INTEGER NOT NULL)")+e.term, name)
		default:
			e.codef(e.c("CREATE SEQUENCE seq_%s START WITH %d INCREMENT BY 1")+e.term, name, x.between(1, 90000))
		}
	default:
		switch e.k {
		case "tsql":
			e.codef("CREATE TRIGGER trg_%s ON dbo.audit_log AFTER INSERT AS", name)
			e.code("BEGIN")
			e.code("    SET NOCOUNT ON")
			e.codef("    UPDATE dbo.audit_log SET checked = %d WHERE id IN (SELECT id FROM inserted)", x.between(0, 1))
			e.code("END")
			e.code("GO")
		case "sqlplus":
			e.codef("CREATE OR REPLACE TRIGGER trg_%s", name)
			e.code("  BEFORE INSERT ON rsch_audit_log FOR EACH ROW")
			e.code("BEGIN")
			e.codef("  :NEW.checked := %d;", x.between(0, 1))
			e.code("END;")
			e.code("/")
		default:
			e.codef(e.c("CREATE INDEX ix_%s_2 ON %s (%s DESC)")+e.term, name, tbl, col)
		}
	}
	if x.chance(0.4) {
		e.note(e.j.g.comment(p.Author, "warnings", "history", "lore"))
	}
}

func (e *sqlEmitter) cleanup() {
	e.blank()
	e.note("Clean up. " + e.j.g.comment(nil, "operators", "warnings"))
	for _, k := range []string{"pct", "lad", "med", "met", "deck"} {
		e.codef(e.c("DROP TABLE %s")+e.term, e.t(k))
	}
	switch e.k {
	case "tsql":
		e.code("GO")
	case "sqlplus":
		e.code("COMMIT;")
		e.code("SPOOL OFF")
		e.code("EXIT")
	case "sqlpl":
		e.code("COMMIT@")
		e.code("CONNECT RESET@")
	default:
		e.code(e.c("COMMIT") + e.term)
	}
}
