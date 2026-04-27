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

	"github.com/octane-project/octane/pkg/report"
	reportjson "github.com/octane-project/octane/pkg/report/json"
	"github.com/octane-project/octane/pkg/runner"
)

// goroutineCount is the number of parallel writers exercised by the
// concurrent write test (AC10: each run uses a distinct run-id).
const goroutineCount = 10

// Test_report_WriteJSON_ConcurrentDistinctRunIDs asserts that N parallel
// WriteJSON calls with distinct run-ids produce no file conflicts and that
// every caller writes its own <run-id>/octane.json to the shared parent dir.
func Test_report_WriteJSON_ConcurrentDistinctRunIDs(t *testing.T) {
	t.Parallel()

	sharedDir := t.TempDir()

	errs := make(chan error, goroutineCount)

	var wg sync.WaitGroup

	wg.Add(goroutineCount)

	for i := range goroutineCount {
		i := i // capture loop variable for Go < 1.22 compatibility

		go func() {
			defer wg.Done()

			result := &runner.RunResult{
				RunID:      fmt.Sprintf("run-%02d", i),
				StartedAt:  time.Now(),
				FinishedAt: time.Now(),
				Stories:    []runner.StoryResult{},
				Summary:    runner.Summary{ //nolint:exhaustruct // zero-value fixture
				},
			}

			outDir := filepath.Join(sharedDir, result.RunID)

			opts := report.JSONOptions{ //nolint:exhaustruct // NoTraceOnPass zero value is correct
				OctaneVersion: "test",
			}

			if err := reportjson.WriteJSON(result, outDir, opts); err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", i, err)
			}
		}()
	}

	wg.Wait()
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
			goroutineCount,
			len(entries),
		)
	}

	// Invariant: each file encodes a distinct RunID matching its directory name.
	seenRunIDs := make(map[string]struct{}, goroutineCount)

	for _, entry := range entries {
		if !entry.IsDir() {
			t.Errorf("expected directory entry, got file %q", entry.Name())

			continue
		}

		reportPath := filepath.Join(sharedDir, entry.Name(), "octane.json")

		data, readErr := os.ReadFile(
			reportPath,
		) //nolint:gosec // G304: t.TempDir path
		if readErr != nil {
			t.Errorf("reading %s: %v", reportPath, readErr)

			continue
		}

		var top map[string]any

		if unmarshalErr := json.Unmarshal(data, &top); unmarshalErr != nil {
			t.Errorf("unmarshalling %s: %v", reportPath, unmarshalErr)

			continue
		}

		runID, ok := top["run_id"].(string)
		if !ok || runID == "" {
			t.Errorf("%s: missing or empty run_id", reportPath)

			continue
		}

		// Invariant: run_id encoded in the file matches the directory name.
		if runID != entry.Name() {
			t.Errorf(
				"%s: run_id %q does not match directory name %q",
				reportPath,
				runID,
				entry.Name(),
			)
		}

		// Invariant: run_id values are unique across all written files.
		if _, dup := seenRunIDs[runID]; dup {
			t.Errorf("duplicate run_id %q across concurrent writes", runID)
		}

		seenRunIDs[runID] = struct{}{}
	}

	if len(seenRunIDs) != goroutineCount {
		t.Errorf(
			"expected %d distinct run IDs, got %d",
			goroutineCount,
			len(seenRunIDs),
		)
	}
}
