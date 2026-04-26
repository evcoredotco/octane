---
description: Execute a single task from the active tasks.md.
argument-hint: T-NNN-MM
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# /implement

You are implementing exactly one task from the active spec's `tasks.md`.
The task ID is `$ARGUMENTS`.

## Steps

1. Locate the row for `$ARGUMENTS` in `specs/NNN-.../tasks.md`. If the
   task ID is missing or ambiguous, stop and ask.
2. Read the row's `Agent` column. Delegate to that subagent. Do not
   execute the task in the main thread.
3. The chosen agent must:
   - Read the linked acceptance criteria in `spec.md`.
   - Read the contracts in `plan.md`.
   - Touch only the files listed in the task's `Files` column.
   - Run the relevant local checks (`make format`, `make lint`,
     `make test`, `make test-reference` as applicable).
   - Reference the task ID in every commit message.
4. On completion, return to the main thread with:
   - One-paragraph summary of the change.
   - List of files modified.
   - Local check results.
   - Any follow-up tasks that surfaced (do not start them — surface
     them to the architect).

## Constraints

- Never mark a task done without green local checks.
- Never expand the task scope. If you need to, stop and update
  `tasks.md` with a new row, then run the original task only.
- Never touch the constitution.
- Security-sensitive tasks (transport, action, CI, fixtures) require a
  follow-up `security` review task; ensure it exists in `tasks.md`.
