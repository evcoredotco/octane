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

// credentialKeys are the three known credential key names used in
// OCTANE connection profiles.
var credentialKeys = []string{"token", "password", "basic"}

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

		for _, key := range credentialKeys {
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
}

// buildRandomAuthMap constructs a map[string]any with a random subset
// of credential keys and randomised string values. The iter parameter
// is incorporated into values to ensure uniqueness.
func buildRandomAuthMap(rng *mrand.Rand, iter int) map[string]any {
	out := make(map[string]any, len(credentialKeys))

	for _, key := range credentialKeys {
		if rng.IntN(2) == 0 {
			out[key] = fmt.Sprintf(
				"secret-%s-%d-%d",
				key,
				iter,
				rng.IntN(10000),
			)
		}
	}

	return out
}
