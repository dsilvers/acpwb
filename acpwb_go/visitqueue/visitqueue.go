// Package visitqueue ports the request-path side of
// apps/core/crawler_queue.py: RPUSH a JSON payload onto the same Redis
// lists (acpwb:crawler_queue / acpwb:archive_queue) that Django's
// drain_crawler_queue / drain_archive_queue management commands already
// drain into CrawlerVisit / ArchiveVisit rows. It also PUBLISHes to the
// same "request_stream" Redis pub/sub channel apps/core/stream_middleware.py
// uses, which ws_service relays to the live dashboard and
// botseed_processor consumes for entropy — acpwb_go bypasses Django's
// middleware chain entirely, so without this, none of its traffic (the
// majority of the site's real traffic after the archive/policy cutover)
// would appear in either.
//
// All of a request's Redis writes (up to two queue RPUSHes plus the stream
// PUBLISH) are sent as a single pipelined round-trip via PushVisit, rather
// than as separate commands — this matters given how much request volume
// this service handles; see deploy/README.md's Redis tcp-backlog incident
// for why extra avoidable Redis round-trips are worth caring about here.
package visitqueue

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	crawlerQueueKey      = "acpwb:crawler_queue"
	archiveQueueKey      = "acpwb:archive_queue"
	requestStreamChannel = "request_stream"
)

// Queue wraps a Redis client for pushing visit records. A nil *Queue (or one
// whose client is unreachable) causes PushVisit to silently no-op, mirroring
// push_crawler_visit()/push_archive_visit()'s "return False, caller falls back"
// contract — except this Go service has no local DB fallback, so a persistent
// Redis outage means visits simply aren't logged rather than blocking the
// response, which matches this service's only job: serve content fast.
type Queue struct {
	client *redis.Client
}

// New creates a Queue from a redis:// URL (e.g. "redis://redis:6379/0"). It
// does not block or fail if Redis is unreachable at startup — every push
// carries its own short timeout and swallows errors, same as the Python
// side's circuit-breaker/fire-and-forget behavior.
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

// censorIP mirrors stream_middleware.py's last-octet censoring for IPv4
// (e.g. "203.0.113.42" -> "203.0.113.xxx"); IPv6 passes through unchanged,
// same as the Python side.
func censorIP(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ipStr
	}
	v4 := ip.To4()
	if v4 == nil {
		return ipStr
	}
	return fmt.Sprintf("%d.%d.%d.xxx", v4[0], v4[1], v4[2])
}

// ArchiveInfo carries the extra fields archive_queue's payload needs, beyond
// what CrawlerVisit/request_stream already have. A nil *ArchiveInfo on Visit
// means "not an archive-trap request" — no ArchiveVisit row is queued.
type ArchiveInfo struct {
	Year, Month, Day, Depth int
	Slug                    string
}

// Visit carries everything about one served request needed to populate the
// crawler queue, the (optional) archive queue, and the live request_stream
// — gathered after rendering completes so Status/ResponseBytes/ResponseMs
// reflect what was actually sent, matching stream_middleware.py's
// end-of-request measurement.
type Visit struct {
	IPAddress, UserAgent, Host, Path, Referrer string
	TrapType, QueryString                      string
	BotType, BotGroup                          string
	Method                                     string
	Status                                     int
	ResponseBytes                              int
	ResponseMs                                 int64
	Archive                                    *ArchiveInfo
}

// PushVisit sends the crawler-queue RPUSH, the optional archive-queue RPUSH,
// and the request_stream PUBLISH as a single pipelined round-trip. Errors
// are swallowed — fire-and-forget, matching queue_crawler_visit/
// queue_archive_visit/RequestStreamMiddleware's "best effort, never block
// or fail the request" contract.
func (q *Queue) PushVisit(v Visit) {
	if q == nil || q.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pipe := q.client.Pipeline()

	crawlerPayload := map[string]any{
		"timestamp":       nowISO(),
		"ip_address":      v.IPAddress,
		"user_agent":      truncate(v.UserAgent, 512),
		"host":            truncate(v.Host, 253),
		"path":            truncate(v.Path, 512),
		"referrer":        truncate(v.Referrer, 256),
		"trap_type":       v.TrapType,
		"query_string":    truncate(v.QueryString, 256),
		"bot_type":        v.BotType,
		"bot_group":       v.BotGroup,
		"idempotency_key": uuid4(),
	}
	if data, err := json.Marshal(crawlerPayload); err == nil {
		pipe.RPush(ctx, crawlerQueueKey, data)
	}

	if v.Archive != nil {
		archivePayload := map[string]any{
			"timestamp":       nowISO(),
			"ip_address":      v.IPAddress,
			"user_agent":      truncate(v.UserAgent, 512),
			"year":            v.Archive.Year,
			"month":           v.Archive.Month,
			"day":             v.Archive.Day,
			"slug":            truncate(v.Archive.Slug, 512),
			"depth":           v.Archive.Depth,
			"idempotency_key": uuid4(),
		}
		if data, err := json.Marshal(archivePayload); err == nil {
			pipe.RPush(ctx, archiveQueueKey, data)
		}
	}

	streamPayload := map[string]any{
		"ip":             censorIP(v.IPAddress),
		"host":           v.Host,
		"path":           v.Path,
		"timestamp":      nowISO(),
		"response_ms":    v.ResponseMs,
		"response_bytes": v.ResponseBytes,
		"method":         v.Method,
		"status":         v.Status,
		"user_agent":     v.UserAgent,
		"bot_type":       v.BotType,
		"bot_group":      v.BotGroup,
	}
	if data, err := json.Marshal(streamPayload); err == nil {
		pipe.Publish(ctx, requestStreamChannel, data)
	}

	_, _ = pipe.Exec(ctx)
}
