// Package rand_test contains black-box tests for the rand package (T-002-24).
package rand_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/engine/rand"
)

// goldenSeed is the fixed seed used for cross-platform golden comparisons.
const goldenSeed uint64 = 0xDEADBEEF

// intnBound is the upper bound for Intn calls in these tests.
const intnBound = 100

// iterRange is the number of iterations for range-check loops.
const iterRange = 20

// goldenIntnCount is the number of Intn values in the golden sequence.
const goldenIntnCount = 10

// goldenInt63Val0 through goldenInt63Val6 are the first seven Int63 golden
// values used as spot-check anchors in the printed sequence.
const (
	goldenInt63Val0 = int64(597728187450255743)
	goldenInt63Val1 = int64(2156080055563053635)
	goldenInt63Val2 = int64(7785570495430472211)
	goldenInt63Val3 = int64(819933583207398863)
	goldenInt63Val4 = int64(5735419280001882410)
	goldenInt63Val5 = int64(1041471264482178515)
	goldenInt63Val6 = int64(546392340153015798)
)

// goldenFloat64Val0 through goldenFloat64Val6 are the first seven Float64
// golden values used as spot-check anchors in the printed sequence.
const (
	goldenFloat64Val0 = 0.7922225762829015
	goldenFloat64Val1 = 0.8832252330479566
	goldenFloat64Val2 = 0.2597513076824052
	goldenFloat64Val3 = 0.7564041323089473
	goldenFloat64Val4 = 0.583139569620237
	goldenFloat64Val5 = 0.4960427280051176
	goldenFloat64Val6 = 0.019301430454382484
)

// goldenIntnVal16 is the repeated Intn golden value that appears at indices
// 0 and 2 of the golden Intn sequence.
const goldenIntnVal16 = 16

// floatZero is the float64 lower bound for Float64 range checks.
const floatZero = 0.0

// intZero is the lower bound for Int63 non-negative assertions.
const intZero = 0

// maxFloat64Exclusive is the exclusive upper bound for Float64 range checks.
const maxFloat64Exclusive = 1.0

// goldenInt63SeqA0 through goldenInt63SeqA6 are the even-indexed interleaved
// Int63 golden values in goldenInt63Seq (the values that follow the named
// goldenInt63ValN anchors in each triplet).
const (
	goldenInt63SeqA0 = int64(5905879438849105626)
	goldenInt63SeqA1 = int64(4286555703373052804)
	goldenInt63SeqA2 = int64(8666312242828484705)
	goldenInt63SeqA3 = int64(1145447176862202584)
	goldenInt63SeqA4 = int64(4177816216003567600)
	goldenInt63SeqA5 = int64(3980076034963716207)
	goldenInt63SeqA6 = int64(6361685935628303573)
)

// goldenInt63SeqB0 through goldenInt63SeqB5 are the odd-indexed interleaved
// Int63 golden values in goldenInt63Seq (the third value in each triplet).
const (
	goldenInt63SeqB0 = int64(6200458422628470231)
	goldenInt63SeqB1 = int64(5935437820250642311)
	goldenInt63SeqB2 = int64(2997161554461854968)
	goldenInt63SeqB3 = int64(7901221965921019868)
	goldenInt63SeqB4 = int64(2526656054580991210)
	goldenInt63SeqB5 = int64(50058807992993355)
)

// goldenFloat64SeqA0 through goldenFloat64SeqA6 are the even-indexed
// interleaved Float64 golden values in goldenFloat64Seq (the values that
// follow the named goldenFloat64ValN anchors in each triplet).
const (
	goldenFloat64SeqA0 = 0.15961138887565152
	goldenFloat64SeqA1 = 0.8992383559120135
	goldenFloat64SeqA2 = 0.371252640695678
	goldenFloat64SeqA3 = 0.8052411702769897
	goldenFloat64SeqA4 = 0.5545883153616629
	goldenFloat64SeqA5 = 0.6815548018092786
	goldenFloat64SeqA6 = 0.5164767780532273
)

// goldenFloat64SeqB0 through goldenFloat64SeqB5 are the odd-indexed
// interleaved Float64 golden values in goldenFloat64Seq (the third value in
// each triplet).
const (
	goldenFloat64SeqB0 = 0.271398117795402
	goldenFloat64SeqB1 = 0.13199738542030615
	goldenFloat64SeqB2 = 0.8601252495371617
	goldenFloat64SeqB3 = 0.07068325071343995
	goldenFloat64SeqB4 = 0.49428323722299106
	goldenFloat64SeqB5 = 0.9964235903383537
)

// goldenIntnCount2 is the second Intn golden value in goldenIntnSeq (the
// element at index 1, between the two goldenIntnVal16 appearances).
const goldenIntnCount2 = 82

// goldenIntnSeqTail contains the Intn golden values at indices 3–9 of
// goldenIntnSeq (the tail after goldenIntnVal16, goldenIntnCount2,
// goldenIntnVal16).
const (
	goldenIntnTail0 = 77
	goldenIntnTail1 = 99
	goldenIntnTail2 = 61
	goldenIntnTail3 = 92
	goldenIntnTail4 = 25
	goldenIntnTail5 = 4
	goldenIntnTail6 = 11
)

// goldenInt63Seq returns the expected sequence of Int63 values from
// Deterministic seeded with goldenSeed. If this sequence ever changes,
// the determinism guarantee (AC6) has been broken.
func goldenInt63Seq() []int64 {
	return []int64{
		goldenInt63Val0, goldenInt63SeqA0, goldenInt63SeqB0,
		goldenInt63Val1, goldenInt63SeqA1, goldenInt63SeqB1,
		goldenInt63Val2, goldenInt63SeqA2, goldenInt63SeqB2,
		goldenInt63Val3, goldenInt63SeqA3, goldenInt63SeqB3,
		goldenInt63Val4, goldenInt63SeqA4, goldenInt63SeqB4,
		goldenInt63Val5, goldenInt63SeqA5, goldenInt63SeqB5,
		goldenInt63Val6, goldenInt63SeqA6,
	}
}

// goldenFloat64Seq returns the expected sequence of Float64 values drawn
// from the same RNG after the Int63 sequence above has been consumed.
func goldenFloat64Seq() []float64 {
	return []float64{
		goldenFloat64Val0, goldenFloat64SeqA0, goldenFloat64SeqB0,
		goldenFloat64Val1, goldenFloat64SeqA1, goldenFloat64SeqB1,
		goldenFloat64Val2, goldenFloat64SeqA2, goldenFloat64SeqB2,
		goldenFloat64Val3, goldenFloat64SeqA3, goldenFloat64SeqB3,
		goldenFloat64Val4, goldenFloat64SeqA4, goldenFloat64SeqB4,
		goldenFloat64Val5, goldenFloat64SeqA5, goldenFloat64SeqB5,
		goldenFloat64Val6, goldenFloat64SeqA6,
	}
}

// goldenIntnSeq returns the expected sequence of Intn(intnBound) values
// drawn from the same RNG after the Int63 and Float64 sequences above have
// been consumed.
func goldenIntnSeq() []int {
	return []int{
		goldenIntnVal16, goldenIntnCount2, goldenIntnVal16,
		goldenIntnTail0, goldenIntnTail1, goldenIntnTail2,
		goldenIntnTail3, goldenIntnTail4, goldenIntnTail5,
		goldenIntnTail6,
	}
}

// checkInt63Seq drains rng.Int63() for each element in wantSeq and
// reports an error for any mismatch.
func checkInt63Seq(
	t *testing.T,
	rng interface{ Int63() int64 },
	wantSeq []int64,
) {
	t.Helper()

	for idx, wantVal := range wantSeq {
		if got := rng.Int63(); got != wantVal {
			t.Errorf("Int63[%d]: got %d, want %d", idx, got, wantVal)
		}
	}
}

// checkFloat64Seq drains rng.Float64() for each element in wantSeq and
// reports an error for any mismatch.
func checkFloat64Seq(
	t *testing.T,
	rng interface{ Float64() float64 },
	wantSeq []float64,
) {
	t.Helper()

	for idx, wantVal := range wantSeq {
		if got := rng.Float64(); got != wantVal {
			t.Errorf("Float64[%d]: got %g, want %g", idx, got, wantVal)
		}
	}
}

// checkIntnSeq drains rng.Intn(bound) for each element in wantSeq and
// reports an error for any mismatch.
func checkIntnSeq(
	t *testing.T,
	rng interface{ Intn(n int) int },
	wantSeq []int,
	bound int,
) {
	t.Helper()

	for idx, wantVal := range wantSeq {
		if got := rng.Intn(bound); got != wantVal {
			t.Errorf("Intn(%d)[%d]: got %d, want %d", bound, idx, got, wantVal)
		}
	}
}

// TestGoldenSequence verifies that Deterministic(goldenSeed) produces the
// exact pre-computed golden sequence. A change to this sequence means the
// determinism guarantee (AC6) is broken and must be investigated before
// updating the golden values.
func TestGoldenSequence(t *testing.T) {
	t.Parallel()

	rng := rand.Deterministic(goldenSeed)

	checkInt63Seq(t, rng, goldenInt63Seq())
	checkFloat64Seq(t, rng, goldenFloat64Seq())
	checkIntnSeq(t, rng, goldenIntnSeq(), intnBound)
}

// assertInt63Identical asserts that count consecutive Int63 calls on rng1
// and rng2 yield identical values.
func assertInt63Identical(
	t *testing.T,
	rng1, rng2 interface{ Int63() int64 },
	count int,
) {
	t.Helper()

	for idx := range count {
		if got1, got2 := rng1.Int63(), rng2.Int63(); got1 != got2 {
			t.Errorf("Int63[%d]: rng1=%d rng2=%d diverge", idx, got1, got2)
		}
	}
}

// assertFloat64Identical asserts that count consecutive Float64 calls on rng1
// and rng2 yield identical values.
func assertFloat64Identical(
	t *testing.T,
	rng1, rng2 interface{ Float64() float64 },
	count int,
) {
	t.Helper()

	for idx := range count {
		if got1, got2 := rng1.Float64(), rng2.Float64(); got1 != got2 {
			t.Errorf("Float64[%d]: rng1=%g rng2=%g diverge", idx, got1, got2)
		}
	}
}

// assertIntnIdentical asserts that count consecutive Intn(bound) calls on
// rng1 and rng2 yield identical values.
func assertIntnIdentical(
	t *testing.T,
	rng1, rng2 interface{ Intn(n int) int },
	count, bound int,
) {
	t.Helper()

	for idx := range count {
		got1, got2 := rng1.Intn(bound), rng2.Intn(bound)
		if got1 != got2 {
			t.Errorf(
				"Intn(%d)[%d]: rng1=%d rng2=%d diverge",
				bound, idx, got1, got2,
			)
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

	assertInt63Identical(t, rng1, rng2, iterations)
	assertFloat64Identical(t, rng1, rng2, iterations)
	assertIntnIdentical(t, rng1, rng2, iterations, intnBound)
}

// TestDeterministicDifferentSeedsDiffer verifies that two different seeds
// produce different sequences, guarding against degenerate implementations.
func TestDeterministicDifferentSeedsDiffer(t *testing.T) {
	t.Parallel()

	rng1 := rand.Deterministic(goldenSeed)
	rng2 := rand.Deterministic(goldenSeed + 1)

	differ := false

	for range iterRange {
		if rng1.Int63() != rng2.Int63() {
			differ = true

			break
		}
	}

	if !differ {
		t.Error("expected different seeds to produce different sequences")
	}
}

// assertInt63NonNegative asserts that count consecutive Int63 calls produce
// non-negative values.
func assertInt63NonNegative(
	t *testing.T,
	rng interface{ Int63() int64 },
	count int,
) {
	t.Helper()

	for idx := range count {
		if val := rng.Int63(); val < intZero {
			t.Errorf("Int63[%d]=%d is negative", idx, val)
		}
	}
}

// assertFloat64InRange asserts that count consecutive Float64 calls produce
// values in [0.0, 1.0).
func assertFloat64InRange(
	t *testing.T,
	rng interface{ Float64() float64 },
	count int,
) {
	t.Helper()

	for idx := range count {
		val := rng.Float64()
		if val < floatZero || val >= maxFloat64Exclusive {
			t.Errorf("Float64[%d]=%g out of [0.0, 1.0)", idx, val)
		}
	}
}

// assertIntnInRange asserts that count consecutive Intn(bound) calls produce
// values in [0, bound).
func assertIntnInRange(
	t *testing.T,
	rng interface{ Intn(n int) int },
	count, bound int,
) {
	t.Helper()

	for idx := range count {
		val := rng.Intn(bound)
		if val < intZero || val >= bound {
			t.Errorf("Intn(%d)[%d]=%d out of [0, %d)", bound, idx, val, bound)
		}
	}
}

// TestRealRandInterface verifies that Real() satisfies the Rand interface
// and produces values in valid ranges.
func TestRealRandInterface(t *testing.T) {
	t.Parallel()

	rng := rand.Real()

	assertInt63NonNegative(t, rng, iterRange)
	assertFloat64InRange(t, rng, iterRange)
	assertIntnInRange(t, rng, iterRange, intnBound)
}

// TestDeterministicPrintGolden is a helper that prints the actual sequence
// when run with -v. It is not a correctness assertion; use it to regenerate
// golden values when the PCG implementation changes under a new Go major
// version.
func TestDeterministicPrintGolden(t *testing.T) {
	t.Parallel()

	rng := rand.Deterministic(goldenSeed)

	t.Log("Int63 sequence:")

	for idx := range iterRange {
		t.Logf("  [%02d] %d", idx, rng.Int63())
	}

	t.Log("Float64 sequence:")

	for idx := range iterRange {
		t.Logf("  [%02d] %v", idx, rng.Float64())
	}

	t.Log("Intn(intnBound) sequence:")

	for idx := range goldenIntnCount {
		t.Logf("  [%02d] %d", idx, rng.Intn(intnBound))
	}
}
