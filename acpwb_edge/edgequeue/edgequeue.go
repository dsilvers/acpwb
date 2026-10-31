// Package edgequeue ships acpwb_edge's records to Redis — one list per
// record kind: connections (acpwb:edge_conn_queue → EdgeConnection) and
// requests (acpwb:edge_request_queue → EdgeRequest), each drained by its
// own Django command into its hypertable — the same RPUSH-then-drain
// pattern acpwb_go's visitqueue and Django's crawler_queue use.
//
// Pushing never blocks a connection: each kind has its own bounded channel
// and worker sending pipelined batches, so a flood of one kind can't crowd
// out the other. A failed push is retried until it lands while the channel
// holds what arrives meanwhile (the drains dedupe, so a push that timed out
// but landed is harmless). Records are dropped only when a channel is full
// or at shutdown, counted per kind; drops are logged once a minute and
// published with the stats.
package edgequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ConnQueue    = "acpwb:edge_conn_queue"
	RequestQueue = "acpwb:edge_request_queue"
	StatsKey     = "acpwb:edge_stats" // hash of process counters, read by precalc_dashboard

	chanCapacity = 64 << 10
	maxBatch     = 500
	batchWait    = 50 * time.Millisecond
	pushTimeout  = 2 * time.Second
	// Pause between retries of a failed push.
	retryWait = time.Second
)

// Kinds lists the queues, with the short name used in stats.
var Kinds = map[string]string{ConnQueue: "conn", RequestQueue: "request"}

type Queue struct {
	client   *redis.Client
	queueMax int64
	lanes    map[string]*lane
	stop     chan struct{}
	stopOnce sync.Once
}

type lane struct {
	key     string
	ch      chan []byte
	done    chan struct{}
	dropped atomic.Int64 // every lost record, whatever the reason
	full    atomic.Int64 // of those, lost to a full channel
	pushed  atomic.Int64
	// Failed pushes, retried; not lost records.
	pushFailures atomic.Int64
	failingSince time.Time // worker only
}

// New connects lazily: an unreachable Redis at startup isn't an error.
func New(redisURL string, queueMax int64) (*Queue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("edgequeue: parsing redis URL: %w", err)
	}
	opt.DialTimeout = time.Second
	opt.ReadTimeout = pushTimeout
	opt.WriteTimeout = pushTimeout
	q := &Queue{
		client:   redis.NewClient(opt),
		queueMax: queueMax,
		lanes:    map[string]*lane{},
		stop:     make(chan struct{}),
	}
	for key := range Kinds {
		l := &lane{key: key, ch: make(chan []byte, chanCapacity), done: make(chan struct{})}
		q.lanes[key] = l
		go q.run(l)
	}
	go q.reportDrops()
	return q, nil
}

// Push queues v (marshalled to JSON) onto the list key without blocking.
// A nil Queue drops everything, so callers needn't check whether Redis was
// configured.
func (q *Queue) Push(key string, v any) {
	if q == nil {
		return
	}
	l := q.lanes[key]
	if l == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		l.dropped.Add(1)
		return
	}
	select {
	case l.ch <- data:
	default:
		l.dropped.Add(1)
		l.full.Add(1)
	}
}

// Dropped, Pushed and PushFailures report key's counters since startup.
func (q *Queue) Dropped(key string) int64      { return q.lanes[key].dropped.Load() }
func (q *Queue) Pushed(key string) int64       { return q.lanes[key].pushed.Load() }
func (q *Queue) PushFailures(key string) int64 { return q.lanes[key].pushFailures.Load() }

// Close flushes what's queued (best effort: a failing push ends it, and the
// rest is dropped) and stops the workers. No Push may follow it.
func (q *Queue) Close() {
	if q == nil {
		return
	}
	q.stopOnce.Do(func() {
		close(q.stop)
		for _, l := range q.lanes {
			close(l.ch)
		}
		for _, l := range q.lanes {
			<-l.done
		}
		q.client.Close()
	})
}

func (q *Queue) run(l *lane) {
	defer close(l.done)
	batch := make([]any, 0, maxBatch)
	for {
		first, ok := <-l.ch
		if !ok {
			return
		}
		batch = append(batch[:0], first)
		closed := false
		timer := time.NewTimer(batchWait)
	fill:
		for len(batch) < maxBatch {
			select {
			case data, ok := <-l.ch:
				if !ok {
					closed = true
					break fill
				}
				batch = append(batch, data)
			case <-timer.C:
				break fill
			}
		}
		timer.Stop()

		q.deliver(l, batch)
		if closed {
			return
		}
	}
}

// deliver pushes batch, retrying until it lands. While it retries nothing
// leaves the channel, so an outage longer than the channel can hold costs
// the newest records instead of whatever arrives during a short failure.
func (q *Queue) deliver(l *lane, batch []any) {
	for {
		if !l.failingSince.IsZero() && q.stopping() {
			l.dropped.Add(int64(len(batch)))
			return
		}
		err := q.send(l.key, batch)
		if err == nil {
			l.pushed.Add(int64(len(batch)))
			if !l.failingSince.IsZero() {
				log.Printf("edgequeue: %s pushes recovered after %v", Kinds[l.key], time.Since(l.failingSince).Round(time.Millisecond))
				l.failingSince = time.Time{}
			}
			return
		}
		l.pushFailures.Add(1)
		if l.failingSince.IsZero() {
			l.failingSince = time.Now()
			log.Printf("edgequeue: pushing %d %s records failed, retrying: %v", len(batch), Kinds[l.key], err)
		}
		select {
		case <-q.stop:
		case <-time.After(retryWait):
		}
	}
}

func (q *Queue) stopping() bool {
	select {
	case <-q.stop:
		return true
	default:
		return false
	}
}

func (q *Queue) send(key string, batch []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()
	pipe := q.client.Pipeline()
	pipe.RPush(ctx, key, batch...)
	// Bounded like the fingerprint queue: keep only the newest queueMax
	// entries if the drain falls behind (LTRIM is O(1) when nothing goes).
	pipe.LTrim(ctx, key, -q.queueMax, -1)
	_, err := pipe.Exec(ctx)
	return err
}

func (q *Queue) reportDrops() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	last := map[string]int64{}
	for {
		select {
		case <-q.stop:
			return
		case <-t.C:
			for key, l := range q.lanes {
				if d := l.dropped.Load(); d != last[key] {
					log.Printf("edgequeue: dropped %d %s records in the last minute (%d total)", d-last[key], Kinds[key], d)
					last[key] = d
				}
			}
		}
	}
}

// Stats returns records_pushed_<kind> / records_dropped_<kind> counters,
// plus records_dropped_full_<kind> (the part of the drops lost to a full
// channel) and push_failures_<kind> (failed attempts, each retried).
func (q *Queue) Stats() map[string]int64 {
	st := map[string]int64{}
	if q == nil {
		return st
	}
	for key, l := range q.lanes {
		st["records_pushed_"+Kinds[key]] = l.pushed.Load()
		st["records_dropped_"+Kinds[key]] = l.dropped.Load()
		st["records_dropped_full_"+Kinds[key]] = l.full.Load()
		st["push_failures_"+Kinds[key]] = l.pushFailures.Load()
	}
	return st
}

// PublishStats writes the edge's process counters to the StatsKey hash, so
// the dashboard can show what never becomes a row: records dropped here,
// overload 503s, parser panics. Best effort, like everything in this
// package.
func (q *Queue) PublishStats(stats map[string]int64) {
	if q == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()
	fields := make(map[string]any, len(stats)+1)
	for k, v := range stats {
		fields[k] = v
	}
	fields["updated_at"] = time.Now().Unix()
	pipe := q.client.Pipeline()
	pipe.HSet(ctx, StatsKey, fields)
	pipe.Expire(ctx, StatsKey, 10*time.Minute) // goes stale if the edge stops
	_, _ = pipe.Exec(ctx)
}
