# ADR 0018: Determinism Primitives — Clock and Rand Injection

- **Status:** Accepted
- **Date:** 2026-04-27
- **Deciders:** Project maintainer, Architect, Backend
- **Constitution principles touched:** IV (Determinism), V (Stdlib-Heavy),
  VI (Test Cases as Code)

## Context

OCTANE's engine drives real CSMS instances over the wire. Every test run must
produce an identical result for identical inputs — a reproducibility guarantee
the constitution calls "determinism" (principle IV).

Two standard library facilities break this guarantee silently:

| Facility | Why it breaks determinism |
|----------|--------------------------|
| `time.Now()` | Returns the wall clock, which advances differently on every invocation and every machine. |
| `crypto/rand` | Returns cryptographically random bytes that differ per call, making any sequence that depends on them non-reproducible. |

Direct calls to these facilities in engine code, keyword bodies, or the runner
mean that running the same `.story` file twice against the same CSMS can
produce different timing measurements, different message correlation IDs, and
therefore different reports. Downstream property — "run the suite twice with
the same seed, get the same report" — becomes untestable.

The canonical solution in Go is **dependency injection**: define an interface
for each facility, inject a production implementation in the binary entrypoint,
and inject a test double in unit tests. This ADR formalises the interface
contracts and the injection model for OCTANE.

## Decision

### Interfaces

Two interfaces live under `pkg/engine/`:

#### `pkg/engine/clock/clock.go` — `Clock`

```go
type Clock interface {
    Now() time.Time
    Sleep(d time.Duration)
}
```

- `Now()` returns the current logical time.
- `Sleep(d)` blocks for the logical duration `d`.

In production, `Real()` delegates to `time.Now()` and `time.Sleep()`.
In tests, `Deterministic(seed time.Time)` advances only when the test
explicitly calls `Tick(d)` on the returned handle.

#### `pkg/engine/rand/rand.go` — `Rand`

```go
type Rand interface {
    Int63() int64
    Float64() float64
}
```

- `Int63()` returns a non-negative pseudo-random `int64`.
- `Float64()` returns a pseudo-random `float64` in `[0.0, 1.0)`.

In production, `Real()` wraps `math/rand/v2` with a time-seeded source.
In tests, `Deterministic(seed uint64)` uses a fixed seed so that any
sequence of calls produces a known, frozen output verified by a golden
test across platforms.

### Injection model

Interfaces are passed as **function parameters**, not stored in global
state or package-level variables. The pattern at every injection site is:

```go
// Good — parameter injection
func SendBootNotification(
    ctx context.Context,
    st transport.Station,
    clk clock.Clock,
    rng rand.Rand,
) error { ... }

// Bad — package-level state
var GlobalClock clock.Clock = clock.Real()
```

Package-level `var` injection is explicitly rejected because:

1. It makes test isolation fragile: a test that forgets to restore the
   global breaks any test that runs after it.
2. It is invisible in function signatures, making the dependency graph
   opaque to reviewers.
3. It couples packages that would otherwise be independent.

The wiring layer (the binary's `main` package and the Action entrypoint)
is the only site that constructs production implementations. Everything
below `main` receives interfaces.

### Forbidden call sites

The `forbidigo` linter rule in `.golangci.yaml` already forbids `time.Now`
and `math/rand.*` in non-test Go files under `pkg/`. This ADR documents why.

Any code that needs the current time or a random number MUST accept a
`clock.Clock` or `rand.Rand` parameter. Pull requests that bypass this
constraint will be rejected at review.

`crypto/rand` for key material is excluded from this rule — it is not
injected because its non-determinism is a feature in the security domain.
However, no OCTANE engine code uses `crypto/rand`; it is only consumed by
TLS infrastructure in the standard library.

## Goroutine scheduling and channel-based coordination

OCTANE's test engine is concurrent: each `.story` file runs keyword steps
sequentially, but multiple stations may execute in parallel (ADR 0008).
Timing coordination between goroutines must never use `time.After` or
`time.NewTimer` in keyword bodies, because those functions call the real
wall clock and cannot be controlled by an injected `Clock`.

The approved coordination pattern is **channel-based**:

```go
// Good — channel rendezvous, no real-time coupling
done := make(chan struct{})
go func() {
    defer close(done)
    _ = st.Expect(ctx)
}()
select {
case <-done:
case <-ctx.Done():
    return ctx.Err()
}

// Bad — real-wall-clock timer, not injectable
select {
case msg := <-inbox:
    _ = msg
case <-time.After(5 * time.Second): // forbidden
    return errors.New("timeout")
}
```

Deadlines are expressed as `context.Context` cancellations. The engine sets
a `context.WithDeadline` at the story level using the injected `Clock.Now()`
as the reference point. Keyword bodies observe the deadline through `ctx.Done()`
and do not compute their own timeouts.

`clock.Clock.Sleep(d)` is the only approved blocking wait inside keyword
bodies when a literal delay is required (e.g. a "Wait 2s" keyword). The
`Deterministic` implementation makes this a no-op in unit tests unless the
test explicitly advances the clock, so tests never actually wait.

## Test double contracts

### Deterministic Clock

```
clock.Deterministic(seed time.Time) (*DeterministicClock, Clock)
```

- `Now()` returns `seed + accumulated_ticks`. No wall-clock reads occur.
- `Sleep(d)` records the duration but does not advance the clock and does
  not block the goroutine. The test controls advancement via `Tick(d)`.
- `Tick(d)` advances `accumulated_ticks` by `d`. Any goroutine blocked in
  `Sleep` is unblocked if its requested duration is now satisfied.
- The clock is safe for concurrent use. Goroutines that call `Sleep` block
  on a channel; `Tick` broadcasts to all waiting goroutines.

A test that wants to assert behavior after a 5-second window:

```go
clk, iface := clock.Deterministic(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
// inject iface into the code under test
clk.Tick(5 * time.Second) // instant in the test; no real sleep
// assert post-5s behavior
```

### Deterministic Rand

```
rand.Deterministic(seed uint64) Rand
```

- `Int63()` and `Float64()` derive from `math/rand/v2`'s `NewPCG(seed, 0)`
  source, which produces a platform-independent sequence for any given seed.
- The golden sequence for seed `0xDEADBEEF` is verified in
  `pkg/engine/rand/determinism_test.go` on every CI run. A test failure
  indicates a toolchain regression, not a code bug.
- The `Deterministic` Rand is not safe for concurrent use. Keyword bodies
  that need concurrent randomness receive separate `Rand` instances.

## Consequences

### Positive

- **Reproducibility is enforced by the type system.** A function that accepts
  `clock.Clock` cannot accidentally call `time.Now()`; the compiler would
  reject an unqualified `Now()` call. The linter rejects the import.
- **Tests are fast.** Unit tests that exercise time-based logic complete
  instantly because `Deterministic.Sleep` does not actually sleep.
- **Golden sequences enable regression detection.** A toolchain upgrade that
  silently changes pseudo-random output is caught immediately.
- **Parameter injection scales.** Adding a new injectable facility (e.g. a
  `Logger` interface) follows the same pattern with no architectural surgery.

### Negative

- **Function signature verbosity.** Functions that orchestrate timing and
  randomness carry two extra parameters. Mitigated by grouping primitives in
  an `engine.Context` struct if the parameter count grows beyond three in a
  single function (to be decided if needed).
- **Coordination complexity.** Channel-based coordination is more verbose than
  `time.After`. This is accepted as the cost of determinism.

### Neutral

- `math/rand/v2` (introduced in Go 1.22) is a standard library package.
  Using it for the `Real` Rand implementation does not add a third-party
  dependency, honoring constitution principle V.

## Alternatives considered

- **Global clock variable** (`var Now = time.Now`). Common in Go codebases,
  but ruled out because test isolation requires external restoration of the
  global, which is error-prone in parallel tests (`t.Parallel()` is mandatory
  on all test cases, making globals inherently unsafe).
- **`context.Context` values for Clock and Rand.** The `context` package
  discourages storing typed values in Context. This pattern is common but
  considered bad practice in the Go community and rejected here in favour of
  explicit parameters.
- **Separate test build tags** (`//go:build !test`) to swap implementations.
  Rejected: build-tag–based swapping hides dependencies from normal builds and
  is incompatible with the requirement that the production binary is testable
  end-to-end.
- **`time.After` with `select`** for keyword timeouts. Rejected: ties keyword
  bodies to the wall clock, breaking the determinism guarantee.

## References

- Constitution: principles IV, V, VI
- ADR 0003 (WebSocket library — nhooyr.io/websocket)
- ADR 0007 (keyword library layering — where Clock and Rand are consumed)
- ADR 0008 (multi-station orchestration — concurrent injection)
- Spec 002 (wire engine) — AC5, AC6
- `math/rand/v2` package: https://pkg.go.dev/math/rand/v2
- PCG random number generator: https://www.pcg-random.org/
