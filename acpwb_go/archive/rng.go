package archive

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

// rngA reproduces `random.Random(hashlib.md5(seed.encode()).hexdigest())` —
// seeding with the hex-digest STRING (Python's version-2 str seeding path).
// This is the pattern views.py uses directly for archive content/nav/related
// generators (_generate_archive_content, _gen_nav_slugs, _gen_related_path_data,
// _gen_cross_year_reports, _gen_related_docs_data, _gen_presentations_count).
func rngA(seed string) *pyrandom.Random {
	return pyrandom.NewFromPythonStringSeed(md5Hex(seed))
}

// rngB reproduces `random.Random(int(hashlib.md5(seed.encode()).hexdigest(), 16))`
// — the FULL hex digest as an integer, no masking. Used by
// policy_generator._rng_from_seed and report_generator._rng_from_seed
// (despite the latter's misleading docstring claiming it matches the mod-2**32
// variant below — it does not; verified against the source).
func rngB(seed string) *pyrandom.Random {
	h := md5Hex(seed)
	n := new(big.Int)
	n.SetString(h, 16)
	return pyrandom.New(n)
}

// rngC reproduces `random.Random(int(hashlib.md5(seed.encode()).hexdigest(), 16) % (2**32))`
// — used by presentations/generators.py's and logo_generator.py's
// _rng_from_seed.
func rngC(seed string) *pyrandom.Random {
	h := md5Hex(seed)
	n := new(big.Int)
	n.SetString(h, 16)
	mod := new(big.Int).Lsh(big.NewInt(1), 32)
	n.Mod(n, mod)
	return pyrandom.New(n)
}

func choice[T any](r *pyrandom.Random, pool []T) T {
	return pyrandom.Choice(r, pool)
}

func sample[T any](r *pyrandom.Random, pool []T, k int) []T {
	return pyrandom.Sample(r, pool, k)
}
