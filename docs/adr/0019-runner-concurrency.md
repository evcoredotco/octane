# ADR 0019: Runner Concurrency Model

- **Status:** Accepted
- **Date:** 2026-04-27
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** II (Two Distribution Surfaces,
  One Engine), IV (Determinism), V (Stdlib-Heavy)

## Context

The runner (spec 005) walks a dependency graph of `.story` files in
topological order and executes each story against a live CSMS over the
wire. A naively sequential walk leaves network I/O idle whenever one
story is blocked waiting for a CSMS response while independent stories
could run in parallel.

Operators need a concurrency knob: `--max-parallel N` limits the number
of stories executing concurrently. The default is 1 (sequential), which
preserves backward-compatible, fully deterministic output ordering.
Values above 1 trade strict ordering for throughput.

Concurrency interacts with three other runner subsystems:

| Subsystem | Interaction |
|-----------|-------------|
| **Dependency graph (ADR 0015)** | A story is eligible for execution only after every transitive prerequisite has completed successfully. Parallelism is bounded by the graph's width at any given level, not just by N. |
| **Cache lock protocol (ADR 0016)** | Two goroutines (or two processes) racing on the same cache key must serialize through `flock(LOCK_EX)` with the double-checked acquire pattern. |
| **Sharding (`--shard N/M`)** | CI fan-out partitions the story set across M parallel jobs. Each shard executes an independent subset; the runner's in-process parallelism applies within each shard. |

This ADR formalises the concurrency model so that the backend
implementation (T-005-44), the lock subsystem (T-005-30 through
T-005-33), and the sharding logic (T-005-45) share a single,
documented contract.

## Decision

### Worker-pool model

The runner maintains a pool of at most N goroutines (set by
`--max-parallel N`, default 1). A scheduler goroutine feeds eligible
stories into a buffered work channel; worker goroutines consume from it.

```
                      +-----------+
                      | Scheduler |
                      +-----+-----+
                            |
              eligible stories dispatched
              in stable topological order
                            |
            +-------+-------+-------+-------+
            |       |       |       |       |
          [W1]    [W2]    [W3]    [W4]   ... [WN]
            |       |       |       |
        cache check + execute + cache write
```

A story becomes **eligible** when all of its transitive prerequisites
have a terminal status (passed, failed, or cached-hit). The scheduler
re-evaluates eligibility each time a worker reports completion.

### Dispatch order and determinism

When multiple stories are eligible simultaneously, the scheduler
dispatches them in **lexicographic order by story ID**. This ensures
that, given the same graph and the same N, the set of stories
dispatched in each batch is identical across runs, satisfying
constitution principle IV.

With `--max-parallel 1`, the execution order is identical to the
sequential topological order produced by Kahn's algorithm with
lexicographic tie-breaking. Increasing N allows stories within the same
topological level to overlap, but the dispatch sequence within each
level remains deterministic.

Report output preserves the original topological order regardless of
worker completion order: `StoryResult.Order` is assigned at dispatch
time, not at completion time.

### Eligible-set computation

The scheduler maintains three sets:

| Set | Contents |
|-----|----------|
| `pending` | Stories not yet dispatched. |
| `running` | Stories dispatched to a worker but not yet complete. |
| `done` | Stories with a terminal status (passed, failed, skipped, cached). |

On each tick (triggered by a worker completion or at startup):

1. For each story in `pending`, check whether every prerequisite is in
   `done` with status passed or cached-hit.
2. If a prerequisite is in `done` with status failed, mark the
   dependent as `skipped` (failure propagation per ADR 0015) and move
   it to `done` without dispatching.
3. Collect all newly eligible stories, sort by story ID, and dispatch
   up to `N - len(running)` of them to the work channel.

This loop runs in the scheduler goroutine only; there is no concurrent
mutation of the sets.

### Interaction with the cache lock protocol

Each worker, upon receiving a story, follows the acquire pattern from
ADR 0016 section "Acquire pattern":

1. **Read** `result.json`. If valid (not expired, schema matches),
   report cached-hit to the scheduler. No lock acquired.
2. **Acquire** `flock(LOCK_EX)` on `locks/<key_hash>.lock`. The lock
   call respects `--lock-timeout` (default 60 seconds) and `--no-wait`
   (fail immediately on contention).
3. **Re-read** `result.json`. If valid (another process completed the
   story while we waited), release the lock, report cached-hit.
4. **Execute** the story.
5. **Write** `trace.json` and `result.json` via the atomic
   temp-file-and-rename protocol.
6. **Release** `flock`.

The double-checked acquire prevents two workers (in-process or
cross-process) from executing the same story. In-process, the
`sync.Map[CacheKey]*sync.Once` described in ADR 0016 provides an
additional fast path before the filesystem lock is consulted.

### Lock timeout and fast-fail

| Flag | Default | Behavior |
|------|---------|----------|
| `--lock-timeout` | `60s` | Maximum time a worker blocks on `flock(LOCK_EX)` before returning an error. The error propagates as a story failure and triggers skip cascading for dependents. |
| `--no-wait` | off | Equivalent to `--lock-timeout 0`. The worker calls `flock(LOCK_EX \| LOCK_NB)` and fails immediately if the lock is held. Intended for CI scripts that do not expect concurrent local runs. |

When `--no-wait` is set, the runner assumes it is the sole writer. Any
lock contention indicates a configuration error and the run aborts with
exit code 9.

### Sharding: `--shard N/M`

Sharding partitions the story set for CI fan-out. Story `s` belongs to
shard `i` (zero-indexed) if and only if:

```
binary.BigEndian.Uint64(sha256(s.id)[:8]) % M == i
```

The first 8 bytes of the SHA-256 digest are interpreted as a big-endian
unsigned 64-bit integer. The modulo assignment is stable: a story moves
shards only when its ID changes or when M changes.

Properties:

- **Stable across platforms.** SHA-256 and big-endian integer decoding
  are deterministic. Two machines running different operating systems
  produce the same shard assignment for the same story ID and M.
- **Uniform distribution.** SHA-256 output is uniformly distributed;
  for any M, shard sizes differ by at most one story in expectation.
- **Prerequisite inclusion.** When a sharded run needs a prerequisite
  that is assigned to a different shard, the prerequisite is still
  included in the current shard's execution plan. The prerequisite's
  cache entry, if available from a prior shard's run, provides a hit;
  if not, it executes locally. This avoids cross-shard coordination.
- **Disjoint cache writes.** Each shard writes cache entries only for
  stories it executes. Merging shard cache directories in a downstream
  CI job produces the union of all entries with no conflicts (cache
  keys are content-addressed).

Sharding composes with `--max-parallel`: within a shard, up to N
stories execute concurrently, subject to the dependency graph.

### Goroutine lifecycle

Workers are started when `runner.Run` begins and stopped when the
scheduler closes the work channel (all stories are in `done`). Workers
do not outlive the `Run` call.

Context cancellation (`ctx.Done()`) is propagated to all workers.
A cancelled context causes in-flight stories to abort; their status is
`failed` with an appropriate error. The scheduler then marks all
remaining `pending` stories as `skipped`.

## Consequences

### Positive

- **Throughput scales with graph width.** A suite of 16 independent
  leaf stories with `--max-parallel 4` completes in roughly one quarter
  the wall-clock time of a sequential run (AC9).
- **No non-determinism in reports.** Dispatch order is deterministic;
  report order is the topological order assigned at dispatch time, not
  completion time. Two runs with the same N and the same inputs produce
  the same report ordering.
- **Lock protocol is reused, not reinvented.** The worker's cache
  interaction is identical to the sequential runner's; parallelism
  is layered on top without modifying the lock contract.
- **Sharding is zero-coordination.** Each shard is a self-contained
  `octane run` invocation. No shared state between CI jobs beyond the
  cache directory.
- **Stdlib only.** The worker pool uses `chan`, `sync.WaitGroup`, and
  `context.Context`. No third-party concurrency library is introduced,
  honoring constitution principle V.

### Negative

- **Dependency graph traversal is the scheduling bottleneck.** The
  scheduler is single-goroutine and re-evaluates eligibility on every
  completion. For very deep, narrow graphs (long chains with no
  branching), parallelism provides no benefit because only one story
  is eligible at a time.
- **Shard imbalance is possible.** SHA-256 modulo produces
  statistically uniform shards, but pathological story ID distributions
  could produce uneven partitions. This is unlikely in practice and
  can be addressed by adjusting M.
- **Lock timeout failures cascade.** A timed-out lock acquisition
  marks the story as failed and skips all dependents. Operators must
  tune `--lock-timeout` for their environment (or use `--no-wait` in
  CI where concurrent local runs are not expected).

### Neutral

- `--max-parallel 1` (the default) makes the runner behave identically
  to a sequential implementation. The worker pool collapses to a single
  goroutine with no scheduling overhead beyond the eligibility check.

## Alternatives considered

- **One goroutine per story, bounded by a semaphore.** Simpler to
  implement but harder to reason about dispatch order. The scheduler
  model provides deterministic dispatch ordering, which the semaphore
  model does not.
- **Pipeline model (one stage per topological level).** Overly rigid:
  forces all stories in a level to complete before the next level
  begins, leaving workers idle when levels have uneven story counts.
  The eligibility-based scheduler dispatches the next story as soon as
  its prerequisites are met, regardless of level boundaries.
- **No in-process parallelism; rely on `--shard` for all
  parallelism.** Rejected because sharding requires CI infrastructure.
  A developer running `octane run` locally benefits from in-process
  parallelism without CI.
- **Work-stealing scheduler.** More complex and harder to make
  deterministic. The simple dispatch-from-sorted-eligible model is
  sufficient for the expected suite sizes (hundreds to low thousands
  of stories).

## References

- Constitution: principles II, IV, V
- ADR 0015 (test dependency graph) -- prerequisite resolution,
  failure propagation, topological ordering
- ADR 0016 (cache and lock subsystem) -- flock protocol,
  double-checked acquire, atomic writes
- ADR 0018 (determinism primitives) -- Clock and Rand injection
- Spec 005 (dependency graph and cache) -- AC9, section 10
