package compliance

import (
	"fmt"
	"math"
	"strings"
)

// Aviation exhibits. Answers follow the real rules: runway numbers and
// headwind, engine intake and blast hazard areas, weight-and-balance
// envelopes, FAA lithium battery rules for passengers, 14 CFR 91.151 VFR
// fuel reserves, 121.542 sterile cockpit, Part 107's 400 ft AGL limit,
// nearest usable exit, and AWC flight categories from a METAR.

func init() {
	for k, g := range map[string]exhibitGen{
		"runway":     runwayExhibit,
		"rampzones":  rampZonesExhibit,
		"cgenvelope": cgEnvelopeExhibit,
		"lithium":    lithiumBagsExhibit,
		"fuelres":    fuelReserveExhibit,
		"sterile":    sterileCockpitExhibit,
		"droneagl":   droneAGLExhibit,
		"cabinexit":  cabinExitExhibit,
		"metar":      metarExhibit,
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
		"faa":           {"runway", "fuelres", "sterile", "metar", "cgenvelope"},
		"ramp":          {"rampzones", "runway", "lightning"},
		"cabin":         {"cabinexit", "lithium"},
		"dgr":           {"lithium"},
		"flightdept":    {"cgenvelope", "fuelres", "metar"},
		"sida":          {"badgezones", "rampzones"},
		"drones":        {"droneagl"},
		"travel":        {"lithium", "cabinexit"},
		"intltravel":    {"lithium"},
		"travelbooking": {"cabinexit"},
		"execcomp":      {"cgenvelope"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// bearing → screen point (north up).
func bearingXY(cx, cy, l float64, deg int) (float64, float64) {
	a := float64(deg) * math.Pi / 180
	return cx + l*math.Sin(a), cy - l*math.Cos(a)
}

func angDiff(a, b int) int {
	d := ((a-b)%360 + 360) % 360
	if d > 180 {
		d = 360 - d
	}
	return d
}

// ── runway selection: land and take off into the wind ───────────────────

func runwayExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	var n1, n2, wind int
	for {
		n1, n2 = r.between(1, 18), r.between(1, 18)
		if angDiff(n1*10, n2*10) < 50 || angDiff(n1*10, n2*10) > 130 {
			continue
		}
		wind = 10 * r.between(1, 36)
		ends := []int{n1, n1 + 18, n2, n2 + 18}
		best, second := 999, 999
		for _, e := range ends {
			d := angDiff(e*10, wind)
			if d < best {
				best, second = d, best
			} else if d < second {
				second = d
			}
		}
		if best <= 30 && second >= best+30 {
			break
		}
	}
	ends := []int{n1, n1 + 18, n2, n2 + 18}
	s := frame(440, 300, "Airport diagram with two runways and a wind indicator", "AIRPORT DIAGRAM — ACTIVE RUNWAY")
	cx, cy := 230.0, 165.0
	for _, n := range []int{n1, n2} {
		x1, y1 := bearingXY(cx, cy, 115, n*10)
		x2, y2 := bearingXY(cx, cy, 115, n*10+180)
		s.line(x1, y1, x2, y2, "#374151", 16)
		s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#fff" stroke-width="1.5" stroke-dasharray="8 8"/>`, x1, y1, x2, y2)
	}
	right := ""
	var names []string
	for _, e := range ends {
		// The number is painted at the end you start from, opposite its heading.
		x, y := bearingXY(cx, cy, 134, e*10+180)
		name := fmt.Sprintf("%02d", e)
		s.rect(x-13, y-9, 26, 16, 3, "#fff", "#374151")
		s.text(x, y+4, 9.5, `text-anchor="middle" font-weight="700"`, name)
		names = append(names, "Runway "+name)
		if angDiff(e*10, wind) <= 30 {
			right = "Runway " + name
		}
	}
	s.rect(14, 30, 104, 78, 4, "#f3f4f6", "#9ca3af")
	s.text(66, 44, 8, `text-anchor="middle" font-weight="700"`, "WIND")
	fx, fy := bearingXY(66, 76, 22, wind)
	tx, ty := bearingXY(66, 76, 18, wind+180)
	s.raw(`<defs><marker id="rw" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto"><path d="M0 0 L8 4 L0 8 Z" fill="#1d4ed8"/></marker></defs>`)
	s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1d4ed8" stroke-width="3" marker-end="url(#rw)"/>`, fx, fy, tx, ty)
	kt := r.between(8, 22)
	s.text(66, 104, 7.5, `text-anchor="middle"`, fmt.Sprintf("%03d° at %d kt", wind, kt))
	s.text(420, 40, 9, `text-anchor="end" font-weight="700" fill="#374151"`, "N ↑")
	q := Question{key: "ex/runway", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. The wind is from %03[2]d degrees at %[3]d knots. Which runway gives the best headwind for takeoff?",
		"Refer to %[1]s. Based on the wind shown, which runway is most likely in use for landing?",
	}), ex, wind, kt)
	q.Choices, q.Correct = oneOf(r, right, names)
	return q
}

// ── engine hazard areas on the ramp ─────────────────────────────────────

func rampZonesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(440, 350, "Top view of a twin-engine jet with intake and exhaust hazard areas and four ground crew positions", "RAMP — ENGINE HAZARD AREAS (ENGINES RUNNING)")
	cx, cy := 220.0, 172.0
	// hazard areas
	for _, ex := range []float64{-52, 52} {
		s.raw(`<path d="M%.1f %.1f A42 42 0 0 1 %.1f %.1f Z" fill="#fca5a5" fill-opacity=".55" stroke="#dc2626"/>`, cx+ex-42, cy-4, cx+ex+42, cy-4)
		s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#fdba74" fill-opacity=".5" stroke="#ea580c"/>`, cx+ex-10, cy+14, cx+ex+10, cy+14, cx+ex+28, cy+170, cx+ex-28, cy+170)
	}
	// aircraft
	s.raw(`<path d="M%.1f %.1f q10 0 12 24 v150 l-4 20 h-16 l-4 -20 v-150 q2 -24 12 -24z" fill="#e5e7eb" stroke="#374151"/>`, cx, cy-80)
	s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#d1d5db" stroke="#374151"/>`, cx-12, cy-20, cx-150, cy+20, cx+150, cy+20, cx+12, cy-20)
	s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#d1d5db" stroke="#374151"/>`, cx-8, cy+90, cx-50, cy+110, cx+50, cy+110, cx+8, cy+90)
	for _, ex := range []float64{-52, 52} {
		s.rect(cx+ex-9, cy-6, 18, 22, 4, "#6b7280", "#111")
	}
	s.circle(cx+150, cy+20, 3, "#16a34a", "")
	s.circle(cx-150, cy+20, 3, "#dc2626", "")
	s.text(20, 342, 7.5, `fill="#b91c1c"`, "Red = intake suction   Orange = jet blast")
	safe := [][2]float64{{cx, cy - 118}, {cx + 180, cy + 40}, {cx - 185, cy - 40}, {cx, cy + 150}}
	unsafe := [][2]float64{{cx + 52, cy - 30}, {cx - 52, cy + 110}, {cx - 52, cy - 28}, {cx + 52, cy + 90}}
	spots := sample(r, safe, 3)
	spots = append([][2]float64{pick(r, unsafe)}, spots...)
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		p := spots[o]
		s.circle(p[0], p[1], 6, "#1d4ed8", "#fff")
		markAt(s, p[0]+16, p[1]-10, i)
		if o == 0 {
			right = i
		}
	}
	q := Question{key: "ex/rampzones", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Both engines are running. Which ground crew position is inside an engine hazard area?",
		"Refer to %s. The anti-collision beacon is on and the engines are running. Which marked crew member must move immediately?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Position", 4, right)
	return q
}

// ── weight and balance envelope ─────────────────────────────────────────

func cgEnvelopeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// envelope in (CG inches, weight lb)
	env := [][2]float64{{82, 1500}, {82, 1950}, {86, 2550}, {93, 2550}, {93, 1500}}
	inside := func(cg, w float64) bool {
		if w < 1500 || w > 2550 || cg > 93 {
			return false
		}
		fwd := 82.0
		if w > 1950 {
			fwd = 82 + (w-1950)/600*4
		}
		return cg >= fwd
	}
	s := frame(440, 300, "Weight and center-of-gravity envelope chart with four plotted loadings", "WEIGHT & BALANCE — CG ENVELOPE")
	x := func(cg float64) float64 { return 60 + (cg-80)*26 }
	y := func(w float64) float64 { return 258 - (w-1400)*0.17 }
	for cg := 80.0; cg <= 94; cg += 2 {
		s.line(x(cg), y(1400), x(cg), y(2600), "#e5e7eb", 1)
		s.text(x(cg), 272, 7.5, `text-anchor="middle" fill="#555"`, fmt.Sprintf("%.0f", cg))
	}
	for w := 1400.0; w <= 2600; w += 200 {
		s.line(x(80), y(w), x(94), y(w), "#e5e7eb", 1)
		s.text(54, y(w)+3, 7.5, `text-anchor="end" fill="#555"`, commas(int(w)))
	}
	var pts []string
	for _, p := range env {
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", x(p[0]), y(p[1])))
	}
	s.raw(`<polygon points="%s" fill="#bfdbfe" fill-opacity=".6" stroke="#1d4ed8" stroke-width="2"/>`, strings.Join(pts, " "))
	s.text(240, 292, 8, `text-anchor="middle" fill="#333"`, "CG (inches aft of datum)")
	s.text(12, 150, 8, `fill="#333" transform="rotate(-90 12 150)" text-anchor="middle"`, "Gross weight (lb)")
	// one loading outside, three inside
	var load [][2]float64
	for len(load) < 3 {
		cg, w := 84+r.float()*8, 1600+r.float()*900
		if inside(cg, w) && inside(cg-0.8, w) && inside(cg+0.8, w) && inside(cg, w+60) && inside(cg, w-60) {
			load = append(load, [2]float64{cg, w})
		}
	}
	out := pick(r, [][2]float64{{93.8 + r.float(), 1700 + r.float()*600}, {83 + r.float(), 2350 + r.float()*150}, {88 + r.float()*3, 2620 + r.float()*20}, {81 + r.float()*0.6, 1600 + r.float()*200}})
	if out[1] > 2590 {
		out[1] = 2590
	}
	load = append([][2]float64{out}, load...)
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		p := load[o]
		s.circle(x(p[0]), y(p[1]), 4, "#c62828", "")
		markAt(s, x(p[0])+13, y(p[1])-9, i)
		if o == 0 {
			right = i
		}
	}
	q := Question{key: "ex/cgenvelope", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. The flight department plotted four possible loadings for tomorrow's trip. Which one is outside the approved envelope?",
		"Refer to %s. Which plotted loading may not be flown as planned?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Loading", 4, right)
	return q
}

// ── lithium batteries: carry-on only ────────────────────────────────────

func lithiumBagsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := pick(r, []string{"Power bank", "Spare camera battery", "Vape pen", "Spare laptop battery"})
	checked := sample(r, []string{"Shoes", "Hair dryer", "Paperback novel", "Umbrella", "Toiletry bag", "Folded suits", "Conference swag"}, 3)
	carry := sample(r, []string{"Laptop", "Phone charger", "Noise-cancelling headphones", "Tablet", "Passport"}, 3)
	all := append([]string{bad}, checked...)
	shuffle(r, all)
	s := frame(460, 250, "A carry-on bag and a checked bag with their packed items", "PACKING CHECK — "+strings.ToUpper(sl["FIRST"])+"'S TRIP")
	bag := func(x float64, title, fill string, items []string) {
		s.rect(x, 50, 190, 170, 14, fill, "#374151")
		s.rect(x+70, 38, 50, 14, 5, "none", "#374151")
		s.text(x+95, 70, 9, `text-anchor="middle" font-weight="700"`, title)
		for i, it := range items {
			s.rect(x+16, 82+float64(i)*32, 158, 24, 5, "#fff", "#9ca3af")
			s.text(x+95, 98+float64(i)*32, 8.5, `text-anchor="middle"`, it)
		}
	}
	bag(24, "CARRY-ON", "#dbeafe", carry)
	bag(246, "CHECKED BAG", "#fef3c7", all)
	s.text(230, 240, 7.5, `text-anchor="middle" fill="#666"`, "Laptop, phone, and tablet are packed switched off.")
	q := Question{key: "ex/lithium", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which item must be moved out of the checked bag before %s's flight?",
		"Refer to %s. The airline asks to gate-check %s's carry-on. Which item in the checked bag already breaks the rules?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, bad, checked)
	return q
}

// ── VFR fuel reserves (91.151: 30 min day, 45 min night) ────────────────

func fuelReserveExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(460, 250, "Four planned flights with flight time and fuel endurance bars", "FUEL PLANNING — VFR FLIGHTS")
	k := 1.3 // px per minute
	for i := 0; i < 4; i++ {
		night := r.chance(0.5)
		need := 30
		if night {
			need = 45
		}
		ft := 5 * r.between(10, 28)
		res := need + 5*r.between(2, 8)
		if i == bad {
			if night && r.chance(0.6) {
				res = 5 * r.between(7, 8) // 35-40: enough for day, not night
			} else {
				res = 5 * r.between(2, need/5-1)
			}
		}
		y := 44 + float64(i)*48
		cond := "Day VFR"
		if night {
			cond = "Night VFR"
		}
		s.text(20, y+10, 9, `font-weight="700" fill="#0a1628"`, "FLIGHT "+markLetters[i])
		s.text(20, y+24, 7.5, `fill="#555"`, cond)
		s.rect(100, y, float64(ft)*k, 13, 2, "#1e3a8a", "")
		s.text(104+float64(ft)*k, y+10, 7.5, `fill="#1e3a8a"`, fmt.Sprintf("flight %d min", ft))
		s.rect(100, y+17, float64(ft+res)*k, 13, 2, "#c9a227", "")
		s.text(104+float64(ft+res)*k, y+27, 7.5, `fill="#8a6d1a"`, fmt.Sprintf("fuel %d min", ft+res))
	}
	s.text(230, 240, 7.5, `text-anchor="middle" fill="#666"`, "Endurance at normal cruising speed. All four are airplanes.")
	q := Question{key: "ex/fuelres", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which planned flight does not carry the minimum VFR fuel reserve?",
		"Refer to %s. The chief pilot rejects one of these flight plans for fuel. Which one?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Flight", 4, bad)
	return q
}

// ── sterile cockpit below 10,000 ft ─────────────────────────────────────

func sterileCockpitExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	topics := sample(r, []string{"Fantasy football draft", "Holiday party gossip", "Best airport burritos", "Mortgage rates", "The CEO's new boat", "Weekend golf scores", "Reality TV finale"}, 4)
	s := frame(470, 250, "Flight altitude profile with four marked conversations", "CREW COMMUNICATION LOG — ALTITUDE PROFILE")
	x := func(t float64) float64 { return 40 + t*4 }    // minutes 0..100
	y := func(a float64) float64 { return 220 - a/200 } // feet 0..36,000
	cruise := float64(r.between(28, 37) * 1000)
	alt := func(t float64) float64 {
		switch {
		case t < 25:
			return cruise * t / 25
		case t < 70:
			return cruise
		default:
			return math.Max(0, cruise*(100-t)/30)
		}
	}
	s.line(x(0), y(0), x(100), y(0), "#555", 1.5)
	for _, a := range []float64{10000, 20000, 30000} {
		s.dash(x(0), y(a), x(100), y(a), "#d1d5db", 1)
		s.text(x(0)-4, y(a)+3, 7, `text-anchor="end" fill="#555"`, commas(int(a)))
	}
	var d []string
	for t := 0.0; t <= 100; t += 2 {
		d = append(d, fmt.Sprintf("%.1f,%.1f", x(t), y(alt(t))))
	}
	s.raw(`<polyline points="%s" fill="none" stroke="#1e3a8a" stroke-width="2.5"/>`, strings.Join(d, " "))
	s.text(x(100), 236, 7.5, `text-anchor="end" fill="#555"`, "Minutes after departure")
	// times: one in climb/descent below 10,000; others at cruise or above 10,000
	var ts []float64
	badT := pick(r, []float64{float64(r.between(2, 5)), float64(r.between(93, 97))})
	for len(ts) < 3 {
		t := float64(r.between(30, 68))
		if r.chance(0.3) {
			t = float64(r.between(14, 22))
		}
		ok := alt(t) > 12000
		for _, u := range ts {
			if math.Abs(u-t) < 12 {
				ok = false
			}
		}
		if ok {
			ts = append(ts, t)
		}
	}
	ts = append([]float64{badT}, ts...)
	for i, t := range ts {
		px, py := x(t), y(alt(t))
		s.circle(px, py, 4, "#c62828", "")
		ly := py - 16 - float64(i%2)*22
		if ly < 36 {
			ly = py + 22 + float64(i%2)*22
		}
		anchor := "middle"
		if px > 400 {
			anchor = "end"
		} else if px < 80 {
			anchor = "start"
		}
		s.text(px, ly, 7.5, `text-anchor="`+anchor+`" font-weight="700"`, topics[i])
		s.text(px, ly+10, 7, `text-anchor="`+anchor+`" fill="#555"`, fmt.Sprintf("%s ft", commas(int(math.Round(alt(t)/100)*100))))
	}
	q := Question{key: "ex/sterile", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Under the sterile cockpit rule, which crew conversation was not allowed?",
		"Refer to %s. The cockpit voice recorder review found four off-topic conversations. Which one happened during a critical phase of flight?",
	}), ex)
	q.Choices, q.Correct = withChoices(r, topics[0], topics[1:])
	return q
}

// ── drone altitude: 400 ft AGL, given in MSL ────────────────────────────

func droneAGLExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ground := 10 * r.between(55, 110)
	bad := r.intn(4)
	s := frame(440, 260, "Side view of four drones with altitude readouts over open ground", "DRONE TELEMETRY — OPEN FIELD SURVEY")
	s.rect(1, 222, 438, 37, 0, "#d9f99d", "")
	s.line(1, 222, 439, 222, "#3f6212", 2)
	s.text(16, 244, 8, `font-weight="700" fill="#3f6212"`, fmt.Sprintf("Ground elevation: %s ft MSL", commas(ground)))
	for i := 0; i < 4; i++ {
		agl := 10 * r.between(12, 38)
		if i == bad {
			agl = 10 * r.between(43, 52)
		}
		x := 50 + float64(i)*96
		yy := 222 - float64(agl)*0.36
		s.dash(x, yy+8, x, 222, "#9ca3af", 1)
		s.rect(x-12, yy-3, 24, 6, 2, "#374151", "")
		for _, dx := range []float64{-16, 16} {
			s.raw(`<ellipse cx="%.1f" cy="%.1f" rx="9" ry="2.5" fill="#9ca3af"/>`, x+dx, yy-5)
		}
		s.rect(x+18, yy-8, 62, 24, 3, "#111827", "")
		s.text(x+49, yy+1, 7, `text-anchor="middle" fill="#a7f3d0"`, "DRONE "+markLetters[i])
		s.text(x+49, yy+12, 8, `text-anchor="middle" font-weight="700" fill="#a7f3d0"`, commas(ground+agl)+" MSL")
	}
	q := Question{key: "ex/droneagl", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. No structures are nearby. Which drone is flying above the Part 107 altitude limit?",
		"Refer to %s. The remote pilot in command must bring one drone down right away. Which one?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Drone", 4, bad)
	return q
}

// ── nearest usable cabin exit ───────────────────────────────────────────

func cabinExitExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	exits := []int{0, 10, 20, 31} // the row each exit sits in front of / after
	names := []string{"Forward exit", "Row 10 exit", "Row 20 exit", "Rear exit"}
	blocked := r.intn(4)
	var row int
	for {
		row = r.between(2, 29)
		best, second, cnt := 999, 999, 0
		for i, e := range exits {
			if i == blocked {
				continue
			}
			d := row - e
			if d < 0 {
				d = -d
			}
			if d < best {
				best, second = d, best
				cnt = 1
			} else if d == best {
				cnt++
			} else if d < second {
				second = d
			}
		}
		if cnt == 1 && second-best >= 3 {
			break
		}
	}
	seat := fmt.Sprintf("%d%s", row, pick(r, []string{"A", "C", "D", "F"}))
	s := frame(470, 200, "Cabin seat map with four exits, one blocked by fire", "CABIN SEAT MAP — EXITS")
	x := func(row int) float64 { return 50 + float64(row)*12.5 }
	s.raw(`<path d="M30 60 Q14 100 30 140 L450 140 Q462 100 450 60 Z" fill="#f8fafc" stroke="#475569" stroke-width="1.5"/>`)
	for rr := 1; rr <= 30; rr++ {
		for k, yy := range []float64{66, 76, 86, 112, 122, 132} {
			fill := "#cbd5e1"
			if rr == row && ((k == 0 && strings.HasSuffix(seat, "A")) || (k == 2 && strings.HasSuffix(seat, "C")) || (k == 3 && strings.HasSuffix(seat, "D")) || (k == 5 && strings.HasSuffix(seat, "F"))) {
				fill = "#1d4ed8"
			}
			s.rect(x(rr)-5, yy-4, 9, 8, 1.5, fill, "")
		}
		if rr%5 == 0 {
			s.text(x(rr), 104, 6.5, `text-anchor="middle" fill="#64748b"`, fmt.Sprint(rr))
		}
	}
	right := ""
	for i, e := range exits {
		ex := x(e) + 6
		if e == 0 {
			ex = x(0) + 2
		}
		s.rect(ex-6, 52, 12, 8, 1, "#16a34a", "")
		s.rect(ex-6, 140, 12, 8, 1, "#16a34a", "")
		s.text(ex, 168, 7.5, `text-anchor="middle" font-weight="700" fill="#166534"`, names[i])
		if i == blocked {
			s.raw(`<path d="M%.1f 46 q4 -14 8 -4 q4 -10 8 4 z" fill="#f97316"/><path d="M%.1f 154 q4 14 8 4 q4 10 8 -4 z" fill="#f97316"/>`, ex-8, ex-8)
			s.text(ex, 180, 7, `text-anchor="middle" fill="#c2410c"`, "fire outside")
		}
	}
	best := 999
	for i, e := range exits {
		if i == blocked {
			continue
		}
		d := row - e
		if d < 0 {
			d = -d
		}
		if d < best {
			best, right = d, names[i]
		}
	}
	s.text(460, 194, 7.5, `text-anchor="end" fill="#1d4ed8" font-weight="700"`, "Blue = seat "+seat)
	q := Question{key: "ex/cabinexit", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. The crew calls for an evacuation and there is fire outside one exit. Which exit should the passenger in seat %[2]s use?",
		"Refer to %[1]s. %[3]s is in seat %[2]s. Which is their nearest usable exit?",
	}), ex, seat, sl["FIRST"])
	q.Choices, q.Correct = oneOf(r, right, names)
	return q
}

// ── METAR flight category ───────────────────────────────────────────────

func metarExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	stations := []string{"KMKE", "KORD", "KMSP", "KDEN", "KBOS", "KSEA", "KATL", "KDFW", "KSFO", "KPHX"}
	type vis struct {
		txt string
		sm  float64
	}
	vs := pick(r, []vis{{"10SM", 10}, {"7SM", 7}, {"6SM", 6}, {"5SM", 5}, {"4SM", 4}, {"3SM", 3}, {"2SM", 2}, {"1 1/2SM", 1.5}, {"1SM", 1}, {"1/2SM", 0.5}})
	var layers []string
	ceiling := -1
	h := r.between(3, 12)
	for i := 0; i < r.between(1, 3); i++ {
		cov := pick(r, []string{"FEW", "SCT", "BKN", "OVC"})
		layers = append(layers, fmt.Sprintf("%s%03d", cov, h))
		if (cov == "BKN" || cov == "OVC") && ceiling < 0 {
			ceiling = h * 100
		}
		if cov == "OVC" {
			break
		}
		h += r.between(5, 60)
	}
	if r.chance(0.3) {
		layers = []string{"SKC"}
		ceiling = -1
	}
	cat := func() string {
		switch {
		case (ceiling >= 0 && ceiling < 500) || vs.sm < 1:
			return "LIFR"
		case (ceiling >= 0 && ceiling < 1000) || vs.sm < 3:
			return "IFR"
		case (ceiling >= 0 && ceiling <= 3000) || vs.sm <= 5:
			return "MVFR"
		}
		return "VFR"
	}()
	wx := ""
	if vs.sm < 6 {
		wx = pick(r, []string{" BR", " -RA", " FG", " HZ", " -SN"})
		if vs.sm >= 1 && wx == " FG" {
			wx = " BR"
		}
	}
	t := r.between(-5, 28)
	dp := t - r.between(0, 8)
	tmp := func(v int) string {
		if v < 0 {
			return fmt.Sprintf("M%02d", -v)
		}
		return fmt.Sprintf("%02d", v)
	}
	metar := fmt.Sprintf("%s %02d%02d%02dZ %03d%02dKT %s%s %s %s/%s A%d", pick(r, stations), r.between(1, 28), r.between(0, 23), 5*r.intn(12), 10*r.between(1, 36), r.between(3, 18), vs.txt, wx, strings.Join(layers, " "), tmp(t), tmp(dp), r.between(2960, 3040))
	s := frame(460, 170, "A METAR weather report with a decoding key", "AVIATION ROUTINE WEATHER REPORT (METAR)")
	s.rect(16, 34, 428, 40, 3, "#111827", "")
	s.text(230, 59, 10.5, `text-anchor="middle" fill="#a7f3d0" font-family="Courier New,monospace" font-weight="700"`, metar)
	key := []string{
		"SM = visibility in statute miles · BKN/OVC = broken/overcast (a ceiling) · FEW/SCT = not a ceiling",
		"Cloud heights are in hundreds of feet above ground: OVC008 = overcast at 800 ft",
		"VFR: ceiling > 3,000 ft and vis > 5 SM · MVFR: 1,000–3,000 ft or 3–5 SM",
		"IFR: 500 to below 1,000 ft or 1 to below 3 SM · LIFR: below 500 ft or below 1 SM",
	}
	for i, k := range key {
		s.text(18, 96+float64(i)*16, 7.5, `fill="#374151"`, k)
	}
	q := Question{key: "ex/metar", Exhibit: s.done()}
	if ceiling >= 0 && r.chance(0.35) {
		q.Text = fmt.Sprintf("Refer to %s. What is the ceiling in this report?", ex)
		q.Choices, q.Correct = numChoices(r, ceiling, []int{-100, 100, ceiling * 9, -ceiling / 2, 1000}, func(n int) string { return commas(n) + " ft" })
		return q
	}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. What flight category does this report indicate?",
		"Refer to %s. Before the corporate jet departs, the flight department checks the destination weather. What flight category is it?",
	}), ex)
	q.Choices, q.Correct = oneOf(r, cat, []string{"VFR", "MVFR", "IFR", "LIFR"})
	return q
}
