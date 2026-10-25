package compliance

import (
	"fmt"
	"strings"
)

// Flowchart exhibits: a chain of yes/no decisions, each of which can end
// in an outcome; the question describes a case and asks where it lands.
// Charts that summarize law follow it (FMLA 29 CFR 825.110, FLSA salary
// test, I-9 timing, GDPR Arts. 33–34, OSHA 1904.7 recordability, 1910.147
// and 1910.146 scope, California's ABC test, FAA lithium battery limits,
// NWS watch vs. warning); the rest are ACPWB policy.

type flowStep struct {
	q      string
	stopOn bool   // the answer that ends the chart here
	out    string // outcome when it ends here
	yes    []string
	no     []string // phrases describing the case; "" adds nothing
}

type flowSpec struct {
	title, subj, ask, final string
	steps                   []flowStep
	// numeric charts: value thresholds, highest first (step i asks "over
	// thresholds[i]?"), with show formatting and a case template.
	thresholds []float64
	show       func(float64) string
	cats       []string
}

func st(q string, stopOn bool, out string, yes, no []string) flowStep {
	return flowStep{q, stopOn, out, yes, no}
}

func ph(p ...string) []string { return p }

var flowSpecs = map[string]flowSpec{
	"fmlaflow": {title: "FMLA ELIGIBILITY", subj: "{N}", ask: "Is the leave FMLA-protected?", final: "Eligible: up to 12 weeks of job-protected leave",
		steps: []flowStep{
			st("Worked for ACPWB at least 12 months?", false, "Not eligible for FMLA", ph("has worked at ACPWB for three years", "joined ACPWB 20 months ago", "has 14 months of service"), ph("was hired 7 months ago", "started at ACPWB 10 months ago")),
			st("Worked at least 1,250 hours in the last 12 months?", false, "Not eligible for FMLA", ph("worked about 1,900 hours last year", "logged 1,400 hours in the past 12 months"), ph("works 20 hours a week, about 1,000 hours a year", "logged 980 hours in the past 12 months")),
			st("50 or more employees within 75 miles of the worksite?", false, "Not eligible for FMLA", ph("works at the Milwaukee headquarters with 600 coworkers", "works at a site with 120 employees"), ph("works at a 12-person office with no other ACPWB site within 75 miles")),
			st("Leave for a qualifying reason?", false, "Use PTO or the personal leave policy", ph("needs time off for surgery", "is expecting a baby", "must care for a parent with a serious health condition"), ph("wants six weeks off to backpack across Europe", "wants extra time off to renovate a kitchen")),
		}, cats: []string{"leave", "parental", "fmla", "accommodation", "managers"}},
	"flsaflow": {title: "EXEMPT OR NON-EXEMPT? (FLSA)", subj: "The position", ask: "How should the position be classified?", final: "Exempt from overtime",
		steps: []flowStep{
			st("Teacher, or licensed doctor or lawyer practicing the profession?", true, "Exempt (the salary test does not apply)", ph("is a staff attorney practicing law", "is a physician at the on-site clinic"), ph("")),
			st("Paid a salary of at least $684 per week?", false, "Non-exempt: overtime required", ph("pays $1,150 a week in salary", "pays an annual salary of $78,000"), ph("pays $640 a week in salary", "pays $15 an hour")),
			st("Duties meet an exemption test (executive, administrative, professional, computer)?", false, "Non-exempt: overtime required", ph("manages a team of six and hires staff", "is a software engineer designing systems"), ph("answers phones and schedules meetings", "processes invoices under close supervision")),
		}, cats: []string{"wage", "classification", "payroll", "timekeeping", "overtime"}},
	"i9flow": {title: "FORM I-9 — WHEN IS IT DUE?", subj: "{N}", ask: "When must the I-9 be completed?", final: "Section 1 by the first day of work; Section 2 within 3 business days",
		steps: []flowStep{
			st("Rehired within 3 years of the original Form I-9?", true, "You may use Supplement B of the original form", ph("is being rehired 2 years after leaving"), ph("is a new hire", "has never worked at ACPWB before")),
			st("Job will last fewer than 3 business days?", true, "Both sections by the end of the first day of work", ph("is hired for a two-day event", "is starting a one-day assignment"), ph("is starting a permanent role", "is starting a six-month contract role")),
		}, cats: []string{"i9", "hiring", "onboarding"}},
	"dataclassflow": {title: "DATA CLASSIFICATION DECISION TREE", subj: "The spreadsheet", ask: "How should it be classified?", final: "Public",
		steps: []flowStep{
			st("Contains SSNs, bank, or health data?", true, "Restricted", ph("includes employee Social Security numbers", "lists bank routing numbers"), ph("has no government ID, bank, or health data")),
			st("Contains personal data about customers or employees?", true, "Confidential", ph("lists customer names and email addresses", "includes employee home addresses"), ph("has no personal data about anyone")),
			st("Meant only for ACPWB staff?", true, "Internal", ph("is marked for internal distribution", "is the internal org chart"), ph("is approved for the public website")),
		}, cats: []string{"classification", "privacy", "records", "customerdata"}},
	"giftflow": {title: "GIFT ACCEPTANCE", subj: "The gift", ask: "What should you do with it?", final: "Accept and log it in the gift register",
		steps: []flowStep{
			st("From a government official or a vendor in an active bid?", true, "Decline the gift", ph("comes from a vendor bidding on our janitorial contract", "comes from a city inspector"), ph("comes from a long-time customer with no open deals")),
			st("Cash or cash equivalent?", true, "Decline the gift", ph("is a $50 Visa gift card", "is an envelope of cash"), ph("is a fruit basket", "is a pair of concert tickets")),
			st("Worth more than $100?", true, "Report to Ethics before accepting", ph("is worth about $250", "is valued at $180"), ph("is worth about $40", "is valued at $25")),
		}, cats: []string{"gifts", "bribery", "ethics", "procurement"}},
	"changeflow": {title: "CHANGE MANAGEMENT — WHICH PATH?", subj: "The change", ask: "Which approval path applies?", final: "Normal change: peer review and manager approval",
		steps: []flowStep{
			st("Is production down or at risk right now?", true, "Emergency change: ECAB approval", ph("restores the payroll system after an outage"), ph("is planned for next month", "is not urgent")),
			st("Pre-approved standard change in the catalog?", true, "Standard change: no CAB review", ph("is a routine password reset template", "is a catalog item: add a printer"), ph("is a custom configuration")),
			st("Could it affect customers or other systems?", true, "Normal change: full CAB review", ph("modifies the customer portal's login", "touches the shared database cluster"), ph("only changes an internal report's font")),
		}, cats: []string{"changemgmt", "releases", "network"}},
	"receiptflow": {title: "EXPENSE RECEIPT RULES", subj: "The expense", ask: "What documentation is required?", final: "Receipt optional; record date, amount, place, and purpose",
		steps: []flowStep{
			st("Claimed as a per diem meal allowance?", true, "No receipt needed under per diem", ph("is a per diem meal allowance"), ph("is an actual-cost expense")),
			st("Is it lodging?", true, "Itemized hotel folio required", ph("is a hotel stay", "is two nights at a conference hotel"), ph("is a taxi ride", "is a team dinner")),
			st("$75 or more?", true, "Receipt required", ph("costs $112", "comes to $96.40"), ph("costs $38", "comes to $21.75")),
		}, cats: []string{"reimburse", "travel", "pcard", "finance"}},
	"breachflow": {title: "PERSONAL DATA BREACH — GDPR NOTIFICATION", subj: "The incident", ask: "What notification does GDPR require?", final: "Notify the supervisory authority within 72 hours",
		steps: []flowStep{
			st("Was personal data affected?", false, "Log as a security incident; not a personal data breach", ph("exposed a list of EU customers' email addresses", "involved stolen HR files"), ph("involved only an encrypted marketing deck with no personal data")),
			st("Unlikely to result in any risk to people?", true, "Document it internally; no notification", ph("involved data that was strongly encrypted with the key still safe"), ph("involved readable data")),
			st("Likely to result in a HIGH risk to people?", true, "Notify the authority within 72 hours AND tell the people affected", ph("exposed passport scans and bank details", "exposed health records"), ph("exposed only work email addresses")),
		}, cats: []string{"privacy2", "privacy", "breach", "gdpr", "ransomware"}},
	"hotworkflow": {title: "HOT WORK PERMIT (NFPA 51B)", subj: "The job", ask: "What is the result?", final: "Issue the hot work permit",
		steps: []flowStep{
			st("Can it be moved to the weld shop or done by a cold method?", true, "Do it there; no permit needed", ph("can be done with a bolt-on bracket instead of welding", "involves a part that can go to the weld shop"), ph("is a pipe fixed in the ceiling")),
			st("Can combustibles within 35 ft be removed or shielded?", false, "Hot work not allowed", ph("is near cardboard that can be moved", "is in a room cleared of combustibles"), ph("is next to a foam wall that cannot be covered")),
			st("Sprinklers in service and a fire watch assigned?", false, "Hot work not allowed", ph("has sprinklers in service and a trained fire watch", "has a fire watch assigned for the job"), ph("is in an area where the sprinklers are shut off for repair")),
		}, cats: []string{"hotwork", "fire", "construction"}},
	"permitflow": {title: "CONFINED SPACE CLASSIFICATION (OSHA 1910.146)", subj: "The space", ask: "How should the space be classified?", final: "Permit-required confined space",
		steps: []flowStep{
			st("Large enough to enter, limited entry/exit, not designed for continuous occupancy?", false, "Not a confined space", ph("is a storm vault reached by a ladder", "is an empty water tank with a manhole"), ph("is a normal storage room with a door", "is a walk-in cooler")),
			st("Has (or could have) a hazardous atmosphere, engulfment, trapping walls, or other serious hazard?", false, "Non-permit confined space", ph("may collect hydrogen sulfide", "contains loose grain"), ph("is dry, ventilated, and has no hazards")),
		}, cats: []string{"spaces", "construction", "warehouse"}},
	"lotoflow": {title: "DOES LOCKOUT/TAGOUT APPLY? (OSHA 1910.147)", subj: "The task", ask: "What does the standard require?", final: "Apply lockout/tagout before starting",
		steps: []flowStep{
			st("Servicing or maintenance, or bypassing a guard?", false, "Normal production: lockout not required", ph("is clearing a jam inside the guard", "is replacing a drive belt"), ph("is loading paper during a normal print run")),
			st("Cord-and-plug equipment, unplugged, plug in your exclusive control?", true, "Exempt: unplug it and keep the plug with you", ph("is on a plug-in bench grinder kept unplugged next to the worker"), ph("is on a hard-wired conveyor")),
		}, cats: []string{"lockout", "machineguarding", "electrical"}},
	"recordflow": {title: "OSHA 300 LOG — IS IT RECORDABLE?", subj: "The injury", ask: "How should it be logged?", final: "Not recordable: first aid only",
		steps: []flowStep{
			st("Is it work-related?", false, "Not recordable: not work-related", ph("happened while stocking shelves", "happened on the loading dock"), ph("happened at home over the weekend")),
			st("Death, days away, restricted work or transfer, or loss of consciousness?", true, "Recordable on the OSHA 300 log", ph("kept the worker home for two days", "led to light duty for a week"), ph("did not cause any lost time")),
			st("Medical treatment beyond first aid?", true, "Recordable on the OSHA 300 log", ph("required stitches", "needed a prescription medication"), ph("was treated with a bandage and ice")),
		}, cats: []string{"injuries", "safety", "workerscomp", "nearmiss"}},
	"wireflow": {title: "OUTGOING WIRE VERIFICATION", subj: "The wire request", ask: "What should Accounts Payable do?", final: "Process the wire normally",
		steps: []flowStep{
			st("New payee or changed bank details?", true, "Call back on a number already on file", ph("asks us to pay a new account number", "updates the vendor's bank details"), ph("goes to a long-standing vendor account")),
			st("Comes with urgency or secrecy?", true, "Escalate to Treasury and verify independently", ph("is marked urgent and confidential", "says not to tell anyone"), ph("follows the normal payment schedule")),
			st("Over $10,000?", true, "Require dual approval", ph("is for $48,000", "is for $12,500"), ph("is for $3,200")),
		}, cats: []string{"wires", "fraud", "finance", "socialeng"}},
	"socialflow": {title: "BEFORE YOU POST", subj: "The post", ask: "What should happen with the post?", final: "Post freely; normal conduct rules apply",
		steps: []flowStep{
			st("Mentions unreleased products, earnings, or customers?", true, "Do not post", ph("mentions next quarter's numbers", "teases an unannounced product"), ph("discusses nothing confidential")),
			st("Speaking on behalf of ACPWB?", true, "Get Communications approval first", ph("is for the official company page"), ph("is on your personal account")),
			st("Do you say you work at ACPWB?", true, "Post with an 'opinions are my own' disclosure", ph("praises ACPWB's products and names your employer"), ph("is about your weekend hike")),
		}, cats: []string{"social", "brand", "disclosure", "media"}},
	"coiflow": {title: "CONFLICT OF INTEREST CHECK", subj: "{N}'s situation", ask: "What must happen?", final: "No disclosure needed",
		steps: []flowStep{
			st("Could it influence a business decision you make?", true, "Disclose and recuse yourself", ph("involves choosing a vendor owned by a cousin"), ph("does not touch any decision you make")),
			st("Involves an ACPWB vendor, customer, or competitor?", true, "Disclose to Ethics for review", ph("involves a side job for a competitor", "involves serving on a customer's advisory board"), ph("involves a local food pantry")),
			st("Uses company time or assets?", true, "Not permitted", ph("uses the company laptop on work hours"), ph("happens on weekends with personal equipment")),
		}, cats: []string{"conflicts", "ethics", "conduct", "moonlighting"}},
	"abcflow": {title: "EMPLOYEE OR CONTRACTOR? (CALIFORNIA ABC TEST)", subj: "The worker", ask: "How must the worker be classified?", final: "Independent contractor",
		steps: []flowStep{
			st("Does a statutory exemption apply (e.g., certain licensed professionals)?", true, "Apply the Borello test instead", ph("is a licensed physician"), ph("")),
			st("A: Free from ACPWB's control and direction?", false, "Employee", ph("sets their own methods and hours"), ph("follows a schedule set by an ACPWB manager")),
			st("B: Work outside the usual course of ACPWB's business?", false, "Employee", ph("fixes the office plumbing"), ph("writes research reports like our staff analysts")),
			st("C: Independently established in that trade?", false, "Employee", ph("runs a plumbing business with other clients"), ph("has no other clients or business of their own")),
		}, cats: []string{"classification", "gig", "contractors", "wage"}},
	"exportflow": {title: "EXPORT SCREENING", subj: "The shipment", ask: "What is the outcome?", final: "Ship with no license required",
		steps: []flowStep{
			st("Destination under comprehensive embargo?", true, "Do not ship: escalate to Trade Compliance", ph("is going to North Korea", "is going to Iran"), ph("is going to Canada", "is going to Germany")),
			st("Any party on a restricted party list?", true, "Hold for Trade Compliance review", ph("names a consignee on the Entity List"), ph("clears restricted party screening")),
			st("Does the item's classification need a license for that country?", true, "Apply for an export license", ph("contains controlled encryption that requires a license there"), ph("is ordinary office furniture")),
		}, cats: []string{"export", "sanctions", "supplychain", "customs"}},
	"lostbadgeflow": {title: "BADGE PROBLEMS", subj: "{N}'s badge", ask: "What should happen next?", final: "No action needed",
		steps: []flowStep{
			st("Lost or stolen?", true, "Report to Security now so it is deactivated", ph("was stolen from a car", "cannot be found after a conference"), ph("is not lost")),
			st("Left at home?", true, "Get a one-day temporary badge at the lobby", ph("is sitting on the kitchen counter at home"), ph("is with them")),
			st("Damaged or not reading?", true, "Request a replacement badge", ph("is cracked and no longer opens doors"), ph("works on every reader")),
		}, cats: []string{"badge", "physicalsec", "security"}},
	"mediaflow": {title: "WHO HANDLES THIS CALL?", subj: "The caller", ask: "Where should the call go?", final: "Handle under normal customer service",
		steps: []flowStep{
			st("Reporter, blogger, or podcaster?", true, "Corporate Communications; do not comment", ph("says she writes for a trade magazine", "hosts an HR podcast"), ph("is not with the media")),
			st("Regulator or government agency?", true, "Legal", ph("is an investigator from a state agency"), ph("is not from any agency")),
			st("Analyst or investor?", true, "Investor Relations", ph("covers ACPWB for a brokerage"), ph("is not an investor")),
		}, cats: []string{"media", "crisiscomms", "disclosure", "brand"}},
	"tornadoflow": {title: "SEVERE WEATHER — WHAT DO I DO?", subj: "This afternoon,", ask: "What should employees do?", final: "Normal operations",
		steps: []flowStep{
			st("Outdoor siren sounding or a building alert issued?", true, "Go to the shelter area now", ph("the outdoor sirens are sounding"), ph("there are no sirens")),
			st("Tornado WARNING for our area?", true, "Go to the shelter area now", ph("the National Weather Service issued a tornado warning for our county"), ph("there is no warning")),
			st("Tornado WATCH for our area?", true, "Stay alert and review shelter locations", ph("a tornado watch covers the region"), ph("there is no watch")),
		}, cats: []string{"weather", "closures", "evacassist"}},
	"lithiumflow": {title: "CAN I FLY WITH THIS BATTERY? (FAA)", subj: "The battery", ask: "What do FAA rules allow?", final: "Allowed: carry-on, or checked only if installed, off, and protected",
		steps: []flowStep{
			st("Over 160 Wh?", true, "Not allowed on passenger aircraft", ph("is rated at 190 Wh", "is rated at 250 Wh"), ph("")),
			st("101 to 160 Wh?", true, "Airline approval required; spares carry-on only, limit two", ph("is rated at 130 Wh", "is rated at 150 Wh"), ph("is rated at 74 Wh", "is rated at 27 Wh")),
			st("A spare battery or power bank?", true, "Carry-on only, with terminals protected", ph("is a spare power bank", "is a loose spare battery"), ph("is installed in a laptop")),
		}, cats: []string{"travel", "dgr", "cabin", "lithium"}},
	"harassflow": {title: "REPORTING HARASSMENT AND DISCRIMINATION", subj: "The behavior", ask: "What is the right next step?", final: "Use the open door policy for any other concern",
		steps: []flowStep{
			st("Is anyone in immediate danger?", true, "Call 911, then Security", ph("includes a threat of violence happening now"), ph("involves no immediate danger")),
			st("Sexual, or based on a protected characteristic?", true, "Report to HR or the Ethics Hotline for investigation", ph("involves repeated comments about a coworker's body", "involves jokes about a coworker's religion", "involves unwanted touching"), ph("has nothing to do with sex or any protected characteristic")),
			st("Rude or disrespectful conduct?", true, "Address it directly or talk to your manager", ph("is a coworker interrupting everyone in meetings"), ph("is a disagreement about a project plan")),
		}, cats: []string{"harassment", "sexharassment", "civility", "eeo", "bystander", "respect"}},
	"bystanderflow": {title: "BYSTANDER INTERVENTION — THE 5 Ds", subj: "At the offsite,", ask: "Which response does the chart recommend?", final: "Delegate, Delay, or Document: get help and check in afterward",
		steps: []flowStep{
			st("Is someone in danger right now?", true, "Call 911 or Security", ph("a guest is being physically blocked from leaving"), ph("no one is in physical danger")),
			st("Safe to speak up directly?", true, "Direct: calmly name the behavior", ph("you know both people and feel safe speaking up"), ph("you do not feel safe confronting the person")),
			st("Can you create a distraction?", true, "Distract: interrupt the situation", ph("you could invite the target to help you with something"), ph("there is no natural way to interrupt")),
		}, cats: []string{"bystander", "harassment", "sexharassment", "respect", "offsites"}},
	"photoflow": {title: "PHOTO AND VIDEO CONSENT", subj: "The photo", ask: "What consent is needed?", final: "Get verbal consent and honor opt-outs",
		steps: []flowStep{
			st("Includes a minor?", true, "Written consent from a parent or guardian", ph("shows a child at Take Your Kids to Work Day"), ph("shows only adults")),
			st("For advertising or the public website?", true, "Signed model release", ph("will appear in a recruiting ad"), ph("is for the internal newsletter")),
		}, cats: []string{"photos", "privacy", "kids", "brand"}},
	"aiflow": {title: "USING GENERATIVE AI", subj: "The task", ask: "What does the AI policy require?", final: "Allowed under the AI acceptable use policy",
		steps: []flowStep{
			st("Prompt includes confidential or personal data?", true, "Use only the approved enterprise AI tool", ph("pastes in a customer contract", "includes employee performance notes"), ph("uses only public information")),
			st("Output published or sent to a customer?", true, "Human review and fact-check required", ph("drafts a customer-facing FAQ"), ph("is a private brainstorm")),
			st("Code going into production?", true, "Normal code review plus license scan", ph("generates a payroll script for production"), ph("writes no code")),
		}, cats: []string{"ai", "chatbots", "genai", "acceptableuse"}},
	"travelflow": {title: "TRAVEL APPROVAL", subj: "The trip", ask: "What approval is needed?", final: "Book through the travel portal",
		steps: []flowStep{
			st("Destination on the high-risk country list?", true, "Security briefing and VP approval", ph("goes to a country on the high-risk list"), ph("goes to a low-risk destination")),
			st("International?", true, "Manager approval and travel registration", ph("goes to Toronto", "goes to London"), ph("stays in the United States")),
			st("Airfare over $1,000?", true, "Manager approval", ph("has a $1,340 fare"), ph("has a $280 fare")),
		}, cats: []string{"travel", "intltravel", "travelbooking"}},
	"rtoflow": {title: "OFFICE ATTENDANCE EXCEPTIONS", subj: "{N}", ask: "What is the attendance expectation?", final: "Three office days per week",
		steps: []flowStep{
			st("Approved ADA accommodation?", true, "Follow the accommodation plan", ph("has an approved accommodation to work from home"), ph("has no accommodation on file")),
			st("Role designated fully remote?", true, "No office days required", ph("holds a designated remote role"), ph("holds a hybrid role")),
			st("Temporary hardship under 30 days?", true, "Manager may approve a short exception", ph("has a two-week family emergency"), ph("has no special circumstances")),
		}, cats: []string{"rto", "hybrid", "remote", "accommodation"}},
	"accomflow": {title: "ADA INTERACTIVE PROCESS", subj: "{N}", ask: "What should the manager do?", final: "Document the process and consult HR and Legal",
		steps: []flowStep{
			st("Asked for a change at work because of a medical condition (in any words)?", false, "No accommodation request yet", ph("asked for a standing desk because of a back condition", "said a migraine condition makes the lights hard to work under"), ph("asked for a bigger monitor because the old one is ugly")),
			st("Requested accommodation effective and not an undue hardship?", true, "Provide the accommodation", ph("asked for something inexpensive that works"), ph("asked for something that would require rebuilding the floor")),
			st("Is there another effective accommodation?", true, "Offer the alternative", ph("could be helped by a different, reasonable change"), ph("has no other workable option identified")),
		}, cats: []string{"accommodation", "accessibility", "eeo", "managers"}},
	"ossflow": {title: "OPEN SOURCE LICENSE REVIEW", subj: "The library", ask: "What should the engineer do?", final: "Submit to the Open Source Review Board",
		steps: []flowStep{
			st("Approved permissive license (MIT, BSD, Apache 2.0)?", true, "Use it and keep the notices", ph("is MIT-licensed", "uses the Apache 2.0 license"), ph("is not under a permissive license")),
			st("Strong copyleft (GPL, AGPL)?", true, "Legal review before shipping or hosting", ph("is under the AGPL", "is GPL-3.0"), ph("is not copyleft")),
			st("No license at all?", true, "Do not use: no license means no permission", ph("has no license file anywhere"), ph("has an unusual custom license")),
		}, cats: []string{"opensource", "ip", "copyright", "licensing"}},
	"retainflow": {title: "CAN THIS BE DESTROYED?", subj: "The record", ask: "What does the retention policy say?", final: "Follow the department retention schedule",
		steps: []flowStep{
			st("Subject to a legal hold?", true, "Keep it: do not destroy", ph("is covered by a legal hold"), ph("is not under any hold")),
			st("Tax or financial record?", true, "Keep for 7 years", ph("is a vendor invoice", "is a tax return workpaper"), ph("is not a financial record")),
			st("Former employee's personnel file?", true, "Keep for 6 years after separation", ph("is a former employee's personnel file"), ph("is a meeting agenda")),
		}, cats: []string{"records", "retention", "legalhold", "doccontrol"}},
	"spillflow": {title: "CHEMICAL SPILL RESPONSE", subj: "The spill", ask: "What is the first response?", final: "Isolate the area and call EHS",
		steps: []flowStep{
			st("Anyone injured or splashed?", true, "Call 911 and use the eyewash or shower", ph("splashed a worker's face"), ph("injured no one")),
			st("Large, unknown, or highly hazardous?", true, "Evacuate and call the emergency team", ph("is an unknown liquid", "is five gallons of acid"), ph("is a small cup of a known cleaner")),
			st("Trained, with a spill kit at hand?", true, "Clean it up with the kit, wearing PPE", ph("happened near a stocked spill kit and a trained worker"), ph("has no trained person nearby")),
		}, cats: []string{"chemical", "hazcom", "lab", "ppe"}},
	"visitorflow": {title: "VISITOR ACCESS", subj: "The visitor", ask: "What access should the visitor get?", final: "Standard visitor badge, escorted",
		steps: []flowStep{
			st("Government inspector with credentials?", true, "Call Legal and EHS, then escort", ph("is an OSHA compliance officer"), ph("is not an inspector")),
			st("Entering a restricted area?", true, "Escorted badge plus signed NDA", ph("needs access to the server room"), ph("only needs a conference room")),
			st("Delivery only?", true, "Leave it at reception; no badge", ph("is dropping off lunch"), ph("is attending a meeting")),
		}, cats: []string{"visitor", "physicalsec", "badge", "lobby"}},
	"concernflow": {title: "RAISING A CONCERN", subj: "{N}", ask: "Where should the concern go?", final: "Raise it with your manager",
		steps: []flowStep{
			st("Fear retaliation or want to stay anonymous?", true, "Anonymous Ethics Hotline", ph("is worried about retaliation", "wants to stay anonymous"), ph("is comfortable being identified")),
			st("Is your manager involved?", true, "HR, Legal, or your manager's manager", ph("thinks the manager is involved"), ph("has a manager who is not involved")),
		}, cats: []string{"whistleblower", "ethics", "retaliation", "speakup"}},
	"sevflow": {title: "INCIDENT SEVERITY TRIAGE", subj: "The incident", ask: "What severity should it be assigned?", final: "Sev 4: ticket in the queue",
		steps: []flowStep{
			st("Customer data exposed?", true, "Sev 1: page the incident commander", ph("exposed customer records"), ph("exposed no data")),
			st("Production service down?", true, "Sev 2", ph("took the customer portal offline"), ph("left all services up")),
			st("Degraded for users?", true, "Sev 3", ph("slowed page loads for users"), ph("affects only an internal test box")),
		}, cats: []string{"incident", "tickets", "sla", "ransomware"}},
	"purchaseflow": {title: "PURCHASE APPROVAL — DELEGATION OF AUTHORITY", subj: "A purchase order for", ask: "Who must approve it?", final: "Manager approval",
		steps: []flowStep{
			{q: "Over $250,000?", stopOn: true, out: "CFO approval"},
			{q: "Over $50,000?", stopOn: true, out: "Vice President approval"},
			{q: "Over $5,000?", stopOn: true, out: "Director approval"},
		}, thresholds: []float64{250000, 50000, 5000}, show: func(v float64) string { return dollars(int(v)) },
		cats: []string{"procurement", "authority", "finance"}},
	"flightclassflow": {title: "AIR TRAVEL CLASS OF SERVICE", subj: "A flight of", ask: "What class may be booked?", final: "Economy",
		steps: []flowStep{
			{q: "Over 10 hours scheduled?", stopOn: true, out: "Business class"},
			{q: "Over 6 hours scheduled?", stopOn: true, out: "Premium economy"},
		}, thresholds: []float64{10, 6}, show: func(v float64) string { return fmt.Sprintf("%.1f hours", v) },
		cats: []string{"travel", "travelbooking", "intltravel"}},
}

func init() {
	gens := map[string]exhibitGen{}
	cats := map[string][]string{}
	for k, sp := range flowSpecs {
		gens[k] = flowGen(k, sp)
		for _, c := range sp.cats {
			cats[c] = append(cats[c], k)
		}
	}
	for _, ks := range cats {
		sortStrings(ks)
	}
	registerExhibits(gens, cats)
}

func yn(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func (s *svgB) textBlock(cx, cy, size float64, attrs string, lines []string) {
	lh := size * 1.25
	y := cy - lh*float64(len(lines)-1)/2 + size*0.35
	for _, l := range lines {
		s.text(cx, y, size, `text-anchor="middle" `+attrs, l)
		y += lh
	}
}

func joinAnd(p []string) string {
	switch len(p) {
	case 0:
		return ""
	case 1:
		return p[0]
	case 2:
		return p[0] + " and " + p[1]
	}
	return strings.Join(p[:len(p)-1], ", ") + ", and " + p[len(p)-1]
}

func flowGen(kind string, sp flowSpec) exhibitGen {
	n := len(sp.steps)
	var outs []string
	for _, stp := range sp.steps {
		if !contains(outs, stp.out) {
			outs = append(outs, stp.out)
		}
	}
	if !contains(outs, sp.final) {
		outs = append(outs, sp.final)
	}
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		k := r.intn(n + 1) // where the case stops; n = falls through to final
		right := sp.final
		if k < n {
			right = sp.steps[k].out
		}
		var caseText string
		if sp.thresholds != nil {
			var v float64
			switch {
			case k == 0:
				v = sp.thresholds[0] * (1.1 + r.float()*1.5)
			case k < n:
				lo, hi := sp.thresholds[k], sp.thresholds[k-1]
				v = lo + (hi-lo)*(0.1+0.8*r.float())
			default:
				v = sp.thresholds[n-1] * (0.15 + 0.75*r.float())
			}
			if sp.thresholds[0] >= 1000 {
				v = float64(int(v/50) * 50)
			} else {
				v = float64(int(v*4)) / 4
			}
			caseText = sp.subj + " " + sp.show(v) + "."
		} else {
			var phrases []string
			for i := 0; i <= k && i < n; i++ {
				stp := sp.steps[i]
				ans := !stp.stopOn
				if i == k {
					ans = stp.stopOn
				}
				pool := stp.no
				if ans {
					pool = stp.yes
				}
				if p := pick(r, pool); p != "" {
					phrases = append(phrases, p)
				}
			}
			subj := strings.ReplaceAll(sp.subj, "{N}", personName(r))
			caseText = subj + " " + joinAnd(phrases) + "."
			if len(phrases) == 0 {
				caseText = ""
			}
		}

		const dx, dw, dh, rowH = 150.0, 270.0, 54.0, 74.0
		h := 46 + rowH*float64(n) + 52
		s := frame(540, h, sp.title+" flowchart", sp.title)
		s.raw(`<defs><marker id="fa-%s" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto"><path d="M0 0 L8 4 L0 8 Z" fill="#334155"/></marker></defs>`, kind)
		arrow := func(x1, y1, x2, y2 float64) {
			s.raw(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#334155" stroke-width="1.5" marker-end="url(#fa-%s)"/>`, x1, y1, x2, y2, kind)
		}
		for i, stp := range sp.steps {
			cy := 46 + rowH*float64(i) + dh/2
			s.raw(`<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="#eff6ff" stroke="#1d4ed8" stroke-width="1.5"/>`,
				dx, cy-dh/2, dx+dw/2, cy, dx, cy+dh/2, dx-dw/2, cy)
			s.textBlock(dx, cy, 8, `fill="#0f172a"`, wrapText(stp.q, 38))
			arrow(dx+dw/2, cy, 336, cy)
			s.text(dx+dw/2+8, cy-5, 8, `font-weight="700" fill="#1d4ed8"`, yn(stp.stopOn))
			lines := wrapText(stp.out, 30)
			bh := 12*float64(len(lines)) + 12
			s.rect(338, cy-bh/2, 190, bh, 6, "#fef9c3", "#ca8a04")
			s.textBlock(433, cy, 8, `font-weight="700" fill="#713f12"`, lines)
			arrow(dx, cy+dh/2, dx, cy+rowH-dh/2-2+(map[bool]float64{true: 6, false: 0}[i == n-1]))
			s.text(dx+6, cy+dh/2+13, 8, `font-weight="700" fill="#1d4ed8"`, yn(!stp.stopOn))
		}
		fy := 46 + rowH*float64(n) + 4
		lines := wrapText(sp.final, 44)
		s.rect(dx-dw/2, fy, dw, 16+12*float64(len(lines)-1)+10, 6, "#dcfce7", "#16a34a")
		s.textBlock(dx, fy+(26+12*float64(len(lines)-1))/2, 8, `font-weight="700" fill="#14532d"`, lines)
		q := Question{key: "ex/" + kind, Exhibit: s.done()}
		if caseText == "" {
			caseText = "Nothing in the chart applies to the case."
		}
		q.Text = fmt.Sprintf("Refer to %s. %s Following the flowchart: %s", ex, caseText, sp.ask)
		var wrongs []string
		for _, o := range outs {
			if o != right {
				wrongs = append(wrongs, o)
			}
		}
		if len(wrongs) > 3 {
			wrongs = sample(r, wrongs, 3)
		}
		q.Choices, q.Correct = withChoices(r, right, wrongs)
		return q
	}
}
