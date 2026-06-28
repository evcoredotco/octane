---
sidebar_position: 3
---

# Multi-Station Patterns

Reach for multiple stations when the conformance behavior requires
independent station state or coordinated wire interactions. This page
collects practical recipes; see the
[multi-station concept](../concepts/multi-station.md) for the model behind
them.

## Declaring and addressing stations

Set the count in `Meta`; the runner allocates `CP01`, `CP02`, …, and steps
address each by handle.

```text
Meta
    Name:     Independent concurrent transactions
    Spec-Ref: OCPP-J 1.6 §6.16 StartTransaction
    Stations: 2
    Timeout:  90s
    Depends:
      - id:    station_boot_accepted
        scope: per-station
```

Because the dependency is `per-station`, `station_boot_accepted` runs once
for `CP01` and once for `CP02` before the scenario body.

## Pattern: sequential coordination

Steps run in declared order by default — ideal when one station's action
must precede another's.

```text
Scenario: Two stations register and report status in turn
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    When  station "CP02" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds

    When  station "CP01" sends StatusNotification for connector 1 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
    When  station "CP02" sends StatusNotification for connector 1 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
```

## Pattern: genuine concurrency

When the scenario must exercise the CSMS under simultaneous load, wrap the
concurrent actions in a `Parallel … End-Parallel` block.

```text
Scenario: Concurrent starts stay independent
    Parallel
        When  station "CP01" starts a transaction on connector 1 with idTag "TAG-A" and meterStart 0
        When  station "CP02" starts a transaction on connector 1 with idTag "TAG-B" and meterStart 0
    End-Parallel

    Then  the CSMS responds to StartTransaction with idTagInfo.status "Accepted" within 30 seconds
    And   the StartTransaction response assigns a positive transactionId
```

:::note Two orders are both valid
Inside a `Parallel` block, the two sends may reach the CSMS in either
order. The report records the order actually observed; do not write
assertions that assume one particular interleaving.
:::

## Choosing a dependency scope

| Use… | When the prerequisite… |
|---|---|
| `per-station` *(default)* | must hold for each station independently (boot, connect, status). |
| `per-run` | is shared by all stations in the run (a one-time fixture). |
| `global` | is shared across the cache validity window. |

Keep prerequisites explicit in `Depends` rather than re-deriving state in
each scenario — it keeps stories small and lets the runner cache and skip
shared setup. See [dependency graph & caching](../concepts/dependency-graph.md).

## Reading multi-station reports

Per-station state, latencies, and wire traces are captured for every
handle, so a report shows exactly what `CP01` and `CP02` each sent and
received. This is invaluable when a concurrency bug only manifests for one
station. See [Reports](../operations/reports.md).

## Next

- **[Multi-station orchestration](../concepts/multi-station.md)** — the
  underlying model.
- **[Keywords reference](./keywords-reference.md)** — the steps used above.
