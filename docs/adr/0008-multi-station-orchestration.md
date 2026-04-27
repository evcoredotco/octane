# ADR 0008: Multi-Station Orchestration

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** XI (Wire conformance), IV
  (Determinism)

## Context

ADR 0005 retired the Test Harness Adapter in favor of wire-only
conformance. Many stateful OCPP scenarios that the THA design
previously addressed via privileged CSMS access are still verifiable
on the wire — provided OCTANE can drive multiple simulated charging
stations from a single run and coordinate their interactions.

Examples that require multi-station orchestration:

- **Concurrent transactions on different stations** validating that the
  CSMS settles each transaction independently.
- **Same id token at two stations** validating the
  `ConcurrentTx` rejection behavior.
- **Reservation handover** validating that a reservation made by
  station A blocks station B.
- **Authorize cache invalidation** across stations.

Without first-class multi-station support these scenarios are
unreachable. With it, a meaningful fraction of stateful conformance
becomes wire-verifiable.

## Decision

OCTANE supports declaring N stations per scenario in the story Meta:

```
Meta
    Spec-Ref: OCPP-1.6 / TC_T_07_CS
    Title:    ConcurrentTx — same id token at two stations
    Tags:     core, transactions, multi-station
    Stations: 2
```

The runner allocates N station handles named `"CP01"`, `"CP02"`, … on
preflight. Stories reference stations by handle in step text:

```
When  station "CP01" sends Authorize with id token "VID:0001"
And   station "CP02" sends Authorize with id token "VID:0001"
Then  the CSMS rejects "CP02" with status "ConcurrentTx"
```

### Concurrency model

- Each station handle owns a dedicated WebSocket connection and a
  goroutine.
- Steps are executed **sequentially** in the order they appear in the
  story. A step targeting `"CP01"` blocks the runner until
  `"CP01"`'s outcome is known.
- Steps that need true concurrency declare a `Parallel` block:

```
Parallel
    When  station "CP01" sends StartTransaction
    When  station "CP02" sends StartTransaction
End-Parallel
```

  Inside a `Parallel` block the steps execute concurrently. The block
  fails if any contained step fails.
- The runner enforces a `--max-stations` ceiling (default 16) to keep
  resource use bounded.

### Determinism (constitution IV)

- Station handles are assigned in declared order from a deterministic
  ID space derived from the run seed.
- WebSocket connections are opened in declared order.
- Within a `Parallel` block, ordering is not deterministic on the wire
  — the CSMS may receive `"CP01"` first or `"CP02"` first. The runner
  records both possibilities in the report and treats either as
  acceptable unless the story specifies an ordering constraint.

### Station bring-up

The default Setup section opens connections for all declared stations
and performs BootNotification, unless a Meta key disables it:

```
Meta
    Stations: 2
    Auto-Boot: false
```

This default makes "ordinary" multi-station scenarios concise. Stories
that test the BootNotification handshake itself disable Auto-Boot to
take control.

### Station identity

Each station's `chargingStationSerial`, `vendor`, and `model` are
deterministically derived from `run_id + station_handle`. This means
the same scenario run twice produces identical station identities,
which keeps reports byte-diffable per principle IV.

Profiles MAY override the identity template via a profile keyword.

### Failure semantics

- A failure on any station immediately marks the whole scenario as
  failed.
- All stations' Teardown still runs, in reverse declaration order.
- The report's per-station section captures every wire event each
  station saw.

## Consequences

### Positive

- A meaningful fraction of stateful OCPP scenarios becomes wire-only
  verifiable.
- The grammar for multi-station is small and explicit; authors do not
  juggle implicit context.
- Concurrency is opt-in (Parallel block), keeping single-station
  stories simple and deterministic.
- Per-station report sections give certifiers granular evidence.

### Negative

- Resource use scales linearly with station count; long scenarios
  with high station counts can pressure CI runners.
  Mitigated by `--max-stations` and by documenting realistic ceilings
  for popular CI providers.
- Parallel block non-determinism complicates report diffing. The
  report design accommodates this by recording observed-wire-order
  separately from declared-step-order.

### Neutral

- A scenario that omits the `Stations` key defaults to 1 station,
  preserving the simplest single-station authoring path.

## Alternatives considered

- **One station per run; chain runs to simulate multi-station.**
  Rejected: cross-run state coordination is exactly the THA-style
  privilege OCTANE is avoiding.
- **Implicit station per step.** Rejected: ambiguous and bug-prone.
- **Threading-style "actors" model with mailboxes.** Rejected: more
  complex than scenario authors need; goroutines + channels keep the
  Go implementation idiomatic.

## References

- Constitution: principles IV, XI
- ADR 0005 (story framework)
- ADR 0006 (story DSL)
