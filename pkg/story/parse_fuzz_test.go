// Package story_test — fuzz test for the story parser (T-001-43).
//
// FuzzParse feeds random byte sequences to Parse and asserts that the parser
// never panics. Errors are acceptable; panics are not.
package story_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/story"
)

// FuzzParse is a property-based fuzz test: any byte sequence fed to
// story.Parse must not cause a panic. Errors are expected and ignored.
// The corpus seeds are representative slices of the .story DSL grammar.
func FuzzParse(f *testing.F) {
	// Seed corpus: minimal valid story.
	f.Add([]byte(
		"Meta\n" +
			"    Name: x\n" +
			"    Id:   x\n" +
			"    Tags: helper\n" +
			"    Stations: 1\n\n" +
			"Scenario: s\n" +
			"    When action\n",
	))

	// Seed corpus: empty input.
	f.Add([]byte(""))

	// Seed corpus: comment-only file.
	f.Add([]byte("# just a comment\n"))

	// Seed corpus: depends block.
	f.Add([]byte(
		"Meta\n" +
			"    Name:     Boot\n" +
			"    Id:       boot\n" +
			"    Spec-Ref: OCPP 1.6 -B01\n" +
			"    Tags:     core\n" +
			"    Stations: 1\n" +
			"    Depends:\n" +
			"      - id:    other\n" +
			"        scope: per-station\n\n" +
			"Scenario: s\n" +
			"    Given precondition\n",
	))

	// Seed corpus: binary noise.
	f.Add([]byte{0x00, 0xff, 0x80, 0x01, 0x7f})

	f.Fuzz(func(_ *testing.T, data []byte) {
		// Must not panic; error is OK.
		_, _ = story.Parse("fuzz.story", data)
	})
}
