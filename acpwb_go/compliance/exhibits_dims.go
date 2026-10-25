package compliance

import (
	"fmt"
	"math"
)

// Dimension exhibits: four side-by-side installations, each with a
// measured dimension; one meets (or breaks) the rule. Measurements are the
// real ADA 2010 Standards, OSHA 1910/1926, IBC, ANSI Z358.1, CPSC, NEC,
// and OSHA ergonomics eTool figures.

type dimSpec struct {
	title, label, noun, unit string
	style                    string  // height, gap, width, depth, above
	lo, hi, dmax             float64 // acceptable range; drawing scale max
	step                     float64
	ft, refBottom            bool   // ft: show as ft/in; refBottom: height measured to the object's bottom
	onlyOne                  bool   // the question always asks for the one in range
	rule, askOK, askBad      string // askOK when exactly one is in range
	other                    string // second object in gap style
	cats                     []string
}

var dimSpecs = map[string]dimSpec{
	"handrailht": {title: "STAIR HANDRAILS", label: "HANDRAIL", noun: "Handrail", style: "height", lo: 34, hi: 38, dmax: 48, step: 1,
		rule: "ADA 505.4: the top of a handrail must be 34–38 in above the stair nosings.", askOK: "Which handrail is mounted at a compliant height?", askBad: "Which handrail is mounted at the wrong height?",
		cats: []string{"accessibility", "slips", "facilities"}},
	"guardrailht": {title: "GUARDRAILS ON A MEZZANINE", label: "TOP RAIL", noun: "Guardrail", style: "height", lo: 39, hi: 45, dmax: 54, step: 1,
		rule: "OSHA 1910.29(b)(1): the top rail must be 42 in, plus or minus 3 in, above the walking-working surface.", askOK: "Which guardrail meets the OSHA height requirement?", askBad: "Which guardrail does NOT meet the OSHA height requirement?",
		cats: []string{"fallprotection", "warehouse", "safety"}},
	"grabbar": {title: "RESTROOM GRAB BARS", label: "GRAB BAR", noun: "Restroom", style: "height", lo: 33, hi: 36, dmax: 48, step: 1,
		rule: "ADA 609.4: grab bars must be 33–36 in above the floor, measured to the top of the gripping surface.", askOK: "Which grab bar is installed at a compliant height?", askBad: "Which grab bar is installed at the wrong height?",
		cats: []string{"accessibility", "accommodation", "facilities"}},
	"counterht": {title: "SERVICE COUNTERS", label: "COUNTER", noun: "Counter", style: "height", lo: 28, hi: 36, dmax: 48, step: 1,
		rule: "ADA 904.4.1: an accessible section of a service counter can be no higher than 36 in.", askOK: "Which counter has an accessible height?", askBad: "Which counter is too high to be the accessible section?",
		cats: []string{"accessibility", "lobby", "customers"}},
	"reachrange": {title: "BADGE READERS ON AN UNOBSTRUCTED WALL", label: "READER", noun: "Reader", style: "height", lo: 15, hi: 48, dmax: 64, step: 1,
		rule: "ADA 308.2.1: an unobstructed forward reach must be no higher than 48 in and no lower than 15 in.", askOK: "Which badge reader is within the ADA reach range?", askBad: "Which badge reader is outside the ADA reach range?",
		cats: []string{"accessibility", "badge", "physicalsec"}},
	"mirrorht": {title: "RESTROOM MIRRORS (ABOVE LAVATORIES)", label: "MIRROR", noun: "Mirror", style: "height", lo: 0, hi: 40, dmax: 56, step: 1, refBottom: true,
		rule: "ADA 603.3: the bottom edge of the reflecting surface must be no higher than 40 in above the floor.", askOK: "Which mirror meets the ADA requirement?", askBad: "Which mirror is mounted too high?",
		cats: []string{"accessibility", "facilities"}},
	"coathook": {title: "COAT HOOKS IN AN ACCESSIBLE STALL", label: "HOOK", noun: "Hook", style: "height", lo: 15, hi: 48, dmax: 64, step: 1,
		rule: "ADA 603.4: coat hooks must be within reach range, 15–48 in above the floor.", askOK: "Which hook is within reach range?", askBad: "Which hook is out of reach range?",
		cats: []string{"accessibility", "accommodation"}},
	"fountain": {title: "DRINKING FOUNTAINS (WHEELCHAIR ACCESSIBLE)", label: "SPOUT", noun: "Fountain", style: "height", lo: 0, hi: 36, dmax: 48, step: 1,
		rule: "ADA 602.4: the spout outlet of a wheelchair-accessible fountain must be no higher than 36 in.", askOK: "Which fountain meets the ADA spout height?", askBad: "Which fountain's spout is too high?",
		cats: []string{"accessibility", "wellness", "kitchen"}},
	"toiletseat": {title: "ACCESSIBLE WATER CLOSETS", label: "SEAT", noun: "Stall", style: "height", lo: 17, hi: 19, dmax: 30, step: 0.5,
		rule: "ADA 604.4: the top of the toilet seat must be 17–19 in above the floor.", askOK: "Which stall's seat height is compliant?", askBad: "Which stall's seat height is NOT compliant?",
		cats: []string{"accessibility", "facilities"}},
	"kneeclear": {title: "LAVATORY KNEE CLEARANCE", label: "APRON", noun: "Lavatory", style: "height", lo: 27, hi: 999, dmax: 40, step: 1, refBottom: true,
		rule: "ADA 306.3.3: knee clearance under a lavatory must be at least 27 in high.", askOK: "Which lavatory provides enough knee clearance?", askBad: "Which lavatory has too little knee clearance?",
		cats: []string{"accessibility", "accommodation"}},
	"protrude": {title: "WALL-MOUNTED OBJECTS ALONG A CORRIDOR", label: "PHONE", noun: "Object", style: "gap", lo: 0, hi: 4, dmax: 12, step: 0.5, other: "WALL",
		rule: "ADA 307.2: objects with leading edges 27–80 in above the floor may protrude no more than 4 in into the circulation path.", askOK: "Which object meets the protrusion limit?", askBad: "Which object protrudes too far into the corridor?",
		cats: []string{"accessibility", "slips", "signage"}},
	"headroom": {title: "OVERHEAD SIGNS IN A HALLWAY", label: "SIGN", noun: "Sign", style: "height", lo: 80, hi: 999, dmax: 100, step: 1, ft: true, refBottom: true,
		rule: "ADA 307.4: vertical clearance must be at least 80 in (6 ft 8 in) above the floor.", askOK: "Which sign provides enough headroom?", askBad: "Which sign hangs too low?",
		cats: []string{"accessibility", "signage", "slips"}},
	"doorwidth": {title: "DOORWAYS — CLEAR OPENING WIDTH", label: "DOOR", noun: "Door", style: "width", lo: 32, hi: 999, dmax: 40, step: 1,
		rule: "ADA 404.2.3: doors must provide a clear opening of at least 32 in.", askOK: "Which doorway meets the ADA clear width?", askBad: "Which doorway is too narrow?",
		cats: []string{"accessibility", "officemove", "evacassist"}},
	"exitwidth": {title: "EXIT ROUTE WIDTH", label: "EXIT ROUTE", noun: "Route", style: "width", lo: 28, hi: 999, dmax: 40, step: 1,
		rule: "OSHA 1910.36(g)(2): an exit access route must be at least 28 in wide at all points.", askOK: "Which exit route is wide enough?", askBad: "Which exit route is too narrow?",
		cats: []string{"fire", "evacassist", "warehouse"}},
	"exitceiling": {title: "EXIT ROUTE CEILING HEIGHT", label: "CEILING", noun: "Corridor", style: "height", lo: 90, hi: 999, dmax: 110, step: 1, ft: true, refBottom: true,
		rule: "OSHA 1910.36(g)(1): an exit route must have a ceiling at least 7 ft 6 in high.", askOK: "Which corridor meets the exit route ceiling height?", askBad: "Which corridor's ceiling is too low for an exit route?",
		cats: []string{"fire", "evacassist", "construction"}},
	"accessaisle": {title: "ACCESSIBLE PARKING — ACCESS AISLES", label: "AISLE", noun: "Space", style: "width", lo: 60, hi: 999, dmax: 84, step: 1,
		rule: "ADA 502.3.1: access aisles must be at least 60 in wide.", askOK: "Which access aisle is wide enough?", askBad: "Which access aisle is too narrow?",
		cats: []string{"parking", "accessibility"}},
	"clearfloor": {title: "CLEAR FLOOR SPACE AT A COPIER", label: "CLEAR SPACE", noun: "Copier", style: "width", lo: 30, hi: 999, dmax: 40, step: 1,
		rule: "ADA 305.3: a clear floor space must be at least 30 in by 48 in.", askOK: "Which copier has a wide enough clear floor space?", askBad: "Which copier's clear floor space is too narrow?",
		cats: []string{"accessibility", "printing"}},
	"riserht": {title: "STAIR RISERS", label: "RISER", noun: "Stair", style: "height", lo: 4, hi: 7, dmax: 10, step: 0.25,
		rule: "IBC 1011.5.2: stair risers must be 4 to 7 in high.", askOK: "Which stair's risers meet the building code?", askBad: "Which stair's risers are too tall?",
		cats: []string{"slips", "construction", "facilities"}},
	"treaddepth": {title: "STAIR TREADS", label: "TREAD", noun: "Stair", style: "width", lo: 11, hi: 999, dmax: 15, step: 0.25,
		rule: "IBC 1011.5.2: stair treads must be at least 11 in deep.", askOK: "Which stair's treads are deep enough?", askBad: "Which stair's treads are too shallow?",
		cats: []string{"slips", "construction"}},
	"ladderext": {title: "EXTENSION LADDERS AT A ROOF LANDING", label: "LADDER", noun: "Ladder", style: "above", lo: 36, hi: 999, dmax: 54, step: 1, ft: true,
		rule: "OSHA 1926.1053(b)(1): ladder side rails must extend at least 3 ft above the upper landing.", askOK: "Which ladder extends far enough above the landing?", askBad: "Which ladder does NOT extend far enough above the landing?",
		cats: []string{"ladder", "fallprotection", "construction"}},
	"toeboard": {title: "TOEBOARDS ON AN ELEVATED PLATFORM", label: "TOEBOARD", noun: "Platform", style: "height", lo: 3.5, hi: 999, dmax: 6, step: 0.5,
		rule: "OSHA 1910.29(k)(1)(ii): toeboards must be at least 3.5 in high.", askOK: "Which toeboard is tall enough?", askBad: "Which toeboard is too short?",
		cats: []string{"fallprotection", "warehouse"}},
	"heaterclear": {title: "SPACE HEATERS UNDER DESKS", label: "HEATER", noun: "Desk", style: "gap", lo: 36, hi: 999, dmax: 48, step: 1, ft: true, other: "CURTAIN",
		rule: "CPSC: keep space heaters at least 3 ft from anything that can burn.", askOK: "Which heater is placed safely?", askBad: "Which heater is too close to something that can burn?",
		cats: []string{"fire", "winter", "electrical"}},
	"gfci": {title: "RECEPTACLES NEAR A BREAK ROOM SINK", label: "OUTLET", noun: "Receptacle", style: "gap", lo: 0, hi: 72, dmax: 120, step: 1, ft: true, other: "SINK", onlyOne: true,
		rule: "NEC 210.8: receptacles within 6 ft of the edge of a sink require GFCI protection.", askOK: "Which receptacle requires GFCI protection?",
		cats: []string{"electrical", "kitchen"}},
	"trenchprot": {title: "TRENCH CROSS-SECTIONS (STABLE ROCK EXCLUDED)", label: "TRENCH", noun: "Trench", style: "depth", lo: 60, hi: 999, dmax: 96, step: 1, ft: true, onlyOne: true,
		rule: "OSHA 1926.652: trenches 5 ft deep or more need a protective system (sloping, shoring, or a shield).", askOK: "Which trench requires a protective system?",
		cats: []string{"construction", "fieldwork", "spaces"}},
	"spoilpile": {title: "SPOIL PILES BESIDE OPEN TRENCHES", label: "SPOIL", noun: "Trench", style: "gap", lo: 24, hi: 999, dmax: 48, step: 1, ft: true, other: "EDGE",
		rule: "OSHA 1926.651(j)(2): excavated material must be kept at least 2 ft from the edge of the excavation.", askOK: "Which spoil pile is placed correctly?", askBad: "Which spoil pile is too close to the edge?",
		cats: []string{"construction", "fieldwork"}},
	"fallht": {title: "UNPROTECTED PLATFORM EDGES (GENERAL INDUSTRY)", label: "PLATFORM", noun: "Platform", style: "height", lo: 48, hi: 999, dmax: 72, step: 1, ft: true, onlyOne: true,
		rule: "OSHA 1910.28(b)(1): general industry requires fall protection at unprotected sides 4 ft or more above a lower level.", askOK: "Which platform requires fall protection?",
		cats: []string{"fallprotection", "warehouse", "ladder"}},
	"scaffoldht": {title: "SCAFFOLD PLATFORM HEIGHTS", label: "SCAFFOLD", noun: "Scaffold", style: "height", lo: 120, hi: 999, dmax: 168, step: 1, ft: true, onlyOne: true,
		rule: "OSHA 1926.451(g)(1): scaffolds more than 10 ft above a lower level need fall protection, such as guardrails.", askOK: "Which scaffold platform must have guardrails or other fall protection?",
		cats: []string{"construction", "fallprotection"}},
	"plankoverhang": {title: "SCAFFOLD PLANK OVERHANG", label: "PLANK END", noun: "Plank", style: "gap", lo: 6, hi: 12, dmax: 20, step: 1, other: "SUPPORT",
		rule: "OSHA 1926.451(b)(4): planks must extend at least 6 in and no more than 12 in past their supports.", askOK: "Which plank's overhang is compliant?", askBad: "Which plank's overhang is NOT compliant?",
		cats: []string{"construction", "fallprotection"}},
	"eyewashht": {title: "EYEWASH STATION HEIGHTS", label: "EYEWASH", noun: "Station", style: "height", lo: 33, hi: 53, dmax: 66, step: 1,
		rule: "ANSI Z358.1: eyewash flushing fluid must be delivered 33–53 in above the floor.", askOK: "Which eyewash is installed at a compliant height?", askBad: "Which eyewash is installed at the wrong height?",
		cats: []string{"hazcom", "lab", "chemical", "ppe"}},
	"exitletters": {title: "EXIT SIGN LETTERING", label: "EXIT", noun: "Sign", style: "letters", lo: 6, hi: 999, dmax: 9, step: 0.5,
		rule: "OSHA 1910.37(b)(7): EXIT lettering must be at least 6 in high, with strokes at least 3/4 in wide.", askOK: "Which sign's lettering is tall enough?", askBad: "Which sign's lettering is too small?",
		cats: []string{"fire", "signage", "evacassist"}},
	"monitordist": {title: "MONITOR VIEWING DISTANCE", label: "MONITOR", noun: "Workstation", style: "gap", lo: 20, hi: 40, dmax: 54, step: 1, other: "EYES",
		rule: "OSHA ergonomics guidance: place the monitor 20–40 in from your eyes.", askOK: "Which workstation's monitor is at a recommended distance?", askBad: "Which workstation's monitor is outside the recommended distance?",
		cats: []string{"ergonomics", "eyestrain", "hotdesk"}},
	"tiretread": {title: "FLEET TRUCK STEER TIRES — TREAD DEPTH", label: "TREAD", noun: "Tire", style: "tread", lo: 4, hi: 999, dmax: 14, step: 1, unit: "/32 in",
		rule: "FMCSA 393.75: a truck's front steer tires need at least 4/32 in of tread depth (2/32 in on other positions).", askOK: "Which steer tire has enough tread?", askBad: "Which steer tire must be replaced?",
		cats: []string{"driving", "vehicles", "fleet"}},
	"rampwidth": {title: "ACCESSIBLE RAMP WIDTH", label: "RAMP", noun: "Ramp", style: "width", lo: 36, hi: 999, dmax: 48, step: 1,
		rule: "ADA 405.5: the clear width of a ramp run must be at least 36 in.", askOK: "Which ramp is wide enough?", askBad: "Which ramp is too narrow?",
		cats: []string{"accessibility", "evacassist", "parking"}},
	"ramplanding": {title: "RAMP LANDING LENGTH", label: "LANDING", noun: "Landing", style: "width", lo: 60, hi: 999, dmax: 80, step: 1,
		rule: "ADA 405.7.3: ramp landings must be at least 60 in long.", askOK: "Which landing is long enough?", askBad: "Which landing is too short?",
		cats: []string{"accessibility", "construction"}},
	"railgrip": {title: "HANDRAIL GRIP DIAMETER", label: "RAIL", noun: "Handrail", style: "width", lo: 1.25, hi: 2, dmax: 3, step: 0.25,
		rule: "ADA 505.7.1: circular handrails must have an outside diameter of 1¼ to 2 in.", askOK: "Which handrail has a compliant grip diameter?", askBad: "Which handrail's grip diameter is NOT compliant?",
		cats: []string{"accessibility", "slips"}},
	"railwall": {title: "HANDRAIL CLEARANCE FROM THE WALL", label: "RAIL", noun: "Handrail", style: "gap", lo: 1.5, hi: 999, dmax: 4, step: 0.25, other: "WALL",
		rule: "ADA 505.5: the clearance between a handrail and the wall must be at least 1½ in.", askOK: "Which handrail has enough wall clearance?", askBad: "Which handrail is too close to the wall?",
		cats: []string{"accessibility", "slips"}},
}

func init() {
	gens := map[string]exhibitGen{}
	cats := map[string][]string{}
	for k, sp := range dimSpecs {
		gens[k] = dimGen(k, sp)
		for _, c := range sp.cats {
			cats[c] = append(cats[c], k)
		}
	}
	for _, ks := range cats {
		sortStrings(ks)
	}
	registerExhibits(gens, cats)
}

func (sp dimSpec) show(v float64) string {
	if sp.ft {
		return ftIn(int(math.Round(v)))
	}
	u := sp.unit
	if u == "" {
		u = " in"
	}
	switch {
	case v == math.Trunc(v):
		return fmt.Sprintf("%d%s", int(v), u)
	case v*2 == math.Trunc(v*2):
		return fmt.Sprintf("%.1f%s", v, u)
	}
	return fmt.Sprintf("%.2f%s", v, u)
}

func (sp dimSpec) in(v float64) bool { return v >= sp.lo-1e-9 && v <= sp.hi+1e-9 }

// pick draws a value on the right side of the range, one step clear of
// its edges (so readings are never on the line).
func (sp dimSpec) pickVal(r *rnd, in bool) float64 {
	lo := math.Ceil(math.Max(sp.step, math.Max(sp.dmax*0.15, sp.lo*0.6))/sp.step) * sp.step
	if sp.lo > 0 && lo > sp.lo-4*sp.step {
		lo = sp.step // leave room below the minimum (tread depth)
	}
	for try := 0; ; try++ {
		if try > 10000 {
			panic("dimension value: no room for " + sp.title)
		}
		v := lo + float64(r.intn(int((sp.dmax-lo)/sp.step)+1))*sp.step
		if in && sp.in(v) || !in && !sp.in(v) && math.Abs(v-sp.lo) >= sp.step && math.Abs(v-sp.hi) >= sp.step {
			return v
		}
	}
}

func dimGen(kind string, sp dimSpec) exhibitGen {
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		oneIn := sp.onlyOne || r.chance(0.5)
		odd := r.intn(4)
		vals := make([]float64, 4)
		for i := range vals {
			in := i == odd
			if !oneIn {
				in = i != odd
			}
			for try := 0; ; try++ {
				if try > 1000 {
					panic("dimension values: not enough distinct values for " + sp.title)
				}
				vals[i] = sp.pickVal(r, in)
				dup := false
				for j := 0; j < i; j++ {
					dup = dup || vals[j] == vals[i]
				}
				if !dup {
					break
				}
			}
		}
		s := frame(540, 232, sp.title+" — four installations with dimensions", sp.title)
		for i, v := range vals {
			drawDim(s, sp, 10+132*float64(i), v)
			markAt(s, 22+132*float64(i), 40, i)
		}
		q := Question{key: "ex/" + kind, Exhibit: s.done()}
		rule := ""
		if r.chance(0.65) {
			rule = " " + sp.rule
		}
		ask := sp.askOK
		if !oneIn {
			ask = sp.askBad
		}
		q.Text = fmt.Sprintf("Refer to %s.%s %s", ex, rule, ask)
		q.Choices, q.Correct = letterQ(r, sp.noun, 4, odd)
		return q
	}
}

// dimArrow draws a dimension line with end ticks and a label.
func dimArrow(s *svgB, x1, y1, x2, y2 float64, label string) {
	s.line(x1, y1, x2, y2, "#b91c1c", 1.2)
	if x1 == x2 {
		s.line(x1-4, y1, x1+4, y1, "#b91c1c", 1.2)
		s.line(x2-4, y2, x2+4, y2, "#b91c1c", 1.2)
		my := (y1 + y2) / 2
		w := 5.6*float64(len([]rune(label))) + 6
		s.rect(x1+4, my-7, w, 13, 2, "#fff", "#fecaca")
		s.text(x1+7, my+3, 8.5, `font-weight="700" fill="#b91c1c"`, label)
		return
	}
	s.line(x1, y1-4, x1, y1+4, "#b91c1c", 1.2)
	s.line(x2, y2-4, x2, y2+4, "#b91c1c", 1.2)
	s.text((x1+x2)/2, y1-6, 8.5, `text-anchor="middle" font-weight="700" fill="#b91c1c" stroke="#fff" stroke-width="3" paint-order="stroke"`, label)
}

func drawDim(s *svgB, sp dimSpec, x0, v float64) {
	const fy = 210.0 // floor
	s.rect(x0, 30, 124, 192, 4, "#f8fafc", "#e2e8f0")
	lab := func(x, y float64, t string) {
		s.text(x, y, 7, `text-anchor="middle" font-weight="700" fill="#334155"`, t)
	}
	switch sp.style {
	case "height", "above":
		k := 150 / sp.dmax
		s.line(x0+4, fy, x0+120, fy, "#475569", 2)
		s.raw(`<rect x="%.1f" y="56" width="6" height="%.1f" fill="#cbd5e1"/>`, x0+112, fy-56)
		y := fy - v*k
		if sp.style == "above" {
			ly := fy - 0.5*150
			s.rect(x0+60, ly, 60, 8, 0, "#94a3b8", "")
			lab(x0+90, ly+18, "LANDING")
			y = ly - v*k*0.5
			s.line(x0+56, fy, x0+56, y, "#a16207", 3)
			s.line(x0+70, fy, x0+70, y, "#a16207", 3)
			for ry := fy - 12; ry > y+4; ry -= 14 {
				s.line(x0+56, ry, x0+70, ry, "#a16207", 2)
			}
			dimArrow(s, x0+32, ly, x0+32, y, sp.show(v))
			return
		}
		switch {
		case sp.refBottom:
			s.rect(x0+48, y-22, 58, 22, 2, "#bfdbfe", "#1d4ed8")
			lab(x0+77, y-8, sp.label)
		default:
			s.rect(x0+46, y, 62, 6, 2, "#1d4ed8", "")
			lab(x0+77, y-5, sp.label)
			s.line(x0+56, y+6, x0+56, fy, "#94a3b8", 1)
			s.line(x0+98, y+6, x0+98, fy, "#94a3b8", 1)
		}
		dimArrow(s, x0+24, fy, x0+24, y, sp.show(v))
	case "gap":
		k := 82 / sp.dmax
		s.line(x0+4, fy, x0+120, fy, "#475569", 2)
		s.rect(x0+10, fy-90, 16, 90, 2, "#cbd5e1", "#64748b")
		lab(x0+18, fy-96, sp.other)
		gx := x0 + 26 + v*k
		s.rect(gx, fy-60, 22, 34, 3, "#bfdbfe", "#1d4ed8")
		lab(gx+11, fy-14, sp.label)
		s.dash(x0+26, fy-70, x0+26, fy-40, "#94a3b8", 1)
		dimArrow(s, x0+26, fy-74, gx, fy-74, sp.show(v))
	case "width":
		k := 92 / sp.dmax
		w := v * k
		left := x0 + 62 - w/2
		s.rect(x0+6, 90, left-x0-6, 16, 0, "#475569", "")
		s.rect(left+w, 90, x0+118-left-w, 16, 0, "#475569", "")
		s.rect(left, 106, w, 70, 0, "#e0f2fe", "")
		lab(x0+62, 150, sp.label)
		lab(x0+62, 200, "TOP VIEW")
		dimArrow(s, left, 84, left+w, 84, sp.show(v))
	case "depth":
		k := 140 / sp.dmax
		gy := 60.0
		s.rect(x0+4, gy, 116, fy-gy, 0, "#d6c4a8", "")
		s.rect(x0+40, gy, 44, v*k, 0, "#f8fafc", "#8b6b3e")
		s.line(x0+4, gy, x0+120, gy, "#5b4325", 2)
		lab(x0+62, gy-6, sp.label)
		dimArrow(s, x0+62, gy, x0+62, gy+v*k, sp.show(v))
	case "tread":
		k := 70 / sp.dmax
		s.rect(x0+14, 80, 96, 90, 10, "#374151", "")
		for _, gx := range []float64{x0 + 32, x0 + 56, x0 + 80} {
			s.rect(gx, 80, 12, v*k+2, 0, "#f8fafc", "")
		}
		lab(x0+62, 186, "TIRE CROSS-SECTION")
		s.text(x0+62, 72, 9, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "Tread: "+sp.show(v))
	case "letters":
		s.rect(x0+12, 90, 100, 50, 4, "#fff", "#15803d")
		size := 8 + v*4
		s.text(x0+62, 115+size*0.35, size, `text-anchor="middle" font-weight="700" fill="#15803d"`, "EXIT")
		s.text(x0+62, 168, 9, `text-anchor="middle" font-weight="700" fill="#b91c1c"`, "Letters: "+sp.show(v))
	}
}
