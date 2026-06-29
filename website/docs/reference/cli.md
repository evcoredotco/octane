---
sidebar_position: 1
---

# CLI Reference

The `octane` binary is the primary surface. This page documents the
command tree, global flags, and per-subcommand flags. For the
authoritative version of any command in your build, run
`octane help <command>`.

```text
octane [global-flags] <command> [command-flags] [args]
```

| Command | Purpose |
|---|---|
| `run` | Run `.story` conformance suites against a CSMS. |
| `validate` | Parse and validate story files without executing them. |
| `keywords` | Inspect the registered keyword vocabulary. |
| `cache` | Manage the content-addressed result cache. |
| `completion` | Generate shell completion scripts. |

## Global flags

Persistent flags apply to every subcommand.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--config` | string | `octane.yml` | Path to the configuration file. |
| `--verbose`, `-v` | bool | `false` | Enable verbose output. |
| `--no-cache` | bool | `false` | Bypass the result cache entirely. |
| `--cache-dir` | string | `$XDG_CACHE_HOME/octane/cache` | Override the cache directory. |

## `octane run`

Run discovers and executes `.story` files against a CSMS endpoint.

```text
octane run [story-paths...] [flags]
```

Story paths may be files or directories (searched recursively). When no
paths are given, OCTANE uses `storiesDir` from `octane.yml` (default
`scenarios`).

| Flag | Type | Default | Description |
|---|---|---|---|
| `--csms-endpoint` | string | `` | Base WebSocket URL of the CSMS under test, e.g. `ws://localhost:9210`. |
| `--max-parallel` | int | `1` | Maximum stories to run concurrently. |
| `--shard` | string | `` | Shard index in `N/M` format (e.g. `1/4`); selects a subset by `sha256(test_id) % M`. |
| `--ocpp-version` | string | `` | Restrict the run to stories declaring this version (e.g. `1.6`). |
| `--lock-timeout` | duration | `60s` | Max wait to acquire a per-cache-key lock. |
| `--no-wait` | bool | `false` | Fail immediately when a cache lock is busy. |
| `--insecure-skip-verify` | bool | `false` | Disable TLS verification. Emits a warning; never use in production. |
| `--param` | stringArray | `` | Story parameter override in `name=value` form. May be repeated. |
| `--fail-on` | string | `any` | Exit non-zero when reached: `any` (default) or `major`. |
| `--report-dir` | string | `reports/` | Directory for per-run report subdirectories. |
| `--no-trace-on-pass` | bool | `false` | Omit wire traces from reports for stories that passed. |

**Output.** A one-line summary on stdout, followed by the report location:

```text
passed=N failed=M skipped=K cache-hits=J
report-dir=reports/<run-id>
```

**Exit codes.** `0` all passed · `1` a story failed · `64` config/flag
error · `74` I/O error · `125` internal error. See
[exit codes](./exit-codes.md).

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210
octane run scenarios/v16/station_boot_accepted.story --csms-endpoint ws://localhost:9210
octane run scenarios/v16/transaction_pluginfirst_accepted.story --param valid_idTag=AABBCC
octane run scenarios/v16 --csms-endpoint ws://localhost:9210 --shard 1/4 --max-parallel 4
```

## `octane validate stories`

Parse and structurally validate `.story` files without running them.

```text
octane validate stories [paths...] [flags]
```

Paths may be files or directories (searched recursively); with none given,
the current directory is searched. One line is printed per file:

```text
OK: scenarios/v16/station_boot_accepted.story
ERROR: scenarios/v16/broken.story: <message>
```

**Exit codes.** `0` when every file is valid; `64` when any file fails to
parse.

## `octane keywords`

Inspect the keyword registry.

```text
octane keywords list
octane keywords resolve <step-text>
```

- **`list`** prints every registered keyword, sorted by layer then OCPP
  version then pattern, one per line:

  ```text
  [primitive] [unknown] wait {duration:duration}
  [domain]    [1.6]     station {station:string} sends Heartbeat
  ```

- **`resolve`** matches a step against the registry and prints the matched
  pattern, layer, and version — or a `no match` message with the closest
  suggestion (by edit distance).

Both commands exit `0`.

:::note Pre-alpha
In the current build, `octane keywords list` wires in only the primitive
layer. The full domain catalog is documented in the
[keyword catalog](./keyword-catalog.md).
:::

## `octane cache`

Manage the [content-addressed cache](../concepts/dependency-graph.md).

| Command | Description |
|---|---|
| `octane cache info` | Print the effective cache directory and statistics. |
| `octane cache prune [--max-age <dur>]` | Remove entries older than `--max-age` (default `24h`) or past TTL. |
| `octane cache clear` | Remove all result entries (preserves the directory structure). |
| `octane cache key <story-id>` | Print the SHA-256 cache key for a story ID. |

```bash
octane cache info
octane cache prune --max-age 168h
octane cache key boot_sequence_accepted
```

Cache commands exit `0` on success and `74` on I/O error.

## `octane completion`

Generate a shell completion script.

```text
octane completion [bash|zsh|fish|powershell]
```

| Shell | Load in the current session |
|---|---|
| bash | `source <(octane completion bash)` |
| zsh | `source <(octane completion zsh)` |
| fish | `octane completion fish \| source` |
| powershell | `octane completion powershell \| Out-String \| Invoke-Expression` |

**Exit codes.** `0` on success; `64` for an unsupported shell; `74` on a
write error.

## Next

- **[Configuration schema](./config-schema.md)** — `octane.yml` and
  environment variables.
- **[Exit codes](./exit-codes.md)** — the full code table.
- **[CI integration](../operations/ci-integration.md)** — these commands in
  a pipeline.
