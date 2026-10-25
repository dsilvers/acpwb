package compliance

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand/v2"
)

// rnd is a deterministic PCG stream seeded from an md5 of a seed string, the
// same house pattern as legacyjob/rng.go (same seed → same output). Nothing
// in this package has a Python counterpart, so Go-native randomness is fine.
type rnd struct{ r *rand.Rand }

func newRnd(seed string) *rnd {
	h := md5.Sum([]byte(seed))
	return &rnd{rand.New(rand.NewPCG(binary.BigEndian.Uint64(h[:8]), binary.BigEndian.Uint64(h[8:])))}
}

func (x *rnd) intn(n int) int {
	if n <= 1 {
		return 0
	}
	return x.r.IntN(n)
}

// between returns an int in [lo, hi] inclusive.
func (x *rnd) between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + x.r.IntN(hi-lo+1)
}

func (x *rnd) chance(p float64) bool { return x.r.Float64() < p }

func (x *rnd) float() float64 { return x.r.Float64() }

func pick[T any](x *rnd, s []T) T {
	var zero T
	if len(s) == 0 {
		return zero
	}
	return s[x.r.IntN(len(s))]
}

func shuffle[T any](x *rnd, s []T) {
	x.r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// sample returns k distinct elements (or all of them, shuffled, if k >= len).
func sample[T any](x *rnd, s []T, k int) []T {
	cp := append([]T(nil), s...)
	shuffle(x, cp)
	if k < len(cp) {
		cp = cp[:k]
	}
	return cp
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
