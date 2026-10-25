package compliance

import (
	"fmt"
	"math"
)

// Gauge exhibits: four dial instruments, each read against a stated rule
// (FDA Food Code, EPA, OSHA, ASHRAE, CDC, USDA). The question asks which
// reading is out of range, how many are in range, or what one dial reads.

type gItem struct {
	name   string
	lo, hi float64 // zero both → the spec's range
}

type gaugeSpec struct {
	title, unit, noun       string
	dmin, dmax, major, step float64
	dec                     int
	lo, hi, margin          float64
	items                   []gItem
	rule, askOut, askCount  string
	cats                    []string
}

var inf = math.Inf(1)

func gi(names ...string) []gItem {
	out := make([]gItem, len(names))
	for i, n := range names {
		out[i] = gItem{name: n}
	}
	return out
}

var gaugeSpecs = map[string]gaugeSpec{
	"coldhold": {"COLD HOLDING — DIAL THERMOMETERS", "°F", "unit", 20, 60, 10, 1, 0, -inf, 41, 2,
		gi("Salad bar", "Deli case", "Sandwich rail", "Milk cooler", "Dessert case", "Walk-in cooler"),
		"The FDA Food Code requires cold food to be held at 41°F or below.",
		"Which unit is NOT holding food at a safe temperature?", "How many of these units are holding food at a safe temperature?",
		[]string{"foodsafety", "potluck", "kitchen", "hospitality"}},
	"hothold": {"HOT HOLDING — DIAL THERMOMETERS", "°F", "unit", 100, 200, 20, 2, 0, 135, inf, 4,
		gi("Steam table", "Soup well", "Chafing dish", "Rice warmer", "Carving station", "Heat lamp shelf"),
		"The FDA Food Code requires hot food to be held at 135°F or above.",
		"Which unit is NOT holding food hot enough?", "How many of these units are holding food at a safe temperature?",
		[]string{"foodsafety", "potluck", "hospitality", "catering"}},
	"freezer": {"FREEZER THERMOMETERS", "°F", "freezer", -20, 40, 10, 1, 0, -inf, 0, 2,
		gi("Walk-in freezer", "Ice cream chest", "Reach-in freezer", "Break room freezer", "Catering freezer"),
		"USDA: freezers should hold food at 0°F or below.",
		"Which freezer is too warm?", "How many freezers are cold enough?",
		[]string{"foodsafety", "kitchen", "catering"}},
	"handsink": {"HAND SINK WATER TEMPERATURE", "°F", "sink", 60, 140, 20, 2, 0, 100, 120, 3,
		gi("Kitchen hand sink", "Café hand sink", "Bar hand sink", "Catering sink", "Break room sink"),
		"Hand sinks must supply water of at least 100°F (FDA Food Code), and Facilities caps tap water at 120°F to prevent scalds.",
		"Which sink's water temperature is outside the acceptable range?", "How many sinks are within the acceptable range?",
		[]string{"foodsafety", "kitchen", "health"}},
	"co2": {"MEETING ROOM CO₂ MONITORS", " ppm", "room", 400, 2000, 400, 50, 0, -inf, 1000, 100,
		gi("Boardroom", "Huddle 4B", "Training Room", "Quiet Room", "Phone Booth 7", "War Room"),
		"ACPWB Facilities increases ventilation in any room above 1,000 ppm CO₂, following ASHRAE indoor air guidance.",
		"Which room needs more ventilation?", "How many rooms are at or below the 1,000 ppm target?",
		[]string{"airquality", "meetings", "health", "facilities"}},
	"humidity": {"HYGROMETERS — RELATIVE HUMIDITY", "% RH", "space", 0, 100, 20, 2, 0, 30, 60, 4,
		gi("Records Room", "Open Office 3", "Library", "Art Storage", "Mail Room"),
		"The EPA recommends keeping indoor relative humidity between 30% and 60%.",
		"Which space is outside the EPA's recommended humidity range?", "How many spaces are within the recommended range?",
		[]string{"airquality", "records", "facilities"}},
	"serverinlet": {"DATA CENTER INLET TEMPERATURES", "°F", "reading", 50, 100, 10, 1, 0, 64.4, 80.6, 2,
		gi("Row 1 inlet", "Row 2 inlet", "Row 3 inlet", "Row 4 inlet", "Network closet"),
		"ASHRAE recommends server inlet air between 64.4°F and 80.6°F (18–27°C).",
		"Which reading is outside the ASHRAE recommended range?", "How many readings are within the recommended range?",
		[]string{"serverroom", "continuity", "network", "backups"}},
	"aqi": {"OUTDOOR CREW LOCATIONS — AIR QUALITY INDEX", " AQI", "location", 0, 300, 50, 5, 0, -inf, 150, 10,
		gi("North lot crew", "Roof crew", "Grounds crew", "Loading dock", "Garden crew"),
		"Under Cal/OSHA's wildfire smoke rule, protective measures are required when the AQI for PM2.5 is 151 or higher.",
		"Which crew's location requires wildfire smoke protections?", "At how many locations is the AQI below the level that requires protections?",
		[]string{"airquality", "fieldwork", "weather", "heat"}},
	"uvindex": {"JOB SITE UV INDEX", " UV", "site", 0, 12, 2, 1, 0, -inf, 2.5, 0.5,
		gi("Parking deck", "Rooftop garden", "Athletic field", "Dock yard", "Solar array"),
		"The EPA recommends sun protection whenever the UV Index is 3 or higher.",
		"At which site is sun protection recommended?", "At how many sites is sun protection NOT yet needed?",
		[]string{"heat", "fieldwork", "sports", "wellness"}},
	"radon": {"RADON TEST RESULTS", " pCi/L", "space", 0, 10, 2, 0.2, 1, -inf, 4, 0.4,
		gi("Basement Archive", "B1 Mail Room", "Parking Office", "Fitness Center", "Records Vault"),
		"The EPA recommends fixing a building when radon is 4 pCi/L or higher.",
		"Which space needs radon mitigation?", "How many spaces are below the EPA action level?",
		[]string{"airquality", "health", "records"}},
	"coalarm": {"CARBON MONOXIDE — 8-HOUR AVERAGE", " ppm", "area", 0, 100, 20, 2, 0, -inf, 50, 4,
		gi("Garage P1", "Garage P2", "Loading Dock", "Generator Room", "Shuttle Bay"),
		"OSHA's permissible exposure limit for carbon monoxide is 50 ppm averaged over 8 hours.",
		"Which area is over the OSHA limit?", "How many areas are within the OSHA limit?",
		[]string{"airquality", "parking", "warehouse", "spaces"}},
	"tirepressure": {"PRE-TRIP INSPECTION — TIRE PRESSURE", " psi", "tire", 0, 60, 10, 1, 0, 33, 37, 1,
		gi("Van 12 FL", "Van 12 FR", "Van 12 RL", "Van 12 RR", "Spare"),
		"Fleet policy: keep tires within 2 psi of the door placard pressure of 35 psi.",
		"Which tire must be adjusted before the trip?", "How many tires are within 2 psi of the placard?",
		[]string{"driving", "vehicles", "fleet", "deliveries"}},
	"fireriser": {"SPRINKLER RISER GAUGES", " psi", "riser", 0, 200, 40, 5, 0, 60, 90, 5,
		gi("Riser 1 North", "Riser 2 South", "Riser 3 Garage", "Riser 4 Annex"),
		"The inspection tag lists normal static pressure for these risers as 60–90 psi.",
		"Which riser reading should be reported to Facilities?", "How many risers read within the normal range?",
		[]string{"fire", "facilities", "construction"}},
	"oxygencyl": {"FIRST-AID OXYGEN CYLINDERS", " psi", "cylinder", 0, 3000, 500, 50, 0, 500, inf, 100,
		gi("Lobby first aid", "Fitness center", "Security desk", "Clinic room 2", "Plant first aid"),
		"Site medical protocol: replace any first-aid oxygen cylinder that reads below 500 psi.",
		"Which cylinder must be replaced?", "How many cylinders are ready for use?",
		[]string{"firstaid", "medical", "cpr"}},
	"airnozzle": {"SHOP AIR REGULATORS", " psi", "air line", 0, 120, 20, 2, 0, -inf, 30, 4,
		gi("Bench 1 blow gun", "Bench 2 blow gun", "Paint booth", "Print shop", "Maintenance cart"),
		"OSHA 1910.242(b): compressed air used for cleaning must be reduced to less than 30 psi.",
		"Which air line is set too high to use for cleaning?", "How many air lines are safe to use for cleaning?",
		[]string{"machineguarding", "safety", "ppe", "warehouse"}},
	"dishrinse": {"DISH MACHINE FINAL RINSE", "°F", "machine", 120, 220, 20, 2, 0, 180, inf, 3,
		gi("Cafeteria", "Catering", "Bistro", "Exec dining", "Pot room"),
		"These high-temperature dish machines require a final rinse of at least 180°F.",
		"Which machine is not sanitizing properly?", "How many machines meet the final rinse minimum?",
		[]string{"foodsafety", "kitchen", "hospitality"}},
	"sanitizer": {"SANITIZER TEST STRIP READER", " ppm", "bucket", 0, 200, 50, 5, 0, 50, 100, 5,
		gi("Prep bucket", "Bar bucket", "Deli bucket", "Bakery bucket", "Grill bucket"),
		"Chlorine sanitizer for wiping cloths should test between 50 and 100 ppm.",
		"Which bucket must be remade?", "How many buckets are mixed correctly?",
		[]string{"foodsafety", "kitchen", "chemical"}},
	"liftwind": {"AERIAL LIFT ANEMOMETERS", " mph", "lift", 0, 50, 10, 1, 0, -inf, 28, 2,
		gi("Lift 1 · Lot A", "Lift 2 · Atrium", "Lift 3 · Dock", "Lift 4 · Garage", "Lift 5 · Roof"),
		"The aerial lifts' decals rate them for winds up to 28 mph (12.5 m/s).",
		"Which lift must be lowered and stopped?", "How many lifts may keep working?",
		[]string{"fallprotection", "construction", "weather", "fieldwork"}},
	"filterdp": {"AIR HANDLER FILTER PRESSURE DROP", " in. w.c.", "air handler", 0, 2, 0.5, 0.05, 2, -inf, 1.0, 0.1,
		gi("AHU-1", "AHU-2", "AHU-3", "AHU-4", "AHU-5"),
		"Facilities replaces air filters when the pressure drop reaches 1.0 in. w.c.",
		"Which air handler needs a filter change?", "How many air handlers can keep their current filters?",
		[]string{"airquality", "facilities", "maintenance"}},
	"coffeetemp": {"COFFEE BREWER WATER TEMPERATURE", "°F", "brewer", 150, 230, 20, 2, 0, 195, 205, 2,
		gi("3rd Floor Brewer", "Lobby Café", "Exec Suite", "Break Room 2", "Training Center"),
		"The Specialty Coffee Association recommends brewing water between 195°F and 205°F.",
		"Which brewer is outside the recommended range?", "How many brewers are within the recommended range?",
		[]string{"coffee", "kitchen", "culture"}},
	"radarspeed": {"CAMPUS RADAR SPEED SIGNS", " mph", "reading", 0, 40, 10, 1, 0, -inf, 15, 1,
		gi("North Drive", "Visitor Loop", "Garage Ramp", "Dock Road", "Shuttle Lane"),
		"The campus speed limit is 15 mph.",
		"On which road was the driver speeding?", "On how many roads was the driver within the limit?",
		[]string{"driving", "parking", "commuting", "vehicles"}},
	"vaccinefridge": {"VACCINE STORAGE THERMOMETERS", "°F", "unit", 20, 60, 10, 1, 0, 36, 46, 1,
		gi("Clinic Fridge 1", "Clinic Fridge 2", "Mobile Cooler", "Pharmacy Fridge"),
		"CDC: store refrigerated vaccines between 36°F and 46°F (2–8°C).",
		"Which vaccine storage unit is out of range?", "How many storage units are in range?",
		[]string{"medical", "pandemic", "wellness"}},
	"cookprobe": {"PROBE THERMOMETERS — COMPANY PICNIC GRILL", "°F", "item", 100, 200, 20, 1, 0, 0, 0, 3,
		[]gItem{{"Chicken thighs", 165, inf}, {"Burgers", 160, inf}, {"Pork chops", 145, inf}, {"Salmon", 145, inf}, {"Turkey breast", 165, inf}, {"Bratwurst", 160, inf}},
		"USDA minimum internal temperatures: poultry 165°F, ground meats 160°F, whole cuts of pork and fish 145°F.",
		"Which item is NOT yet safe to serve?", "How many items have reached a safe internal temperature?",
		[]string{"potluck", "foodsafety", "celebrations", "offsites"}},
	"tepid": {"EYEWASH FLUSHING FLUID TEMPERATURE", "°F", "station", 40, 120, 20, 1, 0, 60, 100, 3,
		gi("Lab A eyewash", "Lab B eyewash", "Shop eyewash", "Battery room", "Loading dock"),
		"ANSI Z358.1 requires tepid flushing fluid, 60°F to 100°F.",
		"Which eyewash station's water is outside the tepid range?", "How many stations deliver tepid water?",
		[]string{"hazcom", "chemical", "ppe", "lab"}},
	"fumehood": {"FUME HOOD FACE VELOCITY MONITORS", " fpm", "hood", 0, 200, 40, 5, 0, 80, 120, 5,
		gi("Hood 1", "Hood 2", "Hood 3", "Hood 4", "Hood 5"),
		"Lab policy: fume hood face velocity must be 80–120 feet per minute.",
		"Which hood should be tagged out of service?", "How many hoods are within the policy range?",
		[]string{"chemical", "lab", "hazcom"}},
	"officetemp": {"OFFICE THERMOSTATS", "°F", "area", 50, 90, 10, 1, 0, 68, 76, 1,
		gi("Open Office 2", "Boardroom", "Call Center", "Reception", "Studio"),
		"OSHA recommends office temperatures between 68°F and 76°F.",
		"Which area is outside OSHA's recommended range?", "How many areas are within the recommended range?",
		[]string{"ergonomics", "openoffice", "airquality", "hotdesk"}},
	"dosimeter": {"NOISE DOSIMETERS — END OF SHIFT", "%", "worker", 0, 200, 40, 5, 0, -inf, 100, 5,
		gi("Press operator", "Forklift driver", "Shredder tech", "Print shop", "Grounds crew"),
		"A noise dose above 100% exceeds OSHA's permissible exposure limit (90 dBA, 8-hour TWA).",
		"Which worker's dosimeter shows an exposure over the OSHA limit?", "How many workers are within the OSHA limit?",
		[]string{"hearing", "warehouse", "ppe", "construction"}},
	"upsruntime": {"UPS RUNTIME AT CURRENT LOAD", " min", "UPS", 0, 60, 10, 1, 0, 15, inf, 2,
		gi("UPS-1 Core", "UPS-2 Storage", "UPS-3 Network", "UPS-4 Phones"),
		"IT standard: every UPS must show at least 15 minutes of runtime at current load.",
		"Which UPS fails the runtime standard?", "How many UPS units meet the standard?",
		[]string{"serverroom", "continuity", "backups", "electrical"}},
	"propanefill": {"PROPANE CYLINDER FILL GAUGES", "%", "cylinder", 0, 100, 20, 2, 0, -inf, 80, 4,
		gi("Tank 1", "Tank 2", "Tank 3", "Tank 4", "Tank 5"),
		"Propane cylinders must not be filled beyond 80% of capacity, leaving room for the liquid to expand.",
		"Which cylinder appears overfilled?", "How many cylinders are filled within the limit?",
		[]string{"forklift", "warehouse", "hazcom", "fire"}},
	"fryer": {"FRYER OIL THERMOMETERS", "°F", "fryer", 250, 450, 50, 5, 0, 325, 375, 5,
		gi("Fryer 1", "Fryer 2", "Fryer 3", "Fryer 4", "Snack bar fryer"),
		"Kitchen standard: hold fryer oil between 325°F and 375°F.",
		"Which fryer is outside the standard range?", "How many fryers are within the standard range?",
		[]string{"kitchen", "fire", "foodsafety"}},
	"parcelscale": {"MAIL ROOM PARCEL SCALE", " lb", "package", 0, 200, 25, 1, 0, -inf, 150, 3,
		gi("Box 1", "Box 2", "Crate 3", "Tube 4", "Box 5"),
		"Ground parcel carriers such as UPS and FedEx accept packages up to 150 lb; heavier ones ship as freight.",
		"Which package must ship as freight?", "How many packages can go by ground parcel service?",
		[]string{"mail", "deliveries", "lifting"}},
	"bagscale": {"AIRPORT CHECK-IN BAG SCALE", " lb", "bag", 0, 80, 10, 1, 0, -inf, 50, 2,
		gi("Bag 1", "Bag 2", "Bag 3", "Bag 4", "Bag 5"),
		"Most U.S. airlines charge an overweight fee for economy checked bags over 50 lb.",
		"Which bag will be charged an overweight fee?", "How many bags are within the 50 lb allowance?",
		[]string{"travel", "travelbooking", "intltravel"}},
}

func init() {
	gens := map[string]exhibitGen{}
	cats := map[string][]string{}
	for k, sp := range gaugeSpecs {
		gens[k] = gaugeGen(k, sp)
		for _, c := range sp.cats {
			cats[c] = append(cats[c], k)
		}
	}
	for _, ks := range cats {
		sortStrings(ks)
	}
	registerExhibits(gens, cats)
}

func (sp gaugeSpec) rangeOf(it gItem) (float64, float64) {
	if it.lo == 0 && it.hi == 0 {
		return sp.lo, sp.hi
	}
	return it.lo, it.hi
}

func (sp gaugeSpec) round(v float64) float64 {
	return math.Round(v/sp.step) * sp.step
}

// value draws a reading inside (in) or clearly outside the item's range.
func (sp gaugeSpec) value(r *rnd, it gItem, in bool) float64 {
	lo, hi := sp.rangeOf(it)
	ok := func(v float64) bool {
		if in {
			return v >= lo+sp.margin-1e-9 && v <= hi-sp.margin+1e-9
		}
		return v <= lo-sp.margin+1e-9 || v >= hi+sp.margin-1e-9
	}
	for i := 0; i < 200; i++ {
		var a, b float64
		if in {
			a, b = math.Max(lo+sp.margin, sp.dmin), math.Min(hi-sp.margin, sp.dmax)
		} else {
			below := lo-sp.margin >= sp.dmin
			above := hi+sp.margin <= sp.dmax
			if below && (!above || r.chance(0.5)) {
				a, b = sp.dmin+sp.step, lo-sp.margin
			} else {
				a, b = hi+sp.margin, sp.dmax-sp.step
			}
		}
		v := sp.round(a + r.float()*(b-a))
		if v >= sp.dmin && v <= sp.dmax && ok(v) {
			return v
		}
	}
	panic("gauge value: no room in range")
}

func (sp gaugeSpec) inRange(it gItem, v float64) bool {
	lo, hi := sp.rangeOf(it)
	return v >= lo && v <= hi
}

func gaugeGen(kind string, sp gaugeSpec) exhibitGen {
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		n := 4
		idx := pickN(r, len(sp.items), n)
		mode := r.intn(3)
		bad, nIn := r.intn(n), r.between(0, n)
		vals := make([]float64, n)
		for i := range vals {
			it := sp.items[idx[i]]
			switch mode {
			case 0:
				vals[i] = sp.value(r, it, i != bad)
			case 1:
				vals[i] = sp.value(r, it, i < nIn)
			default:
				m := sp.minor()
				vals[i] = sp.dmin + m*float64(r.between(1, int(math.Round((sp.dmax-sp.dmin)/m))-1))
			}
		}
		if mode == 1 { // shuffle which gauges are the in-range ones
			p := orderN(r, n)
			nv := make([]float64, n)
			ni := make([]int, n)
			for i := range p {
				nv[i], ni[i] = vals[p[i]], idx[p[i]]
			}
			vals, idx = nv, ni
			nIn = 0
			for i := range vals {
				if sp.inRange(sp.items[idx[i]], vals[i]) {
					nIn++
				}
			}
		}
		s := frame(520, 212, sp.title+" — four dial readings", sp.title)
		for i := 0; i < n; i++ {
			cx := 65 + 130*float64(i)
			drawGauge(s, sp, cx, 112, vals[i], mode != 2)
			s.text(cx, 198, 8.5, `text-anchor="middle" font-weight="700"`, sp.items[idx[i]].name)
			markAt(s, cx-46, 52, i)
		}
		opt := func(i int) string { return fmt.Sprintf("Gauge %s (%s)", markLetters[i], sp.items[idx[i]].name) }
		q := Question{key: "ex/" + kind, Exhibit: s.done()}
		rule := ""
		if r.chance(0.6) {
			rule = " " + sp.rule
		}
		switch mode {
		case 0:
			var wrongs []string
			for i := 0; i < n; i++ {
				if i != bad {
					wrongs = append(wrongs, opt(i))
				}
			}
			q.Text = fmt.Sprintf("Refer to %s.%s %s", ex, rule, sp.askOut)
			q.Choices, q.Correct = withChoices(r, opt(bad), wrongs)
		case 1:
			q.Text = fmt.Sprintf("Refer to %s.%s %s", ex, rule, sp.askCount)
			q.Choices, q.Correct = numChoices(r, nIn, []int{-1, 1, 2, -2, 3}, itoa)
		default:
			g := r.intn(n)
			m := sp.minor()
			k := int(math.Round((vals[g] - sp.dmin) / m))
			mj := int(math.Round(sp.major / m))
			q.Text = fmt.Sprintf("Refer to %s. What does gauge %s (%s) read?", ex, markLetters[g], sp.items[idx[g]].name)
			q.Choices, q.Correct = numChoices(r, k, []int{-1, 1, -2, 2, mj, -mj, mj / 2}, func(k int) string {
				return fmtNum(sp.dmin+float64(k)*m, sp.dec) + sp.unit
			})
		}
		return q
	}
}

// minor is the dial's minor tick: the step, coarsened to at most 50 ticks.
func (sp gaugeSpec) minor() float64 {
	m := sp.step
	for (sp.dmax-sp.dmin)/m > 50 {
		m *= 2
	}
	return m
}

func gaugeAngle(sp gaugeSpec, v float64) float64 {
	return (210 - 240*(v-sp.dmin)/(sp.dmax-sp.dmin)) * math.Pi / 180
}

func drawGauge(s *svgB, sp gaugeSpec, cx, cy, v float64, readout bool) {
	const R = 50.0
	s.circle(cx, cy, R+6, "#f8fafc", "#475569")
	minor := sp.minor()
	for t := sp.dmin; t <= sp.dmax+1e-9; t += minor {
		a := gaugeAngle(sp, t)
		s.line(cx+(R-4)*math.Cos(a), cy-(R-4)*math.Sin(a), cx+R*math.Cos(a), cy-R*math.Sin(a), "#94a3b8", 0.8)
	}
	for t := sp.dmin; t <= sp.dmax+1e-9; t += sp.major {
		a := gaugeAngle(sp, t)
		s.line(cx+(R-9)*math.Cos(a), cy-(R-9)*math.Sin(a), cx+R*math.Cos(a), cy-R*math.Sin(a), "#1f2937", 1.6)
		d := 0
		if sp.major != math.Trunc(sp.major) {
			d = 1
			if sp.major*10 != math.Trunc(sp.major*10) {
				d = 2
			}
		}
		s.text(cx+(R-18)*math.Cos(a), cy-(R-18)*math.Sin(a)+3, 7, `text-anchor="middle" fill="#334155"`, fmtNum(t, d))
	}
	a := gaugeAngle(sp, v)
	s.line(cx, cy, cx+(R-7)*math.Cos(a), cy-(R-7)*math.Sin(a), "#dc2626", 2.2)
	s.circle(cx, cy, 4, "#1f2937", "")
	s.text(cx, cy+20, 7.5, `text-anchor="middle" fill="#475569"`, sp.unit)
	if readout {
		s.rect(cx-38, cy+28, 76, 18, 3, "#0f172a", "")
		s.text(cx, cy+41, 9.5, `text-anchor="middle" font-weight="700" fill="#4ade80"`, fmtNum(v, sp.dec)+sp.unit)
	}
}
