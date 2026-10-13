package edge

// Live receives a copy of every record for the dashboard's live request
// stream (livefeed.Feed). Unlike Sink it sees every connection before
// recordGate aggregates them, and each request with its connection's TLS
// details attached, since a RequestRecord only links to those by conn_id.
//
// Callers check Active first, so nothing is copied or snapshotted while
// nobody is watching. Request and Conn must not block or keep the pointers
// they're given.
type Live interface {
	Active() bool
	Request(rec *RequestRecord, tls ConnTLS)
	Conn(rec *Record)
}

// ConnTLS is what a request's connection negotiated, as far as the edge
// knows when the request finishes.
type ConnTLS struct {
	SNI     string `json:"sni"`
	ALPN    string `json:"alpn"`
	Version string `json:"tls_version"`
	Cipher  string `json:"tls_cipher"`
	JA4     string `json:"ja4"`
	JA3     string `json:"ja3"`
	H2FP    string `json:"h2_fingerprint"`
}

func (ci *connInfo) connTLS() ConnTLS {
	ci.mu.Lock()
	t := ConnTLS{
		SNI: ci.sni, ALPN: ci.alpn, Version: ci.tlsVersion, Cipher: ci.cipher,
		JA4: ci.fps.JA4, JA3: ci.fps.JA3,
	}
	h2 := ci.h2
	ci.mu.Unlock()
	if h2 != nil {
		t.H2FP = clip(h2.Fingerprint(), 512)
	}
	return t
}
