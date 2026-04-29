// Package runner — T-005-45: shard partitioning.
//
// applyShardFilter removes from the provided story slice any story
// whose test_id does not belong to the current shard. Prerequisites
// required by remaining stories are preserved regardless of their
// shard assignment, per ADR 0019 §"Sharding: --shard N/M".

package runner

import "github.com/evcoreco/octane/pkg/story/ast"

// applyShardFilter returns the subset of stories whose test_id
// belongs to shardIndex (zero-based) out of shardTotal shards.
//
// When shardTotal is zero, the full input slice is returned
// unchanged (sharding disabled).
//
// The algorithm is:
//
//	binary.BigEndian.Uint64(sha256(test_id)[:8]) % shardTotal == shardIndex
//
// Stories NOT in the requested shard but required as prerequisites
// of sharded stories are added back by collectPrerequisites (called
// after this function), satisfying ADR 0019 §"Prerequisite inclusion".
func applyShardFilter(
	stories []*ast.Story,
	shardIndex,
	shardTotal int,
) []*ast.Story {
	if shardTotal <= 0 {
		return stories
	}

	filtered := make([]*ast.Story, 0, len(stories))

	for _, storyAST := range stories {
		if inShardFilter(storyAST.Meta.ID, shardIndex, shardTotal) {
			filtered = append(filtered, storyAST)
		}
	}

	return filtered
}

// collectPrerequisites performs a depth-first walk of the Depends
// chains for all stories in roots and returns the union set of all
// stories (roots + all transitive prerequisites). The allStories
// index maps story ID → *ast.Story for prerequisite lookup.
//
// Stories whose IDs are not found in allStories are silently
// skipped; the DAG builder will surface the missing-story error via
// its own validation pass.
func collectPrerequisites(
	roots []*ast.Story,
	allStories map[string]*ast.Story,
) []*ast.Story {
	seen := make(map[string]bool, len(roots))
	out := make([]*ast.Story, 0, len(roots))

	var walk func(storyAST *ast.Story)

	walk = func(storyAST *ast.Story) {
		if seen[storyAST.Meta.ID] {
			return
		}

		seen[storyAST.Meta.ID] = true

		out = append(out, storyAST)

		for _, dep := range storyAST.Meta.Depends {
			if prereq, ok := allStories[dep.ID]; ok {
				walk(prereq)
			}
		}
	}

	for _, storyAST := range roots {
		walk(storyAST)
	}

	return out
}
