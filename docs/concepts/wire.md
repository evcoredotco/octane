# Wire Engine

The wire engine is the layer responsible for WebSocket connectivity and
OCPP-J frame encoding and decoding. It is not concerned with story parsing,
keyword dispatch, or conformance assertions — those live in higher layers.
Keyword authors (spec 003) call the wire engine through the `Station`
interface without coupling to any specific WebSocket implementation.

Two packages make up the wire engine:

| Package         | Responsibility                                              |
|-----------------|-------------------------------------------------------------|
| `pkg/transport` | WebSocket dialling, TLS, subprotocol negotiation, frame I/O |
| `pkg/wire`      | OCPP-J array parsing and serialization                      |

Two supporting packages provide determinism primitives:

| Package            | Responsibility                               |
|--------------------|----------------------------------------------|
| `pkg/engine/clock` | Wall-clock abstraction (`Clock` interface)   |
| `pkg/engine/rand`  | Random-number abstraction (`Rand` interface) |

---

## OCPP-J frame shapes

OCPP-J frames are JSON arrays whose first element is a numeric message type
code (OCPP-J -3.4). Three frame types are relevant for station-side testing.

### CALL (type 2) — station-to-CSMS request

```json
[2, "abc123", "BootNotification", {"chargepointModel": "Acme"}]
```

Elements:

| Index | Type   | Meaning                                                   |
|-------|--------|-----------------------------------------------------------|
| 0     | number | `2` — message type code for CALL                          |
| 1     | string | `UniqueID` — correlation identifier chosen by the station |
| 2     | string | `Action` — OCPP operation name, e.g. `"BootNotification"` |
| 3     | object | `Payload` — operation-specific request body               |

### CALLRESULT (type 3) — CSMS-to-station response

```json
[3, "abc123", {"status": "Accepted", "currentTime": "2026-01-01T00:00:00Z", "interval": 60}]
```

Elements:

| Index | Type   | Meaning                                      |
|-------|--------|----------------------------------------------|
| 0     | number | `3` — message type code for CALLRESULT       |
| 1     | string | `UniqueID` — matches the originating CALL    |
| 2     | object | `Payload` — operation-specific response body |

### CALLERROR (type 4) — CSMS-to-station error

```json
[4, "abc123", "NotImplemented", "Remote trigger not supported", {}]
```

Elements:

| Index | Type   | Meaning                                                   |
|-------|--------|-----------------------------------------------------------|
| 0     | number | `4` — message type code for CALLERROR                     |
| 1     | string | `UniqueID` — matches the originating CALL                 |
| 2     | string | `ErrorCode` — OCPP-J error code, e.g. `"NotImplemented"`  |
| 3     | string | `ErrorDescription` — human-readable description           |
| 4     | object | `Details` — optional additional context; `{}` when absent |

The `pkg/wire` constants `MessageTypeCall`, `MessageTypeResult`, and
`MessageTypeError` hold these numeric values. Parsing entry points are
`ParseCall`, `ParseResult`, and `ParseError` (in `parse.go`). Serialization
is `Encode` (in `encode.go`).

When a frame does not match the expected array shape, the parse functions
return `*wire.ErrFrameShape`. The `Reason` field pinpoints the structural
violation; the `Raw` field carries up to 256 bytes of the malformed frame
for diagnostic logging.

---

## Connecting to a CSMS

`transport.Dial` opens a WebSocket connection and returns a `Station` handle.

```go
import (
    "context"
    "log"

    "github.com/evcoredotco/octane/pkg/transport"
)

func connect(ctx context.Context) (transport.Station, error) {
    opts := transport.DialOptions{
        Subprotocols: []string{"ocpp1.6"},
        // HandshakeTimeout defaults to 30 seconds when zero.
        // MaxFrameBytes defaults to 1 MiB when zero.
    }

    station, err := transport.Dial(ctx, "wss://csms.example.com/ocpp/CP001", opts)
    if err != nil {
        return nil, err
    }

    return station, nil
}
```

The context passed to `Dial` governs only the WebSocket upgrade handshake.
After `Dial` returns, cancelling that context has no effect on the live
`Station`; the station manages its own reader goroutine internally.

`Station` provides three methods:

| Method   | Behaviour                                                                                                                                             |
|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Send`   | Encodes a `[]any` as a canonical OCPP-J JSON array and writes it to the WebSocket. Blocks until the frame is on the wire or the context is cancelled. |
| `Expect` | Blocks until an inbound frame arrives, the context is cancelled, or the connection closes. Delivers frames in FIFO order.                             |
| `Close`  | Gracefully shuts down the connection. Safe to call more than once.                                                                                    |

Both `Send` and `Expect` return `*transport.ErrStationClosed` if the station
has already been closed.

### TLS

TLS verification is on by default. Setting `DialOptions.InsecureSkipVerify`
to `true` disables certificate validation and causes every report produced
during that run to carry a banner-level finding. Do not set it without an
explicit operator opt-in.

### Subprotocol negotiation

Populate `DialOptions.Subprotocols` with the OCPP subprotocols to offer, in
preference order. If the CSMS selects a subprotocol not in that list, or
omits the `Sec-WebSocket-Protocol` response header entirely, `Dial` closes
the connection and returns `*transport.ErrSubprotocolMismatch`.

---

## Determinism primitives

Reproducible test execution requires that all time and randomness sources be
injectable. The engine enforces this via two interfaces:

### `clock.Clock`

```go
type Clock interface {
    Now() time.Time
    Sleep(ctx context.Context, d time.Duration) error
    After(d time.Duration) <-chan time.Time
}
```

Production wiring uses `clock.Real()`. Tests use `clock.Deterministic(seed)`
to advance time under program control. Direct calls to `time.Now()` are
forbidden inside `pkg/keywords/`, `pkg/runner/`, and `pkg/engine/` — the
linter enforces this via `forbidigo`.

### `rand.Rand`

```go
type Rand interface {
    Int63() int64
    Float64() float64
    Intn(n int) int
}
```

Production wiring uses `rand.Real()` (crypto-seeded). Tests use
`rand.Deterministic(seed)` (fixed-seed PCG) so that unique ID generation and
any probability-dependent logic produce identical sequences across runs.
Direct calls to `crypto/rand` or `math/rand` are forbidden inside the same
packages.

Inject both via function parameter, not global state. This is constitution
principle IV.

---

## Error types

| Type                                | Returned by                              | Cause                                                                                                     |
|-------------------------------------|------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| `*transport.ErrSubprotocolMismatch` | `Dial`                                   | CSMS selected a subprotocol not in `DialOptions.Subprotocols`, or omitted the response header             |
| `*transport.ErrTLSValidation`       | `Dial`                                   | TLS handshake failed (expired cert, untrusted CA, hostname mismatch); wraps the underlying x509/tls error |
| `*transport.ErrFrameTooLarge`       | `Station.Expect`                         | Inbound frame exceeded `DialOptions.MaxFrameBytes`; the connection remains open                           |
| `*transport.ErrStationClosed`       | `Station.Send`, `Station.Expect`         | The station was already closed before the call                                                            |
| `*wire.ErrFrameShape`               | `ParseCall`, `ParseResult`, `ParseError` | Inbound JSON does not match the expected OCPP-J array shape                                               |

Use `errors.As` to inspect typed transport errors and access structured
fields (e.g., `ErrSubprotocolMismatch.Requested`, `ErrTLSValidation.Cause`).
