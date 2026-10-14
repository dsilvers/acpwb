// acpwb_edge is the public front-end for acpwb.com and botseed.net. It
// replaced nginx (and the local-dev HAProxy): it terminates TLS itself so
// it can fingerprint every client — JA4/JA3 from the raw ClientHello, JA4H
// from request header order, and the HTTP/2 SETTINGS/WINDOW_UPDATE/
// PRIORITY/pseudo-header fingerprint — forwards those to the apps as
// headers, and queues one record per connection (including failed
// handshakes) for the EdgeConnection table. Routing, static files and gzip
// replicate the nginx config it replaced.
//
// Configuration is all environment variables (see deploy/acpwb-edge.service
// and docker-compose-local.yml).
package main

import (
	"context"
	"crypto/tls"
	"expvar"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"acpwb_edge/certs"
	"acpwb_edge/classify"
	"acpwb_edge/edge"
	"acpwb_edge/edgequeue"
	"acpwb_edge/httpfp"
	"acpwb_edge/livefeed"
	"acpwb_edge/proxy"
	"acpwb_edge/router"
)

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}

// curvePreferences picks the key exchange groups. Go ignores the order of
// this list and always prefers X25519MLKEM768 when a client offers it, so
// leaving it out is the only way to negotiate X25519 the way nginx's
// OpenSSL 3.0 did — a ~1.1 KB smaller ServerHello. Clients that sent only
// an ML-KEM key share get a HelloRetryRequest. pq restores Go's default.
func curvePreferences(pq bool) []tls.CurveID {
	if pq {
		return nil
	}
	return []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521}
}

func envSeconds(key string) time.Duration {
	return time.Duration(envInt(key, 0)) * time.Second
}

func envInt(key string, def int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v > 0 {
		return v
	}
	return def
}

// envSignedInt is envInt that also accepts negatives ("off").
func envSignedInt(key string, def int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v != 0 {
		return v
	}
	return def
}

func main() {
	log.SetFlags(0) // journald timestamps lines already

	dev := envBool("EDGE_DEV")
	certDir := env("EDGE_CERT_DIR", "/etc/acpwb-edge/certs")

	var store *certs.Store
	var err error
	if dev {
		store, err = certs.SelfSigned("acpwb.com", "*.acpwb.com", "acpwb.example", "*.acpwb.example",
			"botseed.net", "localhost", "127.0.0.1", "::1")
	} else {
		store, err = certs.Load(certDir, "acpwb.com")
	}
	if err != nil {
		log.Fatalf("acpwb_edge: %v", err)
	}
	store.CheckExpiry(time.Now())

	up := func(name, addr string, keepalive int, timeout time.Duration) http.Handler {
		return proxy.New(proxy.Upstream{Name: name, Addr: addr, Keepalive: keepalive, Timeout: timeout})
	}
	rc := router.Config{
		Dev:         dev,
		StaticRoot:  env("EDGE_STATIC_ROOT", "/home/acpwb/acpwb/acpwb/staticfiles"),
		BotseedRoot: env("EDGE_BOTSEED_ROOT", "/home/acpwb/acpwb/botseed"),
		ACMERoot:    env("EDGE_ACME_ROOT", "/var/www/certbot"),
		Upstreams: map[string]http.Handler{
			router.Django:     up("django", env("EDGE_DJANGO_UPSTREAM", "unix:/run/acpwb-gunicorn/gunicorn.sock"), 128, 60*time.Second),
			router.AcpwbGo:    up("acpwb_go", env("EDGE_GO_UPSTREAM", "127.0.0.1:8091"), 64, 60*time.Second),
			router.WS:         up("ws", env("EDGE_WS_UPSTREAM", "127.0.0.1:8765"), 4, 3600*time.Second),
			router.BotseedWS:  up("botseed_ws", env("EDGE_BOTSEED_WS_UPSTREAM", "127.0.0.1:8766"), 4, 3600*time.Second),
			router.BotseedAPI: up("botseed_api", env("EDGE_BOTSEED_API_UPSTREAM", "127.0.0.1:8767"), 16, 10*time.Second),
		},
	}

	var sink edge.Sink
	var queue *edgequeue.Queue
	var live edge.Live // stays a nil interface without Redis
	var feed *livefeed.Feed
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if queue, err = edgequeue.New(redisURL, envInt("EDGE_QUEUE_MAX", 1_000_000)); err != nil {
			log.Fatalf("acpwb_edge: %v", err)
		}
		sink = queue
		// The dashboard's live stream: one bundle a second while someone
		// watches /acpwb-dashboard/live/, capped per bundle.
		if feed, err = livefeed.New(redisURL,
			int(envInt("EDGE_LIVE_MAX_REQUESTS", 5000)), int(envInt("EDGE_LIVE_MAX_CONNS", 2000))); err != nil {
			log.Fatalf("acpwb_edge: %v", err)
		}
		live = feed
	} else {
		log.Printf("acpwb_edge: REDIS_URL unset — connection and request records are not being recorded")
	}

	// Signs the acpwb_cid client cookie. Must stay the same across
	// restarts, or every browser gets a new client ID.
	cookieSecret := []byte(os.Getenv("EDGE_COOKIE_SECRET"))
	if len(cookieSecret) == 0 && !dev {
		log.Printf("acpwb_edge: EDGE_COOKIE_SECRET unset — client IDs reset on every restart")
	}

	// Client networks (ASN, hosting vs residential) for request
	// classification. Without the database every network is "unknown";
	// everything else still works. SIGHUP reloads it after
	// download_geoip_db fetches the monthly update.
	asnPath := env("EDGE_ASN_DB", "/home/acpwb/acpwb/acpwb/var/geoip/GeoLite2-ASN.mmdb")
	asnDB := &classify.ASNDB{}
	if err := asnDB.Load(asnPath); err != nil {
		log.Printf("acpwb_edge: ASN database: %v — client networks will be unknown", err)
	}

	srv := edge.New(edge.Config{
		TLSConfig: &tls.Config{
			// nginx accepted TLSv1 and TLSv1.1 too; old-client bots keep
			// reaching the traps (and their JA4s get recorded).
			MinVersion:       tls.VersionTLS10,
			GetCertificate:   store.GetCertificate,
			CurvePreferences: curvePreferences(envBool("EDGE_PQ_KEX")),
		},
		TLS:          router.TLS(rc),
		Plain:        router.Plain(rc, envBool("EDGE_PLAIN_BOTSEED")),
		Sink:         sink,
		Live:         live,
		CookieDomain: rc.CookieDomain,
		CookieSecret: cookieSecret,
		HTTPSBase:    rc.HTTPSBase,
		MaxBodyBytes: envInt("EDGE_MAX_BODY_BYTES", 10<<20),
		ASN:          asnDB,
		// 0 = the edge package defaults.
		HandshakeTimeout:  envSeconds("EDGE_HANDSHAKE_TIMEOUT_SECONDS"),
		ReadHeaderTimeout: envSeconds("EDGE_HEADER_TIMEOUT_SECONDS"),
		IdleTimeout:       envSeconds("EDGE_IDLE_TIMEOUT_SECONDS"),
		Limits: edge.Limits{
			StallTimeout:     envSeconds("EDGE_STALL_TIMEOUT_SECONDS"),
			RequestsInFlight: int(envSignedInt("EDGE_MAX_IN_FLIGHT", 100000)),
			RecordsPerIPMin:  int(envSignedInt("EDGE_RECORDS_PER_IP_MINUTE", 600)),
		},
	})

	httpsAddr, httpAddr := env("EDGE_HTTPS_ADDR", ":443"), env("EDGE_HTTP_ADDR", ":80")
	tl, err := net.Listen("tcp", httpsAddr)
	if err != nil {
		log.Fatalf("acpwb_edge: %v", err)
	}
	go func() {
		if err := srv.ServeTLS(tl); err != nil {
			log.Fatalf("acpwb_edge: %s: %v", httpsAddr, err)
		}
	}()
	if httpAddr != "" {
		pl, err := net.Listen("tcp", httpAddr)
		if err != nil {
			log.Fatalf("acpwb_edge: %v", err)
		}
		go func() {
			if err := srv.ServePlain(pl); err != nil {
				log.Fatalf("acpwb_edge: %s: %v", httpAddr, err)
			}
		}()
	}
	log.Printf("acpwb_edge: listening on %s (TLS) and %s (plain), dev=%v", httpsAddr, httpAddr, dev)

	started := time.Now().Unix()
	stats := func() map[string]int64 {
		st := edge.Stats()
		st["tap_panics"] = httpfp.Panics()
		st["goroutines"] = int64(runtime.NumGoroutine())
		st["started_at"] = started
		for k, v := range queue.Stats() {
			st[k] = v
		}
		for k, v := range feed.Stats() {
			st[k] = v
		}
		return st
	}
	go func() {
		for range time.Tick(30 * time.Second) {
			queue.PublishStats(stats())
		}
	}()

	// Debug listener: expvar counters at /debug/vars (conns, handshakes,
	// queue drops, parser panics, goroutines) and pprof. Bind it to
	// localhost only — it's never for the public.
	if debugAddr := os.Getenv("EDGE_DEBUG_ADDR"); debugAddr != "" {
		expvar.Publish("edge", expvar.Func(func() any { return stats() }))
		go func() {
			// http.DefaultServeMux: where expvar and net/http/pprof register.
			log.Printf("acpwb_edge: debug listener on %s: %v", debugAddr, http.ListenAndServe(debugAddr, nil))
		}()
	}

	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	certHup := make(chan os.Signal, 1)
	go func() {
		for sig := range hup {
			if err := asnDB.Load(asnPath); err != nil {
				log.Printf("acpwb_edge: reloading ASN database: %v", err)
			}
			select {
			case certHup <- sig:
			default: // a reload is already pending
			}
		}
	}()
	stop := make(chan struct{})
	go store.Watch(certHup, time.Hour, stop)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT)
	<-term
	log.Printf("acpwb_edge: shutting down")
	close(stop)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("acpwb_edge: shutdown: %v", err)
	}
	feed.Close()
	queue.Close()
}
