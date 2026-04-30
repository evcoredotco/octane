// from_runner.go implements the projection from runner.RunResult to model.Report.
//
// Task: T-007-03.

package model

import (
	"cmp"
	"slices"

	"github.com/evcoreco/octane/pkg/report/internal/redact"
	"github.com/evcoreco/octane/pkg/runner"
)

// schemaVersion is the integer version of the report schema.
const schemaVersion = 1

// defaultOctaneVersion is used when the caller does not supply a version.
const defaultOctaneVersion = "dev"

// noFindings is the length of an empty findings slice.
const noFindings = 0

// FromRunner projects a runner.RunResult into a model.Report.
// Stories are sorted by (TestID, ScopeKey) for byte-deterministic output.
// The octaneVersion parameter is embedded in the report header; when empty
// it defaults to "dev".
func FromRunner(result *runner.RunResult, octaneVersion string) *Report {
	ver := octaneVersion
	if ver == "" {
		ver = defaultOctaneVersion
	}

	stories := projectStories(result.Stories)

	slices.SortFunc(stories, func(a, b StoryReport) int {
		if n := cmp.Compare(a.TestID, b.TestID); n != 0 {
			return n
		}

		return cmp.Compare(a.ScopeKey, b.ScopeKey)
	})

	return &Report{
		SchemaVersion: schemaVersion,
		OctaneVersion: ver,
		RunID:         result.RunID,
		StartedAt:     result.StartedAt,
		FinishedAt:    result.FinishedAt,
		Summary:       projectSummary(result.Summary),
		Stories:       stories,
	}
}

// projectSummary converts a runner.Summary to a model.Summary.
func projectSummary(src runner.Summary) Summary {
	return Summary{
		Total:     src.Total,
		Passed:    src.Passed,
		Failed:    src.Failed,
		Skipped:   src.Skipped,
		CacheHits: src.CacheHits,
	}
}

// projectStories converts a slice of runner.StoryResult to a slice of
// model.StoryReport.
func projectStories(src []runner.StoryResult) []StoryReport {
	out := make([]StoryReport, len(src))

	for idx, sr := range src {
		out[idx] = projectStory(sr)
	}

	return out
}

// projectStory converts a single runner.StoryResult to a model.StoryReport.
func projectStory(src runner.StoryResult) StoryReport {
	dur := src.FinishedAt.Sub(src.StartedAt).Milliseconds()
	tracePresent := src.Trace != nil

	return StoryReport{
		TestID:       src.TestID,
		ScopeKey:     src.ScopeKey,
		OCPPVersion:  src.OCPPVersion,
		Status:       projectStatus(src.Status),
		CacheStatus:  projectCacheStatus(src.CacheStatus),
		StartedAt:    src.StartedAt,
		FinishedAt:   src.FinishedAt,
		DurationMS:   dur,
		Findings:     projectFindings(src.Findings),
		Trace:        projectTrace(src.Trace),
		TracePresent: tracePresent,
		Cause:        src.Cause,
		CauseChain:   src.CauseChain,
	}
}

// projectStatus maps runner.Status to its canonical lowercase string.
func projectStatus(src runner.Status) string {
	switch src {
	case runner.StatusPassed:
		return "passed"
	case runner.StatusFailed:
		return "failed"
	case runner.StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// projectCacheStatus maps runner.CacheStatus to its canonical hyphenated
// string.
func projectCacheStatus(src runner.CacheStatus) string {
	switch src {
	case runner.CacheHitPass:
		return "hit-pass"
	case runner.CacheHitSkip:
		return "hit-skip"
	case runner.CacheMiss:
		return "miss"
	case runner.CacheBypassed:
		return "bypassed"
	default:
		return "unknown"
	}
}

// projectFindings converts a slice of runner.Finding to a slice of
// model.Finding. Each finding message is scrubbed for JWT patterns and
// other credential-bearing strings before inclusion in the report.
func projectFindings(src []runner.Finding) []Finding {
	if len(src) == noFindings {
		return nil
	}

	out := make([]Finding, len(src))

	for idx, f := range src {
		out[idx] = Finding{
			Message:  redact.FindingMessage(f.Message),
			Severity: f.Severity,
		}
	}

	return out
}

// projectTrace converts a *runner.Trace to a *model.Trace. Returns nil when
// src is nil. Each raw OCPP-J frame is passed through the frame redactor
// before inclusion in the report.
func projectTrace(src *runner.Trace) *Trace {
	if src == nil {
		return nil
	}

	frames := make([]Frame, len(src.Frames))

	for idx, raw := range src.Frames {
		frames[idx] = Frame{Raw: redact.Frame(raw)}
	}

	return &Trace{Frames: frames}
}
