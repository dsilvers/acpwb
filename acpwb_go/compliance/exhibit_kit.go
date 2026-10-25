package compliance

// Small helpers shared by the exhibit generators: an SVG builder, phone and
// paper frames, and answer-choice builders.

import (
	"fmt"
	"strings"
	"time"
)

type svgB struct {
	strings.Builder
}

// newSVG starts an exhibit image w×h with an accessible label.
func newSVG(w, h float64, label string) *svgB {
	s := &svgB{}
	fmt.Fprintf(s, `<svg class="cmp-exhibit-img" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" role="img" aria-label="%s">`, w, h, w, h, esc(label))
	return s
}

func (s *svgB) rect(x, y, w, h, rx float64, fill, stroke string) {
	st := ""
	if stroke != "" {
		st = fmt.Sprintf(` stroke="%s"`, stroke)
	}
	fmt.Fprintf(s, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s"%s/>`, x, y, w, h, rx, fill, st)
}

// text draws escaped text; attrs is extra SVG attributes, such as
// `font-weight="700" fill="#222" text-anchor="middle"`.
func (s *svgB) text(x, y, size float64, attrs, txt string) {
	if !strings.Contains(attrs, "fill=") {
		attrs += ` fill="#222"`
	}
	fmt.Fprintf(s, `<text x="%.1f" y="%.1f" %s font-size="%.1f" %s>%s</text>`, x, y, svgFont, size, attrs, esc(txt))
}

func (s *svgB) line(x1, y1, x2, y2 float64, stroke string, w float64) {
	fmt.Fprintf(s, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f"/>`, x1, y1, x2, y2, stroke, w)
}

func (s *svgB) dash(x1, y1, x2, y2 float64, stroke string, w float64) {
	fmt.Fprintf(s, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-dasharray="5 4"/>`, x1, y1, x2, y2, stroke, w)
}

func (s *svgB) circle(cx, cy, r float64, fill, stroke string) {
	st := ""
	if stroke != "" {
		st = fmt.Sprintf(` stroke="%s" stroke-width="1.5"`, stroke)
	}
	fmt.Fprintf(s, `<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s"%s/>`, cx, cy, r, fill, st)
}

func (s *svgB) raw(format string, a ...any) { fmt.Fprintf(s, format, a...) }

func (s *svgB) done() string {
	s.WriteString(`</svg>`)
	return s.String()
}

// phone draws a phone frame of the given screen height and returns the
// builder; screen content starts at (20, 50) and is 220 wide.
func phone(h float64, label, title string) *svgB {
	s := newSVG(260, h+70, label)
	s.rect(5, 5, 250, h+60, 32, "#111", "")
	s.rect(15, 38, 230, h+12, 6, "#f2f2f7", "")
	s.rect(100, 16, 60, 12, 6, "#222", "")
	if title != "" {
		s.rect(15, 38, 230, 26, 0, "#e5e5ea", "")
		s.text(130, 55, 10, `text-anchor="middle" font-weight="700"`, title)
	}
	return s
}

// paper draws a sheet of paper with a header band.
func paper(w, h float64, label, title, sub string) *svgB {
	s := newSVG(w, h, label)
	s.rect(1, 1, w-2, h-2, 4, "#fffdf7", "#c9c2b0")
	s.rect(1, 1, w-2, 34, 4, "#1a3a5c", "")
	s.text(14, 22, 12, `font-weight="700" fill="#fff"`, title)
	if sub != "" {
		s.text(w-14, 22, 9, `text-anchor="end" fill="#c9d6e3"`, sub)
	}
	return s
}

// wrongsFrom picks up to n distinct items from pool that differ from right.
func wrongsFrom(r *rnd, right string, pool []string, n int) []string {
	var out []string
	for _, p := range sample(r, pool, len(pool)) {
		if p != right && p != "" && !contains(out, p) {
			out = append(out, p)
			if len(out) == n {
				break
			}
		}
	}
	return out
}

// oneOf builds choices from the right answer plus three wrongs from pool.
func oneOf(r *rnd, right string, pool []string) ([]string, int) {
	return withChoices(r, right, wrongsFrom(r, right, pool, 3))
}

func countQ(r *rnd, right int, unit string) ([]string, int) {
	return numChoices(r, right, []int{-1, 1, 2, -2, 3}, func(n int) string { return plural(n, unit) })
}

func moneyQ(r *rnd, cents int, deltas []int) ([]string, int) {
	return numChoices(r, cents, deltas, money)
}

func dollarQ(r *rnd, n int, deltas []int) ([]string, int) {
	return numChoices(r, n, deltas, dollars)
}

func pctQ(r *rnd, n int) ([]string, int) {
	return numChoices(r, n, []int{-5, 5, -10, 10, 3, -3}, func(v int) string { return fmt.Sprintf("%d%%", v) })
}

func dateQ(r *rnd, right time.Time, offs []int) ([]string, int) {
	var wrongs []string
	for _, d := range sample(r, offs, len(offs)) {
		w := longDate(right.AddDate(0, 0, d))
		if d != 0 && !contains(wrongs, w) && len(wrongs) < 3 {
			wrongs = append(wrongs, w)
		}
	}
	return withChoices(r, longDate(right), wrongs)
}

func longDate(t time.Time) string  { return t.Format("January 2, 2006") }
func shortDate(t time.Time) string { return t.Format("Jan 2, 2006") }

// people returns n distinct full names.
func people(r *rnd, n int) []string {
	firsts := sample(r, firstNames, n)
	out := make([]string, n)
	for i, f := range firsts {
		out[i] = f + " " + pick(r, lastNames)
	}
	return out
}

// lettered returns "A", "B", … for n options.
func lettered(i int) string { return string(rune('A' + i)) }

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// addBusinessDays adds n weekdays (Mon–Fri) to t.
func addBusinessDays(t time.Time, n int) time.Time {
	for n > 0 {
		t = t.AddDate(0, 0, 1)
		if t.Weekday() != time.Saturday && t.Weekday() != time.Sunday {
			n--
		}
	}
	return t
}

// weekday returns the next date on or after t that is a Monday–Friday.
func weekday(t time.Time) time.Time {
	for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// registerExhibits adds generators (as "ex-<kind>" too) and appends kinds to
// bank categories. Call it from an exhibits_*.go init, which runs after
// exhibits.go has set up the maps.
func registerExhibits(gens map[string]exhibitGen, cats map[string][]string) {
	for k, g := range gens {
		if _, dup := exhibitGens[k]; dup {
			panic("duplicate exhibit kind " + k)
		}
		exhibitGens[k] = g
		generatedKinds["ex-"+k] = true
	}
	exhibitNames = exhibitNames[:0]
	for k := range exhibitGens {
		exhibitNames = append(exhibitNames, k)
	}
	sortStrings(exhibitNames)
	for cat, ks := range cats {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// fmtNum formats v with d decimals.
func fmtNum(v float64, d int) string {
	return fmt.Sprintf("%.*f", d, v)
}

// pickN returns n distinct indexes from 0..m-1.
func pickN(r *rnd, m, n int) []int {
	o := make([]int, m)
	for i := range o {
		o[i] = i
	}
	shuffle(r, o)
	return o[:n]
}
