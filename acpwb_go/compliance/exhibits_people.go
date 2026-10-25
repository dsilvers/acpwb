package compliance

// People and workplace exhibits: onboarding and leave deadlines (I-9's 3
// business days, USERRA's 14/90-day windows), surveys, schedules, board
// quorum, projects, meetings, perks, celebrations, sustainability, and a
// few items from the ACPWB Archives (Y2K, fax, floppy disks).

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func init() {
	for k, g := range map[string]exhibitGen{
		"i9":           i9Exhibit,
		"usrra":        usrraExhibit,
		"lifeevent":    lifeEventExhibit,
		"surveygroups": surveyGroupsExhibit,
		"restgap":      restGapExhibit,
		"mealpunch":    mealPunchExhibit,
		"dresscal":     dressCalendarExhibit,
		"quorum":       quorumExhibit,
		"gantt":        ganttExhibit,
		"kanban":       kanbanExhibit,
		"okr":          okrExhibit,
		"agenda":       agendaExhibit,
		"meetingcost":  meetingCostExhibit,
		"replyall":     replyAllExhibit,
		"roomscreen":   roomScreenExhibit,
		"evcharge":     evChargeExhibit,
		"dogboard":     dogBoardExhibit,
		"santatags":    santaTagsExhibit,
		"rsvp":         rsvpExhibit,
		"emissions":    emissionsExhibit,
		"recycling":    recyclingExhibit,
		"tradeagenda":  tradeAgendaExhibit,
		"ndaterm":      ndaTermExhibit,
		"y2k":          y2kExhibit,
		"faxcover":     faxCoverExhibit,
		"floppy":       floppyExhibit,
		"parking":      parkingExhibit,
		"orgchart":     orgChartExhibit,
		"napbook":      napBookExhibit,
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
		"onboarding":     {"i9"},
		"hiring":         {"i9"},
		"civicleave":     {"usrra"},
		"leave":          {"usrra"},
		"enrollment":     {"lifeevent"},
		"parental":       {"lifeevent"},
		"recognition":    {"surveygroups"},
		"culture":        {"surveygroups", "replyall", "meetingcost"},
		"nightshift":     {"restgap"},
		"wage":           {"mealpunch"},
		"timeoff":        {"mealpunch", "restgap"},
		"attendance":     {"mealpunch"},
		"dresscode":      {"dresscal"},
		"costumes":       {"dresscal"},
		"board":          {"quorum"},
		"governance":     {"quorum", "orgchart"},
		"agile":          {"gantt", "kanban"},
		"goals":          {"okr"},
		"performance":    {"okr"},
		"meetings":       {"agenda", "meetingcost", "roomscreen"},
		"calendar":       {"agenda", "roomscreen"},
		"travelbooking":  {"roomscreen"},
		"hotdesk":        {"roomscreen"},
		"commuting":      {"evcharge", "parking"},
		"parking":        {"parking", "evcharge"},
		"futurework":     {"napbook"},
		"wellness":       {"napbook"},
		"pets":           {"dogboard"},
		"holiday":        {"santatags", "rsvp"},
		"celebrations":   {"rsvp", "santatags"},
		"gifts":          {"santatags"},
		"offsites":       {"rsvp", "tradeagenda"},
		"climate":        {"emissions"},
		"esg":            {"emissions"},
		"sustainability": {"recycling"},
		"antitrust":      {"tradeagenda"},
		"ip":             {"ndaterm"},
		"contracts":      {"ndaterm"},
		"retro":          {"y2k", "faxcover", "floppy"},
		"records":        {"faxcover"},
		"managers":       {"orgchart"},
		"authority":      {"orgchart"},
		"accessibility":  {"parking"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// monthGrid draws a month calendar with some days marked; marks maps day
// → (fill, label).
func monthGrid(s *svgB, x, y float64, month time.Time, marks map[int][2]string) {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	days := first.AddDate(0, 1, -1).Day()
	s.text(x+3.5*44, y, 12, `text-anchor="middle" font-weight="700" fill="#1a3a5c"`, first.Format("January 2006"))
	for i, d := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		s.text(x+float64(i)*44+22, y+18, 9, `text-anchor="middle" fill="#666" font-weight="700"`, d)
	}
	col := int(first.Weekday())
	row := 0
	for d := 1; d <= days; d++ {
		cx := x + float64(col)*44
		cy := y + 26 + float64(row)*30
		fill := "#fff"
		if m, ok := marks[d]; ok {
			fill = m[0]
		} else if col == 0 || col == 6 {
			fill = "#f1f1f1"
		}
		s.rect(cx+1, cy, 42, 28, 2, fill, "#d6dbe1")
		s.text(cx+5, cy+11, 9, "", fmt.Sprint(d))
		if m, ok := marks[d]; ok && m[1] != "" {
			s.text(cx+22, cy+24, 7, `text-anchor="middle" font-weight="700"`, m[1])
		}
		col++
		if col == 7 {
			col, row = 0, row+1
		}
	}
}

// ── Form I-9 deadline ──────────────────────────────────────────────────

func i9Exhibit(r *rnd, c *Course, sl slots, ex string) Question {
	month := time.Date(r.between(2019, 2026), time.Month(r.between(1, 12)), 1, 0, 0, 0, 0, time.UTC)
	start := weekday(month.AddDate(0, 0, r.between(0, 15)))
	due := addBusinessDays(start, 3)
	for due.Month() != start.Month() { // keep it on one calendar page
		start = start.AddDate(0, 0, -7)
		due = addBusinessDays(start, 3)
	}
	s := newSVG(330, 250, "Calendar with a first day of work")
	monthGrid(s, 10, 20, month, map[int][2]string{start.Day(): {"#c8e6c9", "START"}})
	q := Question{key: "ex/i9", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. %s starts work on the day marked START, and ACPWB is open Monday through Friday. By what date must ACPWB complete Section 2 of Form I-9?", ex, sl["NAME"])
	q.Choices, q.Correct = withChoices(r, longDate(due), []string{longDate(start.AddDate(0, 0, 3)), longDate(addBusinessDays(start, 10)), longDate(start)})
	if due.Equal(start.AddDate(0, 0, 3)) {
		q.Choices, q.Correct = withChoices(r, longDate(due), []string{longDate(addBusinessDays(start, 5)), longDate(addBusinessDays(start, 10)), longDate(start)})
	}
	return q
}

// ── USERRA return windows ──────────────────────────────────────────────

func usrraExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	days := pick(r, []int{r.between(31, 180), r.between(181, 900)})
	end := exDate(r)
	beg := end.AddDate(0, 0, -days)
	window := 14
	if days > 180 {
		window = 90
	}
	rows := [][]string{
		{"Service member", sl["NAME"]},
		{"Branch", pick(r, []string{"Wisconsin Army National Guard", "U.S. Navy Reserve", "U.S. Air Force Reserve", "U.S. Marine Corps Reserve"})},
		{"Service began", longDate(beg)},
		{"Service completed", longDate(end)},
		{"Length of service", fmt.Sprintf("%d days", days)},
	}
	q := Question{key: "ex/usrra", Exhibit: docExhibit("Military Leave Record", "USERRA · reemployment", []string{"Item", "Detail"}, rows, "Service of 31–180 days: apply within 14 days of completing service. Over 180 days: within 90 days.")}
	q.Text = fmt.Sprintf("Refer to %s. By what date must %s apply for reemployment to keep USERRA protection?", ex, sl["FIRST"])
	q.Choices, q.Correct = dateQ(r, end.AddDate(0, 0, window), []int{-window + 14 + 76, -window + 30, -window + 7, -window + 180, -window + 1})
	return q
}

// ── qualifying life event window ───────────────────────────────────────

func lifeEventExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	ev := pick(r, []string{"Birth of a child", "Marriage", "Adoption", "Spouse lost other coverage", "Divorce"})
	window := pick(r, []int{30, 31, 60})
	evDate := exDate(r)
	late := r.chance(0.5)
	delay := r.between(3, window-1)
	if late {
		delay = window + r.between(1, 20)
	}
	req := evDate.AddDate(0, 0, delay)
	rows := [][]string{{"Event", ev}, {"Event date", longDate(evDate)}, {"Change requested", longDate(req)}, {"Plan deadline", fmt.Sprintf("Within %d days of the event", window)}}
	q := Question{key: "ex/lifeevent", Exhibit: docExhibit("Benefits Change Request — "+sl["NAME"], "Mid-year election change", []string{"Item", "Detail"}, rows, "Requests after the deadline wait until open enrollment.")}
	q.Text = fmt.Sprintf("Refer to %s. Can %s's mid-year change be processed now?", ex, sl["FIRST"])
	yes := fmt.Sprintf("Yes; it was requested %d days after the event, within the %d-day window", delay, window)
	no := fmt.Sprintf("No; it was requested %d days after the event, outside the %d-day window, so it waits for open enrollment", delay, window)
	right, wrong := yes, no
	if late {
		right, wrong = no, yes
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Only with the manager's approval", "Only if the employee resubmits on paper"})
	return q
}

// ── survey minimum group size ──────────────────────────────────────────

func surveyGroupsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	minN := pick(r, []int{5, 7, 10})
	teams := sample(r, departments, 6)
	shown := 0
	var rows [][]string
	for _, t := range teams {
		n := r.between(1, 18)
		if n >= minN {
			shown++
		}
		rows = append(rows, []string{t, fmt.Sprint(n), fmt.Sprintf("%.1f", 2.5+float64(r.between(0, 24))/10)})
	}
	q := Question{key: "ex/surveygroups", Exhibit: docExhibit("Engagement Survey — Response Counts", fmt.Sprintf("Results are shown only for groups with at least %d responses", minN), []string{"Team", "Responses", "Engagement score"}, rows, "Smaller groups roll up to the next level to protect anonymity.")}
	q.Text = fmt.Sprintf("Refer to %s. For how many of these teams can results be shown to the team's manager?", ex)
	q.Choices, q.Correct = countQ(r, shown, "team")
	return q
}

// ── rest between shifts ────────────────────────────────────────────────

func restGapExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	minRest := pick(r, []int{8, 10, 11})
	names := people(r, 4)
	bad := r.intn(4)
	var rows [][]string
	for i, n := range names {
		end1 := r.between(17, 23)*60 + pick(r, []int{0, 30})
		gap := minRest*60 + r.between(0, 6)*60
		if i == bad {
			gap = minRest*60 - r.between(1, 4)*60
		}
		start2 := (end1 + gap) % (24 * 60)
		rows = append(rows, []string{n, "Mon " + clock(end1%(24*60)), "Tue " + clock(start2)})
	}
	q := Question{key: "ex/restgap", Exhibit: docExhibit("Shift Schedule — Overnight Operations", fmt.Sprintf("Policy: at least %d hours off between the end of one shift and the start of the next", minRest), []string{"Employee", "Shift ends", "Next shift starts"}, rows, "Swap requests that break the rest rule are declined.")}
	q.Text = fmt.Sprintf("Refer to %s. Whose schedule breaks the rest rule?", ex)
	q.Choices, q.Correct = oneOf(r, names[bad], append(names, "No one"))
	return q
}

// ── time punches and meal break ────────────────────────────────────────

func mealPunchExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	in := r.between(6, 9)*60 + pick(r, []int{0, 15, 30, 45})
	out1 := in + r.between(3, 5)*60 + pick(r, []int{0, 15, 30})
	brk := pick(r, []int{20, 30, 45, 60})
	in2 := out1 + brk
	out2 := in2 + r.between(3, 5)*60 + pick(r, []int{0, 15, 30, 45})
	worked := (out1 - in) + (out2 - in2)
	paid := worked
	if brk < 30 { // a short break is paid
		paid += brk
	}
	rows := [][]string{{"In", clock(in)}, {"Out (meal)", clock(out1)}, {"In", clock(in2)}, {"Out", clock(out2)}}
	q := Question{key: "ex/mealpunch", Exhibit: docExhibit("Time Punches — "+sl["NAME"], "Non-exempt · "+exDate(r).Format("Mon Jan 2"), []string{"Punch", "Time"}, rows, "Meal breaks of 30 minutes or more, fully relieved of duty, are unpaid. Shorter breaks are paid.")}
	f := func(m int) string { return fmt.Sprintf("%d hours %d minutes", m/60, m%60) }
	q.Text = fmt.Sprintf("Refer to %s. How much paid time should %s's timesheet show for this day?", ex, sl["FIRST"])
	q.Choices, q.Correct = numChoices(r, paid, []int{brk, -brk, 30, -15, 60}, f)
	return q
}

// ── dress code calendar ────────────────────────────────────────────────

func dressCalendarExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	codes := map[string]string{}
	client := r.intn(5)
	for i, d := range days {
		codes[d] = "Business casual"
		if i == 4 {
			codes[d] = "Casual Friday (jeans OK)"
		}
	}
	theme := r.intn(4)
	codes[days[theme]] = pick(r, []string{"Wear Your Team Jersey Day", "Packers Spirit Day", "PERCH Conference Polo Day"})
	codes[days[client]] = "Client visit: business professional"
	s := newSVG(460, 140, "Weekly dress code calendar")
	s.rect(0, 0, 460, 140, 8, "#fff", "#ccc")
	for i, d := range days {
		x := 10 + float64(i)*89
		col := "#e3f2fd"
		if strings.HasPrefix(codes[d], "Client") {
			col = "#fff3e0"
		}
		s.rect(x, 14, 84, 112, 6, col, "#90a4ae")
		s.text(x+42, 34, 11, `text-anchor="middle" font-weight="700" fill="#1a3a5c"`, d)
		for j, ln := range wrapText(codes[d], 13) {
			s.text(x+42, 60+float64(j)*14, 9, `text-anchor="middle"`, ln)
		}
	}
	q := Question{key: "ex/dresscal", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. What should %s wear on %s?", ex, sl["FIRST"], days[client])
	q.Choices, q.Correct = withChoices(r, "Business professional, since a client is visiting", []string{"Jeans, since it is casual week", "A team jersey", "A Halloween costume"})
	return q
}

// ── board quorum ───────────────────────────────────────────────────────

func quorumExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	seats := pick(r, []int{7, 9, 11, 12})
	vacant := r.between(0, 1)
	inOffice := seats - vacant
	need := inOffice/2 + 1
	names := people(r, inOffice)
	present := 0
	var rows [][]string
	for _, n := range names {
		st := pick(r, []string{"Present", "Present", "Remote (video)", "Absent", "Absent"})
		if st != "Absent" {
			present++
		}
		rows = append(rows, []string{n, st})
	}
	for i := 0; i < vacant; i++ {
		rows = append(rows, []string{"(vacant seat)", "—"})
	}
	q := Question{key: "ex/quorum", Exhibit: docExhibit("Board Attendance — "+sl["COMMITTEE"], fmt.Sprintf("%d seats · bylaws: quorum is a majority of directors in office; video attendance counts", seats), []string{"Director", "Attendance"}, rows, "The secretary records quorum before any vote.")}
	if present >= need {
		q.Text = fmt.Sprintf("Refer to %s. Is a quorum present?", ex)
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Yes; %d of %d directors in office are attending, and %d are needed", present, inOffice, need), []string{"No; remote directors do not count", fmt.Sprintf("No; all %d seats must be filled", seats), "No; a quorum requires every director"})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. How many more directors must join before the board can vote?", ex)
	q.Choices, q.Correct = countQ(r, need-present, "director")
	return q
}

// ── Gantt chart ────────────────────────────────────────────────────────

func ganttExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type task struct {
		name string
		dur  int
		deps []int
	}
	names := sample(r, []string{"Requirements", "Vendor selection", "Data migration", "Policy draft", "Legal review", "Training build", "Pilot", "Rollout", "Survey design", "Board approval"}, 5)
	tasks := []task{
		{names[0], r.between(1, 3), nil},
		{names[1], r.between(1, 4), []int{0}},
		{names[2], r.between(2, 5), []int{0}},
		{names[3], r.between(1, 3), []int{1, 2}},
		{names[4], r.between(1, 3), []int{3}},
	}
	start := make([]int, len(tasks))
	end := make([]int, len(tasks))
	for i, t := range tasks {
		for _, d := range t.deps {
			if end[d] > start[i] {
				start[i] = end[d]
			}
		}
		end[i] = start[i] + t.dur
	}
	finish := end[len(end)-1]
	s := newSVG(460, 60+float64(len(tasks))*34+30, "Project Gantt chart")
	s.rect(0, 0, 460, 60+float64(len(tasks))*34+30, 8, "#fff", "#ccc")
	s.text(14, 22, 12, `font-weight="700" fill="#1a3a5c"`, "Project "+pick(r, []string{"Vertex", "Flywheel", "Heimdall", "Legato", "Crown"})+" · weeks")
	x0, wk := 200.0, 246.0/float64(finish+1)
	for w := 0; w <= finish; w++ {
		s.line(x0+float64(w)*wk, 34, x0+float64(w)*wk, 50+float64(len(tasks))*34, "#eee", 1)
		s.text(x0+float64(w)*wk+wk/2, 44, 9, `text-anchor="middle" fill="#666"`, fmt.Sprint(w+1))
	}
	for i, t := range tasks {
		y := 54 + float64(i)*34
		s.text(14, y+12, 10, `font-weight="700"`, t.name)
		s.rect(x0+float64(start[i])*wk, y+4, float64(t.dur)*wk-2, 18, 4, "#5b8def", "")
		s.text(x0+float64(start[i])*wk+float64(t.dur)*wk/2, y+17, 8, `text-anchor="middle" fill="#fff" font-weight="700"`, plural(t.dur, "wk"))
		var deps []string
		for _, d := range t.deps {
			deps = append(deps, tasks[d].name)
		}
		if len(deps) > 0 {
			s.text(14, y+25, 8, `fill="#888"`, "after "+strings.Join(deps, " + "))
		}
	}
	q := Question{key: "ex/gantt", Exhibit: s.done()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Each task starts when all the tasks it follows are done. How many weeks does the whole project take?", ex)
		total := 0
		for _, t := range tasks {
			total += t.dur
		}
		q.Choices, q.Correct = numChoices(r, finish, []int{total - finish, 1, -1, 2}, func(n int) string { return plural(n, "week") })
		return q
	}
	k := pick(r, []int{1, 2})
	q.Text = fmt.Sprintf("Refer to %s. If \"%s\" slips by one week, does the project finish date move?", ex, tasks[k].name)
	yes := "Yes; it is on the critical path, so the finish moves by one week"
	no := "No; it has slack, so the finish date stays the same"
	right, wrong := yes, no
	if end[k] < end[3-k] {
		right, wrong = no, yes
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Yes, by two weeks", "No; only the first task can affect the finish date"})
	return q
}

// ── Kanban WIP limits ──────────────────────────────────────────────────

func kanbanExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	cols := []string{"To do", "In progress", "Review", "Blocked", "Done"}
	limits := []int{0, r.between(3, 5), r.between(2, 3), r.between(1, 2), 0}
	over := r.between(1, 3)
	s := newSVG(470, 260, "Kanban board")
	s.rect(0, 0, 470, 260, 8, "#eef1f5", "")
	for i, cname := range cols {
		x := 8 + float64(i)*92
		n := r.between(1, 6)
		if limits[i] > 0 {
			n = r.between(1, limits[i])
			if i == over {
				n = limits[i] + r.between(1, 2)
			}
		}
		head := cname
		if limits[i] > 0 {
			head = fmt.Sprintf("%s (WIP %d)", cname, limits[i])
		}
		s.rect(x, 10, 86, 240, 6, "#dfe5ec", "")
		s.text(x+43, 28, 9, `text-anchor="middle" font-weight="700"`, head)
		for k := 0; k < n; k++ {
			s.rect(x+6, 38+float64(k)*34, 74, 28, 4, "#fff", "#c5ced8")
			s.text(x+12, 56+float64(k)*34, 8, "", fmt.Sprintf("ACP-%d", r.between(100, 999)))
		}
	}
	q := Question{key: "ex/kanban", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which column is over its work-in-progress limit?", ex)
	q.Choices, q.Correct = oneOf(r, cols[over], []string{cols[0], cols[1], cols[2], cols[3], "None of them"})
	return q
}

// ── OKR scoring ────────────────────────────────────────────────────────

func okrExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	krs := sample(r, []string{"Cut audit findings from 12 to 4", "Train 95% of managers on TIPS", "Close 30 open policy exceptions", "Raise survey participation to 80%", "Publish the pay equity report by Q3", "Reduce badge tailgating by half"}, r.between(3, 4))
	var rows [][]string
	sum := 0
	for _, k := range krs {
		v := r.between(0, 10)
		sum += v
		rows = append(rows, []string{k, fmt.Sprintf("%.1f", float64(v)/10)})
	}
	avg := int(math.Round(float64(sum) / float64(len(krs))))
	q := Question{key: "ex/okr", Exhibit: docExhibit("Q"+fmt.Sprint(r.between(1, 4))+" OKR — "+sl["DEPT"], "Objective: "+pick(r, []string{"Make compliance boring again", "Build a speak-up culture", "Earn the auditors' trust"}), []string{"Key result", "Score (0.0–1.0)"}, rows, "The objective's score is the average of its key results, rounded to one decimal.")}
	q.Text = fmt.Sprintf("Refer to %s. What is the objective's score?", ex)
	q.Choices, q.Correct = numChoices(r, avg, []int{1, -1, 2, -2, 3}, func(n int) string { return fmt.Sprintf("%.1f", float64(n)/10) })
	return q
}

// ── agenda timing ──────────────────────────────────────────────────────

func agendaExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	start := r.between(8, 15)*60 + pick(r, []int{0, 30})
	items := sample(r, []string{"Welcome and safety moment", "Q3 results", "Policy updates", "Pay transparency rollout", "Open discussion", "PERCH Conference planning", "Audit findings", "Next steps"}, r.between(4, 6))
	var rows [][]string
	t := start
	var starts []int
	for _, it := range items {
		d := pick(r, []int{5, 10, 15, 20, 25, 30, 45})
		starts = append(starts, t)
		rows = append(rows, []string{it, fmt.Sprintf("%d min", d)})
		t += d
	}
	q := Question{key: "ex/agenda", Exhibit: docExhibit("Agenda — "+sl["COMMITTEE"], "Starts "+clock(start), []string{"Item", "Time"}, rows, "Items run back to back with no breaks.")}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. At what time does the meeting end?", ex)
		q.Choices, q.Correct = numChoices(r, t, []int{-15, 15, 30, -30, 10}, clock)
		return q
	}
	k := r.between(1, len(items)-1)
	q.Text = fmt.Sprintf("Refer to %s. When does \"%s\" start?", ex, items[k])
	q.Choices, q.Correct = numChoices(r, starts[k], []int{-15, 15, 10, -10, 30}, clock)
	return q
}

// ── cost of a meeting ──────────────────────────────────────────────────

func meetingCostExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	roles := []struct {
		role string
		rate int
	}{{"Vice President", 150}, {"Director", 110}, {"Manager", 75}, {"Analyst", 45}, {"Coordinator", 35}}
	mins := pick(r, []int{30, 45, 60, 90})
	var rows [][]string
	perHour := 0
	for _, ro := range sample(r, roles, r.between(3, 5)) {
		n := r.between(1, 6)
		perHour += n * ro.rate
		rows = append(rows, []string{ro.role, fmt.Sprint(n), fmt.Sprintf("$%d", ro.rate)})
	}
	cost := perHour * mins / 60
	q := Question{key: "ex/meetingcost", Exhibit: docExhibit("Recurring Meeting: \"Quick Sync\"", fmt.Sprintf("%d minutes · weekly", mins), []string{"Role", "Attendees", "Loaded cost per hour"}, rows, "Could this meeting have been an email?")}
	q.Text = fmt.Sprintf("Refer to %s. What does one occurrence of this meeting cost in attendee time?", ex)
	q.Choices, q.Correct = dollarQ(r, cost, []int{perHour - cost, cost / 2, -cost / 3, 100, 250})
	return q
}

// ── reply-all ──────────────────────────────────────────────────────────

func replyAllExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	toN, ccN := r.between(2, 9), r.between(0, 40)
	dl := r.chance(0.4)
	s := newSVG(440, 230, "Email header")
	s.rect(0, 0, 440, 230, 6, "#fff", "#bbb")
	s.rect(0, 0, 440, 30, 6, "#e8eaed", "")
	s.text(14, 20, 11, `font-weight="700"`, "RE: RE: RE: Who took my stapler")
	for i, kv := range [][2]string{{"From", sl["NAME2"]}, {"To", fmt.Sprintf("You, and %d others", toN-1)}, {"Cc", fmt.Sprintf("%d people", ccN)}} {
		y := 54 + float64(i)*24
		s.text(14, y, 10, `fill="#666"`, kv[0]+":")
		s.text(60, y, 10, "", kv[1])
	}
	if dl {
		s.text(60, 126, 10, `fill="#c62828"`, "+ all-acpwb-milwaukee@acpwb.com (4,212 members)")
	}
	s.text(14, 160, 10, "", "Please stop replying all.")
	s.rect(14, 180, 70, 26, 4, "#1a3a5c", "")
	s.text(49, 197, 10, `text-anchor="middle" fill="#fff" font-weight="700"`, "Reply")
	s.rect(94, 180, 80, 26, 4, "#c62828", "")
	s.text(134, 197, 10, `text-anchor="middle" fill="#fff" font-weight="700"`, "Reply all")
	q := Question{key: "ex/replyall", Exhibit: s.done()}
	if dl {
		q.Text = fmt.Sprintf("Refer to %s. What is the best way for %s to respond?", ex, sl["FIRST"])
		q.Choices, q.Correct = withChoices(r, "Reply only to the sender, or not at all; never reply all to a company-wide list", []string{"Reply all asking everyone to stop replying all", "Reply all with a GIF", "Reply all to ask to be removed"})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. If %s clicks Reply all, how many people receive the reply (not counting %s)?", ex, sl["FIRST"], sl["FIRST"])
	q.Choices, q.Correct = countQ(r, 1+(toN-1)+ccN, "person")
	for i := range q.Choices {
		q.Choices[i] = strings.Replace(q.Choices[i], "persons", "people", 1)
	}
	return q
}

// ── meeting room screen ────────────────────────────────────────────────

func roomScreenExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	now := r.between(9, 16)*60 + pick(r, []int{0, 10, 20, 40, 50})
	free := r.chance(0.5)
	var nextAt int
	if free {
		nextAt = now + pick(r, []int{20, 30, 45, 60, 90})
	} else {
		nextAt = now - pick(r, []int{10, 20})
	}
	need := pick(r, []int{15, 30, 45, 60})
	room := pick(r, []string{"Huddle 33-A", "Lakeview", "Juneau", "Kilbourn", "Brady"})
	s := newSVG(320, 220, "Meeting room booking screen")
	col := "#2e7d32"
	status := "AVAILABLE"
	if !free {
		col, status = "#c62828", "IN USE"
	}
	s.rect(0, 0, 320, 220, 14, "#222", "")
	s.rect(12, 12, 296, 196, 8, col, "")
	s.text(160, 46, 14, `text-anchor="middle" fill="#fff" font-weight="700"`, room)
	s.text(160, 92, 28, `text-anchor="middle" fill="#fff" font-weight="900"`, status)
	s.text(160, 122, 12, `text-anchor="middle" fill="#fff"`, "Now "+clock(now))
	if free {
		s.text(160, 156, 11, `text-anchor="middle" fill="#fff"`, "Next booking: "+clock(nextAt)+" — "+sl["COMMITTEE"])
	} else {
		s.text(160, 156, 11, `text-anchor="middle" fill="#fff"`, "Booked since "+clock(nextAt)+" until "+clock(nextAt+60))
	}
	q := Question{key: "ex/roomscreen", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. %s needs a room right now for a %d-minute call. Can they use this one without bumping a booking?", ex, sl["FIRST"], need)
	fits := free && now+need <= nextAt
	yes := "Yes; it is free long enough"
	no := "No; it is not free for the full call"
	right, wrong := yes, no
	if !fits {
		right, wrong = no, yes
	}
	q.Choices, q.Correct = withChoices(r, right, []string{wrong, "Yes, if they knock first", "Only if they are senior to the next organizer"})
	return q
}

// ── EV charging log ────────────────────────────────────────────────────

func evChargeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{3, 4})
	names := people(r, 4)
	bad := r.intn(4)
	var rows [][]string
	for i, n := range names {
		in := r.between(7, 12)*60 + pick(r, []int{0, 15, 30, 45})
		dur := r.between(60, limit*60)
		if i == bad {
			dur = limit*60 + r.between(20, 150)
		}
		rows = append(rows, []string{n, fmt.Sprintf("EV-%d", i+1), clock(in), clock(in + dur)})
	}
	q := Question{key: "ex/evcharge", Exhibit: docExhibit("EV Charging Log — Garage Level 1", fmt.Sprintf("Limit: %d hours per session, then move your car", limit), []string{"Driver", "Charger", "Plugged in", "Unplugged"}, rows, "Repeated overstays suspend charging privileges.")}
	q.Text = fmt.Sprintf("Refer to %s. Who exceeded the charging time limit?", ex)
	q.Choices, q.Correct = oneOf(r, names[bad], append(names, "No one"))
	return q
}

// ── nap pod bookings ───────────────────────────────────────────────────

func napBookExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{20, 25, 30})
	names := people(r, 4)
	bad := r.intn(4)
	s := phone(280, "Nap pod booking app", "NapPod · "+sl["FLOOR"]+" floor")
	for i, n := range names {
		start := r.between(12, 15)*60 + pick(r, []int{0, 10, 20, 30, 40, 50})
		dur := r.between(10, limit)
		if i == bad {
			dur = limit + r.between(10, 70)
		}
		y := 80 + float64(i)*52
		s.rect(24, y, 212, 44, 8, "#fff", "#d0d0d8")
		s.text(34, y+18, 10, `font-weight="700"`, n)
		s.text(34, y+34, 9, `fill="#666"`, clock(start)+" – "+clock(start+dur))
		s.text(226, y+26, 18, `text-anchor="end"`, "💤")
	}
	s.text(130, 316, 9, `text-anchor="middle" fill="#666"`, fmt.Sprintf("Max session: %d minutes", limit))
	q := Question{key: "ex/napbook", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Whose nap exceeded the maximum session length?", ex)
	q.Choices, q.Correct = oneOf(r, names[bad], append(names, "No one"))
	return q
}

// ── Bring Your Dog to Work Day board ───────────────────────────────────

func dogBoardExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	event := exDate(r)
	dogs := sample(r, []string{"Biscuit", "Captain", "Pretzel", "Juneau", "Bratwurst", "Kringle", "Lambeau", "Pickles", "Waffles", "Dorothy"}, 4)
	bad := r.intn(4)
	s := newSVG(460, 250, "Dog registration board")
	s.rect(0, 0, 460, 250, 8, "#c69c6d", "")
	s.rect(10, 10, 440, 230, 4, "#e8d3b0", "")
	s.text(230, 34, 13, `text-anchor="middle" font-weight="700" fill="#5d4037"`, "BRING YOUR DOG TO WORK DAY · "+strings.ToUpper(shortDate(event)))
	var names []string
	for i, d := range dogs {
		x := 20 + float64(i)*108
		exp := event.AddDate(0, r.between(1, 14), 0)
		if i == bad {
			exp = event.AddDate(0, 0, -r.between(3, 120))
		}
		s.rect(x, 50, 100, 176, 3, "#fffdf5", "#bbb")
		s.circle(x+50, 50, 4, "#c62828", "")
		s.circle(x+50, 92, 24, "#d7b98e", "")
		s.circle(x+40, 86, 3, "#333", "")
		s.circle(x+60, 86, 3, "#333", "")
		s.raw(`<ellipse cx="%.0f" cy="98" rx="5" ry="4" fill="#333"/>`, x+50)
		s.raw(`<ellipse cx="%.0f" cy="78" rx="8" ry="14" fill="#a1887f" transform="rotate(-25 %.0f 78)"/>`, x+30, x+30)
		s.raw(`<ellipse cx="%.0f" cy="78" rx="8" ry="14" fill="#a1887f" transform="rotate(25 %.0f 78)"/>`, x+70, x+70)
		s.text(x+50, 136, 12, `text-anchor="middle" font-weight="700"`, d)
		s.text(x+50, 154, 8, `text-anchor="middle" fill="#666"`, "Rabies vaccine")
		s.text(x+50, 168, 8, `text-anchor="middle" fill="#666"`, "valid through")
		s.text(x+50, 184, 10, `text-anchor="middle" font-weight="700"`, shortDate(exp))
		s.text(x+50, 208, 8, `text-anchor="middle" fill="#666"`, "Owner: "+firstOf(personName(r)))
		names = append(names, d)
	}
	q := Question{key: "ex/dogboard", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Dogs must have a rabies vaccination that is current on the day of the event. Which dog may not attend?", ex)
	q.Choices, q.Correct = oneOf(r, names[bad], append(names, "None; all may attend"))
	return q
}

// ── Secret Santa price tags ────────────────────────────────────────────

func santaTagsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	limit := pick(r, []int{20, 25, 30})
	gifts := sample(r, []string{"Desk plant", "Coffee mug", "Puzzle", "Candle (unlit, decorative)", "Hot sauce set", "Bluetooth speaker", "Notebook", "Kringle", "Socks", "Board game"}, 4)
	bad := r.intn(4)
	s := newSVG(460, 210, "Gift price tags")
	s.rect(0, 0, 460, 210, 8, "#1b5e20", "")
	s.text(230, 26, 13, `text-anchor="middle" font-weight="700" fill="#fff"`, fmt.Sprintf("SECRET SANTA · LIMIT $%d", limit))
	for i, g := range gifts {
		p := 100*r.between(8, limit-1) + pick(r, []int{0, 49, 99})
		if p > limit*100 {
			p = limit*100 - 1
		}
		if i == bad {
			p = 100*r.between(limit+1, limit+30) + pick(r, []int{0, 99})
		}
		x := 20 + float64(i)*110
		s.rect(x, 50, 96, 140, 6, "#c62828", "")
		s.rect(x+30, 40, 36, 14, 3, "#ffd54f", "")
		s.rect(x+10, 100, 76, 52, 4, "#fff", "#bbb")
		s.text(x+48, 120, 10, `text-anchor="middle" font-weight="700"`, wrapText(g, 12)[0])
		s.text(x+48, 142, 14, `text-anchor="middle" font-weight="900" fill="#c62828"`, money(p))
	}
	q := Question{key: "ex/santatags", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which gift is over the Secret Santa price limit?", ex)
	var labels []string
	for _, g := range gifts {
		labels = append(labels, "The "+strings.ToLower(wrapText(g, 12)[0]))
	}
	q.Choices, q.Correct = oneOf(r, labels[bad], append(labels, "None of them"))
	return q
}

// ── RSVP list ──────────────────────────────────────────────────────────

func rsvpExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	n := r.between(7, 11)
	names := people(r, n)
	diet := pick(r, []string{"Vegetarian", "Gluten-free", "Vegan", "Halal", "Kosher"})
	want := 0
	var rows [][]string
	for _, nm := range names {
		resp := pick(r, []string{"Yes", "Yes", "Yes", "No", "Maybe"})
		d := pick(r, []string{"—", "—", diet, "Vegetarian", "Gluten-free", "Nut allergy"})
		if resp == "Yes" && d == diet {
			want++
		}
		rows = append(rows, []string{nm, resp, d})
	}
	q := Question{key: "ex/rsvp", Exhibit: docExhibit("RSVPs — "+capFirst(sl["EVENT"]), "Catering by "+sl["VENDOR"], []string{"Guest", "Attending", "Dietary needs"}, rows, "Order meals for \"Yes\" responses only.")}
	q.Text = fmt.Sprintf("Refer to %s. How many %s meals should be ordered?", ex, strings.ToLower(diet))
	q.Choices, q.Correct = countQ(r, want, "meal")
	return q
}

// ── emissions by scope ─────────────────────────────────────────────────

func emissionsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type src struct {
		name  string
		scope int
	}
	srcs := sample(r, []src{{"Natural gas for headquarters boilers", 1}, {"Fleet vehicle fuel", 1}, {"Backup generator diesel", 1}, {"Purchased electricity", 2}, {"Purchased district steam", 2}, {"Business air travel", 3}, {"Employee commuting", 3}, {"Purchased goods and services", 3}, {"Hotel stays", 3}, {"Waste sent to landfill", 3}}, 6)
	totals := map[int]int{}
	var rows [][]string
	for _, s := range srcs {
		v := 10 * r.between(5, 400)
		totals[s.scope] += v
		rows = append(rows, []string{s.name, commas(v)})
	}
	scope := pick(r, []int{1, 2, 3})
	if totals[scope] == 0 {
		scope = 3
		for k := range totals {
			if totals[k] > 0 {
				scope = k
				break
			}
		}
		for _, k := range []int{1, 2, 3} {
			if totals[k] > 0 {
				scope = k
				break
			}
		}
	}
	q := Question{key: "ex/emissions", Exhibit: docExhibit("Greenhouse Gas Inventory (draft)", "Metric tons CO2e · scopes not yet assigned", []string{"Source", "tCO2e"}, rows, "Scope 1: direct, owned sources. Scope 2: purchased energy. Scope 3: other value-chain emissions.")}
	q.Text = fmt.Sprintf("Refer to %s. What is the total for Scope %d?", ex, scope)
	var others []int
	for _, k := range []int{1, 2, 3} {
		if k != scope {
			others = append(others, totals[k])
		}
	}
	q.Choices, q.Correct = numChoices(r, totals[scope], []int{others[0], others[1], 100, -50, 250}, func(n int) string { return commas(n) + " tCO2e" })
	return q
}

// ── recycling sort ─────────────────────────────────────────────────────

func recyclingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	bins := []struct{ name, col string }{{"Mixed recycling", "#1565c0"}, {"Compost", "#2e7d32"}, {"Landfill", "#424242"}, {"E-waste & batteries", "#f9a825"}, {"Confidential shred", "#6a1b9a"}}
	items := map[string]int{"Clean soda can": 0, "Empty water bottle": 0, "Banana peel": 1, "Coffee grounds": 1, "Greasy pizza box": 1, "Chip bag": 2, "Styrofoam cup": 2, "Old phone": 3, "AA batteries": 3, "Printed salary report": 4, "Signed contract draft": 4}
	var keys []string
	for k := range items {
		keys = append(keys, k)
	}
	sortStrings(keys)
	item := pick(r, keys)
	s := newSVG(470, 210, "Recycling station")
	s.rect(0, 0, 470, 210, 8, "#eceff1", "")
	s.text(14, 22, 11, `font-weight="700"`, "Posted guidance, "+sl["PLACE"])
	for i, b := range bins {
		x := 12 + float64(i)*92
		s.raw(`<path d="M%.0f 60 h76 l-8 120 h-60 Z" fill="%s"/>`, x, b.col)
		s.rect(x-4, 50, 84, 12, 3, b.col, "")
		for j, ln := range wrapText(b.name, 11) {
			s.text(x+38, 96+float64(j)*13, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, ln)
		}
		var ex []string
		for _, k := range keys {
			if items[k] == i && k != item && len(ex) < 2 {
				ex = append(ex, k)
			}
		}
		for j, e := range ex {
			s.text(x+38, 140+float64(j)*12, 7, `text-anchor="middle" fill="#fff"`, e)
		}
	}
	q := Question{key: "ex/recycling", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the posted guidance, which bin is right for this item: %s?", ex, strings.ToLower(item))
	var names []string
	for _, b := range bins {
		names = append(names, b.name)
	}
	q.Choices, q.Correct = oneOf(r, bins[items[item]].name, names)
	return q
}

// ── trade association agenda ───────────────────────────────────────────

func tradeAgendaExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	okItems := []string{"Welcome and antitrust statement", "Update on new federal leave regulations", "Panel: accessibility in benefits enrollment", "Survey methodology for anonymized pay data (third-party administered)", "Annual conference logistics", "Election of officers", "Lobbying update on state paid sick leave bills"}
	badItems := []string{"Agreeing on next year's merit increase percentages", "Members' planned price increases for next quarter", "Agreement not to recruit each other's employees", "Dividing up client territories in the Midwest"}
	bad := pick(r, badItems)
	agenda := append(sample(r, okItems, 4), bad)
	shuffle(r, agenda)
	var rows [][]string
	for i, it := range agenda {
		rows = append(rows, []string{fmt.Sprint(i + 1), it})
	}
	q := Question{key: "ex/tradeagenda", Exhibit: docExhibit("Midwest Compensation Professionals Association — Spring Meeting", "Members include several direct competitors of ACPWB", []string{"#", "Agenda item"}, rows, "Counsel will attend. Minutes will be kept.")}
	k := 0
	for i, it := range agenda {
		if it == bad {
			k = i
		}
	}
	q.Text = fmt.Sprintf("Refer to %s. Which agenda item should %s ask to have removed before attending?", ex, sl["FIRST"])
	var labels []string
	for i := range agenda {
		labels = append(labels, fmt.Sprintf("Item %d", i+1))
	}
	q.Choices, q.Correct = oneOf(r, labels[k], labels)
	return q
}

// ── NDA term ───────────────────────────────────────────────────────────

func ndaTermExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	eff := exDate(r)
	term := pick(r, []int{2, 3, 5})
	signedLater := eff.AddDate(0, 0, r.between(3, 40))
	rows := [][]string{{"Parties", "ACPWB and " + sl["VENDOR"]}, {"Effective Date", longDate(eff)}, {"Signed", longDate(signedLater)}, {"Confidentiality period", fmt.Sprintf("%d years from the Effective Date", term)}, {"Trade secrets", "Protected for as long as they remain trade secrets"}}
	q := Question{key: "ex/ndaterm", Exhibit: docExhibit("Mutual Nondisclosure Agreement — Key Terms", "Executed copy on file with Legal", []string{"Term", "Detail"}, rows, "Defend Trade Secrets Act whistleblower immunity notice included.")}
	q.Text = fmt.Sprintf("Refer to %s. For ordinary confidential information (not trade secrets), when does the confidentiality period end?", ex)
	end := eff.AddDate(term, 0, 0)
	q.Choices, q.Correct = withChoices(r, longDate(end), []string{longDate(signedLater.AddDate(term, 0, 0)), longDate(eff.AddDate(term+1, 0, 0)), "Never; all information is protected forever"})
	return q
}

// ── Y2K date windowing ─────────────────────────────────────────────────

func y2kExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	pivot := pick(r, []int{30, 40, 50, 60})
	yy := r.between(0, 99)
	full := 1900 + yy
	if yy < pivot {
		full = 2000 + yy
	}
	s := newSVG(440, 220, "Green-screen terminal")
	s.rect(0, 0, 440, 220, 10, "#111", "")
	s.rect(14, 14, 412, 192, 4, "#062006", "")
	for i, l := range []string{
		"ACPWB PAYROLL MASTER FILE     REL 4.2  1999-11-30",
		"------------------------------------------------",
		"DATE WINDOWING: YY < " + fmt.Sprintf("%02d", pivot) + " -> 20YY ELSE 19YY",
		"",
		"EMPLOYEE  " + strings.ToUpper(sl["NAME"]),
		fmt.Sprintf("HIRE DATE 07/01/%02d", yy),
		"",
		"*** Y2K REMEDIATION COMPLETE. HAVE A NICE MILLENNIUM ***",
	} {
		s.text(26, 38+float64(i)*20, 11, `fill="#33ff66" font-family="Courier New,monospace"`, l)
	}
	q := Question{key: "ex/y2k", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Using the windowing rule shown, what full year does the system store for this hire date?", ex)
	alt := 1900 + yy
	if full == alt {
		alt = 2000 + yy
	}
	q.Choices, q.Correct = withChoices(r, fmt.Sprint(full), []string{fmt.Sprint(alt), fmt.Sprint(1800 + yy), fmt.Sprintf("%d", 2100+yy)})
	return q
}

// ── fax cover sheet ────────────────────────────────────────────────────

func faxCoverExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	total := r.between(4, 12)
	got := total - r.between(1, 3)
	if r.chance(0.3) {
		got = total
	}
	s := paper(400, 300, "Fax cover sheet", "FACSIMILE TRANSMITTAL", fmt.Sprintf("Received %d pages", got))
	for i, kv := range [][2]string{{"TO", sl["NAME"] + ", Legal Dept."}, {"FAX", "(414) 555-0" + fmt.Sprint(r.between(100, 199))}, {"FROM", sl["NAME2"] + ", " + sl["VENDOR"]}, {"DATE", exDate(r).AddDate(-20, 0, 0).Format("01/02/2006")}, {"PAGES", fmt.Sprintf("%d (including cover)", total)}, {"RE", "Signed amendment — CONFIDENTIAL"}} {
		y := 64 + float64(i)*26
		s.text(20, y, 11, `font-weight="700" font-family="Courier New,monospace"`, kv[0]+":")
		s.text(90, y, 11, `font-family="Courier New,monospace"`, kv[1])
	}
	for i, l := range wrapText("This facsimile contains confidential information intended only for the named recipient. If you received it in error, notify the sender by telephone and destroy all copies.", 60) {
		s.text(20, 238+float64(i)*13, 9, `fill="#555" font-family="Courier New,monospace"`, l)
	}
	q := Question{key: "ex/faxcover", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. The fax machine's counter shows %d pages received, including the cover. What should %s do?", ex, got, sl["FIRST"])
	if got == total {
		q.Choices, q.Correct = withChoices(r, "Nothing more; all pages arrived, so file it securely", []string{"Ask for the missing pages", "Report it as a data incident", "Send it back"})
		return q
	}
	missing := "1 page is missing"
	if total-got > 1 {
		missing = fmt.Sprintf("%d pages are missing", total-got)
	}
	q.Choices, q.Correct = withChoices(r, "Contact the sender to resend; "+missing,
		[]string{"File it as complete", "Leave it on the fax machine until the rest arrive", "Ask the sender to email it to a personal account"})
	return q
}

// ── floppy disk write protection ───────────────────────────────────────

func floppyExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	prot := r.intn(4)
	labels := sample(r, []string{"PAYROLL 1994", "BACKUP", "WP 5.1 DOCS", "Q3 BUDGET", "DBASE IV", "TAX FORMS 1997", "LOTUS 123"}, 4)
	s := newSVG(460, 200, "Four 3.5-inch floppy disks, back view")
	s.rect(0, 0, 460, 200, 8, "#e0e0e0", "")
	for i := 0; i < 4; i++ {
		x := 14 + float64(i)*112
		s.rect(x, 20, 100, 104, 4, "#263238", "")
		s.rect(x+28, 20, 46, 30, 1, "#b0bec5", "")
		s.rect(x+12, 70, 76, 48, 2, "#fff", "")
		s.text(x+50, 90, 8, `text-anchor="middle" font-weight="700"`, labels[i])
		// write-protect tab: top-right corner from the back; hole open = protected
		if i == prot {
			s.rect(x+84, 26, 10, 12, 1, "#000", "") // slider moved: hole open
		} else {
			s.rect(x+84, 26, 10, 12, 1, "#78909c", "#455a64") // slider closed
		}
		s.text(x+50, 142, 11, `text-anchor="middle" font-weight="700"`, "Disk "+lettered(i))
	}
	s.text(230, 176, 9, `text-anchor="middle" fill="#555"`, "Corner tab shown from the back: black = open hole, gray = closed")
	q := Question{key: "ex/floppy", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which disk is write-protected (its square hole is open)?", ex)
	q.Choices, q.Correct = oneOf(r, "Disk "+lettered(prot), []string{"Disk A", "Disk B", "Disk C", "Disk D"})
	return q
}

// ── parking lot ────────────────────────────────────────────────────────

func parkingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	kinds := []struct{ label, col string }{{"♿", "#1565c0"}, {"EV", "#2e7d32"}, {"VISITOR", "#e65100"}, {"RESERVED", "#6a1b9a"}, {"", "#9e9e9e"}}
	n := 8
	types := make([]int, n)
	for i := range types {
		types[i] = pick(r, []int{0, 1, 2, 3, 4})
	}
	open := r.intn(n)
	types[open] = 4
	occupied := map[int]bool{}
	for i := range types {
		if i != open && types[i] == 4 {
			occupied[i] = true
		} else if i != open && r.chance(0.4) {
			occupied[i] = true
		}
	}
	s := newSVG(460, 200, "Parking lot map")
	s.rect(0, 0, 460, 200, 6, "#555", "")
	for i := 0; i < n; i++ {
		x := 14 + float64(i)*54
		s.line(x, 20, x, 120, "#fff", 2)
		if types[i] != 4 {
			s.rect(x+6, 90, 40, 22, 3, kinds[types[i]].col, "")
			s.text(x+26, 105, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, kinds[types[i]].label)
		}
		if occupied[i] {
			s.rect(x+8, 28, 36, 56, 8, pick(r, []string{"#b71c1c", "#0d47a1", "#e0e0e0", "#212121", "#f9a825"}), "#222")
		}
		s.text(x+26, 138, 11, `text-anchor="middle" fill="#fff" font-weight="700"`, fmt.Sprint(i+1))
	}
	s.line(14+float64(n)*54, 20, 14+float64(n)*54, 120, "#fff", 2)
	s.text(230, 180, 10, `text-anchor="middle" fill="#fff"`, "Unmarked spaces are general employee parking")
	q := Question{key: "ex/parking", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. %s drives a gas car, has no accessible parking placard, and arrives at 8 a.m. Which space may they use?", ex, sl["FIRST"])
	var opts []string
	for i := 0; i < n; i++ {
		if i != open && !occupied[i] {
			opts = append(opts, fmt.Sprintf("Space %d", i+1))
		}
	}
	for i := 0; len(opts) < 3 && i < n; i++ {
		if i != open && !contains(opts, fmt.Sprintf("Space %d", i+1)) {
			opts = append(opts, fmt.Sprintf("Space %d", i+1))
		}
	}
	q.Choices, q.Correct = oneOf(r, fmt.Sprintf("Space %d", open+1), opts)
	return q
}

// ── org chart approvals ────────────────────────────────────────────────

func orgChartExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	names := people(r, 7)
	// 0 VP; 1,2 directors; 3,4 managers under 1; 5,6 analysts under 3/4
	titles := []string{"Vice President", "Director", "Director", "Manager", "Manager", "Analyst", "Analyst"}
	parent := []int{-1, 0, 0, 1, 1, 3, 4}
	pos := [][2]float64{{230, 40}, {120, 110}, {340, 110}, {60, 180}, {180, 180}, {60, 250}, {180, 250}}
	s := newSVG(460, 290, "Organization chart")
	s.rect(0, 0, 460, 290, 8, "#fff", "#ccc")
	for i, p := range parent {
		if p >= 0 {
			s.line(pos[p][0], pos[p][1]+16, pos[i][0], pos[i][1]-16, "#90a4ae", 1.5)
		}
	}
	for i, n := range names {
		x, y := pos[i][0], pos[i][1]
		s.rect(x-56, y-18, 112, 36, 5, "#e8eef5", "#1a3a5c")
		s.text(x, y-3, 9, `text-anchor="middle" font-weight="700"`, n)
		s.text(x, y+11, 8, `text-anchor="middle" fill="#555"`, titles[i])
	}
	q := Question{key: "ex/orgchart", Exhibit: s.done()}
	k := pick(r, []int{5, 6})
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. %s submits an expense report. Under the policy, the employee's direct manager approves it. Who approves?", ex, names[k])
		q.Choices, q.Correct = oneOf(r, names[parent[k]], names)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. %s wants to raise a concern about their manager. Who is the manager's manager, one option for escalating it?", ex, names[k])
	q.Choices, q.Correct = oneOf(r, names[parent[parent[k]]], names)
	return q
}
