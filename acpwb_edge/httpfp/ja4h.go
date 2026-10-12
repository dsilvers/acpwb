// Package httpfp captures what HTTP fingerprints need but net/http throws
// away — request header order (http.Header is a map) and HTTP/2 connection
// preface frames — by watching the decrypted bytes a connection reads, and
// computes JA4H and the Akamai-style HTTP/2 fingerprint from them.
package httpfp

import (
	"net/http"
	"slices"
	"strings"

	"acpwb_edge/tlsfp"
)

// JA4H computes the JA4H HTTP client fingerprint, per
// https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/JA4H.md, from
// r and its header names in wire order. HTTP/2 pseudo-headers are skipped.
// Returns "" if names is nil (capture was lost for this request).
//
// Ported from the openbeverage.com demo.
func JA4H(r *http.Request, names []string) string {
	if names == nil {
		return ""
	}
	method := strings.ToLower(r.Method)
	if len(method) < 2 {
		method += "0"
	}
	version := "11"
	switch {
	case r.ProtoMajor == 1 && r.ProtoMinor == 0:
		version = "10"
	case r.ProtoMajor == 2:
		version = "20"
	}

	var counted []string
	hasCookie, hasReferer := "n", "n"
	for _, n := range names {
		if strings.HasPrefix(n, ":") {
			continue
		}
		switch strings.ToLower(n) {
		case "cookie":
			hasCookie = "c"
		case "referer":
			hasReferer = "r"
		default:
			counted = append(counted, n)
		}
	}

	lang := strings.ToLower(strings.ReplaceAll(r.Header.Get("Accept-Language"), "-", ""))
	lang, _, _ = strings.Cut(lang, ",")
	lang, _, _ = strings.Cut(lang, ";")
	lang = (lang + "0000")[:4]

	var cookieNames, cookiePairs []string
	for _, ck := range r.Cookies() {
		cookieNames = append(cookieNames, ck.Name)
		cookiePairs = append(cookiePairs, ck.Name+"="+ck.Value)
	}
	slices.Sort(cookieNames)
	slices.Sort(cookiePairs)

	a := method[:2] + version + hasCookie + hasReferer + tlsfp.TwoDigits(len(counted)) + lang
	return a + "_" + tlsfp.Hash12(strings.Join(counted, ",")) + "_" +
		tlsfp.Hash12(strings.Join(cookieNames, ",")) + "_" + tlsfp.Hash12(strings.Join(cookiePairs, ","))
}
