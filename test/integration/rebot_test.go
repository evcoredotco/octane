//go:build integration

// Package integration_test contains integration tests that require external
// services or binaries. Build with -tags integration to include these tests.
//
// Task: T-007-33.
package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/report"
	"github.com/evcoreco/octane/pkg/report/robotxml"
	"github.com/evcoreco/octane/pkg/runner"
)

// rebotImage is the Docker image used to run rebot for Robot Framework XML
// validation.
const rebotImage = "ghcr.io/robotframework/rfdocker:7.0"

// fixedRebotTime returns a deterministic time for the rebot test fixture.
func fixedRebotTime(offsetSeconds int) time.Time {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// buildRebotResult constructs a minimal runner.RunResult suitable for rebot
// validation: one passed story, one failed story.
func buildRebotResult() *runner.RunResult {
	return &runner.RunResult{
		RunID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		StartedAt:  fixedRebotTime(0),
		FinishedAt: fixedRebotTime(30),
		Summary: runner.Summary{
			Total:     2,
			Passed:    1,
			Failed:    1,
			Skipped:   0,
			CacheHits: 0,
		},
		Stories: []runner.StoryResult{
			{
				Order:       0,
				TestID:      "tc_boot_notification",
				ScopeKey:    "CP01",
				OCPPVersion: "1.6",
				Status:      runner.StatusPassed,
				CacheStatus: runner.CacheMiss,
				StartedAt:   fixedRebotTime(0),
				FinishedAt:  fixedRebotTime(10),
				Findings:    nil,
				Trace:       nil,
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
				StartedAt:   fixedRebotTime(10),
				FinishedAt:  fixedRebotTime(20),
				Findings: []runner.Finding{
					{
						Message:  "heartbeat interval mismatch: got 600, want 300",
						Severity: "error",
					},
				},
				Trace:      nil,
				Cause:      "",
				CauseChain: nil,
			},
		},
	}
}

// TestRebot_RobotXML pipes the Robot XML output through Docker rebot and
// asserts that rebot exits cleanly, validating the output.xml schema.
//
// Skip conditions:
//   - docker binary not found on PATH
//   - build tag "integration" not set (enforced by the build constraint above)
func TestRebot_RobotXML(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found on PATH; skipping rebot integration test")
	}

	inDir := t.TempDir()
	outDir := t.TempDir()

	result := buildRebotResult()
	opts := report.RobotXMLOptions{SuiteName: "OCTANE Conformance"}

	if err := robotxml.WriteRobotXML(result, inDir, opts); err != nil {
		t.Fatalf("WriteRobotXML: %v", err)
	}

	xmlPath := filepath.Join(inDir, "output.xml")

	if _, err := os.Stat(xmlPath); err != nil {
		t.Fatalf("output.xml not created: %v", err)
	}

	// Run: docker run --rm \
	//   -v <inDir>:/in:ro \
	//   -v <outDir>:/out \
	//   <image> rebot --report /out/report.html /in/output.xml
	//nolint:gosec // G204: intentional use of exec for Docker integration test
	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", inDir+":/in:ro",
		"-v", outDir+":/out",
		rebotImage,
		"rebot", "--report", "/out/report.html", "/in/output.xml",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"rebot exited with error: %v\noutput:\n%s",
			err, string(out),
		)
	}

	// Assert no unexpected stderr output: rebot writes its summary to stdout;
	// any content on stderr indicates a problem.
	combined := string(out)
	if strings.Contains(combined, "Error") ||
		strings.Contains(combined, "ERROR") {
		t.Errorf("rebot output contained error indicators:\n%s", combined)
	}
}
