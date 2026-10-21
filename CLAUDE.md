# CLAUDE.md — OCTANE

> **What this file is:** Claude Code-specific dispatch rules, tool
> shortcuts, file-scope reminders, and domain refusals that extend the
> cross-tool agent contract in `AGENTS.md`.
>
> **What this file is not:** a substitute for `AGENTS.md`. Every rule in
> `AGENTS.md` is binding here. Where this file and `AGENTS.md`
> contradict, `AGENTS.md` wins. When in doubt about agent scope,
> responsibility, or workflow, read `AGENTS.md` first.
>
> **When NOT to use this file:** if you need the binding agent contract,
> lane definitions, or spec-driven workflow — use `AGENTS.md`. This file
> adds Claude-specific shortcuts on top, it does not replace anything.

## Loading order

1. `.specify/memory/constitution.md` — binding principles
2. `AGENTS.md` — agent contract (read first, always authoritative)
3. This file — Claude Code-specific overrides and shortcuts
4. Active spec under `specs/<current-branch>/`

## Default agent dispatch

When the user gives a high-level instruction without specifying an agent,
route as follows:

| Phrase pattern | Default agent |
|----------------|---------------|
| "draft a spec", "let's design", "what should we build" | architect |
| "implement", "code this up", "write the function" | backend |
| "add a keyword", "story DSL", "parser", "scenario file", ".story" | keyword-author |
| "add a test", "write coverage", "fuzz this" | qa |
| "wire CI", "publish the action", "container", "release" | devops |
| "review this", "look at the diff" | reviewer |
| "is this safe", "credentials", "TLS", "secrets" | security |
| "write docs", "update README", "changelog" | docs |

If the request straddles two roles, the architect mediates.

## Slash commands

| Command | What it does |
|---------|--------------|
| `/specify <text>` | Run `.specify/scripts/bash/new-spec.sh`, then draft `spec.md` from the user's description |
| `/plan` | Read the active `spec.md` and fill `plan.md` |
| `/tasks` | Read `spec.md` + `plan.md` and generate `tasks.md` rows |
| `/implement <task-id>` | Execute one task end-to-end under the assigned agent |
| `/adr <title>` | Run `.specify/scripts/bash/new-adr.sh` and draft the ADR |
| `/check` | Run `.specify/scripts/bash/check-spec.sh` against the active spec |

Definitions live under `.claude/commands/`.

## Subagent invocation hints

Claude Code spawns subagents from `.claude/agents/`. To delegate explicitly:

```
> Use the backend subagent to implement T-001-04.
> Use the security subagent to review the WebSocket TLS settings in pkg/transport.
```

When a subagent finishes, return control to the main thread with a one-line
summary; do not chain into another subagent without surfacing the
intermediate result.

## File scope reminders for the main thread

The main Claude session (no subagent) may **read** anything.

**Write targets** (main thread only — no subagent needed):

| Condition | Write target |
|-----------|-------------|
| User is in spec/plan/tasks mode | `specs/<active>/` |
| User asks to draft an ADR | `docs/adr/` |
| User explicitly requests | `CLAUDE.md`, `AGENTS.md` |

**Everything else is delegated.** If a task touches `cmd/`, `internal/`,
`pkg/`, `.github/`, `action/`, `*_test.go`, or `docs/` prose, delegate
to the appropriate subagent per the roster in `AGENTS.md -5`. Do not
write to those paths directly from the main thread.

## Things Claude should refuse to do

- Edit `.specify/memory/constitution.md` without an ADR amendment in flight.
- Disable TLS verification anywhere in `pkg/transport/`.
- Add new third-party Go dependencies without a draft ADR.
- Mark a test case stable before it has passed against the pinned CitrineOS.
- Push directly to `main`. Always work on a `spec/...` or `feat/...` branch.
- Declare any OCPP 1.6 data type (struct, enum, sub-object) locally in OCTANE.
  All OCPP 1.6 types **must** come from `github.com/evcoreco/ocpp16types`
  (ADR 0020). If the required type is missing from that module, stop and
  instruct the user to contribute it upstream first.
- Construct any OCPP 1.6 request or response message outside of the
  constructors exported by `github.com/evcoreco/ocpp16messages` (ADR 0020).
  Raw struct literals for OCPP 1.6 messages are forbidden. If the required
  message sub-package is missing, stop and instruct the user to contribute
  it upstream first.
- Parse, construct, or marshal any OCPP-J JSON message (Call, CallResult,
  CallError) outside of `github.com/evcoreco/ocpp16j` (ADR 0020).
  Specifically forbidden: calling `json.Unmarshal` directly against OCPP-J
  arrays, hand-assembling `[2,"id","Action",{...}]` literals, and building
  `UniqueId` values via raw string casts. All JSON framing **must** go through
  `ocpp16json.Parse()`, `ocpp16json.NewCall()`, `ocpp16json.NewCallResult()`,
  `ocpp16json.NewCallError()`, and `ocpp16json.NewUniqueId()`. Payload decode
  **must** use `ocpp16json.Registry` + `ocpp16json.JSONDecoder[Input, Output]`.
  If the required feature is missing from `ocpp16j`, stop and instruct the
  user to contribute it upstream first.

## EVCore OCPP 1.6 three-layer standard

> **Why this is here:** The three rules in "Things Claude should refuse"
> above all reference this standard. Every agent working in OCTANE must
> know the layer boundaries. Canonical reference: ADR 0020.

```
JSON bytes  →  ocpp16j (Parse/Marshal/Validate)  →  ocpp16messages (Req/Conf)
                                                          ↓
                                                     ocpp16types (field types, enums)
```

| Layer | Module | Scope |
|-------|--------|-------|
| JSON framing | `github.com/evcoreco/ocpp16j` | Call / CallResult / CallError envelopes, UniqueId, ErrorCode, Registry |
| Message construction | `github.com/evcoreco/ocpp16messages` | `Req()` / `Conf()` constructor pairs per OCPP action |
| Data types | `github.com/evcoreco/ocpp16types` | CiString, DateTime, enums, sub-objects |

All three are first-party EVCore infrastructure and must be pinned as direct
`go.mod` dependencies (never `// indirect`).

## Quick references

- OCPP 1.6J spec: https://www.openchargealliance.org/protocols/ocpp-16/
- - - CitrineOS: https://citrineos.github.io/
