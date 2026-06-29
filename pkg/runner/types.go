package runner

import "time"

// statusUnknown is the string label returned by String() methods when the
// enum value has not been recognised (i.e. falls into the default branch).
const statusUnknown = "unknown"

// ---------------------------------------------------------------------------
// Status enum — per-story execution outcome
// ---------------------------------------------------------------------------

// Status represents the execution outcome of a single story.
// A story's status is one of passed, failed, or skipped. The
// zero value is invalid; all valid statuses use iota+1.
type Status int

const (
	// StatusPassed indicates the story executed and all
	// assertions succeeded.
	StatusPassed Status = iota + 1

	// StatusFailed indicates the story executed but at least
	// one assertion did not hold, or a runtime error occurred
	// during execution.
	StatusFailed

	// StatusSkipped indicates the story did not execute because
	// a prerequisite in its dependency chain failed. The
	// StoryResult.Cause field names the failing prerequisite.
	StatusSkipped
)

// String returns the canonical lowercase label for a Status
// value. The labels match the wire format used in result.json
// and operator-facing output.
func (s Status) String() string {
	switch s {
	case StatusPassed:
		return "passed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return statusUnknown
	}
}

// ---------------------------------------------------------------------------
// CacheStatus enum — cache interaction outcome per OQ2
// ---------------------------------------------------------------------------

// CacheStatus describes how the cache participated in a story's
// execution. The four values were resolved in spec 005 OQ2 and
// appear in the report's per-entry cache_status field.
type CacheStatus int

const (
	// CacheHitPass indicates the cache contained a valid,
	// non-expired entry whose recorded status was passed.
	// The story was not re-executed.
	CacheHitPass CacheStatus = iota + 1

	// CacheHitSkip indicates the cache contained a valid,
	// non-expired entry whose recorded status was skipped.
	// The story was not re-executed.
	CacheHitSkip

	// CacheMiss indicates no valid cache entry existed for
	// the story's cache key (either absent or expired). The
	// story was executed and the result was written to the
	// cache.
	CacheMiss

	// CacheBypassed indicates the cache was intentionally not
	// consulted for this story, for example because the
	// operator passed a --no-cache flag or the story's
	// metadata explicitly disables caching.
	CacheBypassed
)

// String returns the canonical hyphenated label for a
// CacheStatus value. The labels match the report's
// cache_status field format defined in spec 005 OQ2.
func (c CacheStatus) String() string {
	switch c {
	case CacheHitPass:
		return "hit-pass"
	case CacheHitSkip:
		return "hit-skip"
	case CacheMiss:
		return "miss"
	case CacheBypassed:
		return "bypassed"
	default:
		return statusUnknown
	}
}

// ---------------------------------------------------------------------------
// Config — runner configuration
// ---------------------------------------------------------------------------

// Config holds the parameters for a single runner.Run invocation.
// Both distribution surfaces (CLI and GitHub Action) construct a
// Config and pass it to runner.Run; no surface-specific code
// paths exist in the runner (constitution principle II).
type Config struct {
	// StoryPaths lists the filesystem paths to .story files
	// or directories containing .story files. The runner
	// recursively discovers .story files in directories.
	StoryPaths []string

	// MaxParallel is the maximum number of stories that may
	// execute concurrently. Zero means sequential execution
	// (no parallelism). See ADR 0019 for the concurrency
	// model.
	MaxParallel int

	// LockTimeout is the maximum duration to wait when
	// acquiring a per-cache-key flock. Defaults to 60s when
	// zero. See spec 005 G6.
	LockTimeout time.Duration

	// NoWait, when true, causes the runner to fail
	// immediately if a cache-key flock cannot be acquired
	// instead of blocking up to LockTimeout. See spec 005
	// G6.
	NoWait bool

	// ShardIndex is the zero-based index of this shard in a
	// sharded CI fan-out. Only stories whose
	// sha256(test_id)[:8] mod ShardTotal == ShardIndex are
	// executed. Ignored when ShardTotal is zero. See spec 005
	// OQ1.
	ShardIndex int

	// ShardTotal is the total number of shards. Zero disables
	// sharding. See spec 005 OQ1.
	ShardTotal int

	// CacheDir is the root directory for the content-addressed
	// cache tree. When empty, the runner uses
	// $XDG_CACHE_HOME/octane/cache/ (falling back to
	// $HOME/.cache/octane/cache/ on POSIX systems). See
	// ADR 0016.
	CacheDir string

	// NoCache, when true, bypasses the cache entirely. Every
	// story executes unconditionally and no cache entries are
	// read or written. The CacheStatus for all stories will
	// be CacheBypassed.
	NoCache bool

	// NoTraceOnPass, when true, suppresses wire trace
	// capture for stories that pass. This reduces cache
	// entry size and report verbosity for green runs.
	NoTraceOnPass bool

	// OCPPVersion restricts the run to stories declaring this
	// OCPP version. When empty, all versions are included.
	OCPPVersion string

	// InsecureSkipVerify disables TLS certificate verification for
	// WebSocket connections. When true the runner threads this flag
	// through to the wire transport and the report emitter injects
	// a banner-level finding (spec 006 AC7, constitution principle X).
	InsecureSkipVerify bool

	// CSMSEndpoint is the base WebSocket URL of the CSMS under test
	// (e.g. "ws://localhost:9210"). Lifecycle domain keywords append
	// "/" + stationHandle to construct per-station URLs. An empty string
	// means no endpoint is configured; lifecycle keywords will return
	// a descriptive error if they need it.
	CSMSEndpoint string

	// Parameters supplies runtime values for placeholders declared by
	// story Meta Parameters. The runner substitutes these into step text
	// before keyword resolution.
	Parameters map[string]string
}
