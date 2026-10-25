package compliance

import (
	"fmt"
	"math"
)

// Plan-distance exhibits: a gridded floor or site plan with equipment and
// four lettered locations. Each rule is a distance: a maximum travel
// distance to the nearest item (walk, along the grid), a coverage radius,
// or a minimum separation. Fire, life-safety and OSHA rules are the real
// ones (NFPA 10/72, IBC, OSHA 1910.253, IFC, CDC, ANSI Z358.1).

type planSpec struct {
	title, target, icon, fill, point string
	cell, limit                      float64
	radius, minDist, wall            bool // wall: targets sit on a building's wall
	rule, ask                        string
	cats                             []string
}

var planSpecs = map[string]planSpec{
	"extwalk": {"FLOOR PLAN — FIRE EXTINGUISHERS", "fire extinguisher", "EXT", "#dc2626", "Workstation", 10, 75, false, false, false,
		"NFPA 10: travel distance to a Class A extinguisher must not exceed 75 ft.",
		"Which workstation is more than 75 ft of travel from the nearest extinguisher?", []string{"fire", "safety", "facilities"}},
	"extbwalk": {"FLAMMABLE LIQUIDS AREA — CLASS B EXTINGUISHERS", "10-B extinguisher", "10-B", "#dc2626", "Work position", 5, 50, false, false, false,
		"NFPA 10: in an ordinary hazard area, travel to a 10-B rated extinguisher must not exceed 50 ft.",
		"Which work position is too far from a 10-B extinguisher?", []string{"fire", "chemical", "hazcom", "warehouse"}},
	"extkwalk": {"COMMERCIAL KITCHEN — CLASS K EXTINGUISHER", "Class K extinguisher", "K", "#dc2626", "Appliance", 3, 30, false, false, false,
		"NFPA 10: a Class K extinguisher must be within 30 ft of travel of cooking appliances that use oils and fats.",
		"Which appliance is more than 30 ft of travel from a Class K extinguisher?", []string{"kitchen", "fire", "foodsafety"}},
	"pullwalk": {"FLOOR PLAN — MANUAL PULL STATIONS", "pull station", "PULL", "#dc2626", "Desk", 20, 200, false, false, false,
		"NFPA 72: no point on the floor may be more than 200 ft of travel from a manual fire alarm pull station.",
		"Which desk is more than 200 ft of travel from a pull station?", []string{"fire", "evacassist"}},
	"exitwalk": {"FLOOR PLAN — EXIT ACCESS TRAVEL DISTANCE (SPRINKLERED)", "exit", "EXIT", "#15803d", "Office", 25, 300, false, false, false,
		"In a sprinklered business occupancy, the building code limits exit access travel distance to 300 ft.",
		"Which office exceeds the allowed travel distance to an exit?", []string{"fire", "evacassist", "officemove"}},
	"aedwalk": {"FLOOR PLAN — AED LOCATIONS", "AED", "AED", "#16a34a", "Work area", 20, 300, false, false, false,
		"ACPWB standard: no work area may be more than 300 ft of travel from an AED.",
		"Which work area is too far from an AED?", []string{"cpr", "firstaid", "medical"}},
	"smokezone": {"SITE PLAN — SMOKING AREAS", "entrance", "DOOR", "#92400e", "Bench", 5, 25, true, true, true,
		"Company policy and local ordinance: no smoking within 25 ft of any entrance.",
		"Which bench is too close to an entrance to be a smoking area?", []string{"substances", "wellness", "health", "lobby"}},
	"genspot": {"SITE PLAN — PORTABLE GENERATOR LOCATIONS", "door or window", "WIN", "#0369a1", "Spot", 5, 20, true, true, true,
		"CDC: keep portable generators at least 20 ft from doors, windows, and vents.",
		"Which generator spot is too close to the building openings?", []string{"continuity", "closures", "weather", "construction"}},
	"hydrantpark": {"STREET PLAN — FIRE HYDRANTS", "hydrant", "H", "#dc2626", "Car", 3, 15, true, true, false,
		"In most U.S. cities, parking within 15 ft of a fire hydrant is prohibited.",
		"Which car is parked too close to a hydrant?", []string{"parking", "driving", "fire"}},
	"cylinderstore": {"CYLINDER STORAGE YARD", "oxygen cylinder", "O2", "#15803d", "Acetylene", 2, 20, true, true, false,
		"OSHA 1910.253: stored oxygen cylinders must be at least 20 ft from fuel-gas cylinders, unless separated by a 5-ft fire-rated barrier.",
		"Which acetylene cylinder is stored too close to the oxygen?", []string{"hotwork", "warehouse", "hazcom", "construction"}},
	"firstaidwalk": {"FLOOR PLAN — FIRST AID KITS", "first aid kit", "+", "#16a34a", "Work area", 15, 150, false, false, false,
		"ACPWB standard: a first aid kit within 150 ft of travel of every work area.",
		"Which work area is too far from a first aid kit?", []string{"firstaid", "safety", "injuries"}},
	"spillwalk": {"LAB FLOOR — SPILL KITS", "spill kit", "SPILL", "#ca8a04", "Cabinet", 10, 100, false, false, false,
		"Lab standard: every chemical storage cabinet within 100 ft of travel of a spill kit.",
		"Which cabinet is too far from a spill kit?", []string{"chemical", "hazcom", "lab"}},
	"coffeewalk": {"FLOOR PLAN — COFFEE STATIONS", "coffee station", "CAFÉ", "#78350f", "Desk", 15, 150, false, false, false,
		"Facilities Service Level Agreement: no desk more than 150 ft of travel from coffee.",
		"Which desk falls outside the coffee Service Level Agreement?", []string{"coffee", "kitchen", "officemove", "hotdesk"}},
	"printwalk": {"FLOOR PLAN — SECURE-RELEASE PRINTERS", "secure printer", "PRT", "#334155", "Desk", 10, 60, false, false, false,
		"Print policy: every desk within 60 ft of travel of a secure-release printer.",
		"Which desk is too far from a secure printer?", []string{"printing", "cleandesk", "privacy"}},
	"camcover": {"SITE PLAN — SECURITY CAMERA COVERAGE", "camera", "CAM", "#334155", "Door", 5, 40, true, false, false,
		"Security standard: every exterior door must be within 40 ft of a camera.",
		"Which door is NOT covered by a camera?", []string{"physicalsec", "security", "badge"}},
	"smokedet": {"CEILING PLAN — SMOKE DETECTORS", "smoke detector", "SD", "#64748b", "Location", 3, 21, true, false, false,
		"NFPA 72: on a smooth ceiling, every point must be within 21 ft (0.7 × the 30-ft spacing) of a smoke detector.",
		"Which location is outside smoke detector coverage?", []string{"fire", "facilities"}},
	"wifiap": {"FLOOR PLAN — WI-FI ACCESS POINTS", "access point", "AP", "#7c3aed", "Desk", 10, 60, true, false, false,
		"IT design standard: every desk within 60 ft of an access point.",
		"Which desk is likely to have weak Wi-Fi?", []string{"network", "hotdesk", "devices"}},
	"paspeaker": {"FLOOR PLAN — PUBLIC ADDRESS SPEAKERS", "speaker", "PA", "#b45309", "Work area", 10, 50, true, false, false,
		"Mass notification design: every work area within 50 ft of a PA speaker.",
		"Which work area may not hear emergency announcements?", []string{"closures", "weather", "threat"}},
	"drenchwalk": {"LAB FLOOR — EMERGENCY SHOWERS", "safety shower", "SHWR", "#16a34a", "Acid station", 5, 55, false, false, false,
		"ANSI Z358.1: an emergency shower must be reachable within 10 seconds, about 55 ft of travel.",
		"Which acid station is too far from a safety shower?", []string{"hazcom", "chemical", "lab", "ppe"}},
	"bleedwalk": {"FLOOR PLAN — BLEEDING CONTROL KITS", "bleeding control kit", "BCK", "#b91c1c", "Area", 20, 200, false, false, false,
		"ACPWB standard: a bleeding control kit within 200 ft of travel of every area.",
		"Which area is too far from a bleeding control kit?", []string{"firstaid", "threat", "violence"}},
	"dumpsterspot": {"SITE PLAN — DUMPSTER LOCATIONS", "building wall", "", "#0369a1", "Dumpster", 1, 5, true, true, true,
		"International Fire Code 304.3.3: dumpsters of 1.5 cubic yards or more must be at least 5 ft from combustible walls, openings, and eaves.",
		"Which dumpster location is too close to the building?", []string{"fire", "sustainability", "facilities", "waste"}},
	"grillspot": {"SITE PLAN — COMPANY PICNIC GRILLS", "building wall", "", "#0369a1", "Grill", 2, 10, true, true, true,
		"International Fire Code 308.1.4: charcoal and propane grills may not be used within 10 ft of combustible construction.",
		"Which grill location is too close to the building?", []string{"potluck", "celebrations", "fire", "offsites"}},
	"shuttlewalk": {"CAMPUS MAP — SHUTTLE STOPS", "shuttle stop", "BUS", "#1d4ed8", "Lot entrance", 100, 1000, false, false, false,
		"Commuter program: every parking lot entrance within 1,000 ft of walking to a shuttle stop.",
		"Which lot entrance is too far from a shuttle stop?", []string{"commuting", "parking"}},
	"radiocover": {"WAREHOUSE — RADIO REPEATERS", "repeater", "RPT", "#0f766e", "Aisle", 20, 150, true, false, false,
		"Warehouse radio plan: every aisle within 150 ft of a repeater.",
		"Which aisle is outside repeater coverage?", []string{"warehouse", "deliveries"}},
	"shelterwalk": {"FLOOR PLAN — SEVERE WEATHER SHELTERS", "storm shelter", "SHLT", "#4d7c0f", "Work area", 25, 300, false, false, false,
		"Severe weather plan: every work area within 300 ft of travel of a designated shelter.",
		"Which work area is too far from a storm shelter?", []string{"weather", "closures", "evacassist"}},
}

func init() {
	gens := map[string]exhibitGen{}
	cats := map[string][]string{}
	for k, sp := range planSpecs {
		gens[k] = planGen(k, sp)
		for _, c := range sp.cats {
			cats[c] = append(cats[c], k)
		}
	}
	for _, ks := range cats {
		sortStrings(ks)
	}
	registerExhibits(gens, cats)
}

type cellXY struct{ x, y int }

// planToWall rules measure to the building wall itself, not to targets.
var planToWall = map[string]bool{"dumpsterspot": true, "grillspot": true}

func planGen(kind string, sp planSpec) exhibitGen {
	const cols, rows, px, ox, oy = 16, 9, 28.0, 36.0, 44.0
	toWall := planToWall[kind]
	wallRow := 3 // building occupies rows 0..wallRow-1 on wall plans
	if toWall {
		wallRow = 2
	}
	lim := sp.limit / sp.cell
	// pos is a location in cell units: points sit at cell centres, wall
	// targets on the building's wall line.
	type pos struct{ x, y float64 }
	at := func(p cellXY, onWall bool) pos {
		if onWall {
			return pos{float64(p.x) + 0.5, float64(wallRow)}
		}
		return pos{float64(p.x) + 0.5, float64(p.y) + 0.5}
	}
	dist := func(a, b pos) float64 {
		dx, dy := a.x-b.x, a.y-b.y
		if sp.radius {
			return math.Hypot(dx, dy)
		}
		return math.Abs(dx) + math.Abs(dy)
	}
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		var targets, pts []pos
		var used []cellXY
		near := func(p pos) float64 {
			if toWall {
				return p.y - float64(wallRow)
			}
			best := 1e9
			for _, t := range targets {
				best = math.Min(best, dist(p, t))
			}
			return best
		}
		free := func(p cellXY) bool {
			for _, q := range used {
				if abs(q.x-p.x) <= 1 && abs(q.y-p.y) <= 1 {
					return false
				}
			}
			return true
		}
		randCell := func() cellXY {
			if sp.wall {
				return cellXY{r.intn(cols), r.between(wallRow, rows-1)}
			}
			return cellXY{r.intn(cols), r.intn(rows)}
		}
		var bad int
		for {
			targets, pts, used = nil, nil, nil
			nt := r.between(2, 3)
			if sp.minDist {
				nt = r.between(2, 4)
			}
			if toWall {
				nt = 0
			}
			for len(targets) < nt {
				p := randCell()
				if sp.wall {
					p.y = wallRow - 1
				}
				if free(p) {
					targets = append(targets, at(p, sp.wall))
					used = append(used, p)
				}
			}
			bad = r.intn(4)
			ok := true
			for i := 0; i < 4 && ok; i++ {
				placed := false
				for try := 0; try < 400; try++ {
					p := randCell()
					if !free(p) {
						continue
					}
					d := near(at(p, false))
					violates, fine := d > lim+0.4, d < lim-0.4
					if sp.minDist {
						violates, fine = d < lim-0.4, d > lim+0.4
					}
					if (i == bad && violates) || (i != bad && fine) {
						pts = append(pts, at(p, false))
						used = append(used, p)
						placed = true
						break
					}
				}
				ok = placed
			}
			if ok {
				break
			}
		}
		X := func(p pos) float64 { return ox + p.x*px }
		Y := func(p pos) float64 { return oy + p.y*px }
		s := frame(520, 330, sp.title+" with four lettered locations", sp.title)
		s.raw(`<defs><clipPath id="clip-%s"><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f"/></clipPath></defs>`, kind, ox, oy, cols*px, rows*px)
		s.rect(ox, oy, cols*px, rows*px, 0, "#f8fafc", "")
		for i := 1; i < cols; i++ {
			s.line(ox+float64(i)*px, oy, ox+float64(i)*px, oy+rows*px, "#e2e8f0", 1)
		}
		for i := 1; i < rows; i++ {
			s.line(ox, oy+float64(i)*px, ox+cols*px, oy+float64(i)*px, "#e2e8f0", 1)
		}
		if sp.wall {
			wy := float64(wallRow) * px
			s.rect(ox, oy, cols*px, wy, 0, "#cbd5e1", "")
			s.line(ox, oy+wy, ox+cols*px, oy+wy, "#334155", 4)
			s.text(ox+cols*px/2, oy+wy/2+4, 11, `text-anchor="middle" font-weight="700" fill="#475569" letter-spacing="3"`, "BUILDING")
		}
		s.raw(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#334155" stroke-width="3"/>`, ox, oy, cols*px, rows*px)
		if sp.radius {
			s.raw(`<g clip-path="url(#clip-%s)">`, kind)
			for _, t := range targets {
				s.raw(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" fill-opacity=".07" stroke="%s" stroke-dasharray="4 3"/>`, X(t), Y(t), lim*px, sp.fill, sp.fill)
			}
			s.raw(`</g>`)
		}
		for _, t := range targets {
			w := 10 + 6.2*float64(len([]rune(sp.icon)))
			s.rect(X(t)-w/2, Y(t)-8, w, 16, 3, sp.fill, "#fff")
			s.text(X(t), Y(t)+3.5, 8.5, `text-anchor="middle" font-weight="700" fill="#fff"`, sp.icon)
		}
		for i, p := range pts {
			markAt(s, X(p), Y(p), i)
		}
		note := fmt.Sprintf("1 square = %s ft · %s = %s", commas(int(sp.cell)), sp.icon, sp.target)
		if toWall {
			note = fmt.Sprintf("1 square = %s ft · the building wall is combustible construction", commas(int(sp.cell)))
		} else if !sp.radius {
			note += " · measure travel along the grid (no diagonal shortcuts)"
		} else if sp.minDist {
			note += fmt.Sprintf(" · dashed circles show %s ft", commas(int(sp.limit)))
		} else {
			note += fmt.Sprintf(" · dashed circles show the %s-ft range", commas(int(sp.limit)))
		}
		s.text(ox, 318, 8.5, `fill="#334155"`, note)
		q := Question{key: "ex/" + kind, Exhibit: s.done()}
		rule := ""
		if r.chance(0.65) {
			rule = " " + sp.rule
		}
		if !sp.radius && r.chance(0.35) {
			i := r.intn(4)
			d := int(math.Round(near(pts[i]) * sp.cell))
			cell := int(sp.cell)
			q.Text = fmt.Sprintf("Refer to %s. How far must someone at %s %s travel to reach the nearest %s?", ex, lowerFirst(sp.point), markLetters[i], sp.target)
			q.Choices, q.Correct = numChoices(r, d, []int{cell, -cell, 2 * cell, -2 * cell, 3 * cell}, func(n int) string { return commas(n) + " ft" })
			return q
		}
		q.Text = fmt.Sprintf("Refer to %s.%s %s", ex, rule, sp.ask)
		q.Choices, q.Correct = letterQ(r, sp.point, 4, bad)
		return q
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
