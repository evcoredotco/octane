// Package registry — white-box test for sort determinism (T-003-22, AC1).
//
// This file uses package registry (not registry_test) so that reset() is
// accessible to isolate this test from the global registry state populated
// by other _test.go files or init() registrations.
package registry

import (
	"fmt"
	"testing"

	"github.com/octane-project/octane/pkg/engine/rand"
	"github.com/octane-project/octane/pkg/keywords/api"
)

// ── constants ─────────────────────────────────────────────────────────────────

const (
	// sortTestSeed is the fixed RNG seed that makes the registration
	// order reproducible across runs and platforms.
	sortTestSeed uint64 = 0xCAFEBABE_DEADBEEF

	// keywordCount is the number of keywords registered in the sort
	// determinism test. Fifty covers all (Layer × OCPPVersion) buckets
	// multiple times and exercises the within-bucket lexicographic sort.
	keywordCount = 50
)

// layerValues enumerates the two legal Layer values in ascending order.
var layerValues = []api.Layer{api.LayerPrimitive, api.LayerDomain}

// versionValues enumerates the three legal OCPPVersion values in
// ascending order.
var versionValues = []api.OCPPVersion{api.OCPP16, api.OCPP201, api.OCPP21}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildShuffledKeywords returns a slice of keywordCount unique api.Keyword
// values in a pseudo-random order determined by rng. Every keyword has a
// unique (Layer, OCPPVersion, Pattern) tuple so no collision panic fires.
func buildShuffledKeywords(rng rand.Rand) []api.Keyword {
	keywords := make([]api.Keyword, keywordCount)

	for idx := range keywords {
		keywords[idx] = api.Keyword{ //nolint:exhaustruct // Func intentionally nil for sort tests
			Layer:       layerValues[idx%len(layerValues)],
			OCPPVersion: versionValues[idx%len(versionValues)],
			Pattern:     fmt.Sprintf("step number %04d executes action", idx),
		}
	}

	// Fisher-Yates shuffle using the injected RNG so the order is
	// deterministic for a fixed seed yet differs from the sorted order.
	for idx := keywordCount - 1; idx > 0; idx-- {
		jdx := rng.Intn(idx + 1)
		keywords[idx], keywords[jdx] = keywords[jdx], keywords[idx]
	}

	return keywords
}

// isSortedByLayerVersionPattern reports whether slice is ordered by
// (Layer asc, OCPPVersion asc, Pattern lex asc), the canonical order
// required by constitution principle IV and AC1.
func isSortedByLayerVersionPattern(keywords []api.Keyword) bool {
	for idx := 1; idx < len(keywords); idx++ {
		prev := keywords[idx-1]
		curr := keywords[idx]

		if prev.Layer > curr.Layer {
			return false
		}

		if prev.Layer == curr.Layer && prev.OCPPVersion > curr.OCPPVersion {
			return false
		}

		sameLayerAndVersion := prev.Layer == curr.Layer &&
			prev.OCPPVersion == curr.OCPPVersion

		if sameLayerAndVersion && prev.Pattern > curr.Pattern {
			return false
		}
	}

	return true
}

// ── tests ─────────────────────────────────────────────────────────────────────

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

	// Invariant: the two results must be identical (idempotent sort).
	for idx := range firstCall {
		firstKw := firstCall[idx]
		secondKw := secondCall[idx]

		if firstKw.Layer != secondKw.Layer ||
			firstKw.OCPPVersion != secondKw.OCPPVersion ||
			firstKw.Pattern != secondKw.Pattern {
			t.Errorf(
				"All()[%d] diverges between calls: "+
					"first={%v %v %q} second={%v %v %q}",
				idx,
				firstKw.Layer, firstKw.OCPPVersion, firstKw.Pattern,
				secondKw.Layer, secondKw.OCPPVersion, secondKw.Pattern,
			)
		}
	}

	// Invariant: the result must be sorted by (Layer, OCPPVersion, Pattern).
	if !isSortedByLayerVersionPattern(firstCall) {
		t.Error("All() result is not sorted by (Layer, OCPPVersion, Pattern)")

		for idx, keyword := range firstCall {
			t.Logf(
				"  [%02d] layer=%v ocpp=%v pattern=%q",
				idx,
				keyword.Layer,
				keyword.OCPPVersion,
				keyword.Pattern,
			)
		}
	}
}
