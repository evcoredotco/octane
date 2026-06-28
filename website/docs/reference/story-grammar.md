---
sidebar_position: 3
---

# Story Grammar

This page is the formal syntax reference for `.story` files. For a
conceptual tour see [Stories](../concepts/stories.md); for a guided build
see [Authoring your first story](../authoring/first-story.md).

## Lexical rules

- A story is a UTF-8 text file with the `.story` extension.
- **Indentation is significant.** Nesting is expressed with leading
  spaces.
- **Tab characters are forbidden** anywhere in the file. The parser
  rejects them.
- Lines beginning with `#` are comments.
- The parser is a hand-written recursive-descent parser with no
  third-party dependency. On error it reports a line and column.

## Document structure

```text
story        = [ comment-block ] meta-section
               [ background-section ]
               [ setup-section ]
               scenario-section
               [ teardown-section ]
```

| Section | Required | Contents |
|---|---|---|
| `Meta` | yes | Key/value declarations (see below). |
| `Background` | no | `Given` steps shared by the scenario. |
| `Setup` | no | Steps run before the scenario. |
| `Scenario:` | yes | A title line plus `When`/`Then`/`And` steps. |
| `Teardown` | no | Cleanup steps, run regardless of outcome. |

## `Meta` keys

```text
Meta
    Name:        <human-readable title>
    Id:          <stable identifier>
    Spec-Ref:    <OCPP section(s)>          # required for conformance stories
    Tags:        <tag>, <tag>, …
    Stations:    <integer>
    Timeout:     <duration>                 # e.g. 30s, 90s
    Parameters:  <name>, <name>, …
    Cache-TTL:   <duration>                 # optional
    Depends:
      - id:    <story-id>
        scope: per-station | per-run | global
```

| Key | Required | Notes |
|---|---|---|
| `Name` | yes | Display title. |
| `Id` | yes | Used by dependencies, cache key, and reports. |
| `Spec-Ref` | conditional | **Required** for conformance stories; **forbidden** for `helper` stories. |
| `Tags` | no | Comma-separated; see [tags](#tags). |
| `Stations` | no | Station handles allocated as `CP01`, `CP02`, …; defaults to 1. |
| `Timeout` | no | Overall scenario budget (Go duration). |
| `Parameters` | no | Names interpolated into steps as `{name}`. |
| `Cache-TTL` | no | Time-based cache invalidation for this story. |
| `Depends` | no | List of prerequisites, each an `id` + `scope`. |

### The helper/conformance rule

The parser enforces a hard invariant:

- a story **tagged `helper`** must **omit** `Spec-Ref`;
- a story **not tagged `helper`** must **include** `Spec-Ref`.

This keeps the conformance set honest: only stories that cite a
specification section count as conformance tests.

## Tags

Recognized classification tags include `wire-only`, `multi-station`,
`operator-assisted`, `helper`, and `pure-protocol`. Free-form domain tags
(`boot`, `lifecycle`, `transaction`, `reservation`, `csms-initiated`,
`charging`, …) are also allowed and useful for filtering.

## Steps

```text
step = ( "Given" | "When" | "Then" | "And" ) WS step-text
```

- `Given` introduces a precondition (typically in `Background`).
- `When` performs an action.
- `Then` / `And` assert an outcome.

Each step's text is matched against the [keyword](../authoring/keywords-reference.md)
registry. A step that matches no keyword fails preflight.

### Parameter interpolation

Parameters declared in `Meta` are substituted where they appear in step
text: bare for numeric placeholders, quoted for strings.

```text
Meta
    Parameters:  connectorId, idTag

Scenario: ...
    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    And   station "CP01" sends Authorize with idTag "{idTag}"
```

## `Depends` scopes

| Scope | Behavior |
|---|---|
| `per-station` *(default)* | Runs once per station handle. |
| `per-run` | Runs once per run. |
| `global` | Runs once across the cache validity window. |

See [dependency graph & caching](../concepts/dependency-graph.md) for how
the runner resolves and caches dependencies.

## Complete example

```text
# A conformance story with a dependency and parameters.
Meta
    Name:        Connector reservation faulted
    Id:          connector_reservation_faulted
    Spec-Ref:    OCPP-J 1.6 §6.40 ReserveNow
    Tags:        reservation, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId, idTag
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable

Scenario: CSMS handles a Faulted reservation response
    When  the CSMS sends ReserveNow with connectorId {connectorId} and idTag "{idTag}" to station "CP01" within 30 seconds
    Then  station "CP01" responds with ReserveNow.conf status "Faulted"
    And   the CSMS accepts the response without error within 10 seconds

Teardown
    Disconnect station "CP01"
```

## Next

- **[Keywords reference](../authoring/keywords-reference.md)** — the step
  vocabulary.
- **[Stories](../concepts/stories.md)** — the conceptual overview.
