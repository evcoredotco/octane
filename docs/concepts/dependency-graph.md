# Dependency Graph

OCTANE runs `.story` files in dependency order. A story can declare that it
requires one or more prerequisite stories to have passed before it executes.
The runner builds a directed acyclic graph (DAG) from those declarations,
computes a stable topological execution order, and drives the worker pool
accordingly.

## Why it exists

OCPP conformance tests rarely stand alone. A reservation test, for instance,
only makes sense after a charge point has successfully booted and its
connector reported `Available`. Rather than duplicating boot logic in every
test, you write a helper story that establishes the known state, and declare
the conformance test as a dependent of that helper.

When a prerequisite fails, the runner marks every downstream dependent as
`skipped` and records the failing prerequisite in the `Cause` field of each
skipped result. This avoids false failures and keeps the report readable.

## The `Depends:` block

In a story's Meta section, declare prerequisites with a `Depends:` block:

```yaml
Meta:
  Id: connector_reservation_faulted
  Depends:
    - id: connector_status_available
      scope: per-station
    - id: station_boot_accepted
      scope: per-run
```

Each entry has two fields:

| Field   | Required | Description                               |
|---------|----------|-------------------------------------------|
| `id`    | yes      | The `Id` value of the prerequisite story  |
| `scope` | yes      | One of `per-station`, `per-run`, `global` |

## Scope types

### `per-station`

The prerequisite executes once for each station handle in the run. If a run
covers three charge points (`CP01`, `CP02`, `CP03`), the runner creates three
independent instances of the prerequisite — one per station — and connects
each conformance test instance to its matching prerequisite instance.

Use `per-station` for any setup that is station-specific: boot sequences,
connector status polls, authorization flows.

### `per-run`

The prerequisite executes once for the entire run, regardless of how many
stations are involved. All conformance tests that declare a `per-run`
dependency share that single prerequisite result.

Use `per-run` for setup that affects all stations simultaneously: global
configuration pushes, network-level authentication.

### `global`

Like `per-run`, but the result is also reused across separate `octane run`
invocations via the cache. The scope key for cache keying is the empty string,
so any run against the same CSMS endpoint and OCPP version shares the cached
outcome.

Use `global` sparingly. It is appropriate for one-time provisioning steps
that are genuinely idempotent and whose results do not change between runs.

## DAG construction

1. Each `(story, scope-key)` pair becomes one DAG node. A story with
   `per-station` scope and `Stations: 3` expands into three nodes:
   `story_id/CP01`, `story_id/CP02`, `story_id/CP03`.
2. Edges are added from each prerequisite node to each dependent node.
3. The runner computes a topological order. Within a topological level, nodes
   are ordered lexicographically by node ID for deterministic dispatch.

## Cycle detection

If the `Depends:` declarations form a cycle, `runner.Run` returns
`runner.ErrCycle`. The error message names the story IDs involved:

```txt
runner: dependency cycle detected: a → b → a
```

No stories execute when a cycle is detected.

## Station count and `per-station` scope

The `Stations: N` Meta key on a story controls how many per-station nodes the
runner creates for that story. When a conformance story declares `Stations: 2`
and depends on a helper with `per-station` scope, the runner instantiates
`helper/CP01` and `helper/CP02` as prerequisites for `test/CP01` and
`test/CP02` respectively.

## Examples

Helper story (no `Spec-Ref`, tagged `helper`):

```yaml
Meta:
  Id: station_boot_accepted
  Tags: [helper, ocpp1.6]
  Stations: 1
```

Conformance story depending on the helper per station:

```yaml
Meta:
  Id: connector_reservation_faulted
  Spec-Ref: "OCPP 1.6J -7.3"
  Tags: [ocpp1.6]
  Stations: 2
  Depends:
    - id: station_boot_accepted
      scope: per-station
```

With `Stations: 2`, the runner creates nodes `connector_reservation_faulted/CP01`
and `connector_reservation_faulted/CP02`, each depending on its matching
`station_boot_accepted/CP01` and `station_boot_accepted/CP02`.
