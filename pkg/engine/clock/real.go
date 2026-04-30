package clock

import (
	"context"
	"time"
)

// realClock delegates all operations to the standard library system clock.
type realClock struct{}

// Real returns a Clock backed by the system wall clock.
func Real() Clock {
	return &realClock{}
}

// Now returns the current wall-clock time via time.Now.
func (*realClock) Now() time.Time {
	return time.Now()
}

// Sleep blocks until d has elapsed or ctx is cancelled.
// Returns ctx.Err() if the context is done before d elapses.
func (*realClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// After returns a channel that fires after d has elapsed.
// The channel is buffered with capacity 1 (guaranteed by time.After).
func (*realClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
