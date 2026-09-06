package archive

import "strconv"

// commaInt reproduces Python's f"{n:,}" formatting for non-negative ints
// (all call sites here pass baseline/current, which max(0, ...) guarantees
// non-negative).
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

// signedInt reproduces Python's f"+{d}" if d >= 0 else str(d).
func signedInt(n int) string {
	if n >= 0 {
		return "+" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func dollarComma(n int) string {
	return "$" + commaInt(n)
}
