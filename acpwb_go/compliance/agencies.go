package compliance

import (
	"fmt"
	"strings"
)

// Mandating agencies: the 853 real ones from the policy pages
// (acpwb_go/data AGENCIES), plus generated regulators whose slug is
// "<region>-<body>-<domain>" (e.g. wi-board-workplace-ergonomics →
// "Wisconsin Board of Workplace Ergonomics"), built from data/bodies/.
// Region and body slugs never contain hyphens, so a generated slug always
// has 3+ parts and parses back unambiguously; real slugs are tried first.

type agency struct {
	Slug, Name, Domain string
	Code               string // short code for seals
}

type bodyPart struct{ Slug, Name, Desc string }

var (
	bodyRegions []bodyPart
	bodyKinds   []bodyPart // Name is the pattern, e.g. "{R} Board of {D}"
	bodyDomains []bodyPart
	regionIdx   = map[string]int{}
	kindIdx     = map[string]int{}
	domainIdx   = map[string]int{}
)

func loadBodies() {
	parse := func(file string, n int) []bodyPart {
		var out []bodyPart
		for _, l := range lines(readData("data/bodies/" + file)) {
			f := strings.Split(l, "|")
			if len(f) != n {
				panic("compliance: bad line in " + file + ": " + l)
			}
			for i := range f {
				f[i] = strings.TrimSpace(f[i])
			}
			p := bodyPart{Slug: f[0], Name: f[1]}
			if n > 2 {
				p.Desc = f[2]
			}
			out = append(out, p)
		}
		return out
	}
	bodyRegions = parse("regions.txt", 2)
	bodyKinds = parse("kinds.txt", 2)
	bodyDomains = parse("domains.txt", 3)
	for i, p := range bodyRegions {
		regionIdx[p.Slug] = i
	}
	for i, p := range bodyKinds {
		kindIdx[p.Slug] = i
	}
	for i, p := range bodyDomains {
		domainIdx[p.Slug] = i
	}
}

func generatedAgency(ri, ki, di int) agency {
	reg, kind, dom := bodyRegions[ri], bodyKinds[ki], bodyDomains[di]
	name := strings.NewReplacer("{R}", reg.Name, "{D}", dom.Name).Replace(kind.Name)
	initials := ""
	for _, w := range strings.Fields(strings.NewReplacer("{R}", "", "{D}", dom.Name, "&", "", " of ", " ", " on ", " ", " for ", " ").Replace(kind.Name)) {
		initials += strings.ToUpper(w[:1])
	}
	return agency{
		Slug:   reg.Slug + "-" + kind.Slug + "-" + dom.Slug,
		Name:   name,
		Domain: dom.Desc,
		Code:   strings.ToUpper(reg.Slug) + "-" + initials,
	}
}

// lookupAgency resolves a real or generated agency slug.
func lookupAgency(slug string) (agency, bool) {
	if i, ok := agencyMap[slug]; ok {
		return agencies[i], true
	}
	parts := strings.SplitN(slug, "-", 3)
	if len(parts) != 3 {
		return agency{}, false
	}
	ri, ok1 := regionIdx[parts[0]]
	ki, ok2 := kindIdx[parts[1]]
	di, ok3 := domainIdx[parts[2]]
	if !ok1 || !ok2 || !ok3 {
		return agency{}, false
	}
	return generatedAgency(ri, ki, di), true
}

func generatedCount() int { return len(bodyRegions) * len(bodyKinds) * len(bodyDomains) }

// agencyAt indexes the whole agency space: real agencies first, then every
// generated one.
func agencyAt(i int) agency {
	if i < len(agencies) {
		return agencies[i]
	}
	i -= len(agencies)
	nk, nd := len(bodyKinds), len(bodyDomains)
	return generatedAgency(i/(nk*nd)%len(bodyRegions), i/nd%nk, i%nd)
}

func agencyTotal() int { return len(agencies) + generatedCount() }

// randomAgency picks a real agency half the time, a generated one otherwise.
func randomAgency(r *rnd) agency {
	if r.chance(0.5) {
		return pick(r, agencies)
	}
	return agencyAt(len(agencies) + r.intn(generatedCount()))
}

func sampleAgencies(r *rnd, n int) []agency {
	out := make([]agency, n)
	for i := range out {
		out[i] = randomAgency(r)
	}
	return out
}

const agenciesPerPage = 120

// agencyDirectoryPage is one page of /compliance/agencies/: the real
// agencies first, then the generated ones in order.
func agencyDirectoryPage(page int) []agency {
	var out []agency
	for i := (page - 1) * agenciesPerPage; i < page*agenciesPerPage && i < agencyTotal(); i++ {
		out = append(out, agencyAt(i))
	}
	return out
}

func agencyPages() int { return (agencyTotal() + agenciesPerPage - 1) / agenciesPerPage }

func agencyURL(a agency) string { return "/compliance/" + a.Slug + "/" }

func (a agency) String() string { return fmt.Sprintf("%s (%s)", a.Name, a.Slug) }
