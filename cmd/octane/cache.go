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

// defaultMaxAge is the default value for the --max-age flag used by
// "octane cache prune". Entries older than this are removed by default.
const defaultMaxAge = 24 * time.Hour

// placeholderSHA is the placeholder hash used by "octane cache key"
// when the real inputs are not available at the CLI layer.
const placeholderSHA = "00000000"

// errFmtOctane is the single-argument error format used throughout cache
// subcommand handlers.
const errFmtOctane = "octane: %v\n"

// cacheDirName is the directory name component used in XDG cache paths.
const cacheDirName = "cache"

// newCacheCmd constructs and returns the "octane cache" subcommand group.
// globalFlags is the parent global-flags struct used to resolve the
// cache directory.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCacheCmd(globalFlags *globalFlagsT) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the content-addressed result cache",
	}

	cmd.AddCommand(newCacheInfoCmd(globalFlags))
	cmd.AddCommand(newCachePruneCmd(globalFlags))
	cmd.AddCommand(newCacheClearCmd(globalFlags))
	cmd.AddCommand(newCacheKeyCmd())

	return cmd
}

// newCacheInfoCmd constructs the "octane cache info" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCacheInfoCmd(globalFlags *globalFlagsT) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print the cache directory location",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cacheInfo(cmd, args, globalFlags)
		},
	}
}

// newCachePruneCmd constructs the "octane cache prune" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCachePruneCmd(globalFlags *globalFlagsT) *cobra.Command {
	var maxAge time.Duration

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove cache entries older than --max-age",
		Long: `prune removes entries whose WrittenAt plus max-age is before now,
and entries whose TTL has expired.

Empty fanout directories are removed after pruning.`,
		RunE: func(c *cobra.Command, args []string) error {
			return cachePrune(c, args, globalFlags, maxAge)
		},
	}

	cmd.Flags().DurationVar(
		&maxAge,
		"max-age",
		defaultMaxAge,
		"maximum age of cache entries to keep (e.g. 24h, 7d)",
	)

	return cmd
}

// newCacheClearCmd constructs the "octane cache clear" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCacheClearCmd(globalFlags *globalFlagsT) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove all entries from the result cache",
		Long: `clear removes the contents of the results/ subdirectory in the
cache directory, effectively invalidating all cached test results.

The cache directory structure itself (version.json, locks/) is
preserved; only result entries are removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cacheClear(cmd, args, globalFlags)
		},
	}
}

// newCacheKeyCmd constructs the "octane cache key" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCacheKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "key <story-id>",
		Short: "Print the cache key hash for a story ID",
		Long: `key prints the SHA-256 hex digest octane uses as the cache key
for the given story ID with placeholder values for the remaining
key components (CSMS endpoint, story content, parameters).

This is useful for locating cached entries on the filesystem or
for debugging cache invalidation behaviour.`,
		Args: cobra.ExactArgs(exactlyOneArg),
		RunE: cacheKey,
	}
}

// resolveCacheDirForCLI returns the effective cache directory using
// the --cache-dir global flag, then $OCTANE_CACHE_DIR,
// then $XDG_CACHE_HOME/octane/cache/, then $HOME/.cache/octane/cache/
// as fallbacks.
func resolveCacheDirForCLI(globalFlags *globalFlagsT) (string, error) {
	if globalFlags.cacheDir != emptyFlagValue {
		return globalFlags.cacheDir, nil
	}

	if envDir := os.Getenv("OCTANE_CACHE_DIR"); envDir != emptyFlagValue {
		return envDir, nil
	}

	if xdgHome := os.Getenv("XDG_CACHE_HOME"); xdgHome != emptyFlagValue {
		return filepath.Join(xdgHome, "octane", cacheDirName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "octane", cacheDirName), nil
}

// cacheInfo is the RunE function for "octane cache info".
func cacheInfo(_ *cobra.Command, _ []string, globalFlags *globalFlagsT) error {
	cacheDir, err := resolveCacheDirForCLI(globalFlags)
	if err != nil {
		dieErrf(exitcode.ToolError, errFmtOctane, err)

		return nil
	}

	_, _ = fmt.Fprintf(os.Stdout, "cache dir: %s\n", cacheDir)

	return nil
}

// cachePrune is the RunE function for "octane cache prune".
func cachePrune(
	_ *cobra.Command,
	_ []string,
	globalFlags *globalFlagsT,
	maxAge time.Duration,
) error {
	cacheDir, err := resolveCacheDirForCLI(globalFlags)
	if err != nil {
		dieErrf(exitcode.ToolError, errFmtOctane, err)

		return nil
	}

	cacheStore, err := cache.Open(cacheDir)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: open cache: %v\n", err)

		return nil
	}

	err = cacheStore.Prune(context.Background(), maxAge)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: prune cache: %v\n", err)

		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "pruned cache")

	return nil
}

// cacheClear is the RunE function for "octane cache clear".
func cacheClear(_ *cobra.Command, _ []string, globalFlags *globalFlagsT) error {
	cacheDir, err := resolveCacheDirForCLI(globalFlags)
	if err != nil {
		dieErrf(exitcode.ToolError, errFmtOctane, err)

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
	storyID := args[firstArgIndex]

	// Use placeholder SHAs as documented in pkg/runner/run.go
	// buildCacheKey. The real values require the CSMS endpoint
	// (spec 002), story content hash (spec 001), and parameter
	// hash (spec 003) which are not yet available at the CLI layer.
	key := cache.Key{
		TestID:          storyID,
		ScopeKey:        emptyFlagValue,
		CSMSEndpointSHA: placeholderSHA,
		OctaneVersion:   "dev",
		OCPPVersion:     "unknown",
		StoryContentSHA: placeholderSHA,
		ParameterSHA:    placeholderSHA,
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", key.Hash())

	return nil
}
