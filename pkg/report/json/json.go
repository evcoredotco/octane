// Package json implements the JSON emitter for OCTANE run reports.
// The public entry point is [WriteJSON], which projects a
// [runner.RunResult] into a byte-deterministic octane.json file.
//
// Task: T-007-20.
package json

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/octane-project/octane/pkg/report"
	"github.com/octane-project/octane/pkg/report/model"
	"github.com/octane-project/octane/pkg/runner"
)

// outputFileName is the name of the JSON report file written into the
// output directory.
const outputFileName = "octane.json"

// jsonReport is the top-level JSON serialization struct for the report.
type jsonReport struct {
	SchemaVersion int         `json:"schema_version"`
	OctaneVersion string      `json:"octane_version"`
	RunID         string      `json:"run_id"`
	StartedAt     string      `json:"started_at"`
	FinishedAt    string      `json:"finished_at"`
	Summary       jsonSummary `json:"summary"`
	Stories       []jsonStory `json:"stories"`
}

// jsonSummary is the JSON serialization struct for the run summary.
type jsonSummary struct {
	Total     int `json:"total"`
	Passed    int `json:"passed"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
	CacheHits int `json:"cache_hits"`
}

// jsonStory is the JSON serialization struct for a single story result.
type jsonStory struct {
	TestID       string        `json:"test_id"`
	ScopeKey     string        `json:"scope_key"`
	OCPPVersion  string        `json:"ocpp_version"`
	Status       string        `json:"status"`
	CacheStatus  string        `json:"cache_status"`
	StartedAt    string        `json:"started_at"`
	FinishedAt   string        `json:"finished_at"`
	DurationMS   int64         `json:"duration_ms"`
	Findings     []jsonFinding `json:"findings"`
	TracePresent bool          `json:"trace_present"`
	Trace        *jsonTrace    `json:"trace,omitempty"`
	Cause        string        `json:"cause,omitempty"`
	CauseChain   []string      `json:"cause_chain,omitempty"`
}

// jsonFinding is the JSON serialization struct for a diagnostic finding.
type jsonFinding struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// rfc3339 is the time layout used throughout the JSON report.
const rfc3339 = "2006-01-02T15:04:05Z07:00"

// WriteJSON projects result into a model.Report, serializes it to
// JSON with 2-space indentation, and writes the result to
// dir/octane.json. The directory is created with os.MkdirAll when it
// does not already exist.
func WriteJSON(
	result *runner.RunResult,
	dir string,
	opts report.JSONOptions,
) error {
	rep := model.FromRunner(result, opts.OctaneVersion)

	jrep := buildJSONReport(rep, opts)

	data, err := json.MarshalIndent(jrep, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	outPath := filepath.Join(dir, outputFileName)

	return os.WriteFile(outPath, data, 0o600)
}

// buildJSONReport converts a model.Report to the jsonReport
// serialization struct.
func buildJSONReport(rep *model.Report, opts report.JSONOptions) jsonReport {
	stories := buildJSONStories(rep.Stories, opts)

	return jsonReport{
		SchemaVersion: rep.SchemaVersion,
		OctaneVersion: rep.OctaneVersion,
		RunID:         rep.RunID,
		StartedAt:     rep.StartedAt.Format(rfc3339),
		FinishedAt:    rep.FinishedAt.Format(rfc3339),
		Summary:       buildJSONSummary(rep.Summary),
		Stories:       stories,
	}
}

// buildJSONSummary converts a model.Summary to a jsonSummary.
func buildJSONSummary(src model.Summary) jsonSummary {
	return jsonSummary{
		Total:     src.Total,
		Passed:    src.Passed,
		Failed:    src.Failed,
		Skipped:   src.Skipped,
		CacheHits: src.CacheHits,
	}
}

// buildJSONStories converts a slice of model.StoryReport to a slice of
// jsonStory, applying trace suppression and sorting findings.
func buildJSONStories(
	stories []model.StoryReport,
	opts report.JSONOptions,
) []jsonStory {
	out := make([]jsonStory, len(stories))

	for idx, sr := range stories {
		out[idx] = buildJSONStory(sr, opts)
	}

	return out
}

// buildJSONStory converts a single model.StoryReport to a jsonStory.
func buildJSONStory(
	src model.StoryReport,
	opts report.JSONOptions,
) jsonStory {
	tracePresent, trace := buildTrace(src.Trace, src.Status, opts)
	findings := buildJSONFindings(src.Findings)
	causeChain := walkCauseChain(src)

	return jsonStory{
		TestID:       src.TestID,
		ScopeKey:     src.ScopeKey,
		OCPPVersion:  src.OCPPVersion,
		Status:       src.Status,
		CacheStatus:  src.CacheStatus,
		StartedAt:    src.StartedAt.Format(rfc3339),
		FinishedAt:   src.FinishedAt.Format(rfc3339),
		DurationMS:   src.DurationMS,
		Findings:     findings,
		TracePresent: tracePresent,
		Trace:        trace,
		Cause:        src.Cause,
		CauseChain:   causeChain,
	}
}

// buildJSONFindings converts a slice of model.Finding to a slice of
// jsonFinding, sorted by (severity desc, message asc).
func buildJSONFindings(src []model.Finding) []jsonFinding {
	if len(src) == 0 {
		return nil
	}

	out := make([]jsonFinding, len(src))

	for idx, f := range src {
		out[idx] = jsonFinding{
			Message:  f.Message,
			Severity: f.Severity,
		}
	}

	sort.Slice(out, func(idx, jdx int) bool {
		if out[idx].Severity != out[jdx].Severity {
			// Higher severity (lexicographically larger) sorts first.
			return out[idx].Severity > out[jdx].Severity
		}

		return out[idx].Message < out[jdx].Message
	})

	return out
}
