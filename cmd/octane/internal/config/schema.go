package config

import "time"

const (
	// defaultLockTimeoutSeconds is the default lock timeout in seconds.
	defaultLockTimeoutSeconds = 60

	// defaultMaxParallel is the default number of parallel story workers.
	defaultMaxParallel = 1
)

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
	StoriesDir string `yaml:"storiesDir"`

	// CacheDir overrides the default cache directory
	// ($XDG_CACHE_HOME/octane/cache/ or ~/.cache/octane/cache/).
	// When empty the runner resolves the directory from environment
	// variables at runtime.
	CacheDir string `yaml:"cacheDir"`

	// MaxParallel is the maximum number of stories that may execute
	// concurrently. A value of 1 (the default) produces sequential
	// execution. Values greater than 1 enable the parallel worker
	// pool (ADR 0019).
	MaxParallel int `yaml:"maxParallel"`

	// OCPPVersion restricts the run to stories declaring this OCPP
	// version (e.g., "1.6"). When empty all
	// stories are eligible regardless of their declared version.
	OCPPVersion string `yaml:"ocppVersion"`

	// LockTimeout is the maximum duration the runner waits when
	// acquiring a per-cache-key flock. Defaults to 60 s (spec 005
	// G6). A zero value in the YAML is replaced by the default at
	// [Default] time.
	LockTimeout time.Duration `yaml:"lockTimeout"`

	// InsecureSkipVerify disables TLS certificate verification for
	// WebSocket connections. Setting this to true causes the CLI to
	// emit a warning banner before executing any stories. It should
	// never be used in production environments.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`

	// FailOn controls which story outcomes are counted as run-level
	// failures for the purposes of the process exit code. Accepted
	// values are "any" (default) and "major". "any" causes the
	// binary to exit with [exitcode.TestFailed] when at least one
	// story has status=failed. "major" is reserved for future use.
	FailOn string `yaml:"failOn"`

	// Parameters supplies runtime values for placeholders declared by
	// story Meta Parameters. Values are strings at the config boundary;
	// keyword pattern matching performs the final type coercion.
	Parameters map[string]string `yaml:"parameters"`
}

// Default returns the built-in baseline configuration. It is used
// when no octane.yml file is present and as the starting point for
// [Load], [ApplyEnv], and [Resolve] layering.
func Default() Config {
	return Config{
		StoriesDir:         "scenarios",
		CacheDir:           "",
		MaxParallel:        defaultMaxParallel,
		OCPPVersion:        "",
		LockTimeout:        defaultLockTimeoutSeconds * time.Second,
		InsecureSkipVerify: false,
		FailOn:             "any",
		Parameters:         map[string]string{},
	}
}
