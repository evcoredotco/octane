# ADR 0014: Intellectual Property and Authoring Guidelines

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect
- **Constitution principles touched:** I (Conformance Above
  Convenience)

## Context

OCTANE asserts conformance to OCPP, a specification published by the
Open Charge Alliance (OCA). The OCA also maintains commercial test
catalogs and certification tooling whose content is independently
copyrighted. Distinguishing the two is essential for OCTANE's
intellectual-property posture:

- **OCPP specifications** describe the public protocol every
  conformant CSMS must implement. The behavior they define is
  factual and not protectable as expression. OCTANE may freely
  test against published specification behavior.
- **Third-party test catalogs and certification tooling** describe
  *test cases* (titles, descriptions, scenario numbering, validation
  steps) which are copyrighted creative works. Reproducing this
  content — verbatim or by close paraphrase — is prohibited.

To keep OCTANE legally clean and unaffiliated with any third-party
testing tool, the project adopts strict authoring rules.

## Decision

### Authoring derives from public specifications, not from third-party catalogs

Every OCTANE conformance story MUST derive its assertions from the
published OCPP specification. The author works from the protocol
document — message schemas, state machines, error codes, sequence
diagrams — and writes original prose describing the intended
behavior. The author does not consult, transcribe, paraphrase, or
restructure third-party test catalogs.

In practice this means:

- The story's `Spec-Ref` Meta key cites the OCPP specification by
  version, section, and message name (e.g. `OCPP-J 1.6 §6.40
  ReserveNow`). It does not cite any third-party scenario catalog.
- The story's `Name` is original prose written by the OCTANE
  contributor.
- The story's narrative `# header comments` are original prose.
- Test parameters and expected outcomes are derived from the
  specification's normative text, not from any external catalog's
  worked example.

### Naming convention: `resource_function_desire`

Story IDs and filenames follow a structured three-slot schema:

```text
<resource>_<function>_<desire>
```

| Slot       | Meaning                                          | Examples                                                                                                   |
|------------|--------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| `resource` | The OCPP entity under test                       | `connector`, `station`, `transaction`, `reservation`, `boot`, `heartbeat`, `authorize`                     |
| `function` | What the resource does or what is being asserted | `notification`, `start`, `stop`, `reservation`, `authorize`, `status`, `update`                            |
| `desire`   | The expected outcome or scenario flavor          | `accepted`, `rejected`, `faulted`, `available`, `concurrent`, `malformed`, `timeout`, `success`, `failure` |

All lowercase, snake_case, underscores between slots. Multi-word
slot values use no separators (`idtoken`, not `id_token`).

Examples:

- `boot_notification_accepted`
- `boot_notification_malformed`
- `connector_reservation_faulted`
- `authorize_concurrent_rejected`
- `connector_status_available`
- `station_boot_accepted`

The `desire` slot prefers a specific protocol-level state when one
applies (`faulted`, `concurrenttx`, `accepted`) over a generic
outcome category (`success`, `failure`).

### No references to third-party tools in published artefacts

OCTANE's published outputs — the binary, story files, ADRs,
documentation, web site, man pages, CLI help text, error messages —
contain no references to third-party CSMS testing tools by name.
Where the project needs to refer to "an external authority that
operates formal certification," it does so generically without
naming the tool or organization.

This applies to:

- Source comments and docstrings
- ADR cross-references
- README / website / man-page prose
- Story Meta blocks and inline comments
- CLI output, error messages, log lines
- Generated reports (JSON, Robot XML)

Internal-only project artefacts (private design notes, slack
discussions, this very ADR's *Context* section by way of historical
record) MAY mention third-party tooling in passing but should
prefer generic descriptions where possible.

### Forbidden authoring patterns

The following are explicitly prohibited:

1. **Verbatim reproduction.** Copying any text from a third-party
   test catalog into an OCTANE story, ADR, or other artefact.
2. **Close paraphrase.** Restating a third-party catalog's
   description while preserving its distinctive structure or
   wording.
3. **Numerical translation.** Mapping third-party test case IDs
   one-to-one onto OCTANE story IDs (e.g. "TC_048_1" → "tc_048_1").
   OCTANE IDs follow the `resource_function_desire` schema and are
   independently chosen.
4. **Catalog-derived structure.** Reproducing a third-party
   catalog's section ordering, scenario grouping, or table layout
   in OCTANE documentation.

### Permitted patterns

The following are explicitly permitted:

1. **Citing the OCPP specification by section.** `Spec-Ref:
   OCPP-J 1.6 §6.40 ReserveNow` is a factual reference to public
   protocol documentation.
2. **Independently authored test descriptions.** Two parties
   describing the same protocol behavior will arrive at similar
   prose; this similarity does not constitute infringement when
   each author derived the text from the specification.
3. **Functional overlap.** OCTANE stories will exercise the same
   protocol behaviors that any conformance test catalog also
   exercises, because both derive from the same specification.
   Functional overlap is not infringement; expression-level
   reproduction is.

## Consequences

### Positive

- OCTANE's IP posture is defensible. The project asserts
  conformance to the OCPP specifications without claiming
  affiliation with any third party and without reproducing any
  third-party copyrighted content.
- The story library has its own coherent identity, with names that
  describe what each test does rather than encoding an external
  catalog's numbering.
- Contributors have clear, mechanical guidance for what is and is
  not acceptable.

### Negative

- Authors familiar with third-party catalogs must consciously work
  from the OCPP specification rather than from the catalog. This is
  more effort but produces cleaner work.
- The OCTANE story library cannot advertise "X% coverage of
  catalog Y" without legal review of how that claim is phrased.

## References

- Constitution principle I (Conformance Above Convenience)
- ADR 0006 (story DSL grammar) — defines `Spec-Ref` Meta key
- `CONTRIBUTING.md` — operational guidance for authors
- `docs/conformance-claim.md` — public positioning of OCTANE's
  conformance claim
