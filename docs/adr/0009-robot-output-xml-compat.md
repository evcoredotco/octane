# ADR 0009: Robot Framework `output.xml` Compatibility

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, DevOps, Docs
- **Constitution principles touched:** II (Two Distribution Surfaces)

## Context

ADR 0005 borrows Robot Framework's structural model but implements
OCTANE natively in Go. Robot Framework's *runtime* dependency is
explicitly rejected (Python in CI, dependency tree).

However, Robot Framework's **output format** (`output.xml`) is a
de-facto industry standard consumed by:

- Allure
- ReportPortal
- Jenkins (Robot Framework plugin)
- GitLab and GitHub Actions test reporters
- Many in-house QA dashboards
- IDE plugins for IntelliJ and VS Code

Producing this format alongside OCTANE's native JSON report unlocks a
large reporting ecosystem for free, without taking a runtime
dependency on Robot Framework itself.

`pytest --junit-xml=results.xml` is direct prior art: pytest does not
depend on JUnit but emits the JUnit XML format. Jenkins, GitLab, and
GitHub Actions all consume it transparently.

## Decision

OCTANE emits two report artifacts per run by default:

| File | Format | Purpose |
|------|--------|---------|
| `report.json` | OCTANE-native JSON, byte-deterministic | Source of truth, certification submissions |
| `output.xml` | Robot Framework 7.x output schema | Ecosystem integration |

The XML emitter:

- Lives in `pkg/report/robotxml/`.
- Is generated from the same in-memory report tree as `report.json`,
  not from a re-parse of the JSON. This guarantees the two artifacts
  agree.
- Is opt-out via `--no-robot-xml`, but on-by-default.
- Does not pull in any third-party dependency. The format is small
  enough to emit with `encoding/xml`.

### Schema scope

OCTANE emits the subset of `output.xml` required for ecosystem tooling:

- `<robot>` root with version 7.0 attribute
- `<suite>` per scenario
- `<test>` per step (Given/When/Then)
- `<kw>` for keyword invocations under each step
- `<msg>` for log lines and OCPP wire events
- `<status status="PASS|FAIL|SKIP" starttime="..." endtime="..."/>`
- `<statistics>` summary

We do not emit:

- `<arguments>` for keyword arguments (OCTANE renders them via `<msg>`)
- `<for>`, `<while>`, `<if>` control flow elements (no DSL equivalent)
- Robot listener artifacts (we are not a Robot runner)

### Validation

A schema validation step in CI uses `xmllint` against the published
Robot Framework XSD (`schema/robot.xsd` in the Robot repo, vendored
at a pinned commit under `test/robotxml/schema.xsd`). PRs that produce
non-validating XML fail CI.

### Versioning

The Robot Framework output schema is versioned. OCTANE pins to
**Robot Framework 7.x output format** for v1. A future Robot version
that breaks compatibility triggers an ADR amendment, not a silent
upgrade.

### Testing

- A golden test under `test/robotxml/golden_test.go` runs OCTANE
  against a fixed scenario and asserts the produced XML matches a
  pinned golden file.
- A consumer test pipes the output through Robot Framework's own
  `rebot` tool (in a CI sidecar container, not as a Go dependency) to
  verify it is consumable upstream.

## Consequences

### Positive

- OCTANE plugs into Allure, ReportPortal, Jenkins, GitLab test
  reporters, and IDE Robot plugins on day one.
- Adoption: a CSMS team already running Robot Framework for other QA
  can fold OCTANE results into the same dashboard with no extra work.
- The strategic value is asymmetric: a 200-line emitter buys a
  multi-thousand-tool ecosystem.

### Negative

- We carry a schema we do not own. Robot Framework's release cadence
  may force OCTANE to track. Mitigated by pinning to a specific Robot
  major version and by golden tests that detect regressions.
- The format is verbose for large runs. Mitigated by `--no-robot-xml`
  for users who only consume `report.json`.

### Neutral

- The native `report.json` remains the source of truth for
  certification and for OCTANE's determinism guarantees. The XML is a
  view, not a contract.

## Alternatives considered

- **JUnit XML.** Considered. Less expressive (no nested keywords, no
  domain-specific message types) but simpler. Decision: emit Robot's
  format because OCTANE is structurally closer to Robot than to JUnit;
  JUnit emission is a future ADR if demand arises.
- **TAP or TestAnything Protocol.** Niche; ecosystem support narrower
  than Robot.
- **No interop format.** Rejected: leaves significant reporting value
  on the table for a small implementation cost.

## References

- Constitution: principle II
- ADR 0005 (story framework)
- Robot Framework output schema:
  https://github.com/robotframework/robotframework/tree/master/doc/schema
- pytest JUnit XML emission as prior art
