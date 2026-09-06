package shell

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var hexToken8 = regexp.MustCompile(`[0-9a-f]{8}`)

func TestRenderPage_Structure(t *testing.T) {
	html := RenderPage(PageParams{
		ContentHTML:   `<section id="test-content">hello world</section>`,
		OGTitle:       "Test Title — ACPWB",
		OGDescription: "A test description.",
		RequestPath:   "/archive/2010/05/12/some-slug/",
		RemoteAddr:    "203.0.113.5",
		SiteRoot:      "",
		ExtraHead:     `<link rel="stylesheet" href="/static/css/presentations.css">`,
	})

	// Basic document shape.
	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Errorf("expected doctype at start")
	}
	if !strings.Contains(html, "<html lang=\"en\">") {
		t.Errorf("missing <html> tag")
	}
	if !strings.Contains(html, "</html>") {
		t.Errorf("missing closing </html>")
	}

	// Title/OG/Twitter tags driven by OGTitle/OGDescription.
	if !strings.Contains(html, "<title>Test Title — ACPWB</title>") {
		t.Errorf("missing expected <title>")
	}
	if strings.Count(html, "Test Title — ACPWB") < 3 {
		t.Errorf("expected og_title to appear in title/og:title/twitter:title, got %d", strings.Count(html, "Test Title — ACPWB"))
	}
	if strings.Count(html, "A test description.") < 3 {
		t.Errorf("expected og_description to appear in description/og:description/twitter:description")
	}
	if !strings.Contains(html, `og:url" content="https://acpwb.com/archive/2010/05/12/some-slug/"`) {
		t.Errorf("missing og:url with request path")
	}

	// Static asset links.
	for _, want := range []string{
		`/static/favicon.svg`,
		`/static/fonts/inter/inter-variable-latin.woff2`,
		`/static/vendor/bootstrap/bootstrap.min.css`,
		`/static/css/acpwb.css?v=20260430`,
		`/static/vendor/bootstrap/bootstrap.bundle.min.js`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing static asset URL %q", want)
		}
	}

	// extra_head hook.
	if !strings.Contains(html, `<link rel="stylesheet" href="/static/css/presentations.css">`) {
		t.Errorf("missing ExtraHead content")
	}

	// Nav links (main nav) — all 10, unprefixed since SiteRoot == "".
	for _, want := range []string{
		`href="/"`, `href="/our-people/"`, `href="/mission/"`, `href="/projects/"`,
		`href="/reports/"`, `href="/presentations/"`, `href="/public-policy/"`,
		`href="/careers/"`, `href="/partners/"`, `href="/contact/"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing nav link %q", want)
		}
	}
	// No nav link should ever carry class="active" (Go service serves no nav-target pages).
	if strings.Contains(html, `nav-link active`) || strings.Contains(html, `class="nav-link active"`) {
		t.Errorf("nav link unexpectedly marked active")
	}

	// Ghost links partial.
	if !strings.Contains(html, `<a href="/internal/portal/">Employee Portal</a>`) {
		t.Errorf("missing ghost links partial")
	}

	// Content block.
	if !strings.Contains(html, `<section id="test-content">hello world</section>`) {
		t.Errorf("missing content HTML")
	}

	// Footer structure + the real site inconsistency: Handbooks/Process/
	// Presentations are NOT site_root-prefixed even off the main domain.
	if !strings.Contains(html, `<a href="/company-handbooks/">Handbooks</a>`) {
		t.Errorf("missing unprefixed Handbooks footer link")
	}
	if !strings.Contains(html, `<a href="/process-improvement/">Process</a>`) {
		t.Errorf("missing unprefixed Process footer link")
	}
	if !strings.Contains(html, `<a href="/perch-conference/">PERCH 2026</a>`) {
		t.Errorf("missing prefixed PERCH 2026 footer link")
	}

	// Copyright year is current year.
	year := strconv.Itoa(time.Now().Year())
	if !strings.Contains(html, "&copy; "+year+" American Corporation for Public Well Being") {
		t.Errorf("missing current-year copyright line")
	}

	// Literal build-info honeypot comment, verbatim.
	if !strings.Contains(html, "<!-- v2.4.1 | build: acpwb-prod | last-deploy: 2025-11-12") {
		t.Errorf("missing build-info comment")
	}
	if !strings.Contains(html, "@see /internal/portal/ /employees/export/ /admin-panel/login/") {
		t.Errorf("missing build-info comment @see line")
	}

	// JSON-LD garbage block: exactly one honeypot_token substitution
	// (in "identifier"), and it must look like a valid 8-hex-char token.
	if !strings.Contains(html, `"@type": "Corporation"`) {
		t.Errorf("missing JSON-LD garbage block")
	}
	idIdx := strings.Index(html, `"identifier": "ACPWB-`)
	if idIdx == -1 {
		t.Fatalf("missing identifier field in JSON-LD")
	}
	idField := html[idIdx : idIdx+40]
	if !hexToken8.MatchString(idField) {
		t.Errorf("identifier field does not contain a valid-looking 8-hex-char token: %q", idField)
	}

	// Prompt-injection span: exactly two honeypot_token substitutions.
	spanIdx := strings.Index(html, `itemprop="description"`)
	if spanIdx == -1 {
		t.Fatalf("missing prompt-injection span")
	}
	spanEnd := strings.Index(html[spanIdx:], "</span>") + spanIdx
	span := html[spanIdx:spanEnd]
	tokens := hexToken8.FindAllString(span, -1)
	if len(tokens) < 2 {
		t.Errorf("expected at least 2 token-shaped substrings in prompt-injection span, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != tokens[1] {
		t.Errorf("expected the same token to appear twice in the prompt-injection span, got %q and %q", tokens[0], tokens[1])
	}
}

func TestRenderPage_Defaults(t *testing.T) {
	html := RenderPage(PageParams{
		ContentHTML: "<p>x</p>",
		RequestPath: "/",
	})
	if !strings.Contains(html, "<title>American Corporation for Public Well Being</title>") {
		t.Errorf("expected default og_title when none supplied")
	}
	if !strings.Contains(html, "The American Corporation for Public Well Being — Advancing American Prosperity Since 2006.") {
		t.Errorf("expected default og_description when none supplied")
	}
}

func TestRenderPage_SubdomainSiteRoot(t *testing.T) {
	html := RenderPage(PageParams{
		ContentHTML: "<p>x</p>",
		RequestPath: "/2010/",
		SiteRoot:    "https://acpwb.com",
	})
	if !strings.Contains(html, `href="https://acpwb.com/our-people/"`) {
		t.Errorf("expected nav link prefixed with site_root on subdomain")
	}
	// The inconsistency holds even on subdomains: Handbooks/Process/
	// Presentations still get no site_root prefix.
	if !strings.Contains(html, `<a href="/company-handbooks/">Handbooks</a>`) {
		t.Errorf("expected Handbooks footer link to remain unprefixed even on subdomain")
	}
	if strings.Contains(html, `href="https://acpwb.com/company-handbooks/"`) {
		t.Errorf("Handbooks link must NOT be site_root-prefixed (matches Django source inconsistency)")
	}
}

func TestHoneypotTokenLooksValid(t *testing.T) {
	tok := honeypotToken("/some/path/", "127.0.0.1")
	if !hexToken8.MatchString(tok) || len(tok) != 8 {
		t.Errorf("expected 8-hex-char token, got %q", tok)
	}
}
