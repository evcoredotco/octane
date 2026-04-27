# Getting Started with OCTANE

This guide takes a CSMS developer from zero to a first conformance run in
under five minutes. It assumes you have a CSMS listening on a WebSocket
endpoint and want to verify its OCPP wire behavior.

## Prerequisites

- **Go 1.23 or later** — required for the `go install` path.
- Or **Docker** — the GHCR image includes the binary; no Go toolchain needed.
- A reachable CSMS WebSocket endpoint (local or remote).

## Install

### Option A: Go toolchain

```bash
go install github.com/evcoreco/octane/cmd/octane@latest
```

The binary is placed in `$GOPATH/bin/octane` (or `$HOME/go/bin/octane`).

### Option B: Docker

```bash
docker pull ghcr.io/evcoreco/octane:latest
```

Use the image in place of the binary:

```bash
docker run --rm ghcr.io/evcoreco/octane:latest octane --version
```

### Option C: Build from source

```bash
git clone https://github.com/evcoreco/octane
cd octane
make build          # binary lands at ./bin/octane
```

## Point OCTANE at your CSMS

Create a minimal `octane.yml` in the directory where you will run OCTANE:

```yaml
schema_version: "1"
csms:
  url: ws://localhost:8080/ocpp/CP001
  ocpp_version: "1.6"
  subprotocol: ocpp1.6
```

`csms.url` is the only required field. All other settings have built-in
defaults. For the full configuration reference see
[`docs/configuration.md`](./configuration.md).

## Run the v16 suite

```bash
octane run scenarios/v16/
```

OCTANE discovers every `.story` file under `scenarios/v16/`, builds the
dependency graph, and executes stories against the endpoint declared in
`octane.yml`. Stories that depend on others wait for their prerequisites
to pass first.

To run a single story:

```bash
octane run scenarios/v16/boot_notification_accepted.story
```

To dry-validate story files without executing them:

```bash
octane validate stories scenarios/v16/
```

## Read the output

A successful run prints one summary line to stdout:

```
passed=12 failed=0 skipped=0 cache-hits=8
```

| Field | Meaning |
|---|---|
| `passed` | Stories that completed and all assertions held. |
| `failed` | Stories where at least one assertion failed or the run errored. |
| `skipped` | Stories skipped because a prerequisite failed. |
| `cache-hits` | Stories satisfied from the content-addressed result cache. |

When one or more stories fail, OCTANE exits with code 1 and the failed
stories are listed above the summary line. Exit code 0 means all stories
passed (or were satisfied by the cache).

## Next steps

- **Configuration reference** — every flag, every env var, every
  `octane.yml` field: [`docs/configuration.md`](./configuration.md)
- **Full CLI reference** — all subcommands and flags:
  [`docs/cli-reference.md`](./cli-reference.md)
- **CI integration examples** — ready-to-use GitHub Actions and GitLab CI
  pipelines: [`examples/ci/`](../examples/ci/)
- **Scenario authoring** — how to write `.story` files:
  [`docs/concepts/dependency-graph.md`](./concepts/dependency-graph.md)
