// Package cache defines the types and interface for OCTANE's
// content-addressed file tree cache per ADR 0016.
//
// The cache stores test results as plain JSON files in a directory
// tree keyed by the SHA-256 of the cache key tuple. It has no
// third-party dependencies (constitution principle V).
//
// # Security
//
// Cache entries MUST NOT contain credentials. The runner MUST
// redact sensitive fields (auth headers, idTags marked sensitive
// in connection profiles) before writing any entry. This
// constraint is mandated by constitution principle X. Callers
// are responsible for ensuring that the Result and Trace byte
// slices passed to [Cache.Put] are already redacted.
//
// # Cache key
//
// A [Key] captures every input that affects a test's outcome.
// Its [Key.Hash] method returns the SHA-256 hex digest of the
// lexicographically ordered, colon-joined tuple of those inputs.
// Any change to any field produces a different hash and therefore
// a different filesystem path; old entries become unreachable and
// are pruned by age.
//
// # Cache entry
//
// An [Entry] holds the serialised result, the optional wire
// trace, and timing/TTL metadata. The [Entry.IsExpired] method
// implements the TTL invalidation described in spec 005 AC10.
//
// # Interface
//
// [Cache] is the minimal surface consumed by the runner. A file
// tree implementation satisfies it (see T-005-22); tests can
// supply an in-memory double.
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned by [Cache.Get] when no valid entry
// exists for the requested key. Callers should use [errors.Is] to
// detect this condition:
//
//	entry, err := c.Get(ctx, key)
//	if errors.Is(err, cache.ErrCacheMiss) {
//	    // execute the test and write back via Put
//	}
var ErrCacheMiss = errors.New("cache: miss")

// Key is the cache key tuple. Every field that can affect a
// test's outcome is included so that any input change produces
// a distinct hash and therefore a distinct filesystem path.
//
// The fields correspond to ADR 0016 "Cache key derivation":
//
//   - TestID:          story Id Meta key.
//   - ScopeKey:        station handle (per-station), run ID
//     (per-run), or empty string (global) per
//     ADR 0015.
//   - CSMSEndpointSHA: SHA-256 hex digest of (URL +
//     subprotocol + auth-mode tuple).
//   - OctaneVersion:   build version string.
//   - OCPPVersion:     OCPP version from story Meta or
//     config (e.g., "1.6", "2.0.1").
//   - StoryContentSHA: SHA-256 hex digest of the story
//     file content plus the transitive
//     content of all prerequisite stories.
//   - ParameterSHA:    SHA-256 hex digest of the bound
//     parameter values.
//
// Callers populate these fields; [Key.Hash] deterministically
// derives the content-address from them.
type Key struct {
	// TestID is the stable identifier of the story under test.
	// It must match the Id Meta key in the .story file.
	TestID string

	// ScopeKey qualifies the test execution context. For
	// per-station scope it is the station handle (e.g.,
	// "CP01"); for per-run scope it is the run ID; for global
	// scope it is the empty string.
	ScopeKey string

	// CSMSEndpointSHA is the SHA-256 hex digest of the CSMS
	// connection tuple (URL, subprotocol, auth-mode). A change
	// in the CSMS configuration invalidates all cached results
	// that were produced against the previous endpoint.
	CSMSEndpointSHA string

	// OctaneVersion is the build version of the octane binary
	// that produced or is consuming this cache entry. Including
	// the version in the key ensures that a new octane release
	// does not silently reuse results from an older version
	// whose behaviour may differ.
	OctaneVersion string

	// OCPPVersion is the OCPP protocol version declared by the
	// story or the run configuration (e.g., "1.6", "2.0.1",
	// "2.1").
	OCPPVersion string

	// StoryContentSHA is the SHA-256 hex digest of the story
	// file's content concatenated with the transitive content
	// of all prerequisite stories, in topological order. A
	// change in any story in the dependency chain invalidates
	// the cache entry.
	StoryContentSHA string

	// ParameterSHA is the SHA-256 hex digest of the bound
	// parameter values. Different parameter bindings produce
	// different cache keys even for the same story.
	ParameterSHA string
}

// Entry is a single cache entry comprising the serialised test
// result, the optional wire trace, and timing metadata.
//
// Result and Trace are the raw JSON bytes of result.json and
// trace.json respectively (per ADR 0016). They MUST NOT contain
// credentials or other sensitive material (constitution
// principle X).
type Entry struct {
	// Result is the JSON-encoded content of result.json.
	// It is always present for a valid cache entry.
	Result []byte

	// Trace is the JSON-encoded content of trace.json.
	// It may be nil when the trace was suppressed (e.g.,
	// --no-trace-on-pass for a passing test).
	Trace []byte

	// WrittenAt is the wall-clock time at which this entry
	// was persisted to disk. It is used together with TTL to
	// determine expiry.
	WrittenAt time.Time

	// TTL is the maximum age of this entry before it is
	// considered stale. A zero value means the entry never
	// expires by TTL (it can still be pruned by max-age).
	//
	// When a story declares a Cache-TTL Meta key (spec 005
	// AC10), the runner sets this field to the parsed duration.
	TTL time.Duration
}

// IsExpired reports whether the entry has exceeded its TTL as of
// the given wall-clock time. An entry with a zero TTL never
// expires by this check.
//
// This implements the TTL invalidation described in spec 005
// AC10: a Cache-TTL Meta key on a helper story causes the runner
// to treat expired entries as cache misses and re-execute the
// story.
func (e *Entry) IsExpired(now time.Time) bool {
	if e.TTL <= 0 {
		return false
	}

	return now.After(e.WrittenAt.Add(e.TTL))
}

// Cache is the interface consumed by the runner to look up,
// store, and prune cached test results. The file tree
// implementation (T-005-22) satisfies this interface; tests may
// supply an in-memory double.
//
// All methods accept a [context.Context] for cancellation and
// timeout propagation.
type Cache interface {
	// Get retrieves the cache entry for the given key. It
	// returns [ErrCacheMiss] if no valid entry exists (the
	// key was never written, the file is corrupt, or the
	// entry's TTL has expired).
	//
	// A successful return means the entry was found, its
	// schema version is supported, and it has not expired.
	// The caller may still inspect [Entry.IsExpired] with a
	// different clock if needed.
	Get(ctx context.Context, key Key) (*Entry, error)

	// Put writes a cache entry for the given key using the
	// atomic temp-file-and-rename protocol described in
	// ADR 0016. The entry MUST NOT contain credentials or
	// other sensitive material (constitution principle X).
	//
	// If an entry already exists for the key, it is
	// overwritten atomically.
	Put(ctx context.Context, key Key, entry Entry) error

	// Prune removes cache entries whose WrittenAt time is
	// older than maxAge, and entries whose TTL has expired.
	// Empty fanout directories are removed after pruning.
	//
	// Prune is safe to call concurrently with Get and Put;
	// the atomic-rename protocol ensures that readers never
	// observe a partially-deleted entry.
	Prune(ctx context.Context, maxAge time.Duration) error
}
