package policy

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// escape reproduces django.utils.html.escape (same behavior as
// archive/htmlgen.go:escape, duplicated here per this package's isolation).
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

// pyCapitalize reproduces Python's str.capitalize(): first character
// uppercased, the rest lowercased.
func pyCapitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + strings.ToLower(string(r[1:]))
}

// pyTitle reproduces Python's str.title().
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

// wordCapitalizeJoin reproduces ' '.join(w.capitalize() for w in s.split()).
func wordCapitalizeJoin(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = pyCapitalize(w)
	}
	return strings.Join(words, " ")
}

// commaInt reproduces Python's f"{n:,}" for non-negative ints (all call
// sites here only ever pass non-negative values).
func commaInt(n int) string {
	s := strconv.Itoa(n)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// pyFloorDiv reproduces Python's `//` (floor division) for ints.
func pyFloorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// pyRound1 reproduces Python's round(x, 1): correctly-rounded to 1 decimal
// place (ties-to-even), by delegating to Go's correctly-rounded decimal
// formatter and parsing the result back.
func pyRound1(x float64) float64 {
	s := strconv.FormatFloat(x, 'f', 1, 64)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// pyFloatStr reproduces Python's str(float) for the specific case of a
// value already rounded to at most 1 decimal place: shortest round-trip
// decimal representation, always with at least one digit after the point.
func pyFloatStr(x float64) string {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// pyFormatStrict is a minimal stand-in for Python's str.format(**kwargs):
// "{name}" placeholders, plus the one format-spec shape actually used
// across the policy template pools — "{name:0Nd}" zero-padding an integer
// field to N digits (verified live: only the footnote pool's `seq` field
// ever carries a spec, as 02d/03d/06d). Every value is otherwise
// pre-formatted into its final string form by the caller before being
// passed in here. ok is false if the template references a name missing
// from kwargs (Python KeyError).
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
			field := tmpl[i+1 : i+1+end]
			name := field
			spec := ""
			if idx := strings.IndexByte(field, ':'); idx >= 0 {
				name = field[:idx]
				spec = field[idx+1:]
			}
			v, ok := kwargs[name]
			if !ok {
				return tmpl, false
			}
			if spec != "" {
				v = applyIntFormatSpec(v, spec)
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

// applyIntFormatSpec handles the "0Nd" zero-pad-integer spec shape (see
// pyFormatStrict's doc comment). Any other spec shape is returned
// unmodified (none appear in this codebase's template pools).
func applyIntFormatSpec(v, spec string) string {
	if len(spec) >= 2 && spec[0] == '0' && spec[len(spec)-1] == 'd' {
		width, err := strconv.Atoi(spec[1 : len(spec)-1])
		if err == nil {
			neg := strings.HasPrefix(v, "-")
			digits := v
			if neg {
				digits = v[1:]
			}
			for len(digits) < width {
				digits = "0" + digits
			}
			if neg {
				return "-" + digits
			}
			return digits
		}
	}
	return v
}

// pyFormat formats with fallback-to-raw-template-on-missing-key.
func pyFormat(tmpl string, kwargs map[string]string) string {
	out, ok := pyFormatStrict(tmpl, kwargs)
	if !ok {
		return tmpl
	}
	return out
}

var monthFullNames = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// isValidDate mirrors whether datetime.date(y, m, d) would succeed.
func isValidDate(y, m, d int) bool {
	if m < 1 || m > 12 || d < 1 {
		return false
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return t.Year() == y && int(t.Month()) == m && t.Day() == d
}

// safeDate mirrors the common `try: datetime.date(y,m,d) except ValueError:
// datetime.date(y,m,1)` pattern used throughout policy_generator.py.
func safeDate(y, m, d int) (int, int, int) {
	if isValidDate(y, m, d) {
		return y, m, d
	}
	return y, m, 1
}

// filingDateDetail reproduces
// datetime.date(year,month,day).strftime('%B %d, %Y').replace(' 0', ' ')
// (generate_policy_document's variant — %04d year fallback) which is
// byte-identical to strftime('%B %-d, %Y') (the _generate_doc_stub variant)
// for any valid date, since both produce a non-zero-padded day number.
func filingDateDetail(year, month, day int) string {
	if !isValidDate(year, month, day) {
		return strconv_Sprintf04(year, month, day)
	}
	return monthFullNames[month-1] + " " + strconv.Itoa(day) + ", " + strconv.Itoa(year)
}

// filingDateStub reproduces _generate_doc_stub's
// strftime('%B %-d, %Y') / f"{year}-{month:02d}-{day:02d}" fallback (year
// NOT zero-padded, unlike filingDateDetail's fallback).
func filingDateStub(year, month, day int) string {
	if !isValidDate(year, month, day) {
		return strconv.Itoa(year) + "-" + pad2(month) + "-" + pad2(day)
	}
	return monthFullNames[month-1] + " " + strconv.Itoa(day) + ", " + strconv.Itoa(year)
}

func pad2(n int) string {
	s := strconv.Itoa(n)
	if len(s) < 2 {
		return "0" + s
	}
	return s
}

func pad4(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

func strconv_Sprintf04(year, month, day int) string {
	return pad4(year) + "-" + pad2(month) + "-" + pad2(day)
}

// dedupeStrings reproduces list(dict.fromkeys(items)): first occurrence of
// each distinct value, in original order.
func dedupeStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, v := range items {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// truncate72 reproduces pyrender/policy.py:_truncate72 — title[:72] + ('…'
// if len(title) > 72 else ”), counting Unicode code points as Python's str
// slicing/len do.
func truncate72(title string) string {
	r := []rune(title)
	if len(r) > 72 {
		return string(r[:72]) + "…"
	}
	return title
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
