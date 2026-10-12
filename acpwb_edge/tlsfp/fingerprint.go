package tlsfp

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Fingerprints is everything computed from one ClientHello.
type Fingerprints struct {
	JA4    string // FoxIO JA4, e.g. t13d1516h2_8daaf6152771_e5627efa2ab1
	JA4R   string // JA4_r: the same, with the sorted raw lists instead of hashes
	JA3    string // MD5 of JA3Str
	JA3Str string // SSLVersion,Ciphers,Extensions,Groups,PointFormats
}

// Compute returns all fingerprints for h. The transport is always TCP:
// acpwb_edge doesn't speak QUIC.
func Compute(h *Hello) Fingerprints {
	versions := h.SupportedVersions
	if len(withoutGREASE(versions)) == 0 {
		versions = []uint16{h.LegacyVersion}
	}
	a, b, c := ja4Parts(versions, h.ServerName != "", h.CipherSuites, h.Extensions, h.SignatureAlgorithms, h.ALPN)
	ja3 := ja3String(h)
	sum := md5.Sum([]byte(ja3))
	return Fingerprints{
		JA4:    a + "_" + hash12(b) + "_" + hash12(c),
		JA4R:   a + "_" + b + "_" + c,
		JA3:    hex.EncodeToString(sum[:]),
		JA3Str: ja3,
	}
}

// JA4 per https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/JA4.md
// — ported from the openbeverage.com demo, which is tested against the
// spec's worked example.

const emptyHash = "000000000000"

func isGREASE(v uint16) bool {
	return v&0x0f0f == 0x0a0a && v>>8 == v&0xff
}

func hash12(s string) string {
	if s == "" {
		return emptyHash
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

func hexList(vs []uint16) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = fmt.Sprintf("%04x", v)
	}
	return strings.Join(parts, ",")
}

func withoutGREASE(vs []uint16) []uint16 {
	out := make([]uint16, 0, len(vs))
	for _, v := range vs {
		if !isGREASE(v) {
			out = append(out, v)
		}
	}
	return out
}

func tlsVersionCode(versions []uint16) string {
	var highest uint16
	for _, v := range withoutGREASE(versions) {
		highest = max(highest, v)
	}
	switch highest {
	case 0x0304:
		return "13"
	case 0x0303:
		return "12"
	case 0x0302:
		return "11"
	case 0x0301:
		return "10"
	case 0x0300:
		return "s3"
	case 0x0002:
		return "s2"
	}
	return "00"
}

func isAlnum(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func alpnCode(protos []string) string {
	if len(protos) == 0 || protos[0] == "" {
		return "00"
	}
	p := protos[0]
	first, last := p[0], p[len(p)-1]
	if isAlnum(first) && isAlnum(last) {
		return string([]byte{first, last})
	}
	h := hex.EncodeToString([]byte(p))
	return string([]byte{h[0], h[len(h)-1]})
}

// TwoDigits is JA4's capped two-digit count; JA4H uses it too.
func TwoDigits(n int) string {
	return fmt.Sprintf("%02d", min(n, 99))
}

// Hash12 is JA4's truncated SHA-256; JA4H uses it too.
func Hash12(s string) string { return hash12(s) }

// ja4From is the demo's entry point, kept for the spec-vector test.
func ja4From(versions []uint16, sni bool, ciphers, exts, sigs []uint16, alpn []string) string {
	a, b, c := ja4Parts(versions, sni, ciphers, exts, sigs, alpn)
	return a + "_" + hash12(b) + "_" + hash12(c)
}

// ja4Parts returns JA4's a section and the unhashed b and c sections —
// JA4 hashes b and c, JA4_r writes them out as-is.
func ja4Parts(versions []uint16, sni bool, ciphers, exts, sigs []uint16, alpn []string) (a, b, c string) {
	ciphers = withoutGREASE(ciphers)
	exts = withoutGREASE(exts)
	sigs = withoutGREASE(sigs)

	sniCode := "i"
	if sni {
		sniCode = "d"
	}
	a = "t" + tlsVersionCode(versions) + sniCode + TwoDigits(len(ciphers)) + TwoDigits(len(exts)) + alpnCode(alpn)

	b = hexList(slices.Sorted(slices.Values(ciphers)))

	var hashedExts []uint16
	for _, e := range exts {
		if e != extServerName && e != extALPN { // counted but not hashed
			hashedExts = append(hashedExts, e)
		}
	}
	slices.Sort(hashedExts)
	c = hexList(hashedExts)
	if c != "" && len(sigs) > 0 {
		c += "_" + hexList(sigs)
	}
	return a, b, c
}

// ja3String is the classic JA3 input: decimal values, dash-joined, GREASE
// removed, ciphers and extensions in wire order.
func ja3String(h *Hello) string {
	dec := func(vs []uint16) string {
		parts := make([]string, 0, len(vs))
		for _, v := range withoutGREASE(vs) {
			parts = append(parts, strconv.Itoa(int(v)))
		}
		return strings.Join(parts, "-")
	}
	points := make([]string, len(h.ECPointFormats))
	for i, p := range h.ECPointFormats {
		points[i] = strconv.Itoa(int(p))
	}
	return strings.Join([]string{
		strconv.Itoa(int(h.LegacyVersion)),
		dec(h.CipherSuites),
		dec(h.Extensions),
		dec(h.SupportedGroups),
		strings.Join(points, "-"),
	}, ",")
}
