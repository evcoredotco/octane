// Package exitcode defines the canonical process exit codes used
// by the octane CLI. The codes follow the BSD sysexits(3) convention
// for values above 1, and the de-facto CI convention of exit 1 for
// test failures.
//
// Callers should invoke [Exec] rather than os.Exit directly so that
// the exit path is easily intercepted in integration tests that
// capture the process exit code.
package exitcode
