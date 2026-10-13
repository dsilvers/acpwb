package httpfp

import (
	"bytes"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const maxHeadBytes = 64 << 10

// H1Conn records the order of request header names as an HTTP/1.x
// connection's bytes are read. HTTP/1.x requests on one connection are
// handled in order, so a FIFO of heads lines up with the requests the
// server hands to its handler: each handler call takes one with NextHead.
//
// Ported from the openbeverage.com demo's headerConn. Capture stops for the
// rest of the connection after a chunked request body (the scanner doesn't
// decode chunking) or an oversized head; CaptureLost reports that.
type H1Conn struct {
	net.Conn

	mu    sync.Mutex
	heads [][]string

	// Only touched by the server's reading goroutine.
	buf  []byte
	skip int64 // body bytes still to pass over

	lost atomic.Bool
}

// NewH1Conn wraps c, which must deliver the decrypted request stream.
func NewH1Conn(c net.Conn) *H1Conn { return &H1Conn{Conn: c} }

func (c *H1Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 && !c.lost.Load() {
		c.safeScan(p[:n])
	}
	return n, err
}

// safeScan: see H2Conn.safeFeed.
func (c *H1Conn) safeScan(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.lost.Store(true)
			c.buf = nil
			notePanic("h1 tap", r)
		}
	}()
	c.scan(data)
}

// NextHead returns the header names of the next request read from this
// connection, in wire order, or nil if capture had stopped.
func (c *H1Conn) NextHead() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.heads) == 0 {
		return nil
	}
	h := c.heads[0]
	c.heads = c.heads[1:]
	return h
}

// CaptureLost reports whether header capture stopped on this connection.
func (c *H1Conn) CaptureLost() bool { return c.lost.Load() }

func (c *H1Conn) scan(data []byte) {
	for len(data) > 0 && !c.lost.Load() {
		if c.skip > 0 {
			n := min(int64(len(data)), c.skip)
			c.skip -= n
			data = data[n:]
			continue
		}
		if len(c.buf) == 0 {
			// Clients may send stray CRLFs between requests.
			data = bytes.TrimLeft(data, "\r\n")
			if len(data) == 0 {
				return
			}
		}
		start := max(0, len(c.buf)-3)
		c.buf = append(c.buf, data...)
		end := bytes.Index(c.buf[start:], []byte("\r\n\r\n"))
		if end < 0 {
			if len(c.buf) > maxHeadBytes {
				c.lost.Store(true)
				c.buf = nil
			}
			return
		}
		end += start + 4
		names, bodyLen, chunked := parseHead(c.buf[:end])
		c.mu.Lock()
		c.heads = append(c.heads, names)
		c.mu.Unlock()
		data = c.buf[end:]
		c.buf = nil
		if chunked {
			c.lost.Store(true)
			return
		}
		c.skip = bodyLen
		// Copy the remainder: data aliases c.buf's backing array, which the
		// next append would otherwise overwrite while we're still reading it.
		data = slices.Clone(data)
	}
}

// parseHead returns the header names in order, plus how to skip the body.
func parseHead(head []byte) (names []string, bodyLen int64, chunked bool) {
	lines := strings.Split(strings.TrimRight(string(head), "\r\n"), "\r\n")
	names = []string{}
	for _, line := range lines[1:] {
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		names = append(names, name)
		value = strings.TrimSpace(value)
		switch strings.ToLower(name) {
		case "content-length":
			bodyLen, _ = strconv.ParseInt(value, 10, 64)
		case "transfer-encoding":
			chunked = strings.Contains(strings.ToLower(value), "chunked")
		}
	}
	return names, bodyLen, chunked
}
