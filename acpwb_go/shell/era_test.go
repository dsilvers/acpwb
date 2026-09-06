package shell

import (
	"strings"
	"testing"
)

func TestRenderEraPage_Structure(t *testing.T) {
	html := RenderEraPage(EraPageParams{
		EraContentHTML: `<section id="era-test-content">hello era world</section>`,
		Title:          "Q3 Earnings Review",
		TitleSuffix:    "ACPWB Archive",
		OGDescription:  "Technology sector engagement documentation archived 2010-05-12.",
		RequestPath:    "/05/12/q3-earnings-review-1234/",
		RemoteAddr:     "203.0.113.5",
		Year:           2010,
		AllYears:       []int{2025, 2010, 1985},
		YearData: EraYearData{
			Bg: "#F8F9FA", TextColor: "#1A1A2E", Accent: "#1E5F74", Accent2: "#4DA6C8",
			FontBody: "Helvetica, Arial, sans-serif", FontHead: "Helvetica, Arial, sans-serif",
			LayoutClass: "era-generic",
		},
	})

	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Errorf("expected doctype at start")
	}
	if !strings.Contains(html, "<html lang=\"en\">") || !strings.Contains(html, "</html>") {
		t.Errorf("missing <html>/</html>")
	}

	if !strings.Contains(html, "<title>Q3 Earnings Review — ACPWB Archive</title>") {
		t.Errorf("missing expected <title>")
	}
	// twitter:title/twitter:description are NEVER overridden by any era
	// content template in the real Jinja2 source (confirmed against
	// templates/jinja2/honeypot/era/{archive,archive_compliance,archive_minutes}.html
	// — only title/og_title/og_description blocks are overridden) — a real
	// inconsistency reproduced verbatim, not a bug in this port. Check the
	// specific tags directly rather than counting substring occurrences
	// across the whole page: og_title/og_description/the site defaults all
	// also appear (legitimately) inside the JSON-LD garbage and
	// prompt-injection partials, which would make a bare Count() fragile.
	if !strings.Contains(html, `<meta property="og:title" content="Q3 Earnings Review — ACPWB Archive">`) {
		t.Errorf("missing expected og:title")
	}
	if !strings.Contains(html, `<meta property="og:description" content="Technology sector engagement documentation archived 2010-05-12.">`) {
		t.Errorf("missing expected og:description")
	}
	if !strings.Contains(html, `<meta name="twitter:title" content="`+defaultOGTitle+`">`) {
		t.Errorf("expected twitter:title to use the site default, not the page's computed title")
	}
	if !strings.Contains(html, `<meta name="twitter:description" content="`+defaultOGDescription+`">`) {
		t.Errorf("expected twitter:description to use the site default, not the page's computed og_description")
	}
	if !strings.Contains(html, `<meta name="description" content="`+defaultOGDescription+`">`) {
		t.Errorf("expected meta description to use the site default, not the page's computed og_description")
	}
	if !strings.Contains(html, `og:url" content="https://acpwb.com/05/12/q3-earnings-review-1234/"`) {
		t.Errorf("missing og:url with request path")
	}

	// The era content fragment must appear verbatim, wrapped in the
	// archive-era-wrapper + year-footer, inside <main>.
	if !strings.Contains(html, `<section id="era-test-content">hello era world</section>`) {
		t.Errorf("era content missing")
	}
	if !strings.Contains(html, `class="archive-era-wrapper era-generic"`) {
		t.Errorf("missing archive-era-wrapper with layout class")
	}
	if !strings.Contains(html, `--era-accent: #1E5F74;`) {
		t.Errorf("missing year-data CSS custom properties")
	}
	if !strings.Contains(html, "ACPWB Institutional Archive") {
		t.Errorf("missing year-footer partial")
	}
	// Current year (2010) badge should be highlighted; others plain.
	if !strings.Contains(html, `https://archives-2010.acpwb.com/" style="display:inline-block;font-size:.65rem;padding:.2rem .5rem;text-decoration:none;border:1px solid rgba(255,255,255,.2);background:var(--era-accent,var(--gold,#c8a951))`) {
		t.Errorf("expected current-year badge to be highlighted")
	}

	// Nav: era shell has its own 8-item list, distinct from base.html's 10.
	if strings.Count(html, `class="nav-link"`) != 8 {
		t.Errorf("expected 8 nav-link items in era nav, got %d", strings.Count(html, `class="nav-link"`))
	}
	if !strings.Contains(html, `navbar-expand-lg`) {
		t.Errorf("expected navbar-expand-lg (era shell), not navbar-expand-xl (main shell)")
	}

	// Ghost links: era variant uses absolute acpwb.com URLs.
	if !strings.Contains(html, `https://acpwb.com/internal/portal/`) {
		t.Errorf("expected absolute-URL era ghost links")
	}

	// honeypot_token substituted into jsonld + prompt-injection partials
	// (shared, byte-identical with the main shell).
	if !strings.Contains(html, `"identifier": "ACPWB-`) {
		t.Errorf("missing jsonld identifier")
	}
	if !strings.Contains(html, `content provenance record`) {
		t.Errorf("missing prompt-injection span")
	}

	// Footer nav columns reused verbatim from the main shell's constants.
	if !strings.Contains(html, `href="https://acpwb.com/archive/">Archives</a>`) {
		t.Errorf("missing footer archives link")
	}
}
