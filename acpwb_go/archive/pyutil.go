package archive

import (
	"strings"
	"unicode"

	"acpwb_go/pyrandom"
)

// pyFormatStrict is a minimal stand-in for Python's str.format(**kwargs),
// covering exactly the subset actually used by the archive/report/
// presentation template pools ported here: plain "{name}" placeholders (no
// format specs, no positional args, no "{{"/"}}" escapes — verified none of
// the source template pools use them). ok is false if the template
// references any name not present in kwargs — mirroring Python's KeyError —
// so callers can replicate the source's `try: ... except (KeyError,
// IndexError): use the raw template` fallback pattern. Several call sites in
// this codebase genuinely hit this path (e.g. _generate_archive_content's
// exec_bullets formatting doesn't pass phase=, but some EXEC_SUMMARY_BULLETS
// templates reference {phase} — the real Django app really does render
// those bullets as raw, unfilled template text, and this port must match
// that, not "fix" it).
func pyFormatStrict(tmpl string, kwargs map[string]string) (string, bool) {
	var b strings.Builder
	i := 0
	for i < len(tmpl) {
		c := tmpl[i]
		if c == '{' {
			end := strings.IndexByte(tmpl[i+1:], '}')
			if end < 0 {
				b.WriteString(tmpl[i:])
				break
			}
			name := tmpl[i+1 : i+1+end]
			v, ok := kwargs[name]
			if !ok {
				return tmpl, false
			}
			b.WriteString(v)
			i = i + 1 + end + 1
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String(), true
}

// pyFormat formats with fallback-to-raw-template-on-missing-key already
// applied, for the (common) call sites where that's exactly the desired
// behavior.
func pyFormat(tmpl string, kwargs map[string]string) string {
	out, ok := pyFormatStrict(tmpl, kwargs)
	if !ok {
		return tmpl
	}
	return out
}

// pyTitle reproduces Python's str.title(): each maximal run of Unicode
// letters gets its first letter uppercased and the rest lowercased; any
// non-letter (space, digit, punctuation, apostrophe...) breaks a run, so a
// letter immediately following one is titlecased again (matching Python's
// well-known "don't" -> "Don'T" behavior).
func pyTitle(s string) string {
	var b strings.Builder
	prevCased := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			if prevCased {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(unicode.ToUpper(r))
			}
			prevCased = true
		} else {
			b.WriteRune(r)
			prevCased = false
		}
	}
	return b.String()
}

// choiceFormatReroll reproduces the compliance/minutes generators' quirky
// fallback pattern, e.g.:
//
//	try:
//	    scope_para = rng.choice(_COMPLIANCE_SCOPE_TEMPLATES).format(**kw)
//	except (KeyError, IndexError):
//	    scope_para = rng.choice(_COMPLIANCE_SCOPE_TEMPLATES)
//
// Unlike the default archive variant's fallback (which just reuses the
// already-chosen raw template — see pyFormat), several of the compliance and
// minutes generators re-roll a FRESH rng.choice() call on format failure and
// use that value unformatted, discarding the first choice entirely. This
// consumes an extra random draw from the RNG stream that must be reproduced
// for byte-identical output, so it can't be modeled as a plain pyFormat
// fallback — it must actually issue a second choice() call to the *same*
// pyrandom.Random when the first one's format fails.
func choiceFormatReroll(rng *pyrandom.Random, pool []string, kwargs map[string]string) string {
	tmpl := choice(rng, pool)
	out, ok := pyFormatStrict(tmpl, kwargs)
	if !ok {
		return choice(rng, pool)
	}
	return out
}

// pyCapitalize reproduces Python's str.capitalize(): first character
// uppercased, the rest of the string lowercased.
func pyCapitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + strings.ToLower(string(r[1:]))
}
