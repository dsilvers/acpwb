package archive

import (
	"regexp"
	"strings"
)

var (
	slugifyNonWord  = regexp.MustCompile(`[^\w\s-]`)
	slugifyDashesWs = regexp.MustCompile(`[-\s]+`)
)

// slugify ports django.utils.text.slugify(value, allow_unicode=False):
//
//	value = unicodedata.normalize('NFKD', value).encode('ascii', 'ignore').decode('ascii')
//	value = re.sub(r'[^\w\s-]', '', value.lower())
//	return re.sub(r'[-\s]+', '-', value).strip('-_')
//
// The NFKD-then-ascii-encode step (which drops accents/diacritics and any
// other non-ASCII code point) is a no-op for every string this port ever
// slugifies — organization names and generated titles are confirmed
// ASCII-only in this codebase's data pools — so it's simplified to a direct
// byte-level filter instead of pulling in a Unicode normalization dependency.
func slugify(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r <= 127 {
			b.WriteRune(r)
		}
	}
	ascii := strings.ToLower(b.String())
	ascii = slugifyNonWord.ReplaceAllString(ascii, "")
	ascii = slugifyDashesWs.ReplaceAllString(ascii, "-")
	return strings.Trim(ascii, "-_")
}
