// T-006-51: golden man-page test.
// Set the environment variable UPDATE_GOLDEN=1 to regenerate the
// testdata/man/ golden files instead of comparing them.

package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra/doc"
)

const (
	// goldenDirPerm is the permission mode for the test golden directory.
	// 0o755 is intentional: test fixture dirs are public.
	goldenDirPerm = 0o755

	// goldenFilePerm is the permission mode for the test golden files.
	// 0o644 is intentional: test fixtures are public.
	goldenFilePerm = 0o644
)

// updateGolden returns true when the UPDATE_GOLDEN environment variable
// is set to "1" or "true", indicating that golden files should be
// regenerated rather than compared.
func updateGolden() bool {
	v := os.Getenv("UPDATE_GOLDEN")

	return v == "1" || v == "true"
}

// updateManGoldens regenerates the golden files under goldenDir from
// the generated man pages in tmpDir. Each non-directory entry is read
// and written to goldenDir.
func updateManGoldens(t *testing.T, tmpDir, goldenDir string) {
	t.Helper()

	mkdirErr := os.MkdirAll(goldenDir, goldenDirPerm)
	if mkdirErr != nil {
		t.Fatalf("mkdir %q: %v", goldenDir, mkdirErr)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read tmpDir: %v", err)
	}

	for _, entry := range entries {
		copyManGolden(t, tmpDir, goldenDir, entry)
	}

	t.Log("golden files updated")
}

// copyManGolden copies a single generated man page file from tmpDir to
// goldenDir. It skips directory entries.
func copyManGolden(t *testing.T, tmpDir, goldenDir string, entry fs.DirEntry) {
	t.Helper()

	if entry.IsDir() {
		return
	}

	generatedPath := filepath.Join(tmpDir, entry.Name())

	//nolint:gosec // G304: path from t.TempDir() + DirEntry.Name()
	data, readErr := os.ReadFile(generatedPath)
	if readErr != nil {
		t.Fatalf("read generated file %q: %v", entry.Name(), readErr)
	}

	dest := filepath.Join(goldenDir, entry.Name())

	//nolint:gosec // G306: public test fixtures; goldenFilePerm intentional
	writeErr := os.WriteFile(dest, data, goldenFilePerm)
	if writeErr != nil {
		t.Fatalf("write golden file %q: %v", dest, writeErr)
	}
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
		compareManGolden(t, tmpDir, goldenDir, entry)
	}
}

// compareManGolden compares a single generated man page in tmpDir against
// the corresponding golden file in goldenDir. It skips directory entries.
func compareManGolden(
	t *testing.T,
	tmpDir, goldenDir string,
	entry fs.DirEntry,
) {
	t.Helper()

	if entry.IsDir() {
		return
	}

	name := entry.Name()

	//nolint:gosec // G304: path from t.TempDir() + DirEntry.Name()
	generated, readErr := os.ReadFile(filepath.Join(tmpDir, name))
	if readErr != nil {
		t.Errorf("read generated %q: %v", name, readErr)

		return
	}

	//nolint:gosec // G304: path from "testdata/man" + DirEntry.Name()
	golden, readGoldenErr := os.ReadFile(filepath.Join(goldenDir, name))
	if readGoldenErr != nil {
		t.Errorf(
			"golden file %q missing; run UPDATE_GOLDEN=1 to create it",
			name,
		)

		return
	}

	if string(generated) != string(golden) {
		t.Errorf(
			"man page %q differs from golden;"+
				" run UPDATE_GOLDEN=1 to refresh",
			name,
		)
	}
}

// Test_octane_ManGolden generates Section 1 man pages from the cobra
// command tree and compares them against the golden files in
// testdata/man/. Set UPDATE_GOLDEN=1 to regenerate the golden files.
func Test_octane_ManGolden(t *testing.T) {
	t.Parallel()

	_, statErr := os.Stat("testdata/man")
	if os.IsNotExist(statErr) && !updateGolden() {
		t.Fatal(
			"testdata/man not found;" +
				" run with UPDATE_GOLDEN=1 to create",
		)
	}

	tmpDir := t.TempDir()

	header := new(doc.GenManHeader)
	header.Title = "OCTANE"
	header.Section = "1"

	root := newRootCmd()

	genErr := doc.GenManTree(root, header, tmpDir)
	if genErr != nil {
		t.Fatalf("GenManTree: %v", genErr)
	}

	goldenDir := filepath.Join("testdata", "man")

	if updateGolden() {
		updateManGoldens(t, tmpDir, goldenDir)

		return
	}

	compareManGoldens(t, tmpDir, goldenDir)
}
