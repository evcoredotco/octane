package clock

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// waiter represents a single goroutine blocked in Sleep or After, holding
// the duration it is waiting for and the deadline (seed + accumulated wait).
type waiter struct {
	// deadline is the absolute clock time at which this waiter unblocks.
	deadline time.Time
	// ch receives the clock time when the waiter is unblocked.
	ch chan time.Time
}

// DeterministicClock is a Clock whose current time advances only when
// Advance is called explicitly. It is safe for concurrent use.
//
// Obtain an instance via Deterministic.
type DeterministicClock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []*waiter
}

// Deterministic returns a Clock frozen at seed. Call Advance on the
// returned *DeterministicClock to advance time in tests.
func Deterministic(seed time.Time) *DeterministicClock {
	return &DeterministicClock{
		mu:      sync.Mutex{},
		now:     seed,
		waiters: make([]*waiter, 0),
	}
}

// Now returns the current frozen time of the clock.
func (c *DeterministicClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

// Sleep blocks until the clock has been advanced by at least d from the
// moment Sleep was called, or until ctx is cancelled.
// Returns ctx.Err() if the context is cancelled before d elapses.
// Returns nil immediately when d == 0.
func (c *DeterministicClock) Sleep(ctx context.Context, d time.Duration) error {
	if d == 0 {
		return nil
	}

	wtr := c.register(d)

	select {
	case <-wtr.ch:
		return nil
	case <-ctx.Done():
		c.deregister(wtr)

		return fmt.Errorf("clock: context cancelled: %w", ctx.Err())
	}
}

// After returns a channel that fires when the clock has been advanced by at
// least d from the moment After is called. The channel is buffered with
// capacity 1.
func (c *DeterministicClock) After(d time.Duration) <-chan time.Time {
	wtr := c.register(d)

	return wtr.ch
}

// Advance moves the clock forward by d and unblocks any Sleep or After
// calls whose wait duration has been fully elapsed. Advance is safe for
// concurrent use from test goroutines.
func (c *DeterministicClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)
	c.drainWaiters()
}

// register creates a waiter for d and appends it to the waiters list.
// If d == 0, the waiter is immediately resolved before returning.
func (c *DeterministicClock) register(d time.Duration) *waiter {
	c.mu.Lock()
	defer c.mu.Unlock()

	wtr := &waiter{
		deadline: c.now.Add(d),
		ch:       make(chan time.Time, 1),
	}

	if !c.now.Before(wtr.deadline) {
		wtr.ch <- c.now

		return wtr
	}

	c.waiters = append(c.waiters, wtr)

	return wtr
}

// deregister removes a waiter from the list without firing it. Used when
// a Sleep's context is cancelled before the clock catches up.
func (c *DeterministicClock) deregister(target *waiter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	remaining := c.waiters[:0]

	for _, wtr := range c.waiters {
		if wtr != target {
			remaining = append(remaining, wtr)
		}
	}

	c.waiters = remaining
}

// drainWaiters fires every waiter whose deadline is at or before c.now.
// Must be called with c.mu held.
//
// The slice reuse (remaining = c.waiters[:0]) is intentional — it keeps
// the backing array in place to avoid a per-Advance allocation. This is safe
// because all sends to wtr.ch are non-blocking (buffered with capacity 1)
// and c.waiters is only accessed while c.mu is held.
func (c *DeterministicClock) drainWaiters() {
	remaining := c.waiters[:0]

	for _, wtr := range c.waiters {
		if !c.now.Before(wtr.deadline) {
			wtr.ch <- c.now
		} else {
			remaining = append(remaining, wtr)
		}
	}

	c.waiters = remaining
}
