// Package json implements the JSON emitter for OCTANE run reports.
// This file implements cause-chain walking helpers.
//
// Task: T-007-22.
package json

import (
	"github.com/octane-project/octane/pkg/report/model"
)

// walkCauseChain returns the cause chain for a story. When the
// model.StoryReport already has a non-empty CauseChain it is returned
// as-is. When the chain is empty but a Cause is set, a single-element
// slice containing the Cause is returned as a defensive fallback. When
// neither field is set, nil is returned.
//
// This defensive fallback exists because the runner always populates
// CauseChain when it populates Cause (spec 005 AC4), but callers from
// other code paths (e.g. tests constructing StoryResult by hand) may
// omit CauseChain.
func walkCauseChain(src model.StoryReport) []string {
	if len(src.CauseChain) > 0 {
		return src.CauseChain
	}

	if src.Cause != "" {
		return []string{src.Cause}
	}

	return nil
}
