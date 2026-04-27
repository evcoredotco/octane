// Package rand_test contains black-box tests for the rand package (T-002-24).
package rand_test

import (
	"testing"

	"github.com/octane-project/octane/pkg/engine/rand"
)

// goldenSeed is the fixed seed used for cross-platform golden comparisons.
const goldenSeed uint64 = 0xDEADBEEF

// goldenInt63 is the expected sequence of Int63 values from Deterministic
// seeded with goldenSeed. If this sequence ever changes, the determinism
// guarantee (AC6) has been broken.
var goldenInt63 = []int64{
	597728187450255743, 5905879438849105626, 6200458422628470231,
	2156080055563053635, 4286555703373052804, 5935437820250642311,
	7785570495430472211, 8666312242828484705, 2997161554461854968,
	819933583207398863, 1145447176862202584, 7901221965921019868,
	5735419280001882410, 4177816216003567600, 2526656054580991210,
	1041471264482178515, 3980076034963716207, 50058807992993355,
	546392340153015798, 6361685935628303573,
}

// goldenFloat64 is the expected sequence of Float64 values drawn from the
// same RNG after the Int63 sequence above has been consumed.
var goldenFloat64 = []float64{
	0.7922225762829015, 0.15961138887565152, 0.271398117795402,
	0.8832252330479566, 0.8992383559120135, 0.13199738542030615,
	0.2597513076824052, 0.371252640695678, 0.8601252495371617,
	0.7564041323089473, 0.8052411702769897, 0.07068325071343995,
	0.583139569620237, 0.5545883153616629, 0.49428323722299106,
	0.4960427280051176, 0.6815548018092786, 0.9964235903383537,
	0.019301430454382484, 0.5164767780532273,
}

// goldenIntn is the expected sequence of Intn(100) values drawn from the
// same RNG after the Int63 and Float64 sequences above have been consumed.
var goldenIntn = []int{
	16, 82, 16, 77, 99, 61, 92, 25, 4, 11,
}

// TestGoldenSequence verifies that Deterministic(goldenSeed) produces the
// exact pre-computed golden sequence. A change to this sequence means the
// determinism guarantee (AC6) is broken and must be investigated before
// updating the golden values.
func TestGoldenSequence(t *testing.T) {
	t.Parallel()

	rng := rand.Deterministic(goldenSeed)

	for idx, want := range goldenInt63 {
		got := rng.Int63()
		if got != want {
			t.Errorf("Int63[%d]: got %d, want %d", idx, got, want)
		}
	}

	for idx, want := range goldenFloat64 {
		got := rng.Float64()
		if got != want {
			t.Errorf("Float64[%d]: got %g, want %g", idx, got, want)
		}
	}

	for idx, want := range goldenIntn {
		got := rng.Intn(100)
		if got != want {
			t.Errorf("Intn(100)[%d]: got %d, want %d", idx, got, want)
		}
	}
}

// TestDeterministicSameSeedIdentical verifies that two Rand instances
// created with the same seed produce byte-identical sequences.
func TestDeterministicSameSeedIdentical(t *testing.T) {
	t.Parallel()

	const iterations = 50

	rng1 := rand.Deterministic(goldenSeed)
	rng2 := rand.Deterministic(goldenSeed)

	for idx := 0; idx < iterations; idx++ {
		got1 := rng1.Int63()
		got2 := rng2.Int63()

		if got1 != got2 {
			t.Errorf("Int63[%d]: rng1=%d rng2=%d diverge", idx, got1, got2)
		}
	}

	for idx := 0; idx < iterations; idx++ {
		got1 := rng1.Float64()
		got2 := rng2.Float64()

		if got1 != got2 {
			t.Errorf("Float64[%d]: rng1=%g rng2=%g diverge", idx, got1, got2)
		}
	}

	for idx := 0; idx < iterations; idx++ {
		got1 := rng1.Intn(100)
		got2 := rng2.Intn(100)

		if got1 != got2 {
			t.Errorf("Intn(100)[%d]: rng1=%d rng2=%d diverge", idx, got1, got2)
		}
	}
}

// TestDeterministicDifferentSeedsDiffer verifies that two different seeds
// produce different sequences, guarding against degenerate implementations.
func TestDeterministicDifferentSeedsDiffer(t *testing.T) {
	t.Parallel()

	rng1 := rand.Deterministic(goldenSeed)
	rng2 := rand.Deterministic(goldenSeed + 1)

	differ := false

	for idx := 0; idx < 20; idx++ {
		if rng1.Int63() != rng2.Int63() {
			differ = true

			break
		}
	}

	if !differ {
		t.Error("expected different seeds to produce different sequences")
	}
}

// TestRealRandInterface verifies that Real() satisfies the Rand interface
// and produces values in valid ranges.
func TestRealRandInterface(t *testing.T) {
	t.Parallel()

	rng := rand.Real()

	for idx := 0; idx < 20; idx++ {
		val := rng.Int63()
		if val < 0 {
			t.Errorf("Int63[%d]=%d is negative", idx, val)
		}
	}

	for idx := 0; idx < 20; idx++ {
		val := rng.Float64()
		if val < 0.0 || val >= 1.0 {
			t.Errorf("Float64[%d]=%g out of [0.0, 1.0)", idx, val)
		}
	}

	for idx := 0; idx < 20; idx++ {
		val := rng.Intn(100)
		if val < 0 || val >= 100 {
			t.Errorf("Intn(100)[%d]=%d out of [0, 100)", idx, val)
		}
	}
}

// TestDeterministicPrintGolden is a helper that prints the actual sequence
// when run with -v. It is not a correctness assertion; use it to regenerate
// golden values when the PCG implementation changes under a new Go major
// version.
func TestDeterministicPrintGolden(t *testing.T) {
	t.Parallel()

	rng := rand.Deterministic(goldenSeed)

	t.Log("Int63 sequence:")

	for idx := 0; idx < 20; idx++ {
		t.Logf("  [%02d] %d", idx, rng.Int63())
	}

	t.Log("Float64 sequence:")

	for idx := 0; idx < 20; idx++ {
		t.Logf("  [%02d] %v", idx, rng.Float64())
	}

	t.Log("Intn(100) sequence:")

	for idx := 0; idx < 10; idx++ {
		t.Logf("  [%02d] %d", idx, rng.Intn(100))
	}
}
