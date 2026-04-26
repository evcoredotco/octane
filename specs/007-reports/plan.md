# Plan 007: Reports — JSON and Robot XML

> **Spec ID:** `007-reports`
> **Status:** Approved (provisional — depends on spec 005)
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement two byte-deterministic report emitters: an
OCTANE-native JSON format (the canonical output) and a Robot
Framework `output.xml` projection (compatibility format). Both
operate on the same `RunResult` data model from spec 005.

> **Note:** Detailed mapping depends on spec 005's `RunResult`
> shape. Marked provisional until spec 005 is implemented.

## 2. Architecture Touchpoints

- `pkg/report/model/` — new; the in-memory shape (subset/projection of `pkg/runner.RunResult`)
- `pkg/report/json/` — new; OCTANE-native emitter
- `pkg/report/robotxml/` — new; Robot XML projection
- `pkg/report/internal/redact/` — new; credential redaction helper
- `pkg/report/testdata/` — new; golden fixtures for both formats
- Read-only consumer: `pkg/runner.RunResult`

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `pkg/report.WriteJSON(result *runner.RunResult, dir string, opts JSONOptions) error` | new | initial |
| `pkg/report.WriteRobotXML(result *runner.RunResult, dir string, opts RobotXMLOptions) error` | new | initial |
| `pkg/report.JSONOptions`, `RobotXMLOptions` | new structs | initial |
| `pkg/report/model.Report`, `StoryReport`, `Trace`, `Frame`, `Finding` | new structs | initial |

## 4. Data Contracts

### JSON schema

Defined in spec 007 §10. Top-level fields: `schema_version: 1`,
`octane_version`, `run_id`, `started_at`, `finished_at`,
`summary`, `stories`. Stories sorted by `test_id`. Findings
within a story sorted by `(severity_desc, message)`. Trace
frames in temporal order.

### Robot XML schema

Robot Framework 7.x `output.xml` format. The mapping table from
spec 007 §10 is the contract. `rebot --report report.html
output.xml` MUST produce a clean HTML report (AC4).

### Redaction contract

| Field | Redacted as |
|-------|-------------|
| Connection profile `auth.token` | `"<redacted>"` |
| Connection profile `auth.password` | `"<redacted>"` |
| Connection profile `auth.basic` | `"<redacted>"` |
| HTTP headers matching `(?i)authorization\|cookie\|x-api-key` | `"<redacted>"` |
| Frame payload fields named `idTag` | preserved (idTags are part of the test data, not credentials) |

## 5. Required ADRs

- [x] ADR 0009 — Robot Framework `output.xml` compatibility

No new ADRs needed.

## 6. Test Strategy

- **Determinism golden tests**: A canned `RunResult` is
  serialized to both formats; the bytes are diffed against
  golden files. Tests run on Linux/macOS/Windows; identical
  bytes everywhere (AC2).
- **rebot consumability test**: An integration test pipes the
  Robot XML through `docker run --rm robot:latest rebot
  --report report.html /input/output.xml`; assert no warnings
  (AC4).
- **Redaction test**: A canned `RunResult` containing a
  populated auth block; serialize; assert the field is
  `"<redacted>"` in the output (AC7).
- **Trace toggle test**: Run a passing story with and without
  `--no-trace-on-pass`; assert presence/absence of the trace
  block per AC5/AC6.
- **Cause chain test**: A canned skipped-story chain; assert
  the JSON entry's `cause_chain` walks back to the root failure
  (AC8).
- **Concurrency test**: Two report writes to a shared parent
  directory with distinct run-ids; assert no file conflicts
  (AC10).

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** N/A. `schema_version: 1` is the
  initial.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Robot XML schema drift between Robot 7.x and our generator | Medium | Medium | Pin Robot version in the integration test; bump deliberately |
| `encoding/json` map iteration leaks non-determinism | Medium | High | Forbid map serialization; use sorted slices; review pass + lint rule |
| Redaction misses a credential field added later | Medium | High | Redaction is a deny-by-default with an allowlist of preserved fields; new auth fields automatically inherit redaction |
| Robot XML is verbose for large traces | Low | Medium | Configure trace inclusion; large traces can use `--trace-files-separate` per spec 007 §10 |
| Cause chain walking has edge cases for diamond-shaped graphs | Medium | Medium | Property test against random DAGs |

## 9. Effort Estimate

- T-shirt size: **M**
- Calendar estimate: 1–1.5 weeks of focused work (after spec 005)
- Parallelizable streams: JSON + Robot XML emitters are
  independent; redaction helper feeds both

---

## Approval

- [x] Architect / Spec author
- [ ] Backend implementer (sanity check after spec 005 lands)
- [x] Maintainer review
