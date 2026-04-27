# Reports

OCTANE writes two report files at the end of every `octane run` invocation.
Both files are placed under a run-specific subdirectory:

```
<report-dir>/<run-id>/
  octane.json   — OCTANE-native JSON report
  output.xml    — Robot Framework 7.x output.xml
```

The `run-id` is a [ULID](https://github.com/ulid/spec) generated at the start
of each run, so successive runs never overwrite one another.

The parent directory is controlled by the `--report-dir` flag (default:
`reports/`).

## JSON format (`octane.json`)

The JSON report is the canonical machine-readable record of a run. It is
written with 2-space indentation and uses `schema_version: 1`.

### Top-level fields

| Field            | Type     | Description                                                          |
|------------------|----------|----------------------------------------------------------------------|
| `schema_version` | integer  | Always `1` for this generation of the report format.                |
| `octane_version` | string   | Version string of the `octane` binary; `"dev"` in local builds.     |
| `run_id`         | string   | ULID for this run.                                                   |
| `started_at`     | string   | RFC 3339 timestamp when the run began.                               |
| `finished_at`    | string   | RFC 3339 timestamp when the run completed.                           |
| `summary`        | object   | Aggregate counts (see below).                                        |
| `stories`        | array    | Per-story results sorted by `(test_id, scope_key)`.                  |

### `summary` object

| Field        | Type    | Description                                        |
|--------------|---------|----------------------------------------------------|
| `total`      | integer | Total stories in the resolved dependency graph.    |
| `passed`     | integer | Stories that passed.                               |
| `failed`     | integer | Stories that failed.                               |
| `skipped`    | integer | Stories skipped due to a prerequisite failure.     |
| `cache_hits` | integer | Stories served from the result cache.              |

### `stories` array entries

| Field           | Type    | Description                                                           |
|-----------------|---------|-----------------------------------------------------------------------|
| `test_id`       | string  | Stable snake_case identifier of the story.                            |
| `scope_key`     | string  | Execution scope instance (station handle or run-ID).                  |
| `ocpp_version`  | string  | OCPP version declared by the story.                                   |
| `status`        | string  | `"passed"`, `"failed"`, or `"skipped"`.                               |
| `cache_status`  | string  | `"hit-pass"`, `"hit-skip"`, `"miss"`, or `"bypassed"`.               |
| `started_at`    | string  | RFC 3339 story start time.                                            |
| `finished_at`   | string  | RFC 3339 story end time.                                              |
| `duration_ms`   | integer | Execution duration in milliseconds.                                   |
| `findings`      | array   | Diagnostic messages; sorted by `(severity desc, message asc)`.       |
| `trace_present` | boolean | `true` when wire trace data was captured for this story.              |
| `trace`         | object  | Wire-level OCPP-J frames; omitted when `trace_present` is `false`.    |
| `cause`         | string  | Prerequisite whose failure triggered a skip (empty otherwise).        |
| `cause_chain`   | array   | Transitive chain of prerequisite failures (empty otherwise).          |

### Realistic example

The following shows a run with two stories: one that passed and one that
failed with a finding.

```json
{
  "schema_version": 1,
  "octane_version": "0.3.1",
  "run_id": "01HZQVG7KXABCDEF1234567890",
  "started_at": "2026-04-27T08:00:00Z",
  "finished_at": "2026-04-27T08:00:04Z",
  "summary": {
    "total": 2,
    "passed": 1,
    "failed": 1,
    "skipped": 0,
    "cache_hits": 0
  },
  "stories": [
    {
      "test_id": "station_boot_accepted",
      "scope_key": "ws://cs001.example.com:8080/ocpp",
      "ocpp_version": "1.6",
      "status": "passed",
      "cache_status": "miss",
      "started_at": "2026-04-27T08:00:00Z",
      "finished_at": "2026-04-27T08:00:01Z",
      "duration_ms": 1024,
      "findings": [],
      "trace_present": true,
      "trace": {
        "frames": [
          [2, "abc123", "BootNotification", {"chargePointModel": "ModelX", "chargePointVendor": "VendorY"}],
          [3, "abc123", "BootNotification", {"currentTime": "2026-04-27T08:00:00Z", "interval": 300, "status": "Accepted"}]
        ]
      }
    },
    {
      "test_id": "connector_reservation_faulted",
      "scope_key": "ws://cs001.example.com:8080/ocpp",
      "ocpp_version": "1.6",
      "status": "failed",
      "cache_status": "miss",
      "started_at": "2026-04-27T08:00:01Z",
      "finished_at": "2026-04-27T08:00:04Z",
      "duration_ms": 2987,
      "findings": [
        {
          "message": "expected ReserveNow.conf status Accepted, got Faulted",
          "severity": "error"
        }
      ],
      "trace_present": true,
      "trace": {
        "frames": [
          [2, "def456", "ReserveNow", {"connectorId": 1, "expiryDate": "2026-04-27T09:00:00Z", "idTag": "TAG001", "reservationId": 1}],
          [3, "def456", "ReserveNow", {"status": "Faulted"}]
        ]
      }
    }
  ]
}
```

## Robot Framework XML (`output.xml`)

The `output.xml` file conforms to the Robot Framework 7.x output schema. It
can be consumed by any tool in the Robot Framework ecosystem, including:

- `rebot` — the Robot Framework log/report post-processor
- Robot Framework listeners and result libraries
- CI plugins for Jenkins, GitLab, and GitHub that understand RF output

Each OCTANE story becomes a `<test>` element inside a single `<suite>`. Wire
trace frames are emitted as `<kw name="trace.frame">` elements within the
test. Story findings become `<msg>` child elements on the test `<status>`
node; findings with severity `"error"` use level `ERROR`, all others use
`WARN`.

OCTANE status values map to Robot Framework statuses as follows:

| OCTANE status | Robot Framework status |
|---------------|------------------------|
| `passed`      | `PASS`                 |
| `failed`      | `FAIL`                 |
| `skipped`     | `SKIP`                 |
| any other     | `NOT RUN`              |

For end-to-end usage with `rebot`, see
[`docs/integrations/robot-framework.md`](../integrations/robot-framework.md).

## Suppressing traces on passing stories

The `--no-trace-on-pass` flag instructs OCTANE to omit wire trace data for
stories that pass. When this flag is set:

- `trace_present` is `false` for every passing story.
- The `trace` field is absent from passing story objects.
- Failing and skipped stories are unaffected; their traces are always included
  when trace capture is active.

This flag is useful in CI environments where report storage is constrained and
debugging information is only needed for failures.

```sh
octane run --stories scenarios/v16 --no-trace-on-pass
```

## Redaction

OCTANE applies deny-by-default redaction to all connection profile data that
appears in wire traces and report metadata.

### Auth block redaction

Every field in a connection profile `auth` block is treated as a credential.
All values are replaced with the literal string `<redacted>` regardless of the
key name. The deny-by-default policy means no allow-list is required; operators
do not need to enumerate credential field names.

### HTTP header redaction

The following HTTP headers are redacted wherever they appear in captured
frames:

| Header name            |
|------------------------|
| `Authorization`        |
| `Cookie`               |
| `Set-Cookie`           |
| `X-Api-Key`            |
| `Proxy-Authorization`  |

Header name matching is case-insensitive. Any header not in this list passes
through unmodified.

Redacted values are replaced with `<redacted>`.
