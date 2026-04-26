# ADR 0002: Go as the Engine Language

- **Status:** Proposed
- **Date:** 2026-04-26
- **Deciders:** Project maintainer
- **Constitution principles touched:** V (Go-First, Stdlib-Heavy)

## Context

OCTANE must be deployable as a single static binary in CI environments
(GitHub Actions, GitLab CI, on-premise Jenkins, Docker-in-Docker). It must
also expose a published `octane-action` GitHub Action that wraps the same
binary.

The OCPP CSMS reference implementation (CitrineOS) is written in
TypeScript/Node, but OCTANE is a *consumer* of CitrineOS, not a fork. Our
runtime requirements differ: we need fast startup, no ambient dependencies,
and easy cross-compilation for Linux/macOS/Windows.

The wider QTech ecosystem already uses Go for EVcore (the AI-native
open-core CSMS), giving us shared idioms and pre-vetted Go conventions
(see Alexis's `golang-master` skill).

## Decision

OCTANE's engine, CLI, and Action entrypoint are written in **Go 1.23**.

## Consequences

### Positive

- Single static binary; no Node, no Python, no JVM in the runtime path.
- Fast startup (< 50 ms target), critical for CI gating.
- Native cross-compilation (`GOOS=linux GOARCH=amd64`, `arm64`, etc.).
- Shared conventions with EVcore.
- Excellent stdlib coverage for HTTP, TLS, JSON, and concurrency — aligns
  with constitutional principle V.

### Negative

- Test cases are typed Go values (constitutional principle VI), not YAML.
  Contributors who prefer config-driven test definitions will face a
  ramp-up cost.
- WebSocket support requires one third-party dependency (Gorilla
  WebSocket or `nhooyr.io/websocket`). Tracked in ADR 0003.

### Neutral

- Go module path: `github.com/<org>/octane`. Final org TBD before tag v1.

## Alternatives considered

- **TypeScript/Node** — alignment with CitrineOS, but introduces a
  heavyweight runtime in CI and degrades startup latency. Rejected.
- **Rust** — excellent fit for protocol work, but ecosystem maturity for
  WebSocket + TLS + structured logging is below Go's, and the project
  cannot afford the velocity penalty at this stage. Reconsider for v2.
- **Python** — fast to prototype, but distribution and startup are weak
  for a CI-gating tool. Rejected.

## References

- Constitution: principle V
- EVcore architecture (Go + Python hybrid; engine in Go)
