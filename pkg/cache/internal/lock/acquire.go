// Package lock is documented in errors.go.
package lock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// ErrLockTimeout is returned by [Acquire] when the caller has been
// unable to obtain the exclusive flock within the configured timeout,
// or immediately when [noWait] is true and the lock is already held.
//
// Callers should use [errors.Is] to distinguish this sentinel from
// transient I/O errors:
//
//	closer, err := lock.Acquire(ctx, path, timeout, noWait)
//	if errors.Is(err, lock.ErrLockTimeout) {
//	    // another process holds the lock; propagate or skip
//	}
var ErrLockTimeout = errors.New("lock: timed out waiting for lock")

// lockCloser wraps an *os.File returned by lockFile and implements
// io.Closer by calling unlockFile on Close.
type lockCloser struct {
	file *os.File
}

// Close releases the exclusive flock and closes the underlying file
// descriptor by delegating to [unlockFile].
func (lc *lockCloser) Close() error {
	if err := unlockFile(lc.file); err != nil {
		return fmt.Errorf("lock: Close: %w", err)
	}

	return nil
}

// backoffState holds the mutable retry-loop state for [Acquire].
// Keeping it in a struct reduces the number of parameters threaded
// through the loop and keeps cognitive complexity below 7.
type backoffState struct {
	delay    time.Duration
	maxDelay time.Duration
}

// next returns the current delay and advances the state to the next
// (doubled) delay, capped at maxDelay.
func (bs *backoffState) next() time.Duration {
	current := bs.delay
	bs.delay *= 2

	if bs.delay > bs.maxDelay {
		bs.delay = bs.maxDelay
	}

	return current
}

// Acquire attempts to obtain an exclusive flock on the file at
// lockPath, returning an [io.Closer] that releases the lock when
// closed.
//
// The retry loop uses exponential backoff starting at 1 ms and
// capped at 100 ms. Two early-exit conditions override the loop:
//
//   - If noWait is true and the first [lockFile] attempt returns
//     [ErrLockBusy], Acquire returns [ErrLockTimeout] immediately
//     without sleeping.
//   - If timeout elapses or ctx is cancelled, Acquire returns
//     [ErrLockTimeout] (for timeout) or ctx.Err() (for cancellation).
//
// A zero timeout means "use only the context deadline"; a negative
// timeout behaves the same.
func Acquire(
	ctx context.Context,
	lockPath string,
	timeout time.Duration,
	noWait bool,
) (io.Closer, error) {
	deadline, cancelFn := buildDeadlineContext(ctx, timeout)
	defer cancelFn()

	backoff := &backoffState{
		delay:    time.Millisecond,
		maxDelay: 100 * time.Millisecond,
	}

	for {
		fileHandle, err := lockFile(lockPath)
		if err == nil {
			return &lockCloser{file: fileHandle}, nil
		}

		if !errors.Is(err, ErrLockBusy) {
			return nil, fmt.Errorf("lock: Acquire: %w", err)
		}

		// Lock is busy. Apply no-wait or backoff logic.
		if noWait {
			return nil, ErrLockTimeout
		}

		sleepDuration := backoff.next()

		if err = sleepOrCancel(deadline, sleepDuration); err != nil {
			return nil, err
		}
	}
}

// buildDeadlineContext returns a child context that expires at the
// earlier of ctx's own deadline and now+timeout (when timeout > 0).
// The returned cancel function must always be called by the caller.
func buildDeadlineContext(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, timeout)
}

// sleepOrCancel waits for sleepDuration, returning early with
// [ErrLockTimeout] if the deadline context has expired, or
// ctx.Err() if the parent context was cancelled for another reason.
func sleepOrCancel(
	deadline context.Context,
	sleepDuration time.Duration,
) error {
	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-deadline.Done():
		return mapContextErr(deadline.Err())

	case <-timer.C:
		return nil
	}
}

// mapContextErr translates a context error into the appropriate
// sentinel for callers of [Acquire].
//
//   - context.DeadlineExceeded → [ErrLockTimeout] (our timeout fired).
//   - any other error (context.Canceled) → returned as-is so callers
//     can distinguish a deliberate cancellation from a timeout.
func mapContextErr(ctxErr error) error {
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return ErrLockTimeout
	}

	return ctxErr
}
