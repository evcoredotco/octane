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

	cfg := noopCfg(storyDir)

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	prereqResults := collectByTestID(result.Stories, "ps_prereq")

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

	assertPerStationScopeKeys(t, prereqResults)
}

// assertPerStationScopeKeys checks that the two prereq results have distinct
// scope keys equal to CP01 and CP02, and that both passed.
func assertPerStationScopeKeys(
	t *testing.T,
	prereqResults []runner.StoryResult,
) {
	t.Helper()

	const (
		// firstKey is the index of the first scope key.
		firstKey = 0
		// secondKey is the index of the second scope key.
		secondKey = 1
	)

	scopeKeys := prereqScopeKeys(prereqResults)

	if scopeKeys[firstKey] == scopeKeys[secondKey] {
		t.Errorf(
			"ps_prereq instances share the same ScopeKey %q",
			scopeKeys[firstKey],
		)
	}

	expectedHandles := map[string]bool{"CP01": true, "CP02": true}

	for _, sk := range scopeKeys {
		if !expectedHandles[sk] {
			t.Errorf(
				"unexpected ScopeKey %q for ps_prereq; want CP01 or CP02",
				sk,
			)
		}
	}

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
