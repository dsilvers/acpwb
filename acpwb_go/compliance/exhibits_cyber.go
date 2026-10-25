package compliance

// Security and privacy exhibits: MFA prompts, passwords, lookalike URLs,
// Wi-Fi evil twins, QR stickers, smishing, file sharing, impossible
// travel, tailgating counts, ransomware, hidden spreadsheet columns,
// classification labels, GDPR clocks (one month; 72 hours), and legal
// hold scope. Answers are computed from what is drawn.

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func init() {
	for k, g := range map[string]exhibitGen{
		"mfaprompt":   mfaPromptExhibit,
		"passwords":   passwordsExhibit,
		"urlbar":      urlBarExhibit,
		"wifilist":    wifiListExhibit,
		"qrposter":    qrPosterExhibit,
		"smishing":    smishingExhibit,
		"sharing":     sharingExhibit,
		"signins":     signInsExhibit,
		"tailgate":    tailgateExhibit,
		"ransomnote":  ransomNoteExhibit,
		"hiddencols":  hiddenColsExhibit,
		"classlabels": classLabelsExhibit,
		"dsarclock":   dsarClockExhibit,
		"breachclock": breachClockExhibit,
		"holdscope":   holdScopeExhibit,
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
		"mfa":           {"mfaprompt", "signins"},
		"passwords":     {"passwords"},
		"phishing":      {"urlbar", "smishing"},
		"email":         {"urlbar"},
		"socialeng":     {"smishing", "qrposter", "urlbar"},
		"network":       {"wifilist"},
		"intltravel":    {"wifilist"},
		"devices":       {"qrposter", "wifilist"},
		"privacy":       {"sharing", "classlabels"},
		"privacy2":      {"dsarclock", "breachclock"},
		"security":      {"mfaprompt", "signins", "sharing"},
		"ransomware":    {"ransomnote", "breachclock"},
		"spreadsheets":  {"hiddencols"},
		"records":       {"classlabels"},
		"doccontrol":    {"classlabels"},
		"legalhold":     {"holdscope"},
		"badge":         {"tailgate"},
		"visitor":       {"tailgate"},
		"physicalsec":   {"tailgate"},
		"lobby":         {"tailgate"},
		"health":        {"sharing"},
		"customerdata":  {"dsarclock", "sharing"},
		"acceptableuse": {"signins"},
	} {
		categoryExhibits[cat] = append(categoryExhibits[cat], ks...)
	}
}

// ── MFA push prompt ────────────────────────────────────────────────────

var mfaCities = []string{"Milwaukee, WI", "Chicago, IL", "Madison, WI", "Lagos, Nigeria", "Bucharest, Romania", "São Paulo, Brazil", "Minsk, Belarus", "Jakarta, Indonesia"}

func mfaPromptExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	city := pick(r, mfaCities[:3])
	foreign := r.chance(0.5)
	shownCity := city
	if foreign {
		shownCity = pick(r, mfaCities[3:])
	}
	initiated := r.chance(0.5)
	num := r.between(10, 99)
	t := r.between(7*60, 23*60)
	s := phone(320, "MFA sign-in request on a phone", "ACPWB Authenticator")
	s.rect(28, 80, 204, 250, 14, "#fff", "#d0d0d8")
	s.circle(130, 118, 22, "#1a3a5c", "")
	s.text(130, 124, 16, `text-anchor="middle" fill="#fff" font-weight="700"`, "A")
	s.text(130, 162, 13, `text-anchor="middle" font-weight="700"`, "Are you trying to sign in?")
	s.text(130, 180, 9, `text-anchor="middle" fill="#666"`, "ACPWB Single Sign-On")
	for i, kv := range [][2]string{{"Location", shownCity}, {"Time", clock(t)}, {"App", pick(r, []string{"Email", "Workday", "Board Portal", "VPN"})}} {
		y := 204 + float64(i)*18
		s.text(44, y, 9, `fill="#888"`, kv[0])
		s.text(216, y, 9, `text-anchor="end" font-weight="700"`, kv[1])
	}
	s.text(130, 272, 9, `text-anchor="middle" fill="#666"`, "Enter the number shown on your screen")
	s.rect(84, 280, 92, 26, 6, "#f2f2f7", "#999")
	s.text(130, 298, 12, `text-anchor="middle" fill="#bbb"`, "_ _")
	s.rect(40, 314, 84, 26, 13, "#e5e5ea", "")
	s.text(82, 331, 11, `text-anchor="middle" font-weight="700" fill="#c62828"`, "Deny")
	s.rect(136, 314, 84, 26, 13, "#1a3a5c", "")
	s.text(178, 331, 11, `text-anchor="middle" font-weight="700" fill="#fff"`, "Approve")
	q := Question{key: "ex/mfaprompt", Exhibit: s.done()}
	situation := fmt.Sprintf("You are at your desk in %s and have not tried to sign in to anything.", city)
	if initiated {
		situation = fmt.Sprintf("You just entered your password on your laptop in %s, and the sign-in screen shows the number %d.", city, num)
	}
	q.Text = fmt.Sprintf("Refer to %s. %s What should you do with this prompt?", ex, situation)
	approve := fmt.Sprintf("Enter %d and approve it", num)
	deny := "Deny it and report it to IT Security"
	right := deny
	if initiated && !foreign {
		right = approve
	}
	wrongs := []string{"Approve it so the prompts stop", "Ignore it and leave it on screen"}
	if right == deny {
		wrongs = append(wrongs, approve)
	} else {
		wrongs = append(wrongs, deny)
	}
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── password policy ────────────────────────────────────────────────────

var passWords = []string{"maple", "orbit", "quiet", "lantern", "harbor", "velvet", "copper", "meadow", "pixel", "thunder", "walnut", "glacier", "saffron", "tundra", "ember", "falcon"}

func passwordsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	first := strings.ToLower(sl["FIRST"])
	breached := []string{"Password123!", "Welcome2024!", "Summer2025!!", "Acpwb!" + sl["YEAR"], "Milwaukee#1fan"}
	good := strings.Join(sample(r, passWords, 3), "-") + fmt.Sprint(r.between(10, 99))
	cands := []string{
		good,
		pick(r, breached[:3]),
		strings.Title(first) + "-" + strings.Join(sample(r, passWords, 2), "-") + "!",
		pick(r, passWords) + fmt.Sprint(r.between(1, 99)) + "!",
		"ACPWB-" + strings.Join(sample(r, passWords, 2), "-"),
	}
	s := paper(400, 250, "Password policy and candidates", "Password Checker", "IT Security")
	s.text(14, 56, 10, `font-weight="700"`, "Policy:")
	for i, l := range []string{"• At least 14 characters", "• Not on the breached-password list", "• Must not contain your name or \"ACPWB\""} {
		s.text(24, 72+float64(i)*14, 10, "", l)
	}
	s.text(14, 128, 10, `font-weight="700"`, "Breached list (excerpt):")
	s.text(24, 144, 10, `font-family="Courier New,monospace"`, strings.Join(breached[:3], "   "))
	s.text(14, 172, 10, `font-weight="700"`, "Account: "+sl["NAME"])
	s.text(14, 196, 9, `fill="#666"`, "Candidates are shown in the question.")
	q := Question{key: "ex/passwords", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which password meets every rule in the policy for %s's account?", ex, sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, cands[0], sample(r, cands[1:], 3))
	return q
}

// ── address bars (lookalike domains) ───────────────────────────────────

func urlBarExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	path := pick(r, []string{"/sso/login", "/benefits/enroll", "/payroll/direct-deposit", "/compliance/", "/boardportal/signin"})
	real := "https://login.acpwb.com" + path
	fakes := []string{
		"https://login.acpwb.com.verify-session.net" + path,
		"https://login-acpwb.com" + path,
		"https://login.acpvvb.com" + path,
		"https://acpwb.login-secure.co" + path,
		"https://login.acpwb-hr.com" + path,
		"https://login.acpwb.cm" + path,
	}
	urls := append([]string{real}, sample(r, fakes, 3)...)
	shuffle(r, urls)
	s := newSVG(460, 40+float64(len(urls))*48, "Browser address bars")
	s.rect(0, 0, 460, 40+float64(len(urls))*48, 8, "#eceff3", "")
	s.text(14, 24, 11, `font-weight="700"`, "Four tabs, four address bars")
	for i, u := range urls {
		y := 40 + float64(i)*48
		s.rect(12, y, 436, 36, 18, "#fff", "#c9ced6")
		s.text(28, y+23, 11, `font-weight="700" fill="#1a3a5c"`, lettered(i))
		s.raw(`<path d="M50 %.1f h10 v8 h-10 Z M52 %.1f v-3 a3 3 0 0 1 6 0 v3" fill="none" stroke="#555" stroke-width="1.4"/>`, y+16, y+16)
		s.text(68, y+23, 11, `font-family="Courier New,monospace"`, u)
	}
	q := Question{key: "ex/urlbar", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. %s received a link to sign in. Which address bar shows the genuine ACPWB site?", ex, sl["FIRST"])
	var labels []string
	right := ""
	for i, u := range urls {
		labels = append(labels, "Address "+lettered(i))
		if u == real {
			right = "Address " + lettered(i)
		}
	}
	q.Choices, q.Correct = oneOf(r, right, labels)
	return q
}

// ── Wi-Fi list (evil twin) ─────────────────────────────────────────────

func wifiListExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	venue := pick(r, []string{"Lakeshore Suites", "Pfister", "MKE Airport", "Brookfield Conference Center", "Third Ward Coffee"})
	ssid := strings.ReplaceAll(venue, " ", "") + "_Guest"
	nets := []struct {
		name   string
		locked bool
	}{{ssid, true}, {ssid, false}, {strings.ReplaceAll(venue, " ", "") + "_FREE", false}, {"FREE_PUBLIC_WIFI", false}, {"ACPWB-Corp-5G", true}}
	order := sample(r, rangeInts(len(nets)), len(nets))
	s := phone(260, "Wi-Fi network list on a phone", "Wi-Fi")
	for i, k := range order {
		n := nets[k]
		y := 80 + float64(i)*42
		s.rect(24, y, 212, 36, 8, "#fff", "")
		s.text(36, y+22, 11, `font-weight="700"`, n.name)
		for b := 0; b < 3; b++ {
			s.rect(196+float64(b)*6, y+24-float64(b)*5, 4, 6+float64(b)*5, 1, "#1a3a5c", "")
		}
		if n.locked {
			s.raw(`<path d="M176 %.1f h9 v8 h-9 Z M178 %.1f v-3 a2.5 2.5 0 0 1 5 0 v3" fill="none" stroke="#333" stroke-width="1.3"/>`, y+15, y+15)
		}
	}
	s.text(130, 300, 9, `text-anchor="middle" fill="#666"`, "🔒 = password required")
	q := Question{key: "ex/wifilist", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. At %s, the front desk card says the network is \"%s\" with a password. Which network should %s join (then connect the VPN)?", ex, venue, ssid, sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, "\""+ssid+"\" with the lock icon", []string{"\"" + ssid + "\" without the lock icon", "\"" + nets[2].name + "\"", "\"FREE_PUBLIC_WIFI\""})
	return q
}

// ── QR code sticker ────────────────────────────────────────────────────

func qrPattern(s *svgB, x, y, size float64, seed *rnd) {
	cell := size / 21
	s.rect(x, y, size, size, 0, "#fff", "")
	for i := 0; i < 21; i++ {
		for j := 0; j < 21; j++ {
			finder := (i < 7 && j < 7) || (i < 7 && j > 13) || (i > 13 && j < 7)
			on := seed.chance(0.45)
			if finder {
				ii, jj := i%14, j%14
				if ii > 6 {
					ii -= 14
				}
				on = ii == 0 || ii == 6 || jj == 0 || jj == 6 || (ii >= 2 && ii <= 4 && jj >= 2 && jj <= 4)
				if i > 6 && i < 14 || j > 6 && j < 14 && !(i < 7 || i > 13) {
					on = seed.chance(0.45)
				}
			}
			if on {
				s.rect(x+float64(j)*cell, y+float64(i)*cell, cell, cell, 0, "#111", "")
			}
		}
	}
}

func qrPosterExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	dom := pick(r, []string{"pay-mke-parking.co", "acpwb-parking-pay.info", "parkmke-secure.net", "mkeparkpay.top"})
	s := newSVG(440, 300, "Parking pay station with QR code")
	s.rect(0, 0, 440, 300, 8, "#dde3e8", "")
	s.rect(30, 20, 200, 260, 10, "#1a3a5c", "")
	s.text(130, 52, 14, `text-anchor="middle" fill="#fff" font-weight="700"`, "ACPWB PARKING")
	s.text(130, 70, 10, `text-anchor="middle" fill="#c9d6e3"`, "Scan to pay · Garage Level 2")
	qrPattern(s, 70, 90, 120, newRnd("qr_orig"))
	// a sticker pasted slightly askew over the original
	s.raw(`<g transform="rotate(-4 130 150)">`)
	s.rect(76, 96, 112, 112, 2, "#fff", "#bbb")
	qrPattern(s, 80, 100, 104, r)
	s.raw(`</g>`)
	s.text(130, 236, 9, `text-anchor="middle" fill="#c9d6e3"`, "Official pay site: parking.acpwb.com")
	s.rect(260, 70, 160, 120, 10, "#fff", "#999")
	s.text(340, 92, 10, `text-anchor="middle" font-weight="700"`, "Camera preview")
	s.text(340, 118, 9, `text-anchor="middle" fill="#666"`, "Open in browser:")
	s.text(340, 138, 10, `text-anchor="middle" fill="#1565c0" font-family="Courier New,monospace"`, dom)
	s.rect(290, 156, 100, 22, 11, "#1565c0", "")
	s.text(340, 171, 10, `text-anchor="middle" fill="#fff" font-weight="700"`, "Open")
	q := Question{key: "ex/qrposter", Exhibit: s.done()}
	if r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. Where would scanning this QR code take %s?", ex, sl["FIRST"])
		q.Choices, q.Correct = withChoices(r, dom, []string{"parking.acpwb.com", "acpwb.com", "The garage's pay station screen"})
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. What should %s do?", ex, sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, "Not open the link, pay at parking.acpwb.com directly, and report the sticker to Security",
		[]string{"Open the link, since it is on an ACPWB sign", "Peel off the sticker and scan the code underneath, then pay", "Open the link but use a personal card"})
	return q
}

// ── smishing ───────────────────────────────────────────────────────────

func smishingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	type sms struct{ from, text string }
	legit := []sms{
		{"ACPWB Alerts", "Reminder: fire drill today at 10:30 a.m. No action needed. Reply STOP to opt out."},
		{"Dr. Patel's Office", "Appointment confirmed for Thu 2:15 p.m. Call 414-555-0142 to reschedule."},
		{"Lakeshore Air", "Flight LA218 MKE→LGA departs Gate C14 on time."},
		{"Mom", "Did you get the photos from Sunday?"},
		{sl["FIRST2"], "Running 5 min late to the 3:00, start without me"},
	}
	bad := pick(r, []sms{
		{"+1 (872) 555-0199", "USPS: Your package is on hold due to an incomplete address. Update within 12 hrs: usps-redeliver-help.com"},
		{"ACPWB-HR", "Your direct deposit will be suspended today. Verify now: acpwb-payroll-verify.net"},
		{"+44 7700 900123", "This is the CEO. I need you to buy 5 gift cards for a client ASAP. Text me the codes."},
		{"IRS Notice", "You have a pending tax refund of $1,422.80. Claim at irs-refund-claims.org"},
		{"Lakeshore Bank", "Unusual sign-in detected. Confirm your identity: lakeshore-bank-secure.co/verify"},
	})
	msgs := append(sample(r, legit, 3), bad)
	shuffle(r, msgs)
	s := phone(360, "Text messages on a phone", "Messages")
	y := 76.0
	for _, m := range msgs {
		lines := wrapText(m.text, 36)
		h := 20 + float64(len(lines))*12
		s.rect(24, y, 212, h+10, 10, "#fff", "#d0d0d8")
		s.text(34, y+14, 9, `font-weight="700"`, m.from)
		for j, ln := range lines {
			s.text(34, y+28+float64(j)*12, 9, "", ln)
		}
		y += h + 18
	}
	q := Question{key: "ex/smishing", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which text message is most likely a smishing attempt?", ex)
	var froms []string
	for _, m := range msgs {
		froms = append(froms, "The message from "+m.from)
	}
	q.Choices, q.Correct = oneOf(r, "The message from "+bad.from, froms)
	return q
}

// ── file sharing settings ──────────────────────────────────────────────

func sharingExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	files := sample(r, []string{"Q3 Board Deck.pptx", "Salary Bands FY.xlsx", "Team Lunch Menu.docx", "PERCH Agenda (Public).pdf", "Open Enrollment FAQ.pdf", "Acquisition Target List.xlsx", "Holiday Party Photos", "Employee SSN Export.csv", "Office Floor Map.pdf"}, 5)
	labels := map[string]string{"Q3 Board Deck.pptx": "Confidential", "Salary Bands FY.xlsx": "Confidential", "Team Lunch Menu.docx": "Internal", "PERCH Agenda (Public).pdf": "Public", "Open Enrollment FAQ.pdf": "Internal", "Acquisition Target List.xlsx": "Restricted", "Holiday Party Photos": "Internal", "Employee SSN Export.csv": "Restricted", "Office Floor Map.pdf": "Internal"}
	var conf []int
	for i, f := range files {
		if labels[f] == "Confidential" || labels[f] == "Restricted" {
			conf = append(conf, i)
		}
	}
	if len(conf) == 0 {
		files[0] = "Employee SSN Export.csv"
		conf = []int{0}
	}
	bad := pick(r, conf)
	s := newSVG(460, 60+float64(len(files))*40, "File sharing settings")
	s.rect(0, 0, 460, 60+float64(len(files))*40, 8, "#fff", "#ccc")
	s.text(14, 24, 12, `font-weight="700" fill="#1a3a5c"`, "Shared Drive — "+sl["DEPT"])
	s.text(14, 44, 9, `font-weight="700" fill="#666"`, "FILE")
	s.text(250, 44, 9, `font-weight="700" fill="#666"`, "LABEL")
	s.text(340, 44, 9, `font-weight="700" fill="#666"`, "GENERAL ACCESS")
	for i, f := range files {
		y := 66 + float64(i)*40
		s.line(14, y-14, 446, y-14, "#eee", 1)
		s.text(14, y+4, 11, "", f)
		col := map[string]string{"Public": "#2e7d32", "Internal": "#1565c0", "Confidential": "#e65100", "Restricted": "#c62828"}[labels[f]]
		s.rect(250, y-9, 76, 18, 9, col, "")
		s.text(288, y+4, 9, `text-anchor="middle" fill="#fff" font-weight="700"`, labels[f])
		access := "Restricted to ACPWB"
		if i == bad {
			access = "Anyone with the link"
		} else if labels[f] == "Public" && r.chance(0.6) {
			access = "Anyone with the link"
		} else if labels[f] == "Restricted" || r.chance(0.3) {
			access = "Only people added"
		}
		s.text(340, y+4, 10, "", access)
	}
	q := Question{key: "ex/sharing", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Which file's sharing setting must be fixed right away?", ex)
	q.Choices, q.Correct = oneOf(r, files[bad], files)
	return q
}

// ── impossible travel ──────────────────────────────────────────────────

var signinCities = []struct {
	name     string
	lat, lon float64
}{{"Milwaukee, US", 43.04, -87.91}, {"Chicago, US", 41.88, -87.63}, {"New York, US", 40.71, -74.01}, {"Denver, US", 39.74, -104.99}, {"London, UK", 51.51, -0.13}, {"Singapore", 1.35, 103.82}, {"São Paulo, BR", -23.55, -46.63}, {"Tokyo, JP", 35.68, 139.69}, {"Lagos, NG", 6.52, 3.38}}

func miles(a, b int) float64 {
	p := math.Pi / 180
	la1, lo1, la2, lo2 := signinCities[a].lat*p, signinCities[a].lon*p, signinCities[b].lat*p, signinCities[b].lon*p
	h := math.Pow(math.Sin((la2-la1)/2), 2) + math.Cos(la1)*math.Cos(la2)*math.Pow(math.Sin((lo2-lo1)/2), 2)
	return 2 * 3959 * math.Asin(math.Sqrt(h))
}

func signInsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	// A plausible day around Milwaukee/Chicago, plus one sign-in that would
	// need faster-than-airliner travel from the previous one.
	t := r.between(6*60, 10*60)
	cur := 0
	var rows [][]string
	type ev struct{ city, at int }
	var evs []ev
	n := r.between(4, 6)
	bad := r.between(1, n-1)
	for i := 0; i < n; i++ {
		if i == bad {
			far := r.between(4, len(signinCities)-1)
			evs = append(evs, ev{far, t + r.between(20, 90)})
			cur = far
		} else if i == bad+1 {
			evs = append(evs, ev{0, evs[i-1].at + r.between(5, 40)})
			cur = 0
		} else {
			next := cur
			if i > 0 && r.chance(0.3) && cur <= 1 {
				next = 1 - cur
			}
			gap := r.between(40, 120)
			if i > 0 && next != cur {
				gap = r.between(150, 240) // a drive between Milwaukee and Chicago
			}
			at := t
			if i > 0 {
				at = evs[i-1].at + gap
			}
			evs = append(evs, ev{next, at})
			cur = next
		}
		t = evs[i].at
	}
	for _, e := range evs {
		rows = append(rows, []string{clock(e.at % (24 * 60)), signinCities[e.city].name, pick(r, []string{"Email", "VPN", "Workday", "Board Portal"}), "Success"})
	}
	q := Question{key: "ex/signins", Exhibit: docExhibit("Sign-in Activity — "+sl["NAME"], "All times Central · same day", []string{"Time", "Location", "App", "Result"}, rows, "Impossible travel: two sign-ins farther apart than an airliner (about 600 mph) could fly in the time between them.")}
	q.Text = fmt.Sprintf("Refer to %s. Which sign-in should be investigated as impossible travel?", ex)
	var labels []string
	for _, e := range evs {
		labels = append(labels, fmt.Sprintf("%s from %s", clock(e.at%(24*60)), signinCities[e.city].name))
	}
	_ = miles
	q.Choices, q.Correct = oneOf(r, labels[bad], labels)
	return q
}

// ── turnstile tailgating ───────────────────────────────────────────────

func tailgateExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	var rows [][]string
	worst, worstN, total := "", -1, 0
	for h := 7; h <= 11; h++ {
		sw := r.between(20, 140)
		extra := r.between(0, 6)
		if r.chance(0.2) {
			extra = r.between(7, 14)
		}
		label := fmt.Sprintf("%s–%s", clock(h*60), clock(h*60+59))
		rows = append(rows, []string{label, fmt.Sprint(sw), fmt.Sprint(sw + extra)})
		total += extra
		if extra > worstN {
			worst, worstN = label, extra
		} else if extra == worstN {
			worst = "" // tie: no single worst hour
		}
	}
	q := Question{key: "ex/tailgate", Exhibit: docExhibit("Turnstile Report — "+sl["DOOR"], "Badge swipes vs. people counted by the overhead sensor", []string{"Hour", "Badge swipes", "People counted"}, rows, "Each person must swipe their own badge.")}
	if worst != "" && r.chance(0.5) {
		q.Text = fmt.Sprintf("Refer to %s. During which hour did the most people enter without swiping?", ex)
		var labels []string
		for _, row := range rows {
			labels = append(labels, row[0])
		}
		q.Choices, q.Correct = oneOf(r, worst, labels)
		return q
	}
	q.Text = fmt.Sprintf("Refer to %s. How many people in total entered without swiping a badge?", ex)
	q.Choices, q.Correct = countQ(r, total, "person")
	for i := range q.Choices {
		q.Choices[i] = strings.Replace(q.Choices[i], "persons", "people", 1)
	}
	return q
}

// ── ransomware screen ──────────────────────────────────────────────────

func ransomNoteExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	hrs := r.between(24, 96)
	btc := float64(r.between(5, 90)) / 10
	s := newSVG(440, 260, "Ransomware message on a screen")
	s.rect(0, 0, 440, 260, 10, "#222", "")
	s.rect(14, 14, 412, 214, 4, "#7a0000", "")
	s.text(220, 50, 18, `text-anchor="middle" fill="#fff" font-weight="700" font-family="Courier New,monospace"`, "YOUR FILES ARE ENCRYPTED")
	for i, l := range []string{"All documents, databases, and backups on this", "network have been encrypted. To recover them,", fmt.Sprintf("send %.1f BTC to the wallet below.", btc), "Do not contact IT. Do not restart."} {
		s.text(220, 80+float64(i)*16, 11, `text-anchor="middle" fill="#ffd7d7" font-family="Courier New,monospace"`, l)
	}
	s.rect(110, 156, 220, 40, 4, "#000", "")
	s.text(220, 182, 20, `text-anchor="middle" fill="#ff5252" font-weight="700" font-family="Courier New,monospace"`, fmt.Sprintf("%02d:%02d:%02d", hrs, r.between(0, 59), r.between(0, 59)))
	s.text(220, 214, 9, `text-anchor="middle" fill="#ffb3b3" font-family="Courier New,monospace"`, "bc1q"+md5Hex(sl["NAME"])[:24])
	s.rect(200, 234, 40, 18, 2, "#333", "")
	q := Question{key: "ex/ransomnote", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. This appears on %s's laptop. What should %s do first?", ex, sl["FIRST"], sl["FIRST"])
	q.Choices, q.Correct = withChoices(r, "Disconnect from the network (unplug the cable or turn off Wi-Fi) and call IT Security right away",
		[]string{"Pay the ransom quickly before the timer runs out", "Restart the laptop several times to clear the message", "Follow the note and avoid contacting IT"})
	return q
}

// ── spreadsheet hidden columns ─────────────────────────────────────────

func hiddenColsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	all := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	heads := []string{"Name", "Department", "Title", "Base salary", "Home address", "Start date", "Rating", "SSN (last 4)"}
	hide := sample(r, []int{2, 3, 4, 5, 6, 7}, r.between(1, 2))
	sortInts(hide)
	var vis []int
	for i := range all {
		if !containsInt(hide, i) {
			vis = append(vis, i)
		}
	}
	w := 70.0
	s := newSVG(30+w*float64(len(vis))+10, 200, "Spreadsheet with column headers")
	s.rect(0, 0, 30+w*float64(len(vis))+10, 200, 4, "#fff", "#bbb")
	s.rect(0, 0, 30+w*float64(len(vis))+10, 22, 0, "#e8eaed", "")
	names := people(r, 5)
	for j, ci := range vis {
		x := 30 + float64(j)*w
		s.line(x, 0, x, 200, "#d0d4d9", 1)
		s.text(x+w/2, 15, 10, `text-anchor="middle" font-weight="700" fill="#444"`, all[ci])
		if j > 0 && vis[j-1] != ci-1 { // the double line where columns are hidden
			s.line(x-1.5, 0, x-1.5, 200, "#888", 1)
			s.line(x+1.5, 0, x+1.5, 200, "#888", 1)
		}
		s.text(x+4, 38, 9, `font-weight="700"`, heads[ci])
		for i := 0; i < 5; i++ {
			v := ""
			switch ci {
			case 0:
				v = names[i]
			case 1:
				v = pick(r, departments)
			case 2:
				v = pick(r, []string{"Analyst", "Manager", "Director", "Specialist"})
			case 5:
				v = exDate(r).Format("01/02/06")
			case 6:
				v = fmt.Sprint(r.between(1, 5))
			default:
				v = "•••"
			}
			if len(v) > 12 {
				v = v[:11] + "…"
			}
			s.text(x+4, 60+float64(i)*24, 9, "", v)
		}
	}
	for i := 0; i < 6; i++ {
		s.text(15, 38+float64(i)*24, 9, `text-anchor="middle" fill="#666"`, fmt.Sprint(i+1))
		s.line(0, 44+float64(i)*24, 30+w*float64(len(vis))+10, 44+float64(i)*24, "#eee", 1)
	}
	q := Question{key: "ex/hiddencols", Exhibit: s.done()}
	var hl []string
	for _, h := range hide {
		hl = append(hl, all[h])
	}
	right := strings.Join(hl, " and ")
	q.Text = fmt.Sprintf("Refer to %s. %s plans to email this spreadsheet to a vendor. Which columns are hidden, but still in the file?", ex, sl["FIRST"])
	var wrongs []string
	for _, o := range [][]int{{1}, {7}, {0}, {2, 6}, {5}, {3}, {4, 7}} {
		var ol []string
		for _, h := range o {
			ol = append(ol, all[h])
		}
		if w := strings.Join(ol, " and "); w != right && !contains(wrongs, w) {
			wrongs = append(wrongs, w)
		}
	}
	wrongs = append(wrongs[:2], "None; hidden columns are removed when the file is saved")
	q.Choices, q.Correct = withChoices(r, right, wrongs)
	return q
}

// ── classification labels ──────────────────────────────────────────────

func classLabelsExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	levels := []string{"Public", "Internal", "Confidential", "Restricted"}
	allowed := map[string][]string{
		"Public":       {"anyone", "employees", "partners under NDA", "named people"},
		"Internal":     {"employees", "partners under NDA", "named people"},
		"Confidential": {"partners under NDA", "named people"},
		"Restricted":   {"named people"},
	}
	recips := []struct{ who, kind string }{{"a reporter", "anyone"}, {"the whole company", "employees"}, {"an outside consultant with a signed NDA", "partners under NDA"}, {"two named board members", "named people"}, {"a vendor with no NDA", "anyone"}, {"a public conference audience", "anyone"}}
	docs := sample(r, []string{"Office holiday schedule", "Draft press release", "Benefits rate sheet", "M&A target memo", "PERCH speaker list", "Pay equity analysis", "Org chart", "Board minutes"}, 4)
	bad := r.intn(4)
	var rows [][]string
	var labels []string
	for i, d := range docs {
		var lvl string
		var rc struct{ who, kind string }
		if i == bad {
			lvl = pick(r, levels[1:])
			for {
				rc = pick(r, recips)
				if !contains(allowed[lvl], rc.kind) {
					break
				}
			}
		} else {
			lvl = pick(r, levels)
			for {
				rc = pick(r, recips)
				if contains(allowed[lvl], rc.kind) {
					break
				}
			}
		}
		rows = append(rows, []string{d, lvl, rc.who})
		labels = append(labels, d)
	}
	rows = append(rows, []string{"—", "—", "—"})
	q := Question{key: "ex/classlabels", Exhibit: docExhibit("Outgoing Documents — Review Queue", "Public: anyone · Internal: employees and NDA partners · Confidential: NDA partners and named people · Restricted: named people only", []string{"Document", "Label", "Planned recipient"}, rows[:4], "Check the label before every send.")}
	q.Text = fmt.Sprintf("Refer to %s. Which planned send violates the classification policy?", ex)
	q.Choices, q.Correct = oneOf(r, labels[bad], labels)
	return q
}

// ── GDPR one-month request clock ───────────────────────────────────────

// oneMonthLater is the GDPR "one month": the same day next month, or the
// last day of next month when that day does not exist.
func oneMonthLater(t time.Time) time.Time {
	y, m, d := t.Date()
	first := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	if d > last {
		d = last
	}
	return time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, time.UTC)
}

func dsarClockExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	got := time.Date(r.between(2019, 2026), time.Month(r.between(1, 12)), pick(r, []int{3, 9, 14, 21, 28, 29, 30, 31}), 0, 0, 0, 0, time.UTC)
	for got.Day() < 3 { // AddDate overflow into next month: move back
		got = got.AddDate(0, 0, -3)
	}
	due := oneMonthLater(got)
	s := newSVG(400, 220, "Data subject request ticket")
	s.rect(0, 0, 400, 220, 8, "#fff", "#ccc")
	s.rect(0, 0, 400, 32, 8, "#4a148c", "")
	s.rect(0, 20, 400, 12, 0, "#4a148c", "")
	s.text(14, 21, 12, `font-weight="700" fill="#fff"`, "PRIVACY REQUEST · DSAR-"+fmt.Sprint(r.between(1000, 9999)))
	for i, kv := range [][2]string{{"Type", "Access request (GDPR Article 15)"}, {"Requester", "Former employee, Berlin office"}, {"Received", longDate(got)}, {"Identity verified", "Yes"}, {"Extension", "None requested"}} {
		y := 58 + float64(i)*26
		s.text(14, y, 10, `fill="#666"`, kv[0])
		s.text(150, y, 11, `font-weight="700"`, kv[1])
	}
	q := Question{key: "ex/dsarclock", Exhibit: s.done()}
	q.Text = fmt.Sprintf("Refer to %s. Without an extension, by what date must the response be sent?", ex)
	q.Choices, q.Correct = withChoices(r, longDate(due), []string{longDate(got.AddDate(0, 0, 30+r.between(5, 15))), longDate(got.AddDate(0, 0, 14)), longDate(oneMonthLater(due))})
	return q
}

// ── 72-hour breach clock ───────────────────────────────────────────────

func breachClockExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	aware := time.Date(r.between(2019, 2026), time.Month(r.between(1, 12)), r.between(1, 25), r.between(0, 23), 15*r.between(0, 3), 0, 0, time.UTC)
	due := aware.Add(72 * time.Hour)
	f := func(t time.Time) string { return t.Format("Mon Jan 2, ") + clock(t.Hour()*60+t.Minute()) }
	rows := [][]string{
		{"Incident detected by monitoring", f(aware.Add(-time.Duration(r.between(2, 30)) * time.Hour))},
		{"Confirmed personal data affected (aware)", f(aware)},
		{"Containment complete", f(aware.Add(time.Duration(r.between(4, 30)) * time.Hour))},
		{"Data subjects affected", commas(r.between(120, 40000))},
	}
	q := Question{key: "ex/breachclock", Exhibit: docExhibit("Incident Timeline — "+sl["DEPT"], "Personal data breach · EU employees affected", []string{"Event", "Time (CET)"}, rows, "GDPR: notify the supervisory authority within 72 hours of becoming aware, where feasible.")}
	q.Text = fmt.Sprintf("Refer to %s. By when must ACPWB notify the supervisory authority?", ex)
	q.Choices, q.Correct = withChoices(r, f(due), []string{f(aware.Add(24 * time.Hour)), f(due.Add(-time.Duration(r.between(6, 30)) * time.Hour)), f(aware.Add(30 * 24 * time.Hour))})
	return q
}

// ── legal hold scope ───────────────────────────────────────────────────

func holdScopeExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	custodians := people(r, 3)
	from := exDate(r)
	to := from.AddDate(0, r.between(4, 14), 0)
	topic := pick(r, []string{"Project " + pick(r, []string{"Vertex", "Flywheel", "Heimdall", "Legato", "Axiom"}), "the " + sl["VENDOR"] + " contract", "the 2019 pay equity review", "warehouse overtime"})
	docs := []struct {
		owner, topic string
		d            time.Time
	}{}
	hit := r.intn(4)
	others := people(r, 2)
	for i := 0; i < 4; i++ {
		if i == hit {
			docs = append(docs, struct {
				owner, topic string
				d            time.Time
			}{pick(r, custodians), topic, from.AddDate(0, 0, r.between(5, 90))})
			continue
		}
		switch r.intn(3) {
		case 0: // custodian, wrong topic
			docs = append(docs, struct {
				owner, topic string
				d            time.Time
			}{pick(r, custodians), "Holiday party planning", from.AddDate(0, 0, r.between(5, 90))})
		case 1: // right topic, before the period
			docs = append(docs, struct {
				owner, topic string
				d            time.Time
			}{pick(r, custodians), topic, from.AddDate(-1, -r.between(1, 6), 0)})
		default: // not a custodian, unrelated
			docs = append(docs, struct {
				owner, topic string
				d            time.Time
			}{pick(r, others), "Office supply orders", from.AddDate(0, 0, r.between(5, 90))})
		}
	}
	var rows [][]string
	var labels []string
	for i, d := range docs {
		rows = append(rows, []string{fmt.Sprintf("Doc %d", i+1), d.owner, d.topic, d.d.Format("01/02/2006")})
		labels = append(labels, fmt.Sprintf("Doc %d", i+1))
	}
	q := Question{key: "ex/holdscope", Exhibit: docExhibit("Legal Hold Notice — Matter "+fmt.Sprint(r.between(1000, 9999)), fmt.Sprintf("Custodians: %s · Topic: %s · Period: %s – %s", strings.Join(custodians, ", "), topic, from.Format("01/02/2006"), to.Format("01/02/2006")), []string{"Item", "Owner", "Subject", "Date"}, rows, "All four items are past their normal retention date and queued for deletion.")}
	q.Text = fmt.Sprintf("Refer to %s. Which item must be pulled from the deletion queue because it is covered by the legal hold?", ex)
	q.Choices, q.Correct = oneOf(r, labels[hit], labels)
	return q
}
