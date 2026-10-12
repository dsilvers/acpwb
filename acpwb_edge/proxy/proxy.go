// Package proxy is acpwb_edge's reverse proxy to the apps behind it
// (gunicorn/Django, acpwb_go, the ws service, botseed). Upstream requests
// look like what nginx sent: plain HTTP/1.1 with keep-alive, the original
// request URI, Host = nginx's $host, and the X-Real-IP / X-Forwarded-* /
// X-Client-Protocol / X-TLS-* headers — plus the fingerprint headers
// (X-JA4, X-JA3, X-JA4H, X-H2-Fingerprint) only the edge can compute.
//
// Every one of those headers is deleted from the client's request first,
// whatever the client sent: they are the apps' only source for the client's
// IP and fingerprints, so a client must never be able to supply them.
package proxy

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"acpwb_edge/meta"
	"acpwb_edge/page"
)

type Upstream struct {
	Name      string
	Addr      string        // "host:port" or "unix:/path/to.sock"
	Keepalive int           // idle connections kept, like nginx's upstream keepalive
	Timeout   time.Duration // time to response headers (nginx proxy_read_timeout)
}

// New returns a handler proxying to u.
func New(u Upstream) http.Handler {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	network, address, urlHost := "tcp", u.Addr, u.Addr
	if path, ok := strings.CutPrefix(u.Addr, "unix:"); ok {
		network, address, urlHost = "unix", path, u.Name
	}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
		MaxIdleConns:          u.Keepalive,
		MaxIdleConnsPerHost:   u.Keepalive,
		IdleConnTimeout:       60 * time.Second,
		ResponseHeaderTimeout: u.Timeout,
		ExpectContinueTimeout: time.Second,
		DisableCompression:    true, // pass Accept-Encoding through untouched, as nginx does
	}
	el := &errLog{name: u.Name}
	return &httputil.ReverseProxy{
		Transport: tr,
		Rewrite: func(pr *httputil.ProxyRequest) {
			// Rewrite has already removed Forwarded and X-Forwarded-*.
			m := meta.From(pr.In.Context())
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = urlHost
			pr.Out.Host = m.Host
			SetHeaders(pr.Out.Header, m)
		},
		ModifyResponse: func(resp *http.Response) error {
			// nginx never passed the upstream's Server header through.
			resp.Header.Del("Server")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(err, context.Canceled) {
				return // client went away
			}
			el.log(err)
			code := http.StatusBadGateway
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() || errors.Is(err, context.DeadlineExceeded) {
				code = http.StatusGatewayTimeout
			}
			page.Error(w, code)
		},
		ErrorLog: log.New(discard{}, "", 0), // ErrorHandler logs, rate-limited
	}
}

// reserved reports whether a request header name is one the edge owns.
// Underscored variants count too: gunicorn would drop them, but anything
// reading raw header names shouldn't get a spoofable lookalike.
func reserved(name string) bool {
	n := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	switch n {
	case "x-real-ip", "x-client-protocol", "forwarded":
		return true
	}
	for _, p := range []string{"x-forwarded-", "x-tls-", "x-ja3", "x-ja4", "x-h2-"} {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

// SetHeaders scrubs every client-supplied copy of the edge's headers from
// h, then sets them from m.
func SetHeaders(h http.Header, m *meta.Request) {
	for k := range h {
		if reserved(k) {
			delete(h, k)
		}
	}
	set := func(k, v string) {
		if v != "" {
			h[k] = []string{v}
		}
	}
	set("X-Real-Ip", m.ClientIP)
	set("X-Forwarded-For", m.ClientIP)
	proto := "http"
	if m.TLS {
		proto = "https"
	}
	set("X-Forwarded-Proto", proto)
	set("X-Client-Protocol", m.ClientProtocol)
	set("X-Tls-Protocol", m.TLSProtocol)
	set("X-Tls-Cipher", m.TLSCipher)
	set("X-Ja4", m.JA4)
	set("X-Ja3", m.JA3)
	set("X-Ja4h", m.JA4H)
	set("X-H2-Fingerprint", m.H2Fingerprint)
}

// errLog logs upstream errors at most once per 10s per upstream, with a
// count of what was suppressed, so a down upstream doesn't flood the journal.
type errLog struct {
	name       string
	mu         sync.Mutex
	last       time.Time
	suppressed int
}

func (e *errLog) log(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if time.Since(e.last) < 10*time.Second {
		e.suppressed++
		return
	}
	log.Printf("proxy %s: %v (%d similar errors suppressed)", e.name, err, e.suppressed)
	e.last, e.suppressed = time.Now(), 0
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
