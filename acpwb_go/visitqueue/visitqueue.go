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
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	crawlerQueueKey      = "acpwb:crawler_queue"
	archiveQueueKey      = "acpwb:archive_queue"
	fingerprintQueueKey  = "acpwb:fingerprint_queue"
	requestStreamChannel = "request_stream"
	firstSeenTTL         = 30 * 24 * time.Hour
)

// browserSignalHeaders must match apps.core.signal_capture.BROWSER_SIGNAL_HEADERS
// exactly — this list is the residential-proxy-detection candidate set, and a
// signal that means different things on the two backends writing
// RequestFingerprint isn't useful. Go's net/http canonicalizes header names
// (e.g. "sec-fetch-site" -> "Sec-Fetch-Site"), so lookups use Header.Get,
// which canonicalizes its argument too.
var browserSignalHeaders = []string{
	"Accept-Language",
	"Sec-Fetch-Site",
	"Sec-Fetch-Mode",
	"Sec-Fetch-Dest",
	"Sec-Fetch-User",
	"Sec-Ch-Ua",
	"Sec-Ch-Ua-Mobile",
	"Sec-Ch-Ua-Platform",
	"Upgrade-Insecure-Requests",
}

// BrowserHeadersPresent returns the sorted, comma-joined subset of
// browserSignalHeaders present on r — same shape and meaning as Django's
// apps.core.signal_capture.browser_headers_present().
func BrowserHeadersPresent(r *http.Request) string {
	var present []string
	for _, h := range browserSignalHeaders {
		if r.Header.Get(h) != "" {
			present = append(present, strings.ToLower(h))
		}
	}
	sort.Strings(present)
	return strings.Join(present, ",")
}

// Queue wraps a Redis client for pushing visit records. A nil *Queue (or one
// whose client is unreachable) causes PushVisit to silently no-op, mirroring
// push_crawler_visit()/push_archive_visit()'s "return False, caller falls back"
// contract — except this Go service has no local DB fallback, so a persistent
// Redis outage means visits simply aren't logged rather than blocking the
// response, which matches this service's only job: serve content fast.
type Queue struct {
	client *redis.Client

	// RequestFingerprint capture gate — same meaning and defaults as Django's
	// FINGERPRINT_CAPTURE_ENABLED / FINGERPRINT_SAMPLE_RATE /
	// FINGERPRINT_QUEUE_MAX settings, read from the same-named env vars.
	fingerprintEnabled    bool
	fingerprintSampleRate float64
	fingerprintQueueMax   int64
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
	q := &Queue{
		client:                redis.NewClient(opt),
		fingerprintEnabled:    envBool("FINGERPRINT_CAPTURE_ENABLED", false),
		fingerprintSampleRate: envFloat("FINGERPRINT_SAMPLE_RATE", 1.0),
		fingerprintQueueMax:   envInt("FINGERPRINT_QUEUE_MAX", 1_000_000),
	}
	return q, nil
}

func envBool(key string, def bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return v
	}
	return def
}

func envInt(key string, def int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v > 0 {
		return v
	}
	return def
}

// FingerprintSampled mirrors apps.core.signal_capture.fingerprint_sampled
// exactly: bucket = first 32 bits of md5(ip), big-endian, captured when
// bucket < rate * 2^32. Both backends must sample the same IPs.
func FingerprintSampled(ip string, rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	if rate <= 0.0 {
		return false
	}
	sum := md5.Sum([]byte(ip))
	return float64(binary.BigEndian.Uint32(sum[:4])) < rate*4294967296.0
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

	Fingerprint
}

// Fingerprint is the residential-proxy-detection raw signal for one request
// — see RequestFingerprint / apps.core.stream_middleware.py's
// _queue_fingerprint for the Django side of this same capture. Everything
// but BrowserHeadersPresent comes from headers acpwb_edge sets after
// stripping any client-supplied copies, since this service, like Django,
// only ever sees the plain HTTP/1.1 connection the edge proxies over.
type Fingerprint struct {
	ClientProtocol, TLSProtocol, TLSCipher string
	JA4, JA3, JA4H, H2Fingerprint          string
	BrowserHeadersPresent                  string
	// acpwb_edge's request ID (X-Request-Id), stored as request_id on the
	// CrawlerVisit/ArchiveVisit/RequestFingerprint rows so they join to
	// EdgeRequest. Empty unless it's a well-formed UUID.
	RequestID string
}

// FingerprintFrom reads r's Fingerprint off the acpwb_edge headers plus the
// fixed browser-header candidate set.
func FingerprintFrom(r *http.Request) Fingerprint {
	return Fingerprint{
		ClientProtocol:        r.Header.Get("X-Client-Protocol"),
		TLSProtocol:           r.Header.Get("X-Tls-Protocol"),
		TLSCipher:             r.Header.Get("X-Tls-Cipher"),
		JA4:                   r.Header.Get("X-Ja4"),
		JA3:                   r.Header.Get("X-Ja3"),
		JA4H:                  r.Header.Get("X-Ja4h"),
		H2Fingerprint:         r.Header.Get("X-H2-Fingerprint"),
		BrowserHeadersPresent: BrowserHeadersPresent(r),
		RequestID:             requestID(r.Header.Get("X-Request-Id")),
	}
}

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// requestID returns v if it's a UUID — anything else would fail the
// UUIDField on the Django side and lose the row.
func requestID(v string) string {
	if uuidRe.MatchString(v) {
		return v
	}
	return ""
}

// requestIDValue is the payload value for a request ID: null when absent.
func requestIDValue(id string) any {
	if id == "" {
		return nil
	}
	return id
}

// PushVisit sends the crawler-queue RPUSH, the optional archive-queue RPUSH,
// the request_stream PUBLISH, and the first-seen-IP check as a single
// pipelined round-trip, then (needing that check's result first) a second,
// separate RPUSH onto the fingerprint queue. Both round-trips happen inside
// the caller's `go vq.PushVisit(...)` goroutine, off the response path, so
// the extra round-trip costs nothing user-facing. Errors are swallowed —
// fire-and-forget, matching queue_crawler_visit/queue_archive_visit/
// RequestStreamMiddleware's "best effort, never block or fail the request"
// contract.
func (q *Queue) PushVisit(v Visit) {
	if q == nil || q.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Decided before anything touches Redis — including the per-IP
	// first-seen key — same as the Django middleware.
	captureFingerprint := q.fingerprintEnabled && FingerprintSampled(v.IPAddress, q.fingerprintSampleRate)

	pipe := q.client.Pipeline()
	var firstSeenCmd *redis.BoolCmd
	if captureFingerprint {
		firstSeenCmd = pipe.SetNX(ctx, "acpwb:ipseen:"+v.IPAddress, "1", firstSeenTTL)
	}

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
		"request_id":      requestIDValue(v.RequestID),
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
			"request_id":      requestIDValue(v.RequestID),
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

	if !captureFingerprint {
		return
	}

	// firstSeenCmd's result is only readable after Exec — a second, separate
	// call rather than folding into the pipeline above.
	firstSeenIP, _ := firstSeenCmd.Result()

	fingerprintPayload := map[string]any{
		"timestamp":               nowISO(),
		"ip_address":              v.IPAddress,
		"host":                    truncate(v.Host, 253),
		"user_agent":              truncate(v.UserAgent, 512),
		"referrer_present":        v.Referrer != "",
		"first_seen_ip":           firstSeenIP,
		"client_protocol":         truncate(v.ClientProtocol, 16),
		"tls_protocol":            truncate(v.TLSProtocol, 16),
		"tls_cipher":              truncate(v.TLSCipher, 64),
		"ja4":                     truncate(v.JA4, 40),
		"ja3":                     truncate(v.JA3, 32),
		"ja4h":                    truncate(v.JA4H, 64),
		"h2_fingerprint":          truncate(v.H2Fingerprint, 512),
		"browser_headers_present": truncate(v.BrowserHeadersPresent, 512),
		"request_id":              requestIDValue(v.RequestID),
		"idempotency_key":         uuid4(),
	}
	if data, err := json.Marshal(fingerprintPayload); err == nil {
		// Bounded like push_fingerprint_signal: LTRIM keeps only the newest
		// fingerprintQueueMax entries if the drain falls behind.
		fp := q.client.Pipeline()
		fp.RPush(ctx, fingerprintQueueKey, data)
		fp.LTrim(ctx, fingerprintQueueKey, -q.fingerprintQueueMax, -1)
		_, _ = fp.Exec(ctx)
	}
}

const (
	codeEmailQueueKey = "acpwb:code_email_queue"
	// Short: /compliance/ is endless, so each (address, path) key is a new
	// Redis key — 30 days of them was ~80M keys, and the queue 13M deep.
	codeEmailSeenTTL = 48 * time.Hour
)

// PushCodeEmails registers the honeypot programmer addresses shown on a
// legacy source-code page (and /compliance/ instructor addresses, keyed by
// course URL), so an inbound email to one of them can be traced
// back to the exact page it was harvested from (drained into
// honeypot.SourceCodeEmail by Django's drain_code_email_queue). A per
// (address, path) SETNX dedupes repeat crawls of the same page.
func (q *Queue) PushCodeEmails(host, path, ip string, addrs []string) {
	if q == nil || q.client == nil || len(addrs) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	pipe := q.client.Pipeline()
	cmds := make([]*redis.BoolCmd, len(addrs))
	for i, a := range addrs {
		cmds[i] = pipe.SetNX(ctx, "acpwb:codemail:"+a+"|"+path, "1", codeEmailSeenTTL)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return
	}
	push := q.client.Pipeline()
	n := 0
	for i, a := range addrs {
		if ok, _ := cmds[i].Result(); !ok {
			continue
		}
		data, err := json.Marshal(map[string]any{
			"address":    a,
			"host":       truncate(host, 253),
			"path":       truncate(path, 512),
			"ip_address": ip,
			"timestamp":  nowISO(),
		})
		if err == nil {
			push.RPush(ctx, codeEmailQueueKey, data)
			n++
		}
	}
	if n > 0 {
		_, _ = push.Exec(ctx)
	}
}

const (
	complianceQueueKey = "acpwb:compliance_queue"
	complianceSeenTTL  = 30 * 24 * time.Hour
	complianceQueueMax = 1_000_000
)

// ComplianceCompletion is one /compliance/ milestone — an attempt finished,
// a certificate served, or a learner identity typed in — drained into
// honeypot.ComplianceCompletion by Django's drain_compliance_queue. Ordinary
// question pages are not queued: their CrawlerVisit path already holds the
// whole answer trail (acpwb_go compliance-score decodes it).
type ComplianceCompletion struct {
	Event, Course, Answers, CertID string
	Attempt, Score                 int
	Passed                         bool
	LearnerName, LearnerEmail      string
	LearnerEmployeeID              string
	BotType                        string // botclassify label ("OpenAI GPTBot", "Other / Browser")
	ClientClass                    string // acpwb_edge's X-Traffic-Class (known_bot, browser, headless, …)
}

// PushComplianceCompletion queues a milestone at most once per
// (IP, course, attempt, event) every 30 days, so a crawler re-walking the
// same finished attempt doesn't add rows.
func (q *Queue) PushComplianceCompletion(r *http.Request, ip string, c ComplianceCompletion) {
	if q == nil || q.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	key := fmt.Sprintf("acpwb:compliance:%s|%s|%d|%s", ip, c.Course, c.Attempt, c.Event)
	if ok, err := q.client.SetNX(ctx, key, "1", complianceSeenTTL).Result(); err != nil || !ok {
		return
	}
	data, err := json.Marshal(map[string]any{
		"event":               c.Event,
		"course":              truncate(c.Course, 80),
		"attempt":             c.Attempt,
		"answers":             truncate(c.Answers, 32),
		"score":               c.Score,
		"passed":              c.Passed,
		"cert_id":             c.CertID,
		"learner_name":        truncate(c.LearnerName, 120),
		"learner_email":       truncate(c.LearnerEmail, 254),
		"learner_employee_id": truncate(c.LearnerEmployeeID, 64),
		"bot_type":            truncate(c.BotType, 64),
		"client_class":        truncate(c.ClientClass, 16),
		"ip_address":          ip,
		"user_agent":          truncate(r.UserAgent(), 512),
		"host":                truncate(r.Host, 253),
		"request_id":          requestIDValue(requestID(r.Header.Get("X-Request-Id"))),
		"timestamp":           nowISO(),
	})
	if err != nil {
		return
	}
	pipe := q.client.Pipeline()
	pipe.RPush(ctx, complianceQueueKey, data)
	pipe.LTrim(ctx, complianceQueueKey, -complianceQueueMax, -1)
	_, _ = pipe.Exec(ctx)
}

// ── /compliance/ dashboard counters ─────────────────────────────────────

// ComplianceStat is one served /compliance/ page, folded into per-day Redis
// counters for the staff dashboard instead of a row per page view (the
// CrawlerVisit row already exists; these are cheap aggregates of it).
type ComplianceStat struct {
	Kind, Topic, Agency, Course, Answers, Exhibit string
	Attempt                                       int
	IP, BotType, ClientClass                      string
}

const (
	complianceStatTTL      = 45 * 24 * time.Hour
	complianceStatMaxDepth = 25
)

// ComplianceStatKey is the Redis key for one day's counter of one kind
// (Django's dashboard reads the same names: apps/core/compliance_stats.py).
func ComplianceStatKey(day, name string) string { return "acpwb:cstat:" + day + ":" + name }

// PushComplianceStats bumps the day's counters for one served page:
//
//	pages      hash  kind and kind|client_class → hits
//	bots       hash  bot type → question-page hits
//	topics     hash  course-type slug → hits
//	exhibits   hash  exhibit kind → question-page hits
//	depth      hash  answers so far → question/results hits
//	agencies   zset  agency slug → hits
//	ips        HLL   distinct IPs
//	att:<d>    HLL   distinct (ip, course, attempt) that reached depth d
//	fan:<d>    HLL   distinct (ip, course, attempt, answers) at depth d
//
// fan:<d> / att:<d> is the average number of answer paths each attempt
// explored at depth d: 1 for a crawler walking one path, up to 4^d for one
// following every choice.
func (q *Queue) PushComplianceStats(st ComplianceStat) {
	if q == nil || q.client == nil || st.Kind == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	day := time.Now().UTC().Format("20060102")
	k := func(name string) string { return ComplianceStatKey(day, name) }
	class := st.ClientClass
	if class == "" {
		class = "unclassified"
	}
	pipe := q.client.Pipeline()
	touched := []string{k("pages"), k("ips")}
	pipe.HIncrBy(ctx, k("pages"), st.Kind, 1)
	pipe.HIncrBy(ctx, k("pages"), st.Kind+"|"+class, 1)
	pipe.PFAdd(ctx, k("ips"), st.IP)
	if st.Topic != "" {
		pipe.HIncrBy(ctx, k("topics"), st.Topic, 1)
		touched = append(touched, k("topics"))
	}
	if st.Agency != "" {
		pipe.ZIncrBy(ctx, k("agencies"), 1, st.Agency)
		touched = append(touched, k("agencies"))
	}
	if st.Kind == "question" || st.Kind == "results" {
		d := len(st.Answers)
		pipe.HIncrBy(ctx, k("depth"), fmt.Sprint(d), 1)
		touched = append(touched, k("depth"))
		if d <= complianceStatMaxDepth {
			att := fmt.Sprintf("%s|%s|%d", st.IP, st.Course, st.Attempt)
			pipe.PFAdd(ctx, k(fmt.Sprintf("att:%d", d)), att)
			pipe.PFAdd(ctx, k(fmt.Sprintf("fan:%d", d)), att+"|"+st.Answers)
			touched = append(touched, k(fmt.Sprintf("att:%d", d)), k(fmt.Sprintf("fan:%d", d)))
		}
	}
	if st.Kind == "question" {
		pipe.HIncrBy(ctx, k("bots"), st.BotType, 1)
		touched = append(touched, k("bots"))
		if st.Exhibit != "" {
			pipe.HIncrBy(ctx, k("exhibits"), st.Exhibit, 1)
			touched = append(touched, k("exhibits"))
		}
	}
	for _, key := range touched {
		pipe.Expire(ctx, key, complianceStatTTL)
	}
	_, _ = pipe.Exec(ctx)
}
