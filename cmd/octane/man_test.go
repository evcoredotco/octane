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

// updateManGoldens regenerates the golden files under goldenDir from
// the generated man pages in tmpDir. Each non-directory entry is read
// and written to goldenDir.
func updateManGoldens(t *testing.T, tmpDir, goldenDir string) {
	t.Helper()

	//nolint:gosec // G301: golden dir is a checked-in test fixture; 0755 is intentional
	mkdirErr := os.MkdirAll(goldenDir, 0o755)
	if mkdirErr != nil {
		t.Fatalf("mkdir %q: %v", goldenDir, mkdirErr)
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
		writeErr := os.WriteFile(dest, data, 0o644)
		if writeErr != nil {
			t.Fatalf("write golden file %q: %v", dest, writeErr)
		}
	}

	t.Log("golden files updated")
}

// compareManGoldens checks each generated man page in tmpDir against the
// corresponding golden file in goldenDir, reporting mismatches via t.Errorf.
func compareManGoldens(t *testing.T, tmpDir, goldenDir string) {
	t.Helper()

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read tmpDir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		//nolint:gosec // G304: path constructed from t.TempDir() + DirEntry.Name()
		generated, readErr := os.ReadFile(filepath.Join(tmpDir, name))
		if readErr != nil {
			t.Errorf("read generated %q: %v", name, readErr)

			continue
		}

		//nolint:gosec // G304: path constructed from "testdata/man" + DirEntry.Name()
		golden, readGoldenErr := os.ReadFile(filepath.Join(goldenDir, name))
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

// Test_octane_ManGolden generates Section 1 man pages from the cobra
// command tree and compares them against the golden files in
// testdata/man/. Run with -update to regenerate the golden files.
func Test_octane_ManGolden(t *testing.T) {
	t.Parallel()

	_, statErr := os.Stat("testdata/man")
	if os.IsNotExist(statErr) && !*updateGolden {
		t.Fatal(
			"testdata/man not found; run: go test -run Test_octane_ManGolden -update ./cmd/octane/",
		)
	}

	tmpDir := t.TempDir()

	header := &doc.GenManHeader{ //nolint:exhaustruct // Date/Source/Manual are optional; cobra fills them
		Title:   "OCTANE",
		Section: "1",
	}

	genErr := doc.GenManTree(rootCmd, header, tmpDir)
	if genErr != nil {
		t.Fatalf("GenManTree: %v", genErr)
	}

	goldenDir := filepath.Join("testdata", "man")

	if *updateGolden {
		updateManGoldens(t, tmpDir, goldenDir)

		return
	}

	compareManGoldens(t, tmpDir, goldenDir)
}
