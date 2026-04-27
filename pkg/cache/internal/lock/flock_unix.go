//go:build !windows

package lock

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// lockFile opens (or creates) the file at path and acquires an
// exclusive, non-blocking flock on it.
//
// The lock file is opened with O_CREATE|O_RDWR and permissions
// 0o600 so it is readable and writable only by the owning user,
// matching the cache directory permission policy (ADR 0016 §
// "Lock file layout").
//
// If the lock is held by another process, lockFile returns
// [ErrLockBusy]. Any other syscall error is wrapped and returned
// as-is. The caller MUST call [unlockFile] on the returned *os.File
// to release the lock and close the descriptor.
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

	//nolint:gosec // G115: uintptr→int conversion is the accepted syscall.Flock idiom
	flockErr := syscall.Flock(
		int(fileHandle.Fd()),
		syscall.LOCK_EX|syscall.LOCK_NB,
	)
	if flockErr != nil {
		_ = fileHandle.Close()

		if errors.Is(flockErr, syscall.EWOULDBLOCK) ||
			errors.Is(flockErr, syscall.EAGAIN) {
			return nil, ErrLockBusy
		}

		return nil, fmt.Errorf("lock: flock acquire: %w", flockErr)
	}

	return fileHandle, nil
}

// unlockFile releases the exclusive flock held on fileHandle and
// closes the file descriptor.
//
// The unlock is performed via LOCK_UN before Close so that another
// waiting process can acquire the lock before the fd is recycled.
// If both Flock and Close fail, the Flock error takes precedence.
//
// This function is called by [lockCloser.Close] in acquire.go (T-005-32).
func unlockFile(fileHandle *os.File) error {
	//nolint:gosec // G115: uintptr→int conversion is the accepted syscall.Flock idiom
	flockErr := syscall.Flock(int(fileHandle.Fd()), syscall.LOCK_UN)
	closeErr := fileHandle.Close()

	if flockErr != nil {
		return fmt.Errorf("lock: flock release: %w", flockErr)
	}

	if closeErr != nil {
		return fmt.Errorf("lock: close lock file: %w", closeErr)
	}

	return nil
}
