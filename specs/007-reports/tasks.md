# Tasks 007: Reports

> **Spec ID:** `007-reports`
> **Plan reference:** `./plan.md`
> **Status:** Ready (provisional — depends on spec 005)

## Conventions

Same as previous specs.

---

## Phase 1 — Data model

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-01 | Define `report/model.Report`, `StoryReport`, `Trace`, `Frame`, `Finding` | architect | S | AC1, AC5 | `pkg/report/model/model.go` |
| T-007-02 | Define `JSONOptions`, `RobotXMLOptions` | architect | P | AC6 | `pkg/report/options.go` |
| T-007-03 | Implement `runner.RunResult → report.model.Report` projection | backend | S | AC1 | `pkg/report/model/from_runner.go` |

## Phase 2 — Redaction

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-10 | Implement deny-by-default redactor for auth fields | backend | S | AC7 | `pkg/report/internal/redact/redact.go` |
| T-007-11 | Implement HTTP-header redactor (regex against known header names) | backend | P | AC7 | `pkg/report/internal/redact/headers.go` |
| T-007-12 | Redaction unit tests — every documented redacted field | qa | S | AC7 | `pkg/report/internal/redact/redact_test.go` |
| T-007-13 | Redaction property test: random data → no `auth.token` ever appears in output | qa | P | AC7 | `pkg/report/internal/redact/property_test.go` |

## Phase 3 — JSON emitter

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-20 | Implement `WriteJSON` with sorted-keys encoding | backend | S | AC1, AC2 | `pkg/report/json/json.go` |
| T-007-21 | Implement trace inclusion / `--no-trace-on-pass` toggle | backend | P | AC5, AC6 | `pkg/report/json/trace.go` |
| T-007-22 | Implement cause-chain walking | backend | P | AC8 | `pkg/report/json/cause.go` |
| T-007-23 | JSON byte-determinism golden tests across Linux/macOS/Windows | qa | S | AC2 | `pkg/report/json/golden_test.go` |
| T-007-24 | JSON schema validation test (against a pinned schema) | qa | P | AC1 | `pkg/report/json/schema_test.go` |

## Phase 4 — Robot XML emitter

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-30 | Implement `WriteRobotXML` mapping per spec 007 §10 | backend | S | AC3, AC9 | `pkg/report/robotxml/robotxml.go` |
| T-007-31 | Implement status mapping (passed/failed/skipped/bypassed → PASS/FAIL/SKIP/NOT RUN) | backend | P | AC9 | `pkg/report/robotxml/status.go` |
| T-007-32 | Robot XML golden tests | qa | S | AC3 | `pkg/report/robotxml/golden_test.go` |
| T-007-33 | rebot-consumability integration test (via Docker) | qa | S | AC4 | `test/integration/rebot_test.go` |

## Phase 5 — Concurrency

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-40 | Concurrent-write safety test (parallel runs to shared parent dir) | qa | S | AC10 | `pkg/report/concurrent_test.go` |

## Phase 6 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-50 | Godoc on every exported symbol | docs | P | — | `pkg/report/*.go` |
| T-007-51 | `docs/concepts/reports.md` with format examples | docs | P | — | `docs/concepts/reports.md` |
| T-007-52 | `docs/integrations/robot-framework.md` (rebot quickstart) | docs | P | AC4 | `docs/integrations/robot-framework.md` |
| T-007-53 | CHANGELOG entry | docs | S | — | `CHANGELOG.md` |

## Phase 7 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-007-60 | Security review: redaction completeness | security | S | AC7 | — |
| T-007-61 | Reviewer pass: schema stability, byte determinism | reviewer | S | AC2 | — |

---

## Definition of Done

- [ ] All 10 acceptance criteria covered by at least one task
- [ ] JSON byte-determinism test green across Linux/macOS/Windows
- [ ] rebot consumes Robot XML output without warnings
- [ ] Redaction property test passes with random adversarial data
- [ ] CHANGELOG updated under `## [Unreleased]`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/007-reports` passes
