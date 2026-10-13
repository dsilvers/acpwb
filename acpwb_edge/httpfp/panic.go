package httpfp

import (
	"log"
	"runtime/debug"
	"sync/atomic"
)

var panics atomic.Int64

// Panics counts recovered tap panics since startup.
func Panics() int64 { return panics.Load() }

// notePanic logs a recovered parser panic — with the stack, since it's a
// bug — at 1, 2, 4, 8… occurrences, so a client replaying the trigger
// can't flood the journal.
func notePanic(where string, r any) {
	n := panics.Add(1)
	if n&(n-1) == 0 {
		log.Printf("httpfp: recovered %s panic #%d (capture disabled for that connection): %v\n%s", where, n, r, debug.Stack())
	}
}
