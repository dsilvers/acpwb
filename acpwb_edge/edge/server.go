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
	"sync"
	"time"

	"golang.org/x/net/http2"

	"acpwb_edge/httpfp"
	"acpwb_edge/meta"
	"acpwb_edge/page"
	"acpwb_edge/router"
)

// Sink receives one Record per closed connection (edgequeue.Queue).
type Sink interface {
	Push(v any)
}

type Config struct {
	TLSConfig *tls.Config // certificates, versions; ALPN is set here
	TLS       http.Handler
	Plain     http.Handler
	Sink      Sink

	HandshakeTimeout time.Duration // default 10s
	IdleTimeout      time.Duration // default 65s (nginx keepalive_timeout)
	MaxBodyBytes     int64         // default 10 MiB
}

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
}

func New(cfg Config) *Server {
	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 10 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 65 * time.Second
	}
	if cfg.MaxBodyBytes == 0 {
		cfg.MaxBodyBytes = 10 << 20
	}
	s := &Server{cfg: cfg}
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
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
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
		MaxConcurrentStreams: 250,
		IdleTimeout:          cfg.IdleTimeout,
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
	ci := newConnInfo(raw, true, s.cfg.Sink)
	wc := &wireConn{Conn: raw, ci: ci, recording: true}
	tc := tls.Server(wc, s.tlsConfig)

	// A socket deadline rather than HandshakeContext keeps the whole
	// handshake on this goroutine (HandshakeContext closes the conn from
	// another one on cancel).
	raw.SetDeadline(time.Now().Add(s.cfg.HandshakeTimeout))
	err := tc.Handshake()
	raw.SetDeadline(time.Time{})
	s.handshake.Done()
	if err != nil {
		ci.handshakeFailed(err)
		if ci.plainHTTPOnTLS() {
			wc.Write([]byte(plainHTTPOnTLSResponse))
		}
		wc.Close()
		return
	}

	st := tc.ConnectionState()
	ci.handshakeDone(st)
	if st.NegotiatedProtocol == "h2" {
		hc := httpfp.NewH2Conn(tc)
		ci.mu.Lock()
		ci.h2 = hc
		ci.mu.Unlock()
		s.h2.ServeConn(&h2Conn{H2Conn: hc, tc: tc}, &http2.ServeConnOpts{
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
	ci := newConnInfo(raw, false, s.cfg.Sink)
	ci.outcome = OutcomePlain
	hc := httpfp.NewH1Conn(&wireConn{Conn: raw, ci: ci})
	ci.h1 = hc
	cl.push(&h1Conn{H1Conn: hc, ci: ci})
}

// nginx's reply to plain HTTP on its TLS port.
const plainHTTPOnTLSResponse = "HTTP/1.1 400 Bad Request\r\nServer: nginx\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n" +
	"<html>\r\n<head><title>400 The plain HTTP request was sent to HTTPS port</title></head>\r\n<body>\r\n" +
	"<center><h1>400 Bad Request</h1></center>\r\n<center>The plain HTTP request was sent to HTTPS port</center>\r\n" +
	"<hr><center>nginx</center>\r\n</body>\r\n</html>\r\n"

// serveRequest is the handler both HTTP servers run: it attaches what the
// connection layer learned to the request, then hands it to the router.
func (s *Server) serveRequest(w http.ResponseWriter, r *http.Request) {
	ci := connInfoFrom(r.Context())
	if ci == nil {
		page.Error(w, http.StatusInternalServerError)
		return
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
	r = r.WithContext(meta.With(r.Context(), m))
	sw := &statusWriter{ResponseWriter: w}
	defer func() { ci.noteRequest(r, m, names, sw.Status()) }()

	w.Header().Set("Server", page.ServerHeader)
	if r.ContentLength > s.cfg.MaxBodyBytes {
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

// statusWriter records the response status. Unwrap keeps Flush/Hijack
// reachable through http.ResponseController (the proxy's WebSocket
// upgrade hijacks through it).
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 || w.status < 200 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		if w.status == 0 {
			w.status = http.StatusOK
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
	return err
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
