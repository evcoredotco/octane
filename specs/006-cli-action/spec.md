# Spec 006: CLI and Action Surface

> **Spec ID:** `006-cli-action`
> **Status:** Approved (provisional — depends on spec 005 final shape)
> **Author:** Alexis Sánchez
> **Created:** 2026-04-26
> **Constitution version:** 1.4.0

---

## 1. Problem Statement

OCTANE has two distribution surfaces per ADR 0010 and constitution
principle II: a CLI binary (`octane run …`) and a GitHub Action
(plus a documented GitLab integration). Both invoke the same
runner code path; their differences are in surface ergonomics
(flags vs YAML inputs, exit codes vs job statuses, reports as
files vs reports as artifacts).

This spec defines:

- The complete `octane` CLI command tree, flag set, exit codes,
  and configuration resolution chain.
- The GitHub Action's `action.yml` inputs, outputs, and the
  `entrypoint.sh` that bridges Action inputs to CLI flags.
- The GitLab CI integration pattern (no GitLab-specific binary
  surface; it's a Docker image plus environment variables).
- The `OCTANE_CACHE_DIR` env var and other CI integration
  contracts already named in ADR 0016.

## 2. Goals

- G1. Implement `cmd/octane/` using `spf13/cobra` (per ADR 0011
      for shell completion compatibility).
- G2. Stable exit codes per the table in -6.
- G3. Configuration resolution chain (per ADR 0010): CLI flags
      override env vars override `octane.yml` overrides defaults.
- G4. The GitHub Action is a thin wrapper: `action.yml` declares
      inputs, `entrypoint.sh` translates to CLI flags, runs the
      binary, exposes outputs.
- G5. Example workflows in `examples/ci/github-actions/` and
      `examples/ci/gitlab-ci/` (already shipped) execute end-to-end
      against the pinned CitrineOS.
- G6. Man pages (per ADR 0011) and shell completions (per ADR
      0012) generated from the cobra command tree.

## 3. Non-Goals

- N1. Writing the runner (spec 005).
- N2. Writing the report emitter (spec 007).
- N3. Connection profile *format* (defined in ADR 0010); only the
      resolution chain is in scope here.
- N4. CSMS-specific authentication adapters (forbidden by
      constitution principle XII).
- N5. Distribution channel publishing (`.deb`, `.rpm`, Homebrew —
      handled by `goreleaser`, already configured).

## 4. User Stories

- **As an operator running OCTANE locally**, I want
  `octane run scenarios/v16/` to do the right thing with sensible
  defaults, then offer flags for every override I might need.
- **As a CI maintainer using GitHub Actions**, I want a single
  `uses: evcoreco/octane-action@v0` block with named inputs
  that hide the CLI surface entirely.
- **As a GitLab CI user**, I want a published Docker image I can
  drop into my pipeline with environment-variable configuration.
- **As a developer debugging a CI run**, I want exit codes that
  unambiguously distinguish "tests failed" from "tool failure"
  from "configuration error", so my workflow's
  `continue-on-error` decisions are correct.

## 5. Constraints from the Constitution

| Principle | Constraint |
|-----------|------------|
| II. Two Distribution Surfaces, One Engine | The CLI and Action MUST share `cmd/octane/` code paths; the Action is `entrypoint.sh` calling `octane run` with translated flags. No Action-specific Go code. |
| V. Stdlib-Heavy | `spf13/cobra` is the only CLI dependency. `viper` is *not* used for config; ADR 0010 specifies a hand-rolled YAML loader against a fixed schema. |
| X. Security | The CLI MUST NOT log credentials. Connection profile auth blocks are redacted in any operator-facing output. |
| XI. Wire Conformance Only | The CLI never reads CSMS state; `octane csms-status` and similar commands are forbidden. |

## 6. Acceptance Criteria

- AC1. **Given** `octane run scenarios/v16/`, **when** the
       command executes against the pinned CitrineOS, **then**
       it returns exit code 0 on suite-pass and exit code 1 on
       any-fail; output is the report path and a one-line summary.
- AC2. **Given** the documented exit code table in -10, **when**
       any documented condition is reached, **then** the CLI
       exits with the matching code. Documented codes are stable
       across releases.
- AC3. **Given** a flag, an env var, and an `octane.yml` value
       all set for the same parameter, **when** `octane run` is
       invoked, **then** the flag wins. **Given** the flag is
       absent, **when** the env var is set, **then** the env var
       wins. **Given** both are absent, **when** `octane.yml`
       sets the value, **then** the YAML value wins.
- AC4. **Given** the GitHub Action used as
       `uses: evcoreco/octane-action@v0` with inputs
       `stories: scenarios/v16/` and `fail-on: major`, **when**
       the workflow runs, **then** the binary executes with
       equivalent CLI flags and the action's `report-path`
       output is set.
- AC5. **Given** the Docker image `ghcr.io/evcoreco/octane`,
       **when** invoked from a GitLab CI job with
       `OCTANE_CACHE_DIR` set, **then** the cache directory is
       used and the cache mechanism in spec 005 / ADR 0016
       activates.
- AC6. **Given** an invalid `octane.yml` (malformed YAML, missing
       required fields), **when** the CLI parses it, **then** it
       exits with code 64 (config error) and prints a message
       citing the file and the offending field.
- AC7. **Given** `--insecure-skip-verify` is set, **when**
       the CLI runs, **then** it emits a banner-level finding
       in the report flagging the unsafe configuration.
- AC8. **Given** `octane completion bash` (and `zsh`, `fish`),
       **when** the command executes, **then** it writes a
       valid completion script to stdout that auto-completes
       commands, flags, and known story paths.
- AC9. **Given** the man page targets in the Makefile, **when**
       `make man` runs, **then** `man octane`, `man octane-run`,
       `man octane-cache`, etc., are generated and consistent
       with `--help` output (golden test compares them).
- AC10. **Given** the example workflows in `examples/ci/`,
        **when** they run end-to-end against the pinned
        CitrineOS, **then** all assertions pass and report
        artifacts are uploaded to the expected paths.

## 7. OCPP Scope

The CLI is OCPP-version-agnostic. Version selection is implicit
(via story Meta) or explicit (via `--ocpp-version` flag, used
when stories don't declare).

## 8. Open Questions

- OQ1. Whether to ship a `--watch` mode that re-runs on file
       changes (developer convenience). Recommendation: yes, in
       a follow-up spec, not this one. Watch mode adds a
       filesystem-events dependency and complicates exit-code
       semantics.
       *(owner: Architect, due: with this spec — DEFERRED to
       a follow-up.)*
- OQ2. Whether `octane.yml` should support YAML anchors/aliases
       for shared configuration blocks. Recommendation: no.
       Anchors complicate the loader and create non-obvious
       inheritance paths. Use environment variables for shared
       config across pipelines instead.
       *(owner: Architect, due: with this spec — RESOLVED.)*
- OQ3. **(speculative — depends on spec 005 implementation)**
       Whether `--shard N/M` flags should be CLI-level or
       runner-level. Recommendation: CLI-level, since CI matrix
       jobs set them per-shard from the matrix index. Final
       decision deferred to spec 005 implementation.
       *(owner: Backend, due: spec 005 implementation.)*

## 9. Out of Scope (parking lot)

- `--watch` mode for local development.
- Interactive `octane init` wizard for new projects.
- Configuration file migration between major versions.
- A `octane upgrade` self-updater (operators manage their
  package manager).
- Telemetry / phone-home (forbidden by Anthropic-style
  open-source ethic and constitution principle XIII).

## 10. Exit code table

| Code | Meaning |
|------|---------|
| 0 | All scenarios passed (or were skipped due to upstream filters). |
| 1 | At least one conformance assertion failed. |
| 2 | Tool error (panic, internal bug). |
| 3 | Reserved for future use. |
| 9 | Cache lock contention exceeded `--lock-timeout` (per ADR 0016). |
| 64 | Configuration error (malformed `octane.yml`, unknown flag, missing required input). |
| 65 | Story parse error (per spec 001). |
| 66 | Keyword resolution error (per spec 003). |
| 70 | Wire transport error (TLS, DNS, subprotocol mismatch — per spec 002). |

Codes follow `sysexits.h` conventions where possible. New codes
are added by amending this spec; never opportunistically.

## 11. Configuration resolution chain

```
CLI flags (highest priority)
    ↓
Environment variables (OCTANE_*)
    ↓
octane.yml at the working directory
    ↓
Built-in defaults (lowest priority)
```

Environment variable names are derived from flag names by
upper-snake-casing and prepending `OCTANE_`:

```
--cache-dir       →  OCTANE_CACHE_DIR
--lock-timeout    →  OCTANE_LOCK_TIMEOUT
--max-parallel    →  OCTANE_MAX_PARALLEL
--csms-url        →  OCTANE_CSMS_URL
```

The `octane.yml` schema is defined in ADR 0010.

---

## Approval

- [x] Architect / Spec author
- [ ] Backend implementer (sanity check after spec 005 lands)
- [x] DevOps / Platform
- [x] Maintainer review

> **Note:** This spec depends on the runner shape from spec 005.
> Some details (especially OQ3 on `--shard`) may need
> revision once spec 005 is implemented. Marked Approved
> Provisional rather than Approved.
