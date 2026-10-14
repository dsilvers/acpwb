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

	"acpwb_edge/classify"
	"acpwb_edge/httpfp"
	"acpwb_edge/meta"
	"acpwb_edge/tlsfp"
)

// connInfo is everything known about one client connection. It becomes a
// Record when the connection closes.
type connInfo struct {
	id       string // EdgeConnection.conn_id / EdgeRequest.conn_id
	ip       string
	accepted time.Time
	tls      bool
	raw      net.Conn             // the TCP socket, for kernel RTT
	emit     func(*Record)        // where the Record goes at close
	emitReq  func(*RequestRecord) // where request records go
	asnDB    *classify.ASNDB

	netOnce sync.Once
	netInfo classify.NetInfo

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
	requests    int // requests that reached a handler
	rejected    int // of those, refused by the edge's overload valve (503)
	unserved    int // requests read that no handler served (see closeRequests)
	clientID    string
	closeReason string
	first       *firstRequest

	// readAborted is set while the read deadline is in the past: net/http
	// sets one after every HTTP/1 request to cancel its background read,
	// and the timeout that read returns isn't the connection timing out.
	readAborted atomic.Bool
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
	if reason == "" || reason == CloseReadTimeout && ci.readAborted.Load() {
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
		"requests_logged":      requestsLogged.Load(),
		"requests_unserved":    requestsUnserved.Load(),
		"parser_panics":        parsePanics.Load(),
	}
}

func newConnInfo(raw net.Conn, isTLS bool, asnDB *classify.ASNDB) *connInfo {
	connsOpen.Add(1)
	connsTotal.Add(1)
	ci := &connInfo{id: uuid4(), accepted: time.Now(), tls: isTLS, raw: raw, asnDB: asnDB}
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

// plainHTTPOnTLS returns the opening bytes if the client sent an HTTP
// request line to the TLS port.
func (ci *connInfo) plainHTTPOnTLS() ([]byte, bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	_, _, ok := requestLine(ci.rawHead)
	return ci.rawHead, ok
}

// startRequest counts a request reaching the handler and returns its
// position on the connection (1 for the first).
func (ci *connInfo) startRequest() int {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.requests++
	return ci.requests
}

// noteRequest keeps a summary of the connection's first request, and its
// first client ID.
func (ci *connInfo) noteRequest(r *http.Request, m *meta.Request, names []string, status int) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	if ci.clientID == "" {
		ci.clientID = m.ClientID
	}
	if ci.first != nil {
		return
	}
	ci.first = &firstRequest{
		host: m.Host, method: r.Method, path: r.URL.RequestURI(), ua: r.UserAgent(),
		ja4h: m.JA4H, names: strings.Join(names, ","), route: m.Route, status: status,
	}
}

// network is the client IP's ASN and network kind, looked up once per
// connection.
func (ci *connInfo) network() classify.NetInfo {
	ci.netOnce.Do(func() { ci.netInfo = ci.asnDB.Lookup(ci.ip) })
	return ci.netInfo
}

// record builds the connection's Record at close.
func (ci *connInfo) record(now time.Time) *Record {
	n := ci.network()
	ci.mu.Lock()
	defer ci.mu.Unlock()
	rec := &Record{
		Timestamp:      isoTime(ci.accepted),
		IPAddress:      ci.ip,
		ASN:            n.ASN,
		Network:        n.Network,
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
		UnservedReqs:   ci.unserved,
		ConnID:         ci.id,
		ClientID:       ci.clientID,
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
		rec.FirstMethod, rec.FirstPath, rec.FirstStatus = method, target, http.StatusMovedPermanently
		rec.FirstRoute = ErrPlainOnTLS
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
				c.ci.rawHead = c.buf[:min(len(c.buf), 4096)]
			}
			c.ci.mu.Unlock()
			c.recording, c.buf = false, nil
		}
	}
	return n, err
}

// SetReadDeadline and SetDeadline see every deadline set above — tls.Conn
// passes them down — so they track net/http's aborted reads.
func (c *wireConn) SetReadDeadline(t time.Time) error {
	c.ci.readAborted.Store(!t.IsZero() && t.Before(time.Now()))
	return c.Conn.SetReadDeadline(t)
}

func (c *wireConn) SetDeadline(t time.Time) error {
	c.ci.readAborted.Store(!t.IsZero() && t.Before(time.Now()))
	return c.Conn.SetDeadline(t)
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
		c.ci.closeRequests()
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

var (
	parsePanics      atomic.Int64
	requestsUnserved atomic.Int64
)

// closeRequests logs the requests the taps read that no handler served:
// HTTP/1 requests the server rejected (it answers those itself — 400, 431,
// 501 — and closes) or whose head never completed, and HTTP/2 streams the
// server refused or the client reset before a handler ran. Runs as the
// connection closes, when the server is done reading.
func (ci *connInfo) closeRequests() {
	ci.mu.Lock()
	h1, h2 := ci.h1, ci.h2
	ci.mu.Unlock()
	now := time.Now()
	var recs []*RequestRecord
	switch {
	case h2 != nil:
		for _, h := range h2.Unserved() {
			recs = append(recs, ci.unservedRecord(h, 0, ErrNotServed, "HTTP/2.0", now))
		}
	case h1 != nil:
		heads, status := h1.Unserved()
		for i, h := range heads {
			kind, st := ErrRejected, 0
			if h.Partial {
				kind = ErrIncomplete
			}
			if i == len(heads)-1 && status >= 400 {
				st = status // the server's reply to the request it gave up on
			}
			recs = append(recs, ci.unservedRecord(h, st, kind, "HTTP/1.1", now))
		}
	}
	if len(recs) == 0 {
		return
	}
	requestsUnserved.Add(int64(len(recs)))
	ci.mu.Lock()
	seq := ci.requests + ci.unserved
	ci.unserved += len(recs)
	ci.mu.Unlock()
	for i, rec := range recs {
		if i >= maxUnservedLogged {
			break
		}
		rec.Seq = seq + i + 1
		ci.pushRequest(rec)
	}
}

func (ci *connInfo) pushRequest(rec *RequestRecord) {
	if ci.emitReq != nil {
		ci.emitReq(rec)
	}
}

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
