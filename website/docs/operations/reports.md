---
sidebar_position: 2
---

# Reports

Every `octane run` produces two report files, written under
`reports/<run-id>/`. Both are built from the **same in-memory report
tree**, so they always agree — one is optimized for OCTANE and
certification, the other for the wider test-tooling ecosystem.

| File | Format | Purpose |
|---|---|---|
| `report.json` | OCTANE-native, byte-deterministic | The source of truth for conformance. |
| `output.xml` | Robot Framework 7.x output schema | Feeds Allure, ReportPortal, Jenkins, GitLab, GitHub Actions. |

The run also prints the location to stdout:

```text
passed=21 failed=0 skipped=0 cache-hits=6
report-dir=reports/run-20260628-1/
```

Change the base directory with `--report-dir` (default `reports/`).

## `report.json`

The native report carries the metadata needed to reproduce and audit a
run, plus a per-scenario tree of steps with wire frames attached:

- the OCTANE version and OCPP version;
- the connection profile identity and version;
- the SHA-256 of the effective configuration;
- the RNG seed;
- for each story: status, timing, findings, and (optionally) the OCPP-J
  wire trace.

Because the engine is [deterministic](../concepts/architecture.md),
identical inputs produce **byte-identical** reports — modulo timestamps.
That makes diffs between runs meaningful: a change in the report is a
change in observed behavior.

```json
{
  "octaneVersion": "0.x",
  "ocppVersion": "1.6",
  "configSha256": "…",
  "seed": 42,
  "summary": { "passed": 21, "failed": 0, "skipped": 0, "cacheHits": 6 },
  "stories": [
    {
      "id": "boot_sequence_accepted",
      "specRef": "OCPP-J 1.6 §6.5 BootNotification",
      "status": "passed",
      "durationMs": 82,
      "steps": [ /* … with attached wire frames … */ ]
    }
  ]
}
```

:::note Illustrative shape
The JSON above shows the *kind* of information `report.json` carries; treat
the actual file emitted by your build as authoritative.
:::

## Story statuses

| Status | Meaning |
|---|---|
| `passed` | Ran; every assertion held. |
| `failed` | Ran; at least one assertion failed. |
| `skipped` | Did not run because a prerequisite failed. The entry points to the failing prerequisite. |

A `skipped` story is not a passing story — it means OCTANE could not reach
the point where the assertion would be exercised. See
[dependency graph & caching](../concepts/dependency-graph.md).

## Wire traces

The wire trace is the sequence of OCPP-J frames exchanged for a story —
the Call, CallResult, and CallError envelopes. It is what you read to see
*exactly* what the CSMS sent.

- **Failed** stories always include their trace.
- **Passing** stories include their trace by default.
- Pass `--no-trace-on-pass` to omit traces for passing stories and shrink
  report and cache size in CI.

## Consuming `output.xml`

`output.xml` conforms to the Robot Framework 7.x schema, so the existing
reporting ecosystem works without adapters:

```bash
# Generate a standalone HTML report and log from output.xml:
rebot reports/run-20260628-1/output.xml
```

Common integrations:

| Tool | How |
|---|---|
| **Jenkins** | Robot Framework plugin reads `output.xml`. |
| **Allure** | Allure's Robot Framework adapter. |
| **ReportPortal** | Robot Framework agent / XML import. |
| **GitLab / GitHub** | Upload `reports/` as a job artifact (see [CI integration](./ci-integration.md)). |

## Next

- **[CI integration](./ci-integration.md)** — uploading and gating on
  reports.
- **[Troubleshooting](./troubleshooting.md)** — reading a failing report.
- **[Exit codes](../reference/exit-codes.md)** — what the process code
  means.
