package legacyjob

import "testing"

// A year outside every language's service span (archives-1984 era pages
// requested by crawlers) must not divide by zero.
func TestIdentityOutOfSpanYear(t *testing.T) {
	for _, y := range []int{1900, 1984, 2100} {
		if id := Identity(y, 12, 13, "q3-earnings-review"); id.LangKey == "" {
			t.Fatalf("year %d: empty language", y)
		}
	}
}
