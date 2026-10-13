package edge

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
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
	raw      net.Conn      // the TCP socket, for kernel RTT
	emit     func(*Record) // where the Record goes at close

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
	rejected    int // requests refused by the edge's limits (503)
	closeReason string
	first       *firstRequest
}

// Close reasons — why a connection ended, as far as the edge can tell:
// the first error seen on the socket, or the edge closing it.
const (
	CloseClient       = "client_closed"  // clean EOF from the client
	CloseClientReset  = "client_reset"   // RST / broken pipe
	CloseReadTimeout  = "read_timeout"   // handshake, header, idle or body stall
	CloseWriteTimeout = "write_timeout"  // client stopped reading the response
	CloseReadError    = "read_error"     // other read failure
	CloseWriteError   = "write_error"    // other write failure
	CloseServer       = "server_closed"  // the edge closed it (protocol error, GOAWAY, idle, shutdown)
	CloseRapidReset   = "h2_rapid_reset" // the edge ended an HTTP/2 rapid-reset attack
)

func classifyErr(err error, write bool) string {
	var ne net.Error
	switch {
	case errors.Is(err, net.ErrClosed):
		return "" // our own Close: not the cause
	case errors.Is(err, io.EOF) && !write:
		return CloseClient
	case errors.As(err, &ne) && ne.Timeout():
		if write {
			return CloseWriteTimeout
		}
		return CloseReadTimeout
	case errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE):
		return CloseClientReset
	case write:
		return CloseWriteError
	}
	return CloseReadError
}

func (ci *connInfo) setCloseReason(reason string) {
	ci.mu.Lock()
	if ci.closeReason == "" {
		ci.closeReason = reason
	}
	ci.mu.Unlock()
}

func (ci *connInfo) noteErr(err error, write bool) {
	reason := classifyErr(err, write)
	if reason == "" {
		return
	}
	ci.mu.Lock()
	if ci.closeReason == "" {
		ci.closeReason = reason
	}
	ci.mu.Unlock()
}

func (ci *connInfo) noteRejectedRequest() {
	ci.mu.Lock()
	ci.rejected++
	ci.mu.Unlock()
}

type firstRequest struct {
	host, method, path, ua, ja4h, names, route string
	status                                     int
}

// Process-wide counters, for the debug listener and the dashboard (see
// Stats).
var (
	connsOpen, connsTotal, handshakesInFlight atomic.Int64
)

// Stats reports the edge's live counters.
func Stats() map[string]int64 {
	return map[string]int64{
		"conns_open":           connsOpen.Load(),
		"conns_total":          connsTotal.Load(),
		"handshakes_in_flight": handshakesInFlight.Load(),
		"requests_in_flight":   inFlight.Load(),
		"rejected_requests":    rejectedRequests.Load(),
		"records_aggregated":   recordsAggregated.Load(),
		"parser_panics":        parsePanics.Load(),
	}
}

func newConnInfo(raw net.Conn, isTLS bool) *connInfo {
	connsOpen.Add(1)
	connsTotal.Add(1)
	ci := &connInfo{accepted: time.Now(), tls: isTLS, raw: raw}
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
		RejectedReqs:   ci.rejected,
		CloseReason:    ci.closeReason,
		ConnCount:      1,
		IdempotencyKey: uuid4(),
	}
	if rec.CloseReason == "" {
		rec.CloseReason = CloseServer
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
		st := ci.h2.Counters()
		rec.H2Streams, rec.H2Resets, rec.H2Continuations, rec.H2Goaway = st.Streams, st.Resets, st.Continuations, st.Goaway
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
	if err != nil {
		c.ci.noteErr(err, false)
	}
	if n > 0 && c.recording {
		c.buf = append(c.buf, p[:n]...)
		h, fps, perr := safeParse(c.buf)
		switch {
		case perr == nil:
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
	if err != nil {
		c.ci.noteErr(err, true)
	}
	return n, err
}

func (c *wireConn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		connsOpen.Add(-1)
		rec := c.ci.record(time.Now()) // before close: kernel RTT needs the socket
		err = c.Conn.Close()
		if c.ci.emit != nil {
			c.ci.emit(rec)
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

// Read also sees TLS-level errors: a client's clean close_notify surfaces
// here as io.EOF, never on the raw socket.
func (c *h1Conn) Read(p []byte) (int, error) {
	n, err := c.H1Conn.Read(p)
	if err != nil {
		c.ci.noteErr(err, false)
	}
	return n, err
}

// h2Conn is what the HTTP/2 server gets. ConnectionState lets x/net/http2
// apply its TLS checks and fill r.TLS, as it would for a *tls.Conn.
type h2Conn struct {
	*httpfp.H2Conn
	tc *tls.Conn
	ci *connInfo
}

func (c *h2Conn) Read(p []byte) (int, error) {
	n, err := c.H2Conn.Read(p)
	if err != nil {
		c.ci.noteErr(err, false)
	}
	return n, err
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

var parsePanics atomic.Int64

// safeParse parses and fingerprints client-controlled bytes; a bug in
// either must cost one connection its fingerprint, never the process.
func safeParse(raw []byte) (h *tlsfp.Hello, fps tlsfp.Fingerprints, err error) {
	defer func() {
		if r := recover(); r != nil {
			h, fps, err = nil, tlsfp.Fingerprints{}, tlsfp.ErrMalformed
			if n := parsePanics.Add(1); n&(n-1) == 0 {
				log.Printf("edge: recovered ClientHello parser panic #%d: %v\n%s", n, r, debug.Stack())
			}
		}
	}()
	if h, err = tlsfp.Parse(raw); err == nil {
		fps = tlsfp.Compute(h)
	}
	return h, fps, err
}
