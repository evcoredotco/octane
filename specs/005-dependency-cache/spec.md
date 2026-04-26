# Spec 005: Dependency Graph and Cache

> **Spec ID:** `005-dependency-cache`
> **Status:** Approved
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCPP scenarios are not independent: a reservation test cannot run
against a CSMS that has not registered the station, and a station
cannot register without first establishing an OCPP-J WebSocket
connection. ADR 0015 specifies the test dependency graph that
captures these prerequisites; ADR 0016 specifies the
content-addressed file tree cache that makes graph traversal
efficient.

This spec implements both. The two are bundled because they share
inputs (the cache key includes the transitively-hashed dependency
chain) and a single review of "the runner" is more coherent than
splitting them.

The runner is the orchestration layer that:

1. Walks a topological ordering of the resolved dependency
   graph for the requested stories.
2. Consults the cache before executing each story; on a hit,
   skips execution and proceeds.
3. On a miss, executes the story by invoking resolved keywords
   against the wire engine (specs 002, 003, 004), captures the
   wire trace, and writes a cache entry.
4. On a story failure, propagates the failure to dependents as
   `skipped` per ADR 0015.
5. Returns a structured `RunResult` that the report emitter
   (spec 007) consumes.

## 2. Goals

- G1. Implement `pkg/runner/` — topological resolver, cycle
      detection, scope-aware traversal (per-station, per-run,
      global per ADR 0015).
- G2. Implement `pkg/cache/` — content-addressed file tree at
      `$XDG_CACHE_HOME/octane/cache/` (overridable via
      `OCTANE_CACHE_DIR`) per ADR 0016.
- G3. Atomic write protocol: temp-file-and-rename for
      `result.json` and `trace.json`, with directory `fsync` for
      crash safety.
- G4. POSIX `flock` per cache key for in-machine concurrency;
      cross-machine concurrency delegated to the CI cache layer
      per ADR 0016.
- G5. Failure propagation: a failed prerequisite causes
      dependents to skip with a finding pointing at the failing
      prerequisite.
- G6. Configurable timeouts (`--lock-timeout`, default 60s) and
      fast-fail (`--no-wait`).

## 3. Non-Goals

- N1. Story parsing (spec 001).
- N2. Wire I/O (spec 002).
- N3. Keyword resolution (spec 003).
- N4. Distributed execution across machines (delegated to CI).
- N5. Cache eviction beyond age-based pruning. LRU and size caps
      are operator-managed via the CI cache layer (GitHub: 10 GB
      per repo; GitLab: configurable).
- N6. Live progress UI (a follow-up; the runner emits structured
      progress events that future tooling can render).

## 4. User Stories

- **As an operator**, I want `octane run scenarios/v16/` to
  resolve every story's prerequisite chain transparently and
  execute prerequisites once even when many stories share them.
- **As a CI maintainer**, I want a green-on-green-cached run to
  complete in seconds, even when the underlying suite would
  take minutes if rerun from scratch.
- **As a parallel-CI user**, I want two parallel jobs running
  disjoint suite partitions to write to disjoint cache paths
  and to merge cleanly when their outputs are restored together
  in a downstream job.
- **As an operator debugging a failure**, I want the report to
  identify the *original* failing prerequisite, not just the
  cascade of skipped dependents.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| II. Two Distribution Surfaces, One Engine | The runner has zero CLI-specific or Action-specific code paths. Both surfaces invoke `runner.Run(cfg)`. |
| IV. Determinism | Topological ordering is stable: ties are broken by lexicographic story ID. Cache key inputs are sorted. |
| V. Stdlib-Heavy | The cache is plain JSON files; no third-party storage library. |
| X. Security | Cache files MUST NOT contain credentials. The runner redacts known sensitive fields (auth headers, idTags marked sensitive in connection profiles) before writing. |

## 6. Acceptance Criteria

- AC1. **Given** a story with a 4-deep `Depends:` chain, **when**
       the runner executes the story for the first time, **then**
       all four prerequisites run in topological order before the
       requested story.
- AC2. **Given** the same story re-executed with no changes,
       **when** the runner consults the cache, **then** every
       prerequisite is a cache hit and execution skips to the
       requested story (which is itself a cache hit).
- AC3. **Given** a `Depends:` graph containing a cycle (story A
       depends on B, B depends on A), **when** the runner runs
       preflight, **then** it returns `ErrCycle` listing the
       offending edges; no execution begins.
- AC4. **Given** a prerequisite that fails on the wire, **when**
       the runner processes the failure, **then** every dependent
       story is marked `skipped` with a finding referencing the
       prerequisite's `test_id` and finding.
- AC5. **Given** a story with `Stations: 2` and a per-station
       prerequisite `station_boot_accepted`, **when** the runner
       walks the graph, **then** the prerequisite executes twice
       (once per station handle) with distinct cache keys.
- AC6. **Given** a prerequisite with scope `per-run`, **when**
       a suite of 50 stories sharing the prerequisite runs, **then**
       the prerequisite executes exactly once.
- AC7. **Given** two `octane run` invocations on the same
       machine racing on the same cache key, **when** the second
       arrives, **then** it acquires `flock` after the first
       releases, re-reads the result, and uses the cached value
       (no double execution).
- AC8. **Given** a CI workflow that restored a partial cache
       (80% of last run's entries), **when** the runner executes,
       **then** it reuses the 80% as cache hits and re-runs only
       the missing 20%.
- AC9. **Given** `octane run --max-parallel 4`, **when** a suite
       of 16 leaf stories with no inter-dependencies runs, **then**
       up to 4 stories execute concurrently, observable as
       overlapping wire traces.
- AC10. **Given** a `Cache-TTL: 1h` Meta key on a helper story
        and a cache entry written 90 minutes ago, **when** the
        runner consults the cache, **then** the entry is treated
        as missing and the helper re-executes.

## 7. OCPP Scope

The runner is OCPP-version-agnostic. Per-version behavior is in
the keyword library; the runner just walks the graph and
invokes keywords.

## 8. Open Questions

- OQ1. Whether to support `--shard 1/4` style suite partitioning
       at the runner level for CI fan-out. Recommendation: yes,
       since it complements the cache design (each shard writes
       to disjoint key paths). Stable shard assignment is by
       `sha256(test_id) mod shard_count`.
       *(owner: Architect, due: with this spec — RESOLVED in
       favor.)*
- OQ2. Whether the cache should distinguish "cached pass" from
       "cached skip" in operator-facing output. Recommendation:
       yes; the report (spec 007) shows a `cache_status` field
       per entry: `hit-pass`, `hit-skip`, `miss`, `bypassed`.
       *(owner: Architect, due: with this spec — RESOLVED.)*

## 9. Out of Scope (parking lot)

- Distributed execution across machines (CI cache layer's job).
- Live progress UI (follow-up tooling, not in this spec).
- Cache LRU eviction (operator-managed via the CI cache layer
  or `octane cache prune`).
- Cache encryption at rest (operator can encrypt the cache
  directory if needed; not OCTANE's concern).

## 10. Implementation notes

### Cache key derivation

The cache key SHA-256 is the lexicographically ordered tuple
defined in ADR 0016. The runner computes this before consulting
the cache; identical-input runs produce identical keys.

### Atomic-write protocol

```
1. Write result.json.tmp in target directory.
2. fsync the temp file.
3. Rename to result.json (atomic on POSIX).
4. fsync the directory entry.
```

The same pattern applies to `trace.json`. A reader sees either
the prior version or the new version, never a torn write.

### Lock acquire pattern

Double-checked locking, per ADR 0016:

```
1. Read result.json. If valid, return cached result.
2. flock(LOCK_EX) on locks/<key_hash>.lock.
3. Re-read result.json. If valid (another runner finished while
   we waited), return cached result.
4. Execute the story.
5. Write trace.json.tmp + atomic-rename.
6. Write result.json.tmp + atomic-rename.
7. Release flock.
```

The double-check handles concurrent local invocations.

### Failure propagation

A story's `RunResult.Status` is one of `passed`, `failed`,
`skipped`. A skipped status carries a `Cause` field naming the
prerequisite whose failure triggered the skip. The transitive
chain of skips is recoverable from the report by walking
`Cause` fields.

### Stable topological ordering

Standard Kahn's algorithm produces a topological order, but ties
within a level (multiple stories whose prerequisites are all
satisfied) must be broken deterministically. The runner sorts
ties by lexicographic story ID. Two runs of the same suite
produce the same execution order.

### Sharding contract

`--shard N/M` partitions stories such that story `s` runs in
shard `i` iff `sha256(s.id)[:8] mod M == i`. The mapping is
stable across runs and platforms; a story moves shards only
when its ID changes.

A sharded run's cache directory contains only the entries for
that shard's stories. Restoration of multiple shard caches in a
downstream job produces the union.

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] DevOps / Platform
- [x] Maintainer review
