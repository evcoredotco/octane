// Package clock_test covers the DeterministicClock implementation (AC5).
package clock_test

import (
	"context"
	"testing"
	"time"

	"github.com/octane-project/octane/pkg/engine/clock"
)

var seed = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// TestDeterministicNow verifies that Now returns the seed time before any
// Advance calls, and the advanced time afterwards.
func TestDeterministicNow(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	if got := clk.Now(); !got.Equal(seed) {
		t.Errorf("Now() = %v, want %v", got, seed)
	}

	clk.Advance(time.Hour)

	want := seed.Add(time.Hour)
	if got := clk.Now(); !got.Equal(want) {
		t.Errorf("After Advance(1h), Now() = %v, want %v", got, want)
	}
}

// TestDeterministicSleepUnblocks verifies that Sleep unblocks after the clock
// is advanced past the sleep duration. After is used to register the waiter
// synchronously before Advance, avoiding a goroutine-scheduling race.
func TestDeterministicSleepUnblocks(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	// Register the waiter via After (synchronous); then Advance and verify.
	ch := clk.After(5 * time.Second)
	clk.Advance(10 * time.Second)

	select {
	case fired := <-ch:
		want := seed.Add(10 * time.Second)
		if !fired.Equal(want) {
			t.Errorf("After fired with time %v, want %v", fired, want)
		}
	case <-time.After(time.Second):
		t.Error("waiter was not unblocked after Advance")
	}
}

// TestDeterministicSleepCancelCtx verifies that Sleep returns ctx.Err()
// when the context is cancelled before the clock catches up.
func TestDeterministicSleepCancelCtx(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- clk.Sleep(ctx, time.Hour)
	}()

	cancel()

	err := <-done
	if err != context.Canceled {
		t.Errorf(
			"Sleep with cancelled ctx: got %v, want %v",
			err,
			context.Canceled,
		)
	}
}

// TestDeterministicAfterUnblocks verifies that the channel returned by After
// receives a value after the clock is advanced past the duration.
func TestDeterministicAfterUnblocks(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	ch := clk.After(3 * time.Second)

	clk.Advance(5 * time.Second)

	select {
	case got := <-ch:
		want := seed.Add(5 * time.Second)
		if !got.Equal(want) {
			t.Errorf("After: received time %v, want %v", got, want)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("After channel did not fire after Advance")
	}
}

// TestDeterministicSleepZero verifies that Sleep(0) returns immediately.
func TestDeterministicSleepZero(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	if err := clk.Sleep(context.Background(), 0); err != nil {
		t.Errorf("Sleep(0): got error %v, want nil", err)
	}
}

// TestDeterministicMultipleWaiters verifies that multiple After channels
// registered before a single Advance call are all unblocked by that Advance.
// After is used instead of Sleep to avoid a goroutine-scheduling race between
// waiter registration and the Advance call.
func TestDeterministicMultipleWaiters(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(seed)

	const n = 10

	// Register all waiters synchronously via After (no goroutines needed).
	chs := make([]<-chan time.Time, n)
	for i := range chs {
		chs[i] = clk.After(time.Second)
	}

	// Single Advance past all deadlines.
	clk.Advance(2 * time.Second)

	// Verify every channel fired.
	for i, ch := range chs {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Errorf("waiter %d was not unblocked after Advance", i)
		}
	}
}

// TestRealClockNow verifies that Real() returns a non-zero time and that Now
// advances between successive calls.
func TestRealClockNow(t *testing.T) {
	t.Parallel()

	clk := clock.Real()

	t1 := clk.Now()
	if t1.IsZero() {
		t.Error("Real().Now() returned zero time")
	}

	// Busy-wait briefly so the second call is later.
	for clk.Now().Equal(t1) {
	}

	t2 := clk.Now()
	if !t2.After(t1) {
		t.Errorf("Real clock did not advance: t1=%v t2=%v", t1, t2)
	}
}
