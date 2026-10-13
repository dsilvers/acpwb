package edge

import (
	"sync"
	"sync/atomic"
	"time"
)

var recordsAggregated atomic.Int64

// recordGate sits between closed connections and the Sink. Each IP gets
// perMin individual records a minute; past that its connections are folded
// into one record per (IP, outcome, JA4) per minute flagged aggregated,
// carrying conn_count and summed bytes/requests — a connection flood is still
// counted exactly, but can't turn into millions of rows (a single local
// client managed 1.8M connections in 20s in testing).
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
	g.mu.Lock()
	g.counts[rec.IPAddress]++
	if g.counts[rec.IPAddress] <= g.perMin {
		g.mu.Unlock()
		g.sink.Push(rec)
		return
	}
	recordsAggregated.Add(1)
	k := aggKey{rec.IPAddress, rec.Outcome, rec.JA4}
	if a := g.agg[k]; a != nil {
		a.ConnCount++
		a.BytesIn += rec.BytesIn
		a.BytesOut += rec.BytesOut
		a.RequestCount += rec.RequestCount
		a.RejectedReqs += rec.RejectedReqs
		a.DurationMillis += rec.DurationMillis
		a.H2Streams += rec.H2Streams
		a.H2Resets += rec.H2Resets
		a.H2Continuations += rec.H2Continuations
	} else {
		cp := *rec // keeps the first connection's details as a sample
		cp.Aggregated = true
		cp.IdempotencyKey = uuid4()
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
		g.sink.Push(rec)
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
