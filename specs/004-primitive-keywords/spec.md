# Spec 004: Primitive Keywords

> **Spec ID:** `004-primitive-keywords`
> **Status:** Approved
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

The keyword library has two layers per ADR 0007: the *domain*
layer (semantic, OCPP-version-scoped: "the CSMS sends ReserveNow
…") and the *primitive* layer (transport-level, version-agnostic:
"open a WebSocket to {url}"). Spec 003 defines the API surface
and registry; this spec ships the primitive layer.

The primitive layer is small by design. It exists for two
reasons:

1. **As a fallback** when no domain keyword matches a step. A
   story testing a CSMS extension that has no formal OCPP
   message can still drive the wire using primitives.
2. **As a building block** for domain keyword authors. A domain
   keyword may delegate to primitives to reduce duplication.

This spec ships only the primitives needed by the runner to
execute conformance stories at all. Domain keywords for OCPP 1.6,
 are deferred to spec 007 and beyond.

## 2. Goals

- G1. Implement `pkg/keywords/primitive/` with a small set of
      transport-level keywords that drive `Station.Send`,
      `Station.Expect`, and basic timing assertions.
- G2. Each primitive keyword is registered at `init()` time per
      ADR 0007 and is invisible to stories that override it with
      a domain keyword.
- G3. Each primitive keyword has unit tests that exercise it
      against a mock `Station` from `pkg/keywords/api/mock`,
      demonstrating mock-friendliness (spec 003 AC8).
- G4. The set of primitives is sufficient to write a smoke test
      that opens a WebSocket, sends a CALL, expects a CALLRESULT,
      and closes — without any domain keyword.

## 3. Non-Goals

- N1. Domain keywords (OCPP-version-specific message keywords
      like `BootNotification` or `Authorize`).
- N2. Connection establishment beyond `Open WebSocket` —
      authentication flows, OCPP token negotiation, etc., are
      domain concerns.
- N3. Concurrency primitives (parallel station I/O is the
      runner's responsibility, spec 005).
- N4. The `octane keywords list` CLI command (spec 006).

## 4. User Stories

- **As a story author** writing a smoke test against a CSMS
  whose authentication flow is non-standard, I want to use
  primitive keywords directly to send raw CALL frames without a
  full domain keyword set.
- **As a domain keyword author** writing the OCPP 1.6 keyword
  catalog, I want a small primitive layer to delegate to,
  rather than reimplementing wire-level concerns in every
  domain keyword.
- **As a reviewer**, I want the primitive layer to be small
  enough to read end-to-end in one sitting, so that adding a
  primitive is a deliberate decision.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| IV. Determinism | Primitive keywords use `state.Now()` and `state.Rand()`, not stdlib alternatives. |
| V. Stdlib-Heavy | Each primitive is implementable in <100 LOC. No external dependencies beyond what spec 002 already pulls in. |
| XI. Wire Conformance Only | Primitives never persist CSMS state, never read CSMS configuration. |
| XII. No CSMS Adaptation Surface | Primitives are pure protocol mechanics; CSMS-specific behavior is forbidden. |

## 6. Acceptance Criteria

- AC1. **Given** a primitive keyword `open a WebSocket to {url:string} as station {station:string}`,
       **when** the keyword executes against a real CSMS,
       **then** a `Station` handle is registered in the runtime
       state under the given handle name.
- AC2. **Given** a primitive keyword `send raw frame {frame:any}
       on station {station:string}`, **when** the keyword
       executes, **then** the frame is encoded by `pkg/wire/`
       and emitted on the station's wire.
- AC3. **Given** a primitive keyword `expect any frame on
       station {station:string} within {timeout:duration}`,
       **when** the keyword executes and a frame arrives within
       the timeout, **then** the keyword returns successfully
       and stashes the frame in per-station scratch space.
- AC4. **Given** the same `expect` keyword and no frame arrives
       within the timeout, **when** the timeout elapses, **then**
       the keyword returns `ErrTimeout` carrying the timeout
       value and the elapsed wall-clock time (per the injected
       deterministic clock).
- AC5. **Given** a primitive keyword `wait {duration:duration}`,
       **when** the keyword executes, **then** it sleeps the
       deterministic clock by exactly that duration; in
       deterministic mode, no real wall-clock time elapses.
- AC6. **Given** the smoke test
       `examples/stories/primitives_only.story` using only
       primitive keywords, **when** the runner executes it
       against the pinned CitrineOS, **then** the story passes
       and emits a complete trace.
- AC7. **Given** a domain keyword and a primitive keyword with
       the same pattern in a story declaring OCPP 1.6, **when**
       the resolver runs, **then** the domain keyword wins
       (spec 003 AC6).

## 7. OCPP Scope

Primitive keywords are OCPP-version-agnostic. They speak only
OCPP-J framing, never OCPP message semantics.

## 8. Open Questions

- OQ1. Whether to ship a primitive keyword for sending malformed
       frames deliberately (for negative-path conformance
       testing). Recommendation: yes, name it
       `send raw bytes {bytes:string} on station {station}`,
       parameterized by a hex string. Domain keywords for
       malformed-message tests can delegate to it.
       *(owner: Architect, due: with this spec — RESOLVED in
       favor.)*
- OQ2. Whether `wait` should accept a real-clock mode override.
       Recommendation: no. If a test needs real time to pass, it
       is non-deterministic by definition; surface this as a
       suite-level configuration, not a per-keyword flag.
       *(owner: Architect, due: with this spec — RESOLVED.)*

## 9. Out of Scope (parking lot)

- Domain keywords for any OCPP version (spec 007+).
- Multi-station orchestration primitives (spec 005).
- Cryptographic primitives (signing, JWT, OCSP) — these are
  domain-level concerns scoped to OCPP 2.x and live in their
  own spec.

## 10. Primitive keyword catalog (initial)

| # | Pattern |
|---|---------|
| 1 | `open a WebSocket to {url:string} as station {station:string}` |
| 2 | `open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}` |
| 3 | `close station {station:string}` |
| 4 | `send raw frame {frame:any} on station {station:string}` |
| 5 | `send raw bytes {bytes:string} on station {station:string}` |
| 6 | `expect any frame on station {station:string} within {timeout:duration}` |
| 7 | `expect a frame of type {messageType:int} on station {station:string} within {timeout:duration}` |
| 8 | `wait {duration:duration}` |
| 9 | `the connection on station {station:string} is open` |
| 10 | `the connection on station {station:string} is closed` |

These are the initial set. Additional primitives are added by
amending this spec; new primitives never come from a downstream
spec opportunistically.

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] Maintainer review
