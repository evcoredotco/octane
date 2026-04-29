// Package dag — topological ordering via Kahn's algorithm.
//
// This file is intentionally separate from dag.go so that the DAG
// construction and traversal concerns stay in distinct source units.

package dag

import "sort"

// TopologicalOrder returns the nodes of g in a stable topological
// execution order. Ties — multiple nodes whose in-degree drops to
// zero simultaneously — are broken by lexicographic [Node.ID],
// satisfying constitution principle IV (determinism).
//
// If the graph contains a cycle, TopologicalOrder returns a non-nil
// [*ErrCycle] whose Edges field lists every edge that participates in
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
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].ID < ready[j].ID
	})

	result := make([]Node, 0, len(grph.nodes))

	for len(ready) > 0 {
		// Pop the lexicographically smallest node.
		current := ready[0]
		ready = ready[1:]

		result = append(result, current)

		// Reduce in-degree for every dependent of current.
		newReady := make([]Node, 0, len(grph.adjacency[current.ID]))

		for _, neighborID := range grph.adjacency[current.ID] {
			inDeg[neighborID]--

			if inDeg[neighborID] == 0 {
				idx := grph.nodeIndex[neighborID]
				newReady = append(newReady, grph.nodes[idx])
			}
		}

		// Sort newly eligible nodes before merging so that the
		// overall ordering is stable without a full re-sort.
		sort.Slice(newReady, func(i, j int) bool {
			return newReady[i].ID < newReady[j].ID
		})

		ready = append(ready, newReady...)

		// Re-sort the combined ready queue so that lexicographic
		// tie-breaking is global across all currently eligible nodes,
		// not just newly added ones.
		sort.Slice(ready, func(i, j int) bool {
			return ready[i].ID < ready[j].ID
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

		return nil, &ErrCycle{Edges: cycleEdges}
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
	out := make([]Node, 0, len(nodes))

	for _, node := range nodes {
		if inDeg[node.ID] == 0 {
			out = append(out, node)
		}
	}

	return out
}
