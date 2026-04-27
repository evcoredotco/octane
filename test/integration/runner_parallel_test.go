// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-56
// AC9: 16 leaf stories with no inter-dependencies all pass when MaxParallel:4.
package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/octane-project/octane/pkg/keywords/primitive"
	"github.com/octane-project/octane/pkg/runner"
)

// parallelLeafStory produces a self-contained passing story for the given index.
func parallelLeafStory(idx int) string {
	return fmt.Sprintf(`Meta
    Name:      Parallel leaf story %02d
    Id:        parallel_leaf_%02d
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Leaf passes
    When  wait 0s
`, idx, idx)
}

// Test_runner_RunParallelLeafStories asserts that 16 independent leaf stories
// all pass when MaxParallel is 4 and the cache is bypassed.
func Test_runner_RunParallelLeafStories(t *testing.T) {
	t.Parallel()

	const totalLeaves = 16
	const maxParallel = 4

	storyDir := t.TempDir()

	for i := range totalLeaves {
		name := filepath.Join(
			storyDir,
			fmt.Sprintf("parallel_leaf_%02d.story", i),
		)
		writeFile(t, name, parallelLeafStory(i))
	}

	cfg := runner.Config{
		StoryPaths:  []string{storyDir},
		NoCache:     true,
		MaxParallel: maxParallel,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	// Invariant: all 16 leaf stories must be present in results.
	if len(result.Stories) != totalLeaves {
		t.Fatalf(
			"len(result.Stories): want %d, got %d",
			totalLeaves,
			len(result.Stories),
		)
	}

	// Invariant: all stories must have passed.
	for _, sr := range result.Stories {
		if sr.Status != runner.StatusPassed {
			t.Errorf(
				"story %q: want StatusPassed, got %s",
				sr.TestID,
				sr.Status,
			)
		}
	}

	// Invariant: cache status must be CacheBypassed for all stories.
	for _, sr := range result.Stories {
		if sr.CacheStatus != runner.CacheBypassed {
			t.Errorf(
				"story %q: want CacheBypassed, got %s",
				sr.TestID,
				sr.CacheStatus,
			)
		}
	}

	// Invariant: summary counts must match.
	if result.Summary.Passed != totalLeaves {
		t.Errorf(
			"Summary.Passed: want %d, got %d",
			totalLeaves,
			result.Summary.Passed,
		)
	}

	if result.Summary.Failed != 0 {
		t.Errorf("Summary.Failed: want 0, got %d", result.Summary.Failed)
	}

	if result.Summary.Skipped != 0 {
		t.Errorf("Summary.Skipped: want 0, got %d", result.Summary.Skipped)
	}
}
