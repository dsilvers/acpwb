package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"acpwb_go/archive"
)

const fixedToken = "fixedtok01"

// reportDiff finds the first byte at which want/got diverge and prints a
// small window of context (mirrors archive/fixtures_test.go's helper).
func reportDiff(t *testing.T, want, got string) {
	t.Helper()
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	i := 0
	for i < n && want[i] == got[i] {
		i++
	}
	lo := i - 100
	if lo < 0 {
		lo = 0
	}
	whi := i + 100
	if whi > len(want) {
		whi = len(want)
	}
	ghi := i + 100
	if ghi > len(got) {
		ghi = len(got)
	}
	t.Errorf(
		"mismatch at byte %d (want len=%d, got len=%d)\nWANT: ...%s...\nGOT:  ...%s...",
		i, len(want), len(got), want[lo:whi], got[lo:ghi],
	)
}

func loadFixtures(t *testing.T, glob string) []map[string]any {
	files, err := filepath.Glob(glob)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no fixture files found for %s — run dump_policy_fixtures and copy them into testdata/", glob)
	}
	var out []map[string]any
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		m["__file"] = filepath.Base(f)
		out = append(out, m)
	}
	return out
}

func policyYearsDesc() []int {
	years := make([]int, 0, 2025-1992)
	for y := 2025; y >= 1993; y-- {
		years = append(years, y)
	}
	return years
}

func mainYearURLFn() PolicyYearURLFunc {
	return func(y int) string { return fmt.Sprintf("/public-policy/%d/", y) }
}
func mainMonthURLFn() PolicyMonthURLFunc {
	return func(y, m int) string { return fmt.Sprintf("/public-policy/%d/%02d/", y, m) }
}
func subYearURLFn() PolicyYearURLFunc {
	return func(y int) string { return fmt.Sprintf("/%d/", y) }
}
func subMonthURLFn() PolicyMonthURLFunc {
	return func(y, m int) string { return fmt.Sprintf("/%d/%02d/", y, m) }
}

func TestPolicyIndexFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/index/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "", RequestPath: "/public-policy/"}
			years := GetPolicyIndexYears()
			got := RenderPolicyIndex(meta, years)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_index: %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicyYearFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/year/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			year := int(fx["year"].(float64))
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "", RequestPath: fmt.Sprintf("/public-policy/%d/", year)}
			yearData := GetPolicyYearData(year)
			months := GetPolicyYearMonths(year)
			got := RenderPolicyYear(meta, year, yearData, months, policyYearsDesc(), year-1, year+1)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_year: %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicyMonthMainFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/month_main/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			year := int(fx["year"].(float64))
			month := int(fx["month"].(float64))
			prevMonth, prevYear := month-1, year
			if month == 1 {
				prevMonth, prevYear = 12, year-1
			}
			nextMonth, nextYear := month+1, year
			if month == 12 {
				nextMonth, nextYear = 1, year+1
			}
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "", RequestPath: fmt.Sprintf("/public-policy/%d/%02d/", year, month)}
			entries := GetPolicyMonthEntries(year, month)
			p := MonthPageParams{
				Year: year, Month: month, Entries: entries, PolicyYears: policyYearsDesc(),
				PolicyIndexURL: "/public-policy/", YearURL: mainYearURLFn()(year),
				PrevMonthURL: mainMonthURLFn()(prevYear, prevMonth), NextMonthURL: mainMonthURLFn()(nextYear, nextMonth),
				PolicyYearURLFn: mainYearURLFn(),
			}
			got := RenderPolicyMonth(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_month (main): %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicyMonthSubdomainFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/month_sub/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			year := int(fx["year"].(float64))
			month := int(fx["month"].(float64))
			agency := fx["agency"].(string)
			prevMonth, prevYear := month-1, year
			if month == 1 {
				prevMonth, prevYear = 12, year-1
			}
			nextMonth, nextYear := month+1, year
			if month == 12 {
				nextMonth, nextYear = 1, year+1
			}
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "https://acpwb.com", RequestPath: fmt.Sprintf("/%d/%02d/", year, month)}
			urlFn := func(y, m, d int, ag, sl string) string { return fmt.Sprintf("/%d/%02d/%02d/%s/", y, m, d, sl) }
			entries := GetPolicyAgencyMonthEntries(agency, year, month, urlFn)
			p := MonthPageParams{
				Year: year, Month: month, Entries: entries, PolicyYears: policyYearsDesc(),
				PolicyIndexURL: "/", YearURL: subYearURLFn()(year),
				PrevMonthURL: subMonthURLFn()(prevYear, prevMonth), NextMonthURL: subMonthURLFn()(nextYear, nextMonth),
				PolicyYearURLFn: subYearURLFn(),
			}
			got := RenderPolicyMonth(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_month (subdomain): %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicyDetailMainFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/detail_main/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			year := int(fx["year"].(float64))
			month := int(fx["month"].(float64))
			day := int(fx["day"].(float64))
			agency := fx["agency"].(string)
			slug := fx["slug"].(string)

			doc := GeneratePolicyDocument(year, month, day, agency, slug)
			related := GenerateRelatedLinks(year, month, day, agency, slug, nil)
			relatedArchive := GetCrossArchiveStubs(year, month, day, agency, slug)
			relatedPres := archive.GeneratePresentationsForContext(
				fmt.Sprintf("policy_pres_%d_%d_%d_%s_%s", year, month, day, agency, truncate(slug, 32)), 4)

			meta := PageMeta{
				HoneypotToken: fixedToken, SiteRoot: "",
				RequestPath: fmt.Sprintf("/public-policy/%d/%02d/%02d/%s/%s/", year, month, day, agency, slug),
				NowYear:     2026,
			}
			p := DetailParams{
				Doc: doc, Related: &related, RelatedArchive: relatedArchive, RelatedPresentations: relatedPres,
				PolicyYears: policyYearsDesc(), PolicyYearURL: mainYearURLFn(), PolicyMonthURL: mainMonthURLFn(),
			}
			got := RenderPolicyDetail(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_detail (main): %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicyDetailSubdomainFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/detail_sub/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			year := int(fx["year"].(float64))
			month := int(fx["month"].(float64))
			day := int(fx["day"].(float64))
			agency := fx["agency"].(string)
			slug := fx["slug"].(string)

			urlFn := URLFunc(func(y, m, d int, ag, sl string) string {
				if ag == agency {
					return fmt.Sprintf("/%d/%02d/%02d/%s/", y, m, d, sl)
				}
				return fmt.Sprintf("https://policy-%s.acpwb.com/%d/%02d/%02d/%s/", ag, y, m, d, sl)
			})

			doc := GeneratePolicyDocument(year, month, day, agency, slug)
			doc.URL = urlFn(year, month, day, agency, slug)
			related := GenerateRelatedLinks(year, month, day, agency, slug, urlFn)
			relatedArchive := GetCrossArchiveStubs(year, month, day, agency, slug)

			meta := PageMeta{
				HoneypotToken: fixedToken, SiteRoot: "https://acpwb.com",
				RequestPath: fmt.Sprintf("/%d/%02d/%02d/%s/", year, month, day, slug),
				NowYear:     2026,
			}
			p := DetailParams{
				Doc: doc, Related: &related, RelatedArchive: relatedArchive, RelatedPresentations: nil,
				PolicyYears: policyYearsDesc(), PolicyYearURL: subYearURLFn(), PolicyMonthURL: subMonthURLFn(),
			}
			got := RenderPolicyDetail(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_detail (subdomain): %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicySubdomainIndexFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/sub_index/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			agency := fx["agency"].(string)
			full, domain, ok := AgencyData(agency)
			if !ok {
				full, domain = "Unknown Agency", "regulatory policy"
			}
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "https://acpwb.com", RequestPath: "/"}
			p := SubdomainIndexParams{
				Agency: agency, AgencyFull: full, PolicyDomain: domain,
				Years:         GetPolicyAgencyYears(agency),
				OGTitle:       fmt.Sprintf("%s Policy Filings — ACPWB", upperAscii(agency)),
				OGDescription: fmt.Sprintf("ACPWB regulatory filings, comment letters, and testimony submitted to the %s.", full),
				PolicyYearURL: subYearURLFn(), PolicyMonthURL: subMonthURLFn(),
			}
			got := RenderPolicySubdomainIndex(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_subdomain_index: %d/%d fixtures matched", passed, len(fixtures))
}

func TestPolicySubdomainYearFixtures(t *testing.T) {
	fixtures := loadFixtures(t, "testdata/sub_year/*.json")
	passed := 0
	for _, fx := range fixtures {
		t.Run(fx["__file"].(string), func(t *testing.T) {
			agency := fx["agency"].(string)
			year := int(fx["year"].(float64))
			full, domain, ok := AgencyData(agency)
			if !ok {
				full, domain = "Unknown Agency", "regulatory policy"
			}
			meta := PageMeta{HoneypotToken: fixedToken, SiteRoot: "https://acpwb.com", RequestPath: fmt.Sprintf("/%d/", year)}
			p := SubdomainYearParams{
				Agency: agency, AgencyFull: full, PolicyDomain: domain, Year: year,
				YearDetail: GetPolicyAgencyYearDetail(agency, year),
				AllYears:   GetPolicyAgencyYears(agency),
				PrevYear:   year - 1, NextYear: year + 1,
				OGTitle:        fmt.Sprintf("%d %s Policy Filings — ACPWB", year, upperAscii(agency)),
				PolicyIndexURL: "/",
				PolicyYearURL:  subYearURLFn(), PolicyMonthURL: subMonthURLFn(),
			}
			got := RenderPolicySubdomainYear(meta, p)
			want := fx["html"].(string)
			if got != want {
				reportDiff(t, want, got)
				return
			}
			passed++
		})
	}
	t.Logf("policy_subdomain_year: %d/%d fixtures matched", passed, len(fixtures))
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

func upperAscii(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}
