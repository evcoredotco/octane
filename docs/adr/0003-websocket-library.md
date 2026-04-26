# ADR 0003: WebSocket Library — `nhooyr.io/websocket`

- **Status:** Proposed
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Backend agent
- **Constitution principles touched:** V (Go-First, Stdlib-Heavy),
  X (Security and Compliance)

## Context

OCPP 1.6J / 2.0.1 / 2.1 transports JSON over WebSocket (WSS in production).
The Go standard library does not include a WebSocket implementation, so
OCTANE must take exactly one third-party dependency to cover this gap.

Constraints:

- Constitutional principle V allows precisely one WebSocket dependency.
- Constitutional principle X requires TLS verification on by default and
  controlled disabling via `--insecure`.
- The library must support both client (CSMS-as-server) and server
  (charging-station-as-server) modes, since OCPP scenarios include
  both directions.
- Context-aware cancellation is mandatory; the engine threads a
  `context.Context` through every operation.
- Active maintenance is required — abandoned WebSocket libraries are a
  recurring supply-chain hazard.

## Decision

Adopt **`nhooyr.io/websocket`** (now mirrored at
`github.com/coder/websocket` after Anmol Sethi joined Coder) as the single
WebSocket dependency for OCTANE. Pin to a specific minor version in
`go.mod`; bump only via an ADR amendment to this record.

## Consequences

### Positive

- Idiomatic `context.Context` support across read/write APIs.
- Minimal API surface (low cognitive load for reviewers).
- TLS handed off to `crypto/tls`; OCTANE controls verification policy.
- Active maintenance under Coder.
- Permissive license (ISC), Apache-2.0 compatible.

### Negative

- Smaller ecosystem than `gorilla/websocket`; fewer third-party examples.
- Some advanced features (per-message deflate variants, custom subprotocol
  negotiation hooks) are simpler than Gorilla's, requiring small wrappers
  in `pkg/transport/`.

### Neutral

- The dependency replaces zero stdlib functionality and is contained in
  one package (`pkg/transport/`); swapping it later affects a bounded
  blast radius.

## Alternatives considered

- **`gorilla/websocket`** — most widely used, but entered maintenance
  mode in 2022, then was revived. The maintenance signal remains noisy.
  Larger API surface and weaker context support. Rejected.
- **`github.com/gobwas/ws`** — lower-level, higher performance, but
  requires building most of the protocol layer ourselves. Rejected on
  cost-of-ownership grounds.
- **`golang.org/x/net/websocket`** — explicitly deprecated by the Go
  team. Rejected.
- **Roll our own** — RFC 6455 is well-specified, but the maintenance
  burden is incompatible with the constitution's "stdlib first" intent.
  Rejected.

## References

- Constitution: principles V, X
- nhooyr.io/websocket: https://github.com/coder/websocket
- OCPP-J transport spec: OCA OCPP-J 1.6 §3 / OCPP-J 2.0.1 §3 / OCPP-J 2.1 §3
