package pyrandom

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
)

// step mirrors one recorded operation from testdata/fixtures.json, generated
// by a real CPython 3 interpreter (see scratchpad/gen_fixtures.py at the
// time this was written). Fields are loosely typed (json.RawMessage) because
// each op uses a different subset of them.
type step struct {
	Op         string          `json:"op"`
	A          json.Number     `json:"a"`
	B          json.Number     `json:"b"`
	K          int             `json:"k"`
	Pop        json.RawMessage `json:"pop"`
	Weights    []float64       `json:"weights"`
	CumWeights []float64       `json:"cum_weights"`
	Out        json.RawMessage `json:"out"`
}

type scenario struct {
	Name  string `json:"name"`
	Seed  string `json:"seed"`
	Steps []step `json:"steps"`
}

func loadScenarios(t *testing.T) []scenario {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures.json")
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}
	var scenarios []scenario
	if err := json.Unmarshal(data, &scenarios); err != nil {
		t.Fatalf("parsing fixtures: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatalf("no scenarios loaded")
	}
	return scenarios
}

func mustFloat(t *testing.T, n json.Number) float64 {
	t.Helper()
	f, err := n.Float64()
	if err != nil {
		t.Fatalf("bad float %q: %v", n, err)
	}
	return f
}

func mustInt64(t *testing.T, n json.Number) int64 {
	t.Helper()
	i, err := n.Int64()
	if err != nil {
		t.Fatalf("bad int %q: %v", n, err)
	}
	return i
}

func decodeStrSlice(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	var s []string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("decoding []string: %v", err)
	}
	return s
}

func decodeIntSlice(t *testing.T, raw json.RawMessage) []int64 {
	t.Helper()
	var s []int64
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("decoding []int64: %v", err)
	}
	return s
}

func TestAgainstCPythonFixtures(t *testing.T) {
	scenarios := loadScenarios(t)

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.Name, func(t *testing.T) {
			seed, ok := new(big.Int).SetString(sc.Seed, 10)
			if !ok {
				t.Fatalf("bad seed %q", sc.Seed)
			}
			r := New(seed)

			for i, st := range sc.Steps {
				ctx := func() string { return sc.Name + "/" + st.Op + "#" + itoa(i) }

				switch st.Op {
				case "random", "random2":
					got := r.Random()
					var want float64
					mustUnmarshal(t, st.Out, &want)
					if got != want {
						t.Fatalf("%s: random() = %v, want %v", ctx(), got, want)
					}

				case "randint", "randint2":
					a := mustInt64(t, st.A)
					b := mustInt64(t, st.B)
					got := r.RandInt(a, b)
					want := mustInt64(t, jsonNumber(t, st.Out))
					if got != want {
						t.Fatalf("%s: RandInt(%d,%d) = %d, want %d", ctx(), a, b, got, want)
					}

				case "uniform", "uniform2":
					a := mustFloat(t, st.A)
					b := mustFloat(t, st.B)
					got := r.Uniform(a, b)
					var want float64
					mustUnmarshal(t, st.Out, &want)
					if got != want {
						t.Fatalf("%s: Uniform(%v,%v) = %v, want %v", ctx(), a, b, got, want)
					}

				case "choice_str", "choice_str2":
					pop := decodeStrSlice(t, st.Pop)
					got := Choice(r, pop)
					var want string
					mustUnmarshal(t, st.Out, &want)
					if got != want {
						t.Fatalf("%s: Choice = %q, want %q", ctx(), got, want)
					}

				case "choice_int":
					pop := decodeIntSlice(t, st.Pop)
					got := Choice(r, pop)
					var want int64
					mustUnmarshal(t, st.Out, &want)
					if got != want {
						t.Fatalf("%s: Choice = %d, want %d", ctx(), got, want)
					}

				case "getrandbits":
					got := r.GetRandBits(st.K)
					var wantStr json.Number
					mustUnmarshal(t, st.Out, &wantStr)
					want, ok := new(big.Int).SetString(wantStr.String(), 10)
					if !ok {
						t.Fatalf("%s: bad want bigint %q", ctx(), wantStr)
					}
					if got.Cmp(want) != 0 {
						t.Fatalf("%s: GetRandBits(%d) = %v, want %v", ctx(), st.K, got, want)
					}

				case "shuffle_int":
					pop := decodeIntSlice(t, st.Pop)
					cp := append([]int64(nil), pop...)
					Shuffle(r, cp)
					want := decodeIntSlice(t, st.Out)
					assertInt64SliceEqual(t, ctx(), cp, want)

				case "shuffle_str":
					pop := decodeStrSlice(t, st.Pop)
					cp := append([]string(nil), pop...)
					Shuffle(r, cp)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), cp, want)

				case "sample_int":
					pop := decodeIntSlice(t, st.Pop)
					got := Sample(r, pop, st.K)
					want := decodeIntSlice(t, st.Out)
					assertInt64SliceEqual(t, ctx(), got, want)

				case "sample_str":
					pop := decodeStrSlice(t, st.Pop)
					got := Sample(r, pop, st.K)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), got, want)

				case "sample_big":
					pop := decodeIntSlice(t, st.Pop)
					got := Sample(r, pop, st.K)
					want := decodeIntSlice(t, st.Out)
					assertInt64SliceEqual(t, ctx(), got, want)

				case "sample_full":
					pop := decodeStrSlice(t, st.Pop)
					got := Sample(r, pop, st.K)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), got, want)

				case "choices_equal":
					pop := decodeStrSlice(t, st.Pop)
					got := Choices(r, pop, nil, nil, st.K)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), got, want)

				case "choices_weighted":
					pop := decodeStrSlice(t, st.Pop)
					weights := make([]float64, len(st.Weights))
					copy(weights, st.Weights)
					got := Choices(r, pop, weights, nil, st.K)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), got, want)

				case "choices_cumweighted":
					pop := decodeStrSlice(t, st.Pop)
					cw := make([]float64, len(st.CumWeights))
					copy(cw, st.CumWeights)
					got := Choices(r, pop, nil, cw, st.K)
					want := decodeStrSlice(t, st.Out)
					assertStrSliceEqual(t, ctx(), got, want)

				default:
					t.Fatalf("%s: unknown op %q", ctx(), st.Op)
				}
			}
		})
	}
}

func mustUnmarshal(t *testing.T, raw json.RawMessage, v any) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("unmarshal %s into %T: %v", raw, v, err)
	}
}

func jsonNumber(t *testing.T, raw json.RawMessage) json.Number {
	t.Helper()
	var n json.Number
	mustUnmarshal(t, raw, &n)
	return n
}

func assertInt64SliceEqual(t *testing.T, ctx string, got, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: length mismatch got=%v want=%v", ctx, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: index %d got=%v want=%v (full got=%v want=%v)", ctx, i, got[i], want[i], got, want)
		}
	}
}

func assertStrSliceEqual(t *testing.T, ctx string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: length mismatch got=%v want=%v", ctx, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: index %d got=%v want=%v (full got=%v want=%v)", ctx, i, got[i], want[i], got, want)
		}
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
