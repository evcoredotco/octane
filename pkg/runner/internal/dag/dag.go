// Package dag implements a directed acyclic graph for story dependency
// resolution. The graph enforces acyclicity at edge-insertion time and
// produces a stable topological ordering with ties broken by
// lexicographic node ID (constitution principle IV: determinism).
//
// All collection fields use slices rather than maps to guarantee
// deterministic iteration and serialization order.
package dag

import "fmt"

// Node is a vertex in the dependency graph. Each node corresponds to
// a single story identified by its stable snake_case ID.
type Node struct {
	// ID is the stable snake_case story identifier. It matches the
	// ID Meta key of the corresponding .story file and serves as the
	// unique key within a [Graph].
	ID string
}

// Edge is a directed dependency edge from a prerequisite node to a
// dependent node. The direction encodes "From must complete before
// To may begin."
type Edge struct {
	// From is the ID of the prerequisite node (the dependency).
	From string
	// To is the ID of the dependent node (the story that requires
	// the prerequisite).
	To string
}

// Graph is the dependency graph container. It holds an ordered set of
// nodes and directed edges, supporting topological traversal and
// cycle detection.
//
// A Graph must be created via [New]; the zero value is not usable.
type Graph struct {
	// nodes stores every added node in insertion order. Keyed
	// lookups go through the nodeIndex map; iteration uses this
	// slice to preserve deterministic ordering.
	nodes []Node

	// nodeIndex maps node ID to the index in the nodes slice,
	// providing O(1) membership checks. This map is internal
	// bookkeeping and is never serialized.
	nodeIndex map[string]int

	// edges stores every directed edge in insertion order.
	edges []Edge

	// adjacency maps each node ID to the list of IDs it has
	// outgoing edges to (its dependents). Used by the topological
	// sort and cycle detection algorithms.
	adjacency map[string][]string

	// inDegree tracks the number of incoming edges for each node
	// ID. Used by Kahn's algorithm for topological sorting.
	inDegree map[string]int
}

// New creates an empty [Graph] ready for use. Callers add vertices
// with [Graph.AddNode] and directed edges with [Graph.AddEdge].
func New() *Graph {
	return &Graph{
		nodes:     nil,
		nodeIndex: make(map[string]int),
		edges:     nil,
		adjacency: make(map[string][]string),
		inDegree:  make(map[string]int),
	}
}

// AddNode inserts a vertex into the graph. If a node with the same
// ID already exists, AddNode is a no-op.
//
// Implementation is provided by T-005-10.
func (g *Graph) AddNode(n Node) {
	_ = n // Stub: implementation in T-005-10.
}

// AddEdge inserts a directed dependency edge from a prerequisite to a
// dependent. Both endpoint node IDs must have been added via [AddNode]
// before calling AddEdge.
//
// AddEdge returns [ErrCycle] if the new edge would introduce a cycle.
// On error the graph is left unchanged.
//
// Implementation is provided by T-005-10.
func (g *Graph) AddEdge(e Edge) error {
	_ = e // Stub: implementation in T-005-10.

	return nil
}

// Nodes returns a copy of the graph's nodes in insertion order.
func (g *Graph) Nodes() []Node {
	out := make([]Node, len(g.nodes))
	copy(out, g.nodes)

	return out
}

// Edges returns a copy of the graph's edges in insertion order.
func (g *Graph) Edges() []Edge {
	out := make([]Edge, len(g.edges))
	copy(out, g.edges)

	return out
}

// ErrCycle is returned by [Graph.AddEdge] when inserting an edge would
// create a cycle in the dependency graph. The Edges field lists every
// edge that participates in the cycle, enabling callers to produce
// actionable diagnostic output identifying the offending dependencies.
//
// Callers should use [errors.As] to extract the typed error:
//
//	var cycle *dag.ErrCycle
//	if errors.As(err, &cycle) {
//	    for _, e := range cycle.Edges {
//	        fmt.Printf("%s -> %s\n", e.From, e.To)
//	    }
//	}
type ErrCycle struct {
	// Edges lists the directed edges that form the cycle, in
	// traversal order. The last edge's To field equals the first
	// edge's From field, closing the loop.
	Edges []Edge
}

// Error returns a human-readable description of the cycle, listing
// every participating edge.
func (e *ErrCycle) Error() string {
	if len(e.Edges) == 0 {
		return "dag: cycle detected (no edge details available)"
	}

	msg := "dag: cycle detected: "

	for i, edge := range e.Edges {
		if i > 0 {
			msg += " -> "
		}

		msg += edge.From
	}

	msg += fmt.Sprintf(" -> %s", e.Edges[0].From)

	return msg
}
