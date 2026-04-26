---
description: Decompose the active plan into atomic, agent-assignable tasks.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# /tasks

The active branch must be `spec/NNN-feature-slug` with a reviewed `spec.md`
and `plan.md`. Your job is to fill `tasks.md` such that any single task
can be executed by exactly one Claude Code subagent in one PR.

## Steps

1. Read `specs/NNN-.../spec.md` and `plan.md`.
2. Delegate to the **architect** subagent.
3. The architect must:
   - Decompose the plan into Phase 1–6 tasks (Contracts, Core, Reference,
     Surfaces, Documentation, Review).
   - Assign exactly one agent per task from the roster in `AGENTS.md`.
   - Mark dependency relationships via the Parallel column (`P` vs. `S`).
   - Reference the acceptance criteria each task covers.
   - List the *exact* files each task may touch.
   - Verify every acceptance criterion is covered by at least one task.

## Constraints

- A task that requires more than ~150 lines of diff is too large; split it.
- A task that touches multiple agents' scopes is forbidden; split it by
  scope.
- A task without an AC reference is deleted, not "kept just in case."
- Phase 6 (Review) is mandatory. Every spec ends with security review and
  code review tasks.

## Output

Run `.specify/scripts/bash/check-spec.sh specs/NNN-...` and paste its
output before finishing. Stop. The user reviews `tasks.md` before
implementation starts.
