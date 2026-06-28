# Spec 007: Reports — JSON and Robot XML

> **Spec ID:** `007-reports`
> **Status:** Approved (provisional — depends on spec 005 final shape)
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCTANE produces two report formats per ADR 0009:

1. **OCTANE-native JSON** — the canonical, machine-readable
   output. Future tooling (HTML viewers, dashboards, scoring
   pipelines) consumes this.
2. **Robot Framework `output.xml`** — a compatibility format that
   makes OCTANE results consumable by the Robot Framework
   ecosystem's tooling: `rebot` for combining and re-rendering
   reports, `robotframework-lsp` for editor integration, and
   so on.

The two formats are emitted from the same in-memory `RunResult`
data model. This spec defines that model, its byte-deterministic
JSON serialization, and the projection to Robot's `output.xml`
schema.

## 2. Goals

- G1. Implement `pkg/report/` with three subpackages: `model/`
      (the in-memory `RunResult`), `json/` (the OCTANE-native
      format), `robotxml/` (the Robot compatibility projection).
- G2. JSON output is **byte-deterministic**: identical inputs
      across runs and platforms produce byte-identical bytes
      (per constitution principle IV).
- G3. Robot XML output validates against the Robot Framework 7.x
      `output.xml` schema and is consumable by `rebot --report
      report.html output.xml` without errors.
- G4. Both emitters operate on the same `RunResult`; they share
      no state and can be invoked in any order or in parallel.
- G5. Wire trace embedding: failed-test traces are always
      embedded; passing-test traces are embedded by default but
      suppressible via `--no-trace-on-pass` (consistent with
      ADR 0016).
- G6. Report file paths follow a predictable schema:
      `reports/<run-id>/octane.json`,
      `reports/<run-id>/output.xml`. The CLI flag
      `--report-dir` (default `reports/`) controls the parent.

## 3. Non-Goals

- N1. HTML report rendering. `rebot` produces this from Robot XML;
      OCTANE does not maintain its own HTML emitter.
- N2. Cache schema (defined in ADR 0016 and implemented in spec
      005). The cache and the report are distinct artifacts;
      they share input data but are not interchangeable.
- N3. Live progress streaming (a follow-up; the runner emits
      structured events, but report files are written at the
      end of a run).
- N4. Report aggregation across multiple `octane run`
      invocations. `rebot` handles this for Robot XML; for
      OCTANE-native JSON, callers concatenate the JSON arrays
      themselves.

## 4. User Stories

- **As an operator**, I want a JSON report I can grep, jq, and
  feed to dashboards.
- **As a Robot Framework user**, I want OCTANE's results to
  appear in `rebot`-rendered HTML alongside my other Robot test
  results, with consistent severity rendering.
- **As a CI maintainer**, I want determinism: the same inputs
  must produce the same report bytes, so I can hash the report
  for change detection.
- **As a debug user**, I want a failed test's wire trace embedded
  directly in the report, not stashed in a separate location I
  have to discover.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| IV. Determinism | JSON output is byte-identical for identical inputs. Map iteration order is forbidden in the serializer; use sorted slices throughout. |
| V. Stdlib-Heavy | JSON serialization uses `encoding/json` only. Robot XML uses `encoding/xml`. No third-party serialization. |
| VI. Test Cases as Code | The Go data model is the source of truth; both formats project from it. |
| X. Security | Reports MUST NOT contain credentials. Connection profile auth blocks are redacted before serialization. |

## 6. Acceptance Criteria

- AC1. **Given** a `RunResult` with N stories executed,
       **when** `report.WriteJSON(result, dir)` is called,
       **then** a file `dir/octane.json` is written containing
       a JSON object with one entry per story, sorted by story
       ID.
- AC2. **Given** the same `RunResult`, **when**
       `report.WriteJSON` is invoked twice into temp directories,
       **then** the two files are byte-identical.
- AC3. **Given** the same `RunResult`, **when**
       `report.WriteRobotXML(result, dir)` is called, **then** a
       file `dir/output.xml` is written that validates against
       the Robot Framework 7.x XSD.
- AC4. **Given** the file from AC3, **when**
       `rebot --report report.html dir/output.xml` runs, **then**
       it produces an HTML report with no warnings or errors.
- AC5. **Given** a failed story, **when** the JSON report is
       written, **then** the story's entry contains a complete
       wire trace (every CALL/CALLRESULT/CALLERROR frame in
       order, with timestamps).
- AC6. **Given** a passing story and `--no-trace-on-pass` set,
       **when** the report is written, **then** the story's
       entry contains a `trace_present: false` field and no
       trace data.
- AC7. **Given** a connection profile with a populated `auth`
       block, **when** the report is written, **then** the
       `auth` section in the report is replaced with `"<redacted>"`.
- AC8. **Given** a `RunResult` containing a skipped story with
       `Cause: "station_boot_accepted/CP01"`, **when** the
       report is written, **then** the skipped entry's
       `cause_chain` field traces back to the original failure.
- AC9. **Given** the Robot XML output, **when** opened in
       `robotframework-lsp` or any Robot 7.x consumer, **then**
       OCTANE story names render correctly, severity labels map
       to Robot's pass/fail/skip, and durations display in
       seconds.
- AC10. **Given** N parallel `octane run` invocations with
        distinct run-ids, **when** each writes to a shared parent
        report directory, **then** no file conflicts occur and
        each invocation's reports are in its own subdirectory.

## 7. OCPP Scope

The report data model is OCPP-version-agnostic. Wire traces
preserve the OCPP version per-frame as observed; the report
itself does not interpret message semantics.

## 8. Open Questions

- OQ1. **(speculative — depends on spec 005)** Whether
       `cache_status` (hit-pass / hit-skip / miss / bypassed)
       per spec 005 OQ2 should be visible in the JSON report.
       Recommendation: yes, as a top-level field per story
       entry. Critical for CI debugging ("why did this run take
       so long").
       *(owner: Architect, due: spec 005 implementation.)*
- OQ2. Whether to support a custom JSON schema version field for
       forward compatibility. Recommendation: yes, top-level
       `schema_version: 1`. Future schema changes increment.
       *(owner: Architect, due: with this spec — RESOLVED.)*
- OQ3. Robot XML's `<status>` element supports `PASS`, `FAIL`,
       `SKIP`, `NOT RUN`. OCTANE's status set is `passed`,
       `failed`, `skipped`. Mapping is straightforward except for
       `bypassed` (a cache-bypassed story). Recommendation: map
       `bypassed` to Robot's `NOT RUN` with a `<msg>` explaining
       the bypass.
       *(owner: Architect, due: with this spec — RESOLVED.)*

## 9. Out of Scope (parking lot)

- HTML emitter (use `rebot`).
- JUnit XML emitter (Robot XML covers this use case via
  `rebot --xunit junit.xml`).
- Streaming reports (per-test write as it completes).
- Report diffing (`octane report diff a.json b.json`) — a
  follow-up tool, not part of this spec.
- Custom report templates.

## 10. Implementation notes

### Determinism strategy

`encoding/json` does not guarantee key order across map
iterations. The emitter MUST NOT serialize maps directly; it
walks an explicit slice of `(key, value)` pairs sorted by key.
The same applies to lists: stories are sorted by ID, findings
within a story are sorted by `(severity_desc, message)`, frames
within a trace are in temporal order (which is naturally
deterministic).

### File layout

```
reports/
└── <run-id>/                 # ULID, ordered by run start time
    ├── octane.json           # OCTANE-native, primary
    ├── output.xml            # Robot Framework, compatibility
    └── traces/               # only if --trace-files-separate
        └── <test_id>.trace.json
```

By default, traces are inline in `octane.json`. The flag
`--trace-files-separate` writes them to sibling files for very
large traces.

### JSON schema (informative)

```json
{
  "schema_version": 1,
  "octane_version": "0.1.0",
  "run_id": "01HXXXXXXXXXXXXXXXXXXXXXXX",
  "started_at": "2026-04-26T08:00:00Z",
  "finished_at": "2026-04-26T08:00:42Z",
  "summary": {
    "total": 12,
    "passed": 10,
    "failed": 1,
    "skipped": 1
  },
  "stories": [
    {
      "test_id": "boot_sequence_accepted",
      "scope_key": "CP01",
      "ocpp_version": "1.6",
      "status": "passed",
      "cache_status": "miss",
      "started_at": "...",
      "finished_at": "...",
      "duration_ms": 1234,
      "findings": [],
      "trace": { ... }
    }
  ]
}
```

### Robot XML mapping

| OCTANE field | Robot XML element |
|--------------|-------------------|
| story | `<test>` inside an outer `<suite>` |
| story.findings (severity≥major) | `<status status="FAIL"><msg>` |
| story.findings (severity<major) | `<status status="PASS">` plus `<msg level="WARN">` |
| story status `passed` | `<status status="PASS">` |
| story status `failed` | `<status status="FAIL">` |
| story status `skipped` | `<status status="SKIP">` |
| story status `bypassed` | `<status status="NOT RUN">` |
| story.duration_ms | `starttime` / `endtime` attributes |
| trace frames | `<kw>`-wrapped log lines per frame |

The mapping is one-way; consuming Robot XML and rehydrating
into a `RunResult` is not in scope.

---

## Approval

- [x] Architect / Spec author
- [ ] Backend implementer (sanity check after spec 005 lands)
- [x] Maintainer review

> **Note:** This spec depends on the `RunResult` shape from
> spec 005. The detailed mapping in -10 may need revision once
> spec 005 lands. Marked Approved Provisional.
