package legacyjob_test

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"acpwb_go/archive"
	"acpwb_go/legacyjob"
)

// pagesFor finds n distinct pages whose compute language is k.
func pagesFor(k string, n int) []legacyjob.Input {
	var out []legacyjob.Input
	for i := 0; len(out) < n && i < 200000; i++ {
		y := 1985 + i%41
		slug := fmt.Sprintf("variety-%s-%d", k, i)
		if legacyjob.Identity(y, 1+i%12, 1+i%28, slug).LangKey == k {
			out = append(out, archive.SourceCodeInput(y, 1+i%12, 1+i%28, slug, i%2 == 0))
		}
	}
	return out
}

var (
	wordRe = regexp.MustCompile(`[A-Za-z_#@$&][A-Za-z0-9_#@$.\-]*`)
	numRe  = regexp.MustCompile(`[0-9]+(\.[0-9]+)?`)
)

// lineSets returns a listing's statement lines (sequence columns 73-80
// dropped), raw and as "shapes" with every identifier and number replaced,
// so renaming alone does not count as variety.
func lineSets(l *legacyjob.Listing) (raw, shape map[string]bool) {
	raw, shape = map[string]bool{}, map[string]bool{}
	for _, t := range legacyjob.CodeLinesOf(l) {
		if len(t) > 72 && (l.ID.LangKey == "f4" || l.ID.LangKey == "f77" || l.ID.LangKey == "cobol") {
			t = t[:72]
		}
		t = strings.TrimSpace(t)
		if len(t) < 6 {
			continue
		}
		raw[t] = true
		shape[numRe.ReplaceAllString(wordRe.ReplaceAllString(t, "w"), "n")] = true
	}
	return
}

func jaccard(sets []map[string]bool) float64 {
	var total float64
	pairs := 0
	for i := 0; i < len(sets); i++ {
		for j := i + 1; j < len(sets); j++ {
			inter := 0
			for l := range sets[i] {
				if sets[j][l] {
					inter++
				}
			}
			if union := len(sets[i]) + len(sets[j]) - inter; union > 0 {
				total += float64(inter) / float64(union)
				pairs++
			}
		}
	}
	if pairs == 0 {
		return 0
	}
	return total / float64(pairs)
}

// languageVariety: mean pairwise Jaccard similarity of statement-line sets
// between pages of one language, raw and by shape (lower is more varied).
func languageVariety(k string, n int) (raw, shape float64) {
	var rs, ss []map[string]bool
	for _, in := range pagesFor(k, n) {
		in := in
		r, s := lineSets(legacyjob.BuildFresh(&in))
		rs, ss = append(rs, r), append(ss, s)
	}
	return jaccard(rs), jaccard(ss)
}

// Every language's pages must differ from each other. VARIETY_REPORT=1
// prints the per-language table.
func TestLanguageVariety(t *testing.T) {
	type row struct {
		k          string
		raw, shape float64
	}
	var rows []row
	for _, k := range legacyjob.LanguageKeys() {
		r, s := languageVariety(k, 12)
		rows = append(rows, row{k, r, s})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].shape > rows[j].shape })
	if os.Getenv("VARIETY_REPORT") != "" {
		for _, r := range rows {
			fmt.Printf("%-10s raw %.3f  shape %.3f\n", r.k, r.raw, r.shape)
		}
	}
	for _, r := range rows {
		if r.raw > maxRawSimilarity || r.shape > maxShapeSimilarity {
			t.Errorf("%s: pages too similar (raw %.3f, shape %.3f)", r.k, r.raw, r.shape)
		}
	}
}

// Ceilings ratchet down as generators gain variety.
const (
	maxRawSimilarity   = 0.07
	maxShapeSimilarity = 0.71
)
