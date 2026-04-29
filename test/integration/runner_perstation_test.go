// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-53
// AC5: Per-station prereq runs once per station. With Stations:2 on the
// dependent, the prereq executes twice (one StoryResult per station handle).

package integration_test

import (
	"context"
	"testing"

	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// storyPerStationPrereq is a per-station prereq story.
const storyPerStationPrereq = `Meta
    Name:      Per-station prereq
    Id:        ps_prereq
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Per-station prereq passes
    When  wait 0s
`

// storyPerStationDependent depends on ps_prereq with per-station scope and
// declares Stations:2 so the prereq must run for both CP01 and CP02.
const storyPerStationDependent = `Meta
    Name:      Per-station dependent
    Id:        ps_dependent
    Tags:      helper
    Stations:  2
    Timeout:   10s
    Depends:
      - id:    ps_prereq
        scope: per-station

Scenario: Per-station dependent passes
    When  wait 0s
`

// Test_runner_RunPerStationPrereqRunsTwice asserts that a per-station prereq
// produces two StoryResult entries (one per station) when the dependent
// declares Stations:2.
func Test_runner_RunPerStationPrereqRunsTwice(t *testing.T) {
	t.Parallel()

	storyDir := t.TempDir()

	writeFile(t, storyDir+"/ps_prereq.story", storyPerStationPrereq)
	writeFile(t, storyDir+"/ps_dependent.story", storyPerStationDependent)

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    true,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	// Collect all StoryResult entries whose TestID is "ps_prereq".
	var prereqResults []runner.StoryResult

	for _, sr := range result.Stories {
		if sr.TestID == "ps_prereq" {
			prereqResults = append(prereqResults, sr)
		}
	}

	// Invariant: the per-station prereq must appear exactly twice (CP01, CP02).
	const expectedInstances = 2
	if len(prereqResults) != expectedInstances {
		t.Fatalf(
			"ps_prereq instance count: want %d, got %d (results: %v)",
			expectedInstances,
			len(prereqResults),
			prereqScopeKeys(prereqResults),
		)
	}

	// Invariant: the two instances must have distinct scope keys.
	scopeKeys := prereqScopeKeys(prereqResults)
	if scopeKeys[0] == scopeKeys[1] {
		t.Errorf("ps_prereq instances share the same ScopeKey %q", scopeKeys[0])
	}

	// Invariant: both must contain one of the expected station handles.
	expectedHandles := map[string]bool{"CP01": true, "CP02": true}
	for _, sk := range scopeKeys {
		if !expectedHandles[sk] {
			t.Errorf(
				"unexpected ScopeKey %q for ps_prereq; want CP01 or CP02",
				sk,
			)
		}
	}

	// Invariant: all prereq instances must have passed.
	for _, sr := range prereqResults {
		if sr.Status != runner.StatusPassed {
			t.Errorf(
				"ps_prereq/%s: want StatusPassed, got %s",
				sr.ScopeKey,
				sr.Status,
			)
		}
	}
}

// prereqScopeKeys extracts the ScopeKey field from each StoryResult.
func prereqScopeKeys(results []runner.StoryResult) []string {
	keys := make([]string, len(results))

	for i, sr := range results {
		keys[i] = sr.ScopeKey
	}

	return keys
}
