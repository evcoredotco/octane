---
description: Validate the active spec's structural integrity.
allowed-tools: Bash, Read, Glob
---

# /check

Run the structural check against the active spec.

## Steps

1. Determine the active spec from the current branch (`spec/NNN-...` or
   `feat/NNN-...`).
2. Run:

   ```bash
   .specify/scripts/bash/check-spec.sh specs/NNN-...
   ```

3. Report the output verbatim. If checks fail, list the missing or
   malformed items and recommend the fastest fix per item.
4. Do not modify any files. The user (or a subsequent agent) decides
   whether to fix the spec, the plan, or the tasks.
