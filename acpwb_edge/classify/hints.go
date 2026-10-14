package classify

import (
	"regexp"
	"strconv"
	"strings"
)

// ClientHints are a request's low-entropy User-Agent Client Hints, which
// every Chromium browser since 89 sends on HTTPS requests.
type ClientHints struct {
	UA       string // sec-ch-ua: "Chromium";v="155", "Not/A)Brand";v="24", "Google Chrome";v="155"
	Mobile   string // sec-ch-ua-mobile: ?0 / ?1
	Platform string // sec-ch-ua-platform: "Windows"
}

var brandRe = regexp.MustCompile(`"([^"]*)"\s*;\s*v="(\d+)`)

// hintsMismatch reports whether a Chromium claim's client hints contradict
// its User-Agent. Chrome builds both from the same version and platform, so
// they always agree; a UA swapped by automation (Puppeteer's setUserAgent,
// a header-rewriting proxy) or written by an HTTP library doesn't. headless
// is set when the hints admit to HeadlessChrome.
func hintsMismatch(ua string, ch ClientHints) (mismatch, headless bool) {
	l := strings.ToLower(ua)
	var chromium int
	for _, m := range brandRe.FindAllStringSubmatch(ch.UA, -1) {
		switch m[1] {
		case "HeadlessChrome":
			headless = true
		case "Chromium":
			chromium, _ = strconv.Atoi(m[2])
		}
	}
	if headless {
		return false, true
	}
	if want := major(l, "chrome/"); chromium == 0 || want != 0 && chromium != want {
		return true, false
	}
	// Phones say "Mobile" in the UA; Android tablets don't, and send ?0.
	if mobile := strings.Contains(l, "mobile"); ch.Mobile == "?1" && !mobile || ch.Mobile == "?0" && mobile {
		return true, false
	}
	if p := uaPlatform(l); p != "" && ch.Platform != "" && strings.Trim(ch.Platform, `"`) != p {
		return true, false
	}
	return false, false
}

// uaPlatform is the sec-ch-ua-platform Chrome sends alongside a UA, or ""
// when the UA doesn't pin it down.
func uaPlatform(l string) string {
	switch {
	case strings.Contains(l, "windows nt"):
		return "Windows"
	case strings.Contains(l, "android"):
		return "Android"
	case strings.Contains(l, "cros "):
		return "Chrome OS"
	case strings.Contains(l, "macintosh"):
		return "macOS"
	}
	// Not Linux: Android's "desktop site" mode sends a Linux UA.
	return ""
}
