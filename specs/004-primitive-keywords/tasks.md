# Tasks 004: Primitive Keywords

> **Spec ID:** `004-primitive-keywords`
> **Plan reference:** `./plan.md`
> **Status:** Ready

## Conventions

Same as previous specs.

---

## Phase 1 — Connection primitives

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-01 | Implement `open a WebSocket to {url} as station {station}` | backend | P | AC1 | `pkg/keywords/primitive/open.go` |
| T-004-02 | Implement variant with `with subprotocol {subprotocol}` | backend | P | AC1 | `pkg/keywords/primitive/open.go` |
| T-004-03 | Implement `close station {station}` | backend | P | — | `pkg/keywords/primitive/close.go` |
| T-004-04 | Implement `the connection on station {station} is open/closed` | backend | P | — | `pkg/keywords/primitive/status.go` |
| T-004-05 | Unit tests for connection primitives against `mock.NewMockStation` | qa | S | AC1 | `pkg/keywords/primitive/open_test.go`, `close_test.go`, `status_test.go` |

## Phase 2 — Frame primitives

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-10 | Implement `send raw frame {frame:any} on station {station}` | backend | P | AC2 | `pkg/keywords/primitive/send.go` |
| T-004-11 | Implement `send raw bytes {bytes:string} on station {station}` (hex-decoded) | backend | P | AC2 | `pkg/keywords/primitive/send.go` |
| T-004-12 | Implement `expect any frame on station {station} within {timeout}` | backend | P | AC3, AC4 | `pkg/keywords/primitive/expect.go` |
| T-004-13 | Implement `expect a frame of type {messageType:int} on station {station} within {timeout}` | backend | P | AC3 | `pkg/keywords/primitive/expect.go` |
| T-004-14 | Unit tests for frame primitives | qa | S | AC2, AC3, AC4 | `pkg/keywords/primitive/send_test.go`, `expect_test.go` |

## Phase 3 — Timing primitives

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-20 | Implement `wait {duration}` against deterministic clock | backend | P | AC5 | `pkg/keywords/primitive/wait.go` |
| T-004-21 | Determinism test: `wait` advances no real time | qa | S | AC5 | `pkg/keywords/primitive/wait_test.go` |

## Phase 4 — Smoke test

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-30 | Author `examples/stories/primitives_only.story` (smoke test using only primitives) | qa | S | AC6 | `examples/stories/primitives_only.story` |
| T-004-31 | Integration test: smoke story executes end-to-end against CitrineOS | qa | S | AC6 | `test/integration/primitives_smoke_test.go` |
| T-004-32 | Domain-vs-primitive precedence test (requires a domain keyword fixture) | qa | P | AC7 | `pkg/keywords/primitive/precedence_test.go` |

## Phase 5 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-40 | Godoc on every primitive | docs | P | — | `pkg/keywords/primitive/*.go` |
| T-004-41 | `docs/keywords/primitives.md` — full primitive catalog with examples | docs | P | — | `docs/keywords/primitives.md` |
| T-004-42 | CHANGELOG entry | docs | S | — | `CHANGELOG.md` |

## Phase 6 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-004-50 | Reviewer pass: ensure no OCPP semantics leak into primitives | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 7 acceptance criteria covered by at least one task
- [ ] `examples/stories/primitives_only.story` runs end-to-end
- [ ] No OCPP message names appear in `pkg/keywords/primitive/*.go`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/004-primitive-keywords` passes
- [ ] CHANGELOG updated under `## [Unreleased]`
