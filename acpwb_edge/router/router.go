// Package router is acpwb_edge's port of the nginx server blocks it
// replaced: nginx/acpwb.com (+ conf.d/upstream-acpwb.conf's host maps) and
// nginx/botseed.net, including nginx's location precedence — regex
// locations, in file order, beat prefix locations; otherwise the longest
// matching prefix wins. router_test.go pins every location.
package router

import (
	"net"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/klauspost/compress/gzhttp"

	"acpwb_edge/meta"
	"acpwb_edge/page"
)

// Route targets — also what EdgeConnection.first_route records.
const (
	Django       = "django"
	AcpwbGo      = "acpwb_go"
	WS           = "ws"
	Static       = "static"
	StaticGraphs = "static_graphs"
	BotseedWS    = "botseed_ws"
	BotseedAPI   = "botseed_api"
	BotseedFiles = "botseed_static"
	Redirect     = "redirect"
	ACME         = "acme"
)

type Config struct {
	Dev bool // also accept acpwb.example / localhost hosts

	StaticRoot  string // /static/ alias
	BotseedRoot string // botseed.net document root
	ACMERoot    string // certbot webroot for /.well-known/acme-challenge/ on :80

	// Upstream proxies, keyed by Django / AcpwbGo / WS / BotseedWS / BotseedAPI.
	Upstreams map[string]http.Handler
}

var (
	archiveHost = regexp.MustCompile(`^archives-\d{4}\.acpwb\.(com|example)$`)
	policyHost  = regexp.MustCompile(`^policy-[a-z0-9][a-z0-9\-]*\.acpwb\.(com|example)$`)
)

// regexLocation is an nginx `location ~ ...` block. target picks the
// upstream, given whether the host is an archive / policy subdomain.
type regexLocation struct {
	re     *regexp.Regexp
	target func(archive, policy bool) string
}

func always(t string) func(bool, bool) string { return func(bool, bool) string { return t } }

// In nginx/acpwb.com file order — order is load-bearing (see the export.csv
// comment there).
var regexLocations = []regexLocation{
	// Archive day-level trap pages: /archive/<year>/<month>/<day>[/<slug>][/]
	{regexp.MustCompile(`^/archive/\d{4}/\d{1,2}/\d{1,2}(/|$)`), always(AcpwbGo)},
	// CSV export isn't ported to acpwb_go; must be checked before the next one.
	{regexp.MustCompile(`^/\d{1,2}/\d{1,2}/.*export\.csv$`), always(Django)},
	// Archive-subdomain day-level content; elsewhere it's Django's
	// archive_trap_yearless.
	{regexp.MustCompile(`^/\d{1,2}/\d{1,2}(/|$)`), func(archive, _ bool) string {
		if archive {
			return AcpwbGo
		}
		return Django
	}},
	// Policy-subdomain year/month/detail pages.
	{regexp.MustCompile(`^/[0-9]{4}(/[0-9]{1,2}(/[0-9]{1,2}/[^/]+)?)?/?$`), func(_, policy bool) string {
		if policy {
			return AcpwbGo
		}
		return Django
	}},
}

// Prefix locations. /static/graphs/ comes from the docker nginx.conf: the
// dashboard's traffic graphs are regenerated every 30 minutes, so they
// mustn't get /static/'s 30-day immutable caching.
var prefixLocations = []struct{ prefix, target string }{
	{"/static/graphs/", StaticGraphs},
	{"/static/", Static},
	{"/archive/", Django},
	{"/source-code/", AcpwbGo},
	{"/compliance/", AcpwbGo},
	{"/public-policy/", AcpwbGo},
	{"/wiki/", Django},
	{"/reports/", Django},
	{"/.well-known/", Django},
	{"/ws/requests/", WS},
	{"/", Django},
}

// AcpwbTarget picks the location for an acpwb.com request, as nginx would.
func AcpwbTarget(host, p string) string {
	p = normalize(p)
	archive, policy := archiveHost.MatchString(host), policyHost.MatchString(host)
	for _, l := range regexLocations {
		if l.re.MatchString(p) {
			return l.target(archive, policy)
		}
	}
	best, target := -1, Django
	for _, l := range prefixLocations {
		if strings.HasPrefix(p, l.prefix) && len(l.prefix) > best {
			best, target = len(l.prefix), l.target
		}
	}
	return target
}

// BotseedTarget picks the location for a botseed.net request.
func BotseedTarget(p string) string {
	p = normalize(p)
	switch {
	case strings.HasPrefix(p, "/ws/"):
		return BotseedWS
	case strings.HasPrefix(p, "/api/"):
		return BotseedAPI
	}
	return BotseedFiles
}

// normalize does what nginx does to a URI before matching locations:
// merge slashes and resolve dot segments, keeping a trailing slash.
func normalize(p string) string {
	if p == "" {
		return "/"
	}
	trailing := strings.HasSuffix(p, "/")
	c := path.Clean("/" + p)
	if trailing && c != "/" {
		c += "/"
	}
	return c
}

// HostOf returns nginx's $host for a Host header value: lowercased, port
// and trailing dot stripped.
func HostOf(hostport string) string {
	h := strings.ToLower(hostport)
	if strings.HasPrefix(h, "[") { // IPv6 literal
		if i := strings.Index(h, "]"); i >= 0 {
			return h[:i+1]
		}
	}
	if i := strings.LastIndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	return strings.TrimSuffix(h, ".")
}

type vhost int

const (
	vhostUnknown vhost = iota
	vhostAcpwb
	vhostAcpwbWWW
	vhostBotseed
	vhostBotseedWWW
)

func (c *Config) classify(host string) vhost {
	switch {
	case host == "www.acpwb.com":
		return vhostAcpwbWWW
	case host == "acpwb.com" || strings.HasSuffix(host, ".acpwb.com"):
		return vhostAcpwb
	case host == "botseed.net":
		return vhostBotseed
	case host == "www.botseed.net":
		return vhostBotseedWWW
	case c.Dev && (host == "acpwb.example" || strings.HasSuffix(host, ".acpwb.example") ||
		host == "localhost" || host == "127.0.0.1" || host == "[::1]"):
		return vhostAcpwb
	}
	return vhostUnknown
}

var gzipTypes = []string{
	"text/html", "text/plain", "text/css", "text/xml", "text/javascript",
	"application/json", "application/javascript", "application/xml",
	"application/atom+xml", "application/rss+xml",
}

// gzipped wraps h with nginx's gzip settings (level 4, ≥1024 bytes, the
// production gzip_types plus text/html, which nginx always compresses).
func gzipped(h http.Handler) http.Handler {
	wrap, err := gzhttp.NewWrapper(
		gzhttp.MinSize(1024),
		gzhttp.CompressionLevel(4),
		gzhttp.ContentTypes(gzipTypes),
	)
	if err != nil {
		panic(err)
	}
	return wrap(h)
}

// TLS returns the handler for the HTTPS listener.
func TLS(c Config) http.Handler {
	targets := map[string]http.Handler{
		Static:       gzipped(files(c.StaticRoot, "/static/", false, staticCache)),
		StaticGraphs: gzipped(files(c.StaticRoot, "/static/", false, noCache)),
		BotseedFiles: gzipped(files(c.BotseedRoot, "/", true, nil)),
	}
	for _, name := range []string{Django, AcpwbGo, BotseedAPI} {
		if h := c.Upstreams[name]; h != nil {
			targets[name] = gzipped(h)
		}
	}
	// WebSockets aren't compressed: gzip wrappers and hijacked conns don't mix.
	for _, name := range []string{WS, BotseedWS} {
		if h := c.Upstreams[name]; h != nil {
			targets[name] = h
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := meta.From(r.Context())
		var target string
		switch c.classify(m.Host) {
		case vhostAcpwb:
			target = AcpwbTarget(m.Host, r.URL.Path)
		case vhostBotseed:
			target = BotseedTarget(r.URL.Path)
		case vhostBotseedWWW:
			m.Route = Redirect
			page.Redirect(w, r, "https://botseed.net"+r.URL.RequestURI(), http.StatusMovedPermanently)
			return
		default: // www.acpwb.com, and nginx's default server for unknown names
			m.Route = Redirect
			page.Redirect(w, r, "https://acpwb.com"+r.URL.RequestURI(), http.StatusMovedPermanently)
			return
		}
		m.Route = target
		h := targets[target]
		if h == nil {
			page.Error(w, http.StatusBadGateway)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Plain returns the handler for the port-80 listener: ACME HTTP-01
// challenges from the certbot webroot for every host (botseed.net renews
// that way), everything else redirected to HTTPS. With serveBotseed (local
// dev's :8090) it instead serves botseed.net's site directly.
func Plain(c Config, serveBotseed bool) http.Handler {
	acme := files(c.ACMERoot, "/", false, nil)
	tlsHandler := TLS(c)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := meta.From(r.Context())
		if serveBotseed {
			m.Host = "botseed.net"
			tlsHandler.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(normalize(r.URL.Path), "/.well-known/acme-challenge/") {
			m.Route = ACME
			acme.ServeHTTP(w, r)
			return
		}
		m.Route = Redirect
		page.Redirect(w, r, c.HTTPSBase(m.Host)+r.URL.RequestURI(), http.StatusMovedPermanently)
	})
}

// HTTPSBase is where plain HTTP for hostport belongs on HTTPS, as a URL
// with no path: the same host for our names (port kept in dev, where the
// TLS listener isn't on 443), https://acpwb.com for anything else — never
// an arbitrary client-supplied host.
func (c Config) HTTPSBase(hostport string) string {
	host := HostOf(hostport)
	switch c.classify(host) {
	case vhostAcpwb, vhostAcpwbWWW:
		if _, port, err := net.SplitHostPort(hostport); c.Dev && err == nil && isPort(port) {
			return "https://" + host + ":" + port
		}
		return "https://" + host
	case vhostBotseed, vhostBotseedWWW:
		return "https://botseed.net"
	}
	return "https://acpwb.com"
}

func isPort(s string) bool {
	if s == "" || len(s) > 5 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// CookieDomain returns the Domain attribute for the edge's client cookie
// on host — the registrable domain, so acpwb.com and every subdomain share
// one client ID — and false for hosts that get no cookie (unknown names,
// which are only ever redirected).
func (c Config) CookieDomain(host string) (string, bool) {
	switch c.classify(host) {
	case vhostAcpwb, vhostAcpwbWWW:
		switch {
		case host == "acpwb.com" || strings.HasSuffix(host, ".acpwb.com"):
			return "acpwb.com", true
		case host == "acpwb.example" || strings.HasSuffix(host, ".acpwb.example"):
			return "acpwb.example", true
		}
		return "", true // dev: localhost, 127.0.0.1 — host-only
	case vhostBotseed, vhostBotseedWWW:
		return "botseed.net", true
	}
	return "", false
}

const thirtyDays = 30 * 24 * time.Hour

func staticCache(h http.Header) {
	h.Set("Expires", time.Now().Add(thirtyDays).UTC().Format(http.TimeFormat))
	h.Set("Cache-Control", "max-age=2592000, public, immutable")
}

func noCache(h http.Header) {
	h.Set("Expires", time.Now().Add(-time.Second).UTC().Format(http.TimeFormat))
	h.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	h.Set("Pragma", "no-cache")
}
