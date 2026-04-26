# Plan 002: Wire Engine

> **Spec ID:** `002-wire-engine`
> **Status:** Approved
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement the OCPP-J transport, frame parser, and determinism
primitives (clock, RNG) that the rest of OCTANE consumes. The
transport wraps `nhooyr.io/websocket` (per ADR 0003); the frame
parser handles the JSON-array shape per OCPP-J §3.4; the clock
and RNG are interfaces with real and deterministic implementations
behind them.

## 2. Architecture Touchpoints

- `pkg/transport/` — new; WebSocket client, dial logic, station handle
- `pkg/wire/` — new; OCPP-J frame parser/encoder (CALL, CALLRESULT, CALLERROR)
- `pkg/engine/clock/` — new; Clock interface + real/deterministic implementations
- `pkg/engine/rand/` — new; Rand interface + real/deterministic implementations
- `test/reference/` — new; pinned CitrineOS docker-compose for integration tests
- `test/integration/` — new; first integration test against CitrineOS

No other packages are touched. Spec 003 will consume the
`Station`, `Clock`, `Rand` interfaces.

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `pkg/transport.Dial(ctx, url, opts) (Station, error)` | new | initial |
| `pkg/transport.Station` interface | new | initial |
| `pkg/transport.DialOptions` | new struct | initial |
| `pkg/wire.MessageTypeCall/Result/Error` | new const | initial |
| `pkg/wire.Call`, `Result`, `Error` | new structs | initial |
| `pkg/wire.ParseCall`, `ParseResult`, `ParseError` | new functions | initial |
| `pkg/wire.Encode(any) ([]byte, error)` | new function | initial |
| `pkg/engine/clock.Clock` interface | new | initial |
| `pkg/engine/clock.Real()`, `Deterministic(seed time.Time)` | new constructors | initial |
| `pkg/engine/rand.Rand` interface | new | initial |
| `pkg/engine/rand.Real()`, `Deterministic(seed uint64)` | new constructors | initial |

## 4. Data Contracts

### Frame shape

OCPP-J §3.4 frames are JSON arrays. Encoding is canonical JSON
(no trailing whitespace, sorted object keys for byte
determinism).

```
CALL:       [2, "<msg-id>", "<Action>", { ... payload ... }]
CALLRESULT: [3, "<msg-id>", { ... payload ... }]
CALLERROR:  [4, "<msg-id>", "<errcode>", "<descr>", { ... details ... }]
```

### Station interface

```go
type Station interface {
    Send(ctx context.Context, frame []any) error
    Expect(ctx context.Context) ([]any, error)
    Close() error
}
```

`Send` blocks until the frame is on the wire. `Expect` blocks
until a frame arrives or `ctx` is canceled.

### Clock interface

```go
type Clock interface {
    Now() time.Time
    Sleep(ctx context.Context, d time.Duration) error
    Tick(d time.Duration) <-chan time.Time
}
```

Deterministic implementation advances only on explicit
`Tick(d)` from the test harness.

## 5. Required ADRs

- [x] ADR 0003 — `nhooyr.io/websocket` library choice
- [x] ADR 0004 — CitrineOS as reference CSMS
- [ ] **ADR 0018** (new) — Determinism primitives: how injected
      Clock and Rand interact with goroutine scheduling and
      cancellation. Drafted alongside this spec.

## 6. Test Strategy

- **Unit tests**: every error path in `pkg/wire/` (malformed
  frames). `pkg/transport/` is harder to unit-test; rely on
  integration tests for the WebSocket path.
- **Integration test against pinned CitrineOS**: bring up
  CitrineOS via docker-compose; run a BootNotification round-trip
  per AC8.
- **Determinism golden tests**: `pkg/engine/rand` produces a known
  golden sequence per seed; assert byte-equality across Linux,
  macOS, and Windows in CI matrix.
- **TLS test**: integration test against a self-signed test
  certificate; verify `--insecure-skip-verify` toggles correctly
  per AC7.
- **Subprotocol negotiation test**: against a fake CSMS that
  returns wrong subprotocol; assert `ErrSubprotocolMismatch`.

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** N/A.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `nhooyr.io/websocket` API churn | Low | Medium | Pin minor version; isolation behind `pkg/transport/` interface |
| CitrineOS upstream changes break the pinned commit | Medium | High | Pin to a SHA; manual upgrade is a documented operator step |
| Determinism leaks through goroutine scheduling | Medium | High | Channel-based coordination only; no `time.After` directly; review pass for any naked `time.Now()` |
| TLS error messages from `crypto/tls` are inscrutable | Medium | Medium | Wrap in `ErrTLSValidation` with a remediation hint |
| Frame size limit cuts a legitimate large response | Low | Medium | Default 1 MiB is generous; configurable via dial opts |

## 9. Effort Estimate

- T-shirt size: **M**
- Calendar estimate: 2–3 weeks of focused work
- Parallelizable streams: `pkg/wire` + `pkg/engine/clock` +
  `pkg/engine/rand` are independent; `pkg/transport` is the
  longest path

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] DevOps / Platform (CitrineOS pinning)
- [x] Security reviewer
- [x] Maintainer review
