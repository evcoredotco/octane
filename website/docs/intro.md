---
sidebar_position: 1
---

# Introduction

**OCTANE** — *OCPP Conformance Testing & Network Evaluation* — is an
open-source conformance harness for **OCPP 1.6J** Charging Station
Management Systems (CSMS).

OCTANE impersonates one or more charging stations over the OCPP-J
WebSocket protocol, drives your CSMS through real message exchanges, and
asserts that it responds exactly as the specification requires. It tests
your system **from the charging-station side, at the wire** — the CSMS
under test needs no modification, no admin API, and no sidecar service.

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210
```

```text
passed=21 failed=0 skipped=0 cache-hits=6
report-dir=reports/run-20260628-1/
```

## What OCTANE does

- **Drives an unmodified CSMS.** OCTANE speaks OCPP-J natively from the
  station side. It dials your endpoint, exchanges messages, and checks
  the bytes that come back.
- **Runs declarative `.story` files.** Each story is a self-contained
  scenario written in a Gherkin-flavored DSL that reads like English and
  traces to a section of the OCPP specification.
- **Resolves a dependency graph.** Stories declare prerequisites; the
  runner builds a DAG, executes prerequisites first, and a
  content-addressed cache skips anything unchanged since the last run.
- **Emits deterministic reports.** Every run produces a byte-stable
  `report.json` and a Robot Framework `output.xml`, ready to upload as CI
  artifacts or feed into a test-management dashboard.
- **Ships two surfaces, one engine.** The `octane` CLI and the
  `octane-action` GitHub Action wrap the same binary, so anything you can
  do locally you can also gate in CI.

## What OCTANE does not do

- It does **not** inspect CSMS-internal behavior that is invisible on the
  OCPP wire — databases, billing pipelines, audit logs, internal state.
  If the CSMS sends the right bytes in the right order, the story passes.
- It does **not** accept per-CSMS tolerances or behavioral overrides. A
  CSMS deviation is a *finding*, not a configurable exception.
- It does **not** issue or imply formal conformance certification.
  Certification by an external authority is a separate process; OCTANE
  asserts only that observable wire behavior matches what the OCPP
  specification requires for the scenarios it exercises.
- It does **not** implement OCPP 2.0.1 or 2.1. OCTANE targets **OCPP
  1.6J** exclusively.

:::info Project status
OCTANE is **pre-alpha**. The core engine is implemented and the OCPP 1.6
keyword layer covers 17 message types across 21 stories. There are no
published packages yet — you [build from source](./installation.md).
:::

## Where to start

- **[Getting started](./getting-started.md)** — go from clone to a green
  run in about five minutes.
- **[Installation](./installation.md)** — build the CLI, enable shell
  completion, read the man pages.
- **[Wire conformance](./concepts/wire-conformance.md)** — the philosophy
  that shapes everything else.
- **[Architecture](./concepts/architecture.md)** — the three-layer model
  (stories → keywords → engine).
- **[Authoring your first story](./authoring/first-story.md)** — write
  and run a conformance scenario from scratch.

For terminal users with the binary installed:

```text
man 7 octane          # concepts overview
man 1 octane-run      # run-command reference
man 5 octane.story    # story DSL reference
man 5 octane.yml      # config file reference
```
