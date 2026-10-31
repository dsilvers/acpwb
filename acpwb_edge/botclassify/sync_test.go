package botclassify

import (
	"bytes"
	"os"
	"testing"
)

// botclassify.go and vendor_ranges.tsv are byte-for-byte copies of acpwb_go's port of
// apps/core/bot_classify.py (separate Go modules can't share a package
// without a replace directive). acpwb_go's fixture test pins it to the
// Python; this pins the copy to acpwb_go's.
func TestSameAsAcpwbGo(t *testing.T) {
	for _, name := range []string{"botclassify.go", "vendor_ranges.tsv"} {
		theirs, err := os.ReadFile("../../acpwb_go/botclassify/" + name)
		if err != nil {
			t.Skip("acpwb_go not checked out alongside:", err)
		}
		ours, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(ours, theirs) {
			t.Fatal(name + " differs from acpwb_go/botclassify/" + name + ": copy it over")
		}
	}
}
