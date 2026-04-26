# ADR 0005: Story-Driven Conformance Framework

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect
- **Constitution principles touched:** I (Conformance Above Convenience),
  II (Two Distribution Surfaces), VI (Test Cases as Code), XI (Wire
  conformance), XII (No CSMS-specific adaptation)

## Context

OCTANE must verify OCPP conformance against an unmodified CSMS.
This rules out designs that require code or services in the CSMS
itself: a conformance tool that demands vendor cooperation is, in
practice, not a conformance tool. The replacement design must:

- Run against an unmodified CSMS.
- Verify conformance from observable wire behavior.
- Be honest about what it cannot verify.
- Borrow proven patterns from existing test frameworks rather than
  invent new ones.

Robot Framework provides the structural metaphor: declarative
scenarios in a DSL, executed by a layered keyword library, producing
machine-readable reports. We take the metaphor without taking the
runtime — Go gives us distribution and determinism that a Python
runtime cannot.

## Decision

Adopt a **story-driven conformance framework** modeled on the
architecture of Robot Framework but implemented natively in Go.
OCTANE becomes a runner over three layered concepts:

| Layer | Artifact | Owner |
|-------|----------|-------|
| Specification | `.story` files (Gherkin-flavored DSL, ADR 0006) | OCTANE project |
| Execution | Two-layer keyword library: primitive + domain (ADR 0007) | OCTANE project |
| Connection | YAML metadata describing how to reach the CSMS (ADR 0010) | Operator (user) |

The runtime contract is:

```
octane run \
  --connection citrineos \
  --story scenarios/v201/TC_B_01_CS.story
```

OCTANE loads the connection metadata, parses the story per ADR 0006,
resolves every step against the keyword library per ADR 0007, opens
WebSocket connections to the CSMS as one or more impersonated
charging stations (ADR 0008 covers multi-station orchestration), and
emits a deterministic JSON report plus an optional Robot Framework
`output.xml` companion (ADR 0009) for ecosystem integration.

### What OCTANE does not do

- It does not require any change to the CSMS.
- It does not host or require a sidecar service.
- It does not assume CSMS admin API access.
- It does not need a vendor-maintained adapter to run.
- It does not adapt domain keywords to specific CSMSes (ADR 0007,
  constitution principle XII).

### What OCTANE explicitly cannot verify

The honest scope is **wire conformance to OCPP**. Scenarios that
require privileged CSMS observability (audit log content, billing
pipeline behavior, internal state transitions invisible on the
wire) are either out of scope or marked **operator-assisted** —
runnable interactively, skipped automatically in CI mode.

This bounded claim is documented prominently in the published
report:

```
Wire conformance:        87/120 scenarios (PASS: 84, FAIL: 3)
Operator-assisted:       12 scenarios skipped (CI mode)
Overall conformance:     wire-tier
```

## Consequences

### Positive

- **Zero adoption cost.** Any CSMS team can run OCTANE against an
  unmodified instance with one CLI command.
- **Defensible conformance claim.** "Wire conformance to OCPP X.Y.Z"
  is a precise, verifiable claim that does not depend on
  vendor-written glue code.
- **Familiar mental model.** Robot Framework's structure is well
  known in the QA community; story files read as executable
  specifications.
- **Multi-station orchestration is natural.** Many stateful
  scenarios (transactions, multi-station auth, reservation
  lifecycle) become reachable through coordinated wire interaction
  (ADR 0008).
- **Constitutional principle VI is preserved.** Story files are a
  declarative *surface*; the keyword library that executes them is
  typed Go code.

### Negative

- **Some scenarios become uncoverable.** Audit log assertions and
  internal observability scenarios cannot be verified on the wire.
  Documented as operator-assisted, accepted as the cost of zero-code
  adoption.
- **DSL design and parsing add scope.** A Gherkin-flavored parser is
  non-trivial. Mitigated by pinning the grammar in ADR 0006 and
  shipping a small recursive-descent parser.

### Neutral

- The constitution carries principles XI (wire-only conformance) and
  XII (no CSMS-specific adaptation). Existing engineering principles
  are unchanged.

## Alternatives considered

- **Vendor-implemented test harness adapters.** Considered early in
  the project. Rejected: vendor cooperation cost is unacceptable
  and conflates "the CSMS is conformant" with "the adapter is
  correct."
- **Reuse Robot Framework directly.** Rejected on distribution
  grounds (Python runtime in CI vs single static Go binary) and on
  parser/runtime drift risk. Robot's `output.xml` format is reused
  for reporting (ADR 0009); the runtime is native.
- **Pure Go scenarios (no DSL).** Rejected because the primary
  readers of conformance scenarios are not Go engineers — they are
  certification reviewers, OCA participants, and CSMS QA leads.
- **YAML scenarios.** Rejected because Given/When/Then carries
  semantic weight that YAML structures awkwardly.

## References

- Constitution: principles I, II, VI, XI, XII
- ADR 0006 (story DSL grammar)
- ADR 0007 (keyword library layering)
- ADR 0008 (multi-station orchestration)
- ADR 0009 (Robot Framework output.xml compatibility)
- ADR 0010 (connection profiles)
- Robot Framework: https://robotframework.org/
