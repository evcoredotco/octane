---
description: Open a new ADR with proper numbering and template.
argument-hint: "<decision title>"
allowed-tools: Bash, Read, Write, Edit
---

# /adr

You are creating an Architecture Decision Record. The title is `$ARGUMENTS`.

## Steps

1. Run:

   ```bash
   .specify/scripts/bash/new-adr.sh "$ARGUMENTS"
   ```

2. Delegate to the **architect** subagent to fill the template.
3. The architect must:
   - Capture context — what is the issue motivating this decision?
   - Capture the decision in one concise paragraph.
   - List positive, negative, and neutral consequences.
   - List at least two alternatives considered.
   - Link to the spec or constitution principle this decision serves.

## Constraints

- ADR status starts at `Proposed`. It moves to `Accepted` only after a
  maintainer review and merge.
- ADRs are immutable once `Accepted`. To revise, open a new ADR with
  `Status: Supersedes ADR XXXX` and update the old one to
  `Status: Superseded by ADR YYYY`.
- ADRs amending the constitution require the title prefix
  `Amend constitution:`.
