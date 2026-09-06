package archive

import (
	"fmt"
	"strings"

	"acpwb_go/pyrandom"
)

// Author carries exactly the fields
// apps/core/htmlgen.py:render_pres_card reads off each authors[] entry.
type Author struct {
	FullName   string
	AvatarSeed string
	Initials   string
}

// Presentation carries exactly the fields render_pres_card reads (plus the
// bookkeeping fields needed to reach them — pres_seed/slug/etc. aren't
// rendered but are kept for parity/debugging).
type Presentation struct {
	OrgSlug        string
	OrgName        string
	Title          string
	PubDateDisplay string
	Industry       string
	SlideCount     int
	ThumbBg        string
	Theme          presTheme
	Authors        []Author
	PresURL        string
}

var monthAbbrev3 = []string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

func pubDateDisplay(year, month, day int) string {
	return fmt.Sprintf("%s %d, %d", monthAbbrev3[month-1], day, year)
}

// fillEight ports presentations/generators.py:_fill's unconditional 8-draw
// dict literal (n, m, x, pct, months, years, regions, year — always drawn in
// this order regardless of which placeholders the template actually uses,
// since Python evaluates every value in a dict literal eagerly), then
// applies pyFormat with those substitutions merged with any extra ones.
func fillEight(rng *pyrandom.Random, template string, extra map[string]string) string {
	kw := map[string]string{
		"n":       itoaCache(int(rng.RandInt(15, 87))),
		"m":       itoaCache(int(rng.RandInt(2, 480))),
		"x":       itoaCache(int(rng.RandInt(2, 10))),
		"pct":     itoaCache(int(rng.RandInt(12, 94))),
		"months":  itoaCache(int(rng.RandInt(3, 18))),
		"years":   itoaCache(int(rng.RandInt(3, 20))),
		"regions": itoaCache(int(rng.RandInt(4, 12))),
		"year":    itoaCache(int(rng.RandInt(2015, 2025))),
	}
	for k, v := range extra {
		kw[k] = v
	}
	return pyFormat(template, kw)
}

// generateTitleForSlug ports presentations/generators.py:_generate_title,
// consuming rng in the exact same order (template choice, adjective choice,
// then fillEight's 8 draws) — its RETURN VALUE matters here (unlike in the
// real app's slide generation) because it feeds slugify() to build the
// slug that generate_presentation_meta later decodes back into the
// displayed card title.
func generateTitleForSlug(rng *pyrandom.Random, industry, domain, verb, noun string) string {
	tmpl := choice(rng, presTitleTemplates)
	adjClean := strings.ReplaceAll(choice(rng, presAdjectives), "-", " ")
	return fillEight(rng, tmpl, map[string]string{
		"industry": industry, "domain": domain, "verb": verb, "adj": adjClean, "noun": noun,
	})
}

func slugFromTitle(title string, num int) string {
	return fmt.Sprintf("%s-%d", slugify(title), num)
}

// GeneratePresentationsForContext ports
// presentations/generators.py:generate_presentations_for_context.
func GeneratePresentationsForContext(contextSeed string, count int) []Presentation {
	results := make([]Presentation, 0, count)
	usedSlugs := make(map[string]bool)

	for i := 0; i < count; i++ {
		itemRng := rngC(fmt.Sprintf("%s_item%d", contextSeed, i))
		orgName := choice(itemRng, presOrganizations)
		orgSlug := slugify(orgName)
		year := int(itemRng.RandInt(2008, 2025))
		month := int(itemRng.RandInt(1, 12))
		day := int(itemRng.RandInt(1, 28))

		industry := choice(itemRng, archiveIndustries) // INDUSTRIES == ARCHIVE_INDUSTRIES
		domain := choice(itemRng, presDomains)
		verb := choice(itemRng, presVerbs)
		noun := choice(itemRng, presNouns)
		title := generateTitleForSlug(itemRng, industry, domain, verb, noun)
		num := int(itemRng.RandInt(1000, 9999))
		slug := slugFromTitle(title, num)
		for usedSlugs[slug] {
			num = int(itemRng.RandInt(1000, 9999))
			slug = slugFromTitle(title, num)
		}
		usedSlugs[slug] = true

		results = append(results, generatePresentationMeta(orgSlug, year, month, day, slug))
	}
	return results
}

// generatePresentationMeta ports presentations/generators.py:generate_presentation_meta.
func generatePresentationMeta(orgSlug string, year, month, day int, slug string) Presentation {
	presSeed := fmt.Sprintf("pres_%s_%d_%02d_%02d_%s", orgSlug, year, month, day, slug)
	rng := rngC(presSeed)

	orgName, ok := presOrgSlugMap[orgSlug]
	if !ok {
		orgName = pyTitle(strings.ReplaceAll(orgSlug, "-", " "))
	}
	industry := choice(rng, archiveIndustries)
	domain := choice(rng, presDomains)
	verb := choice(rng, presVerbs)
	noun := choice(rng, presNouns)

	title := titleFromSlug(slug)

	adj := choice(rng, presAdjectives)
	subtitleTmpl := choice(rng, presSubtitles)
	_ = fillEight(rng, subtitleTmpl, map[string]string{
		"industry": industry, "domain": domain, "org": orgName, "verb": verb, "noun": noun, "adj": adj,
	})

	venue := choice(rng, presVenues)
	venueNameTmpl := venue[1]
	_ = fillEight(rng, venueNameTmpl, map[string]string{"industry": industry, "domain": domain})

	// author_count = rng.choices([1, 2, 3], weights=[30, 45, 25])[0]
	authorCount := int(pyrandom.Choices(rng, []int64{1, 2, 3}, []float64{30, 45, 25}, nil, 1)[0])
	authors := generateAuthors(presSeed, authorCount)

	theme := choice(rng, presThemes)
	slideCount := int(rng.RandInt(10, 20))

	thumbBg := pickBackground(presSeed, 0)

	presURL := fmt.Sprintf("/presentations/%s/%d/%02d/%02d/%s/", orgSlug, year, month, day, slug)

	return Presentation{
		OrgSlug: orgSlug, OrgName: orgName, Title: title,
		PubDateDisplay: pubDateDisplay(year, month, day),
		Industry:       industry, SlideCount: slideCount, ThumbBg: thumbBg,
		Theme: theme, Authors: authors, PresURL: presURL,
	}
}

func generateAuthors(presSeed string, count int) []Author {
	// presentations/generators.py's own _rng_from_seed is the mod-2**32
	// variant (pattern C), unlike report_generator.py's / policy_generator.py's
	// same-named helper (pattern B) — see rng.go's doc comments.
	rng := rngC("authors_" + presSeed)
	usedEmails := make(map[string]bool)
	authors := make([]Author, 0, count)
	for i := 0; i < count; i++ {
		first := choice(rng, firstNames)
		last := choice(rng, lastNames)
		firstFirst := strings.Fields(first)[0]
		emailBase := strings.ToLower(firstFirst) + "." + strings.ToLower(last) + "@acpwb.com"
		email := emailBase
		n := 2
		for usedEmails[email] {
			email = fmt.Sprintf("%s.%s%d@acpwb.com", strings.ToLower(firstFirst), strings.ToLower(last), n)
			n++
		}
		usedEmails[email] = true
		avatarSeed := md5Hex(fmt.Sprintf("%s%s%d", first, last, i))[:16]
		initials := strings.ToUpper(string([]rune(first)[0]) + string([]rune(last)[0]))
		_ = choice(rng, peopleTitles)      // title (unused by render_pres_card)
		_ = choice(rng, peopleDepartments) // department (unused)
		authors = append(authors, Author{
			FullName:   first + " " + last,
			AvatarSeed: avatarSeed,
			Initials:   initials,
		})
	}
	return authors
}

// pickBackground ports presentations/image_selector.py:pick_background. The
// real function returns None if the chosen bg_NNNNN.webp file doesn't exist
// on disk; the full 5000-image pool is confirmed present in this codebase
// (verified against the checked-in static/img/presentations/backgrounds/
// directory), so that branch is unreachable and not reproduced.
func pickBackground(presSeed string, slideNum int) string {
	rng := rngC(fmt.Sprintf("%s_bg_%d", presSeed, slideNum))
	n := int(rng.RandInt(0, 4999))
	return fmt.Sprintf("img/presentations/backgrounds/bg_%05d.webp", n)
}

// titleFromSlug ports presentations/generators.py:_title_from_slug.
func titleFromSlug(slug string) string {
	idx := strings.LastIndex(slug, "-")
	body := slug
	if idx >= 0 {
		suffix := slug[idx+1:]
		if suffix != "" && isAllDigits(suffix) {
			body = slug[:idx]
		}
	}
	return smartTitle(strings.ReplaceAll(body, "-", " "))
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// smartTitle ports presentations/generators.py:_smart_title.
func smartTitle(s string) string {
	words := strings.Fields(s)
	out := make([]string, len(words))
	for i, w := range words {
		lower := strings.ToLower(w)
		if acr, ok := presAcronyms[lower]; ok {
			out[i] = acr
		} else if i == 0 || !presTitleCaseLower[lower] {
			out[i] = pyCapitalize(w)
		} else {
			out[i] = lower
		}
	}
	return strings.Join(out, " ")
}
