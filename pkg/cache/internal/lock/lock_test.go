// Package lock_test contains cross-platform black-box tests for the
// double-checked lock acquire pattern (T-005-32, T-005-33).
//
// These tests exercise [lock.Acquire] as an external caller would:
// via the exported sentinel errors and the returned [io.Closer].
// The tests run on Linux, macOS, and Windows (CI matrix per AC7).
package lock_test

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/octane-project/octane/pkg/cache/internal/lock"
)

// lockPath returns a per-test temp-file path suitable for use as a
// lock file. Using a unique path per test avoids cross-test interference
// when tests run in parallel.
func lockPath(tb testing.TB) string {
	tb.Helper()

	return filepath.Join(tb.TempDir(), "test.lock")
}

// Test_lock_AcquireAndRelease verifies the happy path: a lock can be
// acquired on a fresh path, and after Close the same path can be
// locked again (demonstrating that the flock was released).
func Test_lock_AcquireAndRelease(t *testing.T) {
	t.Parallel()

	path := lockPath(t)

	ctx := context.Background()

	closer, err := lock.Acquire(ctx, path, 0, false)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}

	if err = closer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// The lock should now be free; a second Acquire must succeed.
	closer2, err := lock.Acquire(ctx, path, 0, false)
	if err != nil {
		t.Fatalf("second Acquire after Close: %v", err)
	}

	if err = closer2.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

// Test_lock_NoWaitReturnsBusy verifies that when noWait is true and
// the lock is held by another goroutine, Acquire returns ErrLockTimeout
// immediately without sleeping.
func Test_lock_NoWaitReturnsBusy(t *testing.T) {
	t.Parallel()

	path := lockPath(t)

	ctx := context.Background()

	// Goroutine 1: hold the lock for the duration of the test.
	holder, err := lock.Acquire(ctx, path, 0, false)
	if err != nil {
		t.Fatalf("holder Acquire: %v", err)
	}

	defer func() {
		if closeErr := holder.Close(); closeErr != nil {
			t.Errorf("holder Close: %v", closeErr)
		}
	}()

	// Goroutine 2: attempt with noWait=true; must fail immediately.
	const noWait = true

	_, err = lock.Acquire(ctx, path, 0, noWait)
	if !errors.Is(err, lock.ErrLockTimeout) {
		t.Errorf(
			"noWait Acquire: got %v, want ErrLockTimeout",
			err,
		)
	}
}

// Test_lock_TimeoutExpires verifies that when a non-zero timeout is
// supplied and the lock remains busy, Acquire returns ErrLockTimeout
// after approximately the given duration.
func Test_lock_TimeoutExpires(t *testing.T) {
	t.Parallel()

	path := lockPath(t)

	ctx := context.Background()

	// Goroutine 1: hold the lock throughout.
	holder, err := lock.Acquire(ctx, path, 0, false)
	if err != nil {
		t.Fatalf("holder Acquire: %v", err)
	}

	defer func() {
		if closeErr := holder.Close(); closeErr != nil {
			t.Errorf("holder Close: %v", closeErr)
		}
	}()

	const lockTimeout = 50 * time.Millisecond

	started := time.Now()

	_, err = lock.Acquire(ctx, path, lockTimeout, false)

	elapsed := time.Since(started)

	if !errors.Is(err, lock.ErrLockTimeout) {
		t.Errorf(
			"timeout Acquire: got %v, want ErrLockTimeout",
			err,
		)
	}

	// Sanity-check: elapsed must be at least the timeout duration.
	// We allow a 3x upper bound for slow CI environments.
	const maxFactor = 3

	if elapsed < lockTimeout {
		t.Errorf(
			"returned too early: elapsed %v < timeout %v",
			elapsed,
			lockTimeout,
		)
	}

	if elapsed > maxFactor*lockTimeout {
		t.Logf(
			"warning: elapsed %v greatly exceeds timeout %v "+
				"(CI may be slow)",
			elapsed,
			lockTimeout,
		)
	}
}

// Test_lock_ContextCancel verifies that when the parent context is
// cancelled while Acquire is waiting for a busy lock, Acquire returns
// the context's error (context.Canceled), not ErrLockTimeout.
func Test_lock_ContextCancel(t *testing.T) {
	t.Parallel()

	path := lockPath(t)

	// Goroutine 1: hold the lock throughout.
	holder, err := lock.Acquire(context.Background(), path, 0, false)
	if err != nil {
		t.Fatalf("holder Acquire: %v", err)
	}

	defer func() {
		if closeErr := holder.Close(); closeErr != nil {
			t.Errorf("holder Close: %v", closeErr)
		}
	}()

	cancelCtx, cancel := context.WithCancel(context.Background())

	var (
		acquireErr error
		waitGroup  sync.WaitGroup
	)

	const concurrency = 1

	waitGroup.Add(concurrency)

	go func() {
		defer waitGroup.Done()

		// No timeout; rely solely on context cancellation.
		_, acquireErr = lock.Acquire(cancelCtx, path, 0, false)
	}()

	// Give the goroutine a moment to enter the retry loop before
	// cancelling. A short sleep is acceptable here because the
	// goroutine must reach the select inside sleepOrCancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	waitGroup.Wait()

	if !errors.Is(acquireErr, context.Canceled) {
		t.Errorf(
			"ContextCancel Acquire: got %v, want context.Canceled",
			acquireErr,
		)
	}
}
