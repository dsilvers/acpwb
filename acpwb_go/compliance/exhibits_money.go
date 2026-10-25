package compliance

// Money exhibits: expenses, travel, pay, equity, benefits, approvals, and
// cash handling. Every answer is computed from the figures drawn; rules
// with a legal basis (per diem 75% travel days, clawback lookback, CTR
// threshold, pay transparency postings) follow the real rule.

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func init() {
	for k, g := range map[string]exhibitGen{
		"invoice":      invoiceExhibit,
		"wire":         wireExhibit,
		"pcardstmt":    pcardStatementExhibit,
		"receipt":      receiptExhibit,
		"mileage":      mileageExhibit,
		"perdiem":      perDiemExhibit,
		"boardingpass": boardingPassExhibit,
		"hotelfolio":   hotelFolioExhibit,
		"vesting":      vestingExhibit,
		"options":      optionsExhibit,
		"payratio":     payRatioExhibit,
		"summarycomp":  summaryCompExhibit,
		"sayonpay":     sayOnPayExhibit,
		"clawback":     clawbackExhibit,
		"match401k":    match401kExhibit,
		"retirevest":   retireVestExhibit,
		"hsalimit":     hsaLimitExhibit,
		"planpick":     planPickExhibit,
		"bonuscalc":    bonusCalcExhibit,
		"commission":   commissionExhibit,
		"relocation":   relocationExhibit,
		"paystub":      payStubExhibit,
		"approvals":    approvalMatrixExhibit,
		"payscatter":   payScatterExhibit,
		"jobposting":   jobPostingExhibit,
		"cashlog":      cashLogExhibit,
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
		"travel":          {"receipt", "perdiem", "boardingpass", "hotelfolio", "mileage"},
		"reimburse":       {"receipt", "mileage", "relocation"},
		"pcard":           {"pcardstmt"},
		"procurement":     {"invoice", "approvals"},
		"thirdparty":      {"wire", "invoice"},
		"authority":       {"approvals"},
		"finance":         {"invoice", "approvals"},
		"execcomp":        {"summarycomp", "payratio"},
		"payratio":        {"payratio"},
		"sayonpay":        {"sayonpay"},
		"clawback":        {"clawback"},
		"equity":          {"vesting", "options"},
		"insider":         {"options"},
		"retirement":      {"match401k", "retirevest"},
		"hsa":             {"hsalimit"},
		"enrollment":      {"planpick"},
		"incentives":      {"bonuscalc", "commission"},
		"payroll":         {"paystub"},
		"wage":            {"paystub"},
		"pay":             {"payscatter"},
		"payequity":       {"payscatter"},
		"paytransparency": {"jobposting"},
		"hiring":          {"jobposting"},
		"aml":             {"cashlog"},
		"pettycash":       {"cashlog"},
		"bribery":         {"wire"},
		"hospitality":     {"receipt"},
		"offsites":        {"hotelfolio", "receipt"},
		"travelbooking":   {"boardingpass", "hotelfolio"},
		"finwell":         {"paystub", "match401k"},
		"socialeng":       {"wire"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// ── vendor invoice ─────────────────────────────────────────────────────

var invoiceItems = []string{"Binder clips (box of 12)", "Toner cartridge, black", "Ergonomic footrest", "Conference room chairs", "Whiteboard markers (8 pk)", "Shredding service (monthly)", "Badge lanyards (100)", "Hand sanitizer refill", "Cable organizers", "Banker's boxes (25)", "Desk lamp, LED", "Copy paper (case)", "Wireless mouse", "Printed compliance posters", "Coffee filters (case)"}

func invoiceExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := r.between(3, 5)
	items := sample(r, invoiceItems, n)
	type line struct {
		desc      string
		qty, unit int // unit in cents
		ext       int
	}
	lines := make([]line, n)
	sub := 0
	for i := range lines {
		lines[i] = line{desc: items[i], qty: r.between(1, 24), unit: 100 * r.between(3, 180)}
		lines[i].ext = lines[i].qty * lines[i].unit
		sub += lines[i].ext
	}
	wrong := r.intn(n)
	mode := r.intn(2)
	if mode == 0 { // one extension is miscalculated
		lines[wrong].ext += pick(r, []int{1000, 2000, -1000, 900, 5000})
		if lines[wrong].ext <= 0 {
			lines[wrong].ext += 6000
		}
	}
	shown := 0
	for _, l := range lines {
		shown += l.ext
	}
	taxPct := pick(r, []int{5, 6, 7})
	s := paper(400, 120+float64(n)*22+80, "Vendor invoice", "INVOICE #"+fmt.Sprint(r.between(10000, 99999)), sl["VENDOR"])
	s.text(14, 54, 9, `fill="#555"`, "Bill to: ACPWB Accounts Payable · Net 30")
	y := 76.0
	s.text(14, y, 9, `font-weight="700"`, "Description")
	s.text(250, y, 9, `font-weight="700" text-anchor="end"`, "Qty")
	s.text(310, y, 9, `font-weight="700" text-anchor="end"`, "Unit")
	s.text(386, y, 9, `font-weight="700" text-anchor="end"`, "Amount")
	s.line(14, y+5, 386, y+5, "#999", 1)
	for i, l := range lines {
		ly := y + 22 + float64(i)*22
		s.text(14, ly, 10, "", fmt.Sprintf("%d. %s", i+1, l.desc))
		s.text(250, ly, 10, `text-anchor="end"`, fmt.Sprint(l.qty))
		s.text(310, ly, 10, `text-anchor="end"`, money(l.unit))
		s.text(386, ly, 10, `text-anchor="end"`, money(l.ext))
	}
	y += 22*float64(n) + 18
	s.line(230, y-8, 386, y-8, "#999", 1)
	s.text(310, y+6, 10, `text-anchor="end"`, "Subtotal")
	s.text(386, y+6, 10, `text-anchor="end"`, money(shown))
	s.text(310, y+24, 10, `text-anchor="end"`, fmt.Sprintf("Sales tax %d%%", taxPct))
	tax := shown * taxPct / 100
	s.text(386, y+24, 10, `text-anchor="end"`, money(tax))
	s.text(310, y+44, 11, `text-anchor="end" font-weight="700"`, "TOTAL DUE")
	s.text(386, y+44, 11, `text-anchor="end" font-weight="700"`, money(shown+tax))
	q := Question{key: "ex/invoice", Exhibit: s.done()}
	if mode == 0 {
		q.Text = fmt.Sprintf("Refer to %s. Before approving payment, %s checks the math. Which line's amount does not equal quantity times unit price?", ex, sl["FIRST"])
		var opts []string
		for i := range lines {
			opts = append(opts, fmt.Sprintf("Line %d", i+1))
		}
		q.Choices, q.Correct = oneOf(r, opts[wrong], append(opts, "None; every line is correct"))
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What is the invoice subtotal before tax?", ex)
	q.Choices, q.Correct = moneyQ(r, sub, []int{tax, -lines[0].ext, 1000, -1000, lines[n-1].unit})
	return q
}

// ── wire request vs. vendor master ─────────────────────────────────────

func wireExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	acct := fmt.Sprintf("%010d", r.between(1000000000, 1999999999))
	routing := fmt.Sprintf("0%08d", r.between(71000000, 75999999))
	bank := pick(r, []string{"Lakeshore National Bank", "Cream City Savings", "Great Lakes Trust", "Badger State Bank"})
	contact := personName(r)
	fields := []string{"Payee name", "Bank name", "Routing number", "Account number", "Remit-to email"}
	master := []string{sl["VENDOR"], bank, routing, acct, "ar@" + strings.ToLower(strings.ReplaceAll(strings.Fields(sl["VENDOR"])[0], "'", "")) + ".com"}
	req := append([]string(nil), master...)
	changed := pick(r, []int{1, 2, 3, 4})
	switch changed {
	case 1:
		req[1] = pick(r, []string{"Offshore Capital Bank Ltd.", "First Digital Remit", "Pacific Rim Holdings Bank"})
	case 2:
		req[2] = fmt.Sprintf("0%08d", r.between(26000000, 26999999))
	case 3:
		req[3] = fmt.Sprintf("%010d", r.between(5000000000, 5999999999))
	case 4:
		parts := strings.SplitN(master[4], "@", 2)
		req[4] = parts[0] + "@" + strings.Replace(parts[1], ".com", "-billing.net", 1)
	}
	var rows [][]string
	for i, f := range fields {
		rows = append(rows, []string{f, master[i], req[i]})
	}
	amt := 100 * r.between(8000, 240000)
	q := Question{key: "ex/wire", Exhibit: docExhibit("Wire Request Review — "+money(amt), "Requested by "+contact+" · urgent, same day", []string{"Field", "Vendor master record", "Wire request"}, rows, "Policy: any change to payment details must be verified by calling the vendor at a number already on file.")}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Which field on the wire request does not match the vendor master record?", ex)
		q.Choices, q.Correct = oneOf(r, fields[changed], fields)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What should %s do before releasing this wire?", ex, sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, "Hold it and verify the "+strings.ToLower(fields[changed])+" change by calling the vendor at the number on file",
		[]string{"Release it, since the request is marked urgent", "Reply to the requester's email to confirm the new details", "Release half now and half after confirmation"})
	return q
}

// ── P-card statement (split purchases) ─────────────────────────────────

func pcardStatementExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{2500, 3000, 5000})
	vendors := []string{"Lakefront Office Supply", "Brightline Staffing", "Cream City Catering", "MKE Print Shop", "TechDepot", "Third Ward Florist", "Lakeshore Hardware", "Badger Courier"}
	vs := sample(r, vendors, 5)
	day := exDate(r)
	type tx struct {
		d   time.Time
		v   string
		amt int
	}
	var txs []tx
	split := vs[0]
	sd := day.AddDate(0, 0, r.between(2, 20))
	a := limit*100 - 100*r.between(1, 120)
	b := limit*100 - 100*r.between(1, 300)
	txs = append(txs, tx{sd, split, a}, tx{sd, split, b})
	for i := 0; i < r.between(3, 5); i++ {
		txs = append(txs, tx{day.AddDate(0, 0, r.between(0, 27)), vs[1+r.intn(4)], 100 * r.between(12, limit/2)})
	}
	shuffle(r, txs)
	for i := 1; i < len(txs); i++ { // date order
		for j := i; j > 0 && txs[j].d.Before(txs[j-1].d); j-- {
			txs[j], txs[j-1] = txs[j-1], txs[j]
		}
	}
	var rows [][]string
	total := 0
	for _, t := range txs {
		rows = append(rows, []string{t.d.Format("01/02"), t.v, money(t.amt)})
		total += t.amt
	}
	q := Question{key: "ex/pcardstmt", Exhibit: docExhibit("P-Card Statement — "+sl["NAME"], fmt.Sprintf("Single-transaction limit %s", dollars(limit)), []string{"Date", "Merchant", "Amount"}, rows, "Splitting a purchase to stay under the single-transaction limit is prohibited.")}
	if r.chance(0.6) {
		q.Text = fmt.Sprintf("Refer to %s. Which merchant's charges look like a split purchase?", ex)
		q.Choices, q.Correct = oneOf(r, split, vs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What is the statement total %s must reconcile?", ex, sl["FIRST"])
	q.Choices, q.Correct = moneyQ(r, total, []int{-a, -b, 10000, -10000, 2500})
	return q
}

// ── restaurant receipt ─────────────────────────────────────────────────

func receiptExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	diners := r.between(1, 4)
	capPer := pick(r, []int{50, 60, 75})
	items := []string{"Fish fry", "Bratwurst plate", "Cobb salad", "Butter burger", "Cheese curds", "Walleye sandwich", "Club sandwich", "Tomato soup", "Iced tea", "Coffee", "Sparkling water", "Kringle slice"}
	var subtotal int
	var lines [][2]string
	for _, it := range sample(r, items, diners*2+r.between(0, 2)) {
		p := 100*r.between(3, 24) + pick(r, []int{0, 25, 50, 75, 95})
		subtotal += p
		lines = append(lines, [2]string{it, money(p)})
	}
	alcohol := 0
	if r.chance(0.5) {
		alcohol = 100 * r.between(8, 14) * r.between(1, 3)
		subtotal += alcohol
		lines = append(lines, [2]string{"Craft beer flight", money(alcohol)})
	}
	tax := subtotal * 55 / 1000
	tipPct := pick(r, []int{15, 18, 20, 22, 25, 30})
	tip := (subtotal + tax) * tipPct / 100
	s := newSVG(240, 130+float64(len(lines))*16+110, "Restaurant receipt")
	s.rect(10, 4, 220, 120+float64(len(lines))*16+100, 2, "#fffef8", "#ddd")
	s.text(120, 28, 12, `text-anchor="middle" font-weight="700" font-family="Courier New,monospace"`, "THE LAKEFRONT GRILLE")
	s.text(120, 42, 8, `text-anchor="middle" font-family="Courier New,monospace"`, "Milwaukee, WI · Table 14")
	s.text(120, 56, 8, `text-anchor="middle" font-family="Courier New,monospace"`, fmt.Sprintf("Guests: %d", diners))
	y := 78.0
	for _, l := range lines {
		s.text(22, y, 9, `font-family="Courier New,monospace"`, l[0])
		s.text(218, y, 9, `text-anchor="end" font-family="Courier New,monospace"`, l[1])
		y += 16
	}
	s.dash(22, y-6, 218, y-6, "#999", 1)
	for _, kv := range [][2]string{{"SUBTOTAL", money(subtotal)}, {"TAX", money(tax)}, {fmt.Sprintf("TIP (%d%%)", tipPct), money(tip)}, {"TOTAL", money(subtotal + tax + tip)}} {
		y += 16
		s.text(22, y, 9, `font-family="Courier New,monospace" font-weight="700"`, kv[0])
		s.text(218, y, 9, `text-anchor="end" font-family="Courier New,monospace" font-weight="700"`, kv[1])
	}
	s.text(120, y+30, 8, `text-anchor="middle" font-family="Courier New,monospace"`, "THANK YOU · COME AGAIN")
	q := Question{key: "ex/receipt", Exhibit: s.done()}
	maxTip := 20
	tipOK := tip
	if tipPct > maxTip {
		tipOK = (subtotal + tax) * maxTip / 100
	}
	policy := fmt.Sprintf("ACPWB policy reimburses meals up to %s per person (food, non-alcoholic drinks, and tax), plus a tip of up to %d%%; alcohol is reimbursable only with a client present.", dollars(capPer), maxTip)
	if r.chance(0.5) || alcohol == 0 {
		q.Text = fmt.Sprintf("Refer to %s. %s The tip on this team meal (no client present) is %d%%. How much of the tip is reimbursable?", ex, policy, tipPct)
		q.Choices, q.Correct = moneyQ(r, tipOK, []int{tip - tipOK + 100, 500, -300, 1000, tip - tipOK + 700})
		if tipOK == tip {
			q.Choices, q.Correct = moneyQ(r, tip, []int{-200, -500, 300, 800})
		}
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. %s No client was present. Which part of this receipt is NOT reimbursable?", ex, policy)
	q.Choices, q.Correct = withChoices(r, "The craft beer flight", []string{"The sales tax", "The food items", "The non-alcoholic drinks"})
	return q
}

// ── mileage log ────────────────────────────────────────────────────────

func mileageExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	rate := pick(r, []int{67, 70, 72}) // cents per mile, shown on the form
	places := []string{"Client site, Waukesha", "Records warehouse, Oak Creek", "PERCH venue walkthrough", "Vendor audit, Racine", "Board meeting, Brookfield", "Training center, Mequon", "Field visit, Kenosha"}
	n := r.between(3, 5)
	var rows [][]string
	business := 0
	commute := r.intn(n)
	d := exDate(r)
	for i := 0; i < n; i++ {
		mi := r.between(6, 88)
		purpose := pick(r, places)
		if i == commute {
			purpose = "Home to regular office (commute)"
			mi = r.between(8, 30)
		} else {
			business += mi
		}
		rows = append(rows, []string{d.AddDate(0, 0, i*r.between(1, 4)).Format("01/02"), purpose, fmt.Sprint(mi)})
	}
	q := Question{key: "ex/mileage", Exhibit: docExhibit("Mileage Log — "+sl["NAME"], fmt.Sprintf("Personal vehicle · reimbursement rate %d¢ per mile", rate), []string{"Date", "Destination / purpose", "Miles"}, rows, "Commuting between home and your regular workplace is not reimbursable.")}
	q.Text = fmt.Sprintf("Refer to %s. How much should %s be reimbursed for mileage?", ex, sl["FIRST"])
	total := 0
	for _, row := range rows {
		var m int
		fmt.Sscan(row[2], &m)
		total += m
	}
	q.Choices, q.Correct = moneyQ(r, business*rate, []int{(total - business) * rate, -rate * 10, rate * 10, 500})
	return q
}

// ── per diem (GSA 75% first and last day) ──────────────────────────────

func perDiemExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	mie := pick(r, []int{59, 64, 69, 74, 79, 86})
	days := r.between(2, 6)
	start := exDate(r)
	end := start.AddDate(0, 0, days-1)
	city := pick(r, []string{"Washington, DC", "Chicago, IL", "Denver, CO", "Boston, MA", "Austin, TX", "Seattle, WA", "Atlanta, GA"})
	rows := [][]string{
		{"Destination", city},
		{"Depart", start.Format("Mon Jan 2, 2006") + " (travel day)"},
		{"Return", end.Format("Mon Jan 2, 2006") + " (travel day)"},
		{"M&IE rate", fmt.Sprintf("$%d per full day", mie)},
		{"First and last day", "75% of the M&IE rate"},
	}
	q := Question{key: "ex/perdiem", Exhibit: docExhibit("Travel Authorization — "+sl["NAME"], "Meals and incidental expenses (M&IE) per diem", []string{"Item", "Detail"}, rows, "Meals provided by a conference or host reduce the per diem.")}
	full := mie * 100 * (days - 2)
	partial := mie * 75 * 2
	if days == 1 {
		full, partial = 0, mie*75
	}
	right := full + partial
	q.Text = fmt.Sprintf("Refer to %s. What is %s's total M&IE per diem for the trip?", ex, sl["FIRST"])
	q.Choices, q.Correct = moneyQ(r, right, []int{mie * 50, mie * 25, -mie * 50, mie*100*days - right, mie * 100})
	return q
}

// ── boarding pass (class of service) ───────────────────────────────────

func boardingPassExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	routes := []struct {
		from, to string
		mins     int
	}{{"MKE", "LGA", 130}, {"ORD", "LHR", 495}, {"MKE", "DEN", 165}, {"ORD", "NRT", 780}, {"MKE", "SFO", 270}, {"ORD", "FRA", 520}, {"MKE", "ATL", 125}, {"ORD", "HNL", 545}}
	rt := pick(r, routes)
	cabin := pick(r, []string{"ECONOMY", "PREMIUM ECONOMY", "BUSINESS", "FIRST"})
	threshold := pick(r, []int{6, 8})
	dep := r.between(6*60, 21*60)
	s := newSVG(420, 170, "Boarding pass")
	s.rect(2, 2, 416, 166, 12, "#fff", "#bbb")
	s.rect(2, 2, 416, 36, 12, "#1a3a5c", "")
	s.rect(2, 24, 416, 14, 0, "#1a3a5c", "")
	s.text(16, 26, 13, `font-weight="700" fill="#fff"`, "LAKESHORE AIR · BOARDING PASS")
	s.text(404, 26, 11, `text-anchor="end" fill="#e8c96a" font-weight="700"`, cabin)
	s.text(20, 82, 34, `font-weight="700" fill="#1a3a5c"`, rt.from)
	s.text(150, 82, 34, `font-weight="700" fill="#1a3a5c"`, rt.to)
	s.raw(`<path d="M104 70 l22 0 m-6 -5 l6 5 l-6 5" stroke="#1a3a5c" stroke-width="2" fill="none"/>`)
	for i, kv := range [][2]string{{"PASSENGER", strings.ToUpper(sl["NAME"])}, {"DEPARTS", clock(dep)}, {"SCHEDULED", fmt.Sprintf("%dh %02dm", rt.mins/60, rt.mins%60)}, {"SEAT", fmt.Sprintf("%d%s", r.between(2, 38), pick(r, []string{"A", "C", "D", "F"}))}} {
		x := 20 + float64(i%2)*130
		y := 108 + float64(i/2)*30
		s.text(x, y, 8, `fill="#888"`, kv[0])
		s.text(x, y+13, 11, `font-weight="700"`, kv[1])
	}
	s.dash(300, 44, 300, 162, "#bbb", 1)
	for i := 0; i < 26; i++ { // barcode
		w := float64(1 + r.intn(3))
		s.rect(316+float64(i)*3.4, 60, w, 80, 0, "#111", "")
	}
	q := Question{key: "ex/boardingpass", Exhibit: s.done()}
	allowed := cabin == "ECONOMY" || cabin == "PREMIUM ECONOMY" || (cabin == "BUSINESS" && rt.mins > threshold*60)
	q.Text = fmt.Sprintf("Refer to %s. ACPWB's travel policy allows economy or premium economy on any flight, business class only when the scheduled flight time is over %d hours, and never first class. Is this ticket within policy?", ex, threshold)
	yes := "Yes; the cabin is allowed for this flight's scheduled time"
	no := "No; this cabin is not allowed for this flight under the policy"
	right, wrong := yes, no
	if !allowed {
		right, wrong = no, yes
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Only if the traveler is a manager", "Only if the seat is on the aisle"})
	return q
}

// ── hotel folio ────────────────────────────────────────────────────────

func hotelFolioExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	nights := r.between(1, 4)
	rate := 100 * r.between(129, 289)
	type ch struct {
		desc string
		amt  int
		ok   bool
	}
	var chs []ch
	start := exDate(r)
	for i := 0; i < nights; i++ {
		chs = append(chs, ch{"Room charge " + start.AddDate(0, 0, i).Format("01/02"), rate, true}, ch{"Room tax", rate * 15 / 100, true})
	}
	extras := []ch{{"In-room movie", 100 * r.between(12, 22), false}, {"Minibar", 100*r.between(8, 40) + 50, false}, {"Business center printing", 100 * r.between(2, 9), true}, {"Spa services", 100 * r.between(60, 140), false}, {"Laundry (trip over 5 nights only)", 100 * r.between(15, 35), false}, {"Hotel Wi-Fi", 100 * r.between(10, 15), true}, {"Parking", 100 * r.between(18, 40), true}}
	for _, e := range sample(r, extras, r.between(2, 4)) {
		chs = append(chs, e)
	}
	ok, total := 0, 0
	var rows [][]string
	for _, x := range chs {
		rows = append(rows, []string{x.desc, money(x.amt)})
		total += x.amt
		if x.ok {
			ok += x.amt
		}
	}
	rows = append(rows, []string{"Balance", money(total)})
	q := Question{key: "ex/hotelfolio", Exhibit: docExhibit("Hotel Folio — "+sl["NAME"], fmt.Sprintf("%s · %d night(s)", pick(r, []string{"Pfister Hotel", "Hyatt Regency", "Lakeshore Suites", "Iron Horse Hotel"}), nights), []string{"Charge", "Amount"}, rows, "Reimbursable: room, room tax, business printing, Wi-Fi, and parking. Not reimbursable: movies, minibar, spa, and laundry on trips of 5 nights or fewer.")}
	q.Text = fmt.Sprintf("Refer to %s. How much of this folio is reimbursable?", ex)
	q.Choices, q.Correct = moneyQ(r, ok, []int{total - ok, rate, -rate * 15 / 100, 1500, -900})
	return q
}

// ── RSU vesting ────────────────────────────────────────────────────────

func vestingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	grant := 48 * r.between(10, 100) // divisible by 48 months
	grantDate := time.Date(r.between(2019, 2023), time.Month(r.between(1, 12)), 1, 0, 0, 0, 0, time.UTC)
	months := r.between(6, 47)
	asOf := grantDate.AddDate(0, months, 0)
	vested := 0
	if months >= 12 {
		vested = grant / 48 * months
	}
	s := newSVG(420, 230, "RSU vesting schedule")
	s.rect(0, 0, 420, 230, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, fmt.Sprintf("RSU Grant — %s shares, granted %s", commas(grant), shortDate(grantDate)))
	s.text(14, 38, 9, `fill="#666"`, "4-year schedule · 25% vests at the 1-year cliff, then 1/48 monthly")
	const x0, y0, w, h = 40.0, 200.0, 360.0, 128.0
	s.line(x0, y0, x0+w, y0, "#333", 1)
	s.line(x0, y0, x0, y0-h, "#333", 1)
	for m := 1; m <= 48; m++ {
		v := 0
		if m >= 12 {
			v = grant / 48 * m
		}
		bh := float64(v) / float64(grant) * h
		col := "#c9d6e3"
		if m <= months {
			col = "#1a3a5c"
		}
		s.rect(x0+float64(m-1)*7.5+1, y0-bh, 6, bh, 0, col, "")
	}
	for _, yr := range []int{1, 2, 3, 4} {
		x := x0 + float64(yr*12)*7.5 - 4
		s.text(x, y0+14, 9, `text-anchor="middle"`, fmt.Sprintf("Yr %d", yr))
	}
	ax := x0 + float64(months)*7.5 - 3
	s.dash(ax, y0, ax, y0-h-6, "#c62828", 1.2)
	s.text(ax, y0-h-10, 9, `text-anchor="middle" fill="#c62828" font-weight="700"`, "Today: "+shortDate(asOf))
	q := Question{key: "ex/vesting", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. How many shares have vested as of the date marked \"Today\"?", ex)
	q.Choices, q.Correct = numChoices(r, vested, []int{grant / 4, -grant / 48, grant / 48, grant / 48 * 3, -grant / 4}, func(n int) string { return commas(n) + " shares" })
	if months < 12 {
		q.Choices, q.Correct = withChoices(r, "0 shares", []string{commas(grant/48*months) + " shares", commas(grant/4) + " shares", commas(grant/12) + " shares"})
	}
	return q
}

// ── stock options ──────────────────────────────────────────────────────

func optionsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	strike := r.between(18, 90)
	price := strike + r.between(-15, 40)
	if price < 3 {
		price = 3
	}
	shares := 100 * r.between(5, 60)
	s := newSVG(420, 220, "Stock price chart")
	s.rect(0, 0, 420, 220, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, fmt.Sprintf("ACPWB share price · %s vested options", commas(shares)))
	lo, hi := float64(min(price, strike)-15), float64(max(price, strike)+15)
	yf := func(v float64) float64 { return 190 - (v-lo)/(hi-lo)*150 }
	var pts []string
	v := float64(strike) + float64(r.between(-8, 8))
	for i := 0; i <= 30; i++ {
		target := float64(price)
		v += (target-v)/float64(31-i) + float64(r.between(-3, 3))
		if i == 30 {
			v = float64(price)
		}
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", 30+float64(i)*12.3, yf(v)))
	}
	s.raw(`<polyline points="%s" fill="none" stroke="#1a3a5c" stroke-width="2"/>`, strings.Join(pts, " "))
	s.dash(30, yf(float64(strike)), 400, yf(float64(strike)), "#c62828", 1.2)
	s.text(34, yf(float64(strike))-5, 9, `fill="#c62828" font-weight="700"`, fmt.Sprintf("Exercise price $%d", strike))
	s.circle(399, yf(float64(price)), 4, "#e0a030", "")
	s.text(396, yf(float64(price))-8, 10, `text-anchor="end" font-weight="700"`, fmt.Sprintf("Today $%d", price))
	q := Question{key: "ex/options", Exhibit: s.done()}
	intrinsic := 0
	if price > strike {
		intrinsic = (price - strike) * shares
	}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. What is the pre-tax intrinsic value of these options today (share price minus exercise price, times the number of options)?", ex)
		q.Choices, q.Correct = dollarQ(r, intrinsic, []int{price*shares - intrinsic, strike * shares, shares * 5, -shares * 3})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Are these options \"underwater\" today?", ex)
	yes, no := "Yes; the share price is below the exercise price", "No; the share price is above the exercise price"
	if price == strike {
		no = "No; the share price equals the exercise price, so they are \"at the money\""
	}
	right, wrong := no, yes
	if price < strike {
		right, wrong = yes, no
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Yes, because options are always underwater until sold", "It depends on the weather in Milwaukee"})
	return q
}

// ── CEO pay ratio ──────────────────────────────────────────────────────

func payRatioExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	median := 1000 * r.between(42, 96)
	ratio := r.between(40, 420)
	ceo := median*ratio + 1000*r.between(-20, 20)
	right := int(math.Round(float64(ceo) / float64(median)))
	rows := [][]string{
		{"CEO annual total compensation", dollars(ceo)},
		{"Median employee annual total compensation", dollars(median)},
		{"Employees (excluding de minimis non-U.S.)", commas(r.between(1200, 40000))},
		{"Median employee identified", fmt.Sprint(r.between(2022, 2025))},
	}
	q := Question{key: "ex/payratio", Exhibit: docExhibit("CEO Pay Ratio Disclosure (Item 402(u))", "Proxy statement excerpt", []string{"Item", "Amount"}, rows, "Ratio = CEO total compensation ÷ median employee total compensation.")}
	q.Text = fmt.Sprintf("Refer to %s. What CEO pay ratio should be disclosed, rounded to the nearest whole number?", ex)
	q.Choices, q.Correct = numChoices(r, right, []int{-10, 10, right / 2, -right / 3, 25}, func(n int) string { return fmt.Sprintf("%d to 1", n) })
	return q
}

// ── Summary Compensation Table ─────────────────────────────────────────

func summaryCompExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	titles := []string{"Chief Executive Officer", "Chief Financial Officer", "Chief Operating Officer", "General Counsel", "Chief People Officer"}
	names := people(r, len(titles))
	head := []string{"Name and title", "Salary", "Bonus", "Stock awards", "Non-equity incentive", "All other", "Total"}
	var rows [][]string
	totals := make([]int, len(titles))
	for i := range titles {
		sal := 1000 * r.between(380, 1300)
		if i == 0 {
			sal = 1000 * r.between(1000, 1600)
		}
		vals := []int{sal, 1000 * r.between(0, 400), 1000 * r.between(400, 9000), 1000 * r.between(100, 2400), 1000 * r.between(10, 300)}
		t := 0
		row := []string{names[i] + ", " + titles[i]}
		for _, v := range vals {
			t += v
			row = append(row, dollars(v))
		}
		totals[i] = t
		row = append(row, "—")
		rows = append(rows, row)
	}
	k := r.intn(len(titles))
	q := Question{key: "ex/summarycomp", Exhibit: docExhibit("Summary Compensation Table (excerpt)", fmt.Sprintf("Fiscal %d · named executive officers", r.between(2018, 2025)), head, rows, "The Total column was left blank in this draft.")}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. What belongs in the Total column for %s?", ex, firstOf(names[k])+" "+strings.Fields(names[k])[1])
		var st []int
		for _, col := range rows[k][1:6] {
			var v int
			fmt.Sscan(strings.NewReplacer("$", "", ",", "").Replace(col), &v)
			st = append(st, v)
		}
		q.Choices, q.Correct = dollarQ(r, totals[k], []int{-st[2], -st[4], st[1], -st[0], 100000})
		return q
	}
	best := 0
	for i, t := range totals {
		if t > totals[best] {
			best = i
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. Which named executive officer had the highest total compensation?", ex)
	q.Choices, q.Correct = oneOf(r, names[best], names)
	return q
}

// ── say-on-pay results ─────────────────────────────────────────────────

func sayOnPayExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	forV := r.between(20, 95) * 1_000_000 / 10
	againstV := r.between(5, 60) * 1_000_000 / 10
	abst := r.between(1, 12) * 1_000_000 / 10
	broker := r.between(5, 30) * 1_000_000 / 10
	pct := int(math.Round(float64(forV) * 100 / float64(forV+againstV)))
	s := newSVG(420, 220, "Say-on-pay vote results")
	s.rect(0, 0, 420, 220, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, "Advisory Vote on Executive Compensation — Final Tally")
	tot := float64(forV + againstV + abst)
	cx, cy, rad := 110.0, 125.0, 70.0
	start := -math.Pi / 2
	for _, seg := range []struct {
		v   int
		col string
	}{{forV, "#2e7d32"}, {againstV, "#c62828"}, {abst, "#9e9e9e"}} {
		a := float64(seg.v) / tot * 2 * math.Pi
		x1, y1 := cx+rad*math.Cos(start), cy+rad*math.Sin(start)
		x2, y2 := cx+rad*math.Cos(start+a), cy+rad*math.Sin(start+a)
		large := 0
		if a > math.Pi {
			large = 1
		}
		s.raw(`<path d="M%.1f %.1f L%.1f %.1f A%.0f %.0f 0 %d 1 %.1f %.1f Z" fill="%s" stroke="#fff" stroke-width="2"/>`, cx, cy, x1, y1, rad, rad, large, x2, y2, seg.col)
		start += a
	}
	s.circle(cx, cy, 34, "#fff", "")
	for i, kv := range []struct {
		l, col string
		v      int
	}{{"For", "#2e7d32", forV}, {"Against", "#c62828", againstV}, {"Abstain", "#9e9e9e", abst}, {"Broker non-votes", "#fff", broker}} {
		y := 70 + float64(i)*28
		s.rect(220, y-10, 12, 12, 2, kv.col, "#999")
		s.text(240, y, 11, `font-weight="700"`, kv.l)
		s.text(406, y, 11, `text-anchor="end"`, commas(kv.v))
	}
	s.text(220, 190, 9, `fill="#666"`, "Per the bylaws, support = For ÷ (For + Against).")
	q := Question{key: "ex/sayonpay", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the bylaws' method, what percentage support did the say-on-pay proposal receive?", ex)
	q.Choices, q.Correct = pctQ(r, pct)
	return q
}

// ── clawback lookback (Rule 10D-1) ─────────────────────────────────────

func clawbackExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	fy := r.between(2019, 2025) // the year the restatement is required
	var rows [][]string
	for y := fy - 5; y <= fy; y++ {
		rows = append(rows, []string{fmt.Sprintf("FY%d", y), dollars(1000 * r.between(200, 1800)), pick(r, []string{"Revenue growth", "Operating margin", "EPS", "Adjusted EBITDA"})})
	}
	q := Question{key: "ex/clawback", Exhibit: docExhibit("Incentive Compensation Received — "+sl["NAME"], fmt.Sprintf("Restatement required: board concluded in FY%d that prior financial statements must be restated", fy), []string{"Fiscal year received", "Incentive pay", "Financial reporting measure"}, rows, "Fiscal years end December 31. Rule 10D-1 lookback: the three completed fiscal years before the date the restatement was required.")}
	right := fmt.Sprintf("FY%d, FY%d, and FY%d", fy-3, fy-2, fy-1)
	q.Text = fmt.Sprintf("Refer to %s. Incentive pay received in which fiscal years is subject to recovery under the clawback rule?", ex)
	q.Choices, q.Correct = withChoices(r, right, []string{
		fmt.Sprintf("FY%d, FY%d, and FY%d", fy-2, fy-1, fy),
		fmt.Sprintf("FY%d only", fy),
		fmt.Sprintf("FY%d through FY%d", fy-5, fy),
	})
	return q
}

// ── 401(k) match ───────────────────────────────────────────────────────

func match401kExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	salary := 1000 * r.between(42, 160)
	defer_ := r.between(1, 10)
	tier1, tier2 := pick(r, []int{3, 4}), 2
	m := 0
	if defer_ <= tier1 {
		m = salary * defer_ / 100
	} else {
		m = salary*tier1/100 + salary*min(defer_-tier1, tier2)/200
	}
	rows := [][]string{
		{"Annual eligible pay", dollars(salary)},
		{"Employee deferral rate", fmt.Sprintf("%d%%", defer_)},
		{"Employer match", fmt.Sprintf("100%% of the first %d%% of pay deferred, plus 50%% of the next %d%%", tier1, tier2)},
	}
	q := Question{key: "ex/match401k", Exhibit: docExhibit("401(k) Contribution Summary — "+sl["NAME"], "Plan year", []string{"Item", "Detail"}, rows, "Matching contributions are calculated on eligible pay for the full year.")}
	q.Text = fmt.Sprintf("Refer to %s. What is %s's employer match for the year?", ex, sl["FIRST"])
	q.Choices, q.Correct = dollarQ(r, m, []int{salary*defer_/100 - m, salary / 100, -salary / 100, salary / 50, salary / 200})
	return q
}

// ── retirement plan vesting ────────────────────────────────────────────

func retireVestExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	sched := []int{0, 0, 20, 40, 60, 80, 100} // graded 2-to-6-year
	yrs := r.between(1, 7)
	bal := 100 * r.between(30, 900) * 10
	pct := sched[min(yrs, 6)]
	var rows [][]string
	for y := 1; y <= 6; y++ {
		rows = append(rows, []string{fmt.Sprintf("%d", y), fmt.Sprintf("%d%%", sched[y])})
	}
	q := Question{key: "ex/retirevest", Exhibit: docExhibit("Employer Contribution Vesting Schedule", fmt.Sprintf("%s · %d completed years of service · employer contributions %s", sl["NAME"], yrs, money(bal)), []string{"Years of service", "Vested"}, rows, "Your own deferrals are always 100% vested.")}
	q.Text = fmt.Sprintf("Refer to %s. If %s leaves ACPWB now, how much of the employer contribution balance is vested?", ex, sl["FIRST"])
	q.Choices, q.Correct = moneyQ(r, bal*pct/100, []int{bal / 5, -bal / 5, bal - bal*pct/100, bal / 10})
	return q
}

// ── HSA contribution limit ─────────────────────────────────────────────

func hsaLimitExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	family := r.chance(0.5)
	limit := 4400
	cov := "Self-only HDHP coverage"
	if family {
		limit, cov = 8750, "Family HDHP coverage"
	}
	age55 := r.chance(0.3)
	if age55 {
		limit += 1000
		cov += " · age 55+ catch-up"
	}
	emp := 50 * r.between(0, 30)
	ee := 50 * r.between(10, limit/50-emp/50-2)
	rows := [][]string{
		{"Coverage", cov},
		{"Annual contribution limit", dollars(limit)},
		{"Employer contributions to date", dollars(emp)},
		{"Employee payroll contributions to date", dollars(ee)},
	}
	q := Question{key: "ex/hsalimit", Exhibit: docExhibit("HSA Contribution Statement — "+sl["NAME"], "Plan year to date", []string{"Item", "Amount"}, rows, "Employer and employee contributions both count toward the annual limit.")}
	q.Text = fmt.Sprintf("Refer to %s. How much more can %s contribute this year without exceeding the limit?", ex, sl["FIRST"])
	q.Choices, q.Correct = dollarQ(r, limit-emp-ee, []int{emp, -500, 1000, 500})
	return q
}

// ── health plan comparison ─────────────────────────────────────────────

func planPickExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type plan struct {
		name             string
		prem, ded, coins int // prem per paycheck (26), coins %
		oop              int
	}
	plans := []plan{
		{"Bronze HDHP", 10 * r.between(2, 6), 100 * r.between(30, 45), 20, 100 * r.between(65, 80)},
		{"Silver PPO", 10 * r.between(6, 11), 100 * r.between(10, 20), 20, 100 * r.between(40, 55)},
		{"Gold PPO", 10 * r.between(11, 17), 100 * r.between(2, 6), 10, 100 * r.between(20, 30)},
	}
	spend := 100 * r.between(5, 120)
	cost := func(p plan) int {
		oop := spend
		if spend > p.ded {
			oop = p.ded + (spend-p.ded)*p.coins/100
		}
		if oop > p.oop {
			oop = p.oop
		}
		return p.prem*26 + oop
	}
	var rows [][]string
	best := 0
	for i, p := range plans {
		rows = append(rows, []string{p.name, dollars(p.prem), dollars(p.ded), fmt.Sprintf("%d%%", p.coins), dollars(p.oop)})
		if cost(p) < cost(plans[best]) {
			best = i
		}
	}
	q := Question{key: "ex/planpick", Exhibit: docExhibit("Medical Plan Comparison", "Employee-only coverage · 26 paychecks", []string{"Plan", "Premium per paycheck", "Deductible", "Coinsurance after deductible", "Out-of-pocket max"}, rows, "Out-of-pocket costs = deductible, then coinsurance, up to the out-of-pocket maximum.")}
	tie := false
	for i, p := range plans {
		if i != best && cost(p) == cost(plans[best]) {
			tie = true
		}
	}
	if r.chance(0.5) && !tie {
		q.Text = fmt.Sprintf("Refer to %s. %s expects %s in covered medical expenses next year. Which plan has the lowest total cost (premiums plus out-of-pocket)?", ex, sl["FIRST"], dollars(spend))
		q.Choices, q.Correct = withChoices(r, plans[best].name, func() []string {
			var w []string
			for i, p := range plans {
				if i != best {
					w = append(w, p.name)
				}
			}
			return append(w, "They all cost the same")
		}())
		return q
	}
	k := r.intn(3)
	q.Text = fmt.Sprintf("Refer to %s. With %s in covered medical expenses, what would %s pay in total for the year under the %s (premiums plus out-of-pocket)?", ex, dollars(spend), sl["FIRST"], plans[k].name)
	q.Choices, q.Correct = dollarQ(r, cost(plans[k]), []int{-plans[k].prem * 26, plans[k].ded, -spend / 2, 300})
	return q
}

// ── annual bonus ───────────────────────────────────────────────────────

func bonusCalcExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	salary := 1000 * r.between(55, 220)
	target := pick(r, []int{5, 8, 10, 12, 15, 20, 25})
	co := pick(r, []int{0, 50, 80, 90, 100, 110, 125, 150})
	ind := pick(r, []int{0, 75, 100, 110, 125, 150})
	bonus := salary * target / 100 * co / 100 * ind / 100
	rows := [][]string{
		{"Base salary", dollars(salary)},
		{"Target bonus", fmt.Sprintf("%d%% of salary", target)},
		{"Company performance multiplier", fmt.Sprintf("%d%%", co)},
		{"Individual performance multiplier", fmt.Sprintf("%d%%", ind)},
	}
	q := Question{key: "ex/bonuscalc", Exhibit: docExhibit("Annual Incentive Statement — "+sl["NAME"], "Bonus = salary × target × company multiplier × individual multiplier", []string{"Component", "Value"}, rows, "Bonuses are paid in March after the audit of annual results.")}
	q.Text = fmt.Sprintf("Refer to %s. What bonus will %s receive?", ex, sl["FIRST"])
	q.Choices, q.Correct = dollarQ(r, bonus, []int{salary*target/100 - bonus + 1, salary*target/100*co/100 - bonus + 7, 1000, -1000, 2500})
	return q
}

// ── tiered commission ──────────────────────────────────────────────────

func commissionExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	t1, t2 := 1000*r.between(80, 150), 1000*r.between(200, 320)
	r1, r2, r3 := pick(r, []int{2, 3, 4}), pick(r, []int{5, 6}), pick(r, []int{8, 10})
	sales := 1000 * r.between(40, 450)
	com := 0
	com += min(sales, t1) * r1 / 100
	if sales > t1 {
		com += (min(sales, t2) - t1) * r2 / 100
	}
	if sales > t2 {
		com += (sales - t2) * r3 / 100
	}
	rows := [][]string{
		{"$0 – " + dollars(t1), fmt.Sprintf("%d%%", r1)},
		{dollars(t1) + " – " + dollars(t2), fmt.Sprintf("%d%%", r2)},
		{"Over " + dollars(t2), fmt.Sprintf("%d%%", r3)},
	}
	q := Question{key: "ex/commission", Exhibit: docExhibit("Commission Plan — Quarterly Tiers", fmt.Sprintf("%s · quarter bookings %s", sl["NAME"], dollars(sales)), []string{"Bookings tier", "Rate on bookings in tier"}, rows, "Each rate applies only to the bookings within its tier.")}
	q.Text = fmt.Sprintf("Refer to %s. What commission did %s earn this quarter?", ex, sl["FIRST"])
	top := sales * r3 / 100
	if sales <= t1 {
		top = sales * r2 / 100
	}
	q.Choices, q.Correct = dollarQ(r, com, []int{top - com, sales*r2/100 - com, 500, -500, 1200})
	return q
}

// ── relocation repayment ───────────────────────────────────────────────

func relocationExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	amt := 500 * r.between(10, 60)
	term := pick(r, []int{12, 24})
	left := r.between(1, term-1)
	owed := amt * (term - left) / term
	rows := [][]string{
		{"Relocation paid", dollars(amt)},
		{"Repayment period", fmt.Sprintf("%d months from start date", term)},
		{"Repayment if you resign", "Prorated: the unearned share of the remaining months"},
		{"Months worked before resigning", fmt.Sprint(left)},
	}
	q := Question{key: "ex/relocation", Exhibit: docExhibit("Relocation Repayment Agreement", sl["NAME"]+" · voluntary resignation", []string{"Term", "Detail"}, rows, "No repayment is due after the repayment period ends.")}
	q.Text = fmt.Sprintf("Refer to %s. How much must %s repay?", ex, sl["FIRST"])
	q.Choices, q.Correct = dollarQ(r, owed, []int{amt - owed, amt / term, -amt / term, amt / 4})
	return q
}

// ── pay stub ───────────────────────────────────────────────────────────

func payStubExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	gross := 100 * r.between(1400, 6200)
	ded := []struct {
		name   string
		amt    int
		pretax bool
	}{
		{"401(k) deferral", gross * pick(r, []int{3, 4, 5, 6, 8}) / 100, true},
		{"Medical premium (Section 125)", 100 * r.between(40, 220), true},
		{"Federal income tax", gross * pick(r, []int{8, 10, 12, 14}) / 100, false},
		{"Social Security (6.2%)", gross * 62 / 1000, false},
		{"Medicare (1.45%)", gross * 145 / 10000, false},
		{"State income tax", gross * pick(r, []int{3, 4, 5}) / 100, false},
	}
	if r.chance(0.4) {
		ded = append(ded, struct {
			name   string
			amt    int
			pretax bool
		}{"Charity (United Way)", 100 * r.between(5, 25), false})
	}
	net := gross
	for _, d := range ded {
		net -= d.amt
	}
	s := newSVG(380, 120+float64(len(ded))*18+50, "Pay stub")
	s.rect(1, 1, 378, 120+float64(len(ded))*18+48, 4, "#f9fbf6", "#9fb59a")
	s.text(14, 24, 13, `font-weight="700" fill="#2e5e3a"`, "ACPWB · EARNINGS STATEMENT")
	s.text(366, 24, 9, `text-anchor="end" fill="#555"`, "Pay date "+shortDate(exDate(r)))
	s.text(14, 44, 10, "", sl["NAME"]+" · biweekly")
	s.text(14, 70, 11, `font-weight="700"`, "Gross pay")
	s.text(366, 70, 11, `text-anchor="end" font-weight="700"`, money(gross))
	s.line(14, 80, 366, 80, "#9fb59a", 1)
	y := 98.0
	for _, d := range ded {
		s.text(24, y, 10, "", d.name)
		s.text(366, y, 10, `text-anchor="end"`, "−"+money(d.amt))
		y += 18
	}
	s.line(14, y-8, 366, y-8, "#9fb59a", 1)
	s.text(14, y+12, 12, `font-weight="700"`, "Net pay")
	s.rect(260, y-2, 106, 20, 3, "#e6efe2", "")
	s.text(366, y+12, 12, `text-anchor="end" font-weight="700"`, "?")
	q := Question{key: "ex/paystub", Exhibit: s.done()}
	if r.chance(0.6) {
		q.Text = fmt.Sprintf("Refer to %s. The net pay amount is smudged. What is %s's net pay?", ex, sl["FIRST"])
		q.Choices, q.Correct = moneyQ(r, net, []int{ded[0].amt, -ded[2].amt, ded[3].amt, 1000, -1500})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Which deduction reduces %s's taxable wages for federal income tax?", ex, sl["FIRST"])
	var wrong []string
	for _, d := range ded[2:] {
		wrong = append(wrong, d.name)
	}
	q.Choices, q.Correct = oneOf(r, ded[r.intn(2)].name, wrong)
	return q
}

// ── approval matrix ────────────────────────────────────────────────────

func approvalMatrixExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	roles := []string{"Manager", "Director", "Vice President", "Chief Financial Officer", "Board of Directors"}
	limits := []int{5000, 25000, 100000, 1000000}
	if r.chance(0.5) {
		limits = []int{10000, 50000, 250000, 2000000}
	}
	var rows [][]string
	for i, role := range roles {
		lo := 0
		if i > 0 {
			lo = limits[i-1] + 1
		}
		rng := dollars(lo) + " – " + dollars(limits[min(i, 3)])
		if i == 4 {
			rng = "Over " + dollars(limits[3])
		}
		rows = append(rows, []string{role, rng})
	}
	k := r.intn(5)
	var amt int
	if k == 0 {
		amt = r.between(limits[0]/4, limits[0])
	} else if k == 4 {
		amt = limits[3] + r.between(1, limits[3])
	} else {
		amt = r.between(limits[k-1]+1, limits[k])
	}
	amt = amt / 100 * 100
	if amt <= 0 {
		amt = 100
	}
	if k > 0 && amt <= limits[k-1] {
		amt = limits[k-1] + 100
	}
	what := pick(r, []string{"a consulting statement of work", "a software subscription", "a three-year facilities contract", "an event venue booking for the PERCH Conference", "a records storage agreement"})
	q := Question{key: "ex/approvals", Exhibit: docExhibit("Delegation of Authority — Purchase Commitments", "Total contract value, including all committed years", []string{"Minimum approver", "Commitment value"}, rows, "Splitting a commitment to fit under a lower limit is prohibited.")}
	q.Text = fmt.Sprintf("Refer to %s. Who is the lowest-level approver for %s with a total value of %s?", ex, what, dollars(amt))
	q.Choices, q.Correct = oneOf(r, roles[k], roles)
	return q
}

// ── pay equity scatter ─────────────────────────────────────────────────

func payScatterExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	levels := []string{"L1", "L2", "L3", "L4"}
	bands := [][2]int{{48, 66}, {62, 84}, {78, 104}, {96, 130}}
	names := people(r, 6)
	type pt struct {
		name string
		lvl  int
		pay  int
	}
	var pts []pt
	out := r.intn(len(names))
	perLevel := map[int]int{}
	for i, n := range names {
		l := r.intn(4)
		for perLevel[l] >= 2 { // at most two people per level keeps labels apart
			l = (l + 1) % 4
		}
		perLevel[l]++
		p := r.between(bands[l][0]+2, bands[l][1]-2)
		if i == out {
			p = bands[l][0] - r.between(5, 10)
		}
		pts = append(pts, pt{n, l, p})
	}
	s := newSVG(460, 290, "Pay by level scatter plot")
	s.rect(0, 0, 460, 290, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, "Base pay by level, "+sl["DEPT"])
	x0, y0 := 50.0, 250.0
	yf := func(k int) float64 { return y0 - float64(k-36)/100*210 }
	for k := 40; k <= 140; k += 20 {
		s.line(x0, yf(k), 446, yf(k), "#eee", 1)
		s.text(x0-6, yf(k)+3, 9, `text-anchor="end" fill="#666"`, fmt.Sprintf("$%dk", k))
	}
	for l, b := range bands {
		x := x0 + 50 + float64(l)*98
		s.rect(x-40, yf(b[1]), 80, yf(b[0])-yf(b[1]), 3, "#e8eef5", "#c9d6e3")
		s.text(x, y0+18, 10, `text-anchor="middle" font-weight="700"`, levels[l])
	}
	seen := map[int]int{}
	for _, p := range pts {
		k := seen[p.lvl]
		seen[p.lvl]++
		x := x0 + 50 + float64(p.lvl)*98 - 22 + float64(k)*44
		s.circle(x, yf(p.pay), 5, "#1a3a5c", "#fff")
		s.text(x, yf(p.pay)-9, 9, `text-anchor="middle" fill="#333" font-weight="700"`, strings.Fields(p.name)[0])
	}
	s.text(14, 282, 9, `fill="#666"`, "Shaded bands show each level's approved pay range.")
	q := Question{key: "ex/payscatter", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Whose pay falls below the approved range for their level and should be reviewed first?", ex)
	var firsts []string
	for _, p := range pts {
		firsts = append(firsts, strings.Fields(p.name)[0])
	}
	q.Choices, q.Correct = oneOf(r, firsts[out], firsts)
	return q
}

// ── job postings (pay transparency) ────────────────────────────────────

func jobPostingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	required := []string{"Denver, CO", "Los Angeles, CA", "Seattle, WA", "New York, NY", "Chicago, IL"}
	other := []string{"Milwaukee, WI", "Dallas, TX", "Atlanta, GA", "Phoenix, AZ"}
	roles := sample(r, []string{"Compensation Analyst", "Payroll Specialist", "HR Business Partner", "Benefits Manager", "Data Steward", "Facilities Coordinator"}, 4)
	bad := r.intn(4)
	s := newSVG(440, 300, "Job postings")
	s.rect(0, 0, 440, 300, 8, "#f4f6f9", "")
	type post struct{ role, loc, pay string }
	var ps []post
	for i := 0; i < 4; i++ {
		lo := 1000 * r.between(48, 110)
		p := post{roles[i], pick(r, append(required, other...)), fmt.Sprintf("%s – %s", dollars(lo), dollars(lo+1000*r.between(12, 40)))}
		if i == bad {
			p.loc = pick(r, required)
			p.pay = "Competitive salary"
		} else if r.chance(0.3) && contains(other, p.loc) {
			p.pay = "Competitive salary"
		}
		ps = append(ps, p)
		x, y := 12+float64(i%2)*216, 14+float64(i/2)*142
		s.rect(x, y, 204, 130, 6, "#fff", "#d0d6de")
		s.text(x+12, y+24, 12, `font-weight="700" fill="#1a3a5c"`, p.role)
		s.text(x+12, y+42, 9, `fill="#666"`, "ACPWB · "+p.loc)
		s.text(x+12, y+66, 10, "", "Pay: "+p.pay)
		s.text(x+12, y+86, 9, `fill="#666"`, "Full-time · hybrid")
		s.rect(x+12, y+100, 60, 18, 4, "#1a3a5c", "")
		s.text(x+42, y+113, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, "Apply")
	}
	q := Question{key: "ex/jobposting", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Colorado, California, Washington, New York City, and Illinois require job postings to include a pay range. Which posting is out of compliance?", ex)
	var all []string
	for _, p := range ps {
		all = append(all, p.role+" ("+p.loc+")")
	}
	q.Choices, q.Correct = oneOf(r, all[bad], all)
	return q
}

// ── cash deposit log (structuring) ─────────────────────────────────────

func cashLogExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	custs := people(r, 5)
	smurf := r.intn(5)
	day := exDate(r)
	type dep struct {
		d   time.Time
		who string
		amt int
	}
	var deps []dep
	for i, cname := range custs {
		if i == smurf {
			for k := 0; k < r.between(3, 4); k++ {
				deps = append(deps, dep{day.AddDate(0, 0, k+r.intn(2)), cname, 100 * r.between(9000, 9900)})
			}
			continue
		}
		deps = append(deps, dep{day.AddDate(0, 0, r.between(0, 6)), cname, 100 * r.between(200, 7000)})
		if r.chance(0.3) {
			deps = append(deps, dep{day.AddDate(0, 0, r.between(0, 6)), cname, 100 * r.between(10500, 18000)})
		}
	}
	for i := 1; i < len(deps); i++ {
		for j := i; j > 0 && deps[j].d.Before(deps[j-1].d); j-- {
			deps[j], deps[j-1] = deps[j-1], deps[j]
		}
	}
	var rows [][]string
	for _, d := range deps {
		rows = append(rows, []string{d.d.Format("Mon 01/02"), d.who, money(d.amt)})
	}
	q := Question{key: "ex/cashlog", Exhibit: docExhibit("Cash Deposit Log — Lakeshore Branch", "Currency Transaction Report threshold: cash over $10,000 in one business day", []string{"Date", "Customer", "Cash deposited"}, rows, "Deposits over the threshold are reported automatically.")}
	q.Text = fmt.Sprintf("Refer to %s. Which customer's activity is a red flag for structuring?", ex)
	q.Choices, q.Correct = oneOf(r, custs[smurf], custs)
	return q
}
