# OCTANE

**OCPP Conformance Testing and Network Evaluation**

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://go.dev)

OCTANE is a Go CLI for running declarative `.story` conformance tests against
a Charging Station Management System (CSMS). It acts from the charging station
side of an OCPP-J WebSocket connection, drives protocol exchanges, and reports
whether the CSMS responds as expected on the wire.

Current implementation status:

- OCPP 1.6 domain keywords are implemented under `pkg/keywords/ocpp16`.
- The checked-in suite contains 23 OCPP 1.6 stories under `scenarios/v16`.
- Stories are parsed into an AST, resolved through dependency ordering, run with
  optional parallelism, and cached by content-addressed keys.
- Reports are written as JSON and Robot Framework-compatible XML.
- The reliable installation path today is building from source in this repo.

Last reviewed: 2026-06-29.

## Prerequisites

- Go 1.26 or later.
- A reachable CSMS WebSocket endpoint, for example `ws://localhost:9210`.
- Optional local tooling for development: `golangci-lint`, `gofumpt`, `golines`,
  `gci`, `goreleaser`, `nfpm`, Docker, and pnpm.

Install the optional Go tools with:

```bash
make install-tools
```

## Build

Build the CLI into `bin/octane`:

```bash
make build
```

Or run directly with the Go toolchain:

```bash
go run ./cmd/octane --help
```

## Configure

The repository includes a working starter config at `octane.yml`:

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

The CSMS endpoint is not stored in `octane.yml`. Pass it to `octane run` with
`--csms-endpoint`.

Configuration precedence is:

1. Built-in defaults.
2. `octane.yml`.
3. `OCTANE_*` environment variables.
4. CLI flags.

See [docs/configuration.md](./docs/configuration.md) for the field reference.

## Run

Run the checked-in OCPP 1.6 suite against a local CSMS:

```bash
./bin/octane run --csms-endpoint ws://localhost:9210
```

Run one story:

```bash
./bin/octane run scenarios/v16/station_boot_accepted.story \
  --csms-endpoint ws://localhost:9210
```

Override a story parameter at runtime:

```bash
./bin/octane run scenarios/v16/transaction_pluginfirst_accepted.story \
  --csms-endpoint ws://localhost:9210 \
  --param valid_idTag=AABBCC \
  --param connectorId=1
```

Validate story syntax and structure without opening a CSMS connection:

```bash
./bin/octane validate stories scenarios/v16
```

A completed run prints a summary like:

```text
passed=23 failed=0 skipped=0 cache-hits=0
report-dir=reports/<run-id>
```

Reports are written below `reports/` unless `--report-dir` is changed or set to
an empty value.

## CLI Surface

Global flags:

| Flag | Purpose |
|---|---|
| `--config` | Path to `octane.yml`; defaults to `octane.yml`. |
| `--cache-dir` | Override the content-addressed result cache directory. |
| `--no-cache` | Bypass cached results and execute every selected story. |
| `--verbose`, `-v` | Enable verbose output. |

Main commands:

| Command | Purpose |
|---|---|
| `octane run [story-paths...]` | Run `.story` files or directories against a CSMS. |
| `octane validate stories [paths...]` | Parse and validate stories without executing them. |
| `octane keywords list` | Print registered primitive and OCPP 1.6 keywords. |
| `octane keywords resolve <step-text>` | Resolve a story step to a keyword pattern. |
| `octane cache info` | Print the effective cache location. |
| `octane cache prune --max-age 24h` | Remove old or expired cache entries. |
| `octane cache clear` | Remove all cached result entries. |
| `octane cache key <story-id>` | Print the placeholder cache key hash for a story ID. |
| `octane completion <shell>` | Generate shell completions. |

Useful `run` flags:

| Flag | Purpose |
|---|---|
| `--csms-endpoint` | Base WebSocket URL of the CSMS under test. |
| `--max-parallel` | Maximum concurrently executing stories. |
| `--shard N/M` | Run one shard of the selected story set in CI. |
| `--ocpp-version` | Restrict execution to a declared OCPP version, such as `1.6`. |
| `--lock-timeout` | Maximum wait for a cache lock; default is `60s`. |
| `--no-wait` | Fail immediately if a cache lock is busy. |
| `--insecure-skip-verify` | Disable TLS certificate verification. |
| `--param name=value` | Override a story parameter; may be repeated. |
| `--report-dir` | Directory where per-run report subdirectories are written. |
| `--no-trace-on-pass` | Omit wire traces from reports for passing stories. |

Run `./bin/octane help <command>` for the authoritative command help.

## Story Format

Stories are Gherkin-flavored text files with metadata, setup, scenario steps,
dependencies, and teardown. This is the shape used by the checked-in suite:

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

The parser and validator live under `pkg/story`. Dependency ordering is handled
by `pkg/runner`; dependency graph details are documented in
[docs/concepts/dependency-graph.md](./docs/concepts/dependency-graph.md).

## OCPP 1.6 Coverage

The checked-in `scenarios/v16` suite exercises these OCPP 1.6 flows:

| Area | Stories |
|---|---|
| Connection and boot | `station_connection_established`, `station_boot_accepted`, `boot_sequence_accepted` |
| Status and heartbeat | `connector_status_available`, `meter_values_periodic_accepted` |
| Local transactions | `transaction_pluginfirst_accepted`, `transaction_identificationfirst_accepted`, `transaction_identificationfirst_connection_timeout_available`, `transaction_stop_accepted`, `transaction_evside_disconnect_true_true` |
| Remote transactions | `transaction_remotestart_accepted`, `transaction_remotestop_accepted` |
| Connector control | `connector_unlock_accepted`, `connector_unlock_failed`, `connector_availability_operative`, `connector_availability_inoperative` |
| Station control | `station_reset_soft_accepted`, `station_reset_hard_accepted` |
| Configuration and cache | `configuration_get_accepted`, `configuration_change_accepted`, `cache_clear_accepted` |
| Reservations | `connector_reservation_faulted`, `connector_cancelreservation_accepted` |

Use `./bin/octane keywords list` to see the current registered keyword catalog.

## Repository Layout

| Path | Purpose |
|---|---|
| `cmd/octane` | Cobra CLI, command wiring, config loading, exit behavior. |
| `pkg/story` | `.story` lexer, parser, AST, diagnostics, and parser tests. |
| `pkg/runner` | Story execution, dependency traversal, sharding, skip logic, cache integration. |
| `pkg/keywords` | Primitive, lifecycle, OCPP 1.6, API, and registry packages. |
| `pkg/transport` | WebSocket dialing and station connection management. |
| `pkg/wire` | OCPP-J frame parsing, encoding, and coercion. |
| `pkg/cache` | Content-addressed result cache and file locking. |
| `pkg/report` | Report model plus JSON and Robot XML writers. |
| `pkg/engine` | Deterministic clock and random primitives for repeatable runs. |
| `scenarios/v16` | Checked-in OCPP 1.6 conformance stories. |
| `docs` | Configuration, CLI, concepts, ADRs, manpage sources, and integration docs. |
| `examples` | Consumer examples, including CI snippets and a primitive story. |
| `action` | Docker-based GitHub Action wrapper metadata and entrypoint. |
| `website` | Docusaurus documentation site. |
| `test/integration` | Integration tests for runner, cache, reports, TLS, and keyword behavior. |
| `test/reference` | Docker Compose reference target assets for CitrineOS-oriented testing. |
| `packaging` | `nfpm` packaging configuration. |
| `scripts` | Manpage, completion, and CI helper scripts. |

## Development

Common local commands:

```bash
make format
make lint
make test
make build
```

Additional procedures:

```bash
make test-reference   # run reference-tagged integration tests with Docker Compose
make fuzz             # run fuzz targets for 30 seconds each
make man              # regenerate man pages
make completions      # regenerate shell completions
make docs-html        # build the Docusaurus site
make docs-serve       # serve docs on http://127.0.0.1:3000
make package          # snapshot release artifacts through goreleaser
make clean            # remove generated build artifacts
```

The CI workflows currently live in `.github/workflows/ci.yml`,
`.github/workflows/docs.yml`, and `.github/workflows/release.yml`.

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Command succeeded; for `run`, all selected stories passed or were satisfied from cache. |
| `1` | One or more stories failed. |
| `64` | Configuration, flag, or story validation error. |
| `74` | I/O error. |
| `125` | Internal error. |

## License

Apache-2.0. See [LICENSE](./LICENSE).

OCTANE is not affiliated with or endorsed by the Open Charge Alliance, LF
Energy, or CitrineOS.
