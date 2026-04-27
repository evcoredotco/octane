// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-51
// AC2: Cache hit on second run: both stories are served from cache on re-run.
package integration_test

import (
	"context"
	"testing"

	_ "github.com/octane-project/octane/pkg/keywords/primitive"
	"github.com/octane-project/octane/pkg/runner"
)

// storyPrereq is a simple passing story with no dependencies.
const storyPrereq = `Meta
    Name:      Cache prereq story
    Id:        cache_prereq
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: Prereq passes
    When  wait 0s
`

// storyDependent depends on cache_prereq.
const storyDependent = `Meta
    Name:      Cache dependent story
    Id:        cache_dependent
    Tags:      helper
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    cache_prereq
        scope: per-station

Scenario: Dependent passes
    When  wait 0s
`

// Test_runner_RunCacheHitOnSecondRun asserts that both stories are served from
// the cache (CacheHitPass) on the second identical run.
func Test_runner_RunCacheHitOnSecondRun(t *testing.T) {
	t.Parallel()

	storyDir := t.TempDir()
	cacheDir := t.TempDir()

	writeFile(t, storyDir+"/cache_prereq.story", storyPrereq)
	writeFile(t, storyDir+"/cache_dependent.story", storyDependent)

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    false,
		CacheDir:   cacheDir,
	}

	// First run: both stories are cache misses and get executed.
	firstResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first runner.Run: %v", err)
	}

	const expectedStoriesCount = 2
	if len(firstResult.Stories) != expectedStoriesCount {
		t.Fatalf(
			"first run: len(Stories): want %d, got %d",
			expectedStoriesCount,
			len(firstResult.Stories),
		)
	}

	for _, sr := range firstResult.Stories {
		// Invariant: first run must be a cache miss.
		if sr.CacheStatus != runner.CacheMiss {
			t.Errorf(
				"first run: story %q: want CacheMiss, got %s",
				sr.TestID,
				sr.CacheStatus,
			)
		}
	}

	// Second run with same config: both stories should be cache hits.
	secondResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second runner.Run: %v", err)
	}

	if len(secondResult.Stories) != expectedStoriesCount {
		t.Fatalf(
			"second run: len(Stories): want %d, got %d",
			expectedStoriesCount,
			len(secondResult.Stories),
		)
	}

	for _, sr := range secondResult.Stories {
		// Invariant: second run must be a cache hit (pass or skip).
		isHit := sr.CacheStatus == runner.CacheHitPass ||
			sr.CacheStatus == runner.CacheHitSkip
		if !isHit {
			t.Errorf(
				"second run: story %q: want CacheHit*, got %s",
				sr.TestID,
				sr.CacheStatus,
			)
		}
	}

	// Invariant: Summary.CacheHits must equal the number of stories.
	const expectedCacheHits = 2
	if secondResult.Summary.CacheHits != expectedCacheHits {
		t.Errorf(
			"second run: Summary.CacheHits: want %d, got %d",
			expectedCacheHits,
			secondResult.Summary.CacheHits,
		)
	}
}
