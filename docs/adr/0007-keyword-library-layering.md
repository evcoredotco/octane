# ADR 0007: Keyword Library Layering — Primitive and Domain

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** V (Stdlib-Heavy), VI (Test Cases
  as Code), XII (No CSMS-specific adaptation)

## Context

ADR 0006 defines the `.story` grammar. To execute a story, OCTANE
must resolve each step to a Go function that produces an outcome
(pass / fail / observation). The mapping from step text to Go
function is the **keyword library**.

Per constitution principle XII, OCTANE has no CSMS-specific
adaptation surface: domain keywords are the same for every CSMS
implementing a given OCPP version. A keyword library that allowed
per-CSMS overrides would conflate "the CSMS is conformant" with "the
override hides a deviation," which is exactly what a conformance tool
must not do.

The library therefore has exactly **two layers**, with strict
ownership and resolution rules.

## Decision

```
        ┌────────────────────────────────────────┐
1.  ┌──> Domain keywords         (OCPP-level)     │  highest precedence
2.  │    Primitive keywords      (transport)      │  fallback
    │                                              │
    └─ resolution: domain → primitive  ───────────┘
```

A step matches the first registered pattern in this order. A failure
to match either layer produces a parser-stage error before the run
starts (no silent fallthrough at runtime).

### Layer 1 — Domain keywords

**Owner:** OCTANE project, `pkg/keywords/domain/v16/`,
`pkg/keywords/domain/v201/`, `pkg/keywords/domain/v21/`.

Domain keywords encode OCPP semantics. They are the keywords story
authors use 95% of the time.

Examples:

- `station {station:string} sends BootNotification with reason {reason:string}`
- `the CSMS responds with status {status:string} within {timeout:duration}`
- `station {station:string} starts a transaction with id token {token:string}`
- `the CSMS authorizes id token {token:string}`
- `station {station:string} sends Heartbeat`

Domain keywords are versioned per OCPP version. A keyword in
`domain/v201/` is invisible when a story declares
`Spec-Ref: OCPP-1.6 / ...` and vice versa. This keeps OCPP version
ambiguity out of stories.

Adding a new domain keyword is an ADR-level decision when it
expands the public DSL surface meaningfully. Adjusting the wording
of an existing one is a normal PR.

**Domain keywords do not vary by CSMS.** This is the operative
constitutional rule. If a CSMS deviates from the OCPP specification,
that deviation surfaces as a finding when the keyword runs against
it — not as a different keyword for that CSMS.

### Layer 2 — Primitive keywords

**Owner:** OCTANE project, `pkg/keywords/primitive/`.

Primitives are transport-level operations the engine knows
natively. They never reference OCPP semantics.

Examples:

- `open WebSocket to {url:string}`
- `close WebSocket for station {station:string}`
- `send raw frame {payload:string} on station {station:string}`
- `expect frame within {timeout:duration}`
- `assert frame field {path:string} equals {value:any}`
- `wait {duration:duration}`

Primitives are escape hatches. Stories use them rarely; domain
keywords compose them.

### Resolution rules

1. The runner consults the domain library scoped to the story's
   OCPP version. If a pattern matches, that keyword wins.
2. Otherwise, the runner consults the primitive library.
3. Otherwise, the run fails preflight with a clear diagnostic
   citing the step text and the layers that were searched, plus
   the closest registered patterns by Levenshtein distance ≤ 4.

### Pattern syntax

Patterns use `{name:type}` placeholders. Supported types:

- `string`     — double-quoted text
- `int`        — integer
- `float`      — floating-point number
- `bool`       — `true` / `false`
- `duration`   — `30 seconds`, `5 minutes`, etc.
- `any`        — accepts any of the above; the keyword must coerce
- `station`    — alias of `string`, validated as a registered station handle

Each keyword registers exactly one pattern. Two keywords cannot
register overlapping patterns; the registration step fails the
build with a startup panic.

### Registration

Keywords are registered at package init() with:

```go
registry.Register(api.Keyword{
    Layer:       api.LayerDomain,
    OCPPVersion: api.OCPP201,
    Pattern: "station {station:string} sends BootNotification " +
        "with reason {reason:string}",
    Func: func(ctx context.Context, s api.State, args api.Args) error {
        // ... drives the wire ...
    },
})
```

### Keyword author surface

The `pkg/keywords/api` package defines the contract every keyword
function honors. The surface is small by design — keyword authors
should only need to learn what is documented here.

**The `Func` signature.** Every keyword has the type
`func(ctx context.Context, state State, args Args) error`. A non-nil
return marks the step failed; the error message becomes the finding
text in the report. The `context.Context` carries cancellation and
the per-step timeout.

**The `Args` accessor.** Bound parameters are accessed by typed
methods: `args.String(name)`, `args.Int(name)`, `args.Duration(name)`,
etc. These methods panic if the named argument is missing or has the
wrong type. **This is intentional**: the registry validates the
pattern's `{name:type}` placeholders against the keyword's accesses
at registration time, so reaching the panic at runtime indicates a
registry bug, not an author bug. Authors do not need defensive
type-checking in keyword bodies.

**The `State` interface.** The runtime exposes itself to keywords
through a small interface, not a concrete type:

```go
type State interface {
    Station(handle string) (Station, error)
    Now() time.Time
    Logf(format string, args ...any)
}
```

Keeping `State` as an interface means keyword libraries can be
unit-tested against a mock state without importing the runtime
package. The mock implements the same three methods and the
keyword body cannot tell the difference.

**The `Station` interface.** Wire I/O happens through:

```go
type Station interface {
    Send(ctx context.Context, frame []any) error
    Expect(ctx context.Context) ([]any, error)
}
```

Frames are OCPP-J JSON arrays in their decoded Go form (per
spec 001: arrays decode to `[]any`, numbers to `float64`).

**Determinism rule.** Keywords MUST use `state.Now()` instead of
`time.Now()`. The runtime injects a deterministic clock so reports
can be byte-identical across runs (constitution principle IV). The
linter rejects `time.Now()` calls in `pkg/keywords/`.

**Mock-friendliness contract.** Anything a keyword needs from the
runtime goes through the `State` or `Station` interfaces. A keyword
that imports `pkg/runtime` directly is a code smell; the reviewer
agent flags it.

### Authoring patterns

Two patterns recur often enough to be worth pinning as project
conventions:

**Pattern A: request/response pairs for CSMS-initiated flows.**
When the CSMS sends a CALL and a station responds with a
CALLRESULT (e.g. ReserveNow, RemoteStartTransaction), express the
two halves as two keywords:

```
When  the CSMS sends ReserveNow with connectorId 1 and idTag "X" 
      to station "CP01" within 30 seconds
Then  station "CP01" responds with ReserveNow.conf status "Faulted"
```

The first keyword (the *expectation*) blocks until the inbound
CALL arrives and validates its fields. The second keyword (the
*response*) sends the matching CALLRESULT, echoing the inbound
`messageId` correctly. The runtime carries a small per-station
scratch space (see spec 001) so the response keyword can correlate
with the prior expectation.

This pattern matches the OCPP-J wire protocol literally and keeps
the story DSL readable.

**Pattern B: defensive enum validation inside keywords.** When a
keyword takes a status value as an argument (e.g.
`status "Faulted"`), validate it against the OCPP-defined enum
inside the keyword body and return a clear error for unknown
values:

```go
if !isValidReserveNowStatus(status) {
    return fmt.Errorf(
        "ReserveNow.conf.status: %q is not a valid OCPP 1.6 value "+
        "(Accepted, Faulted, Occupied, Rejected, Unavailable)",
        status,
    )
}
```

This catches story-author typos at preflight rather than producing
a wire-level error during execution.

### Resolver inspection

The CLI exposes the resolved keyword set for inspection:

| Command | Output |
|---------|--------|
| `octane keywords list` | All registered keywords, sorted by `(Layer, OCPPVersion, Pattern)` |
| `octane keywords list --layer domain --ocpp 1.6` | Filtered subset |
| `octane keywords resolve --story foo.story` | For each step in the story, the layer that wins |

The sort order is deterministic (constitution principle IV); two
runs of `octane keywords list` against the same binary produce
byte-identical output.

## Consequences

### Positive

- The keyword library reflects the project's commitment to
  CSMS-agnostic conformance: there is no place to hide a per-CSMS
  override.
- Story authors learn one OCPP-version keyword set per OCPP
  version. The set is the same regardless of which CSMS they target.
- Resolution is deterministic and inspectable: `octane keywords
  resolve --story foo.story` prints the layer that wins for each
  step.
- The reviewer agent can enforce that PRs touching
  `pkg/keywords/domain/` carry an ADR or a strong rationale, since
  they expand the public DSL.

### Negative

- A genuine CSMS quirk that is not a conformance violation but is
  observable on the wire (e.g. CitrineOS auto-commissioning a station
  on first connect) cannot be papered over. Stories that rely on
  such behavior must declare it explicitly and the operator must
  configure the CSMS appropriately. This is a feature, not a bug:
  it forces honesty about what the CSMS does.
- Slightly less flexibility than a three-layer library, but the
  flexibility being given up is precisely the flexibility a
  conformance tool must refuse.

### Neutral

- Connection metadata (URL templates, ports, subprotocol mappings)
  is *not* a keyword concern. It lives in connection profiles
  (ADR 0010) and is consumed by the runtime before any keyword
  executes.

## Alternatives considered

- **Three layers including profile-keyword overrides.** Considered
  during early design and rejected on integrity grounds (see
  constitution principle XII).
- **Flat library, namespace by prefix.** Rejected: pollutes the
  global pattern space.
- **Story-level keyword imports (Robot Framework `*** Settings ***`
  style).** Rejected for v1 because it leaks operational concerns
  into stories, violating constitution principle XII.

## References

- Constitution: principles V, VI, XII
- ADR 0005 (story-driven framework)
- ADR 0006 (story DSL grammar)
- ADR 0010 (connection profiles)
