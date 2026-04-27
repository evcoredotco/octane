// Package json_test contains tests for the JSON emitter.
// Task: T-007-23.
package json_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	reportjson "github.com/evcoreco/octane/pkg/report/json"
	"github.com/evcoreco/octane/pkg/runner"
)

// updateFlag controls whether the golden file is regenerated.
var updateFlag = flag.Bool("update", false, "update golden files")

// goldenFilePath is the path to the golden JSON file, relative to this
// test file's directory.
const goldenFilePath = "testdata/golden.json"

// reportFileName is the output file name produced by WriteJSON.
const reportFileName = "octane.json"

// fixedTime returns a deterministic time for test fixtures.
func fixedTime(offsetSeconds int) time.Time {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// buildGoldenResult constructs a canned runner.RunResult with known
// data: three stories (passed, failed, skipped). The skipped story has
// a cause chain.
func buildGoldenResult() *runner.RunResult {
	passedTrace := &runner.Trace{
		Frames: [][]byte{
			[]byte(`[2,"abc123","BootNotification",{"reason":"PowerUp"}]`),
			[]byte(
				`[3,"abc123",{"currentTime":"2024-01-15T10:00:01Z","interval":300,"status":"Accepted"}]`,
			),
		},
	}

	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  fixedTime(0),
		FinishedAt: fixedTime(30),
		Summary: runner.Summary{
			Total:     3,
			Passed:    1,
			Failed:    1,
			Skipped:   1,
			CacheHits: 1,
		},
		Stories: []runner.StoryResult{
			{
				Order:       0,
				TestID:      "tc_boot_notification",
				ScopeKey:    "CP01",
				OCPPVersion: "1.6",
				Status:      runner.StatusPassed,
				CacheStatus: runner.CacheHitPass,
				StartedAt:   fixedTime(0),
				FinishedAt:  fixedTime(10),
				Findings:    nil,
				Trace:       passedTrace,
				Cause:       "",
				CauseChain:  nil,
			},
			{
				Order:       1,
				TestID:      "tc_heartbeat",
				ScopeKey:    "CP01",
				OCPPVersion: "1.6",
				Status:      runner.StatusFailed,
				CacheStatus: runner.CacheMiss,
				StartedAt:   fixedTime(10),
				FinishedAt:  fixedTime(20),
				Findings: []runner.Finding{
					{
						Message:  "heartbeat interval mismatch: got 600, want 300",
						Severity: "error",
					},
					{
						Message:  "response took 250ms",
						Severity: "warning",
					},
				},
				Trace:      nil,
				Cause:      "",
				CauseChain: nil,
			},
			{
				Order:       2,
				TestID:      "tc_status_notification",
				ScopeKey:    "CP01",
				OCPPVersion: "1.6",
				Status:      runner.StatusSkipped,
				CacheStatus: runner.CacheBypassed,
				StartedAt:   fixedTime(20),
				FinishedAt:  fixedTime(20),
				Findings:    nil,
				Trace:       nil,
				Cause:       "tc_heartbeat/CP01",
				CauseChain:  []string{"tc_heartbeat/CP01"},
			},
		},
	}
}

// readReport reads the octane.json output from a directory produced by
// WriteJSON.
func readReport(t *testing.T, dir string) []byte {
	t.Helper()

	path := filepath.Join(dir, reportFileName)

	data, err := os.ReadFile(path) //nolint:gosec // G304: t.TempDir path
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	return data
}

// Test_json_Golden verifies byte-deterministic JSON output. When the
// -update flag is set, the golden file is regenerated.
func Test_json_Golden(t *testing.T) {
	t.Parallel()

	result := buildGoldenResult()
	opts := report.JSONOptions{
		NoTraceOnPass: false,
		OctaneVersion: "0.1.0-test",
	}

	dir1 := t.TempDir()

	if err := reportjson.WriteJSON(result, dir1, opts); err != nil {
		t.Fatalf("WriteJSON dir1: %v", err)
	}

	got := readReport(t, dir1)

	// AC2: determinism — write a second time and compare bytes.
	dir2 := t.TempDir()

	if err := reportjson.WriteJSON(result, dir2, opts); err != nil {
		t.Fatalf("WriteJSON dir2: %v", err)
	}

	got2 := readReport(t, dir2)

	if !bytes.Equal(got, got2) {
		t.Error("non-deterministic output: first and second runs differ")
	}

	if *updateFlag {
		if err := os.MkdirAll("testdata", 0o750); err != nil {
			t.Fatalf("creating testdata: %v", err)
		}

		if err := os.WriteFile(goldenFilePath, got, 0o600); err != nil {
			t.Fatalf("updating golden file: %v", err)
		}

		t.Logf("golden file updated: %s", goldenFilePath)

		return
	}

	want, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf(
			"reading golden file %s: %v (run with -update to create it)",
			goldenFilePath, err,
		)
	}

	if !bytes.Equal(got, want) {
		t.Errorf(
			"output differs from golden file %s\n--- got ---\n%s\n--- want ---\n%s",
			goldenFilePath,
			got,
			want,
		)
	}
}
