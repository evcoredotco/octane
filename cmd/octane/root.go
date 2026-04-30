package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	// Side-effect import: registers primitive keywords at init time.
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
)

// version is the binary version string, injected at build time by goreleaser
// via -X main.version={{.Version}}. Defaults to "dev" in local builds.
var version = "dev"

// globalFlagsT holds the parsed values of the persistent global flags
// declared on the root command. They are set by cobra's flag binding
// before any RunE function executes.
type globalFlagsT struct {
	configPath string
	verbose    bool
	noCache    bool
	cacheDir   string
}

// exitPanic is the sentinel value used by dieErrf to signal a controlled
// process exit through main's recover handler. Using panic/recover keeps
// os.Exit confined to func main() only, satisfying the revive deep-exit rule.
type exitPanic struct {
	code int
}

// newRootCmd constructs and returns the root cobra command with all
// persistent flags and subcommands wired up. Every invocation returns
// an independent command tree; there are no package-level command
// globals.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newRootCmd() *cobra.Command {
	flags := &globalFlagsT{}

	cmd := &cobra.Command{
		Use:   "octane",
		Short: "OCTANE — OCPP conformance test runner",
		Long: `octane runs .story suites against a CSMS endpoint, validates
story files, and manages the content-addressed result cache.

Global flags apply to all subcommands. Use "octane help <command>"
for subcommand-specific documentation.`,
	}

	persistentFlags := cmd.PersistentFlags()

	persistentFlags.StringVar(
		&flags.configPath,
		"config",
		"octane.yml",
		"path to octane.yml configuration file",
	)

	persistentFlags.BoolVarP(
		&flags.verbose,
		"verbose",
		"v",
		false,
		"enable verbose output",
	)

	persistentFlags.BoolVar(
		&flags.noCache,
		"no-cache",
		false,
		"bypass the result cache entirely",
	)

	persistentFlags.StringVar(
		&flags.cacheDir,
		"cache-dir",
		emptyFlagValue,
		"override the cache directory (default: $XDG_CACHE_HOME/octane/cache)",
	)

	cmd.AddCommand(newRunCmd(flags))
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newCacheCmd(flags))
	cmd.AddCommand(newCompletionCmd(cmd))
	cmd.AddCommand(newKeywordsCmd())
	cmd.AddCommand(newGenManPagesCmd(cmd))

	return cmd
}

// Execute runs the root cobra command. On error it panics with an
// exitPanic so that main's recover handler can call os.Exit with the
// correct exit code. It is the public entry point called from main.
func Execute() {
	root := newRootCmd()

	err := root.Execute()
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: %v\n", err)
	}
}

// dieErrf prints a formatted error message to stderr and exits the
// process with the given exit code. It is the canonical way for
// RunE functions to report fatal errors without returning them
// through cobra (which would print an additional usage hint).
//
// Internally it panics with an exitPanic that main's recover handler
// catches, keeping os.Exit confined to func main() only
// (revive deep-exit rule).
func dieErrf(code int, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)

	panic(exitPanic{code: code})
}
