# OCTANE Configuration Reference

## Resolution Chain

OCTANE assembles the effective configuration from four sources. Later
sources in the list override earlier ones:

```
┌─────────────────────────────────────────────────────────────────┐
│  Priority  │  Source                                            │
│  (lowest)  │                                                    │
│    1        │  Built-in defaults                                │
│    2        │  octane.yml (file on disk)                        │
│    3        │  OCTANE_* environment variables                   │
│    4        │  CLI flags (highest)                              │
│  (highest) │                                                    │
└─────────────────────────────────────────────────────────────────┘
```

A field absent from `octane.yml` falls back to the built-in default.
An environment variable set to a non-empty string overrides the YAML
value and the default. A CLI flag that is explicitly set wins over
everything.

## octane.yml Schema

Place `octane.yml` in the working directory from which you run OCTANE,
or pass an explicit path with `--config <path>`.

```yaml
schema_version: "1"

# Root directory searched for .story files when no paths are given
# on the command line.
stories_dir: "scenarios"

# Override the cache directory. Leave empty to use the XDG default.
cache_dir: ""

# Maximum number of stories that may execute concurrently.
max_parallel: 1

# Restrict execution to stories declaring this OCPP version.
# Valid values: "1.6", "2.0.1", "2.1". Empty means all versions.
ocpp_version: ""

# Maximum time to wait when acquiring a per-cache-key lock.
lock_timeout: 60s

# Disable TLS certificate verification. Never use in production.
insecure_skip_verify: false

# Exit threshold for the run command.
# "any" (default): exit 1 when any story fails.
# "major": reserved for future use.
fail_on: "any"
```

**Field reference**

| Field | Type | Default | env var override |
|---|---|---|---|
| `stories_dir` | string | `scenarios` | — |
| `cache_dir` | string | `` (XDG default) | `OCTANE_CACHE_DIR` |
| `max_parallel` | int | `1` | `OCTANE_MAX_PARALLEL` |
| `ocpp_version` | string | `` | `OCTANE_OCPP_VERSION` |
| `lock_timeout` | duration | `60s` | `OCTANE_LOCK_TIMEOUT` |
| `insecure_skip_verify` | bool | `false` | — |
| `fail_on` | string | `any` | `OCTANE_FAIL_ON` |

## Environment Variables

| Variable | Type | Corresponding flag | octane.yml field | Description |
|---|---|---|---|---|
| `OCTANE_CACHE_DIR` | string | `--cache-dir` | `cache_dir` | Override the cache directory. |
| `OCTANE_MAX_PARALLEL` | int | `--max-parallel` | `max_parallel` | Maximum concurrent stories. |
| `OCTANE_OCPP_VERSION` | string | `--ocpp-version` | `ocpp_version` | Restrict run to this OCPP version. |
| `OCTANE_LOCK_TIMEOUT` | duration | `--lock-timeout` | `lock_timeout` | Cache-lock acquisition timeout. |
| `OCTANE_FAIL_ON` | string | `--fail-on` | `fail_on` | Failure threshold (`any` or `major`). |

Duration variables use Go's duration syntax: `60s`, `5m`, `1h30m`.

Invalid values (non-parseable int or duration) are silently dropped and
the lower-priority source takes effect.

## Exit Codes

| Code | Constant | Meaning |
|---|---|---|
| `0` | `OK` | All stories passed, or a read-only command completed without error. |
| `1` | `TestFailed` | One or more stories failed execution. |
| `64` | `ConfigError` | Configuration file or flag error (malformed YAML, unparseable value, missing required input). |
| `74` | `IOError` | I/O failure (cache directory inaccessible, story file unreadable, report unwritable). Follows BSD `EX_IOERR`. |
| `125` | `InternalError` | Unexpected internal failure; indicates a bug in OCTANE. |

Exit codes 2–63 and 66–73 and 75–124 are reserved for future use.

## Cache Configuration

### Default location

OCTANE resolves the cache directory in this order:

1. `--cache-dir` CLI flag (highest)
2. `OCTANE_CACHE_DIR` environment variable
3. `$XDG_CACHE_HOME/octane/cache`
4. `$HOME/.cache/octane/cache` (lowest)

On Linux the XDG default resolves to `~/.cache/octane/cache/` unless
`XDG_CACHE_HOME` is set. On macOS, `$HOME/Library/Caches` is the
conventional location but OCTANE follows the XDG convention regardless
of platform, matching Go's own build cache behavior.

### Disabling the cache

Pass `--no-cache` or set it globally in CI to force every story to
execute against the live endpoint:

```bash
octane run --no-cache scenarios/v16/
```

When `--no-cache` is active, the `cache-hits` field of the summary is
always `0` and no results are written to or read from disk.

### Lock timeout

When two `octane run` processes compete for the same cache key (e.g.,
two parallel CI jobs running the same story), the second process waits
for the flock to be released. `--lock-timeout` (default `60s`) sets the
maximum wait. Pass `--no-wait` to fail immediately instead of waiting.
