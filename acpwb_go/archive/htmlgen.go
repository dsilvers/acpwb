package archive

import (
	"strconv"
	"strings"
	"unicode"
)

// escape reproduces django.utils.html.escape: & < > " ' -> entities, in that
// substitution order (Python's str.translate applies all mappings in a
// single pass so order doesn't actually matter for correctness, but &amp;
// must be computed before scanning for the other characters it introduces
// wouldn't matter here since translate does one pass over the ORIGINAL
// string, not a series of sequential replaces).
func escape(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#x27;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// truncatechars reproduces Django's `truncatechars` template filter
// (Truncator(value).chars(n), default replacement "…"): if the string is no
// longer than n (counting combining marks as zero-width, per
// unicodedata.combining), return it unchanged; otherwise cut it down so that
// text + "…" is exactly n characters (by the same counting rule) and append
// "…".
func truncatechars(s string, n int) string {
	if n <= 0 {
		return ""
	}
	truncateLen := n - 1 // len("…") == 1, so calculate_truncate_chars_length(n, nil) == n-1
	runes := []rune(s)
	sLen := 0
	endIndex := -1
	for i, r := range runes {
		if isCombining(r) {
			continue
		}
		sLen++
		if endIndex < 0 && sLen > truncateLen {
			endIndex = i
		}
		if sLen > n {
			if endIndex < 0 {
				endIndex = 0
			}
			return string(runes[:endIndex]) + "…"
		}
	}
	return s
}

// truncatewords reproduces Django's `truncatewords` filter: split on
// whitespace (which also normalizes/collapses whitespace, including
// newlines), and if there are more than n words, join the first n with a
// single space and append " …" (a literal two-character suffix, not the
// %(truncated_text)s pattern) unless the joined text already ends with it.
func truncatewords(s string, n int) string {
	if n <= 0 {
		return ""
	}
	words := strings.Fields(s)
	if len(words) > n {
		joined := strings.Join(words[:n], " ")
		if strings.HasSuffix(joined, " …") {
			return joined
		}
		return joined + " …"
	}
	return strings.Join(words, " ")
}

// isCombining approximates Python's unicodedata.combining(char) != 0 using
// Go's Unicode combining-mark range tables. Good enough for the ASCII/basic-
// Latin generated text this codebase produces.
func isCombining(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r)
}

// getArchiveSeal substitutes the year/record_id sentinels in the captured
// static-partial template (see acpwb_go/data/ARCHIVE_SEAL_TEMPLATE.html,
// dumped byte-for-byte from apps.core.htmlgen.get_archive_seal's real
// rendered output — see that function's docstring for why this is safe to
// treat as a fixed template with two substitutions).
func getArchiveSeal(year int, recordID string) string {
	out := archiveSealTemplate
	out = strings.ReplaceAll(out, "__HTMLGEN_SEAL_YEAR__", strconv.Itoa(year))
	out = strings.ReplaceAll(out, "__HTMLGEN_SEAL_RECORD_ID__", recordID)
	return out
}
