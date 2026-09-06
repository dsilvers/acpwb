// era.go ports the SEPARATE page shell used by the archive subdomain
// ("era") pages — templates/jinja2/honeypot/archive_era_base.html plus the
// three honeypot partials it includes from templates/jinja2/partials/ (two
// of which, _jsonld_garbage.html and _prompt_injection.html, are byte-
// identical to the main shell's Django-template versions already ported in
// shell.go and are reused here; _ghost_links.html is NOT identical — the
// jinja2 copy uses absolute https://acpwb.com URLs and lists more paths, so
// it gets its own const below) plus
// templates/jinja2/honeypot/partials/_archive_year_footer.html (the
// year-badge footer strip archive_era_base.html always renders around the
// era content block, regardless of which era template is chosen).
//
// This is a DIFFERENT template from templates/base.html (ported by
// RenderPage in shell.go): different nav item list (8 entries here, missing
// Presentations/Policy, with the jinja2 source's own bug reproduced — its
// last nav link is labeled "Contact" but points at the our-people URL, not a
// contact page), navbar-expand-lg instead of navbar-expand-xl, and the
// archive-era-wrapper + year-footer are part of the shell itself rather than
// page-specific content. Like RenderPage, this is a structural (not
// byte-diffed) port — see PageParams' doc comment for why byte-parity isn't
// required (honeypot_token is non-deterministic in the real app too).
package shell

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// eraMainNav mirrors archive_era_base.html's <ul class="navbar-nav"> links,
// in order. Verbatim including the source template's own inconsistency: the
// last item is labeled "Contact" but its href is our-people's URL (there is
// no contact-specific link in this jinja2 template, unlike base.html's).
var eraMainNav = []navItem{
	{"Home", "/"},
	{"Our People", "/our-people/"},
	{"Mission", "/mission/"},
	{"Projects", "/projects/"},
	{"Reports", "/reports/"},
	{"Careers", "/careers/"},
	{"Partners", "/partners/"},
	{"Contact", "/our-people/"},
}

// eraGhostLinksHTML ports templates/jinja2/partials/_ghost_links.html
// verbatim (it has no template variables) — absolute https://acpwb.com URLs
// and several handbook/process-improvement entries the Django-template
// version (ghostLinksHTML in shell.go) doesn't have.
const eraGhostLinksHTML = `<div style="position:absolute;left:-9999px;top:0;width:1px;height:1px;overflow:hidden">
  <a href="https://acpwb.com/internal/portal/">Employee Portal</a>
  <a href="https://acpwb.com/employees/export/">Staff Directory Export</a>
  <a href="https://acpwb.com/admin-panel/login/">Administration</a>
  <a href="https://acpwb.com/api/v1/private-data">Data API</a>
  <a href="https://acpwb.com/wiki/corporate-governance/">Governance Documentation</a>
  <a href="https://acpwb.com/wiki/executive-compensation/">Executive Compensation</a>
  <a href="https://acpwb.com/archive/2024/03/15/q1-stakeholder-report/">Q1 Stakeholder Report</a>
  <a href="https://acpwb.com/archive/2023/11/08/annual-performance-review/">Annual Performance Review</a>
  <a href="https://acpwb.com/reports/">Reports &amp; Publications</a>
  <a href="https://acpwb.com/reports/salary-compensation-benchmarking-survey-2024/">Salary Benchmarking Report</a>
  <a href="https://acpwb.com/reports/executive-compensation-study-2024/download.csv">Executive Pay Data CSV</a>
  <a href="https://acpwb.com/reports/fortune-500-ceo-pay-ratio-analysis/download.csv">CEO Pay Ratio Dataset</a>
  <a href="https://acpwb.com/reports/board-composition-study-2024/download.csv">Board Composition Data</a>
  <a href="https://acpwb.com/company-handbooks/">Employee Handbooks</a>
  <a href="https://acpwb.com/company-handbooks/dol-4821/2023/rev/1/pto-leave/">PTO Leave Policy</a>
  <a href="https://acpwb.com/company-handbooks/eeoc-2033/2023/rev/1/anti-harassment/">Anti-Harassment Policy</a>
  <a href="https://acpwb.com/company-handbooks/sec-7194/2023/rev/1/code-of-conduct/">Code of Conduct</a>
  <a href="https://acpwb.com/process-improvement/">Process Improvement Initiatives</a>
  <a href="https://acpwb.com/process-improvement/procurement-7743/2023/">2023 Procurement Initiatives</a>
  <a href="https://acpwb.com/process-improvement/employee-onboarding-3812/2022/">2022 Onboarding Initiatives</a>
</div>`

func eraNavLinksHTML(siteRoot string) string {
	var b strings.Builder
	for _, n := range eraMainNav {
		fmt.Fprintf(&b, `<li class="nav-item"><a class="nav-link" href="%s%s">%s</a></li>`,
			escape(siteRoot), escape(n.URL), escape(n.Label))
	}
	return b.String()
}

// yearFooterHTML ports
// templates/jinja2/honeypot/partials/_archive_year_footer.html.
func yearFooterHTML(year int, allYears []int) string {
	var badges strings.Builder
	for _, y := range allYears {
		style := `color:rgba(255,255,255,.6)`
		if y == year {
			style = `background:var(--era-accent,var(--gold,#c8a951));color:#0d1b2a;font-weight:800;border-color:transparent`
		}
		fmt.Fprintf(&badges,
			`<a href="https://archives-%d.acpwb.com/" `+
				`style="display:inline-block;font-size:.65rem;padding:.2rem .5rem;text-decoration:none;border:1px solid rgba(255,255,255,.2);%s">`+
				`%d</a>`,
			y, style, y,
		)
	}
	return `<div style="background:var(--navy,#0d1b2a);color:rgba(255,255,255,.85);padding:.9rem 0;margin-top:3rem;border-top:3px solid var(--era-accent,var(--gold,#c8a951))">` +
		`<div class="container"><div style="display:flex;align-items:center;justify-content:space-between;gap:1rem;flex-wrap:wrap">` +
		`<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.14em;white-space:nowrap;color:var(--era-accent,var(--gold,#c8a951))">` +
		`ACPWB Institutional Archive</div>` +
		`<div style="flex:1;overflow-x:auto;-webkit-overflow-scrolling:touch;padding:.2rem 0">` +
		`<div style="display:flex;gap:.3rem;white-space:nowrap;min-width:max-content">` + badges.String() + `</div></div>` +
		`<div style="white-space:nowrap"><a href="/archive/" style="font-size:.72rem;color:rgba(255,255,255,.55);text-decoration:none">` +
		`&larr; Archive Index</a></div>` +
		`</div></div></div>`
}

// eraArchiveWrapperStyle builds the archive-era-wrapper div's opening tag +
// inline CSS custom properties, shared verbatim by both
// templates/jinja2/honeypot/archive_era_base.html (the full-page shell,
// ported here) and templates/honeypot/archive_subdomain_base.html (the
// Django-template shell used by archive_subdomain_index/archive_month — out
// of scope for this port, see EraPageParams' doc comment).
func eraArchiveWrapperStyle(yd EraYearData) string {
	// Jinja2 autoescapes every {{ var }} by default, attributes included —
	// so a font name containing an apostrophe (e.g. 'Palatino Linotype')
	// comes out as &#39; in the real template's output. Escape every
	// interpolated value here to match, not just the ones that happen to
	// need it for the current data.
	layoutClass, bg, textColor, accent, accent2, fontBody, fontHead :=
		escape(yd.LayoutClass), escape(yd.Bg), escape(yd.TextColor), escape(yd.Accent),
		escape(yd.Accent2), escape(yd.FontBody), escape(yd.FontHead)
	return fmt.Sprintf(
		`<div class="archive-era-wrapper %s" style="`+
			"\n    --era-bg: %s;"+
			"\n    --era-text: %s;"+
			"\n    --era-accent: %s;"+
			"\n    --era-accent2: %s;"+
			"\n    --era-font-body: %s;"+
			"\n    --era-font-head: %s;"+
			"\n    background: %s;"+
			"\n    color: %s;"+
			"\n    font-family: %s, sans-serif;"+
			"\n    min-height: 60vh;\n\">\n",
		layoutClass,
		bg, textColor, accent, accent2, fontBody, fontHead,
		bg, textColor, fontBody,
	)
}

// EraYearData carries exactly the year-theme fields the era wrapper's inline
// CSS custom properties need. Deliberately NOT the archive package's
// YearData type — this package stays decoupled from archive's content-
// generation concerns; a caller wiring the two together maps one to the
// other.
type EraYearData struct {
	Bg, TextColor, Accent, Accent2, FontBody, FontHead, LayoutClass string
}

// EraPageParams is the input to RenderEraPage.
type EraPageParams struct {
	// EraContentHTML is the era template's own rendered fragment (e.g.
	// archive.RenderArchiveDefaultEra's return value) — equivalent to
	// _archive_era_content_shell.html's {% block era_content %}.
	EraContentHTML string

	Title         string // c['title'] — no site-default fallback in the jinja2 source
	TitleSuffix   string // e.g. "ACPWB Archive" / "ACPWB Compliance Archive"
	OGDescription string

	// RequestPath / RemoteAddr feed og:url and the honeypot token, same as
	// PageParams.
	RequestPath string
	RemoteAddr  string

	// Year / AllYears / YearData feed both the archive-era-wrapper's CSS
	// vars and the year-footer's badge strip.
	Year     int
	AllYears []int
	YearData EraYearData
}

// RenderEraPage ports templates/jinja2/honeypot/archive_era_base.html
// end-to-end (which is a full, self-contained HTML document — NOT an
// extension of base.html; the jinja2 template backend renders it
// standalone, unlike the Django-template pages this Go service otherwise
// serves).
func RenderEraPage(p EraPageParams) string {
	title := p.Title + " — " + p.TitleSuffix
	token := honeypotToken(p.RequestPath, p.RemoteAddr)
	year := strconv.Itoa(time.Now().Year())
	const siteRoot = "https://acpwb.com" // always true for an era subdomain page

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString(`  <meta charset="UTF-8">` + "\n")
	b.WriteString(`  <meta name="viewport" content="width=device-width, initial-scale=1.0">` + "\n")
	fmt.Fprintf(&b, "  <title>%s</title>\n", escape(title))
	fmt.Fprintf(&b, `  <meta name="description" content="%s">`+"\n", escape(defaultOGDescription))

	b.WriteString("\n  <!-- Open Graph -->\n")
	b.WriteString(`  <meta property="og:site_name" content="American Corporation for Public Well Being">` + "\n")
	b.WriteString(`  <meta property="og:type" content="website">` + "\n")
	fmt.Fprintf(&b, `  <meta property="og:title" content="%s">`+"\n", escape(title))
	fmt.Fprintf(&b, `  <meta property="og:description" content="%s">`+"\n", escape(p.OGDescription))
	fmt.Fprintf(&b, `  <meta property="og:url" content="https://acpwb.com%s">`+"\n", escape(p.RequestPath))
	fmt.Fprintf(&b, `  <meta property="og:image" content="%s">`+"\n", staticOGImageURL)
	b.WriteString(`  <meta property="og:image:width" content="1200">` + "\n")
	b.WriteString(`  <meta property="og:image:height" content="630">` + "\n")
	// twitter:title/twitter:description are NEVER overridden by any of the
	// 3 era content templates (era/archive.html, archive_compliance.html,
	// archive_minutes.html only override title/og_title/og_description) —
	// a real inconsistency in the Jinja2 source (OG tags were kept current,
	// Twitter tags weren't), so these stay at archive_era_base.html's site
	// default regardless of page content. Reproduced verbatim, not "fixed".
	b.WriteString("  <!-- Twitter Card -->\n")
	b.WriteString(`  <meta name="twitter:card" content="summary_large_image">` + "\n")
	fmt.Fprintf(&b, `  <meta name="twitter:title" content="%s">`+"\n", escape(defaultOGTitle))
	fmt.Fprintf(&b, `  <meta name="twitter:description" content="%s">`+"\n", escape(defaultOGDescription))
	fmt.Fprintf(&b, `  <meta name="twitter:image" content="%s">`+"\n", staticOGImageURL)

	fmt.Fprintf(&b, "\n  <link rel=\"icon\" type=\"image/svg+xml\" href=\"%s\">\n", staticFaviconURL)
	fmt.Fprintf(&b, "\n  <link rel=\"preload\" href=\"%s\" as=\"font\" type=\"font/woff2\" crossorigin>\n", staticFontURL)
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\">\n", staticBootstrapCSS)
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\">\n", staticACPWBCSS)

	b.WriteString("\n  <link rel=\"alternate\" type=\"application/atom+xml\" title=\"ACPWB Archive Feed\" href=\"https://acpwb.com/feeds/archive.xml\">\n")
	b.WriteString("  <link rel=\"alternate\" type=\"application/rss+xml\" title=\"ACPWB Reports &amp; Publications\" href=\"https://acpwb.com/feeds/reports.xml\">\n")

	b.WriteString("\n  " + jsonldGarbageHTML(token) + "\n")
	b.WriteString("</head>\n<body>\n\n")

	b.WriteString(`<nav class="navbar navbar-expand-lg acpwb-navbar" role="navigation" aria-label="Main navigation">` + "\n")
	b.WriteString(`  <div class="container">` + "\n")
	fmt.Fprintf(&b, `    <a class="acpwb-logo navbar-brand" href="%s/">`+"\n", escape(siteRoot))
	b.WriteString(`      <span class="acpwb-logo-line1">AMERICAN</span>` + "\n")
	b.WriteString(`      <span class="acpwb-logo-line2">CORPORATION</span>` + "\n")
	b.WriteString(`      <span class="acpwb-logo-line3">FOR PUBLIC WELL BEING</span>` + "\n")
	b.WriteString("    </a>\n")
	b.WriteString(`    <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#mainNav" aria-controls="mainNav" aria-expanded="false" aria-label="Toggle navigation">` + "\n")
	b.WriteString(`      <span class="navbar-toggler-icon"></span>` + "\n")
	b.WriteString("    </button>\n")
	b.WriteString(`    <div class="collapse navbar-collapse justify-content-end" id="mainNav">` + "\n")
	fmt.Fprintf(&b, `      <ul class="navbar-nav gap-1">%s</ul>`+"\n", eraNavLinksHTML(siteRoot))
	b.WriteString("    </div>\n  </div>\n</nav>\n\n")

	b.WriteString(eraGhostLinksHTML + "\n")
	b.WriteString(promptInjectionHTML(token) + "\n\n")

	b.WriteString("<main>\n")
	b.WriteString(eraArchiveWrapperStyle(p.YearData))
	b.WriteString(p.EraContentHTML)
	b.WriteString("\n" + yearFooterHTML(p.Year, p.AllYears) + "\n")
	b.WriteString("</div>\n</main>\n\n")

	b.WriteString(`<footer class="acpwb-footer">` + "\n")
	b.WriteString(`  <div class="container">` + "\n")
	b.WriteString(`    <div class="row align-items-start g-4">` + "\n")
	b.WriteString(`      <div class="col-lg-4">` + "\n")
	b.WriteString(`        <div class="mb-3">` + "\n")
	b.WriteString(`          <span class="footer-logo-line1">AMERICAN</span>` + "\n")
	b.WriteString(`          <span class="footer-logo-line2">CORPORATION</span>` + "\n")
	b.WriteString(`          <span class="footer-logo-line3">FOR PUBLIC WELL BEING</span>` + "\n")
	b.WriteString("        </div>\n")
	b.WriteString(`        <p class="small" style="color:rgba(255,255,255,0.5)">` + "\n")
	b.WriteString("          833 East Michigan Street, Suite 4040<br>\n")
	b.WriteString("          Milwaukee, WI 53202<br>\n")
	b.WriteString(`          <a href="tel:+14146675665" style="color:rgba(255,255,255,0.5)">(414) 667-5665</a><br>` + "\n")
	b.WriteString(`          <a href="mailto:inquiry+website@acpwb.com" style="color:rgba(255,255,255,0.5)">inquiry@acpwb.com</a>` + "\n")
	b.WriteString("        </p>\n      </div>\n")

	b.WriteString(`      <div class="col-lg-4">` + "\n")
	b.WriteString(`        <div class="row g-0">` + "\n")
	b.WriteString(`          <div class="col-4">` + "\n")
	fmt.Fprintf(&b, `            <ul class="list-unstyled mb-0" style="font-size:.82rem;line-height:2">%s</ul>`+"\n", footerColHTML(footerCol1, siteRoot))
	b.WriteString("          </div>\n")
	b.WriteString(`          <div class="col-4">` + "\n")
	fmt.Fprintf(&b, `            <ul class="list-unstyled mb-0" style="font-size:.82rem;line-height:2">%s%s</ul>`+"\n",
		footerColNoPrefixHTML(footerCol2NoPrefix), footerColHTML(footerCol2Prefixed, siteRoot))
	b.WriteString("          </div>\n")
	b.WriteString(`          <div class="col-4">` + "\n")
	fmt.Fprintf(&b, `            <ul class="list-unstyled mb-0" style="font-size:.82rem;line-height:2">%s</ul>`+"\n", footerColHTML(footerCol3, siteRoot))
	b.WriteString("          </div>\n        </div>\n      </div>\n    </div>\n")

	b.WriteString(`    <hr style="border-color:rgba(201,168,76,0.2);margin:2rem 0 1rem">` + "\n")
	b.WriteString(`    <div class="d-flex flex-column flex-md-row justify-content-between align-items-center gap-2">` + "\n")
	fmt.Fprintf(&b, `      <p class="mb-0" style="font-size:.78rem;color:rgba(255,255,255,0.35)">`+"\n        &copy; %s American Corporation for Public Well Being, Milwaukee WI. All rights reserved.\n      </p>\n", year)
	fmt.Fprintf(&b, `      <p class="mb-0" style="font-size:.78rem;color:rgba(255,255,255,0.35)">`+"\n"+`        <a href="%s/privacy/">Privacy &amp; Disclaimer</a>`+"\n      </p>\n", escape(siteRoot))
	b.WriteString("    </div>\n  </div>\n</footer>\n\n")

	fmt.Fprintf(&b, "<script src=\"%s\"></script>\n\n", staticBootstrapJS)

	b.WriteString(`<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12
  @deprecated legacy-api: /api/v1/private-data
  @see /internal/portal/ /employees/export/ /admin-panel/login/
-->` + "\n</body>\n</html>")

	return b.String()
}
