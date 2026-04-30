// Package report_test contains black-box tests for the report emitter layer.
//
// Task: T-007-40.
package report_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	reportjson "github.com/evcoreco/octane/pkg/report/json"
	"github.com/evcoreco/octane/pkg/runner"
)

// goroutineCount is the number of parallel writers exercised by the
// concurrent write test (AC10: each run uses a distinct run-id).
const goroutineCount = 10

// validateRunIDEntry checks that a single directory entry in the shared output
// dir contains a valid octane.json with a matching run_id and records the
// run_id into seenRunIDs. It reports all invariant violations via t.Errorf and
// returns the run_id (empty string on failure).
func validateRunIDEntry(
	t *testing.T,
	sharedDir string,
	entry os.DirEntry,
	seenRunIDs map[string]struct{},
) {
	t.Helper()

	if !entry.IsDir() {
		t.Errorf("expected directory entry, got file %q", entry.Name())

		return
	}

	reportPath := filepath.Join(sharedDir, entry.Name(), "octane.json")

	data, readErr := os.ReadFile(filepath.Clean(reportPath))
	if readErr != nil {
		t.Errorf("reading %s: %v", reportPath, readErr)

		return
	}

	var top map[string]any

	unmarshalErr := json.Unmarshal(data, &top)
	if unmarshalErr != nil {
		t.Errorf("unmarshalling %s: %v", reportPath, unmarshalErr)

		return
	}

	runID, ok := top["runId"].(string)
	if !ok || runID == "" {
		t.Errorf("%s: missing or empty runId", reportPath)

		return
	}

	if runID != entry.Name() {
		t.Errorf(
			"%s: runId %q does not match directory name %q",
			reportPath, runID, entry.Name(),
		)
	}

	if _, dup := seenRunIDs[runID]; dup {
		t.Errorf("duplicate runId %q across concurrent writes", runID)
	}

	seenRunIDs[runID] = struct{}{}
}

// writeTestReport builds a minimal RunResult and calls reportjson.WriteJSON
// for a single goroutine in the concurrent test. Errors are sent to errs.
func writeTestReport(goroutineIdx int, sharedDir string, errs chan<- error) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	result := &runner.RunResult{
		RunID:      fmt.Sprintf("run-%02d", goroutineIdx),
		StartedAt:  fixedTime,
		FinishedAt: fixedTime.Add(5 * time.Second),
		Stories:    []runner.StoryResult{},
		Summary:    runner.Summary{}, //nolint:exhaustruct // zero-value fixture
	}

	outDir := filepath.Join(sharedDir, result.RunID)

	opts := report.JSONOptions{ //nolint:exhaustruct // NoTraceOnPass zero value is correct
		OctaneVersion: "test",
	}

	writeErr := reportjson.WriteJSON(result, outDir, opts)
	if writeErr != nil {
		errs <- fmt.Errorf("goroutine %d: %w", goroutineIdx, writeErr)
	}
}

// Test_report_WriteJSON_ConcurrentDistinctRunIDs asserts that N parallel
// WriteJSON calls with distinct run-ids produce no file conflicts and that
// every caller writes its own <run-id>/octane.json to the shared parent dir.
func Test_report_WriteJSON_ConcurrentDistinctRunIDs(t *testing.T) {
	t.Parallel()

	sharedDir := t.TempDir()

	errs := make(chan error, goroutineCount)

	var workGroup sync.WaitGroup

	workGroup.Add(goroutineCount)

	for goroutineIdx := range goroutineCount {
		go func() {
			defer workGroup.Done()

			writeTestReport(goroutineIdx, sharedDir, errs)
		}()
	}

	workGroup.Wait()
	close(errs)

	// Invariant: no goroutine reported an error.
	for err := range errs {
		t.Errorf("unexpected write error: %v", err)
	}

	// Invariant: exactly goroutineCount <run-id>/octane.json files exist.
	entries, err := os.ReadDir(sharedDir)
	if err != nil {
		t.Fatalf("reading shared dir: %v", err)
	}

	if len(entries) != goroutineCount {
		t.Fatalf(
			"expected %d run subdirectories, got %d",
			goroutineCount, len(entries),
		)
	}

	// Invariant: each file encodes a distinct RunID matching its directory name.
	seenRunIDs := make(map[string]struct{}, goroutineCount)

	for _, entry := range entries {
		validateRunIDEntry(t, sharedDir, entry, seenRunIDs)
	}

	if len(seenRunIDs) != goroutineCount {
		t.Errorf(
			"expected %d distinct run IDs, got %d",
			goroutineCount, len(seenRunIDs),
		)
	}
}
