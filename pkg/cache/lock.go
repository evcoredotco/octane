package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/evcoreco/octane/pkg/cache/internal/lock"
)

// ErrLockTimeout is returned by [AcquireLock] or [TryLock] when the
// caller has been unable to obtain the exclusive flock within the
// configured timeout, or immediately when the lock is already held.
//
// It re-exports the internal sentinel so that callers in pkg/runner
// do not need to import the internal lock package.
var ErrLockTimeout = lock.ErrLockTimeout

// AcquireLock obtains an exclusive flock on the file at lockPath,
// returning an [io.Closer] that releases the lock when closed.
//
// AcquireLock is the public surface of the double-checked acquire
// pattern described in ADR 0016 and ADR 0019. It delegates directly
// to the internal lock.Acquire implementation.
//
// Parameters:
//   - ctx: carries cancellation; a cancelled ctx causes an immediate
//     return of ctx.Err().
//   - lockPath: filesystem path to the lock file (e.g.
//     <cache-dir>/locks/<key_hash>.lock). The file is created if it
//     does not exist.
//   - timeout: maximum time to wait for the lock. Zero means use
//     only the ctx deadline.
//
// Returns [ErrLockTimeout] when the timeout fires. Any other error
// reflects an I/O failure. To fail immediately when the lock is
// already held, use [TryLock].
func AcquireLock(
	ctx context.Context,
	lockPath string,
	timeout time.Duration,
) (io.Closer, error) {
	closer, err := lock.Acquire(ctx, lockPath, timeout)
	if err != nil {
		if errors.Is(err, lock.ErrLockTimeout) {
			return nil, ErrLockTimeout
		}

		return nil, fmt.Errorf("cache: acquire lock: %w", err)
	}

	return closer, nil
}

// TryLock attempts a single non-blocking flock on the file at
// lockPath. If the lock is already held, [ErrLockTimeout] is
// returned immediately.
//
// TryLock is the non-blocking counterpart of [AcquireLock].
func TryLock(
	ctx context.Context,
	lockPath string,
) (io.Closer, error) {
	closer, err := lock.TryAcquire(ctx, lockPath)
	if err != nil {
		if errors.Is(err, lock.ErrLockTimeout) {
			return nil, ErrLockTimeout
		}

		return nil, fmt.Errorf("cache: try lock: %w", err)
	}

	return closer, nil
}
