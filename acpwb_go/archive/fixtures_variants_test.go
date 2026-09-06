package archive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type variantFixture struct {
	Variant    string `json:"variant"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	Slug       string `json:"slug"`
	VariantInt int    `json:"variant_int"`
	HTML       string `json:"html"`
}

// TestComplianceAgainstDjangoFixtures compares this package's Go port of the
// main-domain 'compliance' archive page variant against fixtures dumped from
// the real Django/Python pyrender path (see
// apps/honeypot/management/commands/dump_archive_compliance_minutes_fixtures.py
// in the Django project, and testdata/compliance_*.json for the dumps).
func TestComplianceAgainstDjangoFixtures(t *testing.T) {
	runVariantFixtures(t, "testdata/compliance_*.json", func(fx variantFixture) string {
		ctx := BuildComplianceContext(fx.Year, fx.Month, fx.Day, fx.Slug)
		return RenderComplianceDefault(&ctx)
	})
}

// TestMinutesAgainstDjangoFixtures is the 'minutes' variant's counterpart to
// TestComplianceAgainstDjangoFixtures.
func TestMinutesAgainstDjangoFixtures(t *testing.T) {
	runVariantFixtures(t, "testdata/minutes_*.json", func(fx variantFixture) string {
		ctx := BuildMinutesContext(fx.Year, fx.Month, fx.Day, fx.Slug)
		return RenderMinutesDefault(&ctx)
	})
}

func runVariantFixtures(t *testing.T, glob string, render func(variantFixture) string) {
	t.Helper()
	files, err := filepath.Glob(glob)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no fixture files found matching %s — run dump_archive_compliance_minutes_fixtures and copy them in", glob)
	}

	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			var fx variantFixture
			if err := json.Unmarshal(raw, &fx); err != nil {
				t.Fatal(err)
			}

			got := render(fx)
			if got != fx.HTML {
				reportDiff(t, fx.HTML, got)
			}
		})
	}
}
