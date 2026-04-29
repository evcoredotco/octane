package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	_ "github.com/evcoreco/octane/pkg/keywords/primitive" // register primitive keywords
)

// version is the binary version string, injected at build time by goreleaser
// via -X main.version={{.Version}}. Defaults to "dev" in local builds.
var version = "dev"

// globalFlags holds the parsed values of the persistent global flags
// declared on the root command. They are set by cobra's flag binding
// before any RunE function executes.
var globalFlags struct {
	configPath string
	verbose    bool
	noCache    bool
	cacheDir   string
}

// rootCmd is the top-level cobra command. Every subcommand is
// registered as a child of rootCmd in their respective source files.
//
//nolint:exhaustruct // cobra.Command has many optional fields
var rootCmd = &cobra.Command{
	Use:   "octane",
	Short: "OCTANE — OCPP conformance test runner",
	Long: `octane runs .story conformance test suites against a CSMS endpoint,
validates story files, and manages the content-addressed result cache.

Global flags apply to all subcommands. Use "octane help <command>"
for subcommand-specific documentation.`,
}

func init() {
	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVar(
		&globalFlags.configPath,
		"config",
		"octane.yml",
		"path to octane.yml configuration file",
	)

	persistentFlags.BoolVarP(
		&globalFlags.verbose,
		"verbose",
		"v",
		false,
		"enable verbose output",
	)

	persistentFlags.BoolVar(
		&globalFlags.noCache,
		"no-cache",
		false,
		"bypass the result cache entirely",
	)

	persistentFlags.StringVar(
		&globalFlags.cacheDir,
		"cache-dir",
		"",
		"override the cache directory (default: $XDG_CACHE_HOME/octane/cache)",
	)
}

// Execute runs the root cobra command and exits with the appropriate
// process exit code on error. It is called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitcode.Exec(exitcode.ToolError)
	}
}

// dieErr prints a formatted error message to stderr and exits the
// process with the given exit code. It is the canonical way for
// RunE functions to report fatal errors without returning them
// through cobra (which would print an additional usage hint).
func dieErr(code int, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)

	exitcode.Exec(code)
}
