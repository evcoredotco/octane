package exitcode

import "os"

// Process exit codes used by the octane CLI.
//
// OK (0) and TestFailed (1) follow universal CI convention.
// Codes 2–63 are reserved for future use.
// Codes 64, 74, and 125 follow the BSD sysexits(3) table.
const (
	// OK is the exit code used when all stories passed or when
	// a read-only command (validate, keywords list, cache info)
	// completes without error.
	OK = 0

	// TestFailed is the exit code used when one or more stories
	// failed execution. CI systems treat any non-zero exit as a
	// build failure; 1 is the canonical value for test failures.
	TestFailed = 1

	// ConfigError is the exit code used when the configuration
	// file or a command-line flag contains a structural or
	// semantic error (EX_USAGE in sysexits.h). It is returned
	// when YAML is malformed, a required flag is missing, or a
	// flag value cannot be parsed.
	ConfigError = 64

	// IOError is the exit code used when an I/O operation fails
	// unexpectedly (EX_IOERR in sysexits.h). It is returned
	// when the cache directory cannot be created, a story file
	// cannot be read, or a report cannot be written.
	IOError = 74

	// InternalError is the exit code used when an unexpected
	// internal failure occurs that is not attributable to user
	// input or I/O. It signals a bug in octane itself.
	InternalError = 125
)

// Exec terminates the current process with code. It is a thin
// wrapper around os.Exit that provides a single call site for
// all process-exit decisions in the CLI, making it easy to
// audit and to intercept in tests that capture the exit code.
func Exec(code int) {
	os.Exit(code)
}
