---
sidebar_position: 4
---

# Multi-Station Testing

OCTANE stories can declare multiple station handles. This lets one run
coordinate interactions that require more than one charging station, such as
parallel transactions or prerequisite chains that apply per station.

Dependencies can be scoped per station, per run, or globally through the
story cache.

For dependency behavior, see `docs/concepts/dependency-graph.md`.

