package legacyjob

import "strings"

// CodeEmail is a programmer's address on a source-code page, built exactly
// like every other generated ACPWB address (apps/people/generators.py,
// archive engagement teams): lowercase "first.last@acpwb.com".
//
// The same address can appear on many pages, so the handler registers every
// (address, page) pair it serves via the acpwb:code_email_queue Redis list
// (drained into honeypot.SourceCodeEmail); an inbound email to the address
// can then be matched back to the page(s) it was harvested from.
func CodeEmail(first, last string) string {
	clean := func(s string) string {
		var b strings.Builder
		for _, r := range strings.ToLower(s) {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			} else if r == ' ' && b.Len() > 0 {
				b.WriteByte('.')
			}
		}
		return b.String()
	}
	f, l := clean(first), clean(last)
	if f == "" {
		return l + "@acpwb.com"
	}
	return f + "." + l + "@acpwb.com"
}
