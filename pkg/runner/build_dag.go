// Package runner — T-005-40: story → DAG node conversion.
//
// buildDAG converts a slice of parsed story ASTs into a dependency
// graph. Each (story, scope-key) pair becomes one DAG node; edges
// encode the Depends: prerequisites declared in each story's Meta
// section. The function returns ErrCycle (wrapping *dag.ErrCycle)
// when a cycle is detected.

package runner

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/runner/internal/dag"
	"github.com/evcoreco/octane/pkg/story/ast"
)

// storyNode is the internal representation of a single DAG node
// before results are collected. It pairs the original AST with the
// scope-specific execution key used for cache isolation.
type storyNode struct {
	// story is the parsed .story AST.
	story *ast.Story

	// scopeKey is the scope-specific identifier for this node.
	// For per-station scope it is the station handle ("CP01");
	// for per-run scope it is the run ID; for global scope it is
	// the empty string.
	scopeKey string

	// nodeID is the unique DAG node identifier combining the story
	// ID and scope key: "<story_id>/<scope_key>" or simply
	// "<story_id>" when scopeKey is empty.
	nodeID string
}

// makeNodeID builds the canonical DAG node identifier for a story
// and its scope key. The format is "<storyID>/<scopeKey>" when
// scopeKey is non-empty, or simply "<storyID>" for global scope.
func makeNodeID(storyID, scopeKey string) string {
	if scopeKey == "" {
		return storyID
	}

	return storyID + "/" + scopeKey
}

// ErrCycle is returned by [Run] when the dependency graph contains
// a cycle. The underlying *dag.ErrCycle carries the offending edges.
// Callers may use errors.As to extract edge details:
//
//	var cycle *dag.ErrCycle
//	if errors.As(err, &cycle) { ... }
var ErrCycle = errors.New("runner: dependency cycle detected")

// storyIndex maps story ID to its parsed AST for O(1) prerequisite
// lookup during DAG construction.
type storyIndex map[string]*ast.Story

// buildDAGResult is the output of buildDAG.
type buildDAGResult struct {
	// graph is the constructed dependency graph.
	graph *dag.Graph

	// nodes is a slice of storyNode values in the same insertion
	// order as the DAG nodes. Used by the traversal layer.
	nodes []storyNode

	// index maps nodeID → position in nodes for O(1) lookup.
	index map[string]int
}

// scopeKeysFor returns the slice of scope keys that should be
// instantiated for a given dependency scope when the dependent
// story declares stationCount stations.
//
//   - ScopePerStation: one key per station handle ("CP01", "CP02", …).
//   - ScopePerRun: a single fixed key (the runID).
//   - ScopeGlobal: a single empty string.
func scopeKeysFor(scope ast.Scope, stationCount int, runID string) []string {
	switch scope {
	case ast.ScopePerRun:
		return []string{runID}

	case ast.ScopeGlobal:
		return []string{""}

	case ast.ScopePerStation:
		keys := make([]string, stationCount)

		for idx := range stationCount {
			keys[idx] = fmt.Sprintf("CP%02d", idx+1)
		}

		return keys
	}

	// Unreachable: all Scope values handled above.
	return nil
}

// buildDAG constructs the dependency DAG from the provided stories.
//
// Each story is expanded into one or more DAG nodes according to
// the station count declared in its Meta section and the scope
// declared in each of its Depends entries. The resulting graph is
// suitable for topological ordering by dag.TopologicalOrder.
//
// buildDAG returns ErrCycle (wrapping *dag.ErrCycle) when a cycle
// is detected during edge insertion.
func buildDAG(
	stories []*ast.Story,
	runID string,
	stationCountOverride int,
) (*buildDAGResult, error) {
	idx := make(storyIndex, len(stories))

	for _, storyAST := range stories {
		idx[storyAST.Meta.ID] = storyAST
	}

	grph := dag.New()
	nodes := make([]storyNode, 0, len(stories))
	nodeIdx := make(map[string]int, len(stories))

	// Pre-scan: collect story IDs that appear as prerequisites in any
	// Depends entry. These stories must NOT be expanded into standalone
	// per-station nodes in the first pass; instead their scoped nodes
	// are created on demand in the second pass (addDepEdges) to avoid
	// creating both a standalone /CP01 node and a /runID node for the
	// same story when scope = per-run or scope = global.
	prereqIDs := collectPrereqIDs(stories)

	// First pass: create standalone nodes for stories that are NOT
	// referenced as prerequisites. Each such story runs independently,
	// once per station handle declared in its Meta.Stations field.
	for _, storyAST := range stories {
		if prereqIDs[storyAST.Meta.ID] {
			// This story will be instantiated by addDepEdges below
			// with the correct scope (per-station, per-run, or global).
			continue
		}

		stationCount := effectiveStationCount(
			storyAST.Meta.Stations,
			stationCountOverride,
		)

		scopeKeys := scopeKeysFor(
			ast.ScopePerStation,
			stationCount,
			runID,
		)

		for _, scopeKey := range scopeKeys {
			nid := makeNodeID(storyAST.Meta.ID, scopeKey)

			if _, exists := nodeIdx[nid]; exists {
				continue
			}

			storyNodeVal := storyNode{
				story:    storyAST,
				scopeKey: scopeKey,
				nodeID:   nid,
			}

			grph.AddNode(dag.Node{ID: nid})

			nodeIdx[nid] = len(nodes)
			nodes = append(nodes, storyNodeVal)
		}
	}

	// Second pass: add edges for each story's declared Depends.
	for _, storyAST := range stories {
		stationCount := effectiveStationCount(
			storyAST.Meta.Stations,
			stationCountOverride,
		)

		dependentScopeKeys := scopeKeysFor(
			ast.ScopePerStation,
			stationCount,
			runID,
		)

		for _, depScopeKey := range dependentScopeKeys {
			dependentNID := makeNodeID(storyAST.Meta.ID, depScopeKey)

			// Ensure the dependent node exists only when this story has
			// dependencies. A story with no Depends skipped in the first
			// pass (because it is a prereq of another story) should NOT
			// be recreated here with its per-station scope key — it will
			// be instantiated by addDepEdges with the correct scope when
			// a dependent story references it.
			if len(storyAST.Meta.Depends) > 0 {
				ensureNodeExists(
					dependentNID,
					storyAST,
					depScopeKey,
					grph,
					&nodes,
					nodeIdx,
				)
			}

			for _, dep := range storyAST.Meta.Depends {
				err := addDepEdges(
					dep,
					dependentNID,
					idx,
					stationCountOverride,
					runID,
					grph,
					&nodes,
					nodeIdx,
				)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &buildDAGResult{
		graph: grph,
		nodes: nodes,
		index: nodeIdx,
	}, nil
}

// effectiveStationCount returns stationCount or stationCountOverride
// when the override is positive, enforcing a minimum of 1.
func effectiveStationCount(stationCount, override int) int {
	count := stationCount

	if override > 0 {
		count = override
	}

	if count < 1 {
		count = 1
	}

	return count
}

// addDepEdges resolves one Depends entry for a given dependent node
// and inserts the corresponding edges into grph. Prerequisite nodes
// that do not yet exist in nodeIdx are created on demand.
func addDepEdges(
	dep ast.Dependency,
	dependentNID string,
	idx storyIndex,
	stationCountOverride int,
	runID string,
	grph *dag.Graph,
	nodes *[]storyNode,
	nodeIdx map[string]int,
) error {
	prereqStory, ok := idx[dep.ID]
	if !ok {
		// Prerequisite not in the loaded set; the validator catches
		// this. Skip silently to avoid a nil-deref.
		return nil
	}

	// For per-station scope, the prerequisite runs once per station of
	// the DEPENDENT story (not once per station of the prereq itself).
	// The dependent's scope key encodes the station handle for this
	// particular instance, so we mirror it onto the prereq.
	var prereqScopeKeys []string

	if dep.Scope == ast.ScopePerStation {
		_, depScopeKey := splitNodeID(dependentNID)
		prereqScopeKeys = []string{depScopeKey}
	} else {
		prereqStationCount := effectiveStationCount(
			prereqStory.Meta.Stations,
			stationCountOverride,
		)

		prereqScopeKeys = scopeKeysFor(dep.Scope, prereqStationCount, runID)
	}

	for _, prereqScopeKey := range prereqScopeKeys {
		prereqNID := makeNodeID(dep.ID, prereqScopeKey)

		ensureNodeExists(
			prereqNID,
			prereqStory,
			prereqScopeKey,
			grph,
			nodes,
			nodeIdx,
		)

		edge := dag.Edge{From: prereqNID, To: dependentNID}

		err := grph.AddEdge(edge)
		if err != nil {
			var errCycle *dag.ErrCycle
			if errors.As(err, &errCycle) {
				return fmt.Errorf("%w: %w", ErrCycle, errCycle)
			}

			return fmt.Errorf(
				"runner: add edge %s→%s: %w",
				prereqNID,
				dependentNID,
				err,
			)
		}
	}

	return nil
}

// ensureNodeExists adds a storyNode to grph and nodeIdx when it
// does not yet exist.
func ensureNodeExists(
	nid string,
	storyAST *ast.Story,
	scopeKey string,
	grph *dag.Graph,
	nodes *[]storyNode,
	nodeIdx map[string]int,
) {
	if _, exists := nodeIdx[nid]; exists {
		return
	}

	storyNodeVal := storyNode{
		story:    storyAST,
		scopeKey: scopeKey,
		nodeID:   nid,
	}

	grph.AddNode(dag.Node{ID: nid})

	nodeIdx[nid] = len(*nodes)
	*nodes = append(*nodes, storyNodeVal)
}

// inShardFilter returns true when the story should be included in
// the current shard, or when sharding is disabled (shardTotal == 0).
//
// The filter implements spec 005 §10 "Sharding contract":
//
//	binary.BigEndian.Uint64(sha256(test_id)[:8]) % shardTotal == shardIndex
func inShardFilter(testID string, shardIndex, shardTotal int) bool {
	if shardTotal <= 0 {
		return true
	}

	digest := sha256Sum(testID)
	shardNum := binary.BigEndian.Uint64(digest[:8])

	// Modulo result is in [0, shardTotal), which fits in int since
	// shardTotal is a positive int. The conversion is safe.
	assignedShard := int(
		shardNum % uint64(shardTotal),
	)

	return assignedShard == shardIndex
}

// sha256Sum returns the SHA-256 digest of text.
// Separated from inShardFilter to keep cyclomatic complexity low.
func sha256Sum(text string) [32]byte {
	return sha256Of([]byte(text))
}

// collectPrereqIDs returns the set of story IDs that appear as
// prerequisites in any Depends entry within the given story collection.
// The set is used by buildDAG to skip standalone node creation for
// prerequisite stories; their scoped instances are created on-demand
// by addDepEdges with the correct scope key.
func collectPrereqIDs(stories []*ast.Story) map[string]bool {
	prereqs := make(map[string]bool)

	for _, storyAST := range stories {
		for _, dep := range storyAST.Meta.Depends {
			prereqs[dep.ID] = true
		}
	}

	return prereqs
}
