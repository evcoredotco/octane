# Plan 004: Primitive Keywords

> **Spec ID:** `004-primitive-keywords`
> **Status:** Approved
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement the small set of transport-level primitive keywords
listed in spec 004 -10. Each keyword is a function plus a
`registry.Register(api.Keyword{...})` call in `init()`. No
business logic; the keywords are thin glue between the keyword
API and the wire engine.

## 2. Architecture Touchpoints

- `pkg/keywords/primitive/` — new; one .go file per keyword
  family (open.go, send.go, expect.go, wait.go, status.go)
- `pkg/keywords/primitive/internal/` — small helpers shared
  across primitives
- Read-only consumers: `pkg/keywords/api`, `pkg/transport`,
  `pkg/wire`, `pkg/engine/clock`
- `examples/stories/primitives_only.story` — new; smoke test
  using only primitives

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| (no exported symbols beyond `init()` registration) | n/a | initial |

The package's external effect is purely `init()`-time
registration. There is no public API surface beyond the
keyword pattern strings.

## 4. Data Contracts

The 10 patterns from spec 004 -10. Each pattern's argument
types are explicit; the resolver coerces step text into
`Args` per spec 003.

Errors returned from primitive keywords:

- `ErrTimeout` (already typed in `pkg/transport`)
- `ErrFrameShape` (already typed in `pkg/wire`)
- `errors.New("station not registered")` for handle lookup
  failures (caught by the runner)

## 5. Required ADRs

- [x] ADR 0007 — Keyword library layering

## 6. Test Strategy

- **Unit tests** per primitive family, against
  `mock.NewMockStation()`. Verify the keyword body sends/
  receives the expected wire shapes.
- **Determinism tests**: `wait` keyword against a deterministic
  clock; assert no real time elapses.
- **Smoke test**: `examples/stories/primitives_only.story`
  executes end-to-end against the pinned CitrineOS, using only
  primitive keywords. Validates spec 004 AC6.

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** N/A.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Primitive layer grows beyond 10–15 keywords | Medium | Medium | Adding a primitive requires amending spec 004 -10; reviewer enforces |
| Primitives leak OCPP semantics | Low | High | Reviewer agent flags any reference to OCPP message names in primitive code |
| `wait` keyword interaction with deterministic clock subtle | Medium | Medium | Integration test exercises the keyword against both clock implementations |

## 9. Effort Estimate

- T-shirt size: **S**
- Calendar estimate: 3–5 days of focused work
- Parallelizable streams: each primitive family is independent;
  all 5 files can be written in parallel

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] Maintainer review
