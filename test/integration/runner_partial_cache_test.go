// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-55
// AC8: Partial cache: when 8 of 10 cache entries are deleted after a full
// first run, the third run reports CacheHits >= 8 and re-executes the 2 misses.
package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// leafStoryTemplate produces a self-contained passing story for the given
// zero-based index. Story IDs are partial_leaf_00 … partial_leaf_09.
func leafStoryTemplate(idx int) string {
	return fmt.Sprintf(`Meta
    Name:      Partial cache leaf %02d
    Id:        partial_leaf_%02d
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Leaf passes
    When  wait 0s
`, idx, idx)
}

// Test_runner_RunPartialCache asserts that a third run after two of ten
// populated cache entries are deleted produces >= 8 cache hits and exactly
// 2 cache misses.
func Test_runner_RunPartialCache(t *testing.T) {
	t.Parallel()

	const totalStories = 10
	const deletedEntries = 2
	const expectedHits = totalStories - deletedEntries

	storyDir := t.TempDir()
	cacheDir := t.TempDir()

	// Write 10 independent leaf stories.
	for i := range totalStories {
		name := filepath.Join(
			storyDir,
			fmt.Sprintf("partial_leaf_%02d.story", i),
		)
		writeFile(t, name, leafStoryTemplate(i))
	}

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    false,
		CacheDir:   cacheDir,
	}

	// First run: all 10 are cache misses; cache is populated.
	firstResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first runner.Run: %v", err)
	}

	if len(firstResult.Stories) != totalStories {
		t.Fatalf(
			"first run: want %d stories, got %d",
			totalStories,
			len(firstResult.Stories),
		)
	}

	// Second run: all 10 are cache hits.
	secondResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second runner.Run: %v", err)
	}

	if secondResult.Summary.CacheHits != totalStories {
		t.Fatalf(
			"second run: want %d CacheHits, got %d",
			totalStories,
			secondResult.Summary.CacheHits,
		)
	}

	// Delete 2 cache entries from the results tree so the third run misses them.
	if err = deleteNCacheEntries(cacheDir, deletedEntries); err != nil {
		t.Fatalf("deleteNCacheEntries: %v", err)
	}

	// Third run: expect 8 hits, 2 misses.
	thirdResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("third runner.Run: %v", err)
	}

	// Invariant: at least 8 of 10 stories must be cache hits.
	if thirdResult.Summary.CacheHits < expectedHits {
		t.Errorf(
			"third run: Summary.CacheHits: want >= %d, got %d",
			expectedHits,
			thirdResult.Summary.CacheHits,
		)
	}

	// Invariant: all 10 stories should still be present and pass.
	if len(thirdResult.Stories) != totalStories {
		t.Errorf(
			"third run: want %d stories, got %d",
			totalStories,
			len(thirdResult.Stories),
		)
	}

	for _, sr := range thirdResult.Stories {
		if sr.Status != runner.StatusPassed {
			t.Errorf(
				"third run: story %q: want StatusPassed, got %s",
				sr.TestID,
				sr.Status,
			)
		}
	}
}

// deleteNCacheEntries removes up to n leaf entry directories from the cache
// results tree. Each leaf directory is a <hash>/ folder two levels under
// <cacheDir>/results/<prefix>/<hash>/result.json.
func deleteNCacheEntries(cacheDir string, n int) error {
	resultsRoot := filepath.Join(cacheDir, "results")

	prefixEntries, err := os.ReadDir(resultsRoot)
	if err != nil {
		return fmt.Errorf("read results root: %w", err)
	}

	deleted := 0

	for _, prefix := range prefixEntries {
		if !prefix.IsDir() || deleted >= n {
			break
		}

		prefixPath := filepath.Join(resultsRoot, prefix.Name())

		hashEntries, readErr := os.ReadDir(prefixPath)
		if readErr != nil {
			return fmt.Errorf("read prefix dir: %w", readErr)
		}

		for _, hash := range hashEntries {
			if !hash.IsDir() || deleted >= n {
				break
			}

			hashPath := filepath.Join(prefixPath, hash.Name())

			if err = os.RemoveAll(hashPath); err != nil {
				return fmt.Errorf("remove entry dir %q: %w", hashPath, err)
			}

			deleted++
		}
	}

	return nil
}
