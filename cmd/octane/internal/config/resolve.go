package config

import "time"

// FlagOverrides carries the values of command-line flags that the
// operator explicitly set. A nil pointer means the flag was not
// provided; a non-nil pointer means the operator explicitly set the
// flag, and its value wins over every lower-priority source
// (environment variables, YAML file, defaults).
//
// Using nil pointers (rather than zero values or a parallel bool
// map) lets callers use the same type for both "not set" and
// "set to the zero value of the type", which is important for
// duration flags where 0 might be a valid explicit value.
type FlagOverrides struct {
	// CacheDir overrides [Config.CacheDir] when non-nil.
	CacheDir *string

	// MaxParallel overrides [Config.MaxParallel] when non-nil.
	MaxParallel *int

	// OCPPVersion overrides [Config.OCPPVersion] when non-nil.
	OCPPVersion *string

	// LockTimeout overrides [Config.LockTimeout] when non-nil.
	LockTimeout *time.Duration

	// FailOn overrides [Config.FailOn] when non-nil.
	FailOn *string

	// InsecureSkipVerify overrides [Config.InsecureSkipVerify] when
	// non-nil.
	InsecureSkipVerify *bool

	// Parameters overlays [Config.Parameters] when non-empty. Individual
	// keys override YAML/env-sourced parameter values.
	Parameters map[string]string
}

// Resolve merges cfg with the explicit flag overrides in flags and
// returns the resolved [Config]. Flag values always win over every
// lower-priority source (environment variables, YAML, defaults).
//
// Only non-nil pointer fields in flags are applied; nil fields leave
// the corresponding cfg value unchanged.
func Resolve(cfg Config, flags FlagOverrides) Config {
	if flags.CacheDir != nil {
		cfg.CacheDir = *flags.CacheDir
	}

	if flags.MaxParallel != nil {
		cfg.MaxParallel = *flags.MaxParallel
	}

	if flags.OCPPVersion != nil {
		cfg.OCPPVersion = *flags.OCPPVersion
	}

	if flags.LockTimeout != nil {
		cfg.LockTimeout = *flags.LockTimeout
	}

	if flags.FailOn != nil {
		cfg.FailOn = *flags.FailOn
	}

	if flags.InsecureSkipVerify != nil {
		cfg.InsecureSkipVerify = *flags.InsecureSkipVerify
	}

	if len(flags.Parameters) > 0 {
		cfg.Parameters = mergeParameters(cfg.Parameters, flags.Parameters)
	}

	return cfg
}

func mergeParameters(
	base map[string]string,
	overrides map[string]string,
) map[string]string {
	merged := make(map[string]string, len(base)+len(overrides))

	for key, value := range base {
		merged[key] = value
	}

	for key, value := range overrides {
		merged[key] = value
	}

	return merged
}
