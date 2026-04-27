# ADR 0016: Cache and Lock Subsystem — Content-Addressed File Tree

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** II (Two Distribution Surfaces,
  One Engine), IV (Determinism), V (Stdlib-Heavy), X (Security)

## Context

ADR 0015 establishes a test dependency graph where every story can
be a prerequisite for others. To avoid redundant execution and to
support cross-process operator workflows, OCTANE caches test results
across `octane run` invocations.

The cache must work correctly in two distinct deployment scenarios:

| Scenario | Constraints |
|----------|-------------|
| **Local CLI** | One operator, one machine, occasional concurrent runs. POSIX file locks (`flock`) are sufficient for cross-process safety. |
| **CI (GitHub Actions / GitLab CI)** | Cache directory restored from a prior job's tarball at job start; uploaded as a tarball at job end. Multiple parallel jobs may write to the same cache key concurrently across machines that do not share a filesystem. |

A previous draft of this ADR specified SQLite as the cache backend.
That choice was reconsidered after the CI deployment scenario was
re-examined: SQLite's strengths (in-database concurrency, ACID
transactions on a single file) work *against* the CI cache model,
where:

- Two parallel CI jobs writing to the same cache key race; whichever
  uploads its tarball last wins, silently overwriting the other.
- A single SQLite file compresses poorly compared to JSON text,
  inflating CI cache upload/download bandwidth.
- Partial cache restoration leaves a corrupt SQLite file; partial
  restoration of a file tree leaves a usable subset.

The pattern that does work for CI caching — and that comparable
tools (Bazel disk cache, ccache, Go build cache, Cargo) have
converged on — is a **content-addressed file tree**.

## Decision

OCTANE's cache is a content-addressed directory tree. Each cache
entry is one or more JSON files at a path derived from the SHA-256
of the cache key tuple. There is no database.

### Layout

```
$XDG_CACHE_HOME/octane/cache/
├── results/
│   ├── ab/
│   │   └── ab12cd34ef56.../              # full SHA-256 of cache key
│   │       ├── result.json               # small, always present
│   │       └── trace.json                # large, optional
│   ├── cd/
│   │   └── cd56ef78ab90.../
│   │       └── result.json
│   └── ...
├── locks/
│   └── ab12cd34ef56.lock                 # POSIX flock target
└── meta/
    ├── octane-version                    # plain text
    └── created-at                        # plain text RFC 3339
```

The two-character fanout under `results/` is the standard pattern
used by Bazel, ccache, Go's build cache, and Git. It bounds
directory size and keeps `ls` output readable.

### Cache key derivation

Each cache entry is keyed by the SHA-256 of the lexicographically
ordered tuple:

```
test_id || ":" || scope_key || ":" || csms_endpoint_sha ||
  ":" || octane_version || ":" || ocpp_version ||
  ":" || story_content_sha || ":" || parameter_sha
```

Where:

| Field | Source |
|-------|--------|
| `test_id` | story `Id` Meta key |
| `scope_key` | station handle (`per-station`), run ID (`per-run`), or empty (`global`) per ADR 0015 |
| `csms_endpoint_sha` | SHA-256 of (URL + subprotocol + auth-mode tuple) |
| `octane_version` | from build info |
| `ocpp_version` | from story Meta or config |
| `story_content_sha` | SHA-256 of story file + transitively all prerequisites' content |
| `parameter_sha` | SHA-256 of bound parameters |

Any change to any field produces a new cache key and therefore a
new path; old entries become unreachable and are pruned by age.

### Result file schema

Every cache entry includes `result.json`:

```json
{
  "schema_version": 1,
  "test_id": "connector_reservation_faulted",
  "scope_key": "CP01",
  "key_hash": "ab12cd34ef56...",
  "octane_version": "0.1.0",
  "ocpp_version": "1.6",
  "csms_endpoint_sha": "9f8a7b6c5d4e...",
  "story_content_sha": "1a2b3c4d5e6f...",
  "parameter_sha": "f1e2d3c4b5a6...",
  "started_at": "2026-04-26T08:00:00.123Z",
  "finished_at": "2026-04-26T08:00:01.456Z",
  "status": "passed",
  "ttl_seconds": 3600,
  "findings": [],
  "trace_present": true,
  "trace_byte_count": 12873
}
```

`status` is one of `passed`, `failed`, `skipped` (per ADR 0015).
`trace_present` indicates whether a sibling `trace.json` exists.

### Wire trace splitting

Wire traces are split into a sibling `trace.json` file in the same
directory. The split applies to **all** cache entries (no
configuration); the rationale is that report rendering frequently
needs `result.json` (status, finding count, timing) without needing
the full wire frame log.

```
ab/ab12cd34ef56.../
├── result.json    # 1–2 KB typical
└── trace.json     # 0 KB to several MB
```

`trace.json` contains the captured OCPP-J frames in order:

```json
{
  "schema_version": 1,
  "key_hash": "ab12cd34ef56...",
  "frames": [
    {
      "direction": "in",
      "captured_at": "2026-04-26T08:00:00.500Z",
      "raw": [2, "msg-1", "ReserveNow", { "connectorId": 1, "idTag": "VID:0001" }]
    },
    {
      "direction": "out",
      "captured_at": "2026-04-26T08:00:01.100Z",
      "raw": [3, "msg-1", { "status": "Faulted" }]
    }
  ]
}
```

A failed test always writes its trace. A passing test writes its
trace by default; `--no-trace-on-pass` suppresses passing-test
traces to reduce cache size when CI bandwidth is constrained.

### Atomic writes

Cache writes use the standard temp-file-and-rename pattern:

1. Write `result.json.tmp` (or `trace.json.tmp`) into the target
   directory.
2. `fsync` the temp file.
3. `rename` it to its final name. POSIX guarantees this is atomic
   on the same filesystem.
4. `fsync` the directory entry.

A reader that opens a result file will either see the full prior
version or the full new version, never a torn write.

### Lock protocol

#### In-process

A `sync.Map[CacheKey]*sync.Once` ensures each (test_id, scope_key)
runs at most once per `octane run` invocation.

#### Cross-process (same machine only)

A `flock`-based exclusive lock on `locks/<key_hash>.lock` prevents
two `octane run` invocations on the same machine from racing on
the same cache key. The lock file is separate from the result file
so the result file's atomic-rename is unencumbered.

- Linux/macOS: `syscall.Flock(fd, syscall.LOCK_EX)`
- Windows: `LockFileEx(handle, LOCKFILE_EXCLUSIVE_LOCK, ...)`

POSIX advisory locks release on process exit, so crashed runners
do not leave permanent stale locks.

#### Cross-machine (CI)

Cross-machine concurrency is **not** managed by OCTANE. It is
managed by the CI cache layer:

- GitHub Actions and GitLab CI both serialize cache uploads per
  cache key at the platform level. The "loser" of a race uploads
  to a different (typically time-suffixed) key; the next job's
  cache restoration picks up the most-recently-saved key matching
  its prefix.
- Within a single CI job's runner, OCTANE's flock-based
  cross-process safety still applies as a defense-in-depth measure
  even though concurrent invocations within one job are unusual.

**This is intentional.** OCTANE is not a distributed system; the
cache is local-only. CI cache infrastructure is the right layer to
solve cross-machine concurrency, and it does.

### Acquire pattern

For each cache key:

```
1. Compute key_hash = sha256(cache key tuple).
2. Compute path = results/<key_hash[:2]>/<key_hash>/result.json
3. Try to read result.json. If it exists and is valid (TTL not
   expired, schema_version matches), use it. END.
4. Acquire flock on locks/<key_hash>.lock (exclusive).
5. Re-read result.json. If it exists and is valid, use it. END.
6. Execute the test, capture result and trace.
7. Write trace.json.tmp + atomic-rename trace.json (if not skipped
   per --no-trace-on-pass).
8. Write result.json.tmp + atomic-rename result.json.
9. Release flock.
10. END.
```

The double-read handles the case where another local runner
completed the test while we waited for the lock.

### Wait or fail

By default, a runner that finds a held lock waits up to
`--lock-timeout` (default `60s`) before failing with exit 9.

A `--no-wait` flag fails immediately on lock contention (intended
for CI scripts that don't expect concurrent local runs).

### Operator surface

```
octane cache info     # cache directory, total size, entry count
octane cache prune    # remove entries older than --max-age (default 30d)
                      # or with TTL exceeded
octane cache clear    # remove all cache content
octane cache key <story> [--scope STATION]  # print the resolved
                                             # cache key + path
octane cache show <key>                      # print result.json
octane cache trace <key>                     # print trace.json
```

`octane cache prune` is implemented as a filesystem walk: any
result file whose `finished_at + ttl_seconds < now` or whose
mtime is older than `--max-age` is deleted along with its trace.
Empty fanout directories are removed.

## CI integration

### Cache directory contract

OCTANE reads `OCTANE_CACHE_DIR`. If unset, falls back to the
XDG/OS-default location. CI workflows set this environment variable
to a path the CI cache action is configured to persist.

### GitHub Actions example

```yaml
name: ocpp-conformance
on: [push, pull_request]

jobs:
  conformance:
    runs-on: ubuntu-latest
    env:
      OCTANE_CACHE_DIR: ${{ github.workspace }}/.octane-cache
    steps:
      - uses: actions/checkout@v4

      - name: Restore OCTANE cache
        uses: actions/cache@v4
        with:
          path: ${{ env.OCTANE_CACHE_DIR }}
          # Key changes when OCTANE version, scenarios, or
          # connection profile change. Suite content drives most
          # invalidation; story_content_sha inside cache entries
          # handles per-test invalidation.
          key: octane-${{ runner.os }}-${{ hashFiles('scenarios/**', 'connections/**', 'octane.yml') }}
          restore-keys: |
            octane-${{ runner.os }}-

      - name: Run OCTANE
        uses: evcoreco/octane-action@v0
        with:
          stories: scenarios/v16/
          fail-on: major

      - name: Upload reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: octane-report
          path: reports/
```

Notes:

- The cache `key` includes the suite content hash so that scenario
  edits invalidate the whole cache. OCTANE's per-entry validity
  check then ensures stale entries inside a partly-restored cache
  are not used.
- `restore-keys:` enables partial cache hits: a new run reuses any
  previous cache for the same OS, then OCTANE invalidates entries
  that no longer match. This is exactly the file-tree's strength.
- For matrix jobs (multiple OCPP versions, multiple CSMSes), use
  one cache key per matrix dimension or accept that all matrix
  runs share one cache and rely on per-entry SHAs for correctness.

### GitLab CI example

```yaml
stages:
  - conformance

variables:
  OCTANE_CACHE_DIR: ${CI_PROJECT_DIR}/.octane-cache

conformance:
  stage: conformance
  image: ghcr.io/evcoreco/octane:latest
  cache:
    key:
      files:
        - scenarios/**/*
        - connections/**/*
        - octane.yml
      prefix: octane-${CI_RUNNER_TAGS}
    paths:
      - .octane-cache/
    policy: pull-push
  script:
    - octane run scenarios/v16/ --fail-on major
  artifacts:
    when: always
    paths:
      - reports/
```

Notes:

- GitLab's `cache.key.files` automatically hashes the listed files.
  The `prefix` namespaces by runner tags so jobs on different
  runners do not contend.
- `policy: pull-push` is the default — restore at job start, save
  at job end. A read-only consumer would use `policy: pull`.
- For parallel jobs (`parallel: 4`), each job has its own cache
  namespace by default; combine with the `prefix` to control
  sharing explicitly.

### `.gitignore` recommendation

```
.octane-cache/
reports/
```

The cache is local-only. It is never committed.

## Pruning and size management

The cache grows monotonically until pruned. Two pruning paths:

| Trigger | Mechanism |
|---------|-----------|
| Manual | `octane cache prune --max-age 7d` |
| Automatic (per-run, opt-in) | `octane.yml` carries `cache.auto_prune_max_age: 30d`; the runner prunes at startup |
| CI | The CI cache layer evicts old keys per its own LRU policy (GitHub: 10GB per repo; GitLab: configurable per project) |

A cache directory with several thousand entries is fine; SQLite
would have been overkill for that scale anyway.

## Consequences

### Positive

- **CI-native.** Two parallel CI jobs write to disjoint paths
  (different `key_hash` prefixes); their cache directories merge
  cleanly when the next job restores both.
- **Inspection is trivial.** `cat result.json | jq` and
  `cat trace.json | jq` cover the operator's debugging needs. No
  SQL, no third-party tool to install.
- **Compression-friendly.** JSON files compress 10–20× under gzip.
  CI cache uploads are dramatically smaller than the SQLite
  equivalent.
- **Partial cache restoration works.** A cache that was 80%
  restored gives 80% hits. With SQLite, a half-restored database
  is corrupt.
- **No third-party Go dependency.** OCTANE reverts to a single
  non-stdlib runtime dep (`nhooyr.io/websocket` per ADR 0003),
  honoring constitution principle V.
- **Atomic-rename writes are simpler than SQLite transactions.**
  POSIX atomic rename is well-understood and works identically
  across Linux, macOS, and Windows.

### Negative

- **No multi-row atomic transactions.** OCTANE doesn't actually
  need any.
- **No SQL-based ad-hoc analytics.** A power user wanting to query
  "which tests failed twice in the last week" needs to write a
  shell pipeline. Acceptable for the workload.
- **Filesystem inode pressure** at very large suite sizes (hundreds
  of thousands of entries). Two-character fanout mitigates this;
  beyond that scale, increase fanout depth (a future schema bump).
- **Cache pruning is filesystem walk-based**, slower than a SQL
  `DELETE` for very large caches. Mitigated by running prune as a
  background task or in CI's idle time.

### Neutral

- The schema is per-file (`schema_version` field). Future evolution
  is mixed-version: old entries with `schema_version=1` remain
  readable while new writes use `schema_version=2`. A migration
  pass can rewrite-in-place if needed.

## Alternatives considered

- **SQLite.** Considered and adopted in a previous draft of this
  ADR; reversed once the CI cache scenario was re-examined.
  Justification recorded in the Context section above.
- **BoltDB or other embedded key-value stores.** Same single-file
  drawback as SQLite for the CI scenario.
- **JSON Lines append-only log.** Considered. Rejected because
  cache entries can have very different sizes (small `result.json`
  vs large `trace.json`), and a log makes per-entry deletion
  expensive.
- **Per-entry directory with one file per field.** Considered.
  Rejected as overkill: `result.json` is small and well-structured.
- **Ignore CI caching entirely; do everything in-memory.**
  Rejected: cross-run caching is the whole point of the cache.

## References

- Constitution principles II, IV, V, X
- ADR 0015 (test dependency graph) — caching is what makes the
  graph efficient
- Bazel disk cache layout: <https://bazel.build/remote/caching>
- ccache layout: <https://ccache.dev/manual/latest.html>
- Go build cache layout: `$GOCACHE`, populated by
  `cmd/go/internal/cache`
- GitHub Actions cache: <https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows>
- GitLab CI cache: <https://docs.gitlab.com/ee/ci/caching/>
