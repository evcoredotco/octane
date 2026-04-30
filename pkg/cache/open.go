package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evcoreco/octane/pkg/engine/clock"
)

const (
	// schemaVersion is the current schema version written into
	// version.json by [Open]. Readers compare this value against
	// the supported range to detect incompatible cache directories.
	schemaVersion = 1

	// cacheDirPerm is the permission bits for cache sub-directories.
	cacheDirPerm = 0o750
)

// versionStamp is the structure written to version.json at
// cache-open time. It lets operators and future OCTANE versions
// detect schema incompatibilities before reading entries.
type versionStamp struct {
	// SchemaVersion is the integer schema version of this
	// cache directory. Currently always 1.
	SchemaVersion int `json:"schemaVersion"`

	// CreatedAt is the RFC 3339 timestamp when the cache
	// directory was first initialised.
	CreatedAt time.Time `json:"createdAt"`
}

// Open creates (or verifies) a content-addressed cache directory
// rooted at dir and returns a [Cache] implementation backed by
// that directory tree.
//
// Open creates the following sub-directories if they do not exist:
//
//	<dir>/results/   — two-character fanout directories live here
//	<dir>/locks/     — flock target files (Phase 4, T-005-30)
//
// A version stamp file (<dir>/version.json) is written on the
// first call; subsequent calls leave the existing stamp intact.
//
// Open returns an error if dir cannot be created, if the directory
// tree is not writable, or if the version stamp cannot be written.
func Open(dir string) (*FileCache, error) {
	subDirs := []string{
		filepath.Join(dir, "results"),
		filepath.Join(dir, "locks"),
	}

	for _, sub := range subDirs {
		err := os.MkdirAll(sub, cacheDirPerm)
		if err != nil {
			return nil, fmt.Errorf(
				"cache: create directory %q: %w",
				sub,
				err,
			)
		}
	}

	err := writeVersionStamp(dir)
	if err != nil {
		return nil, err
	}

	return &FileCache{dir: dir, clk: clock.Real()}, nil
}

// OpenWithClock is identical to [Open] but uses the supplied clock
// for all TTL checks and WrittenAt timestamps inside [FileCache].
// Inject a [clock.DeterministicClock] in tests that need precise
// control over cache expiry without real wall-clock delay
// (constitution principle IV).
func OpenWithClock(dir string, clk clock.Clock) (*FileCache, error) {
	subDirs := []string{
		filepath.Join(dir, "results"),
		filepath.Join(dir, "locks"),
	}

	for _, sub := range subDirs {
		err := os.MkdirAll(sub, cacheDirPerm)
		if err != nil {
			return nil, fmt.Errorf(
				"cache: create directory %q: %w",
				sub,
				err,
			)
		}
	}

	err := writeVersionStamp(dir)
	if err != nil {
		return nil, err
	}

	return &FileCache{dir: dir, clk: clk}, nil
}

// writeVersionStamp writes version.json into dir if the file does
// not already exist. If the file is present (cache was opened
// before), writeVersionStamp is a no-op.
func writeVersionStamp(dir string) error {
	versionPath := filepath.Join(dir, "version.json")

	_, statErr := os.Stat(versionPath)
	if statErr == nil {
		// File already exists; leave it untouched.
		return nil
	}

	stamp := versionStamp{
		SchemaVersion: schemaVersion,
		CreatedAt:     time.Now().UTC(),
	}

	data, err := json.Marshal(stamp)
	if err != nil {
		return fmt.Errorf("cache: marshal version stamp: %w", err)
	}

	err = atomicWriteFile(versionPath, data)
	if err != nil {
		return fmt.Errorf("cache: write version stamp: %w", err)
	}

	return nil
}
