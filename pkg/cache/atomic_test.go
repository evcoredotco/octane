// Package cache_test contains black-box contract tests for pkg/cache.
//
// These tests verify the atomic write protocol (spec 005 AC2, AC8, AC10)
// and concurrent safety of the file-tree cache implementation.
package cache_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/cache"
)

// testKey returns a fully-populated Key whose Hash is deterministic
// for the given testID. All non-ID fields are fixed synthetic values.
func testKey(testID string) cache.Key {
	return cache.Key{
		TestID:   testID,
		ScopeKey: "test-scope",
		// 64-character hex strings (valid SHA-256 placeholders).
		CSMSEndpointSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		OctaneVersion:   "v0.0.0-test",
		OCPPVersion:     "1.6",
		StoryContentSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ParameterSHA:    "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}
}

// Test_cache_RoundTrip verifies that Put followed by Get returns the
// same Result bytes for the same key (happy-path contract).
//
// Invariant: every byte written by Put is returned unchanged by Get.
func Test_cache_RoundTrip(t *testing.T) {
	t.Parallel()

	const valueResult = `{"status":"passed"}`

	cch, err := cache.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ctx := context.Background()
	key := testKey("round-trip-001")

	entry := cache.Entry{
		Result:    []byte(valueResult),
		Trace:     nil,
		WrittenAt: time.Now().UTC(),
		TTL:       0,
	}

	if err = cch.Put(ctx, key, entry); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, err := cch.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if string(got.Result) != valueResult {
		t.Errorf(
			"result mismatch: got %q, want %q",
			got.Result,
			valueResult,
		)
	}
}

// Test_cache_TornWriteRejected verifies that a stray result.json.tmp
// file (simulating an interrupted Put whose rename never completed) is
// never surfaced by Get as a valid entry.
//
// Invariant: Get returns ErrCacheMiss when only the .tmp artefact
// exists; the canonical result.json has not been written.
func Test_cache_TornWriteRejected(t *testing.T) {
	t.Parallel()

	// Use a named variable so we can derive the entry directory path
	// manually — the Cache interface does not expose internal paths.
	tmpDir := t.TempDir()

	cch, err := cache.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ctx := context.Background()
	key := testKey("torn-write-001")
	hash := key.Hash()

	// Build the path that FileCache uses: <dir>/results/<hash[:2]>/<hash>/
	entryDir := filepath.Join(tmpDir, "results", hash[:2], hash)
	if err = os.MkdirAll(entryDir, 0o750); err != nil {
		t.Fatalf("create entry dir: %v", err)
	}

	// Drop the intermediate temp file; the atomic rename never happens.
	tmpPath := filepath.Join(entryDir, "result.json.tmp")
	if err = os.WriteFile(
		tmpPath,
		[]byte(`{"schema_version":1}`),
		0o600,
	); err != nil {
		t.Fatalf("write .tmp artefact: %v", err)
	}

	// Get must not treat the .tmp file as a valid cache entry.
	_, err = cch.Get(ctx, key)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss for torn write, got: %v", err)
	}
}

// Test_cache_TTLExpiry verifies that an entry whose TTL has elapsed is
// treated as a cache miss by Get.
//
// The entry is written with WrittenAt set 2 seconds in the past and a
// TTL of 1 second, so it is already expired by the time Get is called.
// This avoids any real sleep while exercising the TTL invalidation path
// (spec 005 AC10).
//
// Note: Put stores TTL as whole seconds (int64(entry.TTL.Seconds())),
// so the minimum effective TTL is 1 second.
//
// Invariant: Get returns ErrCacheMiss when WrittenAt + TTL < now.
func Test_cache_TTLExpiry(t *testing.T) {
	t.Parallel()

	// TTL of exactly 1 second is the minimum that survives the
	// seconds-granularity serialisation in Put.
	const (
		valueTTL       = time.Second
		valueAgeOffset = 2 * time.Second // WrittenAt is 2s ago, TTL is 1s
	)

	cch, err := cache.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ctx := context.Background()
	key := testKey("ttl-expiry-001")

	// WrittenAt 2s ago with TTL 1s: entry.IsExpired(now) == true.
	entry := cache.Entry{
		Result:    []byte(`{"status":"passed"}`),
		Trace:     nil,
		WrittenAt: time.Now().UTC().Add(-valueAgeOffset),
		TTL:       valueTTL,
	}

	if err = cch.Put(ctx, key, entry); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, err = cch.Get(ctx, key)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf(
			"expected ErrCacheMiss for expired TTL entry, got: %v",
			err,
		)
	}
}

// Test_cache_PruneRemovesOldEntries verifies that Prune removes an entry
// whose WrittenAt predates now by more than maxAge, and that a subsequent
// Get returns ErrCacheMiss.
//
// Invariant: an entry older than maxAge is not retrievable after Prune.
func Test_cache_PruneRemovesOldEntries(t *testing.T) {
	t.Parallel()

	// maxAge of 1s; entry is written 2s in the past to exceed it.
	const (
		valueMaxAge    = time.Second
		valueAgeOffset = 2 * time.Second
	)

	cch, err := cache.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ctx := context.Background()
	key := testKey("prune-old-001")

	entry := cache.Entry{
		Result:    []byte(`{"status":"passed"}`),
		Trace:     nil,
		WrittenAt: time.Now().UTC().Add(-valueAgeOffset),
		TTL:       0,
	}

	if err = cch.Put(ctx, key, entry); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Sanity-check: entry is reachable before pruning.
	if _, err = cch.Get(ctx, key); err != nil {
		t.Fatalf("Get before Prune: %v", err)
	}

	if err = cch.Prune(ctx, valueMaxAge); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	_, err = cch.Get(ctx, key)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss after Prune, got: %v", err)
	}
}

// Test_cache_ConcurrentPutDifferentKeys verifies that two goroutines
// writing disjoint cache keys simultaneously produce no data races and
// that both entries are independently retrievable afterwards.
//
// Invariant: concurrent Put on disjoint keys is race-free and each
// entry survives intact.
func Test_cache_ConcurrentPutDifferentKeys(t *testing.T) {
	t.Parallel()

	cch, err := cache.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ctx := context.Background()

	const (
		valueResultA = `{"status":"passed","id":"key-a"}`
		valueResultB = `{"status":"passed","id":"key-b"}`
	)

	keyA := testKey("concurrent-put-key-a")
	keyB := testKey("concurrent-put-key-b")

	entryA := cache.Entry{
		Result:    []byte(valueResultA),
		Trace:     nil,
		WrittenAt: time.Now().UTC(),
		TTL:       0,
	}

	entryB := cache.Entry{
		Result:    []byte(valueResultB),
		Trace:     nil,
		WrittenAt: time.Now().UTC(),
		TTL:       0,
	}

	var (
		waitGroup sync.WaitGroup
		errA      error
		errB      error
	)

	const valueConcurrency = 2

	waitGroup.Add(valueConcurrency)

	go func() {
		defer waitGroup.Done()

		errA = cch.Put(ctx, keyA, entryA)
	}()

	go func() {
		defer waitGroup.Done()

		errB = cch.Put(ctx, keyB, entryB)
	}()

	waitGroup.Wait()

	if errA != nil {
		t.Fatalf("Put keyA: %v", errA)
	}

	if errB != nil {
		t.Fatalf("Put keyB: %v", errB)
	}

	gotA, err := cch.Get(ctx, keyA)
	if err != nil {
		t.Fatalf("Get keyA: %v", err)
	}

	if string(gotA.Result) != valueResultA {
		t.Errorf(
			"keyA mismatch: got %q, want %q",
			gotA.Result,
			valueResultA,
		)
	}

	gotB, err := cch.Get(ctx, keyB)
	if err != nil {
		t.Fatalf("Get keyB: %v", err)
	}

	if string(gotB.Result) != valueResultB {
		t.Errorf(
			"keyB mismatch: got %q, want %q",
			gotB.Result,
			valueResultB,
		)
	}
}
