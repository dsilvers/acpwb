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
		q.Push(RequestQueue, map[string]int{"i": i})
		if i%10 == 0 {
			q.Push(ConnQueue, map[string]int{"i": i})
		}
	}
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("pushing took %v — Push must not block", el)
	}
	q.Close()
	for key, want := range map[string]int64{RequestQueue: n, ConnQueue: n / 10} {
		if got := q.Dropped(key) + q.Pushed(key); got != want {
			t.Fatalf("%s: dropped+pushed = %d, want %d", key, got, want)
		}
		if q.Pushed(key) != 0 {
			t.Fatalf("%s: pushed %d with no Redis", key, q.Pushed(key))
		}
	}
}

func TestNilQueueIsANoop(t *testing.T) {
	var q *Queue
	q.Push(ConnQueue, 1)
	q.Close()
}
