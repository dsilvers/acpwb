package edge

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"

	"acpwb_edge/certs"
	"acpwb_edge/edgequeue"
	"acpwb_edge/proxy"
	"acpwb_edge/router"
	"acpwb_edge/tlsfp"
)

type sink struct {
	ch   chan *Record
	reqs chan *RequestRecord
}

func (s *sink) Push(key string, v any) {
	switch key {
	case edgequeue.ConnQueue:
		s.ch <- v.(*Record)
	case edgequeue.RequestQueue:
		select {
		case s.reqs <- v.(*RequestRecord):
		default: // tests that don't read requests mustn't block on them
		}
	}
}

func (s *sink) nextReq(t *testing.T) *RequestRecord {
	t.Helper()
	select {
	case r := <-s.reqs:
		return r
	case <-time.After(5 * time.Second):
		t.Fatal("no request record")
		return nil
	}
}

func (s *sink) next(t *testing.T) *Record {
	t.Helper()
	select {
	case r := <-s.ch:
		return r
	case <-time.After(5 * time.Second):
		t.Fatal("no connection record")
		return nil
	}
}

type upstream struct {
	mu   sync.Mutex
	last http.Header
	host string
	uri  string
	all  []string // every request URI seen
}

func (u *upstream) seen() (http.Header, string, string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.last, u.host, u.uri
}

type env struct {
	tlsAddr, plainAddr string
	sink               *sink
	up                 *upstream

	mu       sync.Mutex
	edgeSeen []string // every request URI the edge's own handler saw

	unblock chan struct{} // /wiki/block/ waits on it
	srv     *Server
}

func start(t *testing.T) *env { return startWith(t, nil) }

func startWith(t *testing.T, mod func(*Config)) *env {
	t.Helper()
	up := &upstream{}
	unblock := make(chan struct{})
	upL, _ := net.Listen("tcp", "127.0.0.1:0")
	upSrv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/wiki/block/" {
			select {
			case <-unblock:
			case <-r.Context().Done():
			}
		}
		if r.Header.Get("Upgrade") == "websocket" {
			// Minimal upgrade: 101, then echo.
			c, brw, _ := http.NewResponseController(w).Hijack()
			brw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
			brw.Flush()
			io.Copy(c, brw)
			c.Close()
			return
		}
		up.mu.Lock()
		io.Copy(io.Discard, r.Body)
		up.last, up.host, up.uri = r.Header.Clone(), r.Host, r.RequestURI
		up.all = append(up.all, r.RequestURI)
		up.mu.Unlock()
		w.Header().Set("Server", "gunicorn")
		io.WriteString(w, "upstream ok")
	})}
	go upSrv.Serve(upL)
	t.Cleanup(func() {
		select {
		case <-unblock:
		default:
			close(unblock)
		}
		upSrv.Close()
	})

	p := proxy.New(proxy.Upstream{Name: "django", Addr: upL.Addr().String(), Keepalive: 8, Timeout: 5 * time.Second})
	ws := proxy.New(proxy.Upstream{Name: "ws", Addr: upL.Addr().String(), Keepalive: 2, Timeout: 5 * time.Second})
	rc := router.Config{Upstreams: map[string]http.Handler{router.Django: p, router.AcpwbGo: p, router.WS: ws}}

	store, err := certs.SelfSigned("acpwb.com", "*.acpwb.com")
	if err != nil {
		t.Fatal(err)
	}
	sk := &sink{ch: make(chan *Record, 64), reqs: make(chan *RequestRecord, 1024)}
	e := &env{sink: sk, up: up, unblock: unblock}
	tlsRouter := router.TLS(rc)
	cfg := Config{
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10, GetCertificate: store.GetCertificate},
		TLS: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e.mu.Lock()
			e.edgeSeen = append(e.edgeSeen, r.RequestURI)
			e.mu.Unlock()
			tlsRouter.ServeHTTP(w, r)
		}),
		Plain: router.Plain(rc, false),
		Sink:  sk,
	}
	if mod != nil {
		mod(&cfg)
	}
	srv := New(cfg)
	e.srv = srv
	tl, _ := net.Listen("tcp", "127.0.0.1:0")
	pl, _ := net.Listen("tcp", "127.0.0.1:0")
	go srv.ServeTLS(tl)
	go srv.ServePlain(pl)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	})
	e.tlsAddr, e.plainAddr = tl.Addr().String(), pl.Addr().String()
	return e
}

// clientTLS is for talking to the edge's throwaway self-signed test cert.
func clientTLS(protos ...string) *tls.Config {
	return &tls.Config{ServerName: "acpwb.com", NextProtos: protos, InsecureSkipVerify: true}
}

func client(e *env, cfg *tls.Config) *http.Client {
	tr := &http.Transport{
		TLSClientConfig:   cfg,
		ForceAttemptHTTP2: len(cfg.NextProtos) > 0 && cfg.NextProtos[0] == "h2",
		DialTLSContext:    nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, e.tlsAddr)
		},
	}
	return &http.Client{Transport: tr, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

var ja4Re = regexp.MustCompile(`^t1[0-3][di]\d{4}[0-9a-z]{2}_[0-9a-f]{12}_[0-9a-f]{12}$`)

func TestHTTP1ProxyHeaders(t *testing.T) {
	e := start(t)
	c := client(e, clientTLS("http/1.1"))
	req, _ := http.NewRequest("GET", "https://acpwb.com/wiki/topic/?q=1", nil)
	req.Header.Set("X-JA4", "spoofed")
	req.Header.Set("X-Forwarded-For", "6.6.6.6")
	req.Header.Set("X-Real-IP", "6.6.6.6")
	req.Header.Set("X-TLS-Protocol", "SSLv3")
	req.Header["X_Ja4"] = []string{"spoofed-underscore"}
	req.Header.Set("X-Traffic-Class", "browser")
	req.Header.Set("X-Client-ASN", "15169")
	req.Header.Set("Accept-Language", "en-US")
	// Go's TLS stack claiming Chrome.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.Header.Get("Server") != "acpwb/web" || len(resp.Header.Values("Server")) != 1 {
		t.Errorf("Server header %v — upstream's must be replaced", resp.Header.Values("Server"))
	}

	h, host, uri := e.up.seen()
	if host != "acpwb.com" || uri != "/wiki/topic/?q=1" {
		t.Errorf("upstream Host %q URI %q", host, uri)
	}
	for k, want := range map[string]string{
		"X-Forwarded-For": "127.0.0.1", "X-Real-Ip": "127.0.0.1", "X-Forwarded-Proto": "https",
		"X-Client-Protocol": "HTTP/1.1", "X-Tls-Protocol": "TLSv1.3", "X-Tls-Cipher": "TLS_AES_128_GCM_SHA256",
	} {
		if got := strings.Join(h.Values(k), ","); got != want {
			t.Errorf("upstream %s = %q, want %q", k, got, want)
		}
	}
	if ja4 := h.Get("X-Ja4"); !ja4Re.MatchString(ja4) || len(h.Values("X-Ja4")) != 1 {
		t.Errorf("upstream X-Ja4 = %v", h.Values("X-Ja4"))
	}
	if h.Get("X-Ja3") == "" || !strings.HasPrefix(h.Get("X-Ja4h"), "ge11nn") {
		t.Errorf("X-Ja3 %q X-Ja4h %q", h.Get("X-Ja3"), h.Get("X-Ja4h"))
	}
	if _, ok := h["X_Ja4"]; ok {
		t.Error("underscored spoof header reached upstream")
	}
	if h.Get("X-H2-Fingerprint") != "" {
		t.Error("h2 fingerprint on an HTTP/1.1 request")
	}
	if got := h.Values("X-Traffic-Class"); len(got) != 1 || got[0] != "impersonator" {
		t.Errorf("upstream X-Traffic-Class = %v", got)
	}
	if got := h.Values("X-Client-Asn"); len(got) != 0 {
		t.Errorf("upstream X-Client-Asn = %v: no ASN database, and the client's copy must go", got)
	}

	c.CloseIdleConnections()
	rec := e.sink.next(t)
	if rec.Outcome != OutcomeOK || rec.ALPN != "http/1.1" || rec.TLSVer != "TLSv1.3" || rec.SNI != "acpwb.com" {
		t.Errorf("record %+v", rec)
	}
	if rec.JA4 != h.Get("X-Ja4") || rec.FirstJA4H != h.Get("X-Ja4h") || rec.FirstPath != "/wiki/topic/?q=1" ||
		rec.FirstRoute != router.Django || rec.FirstStatus != 200 || rec.RequestCount != 1 || rec.FirstHost != "acpwb.com" {
		t.Errorf("record %+v", rec)
	}
	if r := e.sink.nextReq(t); r.ClientClass != "impersonator" || !strings.HasPrefix(r.Label, "fake Chrome (") ||
		r.Signals == 0 || r.Network != "unknown" || r.JA4 != rec.JA4 {
		t.Errorf("request class %q label %q signals %b network %q", r.ClientClass, r.Label, r.Signals, r.Network)
	}
	if !strings.HasPrefix(rec.FirstHeaderNames, "Host,User-Agent,") || rec.BytesIn == 0 || rec.BytesOut == 0 {
		t.Errorf("record names %q bytes %d/%d", rec.FirstHeaderNames, rec.BytesIn, rec.BytesOut)
	}
}

func TestHTTP2(t *testing.T) {
	e := start(t)
	c := client(e, clientTLS("h2", "http/1.1"))
	for _, p := range []string{"/wiki/a/", "/reports/"} {
		resp, err := c.Get("https://acpwb.com" + p)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.ProtoMajor != 2 {
			t.Fatalf("negotiated %s, want HTTP/2", resp.Proto)
		}
	}
	h, _, _ := e.up.seen()
	if h.Get("X-Client-Protocol") != "HTTP/2.0" {
		t.Errorf("X-Client-Protocol %q", h.Get("X-Client-Protocol"))
	}
	fp := h.Get("X-H2-Fingerprint")
	// Go's own client sends its pseudo-headers as a,m,p,s (Chrome: m,a,s,p).
	if !regexp.MustCompile(`^[\d:;]+\|\d+\|[\d:,]+\|[amps](,[amps]){3}$`).MatchString(fp) {
		t.Errorf("X-H2-Fingerprint %q", fp)
	}
	if !strings.HasPrefix(h.Get("X-Ja4h"), "ge20nn") {
		t.Errorf("X-Ja4h %q", h.Get("X-Ja4h"))
	}
	c.CloseIdleConnections()
	rec := e.sink.next(t)
	if rec.ALPN != "h2" || rec.H2FP != fp || rec.RequestCount != 2 || rec.FirstPath != "/wiki/a/" {
		t.Errorf("record %+v", rec)
	}
	if !strings.HasPrefix(rec.FirstHeaderNames, ":method,:path") && !strings.HasPrefix(rec.FirstHeaderNames, ":authority") {
		t.Errorf("h2 header names %q", rec.FirstHeaderNames)
	}
}

// nginx accepted TLS 1.0/1.1; old-client bots must still reach the traps.
func TestTLS10ClientIsServed(t *testing.T) {
	e := start(t)
	cfg := clientTLS("http/1.1")
	cfg.MinVersion, cfg.MaxVersion = tls.VersionTLS10, tls.VersionTLS10
	resp, err := client(e, cfg).Get("https://acpwb.com/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	h, _, _ := e.up.seen()
	if h.Get("X-Tls-Protocol") != "TLSv1" || !strings.Contains(h.Get("X-Tls-Cipher"), "-SHA") {
		t.Errorf("TLS 1.0: protocol %q cipher %q", h.Get("X-Tls-Protocol"), h.Get("X-Tls-Cipher"))
	}
	if !strings.HasPrefix(h.Get("X-Ja4"), "t10d") {
		t.Errorf("X-Ja4 %q", h.Get("X-Ja4"))
	}
}

// An old client offering h2 anyway must get HTTP/1.1, not a connection
// x/net/http2 then refuses for inadequate security.
func TestTLS10ClientOfferingH2GetsHTTP1(t *testing.T) {
	e := start(t)
	cfg := clientTLS("h2", "http/1.1")
	cfg.MinVersion, cfg.MaxVersion = tls.VersionTLS10, tls.VersionTLS10
	c, err := tls.Dial("tcp", e.tlsAddr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if p := c.ConnectionState().NegotiatedProtocol; p != "http/1.1" {
		t.Fatalf("negotiated %q", p)
	}
	c.Write([]byte("GET / HTTP/1.1\r\nHost: acpwb.com\r\n\r\n"))
	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("%v %v", resp, err)
	}
}

func captureHello(t *testing.T) []byte {
	t.Helper()
	a, b := net.Pipe()
	go func() { tls.Client(a, clientTLS("h2")).Handshake(); a.Close() }()
	var raw []byte
	buf := make([]byte, 4096)
	for {
		n, err := b.Read(buf)
		raw = append(raw, buf[:n]...)
		if _, perr := tlsfp.Parse(raw); perr == nil {
			b.Close()
			return raw
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

// A connection that never completes the handshake still yields a record
// with the client's fingerprint — traffic the apps never see.
func TestFailedHandshakeIsRecorded(t *testing.T) {
	e := start(t)
	c, err := net.Dial("tcp", e.tlsAddr)
	if err != nil {
		t.Fatal(err)
	}
	c.Write(captureHello(t))
	c.Write([]byte("\x17\x03\x03\x00\x05garbage"))
	io.Copy(io.Discard, c)
	c.Close()
	rec := e.sink.next(t)
	if rec.Outcome != OutcomeHandshakeFailed || !ja4Re.MatchString(rec.JA4) || rec.JA3 == "" || rec.SNI != "acpwb.com" {
		t.Errorf("record %+v", rec)
	}
}

// Plain HTTP sent to the TLS port is redirected to HTTPS — to the same
// host only if it's one of ours.
func TestPlainHTTPOnTLSPort(t *testing.T) {
	rc := router.Config{}
	e := startWith(t, func(c *Config) { c.HTTPSBase = rc.HTTPSBase })
	for _, tc := range []struct{ host, path, want string }{
		{"archives-1999.acpwb.com", "/wp-login.php?x=1", "https://archives-1999.acpwb.com/wp-login.php?x=1"},
		{"evil.example", "/a", "https://acpwb.com/a"},
		{"acpwb.com", "http://evil.example/", "https://acpwb.com/"},
	} {
		c, _ := net.Dial("tcp", e.tlsAddr)
		c.Write([]byte("GET " + tc.path + " HTTP/1.1\r\nHost: " + tc.host + "\r\nUser-Agent: curl/8\r\n\r\n"))
		resp, err := http.ReadResponse(bufio.NewReader(c), nil)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		c.Close()
		if resp.StatusCode != 301 || resp.Header.Get("Location") != tc.want || resp.Header.Get("Server") != "acpwb/web" ||
			!strings.Contains(string(body), "The document has moved") {
			t.Errorf("%s%s: %d %v", tc.host, tc.path, resp.StatusCode, resp.Header)
		}
		req := e.sink.nextReq(t)
		rec := e.sink.next(t)
		if rec.Outcome != OutcomeNoHello || rec.FirstMethod != "GET" || rec.FirstStatus != 301 {
			t.Errorf("conn record %+v", rec)
		}
		if req.Status != 301 || req.Error != ErrPlainOnTLS || req.Host != tc.host || req.UserAgent != "curl/8" ||
			req.ConnID != rec.ConnID || req.ConnID == "" || req.Seq != 1 || req.ResponseBytes == 0 {
			t.Errorf("request record %+v", req)
		}
	}
}

// Every request gets a record linked to its connection by conn_id; its
// request_id reaches the upstream (a client's own X-Request-Id never
// does); the client cookie is issued once and then recognised.
func TestRequestRecords(t *testing.T) {
	rc := router.Config{}
	e := startWith(t, func(c *Config) { c.CookieDomain = rc.CookieDomain; c.CookieSecret = []byte("test") })
	c := client(e, clientTLS("http/1.1"))

	req, _ := http.NewRequest("POST", "https://acpwb.com/wiki/a/?q=1", strings.NewReader("hello"))
	req.Header.Set("X-Request-Id", "spoofed")
	req.Header.Set("User-Agent", "rec-test/1")
	req.Header.Set("Referer", "https://example.com/")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) != 1 || cookies[0].Name != ClientCookieName || cookies[0].Domain != "acpwb.com" ||
		!cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].MaxAge != clientCookieMaxAge {
		t.Fatalf("cookies %+v", cookies)
	}
	r1 := e.sink.nextReq(t)
	h, _, _ := e.up.seen()
	if h.Get("X-Request-Id") != r1.RequestID || len(r1.RequestID) != 36 {
		t.Errorf("upstream X-Request-Id %q, record %q", h.Get("X-Request-Id"), r1.RequestID)
	}
	if r1.Seq != 1 || r1.Method != "POST" || r1.Path != "/wiki/a/" || r1.Query != "q=1" || r1.Status != 200 ||
		r1.Route != router.Django || r1.UserAgent != "rec-test/1" || r1.Referer != "https://example.com/" ||
		r1.RequestBytes != 5 || r1.ResponseBytes != int64(len(body)) || r1.Error != "" || !r1.NewClient ||
		len(r1.ClientID) != 22 || r1.Protocol != "HTTP/1.1" || r1.DurationUS <= 0 || r1.TTFBMicros <= 0 || !r1.TLS ||
		!strings.Contains(r1.HeaderNames, "Referer") {
		t.Errorf("first request %+v", r1)
	}

	// Same connection, cookie sent back: same client, not new.
	req, _ = http.NewRequest("GET", "https://acpwb.com/static/missing.css", nil)
	req.AddCookie(cookies[0])
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if len(resp.Cookies()) != 0 {
		t.Errorf("cookie reissued: %v", resp.Cookies())
	}
	r2 := e.sink.nextReq(t)
	if r2.Seq != 2 || r2.ConnID != r1.ConnID || r2.ClientID != r1.ClientID || r2.NewClient ||
		r2.Status != 404 || r2.Route != router.Static {
		t.Errorf("second request %+v", r2)
	}

	// A forged cookie is replaced, not trusted.
	req, _ = http.NewRequest("GET", "https://acpwb.com/", nil)
	req.AddCookie(&http.Cookie{Name: ClientCookieName, Value: r1.ClientID + ".forgedforgedforg"})
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if r3 := e.sink.nextReq(t); r3.ClientID == r1.ClientID || !r3.NewClient {
		t.Errorf("forged cookie accepted: %+v", r3)
	}

	c.CloseIdleConnections()
	rec := e.sink.next(t)
	if rec.ConnID != r1.ConnID || rec.RequestCount != 3 || rec.ClientID != r1.ClientID || rec.Aggregated {
		t.Errorf("conn record %+v", rec)
	}
}

func TestHTTP2RequestRecord(t *testing.T) {
	e := start(t)
	c := client(e, clientTLS("h2"))
	resp, err := c.Get("https://acpwb.com/reports/")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	r := e.sink.nextReq(t)
	if r.Protocol != "HTTP/2.0" || r.Status != 200 || r.HeaderNames == "" || r.Seq != 1 {
		t.Errorf("record %+v", r)
	}
}

// Requests the HTTP/1 server rejects before any handler runs are still
// logged, with the status it answered.
func TestRejectedRequestIsRecorded(t *testing.T) {
	e := start(t)
	tc, err := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
	if err != nil {
		t.Fatal(err)
	}
	tc.Write([]byte("GET /ok/ HTTP/1.1\r\nHost: acpwb.com\r\n\r\n" +
		"GET /bad/ HTTP/1.1\r\nHost: acpwb.com\r\nUser-Agent: smuggler\r\nBad Header Line\r\n\r\n"))
	io.Copy(io.Discard, tc)
	tc.Close()
	ok := e.sink.nextReq(t)
	bad := e.sink.nextReq(t)
	if ok.Path != "/ok/" || ok.Error != "" || ok.Seq != 1 {
		t.Errorf("served request %+v", ok)
	}
	if bad.Path != "/bad/" || bad.Error != ErrRejected || bad.Status != 400 || bad.UserAgent != "smuggler" ||
		bad.Seq != 2 || bad.ConnID != ok.ConnID {
		t.Errorf("rejected request %+v", bad)
	}
	if rec := e.sink.next(t); rec.RequestCount != 1 || rec.UnservedReqs != 1 {
		t.Errorf("conn record %+v", rec)
	}
}

func TestConnectAndLeave(t *testing.T) {
	e := start(t)
	c, _ := net.Dial("tcp", e.tlsAddr)
	c.Close()
	if rec := e.sink.next(t); rec.Outcome != OutcomeNoHello || rec.BytesIn != 0 || rec.IPAddress != "127.0.0.1" {
		t.Errorf("record %+v", rec)
	}
}

func TestPlainListenerRedirects(t *testing.T) {
	e := start(t)
	c, _ := net.Dial("tcp", e.plainAddr)
	c.Write([]byte("GET /careers/?x=1 HTTP/1.1\r\nHost: archives-2001.acpwb.com\r\nConnection: close\r\n\r\n"))
	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	if resp.StatusCode != 301 || resp.Header.Get("Location") != "https://archives-2001.acpwb.com/careers/?x=1" {
		t.Errorf("%d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	rec := e.sink.next(t)
	if rec.Outcome != OutcomePlain || rec.FirstRoute != router.Redirect || rec.FirstStatus != 301 || rec.FirstJA4H == "" {
		t.Errorf("record %+v", rec)
	}
}

func TestWebSocketUpgradeThroughProxy(t *testing.T) {
	e := start(t)
	c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Write([]byte("GET /ws/requests/?token=x HTTP/1.1\r\nHost: acpwb.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n"))
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 101 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	c.Write([]byte("ping"))
	got := make([]byte, 4)
	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := io.ReadFull(br, got); err != nil || string(got) != "ping" {
		t.Fatalf("echo %q %v", got, err)
	}
}

func TestShutdownStopsAccepting(t *testing.T) {
	store, _ := certs.SelfSigned("acpwb.com")
	srv := New(Config{TLSConfig: &tls.Config{GetCertificate: store.GetCertificate}, TLS: http.NotFoundHandler(), Plain: http.NotFoundHandler()})
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	done := make(chan error, 1)
	go func() { done <- srv.ServeTLS(l) }()
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, net.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ServeTLS didn't return")
	}
}

// Go's HTTP/2 server answers connection-specific headers (h2spec 8.1.2.2)
// with a 400 rather than the RST_STREAM the spec asks for. Either way, the
// property that matters for an h2→HTTP/1.1 proxy is that none of them —
// request-smuggling material — ever reaches the upstream: each request must
// be rejected or forwarded without the header.
func TestH2ConnectionSpecificHeadersNotForwarded(t *testing.T) {
	e := start(t)
	for i, extra := range [][2]string{
		{"connection", "keep-alive, x-evil"}, {"keep-alive", "timeout=5"}, {"proxy-connection", "keep-alive"},
		{"transfer-encoding", "chunked"}, {"upgrade", "h2c"}, {"te", "gzip"}, {"te", "trailers"},
	} {
		path := fmt.Sprintf("/wiki/smuggle-%d/", i)
		status := h2Request(t, e, path, extra, [2]string{"x-evil", "1"})
		h, _, uri := e.up.seen()
		if uri != path {
			if status != 400 {
				t.Errorf("%s: not forwarded, but status %d (want 400 or forwarded)", extra[0], status)
			}
			continue
		}
		canon := http.CanonicalHeaderKey(extra[0])
		if v := h.Get(canon); v != "" && !(canon == "Te" && v == "trailers") {
			t.Errorf("%s: %q reached the upstream", canon, v)
		}
		if extra[0] == "connection" && h.Get("X-Evil") != "" {
			t.Error("a header named in Connection reached the upstream")
		}
	}
}

// h2Request sends one GET over a raw HTTP/2 connection (so headers Go's own
// client refuses to send can be sent) and returns the response status.
func h2Request(t *testing.T, e *env, path string, extra ...[2]string) int {
	t.Helper()
	c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("h2"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Write([]byte(http2.ClientPreface))
	fr := http2.NewFramer(c, c)
	fr.WriteSettings()
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	for _, f := range append([][2]string{{":method", "GET"}, {":authority", "acpwb.com"}, {":scheme", "https"}, {":path", path}}, extra...) {
		enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, BlockFragment: hb.Bytes(), EndStream: true, EndHeaders: true})
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	dec := hpack.NewDecoder(4096, nil)
	status := 0
	for {
		f, err := fr.ReadFrame()
		if err != nil {
			t.Fatal(err)
		}
		switch f := f.(type) {
		case *http2.HeadersFrame:
			fields, _ := dec.DecodeFull(f.HeaderBlockFragment())
			for _, hf := range fields {
				if hf.Name == ":status" {
					status, _ = strconv.Atoi(hf.Value)
				}
			}
			if f.StreamEnded() {
				return status
			}
		case *http2.DataFrame:
			if f.StreamEnded() {
				return status
			}
		case *http2.RSTStreamFrame:
			return -1
		}
	}
}

// Classic HTTP/1.1 desync payloads: each tries to make the edge and the
// upstream disagree about where the first request's body ends, so that
// "GET /smuggled" becomes a request only the upstream sees — one that skips
// the edge's header scrubbing and fingerprinting. Go's server rejects
// ambiguous framing or resolves it per RFC 9112 (where /smuggled is then
// just a pipelined request the edge itself handles), and ReverseProxy
// re-frames what it forwards. Invariant: every request the upstream gets is
// one the edge's handler saw.
func TestHTTP1RequestSmuggling(t *testing.T) {
	e := start(t)
	smuggled := "GET /smuggled HTTP/1.1\r\nHost: acpwb.com\r\n\r\n"
	for name, req := range map[string]string{
		"CL.TE": "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nContent-Length: 6\r\nTransfer-Encoding: chunked\r\n\r\n0\r\n\r\n" + smuggled,
		"TE.CL": "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nTransfer-Encoding: chunked\r\nContent-Length: 3\r\n\r\n" +
			fmt.Sprintf("%x\r\n%s\r\n0\r\n\r\n", len(smuggled), smuggled),
		"double CL":     "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nContent-Length: 0\r\nContent-Length: 44\r\n\r\n" + smuggled,
		"TE obfuscated": "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nTransfer-Encoding: xchunked\r\nContent-Length: 0\r\n\r\n" + smuggled,
		"TE space":      "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nTransfer-Encoding : chunked\r\n\r\n0\r\n\r\n" + smuggled,
		"TE list":       "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nTransfer-Encoding: chunked, identity\r\n\r\n0\r\n\r\n" + smuggled,
		"obs-fold":      "POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nTransfer-Encoding:\r\n chunked\r\nContent-Length: 0\r\n\r\n" + smuggled,
		"bare LF":       "POST /wiki/a/ HTTP/1.1\nHost: acpwb.com\nContent-Length: 0\nTransfer-Encoding: chunked\n\n0\n\n" + smuggled,
	} {
		c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
		if err != nil {
			t.Fatal(err)
		}
		c.Write([]byte(req))
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		io.Copy(io.Discard, c)
		c.Close()
		e.up.mu.Lock()
		e.mu.Lock()
		seen := map[string]int{}
		for _, u := range e.edgeSeen {
			seen[u]++
		}
		for _, u := range e.up.all {
			if seen[u] == 0 {
				t.Errorf("%s: upstream got %q, which the edge never handled", name, u)
			}
			seen[u]--
		}
		e.up.all, e.edgeSeen = nil, nil
		e.mu.Unlock()
		e.up.mu.Unlock()
	}
}

// Slowloris-style clients must not hold connections open: a stalled
// ClientHello and a trickled request body are each cut off by a deadline.
func TestSlowClientsAreCutOff(t *testing.T) {
	e := startWith(t, func(c *Config) {
		c.HandshakeTimeout = 300 * time.Millisecond
		c.Limits.StallTimeout = 300 * time.Millisecond
	})
	closedWithin := func(name string, c net.Conn, d time.Duration) {
		t.Helper()
		c.SetReadDeadline(time.Now().Add(d))
		start := time.Now()
		_, err := io.Copy(io.Discard, c)
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			t.Errorf("%s: still open after %v", name, time.Since(start))
		}
	}

	// Half a ClientHello, then silence.
	raw, _ := net.Dial("tcp", e.tlsAddr)
	raw.Write(captureHello(t)[:40])
	closedWithin("stalled ClientHello", raw, 3*time.Second)
	raw.Close()
	if rec := e.sink.next(t); rec.Outcome != OutcomeTimeout {
		t.Errorf("stalled hello outcome %q, want timeout", rec.Outcome)
	}

	// A POST that promises a body and trickles one byte.
	c, _ := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
	c.Write([]byte("POST /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\nContent-Length: 100000\r\n\r\nx"))
	closedWithin("trickled body", c, 3*time.Second)
	c.Close()
}

// The overload valve is process-wide, never per client.
func TestOverloadValve(t *testing.T) {
	e := startWith(t, func(c *Config) { c.Limits.RequestsInFlight = 1 })
	cl := client(e, clientTLS("h2"))
	go cl.Get("https://acpwb.com/wiki/block/") // holds the only slot
	time.Sleep(200 * time.Millisecond)
	resp, err := cl.Get("https://acpwb.com/wiki/other/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 503 || resp.Header.Get("Retry-After") != "1" {
		t.Fatalf("over the limit: %d %v", resp.StatusCode, resp.Header)
	}
	close(e.unblock)
	time.Sleep(100 * time.Millisecond)
	cl.CloseIdleConnections()
	rec := e.sink.next(t)
	if rec.RejectedReqs != 1 || rec.RequestCount != 2 {
		t.Errorf("record %+v", rec)
	}
}

func TestRecordAggregation(t *testing.T) {
	e := startWith(t, func(c *Config) { c.Limits.RecordsPerIPMin = 2 })
	for range 5 {
		c, _ := net.Dial("tcp", e.tlsAddr)
		c.Write([]byte("junk that isn't TLS"))
		io.Copy(io.Discard, c)
		c.Close()
	}
	a, b := e.sink.next(t), e.sink.next(t)
	if a.Aggregated || b.Aggregated || a.ConnCount != 1 {
		t.Fatalf("first two should be individual: %+v %+v", a, b)
	}
	e.srv.records.flush()
	agg := e.sink.next(t)
	if !agg.Aggregated || agg.ConnCount != 3 || agg.Outcome != OutcomeNoHello || agg.BytesIn != 3*int64(len("junk that isn't TLS")) {
		t.Errorf("aggregated record %+v", agg)
	}
}

func TestH2AbuseCountersAndGoaway(t *testing.T) {
	e := start(t)
	c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("h2"))
	if err != nil {
		t.Fatal(err)
	}
	c.Write([]byte(http2.ClientPreface))
	fr := http2.NewFramer(c, c)
	fr.WriteSettings()
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	for _, f := range [][2]string{{":method", "GET"}, {":authority", "acpwb.com"}, {":scheme", "https"}, {":path", "/wiki/x/"}} {
		enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	for id := uint32(1); id <= 19; id += 2 { // rapid reset x10
		fr.WriteHeaders(http2.HeadersFrameParam{StreamID: id, BlockFragment: hb.Bytes(), EndStream: true, EndHeaders: true})
		fr.WriteRSTStream(id, http2.ErrCodeCancel)
	}
	// Then a protocol violation: DATA on stream 0 → GOAWAY PROTOCOL_ERROR.
	fr.AllowIllegalWrites = true
	fr.WriteData(0, false, []byte("x"))
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	io.Copy(io.Discard, c)
	c.Close()
	rec := e.sink.next(t)
	if rec.H2Streams != 10 || rec.H2Resets != 10 || rec.H2Goaway != "PROTOCOL_ERROR" {
		t.Errorf("streams=%d resets=%d goaway=%q", rec.H2Streams, rec.H2Resets, rec.H2Goaway)
	}
}

func TestCloseReasons(t *testing.T) {
	e := startWith(t, func(c *Config) { c.HandshakeTimeout = 200 * time.Millisecond })
	c, _ := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
	c.Close()
	if rec := e.sink.next(t); rec.CloseReason != CloseClient {
		t.Errorf("client close: %q", rec.CloseReason)
	}
	raw, _ := net.Dial("tcp", e.tlsAddr)
	raw.Write(captureHello(t)[:20])
	io.Copy(io.Discard, raw)
	raw.Close()
	if rec := e.sink.next(t); rec.CloseReason != CloseReadTimeout || rec.Outcome != OutcomeTimeout {
		t.Errorf("stalled hello: %q %q", rec.Outcome, rec.CloseReason)
	}

	// After an HTTP/1 request net/http cancels its background read with a
	// past deadline; that timeout isn't why the connection ended.
	c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("http/1.1"))
	if err != nil {
		t.Fatal(err)
	}
	c.Write([]byte("GET /wiki/a/ HTTP/1.1\r\nHost: acpwb.com\r\n\r\n"))
	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	c.Close()
	if rec := e.sink.next(t); rec.CloseReason != CloseClient || rec.RequestCount != 1 {
		t.Errorf("after a request: close_reason=%q requests=%d", rec.CloseReason, rec.RequestCount)
	}
}

// A connection that resets nearly every stream it opens is ended, and
// recorded as such; a client resetting a few streams is untouched (the
// rapid-reset test above resets 10 and keeps its connection).
func TestRapidResetConnectionIsClosed(t *testing.T) {
	e := start(t)
	c, err := tls.Dial("tcp", e.tlsAddr, clientTLS("h2"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Write([]byte(http2.ClientPreface))
	fr := http2.NewFramer(c, c)
	fr.WriteSettings()
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	for _, f := range [][2]string{{":method", "GET"}, {":authority", "acpwb.com"}, {":scheme", "https"}, {":path", "/wiki/x/"}} {
		enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	go func() {
		for id := uint32(1); id < 20000; id += 2 {
			if fr.WriteHeaders(http2.HeadersFrameParam{StreamID: id, BlockFragment: hb.Bytes(), EndStream: true, EndHeaders: true}) != nil ||
				fr.WriteRSTStream(id, http2.ErrCodeCancel) != nil {
				return
			}
		}
	}()
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = io.Copy(io.Discard, c)
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		t.Fatal("rapid-reset connection was not closed")
	}
	rec := e.sink.next(t)
	if rec.CloseReason != CloseRapidReset || rec.H2Resets < 200 {
		t.Errorf("close_reason=%q resets=%d streams=%d", rec.CloseReason, rec.H2Resets, rec.H2Streams)
	}
}

type liveTap struct {
	active atomic.Bool
	mu     sync.Mutex
	reqs   []*RequestRecord
	tls    []ConnTLS
	conns  []*Record
}

func (l *liveTap) Active() bool { return l.active.Load() }
func (l *liveTap) Request(rec *RequestRecord, t ConnTLS) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := *rec
	l.reqs, l.tls = append(l.reqs, &cp), append(l.tls, t)
}
func (l *liveTap) Conn(rec *Record) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := *rec
	l.conns = append(l.conns, &cp)
}

// The live stream gets each request with its connection's TLS details,
// every connection (before aggregation), and nothing while inactive.
func TestLiveFeed(t *testing.T) {
	lt := &liveTap{}
	e := startWith(t, func(c *Config) { c.Live = lt; c.Limits.RecordsPerIPMin = 1 })

	c := client(e, clientTLS("h2", "http/1.1"))
	get := func(p string) {
		resp, err := c.Get("https://acpwb.com" + p)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	get("/unwatched/")
	c.CloseIdleConnections()
	e.sink.next(t)
	lt.active.Store(true)
	get("/wiki/a/")
	c.CloseIdleConnections()
	rec := e.sink.next(t)
	for range 3 {
		nc, _ := net.Dial("tcp", e.tlsAddr)
		nc.Write([]byte("junk that isn't TLS"))
		io.Copy(io.Discard, nc)
		nc.Close()
	}
	e.sink.next(t)

	lt.mu.Lock()
	defer lt.mu.Unlock()
	if len(lt.reqs) != 1 || lt.reqs[0].Path != "/wiki/a/" {
		t.Fatalf("live requests %+v", lt.reqs)
	}
	ct := lt.tls[0]
	if ct.Version != "TLSv1.3" || ct.ALPN != "h2" || ct.SNI != "acpwb.com" || !ja4Re.MatchString(lt.reqs[0].JA4) || ct.H2FP == "" || ct.H2FP != rec.H2FP {
		t.Errorf("live TLS %+v, request ja4 %q (conn record %+v)", ct, lt.reqs[0].JA4, rec)
	}
	if r := lt.reqs[0]; r.ClientClass == "" || r.Label == "" || r.Group == "" || r.Network != "unknown" {
		t.Errorf("live request classification: %q %q %q %q", r.ClientClass, r.Label, r.Group, r.Network)
	}
	// One conn with the request, then all three junk ones even though the
	// gate aggregated two of them.
	if len(lt.conns) != 4 || lt.conns[0].RequestCount != 1 || lt.conns[3].Outcome != OutcomeNoHello || lt.conns[3].Aggregated {
		t.Errorf("live conns %d: %+v", len(lt.conns), lt.conns)
	}
}
