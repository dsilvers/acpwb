package compliance

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// "Refer to Exhibit 4-B" questions. Each exhibit is drawn (SVG) or laid out
// (HTML table) from the same seed as its answer key, and the right answer is
// computed from what the exhibit shows: the line item over the meal limit,
// the exit nearest the "You are here" star, the number of items left out on
// a desk. Crawlers get a picture and a question that only makes sense
// together.

type exhibitGen func(r *rnd, c *Course, sl slots, ex string) Question

var exhibitGens = map[string]exhibitGen{
	"expense":    expenseExhibit,
	"fridge":     fridgeExhibit,
	"evac":       evacExhibit,
	"visitorlog": visitorLogExhibit,
	"gifts":      giftRegisterExhibit,
	"retention":  retentionExhibit,
	"document":   documentExhibit,
	"desk":       deskExhibit,
	"payrange":   payRangeExhibit,
	"calendar":   calendarExhibit,
	"printqueue": printQueueExhibit,
	"screening":  screeningExhibit,
	"email":      emailExhibit,
}

// exhibitNames is exhibitGens' keys, sorted (the "exhibit" universal kind
// picks from it; map order would break determinism).
var exhibitNames = func() []string {
	var out []string
	for k := range exhibitGens {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}()

// Course mixes may name "exhibit" (any kind) or "ex-<kind>" (one kind).
func init() {
	generatedKinds["exhibit"] = true
	for _, k := range exhibitNames {
		generatedKinds["ex-"+k] = true
	}
}

// categoryExhibits: bank categories that sometimes get an exhibit question
// instead of a bank question.
var categoryExhibits = map[string][]string{
	"travel":      {"expense"},
	"procurement": {"expense", "screening"},
	"export":      {"screening"},
	"kitchen":     {"fridge"},
	"safety":      {"evac"},
	"continuity":  {"evac"},
	"visitor":     {"visitorlog"},
	"badge":       {"visitorlog"},
	"gifts":       {"gifts"},
	"ethics":      {"gifts"},
	"records":     {"retention", "document"},
	"cleandesk":   {"desk", "document"},
	"security":    {"desk", "email"},
	"phishing":    {"email"},
	"email":       {"email"},
	"privacy":     {"document"},
	"pay":         {"payrange"},
	"meetings":    {"calendar"},
	"printing":    {"printqueue"},
	"remote":      {"desk"},
	"cpr":         {"evac"},
	"firstaid":    {"evac"},
}

const exhibitChance = 0.6

// drawnExhibits lists the exhibit kinds that draw a picture (SVG) rather
// than a table; "any exhibit" questions prefer them. Classified once from a
// fixed course and seed, so the list (and every question) is the same in
// every process.
var drawnExhibits = sync.OnceValue(func() []string {
	c, _ := ParseCourse("sec", courseTypes[0].Slug+"-10000")
	var out []string
	for _, k := range exhibitNames {
		r := newRnd("drawn_" + k)
		if strings.Contains(exhibitGens[k](r, c, newSlots(r, c), "Exhibit 1-A").Exhibit, "<svg") {
			out = append(out, k)
		}
	}
	return out
})

// anyExhibit picks an exhibit kind, a drawing 80% of the time.
func anyExhibit(r *rnd) string {
	if d := drawnExhibits(); len(d) > 0 && r.chance(0.8) {
		return pick(r, d)
	}
	return pick(r, exhibitNames)
}

func exhibitQ(r *rnd, c *Course, sl slots, n int, kind string) Question {
	ex := exhibitLabel(r, n)
	q := exhibitGens[kind](r, c, sl, ex)
	q.ExLabel = ex
	if q.Prompt == "" {
		q.Prompt = pick(r, []string{"Refer to the exhibit.", "Review the exhibit, then select the BEST answer.", "Examine the exhibit carefully.", "Use the exhibit to answer."})
	}
	return q
}

// ── helpers ──────────────────────────────────────────────────────────────

func personName(r *rnd) string { return pick(r, firstNames) + " " + pick(r, lastNames) }

func firstOf(name string) string { f, _, _ := strings.Cut(name, " "); return f }

func money(cents int) string {
	neg := ""
	if cents < 0 {
		neg, cents = "-", -cents
	}
	return fmt.Sprintf("%s$%s.%02d", neg, commas(cents/100), cents%100)
}

func dollars(n int) string { return "$" + commas(n) }

// numChoices builds a numeric question: right plus three distinct wrong
// values (right+delta, skipping negatives), formatted.
func numChoices(r *rnd, right int, deltas []int, f func(int) string) ([]string, int) {
	deltas = append([]int(nil), deltas...)
	shuffle(r, deltas)
	seen := map[int]bool{right: true}
	var wrongs []string
	for _, d := range deltas {
		v := right + d
		if v < 0 || seen[v] {
			continue
		}
		seen[v] = true
		wrongs = append(wrongs, f(v))
		if len(wrongs) == 3 {
			break
		}
	}
	for k := 1; len(wrongs) < 3; k++ {
		if v := right + k*7; !seen[v] {
			seen[v] = true
			wrongs = append(wrongs, f(v))
		}
	}
	return withChoices(r, f(right), wrongs)
}

func itoa(n int) string { return fmt.Sprint(n) }

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

// docExhibit lays out a form or log as an HTML table.
func docExhibit(title, meta string, head []string, rows [][]string, note string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="cmp-ex-doc"><div class="cmp-ex-head"><b>%s</b><span>%s</span></div><div class="cmp-table-wrap"><table class="cmp-ex-table"><thead><tr>`, esc(title), esc(meta))
	for _, h := range head {
		fmt.Fprintf(&b, `<th>%s</th>`, esc(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		b.WriteString(`<tr>`)
		for _, cell := range row {
			cls := ""
			if strings.HasPrefix(cell, "\x00") { // shaded cell (calendar bookings)
				cell, cls = cell[1:], ` class="cmp-ex-busy"`
			}
			fmt.Fprintf(&b, `<td%s>%s</td>`, cls, esc(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	if note != "" {
		fmt.Fprintf(&b, `<div class="cmp-ex-note">%s</div>`, esc(note))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func exDate(r *rnd) time.Time {
	return time.Date(r.between(2016, 2026), time.Month(r.between(1, 12)), r.between(1, 24), 0, 0, 0, 0, time.UTC)
}

func clock(min int) string {
	h, m := min/60, min%60
	ap := "a.m."
	if h >= 12 {
		ap = "p.m."
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	return fmt.Sprintf("%d:%02d %s", h, m, ap)
}

// ── expense report ───────────────────────────────────────────────────────

func expenseExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	meal := pick(r, []int{50, 60, 75}) * 100
	receipt := pick(r, []int{25, 50}) * 100
	lodging := pick(r, []int{189, 219, 249}) * 100
	type line struct {
		cat  string
		amt  int
		rcpt bool
	}
	good := func() line {
		switch r.intn(5) {
		case 0:
			a := r.between(1200, meal-300)
			return line{"Meal", a, a > receipt || r.chance(0.5)}
		case 1:
			return line{"Lodging", r.between(lodging-6000, lodging-500), true}
		case 2:
			return line{"Taxi / rideshare", r.between(1400, 6200), true}
		case 3:
			return line{"Parking", r.between(900, 4800), true}
		}
		return line{"Conference registration", r.between(19500, 89500), true}
	}
	var bad line
	kind := r.intn(4)
	switch kind {
	case 0:
		bad = line{"Meal", meal + r.between(500, 4000), true}
	case 1:
		bad = line{"Lodging", lodging + r.between(1500, 9000), true}
	case 2:
		bad = line{pick(r, []string{"Taxi / rideshare", "Parking", "Meal"}), receipt + r.between(300, 2400), false}
		if bad.cat == "Meal" && bad.amt > meal {
			bad.amt = meal - 100
		}
	default:
		bad = line{pick(r, []string{"Hotel minibar", "In-room movie", "Spa service", "Airport lounge pass"}), r.between(1200, 6500), true}
	}
	lines := []line{good(), good(), good(), good()}
	at := r.intn(5)
	lines = append(lines[:at], append([]line{bad}, lines[at:]...)...)
	start := exDate(r)
	var rows [][]string
	okTotal, all := 0, 0
	for i, l := range lines {
		rc := "Yes"
		if !l.rcpt {
			rc = "No"
		}
		rows = append(rows, []string{itoa(i + 1), start.AddDate(0, 0, i).Format("01/02/2006"), l.cat, money(l.amt), rc})
		all += l.amt
		if i != at {
			okTotal += l.amt
		}
	}
	who := sl["NAME"]
	q := Question{key: "ex/expense", Exhibit: docExhibit("Expense Report — "+who, fmt.Sprintf("Trip: %s · Cost center %s", sl["CITY"], sl["FORM"]),
		[]string{"Line", "Date", "Category", "Amount", "Itemized receipt"}, rows,
		fmt.Sprintf("Policy limits: meals %s per day; lodging %s per night; itemized receipt required over %s; personal items are not reimbursable. Non-compliant lines are not reimbursed.", money(meal), money(lodging), money(receipt)))}
	if r.chance(0.6) {
		q.Text = fmt.Sprintf("Refer to %s. Which line item on %s's expense report does NOT comply with the Travel & Expense Policy?", ex, firstOf(who))
		desc := func(i int) string { return fmt.Sprintf("Line %d: %s, %s", i+1, lines[i].cat, money(lines[i].amt)) }
		var wrongs []string
		for _, i := range sample(r, []int{0, 1, 2, 3, 4}, 5) {
			if i != at && len(wrongs) < 3 {
				wrongs = append(wrongs, desc(i))
			}
		}
		q.Choices, q.Correct = withChoices(r, desc(at), wrongs)
		return q
	}
	if r.chance(0.35) {
		if r.chance(0.5) {
			q.Text = fmt.Sprintf("Refer to %s. What is the total of every line on %s's expense report, before any adjustments?", ex, firstOf(who))
			q.Choices, q.Correct = numChoices(r, all, []int{-lines[at].amt, -lines[(at+2)%5].amt, 1000, -1000, 100}, money)
			return q
		}
		n := 0
		for _, l := range lines {
			if l.rcpt {
				n++
			}
		}
		q.Text = fmt.Sprintf("Refer to %s. How many lines on the report have an itemized receipt?", ex)
		q.Choices, q.Correct = numChoices(r, n, []int{-2, -1, 1, 2}, itoa)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What is the total reimbursable amount of %s's expense report?", ex, firstOf(who))
	q.Choices, q.Correct = numChoices(r, okTotal, []int{all - okTotal, -lines[(at+1)%5].amt, 1000, -1000, 10000}, money)
	return q
}

// ── shared refrigerator label ────────────────────────────────────────────

func fridgeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	stored := exDate(r)
	days := pick(r, []int{2, 3, 4, 5, 7})
	q := Question{key: "ex/fridge", Exhibit: fridgeSVG(sl["NAME"], sl["DEPT"], stored, days, r.float()*6-3)}
	last := stored.AddDate(0, 0, days)
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Under the policy shown, what is the last day this item may remain in the refrigerator?", ex)
		f := func(d time.Time) string { return d.Format("Monday, January 2") }
		q.Choices, q.Correct = withChoices(r, f(last), []string{f(last.AddDate(0, 0, -1)), f(last.AddDate(0, 0, 1)), f(stored.AddDate(0, 0, 2*days+1))})
		return q
	}
	on := stored.AddDate(0, 0, r.between(1, 2*days))
	q.Text = fmt.Sprintf("Refer to %s. Facilities finds this item during its refrigerator sweep on %s. Under the policy shown, it should be:", ex, on.Format("Monday, January 2"))
	right, wrong := "Discarded, including the container", "Left in place; it is within the storage period"
	if !on.After(last) {
		right, wrong = wrong, right
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Relabelled with today's date and kept", "Moved to the " + sl["PLACE"]})
	return q
}

// ── evacuation floor plan ────────────────────────────────────────────────

func evacExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	p := buildFloorPlan(r, sl)
	q := Question{key: "ex/evac", Exhibit: floorPlanSVG(p)}
	if r.chance(0.3) {
		q.Text = fmt.Sprintf("Refer to %s, the evacuation map for the %s floor. How many fire extinguishers are shown?", ex, p.floor)
		q.Choices, q.Correct = numChoices(r, len(p.extinguishers), []int{-2, -1, 1, 2, 3}, itoa)
		return q
	}
	if p.nearestAED >= 0 && (mixHas(c.Type.Mix, "cpr") || mixHas(c.Type.Mix, "firstaid") || r.chance(0.2)) && r.chance(0.5) {
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %s. A colleague collapses at the location marked \"You are here.\" Which AED is closest?",
			"Refer to %s. You need to send someone for an AED from the location marked \"You are here.\" Which one is nearest?",
		}), ex)
		var wrongs []string
		for i := range p.aeds {
			if i != p.nearestAED {
				wrongs = append(wrongs, fmt.Sprintf("AED %d", i+1))
			}
		}
		wrongs = append(wrongs, "The AED at the Security Office in Suite 4040")
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("AED %d", p.nearestAED+1), wrongs)
		return q
	}
	if p.blocked >= 0 && r.chance(0.25) {
		q.Text = fmt.Sprintf("Refer to %s. Which exit on the %s floor is currently unavailable?", ex, p.floor)
		var wrongs []string
		for i, e := range p.exits {
			if i != p.blocked {
				wrongs = append(wrongs, "Exit "+e.label)
			}
		}
		q.Choices, q.Correct = withChoices(r, "Exit "+p.exits[p.blocked].label, wrongs)
		return q
	}
	if p.blocked >= 0 {
		q.Text = fmt.Sprintf("Refer to %s. The fire alarm sounds while you are at the location marked \"You are here.\" Which exit should you use?", ex)
	} else {
		q.Text = fmt.Sprintf("Refer to %s. The fire alarm sounds while you are at the location marked \"You are here.\" Which is the nearest exit?", ex)
	}
	var wrongs []string
	for i, e := range p.exits {
		if i != p.answer {
			wrongs = append(wrongs, "Exit "+e.label)
		}
	}
	q.Choices, q.Correct = withChoices(r, "Exit "+p.exits[p.answer].label, wrongs)
	return q
}

// ── visitor log ──────────────────────────────────────────────────────────

func visitorLogExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type visit struct {
		name, company, host, escort string
		in, out                     int // minutes after midnight; out 0 = still on site
	}
	day := exDate(r)
	bad := r.intn(5)
	var vs []visit
	for i := 0; i < 5; i++ {
		v := visit{name: personName(r), company: pick(r, append(append([]string{}, vars["VENDOR"]...), orgs...)), host: personName(r)}
		v.escort = firstOf(v.host) + " " + string([]rune(strings.Fields(v.host)[1])[:1]) + "."
		if r.chance(0.3) {
			n := personName(r)
			v.escort = firstOf(n) + " " + string([]rune(strings.Fields(n)[1])[:1]) + "."
		}
		v.in = 480 + 5*r.between(0, 96)
		if r.chance(0.7) {
			v.out = v.in + 5*r.between(6, 48)
		}
		vs = append(vs, v)
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i].in < vs[j].in })
	variant := r.chance(0.5)
	if variant {
		vs[bad].escort = "(none)"
	}
	t := 60 * r.between(11, 15)
	for i := range vs {
		if vs[i].in == t {
			vs[i].in -= 5
		}
		if vs[i].out == t {
			vs[i].out += 5
		}
	}
	var rows [][]string
	present := 0
	for _, v := range vs {
		out := "—"
		if v.out > 0 {
			out = clock(v.out)
		}
		rows = append(rows, []string{fmt.Sprintf("V-%04d", r.between(1000, 9999)), v.name, v.company, v.host, v.escort, clock(v.in), out})
		if v.in < t && (v.out == 0 || v.out > t) {
			present++
		}
	}
	q := Question{key: "ex/visitorlog", Exhibit: docExhibit("Visitor Sign-In Log — North Lobby", day.Format("Monday, January 2, 2006"),
		[]string{"Badge", "Visitor", "Company", "Host", "Escort", "Time in", "Time out"}, rows,
		"Every visitor must be escorted by an ACPWB associate at all times. Visitor badges are returned at sign-out.")}
	if variant {
		q.Text = fmt.Sprintf("Refer to %s. Which visitor's log entry does NOT comply with the Visitor Escort Standard?", ex)
		var wrongs []string
		for i, v := range vs {
			if i != bad && len(wrongs) < 3 {
				wrongs = append(wrongs, v.name)
			}
		}
		q.Choices, q.Correct = withChoices(r, vs[bad].name, wrongs)
		return q
	}
	if r.chance(0.5) {
		i := r.intn(len(vs))
		if r.chance(0.5) {
			q.Text = fmt.Sprintf("Refer to %s. Who hosted %s's visit?", ex, vs[i].name)
			var wrongs []string
			for j, v := range vs {
				if j != i && v.host != vs[i].host && !contains(wrongs, v.host) && len(wrongs) < 3 {
					wrongs = append(wrongs, v.host)
				}
			}
			for len(wrongs) < 3 {
				wrongs = append(wrongs, personName(r))
			}
			q.Choices, q.Correct = withChoices(r, vs[i].host, wrongs)
			return q
		}
		for j := range vs {
			if vs[(i+j)%len(vs)].out > 0 {
				i = (i + j) % len(vs)
				break
			}
		}
		if vs[i].out > 0 {
			dur := func(m int) string { return fmt.Sprintf("%d h %02d min", m/60, m%60) }
			q.Text = fmt.Sprintf("Refer to %s. According to the log, how long was %s on site?", ex, vs[i].name)
			q.Choices, q.Correct = numChoices(r, vs[i].out-vs[i].in, []int{-60, 60, -15, 15, 30, -30}, dur)
			return q
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. According to the log, how many visitors were on site at %s?", ex, clock(t))
	q.Choices, q.Correct = numChoices(r, present, []int{-2, -1, 1, 2, 3}, itoa)
	return q
}

// ── gift register ────────────────────────────────────────────────────────

var giftItems = []string{"Fruit basket", "Box of chocolates", "Two Brewers tickets", "Two Bucks tickets", "Bottle of wine",
	"Holiday cookie tin", "Logo fleece jacket", "Golf outing", "Dinner for two", "Theater tickets", "Branded water bottle",
	"Desk calendar", "Cheese & sausage gift box", "Conference pass", "Bluetooth speaker", "Coffee sampler"}

func giftRegisterExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{50, 75, 100})
	names := []string{personName(r), personName(r), personName(r)}
	approver := personName(r)
	start := exDate(r)
	bad := r.intn(6)
	type gift struct {
		who, from, item string
		val             int
		ok              bool
	}
	var gs []gift
	for i := 0; i < 6; i++ {
		g := gift{who: pick(r, names), from: pick(r, vars["VENDOR"]), item: pick(r, giftItems)}
		if r.chance(0.5) {
			g.val, g.ok = r.between(10, limit), r.chance(0.2)
		} else {
			g.val, g.ok = r.between(limit+5, limit*5), true
		}
		gs = append(gs, g)
	}
	gs[bad].val, gs[bad].ok = r.between(limit+5, limit*4), false
	var rows [][]string
	totals := map[string]int{}
	for i, g := range gs {
		appr := "—"
		if g.ok {
			appr = approver
		}
		rows = append(rows, []string{itoa(i + 1), start.AddDate(0, 0, 3*i).Format("01/02/2006"), g.who, g.from, g.item, dollars(g.val), appr})
		totals[g.who] += g.val
	}
	q := Question{key: "ex/gifts", Exhibit: docExhibit("Gifts & Entertainment Register", c.Dept,
		[]string{"Entry", "Date", "Recipient", "Given by", "Description", "Est. value", "Pre-approved by"}, rows,
		fmt.Sprintf("Gifts with an estimated value over %s require pre-approval from Compliance before acceptance.", dollars(limit)))}
	if r.chance(0.55) {
		q.Text = fmt.Sprintf("Refer to %s. Which entry required pre-approval but did not receive it?", ex)
		desc := func(i int) string { return fmt.Sprintf("Entry %d: %s from %s", i+1, gs[i].item, gs[i].from) }
		var wrongs []string
		for _, i := range sample(r, []int{0, 1, 2, 3, 4, 5}, 6) {
			if i != bad && len(wrongs) < 3 {
				wrongs = append(wrongs, desc(i))
			}
		}
		q.Choices, q.Correct = withChoices(r, desc(bad), wrongs)
		return q
	}
	if r.chance(0.4) {
		over := 0
		for _, g := range gs {
			if g.val > limit {
				over++
			}
		}
		q.Text = fmt.Sprintf("Refer to %s. How many entries in the register required pre-approval?", ex)
		q.Choices, q.Correct = numChoices(r, over, []int{-2, -1, 1, 2}, itoa)
		return q
	}
	who := gs[r.intn(6)].who
	q.Text = fmt.Sprintf("Refer to %s. What is the total estimated value of the gifts %s received?", ex, who)
	q.Choices, q.Correct = numChoices(r, totals[who], []int{-limit, limit, 25, -25, 10, -10}, dollars)
	return q
}

// ── retention schedule ───────────────────────────────────────────────────

var retentionTypes = []string{"Payroll registers", "Expense reports", "Visitor logs", "Badge access logs", "Training records",
	"Vendor contracts", "Personnel files", "Meeting room bookings", "Gift register entries", "Incident reports",
	"Shared refrigerator logs", "Board minutes (drafts)", "Pay equity analyses", "Time-off requests", "Clean desk audit sheets",
	"Certificates of completion", "Phishing simulation results", "Parking permits"}

func retentionExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	types := sample(r, retentionTypes, 6)
	years := make([]int, len(types))
	var rows [][]string
	for i, t := range types {
		years[i] = r.between(1, 12)
		rows = append(rows, []string{t, plural(years[i], "year"), pick(r, []string{"Records Management", "Legal", "People Operations", "Finance", "Facilities", c.Dept})})
	}
	q := Question{key: "ex/retention", Exhibit: docExhibit("Records Retention Schedule (excerpt)", "Schedule "+sl["POLICY"]+" · rev. "+sl["YEAR"],
		[]string{"Record type", "Retention", "Owner"}, rows,
		"Retention runs from the end of the calendar year in which the record was created. Records may be destroyed on or after January 1 following the end of the retention period, unless a legal hold applies.")}
	if r.chance(0.35) {
		best := 0
		for i := range years {
			if years[i] > years[best] {
				best = i
			}
		}
		years[best] += 3 // unique maximum
		rows[best][1] = plural(years[best], "year")
		q.Exhibit = docExhibit("Records Retention Schedule (excerpt)", "Schedule "+sl["POLICY"]+" · rev. "+sl["YEAR"], []string{"Record type", "Retention", "Owner"}, rows,
			"Retention runs from the end of the calendar year in which the record was created. Records may be destroyed on or after January 1 following the end of the retention period, unless a legal hold applies.")
		q.Text = fmt.Sprintf("Refer to %s. Which record type must be kept the longest?", ex)
		var wrongs []string
		for i, t := range types {
			if i != best && len(wrongs) < 3 {
				wrongs = append(wrongs, t)
			}
		}
		q.Choices, q.Correct = withChoices(r, types[best], wrongs)
		return q
	}
	i := r.intn(len(types))
	created := exDate(r)
	q.Text = fmt.Sprintf("Refer to %s. In what year may %s created on %s first be destroyed?", ex, lowerFirst(types[i]), created.Format("January 2, 2006"))
	q.Choices, q.Correct = numChoices(r, created.Year()+years[i]+1, []int{-1, 1, -2, 2}, itoa)
	return q
}

// ── classified document (report cover from the image pool) ──────────────

var classLevels = []struct{ Label, Share, Store, Print, Dispose string }{
	{"PUBLIC", "Share it; Public documents are approved for release", "Anywhere, including the public ACPWB website",
		"On any printer", "In regular paper recycling"},
	{"INTERNAL", "Decline; Internal documents are for ACPWB associates only", "On ACPWB systems such as the intranet or team sites",
		"On any ACPWB printer, collected promptly", "In an ACPWB recycling bin, never at home"},
	{"CONFIDENTIAL", "Share only with owner approval, under an NDA, through the secure-transfer tool", "Only in access-controlled ACPWB repositories, shared with named people",
		"Only with hold & release, collected immediately", "In a locked confidential shred bin"},
	{"RESTRICTED", "Decline; only the document owner may share it, with Legal approval", "Only in the designated restricted repository, with owner-approved access",
		"Only on the designated secure printer, with the owner's approval", "By witnessed cross-cut shredding, with a destruction certificate"},
}

func coverTitle(slug string) string {
	parts := strings.Split(slug, "-")
	if n := len(parts); n > 1 && len(parts[n-1]) == 4 && strings.Trim(parts[n-1], "0123456789abcdef") == "" && strings.Trim(parts[n-1], "0123456789") != "" {
		parts = parts[:n-1]
	}
	for i, p := range parts {
		switch p {
		case "hr", "ceo", "ai", "it", "dei", "esg", "q1", "q2", "q3", "q4":
			parts[i] = strings.ToUpper(p)
		case "and", "of", "the", "for", "in", "to", "vs":
			if i > 0 {
				continue
			}
			fallthrough
		default:
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func documentExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cover := pick(r, reportCovers)
	lvl := r.intn(len(classLevels))
	L := classLevels[lvl]
	title := coverTitle(cover)
	q := Question{key: "ex/document",
		Exhibit: fmt.Sprintf(`<div class="cmp-ex-cover"><img src="/static/img/report-covers/%s.webp" alt="" loading="lazy"><span class="cmp-ex-stamp cmp-ex-stamp-%d">%s</span><div class="cmp-ex-covercap"><b>%s</b><small>Owner: %s · %s · Classification: %s</small></div></div>`,
			esc(cover), lvl, L.Label, esc(title), esc(sl["DEPT"]), esc(sl["YEAR"]), L.Label)}
	var right string
	var wrongs []string
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. %s from %s asks you for a copy of this document. Based on its classification, you should:", ex, sl["NAME2"], sl["VENDOR"])
		for i, l := range classLevels {
			if i == lvl {
				right = l.Share
			} else {
				wrongs = append(wrongs, l.Share)
			}
		}
	} else {
		field := r.intn(3)
		q.Text = fmt.Sprintf([]string{
			"Refer to %s. Based on its classification, where may this document be stored?",
			"Refer to %s. You need a paper copy of this document for a meeting. Based on its classification, how may it be printed?",
			"Refer to %s. This document has reached the end of its retention period. Based on its classification, how should paper copies be disposed of?",
		}[field], ex)
		for i, l := range classLevels {
			v := []string{l.Store, l.Print, l.Dispose}[field]
			if i == lvl {
				right = v
			} else {
				wrongs = append(wrongs, v)
			}
		}
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── clean desk scene ─────────────────────────────────────────────────────

func deskExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	d := buildDesk(r, sl)
	q := Question{key: "ex/desk", Exhibit: deskSVG(d)}
	viol := d.violations()
	if len(viol) >= 1 && len(d.benign) >= 3 && r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s, a photo of %s's desk taken during an after-hours clean desk audit. Which item should have been secured?", ex, sl["FIRST"])
		var wrongs []string
		for _, b := range sample(r, d.benign, 3) {
			wrongs = append(wrongs, deskItemNames[b])
		}
		q.Choices, q.Correct = withChoices(r, deskItemNames[pick(r, viol)], wrongs)
		return q
	}
	if len(viol) >= 3 && len(d.benign) >= 1 && r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s, a photo of %s's desk after hours. Which item may be left on the desk overnight?", ex, sl["FIRST"])
		var wrongs []string
		for _, v := range sample(r, viol, 3) {
			wrongs = append(wrongs, deskItemNames[v])
		}
		q.Choices, q.Correct = withChoices(r, deskItemNames[pick(r, d.benign)], wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s, a photo of %s's desk taken during an after-hours clean desk audit. How many Clean Desk Policy violations are visible?", ex, sl["FIRST"])
	q.Choices, q.Correct = numChoices(r, len(viol), []int{-2, -1, 1, 2, 3}, itoa)
	return q
}

// ── pay range ────────────────────────────────────────────────────────────

func payRangeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	lo := r.between(42, 120) * 1000
	hi := lo + lo/2/1000*1000
	mid := (lo + hi) / 2
	type emp struct {
		name string
		pay  int
	}
	bad := r.intn(4)
	var es []emp
	for i := 0; i < 4; i++ {
		pay := lo + 500*r.between(4, (hi-lo)/500-4)
		if i == bad {
			if r.chance(0.5) {
				pay = lo - 500*r.between(4, 16)
			} else {
				pay = hi + 500*r.between(4, 16)
			}
		}
		es = append(es, emp{personName(r), pay})
	}
	job := sl["TITLE"]
	var rows [][]string
	for _, e := range es {
		rows = append(rows, []string{e.name, job, dollars(e.pay)})
	}
	q := Question{key: "ex/payrange", Exhibit: payRangeSVG(job, lo, mid, hi) + docExhibit("Current Salaries", sl["DEPT"], []string{"Associate", "Role", "Base salary"}, rows, "")}
	if r.chance(0.55) {
		q.Text = fmt.Sprintf("Refer to %s. Whose base salary falls outside the posted pay range?", ex)
		var wrongs []string
		for i, e := range es {
			if i != bad {
				wrongs = append(wrongs, e.name)
			}
		}
		q.Choices, q.Correct = withChoices(r, es[bad].name, wrongs)
		return q
	}
	// Someone below the midpoint (the out-of-range one if needed).
	k := -1
	for i, e := range es {
		if e.pay < mid && (k < 0 || i != bad) {
			k = i
		}
	}
	if k < 0 {
		es[0].pay = mid - 500*r.between(2, 20)
		k = 0
		rows[0][2] = dollars(es[0].pay)
		q.Exhibit = payRangeSVG(job, lo, mid, hi) + docExhibit("Current Salaries", sl["DEPT"], []string{"Associate", "Role", "Base salary"}, rows, "")
	}
	q.Text = fmt.Sprintf("Refer to %s. How much would %s's salary need to increase to reach the range midpoint?", ex, es[k].name)
	q.Choices, q.Correct = numChoices(r, mid-es[k].pay, []int{-1000, 1000, 2500, -2500, 500, (hi - lo) / 2}, dollars)
	return q
}

// ── meeting room calendar ────────────────────────────────────────────────

var roomNames = []string{"Lake Michigan", "Third Ward", "Riverwalk", "Menomonee", "Bay View", "Brady Street", "Juneau", "Kilbourn", "Mitchell", "Sherman Park", "Walker's Point", "Tosa"}
var bookingTitles = []string{"Budget sync", "1:1", "Vendor call", "Staff meeting", "Interview", "Training", "PERCH planning", "Offsite prep", "Quarterly review", "Hold", "Audit walkthrough", "Lunch & learn", "Board prep", "Fire warden huddle"}

func calendarExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	rooms := sample(r, roomNames, 4)
	const first, slotsN = 9, 8 // 9:00 a.m. … 4:00 p.m.
	start := r.between(0, slotsN-2)
	target := r.intn(4)
	busy := make([][]bool, 4)
	for i := range busy {
		busy[i] = make([]bool, slotsN)
		for s := range busy[i] {
			busy[i][s] = r.chance(0.45)
		}
		if i == target {
			busy[i][start], busy[i][start+1] = false, false
		} else if !busy[i][start] && !busy[i][start+1] {
			busy[i][start+r.intn(2)] = true
		}
	}
	head := []string{"Room"}
	for s := 0; s < slotsN; s++ {
		head = append(head, strings.Replace(clock((first+s)*60), ":00", "", 1))
	}
	var rows [][]string
	for i, room := range rooms {
		row := []string{room}
		for s := 0; s < slotsN; s++ {
			if busy[i][s] {
				row = append(row, "\x00"+pick(r, bookingTitles))
			} else {
				row = append(row, "")
			}
		}
		rows = append(rows, row)
	}
	day := exDate(r)
	q := Question{key: "ex/calendar", Exhibit: docExhibit("Conference Room Availability — "+sl["FLOOR"]+" Floor", day.Format("Monday, January 2, 2006"), head, rows, "Shaded cells are booked. Each column is one hour.")}
	if r.chance(0.35) {
		s := r.intn(slotsN)
		free := 0
		for i := range rooms {
			if !busy[i][s] {
				free++
			}
		}
		q.Text = fmt.Sprintf("Refer to %s. How many of the rooms shown are free at %s?", ex, clock((first+s)*60))
		q.Choices, q.Correct = numChoices(r, free, []int{-2, -1, 1, 2, 3}, itoa)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. %s needs a room from %s to %s for a two-hour meeting. Which room is available for the entire meeting?", ex, sl["NAME"], clock((first+start)*60), clock((first+start+2)*60))
	var wrongs []string
	for i, room := range rooms {
		if i != target {
			wrongs = append(wrongs, room)
		}
	}
	q.Choices, q.Correct = withChoices(r, rooms[target], wrongs)
	return q
}

// ── print queue ──────────────────────────────────────────────────────────

var printDocs = []string{"Q%d salary review.xlsx", "Board deck v%d.pptx", "Offer letter %d.docx", "Lunch menu %d.pdf", "Org chart rev %d.pdf",
	"Parking memo %d.docx", "Merger memo draft %d.docx", "PERCH agenda %d.pdf", "Vendor contract %d.pdf", "Holiday party flyer %d.pdf",
	"Performance review %d.docx", "Expense report %d.pdf", "Benefits FAQ %d.pdf", "Proxy statement draft %d.docx"}

func printQueueExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bad := r.intn(6)
	var rows [][]string
	var jobs []string
	for i := 0; i < 6; i++ {
		owner := personName(r)
		doc := fmt.Sprintf(pick(r, printDocs), r.between(2, 9))
		cls := pick(r, []string{"Public", "Internal", "Internal"})
		mode := pick(r, []string{"Direct", "Hold & release"})
		status := pick(r, []string{"Collected", "Collected", fmt.Sprintf("Not collected (%d h %02d m)", r.between(1, 9), r.between(0, 59))})
		if mode == "Hold & release" && r.chance(0.4) {
			status = "Held — awaiting release"
		}
		if r.chance(0.4) || i == bad {
			cls = pick(r, []string{"Confidential", "Restricted"})
			mode = "Hold & release"
			status = pick(r, []string{"Collected", "Held — awaiting release"})
		}
		if i == bad {
			if r.chance(0.5) {
				mode = "Direct"
			} else {
				status = fmt.Sprintf("Not collected (%d h %02d m)", r.between(1, 9), r.between(0, 59))
			}
		}
		job := fmt.Sprintf("%d", r.between(1000, 9999))
		jobs = append(jobs, fmt.Sprintf("Job %s (%s)", job, firstOf(owner)))
		rows = append(rows, []string{job, owner, doc, cls, mode, status})
	}
	q := Question{key: "ex/printqueue", Exhibit: docExhibit("Print Queue — "+sl["FLOOR"]+" Floor Printer Alcove", "Secure Print Server · "+sl["WEEKDAY"],
		[]string{"Job", "Owner", "Document", "Classification", "Mode", "Status"}, rows,
		"Secure Printing Standard: Confidential and Restricted documents must be printed with hold & release and collected immediately.")}
	q.Text = fmt.Sprintf("Refer to %s. Which print job does NOT comply with the Secure Printing Standard?", ex)
	var wrongs []string
	for _, i := range sample(r, []int{0, 1, 2, 3, 4, 5}, 6) {
		if i != bad && len(wrongs) < 3 {
			wrongs = append(wrongs, jobs[i])
		}
	}
	q.Choices, q.Correct = withChoices(r, jobs[bad], wrongs)
	return q
}

// ── restricted party screening ───────────────────────────────────────────

var tradeWords1 = []string{"Golden", "Northern", "Crescent", "Atlas", "Meridian", "Polar", "Blue River", "Red Star", "Silver Gate", "Eastwind", "Harbor", "Summit", "Ironbridge", "Jade", "Falcon", "Orion"}
var tradeWords2 = []string{"Trading", "Logistics", "Machinery", "Electronics", "Shipping", "Industrial", "Precision", "Components", "Freight", "Instruments"}
var tradeSuffix = [][2]string{{"Co.", "Company"}, {"Ltd.", "Limited"}, {"LLC", "L.L.C."}, {"Group", "Grp."}, {"Corp.", "Corporation"}}

func screeningExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	w1 := sample(r, tradeWords1, 7)
	w2 := sample(r, tradeWords2, 7)
	cities := []string{"Dubai", "Hong Kong", "Istanbul", "Tashkent", "Minsk", "Caracas", "Singapore", "Almaty", "Tbilisi", "Yerevan"}
	var list [][]string
	var listed [][2]string // [base name, suffix index]
	for i := 0; i < 3; i++ {
		si := r.intn(len(tradeSuffix))
		listed = append(listed, [2]string{w1[i] + " " + w2[i], itoa(si)})
		list = append(list, []string{w1[i] + " " + w2[i] + " " + tradeSuffix[si][0], pick(r, cities), pick(r, []string{"Entity List", "SDN", "Denied Persons", "Unverified List"})})
	}
	hit := r.intn(3)
	bad := r.intn(4)
	var orders [][]string
	var names []string
	for i := 0; i < 4; i++ {
		var name string
		if i == bad {
			si := 0
			fmt.Sscan(listed[hit][1], &si)
			base := listed[hit][0]
			switch r.intn(3) {
			case 0:
				name = base + " " + tradeSuffix[si][1]
			case 1:
				name = strings.ToUpper(base) + " " + tradeSuffix[si][0]
			default:
				name = strings.Replace(base, " ", "-", 1) + " " + tradeSuffix[si][1]
			}
		} else {
			name = w1[3+i] + " " + w2[3+i] + " " + pick(r, tradeSuffix)[r.intn(2)]
		}
		names = append(names, name)
		orders = append(orders, []string{fmt.Sprintf("SO-%05d", r.between(10000, 99999)), name, pick(r, cities), pick(r, []string{"Servo controllers", "Network switches", "Office chairs", "Calibration kits", "Label printers", "Badge readers", "Thermostats"})})
	}
	q := Question{key: "ex/screening", Exhibit: docExhibit("Restricted Party List (excerpt)", "Trade Compliance · updated "+exDate(r).Format("01/02/2006"), []string{"Listed party", "City", "List"}, list, "") +
		docExhibit("Orders Pending Shipment", sl["WEEKDAY"], []string{"Order", "Customer", "Ship to", "Goods"}, orders,
			"Hold any order whose customer matches a listed party, including variations in spelling, punctuation, or abbreviation.")}
	q.Text = fmt.Sprintf("Refer to %s. Which order must be held for Trade Compliance review?", ex)
	var wrongs []string
	for i, row := range orders {
		if i != bad {
			wrongs = append(wrongs, row[0]+" ("+names[i]+")")
		}
	}
	q.Choices, q.Correct = withChoices(r, orders[bad][0]+" ("+names[bad]+")", wrongs)
	return q
}

func mixHas(m []mixEntry, cat string) bool {
	for _, e := range m {
		if e.cat == cat {
			return true
		}
	}
	return false
}
