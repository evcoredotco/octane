package reportjson

import (
	encodingjson "encoding/json"

	"github.com/evcoreco/octane/pkg/report"
	"github.com/evcoreco/octane/pkg/report/model"
)

// jsonTrace is the JSON serialization struct for a story's wire trace.
type jsonTrace struct {
	Frames []encodingjson.RawMessage `json:"frames"`
}

// statusPassed is the canonical string for a passing story.
const statusPassed = "passed"

// buildTrace converts a model.Trace to a jsonTrace and the tracePresent
// flag. Returns (false, nil) when:
//   - the trace is nil (no trace was captured), or
//   - opts.NoTraceOnPass is true and the story passed.
func buildTrace(
	src *model.Trace,
	status string,
	opts report.JSONOptions,
) (bool, *jsonTrace) {
	if src == nil {
		return false, nil
	}

	if opts.NoTraceOnPass && status == statusPassed {
		return false, nil
	}

	frames := make([]encodingjson.RawMessage, len(src.Frames))

	for idx, frame := range src.Frames {
		frames[idx] = encodingjson.RawMessage(frame.Raw)
	}

	return true, &jsonTrace{Frames: frames}
}
