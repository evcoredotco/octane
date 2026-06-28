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

const (
	// defaultManSection is the default man page section number.
	defaultManSection = 1

	// outDirPerm is the permission mode for the man page output directory.
	// 0o755 is intentional: man pages are public documentation.
	outDirPerm = 0o755
)

// newGenManPagesCmd constructs and returns the hidden "gen-manpages"
// subcommand. root is the root cobra command passed to doc.GenManTree
// so that man pages are generated for the full command tree.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newGenManPagesCmd(root *cobra.Command) *cobra.Command {
	var section int

	var outDir string

	cmd := &cobra.Command{
		Use:    "gen-manpages",
		Short:  "Generate roff man pages from cobra command tree (hidden)",
		Hidden: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runGenManPages(c, args, root, outDir, section)
		},
	}

	flags := cmd.Flags()

	flags.IntVar(
		&section,
		"section",
		defaultManSection,
		"man section number (currently only section 1 is generated from cobra)",
	)

	flags.StringVar(
		&outDir,
		"out",
		emptyFlagValue,
		"output directory; created if it does not exist (required)",
	)

	err := cmd.MarkFlagRequired("out")
	if err != nil {
		// MarkFlagRequired only errors when the flag name is unknown,
		// which would be a programming error caught at construction time.
		panic(err)
	}

	return cmd
}

// runGenManPages is the RunE function for the hidden gen-manpages
// subcommand. It creates the output directory and delegates to
// cobra/doc.GenManTree.
func runGenManPages(
	_ *cobra.Command,
	_ []string,
	root *cobra.Command,
	outDir string,
	section int,
) error {
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

	err = doc.GenManTree(root, header, outDir)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: gen-manpages: %v\n", err)
	}

	return nil
}
