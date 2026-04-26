# CLAUDE.md — OCTANE

> Claude Code project memory. Loaded automatically at session start.
> See `AGENTS.md` for the cross-tool agent contract; this file adds
> Claude-specific guidance only.

## Loading order

1. `.specify/memory/constitution.md` — binding principles
2. `AGENTS.md` — agent contract (read first)
3. This file — Claude-specific overrides and shortcuts
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

The main Claude session (no subagent) may **read** anything but should
**write** only:

- `specs/<active>/` — when the user is in spec/plan/tasks mode
- `docs/adr/` — when drafting an ADR
- this file (`CLAUDE.md`) and `AGENTS.md` — only on explicit user request

For everything else, delegate to the appropriate subagent.

## Things Claude should refuse to do

- Edit `.specify/memory/constitution.md` without an ADR amendment in flight.
- Disable TLS verification anywhere in `pkg/transport/`.
- Add new third-party Go dependencies without a draft ADR.
- Mark a test case stable before it has passed against the pinned CitrineOS.
- Push directly to `main`. Always work on a `spec/...` or `feat/...` branch.

## Quick references

- OCPP 1.6J spec: https://www.openchargealliance.org/protocols/ocpp-16/
- OCPP 2.0.1 spec: https://www.openchargealliance.org/protocols/ocpp-201/
- OCPP 2.1 spec: https://www.openchargealliance.org/protocols/ocpp-21/
- CitrineOS: https://citrineos.github.io/
