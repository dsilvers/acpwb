package compliance

import (
	"fmt"
	"math"
	"strings"
)

// Diagram exhibits: drawings where the answer is read off the picture —
// where to aim, which stack is too high, which route to take. The safety
// ones are scored on real guidance (NFPA 10 and 13, NEC 110.26, ANSI
// Z358.1, NWS, NSC, ADA 405.2, OSHA 1910.147); the drawing decides which
// option is right.

func init() {
	for k, g := range map[string]exhibitGen{
		"extaim":     extAimExhibit,
		"sprinkler":  sprinklerExhibit,
		"panelclear": panelClearExhibit,
		"extheight":  extHeightExhibit,
		"windleak":   windLeakExhibit,
		"tornado":    tornadoExhibit,
		"shelfload":  shelfLoadExhibit,
		"following":  followingExhibit,
		"screenview": screenViewExhibit,
		"netmap":     netMapExhibit,
		"venn":       vennExhibit,
		"pie":        pieExhibit,
		"worldclock": worldClockExhibit,
		"stepladder": stepladderExhibit,
		"eyewash":    eyewashExhibit,
		"elevload":   elevLoadExhibit,
		"clearance":  clearanceExhibit,
		"cords":      cordsExhibit,
		"stairs":     stairsExhibit,
		"lightning":  lightningExhibit,
		"ramp":       rampExhibit,
		"badgezones": badgeZonesExhibit,
		"lockpoints": lockPointsExhibit,
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
		"fire":           {"extaim", "extheight", "sprinkler", "stairs"},
		"safety":         {"cords", "shelfload", "panelclear", "extaim", "eyewash"},
		"electrical":     {"panelclear", "cords"},
		"warehouse":      {"shelfload", "sprinkler", "clearance"},
		"officemove":     {"shelfload", "elevload", "sprinkler"},
		"elevator":       {"elevload", "stairs"},
		"hazcom":         {"eyewash", "windleak"},
		"spaces":         {"windleak"},
		"ppe":            {"eyewash"},
		"weather":        {"tornado", "lightning"},
		"closures":       {"tornado", "lightning"},
		"fieldwork":      {"lightning", "following", "clearance"},
		"sports":         {"lightning"},
		"offsites":       {"lightning", "venn"},
		"driving":        {"following", "clearance"},
		"deliveries":     {"clearance"},
		"parking":        {"clearance"},
		"commuting":      {"following"},
		"ladder":         {"stepladder"},
		"injuries":       {"cords", "stepladder"},
		"lockout":        {"lockpoints"},
		"construction":   {"lockpoints", "windleak"},
		"evacassist":     {"stairs", "ramp"},
		"accessibility":  {"ramp"},
		"accommodation":  {"ramp"},
		"openoffice":     {"screenview", "cords"},
		"cleandesk":      {"screenview"},
		"privacy":        {"screenview"},
		"privacy2":       {"screenview", "netmap"},
		"physicalsec":    {"badgezones", "screenview"},
		"badge":          {"badgezones"},
		"serverroom":     {"badgezones", "panelclear"},
		"visitor":        {"badgezones"},
		"network":        {"netmap"},
		"devices":        {"netmap"},
		"acceptableuse":  {"netmap"},
		"security":       {"netmap", "badgezones"},
		"globalteams":    {"worldclock"},
		"calendar":       {"worldclock"},
		"remote":         {"worldclock"},
		"intltravel":     {"worldclock"},
		"learning":       {"venn", "pie"},
		"managers":       {"venn"},
		"onboarding":     {"venn"},
		"finance":        {"pie"},
		"audit":          {"pie"},
		"spreadsheets":   {"pie", "venn"},
		"sustainability": {"pie"},
		"coffee":         {"pie"},
		"celebrations":   {"venn"},
		// banks that had no exhibit of their own
		"ai":             {"pie", "venn", "chatthread"},
		"aihiring":       {"venn", "jobposting"},
		"airquality":     {"pie"},
		"bedbugs":        {"pie", "worldclock"},
		"brand":          {"pie", "venn"},
		"careers":        {"orgchart", "venn"},
		"chatbots":       {"chatthread", "pie"},
		"civility":       {"chatthread", "replyall"},
		"classification": {"venn", "orgchart"},
		"conduct":        {"venn", "screenview"},
		"conflicts":      {"orgchart", "venn"},
		"crypto":         {"pie", "netmap"},
		"customers":      {"pie", "chatthread"},
		"deferredcomp":   {"vesting", "pie"},
		"dilemmas":       {"venn", "screenview"},
		"disclosure":     {"pie", "screenview"},
		"drones":         {"badgezones", "lightning"},
		"eeo":            {"venn", "jobposting"},
		"emoji":          {"chatthread", "replyall"},
		"focus":          {"pie", "videogrid"},
		"gambling":       {"pie"},
		"harassment":     {"chatthread", "venn"},
		"innovation":     {"pie", "kanban"},
		"insurance":      {"pie"},
		"investigations": {"badgezones", "screenview", "orgchart"},
		"kids":           {"badgezones", "venn"},
		"languages":      {"worldclock", "venn"},
		"networking":     {"venn", "orgchart"},
		"offboarding":    {"netmap", "badgezones"},
		"pandemic":       {"pie", "worldclock"},
		"photos":         {"screenview"},
		"political":      {"pie"},
		"presentations":  {"pie", "venn"},
		"relationships":  {"orgchart"},
		"sanctions":      {"worldclock"},
		"signage":        {"extheight", "stairs"},
		"signatures":     {"chatthread"},
		"sox":            {"pie", "venn"},
		"substances":     {"pie"},
		"supplychain":    {"pie", "worldclock"},
		"trivia":         {"worldclock", "pie", "venn"},
		"volunteering":   {"venn", "pie"},
		"vr":             {"cords", "screenview"},
		"whistleblower":  {"orgchart", "chatthread"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// frame starts a white diagram with a bold title.
func frame(w, h float64, label, title string) *svgB {
	s := newSVG(w, h, label)
	s.rect(0.5, 0.5, w-1, h-1, 0, "#fff", "#c8ccd4")
	s.text(w/2, 18, 10.5, `text-anchor="middle" font-weight="700" fill="#0a1628"`, title)
	return s
}

func markAt(s *svgB, x, y float64, i int) { markSVG(&s.Builder, x, y, markLetters[i]) }

// letterQ builds "<noun> A" … choices for n lettered options, right being
// the index of the correct letter.
func letterQ(r *rnd, noun string, n, right int) ([]string, int) {
	var wrongs []string
	for i := 0; i < n; i++ {
		if i != right {
			wrongs = append(wrongs, noun+" "+markLetters[i])
		}
	}
	return withChoices(r, noun+" "+markLetters[right], wrongs)
}

// ftIn formats inches as "6 ft 8 in".
func ftIn(in int) string {
	if in < 12 {
		return fmt.Sprintf("%d in", in)
	}
	if in%12 == 0 {
		return fmt.Sprintf("%d ft", in/12)
	}
	return fmt.Sprintf("%d ft %d in", in/12, in%12)
}

// orderN returns 0..n-1 shuffled: order[i] is what letter i marks.
func orderN(r *rnd, n int) []int {
	o := make([]int, n)
	for i := range o {
		o[i] = i
	}
	shuffle(r, o)
	return o
}

// ── where to aim an extinguisher (PASS: aim at the base) ────────────────

func extAimExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(420, 230, "Wastebasket fire with four aiming points", "PORTABLE EXTINGUISHER — WHERE TO AIM")
	s.line(20, 205, 400, 205, "#8a8f98", 2)
	s.raw(`<ellipse cx="200" cy="52" rx="64" ry="15" fill="#cbd5e1" opacity=".85"/>`)
	s.raw(`<path d="M168 150 Q174 108 188 124 Q192 74 204 116 Q216 86 220 122 Q232 104 232 150 Z" fill="#f59e0b" stroke="#d97706"/>`)
	s.raw(`<path d="M180 150 Q186 124 194 134 Q200 110 208 136 Q216 126 222 150 Z" fill="#fde047"/>`)
	s.raw(`<path d="M160 205 L167 148 L233 148 L240 205 Z" fill="#9aa5b1" stroke="#5b6670"/>`)
	s.rect(52, 122, 26, 82, 9, "#c62828", "#7f1d1d")
	s.rect(58, 110, 14, 14, 2, "#333", "")
	s.raw(`<path d="M72 114 Q104 100 116 136" fill="none" stroke="#333" stroke-width="3"/>`)
	s.text(65, 220, 8, `text-anchor="middle" fill="#666"`, "Class ABC")
	pts := []float64{150, 126, 88, 52} // base of the fire (right), flames, flame tips, smoke
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		y := pts[o]
		s.circle(200, y, 3, "#c62828", "")
		s.dash(206, y, 330, y, "#555", 1)
		markAt(s, 342, y, i)
		if o == 0 {
			right = i
		}
	}
	q := Question{key: "ex/extaim", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. %[2]s is using the PASS technique on this wastebasket fire. Where should the nozzle be aimed?",
		"Refer to %[1]s. Which marked point is the correct place to aim a portable fire extinguisher?",
		"Refer to %[1]s. A small fire has started in a wastebasket near %[2]s's desk, and the way out is clear behind them. Where should they aim the extinguisher?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = letterQ(r, "Point", 4, right)
	return q
}

// ── storage clearance below sprinklers (18 in, NFPA 13) ──────────────────

func sprinklerExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	defl := pick(r, []int{108, 114, 120, 126})
	bad := r.intn(4)
	tops := make([]int, 4)
	for i := range tops {
		if i == bad {
			tops[i] = defl - r.between(4, 14)
		} else {
			tops[i] = defl - r.between(24, 54)
		}
	}
	s := frame(460, 265, "Storage room elevation with sprinkler heads and four box stacks", "STORAGE ROOM — CLEARANCE BELOW SPRINKLERS")
	floor := 232.0
	k := 172.0 / float64(defl+10)
	y := func(in int) float64 { return floor - float64(in)*k }
	ceil := y(defl + 8)
	s.text(24, 40, 8.5, `fill="#444"`, "Sprinkler deflectors: "+ftIn(defl)+" above the floor")
	s.line(20, ceil, 440, ceil, "#555", 2)
	s.line(40, ceil+4, 420, ceil+4, "#c62828", 3)
	s.line(20, floor, 440, floor, "#555", 2)
	for i := 0; i < 4; i++ {
		x := 90 + float64(i)*94
		s.line(x, ceil+4, x, y(defl), "#c62828", 2)
		s.raw(`<path d="M%.1f %.1f l-8 5 h16 z" fill="#b45309"/>`, x, y(defl))
		h := float64(tops[i]) * k
		nb := 3 + r.intn(3)
		bh := h / float64(nb)
		for j := 0; j < nb; j++ {
			s.rect(x-28, floor-float64(j+1)*bh, 56, bh, 1, "#d6b98c", "#8a6d45")
		}
		s.rect(x-26, floor-h+3, 52, 14, 2, "#fff", "")
		s.text(x, floor-h+13, 8.5, `text-anchor="middle" font-weight="700"`, ftIn(tops[i]))
		s.text(x, floor+16, 9, `text-anchor="middle" font-weight="700" fill="#0a1628"`, "STACK "+markLetters[i])
	}
	s.text(230, 258, 8, `text-anchor="middle" fill="#666"`, "Heights are to the top of each stack.")
	q := Question{key: "ex/sprinkler", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which stack must be lowered to keep the required clearance below the sprinkler deflectors?",
		"Refer to %s. Facilities is inspecting this storage room before the fire marshal's visit. Which stack is too close to the sprinklers?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Stack", 4, bad)
	return q
}

// ── working space in front of electrical panels (3 ft, NEC 110.26) ──────

func panelClearExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	items := sample(r, []string{"Mop bucket", "Paper cartons", "Rolling chair", "Decor bin", "Step stool", "Floor fan", "Recycling cart", "Spare monitors"}, 4)
	bad := r.intn(4)
	s := frame(440, 280, "Plan view of an electrical room with items at marked distances from the panels", "ELECTRICAL ROOM — PLAN VIEW")
	s.rect(20, 34, 400, 14, 0, "#9ca3af", "")
	for i := 0; i < 4; i++ {
		x := 60 + float64(i)*90
		s.rect(x-10, 48, 80, 12, 1, "#4b5563", "")
		s.text(x+30, 57, 7, `text-anchor="middle" fill="#fff"`, fmt.Sprintf("PANEL LP-%d", i+1))
	}
	s.text(420, 274, 7.5, `text-anchor="end" fill="#666"`, "120/208 V panels · distances from the panel fronts")
	for i, name := range items {
		d := r.between(40, 70)
		if i == bad {
			d = r.between(10, 30)
		}
		x := 60 + float64(i)*90
		y := 60 + float64(d)*2.5
		s.dash(x+30, 60, x+30, y, "#c62828", 1)
		s.rect(x+12, (60+y)/2-7, 36, 13, 2, "#fff", "")
		s.text(x+30, (60+y)/2+3, 8, `text-anchor="middle" font-weight="700" fill="#c62828"`, fmt.Sprintf("%d in", d))
		s.rect(x, y, 60, 22, 3, "#e5e7eb", "#6b7280")
		s.text(x+30, y+14, 7.5, `text-anchor="middle"`, name)
	}
	q := Question{key: "ex/panelclear", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which item is inside the working space that must be kept clear in front of the electrical panels?",
		"Refer to %[1]s. %[2]s is walking the electrical room before an inspection. Which item must be moved?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = oneOf(r, items[bad], items)
	return q
}

// ── extinguisher mounting height (NFPA 10: top ≤ 5 ft, bottom ≥ 4 in) ────

func extHeightExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	badTop := r.chance(0.6)
	s := frame(440, 280, "Wall elevation with four mounted fire extinguishers and their heights", "EXTINGUISHER MOUNTING — CORRIDOR WALL")
	floor := 240.0
	k := 2.4
	s.line(20, floor, 420, floor, "#555", 2)
	for i := 0; i < 4; i++ {
		h := r.between(18, 22)
		top := r.between(40, 58)
		if i == bad {
			if badTop {
				top = r.between(64, 76)
			} else {
				top = h + r.between(0, 2)
			}
		}
		bottom := top - h
		x := 60 + float64(i)*96
		yt, yb := floor-float64(top)*k, floor-float64(bottom)*k
		s.rect(x-3, yt-6, 6, 8, 1, "#333", "")
		s.rect(x-11, yt, 22, yb-yt, 7, "#c62828", "#7f1d1d")
		s.rect(x-6, yt-4, 12, 6, 1, "#222", "")
		s.text(x+16, yt+4, 7.5, `fill="#333"`, "top "+ftIn(top))
		s.text(x+16, yb, 7.5, `fill="#333"`, fmt.Sprintf("bottom %d in", bottom))
		s.text(x, floor+16, 9, `text-anchor="middle" font-weight="700" fill="#0a1628"`, markLetters[i])
	}
	s.text(220, 272, 8, `text-anchor="middle" fill="#666"`, "All four are 10 lb ABC extinguishers. Heights are measured from the floor.")
	q := Question{key: "ex/extheight", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which extinguisher is mounted at an incorrect height?",
		"Refer to %s. During the monthly fire-equipment check, which extinguisher should be re-hung?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Extinguisher", 4, bad)
	return q
}

// ── chemical release: go upwind ─────────────────────────────────────────

var compass8 = map[int]string{0: "east", 45: "southeast", 90: "south", 135: "southwest", 180: "west", 225: "northwest", 270: "north", 315: "northeast"}

func windLeakExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	th := 45 * r.intn(8) // the wind blows toward th (screen degrees, y down)
	s := frame(460, 290, "Site plan with a leak, the wind direction, and four assembly points", "SITE PLAN — CHEMICAL RELEASE")
	cx, cy := 260.0, 155.0
	rad := func(d int) float64 { return float64(d) * math.Pi / 180 }
	// wind box
	s.rect(14, 30, 92, 74, 4, "#f3f4f6", "#9ca3af")
	s.text(60, 44, 8, `text-anchor="middle" font-weight="700" fill="#374151"`, "WIND")
	ax, ay := math.Cos(rad(th)), math.Sin(rad(th))
	s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1d4ed8" stroke-width="3" marker-end="url(#wa)"/>`, 60-ax*22, 72-ay*22, 60+ax*18, 72+ay*18)
	s.raw(`<defs><marker id="wa" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto"><path d="M0 0 L8 4 L0 8 Z" fill="#1d4ed8"/></marker></defs>`)
	s.text(60, 100, 7, `text-anchor="middle" fill="#374151"`, "from the "+compass8[(th+180)%360])
	s.text(440, 40, 9, `text-anchor="end" font-weight="700" fill="#374151"`, "N ↑")
	// the release
	s.rect(cx-30, cy-12, 60, 24, 6, "#d1d5db", "#6b7280")
	s.raw(`<path d="M%.1f %.1f l6 -14 l4 10 l8 -8 l-2 12 l12 2 l-12 6 l6 10 l-12 -4 l-4 10 l-4 -10 l-12 4 l6 -10 l-12 -6 l12 -2 z" fill="#ef4444" opacity=".85"/>`, cx-2, cy-6)
	s.text(cx, cy+28, 8, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "LEAK")
	angles := []int{th + 180, th, th + 50, th - 50} // upwind first
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		a := rad(angles[o])
		x, y := cx+108*math.Cos(a), cy+108*math.Sin(a)
		s.circle(x, y, 13, "#16a34a", "#14532d")
		s.text(x, y+3, 8, `text-anchor="middle" font-weight="700" fill="#fff"`, "AP")
		markAt(s, x+22, y-12, i)
		if o == 0 {
			right = i
		}
	}
	chem := pick(r, []string{"ammonia", "chlorine", "an unknown chemical", "a cleaning concentrate"})
	q := Question{key: "ex/windleak", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. A delivery tanker is leaking %[2]s at the unloading area. Based on the wind, which assembly point should everyone go to?",
		"Refer to %[1]s. The alarm reports a release of %[2]s. %[3]s is told to move to safety. Which assembly point should they choose?",
	}), ex, chem, sl["FIRST"])
	q.Choices, q.Correct = letterQ(r, "Assembly Point", 4, right)
	return q
}

// ── tornado shelter: lowest floor, interior, no windows ──────────────────

func tornadoExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	basement := r.chance(0.5)
	ground := 250.0
	if basement {
		ground = 200
	}
	s := frame(460, 290, "Building cross-section with four marked locations", "BUILDING CROSS-SECTION — TORNADO SHELTER")
	s.rect(20, ground, 420, 290-ground-1, 0, "#e7dcc8", "")
	s.line(20, ground, 440, ground, "#6b5b3e", 2)
	// three floors above ground: left office (windows), core (no windows), right
	for f := 0; f < 3; f++ {
		top := ground - float64(f+1)*50
		s.rect(80, top, 300, 50, 0, "#f8fafc", "#475569")
		s.line(180, top, 180, top+50, "#475569", 1.5)
		s.line(280, top, 280, top+50, "#475569", 1.5)
		s.rect(78, top+12, 4, 24, 0, "#60a5fa", "")
		if f == 0 {
			s.rect(282, top+4, 96, 44, 0, "#bfdbfe", "#60a5fa") // glass lobby
		} else {
			s.rect(378, top+12, 4, 24, 0, "#60a5fa", "")
		}
	}
	s.raw(`<path d="M76 %.1f L230 %.1f L384 %.1f Z" fill="#94a3b8"/>`, ground-150, ground-172, ground-150)
	if basement {
		s.rect(80, ground, 300, 50, 0, "#f1f5f9", "#475569")
		s.line(180, ground, 180, ground+50, "#475569", 1.5)
		s.text(130, ground+30, 7.5, `text-anchor="middle" fill="#64748b"`, "Mechanical")
	}
	s.text(130, ground-8, 7.5, `text-anchor="middle" fill="#64748b"`, "Office")
	s.text(330, ground-8, 7.5, `text-anchor="middle" fill="#1e3a8a"`, "Glass lobby")
	s.text(230, ground-108, 7.5, `text-anchor="middle" fill="#64748b"`, "Restroom")
	s.text(130, ground-108, 7.5, `text-anchor="middle" fill="#64748b"`, "Office")
	type spot struct{ x, y float64 }
	shelter := spot{230, ground - 30}
	label := "Storage (interior)"
	if basement {
		shelter = spot{280, ground + 22}
		label = "Storage"
		s.text(280, ground+44, 7.5, `text-anchor="middle" fill="#64748b"`, label)
	} else {
		s.text(230, ground-8, 7.5, `text-anchor="middle" fill="#64748b"`, label)
	}
	spots := []spot{shelter, {230, ground - 130}, {130, ground - 30}, {330, ground - 30}}
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		markAt(s, spots[o].x, spots[o].y, i)
		if o == 0 {
			right = i
		}
	}
	s.text(440, 284, 7.5, `text-anchor="end" fill="#666"`, "Blue = windows and glass")
	q := Question{key: "ex/tornado", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. A tornado warning has been issued for the area. Which marked location is the best place to shelter?",
		"Refer to %[1]s. The alert says to take shelter now. Where should %[2]s go?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = letterQ(r, "Location", 4, right)
	return q
}

// ── heavy items between knee and shoulder ───────────────────────────────

func shelfLoadExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(440, 270, "Storage shelving with labeled box weights", "SUPPLY ROOM SHELVING")
	boards := []float64{236, 181, 126, 71} // floor, 30 in, 50 in, 72 in
	heights := []string{"floor", "30 in", "50 in", "72 in"}
	s.line(120, 40, 120, 240, "#6b7280", 3)
	s.line(340, 40, 340, 240, "#6b7280", 3)
	for i, y := range boards {
		s.rect(120, y, 220, 4, 0, "#6b7280", "")
		s.text(112, y+3, 7.5, `text-anchor="end" fill="#666"`, heights[i])
	}
	// candidates: top heavy (correct), top light, 30-in heavy, 50-in medium
	cands := []struct{ shelf, lb int }{{3, r.between(35, 55)}, {3, r.between(3, 8)}, {1, r.between(30, 50)}, {2, r.between(10, 15)}}
	filler := [][2]int{{5, 15}, {10, 30}, {5, 15}, {2, 6}}
	slots := [4][3]int{}
	for i := range slots {
		for j := range slots[i] {
			slots[i][j] = -1
		}
	}
	// top shelf holds two candidates; others one
	topSlots := sample(r, []int{0, 1, 2}, 2)
	slots[3][topSlots[0]], slots[3][topSlots[1]] = 0, 1
	slots[1][r.intn(3)] = 2
	slots[2][r.intn(3)] = 3
	order := orderN(r, 4)
	letterOf := make([]int, 4)
	for i, o := range order {
		letterOf[o] = i
	}
	for sh := 0; sh < 4; sh++ {
		for j := 0; j < 3; j++ {
			x := 132 + float64(j)*68
			y := boards[sh] - 38
			lb := r.between(filler[sh][0], filler[sh][1])
			ci := slots[sh][j]
			if ci >= 0 {
				lb = cands[ci].lb
			}
			s.rect(x, y, 58, 38, 2, "#d6b98c", "#8a6d45")
			s.text(x+29, y+23, 9, `text-anchor="middle" font-weight="700"`, fmt.Sprintf("%d lb", lb))
			if ci >= 0 {
				markAt(s, x+54, y+2, letterOf[ci])
			}
		}
	}
	q := Question{key: "ex/shelfload", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Which marked box should be moved to a lower shelf?",
		"Refer to %[1]s. %[2]s is reorganizing the supply room to reduce lifting injuries. Which marked box is stored in the wrong place?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = letterQ(r, "Box", 4, letterOf[0])
	return q
}

// ── following distance (three-second rule) ──────────────────────────────

func followingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(460, 250, "Four lanes, each with a car following a truck, speed and gap labeled", "FLEET TELEMATICS — FOLLOWING DISTANCE")
	for i := 0; i < 4; i++ {
		mph := pick(r, []int{30, 40, 45, 55, 60, 65, 70})
		fps := float64(mph) * 22 / 15
		sec := 3.6 + r.float()*1.8
		if i == bad {
			sec = 1.1 + r.float()*1.1
		}
		gap := int(math.Round(sec*fps/10)) * 10
		y := 34 + float64(i)*52
		s.rect(20, y, 420, 40, 0, "#4b5563", "")
		s.dash(20, y+20, 440, y+20, "#e5e7eb", 1)
		px := float64(gap) * 0.42
		s.rect(350, y+8, 70, 24, 3, "#e5e7eb", "#111")
		s.text(385, y+24, 7.5, `text-anchor="middle"`, "TRUCK")
		cx := 346 - px - 46
		s.rect(cx, y+11, 46, 18, 6, "#2563eb", "#1e3a8a")
		s.text(cx+23, y+24, 8.5, `text-anchor="middle" font-weight="700" fill="#fff"`, markLetters[i])
		s.line(cx+48, y+36, 348, y+36, "#fde047", 1)
		s.text((cx+48+348)/2, y+33, 8, `text-anchor="middle" fill="#fde047" font-weight="700"`, fmt.Sprintf("%d ft", gap))
		s.text(24, y+12, 8, `fill="#fff" font-weight="700"`, fmt.Sprintf("%d mph", mph))
	}
	s.text(230, 244, 7.5, `text-anchor="middle" fill="#666"`, "At 60 mph a vehicle travels 88 feet per second.")
	q := Question{key: "ex/following", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Four fleet drivers are each following a truck. Using the three-second rule, which driver is following too closely?",
		"Refer to %s. Telematics flagged one of these cars for tailgating. Which driver is under three seconds behind the truck?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Driver", 4, bad)
	return q
}

// ── which screen can be read from the corridor ──────────────────────────

func screenViewExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(440, 275, "Office plan with four desks, screen directions, and a glass corridor wall", "OPEN OFFICE — SCREEN VISIBILITY")
	s.rect(20, 30, 400, 206, 0, "#fafaf9", "#78716c")
	s.rect(20, 233, 400, 6, 0, "#93c5fd", "#3b82f6")
	s.text(220, 254, 8.5, `text-anchor="middle" font-weight="700" fill="#1d4ed8"`, "GLASS WALL · PUBLIC CORRIDOR")
	s.text(220, 268, 7.5, `text-anchor="middle" fill="#666"`, "Blue wedge = the direction each screen faces")
	bad := 2 + r.intn(2) // a bottom-row desk faces the glass
	pos := [][2]float64{{120, 80}, {320, 80}, {120, 175}, {320, 175}}
	names := sample(r, firstNames, 4)
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		x, y := pos[o][0], pos[o][1]
		dir := "up"
		if o == bad {
			dir = "down"
		} else if o < 2 {
			dir = pick(r, []string{"up", "left", "right"})
		}
		s.rect(x-45, y-20, 90, 40, 3, "#d6c7a1", "#8a7550")
		switch dir {
		case "up":
			s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#60a5fa" opacity=".45"/>`, x, y-12, x-26, y-46, x+26, y-46)
			s.rect(x-18, y-14, 36, 5, 1, "#111", "")
		case "down":
			s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#60a5fa" opacity=".45"/>`, x, y+12, x-26, y+46, x+26, y+46)
			s.rect(x-18, y+9, 36, 5, 1, "#111", "")
		case "left":
			s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#60a5fa" opacity=".45"/>`, x-30, y, x-74, y-20, x-74, y+20)
			s.rect(x-36, y-14, 5, 28, 1, "#111", "")
		case "right":
			s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#60a5fa" opacity=".45"/>`, x+30, y, x+74, y-20, x+74, y+20)
			s.rect(x+31, y-14, 5, 28, 1, "#111", "")
		}
		s.text(x, y+4, 8, `text-anchor="middle" fill="#3f3a2e"`, names[o])
		markAt(s, x+52, y-22, i)
		if o == bad {
			right = i
		}
	}
	q := Question{key: "ex/screenview", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which desk's screen can be read by people walking past in the public corridor?",
		"Refer to %s. Customer records are open on every screen in this room. Which desk needs a privacy filter or a turned monitor?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Desk", 4, right)
	return q
}

// ── personal device on the corporate network ────────────────────────────

func netMapExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type dev struct{ name, owner string }
	company := sample(r, []dev{{"Laptop", "ACPWB-issued"}, {"Printer", "network"}, {"Badge reader", "lobby"}, {"Room display", "conference"}, {"Desk phone", "VoIP"}, {"Scanner", "mailroom"}}, 3)
	personal := sample(r, []dev{{"Phone", "personal"}, {"Tablet", "personal"}, {"Smartwatch", "personal"}, {"Laptop", "visitor"}, {"Speaker", "personal"}, {"Game console", "personal"}}, 4)
	corp := append(append([]dev{}, company...), personal[0])
	shuffle(r, corp)
	guest := personal[1:]
	s := frame(460, 260, "Network map with corporate and guest networks and their connected devices", "NETWORK MAP — 3RD FLOOR")
	s.rect(190, 30, 80, 26, 4, "#1f2937", "")
	s.text(230, 47, 8.5, `text-anchor="middle" font-weight="700" fill="#fff"`, "FIREWALL")
	sw := func(x float64, label, fill string) {
		s.line(230, 56, x, 96, "#6b7280", 2)
		s.rect(x-70, 96, 140, 26, 4, fill, "")
		s.text(x, 113, 8.5, `text-anchor="middle" font-weight="700" fill="#fff"`, label)
	}
	sw(130, "CORPORATE LAN", "#1e3a8a")
	sw(370, "GUEST Wi-Fi", "#0f766e")
	draw := func(d dev, x, sx float64) {
		s.line(sx, 122, x, 172, "#9ca3af", 1.5)
		s.rect(x-24, 172, 48, 30, 4, "#f3f4f6", "#6b7280")
		s.text(x, 216, 8, `text-anchor="middle" font-weight="700"`, d.name)
		s.text(x, 227, 7, `text-anchor="middle" fill="#666"`, "("+d.owner+")")
	}
	for i, d := range corp {
		draw(d, 40+float64(i)*60, 130)
	}
	for i, d := range guest {
		draw(d, 310+float64(i)*58, 370)
	}
	s.text(230, 250, 7.5, `text-anchor="middle" fill="#444"`, "Policy: personal and visitor devices connect to Guest Wi-Fi only.")
	label := func(d dev) string { return "The " + d.owner + " " + strings.ToLower(d.name) }
	var wrongs []string
	for _, d := range company {
		wrongs = append(wrongs, label(d))
	}
	q := Question{key: "ex/netmap", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which device is connected to the wrong network?",
		"Refer to %s. The network team is reviewing the 3rd floor. Which corporate LAN device breaks the policy shown?",
	}), ex)
	q.Choices, q.Correct = withChoices(r, label(personal[0]), wrongs)
	return q
}

// ── Venn diagram of course assignments ──────────────────────────────────

func vennExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	pair := sample(r, []string{"Export Controls", "Anti-Bribery", "Data Privacy", "Records Retention", "Insider Trading", "Ladder Safety", "Clean Desk", "Hazard Communication", "Fire Warden", "Expense Policy"}, 2)
	nl, nb, nr, nn := r.between(2, 3), r.between(1, 3), r.between(2, 3), r.between(1, 2)
	names := sample(r, firstNames, nl+nb+nr+nn)
	onlyL, both, onlyR, neither := names[:nl], names[nl:nl+nb], names[nl+nb:nl+nb+nr], names[nl+nb+nr:]
	s := frame(480, 270, "Venn diagram of two course assignments with employee names", "WHO MUST COMPLETE WHICH COURSE")
	s.raw(`<circle cx="180" cy="152" r="96" fill="#3b82f6" fill-opacity=".14" stroke="#1d4ed8" stroke-width="2"/>`)
	s.raw(`<circle cx="290" cy="152" r="96" fill="#f59e0b" fill-opacity=".14" stroke="#b45309" stroke-width="2"/>`)
	s.text(150, 46, 8.5, `text-anchor="middle" font-weight="700" fill="#1d4ed8"`, "Must complete: "+pair[0])
	s.text(320, 46, 8.5, `text-anchor="middle" font-weight="700" fill="#b45309"`, "Must complete: "+pair[1])
	col := func(xs []string, x float64) {
		for i, n := range xs {
			s.text(x, 152+(float64(i)-float64(len(xs)-1)/2)*22, 9, `text-anchor="middle"`, n)
		}
	}
	col(onlyL, 128)
	col(both, 235)
	col(onlyR, 342)
	s.rect(408, 196, 66, 18+float64(len(neither))*14, 3, "#f9fafb", "#9ca3af")
	s.text(441, 208, 7, `text-anchor="middle" fill="#666"`, "Neither")
	for i, n := range neither {
		s.text(441, 222+float64(i)*14, 8.5, `text-anchor="middle"`, n)
	}
	q := Question{key: "ex/venn", Exhibit: s.done()}
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. How many employees must complete both %s and %s?", ex, pair[0], pair[1])
		q.Choices, q.Correct = countQ(r, nb, "employee")
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. Which employee must complete %s but not %s?", ex, pair[0], pair[1])
		var wrongs []string
		wrongs = append(wrongs, both[0], onlyR[0], neither[0])
		q.Choices, q.Correct = withChoices(r, onlyL[0], wrongs)
	default:
		q.Text = fmt.Sprintf("Refer to %s. How many employees must complete at least one of the two courses?", ex)
		q.Choices, q.Correct = countQ(r, nl+nb+nr, "employee")
	}
	return q
}

// ── pie chart ───────────────────────────────────────────────────────────

func pieExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type theme struct {
		title string
		cats  []string
		money bool
	}
	t := pick(r, []theme{
		{"FY TRAINING BUDGET", []string{"Compliance courses", "Leadership programs", "Safety training", "Conferences", "Tuition assistance"}, true},
		{"OFFICE SUPPLY SPEND", []string{"Paper", "Toner", "Coffee", "Sticky notes", "Lanyards"}, true},
		{"WHERE THE COMPLIANCE HOURS WENT", []string{"Watching videos", "Retaking quizzes", "Waiting for pages", "Finding the Attest button", "Actual learning"}, false},
		{"HOLIDAY PARTY BUDGET", []string{"Catering", "Venue", "Decorations", "Photo booth", "Ugly sweater prizes"}, true},
		{"IT HELP DESK TICKETS BY TYPE", []string{"Password resets", "Printer jams", "MFA lockouts", "Monitor cables", "Is it plugged in"}, false},
	})
	pcts := make([]int, 5)
	left := 100
	for i := 0; i < 4; i++ {
		pcts[i] = 2 * r.between(4, 13)
		left -= pcts[i]
	}
	pcts[4] = left
	colors := []string{"#1e3a8a", "#c9a227", "#0f766e", "#b91c1c", "#6b7280"}
	s := frame(460, 240, "Pie chart with five labeled categories", t.title)
	a0 := -math.Pi / 2
	for i, p := range pcts {
		a1 := a0 + float64(p)/100*2*math.Pi
		large := 0
		if p > 50 {
			large = 1
		}
		s.raw(`<path d="M130 130 L%.1f %.1f A90 90 0 %d 1 %.1f %.1f Z" fill="%s" stroke="#fff" stroke-width="1.5"/>`,
			130+90*math.Cos(a0), 130+90*math.Sin(a0), large, 130+90*math.Cos(a1), 130+90*math.Sin(a1), colors[i])
		a0 = a1
		y := 60 + float64(i)*30
		s.rect(250, y-10, 14, 14, 2, colors[i], "")
		s.text(272, y+1, 9, ``, fmt.Sprintf("%s — %d%%", t.cats[i], p))
	}
	q := Question{key: "ex/pie", Exhibit: s.done()}
	i, j := 0, 0
	for i == j {
		i, j = r.intn(5), r.intn(5)
	}
	if t.money && r.chance(0.5) {
		total := 10000 * r.between(5, 40)
		q.Text = fmt.Sprintf("Refer to %s. The total is %s. How much went to %s?", ex, dollars(total), strings.ToLower(t.cats[i]))
		amt := total * pcts[i] / 100
		q.Choices, q.Correct = dollarQ(r, amt, []int{-total / 20, total / 20, -total / 10, total / 10, total / 50})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What share is %s and %s combined?", ex, strings.ToLower(t.cats[i]), strings.ToLower(t.cats[j]))
	q.Choices, q.Correct = pctQ(r, pcts[i]+pcts[j])
	return q
}

// ── analog world clocks ─────────────────────────────────────────────────

func worldClockExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cities := sample(r, []string{"Milwaukee", "London", "Tokyo", "Sydney", "São Paulo", "Mumbai", "Dubai", "Toronto", "Johannesburg", "Singapore", "Berlin", "Denver"}, 4)
	type hm struct{ h, m int }
	var times []hm
	for len(times) < 4 {
		t := hm{r.between(1, 12), 5 * r.intn(12)}
		dup := false
		for _, u := range times {
			if u.h == t.h || u.m == t.m {
				dup = true
			}
		}
		if !dup && t.m/5 != t.h%12 {
			times = append(times, t)
		}
	}
	s := frame(460, 175, "Four analog wall clocks labeled with office cities", "OFFICE CLOCKS — LOCAL TIME")
	for i, t := range times {
		cx, cy := 62+float64(i)*112, 88.0
		s.circle(cx, cy, 42, "#fff", "#1f2937")
		for k := 0; k < 12; k++ {
			a := float64(k) * math.Pi / 6
			w := 1.0
			if k%3 == 0 {
				w = 2.5
			}
			s.line(cx+36*math.Sin(a), cy-36*math.Cos(a), cx+40*math.Sin(a), cy-40*math.Cos(a), "#1f2937", w)
		}
		ha := (float64(t.h%12) + float64(t.m)/60) * math.Pi / 6
		ma := float64(t.m) * math.Pi / 30
		s.line(cx, cy, cx+20*math.Sin(ha), cy-20*math.Cos(ha), "#111", 4)
		s.line(cx, cy, cx+33*math.Sin(ma), cy-33*math.Cos(ma), "#111", 2)
		s.circle(cx, cy, 3, "#c62828", "")
		s.text(cx, 150, 9.5, `text-anchor="middle" font-weight="700" fill="#0a1628"`, strings.ToUpper(cities[i]))
	}
	fmtT := func(t hm) string { return fmt.Sprintf("%d:%02d", t.h, t.m) }
	k := r.intn(4)
	q := Question{key: "ex/worldclock", Exhibit: s.done()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. The hallway clocks show local time in four ACPWB offices. Which office's clock reads %s?", ex, fmtT(times[k]))
		q.Choices, q.Correct = oneOf(r, cities[k], cities)
		return q
	}
	t := times[k]
	swapped := hm{t.m / 5, t.h * 5 % 60}
	if swapped.h == 0 {
		swapped.h = 12
	}
	cands := []string{fmtT(swapped), fmtT(hm{t.h%12 + 1, t.m}), fmtT(hm{t.h, (t.m + 30) % 60}), fmtT(hm{(t.h+10)%12 + 1, t.m})}
	q.Text = fmt.Sprintf("Refer to %s. What time does the %s clock show?", ex, cities[k])
	q.Choices, q.Correct = withChoices(r, fmtT(t), wrongsFrom(r, fmtT(t), cands, 3))
	return q
}

// ── stepladder: never the top cap or the top step ───────────────────────

func stepladderExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(320, 280, "Side view of a stepladder with marked positions", "STEPLADDER — STANDING POSITIONS")
	s.line(100, 262, 150, 56, "#d97706", 5)
	s.line(230, 262, 175, 56, "#d97706", 5)
	s.rect(140, 46, 46, 12, 2, "#92400e", "")
	s.line(160, 180, 205, 180, "#6b7280", 2)
	s.line(30, 264, 290, 264, "#555", 2)
	// rungs: top cap (0), top step (1), then 2, 3, 4 down the front rail
	ys := []float64{52, 96, 140, 184, 228}
	for i := 1; i < 5; i++ {
		x := 150 - (ys[i]-56)*50/206
		s.line(x-6, ys[i], x+22, ys[i], "#92400e", 5)
	}
	cands := []int{2, 0, 1, 3} // correct first: the step below the top step
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		y := ys[cands[o]]
		x := 150 - (y-56)*50/206
		if cands[o] == 0 {
			x = 140
		}
		s.dash(x-10, y, 52, y, "#555", 1)
		markAt(s, 40, y, i)
		if o == 0 {
			right = i
		}
	}
	s.text(160, 276, 7.5, `text-anchor="middle" fill="#666"`, "6 ft Type IA stepladder, fully open and locked")
	q := Question{key: "ex/stepladder", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. %[2]s needs to reach a high shelf. What is the highest marked position on this stepladder they may stand on?",
		"Refer to %[1]s. Which marked position is the highest safe place to stand on this stepladder?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = letterQ(r, "Position", 4, right)
	return q
}

// ── eyewash within 10 seconds, same level, unobstructed (ANSI Z358.1) ─────

func eyewashExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(460, 270, "Floor plan with a chemical hazard and four candidate eyewash stations with paths", "EYEWASH STATION PLACEMENT")
	hx, hy := 70.0, 140.0
	s.rect(hx-46, hy-24, 92, 48, 4, "#fee2e2", "#b91c1c")
	s.text(hx, hy-4, 8, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "CORROSIVES")
	s.text(hx, hy+10, 7.5, `text-anchor="middle" fill="#b91c1c"`, "battery charging")
	type kind struct {
		dist int
		note string
	}
	kinds := []kind{{r.between(20, 48), ""}, {r.between(85, 130), ""}, {r.between(20, 40), "through a locked door"}, {r.between(20, 40), "up one flight of stairs"}}
	slotsXY := [][2]float64{{250, 54}, {400, 104}, {400, 196}, {250, 236}}
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		p := slotsXY[i]
		k := kinds[o]
		s.dash(hx+46, hy, p[0]-16, p[1], "#6b7280", 1.5)
		mx, my := (hx+46+p[0]-16)/2, (hy+p[1])/2
		txt := fmt.Sprintf("%d ft", k.dist)
		if k.note != "" {
			txt += " · " + k.note
		}
		s.rect(mx-60, my-8, 120, 14, 3, "#fff", "")
		s.text(mx, my+3, 7.5, `text-anchor="middle" fill="#374151"`, txt)
		s.rect(p[0]-14, p[1]-14, 28, 28, 3, "#16a34a", "")
		s.text(p[0], p[1]+4, 9, `text-anchor="middle" font-weight="700" fill="#fff"`, "EW")
		markAt(s, p[0]+24, p[1]-14, i)
		if o == 0 {
			right = i
		}
	}
	q := Question{key: "ex/eyewash", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which marked eyewash station meets the guidance for reaching it within about 10 seconds from the hazard?",
		"Refer to %s. Facilities can keep only one of these eyewash stations. Which one should stay?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Station", 4, right)
	return q
}

// ── elevator rated load ─────────────────────────────────────────────────

func elevLoadExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cap := pick(r, []int{2500, 3000, 3500, 4000})
	bad := r.intn(4)
	s := frame(470, 240, "Elevator capacity plate and four moving loads with weights", "FREIGHT ELEVATOR — MOVE DAY")
	s.rect(18, 34, 112, 86, 4, "#d4d4d8", "#52525b")
	s.rect(24, 40, 100, 74, 2, "#e4e4e7", "#71717a")
	s.text(74, 60, 8, `text-anchor="middle" font-weight="700"`, "RATED LOAD")
	s.text(74, 80, 12, `text-anchor="middle" font-weight="700"`, commas(cap)+" LB")
	s.text(74, 100, 7.5, `text-anchor="middle"`, fmt.Sprintf("%d PERSONS", cap/150))
	mains := sample(r, []string{"Safe", "Copier", "Server rack", "Vending machine", "Pallet of paper", "Filing cabinets (6)", "Lobby piano", "Conference table"}, 4)
	for i := 0; i < 4; i++ {
		total := cap - r.between(2, 18)*50
		if i == bad {
			total = cap + r.between(1, 8)*50
		}
		movers := r.between(1, 2)
		mw := movers * r.between(17, 22) * 10
		jack := pick(r, []int{0, 150, 180})
		main := total - mw - jack
		y := 40 + float64(i)*48
		s.rect(146, y, 34, 26, 2, "#d6b98c", "#8a6d45")
		s.circle(152, y+30, 4, "#333", "")
		s.circle(174, y+30, 4, "#333", "")
		s.text(194, y+10, 9, `font-weight="700" fill="#0a1628"`, "LOAD "+markLetters[i])
		parts := fmt.Sprintf("%s %s lb", mains[i], commas(main))
		if jack > 0 {
			parts += fmt.Sprintf(" + pallet jack %d lb", jack)
		}
		parts += fmt.Sprintf(" + %s %s lb", plural(movers, "mover"), commas(mw))
		s.text(194, y+25, 8.5, ``, parts)
	}
	q := Question{key: "ex/elevload", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Each load rides up alone with its movers. Which load exceeds the elevator's rated capacity?",
		"Refer to %s. The move coordinator must split one load into two trips. Which one?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Load", 4, bad)
	return q
}

// ── parking garage height clearance ─────────────────────────────────────

func clearanceExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	clr := pick(r, []int{78, 80, 82, 84, 90})
	nbad := 1 + r.intn(2)
	types := []string{"Sedan", "SUV with roof box", "Pickup with ladder rack", "Cargo van"}
	hs := []int{r.between(56, 62), r.between(70, 76), 0, 0}
	for i := 2; i < 4; i++ {
		if i >= 4-nbad {
			hs[i] = clr + r.between(2, 14)
		} else {
			hs[i] = clr - r.between(3, 8)
		}
	}
	if hs[1] >= clr {
		hs[1] = clr - 6
	}
	order := orderN(r, 4)
	s := frame(470, 230, "Garage entrance clearance sign and four vehicles with heights", "PARKING GARAGE ENTRANCE")
	s.rect(18, 40, 86, 54, 4, "#facc15", "#713f12")
	s.text(61, 58, 8, `text-anchor="middle" font-weight="700"`, "CLEARANCE")
	s.text(61, 80, 12, `text-anchor="middle" font-weight="700"`, ftIn(clr))
	s.line(20, 205, 450, 205, "#555", 2)
	k := 1.3
	for i, o := range order {
		x := 120 + float64(i)*86
		h := float64(hs[o]) * k
		s.rect(x, 205-h, 72, h-10, 6, "#94a3b8", "#334155")
		s.rect(x+6, 205-h+6, 24, 14, 2, "#e0f2fe", "")
		s.circle(x+16, 200, 7, "#111", "")
		s.circle(x+56, 200, 7, "#111", "")
		lbl := ftIn(hs[o])
		if o >= 2 && r.chance(0.5) {
			lbl = fmt.Sprintf("%d in", hs[o])
		}
		s.text(x+36, 205-h-6, 8.5, `text-anchor="middle" font-weight="700"`, lbl)
		s.text(x+36, 220, 7.5, `text-anchor="middle" fill="#444"`, types[o])
	}
	q := Question{key: "ex/clearance", Exhibit: s.done()}
	if nbad == 1 {
		q.Text = fmt.Sprintf("Refer to %s. Which vehicle cannot safely enter the garage?", ex)
		q.Choices, q.Correct = oneOf(r, types[3], types)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. How many of these vehicles can safely enter the garage?", ex)
	q.Choices, q.Correct = numChoices(r, 4-nbad, []int{1, 2, -1, -2}, func(n int) string { return plural(n, "vehicle") })
	return q
}

// ── cords across a walkway ──────────────────────────────────────────────

func cordsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	s := frame(440, 230, "Office plan with a walkway and power cords, some under cord covers", "WALKWAY INSPECTION")
	s.rect(20, 92, 400, 50, 0, "#e5e7eb", "")
	s.text(30, 121, 9, `font-weight="700" fill="#6b7280"`, "WALKWAY")
	for i := 0; i < 4; i++ {
		s.rect(40+float64(i)*100, 34, 70, 34, 3, "#d6c7a1", "#8a7550")
		s.rect(40+float64(i)*100, 166, 70, 34, 3, "#d6c7a1", "#8a7550")
	}
	ncross := r.between(3, 6)
	nbare := r.between(1, minInt(4, ncross))
	xs := sample(r, []int{90, 140, 190, 240, 290, 340, 380}, ncross)
	bare := map[int]bool{}
	for _, i := range sample(r, orderN(r, ncross), nbare) {
		bare[i] = true
	}
	for i, x := range xs {
		fx := float64(x)
		dx := float64(r.between(-20, 20))
		s.raw(`<path d="M%.1f 68 C%.1f 100 %.1f 134 %.1f 166" fill="none" stroke="#111" stroke-width="2"/>`, fx, fx+dx, fx-dx, fx+dx/2)
		if !bare[i] {
			mx := fx + dx/4
			s.rect(mx-11, 90, 22, 54, 3, "#facc15", "#a16207")
			s.line(mx-11, 104, mx+11, 104, "#a16207", 1)
			s.line(mx-11, 130, mx+11, 130, "#a16207", 1)
		}
	}
	// cords that stay put
	for i := 0; i < r.between(1, 3); i++ {
		x := 50 + float64(r.intn(4))*100
		s.raw(`<path d="M%.1f 68 q10 14 24 6 q12 -6 22 8" fill="none" stroke="#111" stroke-width="2"/>`, x)
	}
	s.text(420, 222, 7.5, `text-anchor="end" fill="#666"`, "Yellow = cord cover")
	q := Question{key: "ex/cords", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. How many cords cross the walkway without a cord cover?",
		"Refer to %[1]s. %[2]s is doing the monthly walkthrough. How many trip hazards from cords should they report in the walkway?",
	}), ex, sl["FIRST"])
	q.Choices, q.Correct = countQ(r, nbare, "cord")
	return q
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── fire evacuation: clear stairwell, never the elevator ────────────────

func stairsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	floors := 6
	you := r.between(3, 6)
	fire := r.between(2, you)
	smokyLeft := r.chance(0.5)
	s := frame(460, 320, "Building cross-section with two stairwells, an elevator, and a roof door", "FIRE EVACUATION — ROUTE CHOICE")
	fy := func(f int) float64 { return 304 - float64(f)*40 } // floor line under floor f+1
	for f := 0; f <= floors; f++ {
		s.line(40, fy(f), 420, fy(f), "#64748b", 1.5)
	}
	s.line(40, fy(floors), 40, fy(0), "#334155", 2)
	s.line(420, fy(floors), 420, fy(0), "#334155", 2)
	cols := []float64{60, 380}
	for _, x := range cols {
		for f := 0; f < floors; f++ {
			s.raw(`<path d="M%.1f %.1f l12 -10 l12 0 l12 -10 l12 0" fill="none" stroke="#94a3b8" stroke-width="2"/>`, x-24, fy(f)-6)
		}
	}
	s.rect(212, fy(floors), 36, fy(0)-fy(floors), 0, "#e2e8f0", "#475569")
	s.text(230, fy(floors)-6, 7.5, `text-anchor="middle" fill="#475569"`, "ELEVATOR")
	s.text(60, fy(floors)-6, 7.5, `text-anchor="middle" fill="#475569"`, "STAIR 1")
	s.text(380, fy(floors)-6, 7.5, `text-anchor="middle" fill="#475569"`, "STAIR 2")
	s.rect(370, fy(floors)-30, 22, 18, 2, "#fef3c7", "#92400e")
	s.text(381, fy(floors)-36, 7, `text-anchor="middle" fill="#92400e"`, "ROOF")
	// fire and smoke
	s.raw(`<path d="M150 %.1f q8 -24 14 -8 q6 -20 12 0 q6 -12 8 8 z" fill="#f97316"/>`, fy(fire-1)-2)
	sx := 380.0
	if smokyLeft {
		sx = 60
	}
	for f := fire - 1; f < floors; f++ {
		s.raw(`<ellipse cx="%.1f" cy="%.1f" rx="22" ry="11" fill="#475569" opacity=".55"/>`, sx, fy(f)-20)
	}
	yy := fy(you-1) - 20
	s.raw(`<path d="M300 %.1f l3 7 l7 0 l-6 4 l2 7 l-6 -4 l-6 4 l2 -7 l-6 -4 l7 0 z" fill="#1d4ed8"/>`, yy-8)
	s.text(300, yy+22, 7, `text-anchor="middle" fill="#1d4ed8" font-weight="700"`, "YOU")
	clear := 380.0
	if !smokyLeft {
		clear = 60
	}
	spots := [][2]float64{{clear, yy}, {230, yy}, {sx, yy}, {381, fy(floors) - 21}}
	order := orderN(r, 4)
	right := 0
	for i, o := range order {
		markAt(s, spots[o][0]+22, spots[o][1]-6, i)
		if o == 0 {
			right = i
		}
	}
	q := Question{key: "ex/stairs", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. The fire alarm sounds while %[2]s is on floor %[3]d. Smoke has been reported in one stairwell. Which marked route should they take?",
		"Refer to %[1]s. There is a fire on floor %[4]d and you are on floor %[3]d. Which marked route is the right way out?",
	}), ex, sl["FIRST"], you, fire)
	q.Choices, q.Correct = letterQ(r, "Route", 4, right)
	return q
}

// ── lightning: wait 30 minutes after the last thunder (NWS) ─────────────

func lightningExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	start := 60 * r.between(12, 17)
	n := r.between(3, 5)
	var ts []int
	t := start + 5*r.between(1, 4)
	for i := 0; i < n; i++ {
		ts = append(ts, t)
		t += 5 * r.between(2, 5)
	}
	last := ts[len(ts)-1]
	s := frame(480, 170, "Timeline of thunder heard during an outdoor event", "OUTDOOR EVENT LOG — THUNDER HEARD")
	x := func(m int) float64 { return 30 + float64(m-start)*420/150 }
	s.line(30, 110, 450, 110, "#374151", 2)
	for m := start; m <= start+150; m += 15 {
		s.line(x(m), 106, x(m), 114, "#374151", 1.5)
		s.text(x(m), 130, 7.5, `text-anchor="middle" fill="#374151"`, clock(m))
	}
	for i, m := range ts {
		s.raw(`<path d="M%.1f 62 l-8 18 h7 l-5 16 l14 -22 h-7 l6 -12 z" fill="#facc15" stroke="#a16207"/>`, x(m)+2)
		y := 54.0
		if i%2 == 1 {
			y = 46
		}
		s.text(x(m), y, 7.5, `text-anchor="middle" font-weight="700"`, clock(m))
	}
	s.text(240, 156, 7.5, `text-anchor="middle" fill="#666"`, "No thunder has been heard since the last mark.")
	q := Question{key: "ex/lightning", Exhibit: s.done()}
	ev := pick(r, []string{"company softball game", "parking-lot food truck rally", "rooftop all-hands", "volunteer park cleanup", "outdoor team-building ropes course"})
	q.Text = fmt.Sprintf("Refer to %s. The %s moved indoors when thunder was first heard. If no more thunder is heard, what is the earliest time it can go back outside?", ex, ev)
	cands := []string{clock(last + 15), clock(last + 10), clock(last), clock(ts[0] + 30), clock(last + 60), clock(last + 20)}
	q.Choices, q.Correct = withChoices(r, clock(last+30), wrongsFrom(r, clock(last+30), cands, 3))
	return q
}

// ── ramp slope (ADA 405.2: 1:12 maximum) ────────────────────────────────

func rampExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(4)
	s := frame(460, 250, "Four ramps drawn in side view with rise and run labeled", "ENTRANCE RAMPS — SIDE VIEW")
	for i := 0; i < 4; i++ {
		rise := r.between(6, 30)
		var run int // feet
		if i == bad {
			run = int(math.Floor(float64(rise*r.between(7, 10)) / 12))
			if run < 1 {
				run = 1
			}
		} else {
			run = int(math.Ceil(float64(rise*r.between(12, 18)) / 12))
		}
		ox, oy := 30+float64(i%2)*220, 122+float64(i/2)*108
		w := 170.0
		h := w * float64(rise) / float64(run*12) * 3 // vertical exaggerated ×3
		if h > 70 {
			h = 70
		}
		s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#d6d3d1" stroke="#57534e" stroke-width="1.5"/>`, ox, oy, ox+w, oy, ox+w, oy-h)
		s.text(ox+w+6, oy-h/2+3, 8, `fill="#333"`, fmt.Sprintf("rise %d in", rise))
		s.text(ox+w/2, oy+14, 8, `text-anchor="middle" fill="#333"`, fmt.Sprintf("run %d ft", run))
		s.text(ox, oy-80, 9, `font-weight="700" fill="#0a1628"`, "RAMP "+markLetters[i])
	}
	s.text(230, 244, 7.5, `text-anchor="middle" fill="#666"`, "Vertical scale exaggerated.")
	q := Question{key: "ex/ramp", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which ramp is too steep to meet the ADA maximum running slope?",
		"Refer to %s. Facilities is reviewing entrance ramps before the renovation. Which one must be rebuilt?",
	}), ex)
	q.Choices, q.Correct = letterQ(r, "Ramp", 4, bad)
	return q
}

// ── badge access zones ──────────────────────────────────────────────────

func badgeZonesExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type zone struct{ name, fill, stroke string }
	zones := sample(r, []zone{{"GREEN", "#dcfce7", "#16a34a"}, {"BLUE", "#dbeafe", "#2563eb"}, {"ORANGE", "#ffedd5", "#ea580c"}, {"RED", "#fee2e2", "#dc2626"}, {"PURPLE", "#f3e8ff", "#9333ea"}}, 5)
	rooms := sample(r, []string{"Server Room", "Records Vault", "Executive Suite", "Mail Room", "Lab 2", "HR Files", "Finance", "Wellness Room", "Print Shop", "Archive"}, 5)
	s := frame(470, 250, "Floor plan with access zones and an employee badge", "BADGE ACCESS ZONES — 4TH FLOOR")
	pos := [][4]float64{{20, 34, 110, 90}, {130, 34, 110, 90}, {240, 34, 100, 90}, {20, 124, 160, 80}, {180, 124, 160, 80}}
	for i, p := range pos {
		s.rect(p[0], p[1], p[2], p[3], 0, zones[i].fill, zones[i].stroke)
		s.text(p[0]+p[2]/2, p[1]+p[3]/2-2, 8.5, `text-anchor="middle" font-weight="700"`, rooms[i])
		s.text(p[0]+p[2]/2, p[1]+p[3]/2+12, 7, `text-anchor="middle" fill="#444"`, "ZONE "+zones[i].name)
	}
	s.rect(20, 204, 320, 30, 0, "#f5f5f4", "#a8a29e")
	s.text(180, 223, 8, `text-anchor="middle" fill="#57534e"`, "LOBBY / CORRIDOR (all badges)")
	k := r.intn(5)
	s.rect(356, 50, 100, 150, 8, "#fff", "#1e3a8a")
	s.rect(356, 50, 100, 26, 8, "#1e3a8a", "")
	s.text(406, 68, 8.5, `text-anchor="middle" font-weight="700" fill="#fff"`, "ACPWB")
	s.circle(406, 104, 18, "#cbd5e1", "")
	s.text(406, 140, 8.5, `text-anchor="middle" font-weight="700"`, sl["NAME"])
	s.text(406, 160, 7, `text-anchor="middle" fill="#444"`, "ACCESS:")
	s.text(406, 174, 8, `text-anchor="middle" font-weight="700" fill="`+zones[k].stroke+`"`, "LOBBY · "+zones[k].name)
	q := Question{key: "ex/badgezones", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %s. Which room can %s's badge open?",
		"Refer to %s. %s needs to drop off a package without being escorted. Which room is their badge allowed into?",
	}), ex, sl["NAME"])
	q.Choices, q.Correct = oneOf(r, rooms[k], rooms)
	return q
}

// ── lockout: every energy source isolated and locked ────────────────────

func lockPointsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	machine := pick(r, []string{"carton sealer", "mail inserter", "cardboard baler", "dish machine", "trash compactor", "print finisher"})
	srcs := sample(r, []string{"Electrical disconnect", "Compressed-air valve", "Hydraulic pump switch", "Steam supply valve", "Control-panel breaker", "Water supply valve"}, 4)
	bad := r.intn(4)
	s := frame(470, 250, "Machine with four energy isolation points, locks applied to some", "LOCKOUT/TAGOUT — "+strings.ToUpper(machine))
	pos := [][2]float64{{60, 60}, {410, 60}, {60, 200}, {410, 200}}
	for _, p := range pos {
		s.line(p[0], p[1], 235, 135, "#94a3b8", 2)
	}
	s.rect(170, 90, 130, 90, 6, "#cbd5e1", "#334155")
	s.text(235, 140, 9, `text-anchor="middle" font-weight="700" fill="#334155"`, strings.ToUpper(machine))
	for i, p := range pos {
		s.rect(p[0]-32, p[1]-22, 64, 44, 4, "#f8fafc", "#475569")
		s.text(p[0], p[1]-4, 7.5, `text-anchor="middle" font-weight="700"`, "OFF")
		s.text(p[0], p[1]+32, 7.5, `text-anchor="middle" fill="#333"`, srcs[i])
		if i != bad {
			s.raw(`<path d="M%.1f %.1f a6 6 0 0 1 12 0 v6" fill="none" stroke="#111" stroke-width="2.5"/>`, p[0]-6, p[1]+4)
			s.rect(p[0]-9, p[1]+8, 18, 12, 2, "#dc2626", "")
			s.rect(p[0]+12, p[1]+2, 14, 18, 1, "#fef08a", "#a16207")
		}
	}
	q := Question{key: "ex/lockpoints", Exhibit: s.done()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. %[2]s is about to clear a jam in the %[3]s. Which energy source has not been locked out?",
		"Refer to %[1]s. Every switch shows OFF. Why is it not yet safe to service the %[3]s?",
	}), ex, sl["FIRST"], machine)
	if strings.HasPrefix(q.Text, "Refer to "+ex+". Every") {
		right := "The " + strings.ToLower(srcs[bad]) + " has no lock or tag"
		var wrongs []string
		for i, s := range srcs {
			if i != bad {
				wrongs = append(wrongs, "The "+strings.ToLower(s)+" has no lock or tag")
			}
		}
		q.Choices, q.Correct = withChoices(r, right, wrongs)
		return q
	}
	q.Choices, q.Correct = oneOf(r, srcs[bad], srcs)
	return q
}
