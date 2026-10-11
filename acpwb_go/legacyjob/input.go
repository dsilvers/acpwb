package legacyjob

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"unicode"
)

// Input is everything a surface (archive page today, policy pages later)
// hands the generator. It deliberately carries no archive types so this
// package stays surface-agnostic and import-cycle free (archive imports
// legacyjob for Identity, not the other way round).
type Input struct {
	Year, Month, Day int
	Slug             string
	Host             string // request host, for emails registry only
	Path             string // canonical article path (seed for code emails)

	Token    string // md5("acpwb_archive_"+slug)[:8] — page watermark
	RecordID string

	Title, BaseTitle, Org, Industry, Phase string
	EngCode, DocVersion, Distribution      string
	PageCount, FileSizeKB                  int
	PeerGroup                              string

	Metrics []MetricIn
	Bench   []BenchIn

	Team      []Person   // engagement team (name/title/email)
	Revisions []Revision // document revision history

	Sentences     []string // findings / exec bullets / paragraphs / footnotes
	RelatedTitles []string // referenced report titles
	Hex           []string // high-entropy tokens from the page's bulk hex

	Links []CrossLink // other archive pages, for the cross-linked code web

	// The site's shared generated-name pools (same ones the engagement team
	// and employee generators use), so programmers look like every other
	// ACPWB person and get the same first.last@acpwb.com addresses.
	FirstNames, LastNames []string

	ArticleURL    string
	ExportCSVURL  string
	MetricsAnchor string
	BenchAnchor   string

	// Navigation: source-code index pages and neighbouring entries.
	IndexURL, YearURL, MonthURL string
	Prev, Next                  *MoreLink
	More                        []MoreGroup
}

// MetricIn is one Engagement Metrics row, parsed back from the page's own
// formatted strings so the listing can never drift from the table.
type MetricIn struct {
	Name              string
	Baseline, Current int
}

// BenchIn is one Market Benchmark row. P holds the eight published values
// (P10…P95) exactly as displayed; Median == P[3].
type BenchIn struct {
	Name   string
	Median int
	P      [8]int
}

type Person struct{ Name, Title, Email string }

type Revision struct{ Version, Date, Description, Author string }

// MoreLink / MoreGroup feed the "Related software projects" block at the bottom of
// the page (other archive articles' /source-code/ pages).
type MoreLink struct {
	URL, Title, Program, Language, Date, Org, Job string
}

type MoreGroup struct {
	Title     string
	URL, More string // optional "see all" link
	Links     []MoreLink
}

type CrossLink struct {
	URL              string // URL of the target article's /source-code/ page
	Label, Date      string
	Year, Month, Day int
	Slug             string
}

// ladderNum is the published percentile ladder as integer hundredths.
var ladderNum = [8]int{58, 78, 88, 100, 114, 128, 162, 184}
var ladderLit = [8]string{"0.58", "0.78", "0.88", "1.00", "1.14", "1.28", "1.62", "1.84"}
var ladderCol = [8]string{"P10", "P25", "P33", "P50", "P67", "P75", "P90", "P95"}

// ParseInt strips "$", "," and whitespace and parses a signed integer.
func ParseInt(s string) int {
	var b strings.Builder
	for _, r := range s {
		if r == '-' || r == '+' || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	n, _ := strconv.Atoi(b.String())
	return n
}

// WatermarkToken matches the archive CSV export's token formula.
func WatermarkToken(slug string) string {
	sum := md5.Sum([]byte("acpwb_archive_" + slug))
	return hex.EncodeToString(sum[:])[:8]
}

var asciiFold = map[rune]string{
	'—': "-", '–': "-", '‘': "'", '’': "'", '“': "\"", '”': "\"", '…': "...",
	'é': "E", 'è': "E", 'á': "A", 'à': "A", 'ó': "O", 'ö': "O", 'ü': "U", 'ñ': "N",
	'ç': "C", 'í': "I", '&': "AND",
}

// cardText upper-cases and folds a string to the keypunch character set
// (no lowercase, ASCII only), the form everything takes once it is on a card.
func cardText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if rep, ok := asciiFold[r]; ok {
			b.WriteString(rep)
			continue
		}
		if r > 126 {
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// asciiText folds a string to printable ASCII without changing case (for
// languages whose sources are mixed case but still 7-bit).
func asciiText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if rep, ok := asciiFold[r]; ok && r != '&' {
			b.WriteString(strings.ToLower(rep))
			if r == '—' || r == '–' || r == '‘' || r == '’' || r == '“' || r == '”' || r == '…' {
				continue
			}
			continue
		}
		if r > 126 || (r < 32 && r != '\t' && r != '\n') {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padLeft(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return strings.Repeat(" ", n-len(s)) + s
}

// wrapWords greedily wraps text to width (by visible length).
func wrapWords(text string, width int) []string {
	if width < 10 {
		width = 10
	}
	var lines []string
	var cur strings.Builder
	for _, w := range strings.Fields(text) {
		if cur.Len() > 0 && visibleLen(cur.String())+1+visibleLen(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// Link markers embedded in listing text: \x01<index>\x02<text>\x03. The
// renderer turns them into <a> tags; visibleLen / stripMarkers ignore them.
const (
	mkOpen  = '\x01'
	mkMid   = '\x02'
	mkClose = '\x03'
)

func linkMarker(idx int, text string) string {
	return string(mkOpen) + strconv.Itoa(idx) + string(mkMid) + text + string(mkClose)
}

func stripMarkers(s string) string {
	if !strings.ContainsRune(s, mkOpen) {
		return s
	}
	var b strings.Builder
	skip := false
	for _, r := range s {
		switch {
		case r == mkOpen:
			skip = true
		case r == mkMid:
			skip = false
		case r == mkClose:
		case !skip:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func visibleLen(s string) int { return len(stripMarkers(s)) }
