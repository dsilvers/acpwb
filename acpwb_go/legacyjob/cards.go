package legacyjob

import (
	"fmt"
	"strings"
)

// IBM 029 keypunch (EBCDIC) Hollerith code: rune -> punched rows.
// Row ids: 12, 11, 0..9.
var hollerith = func() map[rune][]int {
	m := map[rune][]int{' ': nil}
	for d := 0; d <= 9; d++ {
		m[rune('0'+d)] = []int{d}
	}
	for i, r := range "ABCDEFGHI" {
		m[r] = []int{12, i + 1}
	}
	for i, r := range "JKLMNOPQR" {
		m[r] = []int{11, i + 1}
	}
	for i, r := range "STUVWXYZ" {
		m[r] = []int{0, i + 2}
	}
	set := func(r rune, rows ...int) { m[r] = rows }
	set('&', 12)
	set('-', 11)
	set('/', 0, 1)
	set('.', 12, 3, 8)
	set('<', 12, 4, 8)
	set('(', 12, 5, 8)
	set('+', 12, 6, 8)
	set('|', 12, 7, 8)
	set('!', 11, 2, 8)
	set('$', 11, 3, 8)
	set('*', 11, 4, 8)
	set(')', 11, 5, 8)
	set(';', 11, 6, 8)
	set(',', 0, 3, 8)
	set('%', 0, 4, 8)
	set('_', 0, 5, 8)
	set('>', 0, 6, 8)
	set('?', 0, 7, 8)
	set(':', 2, 8)
	set('#', 3, 8)
	set('@', 4, 8)
	set('\'', 5, 8)
	set('=', 6, 8)
	set('"', 7, 8)
	return m
}()

var rowOrder = []int{12, 11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

// Punches returns the punched rows for each of the 80 columns of a card.
func Punches(text string) [][]int {
	cols := make([][]int, 80)
	for i, r := range []rune(padRight(cardText(text), 80)) {
		if i >= 80 {
			break
		}
		cols[i] = hollerith[r]
	}
	return cols
}

// DecodePunches maps punched rows back to characters (for tests).
func DecodePunches(cols [][]int) string {
	rev := map[string]rune{}
	for r, rows := range hollerith {
		rev[fmt.Sprint(rows)] = r
	}
	var b strings.Builder
	for _, rows := range cols {
		if r, ok := rev[fmt.Sprint(rows)]; ok {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return b.String()
}

var cardColors = []struct{ bg, ink, label string }{
	{"#f3e7c4", "#8a6d3b", "manila"},
	{"#dfeccf", "#4f6b3a", "green"},
	{"#f6d5c6", "#8a4b33", "salmon"},
	{"#e7e2f2", "#5a4f86", "lilac"},
	{"#f4f1e4", "#6b675a", "white"},
}

// cardSVG draws one 80-column card: printed text along the top edge, the
// row digits, and rectangular holes for every punch.
func cardSVG(text, caption string, color int) string {
	c := cardColors[color%len(cardColors)]
	const w, h = 740.0, 326.0
	colX := func(i int) float64 { return 18 + float64(i)*8.85 }
	rowY := func(r int) float64 {
		for i, rr := range rowOrder {
			if rr == r {
				return 30 + float64(i)*23.4
			}
		}
		return 0
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" role="img" aria-label="Punched card: %s">`, w, h, esc(stripMarkers(text)))
	fmt.Fprintf(&b, `<path d="M18 0 H%.0f V%.0f H0 V18 Z" fill="%s" stroke="%s" stroke-width="1"/>`, w, h, c.bg, c.ink)
	// row digits as one text run per row (0-9)
	for d := 0; d <= 9; d++ {
		fmt.Fprintf(&b, `<text x="18" y="%.1f" font-family="IBM Plex Mono,Menlo,monospace" font-size="9" fill="%s" opacity=".55" textLength="700" lengthAdjust="spacing">%s</text>`,
			rowY(d)+4, c.ink, strings.Repeat(fmt.Sprint(d), 80))
	}
	// printed characters along the top edge
	fmt.Fprintf(&b, `<text x="18" y="16" font-family="IBM Plex Mono,Menlo,monospace" font-size="10.5" fill="#222" textLength="700" lengthAdjust="spacing">%s</text>`,
		esc(padRight(cardText(stripMarkers(text)), 80)))
	fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" font-family="IBM Plex Sans,Helvetica,sans-serif" font-size="9" fill="%s" text-anchor="end">%s</text>`, w-10, h-6, c.ink, esc(caption))
	for i, rows := range Punches(stripMarkers(text)) {
		for _, r := range rows {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="5" height="12" fill="#2b2b2b"/>`, colX(i)+1, rowY(r)-4)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func cardsHTML(j *job, srcLines []Line, deck []string) string {
	x := j.x
	var cards [][2]string
	if len(deck) > 0 {
		cards = append(cards, [2]string{deck[0], "SYSIN CARD 1 — CLIENT / RECORD"})
	}
	if len(deck) > 3 {
		cards = append(cards, [2]string{deck[3], "SYSIN CARD 4 — FIRST METRIC"})
	}
	var code []Line
	for _, l := range srcLines {
		if l.Code && len(strings.TrimSpace(l.Text)) > 10 {
			code = append(code, l)
		}
	}
	for n := x.between(1, 3); n > 0 && len(code) > 0; n-- {
		l := pick(x, code)
		cards = append(cards, [2]string{l.Text, "SOURCE DECK"})
	}
	var b strings.Builder
	b.WriteString(`<div class="lj-cards">`)
	base := x.intn(len(cardColors))
	for i, c := range cards {
		b.WriteString(cardSVG(c[0], c[1], base+i))
	}
	b.WriteString(`</div>`)
	return b.String()
}

// rd114HTML renders the change-request form that authorised this run.
func rd114HTML(j *job, l *Listing) string {
	x := j.x
	in := j.in
	team := in.Team
	req := Person{Name: "Engagement team", Title: "Engagement Team"}
	if len(team) > 0 {
		req = team[0]
	}
	appr := j.g.coder()
	appr2 := j.g.otherCoder(appr)
	desc := j.g.comment(appr, "revision", "data", "memos") + " " + j.g.comment(nil)
	field := func(label, val string) string {
		return fmt.Sprintf(`<td><span class="lj-fl">%s</span>%s</td>`, esc(label), esc(val))
	}
	var b strings.Builder
	b.WriteString(`<div class="lj-form"><div style="display:flex;justify-content:space-between;font-weight:700;margin-bottom:.5rem">`)
	fmt.Fprintf(&b, `<span>FORM RD-114 &mdash; PROGRAM CHANGE / RUN AUTHORIZATION</span><span>No. %06d</span></div><table>`, x.between(1000, 999999))
	b.WriteString(`<tr>` + field("Program", l.ID.Program+" rel. "+l.ID.Release) + field("Language", l.ID.Language) + field("Job", l.ID.JobNumber) + `</tr>`)
	b.WriteString(`<tr>` + field("Requested by", req.Name+", "+req.Title) + field("Cost center", in.EngCode) + field("Record", strings.ToUpper(in.RecordID)) + `</tr>`)
	b.WriteString(`<tr><td colspan="3"><span class="lj-fl">Description of change / reason for run</span>` + esc(desc) + `</td></tr>`)
	fmt.Fprintf(&b, `<tr>%s%s%s</tr>`,
		field("Approved (1)", appr.Name()+" — "+CodeEmail(appr.First, appr.Last)+" — "+j.g.date(appr, appr.Start, appr.End)),
		field("Approved (2)", appr2.Name()+" — "+CodeEmail(appr2.First, appr2.Last)+" — "+j.g.date(appr2, appr2.Start, appr2.End)),
		field("Distribution", in.Distribution))
	j.g.emails[CodeEmail(appr.First, appr.Last)] = true
	j.g.emails[CodeEmail(appr2.First, appr2.Last)] = true
	fmt.Fprintf(&b, `<tr><td colspan="3"><span class="lj-fl">Operations use only</span>Received %s &middot; Keypunched by %s &middot; Verified %s &middot; %s</td></tr>`,
		esc(j.g.date(nil, in.Year, in.Year)), esc(j.g.coder().Initials), esc(pick(x, []string{"YES", "YES", "N/A", "SEE BACK"})), esc(cardText(j.g.comment(nil, "superstition", "misc", "hardware"))))
	b.WriteString(`</table></div>`)
	return b.String()
}
