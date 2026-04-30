package rand

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand/v2"
)

// realRand wraps math/rand/v2 with a cryptographically seeded PCG source.
type realRand struct {
	rng *mrand.Rand
}

// Real returns a Rand backed by math/rand/v2 with a cryptographically
// random seed.
func Real() Rand {
	var seed [16]byte

	_, err := crand.Read(seed[:])
	if err != nil {
		panic("clock/rand: crypto/rand unavailable: " + err.Error())
	}

	seed1 := binary.LittleEndian.Uint64(seed[:8])
	seed2 := binary.LittleEndian.Uint64(seed[8:])

	return &realRand{
		//nolint:gosec // G404: math/rand/v2 is intended; seed is crypto/rand.
		rng: mrand.New(mrand.NewPCG(seed1, seed2)),
	}
}

// Int63 returns a non-negative pseudo-random 63-bit integer.
func (r *realRand) Int63() int64 {
	return r.rng.Int64N(1<<63 - 1)
}

// Float64 returns a pseudo-random float64 in [0.0, 1.0).
func (r *realRand) Float64() float64 {
	return r.rng.Float64()
}

// Intn returns a non-negative pseudo-random int in [0, n).
// Panics if n <= 0.
func (r *realRand) Intn(n int) int {
	return r.rng.IntN(n)
}
