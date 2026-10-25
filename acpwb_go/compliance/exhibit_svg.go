package compliance

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Drawings for exhibits.go. Everything the answer key depends on (exit
// positions, items on a desk, a label's date) is drawn from the same values
// the question is scored with.

const svgFont = `font-family="Arial,Helvetica,sans-serif"`

// ── refrigerator label + policy sign ────────────────────────────────────

func fridgeSVG(name, dept string, stored time.Time, days int, tilt float64) string {
	var s strings.Builder
	s.WriteString(`<svg class="cmp-exhibit-img" width="440" height="200" viewBox="0 0 440 200" role="img" aria-label="Food label and refrigerator policy sign">`)
	s.WriteString(`<rect width="440" height="200" fill="#eef2f5"/><rect x="8" y="8" width="424" height="184" rx="10" fill="#f8fafb" stroke="#cfd6dd"/>`)
	// Container lid with masking-tape label.
	s.WriteString(`<rect x="22" y="38" width="206" height="128" rx="16" fill="#cfe3ef" stroke="#9fb9c9"/>`)
	fmt.Fprintf(&s, `<g transform="rotate(%.1f 125 100)"><rect x="34" y="58" width="182" height="86" fill="#f1dfae" stroke="#d8c38c"/>`, tilt)
	hand := `font-family="'Marker Felt','Comic Sans MS','Segoe Print',cursive" fill="#1d2a6b"`
	fmt.Fprintf(&s, `<text x="125" y="76" text-anchor="middle" %s font-size="10">PROPERTY OF</text>`, hand)
	fmt.Fprintf(&s, `<text x="125" y="98" text-anchor="middle" %s font-size="17">%s</text>`, hand, esc(truncate(name, 22)))
	fmt.Fprintf(&s, `<text x="125" y="114" text-anchor="middle" %s font-size="9">%s</text>`, hand, esc(truncate(dept, 32)))
	fmt.Fprintf(&s, `<text x="125" y="134" text-anchor="middle" %s font-size="12">Stored: %s</text></g>`, hand, stored.Format("Mon 1/2"))
	// Policy sign.
	s.WriteString(`<rect x="250" y="24" width="166" height="152" fill="#fff" stroke="#b9c2cb"/><rect x="250" y="24" width="166" height="26" fill="#8b1e1e"/>`)
	fmt.Fprintf(&s, `<text x="333" y="41" text-anchor="middle" %s font-size="10" font-weight="700" fill="#fff" letter-spacing=".5">SHARED REFRIGERATOR</text>`, svgFont)
	for i, l := range []string{"Labelled items may stay", fmt.Sprintf("%d days after the date", days), "stored. After that they", "are discarded, with", "their containers.", "", "Unlabelled items are", "discarded daily."} {
		weight := ""
		if i == 1 {
			weight = ` font-weight="700"`
		}
		fmt.Fprintf(&s, `<text x="262" y="%d" %s font-size="10" fill="#222"%s>%s</text>`, 68+i*13, svgFont, weight, esc(l))
	}
	s.WriteString(`</svg>`)
	return s.String()
}

// ── evacuation floor plan ────────────────────────────────────────────────

type planExit struct {
	label string
	x, y  float64
	horiz bool // on a top/bottom wall
}

type floorPlan struct {
	floor         string
	rooms         [][4]float64
	roomNames     []string
	exits         []planExit
	blocked       int // index into exits, or -1
	youX, youY    float64
	answer        int
	extinguishers [][2]float64
	aeds          [][2]float64 // labelled AED 1, AED 2, …
	nearestAED    int          // -1 when no AED is clearly nearest
}

const planL, planT, planR, planB = 20.0, 44.0, 420.0, 280.0

func buildFloorPlan(r *rnd, sl slots) floorPlan {
	p := floorPlan{floor: sl["FLOOR"], blocked: -1}
	p.rooms = [][4]float64{{planL, planT, 104, 74}, {planR - 94, planT, 94, 74}, {planL, planB - 74, 104, 74}, {planR - 94, planB - 74, 94, 74}, {180, 134, 80, 56}}
	p.roomNames = append(sample(r, roomNames, 4), "Elevators")
	labels := sample(r, []string{"A", "B", "C", "D"}, 4)
	p.exits = []planExit{
		{labels[0], float64(r.between(140, 300)), planT, true},
		{labels[1], float64(r.between(140, 300)), planB, true},
		{labels[2], planL, float64(r.between(130, 200)), false},
		{labels[3], planR, float64(r.between(130, 200)), false},
	}
	block := r.chance(0.45)
	nearestFree := func() (int, float64) {
		best, gap := -1, 0.0
		var ds []float64
		for i, e := range p.exits {
			d := math.Hypot(e.x-p.youX, e.y-p.youY)
			if i == p.blocked {
				d = math.Inf(1)
			}
			ds = append(ds, d)
			if best < 0 || d < ds[best] {
				best = i
			}
		}
		gap = math.Inf(1)
		for i, d := range ds {
			if i != best && d-ds[best] < gap {
				gap = d - ds[best]
			}
		}
		return best, gap
	}
	for try := 0; ; try++ {
		p.youX, p.youY = float64(r.between(50, 390)), float64(r.between(70, 255))
		if p.youX > 165 && p.youX < 275 && p.youY > 120 && p.youY < 205 {
			continue // inside the elevator core
		}
		p.blocked = -1
		if block {
			n, _ := nearestFree()
			p.blocked = n
			if r.chance(0.3) {
				p.blocked = (n + 1 + r.intn(3)) % 4
			}
		}
		a, gap := nearestFree()
		if gap >= 35 || try > 40 {
			p.answer = a
			break
		}
	}
	for i := r.between(2, 6); i > 0; i-- {
		p.extinguishers = append(p.extinguishers, [2]float64{float64(r.between(40, 400)), float64(r.between(62, 262))})
	}
	p.nearestAED = -1
	for i := r.between(2, 3); i > 0; i-- {
		p.aeds = append(p.aeds, [2]float64{float64(r.between(45, 395)), float64(r.between(64, 262))})
	}
	best, second := -1, math.Inf(1)
	for i, a := range p.aeds {
		d := math.Hypot(a[0]-p.youX, a[1]-p.youY)
		if best < 0 || d < math.Hypot(p.aeds[best][0]-p.youX, p.aeds[best][1]-p.youY) {
			if best >= 0 {
				second = math.Hypot(p.aeds[best][0]-p.youX, p.aeds[best][1]-p.youY)
			}
			best = i
		} else if d < second {
			second = d
		}
	}
	if best >= 0 && second-math.Hypot(p.aeds[best][0]-p.youX, p.aeds[best][1]-p.youY) >= 35 {
		p.nearestAED = best
	}
	return p
}

func floorPlanSVG(p floorPlan) string {
	var s strings.Builder
	s.WriteString(`<svg class="cmp-exhibit-img" width="440" height="326" viewBox="0 0 440 326" role="img" aria-label="Evacuation floor plan">`)
	s.WriteString(`<rect width="440" height="326" fill="#fff"/>`)
	fmt.Fprintf(&s, `<text x="220" y="20" text-anchor="middle" %s font-size="11" font-weight="700" fill="#0a1628">EVACUATION PLAN — %s FLOOR</text>`, svgFont, esc(strings.ToUpper(p.floor)))
	fmt.Fprintf(&s, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#f7f4ec" stroke="#333" stroke-width="3"/>`, planL, planT, planR-planL, planB-planT)
	for i, rm := range p.rooms {
		fill := "#e9e4d6"
		if i == 4 {
			fill = "#d5d9e0"
		}
		fmt.Fprintf(&s, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="%s" stroke="#777"/>`, rm[0], rm[1], rm[2], rm[3], fill)
		fmt.Fprintf(&s, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="8" fill="#555">%s</text>`, rm[0]+rm[2]/2, rm[1]+rm[3]/2+3, svgFont, esc(p.roomNames[i]))
	}
	for _, e := range p.extinguishers {
		fmt.Fprintf(&s, `<circle cx="%.0f" cy="%.0f" r="5.5" fill="#c62828"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="7" font-weight="700" fill="#fff">F</text>`, e[0], e[1], e[0], e[1]+2.5, svgFont)
	}
	for i, a := range p.aeds {
		fmt.Fprintf(&s, `<rect x="%.0f" y="%.0f" width="22" height="13" rx="2" fill="#1f7a3a"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="6.5" font-weight="700" fill="#fff">AED</text><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="7" font-weight="700" fill="#1f7a3a">%d</text>`,
			a[0]-11, a[1]-6.5, a[0], a[1]+2.5, svgFont, a[0], a[1]+15, svgFont, i+1)
	}
	for i, e := range p.exits {
		w, h := 34.0, 12.0
		if !e.horiz {
			w, h = 12, 34
		}
		fmt.Fprintf(&s, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#1f7a3a"/>`, e.x-w/2, e.y-h/2, w, h)
		lx, ly := e.x, e.y+3
		switch {
		case e.horiz && e.y == planT:
			ly = e.y - 9
		case e.horiz:
			ly = e.y + 17
		case e.x == planL:
			lx, ly = e.x+28, e.y+3
		default:
			lx, ly = e.x-28, e.y+3
		}
		fmt.Fprintf(&s, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="9" font-weight="700" fill="#1f7a3a">EXIT %s</text>`, lx, ly, svgFont, e.label)
		if i == p.blocked {
			fmt.Fprintf(&s, `<path d="M%.0f %.0f l16 16 M%.0f %.0f l-16 16" stroke="#c62828" stroke-width="4"/>`, e.x-8, e.y-8, e.x+8, e.y-8)
			ty := e.y + 26
			if e.horiz && e.y == planB {
				ty = e.y - 14
			}
			fmt.Fprintf(&s, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="7.5" font-weight="700" fill="#c62828">BLOCKED — CONSTRUCTION</text>`, math.Min(math.Max(e.x, 75), 365), ty, svgFont)
		}
	}
	// You are here.
	fmt.Fprintf(&s, `<path d="%s" fill="#d4a017" stroke="#7a5a00"/>`, starPath(p.youX, p.youY, 9, 4))
	fmt.Fprintf(&s, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="8" font-weight="700" fill="#7a5a00">YOU ARE HERE</text>`, p.youX, p.youY-12, svgFont)
	fmt.Fprintf(&s, `<g %s font-size="8" fill="#333"><path d="%s" fill="#d4a017"/><text x="40" y="318">You are here</text><rect x="110" y="311" width="14" height="8" fill="#1f7a3a"/><text x="130" y="318">Exit</text><circle cx="172" cy="315" r="4.5" fill="#c62828"/><text x="181" y="318">Fire extinguisher</text><path d="M262 310 l9 9 M271 310 l-9 9" stroke="#c62828" stroke-width="2.5"/><text x="278" y="318">Exit blocked</text><rect x="338" y="309" width="16" height="10" rx="2" fill="#1f7a3a"/><text x="358" y="318">AED</text></g>`, svgFont, starPath(30, 315, 6, 2.6))
	s.WriteString(`</svg>`)
	return s.String()
}

func starPath(cx, cy, R, r float64) string {
	var b strings.Builder
	for i := 0; i < 10; i++ {
		rad := R
		if i%2 == 1 {
			rad = r
		}
		a := -math.Pi/2 + float64(i)*math.Pi/5
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&b, "%s%.1f %.1f", cmd, cx+rad*math.Cos(a), cy+rad*math.Sin(a))
	}
	b.WriteString("Z")
	return b.String()
}

// ── clean desk scene ─────────────────────────────────────────────────────

var deskItemNames = map[string]string{
	"screen": "The unlocked computer screen", "sticky": "The sticky note on the monitor", "papers": "The papers marked Confidential",
	"usb": "The USB drive", "badge": "The access badge", "keys": "The filing cabinet keys",
	"mug": "The coffee mug", "plant": "The plant", "photo": "The framed photo", "pens": "The pen cup", "stapler": "The stapler", "headphones": "The headphones",
}

type deskScene struct {
	unlocked, sticky bool
	slotItems        []string // placed in deskSlots order
	benign           []string
	slotViol         []string
	year, doc        string
}

func (d deskScene) violations() []string {
	var v []string
	if d.unlocked {
		v = append(v, "screen")
	}
	if d.sticky {
		v = append(v, "sticky")
	}
	return append(v, d.slotViol...)
}

var deskSlots = [][2]float64{{48, 210}, {118, 252}, {44, 262}, {322, 250}, {394, 210}, {392, 262}}

func buildDesk(r *rnd, sl slots) deskScene {
	d := deskScene{year: sl["YEAR"], doc: sl["DOCUMENT"]}
	chosen := sample(r, []string{"screen", "sticky", "papers", "usb", "badge", "keys"}, r.between(0, 4))
	for _, c := range chosen {
		switch c {
		case "screen":
			d.unlocked = true
		case "sticky":
			d.sticky = true
		default:
			d.slotViol = append(d.slotViol, c)
		}
	}
	free := len(deskSlots) - len(d.slotViol)
	lo := 3
	if free < lo {
		lo = free
	}
	d.benign = sample(r, []string{"mug", "plant", "photo", "pens", "stapler", "headphones"}, r.between(lo, free))
	d.slotItems = append(append([]string{}, d.slotViol...), d.benign...)
	shuffle(r, d.slotItems)
	return d
}

func deskSVG(d deskScene) string {
	var s strings.Builder
	s.WriteString(`<svg class="cmp-exhibit-img" width="440" height="290" viewBox="0 0 440 290" role="img" aria-label="Photo of a workstation">`)
	s.WriteString(`<rect width="440" height="290" fill="#e7e1d3"/><rect y="186" width="440" height="104" fill="#b48c5c"/><rect y="186" width="440" height="5" fill="#8f6b40"/>`)
	// Monitor.
	s.WriteString(`<rect x="208" y="150" width="24" height="30" fill="#333"/><rect x="182" y="176" width="76" height="9" rx="3" fill="#2a2a2a"/>`)
	s.WriteString(`<rect x="146" y="36" width="148" height="116" rx="5" fill="#1d1d1f"/>`)
	if d.unlocked {
		s.WriteString(`<rect x="154" y="44" width="132" height="100" fill="#fff"/><rect x="154" y="44" width="132" height="11" fill="#1f7a3a"/>`)
		fmt.Fprintf(&s, `<text x="158" y="52.5" %s font-size="6.5" fill="#fff">%s.xlsx</text>`, svgFont, esc(capFirst(truncate(d.doc, 26))))
		for i := 0; i < 8; i++ {
			fmt.Fprintf(&s, `<line x1="154" y1="%d" x2="286" y2="%d" stroke="#d9dde3"/>`, 66+i*10, 66+i*10)
		}
		for _, x := range []int{186, 220, 254} {
			fmt.Fprintf(&s, `<line x1="%d" y1="56" x2="%d" y2="144" stroke="#d9dde3"/>`, x, x)
		}
		s.WriteString(`<rect x="187" y="77" width="33" height="9" fill="#fff3b0"/><rect x="221" y="97" width="33" height="9" fill="#ffd6d1"/>`)
	} else {
		s.WriteString(`<rect x="154" y="44" width="132" height="100" fill="#12305a"/>`)
		s.WriteString(`<rect x="208" y="84" width="24" height="20" rx="3" fill="#dfe6f0"/><path d="M212 84 v-6 a8 8 0 0 1 16 0 v6" fill="none" stroke="#dfe6f0" stroke-width="3"/>`)
		fmt.Fprintf(&s, `<text x="220" y="122" text-anchor="middle" %s font-size="7.5" fill="#dfe6f0">Locked · Ctrl+Alt+Del</text>`, svgFont)
	}
	if d.sticky {
		s.WriteString(`<g transform="rotate(7 300 60)"><rect x="276" y="40" width="40" height="38" fill="#ffe96b" stroke="#e0c84a"/>`)
		fmt.Fprintf(&s, `<text x="296" y="55" text-anchor="middle" font-family="'Marker Felt','Comic Sans MS',cursive" font-size="8" fill="#1d2a6b">pw:</text><text x="296" y="68" text-anchor="middle" font-family="'Marker Felt','Comic Sans MS',cursive" font-size="7" fill="#1d2a6b">Summer%s!</text></g>`, esc(d.year))
	}
	// Keyboard and mouse.
	s.WriteString(`<rect x="166" y="200" width="108" height="24" rx="3" fill="#d9d9d9" stroke="#aaa"/>`)
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&s, `<line x1="171" y1="%d" x2="269" y2="%d" stroke="#bbb" stroke-dasharray="5 2"/>`, 206+i*6, 206+i*6)
	}
	s.WriteString(`<ellipse cx="292" cy="214" rx="8" ry="11" fill="#d9d9d9" stroke="#aaa"/>`)
	for i, it := range d.slotItems {
		if i >= len(deskSlots) {
			break
		}
		drawDeskItem(&s, it, deskSlots[i][0], deskSlots[i][1])
	}
	s.WriteString(`</svg>`)
	return s.String()
}

func drawDeskItem(s *strings.Builder, it string, x, y float64) {
	switch it {
	case "mug":
		fmt.Fprintf(s, `<rect x="%.0f" y="%.0f" width="22" height="26" rx="3" fill="#fff" stroke="#888"/><path d="M%.0f %.0f c9 0 9 14 0 14" fill="none" stroke="#888" stroke-width="3"/><path d="M%.0f %.0f q-4 -6 0 -12 M%.0f %.0f q-4 -6 0 -12" fill="none" stroke="#aaa"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="5" font-weight="700" fill="#0a1628">ACPWB</text>`,
			x-11, y-13, x+11, y-7, x-4, y-16, x+4, y-16, x, y+2, svgFont)
	case "plant":
		fmt.Fprintf(s, `<path d="M%.0f %.0f h20 l-3 18 h-14z" fill="#c4683b"/><ellipse cx="%.0f" cy="%.0f" rx="7" ry="14" fill="#3e8e41" transform="rotate(-25 %.0f %.0f)"/><ellipse cx="%.0f" cy="%.0f" rx="7" ry="15" fill="#4caf50" transform="rotate(20 %.0f %.0f)"/><ellipse cx="%.0f" cy="%.0f" rx="6" ry="13" fill="#2e7d32"/>`,
			x-10, y-2, x-7, y-14, x-7, y-14, x+7, y-14, x+7, y-14, x, y-18)
	case "photo":
		fmt.Fprintf(s, `<rect x="%.0f" y="%.0f" width="36" height="28" fill="#6d4c2f"/><rect x="%.0f" y="%.0f" width="28" height="20" fill="#bfe0f2"/><circle cx="%.0f" cy="%.0f" r="3.5" fill="#7a5a44"/><circle cx="%.0f" cy="%.0f" r="3" fill="#a9785a"/><rect x="%.0f" y="%.0f" width="22" height="4" fill="#7cb342"/>`,
			x-18, y-16, x-14, y-12, x-5, y-4, x+5, y-3, x-11, y+4)
	case "pens":
		fmt.Fprintf(s, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#1565c0" stroke-width="2.5"/><line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#c62828" stroke-width="2.5"/><line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#222" stroke-width="2.5"/><rect x="%.0f" y="%.0f" width="18" height="22" fill="#37474f"/>`,
			x-4, y-6, x-8, y-22, x, y-6, x+1, y-24, x+4, y-6, x+9, y-20, x-9, y-8)
	case "stapler":
		fmt.Fprintf(s, `<rect x="%.0f" y="%.0f" width="44" height="8" rx="2" fill="#333"/><path d="M%.0f %.0f h40 a4 4 0 0 1 0 8 h-40z" fill="#c62828"/>`, x-22, y, x-22, y-8)
	case "headphones":
		fmt.Fprintf(s, `<path d="M%.0f %.0f a16 16 0 0 1 32 0" fill="none" stroke="#333" stroke-width="4"/><rect x="%.0f" y="%.0f" width="9" height="14" rx="3" fill="#333"/><rect x="%.0f" y="%.0f" width="9" height="14" rx="3" fill="#333"/>`,
			x-16, y+2, x-20, y-2, x+11, y-2)
	case "papers":
		fmt.Fprintf(s, `<rect x="%.0f" y="%.0f" width="46" height="32" fill="#fafafa" stroke="#bbb" transform="rotate(-6 %.0f %.0f)"/><rect x="%.0f" y="%.0f" width="46" height="32" fill="#fff" stroke="#bbb"/>`, x-21, y-18, x, y, x-23, y-14)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(s, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#ccc"/>`, x-19, y-8+float64(i)*5, x+19, y-8+float64(i)*5)
		}
		fmt.Fprintf(s, `<text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="6.5" font-weight="700" fill="#c62828" transform="rotate(-8 %.0f %.0f)">CONFIDENTIAL</text>`, x, y+12, svgFont, x, y+12)
	case "usb":
		fmt.Fprintf(s, `<rect x="%.0f" y="%.0f" width="28" height="11" rx="2" fill="#1565c0"/><rect x="%.0f" y="%.0f" width="9" height="7" fill="#bdbdbd"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="5" fill="#fff">BACKUP</text>`, x-16, y-5, x+12, y-3, x-2, y+2.5, svgFont)
	case "badge":
		fmt.Fprintf(s, `<path d="M%.0f %.0f q-14 -22 6 -26 q20 4 6 26" fill="none" stroke="#0a1628" stroke-width="2"/><rect x="%.0f" y="%.0f" width="22" height="30" rx="2" fill="#fff" stroke="#888"/><rect x="%.0f" y="%.0f" width="22" height="7" fill="#c9a84c"/><rect x="%.0f" y="%.0f" width="10" height="10" fill="#9aa5b1"/>`,
			x-6, y-14, x-11, y-14, x-11, y-14, x-5, y-4)
	case "keys":
		fmt.Fprintf(s, `<circle cx="%.0f" cy="%.0f" r="7" fill="none" stroke="#9e9e9e" stroke-width="2"/><path d="M%.0f %.0f l16 6 l2 -3 M%.0f %.0f l14 10" stroke="#b0892f" stroke-width="3" fill="none"/><rect x="%.0f" y="%.0f" width="16" height="9" fill="#fff8c4" stroke="#aaa"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="5" fill="#333">CAB 4</text>`,
			x-8, y-4, x-2, y-2, x-2, y+1, x-26, y+2, x-18, y+8.5, svgFont)
	}
}

// ── pay range band ───────────────────────────────────────────────────────

func payRangeSVG(job string, lo, mid, hi int) string {
	axLo, axHi := float64(lo)*0.8, float64(hi)*1.15
	xp := func(v int) float64 { return 30 + (float64(v)-axLo)/(axHi-axLo)*380 }
	var s strings.Builder
	s.WriteString(`<svg class="cmp-exhibit-img" width="440" height="110" viewBox="0 0 440 110" role="img" aria-label="Posted pay range">`)
	s.WriteString(`<rect width="440" height="110" fill="#fff" stroke="#c8ccd4"/>`)
	fmt.Fprintf(&s, `<text x="220" y="18" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">Posted Pay Range — %s</text>`, svgFont, esc(truncate(job, 44)))
	s.WriteString(`<line x1="30" y1="60" x2="410" y2="60" stroke="#999"/>`)
	fmt.Fprintf(&s, `<rect x="%.1f" y="46" width="%.1f" height="28" fill="#f3e7c2" stroke="#c9a84c"/>`, xp(lo), xp(hi)-xp(lo))
	for i, v := range []int{lo, mid, hi} {
		x := xp(v)
		fmt.Fprintf(&s, `<line x1="%.1f" y1="40" x2="%.1f" y2="80" stroke="#0a1628" stroke-width="%s"/>`, x, x, []string{"1.5", "1", "1.5"}[i])
		fmt.Fprintf(&s, `<text x="%.1f" y="94" text-anchor="middle" %s font-size="9" fill="#333">%s</text>`, x, svgFont, dollars(v))
		fmt.Fprintf(&s, `<text x="%.1f" y="36" text-anchor="middle" %s font-size="8" fill="#666">%s</text>`, x, svgFont, []string{"Minimum", "Midpoint", "Maximum"}[i])
	}
	s.WriteString(`</svg>`)
	return s.String()
}
