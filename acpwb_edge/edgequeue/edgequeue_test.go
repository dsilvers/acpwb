package edgequeue

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
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

// fakeRedis answers just enough RESP for a pipelined RPUSH + LTRIM and counts
// the values pushed.
func fakeRedis(t *testing.T, l net.Listener, pushed *atomic.Int64) {
	t.Helper()
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				r := bufio.NewReader(c)
				for {
					args, err := readCommand(r)
					if err != nil {
						return
					}
					reply := "+OK\r\n"
					switch strings.ToLower(args[0]) {
					case "hello":
						reply = "-ERR unknown command\r\n"
					case "rpush":
						pushed.Add(int64(len(args) - 2))
						reply = fmt.Sprintf(":%d\r\n", pushed.Load())
					}
					if _, err := c.Write([]byte(reply)); err != nil {
						return
					}
				}
			}()
		}
	}()
}

func readCommand(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}
	args := make([]string, n)
	for i := range args {
		if _, err := r.ReadString('\n'); err != nil { // $<len>
			return nil, err
		}
		if args[i], err = r.ReadString('\n'); err != nil {
			return nil, err
		}
		args[i] = strings.TrimSpace(args[i])
	}
	return args, nil
}

// A short Redis failure must cost nothing: the failed batch is retried and
// what arrives meanwhile waits in the channel.
func TestFailedPushIsRetriedWithoutLoss(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close() // down at first

	q, err := New("redis://"+addr+"/0", 1000)
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	q.Push(ConnQueue, 0)
	deadline := time.Now().Add(10 * time.Second)
	for q.PushFailures(ConnQueue) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("no failed push recorded")
		}
		time.Sleep(10 * time.Millisecond)
	}
	const n = 1000
	for i := 1; i < n; i++ {
		q.Push(ConnQueue, i)
	}

	if l, err = net.Listen("tcp", addr); err != nil {
		t.Fatal(err)
	}
	var pushed atomic.Int64
	fakeRedis(t, l, &pushed)
	for q.Pushed(ConnQueue) < n {
		if time.Now().After(deadline) {
			t.Fatalf("pushed %d of %d", q.Pushed(ConnQueue), n)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := pushed.Load(); got != n {
		t.Fatalf("redis received %d, want %d", got, n)
	}
	if d := q.Dropped(ConnQueue); d != 0 {
		t.Fatalf("dropped %d", d)
	}
}

func TestNilQueueIsANoop(t *testing.T) {
	var q *Queue
	q.Push(ConnQueue, 1)
	q.Close()
}
