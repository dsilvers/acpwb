package compliance

import (
	"fmt"
	"strconv"
	"strings"
)

// File exhibits (data/exhibits/files.txt): a numbered file, log, or
// message list drawn as an editor window; one line is the problem.

type fileLine struct {
	text string
	slot int // 0 = no slot
}

type fileSpec struct {
	kind, title string
	fixed       bool
	asks        []string
	good        []fileLine
	bad         []fileLine
}

func loadFileSpecs() []fileSpec {
	raw, err := dataFS.ReadFile("data/exhibits/files.txt")
	if err != nil {
		panic(err)
	}
	var out []fileSpec
	var cur *fileSpec
	for _, ln := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(ln, "== "):
			parts := strings.Split(ln[3:], " | ")
			if len(parts) != 3 {
				panic("files.txt: bad header " + ln)
			}
			out = append(out, fileSpec{kind: strings.TrimSpace(parts[0]), title: parts[1], fixed: strings.TrimSpace(parts[2]) == "fixed"})
			cur = &out[len(out)-1]
		case cur == nil || strings.HasPrefix(ln, "#") || strings.TrimSpace(ln) == "":
		case strings.HasPrefix(ln, "cats: "):
		case strings.HasPrefix(ln, "ask: "):
			cur.asks = append(cur.asks, ln[5:])
		case strings.HasPrefix(ln, "+ "):
			cur.good = append(cur.good, fileLine{ln[2:], 0})
		case strings.HasPrefix(ln, "- "):
			cur.bad = append(cur.bad, fileLine{ln[2:], 0})
		case ln[0] == '?' || ln[0] == '-':
			n, err := strconv.Atoi(ln[1:2])
			if err != nil || len(ln) < 3 || ln[2] != ' ' {
				panic("files.txt: bad line " + ln)
			}
			if ln[0] == '?' {
				cur.good = append(cur.good, fileLine{ln[3:], n})
			} else {
				cur.bad = append(cur.bad, fileLine{ln[3:], n})
			}
		default:
			panic("files.txt: bad line " + ln)
		}
	}
	return out
}

// fileCats reads each block's "cats:" line.
func fileCats() map[string][]string {
	raw, _ := dataFS.ReadFile("data/exhibits/files.txt")
	cats := map[string][]string{}
	kind := ""
	for _, ln := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(ln, "== ") {
			kind = strings.TrimSpace(strings.Split(ln[3:], " | ")[0])
		} else if strings.HasPrefix(ln, "cats: ") {
			for _, c := range strings.Split(ln[6:], ",") {
				cats[strings.TrimSpace(c)] = append(cats[strings.TrimSpace(c)], kind)
			}
		}
	}
	return cats
}

func init() {
	gens := map[string]exhibitGen{}
	for _, sp := range loadFileSpecs() {
		if len(sp.asks) == 0 || len(sp.bad) == 0 || len(sp.good) < 4 {
			panic("files.txt: incomplete block " + sp.kind)
		}
		gens[sp.kind] = fileGen(sp)
	}
	registerExhibits(gens, fileCats())
}

func fileGen(sp fileSpec) exhibitGen {
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		bad := pick(r, sp.bad)
		var lines []string
		badAt := -1
		if sp.fixed {
			// Bad lines for a slot replace that slot; an unslotted bad line
			// goes anywhere after the first line.
			for _, g := range sp.good {
				switch {
				case g.slot != 0 && g.slot == bad.slot:
					badAt = len(lines)
					lines = append(lines, bad.text)
				default:
					lines = append(lines, g.text)
				}
			}
			if badAt < 0 {
				badAt = r.between(1, len(lines))
				lines = append(lines[:badAt], append([]string{bad.text}, lines[badAt:]...)...)
			}
		} else {
			var goods []string
			for _, g := range sp.good {
				goods = append(goods, g.text)
			}
			n := minInt(len(goods), r.between(5, 8))
			lines = sample(r, goods, n)
			badAt = r.intn(n + 1)
			lines = append(lines[:badAt], append([]string{bad.text}, lines[badAt:]...)...)
		}
		q := Question{key: "ex/" + sp.kind, Exhibit: fileSVG(sp.title, lines)}
		q.Text = fmt.Sprintf("Refer to %s. %s", ex, pick(r, sp.asks))
		var wrongs []string
		for _, i := range sample(r, orderN(r, len(lines)), len(lines)) {
			if i != badAt && len(wrongs) < 3 {
				wrongs = append(wrongs, fmt.Sprintf("Line %d", i+1))
			}
		}
		q.Choices, q.Correct = withChoices(r, fmt.Sprintf("Line %d", badAt+1), wrongs)
		return q
	}
}

// fileSVG draws an editor window with numbered lines. Braces are written as
// character references so code samples survive the "no unfilled slot" checks.
func fileSVG(title string, lines []string) string {
	h := 46 + 17*float64(len(lines)) + 10
	s := newSVG(560, h, title+" with numbered lines")
	s.rect(0.5, 0.5, 559, h-1, 6, "#1e1e2e", "#0f0f17")
	s.rect(0.5, 0.5, 559, 28, 6, "#2d2d3f", "")
	for i, c := range []string{"#ff5f57", "#febc2e", "#28c840"} {
		s.circle(16+14*float64(i), 14.5, 4.5, c, "")
	}
	s.text(66, 18.5, 9.5, `font-weight="700" fill="#e5e7eb"`, title)
	for i, l := range lines {
		y := 48 + 17*float64(i)
		s.text(32, y, 9.5, `text-anchor="end" fill="#6b7280" font-family="ui-monospace,Menlo,Consolas,monospace"`, itoa(i+1))
		t := esc(l)
		t = strings.ReplaceAll(strings.ReplaceAll(t, "{", "&#123;"), "}", "&#125;")
		fmt.Fprintf(s, `<text x="44" y="%.1f" font-size="9.5" fill="#e5e7eb" xml:space="preserve" font-family="ui-monospace,Menlo,Consolas,monospace">%s</text>`, y, t)
	}
	return s.done()
}
