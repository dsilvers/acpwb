package legacyjob

import (
	"path"
	"strconv"
	"strings"
)

type licenseTpl struct {
	id, name, kind string
	from, to       int
	langs          []string
	extra          bool // may be stapled onto another license as a second notice
	body           string
}

func loadLicenses() []licenseTpl {
	var out []licenseTpl
	entries, _ := fragmentFS.ReadDir("fragments/licenses")
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".txt") {
			continue
		}
		raw, _ := fragmentFS.ReadFile(path.Join("fragments/licenses", ent.Name()))
		t := licenseTpl{id: strings.TrimSuffix(ent.Name(), ".txt"), from: 1960, to: 2100}
		var body []string
		for _, l := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(l, "#") && strings.Contains(l, ":") && len(body) == 0 {
				kv := strings.SplitN(strings.TrimPrefix(l, "#"), ":", 2)
				v := strings.TrimSpace(kv[1])
				switch strings.TrimSpace(kv[0]) {
				case "name":
					t.name = v
				case "kind":
					t.kind = v
				case "from":
					t.from, _ = strconv.Atoi(v)
				case "to":
					t.to, _ = strconv.Atoi(v)
				case "langs":
					t.langs = strings.Split(v, ",")
				case "extra":
					t.extra = v == "yes"
				}
				continue
			}
			body = append(body, l)
		}
		t.body = strings.TrimSpace(strings.Join(body, "\n"))
		out = append(out, t)
	}
	return out
}

var licenseKindWeight = map[string]int{"proprietary": 30, "commercial": 30, "noncommercial": 18, "open": 14, "public": 6, "dual": 8}

// wrapParas wraps each blank-line-separated paragraph, keeping lines that
// look pre-formatted (indented, list items, signature lines) as they are.
func wrapParas(text string, width int) string {
	var out []string
	for _, para := range strings.Split(text, "\n\n") {
		lines := strings.Split(para, "\n")
		pre := false
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(l, " ") || (len(t) > 1 && (t[1] == '.' || t[1] == ')')) || strings.HasPrefix(t, "-") || strings.Contains(l, "    ") {
				pre = true
			}
		}
		if pre || len(lines) > 1 {
			for _, l := range lines {
				if len(l) > width && !strings.HasPrefix(l, " ") {
					out = append(out, wrapWords(l, width)...)
				} else {
					out = append(out, l)
				}
			}
		} else {
			out = append(out, wrapWords(para, width)...)
		}
		out = append(out, "")
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

// buildLicense picks (or composes) an era-appropriate license for the page.
func buildLicense(j *job) (name, text string) {
	x := j.x.fork("license")
	g := j.g
	in := j.in
	author := pick(x, g.roster)
	var tpls []licenseTpl
	var ws []int
	for _, t := range loadLicenses() {
		if in.Year < t.from || in.Year > t.to {
			continue
		}
		if len(t.langs) > 0 && !contains(t.langs, j.id.LangKey) && !contains(t.langs, langFamily(j.id.LangKey)) {
			continue
		}
		w := licenseKindWeight[t.kind]
		if len(t.langs) > 0 {
			w *= 3 // language-native licenses are very likely when they apply
		}
		tpls = append(tpls, t)
		ws = append(ws, w)
	}
	if len(tpls) == 0 || x.chance(0.3) {
		name, text = composeLicense(j, x, author)
	} else {
		t := tpls[x.weighted(ws)]
		name = t.name
		text = g.fill(t.body, author)
	}
	// Occasionally a second notice is stapled on.
	if x.chance(0.25) {
		var extras []licenseTpl
		for _, t := range tpls {
			if (t.extra || t.id == "export-control" || t.id == "ai-training" || t.id == "internal-use" || t.id == "escrow" || t.id == "employee-agreement") && t.name != name {
				extras = append(extras, t)
			}
		}
		if len(extras) > 0 {
			e := pick(x, extras)
			text += "\n\n" + strings.Repeat("-", 40) + "\n\n" + g.fill(e.body, author)
			name += " + " + e.name
		}
	}
	return name, wrapParas(asciiText(text), 78)
}

func composeLicense(j *job, x *rnd, author *Coder) (string, string) {
	L := loadClauses()
	g := j.g
	title := g.fill(pick(x, L["title"]), author)
	var b strings.Builder
	b.WriteString(title + "\n\n")
	b.WriteString(g.fill(pick(x, L["preamble"]), author) + "\n\n")
	n := 1
	// Each section heading has period-appropriate synonyms.
	heads := map[string][]string{
		"definitions":     {"Definitions", "Defined Terms", "Interpretation"},
		"grant":           {"License Grant", "Grant of License", "Scope of License", "Permitted Use"},
		"restriction":     {"Restrictions", "Prohibited Uses", "Limitations on Use", "Conditions of Use"},
		"fees":            {"Fees and Payment", "License Charges", "Payment Terms", "Consideration"},
		"output":          {"Output and Results", "Use of Output", "Reports and Figures"},
		"confidentiality": {"Confidentiality", "Proprietary Rights", "Protection of the Program", "Trade Secrets"},
		"support":         {"Support and Maintenance", "Maintenance Services", "Program Services", "Updates"},
		"audit":           {"Audit Rights", "Verification", "Records and Inspection"},
		"export":          {"Export", "Export Control", "Territory"},
		"govt":            {"Government Use", "U.S. Government Rights", "Restricted Rights"},
		"termination":     {"Term and Termination", "Termination", "Duration"},
		"warranty":        {"Warranty", "Disclaimer of Warranty", "Limited Warranty"},
		"liability":       {"Limitation of Liability", "Limitation of Remedies", "Exclusion of Damages"},
		"law":             {"Governing Law", "Jurisdiction", "Applicable Law"},
		"acceptance":      {"Acceptance", "Acknowledgement", "Assent"},
		"notices":         {"Notices", "Communications"},
		"misc":            {"General", "Miscellaneous", "General Provisions"},
	}
	section := func(cat string, lo, hi int, p float64) {
		if len(L[cat]) == 0 || !x.chance(p) {
			return
		}
		head := pick(x, heads[cat])
		for _, it := range sample(x, L[cat], x.between(lo, hi)) {
			b.WriteString(strconv.Itoa(n) + ". " + head + ". " + g.fill(it, author) + "\n\n")
			n++
		}
	}
	section("definitions", 1, 3, 0.35)
	section("grant", 1, 2, 1)
	section("restriction", 3, 6, 1)
	section("fees", 1, 2, 0.4)
	section("output", 1, 2, 0.45)
	section("confidentiality", 1, 2, 1)
	section("support", 1, 2, 0.35)
	section("audit", 1, 2, 0.3)
	section("export", 1, 1, 0.25)
	section("govt", 1, 1, 0.2)
	section("termination", 1, 2, 1)
	section("warranty", 1, 1, 1)
	section("liability", 1, 2, 1)
	section("law", 1, 1, 1)
	section("notices", 1, 1, 0.3)
	section("acceptance", 1, 1, 0.3)
	section("misc", 1, 3, 1)
	if sig := L["signature"]; len(sig) > 0 && x.chance(0.5) {
		b.WriteString(g.fill(pick(x, sig), author) + "\n\n")
	}
	b.WriteString(g.fill("Copyright (C) {YEARRANGE} {HOLDER}. All rights reserved. License No. {LICNO}.", author))
	return strings.Title(strings.ToLower(title)), b.String()
}

func loadClauses() map[string][]string { return loadDir("fragments/licenses/clauses") }
