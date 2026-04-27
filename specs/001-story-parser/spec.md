# Spec 001: Story DSL Parser

> **Spec ID:** `001-story-parser`
> **Status:** Approved
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCTANE's user-facing surface for declaring tests is the `.story`
file (per ADR 0006). Every other component — the runtime, the
keyword library, the dependency resolver, the cache, the report
emitter — reads its inputs from a parsed story. Until a working
parser exists, nothing else can run.

The parser must:

- Read `.story` files in the Gherkin-flavored DSL grammar pinned
  by ADR 0006.
- Validate the Meta section against the schema, distinguishing
  conformance stories (require `Spec-Ref`) from helper stories
  (forbid `Spec-Ref`, require `helper` tag) per ADR 0014.
- Surface human-readable errors with line and column information.
- Produce a typed AST that downstream consumers (resolver, runner)
  can walk without re-parsing.
- Be deterministic: parsing the same bytes twice produces
  byte-identical AST representations (constitution principle IV).
- Have no I/O, no concurrency, no side effects. Pure
  text-to-struct.

This spec is the smallest implementable unit in the project. It
unblocks specs 003–007.

## 2. Goals

- G1. Implement a recursive-descent parser for the `.story` grammar
      defined in ADR 0006.
- G2. Distinguish conformance stories from helper stories at parse
      time using the rules in ADR 0014.
- G3. Surface diagnostic errors with `(line, column, message,
      suggestion)` for every parse failure.
- G4. Produce a typed AST with stable Go field ordering so that
      downstream serialization is deterministic.
- G5. Validate every example story under `scenarios/` parses
      successfully.

## 3. Non-Goals

- N1. Step-text-to-keyword resolution (that is spec 003's
      problem).
- N2. Dependency-graph traversal (spec 005).
- N3. Story execution against a CSMS (specs 002 + 005).
- N4. LSP / IDE integration (deferred indefinitely).
- N5. Story file *generation* (only consumption is in scope).

## 4. User Stories

- **As a story author**, I want clear error messages when I make a
  syntax mistake, so I can fix it without consulting the parser
  source.
- **As a CI maintainer**, I want `octane validate stories` to fail
  loudly on malformed stories before any wire activity begins.
- **As a downstream component author** (runner, resolver, reporter),
  I want a typed AST whose Go fields are documented so I can
  consume parser output without re-parsing.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| IV. Determinism | Parser output is byte-deterministic. Map iteration order is forbidden in serialization paths; use sorted slices. |
| V. Stdlib-Heavy | Parser uses only Go stdlib. No third-party parsing libraries (no PEG generators, no participle, no goyacc). |
| VI. Test Cases as Code | The parser is the typed-Go layer behind the declarative DSL surface. |
| VIII. Spec-Driven | This spec is approved before any code lands. |

## 6. Acceptance Criteria

- AC1. **Given** any `.story` file under `scenarios/`, **when**
       the parser runs, **then** it produces a complete typed AST
       with all Meta keys, Background, Scenario, and Teardown
       sections populated correctly.
- AC2. **Given** a story missing a required Meta key (`Name`,
       `Id`, `Stations`), **when** the parser runs, **then** it
       returns an error citing the file, line, column, and the
       missing key by name.
- AC3. **Given** a conformance story (no `helper` tag) without
       `Spec-Ref`, **when** the parser runs, **then** it returns
       a typed `ErrMissingSpecRef` with line and column.
- AC4. **Given** a helper story (tagged `helper`) with a `Spec-Ref`
       key present, **when** the parser runs, **then** it returns
       a typed `ErrSpecRefOnHelper` with line and column.
- AC5. **Given** the same `.story` bytes parsed twice, **when**
       the resulting ASTs are JSON-serialized with sorted keys,
       **then** the serialized output is byte-identical.
- AC6. **Given** a story whose `Depends:` block is malformed (not
       a YAML list, missing `id`, unknown `scope` value), **when**
       the parser runs, **then** it returns a structured error
       identifying which dependency entry is malformed.
- AC7. **Given** a syntactically valid story whose step text
       references unbound parameters (e.g., `{idTag}` not declared
       in `Parameters:`), **when** the parser runs, **then** it
       returns `ErrUnboundParameter` listing every unbound
       reference.
- AC8. **Given** the existing 10 `.story` files under
       `scenarios/v16/` and `scenarios/`, **when**
       `octane validate stories` runs, **then** all 10 parse
       successfully and the exit code is 0.

## 7. OCPP Scope

Not applicable. The parser operates on `.story` files
independently of OCPP version; OCPP-specific validation lives in
downstream specs.

## 8. Open Questions

- OQ1. Whether to support Windows line endings (CRLF) in addition
       to LF. Recommendation: yes, normalize at the lexer level;
       no separate CRLF parser path.
       *(owner: Architect, due: with implementation)*
- OQ2. Whether the parser produces position information for every
       AST node or only for top-level Meta keys. Recommendation:
       every node, since downstream tooling (the keyword resolver
       in spec 003) wants Levenshtein hints attached to step
       tokens.
       *(owner: Architect, due: with implementation)*

## 9. Out of Scope (parking lot)

- Auto-formatting of `.story` files (`octane fmt`).
- LSP server for editor integration.
- Story file *generation* from OCPP message schemas.
- Cross-version schema evolution (future ADR if needed).

---

## Approval

- [x] Architect / Spec author
- [x] Maintainer review
