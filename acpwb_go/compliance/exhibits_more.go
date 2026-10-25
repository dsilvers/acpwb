package compliance

// More exhibits: ladder setup (4-to-1 rule), fire extinguisher label and
// gauge, heat index readout (NWS formula), team chat channel, non-exempt
// timesheet (FLSA overtime), layoff notice timeline (WARN 60 days),
// interview guide (EEOC), and a video call grid. The life-safety and
// employment-law answers are computed from real rules; keep them accurate.
//
// This file sorts after exhibits.go, so its init runs last and its
// category mappings append to the ones set by the other exhibit files.

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func init() {
	for k, g := range map[string]exhibitGen{
		"ladder":       ladderExhibit,
		"extinguisher": extinguisherExhibit,
		"heatindex":    heatIndexExhibit,
		"chatthread":   chatThreadExhibit,
		"timesheet":    timesheetExhibit,
		"warn":         warnExhibit,
		"interview":    interviewExhibit,
		"videogrid":    videoGridExhibit,
	} {
		exhibitGens[k] = g
		generatedKinds["ex-"+k] = true
	}
	exhibitNames = exhibitNames[:0]
	for k := range exhibitGens {
		exhibitNames = append(exhibitNames, k)
	}
	sortStrings(exhibitNames)
	for cat, ks := range map[string][]string{
		"fire":     {"extinguisher", "evac"},
		"heat":     {"heatindex"},
		"ladder":   {"ladder"},
		"wage":     {"timesheet"},
		"layoff":   {"warn"},
		"hiring":   {"interview"},
		"chat":     {"chatthread"},
		"video":    {"videogrid"},
		"social":   {"chatthread"},
		"safety":   {"extinguisher", "ladder"},
		"meetings": {"videogrid"},
		"remote":   {"videogrid", "chatthread"},
		"weather":  {"heatindex"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// ── ladder (4-to-1 rule) ────────────────────────────────────────────────

func ladderExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	h := 4 * r.between(2, 7) // feet to the support point: 8–28
	right := h / 4
	const groundY, wallX = 236.0, 300.0
	ppf := 180 / float64(h) // the ladder always fills the frame
	var base float64        // feet out from the wall, as drawn
	setup := r.intn(3)      // 0 correct, 1 too close, 2 too far
	switch setup {
	case 0:
		base = float64(right)
	case 1:
		base = math.Max(1, float64(right)/2.5)
	default:
		base = float64(right) * 2
	}
	base = math.Round(base*2) / 2
	if setup != 0 && base == float64(right) {
		base += 1.5
	}
	topY := groundY - float64(h)*ppf
	bx := wallX - base*ppf
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="cmp-exhibit-img" width="380" height="270" viewBox="0 0 380 270" role="img" aria-label="Ladder against a wall"><rect width="380" height="270" fill="#eef3f7"/>`)
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="30" height="%.0f" fill="#b0855b"/><rect x="0" y="%.0f" width="380" height="34" fill="#8d9a6b"/>`, wallX, topY-14, groundY-topY+14, groundY)
	fmt.Fprintf(&b, `<path d="M%.0f %.0f h40 l-6 10 h-34 Z" fill="#6d4c33"/><rect x="%.0f" y="%.0f" width="36" height="6" rx="3" fill="#90a4ae"/>`, wallX-6, topY-24, wallX-10, topY-4)
	for y := topY - 4; y < groundY; y += 12 { // brick courses
		fmt.Fprintf(&b, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#9a6e48" stroke-width="1"/>`, wallX, y, wallX+30, y)
	}
	// rails: the ladder leans from (bx, ground) to (wallX, topY)
	dx, dy := wallX-bx, topY-groundY
	l := math.Hypot(dx, dy)
	nx, ny := -dy/l*7, dx/l*7
	for _, s := range []float64{-1, 1} {
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#c9a227" stroke-width="4" stroke-linecap="round"/>`, bx+s*nx, groundY+s*ny, wallX+s*nx, topY+s*ny)
	}
	for t := 0.08; t < 0.98; t += 0.09 {
		x, y := bx+dx*t, groundY+dy*t
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#c9a227" stroke-width="3"/>`, x-nx, y-ny, x+nx, y+ny)
	}
	// dimensions
	fmt.Fprintf(&b, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#1a3a5c" stroke-width="1.2" stroke-dasharray="4 3"/>`, wallX+44, topY, wallX+44, groundY)
	fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" %s font-size="11" font-weight="700" fill="#1a3a5c" transform="rotate(-90 %.0f %.0f)" text-anchor="middle">%d ft to support point</text>`, wallX+40, (topY+groundY)/2, svgFont, wallX+40, (topY+groundY)/2, h)
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#1a3a5c" stroke-width="1.2" stroke-dasharray="4 3"/>`, bx, groundY+16, wallX, groundY+16)
	fmt.Fprintf(&b, `<text x="%.1f" y="%.0f" %s font-size="11" font-weight="700" fill="#fff" text-anchor="middle">%s ft</text>`, (bx+wallX)/2, groundY+30, svgFont, strings.TrimSuffix(fmt.Sprintf("%.1f", base), ".0"))
	fmt.Fprintf(&b, `<text x="12" y="22" %s font-size="11" font-weight="700" fill="#1a3a5c">Extension ladder · %s</text>`, svgFont, esc(sl["PLACE"]))
	b.WriteString(`</svg>`)
	q := Question{key: "ex/ladder", Exhibit: b.String()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. %s set up this ladder to reach a gutter. Using the 4-to-1 rule, is the setup correct?", ex, sl["NAME"])
		opts := []string{
			"Yes; the base is about 1 foot out for every 4 feet of height",
			"No; the base is too close to the wall, so the ladder could tip backward",
			"No; the base is too far from the wall, so the feet could slide out",
		}
		q.Choices, q.Correct = withChoices(r, opts[setup], append(append([]string{}, opts[:setup]...), append(opts[setup+1:], "No; extension ladders must stand straight up against the wall")...))
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. The ladder reaches a support point %d feet up. Following the 4-to-1 rule, how far from the wall should its base be?", ex, h)
	q.Choices, q.Correct = numChoices(r, right, []int{-1, 1, 2, -2, h - right, 3}, func(n int) string { return plural(n, "foot") })
	for i, ch := range q.Choices {
		q.Choices[i] = strings.Replace(ch, "foots", "feet", 1)
	}
	return q
}

// ── fire extinguisher ───────────────────────────────────────────────────

var fireClasses = []struct{ Letter, Fire string }{
	{"A", "a wastebasket of paper burning in the copy room"},
	{"B", "a small pool of spilled solvent burning in the maintenance shop"},
	{"C", "a smoking, still-plugged-in power strip under a desk"},
	{"K", "a deep fryer full of cooking oil in the cafeteria kitchen"},
}

type extType struct {
	Name    string
	Classes string
	Color   string
}

var extinguisherTypes = []extType{
	{"WATER", "A", "#c62828"},
	{"CO2", "BC", "#37474f"},
	{"ABC DRY CHEMICAL", "ABC", "#c62828"},
	{"WET CHEMICAL", "AK", "#b0bec5"},
	{"BC DRY CHEMICAL", "BC", "#c62828"},
}

func classIcon(b *strings.Builder, letter string, x, y float64) {
	const s = 26.0
	cx, cy := x+s/2, y+s/2
	switch letter {
	case "A":
		fmt.Fprintf(b, `<polygon points="%.0f,%.0f %.0f,%.0f %.0f,%.0f" fill="#2e7d32"/>`, cx, y, x, y+s, x+s, y+s)
	case "B":
		fmt.Fprintf(b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#c62828"/>`, x+1, y+1, s-2, s-2)
	case "C":
		fmt.Fprintf(b, `<circle cx="%.0f" cy="%.0f" r="%.0f" fill="#1565c0"/>`, cx, cy, s/2)
	case "K":
		var pts []string
		for i := 0; i < 6; i++ {
			a := math.Pi/6 + float64(i)*math.Pi/3
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", cx+s/2*math.Cos(a), cy+s/2*math.Sin(a)))
		}
		fmt.Fprintf(b, `<polygon points="%s" fill="#212121"/>`, strings.Join(pts, " "))
	}
	fmt.Fprintf(b, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="13" font-weight="700" fill="#fff">%s</text>`, cx, cy+(map[string]float64{"A": 4}[letter])+5, svgFont, letter)
}

func extinguisherExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	t := pick(r, extinguisherTypes)
	gauge := r.intn(4) // 0,1 green; 2 red low; 3 red high
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="340" height="300" viewBox="0 0 340 300" role="img" aria-label="Fire extinguisher"><rect width="340" height="300" fill="#f4f1ea"/>`)
	// body
	fmt.Fprintf(&b, `<rect x="40" y="90" width="90" height="190" rx="30" fill="%s" stroke="#222" stroke-width="1.5"/>`, t.Color)
	b.WriteString(`<rect x="70" y="56" width="30" height="40" fill="#555"/><path d="M60 58 h70 l10 -10" stroke="#333" stroke-width="7" fill="none" stroke-linecap="round"/><path d="M100 62 q50 10 40 80 q-6 30 -18 40" stroke="#222" stroke-width="6" fill="none"/><circle cx="62" cy="56" r="4" fill="#ffd54f"/>`)
	// gauge
	b.WriteString(`<circle cx="85" cy="40" r="22" fill="#fff" stroke="#222" stroke-width="2"/><path d="M68 47 A18 18 0 0 1 72 28" stroke="#c62828" stroke-width="5" fill="none"/><path d="M73 27 A18 18 0 0 1 97 27" stroke="#2e7d32" stroke-width="5" fill="none"/><path d="M98 28 A18 18 0 0 1 102 47" stroke="#c62828" stroke-width="5" fill="none"/>`)
	ang := map[int]float64{0: -90, 1: -78, 2: -165, 3: -15}[gauge]
	nx, ny := 85+16*math.Cos(ang*math.Pi/180), 40+16*math.Sin(ang*math.Pi/180)
	fmt.Fprintf(&b, `<line x1="85" y1="40" x2="%.1f" y2="%.1f" stroke="#111" stroke-width="2"/><circle cx="85" cy="40" r="2.5" fill="#111"/>`, nx, ny)
	// label
	b.WriteString(`<rect x="52" y="140" width="66" height="96" rx="4" fill="#fff" stroke="#999"/>`)
	for i, line := range wrapText(t.Name, 9) {
		fmt.Fprintf(&b, `<text x="85" y="%d" text-anchor="middle" %s font-size="9" font-weight="700" fill="#222">%s</text>`, 154+i*11, svgFont, esc(line))
	}
	// class panel to the right
	b.WriteString(`<rect x="170" y="70" width="150" height="190" rx="8" fill="#fff" stroke="#bbb"/>`)
	fmt.Fprintf(&b, `<text x="245" y="92" text-anchor="middle" %s font-size="11" font-weight="700" fill="#1a3a5c">LABEL (ENLARGED)</text>`, svgFont)
	y := 104.0
	for _, fc := range fireClasses {
		if !strings.Contains(t.Classes, fc.Letter) {
			continue
		}
		classIcon(&b, fc.Letter, 184, y)
		desc := map[string]string{"A": "Ordinary combustibles", "B": "Flammable liquids", "C": "Electrical equipment", "K": "Cooking oils and fats"}[fc.Letter]
		for i, ln := range wrapText(desc, 14) {
			fmt.Fprintf(&b, `<text x="218" y="%.0f" %s font-size="10" fill="#222">%s</text>`, y+11+float64(i)*11, svgFont, esc(ln))
		}
		y += 38
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/extinguisher", Exhibit: b.String()}
	if r.chance(0.35) {
		q.Text = fmt.Sprintf("Refer to %s. During the monthly inspection, %s checks this extinguisher's pressure gauge. What should happen?", ex, sl["FIRST"])
		ok := "Nothing; the needle is in the green, so it is charged and ready"
		bad := "Report it for service; the needle is out of the green zone"
		right, wrong := ok, bad
		if gauge >= 2 {
			right, wrong = bad, ok
		}
		q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Shake it until the needle moves", "Spray a short test burst in the hallway"})
		return q
	}
	var can, cannot []string
	for _, fc := range fireClasses {
		if strings.Contains(t.Classes, fc.Letter) {
			can = append(can, fc.Fire)
		} else {
			cannot = append(cannot, fc.Fire)
		}
	}
	if len(can) >= 3 || (len(cannot) == 1) {
		q.Text = fmt.Sprintf("Refer to %s. For which of these small fires should this extinguisher NOT be used?", ex)
		q.Choices, q.Correct = withChoices(r, capFirst(pick(r, cannot)), capAll(sample(r, can, 3)))
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. The alarm has been raised and the exit is clear behind you. Which small fire is this extinguisher rated for?", ex)
	q.Choices, q.Correct = withChoices(r, capFirst(pick(r, can)), capAll(sample(r, cannot, min(3, len(cannot)))))
	return q
}

// ── heat index (NWS) ────────────────────────────────────────────────────

// heatIndex is the National Weather Service heat index (°F) for air
// temperature t (°F) and relative humidity rh (%): the Rothfusz regression
// with NWS's adjustments, or the simple formula below 80°F.
func heatIndex(t, rh float64) float64 {
	simple := 0.5 * (t + 61 + (t-68)*1.2 + rh*0.094)
	if (simple+t)/2 < 80 {
		return simple
	}
	hi := -42.379 + 2.04901523*t + 10.14333127*rh - 0.22475541*t*rh - 0.00683783*t*t -
		0.05481717*rh*rh + 0.00122874*t*t*rh + 0.00085282*t*rh*rh - 0.00000199*t*t*rh*rh
	if rh < 13 && t >= 80 && t <= 112 {
		hi -= (13 - rh) / 4 * math.Sqrt((17-math.Abs(t-95))/17)
	} else if rh > 85 && t >= 80 && t <= 87 {
		hi += (rh - 85) / 10 * ((87 - t) / 5)
	}
	return hi
}

var heatCategories = []struct {
	Lo   float64
	Name string
}{
	{80, "Caution: fatigue is possible with prolonged exposure and activity"},
	{90, "Extreme Caution: heat cramps and heat exhaustion are possible"},
	{103, "Danger: heat cramps and heat exhaustion are likely, and heat stroke is possible"},
	{125, "Extreme Danger: heat stroke is highly likely"},
}

func heatCategory(hi float64) int {
	k := -1
	for i, c := range heatCategories {
		if hi >= c.Lo {
			k = i
		}
	}
	return k
}

func heatIndexExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	var t, rh, hi float64
	for try := 0; ; try++ {
		t = float64(r.between(84, 108))
		rh = float64(5 * r.between(4, 16))
		hi = heatIndex(t, rh)
		k := heatCategory(hi)
		if k < 0 || hi > 140 {
			continue
		}
		// keep clear of category edges so rounding cannot change the answer
		near := false
		for _, c := range heatCategories {
			if math.Abs(hi-c.Lo) < 1.5 {
				near = true
			}
		}
		if !near || try > 40 {
			break
		}
	}
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="300" height="190" viewBox="0 0 300 190" role="img" aria-label="Weather station readout"><rect width="300" height="190" rx="14" fill="#1f2a36"/>`)
	fmt.Fprintf(&b, `<text x="150" y="26" text-anchor="middle" %s font-size="11" font-weight="700" fill="#9fb3c8" letter-spacing="1">ACPWB LOADING DOCK · WEATHER STATION</text>`, svgFont)
	for i, f := range []struct{ label, val string }{{"AIR TEMP", fmt.Sprintf("%.0f°F", t)}, {"REL. HUMIDITY", fmt.Sprintf("%.0f%%", rh)}} {
		x := 20 + float64(i)*140
		fmt.Fprintf(&b, `<rect x="%.0f" y="42" width="120" height="96" rx="8" fill="#d7e8c3"/><text x="%.0f" y="62" text-anchor="middle" %s font-size="10" font-weight="700" fill="#33402a">%s</text><text x="%.0f" y="112" text-anchor="middle" font-family="Courier New,monospace" font-size="34" font-weight="700" fill="#1b2a10">%s</text>`,
			x, x+60, svgFont, f.label, x+60, f.val)
	}
	fmt.Fprintf(&b, `<text x="150" y="164" text-anchor="middle" %s font-size="9" fill="#9fb3c8">Shade reading · %s · wind calm</text>`, svgFont, clock(r.between(12, 16)*60+r.between(0, 59)))
	b.WriteString(`</svg>`)
	q := Question{key: "ex/heatindex", Exhibit: b.String()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Using the National Weather Service heat index, which risk category applies to %s's crew working outside right now?", ex, sl["FIRST"])
		k := heatCategory(hi)
		var wrongs []string
		for i, c := range heatCategories {
			if i != k {
				wrongs = append(wrongs, c.Name)
			}
		}
		q.Choices, q.Correct = withChoices(r, heatCategories[k].Name, wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. About what does it feel like outside, according to the National Weather Service heat index?", ex)
	right := int(math.Round(hi/5) * 5)
	q.Choices, q.Correct = numChoices(r, right, []int{-15, -10, 10, 15, 20, int(t) - right}, func(n int) string { return fmt.Sprintf("%d°F", n) })
	return q
}

// ── team chat channel ───────────────────────────────────────────────────

var chatBenign = []string{
	"Morning! Standup moved to 9:30 today.",
	"Draft is in the shared folder, comments welcome by Thursday.",
	"Thanks for covering the client call yesterday 🙏",
	"Reminder: the {COMMITTEE} deck is due Friday.",
	"Anyone know if the {PLACE} printer is fixed?",
	"Out at the dentist 2–3, back after.",
	"Great work on the quarterly numbers, team.",
	"Can someone review my PR before lunch?",
	"Heads up: {SYSTEM} will be down for maintenance at 6 p.m.",
	"I'll take the notes for the retro.",
	"Shared the agenda for Monday in the thread.",
	"Is the 3:00 still happening?",
}

var chatSensitive = []string{
	"{SYSTEM} admin password is Spring{YEAR}! if anyone needs it",
	"customer called about account 4417-2210-9983, SSN on file ends 6621, can someone help",
	"here's the unreleased Q3 earnings draft before the announcement: [attachment]",
	"pasting the full candidate list with home addresses and salaries below",
	"my badge PIN is 4471 if anyone needs to badge me in tomorrow",
}

var chatEveryone = []string{
	"@channel there's leftover cake in the {PLACE}",
	"@here who took my stapler",
	"@channel lunch?",
	"@here anyone else's coffee cold today",
}

func chatThreadExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := r.between(5, 7)
	type msg struct {
		who, text string
		at        int
	}
	names := sample(r, firstNames, n)
	benign := sample(r, chatBenign, n)
	msgs := make([]msg, n)
	t0 := r.between(8*60+30, 15*60)
	for i := range msgs {
		msgs[i] = msg{names[i] + " " + pick(r, lastNames)[:1] + ".", sl.fill(benign[i]), t0 + i*r.between(2, 9)}
		if i > 0 && msgs[i].at <= msgs[i-1].at {
			msgs[i].at = msgs[i-1].at + 1
		}
	}
	mode := r.intn(2) // 0 one sensitive message; 1 count everyone-pings
	bad := r.intn(n)
	pings := 0
	if mode == 0 {
		msgs[bad].text = sl.fill(pick(r, chatSensitive))
		if r.chance(0.5) { // a ping too, so it is not the only odd one
			o := (bad + 1 + r.intn(n-1)) % n
			msgs[o].text = sl.fill(pick(r, chatEveryone))
		}
	} else {
		pings = r.between(0, 3)
		for _, i := range sample(r, rangeInts(n), pings) {
			msgs[i].text = sl.fill(pick(r, chatEveryone))
		}
	}
	channel := pick(r, []string{"#general", "#ops-team", "#finance-close", "#records-mgmt", "#facilities", "#hr-partners", "#q3-planning", "#exec-assistants", "#compliance-help", "#random", "#perch-planning", "#data-stewards"})
	var b strings.Builder
	y := 52.0
	var body strings.Builder
	for _, m := range msgs {
		lines := wrapText(m.text, 46)
		fmt.Fprintf(&body, `<circle cx="34" cy="%.0f" r="12" fill="%s"/><text x="34" y="%.0f" text-anchor="middle" %s font-size="10" font-weight="700" fill="#fff">%s</text>`,
			y+10, pick(r, []string{"#5b8def", "#e0a030", "#4caf50", "#c48b9f", "#7e57c2", "#26a69a"}), y+14, svgFont, esc(m.who[:1]))
		fmt.Fprintf(&body, `<text x="54" y="%.0f" %s font-size="10" font-weight="700" fill="#1d1c1d">%s <tspan font-weight="400" fill="#888">%s</tspan></text>`, y+6, svgFont, esc(m.who), clock(m.at))
		for j, ln := range lines {
			fmt.Fprintf(&body, `<text x="54" y="%.0f" %s font-size="10" fill="#1d1c1d">%s</text>`, y+20+float64(j)*13, svgFont, esc(ln))
		}
		y += 22 + float64(len(lines))*13 + 8
	}
	fmt.Fprintf(&b, `<svg class="cmp-exhibit-img" width="380" height="%.0f" viewBox="0 0 380 %.0f" role="img" aria-label="Team chat channel"><rect width="380" height="%.0f" rx="8" fill="#fff" stroke="#ddd"/><rect width="380" height="34" rx="8" fill="#3f0e40"/><text x="14" y="22" %s font-size="12" font-weight="700" fill="#fff">%s</text><text x="366" y="22" text-anchor="end" %s font-size="9" fill="#cdb8ce">%d members</text>`,
		y+6, y+6, y+6, svgFont, esc(channel), svgFont, r.between(40, 900))
	b.WriteString(body.String())
	b.WriteString(`</svg>`)
	q := Question{key: "ex/chatthread", Exhibit: b.String()}
	if mode == 0 {
		q.Text = fmt.Sprintf("Refer to %s. Which message should be deleted right away and reported under the data handling policy?", ex)
		label := func(m msg) string { return fmt.Sprintf("The %s message from %s", clock(m.at), m.who) }
		var wrongs []string
		for i, m := range msgs {
			if i != bad {
				wrongs = append(wrongs, label(m))
			}
		}
		q.Choices, q.Correct = withChoices(r, label(msgs[bad]), sample(r, wrongs, 3))
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. How many of these messages notify everyone in the channel?", ex)
	q.Choices, q.Correct = numChoices(r, pings, []int{1, 2, 3, -1, 4}, func(n int) string { return plural(n, "message") })
	return q
}

func rangeInts(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// ── timesheet (FLSA overtime) ───────────────────────────────────────────

func timesheetExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var hrs [7]int // in quarter hours
	for d := 1; d <= 5; d++ {
		hrs[d] = 4*r.between(7, 10) + pick(r, []int{0, 0, 1, 2})
	}
	if r.chance(0.5) {
		hrs[6] = 4 * r.between(2, 6)
	}
	if r.chance(0.3) {
		hrs[0] = 4 * r.between(1, 4)
	}
	total, daily := 0, 0
	for _, h := range hrs {
		total += h
		if h > 32 {
			daily += h - 32
		}
	}
	ot := total - 160
	if ot < 0 {
		ot = 0
	}
	rate := 100 * r.between(18, 42) // cents
	start := exDate(r)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	q4 := func(n int) string {
		s := fmt.Sprintf("%.2f", float64(n)/4)
		return strings.TrimSuffix(strings.TrimSuffix(s, "0"), ".0")
	}
	var rows [][]string
	for d, h := range hrs {
		cell := q4(h)
		if h == 0 {
			cell = "—"
		}
		rows = append(rows, []string{days[d] + " " + start.AddDate(0, 0, d).Format("1/2"), cell})
	}
	rows = append(rows, []string{"Total", q4(total)})
	q := Question{key: "ex/timesheet", Exhibit: docExhibit("Weekly Timesheet — "+sl["NAME"],
		fmt.Sprintf("Non-exempt · %s per hour · workweek Sunday–Saturday", money(rate)), []string{"Day", "Hours"}, rows,
		"Federal FLSA rules apply. This location has no state daily-overtime rule.")}
	hours := func(n int) string { return q4(n) + " hours" }
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. How many overtime hours must %s be paid for this workweek?", ex, sl["FIRST"])
		var wrongs []string
		seen := map[int]bool{ot: true}
		for _, v := range []int{daily, ot + 8, ot + 4, ot - 4, 0, total - 128} {
			if v >= 0 && !seen[v] && len(wrongs) < 3 {
				seen[v] = true
				wrongs = append(wrongs, hours(v))
			}
		}
		q.Choices, q.Correct = withChoices(r, hours(ot), wrongs)
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. What are %s's gross wages for this workweek, including any overtime?", ex, sl["FIRST"])
		reg := total - ot
		right := rate*reg/4 + rate*3*ot/8
		if ot == 0 {
			q.Choices, q.Correct = numChoices(r, right, []int{-rate, rate, rate * 2, -rate * 2}, money)
			break
		}
		flat := rate * total / 4                       // overtime paid at straight time
		byDay := rate*(total-daily)/4 + rate*3*daily/8 // daily overtime, not federal
		double := rate*reg/4 + rate*2*ot/4             // double time
		var wrongs []string
		for _, v := range []int{flat, byDay, double, right + rate} {
			if v != right && !contains(wrongs, money(v)) && len(wrongs) < 3 {
				wrongs = append(wrongs, money(v))
			}
		}
		q.Choices, q.Correct = withChoices(r, money(right), wrongs)
	default:
		q.Text = fmt.Sprintf("Refer to %s. %s worked the Saturday hours without approval. Under the FLSA, those hours must be:", ex, sl["FIRST"])
		if hrs[6] == 0 {
			q.Text = fmt.Sprintf("Refer to %s. %s also answered emails at home for an hour on Wednesday night, not shown on the timesheet. That hour must be:", ex, sl["FIRST"])
		}
		q.Choices, q.Correct = withChoices(r, "Paid, counted toward the 40-hour overtime threshold; any policy issue is handled separately",
			[]string{"Unpaid, because they were not approved", "Paid as comp time next month instead of wages", "Paid at straight time only, never as overtime"})
	}
	return q
}

// ── layoff notice (WARN Act) ────────────────────────────────────────────

func warnExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	sep := exDate(r).AddDate(0, 0, r.between(30, 120))
	notice := sep.AddDate(0, 0, -pick(r, []int{14, 30, 45, 59, 60, 61, 75, 90}))
	days := int(sep.Sub(notice).Hours()/24 + 0.5)
	// A WARN "mass layoff": 50–499 positions that are at least a third of
	// the site, or 500+. Keep the numbers inside it so the Act applies.
	affected := r.between(60, 420)
	site := affected + r.between(0, affected*2)
	if r.chance(0.25) {
		affected = r.between(500, 900)
		site = affected + r.between(100, 2000)
	}
	euph := pick(r, []string{"Strategic Workforce Realignment", "Organizational Right-Sizing Initiative", "Synergy Capture Program", "Headcount Optimization Event", "Talent Footprint Rebalancing"})
	rows := [][]string{
		{"Program", euph},
		{"Site", "ACPWB " + sl["CITY"] + " Operations Center"},
		{"Employees at site", fmt.Sprint(site)},
		{"Positions eliminated", fmt.Sprint(affected)},
		{"Written notice given", notice.Format("January 2, 2006")},
		{"Separation date", sep.Format("January 2, 2006")},
	}
	q := Question{key: "ex/warn", Exhibit: docExhibit("Notice of Employment Action", "ACPWB employs 4,000+ · federal WARN Act applies · no state mini-WARN", []string{"Item", "Detail"}, rows,
		"Employees will receive information about severance, COBRA continuation, and the Employee Assistance Program.")}
	fmtD := func(t time.Time) string { return t.Format("January 2, 2006") }
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. Does this notice meet the federal WARN Act's advance notice requirement?", ex)
		yes := fmt.Sprintf("Yes; it gives %d days, and at least 60 calendar days are required", days)
		no := fmt.Sprintf("No; it gives %d days, and at least 60 calendar days are required", days)
		right, wrong := yes, no
		if days < 60 {
			right, wrong = no, yes
		}
		q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Yes; any written notice satisfies the WARN Act", "No; the WARN Act requires one year of notice"})
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. Under the federal WARN Act, what is the latest date written notice could have been given for this separation date?", ex)
		latest := sep.AddDate(0, 0, -60)
		q.Choices, q.Correct = withChoices(r, fmtD(latest), []string{fmtD(sep.AddDate(0, 0, -14)), fmtD(sep.AddDate(0, 0, -30)), fmtD(latest.AddDate(0, 0, 1))})
	default:
		q.Text = fmt.Sprintf("Refer to %s. How many calendar days of notice did the affected employees receive?", ex)
		q.Choices, q.Correct = numChoices(r, days, []int{-1, 1, -7, 7, 30, -30}, func(n int) string { return plural(n, "day") })
	}
	return q
}

// ── interview guide (EEOC) ──────────────────────────────────────────────

var interviewOK = []string{
	"Tell me about a project you led from start to finish.",
	"Why are you interested in this role?",
	"Are you legally authorized to work in the United States?",
	"Can you perform the essential functions of this job, with or without reasonable accommodation?",
	"Describe a time you disagreed with a coworker. What happened?",
	"This role requires travel about {N} times a year. Is that workable for you?",
	"What experience do you have with {SYSTEM}?",
	"The schedule includes some early mornings. Are you available to work it?",
	"What would your first 90 days in this role look like?",
	"Tell me about a mistake you made and what you learned.",
	"Do you have any questions for us?",
	"How do you prioritize when everything is due on {WEEKDAY}?",
}

var interviewBad = []struct{ Q, Why string }{
	{"Are you planning to have children soon?", "Sex, pregnancy, and family status"},
	{"Do you have kids at home? Who watches them while you work?", "Sex, pregnancy, and family status"},
	{"What year did you graduate from high school?", "Age"},
	{"How much longer do you plan to work before you retire?", "Age"},
	{"Which church do you attend?", "Religion"},
	{"Will you need time off for religious holidays?", "Religion"},
	{"Where are your parents from originally?", "National origin"},
	{"Is English your first language?", "National origin"},
	{"Do you have any disabilities or medical conditions?", "Disability"},
	{"Have you ever filed a workers' compensation claim?", "Disability"},
	{"Has anyone in your family had cancer or heart disease?", "Genetic information"},
	{"Are you married? What does your spouse do?", "Sex, pregnancy, and family status"},
}

func interviewExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := r.between(5, 7)
	oks := sample(r, interviewOK, n-1)
	bad := pick(r, interviewBad)
	at := r.intn(n)
	var rows [][]string
	k := 0
	for i := 0; i < n; i++ {
		q := ""
		if i == at {
			q = bad.Q
		} else {
			q = sl.fill(oks[k])
			k++
		}
		rows = append(rows, []string{fmt.Sprint(i + 1), q})
	}
	role := pick(r, []string{"Senior Analyst", "Program Coordinator", "Compliance Specialist", "Associate Director", "Facilities Manager", "Data Steward"})
	q := Question{key: "ex/interview", Exhibit: docExhibit("Structured Interview Guide — "+role+", "+sl["DEPT"], "Prepared by "+sl["NAME"]+" · Round 2", []string{"#", "Question"}, rows,
		"Ask every candidate the same questions in the same order. Score answers 1–5 on the rubric.")}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Which question should be removed from the guide before interviews begin?", ex)
		var wrongs []string
		for _, i := range sample(r, rangeInts(n), n) {
			if i != at && len(wrongs) < 3 {
				wrongs = append(wrongs, fmt.Sprintf("Question %d", i+1))
			}
		}
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Question %d", at+1), wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Question %d should be removed because it seeks information about which protected characteristic?", ex, at+1)
	var wrongs []string
	for _, w := range []string{"Age", "Religion", "National origin", "Disability", "Genetic information", "Sex, pregnancy, and family status"} {
		if w != bad.Why {
			wrongs = append(wrongs, w)
		}
	}
	q.Choices, q.Correct = withChoices(r, bad.Why, sample(r, wrongs, 3))
	return q
}

// ── video call grid ─────────────────────────────────────────────────────

func videoGridExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := pick(r, []int{4, 6, 6, 9})
	cols := 3
	if n == 4 {
		cols = 2
	}
	rows := (n + cols - 1) / cols
	const tw, th, gap = 120.0, 80.0, 6.0
	type tile struct {
		name         string
		muted, camOn bool
		whiteboard   bool
	}
	names := sample(r, firstNames, n)
	tiles := make([]tile, n)
	unmuted, camOff := 0, 0
	for i := range tiles {
		tiles[i] = tile{name: names[i] + " " + pick(r, lastNames), muted: r.chance(0.65), camOn: r.chance(0.6)}
		if !tiles[i].muted {
			unmuted++
		}
		if !tiles[i].camOn {
			camOff++
		}
	}
	wb := -1
	mode := r.intn(3)
	if mode == 2 {
		wb = r.intn(n)
		tiles[wb].camOn = true
		tiles[wb].whiteboard = true
		camOff = 0
		for _, t := range tiles {
			if !t.camOn {
				camOff++
			}
		}
	}
	W := float64(cols)*(tw+gap) + gap
	H := float64(rows)*(th+gap) + gap + 30
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="cmp-exhibit-img" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" role="img" aria-label="Video call"><rect width="%.0f" height="%.0f" rx="8" fill="#1c1c1e"/>`, W, H, W, H, W, H)
	for i, t := range tiles {
		x := gap + float64(i%cols)*(tw+gap)
		y := gap + float64(i/cols)*(th+gap)
		bg := "#3a3a3c"
		if t.camOn {
			bg = pick(r, []string{"#6d7f8f", "#8a7a66", "#5f7a6a", "#7a6f86"})
		}
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="6" fill="%s"/>`, x, y, tw, th, bg)
		if t.camOn {
			if t.whiteboard {
				fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="66" height="40" fill="#fafafa" stroke="#999"/><text x="%.0f" y="%.0f" font-family="Comic Sans MS,cursive" font-size="7" fill="#c62828">LAYOFF LIST Q3</text><text x="%.0f" y="%.0f" font-family="Comic Sans MS,cursive" font-size="6" fill="#1565c0">1. Ops  2. Legal</text><text x="%.0f" y="%.0f" font-family="Comic Sans MS,cursive" font-size="6" fill="#1565c0">3. Facilities</text>`,
					x+6, y+6, x+9, y+16, x+9, y+27, x+9, y+37)
			} else {
				fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="22" height="30" fill="#00000022"/>`, x+8+float64(r.intn(70)), y+6)
			}
			fmt.Fprintf(&b, `<circle cx="%.0f" cy="%.0f" r="11" fill="#e8d5c0"/><path d="M%.0f %.0f q20 -22 40 0 v4 h-40 Z" fill="#e8d5c0"/>`, x+tw/2+20, y+36, x+tw/2, y+th-4)
		} else {
			parts := strings.Fields(t.name)
			fmt.Fprintf(&b, `<circle cx="%.0f" cy="%.0f" r="18" fill="#5b8def"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="13" font-weight="700" fill="#fff">%s%s</text>`,
				x+tw/2, y+th/2-6, x+tw/2, y+th/2-1, svgFont, parts[0][:1], parts[1][:1])
		}
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="16" fill="#00000088"/><text x="%.0f" y="%.0f" %s font-size="9" fill="#fff">%s</text>`, x, y+th-16, tw, x+20, y+th-5, svgFont, esc(t.name))
		mx, my := x+10, y+th-8
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="5" height="8" rx="2.5" fill="#fff"/>`, mx-2.5, my-6)
		if t.muted {
			fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#ff453a" stroke-width="1.8"/>`, mx-5, my+3, mx+5, my-7)
		}
	}
	fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="10" fill="#aaa">%s · %d participants · ● REC</text>`, W/2, H-10, svgFont, esc(sl["COMMITTEE"]), n)
	b.WriteString(`</svg>`)
	q := Question{key: "ex/videogrid", Exhibit: b.String()}
	switch mode {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. A red slash on the microphone icon means a participant is muted. How many participants are unmuted?", ex)
		q.Choices, q.Correct = numChoices(r, unmuted, []int{1, -1, 2, -2, 3}, func(n int) string { return plural(n, "participant") })
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. How many participants have their camera turned off?", ex)
		q.Choices, q.Correct = numChoices(r, camOff, []int{1, -1, 2, -2, 3}, func(n int) string { return plural(n, "participant") })
	default:
		q.Text = fmt.Sprintf("Refer to %s. The meeting is being recorded. Whose video is showing confidential information?", ex)
		var wrongs []string
		for i, t := range tiles {
			if i != wb && len(wrongs) < 3 {
				wrongs = append(wrongs, t.name)
			}
		}
		q.Choices, q.Correct = withChoices(r, tiles[wb].name, wrongs)
	}
	return q
}
