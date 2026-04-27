package main_test

import (
	"testing"

	"github.com/octane-project/octane/cmd/octane/internal/exitcode"
)

// TestExitCodesUnique asserts that each exit code constant is a
// distinct value and belongs to the expected set.
func TestExitCodesUnique(t *testing.T) {
	t.Parallel()

	codeMap := map[string]int{
		"OK":            exitcode.OK,
		"TestFailed":    exitcode.TestFailed,
		"ConfigError":   exitcode.ConfigError,
		"IOError":       exitcode.IOError,
		"InternalError": exitcode.InternalError,
	}

	allowed := map[int]bool{
		0:   true,
		1:   true,
		64:  true,
		74:  true,
		125: true,
	}

	seen := make(map[int]string, len(codeMap))

	for name, code := range codeMap {
		if prev, dup := seen[code]; dup {
			t.Errorf(
				"exit code %d assigned to both %q and %q",
				code,
				prev,
				name,
			)
		}

		seen[code] = name

		if !allowed[code] {
			t.Errorf(
				"exit code %q = %d is not in the allowed set",
				name,
				code,
			)
		}
	}
}
