// Package robotxml implements the Robot Framework output.xml emitter for
// OCTANE run reports. The public entry point is [WriteRobotXML], which
// projects a [runner.RunResult] into a byte-deterministic output.xml file
// that conforms to the Robot Framework 7.x output schema.
//
// Task: T-007-31.
package robotxml

import (
	"time"

	"github.com/evcoreco/octane/pkg/report/model"
)

// robotTimeLayout is the Robot Framework datetime format.
// Example: "20260426 08:00:01.000"
const robotTimeLayout = "20060102 15:04:05.000"

// statusString maps an OCTANE story status string to the Robot Framework
// status attribute value. "passed" → "PASS", "failed" → "FAIL",
// "skipped" → "SKIP", any other value (including "bypassed") → "NOT RUN".
func statusString(status string) string {
	switch status {
	case "passed":
		return "PASS"
	case "failed":
		return "FAIL"
	case "skipped":
		return "SKIP"
	default:
		return "NOT RUN"
	}
}

// suiteStatus computes the suite-level Robot Framework status from all
// story statuses. Returns "FAIL" when any story failed; "PASS" otherwise.
func suiteStatus(stories []model.StoryReport) string {
	for _, s := range stories {
		if s.Status == "failed" {
			return "FAIL"
		}
	}

	return "PASS"
}

// robotTime formats a time.Time value for Robot Framework datetime fields.
// The format is "YYYYMMDD HH:MM:SS.mmm" as required by Robot Framework 7.x.
func robotTime(t time.Time) string {
	return t.UTC().Format(robotTimeLayout)
}
