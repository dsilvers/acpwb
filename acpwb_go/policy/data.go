package policy

import (
	"encoding/json"
	"fmt"
	"strconv"

	"acpwb_go/data"
)

// Data pools loaded from acpwb_go/data (see policy_data.py / policy_generator.py
// for the Python source of truth these mirror).
var (
	agencies            = data.TupleMap("AGENCIES")
	agencyKeysOrder     = data.Strings("AGENCY_KEYS_ORDER")
	policySlugs         = data.Strings("POLICY_SLUGS")
	documentTypes       = data.TuplePairs("DOCUMENT_TYPES")
	signatoryTitles     = data.Strings("SIGNATORY_TITLES")
	credentials         = data.Strings("CREDENTIALS")
	legislation         = data.Strings("LEGISLATION")
	summaryTemplates    = data.StringSliceMap("SUMMARY_TEMPLATES")
	sectionHeadings     = loadSectionHeadings()
	optionalSectionPool = data.Strings("OPTIONAL_SECTION_POOL")
	paragraphTemplates  = data.Strings("PARAGRAPH_TEMPLATES")
	recommendationTmpls = data.Strings("RECOMMENDATION_TEMPLATES")
	positions           = data.TuplePairs("POSITIONS")
	monthsLong          = data.Strings("MONTHS_LONG")
	policyFootnoteTmpls = data.Strings("POLICY_FOOTNOTE_TEMPLATES")
	stubTitlePrefixes   = data.StringSliceMap("STUB_TITLE_PREFIXES")
	featuredSeeds       = loadFeaturedSeeds()
	ceoNames            = loadCEONames()
	yearEraThemes       = data.StringSliceMap("YEAR_ERA_THEMES")
	ceoMessageTemplates = data.Strings("CEO_MESSAGE_TEMPLATES")
	yearAnnualLetters   = loadYearAnnualLetters()
	expertTypes         = data.Strings("EXPERT_TYPES")
	industrySectors     = data.Strings("INDUSTRY_SECTORS")
	timeframes          = data.Strings("TIMEFRAMES")
	comparisonGroups    = data.Strings("COMPARISON_GROUPS")
	findingsBrief       = data.Strings("FINDINGS_BRIEF")
	firstNames          = data.Strings("FIRST_NAMES")
	lastNames           = data.Strings("LAST_NAMES")
	archiveSlugs        = data.Strings("ARCHIVE_SLUGS")
)

func loadSectionHeadings() map[string][][]string {
	var out map[string][][]string
	data.Unmarshal("SECTION_HEADINGS", &out)
	return out
}

// featuredSeed mirrors one _FEATURED_SEEDS tuple: (year, month, day, agency, slug).
type featuredSeed struct {
	Year   int
	Month  int
	Day    int
	Agency string
	Slug   string
}

func loadFeaturedSeeds() []featuredSeed {
	var raw []struct {
		Tuple []json.RawMessage `json:"__tuple__"`
	}
	data.Unmarshal("FEATURED_SEEDS", &raw)
	out := make([]featuredSeed, len(raw))
	for i, t := range raw {
		if len(t.Tuple) != 5 {
			panic(fmt.Sprintf("FEATURED_SEEDS entry %d is not a 5-tuple", i))
		}
		var y, m, d int
		var ag, sl string
		_ = json.Unmarshal(t.Tuple[0], &y)
		_ = json.Unmarshal(t.Tuple[1], &m)
		_ = json.Unmarshal(t.Tuple[2], &d)
		_ = json.Unmarshal(t.Tuple[3], &ag)
		_ = json.Unmarshal(t.Tuple[4], &sl)
		out[i] = featuredSeed{y, m, d, ag, sl}
	}
	return out
}

// ceoName mirrors one _CEO_NAMES tuple: (start_year, end_year, name, title).
type ceoName struct {
	Start int
	End   int
	Name  string
	Title string
}

func loadCEONames() []ceoName {
	var raw []struct {
		Tuple []json.RawMessage `json:"__tuple__"`
	}
	data.Unmarshal("CEO_NAMES", &raw)
	out := make([]ceoName, len(raw))
	for i, t := range raw {
		if len(t.Tuple) != 4 {
			panic(fmt.Sprintf("CEO_NAMES entry %d is not a 4-tuple", i))
		}
		var start, end int
		var name, title string
		_ = json.Unmarshal(t.Tuple[0], &start)
		_ = json.Unmarshal(t.Tuple[1], &end)
		_ = json.Unmarshal(t.Tuple[2], &name)
		_ = json.Unmarshal(t.Tuple[3], &title)
		out[i] = ceoName{start, end, name, title}
	}
	return out
}

func loadYearAnnualLetters() map[int][]string {
	var raw map[string][]string
	data.Unmarshal("YEAR_ANNUAL_LETTERS", &raw)
	out := make(map[int][]string, len(raw))
	for k, v := range raw {
		y, err := strconv.Atoi(k)
		if err != nil {
			panic(fmt.Sprintf("YEAR_ANNUAL_LETTERS: bad year key %q: %v", k, err))
		}
		out[y] = v
	}
	return out
}

// AgencyData returns (full_name, policy_domain) for a lowercased agency
// acronym, mirroring AGENCIES.get(agency.lower()) with the
// f"{agency.upper()} Regulatory Authority" fallback used throughout
// policy_generator.py.
func AgencyData(agency string) (full string, domain string, ok bool) {
	v, ok := agencies[agency]
	if !ok {
		return "", "", false
	}
	return v[0], v[1], true
}
