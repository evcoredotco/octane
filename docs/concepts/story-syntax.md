# Story DSL Syntax Reference

OCTANE test scenarios are written in `.story` files using a structured
plain-text DSL. Each file describes one test: its metadata, prerequisite
state, and the Gherkin-style steps that exercise the CSMS wire behavior.

---

## File layout

A `.story` file follows this section order:

```
[comments]
Meta
    <key-value entries>

[Background]
    <steps>

[Setup]
    <steps>

Scenario: <title>
    <steps>

[Scenario: <additional-title>]
    <steps>

[Teardown]
    <steps>
```

Rules:
- **Meta** is the only mandatory section and must come first (after optional
  file-level comments).
- At least one **Scenario** is required.
- **Background**, **Setup**, and **Teardown** are optional.
- Sections appear in the order shown above; extra Scenario sections follow
  the first.

---

## Comments

Lines beginning with `#` are comments and are ignored by the parser.
Comments may appear at the top of the file (before `Meta`) and between
sections. Inline comments (trailing `#` on a content line) are not
supported.

```story
# This is a file-level comment explaining the test purpose.

Meta
    Name: My Test
```

---

## Meta section

The Meta section declares structured metadata about the test. Each entry is
indented with exactly four spaces and uses the form `Key: value`.

### Required keys

| Key        | Type           | Description |
|------------|----------------|-------------|
| `Name`     | free text      | Human-readable name for the test (displayed in reports). |
| `Id`       | `snake_case`   | Stable identifier used in `Depends` references. Must be unique across the test suite. |
| `Tags`     | CSV            | Comma-separated list of tags. At least one is required. The `helper` tag is structural (see Spec-Ref rules). |
| `Stations` | integer >= 1   | Number of charging-station handles the test requires. |

### Optional keys

| Key          | Type             | Default          | Description |
|--------------|------------------|------------------|-------------|
| `Spec-Ref`   | free text        | (none)           | OCPP specification section reference (e.g. `OCPP 1.6 §B01 BootNotification`). Required for conformance stories; forbidden for helper stories. |
| `Timeout`    | Go duration      | global default   | Per-step timeout (e.g. `30s`, `2m`). |
| `Parameters` | CSV identifiers  | (none)           | Comma-separated list of parameter names that the story accepts from `octane.yml`. |
| `Cache-TTL`  | Go duration      | 1h (helpers) / ∞ | Cache validity window for this story's result. |
| `Depends`    | YAML list        | (none)           | Prerequisite stories (see Depends block). |

### Spec-Ref and helper-tag rules

- Stories tagged `helper` must **not** have `Spec-Ref`.
- Stories without the `helper` tag **must** have `Spec-Ref`.

```story
# Conformance story — Spec-Ref required, no helper tag.
Meta
    Name:      Boot notification accepted
    Id:        boot_notification_accepted
    Spec-Ref:  OCPP 1.6 §B01 BootNotification
    Tags:      core, boot, wire-only
    Stations:  1
    Timeout:   30s
```

```story
# Helper story — no Spec-Ref, helper tag required.
Meta
    Name:     Station connection established
    Id:       station_connection_established
    Tags:     helper, lifecycle
    Stations: 1
    Timeout:  10s
```

---

## Depends block

The `Depends` key introduces a YAML-style list of prerequisite stories. Each
entry is a bullet starting with `- id:` indented at eight spaces (two levels).

```story
    Depends:
      - id:    station_boot_accepted
        scope: per-station
      - id:    another_prerequisite
        scope: per-run
```

### Depends entry fields

| Field   | Required | Values                              | Default       |
|---------|----------|-------------------------------------|---------------|
| `id`    | yes      | `snake_case` story identifier       | —             |
| `scope` | no       | `per-station`, `per-run`, `global`  | `per-station` |

Scope meanings:
- `per-station` — prerequisite runs once per station handle in the test.
- `per-run` — prerequisite runs once per `octane run` invocation.
- `global` — prerequisite runs once within the result cache validity window.

---

## Step keywords

Steps inside Background, Setup, Scenario, and Teardown sections use standard
Gherkin keywords, indented with four spaces:

| Keyword | Meaning |
|---------|---------|
| `Given` | Establishes a precondition. |
| `When`  | Describes an action or event. |
| `Then`  | Asserts an expected outcome. |
| `And`   | Continues the preceding keyword's intent. |
| `But`   | Introduces a negative continuation. |

Teardown sections may also use bare action lines (no keyword prefix) for
cleanup commands such as `Disconnect station "CP01"`.

```story
Scenario: Station registers with the CSMS
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   station "CP01" is in the registered state
```

---

## Parameters

Parameters allow story steps to reference values provided at runtime via
`octane.yml`. Declare parameter names with `Parameters:` and reference them
in step text with `{name}` syntax.

```story
Meta
    Name:       Connector reservation faulted
    Id:         connector_reservation_faulted
    Spec-Ref:   OCPP-J 1.6 §6.40 ReserveNow
    Tags:       reservation, wire-only
    Stations:   1
    Parameters: connectorId, idTag

Scenario: CSMS handles a Faulted reservation response
    When  the CSMS sends ReserveNow with connectorId {connectorId}
          and idTag "{idTag}" to station "CP01"
    Then  station "CP01" responds with ReserveNow.conf status "Faulted"
```

Every `{placeholder}` in step text must be declared in `Parameters:`. The
parser reports an error for any undeclared placeholder.

---

## Parallel blocks (reserved)

The `Parallel` / `End-Parallel` keyword pair is reserved for future
concurrent multi-station orchestration. Steps inside a Parallel block are
currently flattened into the enclosing scenario's step list.

```story
Scenario: Concurrent authorization
    When  station "CP01" starts a transaction
    Parallel
        When  station "CP02" sends Authorize
    End-Parallel
    Then  the CSMS rejects "CP02" with status "ConcurrentTx"
```
