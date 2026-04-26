# Contributing to OCTANE

Thank you for considering a contribution to OCTANE. This document
covers what you need to know to author conformance scenarios,
keyword library entries, and supporting code in line with the
project's conventions.

> **Read first:** [`.specify/memory/constitution.md`](./.specify/memory/constitution.md).
> Every contribution must comply with the constitution; the
> conventions in this file operationalize it. If anything in this
> file conflicts with the constitution, the constitution wins.

## Table of contents

- [Spec-driven development](#spec-driven-development)
- [Authoring conformance stories](#authoring-conformance-stories)
- [Authoring helper stories](#authoring-helper-stories)
- [Adding keywords](#adding-keywords)
- [Code style](#code-style)
- [Commits and PRs](#commits-and-prs)

## Spec-driven development

OCTANE follows a strict spec-driven workflow (constitution principle
VIII). Code does not land before its spec merges:

1. `/specify <feature>` — draft `specs/NNN-feature/spec.md`
2. `/plan` — fill `plan.md` with technical approach + ADR drafts
3. `/tasks` — decompose into atomic, agent-assignable tasks
4. `/implement T-NNN-MM` — execute one task

For trivial fixes (typos, comment improvements, single-line bug
fixes), open a PR directly without a spec. The reviewer will tell
you if a spec is needed.

## Authoring conformance stories

OCTANE conformance stories are independent original work derived
from the published OCPP specifications. The rules in this section
operationalize ADR 0014 (IP and authoring guidelines) and apply
without exception.

### Source of truth

Author from the OCPP specification document, not from any
third-party test catalog or testing tool's documentation. The
specification is the public, canonical description of what
conformant CSMS behavior looks like; that is what OCTANE tests.

In practice:

- Open the relevant OCPP spec PDF (1.6J, 2.0.1, or 2.1).
- Locate the section describing the message or behavior you want
  to test.
- Read the normative text — request schema, response schema,
  state-machine transitions, error conditions.
- Write a story whose `Spec-Ref` Meta key cites that section, and
  whose assertions express what the specification requires.
- Write the prose narrative (the comment block at the top of the
  story file) in your own words. Describe what the test does, why
  it matters for conformance, and what state it assumes.

If you find yourself reaching for a third-party catalog to
"translate" its description into OCTANE form, stop. That is the
exact pattern ADR 0014 forbids. Go back to the OCPP specification.

### Naming convention

Story filenames and IDs follow `<resource>_<function>_<desire>`
(snake_case lowercase). Examples:

| Filename | What it tests |
|----------|---------------|
| `boot_notification_accepted.story` | Successful boot registration |
| `boot_notification_malformed.story` | Wire-level rejection of malformed boot |
| `connector_reservation_faulted.story` | CSMS handles a Faulted reservation response |
| `authorize_concurrent_rejected.story` | Concurrent authorize attempt rejected |

The `desire` slot prefers a specific protocol-level state when one
applies (`faulted`, `concurrenttx`, `accepted`) over a generic
outcome category (`success`, `failure`).

### Required Meta keys for conformance stories

```
Meta
    Name:        <human-readable name in your own words>
    Id:          <snake_case slug matching filename>
    Spec-Ref:    OCPP <version> §<section> <message-or-behavior>
    Tags:        <comma list, must include one of:
                  wire-only | multi-station | operator-assisted>
    Stations:    <integer >= 1>
    Timeout:     <duration; optional, default from config>
    Parameters:  <comma list of names referenced in steps; optional>
    Depends:     <YAML list of prereq IDs; optional>
```

`Spec-Ref` MUST cite the OCPP specification, not a third-party
testing tool. The format is one of:

- `OCPP 2.0.1 §B01 BootNotification`
- `OCPP-J 1.6 §6.40 ReserveNow`
- `OCPP 2.1 §C01 Authorize`

### Required prose comment block

The first non-blank lines of every story file are a `#`-prefixed
narrative explaining what the test does and what it depends on.
This is the equivalent of a function docstring. Write it in your
own words. Do not copy from any third-party source.

Example:

```
# Validates that a CSMS implementing OCPP 2.0.1 §B01 BootNotification
# replies to a well-formed BootNotification.req with a
# BootNotificationResponse carrying status "Accepted" and a
# heartbeatInterval within the spec-permitted range.
#
# Single-station, wire-only conformance test. Depends on a
# successful WebSocket connection but assumes no prior CSMS state.
```

## Authoring helper stories

Helper stories exist to bring the system to a known state so that
downstream conformance tests can run from a defined starting point.

Differences from conformance stories:

- **No `Spec-Ref`** — helpers do not assert conformance to a
  specification section in their own right.
- **Tag `helper`** — required.
- **Filename matches the ID** — kebab-case snake_case as before.
- **Lives alongside conformance stories** — under
  `scenarios/v16/`, `scenarios/v201/`, etc. (no separate
  `helpers/` directory).

The parser enforces the distinction: a story tagged `helper` MUST
omit `Spec-Ref`, and a story not tagged `helper` MUST include it.

## Adding keywords

Keywords are typed Go functions that map step text to wire actions.
They live under `pkg/keywords/`:

- `pkg/keywords/api/`         — public surface (do not modify lightly)
- `pkg/keywords/registry/`    — self-registration mechanism
- `pkg/keywords/primitive/`   — transport-level escape hatches
- `pkg/keywords/domain/v16/`  — OCPP 1.6 keywords
- `pkg/keywords/domain/v201/` — OCPP 2.0.1 keywords
- `pkg/keywords/domain/v21/`  — OCPP 2.1 keywords

Each keyword registers exactly one pattern. Domain keywords are
identical for every CSMS implementing the OCPP version they target;
there is no per-CSMS override layer (constitution principle XII).

When adding a domain keyword, include:

1. The pattern (with `{name:type}` placeholders).
2. The implementation Go function.
3. A black-box test in the same package's `_test.go` file
   asserting the pattern is registered and the function exhibits
   the documented behavior on representative inputs.

## Code style

Go code follows the conventions in
[`mnt/skills/user/golang-master/SKILL.md`](https://example.invalid/golang-master)
where applicable: gofmt, line length 80, function complexity ≤ 7,
no `time.Now()` (use the injected clock), no global state outside
the registry.

Run before pushing:

```bash
make format     # gofumpt + goimports
make lint       # golangci-lint
make test       # go test -race ./...
make spec-check # validates spec structure
```

## Commits and PRs

Commits follow the Conventional Commits specification with a JIRA
prefix when applicable, GPG-signed. Use the `qtech-commit` skill
if you have it.

PRs are scoped: one feature, one spec, one ADR, or one fix. PRs
larger than ~600 lines are split.

The reviewer is responsible for verifying:

- Constitution compliance.
- ADR coverage for any new architectural decision.
- IP cleanliness per ADR 0014 (no third-party catalog references).
- Test coverage at appropriate granularity.

Welcome aboard.
