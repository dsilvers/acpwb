package archive

import (
	"fmt"
	"strings"

	"acpwb_go/data"
	"acpwb_go/pyrandom"
)

// Minutes content pools (apps/honeypot/archive_data_minutes.py via
// export_render_data2.py).
var (
	committeeNames           = data.Strings("COMMITTEE_NAMES")
	meetingLocations         = data.Strings("MEETING_LOCATIONS")
	committeeRoles           = data.Strings("COMMITTEE_ROLES")
	agendaItemTitles         = data.Strings("AGENDA_ITEM_TITLES")
	agendaDiscussionTemplate = data.Strings("AGENDA_DISCUSSION_TEMPLATES")
	resolutionTemplates      = data.Strings("RESOLUTION_TEMPLATES")
	motionVerbs              = data.Strings("MOTION_VERBS")
	actionItemTemplates      = data.Strings("ACTION_ITEM_TEMPLATES")
)

// CommitteeMember is one attendee row.
type CommitteeMember struct {
	Name    string
	Title   string
	Role    string
	Present bool
}

// Motion is the optional formal motion attached to an agenda item.
type Motion struct {
	Verb       string
	Text       string
	MovedBy    string
	SecondedBy string
	Yea        int
	Nay        int
	Abstain    int
	Carried    bool
}

// AgendaItem is one agenda entry.
type AgendaItem struct {
	Number     int
	Title      string
	Discussion string
	Motion     *Motion
	Exhibit    string
}

// MinutesActionItem is one action-item row.
type MinutesActionItem struct {
	Number      int
	Description string
	Owner       string
	DueDate     string
}

// MinutesContent is the Go equivalent of _generate_minutes_content()'s
// return dict (apps/honeypot/views.py:459).
type MinutesContent struct {
	Title       string
	Org         string
	Industry    string
	RecordID    string
	Committee   string
	Location    string
	CallToOrder string
	AdjournTime string
	MeetingRef  string
	EngCode     string
	DateStr     string
	Members     []CommitteeMember
	Quorum      bool
	TotalSeats  int
	NumPresent  int
	Secretary   CommitteeMember
	Items       []AgendaItem
	ActionItems []MinutesActionItem
	NextMeeting string
	BulkHexJS   []string
	BulkHexCSS  []string
}

// GenerateMinutesContent ports apps/honeypot/views.py:_generate_minutes_content.
func GenerateMinutesContent(year, month, day int, slug string) MinutesContent {
	rng := rngA(fmt.Sprintf("minutes_%d%d%d%s", year, month, day, slug))

	org := choice(rng, archiveOrgs)
	industry := choice(rng, archiveIndustries)
	dateStr := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	q := int(rng.RandInt(1, 4))
	n := int(rng.RandInt(12, 180))
	n2 := int(rng.RandInt(10, 50))
	regions := int(rng.RandInt(2, 18))
	pct := int(rng.RandInt(4, 20))
	projectName := choice(rng, complianceProjectNames)

	committee := choice(rng, committeeNames)
	location := choice(rng, meetingLocations)
	hour := int(rng.RandInt(8, 16))
	minuteOfHour := choice(rng, []int64{0, 15, 30, 45})
	ampm := "AM"
	if hour >= 12 {
		ampm = "PM"
	}
	displayHour := hour
	if hour > 12 {
		displayHour = hour - 12
	}
	callToOrder := fmt.Sprintf("%d:%02d %s CT", displayHour, minuteOfHour, ampm)
	adjHour := hour + int(rng.RandInt(1, 3))
	adjAmpm := "AM"
	if adjHour >= 12 {
		adjAmpm = "PM"
	}
	adjDisplay := adjHour
	if adjHour > 12 {
		adjDisplay = adjHour - 12
	}
	adjournTime := fmt.Sprintf("%d:%02d %s CT", adjDisplay, choice(rng, []int64{0, 15, 30, 45}), adjAmpm)

	meetingRef := fmt.Sprintf("MIN-%d-%02d-%d", year, month, rng.RandInt(1000, 9999))
	engCode := fmt.Sprintf("ENG-%d-%s-%d", year, choice(rng, engagementCodes), rng.RandInt(10000, 99999))

	// Title from slug
	tail := slug
	if slug != "" {
		parts := strings.Split(slug, "/")
		tail = parts[len(parts)-1]
	} else {
		tail = fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}
	cleanTail := trailingNumericID.ReplaceAllString(tail, "")
	title := committee + " — " + pyTitle(strings.ReplaceAll(cleanTail, "-", " "))

	// Attendance: 5-9 members
	numMembers := int(rng.RandInt(5, 9))
	totalSeats := numMembers + int(rng.RandInt(0, 2))
	rolesPool := make([]string, len(committeeRoles))
	copy(rolesPool, committeeRoles)
	pyrandom.Shuffle(rng, rolesPool)
	members := make([]CommitteeMember, 0, numMembers)
	for i := 0; i < numMembers; i++ {
		fn := choice(rng, firstNames)
		ln := choice(rng, lastNames)
		members = append(members, CommitteeMember{
			Name:    fn + " " + ln,
			Title:   choice(rng, peopleTitles),
			Role:    rolesPool[i%len(rolesPool)],
			Present: rng.Random() > 0.15,
		})
	}
	numPresent := 0
	for _, m := range members {
		if m.Present {
			numPresent++
		}
	}
	quorum := numPresent >= (totalSeats/2 + 1)
	secretary := members[len(members)-1]
	for _, m := range members {
		if m.Role == "Secretary" {
			secretary = m
			break
		}
	}

	// Agenda items: 3-5
	frameworksSample := sample(rng, complianceFrameworks, 3)
	exhibitLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	numAgenda := int(rng.RandInt(3, 5))
	agendaTitlesSample := sample(rng, agendaItemTitles, numAgenda)
	items := make([]AgendaItem, 0, numAgenda)
	for i, itemTitleRaw := range agendaTitlesSample {
		itemTitle := pyFormat(itemTitleRaw, map[string]string{
			"org": org, "industry": industry,
			"frameworks": frameworksSample[i%len(frameworksSample)],
			"q":          itoaCache(q), "year": itoaCache(year), "n": itoaCache(n),
			"project_name": projectName, "regions": itoaCache(regions),
		})

		presenter := choice(rng, members)
		exhibit := exhibitLetters[i%len(exhibitLetters)]
		dueMonthNum := ((month-1+int(rng.RandInt(1, 3)))%12 + 1)
		dueYear := year
		if dueMonthNum < month {
			dueYear = year + 1
		}
		dueDate := fmt.Sprintf("%d-%02d-28", dueYear, dueMonthNum)

		discussion := choiceFormatReroll(rng, agendaDiscussionTemplate, map[string]string{
			"chair": members[0].Name, "presenter": presenter.Name, "item_title": itemTitle,
			"org": org, "industry": industry, "n": itoaCache(n), "n2": itoaCache(n2),
			"regions": itoaCache(regions), "eng_code": engCode, "q": itoaCache(q),
			"pct": itoaCache(pct), "year": itoaCache(year), "exhibit": exhibit,
			"date": dateStr, "due_date": dueDate, "project_name": projectName,
		})

		// ~60% of items have a formal motion
		var motion *Motion
		if rng.Random() < 0.6 {
			presentMembers := make([]CommitteeMember, 0, len(members))
			for _, m := range members {
				if m.Present {
					presentMembers = append(presentMembers, m)
				}
			}
			if len(presentMembers) >= 2 {
				yea := int(rng.RandInt(int64(len(presentMembers)/2+1), int64(len(presentMembers))))
				nay := int(rng.RandInt(0, int64(len(presentMembers)-yea)))
				abstain := len(presentMembers) - yea - nay
				mover := choice(rng, presentMembers)
				others := make([]CommitteeMember, 0, len(presentMembers))
				for _, m := range presentMembers {
					if m != mover {
						others = append(others, m)
					}
				}
				seconder := mover
				if len(others) > 0 {
					seconder = choice(rng, others)
				}
				resolutionText := choiceFormatReroll(rng, resolutionTemplates, map[string]string{
					"committee": committee, "item_title": itemTitle, "org": org,
					"date": dateStr, "due_date": dueDate, "exhibit": exhibit,
					"eng_code": engCode, "year": itoaCache(year), "n": itoaCache(n),
					"industry": industry, "frameworks": frameworksSample[0],
					"regions": itoaCache(regions),
				})
				motion = &Motion{
					Verb: choice(rng, motionVerbs), Text: resolutionText,
					MovedBy: mover.Name, SecondedBy: seconder.Name,
					Yea: yea, Nay: nay, Abstain: abstain,
					Carried: yea > len(presentMembers)/2,
				}
			}
		}

		items = append(items, AgendaItem{
			Number: i + 1, Title: itemTitle, Discussion: discussion, Motion: motion, Exhibit: exhibit,
		})
	}

	// Action items: 3-6
	// Pre-resolve agenda titles so nested placeholders (e.g. {frameworks})
	// don't leak, matching the Python source's separate resolution pass.
	resolvedAgendaTitles := make([]string, 0, len(agendaTitlesSample))
	for _, raw := range agendaTitlesSample {
		resolvedAgendaTitles = append(resolvedAgendaTitles, pyFormat(raw, map[string]string{
			"org": org, "industry": industry, "frameworks": frameworksSample[0],
			"q": itoaCache(q), "year": itoaCache(year), "n": itoaCache(n),
			"project_name": projectName, "regions": itoaCache(regions),
		}))
	}

	numAction := int(rng.RandInt(3, 6))
	actionItems := make([]MinutesActionItem, 0, numAction)
	for j := 0; j < numAction; j++ {
		fn := choice(rng, firstNames)
		ln := choice(rng, lastNames)
		owner := fn + " " + ln
		exhibit := exhibitLetters[j%len(exhibitLetters)]
		dueMonthNum := ((month-1+int(rng.RandInt(1, 2)))%12 + 1)
		dueYear := year
		if dueMonthNum < month {
			dueYear = year + 1
		}
		dueDate := fmt.Sprintf("%d-%02d-28", dueYear, dueMonthNum)
		// NOTE: order matters here. Python's
		//   rng.choice(_ACTION_ITEM_TEMPLATES).format(item_title=rng.choice(resolved_agenda_titles), ...)
		// picks the action template FIRST, then evaluates the keyword
		// arguments (which includes a SECOND rng.choice call for
		// item_title) before formatting — so the template choice must
		// happen before the item_title choice, not as part of a single
		// combined-helper call (which would evaluate the map literal,
		// and thus the item_title choice, before entering the helper).
		actionTmpl := choice(rng, actionItemTemplates)
		actionKw := map[string]string{
			"item_title": choice(rng, resolvedAgendaTitles), "exhibit": exhibit,
			"frameworks": frameworksSample[j%len(frameworksSample)], "org": org,
			"n": itoaCache(n), "q": itoaCache(q), "due_date": dueDate, "industry": industry,
			"year": itoaCache(year), "regions": itoaCache(regions), "project_name": projectName,
		}
		actionDesc, ok := pyFormatStrict(actionTmpl, actionKw)
		if !ok {
			actionDesc = choice(rng, actionItemTemplates)
		}
		actionItems = append(actionItems, MinutesActionItem{
			Number: j + 1, Description: actionDesc, Owner: owner, DueDate: dueDate,
		})
	}

	nextMonthNum := (month % 12) + 1
	nextYear := year
	if nextMonthNum <= month {
		nextYear = year + 1
	}
	nextMeeting := fmt.Sprintf("%d-%02d-%02d", nextYear, nextMonthNum, rng.RandInt(5, 25))

	recordID := md5Hex(fmt.Sprintf("minutes_%d_%d_%d_%s", year, month, day, slug))[:8]
	bulkHex := make([]string, 200)
	for i := range bulkHex {
		bulkHex[i] = fmt.Sprintf("%016x", rng.GetRandBits64(64))
	}

	return MinutesContent{
		Title: title, Org: org, Industry: industry, RecordID: recordID,
		Committee: committee, Location: location, CallToOrder: callToOrder,
		AdjournTime: adjournTime, MeetingRef: meetingRef, EngCode: engCode,
		DateStr: dateStr, Members: members, Quorum: quorum, TotalSeats: totalSeats,
		NumPresent: numPresent, Secretary: secretary, Items: items, ActionItems: actionItems,
		NextMeeting: nextMeeting, BulkHexJS: bulkHex[:100], BulkHexCSS: bulkHex[100:],
	}
}
