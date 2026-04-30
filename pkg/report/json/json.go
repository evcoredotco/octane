// Package reportjson implements the JSON emitter for OCTANE run reports.
// The public entry point is [WriteJSON], which projects a
// [runner.RunResult] into a byte-deterministic octane.json file.
//
// Task: T-007-20.
package reportjson

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	"github.com/evcoreco/octane/pkg/report/model"
	"github.com/evcoreco/octane/pkg/runner"
)

const (
	// outputFileName is the name of the JSON report file written into the
	// output directory.
	outputFileName = "octane.json"

	// dirPerm is the permission bits for the output directory created by
	// WriteJSON.
	dirPerm = 0o700

	// filePerm is the permission bits for the JSON report file written by
	// WriteJSON.
	filePerm = 0o600

	// emptyLen is the sentinel zero used in len() == 0 guards.
	emptyLen = 0
)

// jsonReport is the top-level JSON serialization struct for the report.
type jsonReport struct {
	SchemaVersion int         `json:"schemaVersion"`
	OctaneVersion string      `json:"octaneVersion"`
	RunID         string      `json:"runId"`
	StartedAt     string      `json:"startedAt"`
	FinishedAt    string      `json:"finishedAt"`
	Summary       jsonSummary `json:"summary"`
	Stories       []jsonStory `json:"stories"`
}

// jsonSummary is the JSON serialization struct for the run summary.
type jsonSummary struct {
	Total     int `json:"total"`
	Passed    int `json:"passed"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
	CacheHits int `json:"cacheHits"`
}

// jsonStory is the JSON serialization struct for a single story result.
type jsonStory struct {
	TestID       string        `json:"testId"`
	ScopeKey     string        `json:"scopeKey"`
	OCPPVersion  string        `json:"ocppVersion"`
	Status       string        `json:"status"`
	CacheStatus  string        `json:"cacheStatus"`
	StartedAt    string        `json:"startedAt"`
	FinishedAt   string        `json:"finishedAt"`
	DurationMS   int64         `json:"durationMs"`
	Findings     []jsonFinding `json:"findings"`
	TracePresent bool          `json:"tracePresent"`
	Trace        *jsonTrace    `json:"trace,omitempty"`
	Cause        string        `json:"cause,omitempty"`
	CauseChain   []string      `json:"causeChain,omitempty"`
}

// jsonFinding is the JSON serialization struct for a diagnostic finding.
type jsonFinding struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// rfc3339 is the time layout used throughout the JSON report.
const rfc3339 = time.RFC3339

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
		return fmt.Errorf("report: marshal JSON: %w", err)
	}

	err = os.MkdirAll(dir, dirPerm)
	if err != nil {
		return fmt.Errorf("report: create output dir: %w", err)
	}

	outPath := filepath.Join(dir, outputFileName)

	err = os.WriteFile(outPath, data, filePerm)
	if err != nil {
		return fmt.Errorf("report: write output file: %w", err)
	}

	return nil
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
	if len(src) == emptyLen {
		return nil
	}

	out := make([]jsonFinding, len(src))

	for idx, f := range src {
		out[idx] = jsonFinding{
			Message:  f.Message,
			Severity: f.Severity,
		}
	}

	slices.SortFunc(out, func(a, b jsonFinding) int {
		// Higher severity (lexicographically larger) sorts first, then
		// sort by message for determinism within the same severity.
		return cmp.Or(
			cmp.Compare(b.Severity, a.Severity),
			cmp.Compare(a.Message, b.Message),
		)
	})

	return out
}
