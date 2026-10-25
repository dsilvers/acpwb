package compliance

import (
	"fmt"
	"math"
)

// Trend chart exhibits: a labelled line chart with a target or limit line.
// Questions ask when the metric first crossed the line, how many periods
// were on the wrong side, where the biggest increase was, or one value.

type trendSpec struct {
	title, metric, prefix, unit string
	dec                         int
	lo, hi, thr                 float64
	thrLabel                    string
	badHigh                     bool
	period                      string // month, quarter, week, year, day, hour
	cats                        []string
}

var trendSpecs = map[string]trendSpec{
	"turnover":      {"VOLUNTARY TURNOVER (ANNUALIZED)", "voluntary turnover", "", "%", 1, 8, 22, 15, "Target max", true, "month", []string{"retention", "managers", "careers", "offboarding"}},
	"engagement":    {"EMPLOYEE ENGAGEMENT INDEX", "the engagement index", "", "", 0, 55, 90, 70, "Goal", false, "quarter", []string{"culture", "managers", "surveys", "recognition"}},
	"payratiotrend": {"CEO PAY RATIO BY FISCAL YEAR", "the CEO pay ratio", "", ":1", 0, 150, 420, 300, "Board cap", true, "year", []string{"payratio", "execcomp", "sayonpay", "governance"}},
	"crawlerhits":   {"CRAWLER REQUESTS PER HOUR (THOUSANDS)", "crawler traffic", "", "k", 0, 20, 140, 100, "Alert threshold", true, "hour", []string{"network", "scraping", "security", "botpolicy"}},
	"phishclick":    {"PHISHING SIMULATION CLICK RATE", "the click rate", "", "%", 1, 2, 18, 8, "Target max", true, "month", []string{"phishing", "socialeng", "security"}},
	"ticketbacklog": {"OPEN HELP DESK TICKETS", "the ticket backlog", "", "", 0, 80, 400, 250, "SLA limit", true, "week", []string{"tickets", "devices", "sla"}},
	"overtimetrend": {"TEAM OVERTIME HOURS", "overtime", "", " h", 0, 40, 260, 160, "Budget", true, "week", []string{"wage", "payroll", "timekeeping"}},
	"trir":          {"TOTAL RECORDABLE INCIDENT RATE", "the incident rate", "", "", 1, 0.5, 5.0, 2.5, "Goal", true, "quarter", []string{"safety", "injuries", "nearmiss", "workerscomp"}},
	"energyuse":     {"OFFICE ELECTRICITY USE (MWh)", "electricity use", "", " MWh", 0, 80, 200, 140, "Target", true, "month", []string{"sustainability", "carbon", "facilities"}},
	"cloudspend":    {"CLOUD SPEND ($ THOUSANDS)", "cloud spend", "$", "k", 0, 40, 160, 110, "Budget", true, "month", []string{"cloudcost", "finance", "procurement"}},
	"badgeins":      {"AVERAGE DAILY BADGE-INS", "average badge-ins", "", "", 0, 120, 480, 300, "RTO target", false, "week", []string{"rto", "hybrid", "badge"}},
	"spamvolume":    {"SPAM RECEIVED AT ACPWB.COM (THOUSANDS/DAY)", "spam volume", "", "k", 0, 5, 60, 40, "Filter capacity", true, "day", []string{"email", "spam", "phishing"}},
	"timetofill":    {"AVERAGE TIME TO FILL (DAYS)", "time to fill", "", " d", 0, 25, 80, 45, "Target max", true, "month", []string{"hiring", "recruiting", "careers"}},
	"offeraccept":   {"OFFER ACCEPTANCE RATE", "offer acceptance", "", "%", 0, 60, 98, 85, "Goal", false, "quarter", []string{"hiring", "recruiting", "referrals"}},
	"trainingcomp":  {"MANDATORY TRAINING COMPLETION", "completion", "", "%", 0, 55, 100, 90, "Required level", false, "week", []string{"learning", "onboarding", "ethics"}},
	"auditfindings": {"OPEN AUDIT FINDINGS", "open findings", "", "", 0, 4, 40, 20, "Committee limit", true, "quarter", []string{"audit", "sox", "governance"}},
	"dsotrend":      {"DAYS SALES OUTSTANDING", "DSO", "", " d", 0, 30, 75, 50, "Target max", true, "month", []string{"finance", "collections", "revenue"}},
	"shrink":        {"INVENTORY SHRINK", "shrink", "", "%", 1, 0.5, 4.0, 2.0, "Tolerance", true, "quarter", []string{"warehouse", "inventory", "fraud"}},
	"npstrend":      {"NET PROMOTER SCORE", "NPS", "", "", 0, 5, 70, 30, "Goal", false, "quarter", []string{"customers", "complaints", "brand"}},
	"slabreach":     {"SLA BREACHES", "SLA breaches", "", "", 0, 0, 24, 10, "Contract limit", true, "month", []string{"sla", "contracts", "thirdparty"}},
	"dbcpu":         {"DATABASE SERVER CPU", "CPU use", "", "%", 0, 20, 98, 80, "Alert level", true, "hour", []string{"serverroom", "network", "cloudcost"}},
	"mfaenroll":     {"MFA ENROLLMENT", "MFA enrollment", "", "%", 0, 50, 100, 95, "Required level", false, "week", []string{"mfa", "passwords", "security"}},
	"ptoliability":  {"ACCRUED PTO LIABILITY ($ MILLIONS)", "PTO liability", "$", "M", 1, 1.0, 6.0, 4.0, "Policy cap", true, "quarter", []string{"pto", "timeoff", "finance"}},
	"lotoccupancy":  {"PARKING LOT OCCUPANCY AT 10 A.M.", "occupancy", "", "%", 0, 40, 100, 90, "Capacity alert", true, "day", []string{"parking", "commuting", "rto"}},
	"coffeecups":    {"COFFEE CUPS SERVED (HUNDREDS)", "coffee served", "", "", 0, 8, 30, 20, "Contract volume", true, "week", []string{"coffee", "kitchen"}},
	"printpages":    {"PAGES PRINTED (THOUSANDS)", "printing", "", "k", 0, 10, 80, 45, "Green target", true, "month", []string{"printing", "sustainability"}},
	"meetinghours":  {"AVERAGE MEETING HOURS PER PERSON", "meeting hours", "", " h", 1, 8, 24, 15, "Wellbeing limit", true, "week", []string{"meetings", "focus", "calendar"}},
	"perchregs":     {"PERCH CONFERENCE REGISTRATIONS PER WEEK", "registrations", "", "", 0, 20, 140, 60, "Pace needed", false, "week", []string{"conferences", "events", "offsites"}},
	"archiveviews":  {"ARCHIVE PAGE VIEWS (MILLIONS)", "archive page views", "", "M", 1, 1, 12, 8, "Planned capacity", true, "month", []string{"retro", "records", "scraping"}},
	"contractorpct": {"CONTRACTOR SHARE OF WORKFORCE", "the contractor share", "", "%", 0, 5, 35, 20, "Policy max", true, "quarter", []string{"gig", "contractors", "classification"}},
	"paygaptrend":   {"ADJUSTED GENDER PAY GAP", "the adjusted pay gap", "", "%", 1, 0, 8, 3, "Target max", true, "year", []string{"payequity", "paytransparency", "eeo"}},
	"breachedpw":    {"ACCOUNTS WITH BREACHED PASSWORDS", "breached-password accounts", "", "", 0, 0, 60, 25, "Risk tolerance", true, "week", []string{"passwords", "mfa", "identity"}},
	"wateruse":      {"CAMPUS WATER USE (THOUSAND GALLONS)", "water use", "", "k gal", 0, 40, 120, 90, "Target", true, "month", []string{"sustainability", "water", "facilities"}},
}

func init() {
	gens := map[string]exhibitGen{}
	cats := map[string][]string{}
	for k, sp := range trendSpecs {
		gens[k] = trendGen(k, sp)
		for _, c := range sp.cats {
			cats[c] = append(cats[c], k)
		}
	}
	for _, ks := range cats {
		sortStrings(ks)
	}
	registerExhibits(gens, cats)
}

var monthAbbr = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

func periodLabels(r *rnd, period string, n int) ([]string, string) {
	out := make([]string, n)
	switch period {
	case "month":
		s, y := r.intn(12), r.between(2018, 2025)
		for i := range out {
			m := s + i
			out[i] = fmt.Sprintf("%s '%02d", monthAbbr[m%12], (y+m/12)%100)
		}
		return out, "month"
	case "quarter":
		s, y := r.intn(4), r.between(2018, 2024)
		for i := range out {
			q := s + i
			out[i] = fmt.Sprintf("Q%d '%02d", q%4+1, (y+q/4)%100)
		}
		return out, "quarter"
	case "year":
		y := r.between(2010, 2017)
		for i := range out {
			out[i] = itoa(y + i)
		}
		return out, "year"
	case "week":
		s := r.between(1, 40)
		for i := range out {
			out[i] = fmt.Sprintf("Wk %d", s+i)
		}
		return out, "week"
	case "day":
		m, d := r.intn(12), r.between(1, 20)
		for i := range out {
			out[i] = fmt.Sprintf("%s %d", monthAbbr[m], d+i)
		}
		return out, "day"
	default:
		h := r.between(7, 12)
		for i := range out {
			out[i] = clockShort((h + i) % 24)
		}
		return out, "hour"
	}
}

func clockShort(h int) string {
	switch {
	case h == 0:
		return "12 a.m."
	case h < 12:
		return fmt.Sprintf("%d a.m.", h)
	case h == 12:
		return "12 p.m."
	}
	return fmt.Sprintf("%d p.m.", h-12)
}

func trendGen(kind string, sp trendSpec) exhibitGen {
	scale := math.Pow(10, float64(sp.dec))
	lo, hi, thr := int(math.Round(sp.lo*scale)), int(math.Round(sp.hi*scale)), int(math.Round(sp.thr*scale))
	margin := (hi - lo) / 25
	if margin < 1 {
		margin = 1
	}
	bad := func(v int) bool {
		if sp.badHigh {
			return v > thr
		}
		return v < thr
	}
	// side returns a value on the good (false) or bad (true) side, kept
	// away from the line.
	side := func(r *rnd, isBad bool) int {
		if isBad == sp.badHigh {
			return r.between(thr+margin, hi)
		}
		return r.between(lo, thr-margin)
	}
	show := func(v int) string { return sp.prefix + fmtNum(float64(v)/scale, sp.dec) + sp.unit }
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		const n = 8
		labels, unitName := periodLabels(r, sp.period, n)
		vals := make([]int, n)
		mode := r.intn(4)
		first := r.between(2, 6)
		switch mode {
		case 0:
			for i := range vals {
				switch {
				case i < first:
					vals[i] = side(r, false)
				case i == first:
					vals[i] = side(r, true)
				default:
					vals[i] = side(r, r.chance(0.5))
				}
			}
		case 1:
			for i := range vals {
				vals[i] = side(r, r.chance(0.5))
			}
		default:
			v := r.between(lo, hi)
			for i := range vals {
				v += r.between(-(hi-lo)/5, (hi-lo)/5)
				v = clampInt(v, lo, hi)
				vals[i] = v
			}
		}
		if mode == 2 { // make one increase clearly the biggest
			for tries := 0; ; tries++ {
				if tries > 100 {
					for i := range vals {
						vals[i] = lo + (hi-lo)/5
						if i >= first {
							vals[i] = hi - (hi-lo)/5
						}
					}
				}
				best, second, at := -1<<30, -1<<30, 0
				for i := 1; i < n; i++ {
					d := vals[i] - vals[i-1]
					if d > best {
						best, second, at = d, best, i
					} else if d > second {
						second = d
					}
				}
				if best > 0 && best-second >= margin {
					first = at
					break
				}
				i := r.between(1, n-1)
				vals[i] = clampInt(vals[i-1]+r.between((hi-lo)/4, (hi-lo)/2), lo, hi)
			}
		}

		s := frame(540, 268, sp.title+" line chart", sp.title)
		x0, x1, y0, y1 := 56.0, 520.0, 52.0, 222.0
		ys := func(v int) float64 { return y1 - (y1-y0)*float64(v-lo)/float64(hi-lo) }
		xs := func(i int) float64 { return x0 + 18 + (x1-x0-36)*float64(i)/float64(n-1) }
		for k := 0; k <= 4; k++ {
			v := lo + (hi-lo)*k/4
			y := ys(v)
			s.line(x0, y, x1, y, "#e5e7eb", 1)
			s.text(x0-5, y+3, 7.5, `text-anchor="end" fill="#475569"`, show(v))
		}
		s.line(x0, y0, x0, y1, "#64748b", 1)
		s.line(x0, y1, x1, y1, "#64748b", 1)
		ty := ys(thr)
		s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#dc2626" stroke-width="1.4" stroke-dasharray="6 4"/>`, x0, ty, x1, ty)
		s.raw(`<line x1="%.1f" y1="34" x2="%.1f" y2="34" stroke="#dc2626" stroke-width="1.4" stroke-dasharray="6 4"/>`, x1-150, x1-128)
		s.text(x1-124, 37, 8, `font-weight="700" fill="#dc2626"`, sp.thrLabel+": "+show(thr))
		pts := ""
		for i, v := range vals {
			pts += fmt.Sprintf("%.1f,%.1f ", xs(i), ys(v))
		}
		s.raw(`<polyline points="%s" fill="none" stroke="#1d4ed8" stroke-width="2"/>`, pts)
		for i, v := range vals {
			s.circle(xs(i), ys(v), 3.2, "#1d4ed8", "")
			ly := ys(v) - 7
			if ly < y0 {
				ly = ys(v) + 13
			}
			s.text(xs(i), ly, 7.5, `text-anchor="middle" font-weight="700" fill="#1e3a8a" stroke="#fff" stroke-width="3" paint-order="stroke"`, show(v))
			s.text(xs(i), y1+14, 8, `text-anchor="middle" fill="#334155"`, labels[i])
		}
		q := Question{key: "ex/" + kind, Exhibit: s.done()}
		dir := "rise above"
		wrong := "above"
		if !sp.badHigh {
			dir, wrong = "fall below", "below"
		}
		lbl := lowerFirst(sp.thrLabel)
		switch mode {
		case 0:
			var wrongs []string
			for _, i := range sample(r, []int{0, 1, 2, 3, 4, 5, 6, 7}, 8) {
				if i != first && len(wrongs) < 3 {
					wrongs = append(wrongs, labels[i])
				}
			}
			q.Text = fmt.Sprintf("Refer to %s. In which %s did %s first %s the %s?", ex, unitName, sp.metric, dir, lbl)
			q.Choices, q.Correct = withChoices(r, labels[first], wrongs)
		case 1:
			nb := 0
			for _, v := range vals {
				if bad(v) {
					nb++
				}
			}
			q.Text = fmt.Sprintf("Refer to %s. In how many of the %ss shown was %s %s the %s?", ex, unitName, sp.metric, wrong, lbl)
			q.Choices, q.Correct = numChoices(r, nb, []int{-1, 1, 2, -2, 3}, itoa)
		case 2:
			span := func(i int) string { return labels[i-1] + " to " + labels[i] }
			var wrongs []string
			for _, i := range sample(r, []int{1, 2, 3, 4, 5, 6, 7}, 7) {
				if i != first && len(wrongs) < 3 {
					wrongs = append(wrongs, span(i))
				}
			}
			q.Text = fmt.Sprintf("Refer to %s. Between which two consecutive %ss did %s increase the most?", ex, unitName, sp.metric)
			q.Choices, q.Correct = withChoices(r, span(first), wrongs)
		default:
			i := r.intn(n)
			d := (hi - lo) / 10
			prep := map[string]string{"hour": "at", "day": "on"}[sp.period]
			if prep == "" {
				prep = "in"
			}
			q.Text = fmt.Sprintf("Refer to %s. What was %s %s %s?", ex, sp.metric, prep, labels[i])
			q.Choices, q.Correct = numChoices(r, vals[i], []int{-margin, margin, d, -d, 2 * d, -2 * d}, show)
		}
		return q
	}
}
