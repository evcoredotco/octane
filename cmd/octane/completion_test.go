// T-006-53: completion syntax smoke tests.
// Verifies that the bash and zsh completion scripts produced by cobra
// are syntactically valid by running the shell's own -n checker.

package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
)

// completionCase describes a single shell completion syntax check.
type completionCase struct {
	// shell is the shell binary to check PATH for and to invoke with -n.
	shell string
	// ext is the file extension for the temp file (e.g. ".bash", ".zsh").
	ext string
	// generate writes the cobra completion script to w.
	generate func(w io.Writer) error
}

// runCompletionSyntaxCheck is the shared helper for completion syntax smoke
// tests. It writes the completion script to a temp file and verifies it
// parses with "<shell> -n".
func runCompletionSyntaxCheck(t *testing.T, tcase completionCase) {
	t.Helper()

	_, lookPathErr := exec.LookPath(tcase.shell)
	if lookPathErr != nil {
		t.Skipf("%s not on PATH", tcase.shell)
	}

	var buf bytes.Buffer

	genErr := tcase.generate(&buf)
	if genErr != nil {
		t.Fatalf("generate %s completion: %v", tcase.shell, genErr)
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "octane-completion-*"+tcase.ext)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	_, err = tmpFile.Write(buf.Bytes())
	if err != nil {
		t.Fatalf("write completion to temp file: %v", err)
	}

	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	//nolint:gosec // G204: temp-file path is controlled by os.CreateTemp
	out, runErr := exec.CommandContext(t.Context(), tcase.shell, "-n", tmpFile.Name()).
		CombinedOutput()
	if runErr != nil {
		t.Errorf("%s -n failed: %v\n%s", tcase.shell, runErr, out)
	}
}

// Test_octane_CompletionBashSyntax captures the bash completion script
// produced by cobra and runs "bash -n" against it to confirm it is
// syntactically valid bash.
func Test_octane_CompletionBashSyntax(t *testing.T) {
	t.Parallel()

	runCompletionSyntaxCheck(t, completionCase{
		shell:    "bash",
		ext:      ".bash",
		generate: rootCmd.GenBashCompletion,
	})
}

// Test_octane_CompletionZshSyntax captures the zsh completion script
// produced by cobra and runs "zsh -n" against it to confirm it is
// syntactically valid zsh.
func Test_octane_CompletionZshSyntax(t *testing.T) {
	t.Parallel()

	runCompletionSyntaxCheck(t, completionCase{
		shell:    "zsh",
		ext:      ".zsh",
		generate: rootCmd.GenZshCompletion,
	})
}
