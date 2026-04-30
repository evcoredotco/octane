// Package primitive_test exercises the wait primitive keyword
// (spec 004 §10, item 8) against a [clock.DeterministicClock].
//
// Task: T-004-21
// AC5: Given a primitive keyword "wait {duration:duration}", when the
// keyword executes, then it sleeps the deterministic clock by exactly
// that duration; in deterministic mode, no real wall-clock time elapses.

package primitive_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/engine/clock"
	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── Named constants ──────────────────────────────────────────────────────────

const (
	// patternWait is the step text for the wait keyword.
	patternWait = "wait {duration:duration}"

	// waitDuration is the duration used in wait tests.
	waitDuration = 5 * time.Second

	// advanceDelay is the time the test goroutine waits before advancing
	// the deterministic clock, expressed as a real-time duration used only
	// to avoid a data-race where Advance fires before Sleep registers its
	// waiter. It is kept very short so the test suite stays fast.
	advanceDelay = time.Millisecond

	// testYear is the year used in deterministic clock seeds for wait tests.
	testYear = 2026

	// chanBufOne is the buffer size for single-result goroutine channels.
	chanBufOne = 1

	// realTimeLimitSec is the real-time ceiling (seconds) for
	// determinism tests.
	realTimeLimitSec = 2

	// clockMonthJanuary is the month component (January) for
	// deterministic clock seed values in wait tests.
	clockMonthJanuary = 1

	// clockDayFirst is the day component (1st) for deterministic clock
	// seed values in wait tests.
	clockDayFirst = 1

	// zeroHMSN is the zero value for hour/minute/second/nanosecond in
	// deterministic clock seeds.
	zeroHMSN = 0
)

// ── tests ────────────────────────────────────────────────────────────────────

// Test_primitive_wait_ReturnsNil verifies that the wait keyword returns nil
// when the deterministic clock is advanced to satisfy the sleep (AC5).
func Test_primitive_wait_ReturnsNil(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(time.Date(
		testYear, clockMonthJanuary, clockDayFirst,
		zeroHMSN, zeroHMSN, zeroHMSN, zeroHMSN, time.UTC,
	))

	state := mock.NewMockState()
	state.SetClock(clk)

	keywordFunc := resolveFunc(t, patternWait)

	args := api.NewArgs(map[string]any{
		"duration": waitDuration,
	})

	done := make(chan error, chanBufOne)

	go func() {
		done <- keywordFunc(context.Background(), state, args)
	}()

	// Advance the deterministic clock from a separate goroutine after a
	// brief real delay to guarantee Sleep has registered its waiter.
	// The delay is measured in real milliseconds (not engine time) and is
	// intentionally tiny so the test suite stays fast.
	time.Sleep(advanceDelay)
	clk.Advance(waitDuration)

	err := <-done
	if err != nil {
		t.Errorf("wait keyword: want nil error, got %v", err)
	}
}

// Test_primitive_wait_NoDeterministicClockRealTimeElapsed verifies that the
// wait keyword with a deterministic clock completes without any meaningful
// real wall-clock time elapsing (AC5 determinism invariant).
//
// The structural proof: [clock.DeterministicClock.Sleep] blocks on a channel
// that fires only when [clock.DeterministicClock.Advance] is called.  Because
// Advance is called immediately (after a 1 ms goroutine-start delay), the
// keyword returns in real-time close to zero regardless of the logical wait
// duration.  We assert completion within 2 seconds, a threshold that would
// be impossible to satisfy if real-time Sleep were used for a 5-second wait.
func Test_primitive_wait_NoDeterministicClockRealTimeElapsed(t *testing.T) {
	t.Parallel()

	clk := clock.Deterministic(time.Date(
		testYear, clockMonthJanuary, clockDayFirst,
		zeroHMSN, zeroHMSN, zeroHMSN, zeroHMSN, time.UTC,
	))

	state := mock.NewMockState()
	state.SetClock(clk)

	keywordFunc := resolveFunc(t, patternWait)

	args := api.NewArgs(map[string]any{
		"duration": waitDuration,
	})

	done := make(chan error, chanBufOne)

	go func() {
		done <- keywordFunc(context.Background(), state, args)
	}()

	// Give the goroutine time to enter Sleep before we advance.
	time.Sleep(advanceDelay)
	clk.Advance(waitDuration)

	// Use a real-time timeout well below the logical wait duration to
	// prove that no real wall-clock delay occurred.
	realTimeLimit := realTimeLimitSec * time.Second
	limitCtx, cancel := context.WithTimeout(context.Background(), realTimeLimit)

	defer cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf(
				"wait keyword: want nil error, got %v",
				err,
			)
		}
	case <-limitCtx.Done():
		t.Errorf(
			"wait keyword: real wall-clock time exceeded %v; "+
				"deterministic clock did not suppress real sleep",
			realTimeLimit,
		)
	}
}

// Test_primitive_wait_ContextCancelled verifies that the wait keyword
// returns an error when the context is cancelled before the duration
// elapses (AC5 cancellation path).
func Test_primitive_wait_ContextCancelled(t *testing.T) {
	t.Parallel()

	// Use a deterministic clock that we deliberately never advance.
	// The keyword should unblock via context cancellation instead.
	clk := clock.Deterministic(time.Date(
		testYear, clockMonthJanuary, clockDayFirst,
		zeroHMSN, zeroHMSN, zeroHMSN, zeroHMSN, time.UTC,
	))

	state := mock.NewMockState()
	state.SetClock(clk)

	keywordFunc := resolveFunc(t, patternWait)

	args := api.NewArgs(map[string]any{
		"duration": waitDuration,
	})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, chanBufOne)

	go func() {
		done <- keywordFunc(ctx, state, args)
	}()

	// Cancel the context after the goroutine has had time to enter Sleep.
	time.Sleep(advanceDelay)
	cancel()

	err := <-done

	// Invariant: a cancelled context must produce a non-nil error.
	if err == nil {
		t.Fatal("wait keyword: want error on context cancel, got nil")
	}

	// Invariant: the error chain must contain context.Canceled.
	if !errors.Is(err, context.Canceled) {
		t.Errorf(
			"wait keyword: want errors.Is(err, context.Canceled), got %v",
			err,
		)
	}
}
