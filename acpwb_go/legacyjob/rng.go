// Package legacyjob generates the "legacy source code" listings that sit
// behind archive pages' generated statistics tables: a deterministic,
// per-page batch job (JCL / DCL / Makefile / Jenkinsfile wrapper, compute
// program in FORTRAN / COBOL / BASIC / C / Perl, input deck, compiler
// listing, printed output) whose compute stage genuinely reproduces the
// numbers shown on the page.
//
// Nothing here has a Python counterpart except Identity (identity.go), which
// the article page renders and Django's pyrender mirrors for fixture parity.
// Everything else is free to use Go-native randomness.
package legacyjob

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand/v2"
	"strconv"
)

// rnd is a deterministic PCG stream seeded from an md5 of a seed string —
// same house pattern as the rest of the site (same seed → same output), just
// not Python-compatible since no Python code ever needs to reproduce it.
type rnd struct{ r *rand.Rand }

func newRnd(seed string) *rnd {
	h := md5.Sum([]byte(seed))
	return &rnd{rand.New(rand.NewPCG(binary.BigEndian.Uint64(h[:8]), binary.BigEndian.Uint64(h[8:])))}
}

// fork derives an independent child stream, so one subsystem drawing more
// or fewer values never shifts what another subsystem generates.
func (x *rnd) fork(label string) *rnd {
	return newRnd(label + "\x00" + strconv.FormatUint(x.r.Uint64(), 16))
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

func (x *rnd) weighted(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return 0
	}
	n := x.r.IntN(total)
	for i, w := range weights {
		if n < w {
			return i
		}
		n -= w
	}
	return len(weights) - 1
}

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
