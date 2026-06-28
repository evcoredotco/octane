---
sidebar_position: 6
---

# Exit Codes

OCTANE uses stable process exit codes so CI jobs can distinguish a failed
conformance assertion from a configuration mistake, an I/O problem, or an
internal bug. The canonical values are defined in
`cmd/octane/internal/exitcode`.

| Code | Constant | Meaning |
|---|---|---|
| `0` | `OK` | All stories passed, or a read-only command completed without error. |
| `1` | `TestFailed` | One or more stories failed execution. |
| `64` | `ConfigError` | Configuration or flag error — malformed YAML, an unparseable value, or a missing required input. |
| `74` | `IOError` | I/O failure — cache directory inaccessible, story file unreadable, or report unwritable. Follows BSD `EX_IOERR`. |
| `125` | `InternalError` | Unexpected internal failure; indicates a bug in OCTANE. |

Exit codes `2`–`63`, `66`–`73`, and `75`–`124` are reserved for future
use.

## Notes by command

- **`octane run`** exits `0` when nothing failed and `1` when any story
  failed. The `--fail-on` flag selects the threshold (`any`, the default,
  fails on the first failed story; `major` is reserved). A bad config or
  flag yields `64`; a cache/file/report I/O problem yields `74`.
- **`octane validate stories`** exits `0` when every file is valid and
  `64` when any file fails to parse.
- **`octane keywords list` / `resolve`** exit `0`. A "no match" from
  `resolve` is reported in the output, not via a non-zero code.
- **`octane cache …`** exits `0` on success and `74` on I/O error.
- **`octane completion`** exits `0` on success, `64` for an unsupported
  shell name, and `74` on a write error.

## Using exit codes in CI

A required conformance check needs no special handling — a non-zero exit
fails the step automatically:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210
# step fails the job if any story failed (exit 1) or the config is bad (64)
```

To distinguish *categories* of failure in a script:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210
case $? in
  0)   echo "conformant" ;;
  1)   echo "conformance failure — see reports/" ;;
  64)  echo "configuration error" ;;
  74)  echo "I/O error" ;;
  125) echo "internal error — please file a bug" ;;
esac
```

## Next

- **[CLI reference](./cli.md)** — per-command behavior.
- **[Troubleshooting](../operations/troubleshooting.md)** — diagnosing each
  code.
- **[Reports](../operations/reports.md)** — what a `1` looks like in detail.
