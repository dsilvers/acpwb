// Package visitqueue ports the request-path side of
// apps/core/crawler_queue.py: RPUSH a JSON payload onto the same Redis
// lists (acpwb:crawler_queue / acpwb:archive_queue) that Django's
// drain_crawler_queue / drain_archive_queue management commands already
// drain into CrawlerVisit / ArchiveVisit rows. This package only needs to
// write in the same shape Django's consumer already expects — the consumer
// side (drain commands, dashboard, bot backfill) is untouched.
package visitqueue

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	crawlerQueueKey = "acpwb:crawler_queue"
	archiveQueueKey = "acpwb:archive_queue"
)

// Queue wraps a Redis client for pushing visit records. A nil *Queue (or one
// whose client is unreachable) causes Push* calls to silently no-op, mirroring
// push_crawler_visit()/push_archive_visit()'s "return False, caller falls back"
// contract — except this Go service has no local DB fallback, so a persistent
// Redis outage means visits simply aren't logged rather than blocking the
// response, which matches this service's only job: serve content fast.
type Queue struct {
	client *redis.Client
}

// New creates a Queue from a redis:// URL (e.g. "redis://redis:6379/0"). It
// does not block or fail if Redis is unreachable at startup — every Push
// call carries its own short timeout and swallows errors, same as the
// Python side's circuit-breaker/fire-and-forget behavior.
func New(redisURL string) (*Queue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("visitqueue: parsing redis URL: %w", err)
	}
	return &Queue{client: redis.NewClient(opt)}, nil
}

func uuid4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func nowISO() string {
	// Matches the shape of Python's timezone.now().isoformat() closely
	// enough for django.utils.dateparse.parse_datetime() to accept it
	// (ISO-8601 with an explicit UTC offset); exact formatting fidelity
	// doesn't matter here the way it does for rendered page content.
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000+00:00")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// push RPUSHes payload (with a freshly-minted idempotency_key merged in,
// matching push_crawler_visit/push_archive_visit) onto key. Fire-and-forget:
// errors are swallowed, matching queue_crawler_visit/queue_archive_visit's
// "best effort, never block the request" contract.
func (q *Queue) push(key string, payload map[string]any) {
	if q == nil || q.client == nil {
		return
	}
	payload["idempotency_key"] = uuid4()
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_ = q.client.RPush(ctx, key, data).Err()
}

// PushCrawlerVisit ports apps/honeypot/views.py:_log_crawler's payload shape.
// botType/botGroup may be empty — the existing backfill_bot_types management
// command already exists to fill these in on rows that lack them.
func (q *Queue) PushCrawlerVisit(ipAddress, userAgent, host, path, referrer, trapType, queryString, botType, botGroup string) {
	q.push(crawlerQueueKey, map[string]any{
		"timestamp":    nowISO(),
		"ip_address":   ipAddress,
		"user_agent":   truncate(userAgent, 512),
		"host":         truncate(host, 253),
		"path":         truncate(path, 512),
		"referrer":     truncate(referrer, 256),
		"trap_type":    trapType,
		"query_string": truncate(queryString, 256),
		"bot_type":     botType,
		"bot_group":    botGroup,
	})
}

// PushArchiveVisit ports apps/honeypot/views.py:archive_trap's
// queue_archive_visit(data) payload shape.
func (q *Queue) PushArchiveVisit(ipAddress, userAgent string, year, month, day, depth int, slug string) {
	q.push(archiveQueueKey, map[string]any{
		"timestamp":  nowISO(),
		"ip_address": ipAddress,
		"user_agent": truncate(userAgent, 512),
		"year":       year,
		"month":      month,
		"day":        day,
		"slug":       truncate(slug, 512),
		"depth":      depth,
	})
}
