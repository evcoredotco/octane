// Package rand defines the Rand interface and its implementations.
//
// Code that depends on randomness must consume a Rand injected via function
// parameter. Direct calls to crypto/rand or math/rand are forbidden inside
// pkg/keywords/, pkg/runner/, and pkg/engine/ (the linter enforces this via
// forbidigo). Use rand.Real() in production wiring and
// rand.Deterministic(seed) in tests.
package rand

// Rand abstracts random-number generation so that code depending on
// randomness (e.g. unique ID generation) can be tested deterministically.
// Inject Rand via function parameter; never call crypto/rand or math/rand
// directly in pkg/keywords/, pkg/runner/, or pkg/engine/.
type Rand interface {
	// Int63 returns a non-negative pseudo-random 63-bit integer.
	Int63() int64

	// Float64 returns a pseudo-random float64 in [0.0, 1.0).
	Float64() float64

	// Intn returns a non-negative pseudo-random int in [0, n).
	// Panics if n <= 0.
	Intn(n int) int
}
