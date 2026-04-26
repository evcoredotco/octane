# Spec 002: Wire Engine

> **Spec ID:** `002-wire-engine`
> **Status:** Approved
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCTANE drives a real CSMS over the wire. The wire engine is the
layer that opens an OCPP-J WebSocket connection, frames messages
per OCPP-J §3.4, and exchanges CALL / CALLRESULT / CALLERROR
frames with deterministic timing.

Two concerns are bundled in this spec because they are inseparable
in practice:

1. **Transport and framing.** A `Station` handle that sends and
   receives OCPP-J frames over a WebSocket, with TLS,
   subprotocol negotiation, and timeout handling.
2. **Determinism primitives.** A `Clock` and `Rand` interface
   that the runtime injects into keywords, replacing direct
   calls to `time.Now()` and `crypto/rand`. Without these, every
   downstream test would be non-reproducible (constitution
   principle IV).

These two together are the smallest unit that produces an
observable wire effect. Spec 003 (keyword API) consumes the
`Station`, `Clock`, and `Rand` interfaces this spec defines.

## 2. Goals

- G1. Implement `pkg/transport/` — a WebSocket client wrapping
      `nhooyr.io/websocket` (per ADR 0003) with TLS,
      subprotocol negotiation, and configurable timeouts.
- G2. Implement `pkg/wire/` — OCPP-J frame parsing and
      serialization for CALL (type 2), CALLRESULT (type 3), and
      CALLERROR (type 4) per OCPP-J §3.4.
- G3. Implement `pkg/engine/clock` — a `Clock` interface with a
      real-clock implementation and a deterministic-clock test
      double that advances on explicit ticks.
- G4. Implement `pkg/engine/rand` — a seedable RNG with a real
      implementation backed by `math/rand/v2` and a test double
      that produces a known sequence per seed.
- G5. Round-trip every frame in OCTANE's example trace fixtures
      through `Encode → bytes → Decode → struct` without loss.
- G6. Connect to the pinned CitrineOS instance at `test/reference/`
      and exchange a BootNotification handshake successfully.

## 3. Non-Goals

- N1. Story parsing (spec 001).
- N2. Step-text-to-keyword resolution (spec 003).
- N3. Domain-specific keyword bodies — Authorize, StartTransaction,
      etc. (spec 004 + later).
- N4. Dependency-graph orchestration (spec 005).
- N5. Report emission (spec 007).
- N6. CSMS-specific protocol quirks (per constitution XII, OCTANE
      tests conformance to the published OCPP specification, not
      to any vendor's interpretation).

## 4. User Stories

- **As a keyword author** (spec 003 consumer), I want a `Station`
  interface I can call `Send` and `Expect` on, without knowing
  whether the underlying transport is a real WebSocket or a test
  double.
- **As a determinism reviewer**, I want every time-dependent path
  in OCTANE to consume an injected `Clock` so that running the
  same suite twice with the same seed produces identical reports.
- **As a CI maintainer**, I want connection-level errors (DNS
  failure, TLS handshake failure, subprotocol mismatch) to surface
  as typed errors with clear remediation hints.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| II. Two Distribution Surfaces, One Engine | The wire engine is the engine; both CLI and Action invoke the same code path. |
| IV. Determinism | `Clock` and `Rand` are mandatory injection points. `time.Now()` and `crypto/rand` are forbidden in `pkg/keywords/`, `pkg/runner/`, `pkg/engine/`. The linter rejects them. |
| V. Stdlib-Heavy | Single non-stdlib runtime dep: `nhooyr.io/websocket` (ADR 0003). No HTTP framework, no codec library beyond `encoding/json`. |
| X. Security | TLS verification is on by default. `--insecure-skip-verify` exists but emits a banner-level finding in every report it appears in. |
| XI. Wire Conformance Only | The transport never persists CSMS state, never reads CSMS configuration. |

## 6. Acceptance Criteria

- AC1. **Given** a CSMS reachable at `wss://host/path`, **when**
       `transport.Dial(ctx, url, opts)` is called, **then** it
       returns a `Station` handle with a successfully completed
       WebSocket handshake using the OCPP subprotocol.
- AC2. **Given** a `Station` handle, **when** the caller invokes
       `Send(ctx, frame)` with a well-formed CALL frame, **then**
       the bytes on the wire are the canonical OCPP-J encoding
       per §3.4 and the frame round-trips through `wire.ParseCall`.
- AC3. **Given** an inbound CALLRESULT (type 3), **when** the
       station's reader goroutine processes the message, **then**
       the parsed `Result` is delivered through `Expect(ctx)` in
       FIFO order relative to other inbound frames.
- AC4. **Given** a malformed inbound frame (wrong array length,
       bad message type, non-string action), **when**
       `wire.ParseCall` is invoked, **then** it returns
       `ErrFrameShape` wrapping a precise reason.
- AC5. **Given** an injected deterministic `Clock` seeded at
       `2026-01-01T00:00:00Z`, **when** code calls `clock.Now()`
       and `clock.Sleep(d)`, **then** the returned times advance
       only when the test explicitly ticks the clock; no
       wall-clock drift is observable.
- AC6. **Given** an injected deterministic `Rand` seeded with
       `0xDEADBEEF`, **when** any number of `Int63()` /
       `Float64()` calls are made, **then** the sequence matches
       a frozen golden across runs and platforms.
- AC7. **Given** a TLS endpoint with an expired certificate,
       **when** `transport.Dial` is invoked without
       `--insecure-skip-verify`, **then** it returns a typed
       `ErrTLSValidation` containing the underlying x509 error.
- AC8. **Given** the pinned CitrineOS instance under
       `test/reference/`, **when** the integration test
       `TestBootNotificationHandshake` runs, **then** it
       completes a real OCPP-J 1.6 BootNotification round-trip
       in under 30 seconds.

## 7. OCPP Scope

This spec covers the OCPP-J framing layer per OCPP-J 1.6 §3.4
and OCPP 2.0.1 §3 (which defines an analogous JSON-array frame
format). Message-level semantics (Authorize.req payload schema,
BootNotification.req schema) are out of scope; they live in the
domain keyword specs (007+).

The wire engine is OCPP-version-agnostic. Subprotocol selection
(`ocpp1.6`, `ocpp2.0.1`) is a transport configuration, not a wire
behavior.

## 8. Open Questions

- OQ1. Whether to support OCPP-S (SOAP) in any form. Recommendation:
       no, ever. SOAP is deprecated in OCPP 2.x and absent from the
       roadmap.
       *(owner: Architect, due: with this spec)*
- OQ2. Buffer sizing for inbound frame queues. Recommendation:
       unbounded channel; a slow consumer is a programming error
       in keyword authorship, surfaced by a `--queue-watermark`
       diagnostic. Producing a backpressure mechanism would push
       complexity into every keyword author.
       *(owner: Backend, due: implementation)*
- OQ3. Whether to support automatic reconnection on transient
       network errors. Recommendation: no. A dropped WebSocket is
       a test failure, not a recoverable state. Reconnection logic
       belongs in operator scripts, not the engine.
       *(owner: Architect, due: with this spec)*

## 9. Out of Scope (parking lot)

- OCPP-S (SOAP) transport.
- HTTP/2 or HTTP/3 for OCPP-J (the spec defines OCPP over
  HTTP/1.1 WebSocket only).
- Connection pooling or reuse across `octane run` invocations.
- Custom WebSocket extensions (e.g., per-message-deflate
  configuration beyond what `nhooyr.io/websocket` exposes by
  default).

## 10. Implementation notes

### JSON decoding quirk

OCPP-J frames are JSON arrays. Go's `encoding/json` decodes
arbitrary JSON arrays to `[]any` and numbers to `float64`,
regardless of whether the source was an integer. Wire-level code
coerces numeric fields (`connectorId`, `messageType`,
`transactionId`) from `float64` back to `int` when extracting
them. The keyword library's `Args` accessor (defined in spec 003)
hides this from keyword authors.

### MessageType constants

OCPP-J §3.4 defines three numeric type codes:

| Code | Meaning |
|------|---------|
| 2 | CALL (request) |
| 3 | CALLRESULT (response) |
| 4 | CALLERROR (error response) |

`pkg/wire/` exposes these as named constants. The numeric values
are dictated by the specification and MUST NOT be renumbered.

### Subprotocol negotiation

The `Sec-WebSocket-Protocol` request header carries a list of
subprotocols the station accepts (`ocpp1.6`, `ocpp2.0.1`,
`ocpp2.1`). The CSMS's response selects one. If the response
omits the header or selects a value not in the request list,
`Dial` returns `ErrSubprotocolMismatch` with both lists in the
error message.

### Frame size limits

OCPP-J 1.6 §3.4 does not specify a frame size cap. OCTANE
applies a configurable `MaxFrameBytes` limit (default 1 MiB) to
guard against pathological CSMS responses. Frames exceeding the
limit return `ErrFrameTooLarge` and are not parsed.

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] DevOps / Platform
- [x] Security reviewer
- [x] Maintainer review
