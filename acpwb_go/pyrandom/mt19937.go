// Package pyrandom is a byte-for-byte-faithful reimplementation of CPython's
// random.Random (the Mersenne Twister based PRNG), restricted to the subset
// of the API actually used by the ACPWB deterministic content generators:
//
//	seed(int), random(), randint(a, b), choice(seq), sample(population, k),
//	uniform(a, b), shuffle(x), getrandbits(k),
//	choices(population, weights, cum_weights, k)
//
// The core MT19937 generator (seeding via init_by_array, genrand_uint32,
// random(), getrandbits(k)) mirrors CPython's Modules/_randommodule.c.
// The higher-level methods (randrange/randint, choice, shuffle, sample,
// choices, uniform, and the internal _randbelow helper) are transcribed
// directly from CPython's Lib/random.py (Random.__init_subclass__ always
// wires _randbelow to _randbelow_with_getrandbits because this type defines
// getrandbits, which is exactly what happens here).
//
// Only integer seeds are supported (that's all the generators use); the
// str/bytes/None seeding paths from Lib/random.py's Random.seed() are not
// implemented.
package pyrandom

import "math/big"

const (
	n         = 624
	m         = 397
	matrixA   = 0x9908b0df
	upperMask = 0x80000000
	lowerMask = 0x7fffffff
)

// Random is a Mersenne-Twister PRNG matching CPython's random.Random for
// integer seeds.
type Random struct {
	mt  [n]uint32
	mti int // index into mt; n+1 means "uninitialized"
}

// New creates a Random seeded the same way random.Random(seed) would be in
// Python, where seed is a non-negative (or negative -- CPython takes abs())
// integer.
func New(seed *big.Int) *Random {
	r := &Random{}
	r.Seed(seed)
	return r
}

// NewFromInt64 is a convenience wrapper around New for small seeds.
func NewFromInt64(seed int64) *Random {
	return New(big.NewInt(seed))
}

// Seed re-seeds the generator exactly as CPython's _randommodule.c
// random_seed() does for an integer argument:
//
//   - take abs(a)
//   - decompose into 32-bit little-endian "digits" (base 2**32); zero
//     becomes the single-word key [0]
//   - run init_by_array(key, len(key))
func (r *Random) Seed(seed *big.Int) {
	a := new(big.Int).Abs(seed)

	var key []uint32
	if a.Sign() == 0 {
		key = []uint32{0}
	} else {
		mask := big.NewInt(0xffffffff)
		tmp := new(big.Int).Set(a)
		for tmp.Sign() != 0 {
			word := new(big.Int).And(tmp, mask)
			key = append(key, uint32(word.Uint64()))
			tmp.Rsh(tmp, 32)
		}
	}

	r.initByArray(key)
}

func (r *Random) initGenrand(s uint32) {
	r.mt[0] = s
	for i := 1; i < n; i++ {
		prev := r.mt[i-1]
		r.mt[i] = (1812433253*(prev^(prev>>30)) + uint32(i))
	}
	r.mti = n
}

func (r *Random) initByArray(key []uint32) {
	r.initGenrand(19650218)
	i, j := 1, 0
	k := n
	if len(key) > k {
		k = len(key)
	}
	for ; k != 0; k-- {
		prev := r.mt[i-1]
		r.mt[i] = (r.mt[i] ^ ((prev ^ (prev >> 30)) * 1664525)) + key[j] + uint32(j)
		i++
		j++
		if i >= n {
			r.mt[0] = r.mt[n-1]
			i = 1
		}
		if j >= len(key) {
			j = 0
		}
	}
	for k = n - 1; k != 0; k-- {
		prev := r.mt[i-1]
		r.mt[i] = (r.mt[i] ^ ((prev ^ (prev >> 30)) * 1566083941)) - uint32(i)
		i++
		if i >= n {
			r.mt[0] = r.mt[n-1]
			i = 1
		}
	}
	r.mt[0] = 0x80000000
}

var mag01 = [2]uint32{0x0, matrixA}

// genrandUint32 returns the next raw 32-bit tempered MT19937 output.
func (r *Random) genrandUint32() uint32 {
	if r.mti >= n {
		var y uint32
		var kk int
		for kk = 0; kk < n-m; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+m] ^ (y >> 1) ^ mag01[y&0x1]
		}
		for ; kk < n-1; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+(m-n)] ^ (y >> 1) ^ mag01[y&0x1]
		}
		y = (r.mt[n-1] & upperMask) | (r.mt[0] & lowerMask)
		r.mt[n-1] = r.mt[m-1] ^ (y >> 1) ^ mag01[y&0x1]
		r.mti = 0
	}

	y := r.mt[r.mti]
	r.mti++

	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18

	return y
}

// Random returns the next float64 in [0.0, 1.0), matching CPython's
// random_random(): 53 bits of precision built from two 32-bit draws.
func (r *Random) Random() float64 {
	a := r.genrandUint32() >> 5 // top 27 bits
	b := r.genrandUint32() >> 6 // top 26 bits
	return (float64(a)*67108864.0 + float64(b)) * (1.0 / 9007199254740992.0)
}

// GetRandBits returns an arbitrary-precision unsigned integer with exactly k
// random bits, matching CPython's random_getrandbits() chunking: 32-bit
// chunks are filled from least-significant to most-significant, with the
// final (most significant) chunk right-shifted to keep only the remaining
// bits.
func (r *Random) GetRandBits(k int) *big.Int {
	if k <= 0 {
		return big.NewInt(0)
	}
	if k <= 32 {
		v := r.genrandUint32() >> (32 - uint(k))
		return new(big.Int).SetUint64(uint64(v))
	}

	result := new(big.Int)
	shift := uint(0)
	remaining := k
	for remaining > 0 {
		var word uint32
		if remaining >= 32 {
			word = r.genrandUint32()
		} else {
			word = r.genrandUint32() >> (32 - uint(remaining))
		}
		wordBig := new(big.Int).Lsh(new(big.Int).SetUint64(uint64(word)), shift)
		result.Or(result, wordBig)
		shift += 32
		remaining -= 32
	}
	return result
}

// GetRandBits64 is a convenience wrapper for the common case k <= 64, used
// internally for _randbelow etc. since population sizes used by the
// generators never approach 2**64.
func (r *Random) GetRandBits64(k int) uint64 {
	if k <= 32 {
		return uint64(r.genrandUint32() >> (32 - uint(k)))
	}
	lo := uint64(r.genrandUint32())
	hi := r.genrandUint32() >> (64 - uint(k))
	return lo | (uint64(hi) << 32)
}
