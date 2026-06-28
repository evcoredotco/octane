// T-005-41: scope-aware traversal.
//
// This file implements the scheduler's eligibility computation and
// the scope-aware deduplication logic described in ADR 0019 §
// "Eligible-set computation". It operates on nodeID strings, not
// story ASTs, so that the same traversal logic works for all three
// scope types (per-station, per-run, global).

package runner

import "slices"

// nodeStatus is the internal execution state of a single DAG node
// within the scheduler's three-set model (ADR 0019).
type nodeStatus int

const (
	// nodePending indicates the node has not yet been dispatched.
	nodePending nodeStatus = iota

	// nodeRunning indicates the node has been dispatched to a
	// worker but has not yet completed.
	nodeRunning

	// nodeDone indicates the node has reached a terminal status
	// (passed, failed, skipped, or cached).
	nodeDone
)

// zeroRunningCount is the initial running-goroutine counter value.
const zeroRunningCount = 0

// emptySliceLen is the zero-length sentinel for make([]T, 0, n) calls.
const emptySliceLen = 0

// zeroOrder is the zero-value sentinel for the Order field of StoryResult.
const zeroOrder = 0

// zeroPendingCount is the initial pending-node counter value.
const zeroPendingCount = 0

// schedulerState holds the mutable state that the scheduler goroutine
// updates on every tick. All fields are accessed from the scheduler
// goroutine only; there is no concurrent mutation.
type schedulerState struct {
	// status maps nodeID → nodeStatus. Every node in the DAG has
	// an entry; the initial value is nodePending.
	status map[string]nodeStatus

	// result maps nodeID → terminal StoryResult. A node has an
	// entry only after it enters nodeDone.
	result map[string]StoryResult

	// prereqs maps nodeID → slice of prerequisite nodeIDs derived
	// from the DAG edges. Built once from the DAG before the run
	// starts.
	prereqs map[string][]string

	// dependents maps nodeID → slice of dependent nodeIDs. Built
	// from the same DAG edges, used for failure propagation.
	dependents map[string][]string

	// order is the stable topological ordering of all nodes in
	// the DAG. The scheduler visits this slice on every tick to
	// find eligible nodes.
	order []string

	// running is the count of nodes currently in nodeRunning.
	running int
}

// newSchedulerState builds the initial scheduler state from a
// resolved DAG result and the stable topological node order.
func newSchedulerState(
	dagResult *buildDAGResult,
	topoOrder []string,
) *schedulerState {
	nodeCount := len(dagResult.nodes)

	status := make(map[string]nodeStatus, nodeCount)
	prereqs := make(map[string][]string, nodeCount)
	dependents := make(map[string][]string, nodeCount)

	for _, sn := range dagResult.nodes {
		status[sn.nodeID] = nodePending
		prereqs[sn.nodeID] = nil
		dependents[sn.nodeID] = nil
	}

	for _, edge := range dagResult.graph.Edges() {
		prereqs[edge.To] = append(prereqs[edge.To], edge.From)
		dependents[edge.From] = append(dependents[edge.From], edge.To)
	}

	return &schedulerState{
		status:     status,
		result:     make(map[string]StoryResult, nodeCount),
		prereqs:    prereqs,
		dependents: dependents,
		order:      topoOrder,
		running:    zeroRunningCount,
	}
}

// eligibleNodes returns the slice of pending node IDs whose every
// prerequisite has a terminal status of passed or cached-hit.
// The returned slice is sorted by nodeID for deterministic dispatch
// (ADR 0019 §"Dispatch order and determinism").
func (ss *schedulerState) eligibleNodes() []string {
	eligible := make([]string, emptySliceLen, len(ss.order))

	for _, nodeID := range ss.order {
		if ss.status[nodeID] != nodePending {
			continue
		}

		if ss.allPrereqsPassed(nodeID) {
			eligible = append(eligible, nodeID)
		}
	}

	slices.Sort(eligible)

	return eligible
}

// allPrereqsPassed reports whether every prerequisite of nodeID has
// completed with a passing or cached status.
func (ss *schedulerState) allPrereqsPassed(nodeID string) bool {
	for _, prereq := range ss.prereqs[nodeID] {
		result, done := ss.result[prereq]
		if !done {
			return false
		}

		if !isPassingStatus(result.Status, result.CacheStatus) {
			return false
		}
	}

	return true
}

// isPassingStatus returns true when the story result represents a
// successful execution (passed or a cache hit). Skipped results from
// failure propagation do NOT count as passing for prerequisite
// satisfaction.
func isPassingStatus(status Status, cacheStatus CacheStatus) bool {
	switch cacheStatus {
	case CacheHitPass, CacheHitSkip:
		return true

	case CacheMiss, CacheBypassed:
		return status == StatusPassed
	}

	// Unreachable: all CacheStatus values handled above.
	return false
}

// propagateFailures marks every dependent of failedNodeID as
// StatusSkipped (recursively) and returns the list of newly-skipped
// node IDs.
//
// The Cause and CauseChain fields are populated per spec 005 §10
// "Failure propagation".
func (ss *schedulerState) propagateFailures(
	failedNodeID string,
	failedResult StoryResult,
) []string {
	var skipped []string

	ss.skipDependents(failedNodeID, failedResult, &skipped)

	return skipped
}

// skipDependents recursively marks all dependents of originID as
// skipped. The originResult is the terminal result of the node
// whose failure triggered the cascade.
func (ss *schedulerState) skipDependents(
	originID string,
	originResult StoryResult,
	skipped *[]string,
) {
	for _, depID := range ss.dependents[originID] {
		if ss.status[depID] != nodePending {
			continue
		}

		causeChain := buildCauseChain(originResult)

		depParts := splitNodeID(depID)
		depStoryID := depParts.storyID
		depScopeKey := depParts.scopeKey

		skipResult := StoryResult{
			Order:       zeroOrder,
			TestID:      depStoryID,
			ScopeKey:    depScopeKey,
			OCPPVersion: "",
			Status:      StatusSkipped,
			CacheStatus: CacheMiss,
			StartedAt:   originResult.FinishedAt,
			FinishedAt:  originResult.FinishedAt,
			Findings: []Finding{
				{
					Message: "skipped: prerequisite " +
						originResult.TestID + " failed",
					Severity: severityError,
				},
			},
			Trace:      nil,
			Cause:      originResult.TestID,
			CauseChain: causeChain,
		}

		ss.status[depID] = nodeDone
		ss.result[depID] = skipResult

		*skipped = append(*skipped, depID)

		// Recurse so that transitive dependents are also skipped.
		ss.skipDependents(depID, skipResult, skipped)
	}
}

// buildCauseChain constructs the transitive cause chain for a newly
// skipped node. The chain starts with the immediate failing
// prerequisite and extends with any prior chain from that result.
func buildCauseChain(originResult StoryResult) []string {
	chain := make([]string, emptySliceLen, 1+len(originResult.CauseChain))

	if originResult.Cause != "" {
		chain = append(chain, originResult.Cause)
		chain = append(chain, originResult.CauseChain...)
	} else {
		// The origin itself failed; it is the root cause.
		chain = append(chain, originResult.TestID)
	}

	return chain
}

// pendingCount returns the number of nodes still in nodePending or
// nodeRunning. When this reaches zero the run is complete.
func (ss *schedulerState) pendingCount() int {
	count := zeroPendingCount

	for _, nodeState := range ss.status {
		if nodeState == nodePending || nodeState == nodeRunning {
			count++
		}
	}

	return count
}
