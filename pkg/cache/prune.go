package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// noMaxAge indicates that age-based pruning is disabled (maxAge == 0).
const noMaxAge = 0

// emptyDirLen is the entry count that identifies an empty directory.
const emptyDirLen = 0

// Prune removes cache entries that have exceeded their maximum age
// or whose TTL has expired, then removes any empty fanout
// directories under <dir>/results/.
//
// The walk visits every result.json under <dir>/results/ and
// removes the containing entry directory when either condition
// holds:
//
//   - The entry's WrittenAt plus maxAge is before time.Now().
//   - The entry's TTL is non-zero and has expired per
//     [Entry.IsExpired].
//
// If result.json is unreadable or its JSON is malformed, the entry
// directory is removed (treat corrupt entries as expired).
//
// Prune is safe to call while [Get] and [Put] are in use: a reader
// will observe either the full result.json or nothing (never a
// partial file) because [Put] uses atomic rename. A prune that
// removes an entry directory between a [Get]'s ReadFile call and
// any subsequent use of the returned [Entry] has no ill effect.
func (fc *FileCache) Prune(
	ctx context.Context,
	maxAge time.Duration,
) error {
	resultsDir := filepath.Join(fc.dir, "results")
	now := fc.clk.Now()

	// Walk fanout prefix directories (two-character, e.g. "ab/").
	fanouts, err := os.ReadDir(resultsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cache: prune: read results dir: %w", err)
	}

	for _, fanout := range fanouts {
		err = pruneOneFanout(ctx, fanout, resultsDir, now, maxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

// pruneOneFanout checks ctx cancellation and delegates to [pruneFanout]
// for a single fanout directory entry. Extracted to keep [Prune]'s
// cognitive complexity within the configured limit.
func pruneOneFanout(
	ctx context.Context,
	fanout os.DirEntry,
	resultsDir string,
	now time.Time,
	maxAge time.Duration,
) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("cache: prune: %w", err)
	}

	return pruneFanout(fanout, resultsDir, now, maxAge)
}

// pruneFanout prunes a single fanout directory entry. Non-directory
// entries are skipped. After pruning expired entries inside
// fanoutEntry, the fanout directory itself is removed when empty.
func pruneFanout(
	fanoutEntry os.DirEntry,
	resultsDir string,
	now time.Time,
	maxAge time.Duration,
) error {
	if !fanoutEntry.IsDir() {
		return nil
	}

	fanoutPath := filepath.Join(resultsDir, fanoutEntry.Name())

	err := pruneEntriesUnder(fanoutPath, now, maxAge)
	if err != nil {
		return err
	}

	// Remove the fanout directory if it is now empty.
	if dirIsEmpty(fanoutPath) {
		_ = os.Remove(fanoutPath)
	}

	return nil
}

// pruneEntriesUnder removes expired entry directories directly
// inside fanoutDir. It reads each entry's result.json to obtain
// WrittenAt and TTL. Directories that cannot be parsed are removed.
func pruneEntriesUnder(
	fanoutDir string,
	now time.Time,
	maxAge time.Duration,
) error {
	entries, err := os.ReadDir(fanoutDir)
	if err != nil {
		return fmt.Errorf(
			"cache: prune: read fanout dir %q: %w",
			fanoutDir,
			err,
		)
	}

	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}

		entryDir := filepath.Join(fanoutDir, ent.Name())

		err = pruneEntryDir(entryDir, now, maxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

// pruneEntryDir removes entryDir when it should be pruned, ignoring
// [os.ErrNotExist] (concurrent prune or already removed).
func pruneEntryDir(
	entryDir string,
	now time.Time,
	maxAge time.Duration,
) error {
	if !shouldPrune(entryDir, now, maxAge) {
		return nil
	}

	err := os.RemoveAll(entryDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(
			"cache: prune: remove entry %q: %w",
			entryDir,
			err,
		)
	}

	return nil
}

// shouldPrune reports whether the entry at entryDir should be
// removed. It reads result.json and checks:
//
//  1. WrittenAt + maxAge < now  (age-based pruning).
//  2. TTL expired per [Entry.IsExpired]  (TTL invalidation).
//
// Corrupt or missing result.json returns true (prune the entry).
func shouldPrune(entryDir string, now time.Time, maxAge time.Duration) bool {
	resultPath := filepath.Join(entryDir, "result.json")

	data, err := cacheReadFile(resultPath)
	if err != nil {
		// Missing or unreadable result — prune.
		return true
	}

	var env resultEnvelope

	err = json.Unmarshal(data, &env)
	if err != nil {
		// Corrupt JSON — prune.
		return true
	}

	// Age-based pruning: entry is older than maxAge.
	if maxAge > noMaxAge && now.After(env.WrittenAt.Add(maxAge)) {
		return true
	}

	// TTL-based pruning: re-use Entry.IsExpired for consistency
	// with the Get path.
	entry := &Entry{
		Result:    []byte(env.Result),
		Trace:     nil,
		WrittenAt: env.WrittenAt,
		TTL:       time.Duration(env.TTLSeconds) * time.Second,
	}

	return entry.IsExpired(now)
}

// dirIsEmpty reports whether dir contains no entries.
func dirIsEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	return len(entries) == emptyDirLen
}
