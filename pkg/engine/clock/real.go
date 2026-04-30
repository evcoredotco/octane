package clock

import (
	"context"
	"fmt"
	"time"
)

// RealClock delegates all operations to the standard library system clock.
type RealClock struct{}

// Real returns a *RealClock backed by the system wall clock.
func Real() *RealClock {
	return &RealClock{}
}

// Now returns the current wall-clock time via time.Now.
func (*RealClock) Now() time.Time {
	return time.Now()
}

// Sleep blocks until d has elapsed or ctx is cancelled.
// Returns ctx.Err() if the context is done before d elapses.
func (*RealClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("clock: context cancelled: %w", ctx.Err())
	}
}

// After returns a channel that fires after d has elapsed.
// The channel is buffered with capacity 1 (guaranteed by time.After).
func (*RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
