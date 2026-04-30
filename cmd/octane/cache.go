package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	"github.com/evcoreco/octane/pkg/cache"
)

// dirMode is the permission bits for directories created by cache commands.
const dirMode = 0o750

//nolint:exhaustruct // cobra.Command has many optional fields
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the content-addressed result cache",
}

//nolint:exhaustruct // cobra.Command has many optional fields
var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print the cache directory location",
	RunE:  cacheInfo,
}

// cachePruneFlags holds the flags for "octane cache prune".
var cachePruneFlags struct {
	maxAge time.Duration
}

//nolint:exhaustruct // cobra.Command has many optional fields
var cachePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove cache entries older than --max-age",
	Long: `prune removes cache entries whose WrittenAt timestamp plus max-age
is before now, and entries whose TTL has expired.

Empty fanout directories are removed after pruning.`,
	RunE: cachePrune,
}

//nolint:exhaustruct // cobra.Command has many optional fields
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all entries from the result cache",
	Long: `clear removes the contents of the results/ subdirectory in the
cache directory, effectively invalidating all cached test results.

The cache directory structure itself (version.json, locks/) is
preserved; only result entries are removed.`,
	RunE: cacheClear,
}

//nolint:exhaustruct // cobra.Command has many optional fields
var cacheKeyCmd = &cobra.Command{
	Use:   "key <story-id>",
	Short: "Print the cache key hash for a story ID",
	Long: `key prints the SHA-256 hex digest that octane uses as the cache key
for the given story ID with placeholder values for the remaining
key components (CSMS endpoint, story content, parameters).

This is useful for locating cached entries on the filesystem or
for debugging cache invalidation behaviour.`,
	Args: cobra.ExactArgs(1),
	RunE: cacheKey,
}

// defaultMaxAge is the default value for the --max-age flag used by
// "octane cache prune". Entries older than this are removed by default.
const defaultMaxAge = 24 * time.Hour

func init() {
	cachePruneCmd.Flags().DurationVar(
		&cachePruneFlags.maxAge,
		"max-age",
		defaultMaxAge,
		"maximum age of cache entries to keep (e.g. 24h, 7d)",
	)

	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cachePruneCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheKeyCmd)
	rootCmd.AddCommand(cacheCmd)
}

// resolveCacheDirForCLI returns the effective cache directory using
// the --cache-dir global flag, then $XDG_CACHE_HOME/octane/cache/,
// then $HOME/.cache/octane/cache/ as fallbacks.
func resolveCacheDirForCLI() (string, error) {
	if globalFlags.cacheDir != "" {
		return globalFlags.cacheDir, nil
	}

	if envDir := os.Getenv("OCTANE_CACHE_DIR"); envDir != "" {
		return envDir, nil
	}

	if xdgHome := os.Getenv("XDG_CACHE_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "octane", "cache"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "octane", "cache"), nil
}

// cacheInfo is the RunE function for "octane cache info".
func cacheInfo(_ *cobra.Command, _ []string) error {
	cacheDir, err := resolveCacheDirForCLI()
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: %v\n", err)

		return nil
	}

	_, _ = fmt.Fprintf(os.Stdout, "cache dir: %s\n", cacheDir)

	return nil
}

// cachePrune is the RunE function for "octane cache prune".
func cachePrune(_ *cobra.Command, _ []string) error {
	cacheDir, err := resolveCacheDirForCLI()
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: %v\n", err)

		return nil
	}

	cacheStore, err := cache.Open(cacheDir)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: open cache: %v\n", err)

		return nil
	}

	err = cacheStore.Prune(context.Background(), cachePruneFlags.maxAge)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: prune cache: %v\n", err)

		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "pruned cache")

	return nil
}

// cacheClear is the RunE function for "octane cache clear".
func cacheClear(_ *cobra.Command, _ []string) error {
	cacheDir, err := resolveCacheDirForCLI()
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: %v\n", err)

		return nil
	}

	resultsDir := filepath.Join(cacheDir, "results")

	err = os.RemoveAll(resultsDir)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: remove results dir: %v\n", err)

		return nil
	}

	err = os.MkdirAll(resultsDir, dirMode)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: recreate results dir: %v\n", err)

		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "cache cleared")

	return nil
}

// cacheKey is the RunE function for "octane cache key".
func cacheKey(_ *cobra.Command, args []string) error {
	storyID := args[0]

	// Use placeholder SHAs as documented in pkg/runner/run.go
	// buildCacheKey. The real values require the CSMS endpoint
	// (spec 002), story content hash (spec 001), and parameter
	// hash (spec 003) which are not yet available at the CLI layer.
	key := cache.Key{
		TestID:          storyID,
		ScopeKey:        "",
		CSMSEndpointSHA: "00000000",
		OctaneVersion:   "dev",
		OCPPVersion:     "unknown",
		StoryContentSHA: "00000000",
		ParameterSHA:    "00000000",
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", key.Hash())

	return nil
}
