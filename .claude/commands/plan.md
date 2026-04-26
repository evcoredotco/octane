---
description: Translate the active spec into a technical plan.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# /plan

The active branch must be `spec/NNN-feature-slug` with a merged-or-reviewed
`spec.md`. Your job is to fill `plan.md` so that the implementation phase
can begin.

## Steps

1. Identify the active spec from the current branch name.
2. Read `specs/NNN-.../spec.md` end to end.
3. Delegate to the **architect** subagent.
4. The architect must:
   - Identify every architecture touchpoint and tick it in §2.
   - Enumerate every public API change and label its semver impact (§3).
   - Define data contracts (OCPP message shapes, report schema additions)
     and link them to the OCPP version-specific schema (§4).
   - List required ADRs in §5; if any do not yet exist, run
     `.specify/scripts/bash/new-adr.sh "<title>"` to scaffold them.
   - Map every acceptance criterion in `spec.md` to a test strategy
     (unit / integration vs. CitrineOS / fuzz / determinism) in §6.
   - Capture risks, mitigations, rough effort (§7–§9).

## Constraints

- The plan must not introduce requirements absent from `spec.md`. If new
  requirements emerge, stop and amend the spec first.
- Every ADR draft is a real file under `docs/adr/` — not a placeholder
  bullet.
- Stop at the end of `plan.md`. Do **not** advance to `/tasks`.
