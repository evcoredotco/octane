// Package robotxml_test contains tests for the Robot XML emitter.
//
// Task: T-007-32.
package robotxml_test

import (
	"bytes"
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	"github.com/evcoreco/octane/pkg/report/robotxml"
	"github.com/evcoreco/octane/pkg/runner"
)

// updateFlag controls whether the golden file is regenerated.
var updateFlag = flag.Bool("update", false, "update golden files")

// goldenFilePath is the path to the golden XML file, relative to this test
// file's directory.
const goldenFilePath = "testdata/output.xml"

// outputFileName is the file name produced by WriteRobotXML.
const outputFileName = "output.xml"

// fixedTime returns a deterministic time for test fixtures.
func fixedTime(offsetSeconds int) time.Time {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// buildGoldenResult constructs the same canned runner.RunResult used by the
// JSON golden test: three stories (passed, failed, skipped). The skipped
// story has a cause chain.
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

// readOutput reads the output.xml file from a directory produced by
// WriteRobotXML.
func readOutput(t *testing.T, dir string) []byte {
	t.Helper()

	path := filepath.Join(dir, outputFileName)

	data, err := os.ReadFile(path) //nolint:gosec // G304: t.TempDir path
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	return data
}

// Test_robotxml_Golden verifies byte-deterministic Robot XML output. When the
// -update flag is set, the golden file is regenerated. The test also verifies
// that the produced XML is well-formed by parsing it back with encoding/xml.
func Test_robotxml_Golden(t *testing.T) {
	t.Parallel()

	result := buildGoldenResult()
	opts := report.RobotXMLOptions{
		SuiteName: "OCTANE Conformance",
	}

	dir1 := t.TempDir()

	if err := robotxml.WriteRobotXML(result, dir1, opts); err != nil {
		t.Fatalf("WriteRobotXML dir1: %v", err)
	}

	got := readOutput(t, dir1)

	// Determinism check: write a second time and compare bytes.
	dir2 := t.TempDir()

	if err := robotxml.WriteRobotXML(result, dir2, opts); err != nil {
		t.Fatalf("WriteRobotXML dir2: %v", err)
	}

	got2 := readOutput(t, dir2)

	if !bytes.Equal(got, got2) {
		t.Error("non-deterministic output: first and second runs differ")
	}

	// Well-formedness check: parse the output back with encoding/xml.
	var parsed interface{}
	if err := xml.Unmarshal(got, &parsed); err != nil {
		t.Errorf("output.xml is not well-formed XML: %v", err)
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
