package compliance

import (
	"fmt"
	"strings"
)

// The phishing email exhibit: a mock message with zero to three planted
// warning signs (lookalike sender domain, link whose real destination
// differs, outside Reply-To, risky attachment type, external-sender banner
// on a message claiming to be internal), and several kinds of question
// about it, all answered from the fields shown.

type emailLure struct {
	Display, Local, Subject, Body, Path, Attach string
	Vendor                                      bool // sent from {VENDOR}'s domain, not acpwb.com
	Pushy                                       bool // a classic fraud pattern in itself: never shown as clean
}

var emailLures = []emailLure{
	{"ACPWB Payroll", "payroll", "[Action needed|Reminder]: [confirm|verify] your direct deposit [details|information]", "Our records show your direct deposit information [is incomplete|needs to be re-verified|was flagged during the {WEEKDAY} payroll run]. Please confirm your details before the next pay date to avoid a delay.", "payroll/direct-deposit", "Payroll_Notice.pdf", false, false},
	{"ACPWB Benefits", "benefits", "Your {YEAR} benefits [enrollment summary|confirmation statement] is ready", "Your [enrollment summary|benefits statement] is ready. Review your elections and confirm them before the window closes on {WEEKDAY}.", "benefits/summary", "Enrollment_Summary.pdf", false, false},
	{"IT Service Desk", "it.servicedesk", "Your [network|email] password expires in {N} hours", "Your [network|email] password is due to expire. Keep your current password by signing in to the self-service portal[? before {TIMEOFDAY}].", "password/keep", "Instructions.pdf", false, false},
	{"Facilities Management", "facilities", "[Badge reader upgrade|Access control update]: re-register your badge", "The {DOOR} readers were [upgraded|replaced] overnight. Re-register your badge [to keep your floor access|by {WEEKDAY}] using the link below.", "badges/register", "Badge_FAQ.pdf", false, false},
	{"ACPWB Compliance Office", "compliance", "Reminder: {COURSE} is available", "A new session of {COURSE} is open. Sign in to [start|resume] the course at your convenience.", "learning/start", "Course_Overview.pdf", false, false},
	{"Accounts Payable", "accounts.payable", "Invoice {FORM}[| — updated remittance details]", "Please find the attached invoice[? for services rendered in {CITY}]. Note our [new|updated] remittance details, effective [immediately|on {WEEKDAY}].", "invoices/view", "Invoice.pdf", true, true},
	{"Office of the CEO", "ceo.office", "[Quick favor|Are you at your desk?|Confidential request]", "I need you to handle something [quickly|discreetly] for a client event. [Reply with your mobile number.|Let me know when you see this.|I'm in meetings all day, so email only.]", "requests", "Event_Details.docx", false, true},
	{"People Operations", "people.ops", "[Updated|Revised] {DOCUMENT} policy — acknowledgement required", "Policy {POLICY} has been [updated|revised]. Please review the changes and record your acknowledgement[? within {DAYS} days].", "policies/acknowledge", "Policy_Changes.pdf", false, false},
	{"ACPWB Travel Desk", "travel", "Itinerary change for your trip to {CITY}", "Your [flight|hotel reservation] for the upcoming trip to {CITY} has changed. Review the new itinerary and confirm[? by {TIMEOFDAY}].", "travel/itinerary", "Itinerary.pdf", false, false},
	{"Parking Office", "parking", "Your parking permit [expires soon|has been suspended]", "Your parking structure permit [expires on {WEEKDAY}|was flagged in an audit]. Renew it online to keep your space on level P{N}.", "parking/renew", "Permit_Terms.pdf", false, false},
	{"PERCH Registration", "perch", "PERCH [registration|room block] confirmation", "Thank you for registering for PERCH. [Confirm your session choices|Complete your hotel booking|Upload your speaker photo] to finalize your registration.", "perch/confirm", "PERCH_Agenda.pdf", false, false},
	{"ACPWB Mailroom", "mailroom", "[A package|Two packages] could not be delivered to your desk", "[A package|A parcel] addressed to you is being held in the mailroom. Schedule a redelivery to the {FLOOR} floor.", "mailroom/redeliver", "Delivery_Notice.pdf", false, false},
	{"Corporate Security", "security", "[Unusual sign-in|New sign-in] detected on your account", "We detected a sign-in to your account from {CITY} at {TIMEOFDAY}. If this wasn't you, secure your account now.", "security/review", "Sign-in_Report.pdf", false, false},
	{"Legal Department", "legal", "Legal hold notice — [matter|case] {FORM}", "You have been identified as a custodian in matter {FORM}. Review the hold notice and confirm that you will preserve all relevant records.", "legal/hold", "Hold_Notice.pdf", false, false},
	{"Wellness Program", "wellness", "[Your wellness reward is waiting|Claim your wellness incentive]", "You have earned a {AMOUNT} wellness incentive. Claim it [before {WEEKDAY}|this week] by signing in to the wellness portal.", "wellness/claim", "Reward_Details.pdf", false, false},
	{"IT Service Desk", "it.servicedesk", "Multi-factor authentication [re-enrollment|reset] required", "Your multi-factor authentication device must be re-enrolled after [the {WEEKDAY} upgrade|a system migration]. Complete enrollment to avoid losing access.", "mfa/enroll", "MFA_Guide.pdf", false, false},
	{"Document Share", "share", "{NAME2} shared \"{DOCUMENT}\" with you", "{NAME2} has shared a document with you: {DOCUMENT}. [Sign in to view it.|Open it to review the changes.|The link expires in {N} days.]", "files/open", "Shared_File.pdf", true, false},
	{"Expense Audit", "expense.audit", "Expense report [returned|flagged] for review", "Your expense report for the {CITY} trip was [returned|flagged]. Upload the missing receipts to avoid a payroll deduction.", "expenses/review", "Audit_Notes.pdf", false, false},
	{"Employee Survey", "survey", "[Your voice matters|Last chance]: the {YEAR} engagement survey", "The annual engagement survey closes on {WEEKDAY}. Responses are confidential. [Take the survey now.|It takes about {N} minutes.]", "survey/start", "Survey_FAQ.pdf", false, false},
	{"Social Committee", "social.committee", "Holiday party RSVP — {PLACE}", "This year's holiday party is in the {PLACE}. RSVP and choose your meal [by {WEEKDAY}|this week].", "events/rsvp", "Menu.pdf", false, false},
	{"Facilities Management", "facilities", "Refrigerator clean-out [this {WEEKDAY}|reminder]", "The {FLOOR} floor refrigerator will be emptied on {WEEKDAY} at 3:00 p.m. Claim your items or reserve shelf space using the link below.", "facilities/fridge", "Cleanout_Schedule.pdf", false, false},
	{"Voicemail System", "voicemail", "New voicemail from {NAME2} ({DEPT})", "You have a new voicemail from {NAME2}. [Listen to it online.|The recording is attached.]", "voicemail/play", "Voicemail.wav", false, false},
	{"eSignature Service", "esign", "Please [sign|review and sign]: {DOCUMENT}", "{NAME2} sent you a document to sign: {DOCUMENT}. [Review and sign by {WEEKDAY}.|The signing link expires in {N} days.]", "sign/review", "Document.pdf", true, false},
	{"Print Services", "print", "You have [a held print job|{N} held print jobs]", "A print job is waiting at the {FLOOR} floor printer alcove. Release it from the web portal before it is deleted.", "print/release", "Print_Summary.pdf", false, false},
}

var emailFlagNames = []string{
	"The sender's address", "The link's actual destination", "The Reply-To address", "The attachment's file type", "The sender verification banner",
}

func vendorDomain(v string) string {
	var b strings.Builder
	for _, ch := range strings.ToLower(v) {
		if ch >= 'a' && ch <= 'z' {
			b.WriteRune(ch)
		}
	}
	return b.String() + ".com"
}

func lookalike(r *rnd, domain string) string {
	base := strings.TrimSuffix(domain, ".com")
	switch r.intn(6) {
	case 0:
		return base + ".co"
	case 1:
		return base + "-" + pick(r, []string{"secure", "portal", "login", "verify", "hr", "billing"}) + ".com"
	case 2:
		return domain + "." + pick(r, []string{"mail-secure.io", "auth-session.co", "verify-login.net", "sso-check.com"})
	case 3:
		for _, sw := range [][2]string{{"w", "vv"}, {"l", "1"}, {"o", "0"}, {"m", "rn"}, {"i", "1"}, {"b", "6"}} {
			if strings.Contains(base, sw[0]) {
				return strings.Replace(base, sw[0], sw[1], 1) + ".com"
			}
		}
		return base + "s.com"
	case 4:
		return pick(r, []string{"secure-", "my", "login-", "portal-"}) + base + ".com"
	}
	return base + ".net"
}

type emailMsg struct {
	from, display, replyTo, to, subject, body, linkText, linkReal, attach string
	external                                                              bool
	flags                                                                 []int
}

func buildEmail(r *rnd, sl slots) emailMsg {
	g := gen{r, sl}
	l := pick(r, emailLures)
	domain := "acpwb.com"
	linkHost := "intranet.acpwb.com"
	if l.Vendor {
		domain = vendorDomain(sl["VENDOR"])
		linkHost = "portal." + domain
	}
	m := emailMsg{
		display: l.Display, from: l.Local + "@" + domain,
		to:      strings.ToLower(strings.ReplaceAll(sl["NAME"], " ", ".")) + "@acpwb.com",
		subject: g.expand(l.Subject), body: g.expand(l.Body),
		linkText: "https://" + linkHost + "/" + l.Path, attach: l.Attach,
		external: l.Vendor,
	}
	if l.Vendor {
		m.display = sl["VENDOR"]
	}
	m.replyTo, m.linkReal = m.from, m.linkText
	k := weighted(r, []int{15, 55, 25, 5})
	if l.Pushy && k == 0 {
		k = 1
	}
	m.flags = sample(r, []int{0, 1, 2, 3, 4}, k)
	for _, f := range m.flags {
		switch f {
		case 0:
			m.from = l.Local + "@" + lookalike(r, domain)
			if !contains(intsToStrings(m.flags), "2") {
				m.replyTo = m.from
			}
			m.external = true
		case 1:
			m.linkReal = "http://" + lookalike(r, linkHost) + "/" + l.Path + "?u=" + md5Hex(sl["NAME"] + l.Path)[:6]
		case 2:
			m.replyTo = strings.ToLower(strings.ReplaceAll(sl["NAME2"], " ", ".")) + pick(r, []string{"@gmail.com", "@outlook.com", "@proton.me", "@yahoo.com", "@mail.ru", "@icloud.com"})
		case 3:
			stem := strings.TrimSuffix(m.attach, m.attach[strings.LastIndex(m.attach, "."):])
			m.attach = stem + pick(r, []string{".pdf.exe", ".zip", ".html", ".docm", ".pdf.js", ".iso", ".scr", ".xlsm", ".lnk"})
		case 4:
			m.external = !l.Vendor // claims to be internal but came from outside
			if l.Vendor {
				// A vendor message flagged as internal is not a tell; use the
				// lookalike sender instead.
				m.from = l.Local + "@" + lookalike(r, domain)
				m.replyTo = m.from
				m.flags = replaceInt(m.flags, 4, 0)
			}
		}
	}
	m.flags = uniqInts(m.flags)
	return m
}

func intsToStrings(xs []int) []string {
	var out []string
	for _, x := range xs {
		out = append(out, fmt.Sprint(x))
	}
	return out
}

func replaceInt(xs []int, from, to int) []int {
	out := append([]int(nil), xs...)
	for i, x := range out {
		if x == from {
			out[i] = to
		}
	}
	return out
}

func uniqInts(xs []int) []int {
	var out []int
	for _, x := range xs {
		dup := false
		for _, y := range out {
			dup = dup || x == y
		}
		if !dup {
			out = append(out, x)
		}
	}
	return out
}

func hostOf(u string) string {
	u = strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
	h, _, _ := strings.Cut(u, "/")
	return h
}

func emailHTML(m emailMsg, sl slots) string {
	banner := `<span class="cmp-ex-ok">✓ Internal sender (verified)</span>`
	if m.external {
		banner = `<span class="cmp-ex-warn">⚠ EXTERNAL: This message came from outside ACPWB.</span>`
	}
	return fmt.Sprintf(`<div class="cmp-ex-mail"><div class="cmp-ex-mailbar">Inbox — %s</div><div class="cmp-ex-banner">%s</div><dl><dt>From</dt><dd>%s &lt;%s&gt;</dd><dt>Reply-To</dt><dd>%s</dd><dt>To</dt><dd>%s</dd><dt>Subject</dt><dd><b>%s</b></dd><dt>Attachment</dt><dd>📎 %s</dd></dl><div class="cmp-ex-mailbody"><p>Hello %s,</p><p>%s</p><p><u>%s</u><br><span class="cmp-ex-hover">Link preview (on hover): %s</span></p><p>Thank you,<br>%s</p></div></div>`,
		esc(m.to), banner, esc(m.display), esc(m.from), esc(m.replyTo), esc(m.to), esc(m.subject), esc(m.attach), esc(sl["FIRST"]), esc(m.body), esc(m.linkText), esc(m.linkReal), esc(m.display))
}

func emailExhibit(r *rnd, c *Course, sl slots, ex string) Question {
	m := buildEmail(r, sl)
	q := Question{key: "ex/email", Exhibit: emailHTML(m, sl)}
	const none = "None of these; the message shows no warning signs"
	switch v := r.intn(5); {
	case v == 0 && len(m.flags) <= 1, v == 1 && len(m.flags) <= 1:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %[1]s. Which detail is the clearest sign that this message is a phishing attempt?",
			"Refer to %[1]s. Which part of this message should make you suspicious?",
			"Refer to %[1]s. Before acting on this message, which detail should you question?",
			"Refer to %[1]s. %[2]s forwards you this message and asks whether it is safe. Which detail, if any, is a warning sign?",
		}), ex, sl["NAME2"])
		var fields []int
		for _, f := range sample(r, []int{0, 1, 2, 3, 4}, 5) {
			// An EXTERNAL banner on a genuinely external message is not a
			// tell, so it can't be offered as a distractor either.
			if f == 4 && m.external {
				continue
			}
			if !contains(intsToStrings(m.flags), fmt.Sprint(f)) {
				fields = append(fields, f)
			}
		}
		if len(m.flags) == 0 {
			q.Choices, q.Correct = withChoices(r, none, []string{emailFlagNames[fields[0]], emailFlagNames[fields[1]], emailFlagNames[fields[2]]})
		} else {
			q.Choices, q.Correct = withChoices(r, emailFlagNames[m.flags[0]], []string{emailFlagNames[fields[0]], emailFlagNames[fields[1]], none})
		}
	case v == 2:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %s. Where does the link in this message actually lead?",
			"Refer to %s. If you click the link, which site will open?",
			"Refer to %s. Which domain does the link point to?",
		}), ex)
		real, shown := hostOf(m.linkReal), hostOf(m.linkText)
		wrongs := []string{}
		if shown != real {
			wrongs = append(wrongs, shown)
		}
		for len(wrongs) < 3 {
			if w := lookalike(r, shown); w != real && !contains(wrongs, w) {
				wrongs = append(wrongs, w)
			}
		}
		q.Choices, q.Correct = withChoices(r, real, wrongs)
	case v == 3:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %s. If you click Reply, who will receive your response?",
			"Refer to %s. Replies to this message will be delivered to:",
		}), ex)
		wrongs := []string{}
		for _, w := range []string{m.from, m.to, "compliance@acpwb.com", "security@acpwb.com", "it.servicedesk@acpwb.com", "noreply@acpwb.com"} {
			if w != m.replyTo && !contains(wrongs, w) && len(wrongs) < 3 {
				wrongs = append(wrongs, w)
			}
		}
		q.Choices, q.Correct = withChoices(r, m.replyTo, wrongs)
	case v == 4 && r.chance(0.5):
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %s. Check the sender address, Reply-To, link destination, attachment type, and sender banner. How many of these show a warning sign?",
			"Refer to %s. Counting only the sender address, Reply-To, link destination, attachment type, and sender banner, how many warning signs does this message show?",
		}), ex)
		q.Choices, q.Correct = numChoices(r, len(m.flags), []int{-2, -1, 1, 2, 3}, itoa)
	default:
		q.Text = fmt.Sprintf(pick(r, []string{
			"Refer to %s. What is the BEST action to take with this message?",
			"Refer to %s. This message arrives in your inbox {TIMEOFDAY}. What should you do?",
			"Refer to %s. Under the ACPWB Email Security Standard, you should:",
		}), ex)
		q.Text = sl.fill(q.Text)
		right := "Report it with the Report Phish button without clicking, opening, or replying"
		wrongs := []string{"Reply and ask the sender to confirm it is genuine", "Forward it to your team as a warning", "Open the attachment in a private browser window to check it"}
		if len(m.flags) == 0 {
			right, wrongs[2] = "Proceed; it passes every check covered in this course", right
		}
		q.Choices, q.Correct = withChoices(r, right, wrongs)
	}
	return q
}
