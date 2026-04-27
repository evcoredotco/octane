# ADR 0012: Shell Completion — Static + Dynamic, bash and zsh

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, DevOps, Security reviewer
- **Constitution principles touched:** II (Two Distribution Surfaces),
  X (Security and Compliance)

## Context

OCTANE is a CLI users interact with daily during conformance work
(authoring stories, validating config, running suites). Shell
completion materially improves ergonomics — both for static surface
(subcommands, flags) and for high-value dynamic surface (story file
paths, profile names, registered keywords).

Cobra (the CLI framework adopted in ADR 0018) generates static
completions for bash, zsh, fish, and PowerShell out of the box.
Dynamic completion is supported via per-command `ValidArgsFunction`
hooks that the shell invokes via a hidden `__complete` subcommand at
TAB time.

Two design choices need pinning:

1. Which shells are first-class in v1.
2. The security boundary for dynamic completion functions.

## Decision

### Shells supported in v1

| Shell | Status | Install path |
|-------|--------|--------------|
| bash  | first-class | `/usr/share/bash-completion/completions/octane` |
| zsh   | first-class | `/usr/share/zsh/vendor-completions/_octane` |
| fish  | tracked, not v1 | follow-up ADR |
| PowerShell | tracked, not v1 | follow-up ADR |

The `octane completion <shell>` subcommand emits the script for any
of bash, zsh, fish, or PowerShell — cobra produces them all
regardless. Only bash and zsh are *installed automatically* by the
distribution packages in v1. Fish and PowerShell users invoke the
subcommand manually.

### Static completions

- Subcommand names (`run`, `validate`, `keywords`, `profile`, …)
- Global flags and per-command flags
- Enum-typed flag values (e.g. `--format json|robot-xml`)

These are free; cobra emits them from the CLI definition.

### Dynamic completions

The following arguments and flag values are completed dynamically:

| Token | Source | Example |
|-------|--------|---------|
| `octane run <PATH>` | filesystem walk for `*.story` files | `octane run scenarios/<TAB>` |
| `--profile <NAME>` | local cache + Go module index | `octane run --profile <TAB>` |
| `octane keywords show <NAME>` | registered keyword library | `octane keywords show station<TAB>` |
| `--config <PATH>` | filesystem walk for `*.yml`/`*.yaml` | `octane run --config <TAB>` |
| `octane validate <SUBJECT>` | enum: `config`, `story`, `profile` | `octane validate <TAB>` |
| `--ocpp-version <VERSION>` | enum: `1.6`, `1.6`, `2.1` | `--ocpp-version <TAB>` |

### Security boundary — read-only and side-effect-free

This is the load-bearing rule of this ADR.

Completion functions execute the OCTANE binary in a special mode
(`octane __complete ...`) every time the user presses TAB. The
function therefore runs in the user's shell environment with full
filesystem access. A buggy or malicious completion function can:

- Trigger network calls on every keystroke.
- Read files that should remain unread (credentials, history).
- Spawn long-running operations that hang the shell.

To prevent this, **every dynamic completion function in OCTANE must
be read-only and side-effect-free.** Concretely:

- ✅ Allowed: filesystem reads (`os.ReadDir`, `os.Stat`).
- ✅ Allowed: in-process registry queries (keyword library, enum
  values).
- ✅ Allowed: parsing the local config file from disk to suggest
  values it already contains.
- ❌ Forbidden: any network call, including DNS-affecting calls.
- ❌ Forbidden: spawning subprocesses (`exec.Command`).
- ❌ Forbidden: writing to disk for any reason, including caches.
- ❌ Forbidden: reading files outside `~/.config/octane/`,
  `$XDG_CONFIG_HOME/octane/`, the project root, and explicitly
  user-named paths from the partial argument.
- ❌ Forbidden: completion-time profile downloads or Go module fetches.
  If the local cache is empty, completion returns an empty list and
  prints a one-line stderr hint to the shell user.

A linter check enforces this at build time:
`go vet -vettool=$(which octane-complete-vet)` rejects any package
under `pkg/cli/complete/` that imports `net`, `net/http`, `os/exec`,
or `database/sql`. The vettool ships in `internal/tools/`.

### Performance budget

- Static completion must complete in ≤ 5ms (no I/O).
- Dynamic completion must complete in ≤ 50ms at the 95th percentile
  for a workspace of up to 500 stories. CI measures and fails the
  build above this threshold.

### Packaging

- `.deb`/`.rpm` post-install drops the bash and zsh completion
  scripts in their FHS locations (above) and updates the appropriate
  caches (`/etc/bash_completion`, `compinit` is the user's
  responsibility).
- Homebrew uses `bash_completion.install` and `zsh_completion.install`.
- Docker image installs both completions and exposes a doc note in
  the image README.

### Versioning

- Adding a new dynamic completion source is a minor version bump.
- Removing a completion source is a major version bump (it removes
  ergonomic surface).
- Changing static completion (flag names, subcommand names) follows
  the CLI's own semver; man pages and completions update together.

## Consequences

### Positive

- Authoring `.story` files and resolving profiles becomes
  significantly faster at the shell.
- Reviewers and operators discover commands by tabbing rather than
  reading the full `--help`.
- The read-only rule is small enough to enforce mechanically and
  large enough to prevent the well-known classes of completion
  abuse.

### Negative

- Two more scripts in the install footprint (~6 KB each, negligible).
- Maintaining the vettool is OCTANE's burden. Mitigated by keeping
  the rule list short and stable; the vettool is ~150 LoC.

### Neutral

- Fish and PowerShell users get a usable but unmanaged completion via
  `octane completion <shell>`. Promoting them to first-class in a
  later ADR is mechanical: add packaging hooks, no design work.

## Alternatives considered

- **Static completions only.** Rejected: dynamic completion is the
  reason `kubectl get pods <TAB>` feels modern and `octane run
  --profile <TAB>` would feel obviously stale.
- **Allow network in completion.** Rejected: predictable, exploitable
  attack surface and a hostile UX (lag on every TAB).
- **Cache completion results in `~/.cache/octane/complete/`.**
  Considered. Rejected for v1: complicates the read-only rule and
  introduces stale-cache bugs; revisit if the 50ms budget is
  routinely missed.

## References

- Constitution: principles II, X
- ADR 0018 (CLI surface)
- ADR 0019 (config schema)
- ADR 0020 (distribution)
- ADR 0011 (manual pages)
- Cobra completion docs:
  https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md
