package tlsfp

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"regexp"
	"slices"
	"testing"
	"time"
)

// The worked example from the FoxIO JA4 technical details.
func TestJA4SpecExample(t *testing.T) {
	ciphers := []uint16{0x0a0a, 0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8, 0xc013, 0xc014, 0x009c, 0x009d, 0x002f, 0x0035}
	exts := []uint16{0x1a1a, 0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0010, 0x0005, 0x000d, 0x0012, 0x0033, 0x002d, 0x002b, 0x001b, 0x0015, 0x4469}
	sigs := []uint16{0x0403, 0x0804, 0x0401, 0x0503, 0x0805, 0x0501, 0x0806, 0x0601}
	got := ja4From([]uint16{0x2a2a, tls.VersionTLS13, tls.VersionTLS12}, true, ciphers, exts, sigs, []string{"h2", "http/1.1"})
	if want := "t13d1516h2_8daaf6152771_e5627efa2ab1"; got != want {
		t.Fatalf("ja4 = %s, want %s", got, want)
	}
}

// captureHello runs a real crypto/tls client handshake and returns the raw
// bytes it sent up to the end of its ClientHello.
func captureHello(t *testing.T, cfg *tls.Config) []byte {
	t.Helper()
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		_ = tls.Client(client, cfg).Handshake()
		client.Close()
	}()
	var raw []byte
	buf := make([]byte, 7) // small reads: exercises reassembly
	for {
		server.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := server.Read(buf)
		raw = append(raw, buf[:n]...)
		if _, perr := Parse(raw); !errors.Is(perr, ErrNeedMore) {
			if perr != nil {
				t.Fatalf("parse: %v", perr)
			}
			return raw
		}
		if err != nil {
			t.Fatalf("read: %v (after %d bytes)", err, len(raw))
		}
	}
}

// infoFor feeds raw to a crypto/tls server and returns the ClientHelloInfo
// crypto/tls itself parsed from it.
func infoFor(t *testing.T, raw []byte) *tls.ClientHelloInfo {
	t.Helper()
	var info *tls.ClientHelloInfo
	conn := &replayConn{Reader: bytes.NewReader(raw)}
	_ = tls.Server(conn, &tls.Config{GetConfigForClient: func(h *tls.ClientHelloInfo) (*tls.Config, error) {
		info = h
		return nil, errors.New("stop")
	}}).Handshake()
	if info == nil {
		t.Fatal("crypto/tls never parsed the hello")
	}
	return info
}

type replayConn struct {
	net.Conn
	io.Reader
}

func (c *replayConn) Read(p []byte) (int, error)       { return c.Reader.Read(p) }
func (c *replayConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *replayConn) Close() error                     { return nil }
func (c *replayConn) SetDeadline(time.Time) error      { return nil }
func (c *replayConn) SetReadDeadline(time.Time) error  { return nil }
func (c *replayConn) SetWriteDeadline(time.Time) error { return nil }
func (c *replayConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (c *replayConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }

// ja4FromInfo is the demo's ClientHelloInfo-based JA4.
func ja4FromInfo(h *tls.ClientHelloInfo) string {
	sigs := make([]uint16, len(h.SignatureSchemes))
	for i, s := range h.SignatureSchemes {
		sigs[i] = uint16(s)
	}
	return ja4From(h.SupportedVersions, h.ServerName != "", h.CipherSuites, h.Extensions, sigs, h.SupportedProtos)
}

func TestParseMatchesCryptoTLS(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  *tls.Config
		re   string
	}{
		{"tls13-h2-sni", &tls.Config{ServerName: "acpwb.com", NextProtos: []string{"h2", "http/1.1"}, InsecureSkipVerify: true}, `^t13d\d{4}h2_`},
		{"no-sni-no-alpn", &tls.Config{InsecureSkipVerify: true}, `^t13i\d{4}00_`},
		{"tls10-only", &tls.Config{ServerName: "acpwb.com", MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS10, InsecureSkipVerify: true}, `^t10d\d{4}00_`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := captureHello(t, tc.cfg)
			h, err := Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			fp := Compute(h)
			info := infoFor(t, raw)

			if want := ja4FromInfo(info); fp.JA4 != want {
				t.Errorf("JA4 = %s, crypto/tls-derived %s", fp.JA4, want)
			}
			if !regexp.MustCompile(tc.re).MatchString(fp.JA4) {
				t.Errorf("JA4 %s doesn't match %s", fp.JA4, tc.re)
			}
			if h.ServerName != info.ServerName {
				t.Errorf("SNI %q vs %q", h.ServerName, info.ServerName)
			}
			if !slices.Equal(h.Extensions, info.Extensions) {
				t.Errorf("extension order %v vs %v", h.Extensions, info.Extensions)
			}
			if !slices.Equal(h.ALPN, info.SupportedProtos) {
				t.Errorf("ALPN %v vs %v", h.ALPN, info.SupportedProtos)
			}
			sum := md5.Sum([]byte(fp.JA3Str))
			if fp.JA3 != hex.EncodeToString(sum[:]) {
				t.Errorf("JA3 isn't md5(JA3Str)")
			}
			if !regexp.MustCompile(`^\d+,[\d-]+,[\d-]+,[\d-]*,[\d-]*$`).MatchString(fp.JA3Str) {
				t.Errorf("JA3Str malformed: %s", fp.JA3Str)
			}
			if fp.JA4R[:10] != fp.JA4[:10] {
				t.Errorf("JA4_r prefix %s vs %s", fp.JA4R[:10], fp.JA4[:10])
			}
		})
	}
}

// Clients may split a ClientHello across several TLS records.
func TestParseFragmentedRecords(t *testing.T) {
	raw := captureHello(t, &tls.Config{ServerName: "acpwb.com", InsecureSkipVerify: true})
	whole, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	body := raw[5:]
	var split []byte
	for len(body) > 0 {
		n := min(len(body), 100)
		split = append(split, 0x16, 0x03, 0x01, byte(n>>8), byte(n))
		split = append(split, body[:n]...)
		body = body[n:]
	}
	got, err := Parse(split)
	if err != nil {
		t.Fatal(err)
	}
	if Compute(got) != Compute(whole) {
		t.Fatal("fragmented hello fingerprints differently")
	}
	if _, err := Parse(split[:len(split)-1]); !errors.Is(err, ErrNeedMore) {
		t.Fatalf("truncated: got %v, want ErrNeedMore", err)
	}
}

func TestParseRejectsNonTLS(t *testing.T) {
	for _, in := range []string{"GET / HTTP/1.1\r\n", "\x16\x02\x00\x00", "SSH-2.0-OpenSSH"} {
		if _, err := Parse([]byte(in)); !errors.Is(err, ErrNotTLS) {
			t.Errorf("%q: got %v, want ErrNotTLS", in, err)
		}
	}
	// A handshake record whose message isn't a ClientHello.
	if _, err := Parse([]byte{0x16, 0x03, 0x01, 0x00, 0x04, 0x02, 0x00, 0x00, 0x00}); !errors.Is(err, ErrMalformed) {
		t.Errorf("ServerHello: got %v, want ErrMalformed", err)
	}
}
