// Package runner orchestrates the execution of .story files by
// walking the test dependency graph (ADR 0015), consulting the
// content-addressed cache (ADR 0016), and returning structured
// results that the report emitter (spec 007) consumes.
//
// This file defines the public data types: runner configuration,
// per-story and aggregate results, execution status, and cache
// status. It contains no implementation logic and imports only
// the standard library.

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
// Finding — a single diagnostic produced during story execution
// ---------------------------------------------------------------------------

// Finding is a single diagnostic message produced during story
// execution. Findings capture assertion failures, runtime errors,
// and informational notes. They appear in the run report under
// the story that produced them.
type Finding struct {
	// Message is the human-readable diagnostic text. For
	// assertion failures it describes the mismatch; for
	// skipped stories it names the failing prerequisite.
	Message string

	// Severity classifies the finding. Typical values are
	// "error" (assertion failure), "warning" (non-fatal
	// observation), and "info" (diagnostic note). The set is
	// not enumerated here because the report emitter owns
	// presentation; the runner treats it as an opaque label.
	Severity string
}

// ---------------------------------------------------------------------------
// Trace — captured wire-level I/O for a single story execution
// ---------------------------------------------------------------------------

// Trace holds the raw wire-level I/O captured during a story's
// execution. It is serialized to trace.json in the cache entry.
// When the --no-trace-on-pass flag is active and the story
// passes, Trace is nil in the StoryResult.
type Trace struct {
	// Frames contains the ordered sequence of OCPP-J frames
	// exchanged during execution, encoded as raw JSON bytes.
	// Each entry is one complete JSON array as it appeared on
	// the wire. The slice preserves send/receive ordering.
	Frames [][]byte
}

// ---------------------------------------------------------------------------
// StoryResult — per-story outcome
// ---------------------------------------------------------------------------

// StoryResult captures the outcome of a single story execution.
// The runner produces one StoryResult per story in the resolved
// dependency graph. Results are sorted by the Order field in the
// enclosing RunResult.Stories slice, reflecting the stable
// topological execution order (spec 005 S10, constitution
// principle IV).
type StoryResult struct {
	// Order is the zero-based position of this story in the
	// stable topological execution sequence. Ties within a
	// topological level are broken by lexicographic story ID.
	Order int

	// TestID is the stable snake_case identifier from the
	// story's Meta.ID field.
	TestID string

	// ScopeKey identifies the execution scope instance. For
	// per-station scope this is the station handle (e.g.
	// "CP01"); for per-run scope it is the run ID; for global
	// scope it is the empty string.
	ScopeKey string

	// OCPPVersion is the OCPP version declared by the story
	// (e.g. "1.6").
	OCPPVersion string

	// Status is the execution outcome: passed, failed, or
	// skipped.
	Status Status

	// CacheStatus describes whether the cache was consulted
	// and whether it held a valid entry. See the CacheStatus
	// enum for the four possible values.
	CacheStatus CacheStatus

	// StartedAt is the wall-clock time when execution of this
	// story began (or when the cache lookup started, for
	// cache hits).
	StartedAt time.Time

	// FinishedAt is the wall-clock time when execution of
	// this story completed (or when the cache lookup
	// completed, for cache hits).
	FinishedAt time.Time

	// Findings holds the diagnostic messages produced during
	// execution. For a skipped story, it contains a single
	// finding referencing the failing prerequisite (AC4).
	Findings []Finding

	// Trace holds the captured wire-level I/O. It is nil when
	// --no-trace-on-pass is active and the story passed, or
	// when the result was served from the cache.
	Trace *Trace

	// Cause names the prerequisite whose failure triggered
	// this story's skip. The format is "test_id/scope_key" of
	// the failing prerequisite. Empty when Status is not
	// StatusSkipped.
	Cause string

	// CauseChain is the transitive chain of prerequisite
	// failures leading to this story's skip, ordered from the
	// immediate parent to the root cause. Empty when Status
	// is not StatusSkipped. The chain is recoverable from the
	// report by walking Cause fields, but is materialized
	// here for convenience.
	CauseChain []string
}

// ---------------------------------------------------------------------------
// Summary — aggregate pass/fail/skip counts
// ---------------------------------------------------------------------------

// Summary holds the aggregate counts for a completed run. It is
// embedded in RunResult and provides the data for the one-line
// summary the CLI prints at the end of a run.
type Summary struct {
	// Total is the number of stories in the resolved
	// dependency graph.
	Total int

	// Passed is the number of stories whose Status is
	// StatusPassed.
	Passed int

	// Failed is the number of stories whose Status is
	// StatusFailed.
	Failed int

	// Skipped is the number of stories whose Status is
	// StatusSkipped.
	Skipped int

	// CacheHits is the number of stories served from the
	// cache (CacheStatus is CacheHitPass or CacheHitSkip).
	CacheHits int
}

// ---------------------------------------------------------------------------
// RunResult — top-level result returned by runner.Run
// ---------------------------------------------------------------------------

// RunResult is the top-level result returned by runner.Run. It
// aggregates per-story outcomes and timing information. The
// report emitter (spec 007) consumes this struct to produce
// operator-facing output.
type RunResult struct {
	// RunID is a unique identifier for this run, generated as
	// a ULID at the start of execution.
	RunID string

	// StartedAt is the wall-clock time when the run began.
	StartedAt time.Time

	// FinishedAt is the wall-clock time when the run
	// completed (all stories finished or skipped).
	FinishedAt time.Time

	// Stories holds the per-story results sorted by the
	// Order field, reflecting the stable topological
	// execution order.
	Stories []StoryResult

	// Summary holds the aggregate pass/fail/skip/cache-hit
	// counts.
	Summary Summary
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
}
