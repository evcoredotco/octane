# ADR 0015: Test Dependency Graph

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** I (Conformance Above
  Convenience), IV (Determinism)

## Context

OCPP scenarios are not independent. A reservation test cannot run
against a CSMS that has not registered the station, and a station
cannot register without first establishing an OCPP-J WebSocket
connection. Most scenarios require a chain of prior state to reach
the point where the actual conformance assertion can be exercised.

Three obvious approaches to this problem all have problems:

1. **Each story carries its full prelude inline.** Every reservation
   test repeats the same six setup lines for connection, boot, and
   status. Duplication grows linearly with the suite size and the
   maintenance cost is high.
2. **A hidden auto-prelude controlled by Meta keys.** A flag like
   `Auto-Prelude: standard` invokes a built-in setup sequence. This
   hides what is happening from story readers and makes any
   variation impossible to express.
3. **Resource-file imports (Robot Framework's `*** Settings ***`
   pattern).** Allows reuse but introduces a separate file type
   whose semantics differ from stories. Authors must learn two
   formats.

A fourth approach — borrowed from build systems and pytest fixtures
— is to model the entire test suite as a **directed acyclic graph
of test cases**, where some test cases are referenced as
prerequisites by others. This is the design adopted here.

## Decision

### Every test case can be a prerequisite

There is one artifact type: the `.story` file. A story's role at
runtime is determined by its position in the dependency graph, not
by its location or extension.

A story declares its dependencies through the `Depends:` Meta key:

```text
Meta
    Name:     Connector reservation faulted
    Id:       connector_reservation_faulted
    Spec-Ref: OCPP-J 1.6 -6.40 ReserveNow
    Stations: 1
    Depends:
      - id:    connector_status_available
        scope: per-station
```

The runner walks the dependency chain transitively, executes the
prerequisites in topological order, then executes the story. Each
prerequisite is itself a story that may have its own dependencies.

### Helper stories vs conformance stories

A conformance story carries a `Spec-Ref` Meta key (per ADR 0014)
and asserts conformance to a section of the OCPP specification.

A helper story omits `Spec-Ref` and is tagged `helper`. Helpers
exist purely as dependency targets — they bring the system to a
known state so conformance stories can run from a defined starting
point. Helpers run their own assertions, but those assertions do
not contribute to the conformance summary in the report.

The parser enforces the distinction:

- A story tagged `helper` MUST omit `Spec-Ref`.
- A story not tagged `helper` MUST include `Spec-Ref`.

Both helpers and conformance stories live under `scenarios/` mixed
together (no separate `helpers/` directory, per project decision).

### Dependency syntax

`Depends:` is a YAML list. Each entry is an object with required
`id` and optional `scope`:

```yaml
Depends:
  - id:    station_connection_established
    scope: per-station
  - id:    csms_supports_reservation_feature_profile
    scope: per-run
```

Three scope values are supported:

| Scope                   | Cache key includes                           | Use case                                                           |
|-------------------------|----------------------------------------------|--------------------------------------------------------------------|
| `per-station` (default) | station handle                               | Lifecycle prereqs that vary by station (boot, connect, status)     |
| `per-run`               | run ID (config SHA + seed)                   | Run-level prereqs (CSMS reachability, feature support detection)   |
| `global`                | empty (cache forever within validity window) | CSMS capability checks that don't change between operator sessions |

If `scope` is omitted, the default is `per-station`.

### Resolution algorithm

At preflight (before any wire activity):

1. Parse every story file under `scenarios/`.
2. Build a graph: each story is a node; each `Depends:` entry is a
   directed edge from the dependent to the prerequisite.
3. For the requested run set (the stories selected by `octane run`
   arguments), compute the **transitive closure of prerequisites**.
4. **Topologically sort** the closure. The runner will execute in
   this order: deepest prerequisites first, the user-requested
   stories last.
5. **Detect cycles.** A cycle in the prerequisite graph is a
   programming error and aborts the run with exit 4 (preflight
   failure) before any wire activity happens.
6. **Detect unresolved IDs.** A `Depends:` entry referencing a
   non-existent story ID also aborts at preflight.

### Execution and caching

For each story in topological order:

1. **Compute the cache key.** Per ADR 0016, the key is
   `(test_id, scope_key, csms_endpoint_sha, octane_version,
   ocpp_version, story_content_sha, parameter_sha)` where the
   story_content_sha is computed transitively over the story and
   all its prerequisites' content.
2. **Check the cache.** If a valid cached result exists, reuse it
   and skip execution.
3. **Acquire the cache key lock** (in-process and cross-process,
   per ADR 0016) to prevent two runners from executing the same
   test case in parallel.
4. **Re-check the cache** after acquiring the lock (another runner
   may have completed it while we waited).
5. **Execute the story.** The runner drives the wire, applies
   keywords, accumulates findings.
6. **On success:** write the result to the cache, release the lock,
   continue to the next story in topological order.
7. **On failure:** mark the story failed, write the failure to the
   cache (with appropriate TTL), release the lock, and skip every
   story whose transitive prerequisites include this one.

### Failure propagation

When a prerequisite fails, dependent stories are **skipped**, not
failed. The report distinguishes:

- `passed`     — story ran and all assertions held.
- `failed`     — story ran and at least one assertion failed.
- `skipped`    — story did not run because a prerequisite failed.

A skipped story carries a pointer to the failing prerequisite in its
report entry so an operator can see the chain.

This is the default behavior. There is no `--continue-on-prereq-fail`
escape hatch in v1; if operators discover they need one, that is a
follow-up ADR.

### Multi-station scoping

When a story declares `Stations: 2`, the runner allocates handles
`CP01` and `CP02`. A `per-station`-scoped prerequisite runs once per
station handle (so `station_boot_accepted` runs twice when its
dependent declares two stations). The cache key for each invocation
includes the station handle.

A `per-run`-scoped prerequisite runs once regardless of station
count.

## Consequences

### Positive

- Test isolation is solved without introducing a new artifact
  type. Every story is a story.
- Common preludes are written once and referenced as dependencies.
  A 50-story suite with 6 lines of common prelude ships as a few
  small helper stories instead of 300 duplicated lines.
- The dependency graph is inspectable. `octane deps show
  connector_reservation_faulted` prints the chain before any wire
  activity happens.
- The cache (per ADR 0016) makes re-running the same suite cheap.
  A helper that ran successfully five minutes ago is reused, not
  re-executed, when its content has not changed.
- Constitution principle IV (determinism) is preserved: the
  topological sort is deterministic given a stable graph; the cache
  key includes every input that could change the result.

### Negative

- Authors must understand the graph model. A story that depends on
  five layers of prerequisites takes longer to reason about than a
  self-contained story.
- A non-conformant CSMS that fails an early prerequisite causes
  large portions of the suite to be skipped. This is the correct
  behavior (you cannot meaningfully test reservation against a CSMS
  that does not accept BootNotifications), but the report needs to
  surface the cascade clearly.

### Neutral

- This is genuinely novel — it is not borrowed from any existing
  conformance tool. The closest precedent is pytest fixtures with
  scoping and Bazel's build graph, neither of which targets
  declarative DSL-driven testing.

## Alternatives considered

- **Inline preludes per story.** Rejected: duplication grows
  linearly with suite size.
- **Hidden auto-prelude controlled by a Meta flag.** Rejected:
  hides setup from story readers, prevents variation.
- **Resource-file imports.** Rejected: introduces a second file
  format with different semantics, conflicts with the
  one-artifact-type principle.
- **Per-story `Setup:` and `Teardown:` blocks composed from
  imported helpers.** Rejected as a partial solution: still
  requires authors to write boilerplate setup, and the cross-story
  reuse is achieved through a different mechanism than the
  intuitive "this test depends on that test."

## References

- Constitution principles I, IV
- ADR 0006 (story DSL grammar) — Meta keys including `Depends`
- ADR 0014 (IP and authoring guidelines) — `Spec-Ref` requirement
  for conformance stories, helper distinction
- ADR 0016 (cache and lock subsystem) — cache key construction,
  invalidation rules, lock protocol
