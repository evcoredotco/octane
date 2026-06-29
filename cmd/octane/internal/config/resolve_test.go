package config_test

import (
	"testing"
	"time"

	"github.com/evcoreco/octane/cmd/octane/internal/config"
)

const (
	// flagCacheDir is the cache directory path set via flag override.
	flagCacheDir = "/flag/cache"

	// maxParallelEnv is the MaxParallel value simulating an env-var layer.
	maxParallelEnv = 8

	// ocppVersion16 is the OCPP 1.6 version string used in config tests.
	ocppVersion16 = "1.6"

	// defaultLockTimeoutSec is the default LockTimeout in seconds.
	defaultLockTimeoutSec = 60

	// defaultMaxParallel is the default MaxParallel worker count.
	defaultMaxParallel = 1

	// preservedCacheDir is the CacheDir value tested for nil-flag preservation.
	preservedCacheDir = "/some/dir"
)

// TestResolve_FlagWinsOverEnv asserts that a non-nil flag override
// takes precedence over an environment-variable value.
func TestResolve_FlagWinsOverEnv(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.CacheDir = "/env/cache" // simulate env-var layer result

	cacheDirFlag := flagCacheDir

	flags := config.FlagOverrides{
		CacheDir:           &cacheDirFlag,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
		Parameters:         nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.CacheDir != flagCacheDir {
		t.Errorf(
			"expected CacheDir=%q, got %q",
			flagCacheDir,
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
	// overridden by the env-var layer to maxParallelEnv.
	base := config.Default()
	base.MaxParallel = maxParallelEnv // env-var layer already applied

	// No flag override for MaxParallel.
	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
		Parameters:         nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.MaxParallel != maxParallelEnv {
		t.Errorf(
			"expected MaxParallel=%d, got %d",
			maxParallelEnv,
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
	base.OCPPVersion = ocppVersion16 // YAML-sourced value

	// No flag override for OCPPVersion.
	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
		Parameters:         nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.OCPPVersion != ocppVersion16 {
		t.Errorf(
			"expected OCPPVersion=%q, got %q",
			ocppVersion16,
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
		Parameters:         nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.LockTimeout != defaultLockTimeoutSec*time.Second {
		t.Errorf(
			"expected LockTimeout=%v, got %v",
			defaultLockTimeoutSec*time.Second,
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

	if resolved.MaxParallel != defaultMaxParallel {
		t.Errorf(
			"expected MaxParallel=%d, got %d",
			defaultMaxParallel,
			resolved.MaxParallel,
		)
	}
}

// TestResolve_NilFlagLeavesFieldUnchanged asserts that a nil flag
// pointer leaves the corresponding Config field unchanged.
func TestResolve_NilFlagLeavesFieldUnchanged(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.CacheDir = preservedCacheDir

	flags := config.FlagOverrides{
		CacheDir:           nil, // not provided by operator
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
		Parameters:         nil,
	}

	resolved := config.Resolve(base, flags)

	if resolved.CacheDir != preservedCacheDir {
		t.Errorf(
			"expected CacheDir=%q, got %q",
			preservedCacheDir,
			resolved.CacheDir,
		)
	}
}

func TestResolve_ParameterOverridesMergeWithYAML(t *testing.T) {
	t.Parallel()

	base := config.Default()
	base.Parameters = map[string]string{
		"connectorId": "1",
		"valid_idTag": "YAML",
	}

	flags := config.FlagOverrides{
		CacheDir:           nil,
		MaxParallel:        nil,
		OCPPVersion:        nil,
		LockTimeout:        nil,
		FailOn:             nil,
		InsecureSkipVerify: nil,
		Parameters: map[string]string{
			"valid_idTag": "CLI",
			"meterStart":  "0",
		},
	}

	resolved := config.Resolve(base, flags)

	if resolved.Parameters["connectorId"] != "1" {
		t.Errorf("connectorId: want %q, got %q", "1", resolved.Parameters["connectorId"])
	}

	if resolved.Parameters["valid_idTag"] != "CLI" {
		t.Errorf(
			"valid_idTag: want %q, got %q",
			"CLI",
			resolved.Parameters["valid_idTag"],
		)
	}

	if resolved.Parameters["meterStart"] != "0" {
		t.Errorf("meterStart: want %q, got %q", "0", resolved.Parameters["meterStart"])
	}
}
