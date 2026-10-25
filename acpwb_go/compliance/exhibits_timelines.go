package compliance

import (
	"fmt"
	"math"
	"strings"
)

// Timeline exhibits: rows of bars on a time axis — driver logs, shifts,
// permits, escorts, change windows. Rules that come from law or a standard
// are the real ones: FMCSA hours of service (49 CFR 395.3), federal child
// labor hours for 14- and 15-year-olds (29 CFR 570.35), NFPA 51B fire
// watch, OSHA 1910.146(i) attendants, 1910.147(f)(4) shift-change lockout,
// FDA Food Code 3-501.19 time as a control, OSHA 1910.95 noise dose, and
// DOT post-accident testing windows (49 CFR 382.303).

type tlBar struct {
	a, b  float64
	color string
	label string
	half  int // 0 full height, 1 top half, 2 bottom half
}

type tlRow struct {
	name  string
	bars  []tlBar
	ticks []float64 // point events (check-ins, badge swipes, trades)
	tickC string
}

type timeline struct {
	title      string
	from, to   float64
	step       float64
	axis       func(float64) string
	rows       []tlRow
	shade      [][2]float64 // shaded windows
	shadeLabel string
	legend     [][2]string // color, label
	note       string
	labelTicks bool // print tick times above ticks
	tickFormat func(float64) string
	shadeColor string
	nameWidth  float64
	centered   bool // labels sit between grid lines (days)
}

const (
	tlBlue  = "#2563eb"
	tlAmber = "#f59e0b"
	tlGray  = "#94a3b8"
	tlRed   = "#dc2626"
	tlGreen = "#16a34a"
)

func (t timeline) svg() string {
	nw := t.nameWidth
	if nw == 0 {
		nw = 118
	}
	x0, x1 := nw+8, 528.0
	top := 50.0
	if len(t.legend) > 0 {
		top = 62
	}
	rh := 30.0
	h := top + rh*float64(len(t.rows)) + 34
	if t.note != "" {
		h += 14
	}
	s := frame(540, h, t.title+" timeline", t.title)
	X := func(v float64) float64 { return x0 + (x1-x0)*(v-t.from)/(t.to-t.from) }
	lx := 14.0
	for _, l := range t.legend {
		s.rect(lx, 30, 12, 9, 2, l[0], "")
		s.text(lx+16, 38, 8, `fill="#334155"`, l[1])
		lx += 26 + 5.2*float64(len([]rune(l[1])))
	}
	bot := top + rh*float64(len(t.rows))
	sc := t.shadeColor
	if sc == "" {
		sc = "#bbf7d0"
	}
	for _, w := range t.shade {
		s.rect(X(w[0]), top-4, X(w[1])-X(w[0]), bot-top+4, 0, sc, "")
		if t.shadeLabel != "" {
			s.text((X(w[0])+X(w[1]))/2, top-7, 7.5, `text-anchor="middle" font-weight="700" fill="#166534"`, t.shadeLabel)
		}
	}
	for v := t.from; v <= t.to+1e-9; v += t.step {
		s.line(X(v), top-4, X(v), bot, "#e2e8f0", 1)
		if !t.centered {
			s.text(X(v), bot+12, 7.5, `text-anchor="middle" fill="#475569"`, t.axis(v))
		} else if v+t.step <= t.to+1e-9 {
			s.text(X(v+t.step/2), bot+12, 7.5, `text-anchor="middle" fill="#475569"`, t.axis(v+t.step/2))
		}
	}
	for i, row := range t.rows {
		y := top + rh*float64(i)
		s.line(x0, y+rh, x1, y+rh, "#f1f5f9", 1)
		s.text(nw, y+rh/2+3, 8.5, `text-anchor="end" font-weight="700"`, row.name)
		for _, b := range row.bars {
			by, bh := y+6, rh-12
			switch b.half {
			case 1:
				by, bh = y+4, rh/2-5
			case 2:
				by, bh = y+rh/2+1, rh/2-5
			}
			s.rect(X(b.a), by, X(b.b)-X(b.a), bh, 2, b.color, "")
			if b.label != "" && X(b.b)-X(b.a) > 5.5*float64(len([]rune(b.label)))+4 {
				fill := "#fff"
				if b.color == "#cbd5e1" {
					fill = "#334155"
				}
				s.text((X(b.a)+X(b.b))/2, by+bh/2+3, 7.5, `text-anchor="middle" font-weight="700" fill="`+fill+`"`, b.label)
			}
		}
		tc := row.tickC
		if tc == "" {
			tc = "#111827"
		}
		for _, tk := range row.ticks {
			s.raw(`<path d="M%.1f %.1f l-4 -7 h8 z" fill="%s"/>`, X(tk), y+rh/2+2, tc)
			s.line(X(tk), y+rh/2+2, X(tk), y+rh-5, tc, 1.5)
			if t.labelTicks {
				f := t.tickFormat
				if f == nil {
					f = t.axis
				}
				s.text(X(tk), y+8, 6.5, `text-anchor="middle" fill="#334155"`, f(tk))
			}
		}
	}
	if t.note != "" {
		s.text(14, h-8, 8, `fill="#334155"`, t.note)
	}
	return s.done()
}

// hourAxis labels hours (fractional, may exceed 24) as clock times.
func hourAxis(v float64) string { return clockShort(int(math.Round(v)) % 24) }

// hm formats fractional hours as a clock time.
func hm(v float64) string {
	m := int(math.Round(v*60)) % (24 * 60)
	return clock(m)
}

func hmRange(a, b float64) string { return hm(a) + "–" + hm(b) }

// hmBare is a clock time without a.m./p.m., for crowded axes.
func hmBare(v float64) string {
	t := hm(v)
	return t[:strings.IndexByte(t, ' ')]
}

func init() {
	registerExhibits(map[string]exhibitGen{
		"hoslog":       hosLogExhibit,
		"minorhours":   minorHoursExhibit,
		"loneworker":   loneWorkerExhibit,
		"firewatch":    fireWatchExhibit,
		"attendant":    attendantExhibit,
		"escortlog":    escortLogExhibit,
		"changewindow": changeWindowExhibit,
		"blackout":     blackoutExhibit,
		"doublebooked": doubleBookedExhibit,
		"focusblock":   focusBlockExhibit,
		"meetingslot":  meetingSlotExhibit,
		"corehours":    coreHoursExhibit,
		"fatigue":      fatigueExhibit,
		"shuttle":      shuttleExhibit,
		"afterhours":   afterHoursExhibit,
		"screenbreak":  screenBreakExhibit,
		"oncall":       onCallExhibit,
		"tphc":         tphcExhibit,
		"noisedose":    noiseDoseExhibit,
		"posttest":     postTestExhibit,
		"dockhours":    dockHoursExhibit,
		"lockhandoff":  lockHandoffExhibit,
		"backupwindow": backupWindowExhibit,
		"coverage":     coverageExhibit,
	}, map[string][]string{
		"driving":         {"hoslog", "posttest", "fatigue"},
		"vehicles":        {"hoslog", "posttest"},
		"fleet":           {"hoslog", "posttest"},
		"deliveries":      {"hoslog", "dockhours"},
		"substances":      {"posttest"},
		"internships":     {"minorhours"},
		"wage":            {"minorhours"},
		"hiring":          {"minorhours"},
		"kids":            {"minorhours"},
		"nightshift":      {"loneworker", "fatigue"},
		"fieldwork":       {"loneworker"},
		"safety":          {"loneworker", "firewatch"},
		"hotwork":         {"firewatch"},
		"fire":            {"firewatch"},
		"construction":    {"firewatch", "attendant"},
		"spaces":          {"attendant"},
		"lockout":         {"lockhandoff"},
		"machineguarding": {"lockhandoff"},
		"visitor":         {"escortlog"},
		"physicalsec":     {"escortlog", "afterhours"},
		"badge":           {"afterhours", "escortlog"},
		"changemgmt":      {"changewindow"},
		"releases":        {"changewindow"},
		"network":         {"changewindow", "backupwindow"},
		"backups":         {"backupwindow"},
		"insider":         {"blackout"},
		"disclosure":      {"blackout"},
		"equity":          {"blackout"},
		"conferences":     {"doublebooked"},
		"events":          {"doublebooked"},
		"calendar":        {"focusblock", "meetingslot"},
		"focus":           {"focusblock"},
		"meetings":        {"meetingslot", "focusblock"},
		"hybrid":          {"corehours"},
		"remote":          {"corehours"},
		"globalteams":     {"corehours", "meetingslot"},
		"commuting":       {"shuttle"},
		"parking":         {"shuttle"},
		"ergonomics":      {"screenbreak"},
		"eyestrain":       {"screenbreak"},
		"oncall":          {"oncall"},
		"tickets":         {"oncall"},
		"foodsafety":      {"tphc"},
		"potluck":         {"tphc"},
		"catering":        {"tphc"},
		"hearing":         {"noisedose"},
		"warehouse":       {"noisedose", "dockhours"},
		"mail":            {"dockhours"},
		"pto":             {"coverage"},
		"timeoff":         {"coverage"},
		"managers":        {"coverage"},
	})
}

// names4 returns n distinct "J. Smith" style names.
func names4(r *rnd, n int) []string {
	for {
		out := people(r, n)
		seen := map[string]bool{}
		for i, p := range out {
			f, l, _ := strings.Cut(p, " ")
			out[i] = f[:1] + ". " + l
			seen[out[i]] = true
		}
		if len(seen) == n {
			return out
		}
	}
}

func whoQ(r *rnd, names []string, bad int) ([]string, int) {
	var wrongs []string
	for i, n := range names {
		if i != bad {
			wrongs = append(wrongs, n)
		}
	}
	return withChoices(r, names[bad], wrongs)
}

// ── FMCSA hours of service ──────────────────────────────────────────────

type hosSeg struct {
	drive bool
	off   bool
	d     float64
}

// hosViolation: >11 h driving, driving after the 14th hour, or more than
// 8 h of driving without a 30-minute interruption (on-duty not driving
// counts as a break since the 2020 rule).
func hosViolation(segs []hosSeg) bool {
	t, drive, since := 0.0, 0.0, 0.0
	for _, s := range segs {
		if s.drive {
			drive += s.d
			since += s.d
			if drive > 11 || t+s.d > 14 || since > 8 {
				return true
			}
		} else if s.d >= 0.5 {
			since = 0
		}
		t += s.d
	}
	return false
}

func hosDay(r *rnd, wantBad bool) []hosSeg {
	for {
		segs := []hosSeg{{false, false, 0.5}}
		drive := 0.0
		target := 8.5 + r.float()*3.5
		for drive < target {
			d := float64(r.between(6, 18)) / 4
			segs = append(segs, hosSeg{true, false, d})
			drive += d
			nd := float64(r.between(1, 8)) / 4
			off := r.chance(0.4)
			if r.chance(0.15) {
				nd = float64(r.between(8, 14)) / 4
			}
			segs = append(segs, hosSeg{false, off, nd})
		}
		total := 0.0
		for _, sg := range segs {
			total += sg.d
		}
		if total <= 20 && hosViolation(segs) == wantBad {
			return segs
		}
	}
}

func hosLogExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i := range names {
		t := float64(r.between(16, 28)) / 4
		var bars []tlBar
		for _, s := range hosDay(r, i == bad) {
			col, lbl := tlAmber, ""
			if s.drive {
				col, lbl = tlBlue, fmt.Sprintf("%.4gh", s.d)
			} else if s.off {
				col = tlGray
			}
			bars = append(bars, tlBar{t, t + s.d, col, lbl, 0})
			t += s.d
		}
		rows = append(rows, tlRow{name: names[i], bars: bars})
	}
	tl := timeline{title: "DRIVER DAILY LOGS — PROPERTY-CARRYING CMV", from: 4, to: 28, step: 2, axis: hourAxis, rows: rows,
		legend: [][2]string{{tlBlue, "Driving"}, {tlAmber, "On duty, not driving"}, {tlGray, "Off duty"}},
		note:   "Each driver starts the duty day after 10 or more consecutive hours off duty."}
	q := Question{key: "ex/hoslog", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Under FMCSA hours-of-service rules (11 hours of driving, a 14-hour duty window, and a 30-minute break after 8 hours of driving), which driver's log shows a violation?",
		"Refer to %[1]s. Which driver violated the federal hours-of-service rules for property-carrying drivers?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── 14- and 15-year-olds on a school day ────────────────────────────────

func minorHoursExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	tooLong := r.chance(0.5)
	var rows []tlRow
	for i := range names {
		var a, b float64
		for {
			a = float64(r.between(62, 68)) / 4 // 3:30–5:00 p.m.
			b = a + float64(r.between(6, 12))/4
			if b <= 19 {
				break
			}
		}
		if i == bad {
			if tooLong {
				a = float64(r.between(60, 62)) / 4
				b = a + float64(r.between(14, 15))/4
			} else {
				b = float64(r.between(77, 82)) / 4
				a = b - float64(r.between(8, 11))/4
			}
		}
		rows = append(rows, tlRow{name: names[i] + " (15)", bars: []tlBar{{a, b, tlBlue, "", 0}}})
	}
	tl := timeline{title: "AFTER-SCHOOL SHIFTS — TUESDAY, OCTOBER 14 (SCHOOL DAY)", from: 14, to: 22, step: 0.5, axis: hmBare, rows: rows,
		note: "All four associates are 15 years old and work in the café."}
	q := Question{key: "ex/minorhours", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Under federal child labor rules for 14- and 15-year-olds, which shift is NOT allowed on a school day?",
		"Refer to %[1]s. Federal rules limit 14- and 15-year-olds to 3 hours on a school day, and (outside summer) no work after 7 p.m. Which shift breaks the rules?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, rows4Names(rows), bad)
	return q
}

func rows4Names(rows []tlRow) []string {
	out := make([]string, len(rows))
	for i, rw := range rows {
		out[i] = rw.name
	}
	return out
}

// ── lone worker check-ins ───────────────────────────────────────────────

func loneWorkerExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i := range names {
		start := float64(r.between(18, 22))
		n := r.between(4, 6)
		longAt := -1
		if i == bad {
			longAt = r.intn(n + 1)
		}
		t := start
		var ticks []float64
		for k := 0; k <= n; k++ {
			gap := float64(r.between(30, 55)) / 60
			if k == longAt {
				gap = float64(r.between(75, 100)) / 60
			}
			t += gap
			if k < n {
				ticks = append(ticks, t)
			}
		}
		rows = append(rows, tlRow{name: names[i], bars: []tlBar{{start, t, "#cbd5e1", "", 0}}, ticks: ticks, tickC: tlBlue})
	}
	tl := timeline{title: "LONE WORKER APP — NIGHT SHIFT CHECK-INS", from: 17, to: 29, step: 1, axis: hourAxis, rows: rows,
		legend: [][2]string{{"#cbd5e1", "Shift"}, {tlBlue, "▼ Check-in"}}, note: "Shift start and shift end count as check-ins."}
	q := Question{key: "ex/loneworker", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Lone workers must check in at least every 60 minutes. Which worker went more than an hour without checking in?",
		"Refer to %[1]s. Which worker's check-ins would have triggered a welfare call under the 60-minute rule?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── hot work fire watch (NFPA 51B: 60 minutes after) ────────────────────

func fireWatchExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	jobs := sample(r, []string{"Roof drain weld", "Dock rail repair", "Duct hanger cut", "Gate hinge weld", "Pipe sweat joint", "Stair nosing grind", "Sign post cut", "Boiler flange weld"}, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i, j := range jobs {
		a := float64(r.between(32, 52)) / 4
		b := a + float64(r.between(4, 10))/4
		after := float64(r.between(60, 90)) / 60
		if i == bad {
			after = float64(r.between(15, 40)) / 60
		}
		rows = append(rows, tlRow{name: j, bars: []tlBar{
			{a, b, tlRed, "Hot work", 1},
			{a, b + after, tlAmber, "Fire watch", 2},
		}})
	}
	tl := timeline{title: "HOT WORK PERMITS — TODAY", from: 7, to: 18, step: 1, axis: hourAxis, rows: rows,
		legend: [][2]string{{tlRed, "Hot work"}, {tlAmber, "Fire watch on duty"}}}
	q := Question{key: "ex/firewatch", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. NFPA 51B requires a fire watch during hot work and for at least 60 minutes after it ends. Which job ended its fire watch too early?",
		"Refer to %[1]s. Which hot work permit was closed out without a long enough fire watch?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, jobs, bad)
	return q
}

// ── confined space attendant ────────────────────────────────────────────

func attendantExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 3)
	start := float64(r.between(32, 40)) / 4
	var rows []tlRow
	lo, hi := 99.0, 0.0
	for i := 0; i < 2; i++ {
		a := start + float64(r.between(0, 4))/4
		b := a + float64(r.between(8, 14))/4
		lo, hi = math.Min(lo, a), math.Max(hi, b)
		rows = append(rows, tlRow{name: "Entrant " + names[i], bars: []tlBar{{a, b, tlBlue, "In space", 0}}})
	}
	// Attendant leaves for 20–30 minutes while an entrant is inside.
	e := rows[r.intn(2)].bars[0]
	g1 := e.a + 0.5 + float64(r.between(0, int((e.b-e.a-1.5)*4)))/4
	g2 := g1 + float64(r.between(20, 30))/60
	rows = append(rows, tlRow{name: "Attendant " + names[2], bars: []tlBar{{lo - 0.25, g1, tlGreen, "On duty", 0}, {g2, hi + 0.25, tlGreen, "On duty", 0}}})
	right := hmRange(g1, g2)
	var wrongs []string
	for _, w := range []float64{lo - 0.25 - 0.5, g2 + 0.5, g1 - 1, hi + 0.5, lo + 0.25} {
		if (w+0.4 <= g1 || w >= g2) && len(wrongs) < 3 {
			if s := hmRange(w, w+0.4); s != right && !contains(wrongs, s) {
				wrongs = append(wrongs, s)
			}
		}
	}
	for k := 1; len(wrongs) < 3; k++ {
		wrongs = append(wrongs, hmRange(hi+float64(k), hi+float64(k)+0.4))
	}
	tl := timeline{title: "PERMIT-REQUIRED CONFINED SPACE — VAULT 3 ENTRY LOG", from: math.Floor(lo - 1), to: math.Ceil(hi + 1), step: 0.5, axis: hmBare, rows: rows,
		legend: [][2]string{{tlBlue, "Entrant inside"}, {tlGreen, "Attendant at the entry"}}, nameWidth: 140}
	q := Question{key: "ex/attendant", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. OSHA requires an attendant outside the space for as long as anyone is inside. During which period was that rule broken?",
		"Refer to %[1]s. When was an entrant inside the permit space with no attendant on duty?",
	}), ex)
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── visitor escort ──────────────────────────────────────────────────────

func escortLogExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i := range names {
		a := float64(r.between(32, 52)) / 4
		b := a + float64(r.between(4, 12))/4
		bars := []tlBar{{a, b, "#cbd5e1", "On site", 1}}
		if i == bad {
			g := a + 0.25 + float64(r.between(0, int((b-a-1)*4)))/4
			bars = append(bars, tlBar{a, g, tlGreen, "Escorted", 2}, tlBar{g + float64(r.between(25, 45))/60, b, tlGreen, "", 2})
		} else {
			bars = append(bars, tlBar{a, b, tlGreen, "Escorted", 2})
		}
		rows = append(rows, tlRow{name: names[i] + " (visitor)", bars: bars})
	}
	tl := timeline{title: "VISITOR MANAGEMENT — BADGE AND ESCORT TIMES", from: 7, to: 19, step: 1, axis: hourAxis, rows: rows,
		legend: [][2]string{{"#cbd5e1", "Visitor badge active"}, {tlGreen, "Escort with visitor"}}, nameWidth: 132}
	q := Question{key: "ex/escortlog", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Visitors must be escorted for their entire time on site. Which visitor was unescorted for part of the visit?",
		"Refer to %[1]s. Which visit should be reported to Security as a lapse in escort?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, rows4Names(rows), bad)
	return q
}

// ── change windows ──────────────────────────────────────────────────────

func changeWindowExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ids := make([]string, 4)
	for i := range ids {
		ids[i] = fmt.Sprintf("CHG-%05d", r.between(10000, 99999))
	}
	bad := r.intn(4)
	var rows []tlRow
	for i := range ids {
		a := float64(r.between(88, 96)) / 4
		b := a + float64(r.between(2, 8))/4
		if b > 26 {
			a, b = a-(b-26), 26
		}
		if i == bad {
			if r.chance(0.5) {
				a = float64(r.between(80, 86)) / 4
				b = a + float64(r.between(4, 8))/4
			} else {
				b = float64(r.between(106, 112)) / 4
				a = b - float64(r.between(4, 8))/4
			}
		}
		rows = append(rows, tlRow{name: ids[i], bars: []tlBar{{a, b, tlBlue, "", 0}}})
	}
	tl := timeline{title: "SATURDAY NIGHT CHANGE CALENDAR", from: 18, to: 30, step: 1, axis: hourAxis, rows: rows,
		shade: [][2]float64{{22, 26}}, shadeLabel: "APPROVED MAINTENANCE WINDOW", legend: [][2]string{{tlBlue, "Change executed"}}}
	q := Question{key: "ex/changewindow", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which change ran outside the approved maintenance window and needs a post-implementation review?",
		"Refer to %[1]s. Approved changes must start and finish inside the window. Which change did not?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, ids, bad)
	return q
}

// ── trading blackout ────────────────────────────────────────────────────

func blackoutExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	b0, b1 := 14.0, float64(r.between(52, 56)) // Mar 15 – Apr 22..26 (day index from Mar 1)
	day := func(v float64) string {
		d := int(math.Round(v))
		if d < 31 {
			return fmt.Sprintf("Mar %d", d+1)
		}
		return fmt.Sprintf("Apr %d", d-30)
	}
	var rows []tlRow
	for i := range names {
		var d float64
		if i == bad {
			d = float64(r.between(int(b0)+2, int(b1)-2))
		} else if r.chance(0.5) {
			d = float64(r.between(1, int(b0)-2))
		} else {
			d = float64(r.between(int(b1)+1, 59))
		}
		rows = append(rows, tlRow{name: names[i], ticks: []float64{d}, tickC: tlRed})
	}
	tl := timeline{title: "Q1 QUARTERLY BLACKOUT — PRE-CLEARED TRADES", from: 0, to: 60, step: 6, axis: day, rows: rows,
		shade: [][2]float64{{b0, b1}}, shadeLabel: "BLACKOUT PERIOD", shadeColor: "#fecaca", labelTicks: true, tickFormat: day,
		legend: [][2]string{{tlRed, "▼ Trade executed"}}, note: "The window reopens after one full trading day following the earnings release."}
	q := Question{key: "ex/blackout", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which insider traded during the blackout period?",
		"Refer to %[1]s. Which trade should the General Counsel's office investigate?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── PERCH sessions double-booked ────────────────────────────────────────

func doubleBookedExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	codes := []string{"K1", "B2", "W3", "P4", "R5", "T6", "L7", "M8", "S9"}
	var rows []tlRow
	for i := range names {
		t := 9.0
		var bars []tlBar
		for k := 0; k < 3; k++ {
			t += float64(r.between(0, 2)) / 2
			d := float64(r.between(2, 3)) / 2
			bars = append(bars, tlBar{t, t + d, tlBlue, pick(r, codes), 0})
			t += d
		}
		if i == bad {
			k := r.intn(2)
			bars[k].half = 1
			na := bars[k].a + 0.25
			bars = append(bars, tlBar{na, na + 1, "#7c3aed", pick(r, codes), 2})
		}
		rows = append(rows, tlRow{name: names[i], bars: bars})
	}
	tl := timeline{title: "PERCH CONFERENCE — REGISTERED BREAKOUT SESSIONS", from: 8, to: 16, step: 1, axis: hourAxis, rows: rows,
		note: "Session codes from the PERCH program. Badges are scanned at each door; you can attend one room at a time."}
	q := Question{key: "ex/doublebooked", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which attendee is registered for two sessions that overlap?",
		"Refer to %[1]s. Whose PERCH schedule will produce a badge scan conflict?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── focus time ──────────────────────────────────────────────────────────

func focusBlockExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	bad := r.intn(5)
	var rows []tlRow
	for i, d := range days {
		for {
			var bars []tlBar
			t := 9.0
			for t < 17 {
				gap := float64(r.between(0, 4)) / 2
				if i == bad {
					gap = float64(r.between(0, 3)) / 2
				}
				t += gap
				if t >= 17 {
					break
				}
				m := math.Min(float64(r.between(1, 3))/2, 17-t)
				bars = append(bars, tlBar{t, t + m, tlBlue, "", 0})
				t += m
			}
			free, prev := 0.0, 9.0
			for _, b := range bars {
				free = math.Max(free, b.a-prev)
				prev = b.b
			}
			free = math.Max(free, 17-prev)
			if (free >= 2) != (i == bad) {
				rows = append(rows, tlRow{name: d, bars: bars})
				break
			}
		}
	}
	tl := timeline{title: "YOUR CALENDAR — NEXT WEEK", from: 9, to: 17, step: 1, axis: hourAxis, rows: rows, legend: [][2]string{{tlBlue, "Busy"}}}
	q := Question{key: "ex/focusblock", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. On which day can you NOT schedule an uninterrupted 2-hour focus block between 9 a.m. and 5 p.m.?",
		"Refer to %[1]s. The Focus Time policy asks for one 2-hour block per day. Which day has no room for it?",
	}), ex)
	var wrongs []string
	for i, d := range days {
		if i != bad && len(wrongs) < 3 {
			wrongs = append(wrongs, d)
		}
	}
	q.Choices, q.Correct = withChoices(r, days[bad], sample(r, wrongs, 3))
	return q
}

func meetingSlotExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	slot := r.between(9, 16)
	busy := make([][]bool, 4)
	for i := range busy {
		busy[i] = make([]bool, 24)
	}
	for h := 9; h < 17; h++ {
		if h != slot {
			busy[r.intn(4)][h] = true
		}
	}
	for i := range busy {
		for h := 9; h < 17; h++ {
			if h != slot && r.chance(0.3) {
				busy[i][h] = true
			}
		}
	}
	var rows []tlRow
	for i := range names {
		var bars []tlBar
		for h := 9; h < 17; h++ {
			if busy[i][h] {
				bars = append(bars, tlBar{float64(h), float64(h) + 1, tlBlue, "Busy", 0})
			}
		}
		rows = append(rows, tlRow{name: names[i], bars: bars})
	}
	tl := timeline{title: "SCHEDULING ASSISTANT — THURSDAY", from: 9, to: 17, step: 1, axis: hourAxis, rows: rows}
	q := Question{key: "ex/meetingslot", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Which one-hour slot works for all four attendees?", ex)
	sh := func(h int) string { return clockShort(h) + "–" + clockShort(h+1) }
	var wrongs []string
	for _, h := range sample(r, []int{9, 10, 11, 12, 13, 14, 15, 16}, 8) {
		if h != slot && len(wrongs) < 3 {
			wrongs = append(wrongs, sh(h))
		}
	}
	q.Choices, q.Correct = withChoices(r, sh(slot), wrongs)
	return q
}

func coreHoursExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	cities := []string{"Milwaukee", "Denver", "Phoenix", "Boston", "Portland", "Atlanta", "Seattle", "Austin"}
	var rows []tlRow
	for i := range names {
		a := float64(r.between(28, 40)) / 4
		b := a + 8 + float64(r.between(0, 4))/4
		if i == bad {
			if r.chance(0.5) {
				a = float64(r.between(42, 46)) / 4
				b = a + 8
			} else {
				a = float64(r.between(24, 26)) / 4
				b = float64(r.between(54, 58)) / 4
			}
		}
		rows = append(rows, tlRow{name: names[i] + " · " + pick(r, cities), bars: []tlBar{{a, b, tlBlue, hmRange(a, b), 0}}})
	}
	tl := timeline{title: "TEAM WORKING HOURS (SHOWN IN CENTRAL TIME)", from: 6, to: 20, step: 1, axis: hourAxis, rows: rows,
		shade: [][2]float64{{10, 15}}, shadeLabel: "CORE HOURS 10 A.M.–3 P.M. CT", nameWidth: 150}
	q := Question{key: "ex/corehours", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. The hybrid work policy requires everyone to be available during core hours. Whose working hours do not cover them?", ex)
	q.Choices, q.Correct = whoQ(r, rows4Names(rows), bad)
	return q
}

// ── fatigue policy: 16 h on, 8 h off ────────────────────────────────────

func fatigueExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	long := r.chance(0.5)
	day := func(v float64) string {
		switch h := int(math.Round(v)); {
		case h == 0:
			return "Mon"
		case h == 24:
			return "Tue"
		case h == 48:
			return "Wed"
		case h%24 < 12:
			return fmt.Sprintf("%da", h%24)
		case h%24 == 12:
			return "12p"
		default:
			return fmt.Sprintf("%dp", h%24-12)
		}
	}
	var rows []tlRow
	for i := range names {
		a := float64(r.between(5, 9))
		b := a + float64(r.between(8, 14))
		rest := float64(r.between(9, 13))
		c2 := b + rest
		d2 := math.Min(c2+float64(r.between(6, 10)), 47)
		if i == bad {
			if long {
				b = a + float64(r.between(17, 19))
				c2 = b + float64(r.between(9, 11))
				d2 = math.Min(c2+6, 47)
			} else {
				c2 = b + float64(r.between(4, 6))
				d2 = math.Min(c2+8, 47)
			}
		}
		rows = append(rows, tlRow{name: names[i], bars: []tlBar{{a, b, tlBlue, fmt.Sprintf("%gh", b-a), 0}, {c2, d2, tlBlue, fmt.Sprintf("%gh", d2-c2), 0}}})
	}
	tl := timeline{title: "MAINTENANCE CREW — DUTY HOURS", from: 0, to: 48, step: 4, axis: day, rows: rows,
		legend: [][2]string{{tlBlue, "On duty"}}, note: "Fatigue policy: no more than 16 consecutive hours on duty, and at least 8 hours off between shifts."}
	q := Question{key: "ex/fatigue", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Which crew member's schedule violates the fatigue policy?", ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── shuttle departures ──────────────────────────────────────────────────

func shuttleExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	lines := []string{"Red Line", "Blue Line", "Green Line"}
	var rows []tlRow
	for _, l := range lines {
		var ticks []float64
		t := 16 + float64(r.between(0, 20))/60
		for t < 19 {
			ticks = append(ticks, t)
			t += float64(r.between(18, 40)) / 60
		}
		rows = append(rows, tlRow{name: l, ticks: ticks, tickC: tlBlue})
	}
	k := r.intn(3)
	tk := rows[k].ticks
	i := r.between(1, len(tk)-2)
	arrive := tk[i-1] + (tk[i]-tk[i-1])*0.5
	arrive = math.Floor(arrive*60) / 60
	if arrive <= tk[i-1] {
		arrive = tk[i-1] + 1.0/60
	}
	tl := timeline{title: "CAMPUS SHUTTLE — WEEKDAY EVENING DEPARTURES", from: 16, to: 19, step: 0.5, axis: hm, rows: rows, note: "All times p.m.",
		labelTicks: true, tickFormat: func(v float64) string { return strings.TrimSuffix(strings.TrimSuffix(hm(v), " p.m."), " a.m.") }}
	q := Question{key: "ex/shuttle", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. If you reach the %s stop at %s, when is the next %s departure?", ex, lines[k], hm(arrive), lines[k])
	var wrongs []string
	for _, w := range []float64{tk[i-1], tk[i+1], tk[i] + 0.25, tk[i] - 1.0/12} {
		if s := hm(w); s != hm(tk[i]) && !contains(wrongs, s) && len(wrongs) < 3 {
			wrongs = append(wrongs, s)
		}
	}
	q.Choices, q.Correct = withChoices(r, hm(tk[i]), wrongs)
	return q
}

// ── badge-ins outside the scheduled shift ───────────────────────────────

func afterHoursExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i := range names {
		a := float64(r.between(6, 10))
		b := a + 8
		var tk []float64
		tk = append(tk, a-float64(r.between(5, 25))/60)
		tk = append(tk, a+float64(r.between(20, 44))/4)
		if i == bad {
			tk = append(tk, b+float64(r.between(8, 16))/4)
		}
		rows = append(rows, tlRow{name: names[i], bars: []tlBar{{a, b, "#cbd5e1", "Scheduled shift", 0}}, ticks: tk, tickC: tlRed})
	}
	tl := timeline{title: "BADGE READER LOG vs. SCHEDULED SHIFTS", from: 4, to: 24, step: 2, axis: hourAxis, rows: rows,
		legend: [][2]string{{"#cbd5e1", "Scheduled shift"}, {tlRed, "▼ Badge-in"}}, note: "Policy: building access outside your scheduled shift (beyond 30 minutes early) requires manager pre-approval."}
	q := Question{key: "ex/afterhours", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Whose badge activity needs an after-hours access approval on file?", ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

func screenBreakExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i := range names {
		t := 9.0
		var bars []tlBar
		long := i == bad
		for t < 13 {
			d := float64(r.between(30, 55)) / 60
			if long && t > 9.5 {
				d = float64(r.between(85, 110)) / 60
				long = false
			}
			d = math.Min(d, 13-t)
			bars = append(bars, tlBar{t, t + d, tlBlue, "", 0})
			t += d + 1.0/12
		}
		rows = append(rows, tlRow{name: names[i], bars: bars})
	}
	tl := timeline{title: "SCREEN TIME LOG — MORNING", from: 9, to: 13, step: 0.5, axis: hmBare, rows: rows,
		legend: [][2]string{{tlBlue, "Continuous screen work"}}, note: "Gaps between bars are 5-minute breaks."}
	q := Question{key: "ex/screenbreak", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. The ergonomics program asks for a short break at least every 60 minutes of continuous screen work. Who skipped one?", ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── on-call rotation gap ────────────────────────────────────────────────

func onCallExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	gap := r.between(2, 25) // 6-hour block index
	var rows []tlRow
	blocks := make([]int, 28)
	for b := range blocks {
		blocks[b] = (b / 7) % 4
	}
	for i := range names {
		var bars []tlBar
		for b := 0; b < 28; b++ {
			if blocks[b] == i && b != gap {
				bars = append(bars, tlBar{float64(b) * 6, float64(b+1) * 6, tlBlue, "", 0})
			}
		}
		rows = append(rows, tlRow{name: names[i], bars: bars})
	}
	ax := func(v float64) string { return days[int(v/24)%7] }
	tl := timeline{title: "PAGER ROTATION — PRIMARY ON CALL", from: 0, to: 168, step: 24, axis: ax, rows: rows,
		legend: [][2]string{{tlBlue, "On call"}}, note: "Each day is split into four 6-hour blocks starting at midnight."}
	blk := func(b int) string {
		h := (b % 4) * 6
		return fmt.Sprintf("%s %s–%s", days[b/4], clockShort(h), clockShort((h+6)%24))
	}
	q := Question{key: "ex/oncall", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. When does the rotation leave no one on call?", ex)
	var wrongs []string
	for _, b := range sample(r, []int{gap - 2, gap - 1, gap + 1, gap + 2, gap + 4}, 5) {
		if b >= 0 && b < 28 && len(wrongs) < 3 {
			wrongs = append(wrongs, blk(b))
		}
	}
	q.Choices, q.Correct = withChoices(r, blk(gap), wrongs)
	return q
}

// ── time as a public health control (4 hours) ───────────────────────────

func tphcExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	items := sample(r, []string{"Sushi tray", "Deviled eggs", "Chicken salad", "Cut melon", "Pasta salad", "Sliders", "Spring rolls", "Quiche bites"}, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i, it := range items {
		a := float64(r.between(42, 50)) / 4
		d := float64(r.between(8, 15)) / 4
		if i == bad {
			d = float64(r.between(18, 22)) / 4
		}
		rows = append(rows, tlRow{name: it, bars: []tlBar{{a, a + d, tlAmber, "", 0}}, ticks: []float64{a + d}, tickC: tlRed})
	}
	tl := timeline{title: "CATERED LUNCH — TIME AS A PUBLIC HEALTH CONTROL LOG", from: 10, to: 18, step: 1, axis: hourAxis, rows: rows,
		legend: [][2]string{{tlAmber, "Out of temperature control"}, {tlRed, "▼ Discarded"}}}
	q := Question{key: "ex/tphc", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Under the FDA Food Code's time-as-a-control rule, food held without temperature control must be discarded within 4 hours. Which item was kept out too long?",
		"Refer to %[1]s. Which item should have been thrown away earlier?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, items, bad)
	return q
}

// ── OSHA noise dose ─────────────────────────────────────────────────────

func noiseDoseExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 4)
	bad := r.intn(4)
	levels := []int{85, 90, 95, 100}
	cols := map[int]string{85: "#86efac", 90: "#fde047", 95: "#fb923c", 100: "#ef4444"}
	var rows []tlRow
	for i := range names {
		for {
			t := 7.0
			dose := 0.0
			var bars []tlBar
			for t < 15 {
				d := math.Min(float64(r.between(2, 6))/2, 15-t)
				l := pick(r, levels)
				if l >= 90 {
					dose += d / (8 / math.Pow(2, float64(l-90)/5))
				}
				bars = append(bars, tlBar{t, t + d, cols[l], fmt.Sprintf("%d dBA", l), 0})
				t += d
			}
			if (i == bad && dose > 1.15) || (i != bad && dose < 0.85) {
				rows = append(rows, tlRow{name: names[i], bars: bars})
				break
			}
		}
	}
	tl := timeline{title: "NOISE EXPOSURE BY TASK — 8-HOUR SHIFT", from: 7, to: 15, step: 1, axis: hourAxis, rows: rows,
		note: "OSHA PEL: 90 dBA for 8 hours, halving for every 5 dBA increase; levels below 90 dBA do not count toward the PEL."}
	q := Question{key: "ex/noisedose", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Which worker's daily noise dose exceeds OSHA's permissible exposure limit?", ex)
	q.Choices, q.Correct = whoQ(r, names, bad)
	return q
}

// ── DOT post-accident testing windows ───────────────────────────────────

func postTestExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ids := make([]string, 4)
	for i := range ids {
		ids[i] = fmt.Sprintf("Accident %d-%03d", r.between(24, 26), r.between(1, 999))
	}
	bad := r.intn(4)
	alcLate := r.chance(0.5)
	var rows []tlRow
	for i := range ids {
		alc := float64(r.between(3, 7)) / 4 * 1.5
		drug := float64(r.between(6, 24))
		if i == bad {
			if alcLate {
				alc = float64(r.between(34, 44)) / 4
			} else {
				drug = float64(r.between(33, 35))
			}
		}
		rows = append(rows, tlRow{name: ids[i], bars: []tlBar{{alc - 0.2, alc + 0.2, tlBlue, "", 1}, {drug - 0.2, drug + 0.2, "#7c3aed", "", 2}},
			ticks: []float64{alc, drug}, tickC: "#334155"})
	}
	tl := timeline{title: "POST-ACCIDENT TESTING — HOURS AFTER THE ACCIDENT", from: 0, to: 36, step: 4, axis: func(v float64) string { return fmt.Sprintf("%gh", v) }, rows: rows,
		labelTicks: true, tickFormat: func(v float64) string { return fmt.Sprintf("%.1fh", v) }, nameWidth: 128,
		legend: [][2]string{{tlBlue, "Alcohol test"}, {"#7c3aed", "Drug test"}}, note: "Hour 0 is the time of the accident."}
	q := Question{key: "ex/posttest", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. DOT rules require a post-accident alcohol test within 8 hours and a drug test within 32 hours. Which accident's testing missed a deadline?",
		"Refer to %[1]s. Which accident needs a record explaining why a required DOT test was not completed in time?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, ids, bad)
	return q
}

func dockHoursExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	carriers := sample(r, []string{"Badger Freight", "Lakeshore LTL", "Cream City Courier", "Northwoods Logistics", "Brew City Express", "Great Lakes Haulers", "Kettle Moraine Transport"}, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i, cr := range carriers {
		var a float64
		if i == bad {
			a = pick(r, []float64{6, 11.5, 12, 15.5, 16})
		} else if r.chance(0.5) {
			a = float64(r.between(28, 41)) / 4
		} else {
			a = float64(r.between(52, 57)) / 4
		}
		rows = append(rows, tlRow{name: cr, ticks: []float64{a}, tickC: tlBlue})
	}
	tl := timeline{title: "RECEIVING DOCK — SCHEDULED DELIVERIES", from: 5, to: 18, step: 1, axis: hourAxis, rows: rows, nameWidth: 140,
		shade: [][2]float64{{7, 11}, {13, 15.25}}, shadeLabel: "DOCK OPEN", labelTicks: true, tickFormat: hm}
	q := Question{key: "ex/dockhours", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Which delivery is booked when the receiving dock is closed?", ex)
	q.Choices, q.Correct = whoQ(r, carriers, bad)
	return q
}

// ── lockout across a shift change (1910.147(f)(4)) ──────────────────────

func lockHandoffExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	machines := sample(r, []string{"Press P-4", "Conveyor C-2", "Baler B-1", "Mixer M-7", "Shredder S-3", "Compressor K-9", "Sorter X-5"}, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i, m := range machines {
		off := 15 + float64(r.between(-2, 2))/4
		on := off - float64(r.between(1, 3))/4
		if i == bad {
			on = off + float64(r.between(1, 3))/4
		}
		rows = append(rows, tlRow{name: m, bars: []tlBar{{11 + float64(r.between(0, 4))/4, off, tlRed, "1st shift lock", 1}, {on, 19, "#7c3aed", "2nd shift lock", 2}}})
	}
	tl := timeline{title: "LOCKOUT/TAGOUT — SHIFT CHANGE HANDOFF", from: 10, to: 20, step: 1, axis: hourAxis, rows: rows,
		legend: [][2]string{{tlRed, "Off-going worker's lock"}, {"#7c3aed", "Oncoming worker's lock"}}}
	q := Question{key: "ex/lockhandoff", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. OSHA requires lockout protection to continue through a shift change. On which machine was the protection interrupted?",
		"Refer to %[1]s. Which machine was left without any lock applied at shift change?",
	}), ex)
	q.Choices, q.Correct = whoQ(r, machines, bad)
	return q
}

func backupWindowExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	servers := sample(r, []string{"FILE-01", "SQL-PAYROLL", "MAIL-02", "ERP-PRD", "HR-DOCS", "WIKI-01", "BUILD-03", "VAULT-07"}, 4)
	bad := r.intn(4)
	var rows []tlRow
	for i, s := range servers {
		a := 24 + float64(r.between(0, 8))/4
		d := float64(r.between(6, 14)) / 4
		if i == bad {
			d = 30 - a + float64(r.between(2, 6))/4
		}
		rows = append(rows, tlRow{name: s, bars: []tlBar{{a, a + d, tlBlue, fmt.Sprintf("%.4gh", d), 0}}})
	}
	tl := timeline{title: "NIGHTLY BACKUP JOBS", from: 22, to: 32, step: 1, axis: hourAxis, rows: rows,
		shade: [][2]float64{{22, 30}}, shadeLabel: "BACKUP WINDOW — MUST FINISH BY 6:00 A.M."}
	q := Question{key: "ex/backupwindow", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. Which server's backup overran the window and will slow the morning start?", ex)
	q.Choices, q.Correct = whoQ(r, servers, bad)
	return q
}

func coverageExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := names4(r, 5)
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	full := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	bad := r.intn(5)
	out := make([][]bool, 5)
	for i := range out {
		out[i] = make([]bool, 5)
	}
	for d := 0; d < 5; d++ {
		n := r.between(0, 2)
		if d == bad {
			n = r.between(4, 5)
		}
		for _, p := range pickN(r, 5, n) {
			out[p][d] = true
		}
	}
	var rows []tlRow
	for p := range names {
		var bars []tlBar
		for d := 0; d < 5; d++ {
			if out[p][d] {
				bars = append(bars, tlBar{float64(d) + 0.05, float64(d) + 0.95, tlAmber, "PTO", 0})
			}
		}
		rows = append(rows, tlRow{name: names[p], bars: bars})
	}
	tl := timeline{title: "TEAM PTO CALENDAR — WEEK OF THE 14TH", from: 0, to: 5, step: 1, centered: true, axis: func(v float64) string { return days[int(v)] }, rows: rows, legend: [][2]string{{tlAmber, "Approved PTO"}}, note: "Minimum staffing: at least two of the five team members working each weekday."}
	q := Question{key: "ex/coverage", Exhibit: tl.svg()}
	q.Text = fmt.Sprintf("Refer to %s. On which day does the team fall below minimum staffing?", ex)
	var wrongs []string
	for d := range full {
		if d != bad {
			wrongs = append(wrongs, full[d])
		}
	}
	q.Choices, q.Correct = withChoices(r, full[bad], sample(r, wrongs, 3))
	return q
}
