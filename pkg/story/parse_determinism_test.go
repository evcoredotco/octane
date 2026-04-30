// Package story_test — black-box tests for the story parser (T-001-31).

package story_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/evcoreco/octane/pkg/story"
	"github.com/evcoreco/octane/pkg/story/internal/serialize"
)

const (
	noStories    = 0
	firstIterIdx = 1
)

// TestDeterminism parses every .story file under scenarios/ 1 000 times and
// asserts that the JSON-serialized AST is byte-identical on every iteration.
// This covers AC5: the parser must produce the same output for the same input
// regardless of invocation count or runtime state.
func TestDeterminism(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "scenarios")

	storyPaths := collectDeterminismPaths(t, root)

	for _, p := range storyPaths {
		t.Run(filepath.ToSlash(p), func(t *testing.T) {
			t.Parallel()
			runDeterminismCheck(t, p)
		})
	}
}

// collectDeterminismPaths walks root and returns all .story file paths.
// It calls t.Fatal when the walk fails or yields no files.
func collectDeterminismPaths(t *testing.T, root string) []string {
	t.Helper()

	var storyPaths []string

	err := filepath.WalkDir(root, collectStoryEntry(&storyPaths))
	if err != nil {
		t.Fatalf("walking scenarios/: %v", err)
	}

	if len(storyPaths) == noStories {
		t.Fatal("no .story files found under scenarios/")
	}

	return storyPaths
}

// collectStoryEntry returns a WalkDir callback that appends .story file paths
// to dst.
func collectStoryEntry(dst *[]string) func(string, os.DirEntry, error) error {
	return func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && filepath.Ext(path) == ".story" {
			*dst = append(*dst, path)
		}

		return nil
	}
}

// runDeterminismCheck parses path 1 000 times and asserts byte-identical JSON
// on every iteration. It is extracted from TestDeterminism to keep the parent
// function's cognitive complexity within the configured limit.
func runDeterminismCheck(t *testing.T, path string) {
	t.Helper()

	src, readErr := os.ReadFile(filepath.Clean(path))
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

	for idx := firstIterIdx; idx < iterations; idx++ {
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
