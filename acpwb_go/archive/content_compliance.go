package archive

import (
	"fmt"
	"strings"

	"acpwb_go/data"
	"acpwb_go/pyrandom"
)

// Compliance content pools (apps/honeypot/archive_data_compliance.py via
// export_render_data2.py).
var (
	auditRefPrefixes          = data.Strings("AUDIT_REF_PREFIXES")
	complianceFrameworks      = data.Strings("COMPLIANCE_FRAMEWORKS")
	complianceFindingTypes    = data.Strings("COMPLIANCE_FINDING_TYPES")
	complianceRiskLevels      = data.Strings("COMPLIANCE_RISK_LEVELS")
	complianceStatuses        = data.Strings("COMPLIANCE_STATUSES")
	complianceScopeTemplates  = data.Strings("COMPLIANCE_SCOPE_TEMPLATES")
	complianceMethodTemplates = data.Strings("COMPLIANCE_METHODOLOGY_TEMPLATES")
	correctiveActionTemplates = data.Strings("CORRECTIVE_ACTION_TEMPLATES")
	mgmtResponseTemplates     = data.Strings("MGMT_RESPONSE_TEMPLATES")
	complianceProjectNames    = data.Strings("PROJECT_NAMES")
	complianceDocVersions     = data.Strings("COMPLIANCE_DOC_VERSIONS")
	complianceTitlePrefixes   = data.Strings("COMPLIANCE_TITLE_PREFIXES")
)

// ComplianceFinding is one row of the compliance findings table/cards.
type ComplianceFinding struct {
	ID               string
	Risk             string
	Status           string
	Owner            string
	Description      string
	CorrectiveAction string
	MgmtResponse     string
	DueDate          string
}

// ComplianceDistEntry is one row of the distribution list.
type ComplianceDistEntry struct {
	Name  string
	Title string
	Email string
}

// ComplianceContent is the Go equivalent of _generate_compliance_content()'s
// return dict (apps/honeypot/views.py:313).
type ComplianceContent struct {
	Title           string
	Org             string
	Industry        string
	RecordID        string
	AuditRef        string
	DocVersion      string
	DateStr         string
	Assessor        string
	AssessorTitle   string
	AssessorEmail   string
	FrameworksCited []string
	ScopePara       string
	MethodPara      string
	Findings        []ComplianceFinding
	DistList        []ComplianceDistEntry
	N               int
	Regions         int
	Pct             int
	BulkHexJS       []string
	BulkHexCSS      []string
}

// GenerateComplianceContent ports apps/honeypot/views.py:_generate_compliance_content.
func GenerateComplianceContent(year, month, day int, slug string) ComplianceContent {
	rng := rngA(fmt.Sprintf("compliance_%d%d%d%s", year, month, day, slug))

	org := choice(rng, archiveOrgs)
	industry := choice(rng, archiveIndustries)
	dateStr := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	n := int(rng.RandInt(18, 340))
	n2 := int(rng.RandInt(10, 80))
	regions := int(rng.RandInt(2, 24))
	endYear := min(year+int(rng.RandInt(0, 2)), 2024)
	endDate := fmt.Sprintf("%d-%02d-28", endYear, min(month+2, 12))
	q := int(rng.RandInt(1, 4))
	pct := int(rng.RandInt(4, 22))
	docVersion := choice(rng, complianceDocVersions)
	projectName := choice(rng, complianceProjectNames)

	auditPrefix := choice(rng, auditRefPrefixes)
	auditRef := fmt.Sprintf("%s-%d-Q%d-%04d", auditPrefix, year, q, rng.RandInt(1000, 9999))

	// Title from slug
	tail := slug
	if slug != "" {
		parts := strings.Split(slug, "/")
		tail = parts[len(parts)-1]
	} else {
		tail = fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}
	cleanTail := trailingNumericID.ReplaceAllString(tail, "")
	title := choice(rng, complianceTitlePrefixes) + " — " + pyTitle(strings.ReplaceAll(cleanTail, "-", " "))

	// Assessor
	fname := choice(rng, firstNames)
	lname := choice(rng, lastNames)
	assessor := fname + " " + lname
	assessorTitle := choice(rng, consultantTitles)
	assessorEmail := strings.ToLower(fname) + "." + strings.ToLower(lname) + "@acpwb.com"

	// Frameworks cited
	frameworksCited := sample(rng, complianceFrameworks, int(rng.RandInt(2, 4)))
	frameworksStr := strings.Join(frameworksCited, "; ")

	kwBase := func() map[string]string {
		return map[string]string{
			"org": org, "industry": industry, "regions": itoaCache(regions),
			"year": itoaCache(year), "endyear": itoaCache(endYear),
			"date": dateStr, "enddate": endDate, "n": itoaCache(n), "n2": itoaCache(n2),
			"q": itoaCache(q), "frameworks": frameworksStr, "doc_version": docVersion,
		}
	}
	scopePara := choiceFormatReroll(rng, complianceScopeTemplates, kwBase())

	methodPara := choiceFormatReroll(rng, complianceMethodTemplates, map[string]string{
		"org": org, "industry": industry, "n": itoaCache(n), "pct": itoaCache(pct),
		"regions": itoaCache(regions), "year": itoaCache(year), "doc_version": docVersion,
		"frameworks": frameworksStr,
	})

	// Findings: 4-7 rows
	numFindings := int(rng.RandInt(4, 7))
	riskWeights := []float64{1, 3, 5, 3}
	findings := make([]ComplianceFinding, 0, numFindings)
	for i := 0; i < numFindings; i++ {
		findingID := fmt.Sprintf("%s-%04d-%03d", auditPrefix, rng.RandInt(1000, 9999), i+1)
		risk := pyrandom.Choices(rng, complianceRiskLevels, riskWeights, nil, 1)[0]
		status := choice(rng, complianceStatuses)
		ownerTitle := choice(rng, peopleTitles)
		findingType := choice(rng, complianceFindingTypes)
		description := pyFormat(findingType, map[string]string{
			"org": org, "industry": industry, "n": itoaCache(n), "regions": itoaCache(regions),
			"pct": itoaCache(pct), "doc_version": docVersion, "frameworks": frameworksStr,
			"year": itoaCache(year),
		})
		corrective := choiceFormatReroll(rng, correctiveActionTemplates, map[string]string{
			"org": org, "n": itoaCache(n), "regions": itoaCache(regions), "industry": industry,
		})
		mgmtResp := choiceFormatReroll(rng, mgmtResponseTemplates, map[string]string{
			"org": org, "pct": itoaCache(pct), "regions": itoaCache(regions), "date": dateStr,
			"project_name": projectName, "q": itoaCache(q), "endyear": itoaCache(endYear),
			"n": itoaCache(n), "n2": itoaCache(n2),
		})

		dueDeltaMonths := int(rng.RandInt(1, 4))
		dueMonth := ((month-1+dueDeltaMonths)%12 + 1)
		dueYear := year
		if dueMonth < month {
			dueYear = year + 1
		}
		dueDate := fmt.Sprintf("%d-%02d-28", dueYear, dueMonth)

		findings = append(findings, ComplianceFinding{
			ID: findingID, Risk: risk, Status: status, Owner: ownerTitle,
			Description: description, CorrectiveAction: corrective, MgmtResponse: mgmtResp,
			DueDate: dueDate,
		})
	}

	// Distribution list
	numDist := int(rng.RandInt(3, 5))
	distList := make([]ComplianceDistEntry, 0, numDist)
	for i := 0; i < numDist; i++ {
		fn := choice(rng, firstNames)
		ln := choice(rng, lastNames)
		distList = append(distList, ComplianceDistEntry{
			Name:  fn + " " + ln,
			Title: choice(rng, peopleTitles),
			Email: strings.ToLower(fn) + "." + strings.ToLower(ln) + "@acpwb.com",
		})
	}

	recordID := md5Hex(fmt.Sprintf("compliance_%d_%d_%d_%s", year, month, day, slug))[:8]
	bulkHex := make([]string, 200)
	for i := range bulkHex {
		bulkHex[i] = fmt.Sprintf("%016x", rng.GetRandBits64(64))
	}

	return ComplianceContent{
		Title: title, Org: org, Industry: industry, RecordID: recordID, AuditRef: auditRef,
		DocVersion: docVersion, DateStr: dateStr, Assessor: assessor, AssessorTitle: assessorTitle,
		AssessorEmail: assessorEmail, FrameworksCited: frameworksCited, ScopePara: scopePara,
		MethodPara: methodPara, Findings: findings, DistList: distList,
		N: n, Regions: regions, Pct: pct,
		BulkHexJS: bulkHex[:100], BulkHexCSS: bulkHex[100:],
	}
}
