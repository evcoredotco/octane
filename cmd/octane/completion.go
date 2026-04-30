package main

import (
	"os"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	"github.com/spf13/cobra"
)

//nolint:exhaustruct // cobra.Command has many optional fields
var completionCmd = &cobra.Command{
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
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE:      generateCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

// generateCompletion is the RunE function for "octane completion".
// It delegates to cobra's built-in shell completion generators.
func generateCompletion(_ *cobra.Command, args []string) error {
	shell := args[0]

	var err error

	switch shell {
	case "bash":
		err = rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		err = rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		err = rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
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
