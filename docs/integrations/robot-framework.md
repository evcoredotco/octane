# Robot Framework integration

OCTANE writes a `output.xml` file that conforms to the Robot Framework 7.x
output schema at the end of every `octane run`. Any tool that understands
Robot Framework output can consume it.

## What is `rebot`?

`rebot` is the Robot Framework result post-processor. It reads one or more
`output.xml` files and produces HTML logs, HTML reports, and a merged
`output.xml`. It is distributed as part of the `robotframework` Python package
and is also available as a pre-built Docker image.

## Quickstart: generate an HTML report with Docker

The command below reads every `output.xml` produced by a set of OCTANE runs
and writes `log.html` and `report.html` into `$PWD/reports`:

```sh
docker run --rm \
  -v "$PWD/reports:/reports" \
  ghcr.io/robotframework/rfdocker:7.0 \
  rebot --outputdir /reports /reports/*/output.xml
```

- `--outputdir /reports` — write the HTML artefacts back into the same volume.
- `/reports/*/output.xml` — the glob expands to every per-run subdirectory
  written by OCTANE (`<report-dir>/<run-id>/output.xml`).

Open `reports/report.html` in a browser to review pass/fail status, timing,
and the per-story wire trace frames.

## Merging multiple shard results

When `octane run` is split across CI shards (using `--shard N/M`), each shard
writes its own `output.xml`. Use `rebot --merge` to combine them into a single
result tree:

```sh
docker run --rm \
  -v "$PWD/reports:/reports" \
  ghcr.io/robotframework/rfdocker:7.0 \
  rebot --merge \
        --output /reports/merged.xml \
        --outputdir /reports \
        /reports/*/output.xml
```

- `--merge` — combines suites with the same name rather than listing them as
  separate top-level suites. This is the correct flag when all shards ran the
  same suite name (the default `"OCTANE Conformance"` or the value you set via
  `--suite-name`).
- `--output /reports/merged.xml` — writes the merged `output.xml` for further
  downstream processing.

Pass `--merge` in addition to `--output` and `--outputdir`; omitting it will
cause `rebot` to nest the shard suites rather than flatten them.

## OCTANE status mapping

OCTANE story statuses map to Robot Framework statuses as follows:

| OCTANE status        | Robot Framework status | Meaning                                                   |
|----------------------|------------------------|-----------------------------------------------------------|
| `passed`             | `PASS`                 | All assertions succeeded.                                 |
| `failed`             | `FAIL`                 | At least one assertion failed or an error was raised.     |
| `skipped`            | `SKIP`                 | A prerequisite story failed; this story was not executed. |
| any other (e.g. cache bypass edge cases) | `NOT RUN` | Story was not attempted.                    |

The `SKIP` status propagates through `rebot`'s summary counts exactly as it
does in a native Robot Framework run, so existing dashboards and CI gates that
watch the RF skip count will reflect OCTANE dependency-chain skips correctly.

## Consumed fields in `output.xml`

The following elements are written by OCTANE and are readable by `rebot`:

- `<robot generator="octane/...">` — top-level element with suite and
  statistics placeholders.
- `<suite name="OCTANE Conformance">` (or the custom suite name) — wraps all
  test cases.
- `<test id="s1-tN" name="<test_id> (<scope_key>)">` — one element per story.
- `<kw name="trace.frame">` — one keyword per OCPP-J wire frame captured for
  the story. Each keyword carries the raw JSON bytes as its `<msg>` text and a
  `PASS` status with start/end timestamps derived from the story start time.
- `<status>` with `<msg level="ERROR|WARN">` — story findings surfaced as
  Robot Framework messages.

## Installing `rebot` without Docker

If you prefer a local Python installation:

```sh
pip install robotframework
rebot --outputdir reports reports/*/output.xml
```

Robot Framework 7.0 or later is required to parse the timestamp format OCTANE
uses (`YYYYMMDD HH:MM:SS.mmm`).
