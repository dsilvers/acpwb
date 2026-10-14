package classify

import (
	"regexp"
	"slices"
	"strings"
)

// Chromium's request header orders (HTTP/2, pseudo-headers aside), as
// Chrome 155 sends them through the edge: a page navigation and a
// subresource / fetch(). Headers outside these lists (origin, content-*,
// cache validators, custom ones) don't take part in the check.
var (
	chromiumNavigation = []string{
		"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent",
		"accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "referer",
		"accept-encoding", "accept-language", "cookie", "priority",
	}
	chromiumSubresource = []string{
		"sec-ch-ua-platform", "user-agent", "sec-ch-ua", "sec-ch-ua-mobile", "accept", "sec-fetch-site",
		"sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language", "cookie", "priority",
	}
	chromiumChecked = func() map[string]bool {
		m := map[string]bool{}
		for _, n := range append(slices.Clone(chromiumNavigation), chromiumSubresource...) {
			m[n] = true
		}
		return m
	}()
)

// chromiumHeaderOrderOK reports whether the request's headers come in an
// order Chromium sends. Automation that sets headers on a real Chrome
// (Puppeteer's setExtraHTTPHeaders and the like) and Chrome-impersonating
// HTTP libraries put accept-language early; Chrome never does.
func chromiumHeaderOrderOK(names []string) bool {
	var seq []string
	for _, n := range names {
		n = strings.ToLower(n)
		if !chromiumChecked[n] || (n == "cookie" && slices.Contains(seq, "cookie")) {
			continue // HTTP/2 splits cookies into several headers
		}
		seq = append(seq, n)
	}
	return inOrder(seq, chromiumNavigation) || inOrder(seq, chromiumSubresource)
}

// inOrder reports whether every name in seq is in tpl, in tpl's order.
func inOrder(seq, tpl []string) bool {
	pos := -1
	for _, n := range seq {
		p := slices.Index(tpl, n)
		if p <= pos {
			return false
		}
		pos = p
	}
	return true
}

// hiddenLink reports whether path is under one of the trap sections that
// only invisible links point to (templates/partials/_ghost_links.html and
// the "@see" HTML comments; Django logs them as ghost_link, see
// BotTrackingMiddleware._classify_path).
func hiddenLink(path string) bool {
	for _, p := range [...]string{"/internal/", "/employees/", "/admin-panel/"} {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func hasHeader(names []string, want string) bool {
	for _, n := range names {
		if strings.EqualFold(n, want) {
			return true
		}
	}
	return false
}

var (
	// "(compatible; KeenableBot/1.0; +https://…)", "…); compatible; ShapBot/0.1.0"
	compatibleRe = regexp.MustCompile(`(?i)compatible;\s*([^;()]+?)\s*(?:[;)]|$)`)
	botWordRe    = regexp.MustCompile(`(?i)bot|crawl|spider|scan|fetch|scrap|archiv|monitor|preview`)
)

// declaredClient recognises clients that say what they are without being
// in botclassify's list: crawlers that name themselves the conventional way
// ("compatible; NameBot/1.0; +https://…"), and anything that doesn't even
// pretend to be a browser (every browser UA starts with "Mozilla/").
// bot is false for an HTTP tool or library ("Cpanel-HTTP-Client/1.0").
func declaredClient(ua string) (name string, bot, ok bool) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "", false, false
	}
	if !strings.HasPrefix(strings.ToLower(ua), "mozilla/") {
		return productName(ua), botWordRe.MatchString(ua), true
	}
	for _, m := range compatibleRe.FindAllStringSubmatch(ua, -1) {
		n := productName(m[1])
		if n == "" || strings.HasPrefix(strings.ToUpper(n), "MSIE") || strings.HasPrefix(n, "+") {
			continue // "compatible; MSIE 10.0" is a (fake) browser claim
		}
		return n, true, true
	}
	if strings.Contains(ua, "+http") || botWordRe.MatchString(ua) && claimBrowser(ua).Name == "" {
		return "Unnamed bot", true, true
	}
	return "", false, false
}

// productName is a UA's leading product name: up to the version slash, a
// parenthesis or a comma, at most 40 bytes.
func productName(s string) string {
	if i := strings.IndexAny(s, "/(,;"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	return clipName(s, 40)
}

func clipName(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && s[n]&0xC0 == 0x80 {
		n--
	}
	return s[:n]
}
