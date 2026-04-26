# Tasks 003: Keyword API and Registry

> **Spec ID:** `003-keyword-api`
> **Plan reference:** `./plan.md`
> **Status:** Ready

## Conventions

Same as previous specs.

---

## Phase 1 — Contracts

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-01 | Define `api.Layer` enum and `api.OCPPVersion` enum | architect | S | AC6, AC7 | `pkg/keywords/api/api.go` |
| T-003-02 | Define `api.Args`, typed accessors, panic semantics | architect | S | AC5 | `pkg/keywords/api/args.go` |
| T-003-03 | Define `api.State`, `api.Station` interfaces | architect | S | AC8 | `pkg/keywords/api/api.go` |
| T-003-04 | Define `api.Func`, `api.Keyword` types | architect | S | AC1 | `pkg/keywords/api/api.go` |
| T-003-05 | Define typed errors `ErrNoMatch`, `ErrTypeMismatch` | architect | P | AC4, AC5 | `pkg/keywords/registry/errors.go` |

## Phase 2 — Pattern matcher

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-10 | Pattern parser: `{name:type}` token extraction | backend | S | AC3, AC5 | `pkg/keywords/registry/internal/pattern/parse.go` |
| T-003-11 | Pattern matcher: case-insensitive literal + typed bind | backend | S | AC3 | `pkg/keywords/registry/internal/pattern/match.go` |
| T-003-12 | Type coercion: string → int/float/bool/duration | backend | S | AC5 | `pkg/keywords/registry/internal/pattern/coerce.go` |
| T-003-13 | Pattern matcher unit tests | qa | S | AC3, AC5 | `pkg/keywords/registry/internal/pattern/match_test.go` |

## Phase 3 — Registry

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-20 | Implement `Register` with collision panic | backend | S | AC2 | `pkg/keywords/registry/registry.go` |
| T-003-21 | Implement `All` with stable sort by `(Layer, OCPPVersion, Pattern)` | backend | S | AC1 | `pkg/keywords/registry/registry.go` |
| T-003-22 | Determinism test: 50 random-order registrations → stable sort | qa | P | AC1 | `pkg/keywords/registry/sort_test.go` |
| T-003-23 | Collision test: duplicate panics with both sites named | qa | P | AC2 | `pkg/keywords/registry/collision_test.go` |

## Phase 4 — Resolver

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-30 | Implement `Resolve(step, ocppVersion) (Match, error)` | backend | S | AC3, AC4, AC6, AC7 | `pkg/keywords/registry/resolve.go` |
| T-003-31 | Implement Levenshtein-suggestion helper for `ErrNoMatch` | backend | P | AC4 | `pkg/keywords/registry/internal/levenshtein/levenshtein.go` |
| T-003-32 | Layered resolution: domain wins over primitive | backend | S | AC6, AC7 | `pkg/keywords/registry/resolve.go` |
| T-003-33 | Resolver unit tests covering every resolution path | qa | S | AC3, AC4, AC6, AC7 | `pkg/keywords/registry/resolve_test.go` |

## Phase 5 — Mock package

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-40 | Implement `mock.NewMockState()`, `NewMockStation()` | backend | P | AC8 | `pkg/keywords/api/mock/mock.go` |
| T-003-41 | Test that mock package has zero imports of `pkg/runner` or `pkg/transport` | qa | S | AC8 | `pkg/keywords/api/mock/imports_test.go` |
| T-003-42 | Sample external keyword in `testdata/external/` exercising mocks | qa | S | AC8 | `pkg/keywords/api/mock/testdata/external/keyword.go` |

## Phase 6 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-50 | Godoc on every exported symbol; package-level overview | docs | P | — | `pkg/keywords/api/*.go`, `pkg/keywords/registry/*.go` |
| T-003-51 | Update `CONTRIBUTING.md` with keyword-author tutorial | docs | P | — | `CONTRIBUTING.md` |
| T-003-52 | Update `docs/concepts/keywords.md` | docs | P | — | `docs/concepts/keywords.md` |
| T-003-53 | CHANGELOG entry | docs | S | — | `CHANGELOG.md` |

## Phase 7 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-003-60 | Security review: panic semantics, no global mutable state outside registry | security | S | — | — |
| T-003-61 | Reviewer pass: API surface stability, error message clarity | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 8 acceptance criteria covered by at least one task
- [ ] Mock package has zero `pkg/runner` or `pkg/transport` imports
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/003-keyword-api` passes
- [ ] CHANGELOG updated under `## [Unreleased]`
