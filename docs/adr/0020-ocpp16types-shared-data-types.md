# ADR 0020: Shared OCPP 1.6 Data Types via `ocpp16types`

- **Status:** Accepted
- **Date:** 2026-04-27
- **Deciders:** Project maintainer, Architect, Backend, Keyword Author
- **Constitution principles touched:** I (Conformance Above Convenience),
  V (Go-First, Stdlib-Heavy), VI (Test Cases Are Code, Not Configuration),
  XII (Scenarios Are Declarative; Adaptation Lives in Profiles)

---

## Context

OCTANE models every OCPP 1.6 message, field, and enumeration as typed Go
values. Without a canonical home for those types, each EVCore application
that speaks OCPP 1.6 would independently redeclare the same structs —
leading to subtle divergences (wrong field tags, missing validations,
silently incorrect enum values) that compound over time.

The EVCore organisation maintains a dedicated, versioned Go module:

```
github.com/evcoreco/ocpp16types
```

This module is the **single source of truth** for all OCPP 1.6 data types
across every EVCore application. It was designed specifically so that all
EVCore software that deals with OCPP 1.6 shares identical struct layouts,
JSON tags, enum constants, and validation logic — eliminating the class of
divergence bugs described above.

Constitution principle V (Stdlib-Heavy) requires an ADR for every new
dependency. This ADR is that record. The dependency is classified as
**first-party EVCore infrastructure**, not a third-party library: it is
owned and maintained by the same organisation that owns OCTANE and carries
the same quality and stability bar.

---

## Decision

**All OCPP 1.6 data types used anywhere in OCTANE MUST come from
`github.com/evcoreco/ocpp16types`.**

This rule is absolute. There are no exceptions.

Concretely:

- Any Go struct that represents an OCPP 1.6 message (request, response,
  notification), field type, or enumeration **must** be imported from
  `github.com/evcoreco/ocpp16types`, not declared locally.
- Local re-declarations, type aliases, and shadow copies of OCPP 1.6
  types are **forbidden**, even for test-only packages.
- When `github.com/evcoreco/ocpp16types` does not yet expose a type that
  OCTANE needs, the correct action is to **contribute that type upstream**
  to the shared module, not to declare it locally as a stopgap.
- Code review (the `reviewer` agent) must reject any PR that introduces a
  locally declared OCPP 1.6 data type.

### What counts as an OCPP 1.6 data type

Anything whose shape, field names, or allowed values are specified in the
OCPP 1.6 (or 1.6J) specification:

- Request and response structs (e.g. `BootNotificationRequest`,
  `BootNotificationResponse`)
- Notification structs (e.g. `MeterValuesRequest`)
- Enumerations defined by the spec (e.g. `RegistrationStatus`,
  `ChargePointStatus`, `AuthorizationStatus`)
- Embedded sub-objects defined by the spec (e.g. `IdTagInfo`,
  `MeterValue`, `SampledValue`)
- Field-level types with spec-mandated constraints (e.g. `CiString20Type`,
  `CiString50Type`)

OCTANE-internal orchestration types (e.g. `runner.RunResult`,
`report.Finding`) are **not** OCPP data types and are not subject to this
rule.

### Import path convention

```go
import "github.com/evcoreco/ocpp16types"

// Use the package alias "ocpp16" for readability.
import ocpp16 "github.com/evcoreco/ocpp16types"
```

Keyword functions reference types as `ocpp16.BootNotificationRequest`,
`ocpp16.RegistrationStatus`, etc.

### Contributing types upstream

When a needed OCPP 1.6 type is absent from `ocpp16types`:

1. Open a PR against `github.com/evcoreco/ocpp16types` with the new type.
2. Reference that upstream PR in the OCTANE task that needs it.
3. Do **not** merge the OCTANE task until the upstream release is tagged
   and the `go.mod` pin updated.
4. Record the dependency in the task's `plan.md` entry as a blocker.

---

## Consequences

### Positive

- OCPP 1.6 struct definitions exist in exactly one place across all EVCore
  applications. A spec correction (e.g. a field renamed in an errata) is
  applied once and propagated via a single version bump.
- Tests, keywords, and wire code all refer to the same Go type. A mismatch
  that would otherwise surface only at JSON serialisation time is now a
  compile error.
- OCTANE's dependency set remains minimal: `ocpp16types` is first-party and
  does not drag in transitive third-party libraries.
- Certification reviewers can audit OCPP type fidelity by inspecting one
  module rather than searching multiple repos.

### Negative

- OCTANE development blocks on upstream `ocpp16types` releases when a
  needed type is missing. This is the correct trade-off: it prevents the
  short-cut of declaring a local copy.
- Agents must resist the temptation to declare a local type as a stopgap
  when the upstream release is slow. The workflow above (upstream PR first,
  OCTANE task blocked until the release) must be followed.

### Neutral

- The `go.mod` entry for `ocpp16types` must be kept in the `require ()`
  block alongside the other direct dependencies. It is never an indirect
  dependency.

---

## Alternatives considered

- **Declare types locally in `pkg/scenarios/v16/`.** Rejected: this is
  exactly the divergence that `ocpp16types` was created to prevent.
- **Generate types from the OCPP schema.** Considered for a future version.
  Would still consume `ocpp16types` as the canonical output target; this ADR
  remains applicable.
- **Vendor the type definitions inline.** Rejected: creates a second source
  of truth and breaks the cross-EVCore consistency guarantee.

---

## References

- `github.com/evcoreco/ocpp16types` — the shared type module
- Constitution principle I (Conformance Above Convenience)
- Constitution principle V (Go-First, Stdlib-Heavy)
- ADR 0002 (Go engine language)
- ADR 0007 (Keyword library layering)
