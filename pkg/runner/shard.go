package runner

import "github.com/evcoreco/octane/pkg/story/ast"

// minShardTotal is the minimum valid shard total; values at or below this
// indicate sharding is disabled.
const minShardTotal = 0

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
	if shardTotal <= minShardTotal {
		return stories
	}

	filtered := make([]*ast.Story, emptySliceLen, len(stories))

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
	walker := &prereqWalker{
		seen:       make(map[string]bool, len(roots)),
		out:        make([]*ast.Story, emptySliceLen, len(roots)),
		allStories: allStories,
	}

	for _, storyAST := range roots {
		walker.walk(storyAST)
	}

	return walker.out
}

// prereqWalker performs a depth-first walk over story prerequisites,
// collecting each story exactly once.
type prereqWalker struct {
	seen       map[string]bool
	out        []*ast.Story
	allStories map[string]*ast.Story
}

// walk visits storyAST and its transitive prerequisites once each.
func (pw *prereqWalker) walk(storyAST *ast.Story) {
	if pw.seen[storyAST.Meta.ID] {
		return
	}

	pw.seen[storyAST.Meta.ID] = true
	pw.out = append(pw.out, storyAST)

	for _, dep := range storyAST.Meta.Depends {
		if prereq, ok := pw.allStories[dep.ID]; ok {
			pw.walk(prereq)
		}
	}
}
