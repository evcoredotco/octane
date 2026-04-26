# Tasks 002: Wire Engine

> **Spec ID:** `002-wire-engine`
> **Plan reference:** `./plan.md`
> **Status:** Ready

## Conventions

- ID format: `T-002-MM` (zero-padded)
- One agent per task
- `P` = parallel-eligible; `S` = strict ordering after prior

---

## Phase 1 — Contracts

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-01 | Define `transport.Station` interface, `DialOptions`, typed errors | architect | S | AC1, AC4, AC7 | `pkg/transport/transport.go`, `pkg/transport/errors.go` |
| T-002-02 | Define `wire.MessageType*` constants, `wire.Call`/`Result`/`Error` structs | architect | P | AC2, AC4 | `pkg/wire/types.go` |
| T-002-03 | Define `clock.Clock` interface | architect | P | AC5 | `pkg/engine/clock/clock.go` |
| T-002-04 | Define `rand.Rand` interface | architect | P | AC6 | `pkg/engine/rand/rand.go` |
| T-002-05 | Draft ADR 0018 (determinism primitives interaction model) | architect | S | AC5, AC6 | `docs/adr/0018-determinism-primitives.md` |

## Phase 2 — Wire

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-10 | Implement `wire.ParseCall`, `ParseResult`, `ParseError` | backend | S | AC2, AC4 | `pkg/wire/parse.go` |
| T-002-11 | Implement `wire.Encode` with sorted-keys canonical JSON | backend | P | AC2 | `pkg/wire/encode.go` |
| T-002-12 | Wire unit tests: every malformed-frame variant | qa | S | AC4 | `pkg/wire/parse_test.go` |
| T-002-13 | JSON quirk test: float64-as-int coercion | qa | P | AC2 | `pkg/wire/coerce_test.go` |

## Phase 3 — Determinism

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-20 | Implement `clock.Real()` (production wallclock) | backend | P | AC5 | `pkg/engine/clock/real.go` |
| T-002-21 | Implement `clock.Deterministic(seed time.Time)` test double | backend | P | AC5 | `pkg/engine/clock/deterministic.go` |
| T-002-22 | Implement `rand.Real()` (math/rand/v2 backed) | backend | P | AC6 | `pkg/engine/rand/real.go` |
| T-002-23 | Implement `rand.Deterministic(seed uint64)` | backend | P | AC6 | `pkg/engine/rand/deterministic.go` |
| T-002-24 | Cross-platform golden test for deterministic Rand sequence | qa | S | AC6 | `pkg/engine/rand/determinism_test.go` |

## Phase 4 — Transport

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-30 | Implement `transport.Dial` with TLS + subprotocol negotiation | backend | S | AC1, AC7 | `pkg/transport/dial.go` |
| T-002-31 | Implement Station handle with reader/writer goroutines | backend | S | AC1, AC2, AC3 | `pkg/transport/station.go` |
| T-002-32 | Implement `Send` with frame encoding | backend | S | AC2 | `pkg/transport/station.go` |
| T-002-33 | Implement `Expect` with FIFO inbound channel | backend | S | AC3 | `pkg/transport/station.go` |
| T-002-34 | Implement `MaxFrameBytes` enforcement and `ErrFrameTooLarge` | backend | P | — | `pkg/transport/station.go` |
| T-002-35 | Implement `ErrSubprotocolMismatch` and `ErrTLSValidation` typed errors | backend | P | AC7 | `pkg/transport/errors.go` |

## Phase 5 — Reference validation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-40 | Pin CitrineOS commit; docker-compose fixture | devops | S | AC8 | `test/reference/citrineos.version`, `test/reference/docker-compose.yaml` |
| T-002-41 | `make test-reference` Makefile target | devops | S | AC8 | `Makefile` |
| T-002-42 | Reference workflow `.github/workflows/reference.yml` | devops | S | AC8 | `.github/workflows/reference.yml` |
| T-002-43 | Integration test `TestBootNotificationHandshake` | qa | S | AC8 | `test/integration/bootnotification_test.go` |
| T-002-44 | Integration test for TLS error paths (self-signed cert fixture) | qa | P | AC7 | `test/integration/tls_test.go` |
| T-002-45 | Integration test for subprotocol mismatch (fake CSMS) | qa | P | AC1 | `test/integration/subprotocol_test.go` |

## Phase 6 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-50 | Godoc on every exported symbol | docs | P | — | `pkg/transport/*.go`, `pkg/wire/*.go`, `pkg/engine/clock/*.go`, `pkg/engine/rand/*.go` |
| T-002-51 | Update `docs/concepts/wire.md` with frame examples | docs | P | — | `docs/concepts/wire.md` |
| T-002-52 | CHANGELOG entry under `[Unreleased]` | docs | S | — | `CHANGELOG.md` |

## Phase 7 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-002-60 | Security review: TLS defaults, frame size limits, redaction | security | S | — | — |
| T-002-61 | Reviewer pass: determinism guarantees, error messages | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 8 acceptance criteria covered by at least one task
- [ ] `make test-reference` green against pinned CitrineOS
- [ ] Deterministic Rand test produces byte-identical sequence on Linux/macOS/Windows
- [ ] Security review signed off
- [ ] ADR 0018 merged
- [ ] CHANGELOG updated under `## [Unreleased]`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/002-wire-engine` passes
