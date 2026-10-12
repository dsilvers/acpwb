// Package certs serves acpwb_edge's TLS certificates. certbot still issues
// and renews them (acpwb.com's wildcard needs DNS-01, which the edge can't
// do); its deploy hook (deploy/certbot-deploy-hook.sh) copies each renewed
// lineage into <dir>/<name>/{fullchain,privkey}.pem and SIGHUPs the edge.
// The store also re-checks file mtimes hourly in case a reload was missed.
package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ExpiryWarning is how close to expiry a loaded certificate starts logging
// a warning on every check.
const ExpiryWarning = 14 * 24 * time.Hour

type set struct {
	certs []*tls.Certificate // sorted by lineage name
	def   *tls.Certificate
}

type Store struct {
	dir, defName string

	cur atomic.Pointer[set]

	mu     sync.Mutex
	mtimes map[string]time.Time
}

// Load reads every <dir>/<name>/{fullchain,privkey}.pem. defaultName is the
// lineage served when SNI is missing or matches nothing.
func Load(dir, defaultName string) (*Store, error) {
	s := &Store{dir: dir, defName: defaultName}
	if _, err := s.Reload(); err != nil {
		return nil, err
	}
	return s, nil
}

// Reload re-reads the certificates if any file changed. On error the
// previously loaded set stays in service.
func (s *Store) Reload() (changed bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dirs, err := os.ReadDir(s.dir)
	if err != nil {
		return false, fmt.Errorf("certs: %w", err)
	}
	mtimes := map[string]time.Time{}
	var names []string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		for _, f := range []string{"fullchain.pem", "privkey.pem"} {
			fi, err := os.Stat(filepath.Join(s.dir, d.Name(), f))
			if err != nil {
				return false, fmt.Errorf("certs: %w", err)
			}
			mtimes[filepath.Join(d.Name(), f)] = fi.ModTime()
		}
		names = append(names, d.Name())
	}
	if len(names) == 0 {
		return false, fmt.Errorf("certs: no certificate directories in %s", s.dir)
	}
	if s.cur.Load() != nil && mapsEqual(mtimes, s.mtimes) {
		return false, nil
	}

	slices.Sort(names)
	next := &set{}
	for _, name := range names {
		c, err := tls.LoadX509KeyPair(filepath.Join(s.dir, name, "fullchain.pem"), filepath.Join(s.dir, name, "privkey.pem"))
		if err != nil {
			return false, fmt.Errorf("certs: %s: %w", name, err)
		}
		next.certs = append(next.certs, &c)
		if name == s.defName {
			next.def = &c
		}
	}
	if next.def == nil {
		next.def = next.certs[0]
	}
	s.cur.Store(next)
	s.mtimes = mtimes
	return true, nil
}

func mapsEqual(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !b[k].Equal(v) {
			return false
		}
	}
	return true
}

// GetCertificate is tls.Config.GetCertificate: the first certificate valid
// for the SNI name (wildcards included), else the default.
func (s *Store) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cs := s.cur.Load()
	if cs == nil {
		return nil, errors.New("certs: none loaded")
	}
	if name := strings.TrimSuffix(strings.ToLower(hello.ServerName), "."); name != "" {
		for _, c := range cs.certs {
			if c.Leaf != nil && c.Leaf.VerifyHostname(name) == nil {
				return c, nil
			}
		}
	}
	return cs.def, nil
}

// CheckExpiry logs a warning for every loaded certificate expiring soon.
func (s *Store) CheckExpiry(now time.Time) {
	cs := s.cur.Load()
	if cs == nil {
		return
	}
	for _, c := range cs.certs {
		if c.Leaf != nil && c.Leaf.NotAfter.Sub(now) < ExpiryWarning {
			log.Printf("certs: WARNING %v expires %s — is certbot renewing and the deploy hook copying?",
				c.Leaf.DNSNames, c.Leaf.NotAfter.Format(time.RFC3339))
		}
	}
}

// Watch reloads on every receive from reload (SIGHUP) and every interval,
// and checks expiry each time, until stop is closed.
func (s *Store) Watch(reload <-chan os.Signal, interval time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-reload:
		case <-t.C:
		}
		if s.dir == "" {
			continue // self-signed
		}
		if changed, err := s.Reload(); err != nil {
			log.Printf("certs: reload failed, keeping current certificates: %v", err)
		} else if changed {
			log.Printf("certs: reloaded from %s", s.dir)
		}
		s.CheckExpiry(time.Now())
	}
}

// SelfSigned returns a store holding one fresh self-signed certificate for
// hosts (names or IPs) — local development only.
func SelfSigned(hosts ...string) (*Store, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: hosts[0]},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	c := &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
	s := &Store{}
	s.cur.Store(&set{certs: []*tls.Certificate{c}, def: c})
	return s, nil
}
