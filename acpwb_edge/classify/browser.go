package classify

import (
	"strconv"
	"strings"
)

// Claim is the browser a User-Agent says it is, and the TLS stack that
// browser would really use.
type Claim struct {
	Name   string // "Chrome", "Edge", "Opera", "Firefox", "Safari"; "" if no browser is claimed
	Family string // chromium, firefox, apple
	Major  int    // major version, 0 if unknown
}

// claimBrowser reads a UA the way the live view's _BROWSER_FAMILIES does
// (Edge and Opera before Chrome, Chrome before Safari), plus one thing it
// doesn't: every iOS browser is WebKit on Apple's TLS stack, whatever it
// calls itself.
func claimBrowser(ua string) Claim {
	l := strings.ToLower(ua)
	if !strings.Contains(l, "mozilla/") {
		return Claim{}
	}
	ios := strings.Contains(l, "iphone") || strings.Contains(l, "ipad") || strings.Contains(l, "ipod")
	var c Claim
	switch {
	case strings.Contains(l, "edg/"), strings.Contains(l, "edga/"):
		c = Claim{Name: "Edge", Family: "chromium", Major: major(l, "edg/")}
	case strings.Contains(l, "edgios/"):
		c = Claim{Name: "Edge", Family: "apple"}
	case strings.Contains(l, "opr/"):
		c = Claim{Name: "Opera", Family: "chromium", Major: major(l, "chrome/")}
	case strings.Contains(l, "crios/"):
		c = Claim{Name: "Chrome", Family: "apple", Major: major(l, "crios/")}
	case strings.Contains(l, "fxios/"):
		c = Claim{Name: "Firefox", Family: "apple", Major: major(l, "fxios/")}
	case strings.Contains(l, "firefox/"):
		c = Claim{Name: "Firefox", Family: "firefox", Major: major(l, "firefox/")}
	case strings.Contains(l, "chrome/"):
		c = Claim{Name: "Chrome", Family: "chromium", Major: major(l, "chrome/")}
	case strings.Contains(l, "safari/") && strings.Contains(l, "version/"):
		c = Claim{Name: "Safari", Family: "apple", Major: major(l, "version/")}
	case strings.Contains(l, "safari/") && strings.Contains(l, "applewebkit/"):
		c = Claim{Name: "Safari", Family: "apple"} // Safari 2 and older had no Version/ token
	default:
		return Claim{}
	}
	if ios {
		c.Family = "apple"
	}
	return c
}

func major(l, token string) int {
	_, rest, ok := strings.Cut(l, token)
	if !ok {
		return 0
	}
	end := strings.IndexFunc(rest, func(r rune) bool { return r < '0' || r > '9' })
	if end < 0 {
		end = len(rest)
	}
	n, _ := strconv.Atoi(rest[:end])
	return n
}
