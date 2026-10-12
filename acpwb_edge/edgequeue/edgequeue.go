// Package edgequeue ships acpwb_edge's per-connection records to Redis
// (acpwb:edge_conn_queue), where Django's drain_edge_conn_queue picks them
// up into the EdgeConnection hypertable — the same RPUSH-then-drain pattern
// acpwb_go's visitqueue and Django's crawler_queue use.
//
// Pushing never blocks a connection: records go onto a bounded channel and
// a single worker sends them in pipelined batches. If the channel is full
// or Redis is down, records are dropped and counted; the count is logged
// once a minute.
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
	QueueKey     = "acpwb:edge_conn_queue"
	chanCapacity = 64 << 10
	maxBatch     = 500
	batchWait    = 50 * time.Millisecond
	pushTimeout  = 2 * time.Second
	// After a failed push, records are dropped for this long instead of
	// piling up retries against a Redis that's down.
	backoff = 5 * time.Second
)

type Queue struct {
	client   *redis.Client
	queueMax int64
	ch       chan []byte
	done     chan struct{}
	stopOnce sync.Once

	dropped atomic.Int64
	pushed  atomic.Int64
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
		ch:       make(chan []byte, chanCapacity),
		done:     make(chan struct{}),
	}
	go q.run()
	go q.reportDrops()
	return q, nil
}

// Push queues v (marshalled to JSON) without blocking. A nil Queue drops
// everything, so callers needn't check whether Redis was configured.
func (q *Queue) Push(v any) {
	if q == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		q.dropped.Add(1)
		return
	}
	select {
	case q.ch <- data:
	default:
		q.dropped.Add(1)
	}
}

// Dropped and Pushed report counters since startup.
func (q *Queue) Dropped() int64 { return q.dropped.Load() }
func (q *Queue) Pushed() int64  { return q.pushed.Load() }

// Close flushes what's queued (best effort, bounded by pushTimeout) and
// stops the worker.
func (q *Queue) Close() {
	if q == nil {
		return
	}
	q.stopOnce.Do(func() {
		close(q.ch)
		<-q.done
		q.client.Close()
	})
}

func (q *Queue) run() {
	defer close(q.done)
	var downUntil time.Time
	batch := make([]any, 0, maxBatch)
	for {
		first, ok := <-q.ch
		if !ok {
			return
		}
		batch = append(batch[:0], first)
		closed := false
		timer := time.NewTimer(batchWait)
	fill:
		for len(batch) < maxBatch {
			select {
			case data, ok := <-q.ch:
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

		if time.Now().Before(downUntil) {
			q.dropped.Add(int64(len(batch)))
		} else if err := q.send(batch); err != nil {
			q.dropped.Add(int64(len(batch)))
			downUntil = time.Now().Add(backoff)
		} else {
			q.pushed.Add(int64(len(batch)))
		}
		if closed {
			return
		}
	}
}

func (q *Queue) send(batch []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()
	pipe := q.client.Pipeline()
	pipe.RPush(ctx, QueueKey, batch...)
	// Bounded like the fingerprint queue: keep only the newest queueMax
	// entries if the drain falls behind (LTRIM is O(1) when nothing goes).
	pipe.LTrim(ctx, QueueKey, -q.queueMax, -1)
	_, err := pipe.Exec(ctx)
	return err
}

func (q *Queue) reportDrops() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	var last int64
	for {
		select {
		case <-q.done:
			return
		case <-t.C:
			if d := q.dropped.Load(); d != last {
				log.Printf("edgequeue: dropped %d connection records in the last minute (%d total)", d-last, d)
				last = d
			}
		}
	}
}
