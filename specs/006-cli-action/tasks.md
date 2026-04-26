# Tasks 006: CLI and Action Surface

> **Spec ID:** `006-cli-action`
> **Plan reference:** `./plan.md`
> **Status:** Ready (provisional — revisit after spec 005 lands)

## Conventions

Same as previous specs.

---

## Phase 1 — Command tree scaffold

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-01 | Set up cobra root command, global flags | backend | S | AC1 | `cmd/octane/main.go`, `cmd/octane/root.go` |
| T-006-02 | Wire `octane run` subcommand | backend | S | AC1, AC2 | `cmd/octane/run.go` |
| T-006-03 | Wire `octane validate stories` subcommand | backend | P | AC6 | `cmd/octane/validate.go` |
| T-006-04 | Wire `octane keywords list/resolve` subcommands | backend | P | — | `cmd/octane/keywords.go` |
| T-006-05 | Wire `octane cache info/prune/clear/key/show/trace` subcommands | backend | P | — | `cmd/octane/cache.go` |
| T-006-06 | Wire `octane completion` subcommand (cobra-stock) | backend | P | AC8 | `cmd/octane/completion.go` |

## Phase 2 — Configuration loader

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-10 | Implement YAML schema struct mirroring ADR 0010 | backend | S | AC3, AC6 | `cmd/octane/internal/config/schema.go` |
| T-006-11 | Implement loader: YAML file → typed config | backend | S | AC3 | `cmd/octane/internal/config/load.go` |
| T-006-12 | Implement env var overlay | backend | S | AC3 | `cmd/octane/internal/config/env.go` |
| T-006-13 | Implement flag-vs-env-vs-yaml resolution chain | backend | S | AC3 | `cmd/octane/internal/config/resolve.go` |
| T-006-14 | Loader unit tests covering every priority path | qa | S | AC3 | `cmd/octane/internal/config/resolve_test.go` |
| T-006-15 | Malformed-YAML test → exit 64 with file+field error | qa | P | AC6 | `cmd/octane/internal/config/load_test.go` |

## Phase 3 — Exit codes

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-20 | Centralize exit-code constants per spec 006 §10 | backend | S | AC2 | `cmd/octane/internal/exitcode/exitcode.go` |
| T-006-21 | Wire each error type to its documented exit code | backend | S | AC2 | (across cmd/octane) |
| T-006-22 | Stability test: every documented code reachable & unique | qa | S | AC2 | `cmd/octane/exitcode_test.go` |
| T-006-23 | Insecure-skip-verify banner test | qa | P | AC7 | `cmd/octane/insecure_test.go` |

## Phase 4 — Action wrapper

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-30 | Update `action/action.yml` with full input/output set per spec 006 §4 | devops | P | AC4 | `action/action.yml` |
| T-006-31 | Implement `action/entrypoint.sh` translating inputs to flags | devops | S | AC4 | `action/entrypoint.sh` |
| T-006-32 | Action smoke workflow `_test-action.yml` | devops | S | AC4 | `.github/workflows/_test-action.yml` |
| T-006-33 | Pin Docker base image; multi-arch build | devops | P | AC5 | `action/Dockerfile` |

## Phase 5 — GitLab integration

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-40 | Validate `examples/ci/gitlab-ci/.gitlab-ci.yml` against pinned CitrineOS | devops | S | AC5, AC10 | (test artifact) |
| T-006-41 | Document Docker image publishing to GHCR | devops | P | AC5 | `docs/distribution.md` |

## Phase 6 — Man pages and completion

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-50 | `make man` generates man pages from cobra tree | devops | S | AC9 | `Makefile`, `scripts/gen-manpages.sh` |
| T-006-51 | Golden test: man page diff against `testdata/man/` | qa | S | AC9 | `cmd/octane/man_test.go` |
| T-006-52 | `make completions` generates bash/zsh/fish completions | devops | S | AC8 | `Makefile`, `scripts/gen-completions.sh` |
| T-006-53 | Completion smoke test: `bash -n` on each output | qa | S | AC8 | `cmd/octane/completion_test.go` |

## Phase 7 — Example workflows validation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-60 | GitHub Actions example workflow runs end-to-end in CI | devops | S | AC10 | `.github/workflows/_test-examples.yml` |
| T-006-61 | GitLab example tested via `gitlab-runner exec docker` smoke | devops | P | AC10 | `scripts/test-gitlab-example.sh` |

## Phase 8 — Documentation

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-70 | `docs/getting-started.md` (5-minute first run) | docs | P | — | `docs/getting-started.md` |
| T-006-71 | `docs/cli-reference.md` (auto-generated from cobra) | docs | P | — | `docs/cli-reference.md` |
| T-006-72 | `docs/configuration.md` (flag/env/YAML inventory) | docs | P | — | `docs/configuration.md` |
| T-006-73 | Update `README.md` quickstart | docs | S | — | `README.md` |
| T-006-74 | CHANGELOG entry | docs | S | — | `CHANGELOG.md` |

## Phase 9 — Review

| ID | Title | Agent | P/S | AC | Files |
|----|-------|-------|-----|----|-------|
| T-006-80 | Security review: credential handling, no secret leaking in logs | security | S | — | — |
| T-006-81 | DevOps review: Action stability, GHCR publishing | devops | S | — | — |
| T-006-82 | Reviewer pass: CLI ergonomics | reviewer | S | — | — |

---

## Definition of Done

- [ ] All 10 acceptance criteria covered by at least one task
- [ ] Example GitHub Actions and GitLab workflows execute green
- [ ] Man pages generated and verified
- [ ] Shell completions parse cleanly in bash/zsh/fish
- [ ] Security review signed off
- [ ] DevOps review of Action surface signed off
- [ ] CHANGELOG updated under `## [Unreleased]`
- [ ] `bash .specify/scripts/bash/check-spec.sh specs/006-cli-action` passes
