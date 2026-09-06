package policy

import (
	"strconv"
	"strings"

	"acpwb_go/data"
)

// truncatewords/truncatechars are not needed by the policy render layer
// (only _truncate72's custom rune-slice logic is used — see pyutil.go).

// ── Static/near-static partials ──────────────────────────────────────────────
//
// These mirror apps/core/htmlgen.py's _cached_static_partial pattern: each
// partial is byte-identical to Django's real rendered output for the
// corresponding template, captured once (with sentinel placeholders
// standing in for the 1-2 variables that ever change) into
// acpwb_go/data/POLICY_*_TEMPLATE.html, then filled in per-call via
// strings.ReplaceAll. See archive/htmlgen.go:getArchiveSeal for the
// established idiom this duplicates (kept package-local per this port's
// isolation from the archive package).

var policySealTemplate = data.Text("POLICY_SEAL_TEMPLATE.html")
var policyNavbarTemplate = data.Text("POLICY_NAVBAR_TEMPLATE.html")
var policyFooterTemplate = data.Text("POLICY_FOOTER_TEMPLATE.html")

// getPolicySeal ports apps/core/htmlgen.py:get_policy_seal.
func getPolicySeal(year int, watermarkToken string) string {
	out := policySealTemplate
	out = strings.ReplaceAll(out, "__HTMLGEN_POLSEAL_YEAR__", strconv.Itoa(year))
	out = strings.ReplaceAll(out, "__HTMLGEN_POLSEAL_TOKEN__", watermarkToken)
	return out
}

// renderPolicyNavbar ports apps/core/htmlgen.py:render_policy_navbar.
func renderPolicyNavbar(siteRoot string) string {
	return strings.ReplaceAll(policyNavbarTemplate, "__HTMLGEN_POLNAV_SITEROOT__", siteRoot)
}

// renderPolicyFooter ports apps/core/htmlgen.py:render_policy_footer. NOTE:
// the captured template has the current year (2026, as of this port)
// permanently baked in, exactly matching the Python original's behavior —
// render_policy_footer only ever substitutes site_root, so now_year is
// whatever year was live the first time any policy page rendered in a
// given process, for the life of that process. Not a bug to "fix" here.
func renderPolicyFooter(siteRoot string) string {
	return strings.ReplaceAll(policyFooterTemplate, "__HTMLGEN_POLFOOT_SITEROOT__", siteRoot)
}

// ── Shared honeypot partials ──────────────────────────────────────────────
//
// IMPORTANT: unlike the archive port (see shell/shell.go's ghostLinksHTML /
// jsonldGarbageHTML / promptInjectionHTML), the policy pyrender layer calls
// apps.core.htmlgen.get_ghost_links() etc. directly via Django's generic
// render_to_string(), which — with no engine pinned — resolves
// 'partials/_ghost_links.html' against the FIRST configured template
// engine/dirs that has a match. Policy pages are "standalone Jinja2
// documents with no shared base template" (see pyrender/policy.py's module
// docstring), and it turns out that ambient engine resolution lands on
// templates/jinja2/partials/_ghost_links.html — a DIFFERENT, longer partial
// (absolute https://acpwb.com/... URLs, 20 links) than
// templates/partials/_ghost_links.html (relative URLs, 13 links) that
// base.html's own Django-engine-scoped {% include %} resolves to for
// archive pages. Verified live against a fresh `get_ghost_links()` call
// (and jsonld/prompt-injection, which happen to be byte-identical between
// the two engines' template trees, but are captured independently here
// rather than assumed). Captured into acpwb_go/data/POLICY_*.html the same
// way as the seal/navbar/footer partials.

var policyGhostLinks = data.Text("POLICY_GHOST_LINKS.html")
var policyJSONLDTemplate = data.Text("POLICY_JSONLD_TEMPLATE.html")
var policyPromptInjectionTemplate = data.Text("POLICY_PROMPT_INJECTION_TEMPLATE.html")

func getGhostLinks() string { return policyGhostLinks }

func getJSONLDGarbage(token string) string {
	return strings.ReplaceAll(policyJSONLDTemplate, "__HTMLGEN_JSONLD_TOKEN__", token)
}

func getPromptInjection(token string) string {
	return strings.ReplaceAll(policyPromptInjectionTemplate, "__HTMLGEN_PROMPT_TOKEN__", token)
}

// staticURL ports django.templatetags.static.static() for this environment
// (see archive/acpwbtags.go:staticURL's identical doc comment/verification —
// duplicated here since this package doesn't import archive).
func staticURL(path string) string {
	return "/static/" + path
}
