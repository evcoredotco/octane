// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-50
// AC1: A 4-deep dependency chain executes in topological order and all stories pass.
package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	// Registers wait {duration:duration} and other primitive keywords.
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// storyChainA is the root of a 4-deep chain: no dependencies.
const storyChainA = `Meta
    Name:      Chain story A
    Id:        chain_a
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: A passes
    When  wait 0s
`

// storyChainB depends on chain_a.
const storyChainB = `Meta
    Name:      Chain story B
    Id:        chain_b
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    chain_a
        scope: per-station

Scenario: B passes
    When  wait 0s
`

// storyChainC depends on chain_b.
const storyChainC = `Meta
    Name:      Chain story C
    Id:        chain_c
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    chain_b
        scope: per-station

Scenario: C passes
    When  wait 0s
`

// storyChainD depends on chain_c (leaf of the chain).
const storyChainD = `Meta
    Name:      Chain story D
    Id:        chain_d
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    chain_c
        scope: per-station

Scenario: D passes
    When  wait 0s
`

// Test_runner_RunChain asserts that a 4-deep dependency chain executes in
// topological order, all stories pass, and all cache statuses are CacheBypassed.
func Test_runner_RunChain(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Write story files to temp directory.
	writeFile(t, filepath.Join(dir, "chain_a.story"), storyChainA)
	writeFile(t, filepath.Join(dir, "chain_b.story"), storyChainB)
	writeFile(t, filepath.Join(dir, "chain_c.story"), storyChainC)
	writeFile(t, filepath.Join(dir, "chain_d.story"), storyChainD)

	cfg := runner.Config{
		StoryPaths: []string{dir},
		NoCache:    true,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runner.Run: %v", err)
	}

	// Invariant: all 4 stories must be present in the result.
	const expectedStories = 4
	if len(result.Stories) != expectedStories {
		t.Fatalf(
			"len(result.Stories): want %d, got %d",
			expectedStories,
			len(result.Stories),
		)
	}

	// Invariant: all stories must have passed with cache bypassed.
	for _, sr := range result.Stories {
		if sr.Status != runner.StatusPassed {
			t.Errorf(
				"story %q: want StatusPassed, got %s",
				sr.TestID,
				sr.Status,
			)
		}

		if sr.CacheStatus != runner.CacheBypassed {
			t.Errorf(
				"story %q: want CacheBypassed, got %s",
				sr.TestID,
				sr.CacheStatus,
			)
		}
	}

	// Build a map of testID → Order to verify topological ordering.
	orderByID := make(map[string]int, len(result.Stories))
	for _, sr := range result.Stories {
		orderByID[sr.TestID] = sr.Order
	}

	requiredIDs := []string{"chain_a", "chain_b", "chain_c", "chain_d"}
	for _, id := range requiredIDs {
		if _, ok := orderByID[id]; !ok {
			t.Errorf("story %q missing from results", id)
		}
	}

	// Invariant: topological order must hold: a < b < c < d.
	type orderCheck struct {
		before string
		after  string
	}

	checks := []orderCheck{
		{before: "chain_a", after: "chain_b"},
		{before: "chain_b", after: "chain_c"},
		{before: "chain_c", after: "chain_d"},
	}

	for _, chk := range checks {
		if orderByID[chk.before] >= orderByID[chk.after] {
			t.Errorf(
				"order violation: %q (order=%d) must precede %q (order=%d)",
				chk.before, orderByID[chk.before],
				chk.after, orderByID[chk.after],
			)
		}
	}

	// Invariant: result.Stories must be sorted by Order.
	if !sort.SliceIsSorted(result.Stories, func(i, j int) bool {
		return result.Stories[i].Order < result.Stories[j].Order
	}) {
		t.Error("result.Stories is not sorted by Order field")
	}
}

// writeFile is a test helper that writes content to path, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeFile(%q): %v", path, err)
	}
}
