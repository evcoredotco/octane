package exitcode

// Process exit codes used by the octane CLI (spec 006 -10).
//
// OK (0) and TestFailed (1) follow universal CI convention.
// Code 3 is reserved.
// Codes 64–70 align with BSD sysexits(3) ranges for operator-facing
// errors. Code 9 is used for lock-contention timeout.
const (
	// OK is the exit code used when all stories passed or when
	// a read-only command (validate, keywords list, cache info)
	// completes without error.
	OK = 0

	// TestFailed is the exit code used when one or more stories
	// failed execution. CI systems treat any non-zero exit as a
	// build failure; 1 is the canonical value for test failures.
	TestFailed = 1

	// ToolError is the exit code used when an unexpected internal
	// failure occurs that is not attributable to user input — for
	// example a panic, an I/O error, or any other bug in octane
	// itself.
	ToolError = 2

	// CacheLockTimeout is the exit code used when cache lock
	// contention is not resolved within the duration specified by
	// --lock-timeout. See spec 006 -10.
	CacheLockTimeout = 9

	// ConfigError is the exit code used when the configuration
	// file or a command-line flag contains a structural or
	// semantic error (EX_USAGE in sysexits.h). It is returned
	// when YAML is malformed, a required flag is missing, or a
	// flag value cannot be parsed.
	ConfigError = 64

	// StoryParseError is the exit code used when a .story file
	// cannot be parsed. See spec 001.
	StoryParseError = 65

	// KeywordError is the exit code used when keyword resolution
	// fails. See spec 003.
	KeywordError = 66

	// TransportError is the exit code used when the wire
	// transport fails to connect or communicate. See spec 002.
	TransportError = 70
)
