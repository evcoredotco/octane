# ADR 0006: `.story` Gherkin-Flavored DSL

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect
- **Constitution principles touched:** VI (Test Cases as Code), XII
  (Declarative scenarios)

## Context

ADR 0005 established that conformance scenarios are written as `.story`
files in a Gherkin-flavored DSL. This ADR pins the grammar, file
layout, and parsing rules so authors and tooling agree on a single
shape.

The grammar must:

- Be human-readable enough that certification reviewers can read it
  without knowing Go.
- Be machine-parseable with a small, deterministic Go parser
  (no regex spaghetti, no third-party DSL framework — constitution V).
- Carry OCPP specification traceability metadata in a structured,
  optional header (mandatory for conformance tests, omitted for
  helper stories).
- Support multi-station scenarios (ADR 0008) without fighting the
  grammar.
- Be diffable and Git-friendly.

## Decision

### File layout

A `.story` file has three required sections in fixed order, each
introduced by a top-level keyword on its own line:

```
Meta
    Name:        Boot notification with accepted registration
    Id:          boot_notification_accepted
    Spec-Ref:    OCPP 1.6 §B01 BootNotification
    Tags:        core, boot, wire-only
    Stations:    1
    Timeout:     30s

Background
    Given the CSMS is reachable
    And   the profile defines station "CP01"

Scenario: Accepted registration on first boot
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   the response includes a heartbeatInterval between 30 and 86400
    And   the response interval is a positive integer
```

Optional `Setup` and `Teardown` sections may follow `Background` and
the last `Scenario`, respectively.

### Grammar (BNF, condensed)

```
story        = meta_section background? setup? scenario+ teardown?
meta_section = "Meta" NEWLINE meta_entry+
meta_entry   = INDENT IDENT ":" value NEWLINE
background   = "Background" NEWLINE step+
setup        = "Setup" NEWLINE step+
scenario     = "Scenario" ":" text NEWLINE step+
teardown     = "Teardown" NEWLINE step+
step         = INDENT step_keyword text NEWLINE
step_keyword = "Given" | "When" | "Then" | "And" | "But"
```

Indentation is whitespace-significant: section bodies are indented by
exactly four spaces. Tab characters are forbidden by the parser to
keep diffs deterministic.

### Meta keys

| Key | Required | Format | Purpose |
|-----|----------|--------|---------|
| `Name` | yes | free text | Human-readable test name |
| `Id` | yes | snake_case slug | Stable identifier for `Depends:` references |
| `Spec-Ref` | conformance only | `OCPP-<version> §<section> <message>` | OCPP specification traceability (constitution principle I) |
| `Tags` | yes | comma list | At least one of `wire-only`, `multi-station`, `operator-assisted`, `helper` |
| `Stations` | yes | integer ≥ 1 | Declared station count for preflight resource allocation |
| `Timeout` | no | duration | Default per-step timeout; overrides `--default-timeout` |
| `Depends` | no | YAML list | Other story IDs this story depends on (per ADR 0015) |

`Spec-Ref` is required for conformance tests and forbidden for helper
stories tagged `helper`. The parser enforces this distinction.

The parser rejects a story missing any required Meta key.

### Step text and parameters

Step text is matched against keyword patterns registered in the
keyword library (ADR 0007). Patterns use `{name:type}` placeholders:

```go
Keyword: "the CSMS responds with status {status:string} within {timeout:duration}"
```

A story line:

```
Then the CSMS responds with status "Accepted" within 30 seconds
```

binds `status="Accepted"` and `timeout=30s`. The parser does not
resolve keywords; it tokenizes and hands tokens to the runner.
Resolution happens at run time so unknown keywords surface as a single,
clear error: `unknown keyword: "the CSMS responds with status"`.

### Quoting and types

- Double-quoted strings are literal text. Newlines forbidden.
- Bare integers and floats are numeric.
- `"Accepted"` is a string; `30 seconds` is a duration; `42` is an integer.
- Booleans are `true` / `false`, lowercase.
- Lists use square brackets: `["a", "b", "c"]`.

### Comments

Lines starting with `#` are comments. Trailing comments are not
permitted (forces clean, diffable lines).

### File extension and discovery

- Extension: `.story`.
- Default discovery root: `scenarios/`.
- Scenarios for OCPP 1.6 live under `scenarios/`, etc.
- A story file MAY contain only one `Scenario:` block in v1 to keep
  the data model simple. Multi-scenario files are an ADR-level
  decision deferred to v1.1.

## Consequences

### Positive

- The DSL is small enough that a recursive-descent parser fits in
  `pkg/story/parser.go` without external dependencies.
- The grammar is unambiguous: every line type is determined by its
  prefix.
- Whitespace-significant indentation makes diffs clean and stable.
- OCPP specification traceability is structurally enforced by the
  Meta section.

### Negative

- Multi-line step values (e.g. embedded JSON payloads) require an
  extension. Reserved syntax (heredoc-style `"""`) is documented as
  future work, not implemented in v1.
- Indentation strictness will frustrate authors who use tab-indented
  editors. Documented in `docs/keywords/authoring.md`.

### Neutral

- The grammar borrows from Gherkin but is not Gherkin-compatible. We
  do not promise interoperability with `cucumber-go` or `godog`. This
  is intentional: drifting from Gherkin where it serves OCPP-specific
  needs (Stations, Timeout) is preferable to a half-Gherkin compromise.

## Alternatives considered

- **Strict Gherkin (godog/cucumber).** Rejected: forces meta into tags
  or comments, no first-class Stations declaration.
- **TOML or YAML scenarios.** Rejected per ADR 0005.
- **Custom binary format.** Rejected: hostile to certification reviewers.

## References

- Constitution: principles VI, XII
- ADR 0005 (story-driven framework)
- ADR 0007 (keyword library)
- Robot Framework `.robot` syntax: https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html#test-data-syntax
