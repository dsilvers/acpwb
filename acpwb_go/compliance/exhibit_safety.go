package compliance

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Life-safety and active-threat exhibits. As with the bank questions in
// these categories, the answer each exhibit is scored on is real guidance
// (AHA / Red Cross CPR and first aid, DHS Run-Hide-Fight, OSHA GHS
// pictograms); the drawing decides which option is right.

func init() {
	for k, g := range map[string]exhibitGen{
		"cprhands":      cprHandsExhibit,
		"aedpads":       aedPadsExhibit,
		"cprfeedback":   cprFeedbackExhibit,
		"firstaidkit":   firstAidKitExhibit,
		"tourniquet":    tourniquetExhibit,
		"timer":         timerExhibit,
		"hideroom":      hideRoomExhibit,
		"threatescape":  threatEscapeExhibit,
		"alertphone":    alertPhoneExhibit,
		"ghs":           ghsExhibit,
		"bombchecklist": bombChecklistExhibit,
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
		"cpr":        {"cprhands", "aedpads", "cprfeedback", "evac"},
		"firstaid":   {"firstaidkit", "tourniquet", "timer", "evac"},
		"threat":     {"hideroom", "threatescape", "alertphone"},
		"violence":   {"alertphone", "hideroom"},
		"weather":    {"alertphone", "evac"},
		"bombthreat": {"bombchecklist", "alertphone"},
		"hazcom":     {"ghs", "timer"},
		"bloodborne": {"firstaidkit"},
		"elevator":   {"alertphone"},
		"safety":     {"evac", "ghs", "firstaidkit"},
	} {
		categoryExhibits[cat] = ks
	}
}

// labelled option letters drawn in exhibits ("Position A").
var markLetters = []string{"A", "B", "C", "D", "E"}

func markSVG(b *strings.Builder, x, y float64, label string) {
	fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="9" fill="#fff" stroke="#c62828" stroke-width="2"/><text x="%.1f" y="%.1f" text-anchor="middle" %s font-size="10" font-weight="700" fill="#c62828">%s</text>`, x, y, x, y+3.5, svgFont, label)
}

// torso draws a simple front-view adult torso outline at (x, y) scaled by k.
func torso(b *strings.Builder, x, y, k float64) {
	fmt.Fprintf(b, `<g transform="translate(%.1f %.1f) scale(%.2f)">`, x, y, k)
	b.WriteString(`<circle cx="100" cy="28" r="22" fill="#f1d3b8" stroke="#8a6a52" stroke-width="1.5"/>`)
	b.WriteString(`<rect x="90" y="48" width="20" height="12" fill="#f1d3b8" stroke="#8a6a52" stroke-width="1.5"/>`)
	b.WriteString(`<path d="M40 70 Q100 52 160 70 L168 150 Q170 230 150 250 L50 250 Q30 230 32 150 Z" fill="#f1d3b8" stroke="#8a6a52" stroke-width="1.5"/>`)
	b.WriteString(`<path d="M62 72 Q85 80 98 76 M138 72 Q115 80 102 76" fill="none" stroke="#b08a6e" stroke-width="1.5"/>`)                                 // collarbones
	b.WriteString(`<line x1="100" y1="80" x2="100" y2="150" stroke="#c9a487" stroke-width="2" stroke-dasharray="4 3"/>`)                                    // breastbone
	b.WriteString(`<circle cx="74" cy="122" r="2.5" fill="#b07a5e"/><circle cx="126" cy="122" r="2.5" fill="#b07a5e"/>`)                                    // nipples
	b.WriteString(`<path d="M60 150 Q100 170 140 150" fill="none" stroke="#d4b298" stroke-width="1.5"/><circle cx="100" cy="205" r="2.5" fill="#b07a5e"/>`) // ribs, navel
	b.WriteString(`</g>`)
}

// ── CPR hand placement ──────────────────────────────────────────────────

func cprHandsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type spot struct {
		x, y float64
		name string
	}
	spots := []spot{
		{100, 132, "the center of the chest, on the lower half of the breastbone"},
		{128, 104, "the upper left chest, over the heart"},
		{100, 190, "the upper abdomen, below the ribs"},
		{150, 165, "the lower right ribs, at the side"},
		{100, 64, "the base of the neck"},
	}
	order := sample(r, []int{0, 1, 2, 3, 4}, 4)
	if !containsInt(order, 0) {
		order[r.intn(4)] = 0
	}
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="300" height="300" viewBox="0 0 300 300" role="img" aria-label="Torso diagram with marked hand positions"><rect width="300" height="300" fill="#fff" stroke="#c8ccd4"/>`)
	fmt.Fprintf(&b, `<text x="150" y="18" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">ADULT CPR — HAND POSITION</text>`, svgFont)
	torso(&b, 50, 30, 1.0)
	right := ""
	var wrongs []string
	for i, o := range order {
		s := spots[o]
		markSVG(&b, 50+s.x, 30+s.y, markLetters[i])
		if o == 0 {
			right = "Position " + markLetters[i]
		} else {
			wrongs = append(wrongs, "Position "+markLetters[i])
		}
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/cprhands", Exhibit: b.String()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. %[2]s has collapsed and is not breathing normally. Where should you place your hands for chest compressions?",
		"Refer to %[1]s. Which marked position is correct for adult chest compressions?",
		"Refer to %[1]s. You are about to begin hands-only CPR on an adult. Which position should the heel of your hand be on?",
	}), ex, sl["NAME2"])
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// ── AED pad placement ───────────────────────────────────────────────────

func aedPadsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// Pad pairs in torso coordinates (pad centers). Index 0 is the standard
	// adult anterolateral placement: upper right chest below the collarbone,
	// lower left side below the armpit.
	layouts := [][2][2]float64{
		{{70, 100}, {148, 160}},
		{{78, 200}, {122, 200}},  // both on the abdomen
		{{124, 98}, {132, 140}},  // both on the left chest, touching
		{{100, 128}, {104, 138}}, // overlapping in the center
		{{56, 76}, {100, 214}},   // on the shoulder and the navel
	}
	order := sample(r, []int{1, 2, 3, 4}, 3)
	order = append(order, 0)
	shuffle(r, order)
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="190" viewBox="0 0 460 190" role="img" aria-label="Four AED pad placements"><rect width="460" height="190" fill="#fff" stroke="#c8ccd4"/>`)
	right := ""
	var wrongs []string
	for i, o := range order {
		x := 10 + float64(i)*112
		torso(&b, x, 14, 0.5)
		for _, p := range layouts[o] {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="16" height="22" rx="3" fill="#7f8c99" stroke="#3d4a57" transform="rotate(-20 %.1f %.1f)"/>`, x+p[0]*0.5-8, 14+p[1]*0.5-11, x+p[0]*0.5, 14+p[1]*0.5)
		}
		fmt.Fprintf(&b, `<text x="%.1f" y="176" text-anchor="middle" %s font-size="11" font-weight="700" fill="#0a1628">Diagram %d</text>`, x+50, svgFont, i+1)
		if o == 0 {
			right = fmt.Sprintf("Diagram %d", i+1)
		} else {
			wrongs = append(wrongs, fmt.Sprintf("Diagram %d", i+1))
		}
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/aedpads", Exhibit: b.String()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. The AED from the %[2]s has arrived. Which diagram shows correct pad placement for an adult?",
		"Refer to %[1]s. Which diagram follows the pictures printed on standard adult AED pads?",
		"Refer to %[1]s. A colleague is applying AED pads to an adult in the %[2]s. Which placement is correct?",
	}), ex, sl["PLACE"])
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── CPR feedback device ─────────────────────────────────────────────────

func cprFeedbackExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	depthQ := r.chance(0.4)
	var val float64
	var lo, hi, min, max float64
	var unit, title string
	if depthQ {
		lo, hi, min, max, unit, title = 2.0, 2.4, 0.5, 3.5, "in", "COMPRESSION DEPTH"
		val = pick(r, []float64{0.9, 1.2, 1.4, 1.6, 1.8, 2.1, 2.2, 2.3, 2.7, 2.9, 3.1})
	} else {
		lo, hi, min, max, unit, title = 100, 120, 60, 160, "/min", "COMPRESSION RATE"
		val = float64(pick(r, []int{72, 78, 84, 88, 92, 96, 104, 108, 112, 116, 126, 132, 138, 146}))
	}
	ang := func(v float64) float64 { return math.Pi + (v-min)/(max-min)*math.Pi }
	pt := func(v, rad float64) (float64, float64) { return 150 + rad*math.Cos(ang(v)), 172 + rad*math.Sin(ang(v)) }
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="300" height="222" viewBox="0 0 300 222" role="img" aria-label="CPR feedback gauge"><rect width="300" height="222" rx="12" fill="#1e2630"/>`)
	fmt.Fprintf(&b, `<text x="150" y="22" text-anchor="middle" %s font-size="10" font-weight="700" fill="#cfd8e3" letter-spacing="1">CPR FEEDBACK · %s</text>`, svgFont, title)
	b.WriteString(`<path d="M40 172 A110 110 0 0 1 260 172" fill="none" stroke="#3b4652" stroke-width="18"/>`)
	x1, y1 := pt(lo, 110)
	x2, y2 := pt(hi, 110)
	fmt.Fprintf(&b, `<path d="M%.1f %.1f A110 110 0 0 1 %.1f %.1f" fill="none" stroke="#2e9d4f" stroke-width="18"/>`, x1, y1, x2, y2)
	nx, ny := pt(val, 96)
	fmt.Fprintf(&b, `<line x1="150" y1="172" x2="%.1f" y2="%.1f" stroke="#ffcc33" stroke-width="4" stroke-linecap="round"/><circle cx="150" cy="172" r="7" fill="#ffcc33"/>`, nx, ny)
	for _, v := range []float64{min, lo, hi, max} {
		tx, ty := pt(v, 130)
		f := "%.0f"
		if depthQ {
			f = "%.1f"
		}
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" %s font-size="8.5" fill="#cfd8e3">`+f+`</text>`, tx, ty+3, svgFont, v)
	}
	f := "%.0f"
	if depthQ {
		f = "%.1f"
	}
	fmt.Fprintf(&b, `<text x="150" y="208" text-anchor="middle" font-family="Courier New,monospace" font-size="20" font-weight="700" fill="#ffcc33">`+f+` %s</text>`, val, unit)
	b.WriteString(`</svg>`)
	q := Question{key: "ex/cprfeedback", Exhibit: b.String()}
	var right string
	var wrongs []string
	if depthQ {
		q.Text = fmt.Sprintf("Refer to %s, a CPR feedback device on a training manikin during %s's practice session. What should %s do?", ex, sl["NAME"], sl["FIRST"])
		opts := []string{"Push harder; the compressions are too shallow", "Maintain depth; it is within the target range", "Push less deeply; the compressions are too deep"}
		switch {
		case val < lo:
			right = opts[0]
		case val <= hi:
			right = opts[1]
		default:
			right = opts[2]
		}
		for _, o := range opts {
			if o != right {
				wrongs = append(wrongs, o)
			}
		}
		wrongs = append(wrongs, "Stop compressions and check for a pulse")
	} else {
		q.Text = fmt.Sprintf("Refer to %s, a CPR feedback device during %s's practice session. Based on the reading, %s should:", ex, sl["NAME"], sl["FIRST"])
		opts := []string{"Speed up; the rate is below 100 per minute", "Keep the same pace; the rate is within 100 to 120 per minute", "Slow down; the rate is above 120 per minute"}
		switch {
		case val < lo:
			right = opts[0]
		case val <= hi:
			right = opts[1]
		default:
			right = opts[2]
		}
		for _, o := range opts {
			if o != right {
				wrongs = append(wrongs, o)
			}
		}
		wrongs = append(wrongs, "Pause for 10 seconds between every set of compressions")
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── first aid kit ───────────────────────────────────────────────────────

var kitItems = []struct{ Name, Label, Color string }{
	{"Sterile gauze pads", "GAUZE", "#ffffff"},
	{"Adhesive bandages", "BANDAGES", "#f6d6b8"},
	{"Disposable gloves", "GLOVES", "#9fd3f2"},
	{"Tourniquet", "TOURNIQUET", "#2b2b2b"},
	{"Trauma shears", "SHEARS", "#e04848"},
	{"CPR face shield", "CPR SHIELD", "#4caf50"},
	{"Instant cold pack", "COLD PACK", "#5aa0ff"},
	{"Burn dressing", "BURN GEL", "#ffb347"},
	{"Antiseptic wipes", "WIPES", "#ececec"},
	{"Emergency blanket", "BLANKET", "#d4af37"},
	{"Elastic bandage", "WRAP", "#c9b49a"},
	{"Eye wash", "EYE WASH", "#bfe6ff"},
}

func firstAidKitExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	idx := sample(r, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 9)
	required := idx[:7] // listed on the lid
	missing := required[r.intn(7)]
	var present []int
	for _, i := range idx {
		if i != missing {
			present = append(present, i)
		}
	}
	shuffle(r, present)
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="260" viewBox="0 0 460 260" role="img" aria-label="Open first aid kit with contents list"><rect width="460" height="260" fill="#eef1f4"/>`)
	// Lid with contents label.
	b.WriteString(`<rect x="14" y="14" width="160" height="232" rx="10" fill="#c62828"/><rect x="26" y="26" width="136" height="208" rx="4" fill="#fff"/>`)
	b.WriteString(`<rect x="80" y="34" width="28" height="8" fill="#c62828"/><rect x="90" y="24" width="8" height="28" fill="#c62828"/>`)
	fmt.Fprintf(&b, `<text x="94" y="66" text-anchor="middle" %s font-size="9" font-weight="700" fill="#0a1628">REQUIRED CONTENTS</text><text x="94" y="77" text-anchor="middle" %s font-size="7" fill="#555">%s Floor · Kit %s</text>`, svgFont, svgFont, esc(sl["FLOOR"]), esc(sl["FORM"]))
	reqSorted := append([]int(nil), required...)
	sortInts(reqSorted)
	for i, k := range reqSorted {
		fmt.Fprintf(&b, `<rect x="34" y="%d" width="7" height="7" fill="none" stroke="#333"/><text x="46" y="%d" %s font-size="8.5" fill="#222">%s</text>`, 88+i*20, 95+i*20, svgFont, esc(kitItems[k].Name))
	}
	// Tray with items.
	b.WriteString(`<rect x="188" y="14" width="258" height="232" rx="10" fill="#c62828"/><rect x="198" y="24" width="238" height="212" rx="4" fill="#f7f7f7"/>`)
	for i, k := range present {
		col, row := i%3, i/3
		x, y := 206+float64(col)*77, 32+float64(row)*68
		it := kitItems[k]
		ink := "#222"
		if it.Color == "#2b2b2b" || it.Color == "#e04848" || it.Color == "#4caf50" || it.Color == "#5aa0ff" {
			ink = "#fff"
		}
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="70" height="56" rx="6" fill="%s" stroke="#999"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="8" font-weight="700" fill="%s">%s</text>`,
			x, y, it.Color, x+35, y+32, svgFont, ink, esc(it.Label))
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/firstaidkit", Exhibit: b.String()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. During the monthly kit check on the %[2]s floor, which required item is missing?",
		"Refer to %[1]s. Compare the kit's contents to the list on its lid. Which item must be restocked?",
		"Refer to %[1]s. %[3]s used this kit yesterday. Which required item was not replaced?",
	}), ex, sl["FLOOR"], sl["NAME"])
	var wrongs []string
	for _, k := range required {
		if k != missing && len(wrongs) < 3 {
			wrongs = append(wrongs, kitItems[k].Name)
		}
	}
	q.Choices, q.Correct = withChoices(r, kitItems[missing].Name, wrongs)
	return q
}

// ── tourniquet placement ────────────────────────────────────────────────

func tourniquetExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// A leg, hip at left, foot at right; wound on the thigh. Positions:
	// 0 = 2-3 inches above the wound (toward the hip): correct.
	// 1 = directly on the wound, 2 = on the knee joint, 3 = on the calf.
	wx := float64(r.between(150, 175))
	pos := []float64{wx - 42, wx, 262, 330}
	order := []int{0, 1, 2, 3}
	shuffle(r, order)
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="170" viewBox="0 0 460 170" role="img" aria-label="Leg with wound and tourniquet positions"><rect width="460" height="170" fill="#fff" stroke="#c8ccd4"/>`)
	fmt.Fprintf(&b, `<text x="230" y="18" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">BLEEDING CONTROL — TOURNIQUET POSITION</text>`, svgFont)
	b.WriteString(`<path d="M30 60 Q60 52 150 58 L262 66 Q270 64 276 68 L390 72 Q420 70 430 88 L432 110 Q420 112 392 104 L276 102 Q270 106 262 104 L150 112 Q60 118 30 110 Z" fill="#f1d3b8" stroke="#8a6a52" stroke-width="1.5"/>`)
	fmt.Fprintf(&b, `<text x="40" y="140" %s font-size="9" fill="#666">HIP</text><text x="262" y="140" text-anchor="middle" %s font-size="9" fill="#666">KNEE</text><text x="425" y="140" text-anchor="end" %s font-size="9" fill="#666">FOOT</text>`, svgFont, svgFont, svgFont)
	fmt.Fprintf(&b, `<ellipse cx="%.0f" cy="85" rx="12" ry="7" fill="#9b1c1c"/><path d="M%.0f 92 q2 10 -2 18 M%.0f 92 q-1 8 2 14" stroke="#9b1c1c" stroke-width="2" fill="none"/>`, wx, wx-4, wx+5)
	right := ""
	var wrongs []string
	for i, o := range order {
		x := pos[o]
		fmt.Fprintf(&b, `<line x1="%.0f" y1="46" x2="%.0f" y2="124" stroke="#555" stroke-dasharray="3 3"/>`, x, x)
		markSVG(&b, x, 36, markLetters[i])
		if o == 0 {
			right = "Position " + markLetters[i]
		} else {
			wrongs = append(wrongs, "Position "+markLetters[i])
		}
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/tourniquet", Exhibit: b.String()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. Firm pressure has not stopped life-threatening bleeding from %[2]s's thigh. Where should the tourniquet be applied?",
		"Refer to %[1]s. Which marked position is correct for a tourniquet on this leg wound?",
	}), ex, sl["NAME2"])
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── emergency timer ─────────────────────────────────────────────────────

func timerExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type scenario struct {
		label, text string
		limit       int // seconds
		before      string
		after       string
	}
	ss := []scenario{
		{"SEIZURE TIMER", "%s is having a seizure in the %s. You started this timer when it began, and the seizure is still going. What should you do?", 300,
			"Keep timing, cushion their head, keep the area clear, and call 911 if it passes 5 minutes",
			"Call 911 now; a seizure lasting more than 5 minutes is an emergency"},
		{"EYEWASH TIMER", "%s got cleaning solution in their eyes in the %s and is at the eyewash station. The timer shows how long they have been flushing. What should happen now?", 900,
			"Keep flushing until at least 15 minutes have passed, then seek medical help",
			"Flushing has reached 15 minutes; get medical help and bring the product's Safety Data Sheet"},
		{"BURN COOLING", "%s burned their hand on the coffee machine in the %s and is cooling it under running water. The timer shows how long it has been cooled. What should happen now?", 600,
			"Keep cooling the burn under cool running water until at least 10 minutes have passed",
			"It has been cooled for 10 minutes; cover it loosely with a clean dressing and seek medical advice if needed"},
	}
	s := pick(r, ss)
	secs := s.limit + r.between(-4, 4)*30 + r.between(0, 29)
	if secs == s.limit {
		secs += 17
	}
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="260" height="150" viewBox="0 0 260 150" role="img" aria-label="Timer"><rect width="260" height="150" rx="16" fill="#263238"/>`)
	fmt.Fprintf(&b, `<text x="130" y="30" text-anchor="middle" %s font-size="11" font-weight="700" fill="#b0bec5" letter-spacing="1">%s</text>`, svgFont, s.label)
	b.WriteString(`<rect x="24" y="44" width="212" height="70" rx="6" fill="#c5e1a5"/>`)
	fmt.Fprintf(&b, `<text x="130" y="96" text-anchor="middle" font-family="Courier New,monospace" font-size="40" font-weight="700" fill="#1b2a10">%02d:%02d</text>`, secs/60, secs%60)
	fmt.Fprintf(&b, `<text x="130" y="136" text-anchor="middle" %s font-size="9" fill="#90a4ae">MIN : SEC · RUNNING</text>`, svgFont)
	b.WriteString(`</svg>`)
	q := Question{key: "ex/timer", Exhibit: b.String()}
	q.Text = fmt.Sprintf("Refer to %s. "+s.text, ex, sl["NAME2"], sl["PLACE"])
	right, other := s.before, s.after
	if secs > s.limit {
		right, other = other, right
	}
	q.Choices, q.Correct = withChoices(r, right, []string{other, "Stop, since the timer has been running long enough to file a report", "Wait for the floor warden before doing anything"})
	return q
}

// ── hiding spot (active threat) ─────────────────────────────────────────

func hideRoomExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// Top-down room 300×200 at (80,40). Door on the bottom wall with a
	// window; a glass wall on the left or right. The door window's view is
	// drawn as a cone. Spot types: 0 out of sight on the door wall, beside
	// the door's hinge side, away from the glass (correct); 1 beside the
	// glass wall; 2 in the cone, across from the door; 3 in the cone,
	// under the table.
	const L, T, W, H = 80.0, 40.0, 300.0, 200.0
	glassLeft := r.chance(0.5)
	doorX := L + 150 + float64(r.between(-30, 30))
	gx := L + W - 6
	if glassLeft {
		gx = L
	}
	spots := map[int][2]float64{}
	// Out of sight: the corner on the door wall, on the side away from the glass.
	if glassLeft {
		spots[0] = [2]float64{L + W - 22, T + H - 22}
		spots[1] = [2]float64{L + 22, T + 40 + float64(r.between(0, 60))}
	} else {
		spots[0] = [2]float64{L + 22, T + H - 22}
		spots[1] = [2]float64{L + W - 22, T + 40 + float64(r.between(0, 60))}
	}
	spots[2] = [2]float64{doorX + float64(r.between(-20, 20)), T + 26}
	spots[3] = [2]float64{doorX + float64(r.between(-10, 10)), T + 100}
	order := []int{0, 1, 2, 3}
	shuffle(r, order)
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="290" viewBox="0 0 460 290" role="img" aria-label="Conference room plan with door window view"><rect width="460" height="290" fill="#fff"/>`)
	fmt.Fprintf(&b, `<text x="230" y="22" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">%s CONFERENCE ROOM — PLAN VIEW</text>`, svgFont, esc(strings.ToUpper(pick(r, roomNames))))
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#f7f4ec" stroke="#333" stroke-width="5"/>`, L, T, W, H)
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="6" height="%.0f" fill="#bfe0f2" stroke="#7fb3d5"/>`, gx, T, H)
	// View cone from the door window up into the room.
	fmt.Fprintf(&b, `<path d="M%.0f %.0f L%.0f %.0f L%.0f %.0f Z" fill="#ffcc33" opacity=".22"/>`, doorX, T+H, doorX-60, T, doorX+60, T)
	// Door.
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="44" height="7" fill="#8d6e63"/><rect x="%.0f" y="%.0f" width="10" height="7" fill="#bfe0f2"/>`, doorX-22, T+H-3, doorX-5, T+H-3)
	// Table and chairs.
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="120" height="44" rx="6" fill="#c8a27a" stroke="#8d6e63"/>`, doorX-60, T+78)
	fmt.Fprintf(&b, `<g %s font-size="8" fill="#555"><text x="%.0f" y="%.0f" text-anchor="middle">DOOR (window, no lock)</text><text x="%.0f" y="%.0f" text-anchor="middle" transform="rotate(-90 %.0f %.0f)">GLASS WALL — CORRIDOR</text></g>`,
		svgFont, doorX, T+H+18, gx+map[bool]float64{true: -10, false: 18}[glassLeft], T+H/2, gx+map[bool]float64{true: -10, false: 18}[glassLeft], T+H/2)
	fmt.Fprintf(&b, `<g %s font-size="8" fill="#333"><rect x="90" y="266" width="12" height="9" fill="#ffcc33" opacity=".5"/><text x="106" y="274">Visible through the door window</text><rect x="268" y="266" width="12" height="9" fill="#bfe0f2" stroke="#7fb3d5"/><text x="284" y="274">Glass wall</text></g>`, svgFont)
	right := ""
	var wrongs []string
	for i, o := range order {
		p := spots[o]
		markSVG(&b, p[0], p[1], markLetters[i])
		if o == 0 {
			right = "Spot " + markLetters[i]
		} else {
			wrongs = append(wrongs, "Spot "+markLetters[i])
		}
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/hideroom", Exhibit: b.String()}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. An active shooter is reported in the building and you cannot safely escape. You have barricaded the door. Where in the room should you hide?",
		"Refer to %[1]s. During a lockdown you are in this room with %[2]s. Which spot keeps you out of sight of the door window and the glass wall?",
		"Refer to %[1]s. You must hide in this room during an active threat. Which marked spot is the best choice?",
	}), ex, sl["NAME2"])
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── escape route away from a threat ─────────────────────────────────────

func threatEscapeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	for try := 0; ; try++ {
		p := buildFloorPlan(r, sl)
		p.blocked = -1
		// Threat somewhere in the open floor; the safest exit is the one
		// whose straight path from "You are here" stays farthest from it.
		tx, ty := float64(r.between(60, 380)), float64(r.between(70, 255))
		if math.Hypot(tx-p.youX, ty-p.youY) < 70 {
			continue
		}
		clear := make([]float64, len(p.exits))
		best := 0
		for i, e := range p.exits {
			clear[i] = segPointDist(p.youX, p.youY, e.x, e.y, tx, ty)
			if clear[i] > clear[best] {
				best = i
			}
		}
		gap := math.Inf(1)
		for i := range clear {
			if i != best && clear[best]-clear[i] < gap {
				gap = clear[best] - clear[i]
			}
		}
		if gap < 30 && try < 40 {
			continue
		}
		svg := floorPlanSVG(p)
		threat := fmt.Sprintf(`<g><circle cx="%.0f" cy="%.0f" r="13" fill="#c62828" opacity=".25"/><circle cx="%.0f" cy="%.0f" r="6" fill="#c62828"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="7.5" font-weight="700" fill="#c62828">REPORTED THREAT</text></g></svg>`, tx, ty, tx, ty, tx, ty-16, svgFont)
		svg = strings.TrimSuffix(svg, "</svg>") + threat
		q := Question{key: "ex/threatescape", Exhibit: svg}
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %[1]s. An active shooter is reported at the marked location on the %[2]s floor. You are at \"You are here\" and can leave safely. Which exit should you use?",
			"Refer to %[1]s. A threat is reported where shown. Following Run, Hide, Fight, which exit gives you the safest escape route?",
		}), ex, p.floor)
		var wrongs []string
		for i, e := range p.exits {
			if i != best {
				wrongs = append(wrongs, "Exit "+e.label)
			}
		}
		q.Choices, q.Correct = withChoices(r, "Exit "+p.exits[best].label, wrongs)
		return q
	}
}

func segPointDist(x1, y1, x2, y2, px, py float64) float64 {
	dx, dy := x2-x1, y2-y1
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(x1+t*dx-px, y1+t*dy-py)
}

// ── emergency notification on a phone ───────────────────────────────────

func alertPhoneExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type alert struct {
		kind, text string
	}
	floor := r.between(12, 40)
	lo := floor - r.between(0, 4)
	hi := lo + r.between(2, 6)
	if r.chance(0.35) { // your floor outside the range
		lo, hi = floor+r.between(2, 5), floor+r.between(6, 10)
	}
	in := floor >= lo && floor <= hi
	alerts := []alert{
		{"lockdown", "LOCKDOWN in effect for the entire building. Active threat reported. Lock or barricade doors, silence phones, stay out of sight. Await instructions."},
		{"evac", fmt.Sprintf("FIRE ALARM: Floors %d–%d EVACUATE now by the stairs. All other floors remain in place and await instructions.", lo, hi)},
		{"shelter", "TORNADO WARNING for Milwaukee County. All associates move now to interior rooms or stairwells on the lowest floor possible, away from windows."},
		{"allclear", "ALL CLEAR. Police have secured the building. The lockdown has ended. Follow instructions from Security and law enforcement."},
		{"test", "TEST — This is a test of the ACPWB Emergency Notification System. No action is required."},
	}
	latest := pick(r, alerts)
	var thread []alert
	if latest.kind == "allclear" {
		thread = append(thread, alerts[0])
	} else if r.chance(0.4) {
		thread = append(thread, alerts[4])
	}
	thread = append(thread, latest)
	var b strings.Builder
	fmt.Fprintf(&b, `<text x="125" y="58" text-anchor="middle" %s font-size="9" fill="#666">ACPWB Alerts · Today</text>`, svgFont)
	y := 70.0
	t0 := r.between(9*60, 16*60)
	for i, a := range thread {
		lines := wrapText(a.text, 34)
		h := 26 + float64(len(lines))*12
		fill := "#fff"
		if a.kind == "lockdown" || a.kind == "evac" || a.kind == "shelter" {
			fill = "#ffe5e5"
		}
		fmt.Fprintf(&b, `<rect x="24" y="%.0f" width="202" height="%.0f" rx="10" fill="%s" stroke="#d0d0d8"/>`, y, h, fill)
		fmt.Fprintf(&b, `<text x="34" y="%.0f" %s font-size="8" font-weight="700" fill="#c62828">EMERGENCY ALERT</text><text x="216" y="%.0f" text-anchor="end" %s font-size="8" fill="#888">%s</text>`, y+14, svgFont, y+14, svgFont, clock(t0+i*r.between(8, 40)))
		for j, ln := range lines {
			fmt.Fprintf(&b, `<text x="34" y="%.0f" %s font-size="9" fill="#111">%s</text>`, y+28+float64(j)*12, svgFont, esc(ln))
		}
		y += h + 10
	}
	ph := math.Max(260, y+30)
	b.WriteString(`</svg>`)
	b2 := fmt.Sprintf(`<svg class="cmp-exhibit-img" width="250" height="%.0f" viewBox="0 0 250 %.0f" role="img" aria-label="Emergency alert on a phone"><rect x="5" y="5" width="240" height="%.0f" rx="32" fill="#111"/><rect x="15" y="40" width="220" height="%.0f" rx="6" fill="#f2f2f7"/><rect x="95" y="16" width="60" height="12" rx="6" fill="#222"/>`, ph+20, ph+20, ph+10, ph-30) + b.String()
	acts := map[string]string{
		"lockdown": "Lock or barricade the door, silence your phone, and stay out of sight until told otherwise",
		"leave":    "Leave now by the nearest stairwell, without using the elevators",
		"stay":     "Stay on your floor, away from the stairwells, and await further instructions",
		"shelter":  "Move to an interior room or stairwell on the lowest floor possible, away from windows",
		"allclear": "Follow instructions from Security and law enforcement as normal activity resumes",
		"none":     "No action is required",
	}
	var key string
	switch latest.kind {
	case "lockdown":
		key = "lockdown"
	case "evac":
		key = "stay"
		if in {
			key = "leave"
		}
	case "shelter":
		key = "shelter"
	case "allclear":
		key = "allclear"
	default:
		key = "none"
	}
	var wrongs []string
	for _, k := range sample(r, []string{"lockdown", "leave", "stay", "shelter", "allclear", "none"}, 6) {
		if k != key && len(wrongs) < 3 {
			wrongs = append(wrongs, acts[k])
		}
	}
	q := Question{key: "ex/alertphone", Exhibit: b2}
	q.Text = fmt.Sprintf(pick(r, []string{
		"Refer to %[1]s. You are at your desk on the %[2]s floor when these alerts arrive. Based on the most recent alert, what should you do?",
		"Refer to %[1]s. %[3]s, who sits on the %[2]s floor, shows you these alerts on their phone. What should you both do now?",
	}), ex, ordinal(floor), sl["NAME2"], ordinal(floor))
	q.Choices, q.Correct = withChoices(r, acts[key], wrongs)
	return q
}

func wrapText(s string, n int) []string {
	var lines []string
	cur := ""
	for _, w := range strings.Fields(s) {
		if cur != "" && len(cur)+1+len(w) > n {
			lines = append(lines, cur)
			cur = w
		} else if cur == "" {
			cur = w
		} else {
			cur += " " + w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// ── GHS hazard pictograms ───────────────────────────────────────────────

var ghsSymbols = []struct{ Name, Meaning, Draw string }{
	{"flame", "Flammable", `<path d="M50 74 C36 72 32 58 40 46 C42 54 46 54 46 50 C44 40 50 30 56 26 C54 36 62 40 64 50 C66 44 64 40 64 38 C72 46 70 70 50 74 Z" fill="#111"/><rect x="30" y="74" width="40" height="4" fill="#111"/>`},
	{"exclamation", "Irritant or harmful (for example, skin or eye irritation)", `<rect x="45" y="26" width="10" height="34" rx="4" fill="#111"/><circle cx="50" cy="69" r="6" fill="#111"/>`},
	{"skull", "Acutely toxic (can be fatal or toxic in small amounts)", `<circle cx="50" cy="44" r="15" fill="#111"/><rect x="42" y="54" width="16" height="10" fill="#111"/><circle cx="44" cy="43" r="4" fill="#fff"/><circle cx="56" cy="43" r="4" fill="#fff"/><path d="M30 64 L70 78 M70 64 L30 78" stroke="#111" stroke-width="5"/>`},
	{"cylinder", "Gas under pressure", `<rect x="30" y="40" width="40" height="20" rx="10" fill="#111" transform="rotate(-25 50 50)"/><rect x="66" y="34" width="8" height="8" fill="#111" transform="rotate(-25 50 50)"/>`},
	{"health", "A serious health hazard (for example, a carcinogen or respiratory sensitizer)", `<circle cx="50" cy="30" r="8" fill="#111"/><path d="M36 76 L38 44 Q50 38 62 44 L64 76 Z" fill="#111"/><path d="M50 50 l3 7 l7 1 l-5 5 l2 7 l-7 -4 l-7 4 l2 -7 l-5 -5 l7 -1 Z" fill="#fff"/>`},
	{"corrosion", "Corrosive (can burn skin, damage eyes, or corrode metal)", `<rect x="30" y="24" width="9" height="22" rx="2" fill="#111" transform="rotate(35 34 35)"/><rect x="60" y="24" width="9" height="22" rx="2" fill="#111" transform="rotate(-35 65 35)"/><circle cx="36" cy="52" r="2.2" fill="#111"/><circle cx="34" cy="58" r="2" fill="#111"/><circle cx="64" cy="52" r="2.2" fill="#111"/><circle cx="66" cy="58" r="2" fill="#111"/><path d="M24 64 h18 v8 h-18 Z M34 64 q2 4 6 0" fill="#111"/><path d="M56 72 v-6 q0 -3 3 -3 v-3 q0 -2 2 -2 q2 0 2 2 v2 q0 -2 2 -2 q2 0 2 2 v2 q0 -2 2 -2 q2 0 2 2 v8 q0 3 -3 4 h-10 Z" fill="#111"/><path d="M62 62 q2 3 5 0" stroke="#fff" stroke-width="1.5" fill="none"/>`},
}

func ghsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	idx := sample(r, []int{0, 1, 2, 3, 4, 5}, 4)
	target := r.intn(4)
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="440" height="150" viewBox="0 0 440 150" role="img" aria-label="Four GHS hazard pictograms"><rect width="440" height="150" fill="#fff" stroke="#c8ccd4"/>`)
	for i, k := range idx {
		x := 10 + float64(i)*108
		fmt.Fprintf(&b, `<g transform="translate(%.0f 8)"><path d="M50 4 L96 50 L50 96 L4 50 Z" fill="#fff" stroke="#d50000" stroke-width="7" stroke-linejoin="round"/>%s</g>`, x, ghsSymbols[k].Draw)
		fmt.Fprintf(&b, `<text x="%.0f" y="132" text-anchor="middle" %s font-size="11" font-weight="700" fill="#0a1628">Pictogram %d</text>`, x+50, svgFont, i+1)
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/ghs", Exhibit: b.String()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %[1]s. A product under the %[2]s sink carries a label meaning \"%[3]s.\" Which pictogram would appear on it?",
			"Refer to %[1]s. Which pictogram indicates a product that is: %[3]s?",
		}), ex, sl["PLACE"], strings.ToLower(ghsSymbols[idx[target]].Meaning))
		var wrongs []string
		for i := range idx {
			if i != target {
				wrongs = append(wrongs, fmt.Sprintf("Pictogram %d", i+1))
			}
		}
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Pictogram %d", target+1), wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What does Pictogram %d mean on a chemical label?", ex, target+1)
	var wrongs []string
	for i, k := range idx {
		if i != target {
			wrongs = append(wrongs, ghsSymbols[k].Meaning)
		}
	}
	q.Choices, q.Correct = withChoices(r, ghsSymbols[idx[target]].Meaning, wrongs)
	return q
}

// ── bomb threat checklist ───────────────────────────────────────────────

func bombChecklistExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	start := 60*r.between(8, 16) + r.between(0, 59)
	dur := r.between(40, 420)
	end := start*60 + dur
	where := pick(r, []string{"the " + sl["PLACE"], "the parking structure, level P" + sl["N"], "a backpack on the " + sl["FLOOR"] + " floor", "the mailroom", "the loading dock"})
	when := pick(r, []string{"at 3:00 p.m.", "in one hour", "before the end of the day", "\"soon\""})
	fields := [][]string{
		{"Time call received", clock(start)},
		{"Time call ended", fmt.Sprintf("%s (%d sec later)", clock(end/60), dur)},
		{"Exact words of the threat", fmt.Sprintf("\"There's a device in %s. It goes off %s.\"", where, strings.Trim(when, "\""))},
		{"Where is it?", where},
		{"When will it go off?", strings.Trim(when, "\"")},
		{"Caller's voice", pick(r, []string{"Calm, male, slow", "Angry, female, fast", "Whispered, accent unclear", "Calm, disguised, slow"})},
		{"Background sounds", pick(r, []string{"Traffic, a horn", "Office chatter", "Machinery", "Music, then silence", "Static"})},
		{"Number called from", pick(r, []string{"Blocked", "Unknown", "(414) 555-0" + fmt.Sprint(r.between(100, 199))})},
		{"Reported to 911 at", clock(end/60 + r.between(1, 3))},
	}
	blank := r.between(5, 7)
	fields[blank][1] = ""
	var rows [][]string
	for _, f := range fields {
		v := f[1]
		if v == "" {
			v = "—"
		}
		rows = append(rows, []string{f[0], v})
	}
	q := Question{key: "ex/bombchecklist", Exhibit: docExhibit("Bomb Threat Call Checklist", "Completed by "+sl["NAME"]+" · Ext. "+fmt.Sprint(r.between(4000, 4999)), []string{"Item", "Recorded"}, rows,
		"Keep the caller talking. Do not hang up. Call 911 from another phone.")}
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. According to the checklist, where did the caller say the device is?", ex)
		alts := []string{"the " + sl["PLACE"], "the parking structure, level P" + sl["N"], "a backpack on the " + sl["FLOOR"] + " floor", "the mailroom", "the loading dock"}
		var wrongs []string
		for _, a := range sample(r, alts, len(alts)) {
			if a != where && !contains(wrongs, capFirst(a)) && len(wrongs) < 3 {
				wrongs = append(wrongs, capFirst(a))
			}
		}
		for len(wrongs) < 3 {
			wrongs = append(wrongs, capFirst(pick(r, []string{"the roof", "the executive wing", "the lobby", "the fitness center", "stairwell C"})))
			if wrongs[len(wrongs)-1] == capFirst(where) || contains(wrongs[:len(wrongs)-1], wrongs[len(wrongs)-1]) {
				wrongs = wrongs[:len(wrongs)-1]
			}
		}
		q.Choices, q.Correct = withChoices(r, capFirst(where), wrongs)
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. How long did the call last?", ex)
		f := func(s int) string { return fmt.Sprintf("%d min %02d sec", s/60, s%60) }
		q.Choices, q.Correct = numChoices(r, dur, []int{-60, 60, -30, 30, 120}, f)
	default:
		q.Text = fmt.Sprintf("Refer to %s. Which item on the checklist was left blank and should be completed from memory as soon as possible?", ex)
		var wrongs []string
		for i, f := range fields {
			if i != blank && i >= 2 && len(wrongs) < 3 {
				wrongs = append(wrongs, f[0])
			}
		}
		q.Choices, q.Correct = withChoices(r, fields[blank][0], sample(r, wrongs, 3))
	}
	return q
}

func sortStrings(s []string) { sort.Strings(s) }
func sortInts(s []int)       { sort.Ints(s) }
