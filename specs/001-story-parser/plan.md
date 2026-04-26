# Plan 001: Story DSL Parser

> **Spec ID:** `001-story-parser`
> **Status:** Approved
> **Author:** Alexis Sánchez

---

## 1. Summary

Implement a hand-written recursive-descent parser for the `.story`
DSL in `pkg/story/`. Input is a byte slice; output is a typed AST
with position information on every node. No I/O, no concurrency,
no third-party parsing libraries. The parser is pure
text-to-struct.

## 2. Architecture Touchpoints

- `pkg/story/` — new package, the parser itself
- `pkg/story/ast/` — typed AST node definitions
- `pkg/story/lex/` — lexer
- `pkg/story/diag/` — diagnostic types (ErrMissingSpecRef, etc.)
- `pkg/story/testdata/` — golden parser-output fixtures
- `scenarios/` — read-only consumer (parse every example story
  in CI)

No other packages are touched. Spec 003 will depend on this
package; specs 005–007 will consume the AST through spec 003.

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `pkg/story.Parse([]byte) (*ast.Story, error)` | new | initial |
| `pkg/story/ast.Story` | new struct | initial |
| `pkg/story/ast.Step`, `Background`, `Scenario`, `Teardown` | new structs | initial |
| `pkg/story/diag.ErrMissingSpecRef` | new typed error | initial |
| `pkg/story/diag.ErrSpecRefOnHelper` | new typed error | initial |
| `pkg/story/diag.ErrUnboundParameter` | new typed error | initial |

Pre-1.0; semver applies after the first tagged release.

## 4. Data Contracts

### AST shape (informative)

```go
type Story struct {
    Path       string
    Meta       Meta
    Background []Step
    Scenarios  []Scenario
    Teardown   []Step
    Position   Position // line:col of file start
}

type Meta struct {
    Name       string
    Id         string
    SpecRef    *string  // nil for helper stories
    Tags       []string // "helper" tag is structural
    Stations   int
    Timeout    time.Duration
    Parameters []string
    Depends    []Dependency
    CacheTTL   *time.Duration
}

type Step struct {
    Kind     StepKind // Given, When, Then, And, But
    Text     string   // verbatim, with {placeholder} tokens preserved
    Position Position
}
```

Field ordering matches the textual order in the source file.
Serialization sorts on each level by stable, documented criteria
(per spec AC5).

## 5. Required ADRs

- [x] ADR 0006 — Story DSL grammar
- [x] ADR 0014 — IP and authoring guidelines (helper-vs-conformance
      distinction)
- [x] ADR 0015 — Test dependency graph (the `Depends:` block this
      parser validates)

No new ADRs needed.

## 6. Test Strategy

- **Unit tests** (`pkg/story/...`): every parser branch covered.
  Particular focus on error paths — every typed error is
  triggered by at least one fixture.
- **Golden tests**: every `.story` file under `scenarios/` has a
  corresponding `<file>.story.golden.json` checked into
  `testdata/`. Mismatch fails CI; intentional updates require
  `go test -update`.
- **Property tests** (`testing/quick`): fuzz the parser with
  random byte inputs; assert it never panics, always returns an
  error or a valid AST.
- **Round-trip test**: parse → serialize sorted JSON → parse
  again from a re-emitted form (if a serializer is ever added);
  for now, just assert byte-determinism of the JSON output of
  the same input.

## 7. Rollout

- **Feature flag:** none. Parser is internal infrastructure.
- **Backwards compatibility:** N/A (no prior parser).
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Grammar drift between ADR 0006 and the parser | Medium | High | Lock ADR 0006 before implementation; review changes in lockstep |
| Determinism leak via map iteration | Low | High | Forbid `map[string]X` in serialized output; use sorted slices |
| Error messages too cryptic for authors | Medium | Medium | Every typed error carries a `Suggestion` field; reviewer agent checks fixtures |
| Parser becomes a hot performance path | Low | Low | Stories are tiny; benchmark only if a >1MB story exists in practice |

## 9. Effort Estimate

- T-shirt size: **M**
- Calendar estimate: 1.5–2 weeks of focused work
- Parallelizable streams: lexer + AST types + diagnostic types
  can advance in parallel after the contracts land

---

## Approval

- [x] Architect / Spec author
- [x] Backend implementer
- [x] QA reviewer (golden fixture strategy)
- [x] Maintainer review
