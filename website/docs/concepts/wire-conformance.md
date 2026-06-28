---
sidebar_position: 1
---

# Wire Conformance

OCTANE validates observable OCPP-J behavior on the WebSocket wire. It does
not inspect CSMS databases, internal jobs, billing state, or administrative
APIs.

Each conformance story drives station-side frames and checks the CSMS
response against the published OCPP specification section named by the
story metadata.

For implementation details, see `docs/concepts/wire.md`.

