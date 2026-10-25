package compliance

import (
	"fmt"
	"math"
	"strings"
)

// Signs and labels: hazard diamonds, signal words, placards, resin codes,
// plugs, road sign shapes, battery and arc flash labels, pipe markings,
// hand signals, tire date codes, alert maps, wiring colors, storage
// compatibility, ladder ratings, high-visibility classes, valve handles.
// All answers follow the published standard named in each generator.

func init() {
	registerExhibits(map[string]exhibitGen{
		"nfpa704":       nfpa704Exhibit,
		"signalwords":   signalWordsExhibit,
		"placards":      placardsExhibit,
		"resincodes":    resinCodesExhibit,
		"safetycolors":  safetyColorsExhibit,
		"plugtypes":     plugTypesExhibit,
		"roadsigns":     roadSignsExhibit,
		"batterywh":     batteryWhExhibit,
		"arcflash":      arcFlashExhibit,
		"pipemarkers":   pipeMarkersExhibit,
		"cylinderlabel": cylinderLabelExhibit,
		"adaparking":    adaParkingExhibit,
		"marshalling":   marshallingExhibit,
		"cranesignals":  craneSignalsExhibit,
		"tiredate":      tireDateExhibit,
		"watchwarning":  watchWarningExhibit,
		"aqicolors":     aqiColorsExhibit,
		"laserclass":    laserClassExhibit,
		"wirecolors":    wireColorsExhibit,
		"breakerdir":    breakerDirExhibit,
		"storagecompat": storageCompatExhibit,
		"ladderduty":    ladderDutyExhibit,
		"hivis":         hiVisExhibit,
		"valvepos":      valvePosExhibit,
	}, map[string][]string{
		"hazcom":         {"nfpa704", "placards", "storagecompat", "pipemarkers", "cylinderlabel"},
		"chemical":       {"nfpa704", "storagecompat", "cylinderlabel"},
		"lab":            {"nfpa704", "storagecompat", "laserclass"},
		"fire":           {"nfpa704", "pipemarkers"},
		"safety":         {"signalwords", "safetycolors", "valvepos"},
		"signage":        {"signalwords", "safetycolors", "roadsigns"},
		"ppe":            {"signalwords", "hivis", "arcflash"},
		"dgr":            {"placards", "batterywh"},
		"mail":           {"placards", "batterywh"},
		"deliveries":     {"placards", "tiredate"},
		"sustainability": {"resincodes"},
		"waste":          {"resincodes"},
		"kitchen":        {"resincodes"},
		"facilities":     {"safetycolors", "pipemarkers", "breakerdir", "valvepos"},
		"intltravel":     {"plugtypes"},
		"travel":         {"plugtypes", "batterywh"},
		"devices":        {"plugtypes", "batterywh"},
		"driving":        {"roadsigns", "tiredate", "hivis"},
		"commuting":      {"roadsigns"},
		"vehicles":       {"tiredate"},
		"fleet":          {"tiredate"},
		"lithium":        {"batterywh"},
		"cabin":          {"batterywh"},
		"electrical":     {"arcflash", "wirecolors", "breakerdir"},
		"lockout":        {"arcflash", "breakerdir", "valvepos"},
		"hotwork":        {"cylinderlabel"},
		"construction":   {"cranesignals", "hivis", "ladderduty"},
		"parking":        {"adaparking"},
		"accessibility":  {"adaparking"},
		"ramp":           {"marshalling"},
		"faa":            {"marshalling"},
		"flightdept":     {"marshalling"},
		"warehouse":      {"cranesignals", "ladderduty"},
		"weather":        {"watchwarning", "aqicolors"},
		"closures":       {"watchwarning"},
		"airquality":     {"aqicolors"},
		"fieldwork":      {"aqicolors", "hivis"},
		"presentations":  {"laserclass"},
		"ladder":         {"ladderduty"},
		"fallprotection": {"ladderduty"},
		"maintenance":    {"valvepos", "breakerdir"},
	})
}

// ── NFPA 704 diamonds ───────────────────────────────────────────────────

func nfpa704Exhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type d704 struct {
		h, f, i int
		sp      string
	}
	mode := r.intn(4)
	ds := make([]d704, 4)
	target := r.intn(4)
	for {
		for k := range ds {
			ds[k] = d704{r.between(0, 3), r.between(0, 3), r.between(0, 2), ""}
		}
		switch mode {
		case 2:
			ds[target].sp = "W"
		case 3:
			ds[target].sp = "OX"
		}
		ok := true
		for k := range ds {
			if k == target {
				continue
			}
			if mode == 0 && ds[k].f >= ds[target].f || mode == 1 && ds[k].h >= ds[target].h {
				ok = false
			}
		}
		if ok {
			break
		}
		if mode == 0 {
			ds[target].f = 4
		} else if mode == 1 {
			ds[target].h = 4
		}
	}
	s := frame(540, 200, "Four NFPA 704 hazard diamonds", "NFPA 704 PLACARDS — CHEMICAL STORAGE ROOM")
	for k, d := range ds {
		cx, cy := 72+132*float64(k), 110.0
		q := func(dx, dy float64, fill string, txt string, tc string) {
			s.raw(`<path d="M%.1f %.1f l26 -26 l26 26 l-26 26 Z" fill="%s" stroke="#111" stroke-width="1.5"/>`, cx+dx-26, cy+dy, fill)
			s.text(cx+dx, cy+dy+6, 16, `text-anchor="middle" font-weight="700" fill="`+tc+`"`, txt)
		}
		q(0, -26, "#dc2626", itoa(d.f), "#fff")
		q(-26, 0, "#2563eb", itoa(d.h), "#fff")
		q(26, 0, "#facc15", itoa(d.i), "#111")
		sp := d.sp
		if sp == "W" {
			sp = "W̶"
		}
		q(0, 26, "#fff", sp, "#111")
		markAt(s, cx-50, cy-48, k)
	}
	asks := []string{
		"Which chemical has the highest flammability rating?",
		"Which chemical has the most severe health hazard rating?",
		"Which chemical reacts dangerously with water?",
		"Which chemical is an oxidizer?",
	}
	qq := Question{key: "ex/nfpa704", Exhibit: s.done()}
	qq.Text = fmt.Sprintf("Refer to %s. %s", ex, asks[mode])
	qq.Choices, qq.Correct = letterQ(r, "Diamond", 4, target)
	return qq
}

// ── ANSI Z535 signal words ──────────────────────────────────────────────

func signalWordsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	words := []struct{ w, bg, fg, means string }{
		{"DANGER", "#dc2626", "#fff", "a hazard that WILL cause death or serious injury if not avoided"},
		{"WARNING", "#f97316", "#111", "a hazard that COULD cause death or serious injury if not avoided"},
		{"CAUTION", "#facc15", "#111", "a hazard that could cause minor or moderate injury"},
		{"NOTICE", "#2563eb", "#fff", "information about property damage, not personal injury"},
	}
	order := orderN(r, 4)
	s := frame(540, 170, "Four safety sign headers", "SAFETY SIGNS — WHICH HEADER?")
	for i, k := range order {
		w := words[k]
		x := 14 + 131*float64(i)
		s.rect(x, 44, 120, 100, 4, "#fff", "#111")
		s.rect(x+4, 48, 112, 30, 2, w.bg, "")
		if w.w == "CAUTION" || w.w == "WARNING" || w.w == "DANGER" {
			s.raw(`<path d="M%.1f 72 l9 -17 l9 17 Z" fill="%s" stroke="%s"/>`, x+10, w.fg, w.fg)
			s.text(x+19, 70, 10, `text-anchor="middle" font-weight="700" fill="`+w.bg+`"`, "!")
		}
		s.text(x+68, 69, 13, `text-anchor="middle" font-weight="700" fill="`+w.fg+`"`, w.w)
		for l := 0; l < 3; l++ {
			s.rect(x+14, 92+14*float64(l), 92-float64(l*18), 5, 2, "#cbd5e1", "")
		}
		markAt(s, x+110, 140, i)
	}
	t := r.intn(4)
	pos := 0
	for i, k := range order {
		if k == t {
			pos = i
		}
	}
	q := Question{key: "ex/signalwords", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Under ANSI Z535, which sign header should be used for %s?", ex, words[t].means)
	q.Choices, q.Correct = letterQ(r, "Sign", 4, pos)
	return q
}

// ── DOT hazard class placards ───────────────────────────────────────────

func placardsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type pl struct{ num, name, top, bot, fg string }
	all := []pl{
		{"1.1", "explosives", "#f97316", "#f97316", "#111"},
		{"2", "a non-flammable compressed gas", "#16a34a", "#16a34a", "#fff"},
		{"3", "a flammable liquid", "#dc2626", "#dc2626", "#fff"},
		{"5.1", "an oxidizer", "#facc15", "#facc15", "#111"},
		{"6", "a poison (toxic) substance", "#fff", "#fff", "#111"},
		{"8", "a corrosive", "#fff", "#111", "#fff"},
		{"9", "miscellaneous dangerous goods, such as lithium batteries", "stripes", "#fff", "#111"},
	}
	show := sample(r, all, 4)
	t := r.intn(4)
	s := frame(540, 190, "Four DOT hazard placards", "TRUCK AT THE DOCK — HAZARD PLACARDS")
	for i, p := range show {
		cx, cy := 72+132*float64(i), 108.0
		s.raw(`<path d="M%.1f %.1f l52 -52 l52 52 l-52 52 Z" fill="#fff" stroke="#111" stroke-width="2"/>`, cx-52, cy)
		switch p.top {
		case "stripes":
			for k := 0; k < 7; k++ {
				s.raw(`<rect x="%.1f" y="%.1f" width="5" height="%.1f" fill="#111"/>`, cx-24+float64(k)*7, cy-46+math.Abs(float64(k)-3)*6, 46-math.Abs(float64(k)-3)*6)
			}
		default:
			s.raw(`<path d="M%.1f %.1f l48 -48 l48 48 Z" fill="%s"/>`, cx-48, cy, p.top)
		}
		s.raw(`<path d="M%.1f %.1f l48 48 l48 -48 Z" fill="%s"/>`, cx-48, cy, p.bot)
		s.raw(`<path d="M%.1f %.1f l52 -52 l52 52 l-52 52 Z" fill="none" stroke="#111" stroke-width="2"/>`, cx-52, cy)
		s.text(cx, cy+36, 15, `text-anchor="middle" font-weight="700" fill="`+p.fg+`" stroke="`+map[bool]string{true: "#fff", false: "none"}[p.fg == "#111" && p.bot == "#111"]+`"`, p.num)
		markAt(s, cx-54, cy-50, i)
	}
	q := Question{key: "ex/placards", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which placard tells responders the load is %s?", ex, show[t].name)
	q.Choices, q.Correct = letterQ(r, "Placard", 4, t)
	return q
}

// ── resin identification codes ──────────────────────────────────────────

func resinCodesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := map[int]string{1: "PET (polyethylene terephthalate), such as water bottles", 2: "HDPE (high-density polyethylene), such as milk jugs", 3: "PVC (vinyl)", 4: "LDPE (low-density polyethylene), such as bags and films", 5: "polypropylene (PP), such as yogurt tubs", 6: "polystyrene (PS), such as foam cups", 7: "\"other\" plastics or mixed resins"}
	codes := sample(r, []int{1, 2, 3, 4, 5, 6, 7}, 4)
	t := r.intn(4)
	s := frame(540, 170, "Four plastic containers with resin codes", "BREAK ROOM CONTAINERS — RESIN CODES")
	for i, code := range codes {
		cx := 72 + 132*float64(i)
		s.rect(cx-30, 44, 60, 84, 10, "#e0f2fe", "#64748b")
		s.raw(`<path d="M%.1f 106 l14 -24 l14 24 Z" fill="none" stroke="#111" stroke-width="2"/>`, cx-14)
		s.text(cx, 103, 11, `text-anchor="middle" font-weight="700"`, itoa(code))
		markAt(s, cx-40, 48, i)
	}
	q := Question{key: "ex/resincodes", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which container is made of %s?", ex, names[codes[t]])
	q.Choices, q.Correct = letterQ(r, "Container", 4, t)
	return q
}

// ── ANSI Z535.1 safety colors ───────────────────────────────────────────

func safetyColorsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cols := []struct{ hex, means string }{
		{"#dc2626", "fire protection equipment and DANGER"},
		{"#f97316", "WARNING and dangerous machine parts"},
		{"#facc15", "CAUTION and physical hazards such as tripping or striking"},
		{"#16a34a", "safety equipment and first aid"},
		{"#2563eb", "NOTICE and safety information"},
	}
	show := sample(r, orderN(r, 5), 4)
	t := r.intn(4)
	s := frame(540, 170, "Four color-coded cabinets", "COLOR-CODED EQUIPMENT CABINETS")
	for i, k := range show {
		x := 24 + 130*float64(i)
		s.rect(x, 44, 96, 100, 6, cols[k].hex, "#111")
		s.line(x+48, 50, x+48, 138, "#0004", 1.5)
		s.rect(x+38, 88, 6, 14, 2, "#e5e7eb", "")
		s.rect(x+52, 88, 6, 14, 2, "#e5e7eb", "")
		markAt(s, x+8, 52, i)
	}
	q := Question{key: "ex/safetycolors", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Under ANSI Z535.1, which cabinet's color identifies %s?", ex, cols[show[t]].means)
	q.Choices, q.Correct = letterQ(r, "Cabinet", 4, t)
	return q
}

// ── plug types ──────────────────────────────────────────────────────────

func plugTypesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	dests := map[string][]string{
		"A": {"Tokyo"},
		"C": {"Berlin", "Madrid", "Amsterdam", "Vienna"},
		"G": {"London", "Dublin", "Singapore", "Hong Kong"},
		"I": {"Sydney", "Auckland", "Melbourne"},
	}
	types := []string{"A", "B", "C", "G", "I"}
	var show []string
	var dest, want string
	for {
		show = sample(r, types, 4)
		want = pick(r, []string{"A", "C", "G", "I"})
		if !contains(show, want) || (want == "A" && contains(show, "B")) {
			continue
		}
		dest = pick(r, dests[want])
		break
	}
	s := frame(540, 170, "Four travel plug types", "TRAVEL ADAPTERS — PLUG FACES")
	for i, t := range show {
		cx, cy := 72+132*float64(i), 96.0
		s.rect(cx-40, cy-40, 80, 80, 14, "#f1f5f9", "#334155")
		switch t {
		case "A":
			s.rect(cx-16, cy-14, 6, 22, 1, "#94a3b8", "#334155")
			s.rect(cx+10, cy-14, 6, 22, 1, "#94a3b8", "#334155")
		case "B":
			s.rect(cx-16, cy-18, 6, 22, 1, "#94a3b8", "#334155")
			s.rect(cx+10, cy-18, 6, 22, 1, "#94a3b8", "#334155")
			s.circle(cx, cy+16, 5, "#94a3b8", "#334155")
		case "C":
			s.circle(cx-14, cy, 6, "#94a3b8", "#334155")
			s.circle(cx+14, cy, 6, "#94a3b8", "#334155")
		case "G":
			s.rect(cx-4, cy-24, 8, 18, 1, "#94a3b8", "#334155")
			s.rect(cx-24, cy+6, 16, 8, 1, "#94a3b8", "#334155")
			s.rect(cx+8, cy+6, 16, 8, 1, "#94a3b8", "#334155")
		case "I":
			s.raw(`<path d="M%.1f %.1f l-9 16 l5 3 l9 -16 Z" fill="#94a3b8" stroke="#334155"/>`, cx-8, cy-14)
			s.raw(`<path d="M%.1f %.1f l9 16 l-5 3 l-9 -16 Z" fill="#94a3b8" stroke="#334155"/>`, cx+8, cy-14)
			s.rect(cx-3, cy+10, 6, 16, 1, "#94a3b8", "#334155")
		}
		markAt(s, cx-46, cy-44, i)
	}
	pos := 0
	for i, t := range show {
		if t == want {
			pos = i
		}
	}
	q := Question{key: "ex/plugtypes", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. You are traveling to %s for the regional sales meeting. Which plug fits the wall outlets there?", ex, dest)
	q.Choices, q.Correct = letterQ(r, "Plug", 4, pos)
	return q
}

// ── road sign shapes (MUTCD) ────────────────────────────────────────────

func roadSignsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type sg struct{ shape, means string }
	all := []sg{
		{"octagon", "come to a complete stop"},
		{"yield", "yield the right of way"},
		{"pentagon", "a school zone or school crossing"},
		{"diamond", "a general warning of a hazard ahead"},
		{"circle", "a railroad crossing ahead"},
		{"pennant", "a no-passing zone"},
	}
	show := sample(r, all, 4)
	t := r.intn(4)
	s := frame(540, 170, "Four blank road sign shapes", "ROAD SIGN SHAPES (TEXT REMOVED)")
	for i, g := range show {
		cx, cy := 72+132*float64(i), 98.0
		switch g.shape {
		case "octagon":
			var pts []string
			for k := 0; k < 8; k++ {
				a := math.Pi/8 + float64(k)*math.Pi/4
				pts = append(pts, fmt.Sprintf("%.1f,%.1f", cx+42*math.Cos(a), cy+42*math.Sin(a)))
			}
			s.raw(`<polygon points="%s" fill="#dc2626" stroke="#fff" stroke-width="3"/>`, strings.Join(pts, " "))
		case "yield":
			s.raw(`<path d="M%.1f %.1f h84 l-42 72 Z" fill="#fff" stroke="#dc2626" stroke-width="9"/>`, cx-42, cy-36)
		case "pentagon":
			s.raw(`<path d="M%.1f %.1f l34 -30 l34 30 v42 h-68 Z" fill="#c6f432" stroke="#111" stroke-width="2"/>`, cx-34, cy-12)
		case "diamond":
			s.raw(`<path d="M%.1f %.1f l40 -40 l40 40 l-40 40 Z" fill="#facc15" stroke="#111" stroke-width="2"/>`, cx-40, cy)
		case "circle":
			s.circle(cx, cy, 40, "#facc15", "#111")
		case "pennant":
			s.raw(`<path d="M%.1f %.1f v60 l84 -30 Z" fill="#facc15" stroke="#111" stroke-width="2"/>`, cx-42, cy-30)
		}
		markAt(s, cx-50, cy-46, i)
	}
	q := Question{key: "ex/roadsigns", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which sign shape means %s?", ex, show[t].means)
	q.Choices, q.Correct = letterQ(r, "Sign", 4, t)
	return q
}

// ── battery watt-hours (FAA) ────────────────────────────────────────────

func batteryWhExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	approval := r.chance(0.5)
	t := r.intn(4)
	var labels [][2]float64 // volts, mAh
	var whs []float64
	for i := 0; i < 4; i++ {
		v := pick(r, []float64{3.6, 3.7, 7.4, 11.1, 14.4})
		var lo, hi float64 = 10, 95
		if i == t {
			if approval {
				lo, hi = 105, 155
			} else {
				lo, hi = 170, 280
			}
		}
		wh := lo + r.float()*(hi-lo)
		mah := math.Round(wh*1000/v/100) * 100
		labels = append(labels, [2]float64{v, mah})
		whs = append(whs, v*mah/1000)
	}
	s := frame(540, 180, "Four spare battery labels", "SPARE BATTERIES IN YOUR CARRY-ON")
	for i, l := range labels {
		x := 18 + 131*float64(i)
		s.rect(x, 44, 112, 110, 8, "#111827", "")
		s.rect(x+8, 52, 96, 94, 4, "#f8fafc", "")
		s.text(x+56, 72, 9, `text-anchor="middle" font-weight="700"`, "Li-ion Battery Pack")
		s.text(x+56, 94, 10, `text-anchor="middle" font-weight="700" fill="#1d4ed8"`, fmt.Sprintf("%.1f V", l[0]))
		s.text(x+56, 112, 10, `text-anchor="middle" font-weight="700" fill="#1d4ed8"`, commas(int(l[1]))+" mAh")
		s.text(x+56, 134, 7, `text-anchor="middle" fill="#64748b"`, "Do not incinerate")
		markAt(s, x+10, 50, i)
	}
	q := Question{key: "ex/batterywh", Exhibit: s.done()}
	ask := "is over 160 Wh and may not be carried on a passenger aircraft at all"
	if approval {
		ask = "is between 101 and 160 Wh and needs airline approval"
	}
	q.Text = fmt.Sprintf("Refer to %s. Watt-hours = volts × milliamp-hours ÷ 1,000. Which spare battery %s?", ex, ask)
	q.Choices, q.Correct = letterQ(r, "Battery", 4, t)
	return q
}

// ── NFPA 70E arc flash label ────────────────────────────────────────────

func arcFlashExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ie := float64(r.between(12, 380)) / 10
	afb := r.between(24, 120)
	volts := pick(r, []int{208, 480, 600})
	lab := map[int]int{208: 42, 480: 42, 600: 42}[volts]
	rab := map[int]int{208: 12, 480: 12, 600: 12}[volts]
	minRating := int(math.Ceil(ie))
	s := frame(400, 250, "Arc flash warning label", "PANEL MDP-2 — EQUIPMENT LABEL")
	s.rect(40, 32, 320, 206, 4, "#fff", "#111")
	s.rect(44, 36, 312, 34, 2, "#f97316", "")
	s.text(200, 59, 16, `text-anchor="middle" font-weight="700" fill="#111"`, "⚠ WARNING")
	s.text(200, 86, 10, `text-anchor="middle" font-weight="700"`, "Arc Flash and Shock Hazard — Appropriate PPE Required")
	rows := [][2]string{
		{"Arc flash boundary", fmt.Sprintf("%d in", afb)},
		{"Incident energy at 18 in", fmt.Sprintf("%.1f cal/cm²", ie)},
		{"Shock hazard when cover is removed", fmt.Sprintf("%d VAC", volts)},
		{"Limited approach boundary", fmt.Sprintf("%d in", lab)},
		{"Restricted approach boundary", fmt.Sprintf("%d in", rab)},
		{"Equipment", "MDP-2 · Study date 03/2025"},
	}
	for i, rw := range rows {
		y := 108 + 20*float64(i)
		s.text(56, y, 9.5, ``, rw[0])
		s.text(344, y, 9.5, `text-anchor="end" font-weight="700"`, rw[1])
	}
	q := Question{key: "ex/arcflash", Exhibit: s.done()}
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. Under NFPA 70E, what is the minimum arc rating for the PPE worn inside the arc flash boundary of this panel?", ex)
		q.Choices, q.Correct = numChoices(r, minRating, []int{-2, 4, 8, 12, -4}, func(n int) string { return fmt.Sprintf("%d cal/cm²", n) })
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. Inside what distance must arc-rated PPE be worn when this panel is open?", ex)
		q.Choices, q.Correct = numChoices(r, afb, []int{-12, 12, 6, -6, 18}, func(n int) string { return fmt.Sprintf("%d in", n) })
	default:
		q.Text = fmt.Sprintf("Refer to %s. An unqualified person may not cross which distance unless escorted by a qualified person?", ex)
		var w []string
		for _, rw := range rows[:5] {
			if rw[0] != "Limited approach boundary" {
				w = append(w, rw[0]+" ("+rw[1]+")")
			}
		}
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Limited approach boundary (%d in)", lab), sample(r, w, 3))
	}
	return q
}

// ── ASME A13.1 pipe marker colors ───────────────────────────────────────

func pipeMarkersExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type pm struct{ bg, fg, means string }
	all := []pm{
		{"#dc2626", "#fff", "fire-quenching water, such as a sprinkler line"},
		{"#f97316", "#111", "a toxic or corrosive fluid"},
		{"#facc15", "#111", "a flammable or oxidizing fluid"},
		{"#7c4a1e", "#fff", "a combustible fluid"},
		{"#16a34a", "#fff", "potable, cooling, or boiler feed water"},
		{"#2563eb", "#fff", "compressed air"},
	}
	show := sample(r, all, 4)
	t := r.intn(4)
	s := frame(540, 220, "Four labeled pipes with color bands", "MECHANICAL ROOM — PIPE MARKERS (TEXT OBSCURED)")
	for i, p := range show {
		y := 50 + 40*float64(i)
		s.rect(60, y, 440, 22, 11, "#cbd5e1", "#64748b")
		s.rect(200, y-2, 160, 26, 3, p.bg, "#111")
		for k := 0; k < 5; k++ {
			s.rect(215+float64(k)*26, y+9, 18, 6, 2, p.fg, "")
		}
		s.raw(`<path d="M380 %.1f h50 v-6 l14 11 l-14 11 v-6 h-50 Z" fill="%s" stroke="#111"/>`, y+5, p.bg)
		markAt(s, 40, y+11, i)
	}
	q := Question{key: "ex/pipemarkers", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the ASME A13.1 color scheme, which pipe carries %s?", ex, show[t].means)
	q.Choices, q.Correct = letterQ(r, "Pipe", 4, t)
	return q
}

// ── gas cylinders: read the label, not the color ────────────────────────

func cylinderLabelExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	gases := sample(r, []string{"OXYGEN", "ACETYLENE", "NITROGEN", "ARGON", "HELIUM", "CARBON DIOXIDE"}, 4)
	colors := sample(r, []string{"#16a34a", "#dc2626", "#64748b", "#1d4ed8", "#a16207", "#111827"}, 4)
	t := r.intn(4)
	s := frame(540, 220, "Four compressed gas cylinders", "CYLINDER CAGE — DELIVERY FROM TWO SUPPLIERS")
	for i, g := range gases {
		cx := 72 + 132*float64(i)
		s.rect(cx-22, 52, 44, 148, 18, colors[i], "#111")
		s.rect(cx-8, 38, 16, 16, 3, "#9ca3af", "#111")
		s.rect(cx-38, 110, 76, 30, 3, "#fff", "#111")
		s.text(cx, 124, 7, `text-anchor="middle" font-weight="700"`, g)
		s.text(cx, 134, 6, `text-anchor="middle" fill="#475569"`, "Compressed gas")
		markAt(s, cx-34, 46, i)
	}
	q := Question{key: "ex/cylinderlabel", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Cylinder colors are not standardized between U.S. suppliers. Which cylinder contains %s?", ex, strings.ToLower(gases[t]))
	q.Choices, q.Correct = letterQ(r, "Cylinder", 4, t)
	return q
}

// ── ADA 208.2 accessible parking count ──────────────────────────────────

func adaRequired(n int) int {
	switch {
	case n <= 25:
		return 1
	case n <= 50:
		return 2
	case n <= 75:
		return 3
	case n <= 100:
		return 4
	case n <= 150:
		return 5
	case n <= 200:
		return 6
	case n <= 300:
		return 7
	case n <= 400:
		return 8
	case n <= 500:
		return 9
	case n <= 1000:
		return int(math.Ceil(float64(n) * 0.02))
	}
	return 20 + int(math.Ceil(float64(n-1000)/100))
}

func adaParkingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	rows := r.between(3, 5)
	var counts []int
	total := 0
	for i := 0; i < rows; i++ {
		n := r.between(14, 70)
		counts = append(counts, n)
		total += n
	}
	s := frame(540, 60+44*float64(rows)+20, "Parking lot plan with stall counts", "LOT C — PARKING PLAN")
	for i, n := range counts {
		y := 44 + 44*float64(i)
		s.rect(20, y, 500, 32, 2, "#e5e7eb", "")
		shown := minInt(n, 40)
		for k := 0; k <= shown; k++ {
			x := 24 + float64(k)*(400/float64(shown))
			s.line(x, y+2, x, y+30, "#fff", 2)
		}
		s.text(510, y+20, 9, `text-anchor="end" font-weight="700"`, fmt.Sprintf("Row %s: %d spaces", lettered(i), n))
	}
	s.text(20, 60+44*float64(rows)+10, 8.5, `fill="#334155"`, "Stall lines are schematic; use the counts. The lot serves the headquarters building.")
	need := adaRequired(total)
	q := Question{key: "ex/adaparking", Exhibit: s.done()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Using the ADA 208.2 table, how many accessible parking spaces does this lot need in total?", ex)
		q.Choices, q.Correct = numChoices(r, need, []int{-1, 1, 2, -2, 3}, itoa)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. The lot needs %d accessible spaces. Under ADA 208.2.4, at least how many of them must be van-accessible?", ex, need)
	q.Choices, q.Correct = numChoices(r, (need+5)/6, []int{1, 2, 3, -1}, itoa)
	return q
}

// ── hand signals (FAA AIM marshalling, ASME B30.5 crane) ────────────────

type pose struct {
	name, means string
	// shoulder→elbow→hand angles (degrees, 0 = right, 90 = up) for each arm
	l1, l2, r1, r2 float64
	extra          string // "throat", "circleup", "circledown"
}

func drawPose(s *svgB, cx, cy float64, p pose, wands bool) {
	s.circle(cx, cy-46, 9, "#fde68a", "#111")
	s.line(cx, cy-37, cx, cy+6, "#111", 3)
	s.line(cx, cy+6, cx-12, cy+40, "#111", 3)
	s.line(cx, cy+6, cx+12, cy+40, "#111", 3)
	arm := func(sx, a1, a2 float64) (float64, float64) {
		ex := sx + 18*math.Cos(a1*math.Pi/180)
		ey := cy - 30 - 18*math.Sin(a1*math.Pi/180)
		hx := ex + 18*math.Cos(a2*math.Pi/180)
		hy := ey - 18*math.Sin(a2*math.Pi/180)
		s.line(sx, cy-30, ex, ey, "#111", 3)
		s.line(ex, ey, hx, hy, "#111", 3)
		return hx, hy
	}
	lx, ly := arm(cx-4, p.l1, p.l2)
	rx, ry := arm(cx+4, p.r1, p.r2)
	if wands {
		for _, h := range [][2]float64{{lx, ly}, {rx, ry}} {
			s.circle(h[0], h[1], 4, "#f97316", "")
		}
	}
	switch p.extra {
	case "circleup":
		s.raw(`<circle cx="%.1f" cy="%.1f" r="6" fill="none" stroke="#1d4ed8" stroke-dasharray="3 2"/>`, rx, ry-8)
	case "circledown":
		s.raw(`<circle cx="%.1f" cy="%.1f" r="6" fill="none" stroke="#1d4ed8" stroke-dasharray="3 2"/>`, rx, ry+8)
	case "wave":
		s.raw(`<path d="M%.1f %.1f h-14 m28 0 h-14" stroke="#1d4ed8" stroke-width="1.5" stroke-dasharray="3 2"/>`, rx, ry-6)
	case "wave2":
		s.raw(`<path d="M%.1f %.1f h-14 m28 0 h-14 M%.1f %.1f h-14 m28 0 h-14" stroke="#1d4ed8" stroke-width="1.5" stroke-dasharray="3 2"/>`, rx, ry-6, lx, ly-6)
	case "updown":
		s.raw(`<path d="M%.1f %.1f v-12 m0 24 v-12 M%.1f %.1f v-12 m0 24 v-12" stroke="#1d4ed8" stroke-width="1.5" stroke-dasharray="3 2"/>`, rx+6, ry, lx-6, ly)
	}
}

func signalExhibit(r *rnd, key, title, who string, poses []pose, wands bool, ex string) Question {
	show := sample(r, poses, 4)
	t := r.intn(4)
	s := frame(540, 170, title, title)
	for i, p := range show {
		cx := 72 + 132*float64(i)
		drawPose(s, cx, 112, p, wands)
		markAt(s, cx-48, 44, i)
	}
	q := Question{key: key, Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which signal from the %s means %s?", ex, who, show[t].means)
	q.Choices, q.Correct = letterQ(r, "Signal", 4, t)
	return q
}

func marshallingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	return signalExhibit(r, "ex/marshalling", "AIRCRAFT MARSHALLING SIGNALS", "marshaller", []pose{
		{"stop", "STOP (arms crossed above the head)", 70, 25, 110, 155, ""},
		{"cut", "CUT ENGINES (hand drawn across the throat)", 230, 60, 280, 270, "throat"},
		{"slow", "SLOW DOWN (arms down, palms toward the ground, moved up and down)", 225, 260, 315, 280, "updown"},
		{"ahead", "MOVE AHEAD (arms raised, beckoning backward)", 135, 100, 45, 80, ""},
	}, true, ex)
}

func craneSignalsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	return signalExhibit(r, "ex/cranesignals", "CRANE HAND SIGNALS (ASME B30.5)", "signal person", []pose{
		{"stop", "STOP (one arm extended, palm down, moved back and forth)", 250, 270, 0, 0, "wave"},
		{"estop", "EMERGENCY STOP (both arms extended, palms down, moved back and forth)", 180, 180, 0, 0, "wave2"},
		{"hoist", "HOIST (forearm vertical, forefinger up, small circles)", 250, 270, 0, 90, "circleup"},
		{"lower", "LOWER (arm extended down, forefinger down, small circles)", 250, 270, 290, 270, "circledown"},
	}, false, ex)
}

// ── DOT tire date codes ─────────────────────────────────────────────────

func tireDateExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	t := r.intn(4)
	var codes []string
	var years []int
	for i := 0; i < 4; i++ {
		y := r.between(2020, 2025)
		if i == t {
			y = r.between(2014, 2018)
		}
		w := r.between(1, 52)
		years = append(years, y)
		codes = append(codes, fmt.Sprintf("%02d%02d", w, y%100))
	}
	s := frame(540, 190, "Four tire sidewall DOT codes", "FLEET TIRES — SIDEWALL DOT CODES")
	for i, cd := range codes {
		cx := 72 + 132*float64(i)
		s.circle(cx, 112, 52, "#1f2937", "")
		s.circle(cx, 112, 30, "#9ca3af", "#111")
		s.raw(`<path id="tp%d" d="M%.1f 112 A42 42 0 0 1 %.1f 112" fill="none"/>`, i, cx-42, cx+42)
		fmt.Fprintf(s, `<text %s font-size="8" font-weight="700" fill="#f8fafc"><textPath href="#tp%d" startOffset="50%%" text-anchor="middle">DOT U2LL LMLR %s</textPath></text>`, svgFont, i, cd)
		s.text(cx, 180, 9, `text-anchor="middle" font-weight="700"`, "…LMLR "+cd)
		markAt(s, cx-50, 58, i)
	}
	q := Question{key: "ex/tiredate", Exhibit: s.done()}
	if r.chance(0.6) {
		q.Text = fmt.Sprintf("Refer to %s. The last four digits of a DOT code are the week and year of manufacture. Fleet policy replaces tires more than 6 years old, and today is October 2025. Which tire must be replaced?", ex)
		q.Choices, q.Correct = letterQ(r, "Tire", 4, t)
		return q
	}
	k := r.intn(4)
	w := codes[k][:2]
	right := fmt.Sprintf("Week %s of %d", strings.TrimPrefix(w, "0"), years[k])
	wrongs := []string{
		fmt.Sprintf("Week %d of %d", years[k]%100, 2000+atoiSafe(w)),
		fmt.Sprintf("%s %d", monthAbbr[(atoiSafe(w)+11)%12], years[k]),
		fmt.Sprintf("Week %s of %d", strings.TrimPrefix(w, "0"), years[k]-10),
	}
	q.Text = fmt.Sprintf("Refer to %s. When was tire %s made?", ex, markLetters[k])
	q.Choices, q.Correct = withChoices(r, right, dedupeNot(wrongs, right))
	return q
}

func atoiSafe(s string) int {
	n := 0
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

func dedupeNot(xs []string, not string) []string {
	var out []string
	for _, x := range xs {
		if x != not && !contains(out, x) {
			out = append(out, x)
		}
	}
	return out
}

// ── NWS watch vs. warning map ───────────────────────────────────────────

func watchWarningExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	const cols, rows = 6, 4
	state := make([]int, cols*rows) // 0 none, 1 watch, 2 warning
	wx0, wy0 := r.between(0, 2), r.between(0, 1)
	for y := wy0; y < wy0+3 && y < rows; y++ {
		for x := wx0; x < wx0+4 && x < cols; x++ {
			state[y*cols+x] = 1
		}
	}
	warn := (wy0+r.between(0, 1))*cols + wx0 + r.between(0, 2)
	state[warn] = 2
	// offices: one in the warning county, others in watch or clear counties
	var offices []int
	offices = append(offices, warn)
	for len(offices) < 4 {
		k := r.intn(cols * rows)
		if state[k] != 2 && !containsInt(offices, k) {
			offices = append(offices, k)
		}
	}
	order := orderN(r, 4)
	s := frame(540, 250, "County map with tornado watch and warning", "NATIONAL WEATHER SERVICE — ACTIVE ALERTS")
	for k, st := range state {
		x, y := 40+float64(k%cols)*76, 40+float64(k/cols)*44
		fill := map[int]string{0: "#f1f5f9", 1: "#fde047", 2: "#dc2626"}[st]
		s.rect(x, y, 76, 44, 0, fill, "#64748b")
	}
	for i, o := range order {
		k := offices[o]
		markAt(s, 40+float64(k%cols)*76+38, 40+float64(k/cols)*44+22, i)
	}
	s.rect(40, 226, 12, 10, 0, "#dc2626", "")
	s.text(56, 235, 8.5, ``, "Tornado Warning")
	s.rect(160, 226, 12, 10, 0, "#fde047", "")
	s.text(176, 235, 8.5, ``, "Tornado Watch")
	pos := 0
	for i, o := range order {
		if o == 0 {
			pos = i
		}
	}
	q := Question{key: "ex/watchwarning", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which office should send employees to shelter immediately?", ex)
	q.Choices, q.Correct = letterQ(r, "Office", 4, pos)
	return q
}

// ── EPA AQI colors ──────────────────────────────────────────────────────

func aqiColorsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cats := []struct{ hex, name string }{
		{"#00e400", "Good"},
		{"#ffff00", "Moderate"},
		{"#ff7e00", "Unhealthy for Sensitive Groups"},
		{"#ff0000", "Unhealthy"},
		{"#8f3f97", "Very Unhealthy"},
		{"#7e0023", "Hazardous"},
	}
	show := sample(r, orderN(r, 6), 4)
	t := r.intn(4)
	s := frame(540, 160, "Four AQI color badges for job sites", "TODAY'S AIR QUALITY — JOB SITES")
	for i, k := range show {
		cx := 72 + 132*float64(i)
		s.circle(cx, 92, 36, cats[k].hex, "#111")
		s.text(cx, 96, 9, `text-anchor="middle" font-weight="700" fill="#111" stroke="#fff" stroke-width="2.5" paint-order="stroke"`, "AQI")
		s.text(cx, 146, 8.5, `text-anchor="middle" font-weight="700"`, "Site "+markLetters[i])
		markAt(s, cx-40, 52, i)
	}
	q := Question{key: "ex/aqicolors", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the EPA's AQI color scale, which site's air is rated \"%s\"?", ex, cats[show[t]].name)
	q.Choices, q.Correct = letterQ(r, "Site", 4, t)
	return q
}

// ── laser pointer classes ───────────────────────────────────────────────

func laserClassExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type lz struct {
		class string
		mw    float64
	}
	ok := pick(r, []lz{{"3R", 4.5}, {"2", 0.9}, {"3R", 4.9}, {"2", 0.5}})
	bad := sample(r, []lz{{"3B", 50}, {"3B", 120}, {"4", 1000}, {"4", 500}, {"3B", 300}}, 3)
	all := append([]lz{ok}, bad...)
	order := orderN(r, 4)
	s := frame(540, 160, "Four laser pointer labels", "PRESENTATION LASER POINTERS — LABELS")
	pos := 0
	for i, k := range order {
		l := all[k]
		if k == 0 {
			pos = i
		}
		x := 18 + 131*float64(i)
		s.rect(x, 50, 112, 80, 4, "#facc15", "#111")
		s.raw(`<path d="M%.1f 62 l10 -8 v16 Z" fill="#111"/>`, x+10)
		s.text(x+60, 72, 8, `text-anchor="middle" font-weight="700"`, "LASER RADIATION")
		s.text(x+56, 92, 10, `text-anchor="middle" font-weight="700"`, "CLASS "+l.class)
		pw := fmt.Sprintf("%g mW", l.mw)
		if l.mw >= 1000 {
			pw = fmt.Sprintf("%g W", l.mw/1000)
		}
		s.text(x+56, 110, 8.5, `text-anchor="middle"`, "Max output "+pw)
		markAt(s, x+100, 48, i)
	}
	q := Question{key: "ex/laserclass", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. The FDA limits laser pointers to 5 mW. Which pointer may be used in the training room?", ex)
	q.Choices, q.Correct = letterQ(r, "Pointer", 4, pos)
	return q
}

// ── NEC wire colors ─────────────────────────────────────────────────────

func wireColorsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	wires := []struct{ hex, name, role string }{
		{"#111", "black", "an ungrounded (hot) conductor"},
		{"#dc2626", "red", "an ungrounded (hot) conductor"},
		{"#f8fafc", "white", "the grounded (neutral) conductor"},
		{"#16a34a", "green", "the equipment grounding conductor"},
	}
	order := orderN(r, 4)
	s := frame(400, 210, "Open electrical box with four conductors", "OPEN JUNCTION BOX — 120/240 V CIRCUIT")
	s.rect(120, 40, 160, 150, 6, "#9ca3af", "#475569")
	s.rect(130, 50, 140, 130, 4, "#e5e7eb", "")
	for i, k := range order {
		x := 150 + 33*float64(i)
		s.raw(`<path d="M%.1f 60 C%.1f 100 %.1f 120 %.1f 170" stroke="#334155" stroke-width="8" fill="none"/>`, x, x+10, x-10, x)
		s.raw(`<path d="M%.1f 60 C%.1f 100 %.1f 120 %.1f 170" stroke="%s" stroke-width="6" fill="none"/>`, x, x+10, x-10, x, wires[k].hex)
		markAt(s, x, 196-4, i)
	}
	ask := []int{2, 3}[r.intn(2)]
	pos := 0
	for i, k := range order {
		if k == ask {
			pos = i
		}
	}
	q := Question{key: "ex/wirecolors", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using standard U.S. (NEC) color coding, which conductor is %s?", ex, wires[ask].role)
	q.Choices, q.Correct = letterQ(r, "Wire", 4, pos)
	return q
}

// ── breaker panel directory ─────────────────────────────────────────────

func breakerDirExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	loads := sample(r, []string{"Break room microwave", "Break room fridge", "Copier 3F", "Lobby lights", "Conference room A", "Server closet UPS", "Coffee maker", "Hall receptacles", "Exit signs", "Boardroom AV", "Vending machine", "Restroom hand dryers", "Mail room", "Desk row 3F-East", "Desk row 3F-West", "Parking lot lights"}, 12)
	s := frame(440, 260, "Breaker panel directory", "PANEL LP-3 — CIRCUIT DIRECTORY")
	s.rect(30, 34, 380, 214, 4, "#f8fafc", "#475569")
	for i, l := range loads {
		col, row := i%2, i/2
		x := 40 + 186*float64(col)
		y := 52 + 32*float64(row)
		n := 2*row + col + 1
		s.rect(x, y-12, 24, 20, 2, "#334155", "")
		s.text(x+12, y+2, 9, `text-anchor="middle" font-weight="700" fill="#fff"`, itoa(n))
		s.text(x+30, y+2, 9, ``, l)
	}
	t := r.intn(12)
	q := Question{key: "ex/breakerdir", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Before replacing a receptacle for the %s, which breaker should be locked out?", ex, strings.ToLower(loads[t]))
	q.Choices, q.Correct = numChoices(r, t+1, []int{1, -1, 2, -2, 3}, func(n int) string { return fmt.Sprintf("Breaker %d", n) })
	return q
}

// ── chemical storage compatibility ──────────────────────────────────────

func storageCompatExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	classes := []struct{ name, hex string }{
		{"Acids", "#dc2626"}, {"Bases", "#2563eb"}, {"Flammables", "#f97316"}, {"Oxidizers", "#facc15"},
	}
	badPairs := [][2]int{{0, 1}, {2, 3}}
	t := r.intn(4)
	s := frame(540, 210, "Four shelves with chemical containers", "CHEMICAL STORAGE — SHELF INSPECTION")
	for i := 0; i < 4; i++ {
		var a, b int
		if i == t {
			bp := pick(r, badPairs)
			a, b = bp[0], bp[1]
			if r.chance(0.5) {
				a, b = b, a
			}
		} else {
			a = r.intn(4)
			b = a
		}
		y := 48 + 40*float64(i)
		s.rect(60, y+28, 440, 5, 1, "#78716c", "")
		for k, cl := range []int{a, b, a, b} {
			x := 90 + 100*float64(k)
			s.rect(x, y, 60, 28, 4, "#fff", classes[cl].hex)
			s.text(x+30, y+17, 8.5, `text-anchor="middle" font-weight="700" fill="#111"`, classes[cl].name)
		}
		markAt(s, 40, y+16, i)
	}
	q := Question{key: "ex/storagecompat", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which shelf stores incompatible chemicals together?", ex)
	q.Choices, q.Correct = letterQ(r, "Shelf", 4, t)
	return q
}

// ── ladder duty ratings (ANSI A14) ──────────────────────────────────────

func ladderDutyExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	types := []struct {
		name string
		lb   int
	}{{"Type III", 200}, {"Type II", 225}, {"Type I", 250}, {"Type IA", 300}, {"Type IAA", 375}}
	show := sample(r, orderN(r, 5), 4)
	sortInts(show)
	order := orderN(r, 4)
	// Pick a load so the lightest adequate ladder is among those shown, and
	// at least one shown ladder is too light.
	var load, right int
	for {
		load = r.between(190, 360)
		right = -1
		for _, k := range show {
			if types[k].lb >= load {
				right = k
				break
			}
		}
		if right >= 0 && types[show[0]].lb < load {
			break
		}
	}
	s := frame(540, 190, "Four ladder duty rating labels", "LADDER ROOM — DUTY RATING LABELS")
	pos := 0
	for i, o := range order {
		k := show[o]
		if k == right {
			pos = i
		}
		x := 18 + 131*float64(i)
		s.rect(x, 46, 112, 120, 4, "#fff", "#111")
		s.rect(x+4, 50, 104, 22, 2, "#f97316", "")
		s.text(x+56, 65, 9, `text-anchor="middle" font-weight="700"`, "DUTY RATING")
		s.text(x+56, 98, 15, `text-anchor="middle" font-weight="700"`, types[k].name)
		s.text(x+56, 122, 10, `text-anchor="middle"`, fmt.Sprintf("%d lb", types[k].lb))
		s.text(x+56, 140, 7, `text-anchor="middle" fill="#475569"`, "Person + tools + materials")
		markAt(s, x+100, 44, i)
	}
	w := load - r.between(20, 50)
	q := Question{key: "ex/ladderduty", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. A technician weighing %d lb will carry %d lb of tools and materials. Which is the lowest-rated ladder they may use?", ex, w, load-w)
	q.Choices, q.Correct = letterQ(r, "Ladder", 4, pos)
	return q
}

// ── ANSI/ISEA 107 high-visibility classes ───────────────────────────────

func hiVisExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type job struct{ desc, class string }
	jobs := []job{
		{"collecting carts in the company parking lot, where traffic moves under 25 mph", "Class 1"},
		{"flagging traffic on a 35 mph city street", "Class 2"},
		{"surveying beside a 65 mph highway at night", "Class 3"},
	}
	j := pick(r, jobs)
	order := orderN(r, 3)
	s := frame(480, 200, "Three high-visibility garments", "HIGH-VISIBILITY GARMENTS — ANSI/ISEA 107")
	classes := []string{"Class 1", "Class 2", "Class 3"}
	pos := 0
	for i, k := range order {
		if classes[k] == j.class {
			pos = i
		}
		cx := 90 + 150*float64(i)
		switch k {
		case 0:
			s.rect(cx-30, 60, 60, 60, 6, "#d9f99d", "#65a30d")
			s.rect(cx-30, 92, 60, 6, 0, "#e5e7eb", "")
		case 1:
			s.raw(`<path d="M%.1f 52 h60 v84 h-60 Z" fill="#bef264" stroke="#65a30d"/>`, cx-30)
			s.rect(cx-30, 80, 60, 7, 0, "#e5e7eb", "")
			s.rect(cx-30, 110, 60, 7, 0, "#e5e7eb", "")
			s.rect(cx-12, 52, 6, 84, 0, "#e5e7eb", "")
			s.rect(cx+6, 52, 6, 84, 0, "#e5e7eb", "")
		case 2:
			s.raw(`<path d="M%.1f 52 h60 l26 50 l-12 6 l-14 -22 v52 h-60 v-52 l-14 22 l-12 -6 Z" fill="#bef264" stroke="#65a30d"/>`, cx-30)
			s.rect(cx-30, 80, 60, 7, 0, "#e5e7eb", "")
			s.rect(cx-30, 110, 60, 7, 0, "#e5e7eb", "")
			s.raw(`<path d="M%.1f 84 l20 6 M%.1f 84 l-20 6" stroke="#e5e7eb" stroke-width="5"/>`, cx+32, cx-32)
		}
		markAt(s, cx-44, 50, i)
	}
	q := Question{key: "ex/hivis", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. The garments are Class 1 (least coverage), Class 2, and Class 3 (sleeves, most coverage). Which garment is appropriate for a worker %s?", ex, j.desc)
	var wrongs []string
	for i := 0; i < 3; i++ {
		if i != pos {
			wrongs = append(wrongs, "Garment "+markLetters[i])
		}
	}
	q.Choices, q.Correct = withChoices(r, "Garment "+markLetters[pos], wrongs)
	return q
}

// ── quarter-turn valve handles ──────────────────────────────────────────

func valvePosExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	closed := r.chance(0.5)
	t := r.intn(4)
	s := frame(540, 220, "Four ball valves on pipes", "QUARTER-TURN BALL VALVES — LOCKOUT VERIFICATION")
	for i := 0; i < 4; i++ {
		y := 52 + 40*float64(i)
		s.rect(60, y, 440, 16, 8, "#94a3b8", "#475569")
		s.rect(250, y-6, 40, 28, 4, "#b45309", "#78350f")
		isClosed := (i == t) == closed
		if isClosed {
			s.rect(266, y-24, 8, 40, 2, "#dc2626", "#7f1d1d")
		} else {
			s.rect(230, y-12, 80, 8, 2, "#dc2626", "#7f1d1d")
		}
		markAt(s, 40, y+8, i)
	}
	q := Question{key: "ex/valvepos", Exhibit: s.done()}
	state := "open"
	if closed {
		state = "closed"
	}
	q.Text = fmt.Sprintf("Refer to %s. On a quarter-turn ball valve, a handle in line with the pipe is open and a handle across the pipe is closed. Which valve is %s?", ex, state)
	q.Choices, q.Correct = letterQ(r, "Valve", 4, t)
	return q
}
