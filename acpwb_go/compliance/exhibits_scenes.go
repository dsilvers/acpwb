package compliance

import (
	"fmt"
	"math"
	"strings"
)

// One-off scene exhibits. Rules from standards are the real ones: FDA Food
// Code 3-501.14 two-stage cooling, OSHA computer workstation eTool, NIOSH
// lifting guidance, OSHA 1926 Subpart P Appendix B slopes, forklift data
// plates (1910.178), sharps containers (replace at the fill line), FMCSA
// truck No-Zones, rack load plaques, a typical fall clearance calculation,
// and IBM card (Hollerith) punch codes.

func init() {
	registerExhibits(map[string]exhibitGen{
		"coolingcurve":  coolingCurveExhibit,
		"monitorheight": monitorHeightExhibit,
		"liftposture":   liftPostureExhibit,
		"dinnertables":  dinnerTablesExhibit,
		"confbadges":    confBadgesExhibit,
		"trenchslope":   trenchSlopeExhibit,
		"forkliftplate": forkliftPlateExhibit,
		"hotaisle":      hotAisleExhibit,
		"patentfig":     patentFigExhibit,
		"reachzones":    reachZonesExhibit,
		"punchcard":     punchCardExhibit,
		"sitetraffic":   siteTrafficExhibit,
		"gfsbackup":     gfsBackupExhibit,
		"annunciator":   annunciatorExhibit,
		"calltree":      callTreeExhibit,
		"sharpsbox":     sharpsBoxExhibit,
		"blindspots":    blindSpotsExhibit,
		"rackload":      rackLoadExhibit,
		"fallclearance": fallClearanceExhibit,
		"chairfit":      chairFitExhibit,
		"headcount":     headcountExhibit,
	}, map[string][]string{
		"foodsafety":     {"coolingcurve"},
		"kitchen":        {"coolingcurve"},
		"catering":       {"coolingcurve", "dinnertables"},
		"potluck":        {"coolingcurve"},
		"ergonomics":     {"monitorheight", "reachzones", "chairfit"},
		"eyestrain":      {"monitorheight"},
		"hotdesk":        {"monitorheight", "chairfit"},
		"remote":         {"chairfit"},
		"lifting":        {"liftposture", "rackload"},
		"injuries":       {"liftposture", "sharpsbox"},
		"warehouse":      {"liftposture", "forkliftplate", "rackload"},
		"mail":           {"liftposture"},
		"conferences":    {"dinnertables", "confbadges"},
		"events":         {"dinnertables", "confbadges"},
		"celebrations":   {"dinnertables"},
		"media":          {"confbadges"},
		"construction":   {"trenchslope", "fallclearance"},
		"fieldwork":      {"trenchslope"},
		"forklift":       {"forkliftplate", "rackload"},
		"serverroom":     {"hotaisle", "gfsbackup"},
		"sustainability": {"hotaisle"},
		"ip":             {"patentfig"},
		"patents":        {"patentfig"},
		"innovation":     {"patentfig"},
		"retro":          {"punchcard"},
		"y2k":            {"punchcard"},
		"records":        {"punchcard", "gfsbackup"},
		"scraping":       {"sitetraffic"},
		"botpolicy":      {"sitetraffic"},
		"ai":             {"sitetraffic"},
		"backups":        {"gfsbackup"},
		"continuity":     {"gfsbackup", "calltree"},
		"fire":           {"annunciator"},
		"evacassist":     {"annunciator", "headcount"},
		"closures":       {"calltree"},
		"weather":        {"calltree", "headcount"},
		"bloodborne":     {"sharpsbox"},
		"medical":        {"sharpsbox"},
		"firstaid":       {"sharpsbox"},
		"driving":        {"blindspots"},
		"commuting":      {"blindspots"},
		"deliveries":     {"blindspots"},
		"fallprotection": {"fallclearance"},
		"ladder":         {"fallclearance"},
		"safety":         {"headcount", "annunciator"},
	})
}

// ── FDA two-stage cooling ───────────────────────────────────────────────

func coolingCurveExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	foods := sample(r, []string{"Chili", "Rice", "Chicken stock", "Mashed potatoes", "Marinara", "Beef stew", "Pinto beans"}, 4)
	cols := []string{"#1d4ed8", "#dc2626", "#16a34a", "#7c3aed"}
	bad := r.intn(4)
	firstFail := r.chance(0.5)
	type cv struct{ t70, t41 float64 }
	cvs := make([]cv, 4)
	for i := range cvs {
		t70 := float64(r.between(5, 7)) / 4 // 1.25–1.75 h
		t41 := t70 + float64(r.between(8, 16))/4
		if t41 > 5.75 {
			t41 = 5.75
		}
		if i == bad {
			if firstFail {
				t70 = float64(r.between(10, 12)) / 4 // 2.5–3 h
				t41 = math.Min(t70+2.5, 5.75)
			} else {
				t41 = float64(r.between(27, 30)) / 4 // 6.75–7.5 h
			}
		}
		cvs[i] = cv{t70, t41}
	}
	s := frame(540, 290, "Cooling log chart for four foods", "COOLING LOG — CAFETERIA KITCHEN")
	x0, x1, y0, y1 := 60.0, 520.0, 50.0, 240.0
	X := func(h float64) float64 { return x0 + (x1-x0)*h/8 }
	Y := func(f float64) float64 { return y1 - (y1-y0)*(f-30)/(140-30) }
	for h := 0; h <= 8; h++ {
		s.line(X(float64(h)), y0, X(float64(h)), y1, "#e5e7eb", 1)
		s.text(X(float64(h)), y1+13, 8, `text-anchor="middle" fill="#475569"`, fmt.Sprintf("%dh", h))
	}
	for _, f := range []float64{41, 70, 135} {
		s.dash(x0, Y(f), x1, Y(f), "#dc2626", 1)
		s.text(x0-4, Y(f)+3, 8, `text-anchor="end" font-weight="700" fill="#dc2626"`, fmt.Sprintf("%.0f°F", f))
	}
	s.dash(X(2), y0, X(2), y1, "#64748b", 1)
	s.dash(X(6), y0, X(6), y1, "#64748b", 1)
	for i, cvv := range cvs {
		pts := fmt.Sprintf("%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f", X(0), Y(135), X(cvv.t70), Y(70), X(cvv.t41), Y(41), X(8), Y(38))
		s.raw(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2.2"/>`, pts, cols[i])
		s.rect(x0+10+112*float64(i), 30, 10, 8, 1, cols[i], "")
		s.text(x0+24+112*float64(i), 37.5, 8.5, `font-weight="700"`, markLetters[i]+": "+foods[i])
	}
	s.text(x0, 280, 8, `fill="#334155"`, "Time measured from 135°F. Dashed lines mark 70°F and 41°F, and the 2-hour and 6-hour marks.")
	q := Question{key: "ex/coolingcurve", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. The FDA Food Code requires cooling from 135°F to 70°F within 2 hours, and to 41°F within 6 hours total. Which batch failed?",
		"Refer to %[1]s. Which batch must be reheated or discarded because it did not cool fast enough?",
	}), ex)
	right := fmt.Sprintf("%s (%s)", markLetters[bad], foods[bad])
	var wrongs []string
	for i := range foods {
		if i != bad {
			wrongs = append(wrongs, fmt.Sprintf("%s (%s)", markLetters[i], foods[i]))
		}
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── monitor height ──────────────────────────────────────────────────────

func monitorHeightExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(540, 210, "Four workstations, side view", "WORKSTATION SETUP — SIDE VIEW")
	for i := 0; i < 4; i++ {
		x := 14 + 131*float64(i)
		s.rect(x, 32, 122, 168, 4, "#f8fafc", "#e2e8f0")
		eye := 92.0
		top := eye + float64(r.between(1, 3))*3 // top of screen at or a little below eye level
		if i == bad {
			top = eye - float64(r.between(5, 8))*3
		}
		// person
		s.circle(x+30, eye-4, 11, "#fde68a", "#111")
		s.circle(x+36, eye-5, 1.6, "#111", "")
		s.line(x+28, eye+7, x+24, 160, "#111", 3)
		s.line(x+24, 160, x+50, 160, "#111", 3)
		s.line(x+50, 160, x+50, 190, "#111", 3)
		// desk + monitor
		s.rect(x+56, 150, 62, 5, 1, "#a16207", "")
		s.line(x+96, 150, x+96, top+40, "#334155", 3)
		s.rect(x+92, top, 7, 44, 1, "#111827", "")
		s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#2563eb" stroke-width="1" stroke-dasharray="3 2"/>`, x+36, eye-5, x+118, eye-5)
		markAt(s, x+12, 44, i)
	}
	s.text(14, 206, 7.5, `fill="#2563eb"`, "Dashed line = eye level")
	q := Question{key: "ex/monitorheight", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. OSHA's workstation guidance places the top of the screen at or slightly below eye level. Which setup has the monitor too high?", ex)
	q.Choices, q.Correct = letterQ(r, "Setup", 4, bad)
	return q
}

// ── lifting posture ─────────────────────────────────────────────────────

func liftPostureExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// Pose 0 is correct: knees bent, back straight, load close.
	order := orderN(r, 4)
	s := frame(540, 200, "Four people lifting a box", "LIFTING TECHNIQUE")
	pos := 0
	for i, k := range order {
		if k == 0 {
			pos = i
		}
		cx := 72 + 132*float64(i)
		fy := 180.0
		switch k {
		case 0: // squat, straight back, box between knees
			s.circle(cx-4, 96, 9, "#fde68a", "#111")
			s.line(cx-6, 105, cx-10, 140, "#111", 3)
			s.line(cx-10, 140, cx+12, 150, "#111", 3)
			s.line(cx+12, 150, cx+4, fy, "#111", 3)
			s.line(cx-6, 112, cx+10, 140, "#111", 3)
			s.rect(cx+6, 136, 24, 22, 2, "#d6a15b", "#7c5320")
		case 1: // stiff legs, rounded back
			s.raw(`<path d="M%.1f 110 Q%.1f 100 %.1f 140" stroke="#111" stroke-width="3" fill="none"/>`, cx-14, cx+10, cx+20)
			s.circle(cx+24, 146, 9, "#fde68a", "#111")
			s.line(cx-14, 110, cx-14, fy, "#111", 3)
			s.line(cx+10, 112, cx+30, 160, "#111", 3)
			s.rect(cx+22, 158, 24, 22, 2, "#d6a15b", "#7c5320")
		case 2: // twisting while lifting
			s.circle(cx, 92, 9, "#fde68a", "#111")
			s.line(cx, 101, cx-4, 140, "#111", 3)
			s.line(cx-4, 140, cx-14, fy, "#111", 3)
			s.line(cx-4, 140, cx+8, fy, "#111", 3)
			s.line(cx, 110, cx+30, 128, "#111", 3)
			s.rect(cx+26, 118, 24, 22, 2, "#d6a15b", "#7c5320")
			s.raw(`<path d="M%.1f 84 a16 8 0 1 1 -2 0" stroke="#dc2626" stroke-width="1.6" fill="none" marker-end=""/>`, cx-14)
		case 3: // load held far from body
			s.circle(cx-16, 92, 9, "#fde68a", "#111")
			s.line(cx-16, 101, cx-16, 140, "#111", 3)
			s.line(cx-16, 140, cx-22, fy, "#111", 3)
			s.line(cx-16, 140, cx-8, fy, "#111", 3)
			s.line(cx-16, 110, cx+30, 112, "#111", 3)
			s.rect(cx+28, 100, 24, 22, 2, "#d6a15b", "#7c5320")
		}
		s.line(cx-50, fy, cx+50, fy, "#94a3b8", 2)
		markAt(s, cx-48, 44, i)
	}
	q := Question{key: "ex/liftposture", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which person is using safe lifting technique?",
		"Refer to %[1]s. Which lift keeps the back straight, bends the knees, and holds the load close?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Person", 4, pos)
	return q
}

// ── PERCH dinner tables ─────────────────────────────────────────────────

func dinnerTablesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := 10
	bad := r.intn(6)
	s := frame(540, 290, "Banquet seating chart with six tables", "PERCH AWARDS DINNER — SEATING CHART")
	for i := 0; i < 6; i++ {
		cx := 100 + 170*float64(i%3)
		cy := 92 + 112*float64(i/3)
		n := r.between(6, limit)
		if i == bad {
			n = r.between(limit+1, limit+3)
		}
		s.circle(cx, cy, 28, "#fef3c7", "#a16207")
		s.text(cx, cy+4, 10, `text-anchor="middle" font-weight="700"`, fmt.Sprintf("Table %d", i+1))
		for k := 0; k < n; k++ {
			a := 2 * math.Pi * float64(k) / float64(n)
			s.circle(cx+40*math.Cos(a), cy+40*math.Sin(a), 6, "#1d4ed8", "")
		}
	}
	s.text(20, 280, 8.5, `fill="#334155"`, "● = one seated guest. Venue rule: no more than 10 guests per round table.")
	q := Question{key: "ex/dinnertables", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which table breaks the venue's seating limit?", ex)
	q.Choices, q.Correct = numChoices(r, bad+1, []int{1, -1, 2, -2, 3, -3}, func(n int) string {
		if n < 1 || n > 6 {
			return fmt.Sprintf("Table %d (not shown)", n)
		}
		return fmt.Sprintf("Table %d", n)
	})
	return q
}

// ── conference badges ───────────────────────────────────────────────────

func confBadgesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	types := []struct{ role, hex string }{{"SPEAKER", "#dc2626"}, {"PRESS", "#2563eb"}, {"ATTENDEE", "#64748b"}, {"STAFF", "#16a34a"}, {"EXHIBITOR", "#a16207"}}
	rooms := []struct {
		room string
		ok   []string
	}{
		{"the press room", []string{"PRESS", "STAFF"}},
		{"the speaker ready room", []string{"SPEAKER", "STAFF"}},
		{"the exhibit hall before it opens", []string{"EXHIBITOR", "STAFF"}},
	}
	rm := pick(r, rooms)
	var show []int
	for {
		show = sample(r, []int{0, 1, 2, 3, 4}, 4)
		n := 0
		for _, k := range show {
			if contains(rm.ok, types[k].role) {
				n++
			}
		}
		if n == 1 {
			break
		}
	}
	names := people(r, 4)
	s := frame(540, 200, "Four conference badges", "PERCH CONFERENCE BADGES")
	pos := 0
	for i, k := range show {
		if contains(rm.ok, types[k].role) {
			pos = i
		}
		x := 22 + 130*float64(i)
		s.rect(x, 40, 106, 140, 6, "#fff", "#94a3b8")
		s.rect(x+38, 46, 30, 6, 3, "#cbd5e1", "")
		s.text(x+53, 76, 9, `text-anchor="middle" font-weight="700" fill="#0a1628"`, "PERCH")
		s.text(x+53, 106, 9, `text-anchor="middle" font-weight="700"`, firstOf(names[i]))
		s.text(x+53, 120, 8, `text-anchor="middle" fill="#475569"`, strings.TrimPrefix(names[i], firstOf(names[i])+" "))
		s.rect(x, 146, 106, 34, 0, types[k].hex, "")
		s.text(x+53, 168, 10, `text-anchor="middle" font-weight="700" fill="#fff"`, types[k].role)
		markAt(s, x+8, 48, i)
	}
	q := Question{key: "ex/confbadges", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Access to %s is limited to %s badges. Which badge holder may enter?", ex, rm.room, strings.ToLower(strings.Join(rm.ok, " and ")))
	q.Choices, q.Correct = letterQ(r, "Badge", 4, pos)
	return q
}

// ── OSHA maximum allowable slopes (1926 Subpart P, App. B) ──────────────

func trenchSlopeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	soils := []struct {
		name string
		h    float64 // horizontal per 1 vertical
	}{{"Type A", 0.75}, {"Type B", 1}, {"Type C", 1.5}}
	soil := soils[r.intn(3)]
	ratios := []float64{0.5, 0.75, 1, 1.5, 2}
	right := r.intn(4)
	var shown []float64
	for {
		shown = sample(r, ratios, 4)
		n := 0
		for _, h := range shown {
			if h >= soil.h {
				n++
			}
		}
		// exactly one adequate slope, or exactly one too steep
		if n == 1 || n == 3 {
			break
		}
	}
	oneOK := false
	n := 0
	for _, h := range shown {
		if h >= soil.h {
			n++
		}
	}
	oneOK = n == 1
	for i, h := range shown {
		if (h >= soil.h) == oneOK {
			right = i
		}
	}
	s := frame(540, 190, "Four sloped trench cross-sections", "EXCAVATION CROSS-SECTIONS — 8 FT DEEP")
	for i, h := range shown {
		x := 14 + 131*float64(i)
		s.rect(x, 40, 122, 140, 4, "#f8fafc", "#e2e8f0")
		depth := 60.0
		run := depth * h * 0.42
		gy := 70.0
		mid := x + 61
		s.raw(`<path d="M%.1f %.1f H%.1f L%.1f %.1f H%.1f L%.1f %.1f H%.1f V176 H%.1f Z" fill="#d6c4a8" stroke="#8b6b3e"/>`,
			x+2, gy, mid-12-run, mid-12, gy+depth, mid+12, mid+12+run, gy, x+120, x+2)
		s.text(mid, gy+depth+16, 9, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, fmt.Sprintf("%g : 1 (H:V)", h))
		markAt(s, x+12, 52, i)
	}
	q := Question{key: "ex/trenchslope", Exhibit: s.done()}
	ask := "Which excavation is sloped enough?"
	if !oneOK {
		ask = "Which excavation is sloped too steeply?"
	}
	q.Text = fmt.Sprintf("Refer to %s. The soil is classified %s. Using OSHA's maximum allowable slopes (Type A ¾:1, Type B 1:1, Type C 1½:1), %s", ex, soil.name, lowerFirst(ask))
	q.Choices, q.Correct = letterQ(r, "Excavation", 4, right)
	return q
}

// ── forklift data plate ─────────────────────────────────────────────────

func forkliftPlateExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	capLb := pick(r, []int{3000, 4000, 5000, 6000})
	lc := 24
	right := r.intn(4)
	type load struct{ lb, depth int }
	loads := make([]load, 4)
	for i := range loads {
		if i == right {
			loads[i] = load{capLb - r.between(2, 10)*50, 2 * r.between(18, 24)}
			continue
		}
		if r.chance(0.5) {
			loads[i] = load{capLb + r.between(2, 16)*50, 2 * r.between(18, 24)} // too heavy
		} else {
			loads[i] = load{capLb - r.between(2, 10)*50, 2 * r.between(30, 36)} // load center beyond 24 in
		}
	}
	s := frame(540, 250, "Forklift data plate and four loads", "FORKLIFT FL-07 — DATA PLATE AND TODAY'S LOADS")
	s.rect(16, 38, 170, 120, 3, "#e5e7eb", "#475569")
	s.text(101, 56, 9, `text-anchor="middle" font-weight="700"`, "CAPACITY PLATE")
	s.text(26, 80, 9, ``, "Model: LP-40")
	s.text(26, 98, 9, `font-weight="700"`, fmt.Sprintf("Capacity: %s lb", commas(capLb)))
	s.text(26, 116, 9, `font-weight="700"`, fmt.Sprintf("Load center: %d in", lc))
	s.text(26, 134, 9, ``, "Max lift height: 188 in")
	s.text(26, 150, 7, `fill="#475569"`, "Load center = half the load's depth")
	for i, l := range loads {
		x := 206 + 82*float64(i)
		w := float64(l.depth) * 1.1
		s.rect(x, 150-30, w*0.7, 30, 2, "#d6a15b", "#7c5320")
		s.text(x+w*0.35, 140, 8, `text-anchor="middle" font-weight="700"`, commas(l.lb)+" lb")
		s.text(x+w*0.35, 178, 8, `text-anchor="middle"`, fmt.Sprintf("%d in deep", l.depth))
		s.line(x, 162, x+w*0.7, 162, "#b91c1c", 1)
		markAt(s, x+10, 100, i)
	}
	s.text(16, 236, 8.5, `fill="#334155"`, "Each load is evenly distributed and lifted with the forks fully under it.")
	q := Question{key: "ex/forkliftplate", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which load can this forklift safely lift as rated on its data plate?", ex)
	q.Choices, q.Correct = letterQ(r, "Load", 4, right)
	return q
}

// ── hot aisle / cold aisle ──────────────────────────────────────────────

func hotAisleExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(8)
	s := frame(540, 250, "Data center rack rows, top view", "DATA HALL — RACK ORIENTATION (TOP VIEW)")
	s.rect(40, 100, 460, 40, 0, "#dbeafe", "")
	s.text(270, 124, 10, `text-anchor="middle" font-weight="700" fill="#1d4ed8"`, "COLD AISLE (perforated tiles)")
	s.rect(40, 30, 460, 26, 0, "#fee2e2", "")
	s.text(270, 47, 9, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "HOT AISLE")
	s.rect(40, 186, 460, 26, 0, "#fee2e2", "")
	s.text(270, 203, 9, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "HOT AISLE")
	for i := 0; i < 8; i++ {
		top := i < 4
		x := 60 + 110*float64(i%4)
		y := 58.0
		if !top {
			y = 142
		}
		s.rect(x, y, 90, 42, 3, "#334155", "")
		facesCold := true
		if i == bad {
			facesCold = false
		}
		frontY := y + 38
		if !top {
			frontY = y
		}
		if !facesCold {
			if top {
				frontY = y
			} else {
				frontY = y + 38
			}
		}
		s.rect(x, frontY, 90, 4, 0, "#38bdf8", "")
		s.text(x+45, y+25, 9, `text-anchor="middle" font-weight="700" fill="#fff"`, fmt.Sprintf("Rack %s", lettered(i)))
	}
	s.text(40, 236, 8.5, `fill="#334155"`, "Blue edge = rack front (air intake). Servers draw air in the front and exhaust it out the back.")
	q := Question{key: "ex/hotaisle", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which rack is installed backward, pulling in hot exhaust air?", ex)
	var wrongs []string
	for _, i := range sample(r, []int{0, 1, 2, 3, 4, 5, 6, 7}, 8) {
		if i != bad && len(wrongs) < 3 {
			wrongs = append(wrongs, "Rack "+lettered(i))
		}
	}
	q.Choices, q.Correct = withChoices(r, "Rack "+lettered(bad), wrongs)
	return q
}

// ── patent drawing ──────────────────────────────────────────────────────

func patentFigExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	parts := []string{"base plate", "pivot arm", "return spring", "staple magazine", "anvil", "ergonomic grip"}
	nums := []int{10, 12, 14, 16, 18, 20}
	perm := orderN(r, 6)
	s := frame(540, 250, "Patent drawing FIG. 1 with reference numerals", "U.S. PATENT 9,447,106 — ACPWB ERGONOMIC STAPLER — FIG. 1")
	// stapler outline
	s.rect(90, 180, 300, 14, 2, "#fff", "#111")                                                    // base plate
	s.raw(`<path d="M110 176 L120 120 L370 108 L380 128 L130 150 Z" fill="#fff" stroke="#111"/>`)  // pivot arm
	s.raw(`<path d="M150 176 l6 -10 l6 10 l6 -10 l6 10 l6 -10 l6 10" fill="none" stroke="#111"/>`) // spring
	s.rect(140, 128, 220, 10, 1, "#fff", "#111")                                                   // magazine
	s.rect(330, 172, 40, 8, 1, "#fff", "#111")                                                     // anvil
	s.raw(`<path d="M250 112 q40 -30 90 -10" fill="none" stroke="#111" stroke-width="2"/>`)        // grip
	anchor := [][2]float64{{240, 194}, {125, 132}, {168, 170}, {250, 133}, {350, 176}, {300, 100}}
	labelAt := [][2]float64{{240, 226}, {60, 100}, {150, 220}, {430, 150}, {450, 200}, {320, 60}}
	for i := 0; i < 6; i++ {
		a, l := anchor[i], labelAt[i]
		s.line(l[0], l[1]-4, a[0], a[1], "#111", 0.8)
		s.text(l[0], l[1], 10, `text-anchor="middle" font-weight="700"`, itoa(nums[perm[i]]))
	}
	t := r.intn(6)
	q := Question{key: "ex/patentfig", Exhibit: s.done()}
	var wrongs []string
	for i := 0; i < 6; i++ {
		if i != t {
			wrongs = append(wrongs, parts[i])
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. In FIG. 1, which part does reference numeral %d point to?", ex, nums[perm[t]])
	q.Choices, q.Correct = withChoices(r, parts[t], sample(r, wrongs, 3))
	return q
}

// ── desk reach zones ────────────────────────────────────────────────────

func reachZonesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	items := sample(r, []string{"Mouse", "Phone", "Notepad", "Stapler", "Water bottle", "Calculator", "Headset"}, 4)
	bad := r.intn(4)
	s := frame(540, 250, "Desk top view with reach zones", "DESK LAYOUT — REACH ZONES (TOP VIEW)")
	cx, cy := 270.0, 230.0
	s.rect(60, 50, 420, 180, 4, "#fef3c7", "#a16207")
	s.raw(`<path d="M%.1f %.1f A170 170 0 0 1 %.1f %.1f" fill="#fde68a" stroke="#a16207" stroke-dasharray="4 3"/>`, cx-170, cy, cx+170, cy)
	s.raw(`<path d="M%.1f %.1f A100 100 0 0 1 %.1f %.1f" fill="#bbf7d0" stroke="#15803d" stroke-dasharray="4 3"/>`, cx-100, cy, cx+100, cy)
	s.rect(70, 58, 10, 8, 1, "#bbf7d0", "#15803d")
	s.text(84, 65, 8.5, `font-weight="700" fill="#15803d"`, "Normal reach (frequent items)")
	s.rect(70, 72, 10, 8, 1, "#fde68a", "#a16207")
	s.text(84, 79, 8.5, `font-weight="700" fill="#a16207"`, "Maximum reach (occasional items)")
	s.rect(cx-40, cy-30, 80, 22, 3, "#475569", "")
	s.text(cx, cy-15, 8, `text-anchor="middle" fill="#fff"`, "Keyboard")
	for i, it := range items {
		var rr float64
		if i == bad {
			rr = 125 + float64(r.between(0, 30))
		} else {
			rr = 55 + float64(r.between(0, 25))
		}
		a := (160 - float64(i)*46) * math.Pi / 180
		x, y := cx+rr*math.Cos(a), cy-rr*math.Sin(a)
		s.rect(x-24, y-8, 48, 16, 3, "#fff", "#334155")
		s.text(x, y+3, 7.5, `text-anchor="middle" font-weight="700"`, it)
	}
	q := Question{key: "ex/reachzones", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. The worker uses all four items many times an hour. Which item should be moved closer?", ex)
	var wrongs []string
	for i, it := range items {
		if i != bad {
			wrongs = append(wrongs, it)
		}
	}
	q.Choices, q.Correct = withChoices(r, items[bad], wrongs)
	return q
}

// ── IBM punched card (Hollerith) ────────────────────────────────────────

// holes returns the punched rows for a character: 12, 11, 0..9 as
// indexes 0..11 (12=0, 11=1, 0=2, 1=3 … 9=11).
func holes(ch rune) []int {
	switch {
	case ch >= '0' && ch <= '9':
		return []int{int(ch-'0') + 2}
	case ch >= 'A' && ch <= 'I':
		return []int{0, int(ch-'A') + 3}
	case ch >= 'J' && ch <= 'R':
		return []int{1, int(ch-'J') + 3}
	case ch >= 'S' && ch <= 'Z':
		return []int{2, int(ch-'S') + 4}
	}
	return nil
}

func punchCardExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	words := []string{"PAYROLL", "PERCH", "ACPWB", "BONUS", "AUDIT", "LEDGER", "MERIT", "SALARY", "PENSION", "BADGE", "RETAIN", "QUORUM", "VOUCHER", "MEMO"}
	w := pick(r, words)
	s := frame(540, 250, "IBM 80-column punched card, first columns", "PUNCHED CARD — COLUMNS 1–"+itoa(len(w)))
	s.rect(20, 30, 500, 210, 6, "#f5e6c4", "#a8956a")
	rowNames := []string{"12", "11", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for i, rn := range rowNames {
		s.text(36, 58+13.5*float64(i), 7.5, `text-anchor="end" fill="#8a7a55"`, rn)
	}
	for col := 0; col < 10; col++ {
		x := 60 + 44*float64(col)
		s.text(x+5, 42, 7, `text-anchor="middle" font-weight="700" fill="#6b5d3d"`, itoa(col+1))
		for i := 2; i < 12; i++ {
			s.text(x+5, 58+13.5*float64(i), 7, `text-anchor="middle" fill="#c9b78f"`, rowNames[i])
		}
	}
	for col, ch := range w {
		x := 60 + 44*float64(col)
		for _, h := range holes(ch) {
			s.rect(x+1, 49+13.5*float64(h), 8, 12, 1, "#111", "")
		}
	}
	s.text(36, 232, 7.5, `fill="#6b5d3d"`, "Zones: 12 + 1–9 = A–I · 11 + 1–9 = J–R · 0 + 2–9 = S–Z · a single punch 0–9 = that digit")
	q := Question{key: "ex/punchcard", Exhibit: s.done()}
	// Wrong answers: shift one letter's zone.
	shift := func(word string, k int) string {
		b := []rune(word)
		ch := b[k]
		switch {
		case ch >= 'A' && ch <= 'I':
			b[k] = ch + 9
		case ch >= 'J' && ch <= 'R':
			if ch == 'J' {
				b[k] = 'A'
			} else {
				b[k] = ch + 8
			}
		default:
			b[k] = 'A' + (ch-'S'+1)%9
		}
		return string(b)
	}
	var wrongs []string
	for _, k := range sample(r, orderN(r, len(w)), len(w)) {
		if x := shift(w, k); x != w && !contains(wrongs, x) && len(wrongs) < 3 {
			wrongs = append(wrongs, x)
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. Using the Hollerith code shown on the card, what word is punched in columns 1–%d?", ex, len(w))
	q.Choices, q.Correct = withChoices(r, w, wrongs)
	return q
}

// ── bot share of traffic (stacked bars) ─────────────────────────────────

func siteTrafficExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := 8
	h0 := r.between(0, 15)
	bad := r.intn(n)
	type hr struct{ human, bot int }
	hs := make([]hr, n)
	for i := range hs {
		tot := r.between(40, 120)
		pct := r.between(35, 72)
		if i == bad {
			pct = r.between(84, 95)
		}
		b := tot * pct / 100
		hs[i] = hr{tot - b, b}
	}
	s := frame(540, 260, "Stacked bar chart of human and bot requests", "ACPWB.COM REQUESTS PER HOUR (THOUSANDS)")
	x0, y1 := 60.0, 220.0
	k := 150.0 / 120
	for i, h := range hs {
		x := x0 + 56*float64(i)
		s.rect(x, y1-float64(h.human)*k, 36, float64(h.human)*k, 0, "#2563eb", "")
		s.rect(x, y1-float64(h.human+h.bot)*k, 36, float64(h.bot)*k, 0, "#f97316", "")
		s.text(x+18, y1-float64(h.human+h.bot)*k-4, 7.5, `text-anchor="middle" font-weight="700"`, fmt.Sprintf("%d/%d", h.bot, h.human+h.bot))
		s.text(x+18, y1+13, 8, `text-anchor="middle"`, clockShort((h0+i)%24))
	}
	s.line(x0-6, y1, 520, y1, "#64748b", 1)
	s.rect(60, 32, 10, 8, 0, "#f97316", "")
	s.text(74, 39, 8, ``, "Automated (bots)")
	s.rect(170, 32, 10, 8, 0, "#2563eb", "")
	s.text(184, 39, 8, ``, "Human")
	s.text(300, 39, 8, `fill="#475569"`, "Labels: bot requests / total")
	q := Question{key: "ex/sitetraffic", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. In which hour did automated traffic exceed 80%% of all requests?", ex)
	var wrongs []string
	for _, i := range sample(r, orderN(r, n), n) {
		if i != bad && len(wrongs) < 3 {
			wrongs = append(wrongs, clockShort((h0+i)%24))
		}
	}
	q.Choices, q.Correct = withChoices(r, clockShort((h0+bad)%24), wrongs)
	return q
}

// ── grandfather-father-son backup rotation ──────────────────────────────

func gfsBackupExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// A month starting on a given weekday (0 = Monday), 30 or 31 days.
	start := r.intn(7)
	days := pick(r, []int{30, 31})
	month := pick(r, []string{"March", "May", "August", "October"})
	s := frame(540, 270, "Backup rotation calendar", "BACKUP ROTATION — "+strings.ToUpper(month))
	dn := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for i, d := range dn {
		s.text(64+66*float64(i), 44, 9, `text-anchor="middle" font-weight="700"`, d)
	}
	lastFri := 0
	for d := 1; d <= days; d++ {
		if (start+d-1)%7 == 4 {
			lastFri = d
		}
	}
	kind := func(d int) string {
		wd := (start + d - 1) % 7
		switch {
		case wd >= 5:
			return "No backup"
		case wd == 4 && d == lastFri:
			return "Monthly tape (Grandfather)"
		case wd == 4:
			return "Weekly tape (Father)"
		}
		return fmt.Sprintf("Daily tape: %s (Son)", dn[wd])
	}
	for d := 1; d <= days; d++ {
		cell := start + d - 1
		x, y := 32+66*float64(cell%7), 50+34*float64(cell/7)
		s.rect(x, y, 64, 32, 2, "#f8fafc", "#cbd5e1")
		s.text(x+6, y+12, 8.5, `font-weight="700"`, itoa(d))
	}
	s.text(32, 262, 8, `fill="#334155"`, "Rotation: daily tapes Mon–Thu, a weekly tape each Friday, and a monthly tape on the last Friday of the month.")
	var d int
	for {
		d = r.between(1, days)
		if (start+d-1)%7 < 5 {
			break
		}
	}
	if r.chance(0.4) {
		d = lastFri
	}
	right := kind(d)
	pool := []string{"Monthly tape (Grandfather)", "Weekly tape (Father)", "Daily tape: Mon (Son)", "Daily tape: Tue (Son)", "Daily tape: Wed (Son)", "Daily tape: Thu (Son)", "No backup"}
	q := Question{key: "ex/gfsbackup", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which tape is written on %s %d?", ex, month, d)
	q.Choices, q.Correct = oneOf(r, right, pool)
	return q
}

// ── fire alarm annunciator ──────────────────────────────────────────────

func annunciatorExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	zones := sample(r, []string{"Loading dock", "Mail room", "3rd floor east", "Cafeteria kitchen", "Server room", "Parking level P1", "Records vault", "Lobby", "Boardroom", "Print shop"}, 6)
	alarm := r.intn(6)
	trouble := (alarm + r.between(1, 5)) % 6
	s := frame(540, 240, "Fire alarm annunciator panel", "FIRE ALARM ANNUNCIATOR — MAIN LOBBY")
	s.rect(30, 34, 480, 194, 6, "#1f2937", "")
	for i, z := range zones {
		y := 54 + 28*float64(i)
		state, col := "NORMAL", "#22c55e"
		switch i {
		case alarm:
			state, col = "ALARM", "#ef4444"
		case trouble:
			state, col = "TROUBLE", "#f59e0b"
		}
		s.circle(56, y, 7, col, "#111")
		s.text(74, y+4, 10, `fill="#f9fafb" font-weight="700"`, fmt.Sprintf("ZONE %d", i+1))
		s.text(150, y+4, 10, `fill="#d1d5db"`, z)
		s.text(490, y+4, 10, `text-anchor="end" font-weight="700" fill="`+col+`"`, state)
	}
	q := Question{key: "ex/annunciator", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Firefighters arrive at the lobby. Where should you direct them?",
		"Refer to %[1]s. Which area is reporting a fire alarm (not a trouble signal)?",
	}), ex)
	var wrongs []string
	for i, z := range zones {
		if i != alarm {
			wrongs = append(wrongs, z)
		}
	}
	w := []string{zones[trouble]}
	for _, x := range sample(r, wrongs, len(wrongs)) {
		if x != zones[trouble] && len(w) < 3 {
			w = append(w, x)
		}
	}
	q.Choices, q.Correct = withChoices(r, zones[alarm], w)
	return q
}

// ── emergency call tree ─────────────────────────────────────────────────

func callTreeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ns := people(r, 7)
	// 0 calls 1,2; 1 calls 3,4; 2 calls 5,6
	s := frame(540, 220, "Emergency call tree", "SITE CLOSURE CALL TREE")
	pos := [][2]float64{{270, 56}, {150, 120}, {390, 120}, {90, 186}, {210, 186}, {330, 186}, {450, 186}}
	edges := [][2]int{{0, 1}, {0, 2}, {1, 3}, {1, 4}, {2, 5}, {2, 6}}
	for _, e := range edges {
		a, b := pos[e[0]], pos[e[1]]
		s.line(a[0], a[1]+12, b[0], b[1]-12, "#64748b", 1.5)
	}
	for i, p := range pos {
		fill := "#e0f2fe"
		if i == 0 {
			fill = "#fde68a"
		}
		s.rect(p[0]-56, p[1]-12, 112, 24, 4, fill, "#334155")
		s.text(p[0], p[1]+4, 8.5, `text-anchor="middle" font-weight="700"`, ns[i])
	}
	s.text(20, 214, 8, `fill="#334155"`, "Rule: if someone can't be reached, the person who called them calls that person's contacts.")
	q := Question{key: "ex/calltree", Exhibit: s.done()}
	mid := r.between(1, 2)
	kids := []int{2*mid + 1, 2*mid + 2}
	q.Text = fmt.Sprintf("Refer to %s. %s cannot be reached. Who should call %s and %s?", ex, ns[mid], ns[kids[0]], ns[kids[1]])
	other := 3 - mid
	q.Choices, q.Correct = withChoices(r, ns[0], []string{ns[other], ns[kids[0]], ns[2*other+1]})
	return q
}

// ── sharps container fill line ──────────────────────────────────────────

func sharpsBoxExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(540, 210, "Four sharps containers", "SHARPS CONTAINERS — FIRST AID ROOMS")
	for i := 0; i < 4; i++ {
		x := 40 + 128*float64(i)
		top, h := 56.0, 130.0
		fill := 0.15 + 0.45*r.float()
		if i == bad {
			fill = 0.8 + 0.12*r.float()
		}
		s.rect(x, top, 80, h, 6, "#fee2e2", "#dc2626")
		s.rect(x+2, top+h*(1-fill), 76, h*fill-2, 4, "#fca5a5", "")
		fy := top + h*0.25
		s.line(x, fy, x+80, fy, "#111", 1.5)
		s.text(x+40, fy-4, 7, `text-anchor="middle" font-weight="700"`, "FILL LINE")
		s.rect(x+20, top-10, 40, 10, 2, "#dc2626", "")
		markAt(s, x-8, top, i)
	}
	q := Question{key: "ex/sharpsbox", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Sharps containers must be closed and replaced when they reach the fill line (about three-quarters full). Which container must be replaced now?", ex)
	q.Choices, q.Correct = letterQ(r, "Container", 4, bad)
	return q
}

// ── truck No-Zones ──────────────────────────────────────────────────────

func blindSpotsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(540, 260, "Top view of a truck with blind spots", "TRACTOR-TRAILER NO-ZONES (TOP VIEW, TRUCK HEADING RIGHT)")
	// truck: trailer x 160..380, cab 380..430 at y 120..150
	s.raw(`<path d="M430 120 L520 110 L520 160 L430 150 Z" fill="#fecaca" fill-opacity=".8"/>`) // front
	s.raw(`<path d="M160 120 L40 112 L40 158 L160 150 Z" fill="#fecaca" fill-opacity=".8"/>`)   // rear
	s.raw(`<path d="M400 152 L230 152 L150 230 L400 230 Z" fill="#fecaca" fill-opacity=".8"/>`) // right side (passenger)
	s.raw(`<path d="M420 118 L380 118 L360 70 L420 70 Z" fill="#fecaca" fill-opacity=".8"/>`)   // left, under mirror
	s.rect(160, 120, 220, 30, 2, "#94a3b8", "#334155")
	s.rect(382, 122, 48, 26, 4, "#334155", "")
	s.text(270, 139, 9, `text-anchor="middle" font-weight="700" fill="#fff"`, "TRAILER")
	s.line(20, 104, 520, 104, "#e5e7eb", 1)
	s.line(20, 166, 520, 166, "#e5e7eb", 1)
	safe := [][2]float64{{90, 70}, {250, 76}, {470, 210}, {310, 60}, {60, 210}}
	inside := [][2]float64{{480, 136}, {90, 136}, {300, 196}, {400, 88}}
	bad := r.intn(4)
	in := pick(r, inside)
	outs := sample(r, safe, 3)
	var pts [][2]float64
	for i := 0; i < 4; i++ {
		if i == bad {
			pts = append(pts, in)
		} else {
			pts = append(pts, outs[0])
			outs = outs[1:]
		}
	}
	for i, p := range pts {
		s.rect(p[0]-16, p[1]-9, 32, 18, 4, "#1d4ed8", "")
		markAt(s, p[0], p[1], i)
	}
	s.text(20, 250, 8, `fill="#334155"`, "Shaded = areas the truck driver cannot see. Lanes are illustrative.")
	q := Question{key: "ex/blindspots", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which car is in one of the truck's No-Zones?", ex)
	q.Choices, q.Correct = letterQ(r, "Car", 4, bad)
	return q
}

// ── rack beam loads ─────────────────────────────────────────────────────

func rackLoadExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{2000, 2500, 3000})
	bad := r.intn(4)
	s := frame(540, 260, "Pallet rack with four beam levels", "PALLET RACK BAY 14 — LOAD PLAQUE AND CURRENT LOADS")
	s.rect(120, 34, 8, 210, 0, "#1d4ed8", "")
	s.rect(400, 34, 8, 210, 0, "#1d4ed8", "")
	for lv := 0; lv < 4; lv++ {
		y := 230 - 50*float64(lv)
		s.rect(120, y, 288, 7, 0, "#f97316", "")
		var a, b int
		for {
			a, b = r.between(4, 20)*100, r.between(4, 20)*100
			if (lv == bad) == (a+b > limit) && a+b != limit {
				break
			}
		}
		for k, w := range []int{a, b} {
			x := 140 + 130*float64(k)
			s.rect(x, y-36, 118, 34, 2, "#d6a15b", "#7c5320")
			s.text(x+59, y-15, 9, `text-anchor="middle" font-weight="700"`, commas(w)+" lb")
		}
		s.text(110, y+3, 9, `text-anchor="end" font-weight="700"`, fmt.Sprintf("Level %d", lv+1))
	}
	s.rect(420, 60, 110, 70, 3, "#fef3c7", "#a16207")
	s.text(475, 80, 8, `text-anchor="middle" font-weight="700"`, "LOAD PLAQUE")
	s.text(475, 98, 8, `text-anchor="middle"`, "Max per beam level")
	s.text(475, 116, 10, `text-anchor="middle" font-weight="700"`, commas(limit)+" lb")
	q := Question{key: "ex/rackload", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which beam level is loaded beyond the rack's posted capacity?", ex)
	q.Choices, q.Correct = numChoices(r, bad+1, []int{1, -1, 2, -2, 3, -3}, func(n int) string { return fmt.Sprintf("Level %d", n) })
	return q
}

// ── fall clearance ──────────────────────────────────────────────────────

func fallClearanceExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	need := 18.5
	bad := r.intn(4)
	s := frame(540, 250, "Four anchor points and the levels below", "FALL CLEARANCE — 6-FT SHOCK-ABSORBING LANYARD")
	for i := 0; i < 4; i++ {
		x := 18 + 131*float64(i)
		h := 19.0 + float64(r.between(1, 10))
		if i == bad {
			h = 11 + float64(r.between(0, 6))
		}
		k := 6.0
		s.rect(x, 36, 120, 200, 4, "#f8fafc", "#e2e8f0")
		gy := 226.0
		ay := gy - h*k
		s.line(x+4, gy, x+116, gy, "#475569", 2)
		s.text(x+60, gy-4, 7, `text-anchor="middle" fill="#475569"`, "lower level")
		s.circle(x+34, ay, 5, "#dc2626", "#111")
		s.line(x+10, ay+6, x+60, ay+6, "#475569", 3)
		dimArrow(s, x+86, gy, x+86, ay, fmt.Sprintf("%g ft", h))
		markAt(s, x+14, 48, i)
	}
	s.text(18, 246, 7.5, `fill="#334155"`, "Clearance from anchor: 6 ft lanyard + 3.5 ft deceleration + 6 ft worker height + 3 ft safety margin = 18.5 ft.")
	q := Question{key: "ex/fallclearance", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the clearance calculation shown (%.1f ft), at which anchor would a worker hit the lower level before the lanyard stops the fall?", ex, need)
	q.Choices, q.Correct = letterQ(r, "Anchor", 4, bad)
	return q
}

// ── chair fit ───────────────────────────────────────────────────────────

func chairFitExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(540, 200, "Four seated workers, side view", "CHAIR ADJUSTMENT — SIDE VIEW")
	for i := 0; i < 4; i++ {
		x := 14 + 131*float64(i)
		fy := 186.0
		seat := 128.0
		if i == bad {
			seat = 104 // seat too high: feet dangle
		}
		s.rect(x, 36, 122, 156, 4, "#f8fafc", "#e2e8f0")
		s.line(x+4, fy, x+118, fy, "#475569", 2)
		s.rect(x+24, seat, 44, 6, 2, "#334155", "")
		s.line(x+46, seat+6, x+46, fy, "#334155", 3)
		s.line(x+24, seat, x+22, seat-48, "#334155", 4)
		// person
		s.circle(x+34, seat-60, 9, "#fde68a", "#111")
		s.line(x+32, seat-51, x+30, seat-2, "#111", 3)
		s.line(x+30, seat-2, x+66, seat-2, "#111", 3)
		foot := math.Min(seat-2+42, fy)
		s.line(x+66, seat-2, x+68, foot, "#111", 3)
		s.line(x+68, foot, x+80, foot, "#111", 3)
		s.rect(x+74, seat-24, 40, 4, 1, "#a16207", "")
		markAt(s, x+110, 48, i)
	}
	q := Question{key: "ex/chairfit", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. OSHA's workstation guidance says feet should rest flat on the floor or a footrest, with thighs about parallel to the floor. Which worker needs a lower seat or a footrest?", ex)
	q.Choices, q.Correct = letterQ(r, "Worker", 4, bad)
	return q
}

// ── assembly point head count ───────────────────────────────────────────

func headcountExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := r.between(12, 16)
	ns := people(r, n)
	missing := r.between(1, 3)
	away := map[int]bool{}
	for _, k := range pickN(r, n, missing+1) {
		away[k] = true
	}
	// one of the "away" is excused (on PTO), the rest are unaccounted for
	var pto int
	for k := range away {
		pto = k
		break
	}
	for k := range away {
		if k < pto {
			pto = k
		}
	}
	s := frame(540, 40+18*float64((n+1)/2)+30, "Assembly point roster", "ASSEMBLY POINT B — FLOOR 3 ROSTER")
	for i, nm := range ns {
		x := 30 + 250*float64(i%2)
		y := 50 + 18*float64(i/2)
		mark, col := "✔", "#16a34a"
		if away[i] {
			mark, col = "☐", "#dc2626"
			if i == pto {
				mark, col = "PTO", "#64748b"
			}
		}
		s.text(x, y, 9.5, `font-weight="700" fill="`+col+`"`, mark)
		s.text(x+32, y, 9.5, ``, nm)
	}
	s.text(30, 40+18*float64((n+1)/2)+18, 8, `fill="#334155"`, "✔ = checked in at the assembly point · PTO = on approved leave today")
	q := Question{key: "ex/headcount", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. During the evacuation, how many people on this roster are unaccounted for and must be reported to the fire department?", ex)
	q.Choices, q.Correct = numChoices(r, missing, []int{1, 2, 3, -1}, itoa)
	return q
}
