---
sidebar_position: 3
---

# Multi-Station Patterns

Use multiple station handles when the conformance behavior requires
independent station state or coordinated wire interactions.

Keep prerequisites explicit through `Depends` metadata. Use `per-station`
scope for setup that must run once for each station handle, and `per-run`
scope for setup shared by all station handles in a run.

For the dependency model, see `docs/concepts/dependency-graph.md`.

