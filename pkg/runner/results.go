package runner

import "time"

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
