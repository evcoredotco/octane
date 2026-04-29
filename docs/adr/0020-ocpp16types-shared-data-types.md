# ADR 0020: Shared OCPP 1.6 Infrastructure via `ocpp16types`, `ocpp16messages`, and `ocpp16j`

- **Status:** Accepted (amended 2026-04-29)
- **Date:** 2026-04-27
- **Amended:** 2026-04-29 — expanded to mandate `ocpp16messages` for all OCPP 1.6 message construction
- **Amended:** 2026-04-29 — expanded to mandate `ocpp16j` for all OCPP-J JSON framing, validation, and marshaling; completing the three-layer EVCore standard
- **Deciders:** Project maintainer, Architect, Backend, Keyword Author
- **Constitution principles touched:** I (Conformance Above Convenience),
  V (Go-First, Stdlib-Heavy), VI (Test Cases Are Code, Not Configuration),
  XII (Scenarios Are Declarative; Adaptation Lives in Profiles)

---

## Context

OCTANE models every OCPP 1.6 message, field, and enumeration as typed Go
values. Without canonical homes for those types, message constructors, and
JSON framing, each EVCore application that speaks OCPP 1.6 would independently
redeclare the same structs, rebuild the same serialisation/validation logic,
and re-implement the same OCPP-J wire protocol — leading to subtle divergences
(wrong field tags, missing validations, silently incorrect enum values,
malformed JSON arrays) that compound over time.

The EVCore organisation maintains three dedicated, versioned Go modules that
together form the **complete OCPP 1.6 infrastructure layer**:

```text
github.com/evcoreco/ocpp16types     — primitive types, enumerations, sub-objects
github.com/evcoreco/ocpp16messages  — complete request / response message constructors
github.com/evcoreco/ocpp16j         — OCPP-J JSON framing, parsing, validation, marshaling
```

These three modules are the **single source of truth** for all OCPP 1.6 data
structures, message construction, and JSON wire-format handling across every
EVCore application. They are designed so that all EVCore software shares
identical struct layouts, JSON tags, enum constants, constructor behaviour,
validation logic, and JSON envelope handling — eliminating the class of
divergence bugs described above.

### Layer responsibilities

```text
JSON bytes  ─→  ocpp16j                  ─→  ocpp16messages
               Parse / Validate / Marshal     Req() / Conf() constructors
               Call / CallResult / CallError  Typed, validated message bodies
               UniqueId / ErrorCode           Uses ocpp16types internally
               Registry + JSONDecoder
                    │
                    └─→  ocpp16types
                         Primitive types, enumerations, sub-objects
                         CiString, DateTime, IdTagInfo, ChargingProfile, …
```

- **`ocpp16j`** owns the JSON array envelope defined in OCPP-J 1.6 section 4.
  It does not know OCPP semantics; it delegates payload validation entirely to
  `ocpp16messages`.
- **`ocpp16messages`** owns the payload layer: typed request/response message
  bodies and the `Req()`/`Conf()` constructor pairs that validate field values.
- **`ocpp16types`** owns every field-level constrained type, enumeration, and
  composite sub-object referenced by the OCPP 1.6 specification.

`ocpp16messages` re-exports `ocpp16types`, so code that imports `ocpp16messages`
can reach all primitive types through its `types/` re-export. Direct use of
`ocpp16types` is still preferred when only field-level types or enumerations
are needed and no message construction is involved.

`ocpp16j` depends on `ocpp16messages` for its `JSONDecoder` integration and on
`ocpp16types` transitively. All three must be pinned as direct dependencies in
`go.mod`.

Constitution principle V (Stdlib-Heavy) requires an ADR for every new
dependency. This ADR covers all three modules. All three are classified as
**first-party EVCore infrastructure**, not third-party libraries: they are
owned and maintained by the same organisation that owns OCTANE, carry the
same quality and stability bar, and introduce no transitive external
dependencies.

---

## Decision

### Rule 1 — All OCPP 1.6 data types MUST come from `ocpp16types`

**All OCPP 1.6 data types used anywhere in OCTANE MUST come from
`github.com/evcoreco/ocpp16types`.**

This rule is absolute. There are no exceptions.

Concretely:

- Any Go struct that represents an OCPP 1.6 field type, enumeration, or
  composite sub-object **must** be imported from `github.com/evcoreco/ocpp16types`,
  not declared locally.
- Local re-declarations, type aliases, and shadow copies of OCPP 1.6
  types are **forbidden**, even for test-only packages.
- When `github.com/evcoreco/ocpp16types` does not yet expose a type that
  OCTANE needs, the correct action is to **contribute that type upstream**
  to the shared module, not to declare it locally as a stopgap.

### Rule 2 — All OCPP 1.6 message construction MUST use `ocpp16messages`

**All construction and use of OCPP 1.6 request and response messages anywhere
in OCTANE MUST go through `github.com/evcoreco/ocpp16messages`.**

This rule is absolute. There are no exceptions.

Concretely:

- Building an OCPP 1.6 request or confirmation message is done exclusively
  via the per-message constructor exported from the relevant sub-package
  (`Req(input ReqInput)` and `Conf(input ConfInput)`).
- Raw struct literals that construct request or response message bodies
  without using these constructors are **forbidden**.
- When `ocpp16messages` does not yet expose a message that OCTANE needs,
  the correct action is to **contribute that message upstream**, not to
  hand-roll a local struct.
- Code review (the `reviewer` agent) must reject any PR that introduces a
  locally constructed OCPP 1.6 message outside of these constructors.

### Rule 3 — All OCPP-J JSON framing MUST use `ocpp16j`

**All parsing, construction, validation, and marshaling of OCPP-J JSON
messages anywhere in OCTANE MUST go through `github.com/evcoreco/ocpp16j`.**

This rule is absolute. There are no exceptions.

Concretely:

- Parsing incoming OCPP-J bytes is done exclusively via `ocpp16json.Parse()`.
  Raw calls to `json.Unmarshal` against OCPP-J arrays are **forbidden**.
- Building a `Call` (request), `CallResult` (response), or `CallError` is done
  exclusively via `ocpp16json.NewCall()`, `ocpp16json.NewCallResult()`, and
  `ocpp16json.NewCallError()` respectively.
- Raw JSON array literals of the form `[2,"id","Action",{...}]` constructed
  by hand are **forbidden**.
- Marshaling a message to wire bytes is done exclusively via `json.Marshal`
  applied to a value produced by the constructors above — never by
  hand-assembling a JSON string or byte slice.
- Decoding a typed payload from a `Call` or `CallResult` is done via
  `ocpp16json.Registry.Decode()` using a `JSONDecoder[Input, Output]` that
  bridges the wire struct to an `ocpp16messages` constructor.
- `UniqueId` values MUST be created via `ocpp16json.NewUniqueId()` — raw
  string casts to the wire are **forbidden**.
- `ErrorCode` values MUST use the sentinel constants exported from `ocpp16j`
  (`ocpp16json.NotImplemented`, `ocpp16json.InternalError`, etc.).
- When `ocpp16j` does not yet expose a feature that OCTANE needs, the correct
  action is to **contribute it upstream**, not to implement it locally.
- Code review (the `reviewer` agent) must reject any PR that bypasses these
  constructors or parses OCPP-J bytes without `ocpp16json.Parse()`.

### What counts as an OCPP 1.6 data type (Rule 1 scope)

Anything whose shape, field names, or allowed values are specified in the
OCPP 1.6 (or 1.6J) specification at the field or sub-object level:

| Category                      | Examples                                                                                                                                                                                |
|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Field-level constrained types | `CiString20Type`, `CiString50Type`, `CiString255Type`, `CiString500Type`, `DateTime`, `Integer`                                                                                         |
| Enumerations                  | `AuthorizationStatus`, `ChargePointStatus`, `RegistrationStatus`, `DiagnosticsStatus`, `ChargingProfileKindType`, `ChargingRateUnit`, `Measurand`, `Phase`, `Location`, `UnitOfMeasure` |
| Composite sub-objects         | `IdTagInfo`, `IDToken`, `MeterValue`, `SampledValue`, `ChargingSchedule`, `ChargingSchedulePeriod`, `ChargingProfile`, `AuthorizationData`, `KeyValue`                                  |
| Error sentinels               | `ErrEmptyValue`, `ErrInvalidValue`                                                                                                                                                      |

### What counts as an OCPP 1.6 message (Rule 2 scope)

Any request or response frame whose structure is defined by the OCPP 1.6
specification, delivered as a constructor pair in `ocpp16messages`:

| Sub-package                            | Constructors  |
|----------------------------------------|---------------|
| `authorize`                            | `Req`, `Conf` |
| `bootnotification`                     | `Req`, `Conf` |
| `starttransaction`                     | `Req`, `Conf` |
| `stoptransaction`                      | `Req`, `Conf` |
| `metering`                             | `Req`, `Conf` |
| `setchargingprofile`                   | `Req`, `Conf` |
| _(and all other message sub-packages)_ | `Req`, `Conf` |

OCTANE-internal orchestration types (e.g. `runner.RunResult`,
`report.Finding`) are **not** OCPP data types or messages and are not subject
to Rules 1 or 2.

### What counts as OCPP-J JSON framing (Rule 3 scope)

Any value or operation defined in OCPP-J 1.6 section 4 (the JSON RPC layer):

| Concept                        | `ocpp16j` surface                                                             |
|--------------------------------|-------------------------------------------------------------------------------|
| Incoming bytes → typed message | `ocpp16json.Parse([]byte)`                                                    |
| Message type detection         | `ocpp16json.IsCall()`, `IsCallResult()`, `IsCallError()`                      |
| Concrete struct extraction     | `ocpp16json.AsCall()`, `AsCallResult()`, `AsCallError()`                      |
| Build a CALL (request)         | `ocpp16json.NewCall(uniqueId, action, payload)`                               |
| Build a CALLRESULT (response)  | `ocpp16json.NewCallResult(uniqueId, payload)`                                 |
| Build a CALLERROR (error)      | `ocpp16json.NewCallError(uniqueId, errorCode, desc, details)`                 |
| Wire bytes                     | `json.Marshal(call)` / `json.Marshal(callResult)` / `json.Marshal(callError)` |
| Payload decode pipeline        | `ocpp16json.Registry` + `ocpp16json.JSONDecoder[Input, Output](constructor)`  |
| Correlation identifier         | `ocpp16json.NewUniqueId(string)` → `ocpp16json.UniqueId`                      |
| Error codes                    | `ocpp16json.NotImplemented`, `ocpp16json.InternalError`, etc.                 |
| Application-layer wrapper      | `ocpp16json.NewDecodedCall[T]()`, `NewDecodedCallResult[T]()`                 |

### Import path conventions

```go
// For field types, enumerations, and sub-objects only:
import ocpp16 "github.com/evcoreco/ocpp16types"

// For message construction (Req / Conf):
import "github.com/evcoreco/ocpp16messages/authorize"
import "github.com/evcoreco/ocpp16messages/bootnotification"
// ... one import per message sub-package as needed

// For JSON framing, parsing, and marshaling:
import ocpp16json "github.com/evcoreco/ocpp16j"

// ocpp16messages re-exports ocpp16types; when a file uses both
// message constructors and field types, the types can be reached via
// the re-export. Direct import of ocpp16types is still preferred for
// files that only need field types or enumerations.
```

Canonical usage pattern — inbound message pipeline:

```go
// 1. Parse the wire bytes.
message, err := ocpp16json.Parse(rawBytes)
if err != nil { /* handle */ }

// 2. Detect type and extract concrete struct.
if ocpp16json.IsCall(message) {
    call, _ := ocpp16json.AsCall(message)

    // 3. Decode the payload through the Registry.
    //    JSONDecoder bridges ReqInput (wire struct) → Req() constructor.
    result, err := registry.Decode(call.Action, call.Payload)

    // 4. Type-assert to the validated domain type.
    req := result.(bootnotification.ReqMessage)
    _ = req
}
```

Canonical usage pattern — outbound message:

```go
// 1. Build the validated payload via ocpp16messages.
conf, err := bootnotification.Conf(bootnotification.ConfInput{
    CurrentTime:      ocpp16.DateTime(time.Now()),
    Interval:         300,
    RegistrationStatus: ocpp16.RegistrationStatusAccepted,
})

// 2. Wrap it in a CALLRESULT envelope.
uid, _ := ocpp16json.NewUniqueId(correlationId)
callResult, err := ocpp16json.NewCallResult(uid, conf)

// 3. Marshal to wire bytes.
wireBytes, err := json.Marshal(callResult)
// wireBytes: [3,"<correlationId>",{...}]
```

Keyword functions reference types as `ocpp16.AuthorizationStatus`,
constructors as `bootnotification.Req(bootnotification.ReqInput{...})`,
and JSON framing as `ocpp16json.NewCall(uid, "BootNotification", payload)`.

### Contributing types, messages, and framing features upstream

When a needed OCPP 1.6 type, message, or `ocpp16j` feature is absent from the
upstream module:

1. Open a PR against the relevant upstream repo (`ocpp16types`, `ocpp16messages`,
   or `ocpp16j`) with the new type, message, or feature.
2. Reference that upstream PR in the OCTANE task that needs it.
3. Do **not** merge the OCTANE task until the upstream release is tagged
   and the `go.mod` pin updated.
4. Record the dependency in the task's `plan.md` entry as a blocker.

---

## Consequences

### Positive

- OCPP 1.6 struct definitions, message construction logic, and JSON wire-format
  handling exist in exactly one place across all EVCore applications. A spec
  correction is applied once and propagated via a single version bump.
- Tests, keywords, and wire code all refer to the same Go types, the same
  constructor paths, and the same JSON framing layer. A mismatch that would
  otherwise surface only at JSON serialisation time is now a compile error.
- OCTANE's dependency set remains minimal: all three modules are first-party
  and introduce no transitive external dependencies.
- Certification reviewers can audit OCPP type fidelity, message conformance,
  and wire-format correctness by inspecting three modules rather than searching
  multiple repos.
- Constructor validation (`errors.Join()` over multiple field errors) surfaces
  invalid message construction at build time, not at runtime against a live CSMS.
- `ocpp16j`'s `Registry` + `JSONDecoder[Input, Output]` pattern provides a
  single, thread-safe decode pipeline that ties raw JSON bytes directly to
  validated `ocpp16messages` domain types with no intermediate raw-map access.

### Negative

- OCTANE development blocks on upstream releases when a needed type, message,
  or framing feature is missing. This is the correct trade-off: it prevents the
  short-cut of declaring a local copy or rolling a hand-crafted JSON array.
- Agents must resist the temptation to declare a local type, hand-roll a message
  struct, or build OCPP-J arrays by hand. The workflow above (upstream PR first,
  OCTANE task blocked until the release) must be followed.

### Neutral

- All three `go.mod` entries (`ocpp16types`, `ocpp16messages`, `ocpp16j`) must
  be kept in the `require ()` block as direct dependencies. None is ever an
  indirect dependency.
- `ocpp16j` depends on `ocpp16messages`, which depends on `ocpp16types`.
  OCTANE's `go.mod` pins all three explicitly to guarantee version alignment.
- `go mod tidy` will remove entries that are not yet imported by any Go source
  file. Entries are re-established automatically via `go get` when implementation
  tasks begin importing these packages.

---

## Alternatives considered

- **Declare types, message structs, and JSON framing locally in `pkg/`.** Rejected:
  this is exactly the divergence that these modules were created to prevent.
- **Use only `ocpp16messages` and never import `ocpp16types` directly.**
  Rejected: when a file only needs enumerations or field types, the direct
  import is cleaner and avoids pulling in the full message sub-package graph.
- **Use only `ocpp16j` for everything and never import `ocpp16messages` directly.**
  Rejected: `ocpp16j` delegates payload validation to `ocpp16messages`; keyword
  and test code needs the typed constructors directly, not just through the
  decode pipeline.
- **Generate types from the OCPP schema.** Considered for a future version.
  Would still target `ocpp16types` as the canonical output module; this ADR
  remains applicable.
- **Vendor the type definitions inline.** Rejected: creates a second source
  of truth and breaks the cross-EVCore consistency guarantee.
- **Hand-assemble OCPP-J arrays with `json.Marshal` on maps or raw strings.**
  Rejected: bypasses envelope validation (UniqueId length, MessageTypeId range,
  ErrorCode whitelist) and couples OCTANE to the spec's byte-level details
  rather than its typed abstractions.

---

## References

- `github.com/evcoreco/ocpp16types` — primitive types, enumerations, sub-objects
- `github.com/evcoreco/ocpp16messages` — request / response message constructors
- `github.com/evcoreco/ocpp16j` — OCPP-J JSON framing, parsing, validation, marshaling
- OCPP-J 1.6 specification, section 4 (JSON RPC message structures)
- Constitution principle I (Conformance Above Convenience)
- Constitution principle V (Go-First, Stdlib-Heavy)
- ADR 0002 (Go engine language)
- ADR 0007 (Keyword library layering)
