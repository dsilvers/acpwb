package botclassify

import (
	"bytes"
	"os"
	"testing"
)

// botclassify.go is a byte-for-byte copy of acpwb_go's port of
// apps/core/bot_classify.py (separate Go modules can't share a package
// without a replace directive). acpwb_go's fixture test pins it to the
// Python; this pins the copy to acpwb_go's.
func TestSameAsAcpwbGo(t *testing.T) {
	theirs, err := os.ReadFile("../../acpwb_go/botclassify/botclassify.go")
	if err != nil {
		t.Skip("acpwb_go not checked out alongside:", err)
	}
	ours, err := os.ReadFile("botclassify.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ours, theirs) {
		t.Fatal("botclassify.go differs from acpwb_go/botclassify/botclassify.go: copy it over")
	}
}
