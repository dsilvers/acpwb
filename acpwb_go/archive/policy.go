package archive

import (
	"fmt"
	"strings"

	"acpwb_go/data"
)

var agencyKeysOrder = data.Strings("AGENCY_KEYS_ORDER")

// PolicyStub carries exactly the fields
// pyrender/archive_main.py:_sidebar_related_policy_html reads from a stub
// dict built by policy_generator.py:_generate_doc_stub.
type PolicyStub struct {
	URL           string
	Title         string
	FilingDate    string
	DocumentType  string
	AgencyAcronym string
}

// GetCrossPolicyStubs ports policy_generator.py:get_cross_policy_stubs.
// Returns nil for the ~70% of inputs where the real function returns Python
// None (the sidebar section is then omitted, matching
// _sidebar_related_policy_html's `if not c['related_policy']`).
func GetCrossPolicyStubs(year, month, day int, slug string) []PolicyStub {
	rng := rngB(fmt.Sprintf("crosslink_policy_%d_%02d_%02d_%s", year, month, day, slug))
	if rng.Random() >= 0.30 {
		return nil
	}
	count := int(rng.RandInt(2, 4))
	stubs := make([]PolicyStub, 0, count)
	for i := 0; i < count; i++ {
		py := int(rng.RandInt(1993, 2025))
		pm := int(rng.RandInt(1, 12))
		pd := int(rng.RandInt(1, 28))
		pagency := choice(rng, agencyKeysOrder)
		pslug := choice(rng, policySlugs)
		useSubdomainURL := rng.Random() < 0.5
		stubs = append(stubs, generateDocStub(py, pm, pd, pagency, pslug, useSubdomainURL))
	}
	return stubs
}

// generateDocStub ports policy_generator.py:_generate_doc_stub. Its
// signatory/docket-number RNG "replay" calls are intentionally not
// reproduced: they're drawn from a rng LOCAL to this function, strictly
// after every field this port renders is already computed, and nothing
// downstream reads that rng again — so they cannot affect this function's
// observable output and are safe to skip entirely.
func generateDocStub(year, month, day int, agency, slug string, useSubdomainURL bool) PolicyStub {
	seed := fmt.Sprintf("acpwb_policy_%d_%02d_%02d_%s_%s", year, month, day, agency, slug)
	rng := rngB(seed)

	docType := choice(rng, documentTypes)
	docTypeSlug, docTypeLabel := docType[0], docType[1]

	prefixPool, ok := stubTitlePrefixes[docTypeSlug]
	if !ok || len(prefixPool) == 0 {
		prefixPool = []string{"Filing on"}
	}
	titlePrefix := choice(rng, prefixPool)

	words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
	for i, w := range words {
		words[i] = pyCapitalize(w)
	}
	topicTitle := strings.Join(words, " ")
	title := titlePrefix + " " + topicTitle

	filingDate := strftimeBMonthDayYear(year, month, day)

	var url string
	if useSubdomainURL {
		url = fmt.Sprintf("https://policy-%s.acpwb.com/%d/%02d/%02d/%s/", agency, year, month, day, slug)
	} else {
		url = fmt.Sprintf("/public-policy/%d/%02d/%02d/%s/%s/", year, month, day, agency, slug)
	}

	return PolicyStub{
		URL: url, Title: title, FilingDate: filingDate,
		DocumentType: docTypeLabel, AgencyAcronym: strings.ToUpper(agency),
	}
}

var monthFullNames = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// strftimeBMonthDayYear reproduces datetime.date(y,m,d).strftime('%B %-d, %Y')
// (full month name, day with no leading zero, 4-digit year), falling back to
// "YYYY-MM-DD" the way the Python does on an invalid calendar date (e.g.
// Feb 30 — datetime.date raises ValueError there). day is always in [1, 28]
// at this port's only call site, so the fallback path is unreachable here,
// but is kept for parity.
func strftimeBMonthDayYear(year, month, day int) string {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}
	return fmt.Sprintf("%s %d, %d", monthFullNames[month-1], day, year)
}
