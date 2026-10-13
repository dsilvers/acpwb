// Package livefeed feeds the dashboard's live request stream
// (/acpwb-dashboard/live/): it buffers the edge's request and connection
// records and PUBLISHes them to Redis as one bundle a second on Channel,
// where ws_service filters and aggregates them for each viewer.
//
// It only does anything while someone is watching: ws_service keeps
// WatchKey alive (a few seconds' TTL) while it has a viewer connected, and
// the feed checks it once a second. Until then Active is false and the
// edge skips it entirely — no copies, no TLS snapshots, no JSON.
//
// A bundle holds at most maxRequests requests and maxConns connections;
// past that, records are only counted (requests_total / conns_total stay
// exact), so a flood can't turn into megabytes a second through Redis.
package livefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"acpwb_edge/edge"
)

const (
	Channel  = "edge_live"
	WatchKey = "acpwb:live:watching"

	interval    = time.Second
	sendTimeout = 500 * time.Millisecond
)

// request is a RequestRecord with its connection's TLS details.
type request struct {
	edge.RequestRecord
	edge.ConnTLS
}

// conn is the slice of a connection Record the live view uses — the full
// one carries several KB of JA3/JA4_r strings and header names.
type conn struct {
	Timestamp    string `json:"timestamp"`
	ConnID       string `json:"conn_id"`
	IPAddress    string `json:"ip_address"`
	SNI          string `json:"sni"`
	ALPN         string `json:"alpn"`
	TLSVer       string `json:"tls_version"`
	TLSCipher    string `json:"tls_cipher"`
	Outcome      string `json:"outcome"`
	JA4          string `json:"ja4"`
	JA3          string `json:"ja3"`
	RTTMicros    int64  `json:"rtt_us"`
	HandshakeUS  int64  `json:"handshake_us"`
	DurationMS   int64  `json:"duration_ms"`
	BytesIn      int64  `json:"bytes_in"`
	BytesOut     int64  `json:"bytes_out"`
	RequestCount int    `json:"request_count"`
	UnservedReqs int    `json:"unserved_requests"`
	CloseReason  string `json:"close_reason"`
	H2Resets     int    `json:"h2_resets"`
	H2Goaway     string `json:"h2_goaway"`
	UserAgent    string `json:"first_user_agent"`
}

type bundle struct {
	T             int64     `json:"t"` // unix ms at publish
	IntervalMS    int64     `json:"interval_ms"`
	RequestsTotal int       `json:"requests_total"`
	ConnsTotal    int       `json:"conns_total"`
	Requests      []request `json:"requests"`
	Conns         []conn    `json:"conns"`
}

type Feed struct {
	client      *redis.Client
	maxRequests int
	maxConns    int
	active      atomic.Bool

	mu        sync.Mutex
	reqs      []request
	conns     []conn
	reqTotal  int
	connTotal int

	published atomic.Int64
	stop      chan struct{}
	done      chan struct{}
	once      sync.Once
}

// New connects lazily, like edgequeue: an unreachable Redis just means no
// viewer is ever seen.
func New(redisURL string, maxRequests, maxConns int) (*Feed, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("livefeed: parsing redis URL: %w", err)
	}
	opt.DialTimeout = time.Second
	opt.ReadTimeout = sendTimeout
	opt.WriteTimeout = sendTimeout
	opt.PoolSize = 2
	f := &Feed{
		client:      redis.NewClient(opt),
		maxRequests: maxRequests,
		maxConns:    maxConns,
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
	go f.run()
	return f, nil
}

func (f *Feed) Active() bool { return f != nil && f.active.Load() }

func (f *Feed) Request(rec *edge.RequestRecord, t edge.ConnTLS) {
	f.mu.Lock()
	f.reqTotal++
	if len(f.reqs) < f.maxRequests {
		f.reqs = append(f.reqs, request{*rec, t})
	}
	f.mu.Unlock()
}

func (f *Feed) Conn(rec *edge.Record) {
	f.mu.Lock()
	f.connTotal++
	if len(f.conns) < f.maxConns {
		f.conns = append(f.conns, conn{
			Timestamp: rec.Timestamp, ConnID: rec.ConnID, IPAddress: rec.IPAddress,
			SNI: rec.SNI, ALPN: rec.ALPN, TLSVer: rec.TLSVer, TLSCipher: rec.TLSCipher,
			Outcome: rec.Outcome, JA4: rec.JA4, JA3: rec.JA3,
			RTTMicros: rec.RTTMicros, HandshakeUS: rec.HandshakeMicros, DurationMS: rec.DurationMillis,
			BytesIn: rec.BytesIn, BytesOut: rec.BytesOut,
			RequestCount: rec.RequestCount, UnservedReqs: rec.UnservedReqs,
			CloseReason: rec.CloseReason, H2Resets: rec.H2Resets, H2Goaway: rec.H2Goaway,
			UserAgent: rec.FirstUserAgent,
		})
	}
	f.mu.Unlock()
}

// take empties the buffers into a bundle.
func (f *Feed) take(now time.Time) bundle {
	f.mu.Lock()
	defer f.mu.Unlock()
	b := bundle{
		T: now.UnixMilli(), IntervalMS: interval.Milliseconds(),
		RequestsTotal: f.reqTotal, ConnsTotal: f.connTotal,
		Requests: f.reqs, Conns: f.conns,
	}
	if b.Requests == nil {
		b.Requests = []request{}
	}
	if b.Conns == nil {
		b.Conns = []conn{}
	}
	f.reqs, f.conns = nil, nil
	f.reqTotal, f.connTotal = 0, 0
	return b
}

func (f *Feed) run() {
	defer close(f.done)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-f.stop:
			return
		case now := <-t.C:
			f.tick(now)
		}
	}
}

// tick publishes the last second's bundle (if anyone was watching) and
// checks whether anyone still is, in one round trip.
func (f *Feed) tick(now time.Time) {
	wasActive := f.active.Load()
	b := f.take(now)
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	pipe := f.client.Pipeline()
	if wasActive {
		if data, err := json.Marshal(b); err == nil {
			pipe.Publish(ctx, Channel, data)
		}
	}
	watching := pipe.Exists(ctx, WatchKey)
	if _, err := pipe.Exec(ctx); err != nil {
		f.active.Store(false)
		return
	}
	if wasActive {
		f.published.Add(1)
	}
	f.active.Store(watching.Val() > 0)
}

// Stats reports bundles published since startup.
func (f *Feed) Stats() map[string]int64 {
	if f == nil {
		return map[string]int64{}
	}
	active := int64(0)
	if f.active.Load() {
		active = 1
	}
	return map[string]int64{"live_bundles_published": f.published.Load(), "live_active": active}
}

func (f *Feed) Close() {
	if f == nil {
		return
	}
	f.once.Do(func() {
		close(f.stop)
		<-f.done
		f.client.Close()
	})
}
