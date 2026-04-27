package config_test

import (
	"testing"
	"time"

	"github.com/evcoreco/octane/cmd/octane/internal/config"
)

// ptr is a generic helper that returns a pointer to the given value.
// It is used to construct FlagOverrides without intermediate
// variables throughout this test file.
func ptr[T any](value T) *T {
	return &value
}

// TestResolve_FlagWinsOverEnv asserts that a non-nil flag override
// takes precedence over an environment-variable value.
func TestResolve_FlagWinsOverEnv(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.CacheDir = "/env/cache" // simulate env-var layer result

	flags := config.FlagOverrides{
		CacheDir:           ptr("/flag/cache"),
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.CacheDir != "/flag/cache" {
		t.Errorf(
			"expected CacheDir=%q, got %q",
			"/flag/cache",
			resolved.CacheDir,
		)
	}
}

// TestResolve_EnvWinsOverYAML asserts that env-var values (already
// applied to cfg before Resolve is called) take precedence over
// YAML-sourced values when no flag override is provided.
func TestResolve_EnvWinsOverYAML(t *testing.T) {
	t.Parallel()

	// Simulate a config loaded from YAML with MaxParallel=4, then
	// overridden by the env-var layer to 8.
	base := config.Default()
	base.MaxParallel = 8 // env-var layer already applied

	// No flag override for MaxParallel.
	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.MaxParallel != 8 {
		t.Errorf(
			"expected MaxParallel=8, got %d",
			resolved.MaxParallel,
		)
	}
}

// TestResolve_YAMLWinsOverDefault asserts that a value read from the
// YAML file takes precedence over the built-in default when neither
// env-var nor flag overrides are present.
func TestResolve_YAMLWinsOverDefault(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.OCPPVersion = "1.6" // YAML-sourced value

	// No flag override for OCPPVersion.
	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.OCPPVersion != "1.6" {
		t.Errorf(
			"expected OCPPVersion=%q, got %q",
			"1.6",
			resolved.OCPPVersion,
		)
	}
}

// TestResolve_DefaultUsedWhenNoOverride asserts that default values
// are preserved when no YAML, env-var, or flag override is present.
func TestResolve_DefaultUsedWhenNoOverride(t *testing.T) {
	t.Parallel()

	base := config.Default()

	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.LockTimeout != 60*time.Second {
		t.Errorf(
			"expected LockTimeout=%v, got %v",
			60*time.Second,
			resolved.LockTimeout,
		)
	}

	if resolved.FailOn != "any" {
		t.Errorf(
			"expected FailOn=%q, got %q",
			"any",
			resolved.FailOn,
		)
	}

	if resolved.MaxParallel != 1 {
		t.Errorf(
			"expected MaxParallel=1, got %d",
			resolved.MaxParallel,
		)
	}
}

// TestResolve_NilFlagLeavesFieldUnchanged asserts that a nil flag
// pointer leaves the corresponding Config field unchanged.
func TestResolve_NilFlagLeavesFieldUnchanged(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.CacheDir = "/some/dir"

	flags := config.FlagOverrides{
		CacheDir:           nil, // not provided by operator
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.CacheDir != "/some/dir" {
		t.Errorf(
			"expected CacheDir=%q, got %q",
			"/some/dir",
			resolved.CacheDir,
		)
	}
}
