package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evcoreco/octane/pkg/engine/clock"
)

// cacheDirMode is the permission bits for cache entry directories (rwxr-x---).
const cacheDirMode = 0o750

// cacheFileMode is the permission bits for cache data files (rw-------).
const cacheFileMode = 0o600

// resultEnvelope is the JSON structure persisted as result.json.
// It wraps [Entry] fields with metadata required by ADR 0016
// §"Result file schema".
//
// The WrittenAt and TTLSeconds fields enable TTL invalidation
// (spec 005 AC10) without storing a Go-specific duration type in
// the file.
type resultEnvelope struct {
	// SchemaVersion is always [schemaVersion] (currently 1).
	SchemaVersion int `json:"schemaVersion"`

	// WrittenAt is the UTC wall-clock time when this entry was
	// persisted. Used for TTL checks and age-based pruning.
	WrittenAt time.Time `json:"writtenAt"`

	// TTLSeconds is the maximum age of this entry in seconds
	// before it is considered stale. Zero means no TTL.
	TTLSeconds int64 `json:"ttlSeconds"`

	// Result holds the raw JSON bytes of the test result. It MUST
	// NOT contain credentials (constitution principle X).
	Result json.RawMessage `json:"result"`

	// TracePresent indicates whether a sibling trace.json file
	// was written for this entry.
	TracePresent bool `json:"tracePresent"`
}

// FileCache is the content-addressed file tree implementation of
// [Cache]. Each entry is stored as one or two JSON files in a
// two-character fanout directory under <dir>/results/.
//
// FileCache is not safe for concurrent use without external
// synchronisation. Phase 4 (T-005-30 through T-005-33) adds
// flock-based locking on top of this struct.
type FileCache struct {
	// dir is the root of the cache directory tree, as received
	// from [Open]. All paths are derived from this root.
	dir string

	// clk is the clock used for TTL expiry checks (Get) and
	// WrittenAt timestamps (Put). Defaults to clock.Real();
	// inject a deterministic clock in tests via [OpenWithClock].
	clk clock.Clock
}

// Get retrieves the cache entry for the given key.
//
// Get computes the key hash, reads
// <dir>/results/<hash[:2]>/<hash>/result.json, unmarshals the
// envelope, checks schema version, and then delegates to
// [Entry.IsExpired] for TTL invalidation (spec 005 AC10).
//
// It returns [ErrCacheMiss] when:
//   - the result file does not exist (normal miss, AC8),
//   - the JSON is malformed (treat corrupt entry as a miss),
//   - the schema version is not supported,
//   - the entry's TTL has expired (AC10).
//
// Any other I/O error is returned as-is.
func (fc *FileCache) Get(
	ctx context.Context,
	key Key,
) (*Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("cache: Get: %w", err)
	}

	hash := key.Hash()
	resultPath := filepath.Join(fc.resultDir(hash), "result.json")

	env, err := readResultEnvelope(resultPath)
	if err != nil {
		return nil, err
	}

	entry := &Entry{
		Result:    []byte(env.Result),
		Trace:     nil,
		WrittenAt: env.WrittenAt,
		TTL:       time.Duration(env.TTLSeconds) * time.Second,
	}

	// TTL invalidation: spec 005 AC10.
	if entry.IsExpired(fc.clk.Now()) {
		return nil, ErrCacheMiss
	}

	// Read the optional trace file. A missing trace is non-fatal;
	// a partially-restored cache (AC8) may be missing trace files
	// while result.json is intact.
	if env.TracePresent {
		entry.Trace = readOptionalTrace(fc.resultDir(hash))
	}

	return entry, nil
}

// readResultEnvelope reads and decodes result.json at resultPath.
// It translates os.ErrNotExist and JSON parse failures to
// [ErrCacheMiss]. Any other I/O error is returned as-is.
func readResultEnvelope(resultPath string) (*resultEnvelope, error) {
	data, err := cacheReadFile(resultPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrCacheMiss
		}

		return nil, fmt.Errorf("cache: read result.json: %w", err)
	}

	var env resultEnvelope

	if err = json.Unmarshal(data, &env); err != nil {
		// Treat corrupt entry as a cache miss so the runner
		// re-executes and overwrites the bad file.
		return nil, ErrCacheMiss
	}

	if env.SchemaVersion != schemaVersion {
		return nil, ErrCacheMiss
	}

	return &env, nil
}

// readOptionalTrace attempts to read trace.json from entryDir.
// A missing or unreadable file returns nil (non-fatal per AC8).
func readOptionalTrace(entryDir string) []byte {
	tracePath := filepath.Join(entryDir, "trace.json")

	data, err := cacheReadFile(tracePath)
	if err != nil {
		return nil
	}

	return data
}

// Put writes a cache entry for the given key using the atomic
// temp-file-and-rename protocol defined in ADR 0016 §"Atomic
// writes" and spec 005 §10:
//
//  1. Write result.json.tmp (fsync).
//  2. Rename to result.json (atomic on POSIX).
//  3. fsync the directory.
//
// If entry.Trace is non-nil, trace.json is written with the same
// protocol before result.json, matching ADR 0016 §"Wire trace
// splitting".
//
// Callers are responsible for ensuring that entry.Result and
// entry.Trace contain no credentials (constitution principle X).
func (fc *FileCache) Put(
	ctx context.Context,
	key Key,
	entry Entry,
) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("cache: Put: %w", err)
	}

	hash := key.Hash()
	dir := fc.resultDir(hash)

	err = os.MkdirAll(dir, cacheDirMode)
	if err != nil {
		return fmt.Errorf("cache: create entry dir: %w", err)
	}

	tracePresent := len(entry.Trace) > 0

	if tracePresent {
		err := atomicWriteFile(
			filepath.Join(dir, "trace.json"),
			entry.Trace,
		)
		if err != nil {
			return fmt.Errorf("cache: write trace.json: %w", err)
		}
	}

	writtenAt := resolveWrittenAt(entry.WrittenAt, fc.clk.Now)

	env := buildEnvelope(entry, writtenAt, tracePresent)

	return writeEnvelope(dir, env)
}

// resolveWrittenAt returns provided if non-zero, otherwise calls now() and
// returns the result in UTC. This keeps Put's branch count low.
func resolveWrittenAt(provided time.Time, now func() time.Time) time.Time {
	if !provided.IsZero() {
		return provided
	}

	return now().UTC()
}

// buildEnvelope constructs the [resultEnvelope] for the given entry,
// resolved writtenAt timestamp, and tracePresent flag.
func buildEnvelope(
	entry Entry,
	writtenAt time.Time,
	tracePresent bool,
) resultEnvelope {
	ttlSecs := int64(0)
	if entry.TTL > 0 {
		ttlSecs = int64(entry.TTL.Seconds())
	}

	return resultEnvelope{
		SchemaVersion: schemaVersion,
		WrittenAt:     writtenAt,
		TTLSeconds:    ttlSecs,
		Result:        json.RawMessage(entry.Result),
		TracePresent:  tracePresent,
	}
}

// writeEnvelope marshals env to JSON and atomically writes it as
// result.json inside dir, then fsyncs the directory.
func writeEnvelope(dir string, env resultEnvelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("cache: marshal result envelope: %w", err)
	}

	resultPath := filepath.Join(dir, "result.json")

	if err = atomicWriteFile(resultPath, data); err != nil {
		return fmt.Errorf("cache: write result.json: %w", err)
	}

	if err = fsyncDir(dir); err != nil {
		return fmt.Errorf("cache: fsync entry dir: %w", err)
	}

	return nil
}

// resultDir returns the directory path for a given hash, including
// the two-character fanout prefix:
//
//	<dir>/results/<hash[:2]>/<hash>/
func (fc *FileCache) resultDir(hash string) string {
	return filepath.Join(fc.dir, "results", hash[:2], hash)
}

// cacheReadFile reads a cache file at the given path.
//
// All paths passed to this function are derived from the cache
// root directory plus a SHA-256 hash component; they are never
// constructed from unvalidated user input. The gosec G304
// suppression is intentional: the caller guarantees the path is
// cache-internal.
func cacheReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: cache path
	if err != nil {
		return nil, fmt.Errorf("cache: read file: %w", err)
	}

	return data, nil
}

// atomicWriteFile writes data to path using the atomic
// temp-file-and-rename protocol:
//
//  1. Write data to path+".tmp".
//  2. fsync the temp file.
//  3. Rename path+".tmp" to path (atomic on POSIX; uses
//     MoveFileEx on Windows via os.Rename).
//
// The containing directory is fsynced by the caller ([Put] and
// [writeVersionStamp]).
func atomicWriteFile(path string, data []byte) error {
	tmp := path + ".tmp"

	//nolint:gosec // G304: tmp path is from a cache path, not user input
	file, err := os.OpenFile(
		tmp,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		cacheFileMode,
	)
	if err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}

	_, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf("write temp file: %w", writeErr)
	}

	if syncErr != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf("fsync temp file: %w", syncErr)
	}

	if closeErr != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf("close temp file: %w", closeErr)
	}

	err = os.Rename(tmp, path)
	if err != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// fsyncDir opens dir and calls Sync() on it to flush the directory
// entry to durable storage, completing the atomic-write protocol
// for the directory itself (step 4 of spec 005 §10).
//
// On platforms where directory Sync is unsupported (e.g. Windows),
// the error from Sync is silently ignored to preserve
// cross-platform compatibility.
func fsyncDir(dir string) error {
	//nolint:gosec // G304: dir is the cache root directory, not user input
	dirHandle, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open dir for fsync: %w", err)
	}

	// Sync error on directories is non-fatal on non-Linux platforms.
	_ = dirHandle.Sync()

	err = dirHandle.Close()
	if err != nil {
		return fmt.Errorf("close dir after fsync: %w", err)
	}

	return nil
}
