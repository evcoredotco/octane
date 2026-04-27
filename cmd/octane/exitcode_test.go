package main_test

import (
	"testing"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
)

// TestExitCodesUnique asserts that each exit code constant is a
// distinct value, exists with its expected numeric value, and that
// the full set matches spec 006 §10.
func TestExitCodesUnique(t *testing.T) {
	t.Parallel()

	// Verify each constant has its spec-mandated value.
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"OK", exitcode.OK, 0},
		{"TestFailed", exitcode.TestFailed, 1},
		{"ToolError", exitcode.ToolError, 2},
		{"CacheLockTimeout", exitcode.CacheLockTimeout, 9},
		{"ConfigError", exitcode.ConfigError, 64},
		{"StoryParseError", exitcode.StoryParseError, 65},
		{"KeywordError", exitcode.KeywordError, 66},
		{"TransportError", exitcode.TransportError, 70},
	}

	seen := make(map[int]string, len(cases))

	for _, testCase := range cases {
		if testCase.got != testCase.want {
			t.Errorf(
				"exitcode.%s = %d; want %d",
				testCase.name,
				testCase.got,
				testCase.want,
			)
		}

		if prev, dup := seen[testCase.got]; dup {
			t.Errorf(
				"exit code %d assigned to both %q and %q",
				testCase.got,
				prev,
				testCase.name,
			)
		}

		seen[testCase.got] = testCase.name
	}
}
