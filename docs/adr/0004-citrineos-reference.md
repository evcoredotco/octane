# ADR 0004: CitrineOS as the Reference CSMS During Development

- **Status:** Proposed
- **Date:** 2026-04-26
- **Deciders:** Project maintainer
- **Constitution principles touched:** III (Reference-Validated Test Cases)

## Context

OCTANE generates conformance test cases derived from the published
OCPP specifications. To gain confidence that those test cases are
*correct* (not just runnable), each one must pass against a known-good
CSMS before being marked stable.

The reference CSMS must be:

- Open-source, so the OCTANE pipeline can build it from source.
- Actively maintained.
- Conformant to the OCPP versions OCTANE supports (1.6J, 1.6, 2.1).
- Self-contained enough to spin up via `docker-compose` in CI.
- Compatibly licensed (Apache-2.0 or similar) for downstream embedding
  in our test harness.

## Decision

Use **CitrineOS** (<https://citrineos.github.io/>) as the reference CSMS
during OCTANE's own development. Pin a specific upstream commit in
`test/reference/citrineos.version`; bumping that pin is a deliberate,
reviewed action, not an automatic update.

## Consequences

### Positive

- Apache-2.0 licensed; embeddable in our test harness without friction.
- Active development with first-party support for OCPP 1.6 and
  emerging 2.1.
- Modular TypeScript codebase with a documented configuration surface,
  making `docker-compose`-based fixtures straightforward.
- Established adoption signals from the OCA community.

### Negative

- 1.6J coverage is weaker than 1.6 in CitrineOS. OCTANE's 1.6J
  conformance tests will need a secondary reference for scenarios
  CitrineOS does not exercise. Tracked as a follow-up ADR.
- TypeScript runtime in our test rig (Docker-isolated; not a contributor
  ergonomic concern, but a CI image-size concern).

### Neutral

- The pin lives at `test/reference/citrineos.version`. Bumping requires:
  1. A successful `make test-reference` run on the new pin.
  2. A CHANGELOG entry under the next release's `## [Unreleased]`.
  3. A maintainer review of the diff (defects in upstream may shift
     OCTANE expectations).

## Alternatives considered

- **SteVe** — mature OCPP 1.6 server (Java), Apache-2.0. Strong 1.6
  coverage but no 1.6 / 2.1 path. Useful as a *secondary* 1.6 reference
  in a follow-up ADR, not as the primary.
- **MaEVe (Thoughtworks)** — Go, OCPP 1.6. Promising but smaller user
  base than CitrineOS at the time of this decision.
- **Commercial reference** — vendor CSMSes are not open-source and
  cannot be embedded in OCTANE's CI pipeline.

## References

- Constitution: principle III
- CitrineOS: <https://github.com/citrineos/citrineos-core>
- OCA OCPP versions: <https://www.openchargealliance.org/protocols/>
