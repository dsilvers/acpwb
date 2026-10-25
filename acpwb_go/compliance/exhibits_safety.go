package compliance

// Safety exhibits, round three. Accurate rules: OSHA's 85 dBA action
// level, confined-space entry limits printed on the permit, the NWS wind
// chill formula, the 40–140°F danger zone, GHS label elements, SDS section
// numbers, USPS suspicious-mail indicators, Drop/Cover/Hold On, and
// monitor placement (top at or slightly below eye level, an arm's length
// away).

import (
	"fmt"
	"math"
	"strings"
)

func init() {
	for k, g := range map[string]exhibitGen{
		"noisemeter":    noiseMeterExhibit,
		"gasmonitor":    gasMonitorExhibit,
		"windchill":     windChillExhibit,
		"foodtemps":     foodTempsExhibit,
		"allergens":     allergenMenuExhibit,
		"liftweights":   liftWeightsExhibit,
		"ppecheck":      ppeCheckExhibit,
		"walkways":      walkwaysExhibit,
		"quakeroom":     quakeRoomExhibit,
		"inspectiontag": inspectionTagExhibit,
		"chemlabel":     chemLabelExhibit,
		"sds":           sdsExhibit,
		"powerstrips":   powerStripsExhibit,
		"suspiciouspkg": suspiciousPackageExhibit,
		"monitorsetup":  monitorSetupExhibit,
		"occupancy":     occupancyExhibit,
		"firedrill":     fireDrillExhibit,
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
		"ppe":          {"noisemeter", "ppecheck", "liftweights"},
		"spaces":       {"gasmonitor"},
		"winter":       {"windchill"},
		"foodsafety":   {"foodtemps", "allergens"},
		"potluck":      {"foodtemps", "allergens"},
		"kitchen":      {"foodtemps"},
		"medical":      {"allergens"},
		"hospitality":  {"allergens"},
		"holiday":      {"allergens", "powerstrips"},
		"warehouse":    {"walkways", "liftweights", "ppecheck"},
		"officemove":   {"liftweights"},
		"construction": {"ppecheck"},
		"earthquake":   {"quakeroom"},
		"fire":         {"inspectiontag", "occupancy", "firedrill"},
		"safety":       {"inspectiontag", "powerstrips", "firedrill"},
		"hazcom":       {"chemlabel", "sds"},
		"electrical":   {"powerstrips"},
		"celebrations": {"powerstrips"},
		"mail":         {"suspiciouspkg"},
		"bombthreat":   {"suspiciouspkg"},
		"deliveries":   {"suspiciouspkg"},
		"ergonomics":   {"monitorsetup"},
		"eyestrain":    {"monitorsetup"},
		"workation":    {"monitorsetup"},
		"meetings":     {"occupancy"},
		"offsites":     {"occupancy"},
		"evacassist":   {"firedrill"},
		"continuity":   {"firedrill"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// ── sound level readings (85 dBA action level) ─────────────────────────

func noiseMeterExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	stations := sample(r, []string{"Print room", "Mailroom sorter", "Loading dock", "Generator room", "Shredder station", "Warehouse aisle 4", "Cafeteria dish line", "Server room", "Lobby", "Records vault"}, 5)
	over := 0
	s := newSVG(440, 250, "Sound level survey")
	s.rect(0, 0, 440, 250, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, "Noise Survey — 8-hour time-weighted averages (dBA)")
	x0, y0 := 120.0, 50.0
	for i, st := range stations {
		v := r.between(68, 96)
		if v == 85 && r.chance(0.5) {
			v = 84
		}
		if v >= 85 {
			over++
		}
		y := y0 + float64(i)*36
		s.text(x0-8, y+15, 10, `text-anchor="end"`, st)
		w := float64(v-60) / 40 * 260
		col := "#2e7d32"
		if v >= 85 {
			col = "#e65100"
		}
		s.rect(x0, y+2, w, 20, 3, col, "")
		s.text(x0+w+6, y+16, 10, `font-weight="700"`, fmt.Sprintf("%d dBA", v))
	}
	ax := x0 + 25.0/40*260
	s.dash(ax, 40, ax, 230, "#c62828", 1.2)
	s.text(ax, 242, 9, `text-anchor="middle" fill="#c62828"`, "85 dBA")
	q := Question{key: "ex/noisemeter", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. OSHA requires a hearing conservation program where the 8-hour average reaches 85 dBA. How many of these stations are at or above that level?", ex)
	q.Choices, q.Correct = countQ(r, over, "station")
	return q
}

// ── four-gas monitor ───────────────────────────────────────────────────

func gasMonitorExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	o2 := float64(r.between(200, 215)) / 10
	lel := r.between(0, 6)
	h2s := r.between(0, 6)
	co := r.between(0, 20)
	fault := r.intn(5) // 0 = all good
	switch fault {
	case 1:
		o2 = float64(r.between(170, 192)) / 10
	case 2:
		lel = r.between(11, 30)
	case 3:
		h2s = r.between(11, 40)
	case 4:
		co = r.between(36, 120)
	}
	s := newSVG(420, 240, "Four-gas monitor readings")
	s.rect(0, 0, 420, 240, 8, "#e8ebe6", "")
	s.rect(20, 14, 180, 212, 18, "#f2b705", "#8a6a00")
	s.rect(36, 40, 148, 140, 6, "#1c2a1c", "")
	for i, kv := range []struct{ l, v string }{{"O2 %", fmt.Sprintf("%.1f", o2)}, {"LEL %", fmt.Sprint(lel)}, {"H2S ppm", fmt.Sprint(h2s)}, {"CO ppm", fmt.Sprint(co)}} {
		y := 64 + float64(i)*32
		s.text(46, y, 10, `fill="#9fe39f" font-family="Courier New,monospace"`, kv.l)
		s.text(174, y, 18, `text-anchor="end" fill="#9fe39f" font-weight="700" font-family="Courier New,monospace"`, kv.v)
	}
	s.circle(110, 204, 12, "#333", "")
	s.rect(220, 30, 186, 180, 6, "#fff", "#999")
	s.text(232, 52, 11, `font-weight="700"`, "Entry Permit — acceptable")
	for i, l := range []string{"Oxygen: 19.5% – 23.5%", "Flammable (LEL): below 10%", "Hydrogen sulfide: below 10 ppm", "Carbon monoxide: below 35 ppm", "", "Test order: O2, LEL, toxics"} {
		s.text(232, 76+float64(i)*20, 10, "", l)
	}
	q := Question{key: "ex/gasmonitor", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. The attendant tests the vault's atmosphere before %s enters. Based on the permit, may entry begin?", ex, sl["FIRST"])
	opts := []string{"Yes; every reading is within the permit's limits", "No; oxygen is outside the acceptable range", "No; the flammable gas level is too high", "No; hydrogen sulfide is too high", "No; carbon monoxide is too high"}
	q.Choices, q.Correct = oneOf(r, opts[fault], opts)
	return q
}

// ── wind chill (NWS) ───────────────────────────────────────────────────

func windChill(t, v float64) float64 {
	p := math.Pow(v, 0.16)
	return 35.74 + 0.6215*t - 35.75*p + 0.4275*t*p
}

func windChillExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	t := float64(5 * r.between(-4, 6))
	v := float64(5 * r.between(2, 10))
	wc := int(math.Round(windChill(t, v)))
	s := newSVG(320, 180, "Outdoor temperature and wind readout")
	s.rect(0, 0, 320, 180, 14, "#16324f", "")
	s.text(160, 26, 11, `text-anchor="middle" fill="#9ec5e8" font-weight="700"`, "PARKING STRUCTURE · ROOF LEVEL")
	for i, kv := range [][2]string{{"AIR TEMP", fmt.Sprintf("%.0f°F", t)}, {"WIND", fmt.Sprintf("%.0f mph", v)}} {
		x := 20 + float64(i)*150
		s.rect(x, 42, 130, 90, 8, "#dbe9f7", "")
		s.text(x+65, 62, 10, `text-anchor="middle" font-weight="700" fill="#1a3a5c"`, kv[0])
		s.text(x+65, 108, 30, `text-anchor="middle" font-weight="700" fill="#0d2238" font-family="Courier New,monospace"`, kv[1])
	}
	s.text(160, 160, 9, `text-anchor="middle" fill="#9ec5e8"`, "Wind chill = 35.74 + 0.6215T − 35.75V^0.16 + 0.4275TV^0.16")
	q := Question{key: "ex/windchill", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. About what is the wind chill for %s, who is clearing ice on the roof level?", ex, sl["FIRST"])
	vals := []int{wc + 10, wc - 10, int(t)}
	if int(t) == wc+10 || int(t) == wc-10 || int(t) == wc {
		vals[2] = wc + 20
	}
	var ws []string
	for _, v := range vals {
		ws = append(ws, fmt.Sprintf("%d°F", v))
	}
	q.Choices, q.Correct = withChoices(r, fmt.Sprintf("%d°F", wc), ws)
	return q
}

// ── food holding temperatures ──────────────────────────────────────────

func foodTempsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type food struct {
		name string
		hot  bool
	}
	foods := sample(r, []food{{"Chili", true}, {"Mac and cheese", true}, {"Bratwurst", true}, {"Meatballs", true}, {"Potato salad", false}, {"Cut fruit", false}, {"Chicken salad", false}, {"Shrimp cocktail", false}, {"Yogurt parfaits", false}, {"Tomato soup", true}}, 4)
	bad := r.intn(4)
	s := newSVG(440, 200, "Buffet food thermometer readings")
	s.rect(0, 0, 440, 200, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, "Buffet at "+sl["EVENT"]+" · 1:30 p.m. check")
	var names []string
	for i, f := range foods {
		temp := 0
		if f.hot {
			temp = r.between(142, 175)
			if i == bad {
				temp = r.between(95, 135)
			}
		} else {
			temp = r.between(33, 40)
			if i == bad {
				temp = r.between(48, 70)
			}
		}
		x := 20 + float64(i)*104
		s.rect(x, 40, 92, 140, 8, "#f7f3ea", "#ddd")
		s.text(x+46, 60, 10, `text-anchor="middle" font-weight="700"`, f.name)
		s.rect(x+40, 72, 12, 70, 6, "#fff", "#999")
		frac := math.Min(1, math.Max(0, float64(temp-20)/170))
		s.rect(x+42, 72+68*(1-frac), 8, 68*frac, 4, "#c62828", "")
		s.circle(x+46, 148, 9, "#c62828", "#999")
		s.text(x+46, 174, 12, `text-anchor="middle" font-weight="700"`, fmt.Sprintf("%d°F", temp))
		names = append(names, f.name)
	}
	q := Question{key: "ex/foodtemps", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Hot food must be held at 140°F or above and cold food at 40°F or below. Which item is in the danger zone?", ex)
	q.Choices, q.Correct = oneOf(r, names[bad], append(names, "None of them"))
	return q
}

// ── allergen menu ──────────────────────────────────────────────────────

var allergenList = []string{"Peanut", "Tree nut", "Milk", "Egg", "Wheat", "Soy", "Fish", "Shellfish", "Sesame"}

func allergenMenuExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	avoid := sample(r, allergenList, 2)
	dishes := sample(r, []string{"Thai noodle bowl", "Kale Caesar", "Fish fry", "Sesame tofu", "Cheese curds", "Turkey club", "Shrimp tacos", "Pesto pasta", "Grilled chicken plate", "Fruit and rice bowl", "Egg salad croissant"}, 5)
	safe := r.intn(5)
	tags := make([][]string, 5)
	for i := range dishes {
		var t []string
		for _, a := range sample(r, allergenList, r.between(1, 3)) {
			if i == safe && contains(avoid, a) {
				continue
			}
			t = append(t, a)
		}
		if i != safe && !contains(t, avoid[0]) && !contains(t, avoid[1]) {
			t = append(t, pick(r, avoid))
		}
		if len(t) == 0 {
			t = []string{"None of the major allergens"}
		}
		tags[i] = t
	}
	var rows [][]string
	for i, d := range dishes {
		rows = append(rows, []string{d, strings.Join(tags[i], ", ")})
	}
	q := Question{key: "ex/allergens", Exhibit: docExhibit("Catering Menu — "+capFirst(sl["EVENT"]), "Allergens as labeled by "+sl["VENDOR"], []string{"Dish", "Contains"}, rows, "Ask the caterer about cross-contact if an allergy is severe.")}
	q.Text = fmt.Sprintf("Refer to %s. %s is allergic to %s and %s. Which dish is labeled free of both?", ex, sl["FIRST"], strings.ToLower(avoid[0]), strings.ToLower(avoid[1]))
	q.Choices, q.Correct = oneOf(r, dishes[safe], dishes)
	return q
}

// ── box weights ────────────────────────────────────────────────────────

func liftWeightsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{35, 40, 50})
	n := r.between(5, 7)
	heavy := 0
	s := newSVG(460, 210, "Boxes with weight labels")
	s.rect(0, 0, 460, 210, 8, "#efeae0", "")
	s.text(14, 22, 12, `font-weight="700" fill="#5d4037"`, "Records move · "+sl["PLACE"])
	for i := 0; i < n; i++ {
		w := r.between(12, 78)
		if w > limit {
			heavy++
		}
		x := 14 + float64(i)*63
		bh := 40 + float64(w)
		s.rect(x, 190-bh, 56, bh, 2, "#c69c6d", "#8d6e4f")
		s.line(x, 190-bh+12, x+56, 190-bh+12, "#8d6e4f", 1)
		s.rect(x+8, 190-bh+20, 40, 18, 2, "#fff", "")
		s.text(x+28, 190-bh+33, 10, `text-anchor="middle" font-weight="700"`, fmt.Sprintf("%d lb", w))
	}
	q := Question{key: "ex/liftweights", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. ACPWB's policy limits one-person lifts to %d pounds; heavier boxes need two people or a cart. How many of these boxes need two people or a cart?", ex, limit)
	q.Choices, q.Correct = countQ(r, heavy, "box")
	for i := range q.Choices {
		q.Choices[i] = strings.Replace(q.Choices[i], "boxs", "boxes", 1)
	}
	return q
}

// ── PPE check ──────────────────────────────────────────────────────────

func ppeCheckExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	gear := []string{"hard hat", "safety glasses", "high-visibility vest", "gloves", "hearing protection"}
	req := sample(r, gear, 3)
	names := people(r, 4)
	bad := r.intn(4)
	missing := pick(r, req)
	s := newSVG(460, 260, "Workers and required PPE sign")
	s.rect(0, 0, 460, 260, 8, "#eef1f4", "")
	s.rect(14, 12, 432, 40, 4, "#1565c0", "")
	s.text(230, 30, 11, `text-anchor="middle" fill="#fff" font-weight="700"`, "PPE REQUIRED BEYOND THIS POINT")
	s.text(230, 45, 10, `text-anchor="middle" fill="#fff"`, strings.ToUpper(strings.Join(req, " · ")))
	for i, n := range names {
		x := 60 + float64(i)*110
		has := map[string]bool{}
		for _, g := range gear {
			has[g] = r.chance(0.5) || contains(req, g)
		}
		if i == bad {
			has[missing] = false
		}
		// body
		s.circle(x, 92, 14, "#e8d5c0", "#999")
		vest := "#607d8b"
		if has["high-visibility vest"] {
			vest = "#c6ff00"
		}
		s.rect(x-18, 108, 36, 52, 6, vest, "#777")
		if has["high-visibility vest"] {
			s.rect(x-18, 126, 36, 5, 0, "#bdbdbd", "")
		}
		s.rect(x-16, 160, 12, 40, 3, "#455a64", "")
		s.rect(x+4, 160, 12, 40, 3, "#455a64", "")
		if has["hard hat"] {
			s.raw(`<path d="M%.0f 88 a16 14 0 0 1 32 0 h4 v4 h-40 v-4 Z" fill="#fbc02d" stroke="#a07800"/>`, x-16)
		}
		if has["safety glasses"] {
			s.rect(x-11, 88, 22, 6, 3, "#4fc3f7", "#0277bd")
		}
		if has["hearing protection"] {
			s.rect(x-17, 88, 5, 10, 2, "#e53935", "")
			s.rect(x+12, 88, 5, 10, 2, "#e53935", "")
		}
		hand := "#e8d5c0"
		if has["gloves"] {
			hand = "#ff7043"
		}
		s.circle(x-22, 158, 5, hand, "#999")
		s.circle(x+22, 158, 5, hand, "#999")
		s.text(x, 222, 10, `text-anchor="middle" font-weight="700"`, strings.Fields(n)[0])
	}
	s.text(230, 248, 9, `text-anchor="middle" fill="#666"`, "Orange hands = gloves · red ear cups = hearing protection · blue band = glasses")
	q := Question{key: "ex/ppecheck", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Who is missing PPE required by the sign?", ex)
	var firsts []string
	for _, n := range names {
		firsts = append(firsts, strings.Fields(n)[0])
	}
	q.Choices, q.Correct = oneOf(r, firsts[bad], append(firsts, "No one"))
	return q
}

// ── warehouse walkways ─────────────────────────────────────────────────

func walkwaysExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := newSVG(460, 300, "Warehouse floor plan with walkways")
	s.rect(0, 0, 460, 300, 6, "#f5f5f0", "#888")
	// racks
	for i := 0; i < 4; i++ {
		s.rect(110+float64(i)*80, 40, 40, 170, 2, "#b0bec5", "#78909c")
	}
	// pedestrian walkway: along the bottom then up the right side
	s.rect(20, 230, 420, 26, 0, "#fff59d", "#f9a825")
	s.rect(414, 30, 26, 226, 0, "#fff59d", "#f9a825")
	s.text(230, 247, 9, `text-anchor="middle" fill="#8d6e00" font-weight="700"`, "PEDESTRIAN WALKWAY")
	s.text(56, 226, 10, `font-weight="700"`, "Dock door ↙")
	s.rect(20, 240, 30, 40, 2, "#5d4037", "")
	s.rect(392, 4, 60, 26, 3, "#1a3a5c", "")
	s.text(422, 21, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, "OFFICE")
	// forklift aisles marked
	s.text(190, 222, 9, `fill="#c62828"`, "Forklift aisles between racks")
	good := r.intn(4)
	cols := []string{"#1565c0", "#6a1b9a", "#00897b", "#d84315"}
	paths := []string{
		"M35 250 L427 250 L427 30",          // stays in walkway
		"M35 250 L190 250 L190 120 L427 30", // cuts through aisle
		"M35 250 L35 120 L360 120 L427 30",  // across rack area
		"M35 250 L270 250 L270 40 L427 30",  // up an aisle
	}
	order := sample(r, []int{1, 2, 3}, 3)
	assign := make([]int, 4) // label i → path index
	k := 0
	for i := range assign {
		if i == good {
			assign[i] = 0
		} else {
			assign[i] = order[k]
			k++
		}
	}
	for i, pi := range assign {
		s.raw(`<path d="%s" fill="none" stroke="%s" stroke-width="3" stroke-dasharray="7 4" opacity="0.85" transform="translate(%d %d)"/>`, paths[pi], cols[i], (i-1)*3, (i-1)*3)
	}
	for i := range assign {
		s.rect(60+float64(i)*70, 272, 60, 20, 4, cols[i], "")
		s.text(90+float64(i)*70, 286, 10, `text-anchor="middle" fill="#fff" font-weight="700"`, "Route "+lettered(i))
	}
	q := Question{key: "ex/walkways", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. %s needs to walk from the dock door to the office. Which route keeps them in the marked pedestrian walkway?", ex, sl["FIRST"])
	var labels []string
	for i := range assign {
		labels = append(labels, "Route "+lettered(i))
	}
	q.Choices, q.Correct = oneOf(r, labels[good], labels)
	return q
}

// ── earthquake: where to take cover ────────────────────────────────────

func quakeRoomExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := newSVG(440, 280, "Office floor plan for earthquake cover")
	s.rect(40, 30, 360, 220, 0, "#fafafa", "#333")
	// window wall (top)
	s.rect(40, 26, 360, 8, 0, "#90caf9", "#1565c0")
	s.text(220, 20, 9, `text-anchor="middle" fill="#1565c0"`, "Floor-to-ceiling windows")
	// tall bookcase on left wall
	s.rect(44, 70, 18, 120, 0, "#8d6e63", "")
	s.text(70, 66, 9, `fill="#5d4037"`, "Tall bookcase")
	// sturdy table in middle-right away from windows
	s.rect(250, 150, 110, 60, 4, "#a1887f", "#5d4037")
	s.text(305, 184, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, "Sturdy table")
	// door on bottom wall
	s.rect(170, 246, 50, 8, 0, "#fff", "#333")
	s.raw(`<path d="M170 250 A50 50 0 0 1 220 200" fill="none" stroke="#999" stroke-dasharray="3 3"/>`)
	s.text(195, 268, 9, `text-anchor="middle"`, "Doorway")
	spots := [][2]float64{{305, 180}, {75, 130}, {220, 48}, {195, 236}} // table, bookcase, window, doorway
	order := sample(r, rangeInts(4), 4)
	for i, k := range order {
		markSVG(&s.Builder, spots[k][0]+map[int]float64{0: -44}[k], spots[k][1], lettered(i))
	}
	q := Question{key: "ex/quakeroom", Exhibit: s.done()}
	right := ""
	var labels []string
	for i, k := range order {
		labels = append(labels, "Position "+lettered(i))
		if k == 0 {
			right = "Position " + lettered(i)
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. The shaking starts while %s is in this room. Where should they Drop, Cover, and Hold On?", ex, sl["FIRST"])
	q.Choices, q.Correct = oneOf(r, right, labels)
	return q
}

// ── extinguisher inspection tag ────────────────────────────────────────

func inspectionTagExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	now := r.between(4, 11) // current month index; months 0..now-1 should be punched
	missed := r.between(0, now-1)
	s := newSVG(240, 330, "Fire extinguisher inspection tag")
	s.raw(`<path d="M20 30 h200 v290 h-200 Z" fill="#fff59d" stroke="#c0a000" stroke-width="2"/>`)
	s.circle(120, 48, 9, "#fff", "#c0a000")
	s.text(120, 80, 13, `text-anchor="middle" font-weight="700" fill="#c62828"`, "INSPECTION RECORD")
	s.text(120, 96, 9, `text-anchor="middle"`, "Monthly visual check · punch the month")
	for i, m := range months {
		x := 44 + float64(i%3)*64
		y := 116 + float64(i/3)*42
		s.rect(x, y, 52, 32, 3, "#fff", "#999")
		s.text(x+24, y+21, 11, `text-anchor="middle" font-weight="700"`, m)
		if i < now && i != missed {
			s.circle(x+45, y+9, 4.5, "#333", "")
		}
	}
	s.text(120, 300, 9, `text-anchor="middle"`, "Annual maintenance: see back")
	q := Question{key: "ex/inspectiontag", Exhibit: s.done()}
	cur := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	q.Text = fmt.Sprintf("Refer to %s. It is now early %s. Which monthly inspection was missed?", ex, cur[now])
	var opts []string
	for i := 0; i < now; i++ {
		opts = append(opts, cur[i])
	}
	opts = append(opts, "None; every month has been punched")
	q.Choices, q.Correct = oneOf(r, cur[missed], opts)
	return q
}

// ── GHS shipping label ─────────────────────────────────────────────────

func chemLabelExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	elements := []string{"Product identifier", "Signal word", "Hazard statements", "Pictograms", "Precautionary statements", "Supplier name, address, and phone"}
	miss := r.intn(len(elements))
	prod := pick(r, []string{"KLEENRITE Degreaser 40", "Lakefront Toner Solvent", "ProShine Floor Stripper", "Badger Bleach Concentrate"})
	s := newSVG(440, 260, "Chemical container label")
	s.rect(2, 2, 436, 256, 6, "#fff", "#333")
	if miss != 0 {
		s.text(20, 34, 16, `font-weight="700"`, prod)
	}
	if miss != 1 {
		s.text(20, 60, 18, `font-weight="900" fill="#c62828"`, pick(r, []string{"DANGER", "WARNING"}))
	}
	if miss != 3 {
		for i := 0; i < 2; i++ {
			x := 300 + float64(i)*66
			s.raw(`<path d="M%.0f 20 l28 28 l-28 28 l-28 -28 Z" fill="#fff" stroke="#c62828" stroke-width="4"/>`, x+28)
			s.text(x+28, 54, 16, `text-anchor="middle" font-weight="900"`, "!")
		}
	}
	y := 92.0
	if miss != 2 {
		s.text(20, y, 10, `font-weight="700"`, "Hazard statements:")
		s.text(20, y+14, 10, "", "Causes skin irritation. Causes serious eye irritation. Harmful if inhaled.")
		y += 40
	}
	if miss != 4 {
		s.text(20, y, 10, `font-weight="700"`, "Precautionary statements:")
		s.text(20, y+14, 10, "", "Wear protective gloves and eye protection. Use only in a well-ventilated area.")
		s.text(20, y+28, 10, "", "IF IN EYES: Rinse cautiously with water for several minutes.")
		y += 52
	}
	if miss != 5 {
		s.text(20, 238, 9, `fill="#555"`, sl["VENDOR"]+" · 1200 N. Water St., Milwaukee, WI · (414) 555-0170")
	}
	q := Question{key: "ex/chemlabel", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. A shipped container's label must have six GHS elements. Which required element is missing from this label?", ex)
	q.Choices, q.Correct = oneOf(r, elements[miss], elements)
	return q
}

// ── SDS sections ───────────────────────────────────────────────────────

var sdsSections = []string{"Identification", "Hazard(s) identification", "Composition/information on ingredients", "First-aid measures", "Fire-fighting measures", "Accidental release measures", "Handling and storage", "Exposure controls/personal protection", "Physical and chemical properties", "Stability and reactivity", "Toxicological information", "Ecological information", "Disposal considerations", "Transport information", "Regulatory information", "Other information"}

func sdsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	var rows [][]string
	for i, t := range sdsSections {
		rows = append(rows, []string{fmt.Sprintf("Section %d", i+1), t})
	}
	prod := pick(r, []string{"KLEENRITE Degreaser 40", "Lakefront Toner Solvent", "ProShine Floor Stripper", "Badger Bleach Concentrate"})
	q := Question{key: "ex/sds", Exhibit: docExhibit("Safety Data Sheet — "+prod, "Table of contents", []string{"Section", "Title"}, rows, "Safety Data Sheets follow the same 16-section order.")}
	sc := pick(r, []struct {
		text string
		sec  int
	}{
		{"%s spilled a gallon in the %s. Which section explains how to contain and clean up the spill?", 6},
		{"%s splashed some in their eyes. Which section gives first-aid instructions?", 4},
		{"%s needs to know which gloves and respirator to wear. Which section covers personal protection?", 8},
		{"%s is setting up storage in the %s. Which section covers safe storage?", 7},
		{"A small fire involving the product breaks out. Which section tells responders %s and the floor warden how to fight it?", 5},
		{"%s needs to dispose of an expired drum. Which section covers disposal?", 13},
	})
	q.Text = fmt.Sprintf("Refer to %s. "+sc.text, ex, sl["NAME2"], sl["PLACE"])
	if strings.Count(sc.text, "%s") == 1 {
		q.Text = fmt.Sprintf("Refer to %s. "+sc.text, ex, sl["NAME2"])
	}
	var opts []string
	for _, k := range []int{2, 4, 5, 6, 7, 8, 11, 13, 16} {
		opts = append(opts, fmt.Sprintf("Section %d", k))
	}
	q.Choices, q.Correct = oneOf(r, fmt.Sprintf("Section %d", sc.sec), opts)
	return q
}

// ── power strip setups ─────────────────────────────────────────────────

func powerStripsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type setup struct {
		desc []string // chain from the wall
		ok   bool
	}
	setups := []setup{
		{[]string{"Wall outlet", "Power strip", "Laptop + monitor"}, true},
		{[]string{"Wall outlet", "Power strip", "Power strip", "Desk lamp"}, false},
		{[]string{"Wall outlet", "Power strip", "Space heater"}, false},
		{[]string{"Wall outlet", "Extension cord (under rug)", "Printer"}, false},
		{[]string{"Wall outlet", "Extension cord", "Power strip", "Mini fridge"}, false},
	}
	pickS := append([]setup{setups[0]}, sample(r, setups[1:], 3)...)
	shuffle(r, pickS)
	s := newSVG(460, 60+4*52, "Four desk power setups")
	s.rect(0, 0, 460, 60+4*52, 8, "#fff", "#ccc")
	s.text(14, 24, 12, `font-weight="700" fill="#1a3a5c"`, "Desk power setups, "+sl["FLOOR"]+" floor")
	right := ""
	for i, st := range pickS {
		y := 44 + float64(i)*52
		s.text(18, y+24, 13, `font-weight="700" fill="#1a3a5c"`, lettered(i))
		x := 40.0
		for j, d := range st.desc {
			w := 6.5*float64(len(d)) + 16
			col := "#eceff1"
			if j == 0 {
				col = "#cfd8dc"
			}
			s.rect(x, y+8, w, 26, 4, col, "#90a4ae")
			s.text(x+w/2, y+25, 9, `text-anchor="middle"`, d)
			if j < len(st.desc)-1 {
				s.line(x+w, y+21, x+w+14, y+21, "#455a64", 2)
			}
			x += w + 14
		}
		if st.ok {
			right = "Setup " + lettered(i)
		}
	}
	q := Question{key: "ex/powerstrips", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which desk power setup follows ACPWB's electrical safety rules?", ex)
	q.Choices, q.Correct = oneOf(r, right, []string{"Setup A", "Setup B", "Setup C", "Setup D"})
	return q
}

// ── suspicious package ─────────────────────────────────────────────────

func suspiciousPackageExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	signs := []string{"postage", "noreturn", "restrictive", "misspelled", "stain", "wire"}
	have := sample(r, signs, r.between(1, 5))
	on := func(k string) bool { return contains(have, k) }
	s := newSVG(440, 280, "Package on a mailroom table")
	s.rect(0, 0, 440, 280, 8, "#e9e4da", "")
	s.raw(`<path d="M60 70 L380 52 L392 232 L52 246 Z" fill="#c8a27a" stroke="#8d6e4f" stroke-width="2"/>`)
	if on("stain") {
		s.raw(`<ellipse cx="300" cy="190" rx="34" ry="20" fill="#7a5a32" opacity="0.55"/>`)
	}
	n := 1
	if on("postage") {
		n = 9
	}
	for i := 0; i < n; i++ {
		x := 300 + float64(i%3)*24
		y := 66 + float64(i/3)*26
		s.rect(x, y, 20, 22, 1, pick(r, []string{"#e57373", "#64b5f6", "#81c784", "#ffd54f"}), "#fff")
	}
	if !on("noreturn") {
		s.text(76, 96, 9, `font-family="Courier New,monospace"`, "FROM: "+sl["VENDOR"])
		s.text(76, 108, 9, `font-family="Courier New,monospace"`, "      1200 N. Water St, Milwaukee WI")
	}
	to := "Chief Executive Officer"
	if on("misspelled") {
		to = "Cheif Executve Oficer"
	}
	s.text(150, 150, 12, `font-family="Courier New,monospace" font-weight="700"`, "TO: "+to)
	hq := "ACPWB HEADQUARTERS"
	if on("misspelled") {
		hq = "ACPWB HEADQUATERS"
	}
	s.text(150, 166, 12, `font-family="Courier New,monospace" font-weight="700"`, "    "+hq)
	if on("restrictive") {
		s.raw(`<g transform="rotate(-8 140 210)">`)
		s.rect(80, 196, 150, 26, 2, "none", "#c62828")
		s.text(155, 214, 12, `text-anchor="middle" font-weight="900" fill="#c62828"`, "PERSONAL · CONFIDENTIAL")
		s.raw(`</g>`)
	}
	if on("wire") {
		s.raw(`<path d="M380 120 q22 4 30 -14 q6 -12 14 -8" fill="none" stroke="#424242" stroke-width="2.5"/>`)
	}
	s.text(220, 268, 9, `text-anchor="middle" fill="#555"`, "Arrived by courier · no tracking number")
	q := Question{key: "ex/suspiciouspkg", Exhibit: s.done()}
	count := len(have)
	q.Text = fmt.Sprintf("Refer to %s. Using the Postal Service's list of suspicious mail indicators (excessive postage, no return address, restrictive markings, misspellings, stains, protruding wires), how many indicators does this package show?", ex)
	q.Choices, q.Correct = countQ(r, count, "indicator")
	return q
}

// ── monitor setup ──────────────────────────────────────────────────────

func monitorSetupExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// setups: 0 correct, 1 too high, 2 too low (laptop flat), 3 too far
	order := sample(r, rangeInts(4), 4)
	s := newSVG(460, 210, "Four monitor setups, side view")
	s.rect(0, 0, 460, 210, 8, "#fff", "#ccc")
	right := ""
	for i, k := range order {
		x := 12 + float64(i)*112
		s.rect(x, 10, 104, 170, 6, "#f5f7fa", "#d0d6de")
		s.text(x+10, 28, 12, `font-weight="700" fill="#1a3a5c"`, lettered(i))
		// seated person: head, torso, chair
		hx, hy := x+30, 70.0
		s.circle(hx, hy, 10, "#e8d5c0", "#999")
		s.rect(hx-8, hy+10, 16, 46, 4, "#5c6bc0", "")
		s.rect(hx-12, hy+56, 30, 6, 2, "#424242", "")
		s.line(hx+2, hy+62, hx+2, 170, "#424242", 3)
		s.dash(hx+10, hy-2, x+100, hy-2, "#c62828", 0.8) // eye level
		s.line(hx+6, hy+22, hx+34, hy+30, "#5c6bc0", 4)  // arm, about an arm's length
		s.circle(hx+35, hy+30, 3, "#e8d5c0", "")
		// desk
		s.rect(x+50, 130, 52, 5, 1, "#8d6e63", "")
		s.line(x+96, 135, x+96, 170, "#8d6e63", 3)
		var top, dist float64
		switch k {
		case 0:
			top, dist = hy+2, 30
		case 1:
			top, dist = hy-30, 30
		case 2:
			top, dist = 118, 26
		case 3:
			top, dist = hy+2, 66
		}
		mx := hx + dist
		if k == 2 {
			s.rect(mx, 126, 24, 4, 1, "#37474f", "")
			s.raw(`<path d="M%.0f 126 l-6 -12" stroke="#37474f" stroke-width="3"/>`, mx)
		} else {
			s.rect(mx, top, 6, 34, 1, "#37474f", "")
			s.line(mx+3, top+34, mx+3, 130, "#78909c", 2)
		}
		if k == 0 {
			right = "Setup " + lettered(i)
		}
	}
	s.text(230, 200, 9, `text-anchor="middle" fill="#c62828"`, "Dashed line = eye level")
	q := Question{key: "ex/monitorsetup", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which setup places the monitor correctly, with the top of the screen at or slightly below eye level and about an arm's length away?", ex)
	q.Choices, q.Correct = oneOf(r, right, []string{"Setup A", "Setup B", "Setup C", "Setup D"})
	return q
}

// ── room occupancy ─────────────────────────────────────────────────────

func occupancyExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	maxOcc := pick(r, []int{12, 16, 20, 24, 30, 49})
	attending := maxOcc + r.between(-6, 8)
	if attending == maxOcc {
		attending++
	}
	room := pick(r, []string{"Lakeview Room", "Room 40-12", "Pfister Boardroom", "Cream City Room", "Training Room B"})
	s := newSVG(440, 220, "Occupancy sign and meeting invite")
	s.rect(0, 0, 440, 220, 8, "#eceff1", "")
	s.rect(20, 20, 150, 180, 6, "#fff", "#c62828")
	s.text(95, 48, 12, `text-anchor="middle" font-weight="700" fill="#c62828"`, "MAXIMUM")
	s.text(95, 64, 12, `text-anchor="middle" font-weight="700" fill="#c62828"`, "OCCUPANCY")
	s.text(95, 130, 52, `text-anchor="middle" font-weight="900" fill="#c62828"`, fmt.Sprint(maxOcc))
	s.text(95, 170, 9, `text-anchor="middle"`, room)
	s.text(95, 186, 8, `text-anchor="middle" fill="#666"`, "By order of the Fire Marshal")
	s.rect(190, 30, 230, 160, 8, "#fff", "#bbb")
	s.text(204, 54, 12, `font-weight="700" fill="#1a3a5c"`, sl["COMMITTEE"])
	s.text(204, 72, 10, `fill="#666"`, room+" · "+pick(r, []string{"Tue", "Wed", "Thu"})+" "+clock(r.between(9, 15)*60))
	s.text(204, 104, 11, "", fmt.Sprintf("Accepted: %d", attending))
	s.text(204, 124, 11, "", fmt.Sprintf("Declined: %d", r.between(0, 9)))
	s.text(204, 144, 11, "", fmt.Sprintf("No response: %d", r.between(0, 6)))
	s.text(204, 172, 9, `fill="#666"`, "Organizer: "+sl["NAME"])
	q := Question{key: "ex/occupancy", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. If everyone who accepted attends in person, does the meeting fit the room's posted occupancy?", ex)
	if attending <= maxOcc {
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Yes; %d accepted, within the limit of %d", attending, maxOcc), []string{fmt.Sprintf("No; it is over by %d", maxOcc-attending+1), "Only if the door is propped open", "Only if half the attendees stand"})
		return q
	}
	q.Choices, q.Correct = withChoices(r, fmt.Sprintf("No; %d accepted, which is %d over the limit of %d, so move rooms or add a remote option", attending, attending-maxOcc, maxOcc),
		[]string{fmt.Sprintf("Yes; %d is close enough to %d", attending, maxOcc), "Yes, as long as the extra people stand near the door", "Yes, because occupancy signs are recommendations"})
	return q
}

// ── fire drill times ───────────────────────────────────────────────────

func fireDrillExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	target := pick(r, []int{4, 5, 6})
	floors := sample(r, []int{12, 14, 18, 21, 23, 27, 31, 33, 36, 40}, 5)
	sortInts(floors)
	var rows [][]string
	slow := 0
	worst, worstT := 0, -1
	used := map[int]bool{}
	for _, f := range floors {
		sec := r.between(150, target*60+150)
		for used[sec] {
			sec++
		}
		used[sec] = true
		if sec > target*60 {
			slow++
		}
		if sec > worstT {
			worst, worstT = f, sec
		}
		rows = append(rows, []string{ordinal(f), fmt.Sprintf("%d:%02d", sec/60, sec%60), fmt.Sprint(r.between(18, 140))})
	}
	q := Question{key: "ex/firedrill", Exhibit: docExhibit("Fire Drill Results — "+exDate(r).Format("January 2006"), fmt.Sprintf("Target: each floor clear in %d:00 or less", target), []string{"Floor", "Time to clear", "People evacuated"}, rows, "Floors over target repeat the drill next quarter.")}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. How many floors must repeat the drill?", ex)
		q.Choices, q.Correct = countQ(r, slow, "floor")
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Which floor took the longest to clear?", ex)
	var opts []string
	for _, f := range floors {
		opts = append(opts, "The "+ordinal(f)+" floor")
	}
	q.Choices, q.Correct = oneOf(r, "The "+ordinal(worst)+" floor", opts)
	return q
}
