package compliance

import "sync"

// cache is a small bounded memo for generated values. The key spaces here
// are effectively unbounded, so instead of tracking recency it is simply
// emptied when full; crawlers walk one course/attempt at a time, so the
// working set refills almost immediately.
type cache[T any] struct {
	mu  sync.Mutex
	max int
	m   map[string]T
}

func newCache[T any](max int) *cache[T] { return &cache[T]{max: max, m: map[string]T{}} }

func (c *cache[T]) get(k string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[k]
	return v, ok
}

func (c *cache[T]) put(k string, v T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.m) >= c.max {
		c.m = map[string]T{}
	}
	c.m[k] = v
}
