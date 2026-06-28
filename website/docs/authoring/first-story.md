---
sidebar_position: 1
---

# Authoring Your First Story

This walkthrough builds a real conformance story from scratch — the
plug-in-first transaction start — explaining each section as it goes. By
the end you will have a story you can validate and run against a CSMS.

## 1. Pick the behavior and find its spec section

Conformance starts with the specification. We will test the **plug-in-first
transaction start**: the driver plugs in *before* presenting a token, the
connector moves to `Preparing`, the station authorizes, then starts the
transaction. That touches three OCPP 1.6 sections: StatusNotification
(§6.13), Authorize (§6.2), and StartTransaction (§6.16).

Every conformance story must record this traceability in `Spec-Ref`.

## 2. Write the `Meta` block

```text
Meta
    Name:        Transaction start plugin-first accepted
    Id:          transaction_pluginfirst_accepted
    Spec-Ref:    OCPP-J 1.6 §6.13 StatusNotification, §6.2 Authorize, §6.16 StartTransaction
    Tags:        transaction, charging, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, idTag, meterStart
    Depends:
      - id:    connector_status_available
        scope: per-station
```

- `Id` is the stable identifier used by dependencies, the cache, and
  reports.
- `Spec-Ref` is mandatory here because this is a conformance story (a
  [helper](../concepts/stories.md) would omit it and add the `helper` tag).
- `Parameters` declares inputs interpolated into steps as `{connectorId}`,
  `"{idTag}"`, and `{meterStart}`.
- `Depends` says this story needs the connector in the `Available` state
  first; the runner runs that prerequisite (and *its* prerequisites)
  automatically. See [dependency graph](../concepts/dependency-graph.md).

## 3. Set up the `Background`

`Background` holds the `Given` preconditions. Some preconditions are
genuinely external — OCTANE cannot provision an `idTag` over the wire
(that would be CSMS-specific adaptation), so it states the requirement and
leaves it to the operator.

```text
Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{idTag}" with status "Accepted"
```

## 4. Write the `Scenario` steps

Each step maps to a [keyword](./keywords-reference.md). `When` performs an
action; `Then` / `And` assert the response.

```text
Scenario: Plug-in precedes authorization; transaction starts cleanly
    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends Authorize with idTag "{idTag}"
    Then  the CSMS responds to Authorize with idTagInfo.status "Accepted" within 30 seconds

    When  station "CP01" starts a transaction on connector {connectorId} with idTag "{idTag}" and meterStart {meterStart}
    Then  the CSMS responds to StartTransaction with idTagInfo.status "Accepted" within 30 seconds
    And   the StartTransaction response assigns a positive transactionId

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Charging"
    Then  the CSMS acknowledges the status within 10 seconds
```

## 5. Clean up with `Teardown`

`Teardown` runs regardless of how the scenario ended.

```text
Teardown
    Disconnect station "CP01"
```

## 6. Validate before running

Parse and structurally check the story without touching a CSMS:

```bash
octane validate stories scenarios/v16/transaction_pluginfirst_accepted.story
```

You will get one line per file:

```text
OK: scenarios/v16/transaction_pluginfirst_accepted.story
```

An invalid file reports `ERROR: <path>: <message>` and the command exits
non-zero.

:::warning Tabs are not allowed
The `.story` parser is whitespace-significant and rejects tab characters.
Indent with spaces. `octane validate` will flag the problem with a line
and column.
:::

## 7. Run it

```bash
octane run scenarios/v16/transaction_pluginfirst_accepted.story \
    --csms-endpoint ws://localhost:9210
```

The runner resolves the `Depends` chain first (connecting and booting the
station), then executes your scenario and prints a summary plus the report
location.

## How a step finds its keyword

If a step does not match any registered keyword, the run fails at preflight
with the closest suggestions. To check a step interactively:

```bash
octane keywords resolve "station \"CP01\" sends Heartbeat"
```

## Next

- **[Keywords reference](./keywords-reference.md)** — the full vocabulary
  of steps.
- **[Multi-station patterns](./multi-station-patterns.md)** — scenarios
  with more than one station.
- **[Story grammar](../reference/story-grammar.md)** — the formal syntax.
