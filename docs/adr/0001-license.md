# ADR 0001: Adopt Apache-2.0 License

- **Status:** Proposed
- **Date:** 2026-04-26
- **Deciders:** Project maintainer
- **Constitution principles touched:** I (Conformance Above Convenience),
  II (Two Distribution Surfaces)

## Context

OCTANE ships two consumer-facing artifacts: a Go CLI and a published
GitHub Action. Both are intended to be embedded in third-party CSMS
projects' CI pipelines. The license must:

- Permit commercial reuse by EV-industry vendors.
- Be compatible with the licenses of the upstream tools we run against
  (notably CitrineOS, which is Apache-2.0).
- Carry an explicit patent grant, since OCPP is a protocol space with
  active patent activity around smart-charging features.
- Be familiar to the OCA (Open Charge Alliance) ecosystem, where MIT
  and Apache-2.0 dominate.

## Decision

Adopt **Apache License, Version 2.0** for OCTANE.

## Consequences

### Positive

- Compatibility with CitrineOS and the broader OCA ecosystem.
- Explicit patent grant (Section 3 of the license).
- Permissive enough to encourage adoption by commercial CSMS vendors.

### Negative

- More verbose than MIT in source-file headers; tooling must enforce
  consistent header application.
- Slightly more restrictive than 0BSD/Unlicense for trivial reuse.

### Neutral

- Standard SPDX identifier (`Apache-2.0`) is recognized by every package
  registry and license scanner.

## Alternatives considered

- **MIT** — simpler, but lacks an explicit patent grant. Risky in the
  OCPP/EV-charging space.
- **MPL-2.0** — file-level copyleft. Discourages embedding in
  commercial CSMS code paths.
- **AGPL-3.0** — strong network copyleft. Incompatible with the goal of
  having OCTANE embedded in commercial CSMS CI pipelines.

## References

- Constitution: principles I, II
- CitrineOS license: Apache-2.0
- OCA software policy:
  <https://www.openchargealliance.org/about-us/policies/>
