package edge

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"time"
)

// Record is one client connection, as pushed to acpwb:edge_conn_queue.
// Field names and lengths match honeypot.EdgeConnection — Django's drain
// builds rows straight from these keys, and an over-long value would fail
// the insert, so everything is truncated to the column size here.
type Record struct {
	Timestamp string `json:"timestamp"`
	IPAddress string `json:"ip_address"`
	SNI       string `json:"sni"`
	ALPN      string `json:"alpn"`
	TLSVer    string `json:"tls_version"`
	TLSCipher string `json:"tls_cipher"`
	Outcome   string `json:"outcome"`

	JA4    string `json:"ja4"`
	JA4R   string `json:"ja4_r"`
	JA3    string `json:"ja3"`
	JA3Str string `json:"ja3_str"`
	H2FP   string `json:"h2_fingerprint"`

	RTTMicros       int64  `json:"rtt_us"`
	HandshakeMicros int64  `json:"handshake_us"`
	DurationMillis  int64  `json:"duration_ms"`
	BytesIn         int64  `json:"bytes_in"`
	BytesOut        int64  `json:"bytes_out"`
	RequestCount    int    `json:"request_count"`     // requests that reached a handler
	RejectedReqs    int    `json:"rejected_requests"` // of those, refused by the overload valve
	UnservedReqs    int    `json:"unserved_requests"` // requests read that no handler served
	ConnID          string `json:"conn_id"`
	ClientID        string `json:"client_id"` // the first client ID seen on the connection
	HdrCaptureLost  bool   `json:"hdr_capture_lost"`
	CloseReason     string `json:"close_reason"`

	// HTTP/2 client behaviour: streams opened, RST_STREAMs sent (rapid
	// reset), CONTINUATION frames (CONTINUATION flood), and the error code
	// of any GOAWAY the server sent (ENHANCE_YOUR_CALM, PROTOCOL_ERROR…).
	H2Streams       int    `json:"h2_streams"`
	H2Resets        int    `json:"h2_resets"`
	H2Continuations int    `json:"h2_continuations"`
	H2Goaway        string `json:"h2_goaway"`

	// Connections this row stands for: 1, or more for an aggregated row
	// (requestless connections from an IP over the per-minute record
	// budget; see recordGate), whose other fields describe the first of
	// them as a sample.
	ConnCount  int  `json:"conn_count"`
	Aggregated bool `json:"aggregated"`

	FirstHost        string `json:"first_host"`
	FirstMethod      string `json:"first_method"`
	FirstPath        string `json:"first_path"`
	FirstUserAgent   string `json:"first_user_agent"`
	FirstJA4H        string `json:"first_ja4h"`
	FirstHeaderNames string `json:"first_header_names"`
	FirstStatus      int    `json:"first_status"`
	FirstRoute       string `json:"first_route"`

	IdempotencyKey string `json:"idempotency_key"`
}

// Outcomes — honeypot.EdgeConnection.OUTCOME_CHOICES.
const (
	OutcomeOK              = "ok"
	OutcomeHandshakeFailed = "handshake_failed"
	OutcomeNoHello         = "no_hello"
	OutcomeTimeout         = "timeout"
	OutcomePlain           = "plain"
)

func (r *Record) truncate() {
	t := func(s *string, n int) { *s = clip(*s, n) }
	t(&r.SNI, 253)
	t(&r.ALPN, 16)
	t(&r.TLSVer, 16)
	t(&r.TLSCipher, 64)
	t(&r.JA4, 40)
	t(&r.JA4R, 4096)
	t(&r.JA3, 32)
	t(&r.JA3Str, 4096)
	t(&r.H2FP, 512)
	t(&r.FirstHost, 253)
	t(&r.FirstMethod, 16)
	t(&r.FirstPath, 512)
	t(&r.FirstUserAgent, 512)
	t(&r.FirstJA4H, 64)
	t(&r.FirstHeaderNames, 4096)
	t(&r.FirstRoute, 32)
	t(&r.CloseReason, 32)
	t(&r.H2Goaway, 32)
}

func isoTime(t time.Time) string {
	// Parses with django.utils.dateparse.parse_datetime, same shape as
	// acpwb_go's visitqueue timestamps.
	return t.UTC().Format("2006-01-02T15:04:05.000000+00:00")
}

func uuid4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// tlsVersionName is nginx's $ssl_protocol naming, which reputation_scoring's
// _OLD_TLS_PROTOCOLS compares against.
func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLSv1"
	case tls.VersionTLS11:
		return "TLSv1.1"
	case tls.VersionTLS12:
		return "TLSv1.2"
	case tls.VersionTLS13:
		return "TLSv1.3"
	}
	return ""
}

// openSSLCipherNames maps the suites crypto/tls can negotiate to the
// OpenSSL names nginx's $ssl_cipher reported, so RequestFingerprint.
// tls_cipher stays comparable across the cutover. TLS 1.3 names are the
// same in both.
var openSSLCipherNames = map[uint16]string{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:       "ECDHE-ECDSA-AES128-GCM-SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:       "ECDHE-ECDSA-AES256-GCM-SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256: "ECDHE-ECDSA-CHACHA20-POLY1305",
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:         "ECDHE-RSA-AES128-GCM-SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:         "ECDHE-RSA-AES256-GCM-SHA384",
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:   "ECDHE-RSA-CHACHA20-POLY1305",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:          "ECDHE-ECDSA-AES128-SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:          "ECDHE-ECDSA-AES256-SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:            "ECDHE-RSA-AES128-SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:            "ECDHE-RSA-AES256-SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:       "ECDHE-ECDSA-AES128-SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:         "ECDHE-RSA-AES128-SHA256",
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256:               "AES128-GCM-SHA256",
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384:               "AES256-GCM-SHA384",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA:                  "AES128-SHA",
	tls.TLS_RSA_WITH_AES_256_CBC_SHA:                  "AES256-SHA",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256:               "AES128-SHA256",
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:           "ECDHE-RSA-DES-CBC3-SHA",
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:                 "DES-CBC3-SHA",
}

func cipherName(id uint16) string {
	if n, ok := openSSLCipherNames[id]; ok {
		return n
	}
	return tls.CipherSuiteName(id)
}
