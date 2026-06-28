package runner

import "context"

// cancelPendingNodes marks all remaining pending nodes in the
// scheduler state as StatusSkipped with an appropriate finding.
// This is called when ctx is cancelled so that the RunResult
// reflects every story's final status.
func cancelPendingNodes(
	ctx context.Context,
	state *schedulerState,
) {
	cancelMsg := "skipped: run cancelled"

	ctxErr := ctx.Err()
	if ctxErr != nil {
		cancelMsg = "skipped: run cancelled: " + ctxErr.Error()
	}

	for nodeID, nodeState := range state.status {
		if nodeState != nodePending {
			continue
		}

		nodeParts := splitNodeID(nodeID)
		storyID := nodeParts.storyID
		scopeKey := nodeParts.scopeKey

		state.status[nodeID] = nodeDone
		state.result[nodeID] = StoryResult{
			Order:       zeroOrder,
			TestID:      storyID,
			ScopeKey:    scopeKey,
			OCPPVersion: "",
			Status:      StatusSkipped,
			CacheStatus: CacheMiss,
			StartedAt:   state.result[nodeID].StartedAt,
			FinishedAt:  state.result[nodeID].FinishedAt,
			Findings: []Finding{
				{
					Message:  cancelMsg,
					Severity: severityError,
				},
			},
			Trace:      nil,
			Cause:      "",
			CauseChain: nil,
		}
	}
}
