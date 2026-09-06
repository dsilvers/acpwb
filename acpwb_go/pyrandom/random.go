package pyrandom

import (
	"math"
	"math/bits"
)

// randBelow ports Random._randbelow_with_getrandbits(n) from Lib/random.py.
// Requires n > 0.
func (r *Random) randBelow(n int64) int64 {
	k := bits.Len64(uint64(n))
	v := r.GetRandBits64(k)
	for v >= uint64(n) {
		v = r.GetRandBits64(k)
	}
	return int64(v)
}

// RandRange ports the fast (step==1) path of Random.randrange(start, stop).
// Panics if the range is empty, matching Python's ValueError behavior in
// spirit (callers in this codebase never hit that case).
func (r *Random) RandRange(start, stop int64) int64 {
	width := stop - start
	if width <= 0 {
		panic("pyrandom: empty range in RandRange")
	}
	return start + r.randBelow(width)
}

// RandInt ports Random.randint(a, b): inclusive of both endpoints.
func (r *Random) RandInt(a, b int64) int64 {
	return r.RandRange(a, b+1)
}

// noFuseMul multiplies two float64s across a real (non-inlined) function
// call boundary. This exists solely to defeat compiler fused-multiply-add
// contraction (permitted by the Go spec for floating-point expressions),
// which on some architectures (observed on arm64) can round
// a + (b-a)*x differently in the last bit than CPython's plain
// multiply-then-add. Splitting the multiply into an uninlinable call keeps
// Uniform's result bit-for-bit identical to CPython's.
//
//go:noinline
func noFuseMul(a, b float64) float64 {
	return a * b
}

// Uniform ports Random.uniform(a, b).
func (r *Random) Uniform(a, b float64) float64 {
	return a + noFuseMul(b-a, r.Random())
}

// Choice ports Random.choice(seq). Panics on an empty sequence, matching
// Python's IndexError in spirit.
func Choice[T any](r *Random, seq []T) T {
	if len(seq) == 0 {
		panic("pyrandom: cannot choose from an empty sequence")
	}
	return seq[r.randBelow(int64(len(seq)))]
}

// Shuffle ports Random.shuffle(x): in-place Fisher-Yates using the exact
// same traversal order and randBelow calls as CPython.
func Shuffle[T any](r *Random, x []T) {
	for i := len(x) - 1; i >= 1; i-- {
		j := r.randBelow(int64(i + 1))
		x[i], x[j] = x[j], x[i]
	}
}

// pow4CeilLog4 replicates Python's:
//
//	4 ** ceil(log(k * 3, 4))
//
// used by Random.sample() to size its "small set" threshold.
func pow4CeilLog4(k int) int64 {
	exp := int64(math.Ceil(math.Log(float64(k*3)) / math.Log(4)))
	result := int64(1)
	for i := int64(0); i < exp; i++ {
		result *= 4
	}
	return result
}

// Sample ports Random.sample(population, k) (without the counts=
// parameter, which the generators in this codebase never use). Returns a
// new slice; population is left unmodified.
func Sample[T any](r *Random, population []T, k int) []T {
	n := len(population)
	if k < 0 || k > n {
		panic("pyrandom: sample larger than population or is negative")
	}
	result := make([]T, k)
	setsize := int64(21)
	if k > 5 {
		setsize += pow4CeilLog4(k)
	}
	if int64(n) <= setsize {
		pool := make([]T, n)
		copy(pool, population)
		for i := 0; i < k; i++ {
			j := r.randBelow(int64(n - i))
			result[i] = pool[j]
			pool[j] = pool[n-i-1]
		}
	} else {
		selected := make(map[int64]bool, k)
		for i := 0; i < k; i++ {
			j := r.randBelow(int64(n))
			for selected[j] {
				j = r.randBelow(int64(n))
			}
			selected[j] = true
			result[i] = population[j]
		}
	}
	return result
}

// bisectRight ports bisect.bisect_right(a, x, lo, hi) restricted to the
// float64 slices used by Choices.
func bisectRight(a []float64, x float64, lo, hi int) int {
	for lo < hi {
		mid := (lo + hi) / 2
		if x < a[mid] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// Choices ports Random.choices(population, weights, cum_weights, k).
// Pass weights == nil and cumWeights == nil for equal-probability sampling
// with replacement. Passing both is invalid (mirrors Python's error there)
// and panics.
func Choices[T any](r *Random, population []T, weights []float64, cumWeights []float64, k int) []T {
	if weights != nil && cumWeights != nil {
		panic("pyrandom: cannot specify both weights and cumWeights")
	}
	n := len(population)
	result := make([]T, k)

	if cumWeights == nil {
		if weights == nil {
			nf := float64(n)
			for i := 0; i < k; i++ {
				idx := int(math.Floor(r.Random() * nf))
				result[i] = population[idx]
			}
			return result
		}
		cumWeights = make([]float64, len(weights))
		sum := 0.0
		for i, w := range weights {
			sum += w
			cumWeights[i] = sum
		}
	}

	if len(cumWeights) != n {
		panic("pyrandom: the number of weights does not match the population")
	}
	total := cumWeights[len(cumWeights)-1]
	if total <= 0.0 {
		panic("pyrandom: total of weights must be greater than zero")
	}
	hi := n - 1
	for i := 0; i < k; i++ {
		x := r.Random() * total
		idx := bisectRight(cumWeights, x, 0, hi)
		result[i] = population[idx]
	}
	return result
}
