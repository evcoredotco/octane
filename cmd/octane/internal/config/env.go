package config

import (
	"os"
	"strconv"
	"time"
)

// ApplyEnv overlays environment variable values on top of cfg and
// returns the updated Config. Fields whose corresponding environment
// variable is unset or empty are left unchanged.
//
// Supported variables:
//
//   - OCTANE_CACHE_DIR             → [Config.CacheDir]
//   - OCTANE_MAX_PARALLEL          → [Config.MaxParallel] (parsed as int)
//   - OCTANE_OCPP_VERSION          → [Config.OCPPVersion]
//   - OCTANE_LOCK_TIMEOUT          → [Config.LockTimeout] (parsed as duration)
//   - OCTANE_FAIL_ON               → [Config.FailOn]
//   - OCTANE_INSECURE_SKIP_VERIFY  → [Config.InsecureSkipVerify] ("true" or "1")
//
// If a numeric or duration variable is present but cannot be parsed,
// ApplyEnv leaves that field unchanged and silently drops the
// invalid value. Strict validation of the final [Config] is the
// caller's responsibility after all layers have been applied.
func ApplyEnv(cfg Config) Config {
	if cacheDir := os.Getenv("OCTANE_CACHE_DIR"); cacheDir != "" {
		cfg.CacheDir = cacheDir
	}

	if maxParallel := os.Getenv("OCTANE_MAX_PARALLEL"); maxParallel != "" {
		if parsed, err := strconv.Atoi(maxParallel); err == nil {
			cfg.MaxParallel = parsed
		}
	}

	if ocppVersion := os.Getenv("OCTANE_OCPP_VERSION"); ocppVersion != "" {
		cfg.OCPPVersion = ocppVersion
	}

	if lockTimeout := os.Getenv("OCTANE_LOCK_TIMEOUT"); lockTimeout != "" {
		if parsed, err := time.ParseDuration(lockTimeout); err == nil {
			cfg.LockTimeout = parsed
		}
	}

	if failOn := os.Getenv("OCTANE_FAIL_ON"); failOn != "" {
		cfg.FailOn = failOn
	}

	// OCTANE_INSECURE_SKIP_VERIFY is intentionally supported via env var
	// (needed for headless CI environments that cannot pass CLI flags).
	// A warning banner is always emitted by the CLI when this is active.
	if skipVerify := os.Getenv("OCTANE_INSECURE_SKIP_VERIFY"); skipVerify == "true" ||
		skipVerify == "1" {
		cfg.InsecureSkipVerify = true
	}

	return cfg
}
