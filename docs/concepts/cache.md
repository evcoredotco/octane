# Cache

OCTANE uses a content-addressed file cache to avoid re-running stories whose
outcomes are already known. This matters most in CI: OCPP conformance suites
include expensive prerequisites — boot sequences, authorization flows, full
connector lifecycle checks — that can take minutes per station. The cache lets
subsequent pipeline runs skip those stories when nothing has changed.

## What is cached

Each story execution produces a cache entry keyed by the SHA-256 of:

| Field               | Source                                                     |
|---------------------|------------------------------------------------------------|
| `test_id`           | Story `Id` Meta key                                        |
| `scope_key`         | Station handle, run ID, or empty string (per scope type)   |
| `csms_endpoint_sha` | SHA-256 of the CSMS URL, subprotocol, and auth mode        |
| `octane_version`    | Build version of the `octane` binary                       |
| `ocpp_version`      | OCPP version from story Meta or run config                 |
| `story_content_sha` | SHA-256 of the story file and all transitive prerequisites |
| `parameter_sha`     | SHA-256 of bound parameter values                          |

Any change to any field produces a different key, so a modified story, a new
CSMS endpoint, or an OCTANE upgrade automatically invalidates relevant entries.
Old unreachable entries are removed by `Prune`.

The entry stores the pass/fail/skip status as `result.json` and, optionally,
the OCPP-J wire trace as `trace.json`.

## Cache location

The cache lives at:

```txt
$XDG_CACHE_HOME/octane/cache/
```

When `XDG_CACHE_HOME` is not set, OCTANE falls back to:

```txt
~/.cache/octane/cache/
```

Override the location with the `OCTANE_CACHE_DIR` environment variable or the
`--cache-dir` flag (spec 006).

Inside the cache root, entries are stored under a two-character fanout
directory that matches the first two hex characters of the key hash:

```txt
<cache-dir>/results/<ab>/<abcdef...>/result.json
<cache-dir>/results/<ab>/<abcdef...>/trace.json   # when trace is present
```

This layout matches the pattern used by Bazel, ccache, and Go's build cache,
and compresses well for CI artifact upload.

## Entry lifetime

Set a per-story TTL via the `Cache-TTL:` Meta key:

```yaml
Meta:
  Id: station_boot_accepted
  Cache-TTL: 24h
```

A `Cache-TTL` of `0` (or the key being absent) means the entry never expires
by TTL and is only removed by explicit `Prune` with a max-age argument.
Helper stories default to `Cache-TTL: 0`.

## Atomic write protocol

Every cache write follows this sequence to prevent readers from observing
partial files:

1. Write data to `result.json.tmp`.
2. `fsync` the temp file.
3. Rename `result.json.tmp` → `result.json` (atomic on POSIX).
4. `fsync` the entry directory.

A reader that opens `result.json` always sees either the previous complete
file or the new complete file, never a partial write.

## Flock-based in-machine locking

When two `octane run` processes target the same story on the same machine,
both will see a cache miss simultaneously. Without coordination, both would
execute the story and write the result. OCTANE prevents this with POSIX
advisory locks (`flock`) on per-key lock files under `<cache-dir>/locks/`.

The acquire sequence is:

1. Read cache (fast path, no lock).
2. Acquire exclusive flock on `<hash>.lock`.
3. Read cache again (double-check inside the lock).
4. If still a miss, execute the story.
5. Write the result.
6. Release the flock.

The `--lock-timeout` flag (default 60 s) sets how long OCTANE waits for a
lock. Pass `--no-wait` to fail immediately if the lock is held.

Cross-machine concurrency is delegated to the CI caching layer: two parallel
jobs writing the same key will each produce a valid, identical result, and
whichever artifact upload lands last wins.

## Bypassing the cache

Pass `--no-cache` to skip all cache reads and writes. Every story executes
unconditionally and the `cache_status` field in the report is `bypassed` for
all entries.

## CI sharding

`--shard N/M` distributes stories across M parallel CI jobs. Job N (zero-
based) runs only stories where:

```txt
binary.BigEndian.Uint64(sha256(test_id)[:8]) % M == N
```

Prerequisites of sharded stories are always included regardless of their own
shard assignment, so the dependency graph remains intact.

When a prior run's cache is partially restored (for example, only 80% of
entries are available from the CI artifact cache), the missing 20% re-execute
and the full 100% are available for the next run.

## `cache.Cache` interface

The `pkg/cache` package exposes a `Cache` interface with three methods:

```go
Get(ctx context.Context, key Key) (*Entry, error)
Put(ctx context.Context, key Key, entry Entry) error
Prune(ctx context.Context, maxAge time.Duration) error
```

`Get` returns `cache.ErrCacheMiss` when no valid entry exists. Tests can
supply an in-memory implementation without touching the filesystem.
