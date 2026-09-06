package archive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixture struct {
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	Slug       string `json:"slug"`
	VariantInt int    `json:"variant_int"`
	HTML       string `json:"html"`
}

// TestAgainstDjangoFixtures compares this package's Go port of the
// main-domain 'default' archive page variant against fixtures dumped
// straight from the real Django/Python pyrender path (see
// apps/honeypot/management/commands/dump_archive_default_fixtures.py in the
// Django project, and acpwb_go/archive/testdata/*.json for the dumps).
func TestAgainstDjangoFixtures(t *testing.T) {
	// testdata/ also holds compliance_*.json and minutes_*.json fixtures
	// (see fixtures_variants_test.go) for the other two archive variants —
	// only match this variant's own dumps here.
	files, err := filepath.Glob("testdata/default_*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no fixture files found under testdata/ — run dump_archive_default_fixtures and copy them in")
	}

	passed := 0
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			var fx fixture
			if err := json.Unmarshal(raw, &fx); err != nil {
				t.Fatal(err)
			}

			ctx := BuildContext(fx.Year, fx.Month, fx.Day, fx.Slug)
			got := RenderArchiveDefault(&ctx)

			if got != fx.HTML {
				reportDiff(t, fx.HTML, got)
				return
			}
			passed++
		})
	}
}

// reportDiff finds the first byte at which want/got diverge and prints a
// small window of context around it, since the full HTML strings are large
// (100KB+) and a raw diff isn't useful in test output.
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
	lo := i - 80
	if lo < 0 {
		lo = 0
	}
	whi := i + 80
	if whi > len(want) {
		whi = len(want)
	}
	ghi := i + 80
	if ghi > len(got) {
		ghi = len(got)
	}
	t.Errorf(
		"mismatch at byte %d (want len=%d, got len=%d)\nWANT: ...%s...\nGOT:  ...%s...",
		i, len(want), len(got), want[lo:whi], got[lo:ghi],
	)
}
