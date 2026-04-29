// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-54
// AC6: Per-run prereq runs exactly once regardless of how many dependents
// reference it. Three dependents sharing a per-run prereq produce exactly
// one prereq StoryResult entry.

package integration_test

import (
	"context"
	"testing"

	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// storyPerRunPrereq is a per-run-scoped prereq used by three stories.
const storyPerRunPrereq = `Meta
    Name:      Per-run prereq
    Id:        pr_prereq
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Per-run prereq passes
    When  wait 0s
`

// storyPerRunDep1 depends on pr_prereq with per-run scope.
const storyPerRunDep1 = `Meta
    Name:      Per-run dependent 1
    Id:        pr_dep1
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    pr_prereq
        scope: per-run

Scenario: Dependent 1 passes
    When  wait 0s
`

// storyPerRunDep2 depends on pr_prereq with per-run scope.
const storyPerRunDep2 = `Meta
    Name:      Per-run dependent 2
    Id:        pr_dep2
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    pr_prereq
        scope: per-run

Scenario: Dependent 2 passes
    When  wait 0s
`

// storyPerRunDep3 depends on pr_prereq with per-run scope.
const storyPerRunDep3 = `Meta
    Name:      Per-run dependent 3
    Id:        pr_dep3
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    pr_prereq
        scope: per-run

Scenario: Dependent 3 passes
    When  wait 0s
`

// Test_runner_RunPerRunPrereqRunsOnce asserts that a per-run prereq appears
// exactly once in the results (with ScopeKey == RunID) even though three
// dependents reference it.
func Test_runner_RunPerRunPrereqRunsOnce(t *testing.T) {
	t.Parallel()

	storyDir := t.TempDir()

	writeFile(t, storyDir+"/pr_prereq.story", storyPerRunPrereq)
	writeFile(t, storyDir+"/pr_dep1.story", storyPerRunDep1)
	writeFile(t, storyDir+"/pr_dep2.story", storyPerRunDep2)
	writeFile(t, storyDir+"/pr_dep3.story", storyPerRunDep3)

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    true,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	// Collect prereq entries.
	var prereqEntries []runner.StoryResult

	for _, sr := range result.Stories {
		if sr.TestID == "pr_prereq" {
			prereqEntries = append(prereqEntries, sr)
		}
	}

	// Invariant: per-run prereq must appear exactly once.
	const expectedPrereqCount = 1
	if len(prereqEntries) != expectedPrereqCount {
		t.Fatalf(
			"pr_prereq instance count: want %d, got %d",
			expectedPrereqCount,
			len(prereqEntries),
		)
	}

	// Invariant: the single prereq instance ScopeKey must equal the RunID.
	prereqScopeKey := prereqEntries[0].ScopeKey
	if prereqScopeKey != result.RunID {
		t.Errorf(
			"pr_prereq ScopeKey: want RunID %q, got %q",
			result.RunID,
			prereqScopeKey,
		)
	}

	// Invariant: prereq must have passed.
	if prereqEntries[0].Status != runner.StatusPassed {
		t.Errorf(
			"pr_prereq: want StatusPassed, got %s",
			prereqEntries[0].Status,
		)
	}

	// Invariant: all three dependents must have passed.
	dependentIDs := []string{"pr_dep1", "pr_dep2", "pr_dep3"}
	byID := storyResultsByTestID(result.Stories)

	for _, id := range dependentIDs {
		sr, ok := byID[id]
		if !ok {
			t.Errorf("story %q missing from results", id)

			continue
		}

		if sr.Status != runner.StatusPassed {
			t.Errorf("story %q: want StatusPassed, got %s", id, sr.Status)
		}
	}

	// Invariant: exactly 4 stories total (1 prereq + 3 dependents).
	const expectedTotal = 4
	if len(result.Stories) != expectedTotal {
		t.Errorf(
			"total stories: want %d, got %d",
			expectedTotal,
			len(result.Stories),
		)
	}
}
