package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
)

// newCompletionCmd constructs and returns the "octane completion" subcommand.
// root is the root cobra command used to generate the completion scripts;
// it is captured via closure so the completion handler can call the
// appropriate Gen*Completion method.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for octane.

To load completions in the current shell session:

  Bash:
    source <(octane completion bash)

  Zsh:
    source <(octane completion zsh)

  Fish:
    octane completion fish | source

  PowerShell:
    octane completion powershell | Out-String | Invoke-Expression`,
		Args:      cobra.ExactArgs(exactlyOneArg),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateCompletion(cmd, args, root)
		},
	}
}

// generateCompletion is the RunE function for "octane completion".
// It delegates to cobra's built-in shell completion generators.
func generateCompletion(
	_ *cobra.Command,
	args []string,
	root *cobra.Command,
) error {
	shell := args[firstArgIndex]

	var err error

	switch shell {
	case "bash":
		err = root.GenBashCompletion(os.Stdout)
	case "zsh":
		err = root.GenZshCompletion(os.Stdout)
	case "fish":
		err = root.GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = root.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		dieErrf(
			exitcode.ConfigError,
			"octane: unsupported shell %q;"+
				" valid values: bash, zsh, fish, powershell\n",
			shell,
		)

		return nil
	}

	if err != nil {
		dieErrf(exitcode.ToolError, "octane: generate completion: %v\n", err)
	}

	return nil
}
