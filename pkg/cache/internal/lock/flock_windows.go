//go:build windows

package lock

import (
	"fmt"
	"os"
)

// lockFile opens (or creates) the file at path and attempts to
// acquire an exclusive lock on it.
//
// STUB: A full implementation using the Windows LockFileEx API
// requires golang.org/x/sys/windows, which is not currently in
// go.mod. Adding that dependency requires an ADR per the
// constitution (principle V). Until that ADR is merged and
// golang.org/x/sys is added, this stub always returns
// [ErrLockBusy] so that callers fall back to the no-lock path
// rather than silently skipping the lock.
//
// To implement the real LockFileEx path:
//  1. Draft an ADR proposing golang.org/x/sys as a dependency.
//  2. Run: go get golang.org/x/sys/windows
//  3. Replace this stub with a call to windows.LockFileEx using
//     LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY flags.
//
// This function is called by [Acquire] in acquire.go (T-005-32).
func lockFile(path string) (*os.File, error) {
	//nolint:gosec // G304: path is derived from the cache root + key hash, not user input
	fileHandle, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR,
		0o600,
	)
	if err != nil {
		return nil, fmt.Errorf("lock: open lock file: %w", err)
	}

	_ = fileHandle.Close()

	// Real LockFileEx implementation requires golang.org/x/sys/windows.
	// See the stub comment above for the upgrade path.
	return nil, ErrLockBusy
}

// unlockFile releases the lock held on fileHandle and closes it.
//
// STUB: Because [lockFile] always returns [ErrLockBusy] on Windows,
// this function is never called in practice. It is provided so that
// acquire.go compiles on Windows without a build-tag guard on the
// call sites.
//
// This function is called by [lockCloser.Close] in acquire.go (T-005-32).
func unlockFile(fileHandle *os.File) error {
	err := fileHandle.Close()
	if err != nil {
		return fmt.Errorf("lock: close lock file: %w", err)
	}

	return nil
}
