package policy

import (
	"crypto/md5"
	"encoding/hex"
	"math/big"

	"acpwb_go/pyrandom"
)

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func md5Hex8(s string) string {
	return md5Hex(s)[:8]
}

// rngFromSeed reproduces policy_generator.py:_rng_from_seed —
// random.Random(int(hashlib.md5(seed_str.encode()).hexdigest(), 16)). This
// is the SAME seeding convention as archive package's rngB (verified there
// against the real Python source); duplicated here rather than imported
// because policy is deliberately a separate package (see task notes).
func rngFromSeed(seed string) *pyrandom.Random {
	h := md5Hex(seed)
	n := new(big.Int)
	n.SetString(h, 16)
	return pyrandom.New(n)
}

func choice[T any](r *pyrandom.Random, pool []T) T {
	return pyrandom.Choice(r, pool)
}

func sample[T any](r *pyrandom.Random, pool []T, k int) []T {
	return pyrandom.Sample(r, pool, k)
}

func shuffle[T any](r *pyrandom.Random, x []T) {
	pyrandom.Shuffle(r, x)
}

// choiceByte reproduces rng.choice(s) for a Python string s — a single
// character/byte chosen uniformly (all call sites here use plain ASCII
// strings, e.g. "ABCDEFGH").
func choiceByte(r *pyrandom.Random, s string) byte {
	return choice(r, []byte(s))
}
