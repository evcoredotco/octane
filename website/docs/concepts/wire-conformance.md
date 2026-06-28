---
sidebar_position: 1
---

# Wire Conformance

Wire conformance is the idea that shapes every other decision in OCTANE:
**a CSMS is conformant if, and only if, it produces the right bytes in the
right order on the OCPP-J WebSocket.** OCTANE validates that and nothing
else.

## Black-box, from the station side

OCTANE connects to your CSMS the same way a real charging station does —
over an OCPP-J WebSocket — and drives the protocol from the station side.
It observes only what the CSMS sends back on that connection.

It never:

- reads CSMS databases, queues, or internal job state;
- calls an administrative or vendor API;
- requires a sidecar, an adapter, or any change to the system under test.

This is **zero-cooperation-cost** adoption: any team can point OCTANE at an
unmodified deployment and get a conformance signal with a single command.
A conformance tool that requires the vendor's cooperation is, in practice,
not a conformance tool.

## What OCTANE checks — and what it cannot

| OCTANE validates | OCTANE does **not** validate |
|---|---|
| The OCPP-J frames the CSMS sends (Call / CallResult / CallError) | CSMS-internal business logic invisible on the wire |
| Field values, enums, and structure against the OCPP 1.6 spec | Billing pipelines, audit-log content, database state |
| Message ordering and timing within a scenario | Anything reachable only through an admin API |
| That responses arrive within declared timeouts | OCPP 2.0.1 / 2.1 behavior (out of scope) |

If a behavior cannot be observed on the wire, OCTANE makes no claim about
it. Scenarios that genuinely require external setup (for example,
provisioning an `idTag` in the CSMS) declare that precondition explicitly
and leave it to the operator.

## A deviation is a finding, not a tolerance

OCTANE has **no** layer for per-CSMS behavioral adaptation. Domain
keywords are identical for every CSMS that implements a given OCPP
version. When a CSMS behaves differently, that difference is reported as a
finding — it cannot be configured away.

Three tempting features were considered and deliberately **rejected** to
protect the integrity of the conformance signal:

- **Vendor-implemented test adapters.** A sidecar the CSMS team writes so
  OCTANE can set up state and read outcomes. Rejected on adoption-cost
  grounds.
- **Per-CSMS keyword overrides.** A third resolution layer where a profile
  could redefine domain keywords for a CSMS's quirks. Rejected: a CSMS
  that needs special handling to pass is, by definition, not conformant.
- **Operator escape hatches for known deviations.** `--accept-deviation`,
  severity overrides, curated deviation registries. Rejected for the same
  reason.

What *is* legitimately CSMS-specific — how to reach the endpoint on the
network — lives in [connection profiles](./profiles.md), which carry
connection metadata only and never conformance logic.

## Relationship to certification

OCTANE does not issue or imply formal conformance certification.
Certification by an external authority is a separate process with its own
scope and rules. OCTANE asserts only that observable wire behavior matches
what the OCPP 1.6 specification requires for the scenarios it exercises —
a precise, automatable, and honest claim.

## Next

- **[Architecture](./architecture.md)** — how the three layers turn a
  story into wire traffic.
- **[Stories](./stories.md)** — the unit of conformance.
- **[Connection profiles](./profiles.md)** — the one place CSMS-specific
  metadata is allowed to live.
