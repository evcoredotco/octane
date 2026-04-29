// T-006-50: hidden gen-manpages subcommand invoked by `make man`.
// It generates Section 1 roff man pages for every cobra subcommand
// via github.com/spf13/cobra/doc.

package main

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
)

// genManPagesFlags holds the parsed flag values for the gen-manpages
// hidden subcommand.
var genManPagesFlags struct {
	section int
	outDir  string
}

//nolint:exhaustruct // cobra.Command has many optional fields
var genManPagesCmd = &cobra.Command{
	Use:    "gen-manpages",
	Short:  "Generate roff man pages from cobra command tree (hidden)",
	Hidden: true,
	RunE:   runGenManPages,
}

func init() {
	flags := genManPagesCmd.Flags()

	flags.IntVar(
		&genManPagesFlags.section,
		"section",
		1,
		"man section number (currently only section 1 is generated from cobra)",
	)

	flags.StringVar(
		&genManPagesFlags.outDir,
		"out",
		"",
		"output directory; created if it does not exist (required)",
	)

	if err := genManPagesCmd.MarkFlagRequired("out"); err != nil {
		// MarkFlagRequired only errors when the flag name is unknown,
		// which would be a programming error caught at init time.
		panic(err)
	}

	rootCmd.AddCommand(genManPagesCmd)
}

// runGenManPages is the RunE function for the hidden gen-manpages
// subcommand. It creates the output directory and delegates to
// cobra/doc.GenManTree.
func runGenManPages(_ *cobra.Command, _ []string) error {
	outDir := genManPagesFlags.outDir
	section := genManPagesFlags.section

	//nolint:gosec // G301: outDir is operator-supplied; 0755 is intentional for public man directories
	if err := os.MkdirAll(outDir, 0o755); err != nil { //nolint:mnd // 0755 is conventional dir perms
		dieErr(
			exitcode.ToolError,
			"octane: gen-manpages: mkdir %q: %v\n",
			outDir,
			err,
		)

		return nil
	}

	header := &doc.GenManHeader{ //nolint:exhaustruct // Date/Source/Manual are optional; cobra fills them
		Title:   "OCTANE",
		Section: strconv.Itoa(section),
	}

	if err := doc.GenManTree(rootCmd, header, outDir); err != nil {
		dieErr(exitcode.ToolError, "octane: gen-manpages: %v\n", err)
	}

	return nil
}
