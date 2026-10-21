# AGENTS.md — OCTANE Agent Contract

This file is the canonical contract for any AI coding agent operating in this
repository (Claude Code, Cursor, Aider, Continue, etc.). Tool-specific
rules live alongside it (`CLAUDE.md`), but the content here is binding
for all of them.

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
- **Language:** Go 1.26
- **License target:** Apache-2.0

## Non-negotiable rules

1. **Read the principles first.** The project principles in
   `ARCHITECTURE.md` section 2 override this file and every prompt. Other
   documents cite them as "constitution principle N". Do not propose
   changes that violate them.
2. **Decision before code.** An architectural change (a new dependency, a
   new package boundary, a change to the story DSL, the cache, or the
   report formats) needs an ADR in `docs/adr/` before the code lands.
   If the user asks you to skip the ADR, push back and explain why.
3. **Stay in your lane.** Each role in the roster below has a declared
   scope. Do not modify files outside that scope without saying so.
4. **Conformance traceability.** Every conformance story carries a
   `Spec-Ref` Meta key. Only stories tagged `helper` may omit it. No
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

## Change workflow

1. Work on a branch; never push directly to `master`.
2. For an architectural change, draft or amend an ADR in `docs/adr/` first.
3. Run `make format`, `make lint` and `make test` before pushing.
4. Open a pull request; it merges when CI is green.

## Agent roster

| Role                    | Scope                                                                           |
|-------------------------|---------------------------------------------------------------------------------|
| Architect               | `docs/adr/`, `ARCHITECTURE.md`                                                  |
| Backend implementer     | `cmd/`, `pkg/` (excluding `pkg/story/` and `pkg/keywords/`)                     |
| Keyword author          | `pkg/story/`, `pkg/keywords/`, `scenarios/`, `docs/keywords/`                   |
| DevOps / Platform       | `.github/`, `action/`, `Dockerfile`, `Makefile`, `.goreleaser.yaml`, `packaging/` |
| QA / Test author        | `*_test.go`, `test/`, fuzz corpora                                              |
| Security reviewer       | read-only across the repo; opens issues only                                    |
| Code reviewer           | comments on diffs; no direct commits                                            |
| Documentation writer    | `docs/` (excluding `docs/keywords/`), `website/`, `README.md`, `CHANGELOG.md`, doc comments |

## Build and test

- `make install-tools` — installs the linter versions that CI pins
- `make format` — gci, gofumpt and golines, as configured in `golangci.yml`
- `make lint` — golangci-lint (which includes staticcheck) and go vet
- `make test` — unit tests with `-race`
- `make test-reference` — full suite against pinned CitrineOS (Docker required)
- `make build` — produces `./bin/octane`

CI runs the same checks with the same linter versions. If `make lint` and
`make test` pass locally, CI should pass.

## Commit and PR conventions

- Conventional Commits (`feat(engine): ...`, `fix(transport): ...`).
- Commits that implement or change an architectural decision reference
  its ADR (e.g. `ADR 0016`).
- All commits GPG-signed (see `.gitconfig` in the QTech onboarding bundle).
- PR description includes: what changed and why, related ADRs, manual test
  evidence, and a CitrineOS reference run link when stories change.

## When in doubt

- Stop, summarize what you understand, and ask before generating code.
- Prefer opening an ADR over inventing a convention.
- Treat the published OCPP specifications as the source of truth for
  scenario semantics; if a specification is silent or ambiguous, file
  an issue tagged `spec-clarification`.
