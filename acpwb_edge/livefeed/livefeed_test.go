package livefeed

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"acpwb_edge/edge"
)

func deadRedis(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	return "redis://" + addr + "/0"
}

// Past the caps, records are only counted, and a bundle empties the
// buffers.
func TestBundleCapsAndCounts(t *testing.T) {
	f, err := New(deadRedis(t), 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for i := range 10 {
		f.Request(&edge.RequestRecord{Path: "/p", Seq: i}, edge.ConnTLS{Version: "TLSv1.3", JA4: "t13d"})
	}
	f.Conn(&edge.Record{Outcome: "ok", JA3Str: "long"})
	f.Conn(&edge.Record{Outcome: "no_hello"})

	b := f.take(time.UnixMilli(1000))
	if b.RequestsTotal != 10 || len(b.Requests) != 3 || b.ConnsTotal != 2 || len(b.Conns) != 1 {
		t.Fatalf("bundle totals %d/%d reqs, %d/%d conns", len(b.Requests), b.RequestsTotal, len(b.Conns), b.ConnsTotal)
	}
	data, _ := json.Marshal(b)
	var got struct {
		Requests []map[string]any `json:"requests"`
		Conns    []map[string]any `json:"conns"`
	}
	json.Unmarshal(data, &got)
	r := got.Requests[0]
	if r["path"] != "/p" || r["tls_version"] != "TLSv1.3" || r["ja4"] != "t13d" {
		t.Errorf("request JSON %v", r)
	}
	if _, ok := got.Conns[0]["ja3_str"]; ok || got.Conns[0]["outcome"] != "ok" {
		t.Errorf("conn JSON %v", got.Conns[0])
	}

	empty := f.take(time.UnixMilli(2000))
	if empty.RequestsTotal != 0 || len(empty.Requests) != 0 || empty.Requests == nil {
		t.Errorf("second bundle %+v", empty)
	}
}

// With Redis down nobody can be watching, so the feed stays inactive.
func TestInactiveWithoutRedis(t *testing.T) {
	f, _ := New(deadRedis(t), 10, 10)
	defer f.Close()
	f.active.Store(true)
	f.tick(time.Now())
	if f.Active() {
		t.Fatal("active with Redis down")
	}
	var nilFeed *Feed
	if nilFeed.Active() {
		t.Fatal("nil feed active")
	}
	nilFeed.Close()
}
