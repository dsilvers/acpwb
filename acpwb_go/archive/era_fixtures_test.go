package archive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type eraFixture struct {
	Variant    string `json:"variant"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	Slug       string `json:"slug"`
	VariantInt int    `json:"variant_int"`
	HTML       string `json:"html"`
}

// TestEraDefaultAgainstDjangoFixtures compares this package's Go port of the
// archives-YYYY.acpwb.com subdomain ("era") 'default' variant against
// fixtures dumped from the real Django/Python pyrender path (see
// apps/honeypot/management/commands/dump_archive_era_fixtures.py in the
// Django project, and testdata/era_default_*.json for the dumps).
func TestEraDefaultAgainstDjangoFixtures(t *testing.T) {
	runEraFixtures(t, "testdata/era_default_*.json", func(fx eraFixture) string {
		ctx := BuildEraContext(fx.Year, fx.Month, fx.Day, fx.Slug)
		return RenderArchiveDefaultEra(&ctx)
	})
}

// TestEraComplianceAgainstDjangoFixtures is the 'compliance' era variant's
// counterpart to TestEraDefaultAgainstDjangoFixtures.
func TestEraComplianceAgainstDjangoFixtures(t *testing.T) {
	runEraFixtures(t, "testdata/era_compliance_*.json", func(fx eraFixture) string {
		ctx := BuildEraComplianceContext(fx.Year, fx.Month, fx.Day, fx.Slug)
		return RenderComplianceDefaultEra(&ctx)
	})
}

// TestEraMinutesAgainstDjangoFixtures is the 'minutes' era variant's
// counterpart to TestEraDefaultAgainstDjangoFixtures.
func TestEraMinutesAgainstDjangoFixtures(t *testing.T) {
	runEraFixtures(t, "testdata/era_minutes_*.json", func(fx eraFixture) string {
		ctx := BuildEraMinutesContext(fx.Year, fx.Month, fx.Day, fx.Slug)
		return RenderMinutesDefaultEra(&ctx)
	})
}

func runEraFixtures(t *testing.T, glob string, render func(eraFixture) string) {
	t.Helper()
	files, err := filepath.Glob(glob)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no fixture files found matching %s — run dump_archive_era_fixtures and copy them in", glob)
	}

	passed := 0
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			var fx eraFixture
			if err := json.Unmarshal(raw, &fx); err != nil {
				t.Fatal(err)
			}

			got := render(fx)
			if got != fx.HTML {
				reportDiff(t, fx.HTML, got)
				return
			}
			passed++
		})
	}
}
