package compliance

import (
	"path"
	"regexp"
	"strings"
)

// Question-bank text is a small grammar, like the archive and source-code
// phrase libraries:
//
//	[a|b|c]    one alternative, chosen at random (nestable, may hold slots)
//	[?text]    optional: included about half the time ([?a|b]: optional choice)
//	{SLOT}     a slot value (newSlots / data/vars/<SLOT>.txt), fixed per question
//	{@name}    a random line of data/fragments/<name>.txt, itself expanded
//
// One question's text and all its answers share a single slot set, so
// "{NAME}" is the same person throughout.

var fragments = map[string][]string{}

func loadFragments() {
	entries, _ := dataFS.ReadDir("data/fragments")
	for _, e := range entries {
		fragments[strings.TrimSuffix(e.Name(), ".txt")] = lines(readData(path.Join("data/fragments", e.Name())))
	}
}

type gen struct {
	r  *rnd
	sl slots
}

func (g gen) expand(s string) string { return fixArticles(g.expandDepth(s, 0)) }

var aBeforeVowel = regexp.MustCompile(`\b([Aa]) ([aeio][a-z]*)`)

// fixArticles turns "a" into "an" before a slot value that starts with a
// vowel ("a employee roster" → "an employee roster"). Words beginning with
// "one" keep "a"; "u" words are left alone ("a user", "an update" are both
// written explicitly).
func fixArticles(s string) string {
	return aBeforeVowel.ReplaceAllStringFunc(s, func(m string) string {
		art, word, _ := strings.Cut(m, " ")
		if strings.HasPrefix(word, "one") {
			return m
		}
		return art + "n " + word
	})
}

func (g gen) expandDepth(s string, depth int) string {
	if depth > 8 {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			end := matching(s, i, '[', ']')
			if end < 0 {
				b.WriteString(s[i:])
				return b.String()
			}
			inner := s[i+1 : end]
			if strings.HasPrefix(inner, "?") {
				// [?a|b] is optional, and picks one alternative when present.
				if g.r.chance(0.5) {
					b.WriteString(g.expandDepth(pick(g.r, splitTop(inner[1:])), depth+1))
				}
			} else {
				b.WriteString(g.expandDepth(pick(g.r, splitTop(inner)), depth+1))
			}
			i = end
		case '{':
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				b.WriteString(s[i:])
				return b.String()
			}
			name := s[i+1 : i+end]
			switch {
			case strings.HasPrefix(name, "@") && len(fragments[name[1:]]) > 0:
				b.WriteString(g.expandDepth(pick(g.r, fragments[name[1:]]), depth+1))
			case g.sl[name] != "":
				b.WriteString(g.sl[name])
			default:
				b.WriteString(s[i : i+end+1]) // left visible; tests catch it
			}
			i += end
		default:
			b.WriteByte(s[i])
		}
	}
	if depth > 0 {
		return b.String() // only the finished text is tidied; inner spaces matter
	}
	return tidy(b.String())
}

// matching finds the bracket closing the one at s[i].
func matching(s string, i int, open, close byte) int {
	depth := 0
	for j := i; j < len(s); j++ {
		switch s[j] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}

// splitTop splits on '|' outside nested brackets.
func splitTop(s string) []string {
	var out []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
		case '|':
			if depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	return append(out, s[start:])
}

// tidy collapses doubled spaces and spaces before punctuation left by
// skipped optional fragments.
func tidy(s string) string {
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	for _, p := range []string{" ,", " .", " ?", " :", " ;"} {
		s = strings.ReplaceAll(s, p, p[1:])
	}
	return strings.TrimSpace(s)
}
