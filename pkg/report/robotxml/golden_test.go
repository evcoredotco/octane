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

const (
	// goldenBaseYear is the year used in fixedTime for golden test fixtures.
	goldenBaseYear = 2024

	// goldenBaseMonth is the month (January) used in fixedTime.
	goldenBaseMonth = 1

	// goldenBaseDay is the day-of-month used in fixedTime.
	goldenBaseDay = 15

	// goldenBaseHour is the hour used in fixedTime.
	goldenBaseHour = 10

	// scopeKeyCP01 is the station scope key used in golden test fixtures.
	scopeKeyCP01 = "CP01"

	// ocppVersion16 is the OCPP version string for OCPP 1.6.
	ocppVersion16 = "1.6"

	// finishedAt10 is the story finish offset in seconds for the first story.
	finishedAt10 = 10

	// startAt10 is the offset in seconds for stories starting at T+10.
	startAt10 = 10

	// startAt20 is the offset in seconds for stories starting at T+20.
	startAt20 = 20

	// finishedAt20 is the finish offset in seconds for the second story.
	finishedAt20 = 20

	// orderFirst is the Order index for the first non-zero story (second
	// position, one-based).
	orderFirst = 1

	// orderSkipped is the Order index for the skipped (third) story.
	orderSkipped = 2

	// finishedAtSec is the run finish offset in seconds from the base time.
	finishedAtSec = 30

	// totalStories is the total number of stories in the golden result.
	totalStories = 3

	// countOne is used for result counts that equal one (1 passed, 1 failed,
	// 1 skipped, 1 cache hit) in the golden test fixture.
	countOne = 1

	// zeroOffset is the zero time offset used in fixedTime(zeroOffset) calls.
	zeroOffset = 0

	// zeroTimeField is the zero value for minute, second, and nanosecond
	// arguments in time.Date calls.
	zeroTimeField = 0

	// orderZero is the Order index for the first story (position zero).
	orderZero = 0

	// dirPerms is the directory permission bits used when creating testdata.
	dirPerms = 0o750

	// filePerms is the file permission bits used when writing golden files.
	filePerms = 0o600
)

// updateFlag controls whether the golden file is regenerated.
// It must be package-level for flag.Bool registration.
//
//nolint:gochecknoglobals
var updateFlag = flag.Bool("update", false, "update golden files")

// goldenFilePath is the path to the golden XML file, relative to this test
// file's directory.
const goldenFilePath = "testdata/output.xml"

// outputFileName is the file name produced by WriteRobotXML.
const outputFileName = "output.xml"

// fixedTime returns a deterministic time for test fixtures.
func fixedTime(offsetSeconds int) time.Time {
	base := time.Date(
		goldenBaseYear, goldenBaseMonth, goldenBaseDay,
		goldenBaseHour, zeroTimeField, zeroTimeField, zeroTimeField,
		time.UTC,
	)

	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// xmlPassedStory returns the passed BootNotification story fixture.
func xmlPassedStory() runner.StoryResult {
	passedTrace := &runner.Trace{
		Frames: [][]byte{
			[]byte(`[2,"abc123","BootNotification",{"reason":"PowerUp"}]`),
			[]byte(
				`[3,"abc123",{"currentTime":"2024-01-15T10:00:01Z",` +
					`"interval":300,"status":"Accepted"}]`,
			),
		},
	}

	return runner.StoryResult{
		Order:       orderZero,
		TestID:      "tc_boot_notification",
		ScopeKey:    scopeKeyCP01,
		OCPPVersion: ocppVersion16,
		Status:      runner.StatusPassed,
		CacheStatus: runner.CacheHitPass,
		StartedAt:   fixedTime(zeroOffset),
		FinishedAt:  fixedTime(finishedAt10),
		Findings:    nil,
		Trace:       passedTrace,
		Cause:       "",
		CauseChain:  nil,
	}
}

// xmlFailedStory returns the failed Heartbeat story fixture.
func xmlFailedStory() runner.StoryResult {
	return runner.StoryResult{
		Order:       orderFirst,
		TestID:      "tc_heartbeat",
		ScopeKey:    scopeKeyCP01,
		OCPPVersion: ocppVersion16,
		Status:      runner.StatusFailed,
		CacheStatus: runner.CacheMiss,
		StartedAt:   fixedTime(startAt10),
		FinishedAt:  fixedTime(finishedAt20),
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

// xmlSkippedStory returns the skipped StatusNotification story fixture.
func xmlSkippedStory() runner.StoryResult {
	return runner.StoryResult{
		Order:       orderSkipped,
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

// buildGoldenResult constructs the same canned runner.RunResult used by the
// JSON golden test: three stories (passed, failed, skipped). The skipped
// story has a cause chain.
func buildGoldenResult() *runner.RunResult {
	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  fixedTime(zeroOffset),
		FinishedAt: fixedTime(finishedAtSec),
		Summary: runner.Summary{
			Total:     totalStories,
			Passed:    countOne,
			Failed:    countOne,
			Skipped:   countOne,
			CacheHits: countOne,
		},
		Stories: []runner.StoryResult{
			xmlPassedStory(),
			xmlFailedStory(),
			xmlSkippedStory(),
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

// updateGoldenXMLFile rewrites the golden output.xml file with got.
// Called by Test_robotxml_Golden when the -update flag is set.
func updateGoldenXMLFile(t *testing.T, got []byte) {
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

	err := robotxml.WriteRobotXML(result, dir1, opts)
	if err != nil {
		t.Fatalf("WriteRobotXML dir1: %v", err)
	}

	got := readOutput(t, dir1)

	// Determinism check: write a second time and compare bytes.
	dir2 := t.TempDir()

	err = robotxml.WriteRobotXML(result, dir2, opts)
	if err != nil {
		t.Fatalf("WriteRobotXML dir2: %v", err)
	}

	got2 := readOutput(t, dir2)

	if !bytes.Equal(got, got2) {
		t.Error("non-deterministic output: first and second runs differ")
	}

	// Well-formedness check: parse the output back with encoding/xml.
	var parsed any

	err = xml.Unmarshal(got, &parsed)
	if err != nil {
		t.Errorf("output.xml is not well-formed XML: %v", err)
	}

	if *updateFlag {
		updateGoldenXMLFile(t, got)

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
