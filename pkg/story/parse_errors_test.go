// Package story_test — negative-case tests for the story parser.
// Each test loads a fixture from testdata/negative/ and asserts that
// the parser returns the correct typed error (AC2–AC7).

package story_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/evcoreco/octane/pkg/story"
	"github.com/evcoreco/octane/pkg/story/diag"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "negative", name)

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}

	return src
}

// TestNegative_MissingName covers AC2: missing required Meta key "Name".
func TestNegative_MissingName(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "missing_name.story")

	_, err := story.Parse("missing_name.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMissingKey
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMissingKey, got %T: %v", err, err)
	}

	if target.Key != "Name" {
		t.Errorf("Key = %q, want %q", target.Key, "Name")
	}
}

// TestNegative_MissingID covers AC2: missing required Meta key "Id".
func TestNegative_MissingID(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "missing_id.story")

	_, err := story.Parse("missing_id.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMissingKey
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMissingKey, got %T: %v", err, err)
	}

	if target.Key != "Id" {
		t.Errorf("Key = %q, want %q", target.Key, "Id")
	}
}

// TestNegative_MissingStations covers AC2: missing required Meta key "Stations".
func TestNegative_MissingStations(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "missing_stations.story")

	_, err := story.Parse("missing_stations.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMissingKey
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMissingKey, got %T: %v", err, err)
	}

	if target.Key != "Stations" {
		t.Errorf("Key = %q, want %q", target.Key, "Stations")
	}
}

// TestNegative_SpecRefOnHelper covers AC4: helper story with Spec-Ref present.
func TestNegative_SpecRefOnHelper(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "spec_ref_on_helper.story")

	_, err := story.Parse("spec_ref_on_helper.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrSpecRefOnHelper
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrSpecRefOnHelper, got %T: %v", err, err)
	}
}

// TestNegative_MissingSpecRef covers AC3: conformance story without Spec-Ref.
func TestNegative_MissingSpecRef(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "missing_spec_ref.story")

	_, err := story.Parse("missing_spec_ref.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMissingSpecRef
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMissingSpecRef, got %T: %v", err, err)
	}
}

// TestNegative_MalformedDependsNoID covers AC6: Depends entry missing id field.
func TestNegative_MalformedDependsNoID(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "malformed_depends_no_id.story")

	_, err := story.Parse("malformed_depends_no_id.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMalformedDepends
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMalformedDepends, got %T: %v", err, err)
	}
}

// TestNegative_MalformedDependsBadScope covers AC6: Depends entry with unknown scope.
func TestNegative_MalformedDependsBadScope(t *testing.T) {
	t.Parallel()

	src := readFixture(t, "malformed_depends_bad_scope.story")

	_, err := story.Parse("malformed_depends_bad_scope.story", src)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var target *diag.ErrMalformedDepends
	if !errors.As(err, &target) {
		t.Fatalf("expected *diag.ErrMalformedDepends, got %T: %v", err, err)
	}
}
