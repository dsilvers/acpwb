package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"acpwb_edge/meta"
)

// One row per nginx/acpwb.com location, plus the precedence cases that
// depend on nginx's regex-before-prefix and file-order rules.
func TestAcpwbTargetMatchesNginx(t *testing.T) {
	const main, arch, pol = "acpwb.com", "archives-1999.acpwb.com", "policy-epa.acpwb.com"
	for _, tc := range []struct{ host, path, want string }{
		// Regex 1: archive day-level → acpwb_go (beats the /archive/ prefix).
		{main, "/archive/1999/06/15/", AcpwbGo},
		{main, "/archive/1999/6/5", AcpwbGo},
		{main, "/archive/1999/06/15/some-slug/source-code/", AcpwbGo},
		// /archive/ prefix → Django (index + year index).
		{main, "/archive/", Django},
		{main, "/archive/1999/", Django},
		{main, "/archive/1999/06/", Django},
		// Regex 2: export.csv → Django, even on an archive subdomain.
		{arch, "/06/15/slug/export.csv", Django},
		{main, "/06/15/slug/export.csv", Django},
		// Regex 3: day-level on an archive subdomain → acpwb_go; on the
		// main domain it's archive_trap_yearless → Django.
		{arch, "/06/15/", AcpwbGo},
		{arch, "/06/15/some-slug/", AcpwbGo},
		{arch, "/6/1", AcpwbGo},
		{main, "/06/15/some-slug/", Django},
		{arch, "/", Django},    // subdomain root (archive index)
		{arch, "/06/", Django}, // month index
		// Regex 4: policy year/month/detail → acpwb_go only on policy hosts.
		{pol, "/2024/", AcpwbGo},
		{pol, "/2024/03/", AcpwbGo},
		{pol, "/2024/03/15/rule-slug/", AcpwbGo},
		{pol, "/2024/03/15/rule-slug", AcpwbGo},
		{pol, "/", Django},
		{pol, "/robots.txt", Django},
		{pol, "/sitemap.xml", Django},
		{main, "/2024/", Django},
		// Prefix locations.
		{main, "/source-code/", AcpwbGo},
		{main, "/source-code/1999/06/", AcpwbGo},
		{main, "/public-policy/", AcpwbGo},
		{main, "/public-policy/2024/03/", AcpwbGo},
		{main, "/wiki/some-topic/", Django},
		{main, "/reports/", Django},
		{main, "/.well-known/tokens/abc/ping", Django},
		{main, "/ws/requests/", WS},
		{main, "/static/css/site.css", Static},
		{main, "/static/graphs/traffic_1h.png", StaticGraphs},
		{main, "/", Django},
		{main, "/our-people/", Django},
		{main, "/wp-login.php", Django},
		{main, "/.env", Django},
		// nginx normalizes before matching.
		{main, "//archive//1999/06/15/", AcpwbGo},
		{main, "/wiki/../source-code/", AcpwbGo},
		{"ARCHIVES-1999.ACPWB.COM", "/06/15/", Django}, // HostOf lowercases first; raw uppercase doesn't match
		{"archives-1999.acpwb.example", "/06/15/", AcpwbGo},
	} {
		if got := AcpwbTarget(tc.host, tc.path); got != tc.want {
			t.Errorf("%s%s → %s, want %s", tc.host, tc.path, got, tc.want)
		}
	}
}

func TestBotseedTarget(t *testing.T) {
	for p, want := range map[string]string{
		"/ws/": BotseedWS, "/ws/stream": BotseedWS, "/api/random?n=5": BotseedAPI,
		"/": BotseedFiles, "/docs/": BotseedFiles, "/api": BotseedFiles,
	} {
		if got := BotseedTarget(strings.SplitN(p, "?", 2)[0]); got != want {
			t.Errorf("%s → %s, want %s", p, got, want)
		}
	}
}

func TestHostOf(t *testing.T) {
	for in, want := range map[string]string{
		"acpwb.com": "acpwb.com", "ACPWB.com:443": "acpwb.com", "acpwb.com.": "acpwb.com",
		"[::1]:8001": "[::1]", "127.0.0.1:8001": "127.0.0.1", "": "",
	} {
		if got := HostOf(in); got != want {
			t.Errorf("HostOf(%q) = %q, want %q", in, got, want)
		}
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

type fixture struct {
	tls, plain, devPlain http.Handler
	hit                  map[string]int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	write := func(p, body string) {
		full := filepath.Join(dir, p)
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte(body), 0o644)
	}
	write("static/css/site.css", strings.Repeat("body{color:red}\n", 200))
	write("static/graphs/traffic_1h.png", "png")
	write("botseed/index.html", "<h1>botseed</h1>")
	write("botseed/docs/index.html", "<h1>docs</h1>")
	write("certbot/.well-known/acme-challenge/tok", "challenge-response")
	write("secret.txt", "outside")

	f := &fixture{hit: map[string]int{}}
	up := func(name string, body string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f.hit[name]++
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, body)
		})
	}
	big := strings.Repeat("<p>acpwb</p>", 200)
	cfg := Config{
		StaticRoot:  filepath.Join(dir, "static"),
		BotseedRoot: filepath.Join(dir, "botseed"),
		ACMERoot:    filepath.Join(dir, "certbot"),
		Upstreams: map[string]http.Handler{
			Django: up(Django, big), AcpwbGo: up(AcpwbGo, big), WS: up(WS, big),
			BotseedWS: up(BotseedWS, "ws"), BotseedAPI: up(BotseedAPI, "{}"),
		},
	}
	f.tls, f.plain, f.devPlain = TLS(cfg), Plain(cfg, false), Plain(cfg, true)
	return f
}

func serve(h http.Handler, method, host, target string, hdr ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	r.Host = host
	for i := 0; i < len(hdr); i += 2 {
		r.Header.Set(hdr[i], hdr[i+1])
	}
	r = r.WithContext(meta.With(r.Context(), &meta.Request{Host: HostOf(host)}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestStaticFiles(t *testing.T) {
	f := newFixture(t)
	w := serve(f.tls, "GET", "acpwb.com", "/static/css/site.css")
	if w.Code != 200 || !strings.Contains(w.Header().Get("Cache-Control"), "immutable") || w.Header().Get("Expires") == "" {
		t.Fatalf("static: %d %v", w.Code, w.Header())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Fatalf("content type %q", ct)
	}
	w = serve(f.tls, "GET", "acpwb.com", "/static/graphs/traffic_1h.png")
	if w.Code != 200 || !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("graphs: %d %v", w.Code, w.Header())
	}
	for target, want := range map[string]int{
		"/static/css/":        403, // directory, no index
		"/static/missing.css": 404,
	} {
		if w := serve(f.tls, "GET", "acpwb.com", target); w.Code != want {
			t.Errorf("%s: %d, want %d", target, w.Code, want)
		}
	}
	// Dot segments are resolved before location matching, as nginx does, so
	// these leave /static/ entirely and go to Django — never to a file
	// outside the static root.
	for _, target := range []string{"/static/../secret.txt", "/static/%2e%2e/secret.txt", "/static/css/../../secret.txt"} {
		if w := serve(f.tls, "GET", "acpwb.com", target); strings.Contains(w.Body.String(), "outside") {
			t.Errorf("%s served a file outside the static root", target)
		}
	}
	if f.hit[Django] != 3 {
		t.Errorf("dot-segment paths should have gone to Django, hits %v", f.hit)
	}
	if w := serve(f.tls, "POST", "acpwb.com", "/static/css/site.css"); w.Code != 405 {
		t.Errorf("POST static: %d", w.Code)
	}
}

func TestGzip(t *testing.T) {
	f := newFixture(t)
	w := serve(f.tls, "GET", "acpwb.com", "/wiki/x/", "Accept-Encoding", "gzip")
	if w.Header().Get("Content-Encoding") != "gzip" || !strings.Contains(w.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatalf("proxied HTML not gzipped: %v", w.Header())
	}
	w = serve(f.tls, "GET", "acpwb.com", "/static/css/site.css", "Accept-Encoding", "gzip")
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("static CSS not gzipped: %v", w.Header())
	}
	w = serve(f.tls, "GET", "acpwb.com", "/ws/requests/", "Accept-Encoding", "gzip")
	if w.Header().Get("Content-Encoding") != "" {
		t.Fatal("WebSocket route must not be gzip-wrapped")
	}
}

func TestVhostsAndRedirects(t *testing.T) {
	f := newFixture(t)
	check := func(h http.Handler, host, target string, code int, loc string) {
		t.Helper()
		w := serve(h, "GET", host, target)
		if w.Code != code || w.Header().Get("Location") != loc {
			t.Errorf("%s%s: %d %q, want %d %q", host, target, w.Code, w.Header().Get("Location"), code, loc)
		}
	}
	check(f.tls, "www.acpwb.com", "/wiki/?a=1", 301, "https://acpwb.com/wiki/?a=1")
	check(f.tls, "www.botseed.net", "/x", 301, "https://botseed.net/x")
	check(f.tls, "evil.example.org", "/", 301, "https://acpwb.com/")
	check(f.tls, "203.0.113.5", "/", 301, "https://acpwb.com/")
	check(f.tls, "localhost:8001", "/", 301, "https://acpwb.com/") // not dev mode
	check(f.plain, "archives-1999.acpwb.com", "/06/15/?x", 301, "https://archives-1999.acpwb.com/06/15/?x")
	check(f.plain, "www.botseed.net", "/a", 301, "https://botseed.net/a")
	check(f.plain, "evil.example.org", "/a", 301, "https://acpwb.com/a")

	if w := serve(f.plain, "GET", "botseed.net", "/.well-known/acme-challenge/tok"); w.Code != 200 || w.Body.String() != "challenge-response" {
		t.Errorf("ACME: %d %q", w.Code, w.Body.String())
	}
	if w := serve(f.plain, "GET", "acpwb.com", "/.well-known/acme-challenge/nope"); w.Code != 404 {
		t.Errorf("missing ACME token: %d", w.Code)
	}

	if w := serve(f.tls, "GET", "botseed.net", "/"); w.Code != 200 || w.Body.String() != "<h1>botseed</h1>" {
		t.Errorf("botseed index: %d %q", w.Code, w.Body.String())
	}
	if w := serve(f.tls, "GET", "botseed.net", "/docs"); w.Code != 200 || w.Body.String() != "<h1>docs</h1>" {
		t.Errorf("botseed try_files $uri/: %d %q", w.Code, w.Body.String())
	}
	if w := serve(f.devPlain, "GET", "localhost:8090", "/api/x"); w.Code != 200 || f.hit[BotseedAPI] != 1 {
		t.Errorf("dev botseed plain listener: %d", w.Code)
	}
	serve(f.tls, "GET", "archives-2001.acpwb.com", "/03/04/x/")
	if f.hit[AcpwbGo] != 1 {
		t.Errorf("archive day-level didn't reach acpwb_go: %v", f.hit)
	}
}
