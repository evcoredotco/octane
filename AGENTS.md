# AGENTS.md — OCTANE Agent Contract

This file is the canonical contract for any AI coding agent operating in this
repository (Claude Code, Cursor, Aider, Continue, etc.). Tool-specific
configurations live alongside it (`CLAUDE.md`, `.cursor/rules/`), but the
content here is binding for all of them.

## Project at a glance

- **Name:** OCTANE — OCPP Conformance Testing & Network Evaluation
- **Goal:** Hardened, scriptable conformance harness for OCPP 1.6J
- **Architecture:** Story-driven framework (per ADR 0005) — `.story`
  files in a Gherkin-flavored DSL drive the CSMS over the OCPP wire,
  parameterized by user-owned connection metadata (ADR 0010).
  Domain keywords are identical across CSMSes (constitution
  principle XII).
- **Surfaces:** `octane` Go CLI + `octane-action` GitHub Action
- **Reference CSMS during dev:** CitrineOS (pinned commit in
  `test/reference/citrineos.version`); a sample connection profile
  ships at `connections/citrineos.yaml`.
- **Language:** Go 1.23
- **License target:** Apache-2.0

## Non-negotiable rules

1. **Read the constitution first.** `.specify/memory/constitution.md` overrides
   this file and every prompt. Do not propose changes that violate it.
2. **Spec before code.** Never produce production code for a feature that does
   not have a merged spec at `specs/NNN-feature-slug/spec.md`. If the user
   asks you to skip the spec, push back and explain why.
3. **Stay in your lane.** Each agent has a declared scope under
   `.claude/agents/`. Do not modify files outside that scope; delegate to the
   correct agent instead.
4. **Conformance traceability.** Every test case carries an `spec_ref`. No
   exceptions for "small additions."
5. **Determinism.** No `time.Now()`, `rand.Int*` (without explicit seeding), or
   map iteration order assumptions in engine code. Use the injected
   `Clock` and `Rand` interfaces from `pkg/engine`.
6. **Stdlib first.** New dependencies require an ADR. Existing pinned
   dependencies are listed in `go.mod` — extend them, do not replace.
7. **Two surfaces, one engine.** Anything reachable from the CLI must also be
   reachable from the GitHub Action and vice versa.
8. **No secrets in code, fixtures, or reports.** Use environment variables and
   document them in `docs/configuration.md`.
9. **All OCPP 1.6 data types come from `github.com/evcoreco/ocpp16types`.**
   This is an absolute rule with no exceptions (ADR 0020). Never declare a
   local struct, type alias, or shadow copy of any OCPP 1.6 message,
   enumeration, or sub-object. If the type is absent from the shared module,
   contribute it upstream and block the OCTANE task until the release is tagged.

## Spec-driven workflow

```
/specify "feature description"   → specs/NNN-feature/spec.md
/plan                            → specs/NNN-feature/plan.md
/tasks                           → specs/NNN-feature/tasks.md
/implement T-NNN-MM              → executes a single task with the right agent
```

Slash commands are defined under `.claude/commands/`. The scripts they invoke
live under `.specify/scripts/bash/` and are tool-agnostic (any agent can shell
out to them).

## Agent roster

| Role | File | Scope |
|------|------|-------|
| Architect / Spec author | `.claude/agents/architect.md` | `specs/`, `docs/adr/`, `.specify/` |
| Backend implementer    | `.claude/agents/backend.md`   | `cmd/`, `internal/`, `pkg/` (excluding `pkg/story/` and `pkg/keywords/`) |
| Keyword author         | `.claude/agents/keyword-author.md` | `pkg/story/`, `pkg/keywords/`, `scenarios/`, `docs/keywords/` |
| DevOps / Platform      | `.claude/agents/devops.md`    | `.github/`, `action/`, `Dockerfile`, `Makefile` |
| QA / Test author       | `.claude/agents/qa.md`        | `*_test.go`, `test/`, fuzz corpora |
| Security reviewer      | `.claude/agents/security.md`  | read-only across the repo; opens issues only |
| Code reviewer          | `.claude/agents/reviewer.md`  | comments on diffs; no direct commits |
| Documentation writer   | `.claude/agents/docs.md`      | `docs/` (excluding `docs/keywords/`), `README.md`, `CHANGELOG.md`, doc comments |

## Build and test

- `make format` — gofmt, gofumpt, golines, gci
- `make lint` — golangci-lint, go vet, staticcheck
- `make test` — unit tests with `-race`
- `make test-reference` — full suite against pinned CitrineOS (Docker required)
- `make build` — produces `./bin/octane`

CI mirrors these targets exactly. If `make` works locally, CI passes.

## Commit and PR conventions

- Conventional Commits (`feat(engine): ...`, `fix(transport): ...`).
- Every commit references a task ID from `tasks.md` (e.g. `T-001-04`) or an
  ADR ID for governance changes.
- All commits GPG-signed (see `.gitconfig` in the QTech onboarding bundle).
- PR description includes: spec link, task IDs closed, manual test evidence,
  CitrineOS reference run link.

## When in doubt

- Stop, summarize what you understand, and ask before generating code.
- Prefer opening an ADR over inventing a convention.
- Treat the published OCPP specifications as the source of truth for
  scenario semantics; if a specification is silent or ambiguous, file
  an issue tagged `spec-clarification`.
