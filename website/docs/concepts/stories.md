---
sidebar_position: 3
---

# Stories

A **story** is a single OCPP conformance scenario, written as a `.story`
file in OCTANE's Gherkin-flavored DSL. It is both the human-readable
specification of a test and the exact thing the runner executes — there is
no hidden test code behind it.

## Anatomy of a `.story` file

A story is made of sections. `Meta` and `Scenario` are required;
`Background`, `Setup`, and `Teardown` are optional.

```text
Meta
    Name:        Cold boot sequence accepted
    Id:          boot_sequence_accepted
    Spec-Ref:    OCPP-J 1.6 §6.5 BootNotification, §6.7 Heartbeat
    Tags:        boot, lifecycle, wire-only
    Stations:    1
    Timeout:     90s
    Depends:
      - id:    station_connection_established
        scope: per-station

Background
    Given the CSMS is reachable

Scenario: Station completes the cold-boot registration sequence
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   the response includes a heartbeatInterval between 30 and 86400
    And   the response includes a currentTime in ISO-8601 format
    When  station "CP01" sends Heartbeat
    Then  the CSMS responds to the Heartbeat within 10 seconds

Teardown
    Disconnect station "CP01"
```

| Section | Required | Purpose |
|---|---|---|
| `Meta` | yes | Identity, traceability, declarations (stations, dependencies, parameters). |
| `Background` | no | `Given` preconditions shared by the scenario. |
| `Setup` | no | Steps that run before the scenario. |
| `Scenario` | yes | The `When`/`Then`/`And` steps that drive and assert. |
| `Teardown` | no | Cleanup steps, run regardless of outcome. |

:::warning Indentation is significant — and tabs are forbidden
The parser is whitespace-sensitive and rejects tab characters. Indent with
spaces. Run `octane validate stories` to catch structural problems before
a run.
:::

## Meta keys

| Key | Meaning |
|---|---|
| `Name` | Human-readable title. |
| `Id` | Stable identifier; used for dependencies, the cache key, and reports. |
| `Spec-Ref` | The OCPP section this story asserts. **Required for conformance stories.** |
| `Tags` | Comma-separated classification (see below). |
| `Stations` | How many simulated stations the runner allocates (`CP01`, `CP02`, …). |
| `Timeout` | Overall scenario budget (Go duration, e.g. `30s`, `90s`). |
| `Parameters` | Story inputs interpolated into steps as `{name}`. |
| `Depends` | Prerequisite stories, each with an `id` and a `scope`. |
| `Cache-TTL` | Optional time-based cache invalidation for this story. |

## Tags

Tags classify a story. Recognized tags include `wire-only`,
`multi-station`, `operator-assisted`, `helper`, and `pure-protocol`,
alongside free-form domain tags such as `boot`, `lifecycle`,
`transaction`, `reservation`, `csms-initiated`, and `charging`.

## Conformance stories vs. helper stories

Stories come in two kinds, and the parser enforces the distinction:

- A **conformance story** asserts conformance to a specification section
  and **must** carry a `Spec-Ref`.
- A **helper story** exists only to bring the system to a known state for
  other stories to depend on. It **must omit** `Spec-Ref` and **must** be
  tagged `helper`.

```text
Meta
    Name:      Station connection established
    Id:        station_connection_established
    Tags:      helper, lifecycle, wire-only
    Stations:  1
    Timeout:   10s

Scenario: Station opens an OCPP-J WebSocket to the CSMS
    When  station "CP01" connects to the CSMS
    Then  the OCPP-J handshake completes within 5 seconds
    And   station "CP01" is in the connected state
```

Helpers and conformance stories live side by side under `scenarios/` —
there is no separate `helpers/` directory.

## Parameters

Declaring `Parameters` lets a story vary by deployment. Each name is
interpolated where it appears in steps — bare for numbers (`{connectorId}`)
or quoted for strings (`"{idTag}"`):

```text
Meta
    Parameters:  connectorId, idTag, meterStart

Scenario: ...
    When  station "CP01" sends Authorize with idTag "{idTag}"
```

## Dependencies

Stories rarely stand alone — a reservation test needs a registered
station, which needs an open connection. The `Depends` block declares
prerequisites and the runner resolves them into a DAG. See
[Dependency graph & caching](./dependency-graph.md) for scopes
(`per-station`, `per-run`, `global`) and how failures propagate.

## Next

- **[Author your first story](../authoring/first-story.md)** — a guided
  walkthrough.
- **[Story grammar](../reference/story-grammar.md)** — the formal syntax
  reference.
- **[Keywords reference](../authoring/keywords-reference.md)** — the steps
  you can write.
