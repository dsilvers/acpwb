package legacyjob

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

func TestEverySlotHasAHandler(t *testing.T) {
	re := regexp.MustCompile(`\{([A-Z][A-Z0-9.]*)\}`)
	slots := map[string]bool{}
	_ = fs.WalkDir(fragmentFS, "fragments", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, _ := fragmentFS.ReadFile(p)
		for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
			slots[m[1]] = true
		}
		return nil
	})
	var missing []string
	orig := unknownSlot
	defer func() { unknownSlot = orig }()
	unknownSlot = func(name string) string { missing = append(missing, name); return "" }
	in := &Input{Year: 2001, Month: 5, Day: 9, Slug: "x", Token: "abcdef12", Org: "Org", Path: "/p"}
	g := &gen{x: newRnd("t"), in: in, emails: map[string]bool{}, usedTpl: map[string]int{}, holder: "H", yearFrom: 1990, id: Identity(2001, 5, 9, "x")}
	g.roster = []*Coder{{First: "Ann", Last: "Lee", Initials: "AL", Start: 1990, End: 2001}, {First: "Bo", Last: "Ng", Initials: "BN", Start: 1995, End: 2001}}
	for s := range slots {
		g.fill("{"+s+"}", g.roster[0])
	}
	if len(missing) > 0 {
		t.Fatalf("slots without handlers: %s", strings.Join(missing, ", "))
	}
}

// Every fragments/vars slot must have at least one value for every archive
// year, or the page would print the bare slot name.
func TestVarsCoverEveryYear(t *testing.T) {
	for name, vals := range getLibrary().vars {
		for y := 1985; y <= 2025; y++ {
			ok := false
			for _, v := range vals {
				if y >= v.From && y <= v.To {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("vars/%s has no value for %d", strings.ToLower(name), y)
			}
		}
	}
}

func TestEraPrefix(t *testing.T) {
	for in, want := range map[string]eraVal{
		"1990-|x":       {1990, 9999, "x"},
		"-1995|y":       {0, 1995, "y"},
		"1986-1986|z":   {1986, 1986, "z"},
		"plain|text":    {0, 9999, "plain|text"},
		"no prefix":     {0, 9999, "no prefix"},
		"1990-2000|  w": {1990, 2000, "w"},
	} {
		if got := parseEra(in); got != want {
			t.Errorf("parseEra(%q) = %+v, want %+v", in, got, want)
		}
	}
}
