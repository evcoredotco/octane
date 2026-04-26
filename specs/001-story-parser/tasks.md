# Tasks 001: Story DSL Parser

> **Spec ID:** `001-story-parser`
> **Plan reference:** `./plan.md`
> **Status:** Ready

## Conventions

- ID format: `T-001-MM` (zero-padded)
- One agent per task; multi-agent tasks are split
- `P` = parallel-eligible; `S` = strict ordering after prior task
- AC column references `spec.md` acceptance criteria

---

## Phase 1 — Contracts

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-00 | Bootstrap Go module (`go.mod`, `go.sum`) | backend | S | — | `go.mod` |
| T-001-01 | Define `ast.Story`, `ast.Meta`, `ast.Step`, `ast.Position` | architect | S | AC1, AC4 | `pkg/story/ast/ast.go` |
| T-001-02 | Define `ast.Dependency`, `ast.Scope` enum | architect | P | AC6 | `pkg/story/ast/dependency.go` |
| T-001-03 | Define typed errors `ErrMissingSpecRef`, `ErrSpecRefOnHelper`, `ErrUnboundParameter`, `ErrMalformedDepends` | architect | P | AC2, AC3, AC4, AC6, AC7 | `pkg/story/diag/errors.go` |
| T-001-04 | Define `lex.Token`, `lex.Lexer` interface | architect | S | AC1 | `pkg/story/lex/lex.go` |

## Phase 2 — Lexer

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-10 | Implement byte-stream lexer with line/column tracking | backend | S | AC1, AC2 | `pkg/story/lex/lex.go` |
| T-001-11 | Implement CRLF→LF normalization at lexer boundary | backend | P | AC1 | `pkg/story/lex/lex.go` |
| T-001-12 | Lexer unit tests covering every token type and error path | qa | S | AC1 | `pkg/story/lex/lex_test.go` |

## Phase 3 — Parser core

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-20 | Implement Meta-block parser (Name, Id, Stations, Timeout) | backend | S | AC1, AC2 | `pkg/story/parser_meta.go` |
| T-001-21 | Implement Spec-Ref / helper-tag mutual-exclusion check | backend | S | AC3, AC4 | `pkg/story/parser_meta.go` |
| T-001-22 | Implement `Depends:` YAML-list parser with scope validation | backend | S | AC6 | `pkg/story/parser_depends.go` |
| T-001-23 | Implement `Parameters:` declaration parser | backend | P | AC7 | `pkg/story/parser_meta.go` |
| T-001-24 | Implement Background / Scenario / Teardown step parser | backend | S | AC1 | `pkg/story/parser_steps.go` |
| T-001-25 | Implement parameter-reference validator (cross-check `{x}` against `Parameters:`) | backend | S | AC7 | `pkg/story/parser_validate.go` |
| T-001-26 | Top-level `Parse([]byte) (*ast.Story, error)` entrypoint | backend | S | AC1 | `pkg/story/parse.go` |

## Phase 4 — Determinism

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-30 | Sorted-keys JSON serializer for AST (test-only utility) | backend | P | AC5 | `pkg/story/internal/serialize/serialize.go` |
| T-001-31 | Property test: 1000 random parses → identical JSON across runs | qa | S | AC5 | `pkg/story/parse_determinism_test.go` |

## Phase 5 — Validation suite

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-40 | Golden fixtures for every example story under `scenarios/` | qa | P | AC1, AC8 | `pkg/story/testdata/*.golden.json` |
| T-001-41 | Implement `octane validate stories` test that parses all `scenarios/` | qa | S | AC8 | `pkg/story/scenarios_test.go` |
| T-001-42 | Negative-fixture suite: malformed Meta, missing Spec-Ref, etc. | qa | P | AC2, AC3, AC4, AC6, AC7 | `pkg/story/testdata/negative/*.story` |
| T-001-43 | Property fuzz test: random byte input never panics | qa | P | — | `pkg/story/parse_fuzz_test.go` |

## Phase 6 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-50 | Godoc on every exported symbol; package-level overview | docs | P | — | `pkg/story/*.go` |
| T-001-51 | Update `docs/concepts/story-syntax.md` with parser-validated examples | docs | P | — | `docs/concepts/story-syntax.md` |
| T-001-52 | CHANGELOG entry under `[Unreleased]` | docs | S | — | `CHANGELOG.md` |

## Phase 7 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-001-60 | Security review: parser robustness against adversarial input | security | S | — | — |
| T-001-61 | Reviewer pass: API surface, error messages, godoc | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 8 acceptance criteria covered by at least one task
- [ ] Every task has a single owning agent
- [ ] All `scenarios/` files parse successfully
- [ ] Property fuzz test runs 5 minutes without panic
- [ ] Determinism test passes on Linux/macOS/Windows
- [ ] Security review signed off
- [ ] CHANGELOG updated under `## [Unreleased]`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/001-story-parser` passes
