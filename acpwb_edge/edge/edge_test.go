package edge

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"acpwb_edge/certs"
	"acpwb_edge/proxy"
	"acpwb_edge/router"
	"acpwb_edge/tlsfp"
)

type sink struct{ ch chan *Record }

func (s *sink) Push(v any) { s.ch <- v.(*Record) }

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
}

func start(t *testing.T) *env {
	t.Helper()
	up := &upstream{}
	upL, _ := net.Listen("tcp", "127.0.0.1:0")
	upSrv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		up.last, up.host, up.uri = r.Header.Clone(), r.Host, r.RequestURI
		up.mu.Unlock()
		w.Header().Set("Server", "gunicorn")
		io.WriteString(w, "upstream ok")
	})}
	go upSrv.Serve(upL)
	t.Cleanup(func() { upSrv.Close() })

	p := proxy.New(proxy.Upstream{Name: "django", Addr: upL.Addr().String(), Keepalive: 8, Timeout: 5 * time.Second})
	ws := proxy.New(proxy.Upstream{Name: "ws", Addr: upL.Addr().String(), Keepalive: 2, Timeout: 5 * time.Second})
	rc := router.Config{Upstreams: map[string]http.Handler{router.Django: p, router.AcpwbGo: p, router.WS: ws}}

	store, err := certs.SelfSigned("acpwb.com", "*.acpwb.com")
	if err != nil {
		t.Fatal(err)
	}
	sk := &sink{ch: make(chan *Record, 64)}
	srv := New(Config{
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10, GetCertificate: store.GetCertificate},
		TLS:       router.TLS(rc),
		Plain:     router.Plain(rc, false),
		Sink:      sk,
	})
	tl, _ := net.Listen("tcp", "127.0.0.1:0")
	pl, _ := net.Listen("tcp", "127.0.0.1:0")
	go srv.ServeTLS(tl)
	go srv.ServePlain(pl)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	})
	return &env{tlsAddr: tl.Addr().String(), plainAddr: pl.Addr().String(), sink: sk, up: up}
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
	req.Header.Set("Accept-Language", "en-US")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.Header.Get("Server") != "nginx" || len(resp.Header.Values("Server")) != 1 {
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

	c.CloseIdleConnections()
	rec := e.sink.next(t)
	if rec.Outcome != OutcomeOK || rec.ALPN != "http/1.1" || rec.TLSVer != "TLSv1.3" || rec.SNI != "acpwb.com" {
		t.Errorf("record %+v", rec)
	}
	if rec.JA4 != h.Get("X-Ja4") || rec.FirstJA4H != h.Get("X-Ja4h") || rec.FirstPath != "/wiki/topic/?q=1" ||
		rec.FirstRoute != router.Django || rec.FirstStatus != 200 || rec.RequestCount != 1 || rec.FirstHost != "acpwb.com" {
		t.Errorf("record %+v", rec)
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

func TestPlainHTTPOnTLSPort(t *testing.T) {
	e := start(t)
	c, _ := net.Dial("tcp", e.tlsAddr)
	c.Write([]byte("GET /wp-login.php HTTP/1.1\r\nHost: acpwb.com\r\n\r\n"))
	body, _ := io.ReadAll(c)
	c.Close()
	if !strings.Contains(string(body), "The plain HTTP request was sent to HTTPS port") {
		t.Errorf("response %q", body)
	}
	rec := e.sink.next(t)
	if rec.Outcome != OutcomeNoHello || rec.FirstMethod != "GET" || rec.FirstPath != "/wp-login.php" {
		t.Errorf("record %+v", rec)
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
