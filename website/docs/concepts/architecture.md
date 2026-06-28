---
sidebar_position: 2
---

# Architecture

OCTANE is organized into three layers. Each layer is the contract for the
one above it, and each is deliberately ignorant of the layers below:
**stories never import Go, keywords never know which CSMS they are talking
to, and the engine never knows which OCPP version it is driving.** That
separation is what makes the system composable and the conformance signal
trustworthy.

```text
┌────────────────────────────────────────────────────────────┐
│  Layer 1 — Stories (.story files)                            │
│  Declarative, Gherkin-flavored scenarios. One per OCPP       │
│  behavior under test. Version-controlled, reviewer-readable, │
│  spec-traceable via Spec-Ref.                                │
└────────────────────────────────────────────────────────────┘
                    ▼  resolves each step to a keyword
┌────────────────────────────────────────────────────────────┐
│  Layer 2 — Keyword library (Go)                              │
│  Two sub-layers, resolved domain → primitive:                │
│    • Domain     — OCPP 1.6 message semantics                 │
│    • Primitive  — raw transport escape hatch                 │
│  No per-CSMS override layer.                                 │
└────────────────────────────────────────────────────────────┘
                    ▼  drives the wire via
┌────────────────────────────────────────────────────────────┐
│  Layer 3 — Engine                                            │
│  WebSocket transport, OCPP-J framing, DAG runner + worker    │
│  pool, content-addressed cache, deterministic clock & RNG.   │
└────────────────────────────────────────────────────────────┘
                    ▼  emits
        report.json (native, byte-deterministic)
        output.xml  (Robot Framework 7.x schema)
```

## Layer 1 — Stories

A [`.story` file](./stories.md) declares one scenario in a small
Gherkin-flavored DSL. It is the artifact a certification reviewer reads
and the artifact the runner executes — there is no separate "test code"
behind it. Stories carry OCPP traceability (`Spec-Ref`), declare how many
stations they need, and list their prerequisites. They never contain Go.

## Layer 2 — The keyword library

Each `Given` / `When` / `Then` / `And` step in a story is matched to a
**keyword**: a typed Go function that knows how to perform one wire action
— send an OCPP message, wait for a response, and assert its fields.

Keywords self-register at process start and are organized into two
sub-layers:

| Sub-layer | Scope | Example |
|---|---|---|
| **Domain** | OCPP 1.6 message semantics | `station {station:string} sends BootNotification with reason {reason:string}` |
| **Primitive** | Raw transport, an escape hatch | `send raw frame {frame:any} on station {station:string}` |

**Resolution order.** For a given step, OCTANE tries domain keywords
scoped to the story's OCPP version first; the first match wins. If none
match, it falls back to primitive keywords. A step that matches neither
layer fails preflight with a diagnostic listing the layers searched and
the closest registered patterns (by edit distance). Pattern collisions
within the same `(layer, version)` tuple panic at startup — they are
caught in CI, never at runtime.

There is intentionally **no third "profile" layer**. Domain keywords are
identical for every CSMS implementing a given OCPP version; see
[wire conformance](./wire-conformance.md).

See the [keywords reference](../authoring/keywords-reference.md) and the
[keyword catalog](../reference/keyword-catalog.md) for the full vocabulary.

## Layer 3 — The engine

The engine is plain Go that turns keyword calls into wire traffic and
results:

- **Transport** — manages the OCPP-J WebSocket connection per station.
- **Wire framing** — encodes and decodes OCPP-J envelopes (Call,
  CallResult, CallError).
- **Runner** — resolves the [dependency DAG](./dependency-graph.md),
  executes stories in topological order using a configurable worker pool,
  and consults the content-addressed cache.
- **Determinism** — a single injected `Clock` and seeded `Rand` mean the
  same inputs produce the same outputs. Engine code never calls
  `time.Now()` or unseeded randomness directly.
- **Reports** — builds one in-memory report tree and serializes it twice.

## The Robot Framework metaphor (without the runtime)

OCTANE borrows Robot Framework's *structural* mental model — declarative
scenarios, a layered keyword library, machine-readable output — and emits
Robot's `output.xml` for ecosystem compatibility. It does **not** run
Robot Framework. There is no Python interpreter in the loop; OCTANE is a
single Go binary, which keeps CI light and distribution simple.

| Robot Framework | OCTANE |
|---|---|
| `@keyword("Pattern ${arg}")` | `registry.Register(api.Keyword{Pattern: "..."})` |
| Library file (one layer) | Domain layer / primitive layer |
| `types={"count": int}` | `{name:int}` placeholders |
| `raise AssertionError` | `return fmt.Errorf(...)` |
| `output.xml` | `output.xml` (same schema) |

## Reports fall out of one tree

Every run produces both `report.json` (OCTANE-native, byte-deterministic,
the source of truth) and `output.xml` (Robot Framework 7.x schema) from
the *same* in-memory report tree, so the two always agree. See
[Reports](../operations/reports.md).

## Next

- **[Stories](./stories.md)** — the anatomy of Layer 1.
- **[Dependency graph & caching](./dependency-graph.md)** — how the runner
  orders and skips work.
- **[Keywords reference](../authoring/keywords-reference.md)** — Layer 2 in
  detail.
