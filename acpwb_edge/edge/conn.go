package edge

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"acpwb_edge/httpfp"
	"acpwb_edge/meta"
	"acpwb_edge/tlsfp"
)

// connInfo is everything known about one client connection. It becomes a
// Record when the connection closes.
type connInfo struct {
	ip       string
	accepted time.Time
	tls      bool
	raw      net.Conn // the TCP socket, for kernel RTT
	sink     Sink

	in, out atomic.Int64

	mu sync.Mutex

	hello    *tlsfp.Hello
	fps      tlsfp.Fingerprints
	helloErr error
	rawHead  []byte // first bytes, kept when they weren't TLS

	outcome     string
	handshakeAt time.Time
	sni, alpn   string
	tlsVersion  string
	cipher      string
	h1          *httpfp.H1Conn
	h2          *httpfp.H2Conn
	requests    int
	first       *firstRequest
}

type firstRequest struct {
	host, method, path, ua, ja4h, names, route string
	status                                     int
}

func newConnInfo(raw net.Conn, isTLS bool, sink Sink) *connInfo {
	ci := &connInfo{accepted: time.Now(), tls: isTLS, raw: raw, sink: sink}
	if a, ok := raw.RemoteAddr().(*net.TCPAddr); ok {
		ip := a.IP
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		ci.ip = ip.String()
	} else if host, _, err := net.SplitHostPort(raw.RemoteAddr().String()); err == nil {
		ci.ip = host
	}
	return ci
}

func (ci *connInfo) handshakeDone(st tls.ConnectionState) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.outcome = OutcomeOK
	ci.handshakeAt = time.Now()
	ci.sni = st.ServerName
	ci.alpn = st.NegotiatedProtocol
	ci.tlsVersion = tlsVersionName(st.Version)
	ci.cipher = cipherName(st.CipherSuite)
}

func (ci *connInfo) handshakeFailed(err error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	outcome := OutcomeHandshakeFailed
	var ne net.Error
	switch {
	case ci.in.Load() == 0:
		outcome = OutcomeNoHello
	case ci.hello == nil && ci.helloErr != nil:
		outcome = OutcomeNoHello // the client didn't speak TLS at all
	case errors.Is(err, context.DeadlineExceeded) || errors.As(err, &ne) && ne.Timeout():
		outcome = OutcomeTimeout
	}
	ci.outcome = outcome
	if ci.hello != nil {
		ci.sni = ci.hello.ServerName
	}
}

// fingerprints returns the ClientHello fingerprints (zero if none parsed).
func (ci *connInfo) fingerprints() tlsfp.Fingerprints {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.fps
}

// plainHTTPOnTLS reports whether the client sent an HTTP request line to
// the TLS port.
func (ci *connInfo) plainHTTPOnTLS() bool {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	_, _, ok := requestLine(ci.rawHead)
	return ok
}

// noteRequest counts a request and keeps a summary of the first one.
func (ci *connInfo) noteRequest(r *http.Request, m *meta.Request, names []string, status int) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.requests++
	if ci.first != nil {
		return
	}
	ci.first = &firstRequest{
		host: m.Host, method: r.Method, path: r.URL.RequestURI(), ua: r.UserAgent(),
		ja4h: m.JA4H, names: strings.Join(names, ","), route: m.Route, status: status,
	}
}

// record builds the connection's Record at close.
func (ci *connInfo) record(now time.Time) *Record {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	rec := &Record{
		Timestamp:      isoTime(ci.accepted),
		IPAddress:      ci.ip,
		SNI:            ci.sni,
		ALPN:           ci.alpn,
		TLSVer:         ci.tlsVersion,
		TLSCipher:      ci.cipher,
		Outcome:        ci.outcome,
		JA4:            ci.fps.JA4,
		JA4R:           ci.fps.JA4R,
		JA3:            ci.fps.JA3,
		JA3Str:         ci.fps.JA3Str,
		RTTMicros:      kernelRTT(ci.raw).Microseconds(),
		DurationMillis: now.Sub(ci.accepted).Milliseconds(),
		BytesIn:        ci.in.Load(),
		BytesOut:       ci.out.Load(),
		RequestCount:   ci.requests,
		IdempotencyKey: uuid4(),
	}
	if !ci.handshakeAt.IsZero() {
		rec.HandshakeMicros = ci.handshakeAt.Sub(ci.accepted).Microseconds()
	}
	if ci.h1 != nil {
		rec.HdrCaptureLost = ci.h1.CaptureLost()
	}
	if ci.h2 != nil {
		rec.H2FP = ci.h2.Fingerprint()
		rec.HdrCaptureLost = ci.h2.CaptureLost()
	}
	if f := ci.first; f != nil {
		rec.FirstHost, rec.FirstMethod, rec.FirstPath = f.host, f.method, f.path
		rec.FirstUserAgent, rec.FirstJA4H, rec.FirstHeaderNames = f.ua, f.ja4h, f.names
		rec.FirstStatus, rec.FirstRoute = f.status, f.route
	} else if method, target, ok := requestLine(ci.rawHead); ok {
		// Plain HTTP sent to the TLS port.
		rec.FirstMethod, rec.FirstPath = method, target
	}
	rec.truncate()
	return rec
}

// wireConn is the raw TCP connection: it counts bytes both ways and, on a
// TLS listener, tees the opening bytes into the ClientHello parser until a
// hello is complete (or clearly isn't coming). When it closes — however the
// connection ended — the connection's Record is emitted.
type wireConn struct {
	net.Conn
	ci        *connInfo
	recording bool
	buf       []byte
	closeOnce sync.Once
}

func (c *wireConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.ci.in.Add(int64(n))
	if n > 0 && c.recording {
		c.buf = append(c.buf, p[:n]...)
		h, perr := tlsfp.Parse(c.buf)
		switch {
		case perr == nil:
			fps := tlsfp.Compute(h)
			c.ci.mu.Lock()
			c.ci.hello, c.ci.fps = h, fps
			c.ci.mu.Unlock()
			c.recording, c.buf = false, nil
		case errors.Is(perr, tlsfp.ErrNeedMore) && len(c.buf) <= tlsfp.MaxHelloBytes:
		default:
			c.ci.mu.Lock()
			c.ci.helloErr = perr
			if errors.Is(perr, tlsfp.ErrNotTLS) {
				c.ci.rawHead = c.buf[:min(len(c.buf), 1024)]
			}
			c.ci.mu.Unlock()
			c.recording, c.buf = false, nil
		}
	}
	return n, err
}

func (c *wireConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.ci.out.Add(int64(n))
	return n, err
}

func (c *wireConn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		rec := c.ci.record(time.Now()) // before close: kernel RTT needs the socket
		err = c.Conn.Close()
		if c.ci.sink != nil {
			c.ci.sink.Push(rec)
		}
	})
	return err
}

// h1Conn is what the HTTP/1 server gets: the header-order tap, plus the
// connInfo for ConnContext.
type h1Conn struct {
	*httpfp.H1Conn
	ci *connInfo
}

// h2Conn is what the HTTP/2 server gets. ConnectionState lets x/net/http2
// apply its TLS checks and fill r.TLS, as it would for a *tls.Conn.
type h2Conn struct {
	*httpfp.H2Conn
	tc *tls.Conn
}

func (c *h2Conn) ConnectionState() tls.ConnectionState { return c.tc.ConnectionState() }

type ciKey struct{}

func withConnInfo(ctx context.Context, ci *connInfo) context.Context {
	return context.WithValue(ctx, ciKey{}, ci)
}

func connInfoFrom(ctx context.Context) *connInfo {
	ci, _ := ctx.Value(ciKey{}).(*connInfo)
	return ci
}

// chanListener feeds connections the edge has already accepted (and
// handshaken) to an http.Server.
type chanListener struct {
	ch   chan net.Conn
	done chan struct{}
	once sync.Once
	addr net.Addr
}

func newChanListener(addr net.Addr) *chanListener {
	return &chanListener{ch: make(chan net.Conn), done: make(chan struct{}), addr: addr}
}

func (l *chanListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.ch:
		return c, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *chanListener) push(c net.Conn) {
	select {
	case l.ch <- c:
	case <-l.done:
		c.Close()
	}
}

func (l *chanListener) Close() error {
	l.once.Do(func() { close(l.done) })
	return nil
}

func (l *chanListener) Addr() net.Addr { return l.addr }

// requestLine parses an HTTP request line off the start of b.
func requestLine(b []byte) (method, target string, ok bool) {
	line, _, found := strings.Cut(string(b), "\r\n")
	if !found {
		line, _, found = strings.Cut(line, "\n")
	}
	if !found {
		return "", "", false
	}
	parts := strings.Split(line, " ")
	if len(parts) != 3 || !strings.HasPrefix(parts[2], "HTTP/") {
		return "", "", false
	}
	for _, ch := range parts[0] {
		if ch < 'A' || ch > 'Z' {
			return "", "", false
		}
	}
	return parts[0], parts[1], true
}
