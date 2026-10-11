package legacyjob

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:embed fragments
var fragmentFS embed.FS

type library struct {
	comments map[string][]string // category -> templates
	cats     []string            // sorted category names
	grammar  map[string][]string // part -> options
	names    map[string][]string // pool -> entries
	vars     map[string][]eraVal // SLOT -> era-gated values (fragments/vars)
}

// eraVal is one value of a fragments/vars slot. A line may start with
// "YYYY-YYYY|", "YYYY-|" or "-YYYY|" to restrict it to articles dated in
// that range (so nobody mentions a laptop in 1987).
type eraVal struct {
	From, To int
	Text     string
}

// parseEra splits an optional "YYYY-YYYY|" / "YYYY-|" / "-YYYY|" prefix off
// a fragment line.
func parseEra(l string) eraVal {
	v := eraVal{From: 0, To: 9999, Text: l}
	bar := strings.IndexByte(l, '|')
	if bar <= 0 || bar > 10 {
		return v
	}
	a, b, ok := strings.Cut(l[:bar], "-")
	if !ok {
		return v
	}
	from, err1 := strconv.Atoi(a)
	to, err2 := strconv.Atoi(b)
	if (a != "" && err1 != nil) || (b != "" && err2 != nil) {
		return v
	}
	if a != "" {
		v.From = from
	}
	if b != "" {
		v.To = to
	}
	v.Text = strings.TrimSpace(l[bar+1:])
	return v
}

// eraText strips a fragment line's era prefix and reports whether the line
// may be used for an article dated in year.
func eraText(l string, year int) (string, bool) {
	v := parseEra(l)
	return v.Text, year >= v.From && year <= v.To
}

func loadVars(dir string) map[string][]eraVal {
	out := map[string][]eraVal{}
	for name, lines := range loadDir(dir) {
		key := strings.ToUpper(name)
		for _, l := range lines {
			v := parseEra(l)
			out[key] = append(out[key], v)
		}
	}
	return out
}

var (
	libOnce sync.Once
	lib     *library
)

func readLines(p string) []string {
	raw, err := fragmentFS.ReadFile(p)
	if err != nil {
		return nil
	}
	var out []string
	for _, l := range strings.Split(string(raw), "\n") {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

func loadDir(dir string) map[string][]string {
	m := map[string][]string{}
	entries, _ := fragmentFS.ReadDir(dir)
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".txt") {
			continue
		}
		m[strings.TrimSuffix(ent.Name(), ".txt")] = readLines(path.Join(dir, ent.Name()))
	}
	return m
}

func getLibrary() *library {
	libOnce.Do(func() {
		lib = &library{
			comments: loadDir("fragments/comments"),
			grammar:  loadDir("fragments/grammar"),
			names:    loadDir("fragments/names"),
			vars:     loadVars("fragments/vars"),
		}
		for c := range lib.comments {
			lib.cats = append(lib.cats, c)
		}
		sort.Strings(lib.cats)
	})
	return lib
}

// ── Roster ──────────────────────────────────────────────────────────────────

// Coder is one fictional programmer in a page's shop history.
type Coder struct {
	First, Last    string
	Initials       string
	Role           string
	Start, End     int // tenure years (End <= article year)
	Trait          string
	DateStyle      int
	UserID, Email  string
	Canon, Team    bool
	SignatureStyle int
}

func (c *Coder) Name() string { return c.First + " " + c.Last }

// Display renders the coder's name in one of the forms programmers actually
// signed listings with.
func (c *Coder) Display(x *rnd) string {
	if c.Canon && c.First == "" {
		return c.Last
	}
	switch x.intn(6) {
	case 0:
		return c.First[:1] + ". " + c.Last
	case 1:
		return c.Last + "," + c.First[:1]
	case 2:
		return c.First + " " + c.Last
	case 3:
		return c.Initials + "/" + "ACPWB"
	case 4:
		return c.First + " " + c.Last[:1] + "."
	default:
		return c.First[:1] + ". " + c.Last
	}
}

type canonChar struct {
	first, last, role string
	start, end        int
}

// Canon characters recur across many pages so the corpus reads as one firm's
// institutional history.
var canonCast = []canonChar{
	{"Robert", "Kowalczyk", "Original author, programmer/analyst", 1974, 1991},
	{"Theresa", "Abernathy", "Senior programmer/analyst", 1981, 1999},
	{"Dennis", "Oyelaran", "Systems programmer", 1985, 2004},
	{"Marjorie", "Lindqvist", "DP manager", 1972, 1988},
	{"Harold", "Pettibone", "Vice President, Research Division", 1970, 1994},
	{"Gail", "Hargrove", "Night operator, later programmer", 1983, 2011},
	{"Wojciech", "Nowak", "Contract programmer", 1996, 2001},
	{"Priya", "Venkataraman", "Lead developer", 2003, 2022},
	{"Lyle", "Fennimore", "Database administrator", 1988, 2009},
	{"", "Y2K Task Force", "Remediation team", 1997, 2000},
	{"Esperanza", "Delgado", "Application support", 2008, 2024},
	{"Kurt", "Brandt", "Statistician", 1976, 1997},
	{"Chidi", "Okonkwo", "Platform engineer", 2012, 2025},
	{"Seymour", "Feldman", "Summer intern", 1986, 1986},
	{"Bronwyn", "Ashworth", "Compensation methodologist", 1990, 2015},
}

var roles = []string{
	"Programmer/analyst", "Senior programmer/analyst", "Systems programmer", "Application programmer",
	"DP manager", "Summer intern", "Contract programmer", "Consultant", "Y2K contractor",
	"Data administrator", "Operations analyst", "Statistician", "Methodologist", "Technical lead",
	"Production support", "Night operator", "Keypunch supervisor", "Report analyst",
	"Junior programmer", "Lead developer", "Maintenance programmer", "Offshore team lead",
}

var traits = []string{"terse", "apologetic", "furious", "pedantic", "philosophical", "boastful", "anxious", "cheerful"}

// traitCats biases which phrase categories a coder's comments come from.
var traitCats = map[string][]string{
	"terse":         {"warnings", "todo", "revision", "routine_purpose", "jcl", "deck"},
	"apologetic":    {"apologies", "handoff", "todo", "debugging"},
	"furious":       {"blame", "warnings", "politics", "clients", "vendors"},
	"pedantic":      {"magic", "units", "memos", "regulatory", "formats", "printing"},
	"philosophical": {"philosophy", "history", "misc", "lore"},
	"boastful":      {"boasts", "history", "performance"},
	"anxious":       {"warnings", "superstition", "security", "testing", "operators"},
	"cheerful":      {"misc", "data", "questions", "office", "lore"},
}

func buildRoster(x *rnd, in *Input) []*Coder {
	L := getLibrary()
	n := x.between(4, 12)
	var out []*Coder
	seen := map[string]bool{}
	add := func(c *Coder) {
		key := c.First + c.Last
		if seen[key] {
			return
		}
		seen[key] = true
		if c.First != "" {
			c.Initials = strings.ToUpper(c.First[:1]) + pick(x, []string{"", "J", "A", "M", "E", "R", "L", "W"}) + strings.ToUpper(c.Last[:1])
		} else {
			c.Initials = "Y2KTF"
		}
		c.Trait = pick(x, traits)
		c.DateStyle = x.intn(5)
		c.SignatureStyle = x.intn(10)
		c.UserID = strings.ToUpper(c.Initials) + fmt.Sprintf("%04X", x.intn(65536))
		out = append(out, c)
	}
	// Canon on roughly 30% of pages: 1-2 characters whose tenure fits.
	if x.chance(0.3) {
		for _, cc := range sample(x, canonCast, x.between(1, 3)) {
			if cc.start <= in.Year {
				end := cc.end
				if end > in.Year {
					end = in.Year
				}
				add(&Coder{First: cc.first, Last: cc.last, Role: cc.role, Start: cc.start, End: end, Canon: true})
			}
		}
	}
	// Engagement team members join as requestors / approvers.
	for _, p := range in.Team {
		if x.chance(0.6) {
			f := strings.Fields(p.Name)
			if len(f) >= 2 {
				add(&Coder{First: f[0], Last: f[len(f)-1], Role: p.Title, Start: in.Year - x.between(0, 6), End: in.Year, Team: true})
			}
		}
	}
	// Document revision authors become later maintainers.
	for _, r := range in.Revisions {
		f := strings.Fields(r.Author)
		if len(f) >= 2 && x.chance(0.5) {
			add(&Coder{First: f[0], Last: f[len(f)-1], Role: pick(x, roles), Start: in.Year - x.between(1, 10), End: in.Year})
		}
	}
	for len(out) < n {
		start := x.between(1966, in.Year)
		span := x.between(0, 25)
		end := start + span
		if end > in.Year {
			end = in.Year
		}
		firsts, lasts := in.FirstNames, in.LastNames
		if len(firsts) == 0 || len(lasts) == 0 {
			firsts, lasts = L.names["first"], L.names["last"]
		}
		add(&Coder{First: pick(x, firsts), Last: pick(x, lasts), Role: pick(x, roles), Start: start, End: end})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	return out
}

// ── Era lore ────────────────────────────────────────────────────────────────

type dated struct {
	name string
	year int
}

func eraPick(x *rnd, items []dated, maxYear int) string {
	var ok []string
	for _, it := range items {
		if it.year <= maxYear {
			ok = append(ok, it.name)
		}
	}
	if len(ok) == 0 {
		return items[0].name
	}
	return pick(x, ok)
}

var hardware = []dated{
	{"IBM 360/50", 1965}, {"IBM 360/65", 1966}, {"IBM 370/158", 1972}, {"IBM 370/168", 1973},
	{"IBM 3033", 1977}, {"IBM 4341", 1979}, {"IBM 3081", 1981}, {"IBM 4381", 1983}, {"IBM 3090", 1985},
	{"IBM ES/9000", 1990}, {"IBM 9672 G5", 1998}, {"IBM z900", 2000}, {"IBM z196", 2010}, {"IBM z15", 2019},
	{"DEC PDP-11/70", 1975}, {"VAX-11/780", 1977}, {"VAX 8600", 1984}, {"MicroVAX II", 1985},
	{"VAXstation 3100", 1989}, {"AlphaServer 2100", 1994}, {"Honeywell 6000", 1970}, {"Burroughs B6700", 1971},
	{"Univac 1108", 1965}, {"Data General Eclipse", 1974}, {"Prime 750", 1979}, {"HP 3000", 1972},
	{"Wang VS", 1977}, {"IBM PC/XT", 1983}, {"IBM PC/AT", 1984}, {"Compaq Deskpro 386", 1986},
	{"Sun SPARCstation 20", 1994}, {"Sun Ultra 10", 1998}, {"Sun Fire V880", 2002}, {"HP 9000/735", 1992},
	{"Dell PowerEdge 2650", 2002}, {"Dell PowerEdge R710", 2009}, {"a VMware cluster", 2008},
	{"an m5.large instance", 2017}, {"a Kubernetes node", 2016},
}

var operatingSystems = []dated{
	{"OS/360", 1966}, {"OS/VS2 MVS", 1974}, {"MVS/SP", 1980}, {"MVS/XA", 1983}, {"MVS/ESA", 1988},
	{"OS/390", 1996}, {"z/OS", 2001}, {"VM/CMS", 1972}, {"VAX/VMS", 1978}, {"OpenVMS", 1992},
	{"RSX-11M", 1974}, {"MPE", 1974}, {"PC-DOS 2.0", 1983}, {"MS-DOS 3.3", 1987}, {"Novell NetWare 3", 1989},
	{"SunOS 4", 1988}, {"Solaris 2.6", 1997}, {"Solaris 8", 2000}, {"HP-UX 10", 1995}, {"AIX 4", 1994},
	{"Windows NT 4.0", 1996}, {"Windows 2000 Server", 2000}, {"Red Hat Enterprise Linux 5", 2007},
	{"Ubuntu 18.04", 2018}, {"Debian 11", 2021},
}

var compilers = map[string][]dated{
	"fortran": {{"IBM FORTRAN IV G", 1966}, {"IBM FORTRAN IV H Extended", 1971}, {"VS FORTRAN 1.4", 1983},
		{"VS FORTRAN 2.5", 1990}, {"VAX FORTRAN 4.0", 1985}, {"DEC FORTRAN", 1993}, {"Sun f77 4.2", 1996},
		{"g77 0.5.24", 1999}, {"gfortran 4.4", 2009}, {"gfortran 9", 2019}, {"Intel ifort 11", 2008}},
	"cobol": {{"IBM OS/VS COBOL", 1974}, {"VS COBOL II R3", 1988}, {"COBOL/370", 1991},
		{"IBM COBOL for OS/390", 1997}, {"Enterprise COBOL 3.4", 2005}, {"Enterprise COBOL 6.3", 2019},
		{"Micro Focus COBOL", 1985}, {"GnuCOBOL 3.1", 2020}},
	"basic": {{"Dartmouth BASIC", 1964}, {"BASIC-PLUS", 1972}, {"HP BASIC/3000", 1976}, {"Microsoft BASIC-80", 1979},
		{"GW-BASIC 3.2", 1987}, {"QuickBASIC 4.5", 1988}, {"VAX BASIC 3.4", 1989}},
	"c": {{"Whitesmiths C", 1980}, {"Lattice C", 1983}, {"Microsoft C 5.1", 1988}, {"Turbo C 2.0", 1988},
		{"SunPro C", 1994}, {"gcc 2.95", 1999}, {"gcc 3.4", 2004}, {"gcc 4.8", 2013}, {"clang 10", 2020}},
	"py2": {{"Python 2.2", 2001}, {"Python 2.3.5", 2003}, {"Python 2.4.3", 2005}, {"Python 2.5.2", 2007},
		{"Python 2.6.6", 2008}, {"Python 2.7.3", 2010}, {"Python 2.7.18", 2020}},
	"atari":    {{"Atari BASIC Rev. B", 1983}, {"Atari BASIC Rev. C", 1985}},
	"pick":     {{"Pick R83 BASIC", 1983}, {"ADDS Mentor BASIC", 1986}, {"UniVerse 7.3 BASIC", 1994}, {"UniVerse 10.1", 2003}},
	"mumps":    {{"DSM-11", 1980}, {"VAX DSM V6", 1988}, {"MSM 4.0", 1992}, {"GT.M V4.3", 2001}, {"Cache 5.2", 2006}, {"GT.M V6.3", 2016}},
	"ada":      {{"DEC Ada V1.5", 1985}, {"Alsys Ada", 1988}, {"Verdix Ada", 1990}, {"GNAT 3.15p", 1998}, {"GNAT 4.3", 2008}, {"GNAT Community 2019", 2019}},
	"pascal":   {{"Turbo Pascal 3.0", 1985}, {"Turbo Pascal 5.5", 1989}, {"Turbo Pascal 7.0", 1992}, {"VAX Pascal V3.4", 1986}},
	"pli":      {{"PL/I Optimizing Compiler V1R5", 1980}, {"PL/I for MVS & VM 1.1", 1994}, {"VisualAge PL/I 2.2", 1999}, {"Enterprise PL/I V3R9", 2006}, {"Enterprise PL/I 5.3", 2019}},
	"pal":      {{"Paradox 2.0 PAL", 1987}, {"Paradox 3.5 PAL", 1990}, {"Paradox 4.5 PAL", 1993}, {"Paradox for Windows 7", 1996}},
	"xbase":    {{"dBASE IV 1.1", 1989}, {"Clipper Summer '87", 1988}, {"Clipper 5.2e", 1993}, {"FoxPro 2.6", 1994}, {"Visual FoxPro 6.0", 1998}},
	"natural":  {{"Natural 1.2", 1985}, {"Natural 2.2", 1990}, {"Natural 3.1", 1997}, {"Natural 4.2", 2006}, {"Natural 8.2", 2016}},
	"informix": {{"Informix-4GL 1.10", 1987}, {"Informix-4GL 4.10", 1991}, {"Informix-4GL 7.30", 1999}, {"Informix 4GL 7.32", 2004}},
	"vb":       {{"Visual Basic 3.0", 1993}, {"Visual Basic 4.0", 1995}, {"Visual Basic 6.0 SP6", 1998}},
	"vba":      {{"Excel 5.0 VBA", 1994}, {"Excel 97 VBA", 1997}, {"Excel 2003 VBA", 2003}, {"Excel 2010 VBA 7", 2010}, {"Excel 2016 VBA", 2016}},
	"vbnet":    {{"Visual Basic .NET 2002", 2002}, {"VB 2005", 2005}, {"VB 2010", 2010}, {"VB 16 (.NET 6)", 2021}},
	"csharp":   {{"C# 1.0 (csc 7.0)", 2002}, {"C# 2.0", 2005}, {"C# 4.0", 2010}, {"C# 7.3", 2018}, {"C# 10", 2021}},
	"php":      {{"PHP 4.0.6", 2000}, {"PHP 4.4.9", 2004}, {"PHP 5.2.17", 2006}, {"PHP 5.6.40", 2014}, {"PHP 7.4", 2019}},
	"perl4":    {{"Perl 4.019", 1991}, {"Perl 4.036", 1993}},
	"cbtr":     {{"Microsoft C 6.0 + Btrieve 5.1", 1990}, {"Watcom C 10.6 + Btrieve 6.15", 1995}, {"Visual C++ 6.0 + Pervasive.SQL 2000", 1999}},
	"sql":      {{"Ingres 6.4 isql", 1990}, {"Sybase isql", 1993}, {"PostgreSQL 7.4 psql", 2003}, {"PostgreSQL 9.6 psql", 2016}},
	"db2sql":   {{"DB2 V1.3 DSNTEP2", 1985}, {"DB2 V3 DSNTEP2", 1993}, {"DB2 V7 DSNTEP2", 2001}, {"DB2 V10 for z/OS", 2010}, {"Db2 12 for z/OS", 2016}},
	"sqlds":    {{"SQL/DS V1R4 ISQL", 1985}, {"SQL/DS V3R3", 1991}},
	"sqlplus":  {{"SQL*Plus 2.0 (Oracle 5)", 1988}, {"SQL*Plus 3.1 (Oracle 7)", 1993}, {"SQL*Plus 8.1.7", 1999}, {"SQL*Plus 10.2", 2005}},
	"plsql":    {{"PL/SQL 2.0 (Oracle 7)", 1992}, {"PL/SQL 8.1.7", 1999}, {"PL/SQL 10.2", 2005}, {"PL/SQL 12.1", 2013}, {"PL/SQL 19c", 2019}},
	"tsql":     {{"Sybase SQL Server 4.2", 1993}, {"SQL Server 6.5", 1996}, {"SQL Server 2000", 2000}, {"SQL Server 2008 R2", 2010}, {"SQL Server 2019", 2019}},
	"sqlpl":    {{"DB2 UDB 7.1 SQL PL", 1999}, {"DB2 UDB 8.2", 2004}, {"DB2 9.7 LUW", 2009}, {"Db2 11.5 LUW", 2019}},
	"perl": {{"Perl 5.004", 1997}, {"Perl 5.005_03", 1999}, {"Perl 5.6.1", 2001}, {"Perl 5.8.8", 2006},
		{"Perl 5.10.1", 2009}, {"Perl 5.16.3", 2013}, {"Perl 5.26", 2017}, {"Perl 5.32", 2020}},
}

func langFamily(langKey string) string {
	if m, ok := langMetas[langKey]; ok {
		return m.family
	}
	return langKey
}

var abends = []string{"S0C7", "S0C4", "S0C1", "S322", "S806", "S878", "SB37", "SD37", "SE37", "U4038", "S013", "S213"}
var dayNames = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
var monthFull = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

var monthAbbr = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
var timesOfDay = []string{"6:00 AM", "7:30 AM", "noon", "2:00 PM", "4:45 PM", "6:00 PM", "9:00 PM", "11:30 PM", "midnight", "3:00 AM"}

// ── Slot filling ────────────────────────────────────────────────────────────

// gen holds the per-page state the slot filler draws from.
type gen struct {
	x        *rnd
	in       *Input
	id       ID
	roster   []*Coder
	vars     []string
	routines []string
	labels   []int
	langName string
	lang2    string
	vp       string
	dist     string
	emails   map[string]bool // honeypot addresses that appear on the page
	usedTpl  map[string]int
	holder   string
	yearFrom int
	depth    int // nested fill depth for fragments/vars values
}

func (g *gen) coder() *Coder { return pick(g.x, g.roster) }

func (g *gen) otherCoder(not *Coder) *Coder {
	for i := 0; i < 6; i++ {
		c := pick(g.x, g.roster)
		if c != not {
			return c
		}
	}
	return pick(g.x, g.roster)
}

// date returns a date string in [fromYear, toYear], never after the article
// date, formatted in the coder's habitual style.
func (g *gen) date(c *Coder, fromYear, toYear int) string {
	if toYear > g.in.Year {
		toYear = g.in.Year
	}
	if fromYear > toYear {
		fromYear = toYear
	}
	y := g.x.between(fromYear, toYear)
	m := g.x.between(1, 12)
	d := g.x.between(1, 28)
	if y == g.in.Year {
		if m > g.in.Month {
			m = g.in.Month
		}
		if m == g.in.Month && d > g.in.Day {
			d = g.in.Day
		}
		if d < 1 {
			d = 1
		}
	}
	style := 0
	if c != nil {
		style = c.DateStyle
	}
	switch style {
	case 1:
		return fmt.Sprintf("%02d/%02d/%02d", m, d, y%100)
	case 2:
		return fmt.Sprintf("%02d/%02d/%02d", d, m, y%100)
	case 3:
		return fmt.Sprintf("%02d-%s-%04d", d, monthAbbr[m-1], y)
	case 4:
		return fmt.Sprintf("%02d.%03d", y%100, (m-1)*30+d)
	default:
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}
}

// dateIn formats a specific year (random month/day, clamped to the article
// date) in the coder's style.
func (g *gen) dateIn(c *Coder, y int) string { return g.date(c, y, y) }

func (g *gen) memo() string {
	switch g.x.intn(5) {
	case 0:
		return fmt.Sprintf("%02d-%03d", g.x.between(70, 99), g.x.between(1, 450))
	case 1:
		return fmt.Sprintf("RD-%04d", g.x.between(10, 9999))
	case 2:
		return fmt.Sprintf("DPM/%02d/%02d", g.x.between(72, 99), g.x.between(1, 60))
	case 3:
		return fmt.Sprintf("CM-%d-%02d", g.x.between(1975, maxi(1976, g.in.Year)), g.x.between(1, 40))
	default:
		return fmt.Sprintf("%s-%03d", pick(g.x, []string{"HR", "IS", "EXEC", "COMP", "AUD", "OPS"}), g.x.between(1, 999))
	}
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (g *gen) quote() string {
	if len(g.in.Sentences) == 0 {
		return "the numbers are what they are"
	}
	s := pick(g.x, g.in.Sentences)
	s = strings.NewReplacer("\"", "", "{", "", "}", "").Replace(s)
	words := strings.Fields(s)
	if len(words) > 18 {
		start := g.x.intn(len(words) - 14)
		words = words[start : start+g.x.between(10, 14)]
	}
	return strings.TrimRight(strings.Join(words, " "), ".,;:")
}

func (g *gen) teamMember() Person {
	if len(g.in.Team) == 0 {
		return Person{Name: "the engagement team", Title: "Engagement Team"}
	}
	return pick(g.x, g.in.Team)
}

func (g *gen) hexToken() string {
	if len(g.in.Hex) == 0 {
		return strings.ToUpper(g.in.Token)
	}
	return strings.ToUpper(pick(g.x, g.in.Hex))
}

func (g *gen) metricName() string {
	var names []string
	for _, m := range g.in.Metrics {
		names = append(names, m.Name)
	}
	for _, b := range g.in.Bench {
		names = append(names, b.Name)
	}
	if len(names) == 0 {
		return "headcount"
	}
	return pick(g.x, names)
}

// fill expands every {SLOT} in tpl. author is the comment's author (the
// {CODER} slot); other coders are drawn per occurrence.
func (g *gen) fill(tpl string, author *Coder) string {
	var b strings.Builder
	second := g.otherCoder(author)
	third := g.otherCoder(second)
	for {
		i := strings.IndexByte(tpl, '{')
		if i < 0 {
			b.WriteString(tpl)
			break
		}
		j := strings.IndexByte(tpl[i:], '}')
		if j < 0 {
			b.WriteString(tpl)
			break
		}
		b.WriteString(tpl[:i])
		b.WriteString(g.slot(tpl[i+1:i+j], author, second, third))
		tpl = tpl[i+j+1:]
	}
	return b.String()
}

func (g *gen) slot(name string, a, c2, c3 *Coder) string {
	x := g.x
	in := g.in
	switch name {
	case "CODER":
		return a.Display(x)
	case "CODER.I":
		return a.Initials
	case "CODER2":
		return c2.Display(x)
	case "CODER2.I":
		return c2.Initials
	case "CODER3":
		return c3.Display(x)
	case "CODER3.I":
		return c3.Initials
	case "DATE", "DATE2", "DATE3":
		return g.date(a, a.Start, a.End)
	case "SHORTDATE":
		return fmt.Sprintf("%02d/%02d", x.between(1, 12), x.between(1, 28))
	case "OLDDATE":
		return g.date(a, 1966, mini(1985, in.Year))
	case "YEAR":
		return strconv.Itoa(x.between(mini(1975, in.Year), in.Year))
	case "OLDYEAR":
		return strconv.Itoa(x.between(1966, mini(1989, in.Year)))
	case "YEAR2":
		return strconv.Itoa(in.Year + x.between(1, 6))
	case "Y2KDATE", "Y2KDATE2":
		if in.Year < 1998 {
			return g.date(a, 1980, in.Year)
		}
		return g.date(a, 1998, mini(1999, in.Year))
	case "Y2KYEAR":
		return strconv.Itoa(x.between(1997, mini(1999, maxi(1997, in.Year))))
	case "PIVOT":
		return strconv.Itoa(x.between(20, 60))
	case "MEMO":
		return g.memo()
	case "AUDIT":
		return fmt.Sprintf("%02d-%02d", x.between(78, 99), x.between(1, 40))
	case "FORM":
		return fmt.Sprintf("RD-%d", x.between(100, 990))
	case "TICKET":
		return fmt.Sprintf("%s-%05d", pick(x, []string{"PMR", "APAR", "CR", "INC", "TKT", "SR", "PTR"}), x.between(100, 99999))
	case "HW":
		return eraPick(x, hardware, in.Year)
	case "HW2":
		return eraPick(x, hardware, in.Year)
	case "OS":
		return eraPick(x, operatingSystems, in.Year)
	case "COMPILER", "COMPILER2":
		return eraPick(x, compilers[langFamily(g.id.LangKey)], in.Year)
	case "DEPT", "DEPT2":
		return pick(x, getLibrary().names["depts"])
	case "DEPTCODE":
		return pick(x, []string{"DP", "IS", "MIS", "RD", "OPS", "TAB", "SYS", "APP"})
	case "VP":
		return g.vp
	case "FLOOR":
		return fmt.Sprintf("the %s floor", ordinal(x.between(2, 14)))
	case "BLDG":
		return pick(x, []string{"the Wells Street building", "Building C", "the annex", "the Mitchell Street office", "the old warehouse", "Tower 2", "the basement data center"})
	case "ROOM":
		return strconv.Itoa(x.between(100, 1499))
	case "EXT":
		return strconv.Itoa(x.between(2000, 7999))
	case "TERM":
		return fmt.Sprintf("T%04d", x.between(1, 4095))
	case "PRT":
		return fmt.Sprintf("PRT%d", x.between(1, 9))
	case "CLASS":
		return string(rune('A' + x.intn(8)))
	case "USERID":
		return a.UserID
	case "HOST":
		return pick(x, []string{"LICSRV1", "acpwb-lic01", "MKEVM01", "SYSB", "ntserver3", "dc1-batch02"})
	case "VOLSER":
		return fmt.Sprintf("%s%04d", pick(x, []string{"AC", "RD", "TP", "SC", "BK"}), x.between(1, 9999))
	case "ABEND":
		return pick(x, abends)
	case "TIME":
		return pick(x, timesOfDay)
	case "DAYNAME":
		return pick(x, dayNames)
	case "SEQ":
		return fmt.Sprintf("%s%05d", g.id.Program[:3], x.between(1, 900)*10)
	case "DIST":
		if g.dist != "" {
			return g.dist
		}
		return "INTERNAL USE ONLY"
	case "PROJECT":
		return pick(x, getLibrary().names["projects"])
	case "PROGRAM":
		return g.id.Program
	case "PROGRAM2":
		return pick(x, programNames)
	case "REL":
		return g.id.Release
	case "LANG":
		return g.langName
	case "LANG2":
		return g.lang2
	case "N", "N2":
		return strconv.Itoa(x.between(12, 400))
	case "SMALLN", "SMALLN2", "SMALLN3":
		return strconv.Itoa(x.between(2, 9))
	case "BIGN":
		return pick(x, []string{"999999", "9999999", "500000", "2147483647", "32767", "65535", "1000000"})
	case "PCT":
		return strconv.Itoa(x.between(1, 15))
	case "CPU":
		return fmt.Sprintf("%d.%02d", x.between(0, 3), x.between(1, 99))
	case "LIT":
		return pick(x, ladderLit[:])
	case "HEXN":
		return g.hexToken()
	case "METRIC":
		return g.metricName()
	case "PEER":
		return in.PeerGroup
	case "CLIENT":
		return in.Org
	case "INDUSTRY":
		return in.Industry
	case "RECORD":
		return strings.ToUpper(in.RecordID)
	case "ENGCODE":
		return in.EngCode
	case "TEAM":
		return g.teamMember().Name
	case "TEAMTITLE":
		return g.teamMember().Title
	case "REPORT":
		if len(in.RelatedTitles) == 0 {
			return in.Title
		}
		return pick(x, in.RelatedTitles)
	case "QUOTE":
		return g.quote()
	case "EMAIL", "EMAIL2":
		c := a
		if name == "EMAIL2" {
			c = c2
		}
		return g.email(c)
	case "VAR", "VAR2":
		if len(g.vars) == 0 {
			return "IDEBUG"
		}
		return pick(x, g.vars)
	case "ROUTINE", "ROUTINE2":
		if len(g.routines) == 0 {
			return "MAIN"
		}
		return pick(x, g.routines)
	case "LABEL", "LABEL2":
		if len(g.labels) == 0 {
			return strconv.Itoa(x.between(1, 99) * 10)
		}
		return strconv.Itoa(pick(x, g.labels))
	case "VENDOR":
		return pick(x, getLibrary().names["vendors"])
	case "HOLDER":
		return g.holder
	case "HOLDERUP":
		return strings.ToUpper(g.holder)
	case "YEARRANGE":
		if g.yearFrom >= in.Year {
			return strconv.Itoa(in.Year)
		}
		return fmt.Sprintf("%d, %d", g.yearFrom, in.Year)
	case "ADDRESS":
		return pick(x, []string{"100 East Wisconsin Avenue, Milwaukee, Wisconsin 53202", "P.O. Box 2006, Milwaukee, WI 53201",
			"777 North Water Street, Suite " + strconv.Itoa(x.between(100, 2400)) + ", Milwaukee, Wisconsin 53202",
			"Research Division, 1 Public Well Being Plaza, Milwaukee, WI 53203"})
	case "STATE":
		return "Wisconsin"
	case "COUNTY":
		return "Milwaukee County"
	case "CITY":
		return "Milwaukee"
	case "LICNO":
		return fmt.Sprintf("%s-%d-%05d", pick(x, []string{"LIC", "RD", "SL", "ACP", "SW"}), in.Year, x.between(1, 99999))
	case "PRODNO":
		return fmt.Sprintf("%d-%s%d", x.between(5600, 5799), pick(x, []string{"A", "B", "C", "D", "R"}), x.between(10, 99))
	case "SERIAL":
		return fmt.Sprintf("%02X-%05d", x.between(0x10, 0xFF), x.between(1, 99999))
	case "FEE":
		return fmt.Sprintf("%d,%03d", x.between(1, 48), x.between(0, 999))
	case "SEATS":
		return strconv.Itoa(pick(x, []int{1, 2, 5, 10, 25, 50}))
	case "TERMDAYS":
		return strconv.Itoa(pick(x, []int{15, 30, 45, 60, 90, 180}))
	case "LICVER":
		return fmt.Sprintf("%d.%d", x.between(1, 3), x.between(0, 9))
	case "DRINK":
		return pick(x, []string{"coffee", "Leinenkugel's", "Spotted Cow", "root beer", "brandy old fashioned", "cup of tea", "Pabst"})
	case "CONTRACT":
		return fmt.Sprintf("%s-%02d-C-%04d", pick(x, []string{"DOL", "OPM", "GS", "HHS", "DAAB"}), in.Year%100, x.between(1, 9999))
	case "AGENCY":
		return pick(x, []string{"U.S. Department of Labor", "U.S. Office of Personnel Management", "General Services Administration", "Department of Health and Human Services"})
	case "FARCLAUSE":
		return pick(x, []string{"FAR 52.227-14 (Rights in Data - General)", "DFARS 252.227-7014", "FAR 52.227-19"})
	case "TIER":
		return pick(x, []string{"Group 10", "Group 40", "AWLC Level 3", "Base", "Tier C"})
	case "SEAL":
		return strings.ToUpper(in.Token)
	case "CERT":
		return fmt.Sprintf("Y2K-%s-%04d", pick(x, []string{"RD", "IS", "TF", "CERT"}), x.between(1, 9999))
	case "TITLE":
		return in.Title
	}
	if v, ok := g.varSlot(name, a); ok {
		return v
	}
	return unknownSlot(name)
}

// varSlot fills a slot from fragments/vars/<slot>.txt, keeping only values
// whose era range covers the article year. Values may themselves contain
// slots (one level of nesting).
func (g *gen) varSlot(name string, a *Coder) (string, bool) {
	pool, ok := getLibrary().vars[name]
	if !ok {
		return "", false
	}
	var fit []string
	for _, v := range pool {
		if g.in.Year >= v.From && g.in.Year <= v.To {
			fit = append(fit, v.Text)
		}
	}
	if len(fit) == 0 {
		return strings.ToLower(name), true
	}
	s := pick(g.x, fit)
	if strings.IndexByte(s, '{') >= 0 && g.depth < 2 {
		g.depth++
		s = g.fill(s, a)
		g.depth--
	}
	return s, true
}

// unknownSlot is the fallback for a slot with no handler; tests assert it is
// never reached for any slot used in the fragment files.
var unknownSlot = func(name string) string { return strings.ToLower(name) }

func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ordinal(n int) string {
	suf := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suf = "st"
		case 2:
			suf = "nd"
		case 3:
			suf = "rd"
		}
	}
	return strconv.Itoa(n) + suf
}

// ── Comment generation ──────────────────────────────────────────────────────

// comment returns one generated comment (a paragraph of text, unwrapped),
// either from a whole-line template in one of cats (or the author's
// trait-biased categories) or composed from grammar parts.
func (g *gen) comment(author *Coder, cats ...string) string {
	L := getLibrary()
	if author == nil {
		author = g.coder()
	}
	if len(cats) == 0 {
		if g.x.chance(0.55) {
			cats = traitCats[author.Trait]
		} else {
			cats = L.cats
		}
	}
	if len(cats) == 1 || g.x.chance(0.7) {
		for tries := 0; tries < 8; tries++ {
			cat := pick(g.x, cats)
			if cat == "y2k" && g.in.Year < 1998 {
				continue
			}
			tpls := L.comments[cat]
			if len(tpls) == 0 {
				continue
			}
			tpl, ok := eraText(pick(g.x, tpls), g.in.Year)
			if !ok {
				continue
			}
			if g.usedTpl[tpl] > 0 && tries < 7 {
				continue
			}
			if g.in.Year < 1998 && strings.Contains(tpl, "{Y2K") {
				continue
			}
			g.usedTpl[tpl]++
			return g.fill(tpl, author)
		}
	}
	return g.composed(author)
}

// composed builds a sentence from the grammar parts:
// opener × subject × verb × complication × closer × signoff.
func (g *gen) composed(author *Coder) string {
	gr := getLibrary().grammar
	parts := []string{}
	if g.x.chance(0.7) {
		parts = append(parts, pick(g.x, gr["opener"]))
	}
	subj := pick(g.x, gr["subject"])
	// Capitalize when the subject starts a sentence (no opener, or an opener
	// that is itself a complete sentence such as "Hello again.").
	if len(parts) == 0 || strings.ContainsAny(parts[0][len(parts[0])-1:], ".!?") || strings.HasSuffix(parts[0], "}") {
		subj = strings.ToUpper(subj[:1]) + subj[1:]
	}
	parts = append(parts, subj, pick(g.x, gr["verb"]))
	if g.x.chance(0.8) {
		parts = append(parts, pick(g.x, gr["complication"]))
	}
	s := strings.Join(parts, " ") + "."
	if g.x.chance(0.6) {
		s += " " + pick(g.x, gr["closer"])
	}
	if g.x.chance(0.5) {
		s += " " + pick(g.x, gr["signoff"])
	}
	return g.fill(s, author)
}

// email returns (and registers) a coder's first.last@acpwb.com address.
// Registered addresses are pushed to the code-email queue by the handler so
// inbound spam can be traced back to this page.
func (g *gen) email(c *Coder) string {
	addr := CodeEmail(c.First, c.Last)
	g.emails[addr] = true
	return addr
}
