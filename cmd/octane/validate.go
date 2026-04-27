package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/octane-project/octane/cmd/octane/internal/exitcode"
	"github.com/octane-project/octane/pkg/story"
)

//nolint:exhaustruct // cobra.Command has many optional fields
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate story files and configuration",
}

//nolint:exhaustruct // cobra.Command has many optional fields
var validateStoriesCmd = &cobra.Command{
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

func init() {
	validateCmd.AddCommand(validateStoriesCmd)
	rootCmd.AddCommand(validateCmd)
}

// validateStories is the RunE function for "octane validate stories".
func validateStories(_ *cobra.Command, storyPaths []string) error {
	if len(storyPaths) == 0 {
		storyPaths = []string{"."}
	}

	var storyFiles []string

	for _, root := range storyPaths {
		found, err := collectStoryFiles(root)
		if err != nil {
			dieErr(exitcode.IOError, "octane: walk %q: %v\n", root, err)

			return nil
		}

		storyFiles = append(storyFiles, found...)
	}

	anyFailed := false

	for _, path := range storyFiles {
		data, err := readStoryFile(path)
		if err != nil {
			anyFailed = true
			_, _ = fmt.Fprintf(os.Stdout, "ERROR: %s: %v\n", path, err)

			continue
		}

		if _, parseErr := story.Parse(path, data); parseErr != nil {
			anyFailed = true
			_, _ = fmt.Fprintf(os.Stdout, "ERROR: %s: %v\n", path, parseErr)

			continue
		}

		_, _ = fmt.Fprintf(os.Stdout, "OK: %s\n", path)
	}

	if anyFailed {
		exitcode.Exec(exitcode.ConfigError)
	}

	return nil
}

// readStoryFile reads the story file at path. The path originates
// from CLI arguments and is therefore operator-controlled; the gosec
// G304 warning is intentionally suppressed.
func readStoryFile(path string) ([]byte, error) {
	return os.ReadFile(
		path,
	) //nolint:gosec // G304: path from CLI args; operator-controlled
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

	var files []string

	err = filepath.WalkDir(
		root,
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

	return files, err
}
