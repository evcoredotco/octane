// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-52
// AC4: Dependent story is skipped when its prerequisite fails. The Findings
// field on the skipped story references the failing prerequisite's ID.

package integration_test

import (
	"context"
	"strings"
	"testing"

	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// storyAlwaysFails uses an unrecognised step so the runner produces ErrNoMatch.
const storyAlwaysFails = `Meta
    Name:      Always failing story
    Id:        always_fails
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: This will fail
    When  this step always fails
`

// storySkipDependent depends on always_fails and must be skipped.
const storySkipDependent = `Meta
    Name:      Skip dependent story
    Id:        skip_dependent
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    always_fails
        scope: per-station

Scenario: Should be skipped
    When  wait 0s
`

// Test_runner_RunDependentSkippedOnPrereqFailure asserts that when a prereq
// fails, its dependent is marked StatusSkipped with a finding referencing the
// prereq ID.
func Test_runner_RunDependentSkippedOnPrereqFailure(t *testing.T) {
	t.Parallel()

	storyDir := t.TempDir()

	writeFile(t, storyDir+"/always_fails.story", storyAlwaysFails)
	writeFile(t, storyDir+"/skip_dependent.story", storySkipDependent)

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    true,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	const expectedStories = 2
	if len(result.Stories) != expectedStories {
		t.Fatalf(
			"len(result.Stories): want %d, got %d",
			expectedStories,
			len(result.Stories),
		)
	}

	byID := storyResultsByTestID(result.Stories)

	// Invariant: the prerequisite must be StatusFailed.
	prereq, ok := byID["always_fails"]
	if !ok {
		t.Fatal("story always_fails not found in results")
	}

	if prereq.Status != runner.StatusFailed {
		t.Errorf("always_fails: want StatusFailed, got %s", prereq.Status)
	}

	// Invariant: the dependent must be StatusSkipped.
	dependent, ok := byID["skip_dependent"]
	if !ok {
		t.Fatal("story skip_dependent not found in results")
	}

	if dependent.Status != runner.StatusSkipped {
		t.Errorf("skip_dependent: want StatusSkipped, got %s", dependent.Status)
	}

	// Invariant: the dependent's Findings must reference the failing prereq ID.
	if !findingsContain(dependent.Findings, "always_fails") {
		t.Errorf(
			"skip_dependent Findings must reference 'always_fails'; got %v",
			dependent.Findings,
		)
	}
}

// storyResultsByTestID indexes a slice of StoryResult by TestID for O(1) lookup.
func storyResultsByTestID(
	stories []runner.StoryResult,
) map[string]runner.StoryResult {
	m := make(map[string]runner.StoryResult, len(stories))

	for _, sr := range stories {
		m[sr.TestID] = sr
	}

	return m
}

// findingsContain reports whether any finding message in findings contains substr.
func findingsContain(findings []runner.Finding, substr string) bool {
	for _, f := range findings {
		if strings.Contains(f.Message, substr) {
			return true
		}
	}

	return false
}
