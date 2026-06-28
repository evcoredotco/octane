---
sidebar_position: 2
---

# Getting Started

OCTANE runs OCPP conformance stories against a CSMS over the OCPP-J
WebSocket interface.

## Basic flow

1. Install the `octane` CLI.
2. Create or choose a connection profile for the target CSMS.
3. Run one or more `.story` files.
4. Review the generated JSON, trace, and Robot XML reports.

```bash
octane run --connection-path connections/citrineos.yaml scenarios/v16
```

For the current repository-level guide, see `docs/getting-started.md`.

