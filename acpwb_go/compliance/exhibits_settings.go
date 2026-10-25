package compliance

import (
	"fmt"
	"strings"
)

// Settings exhibits (data/exhibits/settings.txt): a settings screen where
// one value breaks policy ("which setting"), or zero to two do ("how many").

type setRow struct{ label, good, bad string }

type setSpec struct {
	kind, title, noun string
	ask, count        string
	rows              []setRow
}

func loadSetSpecs() ([]setSpec, map[string][]string) {
	raw, err := dataFS.ReadFile("data/exhibits/settings.txt")
	if err != nil {
		panic(err)
	}
	var out []setSpec
	cats := map[string][]string{}
	for _, ln := range strings.Split(string(raw), "\n") {
		cur := len(out) - 1
		switch {
		case strings.HasPrefix(ln, "#") || strings.TrimSpace(ln) == "":
		case strings.HasPrefix(ln, "== "):
			p := strings.Split(ln[3:], " | ")
			if len(p) != 3 {
				panic("settings.txt: bad header " + ln)
			}
			out = append(out, setSpec{kind: strings.TrimSpace(p[0]), title: p[1], noun: strings.TrimSpace(p[2])})
		case strings.HasPrefix(ln, "cats: "):
			for _, c := range strings.Split(ln[6:], ",") {
				cats[strings.TrimSpace(c)] = append(cats[strings.TrimSpace(c)], out[cur].kind)
			}
		case strings.HasPrefix(ln, "ask: "):
			out[cur].ask = ln[5:]
		case strings.HasPrefix(ln, "count: "):
			out[cur].count = ln[7:]
		case strings.HasPrefix(ln, "~ "):
			p := strings.Split(ln[2:], " | ")
			if len(p) != 3 {
				panic("settings.txt: bad row " + ln)
			}
			out[cur].rows = append(out[cur].rows, setRow{p[0], p[1], p[2]})
		default:
			panic("settings.txt: bad line " + ln)
		}
	}
	return out, cats
}

func init() {
	specs, cats := loadSetSpecs()
	gens := map[string]exhibitGen{}
	for _, sp := range specs {
		nbad := 0
		for _, rw := range sp.rows {
			if rw.good != rw.bad {
				nbad++
			}
		}
		if sp.ask == "" || sp.count == "" || nbad < 4 || len(sp.rows) < 5 {
			panic("settings.txt: incomplete block " + sp.kind)
		}
		gens[sp.kind] = setGen(sp)
	}
	registerExhibits(gens, cats)
}

func setGen(sp setSpec) exhibitGen {
	return func(r *rnd, c *Course, sl slots, ex string) Question {
		n := minInt(len(sp.rows), r.between(5, 7))
		idx := sample(r, orderN(r, len(sp.rows)), n)
		sortInts(idx)
		var badable []int // positions in idx whose row can be shown bad
		for i, k := range idx {
			if sp.rows[k].good != sp.rows[k].bad {
				badable = append(badable, i)
			}
		}
		for len(badable) < 3 { // need enough candidates; resample
			return setGen(sp)(r, c, sl, ex)
		}
		countMode := r.chance(0.35)
		nb := 1
		if countMode {
			nb = r.between(0, 2)
		}
		badSet := map[int]bool{}
		for _, i := range sample(r, badable, nb) {
			badSet[i] = true
		}
		h := 48 + 30*float64(n) + 8
		s := newSVG(520, h, sp.title+" settings screen")
		s.rect(0.5, 0.5, 519, h-1, 10, "#f8fafc", "#cbd5e1")
		s.rect(0.5, 0.5, 519, 36, 10, "#1e293b", "")
		s.rect(0.5, 26, 519, 11, 0, "#1e293b", "")
		s.text(18, 23, 10.5, `font-weight="700" fill="#f8fafc"`, "⚙  "+sp.title)
		for i, k := range idx {
			rw := sp.rows[k]
			v := rw.good
			if badSet[i] {
				v = rw.bad
			}
			y := 48 + 30*float64(i)
			s.rect(12, y, 496, 26, 5, "#fff", "#e2e8f0")
			s.text(24, y+16.5, 9.5, `fill="#0f172a"`, rw.label)
			switch v {
			case "On", "Off":
				on := v == "On"
				bg := "#cbd5e1"
				kx := 468.0
				if on {
					bg, kx = "#16a34a", 482
				}
				s.rect(460, y+6, 36, 14, 7, bg, "")
				s.circle(kx, y+13, 5.5, "#fff", "")
				s.text(452, y+16.5, 8.5, `text-anchor="end" fill="#475569"`, v)
			default:
				s.text(496, y+16.5, 9, `text-anchor="end" font-weight="700" fill="#1e3a8a"`, v)
			}
		}
		q := Question{key: "ex/" + sp.kind, Exhibit: s.done()}
		if countMode {
			q.Text = fmt.Sprintf("Refer to %s. %s", ex, sp.count)
			q.Choices, q.Correct = numChoices(r, nb, []int{1, 2, 3, -1, -2}, itoa)
			return q
		}
		var right string
		var wrongs []string
		for i, k := range idx {
			if badSet[i] {
				right = sp.rows[k].label
			} else {
				wrongs = append(wrongs, sp.rows[k].label)
			}
		}
		q.Text = fmt.Sprintf("Refer to %s. %s", ex, sp.ask)
		q.Choices, q.Correct = withChoices(r, right, sample(r, wrongs, 3))
		return q
	}
}
