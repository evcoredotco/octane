// Package registry — white-box test for sort determinism (T-003-22, AC1).
//
// This file uses package registry (not registry_test) so that reset() is
// accessible to isolate this test from the global registry state populated
// by other _test.go files or init() registrations.

package registry

import (
	"fmt"
	"testing"

	"github.com/evcoreco/octane/pkg/engine/rand"
	"github.com/evcoreco/octane/pkg/keywords/api"
)

// ── constants ────────────────────────────────────────────────────────────────

const (
	// sortTestSeed is the fixed RNG seed that makes the registration
	// order reproducible across runs and platforms.
	sortTestSeed uint64 = 0xCAFEBABE_DEADBEEF

	// keywordCount is the number of keywords registered in the sort
	// determinism test. Fifty covers all (Layer × OCPPVersion) buckets
	// multiple times and exercises the within-bucket lexicographic sort.
	keywordCount = 50

	// shuffleStart is the starting index for the Fisher-Yates shuffle
	// upper bound: shuffle begins from keywordCount-1 and stops when
	// idx > shuffleStop.
	shuffleStop = 0

	// shuffleIncrement is the step added to idx when selecting the
	// random swap partner (idx+shuffleIncrement).
	shuffleIncrement = 1

	// sortStartIdx is the first index checked by isSortedByLayerVersionPattern;
	// pairs are checked starting at [1] vs [0].
	sortStartIdx = 1
)

// layerValues returns the two legal Layer values in ascending order.
func layerValues() []api.Layer {
	return []api.Layer{api.LayerPrimitive, api.LayerDomain}
}

// versionValues returns the supported OCPPVersion values.
func versionValues() []api.OCPPVersion {
	return []api.OCPPVersion{api.OCPP16}
}

// ── helpers ──────────────────────────────────────────────────────────────────

// buildShuffledKeywords returns a slice of keywordCount unique api.Keyword
// values in a pseudo-random order determined by rng. Every keyword has a
// unique (Layer, OCPPVersion, Pattern) tuple so no collision panic fires.
func buildShuffledKeywords(rng rand.Rand) []api.Keyword {
	keywords := make([]api.Keyword, keywordCount)

	lv := layerValues()
	vv := versionValues()

	for idx := range keywords {
		keywords[idx] = api.Keyword{ //nolint:exhaustruct // Func nil for sort
			Layer:       lv[idx%len(lv)],
			OCPPVersion: vv[idx%len(vv)],
			Pattern:     fmt.Sprintf("step number %04d executes action", idx),
		}
	}

	// Fisher-Yates shuffle using the injected RNG so the order is
	// deterministic for a fixed seed yet differs from the sorted order.
	for idx := keywordCount - 1; idx > shuffleStop; idx-- {
		jdx := rng.Intn(idx + shuffleIncrement)
		keywords[idx], keywords[jdx] = keywords[jdx], keywords[idx]
	}

	return keywords
}

// isPairOutOfOrder reports whether prev should come after curr
// in (Layer asc, OCPPVersion asc, Pattern lex asc) order.
func isPairOutOfOrder(prev, curr api.Keyword) bool {
	if prev.Layer != curr.Layer {
		return prev.Layer > curr.Layer
	}

	if prev.OCPPVersion != curr.OCPPVersion {
		return prev.OCPPVersion > curr.OCPPVersion
	}

	return prev.Pattern > curr.Pattern
}

// isSortedByLayerVersionPattern reports whether slice is ordered by
// (Layer asc, OCPPVersion asc, Pattern lex asc), the canonical order
// required by constitution principle IV and AC1.
func isSortedByLayerVersionPattern(keywords []api.Keyword) bool {
	for idx := sortStartIdx; idx < len(keywords); idx++ {
		if isPairOutOfOrder(keywords[idx-sortStartIdx], keywords[idx]) {
			return false
		}
	}

	return true
}

// ── tests ────────────────────────────────────────────────────────────────────

// assertIdempotentSort verifies that two slices from successive All() calls
// contain identical entries in the same order.
func assertIdempotentSort(t *testing.T, first, second []api.Keyword) {
	t.Helper()

	for idx := range first {
		fKw := first[idx]
		sKw := second[idx]

		if fKw.Layer == sKw.Layer &&
			fKw.OCPPVersion == sKw.OCPPVersion &&
			fKw.Pattern == sKw.Pattern {
			continue
		}

		t.Errorf(
			"All()[%d] diverges between calls: "+
				"first={%v %v %q} second={%v %v %q}",
			idx,
			fKw.Layer, fKw.OCPPVersion, fKw.Pattern,
			sKw.Layer, sKw.OCPPVersion, sKw.Pattern,
		)
	}
}

// assertSortOrder verifies that the slice is sorted by
// (Layer asc, OCPPVersion asc, Pattern lex asc) and logs
// the full slice on failure.
func assertSortOrder(t *testing.T, keywords []api.Keyword) {
	t.Helper()

	if isSortedByLayerVersionPattern(keywords) {
		return
	}

	t.Error("All() result is not sorted by (Layer, OCPPVersion, Pattern)")

	for idx, keyword := range keywords {
		t.Logf(
			"  [%02d] layer=%v ocpp=%v pattern=%q",
			idx,
			keyword.Layer,
			keyword.OCPPVersion,
			keyword.Pattern,
		)
	}
}

// Test_registry_All_stableSortAfterRandomRegistrations verifies that
// All() returns keywords in (Layer asc, OCPPVersion asc, Pattern lex asc)
// order regardless of the registration order. Fifty keywords are registered
// in a pseudo-random sequence; All() is called twice and both results must
// be identical and correctly sorted (AC1).
func Test_registry_All_stableSortAfterRandomRegistrations(t *testing.T) {
	t.Parallel()

	// Isolate this test from any keywords registered by other files.
	reset()

	rng := rand.Deterministic(sortTestSeed)
	shuffled := buildShuffledKeywords(rng)

	for _, kw := range shuffled {
		Register(kw)
	}

	firstCall := All()
	secondCall := All()

	// Invariant: both calls must return the same number of entries.
	if len(firstCall) != keywordCount {
		t.Fatalf(
			"All() length: want %d, got %d",
			keywordCount,
			len(firstCall),
		)
	}

	assertIdempotentSort(t, firstCall, secondCall)
	assertSortOrder(t, firstCall)
}
