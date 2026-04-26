# Tasks 005: Dependency Graph and Cache

> **Spec ID:** `005-dependency-cache`
> **Plan reference:** `./plan.md`
> **Status:** Ready

## Conventions

Same as previous specs.

---

## Phase 1 — DAG contracts

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-01 | Define `dag.Node`, `dag.Edge`, `dag.Graph` types | architect | S | AC1, AC3 | `pkg/runner/internal/dag/dag.go` |
| T-005-02 | Define `runner.Config`, `RunResult`, `StoryResult` | architect | S | AC1 | `pkg/runner/types.go` |
| T-005-03 | Define `runner.Status`, `runner.CacheStatus` enums | architect | P | AC4 | `pkg/runner/types.go` |
| T-005-04 | Define `cache.Cache` interface, `cache.Key`, `cache.Entry` | architect | S | AC2, AC10 | `pkg/cache/cache.go` |
| T-005-05 | Draft ADR 0019 (runner concurrency model) | architect | S | AC9 | `docs/adr/0019-runner-concurrency.md` |

## Phase 2 — DAG implementation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-10 | Implement `dag.AddNode`, `dag.AddEdge` with cycle detection | backend | S | AC3 | `pkg/runner/internal/dag/dag.go` |
| T-005-11 | Implement `dag.TopologicalOrder` with stable tie-breaking | backend | S | AC1 | `pkg/runner/internal/dag/topo.go` |
| T-005-12 | Property test: random graphs → either valid topo order or cycle error | qa | S | AC1, AC3 | `pkg/runner/internal/dag/topo_property_test.go` |

## Phase 3 — Cache implementation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-20 | Implement `cache.Open` (creates directory, version stamp) | backend | S | AC2 | `pkg/cache/open.go` |
| T-005-21 | Implement `cache.Key.Hash()` (SHA-256 of tuple) | backend | P | AC2 | `pkg/cache/key.go` |
| T-005-22 | Implement `cache.Get`, `cache.Put` with atomic temp+rename | backend | S | AC2, AC8 | `pkg/cache/file_tree.go` |
| T-005-23 | Implement TTL invalidation per ADR 0016 | backend | P | AC10 | `pkg/cache/file_tree.go` |
| T-005-24 | Implement `cache.Prune` (age-based filesystem walk) | backend | P | — | `pkg/cache/prune.go` |
| T-005-25 | Atomic-write contract test: torn writes never observed | qa | S | AC2 | `pkg/cache/atomic_test.go` |

## Phase 4 — Lock subsystem

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-30 | Implement Linux/macOS flock via `syscall.Flock` | backend | P | AC7 | `pkg/cache/internal/lock/flock_unix.go` |
| T-005-31 | Implement Windows lock via `LockFileEx` | backend | P | AC7 | `pkg/cache/internal/lock/flock_windows.go` |
| T-005-32 | Implement double-checked acquire pattern | backend | S | AC7 | `pkg/cache/internal/lock/acquire.go` |
| T-005-33 | Cross-platform lock test (matrix CI: Linux + macOS + Windows) | qa | S | AC7 | `pkg/cache/internal/lock/lock_test.go` |

## Phase 5 — Runner implementation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-40 | Implement story → DAG node conversion (uses `pkg/story`) | backend | S | AC1 | `pkg/runner/build_dag.go` |
| T-005-41 | Implement scope-aware traversal (per-station, per-run, global) | backend | S | AC5, AC6 | `pkg/runner/traversal.go` |
| T-005-42 | Implement `runner.Run` main loop (cache lookup → execute → cache write) | backend | S | AC2, AC8 | `pkg/runner/run.go` |
| T-005-43 | Implement failure propagation (skipped dependents) | backend | S | AC4 | `pkg/runner/skip.go` |
| T-005-44 | Implement `--max-parallel` worker pool | backend | S | AC9 | `pkg/runner/parallel.go` |
| T-005-45 | Implement `--shard N/M` partitioning | backend | P | — | `pkg/runner/shard.go` |
| T-005-46 | Implement `--lock-timeout`, `--no-wait` flag plumbing | backend | P | AC7 | `pkg/runner/run.go` |

## Phase 6 — Integration

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-50 | Integration: 4-deep chain runs in topological order | qa | S | AC1 | `test/integration/runner_chain_test.go` |
| T-005-51 | Integration: cached re-run skips execution | qa | S | AC2 | `test/integration/runner_cache_test.go` |
| T-005-52 | Integration: failed prereq skips dependents | qa | S | AC4 | `test/integration/runner_skip_test.go` |
| T-005-53 | Integration: per-station scope (Stations: 2 → 2 prereq runs) | qa | S | AC5 | `test/integration/runner_perstation_test.go` |
| T-005-54 | Integration: per-run scope (50 stories → 1 prereq run) | qa | S | AC6 | `test/integration/runner_perrun_test.go` |
| T-005-55 | Integration: 80%-restored cache behavior | qa | S | AC8 | `test/integration/runner_partial_cache_test.go` |
| T-005-56 | Integration: parallel execution observable in trace | qa | S | AC9 | `test/integration/runner_parallel_test.go` |
| T-005-57 | Integration: cache TTL invalidation | qa | S | AC10 | `test/integration/runner_ttl_test.go` |
| T-005-58 | Stress: 1000-story suite with shared prereq → linear time | qa | P | — | `test/integration/runner_stress_test.go` |

## Phase 7 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-60 | Godoc on every exported symbol | docs | P | — | `pkg/runner/*.go`, `pkg/cache/*.go` |
| T-005-61 | `docs/concepts/dependency-graph.md` | docs | P | — | `docs/concepts/dependency-graph.md` |
| T-005-62 | `docs/concepts/cache.md` | docs | P | — | `docs/concepts/cache.md` |
| T-005-63 | CHANGELOG entry | docs | S | — | `CHANGELOG.md` |

## Phase 8 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-005-70 | Security review: cache redaction, lock-file permissions | security | S | — | — |
| T-005-71 | DevOps review: CI cache integration end-to-end | devops | S | AC8 | — |
| T-005-72 | Reviewer pass: API surface stability | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 10 acceptance criteria covered by at least one task
- [ ] Cross-platform lock test green on Linux/macOS/Windows
- [ ] Stress test completes in linear time
- [ ] ADR 0019 merged
- [ ] Security review signed off
- [ ] DevOps review of CI cache integration signed off
- [ ] CHANGELOG updated under `## [Unreleased]`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/005-dependency-cache` passes
