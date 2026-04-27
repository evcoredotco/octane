// T-006-53: completion syntax smoke tests.
// Verifies that the bash and zsh completion scripts produced by cobra
// are syntactically valid by running the shell's own -n checker.
package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

// Test_octane_CompletionBashSyntax captures the bash completion script
// produced by cobra and runs "bash -n" against it to confirm it is
// syntactically valid bash.
func Test_octane_CompletionBashSyntax(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}

	var buf bytes.Buffer

	if err := rootCmd.GenBashCompletion(&buf); err != nil {
		t.Fatalf("GenBashCompletion: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "octane-bash-completion-*.bash")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err = tmpFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("write completion to temp file: %v", err)
	}

	if err = tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	//nolint:gosec // G204: temp-file path is controlled by os.CreateTemp
	out, runErr := exec.Command("bash", "-n", tmpFile.Name()).CombinedOutput()
	if runErr != nil {
		t.Errorf("bash -n failed: %v\n%s", runErr, out)
	}
}

// Test_octane_CompletionZshSyntax captures the zsh completion script
// produced by cobra and runs "zsh -n" against it to confirm it is
// syntactically valid zsh.
func Test_octane_CompletionZshSyntax(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not on PATH")
	}

	var buf bytes.Buffer

	if err := rootCmd.GenZshCompletion(&buf); err != nil {
		t.Fatalf("GenZshCompletion: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "octane-zsh-completion-*.zsh")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err = tmpFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("write completion to temp file: %v", err)
	}

	if err = tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	//nolint:gosec // G204: temp-file path is controlled by os.CreateTemp
	out, runErr := exec.Command("zsh", "-n", tmpFile.Name()).CombinedOutput()
	if runErr != nil {
		t.Errorf("zsh -n failed: %v\n%s", runErr, out)
	}
}
