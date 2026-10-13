package httpfp

import (
	"bytes"
	"testing"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func h2Seed() []byte {
	var wire bytes.Buffer
	wire.WriteString(http2.ClientPreface)
	fr := http2.NewFramer(&wire, nil)
	fr.WriteSettings(http2.Setting{ID: http2.SettingInitialWindowSize, Val: 6291456})
	fr.WriteWindowUpdate(0, 15663105)
	fr.WritePriority(3, http2.PriorityParam{Weight: 200})
	var hb bytes.Buffer
	enc := hpack.NewEncoder(&hb)
	for _, f := range [][2]string{{":method", "GET"}, {":authority", "acpwb.com"}, {":scheme", "https"}, {":path", "/"}, {"accept", "*/*"}} {
		enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	block := hb.Bytes()
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, BlockFragment: block[:4], PadLength: 2})
	fr.WritePriority(1, http2.PriorityParam{Weight: 1}) // inside an open header block
	fr.WriteContinuation(1, true, block[4:])
	fr.WriteData(1, true, []byte("body"))
	return wire.Bytes()
}

// The h2 tap parses client bytes alongside the real server; it must never
// panic, however the bytes are malformed or split across reads.
func FuzzH2Feed(f *testing.F) {
	seed := h2Seed()
	f.Add(seed, uint16(0))
	f.Add(seed, uint16(30))
	f.Add(seed[:40], uint16(7))
	f.Fuzz(func(t *testing.T, data []byte, split uint16) {
		c := NewH2Conn(nil)
		at := int(split) % (len(data) + 1)
		c.feed(data[:at])
		if !c.lost {
			c.feed(data[at:])
		}
		c.Fingerprint()
		c.Claim("GET", "acpwb.com", "/")
	})
}

func FuzzH1Scan(f *testing.F) {
	f.Add([]byte("GET / HTTP/1.1\r\nHost: x\r\nContent-Length: 3\r\n\r\nabcGET /b HTTP/1.1\r\nA: 1\r\n\r\n"), uint16(5))
	f.Add([]byte("POST / HTTP/1.1\r\nTransfer-Encoding: chunked\r\n\r\n0\r\n\r\n"), uint16(0))
	f.Add([]byte("\r\n\r\nGET / HTTP/1.0\r\nContent-Length: -5\r\n\r\n"), uint16(9))
	f.Fuzz(func(t *testing.T, data []byte, split uint16) {
		c := &H1Conn{}
		at := int(split) % (len(data) + 1)
		c.scan(data[:at])
		c.scan(data[at:])
		for c.NextHead() != nil {
		}
	})
}

// Regression: a frame that makes the tap give up mid-buffer (here a second
// HEADERS inside an open header block) used to slice the nil'd buffer and
// panic the server's read goroutine — h2spec 4.3.2 crashed the edge.
func TestH2GiveUpMidBufferDoesNotPanic(t *testing.T) {
	var wire bytes.Buffer
	wire.WriteString(http2.ClientPreface)
	fr := http2.NewFramer(&wire, nil)
	fr.WriteSettings()
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 1, BlockFragment: []byte{0x82}})
	fr.WriteHeaders(http2.HeadersFrameParam{StreamID: 3, BlockFragment: []byte{0x82}, EndHeaders: true})
	fr.WriteData(3, true, []byte("x"))
	c := NewH2Conn(nil)
	c.feed(wire.Bytes()) // must not panic
	if !c.lost {
		t.Fatal("expected the tap to give up")
	}
}

// The server→client side parser only sees the server's own frames, but it
// must still never panic, however writes are split.
func FuzzH2FeedOut(f *testing.F) {
	var out bytes.Buffer
	fr := http2.NewFramer(&out, nil)
	fr.WriteSettings(http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: 100})
	fr.WriteData(1, true, bytes.Repeat([]byte("y"), 300))
	fr.WriteGoAway(1, http2.ErrCodeEnhanceYourCalm, []byte("calm down"))
	f.Add(out.Bytes(), uint16(13))
	f.Fuzz(func(t *testing.T, data []byte, split uint16) {
		c := NewH2Conn(nil)
		at := int(split) % (len(data) + 1)
		c.feedOut(data[:at])
		c.feedOut(data[at:])
	})
}

func TestH2GoawayCapturedAcrossSplitWrites(t *testing.T) {
	var out bytes.Buffer
	fr := http2.NewFramer(&out, nil)
	fr.WriteSettings()
	fr.WriteData(1, true, bytes.Repeat([]byte("y"), 300))
	fr.WriteGoAway(1, http2.ErrCodeEnhanceYourCalm, []byte("calm down"))
	b := out.Bytes()
	for split := 1; split < len(b); split++ {
		c := NewH2Conn(nil)
		c.feedOut(b[:split])
		c.feedOut(b[split:])
		if c.stats.Goaway != "ENHANCE_YOUR_CALM" {
			t.Fatalf("split %d: goaway %q", split, c.stats.Goaway)
		}
	}
}
