package main_test

import (
	"testing"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
)

const (
	wantOK               = 0
	wantTestFailed       = 1
	wantToolError        = 2
	wantCacheLockTimeout = 9
	wantConfigError      = 64
	wantStoryParseError  = 65
	wantKeywordError     = 66
	wantTransportError   = 70
)

// TestExitCodesUnique asserts that each exit code constant is a
// distinct value, exists with its expected numeric value, and that
// the full set matches spec 006 -10.
func TestExitCodesUnique(t *testing.T) {
	t.Parallel()

	// Verify each constant has its spec-mandated value.
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"OK", exitcode.OK, wantOK},
		{"TestFailed", exitcode.TestFailed, wantTestFailed},
		{"ToolError", exitcode.ToolError, wantToolError},
		{"CacheLockTimeout", exitcode.CacheLockTimeout, wantCacheLockTimeout},
		{"ConfigError", exitcode.ConfigError, wantConfigError},
		{"StoryParseError", exitcode.StoryParseError, wantStoryParseError},
		{"KeywordError", exitcode.KeywordError, wantKeywordError},
		{"TransportError", exitcode.TransportError, wantTransportError},
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
