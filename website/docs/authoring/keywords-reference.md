---
sidebar_position: 2
---

# Keywords Reference

Keywords are the executable vocabulary of stories. Each `Given` / `When` /
`Then` / `And` step is matched to a keyword — a typed Go function that
performs one wire action and asserts the result.

## The keyword model

A keyword is a `Pattern` + `Func` pair registered against a layer and an
OCPP version. Patterns contain typed placeholders written `{name:type}`:

```go
registry.Register(api.Keyword{
    Layer:       api.LayerDomain,
    OCPPVersion: api.OCPP16,
    Pattern: "station {station:string} sends BootNotification " +
        "with reason {reason:string}",
    Func: sendBootNotification,
})
```

### Placeholder types

| Type | Matches | Example value |
|---|---|---|
| `string` | a quoted or bare token | `"CP01"`, `Accepted` |
| `int` | an integer | `1`, `86400` |
| `float` | a decimal | `3.14` |
| `bool` | `true` / `false` | `true` |
| `duration` | a Go duration | `30s`, `5m` |
| `station` | a station handle | `CP01` |
| `any` | an opaque value (primitive layer) | a raw frame |

## Two layers, deterministic resolution

Keywords live in two sub-layers (see [architecture](../concepts/architecture.md)):

- **Domain** — OCPP 1.6 message semantics (the bulk of what you write).
- **Primitive** — raw transport, an escape hatch for behavior the domain
  layer does not cover.

For each step, OCTANE matches **domain keywords first** (scoped to the
story's OCPP version); if none match, it falls back to **primitive**
keywords. A step that matches neither fails at preflight with a diagnostic
that lists the layers searched and the nearest registered patterns by edit
distance.

:::tip Inspect the registry at runtime
`octane keywords list` is the authoritative catalog for *your* build — it
prints `[<layer>] [<ocpp-version>] <pattern>` for every registered keyword.
`octane keywords resolve "<step text>"` shows which pattern a step resolves
to, or the closest suggestion.

In the current pre-alpha build, `octane keywords list` wires in only the
**primitive** layer; the domain catalog below is defined in source
(`pkg/keywords/ocpp16`). Treat `octane keywords list` as the source of
truth for what is actually loaded.
:::

## Domain keywords — lifecycle, boot & status

| Pattern |
|---|
| `the CSMS is reachable` |
| `station {station:string} connects to the CSMS` |
| `the OCPP-J handshake completes within {timeout:duration}` |
| `station {station:string} is in the connected state` |
| `station {station:string} is registered to the CSMS` |
| `station {station:string} sends BootNotification with reason {reason:string}` |
| `the CSMS responds with status {status:string} within {timeout:duration}` |
| `station {station:string} is in the registered state` |
| `the response includes a heartbeatInterval between {min:int} and {max:int}` |
| `the response includes a currentTime in ISO-8601 format` |
| `station {station:string} sends StatusNotification for connector {connectorId:int} with status {status:string}` |
| `the CSMS acknowledges the status within {timeout:duration}` |
| `connector {connectorId:int} of station {station:string} is in state {state:string}` |
| `station {station:string} sends Heartbeat` |
| `the CSMS responds to the Heartbeat within {timeout:duration}` |
| `the Heartbeat response includes a currentTime in ISO-8601 format` |
| `the operator has provisioned id token {idTag:string} with status {status:string}` |
| `Disconnect station {station:string}` |

## Domain keywords — authorize, transactions & metering

| Pattern |
|---|
| `station {station:string} sends Authorize with idTag {idTag:string}` |
| `the CSMS responds to Authorize with idTagInfo.status {status:string} within {timeout:duration}` |
| `station {station:string} starts a transaction on connector {connectorId:int} with idTag {idTag:string} and meterStart {meterStart:int}` |
| `the CSMS responds to StartTransaction with idTagInfo.status {status:string} within {timeout:duration}` |
| `the StartTransaction response assigns a positive transactionId` |
| `station {station:string} stops transaction {transactionId:int}` |
| `the CSMS accepts StopTransaction within {timeout:duration}` |
| `station {station:string} sends MeterValues for connector {connectorId:int}` |
| `the CSMS acknowledges MeterValues within {timeout:duration}` |

## Domain keywords — CSMS-initiated commands

| Pattern |
|---|
| `the CSMS sends RemoteStartTransaction with connectorId {connectorId:int} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to RemoteStartTransaction with status {status:string}` |
| `the CSMS sends RemoteStopTransaction with transactionId {transactionId:int} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to RemoteStopTransaction with status {status:string}` |
| `the CSMS sends Reset with type {resetType:string} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to Reset with status {status:string}` |
| `the CSMS sends UnlockConnector with connectorId {connectorId:int} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to UnlockConnector with status {status:string}` |
| `the CSMS sends ChangeAvailability with connectorId {connectorId:int} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to ChangeAvailability with status {status:string}` |
| `the CSMS sends GetConfiguration to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to GetConfiguration` |
| `the CSMS sends ChangeConfiguration with key {key:string} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to ChangeConfiguration with status {status:string}` |
| `the CSMS sends ClearCache to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to ClearCache with status {status:string}` |
| `the CSMS sends ReserveNow with connectorId {connectorId:int} and idTag {idTag:string} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds with ReserveNow.conf status {status:string}` |
| `the CSMS accepts the response without error within {timeout:duration}` |
| `the CSMS sends CancelReservation with reservationId {reservationId:int} to station {station:string} within {timeout:duration}` |
| `station {station:string} responds to CancelReservation with status {status:string}` |

## Primitive keywords — transport escape hatch

Use these only when the domain layer does not cover what you need. They
operate directly on the WebSocket and OCPP-J frames.

| Pattern |
|---|
| `open a WebSocket to {url:string} as station {station:string}` |
| `open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}` |
| `close station {station:string}` |
| `send raw bytes {bytes:string} on station {station:string}` |
| `send raw frame {frame:any} on station {station:string}` |
| `expect any frame on station {station:string} within {timeout:duration}` |
| `expect a frame of type {messageType:int} on station {station:string} within {timeout:duration}` |
| `the connection on station {station:string} is open` |
| `the connection on station {station:string} is closed` |
| `wait {duration:duration}` |

## Next

- **[Keyword catalog](../reference/keyword-catalog.md)** — the same
  vocabulary in a compact, machine-style listing.
- **[Author your first story](./first-story.md)** — see keywords in
  context.
- **[OCPP 1.6 coverage](../reference/ocpp-coverage.md)** — which messages
  have stories today.
