// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Stress test: 100 leaf stories all depending on a single shared prereq.
// Validates correctness under load with MaxParallel:4.
package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/octane-project/octane/pkg/keywords/primitive"
	"github.com/octane-project/octane/pkg/runner"
)

// stressPrereqStory is the single shared prerequisite for all stress leaves.
const stressPrereqStory = `Meta
    Name:      Stress shared prereq
    Id:        stress_prereq
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Stress prereq passes
    When  wait 0s
`

// stressLeafTemplate produces a leaf story that depends on stress_prereq.
func stressLeafTemplate(idx int) string {
	return fmt.Sprintf(`Meta
    Name:      Stress leaf story %03d
    Id:        stress_leaf_%03d
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    stress_prereq
        scope: per-station

Scenario: Stress leaf passes
    When  wait 0s
`, idx, idx)
}

// Test_runner_RunStress verifies correctness of the runner when 100 leaf
// stories all share a single per-station prereq under MaxParallel:4.
func Test_runner_RunStress(t *testing.T) {
	// Skip under -short to keep CI fast.
	if testing.Short() {
		t.Skip("flake: stress test skipped under -short; see T-005-57")
	}

	t.Parallel()

	const totalLeaves = 100
	const maxParallel = 4
	// Total nodes: 1 prereq + 100 leaves.
	const expectedTotal = totalLeaves + 1

	storyDir := t.TempDir()

	writeFile(
		t,
		filepath.Join(storyDir, "stress_prereq.story"),
		stressPrereqStory,
	)

	for i := range totalLeaves {
		name := filepath.Join(
			storyDir,
			fmt.Sprintf("stress_leaf_%03d.story", i),
		)
		writeFile(t, name, stressLeafTemplate(i))
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

	// Invariant: all 101 stories (1 prereq + 100 leaves) must be present.
	if len(result.Stories) != expectedTotal {
		t.Fatalf(
			"len(result.Stories): want %d, got %d",
			expectedTotal,
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

	// Invariant: summary must report all passed.
	if result.Summary.Passed != expectedTotal {
		t.Errorf(
			"Summary.Passed: want %d, got %d",
			expectedTotal,
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
