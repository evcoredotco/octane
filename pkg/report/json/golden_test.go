// Package reportjson_test contains tests for the JSON emitter.
// Task: T-007-23.
package reportjson_test

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
//
// registration
//
//nolint:gochecknoglobals // flag must be package-level for flag.Bool
var updateFlag = flag.Bool("update", false, "update golden files")

// goldenFilePath is the path to the golden JSON file, relative to this
// test file's directory.
const goldenFilePath = "testdata/golden.json"

// reportFileName is the output file name produced by WriteJSON.
const reportFileName = "octane.json"

const (
	// goldenBaseYear is the year used in fixedTime for golden test fixtures.
	goldenBaseYear = 2024

	// goldenBaseMonth is the month (January) used in fixedTime.
	goldenBaseMonth = 1

	// goldenBaseDay is the day-of-month used in fixedTime.
	goldenBaseDay = 15

	// goldenBaseHour is the hour used in fixedTime.
	goldenBaseHour = 10

	// totalStories is the total number of stories in the golden result.
	totalStories = 3

	// orderFirst is the Order index for the first story (zero-based index 1).
	orderFirst = 1

	// orderSecond is the Order index for the second story.
	orderSecond = 2

	// scopeKeyCP01 is the station scope key used in golden test fixtures.
	scopeKeyCP01 = "CP01"

	// ocppVersion16 is the OCPP version string for OCPP 1.6.
	ocppVersion16 = "1.6"

	// finishedAtSec is the run finish offset in seconds from the base time.
	finishedAtSec = 30

	// startAt10 is the offset in seconds for stories starting at T+10.
	startAt10 = 10

	// startAt20 is the offset in seconds for stories starting at T+20.
	startAt20 = 20

	// dirPerms is the directory permission bits used when creating testdata.
	dirPerms = 0o750

	// filePerms is the file permission bits used when writing golden files.
	filePerms = 0o600

	// countOne is used for result counts that equal one (1 passed, 1 failed,
	// 1 skipped, 1 cache hit) in the golden test fixture.
	countOne = 1
)

// fixedTime returns a deterministic time for test fixtures.
func fixedTime(offsetSeconds int) time.Time {
	base := time.Date(
		goldenBaseYear,
		goldenBaseMonth,
		goldenBaseDay,
		goldenBaseHour,
		0, 0, 0,
		time.UTC,
	)

	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// goldenPassedStory returns the passed BootNotification story fixture.
func goldenPassedStory() runner.StoryResult {
	passedTrace := &runner.Trace{
		Frames: [][]byte{
			[]byte(`[2,"abc123","BootNotification",{"reason":"PowerUp"}]`),
			[]byte(`[3,"abc123",{"currentTime":"2024-01-15T10:00:01Z",` +
				`"interval":300,"status":"Accepted"}]`),
		},
	}

	return runner.StoryResult{
		Order:       0,
		TestID:      "tc_boot_notification",
		ScopeKey:    scopeKeyCP01,
		OCPPVersion: ocppVersion16,
		Status:      runner.StatusPassed,
		CacheStatus: runner.CacheHitPass,
		StartedAt:   fixedTime(0),
		FinishedAt:  fixedTime(startAt10),
		Findings:    nil,
		Trace:       passedTrace,
		Cause:       "",
		CauseChain:  nil,
	}
}

// goldenFailedStory returns the failed Heartbeat story fixture.
func goldenFailedStory() runner.StoryResult {
	return runner.StoryResult{
		Order:       orderFirst,
		TestID:      "tc_heartbeat",
		ScopeKey:    scopeKeyCP01,
		OCPPVersion: ocppVersion16,
		Status:      runner.StatusFailed,
		CacheStatus: runner.CacheMiss,
		StartedAt:   fixedTime(startAt10),
		FinishedAt:  fixedTime(startAt20),
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
	}
}

// goldenSkippedStory returns the skipped StatusNotification story fixture.
func goldenSkippedStory() runner.StoryResult {
	return runner.StoryResult{
		Order:       orderSecond,
		TestID:      "tc_status_notification",
		ScopeKey:    scopeKeyCP01,
		OCPPVersion: ocppVersion16,
		Status:      runner.StatusSkipped,
		CacheStatus: runner.CacheBypassed,
		StartedAt:   fixedTime(startAt20),
		FinishedAt:  fixedTime(startAt20),
		Findings:    nil,
		Trace:       nil,
		Cause:       "tc_heartbeat/CP01",
		CauseChain:  []string{"tc_heartbeat/CP01"},
	}
}

// buildGoldenResult constructs a canned runner.RunResult with known
// data: three stories (passed, failed, skipped). The skipped story has
// a cause chain.
func buildGoldenResult() *runner.RunResult {
	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  fixedTime(0),
		FinishedAt: fixedTime(finishedAtSec),
		Summary: runner.Summary{
			Total:     totalStories,
			Passed:    countOne,
			Failed:    countOne,
			Skipped:   countOne,
			CacheHits: countOne,
		},
		Stories: []runner.StoryResult{
			goldenPassedStory(),
			goldenFailedStory(),
			goldenSkippedStory(),
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

// updateGoldenJSONFile rewrites the golden JSON file with got.
// Called by Test_json_Golden when the -update flag is set.
func updateGoldenJSONFile(t *testing.T, got []byte) {
	t.Helper()

	err := os.MkdirAll("testdata", dirPerms)
	if err != nil {
		t.Fatalf("creating testdata: %v", err)
	}

	err = os.WriteFile(goldenFilePath, got, filePerms)
	if err != nil {
		t.Fatalf("updating golden file: %v", err)
	}

	t.Logf("golden file updated: %s", goldenFilePath)
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

	err := reportjson.WriteJSON(result, dir1, opts)
	if err != nil {
		t.Fatalf("WriteJSON dir1: %v", err)
	}

	got := readReport(t, dir1)

	// AC2: determinism — write a second time and compare bytes.
	dir2 := t.TempDir()

	err = reportjson.WriteJSON(result, dir2, opts)
	if err != nil {
		t.Fatalf("WriteJSON dir2: %v", err)
	}

	got2 := readReport(t, dir2)

	if !bytes.Equal(got, got2) {
		t.Error("non-deterministic output: first and second runs differ")
	}

	if *updateFlag {
		updateGoldenJSONFile(t, got)

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
			"output differs from golden file %s\n"+
				"--- got ---\n%s\n--- want ---\n%s",
			goldenFilePath,
			got,
			want,
		)
	}
}
