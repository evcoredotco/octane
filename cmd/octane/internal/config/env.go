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
	if cacheDir := os.Getenv("OCTANE_CACHE_DIR"); cacheDir != emptyEnv {
		cfg.CacheDir = cacheDir
	}

	maxParallel := os.Getenv("OCTANE_MAX_PARALLEL")
	if maxParallel != emptyEnv {
		parsed, err := strconv.Atoi(maxParallel)
		if err == nil {
			cfg.MaxParallel = parsed
		}
	}

	ocppVersion := os.Getenv("OCTANE_OCPP_VERSION")
	if ocppVersion != emptyEnv {
		cfg.OCPPVersion = ocppVersion
	}

	lockTimeout := os.Getenv("OCTANE_LOCK_TIMEOUT")
	if lockTimeout != emptyEnv {
		parsed, lockParseErr := time.ParseDuration(lockTimeout)
		if lockParseErr == nil {
			cfg.LockTimeout = parsed
		}
	}

	if failOn := os.Getenv("OCTANE_FAIL_ON"); failOn != emptyEnv {
		cfg.FailOn = failOn
	}

	// OCTANE_INSECURE_SKIP_VERIFY is intentionally supported via env var
	// (needed for headless CI environments that cannot pass CLI flags).
	// A warning banner is always emitted by the CLI when this is active.
	skipVerify := os.Getenv("OCTANE_INSECURE_SKIP_VERIFY")
	if skipVerify == skipVerifyTrue || skipVerify == skipVerifyOne {
		cfg.InsecureSkipVerify = true
	}

	return cfg
}
