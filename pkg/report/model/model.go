// Package model defines the in-memory representation of an OCTANE run
// result. Both the JSON and Robot XML emitters operate on [Report] and
// its nested types. The struct is populated by [FromRunner] from a
// [runner.RunResult] and is otherwise read-only.
//
// Task: T-007-01.
package model

import "time"

// Report is the top-level in-memory representation of an OCTANE run
// result. Both the JSON and Robot XML emitters operate on this struct.
type Report struct {
	// SchemaVersion is the integer version of the report schema. The
	// current version is 1.
	SchemaVersion int

	// OctaneVersion is the version string of the OCTANE binary that
	// produced the report. Defaults to "dev" when not set.
	OctaneVersion string

	// RunID is the unique ULID identifier for this run.
	RunID string

	// StartedAt is the wall-clock time when the run began.
	StartedAt time.Time

	// FinishedAt is the wall-clock time when the run completed.
	FinishedAt time.Time

	// Summary holds the aggregate pass/fail/skip/cache-hit counts.
	Summary Summary

	// Stories holds the per-story reports sorted by (TestID, ScopeKey)
	// for byte-deterministic output.
	Stories []StoryReport
}

// Summary holds the aggregate counts for a completed run.
type Summary struct {
	// Total is the number of stories in the resolved dependency graph.
	Total int

	// Passed is the number of stories that passed.
	Passed int

	// Failed is the number of stories that failed.
	Failed int

	// Skipped is the number of stories that were skipped.
	Skipped int

	// CacheHits is the number of stories served from the cache.
	CacheHits int
}

// StoryReport captures the report data for a single story execution.
type StoryReport struct {
	// TestID is the stable snake_case identifier of the story.
	TestID string

	// ScopeKey identifies the execution scope instance.
	ScopeKey string

	// OCPPVersion is the OCPP version declared by the story.
	OCPPVersion string

	// Status is the execution outcome: "passed", "failed", or "skipped".
	Status string

	// CacheStatus describes cache participation: "hit-pass", "hit-skip",
	// "miss", or "bypassed".
	CacheStatus string

	// StartedAt is the wall-clock time when execution began.
	StartedAt time.Time

	// FinishedAt is the wall-clock time when execution completed.
	FinishedAt time.Time

	// DurationMS is the story execution duration in milliseconds.
	DurationMS int64

	// Findings holds the diagnostic messages produced during execution.
	Findings []Finding

	// Trace holds the captured wire-level I/O. Nil when not present.
	Trace *Trace

	// TracePresent is true when trace data was captured for this story.
	TracePresent bool

	// Cause names the prerequisite whose failure triggered this story's
	// skip. Empty when Status is not "skipped".
	Cause string

	// CauseChain is the transitive chain of prerequisite failures. Empty
	// when Status is not "skipped".
	CauseChain []string
}

// Finding is a single diagnostic message produced during story execution.
type Finding struct {
	// Message is the human-readable diagnostic text.
	Message string

	// Severity classifies the finding (e.g. "error", "warning", "info").
	Severity string
}

// Trace holds the raw wire-level I/O captured during a story's execution.
type Trace struct {
	// Frames contains the ordered sequence of OCPP-J frames exchanged
	// during execution.
	Frames []Frame
}

// Frame holds a single OCPP-J wire frame.
type Frame struct {
	// Raw is the raw OCPP-J JSON bytes for this frame.
	Raw []byte
}
