// Package dag implements a directed acyclic graph for story dependency
// resolution. The graph enforces acyclicity at edge-insertion time and
// produces a stable topological ordering with ties broken by
// lexicographic node ID (constitution principle IV: determinism).
//
// All collection fields use slices rather than maps to guarantee
// deterministic iteration and serialization order.
package dag

import (
	"errors"
	"fmt"
	"strings"
)

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

// ErrUnknownNode is returned by [Graph.AddEdge] when either endpoint
// of the requested edge has not been registered via [Graph.AddNode].
var ErrUnknownNode = errors.New("dag: unknown node ID")

// AddNode inserts a vertex into the graph. If a node with the same
// ID already exists, AddNode is a no-op.
func (g *Graph) AddNode(node Node) {
	if _, exists := g.nodeIndex[node.ID]; exists {
		return
	}

	g.nodeIndex[node.ID] = len(g.nodes)
	g.nodes = append(g.nodes, node)
	g.adjacency[node.ID] = nil
}

// AddEdge inserts a directed dependency edge from a prerequisite to a
// dependent. Both endpoint node IDs must have been added via [AddNode]
// before calling AddEdge.
//
// AddEdge returns [ErrUnknownNode] (wrapped) when either endpoint has
// not been registered. AddEdge returns [*CycleError] if the new edge
// would introduce a cycle. On any error the graph is left unchanged.
func (g *Graph) AddEdge(edge Edge) error {
	if _, ok := g.nodeIndex[edge.From]; !ok {
		return fmt.Errorf("from %q: %w", edge.From, ErrUnknownNode)
	}

	if _, ok := g.nodeIndex[edge.To]; !ok {
		return fmt.Errorf("to %q: %w", edge.To, ErrUnknownNode)
	}

	// Detect whether adding From→To would create a cycle by checking
	// whether To can already reach From via existing edges (DFS).
	if cyclePath := g.reachablePath(edge.To, edge.From); cyclePath != nil {
		// Build the cycle edge list: the new edge first, then the
		// existing path back, so the sequence closes the loop.
		cycleEdges := make([]Edge, emptyInitialLen, len(cyclePath)+1)
		cycleEdges = append(cycleEdges, edge)
		cycleEdges = append(cycleEdges, cyclePath...)

		return &CycleError{Edges: cycleEdges}
	}

	g.edges = append(g.edges, edge)
	g.adjacency[edge.From] = append(g.adjacency[edge.From], edge.To)
	g.inDegree[edge.To]++

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

// reachablePath returns the sequence of edges that form a path from
// start to target using a depth-first search over existing adjacency.
// It returns nil when no path exists. The returned slice represents
// the edges along the path from start toward target, in traversal
// order.
func (g *Graph) reachablePath(start, target string) []Edge {
	visited := make(map[string]bool, len(g.nodes))

	return g.dfsPath(start, target, visited)
}

// dfsPath is the recursive helper for [Graph.reachablePath]. It
// returns the edge path from current to target, or nil if unreachable.
func (g *Graph) dfsPath(
	current, target string,
	visited map[string]bool,
) []Edge {
	if current == target {
		return []Edge{}
	}

	visited[current] = true

	for _, neighbor := range g.adjacency[current] {
		if visited[neighbor] {
			continue
		}

		edge := Edge{From: current, To: neighbor}

		if tail := g.dfsPath(neighbor, target, visited); tail != nil {
			return append([]Edge{edge}, tail...)
		}
	}

	return nil
}

// CycleError is returned by [Graph.AddEdge] when inserting an edge would
// create a cycle in the dependency graph. The Edges field lists every
// edge that participates in the cycle, enabling callers to produce
// actionable diagnostic output identifying the offending dependencies.
//
// Callers should use [errors.As] to extract the typed error:
//
//	var cycle *dag.CycleError
//	if errors.As(err, &cycle) {
//	    for _, e := range cycle.Edges {
//	        fmt.Printf("%s -> %s\n", e.From, e.To)
//	    }
//	}
type CycleError struct {
	// Edges lists the directed edges that form the cycle, in
	// traversal order. The last edge's To field equals the first
	// edge's From field, closing the loop.
	Edges []Edge
}

// Error returns a human-readable description of the cycle, listing
// every participating edge.
func (e *CycleError) Error() string {
	if len(e.Edges) == zeroInDegree {
		return "dag: cycle detected (no edge details available)"
	}

	msg := "dag: cycle detected: "

	var msgSb strings.Builder

	for i, edge := range e.Edges {
		if i > zeroInDegree {
			msgSb.WriteString(" -> ")
		}

		msgSb.WriteString(edge.From)
	}

	msg += msgSb.String()
	msg += " -> " + e.Edges[0].From

	return msg
}
