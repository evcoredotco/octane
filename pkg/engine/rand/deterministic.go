package rand

import (
	mrand "math/rand/v2"
)

// deterministicRand wraps math/rand/v2 with a fixed PCG seed so that the
// same seed always produces the same sequence across platforms and Go
// versions (within the same Go major version).
type deterministicRand struct {
	rng *mrand.Rand
}

// Deterministic returns a Rand that produces a fixed, reproducible sequence
// for the given seed. The same seed always produces the same sequence
// across platforms and Go versions (within the same Go major version).
func Deterministic(seed uint64) Rand {
	return &deterministicRand{
		//nolint:gosec // G404: math/rand/v2 is intentional; this is a test double
		rng: mrand.New(mrand.NewPCG(seed, seed^0xDEADBEEF)),
	}
}

// Int63 returns a non-negative pseudo-random 63-bit integer.
func (r *deterministicRand) Int63() int64 {
	return r.rng.Int64N(1<<63 - 1)
}

// Float64 returns a pseudo-random float64 in [0.0, 1.0).
func (r *deterministicRand) Float64() float64 {
	return r.rng.Float64()
}

// Intn returns a non-negative pseudo-random int in [0, n).
// Panics if n <= 0.
func (r *deterministicRand) Intn(n int) int {
	return r.rng.IntN(n)
}
