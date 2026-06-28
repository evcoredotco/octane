---
sidebar_position: 2
---

# Getting Started

This guide takes you from a fresh clone to a passing conformance run in
about five minutes. You will build the CLI, point it at a CSMS, run a
single story, then run the whole suite.

## Prerequisites

| You need | Why |
|---|---|
| **Go 1.26+** | OCTANE builds from source; there are no published packages yet. |
| **A reachable CSMS** | The system under test, speaking OCPP-J 1.6 over WebSocket. |
| **Docker** (optional) | The quickest way to stand up the reference CSMS, CitrineOS. |

:::info Pre-alpha
OCTANE is pre-alpha and ships only as source today. See
[Installation](./installation.md) for the full build, shell-completion,
and man-page details.
:::

## 1. Build the CLI

```bash
git clone https://github.com/evcoreco/octane
cd octane
go build ./cmd/octane      # produces ./octane in the repo root
./octane --help
```

You can also run it without building a binary:

```bash
go run ./cmd/octane run --csms-endpoint ws://localhost:9210
```

## 2. Stand up a CSMS

OCTANE tests an existing CSMS — it does not provide one. During
development the reference target is [CitrineOS](https://citrineos.github.io/).
Its OCPP-J WebSocket listens on port **9210** (the REST/admin API on 8080
is not used by OCTANE).

```bash
# From the OCTANE repo — brings up the pinned reference CSMS.
cd test/reference
docker compose up -d --wait
```

The base WebSocket URL is the endpoint **without** a station path —
OCTANE appends `/CP01`, `/CP02`, … per simulated station automatically.

## 3. Create `octane.yml`

Drop an `octane.yml` in the directory you run OCTANE from. The keys are
**camelCase**:

```yaml
storiesDir: scenarios/v16     # where .story files are discovered
ocppVersion: "1.6"            # empty means "all versions"
maxParallel: 1
```

:::note The endpoint is not in `octane.yml`
The CSMS endpoint typically differs between environments, so it is passed
on the command line with `--csms-endpoint` (or via the GitHub Action),
not stored in the config file.
:::

See the [configuration reference](./reference/config-schema.md) for every
field, its default, and the matching environment variable.

## 4. Run your first story

```bash
./octane run scenarios/v16/station_boot_accepted.story \
    --csms-endpoint ws://localhost:9210
```

OCTANE establishes the WebSocket connection, identifies itself as the
station declared in the story, drives the OCPP exchange, and asserts the
CSMS response at each step.

## 5. Run the full suite

```bash
./octane run --csms-endpoint ws://localhost:9210
```

With no story paths given, OCTANE discovers every `.story` file under
`storiesDir`, resolves the [dependency graph](./concepts/dependency-graph.md)
into topological order, and executes it — skipping anything the cache
already knows passed.

## 6. Read the result summary

On completion OCTANE prints a one-line summary and the report location:

```text
passed=12 failed=0 skipped=0 cache-hits=3
report-dir=reports/run-20260628-1/
```

| Field | Meaning |
|---|---|
| `passed` | Stories that ran and whose assertions all held. |
| `failed` | Stories that ran with at least one failed assertion. |
| `skipped` | Stories that did not run because a prerequisite failed. |
| `cache-hits` | Stories skipped because a cached result was still valid. |

The process [exit code](./reference/exit-codes.md) is `0` when everything
passed and `1` when any story failed — exactly what a CI gate needs. Each
run writes a deterministic `report.json` and a Robot Framework
`output.xml` under `reports/<run-id>/`; see [Reports](./operations/reports.md).

## Where to go next

- **[Author your first story](./authoring/first-story.md)** — write a
  conformance scenario from scratch.
- **[Stories](./concepts/stories.md)** — the anatomy of a `.story` file.
- **[CI integration](./operations/ci-integration.md)** — gate pull
  requests on conformance with GitHub Actions or GitLab CI.
- **[CLI reference](./reference/cli.md)** — every subcommand and flag.
