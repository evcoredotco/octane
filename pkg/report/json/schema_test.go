// Task: T-007-24.

package reportjson_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	reportjson "github.com/evcoreco/octane/pkg/report/json"
	"github.com/evcoreco/octane/pkg/runner"
)

// schemaTestYear is the year used in buildMinimalResult time fixtures.
const schemaTestYear = 2024

// schemaTestMonth is the month used in time.Date for schema test fixtures.
const schemaTestMonth = 1

// schemaTestDay is the day-of-month used in time.Date for schema test fixtures.
const schemaTestDay = 15

// schemaTestHour is the hour used in time.Date for schema test fixtures.
const schemaTestHour = 12

// minimalTotalStories is the total story count for the minimal test result.
const minimalTotalStories = 1

// noFailures is the expected failed story count in the passing minimal result.
const noFailures = 0

// noSkipped is the expected skipped story count in the passing minimal result.
const noSkipped = 0

// noCacheHits is the expected cache-hit count in the passing minimal result.
const noCacheHits = 0

// orderZero is the Order index for the first story in the minimal result.
const orderZero = 0

// requiredTopLevelKeys returns the JSON keys that must appear at the top
// level of every octane.json report.
func requiredTopLevelKeys() []string {
	return []string{
		"schemaVersion",
		"octaneVersion",
		"runId",
		"startedAt",
		"finishedAt",
		"summary",
		"stories",
	}
}

// buildMinimalResult creates a minimal runner.RunResult suitable for
// schema validation tests.
func buildMinimalResult() *runner.RunResult {
	now := time.Date(
		schemaTestYear,
		schemaTestMonth,
		schemaTestDay,
		schemaTestHour,
		0, 0, 0,
		time.UTC,
	)

	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  now,
		FinishedAt: now.Add(5 * time.Second),
		Summary: runner.Summary{
			Total:     minimalTotalStories,
			Passed:    minimalTotalStories,
			Failed:    noFailures,
			Skipped:   noSkipped,
			CacheHits: noCacheHits,
		},
		Stories: []runner.StoryResult{
			{
				Order:       orderZero,
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

	writeErr := reportjson.WriteJSON(result, dir, opts)
	if writeErr != nil {
		t.Fatalf("WriteJSON: %v", writeErr)
	}

	outPath := filepath.Join(dir, "octane.json")

	data, err := os.ReadFile(outPath) //nolint:gosec // G304: t.TempDir path
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	var top map[string]any

	unmarshalErr := json.Unmarshal(data, &top)
	if unmarshalErr != nil {
		t.Fatalf("unmarshal: %v", unmarshalErr)
	}

	for _, key := range requiredTopLevelKeys() {
		if _, present := top[key]; !present {
			t.Errorf("missing required top-level key: %q", key)
		}
	}

	assertSchemaVersion(t, top)
	assertOctaneVersion(t, top)
	assertSummaryShape(t, top)
	assertStoriesShape(t, top)
}

// assertOctaneVersion verifies that octaneVersion is a non-empty string.
func assertOctaneVersion(t *testing.T, top map[string]any) {
	t.Helper()

	val, present := top["octaneVersion"]
	if !present {
		return
	}

	str, isStr := val.(string)
	if !isStr || str == "" {
		t.Errorf("octaneVersion: got %v, want non-empty string", val)
	}
}

// assertSchemaVersion verifies that schemaVersion is a non-zero number.
func assertSchemaVersion(t *testing.T, top map[string]any) {
	t.Helper()

	val, present := top["schemaVersion"]
	if !present {
		return
	}

	num, isFloat := val.(float64)
	if !isFloat || num <= 0 {
		t.Errorf("schemaVersion: got %v, want positive integer", val)
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

	summaryKeys := []string{"total", "passed", "failed", "skipped", "cacheHits"}
	for _, key := range summaryKeys {
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
