---
sidebar_position: 3
---

# Troubleshooting

Start with two things: the **exit code** of the process and the
**`report.json`** for the run. The exit code tells you the *category* of
problem; the report tells you the failing step and the wire frames behind
it.

## Triage by exit code

| Exit code | Category | Where to look |
|---|---|---|
| `0` | All stories passed | — |
| `1` | A story failed | `report.json` → failing story → step + wire trace |
| `64` | Config / flag error | the message on stderr; your `octane.yml` and flags |
| `74` | I/O error | cache dir, story file, and report dir permissions |
| `125` | Internal error (bug) | open an issue with the command and output |

See the [exit-codes reference](../reference/exit-codes.md) for the full
table.

## Symptoms and fixes

### `exit 64` — configuration or flag error

Common causes:

- **Wrong YAML key case.** `octane.yml` keys are **camelCase**
  (`storiesDir`, `maxParallel`, `ocppVersion`, `lockTimeout`, `failOn`).
  `snake_case` keys are silently ignored and defaults apply.
- **Unparseable flag value**, e.g. a bad duration.
- **Invalid `--shard`.** It must be `N/M` with `1 ≤ N ≤ M`.

Fix the config or flag and re-run. Validate config and stories up front:

```bash
octane validate stories scenarios/v16
```

### `exit 74` — I/O error

The cache directory is inaccessible, a story file is unreadable, or the
report directory is unwritable. Check permissions and that
`--cache-dir` / `OCTANE_CACHE_DIR` and `--report-dir` point somewhere
writable.

### Connection refused / handshake never completes

- Confirm the CSMS is up and listening.
- Confirm the **endpoint and port**. For CitrineOS the OCPP-J WebSocket is
  `ws://<host>:9210` — **not** the REST/admin port `8080`.
- Remember OCTANE appends `/<stationHandle>` per station, so the
  `--csms-endpoint` value is the **base** URL with no station path.

```bash
octane run scenarios/v16/station_connection_established.story \
    --csms-endpoint ws://localhost:9210
```

### TLS certificate errors

By default OCTANE verifies TLS certificates. For a dev CSMS with a
self-signed certificate you can disable verification — **never in
production**:

```bash
octane run scenarios/v16 --csms-endpoint wss://localhost:9210 --insecure-skip-verify
# or, for headless CI that cannot pass flags:
OCTANE_INSECURE_SKIP_VERIFY=true octane run scenarios/v16 ...
```

OCTANE prints a warning banner to stderr whenever verification is
disabled.

### Everything is `skipped`

A foundational prerequisite failed, and its dependents were held back
rather than failed. Find the **root** failing story — its report entry is
the one with `status: failed`; the skipped entries point back to it. Fix
that prerequisite first. See
[dependency graph & caching](../concepts/dependency-graph.md).

### Lock contention / a run hangs at start

Two runs are competing for the same cache key (common with parallel CI
jobs). The second waits up to `--lock-timeout` (default `60s`). To fail
fast instead of waiting:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210 --no-wait
```

### Stale or unexpected `cache-hits`

The cache key already includes the story content and all of its
prerequisites, so editing a story invalidates it automatically. If you
still want a guaranteed-fresh run:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210 --no-cache
# or wipe results entirely:
octane cache clear
```

Inspect the cache directly:

```bash
octane cache info
octane cache key boot_sequence_accepted
```

### "no match" for a step at preflight

A step did not resolve to any keyword. Check the exact pattern:

```bash
octane keywords list
octane keywords resolve "station \"CP01\" sends Heartbeat"
```

`resolve` prints the matched pattern, or the nearest suggestion when there
is no match. Watch for typos and remember that `.story` files must use
**spaces, not tabs**.

## Quick debugging checklist

```bash
octane validate stories scenarios/v16      # syntax & structure
octane keywords resolve "<your step>"      # does the step resolve?
octane run <one-story> --csms-endpoint <url> --no-cache --verbose
```

## Next

- **[Reports](./reports.md)** — reading a failing report and its wire
  trace.
- **[Exit codes](../reference/exit-codes.md)** — the full code table.
- **[CLI reference](../reference/cli.md)** — flags referenced above.
