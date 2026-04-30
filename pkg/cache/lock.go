package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/evcoreco/octane/pkg/cache/internal/lock"
)

// ErrLockTimeout is returned by [AcquireLock] when the caller has
// been unable to obtain the exclusive flock within the configured
// timeout or immediately when noWait is true and the lock is held.
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
//   - noWait: when true, fail immediately if the lock is held.
//
// Returns [ErrLockTimeout] when the timeout or noWait condition
// fires. Any other error reflects an I/O failure.
func AcquireLock(
	ctx context.Context,
	lockPath string,
	timeout time.Duration,
	noWait bool,
) (io.Closer, error) {
	closer, err := lock.Acquire(ctx, lockPath, timeout, noWait)
	if err != nil {
		if errors.Is(err, lock.ErrLockTimeout) {
			return nil, ErrLockTimeout
		}

		return nil, fmt.Errorf("cache: acquire lock: %w", err)
	}

	return closer, nil
}
