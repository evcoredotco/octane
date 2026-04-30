// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-50
// AC1: A 4-deep dependency chain runs in topological order; all stories pass.

package integration_test

import (
	"cmp"
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	// registers primitive keywords
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

const (
	// chainIDB is the story ID for the second node in the 4-deep chain.
	chainIDB = "chain_b"
	// chainIDC is the story ID for the third node in the 4-deep chain.
	chainIDC = "chain_c"
	// writeFilePerm is the file permission used by the writeFile helper.
	writeFilePerm = 0o600
	// emptyCacheDir is the zero-value cache directory (caching disabled).
	emptyCacheDir = ""
	// emptyOCPPVersion is the zero-value OCPP version string.
	emptyOCPPVersion = ""
	// zeroMaxParallel means unlimited parallelism (runner chooses).
	zeroMaxParallel = 0
	// zeroLockTimeout means no lock timeout.
	zeroLockTimeout = 0
	// zeroShardIndex is the default (no sharding).
	zeroShardIndex = 0
	// zeroShardTotal is the default (no sharding).
	zeroShardTotal = 0
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
// topological order, all stories pass, and all cache statuses are
// CacheBypassed.
func Test_runner_RunChain(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Write story files to temp directory.
	writeFile(t, filepath.Join(dir, "chain_a.story"), storyChainA)
	writeFile(t, filepath.Join(dir, "chain_b.story"), storyChainB)
	writeFile(t, filepath.Join(dir, "chain_c.story"), storyChainC)
	writeFile(t, filepath.Join(dir, "chain_d.story"), storyChainD)

	cfg := noopCfg(dir)

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

	assertAllPassedBypassed(t, result.Stories)
	assertChainOrder(t, result.Stories)
}

// assertAllPassedBypassed verifies every story passed with a bypassed cache.
func assertAllPassedBypassed(t *testing.T, stories []runner.StoryResult) {
	t.Helper()

	for _, story := range stories {
		if story.Status != runner.StatusPassed {
			t.Errorf(
				"story %q: want StatusPassed, got %s",
				story.TestID,
				story.Status,
			)
		}

		if story.CacheStatus != runner.CacheBypassed {
			t.Errorf(
				"story %q: want CacheBypassed, got %s",
				story.TestID,
				story.CacheStatus,
			)
		}
	}
}

// assertChainOrder verifies that the four chain stories appear in topological
// order and that the result slice is sorted by Order.
func assertChainOrder(t *testing.T, stories []runner.StoryResult) {
	t.Helper()

	orderByID := buildOrderMap(stories)

	requiredIDs := []string{"chain_a", chainIDB, chainIDC, "chain_d"}

	checkRequiredPresent(t, orderByID, requiredIDs)
	checkTopoOrder(t, orderByID)
	checkStoriesSorted(t, stories)
}

// buildOrderMap builds a testID → Order map from a stories slice.
func buildOrderMap(stories []runner.StoryResult) map[string]int {
	orderByID := make(map[string]int, len(stories))

	for _, story := range stories {
		orderByID[story.TestID] = story.Order
	}

	return orderByID
}

// checkRequiredPresent asserts that all ids are present in orderByID.
func checkRequiredPresent(
	t *testing.T,
	orderByID map[string]int,
	ids []string,
) {
	t.Helper()

	for _, id := range ids {
		if _, ok := orderByID[id]; !ok {
			t.Errorf("story %q missing from results", id)
		}
	}
}

// checkTopoOrder asserts chain_a < chain_b < chain_c < chain_d by Order.
func checkTopoOrder(t *testing.T, orderByID map[string]int) {
	t.Helper()

	type orderCheck struct {
		before string
		after  string
	}

	checks := []orderCheck{
		{before: "chain_a", after: chainIDB},
		{before: chainIDB, after: chainIDC},
		{before: chainIDC, after: "chain_d"},
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
}

// checkStoriesSorted asserts that stories is sorted ascending by Order field.
func checkStoriesSorted(t *testing.T, stories []runner.StoryResult) {
	t.Helper()

	byOrder := func(a, b runner.StoryResult) int {
		return cmp.Compare(a.Order, b.Order)
	}

	if !slices.IsSortedFunc(stories, byOrder) {
		t.Error("result.Stories is not sorted by Order field")
	}
}

// writeFile is a test helper that writes content to path, failing on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), writeFilePerm)
	if err != nil {
		t.Fatalf("writeFile(%q): %v", path, err)
	}
}

// noopCfg returns a runner.Config with NoCache true and all other fields at
// their zero values, pointed at storyDir.
func noopCfg(storyDir string) runner.Config {
	return runner.Config{
		StoryPaths:         []string{storyDir},
		MaxParallel:        zeroMaxParallel,
		LockTimeout:        zeroLockTimeout,
		NoWait:             false,
		ShardIndex:         zeroShardIndex,
		ShardTotal:         zeroShardTotal,
		CacheDir:           emptyCacheDir,
		NoCache:            true,
		NoTraceOnPass:      false,
		OCPPVersion:        emptyOCPPVersion,
		InsecureSkipVerify: false,
	}
}

// cachedCfg returns a runner.Config with caching enabled, pointed at storyDir
// and cacheDir, with all other fields at their zero values.
func cachedCfg(storyDir, cacheDir string) runner.Config {
	return runner.Config{
		StoryPaths:         []string{storyDir},
		MaxParallel:        zeroMaxParallel,
		LockTimeout:        zeroLockTimeout,
		NoWait:             false,
		ShardIndex:         zeroShardIndex,
		ShardTotal:         zeroShardTotal,
		CacheDir:           cacheDir,
		NoCache:            false,
		NoTraceOnPass:      false,
		OCPPVersion:        emptyOCPPVersion,
		InsecureSkipVerify: false,
	}
}
