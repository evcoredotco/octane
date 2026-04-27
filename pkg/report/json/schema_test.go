// Task: T-007-24.
package json_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/octane-project/octane/pkg/report"
	reportjson "github.com/octane-project/octane/pkg/report/json"
	"github.com/octane-project/octane/pkg/runner"
)

// requiredTopLevelKeys are the JSON keys that must appear at the top
// level of every octane.json report.
var requiredTopLevelKeys = []string{
	"schema_version",
	"run_id",
	"started_at",
	"finished_at",
	"summary",
	"stories",
}

// buildMinimalResult creates a minimal runner.RunResult suitable for
// schema validation tests.
func buildMinimalResult() *runner.RunResult {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  now,
		FinishedAt: now.Add(5 * time.Second),
		Summary: runner.Summary{
			Total:     1,
			Passed:    1,
			Failed:    0,
			Skipped:   0,
			CacheHits: 0,
		},
		Stories: []runner.StoryResult{
			{
				Order:       0,
				TestID:      "tc_example",
				ScopeKey:    "CP01",
				OCPPVersion: "1.6",
				Status:      runner.StatusPassed,
				CacheStatus: runner.CacheMiss,
				StartedAt:   now,
				FinishedAt:  now.Add(5 * time.Second),
				Findings:    nil,
				Trace:       nil,
				Cause:       "",
				CauseChain:  nil,
			},
		},
	}
}

// Test_json_Schema verifies that the JSON output contains all required
// top-level keys and that their types are correct.
func Test_json_Schema(t *testing.T) {
	t.Parallel()

	result := buildMinimalResult()
	opts := report.JSONOptions{
		OctaneVersion: "0.1.0-test",
		NoTraceOnPass: false,
	}
	dir := t.TempDir()

	if err := reportjson.WriteJSON(result, dir, opts); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	outPath := filepath.Join(dir, "octane.json")

	data, err := os.ReadFile(outPath) //nolint:gosec // G304: t.TempDir path
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	var top map[string]any

	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range requiredTopLevelKeys {
		if _, present := top[key]; !present {
			t.Errorf("missing required top-level key: %q", key)
		}
	}

	assertSchemaVersion(t, top)
	assertSummaryShape(t, top)
	assertStoriesShape(t, top)
}

// assertSchemaVersion verifies that schema_version is a non-zero number.
func assertSchemaVersion(t *testing.T, top map[string]any) {
	t.Helper()

	val, present := top["schema_version"]
	if !present {
		return
	}

	num, isFloat := val.(float64)
	if !isFloat || num <= 0 {
		t.Errorf("schema_version: got %v, want positive integer", val)
	}
}

// assertSummaryShape verifies that summary is an object with the
// required keys.
func assertSummaryShape(t *testing.T, top map[string]any) {
	t.Helper()

	summaryVal, present := top["summary"]
	if !present {
		return
	}

	summaryMap, isMap := summaryVal.(map[string]any)
	if !isMap {
		t.Errorf("summary: expected object, got %T", summaryVal)

		return
	}

	for _, key := range []string{"total", "passed", "failed", "skipped", "cache_hits"} {
		if _, exists := summaryMap[key]; !exists {
			t.Errorf("summary: missing required key %q", key)
		}
	}
}

// assertStoriesShape verifies that stories is an array.
func assertStoriesShape(t *testing.T, top map[string]any) {
	t.Helper()

	storiesVal, present := top["stories"]
	if !present {
		return
	}

	if _, isSlice := storiesVal.([]any); !isSlice {
		t.Errorf("stories: expected array, got %T", storiesVal)
	}
}
