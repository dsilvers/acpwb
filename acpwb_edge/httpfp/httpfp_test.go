package httpfp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http/httptest"
	"slices"
	"testing"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// ── HTTP/1 header order (ported from the demo) ───────────────────────────────

func feedH1(chunks ...string) *H1Conn {
	c := &H1Conn{}
	for _, ch := range chunks {
		c.scan([]byte(ch))
	}
	return c
}

func TestH1HeaderOrderSplitReads(t *testing.T) {
	c := feedH1("GET / HTTP/1.1\r\nHost: x\r\nUser-Ag", "ent: curl\r\nAccept: */*\r", "\n\r\n")
	if got := c.NextHead(); !slices.Equal(got, []string{"Host", "User-Agent", "Accept"}) {
		t.Fatalf("got %v", got)
	}
}

func TestH1HeaderOrderKeepAlive(t *testing.T) {
	c := feedH1("GET /a/ HTTP/1.1\r\nHost: x\r\n\r\nGET /b/ HTTP/1.1\r\nAccept: */*\r\nHost: x\r\n\r\n")
	if got := c.NextHead(); !slices.Equal(got, []string{"Host"}) {
		t.Fatalf("first: got %v", got)
	}
	if got := c.NextHead(); !slices.Equal(got, []string{"Accept", "Host"}) {
		t.Fatalf("second: got %v", got)
	}
	if got := c.NextHead(); got != nil {
		t.Fatalf("expected empty queue, got %v", got)
	}
}

func TestH1HeaderOrderSkipsBody(t *testing.T) {
	c := feedH1("POST / HTTP/1.1\r\nHost: x\r\nContent-Length: 20\r\n\r\nGET / HTTP/1.1\r\nX: 1\r\n\r\n", "\r\n\r\nGET / HTTP/1.1\r\nY: 2\r\n\r\n")
	for _, w := range [][]string{{"Host", "Content-Length"}, {"Y"}} {
		if got := c.NextHead(); !slices.Equal(got, w) {
			t.Fatalf("got %v, want %v", got, w)
		}
	}
}

func TestH1ChunkedBodyStopsCapture(t *testing.T) {
	c := feedH1("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n0\r\n\r\nGET / HTTP/1.1\r\nZ: 1\r\n\r\n")
	if got := c.NextHead(); !slices.Equal(got, []string{"Transfer-Encoding"}) {
		t.Fatalf("got %v", got)
	}
	if !c.CaptureLost() || c.NextHead() != nil {
		t.Fatal("capture should stop after a chunked body")
	}
}

func TestH1HeaderOrderOverRealConn(t *testing.T) {
	client, server := net.Pipe()
	hc := NewH1Conn(server)
	go func() {
		client.Write([]byte("GET / HTTP/1.1\r\nB: 1\r\nA: 2\r\n\r\n"))
		client.Close()
	}()
	buf := make([]byte, 7)
	for {
		if _, err := hc.Read(buf); err != nil {
			break
		}
	}
	if got := hc.NextHead(); !slices.Equal(got, []string{"B", "A"}) {
		t.Fatalf("got %v", got)
	}
}

// ── JA4H ─────────────────────────────────────────────────────────────────────

func TestJA4H(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Language", "en-US,en;q=0.9")
	r.Header.Set("Cookie", "b=2; a=1")
	names := []string{"Host", "User-Agent", "Accept", "Accept-Language", "Cookie", "Referer"}
	h := func(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:])[:12] }
	want := "ge11cr04enus_" + h("Host,User-Agent,Accept,Accept-Language") + "_" + h("a,b") + "_" + h("a=1,b=2")
	if got := JA4H(r, names); got != want {
		t.Fatalf("JA4H = %s, want %s", got, want)
	}

	r2 := httptest.NewRequest("GET", "/", nil)
	r2.ProtoMajor, r2.ProtoMinor = 2, 0
	if got := JA4H(r2, []string{":method", ":authority", ":scheme", ":path", "accept"}); got[:12] != "ge20nn010000" {
		t.Fatalf("h2 JA4H prefix = %s (pseudo-headers must not count)", got[:12])
	}
	if JA4H(r, nil) != "" {
		t.Fatal("nil names (capture lost) must give an empty JA4H")
	}
}

// ── HTTP/2 tap ───────────────────────────────────────────────────────────────

type readerConn struct {
	net.Conn
	io.Reader
}

func (c readerConn) Read(p []byte) (int, error) { return c.Reader.Read(p) }

func headerBlock(enc *hpack.Encoder, buf *bytes.Buffer, fields ...string) []byte {
	buf.Reset()
	for i := 0; i < len(fields); i += 2 {
		enc.WriteField(hpack.HeaderField{Name: fields[i], Value: fields[i+1]})
	}
	return slices.Clone(buf.Bytes())
}

func TestH2FingerprintAndClaim(t *testing.T) {
	var wire bytes.Buffer
	wire.WriteString(http2.ClientPreface)
	fr := http2.NewFramer(&wire, nil)
	// Chrome-like opening.
	fr.WriteSettings(
		http2.Setting{ID: http2.SettingHeaderTableSize, Val: 65536},
		http2.Setting{ID: http2.SettingEnablePush, Val: 0},
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: 6291456},
		http2.Setting{ID: http2.SettingMaxHeaderListSize, Val: 262144},
	)
	fr.WriteWindowUpdate(0, 15663105)
	fr.WritePriority(3, http2.PriorityParam{StreamDep: 0, Exclusive: false, Weight: 200})

	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	block1 := headerBlock(enc, &hb,
		":method", "GET", ":authority", "acpwb.com", ":scheme", "https", ":path", "/wiki/x/",
		"user-agent", "Mozilla/5.0", "accept", "*/*", "accept-language", "en-US")
	// Split the first block across HEADERS + CONTINUATION, with padding.
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, BlockFragment: block1[:10], EndStream: true, PadLength: 3})
	fr.WriteContinuation(1, true, block1[10:])
	// Second request reuses dynamic-table entries from the first, so the
	// tap's decoder has to have stayed in step.
	block2 := headerBlock(enc, &hb,
		":method", "GET", ":authority", "acpwb.com", ":scheme", "https", ":path", "/reports/",
		"user-agent", "Mozilla/5.0", "accept", "*/*")
	fr.WriteData(1, false, bytes.Repeat([]byte("x"), 5000)) // skipped, not buffered
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 3, BlockFragment: block2, EndStream: true, EndHeaders: true})

	c := NewH2Conn(readerConn{Reader: &wire})
	buf := make([]byte, 13)
	for {
		if _, err := c.Read(buf); err != nil {
			break
		}
	}
	if c.CaptureLost() {
		t.Fatal("capture lost")
	}
	if got, want := c.Fingerprint(), "1:65536;2:0;4:6291456;6:262144|15663105|3:0:0:201|m,a,s,p"; got != want {
		t.Fatalf("fingerprint\n got %s\nwant %s", got, want)
	}

	if got := c.Claim("GET", "acpwb.com", "/reports/"); !slices.Equal(got,
		[]string{":method", ":authority", ":scheme", ":path", "user-agent", "accept"}) {
		t.Fatalf("stream 3 names: %v", got)
	}
	if got := c.Claim("GET", "ACPWB.com", "/wiki/x/"); len(got) != 7 || got[6] != "accept-language" {
		t.Fatalf("stream 1 names: %v", got)
	}
	if c.Claim("GET", "acpwb.com", "/wiki/x/") != nil {
		t.Fatal("a stream must only be claimable once")
	}
}

func TestH2NoWindowUpdateOrPriority(t *testing.T) {
	var wire bytes.Buffer
	wire.WriteString(http2.ClientPreface)
	fr := http2.NewFramer(&wire, nil)
	fr.WriteSettings(http2.Setting{ID: http2.SettingInitialWindowSize, Val: 65535})
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, EndHeaders: true, BlockFragment: headerBlock(enc, &hb,
		":method", "GET", ":path", "/", ":scheme", "https", ":authority", "x")})
	c := NewH2Conn(readerConn{Reader: &wire})
	io.Copy(io.Discard, c)
	if got, want := c.Fingerprint(), "4:65535|00|0|m,p,s,a"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestH2BadPrefaceGivesUp(t *testing.T) {
	c := NewH2Conn(readerConn{Reader: bytes.NewReader([]byte("GET / HTTP/1.1\r\n\r\n"))})
	io.Copy(io.Discard, c)
	if !c.CaptureLost() || c.Fingerprint() != "" {
		t.Fatal("non-h2 bytes must stop capture")
	}
}
