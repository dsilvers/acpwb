package tlsfp

import (
	"bufio"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	utls "github.com/refraction-networking/utls"
)

// Byte-exact browser ClientHellos (uTLS profiles) and Wireshark's own JA4,
// JA4_r, JA3 and JA3 string for each — an independent implementation.
// Regenerate with scripts/ja4golden.sh.
func TestBrowserHellosMatchWireshark(t *testing.T) {
	dir := filepath.Join("testdata", "hellos")
	names := map[string]string{} // source port → profile
	readTSV(t, filepath.Join(dir, "index.tsv"), func(f []string) { names[f[0]] = f[1] })

	n := 0
	readTSV(t, filepath.Join(dir, "golden.tsv"), func(f []string) {
		if len(f) != 5 {
			t.Fatalf("golden row %v: want 5 fields", f)
		}
		name := names[f[0]]
		raw, err := os.ReadFile(filepath.Join(dir, name+".bin"))
		if err != nil {
			t.Fatal(err)
		}
		h, err := Parse(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := Compute(h)
		for _, c := range []struct{ what, got, want string }{
			{"JA4", got.JA4, f[1]},
			{"JA4_r", got.JA4R, f[2]},
			{"JA3", got.JA3, f[3]},
			{"JA3 string", got.JA3Str, f[4]},
		} {
			if c.got != c.want {
				t.Errorf("%s %s\n ours      %s\n wireshark %s", name, c.what, c.got, c.want)
			}
		}
		n++
	})
	if n < 15 {
		t.Fatalf("only %d golden rows — regenerate with scripts/ja4golden.sh", n)
	}
}

func readTSV(t *testing.T, path string, row func([]string)) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(nil, 1<<20)
	for s.Scan() {
		if line := s.Text(); line != "" {
			row(strings.Split(line, "\t"))
		}
	}
}

// Chrome randomizes GREASE values and (since 106) the order of its
// extensions on every connection. JA4 sorts and drops GREASE, so one
// Chrome yields one JA4 — while JA3, which keeps wire order, doesn't.
// That's the reason to key detection on JA4.
//
// One caveat, also true of real BoringSSL clients: the padding extension
// (0x0015) is only added when the hello would otherwise be 256–511 bytes,
// and a random-length ECH GREASE payload can tip a hello either side of
// that (uTLS's Chrome 120 sends it ~half the time; real Chrome 124+ hellos
// carry an ML-KEM key share and are always far over 512). JA4 counts and
// hashes padding per spec, so such a client has two JA4s, differing only
// by it.
func TestChromeJA4StableAcrossRandomization(t *testing.T) {
	for _, id := range []utls.ClientHelloID{utls.HelloChrome_133, utls.HelloChrome_131, utls.HelloChrome_120, utls.HelloChrome_106_Shuffle} {
		ja4s, ja4sNoPad, ja3s := map[string]bool{}, map[string]bool{}, map[string]bool{}
		for range 16 {
			h, err := Parse(utlsHello(t, id))
			if err != nil {
				t.Fatal(err)
			}
			fp := Compute(h)
			ja4s[fp.JA4], ja3s[fp.JA3] = true, true
			noPad := *h
			noPad.Extensions = nil
			for _, e := range h.Extensions {
				if e != 0x0015 {
					noPad.Extensions = append(noPad.Extensions, e)
				}
			}
			ja4sNoPad[Compute(&noPad).JA4] = true
		}
		if len(ja4sNoPad) != 1 || len(ja4s) > 2 {
			t.Errorf("%s: JA4 varied beyond the padding extension: %v", id.Str(), ja4s)
		}
		if len(ja3s) < 2 {
			t.Errorf("%s: JA3 never varied — is extension shuffling still on?", id.Str())
		}
	}
	// Post-quantum Chrome is always over 512 bytes: exactly one JA4.
	ja4s := map[string]bool{}
	for range 16 {
		h, _ := Parse(utlsHello(t, utls.HelloChrome_133))
		ja4s[Compute(h).JA4] = true
	}
	if len(ja4s) != 1 {
		t.Errorf("Chrome 133: want exactly one JA4, got %v", ja4s)
	}
}

// Every profile's live JA4 matches crypto/tls's own parse of the same bytes.
func TestBrowserProfilesMatchCryptoTLS(t *testing.T) {
	for _, id := range []utls.ClientHelloID{
		utls.HelloChrome_Auto, utls.HelloFirefox_Auto, utls.HelloSafari_Auto, utls.HelloIOS_Auto,
		utls.HelloEdge_Auto, utls.HelloAndroid_11_OkHttp, utls.HelloRandomizedALPN, utls.HelloRandomizedNoALPN,
	} {
		raw := utlsHello(t, id)
		h, err := Parse(raw)
		if err != nil {
			t.Fatalf("%s: %v", id.Str(), err)
		}
		if got, want := Compute(h).JA4, ja4FromInfo(infoFor(t, raw)); got != want {
			t.Errorf("%s: JA4 %s, crypto/tls-derived %s", id.Str(), got, want)
		}
	}
}

func utlsHello(t *testing.T, id utls.ClientHelloID) []byte {
	t.Helper()
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		uc := utls.UClient(client, &utls.Config{ServerName: "acpwb.com", InsecureSkipVerify: true}, id)
		_ = uc.Handshake()
		client.Close()
	}()
	var raw []byte
	buf := make([]byte, 4096)
	for {
		server.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := server.Read(buf)
		raw = append(raw, buf[:n]...)
		if _, perr := Parse(raw); perr == nil {
			return raw
		} else if !errors.Is(perr, ErrNeedMore) {
			t.Fatalf("%s: %v", id.Str(), perr)
		}
		if err != nil {
			t.Fatalf("%s: %v", id.Str(), err)
		}
	}
}
