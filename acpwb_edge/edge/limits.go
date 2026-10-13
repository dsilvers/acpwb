package edge

import (
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// Limits are deliberately not per client: the site exists to give bots
// unthrottled access, so nothing here caps a client's connections or
// request rate. They only protect the process from resources a client
// holds without using them.
type Limits struct {
	// No body/response progress for this long ends the request (nginx's
	// client_body_timeout / send_timeout). Clients that stop reading their
	// responses otherwise park a handler each, up to 100 per HTTP/2
	// connection (~40 KB apiece) — 300 such connections reached 74k
	// goroutines and ~1.9 GB in testing. Default 15s (the slowest
	// download seen in production took 19.7s in all, still making
	// progress).
	StallTimeout time.Duration
	// Last-resort overload valve: past this many in-flight requests across
	// all clients, new requests get an immediate 503 instead of piling
	// onto a process that's about to run out of memory. Production carries
	// ~1-2k; default 100000 (~4 GB worst case). Negative disables it.
	RequestsInFlight int
	// Individual records per IP per minute for connections that carried no
	// request, before the rest are folded into aggregated rows (exact
	// conn_count, nothing dropped). Connections with requests are always
	// recorded individually. Default 600; negative records every
	// connection individually.
	RecordsPerIPMin int
}

func (l *Limits) defaults() {
	if l.StallTimeout == 0 {
		l.StallTimeout = 15 * time.Second
	}
	if l.RequestsInFlight == 0 {
		l.RequestsInFlight = 100000
	}
	if l.RecordsPerIPMin == 0 {
		l.RecordsPerIPMin = 600
	}
}

var (
	rejectedRequests atomic.Int64
	inFlight         atomic.Int64
)

// admitRequest takes an in-flight slot, or refuses if the process-wide
// overload valve is shut.
func (s *Server) admitRequest() (release func(), ok bool) {
	n := inFlight.Add(1)
	if max := s.cfg.Limits.RequestsInFlight; max > 0 && n > int64(max) {
		inFlight.Add(-1)
		return nil, false
	}
	return func() { inFlight.Add(-1) }, true
}

// stallReader extends the read deadline whenever the body makes progress,
// so a request ends only when its body stalls (nginx client_body_timeout),
// not because a large upload is slow.
type stallReader struct {
	io.ReadCloser
	rc      *http.ResponseController
	timeout time.Duration
	last    time.Time
}

func (b *stallReader) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		if now := time.Now(); now.Sub(b.last) > time.Second {
			b.last = now
			_ = b.rc.SetReadDeadline(now.Add(b.timeout))
		}
	}
	return n, err
}
