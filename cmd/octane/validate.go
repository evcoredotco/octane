package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	"github.com/evcoreco/octane/pkg/story"
)

// noStoryPaths is the default search root when no paths are given.
const noStoryPaths = 0

// newValidateCmd constructs and returns the "octane validate" subcommand
// group, including the "octane validate stories" subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate story files and configuration",
	}

	cmd.AddCommand(newValidateStoriesCmd())

	return cmd
}

// newValidateStoriesCmd constructs and returns the "octane validate stories"
// subcommand.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newValidateStoriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stories [paths...]",
		Short: "Validate .story files for syntax and structural correctness",
		Long: `validate stories parses each .story file found at the given paths
and reports any syntax or structural errors.

Paths may be individual .story files or directories. Directories are
searched recursively for .story files.

Output:
  OK: <path>      — file is valid
  ERROR: <path>: <message> — file has a parse error

The command exits 0 when all files are valid, or 64 (config error)
when any file fails to parse.`,
		RunE: validateStories,
	}
}

// validateStories is the RunE function for "octane validate stories".
func validateStories(_ *cobra.Command, storyPaths []string) error {
	if len(storyPaths) == noStoryPaths {
		storyPaths = []string{"."}
	}

	storyFiles, err := collectAllStoryFiles(storyPaths)
	if err != nil {
		dieErrf(exitcode.ToolError, "octane: walk: %v\n", err)

		return nil
	}

	if anyFailed := parseAndReportFiles(storyFiles); anyFailed {
		dieErrf(exitcode.ConfigError, emptyFlagValue)
	}

	return nil
}

// collectAllStoryFiles collects .story files from all provided roots.
func collectAllStoryFiles(roots []string) ([]string, error) {
	var files []string

	for _, root := range roots {
		found, err := collectStoryFiles(root)
		if err != nil {
			return nil, fmt.Errorf("walk %q: %w", root, err)
		}

		files = append(files, found...)
	}

	return files, nil
}

// parseAndReportFiles parses each file and reports results to stdout/stderr.
// It returns true if any file failed to parse.
func parseAndReportFiles(paths []string) bool {
	anyFailed := false

	for _, path := range paths {
		if failed := parseAndReportFile(path); failed {
			anyFailed = true
		}
	}

	return anyFailed
}

// parseAndReportFile parses a single story file and writes the result to
// stdout (OK) or stderr (ERROR). It returns true when the file failed.
func parseAndReportFile(path string) bool {
	data, err := readStoryFile(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", path, err)

		return true
	}

	_, parseErr := story.Parse(path, data)
	if parseErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", path, parseErr)

		return true
	}

	_, _ = fmt.Fprintf(os.Stdout, "OK: %s\n", path)

	return false
}

// readStoryFile reads the story file at path. The path originates
// from CLI arguments and is therefore operator-controlled.
func readStoryFile(path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("validate: read story file: %w", err)
	}

	return data, nil
}

// collectStoryFiles recursively collects all .story file paths under
// root. If root is a file, it is returned as-is regardless of its
// extension (the caller determines which paths to validate). If root
// is a directory it is walked recursively.
func collectStoryFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", root, err)
	}

	if !info.IsDir() {
		return []string{root}, nil
	}

	return walkStoryDir(root)
}

// walkStoryDir walks dir recursively and returns all .story file paths.
func walkStoryDir(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(
		dir,
		func(path string, dirEntry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if !dirEntry.IsDir() && filepath.Ext(path) == ".story" {
				files = append(files, path)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("validate: walk stories: %w", err)
	}

	return files, nil
}
