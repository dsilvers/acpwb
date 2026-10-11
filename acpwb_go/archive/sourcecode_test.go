package archive

import (
	"os"
	"strings"
	"testing"

	"acpwb_go/legacyjob"
)

const exampleSlug = "human-factors-and-usability-engineering-report-9293"

// TestSourceCodeDump writes the example page's listing to $LJ_DUMP when set
// (developer aid), and always checks the printed output matches the tables.
func TestSourceCodeDump(t *testing.T) {
	in := SourceCodeInput(2022, 4, 6, exampleSlug, true)
	l := legacyjob.Build(&in)
	if dir := os.Getenv("LJ_DUMP"); dir != "" {
		var b strings.Builder
		for _, s := range l.Sections {
			b.WriteString("=== " + s.Tab + " ===\n" + s.Text + "\n")
		}
		_ = os.WriteFile(dir+"/listing.txt", []byte(b.String()), 0o644)
		_ = os.WriteFile(dir+"/page.html", []byte(legacyjob.RenderHTML(l)), 0o644)
	}
	var printed string
	for _, s := range l.Sections {
		if s.ID == "sysprint" {
			printed = s.Text
		}
	}
	c := GenerateArchiveContent(2022, 4, 6, exampleSlug)
	for _, r := range c.PercentileTable {
		for _, v := range []string{r.P10, r.P50, r.P95} {
			if !strings.Contains(printed, strings.NewReplacer("$", "", ",", "").Replace(v)) {
				t.Errorf("SYSPRINT missing %s", v)
			}
		}
	}
}
