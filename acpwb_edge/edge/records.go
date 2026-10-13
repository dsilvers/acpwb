package edge

import (
	"sync"
	"sync/atomic"
	"time"

	"acpwb_edge/edgequeue"
)

var recordsAggregated atomic.Int64

// recordGate sits between closed connections and the Sink. A connection
// that carried any request always gets its own record — its EdgeRequest
// rows point at it by conn_id. Requestless connections (failed or
// abandoned handshakes, connect-and-leave) get perMin individual records
// per IP a minute; past that they're folded into one record per (IP,
// outcome, JA4) per minute flagged aggregated, carrying conn_count and
// summed bytes — a connection flood is still counted exactly, but can't
// turn into millions of rows (a single local client managed 1.8M
// connections in 20s in testing).
type recordGate struct {
	sink   Sink
	perMin int

	mu     sync.Mutex
	counts map[string]int
	agg    map[aggKey]*Record

	stop chan struct{}
	done chan struct{}
	once sync.Once
}

type aggKey struct{ ip, outcome, ja4 string }

func newRecordGate(sink Sink, perMin int) *recordGate {
	g := &recordGate{
		sink: sink, perMin: perMin,
		counts: map[string]int{}, agg: map[aggKey]*Record{},
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	go g.run()
	return g
}

func (g *recordGate) emit(rec *Record) {
	if g.sink == nil {
		return
	}
	if rec.RequestCount > 0 || rec.UnservedReqs > 0 || g.perMin < 0 {
		g.sink.Push(edgequeue.ConnQueue, rec)
		return
	}
	g.mu.Lock()
	g.counts[rec.IPAddress]++
	if g.counts[rec.IPAddress] <= g.perMin {
		g.mu.Unlock()
		g.sink.Push(edgequeue.ConnQueue, rec)
		return
	}
	recordsAggregated.Add(1)
	k := aggKey{rec.IPAddress, rec.Outcome, rec.JA4}
	if a := g.agg[k]; a != nil {
		a.ConnCount++
		a.BytesIn += rec.BytesIn
		a.BytesOut += rec.BytesOut
		a.DurationMillis += rec.DurationMillis
		a.H2Streams += rec.H2Streams
		a.H2Resets += rec.H2Resets
		a.H2Continuations += rec.H2Continuations
	} else {
		cp := *rec // keeps the first connection's details as a sample
		cp.Aggregated = true
		cp.IdempotencyKey = uuid4()
		cp.ConnID = "" // stands for many connections
		g.agg[k] = &cp
	}
	g.mu.Unlock()
}

func (g *recordGate) flush() {
	g.mu.Lock()
	agg := g.agg
	g.agg, g.counts = map[aggKey]*Record{}, map[string]int{}
	g.mu.Unlock()
	for _, rec := range agg {
		g.sink.Push(edgequeue.ConnQueue, rec)
	}
}

func (g *recordGate) run() {
	defer close(g.done)
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-g.stop:
			if g.sink != nil {
				g.flush()
			}
			return
		case <-t.C:
			if g.sink != nil {
				g.flush()
			}
		}
	}
}

func (g *recordGate) close() {
	g.once.Do(func() {
		close(g.stop)
		<-g.done
	})
}
