// Task: T-007-13.

package redact_test

import (
	"fmt"
	mrand "math/rand/v2"
	"testing"

	"github.com/evcoreco/octane/pkg/report/internal/redact"
)

// propertyIterations is the number of random inputs tested.
const propertyIterations = 1000

// propertySeed is the deterministic seed for reproducibility
// (constitution principle IV).
const propertySeed uint64 = 0xABCD_1234_5678_EF01

// propertyBinaryChoice is the modulus for a coin-flip in property tests.
const propertyBinaryChoice = 2

// propertyMaxVariance is the upper bound for random value variance in
// property tests.
const propertyMaxVariance = 10000

// credentialKeyNames returns the three known credential key names used in
// OCTANE connection profiles.
func credentialKeyNames() []string {
	return []string{"token", "password", "basic"}
}

// assertNoCredentialLeaks checks that every credential key in got is
// redacted to [redact.Placeholder]. Any non-redacted key is reported
// via t.Errorf.
func assertNoCredentialLeaks(t *testing.T, got map[string]any, iter int) {
	t.Helper()

	for _, key := range credentialKeyNames() {
		val, present := got[key]
		if !present {
			continue
		}

		if val != redact.Placeholder {
			t.Errorf(
				"iter %d: key %q not redacted: got %q",
				iter, key, val,
			)
		}
	}
}

// Test_redact_AuthBlock_property verifies that no credential key in
// a randomly generated auth map retains its original value after
// AuthBlock. Any key whose value is not [redact.Placeholder] after
// redaction would represent a leak.
func Test_redact_AuthBlock_property(t *testing.T) {
	t.Parallel()

	//nolint:gosec // G404: seeded PCG for deterministic tests, not security
	rng := mrand.New(mrand.NewPCG(propertySeed, propertySeed^0xDEADBEEF))

	for iter := range propertyIterations {
		input := buildRandomAuthMap(rng, iter)
		got := redact.AuthBlock(input)

		assertNoCredentialLeaks(t, got, iter)
	}
}

// buildRandomAuthMap constructs a map[string]any with a random subset
// of credential keys and randomised string values. The iter parameter
// is incorporated into values to ensure uniqueness.
func buildRandomAuthMap(rng *mrand.Rand, iter int) map[string]any {
	keys := credentialKeyNames()
	out := make(map[string]any, len(keys))

	for _, key := range keys {
		if rng.IntN(propertyBinaryChoice) == 0 {
			out[key] = fmt.Sprintf(
				"secret-%s-%d-%d",
				key,
				iter,
				rng.IntN(propertyMaxVariance),
			)
		}
	}

	return out
}
