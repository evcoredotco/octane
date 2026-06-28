---
sidebar_position: 6
---

# Multi-Station Orchestration

Many stateful OCPP scenarios are only reachable on the wire by
coordinating two or more charging stations — for example, confirming that
a CSMS keeps two transactions independent, or that a broadcast command
reaches every connected station. OCTANE makes multiple stations a
first-class feature.

## Declaring stations

A story declares how many stations it needs in `Meta`:

```text
Meta
    Name:     Concurrent transactions stay independent
    Spec-Ref: OCPP-J 1.6 §6.16 StartTransaction
    Stations: 2
```

At preflight the runner allocates handles `CP01`, `CP02`, … and connects
each to its own WebSocket (the endpoint with a per-station path appended;
see [connection profiles](./profiles.md)). Steps then reference stations
by handle:

```text
When  station "CP01" sends BootNotification with reason "PowerUp"
When  station "CP02" sends BootNotification with reason "PowerUp"
```

## Sequential by default, concurrent on request

Steps run **sequentially in declared order** — predictable and easy to
reason about. When a scenario genuinely needs concurrency on the wire, opt
in with a `Parallel … End-Parallel` block:

```text
Parallel
    When  station "CP01" sends StartTransaction
    When  station "CP02" sends StartTransaction
End-Parallel
```

Inside the block the enclosed steps are dispatched concurrently.

## Per-station prerequisites

The default `per-station` dependency scope means a prerequisite runs once
for **each** station handle. A story with `Stations: 2` that depends on
`station_boot_accepted` (per-station) boots both `CP01` and `CP02` before
the scenario body runs. Use `per-run` for setup shared across all stations
and `global` for setup shared across the cache window — see
[dependency graph & caching](./dependency-graph.md).

## Reporting and determinism

Per-station state, per-step latencies, and wire traces are recorded for
**every** station, so a report shows exactly what each handle sent and
received.

:::note Determinism modulo concurrency
OCTANE is deterministic — same inputs, same outputs — *except* for the
ordering of genuinely concurrent operations. Two sends inside a `Parallel`
block may reach the CSMS in either order; the report records both observed
orders rather than pretending to a single canonical one.
:::

## Next

- **[Multi-station patterns](../authoring/multi-station-patterns.md)** —
  practical recipes.
- **[Dependency graph & caching](./dependency-graph.md)** — scopes and
  resolution.
