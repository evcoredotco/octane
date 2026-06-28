---
sidebar_position: 3
---

# Troubleshooting

Start with the command exit code and the story findings in the generated
report. For wire-level failures, inspect the trace output and compare the
observed frame to the story's `Spec-Ref`.

Common failure areas include connection profile values, TLS validation,
WebSocket subprotocol negotiation, missing story prerequisites, and CSMS
responses that do not match the OCPP frame shape.

