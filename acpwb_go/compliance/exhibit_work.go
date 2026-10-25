package compliance

import (
	"fmt"
	"strings"
)

// Modern-workplace exhibits: calendar invites, desk-booking maps, hybrid
// team schedules, badge swipe reports, PTO statements, FMLA timelines, and a
// manager's chat during a union campaign. Answers are computed from what is
// shown; FMLA and NLRA answers follow the actual law.

func init() {
	for k, g := range map[string]exhibitGen{
		"invite":       inviteExhibit,
		"deskmap":      deskMapExhibit,
		"hybridgrid":   hybridGridExhibit,
		"badgelog":     badgeLogExhibit,
		"ptostatement": ptoStatementExhibit,
		"fmla":         fmlaExhibit,
		"tipschat":     tipsChatExhibit,
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
		"calendar": {"invite"},
		"meetings": {"calendar", "invite"},
		"hotdesk":  {"deskmap"},
		"hybrid":   {"hybridgrid", "invite"},
		"rto":      {"badgelog", "hybridgrid", "deskmap"},
		"remote":   {"desk", "hybridgrid"},
		"pto":      {"ptostatement"},
		"leave":    {"fmla", "ptostatement"},
		"parental": {"fmla"},
		"union":    {"tipschat"},
	} {
		categoryExhibits[cat] = ks
	}
}

// ── calendar invite ─────────────────────────────────────────────────────

var timeZones = []struct {
	Name string
	Off  int // hours ahead of Central
	City string
}{
	{"Eastern Time", 1, "New York"}, {"Central Time", 0, "Chicago"}, {"Mountain Time", -1, "Denver"}, {"Pacific Time", -2, "Seattle"},
}

var meetingTitles = []string{"{DEPT} weekly sync", "Q{N} planning", "PERCH logistics check-in", "Budget review", "Vendor update: {VENDOR}",
	"Return to Office working group", "Desk neighborhood kickoff", "Policy {POLICY} walkthrough", "Hybrid schedule check-in", "{COMMITTEE} prep",
	"Quick sync", "Retro", "AI tool pilot readout", "Office move logistics", "Holiday party planning"}

func inviteExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	g := gen{r, sl}
	tz := pick(r, []int{0, 2, 3, 3})
	z := timeZones[tz]
	start := 60*r.between(8, 15) + 15*r.intn(4)
	dur := pick(r, []int{15, 25, 30, 45, 50, 60, 90})
	required := r.between(3, 14)
	optional := r.between(0, 6)
	missing := -1 // 0 agenda, 1 location
	if r.chance(0.5) {
		missing = r.intn(2)
	}
	agenda := g.expand("[Review open items from last week, decide on next steps, and confirm owners.|1. Updates (10 min) 2. Decisions needed 3. Next steps|Walk through the draft and agree on changes before {WEEKDAY}.|Goal: agree on the plan. Pre-read attached.]")
	loc := pick(r, []string{"Video call (link in invite)", sl["FLOOR"] + " floor, " + pick(r, roomNames) + " room + video link", "Video call + " + pick(r, roomNames) + " room"})
	if missing == 0 {
		agenda = "(no description)"
	}
	if missing == 1 {
		loc = "(none)"
	}
	recur := pick(r, []string{"Does not repeat", "Weekly on " + sl["WEEKDAY"], "Every two weeks", "Monthly"})
	day := exDate(r)
	title := g.expand(pick(r, meetingTitles))
	rows := [][2]string{
		{"When", fmt.Sprintf("%s, %s – %s (%s)", day.Format("Monday, January 2"), clock(start), clock(start+dur), z.Name)},
		{"Repeats", recur},
		{"Organizer", sl["NAME"] + " (" + z.City + ")"},
		{"Required", fmt.Sprintf("%d people", required)},
		{"Optional", fmt.Sprintf("%d people", optional)},
		{"Location", loc},
		{"Description", agenda},
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="cmp-ex-mail"><div class="cmp-ex-mailbar">📅 Meeting invitation</div><div style="padding:.5rem .7rem;font-weight:700;font-size:.95rem">%s</div><dl>`, esc(title))
	for _, rw := range rows {
		fmt.Fprintf(&b, `<dt>%s</dt><dd>%s</dd>`, esc(rw[0]), esc(rw[1]))
	}
	b.WriteString(`</dl><div class="cmp-ex-mailbody" style="display:flex;gap:.4rem"><span class="cmp-st cmp-st-ok">Accept</span><span class="cmp-st cmp-st-prog">Tentative</span><span class="cmp-st cmp-st-prog">Decline</span></div></div>`)
	q := Question{key: "ex/invite", Exhibit: b.String()}
	switch v := r.intn(4); {
	case v == 0 && missing >= 0:
		q.Text = fmt.Sprintf("Refer to %s. Which element required by the ACPWB Meeting Invite Standard is missing from this invitation?", ex)
		names := []string{"An agenda or description", "A location or meeting link"}
		q.Choices, q.Correct = withChoices(r, names[missing], []string{names[1-missing], "A time zone", "An end time"})
	case v <= 1:
		local := start - z.Off*60
		q.Text = fmt.Sprintf("Refer to %s. You work in Milwaukee (Central Time). What time does this meeting start for you?", ex)
		q.Choices, q.Correct = numChoices(r, local, []int{60, -60, 120, -120, 180}, clock)
	case v == 2:
		q.Text = fmt.Sprintf("Refer to %s. How long is this meeting?", ex)
		q.Choices, q.Correct = numChoices(r, dur, []int{-15, 15, 30, -10, 10, 45}, func(m int) string { return fmt.Sprintf("%d minutes", m) })
	default:
		q.Text = fmt.Sprintf("Refer to %s. How many people are expected to attend (required attendees only)?", ex)
		q.Choices, q.Correct = numChoices(r, required, []int{optional, required + optional - required, -1, 1, 2}, itoa)
	}
	return q
}

// ── desk booking map ────────────────────────────────────────────────────

func deskMapExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type desk struct {
		id     string
		zone   int
		window bool
		status int // 0 available, 1 booked, 2 blocked
		who    string
	}
	zones := sample(r, []string{"Finance", "People Ops", "Legal", "Marketing", "Facilities", "Compliance", "IT", "Research"}, 3)
	zoneColor := []string{"#e3ecf7", "#f6ead2", "#e5f2e3"}
	const rows, cols = 4, 6
	var desks []desk
	floor := sl["FLOOR"]
	fl := strings.TrimRight(floor, "stndrdth")
	for rw := 0; rw < rows; rw++ {
		for cl := 0; cl < cols; cl++ {
			d := desk{id: fmt.Sprintf("%s-%c%d", fl, 'A'+rune(rw), cl+1), zone: cl / 2, window: rw == 0}
			x := r.float()
			switch {
			case x < 0.5:
				d.status = 1
				d.who = strings.ToUpper(string([]rune(pick(r, firstNames))[0:1]) + string([]rune(pick(r, lastNames))[0:1]))
			case x < 0.62:
				d.status = 2
			}
			desks = append(desks, d)
		}
	}
	tz := r.intn(3)
	// Exactly one available window desk in the target zone.
	var win []int
	for i, d := range desks {
		if d.zone == tz && d.window {
			win = append(win, i)
		}
	}
	keep := pick(r, win)
	for _, i := range win {
		if i == keep {
			desks[i].status, desks[i].who = 0, ""
		} else if desks[i].status == 0 {
			desks[i].status, desks[i].who = 1, "JR"
		}
	}
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="300" viewBox="0 0 460 300" role="img" aria-label="Desk booking map"><rect width="460" height="300" fill="#fff"/>`)
	fmt.Fprintf(&b, `<text x="230" y="18" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">DESK BOOKING — %s FLOOR · %s</text>`, svgFont, esc(strings.ToUpper(floor)), esc(strings.ToUpper(sl["WEEKDAY"])))
	b.WriteString(`<rect x="20" y="28" width="420" height="8" fill="#bfe0f2"/>`)
	fmt.Fprintf(&b, `<text x="230" y="35" text-anchor="middle" %s font-size="6.5" fill="#3b6a8a">WINDOWS</text>`, svgFont)
	for z := 0; z < 3; z++ {
		fmt.Fprintf(&b, `<rect x="%d" y="42" width="136" height="214" fill="%s"/><text x="%d" y="250" text-anchor="middle" %s font-size="9" font-weight="700" fill="#333">%s</text>`, 22+z*140, zoneColor[z], 90+z*140, svgFont, esc(strings.ToUpper(zones[z])))
	}
	avail := make([]int, 3)
	for i, d := range desks {
		rw, cl := i/cols, i%cols
		x, y := 30+float64(cl)*70, 50+float64(rw)*48
		fill, ink, label := "#fff", "#1f7a3a", "OPEN"
		switch d.status {
		case 1:
			fill, ink, label = "#2f3b52", "#fff", d.who
		case 2:
			fill, ink, label = "#e0e0e0", "#777", "CLEANING"
		default:
			avail[d.zone]++
		}
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="56" height="34" rx="4" fill="%s" stroke="#9aa5b1"/><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="7" fill="%s">%s</text><text x="%.0f" y="%.0f" text-anchor="middle" %s font-size="8" font-weight="700" fill="%s">%s</text>`,
			x, y, fill, x+28, y+12, svgFont, ink, d.id, x+28, y+26, svgFont, ink, esc(label))
	}
	fmt.Fprintf(&b, `<g %s font-size="8" fill="#333"><rect x="30" y="272" width="12" height="10" fill="#fff" stroke="#9aa5b1"/><text x="46" y="281">Open</text><rect x="90" y="272" width="12" height="10" fill="#2f3b52"/><text x="106" y="281">Booked</text><rect x="160" y="272" width="12" height="10" fill="#e0e0e0" stroke="#9aa5b1"/><text x="176" y="281">Blocked for cleaning</text></g>`, svgFont)
	b.WriteString(`</svg>`)
	q := Question{key: "ex/deskmap", Exhibit: b.String()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. You sit with the %s team and would like a window seat. Which desk can you book?", ex, zones[tz])
		var wrongs []string
		for i, d := range desks {
			if i != keep && (d.zone == tz || d.window) && len(wrongs) < 3 && r.chance(0.7) {
				wrongs = append(wrongs, d.id)
			}
		}
		for i := 0; len(wrongs) < 3; i++ {
			if d := desks[(keep+5+i)%len(desks)]; d.id != desks[keep].id && !contains(wrongs, d.id) {
				wrongs = append(wrongs, d.id)
			}
		}
		q.Choices, q.Correct = withChoices(r, desks[keep].id, wrongs)
		return q
	}
	z := r.intn(3)
	q.Text = fmt.Sprintf("Refer to %s. How many desks are open to book in the %s neighborhood?", ex, zones[z])
	q.Choices, q.Correct = numChoices(r, avail[z], []int{-2, -1, 1, 2, 3}, itoa)
	return q
}

// ── hybrid team schedule ────────────────────────────────────────────────

var weekdays = []string{"Mon", "Tue", "Wed", "Thu", "Fri"}

func hybridGridExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	minDays := pick(r, []int{2, 3, 3})
	anchor := r.intn(5)
	n := 5
	names := make([]string, n)
	grid := make([][]string, n)
	for i := range names {
		names[i] = personName(r)
		grid[i] = make([]string, 5)
		office := sample(r, []int{0, 1, 2, 3, 4}, r.between(minDays, 4))
		if !containsInt(office, anchor) {
			office[0] = anchor
		}
		for d := 0; d < 5; d++ {
			grid[i][d] = "Remote"
			if containsInt(office, d) {
				grid[i][d] = "Office"
			}
		}
	}
	variant := r.intn(3)
	bad := r.intn(n)
	switch variant {
	case 0: // one person below the minimum, no PTO that week
		cnt := 0
		for d := 0; d < 5; d++ {
			if grid[bad][d] == "Office" {
				cnt++
				if cnt >= minDays {
					grid[bad][d] = "Remote"
				}
			}
		}
	case 1: // one person missed the anchor day without PTO
		grid[bad][anchor] = "Remote"
		if c := countOffice(grid[bad]); c < minDays {
			for d := 0; d < 5 && countOffice(grid[bad]) < minDays; d++ {
				if d != anchor {
					grid[bad][d] = "Office"
				}
			}
		}
	}
	// Someone else on PTO for a day, to make it realistic (not the "bad" row).
	pto := (bad + 1 + r.intn(n-1)) % n
	if variant == 2 {
		pto = r.intn(n)
	}
	pd := (anchor + 1 + r.intn(4)) % 5
	if grid[pto][pd] == "Office" && countOffice(grid[pto]) <= minDays {
		pd = -1
	}
	if pd >= 0 {
		grid[pto][pd] = "PTO"
	}
	head := append([]string{"Associate"}, weekdays...)
	var rows [][]string
	for i := range names {
		row := []string{names[i]}
		for d := 0; d < 5; d++ {
			v := grid[i][d]
			if v == "Office" {
				v = "\x00Office"
			}
			row = append(row, v)
		}
		rows = append(rows, row)
	}
	note := fmt.Sprintf("Hybrid Work Policy: at least %d office days per week. %s is the team anchor day; PTO days are excused.", minDays, map[int]string{0: "Monday", 1: "Tuesday", 2: "Wednesday", 3: "Thursday", 4: "Friday"}[anchor])
	q := Question{key: "ex/hybridgrid", Exhibit: docExhibit(sl["DEPT"]+" — Hybrid Schedule", "Week of "+exDate(r).Format("January 2, 2006"), head, rows, note)}
	var wrongs []string
	for i, nm := range names {
		if i != bad && len(wrongs) < 3 {
			wrongs = append(wrongs, nm)
		}
	}
	switch variant {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. Which team member did not meet the minimum number of office days this week?", ex)
		q.Choices, q.Correct = withChoices(r, names[bad], wrongs)
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. Which team member missed the anchor day?", ex)
		q.Choices, q.Correct = withChoices(r, names[bad], wrongs)
	default:
		d := r.intn(5)
		cnt := 0
		for i := range names {
			if grid[i][d] == "Office" {
				cnt++
			}
		}
		q.Text = fmt.Sprintf("Refer to %s. How many team members are in the office on %s?", ex, map[int]string{0: "Monday", 1: "Tuesday", 2: "Wednesday", 3: "Thursday", 4: "Friday"}[d])
		q.Choices, q.Correct = numChoices(r, cnt, []int{-2, -1, 1, 2}, itoa)
	}
	return q
}

func countOffice(row []string) int {
	n := 0
	for _, v := range row {
		if v == "Office" {
			n++
		}
	}
	return n
}

// ── badge swipe report (return to office) ───────────────────────────────

func badgeLogExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	minHours := pick(r, []int{4, 5, 6})
	need := pick(r, []int{3, 3, 4})
	start := exDate(r)
	for start.Weekday() != 1 {
		start = start.AddDate(0, 0, 1)
	}
	var rows [][]string
	qualifying := 0
	short := -1
	for d := 0; d < 5; d++ {
		day := start.AddDate(0, 0, d)
		if r.chance(0.3) {
			rows = append(rows, []string{day.Format("Mon 01/02"), "—", "—", "Remote"})
			continue
		}
		in := 60*r.between(7, 10) + 5*r.intn(12)
		hours := r.between(1, 9)
		if hours == minHours {
			hours++
		}
		if hours < minHours && short >= 0 {
			hours = minHours + r.between(1, 3) // at most one short visit per week
		}
		if hours < minHours && short < 0 {
			short = d
		}
		out := in + hours*60 + 5*r.intn(11)
		mins := out - in
		if mins/60 >= minHours {
			qualifying++
		}
		rows = append(rows, []string{day.Format("Mon 01/02"), clock(in), clock(out), fmt.Sprintf("%d h %02d m", mins/60, mins%60)})
	}
	q := Question{key: "ex/badgelog", Exhibit: docExhibit("Badge Activity — "+sl["NAME"], "North Lobby turnstiles · week of "+start.Format("January 2"), []string{"Date", "First swipe in", "Last swipe out", "Time on site"}, rows,
		fmt.Sprintf("Return to Office Policy: %d in-office days per week; a day counts when the associate is on site for at least %d hours.", need, minHours))}
	if short >= 0 && r.chance(0.4) {
		q.Text = fmt.Sprintf("Refer to %s. On which day did %s badge in but stay too briefly for the day to count?", ex, sl["FIRST"])
		var wrongs []string
		for d := 0; d < 5; d++ {
			if d != short && len(wrongs) < 3 {
				wrongs = append(wrongs, rows[d][0])
			}
		}
		q.Choices, q.Correct = withChoices(r, rows[short][0], wrongs)
		return q
	}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. How many qualifying in-office days did %s have this week?", ex, sl["FIRST"])
		q.Choices, q.Correct = numChoices(r, qualifying, []int{-2, -1, 1, 2}, itoa)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Did %s meet the Return to Office requirement this week?", ex, sl["FIRST"])
	yes := fmt.Sprintf("Yes; %d qualifying days", qualifying)
	no := fmt.Sprintf("No; only %d qualifying days", qualifying)
	right, wrong := no, fmt.Sprintf("Yes; %d qualifying days", qualifying)
	if qualifying >= need {
		right, wrong = yes, fmt.Sprintf("No; only %d qualifying days", qualifying)
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "It cannot be determined from badge data", "Yes; any badge swipe counts as a full day"})
	return q
}

// ── PTO balance statement ───────────────────────────────────────────────

func ptoStatementExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	startBal := r.between(0, 60)
	accrual := pick(r, []int{4, 5, 6, 8})
	periods := r.between(4, 20)
	var rows [][]string
	used := 0
	day := exDate(r)
	reasons := []string{"Vacation", "Sick", "Personal", "Vacation", "Medical appointment", "Family event"}
	nUses := r.between(2, 5)
	for i := 0; i < nUses; i++ {
		h := pick(r, []int{4, 8, 8, 16, 24, 40})
		used += h
		rows = append(rows, []string{day.AddDate(0, 0, 9*i).Format("01/02/2006"), pick(r, reasons), fmt.Sprintf("-%d h", h)})
	}
	bal := startBal + accrual*periods - used
	for bal < 0 {
		periods += 2
		bal = startBal + accrual*periods - used
	}
	cap := pick(r, []int{80, 120, 160})
	summary := fmt.Sprintf("Opening balance %d h · accrual %d h per pay period × %d pay periods · carryover cap %d h", startBal, accrual, periods, cap)
	q := Question{key: "ex/ptostatement", Exhibit: docExhibit("PTO Statement — "+sl["NAME"], summary, []string{"Date", "Type", "Hours"}, rows, "Balances are shown in hours. Requests are deducted when approved.")}
	hours := func(n int) string { return fmt.Sprintf("%d hours", n) }
	switch r.intn(3) {
	case 0:
		q.Text = fmt.Sprintf("Refer to %s. What is %s's current PTO balance?", ex, sl["FIRST"])
		q.Choices, q.Correct = numChoices(r, bal, []int{-accrual, accrual, -8, 8, used, -startBal}, hours)
	case 1:
		q.Text = fmt.Sprintf("Refer to %s. How many hours of PTO has %s used in total?", ex, sl["FIRST"])
		q.Choices, q.Correct = numChoices(r, used, []int{-8, 8, -4, 16, -16}, hours)
	default:
		more := r.between(2, 12)
		proj := bal + accrual*more
		lost := proj - cap
		if lost < 0 {
			lost = 0
		}
		q.Text = fmt.Sprintf("Refer to %s. If %s takes no more time off, %d more pay periods remain this year, and the carryover cap applies, how many hours would be forfeited at year end?", ex, sl["FIRST"], more)
		q.Choices, q.Correct = numChoices(r, lost, []int{8, 16, 24, -8, accrual}, hours)
	}
	return q
}

// ── FMLA leave timeline ─────────────────────────────────────────────────

func fmlaExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type block struct {
		label string
		weeks int
		color string
	}
	kinds := []block{{"Pregnancy-related", 0, "#c48b9f"}, {"Bonding", 0, "#5b8def"}, {"Own serious health condition", 0, "#e0a030"}, {"Care for a parent", 0, "#4caf50"}}
	var used []block
	total := 0
	chosen := sample(r, []int{0, 1, 2, 3}, r.between(1, 3))
	sortInts(chosen) // chronological: pregnancy-related before bonding
	for _, ki := range chosen {
		k := kinds[ki]
		w := r.between(1, 5)
		if total+w > 11 {
			break
		}
		k.weeks = w
		total += w
		used = append(used, k)
	}
	var b strings.Builder
	b.WriteString(`<svg class="cmp-exhibit-img" width="460" height="170" viewBox="0 0 460 170" role="img" aria-label="FMLA leave timeline"><rect width="460" height="170" fill="#fff" stroke="#c8ccd4"/>`)
	fmt.Fprintf(&b, `<text x="230" y="18" text-anchor="middle" %s font-size="10.5" font-weight="700" fill="#0a1628">FMLA LEAVE — %s · 12-MONTH PERIOD</text>`, svgFont, esc(strings.ToUpper(sl["NAME"])))
	b.WriteString(`<rect x="20" y="40" width="420" height="30" fill="#f1f3f6" stroke="#9aa5b1"/>`)
	for i := 0; i <= 52; i += 4 {
		x := 20 + float64(i)*420/52
		fmt.Fprintf(&b, `<line x1="%.1f" y1="70" x2="%.1f" y2="76" stroke="#9aa5b1"/><text x="%.1f" y="86" text-anchor="middle" %s font-size="7" fill="#666">wk %d</text>`, x, x, x, svgFont, i)
	}
	week := r.between(2, 10)
	for i, u := range used {
		x := 20 + float64(week)*420/52
		w := float64(u.weeks) * 420 / 52
		fmt.Fprintf(&b, `<rect x="%.1f" y="40" width="%.1f" height="30" fill="%s"/>`, x, w, u.color)
		fmt.Fprintf(&b, `<rect x="24" y="%d" width="10" height="10" fill="%s"/><text x="40" y="%d" %s font-size="8.5" fill="#222">%s — %s</text>`, 98+i*16, u.color, 107+i*16, svgFont, esc(u.label), plural(u.weeks, "week"))
		week += u.weeks + r.between(2, 10)
	}
	b.WriteString(`</svg>`)
	q := Question{key: "ex/fmla", Exhibit: b.String()}
	weeks := func(n int) string { return plural(n, "week") }
	if r.chance(0.6) {
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %[1]s. %[2]s is an eligible employee. How many weeks of job-protected FMLA leave remain in this 12-month period?",
			"Refer to %[1]s. %[2]s's child is born next month. How many more weeks of FMLA leave can %[2]s take in this 12-month period?",
		}), ex, sl["FIRST"])
		q.Choices, q.Correct = numChoices(r, 12-total, []int{-1, 1, 2, -2, total}, weeks)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. How many weeks of FMLA leave has %s used so far in this 12-month period?", ex, sl["FIRST"])
	q.Choices, q.Correct = numChoices(r, total, []int{-1, 1, 2, -2, 12 - 2*total}, weeks)
	return q
}

// ── manager chat during a union campaign (TIPS) ─────────────────────────

var lawfulCampaignMsgs = []string{
	"Our benefits are already among the best in Milwaukee. I'd encourage everyone to compare before deciding.",
	"You have the right to vote however you choose, and the election is by secret ballot.",
	"If you have questions about the union, I'm happy to share the facts I know.",
	"Dues are set by the union, so ask them what dues would be if they win.",
	"Whatever you decide, nobody will be treated differently because of how they vote.",
	"Reminder: the same rule applies to every non-work topic — please keep those conversations to breaks and lunch.",
	"The NLRB runs the election. I'll share the date and location as soon as they're posted.",
}

var unlawfulCampaignMsgs = []struct{ Kind, Text string }{
	{"Threaten", "If the union wins, I can't promise this office stays open."},
	{"Threaten", "Anyone who signs a card can forget about the promotion list."},
	{"Interrogate", "Quick poll: who went to the union meeting on Tuesday?"},
	{"Interrogate", "Be honest with me — are you voting yes?"},
	{"Promise", "Vote no and I'll make sure everyone gets an extra week of PTO."},
	{"Promise", "If we stay union-free, those raises you asked for are yours."},
	{"Spy", "I'll be keeping an eye on who shows up at the union hall."},
	{"Spy", "Security will be watching the break room to see who's talking union."},
}

func tipsChatExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	lawful := sample(r, lawfulCampaignMsgs, 3)
	bad := pick(r, unlawfulCampaignMsgs)
	at := r.intn(4)
	msgs := append(append(append([]string{}, lawful[:at]...), bad.Text), lawful[at:]...)
	mgr := sl["NAME"]
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="cmp-ex-mail"><div class="cmp-ex-mailbar">💬 #%s-team · messages from %s (people manager)</div><div class="cmp-ex-mailbody">`, esc(strings.ToLower(strings.ReplaceAll(sl["DEPT"], " ", "-"))), esc(mgr))
	t := r.between(9*60, 16*60)
	for i, m := range msgs {
		t += r.between(2, 25)
		fmt.Fprintf(&b, `<p><b>Message %d</b> <span class="cmp-muted">%s</span><br>%s</p>`, i+1, clock(t), esc(m))
	}
	b.WriteString(`</div></div>`)
	q := Question{key: "ex/tipschat", Exhibit: b.String()}
	if r.chance(0.6) {
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %[1]s. During a union organizing campaign, which message from %[2]s violates the National Labor Relations Act?",
			"Refer to %[1]s. Which message crosses the line under the TIPS rules for managers?",
		}), ex, firstOf(mgr))
		var wrongs []string
		for i := range msgs {
			if i != at {
				wrongs = append(wrongs, fmt.Sprintf("Message %d", i+1))
			}
		}
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Message %d", at+1), wrongs)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. Message %d is an example of which prohibited manager behavior?", ex, at+1)
	var wrongs []string
	for _, k := range []string{"Threaten", "Interrogate", "Promise", "Spy"} {
		if k != bad.Kind {
			wrongs = append(wrongs, k)
		}
	}
	q.Choices, q.Correct = withChoices(r, bad.Kind, wrongs)
	return q
}
