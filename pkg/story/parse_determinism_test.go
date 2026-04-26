// Package story_test — black-box tests for the story parser (T-001-31).
package story_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/octane-project/octane/pkg/story"
	"github.com/octane-project/octane/pkg/story/internal/serialize"
)

// TestDeterminism parses every .story file under scenarios/ 1 000 times and
// asserts that the JSON-serialized AST is byte-identical on every iteration.
// This covers AC5: the parser must produce the same output for the same input
// regardless of invocation count or runtime state.
func TestDeterminism(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "scenarios")

	var storyPaths []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && filepath.Ext(path) == ".story" {
			storyPaths = append(storyPaths, path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking scenarios/: %v", err)
	}

	if len(storyPaths) == 0 {
		t.Fatal("no .story files found under scenarios/")
	}

	for _, p := range storyPaths {
		p := p // capture for sub-test

		t.Run(filepath.ToSlash(p), func(t *testing.T) {
			t.Parallel()
			runDeterminismCheck(t, p)
		})
	}
}

// runDeterminismCheck parses path 1 000 times and asserts byte-identical JSON
// on every iteration. It is extracted from TestDeterminism to keep the parent
// function's cognitive complexity within the configured limit.
func runDeterminismCheck(t *testing.T, path string) {
	t.Helper()

	src, readErr := os.ReadFile(path) //nolint:gosec // test file, path from walk
	if readErr != nil {
		t.Fatalf("reading %s: %v", path, readErr)
	}

	first, parseErr := story.Parse(path, src)
	if parseErr != nil {
		t.Fatalf("first parse of %s failed: %v", path, parseErr)
	}

	ref, serErr := serialize.Serialize(first)
	if serErr != nil {
		t.Fatalf("serialize first parse of %s: %v", path, serErr)
	}

	const iterations = 1000

	for idx := 1; idx < iterations; idx++ {
		assertIterationMatch(t, path, src, ref, idx)
	}
}

// assertIterationMatch performs one parse iteration and compares against ref.
func assertIterationMatch(
	t *testing.T,
	path string,
	src []byte,
	ref []byte,
	idx int,
) {
	t.Helper()

	got, parseErr := story.Parse(path, src)
	if parseErr != nil {
		t.Fatalf("iteration %d: parse of %s failed: %v", idx, path, parseErr)
	}

	gotBytes, serErr := serialize.Serialize(got)
	if serErr != nil {
		t.Fatalf("iteration %d: serialize of %s failed: %v", idx, path, serErr)
	}

	if !bytes.Equal(ref, gotBytes) {
		t.Errorf(
			"iteration %d: parse of %s is not deterministic: "+
				"JSON differs from first parse",
			idx, path,
		)
		t.Logf("reference: %s", ref)
		t.Logf("got:       %s", gotBytes)
	}
}
