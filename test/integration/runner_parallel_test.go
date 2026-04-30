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

	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// parallelLeafStory produces a self-contained passing story for the given
// index.
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

	const (
		totalLeaves = 16
		maxParallel = 4
	)

	storyDir := t.TempDir()

	for i := range totalLeaves {
		name := filepath.Join(
			storyDir,
			fmt.Sprintf("parallel_leaf_%02d.story", i),
		)
		writeFile(t, name, parallelLeafStory(i))
	}

	cfg := noopCfg(storyDir)
	cfg.MaxParallel = maxParallel

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

	assertAllPassedBypassed(t, result.Stories)
	assertSummaryAllPassed(t, result.Summary, totalLeaves)
}

// assertSummaryAllPassed checks that Summary reflects all-passed with no
// failures or skips.
func assertSummaryAllPassed(
	t *testing.T,
	summary runner.Summary,
	want int,
) {
	t.Helper()

	if summary.Passed != want {
		t.Errorf("Summary.Passed: want %d, got %d", want, summary.Passed)
	}

	const wantZero = 0
	if summary.Failed != wantZero {
		t.Errorf("Summary.Failed: want 0, got %d", summary.Failed)
	}

	if summary.Skipped != wantZero {
		t.Errorf("Summary.Skipped: want 0, got %d", summary.Skipped)
	}
}
