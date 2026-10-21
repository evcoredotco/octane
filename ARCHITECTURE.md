# OCTANE — Architecture

> **What this document is:** a design narrative that consolidates the
> ratified ADRs and the project principles into a single readable
> reference. It is a *summary*, not a specification — where this
> document and an ADR disagree, the ADR wins.
>
> **What this document is not:** a specification, a test plan, or an
> implementation guide. For binding rules read the principles in
> section 2 and the individual ADRs in `docs/adr/`. For agent operating
> rules read `AGENTS.md`.
>
> **Audience:** engineers, certification reviewers, and contributors.
>
> **Scope:** OCTANE v0 (pre-alpha). Decisions tracked through ADR 0020.
> **Last reviewed:** 2026-10-21

---

## 1. What OCTANE is

**In one sentence:** OCTANE drives an unmodified CSMS over OCPP-J 1.6
and validates wire-level protocol conformance against the published
specification — without requiring any changes to the CSMS under test.

OCTANE — *OCPP Conformance Testing & Network Evaluation* — is an
open-source conformance harness for OCPP 1.6J
Charging Station Management Systems. It runs against an unmodified
CSMS and verifies wire-level conformance to the publicly published
OCPP specifications, automated and CI-friendly.

OCTANE ships as a single Go binary distributed via `.deb`, `.rpm`,
Homebrew, Scoop, Docker, and direct download. A companion GitHub
Action wraps the same binary so any CSMS project can gate its CI
pipeline on conformance with a single workflow step.

The published artefact set is small and stable:

- `octane` — the CLI binary
- `octane-action` — the GitHub Action wrapper
- Sample connection profiles bundled with the binary at
  `/usr/share/octane/connections/<n>.yaml` (per ADR 0010)

## 2. Foundational principles

OCTANE's constitution is the only set of rules that overrides the rest.
Every ADR and design decision must comply with it. The full constitution
text is not kept in this repository; the table below is the reference
for the principles that ADRs, code comments, and other documents cite as
"constitution principle N".

| #    | Principle                                               | Operational meaning                                                                                 |
|------|---------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
| I    | Conformance Above Convenience                           | Every conformance test traces to a section of the OCPP specification.                               |
| II   | Two Distribution Surfaces, One Engine                   | CLI and Action expose the same engine; no surface-only features.                                    |
| III  | Reference-Validated                                     | Every test passes against a pinned CitrineOS commit before being marked stable.                     |
| IV   | Determinism and Reproducibility                         | Same inputs, same outputs (modulo timestamps). Seeded RNG, injected clock, sorted serializers.      |
| V    | Go-First, Stdlib-Heavy                                  | Go 1.26. Third-party deps require an ADR. WebSocket is the single allowed exception.                |
| VI   | Test Cases Are Code, Not Configuration                  | Stories are a declarative *surface*; the keyword library that executes them is typed Go.            |
| VII  | Public API Stability                                    | `pkg/` follows semver. `internal/` does not.                                                        |
| VIII | Spec-Driven Development                                 | Architectural decisions are recorded as ADRs before the code lands.                                 |
| IX   | AI Agents Are Bounded Contributors                      | Agents have declared scopes; agent commits get human sign-off.                                      |
| X    | Security and Compliance Are Continuous                  | TLS on by default, secrets never in artefacts, dependencies scanned on every PR.                    |
| XI   | Conformance Is Verified On the Wire                     | OCTANE never demands changes to the CSMS, never uses admin APIs, never depends on a vendor adapter. |
| XII  | Scenarios Are Declarative; Adaptation Lives in Profiles | Stories carry no CSMS-specific assumptions; profiles carry no conformance logic.                    |

Principles XI and XII are the two pivots that distinguish the current
design from earlier drafts. They emerged from a deliberate rejection
of an adapter-based model (see section 13) and align OCTANE with a
**zero-cooperation-cost** adoption philosophy: any CSMS team can
run OCTANE against an unmodified deployment with one CLI command.

## 2.1 What OCTANE does NOT do

These non-responsibilities are architectural constants, not temporary
limitations. They will not change without an ADR that amends the
principles in section 2.

| OCTANE does NOT… | Responsibility delegated to |
|------------------|-----------------------------|
| Modify CSMS state or call CSMS admin APIs | Ruled out by principle XI; OCTANE is read-only on the wire |
| Accept per-CSMS keyword overrides or behavioral tolerance layers | Ruled out by principle XII; adaptation belongs in connection profiles |
| Guarantee distributed cross-machine cache coherence | Delegated to CI infrastructure (cache artifacts, shared runners) |
| Validate CSMS-internal business logic beyond the wire spec | Out of scope; OCTANE tests OCPP-J protocol conformance only |
| Require any deployment changes on the CSMS side | Zero-cooperation-cost adoption is a hard constraint, not a goal |
| Encode OCPP payload types or message constructors locally | Delegated to `github.com/evcoreco/ocpp16types` and `ocpp16messages` (ADR 0020) |
| Parse or frame OCPP-J JSON arrays inline | Delegated to `github.com/evcoreco/ocpp16j` (ADR 0020) |
| Implement OCPP 2.0.1 or 2.1 | Out of scope; OCTANE targets OCPP 1.6J only |

## 3. The three-layer model

```text
┌──────────────────────────────────────────────────────────────────┐
│  Layer 1 — Stories (.story files)                                │
│  ──────────────────────────────────────────────────────────────  │
│  Declarative Gherkin-flavored scenarios. One per OCPP section/   │
│  protocol behavior under test.                                   │
│  Live in scenarios/v16/.                                         │
│  Read by certification reviewers. Version-controlled. ADR 0006.  │
└──────────────────────────────────────────────────────────────────┘
                              ▼  resolves keywords against
┌──────────────────────────────────────────────────────────────────┐
│  Layer 2 — Keyword library (Go code)                             │
│  ──────────────────────────────────────────────────────────────  │
│  Two sub-layers per ADR 0007, resolved domain → primitive:       │
│                                                                  │
│  • Domain      — OCPP-version-scoped semantics                   │
│  • Primitive   — transport-level escape hatch                    │
│                                                                  │
│  No per-CSMS override layer (constitution principle XII).        │
└──────────────────────────────────────────────────────────────────┘
                              ▼  drives the wire via
┌──────────────────────────────────────────────────────────────────┐
│  Layer 3 — Engine (pkg/engine, pkg/transport, pkg/wire)          │
│  ──────────────────────────────────────────────────────────────  │
│  WebSocket transport, OCPP-J frame parser, multi-station         │
│  orchestration, deterministic clock and RNG, report builder.     │
└──────────────────────────────────────────────────────────────────┘
                              ▼  emits
┌──────────────────────────────────────────────────────────────────┐
│  Reports                                                         │
│  ──────────────────────────────────────────────────────────────  │
│  • octane.json     — OCTANE-native, byte-deterministic           │
│  • output.xml      — Robot Framework 7.x format (ADR 0009)       │
└──────────────────────────────────────────────────────────────────┘
```

Each layer is the contract for the layer above it. Stories never
import Go; keywords never know which CSMS they are talking to;
engine never knows which OCPP version it is dealing with. This
separation is what makes the architecture composable.

## 4. The story DSL

A `.story` file declares one OCPP scenario in three required sections
(`Meta`, `Background`, `Scenario`) plus optional `Setup` and `Teardown`.
Indentation is whitespace-significant; tabs are forbidden; the parser
is recursive-descent with no third-party dependency.

```text
Meta
    Name:        Connector reservation faulted
    Id:          connector_reservation_faulted
    Spec-Ref:    OCPP-J 1.6 -6.40 ReserveNow
    Tags:        reservation, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId, idTag
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS handles a Faulted reservation response
    When  the CSMS sends ReserveNow with connectorId {connectorId}
          and idTag "{idTag}" to station "CP01" within 30 seconds
    Then  station "CP01" responds with ReserveNow.conf status "Faulted"
    And   the CSMS accepts the response without error within 10 seconds

Teardown
    Disconnect station "CP01"
```

Required Meta keys carry OCPP specification traceability (`Spec-Ref`,
required for conformance tests), declare station count for preflight
resource allocation (`Stations`), and classify the scenario through
`Tags` (one of `wire-only`, `multi-station`, `operator-assisted`,
`helper`; optionally `pure-protocol` to declare independence from
any fixture).

The `Parameters` key declares story inputs. They resolve from the
`parameters` map in `octane.yml` and can be overridden per run with
`octane run --param name=value`. Without them, stories cannot vary by
deployment. The schema is implemented but not yet ratified into an ADR.

The grammar is small enough to fit a recursive-descent parser in
~600–800 lines of Go (target). Diagnostic quality (line + column
\+ Levenshtein-suggested keyword on resolution failure) is where most
of the engineering goes.

Defined in **ADR 0006**.

## 5. The keyword library

Keywords are typed Go functions that map step text to wire actions.
They self-register at process start via `init()`. Pattern collisions
inside the same `(layer, ocpp-version)` tuple panic at startup —
caught in CI, never at runtime.

### Two layers, deterministic resolution

```yml
Step text:
    "the CSMS sends ReserveNow with connectorId 1
     and idTag "VID:0001" to station "CP01" within 30 seconds"

Resolution order at runtime:
    1. Domain keywords scoped to the story's OCPP  → first match wins
    2. Primitive keywords                          → fallback
```

A step that does not match either layer fails preflight (exit 4) with
a diagnostic that lists the layers searched and the closest registered
patterns by Levenshtein distance ≤ 4.

There is no third "profile" layer. Per constitution principle XII,
domain keywords are identical for every CSMS implementing a given OCPP
version. A CSMS deviation is a finding, not a configurable tolerance.

### Keyword shape

A keyword is a `Pattern + Func` pair registered against a layer:

```go
registry.Register(api.Keyword{
    Layer:       api.LayerDomain,
    OCPPVersion: api.OCPP16,
    Pattern: "the CSMS sends ReserveNow with " +
        "connectorId {connectorId:int} and " +
        "idTag {idTag:string} to station {station:string} " +
        "within {timeout:duration}",
    Func: expectReserveNow,
})
```

Placeholders use `{name:type}` syntax. Supported types: `string`,
`int`, `float`, `bool`, `duration`, `station`, `any`. Each keyword
registers exactly one pattern; the registry rejects collisions with
a startup panic.

### Robot Framework parallel

The keyword library is the architectural equivalent of a Robot
Framework Python library:

| Robot Framework                        | OCTANE                                                           |
|----------------------------------------|------------------------------------------------------------------|
| `@keyword("Pattern ${arg}")` decorator | `registry.Register(api.Keyword{Pattern: "..."})`                 |
| Library file (single layer)            | `LayerDomain` for OCPP semantics, `LayerPrimitive` for transport |
| `types={"count": int}`                 | `{name:int}` placeholders                                        |
| `raise AssertionError`                 | `return fmt.Errorf(...)`                                         |
| `from robot.api import logger`         | `state.Logf(...)`                                                |
| Per-test scope                         | `api.State` per-scenario by construction                         |

OCTANE does not use Robot Framework at runtime, but borrows its
mental model and emits its `output.xml` format (see section 11).

Defined in **ADR 0007**.

## 6. Shared OCPP 1.6 data types and messages

Two first-party EVCore modules are the **exclusive** source of all OCPP 1.6
data structures in OCTANE. Neither module carries external transitive
dependencies.

```text
github.com/evcoreco/ocpp16types     — field types, enumerations, sub-objects
github.com/evcoreco/ocpp16messages  — request / response message constructors
```

These are **hard architectural rules**, not conventions. Every EVCore
application that speaks OCPP 1.6 operates on identical Go types and
identical constructor paths — same field names, same JSON tags, same enum
values, same validation logic — so that a protocol-level discrepancy can
never hide behind a local type variant or a hand-rolled struct literal.

### `ocpp16types` — primitive data types

Use for field-level constrained types, enumerations, and composite
sub-objects whenever no message construction is involved.

| Category                      | Examples                                                                                                                                                                                |
|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Field-level constrained types | `CiString20Type`, `CiString50Type`, `CiString255Type`, `CiString500Type`, `DateTime`, `Integer`                                                                                         |
| Enumerations                  | `AuthorizationStatus`, `ChargePointStatus`, `RegistrationStatus`, `DiagnosticsStatus`, `ChargingProfileKindType`, `ChargingRateUnit`, `Measurand`, `Phase`, `Location`, `UnitOfMeasure` |
| Composite sub-objects         | `IdTagInfo`, `IDToken`, `MeterValue`, `SampledValue`, `ChargingSchedule`, `ChargingSchedulePeriod`, `ChargingProfile`, `AuthorizationData`, `KeyValue`                                  |
| Error sentinels               | `ErrEmptyValue`, `ErrInvalidValue`                                                                                                                                                      |

```go
import ocpp16 "github.com/evcoreco/ocpp16types"

// Usage in keyword functions:
//   ocpp16.AuthorizationStatus
//   ocpp16.NewIDTagInfo(ocpp16.AuthorizationStatusAccepted)
//   ocpp16.RegistrationStatus
```

### `ocpp16messages` — message construction

Use for building every OCPP 1.6 request and confirmation frame. Each
message lives in its own sub-package and exports a `Req(ReqInput)` /
`Conf(ConfInput)` constructor pair. Constructors validate all fields and
return a joined error for every violation — catching malformed messages at
construction time, not on the wire.

```go
import "github.com/evcoreco/ocpp16messages/bootnotification"
import "github.com/evcoreco/ocpp16messages/authorize"
import "github.com/evcoreco/ocpp16messages/starttransaction"
// ... one import per message sub-package as needed

// Building a request:
req, err := bootnotification.Req(bootnotification.ReqInput{
    ChargePointModel:  "Model-X",
    ChargePointVendor: "VendorCo",
})

// Building a confirmation:
conf, err := bootnotification.Conf(bootnotification.ConfInput{
    CurrentTime: ocpp16.DateTime(time.Now()),
    Interval:    300,
    Status:      ocpp16.RegistrationStatusAccepted,
})
```

`ocpp16messages` re-exports the `ocpp16types` types package internally. A
file that uses both message constructors and field types may reach the types
through that re-export, but a direct `ocpp16types` import is preferred for
files that only need enumerations or field types.

### What is forbidden

- Local re-declarations of any OCPP 1.6 type, even in test packages.
- Raw struct literals that construct request or response message bodies
  without using the `Req` / `Conf` constructors from `ocpp16messages`.
- Type aliases or shadow copies of spec-defined types anywhere in OCTANE.

When either module does not yet expose a needed type or message, the correct
action is to contribute it upstream first, then unblock the OCTANE task once
the release is tagged. Reviewers reject any PR that violates either rule.

**Status:** not yet adopted. None of these modules is in `go.mod` yet, and
no OCTANE package imports them.

Defined in **ADR 0020**.

## 7. Connection profiles

OCTANE has no CSMS-specific behavioral adaptation surface
(constitution principle XII). What is legitimately CSMS-specific is
**connection metadata** — how to reach the CSMS at the network
level. This lives in small YAML files called connection profiles
(ADR 0010), owned by the operator, not the vendor.

A connection profile is roughly 30 lines:

```yaml
schema_version: "1"
name: citrineos
ocpp_versions: ["1.6", "1.6"]
connection:
  url_template: "ws://{host}:{port}/ocpp/{station_id}"
  default_host: localhost
  default_port: 8081
  subprotocol_by_ocpp_version:
    "1.6":   "ocpp1.6"
    "1.6": "ocpp1.6"
auth:
  modes_supported: [none, basic, bearer]
  default_mode: none
```

The OCTANE binary ships sample connection profiles for well-known
open-source CSMSes (CitrineOS, SteVe, MaEVe). Operators with
non-trivial deployments adapt the samples or write their own. A
connection profile is resolved through a chain (project → user-global
→ system → bundled samples); the first match wins.

Connection profiles **never** contain keyword overrides, behavioral
tolerances, expected deviations, or anything that could let a
non-conformant CSMS pass a conformance test. They are network
metadata, full stop.

Defined in **ADR 0010**.

## 8. Test dependency graph

OCPP scenarios are not independent: a reservation test cannot run
against a CSMS that has not registered the station, and a station
cannot register without first establishing an OCPP-J WebSocket
connection. Most scenarios require a chain of prior state to reach
the point where the actual conformance assertion can be exercised.

OCTANE models the test suite as a **directed acyclic graph of test
cases**. Every story can be a prerequisite for another via the
`Depends:` Meta key:

```yml
Meta
    Name:     Connector reservation faulted
    Id:       connector_reservation_faulted
    Spec-Ref: OCPP-J 1.6 -6.40 ReserveNow
    Stations: 1
    Depends:
      - id:    connector_status_available
        scope: per-station
```

The runner walks the dependency chain transitively, executes the
prerequisites in topological order, then executes the requested
story. Each prerequisite is itself a story that may have its own
dependencies. For the example above, the resolved chain is:

```text
connector_reservation_faulted
  └── connector_status_available
        └── station_boot_accepted
              └── station_connection_established
```

### Helper stories vs conformance stories

A conformance story carries `Spec-Ref` (citing an OCPP specification
section) and asserts conformance to that section. A helper story
omits `Spec-Ref` and is tagged `helper`; helpers exist purely as
dependency targets that bring the system to a known state.

The parser enforces the distinction: a story tagged `helper` MUST
omit `Spec-Ref`; a story not tagged `helper` MUST include it. Both
live alongside each other under `scenarios/`; there is no separate
`helpers/` directory.

### Failure propagation

When a prerequisite fails, dependent stories are **skipped**, not
failed. The report distinguishes:

| Status    | Meaning                                          |
|-----------|--------------------------------------------------|
| `passed`  | Story ran and all assertions held.               |
| `failed`  | Story ran and at least one assertion failed.     |
| `skipped` | Story did not run because a prerequisite failed. |

A skipped story carries a pointer to the failing prerequisite in its
report entry, so an operator can see the chain rather than the
cascade.

### Multi-station scoping

A `per-station` prerequisite (the default) runs once per station
handle. So a story declaring `Stations: 2` causes
`station_boot_accepted` to run twice, once for `CP01` and once for
`CP02`. A `per-run` prerequisite runs once regardless of station
count. A `global` prerequisite runs once across the validity window
defined by the cache (see section 9).

Defined in **ADR 0015**.

## 9. Cache and locks

To make the dependency graph efficient, OCTANE caches test results
across `octane run` invocations. The cache is a **content-addressed
file tree** at `$XDG_CACHE_HOME/octane/cache/` (overridable via
`OCTANE_CACHE_DIR`). Each cache entry is two JSON files: a small
`result.json` and an optional sibling `trace.json` with the wire
frames.

```text
$XDG_CACHE_HOME/octane/cache/
├── results/
│   ├── ab/ab12cd34.../
│   │   ├── result.json    # status, timing, findings (1–2 KB)
│   │   └── trace.json     # OCPP-J wire frames (0–MBs)
│   └── ...
├── locks/                 # POSIX flock targets, one per cache key
└── meta/                  # plain-text version stamp
```

The two-character fanout matches the convention used by Bazel,
ccache, and Go's build cache — bounding directory size and keeping
inspection trivial.

### Cache key

Each cache entry is keyed by the SHA-256 of the tuple:

| Field               | Source                                                                     |
|---------------------|----------------------------------------------------------------------------|
| `test_id`           | story `Id` Meta key                                                        |
| `scope_key`         | station handle for `per-station`, run ID for `per-run`, empty for `global` |
| `csms_endpoint_sha` | SHA-256 of (URL + subprotocol + auth-mode)                                 |
| `octane_version`    | from build info                                                            |
| `ocpp_version`      | from story Meta or config                                                  |
| `story_content_sha` | SHA-256 of story file + transitively all prerequisites' content            |
| `parameter_sha`     | SHA-256 of bound parameters                                                |

A cached result is valid only if all fields match the current
invocation. Editing any file in the dependency chain, upgrading
OCTANE, switching CSMS endpoints, or running with different
parameters all invalidate cleanly — the new key resolves to a new
path; old paths become unreachable and are pruned by age.

A per-test `Cache-TTL: <duration>` Meta key adds time-based
invalidation. Defaults: `1h` for helpers, `infinite` for conformance
tests.

### Wire trace splitting

Wire traces are written as a sibling `trace.json` next to
`result.json`. Reports that only need pass/fail status read
`result.json` (fast); reports that need wire-level detail read
both. Failed tests always write traces; passing tests write traces
by default, suppressible with `--no-trace-on-pass` to reduce CI
cache size.

### Atomic writes

Every cache write uses temp-file-and-rename:

1. Write `result.json.tmp` (and/or `trace.json.tmp`) into the
   target directory.
2. `fsync` the temp file.
3. `rename` to its final name (atomic on POSIX).
4. `fsync` the directory entry.

A reader sees either the prior version or the new version, never a
torn write.

### Lock protocol

Cross-process safety on a single machine uses POSIX advisory locks
on `locks/<key_hash>.lock` files separate from the result files:

- **In-process:** `sync.Map[CacheKey]*sync.Once` per cache key.
- **Linux/macOS:** `syscall.Flock(LOCK_EX)`.
- **Windows:** `LockFileEx`.

The acquire pattern is double-checked: read result, take exclusive
flock, re-read result, execute if still missing, write atomically,
release lock. The double-read handles the case where another local
runner completed the test while we waited.

Default `--lock-timeout` is 60 seconds. `--no-wait` fails fast on
lock contention.

**Cross-machine concurrency is delegated to the CI cache layer.**
OCTANE is not a distributed system; coordinating concurrent runs
across machines that do not share a filesystem is exactly what
GitHub Actions and GitLab CI's cache infrastructure does well.

### CI cache integration

The file-tree layout is designed for CI cache restoration. Two
parallel CI jobs running disjoint partitions of a suite write to
disjoint key-hash prefixes; their cache directories merge cleanly
when the next job restores both. Partial restoration leaves a
usable subset (entries that fully restored are valid; missing
entries simply re-run).

Example GitHub Actions integration:

```yaml
- uses: actions/cache@v4
  with:
    path: ${{ env.OCTANE_CACHE_DIR }}
    key: octane-${{ runner.os }}-${{ hashFiles('scenarios/**') }}
    restore-keys: |
      octane-${{ runner.os }}-
```

GitLab CI is analogous; full examples live in ADR 0016.

### Operator commands

```shell
octane cache info     # cache directory, total size, entry count
octane cache prune    # remove entries older than --max-age (default 30d)
octane cache clear    # remove all cache content
octane cache key <story> [--scope STATION]   # print resolved key
octane cache show <key>                       # cat result.json
octane cache trace <key>                      # cat trace.json
```

Inspection is `cat | jq`. No SQL, no third-party tool to install.

Defined in **ADR 0016**.

## 10. Multi-station orchestration

Many stateful OCPP scenarios are reachable on the wire by coordinating
two or more simulated stations. A story declares the count in `Meta`:

```yml
Meta
    Spec-Ref:    OCPP-J 1.6 -6.2 Authorize
    Stations:    2
```

The runner allocates handles `"CP01"`, `"CP02"`, … on preflight.
Steps reference stations by handle. Steps run sequentially in
declared order; concurrency is opt-in via `Parallel ... End-Parallel`
blocks:

```yml
Parallel
    When  station "CP01" sends StartTransaction
    When  station "CP02" sends StartTransaction
End-Parallel
```

Per-station state is captured in the report; per-step latencies and
wire traces are recorded for every station. Determinism (principle IV)
applies modulo concurrency: two parallel sends may reach the CSMS in
either order, and the report records both observed orders.

Defined in **ADR 0008**.

## 11. Reports

Every run produces two artefacts:

| File          | Format                                 | Purpose                                                                     |
|---------------|----------------------------------------|-----------------------------------------------------------------------------|
| `octane.json` | OCTANE-native JSON, byte-deterministic | Source of truth for certification                                           |
| `output.xml`  | Robot Framework 7.x output schema      | Ecosystem reporting (Allure, ReportPortal, Jenkins, GitLab, GitHub Actions) |

Both files are produced from the same in-memory report tree. They
agree by construction.

The native JSON carries the OCTANE version, OCPP version, profile
identity and version, the SHA-256 of the configuration, the seed,
and a per-scenario tree of steps with wire frames attached. Identical
inputs produce byte-identical reports modulo timestamps.

The Robot XML emitter pins to the Robot Framework 7.x schema. A
golden test asserts the produced XML matches a fixed file; a CI step
pipes the XML through Robot's own `rebot` tool to verify upstream
consumability.

Defined in **ADR 0009**.

## 12. Distribution and operator surface

### CLI

The canonical invocation is:

```shell
octane run --csms-endpoint ws://localhost:9210                        # stories in storiesDir
octane run scenarios/v16/station_boot_accepted.story --csms-endpoint ws://localhost:9210
octane validate stories scenarios/v16                                 # no CSMS connection
```

Configuration resolves in this order, each level overriding the
previous one:

1. Built-in defaults
2. `octane.yml` (path set with `--config`; default `./octane.yml`)
3. `OCTANE_*` environment variables
4. Command-line flags

The CSMS endpoint is not stored in `octane.yml`; pass it with
`--csms-endpoint`. The checked-in `octane.yml`:

```yaml
storiesDir:  scenarios/v16
ocppVersion: "1.6"
maxParallel: 1
parameters:
  connectorId: "1"
  valid_idTag: "AABBCC"
  connectionTimeOut: "30s"
  meterStart: "0"
  meterStop: "1000"
```

The full field reference is in `docs/configuration.md`.

### Distribution channels

`.goreleaser.yaml` and `packaging/nfpm.yaml` configure these release
artefacts:

| Channel          | Artefact                                                  |
|------------------|-----------------------------------------------------------|
| Debian/Ubuntu    | `.deb` package (nfpm)                                     |
| Fedora/RHEL      | `.rpm` package (nfpm)                                     |
| macOS / Linux    | Homebrew formula in `evcoredotco/homebrew-octane` (tap not created yet) |
| Docker           | `ghcr.io/evcoredotco/octane` images for amd64 and arm64   |
| Direct           | archives with SBOMs and cosign-signed checksums           |
| Go               | `go install github.com/evcoredotco/octane/cmd/octane@latest` |

Releases are published from `vX.Y.Z` tags, starting with `v0.1.0`.
Release artefacts include man pages, shell completions, and the
LICENSE.

### Documentation

OCTANE ships professional Unix-style documentation (ADR 0011):

| Section                                 | Content                  | Tool                   |
|-----------------------------------------|--------------------------|------------------------|
| `man 1 octane`, `man 1 octane-run`, ... | Per-subcommand reference | Generated from cobra   |
| `man 5 octane.yml`                      | Config file reference    | Hand-written via scdoc |
| `man 5 octane.story`                    | Story DSL reference      | Hand-written via scdoc |
| `man 7 octane`                          | Concepts overview        | Hand-written via scdoc |

Shell completion (ADR 0012) ships for bash and zsh, with both static
(subcommands, flags) and dynamic completion (story file paths,
profile names, registered keywords). Dynamic completion functions
are bound by a hard rule: read-only and side-effect-free. No tool
enforces the rule yet; reviewers check it.

A separate Docusaurus website in `website/` (ADR 0013), deployed to
GitHub Pages by `.github/workflows/docs.yml`, covers tutorials,
conceptual narrative, and the keyword catalog.
It is **not** an HTML render of the man pages — different reading
modes deserve different content shapes. Three of its reference pages
are mechanically generated (CLI, config schema, keyword catalog); the
rest is hand-written.

## 13. Design history

OCTANE's current shape emerged from a sequence of design decisions
that explicitly rejected some intuitive but problematic alternatives.
The rejections are worth recording because they illuminate why the
current architecture is what it is.

### What was considered and rejected

**Vendor-implemented test harness adapters.** Early design proposed
that the CSMS team write a sidecar service exposing an OpenAPI
contract that OCTANE could call to set up state and verify outcomes.
Rejected on adoption-cost grounds: a conformance tool that requires
vendor cooperation is not, in practice, a conformance tool.

**Per-CSMS keyword overrides.** Subsequent design proposed a third
keyword resolution layer where CSMS profiles could override domain
keywords for behavioral quirks. Rejected on integrity grounds (now
codified as constitution principle XII): a CSMS that requires special
handling to pass a conformance test is, by definition, not conformant.
Domain keywords are identical across CSMSes.

**Operator escape hatches for known deviations.** Considered as a
softer version of the previous: `--accept-deviation`, severity
overrides, curated known-deviations registries. Rejected for the same
reason. The integrity of the conformance signal is non-negotiable.

**Reuse of Robot Framework as a runtime.** Rejected on distribution
grounds. A Python interpreter in CI is heavier than a single static
binary; the value of Robot's structural metaphor does not require
its runtime. OCTANE emits Robot's `output.xml` format (ADR 0009) for
ecosystem compatibility while keeping the core in Go.

### What survived

- **The wire-only conformance model**, codified in constitution
  principle XI.
- **The Robot Framework structural metaphor** — declarative scenarios,
  layered keyword library, machine-readable reports — without the
  Python runtime.
- **Connection metadata as legitimate operator configuration**, not
  vendor adaptation (ADR 0010).
- **Multi-station orchestration as a first-class feature** (ADR 0008),
  recovering the stateful-scenario coverage that adapter-based
  designs were trying to address.

## 14. Repository layout

```text
octane/
├── .github/workflows/               # ci.yml, docs.yml, release.yml
├── action/                          # GitHub Action wrapper (Docker-based)
├── cmd/octane/                      # Cobra CLI, configuration, exit codes
├── connections/                     # sample connection profile (citrineos.yaml)
├── docs/
│   ├── adr/                         # architecture decision records
│   ├── concepts/                    # concept documentation
│   ├── integrations/                # CI and Robot Framework integration guides
│   ├── keywords/                    # keyword reference
│   ├── man/                         # scdoc sources for man sections 5 and 7
│   └── conformance-claim.md         # public conformance scope statement
├── examples/                        # consumer CI snippets and example stories
├── packaging/                       # nfpm.yaml for .deb/.rpm
├── pkg/
│   ├── cache/                       # content-addressed result cache and locks
│   ├── engine/                      # deterministic clock and rand
│   ├── keywords/                    # api, registry, primitive, lifecycle, ocpp16
│   ├── report/                      # report model, JSON and Robot XML writers
│   ├── runner/                      # dependency graph, sharding, execution
│   ├── story/                       # .story lexer, parser, AST, diagnostics
│   ├── transport/                   # WebSocket dialing and station connections
│   └── wire/                        # OCPP-J frame parsing and encoding
├── scenarios/v16/                   # OCPP 1.6 conformance and helper stories
├── scripts/                         # man page, completion and CI helper scripts
├── test/
│   ├── integration/                 # integration tests
│   └── reference/                   # Docker Compose CitrineOS reference target
├── website/                         # Docusaurus site
├── AGENTS.md                        # cross-tool agent contract
├── ARCHITECTURE.md                  # this document
├── CHANGELOG.md
├── CLAUDE.md                        # Claude Code project rules
├── CONTRIBUTING.md                  # author / contributor guide
├── Dockerfile                       # release image (binary supplied by goreleaser)
├── LICENSE                          # Apache-2.0
├── Makefile                         # format, lint, test, build, docs, packaging
├── golangci.yml                     # lint and formatter configuration
├── octane.yml                       # starter configuration
└── .goreleaser.yaml                 # release orchestration
```

## 15. ADR index

The ADRs are numbered 0001–0020. Number 0017 is not used. ADRs
0001–0004 have status Proposed; the others are Accepted.

| #    | Title                                                                          |
|------|--------------------------------------------------------------------------------|
| 0001 | Adopt Apache-2.0 License                                                       |
| 0002 | Go as the Engine Language                                                      |
| 0003 | WebSocket Library (`nhooyr.io/websocket`, now `github.com/coder/websocket`)     |
| 0004 | CitrineOS as the Reference CSMS                                                |
| 0005 | Story-Driven Conformance Framework                                             |
| 0006 | `.story` Gherkin-Flavored DSL                                                  |
| 0007 | Keyword Library Layering — Primitive and Domain                                |
| 0008 | Multi-Station Orchestration                                                    |
| 0009 | Robot Framework `output.xml` Compatibility                                     |
| 0010 | Connection Profiles — User-Owned YAML                                          |
| 0011 | Manual Pages (Cobra + scdoc)                                                   |
| 0012 | Shell Completion (bash + zsh, dynamic)                                         |
| 0013 | Web Documentation (Docusaurus)                                                 |
| 0014 | Intellectual Property and Authoring Guidelines                                 |
| 0015 | Test Dependency Graph                                                          |
| 0016 | Cache and Lock Subsystem — Content-Addressed File Tree                         |
| 0018 | Determinism Primitives (Clock and Rand interfaces)                             |
| 0019 | Runner Concurrency Model                                                       |
| 0020 | Shared OCPP 1.6 Data Types and Messages via `ocpp16types` and `ocpp16messages` |

Decisions that are implemented but not yet recorded as ADRs:

| Title                                | Notes                                                    |
|--------------------------------------|----------------------------------------------------------|
| CLI Surface and Subcommand Structure | `octane run`, exit codes, global flags                   |
| Configuration Resolution and Schema  | defaults, `octane.yml`, `OCTANE_*`, flags                |
| Distribution Channels                | `.deb`, `.rpm`, Homebrew, Docker, signing                |
| Story Parameters                     | `Parameters` Meta key, `parameters` map, `--param` flag  |

## 16. What is implemented today

| Component                                                              | Status                                         |
|------------------------------------------------------------------------|------------------------------------------------|
| Story lexer, parser and validator (`pkg/story`)                        | Done                                           |
| Keyword API, registry, primitive, lifecycle and OCPP 1.6 keywords      | Done                                           |
| Runner: dependency graph, sharding, parallelism, skip logic            | Done                                           |
| Content-addressed cache and file locks                                 | Done                                           |
| Reports: `octane.json` and Robot Framework `output.xml`                | Done                                           |
| CLI: `run`, `validate`, `keywords`, `cache`, `completion`              | Done                                           |
| Man pages and bash/zsh completion                                      | Done                                           |
| OCPP 1.6 stories in `scenarios/v16/`                                   | 23 stories                                     |
| CI: lint, tests, CodeQL, website deployment                            | Done                                           |
| Releases (`.goreleaser.yaml`, `packaging/nfpm.yaml`)                   | Published from `v0.1.0`; no Homebrew tap yet   |
| GitHub Action (`action/`)                                              | `evcoredotco/octane/action@v0`                 |
| Docusaurus website                                                     | Deployed to GitHub Pages                       |
| Connection profiles (ADR 0010)                                         | Sample profile only; the CLI does not load it |
| Shared OCPP 1.6 modules (ADR 0020)                                     | Not adopted                                    |
| OCPP 2.0.1 and 2.1                                                     | Out of scope                                   |

## 17. Reading order for new contributors

1. This document (`ARCHITECTURE.md`), especially the principles in
   section 2
2. `AGENTS.md`
3. `CONTRIBUTING.md`
4. The ADRs in `docs/adr/` for the area you are changing
5. The stories in `scenarios/v16/`, to see how the design plays out in
   practice. Run `octane keywords list` for the full keyword catalog.

For Claude Code users specifically, also read `CLAUDE.md`.

## 18. References

- **OCPP 1.6:** <https://www.openchargealliance.org/protocols/ocpp-16/>
- **CitrineOS:** <https://citrineos.github.io/>
- **Robot Framework user guide (mental-model precedent):**
  <https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html>
