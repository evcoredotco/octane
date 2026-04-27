// Package robotxml implements the Robot Framework output.xml emitter for
// OCTANE run reports.
//
// Task: T-007-30.
package robotxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/evcoreco/octane/pkg/report"
	"github.com/evcoreco/octane/pkg/report/model"
	"github.com/evcoreco/octane/pkg/runner"
)

// outputFileName is the name of the Robot XML file written into the output
// directory.
const outputFileName = "output.xml"

// xmlHeader is written verbatim before the marshaled XML body.
const xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

// octaneGenerator is the generator string embedded in the <robot> element.
const octaneGenerator = "octane/0.1.0"

// defaultSuiteName is used when RobotXMLOptions.SuiteName is empty.
const defaultSuiteName = "OCTANE Conformance"

// suiteID is the fixed ID for the outer <suite> element.
const suiteID = "s1"

// severityMajor is the severity threshold at which a finding is mapped to an
// ERROR-level message. Findings with severity "error" meet this threshold;
// "warning" and "info" become WARN-level messages.
const severityMajor = "error"

// ---------------------------------------------------------------------------
// XML struct types
// ---------------------------------------------------------------------------

// xmlRobot is the root <robot> element.
type xmlRobot struct {
	XMLName   xml.Name `xml:"robot"`
	Generated string   `xml:"generated,attr"`
	Generator string   `xml:"generator,attr"`
	Suite     xmlSuite `xml:"suite"`
	Stats     xmlEmpty `xml:"statistics"`
	Errors    xmlEmpty `xml:"errors"`
}

// xmlEmpty represents an empty XML element with no attributes or children.
type xmlEmpty struct{}

// xmlSuite is the <suite> element wrapping all test cases.
type xmlSuite struct {
	ID     string    `xml:"id,attr"`
	Name   string    `xml:"name,attr"`
	Source string    `xml:"source,attr"`
	Tests  []xmlTest `xml:"test"`
	Status xmlStatus `xml:"status"`
}

// xmlTest is a single <test> element representing one story execution.
type xmlTest struct {
	ID       string    `xml:"id,attr"`
	Name     string    `xml:"name,attr"`
	Keywords []xmlKw   `xml:"kw,omitempty"`
	Status   xmlStatus `xml:"status"`
}

// xmlKw is a <kw> element representing a single trace frame as a log entry.
type xmlKw struct {
	Name   string      `xml:"name,attr"`
	Status xmlKwStatus `xml:"status"`
	Msg    string      `xml:"msg,omitempty"`
}

// xmlKwStatus is the <status> element inside a <kw>. It carries no child
// messages, only timing attributes.
type xmlKwStatus struct {
	Status    string `xml:"status,attr"`
	StartTime string `xml:"starttime,attr"`
	EndTime   string `xml:"endtime,attr"`
}

// xmlStatus is the <status> element on a test or suite. When Messages is
// non-empty, child <msg> elements are written beneath the status attributes.
type xmlStatus struct {
	Status    string   `xml:"status,attr"`
	StartTime string   `xml:"starttime,attr"`
	EndTime   string   `xml:"endtime,attr"`
	Messages  []xmlMsg `xml:"msg,omitempty"`
}

// xmlMsg is a <msg> child element inside <status> carrying a finding text.
type xmlMsg struct {
	Level string `xml:"level,attr"`
	Text  string `xml:",chardata"`
}

// ---------------------------------------------------------------------------
// Public entry point
// ---------------------------------------------------------------------------

// WriteRobotXML projects result into a [model.Report], builds a Robot
// Framework 7.x output.xml structure, and writes it to dir/output.xml.
// The directory is created with os.MkdirAll when it does not already exist.
func WriteRobotXML(
	result *runner.RunResult,
	dir string,
	opts report.RobotXMLOptions,
) error {
	rep := model.FromRunner(result, "")

	root := buildRobotXML(rep, opts)

	data, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	outPath := filepath.Join(dir, outputFileName)

	payload := make([]byte, 0, len(xmlHeader)+len(data)+1)
	payload = append(payload, xmlHeader...)
	payload = append(payload, data...)
	payload = append(payload, '\n')

	return os.WriteFile(outPath, payload, 0o600)
}

// ---------------------------------------------------------------------------
// Build helpers
// ---------------------------------------------------------------------------

// buildRobotXML constructs the xmlRobot document from a model.Report.
func buildRobotXML(rep *model.Report, opts report.RobotXMLOptions) xmlRobot {
	suiteName := opts.SuiteName
	if suiteName == "" {
		suiteName = defaultSuiteName
	}

	tests := buildTests(rep.Stories)
	suiteStatusStr := suiteStatus(rep.Stories)

	suite := xmlSuite{
		ID:     suiteID,
		Name:   suiteName,
		Source: "",
		Tests:  tests,
		Status: xmlStatus{
			Status:    suiteStatusStr,
			StartTime: robotTime(rep.StartedAt),
			EndTime:   robotTime(rep.FinishedAt),
			Messages:  nil,
		},
	}

	return xmlRobot{
		XMLName:   xml.Name{Space: "", Local: ""},
		Generated: robotTime(rep.StartedAt),
		Generator: octaneGenerator,
		Suite:     suite,
		Stats:     xmlEmpty{},
		Errors:    xmlEmpty{},
	}
}

// buildTests converts a slice of model.StoryReport to a slice of xmlTest.
func buildTests(stories []model.StoryReport) []xmlTest {
	out := make([]xmlTest, len(stories))

	for idx, story := range stories {
		out[idx] = buildTest(idx+1, story)
	}

	return out
}

// buildTest converts a single model.StoryReport to an xmlTest element.
// The test ID is "s1-tN" where N is the one-based index.
func buildTest(idx int, story model.StoryReport) xmlTest {
	name := testName(story)
	status := statusString(story.Status)
	msgs := buildMessages(story.Findings)
	kws := buildKeywords(story)

	return xmlTest{
		ID:       fmt.Sprintf("%s-t%d", suiteID, idx),
		Name:     name,
		Keywords: kws,
		Status: xmlStatus{
			Status:    status,
			StartTime: robotTime(story.StartedAt),
			EndTime:   robotTime(story.FinishedAt),
			Messages:  msgs,
		},
	}
}

// testName formats the test display name. When ScopeKey is non-empty the
// format is "<test_id> (<scope_key>)", otherwise just "<test_id>".
func testName(story model.StoryReport) string {
	if story.ScopeKey == "" {
		return story.TestID
	}

	return fmt.Sprintf("%s (%s)", story.TestID, story.ScopeKey)
}

// buildMessages converts story findings to <msg> child elements. Findings
// with severity "error" become ERROR-level; all others become WARN-level.
func buildMessages(findings []model.Finding) []xmlMsg {
	if len(findings) == 0 {
		return nil
	}

	out := make([]xmlMsg, 0, len(findings))

	for _, finding := range findings {
		level := "WARN"
		if finding.Severity == severityMajor {
			level = "ERROR"
		}

		out = append(out, xmlMsg{
			Level: level,
			Text:  finding.Message,
		})
	}

	return out
}

// buildKeywords converts trace frames to <kw> elements. Each frame becomes a
// keyword named "trace.frame" with a log message containing the raw JSON.
// Returns nil when no trace is present.
func buildKeywords(story model.StoryReport) []xmlKw {
	if story.Trace == nil || len(story.Trace.Frames) == 0 {
		return nil
	}

	out := make([]xmlKw, len(story.Trace.Frames))
	frameTime := robotTime(story.StartedAt)

	for idx, frame := range story.Trace.Frames {
		out[idx] = xmlKw{
			Name: "trace.frame",
			Status: xmlKwStatus{
				Status:    "PASS",
				StartTime: frameTime,
				EndTime:   frameTime,
			},
			Msg: string(frame.Raw),
		}
	}

	return out
}
