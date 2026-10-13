package edge

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
)

// ClientCookieName is the first-party cookie carrying a browser's client
// ID, so every request from one browser can be grouped across connections
// and IP addresses. It holds a random ID and an HMAC of it — nothing about
// the visitor — and is set on HTTPS responses only.
const ClientCookieName = "acpwb_cid"

const clientCookieMaxAge = 365 * 24 * 60 * 60

// clientIDs issues and verifies client cookies. The signature stops a
// client from claiming another client's ID (polluting that client's
// history); it isn't secrecy — the ID itself is random.
type clientIDs struct {
	key    []byte
	domain func(host string) (string, bool)
}

func newClientIDs(secret []byte, domain func(string) (string, bool)) *clientIDs {
	if domain == nil {
		return nil
	}
	key := sha256.Sum256(secret)
	if len(secret) == 0 {
		_, _ = rand.Read(key[:]) // dev: cookies last until the edge restarts
	}
	return &clientIDs{key: key[:], domain: domain}
}

func (c *clientIDs) sign(id string) string {
	m := hmac.New(sha256.New, c.key)
	m.Write([]byte(id))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil)[:12])
}

// verify returns the client ID in a cookie value, or "" if it isn't one
// this edge issued.
func (c *clientIDs) verify(v string) string {
	id, mac, ok := strings.Cut(v, ".")
	if !ok || len(id) != 22 || !hmac.Equal([]byte(mac), []byte(c.sign(id))) {
		return ""
	}
	return id
}

// identify returns r's client ID, issuing a new one (and setting the
// cookie on w) if r carries no valid one. A nil clientIDs, or a host that
// gets no cookie, yields "".
func (c *clientIDs) identify(w http.ResponseWriter, r *http.Request, host string) (id string, issued bool) {
	if c == nil {
		return "", false
	}
	for _, ck := range r.CookiesNamed(ClientCookieName) {
		if id := c.verify(ck.Value); id != "" {
			return id, false
		}
	}
	domain, ok := c.domain(host)
	if !ok {
		return "", false
	}
	var b [16]byte
	_, _ = rand.Read(b[:])
	id = base64.RawURLEncoding.EncodeToString(b[:])
	ck := &http.Cookie{
		Name:     ClientCookieName,
		Value:    id + "." + c.sign(id),
		Path:     "/",
		Domain:   domain,
		MaxAge:   clientCookieMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	w.Header().Add("Set-Cookie", ck.String())
	return id, true
}
