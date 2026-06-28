# Spec 003: Keyword API and Registry

> **Spec ID:** `003-keyword-api`
> **Status:** Approved
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCTANE's story DSL is declarative; every step in a `.story` file
is an English-shaped sentence that must resolve to a Go function
that drives the wire. The contract between story authors and
keyword authors — the *keyword API* — is the most architecturally
load-bearing surface in the project.

This spec defines:

- The Go interfaces that every keyword function honors (`Func`,
  `Args`, `State`, `Station`).
- The registration mechanism by which keyword libraries make
  themselves discoverable to the runtime (`registry.Register`).
- The pattern grammar for matching step text against keyword
  patterns (the `{name:type}` placeholder syntax from ADR 0007).
- The resolver that walks an AST step and selects the correct
  keyword from the registered set.

This spec ships **only the contracts and the registration/resolver
machinery**. It ships **zero** keyword bodies. Spec 004 ships the
small set of primitive (transport-level) keywords. Domain
keywords (Authorize, BootNotification, etc.) are deferred to spec
007 and beyond.

The split exists so that the API surface can be reviewed,
frozen, and depended on by downstream specs without waiting for
keyword bodies to land.

## 2. Goals

- G1. Implement `pkg/keywords/api/` — the public interfaces and
      types every keyword consumes (`Func`, `Args`, `State`,
      `Station`, `Layer`, `OCPPVersion`, `Keyword`).
- G2. Implement `pkg/keywords/registry/` — the global registry
      with `Register`, `All`, `Resolve`. Self-registration via
      `init()` per ADR 0007.
- G3. Implement the `{name:type}` pattern matcher with support
      for types `string`, `int`, `float`, `bool`, `duration`,
      `station`, `any`.
- G4. Implement layered resolution per ADR 0007: domain layer
      wins over primitive layer for the same pattern within the
      same OCPP version scope.
- G5. Reject pattern collisions at registration time with a
      panic that names both keywords involved.
- G6. Provide mock-friendly `State` and `Station` interfaces so
      keyword libraries can be unit-tested without importing
      `pkg/runner/`.

## 3. Non-Goals

- N1. Keyword *bodies* (spec 004 + later).
- N2. The runtime that *invokes* keywords (spec 005).
- N3. Wire I/O implementation (spec 002).
- N4. Story parsing (spec 001).
- N5. CLI command `octane keywords list` — the resolver exposes
      the data; the command lives in spec 006.

## 4. User Stories

- **As a keyword author**, I want to write
  `func myKeyword(ctx, state, args) error` and register it with
  `registry.Register(api.Keyword{...})` without thinking about
  the runtime, the parser, or the wire layer.
- **As a downstream spec author** (spec 004, 005, 006, 007), I
  want a stable contract surface I can build against without
  waiting for keyword bodies to land.
- **As a story author**, I want unambiguous error messages when
  a step text does not match any registered keyword, with a
  Levenshtein-suggested closest match if one exists.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| IV. Determinism | `registry.All()` returns results sorted by `(Layer, OCPPVersion, Pattern)`. The resolver is deterministic given identical input. |
| V. Stdlib-Heavy | API and registry use only Go stdlib. No reflection-based magic; the contract is plain Go interfaces. |
| VI. Test Cases as Code | This is the typed-Go layer behind the declarative DSL. |
| XII. No CSMS Adaptation Surface | The registry has exactly two layers: primitive and domain. There is no per-CSMS override layer. |

## 6. Acceptance Criteria

- AC1. **Given** a keyword registered with
       `registry.Register(api.Keyword{...})`, **when**
       `registry.All()` is called, **then** the entry appears in
       the returned slice in `(Layer, OCPPVersion, Pattern)` sort
       order.
- AC2. **Given** two keyword registrations with the same
       `(Layer, OCPPVersion, Pattern)` tuple, **when** the second
       `Register` call executes, **then** the program panics with
       a message naming both registration sites.
- AC3. **Given** an AST step `the CSMS sends ReserveNow with
       connectorId 1 and idTag "X" to station "CP01" within 30
       seconds`, **when** the resolver runs against a registry
       containing the matching pattern, **then** it returns the
       keyword's `Func` and an `Args` value bound with
       `connectorId=1`, `idTag="X"`, `station="CP01"`,
       `timeout=30s`.
- AC4. **Given** an AST step that matches no registered pattern,
       **when** the resolver runs, **then** it returns
       `ErrNoMatch` carrying the closest pattern by Levenshtein
       distance (if within edit-distance 5) and its location.
- AC5. **Given** a keyword pattern declares `{n:int}` and the
       step text supplies a non-integer token, **when** the
       resolver runs, **then** it returns `ErrTypeMismatch` with
       the parameter name, expected type, and got-value.
- AC6. **Given** a domain-layer keyword for OCPP 1.6 and a
       primitive-layer keyword with the same pattern, **when**
       the resolver runs against a story declaring OCPP 1.6,
       **then** the domain keyword wins.
- AC7. **Given** a domain-layer keyword for OCPP 1.6 only and
       a story declaring OCPP 1.6, **when** the resolver runs,
       **then** the OCPP 1.6 keyword is invisible and the
       resolver falls through to the primitive layer.
- AC8. **Given** a mock `State` and `Station` from
       `pkg/keywords/api/mock`, **when** a third-party keyword
       is unit-tested against them, **then** the test does not
       require importing `pkg/runner/`, `pkg/transport/`, or any
       network library.

## 7. OCPP Scope

The keyword API is OCPP-version-agnostic. The `OCPPVersion`
enum (`1.6`, `1.6`, `2.1`) exists only as a registry filter;
no version-specific message logic appears in this package.

## 8. Open Questions

- OQ1. Whether `Args` should panic or return `(value, error)` on
       missing arguments. Recommendation: panic. The resolver
       validates pattern-vs-keyword at registration time, so a
       runtime missing argument is by definition a registry bug.
       Defensive `(value, error)` would push noise into every
       keyword body without catching a real failure mode.
       *(owner: Architect, due: with this spec — RESOLVED in favor
       of panic; documented in -10.)*
- OQ2. Whether to support optional parameters in patterns
       (e.g., `[with idTag {idTag:string}]`). Recommendation:
       no, in this spec. If the need arises, two patterns can
       point to the same `Func`. Optional parameters add real
       complexity to the matcher.
       *(owner: Architect, due: with this spec — RESOLVED.)*

## 9. Out of Scope (parking lot)

- Optional parameters in patterns.
- Regex-style wildcards in patterns.
- Keyword aliasing (one pattern → another pattern).
- Plugin loading from `.so` files (constitution principle XII
  forbids the per-CSMS override layer that this would imply).

## 10. Implementation notes

### Args panic semantics

`args.String(name)`, `args.Int(name)`, `args.Duration(name)`
panic if the named argument is missing or has the wrong type.
This is intentional. The registry validates that every
`{name:type}` placeholder in a pattern is consumed by exactly
one `Args.<Type>(name)` call in the keyword body — the check
runs at `init()` time. Reaching the runtime panic indicates a
registry bug, not an authoring bug.

The registry-time check is implemented by parsing the pattern
into placeholder tokens and walking the keyword's source for
matching `Args` access calls. (Static analysis, not reflection.)

### State interface contract

```go
type State interface {
    Station(handle string) (Station, error)
    Now() time.Time
    Logf(format string, args ...any)
    StashPendingCallId(stationHandle, messageId string)
    PopPendingCallId(stationHandle string) (string, bool)
}
```

The `StashPendingCallId` / `PopPendingCallId` pair supports the
two-keyword authoring pattern (request expectation + response
emission) where the response keyword needs the inbound CALL's
`messageId` to echo correctly. The runtime carries a small
per-station scratch space; this is the keyword-side surface to
it.

### Mock-friendliness contract

`State` is an interface, not a concrete type, so a mock state
can satisfy it without pulling in the runtime. The `pkg/keywords/api/mock`
package provides `NewMockState()` and `NewMockStation()` for
keyword unit tests.

A keyword that imports `pkg/runner/`, `pkg/transport/`, or
`net/http` directly is a code smell flagged by the reviewer
agent.

### Determinism rule

Keywords MUST use `state.Now()` instead of `time.Now()`. The
runtime injects a deterministic clock (per spec 002) so reports
are byte-identical across runs. The linter rejects `time.Now()`
calls in `pkg/keywords/`.

### Resolver inspection commands

The resolver's data is exposed via `registry.All()` and
`resolver.Resolve(astStep, ocppVersion)`. The CLI commands
`octane keywords list` and `octane keywords resolve --story foo`
that consume these are defined in spec 006.

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] Maintainer review
