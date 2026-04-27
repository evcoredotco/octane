# octane CLI Reference

> This document reflects the cobra command tree as of spec 006.
> Run `octane help <command>` for the authoritative in-binary version.

## Global Flags

Global flags are persistent — they apply to every subcommand.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--config` | string | `octane.yml` | Path to the `octane.yml` configuration file. |
| `--verbose`, `-v` | bool | `false` | Enable verbose output. |
| `--no-cache` | bool | `false` | Bypass the result cache entirely. All stories run regardless of cached results. |
| `--cache-dir` | string | `$XDG_CACHE_HOME/octane/cache` | Override the cache directory. |

---

## octane run

Run `.story` conformance test suites against a CSMS endpoint.

**Synopsis**

```
octane run [story-paths...] [flags]
```

Story paths may be individual `.story` files or directories. Directories
are searched recursively. When no paths are given, OCTANE searches the
`stories_dir` field from `octane.yml` (default: `scenarios`).

**Flags**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--max-parallel` | int | `1` | Maximum number of stories to run concurrently. |
| `--shard` | string | `` | Shard index in `N/M` format (e.g. `1/4`). Distributes stories across parallel CI workers by `sha256(test_id) % M`. |
| `--ocpp-version` | string | `` | Restrict the run to stories declaring this OCPP version (`1.6`, `1.6`, or `2.1`). When empty all versions are included. |
| `--lock-timeout` | duration | `60s` | Maximum time to wait when acquiring a per-cache-key lock. |
| `--no-wait` | bool | `false` | Fail immediately when a cache lock is busy instead of waiting. |
| `--insecure-skip-verify` | bool | `false` | Disable TLS certificate verification. Emits a warning banner. Do not use in production. |
| `--fail-on` | string | `any` | Exit with code 1 when this threshold is reached. `any` (default) fails on the first failed story. `major` is reserved. |

**Output**

On completion, a one-line summary is written to stdout:

```
passed=N failed=M skipped=K cache-hits=J
```

**Exit codes**

| Code | Meaning |
|---|---|
| `0` | All stories passed. |
| `1` | One or more stories failed. |
| `64` | Configuration or flag error. |
| `74` | I/O error (cache, file read, report write). |
| `125` | Internal error (bug in OCTANE). |

---

## octane validate stories

Validate `.story` files for syntax and structural correctness without
executing them against a CSMS.

**Synopsis**

```
octane validate stories [paths...] [flags]
```

Paths may be individual `.story` files or directories. Directories are
searched recursively. When no paths are given, OCTANE searches the
current directory (`.`).

**Output**

Each file produces one output line:

```
OK: <path>
ERROR: <path>: <message>
```

**Exit codes:** `0` when all files are valid; `64` when any file fails to
parse.

---

## octane keywords list

Print all keywords registered in the global keyword registry.

**Synopsis**

```
octane keywords list [flags]
```

Keywords are printed sorted by layer (primitive first, then domain),
OCPP version, and pattern. Each line has the form:

```
[<layer>] [<ocpp-version>] <pattern>
```

**Exit codes:** `0` always.

---

## octane keywords resolve

Resolve a step text to a keyword pattern.

**Synopsis**

```
octane keywords resolve <step-text> [flags]
```

Matches `<step-text>` against the registered keywords. On a successful
match, prints the matched pattern, layer, and OCPP version. On no match,
prints a `no match` message and the closest suggestion (Levenshtein
distance ≤ 5), if available.

**Exit codes:** `0` always (no match is not an error at the CLI level).

---

## octane cache info

Print the effective cache directory location.

**Synopsis**

```
octane cache info [flags]
```

The directory is resolved by the following chain: `--cache-dir` flag,
then `OCTANE_CACHE_DIR` env var, then `$XDG_CACHE_HOME/octane/cache`,
then `$HOME/.cache/octane/cache`.

**Exit codes:** `0` on success; `74` if the home directory cannot be
determined.

---

## octane cache prune

Remove cache entries that are older than `--max-age` or whose TTL has
expired.

**Synopsis**

```
octane cache prune [flags]
```

After removing entries, empty fanout directories are cleaned up.

**Flags**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--max-age` | duration | `24h` | Remove entries whose `WrittenAt` timestamp is older than this value. |

**Exit codes:** `0` on success; `74` on I/O error.

---

## octane cache clear

Remove all result entries from the cache.

**Synopsis**

```
octane cache clear [flags]
```

Removes the `results/` subdirectory and recreates it empty. The cache
directory structure (`version.json`, `locks/`) is preserved.

**Exit codes:** `0` on success; `74` on I/O error.

---

## octane cache key

Print the cache key SHA-256 hash for a story ID.

**Synopsis**

```
octane cache key <story-id> [flags]
```

Prints the hex digest used as the filesystem cache key for the given
story ID. The remaining key components (CSMS endpoint, story content,
parameters) are filled with placeholder values. Useful for locating
cached entries on disk or debugging cache invalidation.

**Exit codes:** `0` always.

---

## octane completion

Generate shell completion scripts for octane.

**Synopsis**

```
octane completion [bash|zsh|fish|powershell] [flags]
```

**Shells**

| Shell | Load in current session |
|---|---|
| `bash` | `source <(octane completion bash)` |
| `zsh` | `source <(octane completion zsh)` |
| `fish` | `octane completion fish \| source` |
| `powershell` | `octane completion powershell \| Out-String \| Invoke-Expression` |

**Exit codes:** `0` on success; `64` for an unsupported shell name; `74`
on I/O error writing the script.
