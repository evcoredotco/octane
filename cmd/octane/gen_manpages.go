// T-006-50: hidden gen-manpages subcommand invoked by `make man`.
// It generates Section 1 roff man pages for every cobra subcommand
// via github.com/spf13/cobra/doc.

package main

import (
	"os"
	"strconv"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	// defaultManSection is the default man page section number.
	defaultManSection = 1

	// outDirPerm is the permission mode for the man page output directory.
	// 0o755 is intentional: man pages are public documentation.
	outDirPerm = 0o755
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
		defaultManSection,
		"man section number (currently only section 1 is generated from cobra)",
	)

	flags.StringVar(
		&genManPagesFlags.outDir,
		"out",
		"",
		"output directory; created if it does not exist (required)",
	)

	err := genManPagesCmd.MarkFlagRequired("out")
	if err != nil {
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

	err := os.MkdirAll(outDir, outDirPerm)
	if err != nil {
		dieErrf(
			exitcode.ToolError,
			"octane: gen-manpages: mkdir %q: %v\n",
			outDir,
			err,
		)

		return nil
	}

	header := new(doc.GenManHeader)
	header.Title = "OCTANE"
	header.Section = strconv.Itoa(section)

	err = doc.GenManTree(rootCmd, header, outDir)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: gen-manpages: %v\n", err)
	}

	return nil
}
