# ADR 0010: Connection Profiles — User-Owned YAML

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect
- **Constitution principles touched:** X (Security), XI (Wire
  conformance), XII (No CSMS-specific adaptation)

## Context

Different CSMSes are deployed at different URLs, on different ports,
under different WebSocket subprotocols, with different authentication
modes:

- CitrineOS development default: `ws://localhost:8081/ocpp/{station_id}`,
  subprotocol `ocpp1.6` or `ocpp2.0.1`, auth optional.
- SteVe: `ws://host:port/steve/websocket/CentralSystemService/{station_id}`,
  subprotocol `ocpp1.6`.
- Production deployments: arbitrary URLs over WSS with mTLS or basic
  auth.

OCTANE needs to know how to reach a CSMS without that knowledge
being CSMS-specific behavioral adaptation. Per constitution
principle XII, behavior is OCPP and is identical across CSMSes;
**connection metadata is operator configuration**, not vendor
adaptation.

The previous design (now-removed ADRs from the deferred profile-
mechanics work) considered community-maintained `octane-profile-<csms>`
repositories carrying both connection metadata and behavioral
overrides. The behavioral-overrides part is rejected by constitution
principle XII. The connection-metadata part remains legitimate, but
does not need a code repository to express.

## Decision

A **connection profile** is a small YAML file that describes how to
reach a CSMS. It is loaded by the OCTANE runtime before any story
executes and is consumed by `pkg/transport` to open WebSocket
connections.

### Schema

```yaml
schema_version: "1"
name: citrineos
description: |
  CitrineOS development defaults. Adapt for production deployments.
ocpp_versions: ["1.6", "2.0.1"]
connection:
  url_template: "ws://{host}:{port}/ocpp/{station_id}"
  default_host: localhost
  default_port: 8081
  subprotocol_by_ocpp_version:
    "1.6":   "ocpp1.6"
    "2.0.1": "ocpp2.0.1"
auth:
  modes_supported: [none, basic, bearer]
  default_mode: none
notes: |
  CitrineOS dev port 8081 accepts unknown stations via auto-
  commissioning. This is suitable for testing only; production
  deployments will use authenticated endpoints over WSS.
```

The schema is validated by `octane validate connection <path>` and
documented at `docs/connections/schema.yaml`.

### Resolution chain

OCTANE resolves the active connection profile in this order:

1. `--connection-path <path>` flag (explicit file path).
2. `--connection <name>` flag, looked up against the search path
   below.
3. `connections.active` key in the active `octane.yml`.

The search path for `--connection <name>`:

1. `./connections/<name>.yaml` (project-local).
2. `$XDG_CONFIG_HOME/octane/connections/<name>.yaml` (user-global).
3. `/etc/octane/connections/<name>.yaml` (system-wide).
4. Connection profiles bundled with the OCTANE binary at
   `/usr/share/octane/connections/<name>.yaml` (sample profiles
   shipped via `.deb`/`.rpm`/Homebrew).

The first match wins. There is no merging across files; a connection
profile is loaded as a single document.

### Bundled samples

OCTANE ships sample connection profiles for well-known open-source
CSMSes. These are samples, not endorsements:

- `connections/citrineos.yaml` — CitrineOS dev defaults
- `connections/steve.yaml`     — SteVe dev defaults
- `connections/maeve.yaml`     — MaEVe dev defaults
- `connections/localhost.yaml` — generic local CSMS template

Adding a sample to the repository is a normal PR, gated only by
review of the YAML correctness and a `make test-connection-samples`
target that validates the file against the schema.

### What connection profiles MUST NOT contain

By constitutional rule (principle XII), connection profiles never
carry:

- Per-CSMS keyword overrides.
- Behavioral tolerances or expected deviations.
- Conformance findings to suppress.
- Test-data fixtures the CSMS is expected to ship with (such
  fixtures are story preconditions if they are conformance-relevant
  and out of scope otherwise).

A PR proposing to add any of the above is closed with a pointer to
this ADR.

### Credentials

Connection profiles never contain secrets. Credential material is
sourced exclusively from environment variables referenced by name:

```yaml
auth:
  default_mode: basic
  basic:
    username_env: OCTANE_BASIC_USER
    password_env: OCTANE_BASIC_PASS
```

This keeps credentials out of git and keeps connection profiles
shareable without redaction.

## Consequences

### Positive

- Connection metadata is data, not code. Adding support for a new
  CSMS is a YAML file, not a Go module or a separate repository.
- Constitution principle XII is operationalized cleanly: there is
  one place where CSMS-specific information legitimately lives, and
  that place is bounded and inspectable.
- Sample connection profiles in the OCTANE repo serve as both
  documentation and quick-start material.
- The resolution chain follows familiar Unix conventions (project →
  user → system) and integrates cleanly with the `octane.yml`
  resolution chain.

### Negative

- Operators with non-trivial deployments will need to adapt the
  bundled samples or write their own. Mitigated by clear schema
  documentation and `octane validate connection`.
- Connection profile schema evolution requires a migration story.
  Mitigated by `schema_version` and a documented compatibility
  policy: minor schema changes are additive and forward-compatible;
  major changes require a tool-level migration command.

### Neutral

- The schema is small enough (~30 lines) that adding a new field
  is mechanical and reviewable. The schema's conservatism is
  deliberate: anything that grows it should be challenged against
  principle XII.

## Alternatives considered

- **Connection metadata in `octane.yml` directly.** Considered.
  Rejected because it conflates CSMS identity (re-usable across
  projects) with project-specific config (story paths, fail-on
  thresholds, output dirs).
- **Vendor-maintained repositories of connection profiles.**
  Rejected: connection metadata changes too rarely to justify a
  repository per CSMS, and PRs against the OCTANE main repo are a
  more visible review surface for newcomers.
- **Auto-discovery via OCPP probe.** Rejected: discovery requires
  speaking the protocol we are testing, which creates a
  bootstrapping circularity.

## References

- Constitution: principles X, XI, XII
- ADR 0005 (story-driven framework)
- ADR 0007 (keyword library layering)
