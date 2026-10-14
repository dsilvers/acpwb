// Package edge is acpwb_edge's connection layer: it accepts TCP itself,
// fingerprints each TLS ClientHello from the raw bytes, does the TLS
// handshake, and hands the connection to an HTTP/1.1 or HTTP/2 server
// depending on ALPN — with taps on the decrypted stream that recover header
// order (JA4H) and the HTTP/2 preface frames. Every request carries what
// was learned (package meta) to the router and proxy, and every connection
// produces one Record for the analytics queue when it closes.
package edge

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"

	"acpwb_edge/classify"
	"acpwb_edge/edgequeue"
	"acpwb_edge/httpfp"
	"acpwb_edge/meta"
	"acpwb_edge/page"
	"acpwb_edge/router"
)

// Sink receives records (edgequeue.Queue): one *Record per closed
// connection on edgequeue.ConnQueue, one *RequestRecord per request on
// edgequeue.RequestQueue.
type Sink interface {
	Push(key string, v any)
}

type Config struct {
	TLSConfig *tls.Config // certificates, versions; ALPN is set here
	TLS       http.Handler
	Plain     http.Handler
	Sink      Sink
	Live      Live // the dashboard's live stream; nil: none

	// CookieDomain gives the client cookie's Domain for a host, or false
	// for hosts that get none (router.Config.CookieDomain). Nil: no cookie.
	CookieDomain func(host string) (string, bool)
	// CookieSecret signs client cookies; empty means a random key per
	// process (dev — every restart issues new IDs).
	CookieSecret []byte
	// HTTPSBase gives the HTTPS URL (no path) that plain HTTP sent to the
	// TLS port for a Host is redirected to (router.Config.HTTPSBase).
	// Default https://acpwb.com.
	HTTPSBase func(hostport string) string

	// Timeouts, from production data (Oct 2026). Successful handshakes
	// run to 7.7s at p99.9 through slow residential proxies, and those
	// clients give up at ~8s themselves. HTTP/1.1 clients reuse a
	// connection within 2.4s at p99; HTTP/2 reuses have a long tail (p99
	// 50s) but a reconnect costs them one ~240ms handshake.
	// ASN looks up client networks for classification; nil means every
	// network is "unknown".
	ASN *classify.ASNDB

	HandshakeTimeout  time.Duration // default 8s
	ReadHeaderTimeout time.Duration // HTTP/1 request head, once started; default 5s
	IdleTimeout       time.Duration // keep-alive, both protocols; default 15s
	MaxBodyBytes      int64         // default 10 MiB
	Limits            Limits
}

// MaxHeaderBytes bounds a request's headers on both protocols (nginx's
// large_client_header_buffers allowed ~32 KB); it's also the h1 tap's limit.
const MaxHeaderBytes = 64 << 10

// maxFrameSize is the HTTP/2 SETTINGS_MAX_FRAME_SIZE we advertise: the
// spec's default rather than x/net's 1 MiB, so neither the server nor the
// h2 tap ever buffers more than this per connection for one frame.
const maxFrameSize = 16 << 10

type Server struct {
	cfg       Config
	tlsConfig *tls.Config

	h1      *http.Server
	h2      *http2.Server
	baseCtx context.Context
	cancel  context.CancelFunc

	mu        sync.Mutex
	listeners []net.Listener
	chans     []*chanListener
	closing   bool
	handshake sync.WaitGroup

	records *recordGate
	clients *clientIDs
}

func New(cfg Config) *Server {
	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 8 * time.Second
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 15 * time.Second
	}
	if cfg.MaxBodyBytes == 0 {
		cfg.MaxBodyBytes = 10 << 20
	}
	if cfg.HTTPSBase == nil {
		cfg.HTTPSBase = func(string) string { return "https://acpwb.com" }
	}
	cfg.Limits.defaults()
	s := &Server{
		cfg:     cfg,
		records: newRecordGate(cfg.Sink, cfg.Limits.RecordsPerIPMin),
		clients: newClientIDs(cfg.CookieSecret, cfg.CookieDomain),
	}
	s.baseCtx, s.cancel = context.WithCancel(context.Background())

	if cfg.TLSConfig != nil {
		h1Only := cfg.TLSConfig.Clone()
		h1Only.NextProtos = []string{"http/1.1"}
		s.tlsConfig = cfg.TLSConfig.Clone()
		s.tlsConfig.NextProtos = []string{"h2", "http/1.1"}
		// HTTP/2 requires TLS 1.2+ with an AEAD cipher, and x/net/http2
		// drops connections that negotiated it without one. Clients that
		// can't meet that (old bots offering h2 anyway) get HTTP/1.1, as
		// they did from nginx.
		s.tlsConfig.GetConfigForClient = func(h *tls.ClientHelloInfo) (*tls.Config, error) {
			if !h2Capable(h) {
				return h1Only, nil
			}
			return nil, nil
		}
	}
	s.h1 = &http.Server{
		Handler:           http.HandlerFunc(s.serveRequest),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			if hc, ok := c.(*h1Conn); ok {
				return withConnInfo(ctx, hc.ci)
			}
			return ctx
		},
		BaseContext: func(net.Listener) context.Context { return s.baseCtx },
		ErrorLog:    log.New(io.Discard, "", 0), // client garbage isn't news
	}
	s.h2 = &http2.Server{
		MaxConcurrentStreams: 100, // nginx's http2_max_concurrent_streams is 128
		IdleTimeout:          cfg.IdleTimeout,
		MaxReadFrameSize:     maxFrameSize,
	}
	// Registers h2's graceful GOAWAY with s.h1.Shutdown. Our h1 conns are
	// never *tls.Conn, so the TLSNextProto it installs is never used.
	if err := http2.ConfigureServer(s.h1, s.h2); err != nil {
		panic(err)
	}
	return s
}

// ServeTLS accepts on l (the :443 listener) until Shutdown.
func (s *Server) ServeTLS(l net.Listener) error {
	cl := s.track(l)
	if cl == nil {
		return net.ErrClosed
	}
	go s.h1.Serve(cl)
	return s.accept(l, func(c net.Conn) { s.handleTLS(c, cl) })
}

// ServePlain accepts plain HTTP on l (the :80 listener) until Shutdown.
func (s *Server) ServePlain(l net.Listener) error {
	cl := s.track(l)
	if cl == nil {
		return net.ErrClosed
	}
	go s.h1.Serve(cl)
	return s.accept(l, func(c net.Conn) { s.handlePlain(c, cl) })
}

func (s *Server) track(l net.Listener) *chanListener {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return nil
	}
	cl := newChanListener(l.Addr())
	s.listeners = append(s.listeners, l)
	s.chans = append(s.chans, cl)
	return cl
}

func (s *Server) accept(l net.Listener, handle func(net.Conn)) error {
	var backoff time.Duration
	for {
		c, err := l.Accept()
		if err != nil {
			s.mu.Lock()
			closing := s.closing
			s.mu.Unlock()
			if closing || errors.Is(err, net.ErrClosed) {
				return nil
			}
			// e.g. EMFILE: back off rather than spin.
			backoff = min(max(backoff*2, 5*time.Millisecond), time.Second)
			log.Printf("edge: accept: %v; retrying in %v", err, backoff)
			time.Sleep(backoff)
			continue
		}
		backoff = 0
		go handle(c)
	}
}

func (s *Server) handleTLS(raw net.Conn, cl *chanListener) {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		raw.Close()
		return
	}
	s.handshake.Add(1) // under mu: Shutdown's Wait never races an Add
	s.mu.Unlock()
	ci := s.newConn(raw, true)
	wc := &wireConn{Conn: raw, ci: ci, recording: true}
	tc := tls.Server(wc, s.tlsConfig)

	// A socket deadline rather than HandshakeContext keeps the whole
	// handshake on this goroutine (HandshakeContext closes the conn from
	// another one on cancel).
	raw.SetDeadline(time.Now().Add(s.cfg.HandshakeTimeout))
	handshakesInFlight.Add(1)
	err := tc.Handshake()
	handshakesInFlight.Add(-1)
	raw.SetDeadline(time.Time{})
	s.handshake.Done()
	if err != nil {
		ci.handshakeFailed(err)
		if head, ok := ci.plainHTTPOnTLS(); ok {
			s.redirectPlainOnTLS(wc, ci, head)
		}
		wc.Close()
		return
	}

	st := tc.ConnectionState()
	ci.handshakeDone(st)
	if st.NegotiatedProtocol == "h2" {
		hc := httpfp.NewH2Conn(tc)
		hc.SetRapidResetLimit(func() {
			ci.setCloseReason(CloseRapidReset)
			wc.Close() // the h2 server sees the read error and tears down
		})
		ci.mu.Lock()
		ci.h2 = hc
		ci.mu.Unlock()
		s.h2.ServeConn(&h2Conn{H2Conn: hc, tc: tc, ci: ci}, &http2.ServeConnOpts{
			Context:    withConnInfo(s.baseCtx, ci),
			BaseConfig: s.h1,
			Handler:    http.HandlerFunc(s.serveRequest),
		})
		return
	}
	hc := httpfp.NewH1Conn(tc)
	ci.mu.Lock()
	ci.h1 = hc
	ci.mu.Unlock()
	cl.push(&h1Conn{H1Conn: hc, ci: ci})
}

func (s *Server) handlePlain(raw net.Conn, cl *chanListener) {
	ci := s.newConn(raw, false)
	ci.outcome = OutcomePlain
	hc := httpfp.NewH1Conn(&wireConn{Conn: raw, ci: ci})
	ci.h1 = hc
	cl.push(&h1Conn{H1Conn: hc, ci: ci})
}

// redirectPlainOnTLS answers plain HTTP sent to the TLS port (e.g.
// http://acpwb.com:443/) with a redirect to HTTPS, where nginx sent a 400,
// and logs it as a request.
func (s *Server) redirectPlainOnTLS(wc *wireConn, ci *connInfo, raw []byte) {
	start := time.Now()
	head := httpfp.ParseHead(raw)
	target := head.Path
	if !strings.HasPrefix(target, "/") {
		target = "/" // absolute-form or junk: don't reflect it
	}
	loc := s.cfg.HTTPSBase(head.Authority) + target
	body := page.Render(http.StatusMovedPermanently, loc)
	var b strings.Builder
	b.WriteString("HTTP/1.1 301 Moved Permanently\r\nServer: " + page.ServerHeader +
		"\r\nContent-Type: text/html; charset=utf-8\r\nLocation: " + headerSafe(loc) +
		"\r\nContent-Length: " + strconv.Itoa(len(body)) + "\r\nConnection: close\r\n\r\n")
	if head.Method != http.MethodHead {
		b.Write(body)
	}
	n, _ := wc.Write([]byte(b.String()))
	rec := ci.unservedRecord(head, http.StatusMovedPermanently, ErrPlainOnTLS, "HTTP/1.1", start)
	rec.Seq = 1
	rec.Route = router.Redirect
	rec.ResponseBytes = int64(max(0, n-(b.Len()-len(body))))
	rec.DurationUS = time.Since(start).Microseconds()
	ci.pushRequest(rec)
}

// headerSafe drops anything that could end a header line.
func headerSafe(v string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, v)
}

// serveRequest is the handler both HTTP servers run: it attaches what the
// connection layer learned to the request, then hands it to the router.
func (s *Server) serveRequest(w http.ResponseWriter, r *http.Request) {
	ci := connInfoFrom(r.Context())
	if ci == nil {
		page.Error(w, http.StatusInternalServerError)
		return
	}
	if r.Context().Err() != nil {
		return // HTTP/2 stream already reset (rapid reset): do no work
	}
	ci.mu.Lock()
	h1, h2 := ci.h1, ci.h2
	tlsVersion, cipher := ci.tlsVersion, ci.cipher
	ci.mu.Unlock()

	var names []string
	var h2fp string
	switch {
	case h2 != nil:
		names = h2.Claim(r.Method, r.Host, r.RequestURI)
		h2fp = h2.Fingerprint()
	case h1 != nil:
		names = h1.NextHead()
	}
	fps := ci.fingerprints()
	host := router.HostOf(r.Host)
	if host == "" {
		host = "acpwb.com" // nginx falls back to the first server_name
	}
	m := &meta.Request{
		ClientIP:       ci.ip,
		Host:           host,
		TLS:            ci.tls,
		ClientProtocol: r.Proto,
		TLSProtocol:    tlsVersion,
		TLSCipher:      cipher,
		JA4:            fps.JA4,
		JA3:            fps.JA3,
		JA4H:           httpfp.JA4H(r, names),
		H2Fingerprint:  h2fp,
	}
	start := time.Now()
	seq := ci.startRequest()
	m.RequestID = uuid4()
	r = r.WithContext(meta.With(r.Context(), m))
	sw := &statusWriter{ResponseWriter: w, start: start}
	var issued bool
	if ci.tls {
		m.ClientID, issued = s.clients.identify(w, r, host)
	}
	cls := classify.Classify(classify.Input{
		UA:             r.UserAgent(),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		IP:             ci.ip,
		JA4:            fps.JA4,
		H2Fingerprint:  h2fp,
		CookieDropped:  issued && seq > 1,
		Net:            ci.network(),
	})
	m.TrafficClass, m.ASN = cls.Class, ci.network().ASN
	if r.Body != nil && r.Body != http.NoBody {
		r.Body = &countingBody{ReadCloser: r.Body}
	}
	body := r.Body
	defer func() {
		ci.noteRequest(r, m, names, sw.Status())
		rec := s.requestRecord(ci, r, m, sw, names, body, seq, issued, start)
		rec.setClass(fps.JA4, ci.network(), cls)
		ci.pushRequest(rec)
	}()

	w.Header().Set("Server", page.ServerHeader)
	release, ok := s.admitRequest()
	if !ok {
		rejectedRequests.Add(1)
		ci.noteRejectedRequest()
		m.Route, m.Error = "overload", ErrOverload
		sw.Header().Set("Retry-After", "1")
		page.Error(sw, http.StatusServiceUnavailable)
		return
	}
	defer release()
	if !isUpgrade(r) {
		// Stall timeouts (nginx client_body_timeout / send_timeout): the
		// deadlines move forward whenever the body or response progresses.
		rc := http.NewResponseController(w)
		stall := s.cfg.Limits.StallTimeout
		now := time.Now()
		_ = rc.SetReadDeadline(now.Add(stall))
		_ = rc.SetWriteDeadline(now.Add(stall))
		sw.rc, sw.stall, sw.last = rc, stall, now
		if r.Body != nil && r.Body != http.NoBody {
			r.Body = &stallReader{ReadCloser: r.Body, rc: rc, timeout: stall, last: now}
		}
	}
	if r.ContentLength > s.cfg.MaxBodyBytes {
		m.Error = ErrBodyTooLarge
		page.Error(sw, http.StatusRequestEntityTooLarge)
		return
	}
	r.Body = http.MaxBytesReader(sw, r.Body, s.cfg.MaxBodyBytes)

	h := s.cfg.Plain
	if ci.tls {
		h = s.cfg.TLS
	}
	h.ServeHTTP(sw, r)
}

// requestRecord describes a request a handler served, as it finishes.
func (s *Server) requestRecord(ci *connInfo, r *http.Request, m *meta.Request, sw *statusWriter,
	names []string, body io.ReadCloser, seq int, issued bool, start time.Time) *RequestRecord {
	path, query := splitURI(r.RequestURI)
	rec := &RequestRecord{
		Timestamp:      isoTime(start),
		RequestID:      m.RequestID,
		ConnID:         ci.id,
		ClientID:       m.ClientID,
		NewClient:      issued,
		Seq:            seq,
		IPAddress:      ci.ip,
		TLS:            ci.tls,
		Protocol:       protoName(r),
		Host:           m.Host,
		Method:         r.Method,
		Path:           path,
		Query:          query,
		UserAgent:      r.UserAgent(),
		Referer:        r.Referer(),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		JA4H:           m.JA4H,
		HeaderNames:    strings.Join(names, ","),
		Status:         sw.Status(),
		Route:          m.Route,
		Error:          m.Error,
		ContentType:    sw.Header().Get("Content-Type"),
		RequestBytes:   requestBytes(body),
		ResponseBytes:  sw.written,
		DurationUS:     time.Since(start).Microseconds(),
	}
	if !sw.headAt.IsZero() {
		rec.TTFBMicros = sw.headAt.Sub(start).Microseconds()
	}
	switch {
	case rec.Error != "":
	case r.Context().Err() != nil:
		// The client reset the stream or went away before the response
		// was done (nginx logs these as 499).
		rec.Error = ErrClientClosed
	case sw.writeErr:
		rec.Error = ErrWriteFailed
	}
	if rec.Status == 0 && rec.Error == ErrClientClosed {
		rec.Status = 499
	}
	rec.truncate()
	return rec
}

// statusWriter records the response status, when the head went out, and
// body bytes written. Unwrap keeps Flush/Hijack reachable through
// http.ResponseController (the proxy's WebSocket upgrade hijacks through
// it).
type statusWriter struct {
	http.ResponseWriter
	status   int
	start    time.Time
	headAt   time.Time
	written  int64
	writeErr bool

	rc    *http.ResponseController // set: extend the write deadline on progress
	stall time.Duration
	last  time.Time
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 || w.status < 200 {
		w.status = code
		if code >= 200 || code == http.StatusSwitchingProtocols {
			w.headAt = time.Now()
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
		w.headAt = time.Now()
	}
	n, err := w.ResponseWriter.Write(p)
	w.written += int64(n)
	if err != nil {
		w.writeErr = true
	}
	if n > 0 && w.rc != nil {
		// At most once a second: on HTTP/2 each call is a message to the
		// connection's serve loop.
		if now := time.Now(); now.Sub(w.last) > time.Second {
			w.last = now
			_ = w.rc.SetWriteDeadline(now.Add(w.stall))
		}
	}
	return n, err
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		if w.status == 0 {
			w.status = http.StatusOK
			w.headAt = time.Now()
		}
		f.Flush()
	}
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusWriter) Status() int { return w.status }

// Shutdown stops accepting, lets in-flight handshakes finish, and drains
// HTTP/1 and HTTP/2 connections gracefully until ctx is done.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.closing = true
	for _, l := range s.listeners {
		l.Close()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() { s.handshake.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}

	err := s.h1.Shutdown(ctx) // also closes the chanListeners, GOAWAYs h2
	s.mu.Lock()
	for _, cl := range s.chans {
		cl.Close()
	}
	s.mu.Unlock()
	s.cancel()
	s.records.close() // flush aggregated records
	return err
}

// newConn starts tracking an accepted connection.
func (s *Server) newConn(raw net.Conn, isTLS bool) *connInfo {
	ci := newConnInfo(raw, isTLS, s.cfg.ASN)
	sink, live := s.cfg.Sink, s.cfg.Live
	ci.emit = s.records.emit
	if live != nil {
		ci.emit = func(rec *Record) {
			if live.Active() {
				live.Conn(rec)
			}
			s.records.emit(rec)
		}
	}
	if sink != nil || live != nil {
		ci.emitReq = func(rec *RequestRecord) {
			if live != nil && live.Active() {
				live.Request(rec, ci.connTLS())
			}
			if sink != nil {
				requestsLogged.Add(1)
				sink.Push(edgequeue.RequestQueue, rec)
			}
		}
	}
	return ci
}

// h2Capable reports whether a client can meet HTTP/2's TLS requirements
// with us: TLS 1.3 (all AEAD), or TLS 1.2 with an ECDHE AEAD suite.
func h2Capable(h *tls.ClientHelloInfo) bool {
	for _, v := range h.SupportedVersions {
		if v == tls.VersionTLS13 {
			return true
		}
	}
	if !slices.Contains(h.SupportedVersions, tls.VersionTLS12) {
		return false
	}
	for _, c := range h.CipherSuites {
		switch c {
		case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
			return true
		}
	}
	return false
}

// isUpgrade reports a WebSocket (or other) protocol upgrade, whose hijacked
// connection must not inherit a request's deadlines.
func isUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") != "" && httpguts.HeaderValuesContainsToken(r.Header["Connection"], "upgrade")
}
