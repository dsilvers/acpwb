package edge

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"acpwb_edge/classify"
	"acpwb_edge/httpfp"
)

// RequestRecord is one HTTP request, as pushed to acpwb:edge_request_queue
// — every request the edge sees, static files and redirects included, and
// the ones that failed: rejected by the HTTP server before any handler ran
// (malformed, oversized), HTTP/2 streams never served, the client going
// away, upstream errors. Field names and lengths match honeypot.EdgeRequest.
//
// It links to its connection by conn_id (EdgeConnection.conn_id), to the
// browser by client_id (the acpwb_cid cookie), and to the app-side rows it
// produced — CrawlerVisit, ArchiveVisit, RequestFingerprint — by
// request_id, which the proxy forwards upstream as X-Request-Id.
type RequestRecord struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
	ConnID    string `json:"conn_id"`
	ClientID  string `json:"client_id"`
	NewClient bool   `json:"new_client"` // this response set the client cookie
	Seq       int    `json:"seq"`        // 1 for a connection's first request

	IPAddress string `json:"ip_address"`
	ASN       uint32 `json:"asn,omitempty"`
	Network   string `json:"network"`
	TLS       bool   `json:"tls"`
	Protocol  string `json:"protocol"` // HTTP/1.1, HTTP/2.0

	Host           string `json:"host"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	Query          string `json:"query"`
	UserAgent      string `json:"user_agent"`
	Referer        string `json:"referer"`
	AcceptLanguage string `json:"accept_language"`
	JA4H           string `json:"ja4h"`
	HeaderNames    string `json:"header_names"`

	Status        int    `json:"status"`
	Route         string `json:"route"`
	Error         string `json:"error"`
	ContentType   string `json:"content_type"`
	RequestBytes  int64  `json:"request_bytes"`  // body bytes read
	ResponseBytes int64  `json:"response_bytes"` // body bytes written (compressed, as sent)
	TTFBMicros    int64  `json:"ttfb_us"`        // until the response head was written
	DurationUS    int64  `json:"duration_us"`

	// What the client really is (package classify). Label and Group are
	// for the live view; EdgeRequest stores the rest.
	JA4         string `json:"ja4"`
	ClientClass string `json:"client_class"`
	Stack       string `json:"stack"`
	Signals     uint32 `json:"signals"`
	Label       string `json:"label"`
	Group       string `json:"group"`
}

func (r *RequestRecord) setClass(ja4 string, n classify.NetInfo, c classify.Result) {
	r.JA4, r.ASN, r.Network = ja4, n.ASN, n.Network
	r.ClientClass, r.Stack, r.Signals, r.Label, r.Group = c.Class, c.Stack, c.Signals, c.Label, c.Group
	r.JA4 = clip(r.JA4, 40)
	r.Stack = clip(r.Stack, 48)
}

// Request errors — honeypot.EdgeRequest.error. Empty for a request that
// was served, whatever its status.
const (
	ErrClientClosed    = "client_closed"    // client went away / reset the stream mid-request
	ErrWriteFailed     = "write_failed"     // response write failed or stalled
	ErrOverload        = "overload"         // the edge's overload valve answered 503
	ErrBodyTooLarge    = "body_too_large"   // 413
	ErrUpstream        = "upstream_error"   // 502
	ErrUpstreamTimeout = "upstream_timeout" // 504
	ErrRejected        = "rejected"         // HTTP/1 server refused it before any handler (400, 431, 501…)
	ErrIncomplete      = "incomplete"       // HTTP/1 head never completed before the connection ended
	ErrNotServed       = "h2_not_served"    // HTTP/2 stream refused by the server or reset by the client before a handler
	ErrPlainOnTLS      = "plain_on_tls"     // plain HTTP on the TLS port, answered with a redirect
)

// Unserved requests logged individually per connection; the rest are only
// counted (EdgeConnection.unserved_requests, and h2_resets for reset
// streams). A rapid-reset connection otherwise turns into a row per reset
// stream: 1,600 attack connections produced ~290k request records in 8s
// in testing.
const maxUnservedLogged = 20

var requestsLogged atomic.Int64

func (r *RequestRecord) truncate() {
	t := func(s *string, n int) { *s = clip(*s, n) }
	t(&r.Host, 253)
	t(&r.Method, 16)
	t(&r.Path, 1024)
	t(&r.Query, 1024)
	t(&r.UserAgent, 512)
	t(&r.Referer, 512)
	t(&r.AcceptLanguage, 128)
	t(&r.JA4H, 64)
	t(&r.HeaderNames, 2048)
	t(&r.Route, 32)
	t(&r.ContentType, 128)
	t(&r.Protocol, 16)
}

// clip truncates s to n bytes, without splitting a UTF-8 sequence, and
// drops NUL bytes (PostgreSQL text can't hold them; raw bytes of rejected
// requests can contain anything).
func clip(s string, n int) string {
	if strings.IndexByte(s, 0) >= 0 {
		s = strings.ReplaceAll(s, "\x00", "")
	}
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

func splitURI(uri string) (path, query string) {
	path, query, _ = strings.Cut(uri, "?")
	return path, query
}

// unservedRecord describes a request the tap saw but no handler served.
func (ci *connInfo) unservedRecord(h httpfp.Head, status int, errKind, proto string, at time.Time) *RequestRecord {
	path, query := splitURI(h.Path)
	rec := &RequestRecord{
		Timestamp:   isoTime(at),
		RequestID:   uuid4(),
		ConnID:      ci.id,
		IPAddress:   ci.ip,
		TLS:         ci.tls,
		Protocol:    proto,
		Host:        h.Authority,
		Method:      h.Method,
		Path:        path,
		Query:       query,
		UserAgent:   h.UserAgent,
		HeaderNames: strings.Join(h.Names, ","),
		Status:      status,
		Error:       errKind,
	}
	ja4 := ci.fingerprints().JA4
	rec.setClass(ja4, ci.network(), classify.Classify(classify.Input{UA: h.UserAgent, IP: ci.ip, JA4: ja4, Net: ci.network()}))
	rec.truncate()
	return rec
}

// countingBody counts request body bytes as the handler reads them.
type countingBody struct {
	io.ReadCloser
	n int64
}

func (b *countingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.n += int64(n)
	return n, err
}

func requestBytes(b io.ReadCloser) int64 {
	if cb, ok := b.(*countingBody); ok {
		return cb.n
	}
	return 0
}

func protoName(r *http.Request) string {
	if r.ProtoMajor == 2 {
		return "HTTP/2.0"
	}
	return r.Proto
}
