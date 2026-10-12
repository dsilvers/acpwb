package proxy

import (
	"net/http"
	"testing"

	"acpwb_edge/meta"
)

func TestSetHeadersScrubsClientCopies(t *testing.T) {
	h := http.Header{}
	for _, k := range []string{
		"X-Real-Ip", "X-Forwarded-For", "X-Forwarded-Host", "Forwarded", "X-Client-Protocol",
		"X-Tls-Protocol", "X-Tls-Cipher", "X-Ja4", "X-Ja4h", "X-Ja4-R", "X-Ja3", "X-H2-Fingerprint",
	} {
		h.Set(k, "spoofed")
	}
	h["X_Ja4"] = []string{"spoofed"}
	h["x_real_ip"] = []string{"spoofed"}
	h.Set("Accept", "text/html")
	h.Set("X-Jasper", "kept") // not one of ours

	SetHeaders(h, &meta.Request{ClientIP: "203.0.113.7", TLS: true, ClientProtocol: "HTTP/2.0", JA4: "t13d_x_y"})

	for k, vs := range h {
		for _, v := range vs {
			if v == "spoofed" {
				t.Errorf("%s still carries the client's value", k)
			}
		}
	}
	want := map[string]string{
		"X-Real-Ip": "203.0.113.7", "X-Forwarded-For": "203.0.113.7", "X-Forwarded-Proto": "https",
		"X-Client-Protocol": "HTTP/2.0", "X-Ja4": "t13d_x_y", "Accept": "text/html", "X-Jasper": "kept",
	}
	for k, v := range want {
		if got := h.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	// Empty values are omitted, as nginx omits empty headers.
	for _, k := range []string{"X-Tls-Protocol", "X-Ja3", "X-H2-Fingerprint"} {
		if _, ok := h[k]; ok {
			t.Errorf("%s should be absent when empty", k)
		}
	}
}
