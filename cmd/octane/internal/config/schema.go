package config

import "time"

// Config is the authoritative runtime configuration for the octane
// CLI. It is populated by combining defaults, YAML file values,
// environment variable overrides, and command-line flag overrides
// in that priority order (lowest to highest).
//
// All duration fields use [time.Duration] regardless of the source;
// the YAML decoder and [ApplyEnv] parse string representations into
// durations before populating this struct.
type Config struct {
	// StoriesDir is the root directory that [octane run] searches
	// for .story files when no positional story paths are given.
	// Default: "scenarios".
	StoriesDir string `yaml:"stories_dir"`

	// CacheDir overrides the default cache directory
	// ($XDG_CACHE_HOME/octane/cache/ or ~/.cache/octane/cache/).
	// When empty the runner resolves the directory from environment
	// variables at runtime.
	CacheDir string `yaml:"cache_dir"`

	// MaxParallel is the maximum number of stories that may execute
	// concurrently. A value of 1 (the default) produces sequential
	// execution. Values greater than 1 enable the parallel worker
	// pool (ADR 0019).
	MaxParallel int `yaml:"max_parallel"`

	// OCPPVersion restricts the run to stories declaring this OCPP
	// version (e.g., "1.6", "2.0.1", "2.1"). When empty all
	// stories are eligible regardless of their declared version.
	OCPPVersion string `yaml:"ocpp_version"`

	// LockTimeout is the maximum duration the runner waits when
	// acquiring a per-cache-key flock. Defaults to 60 s (spec 005
	// G6). A zero value in the YAML is replaced by the default at
	// [Default] time.
	LockTimeout time.Duration `yaml:"lock_timeout"`

	// InsecureSkipVerify disables TLS certificate verification for
	// WebSocket connections. Setting this to true causes the CLI to
	// emit a warning banner before executing any stories. It should
	// never be used in production environments.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`

	// FailOn controls which story outcomes are counted as run-level
	// failures for the purposes of the process exit code. Accepted
	// values are "any" (default) and "major". "any" causes the
	// binary to exit with [exitcode.TestFailed] when at least one
	// story has status=failed. "major" is reserved for future use.
	FailOn string `yaml:"fail_on"`
}

// Default returns the built-in baseline configuration. It is used
// when no octane.yml file is present and as the starting point for
// [Load], [ApplyEnv], and [Resolve] layering.
func Default() Config {
	return Config{
		StoriesDir:         "scenarios",
		CacheDir:           "",
		MaxParallel:        1,
		OCPPVersion:        "",
		LockTimeout:        60 * time.Second,
		InsecureSkipVerify: false,
		FailOn:             "any",
	}
}
