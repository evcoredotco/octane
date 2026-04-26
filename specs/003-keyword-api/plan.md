# Plan 003: Keyword API and Registry

> **Spec ID:** `003-keyword-api`
> **Status:** Approved
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement the Go interfaces and registration mechanism that
keyword libraries consume. This spec ships zero keyword bodies;
it ships only the contract surface (api package), the global
registry, the pattern matcher, the layered resolver, and the
mock-friendly test doubles.

The deliverable is a stable contract that specs 004, 005, 006,
and 007 can build against without waiting on keyword bodies.

## 2. Architecture Touchpoints

- `pkg/keywords/api/` — new; public types and interfaces
- `pkg/keywords/api/mock/` — new; mock State and Station for keyword unit tests
- `pkg/keywords/registry/` — new; global registry and resolver
- `pkg/keywords/registry/internal/pattern/` — new; pattern matcher (placeholder syntax)
- Read-only consumers: `pkg/story/ast/` (the AST steps the
  resolver matches against), `pkg/transport.Station` and
  `pkg/engine/clock.Clock` (the runtime services exposed
  through `State`)

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `pkg/keywords/api.Layer` | new enum | initial |
| `pkg/keywords/api.OCPPVersion` | new enum | initial |
| `pkg/keywords/api.Args` | new struct | initial |
| `pkg/keywords/api.State` interface | new | initial |
| `pkg/keywords/api.Station` interface | new | initial |
| `pkg/keywords/api.Func` type alias | new | initial |
| `pkg/keywords/api.Keyword` struct | new | initial |
| `pkg/keywords/api/mock.NewMockState() State` | new | initial |
| `pkg/keywords/api/mock.NewMockStation() Station` | new | initial |
| `pkg/keywords/registry.Register(api.Keyword)` | new | initial |
| `pkg/keywords/registry.All() []Entry` | new | initial |
| `pkg/keywords/registry.Resolve(step ast.Step, ocpp OCPPVersion) (Match, error)` | new | initial |
| `pkg/keywords/registry.ErrNoMatch`, `ErrTypeMismatch` | new typed errors | initial |

## 4. Data Contracts

### Pattern grammar

A keyword pattern is a string with `{name:type}` placeholders.
Types are `string`, `int`, `float`, `bool`, `duration`,
`station`, `any`. Whitespace inside the pattern matches one or
more whitespace characters in the step; literal text matches
case-insensitively.

### State interface

The interface defined in spec 003 §10. Unmodified.

### Resolution rules

1. Filter registered keywords by `(Layer, OCPPVersion)`:
   domain-layer keywords matching the active OCPP version, plus
   all primitive-layer keywords.
2. Within the filtered set, apply patterns in
   `(Layer descending, Pattern length descending)` order. Domain
   wins over primitive; longer patterns win over shorter ones for
   ambiguous cases.
3. The first match returns; remaining patterns are not consulted.
4. No match returns `ErrNoMatch` with a Levenshtein-suggested
   closest pattern (within edit distance 5).

## 5. Required ADRs

- [x] ADR 0007 — Keyword library layering (now includes the
      Keyword author surface section that this spec implements)

No new ADRs needed.

## 6. Test Strategy

- **Unit tests**: pattern matcher correctness, type coercion
  (string → int, string → duration via `time.ParseDuration`),
  resolver layer precedence, error paths.
- **Mock-friendliness test**: a third-party keyword in
  `testdata/external_keyword/` registers itself, runs against
  `mock.NewMockState()`, and asserts the call ordering — without
  importing `pkg/runner/`. Validates spec 003 AC8.
- **Determinism test**: register 50 keywords in random order;
  assert `registry.All()` always returns them in the documented
  sort order.
- **Collision test**: register two keywords with the same
  `(Layer, OCPPVersion, Pattern)` tuple; assert the second
  `Register` panics with both registration sites named.

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** N/A.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `Args` panic semantics surprise authors | Medium | Medium | Documented in CONTRIBUTING.md and the api package godoc; reviewer agent flags missing pattern-vs-Args correspondence |
| Pattern matcher regex performance | Low | Low | Patterns are tiny; benchmark only if a keyword set exceeds 1000 entries |
| Levenshtein hint is wrong/misleading | Low | Low | Test against fixtures; cap suggestion at edit distance 5 |
| State interface grows a god-object surface | Medium | High | Reviewer agent enforces "small interfaces" per principle V; new methods require an ADR amendment |

## 9. Effort Estimate

- T-shirt size: **S**
- Calendar estimate: 1 week of focused work
- Parallelizable streams: api package + registry + pattern
  matcher are independent; mock package can land first

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] Maintainer review
