// Package runner orchestrates the execution of .story files by
// walking the test dependency graph (ADR 0015), consulting the
// content-addressed cache (ADR 0016), and returning structured
// results that the report emitter (spec 007) consumes.
//
// # Entry point
//
// [Run] is the single public entry point. Both distribution surfaces
// (CLI and GitHub Action) construct a [Config] and call Run; there
// are no surface-specific code paths in this package (constitution
// principle II).
//
// # Worker pool
//
// Run maintains a pool of at most Config.MaxParallel goroutines.
// A scheduler goroutine dispatches eligible stories in stable
// topological order; workers execute stories and report completion.
// See ADR 0019 for the full concurrency model.
//
// # Scope-aware traversal
//
// Stories declare prerequisite scope (per-station, per-run, global)
// via the Depends: block. The runner expands scoped prerequisites
// into distinct execution nodes before building the DAG so that
// per-station prerequisites execute once per station handle (AC5)
// and per-run prerequisites execute exactly once across the suite
// (AC6).
//
// # Failure propagation
//
// When a story fails, all of its dependents (and their transitive
// dependents) are marked StatusSkipped. The Cause and CauseChain
// fields on StoryResult identify the original failure (AC4).
//
// # Sharding
//
// When Config.ShardTotal > 0, only stories whose
// sha256(test_id)[:8] mod ShardTotal == ShardIndex are executed
// (spec 005 OQ1). Prerequisites outside the shard are still
// included and may produce cache hits from prior shard runs.
//
// # Non-responsibilities
//
// This package does NOT parse .story files (that is pkg/story),
// emit report artefacts (that is pkg/report, spec 007), manage
// WebSocket sessions (that is pkg/transport), or validate OCPP
// payload fields (that is github.com/evcoreco/ocpp16messages).
package runner
