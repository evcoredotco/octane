---
sidebar_position: 1
---

# Introduction

**OCTANE** — *OCPP Conformance Testing & Network Evaluation* — is an
open-source conformance harness for OCPP 1.6J, 2.0.1, and 2.1
Charging Station Management Systems (CSMS).

## What OCTANE does

OCTANE drives a CSMS by impersonating one or more charging stations
over OCPP-J WebSocket connections. It runs `.story` files written in
a Gherkin-flavored DSL, observes the CSMS's wire responses, and
produces a deterministic conformance report.

OCTANE is designed to:

- Run against an **unmodified CSMS**. No code changes, no sidecar
  service, no privileged admin API.
- Gate every commit in CI. The binary distribution ships as `.deb`,
  `.rpm`, Homebrew, and Docker, and runs in seconds.
- Trace every conformance test to the OCPP specification section it
  exercises.
- Produce reports consumable by Robot Framework dashboards (Allure,
  ReportPortal, Jenkins, GitLab, GitHub Actions).

## What OCTANE does not do

- It does **not** verify CSMS-internal behavior that is invisible on
  the OCPP wire — audit-log content, billing pipelines, internal
  state transitions. Those scenarios are documented as
  *operator-assisted* and are out of OCTANE's automated scope.
- It does **not** issue or imply formal conformance certification.
  Certification by an external authority is a separate process with
  its own scope and rules; OCTANE asserts only that observable wire
  behavior matches what the OCPP specifications require for the
  scenarios it exercises.

## Where to start

- [Install OCTANE](./installation.md)
- [Getting started in 5 minutes](./getting-started.md)
- [Wire-conformance concept](./concepts/wire-conformance.md)
- [Authoring your first story](./authoring/first-story.md)

For terminal users with the binary installed:

```
man 7 octane          # concepts overview
man 1 octane-run      # run-command reference
man 5 octane.story    # story DSL reference
man 5 octane.yml      # config file reference
```
