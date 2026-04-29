# OCTANE

**OCPP Conformance Testing & Network Evaluation**

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)

OCTANE is an open-source, AI-native conformance harness for OCPP 1.6J,
 Charging Station Management Systems (CSMS). It runs
against an **unmodified CSMS** and verifies wire-level conformance to
the publicly published OCPP specifications, automated and CI-friendly.

OCTANE has zero adoption cost for CSMS teams: one CLI command,
no code changes, no sidecar service, no privileged admin API.

```bash
# Install via Go toolchain
go install github.com/evcoreco/octane/cmd/octane@latest

# Run the full OCPP 1.6J suite against your CSMS
octane run scenarios/v16/

# Or use the GitHub Action
# See examples/ci/github-actions/ocpp-conformance.yml
```

> **Status:** pre-alpha. The architecture is designed and specified;
> implementation is in progress. See [§ "What's implemented
> today"](#whats-implemented-today) for the honest inventory.

---

## Why OCTANE

Conformance verification of an OCPP CSMS traditionally happens through
manual, operator-driven workflows that don't scale to CI-gated
development. OCTANE fills that gap with a single static binary that
runs the same wire-level checks unattended, in seconds, against an
unmodified CSMS.

| | OCTANE |
|--|--------|
| Runs in CI without operator interaction | ✅ |
| Requires CSMS-side changes | No |
| Verifies internal CSMS state | Out of scope (wire-only by design) |
| Verifies wire behavior against the OCPP spec | ✅ |
| Open source | ✅ |

OCTANE is the day-to-day tool that gets your CSMS to a state where
formal certification by an external authority is uneventful. See
[`docs/conformance-claim.md`](./docs/conformance-claim.md) for the
precise scope of OCTANE's conformance assertion.

## Quick start

### Install

| Channel | Command |
|---------|---------|
| Debian/Ubuntu | `sudo apt install octane` |
| Fedora/RHEL/CentOS | `sudo dnf install octane` |
| macOS | `brew install evcoreco/octane/octane` |
| Windows | `scoop install octane` |
| Docker | `docker pull ghcr.io/evcoreco/octane` |
| From source | `git clone … && make build` |

> Distribution channels are defined and packaged but the public APT
> repo and Homebrew tap are not yet hosted. Build from source for now.

### Run against CitrineOS locally

```bash
# 1. Bring up CitrineOS
git clone https://github.com/citrineos/citrineos-core
cd citrineos-core/Server
docker compose up -d

# 2. In another shell, set up an OCTANE project
mkdir my-conformance && cd my-conformance
cat > octane.yml <<'YAML'
schema_version: "1"
csms:
  url: ws://localhost:8081/ocpp/CP01
  ocpp_version: "1.6"
  subprotocol: ocpp1.6
profile: citrineos
auth:
  mode: none
defaults:
  timeout: 30s
  seed: 42
report:
  output_dir: ./reports
  formats: [json, robot-xml]
YAML

# 3. Drop a story under scenarios/v16/, then run
octane run scenarios/v16/connector_reservation_faulted.story

# 4. Inspect the report
jq '.summary' reports/report.json
```

### CI integration

```yaml
# .github/workflows/conformance.yml
name: OCPP Conformance
on: [push, pull_request]

jobs:
  conformance:
    runs-on: ubuntu-latest
    services:
      citrineos:
        image: ghcr.io/citrineos/citrineos:1.6.0
        ports: ["8081:8081"]
    steps:
      - uses: actions/checkout@v4
      - uses: evcoreco/octane-action@v0
        with:
          config: octane.yml
          stories: scenarios/v16/
          fail-on: major
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: conformance-reports
          path: reports/
```

## How it works

OCTANE drives the CSMS by impersonating one or more charging stations
over the OCPP-J WebSocket protocol. Tests are declarative `.story`
files in a Gherkin-flavored DSL. CSMS-specific *connection metadata*
(URL templates, ports, subprotocol mappings) lives in small YAML
files called connection profiles, owned by the operator. There is no
CSMS-specific *behavioral* adaptation — domain keywords are identical
for every CSMS implementing a given OCPP version.

```
┌────────────────────────┐         ┌──────────────────┐
│ scenarios/             │         │ CSMS             │
│   v16/                 │         │ (unmodified)     │
│     connector_reservation_     │  WSS    │                  │
│       story            │  ────▶  │ CitrineOS, SteVe │
│                        │         │ MaEVe, vendor X  │
│ ↓ resolved by          │  ◀────  │                  │
│ ↓ keyword library      │         │                  │
│ ↓ + active profile     │         └──────────────────┘
│                        │
│ pkg/engine             │         ┌──────────────────┐
│   ↓                    │  emit   │ report.json      │
│ pkg/transport          │  ────▶  │ output.xml (Robot│
│   ↓                    │         │   Framework      │
│ pkg/wire (OCPP-J)      │         │   compatible)    │
└────────────────────────┘         └──────────────────┘
```

For the full picture, read [`ARCHITECTURE.md`](./ARCHITECTURE.md). For
the binding rules, read [`.specify/memory/constitution.md`](./.specify/memory/constitution.md).

## What a story looks like

```
# scenarios/v16/connector_reservation_faulted.story
Meta
    Name:        Connector reservation faulted
    Id:          connector_reservation_faulted
    Spec-Ref:    OCPP-J 1.6 §6.40 ReserveNow
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

Stories are read by certification reviewers, version-controlled with
the project being tested, and trace to specific sections of the
published OCPP specifications via the `Spec-Ref` meta key.

## Repository layout

```
.
├── .specify/                    # Spec-Kit scaffolding (constitution, templates, scripts)
├── .claude/                     # Claude Code agents and slash commands
├── .github/workflows/           # CI: ci.yml, reference.yml, release.yml, docs.yml
├── ARCHITECTURE.md              # Full design narrative (start here)
├── AGENTS.md                    # Cross-tool agent contract
├── CLAUDE.md                    # Claude Code project memory
├── docs/
│   ├── adr/                     # Architecture Decision Records (0001–0016)
│   ├── conformance-claim.md     # public conformance scope statement
│   └── man/                     # scdoc sources for §5 and §7 man pages
├── specs/
│   ├── 001-story-parser/        # .story → AST
│   ├── 002-wire-engine/         # transport + framing + determinism
│   ├── 003-keyword-api/         # API surface + registry + resolver
│   ├── 004-primitive-keywords/  # transport-level primitives
│   ├── 005-dependency-cache/    # runner (DAG) + cache
│   ├── 006-cli-action/          # CLI + GitHub Action + GitLab
│   └── 007-reports/             # JSON + Robot XML
├── scenarios/
│   ├── v16/                     # OCPP 1.6 stories (helpers + reservation)
│   └──                     # OCPP 1.6 stories (boot, authorize)
├── action/                      # GitHub Action manifest
├── packaging/                   # nfpm.yaml for .deb/.rpm
├── scripts/                     # gen-manpages.sh, gen-completions.sh
├── website/                     # Docusaurus site
├── .goreleaser.yaml             # Release orchestration (activates with binary)
├── Makefile                     # Targets reserved for when code lands
├── CHANGELOG.md
└── LICENSE                      # Apache-2.0
```

> The `pkg/` Go tree and `go.mod` are intentionally absent from
> this scaffolding — see [§ "What's implemented today"](#whats-implemented-today).

CSMS connection profiles ship as sample YAML files in OCTANE itself
(see `connections/`). Operators adapt them or write their own. There
are no separate per-CSMS code repositories.

## What's implemented today

Implementation is underway following the
[GitHub Spec-Kit](https://github.com/github/spec-kit) workflow
defined in `.specify/`. Specs 001–006 are fully implemented; spec 007
(reports) is in progress. Each piece of the design lands as code only
after its spec has been refined to implementation-ready detail.

### Done

- Constitution v1.4.0 with 12 ratified principles
- 16 ADRs covering license, language, transport, reference CSMS,
  story framework, DSL grammar, keyword library, multi-station,
  reporting, connection profiles, man pages, shell completion,
  website, IP and authoring guidelines, test dependency graph,
  and cache subsystem (content-addressed file tree)
- 7 specs (`001-story-parser`, `002-wire-engine`,
  `003-keyword-api`, `004-primitive-keywords`,
  `005-dependency-cache`, `006-cli-action`, `007-reports`) with
  full acceptance criteria, plans, and atomic tasks
- `.specify/` Spec-Kit scaffolding (constitution, templates, helper
  scripts, slash commands)
- 8 Claude Code subagents (architect, backend, keyword-author,
  devops, qa, security, reviewer, docs)
- Story DSL parser (`pkg/story/`) — spec 001
- WebSocket transport, OCPP-J frame parser, deterministic clock/rand
  (`pkg/transport/`, `pkg/wire/`, `pkg/engine/`) — spec 002
- Keyword API surface, registry, resolver
  (`pkg/keywords/api/`, `pkg/keywords/registry/`) — spec 003
- Primitive keyword layer (`pkg/keywords/primitive/`) — spec 004
- Story runner with DAG, worker pool, sharding, and content-addressed
  file cache (`pkg/runner/`, `pkg/cache/`) — spec 005
- Complete `octane` CLI built on cobra: `run`, `validate stories`,
  `keywords list/resolve`, `cache info/prune/clear/key`,
  `completion bash|zsh|fish|powershell` (`cmd/octane/`) — spec 006
- GitHub Action (`action/action.yml`, `action/Dockerfile`,
  `action/entrypoint.sh`) — spec 006
- Example `.story` files: 7 conformance stories and 3 helpers
- `CONTRIBUTING.md`, `docs/conformance-claim.md`,
  `docs/getting-started.md`, `docs/cli-reference.md`,
  `docs/configuration.md`
- Man-page sources (§5 for config and story, §7 for concepts)
- Packaging via goreleaser + nfpm (`.deb`, `.rpm`, Homebrew, SBOM)
- CI workflow files

### In progress / not yet written

- Robot XML emitter (`pkg/report/`) — spec 007
- OCPP 1.6 / 1.6 / 2.1 domain keyword layers
- Sample connection profile YAML files
- Public APT/RPM repos and Homebrew tap

## For contributors

OCTANE follows **spec-driven development**. No code lands without a
merged spec. The flow:

1. `/specify <feature>` — draft `specs/NNN-feature/spec.md`
2. `/plan` — fill `plan.md` with technical approach + ADR drafts
3. `/tasks` — decompose into atomic, agent-assignable tasks
4. `/implement T-NNN-MM` — execute one task under the right subagent

Slash commands are defined under `.claude/commands/`. The full agent
contract is `AGENTS.md`. Read the constitution
(`.specify/memory/constitution.md`) before opening a PR.

### Local development

```bash
# Install build tooling
make install-tools

# Format, lint, test, build
make format
make lint
make test
make build

# Generate man pages (requires scdoc)
sudo apt install scdoc
make man

# Generate shell completions
make completions

# Build a snapshot release with .deb and .rpm
make package

# Run docs site locally
make docs-serve
```

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
