package dag

// Topological ordering via Kahn's algorithm.
// This file is intentionally separate from dag.go so that the DAG
// construction and traversal concerns stay in distinct source units.

import (
	"cmp"
	"slices"
)

// emptyInitialLen is the initial length for pre-allocated slices.
const emptyInitialLen = 0

// firstElem is the index of the first element in a slice.
const firstElem = 0

// dropFirst is the start index used to remove the first element.
const dropFirst = 1

// zeroInDegree is the in-degree value that marks a ready node.
const zeroInDegree = 0

// TopologicalOrder returns the nodes of g in a stable topological
// execution order. Ties — multiple nodes whose in-degree drops to
// zero simultaneously — are broken by lexicographic [Node.ID],
// satisfying constitution principle IV (determinism).
//
// If the graph contains a cycle, TopologicalOrder returns a non-nil
// [*CycleError] whose Edges field lists every edge that participates in
// the remaining subgraph. The partial result slice is nil on error.
//
// The algorithm is Kahn's (BFS variant). Time complexity is O(V + E).
func TopologicalOrder(grph *Graph) ([]Node, error) {
	// Work with local copies of in-degree so the original graph is
	// not mutated.
	inDeg := make(map[string]int, len(grph.nodes))

	for _, node := range grph.nodes {
		inDeg[node.ID] = grph.inDegree[node.ID]
	}

	// Seed the ready queue with all nodes whose in-degree is zero,
	// then sort for a stable start state.
	ready := collectZeroInDegree(grph.nodes, inDeg)
	slices.SortFunc(ready, func(a, b Node) int {
		return cmp.Compare(a.ID, b.ID)
	})

	result := make([]Node, emptyInitialLen, len(grph.nodes))

	for len(ready) > zeroInDegree {
		// Pop the lexicographically smallest node.
		current := ready[firstElem]
		ready = ready[dropFirst:]

		result = append(result, current)

		// Reduce in-degree for every dependent of current.
		adjacentLen := len(grph.adjacency[current.ID])
		newReady := make([]Node, emptyInitialLen, adjacentLen)

		for _, neighborID := range grph.adjacency[current.ID] {
			inDeg[neighborID]--

			if inDeg[neighborID] == zeroInDegree {
				idx := grph.nodeIndex[neighborID]
				newReady = append(newReady, grph.nodes[idx])
			}
		}

		// Sort newly eligible nodes before merging so that the
		// overall ordering is stable without a full re-sort.
		slices.SortFunc(newReady, func(a, b Node) int {
			return cmp.Compare(a.ID, b.ID)
		})

		ready = append(ready, newReady...)

		// Re-sort the combined ready queue so that lexicographic
		// tie-breaking is global across all currently eligible nodes,
		// not just newly added ones.
		slices.SortFunc(ready, func(a, b Node) int {
			return cmp.Compare(a.ID, b.ID)
		})
	}

	// If some nodes were not processed, the graph contains a cycle.
	// Report the edges that are still "active" (both endpoints remain
	// unprocessed, i.e. still have non-zero in-degree or were never
	// reached).
	if len(result) < len(grph.nodes) {
		processed := make(map[string]bool, len(result))

		for _, node := range result {
			processed[node.ID] = true
		}

		var cycleEdges []Edge

		for _, edge := range grph.edges {
			if !processed[edge.From] || !processed[edge.To] {
				cycleEdges = append(cycleEdges, edge)
			}
		}

		return nil, &CycleError{Edges: cycleEdges}
	}

	return result, nil
}

// collectZeroInDegree returns the subset of nodes whose in-degree is
// zero according to the provided inDeg map. It preserves the
// caller-supplied slice order.
func collectZeroInDegree(
	nodes []Node,
	inDeg map[string]int,
) []Node {
	out := make([]Node, emptyInitialLen, len(nodes))

	for _, node := range nodes {
		if inDeg[node.ID] == zeroInDegree {
			out = append(out, node)
		}
	}

	return out
}
