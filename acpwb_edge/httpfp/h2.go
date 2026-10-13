package httpfp

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	h2Preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

	frameHeaders      = 0x1
	framePriority     = 0x2
	frameRSTStream    = 0x3
	frameSettings     = 0x4
	frameGoAway       = 0x7
	frameWindowUpdate = 0x8
	frameContinuation = 0x9

	flagAck        = 0x1
	flagEndHeaders = 0x4
	flagPadded     = 0x8
	flagPriority   = 0x20

	// The edge advertises SETTINGS_MAX_FRAME_SIZE 16 KiB, so a bigger frame
	// is one the server rejects anyway; the tap gives up rather than
	// buffer it.
	maxTapFrame = 16<<10 + 9
	// Header blocks are bounded by the server's MaxHeaderBytes (64 KiB).
	maxHeaderBlock = 64 << 10
	// Streams decoded but not yet claimed by a handler. Requests the
	// server refuses never get claimed; this bounds that leftover.
	maxUnclaimed = 512

	// Must equal the server's SETTINGS_HEADER_TABLE_SIZE (x/net/http2's
	// default MaxDecoderHeaderTableSize), or this decoder's dynamic table
	// would diverge from the server's.
	headerTableSize = 4096
)

type h2Stream struct {
	id                      uint32
	names                   []string
	method, authority, path string
}

// H2Conn watches the decrypted client→server bytes of an HTTP/2 connection
// with a frame parser of its own, running alongside the real server's. It
// produces the connection's Akamai-style fingerprint
//
//	SETTINGS|WINDOW_UPDATE|PRIORITY|pseudo-header order
//	e.g. 1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p
//
// from the frames sent before the first HEADERS, and, by decoding every
// header block with an HPACK decoder kept in step with the server's, the
// header names of each request in wire order.
//
// net/http doesn't tell handlers their stream ID, so a handler finds its
// headers with Claim, which takes the lowest unclaimed stream matching its
// method, authority and path. Two concurrent requests that match on all
// three carry the same headers in practice, so mixing them up is harmless.
type H2Conn struct {
	net.Conn

	mu   sync.Mutex
	buf  []byte
	skip int // payload bytes of an uninteresting frame still to pass over
	pre  bool
	lost bool

	dec *hpack.Decoder

	settings     []string
	settingsSeen bool
	windowUpdate string
	priorities   []string
	fingerprint  string

	blockStream uint32
	block       []byte
	inBlock     bool

	streams []*h2Stream

	// Abuse counters — counted even after header capture gives up.
	stats     H2Stats
	maxStream uint32

	onAbuse func() // see SetRapidResetLimit
	abused  bool

	// Server→client side: just enough parsing to see GOAWAYs.
	out     []byte // partial frame header (+ GOAWAY payload prefix)
	outSkip int
}

// H2Stats counts what an HTTP/2 client did on one connection.
type H2Stats struct {
	Streams       int    // streams opened (HEADERS on a new stream ID)
	Resets        int    // RST_STREAM frames sent (rapid reset: Resets ≈ Streams)
	Continuations int    // CONTINUATION frames sent (CONTINUATION flood)
	Goaway        string // error code of a GOAWAY the server sent, e.g. ENHANCE_YOUR_CALM
}

// Rapid reset (CVE-2023-44487): a client opening streams and immediately
// resetting them makes the server start and cancel a handler for each, at
// no cost to itself. Real clients reset a handful of streams, never
// hundreds — so past rapidResetMin resets, at least rapidResetRatio of all
// streams opened, the connection is ended.
const (
	rapidResetMin   = 200
	rapidResetRatio = 0.9
)

// SetRapidResetLimit arranges for fn to run (once, from the reading
// goroutine) when the client crosses the rapid-reset threshold.
func (c *H2Conn) SetRapidResetLimit(fn func()) {
	c.mu.Lock()
	c.onAbuse = fn
	c.mu.Unlock()
}

func (c *H2Conn) checkAbuse() {
	if c.abused || c.onAbuse == nil || c.stats.Resets < rapidResetMin {
		return
	}
	if float64(c.stats.Resets) >= rapidResetRatio*float64(c.stats.Streams) {
		c.abused = true
		go c.onAbuse() // not under c.mu: it closes the connection
	}
}

// Counters returns the connection's H2Stats.
func (c *H2Conn) Counters() H2Stats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

func (c *H2Conn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if n > 0 {
		c.mu.Lock()
		c.safeFeedOut(p[:n])
		c.mu.Unlock()
	}
	return n, err
}

func (c *H2Conn) safeFeedOut(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.outSkip, c.out = 1<<62, nil // stop parsing this side
			notePanic("h2 tap (out)", r)
		}
	}()
	c.feedOut(data)
}

// feedOut follows the server's frames (it never sends a preface) and
// records the error code of the first GOAWAY.
func (c *H2Conn) feedOut(data []byte) {
	for len(data) > 0 {
		if c.outSkip > 0 {
			n := min(c.outSkip, len(data))
			c.outSkip -= n
			data = data[n:]
			continue
		}
		need := 9
		if len(c.out) >= 9 && c.out[3] == frameGoAway {
			need = 9 + 8 // last-stream-id + error code
		}
		take := min(need-len(c.out), len(data))
		c.out = append(c.out, data[:take]...)
		data = data[take:]
		if len(c.out) < need {
			continue
		}
		length := int(c.out[0])<<16 | int(c.out[1])<<8 | int(c.out[2])
		if c.out[3] == frameGoAway && need == 9 {
			if length < 8 {
				c.outSkip, c.out = length, c.out[:0]
			}
			continue // collect the payload prefix next
		}
		if c.out[3] == frameGoAway {
			if c.stats.Goaway == "" {
				c.stats.Goaway = http2.ErrCode(binary.BigEndian.Uint32(c.out[13:17])).String()
			}
			c.outSkip = length - 8
		} else {
			c.outSkip = length
		}
		c.out = c.out[:0]
	}
}

// NewH2Conn wraps c, which must deliver the decrypted HTTP/2 stream.
func NewH2Conn(c net.Conn) *H2Conn {
	return &H2Conn{Conn: c, dec: hpack.NewDecoder(headerTableSize, nil)}
}

func (c *H2Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.mu.Lock()
		if !c.lost {
			c.safeFeed(p[:n])
		}
		c.mu.Unlock()
	}
	return n, err
}

// safeFeed runs the parser on client-controlled bytes. A bug in it must
// only ever cost this connection its fingerprint — a panic here would
// otherwise take down the whole process from the server's read goroutine.
func (c *H2Conn) safeFeed(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.giveUp()
			notePanic("h2 tap", r)
		}
	}()
	c.feed(data)
}

// Fingerprint returns the Akamai-style HTTP/2 fingerprint, or "" if the
// client hasn't sent its first HEADERS frame yet (or capture was lost).
func (c *H2Conn) Fingerprint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fingerprint
}

// CaptureLost reports whether the tap gave up on this connection.
func (c *H2Conn) CaptureLost() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lost
}

// Claim returns the header names, in wire order, of the request on this
// connection with the given method, :authority (or Host) and :path, and
// forgets it. Returns nil if there's no such request.
func (c *H2Conn) Claim(method, authority, path string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, s := range c.streams {
		if s.method == method && s.path == path && strings.EqualFold(s.authority, authority) {
			c.streams = append(c.streams[:i], c.streams[i+1:]...)
			return s.names
		}
	}
	return nil
}

func (c *H2Conn) giveUp() {
	c.lost = true
	c.buf, c.block, c.streams = nil, nil, nil
}

func (c *H2Conn) feed(data []byte) {
	c.buf = append(c.buf, data...)
	if !c.pre {
		if len(c.buf) < len(h2Preface) {
			if !strings.HasPrefix(h2Preface, string(c.buf)) {
				c.giveUp()
			}
			return
		}
		if string(c.buf[:len(h2Preface)]) != h2Preface {
			c.giveUp()
			return
		}
		c.pre = true
		c.buf = c.buf[len(h2Preface):]
	}
	for !c.lost {
		if c.skip > 0 {
			n := min(c.skip, len(c.buf))
			c.skip -= n
			c.buf = c.buf[n:]
			if c.skip > 0 {
				break
			}
			continue
		}
		if len(c.buf) < 9 {
			break
		}
		length := int(c.buf[0])<<16 | int(c.buf[1])<<8 | int(c.buf[2])
		typ, flags := c.buf[3], c.buf[4]
		stream := binary.BigEndian.Uint32(c.buf[5:9]) & 0x7fffffff
		if length+9 > maxTapFrame {
			c.giveUp()
			return
		}
		switch typ {
		case frameHeaders, framePriority, frameSettings, frameWindowUpdate, frameContinuation:
		default:
			if typ == frameRSTStream {
				c.stats.Resets++
				c.checkAbuse()
			}
			c.skip = length
			c.buf = c.buf[9:]
			continue
		}
		if len(c.buf) < 9+length {
			break
		}
		payload := c.buf[9 : 9+length]
		c.frame(typ, flags, stream, payload)
		if c.lost {
			return // giveUp dropped the buffer
		}
		c.buf = c.buf[9+length:]
	}
	// Don't let a consumed prefix pin a large backing array.
	if len(c.buf) == 0 {
		c.buf = nil
	}
}

func (c *H2Conn) frame(typ, flags byte, stream uint32, p []byte) {
	switch typ {
	case frameSettings:
		if stream != 0 || flags&flagAck != 0 || c.settingsSeen {
			return
		}
		c.settingsSeen = true
		for len(p) >= 6 {
			id := binary.BigEndian.Uint16(p[0:2])
			val := binary.BigEndian.Uint32(p[2:6])
			c.settings = append(c.settings, strconv.Itoa(int(id))+":"+strconv.FormatUint(uint64(val), 10))
			p = p[6:]
		}
	case frameWindowUpdate:
		if stream == 0 && c.fingerprint == "" && c.windowUpdate == "" && len(p) == 4 {
			c.windowUpdate = strconv.FormatUint(uint64(binary.BigEndian.Uint32(p)&0x7fffffff), 10)
		}
	case framePriority:
		if c.fingerprint == "" && len(p) == 5 {
			c.priorities = append(c.priorities, priorityString(stream, p))
		}
	case frameHeaders:
		if stream > c.maxStream {
			c.maxStream = stream
			c.stats.Streams++
		}
		if c.inBlock {
			c.giveUp() // HEADERS inside an unfinished block: protocol error
			return
		}
		if flags&flagPadded != 0 {
			if len(p) < 1 || int(p[0]) > len(p)-1 {
				c.giveUp()
				return
			}
			p = p[1 : len(p)-int(p[0])]
		}
		if flags&flagPriority != 0 {
			if len(p) < 5 {
				c.giveUp()
				return
			}
			p = p[5:]
		}
		c.blockStream, c.inBlock = stream, true
		c.block = append(c.block[:0], p...)
		if flags&flagEndHeaders != 0 {
			c.endBlock()
		}
	case frameContinuation:
		c.stats.Continuations++
		if !c.inBlock || stream != c.blockStream {
			c.giveUp()
			return
		}
		if len(c.block)+len(p) > maxHeaderBlock {
			c.giveUp()
			return
		}
		c.block = append(c.block, p...)
		if flags&flagEndHeaders != 0 {
			c.endBlock()
		}
	}
}

func priorityString(stream uint32, p []byte) string {
	dep := binary.BigEndian.Uint32(p[0:4])
	excl := "0"
	if dep&0x80000000 != 0 {
		excl = "1"
	}
	// Weight is sent as weight-1.
	return strconv.FormatUint(uint64(stream), 10) + ":" + excl + ":" +
		strconv.FormatUint(uint64(dep&0x7fffffff), 10) + ":" + strconv.Itoa(int(p[4])+1)
}

// endBlock decodes a complete header block. Every block is decoded, even
// trailers and ones for streams already seen, to keep the HPACK dynamic
// table in step with the server's.
func (c *H2Conn) endBlock() {
	c.inBlock = false
	fields, err := c.dec.DecodeFull(c.block)
	c.block = c.block[:0]
	if err != nil {
		c.giveUp()
		return
	}
	s := &h2Stream{id: c.blockStream, names: make([]string, 0, len(fields))}
	var pseudo []string
	var host string
	for _, f := range fields {
		s.names = append(s.names, f.Name)
		switch f.Name {
		case ":method":
			s.method = f.Value
			pseudo = append(pseudo, "m")
		case ":authority":
			s.authority = f.Value
			pseudo = append(pseudo, "a")
		case ":scheme":
			pseudo = append(pseudo, "s")
		case ":path":
			s.path = f.Value
			pseudo = append(pseudo, "p")
		case "host":
			host = f.Value
		}
	}
	if s.authority == "" {
		s.authority = host
	}
	if c.fingerprint == "" {
		c.fingerprint = strings.Join(c.settings, ";") + "|" +
			orDefault(c.windowUpdate, "00") + "|" +
			orDefault(strings.Join(c.priorities, ","), "0") + "|" +
			strings.Join(pseudo, ",")
	}
	if s.method == "" {
		return // trailers
	}
	for _, old := range c.streams {
		if old.id == s.id {
			return
		}
	}
	if len(c.streams) >= maxUnclaimed {
		c.streams = c.streams[1:]
	}
	c.streams = append(c.streams, s)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
