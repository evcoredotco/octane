---
sidebar_position: 5
---

# OCPP 1.6 Coverage

OCTANE targets **OCPP 1.6J** exclusively. This page is the honest
inventory of which message types have conformance stories today. "Stories"
counts the `.story` files exercising each message type.

## Covered message types

| OCPP message | Direction | Stories | Notes |
|---|---|---|---|
| BootNotification | CS → CSMS | 2 | Accepted; full boot sequence |
| StatusNotification | CS → CSMS | 1 | Available state |
| Heartbeat | CS → CSMS | 1 | Part of the boot sequence |
| Authorize | CS → CSMS | 1 | Plug-in-first flow |
| StartTransaction | CS → CSMS | 2 | Plug-in-first; identification-first |
| StopTransaction | CS → CSMS | 1 | Normal stop |
| MeterValues | CS → CSMS | 1 | Periodic sampling |
| RemoteStartTransaction | CSMS → CS | 1 | Accepted |
| RemoteStopTransaction | CSMS → CS | 1 | Accepted |
| Reset | CSMS → CS | 2 | Soft; Hard |
| UnlockConnector | CSMS → CS | 2 | Accepted; Failed |
| ChangeAvailability | CSMS → CS | 2 | Operative; Inoperative |
| GetConfiguration | CSMS → CS | 1 | Key list returned |
| ChangeConfiguration | CSMS → CS | 1 | Accepted |
| ClearCache | CSMS → CS | 1 | Accepted |
| ReserveNow | CSMS → CS | 1 | Faulted response |
| CancelReservation | CSMS → CS | 1 | Accepted |

That is **17 message types** across **21 `.story` files** under
`scenarios/v16/` (conformance stories plus the lifecycle helpers they
depend on).

*Direction:* CS → CSMS means the charging station initiates and the CSMS
responds; CSMS → CS means the CSMS initiates a command and OCTANE (as the
station) responds, then asserts the CSMS handled the response correctly.

## Not yet covered

Keywords and stories for these OCPP 1.6 message types have not been written
yet:

- DiagnosticsStatusNotification
- FirmwareStatusNotification
- DataTransfer
- TriggerMessage
- SendLocalList
- GetLocalListVersion

:::note A growing inventory
Coverage expands as keywords and stories are added. The authoritative,
up-to-date list for your checkout is the set of `.story` files under
`scenarios/v16/` plus `octane keywords list`.
:::

## How coverage maps to keywords

Each message type above is exercised by one or more keywords in the
[keyword catalog](./keyword-catalog.md) — typically a `When …` keyword that
sends or receives the message and a `Then …` keyword that asserts the
response status or field. Stories assemble these into spec-traceable
scenarios via [`Spec-Ref`](./story-grammar.md).

## Next

- **[Keyword catalog](./keyword-catalog.md)** — the patterns behind the
  coverage.
- **[Wire conformance](../concepts/wire-conformance.md)** — what "covered"
  means and does not mean.
