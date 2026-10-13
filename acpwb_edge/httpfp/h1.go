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

	mu         sync.Mutex
	heads      []Head
	lastStatus int // status of the last response head written (see Write)

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
	return h.Names
}

// Write notes the status of each response head the server writes. The
// last one tells what the server answered a request it rejected before
// any handler ran (400, 431, 501…), which it writes straight to the conn.
func (c *H1Conn) Write(p []byte) (int, error) {
	if code, ok := statusOf(p); ok {
		c.mu.Lock()
		c.lastStatus = code
		c.mu.Unlock()
	}
	return c.Conn.Write(p)
}

// Unserved returns the requests read from this connection that no handler
// took with NextHead — ones the server rejected — plus a final request
// whose head never completed, if bytes of one were waiting. Call it once
// the connection is done. status is the last response status written
// (see Write), 0 if none.
func (c *H1Conn) Unserved() (reqs []Head, status int) {
	c.mu.Lock()
	reqs, status = c.heads, c.lastStatus
	c.heads = nil
	c.mu.Unlock()
	// buf is the reading goroutine's, which is done by now.
	if len(c.buf) > 0 && !c.lost.Load() {
		h := parseHeadInfo(c.buf)
		h.Partial = true
		reqs = append(reqs, h)
	}
	return reqs, status
}

func statusOf(p []byte) (int, bool) {
	// "HTTP/1.1 400 "
	if len(p) < 13 || !bytes.HasPrefix(p, []byte("HTTP/1.")) || p[8] != ' ' {
		return 0, false
	}
	code, err := strconv.Atoi(string(p[9:12]))
	return code, err == nil && code >= 100 && code <= 999
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
		head := parseHeadInfo(c.buf[:end])
		c.mu.Lock()
		c.heads = append(c.heads, head)
		c.mu.Unlock()
		bodyLen, chunked := head.bodyLen, head.chunked
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

// Head is one request as the tap read it off the wire. Requests the server
// serves only contribute Names (via NextHead); the rest describe requests
// it rejected, for logging (see Unserved).
type Head struct {
	Method, Authority, Path, UserAgent string
	Names                              []string
	Partial                            bool // the head never completed

	bodyLen int64
	chunked bool
}

// ParseHead parses a request head, or the start of one: the request line,
// header names in order, Host and User-Agent.
func ParseHead(head []byte) Head { return parseHeadInfo(head) }

func parseHeadInfo(head []byte) Head {
	lines := strings.Split(strings.TrimRight(string(head), "\r\n"), "\r\n")
	h := Head{Names: []string{}}
	if parts := strings.SplitN(lines[0], " ", 3); len(parts) >= 2 {
		h.Method, h.Path = parts[0], parts[1]
	} else {
		h.Path = lines[0]
	}
	for _, line := range lines[1:] {
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		h.Names = append(h.Names, name)
		value = strings.TrimSpace(value)
		switch strings.ToLower(name) {
		case "content-length":
			h.bodyLen, _ = strconv.ParseInt(value, 10, 64)
		case "transfer-encoding":
			h.chunked = strings.Contains(strings.ToLower(value), "chunked")
		case "host":
			h.Authority = value
		case "user-agent":
			h.UserAgent = value
		}
	}
	return h
}
