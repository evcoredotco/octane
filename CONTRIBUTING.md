# Contributing to OCTANE

Thank you for considering a contribution to OCTANE. This document
covers what you need to know to author conformance scenarios,
keyword library entries, and supporting code in line with the
project's conventions.

> **Read first:** [`.specify/memory/constitution.md`](./.specify/memory/constitution.md).
> Every contribution must comply with the constitution; the
> conventions in this file operationalize it. If anything in this
> file conflicts with the constitution, the constitution wins.

## Table of contents

- [Spec-driven development](#spec-driven-development)
- [Authoring conformance stories](#authoring-conformance-stories)
- [Authoring helper stories](#authoring-helper-stories)
- [Adding keywords](#adding-keywords)
- [Keyword author guide](#keyword-author-guide)
- [Code style](#code-style)
- [Commits and PRs](#commits-and-prs)

## Spec-driven development

OCTANE follows a strict spec-driven workflow (constitution principle
VIII). Code does not land before its spec merges:

1. `/specify <feature>` — draft `specs/NNN-feature/spec.md`
2. `/plan` — fill `plan.md` with technical approach + ADR drafts
3. `/tasks` — decompose into atomic, agent-assignable tasks
4. `/implement T-NNN-MM` — execute one task

For trivial fixes (typos, comment improvements, single-line bug
fixes), open a PR directly without a spec. The reviewer will tell
you if a spec is needed.

## Authoring conformance stories

OCTANE conformance stories are independent original work derived
from the published OCPP specifications. The rules in this section
operationalize ADR 0014 (IP and authoring guidelines) and apply
without exception.

### Source of truth

Author from the OCPP specification document, not from any
third-party test catalog or testing tool's documentation. The
specification is the public, canonical description of what
conformant CSMS behavior looks like; that is what OCTANE tests.

In practice:

- Open the relevant OCPP spec PDF (1.6J, 2.0.1, or 2.1).
- Locate the section describing the message or behavior you want
  to test.
- Read the normative text — request schema, response schema,
  state-machine transitions, error conditions.
- Write a story whose `Spec-Ref` Meta key cites that section, and
  whose assertions express what the specification requires.
- Write the prose narrative (the comment block at the top of the
  story file) in your own words. Describe what the test does, why
  it matters for conformance, and what state it assumes.

If you find yourself reaching for a third-party catalog to
"translate" its description into OCTANE form, stop. That is the
exact pattern ADR 0014 forbids. Go back to the OCPP specification.

### Naming convention

Story filenames and IDs follow `<resource>_<function>_<desire>`
(snake_case lowercase). Examples:

| Filename | What it tests |
|----------|---------------|
| `boot_notification_accepted.story` | Successful boot registration |
| `boot_notification_malformed.story` | Wire-level rejection of malformed boot |
| `connector_reservation_faulted.story` | CSMS handles a Faulted reservation response |
| `authorize_concurrent_rejected.story` | Concurrent authorize attempt rejected |

The `desire` slot prefers a specific protocol-level state when one
applies (`faulted`, `concurrenttx`, `accepted`) over a generic
outcome category (`success`, `failure`).

### Required Meta keys for conformance stories

```
Meta
    Name:        <human-readable name in your own words>
    Id:          <snake_case slug matching filename>
    Spec-Ref:    OCPP <version> §<section> <message-or-behavior>
    Tags:        <comma list, must include one of:
                  wire-only | multi-station | operator-assisted>
    Stations:    <integer >= 1>
    Timeout:     <duration; optional, default from config>
    Parameters:  <comma list of names referenced in steps; optional>
    Depends:     <YAML list of prereq IDs; optional>
```

`Spec-Ref` MUST cite the OCPP specification, not a third-party
testing tool. The format is one of:

- `OCPP 2.0.1 §B01 BootNotification`
- `OCPP-J 1.6 §6.40 ReserveNow`
- `OCPP 2.1 §C01 Authorize`

### Required prose comment block

The first non-blank lines of every story file are a `#`-prefixed
narrative explaining what the test does and what it depends on.
This is the equivalent of a function docstring. Write it in your
own words. Do not copy from any third-party source.

Example:

```
# Validates that a CSMS implementing OCPP 2.0.1 §B01 BootNotification
# replies to a well-formed BootNotification.req with a
# BootNotificationResponse carrying status "Accepted" and a
# heartbeatInterval within the spec-permitted range.
#
# Single-station, wire-only conformance test. Depends on a
# successful WebSocket connection but assumes no prior CSMS state.
```

## Authoring helper stories

Helper stories exist to bring the system to a known state so that
downstream conformance tests can run from a defined starting point.

Differences from conformance stories:

- **No `Spec-Ref`** — helpers do not assert conformance to a
  specification section in their own right.
- **Tag `helper`** — required.
- **Filename matches the ID** — kebab-case snake_case as before.
- **Lives alongside conformance stories** — under
  `scenarios/v16/`, `scenarios/v201/`, etc. (no separate
  `helpers/` directory).

The parser enforces the distinction: a story tagged `helper` MUST
omit `Spec-Ref`, and a story not tagged `helper` MUST include it.

## Adding keywords

Keywords are typed Go functions that map step text to wire actions.
They live under `pkg/keywords/`:

- `pkg/keywords/api/`         — public surface (do not modify lightly)
- `pkg/keywords/registry/`    — self-registration mechanism
- `pkg/keywords/primitive/`   — transport-level escape hatches
- `pkg/keywords/domain/v16/`  — OCPP 1.6 keywords
- `pkg/keywords/domain/v201/` — OCPP 2.0.1 keywords
- `pkg/keywords/domain/v21/`  — OCPP 2.1 keywords

Each keyword registers exactly one pattern. Domain keywords are
identical for every CSMS implementing the OCPP version they target;
there is no per-CSMS override layer (constitution principle XII).

When adding a domain keyword, include:

1. The pattern (with `{name:type}` placeholders).
2. The implementation Go function.
3. A black-box test in the same package's `_test.go` file
   asserting the pattern is registered and the function exhibits
   the documented behavior on representative inputs.

## Keyword author guide

This section is for contributors writing Go functions that implement
story steps. It covers what a keyword is, how to author one, how to
register it, and how to write a unit test for it without a network.

### What is a keyword and when should you write one

A keyword is a Go function with the signature:

```go
func(ctx context.Context, state api.State, args api.Args) error
```

It maps one sentence in a `.story` file to one unit of wire behavior.
A non-nil return marks the step failed; the error message becomes the
finding text in the run report.

There are two layers. Write a **primitive** keyword when the operation
is transport-level and OCPP-version-agnostic (for example, opening a
WebSocket or sending a raw frame). Write a **domain** keyword when the
operation encodes OCPP semantics for a specific version (for example,
sending a BootNotification CALL and validating the CALLRESULT).

Domain keywords live under `pkg/keywords/domain/v16/`, `v201/`, or
`v21/` depending on the OCPP version. Primitive keywords live under
`pkg/keywords/primitive/`. Do not write domain keywords that vary by
CSMS: OCTANE has no per-CSMS override layer (constitution principle
XII). If a CSMS deviates from the spec, that deviation surfaces as a
finding — not as a different keyword.

### The `api.Func` signature

```go
import (
    "context"
    "github.com/evcoreco/octane/pkg/keywords/api"
)

func myKeyword(ctx context.Context, state api.State, args api.Args) error {
    // ... drive the wire ...
    return nil
}
```

- `ctx` carries the per-step timeout and cancellation. Pass it to
  every blocking call (`station.Send`, `station.Expect`).
- `state` provides station lookup, a deterministic clock, and
  structured logging. Use `state.Now()` — never `time.Now()` — so
  that reports are byte-identical across runs (constitution
  principle IV). The linter rejects `time.Now()` in `pkg/keywords/`.
- `args` holds the named parameter values extracted from the step
  text. Use typed accessors to retrieve them.

### How `api.Args` accessors work (and why they panic)

The resolver binds pattern placeholders to named Go values before your
function runs. Retrieve them with:

```go
station := args.Station("station")   // {station:station} → string
count   := args.Int("count")         // {count:int}       → int
label   := args.String("label")      // {label:string}    → string
timeout := args.Duration("timeout")  // {timeout:duration}→ time.Duration
ratio   := args.Float("ratio")       // {ratio:float}     → float64
enabled := args.Bool("enabled")      // {enabled:bool}    → bool
raw     := args.Any("payload")       // {payload:any}     → any
```

Every accessor **panics** when the requested key is absent or has the
wrong type. This is intentional. The registry validates that every
`{name:type}` placeholder declared in a pattern has a corresponding
accessor call in the keyword body at `init()` time (static analysis,
not reflection). A runtime panic means a registry bug, not an authoring
bug. Do not wrap accessor calls in `recover()` — fix the pattern or the
registration instead.

### Registering a keyword with `registry.Register`

Register the keyword from your package's `init()` function:

```go
package boot

import (
    "context"

    "github.com/evcoreco/octane/pkg/keywords/api"
    "github.com/evcoreco/octane/pkg/keywords/registry"
)

func init() {
    registry.Register(api.Keyword{
        Pattern:     "station {station:station} sends BootNotification with reason {reason:string}",
        Layer:       api.LayerDomain,
        OCPPVersion: api.OCPP201,
        Func:        sendBootNotification,
    })
}

func sendBootNotification(
    ctx context.Context,
    state api.State,
    args api.Args,
) error {
    handle := args.Station("station")
    reason := args.String("reason")

    station, err := state.Station(handle)
    if err != nil {
        return fmt.Errorf("station %q not available: %w", handle, err)
    }

    // build and send the CALL frame, then await the CALLRESULT ...
    _ = reason
    _ = station
    return nil
}
```

Registration rules:

- Each keyword registers exactly one pattern.
- Two keywords with the same `(Layer, OCPPVersion, Pattern)` triple
  cause a panic at startup naming both registration sites.
- Domain-layer keywords must set `OCPPVersion` to a non-zero value.
- Primitive-layer keywords leave `OCPPVersion` as the zero value.

### Pattern placeholder syntax

Patterns use `{name:type}` placeholders. Supported types:

| Type       | Go type         | Accepts |
|------------|-----------------|---------|
| `string`   | `string`        | any whitespace-delimited token |
| `int`      | `int`           | base-10 integer |
| `float`    | `float64`       | decimal number |
| `bool`     | `bool`          | `true` or `false` (case-insensitive) |
| `duration` | `time.Duration` | Go duration string (`30s`, `1m30s`) |
| `station`  | `string`        | station handle (semantically validated) |
| `any`      | `string`        | any token, stored as raw string |

Each placeholder name must be unique within a pattern. Optional
parameters are not supported; use two separate patterns pointing to the
same `Func` if both short and long forms are needed.

### Testing a keyword with `mock.NewMockState()` and `mock.NewMockStation()`

Keyword unit tests must not import `pkg/runner/`, `pkg/transport/`, or
any network library. Use the test doubles in
`pkg/keywords/api/mock` instead:

```go
package boot_test

import (
    "context"
    "testing"
    "time"

    "github.com/evcoreco/octane/pkg/keywords/api"
    "github.com/evcoreco/octane/pkg/keywords/api/mock"
)

func TestSendBootNotification_Accepted(t *testing.T) {
    state   := mock.NewMockState()
    station := mock.NewMockStation()

    state.RegisterStation("CP01", station)
    state.SetNow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

    // Queue the CALLRESULT the CSMS will return.
    station.QueueFrame([]any{
        3, "msg-001",
        map[string]any{
            "currentTime": "2024-01-01T00:00:00Z",
            "interval":    float64(300),
            "status":      "Accepted",
        },
    })

    args := api.NewArgs(map[string]any{
        "station": "CP01",
        "reason":  "PowerUp",
    })

    err := sendBootNotification(
        context.Background(), state, args,
    )
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    sent := station.SentFrames()
    if len(sent) != 1 {
        t.Fatalf("expected 1 sent frame, got %d", len(sent))
    }
}
```

`mock.State` methods:

| Method | Purpose |
|--------|---------|
| `NewMockState()` | Returns a ready-to-use `*mock.State` |
| `RegisterStation(handle, station)` | Associates a handle with a `*mock.Station` |
| `SetNow(t time.Time)` | Sets the frozen clock returned by `Now()` |
| `Logs() []string` | Returns all messages passed to `Logf` |

`mock.Station` methods:

| Method | Purpose |
|--------|---------|
| `NewMockStation()` | Returns a ready-to-use `*mock.Station` |
| `QueueFrame(frame []any)` | Pre-queues a frame for the next `Expect` call |
| `SentFrames() [][]any` | Returns all frames recorded by `Send` |
| `SetSendError(err)` | Makes all subsequent `Send` calls return `err` |
| `SetExpectError(err)` | Makes all subsequent `Expect` calls return `err` |

### Minimal end-to-end example

A runnable standalone demonstration lives at
`pkg/keywords/api/mock/testdata/external/keyword.go`. It exercises
`mock.State` and `mock.Station` without any network dependency and can
be run with:

```bash
go run ./pkg/keywords/api/mock/testdata/external/keyword.go
```

The file carries `//go:build ignore` and is not compiled by
`go test ./...`; it is documentation as code.

---

## Code style

Go code follows the conventions in
[`mnt/skills/user/golang-master/SKILL.md`](https://example.invalid/golang-master)
where applicable: gofmt, line length 80, function complexity ≤ 7,
no `time.Now()` (use the injected clock), no global state outside
the registry.

Run before pushing:

```bash
make format     # gofumpt + goimports
make lint       # golangci-lint
make test       # go test -race ./...
make spec-check # validates spec structure
```

## Commits and PRs

Commits follow the Conventional Commits specification with a JIRA
prefix when applicable, GPG-signed. Use the `qtech-commit` skill
if you have it.

PRs are scoped: one feature, one spec, one ADR, or one fix. PRs
larger than ~600 lines are split.

The reviewer is responsible for verifying:

- Constitution compliance.
- ADR coverage for any new architectural decision.
- IP cleanliness per ADR 0014 (no third-party catalog references).
- Test coverage at appropriate granularity.

Welcome aboard.
