// Package meta carries what acpwb_edge knows about a request's client —
// filled in by the connection layer, read by the proxy to set upstream
// headers, annotated by the router with the route it chose.
package meta

import "context"

type Request struct {
	ClientIP string
	Host     string // nginx $host: lowercased, port stripped
	TLS      bool

	ClientProtocol string // "HTTP/1.1", "HTTP/2.0" (nginx $server_protocol naming)
	TLSProtocol    string // "TLSv1.3" (nginx $ssl_protocol naming)
	TLSCipher      string // OpenSSL name (nginx $ssl_cipher naming)

	JA4, JA3, JA4H, H2Fingerprint string

	RequestID string // forwarded upstream as X-Request-Id; EdgeRequest.request_id
	ClientID  string // the acpwb_cid cookie's client ID, if any

	TrafficClass string // classify.Class*; forwarded as X-Traffic-Class
	ASN          uint32 // the client IP's ASN, 0 if unknown; forwarded as X-Client-Asn

	Route string // set by the router: which location handled it
	Error string // set by the proxy when the upstream failed (edge.ErrUpstream…)
}

type key struct{}

func With(ctx context.Context, m *Request) context.Context {
	return context.WithValue(ctx, key{}, m)
}

// From returns the request's metadata, or an empty one if there is none.
func From(ctx context.Context) *Request {
	if m, ok := ctx.Value(key{}).(*Request); ok {
		return m
	}
	return &Request{}
}
