package pyrandom

import (
	"crypto/sha512"
	"math/big"
)

// NewFromPythonStringSeed reproduces CPython's random.Random(a) construction
// when `a` is a str (the version-2 string-seeding path in Lib/random.py's
// Random.seed(), which is what a bare `random.Random(some_string)` call
// uses):
//
//	a = a.encode()             # utf-8
//	a = int.from_bytes(a + sha512(a).digest(), 'big')
//
// The ACPWB content generators frequently do exactly this:
//
//	random.Random(hashlib.md5(seed_str.encode()).hexdigest())
//
// i.e. seed with the 32-character hex-digest STRING itself (not an int), so
// this — not New/NewFromInt64 — is the correct constructor for that pattern.
func NewFromPythonStringSeed(s string) *Random {
	b := []byte(s)
	sum := sha512.Sum512(b)
	combined := make([]byte, 0, len(b)+len(sum))
	combined = append(combined, b...)
	combined = append(combined, sum[:]...)
	n := new(big.Int).SetBytes(combined)
	return New(n)
}
