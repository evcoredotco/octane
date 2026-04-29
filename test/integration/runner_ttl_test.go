// Package integration_test — runner integration tests for Spec 005 Phase 6.
//
// Task: T-005-57
// AC10: Cache-TTL: the cache treats an entry with an expired TTL as a miss and
// causes the runner to re-execute the story.
package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/cache"
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/runner"
)

// storyTTL is a single passing story used for the TTL expiry test.
const storyTTL = `Meta
    Name:      TTL test story
    Id:        ttl_story
    Tags:      helper
    Stations:  1
    Timeout:   10s

Scenario: TTL story passes
    When  wait 0s
`

// valueTTLOne is the 1-second TTL used in the expiry assertions.
const valueTTLOne = 1 * time.Second

// valueTTLWait is the wait period (1100ms) that guarantees the 1s TTL has expired.
const valueTTLWait = 1100 * time.Millisecond

// Test_runner_CacheTTLDirectExpiry validates that a cache entry written with a
// 1-second TTL is treated as ErrCacheMiss after the TTL elapses. This test
// exercises the cache layer's TTL invalidation path directly.
func Test_runner_CacheTTLDirectExpiry(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()

	storyCache, err := cache.Open(cacheDir)
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}

	resultJSON, marshalErr := json.Marshal(struct {
		Status string `json:"status"`
	}{Status: "passed"})
	if marshalErr != nil {
		t.Fatalf("json.Marshal: %v", marshalErr)
	}

	key := cache.Key{
		TestID:          "ttl_story",
		ScopeKey:        "CP01",
		CSMSEndpointSHA: "00000000",
		OctaneVersion:   "dev",
		OCPPVersion:     "unknown",
		StoryContentSHA: "00000000",
		ParameterSHA:    "00000000",
	}

	entry := cache.Entry{
		Result:    resultJSON,
		WrittenAt: time.Now().UTC(),
		TTL:       valueTTLOne,
	}

	if err = storyCache.Put(context.Background(), key, entry); err != nil {
		t.Fatalf("cache.Put: %v", err)
	}

	// Invariant: entry is readable immediately (not yet expired).
	if _, getErr := storyCache.Get(context.Background(), key); getErr != nil {
		t.Fatalf("cache.Get before expiry: want hit, got %v", getErr)
	}

	// Wait for the TTL to expire.
	time.Sleep(valueTTLWait)

	// Invariant: after TTL expiry, Get must return ErrCacheMiss.
	_, getErrAfter := storyCache.Get(context.Background(), key)
	if !errors.Is(getErrAfter, cache.ErrCacheMiss) {
		t.Errorf(
			"cache.Get after TTL expiry: want ErrCacheMiss, got %v",
			getErrAfter,
		)
	}
}

// Test_runner_CacheTTLRunnerSecondRunIsHit validates the positive case:
// when caching is enabled and no TTL expiry occurs between two runs,
// the second run is served from the cache.
func Test_runner_CacheTTLRunnerSecondRunIsHit(t *testing.T) {
	t.Parallel()

	storyDir := t.TempDir()
	cacheDir := t.TempDir()

	writeFile(t, storyDir+"/ttl_story.story", storyTTL)

	cfg := runner.Config{
		StoryPaths: []string{storyDir},
		NoCache:    false,
		CacheDir:   cacheDir,
	}

	// First run: story executes, result is cached (CacheMiss → write).
	firstResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first runner.Run: %v", err)
	}

	if len(firstResult.Stories) != 1 {
		t.Fatalf("first run: want 1 story, got %d", len(firstResult.Stories))
	}

	// Invariant: first run must be a cache miss.
	if firstResult.Stories[0].CacheStatus != runner.CacheMiss {
		t.Errorf(
			"first run: want CacheMiss, got %s",
			firstResult.Stories[0].CacheStatus,
		)
	}

	// Second run immediately: must be a cache hit since no TTL has been set.
	secondResult, err := runner.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second runner.Run: %v", err)
	}

	if len(secondResult.Stories) != 1 {
		t.Fatalf("second run: want 1 story, got %d", len(secondResult.Stories))
	}

	// Invariant: second run (no TTL expiry) must be a cache hit.
	isHit := secondResult.Stories[0].CacheStatus == runner.CacheHitPass ||
		secondResult.Stories[0].CacheStatus == runner.CacheHitSkip
	if !isHit {
		t.Errorf(
			"second run: want CacheHit*, got %s",
			secondResult.Stories[0].CacheStatus,
		)
	}

	// Invariant: Summary.CacheHits must be 1.
	const expectedCacheHits = 1
	if secondResult.Summary.CacheHits != expectedCacheHits {
		t.Errorf(
			"second run: Summary.CacheHits: want %d, got %d",
			expectedCacheHits,
			secondResult.Summary.CacheHits,
		)
	}
}
