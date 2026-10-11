package visitqueue

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestPushCodeEmailsNilSafe(t *testing.T) {
	var q *Queue
	q.PushCodeEmails("acpwb.com", "/p/", "1.2.3.4", []string{"a.b@acpwb.com"}) // must not panic
}

// TestPushCodeEmailsRedis runs against a real Redis when ACPWB_TEST_REDIS is
// set (e.g. redis://localhost:6379/15).
func TestPushCodeEmailsRedis(t *testing.T) {
	url := os.Getenv("ACPWB_TEST_REDIS")
	if url == "" {
		t.Skip("ACPWB_TEST_REDIS not set")
	}
	q, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	path := "/archive/2022/04/06/test-codeemail/source-code/"
	addrs := []string{"xiao.whitehead@acpwb.com", "gail.hargrove@acpwb.com"}
	q.client.Del(ctx, codeEmailQueueKey)
	for _, a := range addrs {
		q.client.Del(ctx, "acpwb:codemail:"+a+"|"+path)
	}
	q.PushCodeEmails("acpwb.com", path, "10.0.0.1", addrs)
	q.PushCodeEmails("acpwb.com", path, "10.0.0.2", addrs) // deduped
	items, err := q.client.LRange(ctx, codeEmailQueueKey, 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("queued %d items, want 2 (second push deduped)", len(items))
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(items[0]), &m); err != nil || m["path"] != path || m["ip_address"] != "10.0.0.1" {
		t.Fatalf("bad payload %v (%v)", m, err)
	}
	q.client.Del(ctx, codeEmailQueueKey)
}
