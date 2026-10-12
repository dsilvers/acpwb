package edgequeue

import (
	"net"
	"testing"
	"time"
)

// With Redis unreachable, Push must never block a connection, and every
// record must be accounted for as dropped.
func TestPushNeverBlocksWhenRedisDown(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close() // nothing listening there now

	q, err := New("redis://"+addr+"/0", 1000)
	if err != nil {
		t.Fatal(err)
	}
	const n = 200_000 // > channel capacity
	start := time.Now()
	for i := range n {
		q.Push(map[string]int{"i": i})
	}
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("pushing took %v — Push must not block", el)
	}
	q.Close()
	if got := q.Dropped() + q.Pushed(); got != n {
		t.Fatalf("dropped+pushed = %d, want %d", got, n)
	}
	if q.Pushed() != 0 {
		t.Fatalf("pushed %d with no Redis", q.Pushed())
	}
}

func TestNilQueueIsANoop(t *testing.T) {
	var q *Queue
	q.Push(1)
	q.Close()
}
