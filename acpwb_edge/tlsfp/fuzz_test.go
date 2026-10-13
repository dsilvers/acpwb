package tlsfp

import (
	"os"
	"path/filepath"
	"testing"
)

// Parse and Compute run on bytes straight off the network: they must never
// panic, whatever they're given.
func FuzzParse(f *testing.F) {
	bins, _ := filepath.Glob(filepath.Join("testdata", "hellos", "*.bin"))
	for _, b := range bins {
		raw, err := os.ReadFile(b)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(raw)
		f.Add(raw[:len(raw)/2])
	}
	f.Add([]byte{0x16, 0x03, 0x01, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00})
	f.Fuzz(func(t *testing.T, raw []byte) {
		h, err := Parse(raw)
		if err == nil {
			Compute(h)
		}
	})
}
