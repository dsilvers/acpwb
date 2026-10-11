package legacyjob_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"acpwb_go/archive"
	"acpwb_go/legacyjob"
)

var testSlugs = []string{
	"human-factors-and-usability-engineering-report-9293", "board-minutes-draft-0007", "vendor-contract-renewal-3345",
	"q3-earnings-review-1234", "workforce-reduction-notice-6654", "legal/settlement-summary-8890", "a/b/c-nested-path-3321", "",
}

type sample struct {
	in legacyjob.Input
	l  *legacyjob.Listing
}

func corpus(t testing.TB, n int) []sample {
	t.Helper()
	var out []sample
	for i := 0; len(out) < n; i++ {
		year := 1985 + i%41
		month := 1 + (i*7)%12
		day := 1 + (i*11)%28
		slug := testSlugs[i%len(testSlugs)]
		if i >= len(testSlugs) {
			slug = fmt.Sprintf("%s-%d", strings.TrimSuffix(testSlugs[i%len(testSlugs)], "/"), i)
		}
		in := archive.SourceCodeInput(year, month, day, slug, i%2 == 0)
		out = append(out, sample{in: in, l: legacyjob.BuildFresh(&in)})
	}
	return out
}

func section(l *legacyjob.Listing, id string) string {
	for _, s := range l.Sections {
		if s.ID == id {
			return s.Text
		}
	}
	return ""
}

func TestIdentityEraPools(t *testing.T) {
	seen := map[string]bool{}
	for y := 1985; y <= 2025; y++ {
		for i := 0; i < 400; i++ {
			id := legacyjob.Identity(y, 1+i%12, 1+i%28, fmt.Sprintf("slug-%d", i))
			if !legacyjob.LanguageLive(id.LangKey, y) {
				t.Fatalf("%d: language %s outside its years", y, id.LangKey)
			}
			seen[id.LangKey] = true
			if id != legacyjob.Identity(y, 1+i%12, 1+i%28, fmt.Sprintf("slug-%d", i)) {
				t.Fatal("Identity not deterministic")
			}
		}
	}
	for _, k := range legacyjob.LanguageKeys() {
		if !seen[k] {
			t.Errorf("language %s never selected", k)
		}
	}
}

// oneOf finds an archive page whose job is written in language k.
func oneOf(t testing.TB, k string) (legacyjob.Input, bool) {
	for y := 1985; y <= 2025; y++ {
		for i := 0; i < 300; i++ {
			slug := fmt.Sprintf("language-probe-%s-%d", k, i)
			if legacyjob.Identity(y, 1+i%12, 1+i%28, slug).LangKey == k {
				return archive.SourceCodeInput(y, 1+i%12, 1+i%28, slug, i%2 == 0), true
			}
		}
	}
	return legacyjob.Input{}, false
}

func TestDeterministic(t *testing.T) {
	in := archive.SourceCodeInput(2022, 4, 6, testSlugs[0], true)
	a, b := legacyjob.RenderHTML(legacyjob.BuildFresh(&in)), legacyjob.RenderHTML(legacyjob.BuildFresh(&in))
	if a != b {
		t.Fatal("same input produced different pages")
	}
}

func TestOutputMatchesPublishedTables(t *testing.T) {
	for _, s := range corpus(t, 120) {
		out := section(s.l, "sysprint")
		for _, m := range s.in.Metrics {
			want := fmt.Sprintf("%9d%9d%+9d", m.Baseline, m.Current, m.Current-m.Baseline)
			if !strings.Contains(out, want) {
				t.Fatalf("%d-%d-%d %s: metric row %q missing", s.in.Year, s.in.Month, s.in.Day, s.in.Slug, want)
			}
		}
		for _, b := range s.in.Bench {
			var want string
			for _, v := range b.P {
				want += fmt.Sprintf("%9d", v)
			}
			if !strings.Contains(out, want) {
				t.Fatalf("%d %s: benchmark row missing", s.in.Year, s.in.Slug)
			}
		}
	}
}

var slotRe = regexp.MustCompile(`\{[A-Z][A-Z0-9.]*\}`)
var isoRe = regexp.MustCompile(`\b(19[6-9]\d|20[0-4]\d)-(\d\d)-(\d\d)\b`)

func TestNoUnfilledSlotsAndNoFutureDates(t *testing.T) {
	for _, s := range corpus(t, 120) {
		for _, sec := range s.l.Sections {
			text := sec.Text + sec.HTML
			if m := slotRe.FindString(text); m != "" && sec.ID != "sysprint" {
				t.Fatalf("%d %s section %s: unfilled slot %s", s.in.Year, s.in.Slug, sec.ID, m)
			}
		}
		article := fmt.Sprintf("%04d-%02d-%02d", s.in.Year, s.in.Month, s.in.Day)
		src := section(s.l, "source")
		for _, m := range isoRe.FindAllString(src, -1) {
			if m > article {
				t.Fatalf("%s %s: source mentions %s, after the article date", article, s.in.Slug, m)
			}
		}
	}
}

func TestFixedFormColumns(t *testing.T) {
	for _, s := range corpus(t, 160) {
		if k := s.l.ID.LangKey; k != "f4" && k != "f77" && k != "gf" && k != "cobol" {
			continue
		}
		in := s.in
		l := legacyjob.BuildFresh(&in)
		for _, sec := range l.Sections {
			if sec.ID != "source" || strings.Contains(sec.Tab, "LISTING") {
				continue
			}
			for i, ln := range strings.Split(sec.Text, "\n") {
				if len(ln) > 80 {
					t.Fatalf("%s line %d is %d columns: %q", l.ID.LangKey, i, len(ln), ln)
				}
			}
		}
	}
}

func TestWatermarkInCode(t *testing.T) {
	for _, s := range corpus(t, 120) {
		src := strings.ToUpper(section(s.l, "source"))
		tok := strings.ToUpper(s.in.Token)
		// The seal must appear as a code literal (quoted string or FORTRAN IV
		// Hollerith), not only in comments.
		found := strings.Contains(src, "'"+tok+"'") || strings.Contains(src, "\""+tok+"\"") || strings.Contains(src, "4H"+tok[:4])
		if !found {
			t.Fatalf("%s %d %s: watermark %s not found in any code line", s.l.ID.LangKey, s.in.Year, s.in.Slug, tok)
		}
	}
}

func lineSet(text string) map[string]bool {
	m := map[string]bool{}
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSpace(l)
		if len(l) > 24 {
			m[l] = true
		}
	}
	return m
}

func TestVariety(t *testing.T) {
	c := corpus(t, 80)
	var total float64
	pairs := 0
	for i := 0; i+1 < len(c); i += 2 {
		a, b := lineSet(section(c[i].l, "source")), lineSet(section(c[i+1].l, "source"))
		inter := 0
		for k := range a {
			if b[k] {
				inter++
			}
		}
		union := len(a) + len(b) - inter
		if union > 0 {
			total += float64(inter) / float64(union)
			pairs++
		}
	}
	if avg := total / float64(pairs); avg > 0.2 {
		t.Fatalf("listings too similar: mean line-set Jaccard %.3f", avg)
	}
	seenProg := map[string]bool{}
	for _, s := range c {
		seenProg[s.l.ID.Program+s.l.ID.Release] = true
	}
	if len(seenProg) < 50 {
		t.Fatalf("only %d distinct program identities across %d pages", len(seenProg), len(c))
	}
}

func TestPunchesRoundTrip(t *testing.T) {
	text := "HEDGECROFT INGRAM ADVISORY  1AA96252 $(+-*/)=.,'"
	got := strings.TrimRight(legacyjob.DecodePunches(legacyjob.Punches(text)), " ")
	if got != text {
		t.Fatalf("round trip: got %q want %q", got, text)
	}
}

func TestCodeEmailFormat(t *testing.T) {
	if got := legacyjob.CodeEmail("Xiao", "Whitehead"); got != "xiao.whitehead@acpwb.com" {
		t.Fatal(got)
	}
	in := archive.SourceCodeInput(2022, 4, 6, testSlugs[0], true)
	l := legacyjob.BuildFresh(&in)
	re := regexp.MustCompile(`^[a-z0-9]+(\.[a-z0-9]+)+@acpwb\.com$|^[a-z0-9.]+@acpwb\.com$`)
	if len(l.Emails) == 0 {
		t.Fatal("no emails registered")
	}
	for _, e := range l.Emails {
		if !re.MatchString(e) {
			t.Fatalf("email %q not in first.last format", e)
		}
	}
	page := legacyjob.RenderHTML(l)
	for _, m := range in.Team {
		if strings.Contains(page, legacyjob.CodeEmail(strings.Fields(m.Name)[0], strings.Fields(m.Name)[1])) && m.Email != legacyjob.CodeEmail(strings.Fields(m.Name)[0], strings.Fields(m.Name)[1]) {
			t.Fatalf("engagement team email mismatch for %s", m.Name)
		}
	}
}

func TestContentsAnchors(t *testing.T) {
	for _, s := range corpus(t, 40) {
		page := legacyjob.RenderHTML(s.l)
		for _, sec := range s.l.Sections {
			for _, a := range sec.Anchors {
				id := fmt.Sprintf(`id="acpwb-src-%s-%s"`, strings.ToLower(s.in.RecordID), a.ID)
				if !strings.Contains(page, id) {
					t.Fatalf("anchor %s listed but not present", a.ID)
				}
			}
		}
		if len(s.l.Sections) < 7 {
			t.Fatalf("expected full section set, got %d", len(s.l.Sections))
		}
		cards := false
		for _, sec := range s.l.Sections {
			cards = cards || sec.ID == "cards"
		}
		if cards && (s.l.Wrapper != "jcl" && s.l.Wrapper != "dcl" || s.in.Year > 1995 && s.l.ID.LangKey != "f4") {
			t.Fatalf("%d %s/%s: punch cards outside the card era", s.in.Year, s.l.Wrapper, s.l.ID.LangKey)
		}
	}
}

var languageMarkers = map[string]string{"f4": "SUBROUTINE", "f77": "SUBROUTINE", "gf": "SUBROUTINE", "cobol": "PROCEDURE DIVISION",
	"basic": "GOSUB", "atari": "INPUT #1,CARD$", "c": "#include", "cbtr": "BTRV(B_INSERT", "perl": "sub ", "perl4": "sub ", "py2": "def ",
	"ada": "end loop;", "pascal": "begin", "pascalbtr": "BTRV(B_INSERT", "pli": "PROCEDURE OPTIONS(MAIN)", "pal": "ENDPROC",
	"dbase": "ENDDO", "clipper": "RETURN NIL", "foxpro": "ENDFOR", "natural": "END-DEFINE", "informix": "END MAIN",
	"vb": "End Sub", "vba": "wsOut.Cells", "vbnet": "End Module", "csharp": "static void Main", "php": "<?php",
	"pick": "GOSUB", "mumps": "Q", "plsql": "UTL_FILE.GET_LINE", "sql": "INSERT INTO", "db2sql": "INSERT INTO",
	"sqlds": "INSERT INTO", "sqlplus": "SPOOL", "tsql": "@@FETCH_STATUS", "sqlpl": "LANGUAGE SQL",
	"rpg": "QSYSPRT", "rexx": "EXECIO", "clist": "GETFILE DECK", "cl400": "RCVF", "ezt": "JOB INPUT NULL", "abap": "READ DATASET",
	"apl": "⍞", "focus": "TABLE FILE DECK", "ramis": "TABLE FILE DECK", "nomad": "SCHEMA", "sas": "_NULL_", "spss": "DATA LIST",
	"stata": "infix", "splus": "function(", "r": "function(", "matlab": "fgetl", "octave": "fgetl", "awk": "getline",
	"ksh": "read -r", "tcl": "gets $deck", "bc": "read()", "dc": "0k", "cfml": "<cfset", "asp": "Response.Write",
	"modula2": "MODULE", "lisp": "(defun", "prolog": "between(1, 8", "forth": "CMOVE", "smalltalk": "methodsFor:"}

func TestEveryLanguageEmits(t *testing.T) {
	for _, k := range legacyjob.LanguageKeys() {
		in, ok := oneOf(t, k)
		if !ok {
			t.Errorf("no page found for %s", k)
			continue
		}
		l := legacyjob.BuildFresh(&in)
		src := section(l, "source")
		marker, ok := languageMarkers[k]
		if !ok {
			t.Errorf("no marker for %s", k)
			continue
		}
		if !strings.Contains(strings.ToUpper(src), strings.ToUpper(marker)) {
			t.Errorf("%s listing lacks %q", k, marker)
		}
		if n := strings.Count(src, "\n"); n < 150 {
			t.Errorf("%s listing only %d lines", k, n)
		}
		if section(l, "sysprint") == "" || section(l, "jobcontrol") == "" {
			t.Errorf("%s: missing output or job control", k)
		}
		if strings.Contains(src, "%!") || strings.Contains(src, "(MISSING)") || strings.Contains(src, "%(") {
			t.Errorf("%s: format verb leaked into source", k)
		}
		tok := strings.ToUpper(in.Token)
		up := strings.ToUpper(src)
		if !strings.Contains(up, "'"+tok+"'") && !strings.Contains(up, "\""+tok+"\"") && !strings.Contains(up, "4H"+tok[:4]) {
			t.Errorf("%s: seal %s not present as a code literal", k, tok)
		}
	}
}

func BenchmarkBuild(b *testing.B) {
	in := archive.SourceCodeInput(2022, 4, 6, testSlugs[0], true)
	for i := 0; i < b.N; i++ {
		legacyjob.RenderHTML(legacyjob.BuildFresh(&in))
	}
}

func TestLicenses(t *testing.T) {
	names := map[string]int{}
	for _, s := range corpus(t, 200) {
		last := s.l.Sections[len(s.l.Sections)-1]
		if last.ID != "license" || strings.TrimSpace(last.Text) == "" {
			t.Fatalf("%d %s: license is not the last section", s.in.Year, s.in.Slug)
		}
		page := legacyjob.RenderHTML(s.l)
		if !strings.Contains(page, `-license">License</a>`) || !strings.Contains(page, "<th>License</th>") {
			t.Fatalf("license missing from contents or summary")
		}
		n := s.l.License
		names[strings.SplitN(n, " + ", 2)[0]]++
		y := s.in.Year
		switch {
		case strings.Contains(n, "v3") && y < 2007, strings.Contains(n, "Apache License 2.0") && y < 2004,
			strings.Contains(n, "Shareware") && y > 1999, strings.Contains(n, "Creative Commons Attribution-NonCommercial 4.0") && y < 2013,
			strings.Contains(n, "Python Software Foundation") && s.l.ID.LangKey != "py2", strings.Contains(n, "Artistic") && s.l.ID.LangKey != "perl" && s.l.ID.LangKey != "perl4":
			t.Fatalf("%d %s: anachronistic or mismatched license %q", y, s.l.ID.LangKey, n)
		}
	}
	if len(names) < 20 {
		t.Fatalf("only %d distinct licenses across 200 pages", len(names))
	}
}

// Dead code may carry a debug checkpoint, but never a run of them.
func TestDebugCheckpointsNotRepeated(t *testing.T) {
	for _, s := range corpus(t, 200) {
		prev := false
		for _, ln := range strings.Split(section(s.l, "source"), "\n") {
			up := strings.ToUpper(ln)
			// FORTRAN gathers a routine's FORMAT statements at its end
			hit := strings.Contains(up, "DEBUG CHECKPOINT") && !strings.Contains(up, "FORMAT")
			if hit && prev {
				t.Fatalf("%d %s (%s): consecutive debug checkpoints near %q", s.in.Year, s.in.Slug, s.l.ID.LangKey, strings.TrimSpace(ln))
			}
			prev = hit
		}
	}
}
