---
description: Create a new spec from a one-line problem statement.
argument-hint: "<feature description>"
allowed-tools: Bash, Read, Write, Edit, Glob, Grep
---

# /specify

You are entering the **spec authoring** phase of the OCTANE spec-driven
workflow. The user has described a feature in one or two sentences. Your job
is to scaffold the spec directory and produce a high-quality first draft of
`spec.md`.

## Steps

1. Run the scaffolding script:

   ```bash
   .specify/scripts/bash/new-spec.sh "$ARGUMENTS"
   ```

   This creates `specs/NNN-feature-slug/` with `spec.md`, `plan.md`,
   `tasks.md`, and switches to a `spec/NNN-feature-slug` branch.

2. Delegate to the **architect** subagent. The architect must:
   - Read `.specify/memory/constitution.md`.
   - Open the freshly created `spec.md` and replace placeholders.
   - Produce concrete acceptance criteria (Given/When/Then).
   - List the OCPP versions and specification sections in scope (§7).
   - Capture any unresolved decisions in §8 with owner + due date.

3. Run the structural check:

   ```bash
   .specify/scripts/bash/check-spec.sh specs/NNN-feature-slug
   ```

4. Stop. Do **not** advance to `/plan`. The user reviews `spec.md`
   before planning starts.

## Reminders

- `spec.md` describes **what** and **why**. No tech, no libraries, no
  Go types.
- If the user's description is ambiguous, ask one clarifying question
  before writing — never invent constraints.
- The branch is `spec/NNN-...` until the spec is merged; afterwards it
  becomes `feat/NNN-...` for implementation work.
