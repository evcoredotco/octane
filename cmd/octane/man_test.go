// T-006-51: golden man-page test.
// Run with -update to generate or refresh the testdata/man/ golden files.
package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra/doc"
)

// updateGolden controls whether the golden files are regenerated
// instead of compared. Pass -update on the command line.
var updateGolden = flag.Bool("update", false, "update golden files")

// Test_octane_ManGolden generates Section 1 man pages from the cobra
// command tree and compares them against the golden files in
// testdata/man/. Run with -update to regenerate the golden files.
func Test_octane_ManGolden(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("testdata/man"); os.IsNotExist(err) && !*updateGolden {
		t.Skip(
			"testdata/man not found; run with -update to create golden files",
		)
	}

	tmpDir := t.TempDir()

	header := &doc.GenManHeader{ //nolint:exhaustruct // Date/Source/Manual are optional; cobra fills them
		Title:   "OCTANE",
		Section: "1",
	}

	if err := doc.GenManTree(rootCmd, header, tmpDir); err != nil {
		t.Fatalf("GenManTree: %v", err)
	}

	goldenDir := filepath.Join("testdata", "man")

	if *updateGolden {
		//nolint:gosec // G301: golden dir is a checked-in test fixture; 0755 is intentional
		if err := os.MkdirAll(goldenDir, 0o755); err != nil { //nolint:mnd // conventional dir perms
			t.Fatalf("mkdir %q: %v", goldenDir, err)
		}

		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("read tmpDir: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			generatedPath := filepath.Join(tmpDir, entry.Name())

			//nolint:gosec // G304: path constructed from t.TempDir() + DirEntry.Name()
			data, readErr := os.ReadFile(generatedPath)
			if readErr != nil {
				t.Fatalf("read generated file %q: %v", entry.Name(), readErr)
			}

			dest := filepath.Join(goldenDir, entry.Name())

			//nolint:gosec // G306: golden files are public test fixtures; 0644 is intentional
			if writeErr := os.WriteFile(dest, data, 0o644); writeErr != nil { //nolint:mnd // conventional file perms
				t.Fatalf("write golden file %q: %v", dest, writeErr)
			}
		}

		t.Log("golden files updated")

		return
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read tmpDir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		generatedPath := filepath.Join(tmpDir, name)

		//nolint:gosec // G304: path constructed from t.TempDir() + DirEntry.Name()
		generated, readErr := os.ReadFile(generatedPath)
		if readErr != nil {
			t.Errorf("read generated %q: %v", name, readErr)

			continue
		}

		goldenPath := filepath.Join(goldenDir, name)

		//nolint:gosec // G304: path constructed from "testdata/man" + DirEntry.Name()
		golden, readGoldenErr := os.ReadFile(goldenPath)
		if readGoldenErr != nil {
			t.Errorf(
				"golden file %q missing; run with -update to create it",
				name,
			)

			continue
		}

		if string(generated) != string(golden) {
			t.Errorf(
				"man page %q differs from golden; run with -update to refresh",
				name,
			)
		}
	}
}
