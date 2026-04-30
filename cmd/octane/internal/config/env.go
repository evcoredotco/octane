package config

import (
	"os"
	"strconv"
	"time"
)

// emptyEnv is the named empty-string constant for environment variable checks.
const emptyEnv = ""

// skipVerifyTrue is the string "true" accepted by OCTANE_INSECURE_SKIP_VERIFY.
const skipVerifyTrue = "true"

// skipVerifyOne is the value "1" accepted by OCTANE_INSECURE_SKIP_VERIFY.
const skipVerifyOne = "1"

// ApplyEnv overlays environment variable values on top of cfg and
// returns the updated Config. Fields whose corresponding environment
// variable is unset or empty are left unchanged.
//
// Supported variables:
//
//   - OCTANE_CACHE_DIR             → [Config.CacheDir]
//   - OCTANE_MAX_PARALLEL          → [Config.MaxParallel] (parsed as int)
//   - OCTANE_OCPP_VERSION          → [Config.OCPPVersion]
//   - OCTANE_LOCK_TIMEOUT          → [Config.LockTimeout]
//     (parsed as duration)
//   - OCTANE_FAIL_ON               → [Config.FailOn]
//   - OCTANE_INSECURE_SKIP_VERIFY  → [Config.InsecureSkipVerify]
//     (accepts "true" or "1")
//
// If a numeric or duration variable is present but cannot be parsed,
// ApplyEnv leaves that field unchanged and silently drops the
// invalid value. Strict validation of the final [Config] is the
// caller's responsibility after all layers have been applied.
func ApplyEnv(cfg Config) Config {
	cfg = applyEnvStrings(cfg)
	cfg = applyEnvNumerics(cfg)
	cfg = applyEnvSkipVerify(cfg)

	return cfg
}

// applyEnvStrings applies string-typed environment variables to cfg.
func applyEnvStrings(cfg Config) Config {
	if v := os.Getenv("OCTANE_CACHE_DIR"); v != emptyEnv {
		cfg.CacheDir = v
	}

	if v := os.Getenv("OCTANE_OCPP_VERSION"); v != emptyEnv {
		cfg.OCPPVersion = v
	}

	if v := os.Getenv("OCTANE_FAIL_ON"); v != emptyEnv {
		cfg.FailOn = v
	}

	return cfg
}

// applyEnvNumerics applies numeric and duration environment variables to cfg.
func applyEnvNumerics(cfg Config) Config {
	cfg = applyEnvMaxParallel(cfg)
	cfg = applyEnvLockTimeout(cfg)

	return cfg
}

// applyEnvMaxParallel applies OCTANE_MAX_PARALLEL to cfg when the variable
// is set and contains a valid integer.
func applyEnvMaxParallel(cfg Config) Config {
	v := os.Getenv("OCTANE_MAX_PARALLEL")
	if v == emptyEnv {
		return cfg
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return cfg
	}

	cfg.MaxParallel = parsed

	return cfg
}

// applyEnvLockTimeout applies OCTANE_LOCK_TIMEOUT to cfg when the variable
// is set and contains a valid duration string.
func applyEnvLockTimeout(cfg Config) Config {
	v := os.Getenv("OCTANE_LOCK_TIMEOUT")
	if v == emptyEnv {
		return cfg
	}

	parsed, err := time.ParseDuration(v)
	if err != nil {
		return cfg
	}

	cfg.LockTimeout = parsed

	return cfg
}

// applyEnvSkipVerify applies OCTANE_INSECURE_SKIP_VERIFY to cfg.
// OCTANE_INSECURE_SKIP_VERIFY is intentionally supported via env var
// (needed for headless CI environments that cannot pass CLI flags).
// A warning banner is always emitted by the CLI when this is active.
func applyEnvSkipVerify(cfg Config) Config {
	v := os.Getenv("OCTANE_INSECURE_SKIP_VERIFY")
	if v == skipVerifyTrue || v == skipVerifyOne {
		cfg.InsecureSkipVerify = true
	}

	return cfg
}
