---
sidebar_position: 4
---

# Dependency Graph & Caching

OCPP scenarios are not independent. A reservation test cannot run against
a CSMS that has not registered the station; a station cannot register
without first opening an OCPP-J WebSocket. OCTANE models the whole suite
as a **directed acyclic graph (DAG) of stories** and caches results so
large suites stay fast.

## Declaring prerequisites

Any story can require another via the `Depends` Meta key:

```text
Meta
    Id:       connector_reservation_faulted
    Spec-Ref: OCPP-J 1.6 §6.40 ReserveNow
    Stations: 1
    Depends:
      - id:    connector_status_available
        scope: per-station
```

The runner walks the dependency chain transitively, executes prerequisites
in topological order, then executes the requested story. For the example
above, the resolved chain is:

```text
connector_reservation_faulted
  └── connector_status_available
        └── station_boot_accepted
              └── station_connection_established
```

## Dependency scopes

| Scope | Runs… |
|---|---|
| `per-station` *(default)* | once per station handle — `Stations: 2` runs the prerequisite for both `CP01` and `CP02`. |
| `per-run` | once for the whole run, regardless of station count. |
| `global` | once across the cache validity window. |

## Failure propagation: skip, don't cascade-fail

When a prerequisite **fails**, the stories that depend on it are
**skipped**, not failed. Each skipped entry carries a pointer to the
failing prerequisite, so an operator sees the root cause rather than a
cascade of red.

| Status | Meaning |
|---|---|
| `passed` | The story ran and every assertion held. |
| `failed` | The story ran and at least one assertion failed. |
| `skipped` | The story did not run because a prerequisite failed. |

This is why a run summary can read `failed=1 skipped=7`: one real failure,
seven dependents quietly held back. Fix the root prerequisite first.

## The content-addressed cache

To make the DAG efficient across runs, OCTANE caches results in a
content-addressed file tree (this is what produces `cache-hits` in the run
summary).

```text
$XDG_CACHE_HOME/octane/cache/
├── results/
│   └── ab/ab12cd34…/
│       ├── result.json    # status, timing, findings (~1–2 KB)
│       └── trace.json     # OCPP-J wire frames (0–MBs, optional)
├── locks/                 # one advisory-lock target per cache key
└── meta/                  # version stamp
```

The two-character fan-out matches the convention used by Bazel, ccache,
and Go's build cache: it bounds directory size and keeps inspection
trivial (`cat | jq`). Override the location with `--cache-dir` or
`OCTANE_CACHE_DIR`.

### The cache key

Each entry is keyed by the SHA-256 of a tuple. A cached result is reused
only if **all** components match the current invocation:

| Component | Source |
|---|---|
| `test_id` | the story's `Id` |
| `scope_key` | station handle (`per-station`), run ID (`per-run`), or empty (`global`) |
| `csms_endpoint_sha` | hash of URL + subprotocol + auth mode |
| `octane_version` | the binary's build info |
| `ocpp_version` | from the story or config |
| `story_content_sha` | the story file **plus all of its prerequisites' content** |
| `parameter_sha` | the bound parameters |

Editing any file in the dependency chain, upgrading OCTANE, switching CSMS
endpoints, or changing parameters all invalidate cleanly: the new key
resolves to a new path, and the old path simply becomes unreachable and is
pruned by age. A per-story `Cache-TTL` adds time-based invalidation
(defaults: `1h` for helpers, infinite for conformance stories).

### Wire-trace splitting

`result.json` holds status, timing, and findings; the optional sibling
`trace.json` holds the OCPP-J frames. Failed stories always write traces;
passing stories write traces by default. Pass `--no-trace-on-pass` to omit
them and shrink CI cache size.

### Atomic writes and locking

Every write uses temp-file → `fsync` → `rename` → `fsync` directory, so a
reader sees either the old version or the new one — never a torn write.
Cross-process safety on one machine uses POSIX advisory locks
(`LockFileEx` on Windows) per cache key:

- `--lock-timeout` (default `60s`) bounds how long a run waits for a busy
  lock.
- `--no-wait` fails fast instead of waiting.

**Cross-machine coherence is delegated to the CI cache layer.** OCTANE is
not a distributed system; coordinating runs across machines that do not
share a filesystem is exactly what GitHub Actions and GitLab CI caches do
well. See [CI integration](../operations/ci-integration.md).

### Cache commands

```bash
octane cache info             # location, size, entry count
octane cache prune --max-age 24h   # drop entries older than the cutoff
octane cache clear            # remove all results
octane cache key <story-id>   # print the resolved cache key
```

To force a clean run, pass `--no-cache` (every story executes; `cache-hits`
is always `0`).

## Next

- **[Multi-station orchestration](./multi-station.md)** — how `per-station`
  scope interacts with multiple handles.
- **[CI integration](../operations/ci-integration.md)** — restoring the
  cache across CI jobs.
- **[CLI reference](../reference/cli.md)** — the `cache` subcommands.
