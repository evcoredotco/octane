// Package clock defines the Clock interface and its implementations.
//
// Code that depends on the passage of time must consume a Clock injected via
// function parameter. Direct calls to time.Now() are forbidden inside
// pkg/keywords/, pkg/runner/, and pkg/engine/ (the linter enforces this via
// forbidigo). Use clock.Real() in production wiring and
// clock.Deterministic(seed) in tests.
package clock

import (
	"context"
	"time"
)

// Clock abstracts wall-clock access so that code depending on time can be
// tested deterministically. Inject Clock via function parameter; never call
// time.Now() directly in pkg/keywords/, pkg/runner/, or pkg/engine/.
type Clock interface {
	// Now returns the current time according to this clock.
	Now() time.Time

	// Sleep blocks until d has elapsed on this clock, or ctx is cancelled.
	// Returns ctx.Err() if cancelled.
	Sleep(ctx context.Context, d time.Duration) error

	// After returns a channel that receives the clock's current time after
	// d has elapsed on this clock. The channel is buffered with capacity 1.
	After(d time.Duration) <-chan time.Time
}
