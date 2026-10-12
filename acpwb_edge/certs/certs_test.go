package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func writeLineage(t *testing.T, dir, name string, hosts ...string) {
	t.Helper()
	ss, err := SelfSigned(hosts...)
	if err != nil {
		t.Fatal(err)
	}
	c := ss.cur.Load().def
	keyDER, err := x509.MarshalPKCS8PrivateKey(c.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	d := filepath.Join(dir, name)
	os.MkdirAll(d, 0o755)
	os.WriteFile(filepath.Join(d, "fullchain.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Certificate[0]}), 0o644)
	os.WriteFile(filepath.Join(d, "privkey.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600)
}

func names(t *testing.T, s *Store, sni string) []string {
	t.Helper()
	c, err := s.GetCertificate(&tls.ClientHelloInfo{ServerName: sni})
	if err != nil {
		t.Fatal(err)
	}
	return c.Leaf.DNSNames
}

func TestSNISelectionAndReload(t *testing.T) {
	dir := t.TempDir()
	writeLineage(t, dir, "acpwb.com", "acpwb.com", "*.acpwb.com")
	writeLineage(t, dir, "botseed.net", "botseed.net", "www.botseed.net")

	s, err := Load(dir, "acpwb.com")
	if err != nil {
		t.Fatal(err)
	}
	for sni, want := range map[string]string{
		"archives-1999.acpwb.com": "*.acpwb.com",
		"ACPWB.com.":              "acpwb.com",
		"www.botseed.net":         "www.botseed.net",
		"":                        "acpwb.com", // no SNI → default
		"example.org":             "acpwb.com", // unknown → default
	} {
		if got := names(t, s, sni); !slices.Contains(got, want) {
			t.Errorf("SNI %q got cert for %v, want one covering %s", sni, got, want)
		}
	}

	if changed, err := s.Reload(); err != nil || changed {
		t.Fatalf("unchanged reload: changed=%v err=%v", changed, err)
	}

	// certbot renews botseed.net: new cert, new mtime.
	future := time.Now().Add(time.Minute)
	writeLineage(t, dir, "botseed.net", "botseed.net", "renewed.botseed.net")
	os.Chtimes(filepath.Join(dir, "botseed.net", "fullchain.pem"), future, future)
	if changed, err := s.Reload(); err != nil || !changed {
		t.Fatalf("renewed reload: changed=%v err=%v", changed, err)
	}
	if got := names(t, s, "renewed.botseed.net"); !slices.Contains(got, "renewed.botseed.net") {
		t.Fatalf("renewed cert not served: %v", got)
	}

	// A broken renewal keeps the working certificates in service.
	os.WriteFile(filepath.Join(dir, "botseed.net", "privkey.pem"), []byte("junk"), 0o600)
	later := future.Add(time.Minute)
	os.Chtimes(filepath.Join(dir, "botseed.net", "privkey.pem"), later, later)
	if _, err := s.Reload(); err == nil {
		t.Fatal("expected an error for a broken key")
	}
	if got := names(t, s, "renewed.botseed.net"); !slices.Contains(got, "renewed.botseed.net") {
		t.Fatal("a failed reload must keep the previous certificates")
	}
}

func TestLoadFailsWithNoCerts(t *testing.T) {
	if _, err := Load(t.TempDir(), "acpwb.com"); err == nil {
		t.Fatal("expected an error for an empty cert dir")
	}
}
