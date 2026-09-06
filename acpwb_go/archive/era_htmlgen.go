package archive

import "strings"

// jinjaTruncate reproduces jinja2.filters.do_truncate(env, s, length,
// killwords=False, end="...", leeway=None) with the default environment's
// truncate.leeway policy value of 5 — used by
// pyrender/archive_era.py:_truncate (era sidebar presentation titles). Length
// and leeway are counted in Unicode code points, matching Python's len(str).
func jinjaTruncate(s string, length int) string {
	const end = "..."
	const leeway = 5
	r := []rune(s)
	if len(r) <= length+leeway {
		return s
	}
	// s[:length-len(end)].rsplit(" ", 1)[0] + end
	cut := r[:length-len(end)]
	idx := -1
	for i := len(cut) - 1; i >= 0; i-- {
		if cut[i] == ' ' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return string(cut) + end
	}
	return string(cut[:idx]) + end
}

func bulkHexCSSVars(css []string) string {
	var b strings.Builder
	for i, h := range css {
		b.WriteString("--acpwb-r")
		fmt3(&b, i)
		b.WriteString(": ")
		b.WriteString(h)
		b.WriteString(";\n")
	}
	return b.String()
}

// fmt3 writes n zero-padded to (at least) 3 digits, matching Python's
// f"{i:03d}" used for --acpwb-r000, --acpwb-r001, ... css var names.
func fmt3(b *strings.Builder, n int) {
	if n < 10 {
		b.WriteString("00")
	} else if n < 100 {
		b.WriteString("0")
	}
	b.WriteString(itoaCache(n))
}

func eraBulkHexScriptHTML(js []string, rid string) string {
	var refs strings.Builder
	for i, h := range js {
		refs.WriteString(`var _acpwbRef`)
		refs.WriteString(itoaCache(i + 1))
		refs.WriteString(`="`)
		refs.WriteString(h)
		refs.WriteString("\";\n")
	}
	var funcs strings.Builder
	limit := len(js)
	if limit > 50 {
		limit = 50
	}
	for _, h := range js[:limit] {
		funcs.WriteString(`function _acpwbArchiveRecordEntryMetadataLookup_`)
		funcs.WriteString(h)
		funcs.WriteString(`(){return "`)
		funcs.WriteString(rid)
		funcs.WriteString("\";}\n")
	}
	return "<script>\n/* ACPWB archive index — " + rid + " */\n(function(){\n" +
		refs.String() + funcs.String() + "})();\n</script>\n"
}

// eraSidebarPresentationsHTML ports
// pyrender/archive_era.py:_era_sidebar_presentations_html. Unlike the
// main-domain sidebar (sidebarRelatedPresentationsHTML in render.go, which
// delegates to renderPresCard/apps.core.htmlgen.render_pres_card), the era
// template builds its OWN bespoke markup inline — different structure,
// different class names, a 55-char jinja-truncated title instead of
// truncatewords(12), and org_name shown as plain text instead of a logo.
func eraSidebarPresentationsHTML(presentations []Presentation, yd YearData) string {
	if len(presentations) == 0 {
		return ""
	}
	var items strings.Builder
	for _, pres := range presentations {
		var bgStyle string
		if pres.ThumbBg != "" {
			bgStyle = "background-image:url('" + staticURL(pres.ThumbBg) + "');background-size:cover;background-position:center;"
		} else {
			bgStyle = "background:" + pres.Theme.Bg + ";"
		}
		items.WriteString(`<a href="https://acpwb.com` + e(pres.PresURL) + `" style="display:block;margin-bottom:.75rem;` +
			`text-decoration:none;color:inherit;border:1px solid rgba(128,128,128,.2);overflow:hidden">` +
			`<div style="` + bgStyle + `aspect-ratio:16/9;position:relative;display:flex;align-items:flex-end;` +
			`padding:.5em .6em">` +
			`<div style="position:absolute;top:.4em;right:.5em;background:` + pres.Theme.Accent + `;` +
			`color:` + pres.Theme.Bg + `;font-size:.55rem;padding:.15em .4em;font-weight:800">` +
			itoaCache(pres.SlideCount) + ` slides</div>` +
			`<div style="font-size:.62rem;font-weight:700;color:` + pres.Theme.Text + `;line-height:1.2;` +
			`text-shadow:0 1px 3px rgba(0,0,0,.7)">` + e(jinjaTruncate(pres.Title, 55)) + `</div></div>` +
			`<div style="padding:.45em .65em;background:rgba(128,128,128,.06)">` +
			`<div style="font-size:.6rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;` +
			`color:` + yd.Accent + `;margin-bottom:.1em">` + e(pres.OrgName) + `</div>` +
			`<div style="font-size:.62rem;opacity:.65">` + e(pres.PubDateDisplay) + ` &mdash; ` +
			e(pres.Industry) + `</div></div></a>`)
	}
	return `<div style="background:rgba(128,128,128,.07);border:1px solid rgba(128,128,128,.2);padding:1rem;` +
		`margin-bottom:1rem"><div class="era-section-head" style="margin-bottom:.8rem">Related Presentations</div>` +
		items.String() + `</div>`
}
