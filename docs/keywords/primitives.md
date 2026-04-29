# Primitive keyword catalog

Primitive keywords are the transport-level building blocks of OCTANE's story
DSL (spec 004). They speak only OCPP-J framing — opening WebSocket connections,
sending and receiving frames, asserting connection state, and sleeping the
deterministic clock. They carry no OCPP message semantics.

**Version-agnostic.** Primitive keywords are eligible for every story
regardless of the OCPP version declared in the story's `Meta` block. They are
registered at `api.LayerPrimitive` with a zero `OCPPVersion` value.

**Domain keywords take precedence.** When a domain keyword registered for the
story's declared OCPP version has the same pattern as a primitive, the domain
keyword wins. Primitives serve as fallbacks and composable building blocks, not
as the primary authoring surface.

**Import activation.** Import `pkg/keywords/primitive` (blank import is
sufficient) to activate all primitives. Registration happens at `init()` time;
no explicit setup call is required.

---

## Timeout behavior

Several keywords accept a `{timeout:duration}` argument. The deadline is always
derived from `state.Now().Add(timeout)` — never from `time.Now()`. In
deterministic-clock mode (the default in test suites) no real wall-clock time
elapses.

When the deadline passes before a matching frame arrives, the keyword returns
`*primitive.ErrTimeout`. Use `errors.As` to inspect the fields:

```go
var te *primitive.ErrTimeout
if errors.As(err, &te) {
    fmt.Println("station:", te.Station)   // handle name
    fmt.Println("timeout:", te.Timeout)   // configured duration
    fmt.Println("deadline:", te.Deadline) // state.Now() + timeout
}
```

---

## Primitive catalog

### 1. Open WebSocket

```text
open a WebSocket to {url:string} as station {station:string}
```

| Argument    | Type     | Description                                        |
|-------------|----------|----------------------------------------------------|
| `url`       | `string` | WebSocket URL, e.g. `ws://localhost:9000/CP001`    |
| `station`   | `string` | Handle name to register the connection under       |

Dials the given URL with no subprotocol preference. On success, registers the
live `api.Station` in runtime state under `station` so that subsequent steps
can reference it by name.

Returns a wrapped error if the dial fails (unreachable host, TLS error, etc.).

**Example story step:**

```text
Given open a WebSocket to ws://csms.example.com/CP001 as station CP001
```

---

### 2. Open WebSocket with subprotocol

```
open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}
```

| Argument      | Type     | Description                                      |
|---------------|----------|--------------------------------------------------|
| `url`         | `string` | WebSocket URL                                    |
| `station`     | `string` | Handle name to register the connection under     |
| `subprotocol` | `string` | OCPP subprotocol identifier, e.g. `ocpp1.6`     |

Identical to primitive 1 but offers a single subprotocol during the WebSocket
handshake. Returns an error if the server does not agree on the requested
subprotocol.

**Example story step:**

```
Given open a WebSocket to ws://csms.example.com/CP001 as station CP001 with subprotocol ocpp1.6
```

---

### 3. Close station

```
close station {station:string}
```

| Argument  | Type     | Description                                          |
|-----------|----------|------------------------------------------------------|
| `station` | `string` | Handle name of the station to close                  |

Looks up the named station in runtime state and calls `Close()` on its
WebSocket connection. Returns an error if the handle is not registered or if
the underlying close operation fails.

**Example story step:**

```
Then close station CP001
```

---

### 4. Send raw frame

```
send raw frame {frame:any} on station {station:string}
```

| Argument  | Type    | Description                                                  |
|-----------|---------|--------------------------------------------------------------|
| `frame`   | `any`   | A `[]any` value — the decoded Go form of an OCPP-J JSON array |
| `station` | `string`| Handle name of the target station                            |

Encodes the `[]any` frame via the transport layer and writes it to the
station's WebSocket connection. The frame element layout follows ADR 0006:
JSON arrays decode to `[]any`, JSON numbers to `float64`.

Returns an error if:
- `frame` is not a `[]any` (wrong Go type),
- the station handle is not registered, or
- the underlying send fails.

**Example story step:**

```
When send raw frame [2,"msg-1","BootNotification",{}] on station CP001
```

---

### 5. Send raw bytes

```
send raw bytes {bytes:string} on station {station:string}
```

| Argument  | Type     | Description                                              |
|-----------|----------|----------------------------------------------------------|
| `bytes`   | `string` | Hex-encoded byte string, e.g. `5b322c226964225d`        |
| `station` | `string` | Handle name of the target station                        |

Decodes the hex string to bytes, parses the bytes as a JSON array into a
`[]any` value, and sends it via the station's `Send` method. Intended for
negative-path conformance testing — constructing deliberately malformed or
extension OCPP-J frames (spec 004 OQ1).

Returns an error if:
- the hex string is invalid,
- the decoded bytes do not parse as a JSON array, or
- the underlying send fails.

**Example story step:**

```
When send raw bytes 5b322c226964222c22426f6f744e6f74696669636174696f6e222c7b7d5d on station CP001
```

---

### 6. Expect any frame

```
expect any frame on station {station:string} within {timeout:duration}
```

| Argument  | Type       | Description                                          |
|-----------|------------|------------------------------------------------------|
| `station` | `string`   | Handle name of the station to read from              |
| `timeout` | `duration` | Maximum wait duration, e.g. `5s`, `500ms`            |

Blocks until one OCPP-J frame arrives on the station's WebSocket connection or
the deadline elapses. The deadline is `state.Now().Add(timeout)`.

Returns `nil` on success. Returns `*ErrTimeout` if the deadline elapses before
any frame arrives. Returns a wrapped station error for other read failures.

**Example story step:**

```
Then expect any frame on station CP001 within 5s
```

---

### 7. Expect frame of type

```
expect a frame of type {messageType:int} on station {station:string} within {timeout:duration}
```

| Argument      | Type       | Description                                        |
|---------------|------------|----------------------------------------------------|
| `messageType` | `int`      | OCPP-J message-type code: `2` (CALL), `3` (CALLRESULT), `4` (CALLERROR) |
| `station`     | `string`   | Handle name of the station to read from            |
| `timeout`     | `duration` | Maximum wait duration for the matching frame       |

Loops over received frames under a single shared deadline
(`state.Now().Add(timeout)`). Frames whose first element does not equal
`messageType` are silently discarded and the loop continues. The first
matching frame causes the keyword to return `nil`.

Returns `*ErrTimeout` if no matching frame arrives before the deadline.
Returns a wrapped station error for non-timeout read failures.

**Example story step:**

```
Then expect a frame of type 3 on station CP001 within 10s
```

---

### 8. Wait

```
wait {duration:duration}
```

| Argument   | Type       | Description                              |
|------------|------------|------------------------------------------|
| `duration` | `duration` | How long to pause, e.g. `1s`, `200ms`   |

Calls `state.Sleep(ctx, duration)`. In deterministic-clock mode the injected
clock advances without any real wall-clock delay (constitution principle IV,
spec 004 AC5). In production mode the runner's real clock is used.

Returns a wrapped context error if the context is cancelled before the duration
elapses.

**Example story step:**

```
When wait 2s
```

---

### 9. Assert connection is open

```
the connection on station {station:string} is open
```

| Argument  | Type     | Description                                          |
|-----------|----------|------------------------------------------------------|
| `station` | `string` | Handle name of the station to check                  |

Asserts that the named station's WebSocket connection is currently open by
calling `IsOpen()`. Fails (returns an error) if the handle is not registered or
if `IsOpen()` returns `false`.

**Example story step:**

```
Then the connection on station CP001 is open
```

---

### 10. Assert connection is closed

```
the connection on station {station:string} is closed
```

| Argument  | Type     | Description                                          |
|-----------|----------|------------------------------------------------------|
| `station` | `string` | Handle name of the station to check                  |

Asserts that the named station's WebSocket connection is currently closed by
calling `IsOpen()`. Fails (returns an error) if the handle is not registered or
if `IsOpen()` returns `true`.

**Example story step:**

```
Then the connection on station CP001 is closed
```

---

## Quick-reference table

| # | Pattern | File |
|---|---------|------|
| 1 | `open a WebSocket to {url:string} as station {station:string}` | `open.go` |
| 2 | `open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}` | `open.go` |
| 3 | `close station {station:string}` | `close.go` |
| 4 | `send raw frame {frame:any} on station {station:string}` | `send.go` |
| 5 | `send raw bytes {bytes:string} on station {station:string}` | `send.go` |
| 6 | `expect any frame on station {station:string} within {timeout:duration}` | `expect.go` |
| 7 | `expect a frame of type {messageType:int} on station {station:string} within {timeout:duration}` | `expect.go` |
| 8 | `wait {duration:duration}` | `wait.go` |
| 9 | `the connection on station {station:string} is open` | `status.go` |
| 10 | `the connection on station {station:string} is closed` | `status.go` |

---

## Related

- [docs/concepts/keywords.md](../concepts/keywords.md) — two-layer model and resolver rules.
- [ADR 0007](../adr/0007-keyword-library-layering.md) — design rationale for layering and precedence.
- `pkg/keywords/primitive/` — Go source for all primitives.
- `pkg/keywords/api/` — `Func`, `State`, `Station`, `Args` interfaces.
- `pkg/keywords/api/mock/` — test doubles for unit-testing keyword libraries.
