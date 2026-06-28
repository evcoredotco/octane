# OCTANE

**OCPP Conformance Testing & Network Evaluation**

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://go.dev)

OCTANE is an open-source conformance harness for OCPP 1.6J. It simulates
one or more charging stations over a WebSocket connection and verifies that
your Charging Station Management System (CSMS) responds correctly to each
OCPP 1.6 message type.

The CSMS under test needs no modification. OCTANE speaks OCPP-J natively
from the charging station side: it dials your endpoint, exchanges messages,
and asserts wire-level conformance. Tests are declarative `.story` files that
read like plain English and map directly to sections of the OCPP specification.

> **Status:** pre-alpha. The core is fully implemented and the OCPP 1.6
> keyword layer covers 17 message types. See [OCPP 1.6 coverage](#ocpp-16-coverage)
> for the honest inventory. There are no published packages yet; build from
> source.

---

## What it does

OCTANE impersonates a charging station over the OCPP-J WebSocket protocol.
Each `.story` file is a self-contained test scenario: it establishes a
connection, drives the CSMS through a sequence of OCPP exchanges, and asserts
the CSMS response at each step. Stories declare their dependencies, and the
runner resolves those into a DAG so prerequisite scenarios execute first.

A content-addressed cache prevents redundant runs: a story that passed
against the same CSMS build is skipped until the story file or the binary
itself changes. The cache makes large suites fast in CI without sacrificing
repeatability.

Reports emit as JSON and Robot Framework-compatible XML, suitable for
uploading as CI artifacts or feeding into a test-management system.

OCTANE makes no assertion about internal CSMS state. It observes only the
wire. If the CSMS sends the right bytes in the right order, the story passes.

## Quick start

### Install (build from source)

No packages are published yet. Build from source with the standard Go
toolchain (Go 1.26+):

```bash
git clone https://github.com/evcoreco/octane
cd octane
go build ./cmd/octane
```

Or run without building:

```bash
go run ./cmd/octane run --csms-endpoint ws://localhost:9210
```

### Minimal octane.yml

Create `octane.yml` in the root of your project:

```yaml
storiesDir:  scenarios/v16
ocppVersion: "1.6"
maxParallel: 1
```

The `--csms-endpoint` flag (or the `OCTANE_CSMS_ENDPOINT` environment
variable) tells OCTANE where your CSMS WebSocket endpoint is. It is
not stored in `octane.yml` because it typically differs between environments.

### Run one story

```bash
./octane run scenarios/v16/station_boot_accepted.story \
    --csms-endpoint ws://localhost:9210
```

### Run the full suite

```bash
./octane run --csms-endpoint ws://localhost:9210
```

OCTANE discovers all `.story` files under `storiesDir`, resolves the
dependency graph, and executes them in order.

### GitHub Action

```yaml
# .github/workflows/conformance.yml
name: OCPP Conformance
on: [push, pull_request]

jobs:
  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: evcoreco/octane-action@v0
        with:
          csms-endpoint: ws://your-csms-host:9210
          stories: scenarios/v16/
          fail-on: any
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: conformance-reports
          path: reports/
```

See [`action/README.md`](./action/README.md) for the full input reference.

---

## OCPP 1.6 coverage

The following message types have conformance stories. "Stories" is the count
of `.story` files exercising that message type.

| OCPP Message | Direction | Stories | Notes |
|---|---|---|---|
| BootNotification | CS → CSMS | 2 | Accepted; full boot sequence |
| StatusNotification | CS → CSMS | 1 | Available state |
| Heartbeat | CS → CSMS | 1 | Part of boot sequence |
| Authorize | CS → CSMS | 1 | Plug-in-first flow |
| StartTransaction | CS → CSMS | 2 | Plug-in-first; identification-first |
| StopTransaction | CS → CSMS | 1 | Normal stop |
| MeterValues | CS → CSMS | 1 | Periodic sampling |
| RemoteStartTransaction | CSMS → CS | 1 | Accepted |
| RemoteStopTransaction | CSMS → CS | 1 | Accepted |
| Reset | CSMS → CS | 2 | Soft; Hard |
| UnlockConnector | CSMS → CS | 2 | Accepted; Failed |
| ChangeAvailability | CSMS → CS | 2 | Operative; Inoperative |
| GetConfiguration | CSMS → CS | 1 | Key list returned |
| ChangeConfiguration | CSMS → CS | 1 | Accepted |
| ClearCache | CSMS → CS | 1 | Accepted |
| ReserveNow | CSMS → CS | 1 | Faulted response |
| CancelReservation | CSMS → CS | 1 | Accepted |

**Not yet covered:** DiagnosticsStatusNotification, FirmwareStatusNotification,
DataTransfer, TriggerMessage, SendLocalList, GetLocalListVersion. Keywords and
stories for these message types have not been written yet.

---

## Story format

Stories are declarative text files in a Gherkin-flavored DSL. Here is
`scenarios/v16/connector_reservation_faulted.story`:

```
# Connector reservation faulted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -6.40 ReserveNow
# correctly handles the case where a charging station rejects a
# reservation request by responding with status "Faulted". The CSMS
# must accept the response without raising an OCPP-level error;
# whether or not it surfaces the rejection to upstream operator
# tooling is out of OCTANE's wire-only scope.
#
# This is a CSMS-initiated, single-station scenario. The dependency
# chain ensures the connector under test is in the "Available" state
# (per OCPP-J 1.6 -4.7) before the reservation request is sent;
# without that prerequisite, the CSMS may legitimately respond
# differently and the test would fail for the wrong reason.

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

The `Depends` block declares that `connector_status_available` must pass
first (scoped per-station). The runner enforces this automatically via DAG
resolution. Stories are version-controlled alongside the CSMS under test and
trace to specific OCPP spec sections via `Spec-Ref`.

See [`docs/configuration.md`](./docs/configuration.md) for the full story
metadata reference.

---

## CLI reference

| Subcommand | Description |
|---|---|
| `octane run [stories...] --csms-endpoint <url>` | Run conformance stories against a CSMS |
| `octane validate [stories...]` | Parse and validate story files without running them |
| `octane keywords list` | List all registered keywords |
| `octane keywords resolve <pattern>` | Resolve a keyword pattern against the registry |
| `octane cache info` | Show cache statistics |
| `octane cache prune` | Remove stale cache entries |
| `octane cache clear` | Remove all cache entries |
| `octane cache key` | Print the cache key for a story file |
| `octane completion bash\|zsh\|fish\|powershell` | Generate shell completion script |

**Selected `octane run` flags:**

| Flag | Default | Description |
|---|---|---|
| `--csms-endpoint` | (required) | Base WebSocket URL of the CSMS under test |
| `--max-parallel` | `1` | Maximum concurrently executing stories |
| `--shard N/M` | (none) | Run the Nth of M shards (for CI fan-out) |
| `--ocpp-version` | (all) | Restrict run to stories declaring this version |
| `--fail-on` | `any` | Exit non-zero on `any` failure or only on `major` |
| `--report-dir` | `reports/` | Directory for JSON and Robot XML reports |
| `--no-trace-on-pass` | false | Omit wire trace from reports for passing stories |
| `--insecure-skip-verify` | false | Disable TLS certificate verification (insecure) |

Full flag documentation: [`docs/cli-reference.md`](./docs/cli-reference.md).

---

## How it works

OCTANE opens a WebSocket connection to the CSMS endpoint, identifies itself
as a charging station using the station ID embedded in each story, and drives
the protocol exchange step by step. Each step in a story resolves to a
*keyword* — a Go function in the keyword library that knows how to send a
specific OCPP message, wait for a response, and assert the response fields.
The keyword library has three layers: primitive (raw WebSocket send/receive),
domain (OCPP 1.6 message types), and lifecycle (connect/disconnect/register).
The runner resolves story dependencies into a DAG, executes stories in
dependency order using a configurable worker pool, consults the content-
addressed file cache to skip unchanged stories, and emits a JSON report plus
a Robot XML report on completion.

---

## Repository layout

```
.
├── .specify/                    # Spec-Kit scaffolding (constitution, templates, scripts)
├── .claude/                     # Claude Code agents and slash commands
├── .github/workflows/           # CI: ci.yml, reference.yml, release.yml, docs.yml
├── ARCHITECTURE.md              # Full design narrative
├── AGENTS.md                    # Cross-tool agent contract
├── CLAUDE.md                    # Claude Code project memory
├── CONTRIBUTING.md
├── action/                      # GitHub Action (action.yml, Dockerfile, entrypoint.sh)
├── cmd/
│   └── octane/                  # CLI entry point and subcommands
│       └── internal/config/     # Config struct, env-var layering, flag overrides
├── docs/
│   ├── adr/                     # Architecture Decision Records (0001–0020)
│   ├── cli-reference.md
│   ├── configuration.md
│   ├── conformance-claim.md
│   └── getting-started.md
├── go.mod
├── octane.yml                   # Project-level OCTANE config (storiesDir, ocppVersion)
├── packaging/                   # nfpm.yaml for .deb/.rpm
├── pkg/
│   ├── cache/                   # Content-addressed file cache
│   ├── engine/                  # Deterministic clock and rand
│   ├── keywords/
│   │   ├── api/                 # Keyword type definitions and interfaces
│   │   ├── lifecycle/           # Connection lifecycle keywords
│   │   ├── ocpp16/              # OCPP 1.6 domain keywords (30 keywords, 17 message types)
│   │   ├── primitive/           # Raw send/receive transport keywords
│   │   └── registry/            # Keyword registration and lookup
│   ├── report/                  # Report types; JSON and Robot XML emitters
│   ├── runner/                  # DAG resolver, worker pool, story executor
│   ├── story/                   # .story DSL parser → AST
│   ├── transport/               # WebSocket connection management
│   └── wire/                    # OCPP-J framing (Call, CallResult, CallError)
├── scenarios/
│   └── v16/                     # 21 OCPP 1.6 conformance stories
├── scripts/                     # gen-manpages.sh, gen-completions.sh
├── specs/                       # Spec-Kit specs (001–007)
├── test/                        # Integration and end-to-end tests
├── website/                     # Docusaurus documentation site
├── CHANGELOG.md
└── LICENSE                      # Apache-2.0
```

---

## Contributing

OCTANE uses spec-driven development. No code lands without a merged spec.
The flow:

1. `/specify <feature>` — draft `specs/NNN-feature/spec.md`
2. `/plan` — fill `plan.md` with technical approach and any ADR drafts
3. `/tasks` — decompose into atomic, agent-assignable tasks
4. `/implement T-NNN-MM` — execute one task under the right subagent

Slash commands are defined under `.claude/commands/`. The full agent
contract is [`AGENTS.md`](./AGENTS.md). Read the constitution
(`.specify/memory/constitution.md`) before opening a PR.

### Local development

```bash
# Format, lint, test, build
make format
make lint
make test
make build

# Generate shell completions
make completions

# Run the docs site locally
make docs-serve
```

---

## License

Apache-2.0. See [`LICENSE`](./LICENSE) and
[ADR 0001](./docs/adr/0001-license.md) for the rationale.

## Acknowledgements

- The **Open Charge Alliance** for publishing the OCPP specifications.
- **CitrineOS** (LF Energy) for being a credible open-source CSMS
  reference target.
- **Robot Framework** for the structural metaphor that informs
  OCTANE's keyword library design.

OCTANE is not affiliated with or endorsed by the Open Charge Alliance,
LF Energy, or CitrineOS.
