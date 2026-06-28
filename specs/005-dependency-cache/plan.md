# Plan 005: Dependency Graph and Cache

> **Spec ID:** `005-dependency-cache`
> **Status:** Approved
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement the runner that walks the test dependency graph (per
ADR 0015), and the cache that makes graph walks efficient (per
ADR 0016). Bundled into one spec because their inputs and output
schemas are interlocked.

The runner exposes a single `Run(ctx, cfg) (*RunResult, error)`
function that the CLI (spec 006) and Action (spec 006) invoke.

## 2. Architecture Touchpoints

- `pkg/runner/` — new; resolver, cycle detection, scope-aware traversal
- `pkg/runner/internal/dag/` — new; DAG implementation, topological sort
- `pkg/cache/` — new; content-addressed file tree per ADR 0016
- `pkg/cache/internal/lock/` — new; flock-based per-key locks
- Read-only consumers: `pkg/story`, `pkg/keywords/registry`,
  `pkg/transport`, `pkg/engine/clock`
- `pkg/runner.RunResult` — new; the data model that spec 007
  consumes

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `pkg/runner.Run(ctx, cfg) (*RunResult, error)` | new | initial |
| `pkg/runner.Config` | new struct | initial |
| `pkg/runner.RunResult` | new struct | initial |
| `pkg/runner.StoryResult` | new struct | initial |
| `pkg/runner.Status` (passed/failed/skipped/bypassed) | new enum | initial |
| `pkg/runner.CacheStatus` (hit-pass/hit-skip/miss/bypassed) | new enum | initial |
| `pkg/runner.ErrCycle` | new typed error | initial |
| `pkg/cache.Cache` interface | new | initial |
| `pkg/cache.Open(dir string) (Cache, error)` | new constructor | initial |
| `pkg/cache.Key` | new struct | initial |
| `pkg/cache.Entry` | new struct | initial |

## 4. Data Contracts

### Cache key SHA-256

Per ADR 0016: SHA-256 of the colon-joined tuple of `test_id`,
`scope_key`, `csms_endpoint_sha`, `octane_version`,
`ocpp_version`, `story_content_sha`, `parameter_sha`.

### Cache file paths

```
<cache-dir>/results/<key[0:2]>/<key>/result.json
<cache-dir>/results/<key[0:2]>/<key>/trace.json
<cache-dir>/locks/<key>.lock
```

### result.json schema

The schema in ADR 0016 -"Result file schema". `schema_version: 1`.

### RunResult shape

```go
type RunResult struct {
    RunID      string         // ULID
    StartedAt  time.Time
    FinishedAt time.Time
    Stories    []StoryResult  // sorted by Order field
    Summary    Summary
}

type StoryResult struct {
    Order        int
    TestID       string
    ScopeKey     string
    OCPPVersion  string
    Status       Status
    CacheStatus  CacheStatus
    StartedAt    time.Time
    FinishedAt  time.Time
    Findings     []Finding
    Trace        *Trace  // nil if --no-trace-on-pass
    Cause        string  // for skipped: TestID/ScopeKey of failing prereq
    CauseChain   []string // transitive cause chain
}
```

## 5. Required ADRs

- [x] ADR 0015 — Test dependency graph
- [x] ADR 0016 — Cache and lock subsystem (content-addressed file tree)
- [ ] **ADR 0019** (new) — Runner concurrency model: how
      `--max-parallel` interacts with the DAG. Drafted alongside
      this spec.

## 6. Test Strategy

- **Unit tests** of `pkg/runner/internal/dag/`: cycle detection,
  topological order stability, scope-aware traversal
  (per-station/per-run/global).
- **Unit tests** of `pkg/cache/`: atomic-write contract, lock
  acquire pattern, double-checked locking under simulated
  contention.
- **Integration tests** against the pinned CitrineOS:
  - First-run miss: AC1 (4-deep chain executes in order)
  - Second-run hit: AC2 (every prereq is a cache hit)
  - Failure propagation: AC4 (skipped dependents)
  - Per-station scope: AC5 (prereq runs twice for `Stations: 2`)
  - Per-run scope: AC6 (prereq runs once across 50 stories)
  - Concurrent local runs: AC7 (flock serialization)
  - Partial cache restoration: AC8 (80%-restored cache reuses
    80%, re-runs 20%)
  - TTL invalidation: AC10
- **Stress test**: a synthetic suite of 1000 stories with a
  shared prereq; assert prereq runs once and the suite completes
  in linear time.

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** N/A.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cycle detection misses indirect cycles | Medium | High | Property test with random-graph fixtures; cycles are detected at preflight, not runtime |
| flock semantics differ between Linux/macOS/Windows | Medium | High | Cross-platform integration tests in CI matrix; abstract behind an interface |
| Atomic-rename not actually atomic on Windows | Low | High | Use `os.Rename` (which uses `MoveFileEx` on Windows); test with concurrent writers |
| Cache key derivation changes between versions silently | Medium | High | The key includes `octane_version`; cache-key changes are forced to invalidate by definition |
| Determinism leak through map iteration in serializer | Medium | High | Same mitigation as spec 001: never serialize maps directly |
| `--max-parallel` semantics interact badly with per-station scope | Medium | Medium | Document the interaction in ADR 0019; integration test covers both |

## 9. Effort Estimate

- T-shirt size: **L**
- Calendar estimate: 2–3 weeks of focused work
- Parallelizable streams: dag + cache are independent; once both
  land, the runner glue is fast

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] DevOps / Platform (cross-platform flock testing)
- [x] Maintainer review
