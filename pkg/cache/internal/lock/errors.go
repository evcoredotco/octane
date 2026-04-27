// Package lock provides a platform-specific flock-based exclusive
// lock for individual cache key paths.
//
// The public surface consumed by acquire.go (T-005-32) is two
// unexported functions:
//
//   - lockFile(path string) (*os.File, error)
//   - unlockFile(f *os.File) error
//
// These functions are provided by build-tag-selected files:
//   - flock_unix.go    — Linux / macOS  (//go:build !windows)
//   - flock_windows.go — Windows        (//go:build windows)
//
// [ErrLockBusy] is the sentinel returned when the lock file is
// held by another process and the caller should retry or wait.
package lock

import "errors"

// ErrLockBusy is returned by lockFile when the target lock file
// is already held by another process. The caller (acquire.go,
// T-005-32) should either wait-and-retry or propagate the error
// to the runner as a lock-timeout failure.
var ErrLockBusy = errors.New("lock: file is busy")
