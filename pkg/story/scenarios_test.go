// Package story_test — black-box validation suite for the story parser
// (T-001-40, T-001-41).
//
// TestScenariosParseClean covers AC8: every story under scenarios/ must parse
// without error. TestScenariosGolden compares serialized ASTs against golden
// JSON fixtures, catching unintended AST structure changes.

package story_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/evcoreco/octane/pkg/story"
)

const (
	scenariosDir  = "scenarios"
	parentDir     = ".."
	noFiles       = 0
	goldenDirPerm = 0o750
	storyFileMode = 0o600
)

// shouldUpdateGoldens reports whether the -update flag was passed. It reads
// the flag value at call time so that no package-level variable is needed.
func shouldUpdateGoldens() bool {
	f := flag.Lookup("update")
	if f == nil {
		return false
	}

	return f.Value.String() == "true"
}

func TestMain(m *testing.M) {
	flag.Bool("update", false, "regenerate golden fixture files")
	m.Run()
}

// TestScenariosParseClean asserts that every .story file under scenarios/
// parses without error (AC8). Failures here are blockers: they indicate
// either a broken story file or a parser regression.
func TestScenariosParseClean(t *testing.T) {
	t.Parallel()

	paths := collectStoryPaths(
		t,
		filepath.Join(parentDir, parentDir, scenariosDir),
	)

	for _, path := range paths {
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			t.Parallel()
			assertParsesClean(t, path)
		})
	}
}

// assertParsesClean reads and parses a single .story file, failing the test
// on any error.
func assertParsesClean(t *testing.T, path string) {
	t.Helper()

	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if _, parseErr := story.Parse(path, src); parseErr != nil {
		t.Errorf("parse %s: %v", path, parseErr)
	}
}

// TestScenariosGolden parses each .story file under scenarios/ and compares
// the indented JSON serialization of the resulting AST against a golden file
// stored at pkg/story/testdata/scenarios/<rel-path>.golden.json.
//
// On the first run (no golden file present) the fixture is written
// automatically. Pass -update to force-regenerate all goldens.
func TestScenariosGolden(t *testing.T) {
	t.Parallel()

	scenariosRoot := filepath.Join(parentDir, parentDir, scenariosDir)
	goldenRoot := filepath.Join("testdata", scenariosDir)

	paths := collectStoryPaths(t, scenariosRoot)

	for _, path := range paths {
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			t.Parallel()
			runGoldenCheck(t, path, scenariosRoot, goldenRoot)
		})
	}
}

// runGoldenCheck performs the golden fixture comparison for a single story
// file. It is extracted to keep TestScenariosGolden's cognitive complexity
// within the configured limit.
func runGoldenCheck(
	t *testing.T,
	path string,
	scenariosRoot string,
	goldenRoot string,
) {
	t.Helper()

	got := parseAndMarshal(t, path)

	rel, err := filepath.Rel(scenariosRoot, path)
	if err != nil {
		t.Fatalf("rel path for %s: %v", path, err)
	}

	goldenPath := filepath.Join(goldenRoot, rel+".golden.json")

	compareWithGolden(t, path, goldenPath, got)
}

// parseAndMarshal parses path and marshals the AST to indented JSON.
func parseAndMarshal(t *testing.T, path string) []byte {
	t.Helper()

	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	parsed, err := story.Parse(path, src)
	if err != nil {
		t.Fatalf("parse %s: %v (run TestScenariosParseClean first)", path, err)
	}

	got, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}

	return got
}

// compareWithGolden compares got against the golden file. It writes or
// regenerates the golden when requested or when the file does not yet exist.
func compareWithGolden(
	t *testing.T,
	path string,
	goldenPath string,
	got []byte,
) {
	t.Helper()

	if shouldUpdateGoldens() {
		writeGolden(t, goldenPath, got)

		return
	}

	existing, readErr := os.ReadFile(filepath.Clean(goldenPath))
	if os.IsNotExist(readErr) {
		writeGolden(t, goldenPath, got)

		return
	}

	if readErr != nil {
		t.Fatalf("read golden %s: %v", goldenPath, readErr)
	}

	if string(existing) != string(got) {
		t.Errorf(
			"golden mismatch for %s:\n"+
				"  golden file: %s\n"+
				"  want: %s\n"+
				"  got:  %s",
			path, goldenPath, existing, got,
		)
	}
}

// collectStoryPaths walks root and returns the paths of all .story files.
// It calls t.Fatal if the walk fails or yields no files.
func collectStoryPaths(t *testing.T, root string) []string {
	t.Helper()

	var paths []string

	err := filepath.WalkDir(
		root,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(path) == ".story" {
				paths = append(paths, path)
			}

			return nil
		},
	)
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}

	if len(paths) == noFiles {
		t.Fatalf("no .story files found under %s", root)
	}

	return paths
}

// writeGolden writes data to path, creating parent directories as needed.
// Permissions 0750/0600 are used for directory/file to satisfy gosec G301/G306.
func writeGolden(t *testing.T, path string, data []byte) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), goldenDirPerm)
	if err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	err = os.WriteFile(filepath.Clean(path), data, storyFileMode)
	if err != nil {
		t.Fatalf("write golden %s: %v", path, err)
	}
}
