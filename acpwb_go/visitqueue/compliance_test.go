package visitqueue

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPushComplianceCompletionNilSafe(t *testing.T) {
	var q *Queue
	q.PushComplianceCompletion(httptest.NewRequest("GET", "/compliance/", nil), "1.2.3.4", ComplianceCompletion{Event: "completed"})
}

// Runs against a real Redis when ACPWB_TEST_REDIS is set (e.g. redis://localhost:6379/15).
func TestPushComplianceCompletionRedis(t *testing.T) {
	url := os.Getenv("ACPWB_TEST_REDIS")
	if url == "" {
		t.Skip("ACPWB_TEST_REDIS not set")
	}
	q, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	c := ComplianceCompletion{Event: "completed", Course: "insider-trading-awareness-48213", Attempt: 2, Answers: "abcd", Score: 75, Passed: true, CertID: "0123abcd"}
	q.client.Del(ctx, complianceQueueKey, "acpwb:compliance:10.0.0.1|insider-trading-awareness-48213|2|completed")
	r := httptest.NewRequest("GET", "/compliance/insider-trading-awareness-48213/attempt-2/abcd/", nil)
	r.Header.Set("X-Request-Id", "0f8fad5b-d9cb-469f-a165-70867728950e")
	q.PushComplianceCompletion(r, "10.0.0.1", c)
	q.PushComplianceCompletion(r, "10.0.0.1", c) // deduped
	items, err := q.client.LRange(ctx, complianceQueueKey, 0, -1).Result()
	if err != nil || len(items) != 1 {
		t.Fatalf("queued %d items (%v), want 1", len(items), err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(items[0]), &m); err != nil || m["course"] != c.Course || m["passed"] != true || m["request_id"] != "0f8fad5b-d9cb-469f-a165-70867728950e" {
		t.Fatalf("bad payload %v (%v)", m, err)
	}
	q.client.Del(ctx, complianceQueueKey)
}
