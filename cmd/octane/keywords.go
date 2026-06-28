package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// ocppVersionAny is the zero value for api.OCPPVersion, meaning that
// keyword resolution is not restricted to a particular OCPP version.
// Primitive keywords (the only ones registered without a domain import)
// are always eligible at this version value.
const ocppVersionAny api.OCPPVersion = 0

// newKeywordsCmd constructs and returns the "octane keywords" subcommand
// group, including "list" and "resolve" subcommands.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newKeywordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keywords",
		Short: "Inspect registered keywords",
	}

	cmd.AddCommand(newKeywordsListCmd())
	cmd.AddCommand(newKeywordsResolveCmd())

	return cmd
}

// newKeywordsListCmd constructs the "octane keywords list" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newKeywordsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Print all registered keywords",
		Long: `list prints every keyword registered in the keyword registry,
sorted by layer (primitive then domain), OCPP version, and pattern.

Each line is formatted as:
  [<layer>] [<ocpp-version>] <pattern>`,
		RunE: keywordsList,
	}
}

// newKeywordsResolveCmd constructs the "octane keywords resolve" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newKeywordsResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <step-text>",
		Short: "Resolve a step text to a keyword pattern",
		Long: `resolve matches <step-text> against registered keywords
and prints the matched pattern and extracted arguments.

If no keyword matches, the command prints a "no match" message with
the closest suggestion (if any) and exits 0.`,
		Args: cobra.ExactArgs(exactlyOneArg),
		RunE: keywordsResolve,
	}
}

// keywordsList is the RunE function for "octane keywords list".
func keywordsList(_ *cobra.Command, _ []string) error {
	all := registry.All()

	if len(all) == zeroIntDefault {
		_, _ = fmt.Fprintln(os.Stdout, "(no keywords registered)")

		return nil
	}

	for _, keyword := range all {
		_, _ = fmt.Fprintf(
			os.Stdout,
			"[%s] [%s] %s\n",
			keyword.Layer,
			keyword.OCPPVersion,
			keyword.Pattern,
		)
	}

	return nil
}

// keywordsResolve is the RunE function for "octane keywords resolve".
// It resolves the given step text against registered keywords and
// prints the matched pattern, layer, and OCPP version, or a "no
// match" message with the closest suggestion when resolution fails.
func keywordsResolve(_ *cobra.Command, args []string) error {
	stepText := args[firstArgIndex]

	// Resolve with ocppVersionAny; primitive keywords (the
	// only ones registered without a domain import) are always
	// eligible. Domain keywords require a version filter — pass
	// ocppVersionAny to enumerate all version-agnostic primitives.
	match, err := registry.Resolve(stepText, ocppVersionAny)
	if err != nil {
		var noMatch *registry.NoMatchError

		if errors.As(err, &noMatch) {
			_, _ = fmt.Fprintf(os.Stdout, "no match for: %q\n", stepText)

			if noMatch.Closest != "" {
				_, _ = fmt.Fprintf(os.Stdout, "closest: %q\n", noMatch.Closest)
			}

			return nil
		}

		return fmt.Errorf("resolve: %w", err)
	}

	_, _ = fmt.Fprintf(
		os.Stdout,
		"pattern: %q\nlayer:   %s\nversion: %s\n",
		match.Keyword.Pattern,
		match.Keyword.Layer,
		match.Keyword.OCPPVersion,
	)

	return nil
}
