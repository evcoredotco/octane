package lock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// ErrLockTimeout is returned by [Acquire] or [TryAcquire] when the
// caller has been unable to obtain the exclusive flock within the
// configured timeout, or immediately when [TryAcquire] finds the
// lock already held.
//
// Callers should use [errors.Is] to distinguish this sentinel from
// transient I/O errors:
//
//	closer, err := lock.Acquire(ctx, path, timeout)
//	if errors.Is(err, lock.ErrLockTimeout) {
//	    // another process holds the lock; propagate or skip
//	}
var ErrLockTimeout = errors.New("lock: timed out waiting for lock")

// errRetry is an internal sentinel used by [tryAcquireOnce] to
// signal that the lock was busy and the caller should loop again.
// It is never returned to external callers.
var errRetry = errors.New("lock: busy, retry")

// lockCloser wraps an *os.File returned by lockFile and implements
// io.Closer by calling unlockFile on Close.
type lockCloser struct {
	file *os.File
}

// Close releases the exclusive flock and closes the underlying file
// descriptor by delegating to [unlockFile].
func (lc *lockCloser) Close() error {
	err := unlockFile(lc.file)
	if err != nil {
		return fmt.Errorf("lock: Close: %w", err)
	}

	return nil
}

// maxBackoffDelay is the upper bound for exponential backoff in [Acquire].
const maxBackoffDelay = 100 * time.Millisecond

// noTimeout represents a zero (or negative) timeout duration, meaning
// no deadline is applied beyond the parent context's own deadline.
const noTimeout = 0

// backoffMultiplier is the doubling factor applied to the retry delay.
const backoffMultiplier = 2

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
	bs.delay *= backoffMultiplier

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
// capped at 100 ms. If timeout elapses or ctx is cancelled,
// Acquire returns [ErrLockTimeout] (for timeout) or ctx.Err()
// (for cancellation).
//
// A zero timeout means "use only the context deadline"; a negative
// timeout behaves the same.
//
// To fail immediately when the lock is already held, use [TryAcquire].
func Acquire(
	ctx context.Context,
	lockPath string,
	timeout time.Duration,
) (io.Closer, error) {
	deadline, cancelFn := buildDeadlineContext(ctx, timeout)
	defer cancelFn()

	backoff := &backoffState{
		delay:    time.Millisecond,
		maxDelay: maxBackoffDelay,
	}

	for {
		closer, err := tryAcquireOnce(deadline, lockPath, backoff)
		if err != nil && !errors.Is(err, errRetry) {
			return nil, err
		}

		if closer != nil {
			return closer, nil
		}
	}
}

// TryAcquire attempts a single non-blocking flock on the file at
// lockPath. If the lock is already held by another process,
// [ErrLockTimeout] is returned immediately without sleeping.
//
// A cancelled ctx is checked before the flock attempt; callers may
// use this to abort before any filesystem syscall.
//
// TryAcquire is the non-blocking counterpart of [Acquire].
func TryAcquire(
	ctx context.Context,
	lockPath string,
) (io.Closer, error) {
	err := ctx.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"lock: TryAcquire: context done: %w", err,
		)
	}

	fileHandle, err := lockFile(lockPath)
	if err == nil {
		return &lockCloser{file: fileHandle}, nil
	}

	if errors.Is(err, ErrLockBusy) {
		return nil, ErrLockTimeout
	}

	return nil, fmt.Errorf("lock: TryAcquire: %w", err)
}

// tryAcquireOnce performs a single attempt to lock lockPath.
// It returns a non-nil [io.Closer] on success, [errRetry] when the
// lock is busy and the caller should sleep-and-retry, or a terminal
// error when the attempt must be abandoned (non-busy I/O error or
// timeout).
func tryAcquireOnce(
	deadline context.Context,
	lockPath string,
	backoff *backoffState,
) (io.Closer, error) {
	fileHandle, err := lockFile(lockPath)
	if err == nil {
		return &lockCloser{file: fileHandle}, nil
	}

	if !errors.Is(err, ErrLockBusy) {
		return nil, fmt.Errorf("lock: Acquire: %w", err)
	}

	sleepErr := sleepOrCancel(deadline, backoff.next())
	if sleepErr != nil {
		return nil, sleepErr
	}

	return nil, errRetry
}

// buildDeadlineContext returns a child context that expires at the
// earlier of ctx's own deadline and now+timeout (when timeout > 0).
// The returned cancel function must always be called by the caller.
func buildDeadlineContext(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if timeout <= noTimeout {
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
