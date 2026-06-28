---
sidebar_position: 4
---

# Keyword Catalog

A compact, machine-style listing of every registered keyword pattern,
grouped by layer. For the conceptual model (placeholder types, resolution
order, how steps match) see the
[keywords reference](../authoring/keywords-reference.md).

:::tip Your build is the source of truth
`octane keywords list` prints the keywords actually compiled into your
binary, in the form `[<layer>] [<ocpp-version>] <pattern>`. In the current
pre-alpha build that command wires in only the **primitive** layer; the
**domain** patterns below are defined in `pkg/keywords/ocpp16` and are
listed here for authoring reference.
:::

## Primitive layer

Transport-level steps. Use them only when no domain keyword covers the
behavior you need.

```text
[primitive] open a WebSocket to {url:string} as station {station:string}
[primitive] open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}
[primitive] close station {station:string}
[primitive] send raw bytes {bytes:string} on station {station:string}
[primitive] send raw frame {frame:any} on station {station:string}
[primitive] expect any frame on station {station:string} within {timeout:duration}
[primitive] expect a frame of type {messageType:int} on station {station:string} within {timeout:duration}
[primitive] the connection on station {station:string} is open
[primitive] the connection on station {station:string} is closed
[primitive] wait {duration:duration}
```

## Domain layer — lifecycle, boot & status

```text
[domain] the CSMS is reachable
[domain] station {station:string} connects to the CSMS
[domain] the OCPP-J handshake completes within {timeout:duration}
[domain] station {station:string} is in the connected state
[domain] station {station:string} is registered to the CSMS
[domain] station {station:string} sends BootNotification with reason {reason:string}
[domain] the CSMS responds with status {status:string} within {timeout:duration}
[domain] station {station:string} is in the registered state
[domain] the response includes a heartbeatInterval between {min:int} and {max:int}
[domain] the response includes a currentTime in ISO-8601 format
[domain] station {station:string} sends StatusNotification for connector {connectorId:int} with status {status:string}
[domain] the CSMS acknowledges the status within {timeout:duration}
[domain] connector {connectorId:int} of station {station:string} is in state {state:string}
[domain] station {station:string} sends Heartbeat
[domain] the CSMS responds to the Heartbeat within {timeout:duration}
[domain] the Heartbeat response includes a currentTime in ISO-8601 format
[domain] the operator has provisioned id token {idTag:string} with status {status:string}
[domain] Disconnect station {station:string}
```

## Domain layer — authorize, transactions & metering

```text
[domain] station {station:string} sends Authorize with idTag {idTag:string}
[domain] the CSMS responds to Authorize with idTagInfo.status {status:string} within {timeout:duration}
[domain] station {station:string} starts a transaction on connector {connectorId:int} with idTag {idTag:string} and meterStart {meterStart:int}
[domain] the CSMS responds to StartTransaction with idTagInfo.status {status:string} within {timeout:duration}
[domain] the StartTransaction response assigns a positive transactionId
[domain] station {station:string} stops transaction {transactionId:int}
[domain] the CSMS accepts StopTransaction within {timeout:duration}
[domain] station {station:string} sends MeterValues for connector {connectorId:int}
[domain] the CSMS acknowledges MeterValues within {timeout:duration}
```

## Domain layer — CSMS-initiated commands

```text
[domain] the CSMS sends RemoteStartTransaction with connectorId {connectorId:int} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to RemoteStartTransaction with status {status:string}
[domain] the CSMS sends RemoteStopTransaction with transactionId {transactionId:int} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to RemoteStopTransaction with status {status:string}
[domain] the CSMS sends Reset with type {resetType:string} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to Reset with status {status:string}
[domain] the CSMS sends UnlockConnector with connectorId {connectorId:int} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to UnlockConnector with status {status:string}
[domain] the CSMS sends ChangeAvailability with connectorId {connectorId:int} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to ChangeAvailability with status {status:string}
[domain] the CSMS sends GetConfiguration to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to GetConfiguration
[domain] the CSMS sends ChangeConfiguration with key {key:string} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to ChangeConfiguration with status {status:string}
[domain] the CSMS sends ClearCache to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to ClearCache with status {status:string}
[domain] the CSMS sends ReserveNow with connectorId {connectorId:int} and idTag {idTag:string} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds with ReserveNow.conf status {status:string}
[domain] the CSMS accepts the response without error within {timeout:duration}
[domain] the CSMS sends CancelReservation with reservationId {reservationId:int} to station {station:string} within {timeout:duration}
[domain] station {station:string} responds to CancelReservation with status {status:string}
```

## Next

- **[Keywords reference](../authoring/keywords-reference.md)** — types and
  resolution.
- **[OCPP 1.6 coverage](./ocpp-coverage.md)** — which messages these
  keywords exercise.
