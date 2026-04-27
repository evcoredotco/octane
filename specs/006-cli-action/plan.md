# Plan 006: CLI and Action Surface

> **Spec ID:** `006-cli-action`
> **Status:** Approved (provisional — depends on spec 005)
> **Author:** Alexis Sánchez

---

## 1. Summary

Build the operator-facing surface: the `octane` CLI binary using
cobra, plus a thin `entrypoint.sh` wrapper for the GitHub Action.
Both call into `pkg/runner.Run(ctx, cfg)`. The Action is a Docker
image action (already scaffolded under `action/`) that translates
inputs to CLI flags.

> **Note:** Some details (notably `--shard` flag wiring) depend
> on spec 005's final shape. Marked provisional until spec 005
> is implemented.

## 2. Architecture Touchpoints

- `cmd/octane/` — new; cobra command tree
- `cmd/octane/run.go`, `cmd/octane/cache.go`, etc. — one file per top-level subcommand
- `cmd/octane/internal/config/` — new; YAML loader + resolution chain
- `action/` — already scaffolded; Dockerfile, action.yml, entrypoint.sh
- `examples/ci/` — already scaffolded; verify they execute against CitrineOS
- Read-only consumers: `pkg/runner`, `pkg/story`, `pkg/cache`,
  `pkg/keywords/registry`

## 3. Public API Changes

| Symbol | Change | Semver impact |
|--------|--------|---------------|
| `octane run <stories...>` | new command | initial |
| `octane validate stories <stories...>` | new command | initial |
| `octane keywords list` | new command | initial |
| `octane keywords resolve --story <path>` | new command | initial |
| `octane cache info`/`prune`/`clear`/`key`/`show`/`trace` | new commands | initial |
| `octane completion <shell>` | new command (cobra-stock) | initial |
| Exit codes per spec 006 §10 | new contract | initial |
| Env vars `OCTANE_*` per spec 006 §11 | new contract | initial |

The CLI is a public contract from day one; flag and env names
follow semver from the first tagged release.

## 4. Data Contracts

### `octane.yml` schema

Defined in ADR 0010. The CLI's loader validates against the
schema; unknown keys return exit 64 with a clear error. The
schema is not duplicated here.

### Action inputs (action.yml)

```yaml
inputs:
  stories:
    description: "Path or glob of .story files"
    required: true
  fail-on:
    description: "Lowest finding severity that fails the run (info|minor|major|critical)"
    default: "major"
  config:
    description: "Path to octane.yml"
    default: "octane.yml"
  cache-dir:
    description: "Cache directory; defaults to OCTANE_CACHE_DIR or XDG default"
  report-dir:
    description: "Output directory for reports"
    default: "reports/"
  ocpp-version:
    description: "Override OCPP version (1.6)"
    default: ""
outputs:
  report-path:
    description: "Path to the JSON report"
  exit-code:
    description: "Numeric exit code from octane run"
```

`entrypoint.sh` translates each input to its CLI flag and runs
the binary.

## 5. Required ADRs

- [x] ADR 0010 — Connection profiles (the YAML schema CLI loads)
- [x] ADR 0011 — Manual pages (cobra → scdoc)
- [x] ADR 0012 — Shell completion (cobra-stock)

## 6. Test Strategy

- **Unit tests** of `cmd/octane/internal/config/`: every
  resolution-chain path exercised. Flag wins, env wins, YAML
  wins, default wins.
- **Unit tests** of cobra command tree: invoke each command with
  `--help`; assert no panic, golden help-text matching.
- **Integration test**: `octane run` against pinned CitrineOS;
  assert exit code 0 and report file written.
- **Action surface test**: a workflow under
  `.github/workflows/_test-action.yml` invokes the local Action
  against CitrineOS and uploads the artifact (AC4).
- **Example workflows test**: `examples/ci/github-actions/` and
  `examples/ci/gitlab-ci/` execute end-to-end as part of CI
  (AC10).
- **Man-page golden test**: `make man` is run; the output is
  diffed against `testdata/man/` (AC9).
- **Completion test**: `octane completion bash | bash -n` (and
  zsh, fish equivalents); assert no syntax errors (AC8).

## 7. Rollout

- **Feature flag:** none.
- **Backwards compatibility:** Pre-1.0; flag set is new.
- **Migration:** N/A.

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| cobra major-version churn | Low | Medium | Pin minor version |
| Action input semantics drift from CLI flags | Medium | High | `entrypoint.sh` is generated from a single source of truth (a YAML map); golden test asserts consistency |
| Exit codes accidentally renumbered | Low | High | Stability test asserts every documented code is reachable and unique |
| `octane.yml` schema drift between CLI and ADR 0010 | Medium | Medium | Schema lives in one Go file; ADR references it by line number |
| Speculative `--shard` flag misses spec 005's actual shape | Medium | Low | Marked provisional; revisit after spec 005 lands |

## 9. Effort Estimate

- T-shirt size: **M**
- Calendar estimate: 1.5–2 weeks of focused work
- Parallelizable streams: command tree + config loader can
  develop in parallel; Action wrapper is independent

---

## Approval

- [x] Architect / Spec author
- [ ] Backend implementer (sanity check after spec 005 lands)
- [x] DevOps / Platform
- [x] Maintainer review
