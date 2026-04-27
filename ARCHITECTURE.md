# OCTANE — Architecture

> **Audience:** engineers, certification reviewers, and contributors who
> want a single document explaining how OCTANE is designed and why. It
> consolidates the ratified ADRs and the constitution into a coherent
> narrative. Where this document and an ADR disagree, the ADR wins —
> this is summary, not specification.
>
> **Scope:** OCTANE v0 (pre-alpha). Decisions tracked through ADR 0013.
>
> **Last reviewed:** 2026-04-26

---

## 1. What OCTANE is

OCTANE — *OCPP Conformance Testing & Network Evaluation* — is an
open-source conformance harness for OCPP 1.6J, 2.0.1, and 2.1
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

OCTANE's constitution (`.specify/memory/constitution.md`, currently at
v1.2.0) is the only document that overrides the rest. Every spec, ADR,
and design decision must comply with it. The principles, in summary:

| # | Principle | Operational meaning |
|---|-----------|---------------------|
| I | Conformance Above Convenience | Every conformance test traces to a section of the OCPP specification. |
| II | Two Distribution Surfaces, One Engine | CLI and Action expose the same engine; no surface-only features. |
| III | Reference-Validated | Every test passes against a pinned CitrineOS commit before being marked stable. |
| IV | Determinism and Reproducibility | Same inputs, same outputs (modulo timestamps). Seeded RNG, injected clock, sorted serializers. |
| V | Go-First, Stdlib-Heavy | Go 1.23. Third-party deps require an ADR. WebSocket is the single allowed exception. |
| VI | Test Cases Are Code, Not Configuration | Stories are a declarative *surface*; the keyword library that executes them is typed Go. |
| VII | Public API Stability | `pkg/` follows semver. `internal/` does not. |
| VIII | Spec-Driven Development | Code does not land before its spec merges. |
| IX | AI Agents Are Bounded Contributors | Agents have declared scopes; agent commits get human sign-off. |
| X | Security and Compliance Are Continuous | TLS on by default, secrets never in artefacts, dependencies scanned on every PR. |
| XI | Conformance Is Verified On the Wire | OCTANE never demands changes to the CSMS, never uses admin APIs, never depends on a vendor adapter. |
| XII | Scenarios Are Declarative; Adaptation Lives in Profiles | Stories carry no CSMS-specific assumptions; profiles carry no conformance logic. |

Principles XI and XII are the two pivots that distinguish the current
design from earlier drafts. They emerged from a deliberate rejection
of an adapter-based model (see §12) and align OCTANE with a
**zero-cooperation-cost** adoption philosophy: any CSMS team can
run OCTANE against an unmodified deployment with one CLI command.

## 3. The three-layer model

```
┌──────────────────────────────────────────────────────────────────┐
│  Layer 1 — Stories (.story files)                                │
│  ──────────────────────────────────────────────────────────────  │
│  Declarative Gherkin-flavored scenarios. One per OCPP section/  │
│  protocol behavior under test.                                  │
│  Live in scenarios/v16/, scenarios/v201/, scenarios/v21/.        │
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
│  • report.json     — OCTANE-native, byte-deterministic           │
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

```
Meta
    Spec-Ref:    OCPP-1.6 / TC_048_1_CSMS
    Title:       Reservation of a Connector — Faulted
    Tags:        reservation, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId, idTag

Background
    Given the CSMS is reachable
    And   the profile defines station "CP01"

Scenario: CSMS handles a Faulted reservation response
    When  the CSMS sends ReserveNow with connectorId {connectorId}
          and idTag "{idTag}" to station "CP01"
    Then  station "CP01" responds with ReserveNow.conf status "Faulted"

Teardown
    Disconnect station "CP01"
```

Required Meta keys carry OCPP specification traceability (`Spec-Ref`,
required for conformance tests), declare station count for preflight
resource allocation (`Stations`), and classify the scenario through
`Tags` (one of `wire-only`, `multi-station`, `operator-assisted`,
`helper`; optionally `pure-protocol` to declare independence from
any fixture).

The `Parameters` key declares story inputs that resolve from
`octane.yml`. Without it, stories cannot vary by deployment. The exact
schema is sketched but not yet ratified into an ADR.

The grammar is small enough to fit a recursive-descent parser in
~600–800 lines of Go (target). Diagnostic quality (line + column
+ Levenshtein-suggested keyword on resolution failure) is where most
of the engineering goes.

Defined in **ADR 0006**.

## 5. The keyword library

Keywords are typed Go functions that map step text to wire actions.
They self-register at process start via `init()`. Pattern collisions
inside the same `(layer, ocpp-version)` tuple panic at startup —
caught in CI, never at runtime.

### Two layers, deterministic resolution

```
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

| Robot Framework | OCTANE |
|---|---|
| `@keyword("Pattern ${arg}")` decorator | `registry.Register(api.Keyword{Pattern: "..."})` |
| Library file (single layer) | `LayerDomain` for OCPP semantics, `LayerPrimitive` for transport |
| `types={"count": int}` | `{name:int}` placeholders |
| `raise AssertionError` | `return fmt.Errorf(...)` |
| `from robot.api import logger` | `state.Logf(...)` |
| Per-test scope | `api.State` per-scenario by construction |

OCTANE does not use Robot Framework at runtime, but borrows its
mental model and emits its `output.xml` format (see §8).

Defined in **ADR 0007**.

## 6. Connection profiles

OCTANE has no CSMS-specific behavioral adaptation surface
(constitution principle XII). What is legitimately CSMS-specific is
**connection metadata** — how to reach the CSMS at the network
level. This lives in small YAML files called connection profiles
(ADR 0010), owned by the operator, not the vendor.

A connection profile is roughly 30 lines:

```yaml
schema_version: "1"
name: citrineos
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

## 7. Test dependency graph

OCPP scenarios are not independent: a reservation test cannot run
against a CSMS that has not registered the station, and a station
cannot register without first establishing an OCPP-J WebSocket
connection. Most scenarios require a chain of prior state to reach
the point where the actual conformance assertion can be exercised.

OCTANE models the test suite as a **directed acyclic graph of test
cases**. Every story can be a prerequisite for another via the
`Depends:` Meta key:

```
Meta
    Name:     Connector reservation faulted
    Id:       connector_reservation_faulted
    Spec-Ref: OCPP-J 1.6 §6.40 ReserveNow
    Stations: 1
    Depends:
      - id:    connector_status_available
        scope: per-station
```

The runner walks the dependency chain transitively, executes the
prerequisites in topological order, then executes the requested
story. Each prerequisite is itself a story that may have its own
dependencies. For the example above, the resolved chain is:

```
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

| Status | Meaning |
|--------|---------|
| `passed` | Story ran and all assertions held. |
| `failed` | Story ran and at least one assertion failed. |
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
defined by the cache (see §8).

Defined in **ADR 0015**.

## 8. Cache and locks

To make the dependency graph efficient, OCTANE caches test results
across `octane run` invocations. The cache is a **content-addressed
file tree** at `$XDG_CACHE_HOME/octane/cache/` (overridable via
`OCTANE_CACHE_DIR`). Each cache entry is two JSON files: a small
`result.json` and an optional sibling `trace.json` with the wire
frames.

```
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

| Field | Source |
|-------|--------|
| `test_id` | story `Id` Meta key |
| `scope_key` | station handle for `per-station`, run ID for `per-run`, empty for `global` |
| `csms_endpoint_sha` | SHA-256 of (URL + subprotocol + auth-mode) |
| `octane_version` | from build info |
| `ocpp_version` | from story Meta or config |
| `story_content_sha` | SHA-256 of story file + transitively all prerequisites' content |
| `parameter_sha` | SHA-256 of bound parameters |

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

```
octane cache info     # cache directory, total size, entry count
octane cache prune    # remove entries older than --max-age (default 30d)
octane cache clear    # remove all cache content
octane cache key <story> [--scope STATION]   # print resolved key
octane cache show <key>                       # cat result.json
octane cache trace <key>                      # cat trace.json
```

Inspection is `cat | jq`. No SQL, no third-party tool to install.

Defined in **ADR 0016**.

## 9. Multi-station orchestration

Many stateful OCPP scenarios are reachable on the wire by coordinating
two or more simulated stations. A story declares the count in `Meta`:

```
Meta
    Spec-Ref:    OCPP-2.0.1 / TC_E_07_CS
    Stations:    2
```

The runner allocates handles `"CP01"`, `"CP02"`, … on preflight.
Steps reference stations by handle. Steps run sequentially in
declared order; concurrency is opt-in via `Parallel ... End-Parallel`
blocks:

```
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

## 10. Reports

Every run produces two artefacts:

| File | Format | Purpose |
|------|--------|---------|
| `report.json` | OCTANE-native JSON, byte-deterministic | Source of truth for certification |
| `output.xml` | Robot Framework 7.x output schema | Ecosystem reporting (Allure, ReportPortal, Jenkins, GitLab, GitHub Actions) |

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

## 11. Distribution and operator surface

### CLI

The canonical invocation is:

```
octane run scenarios/v16/TC_048_1_CSMS.story
octane run scenarios/v201/                     # whole directory
octane run scenarios/                           # everything
```

Configuration resolves through a hierarchy (ADR 0019, in scope but
not yet ADR'd in this revision):

1. `--config` flag
2. `OCTANE_CONFIG` environment variable
3. `./octane.yml` or `./.octane/config.yml`
4. `$XDG_CONFIG_HOME/octane/config.yml`
5. `/etc/octane/config.yml`

A sample `octane.yml`:

```yaml
schema_version: "1"
csms:
  url: ws://localhost:8081/ocpp/CP01
  ocpp_version: "1.6"
connection: citrineos                  # resolves a connection profile per ADR 0010
auth:
  mode: none
defaults:
  timeout: 30s
  seed: 42
report:
  output_dir: ./reports
  formats: [json, robot-xml]
```

### Distribution channels

OCTANE distributes through standard Linux/macOS/Windows channels:

| Channel | Path |
|---------|------|
| Debian/Ubuntu | `apt install octane` (via APT repo) |
| Fedora/RHEL | `dnf install octane` (via RPM repo) |
| macOS | `brew install evcoreco/octane/octane` |
| Windows | `scoop install octane` |
| Docker | `docker pull ghcr.io/evcoreco/octane` |
| Direct | static binaries, signed via cosign, SBOM-attested |

Packaging is orchestrated by `goreleaser` reading `.goreleaser.yaml`
and `packaging/nfpm.yaml`. Release artefacts include man pages, shell
completions, and the LICENSE.

### Documentation

OCTANE ships professional Unix-style documentation (ADR 0011):

| Section | Content | Tool |
|---------|---------|------|
| `man 1 octane`, `man 1 octane-run`, ... | Per-subcommand reference | Generated from cobra |
| `man 5 octane.yml` | Config file reference | Hand-written via scdoc |
| `man 5 octane.story` | Story DSL reference | Hand-written via scdoc |
| `man 7 octane` | Concepts overview | Hand-written via scdoc |

Shell completion (ADR 0012) ships for bash and zsh, with both static
(subcommands, flags) and dynamic completion (story file paths,
profile names, registered keywords). Dynamic completion functions
are bound by a hard rule: read-only and side-effect-free, enforced
by a build-time vettool.

A separate Docusaurus website at `https://octane.dev/` (ADR 0013)
covers tutorials, conceptual narrative, and the keyword catalog.
It is **not** an HTML render of the man pages — different reading
modes deserve different content shapes. Three of its reference pages
are mechanically generated (CLI, config schema, keyword catalog); the
rest is hand-written.

## 12. Design history

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

## 13. Repository layout

```
octane/
├── .specify/
│   ├── memory/
│   │   └── constitution.md          # binding principles, v1.4.0
│   ├── scripts/bash/                # spec-kit helper scripts
│   └── templates/                   # spec / plan / tasks templates
├── .claude/
│   ├── agents/                      # subagent role definitions
│   └── commands/                    # /specify /plan /tasks /implement /adr /check
├── .github/workflows/               # ci.yml, reference.yml, release.yml, docs.yml
├── docs/
│   ├── adr/                         # 17 ADRs, all active, sequentially numbered
│   ├── conformance-claim.md         # public conformance scope statement
│   └── man/                         # scdoc sources for §5 and §7
├── specs/
│   ├── 001-story-parser/            # .story → AST (no I/O)
│   ├── 002-wire-engine/              # transport + wire framing + clock/rand
│   ├── 003-keyword-api/              # api package + registry + resolver
│   ├── 004-primitive-keywords/       # transport-level primitive keywords
│   ├── 005-dependency-cache/         # runner (DAG) + content-addressed cache
│   ├── 006-cli-action/               # CLI + GitHub Action + GitLab integration
│   └── 007-reports/                  # JSON + Robot XML emitters
├── scenarios/
│   ├── v16/                         # OCPP 1.6 stories: helpers + connector_reservation_faulted
│   └── v201/                        # OCPP 2.0.1 stories: boot_notification_*, authorize_concurrent_rejected
├── action/                          # GitHub Action wrapper (manifest only; binary lands later)
├── examples/                        # Consumer copy-paste artifacts (CI workflows for GH/GL)
├── packaging/                       # nfpm.yaml for .deb/.rpm
├── scripts/                         # gen-manpages.sh, gen-completions.sh
├── website/                         # Docusaurus site
├── AGENTS.md                        # cross-tool agent contract
├── CLAUDE.md                        # Claude Code project memory
├── ARCHITECTURE.md                  # this document
├── CONTRIBUTING.md                  # author / contributor guide
├── CHANGELOG.md
├── LICENSE                          # Apache-2.0
├── Makefile                         # targets reserved; activate when code lands
├── .goreleaser.yaml                 # release orchestration (activates with binary)
└── .golangci.yaml                   # lint config (activates with code)
```

> The `pkg/` directory tree and `go.mod` are intentionally absent.
> All Go code is specced (ADRs 0007, 0015, 0016, 0017; specs 001
> and 002) but not yet written. Implementation follows the
> Spec-Kit workflow.

## 14. ADR index

All ADRs are active and accepted. The numbering is sequential without
gaps; superseded design history was removed during a 2026-04-26
cleanup so that new contributors see only the current design.

| # | Title |
|---|-------|
| 0001 | Adopt Apache-2.0 License |
| 0002 | Go as the Engine Language |
| 0003 | WebSocket Library (`nhooyr.io/websocket`) |
| 0004 | CitrineOS as the Reference CSMS |
| 0005 | Story-Driven Conformance Framework |
| 0006 | `.story` Gherkin-Flavored DSL |
| 0007 | Keyword Library Layering — Primitive and Domain |
| 0008 | Multi-Station Orchestration |
| 0009 | Robot Framework `output.xml` Compatibility |
| 0010 | Connection Profiles — User-Owned YAML |
| 0011 | Manual Pages (Cobra + scdoc) |
| 0012 | Shell Completion (bash + zsh, dynamic) |
| 0013 | Web Documentation (Docusaurus) |
| 0014 | Intellectual Property and Authoring Guidelines |
| 0015 | Test Dependency Graph |
| 0016 | Cache and Lock Subsystem — Content-Addressed File Tree |

Forthcoming (sketched in conversation, not yet written):

| Title | Notes |
|-------|-------|
| CLI Surface and Subcommand Structure | Locks `octane run`, exit codes, global flags |
| Configuration Resolution and Schema | XDG chain, YAML schema, validation |
| Distribution Channels | `.deb`, `.rpm`, Homebrew, Scoop, signing |
| Story Parameters | Project-supplied inputs in story Meta |

## 15. What is implemented today

This is honest scaffolding inventory, not aspirational marketing.

| Component | Status |
|-----------|--------|
| Constitution v1.4.0 | Done |
| ADRs 0001–0016 | Done |
| Spec 001 (bootstrap engine), Spec 002 (story framework) | Specs merged, implementation pending |
| `.specify/` Spec-Kit scaffolding | Done |
| `.claude/agents/` (8 subagents) and `/slash` commands | Done |
| Example `.story` files: 7 conformance stories (`boot_notification_accepted`, `boot_notification_malformed`, `authorize_concurrent_rejected`, `boot_sequence_accepted`, `connector_reservation_faulted`, `transaction_pluginfirst_accepted`, `transaction_identificationfirst_accepted`) and 3 helpers | Done |
| All Go code (`pkg/keywords`, `pkg/wire`, `pkg/cache`, `pkg/engine`, `pkg/transport`, `pkg/story` parser, `cmd/octane`) | Specced, not implemented |
| Sample connection profile YAML files (`connections/citrineos.yaml`, etc.) | Specced, not bootstrapped |
| `CONTRIBUTING.md` and `docs/conformance-claim.md` | Done |
| Man-page sources (§5, §7) | Done |
| Man-page generation script + scdoc + cobra hooks | Done (cobra hooks fire when binary exists) |
| Shell completion script | Done (depends on binary) |
| Packaging (`.goreleaser.yaml`, `packaging/nfpm.yaml`) | Done (depends on binary) |
| Docusaurus website skeleton | Done |
| CI workflows (ci, reference, release, docs) | Done (Go jobs activate when code lands) |

The project is in a deliberate **design-complete, code-empty** state.
All architectural decisions are committed in ADRs and specs; no Go
code has been written yet. Implementation will follow the GitHub
Spec-Kit workflow defined in `.specify/`, with each piece of the
design landing as code only after its spec has been refined to
implementation-ready detail.

## 16. Reading order for new contributors

1. This document (`ARCHITECTURE.md`)
2. `.specify/memory/constitution.md`
3. `AGENTS.md`
4. The current spec under `specs/<NNN>-<slug>/`
5. The matching ADRs from `docs/adr/`
6. The example stories under `scenarios/v16/` and `scenarios/v201/`
   to see how the design plays out in practice. The starter keyword
   catalog in spec 002 §10 lists every pattern the example stories
   reference.
7. The example stories under `scenarios/v201/`

For Claude Code users specifically, also read `CLAUDE.md` to
understand the dispatch table and slash-command surface.

## 17. References

- **Constitution:** `.specify/memory/constitution.md`
- **OCPP specifications:**
  - OCPP 1.6: <https://www.openchargealliance.org/protocols/ocpp-16/>
  - OCPP 2.0.1: <https://www.openchargealliance.org/protocols/ocpp-201/>
  - OCPP 2.1: <https://www.openchargealliance.org/protocols/ocpp-21/>
- **CitrineOS:** <https://citrineos.github.io/>
- **Robot Framework user guide (mental-model precedent):**
  <https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html>
- **Spec-Kit pattern:** `.specify/templates/`
